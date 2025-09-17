package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Morningstarl2504/BalkanID_repo/internal/database"
	"github.com/Morningstarl2504/BalkanID_repo/internal/models"
	"github.com/Morningstarl2504/BalkanID_repo/internal/utils"
)

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

func (s *FileService) UploadFile(userID uint, fileHeader *multipart.FileHeader, tags []string, folderID *uint) (*models.File, error) {
	// Check file size
	if fileHeader.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.maxFileSize)
	}

	// Check storage quota
	if err := s.storageService.CheckStorageQuota(userID, fileHeader.Size); err != nil {
		return nil, err
	}

	// Open uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Create temporary file
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "upload_*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded file to temp file
	_, err = io.Copy(tempFile, src)
	if err != nil {
		return nil, err
	}

	// Get MIME type from file header
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Validate file type
	if err := utils.ValidateFileType(tempFile.Name(), mimeType); err != nil {
		return nil, fmt.Errorf("file validation failed: %v", err)
	}

	// Calculate file hash
	contentHash, err := utils.CalculateSHA256(tempFile.Name())
	if err != nil {
		return nil, err
	}

	// Find or create file content (handles deduplication)
	fileContent, err := s.storageService.FindOrCreateFileContent(
		tempFile.Name(), contentHash, mimeType, fileHeader.Size)
	if err != nil {
		return nil, err
	}

	// Generate unique filename
	originalFilename := fileHeader.Filename
	filename := fmt.Sprintf("%s_%s", uuid.New().String(), originalFilename)

	// Create file record
	file := &models.File{
		Filename:         filename,
		OriginalFilename: originalFilename,
		ContentID:        fileContent.ID,
		OwnerID:          userID,
		FolderID:         folderID,
		IsPublic:         false,
		Tags:             tags,
		DownloadCount:    0,
	}

	if err := database.DB.Create(file).Error; err != nil {
		// If file creation fails, decrement reference count
		s.storageService.DeleteFileContent(fileContent.ID)
		return nil, err
	}

	// Create audit log
	s.createAuditLog(userID, "upload", "file", file.ID, map[string]interface{}{
		"filename": originalFilename,
		"size":     fileHeader.Size,
		"hash":     contentHash,
	})

	// Load complete file with relations
	if err := database.DB.Preload("Content").Preload("Owner").First(file, file.ID).Error; err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) GetUserFiles(userID uint, filters *models.SearchFilters) ([]models.File, int64, error) {
	query := database.DB.Model(&models.File{}).
		Preload("Content").
		Preload("Owner").
		Preload("Folder").
		Where("owner_id = ?", userID)

	// Apply filters
	query = s.applyFilters(query, filters)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	var files []models.File
	if err := query.Order("uploaded_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (s *FileService) GetPublicFiles(filters *models.SearchFilters) ([]models.File, int64, error) {
	query := database.DB.Model(&models.File{}).
		Preload("Content").
		Preload("Owner").
		Preload("Folder").
		Where("is_public = ?", true)

	// Apply filters
	query = s.applyFilters(query, filters)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	var files []models.File
	if err := query.Order("uploaded_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (s *FileService) applyFilters(query *gorm.DB, filters *models.SearchFilters) *gorm.DB {
	if filters == nil {
		return query
	}

	// Join with content for filtering
	query = query.Joins("JOIN file_contents ON files.content_id = file_contents.id")

	// Filename filter
	if filters.Filename != "" {
		query = query.Where("files.original_filename ILIKE ?", "%"+filters.Filename+"%")
	}

	// MIME type filter
	if filters.MimeType != "" {
		query = query.Where("file_contents.mime_type = ?", filters.MimeType)
	}

	// Size filters
	if filters.MinSize > 0 {
		query = query.Where("file_contents.file_size >= ?", filters.MinSize)
	}
	if filters.MaxSize > 0 {
		query = query.Where("file_contents.file_size <= ?", filters.MaxSize)
	}

	// Date filters
	if filters.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			query = query.Where("files.uploaded_at >= ?", startDate)
		}
	}
	if filters.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			query = query.Where("files.uploaded_at <= ?", endDate.Add(24*time.Hour))
		}
	}

	// Tags filter
	if len(filters.Tags) > 0 {
		query = query.Where("files.tags && ?", filters.Tags)
	}

	// Uploader name filter
	if filters.UploaderName != "" {
		query = query.Joins("JOIN users ON files.owner_id = users.id").
			Where("users.username ILIKE ?", "%"+filters.UploaderName+"%")
	}

	return query
}

