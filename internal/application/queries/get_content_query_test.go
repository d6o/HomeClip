package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestGetContentQueryHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	
	// Create test document with content
	content, _ := valueobjects.NewContent("test content")
	testDoc := entities.RestoreDocument(
		documentID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	// Execute query
	query := GetContentQuery{
		DocumentID: string(documentID),
	}
	
	response, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if response == nil {
		t.Fatal("Expected response, got nil")
	}
	
	if response.Content != "test content" {
		t.Errorf("Expected content 'test content', got %v", response.Content)
	}
	
	if response.Version != 1 {
		t.Errorf("Expected version 1, got %v", response.Version)
	}
	
	if len(response.Attachments) != 0 {
		t.Errorf("Expected no attachments, got %v", len(response.Attachments))
	}
}

func TestGetContentQueryHandler_Handle_WithAttachments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	
	// Create test document with attachments
	content, _ := valueobjects.NewContent("test content")
	testDoc := entities.RestoreDocument(
		documentID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	// Add attachment
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)
	attachment := entities.NewAttachment(
		"att-1",
		documentID,
		fileName,
		mimeType,
		fileSize,
	)
	testDoc.AddAttachment(attachment)

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	// Execute query
	query := GetContentQuery{
		DocumentID: string(documentID),
	}
	
	response, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(response.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %v", len(response.Attachments))
	}
	
	att := response.Attachments[0]
	if att.ID != "att-1" {
		t.Errorf("Expected attachment ID 'att-1', got %v", att.ID)
	}
	
	if att.FileName != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got %v", att.FileName)
	}
	
	if att.MimeType != "text/plain" {
		t.Errorf("Expected mime type 'text/plain', got %v", att.MimeType)
	}
	
	if att.Size != 100 {
		t.Errorf("Expected size 100, got %v", att.Size)
	}
}

func TestGetContentQueryHandler_Handle_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := context.Background()
	
	// Create test document
	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	// Setup expectations - should use default document ID
	mockRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	// Execute query with empty document ID
	query := GetContentQuery{
		DocumentID: "",
	}
	
	response, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if response == nil {
		t.Fatal("Expected response, got nil")
	}
}

func TestGetContentQueryHandler_Handle_NewDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := context.Background()
	documentID := entities.DocumentID("new-doc")
	
	// First call returns not found, second call saves new document
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, entities.ErrDocumentNotFound)
	
	mockRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			if doc.ID() != documentID {
				t.Errorf("Expected document ID %v, got %v", documentID, doc.ID())
			}
			return nil
		})

	// Execute query
	query := GetContentQuery{
		DocumentID: string(documentID),
	}
	
	response, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if response == nil {
		t.Fatal("Expected response, got nil")
	}
	
	if response.Content != "" {
		t.Errorf("Expected empty content for new document, got %v", response.Content)
	}
	
	if response.Version != 1 {
		t.Errorf("Expected version 1 for new document, got %v", response.Version)
	}
}

func TestGetContentQueryHandler_Handle_ExpiredAttachments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	
	// Create test document
	content, _ := valueobjects.NewContent("test content")
	testDoc := entities.RestoreDocument(
		documentID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	// Add expired attachment
	fileName, _ := valueobjects.NewFileName("expired.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)
	
	expiredAttachment := entities.RestoreAttachment(
		"att-expired",
		documentID,
		fileName,
		mimeType,
		fileSize,
		valueobjects.TimestampFrom(time.Now().Add(-25*time.Hour)),
		valueobjects.ExpirationTimeFrom(time.Now().Add(-1*time.Hour)),
	)
	
	// Add active attachment
	activeAttachment := entities.NewAttachment(
		"att-active",
		documentID,
		fileName,
		mimeType,
		fileSize,
	)
	
	testDoc.AddAttachment(expiredAttachment)
	testDoc.AddAttachment(activeAttachment)

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	// Execute query
	query := GetContentQuery{
		DocumentID: string(documentID),
	}
	
	response, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Should only return non-expired attachment
	if len(response.Attachments) != 1 {
		t.Fatalf("Expected 1 non-expired attachment, got %v", len(response.Attachments))
	}
	
	if response.Attachments[0].ID != "att-active" {
		t.Errorf("Expected active attachment, got %v", response.Attachments[0].ID)
	}
}

func TestGetContentQueryHandler_Handle_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	expectedErr := errors.New("repository error")
	
	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, expectedErr)

	// Execute query
	query := GetContentQuery{
		DocumentID: string(documentID),
	}
	
	response, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	
	if response != nil {
		t.Error("Expected nil response on error")
	}
}