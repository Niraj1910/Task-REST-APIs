package config

import (
	"fmt"

	"os"

	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Could not load .env file →", err)
	}

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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	err = db.AutoMigrate(&model.User{}, &model.Task{})
	if err != nil {
		panic("failed to auto-migrate: " + err.Error())
	}

	fmt.Println("→ Database connected and models migrated successfully")
	return db
}
