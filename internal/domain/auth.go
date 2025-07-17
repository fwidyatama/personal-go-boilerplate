package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Auth represents authentication entity
type Auth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

// Claims represents JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	SessionID string    `json:"session_id"`
	jwt.RegisteredClaims
}

//go:generate moq -pkg mock -out ../usecase/mock/auth_usecase_moq.go ../domain AuthUseCase
type AuthUseCase interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, email, password string) (*Auth, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Auth, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, token, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}
