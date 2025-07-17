package registry

import (
	"go-boilerplate/internal/auth"
	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/redis"
)

type UsecaseRegistry struct {
	UserUsecase domain.UserUseCase
	AuthUsecase domain.AuthUseCase
	Redis       redis.RedisClient
}

func NewUsecaseRegistry(repos *RepoRegistry, authService auth.AuthService, redisClient redis.RedisClient) *UsecaseRegistry {
	userUsecase := usecase.NewUserUseCase(repos.UserRepo, authService, redisClient)
	authUsecase := usecase.NewAuthUseCase(repos.UserRepo, authService, redisClient)

	return &UsecaseRegistry{
		UserUsecase: userUsecase,
		AuthUsecase: authUsecase,
		Redis:       redisClient,
	}
}
