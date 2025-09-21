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

	// Join with file_contents for filtering on content properties
	query = query.Joins("JOIN file_contents ON file_contents.id = files.file_content_id")

	if filters.Filename != "" {
		query = query.Where("files.original_filename LIKE ?", "%"+filters.Filename+"%")
	}
	if filters.MimeType != "" {
		// Handle cases like "image" which should match "image/png", "image/jpeg", etc.
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
			// Go to the end of the selected day
			query = query.Where("files.created_at <= ?", t.Add(24*time.Hour-time.Nanosecond))
		}
	}

	// Pagination should be applied in the handler, not the service, for better separation of concerns.
	// But for simplicity in this context, we assume the handler does it.
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

func (s *FileService) Delete(id uint) error {
	return database.DB.Delete(&models.File{}, id).Error
}

func (s *FileService) DeleteFileAndContent(fileID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var file models.File
		if err := tx.Preload("Content").First(&file, fileID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&models.File{}, fileID).Error; err != nil {
			return err
		}

		var content models.FileContent
		if err := tx.First(&content, file.FileContentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if content.ReferenceCount > 1 {
			return tx.Model(&content).Update("reference_count", gorm.Expr("reference_count - 1")).Error
		}
		
		if err := s.storageService.Delete(content.SHA256Hash); err != nil {
			// Log this error but don't fail the transaction
		}

		return tx.Delete(&content).Error
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