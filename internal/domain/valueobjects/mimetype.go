package valueobjects

import (
	"strings"
)

type MimeType struct {
	value string
}

func NewMimeType(mimeType string) (MimeType, error) {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = mimeType[:idx]
	}

	return MimeType{value: mimeType}, nil
}

func (m MimeType) Value() string {
	return m.value
}
