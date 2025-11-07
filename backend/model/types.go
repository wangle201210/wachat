package model

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Messages  []*schema.Message `json:"messages"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// DBConversation represents conversation table in database
type DBConversation struct {
	ID        string `gorm:"primaryKey"`
	Title     string
	CreatedAt int64
	UpdatedAt int64
}

// DBMessage represents message table in database
type DBMessage struct {
	ID             string `gorm:"primaryKey"`
	ConversationID string `gorm:"index"`
	Role           string // user or assistant
	Content        string `gorm:"type:text"`
	Timestamp      int64
	Status         string // sent, pending, error
	ModelName      string
	ModelID        string
	ModelProvider  string

	// Token usage stats
	InputTokens  int
	OutputTokens int
	TotalTokens  int

	ParentID string
}
