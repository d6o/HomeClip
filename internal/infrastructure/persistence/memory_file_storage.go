package persistence

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

type MemoryFileStorage struct {
	mu    sync.RWMutex
	files map[entities.AttachmentID][]byte
}

func NewMemoryFileStorage() repositories.FileStorageRepository {
	return &MemoryFileStorage{
		files: make(map[entities.AttachmentID][]byte),
	}
}

func (s *MemoryFileStorage) Store(ctx context.Context, attachmentID entities.AttachmentID, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.files[attachmentID] = data
	return nil
}

func (s *MemoryFileStorage) Retrieve(ctx context.Context, attachmentID entities.AttachmentID) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.files[attachmentID]
	if !exists {
		return nil, ErrFileNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *MemoryFileStorage) Delete(ctx context.Context, attachmentID entities.AttachmentID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Idempotent delete - no error if file doesn't exist
	delete(s.files, attachmentID)
	return nil
}

func (s *MemoryFileStorage) Exists(ctx context.Context, attachmentID entities.AttachmentID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.files[attachmentID]
	return exists, nil
}