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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")

	content, err := valueobjects.NewContent("test content")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}
	testDoc := entities.RestoreDocument(
		documentID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	query := GetContentQuery{
		DocumentID: string(documentID),
	}

	response, err := handler.Handle(ctx, query)
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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")

	content, err := valueobjects.NewContent("test content")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}
	testDoc := entities.RestoreDocument(
		documentID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	fileName, err := valueobjects.NewFileName("test.txt")
	if err != nil {
		t.Fatalf("Failed to create fileName: %v", err)
	}
	mimeType, err := valueobjects.NewMimeType("text/plain")
	if err != nil {
		t.Fatalf("Failed to create mimeType: %v", err)
	}
	fileSize, err := valueobjects.NewFileSize(100)
	if err != nil {
		t.Fatalf("Failed to create fileSize: %v", err)
	}
	attachment := entities.NewAttachment(
		"att-1",
		documentID,
		fileName,
		mimeType,
		fileSize,
	)
	if err := testDoc.AddAttachment(attachment); err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	query := GetContentQuery{
		DocumentID: string(documentID),
	}

	response, err := handler.Handle(ctx, query)
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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := t.Context()

	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	mockRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	query := GetContentQuery{
		DocumentID: "",
	}

	response, err := handler.Handle(ctx, query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}
}

func TestGetContentQueryHandler_Handle_NewDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := t.Context()
	documentID := entities.DocumentID("new-doc")

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

	query := GetContentQuery{
		DocumentID: string(documentID),
	}

	response, err := handler.Handle(ctx, query)
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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")

	content, err := valueobjects.NewContent("test content")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}
	testDoc := entities.RestoreDocument(
		documentID,
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	fileName, err := valueobjects.NewFileName("expired.txt")
	if err != nil {
		t.Fatalf("Failed to create fileName: %v", err)
	}
	mimeType, err := valueobjects.NewMimeType("text/plain")
	if err != nil {
		t.Fatalf("Failed to create mimeType: %v", err)
	}
	fileSize, err := valueobjects.NewFileSize(100)
	if err != nil {
		t.Fatalf("Failed to create fileSize: %v", err)
	}

	expiredAttachment := entities.RestoreAttachment(
		"att-expired",
		documentID,
		fileName,
		mimeType,
		fileSize,
		valueobjects.TimestampFrom(time.Now().Add(-25*time.Hour)),
		valueobjects.ExpirationTimeFrom(time.Now().Add(-1*time.Hour)),
	)

	activeAttachment := entities.NewAttachment(
		"att-active",
		documentID,
		fileName,
		mimeType,
		fileSize,
	)

	if err := testDoc.AddAttachment(expiredAttachment); err != nil {
		t.Fatalf("Failed to add expired attachment: %v", err)
	}
	if err := testDoc.AddAttachment(activeAttachment); err != nil {
		t.Fatalf("Failed to add active attachment: %v", err)
	}

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	query := GetContentQuery{
		DocumentID: string(documentID),
	}

	response, err := handler.Handle(ctx, query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.Attachments) != 1 {
		t.Fatalf("Expected 1 non-expired attachment, got %v", len(response.Attachments))
	}

	if response.Attachments[0].ID != "att-active" {
		t.Errorf("Expected active attachment, got %v", response.Attachments[0].ID)
	}
}

func TestGetContentQueryHandler_Handle_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockRepo)
	handler := NewGetContentQueryHandler(documentService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	expectedErr := errors.New("repository error")

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, expectedErr)

	query := GetContentQuery{
		DocumentID: string(documentID),
	}

	response, err := handler.Handle(ctx, query)

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if response != nil {
		t.Error("Expected nil response on error")
	}
}
