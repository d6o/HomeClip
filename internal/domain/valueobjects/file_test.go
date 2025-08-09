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

func TestFileName_Extension(t *testing.T) {
	fileName, _ := NewFileName("test.txt")
	if fileName.Extension() != ".txt" {
		t.Errorf("Expected extension .txt, got %v", fileName.Extension())
	}

	fileName2, _ := NewFileName("test.PDF")
	if fileName2.Extension() != ".pdf" {
		t.Errorf("Expected extension .pdf, got %v", fileName2.Extension())
	}
}

func TestFileName_NameWithoutExtension(t *testing.T) {
	fileName, _ := NewFileName("test.txt")
	if fileName.NameWithoutExtension() != "test" {
		t.Errorf("Expected name without extension 'test', got %v", fileName.NameWithoutExtension())
	}
}

func TestNewFileSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{
			name:    "valid size",
			size:    100,
			wantErr: false,
		},
		{
			name:    "zero size",
			size:    0,
			wantErr: true,
		},
		{
			name:    "negative size",
			size:    -1,
			wantErr: true,
		},
		{
			name:    "size exceeds max",
			size:    MaxFileSize + 1,
			wantErr: true,
		},
		{
			name:    "size at max",
			size:    MaxFileSize,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSize, err := NewFileSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && fileSize.Value() != tt.size {
				t.Errorf("Expected size %v, got %v", tt.size, fileSize.Value())
			}
		})
	}
}

func TestFileSize_Conversions(t *testing.T) {
	fileSize, _ := NewFileSize(1024 * 1024) // 1MB
	
	if fileSize.Bytes() != 1024*1024 {
		t.Errorf("Expected 1048576 bytes, got %v", fileSize.Bytes())
	}
	
	if fileSize.Kilobytes() != 1024 {
		t.Errorf("Expected 1024 KB, got %v", fileSize.Kilobytes())
	}
	
	if fileSize.Megabytes() != 1 {
		t.Errorf("Expected 1 MB, got %v", fileSize.Megabytes())
	}
}

func TestNewMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     string
		wantErr  bool
	}{
		{
			name:     "valid text/plain",
			mimeType: "text/plain",
			want:     "text/plain",
			wantErr:  false,
		},
		{
			name:     "valid with charset",
			mimeType: "text/plain; charset=utf-8",
			want:     "text/plain",
			wantErr:  false,
		},
		{
			name:     "valid uppercase",
			mimeType: "TEXT/PLAIN",
			want:     "text/plain",
			wantErr:  false,
		},
		{
			name:     "empty mime type",
			mimeType: "",
			wantErr:  true,
		},
		{
			name:     "not allowed mime type",
			mimeType: "application/x-executable",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, err := NewMimeType(tt.mimeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMimeType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mimeType.Value() != tt.want {
				t.Errorf("Expected mime type %v, got %v", tt.want, mimeType.Value())
			}
		})
	}
}

func TestMimeType_IsImage(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     bool
	}{
		{"image/png", "image/png", true},
		{"image/jpeg", "image/jpeg", true},
		{"text/plain", "text/plain", false},
		{"application/pdf", "application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, _ := NewMimeType(tt.mimeType)
			if got := mimeType.IsImage(); got != tt.want {
				t.Errorf("IsImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMimeType_IsText(t *testing.T) {
	mimeType, _ := NewMimeType("text/plain")
	if !mimeType.IsText() {
		t.Error("Expected text/plain to return true for IsText()")
	}
	
	pdfType, _ := NewMimeType("application/pdf")
	if pdfType.IsText() {
		t.Error("Expected application/pdf to return false for IsText()")
	}
}

func TestMimeType_IsPDF(t *testing.T) {
	pdfType, _ := NewMimeType("application/pdf")
	if !pdfType.IsPDF() {
		t.Error("Expected application/pdf to return true for IsPDF()")
	}
	
	textType, _ := NewMimeType("text/plain")
	if textType.IsPDF() {
		t.Error("Expected text/plain to return false for IsPDF()")
	}
}