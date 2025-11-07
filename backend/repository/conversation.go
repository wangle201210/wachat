package repository

import (
	"github.com/wangle201210/wachat/backend/model"

	"gorm.io/gorm"
)

// ConversationRepository handles conversation data access
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create inserts a new conversation
func (r *ConversationRepository) Create(conv *model.DBConversation) error {
	return r.db.Create(conv).Error
}

// Get retrieves a conversation by ID
func (r *ConversationRepository) Get(id string) (*model.DBConversation, error) {
	var conv model.DBConversation
	if err := r.db.First(&conv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// List returns all conversations ordered by update time
func (r *ConversationRepository) List() ([]*model.DBConversation, error) {
	var convs []*model.DBConversation
	if err := r.db.Order("updated_at DESC").Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

// Update updates a conversation
func (r *ConversationRepository) Update(conv *model.DBConversation) error {
	return r.db.Save(conv).Error
}

// Delete deletes a conversation and its messages
func (r *ConversationRepository) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete messages first
		if err := tx.Where("conversation_id = ?", id).Delete(&model.DBMessage{}).Error; err != nil {
			return err
		}
		// Delete conversation
		return tx.Delete(&model.DBConversation{}, "id = ?", id).Error
	})
}
