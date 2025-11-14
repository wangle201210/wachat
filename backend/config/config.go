package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"
)

var (
	globalConfig *Config
	configPath   string
	configMutex  sync.RWMutex
	watcher      *fsnotify.Watcher
	reloadChan   chan struct{}
	// 配置变更回调
	onConfigChange func()
	// 标记是否跳过下一次自动重载（用于写入配置文件时避免循环重载）
	skipNextReload  bool
	skipReloadMutex sync.Mutex
)

// Config holds all configuration (GoFrame style)
type Config struct {
	AI       *AIConfig       `json:"ai"`
	Binaries *BinariesConfig `json:"binaries"`
	RAG      *RAGConfig      `json:"rag"`
	Qdrant   *QdrantConfig   `json:"qdrant"`
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
	Address string `json:"address"`
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
	AutoStart            bool          `json:"autoStart"`            // 是否自动启动 RAG 服务器（默认 false）
	TopK                 int           `json:"topK"`                 // 检索返回的文档数量
	DefaultKnowledgeBase string        `json:"defaultKnowledgeBase"` // 默认知识库名称（用于自动 RAG 增强）
	DownloadURL          string        `json:"downloadURL"`          // go-rag 下载地址（GitHub Releases）
	InstallPath          string        `json:"installPath"`          // go-rag 安装路径
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

// QdrantConfig holds Qdrant configuration
type QdrantConfig struct {
	Enabled     bool   `json:"enabled"`     // 是否启用 Qdrant
	AutoStart   bool   `json:"autoStart"`   // 是否自动启动 Qdrant
	Port        int    `json:"port"`        // HTTP 端口（默认 6333）
	GrpcPort    int    `json:"grpcPort"`    // gRPC 端口（默认 6334）
	DownloadURL string `json:"downloadURL"` // Qdrant 下载地址（GitHub Releases）
	InstallPath string `json:"installPath"` // Qdrant 安装路径
}

// IsEnabled returns whether Qdrant is enabled
func (c *QdrantConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

// isDevMode checks if running in development mode (wails dev)
func isDevMode() bool {
	// Check if go.mod exists in current directory (dev mode indicator)
	if cwd, err := os.Getwd(); err == nil {
		goModPath := filepath.Join(cwd, "go.mod")
		if gfile.Exists(goModPath) {
			return true
		}
	}
	return false
}

// findConfigFile finds config.yaml based on running mode
// - Dev mode (wails dev): use project directory config.yaml
// - Production mode (after build): use ~/.wachat/config.yaml
func findConfigFile() string {
	if isDevMode() {
		// Development mode: use project directory config.yaml
		if cwd, err := os.Getwd(); err == nil {
			cfgFile := filepath.Join(cwd, "config.yaml")
			return cfgFile
		}
		return "config.yaml"
	}

	// Production mode: use ~/.wachat/config.yaml
	if homeDir, err := os.UserHomeDir(); err == nil {
		wachatDir := filepath.Join(homeDir, ".wachat")
		cfgFile := filepath.Join(wachatDir, "config.yaml")
		return cfgFile
	}

	// Fallback: current directory
	return "config.yaml"
}

// Load loads configuration using GoFrame
func Load(ctx context.Context) (*Config, error) {
	// Find config file
	configPath = findConfigFile()
	g.Log().Debugf(ctx, "Loading config from: %s", configPath)

	// Check if config file exists
	if !gfile.Exists(configPath) {
		g.Log().Infof(ctx, "Config file not found at %s, creating default config", configPath)

		// In production mode, create default config file
		if !isDevMode() {
			if err := createDefaultConfigFile(ctx, configPath); err != nil {
				g.Log().Warningf(ctx, "Failed to create default config file: %v, using in-memory defaults", err)
				return createDefaultConfig(), nil
			}
			g.Log().Infof(ctx, "Created default config file at %s", configPath)
		} else {
			// In dev mode, just use in-memory defaults (user should copy config.example.yaml)
			g.Log().Warningf(ctx, "Dev mode: using in-memory defaults. Please copy config.example.yaml to config.yaml")
			return createDefaultConfig(), nil
		}
	}
	//
	// // Set config file for GoFrame global instance
	cfgDir := filepath.Dir(configPath)
	g.Log().Debugf(ctx, "Loading cfgDir from: %s", cfgDir)
	//
	// // Create adapter with custom config directory
	adapter, err := gcfg.NewAdapterFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config adapter: %w", err)
	}

	// Set adapter for GoFrame global instance
	// This is important for go-rag to access the config via g.Cfg() and g.DB()
	g.Cfg().SetAdapter(adapter)

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

	// Load Qdrant config
	config.Qdrant = &QdrantConfig{}
	if !cfg.MustGet(ctx, "qdrant").IsNil() {
		if err := cfg.MustGet(ctx, "qdrant").Scan(config.Qdrant); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan qdrant config: %v", err)
		}
	}

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
			Enabled: true,
			TopK:    5,
		},
		Qdrant: &QdrantConfig{
			Enabled: true,
		},
	}
}

