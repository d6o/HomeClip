package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	clearEnvVars()

	cfg := LoadConfig()

	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}

	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("Expected default read timeout 15s, got %v", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("Expected default write timeout 15s, got %v", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("Expected default idle timeout 60s, got %v", cfg.IdleTimeout)
	}

	if cfg.MaxContentSize != 1048576 {
		t.Errorf("Expected default max content size 1048576, got %d", cfg.MaxContentSize)
	}

	if cfg.MaxFileSize != 10485760 {
		t.Errorf("Expected default max file size 10485760, got %d", cfg.MaxFileSize)
	}

	if cfg.MaxFileNameLength != 255 {
		t.Errorf("Expected default max filename length 255, got %d", cfg.MaxFileNameLength)
	}

	if cfg.ExpirationDuration != 24*time.Hour {
		t.Errorf("Expected default expiration duration 24h, got %v", cfg.ExpirationDuration)
	}

	if cfg.CleanupInterval != 5*time.Minute {
		t.Errorf("Expected default cleanup interval 5m, got %v", cfg.CleanupInterval)
	}

	if cfg.ExpirationGracePeriod != 1*time.Hour {
		t.Errorf("Expected default expiration grace period 1h, got %v", cfg.ExpirationGracePeriod)
	}

	if !cfg.EnableFileUploads {
		t.Error("Expected file uploads to be enabled by default")
	}

	if !cfg.EnableAutoCleanup {
		t.Error("Expected auto cleanup to be enabled by default")
	}
}

func TestLoadConfig_FromEnvironment(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("READ_TIMEOUT", "30s")
	os.Setenv("WRITE_TIMEOUT", "45s")
	os.Setenv("IDLE_TIMEOUT", "120s")
	os.Setenv("MAX_CONTENT_SIZE", "2097152")
	os.Setenv("MAX_FILE_SIZE", "20971520")
	os.Setenv("MAX_FILENAME_LENGTH", "512")
	os.Setenv("EXPIRATION_DURATION", "48h")
	os.Setenv("CLEANUP_INTERVAL", "10m")
	os.Setenv("EXPIRATION_GRACE_PERIOD", "2h")
	os.Setenv("ENABLE_FILE_UPLOADS", "false")
	os.Setenv("ENABLE_AUTO_CLEANUP", "false")

	defer clearEnvVars()

	cfg := LoadConfig()

	if cfg.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Port)
	}

	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 45*time.Second {
		t.Errorf("Expected write timeout 45s, got %v", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout != 120*time.Second {
		t.Errorf("Expected idle timeout 120s, got %v", cfg.IdleTimeout)
	}

	if cfg.MaxContentSize != 2097152 {
		t.Errorf("Expected max content size 2097152, got %d", cfg.MaxContentSize)
	}

	if cfg.MaxFileSize != 20971520 {
		t.Errorf("Expected max file size 20971520, got %d", cfg.MaxFileSize)
	}

	if cfg.MaxFileNameLength != 512 {
		t.Errorf("Expected max filename length 512, got %d", cfg.MaxFileNameLength)
	}

	if cfg.ExpirationDuration != 48*time.Hour {
		t.Errorf("Expected expiration duration 48h, got %v", cfg.ExpirationDuration)
	}

	if cfg.CleanupInterval != 10*time.Minute {
		t.Errorf("Expected cleanup interval 10m, got %v", cfg.CleanupInterval)
	}

	if cfg.ExpirationGracePeriod != 2*time.Hour {
		t.Errorf("Expected expiration grace period 2h, got %v", cfg.ExpirationGracePeriod)
	}

	if cfg.EnableFileUploads {
		t.Error("Expected file uploads to be disabled")
	}

	if cfg.EnableAutoCleanup {
		t.Error("Expected auto cleanup to be disabled")
	}
}

func TestConfig_ServerPort(t *testing.T) {
	cfg := &Config{
		Port: "3000",
	}

	if cfg.ServerPort() != "3000" {
		t.Errorf("Expected ServerPort to return 3000, got %s", cfg.ServerPort())
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_ENV_VAR",
			envValue:     "custom_value",
			defaultValue: "default",
			expected:     "custom_value",
		},
		{
			name:         "returns default when env not set",
			key:          "UNSET_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "returns parsed int when valid",
			key:          "TEST_INT",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "returns default when invalid int",
			key:          "TEST_INT",
			envValue:     "not_a_number",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "returns default when not set",
			key:          "UNSET_INT",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getIntEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetInt64Env(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int64
		expected     int64
	}{
		{
			name:         "returns parsed int64 when valid",
			key:          "TEST_INT64",
			envValue:     "9223372036854775807",
			defaultValue: 100,
			expected:     9223372036854775807,
		},
		{
			name:         "returns default when invalid int64",
			key:          "TEST_INT64",
			envValue:     "invalid",
			defaultValue: 100,
			expected:     100,
		},
		{
			name:         "returns default when not set",
			key:          "UNSET_INT64",
			envValue:     "",
			defaultValue: 100,
			expected:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getInt64Env(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "returns true when 'true'",
			key:          "TEST_BOOL",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "returns false when 'false'",
			key:          "TEST_BOOL",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "returns true when '1'",
			key:          "TEST_BOOL",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "returns false when '0'",
			key:          "TEST_BOOL",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "returns default when invalid",
			key:          "TEST_BOOL",
			envValue:     "invalid",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "returns default when not set",
			key:          "UNSET_BOOL",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getBoolEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "returns parsed duration when valid",
			key:          "TEST_DURATION",
			envValue:     "10s",
			defaultValue: 5 * time.Second,
			expected:     10 * time.Second,
		},
		{
			name:         "handles complex duration",
			key:          "TEST_DURATION",
			envValue:     "1h30m45s",
			defaultValue: 5 * time.Second,
			expected:     1*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			name:         "returns default when invalid",
			key:          "TEST_DURATION",
			envValue:     "invalid",
			defaultValue: 5 * time.Second,
			expected:     5 * time.Second,
		},
		{
			name:         "returns default when not set",
			key:          "UNSET_DURATION",
			envValue:     "",
			defaultValue: 5 * time.Second,
			expected:     5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getDurationEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func clearEnvVars() {
	envVars := []string{
		"PORT",
		"READ_TIMEOUT",
		"WRITE_TIMEOUT",
		"IDLE_TIMEOUT",
		"MAX_CONTENT_SIZE",
		"MAX_FILE_SIZE",
		"MAX_FILENAME_LENGTH",
		"EXPIRATION_DURATION",
		"CLEANUP_INTERVAL",
		"EXPIRATION_GRACE_PERIOD",
		"ENABLE_FILE_UPLOADS",
		"ENABLE_AUTO_CLEANUP",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
