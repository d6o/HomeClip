package valueobjects

import "errors"

const MaxFileSize = 50 * 1024 * 1024 // 50MB to support larger files

var (
	ErrFileTooLarge = errors.New("file exceeds maximum allowed size")
	ErrEmptyFile    = errors.New("file cannot be empty")
)

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