func (s *FileService) GetFile(fileID uint, userID uint) (*models.File, error) {
	var file models.File
	query := database.DB.Preload("Content").Preload("Owner").Preload("Folder")

	if err := query.First(&file, fileID).Error; err != nil {
		return nil, err
	}

	// Check access permissions
	if !s.canAccessFile(&file, userID) {
		return nil, errors.New("access denied")
	}

	return &file, nil
}

func (s *FileService) canAccessFile(file *models.File, userID uint) bool {
	// Owner can always access
	if file.OwnerID == userID {
		return true
	}

	// Public files can be accessed by anyone
	if file.IsPublic {
		return true
	}

	// Check if file is shared with user
	var share models.FileShare
	err := database.DB.Where("file_id = ? AND shared_with_user_id = ?", file.ID, userID).First(&share).Error
	return err == nil
}

func (s *FileService) DownloadFile(fileID uint, userID uint) (string, string, error) {
	file, err := s.GetFile(fileID, userID)
	if err != nil {
		return "", "", err
	}

	// Increment download count
	database.DB.Model(file).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1))

	// Create audit log
	s.createAuditLog(userID, "download", "file", file.ID, map[string]interface{}{
		"filename": file.OriginalFilename,
	})

	return file.Content.FilePath, file.OriginalFilename, nil
}

func (s *FileService) DeleteFile(fileID uint, userID uint) error {
	var file models.File
	if err := database.DB.First(&file, fileID).Error; err != nil {
		return err
	}

	// Only owner can delete
	if file.OwnerID != userID {
		return errors.New("only the owner can delete this file")
	}

	// Start transaction
	tx := database.DB.Begin()

	// Soft delete the file
	if err := tx.Delete(&file).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Handle file content reference counting
	if err := s.storageService.DeleteFileContent(file.ContentID); err != nil {
		tx.Rollback()
		return err
	}

	// Create audit log
	s.createAuditLog(userID, "delete", "file", file.ID, map[string]interface{}{
		"filename": file.OriginalFilename,
	})

	tx.Commit()
	return nil
}

func (s *FileService) ShareFile(fileID uint, ownerID uint, isPublic bool, sharedUserIDs []uint) error {
	var file models.File
	if err := database.DB.First(&file, fileID).Error; err != nil {
		return err
	}

	// Only owner can share
	if file.OwnerID != ownerID {
		return errors.New("only the owner can share this file")
	}

	// Update public status
	if err := database.DB.Model(&file).Update("is_public", isPublic).Error; err != nil {
		return err
	}

	// Handle specific user shares
	if len(sharedUserIDs) > 0 {
		// Remove existing shares
		database.DB.Where("file_id = ?", fileID).Delete(&models.FileShare{})

		// Create new shares
		for _, userID := range sharedUserIDs {
			share := models.FileShare{
				FileID:           fileID,
				SharedWithUserID: userID,
				SharedByUserID:   ownerID,
			}
			database.DB.Create(&share)
		}
	}

	// Create audit log
	s.createAuditLog(ownerID, "share", "file", file.ID, map[string]interface{}{
		"filename":        file.OriginalFilename,
		"is_public":       isPublic,
		"shared_with":     sharedUserIDs,
	})

	return nil
}

func (s *FileService) CreateFolder(userID uint, name string, parentID *uint) (*models.Folder, error) {
	folder := &models.Folder{
		Name:     name,
		OwnerID:  userID,
		ParentID: parentID,
		IsPublic: false,
	}

	if err := database.DB.Create(folder).Error; err != nil {
		return nil, err
	}

	// Create audit log
	s.createAuditLog(userID, "create", "folder", folder.ID, map[string]interface{}{
		"name": name,
	})

	return folder, nil
}

func (s *FileService) GetUserFolders(userID uint) ([]models.Folder, error) {
	var folders []models.Folder
	if err := database.DB.Where("owner_id = ?", userID).
		Preload("Owner").
		Order("name").
		Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

func (s *FileService) createAuditLog(userID uint, action, resourceType string, resourceID uint, details map[string]interface{}) {
	auditLog := models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}
	database.DB.Create(&auditLog)
}