package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a user entity
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	RefreshToken *string   `json:"-"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

//go:generate moq -pkg mock -out ../repository/mock/user_repository_moq.go ../domain UserRepository
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*User, error)
}

//go:generate moq -pkg mock -out ../usecase/mock/user_usecase_moq.go ../domain UserUseCase
type UserUseCase interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetMe(ctx context.Context, userID uuid.UUID) (*User, error)
}
