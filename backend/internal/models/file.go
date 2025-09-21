// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/models/file.go
package models

import (
	"time"
	"gorm.io/gorm"
)

type File struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	UserID           uint           `json:"user_id" gorm:"not null;index"`
	FileContentID    uint           `json:"-" gorm:"not null;index"`
	OriginalFilename string         `json:"original_filename" gorm:"not null"`
	IsPublic         bool           `json:"is_public" gorm:"default:false"`
	DownloadCount    int            `json:"download_count" gorm:"default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User    User        `json:"user" gorm:"foreignKey:UserID"`
	Content FileContent `json:"content" gorm:"foreignKey:FileContentID"`
}

func (File) TableName() string {
	return "files"
}