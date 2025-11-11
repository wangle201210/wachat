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

// API is the main entry point for backend functionality (GoFrame version)
type API struct {
	chatService        *service.ChatService
	aiService          *service.AIService
	ragService         *service.RAGServiceImpl
	goragServerService *service.GoRagServerService
}

// NewAPI creates a new backend API instance (GoFrame version)
func NewAPI(ctx context.Context) (*API, error) {
	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize repositories
	convRepo := repository.NewConversationRepository(db.DB)
	msgRepo := repository.NewMessageRepository(db.DB)

	// Get configurations
	aiConfig := config.GetAIConfig()
	ragConfig := config.GetRAGConfig()

	// Initialize RAG service (GoFrame version)
	ragService, err := service.NewRAGService(ctx, ragConfig, aiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to init RAG service: %w", err)
	}

	// Initialize AI service
	// 注意：AIService 需要适配 RAGServiceGF
	// 这里暂时使用一个适配器包装
	aiService := service.NewAIService(aiConfig, &ragServiceAdapter{ragService})

	// Initialize chat service
	chatService := service.NewChatService(convRepo, msgRepo, aiService)

	// Initialize go-rag server service (GoFrame version)
	goragServer := service.NewGoRagServerService(ragConfig)

	return &API{
		chatService:        chatService,
		aiService:          aiService,
		ragService:         ragService,
		goragServerService: goragServer,
	}, nil
}

// ragServiceAdapter 适配器，让 RAGServiceImpl 兼容 AIService 的接口
type ragServiceAdapter struct {
	ragService *service.RAGServiceImpl
}

func (a *ragServiceAdapter) IsEnabled() bool {
	return a.ragService != nil && a.ragService.IsEnabled()
}

func (a *ragServiceAdapter) RetrieveWithContext(ctx context.Context, query string) (string, error) {
	if a.ragService == nil {
		return "", nil
	}
	return a.ragService.RetrieveWithContext(ctx, query)
}

func (a *ragServiceAdapter) RetrieveDocuments(ctx context.Context, query string) ([]*schema.Document, error) {
	if a.ragService == nil {
		return nil, nil
	}
	return a.ragService.RetrieveDocuments(ctx, query)
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

// GetRAGService returns RAG service
func (a *API) GetRAGService() *service.RAGServiceImpl {
	return a.ragService
}

// GetGoRagServerService returns go-rag server service
func (a *API) GetGoRagServerService() *service.GoRagServerService {
	return a.goragServerService
}