// createDefaultConfigFile creates a default config.yaml file at the specified path
func createDefaultConfigFile(ctx context.Context, configPath string) error {
	// Ensure parent directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default configuration YAML content
	defaultYAML := `# wachat Configuration
ai:
    base_url: "https://api.siliconflow.cn/v1"
    api_key: "sk-"
    model: "deepseek-ai/DeepSeek-V3"
binaries:
    enabled: false
    use_embedded: false
    bin_path: "./bin"
    startup_order: []
rag:
    enabled: true
    topK: 5
qdrant:
    enabled: true
server:
    address: ":8000"
logger:
    level: "all"
    stdout: true
    path: "/User/wanna/.wachat/logs"
    file: "{Y-m-d}.log"
`

	// Write config file
	if err := os.WriteFile(configPath, []byte(defaultYAML), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	g.Log().Infof(ctx, "Created default config file at %s", configPath)
	return nil
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
		if cfg.RAG.DownloadURL == "" {
			cfg.RAG.DownloadURL = "https://github.com/wangle201210/go-rag/releases/latest/download"
		}
		if cfg.RAG.InstallPath == "" {
			// 默认安装到用户目录 ~/.wachat/go-rag
			if homeDir, err := os.UserHomeDir(); err == nil {
				cfg.RAG.InstallPath = filepath.Join(homeDir, ".wachat", "go-rag")
			} else {
				cfg.RAG.InstallPath = "./go-rag"
			}
		}
		// Note: Other RAG configs (embedding, rerank, etc.) are managed by go-rag
		// through GoFrame global config, we don't need to set defaults here
	}

	// Qdrant defaults
	if cfg.Qdrant != nil {
		if cfg.Qdrant.Port == 0 {
			cfg.Qdrant.Port = 6333
		}
		if cfg.Qdrant.GrpcPort == 0 {
			cfg.Qdrant.GrpcPort = 6334
		}
		if cfg.Qdrant.DownloadURL == "" {
			cfg.Qdrant.DownloadURL = "https://github.com/qdrant/qdrant/releases/latest/download"
		}
		if cfg.Qdrant.InstallPath == "" {
			// 默认安装到用户目录 ~/.wachat/qdrant
			if homeDir, err := os.UserHomeDir(); err == nil {
				cfg.Qdrant.InstallPath = filepath.Join(homeDir, ".wachat", "qdrant")
			} else {
				cfg.Qdrant.InstallPath = "./qdrant"
			}
		}
	}
}

// Get returns the global config instance (thread-safe)
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

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

// GetQdrantConfig returns Qdrant configuration
func GetQdrantConfig() *QdrantConfig {
	cfg := Get()
	return cfg.Qdrant
}

// SetOnConfigChange sets the callback function to be called when config changes
func SetOnConfigChange(callback func()) {
	onConfigChange = callback
}

