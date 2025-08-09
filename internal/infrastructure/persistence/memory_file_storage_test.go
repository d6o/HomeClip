package persistence

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/d6o/homeclip/internal/domain/entities"
)

func TestMemoryFileStorage_Store_And_Retrieve(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	attachmentID := entities.AttachmentID("test-attachment")
	content := []byte("test file content")
	
	// Store file
	err := storage.Store(ctx, attachmentID, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}
	
	// Retrieve file
	reader, err := storage.Retrieve(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to retrieve file: %v", err)
	}
	defer reader.Close()
	
	// Read and verify content
	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}
	
	if !bytes.Equal(retrieved, content) {
		t.Errorf("Expected content %s, got %s", content, retrieved)
	}
}

func TestMemoryFileStorage_Retrieve_NotFound(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	// Try to retrieve non-existent file
	_, err := storage.Retrieve(ctx, entities.AttachmentID("non-existent"))
	
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestMemoryFileStorage_Delete(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	attachmentID := entities.AttachmentID("test-attachment")
	content := []byte("test content")
	
	// Store file
	err := storage.Store(ctx, attachmentID, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}
	
	// Verify it exists
	exists, err := storage.Exists(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected file to exist")
	}
	
	// Delete file
	err = storage.Delete(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}
	
	// Verify it's deleted
	exists, err = storage.Exists(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected file to be deleted")
	}
	
	// Try to retrieve deleted file
	_, err = storage.Retrieve(ctx, attachmentID)
	if err == nil {
		t.Error("Expected error when retrieving deleted file")
	}
}

func TestMemoryFileStorage_Delete_NonExistent(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	// Delete non-existent file should not error
	err := storage.Delete(ctx, entities.AttachmentID("non-existent"))
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent file, got %v", err)
	}
}

func TestMemoryFileStorage_Exists(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	attachmentID := entities.AttachmentID("test-attachment")
	
	// Check non-existent file
	exists, err := storage.Exists(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected file to not exist")
	}
	
	// Store file
	err = storage.Store(ctx, attachmentID, bytes.NewReader([]byte("content")))
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}
	
	// Check existing file
	exists, err = storage.Exists(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected file to exist")
	}
}

func TestMemoryFileStorage_Overwrite(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	attachmentID := entities.AttachmentID("test-attachment")
	originalContent := []byte("original content")
	newContent := []byte("new content")
	
	// Store original file
	err := storage.Store(ctx, attachmentID, bytes.NewReader(originalContent))
	if err != nil {
		t.Fatalf("Failed to store original file: %v", err)
	}
	
	// Overwrite with new content
	err = storage.Store(ctx, attachmentID, bytes.NewReader(newContent))
	if err != nil {
		t.Fatalf("Failed to overwrite file: %v", err)
	}
	
	// Retrieve and verify new content
	reader, err := storage.Retrieve(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to retrieve file: %v", err)
	}
	defer reader.Close()
	
	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}
	
	if !bytes.Equal(retrieved, newContent) {
		t.Errorf("Expected new content %s, got %s", newContent, retrieved)
	}
}

func TestMemoryFileStorage_LargeFile(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	attachmentID := entities.AttachmentID("large-file")
	
	// Create a 1MB file
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	
	// Store large file
	err := storage.Store(ctx, attachmentID, bytes.NewReader(largeContent))
	if err != nil {
		t.Fatalf("Failed to store large file: %v", err)
	}
	
	// Retrieve and verify
	reader, err := storage.Retrieve(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to retrieve large file: %v", err)
	}
	defer reader.Close()
	
	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read large file content: %v", err)
	}
	
	if len(retrieved) != len(largeContent) {
		t.Errorf("Expected %d bytes, got %d", len(largeContent), len(retrieved))
	}
	
	if !bytes.Equal(retrieved, largeContent) {
		t.Error("Large file content mismatch")
	}
}

func TestMemoryFileStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	// Test concurrent stores
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			attachmentID := entities.AttachmentID(string(rune('a' + id)))
			content := []byte(string(rune('A' + id)))
			
			err := storage.Store(ctx, attachmentID, bytes.NewReader(content))
			if err != nil {
				t.Errorf("Failed to store file %v: %v", attachmentID, err)
			}
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all files were stored correctly
	for i := 0; i < 10; i++ {
		attachmentID := entities.AttachmentID(string(rune('a' + i)))
		expectedContent := []byte(string(rune('A' + i)))
		
		reader, err := storage.Retrieve(ctx, attachmentID)
		if err != nil {
			t.Errorf("Failed to retrieve file %v: %v", attachmentID, err)
			continue
		}
		
		retrieved, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Errorf("Failed to read file %v: %v", attachmentID, err)
			continue
		}
		
		if !bytes.Equal(retrieved, expectedContent) {
			t.Errorf("File %v content mismatch", attachmentID)
		}
	}
}

func TestMemoryFileStorage_EmptyFile(t *testing.T) {
	storage := NewMemoryFileStorage()
	ctx := context.Background()
	
	attachmentID := entities.AttachmentID("empty-file")
	emptyContent := []byte{}
	
	// Store empty file
	err := storage.Store(ctx, attachmentID, bytes.NewReader(emptyContent))
	if err != nil {
		t.Fatalf("Failed to store empty file: %v", err)
	}
	
	// Retrieve and verify
	reader, err := storage.Retrieve(ctx, attachmentID)
	if err != nil {
		t.Fatalf("Failed to retrieve empty file: %v", err)
	}
	defer reader.Close()
	
	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read empty file: %v", err)
	}
	
	if len(retrieved) != 0 {
		t.Errorf("Expected empty file, got %d bytes", len(retrieved))
	}
}