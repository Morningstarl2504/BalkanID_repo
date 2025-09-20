package services

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"filevault-backend/internal/database"
	"filevault-backend/internal/models"
)

type StorageService struct {
	UploadPath string
}

func NewStorageService(uploadPath string) *StorageService {
	return &StorageService{
		UploadPath: uploadPath,
	}
}

// SaveFile saves a multipart file to disk
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

// Get retrieves file content as bytes
func (s *StorageService) Get(filename string) ([]byte, error) {
	filePath := filepath.Join(s.UploadPath, filename)
	return os.ReadFile(filePath)
}

// Delete removes a file from storage
func (s *StorageService) Delete(filename string) error {
	filePath := filepath.Join(s.UploadPath, filename)
	return os.Remove(filePath)
}

// Exists checks if a file exists in storage
func (s *StorageService) Exists(filename string) bool {
	filePath := filepath.Join(s.UploadPath, filename)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetPath returns the full file system path for a filename
func (s *StorageService) GetPath(filename string) string {
	return filepath.Join(s.UploadPath, filename)
}

// FileService handles file operations
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

// Create saves a new file record to the database
func (s *FileService) Create(file *models.File) error {
	return database.DB.Create(file).Error
}

// GetByUserID retrieves all files for a specific user
func (s *FileService) GetByUserID(userID uint) ([]*models.File, error) {
	var files []*models.File
	err := database.DB.Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Find(&files).Error
	return files, err
}

// GetByID retrieves a single file by its ID
func (s *FileService) GetByID(id uint) (*models.File, error) {
	var file models.File
	err := database.DB.Preload("User").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Update saves changes to a file record
func (s *FileService) Update(file *models.File) error {
	return database.DB.Save(file).Error
}

// Delete removes a file record from the database
func (s *FileService) Delete(id uint) error {
	return database.DB.Delete(&models.File{}, id).Error
}

// GetPublicFiles retrieves all public files
func (s *FileService) GetPublicFiles() ([]*models.File, error) {
	var files []*models.File
	err := database.DB.Preload("User").Where("is_public = ?", true).Order("created_at DESC").Find(&files).Error
	return files, err
}

// IncrementDownloadCount increases the download count for a file
func (s *FileService) IncrementDownloadCount(id uint) error {
	return database.DB.Model(&models.File{}).Where("id = ?", id).
		Update("download_count", database.DB.Raw("download_count + 1")).Error
}

// GetMaxFileSize returns the maximum allowed file size
func (s *FileService) GetMaxFileSize() int64 {
	return s.maxFileSize
}