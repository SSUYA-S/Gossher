package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"gossher/internal/inventory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (*Repository, string) {
	tmpDir := t.TempDir()
	repo := &Repository{baseDir: tmpDir}
	return repo, tmpDir
}

func TestInit(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Reset global state
		globalRepository = nil
		repoOnce = sync.Once{}

		err := Init(tmpDir)
		require.NoError(t, err)

		repo := GetRepository()
		assert.NotNil(t, repo)
		assert.Equal(t, tmpDir, repo.baseDir)
	})

	t.Run("initialization fails with empty directory", func(t *testing.T) {
		globalRepository = nil
		repoOnce = sync.Once{}

		err := Init("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestGetRepository(t *testing.T) {
	t.Run("panic when not initialized", func(t *testing.T) {
		globalRepository = nil
		repoMutex = sync.RWMutex{}

		assert.Panics(t, func() {
			GetRepository()
		})
	})
}

func TestWrite(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)

	t.Run("save host", func(t *testing.T) {
		host := &inventory.Host{
			Type:    inventory.TypeHost,
			ID:      "host1",
			Name:    "server1",
			Address: "192.168.1.10",
			Port:    22,
		}

		err := repo.Write("host1.yaml", host)
		require.NoError(t, err)

		// Verify file exists
		path := filepath.Join(tmpDir, "host1.yaml")
		_, err = os.Stat(path)
		assert.NoError(t, err)

		// Verify file content
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "type: host")
		assert.Contains(t, string(data), "name: server1")
		assert.Contains(t, string(data), "id: host1")
	})

	t.Run("save group", func(t *testing.T) {
		group := &inventory.Group{
			Type:    inventory.TypeGroup,
			Name:    "webservers",
			HostIDs: []string{"host1", "host2"},
		}

		err := repo.Write("group1.yaml", group)
		require.NoError(t, err)

		path := filepath.Join(tmpDir, "group1.yaml")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "type: group")
		assert.Contains(t, string(data), "name: webservers")
	})

	t.Run("save credential", func(t *testing.T) {
		cred := &inventory.Credential{
			Type: inventory.TypeCredential,
			ID:   "cred1",
			Name: "admin-key",
			User: "admin",
		}

		err := repo.Write("cred1.yaml", cred)
		require.NoError(t, err)

		path := filepath.Join(tmpDir, "cred1.yaml")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "type: credential")
		assert.Contains(t, string(data), "user: admin")
	})

	t.Run("save config", func(t *testing.T) {
		cfg := &inventory.Config{
			Type:           inventory.TypeConfig,
			Theme:          "dark",
			Language:       "en",
			DefaultSSHPort: 22,
		}

		err := repo.Write("config.yaml", cfg)
		require.NoError(t, err)

		path := filepath.Join(tmpDir, "config.yaml")
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "type: config")
		assert.Contains(t, string(data), "theme: dark")
	})
}

