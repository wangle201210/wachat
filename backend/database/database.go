package database

import (
	"os"
	"path/filepath"

	"github.com/wangle201210/wachat/backend/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps gorm.DB
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates database connection and runs migrations
func NewDatabase() (*Database, error) {
	// Get user home dir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create data directory
	dataDir := filepath.Join(homeDir, ".wachat")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// Open database
	dbPath := filepath.Join(dataDir, "chat.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	if err := db.AutoMigrate(&model.DBConversation{}, &model.DBMessage{}); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}
