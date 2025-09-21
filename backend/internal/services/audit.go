// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/services/audit.go
package services

import (
	"filevault-backend/internal/database"
	"filevault-backend/internal/models"
	"github.com/gin-gonic/gin"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

func (s *AuditService) Log(c *gin.Context, action string, resource string, resourceID *uint, details string) {
	userID, exists := c.Get("userID")
	if !exists {
		return
	}
	userIDUint, _ := userID.(uint)

	log := models.AuditLog{
		UserID:     &userIDUint,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Details:    details,
	}

	database.DB.Create(&log)
}