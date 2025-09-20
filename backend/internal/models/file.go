package models

import (
	"time"
	"gorm.io/gorm"
)

type File struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	UserID           uint           `json:"user_id" gorm:"not null;index"`
	Filename         string         `json:"filename" gorm:"not null;unique"`
	OriginalFilename string         `json:"original_filename" gorm:"not null"`
	MimeType         string         `json:"mime_type"`
	Size             int64          `json:"size" gorm:"not null"`
	IsPublic         bool           `json:"is_public" gorm:"default:false"`
	DownloadCount    int            `json:"download_count" gorm:"default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

func (File) TableName() string {
	return "files"
}

// Response format for API (using different names to avoid conflicts)
type FileResponse struct {
	ID               uint        `json:"id"`
	Filename         string      `json:"filename"`
	OriginalFilename string      `json:"original_filename"`
	Content          ContentInfo `json:"content"`
	Owner            OwnerInfo   `json:"owner"`
	UploadedAt       time.Time   `json:"uploaded_at"`
	DownloadCount    int         `json:"download_count"`
	IsPublic         bool        `json:"is_public"`
}

type ContentInfo struct {
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

type OwnerInfo struct {
	Username string `json:"username"`
}