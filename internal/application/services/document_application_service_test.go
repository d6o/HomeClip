package services

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/d6o/homeclip/internal/application"
	"github.com/d6o/homeclip/internal/application/commands"
	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/application/queries"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestDocumentApplicationService_GetContent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()
	documentID := "test-doc"

	content, err := valueobjects.NewContent("test content")
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}
	testDoc := entities.RestoreDocument(
		entities.DocumentID(documentID),
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(testDoc, nil)

	response, err := appService.GetContent(ctx, documentID)
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
}

func TestDocumentApplicationService_UpdateContent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()
	documentID := "test-doc"
	newContent := "updated content"

	testDoc := entities.NewDocument(entities.DocumentID(documentID))

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(testDoc, nil)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			if doc.Content().Value() != newContent {
				t.Errorf("Expected content %v, got %v", newContent, doc.Content().Value())
			}
			return nil
		})

	response, err := appService.UpdateContent(ctx, documentID, newContent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Version != 2 {
		t.Errorf("Expected version 2 after update, got %v", response.Version)
	}
}

func TestDocumentApplicationService_GetContent_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()

	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	response, err := appService.GetContent(ctx, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}
}

func TestDocumentApplicationService_UpdateContent_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()
	newContent := "updated content"

	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	response, err := appService.UpdateContent(ctx, "", newContent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestDocumentApplicationService_GetContent_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()
	documentID := "test-doc"
	expectedErr := errors.New("repository error")

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(nil, expectedErr)

	response, err := appService.GetContent(ctx, documentID)

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if response != nil {
		t.Error("Expected nil response on error")
	}
}

func TestDocumentApplicationService_UpdateContent_InvalidContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()
	documentID := "test-doc"

	invalidContent := string(make([]byte, valueobjects.MaxContentLength+1))

	response, err := appService.UpdateContent(ctx, documentID, invalidContent)

	if err == nil {
		t.Fatal("Expected error for invalid content, got nil")
	}

	if err != valueobjects.ErrContentTooLarge {
		t.Errorf("Expected ErrContentTooLarge, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response even on error")
	}

	if response.Success {
		t.Error("Expected success to be false on error")
	}
}

func TestDocumentApplicationService_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)

	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)

	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := t.Context()
	documentID := "test-doc"

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(nil, entities.ErrDocumentNotFound)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			if doc.ID() != entities.DocumentID(documentID) {
				t.Errorf("Expected document ID %v, got %v", documentID, doc.ID())
			}
			return nil
		})

	response1, err := appService.GetContent(ctx, documentID)
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	if response1.Content != "" {
		t.Error("Expected empty content for new document")
	}

	newContent := "Hello, World!"
	testDoc := entities.NewDocument(entities.DocumentID(documentID))

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(testDoc, nil)

	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, doc *entities.Document) error {
			if doc.Content().Value() != newContent {
				t.Errorf("Expected content %v, got %v", newContent, doc.Content().Value())
			}
			return nil
		})

	response2, err := appService.UpdateContent(ctx, documentID, newContent)
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	if !response2.Success {
		t.Error("Expected successful update")
	}

	updatedContent, err := valueobjects.NewContent(newContent)
	if err != nil {
		t.Fatalf("Failed to create updated content: %v", err)
	}
	updatedDoc := entities.RestoreDocument(
		entities.DocumentID(documentID),
		updatedContent,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		2,
	)

	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(updatedDoc, nil)

	response3, err := appService.GetContent(ctx, documentID)
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	if response3.Content != newContent {
		t.Errorf("Expected content %v, got %v", newContent, response3.Content)
	}
	if response3.Version != 2 {
		t.Errorf("Expected version 2, got %v", response3.Version)
	}
}

// TestDocumentApplicationService_WithMockedHandlers demonstrates that the service
// now accepts interfaces and can be fully mocked for testing
func TestDocumentApplicationService_WithMockedHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockGetHandler := application.NewMockGetContentQueryHandler(ctrl)
	mockUpdateHandler := application.NewMockUpdateContentCommandHandler(ctrl)

	service := NewDocumentApplicationService(mockUpdateHandler, mockGetHandler)

	t.Run("GetContent with mocked handler", func(t *testing.T) {
		ctx := t.Context()
		documentID := "mock-doc"

		expectedResponse := &dtos.GetContentResponse{
			Content:     "mocked content",
			Version:     42,
			Attachments: []dtos.AttachmentDTO{},
		}

		mockGetHandler.EXPECT().
			Handle(ctx, queries.GetContentQuery{DocumentID: documentID}).
			Return(expectedResponse, nil)

		response, err := service.GetContent(ctx, documentID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.Content != expectedResponse.Content {
			t.Errorf("Expected content %s, got %s", expectedResponse.Content, response.Content)
		}

		if response.Version != expectedResponse.Version {
			t.Errorf("Expected version %d, got %d", expectedResponse.Version, response.Version)
		}
	})

	t.Run("UpdateContent with mocked handler", func(t *testing.T) {
		ctx := t.Context()
		documentID := "mock-doc"
		newContent := "updated mock content"

		expectedResponse := &dtos.UpdateContentResponse{
			Success: true,
			Version: 43,
		}

		mockUpdateHandler.EXPECT().
			Handle(ctx, commands.UpdateContentCommand{
				DocumentID: documentID,
				Content:    newContent,
			}).
			Return(expectedResponse, nil)

		response, err := service.UpdateContent(ctx, documentID, newContent)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}

		if response.Version != expectedResponse.Version {
			t.Errorf("Expected version %d, got %d", expectedResponse.Version, response.Version)
		}
	})

	t.Run("Error handling with mocked handler", func(t *testing.T) {
		ctx := t.Context()
		expectedErr := errors.New("mocked error")

		mockGetHandler.EXPECT().
			Handle(ctx, gomock.Any()).
			Return(nil, expectedErr)

		_, err := service.GetContent(ctx, "any-doc")

		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}
