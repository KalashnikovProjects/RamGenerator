package config

import (
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	"os"
	"strconv"
)

type GRPCConfig struct {
	Hostname string
	Port     int
	Token    string
}

type DatabaseConfig struct {
	Hostname string
	Port     int
	User     string
	Password string
	DBName   string
}

type UsersConfig struct {
	DefaultAvatar          string
	DefaultAvatarBox       *entities.Box
	MaxUsernameLen         int // Это поле лучше лишний раз не трогать, возможно придётся мигрировать бд
	TimeBetweenGenerations int // Время в часах
}

type ImageConfig struct {
	DefaultKandinskyStyle string
	FreeImageHostApiKey   string
}

type Config struct {
	GRPC        GRPCConfig
	Database    DatabaseConfig
	UsersConfig UsersConfig
	Image       ImageConfig
}

func toInt(str string) int {
	res, _ := strconv.Atoi(str)
	return res
}

// New возвращает заполненный Config
func New() *Config {
	return &Config{
		GRPC: GRPCConfig{
			Hostname: getEnv("GRPC_HOSTNAME", "localhost"),
			Port:     toInt(getEnv("GRPC_PORT", "50051")),
			Token:    getEnv("GRPC_SECRET_TOKEN", ""),
		},
		Database: DatabaseConfig{
			Hostname: getEnv("POSTGRES_HOSTNAME", "localhost"),
			Port:     toInt(getEnv("POSTGRES_PORT", "5432")),
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", ""),
		},
		UsersConfig: UsersConfig{
			DefaultAvatar:          "https://www.funnyart.club/uploads/posts/2023-02/1675548431_www-funnyart-club-p-smeshnoi-barashek-shutki-13.jpg",
			DefaultAvatarBox:       &entities.Box{{20, 230}, {530, 760}},
			MaxUsernameLen:         24, // Это поле лучше лишний раз не трогать, возможно придётся мигрировать бд
			TimeBetweenGenerations: 20,
		},
		Image: ImageConfig{
			DefaultKandinskyStyle: "DEFAULT",
			FreeImageHostApiKey:   getEnv("FREE_IMAGE_HOST_API_KEY", ""),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
