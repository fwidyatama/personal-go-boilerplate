package handler

import (
	"net/http"
	"strconv"

	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/formatter"
	"go-boilerplate/internal/types"

	"go-boilerplate/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	rt "github.com/oapi-codegen/runtime/types"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userUseCase domain.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase domain.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req types.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    string(req.Email),
		Password: req.Password,
	}

	if err := h.userUseCase.Create(c.Request.Context(), user); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to create user: %v", err)

		if err.Error() == "user with email "+string(req.Email)+" already exists" {
			formatter.Error(c, http.StatusConflict, err.Error(), "Conflict")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to create user", "Internal Server Error")
		return
	}

	userResponse := ToUserDetailResponse(user)
	formatter.Respond(c, http.StatusCreated, true, "User created successfully", userResponse, "", nil)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Invalid user ID: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid user ID", "Bad Request")
		return
	}

	user, err := h.userUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to get user: %v", err)

		if err.Error() == "user not found: "+id.String() {
			formatter.Error(c, http.StatusNotFound, err.Error(), "Not Found")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to get user", "Internal Server Error")
		return
	}

	userResponse := ToUserDetailResponse(user)
	formatter.Success(c, userResponse, "User retrieved successfully")
}

func (h *UserHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.userUseCase.List(c.Request.Context(), limit, offset)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to list users: %v", err)
		formatter.Error(c, http.StatusInternalServerError, "Failed to list users", "Internal Server Error")
		return
	}

	userListResponse := ToUserListResponse(users)
	formatter.Success(c, userListResponse, "Users retrieved successfully")
}

func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Invalid user ID: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid user ID", "Bad Request")
		return
	}

	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to bind request: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid request body", "Bad Request")
		return
	}

	user := &domain.User{
		ID:       id,
		Name:     req.Name,
		Email:    string(req.Email),
		Password: req.Password,
	}

	if err := h.userUseCase.Update(c.Request.Context(), user); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to update user: %v", err)

		if err.Error() == "user not found: "+id.String() {
			formatter.Error(c, http.StatusNotFound, err.Error(), "Not Found")
			return
		} else if err.Error() == "user with email "+string(req.Email)+" already exists" {
			formatter.Error(c, http.StatusConflict, err.Error(), "Conflict")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to update user", "Internal Server Error")
		return
	}

	userResponse := ToUserDetailResponse(user)
	formatter.Success(c, userResponse, "User updated successfully")
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Invalid user ID: %v", err)
		formatter.Error(c, http.StatusBadRequest, "Invalid user ID", "Bad Request")
		return
	}

	if err := h.userUseCase.Delete(c.Request.Context(), id); err != nil {
		logger.WithRequestIDLogger(c.Request.Context()).Errorf("Failed to delete user: %v", err)

		if err.Error() == "user not found: "+id.String() {
			formatter.Error(c, http.StatusNotFound, err.Error(), "Not Found")
			return
		}

		formatter.Error(c, http.StatusInternalServerError, "Failed to delete user", "Internal Server Error")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		formatter.Error(c, http.StatusUnauthorized, "Unauthorized", "Unauthorized")
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		formatter.Error(c, http.StatusUnauthorized, "Invalid user ID", "Unauthorized")
		return
	}
	user, err := h.userUseCase.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			formatter.Error(c, http.StatusNotFound, "User not found", "Not Found")
			return
		}
		formatter.Error(c, http.StatusInternalServerError, "Internal server error", "Internal Server Error")
		return
	}
	// Map domain.User to UserResponse before returning
	userResp := userToResponse(user)
	formatter.Success(c, userResp, "User profile retrieved successfully")
}

// Mapping from domain.User to types.UserResponse
func ToUserDetailResponse(user *domain.User) types.UserResponse {
	if user == nil {
		return types.UserResponse{}
	}
	return types.UserResponse{
		Id:        &user.ID,
		Name:      &user.Name,
		Email:     (*rt.Email)(&user.Email),
		IsActive:  &user.IsActive,
		CreatedAt: &user.CreatedAt,
		UpdatedAt: &user.UpdatedAt,
	}
}

// Mapping from []*domain.User to types.UserListResponse
func ToUserListResponse(users []*domain.User) types.UserListResponse {
	var userResponses []types.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, ToUserDetailResponse(user))
	}
	return types.UserListResponse{
		Users: &userResponses,
	}
}

// userToResponse maps a domain.User to types.UserResponse
func userToResponse(u *domain.User) types.UserResponse {
	idStr := u.ID.String()
	return types.UserResponse{
		Id:        ptrUUID(idStr),
		Name:      ptrString(u.Name),
		Email:     ptrEmail(u.Email),
		IsActive:  &u.IsActive,
		CreatedAt: &u.CreatedAt,
		UpdatedAt: &u.UpdatedAt,
	}
}

func ptrString(s string) *string  { return &s }
func ptrUUID(s string) *rt.UUID   { v := rt.UUID(uuid.MustParse(s)); return &v }
func ptrEmail(e string) *rt.Email { v := rt.Email(e); return &v }
