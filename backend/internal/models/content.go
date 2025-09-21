// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/models/content.go
package models

// FileContent represents the actual content of a file, stored once for deduplication.
type FileContent struct {
	ID             uint   `json:"-" gorm:"primaryKey"`
	SHA256Hash     string `json:"sha256_hash" gorm:"uniqueIndex;not null"`
	FileSize       int64  `json:"file_size" gorm:"not null"`
	MimeType       string `json:"mime_type" gorm:"not null"`
	ReferenceCount uint   `json:"-" gorm:"not null;default:1"`
}

func (FileContent) TableName() string {
	return "file_contents"
}