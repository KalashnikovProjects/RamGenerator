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

	slog.Info("Rendering templates")
	err := server.RenderTemplates()
	if err != nil {
		panic(err)
	}
	serv := server.NewStaticServer(fmt.Sprintf(":%d", config.Conf.Port))

	slog.Info("Running static server", slog.String("addr", serv.Addr))

	err = server.ServeServer(ctx, serv)
	if err != nil {
		slog.Error("Go static server shutdown with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
