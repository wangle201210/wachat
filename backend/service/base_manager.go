package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// BaseServiceManager 提供通用的服务管理功能
type BaseServiceManager struct {
	ctx         context.Context
	serviceName string
	cmd         *exec.Cmd
	isRunning   bool
	callback    ProgressCallback
	done        chan struct{}
}

// NewBaseServiceManager 创建基础服务管理器
func NewBaseServiceManager(ctx context.Context, serviceName string) *BaseServiceManager {
	return &BaseServiceManager{
		ctx:         ctx,
		serviceName: serviceName,
	}
}

// SetProgressCallback 设置进度回调函数
func (b *BaseServiceManager) SetProgressCallback(callback ProgressCallback) {
	b.callback = callback
}

// NotifyProgress 通知进度
func (b *BaseServiceManager) NotifyProgress(downloaded, total int64, percent float64, status string) {
	if b.callback != nil {
		b.callback(downloaded, total, percent, status)
	}
}

// IsRunning 检查服务是否正在运行
func (b *BaseServiceManager) IsRunning() bool {
	return b.isRunning && b.cmd != nil && b.cmd.Process != nil
}

// CheckTCPHealth 检查 TCP 端口健康状态
func (b *BaseServiceManager) CheckTCPHealth(address string) error {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to %s server: %w", b.serviceName, err)
	}
	conn.Close()
	return nil
}

// CheckHTTPHealth 检查 HTTP 端点健康状态
func (b *BaseServiceManager) CheckHTTPHealth(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status: %s", resp.Status)
	}

	return nil
}

// WaitForHealth 等待服务健康（最多等待指定超时时间）
func (b *BaseServiceManager) WaitForHealth(timeout time.Duration, healthChecker func() error) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for %s server to become healthy", b.serviceName)
		case <-ticker.C:
			if err := healthChecker(); err == nil {
				g.Log().Infof(b.ctx, "%s server is healthy", b.serviceName)
				return nil
			}
			g.Log().Debugf(b.ctx, "Waiting for %s server to become healthy...", b.serviceName)
		}
	}
}

// StartProcess 启动进程（通用方法）
func (b *BaseServiceManager) StartProcess(cmd *exec.Cmd, processName string) error {
	b.cmd = cmd
	b.done = make(chan struct{})

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", processName, err)
	}

	b.isRunning = true
	g.Log().Infof(b.ctx, "%s started successfully (PID: %d)", processName, cmd.Process.Pid)

	// 在后台等待进程结束
	go func(name string, process *exec.Cmd) {
		if err := process.Wait(); err != nil {
			g.Log().Warningf(context.Background(), "%s process exited with error: %v", name, err)
		} else {
			g.Log().Infof(context.Background(), "%s process exited normally", name)
		}
		b.isRunning = false
		close(b.done)
	}(processName, cmd)

	return nil
}

// StopProcess 停止进程（通用方法）
func (b *BaseServiceManager) StopProcess(serviceName string) error {
	if !b.isRunning || b.cmd == nil || b.cmd.Process == nil {
		return fmt.Errorf("%s is not running", serviceName)
	}

	g.Log().Infof(b.ctx, "Stopping %s...", serviceName)

	if err := b.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to stop %s: %w", serviceName, err)
	}

	// 等待进程结束（最多 5 秒）
	select {
	case <-b.done:
		g.Log().Infof(b.ctx, "%s stopped successfully", serviceName)
	case <-time.After(5 * time.Second):
		g.Log().Warningf(b.ctx, "%s did not stop gracefully within timeout", serviceName)
		b.isRunning = false
	}

	b.cmd = nil
	return nil
}

// GetStatus 获取服务状态（通用方法）
func (b *BaseServiceManager) GetStatus(isInstalled bool, healthChecker func() error) map[string]interface{} {
	status := map[string]interface{}{
		"installed": isInstalled,
		"running":   b.IsRunning(),
		"healthy":   false,
	}

	if b.IsRunning() && healthChecker != nil {
		if err := healthChecker(); err == nil {
			status["healthy"] = true
		}
	}

	return status
}
