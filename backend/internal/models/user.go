// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/models/user.go
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
	StorageQuota int64          `json:"storage_quota" gorm:"default:10485760"` // Default 10MB
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Files []File `json:"files" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}