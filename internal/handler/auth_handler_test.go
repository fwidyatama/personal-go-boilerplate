package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"go-boilerplate/internal/domain"
	"go-boilerplate/internal/handler"
	"go-boilerplate/internal/types"
	authUseCaseMock "go-boilerplate/internal/usecase/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AuthHandler", func() {
	var (
		authHandler *handler.AuthHandler
		mockUseCase *authUseCaseMock.AuthUseCaseMock
		ginEngine   *gin.Engine
		recorder    *httptest.ResponseRecorder
		ctx         context.Context
		testUserID  uuid.UUID
		testUser    *domain.User
		testAuth    *domain.Auth
		testTime    time.Time
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		recorder = httptest.NewRecorder()
		ctx = context.Background()
		testUserID = uuid.New()
		testTime = time.Now()
		testUser = &domain.User{
			ID:        testUserID,
			Name:      "John Doe",
			Email:     "john@example.com",
			Password:  "hashedpassword",
			IsActive:  true,
			CreatedAt: testTime,
			UpdatedAt: testTime,
		}
		testAuth = &domain.Auth{
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_456",
			TokenType:    "Bearer",
			ExpiresIn:    900, // 15 minutes
			User:         testUser,
		}

		mockUseCase = &authUseCaseMock.AuthUseCaseMock{}
		authHandler = handler.NewAuthHandler(mockUseCase)
	})

	Describe("Register", func() {
		var registerRequest types.RegisterRequest

		BeforeEach(func() {
			registerRequest = types.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Password123!",
			}
		})

		Context("when registration is successful", func() {
			It("should register user and return auth response", func() {
				// Setup mocks
				mockUseCase.RegisterFunc = func(ctx context.Context, user *domain.User) error {
					Expect(user.Name).To(Equal(registerRequest.Name))
					Expect(user.Email).To(Equal(string(registerRequest.Email)))
					Expect(user.Password).To(Equal(registerRequest.Password))
					Expect(user.IsActive).To(BeTrue())
					return nil
				}

				mockUseCase.LoginFunc = func(ctx context.Context, email, password string) (*domain.Auth, error) {
					Expect(email).To(Equal(string(registerRequest.Email)))
					Expect(password).To(Equal(registerRequest.Password))
					return testAuth, nil
				}

				// Create request
				jsonBody, _ := json.Marshal(registerRequest)
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/register", authHandler.Register)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusCreated))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("User registered successfully"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/register", authHandler.Register)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid request body"))
			})
		})

		Context("when user with email already exists", func() {
			It("should return conflict error", func() {
				// Setup mock
				mockUseCase.RegisterFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("user with email " + string(registerRequest.Email) + " already exists")
				}

				// Create request
				jsonBody, _ := json.Marshal(registerRequest)
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/register", authHandler.Register)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusConflict))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(ContainSubstring("already exists"))
			})
		})

		Context("when registration fails with other error", func() {
			It("should return internal server error", func() {
				// Setup mock
				mockUseCase.RegisterFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("database error")
				}

				// Create request
				jsonBody, _ := json.Marshal(registerRequest)
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/register", authHandler.Register)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to register user"))
			})
		})

		Context("when login after registration fails", func() {
			It("should return internal server error", func() {
				// Setup mocks
				mockUseCase.RegisterFunc = func(ctx context.Context, user *domain.User) error {
					return nil
				}

				mockUseCase.LoginFunc = func(ctx context.Context, email, password string) (*domain.Auth, error) {
					return nil, errors.New("login failed")
				}

				// Create request
				jsonBody, _ := json.Marshal(registerRequest)
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/register", authHandler.Register)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to complete registration"))
			})
		})
	})

	Describe("Login", func() {
		var loginRequest types.LoginRequest

		BeforeEach(func() {
			loginRequest = types.LoginRequest{
				Email:    "john@example.com",
				Password: "Password123!",
			}
		})

		Context("when login is successful", func() {
			It("should return auth response", func() {
				// Setup mock
				mockUseCase.LoginFunc = func(ctx context.Context, email, password string) (*domain.Auth, error) {
					Expect(email).To(Equal(string(loginRequest.Email)))
					Expect(password).To(Equal(loginRequest.Password))
					return testAuth, nil
				}

				// Create request
				jsonBody, _ := json.Marshal(loginRequest)
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/login", authHandler.Login)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Login successful"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/login", authHandler.Login)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid request body"))
			})
		})

		Context("when credentials are invalid", func() {
			It("should return unauthorized error", func() {
				// Setup mock
				mockUseCase.LoginFunc = func(ctx context.Context, email, password string) (*domain.Auth, error) {
					return nil, errors.New("invalid credentials")
				}

				// Create request
				jsonBody, _ := json.Marshal(loginRequest)
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/login", authHandler.Login)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid credentials"))
			})
		})

		Context("when login fails with other error", func() {
			It("should return internal server error", func() {
				// Setup mock
				mockUseCase.LoginFunc = func(ctx context.Context, email, password string) (*domain.Auth, error) {
					return nil, errors.New("database error")
				}

				// Create request
				jsonBody, _ := json.Marshal(loginRequest)
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/login", authHandler.Login)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to login"))
			})
		})
	})

	Describe("Refresh", func() {
		var refreshRequest types.RefreshRequest

		BeforeEach(func() {
			refreshRequest = types.RefreshRequest{
				RefreshToken: "refresh_token_456",
			}
		})

		Context("when refresh is successful", func() {
			It("should return new auth response", func() {
				// Setup mock
				mockUseCase.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*domain.Auth, error) {
					Expect(refreshToken).To(Equal(refreshRequest.RefreshToken))
					return testAuth, nil
				}

				// Create request
				jsonBody, _ := json.Marshal(refreshRequest)
				req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/refresh", authHandler.Refresh)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Token refreshed successfully"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/refresh", authHandler.Refresh)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid request body"))
			})
		})

		Context("when refresh token is invalid", func() {
			It("should return unauthorized error", func() {
				// Setup mock
				mockUseCase.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*domain.Auth, error) {
					return nil, errors.New("invalid refresh token")
				}

				// Create request
				jsonBody, _ := json.Marshal(refreshRequest)
				req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/refresh", authHandler.Refresh)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid refresh token"))
			})
		})
	})

	Describe("Logout", func() {
		Context("when logout is successful", func() {
			It("should logout user successfully", func() {
				// Setup mock
				mockUseCase.LogoutFunc = func(ctx context.Context, userID uuid.UUID) error {
					Expect(userID).To(Equal(testUserID))
					return nil
				}

				// Create request
				req := httptest.NewRequest("POST", "/auth/logout", nil)
				req = req.WithContext(ctx)

				// Setup Gin context with user_id
				ginEngine.POST("/auth/logout", func(c *gin.Context) {
					c.Set("user_id", testUserID)
					authHandler.Logout(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Logout successful"))
			})
		})

		Context("when user_id is not in context", func() {
			It("should return unauthorized error", func() {
				// Create request
				req := httptest.NewRequest("POST", "/auth/logout", nil)
				req = req.WithContext(ctx)

				// Setup Gin context without user_id
				ginEngine.POST("/auth/logout", authHandler.Logout)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("User not authenticated"))
			})
		})

		Context("when user_id is not a UUID", func() {
			It("should return unauthorized error", func() {
				// Create request
				req := httptest.NewRequest("POST", "/auth/logout", nil)
				req = req.WithContext(ctx)

				// Setup Gin context with invalid user_id
				ginEngine.POST("/auth/logout", func(c *gin.Context) {
					c.Set("user_id", "not-a-uuid")
					authHandler.Logout(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid user ID"))
			})
		})

		Context("when logout fails", func() {
			It("should return internal server error", func() {
				// Setup mock
				mockUseCase.LogoutFunc = func(ctx context.Context, userID uuid.UUID) error {
					return errors.New("logout failed")
				}

				// Create request
				req := httptest.NewRequest("POST", "/auth/logout", nil)
				req = req.WithContext(ctx)

				// Setup Gin context with user_id
				ginEngine.POST("/auth/logout", func(c *gin.Context) {
					c.Set("user_id", testUserID)
					authHandler.Logout(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to logout"))
			})
		})
	})

	Describe("ForgotPassword", func() {
		var forgotPasswordRequest types.ForgotPasswordRequest

		BeforeEach(func() {
			forgotPasswordRequest = types.ForgotPasswordRequest{
				Email: "john@example.com",
			}
		})

		Context("when forgot password is successful", func() {
			It("should return success message", func() {
				// Setup mock
				mockUseCase.ForgotPasswordFunc = func(ctx context.Context, email string) error {
					Expect(email).To(Equal(string(forgotPasswordRequest.Email)))
					return nil
				}

				// Create request
				jsonBody, _ := json.Marshal(forgotPasswordRequest)
				req := httptest.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/forgot-password", authHandler.ForgotPassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Password reset email sent (if user exists)"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("POST", "/auth/forgot-password", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/forgot-password", authHandler.ForgotPassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid request body"))
			})
		})

		Context("when user not found", func() {
			It("should return not found error", func() {
				// Setup mock
				mockUseCase.ForgotPasswordFunc = func(ctx context.Context, email string) error {
					return errors.New("user not found")
				}

				// Create request
				jsonBody, _ := json.Marshal(forgotPasswordRequest)
				req := httptest.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/forgot-password", authHandler.ForgotPassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("user not found"))
			})
		})
	})

	Describe("ResetPassword", func() {
		var resetPasswordRequest types.ResetPasswordRequest

		BeforeEach(func() {
			resetPasswordRequest = types.ResetPasswordRequest{
				Email:       "john@example.com",
				Token:       "reset_token_123",
				NewPassword: "NewPassword123!",
			}
		})

		Context("when reset password is successful", func() {
			It("should return success message", func() {
				// Setup mock
				mockUseCase.ResetPasswordFunc = func(ctx context.Context, email, token, newPassword string) error {
					Expect(email).To(Equal(string(resetPasswordRequest.Email)))
					Expect(token).To(Equal(resetPasswordRequest.Token))
					Expect(newPassword).To(Equal(resetPasswordRequest.NewPassword))
					return nil
				}

				// Create request
				jsonBody, _ := json.Marshal(resetPasswordRequest)
				req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/reset-password", authHandler.ResetPassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Password reset successful"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/reset-password", authHandler.ResetPassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid request body"))
			})
		})

		Context("when token is invalid or expired", func() {
			It("should return not found error", func() {
				// Setup mock
				mockUseCase.ResetPasswordFunc = func(ctx context.Context, email, token, newPassword string) error {
					return errors.New("invalid or expired token")
				}

				// Create request
				jsonBody, _ := json.Marshal(resetPasswordRequest)
				req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/auth/reset-password", authHandler.ResetPassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("invalid or expired token"))
			})
		})
	})

	Describe("ChangePassword", func() {
		var changePasswordRequest types.ChangePasswordRequest

		BeforeEach(func() {
			changePasswordRequest = types.ChangePasswordRequest{
				CurrentPassword: "OldPassword123!",
				NewPassword:     "NewPassword123!",
			}
		})

		Context("when change password is successful", func() {
			It("should return success message", func() {
				// Setup mock
				mockUseCase.ChangePasswordFunc = func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
					Expect(userID).To(Equal(testUserID))
					Expect(currentPassword).To(Equal(changePasswordRequest.CurrentPassword))
					Expect(newPassword).To(Equal(changePasswordRequest.NewPassword))
					return nil
				}

				// Create request
				jsonBody, _ := json.Marshal(changePasswordRequest)
				req := httptest.NewRequest("PUT", "/auth/change-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context with user_id
				ginEngine.PUT("/auth/change-password", func(c *gin.Context) {
					c.Set("user_id", testUserID)
					authHandler.ChangePassword(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Password changed successfully"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("PUT", "/auth/change-password", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.PUT("/auth/change-password", authHandler.ChangePassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid request body"))
			})
		})

		Context("when user_id is not in context", func() {
			It("should return unauthorized error", func() {
				// Create request
				jsonBody, _ := json.Marshal(changePasswordRequest)
				req := httptest.NewRequest("PUT", "/auth/change-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context without user_id
				ginEngine.PUT("/auth/change-password", authHandler.ChangePassword)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("User not authenticated"))
			})
		})

		Context("when user_id is not a UUID", func() {
			It("should return unauthorized error", func() {
				// Create request
				jsonBody, _ := json.Marshal(changePasswordRequest)
				req := httptest.NewRequest("PUT", "/auth/change-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context with invalid user_id
				ginEngine.PUT("/auth/change-password", func(c *gin.Context) {
					c.Set("user_id", "not-a-uuid")
					authHandler.ChangePassword(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid user ID"))
			})
		})

		Context("when current password is incorrect", func() {
			It("should return forbidden error", func() {
				// Setup mock
				mockUseCase.ChangePasswordFunc = func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
					return errors.New("current password incorrect")
				}

				// Create request
				jsonBody, _ := json.Marshal(changePasswordRequest)
				req := httptest.NewRequest("PUT", "/auth/change-password", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context with user_id
				ginEngine.PUT("/auth/change-password", func(c *gin.Context) {
					c.Set("user_id", testUserID)
					authHandler.ChangePassword(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusForbidden))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("current password incorrect"))
			})
		})
	})

	Describe("Helper Functions", func() {
		Describe("ToAuthUserDetailResponse", func() {
			It("should convert domain.User to types.UserResponse", func() {
				response := handler.ToAuthUserDetailResponse(testUser)

				Expect(*response.Id).To(Equal(testUserID))
				Expect(*response.Name).To(Equal(testUser.Name))
				Expect(*response.Email).To(Equal(openapi_types.Email(testUser.Email)))
				Expect(*response.IsActive).To(Equal(testUser.IsActive))
			})
		})

		Describe("ToRegisterAuthResponse", func() {
			It("should convert domain.User and domain.Auth to types.AuthResponse", func() {
				response := handler.ToRegisterAuthResponse(testUser, testAuth)

				Expect(*response.AccessToken).To(Equal(testAuth.AccessToken))
				Expect(*response.TokenType).To(Equal(testAuth.TokenType))
				Expect(*response.ExpiresIn).To(Equal(int(testAuth.ExpiresIn)))
				Expect(response.User).ToNot(BeNil())
				Expect(*response.User.Id).To(Equal(testUserID))
			})

			It("should handle nil user", func() {
				response := handler.ToRegisterAuthResponse(nil, testAuth)
				Expect(response).To(BeNil())
			})

			It("should handle nil auth", func() {
				response := handler.ToRegisterAuthResponse(testUser, nil)
				Expect(response).To(BeNil())
			})
		})

		Describe("ToAuthResponse", func() {
			It("should convert domain.Auth to types.AuthResponse", func() {
				response := handler.ToAuthResponse(testAuth)

				Expect(*response.AccessToken).To(Equal(testAuth.AccessToken))
				Expect(*response.RefreshToken).To(Equal(testAuth.RefreshToken))
				Expect(*response.TokenType).To(Equal(testAuth.TokenType))
				Expect(*response.ExpiresIn).To(Equal(int(testAuth.ExpiresIn)))
				Expect(response.User).ToNot(BeNil())
				Expect(*response.User.Id).To(Equal(testUserID))
			})

			It("should handle nil auth", func() {
				response := handler.ToAuthResponse(nil)
				Expect(response).To(BeNil())
			})
		})
	})
})
