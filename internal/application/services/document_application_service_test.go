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
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	documentID := "test-doc"
	
	// Create test document
	content, _ := valueobjects.NewContent("test content")
	testDoc := entities.RestoreDocument(
		entities.DocumentID(documentID),
		content,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(testDoc, nil)

	// Execute
	response, err := appService.GetContent(ctx, documentID)
	
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
}

func TestDocumentApplicationService_UpdateContent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	documentID := "test-doc"
	newContent := "updated content"
	
	// Create test document
	testDoc := entities.NewDocument(entities.DocumentID(documentID))

	// Setup expectations
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

	// Execute
	response, err := appService.UpdateContent(ctx, documentID, newContent)
	
	// Assertions
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
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	
	// Create test document
	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	// Setup expectations - should use default document ID
	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)

	// Execute with empty document ID
	response, err := appService.GetContent(ctx, "")
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if response == nil {
		t.Fatal("Expected response, got nil")
	}
}

func TestDocumentApplicationService_UpdateContent_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	newContent := "updated content"
	
	// Create test document
	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	// Setup expectations - should use default document ID
	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DefaultDocumentID).
		Return(testDoc, nil)
	
	mockDocRepo.EXPECT().
		Save(ctx, gomock.Any()).
		Return(nil)

	// Execute with empty document ID
	response, err := appService.UpdateContent(ctx, "", newContent)
	
	// Assertions
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
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	documentID := "test-doc"
	expectedErr := errors.New("repository error")

	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(ctx, entities.DocumentID(documentID)).
		Return(nil, expectedErr)

	// Execute
	response, err := appService.GetContent(ctx, documentID)
	
	// Assertions
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	
	if response != nil {
		t.Error("Expected nil response on error")
	}
}

func TestDocumentApplicationService_UpdateContent_InvalidContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	documentID := "test-doc"
	// Create content that exceeds max size
	invalidContent := string(make([]byte, valueobjects.MaxContentLength+1))

	// Execute - should fail before reaching repository
	response, err := appService.UpdateContent(ctx, documentID, invalidContent)
	
	// Assertions
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
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := services.NewDocumentService(mockDocRepo)
	
	// Create handlers
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	// Create application service
	appService := NewDocumentApplicationService(
		updateContentHandler,
		getContentHandler,
	)

	ctx := context.Background()
	documentID := "test-doc"
	
	// Scenario: Create new document, update it, then get it
	
	// Step 1: Get non-existent document (creates new one)
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
	
	// Step 2: Update content
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
	
	// Step 3: Get updated content
	updatedContent, _ := valueobjects.NewContent(newContent)
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
	defer ctrl.Finish()

	// Create mock handlers using the generated mocks
	mockGetHandler := application.NewMockGetContentQueryHandler(ctrl)
	mockUpdateHandler := application.NewMockUpdateContentCommandHandler(ctrl)

	// Create service with mocked handler interfaces
	service := NewDocumentApplicationService(mockUpdateHandler, mockGetHandler)

	t.Run("GetContent with mocked handler", func(t *testing.T) {
		ctx := context.Background()
		documentID := "mock-doc"
		
		expectedResponse := &dtos.GetContentResponse{
			Content:     "mocked content",
			Version:     42,
			Attachments: []dtos.AttachmentDTO{},
		}
		
		// Setup mock expectation
		mockGetHandler.EXPECT().
			Handle(ctx, queries.GetContentQuery{DocumentID: documentID}).
			Return(expectedResponse, nil)
		
		// Execute
		response, err := service.GetContent(ctx, documentID)
		
		// Verify
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
		ctx := context.Background()
		documentID := "mock-doc"
		newContent := "updated mock content"
		
		expectedResponse := &dtos.UpdateContentResponse{
			Success: true,
			Version: 43,
		}
		
		// Setup mock expectation
		mockUpdateHandler.EXPECT().
			Handle(ctx, commands.UpdateContentCommand{
				DocumentID: documentID,
				Content:    newContent,
			}).
			Return(expectedResponse, nil)
		
		// Execute
		response, err := service.UpdateContent(ctx, documentID, newContent)
		
		// Verify
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
		ctx := context.Background()
		expectedErr := errors.New("mocked error")
		
		// Setup mock to return error
		mockGetHandler.EXPECT().
			Handle(ctx, gomock.Any()).
			Return(nil, expectedErr)
		
		// Execute
		_, err := service.GetContent(ctx, "any-doc")
		
		// Verify error is propagated
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}