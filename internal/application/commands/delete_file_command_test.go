package commands

import (
	"errors"
	"testing"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
	"go.uber.org/mock/gomock"
)

func TestDeleteFileCommandHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)

	handler := NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	attachmentID := entities.AttachmentID("test-attachment")

	document := entities.NewDocument(documentID)
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
		attachmentID,
		documentID,
		fileName,
		mimeType,
		fileSize,
	)
	if err := document.AddAttachment(attachment); err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(document, nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachmentID).
		Return(nil)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx interface{}, doc *entities.Document) error {
			if doc.AttachmentCount() != 0 {
				t.Errorf("Expected 0 attachments after deletion, got %d", doc.AttachmentCount())
			}
			return nil
		})

	cmd := DeleteFileCommand{
		DocumentID:   string(documentID),
		AttachmentID: string(attachmentID),
	}

	err = handler.Handle(ctx, cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteFileCommandHandler_Handle_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)

	handler := NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)

	ctx := t.Context()
	attachmentID := entities.AttachmentID("test-attachment")

	document := entities.NewDocument(entities.DefaultDocumentID)
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
		attachmentID,
		entities.DefaultDocumentID,
		fileName,
		mimeType,
		fileSize,
	)
	if err := document.AddAttachment(attachment); err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(document, nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachmentID).
		Return(nil)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	cmd := DeleteFileCommand{
		DocumentID:   "",
		AttachmentID: string(attachmentID),
	}

	err = handler.Handle(ctx, cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteFileCommandHandler_Handle_DocumentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)

	handler := NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	attachmentID := entities.AttachmentID("test-attachment")

	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(nil, entities.ErrDocumentNotFound)

	cmd := DeleteFileCommand{
		DocumentID:   string(documentID),
		AttachmentID: string(attachmentID),
	}

	err := handler.Handle(ctx, cmd)

	if !errors.Is(err, entities.ErrDocumentNotFound) {
		t.Errorf("Expected ErrDocumentNotFound, got %v", err)
	}
}

func TestDeleteFileCommandHandler_Handle_AttachmentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)

	handler := NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	attachmentID := entities.AttachmentID("non-existent")

	document := entities.NewDocument(documentID)

	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(document, nil)

	cmd := DeleteFileCommand{
		DocumentID:   string(documentID),
		AttachmentID: string(attachmentID),
	}

	err := handler.Handle(ctx, cmd)

	if !errors.Is(err, entities.ErrAttachmentNotFound) {
		t.Errorf("Expected ErrAttachmentNotFound, got %v", err)
	}
}

func TestDeleteFileCommandHandler_Handle_FileStorageDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)

	handler := NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	attachmentID := entities.AttachmentID("test-attachment")

	document := entities.NewDocument(documentID)
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
		attachmentID,
		documentID,
		fileName,
		mimeType,
		fileSize,
	)
	if err := document.AddAttachment(attachment); err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(document, nil)

	storageErr := errors.New("storage error")
	mockFileStorage.EXPECT().
		Delete(ctx, attachmentID).
		Return(storageErr)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	cmd := DeleteFileCommand{
		DocumentID:   string(documentID),
		AttachmentID: string(attachmentID),
	}

	err = handler.Handle(ctx, cmd)

	if err != nil {
		t.Fatalf("Expected no error even if file storage delete fails, got %v", err)
	}
}

func TestDeleteFileCommandHandler_Handle_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)

	handler := NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	attachmentID := entities.AttachmentID("test-attachment")

	document := entities.NewDocument(documentID)
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
		attachmentID,
		documentID,
		fileName,
		mimeType,
		fileSize,
	)
	if err := document.AddAttachment(attachment); err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	mockDocRepo.EXPECT().
		FindByID(ctx, documentID).
		Return(document, nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachmentID).
		Return(nil)

	saveErr := errors.New("save error")
	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(saveErr)

	cmd := DeleteFileCommand{
		DocumentID:   string(documentID),
		AttachmentID: string(attachmentID),
	}

	err = handler.Handle(ctx, cmd)

	if !errors.Is(err, saveErr) {
		t.Errorf("Expected save error %v, got %v", saveErr, err)
	}
}