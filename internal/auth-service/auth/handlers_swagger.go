package auth

// NOTE: This file contains swagger annotations for the auth handlers.
// Run `swag init` to generate the swagger documentation.
// All types are defined in their respective handler files.

// HandleRegister godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the provided credentials and client association
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body auth.RegisterRequest true "Registration details"
// @Success      201 {object} auth.TokenResponse "User successfully registered with tokens"
// @Failure      400 {object} map[string]interface{} "Invalid request body or validation error"
// @Failure      409 {object} map[string]interface{} "User already exists"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /auth/register [post]
// @ID           registerUser

// HandleLogin godoc
// @Summary      User login
// @Description  Authenticates a user with email and password, returns access and refresh tokens
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body auth.LoginRequest true "Login credentials"
// @Success      200 {object} auth.TokenResponse "Login successful with tokens and user info"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Invalid credentials"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /auth/login [post]
// @ID           loginUser

// HandleRefresh godoc
// @Summary      Refresh access token
// @Description  Uses a valid refresh token to obtain a new access token and refresh token pair
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body auth.RefreshRequest true "Refresh token"
// @Success      200 {object} auth.TokenResponse "New tokens generated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Invalid or expired refresh token"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /auth/refresh [post]
// @ID           refreshToken

// HandleValidate godoc
// @Summary      Validate token
// @Description  Validates the provided access token and returns user information if valid
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Token is valid with user details"
// @Failure      401 {object} map[string]interface{} "Invalid or expired token"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /auth/validate [post]
// @Security     Bearer
// @ID           validateToken

// HandleLogout godoc
// @Summary      Logout user
// @Description  Invalidates the current session and revokes the refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Logout successful"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Failed to logout"
// @Router       /auth/logout [post]
// @Security     Bearer
// @ID           logoutUser
