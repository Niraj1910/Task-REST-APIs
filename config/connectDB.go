package config

import (
	"fmt"

	"os"

	"github.com/Niraj1910/Task-REST-APIs/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {

	host := os.Getenv("HOST")
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")
	dbName := os.Getenv("DBNAME")
	port := os.Getenv("PORT")
	sslMode := os.Getenv("SSLMODE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbName, port, sslMode)

	if os.Getenv("POSTGRES_URI") != "" {
		dsn = os.Getenv("POSTGRES_URI")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	err = db.AutoMigrate(&model.User{}, &model.Task{}, &model.EmailVerification{})
	if err != nil {
		panic("failed to auto-migrate: " + err.Error())
	}

	log.Info().Msg("Database connected and models migrated successfully")
	return db
}
