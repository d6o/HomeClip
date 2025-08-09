package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/d6o/homeclip/internal/application"
	"github.com/d6o/homeclip/internal/application/commands"
	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/application/queries"
)

type FileHandler struct {
	uploadHandler    application.UploadFileCommandHandler
	deleteHandler    application.DeleteFileCommandHandler
	getFileHandler   application.GetFileQueryHandler
	listFilesHandler application.ListFilesQueryHandler
}

func NewFileHandler(
	uploadHandler application.UploadFileCommandHandler,
	deleteHandler application.DeleteFileCommandHandler,
	getFileHandler application.GetFileQueryHandler,
	listFilesHandler application.ListFilesQueryHandler,
) *FileHandler {
	return &FileHandler{
		uploadHandler:    uploadHandler,
		deleteHandler:    deleteHandler,
		getFileHandler:   getFileHandler,
		listFilesHandler: listFilesHandler,
	}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		h.writeErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeErrorResponse(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = getMimeTypeFromFileName(header.Filename)
	}

	cmd := commands.UploadFileCommand{
		DocumentID: "",
		FileName:   header.Filename,
		MimeType:   mimeType,
		Size:       header.Size,
		Reader:     file,
	}

	attachment, err := h.uploadHandler.Handle(r.Context(), cmd)
	if err != nil {
		h.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := dtos.UploadFileResponse{
		Success: true,
		Attachment: dtos.AttachmentDTO{
			ID:         string(attachment.ID()),
			FileName:   attachment.FileName().Value(),
			MimeType:   attachment.MimeType().Value(),
			Size:       attachment.Size().Value(),
			UploadedAt: attachment.UploadedAt().Value(),
			ExpiresAt:  attachment.ExpiresAt().Value(),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	attachmentID := extractAttachmentID(r.URL.Path)
	if attachmentID == "" {
		http.Error(w, "Attachment ID required", http.StatusBadRequest)
		return
	}

	query := queries.GetFileQuery{
		DocumentID:   "",
		AttachmentID: attachmentID,
	}

	result, err := h.getFileHandler.Handle(r.Context(), query)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer result.Reader.Close()

	w.Header().Set("Content-Type", result.Attachment.MimeType().Value())
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Attachment.FileName().Value()+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(result.Attachment.Size().Value(), 10))

	io.Copy(w, result.Reader)
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	attachmentID := extractAttachmentID(r.URL.Path)
	if attachmentID == "" {
		http.Error(w, "Attachment ID required", http.StatusBadRequest)
		return
	}

	cmd := commands.DeleteFileCommand{
		DocumentID:   "",
		AttachmentID: attachmentID,
	}

	err := h.deleteHandler.Handle(r.Context(), cmd)
	if err != nil {
		h.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := dtos.DeleteFileResponse{
		Success: true,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := queries.ListFilesQuery{
		DocumentID: "",
	}

	attachments, err := h.listFilesHandler.Handle(r.Context(), query)
	if err != nil {
		h.writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dtoAttachments := make([]dtos.AttachmentDTO, 0, len(attachments))
	for _, att := range attachments {
		if !att.IsExpired() {
			dtoAttachments = append(dtoAttachments, dtos.AttachmentDTO{
				ID:         string(att.ID()),
				FileName:   att.FileName().Value(),
				MimeType:   att.MimeType().Value(),
				Size:       att.Size().Value(),
				UploadedAt: att.UploadedAt().Value(),
				ExpiresAt:  att.ExpiresAt().Value(),
			})
		}
	}

	h.writeJSON(w, http.StatusOK, dtoAttachments)
}

func (h *FileHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *FileHandler) writeErrorResponse(w http.ResponseWriter, message string, status int) {
	response := dtos.UploadFileResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func extractAttachmentID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func getMimeTypeFromFileName(filename string) string {
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return "application/octet-stream"
	}
	ext := strings.ToLower(strings.TrimSpace(filename[lastDot+1:]))
	mimeTypes := map[string]string{
		"txt":  "text/plain",
		"html": "text/html",
		"css":  "text/css",
		"js":   "text/javascript",
		"json": "application/json",
		"pdf":  "application/pdf",
		"zip":  "application/zip",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"svg":  "image/svg+xml",
		"doc":  "application/msword",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"xls":  "application/vnd.ms-excel",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}
