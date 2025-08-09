package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
	"go.uber.org/mock/gomock"
)

func TestDocumentService_GetOrCreateDocument_ExistingDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	service := NewDocumentService(mockRepo)

	ctx := context.Background()
	id := entities.DocumentID("test-doc")
	existingDoc := entities.NewDocument(id)

	mockRepo.EXPECT().
		FindByID(ctx, id).
		Return(existingDoc, nil)

	doc, err := service.GetOrCreateDocument(ctx, id)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if doc.ID() != id {
		t.Errorf("Expected document ID %v, got %v", id, doc.ID())
	}
}

func TestDocumentService_GetOrCreateDocument_NewDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	service := NewDocumentService(mockRepo)

	ctx := context.Background()
	id := entities.DocumentID("test-doc")

	mockRepo.EXPECT().
		FindByID(ctx, id).
		Return(nil, entities.ErrDocumentNotFound)

	mockRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			if doc.ID() != id {
				t.Errorf("Expected document ID %v, got %v", id, doc.ID())
			}
			return nil
		})

	doc, err := service.GetOrCreateDocument(ctx, id)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if doc.ID() != id {
		t.Errorf("Expected document ID %v, got %v", id, doc.ID())
	}
}

func TestDocumentService_GetOrCreateDocument_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	service := NewDocumentService(mockRepo)

	ctx := context.Background()
	id := entities.DocumentID("test-doc")
	expectedErr := errors.New("repository error")

	mockRepo.EXPECT().
		FindByID(ctx, id).
		Return(nil, expectedErr)

	_, err := service.GetOrCreateDocument(ctx, id)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestDocumentService_UpdateDocumentContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	service := NewDocumentService(mockRepo)

	ctx := context.Background()
	id := entities.DocumentID("test-doc")
	existingDoc := entities.NewDocument(id)
	contentValue := "updated content"

	mockRepo.EXPECT().
		FindByID(ctx, id).
		Return(existingDoc, nil)

	mockRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			if doc.Content().Value() != contentValue {
				t.Errorf("Expected content %v, got %v", contentValue, doc.Content().Value())
			}
			return nil
		})

	doc, err := service.UpdateDocumentContent(ctx, id, contentValue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if doc.Content().Value() != contentValue {
		t.Errorf("Expected content %v, got %v", contentValue, doc.Content().Value())
	}
}

func TestDocumentService_UpdateDocumentContent_InvalidContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	service := NewDocumentService(mockRepo)

	ctx := context.Background()
	id := entities.DocumentID("test-doc")
	// Create content that exceeds max size
	contentValue := strings.Repeat("a", valueobjects.MaxContentLength+1)

	_, err := service.UpdateDocumentContent(ctx, id, contentValue)
	if err == nil {
		t.Fatal("Expected error for invalid content, got nil")
	}

	if err != valueobjects.ErrContentTooLarge {
		t.Errorf("Expected ErrContentTooLarge, got %v", err)
	}
}

func TestDocumentService_UpdateDocumentContent_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositories.NewMockDocumentRepository(ctrl)
	service := NewDocumentService(mockRepo)

	ctx := context.Background()
	id := entities.DocumentID("test-doc")
	existingDoc := entities.NewDocument(id)
	contentValue := "updated content"
	saveErr := errors.New("save failed")

	mockRepo.EXPECT().
		FindByID(ctx, id).
		Return(existingDoc, nil)

	mockRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(saveErr)

	_, err := service.UpdateDocumentContent(ctx, id, contentValue)
	if err != saveErr {
		t.Errorf("Expected error %v, got %v", saveErr, err)
	}
}
