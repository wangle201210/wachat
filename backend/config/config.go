package config

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
)

var (
	globalConfig *Config
	configPath   string
)

// Config holds all configuration (GoFrame style)
type Config struct {
	AI       *AIConfig       `json:"ai"`
	Binaries *BinariesConfig `json:"binaries"`
	RAG      *RAGConfig      `json:"rag"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// BinariesConfig holds binary manager configuration
type BinariesConfig struct {
	Enabled      bool     `json:"enabled"`
	UseEmbedded  bool     `json:"use_embedded"`
	BinPath      string   `json:"bin_path"`
	StartupOrder []string `json:"startup_order"`
}

// IsEnabled returns whether binary manager is enabled
func (c *BinariesConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

// IsUseEmbedded returns whether to use embedded mode
func (c *BinariesConfig) IsUseEmbedded() bool {
	return c != nil && c.UseEmbedded
}

// GetBinPath returns the bin directory path
func (c *BinariesConfig) GetBinPath() string {
	if c == nil {
		return ""
	}
	return c.BinPath
}

// GetStartupOrder returns the startup order list
func (c *BinariesConfig) GetStartupOrder() []string {
	if c == nil {
		return nil
	}
	return c.StartupOrder
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address     string `json:"address"`
	OpenapiPath string `json:"openapiPath"`
	SwaggerPath string `json:"swaggerPath"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string `json:"level"`
	Stdout bool   `json:"stdout"`
	Path   string `json:"path"`
	File   string `json:"file"`
}

// DatabaseDefaultConfig holds default database configuration
type DatabaseDefaultConfig struct {
	Host    string `json:"host"`
	Port    string `json:"port"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Charset string `json:"charset"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Default *DatabaseDefaultConfig `json:"default"`
}

