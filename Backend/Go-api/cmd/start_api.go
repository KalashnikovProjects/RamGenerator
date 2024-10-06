package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/api"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/logs"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/ram_generator"
	pb "github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/proto_generated"
	"log"
	"log/slog"
	"os"
)

// TestContext TODO: ПЕРЕНЕСТИ В ТЕСТЫ, сделать автоматическими
func TestContext(ctx context.Context, db *sql.DB, gRPCConn pb.RamGeneratorClient) {
	id, err := database.CreateUserContext(ctx, db, entities.User{Username: "alo", PasswordHash: "adwdawd", DailyRamGenerationTime: 11111})
	if err != nil {
		log.Fatal("Ошибка при CreateUserContext: ", err)
	}
	log.Print("Создал пользователя: ", id)

	err = database.UpdateUserContext(ctx, db, id, entities.User{Username: "pukek ultra 2", DailyRamGenerationTime: 2})
	if err != nil {
		log.Fatal("Ошибка при UpdateUserContext: ", err)
	}

	ramId, err := database.CreateRamContext(ctx, db, entities.Ram{Description: "крутой баран", ImageUrl: "pukek/alo", UserId: id})
	if err != nil {
		log.Fatal("Ошибка при CreateRamContext: ", err)
	}
	log.Print("Создал барана: ", ramId)

	err = database.UpdateRamContext(ctx, db, ramId, entities.Ram{Description: "гига баран"})
	if err != nil {
		log.Fatal("Ошибка при UpdateUserContext: ", err)
	}

	u, err := database.GetUserContext(ctx, db, id)
	if err != nil {
		log.Fatal("Ошибка при GetUserContext: ", err)
	}
	fmt.Printf("%+v\n", u)

	r, err := database.GetRamContext(ctx, db, ramId)
	if err != nil {
		log.Fatal("Ошибка при GetRamContext: ", err)
	}
	fmt.Printf("%+v\n", r)

	rs, err := database.GetRamsByUsernameContext(ctx, db, u.Username)
	if err != nil {
		log.Fatal("Ошибка при GetRamsByUsernameContext: ", err)
	}
	for i, item := range rs {
		fmt.Printf("Баран %d: %+v\n", i, item)
	}

	err = database.DeleteUserContext(ctx, db, id)
	if err != nil {
		log.Fatal("Ошибка при DeleteUserContext: ", err)
	}
}

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
