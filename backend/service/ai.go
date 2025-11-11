package service

import (
	"context"
	"fmt"
	"io"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/wangle201210/wachat/backend/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// RAGService interface for any RAG implementation
type RAGService interface {
	IsEnabled() bool
	RetrieveWithContext(ctx context.Context, query string) (string, error)
	RetrieveDocuments(ctx context.Context, query string) ([]*schema.Document, error)
}

// AIService handles AI interactions using eino ChatModel
type AIService struct {
	chatModel  *openai.ChatModel
	ctx        context.Context
	config     *config.AIConfig
	ragService RAGService
}

// NewAIService creates AI service with eino ChatModel and optional RAG
func NewAIService(cfg *config.AIConfig, ragService RAGService) *AIService {
	return &AIService{
		ctx:        context.Background(),
		config:     cfg,
		ragService: ragService,
	}
}

// initChatModel lazy initializes the ChatModel
func (a *AIService) initChatModel() error {
	if a.chatModel != nil {
		return nil
	}

	cfg := &openai.ChatModelConfig{
		BaseURL: a.config.BaseURL,
		APIKey:  a.config.APIKey,
		Model:   a.config.Model,
	}

	g.Log().Infof(a.ctx, "AI Config: %+v", cfg)

	cm, err := openai.NewChatModel(a.ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}

	a.chatModel = cm
	return nil
}

// StreamResponse streams AI response using eino with optional RAG enhancement
// Returns the retrieved documents that were used to enhance the response
func (a *AIService) StreamResponse(messages []*schema.Message, responseChan chan<- string, enableRAG bool) ([]*schema.Document, error) {
	defer close(responseChan)

	if err := a.initChatModel(); err != nil {
		return nil, err
	}

	// 增强：如果启用了 RAG，检索相关文档并添加到上下文
	var retrievedDocs []*schema.Document
	enhancedMessages := messages
	if enableRAG && a.ragService != nil && a.ragService.IsEnabled() && len(messages) > 0 {
		// 获取最后一条用户消息作为查询
		lastMsg := messages[len(messages)-1]
		if lastMsg.Role == schema.User {
			// 检索文档（只检索一次）
			docs, err := a.ragService.RetrieveDocuments(a.ctx, lastMsg.Content)
			if err == nil && len(docs) > 0 {
				retrievedDocs = docs

				// 自己格式化上下文（避免再次调用 RetrieveWithContext 导致重复检索）
				contextStr := "以下是相关的知识库信息：\n\n"
				for i, doc := range docs {
					contextStr += fmt.Sprintf("%d. [相关度: %.2f] %s\n\n", i+1, doc.Score(), doc.Content)
				}

				systemMsg := &schema.Message{
					Role:    schema.System,
					Content: contextStr,
				}
				// 创建新的消息切片，插入 RAG 上下文
				enhancedMessages = make([]*schema.Message, 0, len(messages)+1)
				enhancedMessages = append(enhancedMessages, systemMsg)
				enhancedMessages = append(enhancedMessages, messages...)

				g.Log().Infof(a.ctx, "RAG: Retrieved %d documents for context,contextStr: %s", len(docs), contextStr)
			}
		}
	}

	streamResult, err := a.chatModel.Stream(a.ctx, enhancedMessages)
	if err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}
	defer streamResult.Close()

	for {
		chunk, err := streamResult.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return retrievedDocs, err
		}
		if chunk.Content != "" {
			responseChan <- chunk.Content
		}
	}

	return retrievedDocs, nil
}
