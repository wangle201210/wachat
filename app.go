package main

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/wangle201210/wachat/backend"
	"github.com/wangle201210/wachat/backend/config"
	"github.com/wangle201210/wachat/backend/model"
	"github.com/wangle201210/wachat/backend/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed bin/*
var binaries embed.FS

// App struct
type App struct {
	ctx           context.Context
	chatAPI       *backend.API
	binaryManager *service.BinaryManager
}

// NewApp creates new App
func NewApp(cfg *config.Config) *App {
	// Use background context for API initialization
	api, err := backend.NewAPI(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize API: %v", err))
	}

	// Create binary manager from config
	binaryManager, err := service.NewBinaryManagerFromConfig(cfg.Binaries, binaries)
	if err != nil {
		g.Log().Warningf(context.Background(), "Binary manager: %v", err)
	}

	return &App{
		chatAPI:       api,
		binaryManager: binaryManager,
	}
}

// startup is called when app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.chatAPI.SetContext(ctx)

	// Setup config change callback to notify frontend
	config.SetOnConfigChange(func() {
		g.Log().Info(ctx, "Configuration changed, notifying frontend...")
		runtime.EventsEmit(ctx, "config:changed", map[string]interface{}{
			"message": "配置文件已更新",
		})
	})

	// Start watching config file for changes
	if err := config.WatchConfig(ctx); err != nil {
		g.Log().Warningf(ctx, "Warning: Failed to start config watcher: %v", err)
	}

	// Setup RAG manager progress callback
	ragManager := a.chatAPI.GetRAGManager()
	if ragManager != nil {
		ragManager.SetProgressCallback(func(downloaded, total int64, percent float64, status string) {
			runtime.EventsEmit(ctx, "rag:download:progress", map[string]interface{}{
				"downloaded": downloaded,
				"total":      total,
				"percent":    percent,
				"status":     status,
			})
		})
	}

	// Setup Qdrant manager progress callback
	qdrantManager := a.chatAPI.GetQdrantManager()
	if qdrantManager != nil {
		qdrantManager.SetProgressCallback(func(downloaded, total int64, percent float64, status string) {
			runtime.EventsEmit(ctx, "qdrant:download:progress", map[string]interface{}{
				"downloaded": downloaded,
				"total":      total,
				"percent":    percent,
				"status":     status,
			})
		})
	}

	// Start all embedded binaries
	if a.binaryManager != nil {
		if err := a.binaryManager.StartAll(ctx); err != nil {
			g.Log().Warningf(ctx, "Warning: Failed to start binaries: %v", err)
		}
	}

	// Auto-start Qdrant if configured
	qdrantConfig := config.GetQdrantConfig()
	if qdrantConfig != nil && qdrantConfig.AutoStart {
		if qdrantManager != nil && qdrantManager.IsInstalled() {
			g.Log().Info(ctx, "Auto-starting Qdrant service...")
			if err := qdrantManager.Start(); err != nil {
				g.Log().Warningf(ctx, "Warning: Failed to auto-start Qdrant service: %v", err)
			} else {
				g.Log().Info(ctx, "Qdrant service auto-started successfully")
			}
		} else if qdrantManager != nil {
			g.Log().Info(ctx, "Qdrant auto-start is enabled but Qdrant is not installed yet")
		}
	}

	// Only auto-start go-rag server if autoStart is true
	ragConfig := config.GetRAGConfig()
	if ragConfig != nil && ragConfig.AutoStart {
		if ragManager != nil && ragManager.IsInstalled() {
			g.Log().Info(ctx, "Auto-starting RAG service...")
			if err := ragManager.Start(); err != nil {
				g.Log().Warningf(ctx, "Warning: Failed to auto-start RAG service: %v", err)
			} else {
				g.Log().Info(ctx, "RAG service auto-started successfully")
			}
		} else if ragManager != nil {
			g.Log().Info(ctx, "RAG auto-start is enabled but go-rag is not installed yet")
		}
	}
}

// shutdown is called when app stops
func (a *App) shutdown(ctx context.Context) {
	g.Log().Info(ctx, "Application shutting down...")

	// Stop config watcher
	config.StopWatch()

	// Stop RAG manager (which manages the go-rag process)
	ragManager := a.chatAPI.GetRAGManager()
	if ragManager != nil && ragManager.IsRunning() {
		g.Log().Info(ctx, "Stopping RAG manager...")
		if err := ragManager.Stop(); err != nil {
			g.Log().Warningf(ctx, "Warning: Failed to stop RAG manager: %v", err)
		} else {
			g.Log().Info(ctx, "RAG manager stopped successfully")
		}
	}

	// Stop Qdrant manager
	qdrantManager := a.chatAPI.GetQdrantManager()
	if qdrantManager != nil && qdrantManager.IsRunning() {
		g.Log().Info(ctx, "Stopping Qdrant manager...")
		if err := qdrantManager.Stop(); err != nil {
			g.Log().Warningf(ctx, "Warning: Failed to stop Qdrant manager: %v", err)
		} else {
			g.Log().Info(ctx, "Qdrant manager stopped successfully")
		}
	}

	// Cleanup managed binaries
	if a.binaryManager != nil {
		g.Log().Info(ctx, "Cleaning up binary manager...")
		a.binaryManager.Cleanup()
	}

	g.Log().Info(ctx, "Application shutdown complete")
}

