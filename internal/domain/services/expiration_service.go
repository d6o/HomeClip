package services

import (
	"context"
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

// CleanExpiredDocuments removes expired documents and their attachments
func (s *ExpirationService) CleanExpiredDocuments(ctx context.Context) (int, error) {
	// This would typically query for all expired documents
	// For now, we'll handle this at the repository level
	// In a real implementation, you'd have a method like:
	// documents, err := s.documentRepo.FindExpired(ctx)
	
	// The actual cleanup will be handled by the repository implementation
	return 0, nil
}

// IsDocumentExpired checks if a document has expired
func (s *ExpirationService) IsDocumentExpired(document *entities.Document) bool {
	return document.IsExpired()
}

// ShouldCleanup determines if a document should be cleaned up
func (s *ExpirationService) ShouldCleanup(document *entities.Document) bool {
	// Clean up if expired for more than 1 hour (grace period)
	gracePeriod := 1 * time.Hour
	expirationTime := document.ExpiresAt().Value()
	return time.Now().UTC().After(expirationTime.Add(gracePeriod))
}

// ValidateAccess checks if a document can be accessed
func (s *ExpirationService) ValidateAccess(document *entities.Document) error {
	if document.IsExpired() {
		return valueobjects.ErrExpired
	}
	return nil
}

// ExtendExpiration extends the expiration time of a document
func (s *ExpirationService) ExtendExpiration(ctx context.Context, documentID entities.DocumentID, duration time.Duration) error {
	document, err := s.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return err
	}
	
	if document.IsExpired() {
		return valueobjects.ErrExpired
	}
	
	// Update expiration time
	newExpiration := document.ExpiresAt().ExtendBy(duration)
	
	// Create updated document with new expiration
	updatedDocument := entities.RestoreDocument(
		document.ID(),
		document.Content(),
		nil, // attachments will be handled by the repository
		document.LastUpdated(),
		newExpiration,
		document.Version() + 1,
	)
	
	// Copy attachments
	for _, attachment := range document.GetAttachments() {
		updatedDocument.AddAttachment(attachment)
	}
	
	return s.documentRepo.Save(ctx, updatedDocument)
}