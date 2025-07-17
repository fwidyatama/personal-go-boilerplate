package registry

import (
	"go-boilerplate/internal/auth"
	"go-boilerplate/internal/handler"
)

type HandlerRegistry struct {
	UserHandler *handler.UserHandler
	AuthHandler *handler.AuthHandler
}

func NewHandlerRegistry(usecases *UsecaseRegistry, authService auth.AuthService) *HandlerRegistry {
	userHandler := handler.NewUserHandler(usecases.UserUsecase)
	authHandler := handler.NewAuthHandler(usecases.AuthUsecase)

	return &HandlerRegistry{
		UserHandler: userHandler,
		AuthHandler: authHandler,
	}
}
