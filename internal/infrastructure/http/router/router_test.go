package router

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
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
	"github.com/d6o/homeclip/internal/infrastructure/http/handlers"
)

//go:embed static/*
var testStaticFiles embed.FS

func TestNewRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	documentHandler := handlers.NewDocumentHandler(appService)
	
	uploadHandler := commands.NewUploadFileCommandHandler(documentService, mockDocRepo, mockFileStorage)
	deleteHandler := commands.NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)
	getFileHandler := queries.NewGetFileQueryHandler(mockDocRepo, mockFileStorage)
	listFilesHandler := queries.NewListFilesQueryHandler(mockDocRepo)
	fileHandler := handlers.NewFileHandler(uploadHandler, deleteHandler, getFileHandler, listFilesHandler)
	
	router := NewRouter(documentHandler, fileHandler, testStaticFiles, true)
	
	if router == nil {
		t.Fatal("Expected router to be created")
	}
	
	if router.mux == nil {
		t.Error("Expected mux to be initialized")
	}
	
	if router.documentHandler == nil {
		t.Error("Expected document handler to be set")
	}
	
	if router.fileHandler == nil {
		t.Error("Expected file handler to be set")
	}
	
	if router.healthHandler == nil {
		t.Error("Expected health handler to be initialized")
	}
	
	if !router.enableFileUploads {
		t.Error("Expected file uploads to be enabled")
	}
}

func TestRouter_Setup_HealthEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup router
	router := setupTestRouter(ctrl, false)
	handler := router.Setup()
	
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "health endpoint",
			path:     "/api/health",
			expected: "healthy",
		},
		{
			name:     "ready endpoint",
			path:     "/api/ready",
			expected: "ready",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			
			handler.ServeHTTP(rec, req)
			
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
			}
			
			var response map[string]string
			err := json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			
			if response["status"] != tt.expected {
				t.Errorf("Expected status '%s', got %v", tt.expected, response["status"])
			}
		})
	}
}

func TestRouter_Setup_DocumentEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	
	// Create test document
	testContent, _ := valueobjects.NewContent("test content")
	testDoc := entities.RestoreDocument(
		entities.DefaultDocumentID,
		testContent,
		nil,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	// Setup expectations for GET
	mockDocRepo.EXPECT().
		FindByID(gomock.Any(), entities.DefaultDocumentID).
		Return(testDoc, nil).AnyTimes()
	
	// Setup expectations for POST
	mockDocRepo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	documentHandler := handlers.NewDocumentHandler(appService)
	
	uploadHandler := commands.NewUploadFileCommandHandler(documentService, mockDocRepo, mockFileStorage)
	deleteHandler := commands.NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)
	getFileHandler := queries.NewGetFileQueryHandler(mockDocRepo, mockFileStorage)
	listFilesHandler := queries.NewListFilesQueryHandler(mockDocRepo)
	fileHandler := handlers.NewFileHandler(uploadHandler, deleteHandler, getFileHandler, listFilesHandler)
	
	router := NewRouter(documentHandler, fileHandler, testStaticFiles, false)
	handler := router.Setup()
	
	// Test GET /api/content
	t.Run("GET content", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
		
		var response dtos.GetContentResponse
		err := json.NewDecoder(rec.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if response.Content != "test content" {
			t.Errorf("Expected content 'test content', got %v", response.Content)
		}
	})
	
	// Test POST /api/content
	t.Run("POST content", func(t *testing.T) {
		reqBody := dtos.UpdateContentRequest{
			Content: "new content",
		}
		body, _ := json.Marshal(reqBody)
		
		req := httptest.NewRequest(http.MethodPost, "/api/content", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
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
	})
}

func TestRouter_Setup_FileEndpoints_Enabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	
	// Create test document with attachments
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(100)
	attachment := entities.NewAttachment(
		entities.AttachmentID("file1"),
		entities.DefaultDocumentID,
		fileName,
		mimeType,
		fileSize,
	)
	
	attachments := make(map[entities.AttachmentID]*entities.Attachment)
	attachments[attachment.ID()] = attachment
	
	testDoc := entities.RestoreDocument(
		entities.DefaultDocumentID,
		valueobjects.EmptyContent(),
		attachments,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	// Setup expectations
	mockDocRepo.EXPECT().
		FindByID(gomock.Any(), entities.DefaultDocumentID).
		Return(testDoc, nil)
	
	router := setupTestRouterWithFileUploads(ctrl, mockDocRepo, mockFileStorage, true)
	handler := router.Setup()
	
	// Test GET /api/files
	t.Run("list files", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
		
		var response []dtos.AttachmentDTO
		err := json.NewDecoder(rec.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if len(response) != 1 {
			t.Errorf("Expected 1 file, got %d", len(response))
		}
	})
}

func TestRouter_Setup_FileEndpoints_Disabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup router with file uploads disabled
	router := setupTestRouter(ctrl, false)
	handler := router.Setup()
	
	// Test that file endpoints return 404 when disabled
	tests := []string{
		"/api/files",
		"/api/files/upload",
		"/api/files/test.txt",
	}
	
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			
			handler.ServeHTTP(rec, req)
			
			if rec.Code != http.StatusNotFound {
				t.Errorf("Expected status %d for disabled endpoint %s, got %d", 
					http.StatusNotFound, path, rec.Code)
			}
		})
	}
}

