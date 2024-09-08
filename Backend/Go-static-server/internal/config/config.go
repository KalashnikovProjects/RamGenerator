package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"strconv"
)

var Conf *Config

type Config struct {
	ApiUrl        string
	Port          int
	DefaultAvatar string

	RootPath      string
	TemplatesPath string
	CdnFilesPath  string
	FaviconPath   string
}

type yamlConfigData struct {
	Frontend struct {
		ApiUrl string `yaml:"api_url"`
	} `yaml:"frontend"`
	Ports struct {
		GoStaticServer int `yaml:"go_static_server"`
	} `yaml:"ports"`
	Users struct {
		DefaultAvatar string `yaml:"default_avatar"`
	} `yaml:"users"`
}

func InitConfigs() {
	rootPath := getEnv("ROOT_PATH", "../..")

	err := godotenv.Load(fmt.Sprintf("%s/.env", rootPath))
	var configData *yamlConfigData
	yamlFile, err := os.ReadFile(fmt.Sprintf("%s/config.yaml", rootPath))

	if err != nil {
		slog.Error("Yaml file not found error", slog.Any("error", err))
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, &configData)
	if err != nil {
		slog.Error("Unmarshal yaml error", slog.Any("error", err))
		os.Exit(1)
	}
	Conf = &Config{
		ApiUrl:        configData.Frontend.ApiUrl,
		Port:          configData.Ports.GoStaticServer,
		DefaultAvatar: configData.Users.DefaultAvatar,
		RootPath:      rootPath,
		TemplatesPath: fmt.Sprintf("%s/Frontend/templates", rootPath),
		CdnFilesPath:  fmt.Sprintf("%s/Frontend/static", rootPath),
		FaviconPath:   fmt.Sprintf("%s/Frontend/static/img/icon196.ico", rootPath),
	}
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
