package inventory

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config holds application-wide configuration.
type Config struct {
	Type           DocumentType `yaml:"type"`
	DataDir        string       `yaml:"data_dir"`
	Theme          string       `yaml:"theme"`
	Language       string       `yaml:"language"`
	DefaultSSHPort int          `yaml:"default_ssh_port"`
	SSHTimeout     int          `yaml:"ssh_timeout"`

	// Runtime - not saved
	BaseDir    string `yaml:"-"`
	ConfigPath string `yaml:"-"`
}

// Global configuration singleton
var (
	globalConfig *Config
	configMutex  sync.RWMutex
)

// Default returns the default configuration.
func Default() *Config {
	baseDir := defaultBaseDir()

	return &Config{
		Type:           TypeConfig,
		DataDir:        baseDir,
		Theme:          "light",
		Language:       "en",
		DefaultSSHPort: 22,
		SSHTimeout:     30,
		BaseDir:        baseDir,
		ConfigPath:     filepath.Join(baseDir, "config.yaml"),
	}
}

// MustLoad loads configuration from file, creating default if not exists.
func MustLoad() {
	if err := Load(); err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
}

// Load loads configuration from file, or creates default if not exists.
func Load() error {
	baseDir := defaultBaseDir()
	configPath := filepath.Join(baseDir, "config.yaml")

	var cfg *Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return fmt.Errorf("failed to create base directory: %w", err)
		}

		cfg = Default()
		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save default config: %w", err)
		}
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		cfg = &Config{
			BaseDir: baseDir,
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	configMutex.Lock()
	globalConfig = cfg
	configMutex.Unlock()

	return nil
}

// saveConfig saves the configuration to file.
func saveConfig(cfg *Config) error {
	if err := os.MkdirAll(cfg.BaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfg.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Save saves the current global configuration to file.
func Save() error {
	configMutex.RLock()
	cfg := globalConfig
	configMutex.RUnlock()

	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	return saveConfig(cfg)
}

// ===== Getters (Thread-safe read) =====

// GetDataDir returns the configured data directory.
func GetDataDir() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded. Call config.MustLoad() at application startup.")
	}

	if globalConfig.DataDir == "" {
		return globalConfig.BaseDir
	}

	if !filepath.IsAbs(globalConfig.DataDir) {
		return filepath.Join(globalConfig.BaseDir, globalConfig.DataDir)
	}

	return globalConfig.DataDir
}

// GetTheme returns the configured theme.
func GetTheme() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.Theme
}

// GetLanguage returns the configured language.
func GetLanguage() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.Language
}

// GetDefaultSSHPort returns the default SSH port.
func GetDefaultSSHPort() int {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.DefaultSSHPort
}

// GetSSHTimeout returns the SSH timeout in seconds.
func GetSSHTimeout() int {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.SSHTimeout
}

// ===== Setters =====

// SetDataDir updates the data directory and saves the config.
func SetDataDir(dir string) error {
	configMutex.Lock()
	if globalConfig == nil {
		configMutex.Unlock()
		return fmt.Errorf("config not loaded")
	}
	globalConfig.DataDir = dir
	configMutex.Unlock()

	return Save()
}

// SetTheme updates the theme and saves the config.
func SetTheme(theme string) error {
	configMutex.Lock()
	if globalConfig == nil {
		configMutex.Unlock()
		return fmt.Errorf("config not loaded")
	}
	globalConfig.Theme = theme
	configMutex.Unlock()

	return Save()
}

// SetLanguage updates the language and saves the config.
func SetLanguage(lang string) error {
	configMutex.Lock()
	if globalConfig == nil {
		configMutex.Unlock()
		return fmt.Errorf("config not loaded")
	}
	globalConfig.Language = lang
	configMutex.Unlock()

	return Save()
}

// SetDefaultSSHPort updates the default SSH port and saves the config.
func SetDefaultSSHPort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}

	configMutex.Lock()
	if globalConfig == nil {
		configMutex.Unlock()
		return fmt.Errorf("config not loaded")
	}
	globalConfig.DefaultSSHPort = port
	configMutex.Unlock()

	return Save()
}

// SetSSHTimeout updates the SSH timeout and saves the config.
func SetSSHTimeout(timeout int) error {
	if timeout <= 0 {
		return fmt.Errorf("invalid timeout: %d", timeout)
	}

	configMutex.Lock()
	if globalConfig == nil {
		configMutex.Unlock()
		return fmt.Errorf("config not loaded")
	}
	globalConfig.SSHTimeout = timeout
	configMutex.Unlock()

	return Save()
}

// ===== Batch Update =====

// Update allows updating multiple fields atomically.
func Update(fn func(*ConfigEditor) error) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	editor := &ConfigEditor{cfg: globalConfig}

	if err := fn(editor); err != nil {
		return err
	}

	return saveConfig(globalConfig)
}

// ConfigEditor provides controlled write access to config fields.
type ConfigEditor struct {
	cfg *Config
}

// SetDataDir sets the data directory.
func (e *ConfigEditor) SetDataDir(dir string) {
	e.cfg.DataDir = dir
}

// SetTheme sets the theme.
func (e *ConfigEditor) SetTheme(theme string) {
	e.cfg.Theme = theme
}

// SetLanguage sets the language.
func (e *ConfigEditor) SetLanguage(lang string) {
	e.cfg.Language = lang
}

// SetDefaultSSHPort sets the default SSH port.
func (e *ConfigEditor) SetDefaultSSHPort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	e.cfg.DefaultSSHPort = port
	return nil
}

// SetSSHTimeout sets the SSH timeout.
func (e *ConfigEditor) SetSSHTimeout(timeout int) error {
	if timeout <= 0 {
		return fmt.Errorf("invalid timeout: %d", timeout)
	}
	e.cfg.SSHTimeout = timeout
	return nil
}

// ===== Helper Functions =====

// defaultBaseDir returns the default base directory (~/.gossher).
func defaultBaseDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".gossher"
	}
	return filepath.Join(homeDir, ".gossher")
}

// ConfigSnapshot represents a read-only snapshot of configuration.
type ConfigSnapshot struct {
	DataDir        string
	Theme          string
	Language       string
	DefaultSSHPort int
	SSHTimeout     int
}

// GetSnapshot returns a read-only copy of the current configuration.
func GetSnapshot() ConfigSnapshot {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}

	return ConfigSnapshot{
		DataDir:        globalConfig.DataDir,
		Theme:          globalConfig.Theme,
		Language:       globalConfig.Language,
		DefaultSSHPort: globalConfig.DefaultSSHPort,
		SSHTimeout:     globalConfig.SSHTimeout,
	}
}
