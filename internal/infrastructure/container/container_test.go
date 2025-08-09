package container

import (
	"embed"
	"testing"
	"time"
)

//go:embed testdata/*
var testStaticFiles embed.FS

func TestNewContainer(t *testing.T) {
	c := NewContainer(testStaticFiles)

	if c == nil {
		t.Fatal("Expected container to be created")
	}

	if c.Config == nil {
		t.Error("Expected Config to be initialized")
	}

	if c.DocumentRepository == nil {
		t.Error("Expected DocumentRepository to be initialized")
	}

	if c.FileStorageRepository == nil {
		t.Error("Expected FileStorageRepository to be initialized")
	}

	if c.DocumentService == nil {
		t.Error("Expected DocumentService to be initialized")
	}

	if c.ExpirationService == nil {
		t.Error("Expected ExpirationService to be initialized")
	}

	if c.CleanupService == nil {
		t.Error("Expected CleanupService to be initialized")
	}

	if c.UpdateContentHandler == nil {
		t.Error("Expected UpdateContentHandler to be initialized")
	}

	if c.UploadFileHandler == nil {
		t.Error("Expected UploadFileHandler to be initialized")
	}

	if c.DeleteFileHandler == nil {
		t.Error("Expected DeleteFileHandler to be initialized")
	}

	if c.GetContentHandler == nil {
		t.Error("Expected GetContentHandler to be initialized")
	}

	if c.GetFileHandler == nil {
		t.Error("Expected GetFileHandler to be initialized")
	}

	if c.ListFilesHandler == nil {
		t.Error("Expected ListFilesHandler to be initialized")
	}

	if c.DocumentAppService == nil {
		t.Error("Expected DocumentAppService to be initialized")
	}

	if c.DocumentHandler == nil {
		t.Error("Expected DocumentHandler to be initialized")
	}

	if c.FileHandler == nil {
		t.Error("Expected FileHandler to be initialized")
	}

	if c.Router == nil {
		t.Error("Expected Router to be initialized")
	}

	if c.Server == nil {
		t.Error("Expected Server to be initialized")
	}
}

func TestContainer_DependencyWiring(t *testing.T) {
	c := NewContainer(testStaticFiles)

	if c.DocumentService == nil {
		t.Error("DocumentService not initialized")
	}

	if c.ExpirationService == nil {
		t.Error("ExpirationService not initialized")
	}

	if c.CleanupService == nil {
		t.Error("CleanupService not initialized")
	}

	if c.DocumentAppService == nil {
		t.Error("DocumentAppService not initialized")
	}

	if c.Server == nil {
		t.Error("Server not initialized")
	}
	if c.Server.Port() != c.Config.Port {
		t.Errorf("Server port mismatch: expected %s, got %s", c.Config.Port, c.Server.Port())
	}
}

func TestContainer_ConfigurationIntegration(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENABLE_FILE_UPLOADS", "true")
	t.Setenv("CLEANUP_INTERVAL", "10m")
	t.Setenv("MAX_FILE_SIZE", "20971520")

	c := NewContainer(testStaticFiles)

	if c.Config.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", c.Config.Port)
	}

	if !c.Config.EnableFileUploads {
		t.Error("Expected file uploads to be enabled")
	}

	if c.Config.CleanupInterval != 10*time.Minute {
		t.Errorf("Expected cleanup interval 10m, got %v", c.Config.CleanupInterval)
	}

	if c.Config.MaxFileSize != 20971520 {
		t.Errorf("Expected max file size 20971520, got %d", c.Config.MaxFileSize)
	}
}

func TestContainer_RepositoryTypes(t *testing.T) {
	c := NewContainer(testStaticFiles)

	if c.DocumentRepository == nil {
		t.Error("DocumentRepository should not be nil")
	}

	if c.FileStorageRepository == nil {
		t.Error("FileStorageRepository should not be nil")
	}
}

