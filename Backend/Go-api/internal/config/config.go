package config

import (
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"strconv"
)

var RootPath = getEnv("ROOT_PATH", "/")
var Conf *Config

type GRPCConfig struct {
	Host  string
	Token string
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
}

type SecretConfig struct {
	GRPC     GRPCConfig
	Database DatabaseConfig
	Image    ImageConfig
}

type UsersConfig struct {
	DefaultAvatarBox entities.Box `yaml:"default_avatar_box"`
	MaxUsernameLen   int          `yaml:"max_username_len"`
}

type GenerationConfig struct {
	TimeBetweenDaily        int   `yaml:"time_between_daily"`
	TimeBetweenDailyAnother []int `yaml:"time_between_daily_another"`
	MaxPromptLen            int   `yaml:"max_prompt_len"`
}

type ImageConfig struct {
	ImageCDNOpenAPI     string
	ImageCDNInternalAPI string
	ImageCDNApiKey      string
}

type PortsConfig struct {
	Api int `yaml:"go_api"`
}

type ClicksConfig struct {
	CPSLimit  int   `yaml:"cps_limit"`
	FirstRam  int   `yaml:"first_ram"`
	DailyRams []int `yaml:"daily_rams"`
}

type WebsocketConfig struct {
	PingPeriod int `yaml:"ping_period"`
	PongWait   int `yaml:"pong_wait"`
}

type AnotherConfig struct {
	DefaultKandinskyStyle string `yaml:"default_kandinsky_style"`
	TopRamsCount          int    `yaml:"top_rams_count"`
}

type SettingsConfig struct {
	Ports      PortsConfig      `yaml:"ports"`
	Clicks     ClicksConfig     `yaml:"clicks"`
	Users      UsersConfig      `yaml:"users"`
	Generation GenerationConfig `yaml:"generation"`
	Websocket  WebsocketConfig  `yaml:"websocket"`
	Another    AnotherConfig    `yaml:"another"`
}

type Config struct {
	SecretConfig
	SettingsConfig
}

func InitConfigs() {
	RootPath = getEnv("ROOT_PATH", "../..")
	err := godotenv.Load(fmt.Sprintf("%s/.env", RootPath))
	var settings *SettingsConfig
	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/config.yaml", RootPath))

	if err != nil {
		slog.Error("Not found config.yaml", slog.Any("error", err))
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, &settings)
	if err != nil {
		slog.Error("Unmarshal yaml error", slog.Any("error", err))
		os.Exit(1)
	}
	secrets := SecretConfig{
		GRPC: GRPCConfig{
			Host:  getEnv("GRPC_HOST", "localhost:50051"),
			Token: getEnv("GRPC_SECRET_TOKEN", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost:5432"),
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", ""),
		},
		Image: ImageConfig{
			ImageCDNApiKey:      getEnv("IMAGE_CDN_API_KEY", ""),
			ImageCDNOpenAPI:     getEnv("IMAGE_CDN_OPEN_API", "http://localhost/cdn"),
			ImageCDNInternalAPI: getEnv("IMAGE_CDN_INTERNAL_API", "http://localhost:8084"),
		},
	}
	Conf = &Config{SecretConfig: secrets, SettingsConfig: *settings}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func toInt(str string) int {
	res, _ := strconv.Atoi(str)
	return res
}
