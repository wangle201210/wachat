package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wangle201210/wachat/backend/model"
	"github.com/wangle201210/wachat/backend/repository"

	"github.com/cloudwego/eino/schema"
)

// ChatService provides chat functionality
type ChatService struct {
	ctx       context.Context
	convRepo  *repository.ConversationRepository
	msgRepo   *repository.MessageRepository
	aiService *AIService
}

// NewChatService creates ChatService with repositories
func NewChatService(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	aiService *AIService,
) *ChatService {
	return &ChatService{
		convRepo:  convRepo,
		msgRepo:   msgRepo,
		aiService: aiService,
	}
}

// SetContext sets the runtime context
func (c *ChatService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// GetAIService returns AI service
func (c *ChatService) GetAIService() *AIService {
	return c.aiService
}

// CreateConversation creates a new conversation
func (c *ChatService) CreateConversation(title string) (*model.Conversation, error) {
	now := time.Now()
	dbConv := &model.DBConversation{
		ID:        fmt.Sprintf("conv_%d", now.UnixNano()),
		Title:     title,
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
	}

	if err := c.convRepo.Create(dbConv); err != nil {
		return nil, err
	}

	return &model.Conversation{
		ID:        dbConv.ID,
		Title:     dbConv.Title,
		Messages:  make([]*schema.Message, 0),
		CreatedAt: time.Unix(dbConv.CreatedAt, 0),
		UpdatedAt: time.Unix(dbConv.UpdatedAt, 0),
	}, nil
}

// GetConversation retrieves a conversation by ID with messages
func (c *ChatService) GetConversation(id string) (*model.Conversation, error) {
	dbConv, err := c.convRepo.Get(id)
	if err != nil {
		return nil, err
	}

	// Load messages from database
	dbMessages, err := c.msgRepo.GetByConversation(id)
	if err != nil {
		return nil, err
	}

	// Convert DBMessage to schema.Message
	messages := make([]*schema.Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		messages = append(messages, &schema.Message{
			Role:    schema.RoleType(dbMsg.Role),
			Content: dbMsg.Content,
		})
	}

	return &model.Conversation{
		ID:        dbConv.ID,
		Title:     dbConv.Title,
		Messages:  messages,
		CreatedAt: time.Unix(dbConv.CreatedAt, 0),
		UpdatedAt: time.Unix(dbConv.UpdatedAt, 0),
	}, nil
}

// ListConversations returns all conversations
func (c *ChatService) ListConversations() ([]*model.Conversation, error) {
	dbConvs, err := c.convRepo.List()
	if err != nil {
		return nil, err
	}

	convs := make([]*model.Conversation, 0, len(dbConvs))
	for _, dbConv := range dbConvs {
		convs = append(convs, &model.Conversation{
			ID:        dbConv.ID,
			Title:     dbConv.Title,
			Messages:  make([]*schema.Message, 0), // Don't load messages for list view
			CreatedAt: time.Unix(dbConv.CreatedAt, 0),
			UpdatedAt: time.Unix(dbConv.UpdatedAt, 0),
		})
	}
	return convs, nil
}

// DeleteConversation deletes a conversation
func (c *ChatService) DeleteConversation(id string) error {
	return c.convRepo.Delete(id)
}

// UpdateConversationTitle updates conversation title
func (c *ChatService) UpdateConversationTitle(id, title string) error {
	dbConv, err := c.convRepo.Get(id)
	if err != nil {
		return err
	}

	dbConv.Title = title
	dbConv.UpdatedAt = time.Now().Unix()
	return c.convRepo.Update(dbConv)
}

// ClearMessages clears all messages in a conversation
func (c *ChatService) ClearMessages(id string) error {
	if err := c.msgRepo.DeleteByConversation(id); err != nil {
		return err
	}

	// Update conversation timestamp
	dbConv, err := c.convRepo.Get(id)
	if err != nil {
		return err
	}
	dbConv.UpdatedAt = time.Now().Unix()
	return c.convRepo.Update(dbConv)
}

// SaveMessage saves a message to database
func (c *ChatService) SaveMessage(conversationID string, msg *schema.Message) error {
	dbMsg := &model.DBMessage{
		ID:             fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		ConversationID: conversationID,
		Role:           string(msg.Role),
		Content:        msg.Content,
		Timestamp:      time.Now().Unix(),
		Status:         "sent",
	}

	if err := c.msgRepo.Create(dbMsg); err != nil {
		return err
	}

	// Update conversation timestamp
	dbConv, err := c.convRepo.Get(conversationID)
	if err != nil {
		return err
	}
	dbConv.UpdatedAt = time.Now().Unix()
	return c.convRepo.Update(dbConv)
}

// GenerateConversationTitle generates and updates conversation title based on recent messages
func (c *ChatService) GenerateConversationTitle(conversationID string, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "", nil
	}

	// Collect recent messages (up to last 3 rounds = 6 messages)
	recentMessages := conv.Messages
	if len(recentMessages) > 6 {
		recentMessages = recentMessages[len(recentMessages)-6:]
	}

	// Build context for title generation
	var contextBuilder strings.Builder
	contextBuilder.WriteString("对话内容：\n")
	for _, msg := range recentMessages {
		role := "用户"
		if msg.Role == schema.Assistant {
			role = "助手"
		}
		contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// Create messages for title generation
	titleGenMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "请根据以下对话内容，生成一个简洁的标题（不超过15个字）。只返回标题文本，不要有其他内容。",
		},
		{
			Role:    schema.User,
			Content: contextBuilder.String(),
		},
	}

	// Generate title using AI
	responseChan := make(chan string, 100)
	go func() {
		c.aiService.StreamResponse(titleGenMessages, responseChan)
	}()

	// Collect response
	var titleBuilder strings.Builder
	for chunk := range responseChan {
		titleBuilder.WriteString(chunk)
	}

	title := titleBuilder.String()
	title = strings.TrimSpace(title)

	// Remove line breaks
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")

	// Limit length
	runes := []rune(title)
	if len(runes) > 30 {
		title = string(runes[:30]) + "..."
	}

	if title == "" {
		return "", nil
	}

	// Update title in database
	if err := c.UpdateConversationTitle(conversationID, title); err != nil {
		return "", err
	}

	return title, nil
}

