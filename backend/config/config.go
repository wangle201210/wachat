package config

import "os"

// GetEnv gets environment variable with default value
func GetEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// AIConfig holds AI service configuration
type AIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

// GetAIConfig returns AI configuration from environment
func GetAIConfig() *AIConfig {
	return &AIConfig{
		BaseURL: GetEnv("OPENAI_API_URL", "https://api.openai.com/v1"),
		APIKey:  GetEnv("OPENAI_API_KEY", ""),
		Model:   GetEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
	}
}
