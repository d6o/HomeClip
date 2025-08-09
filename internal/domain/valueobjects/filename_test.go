package valueobjects

import (
	"strings"
	"testing"
)

func TestNewFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid file name",
			fileName: "test.txt",
			wantErr:  false,
		},
		{
			name:     "empty file name",
			fileName: "",
			wantErr:  true,
		},
		{
			name:     "file name too long",
			fileName: strings.Repeat("a", MaxFileNameLength+1),
			wantErr:  true,
		},
		{
			name:     "file name with path traversal",
			fileName: "../test.txt",
			wantErr:  true,
		},
		{
			name:     "file name with slash",
			fileName: "test/file.txt",
			wantErr:  true,
		},
		{
			name:     "file name with backslash",
			fileName: "test\\file.txt",
			wantErr:  true,
		},
		{
			name:     "file name with spaces",
			fileName: "test file.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName, err := NewFileName(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && fileName.Value() != tt.fileName {
				t.Errorf("Expected file name %v, got %v", tt.fileName, fileName.Value())
			}
		})
	}
}