// CreateConversation creates new conversation
func (a *App) CreateConversation(title string) (*model.Conversation, error) {
	return a.chatAPI.CreateConversation(title)
}

// ListConversations returns all conversations
func (a *App) ListConversations() ([]*model.Conversation, error) {
	return a.chatAPI.ListConversations()
}

// GetConversation returns conversation with messages
func (a *App) GetConversation(id string) (*model.Conversation, error) {
	return a.chatAPI.GetConversation(id)
}

// DeleteConversation deletes conversation
func (a *App) DeleteConversation(id string) error {
	return a.chatAPI.DeleteConversation(id)
}

// SendMessageStream streams AI response using eino
func (a *App) SendMessageStream(conversationID, content string) error {
	// Create event callback that emits Wails runtime events
	eventCallback := func(eventName string, data interface{}) {
		runtime.EventsEmit(a.ctx, eventName, data)
	}

	// Delegate to service layer
	return a.chatAPI.SendMessageStream(conversationID, content, eventCallback)
}

// RAGServerInfo holds RAG server configuration info
type RAGServerInfo struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

// GetRAGServerInfo returns RAG server configuration
func (a *App) GetRAGServerInfo() *RAGServerInfo {
	cfg := config.GetRAGConfig()

	// Check if RAG is enabled in config
	if cfg == nil || !cfg.Enabled {
		return &RAGServerInfo{
			Enabled: false,
			URL:     "",
		}
	}

	// Construct URL from config
	url := fmt.Sprintf("http://localhost%s", cfg.Server.Address)

	return &RAGServerInfo{
		Enabled: true,
		URL:     url,
	}
}

// RAG Manager Methods

