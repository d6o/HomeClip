package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/d6o/homeclip/internal/clipboard"
	"github.com/d6o/homeclip/internal/filestore"
)

//go:embed static
var staticFiles embed.FS

type textStore interface {
	Get(ctx context.Context) (clipboard.Content, error)
	Set(ctx context.Context, content string) error
}

type fileStore interface {
	Save(ctx context.Context, name string, r io.Reader, size int64) (filestore.Info, error)
	List(ctx context.Context) ([]filestore.Info, error)
	FilePath(name string) (string, error)
	Delete(ctx context.Context, name string) error
}

type Server struct {
	text textStore
	file fileStore
	addr string
}

func NewServer(port string, text textStore, file fileStore) *Server {
	return &Server{
		text: text,
		file: file,
		addr: net.JoinHostPort("", port),
	}
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/text", s.handleGetText)
	mux.HandleFunc("PUT /api/text", s.handleSetText)
	mux.HandleFunc("POST /api/files", s.handleUploadFile)
	mux.HandleFunc("GET /api/files", s.handleListFiles)
	mux.HandleFunc("GET /api/files/{filename}", s.handleDownloadFile)
	mux.HandleFunc("DELETE /api/files/{filename}", s.handleDeleteFile)

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	mux.Handle("GET /", http.FileServerFS(staticFS))

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info("server listening", "addr", s.addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) handleGetText(w http.ResponseWriter, r *http.Request) {
	content, err := s.text.Get(r.Context())
	if err != nil {
		if errors.Is(err, clipboard.ErrEmpty) {
			s.writeJSON(w, http.StatusOK, clipboard.Content{})
			return
		}
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, content)
}

func (s *Server) handleSetText(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.text.Set(r.Context(), string(body)); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024*1024+1024)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	info, err := s.file.Save(r.Context(), header.Filename, file, header.Size)
	if err != nil {
		if errors.Is(err, filestore.ErrTooLarge) {
			s.writeError(w, http.StatusRequestEntityTooLarge, err)
			return
		}
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusCreated, info)
}

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := s.file.List(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	if files == nil {
		files = []filestore.Info{}
	}

	s.writeJSON(w, http.StatusOK, files)
}

func (s *Server) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	path, err := s.file.FilePath(filename)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	http.ServeFile(w, r, path)
}

func (s *Server) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	if err := s.file.Delete(r.Context(), filename); err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	slog.Error("request error", "status", status, "error", err)
	http.Error(w, err.Error(), status)
}
