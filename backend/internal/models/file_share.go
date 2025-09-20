package models

import (
	"time"
	"gorm.io/gorm"
)

type FileShare struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	FileID    uint           `json:"file_id" gorm:"not null;index"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	ShareWith *uint          `json:"share_with" gorm:"index"`
	ShareURL  string         `json:"share_url" gorm:"unique"`
	ExpiresAt *time.Time     `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	File       File  `json:"file" gorm:"foreignKey:FileID"`
	User       User  `json:"user" gorm:"foreignKey:UserID"`
	SharedWith *User `json:"shared_with" gorm:"foreignKey:ShareWith"`
}

func (FileShare) TableName() string {
	return "file_shares"
}