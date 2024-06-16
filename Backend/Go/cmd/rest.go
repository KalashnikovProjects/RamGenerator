package main

import (
	"context"
	"github.com/KalashnikovProjects/RamGenerator/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/internal/database"
	_ "github.com/KalashnikovProjects/RamGenerator/internal/rest_api"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	if err := godotenv.Load("../../.env", ".env"); err != nil {
		log.Print("No .env file found (all is ok)")
	}
}

func main() {
	conf := config.New()
	_, err := database.OpenDb(context.TODO(), conf.Database.ConnectionString)
	log.Print("Чекнул базу данных", err)
}
