package repositories

import (
	"context"
	"io"

	"github.com/d6o/homeclip/internal/domain/entities"
)

//go:generate go tool mockgen -source=file_storage_repository.go -destination=file_storage_repository_mock.go -package=repositories
type FileStorageRepository interface {
	Store(ctx context.Context, attachmentID entities.AttachmentID, reader io.Reader) error
	Retrieve(ctx context.Context, attachmentID entities.AttachmentID) (io.ReadCloser, error)
	Delete(ctx context.Context, attachmentID entities.AttachmentID) error
	Exists(ctx context.Context, attachmentID entities.AttachmentID) (bool, error)
}
