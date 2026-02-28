package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/d6o/homeclip/internal/clipboard"
	"github.com/d6o/homeclip/internal/filestore"
)

// --- mocks ---

type mockTextStore struct {
	content clipboard.Content
	err     error
	setErr  error
	last    string
}

func (m *mockTextStore) Get(_ context.Context) (clipboard.Content, error) {
	return m.content, m.err
}

func (m *mockTextStore) Set(_ context.Context, content string) error {
	m.last = content
	return m.setErr
}

type mockFileStore struct {
	files    []filestore.Info
	listErr  error
	saveInfo filestore.Info
	saveErr  error
	path     string
	pathErr  error
	delErr   error
}

func (m *mockFileStore) Save(_ context.Context, name string, _ io.Reader, _ int64) (filestore.Info, error) {
	return m.saveInfo, m.saveErr
}

func (m *mockFileStore) List(_ context.Context) ([]filestore.Info, error) {
	return m.files, m.listErr
}

func (m *mockFileStore) FilePath(_ string) (string, error) {
	return m.path, m.pathErr
}

func (m *mockFileStore) Delete(_ context.Context, _ string) error {
	return m.delErr
}

// --- helpers ---

func newTestServer(text textStore, file fileStore) *Server {
	return NewServer("0", text, file)
}

func setupMux(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/text", s.handleGetText)
	mux.HandleFunc("PUT /api/text", s.handleSetText)
	mux.HandleFunc("POST /api/files", s.handleUploadFile)
	mux.HandleFunc("GET /api/files", s.handleListFiles)
	mux.HandleFunc("GET /api/files/{filename}", s.handleDownloadFile)
	mux.HandleFunc("DELETE /api/files/{filename}", s.handleDeleteFile)
	return mux
}

// --- GET /api/text ---

func TestHandleGetText_Empty(t *testing.T) {
	ts := &mockTextStore{err: clipboard.ErrEmpty}
	s := newTestServer(ts, &mockFileStore{})
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/text", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var c clipboard.Content
	if err := json.NewDecoder(w.Body).Decode(&c); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if c.Content != "" {
		t.Errorf("expected empty content, got %q", c.Content)
	}
}

func TestHandleGetText_WithContent(t *testing.T) {
	ts := &mockTextStore{
		content: clipboard.Content{Content: "hello", UpdatedAt: time.Now()},
	}
	s := newTestServer(ts, &mockFileStore{})
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/text", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var c clipboard.Content
	if err := json.NewDecoder(w.Body).Decode(&c); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if c.Content != "hello" {
		t.Errorf("expected content %q, got %q", "hello", c.Content)
	}
}

func TestHandleGetText_Error(t *testing.T) {
	ts := &mockTextStore{err: errors.New("disk error")}
	s := newTestServer(ts, &mockFileStore{})
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/text", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- PUT /api/text ---

func TestHandleSetText_Success(t *testing.T) {
	ts := &mockTextStore{}
	s := newTestServer(ts, &mockFileStore{})
	mux := setupMux(s)

	body := bytes.NewBufferString("new clipboard content")
	req := httptest.NewRequest(http.MethodPut, "/api/text", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if ts.last != "new clipboard content" {
		t.Errorf("expected last set to %q, got %q", "new clipboard content", ts.last)
	}
}

func TestHandleSetText_StoreError(t *testing.T) {
	ts := &mockTextStore{setErr: errors.New("write error")}
	s := newTestServer(ts, &mockFileStore{})
	mux := setupMux(s)

	body := bytes.NewBufferString("data")
	req := httptest.NewRequest(http.MethodPut, "/api/text", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- GET /api/files ---

func TestHandleListFiles_Empty(t *testing.T) {
	fs := &mockFileStore{files: nil}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var files []filestore.Info
	if err := json.NewDecoder(w.Body).Decode(&files); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestHandleListFiles_WithFiles(t *testing.T) {
	fs := &mockFileStore{
		files: []filestore.Info{
			{Name: "a.txt", Size: 100, UploadedAt: time.Now()},
			{Name: "b.txt", Size: 200, UploadedAt: time.Now()},
		},
	}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var files []filestore.Info
	if err := json.NewDecoder(w.Body).Decode(&files); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestHandleListFiles_Error(t *testing.T) {
	fs := &mockFileStore{listErr: errors.New("list error")}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- POST /api/files ---

func TestHandleUploadFile_Success(t *testing.T) {
	info := filestore.Info{Name: "upload.txt", Size: 12, UploadedAt: time.Now()}
	fs := &mockFileStore{saveInfo: info}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "upload.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fw.Write([]byte("file content"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var result filestore.Info
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Name != "upload.txt" {
		t.Errorf("expected name %q, got %q", "upload.txt", result.Name)
	}
}

func TestHandleUploadFile_TooLarge(t *testing.T) {
	fs := &mockFileStore{saveErr: filestore.ErrTooLarge}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "big.bin")
	fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", w.Code)
	}
}

func TestHandleUploadFile_SaveError(t *testing.T) {
	fs := &mockFileStore{saveErr: errors.New("save error")}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "fail.txt")
	fw.Write([]byte("data"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/files", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestHandleUploadFile_NoFile(t *testing.T) {
	s := newTestServer(&mockTextStore{}, &mockFileStore{})
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodPost, "/api/files", bytes.NewBufferString("not multipart"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- DELETE /api/files/{filename} ---

func TestHandleDeleteFile_Success(t *testing.T) {
	fs := &mockFileStore{}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodDelete, "/api/files/test.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestHandleDeleteFile_NotFound(t *testing.T) {
	fs := &mockFileStore{delErr: filestore.ErrNotFound}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodDelete, "/api/files/ghost.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleDeleteFile_Error(t *testing.T) {
	fs := &mockFileStore{delErr: errors.New("delete error")}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodDelete, "/api/files/err.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- GET /api/files/{filename} ---

func TestHandleDownloadFile_NotFound(t *testing.T) {
	fs := &mockFileStore{pathErr: filestore.ErrNotFound}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files/missing.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleDownloadFile_PathError(t *testing.T) {
	fs := &mockFileStore{pathErr: errors.New("path error")}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files/err.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestHandleDownloadFile_Success(t *testing.T) {
	dir := t.TempDir()
	tmpFile := dir + "/hello.txt"
	if err := writeFile(tmpFile, "file body"); err != nil {
		t.Fatal(err)
	}

	fs := &mockFileStore{path: tmpFile}
	s := newTestServer(&mockTextStore{}, fs)
	mux := setupMux(s)

	req := httptest.NewRequest(http.MethodGet, "/api/files/hello.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Disposition"); ct == "" {
		t.Error("expected Content-Disposition header")
	}
}

// --- NewServer ---

func TestNewServer(t *testing.T) {
	s := NewServer("8080", &mockTextStore{}, &mockFileStore{})
	if s.addr != ":8080" {
		t.Errorf("expected addr %q, got %q", ":8080", s.addr)
	}
}

// helper
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
