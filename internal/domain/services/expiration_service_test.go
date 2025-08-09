package services

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestNewExpirationService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	
	service := NewExpirationService(mockDocRepo, mockFileStorage)
	
	if service == nil {
		t.Error("Expected service to be created")
	}
	
	if service.documentRepo == nil {
		t.Error("Expected document repository to be set")
	}
	
	if service.fileStorage == nil {
		t.Error("Expected file storage to be set")
	}
}

func TestExpirationService_IsDocumentExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	service := NewExpirationService(mockDocRepo, mockFileStorage)

	// Create an expired document
	expiredDoc := entities.RestoreDocument(
		"doc-1",
		valueobjects.EmptyContent(),
		nil,
		valueobjects.TimestampFrom(time.Now().Add(-25*time.Hour)),
		valueobjects.ExpirationTimeFrom(time.Now().Add(-1*time.Hour)),
		1,
	)
	
	if !service.IsDocumentExpired(expiredDoc) {
		t.Error("Expected document to be expired")
	}
	
	// Create a non-expired document
	activeDoc := entities.RestoreDocument(
		"doc-2",
		valueobjects.EmptyContent(),
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	if service.IsDocumentExpired(activeDoc) {
		t.Error("Expected document to not be expired")
	}
}

func TestExpirationService_ShouldCleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	service := NewExpirationService(mockDocRepo, mockFileStorage)

	tests := []struct {
		name           string
		expirationTime time.Time
		shouldCleanup  bool
	}{
		{
			name:           "expired with grace period passed",
			expirationTime: time.Now().Add(-2 * time.Hour),
			shouldCleanup:  true,
		},
		{
			name:           "expired but within grace period",
			expirationTime: time.Now().Add(-30 * time.Minute),
			shouldCleanup:  false,
		},
		{
			name:           "not expired",
			expirationTime: time.Now().Add(1 * time.Hour),
			shouldCleanup:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := entities.RestoreDocument(
				"doc-test",
				valueobjects.EmptyContent(),
				nil,
				valueobjects.NewTimestamp(),
				valueobjects.ExpirationTimeFrom(tt.expirationTime),
				1,
			)
			
			if service.ShouldCleanup(doc) != tt.shouldCleanup {
				t.Errorf("Expected ShouldCleanup to return %v", tt.shouldCleanup)
			}
		})
	}
}

func TestExpirationService_ValidateAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	service := NewExpirationService(mockDocRepo, mockFileStorage)

	// Test with expired document
	expiredDoc := entities.RestoreDocument(
		"doc-1",
		valueobjects.EmptyContent(),
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.ExpirationTimeFrom(time.Now().Add(-1*time.Hour)),
		1,
	)
	
	err := service.ValidateAccess(expiredDoc)
	if err != valueobjects.ErrExpired {
		t.Errorf("Expected ErrExpired, got %v", err)
	}
	
	// Test with valid document
	validDoc := entities.RestoreDocument(
		"doc-2",
		valueobjects.EmptyContent(),
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	err = service.ValidateAccess(validDoc)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestExpirationService_ExtendExpiration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	service := NewExpirationService(mockDocRepo, mockFileStorage)

	ctx := context.Background()
	docID := entities.DocumentID("doc-1")
	
	// Create a document that expires in 1 hour
	doc := entities.RestoreDocument(
		docID,
		valueobjects.EmptyContent(),
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.ExpirationTimeFrom(time.Now().Add(1*time.Hour)),
		1,
	)
	
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(doc, nil)
	
	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, updatedDoc *entities.Document) error {
			// Verify the document version was incremented
			if updatedDoc.Version() != 2 {
				t.Errorf("Expected version 2, got %v", updatedDoc.Version())
			}
			
			// Verify the expiration was extended
			remaining := updatedDoc.ExpiresAt().TimeRemaining()
			if remaining < 1*time.Hour+50*time.Minute {
				t.Errorf("Expected expiration to be extended, got %v", remaining)
			}
			
			return nil
		})
	
	err := service.ExtendExpiration(ctx, docID, 1*time.Hour)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestExpirationService_ExtendExpiration_ExpiredDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	service := NewExpirationService(mockDocRepo, mockFileStorage)

	ctx := context.Background()
	docID := entities.DocumentID("doc-1")
	
	// Create an expired document
	expiredDoc := entities.RestoreDocument(
		docID,
		valueobjects.EmptyContent(),
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.ExpirationTimeFrom(time.Now().Add(-1*time.Hour)),
		1,
	)
	
	mockDocRepo.EXPECT().
		FindByID(ctx, docID).
		Return(expiredDoc, nil)
	
	err := service.ExtendExpiration(ctx, docID, 1*time.Hour)
	if err != valueobjects.ErrExpired {
		t.Errorf("Expected ErrExpired, got %v", err)
	}
}