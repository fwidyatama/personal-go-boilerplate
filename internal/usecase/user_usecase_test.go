package usecase

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authMock "go-boilerplate/internal/auth/mock"
	"go-boilerplate/internal/domain"
	repoMock "go-boilerplate/internal/repository/mock"
	redisMock "go-boilerplate/pkg/redis/mock"

	"github.com/google/uuid"
)

var _ = Describe("UserUseCase", func() {
	var (
		ctx            context.Context
		userRepository *repoMock.UserRepositoryMock
		authSvc        *authMock.AuthServiceMock
		userUseCase    domain.UserUseCase
		redisClient    *redisMock.RedisClientMock
	)

	BeforeEach(func() {
		ctx = context.Background()
		userRepository = &repoMock.UserRepositoryMock{}
		authSvc = &authMock.AuthServiceMock{}
		redisClient = &redisMock.RedisClientMock{}
		userUseCase = NewUserUseCase(userRepository, authSvc, redisClient)
	})

	Describe("Create", func() {
		It("should return error if validation fails", func() {
			user := &domain.User{Name: "", Email: "", Password: ""}
			err := userUseCase.Create(ctx, user)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("validation failed"))
		})

		It("should return error if user with email already exists", func() {
			user := &domain.User{Name: "John", Email: "john@example.com", Password: "password"}
			userRepository.GetByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{Email: email}, nil
			}
			err := userUseCase.Create(ctx, user)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already exists"))
		})

		It("should create user successfully", func() {
			user := &domain.User{Name: "John", Email: "john@example.com", Password: "password"}
			userRepository.GetByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errors.New("not found")
			}
			userRepository.CreateFunc = func(ctx context.Context, u *domain.User) error {
				return nil
			}
			authSvc.HashPasswordFunc = func(password string) (string, error) {
				return "hashed", nil
			}
			err := userUseCase.Create(ctx, user)
			Expect(err).ToNot(HaveOccurred())
			Expect(user.Password).NotTo(Equal("password"))
		})
	})

	Describe("GetMe", func() {
		It("should return user if found", func() {
			user := &domain.User{ID: uuid.New(), Name: "John", Email: "john@example.com"}
			userRepository.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return user, nil
			}
			result, err := userUseCase.GetMe(ctx, user.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(user))
		})
		It("should return error if user not found", func() {
			userRepository.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return nil, errors.New("not found")
			}
			_, err := userUseCase.GetMe(ctx, uuid.New())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("user not found"))
		})
		It("should return error if userID is invalid", func() {
			_, err := userUseCase.GetMe(ctx, uuid.Nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid user ID"))
		})
	})

})
