package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
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
	ctx       context.Context
	config    *config.RAGConfig
	cmd       *exec.Cmd
	isRunning bool
	callback  ProgressCallback
	done      chan struct{} // 用于通知进程已停止
}

// NewRAGManagerService 创建 RAG 管理器服务
func NewRAGManagerService(ctx context.Context, cfg *config.RAGConfig) *RAGManagerService {
	return &RAGManagerService{
		ctx:    ctx,
		config: cfg,
	}
}

// SetProgressCallback 设置进度回调函数
func (r *RAGManagerService) SetProgressCallback(callback ProgressCallback) {
	r.callback = callback
}

// notifyProgress 通知进度
func (r *RAGManagerService) notifyProgress(downloaded, total int64, percent float64, status string) {
	if r.callback != nil {
		r.callback(downloaded, total, percent, status)
	}
}

// IsInstalled 检查 go-rag 是否已安装
func (r *RAGManagerService) IsInstalled() bool {
	binaryPath := r.getBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// IsRunning 检查 go-rag 是否正在运行
func (r *RAGManagerService) IsRunning() bool {
	return r.isRunning && r.cmd != nil && r.cmd.Process != nil
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

	// 尝试连接
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to go-rag server: %w", err)
	}
	conn.Close()
	return nil
}

// WaitForHealth 等待服务健康（最多等待 30 秒）
func (r *RAGManagerService) WaitForHealth(timeout time.Duration) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for go-rag server to become healthy")
		case <-ticker.C:
			if err := r.CheckHealth(); err == nil {
				g.Log().Info(r.ctx, "Go-rag server is healthy")
				return nil
			}
			g.Log().Debug(r.ctx, "Waiting for go-rag server to become healthy...")
		}
	}
}

// Download 下载 go-rag 二进制文件
func (r *RAGManagerService) Download() error {
	r.notifyProgress(0, 0, 0, "准备下载...")

	// 确定下载 URL
	downloadURL := r.getDownloadURL()
	g.Log().Infof(r.ctx, "Downloading go-rag from: %s", downloadURL)

	// 创建安装目录
	installPath := r.config.InstallPath
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 下载文件
	r.notifyProgress(0, 0, 0, "正在连接...")
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
	r.notifyProgress(0, totalSize, 0, "正在下载...")
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
			r.notifyProgress(downloaded, totalSize, percent, "正在下载...")
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
	r.notifyProgress(downloaded, totalSize, 100, "正在解压...")
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

	r.notifyProgress(downloaded, totalSize, 100, "下载完成")
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
	if r.isRunning {
		return fmt.Errorf("go-rag is already running")
	}

	if !r.IsInstalled() {
		return fmt.Errorf("go-rag is not installed, please download first")
	}

	binaryPath := r.getBinaryPath()
	g.Log().Infof(r.ctx, "Starting go-rag from: %s", binaryPath)

	// 创建命令
	r.cmd = exec.Command(binaryPath)
	r.cmd.Dir = r.config.InstallPath

	// 设置环境变量（传递配置文件路径）
	configPath := os.Getenv("WACHAT_CONFIG_PATH")
	if configPath == "" {
		if cwd, err := os.Getwd(); err == nil {
			configPath = cwd
		}
	}
	r.cmd.Env = append(os.Environ(), "WACHAT_CONFIG_PATH="+configPath)

	// 创建 done channel 用于通知进程结束
	r.done = make(chan struct{})

	// 启动进程
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start go-rag: %w", err)
	}

	r.isRunning = true
	g.Log().Info(r.ctx, "Go-rag started successfully")

	// 在后台等待进程结束
	go func() {
		if err := r.cmd.Wait(); err != nil {
			g.Log().Warningf(context.Background(), "Go-rag process exited with error: %v", err)
		} else {
			g.Log().Info(context.Background(), "Go-rag process exited normally")
		}
		r.isRunning = false
		close(r.done) // 通知进程已结束
	}()

	return nil
}

// Stop 停止 go-rag 服务
func (r *RAGManagerService) Stop() error {
	if !r.isRunning || r.cmd == nil || r.cmd.Process == nil {
		return fmt.Errorf("go-rag is not running")
	}

	g.Log().Info(r.ctx, "Stopping go-rag...")

	// 发送终止信号
	if err := r.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to stop go-rag: %w", err)
	}

	// 等待进程结束（最多 5 秒）
	// 不能再次调用 Wait()，因为 Start() 中的 goroutine 已经在等待
	// 使用 done channel 来等待
	select {
	case <-r.done:
		g.Log().Info(r.ctx, "Go-rag stopped successfully")
	case <-time.After(5 * time.Second):
		g.Log().Warning(r.ctx, "Go-rag did not stop gracefully within timeout")
		r.isRunning = false
	}

	r.cmd = nil
	return nil
}

// GetStatus 获取 RAG 服务状态
func (r *RAGManagerService) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"installed": r.IsInstalled(),
		"running":   r.IsRunning(),
		"healthy":   false,
	}

	if r.IsRunning() {
		if err := r.CheckHealth(); err == nil {
			status["healthy"] = true
		}
	}

	return status
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
