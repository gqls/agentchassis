package auth

// NOTE: This file contains swagger annotations for the auth handlers.
// Run `swag init` to generate the swagger documentation.

// HandleRegister godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the provided credentials
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201 {object} TokenResponse "User successfully registered"
// @Failure      400 {object} gin.H "Bad request"
// @Failure      409 {object} gin.H "User already exists"
// @Router       /api/v1/auth/register [post]
// @ID           registerUser

// HandleLogin godoc
// @Summary      User login
// @Description  Authenticates a user and returns access tokens
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} TokenResponse "Login successful"
// @Failure      401 {object} gin.H "Invalid credentials"
// @Router       /api/v1/auth/login [post]
// @ID           loginUser

// HandleRefresh godoc
// @Summary      Refresh access token
// @Description  Uses a refresh token to obtain a new access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RefreshRequest true "Refresh token"
// @Success      200 {object} TokenResponse "Token refreshed successfully"
// @Failure      401 {object} gin.H "Invalid refresh token"
// @Router       /api/v1/auth/refresh [post]
// @ID           refreshToken

// HandleValidate godoc
// @Summary      Validate token
// @Description  Validates an access token and returns user information
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Token is valid"
// @Failure      401 {object} gin.H "Invalid token"
// @Router       /api/v1/auth/validate [post]
// @Security     BearerAuth
// @ID           validateToken

// HandleLogout godoc
// @Summary      Logout user
// @Description  Invalidates the current session
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200 {object} gin.H "Logout successful"
// @Failure      401 {object} gin.H "Unauthorized"
// @Router       /api/v1/auth/logout [post]
// @Security     BearerAuth
// @ID           logoutUser
