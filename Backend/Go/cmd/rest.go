package main

import (
	"context"
	"database/sql"
	"github.com/KalashnikovProjects/RamGenerator/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/internal/database"
	_ "github.com/KalashnikovProjects/RamGenerator/internal/rest_api"
	"github.com/joho/godotenv"
	"log"
	"time"
)

func init() {
	godotenv.Load("../../.env", ".env")
}

func main() {
	conf := config.New()
	log.Println("Загрузка базы данных")
	var db *sql.DB
	_ = db // TODO: убрать
	var err error
	pgPort := 5432
	connectionString := database.GeneratePostgresConnectionString(conf.Database.User, conf.Database.Password, conf.Database.Host, pgPort, conf.Database.DBName)
	for {
		db, err = database.OpenDb(context.TODO(), connectionString)
		if err == nil {
			break
		}
		log.Print("Ща будет retry после ошибки с дб:", err)
		time.Sleep(2 * time.Second)
	}
	log.Print("Чекнул базу данных, всё круто!")
}
