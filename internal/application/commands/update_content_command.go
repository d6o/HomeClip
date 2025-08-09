package commands

import (
	"context"

	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/services"
)

type UpdateContentCommand struct {
	DocumentID string
	Content    string
}

type UpdateContentCommandHandler struct {
	documentService services.DocumentServiceInterface
}

func NewUpdateContentCommandHandler(documentService services.DocumentServiceInterface) *UpdateContentCommandHandler {
	return &UpdateContentCommandHandler{
		documentService: documentService,
	}
}

func (h *UpdateContentCommandHandler) Handle(ctx context.Context, cmd UpdateContentCommand) (*dtos.UpdateContentResponse, error) {
	documentID := entities.DocumentID(cmd.DocumentID)
	if documentID == "" {
		documentID = entities.DefaultDocumentID
	}

	document, err := h.documentService.UpdateDocumentContent(ctx, documentID, cmd.Content)
	if err != nil {
		return &dtos.UpdateContentResponse{
			Success: false,
		}, err
	}

	return &dtos.UpdateContentResponse{
		Success:     true,
		LastUpdated: document.LastUpdated().Value(),
		Version:     document.Version(),
	}, nil
}
