package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Morningstarl2504/Balkanid_repo/internal/database"
	"github.com/Morningstarl2504/Balkanid_repo/internal/models"
	"github.com/Morningstarl2504/Balkanid_repo/internal/services"
	"github.com/Morningstarl2504/Balkanid_repo/internal/utils"
)

type AdminHandler struct {
	fileService    *services.FileService
	storageService *services.StorageService
}

func NewAdminHandler(fileService *services.FileService, storageService *services.StorageService) *AdminHandler {
	return &AdminHandler{
		fileService:    fileService,
		storageService: storageService,
	}
}

func (h *AdminHandler) GetAllFiles(c *gin.Context) {
	var filters models.SearchFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Set default pagination
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 20
	}

	query := database.DB.Model(&models.File{}).
		Preload("Content").
		Preload("Owner").
		Preload("Folder")

	// Apply filters (reuse from file service)
	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	var files []models.File
	if err := query.Order("uploaded_at DESC").Find(&files).Error; err != nil {
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
		TotalUsers        int64 `json:"total_users"`
		TotalFiles        int64 `json:"total_files"`
		TotalStorage      int64 `json:"total_storage"`
		TotalDownloads    int64 `json:"total_downloads"`
		UniqueFiles       int64 `json:"unique_files"`
		DeduplicationSaved int64 `json:"deduplication_saved"`
	}

	// Get total users
	database.DB.Model(&models.User{}).Count(&stats.TotalUsers)

	// Get total files
	database.DB.Model(&models.File{}).Where("deleted_at IS NULL").Count(&stats.TotalFiles)

	// Get total storage (deduplicated)
	database.DB.Model(&models.FileContent{}).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalStorage)

	// Get total downloads
	database.DB.Model(&models.File{}).Select("COALESCE(SUM(download_count), 0)").Scan(&stats.TotalDownloads)

	// Get unique files (file contents)
	database.DB.Model(&models.FileContent{}).Count(&stats.UniqueFiles)

	// Calculate deduplication savings
	var totalOriginalSize int64
	database.DB.Raw(`
		SELECT COALESCE(SUM(fc.file_size * fc.reference_count), 0)
		FROM file_contents fc
	`).Scan(&totalOriginalSize)
	
	stats.DeduplicationSaved = totalOriginalSize - stats.TotalStorage

	utils.SuccessResponse(c, "System stats retrieved successfully", stats)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Select("id, username, email, is_admin, storage_quota, created_at").
		Find(&users).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Users retrieved successfully", users)
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	page := 1
	limit := 50
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var logs []models.AuditLog
	var total int64

	query := database.DB.Model(&models.AuditLog{}).Preload("User")
	
	// Count total
	query.Count(&total)
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	utils.SuccessResponse(c, "Audit logs retrieved successfully", response)
}