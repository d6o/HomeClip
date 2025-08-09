package valueobjects

import (
	"strings"
	"testing"
)

func TestNewContent(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid content",
			value:   "This is valid content",
			wantErr: false,
		},
		{
			name:    "empty content",
			value:   "",
			wantErr: false,
		},
		{
			name:    "content at max size",
			value:   strings.Repeat("a", MaxContentLength),
			wantErr: false,
		},
		{
			name:    "content exceeds max size",
			value:   strings.Repeat("a", MaxContentLength+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewContent(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && content.Value() != tt.value {
				t.Errorf("Expected content value %v, got %v", tt.value, content.Value())
			}
		})
	}
}

func TestContent_IsEmpty(t *testing.T) {
	emptyContent, _ := NewContent("")
	if !emptyContent.IsEmpty() {
		t.Error("Expected empty content to return true for IsEmpty()")
	}

	nonEmptyContent, _ := NewContent("some text")
	if nonEmptyContent.IsEmpty() {
		t.Error("Expected non-empty content to return false for IsEmpty()")
	}
}

func TestContent_Length(t *testing.T) {
	content, _ := NewContent("hello")
	if content.Length() != 5 {
		t.Errorf("Expected length 5, got %d", content.Length())
	}

	emptyContent, _ := NewContent("")
	if emptyContent.Length() != 0 {
		t.Errorf("Expected length 0, got %d", emptyContent.Length())
	}
}

func TestContent_Equals(t *testing.T) {
	content1, _ := NewContent("test")
	content2, _ := NewContent("test")
	content3, _ := NewContent("different")

	if !content1.Equals(content2) {
		t.Error("Expected equal contents to return true")
	}

	if content1.Equals(content3) {
		t.Error("Expected different contents to return false")
	}
}

func TestEmptyContent(t *testing.T) {
	content := EmptyContent()

	if !content.IsEmpty() {
		t.Error("Expected EmptyContent to return empty content")
	}

	if content.Value() != "" {
		t.Error("Expected EmptyContent value to be empty string")
	}
}
