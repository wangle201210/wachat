package backend

import (
	"context"
	"fmt"

	"github.com/wangle201210/wachat/backend/config"
	"github.com/wangle201210/wachat/backend/database"
	"github.com/wangle201210/wachat/backend/model"
	"github.com/wangle201210/wachat/backend/repository"
	"github.com/wangle201210/wachat/backend/service"

	"github.com/cloudwego/eino/schema"
)

// API is the main entry point for backend functionality
type API struct {
	chatService *service.ChatService
	aiService   *service.AIService
}

// NewAPI creates a new backend API instance
func NewAPI() (*API, error) {
	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize repositories
	convRepo := repository.NewConversationRepository(db.DB)
	msgRepo := repository.NewMessageRepository(db.DB)

	// Initialize services
	aiConfig := config.GetAIConfig()
	aiService := service.NewAIService(aiConfig)
	chatService := service.NewChatService(convRepo, msgRepo, aiService)

	return &API{
		chatService: chatService,
		aiService:   aiService,
	}, nil
}

// SetContext sets the runtime context
func (a *API) SetContext(ctx context.Context) {
	a.chatService.SetContext(ctx)
}

// CreateConversation creates a new conversation
func (a *API) CreateConversation(title string) (*model.Conversation, error) {
	return a.chatService.CreateConversation(title)
}

// ListConversations returns all conversations
func (a *API) ListConversations() ([]*model.Conversation, error) {
	return a.chatService.ListConversations()
}

// GetConversation returns conversation with messages
func (a *API) GetConversation(id string) (*model.Conversation, error) {
	return a.chatService.GetConversation(id)
}

// DeleteConversation deletes conversation
func (a *API) DeleteConversation(id string) error {
	return a.chatService.DeleteConversation(id)
}

// UpdateConversationTitle updates conversation title
func (a *API) UpdateConversationTitle(id, title string) error {
	return a.chatService.UpdateConversationTitle(id, title)
}

// SaveMessage saves a message to database
func (a *API) SaveMessage(conversationID string, msg *schema.Message) error {
	return a.chatService.SaveMessage(conversationID, msg)
}

// SendMessageStream handles the complete message streaming flow
func (a *API) SendMessageStream(conversationID, content string, eventCallback service.EventCallback) error {
	return a.chatService.SendMessageStream(conversationID, content, eventCallback)
}

// GetAIService returns AI service
func (a *API) GetAIService() *service.AIService {
	return a.aiService
}
