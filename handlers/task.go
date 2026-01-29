package handlers

import (
	"net/http"
	"strconv"

	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/Niraj1910/Task-REST-APIs.git/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTask(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		var taskBody struct {
			Title       string `json:"title" binding:"required,min=5,max=200"`
			Description string `json:"description" binding:"required,omitempty,max=1000"`
		}
		err := ctx.ShouldBindBodyWithJSON(&taskBody)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid input",
				"details": err.Error(),
			})
			return
		}

		task := model.Task{
			Title:       taskBody.Title,
			Description: taskBody.Description,
			UserID:      userID,
		}

		err = db.Create(&task).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"id": task.ID, "title": task.Title, "description": task.Description, "userId": task.UserID})

	}
}

func UpdateTask(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		idstr := ctx.Param("id")
		taskID, err := strconv.ParseInt(idstr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Task ID"})
			return
		}

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		var taskBody struct {
			Title       string `json:"title" binding:"omitempty,min=5,max=200"`
			Description string `json:"description" binding:"omitempty,max=1000"`
			Priority    int    `json:"priority" binding:"omitempty,gte=0,lte=10"`
			Status      string `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
		}

		err = ctx.ShouldBindBodyWithJSON(&taskBody)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid input",
				"details": err.Error(),
			})
			return
		}

		updates := map[string]interface{}{}
		if taskBody.Title != "" {
			updates["title"] = taskBody.Title
		}
		if taskBody.Description != "" {
			updates["description"] = taskBody.Description
		}
		if taskBody.Priority >= 0 && taskBody.Priority <= 10 {
			updates["priority"] = taskBody.Priority
		}
		if taskBody.Status != "" {
			updates["status"] = taskBody.Status
		}

		if len(updates) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No fields provided to update"})
			return
		}

		result := db.Model(&model.Task{}).Where("ID = ? AND user_id = ?", uint(taskID), userID).Updates(updates)

		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update task",
				"details": result.Error.Error(),
			})
			return
		}

		if result.RowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found or not owned by you"})
			return
		}

		var updatedTask model.Task
		err = db.First(&updatedTask, taskID).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated task"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":          updatedTask.ID,
			"title":       updatedTask.Title,
			"description": updatedTask.Description,
			"priority":    updatedTask.Priority,
			"status":      updatedTask.Status,
			"user_id":     updatedTask.UserID,
		})

	}
}

func DeleteTask(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		idStr := ctx.Param("id")
		taskId, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Task ID"})
			return
		}

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		result := db.Where("id = ? AND user_id = ?", taskId, userID).Delete(&model.Task{})
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the task"})
			return
		}

		if result.RowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found or not owned by you"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted task", "task_id": taskId})
	}
}

func GetTaskByID(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		idStr := ctx.Param("id")
		taskID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Task ID"})
			return
		}

		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		var task model.Task

		err = db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found or not owned by you"})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"priority":    task.Priority,
			"status":      task.Status,
			"user_id":     task.UserID,
		})
	}
}

func GetTasks(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		var tasks []model.Task

		err := db.Where("user_id = ?", userID).Find(&tasks).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks", "details": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"count": len(tasks),
			"Tasks": tasks,
		})
	}
}
