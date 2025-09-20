package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"filevault-backend/internal/models"
	"filevault-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	fmt.Printf("=== UPLOAD REQUEST RECEIVED ===\n")
	
	// Get user ID from JWT middleware
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Printf("ERROR: User not authenticated\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	fmt.Printf("User ID: %v\n", userID)

	// Parse the multipart form
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Printf("ERROR: Failed to parse form: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data: " + err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		fmt.Printf("ERROR: No files in form\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}
	fmt.Printf("Found %d files to upload\n", len(files))

	var uploadedFiles []map[string]interface{}

	for i, fileHeader := range files {
		fmt.Printf("Processing file %d: %s (size: %d)\n", i+1, fileHeader.Filename, fileHeader.Size)

		// Basic file size validation (10MB limit)
		maxSize := int64(10 * 1024 * 1024) // 10MB
		if fileHeader.Size > maxSize {
			fmt.Printf("ERROR: File too large\n")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File %s is too large (max 10MB)", fileHeader.Filename),
			})
			return
		}

		// Generate unique filename
		ext := filepath.Ext(fileHeader.Filename)
		uniqueFilename := uuid.New().String() + ext
		fmt.Printf("Generated unique filename: %s\n", uniqueFilename)

		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Printf("ERROR: Failed to open file: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
			return
		}
		defer file.Close()

		// Save file using storage service
		fmt.Printf("Saving file to storage...\n")
		err = h.storageService.SaveFile(uniqueFilename, file)
		if err != nil {
			fmt.Printf("ERROR: Storage save failed: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
			return
		}
		fmt.Printf("File saved to storage successfully\n")

		// Create file record in database
		fileRecord := &models.File{
			UserID:           userID.(uint),
			Filename:         uniqueFilename,
			OriginalFilename: fileHeader.Filename,
			MimeType:         fileHeader.Header.Get("Content-Type"),
			Size:             fileHeader.Size,
			IsPublic:         false,
		}

		fmt.Printf("Saving file metadata to database...\n")
		err = h.fileService.Create(fileRecord)
		if err != nil {
			fmt.Printf("ERROR: Database save failed: %v\n", err)
			// Clean up the stored file
			h.storageService.Delete(uniqueFilename)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata: " + err.Error()})
			return
		}
		fmt.Printf("File metadata saved successfully\n")

		uploadedFiles = append(uploadedFiles, map[string]interface{}{
			"id":                fileRecord.ID,
			"filename":          uniqueFilename,
			"original_filename": fileHeader.Filename,
			"size":              fileHeader.Size,
			"mime_type":         fileHeader.Header.Get("Content-Type"),
		})
	}

	fmt.Printf("=== UPLOAD SUCCESS: %d files ===\n", len(uploadedFiles))
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully uploaded %d file(s)", len(uploadedFiles)),
		"data": gin.H{
			"files": uploadedFiles,
		},
	})
}

func (h *FileHandler) GetUserFiles(c *gin.Context) {
	fmt.Printf("=== GET USER FILES REQUEST ===\n")
	
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Printf("ERROR: User not authenticated\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	fmt.Printf("Getting files for user ID: %v\n", userID)

	files, err := h.fileService.GetByUserID(userID.(uint))
	if err != nil {
		fmt.Printf("ERROR: Failed to get files: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve files: " + err.Error()})
		return
	}

	fmt.Printf("Found %d files\n", len(files))

	// Convert to the format expected by your frontend
	var fileResponses []map[string]interface{}
	for _, file := range files {
		fileResponse := map[string]interface{}{
			"id":                file.ID,
			"filename":          file.Filename,
			"original_filename": file.OriginalFilename,
			"content": map[string]interface{}{
				"file_size": file.Size,
				"mime_type": file.MimeType,
			},
			"owner": map[string]interface{}{
				"username": file.User.Username,
			},
			"uploaded_at":    file.CreatedAt,
			"download_count": file.DownloadCount,
			"is_public":      file.IsPublic,
		}
		fileResponses = append(fileResponses, fileResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"files": fileResponses,
		},
	})
	fmt.Printf("=== RETURNED %d FILES ===\n", len(fileResponses))
}

func (h *FileHandler) GetFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"file": file,
		},
	})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Get file data from storage
	fileData, err := h.storageService.Get(file.Filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File data not found"})
		return
	}

	// Increment download count
	h.fileService.IncrementDownloadCount(file.ID)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.OriginalFilename))
	c.Header("Content-Type", file.MimeType)
	c.Data(http.StatusOK, file.MimeType, fileData)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	userID, _ := c.Get("userID")

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check if user owns the file
	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete file from storage
	err = h.storageService.Delete(file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from storage"})
		return
	}

	// Delete file record from database
	err = h.fileService.Delete(uint(fileID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func (h *FileHandler) ShareFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	userID, _ := c.Get("userID")

	file, err := h.fileService.GetByID(uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check if user owns the file
	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Toggle public status
	file.IsPublic = !file.IsPublic
	err = h.fileService.Update(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
		return
	}

	status := "private"
	if file.IsPublic {
		status = "public"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("File is now %s", status),
		"data": gin.H{
			"file": file,
		},
	})
}

func (h *FileHandler) GetPublicFiles(c *gin.Context) {
	files, err := h.fileService.GetPublicFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve public files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"files": files,
		},
	})
}

// Placeholder methods for folder operations
func (h *FileHandler) CreateFolder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Folder created successfully"})
}

func (h *FileHandler) GetUserFolders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"folders": []interface{}{},
		},
	})
}

func (h *FileHandler) GetStorageStats(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	// Get user's files for stats
	files, err := h.fileService.GetByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve storage stats"})
		return
	}

	totalSize := int64(0)
	for _, file := range files {
		totalSize += file.Size
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_files": len(files),
			"total_size":  totalSize,
		},
	})
}