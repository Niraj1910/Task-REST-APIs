package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/Niraj1910/Task-REST-APIs.git/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func GetUserProfile(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		var user model.User
		err := db.Where("id = ?", userID).First(&user).Error
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
				return
			}
			log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch user profile")
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve user profile", "details": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})

	}
}

func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		var userBody struct {
			Name     *string `json:"name" binding:"omitempty,min=5,max=100"`
			Email    *string `json:"email" binding:"omitempty,max=255"`
			Password *string `json:"password" binding:"omitempty,min=5,max=255"`
		}

		err := ctx.ShouldBindBodyWithJSON(&userBody)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid input",
				"details": err.Error(),
			})
			return
		}

		updates := map[string]interface{}{}

		if userBody.Name != nil {
			updates["name"] = userBody.Name
		}

		if userBody.Email != nil {
			// Check if new email is already taken by someone else
			var count int64
			db.Model(&model.User{}).
				Where("email = ? AND id != ?", *userBody.Email, userID).
				Count(&count)
			if count > 0 {
				ctx.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
				return
			}
			updates["email"] = *userBody.Email
		}

		if userBody.Password != nil {
			updates["password"] = utils.HashPassword(*userBody.Password)
		}

		if len(updates) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No fields provided to update"})
			return
		}

		err = db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
		if err != nil {
			log.Error().Err(err).Uint("user_id", userID).Msg("Failed to update user profile")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		var updatedUser model.User
		err = db.First(&updatedUser, userID).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user profile"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":       updatedUser.ID,
			"username": updatedUser.Name,
			"email":    updatedUser.Email,
			"message":  "Profile updated successfully",
		})
	}
}

type TaskListResponse struct {
	Tasks []model.Task `json:"tasks"`
	Meta  struct {
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
	} `json:"meta"`
}

func GetUserTasks(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		pageStr := ctx.DefaultQuery("page", "1")
		limitStr := ctx.DefaultQuery("limit", "10")
		status := ctx.Query("status")
		sort := ctx.DefaultQuery("sort", "created_at:desc")

		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		offset := (page - 1) * limit

		query := db.Where("user_id = ?", userID)

		if status != "" {
			query = query.Where("status = ?", status)
		}

		switch sort {
		case "created_at:asc":
			query = query.Order("created_at ASC")
		case "created_at:desc":
			query = query.Order("created_at DESC")
		case "priority:asc":
			query = query.Order("priority ASC")
		case "priority:desc":
			query = query.Order("priority DESC")
		default:
			query = query.Order("created_at DESC") // fallback
		}

		var total int64
		query.Model(&model.Task{}).Count(&total)

		// 5. Fetch paginated results
		var tasks []model.Task
		err := query.Limit(limit).Offset(offset).Find(&tasks).Error
		if err != nil {
			log.Error().
				Err(err).
				Uint("user_id", userID).
				Msg("Failed to fetch user tasks")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve tasks",
				"details": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, TaskListResponse{
			Tasks: tasks,
			Meta: struct {
				Total int64 `json:"total"`
				Page  int   `json:"page"`
				Limit int   `json:"limit"`
			}{
				Total: total,
				Page:  page,
				Limit: limit,
			},
		})
	}
}
