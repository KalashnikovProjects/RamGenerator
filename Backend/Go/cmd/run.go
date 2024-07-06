package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/internal/api"
	"github.com/KalashnikovProjects/RamGenerator/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	"github.com/KalashnikovProjects/RamGenerator/internal/ram_image_generator"
	pb "github.com/KalashnikovProjects/RamGenerator/proto_generated"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	godotenv.Load("../../.env", ".env")
}

// TestContext TODO: ПЕРЕНЕСТИ В ТЕСТЫ, сделать автоматическими
func TestContext(ctx context.Context, db *sql.DB, gRPCConn pb.RamGeneratorClient) {
	id, err := database.CreateUserContext(ctx, db, entities.User{Username: "alo", PasswordHash: "adwdawd", LastRamGenerated: 11111})
	if err != nil {
		log.Fatal("Ошибка при CreateUserContext: ", err)
	}
	log.Print("Создал пользователя: ", id)

	err = database.UpdateUserContext(ctx, db, id, entities.User{Username: "pukek ultra 2", LastRamGenerated: 2})
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
	log.Println("Запуск go сервиса...")

	ctx := context.Background()
	db := database.CreateDBConnectionContext(ctx)
	gRPCConn := ram_image_generator.CreateGRPCConnection()

	server := api.NewRamGeneratorServer(ctx, ":80", db, gRPCConn)
	err := api.ServeServer(ctx, server)
	if err != nil {
		log.Fatal("server shutdown with error:", err)
	}
}
