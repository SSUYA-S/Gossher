package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gossher/internal/inventory"

	"gopkg.in/yaml.v3"
)

type DocumentType = inventory.DocumentType

const (
	TypeConfig     = inventory.TypeConfig
	TypeHost       = inventory.TypeHost
	TypeGroup      = inventory.TypeGroup
	TypeCredential = inventory.TypeCredential
)

// Repository handles reading and writing YAML files with type discrimination.
type Repository struct {
	baseDir string
	mu      sync.RWMutex
}

// Global repository singleton
var (
	globalRepository *Repository
	repoOnce         sync.Once
	repoMutex        sync.RWMutex
)

// ===== Initialization =====

func Init(baseDir string) error {
	var initErr error
	repoOnce.Do(func() {
		if baseDir == "" {
			initErr = fmt.Errorf("base directory cannot be empty")
			return
		}

		if err := os.MkdirAll(baseDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create base directory: %w", err)
			return
		}

		repoMutex.Lock()
		globalRepository = &Repository{
			baseDir: baseDir,
		}
		repoMutex.Unlock()
	})

	return initErr
}

func GetRepository() *Repository {
	repoMutex.RLock()
	defer repoMutex.RUnlock()

	if globalRepository == nil {
		panic("Repository not initialized. Call storage.Init() at application startup.")
	}

	return globalRepository
}

// ===== Core Operations =====

// Write writes a struct to a YAML file (struct already has type field).
func (r *Repository) Write(filename string, v any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	path := filepath.Join(r.baseDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// / Read reads a YAML file and returns the appropriate typed struct.
func (r *Repository) Read(filename string) (DocumentType, any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.baseDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("file not found: %s", filename)
		}
		return "", nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Step 1: Extract type first
	var typeDoc struct {
		Type DocumentType `yaml:"type"`
	}
	if err := yaml.Unmarshal(data, &typeDoc); err != nil {
		return "", nil, fmt.Errorf("failed to extract type: %w", err)
	}

	// Step 2: Create appropriate struct based on type
	var result any
	switch typeDoc.Type {
	case TypeHost:
		result = &inventory.Host{}
	case TypeGroup:
		result = &inventory.Group{}
	case TypeCredential:
		result = &inventory.Credential{}
	case TypeConfig:
		result = &inventory.Config{} // map 대신 Config 구조체
	default:
		return "", nil, fmt.Errorf("unknown document type: %s", typeDoc.Type)
	}

	// Step 3: Unmarshal into the created struct
	if err := yaml.Unmarshal(data, result); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return typeDoc.Type, result, nil
}

// ReadAs reads a YAML file and unmarshals into the provided struct (legacy support).
func (r *Repository) ReadAs(filename string, v any) (DocumentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.baseDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var typeDoc struct {
		Type DocumentType `yaml:"type"`
	}
	if err := yaml.Unmarshal(data, &typeDoc); err != nil {
		return "", fmt.Errorf("failed to extract type: %w", err)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return typeDoc.Type, nil
}

func (r *Repository) Delete(filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := filepath.Join(r.baseDir, filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}

	return nil
}

func (r *Repository) Exists(filename string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.baseDir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// ===== List Operations =====

func (r *Repository) List() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && isYAMLFile(entry.Name()) {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (r *Repository) ListByType(docType DocumentType) ([]string, error) {
	allFiles, err := r.List()
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, filename := range allFiles {
		// 전체 언마샬 대신 타입만 추출
		path := filepath.Join(r.baseDir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			continue // 읽기 실패 시 건너뜀
		}

		var typeDoc struct {
			Type DocumentType `yaml:"type"`
		}
		if err := yaml.Unmarshal(data, &typeDoc); err != nil {
			continue // 파싱 실패 시 건너뜀
		}

		if typeDoc.Type == docType {
			filtered = append(filtered, filename)
		}
	}

	return filtered, nil
}

// ===== Helper Functions =====

func isYAMLFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".yaml" || ext == ".yml"
}

func (r *Repository) GetBaseDir() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.baseDir
}
