// FILE: internal/auth-service/auth/handlers_test.go
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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
