package service

import (
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

// QdrantManagerService 管理 Qdrant 的下载、安装、启动
type QdrantManagerService struct {
	ctx       context.Context
	config    *config.QdrantConfig
	cmd       *exec.Cmd
	isRunning bool
	callback  ProgressCallback
	done      chan struct{} // 用于通知进程已停止
}

// NewQdrantManagerService 创建 Qdrant 管理器服务
func NewQdrantManagerService(ctx context.Context, cfg *config.QdrantConfig) *QdrantManagerService {
	return &QdrantManagerService{
		ctx:    ctx,
		config: cfg,
	}
}

// SetProgressCallback 设置进度回调函数
func (q *QdrantManagerService) SetProgressCallback(callback ProgressCallback) {
	q.callback = callback
}

// notifyProgress 通知进度
func (q *QdrantManagerService) notifyProgress(downloaded, total int64, percent float64, status string) {
	if q.callback != nil {
		q.callback(downloaded, total, percent, status)
	}
}

// IsInstalled 检查 Qdrant 是否已安装
func (q *QdrantManagerService) IsInstalled() bool {
	binaryPath := q.getBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// IsRunning 检查 Qdrant 是否正在运行
func (q *QdrantManagerService) IsRunning() bool {
	return q.isRunning && q.cmd != nil && q.cmd.Process != nil
}

// CheckHealth 检查 Qdrant 服务是否健康（检测端口）
func (q *QdrantManagerService) CheckHealth() error {
	// Qdrant 默认 HTTP 端口是 6333
	address := fmt.Sprintf("localhost:%d", q.config.Port)

	// 尝试连接
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to Qdrant server: %w", err)
	}
	conn.Close()

	// 尝试访问健康检查端点
	healthURL := fmt.Sprintf("http://localhost:%d/healthz", q.config.Port)
	resp, err := http.Get(healthURL)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status: %s", resp.Status)
	}

	return nil
}

// WaitForHealth 等待服务健康（最多等待 30 秒）
func (q *QdrantManagerService) WaitForHealth(timeout time.Duration) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for Qdrant server to become healthy")
		case <-ticker.C:
			if err := q.CheckHealth(); err == nil {
				g.Log().Info(q.ctx, "Qdrant server is healthy")
				return nil
			}
			g.Log().Debug(q.ctx, "Waiting for Qdrant server to become healthy...")
		}
	}
}

// Download 下载 Qdrant 二进制文件
func (q *QdrantManagerService) Download() error {
	q.notifyProgress(0, 0, 0, "准备下载 Qdrant...")

	// 确定下载 URL
	downloadURL := q.getDownloadURL()
	g.Log().Infof(q.ctx, "Downloading Qdrant from: %s", downloadURL)

	// 创建安装目录
	installPath := q.config.InstallPath
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 下载文件
	q.notifyProgress(0, 0, 0, "正在连接...")
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
	q.notifyProgress(0, totalSize, 0, "正在下载...")
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
			q.notifyProgress(downloaded, totalSize, percent, "正在下载...")
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
	q.notifyProgress(downloaded, totalSize, 100, "正在解压...")
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

	q.notifyProgress(downloaded, totalSize, 100, "下载完成")
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
	if q.isRunning {
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
	q.cmd = exec.Command(binaryPath)
	q.cmd.Dir = q.config.InstallPath
	q.cmd.Stdout = logFile
	q.cmd.Stderr = logFile

	// 设置环境变量
	q.cmd.Env = os.Environ()

	// 创建 done channel 用于通知进程结束
	q.done = make(chan struct{})

	// 启动进程
	if err := q.cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start Qdrant: %w", err)
	}

	q.isRunning = true
	g.Log().Infof(q.ctx, "Qdrant started successfully (PID: %d, logs: %s)", q.cmd.Process.Pid, logPath)

	// 在后台等待进程结束
	go func() {
		if err := q.cmd.Wait(); err != nil {
			g.Log().Warningf(context.Background(), "Qdrant process exited with error: %v (check logs at: %s)", err, logPath)
		} else {
			g.Log().Info(context.Background(), "Qdrant process exited normally")
		}
		q.isRunning = false
		logFile.Close()
		close(q.done) // 通知进程已结束
	}()

	return nil
}

// Stop 停止 Qdrant 服务
func (q *QdrantManagerService) Stop() error {
	if !q.isRunning || q.cmd == nil || q.cmd.Process == nil {
		return fmt.Errorf("Qdrant is not running")
	}

	g.Log().Info(q.ctx, "Stopping Qdrant...")

	// 发送终止信号
	if err := q.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to stop Qdrant: %w", err)
	}

	// 等待进程结束（最多 5 秒）
	select {
	case <-q.done:
		g.Log().Info(q.ctx, "Qdrant stopped successfully")
	case <-time.After(5 * time.Second):
		g.Log().Warning(q.ctx, "Qdrant did not stop gracefully within timeout")
		q.isRunning = false
	}

	q.cmd = nil
	return nil
}

// GetStatus 获取 Qdrant 服务状态
func (q *QdrantManagerService) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"installed": q.IsInstalled(),
		"running":   q.IsRunning(),
		"healthy":   false,
	}

	if q.IsRunning() {
		if err := q.CheckHealth(); err == nil {
			status["healthy"] = true
		}
	}

	return status
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
