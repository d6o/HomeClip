package queries

import (
	"context"

	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/services"
)

type GetContentQuery struct {
	DocumentID string
}

type GetContentQueryHandler struct {
	documentService services.DocumentServiceInterface
}

func NewGetContentQueryHandler(documentService services.DocumentServiceInterface) *GetContentQueryHandler {
	return &GetContentQueryHandler{
		documentService: documentService,
	}
}

func (h *GetContentQueryHandler) Handle(ctx context.Context, query GetContentQuery) (*dtos.GetContentResponse, error) {
	documentID := entities.DocumentID(query.DocumentID)
	if documentID == "" {
		documentID = entities.DefaultDocumentID
	}

	document, err := h.documentService.GetOrCreateDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}

	attachments := document.GetAttachments()
	dtoAttachments := make([]dtos.AttachmentDTO, 0, len(attachments))
	for _, att := range attachments {
		if !att.IsExpired() {
			dtoAttachments = append(dtoAttachments, dtos.AttachmentDTO{
				ID:         string(att.ID()),
				FileName:   att.FileName().Value(),
				MimeType:   att.MimeType().Value(),
				Size:       att.Size().Value(),
				UploadedAt: att.UploadedAt().Value(),
				ExpiresAt:  att.ExpiresAt().Value(),
			})
		}
	}

	return &dtos.GetContentResponse{
		Content:     document.Content().Value(),
		LastUpdated: document.LastUpdated().Value(),
		ExpiresAt:   document.ExpiresAt().Value(),
		Version:     document.Version(),
		Attachments: dtoAttachments,
	}, nil
}
