package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/d6o/homeclip/internal/application/commands"
	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/application/queries"
	"github.com/d6o/homeclip/internal/application/services"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	domainservices "github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestDocumentHandler_GetContent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	handler := NewDocumentHandler(appService)

	testDoc := entities.RestoreDocument(
		entities.DefaultDocumentID,
		valueobjects.EmptyContent(),
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)

	mockDocRepo.EXPECT().
		FindByID(gomock.Any(), entities.DefaultDocumentID).
		Return(testDoc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	rec := httptest.NewRecorder()

	handler.GetContent(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response dtos.GetContentResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Content != "" {
		t.Errorf("Expected empty content, got %v", response.Content)
	}

	if response.Version != 1 {
		t.Errorf("Expected version 1, got %v", response.Version)
	}
}

func TestDocumentHandler_GetContent_MethodNotAllowed(t *testing.T) {
	handler := &DocumentHandler{}

	req := httptest.NewRequest(http.MethodPost, "/api/content", nil)
	rec := httptest.NewRecorder()

	handler.GetContent(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestDocumentHandler_SaveContent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	handler := NewDocumentHandler(appService)

	testDoc := entities.NewDocument(entities.DefaultDocumentID)

	mockDocRepo.EXPECT().
		FindByID(gomock.Any(), entities.DefaultDocumentID).
		Return(testDoc, nil)

	mockDocRepo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil)

	reqBody := dtos.UpdateContentRequest{
		Content: "test content",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/content", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SaveContent(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response dtos.UpdateContentResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestDocumentHandler_SaveContent_InvalidJSON(t *testing.T) {
	handler := &DocumentHandler{}

	req := httptest.NewRequest(http.MethodPost, "/api/content", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SaveContent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDocumentHandler_SaveContent_MethodNotAllowed(t *testing.T) {
	handler := &DocumentHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	rec := httptest.NewRecorder()

	handler.SaveContent(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestDocumentHandler_SaveContent_ContentTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	handler := NewDocumentHandler(appService)

	largeContent := make([]byte, valueobjects.MaxContentLength+1)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	reqBody := dtos.UpdateContentRequest{
		Content: string(largeContent),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/content", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SaveContent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDocumentHandler_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	handler := NewDocumentHandler(appService)

	mockDocRepo.EXPECT().
		FindByID(gomock.Any(), entities.DefaultDocumentID).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	rec := httptest.NewRecorder()

	handler.GetContent(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// Helper function tests
func TestDocumentHandler_WriteJSON(t *testing.T) {
	handler := &DocumentHandler{}

	testData := map[string]string{
		"key": "value",
	}

	rec := httptest.NewRecorder()
	handler.writeJSON(rec, http.StatusOK, testData)

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type header to be application/json")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check body
	var result map[string]string
	err := json.NewDecoder(rec.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected key=value, got %v", result)
	}
}

func TestDocumentHandler_HandleError(t *testing.T) {
	handler := &DocumentHandler{}

	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "bad request error",
			err:        errors.New("bad request"),
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "internal server error",
			err:        errors.New("internal error"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.handleError(rec, tt.err, tt.statusCode)

			if rec.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, rec.Code)
			}

			var response map[string]string
			err := json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response["error"] != tt.err.Error() {
				t.Errorf("Expected error message %v, got %v", tt.err.Error(), response["error"])
			}
		})
	}
}

func TestDocumentHandler_MethodNotAllowed(t *testing.T) {
	handler := &DocumentHandler{}

	rec := httptest.NewRecorder()
	handler.methodNotAllowed(rec)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
