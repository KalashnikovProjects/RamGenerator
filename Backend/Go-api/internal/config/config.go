package config

import (
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

var RootPath = getEnv("ROOT_PATH", "/")
var Conf *Config

type GRPCConfig struct {
	Hostname string
	Token    string
}

type DatabaseConfig struct {
	Hostname string
	Port     int
	User     string
	Password string
	DBName   string
}

type AnotherTokens struct {
	FreeImageHostApiKey string
}

type SecretConfig struct {
	GRPC          GRPCConfig
	Database      DatabaseConfig
	AnotherTokens AnotherTokens
}

type UsersConfig struct {
	DefaultAvatarBox       entities.Box `yaml:"default_avatar_box"`
	MaxUsernameLen         int          `yaml:"max_username_len"`
	TimeBetweenGenerations int          `yaml:"time_between_generations"`
}

type ImageConfig struct {
	DefaultKandinskyStyle string `yaml:"default_kandinsky_style"`
}

type PortsConfig struct {
	Api int `yaml:"go_api"`
}

type ClicksConfig struct {
	FirstRam        int   `yaml:"first_ram"`
	DailyRamsPrices []int `yaml:"daily_rams_prices"`
}

type WebsocketConfig struct {
	PingPeriod int `yaml:"ping_period"`
	PongWait   int `yaml:"pong_wait"`
}

type SettingsConfig struct {
	Users     UsersConfig     `yaml:"users"`
	Image     ImageConfig     `yaml:"image"`
	Ports     PortsConfig     `yaml:"ports"`
	Clicks    ClicksConfig    `yaml:"clicks"`
	Websocket WebsocketConfig `yaml:"websocket"`
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
		log.Fatalf("Not found config.yaml: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &settings)
	if err != nil {
		log.Fatalf("Unmarshal yaml error: %v", err)
	}
	secrets := SecretConfig{
		GRPC: GRPCConfig{
			Hostname: getEnv("GRPC_HOSTNAME", "localhost:50051"),
			Token:    getEnv("GRPC_SECRET_TOKEN", ""),
		},
		Database: DatabaseConfig{
			Hostname: getEnv("POSTGRES_HOSTNAME", "localhost"),
			Port:     toInt(getEnv("POSTGRES_PORT", "5432")),
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", ""),
		},
		AnotherTokens: AnotherTokens{
			FreeImageHostApiKey: getEnv("FREE_IMAGE_HOST_API_KEY", "")},
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
