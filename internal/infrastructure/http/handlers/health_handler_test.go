package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_Health(t *testing.T) {
	handler := NewHealthHandler()
	
	tests := []struct {
		name   string
		method string
		want   int
	}{
		{
			name:   "GET request success",
			method: http.MethodGet,
			want:   http.StatusOK,
		},
		{
			name:   "POST request not allowed",
			method: http.MethodPost,
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "PUT request not allowed",
			method: http.MethodPut,
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "DELETE request not allowed",
			method: http.MethodDelete,
			want:   http.StatusMethodNotAllowed,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/health", nil)
			rec := httptest.NewRecorder()
			
			handler.Health(rec, req)
			
			if rec.Code != tt.want {
				t.Errorf("Expected status %d, got %d", tt.want, rec.Code)
			}
			
			if tt.method == http.MethodGet {
				// Check response body
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				
				if response["status"] != "healthy" {
					t.Errorf("Expected status 'healthy', got %v", response["status"])
				}
				
				// Check Content-Type header
				if rec.Header().Get("Content-Type") != "application/json" {
					t.Error("Expected Content-Type header to be application/json")
				}
			}
		})
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	handler := NewHealthHandler()
	
	tests := []struct {
		name   string
		method string
		want   int
	}{
		{
			name:   "GET request success",
			method: http.MethodGet,
			want:   http.StatusOK,
		},
		{
			name:   "POST request not allowed",
			method: http.MethodPost,
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "PUT request not allowed",
			method: http.MethodPut,
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "DELETE request not allowed",
			method: http.MethodDelete,
			want:   http.StatusMethodNotAllowed,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/ready", nil)
			rec := httptest.NewRecorder()
			
			handler.Ready(rec, req)
			
			if rec.Code != tt.want {
				t.Errorf("Expected status %d, got %d", tt.want, rec.Code)
			}
			
			if tt.method == http.MethodGet {
				// Check response body
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				
				if response["status"] != "ready" {
					t.Errorf("Expected status 'ready', got %v", response["status"])
				}
				
				// Check Content-Type header
				if rec.Header().Get("Content-Type") != "application/json" {
					t.Error("Expected Content-Type header to be application/json")
				}
			}
		})
	}
}

func TestHealthHandler_ConcurrentRequests(t *testing.T) {
	handler := NewHealthHandler()
	
	// Test concurrent health check requests
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rec := httptest.NewRecorder()
			
			handler.Health(rec, req)
			
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
			}
			
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}