// ESConfig holds Elasticsearch configuration
type ESConfig struct {
	Address   string `json:"address"`
	IndexName string `json:"indexName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// ModelConfig holds model API configuration (reused for embedding, rerank, etc.)
type ModelConfig struct {
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
}

// RAGConfig holds RAG configuration for wachat
// Note: go-rag server reads its own config (server, database, es, embedding, etc.)
// from GoFrame global config (g.Cfg()), we don't need to load them here
type RAGConfig struct {
	Enabled              bool          `json:"enabled"`              // wailsChat 控制：是否启用 RAG 功能
	TopK                 int           `json:"topK"`                 // 检索返回的文档数量
	DefaultKnowledgeBase string        `json:"defaultKnowledgeBase"` // 默认知识库名称（用于自动 RAG 增强）
	Server               *ServerConfig `json:"server"`               // go-rag 服务器配置（用于判断是否启动服务器和构建 HTTP 请求）
}

// IsEnabled returns whether RAG is enabled
func (c *RAGConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

// IsServerEnabled returns whether go-rag server should be started
func (c *RAGConfig) IsServerEnabled() bool {
	return c != nil && c.Server != nil && c.Server.Address != ""
}

// findConfigFile tries to find config.yaml in multiple locations
func findConfigFile() string {
	// Priority 1: Environment variable WACHAT_CONFIG_PATH
	if configPath := os.Getenv("WACHAT_CONFIG_PATH"); configPath != "" {
		cfgFile := filepath.Join(configPath, "config.yaml")
		if gfile.Exists(cfgFile) {
			return cfgFile
		}
	}

	// Priority 2: Current working directory
	if cwd, err := os.Getwd(); err == nil {
		cfgFile := filepath.Join(cwd, "config.yaml")
		if gfile.Exists(cfgFile) {
			return cfgFile
		}
	}

	// Priority 3: Executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		if realPath, err := filepath.EvalSymlinks(execPath); err == nil {
			execDir = filepath.Dir(realPath)
		}
		cfgFile := filepath.Join(execDir, "config.yaml")
		if gfile.Exists(cfgFile) {
			return cfgFile
		}
	}

	// Priority 4: User home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		configHome := filepath.Join(homeDir, ".config", "wachat")
		cfgFile := filepath.Join(configHome, "config.yaml")
		if gfile.Exists(cfgFile) {
			return cfgFile
		}
	}

	// Default: current directory
	return "config.yaml"
}

// Load loads configuration using GoFrame
func Load(ctx context.Context) (*Config, error) {
	// // Find config file
	// configPath = findConfigFile()
	// log.Printf("Loading config from: %s", configPath)
	//
	// // Check if config file exists
	// if !gfile.Exists(configPath) {
	// 	log.Printf("Warning: config file not found at %s, using defaults", configPath)
	// 	return createDefaultConfig(), nil
	// }
	//
	// // Set config file for GoFrame global instance
	// cfgDir := filepath.Dir(configPath)
	//
	// // Create adapter with custom config directory
	// adapter, err := gcfg.NewAdapterFile(cfgDir)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create config adapter: %w", err)
	// }
	//
	// // Set adapter for GoFrame global instance
	// // This is important for go-rag to access the config via g.Cfg() and g.DB()
	// g.Cfg().SetAdapter(adapter)

	// Get the global config instance
	cfg := g.Cfg()

	// Parse configuration
	config := &Config{}

	// Load AI config
	config.AI = &AIConfig{}
	if !cfg.MustGet(ctx, "ai").IsNil() {
		if err := cfg.MustGet(ctx, "ai").Scan(config.AI); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan ai config: %v", err)
		}
	}

	// Load Binaries config
	config.Binaries = &BinariesConfig{}
	if !cfg.MustGet(ctx, "binaries").IsNil() {
		if err := cfg.MustGet(ctx, "binaries").Scan(config.Binaries); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan binaries config: %v", err)
		}
	}

	// Load RAG config (go-rag compatible)
	config.RAG = loadRAGConfig(ctx, cfg)

	// Apply defaults
	applyDefaults(config)

	globalConfig = config
	return config, nil
}

// loadRAGConfig loads RAG configuration
// Note: Only load wachat-specific config (enabled, topK, server address)
// go-rag server will read its own config from GoFrame global config (g.Cfg())
func loadRAGConfig(ctx context.Context, cfg *gcfg.Config) *RAGConfig {
	ragCfg := &RAGConfig{}

	// Load rag section (enabled, topK)
	if !cfg.MustGet(ctx, "rag").IsNil() {
		if err := cfg.MustGet(ctx, "rag").Scan(ragCfg); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan rag config: %v", err)
		}
	}

	// Load server config (for determining if go-rag server should start)
	if !cfg.MustGet(ctx, "server").IsNil() {
		ragCfg.Server = &ServerConfig{}
		if err := cfg.MustGet(ctx, "server").Scan(ragCfg.Server); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan server config: %v", err)
		}
	}

	return ragCfg
}

// createDefaultConfig creates a default configuration
func createDefaultConfig() *Config {
	return &Config{
		AI: &AIConfig{
			BaseURL: "https://api.openai.com/v1",
			Model:   "gpt-3.5-turbo",
		},
		Binaries: &BinariesConfig{
			Enabled:     false,
			UseEmbedded: false,
			BinPath:     "./bin",
		},
		RAG: &RAGConfig{
			Enabled: false,
			TopK:    5,
		},
	}
}

// applyDefaults applies default values to config
func applyDefaults(cfg *Config) {
	// AI defaults
	if cfg.AI.BaseURL == "" {
		cfg.AI.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.AI.Model == "" {
		cfg.AI.Model = "gpt-3.5-turbo"
	}

	// Binaries defaults
	if cfg.Binaries.BinPath == "" {
		cfg.Binaries.BinPath = "./bin"
	}

	// RAG defaults
	if cfg.RAG != nil {
		if cfg.RAG.TopK == 0 {
			cfg.RAG.TopK = 5
		}
		// Note: Other RAG configs (embedding, rerank, etc.) are managed by go-rag
		// through GoFrame global config, we don't need to set defaults here
	}
}

// Get returns the global config instance
func Get() *Config {
	if globalConfig == nil {
		g.Log().Fatal(context.Background(), "Config not loaded. Call config.Load() first")
	}
	return globalConfig
}

// GetAIConfig returns AI configuration
func GetAIConfig() *AIConfig {
	cfg := Get()
	return cfg.AI
}

// GetRAGConfig returns RAG configuration
func GetRAGConfig() *RAGConfig {
	cfg := Get()
	return cfg.RAG
}
