package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

type UploadFileCommand struct {
	DocumentID string
	FileName   string
	MimeType   string
	Size       int64
	Reader     io.Reader
}

type UploadFileCommandHandler struct {
	documentService services.DocumentServiceInterface
	documentRepo    repositories.DocumentRepository
	fileStorageRepo repositories.FileStorageRepository
}

func NewUploadFileCommandHandler(
	documentService services.DocumentServiceInterface,
	documentRepo repositories.DocumentRepository,
	fileStorageRepo repositories.FileStorageRepository,
) *UploadFileCommandHandler {
	return &UploadFileCommandHandler{
		documentService: documentService,
		documentRepo:    documentRepo,
		fileStorageRepo: fileStorageRepo,
	}
}

func (h *UploadFileCommandHandler) Handle(ctx context.Context, cmd UploadFileCommand) (*entities.Attachment, error) {
	fileName, err := valueobjects.NewFileName(cmd.FileName)
	if err != nil {
		return nil, err
	}

	mimeType, err := valueobjects.NewMimeType(cmd.MimeType)
	if err != nil {
		return nil, err
	}

	size, err := valueobjects.NewFileSize(cmd.Size)
	if err != nil {
		return nil, err
	}

	attachmentID := entities.AttachmentID(generateID())

	documentID := entities.DocumentID(cmd.DocumentID)
	if documentID == "" {
		documentID = entities.DefaultDocumentID
	}

	document, err := h.documentService.GetOrCreateDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}

	attachment := entities.NewAttachment(
		attachmentID,
		documentID,
		fileName,
		mimeType,
		size,
	)

	if err := h.fileStorageRepo.Store(ctx, attachmentID, cmd.Reader); err != nil {
		return nil, err
	}

	if err := document.AddAttachment(attachment); err != nil {
		h.fileStorageRepo.Delete(ctx, attachmentID)
		return nil, err
	}

	if err := h.documentRepo.Save(ctx, document); err != nil {
		h.fileStorageRepo.Delete(ctx, attachmentID)
		return nil, err
	}

	return attachment, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
