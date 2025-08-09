package commands

import (
	"context"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
)

type DeleteFileCommand struct {
	DocumentID   string
	AttachmentID string
}

type DeleteFileCommandHandler struct {
	documentRepo    repositories.DocumentRepository
	fileStorageRepo repositories.FileStorageRepository
}

func NewDeleteFileCommandHandler(
	documentRepo repositories.DocumentRepository,
	fileStorageRepo repositories.FileStorageRepository,
) *DeleteFileCommandHandler {
	return &DeleteFileCommandHandler{
		documentRepo:    documentRepo,
		fileStorageRepo: fileStorageRepo,
	}
}

func (h *DeleteFileCommandHandler) Handle(ctx context.Context, cmd DeleteFileCommand) error {
	documentID := entities.DocumentID(cmd.DocumentID)
	if documentID == "" {
		documentID = entities.DefaultDocumentID
	}

	document, err := h.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return err
	}

	attachmentID := entities.AttachmentID(cmd.AttachmentID)

	if err := document.RemoveAttachment(attachmentID); err != nil {
		return err
	}

	if err := h.fileStorageRepo.Delete(ctx, attachmentID); err != nil {
	}

	return h.documentRepo.Save(ctx, document)
}
