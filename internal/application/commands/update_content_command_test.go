package commands

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

func TestUpdateContentCommandHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	content := "test content"

	document := entities.NewDocument(documentID)
	contentVO, err := valueobjects.NewContent(content)
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}
	if err := document.UpdateContent(contentVO); err != nil {
		t.Fatalf("Failed to update content: %v", err)
	}

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content).
		Return(document, nil)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Version != document.Version() {
		t.Errorf("Expected version %d, got %d", document.Version(), response.Version)
	}
}

func TestUpdateContentCommandHandler_Handle_EmptyDocumentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	content := "test content"

	document := entities.NewDocument(entities.DefaultDocumentID)
	contentVO, err := valueobjects.NewContent(content)
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}
	if err := document.UpdateContent(contentVO); err != nil {
		t.Fatalf("Failed to update content: %v", err)
	}

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, entities.DefaultDocumentID, content).
		Return(document, nil)

	cmd := UpdateContentCommand{
		DocumentID: "",
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)
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

func TestUpdateContentCommandHandler_Handle_ContentTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	content := "test content"

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content).
		Return(nil, valueobjects.ErrContentTooLarge)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)

	if !errors.Is(err, valueobjects.ErrContentTooLarge) {
		t.Errorf("Expected ErrContentTooLarge, got %v", err)
	}

	expectedResponse := &dtos.UpdateContentResponse{Success: false}
	if response.Success != expectedResponse.Success {
		t.Errorf("Expected response success to be false, got %v", response.Success)
	}
}

func TestUpdateContentCommandHandler_Handle_InvalidContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	content := ""

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content).
		Return(nil, valueobjects.ErrInvalidContent)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)

	if !errors.Is(err, valueobjects.ErrInvalidContent) {
		t.Errorf("Expected ErrInvalidContent, got %v", err)
	}

	if response.Success != false {
		t.Error("Expected response success to be false")
	}
}

func TestUpdateContentCommandHandler_Handle_ExpiredDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	content := "test content"

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content).
		Return(nil, valueobjects.ErrExpired)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)

	if !errors.Is(err, valueobjects.ErrExpired) {
		t.Errorf("Expected ErrExpired, got %v", err)
	}

	if response.Success != false {
		t.Error("Expected response success to be false")
	}
}

func TestUpdateContentCommandHandler_Handle_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	content := "test content"

	serviceErr := errors.New("service error")
	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content).
		Return(nil, serviceErr)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)

	if !errors.Is(err, serviceErr) {
		t.Errorf("Expected service error %v, got %v", serviceErr, err)
	}

	if response.Success != false {
		t.Error("Expected response success to be false")
	}
}

func TestUpdateContentCommandHandler_Handle_LargeContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	largeContent := string(make([]byte, valueobjects.MaxContentLength-1))

	document := entities.NewDocument(documentID)
	contentVO, err := valueobjects.NewContent(largeContent)
	if err != nil {
		t.Fatalf("Failed to create large content: %v", err)
	}
	if err := document.UpdateContent(contentVO); err != nil {
		t.Fatalf("Failed to update content: %v", err)
	}

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, largeContent).
		Return(document, nil)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    largeContent,
	}

	response, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error for content at max size, got %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestUpdateContentCommandHandler_Handle_EmptyContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")
	content := ""

	document := entities.NewDocument(documentID)

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content).
		Return(document, nil)

	cmd := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content,
	}

	response, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error for empty content, got %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestUpdateContentCommandHandler_Handle_MultipleUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockDocService := services.NewMockDocumentServiceInterface(ctrl)
	handler := NewUpdateContentCommandHandler(mockDocService)

	ctx := t.Context()
	documentID := entities.DocumentID("test-doc")

	document := entities.NewDocument(documentID)

	content1 := "first content"
	contentVO1, err := valueobjects.NewContent(content1)
	if err != nil {
		t.Fatalf("Failed to create content1: %v", err)
	}
	if err := document.UpdateContent(contentVO1); err != nil {
		t.Fatalf("Failed to update content: %v", err)
	}

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content1).
		Return(document, nil)

	cmd1 := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content1,
	}

	response1, err := handler.Handle(ctx, cmd1)
	if err != nil {
		t.Fatalf("First update failed: %v", err)
	}

	content2 := "second content"
	contentVO2, err := valueobjects.NewContent(content2)
	if err != nil {
		t.Fatalf("Failed to create content2: %v", err)
	}
	if err := document.UpdateContent(contentVO2); err != nil {
		t.Fatalf("Failed to update content: %v", err)
	}

	mockDocService.EXPECT().
		UpdateDocumentContent(ctx, documentID, content2).
		Return(document, nil)

	cmd2 := UpdateContentCommand{
		DocumentID: string(documentID),
		Content:    content2,
	}

	response2, err := handler.Handle(ctx, cmd2)
	if err != nil {
		t.Fatalf("Second update failed: %v", err)
	}

	if response2.Version <= response1.Version {
		t.Errorf("Expected version to increment, got %d after %d", response2.Version, response1.Version)
	}
}
