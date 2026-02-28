package filestore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	return s
}

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	expected := filepath.Join(dir, "files")
	if s.dir != expected {
		t.Errorf("expected dir %q, got %q", expected, s.dir)
	}

	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("expected files directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected files path to be a directory")
	}
}

func TestStore_SaveAndList(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	info, err := s.Save(ctx, "test.txt", strings.NewReader("hello"), 5)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if info.Name != "test.txt" {
		t.Errorf("expected name %q, got %q", "test.txt", info.Name)
	}
	if info.Size != 5 {
		t.Errorf("expected size 5, got %d", info.Size)
	}

	files, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != "test.txt" {
		t.Errorf("expected name %q, got %q", "test.txt", files[0].Name)
	}
}

func TestStore_SaveTooLarge(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "big.bin", strings.NewReader("data"), maxFileSize+1)
	if err != ErrTooLarge {
		t.Errorf("expected ErrTooLarge, got %v", err)
	}
}

func TestStore_SavePathTraversal(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	info, err := s.Save(ctx, "../../../etc/passwd", strings.NewReader("data"), 4)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if info.Name != "passwd" {
		t.Errorf("expected sanitized name %q, got %q", "passwd", info.Name)
	}
}

func TestStore_FilePath(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "doc.pdf", strings.NewReader("pdf content"), 11)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path, err := s.FilePath("doc.pdf")
	if err != nil {
		t.Fatalf("FilePath failed: %v", err)
	}

	if filepath.Base(path) != "doc.pdf" {
		t.Errorf("expected path ending in doc.pdf, got %q", path)
	}
}

func TestStore_FilePathNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.FilePath("nonexistent.txt")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "remove.txt", strings.NewReader("bye"), 3)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := s.Delete(ctx, "remove.txt"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = s.FilePath("remove.txt")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestStore_DeleteNotFound(t *testing.T) {
	s := newTestStore(t)

	err := s.Delete(context.Background(), "ghost.txt")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_ListEmpty(t *testing.T) {
	s := newTestStore(t)

	files, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestStore_ListSkipsDirs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "file.txt", strings.NewReader("data"), 4)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := os.Mkdir(filepath.Join(s.dir, "subdir"), 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	files, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (dir skipped), got %d", len(files))
	}
}

func TestStore_CleanupExpired(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "old.txt", strings.NewReader("old"), 3)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(filepath.Join(s.dir, "old.txt"), oldTime, oldTime)

	if err := s.Cleanup(ctx, time.Hour); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	files, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files after cleanup, got %d", len(files))
	}
}

func TestStore_CleanupKeepsFresh(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "fresh.txt", strings.NewReader("new"), 3)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := s.Cleanup(ctx, time.Hour); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	files, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file kept, got %d", len(files))
	}
}

func TestStore_CleanupMixed(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.Save(ctx, "old.txt", strings.NewReader("old"), 3)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(filepath.Join(s.dir, "old.txt"), oldTime, oldTime)

	_, err = s.Save(ctx, "new.txt", strings.NewReader("new"), 3)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := s.Cleanup(ctx, time.Hour); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	files, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != "new.txt" {
		t.Errorf("expected new.txt to remain, got %q", files[0].Name)
	}
}
