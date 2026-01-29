package handlers

import (
	"net/http"
	"time"

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

		// check if the user's email already exists
		var existing model.User
		err = db.Where("email = ?", input.Email).Find(&existing).Error
		if err != nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "email already  exists",
			})
			return
		}

		user := model.User{
			Name:     input.UserName,
			Email:    input.Email,
			Password: utils.HashPassword(input.Password),
		}

		err = db.Create(&user).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to create user",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":        user.ID,
			"user name": user.Name,
			"email":     user.Email,
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
			ctx.JSON(http.StatusBadRequest, gin.H{
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
