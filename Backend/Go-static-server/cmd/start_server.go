package main

import (
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/server"
)

func main() {
	config.InitConfigs()

	server.Run()
}
