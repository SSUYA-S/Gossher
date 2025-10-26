package inventory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// EntityType represents the type of inventory entity.
type EntityType string

const (
	EntityTypeHost       EntityType = "host"
	EntityTypeCredential EntityType = "credential"
	EntityTypeGroup      EntityType = "group"
)

// EntityHeader contains metadata present in all entity files.
type EntityHeader struct {
	Type EntityType `yaml:"type"`
}

// loadEntity loads a single entity from YAML data.
func loadEntity(data []byte) (Entity, error) {
	// First pass: read type only
	var header EntityHeader
	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("failed to parse entity header: %w", err)
	}

	// Create appropriate entity based on type
	var entity Entity
	switch header.Type {
	case EntityTypeHost:
		h := &Host{}
		if err := yaml.Unmarshal(data, h); err != nil {
			return nil, fmt.Errorf("failed to parse host: %w", err)
		}
		entity = h

	case EntityTypeCredential:
		c := &Credential{}
		if err := yaml.Unmarshal(data, c); err != nil {
			return nil, fmt.Errorf("failed to parse credential: %w", err)
		}
		entity = c

	case EntityTypeGroup:
		g := &Group{}
		if err := yaml.Unmarshal(data, g); err != nil {
			return nil, fmt.Errorf("failed to parse group: %w", err)
		}
		entity = g

	default:
		return nil, fmt.Errorf("unknown entity type: %s", header.Type)
	}

	// Validate entity
	if err := entity.Validate(); err != nil {
		return nil, fmt.Errorf("invalid entity: %w", err)
	}

	return entity, nil
}

// loadEntitiesFromFile loads multiple entities from a YAML file.
func loadEntitiesFromFile(filePath string) ([]Entity, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split by YAML document separator (---)
	documents := splitYAMLDocuments(string(data))

	var entities []Entity
	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		entity, err := loadEntity([]byte(doc))
		if err != nil {
			return nil, fmt.Errorf("failed to load entity %d in %s: %w", i+1, filePath, err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// splitYAMLDocuments splits a YAML string into separate documents.
func splitYAMLDocuments(content string) []string {
	parts := strings.Split(content, "\n---")

	var documents []string
	for i, part := range parts {
		if i == 0 {
			documents = append(documents, part)
		} else {
			documents = append(documents, part)
		}
	}

	return documents
}

// loadAllEntitiesFromDir recursively loads all entities from a directory.
func loadAllEntitiesFromDir(baseDir string) ([]Entity, error) {
	var entities []Entity

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !isYAMLFile(info.Name()) {
			return nil
		}

		// Skip config file
		if info.Name() == "config.yaml" {
			return nil
		}

		fileEntities, err := loadEntitiesFromFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load %s: %v\n", path, err)
			return nil
		}

		entities = append(entities, fileEntities...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return entities, nil
}
