package services

import (
	"context"
	"errors"
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
	expiredDocuments  map[entities.DocumentID]time.Time
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
	log.Println("Running cleanup check for expired documents...")
}

func (s *CleanupService) CleanupDocument(ctx context.Context, documentID entities.DocumentID) error {
	document, err := s.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		if errors.Is(err, entities.ErrDocumentNotFound) {
			return nil
		}
		return err
	}

	if !s.expirationService.ShouldCleanup(document) {
		return nil
	}

	for _, attachment := range document.GetAttachments() {
		if err := s.fileStorage.Delete(ctx, attachment.ID()); err != nil {
			log.Printf("Failed to delete attachment %s: %v", attachment.ID(), err)
		}
	}

	log.Printf("Cleaned up expired document: %s", documentID)
	return nil
}
