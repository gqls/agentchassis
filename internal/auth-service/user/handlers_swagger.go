package user

// NOTE: This file contains swagger annotations for the user handlers.
// Run `swag init` to generate the swagger documentation.

// HandleGetCurrentUser godoc
// @Summary      Get current user profile
// @Description  Retrieves the profile information of the currently authenticated user
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Success      200 {object} UserProfileResponse "User profile retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /user/profile [get]
// @Security     Bearer
// @ID           getCurrentUser

// HandleUpdateCurrentUser godoc
// @Summary      Update user profile
// @Description  Updates the profile information of the currently authenticated user
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        request body UpdateProfileRequest true "Profile update details"
// @Success      200 {object} UserProfileResponse "Profile updated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Failure      409 {object} map[string]interface{} "Email already in use"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /user/profile [put]
// @Security     Bearer
// @ID           updateUserProfile

// HandleChangePassword godoc
// @Summary      Change password
// @Description  Changes the password for the currently authenticated user
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        request body ChangePasswordRequest true "Password change details"
// @Success      200 {object} map[string]interface{} "Password changed successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request or password requirements not met"
// @Failure      401 {object} map[string]interface{} "Unauthorized or incorrect current password"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /user/password [post]
// @Security     Bearer
// @ID           changePassword

// HandleDeleteAccount godoc
// @Summary      Delete account
// @Description  Permanently deletes the currently authenticated user's account. This action cannot be undone.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        request body DeleteAccountRequest true "Account deletion confirmation"
// @Success      200 {object} map[string]interface{} "Account deleted successfully"
// @Failure      400 {object} map[string]interface{} "Invalid confirmation or request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized or incorrect password"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /user/delete [delete]
// @Security     Bearer
// @ID           deleteAccount
