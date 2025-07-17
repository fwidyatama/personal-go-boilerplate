package usecase

import (
	"context"
	"fmt"

	"go-boilerplate/internal/auth"
	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/handler"
	"go-boilerplate/internal/types"

	"go-boilerplate/pkg/logger"

	"go-boilerplate/pkg/redis"

	"github.com/google/uuid"
)

// UserUseCase implements domain.UserUseCase
type UserUseCase struct {
	userRepo domain.UserRepository
	authSvc  auth.AuthService
	redis    redis.RedisClient
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo domain.UserRepository, authSvc auth.AuthService, redisClient redis.RedisClient) domain.UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		authSvc:  authSvc,
		redis:    redisClient,
	}
}

// Create creates a new user with business logic validation
func (uc *UserUseCase) Create(ctx context.Context, user *domain.User) error {
	if user.Email == "" || user.Password == "" || user.Name == "" {
		return fmt.Errorf("validation failed: missing required fields")
	}

	// Check if user with email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Hash password
	hashedPassword, err := uc.authSvc.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	// Create user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.WithRequestIDLogger(ctx).Infof("User created successfully: %s", user.ID)
	return nil
}

// GetByID retrieves a user by ID
func (uc *UserUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID: %s", id)
	}

	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// List retrieves a list of users with pagination
func (uc *UserUseCase) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	users, err := uc.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Update updates an existing user
func (uc *UserUseCase) Update(ctx context.Context, user *domain.User) error {
	if user.Email == "" || user.Name == "" {
		return fmt.Errorf("validation failed: missing required fields")
	}

	// Check if user exists
	existingUser, err := uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	// Check if email is being changed and if it's already taken
	if user.Email != existingUser.Email {
		userWithEmail, err := uc.userRepo.GetByEmail(ctx, user.Email)
		if err == nil && userWithEmail != nil && userWithEmail.ID != user.ID {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	// Hash password if it's being updated
	if user.Password != "" && user.Password != existingUser.Password {
		hashedPassword, err := uc.authSvc.HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	} else {
		user.Password = existingUser.Password
	}

	// Update user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.WithRequestIDLogger(ctx).Infof("User updated successfully: %s", user.ID)
	return nil
}

// Delete deletes a user by ID
func (uc *UserUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid user ID: %s", id)
	}

	// Check if user exists
	_, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %s", id)
	}

	// Delete user
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.WithRequestIDLogger(ctx).Infof("User deleted successfully: %s", id)
	return nil
}

// Register registers a new user
func (uc *UserUseCase) Register(ctx context.Context, user *domain.User) error {
	if user.Email == "" || user.Password == "" || user.Name == "" {
		return fmt.Errorf("validation failed: missing required fields")
	}

	// Check if user with email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Hash password
	hashedPassword, err := uc.authSvc.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	// Set default values
	user.IsActive = true

	// Create user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.WithRequestIDLogger(ctx).Infof("User registered successfully: %s", user.ID)
	return nil
}

// Login authenticates a user and returns JWT tokens
func (uc *UserUseCase) Login(ctx context.Context, email, password string) (*types.AuthResponse, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Verify password
	if err := uc.authSvc.CheckPassword(password, user.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate auth response
	authResponse, err := uc.authSvc.GenerateAuthResponse(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth response: %w", err)
	}

	// Update refresh token in database
	if err := uc.userRepo.UpdateRefreshToken(ctx, user.ID, authResponse.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to update refresh token: %w", err)
	}

	return handler.ToAuthResponse(authResponse), nil
}

// RefreshToken refreshes an access token using a refresh token
func (uc *UserUseCase) RefreshToken(ctx context.Context, refreshToken string) (*types.AuthResponse, error) {
	// Get user by refresh token
	user, err := uc.userRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Generate new auth response
	authResponse, err := uc.authSvc.GenerateAuthResponse(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth response: %w", err)
	}

	// Update refresh token in database
	if err := uc.userRepo.UpdateRefreshToken(ctx, user.ID, authResponse.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to update refresh token: %w", err)
	}

	return handler.ToAuthResponse(authResponse), nil
}

// Logout logs out a user by clearing their refresh token
func (uc *UserUseCase) Logout(ctx context.Context, userID uuid.UUID) error {
	// Clear refresh token
	if err := uc.userRepo.UpdateRefreshToken(ctx, userID, ""); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	logger.WithRequestIDLogger(ctx).Infof("User logged out successfully: %s", userID)
	return nil
}

func (uc *UserUseCase) GetMe(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}
