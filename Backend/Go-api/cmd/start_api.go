package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/api"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/ram_image_generator"
	pb "github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/proto_generated"
	"log"
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
	config.InitConfigs()
	log.Println("Запуск go сервиса...")

	ctx := context.Background()
	db := database.CreateDBConnectionContext(ctx)
	gRPCConn := ram_image_generator.CreateGRPCConnection()

	server := api.NewRamGeneratorServer(ctx, fmt.Sprintf(":%d", config.Conf.Ports.Api), db, gRPCConn)
	err := api.ServeServer(ctx, server)
	if err != nil {
		log.Fatal("server shutdown with error:", err)
	}
}
