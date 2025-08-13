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
	usecaseMock "go-boilerplate/internal/usecase/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserHandler", func() {
	var (
		userHandler *handler.UserHandler
		mockUseCase *usecaseMock.UserUseCaseMock
		ginEngine   *gin.Engine
		recorder    *httptest.ResponseRecorder
		ctx         context.Context
		testUserID  uuid.UUID
		testUser    *domain.User
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

		mockUseCase = &usecaseMock.UserUseCaseMock{}
		userHandler = handler.NewUserHandler(mockUseCase)
	})

	Describe("Create", func() {
		var createRequest types.CreateUserRequest

		BeforeEach(func() {
			createRequest = types.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Password123!",
			}
		})

		Context("when request is valid", func() {
			It("should create user successfully", func() {
				// Setup mock
				mockUseCase.CreateFunc = func(ctx context.Context, user *domain.User) error {
					Expect(user.Name).To(Equal(createRequest.Name))
					Expect(user.Email).To(Equal(string(createRequest.Email)))
					Expect(user.Password).To(Equal(createRequest.Password))
					user.ID = testUserID
					user.CreatedAt = testTime
					user.UpdatedAt = testTime
					return nil
				}

				// Create request
				jsonBody, _ := json.Marshal(createRequest)
				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/users", userHandler.Create)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusCreated))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("User created successfully"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create invalid request
				req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/users", userHandler.Create)
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
				mockUseCase.CreateFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("user with email " + string(createRequest.Email) + " already exists")
				}

				// Create request
				jsonBody, _ := json.Marshal(createRequest)
				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/users", userHandler.Create)
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

		Context("when usecase returns other error", func() {
			It("should return internal server error", func() {
				// Setup mock
				mockUseCase.CreateFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("database error")
				}

				// Create request
				jsonBody, _ := json.Marshal(createRequest)
				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.POST("/users", userHandler.Create)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to create user"))
			})
		})
	})

	Describe("GetByID", func() {
		Context("when user exists", func() {
			It("should return user successfully", func() {
				// Setup mock
				mockUseCase.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					Expect(id).To(Equal(testUserID))
					return testUser, nil
				}

				// Create request
				req := httptest.NewRequest("GET", "/users/"+testUserID.String(), nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users/:id", userHandler.GetByID)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("User retrieved successfully"))
			})
		})

		Context("when user ID is invalid", func() {
			It("should return bad request error", func() {
				// Create request with invalid UUID
				req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users/:id", userHandler.GetByID)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid user ID"))
			})
		})

		Context("when user not found", func() {
			It("should return not found error", func() {
				// Setup mock
				mockUseCase.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, errors.New("user not found: " + id.String())
				}

				// Create request
				req := httptest.NewRequest("GET", "/users/"+testUserID.String(), nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users/:id", userHandler.GetByID)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(ContainSubstring("user not found"))
			})
		})

		Context("when usecase returns other error", func() {
			It("should return internal server error", func() {
				// Setup mock
				mockUseCase.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, errors.New("database error")
				}

				// Create request
				req := httptest.NewRequest("GET", "/users/"+testUserID.String(), nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users/:id", userHandler.GetByID)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to get user"))
			})
		})
	})

	Describe("List", func() {
		var testUsers []*domain.User

		BeforeEach(func() {
			testUsers = []*domain.User{
				{
					ID:        uuid.New(),
					Name:      "John Doe",
					Email:     "john@example.com",
					IsActive:  true,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
				{
					ID:        uuid.New(),
					Name:      "Jane Smith",
					Email:     "jane@example.com",
					IsActive:  true,
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			}
		})

		Context("when users exist", func() {
			It("should return users successfully with default pagination", func() {
				// Setup mock
				mockUseCase.ListFunc = func(ctx context.Context, limit, offset int) ([]*domain.User, error) {
					Expect(limit).To(Equal(10))
					Expect(offset).To(Equal(0))
					return testUsers, nil
				}

				// Create request
				req := httptest.NewRequest("GET", "/users", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users", userHandler.List)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Users retrieved successfully"))
			})

			It("should handle custom pagination parameters", func() {
				// Setup mock
				mockUseCase.ListFunc = func(ctx context.Context, limit, offset int) ([]*domain.User, error) {
					Expect(limit).To(Equal(5))
					Expect(offset).To(Equal(10))
					return testUsers, nil
				}

				// Create request with custom pagination
				req := httptest.NewRequest("GET", "/users?limit=5&offset=10", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users", userHandler.List)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})

			It("should handle invalid limit parameter", func() {
				// Setup mock
				mockUseCase.ListFunc = func(ctx context.Context, limit, offset int) ([]*domain.User, error) {
					Expect(limit).To(Equal(10)) // Should default to 10
					Expect(offset).To(Equal(0))
					return testUsers, nil
				}

				// Create request with invalid limit
				req := httptest.NewRequest("GET", "/users?limit=invalid&offset=0", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users", userHandler.List)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})

			It("should handle limit greater than 100", func() {
				// Setup mock
				mockUseCase.ListFunc = func(ctx context.Context, limit, offset int) ([]*domain.User, error) {
					Expect(limit).To(Equal(10)) // Should default to 10
					Expect(offset).To(Equal(0))
					return testUsers, nil
				}

				// Create request with limit > 100
				req := httptest.NewRequest("GET", "/users?limit=150&offset=0", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users", userHandler.List)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when usecase returns error", func() {
			It("should return internal server error", func() {
				// Setup mock
				mockUseCase.ListFunc = func(ctx context.Context, limit, offset int) ([]*domain.User, error) {
					return nil, errors.New("database error")
				}

				// Create request
				req := httptest.NewRequest("GET", "/users", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.GET("/users", userHandler.List)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Failed to list users"))
			})
		})
	})

	Describe("Update", func() {
		var updateRequest types.UpdateUserRequest

		BeforeEach(func() {
			updateRequest = types.UpdateUserRequest{
				Name:     "John Updated",
				Email:    "john.updated@example.com",
				Password: "NewPassword123!",
			}
		})

		Context("when update is successful", func() {
			It("should update user successfully", func() {
				// Setup mock
				mockUseCase.UpdateFunc = func(ctx context.Context, user *domain.User) error {
					Expect(user.ID).To(Equal(testUserID))
					Expect(user.Name).To(Equal(updateRequest.Name))
					Expect(user.Email).To(Equal(string(updateRequest.Email)))
					Expect(user.Password).To(Equal(updateRequest.Password))
					return nil
				}

				// Create request
				jsonBody, _ := json.Marshal(updateRequest)
				req := httptest.NewRequest("PUT", "/users/"+testUserID.String(), bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.PUT("/users/:id", userHandler.Update)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("User updated successfully"))
			})
		})

		Context("when user ID is invalid", func() {
			It("should return bad request error", func() {
				// Create request with invalid UUID
				jsonBody, _ := json.Marshal(updateRequest)
				req := httptest.NewRequest("PUT", "/users/invalid-uuid", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.PUT("/users/:id", userHandler.Update)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid user ID"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				// Create request with invalid JSON
				req := httptest.NewRequest("PUT", "/users/"+testUserID.String(), bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.PUT("/users/:id", userHandler.Update)
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
				mockUseCase.UpdateFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("user not found: " + user.ID.String())
				}

				// Create request
				jsonBody, _ := json.Marshal(updateRequest)
				req := httptest.NewRequest("PUT", "/users/"+testUserID.String(), bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.PUT("/users/:id", userHandler.Update)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(ContainSubstring("user not found"))
			})
		})

		Context("when email already exists", func() {
			It("should return conflict error", func() {
				// Setup mock
				mockUseCase.UpdateFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("user with email " + user.Email + " already exists")
				}

				// Create request
				jsonBody, _ := json.Marshal(updateRequest)
				req := httptest.NewRequest("PUT", "/users/"+testUserID.String(), bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.PUT("/users/:id", userHandler.Update)
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
	})

	Describe("Delete", func() {
		Context("when delete is successful", func() {
			It("should delete user successfully", func() {
				// Setup mock
				mockUseCase.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					Expect(id).To(Equal(testUserID))
					return nil
				}

				// Create request
				req := httptest.NewRequest("DELETE", "/users/"+testUserID.String(), nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.DELETE("/users/:id", userHandler.Delete)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("when user ID is invalid", func() {
			It("should return bad request error", func() {
				// Create request with invalid UUID
				req := httptest.NewRequest("DELETE", "/users/invalid-uuid", nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.DELETE("/users/:id", userHandler.Delete)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Invalid user ID"))
			})
		})

		Context("when user not found", func() {
			It("should return not found error", func() {
				// Setup mock
				mockUseCase.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					return errors.New("user not found: " + id.String())
				}

				// Create request
				req := httptest.NewRequest("DELETE", "/users/"+testUserID.String(), nil)
				req = req.WithContext(ctx)

				// Setup Gin context
				ginEngine.DELETE("/users/:id", userHandler.Delete)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(ContainSubstring("user not found"))
			})
		})
	})

	Describe("GetMe", func() {
		Context("when user is authenticated and exists", func() {
			It("should return current user profile", func() {
				// Setup mock
				mockUseCase.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					Expect(id).To(Equal(testUserID))
					return testUser, nil
				}

				// Create request
				req := httptest.NewRequest("GET", "/users/me", nil)
				req = req.WithContext(ctx)

				// Setup Gin context with user_id
				ginEngine.GET("/users/me", func(c *gin.Context) {
					c.Set("user_id", testUserID)
					userHandler.GetMe(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("User profile retrieved successfully"))
			})
		})

		Context("when user_id is not in context", func() {
			It("should return unauthorized error", func() {
				// Create request
				req := httptest.NewRequest("GET", "/users/me", nil)
				req = req.WithContext(ctx)

				// Setup Gin context without user_id
				ginEngine.GET("/users/me", userHandler.GetMe)
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("Unauthorized"))
			})
		})

		Context("when user_id is not a UUID", func() {
			It("should return unauthorized error", func() {
				// Create request
				req := httptest.NewRequest("GET", "/users/me", nil)
				req = req.WithContext(ctx)

				// Setup Gin context with invalid user_id
				ginEngine.GET("/users/me", func(c *gin.Context) {
					c.Set("user_id", "not-a-uuid")
					userHandler.GetMe(c)
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

		Context("when user not found", func() {
			It("should return not found error", func() {
				// Setup mock
				mockUseCase.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, errors.New("user not found")
				}

				// Create request
				req := httptest.NewRequest("GET", "/users/me", nil)
				req = req.WithContext(ctx)

				// Setup Gin context with user_id
				ginEngine.GET("/users/me", func(c *gin.Context) {
					c.Set("user_id", testUserID)
					userHandler.GetMe(c)
				})
				ginEngine.ServeHTTP(recorder, req)

				// Assertions
				Expect(recorder.Code).To(Equal(http.StatusNotFound))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response["success"]).To(BeFalse())
				Expect(response["message"]).To(Equal("User not found"))
			})
		})
	})

	Describe("Helper Functions", func() {
		Describe("ToUserDetailResponse", func() {
			It("should convert domain.User to types.UserResponse", func() {
				response := handler.ToUserDetailResponse(testUser)

				Expect(*response.Id).To(Equal(testUserID))
				Expect(*response.Name).To(Equal(testUser.Name))
				Expect(*response.Email).To(Equal(openapi_types.Email(testUser.Email)))
				Expect(*response.IsActive).To(Equal(testUser.IsActive))
				Expect(*response.CreatedAt).To(Equal(testUser.CreatedAt))
				Expect(*response.UpdatedAt).To(Equal(testUser.UpdatedAt))
			})

			It("should handle nil user", func() {
				response := handler.ToUserDetailResponse(nil)
				Expect(response).To(Equal(types.UserResponse{}))
			})
		})

		Describe("ToUserListResponse", func() {
			It("should convert slice of domain.User to types.UserListResponse", func() {
				users := []*domain.User{testUser}
				response := handler.ToUserListResponse(users)

				Expect(*response.Users).To(HaveLen(1))
				Expect((*response.Users)[0].Id).To(Equal(&testUserID))
				Expect((*response.Users)[0].Name).To(Equal(&testUser.Name))
			})

			It("should handle empty slice", func() {
				users := []*domain.User{}
				response := handler.ToUserListResponse(users)

				Expect(*response.Users).To(HaveLen(0))
			})
		})
	})
})
