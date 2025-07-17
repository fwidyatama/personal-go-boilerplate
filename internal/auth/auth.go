package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"go-boilerplate/internal/domain"

	"go-boilerplate/pkg/redis"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:generate moq -pkg mock -out mock/auth_service_moq.go . AuthService
type AuthService interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) error
	GenerateAuthResponse(user *domain.User) (*domain.Auth, error)
	ValidateAccessToken(tokenString string) (*domain.Claims, error)
}

// Service handles authentication operations
type Service struct {
	jwtSecret     []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	redis         redis.RedisClient
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret, refreshSecret string, accessTTL, refreshTTL time.Duration, redisClient redis.RedisClient) AuthService {
	return &Service{
		jwtSecret:     []byte(jwtSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		redis:         redisClient,
	}
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword checks if a password matches a hash
func (s *Service) CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}

// GenerateAccessToken generates a JWT access token
func (s *Service) GenerateAccessToken(user *domain.User, sessionID string) (string, error) {
	claims := domain.Claims{
		UserID:    user.ID,
		Email:     user.Email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-boilerplate",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateAccessToken validates and parses a JWT access token
func (s *Service) ValidateAccessToken(tokenString string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*domain.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateAuthResponse generates a complete authentication response
func (s *Service) GenerateAuthResponse(user *domain.User) (*domain.Auth, error) {
	sessionID := uuid.NewString()
	// Store sessionID in Redis with TTL = refreshTTL
	if err := s.redis.Set(context.Background(), user.ID.String(), sessionID, int(s.refreshTTL.Seconds())); err != nil {
		return nil, fmt.Errorf("failed to store session in redis: %w", err)
	}

	accessToken, err := s.GenerateAccessToken(user, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.Auth{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		User:         user,
	}, nil
}
