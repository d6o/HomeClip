package persistence

import (
	"context"
	"sync"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
)

type MemoryDocumentRepository struct {
	mu        sync.RWMutex
	documents map[entities.DocumentID]*entities.Document
}

func NewMemoryDocumentRepository() repositories.DocumentRepository {
	return &MemoryDocumentRepository{
		documents: make(map[entities.DocumentID]*entities.Document),
	}
}

func (r *MemoryDocumentRepository) FindByID(ctx context.Context, id entities.DocumentID) (*entities.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	document, exists := r.documents[id]
	if !exists {
		return nil, entities.ErrDocumentNotFound
	}

	// Check if document is expired
	if document.IsExpired() {
		// For expired documents, we still return them but the domain layer will handle the expiration
		// This allows for grace period handling
	}

	return r.cloneDocument(document), nil
}

func (r *MemoryDocumentRepository) Save(ctx context.Context, document *entities.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.documents[document.ID()] = r.cloneDocument(document)
	return nil
}

func (r *MemoryDocumentRepository) Exists(ctx context.Context, id entities.DocumentID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.documents[id]
	return exists, nil
}

func (r *MemoryDocumentRepository) cloneDocument(doc *entities.Document) *entities.Document {
	// Clone attachments map
	attachments := make(map[entities.AttachmentID]*entities.Attachment)
	for _, att := range doc.GetAttachments() {
		attachments[att.ID()] = att
	}
	
	return entities.RestoreDocument(
		doc.ID(),
		doc.Content(),
		attachments,
		doc.LastUpdated(),
		doc.ExpiresAt(),
		doc.Version(),
	)
}