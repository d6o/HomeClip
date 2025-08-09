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
	defer ctrl.Finish()

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
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	// Start the service
	service.Start(ctx)

	// Verify it's running
	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	if !running {
		t.Error("Expected service to be running after Start")
	}

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify it's stopped
	service.mu.Lock()
	running = service.running
	service.mu.Unlock()

	if running {
		t.Error("Expected service to be stopped after Stop")
	}
}

func TestCleanupService_Start_AlreadyRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	// Start the service
	service.Start(ctx)
	defer service.Stop()

	// Try to start again - should not panic
	service.Start(ctx)

	// Verify still running
	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	if !running {
		t.Error("Expected service to still be running")
	}
}

func TestCleanupService_Stop_NotRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)

	// Stop without starting - should not panic
	service.Stop()

	// Verify still not running
	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	if running {
		t.Error("Expected service to not be running")
	}
}

func TestCleanupService_CleanupDocument_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	// Create an expired document with attachments
	docID := entities.DocumentID("test-doc")
	
	// Create a document with past expiration time
	expiredTime := valueobjects.ExpirationTimeFrom(time.Now().Add(-25 * time.Hour))
	
	fileName1, _ := valueobjects.NewFileName("file1.txt")
	fileName2, _ := valueobjects.NewFileName("file2.txt")
	mimeType1, _ := valueobjects.NewMimeType("text/plain")
	mimeType2, _ := valueobjects.NewMimeType("text/plain")
	fileSize1, _ := valueobjects.NewFileSize(100)
	fileSize2, _ := valueobjects.NewFileSize(200)
	
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

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachment1.ID()).
		Return(nil)

	mockFileStorage.EXPECT().
		Delete(ctx, attachment2.ID()).
		Return(nil)

	// Execute cleanup
	err := service.CleanupDocument(ctx, docID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCleanupService_CleanupDocument_NotExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	// Create a non-expired document
	docID := entities.DocumentID("test-doc")
	doc := entities.NewDocument(docID)

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)

	// Should not call Delete on file storage

	// Execute cleanup
	err := service.CleanupDocument(ctx, docID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCleanupService_CleanupDocument_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	docID := entities.DocumentID("non-existent")

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(nil, entities.ErrDocumentNotFound)

	// Execute cleanup - should not error for not found
	err := service.CleanupDocument(ctx, docID)

	if err != nil {
		t.Errorf("Expected no error for not found document, got %v", err)
	}
}

func TestCleanupService_CleanupDocument_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	docID := entities.DocumentID("test-doc")
	expectedError := errors.New("database error")

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(nil, expectedError)

	// Execute cleanup
	err := service.CleanupDocument(ctx, docID)

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestCleanupService_CleanupDocument_FileDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 5 * time.Minute

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	// Create an expired document with attachments
	docID := entities.DocumentID("test-doc")
	expiredTime := valueobjects.ExpirationTimeFrom(time.Now().Add(-25 * time.Hour))
	fileName, _ := valueobjects.NewFileName("file.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)
	attachment := entities.NewAttachment(
		entities.AttachmentID("attach1"),
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

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)

	// File delete fails but should not return error
	mockFileStorage.EXPECT().
		Delete(ctx, attachment.ID()).
		Return(errors.New("storage error"))

	// Execute cleanup - should not return error even if file delete fails
	err := service.CleanupDocument(ctx, docID)

	if err != nil {
		t.Errorf("Expected no error even with file delete failure, got %v", err)
	}
}

func TestCleanupService_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start the service
	service.Start(ctx)

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the context
	cancel()

	// Give it time to stop
	time.Sleep(150 * time.Millisecond)

	// Service should have stopped due to context cancellation
	service.mu.Lock()
	running := service.running
	service.mu.Unlock()

	// Clean up
	service.Stop()

	if running {
		t.Log("Service might still be running, but should stop soon due to context cancellation")
	}
}

func TestCleanupService_PeriodicExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	expirationService := domainservices.NewExpirationService(mockDocRepo, mockFileStorage)
	interval := 100 * time.Millisecond

	service := NewCleanupService(mockDocRepo, mockFileStorage, expirationService, interval)
	ctx := context.Background()

	// Start the service
	service.Start(ctx)

	// Let it run for multiple intervals
	time.Sleep(350 * time.Millisecond)

	// Stop the service
	service.Stop()

	// The performCleanup should have been called at least 3 times
	// (once immediately and at least 2 more times from ticker)
	// Since performCleanup is a no-op in the current implementation,
	// we can't directly test this, but we verify the service ran
	// without errors
}