// swagger:model
// @ignoreEmbedded
package model

import (
	"time"

	"gorm.io/gorm"
)

type EmailVerification struct {
	gorm.Model
	Email        string    `gorm:"size:255;index;not null"`
	TempUsername string    `gorm:"size:50;not null"`
	TempPassword string    `gorm:"size:255;not null"`
	Token        string    `gorm:"size:64;unique;not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	Used         bool      `gorm:"default:false"`
}