func TestRead(t *testing.T) {
	repo, _ := setupTestRepo(t)

	t.Run("read host", func(t *testing.T) {
		// Save first
		original := &inventory.Host{
			Type:    inventory.TypeHost,
			ID:      "host2",
			Name:    "server2",
			Address: "10.0.0.1",
			Port:    2222,
		}
		err := repo.Write("host2.yaml", original)
		require.NoError(t, err)

		// Read
		docType, entity, err := repo.Read("host2.yaml")
		require.NoError(t, err)
		assert.Equal(t, inventory.TypeHost, docType)

		// Type assertion
		loaded, ok := entity.(*inventory.Host)
		require.True(t, ok, "entity should be *inventory.Host")
		assert.Equal(t, original.Type, loaded.Type)
		assert.Equal(t, original.ID, loaded.ID)
		assert.Equal(t, original.Name, loaded.Name)
		assert.Equal(t, original.Address, loaded.Address)
		assert.Equal(t, original.Port, loaded.Port)
	})

	t.Run("read group", func(t *testing.T) {
		original := &inventory.Group{
			Type:    inventory.TypeGroup,
			Name:    "databases",
			HostIDs: []string{"db1", "db2"},
		}
		err := repo.Write("group2.yaml", original)
		require.NoError(t, err)

		docType, entity, err := repo.Read("group2.yaml")
		require.NoError(t, err)
		assert.Equal(t, inventory.TypeGroup, docType)

		loaded, ok := entity.(*inventory.Group)
		require.True(t, ok, "entity should be *inventory.Group")
		assert.Equal(t, original.Type, loaded.Type)
		assert.Equal(t, original.Name, loaded.Name)
		assert.Equal(t, original.HostIDs, loaded.HostIDs)
	})

	t.Run("read credential", func(t *testing.T) {
		original := &inventory.Credential{
			Type: inventory.TypeCredential,
			ID:   "cred2",
			Name: "deploy-key",
			User: "deploy",
		}
		err := repo.Write("cred2.yaml", original)
		require.NoError(t, err)

		docType, entity, err := repo.Read("cred2.yaml")
		require.NoError(t, err)
		assert.Equal(t, inventory.TypeCredential, docType)

		loaded, ok := entity.(*inventory.Credential)
		require.True(t, ok, "entity should be *inventory.Credential")
		assert.Equal(t, original.Type, loaded.Type)
		assert.Equal(t, original.ID, loaded.ID)
		assert.Equal(t, original.User, loaded.User)
	})

	t.Run("read config", func(t *testing.T) {
		original := &inventory.Config{
			Type:           inventory.TypeConfig,
			Theme:          "light",
			Language:       "ko",
			DefaultSSHPort: 2222,
		}
		err := repo.Write("config.yaml", original)
		require.NoError(t, err)

		docType, entity, err := repo.Read("config.yaml")
		require.NoError(t, err)
		assert.Equal(t, inventory.TypeConfig, docType)

		loaded, ok := entity.(*inventory.Config)
		require.True(t, ok, "entity should be *inventory.Config")
		assert.Equal(t, original.Type, loaded.Type)
		assert.Equal(t, original.Theme, loaded.Theme)
		assert.Equal(t, original.Language, loaded.Language)
		assert.Equal(t, original.DefaultSSHPort, loaded.DefaultSSHPort)
	})

	t.Run("fail to read non-existent file", func(t *testing.T) {
		_, _, err := repo.Read("nonexistent.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})

	t.Run("fail to read invalid type", func(t *testing.T) {
		invalidPath := filepath.Join(repo.baseDir, "invalid_type.yaml")
		content := `type: unknown_type
name: test`
		err := os.WriteFile(invalidPath, []byte(content), 0644)
		require.NoError(t, err)

		_, _, err = repo.Read("invalid_type.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown document type")
	})

	t.Run("fail to read file without type field", func(t *testing.T) {
		invalidPath := filepath.Join(repo.baseDir, "no_type.yaml")
		content := `name: test
address: 1.2.3.4`
		err := os.WriteFile(invalidPath, []byte(content), 0644)
		require.NoError(t, err)

		_, _, err = repo.Read("no_type.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown document type")
	})
}

func TestReadAs(t *testing.T) {
	repo, _ := setupTestRepo(t)

	t.Run("read host with ReadAs (backward compatibility)", func(t *testing.T) {
		original := &inventory.Host{
			Type:    inventory.TypeHost,
			ID:      "host3",
			Name:    "server3",
			Address: "192.168.1.20",
			Port:    22,
		}
		err := repo.Write("host3.yaml", original)
		require.NoError(t, err)

		var loaded inventory.Host
		docType, err := repo.ReadAs("host3.yaml", &loaded)
		require.NoError(t, err)
		assert.Equal(t, inventory.TypeHost, docType)
		assert.Equal(t, original.Type, loaded.Type)
		assert.Equal(t, original.Name, loaded.Name)
		assert.Equal(t, original.Address, loaded.Address)
	})

	t.Run("read group with ReadAs", func(t *testing.T) {
		original := &inventory.Group{
			Type:    inventory.TypeGroup,
			Name:    "cache-servers",
			HostIDs: []string{"redis1"},
		}
		err := repo.Write("group3.yaml", original)
		require.NoError(t, err)

		var loaded inventory.Group
		docType, err := repo.ReadAs("group3.yaml", &loaded)
		require.NoError(t, err)
		assert.Equal(t, inventory.TypeGroup, docType)
		assert.Equal(t, original.Name, loaded.Name)
	})
}

func TestList(t *testing.T) {
	repo, _ := setupTestRepo(t)

	t.Run("empty directory", func(t *testing.T) {
		files, err := repo.List()
		require.NoError(t, err)
		assert.Empty(t, files)
	})

	t.Run("list multiple files", func(t *testing.T) {
		// Create files
		repo.Write("host1.yaml", &inventory.Host{
			Type: inventory.TypeHost, ID: "h1", Name: "h1", Address: "1.1.1.1", Port: 22,
		})
		repo.Write("host2.yml", &inventory.Host{
			Type: inventory.TypeHost, ID: "h2", Name: "h2", Address: "2.2.2.2", Port: 22,
		})
		repo.Write("cred1.yaml", &inventory.Credential{
			Type: inventory.TypeCredential, ID: "c1", Name: "c1", User: "user",
		})

		// Create txt file (should be excluded)
		txtPath := filepath.Join(repo.baseDir, "readme.txt")
		os.WriteFile(txtPath, []byte("test"), 0644)

		files, err := repo.List()
		require.NoError(t, err)
		assert.Len(t, files, 3)
		assert.Contains(t, files, "host1.yaml")
		assert.Contains(t, files, "host2.yml")
		assert.Contains(t, files, "cred1.yaml")
		assert.NotContains(t, files, "readme.txt")
	})
}

func TestListByType(t *testing.T) {
	repo, _ := setupTestRepo(t)

	// Create files of various types
	repo.Write("host1.yaml", &inventory.Host{
		Type: inventory.TypeHost, ID: "h1", Name: "h1", Address: "1.1.1.1", Port: 22,
	})
	repo.Write("host2.yaml", &inventory.Host{
		Type: inventory.TypeHost, ID: "h2", Name: "h2", Address: "2.2.2.2", Port: 22,
	})
	repo.Write("cred1.yaml", &inventory.Credential{
		Type: inventory.TypeCredential, ID: "c1", Name: "c1", User: "user",
	})
	repo.Write("group1.yaml", &inventory.Group{
		Type: inventory.TypeGroup, Name: "g1",
	})
	repo.Write("config.yaml", &inventory.Config{
		Type: inventory.TypeConfig, Theme: "dark",
	})

	t.Run("filter host type only", func(t *testing.T) {
		hosts, err := repo.ListByType(inventory.TypeHost)
		require.NoError(t, err)
		assert.Len(t, hosts, 2)
		assert.Contains(t, hosts, "host1.yaml")
		assert.Contains(t, hosts, "host2.yaml")
	})

	t.Run("filter credential type only", func(t *testing.T) {
		creds, err := repo.ListByType(inventory.TypeCredential)
		require.NoError(t, err)
		assert.Len(t, creds, 1)
		assert.Contains(t, creds, "cred1.yaml")
	})

	t.Run("filter group type only", func(t *testing.T) {
		groups, err := repo.ListByType(inventory.TypeGroup)
		require.NoError(t, err)
		assert.Len(t, groups, 1)
		assert.Contains(t, groups, "group1.yaml")
	})

	t.Run("filter config type", func(t *testing.T) {
		configs, err := repo.ListByType(inventory.TypeConfig)
		require.NoError(t, err)
		assert.Len(t, configs, 1)
		assert.Contains(t, configs, "config.yaml")
	})

	t.Run("non-existent type", func(t *testing.T) {
		results, err := repo.ListByType("nonexistent")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("skip invalid YAML files", func(t *testing.T) {
		invalidPath := filepath.Join(repo.baseDir, "invalid.yaml")
		os.WriteFile(invalidPath, []byte("invalid: yaml: content"), 0644)

		hosts, err := repo.ListByType(inventory.TypeHost)
		require.NoError(t, err)
		// Skip invalid files and return only valid ones
		assert.Len(t, hosts, 2)
	})
}

func TestDelete(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)

	t.Run("delete file successfully", func(t *testing.T) {
		// Create file
		host := &inventory.Host{
			Type: inventory.TypeHost, ID: "delete", Name: "delete", Address: "1.2.3.4", Port: 22,
		}
		err := repo.Write("delete_me.yaml", host)
		require.NoError(t, err)

		path := filepath.Join(tmpDir, "delete_me.yaml")
		_, err = os.Stat(path)
		require.NoError(t, err)

		// Delete
		err = repo.Delete("delete_me.yaml")
		require.NoError(t, err)

		// Verify file does not exist
		_, err = os.Stat(path)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("delete non-existent file without error", func(t *testing.T) {
		err := repo.Delete("nonexistent.yaml")
		assert.NoError(t, err)
	})
}

func TestExists(t *testing.T) {
	repo, _ := setupTestRepo(t)

	t.Run("file exists", func(t *testing.T) {
		host := &inventory.Host{
			Type: inventory.TypeHost, ID: "exists", Name: "exists", Address: "1.2.3.4", Port: 22,
		}
		err := repo.Write("exists.yaml", host)
		require.NoError(t, err)

		exists := repo.Exists("exists.yaml")
		assert.True(t, exists)
	})

	t.Run("file does not exist", func(t *testing.T) {
		exists := repo.Exists("not_exists.yaml")
		assert.False(t, exists)
	})
}

func TestGetBaseDir(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)

	baseDir := repo.GetBaseDir()
	assert.Equal(t, tmpDir, baseDir)
}

func TestConcurrency(t *testing.T) {
	repo, _ := setupTestRepo(t)

	t.Run("concurrent writes", func(t *testing.T) {
		const goroutines = 10
		var wg sync.WaitGroup

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				host := &inventory.Host{
					Type:    inventory.TypeHost,
					ID:      fmt.Sprintf("concurrent_%d", idx),
					Name:    fmt.Sprintf("host_%d", idx),
					Address: fmt.Sprintf("10.0.0.%d", idx),
					Port:    22,
				}
				repo.Write(fmt.Sprintf("concurrent_%d.yaml", idx), host)
			}(i)
		}

		wg.Wait()

		// Verify all files created
		files, err := repo.List()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(files), goroutines)
	})

	t.Run("concurrent reads", func(t *testing.T) {
		// Create test file
		host := &inventory.Host{
			Type: inventory.TypeHost, ID: "shared", Name: "shared", Address: "1.1.1.1", Port: 22,
		}
		err := repo.Write("shared.yaml", host)
		require.NoError(t, err)

		const goroutines = 20
		var wg sync.WaitGroup

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := repo.Read("shared.yaml")
				assert.NoError(t, err)
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent read/write", func(t *testing.T) {
		const goroutines = 10
		var wg sync.WaitGroup

		// Write
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				host := &inventory.Host{
					Type:    inventory.TypeHost,
					ID:      fmt.Sprintf("rw_%d", idx),
					Name:    fmt.Sprintf("rw_%d", idx),
					Address: fmt.Sprintf("192.168.1.%d", idx),
					Port:    22,
				}
				repo.Write(fmt.Sprintf("rw_%d.yaml", idx), host)
			}(i)
		}

		// Read
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				repo.Read(fmt.Sprintf("rw_%d.yaml", idx))
			}(i)
		}

		wg.Wait()
	})
}

func TestIsYAMLFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.yaml", true},
		{"test.yml", true},
		{"test.txt", false},
		{"test.json", false},
		{"test", false},
		{".yaml", true},
		{".yml", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isYAMLFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}
