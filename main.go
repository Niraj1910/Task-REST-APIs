package main

import (
	"github.com/Niraj1910/Task-REST-APIs.git/config"
	"github.com/Niraj1910/Task-REST-APIs.git/handlers"
	"github.com/Niraj1910/Task-REST-APIs.git/middlewares"
	"github.com/gin-gonic/gin"
)

func main() {
	// connect to DB
	db := config.ConnectDB()

	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "pong"})
	})

	router.POST("/register", handlers.RegisterUser(db))
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
