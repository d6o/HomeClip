package queries

import (
	"context"
	"io"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
)

type GetFileQuery struct {
	DocumentID   string
	AttachmentID string
}

type FileResult struct {
	Attachment *entities.Attachment
	Reader     io.ReadCloser
}

type GetFileQueryHandler struct {
	documentRepo    repositories.DocumentRepository
	fileStorageRepo repositories.FileStorageRepository
}

func NewGetFileQueryHandler(
	documentRepo repositories.DocumentRepository,
	fileStorageRepo repositories.FileStorageRepository,
) *GetFileQueryHandler {
	return &GetFileQueryHandler{
		documentRepo:    documentRepo,
		fileStorageRepo: fileStorageRepo,
	}
}

func (h *GetFileQueryHandler) Handle(ctx context.Context, query GetFileQuery) (*FileResult, error) {
	documentID := entities.DocumentID(query.DocumentID)
	if documentID == "" {
		documentID = entities.DefaultDocumentID
	}

	document, err := h.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	attachmentID := entities.AttachmentID(query.AttachmentID)
	attachment, err := document.GetAttachment(attachmentID)
	if err != nil {
		return nil, err
	}

	reader, err := h.fileStorageRepo.Retrieve(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	return &FileResult{
		Attachment: attachment,
		Reader:     reader,
	}, nil
}
