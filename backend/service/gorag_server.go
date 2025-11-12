package service

import (
	"context"
	"fmt"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/wangle201210/go-rag/server/server"
	"github.com/wangle201210/wachat/backend/config"
)

// GoRagServerService 管理 go-rag HTTP 服务器（使用 GoFrame 共享配置）
type GoRagServerService struct {
	ctx       context.Context
	cancel    context.CancelFunc
	config    *config.RAGConfig
	isRunning bool
	done      chan struct{}
}

// NewGoRagServerService 创建 go-rag 服务器服务（GoFrame版本）
func NewGoRagServerService(cfg *config.RAGConfig) *GoRagServerService {
	return &GoRagServerService{
		config: cfg,
		done:   make(chan struct{}),
	}
}

// IsEnabled 检查服务是否应该启用
func (s *GoRagServerService) IsEnabled() bool {
	return s.config != nil && s.config.IsServerEnabled()
}

// Start 启动 go-rag 服务器（在 goroutine 中）
// 使用共享的 GoFrame 配置，不需要生成新的配置文件
func (s *GoRagServerService) Start(ctx context.Context) error {
	if !s.IsEnabled() {
		g.Log().Info(ctx, "Go-rag server is disabled (server.address is empty)")
		return nil
	}

	if s.isRunning {
		return fmt.Errorf("go-rag server is already running")
	}

	// 创建可取消的 context，用于控制服务器生命周期
	s.ctx, s.cancel = context.WithCancel(ctx)

	g.Log().Info(ctx, "Starting go-rag server...")
	g.Log().Infof(ctx, "Server address: %s", s.config.Server.Address)

	// Note: go-rag 会使用全局的 g.Cfg() 读取配置
	// 如果配置不正确（如缺少 ES、Database 等），go-rag 启动时会自行报错
	// 我们不需要在这里预先检查这些配置

	// 在单独的 goroutine 中启动 go-rag 服务器
	go func() {
		defer func() {
			s.isRunning = false
			close(s.done)
		}()

		// 等待一小段时间让服务器启动
		time.Sleep(20 * time.Second)
		// 调用 go-rag 的公开 Start 函数
		// 注意：server.Start 是阻塞调用，会启动 HTTP 服务器
		// go-rag 会自动使用 GoFrame 的全局配置
		// 传入可取消的 context，这样 Stop 方法可以通过取消 context 来停止服务器
		server.Start(s.ctx)
		g.Log().Info(context.Background(), "Go-rag server goroutine exited")
	}()

	s.isRunning = true
	g.Log().Infof(ctx, "Go-rag server started successfully on %s", s.config.Server.Address)
	g.Log().Info(ctx, "Access go-rag Web UI at: http://localhost"+s.config.Server.Address)

	return nil
}

// Stop 停止 go-rag 服务器
func (s *GoRagServerService) Stop() error {
	if !s.isRunning {
		g.Log().Debug(context.Background(), "Go-rag server is not running, skip stop")
		return nil
	}

	ctx := context.Background()
	g.Log().Info(ctx, "Stopping go-rag server...")

	// 1. 取消 context，触发 GoFrame 服务器优雅关闭
	if s.cancel != nil {
		s.cancel()
		g.Log().Debug(ctx, "Context cancelled, waiting for server shutdown...")
	}

	// 2. 等待 goroutine 完成（或超时）
	shutdownTimeout := 10 * time.Second
	shutdownTimer := time.NewTimer(shutdownTimeout)
	defer shutdownTimer.Stop()

	select {
	case <-s.done:
		// 服务器正常关闭
		g.Log().Info(ctx, "Go-rag server stopped gracefully")
		return nil

	case <-shutdownTimer.C:
		// 超时：服务器未能在规定时间内关闭
		g.Log().Warning(ctx, "Go-rag server shutdown timeout after %v", shutdownTimeout)

		// 强制标记为未运行（即使 goroutine 可能还在运行）
		s.isRunning = false

		// 返回错误提示用户服务器可能未完全关闭
		return fmt.Errorf("go-rag server shutdown timeout after %v, may still be running in background", shutdownTimeout)
	}
}

// IsRunning 检查服务器是否正在运行
func (s *GoRagServerService) IsRunning() bool {
	return s.isRunning
}
