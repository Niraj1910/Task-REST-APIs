package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string     `gorm:"size:200;not null;index"`
	Description string     `gorm:"type:text"`
	Priority    int        `gorm:"default:0;check:priority >= 0;gte=0;lte=10"`
	Status      string     `gorm:"type:varchar(20);default:'pending'"`
	UserID      uint       `gorm:"index"`
	CompletedAt *time.Time `gorm:"index"`
}
