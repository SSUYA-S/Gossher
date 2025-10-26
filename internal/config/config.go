package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config holds application-wide configuration.
type Config struct {
	dataDir string `yaml:"data_dir"` // Base directory for inventory data

	theme    string `yaml:"theme"`
	language string `yaml:"language"`

	defaultSSHPort int `yaml:"default_ssh_port"`
	sshTimeout     int `yaml:"ssh_timeout"`

	// Runtime - not saved
	baseDir    string `yaml:"-"` // Base directory (.gossher)
	configPath string `yaml:"-"` // Path to config.yaml file
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
		dataDir:        baseDir,
		theme:          "light",
		language:       "en",
		defaultSSHPort: 22,
		sshTimeout:     30,
		baseDir:        baseDir,
		configPath:     filepath.Join(baseDir, "config.yaml"),
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
			baseDir: baseDir,
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
	if err := os.MkdirAll(cfg.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfg.configPath, data, 0644); err != nil {
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

	if globalConfig.dataDir == "" {
		return globalConfig.baseDir
	}

	if !filepath.IsAbs(globalConfig.dataDir) {
		return filepath.Join(globalConfig.baseDir, globalConfig.dataDir)
	}

	return globalConfig.dataDir
}

// GetTheme returns the configured theme.
func GetTheme() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.theme
}

// GetLanguage returns the configured language.
func GetLanguage() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.language
}

// GetDefaultSSHPort returns the default SSH port.
func GetDefaultSSHPort() int {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.defaultSSHPort
}

// GetSSHTimeout returns the SSH timeout in seconds.
func GetSSHTimeout() int {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		panic("Config not loaded")
	}
	return globalConfig.sshTimeout
}

// ===== Setters (Thread-safe write) =====

// SetDataDir updates the data directory and saves the config.
func SetDataDir(dir string) error {
	configMutex.Lock()
	if globalConfig == nil {
		configMutex.Unlock()
		return fmt.Errorf("config not loaded")
	}
	globalConfig.dataDir = dir
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
	globalConfig.theme = theme
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
	globalConfig.language = lang
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
	globalConfig.defaultSSHPort = port
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
	globalConfig.sshTimeout = timeout
	configMutex.Unlock()

	return Save()
}

// ===== Batch Update =====

// Update allows updating multiple fields atomically.
func Update(fn func(*configEditor) error) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	// Create editor wrapper
	editor := &configEditor{cfg: globalConfig}

	// Apply updates
	if err := fn(editor); err != nil {
		return err
	}

	// Save after updates
	return saveConfig(globalConfig)
}

// configEditor provides controlled write access to config fields.
type configEditor struct {
	cfg *Config
}

// SetDataDir sets the data directory.
func (e *configEditor) SetDataDir(dir string) {
	e.cfg.dataDir = dir
}

// SetTheme sets the theme.
func (e *configEditor) SetTheme(theme string) {
	e.cfg.theme = theme
}

// SetLanguage sets the language.
func (e *configEditor) SetLanguage(lang string) {
	e.cfg.language = lang
}

// SetDefaultSSHPort sets the default SSH port.
func (e *configEditor) SetDefaultSSHPort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	e.cfg.defaultSSHPort = port
	return nil
}

// SetSSHTimeout sets the SSH timeout.
func (e *configEditor) SetSSHTimeout(timeout int) error {
	if timeout <= 0 {
		return fmt.Errorf("invalid timeout: %d", timeout)
	}
	e.cfg.sshTimeout = timeout
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
		DataDir:        globalConfig.dataDir,
		Theme:          globalConfig.theme,
		Language:       globalConfig.language,
		DefaultSSHPort: globalConfig.defaultSSHPort,
		SSHTimeout:     globalConfig.sshTimeout,
	}
}