// EventCallback is a function type for event emission
type EventCallback func(eventName string, data interface{})

// SendMessageStream handles the complete message streaming flow
func (c *ChatService) SendMessageStream(conversationID, content string, eventCallback EventCallback) error {
	// Emit stream start event
	eventCallback("stream:start", map[string]interface{}{
		"conversationId": conversationID,
	})

	// Get conversation
	conv, err := c.GetConversation(conversationID)
	if err != nil {
		eventCallback("stream:error", map[string]interface{}{
			"conversationId": conversationID,
			"error":          err.Error(),
		})
		return err
	}

	// Add user message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: content,
	}
	conv.Messages = append(conv.Messages, userMsg)

	// Save user message to database
	if err := c.SaveMessage(conversationID, userMsg); err != nil {
		eventCallback("stream:error", map[string]interface{}{
			"conversationId": conversationID,
			"error":          "Failed to save user message: " + err.Error(),
		})
		return err
	}

	// Generate and update conversation title in background
	go func() {
		if title, err := c.GenerateConversationTitle(conversationID, conv); err == nil && title != "" {
			// Emit title updated event
			eventCallback("conversation:title-updated", map[string]interface{}{
				"conversationId": conversationID,
				"title":          title,
			})
		}
	}()

	// Stream AI response
	responseChan := make(chan string)
	errChan := make(chan error)
	assistantContent := ""

	go func() {
		err := c.aiService.StreamResponse(conv.Messages, responseChan)
		errChan <- err
	}()

	go func() {
		for chunk := range responseChan {
			assistantContent += chunk
			eventCallback("stream:response", map[string]interface{}{
				"conversationId": conversationID,
				"chunk":          chunk,
			})
		}

		// Create assistant message
		assistantMsg := &schema.Message{
			Role:    schema.Assistant,
			Content: assistantContent,
		}
		conv.Messages = append(conv.Messages, assistantMsg)

		// Save assistant message to database
		if err := c.SaveMessage(conversationID, assistantMsg); err != nil {
			eventCallback("stream:error", map[string]interface{}{
				"conversationId": conversationID,
				"error":          "Failed to save assistant message: " + err.Error(),
			})
			return
		}

		eventCallback("stream:end", map[string]interface{}{
			"conversationId": conversationID,
			"message": map[string]string{
				"role":    "assistant",
				"content": assistantContent,
			},
		})
	}()

	if err := <-errChan; err != nil {
		eventCallback("stream:error", map[string]interface{}{
			"conversationId": conversationID,
			"error":          err.Error(),
		})
		return err
	}

	return nil
}
