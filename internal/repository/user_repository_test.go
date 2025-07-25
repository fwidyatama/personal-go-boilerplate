package repository_test

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/repository"
)

var _ = Describe("UserRepository", func() {
	var (
		db   *sql.DB
		mock sqlmock.Sqlmock
		repo domain.UserRepository
		ctx  context.Context
	)

	BeforeEach(func() {
		var err error
		db, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		repo = repository.NewUserRepository(db)
		ctx = context.Background()
	})

	AfterEach(func() {
		_ = db.Close()
	})

	It("should create a user", func() {
		token := "token"
		user := &domain.User{
			Name:         "John Doe",
			Email:        "john@example.com",
			Password:     "hashedpassword",
			RefreshToken: &token,
			IsActive:     true,
		}
		now := time.Now()
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO users (name, email, password, refresh_token, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at`)).
			WithArgs(user.Name, user.Email, user.Password, user.RefreshToken, user.IsActive, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(uuid.New(), now, now))

		err := repo.Create(ctx, user)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should get a user by ID", func() {
		id := uuid.New()
		token := "token"
		now := time.Now()
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id, name, email, password, refresh_token, is_active, created_at, updated_at FROM users WHERE id = $1`)).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "refresh_token", "is_active", "created_at", "updated_at"}).
				AddRow(id, "John Doe", "john@example.com", "hashedpassword", &token, true, now, now))

		user, err := repo.GetByID(ctx, id)
		Expect(err).ToNot(HaveOccurred())
		Expect(user).ToNot(BeNil())
		Expect(user.ID).To(Equal(id))
		Expect(user.RefreshToken).ToNot(BeNil())
		Expect(*user.RefreshToken).To(Equal(token))
	})

	It("should return error if user by ID not found", func() {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id, name, email, password, refresh_token, is_active, created_at, updated_at FROM users WHERE id = $1`)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(ctx, id)
		Expect(err).To(HaveOccurred())
		Expect(user).To(BeNil())
	})

	It("should get a user by email", func() {
		email := "john@example.com"
		token := "token"
		now := time.Now()
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id, name, email, password, refresh_token, is_active, created_at, updated_at FROM users WHERE email = $1`)).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "refresh_token", "is_active", "created_at", "updated_at"}).
				AddRow(uuid.New(), "John Doe", email, "hashedpassword", &token, true, now, now))

		user, err := repo.GetByEmail(ctx, email)
		Expect(err).ToNot(HaveOccurred())
		Expect(user).ToNot(BeNil())
		Expect(user.Email).To(Equal(email))
	})

	It("should update a user", func() {
		id := uuid.New()
		token := "token"
		user := &domain.User{
			ID:           id,
			Name:         "John Doe",
			Email:        "john@example.com",
			Password:     "hashedpassword",
			RefreshToken: &token,
			IsActive:     true,
		}
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE users SET name = $1, email = $2, password = $3, refresh_token = $4, is_active = $5, updated_at = $6 WHERE id = $7`)).
			WithArgs(user.Name, user.Email, user.Password, user.RefreshToken, user.IsActive, sqlmock.AnyArg(), user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, user)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should delete a user", func() {
		id := uuid.New()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM users WHERE id = $1`)).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, id)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should count users", func() {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT COUNT(*) FROM users`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		count, err := repo.Count(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(5))
	})
})

// AnyTime is a sqlmock argument matcher for time.Time
// Use as: AnyTime{} in WithArgs
// See: https://github.com/DATA-DOG/go-sqlmock#matching-arguments-like-time
type AnyTime struct{}

func (a AnyTime) Match(v interface{}) bool {
	_, ok := v.(time.Time)
	return ok
}
