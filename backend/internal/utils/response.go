package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Morningstarl2504/Balkanid_repo/internal/models"
)

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.Response{
		Success: false,
		Error:   message,
	})
}

func ValidationErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, models.Response{
		Success: false,
		Error:   err.Error(),
	})
}
