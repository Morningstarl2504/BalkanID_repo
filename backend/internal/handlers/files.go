// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/handlers/files.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"filevault-backend/internal/database"
	"filevault-backend/internal/models"
	"filevault-backend/internal/services"
	"filevault-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileHandler struct {
	fileService    *services.FileService
	storageService *services.StorageService
	auditService   *services.AuditService
}

func NewFileHandler(fileService *services.FileService, storageService *services.StorageService, auditService *services.AuditService) *FileHandler {
	return &FileHandler{
		fileService:    fileService,
		storageService: storageService,
		auditService:   auditService,
	}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to parse form data: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No files provided")
		return
	}

	var uploadedFiles []map[string]interface{}

	for _, fileHeader := range files {
		if err := utils.ValidateFileType(fileHeader); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Validation failed for %s: %s", fileHeader.Filename, err.Error()))
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open uploaded file")
			return
		}

		hash, err := utils.CalculateSHA256Reader(file) // Corrected function name
		file.Close()

		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate file hash")
			return
		}

		var content models.FileContent
		err = database.DB.Where("sha256_hash = ?", hash).First(&content).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error checking for file content")
			return
		}

		if err == nil {
			database.DB.Model(&content).Update("reference_count", gorm.Expr("reference_count + 1"))
		} else {
			fileForSave, err := fileHeader.Open()
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to re-open uploaded file for saving")
				return
			}

			err = h.storageService.SaveFile(hash, fileForSave)
			fileForSave.Close()

			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file to storage")
				return
			}

			content = models.FileContent{
				SHA256Hash: hash,
				FileSize:   fileHeader.Size,
				MimeType:   fileHeader.Header.Get("Content-Type"),
			}
			if err := database.DB.Create(&content).Error; err != nil {
				h.storageService.Delete(hash)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file metadata")
				return
			}
		}

		fileRecord := &models.File{
			UserID:           userID.(uint),
			FileContentID:    content.ID,
			OriginalFilename: fileHeader.Filename,
		}

		if err := h.fileService.Create(fileRecord); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create file record")
			return
		}

		h.auditService.Log(c, "UPLOAD", "FILE", &fileRecord.ID, fmt.Sprintf("User uploaded file '%s'", fileHeader.Filename))

		uploadedFiles = append(uploadedFiles, map[string]interface{}{
			"id":                fileRecord.ID,
			"original_filename": fileHeader.Filename,
			"size":              content.FileSize,
			"mime_type":         content.MimeType,
		})
	}

	utils.SuccessResponse(c, fmt.Sprintf("Successfully uploaded %d file(s)", len(uploadedFiles)), gin.H{"files": uploadedFiles})
}

func (h *FileHandler) GetUserFiles(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filters models.SearchFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	files, err := h.fileService.GetByUserID(userID.(uint), &filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve files: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "Files retrieved successfully", gin.H{"files": files})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File not found")
		return
	}

	fileData, err := h.storageService.Get(file.Content.SHA256Hash)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File data not found in storage")
		return
	}

	h.fileService.IncrementDownloadCount(file.ID)
	h.auditService.Log(c, "DOWNLOAD", "FILE", &file.ID, fmt.Sprintf("User downloaded file '%s'", file.OriginalFilename))

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.OriginalFilename))
	c.Header("Content-Type", file.Content.MimeType)
	c.Data(http.StatusOK, file.Content.MimeType, fileData)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	userID, _ := c.Get("userID")

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File not found")
		return
	}

	if file.UserID != userID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied: you do not own this file")
		return
	}

	err = h.fileService.DeleteFileAndContent(uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		return
	}

	h.auditService.Log(c, "DELETE", "FILE", &file.ID, fmt.Sprintf("User deleted file '%s'", file.OriginalFilename))
	utils.SuccessResponse(c, "File deleted successfully", nil)
}

func (h *FileHandler) GetStorageStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	stats, err := h.fileService.GetStorageStats(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve storage stats: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "Storage stats retrieved successfully", stats)
}
func (h *FileHandler) ShareFile(c *gin.Context) {
	userID, _ := c.Get("userID")
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File not found")
		return
	}

	if file.UserID != userID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied: you do not own this file")
		return
	}

	// Toggle the public status
	file.IsPublic = !file.IsPublic
	if err := database.DB.Save(file).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update file status")
		return
	}

	h.auditService.Log(c, "SHARE", "FILE", &file.ID, fmt.Sprintf("User toggled public access for '%s' to %v", file.OriginalFilename, file.IsPublic))
	utils.SuccessResponse(c, "File public status updated successfully", file)
}

func (h *FileHandler) PublicDownload(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil || !file.IsPublic {
		utils.ErrorResponse(c, http.StatusNotFound, "File not found or is not public")
		return
	}

	fileData, err := h.storageService.Get(file.Content.SHA256Hash)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "File data not found in storage")
		return
	}

	h.fileService.IncrementDownloadCount(file.ID)
	// Note: We don't log the audit event with a user here, as the download is anonymous.
	
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.OriginalFilename))
	c.Header("Content-Type", file.Content.MimeType)
	c.Data(http.StatusOK, file.Content.MimeType, fileData)
}