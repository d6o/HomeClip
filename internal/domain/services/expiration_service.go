package services

import (
	"time"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

type ExpirationService struct {
	documentRepo repositories.DocumentRepository
	fileStorage  repositories.FileStorageRepository
}

func NewExpirationService(
	documentRepo repositories.DocumentRepository,
	fileStorage repositories.FileStorageRepository,
) *ExpirationService {
	return &ExpirationService{
		documentRepo: documentRepo,
		fileStorage:  fileStorage,
	}
}

func (s *ExpirationService) IsDocumentExpired(document *entities.Document) bool {
	return document.IsExpired()
}

func (s *ExpirationService) ShouldCleanup(document *entities.Document) bool {
	gracePeriod := 1 * time.Hour
	expirationTime := document.ExpiresAt().Value()
	return time.Now().UTC().After(expirationTime.Add(gracePeriod))
}

func (s *ExpirationService) ValidateAccess(document *entities.Document) error {
	if document.IsExpired() {
		return valueobjects.ErrExpired
	}
	return nil
}
