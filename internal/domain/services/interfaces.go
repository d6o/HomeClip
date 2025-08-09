package services

import (
	"context"
	"time"

	"github.com/d6o/homeclip/internal/domain/entities"
)

//go:generate go tool mockgen -source=interfaces.go -destination=services_mock.go -package=services

type DocumentServiceInterface interface {
	GetOrCreateDocument(ctx context.Context, id entities.DocumentID) (*entities.Document, error)
	UpdateDocumentContent(ctx context.Context, id entities.DocumentID, contentValue string) (*entities.Document, error)
}

type ExpirationServiceInterface interface {
	IsExpired(createdAt time.Time) bool
	ExpiresAt(createdAt time.Time) time.Time
	TimeUntilExpiration(createdAt time.Time) time.Duration
}
