package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Morningstarl2504/BalkanID_repo/internal/models"
	"github.com/Morningstarl2504/BalkanID_repo/internal/services"
	"github.com/Morningstarl2504/BalkanID_repo/internal/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	utils.SuccessResponse(c, "User registered successfully", response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.authService.Login(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	utils.SuccessResponse(c, "Login successful", response)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	isAdmin, _ := c.Get("isAdmin")

	user := map[string]interface{}{
		"id":       userID,
		"username": username,
		"is_admin": isAdmin,
	}

	utils.SuccessResponse(c, "Profile retrieved successfully", user)
}