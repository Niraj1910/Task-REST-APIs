package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/Niraj1910/Task-REST-APIs.git/auth"
	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/Niraj1910/Task-REST-APIs.git/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegisterUserBody struct {
	UserName        string `json:"username" binding:"required,min=3,max=20"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=50"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

type LoginUserBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=50"`
}

func RegisterUser(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// validate the req body
		var input RegisterUserBody
		err := ctx.ShouldBindBodyWithJSON(&input)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid input",
				"details": err.Error(),
			})
			return
		}

		if input.Password != input.ConfirmPassword {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Password do not match"})
			return
		}

		// check if the user's email already exists
		var count int64
		db.Model(&model.User{}).Where("email = ?", input.Email).Count(&count)
		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"err": "Email already registered"})
			return
		}

		token := uuid.New().String()
		expires := time.Now().Add(10 * time.Minute)

		// store temp data
		verification := model.EmailVerification{
			Email:        input.Email,
			TempUsername: input.UserName,
			TempPassword: utils.HashPassword(input.Password),
			Token:        token,
			ExpiresAt:    expires,
		}

		err = db.Create(&verification).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
			return
		}

		// Build + send the email to confirm the registration
		baseUrl := os.Getenv("PROD_URL")
		if baseUrl == "" {
			baseUrl = "http://localhost:4000"
		}
		verifyLink := fmt.Sprintf("%s/verify?token=%s&email=%s", baseUrl, token, url.QueryEscape(input.Email))

		go func() {
			err := utils.SendVerificationMail(input.UserName, input.Email, verifyLink)
			if err != nil {
				log.Error().Err(err).Str("email", input.Email).Msg("Failed to send verification email")
			}
		}()

		ctx.JSON(http.StatusCreated, gin.H{
			"message": "Registration request received! Please check your email to verify and complete signup.",
			"email":   input.Email,
		})
	}
}

func VerifyEmailAndRegisterUser(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token := ctx.Query("token")
		email := ctx.Query("email")

		if token == "" || email == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing token or email in verification link"})
			return
		}

		// start transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.Error().Err(tx.Error).Msg("Failed to start transaction")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Error().Interface("panic", r).Msg("Panic in verification – rollback")
			}
		}()

		// find verification record
		var verification model.EmailVerification
		err := tx.Where("token = ? AND email = ? AND expires_at > ? AND used = ?", token, email, time.Now(), false).First(&verification).Error
		if err != nil {

			tx.Rollback()

			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid, expired, or already used verification link. Please register again.",
				})
				return
			}
			log.Error().Err(err).Msg("Database error during verification")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Verification failed due to a server error",
				"details": "Please try again later or contact support",
			})
		}

		user := model.User{
			Name:     verification.TempUsername,
			Email:    verification.Email,
			Password: verification.TempPassword,
		}

		err = db.Create(&user).Error
		if err != nil {

			tx.Rollback()

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to activate account – please try again or contact support",
			})
			return
		}

		// commit transaction
		if err := tx.Commit().Error; err != nil {
			log.Error().Err(err).Msg("Transaction commit failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server error during commit"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Email verified successfully! Your account is now active.",
			"email":   user.Email,
			"next":    "You can now log in at /login",
		})
	}
}

func LoginUser(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var input LoginUserBody
		err := ctx.ShouldBindJSON(&input)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "incorrect input",
				"details": err.Error(),
			})
			return
		}

		var user struct {
			ID       uint   `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err = db.Model(&model.User{}).Where("email = ?", input.Email).First(&user).Error
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "user not found",
				"details": err.Error(),
			})
			return
		}

		// check password
		if !utils.CompareHashedPassword(user.Password, input.Password) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "incorrect password",
			})
			return
		}

		// create the jwt token
		token, err := auth.CreateToken(user.Name, user.Email, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "jwt token creation failed",
				"details": err.Error(),
			})
			return
		}

		ctx.SetCookie("token", token, int(time.Now().Add(5*time.Minute).Unix()), "", "", true, true)

		ctx.JSON(http.StatusOK, gin.H{
			"message": "successfully logged in",
			"token":   token,
		})
	}
}

func LogoutUser(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "", "", true, true)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
