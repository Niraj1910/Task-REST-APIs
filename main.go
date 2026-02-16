package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/robfig/cron/v3"

	"github.com/Niraj1910/Task-REST-APIs/config"
	"github.com/Niraj1910/Task-REST-APIs/handlers"
	"github.com/Niraj1910/Task-REST-APIs/middlewares"
	"github.com/Niraj1910/Task-REST-APIs/model"
	"github.com/Niraj1910/Task-REST-APIs/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/Niraj1910/Task-REST-APIs/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Golang Task REST API
// @version         1.0
// @description     Secure task management API with Golang/Gin Framework, JWT authentication and email verification.

// @contact.name    Niraj Shaw
// @contact.url     https://github.com/Niraj1910
// @contact.email   nirazshaw156@gmail.com

// @BasePath  /
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token (format: Bearer <token>)

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

	orgins := []string{"http://localhost:3000", "http://localhost:5173"}
	cliProd := os.Getenv("CLIENT_PROD_URL")
	if cliProd != "" {
		orgins = append(orgins, cliProd)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     orgins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.GET("/ping", Ping)

	router.GET("/", func(ctx *gin.Context) {

		scheme := "https"
		if os.Getenv("ENV") == "development" {
			scheme = "http"
		}
		baseUrl := scheme + "://" + ctx.Request.Host
		ctx.JSON(http.StatusOK, gin.H{
			"message": "This is Gin server for Task Management REST APIs",
			"info":    "Interactive API documentation is available here:",
			"docs":    baseUrl + "/swagger/index.html",
			"ping":    baseUrl + "/ping",
			"status":  "API is running"})
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

	protectedUserRoute := router.Group("/api/user", middlewares.AuthMiddleware)
	{
		protectedUserRoute.GET("/profile", handlers.GetUserProfile(db))
		protectedUserRoute.GET("/task", handlers.GetUserTasks(db))
		protectedUserRoute.PATCH("/update", handlers.UpdateUser(db))
		// }

		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		router.Run(":4000")
	}
}

// Ping godoc
// @Summary      Health check endpoint
// @Description  Returns a simple pong message to verify the API is running
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string "pong"
// @Router       /ping [get]
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
