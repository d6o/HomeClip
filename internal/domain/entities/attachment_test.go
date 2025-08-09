package entities

import (
	"testing"

	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestNewAttachment(t *testing.T) {
	fileName, err := valueobjects.NewFileName("test.txt")
	if err != nil {
		t.Fatalf("Failed to create file name: %v", err)
	}

	mimeType, err := valueobjects.NewMimeType("text/plain")
	if err != nil {
		t.Fatalf("Failed to create mime type: %v", err)
	}

	fileSize, err := valueobjects.NewFileSize(100)
	if err != nil {
		t.Fatalf("Failed to create file size: %v", err)
	}

	attachmentID := AttachmentID("attachment-1")
	documentID := DocumentID("doc-1")

	attachment := NewAttachment(attachmentID, documentID, fileName, mimeType, fileSize)

	if attachment.ID() != attachmentID {
		t.Errorf("Expected ID %v, got %v", attachmentID, attachment.ID())
	}

	if attachment.DocumentID() != documentID {
		t.Errorf("Expected DocumentID %v, got %v", documentID, attachment.DocumentID())
	}

	if attachment.FileName().Value() != "test.txt" {
		t.Errorf("Expected file name 'test.txt', got %v", attachment.FileName().Value())
	}

	if attachment.MimeType().Value() != "text/plain" {
		t.Errorf("Expected mime type 'text/plain', got %v", attachment.MimeType().Value())
	}

	if attachment.Size().Value() != 100 {
		t.Errorf("Expected size 100, got %v", attachment.Size().Value())
	}

	if attachment.UploadedAt().IsZero() {
		t.Error("Expected UploadedAt to be set")
	}

	if attachment.IsExpired() {
		t.Error("Expected new attachment to not be expired")
	}
}
