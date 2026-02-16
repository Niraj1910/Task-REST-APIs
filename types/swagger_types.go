package types

// types/swagger_types.go

import "time"

// These types are ONLY for Swagger documentation.
// They mirror the real model types but without embedded gorm.Model
// so swag can parse them correctly.

// @Schema
type SwaggerTask struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`

	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Status      string `json:"status"`
	UserID      uint   `json:"user_id"`
}

// @Schema
type SwaggerTaskListResponse struct {
	Tasks []SwaggerTask `json:"tasks"`
	Meta  struct {
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
	} `json:"meta"`
}

// SwaggerUser godoc
// @Schema
type SwaggerUser struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`

	Name  string `json:"name"`
	Email string `json:"email"`
	// Add other fields from your real User model (age, role, is_active, etc.)
	Age      int    `json:"age,omitempty"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
}

// SwaggerUserProfileResponse godoc
// @Schema
type SwaggerUserProfileResponse struct {
	User SwaggerUser `json:"user"`
}
