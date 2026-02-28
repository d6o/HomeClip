package filestore

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxFileSize = 100 * 1024 * 1024 // 100 MB

var (
	ErrTooLarge = errors.New("file exceeds 100 MB limit")
	ErrNotFound = errors.New("file not found")
)

type Store struct {
	dir string
	mu  sync.RWMutex
}

func NewStore(dataDir string) (*Store, error) {
	dir := filepath.Join(dataDir, "files")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return &Store{dir: dir}, nil
}

func (s *Store) Save(_ context.Context, name string, r io.Reader, size int64) (Info, error) {
	if size > maxFileSize {
		return Info{}, ErrTooLarge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dest := filepath.Join(s.dir, filepath.Base(name))

	f, err := os.Create(dest)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()

	limited := io.LimitReader(r, maxFileSize+1)

	written, err := io.Copy(f, limited)
	if err != nil {
		os.Remove(dest)
		return Info{}, err
	}

	if written > maxFileSize {
		os.Remove(dest)
		return Info{}, ErrTooLarge
	}

	stat, err := f.Stat()
	if err != nil {
		return Info{}, err
	}

	return Info{
		Name:       filepath.Base(name),
		Size:       written,
		UploadedAt: stat.ModTime(),
	}, nil
}

func (s *Store) List(_ context.Context) ([]Info, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	files := make([]Info, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		files = append(files, Info{
			Name:       e.Name(),
			Size:       info.Size(),
			UploadedAt: info.ModTime(),
		})
	}

	return files, nil
}

func (s *Store) FilePath(name string) (string, error) {
	clean := filepath.Base(name)
	full := filepath.Join(s.dir, clean)

	if _, err := os.Stat(full); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNotFound
		}
		return "", err
	}

	return full, nil
}

func (s *Store) Delete(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clean := filepath.Base(name)
	full := filepath.Join(s.dir, clean)

	if _, err := os.Stat(full); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return err
	}

	return os.Remove(full)
}

func (s *Store) Cleanup(_ context.Context, maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			os.Remove(filepath.Join(s.dir, e.Name()))
		}
	}

	return nil
}
