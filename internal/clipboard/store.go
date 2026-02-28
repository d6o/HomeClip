package clipboard

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrEmpty = errors.New("clipboard is empty")

type Store struct {
	filePath string
	mu       sync.RWMutex
}

func NewStore(dataDir string) *Store {
	return &Store{
		filePath: filepath.Join(dataDir, "clipboard.json"),
	}
}

func (s *Store) Get(_ context.Context) (Content, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Content{}, ErrEmpty
		}
		return Content{}, err
	}

	var c Content
	if err := json.Unmarshal(data, &c); err != nil {
		return Content{}, err
	}

	return c, nil
}

func (s *Store) Set(_ context.Context, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := Content{
		Content:   content,
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0o644)
}

func (s *Store) Cleanup(_ context.Context, maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var c Content
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}

	if time.Since(c.UpdatedAt) > maxAge {
		return os.Remove(s.filePath)
	}

	return nil
}
