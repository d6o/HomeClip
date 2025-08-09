package valueobjects

import "testing"

func TestNewMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     string
		wantErr  bool
	}{
		{
			name:     "valid mime type",
			mimeType: "text/plain",
			want:     "text/plain",
			wantErr:  false,
		},
		{
			name:     "empty mime type defaults to octet-stream",
			mimeType: "",
			want:     "application/octet-stream",
			wantErr:  false,
		},
		{
			name:     "mime type with charset",
			mimeType: "text/html; charset=utf-8",
			want:     "text/html",
			wantErr:  false,
		},
		{
			name:     "uppercase mime type",
			mimeType: "TEXT/PLAIN",
			want:     "text/plain",
			wantErr:  false,
		},
		{
			name:     "video mime type",
			mimeType: "video/mp4",
			want:     "video/mp4",
			wantErr:  false,
		},
		{
			name:     "audio mime type",
			mimeType: "audio/mpeg",
			want:     "audio/mpeg",
			wantErr:  false,
		},
		{
			name:     "any custom mime type is accepted",
			mimeType: "application/custom-type",
			want:     "application/custom-type",
			wantErr:  false,
		},
		{
			name:     "mime type with multiple parameters",
			mimeType: "text/html; charset=utf-8; boundary=something",
			want:     "text/html",
			wantErr:  false,
		},
		{
			name:     "spaces around mime type",
			mimeType: "  text/plain  ",
			want:     "text/plain",
			wantErr:  false,
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
