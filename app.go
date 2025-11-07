package main

import (
	"context"
	"fmt"

	"github.com/wangle201210/wachat/backend"
	"github.com/wangle201210/wachat/backend/model"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	chatAPI *backend.API
}

// NewApp creates new App
func NewApp() *App {
	api, err := backend.NewAPI()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize API: %v", err))
	}
	return &App{
		chatAPI: api,
	}
}

// startup is called when app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.chatAPI.SetContext(ctx)
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
