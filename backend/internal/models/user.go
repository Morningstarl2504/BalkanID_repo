package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"unique;not null"`
	Email        string         `json:"email" gorm:"unique;not null"`
	PasswordHash string         `json:"-" gorm:"not null"`
	IsAdmin      bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Files []File `json:"files" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}