// Reload reloads configuration from file
func Reload(ctx context.Context) error {
	// Check if we should skip this reload
	skipReloadMutex.Lock()
	if skipNextReload {
		skipNextReload = false
		skipReloadMutex.Unlock()
		g.Log().Debug(ctx, "Skipping config reload (triggered by internal write)")
		return nil
	}
	skipReloadMutex.Unlock()

	configMutex.Lock()
	defer configMutex.Unlock()

	g.Log().Info(ctx, "Reloading configuration...")

	// Find config file
	newConfigPath := findConfigFile()
	if !gfile.Exists(newConfigPath) {
		return fmt.Errorf("config file not found at %s", newConfigPath)
	}

	// Create new adapter
	adapter, err := gcfg.NewAdapterFile(newConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config adapter: %w", err)
	}

	// Set adapter for GoFrame global instance
	g.Cfg().SetAdapter(adapter)

	// Get the global config instance
	cfg := g.Cfg()

	// Parse new configuration
	newConfig := &Config{}

	// Load AI config
	newConfig.AI = &AIConfig{}
	if !cfg.MustGet(ctx, "ai").IsNil() {
		if err := cfg.MustGet(ctx, "ai").Scan(newConfig.AI); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan ai config: %v", err)
		}
	}

	// Load Binaries config
	newConfig.Binaries = &BinariesConfig{}
	if !cfg.MustGet(ctx, "binaries").IsNil() {
		if err := cfg.MustGet(ctx, "binaries").Scan(newConfig.Binaries); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan binaries config: %v", err)
		}
	}

	// Load RAG config
	newConfig.RAG = loadRAGConfig(ctx, cfg)

	// Load Qdrant config
	newConfig.Qdrant = &QdrantConfig{}
	if !cfg.MustGet(ctx, "qdrant").IsNil() {
		if err := cfg.MustGet(ctx, "qdrant").Scan(newConfig.Qdrant); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to scan qdrant config: %v", err)
		}
	}

	// Apply defaults
	applyDefaults(newConfig)

	// Replace global config
	globalConfig = newConfig
	configPath = newConfigPath

	g.Log().Info(ctx, "Configuration reloaded successfully")

	// Trigger callback if registered
	if onConfigChange != nil {
		go onConfigChange()
	}

	return nil
}

// WatchConfig starts watching config file for changes
func WatchConfig(ctx context.Context) error {
	if watcher != nil {
		return fmt.Errorf("config watcher already started")
	}

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Watch config file
	if err := watcher.Add(configPath); err != nil {
		watcher.Close()
		watcher = nil
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	// Initialize reload channel
	reloadChan = make(chan struct{}, 1)

	g.Log().Infof(ctx, "Started watching config file: %s", configPath)

	// Start watching in background
	go func() {
		// Debounce timer to avoid multiple reloads
		var debounceTimer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Only respond to write and create events
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					g.Log().Debugf(ctx, "Config file changed: %s", event.Name)

					// Debounce: wait 500ms before reloading
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
						select {
						case reloadChan <- struct{}{}:
							if err := Reload(ctx); err != nil {
								g.Log().Errorf(ctx, "Failed to reload config: %v", err)
							}
						default:
							// Reload already pending
						}
					})
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				g.Log().Errorf(ctx, "Config watcher error: %v", err)
			}
		}
	}()

	return nil
}

// StopWatch stops watching config file
func StopWatch() {
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
	if reloadChan != nil {
		close(reloadChan)
		reloadChan = nil
	}
}

// writeYAMLConfig writes RAG settings to config file
func writeYAMLConfig(ctx context.Context, topK int, defaultKnowledgeBase string) error {
	// Read current config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML as generic map
	var configMap map[string]interface{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure rag section exists
	if configMap["rag"] == nil {
		configMap["rag"] = make(map[string]interface{})
	}

	// Update RAG settings
	ragMap, ok := configMap["rag"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("rag config is not a map")
	}

	ragMap["topK"] = topK
	ragMap["defaultKnowledgeBase"] = defaultKnowledgeBase

	// Marshal back to YAML
	newData, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Set skip flag to avoid reload loop
	skipReloadMutex.Lock()
	skipNextReload = true
	skipReloadMutex.Unlock()

	// Write to file
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		skipReloadMutex.Lock()
		skipNextReload = false
		skipReloadMutex.Unlock()
		return fmt.Errorf("failed to write config file: %w", err)
	}

	g.Log().Infof(ctx, "Wrote RAG settings to config file: topK=%d, defaultKnowledgeBase=%s", topK, defaultKnowledgeBase)
	return nil
}

