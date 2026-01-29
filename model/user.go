package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"size:100;not null;index"`
	Email    string `gorm:"unique;size:255;not null"`
	Password string `gorm:"varchar(100); not null"`
	Age      uint8
	IsActive bool   `gorm:"default:true"`
	Role     string `gorm:"varchar(20);default:'user'"`
}
