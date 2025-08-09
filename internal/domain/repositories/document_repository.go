package repositories

import (
	"context"
	
	"github.com/d6o/homeclip/internal/domain/entities"
)

//go:generate go tool mockgen -source=document_repository.go -destination=document_repository_mock.go -package=repositories
type DocumentRepository interface {
	FindByID(ctx context.Context, id entities.DocumentID) (*entities.Document, error)
	Save(ctx context.Context, document *entities.Document) error
	Exists(ctx context.Context, id entities.DocumentID) (bool, error)
}