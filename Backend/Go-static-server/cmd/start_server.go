package main

import (
	"context"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/logs"
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/server"
	"log/slog"
	"os"
)

func main() {
	logs.InitStdoutLogs(slog.LevelInfo)
	config.InitConfigs()
	ctx := context.Background()
	serv := server.NewStaticServer(ctx, fmt.Sprintf(":%d", config.Conf.Port))
	err := server.ServeServer(ctx, serv)
	if err != nil {
		slog.Error("Go static server shutdown with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
