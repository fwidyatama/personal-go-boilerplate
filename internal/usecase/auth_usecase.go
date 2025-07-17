package usecase

import (
	"context"
	"fmt"

	"go-boilerplate/internal/auth"
	"go-boilerplate/internal/domain"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/redis"

	"github.com/google/uuid"
)

type AuthUseCase struct {
	userRepo domain.UserRepository
	authSvc  auth.AuthService
	redis    redis.RedisClient
}

func NewAuthUseCase(userRepo domain.UserRepository, authSvc auth.AuthService, redisClient redis.RedisClient) domain.AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		authSvc:  authSvc,
		redis:    redisClient,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, user *domain.User) error {
	if user.Email == "" || user.Password == "" || user.Name == "" {
		return fmt.Errorf("validation failed: missing required fields")
	}

	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	hashedPassword, err := uc.authSvc.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	user.ID = uuid.New()
	user.IsActive = true

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.WithRequestIDLogger(ctx).Infof("User registered successfully: %s", user.ID)
	return nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*domain.Auth, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}
	if err := uc.authSvc.CheckPassword(password, user.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	auth, err := uc.authSvc.GenerateAuthResponse(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth response: %w", err)
	}
	if err := uc.userRepo.UpdateRefreshToken(ctx, user.ID, auth.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to update refresh token: %w", err)
	}
	return auth, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.Auth, error) {
	user, err := uc.userRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	auth, err := uc.authSvc.GenerateAuthResponse(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth response: %w", err)
	}
	if err := uc.userRepo.UpdateRefreshToken(ctx, user.ID, auth.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to update refresh token: %w", err)
	}
	return auth, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := uc.userRepo.UpdateRefreshToken(ctx, userID, ""); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	if err := uc.redis.Del(ctx, userID.String()); err != nil {
		return fmt.Errorf("failed to delete session in redis: %w", err)
	}
	logger.WithRequestIDLogger(ctx).Infof("User logged out: %s", userID)
	return nil
}

func (uc *AuthUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}
	token := uuid.NewString()
	key := "resetpw:" + email
	err = uc.redis.Set(ctx, key, token, 15*60)
	if err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}
	logger.WithRequestIDLogger(ctx).Infof("Password reset token for %s: %s", email, token)
	return nil
}

func (uc *AuthUseCase) ResetPassword(ctx context.Context, email, token, newPassword string) error {
	key := "resetpw:" + email
	storedToken, err := uc.redis.Get(ctx, key)
	if err != nil || storedToken != token {
		return fmt.Errorf("invalid or expired token")
	}
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}
	hashedPassword, err := uc.authSvc.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	_ = uc.redis.Del(ctx, key)
	return nil
}

func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}
	if err := uc.authSvc.CheckPassword(currentPassword, user.Password); err != nil {
		return fmt.Errorf("current password incorrect")
	}
	hashedPassword, err := uc.authSvc.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
