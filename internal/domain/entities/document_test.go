package entities

import (
	"testing"
	"time"

	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestNewDocument(t *testing.T) {
	id := DocumentID("test-doc-id")
	doc := NewDocument(id)

	if doc.ID() != id {
		t.Errorf("Expected ID %v, got %v", id, doc.ID())
	}

	if !doc.Content().IsEmpty() {
		t.Errorf("Expected empty content, got %v", doc.Content().Value())
	}

	if doc.AttachmentCount() != 0 {
		t.Errorf("Expected no attachments, got %v", doc.AttachmentCount())
	}

	if doc.LastUpdated().IsZero() {
		t.Error("Expected LastUpdated to be set")
	}

	if doc.Version() != 1 {
		t.Errorf("Expected version 1, got %v", doc.Version())
	}

	if doc.IsExpired() {
		t.Error("Expected new document to not be expired")
	}
}

func TestDocument_UpdateContent(t *testing.T) {
	doc := NewDocument("test-doc")
	originalUpdatedAt := doc.LastUpdated()
	originalVersion := doc.Version()

	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	content, err := valueobjects.NewContent("test content")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}

	err = doc.UpdateContent(content)
	if err != nil {
		t.Fatalf("Failed to update content: %v", err)
	}

	if doc.Content().Value() != "test content" {
		t.Errorf("Expected content 'test content', got %v", doc.Content().Value())
	}

	if !doc.LastUpdated().After(originalUpdatedAt) {
		t.Error("Expected LastUpdated to be updated")
	}

	if doc.Version() != originalVersion+1 {
		t.Errorf("Expected version to increment, got %v", doc.Version())
	}
}

func TestDocument_AddAttachment(t *testing.T) {
	doc := NewDocument("test-doc")
	
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)

	attachment := NewAttachment("attachment-1", doc.ID(), fileName, mimeType, fileSize)
	err := doc.AddAttachment(attachment)
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	if doc.AttachmentCount() != 1 {
		t.Errorf("Expected 1 attachment, got %v", doc.AttachmentCount())
	}

	retrieved, err := doc.GetAttachment("attachment-1")
	if err != nil {
		t.Fatalf("Failed to get attachment: %v", err)
	}

	if retrieved.ID() != "attachment-1" {
		t.Errorf("Expected attachment ID 'attachment-1', got %v", retrieved.ID())
	}
}

func TestDocument_AddAttachment_Duplicate(t *testing.T) {
	doc := NewDocument("test-doc")
	
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)

	attachment := NewAttachment("attachment-1", doc.ID(), fileName, mimeType, fileSize)
	err := doc.AddAttachment(attachment)
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	// Try to add the same attachment again
	err = doc.AddAttachment(attachment)
	if err != ErrDuplicateAttachment {
		t.Errorf("Expected ErrDuplicateAttachment, got %v", err)
	}
}

func TestDocument_RemoveAttachment(t *testing.T) {
	doc := NewDocument("test-doc")
	
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)

	attachment := NewAttachment("attachment-1", doc.ID(), fileName, mimeType, fileSize)
	err := doc.AddAttachment(attachment)
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	err = doc.RemoveAttachment("attachment-1")
	if err != nil {
		t.Fatalf("Failed to remove attachment: %v", err)
	}

	if doc.AttachmentCount() != 0 {
		t.Errorf("Expected 0 attachments, got %v", doc.AttachmentCount())
	}
}

func TestDocument_RemoveAttachment_NotFound(t *testing.T) {
	doc := NewDocument("test-doc")
	
	err := doc.RemoveAttachment("non-existent")
	if err != ErrAttachmentNotFound {
		t.Errorf("Expected ErrAttachmentNotFound, got %v", err)
	}
}

func TestDocument_GetAttachment_NotFound(t *testing.T) {
	doc := NewDocument("test-doc")
	
	_, err := doc.GetAttachment("non-existent")
	if err != ErrAttachmentNotFound {
		t.Errorf("Expected ErrAttachmentNotFound, got %v", err)
	}
}

func TestDocument_Clear(t *testing.T) {
	doc := NewDocument("test-doc")
	
	// Add content
	content, _ := valueobjects.NewContent("some content")
	doc.UpdateContent(content)
	
	originalVersion := doc.Version()
	
	// Clear the document
	doc.Clear()
	
	if !doc.Content().IsEmpty() {
		t.Error("Expected content to be empty after clear")
	}
	
	if doc.Version() != originalVersion+1 {
		t.Error("Expected version to increment after clear")
	}
}

func TestDocument_GetAttachments(t *testing.T) {
	doc := NewDocument("test-doc")
	
	fileName1, _ := valueobjects.NewFileName("test1.txt")
	fileName2, _ := valueobjects.NewFileName("test2.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)

	attachment1 := NewAttachment("attachment-1", doc.ID(), fileName1, mimeType, fileSize)
	attachment2 := NewAttachment("attachment-2", doc.ID(), fileName2, mimeType, fileSize)
	
	doc.AddAttachment(attachment1)
	doc.AddAttachment(attachment2)
	
	attachments := doc.GetAttachments()
	
	if len(attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %v", len(attachments))
	}
}