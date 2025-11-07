package service

import (
	"context"
	"fmt"
	"io"

	"github.com/wangle201210/wachat/backend/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// AIService handles AI interactions using eino ChatModel
type AIService struct {
	chatModel *openai.ChatModel
	ctx       context.Context
	config    *config.AIConfig
}

// NewAIService creates AI service with eino ChatModel
func NewAIService(cfg *config.AIConfig) *AIService {
	return &AIService{
		ctx:    context.Background(),
		config: cfg,
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

	fmt.Printf("AI Config: %+v\n", cfg)

	cm, err := openai.NewChatModel(a.ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}

	a.chatModel = cm
	return nil
}

// StreamResponse streams AI response using eino
func (a *AIService) StreamResponse(messages []*schema.Message, responseChan chan<- string) error {
	defer close(responseChan)

	if err := a.initChatModel(); err != nil {
		return err
	}

	streamResult, err := a.chatModel.Stream(a.ctx, messages)
	if err != nil {
		return fmt.Errorf("stream error: %w", err)
	}
	defer streamResult.Close()

	for {
		chunk, err := streamResult.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if chunk.Content != "" {
			responseChan <- chunk.Content
		}
	}

	return nil
}
