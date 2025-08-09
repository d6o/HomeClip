package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestUploadFileCommandHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	handler := NewUploadFileCommandHandler(
		documentService,
		mockDocRepo,
		mockFileStorage,
	)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	fileContent := []byte("test file content")
	
	// Create test document
	testDoc := entities.NewDocument(documentID)
	
	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)
	
	mockFileStorage.EXPECT().
		Store(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, attachmentID entities.AttachmentID, reader io.Reader) error {
			// Verify we can read the content
			buf := new(bytes.Buffer)
			buf.ReadFrom(reader)
			if buf.String() != string(fileContent) {
				t.Errorf("Expected file content %s, got %s", fileContent, buf.String())
			}
			return nil
		})
	
	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			// Verify attachment was added
			if doc.AttachmentCount() != 1 {
				t.Errorf("Expected 1 attachment, got %d", doc.AttachmentCount())
			}
			return nil
		})

	// Execute command
	cmd := UploadFileCommand{
		DocumentID: string(documentID),
		FileName:   "test.txt",
		MimeType:   "text/plain",
		Size:       int64(len(fileContent)),
		Reader:     bytes.NewReader(fileContent),
	}
	
	attachment, err := handler.Handle(ctx, cmd)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if attachment == nil {
		t.Fatal("Expected attachment, got nil")
	}
	
	if attachment.FileName().Value() != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got %v", attachment.FileName().Value())
	}
	
	if attachment.MimeType().Value() != "text/plain" {
		t.Errorf("Expected mime type 'text/plain', got %v", attachment.MimeType().Value())
	}
	
	if attachment.Size().Value() != int64(len(fileContent)) {
		t.Errorf("Expected size %d, got %v", len(fileContent), attachment.Size().Value())
	}
}

func TestUploadFileCommandHandler_Handle_InvalidFileName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	handler := NewUploadFileCommandHandler(
		documentService,
		mockDocRepo,
		mockFileStorage,
	)

	ctx := context.Background()
	
	// Execute command with invalid filename
	cmd := UploadFileCommand{
		DocumentID: "test-doc",
		FileName:   "../../../etc/passwd", // Path traversal attempt
		MimeType:   "text/plain",
		Size:       100,
		Reader:     bytes.NewReader([]byte("content")),
	}
	
	attachment, err := handler.Handle(ctx, cmd)
	
	// Assertions
	if err == nil {
		t.Fatal("Expected error for invalid filename, got nil")
	}
	
	if attachment != nil {
		t.Error("Expected nil attachment on error")
	}
}

// TestUploadFileCommandHandler_Handle_InvalidMimeType has been removed
// as the application now accepts all MIME types.
// Previously tested: application/x-executable
// func TestUploadFileCommandHandler_Handle_InvalidMimeType(t *testing.T) {
// 	... test removed ...
// }

func TestUploadFileCommandHandler_Handle_FileTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	handler := NewUploadFileCommandHandler(
		documentService,
		mockDocRepo,
		mockFileStorage,
	)

	ctx := context.Background()
	
	// Execute command with file too large
	cmd := UploadFileCommand{
		DocumentID: "test-doc",
		FileName:   "large.txt",
		MimeType:   "text/plain",
		Size:       valueobjects.MaxFileSize + 1, // Exceeds max size
		Reader:     bytes.NewReader([]byte("content")),
	}
	
	attachment, err := handler.Handle(ctx, cmd)
	
	// Assertions
	if err == nil {
		t.Fatal("Expected error for file too large, got nil")
	}
	
	if attachment != nil {
		t.Error("Expected nil attachment on error")
	}
}

func TestUploadFileCommandHandler_Handle_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	handler := NewUploadFileCommandHandler(
		documentService,
		mockDocRepo,
		mockFileStorage,
	)

	ctx := context.Background()
	fileContent := []byte("test content")
	
	// Create test document with default ID
	testDoc := entities.NewDocument(entities.DefaultDocumentID)
	
	// Setup expectations - should use default document ID
	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)
	
	mockFileStorage.EXPECT().
		Store(ctx, gomock.Any(), gomock.Any()).
		Return(nil)
	
	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	// Execute command with empty document ID
	cmd := UploadFileCommand{
		DocumentID: "",
		FileName:   "test.txt",
		MimeType:   "text/plain",
		Size:       int64(len(fileContent)),
		Reader:     bytes.NewReader(fileContent),
	}
	
	attachment, err := handler.Handle(ctx, cmd)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if attachment == nil {
		t.Fatal("Expected attachment, got nil")
	}
	
	if attachment.DocumentID() != entities.DefaultDocumentID {
		t.Errorf("Expected default document ID, got %v", attachment.DocumentID())
	}
}

func TestUploadFileCommandHandler_Handle_StorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	handler := NewUploadFileCommandHandler(
		documentService,
		mockDocRepo,
		mockFileStorage,
	)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	storageErr := errors.New("storage error")
	
	// Create test document
	testDoc := entities.NewDocument(documentID)
	
	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)
	
	mockFileStorage.EXPECT().
		Store(ctx, gomock.Any(), gomock.Any()).
		Return(storageErr)

	// Execute command
	cmd := UploadFileCommand{
		DocumentID: string(documentID),
		FileName:   "test.txt",
		MimeType:   "text/plain",
		Size:       100,
		Reader:     bytes.NewReader([]byte("content")),
	}
	
	attachment, err := handler.Handle(ctx, cmd)
	
	// Assertions
	if err != storageErr {
		t.Errorf("Expected storage error %v, got %v", storageErr, err)
	}
	
	if attachment != nil {
		t.Error("Expected nil attachment on error")
	}
}

func TestUploadFileCommandHandler_Handle_RollbackOnSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	handler := NewUploadFileCommandHandler(
		documentService,
		mockDocRepo,
		mockFileStorage,
	)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	saveErr := errors.New("save error")
	
	// Create test document
	testDoc := entities.NewDocument(documentID)
	
	var storedAttachmentID entities.AttachmentID
	
	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)
	
	mockFileStorage.EXPECT().
		Store(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, attachmentID entities.AttachmentID, reader io.Reader) error {
			storedAttachmentID = attachmentID
			return nil
		})
	
	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(saveErr)
	
	// Expect rollback: delete the stored file
	mockFileStorage.EXPECT().
		Delete(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, attachmentID entities.AttachmentID) error {
			if attachmentID != storedAttachmentID {
				t.Errorf("Expected to delete attachment %v, got %v", storedAttachmentID, attachmentID)
			}
			return nil
		})

	// Execute command
	cmd := UploadFileCommand{
		DocumentID: string(documentID),
		FileName:   "test.txt",
		MimeType:   "text/plain",
		Size:       100,
		Reader:     bytes.NewReader([]byte("content")),
	}
	
	attachment, err := handler.Handle(ctx, cmd)
	
	// Assertions
	if err != saveErr {
		t.Errorf("Expected save error %v, got %v", saveErr, err)
	}
	
	if attachment != nil {
		t.Error("Expected nil attachment on error")
	}
}