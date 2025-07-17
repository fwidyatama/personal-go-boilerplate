package registry

import (
	"database/sql"
	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/repository"
)

type RepoRegistry struct {
	UserRepo domain.UserRepository
}

func NewRepoRegistry(db *sql.DB) *RepoRegistry {

	userRepo := repository.NewUserRepository(db)

	return &RepoRegistry{
		UserRepo: userRepo,
	}
}