func TestRouter_Setup_StaticFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup router
	router := setupTestRouter(ctrl, false)
	handler := router.Setup()
	
	// Test serving static files
	req := httptest.NewRequest(http.MethodGet, "/test.html", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	
	body := rec.Body.String()
	if body != "<html>test</html>" {
		t.Errorf("Expected static file content, got %v", body)
	}
}

func TestRouter_Setup_FileHandler_Methods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	
	// Create test attachment
	fileName, _ := valueobjects.NewFileName("test.txt")
	mimeType, _ := valueobjects.NewMimeType("text/plain")
	fileSize, _ := valueobjects.NewFileSize(12)
	attachment := entities.NewAttachment(
		entities.AttachmentID("test-file"),
		entities.DefaultDocumentID,
		fileName,
		mimeType,
		fileSize,
	)
	
	attachments := make(map[entities.AttachmentID]*entities.Attachment)
	attachments[attachment.ID()] = attachment
	
	testDoc := entities.RestoreDocument(
		entities.DefaultDocumentID,
		valueobjects.EmptyContent(),
		attachments,
		valueobjects.NewTimestamp(),
		valueobjects.NewDefaultExpirationTime(),
		1,
	)
	
	// Setup expectations for download
	mockDocRepo.EXPECT().
		FindByID(gomock.Any(), entities.DefaultDocumentID).
		Return(testDoc, nil).Times(2)
	
	mockFileStorage.EXPECT().
		Retrieve(gomock.Any(), entities.AttachmentID("test-file")).
		Return(io.NopCloser(bytes.NewReader([]byte("file content"))), nil).Times(1)
	
	// Setup expectations for delete
	mockFileStorage.EXPECT().
		Delete(gomock.Any(), entities.AttachmentID("test-file")).
		Return(nil).Times(1)
	
	mockDocRepo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).Times(1)
	
	router := setupTestRouterWithFileUploads(ctrl, mockDocRepo, mockFileStorage, true)
	handler := router.Setup()
	
	// Test GET /api/files/{id}
	t.Run("download file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/files/test-file", nil)
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
		
		if rec.Body.String() != "file content" {
			t.Errorf("Expected 'file content', got %v", rec.Body.String())
		}
	})
	
	// Test DELETE /api/files/{id}
	t.Run("delete file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/files/test-file", nil)
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
	
	// Test unsupported method
	t.Run("unsupported method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/files/test-file", nil)
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})
}

func TestRouter_ConcurrentRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup router
	router := setupTestRouter(ctrl, false)
	handler := router.Setup()
	
	// Test concurrent requests to different endpoints
	done := make(chan bool, 30)
	
	// Health checks
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("Health check failed with status %d", rec.Code)
			}
			done <- true
		}()
	}
	
	// Ready checks
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("Ready check failed with status %d", rec.Code)
			}
			done <- true
		}()
	}
	
	// Static files
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test.html", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("Static file request failed with status %d", rec.Code)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 30; i++ {
		<-done
	}
}

// Helper functions

func setupTestRouter(ctrl *gomock.Controller, enableFileUploads bool) *Router {
	mockDocRepo := repositories.NewMockDocumentRepository(ctrl)
	mockFileStorage := repositories.NewMockFileStorageRepository(ctrl)
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	documentHandler := handlers.NewDocumentHandler(appService)
	
	uploadHandler := commands.NewUploadFileCommandHandler(documentService, mockDocRepo, mockFileStorage)
	deleteHandler := commands.NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)
	getFileHandler := queries.NewGetFileQueryHandler(mockDocRepo, mockFileStorage)
	listFilesHandler := queries.NewListFilesQueryHandler(mockDocRepo)
	fileHandler := handlers.NewFileHandler(uploadHandler, deleteHandler, getFileHandler, listFilesHandler)
	
	return NewRouter(documentHandler, fileHandler, testStaticFiles, enableFileUploads)
}

func setupTestRouterWithFileUploads(ctrl *gomock.Controller, mockDocRepo repositories.DocumentRepository, mockFileStorage repositories.FileStorageRepository, enable bool) *Router {
	documentService := domainservices.NewDocumentService(mockDocRepo)
	getContentHandler := queries.NewGetContentQueryHandler(documentService)
	updateContentHandler := commands.NewUpdateContentCommandHandler(documentService)
	
	appService := services.NewDocumentApplicationService(updateContentHandler, getContentHandler)
	documentHandler := handlers.NewDocumentHandler(appService)
	
	uploadHandler := commands.NewUploadFileCommandHandler(documentService, mockDocRepo, mockFileStorage)
	deleteHandler := commands.NewDeleteFileCommandHandler(mockDocRepo, mockFileStorage)
	getFileHandler := queries.NewGetFileQueryHandler(mockDocRepo, mockFileStorage)
	listFilesHandler := queries.NewListFilesQueryHandler(mockDocRepo)
	fileHandler := handlers.NewFileHandler(uploadHandler, deleteHandler, getFileHandler, listFilesHandler)
	
	return NewRouter(documentHandler, fileHandler, testStaticFiles, enable)
}