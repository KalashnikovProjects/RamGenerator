package config

import (
	"os"
)

type GRPCConfig struct {
	URL   string
	Token string
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
}

type Config struct {
	GRPC     GRPCConfig
	Database DatabaseConfig
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		GRPC: GRPCConfig{
			URL:   getEnv("GRPC_URL", "localhost:50051"),
			Token: getEnv("GRPC_SECRET_TOKEN", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", ""),
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", ""),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
