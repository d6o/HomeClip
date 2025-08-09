package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestMemoryDocumentRepository_Save_And_FindByID(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	// Create a test document
	docID := entities.DocumentID("test-doc")
	content, _ := valueobjects.NewContent("test content")
	doc := entities.RestoreDocument(
		docID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	// Save the document
	err := repo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	
	// Retrieve the document
	retrieved, err := repo.FindByID(ctx, docID)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}
	
	if retrieved.ID() != docID {
		t.Errorf("Expected document ID %v, got %v", docID, retrieved.ID())
	}
	
	if retrieved.Content().Value() != "test content" {
		t.Errorf("Expected content 'test content', got %v", retrieved.Content().Value())
	}
	
	if retrieved.Version() != 1 {
		t.Errorf("Expected version 1, got %v", retrieved.Version())
	}
}

func TestMemoryDocumentRepository_FindByID_NotFound(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	// Try to find non-existent document
	_, err := repo.FindByID(ctx, entities.DocumentID("non-existent"))
	
	if err != entities.ErrDocumentNotFound {
		t.Errorf("Expected ErrDocumentNotFound, got %v", err)
	}
}

func TestMemoryDocumentRepository_Exists(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	docID := entities.DocumentID("test-doc")
	doc := entities.NewDocument(docID)
	
	// Check non-existent document
	exists, err := repo.Exists(ctx, docID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected document to not exist")
	}
	
	// Save document
	err = repo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	
	// Check existing document
	exists, err = repo.Exists(ctx, docID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected document to exist")
	}
}

func TestMemoryDocumentRepository_Update(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	docID := entities.DocumentID("test-doc")
	doc := entities.NewDocument(docID)
	
	// Save initial document
	err := repo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	
	// Update document
	updatedContent, _ := valueobjects.NewContent("updated content")
	doc.UpdateContent(updatedContent)
	
	err = repo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}
	
	// Retrieve and verify update
	retrieved, err := repo.FindByID(ctx, docID)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}
	
	if retrieved.Content().Value() != "updated content" {
		t.Errorf("Expected updated content, got %v", retrieved.Content().Value())
	}
	
	if retrieved.Version() != 2 {
		t.Errorf("Expected version 2 after update, got %v", retrieved.Version())
	}
}

func TestMemoryDocumentRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	// Test concurrent saves
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			docID := entities.DocumentID(string(rune('a' + id)))
			doc := entities.NewDocument(docID)
			
			err := repo.Save(ctx, doc)
			if err != nil {
				t.Errorf("Failed to save document %v: %v", docID, err)
			}
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all documents were saved
	for i := 0; i < 10; i++ {
		docID := entities.DocumentID(string(rune('a' + i)))
		exists, err := repo.Exists(ctx, docID)
		if err != nil {
			t.Errorf("Failed to check existence: %v", err)
		}
		if !exists {
			t.Errorf("Document %v should exist", docID)
		}
	}
}

func TestMemoryDocumentRepository_WithAttachments(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	docID := entities.DocumentID("test-doc")
	doc := entities.NewDocument(docID)
	
	// Add attachments
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)
	
	attachment1 := entities.NewAttachment("att-1", docID, fileName, mimeType, fileSize)
	attachment2 := entities.NewAttachment("att-2", docID, fileName, mimeType, fileSize)
	
	doc.AddAttachment(attachment1)
	doc.AddAttachment(attachment2)
	
	// Save document with attachments
	err := repo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	
	// Retrieve and verify attachments
	retrieved, err := repo.FindByID(ctx, docID)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}
	
	if retrieved.AttachmentCount() != 2 {
		t.Errorf("Expected 2 attachments, got %v", retrieved.AttachmentCount())
	}
	
	// Verify attachment IDs
	att1, err := retrieved.GetAttachment("att-1")
	if err != nil {
		t.Errorf("Failed to get attachment att-1: %v", err)
	}
	if att1.ID() != "att-1" {
		t.Errorf("Expected attachment ID att-1, got %v", att1.ID())
	}
	
	att2, err := retrieved.GetAttachment("att-2")
	if err != nil {
		t.Errorf("Failed to get attachment att-2: %v", err)
	}
	if att2.ID() != "att-2" {
		t.Errorf("Expected attachment ID att-2, got %v", att2.ID())
	}
}

func TestMemoryDocumentRepository_ExpiredDocument(t *testing.T) {
	repo := NewMemoryDocumentRepository()
	ctx := context.Background()
	
	docID := entities.DocumentID("expired-doc")
	
	// Create an expired document
	expiredTime := valueobjects.ExpirationTimeFrom(time.Now().Add(-1 * time.Hour))
	doc := entities.RestoreDocument(
		docID,
		valueobjects.EmptyContent(),
		nil,
		valueobjects.TimestampFrom(time.Now().Add(-25*time.Hour)),
		expiredTime,
		1,
	)
	
	// Save expired document
	err := repo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	
	// Should still be able to retrieve expired document
	// (cleanup is handled by a separate service)
	retrieved, err := repo.FindByID(ctx, docID)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}
	
	if !retrieved.IsExpired() {
		t.Error("Expected document to be expired")
	}
}