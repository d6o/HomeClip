package valueobjects

import "testing"

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
		{
			name:    "small file",
			size:    1,
			wantErr: false,
		},
		{
			name:    "1MB file",
			size:    1024 * 1024,
			wantErr: false,
		},
		{
			name:    "10MB file",
			size:    10 * 1024 * 1024,
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

func TestFileSize_Value(t *testing.T) {
	size := int64(1024)
	fs, err := NewFileSize(size)
	if err != nil {
		t.Fatalf("Failed to create FileSize: %v", err)
	}

	if fs.Value() != size {
		t.Errorf("Expected Value() to return %d, got %d", size, fs.Value())
	}
}
