package user

// NOTE: This file contains swagger annotations for the user handlers.
// Run `swag init` to generate the swagger documentation.

// HandleGetCurrentUser godoc
// @Summary      Get current user profile
// @Description  Returns the profile information of the authenticated user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200 {object} User "User profile retrieved successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      404 {object} gin.H "User not found"
// @Router       /api/v1/user/profile [get]
// @Security     BearerAuth
// @ID           getCurrentUserProfile

// HandleUpdateCurrentUser godoc
// @Summary      Update current user profile
// @Description  Updates the profile information of the authenticated user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body UpdateUserRequest true "Profile update details"
// @Success      200 {object} User "Profile updated successfully"
// @Failure      400 {object} gin.H "Bad request"
// @Failure      401 {object} gin.H "Unauthorized"
// @Router       /api/v1/user/profile [put]
// @Security     BearerAuth
// @ID           updateCurrentUserProfile

// HandleChangePassword godoc
// @Summary      Change password
// @Description  Changes the user's password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body ChangePasswordRequest true "Password change details"
// @Success      200 {object} gin.H "Password changed successfully"
// @Failure      400 {object} gin.H "Bad request"
// @Failure      401 {object} gin.H "Unauthorized"
// @Router       /api/v1/user/password [post]
// @Security     BearerAuth
// @ID           changePassword

// HandleDeleteAccount godoc
// @Summary      Delete user account
// @Description  Permanently deletes the user's account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200 {object} gin.H "Account deleted successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Router       /api/v1/user/delete [delete]
// @Security     BearerAuth
// @ID           deleteUserAccount
