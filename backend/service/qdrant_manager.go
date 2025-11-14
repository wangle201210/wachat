package service

import (
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

// QdrantManagerService 管理 Qdrant 的下载、安装、启动
type QdrantManagerService struct {
	*BaseServiceManager
	config *config.QdrantConfig
}

// NewQdrantManagerService 创建 Qdrant 管理器服务
func NewQdrantManagerService(ctx context.Context, cfg *config.QdrantConfig) *QdrantManagerService {
	return &QdrantManagerService{
		BaseServiceManager: NewBaseServiceManager(ctx, "Qdrant"),
		config:             cfg,
	}
}

// IsInstalled 检查 Qdrant 是否已安装
func (q *QdrantManagerService) IsInstalled() bool {
	binaryPath := q.getBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// CheckHealth 检查 Qdrant 服务是否健康（检测端口）
func (q *QdrantManagerService) CheckHealth() error {
	// Qdrant 默认 HTTP 端口是 6333
	address := fmt.Sprintf("localhost:%d", q.config.Port)

	// 尝试 TCP 连接
	if err := q.CheckTCPHealth(address); err != nil {
		return err
	}

	// 尝试访问健康检查端点
	healthURL := fmt.Sprintf("http://localhost:%d/healthz", q.config.Port)
	return q.CheckHTTPHealth(healthURL)
}

// WaitForHealth 等待服务健康（最多等待指定超时时间）
func (q *QdrantManagerService) WaitForHealth(timeout time.Duration) error {
	return q.BaseServiceManager.WaitForHealth(timeout, q.CheckHealth)
}

// Download 下载 Qdrant 二进制文件
func (q *QdrantManagerService) Download() error {
	q.NotifyProgress(0, 0, 0, "准备下载 Qdrant...")

	// 确定下载 URL
	downloadURL := q.getDownloadURL()
	g.Log().Infof(q.ctx, "Downloading Qdrant from: %s", downloadURL)

	// 创建安装目录
	installPath := q.config.InstallPath
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 下载文件
	q.NotifyProgress(0, 0, 0, "正在连接...")
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

	// 创建临时文件
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	tmpFile := filepath.Join(installPath, "qdrant-download."+ext)
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// 下载并显示进度
	q.NotifyProgress(0, totalSize, 0, "正在下载...")
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
			q.NotifyProgress(downloaded, totalSize, percent, "正在下载...")
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
	q.NotifyProgress(downloaded, totalSize, 100, "正在解压...")
	if err := q.extractArchive(tmpFile, installPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// 删除临时文件
	os.Remove(tmpFile)

	// 设置可执行权限（Unix 系统）
	if runtime.GOOS != "windows" {
		binaryPath := q.getBinaryPath()
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	q.NotifyProgress(downloaded, totalSize, 100, "下载完成")
	g.Log().Info(q.ctx, "Qdrant downloaded successfully")
	return nil
}

// extractArchive 解压文件
func (q *QdrantManagerService) extractArchive(archivePath, destPath string) error {
	// 使用系统命令解压
	var cmd *exec.Cmd
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		// tar -xzf archive.tar.gz -C destPath
		cmd = exec.Command("tar", "-xzf", archivePath, "-C", destPath)
	} else if strings.HasSuffix(archivePath, ".zip") {
		// unzip -o archive.zip -d destPath
		cmd = exec.Command("unzip", "-o", archivePath, "-d", destPath)
	} else {
		return fmt.Errorf("unsupported archive format")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// Start 启动 Qdrant 服务
func (q *QdrantManagerService) Start() error {
	if q.IsRunning() {
		return fmt.Errorf("Qdrant is already running")
	}

	if !q.IsInstalled() {
		return fmt.Errorf("Qdrant is not installed, please download first")
	}

	binaryPath := q.getBinaryPath()
	g.Log().Infof(q.ctx, "Starting Qdrant from: %s", binaryPath)

	// 创建日志文件
	logPath := filepath.Join(q.config.InstallPath, "qdrant.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// 创建命令
	cmd := exec.Command(binaryPath)
	cmd.Dir = q.config.InstallPath
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = os.Environ()

	// 使用基类的 StartProcess 方法
	if err := q.StartProcess(cmd, "Qdrant"); err != nil {
		logFile.Close()
		return err
	}

	g.Log().Infof(q.ctx, "Qdrant logs: %s", logPath)
	return nil
}

// Stop 停止 Qdrant 服务
func (q *QdrantManagerService) Stop() error {
	return q.StopProcess("Qdrant")
}

// GetStatus 获取 Qdrant 服务状态
func (q *QdrantManagerService) GetStatus() map[string]interface{} {
	return q.BaseServiceManager.GetStatus(q.IsInstalled(), q.CheckHealth)
}

// getDownloadURL 获取下载 URL（根据系统和架构）
func (q *QdrantManagerService) getDownloadURL() string {
	baseURL := q.config.DownloadURL

	// 构建文件名：qdrant-{arch}-{os}.{ext}
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Qdrant 使用的架构名称映射
	archMap := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}
	if mappedArch, ok := archMap[arch]; ok {
		arch = mappedArch
	}

	// Qdrant 使用的操作系统名称映射
	osMap := map[string]string{
		"darwin":  "apple-darwin",
		"linux":   "unknown-linux-musl",
		"windows": "pc-windows-msvc",
	}
	if mappedOS, ok := osMap[osName]; ok {
		osName = mappedOS
	}

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	// Qdrant 文件名格式：qdrant-x86_64-apple-darwin.tar.gz
	filename := fmt.Sprintf("qdrant-%s-%s.%s", arch, osName, ext)
	return baseURL + "/" + filename
}

// getBinaryPath 获取二进制文件路径
func (q *QdrantManagerService) getBinaryPath() string {
	binaryName := "qdrant"
	if runtime.GOOS == "windows" {
		binaryName = "qdrant.exe"
	}
	return filepath.Join(q.config.InstallPath, binaryName)
}
