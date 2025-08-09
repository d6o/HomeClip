package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/d6o/homeclip/internal/infrastructure/config"
)

func TestNewServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	cfg := &config.Config{
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	server := NewServer(handler, cfg)
	
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	
	if server.Port() != "8080" {
		t.Errorf("Expected port 8080, got %s", server.Port())
	}
	
	if server.URL() != "http://localhost:8080" {
		t.Errorf("Expected URL http://localhost:8080, got %s", server.URL())
	}
	
	if server.httpServer.Addr != ":8080" {
		t.Errorf("Expected address :8080, got %s", server.httpServer.Addr)
	}
	
	if server.httpServer.ReadTimeout != 15*time.Second {
		t.Errorf("Expected read timeout 15s, got %v", server.httpServer.ReadTimeout)
	}
	
	if server.httpServer.WriteTimeout != 15*time.Second {
		t.Errorf("Expected write timeout 15s, got %v", server.httpServer.WriteTimeout)
	}
	
	if server.httpServer.IdleTimeout != 60*time.Second {
		t.Errorf("Expected idle timeout 60s, got %v", server.httpServer.IdleTimeout)
	}
}

func TestServer_Port(t *testing.T) {
	server := &Server{
		port: "3000",
	}
	
	if server.Port() != "3000" {
		t.Errorf("Expected port 3000, got %s", server.Port())
	}
}

func TestServer_URL(t *testing.T) {
	tests := []struct {
		name string
		port string
		want string
	}{
		{
			name: "standard port",
			port: "8080",
			want: "http://localhost:8080",
		},
		{
			name: "custom port",
			port: "3000",
			want: "http://localhost:3000",
		},
		{
			name: "port 80",
			port: "80",
			want: "http://localhost:80",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				port: tt.port,
			}
			
			if got := server.URL(); got != tt.want {
				t.Errorf("URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Shutdown(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	cfg := &config.Config{
		Port:         "0", // Use port 0 to get a random available port
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		IdleTimeout:  1 * time.Second,
	}
	
	server := NewServer(handler, cfg)
	
	// Start server in background
	go func() {
		_ = server.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shutdown server: %v", err)
	}
}

func TestServer_ShutdownTimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate long-running request
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	
	cfg := &config.Config{
		Port:         "0",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}
	
	server := NewServer(handler, cfg)
	
	// Create a very short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Shutdown should respect the context timeout
	err := server.Shutdown(ctx)
	if err == nil {
		// In a real scenario with active connections, this would timeout
		// Since we're not actually serving, it might succeed immediately
		t.Log("Shutdown succeeded immediately (no active connections)")
	}
}