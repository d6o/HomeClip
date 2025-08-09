package valueobjects

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	MaxFileNameLength = 255
)

var (
	ErrFileTooLarge    = errors.New("file exceeds maximum allowed size")
	ErrInvalidFileName = errors.New("invalid file name")
	ErrInvalidMimeType = errors.New("invalid mime type")
	ErrEmptyFile       = errors.New("file cannot be empty")
)

var AllowedMimeTypes = map[string]bool{
	"text/plain":               true,
	"text/html":                true,
	"text/css":                 true,
	"text/javascript":          true,
	"application/json":         true,
	"application/pdf":          true,
	"application/zip":          true,
	"application/x-zip-compressed": true,
	"image/jpeg":               true,
	"image/jpg":                true,
	"image/png":                true,
	"image/gif":                true,
	"image/svg+xml":            true,
	"application/msword":       true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

type FileName struct {
	value string
}

func NewFileName(name string) (FileName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return FileName{}, ErrInvalidFileName
	}
	if len(name) > MaxFileNameLength {
		return FileName{}, ErrInvalidFileName
	}
	
	// Basic security check for path traversal
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return FileName{}, ErrInvalidFileName
	}
	
	return FileName{value: name}, nil
}

func (f FileName) Value() string {
	return f.value
}

func (f FileName) Extension() string {
	return strings.ToLower(filepath.Ext(f.value))
}

func (f FileName) NameWithoutExtension() string {
	ext := filepath.Ext(f.value)
	return strings.TrimSuffix(f.value, ext)
}

type FileSize struct {
	value int64
}

func NewFileSize(size int64) (FileSize, error) {
	if size <= 0 {
		return FileSize{}, ErrEmptyFile
	}
	if size > MaxFileSize {
		return FileSize{}, ErrFileTooLarge
	}
	return FileSize{value: size}, nil
}

func (f FileSize) Value() int64 {
	return f.value
}

func (f FileSize) Bytes() int64 {
	return f.value
}

func (f FileSize) Kilobytes() float64 {
	return float64(f.value) / 1024
}

func (f FileSize) Megabytes() float64 {
	return float64(f.value) / (1024 * 1024)
}

type MimeType struct {
	value string
}

func NewMimeType(mimeType string) (MimeType, error) {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if mimeType == "" {
		return MimeType{}, ErrInvalidMimeType
	}
	
	// Extract base mime type (remove charset and other parameters)
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = mimeType[:idx]
	}
	
	if !AllowedMimeTypes[mimeType] {
		return MimeType{}, ErrInvalidMimeType
	}
	
	return MimeType{value: mimeType}, nil
}

func (m MimeType) Value() string {
	return m.value
}

func (m MimeType) IsImage() bool {
	return strings.HasPrefix(m.value, "image/")
}

func (m MimeType) IsText() bool {
	return strings.HasPrefix(m.value, "text/")
}

func (m MimeType) IsPDF() bool {
	return m.value == "application/pdf"
}