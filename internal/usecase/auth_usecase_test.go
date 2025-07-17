package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ucMock "go-boilerplate/internal/auth/mock"
	"go-boilerplate/internal/domain"
	repoMock "go-boilerplate/internal/repository/mock"
	redisMock "go-boilerplate/pkg/redis/mock"
)

var _ = Describe("AuthUseCase Password Management", func() {
	var (
		ctx            context.Context
		userRepository *repoMock.UserRepositoryMock
		redisClient    *redisMock.RedisClientMock
		authUseCase    domain.AuthUseCase
		mockAuthSvc    *ucMock.AuthServiceMock
	)

	BeforeEach(func() {
		ctx = context.Background()
		userRepository = &repoMock.UserRepositoryMock{}
		redisClient = &redisMock.RedisClientMock{}
		mockAuthSvc = &ucMock.AuthServiceMock{}
		authUseCase = NewAuthUseCase(userRepository, mockAuthSvc, redisClient)
	})

	Describe("ForgotPassword", func() {
		It("should return error if user not found", func() {
			userRepository.GetByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errors.New("not found")
			}
			err := authUseCase.ForgotPassword(ctx, "notfound@example.com")
			Expect(err).To(MatchError("user not found"))
		})
		It("should store token in redis if user exists", func() {
			userRepository.GetByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{Email: email}, nil
			}
			called := false
			redisClient.SetFunc = func(ctx context.Context, key string, value interface{}, ttl int) error {
				called = true
				return nil
			}
			err := authUseCase.ForgotPassword(ctx, "user@example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})

	Describe("ResetPassword", func() {
		It("should return error if token invalid", func() {
			redisClient.GetFunc = func(ctx context.Context, key string) (string, error) {
				return "othertoken", nil
			}
			err := authUseCase.ResetPassword(ctx, "user@example.com", "badtoken", "newpass")
			Expect(err).To(MatchError("invalid or expired token"))
		})
		It("should return error if user not found", func() {
			redisClient.GetFunc = func(ctx context.Context, key string) (string, error) {
				return "token", nil
			}
			userRepository.GetByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errors.New("not found")
			}
			err := authUseCase.ResetPassword(ctx, "user@example.com", "token", "newpass")
			Expect(err).To(MatchError("user not found"))
		})
		It("should update password if token valid", func() {
			redisClient.GetFunc = func(ctx context.Context, key string) (string, error) {
				return "token", nil
			}
			userRepository.GetByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{Email: email, Password: "old"}, nil
			}
			userRepository.UpdateFunc = func(ctx context.Context, user *domain.User) error {
				return nil
			}
			redisClient.DelFunc = func(ctx context.Context, key string) error {
				return nil
			}

			mockAuthSvc.HashPasswordFunc = func(password string) (string, error) {
				return "hashed", nil
			}

			err := authUseCase.ResetPassword(ctx, "user@example.com", "token", "newpass")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ChangePassword", func() {
		It("should return error if user not found", func() {
			userRepository.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return nil, errors.New("not found")
			}
			err := authUseCase.ChangePassword(ctx, uuid.Nil, "old", "new")
			Expect(err).To(MatchError("user not found"))
		})
		It("should return error if current password incorrect", func() {
			userRepository.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return &domain.User{Password: "hashed"}, nil
			}
			mockAuthSvc.CheckPasswordFunc = func(password string, hash string) error {
				return errors.New("current password incorrect")
			}
			err := authUseCase.ChangePassword(ctx, uuid.Nil, "wrong", "new")
			Expect(err).To(MatchError("current password incorrect"))
		})
		It("should update password if current password correct", func() {
			userRepository.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return &domain.User{Password: "hashed"}, nil
			}
			userRepository.UpdateFunc = func(ctx context.Context, user *domain.User) error {
				return nil
			}
			mockAuthSvc.HashPasswordFunc = func(password string) (string, error) {
				return "hashed", nil
			}
			mockAuthSvc.CheckPasswordFunc = func(password string, hash string) error {
				return nil
			}

			err := authUseCase.ChangePassword(ctx, uuid.Nil, "hashed", "new")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
