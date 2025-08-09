package queries

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestListFilesQueryHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")

	testDoc := entities.NewDocument(documentID)

	fileName1, err := valueobjects.NewFileName("file1.txt")
	if err != nil {
		t.Fatalf("Failed to create fileName1: %v", err)
	}
	fileName2, err := valueobjects.NewFileName("file2.pdf")
	if err != nil {
		t.Fatalf("Failed to create fileName2: %v", err)
	}
	mimeType1, err := valueobjects.NewMimeType("text/plain")
	if err != nil {
		t.Fatalf("Failed to create mimeType1: %v", err)
	}
	mimeType2, err := valueobjects.NewMimeType("application/pdf")
	if err != nil {
		t.Fatalf("Failed to create mimeType2: %v", err)
	}
	fileSize, err := valueobjects.NewFileSize(100)
	if err != nil {
		t.Fatalf("Failed to create fileSize: %v", err)
	}

	attachment1 := entities.NewAttachment("att-1", documentID, fileName1, mimeType1, fileSize)
	attachment2 := entities.NewAttachment("att-2", documentID, fileName2, mimeType2, fileSize)

	if err := testDoc.AddAttachment(attachment1); err != nil {
		t.Fatalf("Failed to add attachment1: %v", err)
	}
	if err := testDoc.AddAttachment(attachment2); err != nil {
		t.Fatalf("Failed to add attachment2: %v", err)
	}

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	query := ListFilesQuery{
		DocumentID: string(documentID),
	}

	attachments, err := handler.Handle(ctx, query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(attachments) != 2 {
		t.Fatalf("Expected 2 attachments, got %v", len(attachments))
	}

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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := t.Context()

	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	mockRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	query := ListFilesQuery{
		DocumentID: "",
	}

	attachments, err := handler.Handle(ctx, query)
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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := t.Context()
	documentID := entities.DocumentID("non-existent")

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, entities.ErrDocumentNotFound)

	query := ListFilesQuery{
		DocumentID: string(documentID),
	}

	attachments, err := handler.Handle(ctx, query)
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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")

	testDoc := entities.NewDocument(documentID)

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(testDoc, nil)

	query := ListFilesQuery{
		DocumentID: string(documentID),
	}

	attachments, err := handler.Handle(ctx, query)
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
	t.Cleanup(ctrl.Finish)

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	handler := NewListFilesQueryHandler(mockRepo)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	expectedErr := errors.New("repository error")

	mockRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, expectedErr)

	query := ListFilesQuery{
		DocumentID: string(documentID),
	}

	attachments, err := handler.Handle(ctx, query)

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if attachments != nil {
		t.Error("Expected nil attachments on error")
	}
}
