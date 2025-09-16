package models

import (
	"time"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type FileContent struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ContentHash   string `json:"content_hash" gorm:"uniqueIndex;not null;size:64"`
	FilePath      string `json:"file_path" gorm:"not null;size:500"`
	FileSize      int64  `json:"file_size" gorm:"not null"`
	MimeType      string `json:"mime_type" gorm:"not null"`
	ReferenceCount int   `json:"reference_count" gorm:"default:1"`
	CreatedAt     time.Time `json:"created_at"`
}

func (FileContent) TableName() string {
	return "file_contents"
}

type File struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Filename         string         `json:"filename" gorm:"not null"`
	OriginalFilename string         `json:"original_filename" gorm:"not null"`
	ContentID        uint           `json:"content_id"`
	Content          FileContent    `json:"content" gorm:"foreignKey:ContentID"`
	OwnerID          uint           `json:"owner_id"`
	Owner            User           `json:"owner" gorm:"foreignKey:OwnerID"`
	FolderID         *uint          `json:"folder_id"`
	Folder           *Folder        `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	IsPublic         bool           `json:"is_public" gorm:"default:false"`
	DownloadCount    int            `json:"download_count" gorm:"default:0"`
	Tags             pq.StringArray `json:"tags" gorm:"type:text[]"`
	UploadedAt       time.Time      `json:"uploaded_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

func (File) TableName() string {
	return "files"
}

type Folder struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	OwnerID   uint      `json:"owner_id"`
	Owner     User      `json:"owner" gorm:"foreignKey:OwnerID"`
	ParentID  *uint     `json:"parent_id"`
	Parent    *Folder   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	IsPublic  bool      `json:"is_public" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

func (Folder) TableName() string {
	return "folders"
}

type FileShare struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FileID         uint      `json:"file_id"`
	File           File      `json:"file" gorm:"foreignKey:FileID"`
	SharedWithUserID uint    `json:"shared_with_user_id"`
	SharedWithUser User      `json:"shared_with_user" gorm:"foreignKey:SharedWithUserID"`
	SharedByUserID uint      `json:"shared_by_user_id"`
	SharedByUser   User      `json:"shared_by_user" gorm:"foreignKey:SharedByUserID"`
	CreatedAt      time.Time `json:"created_at"`
}

func (FileShare) TableName() string {
	return "file_shares"
}

type AuditLog struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	UserID       uint                   `json:"user_id"`
	User         User                   `json:"user" gorm:"foreignKey:UserID"`
	Action       string                 `json:"action" gorm:"not null;size:50"`
	ResourceType string                 `json:"resource_type" gorm:"not null;size:50"`
	ResourceID   uint                   `json:"resource_id" gorm:"not null"`
	Details      map[string]interface{} `json:"details" gorm:"type:jsonb"`
	CreatedAt    time.Time              `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}