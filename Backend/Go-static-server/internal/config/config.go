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

type BaseTemplateData struct {
	ApiUrl            string
	WebsocketProtocol string
	DefaultAvatar     string
}

type Paths struct {
	Root      string
	Templates string
	CdnFiles  string
	Favicon   string
}

type Config struct {
	Port int
	BaseTemplateData
	Paths
}

type yamlConfigData struct {
	Frontend struct {
		ApiUrl            string `yaml:"api_url"`
		WebsocketProtocol string `yaml:"websocket_protocol"`
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
		Port: configData.Ports.GoStaticServer,
		BaseTemplateData: BaseTemplateData{
			ApiUrl:            configData.Frontend.ApiUrl,
			WebsocketProtocol: configData.Frontend.WebsocketProtocol,
			DefaultAvatar:     configData.Users.DefaultAvatar,
		},
		Paths: Paths{
			Root:      rootPath,
			Templates: fmt.Sprintf("%s/Frontend/templates", rootPath),
			CdnFiles:  fmt.Sprintf("%s/Frontend/static", rootPath),
			Favicon:   fmt.Sprintf("%s/Frontend/static/img/icon196.ico", rootPath),
		},
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
