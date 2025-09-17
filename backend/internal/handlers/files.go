package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Morningstarl2504/BalkanID_repo/backend/models"
	"github.com/Morningstarl2504/BalkanID_repo/backend/services"
	"github.com/Morningstarl2504/BalkanID_repo/backend/utils"
)

type FileHandler struct {
	fileService    *services.FileService
	storageService *services.StorageService
}

func NewFileHandler(fileService *services.FileService, storageService *services.StorageService) *FileHandler {
	return &FileHandler{
		fileService:    fileService,
		storageService: storageService,
	}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get files from form
	form := c.Request.MultipartForm
	files := form.File["files"]
	if len(files) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No files provided")
		return
	}

	// Parse optional fields
	tags := strings.Split(c.PostForm("tags"), ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	var folderID *uint
	if folderIDStr := c.PostForm("folder_id"); folderIDStr != "" {
		if id, err := strconv.ParseUint(folderIDStr, 10, 32); err == nil {
			folderIDUint := uint(id)
			folderID = &folderIDUint
		}
	}

	var uploadedFiles []models.File
	var errors []string

	// Process each file
	for _, fileHeader := range files {
		file, err := h.fileService.UploadFile(userID.(uint), fileHeader, tags, folderID)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		uploadedFiles = append(uploadedFiles, *file)
	}

	response := map[string]interface{}{
		"uploaded_files": uploadedFiles,
		"errors":        errors,
		"success_count": len(uploadedFiles),
		"error_count":   len(errors),
	}

	if len(uploadedFiles) > 0 {
		utils.SuccessResponse(c, "Files uploaded successfully", response)
	} else {
		utils.ErrorResponse(c, http.StatusBadRequest, "All file uploads failed")
	}
}

func (h *FileHandler) GetUserFiles(c *gin.Context) {
	userID, _ := c.Get("userID")

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

	files, total, err := h.fileService.GetUserFiles(userID.(uint), &filters)
	if err != nil {
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

	utils.SuccessResponse(c, "Files retrieved successfully", response)
}

func (h *FileHandler) GetPublicFiles(c *gin.Context) {
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

	files, total, err := h.fileService.GetPublicFiles(&filters)
	if err != nil {
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

	utils.SuccessResponse(c, "Public files retrieved successfully", response)
}

func (h *FileHandler) GetFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	userID, _ := c.Get("userID")

	file, err := h.fileService.GetFile(uint(fileID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, "File retrieved successfully", file)
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	userID, _ := c.Get("userID")

	filePath, originalFilename, err := h.fileService.DownloadFile(uint(fileID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+originalFilename+"\"")
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	userID, _ := c.Get("userID")

	if err := h.fileService.DeleteFile(uint(fileID), userID.(uint)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	utils.SuccessResponse(c, "File deleted successfully", nil)
}

func (h *FileHandler) ShareFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	var req struct {
		IsPublic      bool   `json:"is_public"`
		SharedUserIDs []uint `json:"shared_user_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	userID, _ := c.Get("userID")

	if err := h.fileService.ShareFile(uint(fileID), userID.(uint), req.IsPublic, req.SharedUserIDs); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	utils.SuccessResponse(c, "File sharing updated successfully", nil)
}

func (h *FileHandler) GetStorageStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	stats, err := h.storageService.GetUserStorageStats(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Storage stats retrieved successfully", stats)
}

func (h *FileHandler) CreateFolder(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Name     string `json:"name" binding:"required"`
		ParentID *uint  `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	folder, err := h.fileService.CreateFolder(userID.(uint), req.Name, req.ParentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Folder created successfully", folder)
}

func (h *FileHandler) GetUserFolders(c *gin.Context) {
	userID, _ := c.Get("userID")

	folders, err := h.fileService.GetUserFolders(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Folders retrieved successfully", folders)
}
