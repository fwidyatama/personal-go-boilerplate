package registry

import (
	"database/sql"
	"go-boilerplate/internal/auth"
	"go-boilerplate/pkg/redis"
)

type AppRegistry struct {
	DB       *sql.DB
	Auth     auth.AuthService
	Redis    redis.RedisClient
	Repos    *RepoRegistry
	Usecases *UsecaseRegistry
	Handlers *HandlerRegistry
}

func NewAppRegistry(db *sql.DB, authService auth.AuthService, redisClient redis.RedisClient) *AppRegistry {
	repos := NewRepoRegistry(db)
	usecases := NewUsecaseRegistry(repos, authService, redisClient)
	handlers := NewHandlerRegistry(usecases, authService)

	return &AppRegistry{
		DB:       db,
		Auth:     authService,
		Redis:    redisClient,
		Repos:    repos,
		Usecases: usecases,
		Handlers: handlers,
	}
}
