package handlers

import (
	"net/http"
	"strconv"

	"github.com/Niraj1910/Task-REST-APIs/model"
	_ "github.com/Niraj1910/Task-REST-APIs/types"
	"github.com/Niraj1910/Task-REST-APIs/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateTask godoc
// @Summary      Create a new task
// @Description  Creates a task owned by the authenticated user
// @Tags         Tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body model.Task true "Task data (title required)"
// @Success      201 {object} model.Task
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Server error"
// @Router       /api/task/new [post]
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

// UpdateTask godoc
// @Summary      Update a task
// @Description  Updates task fields (partial update allowed) if owned by the user
// @Tags         Tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Task ID"
// @Param        body body model.Task true "Updated fields"
// @Success      200 {object} model.Task
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Task not found or not owned"
// @Failure      500 {object} map[string]string "Server error"
// @Router       /api/task/{id} [put]
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

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Deletes a task if it belongs to the authenticated user
// @Tags         Tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "Task ID"
// @Success      200 {object} map[string]string "Task deleted"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Task not found or not owned"
// @Router       /api/task/{id} [delete]
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

// GetTaskByID godoc
// @Summary      Get a single task by ID
// @Description  Returns a task if it belongs to the authenticated user
// @Tags         Tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "Task ID"
// @Success      200 {object} model.Task
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Task not found or not owned"
// @Router       /api/task/{id} [get]
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

// GetTasks godoc
// @Summary      List authenticated user's tasks
// @Description  Returns paginated list of tasks belonging to the current user
// @Tags         Tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        page   query     int     false  "Page number"                  default(1)
// @Param        limit  query     int     false  "Items per page"               default(10)
// @Param        status query     string  false  "Filter by status (pending, completed, etc.)"
// @Success      200     {object} types.SwaggerTaskListResponse
// @Failure      401     {object} map[string]string "Unauthorized"
// @Router       /api/task [get]
func GetTasks(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := utils.UserIDFromContext(ctx)
		if !ok {
			return
		}

		pageStr := ctx.DefaultQuery("page", "1")
		limitStr := ctx.DefaultQuery("limit", "10")
		sort := ctx.DefaultQuery("sort", "created_at:desc")
		status := ctx.Query("status")

		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)

		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100 // prevent huge responses
		}

		offset := (page - 1) * limit

		// build base query
		query := db.Where("user_id = ?", userID)

		if status != "" {
			query = query.Where("status = ?", status)
		}

		// Optional sorting
		switch sort {
		case "created_at:asc":
			query = query.Order("created_at ASC")
		case "priority:asc":
			query = query.Order("priority ASC")
		case "priority:desc":
			query = query.Order("priority DESC")
		default:
			query = query.Order("created_at DESC") // newest first
		}

		var total int64
		query.Model(&model.Task{}).Count(&total)

		var tasks []model.Task

		err := query.Limit(limit).Offset(offset).Find(&tasks).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks", "details": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"Total": total,
			"Page":  page,
			"Tasks": tasks,
		})
	}
}
