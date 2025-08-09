package queries

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestListFilesQueryHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	
	// Create test document with attachments
	testDoc := entities.NewDocument(documentID)
	
	// Add attachments
	fileName1, _ := valueobjects.NewFileName("file1.txt")
	fileName2, _ := valueobjects.NewFileName("file2.pdf")
	mimeType1, _ := valueobjects.NewMimeType("text/plain")
	mimeType2, _ := valueobjects.NewMimeType("application/pdf")
	fileSize, _ := valueobjects.NewFileSize(100)
	
	attachment1 := entities.NewAttachment("att-1", documentID, fileName1, mimeType1, fileSize)
	attachment2 := entities.NewAttachment("att-2", documentID, fileName2, mimeType2, fileSize)
	
	testDoc.AddAttachment(attachment1)
	testDoc.AddAttachment(attachment2)

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	// Execute query
	query := ListFilesQuery{
		DocumentID: string(documentID),
	}
	
	attachments, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(attachments) != 2 {
		t.Fatalf("Expected 2 attachments, got %v", len(attachments))
	}
	
	// Check attachment IDs
	foundAtt1 := false
	foundAtt2 := false
	for _, att := range attachments {
		if att.ID() == "att-1" {
			foundAtt1 = true
		}
		if att.ID() == "att-2" {
			foundAtt2 = true
		}
	}
	
	if !foundAtt1 {
		t.Error("Expected to find attachment att-1")
	}
	
	if !foundAtt2 {
		t.Error("Expected to find attachment att-2")
	}
}

func TestListFilesQueryHandler_Handle_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := context.Background()
	
	// Create test document with default ID
	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	// Setup expectations - should use default document ID
	mockRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	// Execute query with empty document ID
	query := ListFilesQuery{
		DocumentID: "",
	}
	
	attachments, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if attachments == nil {
		t.Fatal("Expected attachments array, got nil")
	}
	
	if len(attachments) != 0 {
		t.Errorf("Expected empty attachments for new document, got %v", len(attachments))
	}
}

func TestListFilesQueryHandler_Handle_DocumentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := context.Background()
	documentID := entities.DocumentID("non-existent")

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, entities.ErrDocumentNotFound)

	// Execute query
	query := ListFilesQuery{
		DocumentID: string(documentID),
	}
	
	attachments, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error for non-existent document, got %v", err)
	}
	
	if attachments == nil {
		t.Fatal("Expected empty attachments array, got nil")
	}
	
	if len(attachments) != 0 {
		t.Errorf("Expected empty attachments for non-existent document, got %v", len(attachments))
	}
}

func TestListFilesQueryHandler_Handle_NoAttachments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	
	// Create test document without attachments
	testDoc := entities.NewDocument(documentID)

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	// Execute query
	query := ListFilesQuery{
		DocumentID: string(documentID),
	}
	
	attachments, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if attachments == nil {
		t.Fatal("Expected empty attachments array, got nil")
	}
	
	if len(attachments) != 0 {
		t.Errorf("Expected 0 attachments, got %v", len(attachments))
	}
}

func TestListFilesQueryHandler_Handle_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := context.Background()
	documentID := entities.DocumentID("test-doc")
	expectedErr := errors.New("repository error")

	// Setup expectations
	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, expectedErr)

	// Execute query
	query := ListFilesQuery{
		DocumentID: string(documentID),
	}
	
	attachments, err := handler.Handle(ctx, query)
	
	// Assertions
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	
	if attachments != nil {
		t.Error("Expected nil attachments on error")
	}
}