package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Niraj1910/Task-REST-APIs.git/model"
	"github.com/Niraj1910/Task-REST-APIs.git/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

func init() {
	os.Setenv("JWT_SECRET", "test-secret-key-1234567890abcdef")
}

func TestRegisterUser_DuplicateEmail(t *testing.T) {
	// Setup in-memory DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&model.User{}, &model.EmailVerification{})
	require.NoError(t, err)

	// Pre-create a user with email "test@gmai.com"
	db.Create(&model.User{Email: "test@example.com"})

	// Create test Handler
	testHandler := RegisterUser(db)

	body := `{
		"username": "testuser",
		"email": "test@example.com",
		"password": "strongpass123",
		"confirmPassword": "strongpass123"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rc)
	c.Request = req

	// call handler
	testHandler(c)

	assert.Equal(t, http.StatusConflict, rc.Code)
	assert.Contains(t, rc.Body.String(), "Email already registered")
}

func TestRegisterUser_Success_CreatesVerification(t *testing.T) {

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&model.User{}, &model.EmailVerification{})
	require.NoError(t, err)

	testHandler := RegisterUser(db)

	body := `{
        "username": "niraj",
        "email": "niraj@example.com",
        "password": "strongpass123",
        "confirmPassword": "strongpass123"
    }`

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rc)
	c.Request = req

	testHandler(c)

	assert.Equal(t, http.StatusCreated, rc.Code)
	assert.Contains(t, rc.Body.String(), "Registration request received")

	// Check DB: no real user yet
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	assert.Equal(t, int64(0), userCount, "real user should not be created yet")

	// Check verification record exists
	var ver model.EmailVerification
	err = db.Where("email = ?", "niraj@example.com").First(&ver).Error
	assert.NoError(t, err)
	assert.False(t, ver.Used)
	assert.NotEmpty(t, ver.Token)
	assert.True(t, ver.ExpiresAt.After(time.Now()))
}

func TestLoginUser_Success_ReturnsToken(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{})

	// Pre-create a user
	hashed := utils.HashPassword("strongpass123")
	db.Create(&model.User{
		Name:     "Niraj",
		Email:    "niraj@example.com",
		Password: hashed,
	})

	testHandler := LoginUser(db)

	body := `{
        "email": "niraj@example.com",
        "password": "strongpass123"
    }`

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	testHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "successfully logged in")
	assert.Contains(t, w.Body.String(), "token") // token should be in response
}

func TestLoginUser_WrongPassword_Returns401(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{})

	hashed := utils.HashPassword("strongpass123")
	db.Create(&model.User{Email: "niraj@example.com", Password: hashed})

	handler := LoginUser(db)

	body := `{
        "email": "niraj@example.com",
        "password": "wrongpass"
    }`

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "incorrect password")
}
