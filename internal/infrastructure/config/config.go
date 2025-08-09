package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	MaxContentSize    int64
	MaxFileSize       int64
	MaxFileNameLength int

	ExpirationDuration    time.Duration
	CleanupInterval       time.Duration
	ExpirationGracePeriod time.Duration

	EnableFileUploads bool
	EnableAutoCleanup bool
}

func LoadConfig() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),

		MaxContentSize:    getInt64Env("MAX_CONTENT_SIZE", 1048576),
		MaxFileSize:       getInt64Env("MAX_FILE_SIZE", 10485760),
		MaxFileNameLength: getIntEnv("MAX_FILENAME_LENGTH", 255),

		ExpirationDuration:    getDurationEnv("EXPIRATION_DURATION", 24*time.Hour),
		CleanupInterval:       getDurationEnv("CLEANUP_INTERVAL", 5*time.Minute),
		ExpirationGracePeriod: getDurationEnv("EXPIRATION_GRACE_PERIOD", 1*time.Hour),

		EnableFileUploads: getBoolEnv("ENABLE_FILE_UPLOADS", true),
		EnableAutoCleanup: getBoolEnv("ENABLE_AUTO_CLEANUP", true),
	}
}

func (c *Config) ServerPort() string {
	return c.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
