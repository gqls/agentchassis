// FILE: internal/auth-service/auth/handlers_test.go
package auth

import (
	"bytes"
	"context" // Add this import
	"encoding/json"
	"fmt"
	"github.com/gqls/agentchassis/internal/auth-service/jwt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/auth-service/user" // Add this import
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService for testing
type MockService struct {
	mock.Mock
}

func (m *MockService) Register(ctx context.Context, req *user.CreateUserRequest) (*TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenResponse), args.Error(1)
}

func (m *MockService) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenResponse), args.Error(1)
}

func (m *MockService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenResponse), args.Error(1)
}

func (m *MockService) Logout(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func TestHandleRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    RegisterRequest
		mockResponse   *TokenResponse
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				ClientID:  "client-1",
				FirstName: "Test",
				LastName:  "User",
			},
			mockResponse: &TokenResponse{
				AccessToken: "new_access_token",
				TokenType:   "Bearer",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "new_access_token",
		},
		{
			name: "user already exists",
			requestBody: RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
				ClientID: "client-1",
			},
			mockError:      fmt.Errorf("user with email existing@example.com already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody:   "already exists",
		},
		{
			name: "bad request - missing password",
			requestBody: RegisterRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := &Handlers{service: mockService}

			// Setup mock expectation
			if tt.mockError != nil || tt.mockResponse != nil {
				// We expect the service's Register method to be called
				userReq := &user.CreateUserRequest{
					Email:     tt.requestBody.Email,
					Password:  tt.requestBody.Password,
					ClientID:  tt.requestBody.ClientID,
					FirstName: tt.requestBody.FirstName,
					LastName:  tt.requestBody.LastName,
				}
				mockService.On("Register", mock.Anything, userReq).Return(tt.mockResponse, tt.mockError).Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.HandleRegister(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			// Verify that the mock was called if it was expected
			if tt.mockError != nil || tt.mockResponse != nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    LoginRequest
		mockResponse   *TokenResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockResponse: &TokenResponse{
				AccessToken:  "test_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "wrong_password",
			},
			mockError:      fmt.Errorf("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := &Handlers{service: mockService}

			if tt.mockResponse != nil {
				mockService.On("Login", mock.Anything, tt.requestBody.Email, tt.requestBody.Password).
					Return(tt.mockResponse, nil)
			} else {
				mockService.On("Login", mock.Anything, tt.requestBody.Email, tt.requestBody.Password).
					Return(nil, tt.mockError)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.HandleLogin(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    RefreshRequest
		mockResponse   *TokenResponse
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful refresh",
			requestBody: RefreshRequest{
				RefreshToken: "valid_refresh_token",
			},
			mockResponse: &TokenResponse{
				AccessToken: "new_access_token_from_refresh",
				TokenType:   "Bearer",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "new_access_token_from_refresh",
		},
		{
			name: "invalid refresh token",
			requestBody: RefreshRequest{
				RefreshToken: "invalid_or_expired_token",
			},
			mockError:      fmt.Errorf("invalid refresh token"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := &Handlers{service: mockService}

			// Setup mock expectation
			mockService.On("RefreshToken", mock.Anything, tt.requestBody.RefreshToken).
				Return(tt.mockResponse, tt.mockError).
				Once()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.HandleRefresh(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}
