package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/wangle201210/wachat/backend/config"
)

// ProgressCallback 下载进度回调函数
type ProgressCallback func(downloaded, total int64, percent float64, status string)

// RAGManagerService 管理 go-rag 的下载、安装、启动
type RAGManagerService struct {
	*BaseServiceManager
	config *config.RAGConfig
}

// NewRAGManagerService 创建 RAG 管理器服务
func NewRAGManagerService(ctx context.Context, cfg *config.RAGConfig) *RAGManagerService {
	return &RAGManagerService{
		BaseServiceManager: NewBaseServiceManager(ctx, "RAG"),
		config:             cfg,
	}
}

// IsInstalled 检查 go-rag 是否已安装
func (r *RAGManagerService) IsInstalled() bool {
	binaryPath := r.getBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// CheckHealth 检查 go-rag 服务是否健康（检测端口）
func (r *RAGManagerService) CheckHealth() error {
	if r.config.Server == nil || r.config.Server.Address == "" {
		return fmt.Errorf("server address not configured")
	}

	// 解析地址（如 ":8000"）
	address := r.config.Server.Address
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}

	return r.CheckTCPHealth(address)
}

// WaitForHealth 等待服务健康（最多等待 30 秒）
func (r *RAGManagerService) WaitForHealth(timeout time.Duration) error {
	return r.BaseServiceManager.WaitForHealth(timeout, r.CheckHealth)
}

// Download 下载 go-rag 二进制文件
func (r *RAGManagerService) Download() error {
	r.NotifyProgress(0, 0, 0, "准备下载...")

	// 确定下载 URL
	downloadURL := r.getDownloadURL()
	g.Log().Infof(r.ctx, "Downloading go-rag from: %s", downloadURL)

	// 创建安装目录
	installPath := r.config.InstallPath
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 下载文件
	r.NotifyProgress(0, 0, 0, "正在连接...")
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// 获取文件大小
	totalSize := resp.ContentLength

	// 创建临时文件（保留原始扩展名，以便 extractArchive 识别格式）
	// 从 URL 中提取文件名
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	tmpFile := filepath.Join(installPath, "go-rag-download."+ext)
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// 下载并显示进度
	r.NotifyProgress(0, totalSize, 0, "正在下载...")
	downloaded := int64(0)
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write file: %w", writeErr)
			}
			downloaded += int64(n)
			percent := float64(downloaded) / float64(totalSize) * 100
			r.NotifyProgress(downloaded, totalSize, percent, "正在下载...")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	out.Close()

	// 解压
	r.NotifyProgress(downloaded, totalSize, 100, "正在解压...")
	if err := r.extractArchive(tmpFile, installPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// 删除临时文件
	os.Remove(tmpFile)

	// 设置可执行权限（Unix 系统）
	if runtime.GOOS != "windows" {
		binaryPath := r.getBinaryPath()
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	r.NotifyProgress(downloaded, totalSize, 100, "下载完成")
	g.Log().Info(r.ctx, "Go-rag downloaded successfully")
	return nil
}

// extractArchive 解压文件
func (r *RAGManagerService) extractArchive(archivePath, destPath string) error {
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		return r.extractTarGz(archivePath, destPath)
	} else if strings.HasSuffix(archivePath, ".zip") {
		return r.extractZip(archivePath, destPath)
	}
	return fmt.Errorf("unsupported archive format")
}

// extractTarGz 解压 tar.gz 文件
// 自动去掉第一层目录（类似 tar --strip-components=1）
func (r *RAGManagerService) extractTarGz(archivePath, destPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 去掉第一层目录（strip-components=1）
		// 例如：go-rag/go-rag -> go-rag
		//      go-rag/static/index.html -> static/index.html
		parts := strings.Split(header.Name, "/")
		if len(parts) <= 1 {
			// 跳过顶层目录本身
			continue
		}
		strippedName := filepath.Join(parts[1:]...)
		if strippedName == "" {
			continue
		}

		target := filepath.Join(destPath, strippedName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			// 创建文件时使用原始权限
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// extractZip 解压 zip 文件
func (r *RAGManagerService) extractZip(archivePath, destPath string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		target := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		// 创建文件时使用原始权限
		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}

// Start 启动 go-rag 服务
func (r *RAGManagerService) Start() error {
	if r.IsRunning() {
		return fmt.Errorf("go-rag is already running")
	}

	if !r.IsInstalled() {
		return fmt.Errorf("go-rag is not installed, please download first")
	}

	binaryPath := r.getBinaryPath()
	g.Log().Infof(r.ctx, "Starting go-rag from: %s", binaryPath)

	// 创建命令
	cmd := exec.Command(binaryPath)
	cmd.Dir = r.config.InstallPath

	// 设置环境变量（传递配置文件路径）
	configPath := os.Getenv("WACHAT_CONFIG_PATH")
	if configPath == "" {
		if cwd, err := os.Getwd(); err == nil {
			configPath = cwd
		}
	}
	cmd.Env = append(os.Environ(), "WACHAT_CONFIG_PATH="+configPath)

	// 使用基类的 StartProcess 方法
	return r.StartProcess(cmd, "Go-rag")
}

// Stop 停止 go-rag 服务
func (r *RAGManagerService) Stop() error {
	return r.StopProcess("Go-rag")
}

// GetStatus 获取 RAG 服务状态
func (r *RAGManagerService) GetStatus() map[string]interface{} {
	return r.BaseServiceManager.GetStatus(r.IsInstalled(), r.CheckHealth)
}

// getDownloadURL 获取下载 URL（根据系统和架构）
func (r *RAGManagerService) getDownloadURL() string {
	baseURL := r.config.DownloadURL

	// 构建文件名：go-rag-{os}-{arch}.{ext}
	osName := runtime.GOOS
	arch := runtime.GOARCH
	ext := "tar.gz"
	if osName == "windows" {
		ext = "zip"
	}

	filename := fmt.Sprintf("go-rag-%s-%s.%s", osName, arch, ext)
	return baseURL + "/" + filename
}

// getBinaryPath 获取二进制文件路径
func (r *RAGManagerService) getBinaryPath() string {
	binaryName := "go-rag"
	if runtime.GOOS == "windows" {
		binaryName = "go-rag.exe"
	}
	return filepath.Join(r.config.InstallPath, binaryName)
}

// getConfigPath 获取配置文件路径
func (r *RAGManagerService) getConfigPath() string {
	return filepath.Join(r.config.InstallPath, "config.yaml")
}

// GetConfigContent 读取配置文件内容
func (r *RAGManagerService) GetConfigContent() (string, error) {
	if !r.IsInstalled() {
		return "", fmt.Errorf("go-rag is not installed")
	}

	configPath := r.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	return string(data), nil
}

// SaveConfigContent 保存配置文件内容
func (r *RAGManagerService) SaveConfigContent(content string) error {
	if !r.IsInstalled() {
		return fmt.Errorf("go-rag is not installed")
	}

	configPath := r.getConfigPath()

	// 备份原配置文件
	backupPath := configPath + ".backup"
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Rename(configPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup config file: %w", err)
		}
	}

	// 保存新配置
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		// 如果保存失败，恢复备份
		if _, statErr := os.Stat(backupPath); statErr == nil {
			os.Rename(backupPath, configPath)
		}
		return fmt.Errorf("failed to save config file: %w", err)
	}

	// 删除备份
	os.Remove(backupPath)

	g.Log().Info(r.ctx, "RAG config file saved successfully")
	return nil
}
