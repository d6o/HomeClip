package router

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/d6o/homeclip/internal/infrastructure/http/handlers"
)

type Router struct {
	mux             *http.ServeMux
	documentHandler *handlers.DocumentHandler
	fileHandler     *handlers.FileHandler
	healthHandler   *handlers.HealthHandler
	staticFiles     embed.FS
	enableFileUploads bool
}

func NewRouter(documentHandler *handlers.DocumentHandler, fileHandler *handlers.FileHandler, staticFiles embed.FS, enableFileUploads bool) *Router {
	return &Router{
		mux:             http.NewServeMux(),
		documentHandler: documentHandler,
		fileHandler:     fileHandler,
		healthHandler:   handlers.NewHealthHandler(),
		staticFiles:     staticFiles,
		enableFileUploads: enableFileUploads,
	}
}

func (r *Router) Setup() http.Handler {
	// Health check endpoints
	r.mux.HandleFunc("/api/health", r.healthHandler.Health)
	r.mux.HandleFunc("/api/ready", r.healthHandler.Ready)
	
	// Document endpoints
	r.mux.HandleFunc("/api/content", r.documentHandler.HandleContent)
	
	// File endpoints (if enabled)
	if r.enableFileUploads {
		r.mux.HandleFunc("/api/files", r.fileHandler.ListFiles)
		r.mux.HandleFunc("/api/files/upload", r.fileHandler.UploadFile)
		r.mux.HandleFunc("/api/files/", func(w http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodGet:
				r.fileHandler.DownloadFile(w, req)
			case http.MethodDelete:
				r.fileHandler.DeleteFile(w, req)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
	}
	
	// Static files
	staticFS, err := fs.Sub(r.staticFiles, "static")
	if err != nil {
		panic(err)
	}
	r.mux.Handle("/", http.FileServer(http.FS(staticFS)))

	return r.mux
}