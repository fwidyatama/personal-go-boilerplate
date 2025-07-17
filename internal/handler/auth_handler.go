package handler

import (
	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/formatter"
	"go-boilerplate/internal/types"
	"go-boilerplate/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authUseCase domain.AuthUseCase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authUseCase domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	user := &domain.User{
		ID:       uuid.New(),
		Name:     req.Name,
		Email:    string(req.Email),
		Password: req.Password,
		IsActive: true,
	}

	if err := h.authUseCase.Register(c.Request.Context(), user); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to register user: %v", err)

		if err.Error() == "user with email "+string(req.Email)+" already exists" {
			formatter.Error(c, http.StatusConflict, err.Error(), "Conflict")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to register user", "Internal Server Error")
		return
	}

	authResponse, err := h.authUseCase.Login(c.Request.Context(), string(req.Email), req.Password)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to generate auth response: %v", err)
		formatter.Error(c, http.StatusInternalServerError, "Failed to complete registration", "Internal Server Error")
		return
	}

	typesAuthResponse := ToRegisterAuthResponse(user, authResponse)

	formatter.Respond(c, http.StatusCreated, true, "User registered successfully", typesAuthResponse, "", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	authResponse, err := h.authUseCase.Login(c.Request.Context(), string(req.Email), req.Password)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to login user: %v", err)

		if err.Error() == "invalid credentials" {
			formatter.Error(c, http.StatusUnauthorized, "Invalid credentials", "Unauthorized")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to login", "Internal Server Error")
		return
	}

	typesAuthResponse := ToAuthResponse(authResponse)

	formatter.Success(c, typesAuthResponse, "Login successful")
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req types.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	authResponse, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to refresh token: %v", err)

		if err.Error() == "invalid refresh token" {
			formatter.Error(c, http.StatusUnauthorized, "Invalid refresh token", "Unauthorized")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to refresh token", "Internal Server Error")
		return
	}

	typesAuthResponse := ToAuthResponse(authResponse)

	formatter.Success(c, typesAuthResponse, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		formatter.Error(c, http.StatusUnauthorized, "User not authenticated", "Unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		formatter.Error(c, http.StatusUnauthorized, "Invalid user ID", "Unauthorized")
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), userUUID); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to logout user: %v", err)
		formatter.Error(c, http.StatusInternalServerError, "Failed to logout", "Internal Server Error")
		return
	}

	formatter.Success(c, nil, "Logout successful")
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req types.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	err := h.authUseCase.ForgotPassword(c.Request.Context(), string(req.Email))
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to process forgot password: %v", err)
		if err.Error() == "user not found" {
			formatter.Error(c, http.StatusNotFound, err.Error(), "Not Found")
			return
		}
		formatter.Error(c, http.StatusInternalServerError, "Failed to process forgot password", "Internal Server Error")
		return
	}

	formatter.Success(c, nil, "Password reset email sent (if user exists)")
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req types.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	err := h.authUseCase.ResetPassword(c.Request.Context(), string(req.Email), req.Token, req.NewPassword)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to reset password: %v", err)
		if err.Error() == "invalid or expired token" {
			formatter.Error(c, http.StatusNotFound, err.Error(), "Not Found")
			return
		}
		formatter.Error(c, http.StatusInternalServerError, "Failed to reset password", "Internal Server Error")
		return
	}

	formatter.Success(c, nil, "Password reset successful")
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req types.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		formatter.Error(c, http.StatusUnauthorized, "User not authenticated", "Unauthorized")
		return
	}
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		formatter.Error(c, http.StatusUnauthorized, "Invalid user ID", "Unauthorized")
		return
	}

	err := h.authUseCase.ChangePassword(c.Request.Context(), userUUID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to change password: %v", err)
		if err.Error() == "current password incorrect" {
			formatter.Error(c, http.StatusForbidden, err.Error(), "Forbidden")
			return
		}
		formatter.Error(c, http.StatusInternalServerError, "Failed to change password", "Internal Server Error")
		return
	}

	formatter.Success(c, nil, "Password changed successfully")
}

func intPtr(i int) *int { return &i }

func ToAuthUserDetailResponse(user *domain.User) *types.UserResponse {
	return &types.UserResponse{
		Id:       &user.ID,
		Name:     &user.Name,
		Email:    (*openapi_types.Email)(&user.Email),
		IsActive: &user.IsActive,
	}
}

// Mapping dari domain.Auth ke types.AuthResponse
func ToRegisterAuthResponse(user *domain.User, auth *domain.Auth) *types.AuthResponse {
	if user == nil || auth == nil {
		return nil
	}
	return &types.AuthResponse{
		AccessToken: &auth.AccessToken,
		TokenType:   &auth.TokenType,
		ExpiresIn:   intPtr(int(auth.ExpiresIn)),
		User:        ToAuthUserDetailResponse(user),
	}
}

func ToAuthResponse(auth *domain.Auth) *types.AuthResponse {
	if auth == nil {
		return nil
	}
	return &types.AuthResponse{
		AccessToken:  &auth.AccessToken,
		RefreshToken: &auth.RefreshToken,
		TokenType:    &auth.TokenType,
		ExpiresIn:    intPtr(int(auth.ExpiresIn)),
		User:         ToAuthUserDetailResponse(auth.User),
	}
}
