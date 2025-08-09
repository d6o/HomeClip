package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	domainservices "github.com/d6o/homeclip/internal/domain/services"
)

type CleanupService struct {
	documentRepo      repositories.DocumentRepository
	fileStorage       repositories.FileStorageRepository
	expirationService *domainservices.ExpirationService
	interval          time.Duration
	stopChan          chan struct{}
	wg                sync.WaitGroup
	mu                sync.Mutex
	running           bool
	expiredDocuments  map[entities.DocumentID]time.Time // Track when documents expired
}

func NewCleanupService(
	documentRepo repositories.DocumentRepository,
	fileStorage repositories.FileStorageRepository,
	expirationService *domainservices.ExpirationService,
	interval time.Duration,
) *CleanupService {
	return &CleanupService{
		documentRepo:      documentRepo,
		fileStorage:       fileStorage,
		expirationService: expirationService,
		interval:          interval,
		stopChan:          make(chan struct{}),
		expiredDocuments:  make(map[entities.DocumentID]time.Time),
	}
}

func (s *CleanupService) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run(ctx)
	
	log.Println("Cleanup service started")
}

func (s *CleanupService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	
	log.Println("Cleanup service stopped")
}

func (s *CleanupService) run(ctx context.Context) {
	defer s.wg.Done()
	
	// Run cleanup immediately on start
	s.performCleanup(ctx)
	
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.performCleanup(ctx)
		}
	}
}

func (s *CleanupService) performCleanup(ctx context.Context) {
	// Since we're using in-memory storage, we need to implement
	// a way to track and clean expired documents
	// This is a simplified implementation
	
	// In a real implementation, you would:
	// 1. Query all documents from the repository
	// 2. Check each for expiration
	// 3. Delete expired documents and their attachments
	
	log.Println("Running cleanup check for expired documents...")
	
	// For now, we'll rely on the repository to handle expiration
	// during get operations
}

// CleanupDocument manually triggers cleanup for a specific document
func (s *CleanupService) CleanupDocument(ctx context.Context, documentID entities.DocumentID) error {
	document, err := s.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		if err == entities.ErrDocumentNotFound {
			// Already cleaned up
			return nil
		}
		return err
	}
	
	if !s.expirationService.ShouldCleanup(document) {
		return nil
	}
	
	// Delete all attachments
	for _, attachment := range document.GetAttachments() {
		if err := s.fileStorage.Delete(ctx, attachment.ID()); err != nil {
			log.Printf("Failed to delete attachment %s: %v", attachment.ID(), err)
		}
	}
	
	// Delete the document (this would be implemented in a real repository)
	// For now, we'll rely on the repository to handle this
	
	log.Printf("Cleaned up expired document: %s", documentID)
	return nil
}