// DownloadRAG downloads go-rag binary with progress
func (a *App) DownloadRAG() error {
	runtime.EventsEmit(a.ctx, "rag:download:start", nil)
	err := a.chatAPI.DownloadRAG()
	if err != nil {
		runtime.EventsEmit(a.ctx, "rag:download:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	runtime.EventsEmit(a.ctx, "rag:download:complete", nil)
	return nil
}

// StartRAG starts go-rag service
func (a *App) StartRAG() error {
	// Check if Qdrant is required and start it if needed
	qdrantManager := a.chatAPI.GetQdrantManager()
	qdrantConfig := config.GetQdrantConfig()

	// If Qdrant is enabled and not running, start it first
	if qdrantConfig != nil && qdrantConfig.IsEnabled() && qdrantManager != nil {
		if !qdrantManager.IsRunning() {
			g.Log().Info(a.ctx, "Qdrant is required but not running, starting it first...")
			runtime.EventsEmit(a.ctx, "rag:start:progress", map[string]interface{}{
				"status": "正在启动 Qdrant...",
			})

			// Check if Qdrant is installed
			if !qdrantManager.IsInstalled() {
				err := fmt.Errorf("Qdrant 未安装，请先下载并安装 Qdrant")
				runtime.EventsEmit(a.ctx, "rag:start:error", map[string]interface{}{
					"error": err.Error(),
				})
				return err
			}

			// Start Qdrant
			if err := qdrantManager.Start(); err != nil {
				runtime.EventsEmit(a.ctx, "rag:start:error", map[string]interface{}{
					"error": fmt.Sprintf("启动 Qdrant 失败: %v", err),
				})
				return err
			}

			// Wait for Qdrant to become healthy
			runtime.EventsEmit(a.ctx, "rag:start:progress", map[string]interface{}{
				"status": "等待 Qdrant 服务启动...",
			})
			if err := qdrantManager.WaitForHealth(30 * time.Second); err != nil {
				runtime.EventsEmit(a.ctx, "rag:start:error", map[string]interface{}{
					"error": fmt.Sprintf("Qdrant 服务启动超时: %v", err),
				})
				return err
			}
			g.Log().Info(a.ctx, "Qdrant started successfully")
		} else {
			g.Log().Info(a.ctx, "Qdrant is already running")
		}
	}

	// Now start RAG service
	runtime.EventsEmit(a.ctx, "rag:start:progress", map[string]interface{}{
		"status": "正在启动 RAG 服务...",
	})

	err := a.chatAPI.StartRAG()
	if err != nil {
		runtime.EventsEmit(a.ctx, "rag:start:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Wait for service to become healthy
	ragManager := a.chatAPI.GetRAGManager()
	if ragManager != nil {
		runtime.EventsEmit(a.ctx, "rag:start:progress", map[string]interface{}{
			"status": "等待 RAG 服务启动...",
		})

		if err := ragManager.WaitForHealth(30 * time.Second); err != nil {
			runtime.EventsEmit(a.ctx, "rag:start:error", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	runtime.EventsEmit(a.ctx, "rag:start:complete", nil)
	return nil
}

// StopRAG stops go-rag service
func (a *App) StopRAG() error {
	g.Log().Info(a.ctx, "StopRAG called from frontend")

	runtime.EventsEmit(a.ctx, "rag:stop:progress", map[string]interface{}{
		"status": "正在停止...",
	})

	g.Log().Info(a.ctx, "Calling chatAPI.StopRAG()")
	err := a.chatAPI.StopRAG()
	if err != nil {
		g.Log().Errorf(a.ctx, "StopRAG error: %v", err)
		runtime.EventsEmit(a.ctx, "rag:stop:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	g.Log().Info(a.ctx, "StopRAG completed successfully, emitting rag:stop:complete")
	runtime.EventsEmit(a.ctx, "rag:stop:complete", nil)
	return nil
}

// GetRAGStatus returns RAG service status
func (a *App) GetRAGStatus() map[string]interface{} {
	return a.chatAPI.GetRAGStatus()
}

// CheckRAGHealth checks if RAG service is healthy
func (a *App) CheckRAGHealth() error {
	return a.chatAPI.CheckRAGHealth()
}

// GetRAGConfig reads RAG config file content
func (a *App) GetRAGConfig() (string, error) {
	return a.chatAPI.GetRAGConfigContent()
}

// SaveRAGConfig saves RAG config file content
func (a *App) SaveRAGConfig(content string) error {
	return a.chatAPI.SaveRAGConfigContent(content)
}

// Qdrant Manager Methods

// DownloadQdrant downloads Qdrant binary with progress
func (a *App) DownloadQdrant() error {
	runtime.EventsEmit(a.ctx, "qdrant:download:start", nil)
	err := a.chatAPI.DownloadQdrant()
	if err != nil {
		runtime.EventsEmit(a.ctx, "qdrant:download:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	runtime.EventsEmit(a.ctx, "qdrant:download:complete", nil)
	return nil
}

// StartQdrant starts Qdrant service
func (a *App) StartQdrant() error {
	runtime.EventsEmit(a.ctx, "qdrant:start:progress", map[string]interface{}{
		"status": "正在启动 Qdrant...",
	})

	err := a.chatAPI.StartQdrant()
	if err != nil {
		runtime.EventsEmit(a.ctx, "qdrant:start:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Wait for service to become healthy
	qdrantManager := a.chatAPI.GetQdrantManager()
	if qdrantManager != nil {
		runtime.EventsEmit(a.ctx, "qdrant:start:progress", map[string]interface{}{
			"status": "等待 Qdrant 服务启动...",
		})

		if err := qdrantManager.WaitForHealth(30 * time.Second); err != nil {
			runtime.EventsEmit(a.ctx, "qdrant:start:error", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	runtime.EventsEmit(a.ctx, "qdrant:start:complete", nil)
	return nil
}

// StopQdrant stops Qdrant service
func (a *App) StopQdrant() error {
	g.Log().Info(a.ctx, "StopQdrant called from frontend")

	runtime.EventsEmit(a.ctx, "qdrant:stop:progress", map[string]interface{}{
		"status": "正在停止 Qdrant...",
	})

	err := a.chatAPI.StopQdrant()
	if err != nil {
		g.Log().Errorf(a.ctx, "StopQdrant error: %v", err)
		runtime.EventsEmit(a.ctx, "qdrant:stop:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	g.Log().Info(a.ctx, "StopQdrant completed successfully")
	runtime.EventsEmit(a.ctx, "qdrant:stop:complete", nil)
	return nil
}

// GetQdrantStatus returns Qdrant service status
func (a *App) GetQdrantStatus() map[string]interface{} {
	return a.chatAPI.GetQdrantStatus()
}

// CheckQdrantHealth checks if Qdrant service is healthy
func (a *App) CheckQdrantHealth() error {
	return a.chatAPI.CheckQdrantHealth()
}

// RAGSettings represents RAG configuration settings
type RAGSettings struct {
	TopK                 int    `json:"topK"`
	DefaultKnowledgeBase string `json:"defaultKnowledgeBase"`
}

// GetRAGSettings returns current RAG settings
func (a *App) GetRAGSettings() *RAGSettings {
	topK, defaultKB := config.GetRAGSettings()
	return &RAGSettings{
		TopK:                 topK,
		DefaultKnowledgeBase: defaultKB,
	}
}

// UpdateRAGSettings updates RAG settings
func (a *App) UpdateRAGSettings(topK int, defaultKnowledgeBase string) error {
	if topK < 1 {
		return fmt.Errorf("topK must be at least 1")
	}
	if topK > 100 {
		return fmt.Errorf("topK must not exceed 100")
	}

	return config.UpdateRAGSettings(a.ctx, topK, defaultKnowledgeBase)
}

// GetKnowledgeBases returns list of available knowledge bases
func (a *App) GetKnowledgeBases() ([]string, error) {
	return a.chatAPI.GetKnowledgeBases(a.ctx)
}
