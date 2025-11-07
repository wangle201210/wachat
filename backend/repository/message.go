package repository

import (
	"github.com/wangle201210/wachat/backend/model"

	"gorm.io/gorm"
)

// MessageRepository handles message data access
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message
func (r *MessageRepository) Create(msg *model.DBMessage) error {
	return r.db.Create(msg).Error
}

// GetByConversation retrieves all messages for a conversation
func (r *MessageRepository) GetByConversation(conversationID string) ([]*model.DBMessage, error) {
	var messages []*model.DBMessage
	if err := r.db.Where("conversation_id = ?", conversationID).
		Order("timestamp ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// Update updates a message
func (r *MessageRepository) Update(msg *model.DBMessage) error {
	return r.db.Save(msg).Error
}

// DeleteByConversation deletes all messages in a conversation
func (r *MessageRepository) DeleteByConversation(conversationID string) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&model.DBMessage{}).Error
}
