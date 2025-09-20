package models

import "gorm.io/gorm"

// FileContent represents the actual content of a file, stored once for deduplication.
type FileContent struct {
    gorm.Model
    Hash           string `gorm:"uniqueIndex"` // SHA-256 hash of the file content
    Path           string
    Size           int64
    ReferenceCount int `gorm:"default:1"`
}
