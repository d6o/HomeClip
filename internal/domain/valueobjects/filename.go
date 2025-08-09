package valueobjects

import (
	"errors"
	"strings"
)

const MaxFileNameLength = 255

var ErrInvalidFileName = errors.New("invalid file name")

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
