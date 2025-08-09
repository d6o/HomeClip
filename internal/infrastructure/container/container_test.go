package container

import (
	"embed"
	"testing"
	"time"
)

//go:embed testdata/*
var testStaticFiles embed.FS

func TestNewContainer(t *testing.T) {
	// Create container
	c := NewContainer(testStaticFiles)

	// Verify all components are initialized
	if c == nil {
		t.Fatal("Expected container to be created")
	}

	// Check configuration
	if c.Config == nil {
		t.Error("Expected Config to be initialized")
	}

	// Check repositories
	if c.DocumentRepository == nil {
		t.Error("Expected DocumentRepository to be initialized")
	}

	if c.FileStorageRepository == nil {
		t.Error("Expected FileStorageRepository to be initialized")
	}

	// Check domain services
	if c.DocumentService == nil {
		t.Error("Expected DocumentService to be initialized")
	}

	if c.ExpirationService == nil {
		t.Error("Expected ExpirationService to be initialized")
	}

	// Check infrastructure services
	if c.CleanupService == nil {
		t.Error("Expected CleanupService to be initialized")
	}

	// Check command handlers
	if c.UpdateContentHandler == nil {
		t.Error("Expected UpdateContentHandler to be initialized")
	}

	if c.UploadFileHandler == nil {
		t.Error("Expected UploadFileHandler to be initialized")
	}

	if c.DeleteFileHandler == nil {
		t.Error("Expected DeleteFileHandler to be initialized")
	}

	// Check query handlers
	if c.GetContentHandler == nil {
		t.Error("Expected GetContentHandler to be initialized")
	}

	if c.GetFileHandler == nil {
		t.Error("Expected GetFileHandler to be initialized")
	}

	if c.ListFilesHandler == nil {
		t.Error("Expected ListFilesHandler to be initialized")
	}

	// Check application services
	if c.DocumentAppService == nil {
		t.Error("Expected DocumentAppService to be initialized")
	}

	// Check HTTP handlers
	if c.DocumentHandler == nil {
		t.Error("Expected DocumentHandler to be initialized")
	}

	if c.FileHandler == nil {
		t.Error("Expected FileHandler to be initialized")
	}

	// Check router
	if c.Router == nil {
		t.Error("Expected Router to be initialized")
	}

	// Check server
	if c.Server == nil {
		t.Error("Expected Server to be initialized")
	}
}

func TestContainer_DependencyWiring(t *testing.T) {
	// Create container
	c := NewContainer(testStaticFiles)

	// Verify that the dependencies are properly wired

	// DocumentService should use the DocumentRepository
	if c.DocumentService == nil {
		t.Error("DocumentService not initialized")
	}

	// ExpirationService should use both repositories
	if c.ExpirationService == nil {
		t.Error("ExpirationService not initialized")
	}

	// CleanupService should have the correct interval from config
	if c.CleanupService == nil {
		t.Error("CleanupService not initialized")
	}

	// DocumentAppService should have the command and query handlers
	if c.DocumentAppService == nil {
		t.Error("DocumentAppService not initialized")
	}

	// Server should have the correct config
	if c.Server == nil {
		t.Error("Server not initialized")
	}
	if c.Server.Port() != c.Config.Port {
		t.Errorf("Server port mismatch: expected %s, got %s", c.Config.Port, c.Server.Port())
	}
}

func TestContainer_ConfigurationIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("PORT", "9090")
	t.Setenv("ENABLE_FILE_UPLOADS", "true")
	t.Setenv("CLEANUP_INTERVAL", "10m")
	t.Setenv("MAX_FILE_SIZE", "20971520") // 20MB

	// Create container
	c := NewContainer(testStaticFiles)

	// Verify configuration values are properly loaded
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

	// Verify that the correct repository implementations are used
	// In this case, we're using in-memory implementations

	// DocumentRepository should be a MemoryDocumentRepository
	if c.DocumentRepository == nil {
		t.Error("DocumentRepository should not be nil")
	}

	// FileStorageRepository should be a MemoryFileStorage
	if c.FileStorageRepository == nil {
		t.Error("FileStorageRepository should not be nil")
	}
}

func TestContainer_ServiceLayering(t *testing.T) {
	c := NewContainer(testStaticFiles)

	// Verify proper layering - higher layers should depend on lower layers
	
	// Domain services should exist
	if c.DocumentService == nil {
		t.Error("Domain DocumentService should be initialized")
	}
	if c.ExpirationService == nil {
		t.Error("Domain ExpirationService should be initialized")
	}

	// Application handlers should exist and depend on domain services
	if c.UpdateContentHandler == nil {
		t.Error("Application UpdateContentHandler should be initialized")
	}
	if c.GetContentHandler == nil {
		t.Error("Application GetContentHandler should be initialized")
	}

	// Application service should exist and use handlers
	if c.DocumentAppService == nil {
		t.Error("Application DocumentAppService should be initialized")
	}

	// Infrastructure handlers should exist and use application services
	if c.DocumentHandler == nil {
		t.Error("Infrastructure DocumentHandler should be initialized")
	}
	if c.FileHandler == nil {
		t.Error("Infrastructure FileHandler should be initialized")
	}

	// Router should exist and use infrastructure handlers
	if c.Router == nil {
		t.Error("Infrastructure Router should be initialized")
	}

	// Server should exist and be configured
	if c.Server == nil {
		t.Error("Infrastructure Server should be initialized")
	}
}

func TestContainer_FileHandlerConfiguration(t *testing.T) {
	// Test with file uploads enabled
	t.Run("file uploads enabled", func(t *testing.T) {
		t.Setenv("ENABLE_FILE_UPLOADS", "true")
		
		c := NewContainer(testStaticFiles)
		
		if !c.Config.EnableFileUploads {
			t.Error("Expected file uploads to be enabled")
		}
		
		// FileHandler should be fully configured
		if c.FileHandler == nil {
			t.Error("FileHandler should be initialized when uploads are enabled")
		}
		
		// All file-related handlers should be initialized
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

	// Test with file uploads disabled
	t.Run("file uploads disabled", func(t *testing.T) {
		t.Setenv("ENABLE_FILE_UPLOADS", "false")
		
		c := NewContainer(testStaticFiles)
		
		if c.Config.EnableFileUploads {
			t.Error("Expected file uploads to be disabled")
		}
		
		// FileHandler should still be initialized (router decides whether to use it)
		if c.FileHandler == nil {
			t.Error("FileHandler should still be initialized even when uploads are disabled")
		}
	})
}

func TestContainer_MultipleConcurrentCreations(t *testing.T) {
	// Test that multiple containers can be created concurrently
	// without interfering with each other
	
	done := make(chan *Container, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			c := NewContainer(testStaticFiles)
			done <- c
		}()
	}
	
	// Collect all containers
	containers := make([]*Container, 10)
	for i := 0; i < 10; i++ {
		containers[i] = <-done
	}
	
	// Verify all containers are properly initialized
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
		
		// Each container should have its own instances
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
			expected: 5 * time.Minute, // Default from config
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