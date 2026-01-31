package main

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/robfig/cron/v3"

	"github.com/Niraj1910/Task-REST-APIs.git/config"
	"github.com/Niraj1910/Task-REST-APIs.git/handlers"
	"github.com/Niraj1910/Task-REST-APIs.git/middlewares"
	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/Niraj1910/Task-REST-APIs.git/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	utils.InitLogger()

	if err != nil {
		log.Warn().Err(err).Msg("Could not load .env")
	}

	// connect to DB
	db := config.ConnectDB()

	// clean up the registered user's email temp data
	c := cron.New()
	c.AddFunc("@hourly", func() {
		err = db.Where("expires_at < ?", time.Now()).Delete(&model.EmailVerification{}).Error
		if err != nil {
			log.Error().Err(err).Msg("Failed to delete tempData @hourly from verification table ")
		}
	})
	c.Start()

	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "pong"})
	})

	router.POST("/register", handlers.RegisterUser(db))
	router.GET("/verify", handlers.VerifyEmailAndRegisterUser(db))
	router.POST("/login", handlers.LoginUser(db))
	router.POST("/logout", handlers.LogoutUser)

	protectedTaskRoute := router.Group("/api/task", middlewares.AuthMiddleware)
	{
		protectedTaskRoute.GET("/", handlers.GetTasks(db))
		protectedTaskRoute.POST("/new", handlers.CreateTask(db))
		protectedTaskRoute.PUT("/:id", handlers.UpdateTask(db))
		protectedTaskRoute.GET("/:id", handlers.GetTaskByID(db))
		protectedTaskRoute.DELETE("/:id", handlers.DeleteTask(db))

	}

	router.Run(":4000")
}
