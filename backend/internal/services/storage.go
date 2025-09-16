package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/Morningstarl2504/Balkanid_repo/internal/database"
	"github.com/Morningstarl2504/Balkanid_repo/internal/models"
)

type StorageService struct {
	uploadPath string
}

func NewStorageService(uploadPath string) *StorageService {
	// Create upload directory if it doesn't exist
	os.MkdirAll(uploadPath, 0755)
	return &StorageService{
		uploadPath: uploadPath,
	}
}

func (s *StorageService) GetStoragePath(contentHash string) string {
	// Create subdirectories based on hash prefix for better file organization
	prefix := contentHash[:2]
	dir := filepath.Join(s.uploadPath, prefix)
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, contentHash)
}

func (s *StorageService) GetUserStorageStats(userID uint) (*models.StorageStats, error) {
	var stats struct {
		TotalFiles    int64
		TotalSize     int64
		OriginalSize  int64
	}

	// Get total storage used (deduplicated)
	err := database.DB.Table("files").
		Select("COUNT(*) as total_files, COALESCE(SUM(fc.file_size), 0) as total_size").
		Joins("JOIN file_contents fc ON files.content_id = fc.id").
		Where("files.owner_id = ? AND files.deleted_at IS NULL", userID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Get original size (without deduplication)
	err = database.DB.Table("files").
		Select("COALESCE(SUM(fc.file_size), 0) as original_size").
		Joins("JOIN file_contents fc ON files.content_id = fc.id").
		Where("files.owner_id = ? AND files.deleted_at IS NULL", userID).
		Group("files.id").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Get user quota
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	savingsBytes := stats.OriginalSize - stats.TotalSize
	savingsPercentage := float64(0)
	if stats.OriginalSize > 0 {
		savingsPercentage = (float64(savingsBytes) / float64(stats.OriginalSize)) * 100
	}

	return &models.StorageStats{
		TotalUsed:         stats.TotalSize,
		OriginalSize:      stats.OriginalSize,
		SavingsBytes:      savingsBytes,
		SavingsPercentage: savingsPercentage,
		Quota:             user.StorageQuota,
	}, nil
}

func (s *StorageService) CheckStorageQuota(userID uint, additionalSize int64) error {
	stats, err := s.GetUserStorageStats(userID)
	if err != nil {
		return err
	}

	if stats.TotalUsed + additionalSize > stats.Quota {
		return errors.New("storage quota exceeded")
	}

	return nil
}

func (s *StorageService) FindOrCreateFileContent(filePath, contentHash, mimeType string, fileSize int64) (*models.FileContent, error) {
	var fileContent models.FileContent

	// Try to find existing content
	err := database.DB.Where("content_hash = ?", contentHash).First(&fileContent).Error
	if err == nil {
		// File already exists, increment reference count
		database.DB.Model(&fileContent).UpdateColumn("reference_count", gorm.Expr("reference_count + ?", 1))
		return &fileContent, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new file content record
	storagePath := s.GetStoragePath(contentHash)
	
	// Move file to storage location
	if err := os.Rename(filePath, storagePath); err != nil {
		return nil, fmt.Errorf("failed to move file to storage: %v", err)
	}

	fileContent = models.FileContent{
		ContentHash:    contentHash,
		FilePath:       storagePath,
		FileSize:       fileSize,
		MimeType:       mimeType,
		ReferenceCount: 1,
	}

	if err := database.DB.Create(&fileContent).Error; err != nil {
		// Cleanup file if database insert fails
		os.Remove(storagePath)
		return nil, err
	}

	return &fileContent, nil
}

func (s *StorageService) DeleteFileContent(contentID uint) error {
	var fileContent models.FileContent
	if err := database.DB.First(&fileContent, contentID).Error; err != nil {
		return err
	}

	// Decrement reference count
	result := database.DB.Model(&fileContent).UpdateColumn("reference_count", gorm.Expr("reference_count - ?", 1))
	if result.Error != nil {
		return result.Error
	}

	// If no more references, delete the actual file and record
	if err := database.DB.First(&fileContent, contentID).Error; err != nil {
		return err
	}

	if fileContent.ReferenceCount <= 0 {
		// Delete physical file
		if err := os.Remove(fileContent.FilePath); err != nil && !os.IsNotExist(err) {
			return err
		}

		// Delete database record
		return database.DB.Delete(&fileContent).Error
	}

	return nil
}