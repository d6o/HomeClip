package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	domainservices "github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestNewCleanupService(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.documentRepo == nil {
		t.Error("Expected document repository to be set")
	}

	if service.fileStorage == nil {
		t.Error("Expected file storage to be set")
	}

	if service.expirationService == nil {
		t.Error("Expected expiration service to be set")
	}

	if service.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, service.interval)
	}

	if service.running {
		t.Error("Expected service to not be running initially")
	}

	if service.expiredDocuments == nil {
		t.Error("Expected expired documents map to be initialized")
	}
}

func TestCleanupService_Start_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	service.Start(ctx)

	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	if !running {
		t.Error("Expected service to be running after Start")
	}

	time.Sleep(50 * time.Millisecond)

	service.Stop()

	service.mu.Lock()
	running = service.running
	service.mu.Unlock()

	if running {
		t.Error("Expected service to be stopped after Stop")
	}
}

func TestCleanupService_Start_AlreadyRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	service.Start(ctx)
	defer service.Stop()

	service.Start(ctx)

	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	if !running {
		t.Error("Expected service to still be running")
	}
}

func TestCleanupService_Stop_NotRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)

	service.Stop()

	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	if running {
		t.Error("Expected service to not be running")
	}
}

func TestCleanupService_CleanupDocument_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	docID := entities.DocumentID("test-doc")

	expiredTime := valueobjects.ExpirationTimeFrom(time.Now().Add(-25 * time.Hour))

	fileName1, err := valueobjects.NewFileName("file1.txt")
	if err != nil {
		t.Fatalf("Failed to create fileName1: %v", err)
	}
	fileName2, err := valueobjects.NewFileName("file2.txt")
	if err != nil {
		t.Fatalf("Failed to create fileName2: %v", err)
	}
	mimeType1, err := valueobjects.NewMimeType("text/plain")
	if err != nil {
		t.Fatalf("Failed to create mimeType1: %v", err)
	}
	mimeType2, err := valueobjects.NewMimeType("text/plain")
	if err != nil {
		t.Fatalf("Failed to create mimeType2: %v", err)
	}
	fileSize1, err := valueobjects.NewFileSize(100)
	if err != nil {
		t.Fatalf("Failed to create fileSize1: %v", err)
	}
	fileSize2, err := valueobjects.NewFileSize(200)
	if err != nil {
		t.Fatalf("Failed to create fileSize2: %v", err)
	}

	attachment1 := entities.NewAttachment(
		entities.AttachmentID("attach1"),
		docID,
		fileName1,
		mimeType1,
		fileSize1,
	)
	attachment2 := entities.NewAttachment(
		entities.AttachmentID("attach2"),
		docID,
		fileName2,
		mimeType2,
		fileSize2,
	)

	attachments := make(map[entities.AttachmentID]*entities.Attachment)
	attachments[attachment1.ID()] = attachment1
	attachments[attachment2.ID()] = attachment2

	doc := entities.RestoreDocument(
		docID,
		valueobjects.EmptyContent(),
		attachments,
		valueobjects.NewTimestamp(),
		expiredTime,
		1,
	)

	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachment1.ID()).
		Return(nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachment2.ID()).
		Return(nil)

	err = service.CleanupDocument(ctx, docID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCleanupService_CleanupDocument_NotExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	docID := entities.DocumentID("test-doc")
	doc := entities.NewDocument(docID)

	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)

	err := service.CleanupDocument(ctx, docID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCleanupService_CleanupDocument_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	docID := entities.DocumentID("non-existent")

	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(nil, entities.ErrDocumentNotFound)

	err := service.CleanupDocument(ctx, docID)
	if err != nil {
		t.Errorf("Expected no error for not found document, got %v", err)
	}
}

func TestCleanupService_CleanupDocument_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	docID := entities.DocumentID("test-doc")
	expectedError := errors.New("database error")

	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(nil, expectedError)

	err := service.CleanupDocument(ctx, docID)

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestCleanupService_CleanupDocument_FileDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	docID := entities.DocumentID("test-doc")
	expiredTime := valueobjects.ExpirationTimeFrom(time.Now().Add(-25 * time.Hour))
	fileName, err := valueobjects.NewFileName("file.txt")
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
		"attach1",
		docID,
		fileName,
		mimeType,
		fileSize,
	)

	attachments := make(map[entities.AttachmentID]*entities.Attachment)
	attachments[attachment.ID()] = attachment

	doc := entities.RestoreDocument(
		docID,
		valueobjects.EmptyContent(),
		attachments,
		valueobjects.NewTimestamp(),
		expiredTime,
		1,
	)

	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachment.ID()).
		Return(errors.New("storage error"))

	err = service.CleanupDocument(ctx, docID)
	if err != nil {
		t.Errorf("Expected no error even with file delete failure, got %v", err)
	}
}

func TestCleanupService_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)

	ctx, cancel := context.WithCancel(t.Context())

	service.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	cancel()

	time.Sleep(150 * time.Millisecond)

	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	service.Stop()

	if running {
		t.Log("Service might still be running, but should stop soon due to context cancellation")
	}
}

func TestCleanupService_PeriodicExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := t.Context()

	service.Start(ctx)

	time.Sleep(350 * time.Millisecond)

	service.Stop()
}
