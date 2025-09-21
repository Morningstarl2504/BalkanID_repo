// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/services/file.go
package services

import (
	"errors"
	"filevault-backend/internal/database"
	"filevault-backend/internal/models"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type StorageService struct {
	UploadPath string
}

func NewStorageService(uploadPath string) *StorageService {
	return &StorageService{
		UploadPath: uploadPath,
	}
}

func (s *StorageService) SaveFile(filename string, file multipart.File) error {
	filePath := filepath.Join(s.UploadPath, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	return err
}

func (s *StorageService) Get(filename string) ([]byte, error) {
	filePath := filepath.Join(s.UploadPath, filename)
	return os.ReadFile(filePath)
}

func (s *StorageService) Delete(filename string) error {
	filePath := filepath.Join(s.UploadPath, filename)
	return os.Remove(filePath)
}

type FileService struct {
	storageService *StorageService
	maxFileSize    int64
}

func NewFileService(storageService *StorageService, maxFileSize int64) *FileService {
	return &FileService{
		storageService: storageService,
		maxFileSize:    maxFileSize,
	}
}

func (s *FileService) Create(file *models.File) error {
	return database.DB.Create(file).Error
}

func (s *FileService) GetByUserID(userID uint, filters *models.SearchFilters) ([]*models.File, error) {
	var files []*models.File
	query := database.DB.Preload("Content").Preload("User").Where("user_id = ?", userID)
	query = query.Joins("JOIN file_contents ON file_contents.id = files.file_content_id")

	if filters.Filename != "" {
		query = query.Where("files.original_filename LIKE ?", "%"+filters.Filename+"%")
	}
	if filters.MimeType != "" {
		query = query.Where("file_contents.mime_type LIKE ?", filters.MimeType+"%")
	}
	if filters.MinSize > 0 {
		query = query.Where("file_contents.file_size >= ?", filters.MinSize)
	}
	if filters.MaxSize > 0 {
		query = query.Where("file_contents.file_size <= ?", filters.MaxSize)
	}
	if filters.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			query = query.Where("files.created_at >= ?", t)
		}
	}
	if filters.EndDate != "" {
		if t, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			query = query.Where("files.created_at <= ?", t.Add(24*time.Hour-time.Nanosecond))
		}
	}

	err := query.Order("files.created_at DESC").Find(&files).Error
	return files, err
}

func (s *FileService) GetByID(id uint) (*models.File, error) {
	var file models.File
	err := database.DB.Preload("Content").Preload("User").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// THE FIX: This function is now more robust.
func (s *FileService) DeleteFileAndContent(fileID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Find the file record to get the content ID
		var fileToDelete models.File
		if err := tx.First(&fileToDelete, fileID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // File already deleted, success.
			}
			return err
		}
		contentID := fileToDelete.FileContentID

		// 2. Delete the specific file record
		if err := tx.Delete(&models.File{}, fileID).Error; err != nil {
			return err
		}

		// 3. Count how many other files still reference the same content
		var remainingReferences int64
		tx.Model(&models.File{}).Where("file_content_id = ?", contentID).Count(&remainingReferences)

		// 4. If no files are left referencing this content, delete the content and the physical file
		if remainingReferences == 0 {
			var contentToDelete models.FileContent
			if err := tx.First(&contentToDelete, contentID).Error; err != nil {
				// Content might already be gone, which is fine.
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return err
			}

			// Delete physical file from storage
			if err := s.storageService.Delete(contentToDelete.SHA256Hash); err != nil {
				// Log the error but don't fail the transaction, as the DB record is more critical.
				// In a real app, a cleanup job would handle orphaned files.
			}

			// Delete the content record from the database
			return tx.Delete(&contentToDelete).Error
		} else {
			// Optional: Keep the reference_count column in sync for statistics
			tx.Model(&models.FileContent{}).Where("id = ?", contentID).Update("reference_count", remainingReferences)
		}

		return nil
	})
}


func (s *FileService) IncrementDownloadCount(id uint) error {
	return database.DB.Model(&models.File{}).Where("id = ?", id).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}

func (s *FileService) GetStorageStats(userID uint) (*models.StorageStats, error) {
	var userFiles []models.File
	if err := database.DB.Preload("Content").Where("user_id = ?", userID).Find(&userFiles).Error; err != nil {
		return nil, err
	}

	var originalSize int64 = 0
	uniqueContents := make(map[uint]models.FileContent)
	for _, file := range userFiles {
		originalSize += file.Content.FileSize
		if _, ok := uniqueContents[file.FileContentID]; !ok {
			uniqueContents[file.FileContentID] = file.Content
		}
	}

	var totalUsed int64 = 0
	for _, content := range uniqueContents {
		totalUsed += content.FileSize
	}

	var user models.User
	database.DB.First(&user, userID)

	savingsBytes := originalSize - totalUsed
	savingsPercentage := 0.0
	if originalSize > 0 {
		savingsPercentage = (float64(savingsBytes) / float64(originalSize)) * 100
	}

	return &models.StorageStats{
		TotalUsed:         totalUsed,
		OriginalSize:      originalSize,
		SavingsBytes:      savingsBytes,
		SavingsPercentage: savingsPercentage,
		Quota:             user.StorageQuota,
	}, nil
}