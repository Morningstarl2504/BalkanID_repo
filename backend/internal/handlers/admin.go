// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/handlers/admin.go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"filevault-backend/internal/database"
	"filevault-backend/internal/models"
	"filevault-backend/internal/services"
	"filevault-backend/internal/utils"
)

type AdminHandler struct {
	fileService    *services.FileService
	storageService *services.StorageService
	auditService   *services.AuditService
}

func NewAdminHandler(fileService *services.FileService, storageService *services.StorageService, auditService *services.AuditService) *AdminHandler {
	return &AdminHandler{
		fileService:    fileService,
		storageService: storageService,
		auditService:   auditService,
	}
}

func (h *AdminHandler) GetAllFiles(c *gin.Context) {
	var filters models.SearchFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 20
	}

	query := database.DB.Model(&models.File{}).
		Preload("Content").
		Preload("User")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	var files []models.File
	if err := query.Order("created_at DESC").Find(&files).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"files":       files,
		"total":       total,
		"page":        filters.Page,
		"limit":       filters.Limit,
		"total_pages": (total + int64(filters.Limit) - 1) / int64(filters.Limit),
	}

	utils.SuccessResponse(c, "All files retrieved successfully", response)
}

func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	var stats struct {
		TotalUsers         int64   `json:"total_users"`
		TotalFiles         int64   `json:"total_files"`
		TotalStorageUsed   int64   `json:"total_storage_used"`
		OriginalTotalSize  int64   `json:"original_total_size"`
		DeduplicationSaved int64   `json:"deduplication_saved"`
		SavingsPercentage  float64 `json:"savings_percentage"`
	}

	database.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	database.DB.Model(&models.File{}).Count(&stats.TotalFiles)
	database.DB.Model(&models.FileContent{}).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalStorageUsed)
	database.DB.Model(&models.File{}).Joins("JOIN file_contents ON file_contents.id = files.file_content_id").
		Select("COALESCE(SUM(file_contents.file_size), 0)").Scan(&stats.OriginalTotalSize)

	stats.DeduplicationSaved = stats.OriginalTotalSize - stats.TotalStorageUsed
	if stats.OriginalTotalSize > 0 {
		stats.SavingsPercentage = (float64(stats.DeduplicationSaved) / float64(stats.OriginalTotalSize)) * 100
	}

	utils.SuccessResponse(c, "System stats retrieved successfully", stats)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, "Users retrieved successfully", users)
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	if err := database.DB.Preload("User").Order("created_at DESC").Find(&logs).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, "Audit logs retrieved successfully", logs)
}