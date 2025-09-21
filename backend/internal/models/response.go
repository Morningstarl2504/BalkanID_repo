// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/models/response.go
package models

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type StorageStats struct {
	TotalUsed         int64   `json:"total_storage_used"`
	OriginalSize      int64   `json:"original_storage_usage"`
	SavingsBytes      int64   `json:"storage_savings_bytes"`
	SavingsPercentage float64 `json:"storage_savings_percentage"`
	Quota             int64   `json:"user_quota"`
}

type SearchFilters struct {
	Filename     string   `form:"filename"`
	MimeType     string   `form:"mime_type"`
	MinSize      int64    `form:"min_size"`
	MaxSize      int64    `form:"max_size"`
	StartDate    string   `form:"start_date"`
	EndDate      string   `form:"end_date"`
	Tags         []string `form:"tags"`
	UploaderName string   `form:"uploader_name"`
	Page         int      `form:"page" binding:"min=1"`
	Limit        int      `form:"limit" binding:"min=1,max=100"`
}