package queries

import (
	"context"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
)

type ListFilesQuery struct {
	DocumentID string
}

type ListFilesQueryHandler struct {
	documentRepo repositories.DocumentRepository
}

func NewListFilesQueryHandler(documentRepo repositories.DocumentRepository) *ListFilesQueryHandler {
	return &ListFilesQueryHandler{
		documentRepo: documentRepo,
	}
}

func (h *ListFilesQueryHandler) Handle(ctx context.Context, query ListFilesQuery) ([]*entities.Attachment, error) {
	documentID := entities.DocumentID(query.DocumentID)
	if documentID == "" {
		documentID = entities.DefaultDocumentID
	}

	document, err := h.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		if err == entities.ErrDocumentNotFound {
			return []*entities.Attachment{}, nil
		}
		return nil, err
	}

	return document.GetAttachments(), nil
}