func TestContainer_ServiceLayering(t *testing.T) {
	c := NewContainer(testStaticFiles)

	if c.DocumentService == nil {
		t.Error("Domain DocumentService should be initialized")
	}
	if c.ExpirationService == nil {
		t.Error("Domain ExpirationService should be initialized")
	}

	if c.UpdateContentHandler == nil {
		t.Error("Application UpdateContentHandler should be initialized")
	}
	if c.GetContentHandler == nil {
		t.Error("Application GetContentHandler should be initialized")
	}

	if c.DocumentAppService == nil {
		t.Error("Application DocumentAppService should be initialized")
	}

	if c.DocumentHandler == nil {
		t.Error("Infrastructure DocumentHandler should be initialized")
	}
	if c.FileHandler == nil {
		t.Error("Infrastructure FileHandler should be initialized")
	}

	if c.Router == nil {
		t.Error("Infrastructure Router should be initialized")
	}

	if c.Server == nil {
		t.Error("Infrastructure Server should be initialized")
	}
}

func TestContainer_FileHandlerConfiguration(t *testing.T) {
	t.Run("file uploads enabled", func(t *testing.T) {
		t.Setenv("ENABLE_FILE_UPLOADS", "true")

		c := NewContainer(testStaticFiles)

		if !c.Config.EnableFileUploads {
			t.Error("Expected file uploads to be enabled")
		}

		if c.FileHandler == nil {
			t.Error("FileHandler should be initialized when uploads are enabled")
		}

		if c.UploadFileHandler == nil {
			t.Error("UploadFileHandler should be initialized")
		}
		if c.DeleteFileHandler == nil {
			t.Error("DeleteFileHandler should be initialized")
		}
		if c.GetFileHandler == nil {
			t.Error("GetFileHandler should be initialized")
		}
		if c.ListFilesHandler == nil {
			t.Error("ListFilesHandler should be initialized")
		}
	})

	t.Run("file uploads disabled", func(t *testing.T) {
		t.Setenv("ENABLE_FILE_UPLOADS", "false")

		c := NewContainer(testStaticFiles)

		if c.Config.EnableFileUploads {
			t.Error("Expected file uploads to be disabled")
		}

		if c.FileHandler == nil {
			t.Error("FileHandler should still be initialized even when uploads are disabled")
		}
	})
}

func TestContainer_MultipleConcurrentCreations(t *testing.T) {
	done := make(chan *Container, 10)

	for i := 0; i < 10; i++ {
		go func() {
			c := NewContainer(testStaticFiles)
			done <- c
		}()
	}

	containers := make([]*Container, 10)
	for i := 0; i < 10; i++ {
		containers[i] = <-done
	}

	for i, c := range containers {
		if c == nil {
			t.Errorf("Container %d is nil", i)
			continue
		}

		if c.Config == nil {
			t.Errorf("Container %d: Config is nil", i)
		}
		if c.DocumentRepository == nil {
			t.Errorf("Container %d: DocumentRepository is nil", i)
		}
		if c.Server == nil {
			t.Errorf("Container %d: Server is nil", i)
		}

		for j := i + 1; j < len(containers); j++ {
			if c == containers[j] {
				t.Errorf("Container %d and %d are the same instance", i, j)
			}
			if c.DocumentRepository == containers[j].DocumentRepository {
				t.Errorf("Container %d and %d share the same DocumentRepository", i, j)
			}
		}
	}
}

func TestContainer_CleanupServiceInterval(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{
			name:     "default interval",
			envValue: "",
			expected: 5 * time.Minute,
		},
		{
			name:     "custom interval",
			envValue: "1h",
			expected: 1 * time.Hour,
		},
		{
			name:     "short interval",
			envValue: "30s",
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("CLEANUP_INTERVAL", tt.envValue)
			}

			c := NewContainer(testStaticFiles)

			if c.Config.CleanupInterval != tt.expected {
				t.Errorf("Expected cleanup interval %v, got %v",
					tt.expected, c.Config.CleanupInterval)
			}
		})
	}
}
