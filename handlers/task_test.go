package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Niraj1910/Task-REST-APIs.git/handlers"
	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateTask_OwnershipContext(t *testing.T) {

	db := setupTestDB(t)

	testHandler := handlers.CreateTask(db)

	rc := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rc)
	c.Set("user_id", uint(42))

	body := `{"title": "Buy milk", "description": "Whole milk"}`
	req := httptest.NewRequest("POST", "/api/task", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	testHandler(c)

	assert.Equal(t, http.StatusCreated, rc.Code)

	var task model.Task
	db.First(&task)
	assert.Equal(t, uint(42), task.UserID)
	assert.Equal(t, "Buy milk", task.Title)

}
