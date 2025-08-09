package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/application/services"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

type DocumentHandler struct {
	appService *services.DocumentApplicationService
}

func NewDocumentHandler(appService *services.DocumentApplicationService) *DocumentHandler {
	return &DocumentHandler{
		appService: appService,
	}
}

func (h *DocumentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w)
		return
	}

	ctx := r.Context()
	response, err := h.appService.GetContent(ctx, "")
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *DocumentHandler) SaveContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.methodNotAllowed(w)
		return
	}

	var req dtos.UpdateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, err, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, err := h.appService.UpdateContent(ctx, "", req.Content)
	if err != nil {
		if errors.Is(err, valueobjects.ErrContentTooLarge) ||
			errors.Is(err, valueobjects.ErrInvalidContent) ||
			strings.Contains(err.Error(), "too large") ||
			strings.Contains(err.Error(), "exceeds maximum") {
			h.handleError(w, err, http.StatusBadRequest)
		} else {
			h.handleError(w, err, http.StatusInternalServerError)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *DocumentHandler) HandleContent(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetContent(w, r)
	case http.MethodPost:
		h.SaveContent(w, r)
	default:
		h.methodNotAllowed(w)
	}
}

func (h *DocumentHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *DocumentHandler) handleError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

func (h *DocumentHandler) methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
