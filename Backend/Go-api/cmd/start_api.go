package main

import (
	"context"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/api"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/logs"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/ram_generator"
	"log/slog"
	"os"
)

func main() {
	logs.InitStdoutLogs(slog.LevelInfo)
	config.InitConfigs()
	slog.Info("Go api starting...")

	ctx := context.Background()
	pgConnectionString := database.GeneratePostgresConnectionString(config.Conf.Database.User, config.Conf.Database.Password, config.Conf.Database.Host, config.Conf.Database.DBName)
	db := database.CreateDBConnectionContext(ctx, pgConnectionString)
	gRPCConn := ram_generator.CreateGRPCConnection()

	server := api.NewRamGeneratorServer(ctx, fmt.Sprintf(":%d", config.Conf.Ports.Api), db, gRPCConn)
	err := api.ServeServer(ctx, server)
	if err != nil {
		slog.Error("Go api shutdown with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
