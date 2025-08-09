package dtos

import "time"

type DocumentDTO struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	LastUpdated time.Time `json:"lastUpdated"`
	Version     int       `json:"version"`
}

type UpdateContentRequest struct {
	Content string `json:"content"`
}

type UpdateContentResponse struct {
	Success     bool      `json:"success"`
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
	Version     int       `json:"version,omitempty"`
}

type GetContentResponse struct {
	Content     string          `json:"content"`
	LastUpdated time.Time       `json:"lastUpdated"`
	ExpiresAt   time.Time       `json:"expiresAt"`
	Version     int             `json:"version"`
	Attachments []AttachmentDTO `json:"attachments"`
}

type AttachmentDTO struct {
	ID         string    `json:"id"`
	FileName   string    `json:"fileName"`
	MimeType   string    `json:"mimeType"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type UploadFileResponse struct {
	Success    bool          `json:"success"`
	Attachment AttachmentDTO `json:"attachment,omitempty"`
	Error      string        `json:"error,omitempty"`
}

type DeleteFileResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type FileMetadata struct {
	ID         string    `json:"id"`
	FileName   string    `json:"fileName"`
	MimeType   string    `json:"mimeType"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type ListFilesResponse struct {
	Files []AttachmentDTO `json:"files"`
}
