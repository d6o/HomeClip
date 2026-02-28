package config

import "os"

const (
	defaultPort    = "8080"
	defaultDataDir = "/data"
)

type Config struct {
	Port    string
	DataDir string
}

func NewConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	return Config{
		Port:    port,
		DataDir: dataDir,
	}
}
