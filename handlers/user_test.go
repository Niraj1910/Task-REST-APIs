package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Task{}))
	return db
}

func setupContext(method, path, body string, userID uint) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)

	if body != "" {
		c.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, path, nil)
	}

	return c, w
}

func TestGetUserProfile_Success(t *testing.T) {
	db := setupTestDB(t)
	user := model.User{Name: "Niraj", Email: "niraj@example.com"}
	require.NoError(t, db.Create(&user).Error)

	handler := GetUserProfile(db)
	c, w := setupContext(http.MethodGet, "/users/me", "", user.ID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, float64(user.ID), resp["id"])
	assert.Equal(t, user.Name, resp["username"])
	assert.Equal(t, user.Email, resp["email"])
}

func TestGetUserProfile_NotFound(t *testing.T) {
	db := setupTestDB(t)

	handler := GetUserProfile(db)
	c, w := setupContext(http.MethodGet, "/users/me", "", 999) // non-existent ID

	handler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User profile not found")
}

func TestUpdateUser_Success_PartialName(t *testing.T) {
	db := setupTestDB(t)
	user := model.User{Name: "OldName", Email: "test@example.com"}
	require.NoError(t, db.Create(&user).Error)

	handler := UpdateUser(db)
	body := `{"name": "NewName"}`
	c, w := setupContext(http.MethodPatch, "/users/me", body, user.ID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Profile updated successfully")

	var updated model.User
	db.First(&updated, user.ID)
	assert.Equal(t, "NewName", updated.Name)
	assert.Equal(t, "test@example.com", updated.Email) // unchanged
}

func TestUpdateUser_EmailAlreadyTaken(t *testing.T) {
	db := setupTestDB(t)
	user1 := model.User{Name: "User1", Email: "user1@example.com"}
	user2 := model.User{Name: "User2", Email: "taken@example.com"}
	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	handler := UpdateUser(db)
	body := `{"email": "taken@example.com"}`
	c, w := setupContext(http.MethodPatch, "/users/me", body, user1.ID)

	handler(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Email already in use")
}

func TestUpdateUser_NoFieldsProvided(t *testing.T) {
	db := setupTestDB(t)
	user := model.User{Name: "Test"}
	db.Create(&user)

	handler := UpdateUser(db)
	body := `{}` // empty
	c, w := setupContext(http.MethodPatch, "/users/me", body, user.ID)

	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "No fields provided")
}

func TestGetUserTasks_Success_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	userID := uint(42)

	for i := 1; i <= 15; i++ {
		db.Create(&model.Task{
			Title:  fmt.Sprintf("Task %d", i),
			UserID: userID,
		})
	}

	handler := GetUserTasks(db)
	c, w := setupContext(http.MethodGet, "/tasks?page=2&limit=5", "", userID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp TaskListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Len(t, resp.Tasks, 5)
	assert.Equal(t, int64(15), resp.Meta.Total)
	assert.Equal(t, 2, resp.Meta.Page)
	assert.Equal(t, 5, resp.Meta.Limit)
}
