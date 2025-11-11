package main

import (
	"context"
	"embed"
	"fmt"

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

	// Start all embedded binaries
	if a.binaryManager != nil {
		if err := a.binaryManager.StartAll(ctx); err != nil {
			g.Log().Warningf(ctx, "Warning: Failed to start binaries: %v", err)
		}
	}

	// Start go-rag server if enabled
	goragServer := a.chatAPI.GetGoRagServerService()
	if goragServer != nil && goragServer.IsEnabled() {
		if err := goragServer.Start(ctx); err != nil {
			g.Log().Warningf(ctx, "Warning: Failed to start go-rag server: %v", err)
		}
	}
}

// shutdown is called when app stops
func (a *App) shutdown(ctx context.Context) {
	// Stop go-rag server
	goragServer := a.chatAPI.GetGoRagServerService()
	if goragServer != nil && goragServer.IsRunning() {
		if err := goragServer.Stop(); err != nil {
			g.Log().Warningf(ctx, "Warning: Failed to stop go-rag server: %v", err)
		}
	}

	// Cleanup managed binaries
	if a.binaryManager != nil {
		a.binaryManager.Cleanup()
	}
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
	goragServer := a.chatAPI.GetGoRagServerService()
	if goragServer == nil || !goragServer.IsEnabled() {
		return &RAGServerInfo{
			Enabled: false,
			URL:     "",
		}
	}

	// Construct URL from config
	cfg := config.GetRAGConfig()
	url := fmt.Sprintf("http://localhost%s", cfg.Server.Address)

	return &RAGServerInfo{
		Enabled: true,
		URL:     url,
	}
}