// UpdateRAGSettings updates RAG-specific settings in memory and config file
func UpdateRAGSettings(ctx context.Context, topK int, defaultKnowledgeBase string) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalConfig == nil || globalConfig.RAG == nil {
		return fmt.Errorf("RAG config not initialized")
	}

	// Update in-memory config first
	globalConfig.RAG.TopK = topK
	globalConfig.RAG.DefaultKnowledgeBase = defaultKnowledgeBase

	g.Log().Infof(ctx, "Updated RAG settings in memory: topK=%d, defaultKnowledgeBase=%s", topK, defaultKnowledgeBase)

	// Write to config file for persistence
	if err := writeYAMLConfig(ctx, topK, defaultKnowledgeBase); err != nil {
		g.Log().Warningf(ctx, "Failed to write config file: %v", err)
		// Continue even if file write fails - at least in-memory config is updated
	}

	// Trigger config change callback
	if onConfigChange != nil {
		go onConfigChange()
	}

	return nil
}

// GetRAGSettings returns current RAG settings
func GetRAGSettings() (topK int, defaultKnowledgeBase string) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig != nil && globalConfig.RAG != nil {
		return globalConfig.RAG.TopK, globalConfig.RAG.DefaultKnowledgeBase
	}
	return 5, "" // default values
}

// writeAIConfig writes AI settings to config file
func writeAIConfig(ctx context.Context, baseURL, apiKey, model string) error {
	// Read current config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML as generic map
	var configMap map[string]interface{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure ai section exists
	if configMap["ai"] == nil {
		configMap["ai"] = make(map[string]interface{})
	}

	// Update AI settings
	aiMap, ok := configMap["ai"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("ai config is not a map")
	}

	aiMap["base_url"] = baseURL
	aiMap["api_key"] = apiKey
	aiMap["model"] = model

	// Marshal back to YAML
	newData, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Set skip flag to avoid reload loop
	skipReloadMutex.Lock()
	skipNextReload = true
	skipReloadMutex.Unlock()

	// Write to file
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		skipReloadMutex.Lock()
		skipNextReload = false
		skipReloadMutex.Unlock()
		return fmt.Errorf("failed to write config file: %w", err)
	}

	g.Log().Infof(ctx, "Wrote AI settings to config file: base_url=%s, model=%s", baseURL, model)
	return nil
}

// UpdateAISettings updates AI-specific settings in memory and config file
func UpdateAISettings(ctx context.Context, baseURL, apiKey, model string) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalConfig == nil || globalConfig.AI == nil {
		return fmt.Errorf("AI config not initialized")
	}

	// Update in-memory config first
	globalConfig.AI.BaseURL = baseURL
	globalConfig.AI.APIKey = apiKey
	globalConfig.AI.Model = model

	g.Log().Infof(ctx, "Updated AI settings in memory: base_url=%s, model=%s", baseURL, model)

	// Write to config file for persistence
	if err := writeAIConfig(ctx, baseURL, apiKey, model); err != nil {
		g.Log().Warningf(ctx, "Failed to write config file: %v", err)
		// Continue even if file write fails - at least in-memory config is updated
	}

	// Trigger config change callback
	if onConfigChange != nil {
		go onConfigChange()
	}

	return nil
}

// GetAISettings returns current AI settings
func GetAISettings() (baseURL, apiKey, model string) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig != nil && globalConfig.AI != nil {
		return globalConfig.AI.BaseURL, globalConfig.AI.APIKey, globalConfig.AI.Model
	}
	return "https://api.openai.com/v1", "", "gpt-3.5-turbo" // default values
}

// GetConfigContent reads the entire config file content
func GetConfigContent(ctx context.Context) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}
	return string(data), nil
}

// SaveConfigContent saves the entire config file content
func SaveConfigContent(ctx context.Context, content string) error {
	// Set skip flag to avoid reload loop
	skipReloadMutex.Lock()
	skipNextReload = true
	skipReloadMutex.Unlock()

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		skipReloadMutex.Lock()
		skipNextReload = false
		skipReloadMutex.Unlock()
		return fmt.Errorf("failed to write config file: %w", err)
	}

	g.Log().Infof(ctx, "Wrote config file successfully")

	// Reload configuration to apply changes
	if err := Reload(ctx); err != nil {
		g.Log().Warningf(ctx, "Failed to reload config after save: %v", err)
	}

	return nil
}
