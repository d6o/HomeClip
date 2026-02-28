package clipboard

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	s := NewStore("/tmp/testdata")
	if s.filePath != filepath.Join("/tmp/testdata", "clipboard.json") {
		t.Errorf("unexpected file path: %s", s.filePath)
	}
}

func TestStore_GetEmpty(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	_, err := s.Get(context.Background())
	if err != ErrEmpty {
		t.Errorf("expected ErrEmpty, got %v", err)
	}
}

func TestStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	ctx := context.Background()

	if err := s.Set(ctx, "hello world"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	c, err := s.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if c.Content != "hello world" {
		t.Errorf("expected content %q, got %q", "hello world", c.Content)
	}

	if c.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestStore_SetOverwrite(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	ctx := context.Background()

	if err := s.Set(ctx, "first"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := s.Set(ctx, "second"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	c, err := s.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if c.Content != "second" {
		t.Errorf("expected content %q, got %q", "second", c.Content)
	}
}

func TestStore_GetInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	if err := os.WriteFile(s.filePath, []byte("not json"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := s.Get(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestStore_CleanupNoFile(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	err := s.Cleanup(context.Background(), time.Hour)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestStore_CleanupFresh(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	ctx := context.Background()

	if err := s.Set(ctx, "keep me"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := s.Cleanup(ctx, time.Hour); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	c, err := s.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed after cleanup: %v", err)
	}

	if c.Content != "keep me" {
		t.Errorf("expected content %q, got %q", "keep me", c.Content)
	}
}

func TestStore_CleanupExpired(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	ctx := context.Background()

	old := Content{
		Content:   "old content",
		UpdatedAt: time.Now().Add(-2 * time.Hour),
	}
	data, _ := json.Marshal(old)
	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if err := s.Cleanup(ctx, time.Hour); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	_, err := s.Get(ctx)
	if err != ErrEmpty {
		t.Errorf("expected ErrEmpty after cleanup, got %v", err)
	}
}

func TestStore_CleanupInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	if err := os.WriteFile(s.filePath, []byte("bad"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	err := s.Cleanup(context.Background(), time.Hour)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
