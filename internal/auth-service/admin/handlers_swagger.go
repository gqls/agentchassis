package admin

// NOTE: This file contains swagger annotations for the admin handlers.
// Run `swag init` to generate the swagger documentation.
// All types are defined in their respective handler files.

// HandleListUsers godoc
// @Summary      List users
// @Description  Get a paginated list of all users with filtering and sorting options (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        page         query    int     false  "Page number"                          default(1) minimum(1)
// @Param        page_size    query    int     false  "Items per page"                       default(20) minimum(1) maximum(100)
// @Param        email        query    string  false  "Filter by email (partial match)"
// @Param        client_id    query    string  false  "Filter by client ID"
// @Param        role         query    string  false  "Filter by role"                       Enums(user,admin,moderator)
// @Param        tier         query    string  false  "Filter by subscription tier"          Enums(free,basic,premium,enterprise)
// @Param        is_active    query    boolean false  "Filter by active status"
// @Param        sort_by      query    string  false  "Sort field"                           default(created_at) Enums(created_at,updated_at,email,last_login_at)
// @Param        sort_order   query    string  false  "Sort order"                           default(desc) Enums(asc,desc)
// @Success      200          {object} admin.UserListResponse                 "List of users retrieved successfully"
// @Failure      400          {object} map[string]interface{}                 "Invalid query parameters"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users [get]
// @Security     Bearer
// @ID           adminListUsers

// HandleGetUser godoc
// @Summary      Get user details
// @Description  Get detailed information about a specific user including statistics (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Success      200          {object} map[string]interface{}                 "User details with statistics"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id} [get]
// @Security     Bearer
// @ID           adminGetUser

// HandleUpdateUser godoc
// @Summary      Update user
// @Description  Update user details including role, subscription tier, and status (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Param        request      body     admin.UpdateUserRequest true           "User update details"
// @Success      200          {object} user.User                              "Updated user details"
// @Failure      400          {object} map[string]interface{}                 "Invalid request body or invalid role/tier"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id} [put]
// @Security     Bearer
// @ID           adminUpdateUser

// HandleDeleteUser godoc
// @Summary      Delete user
// @Description  Soft delete a user account (admin only). Cannot delete your own account.
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Success      200          {object} map[string]interface{}                 "User deleted successfully"
// @Failure      400          {object} map[string]interface{}                 "Cannot delete your own account"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id} [delete]
// @Security     Bearer
// @ID           adminDeleteUser

// HandleGetUserActivity godoc
// @Summary      Get user activity
// @Description  Get activity logs for a specific user (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Param        limit        query    int     false  "Maximum number of activities"         default(50) maximum(200)
// @Param        offset       query    int     false  "Number of activities to skip"         default(0)
// @Success      200          {object} map[string]interface{}                 "User activity logs"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id}/activity [get]
// @Security     Bearer
// @ID           adminGetUserActivity

// HandleGrantPermission godoc
// @Summary      Grant permission
// @Description  Grant a specific permission to a user (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Param        request      body     admin.GrantPermissionRequest true      "Permission details"
// @Success      200          {object} map[string]interface{}                 "Permission granted successfully"
// @Failure      400          {object} map[string]interface{}                 "Invalid request body"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      409          {object} map[string]interface{}                 "Permission already exists"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id}/permissions [post]
// @Security     Bearer
// @ID           adminGrantPermission

// HandleRevokePermission godoc
// @Summary      Revoke permission
// @Description  Revoke a specific permission from a user (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id          path     string  true   "User ID"
// @Param        permission_name  path     string  true   "Permission name to revoke"
// @Success      200              {object} map[string]interface{}             "Permission revoked successfully"
// @Failure      401              {object} map[string]interface{}             "Unauthorized - no valid token"
// @Failure      403              {object} map[string]interface{}             "Forbidden - admin access required"
// @Failure      404              {object} map[string]interface{}             "User or permission not found"
// @Failure      500              {object} map[string]interface{}             "Internal server error"
// @Router       /admin/users/{user_id}/permissions/{permission_name} [delete]
// @Security     Bearer
// @ID           adminRevokePermission

// Bulk operations and advanced admin features

// HandleBulkUserOperation godoc
// @Summary      Bulk user operation
// @Description  Perform bulk operations on multiple users at once (admin only)
// @Tags         Admin - Bulk Operations
// @Accept       json
// @Produce      json
// @Param        request      body     admin.BulkUserOperation true           "Bulk operation details"
// @Success      200          {object} admin.BulkOperationResult              "Bulk operation results"
// @Failure      400          {object} map[string]interface{}                 "Invalid request body"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/bulk [post]
// @Security     Bearer
// @ID           adminBulkUserOperation

// HandleExportUsers godoc
// @Summary      Export users
// @Description  Export user data in CSV or JSON format with filtering (admin only)
// @Tags         Admin - Data Export
// @Accept       json
// @Produce      json,text/csv
// @Param        request      body     admin.UserExportRequest true           "Export configuration"
// @Success      200          {file}   file                                   "Exported user data (CSV or JSON)"
// @Failure      400          {object} map[string]interface{}                 "Invalid request body"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/export [post]
// @Security     Bearer
// @ID           adminExportUsers

// HandleImportUsers godoc
// @Summary      Import users
// @Description  Bulk import users from CSV file (admin only)
// @Tags         Admin - Data Import
// @Accept       multipart/form-data
// @Produce      json
// @Param        file         formData file    true   "CSV file containing user data"
// @Param        client_id    formData string  true   "Client ID for imported users"
// @Success      200          {object} admin.UserImportResult                "Import results with created user IDs"
// @Failure      400          {object} map[string]interface{}                 "Invalid file or missing client_id"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/import [post]
// @Security     Bearer
// @ID           adminImportUsers

// HandleGetUserSessions godoc
// @Summary      Get user sessions
// @Description  Get all active sessions for a specific user (admin only)
// @Tags         Admin - Session Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Success      200          {object} map[string]interface{}                 "List of active sessions"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id}/sessions [get]
// @Security     Bearer
// @ID           adminGetUserSessions

// HandleTerminateUserSessions godoc
// @Summary      Terminate user sessions
// @Description  Terminate all active sessions for a user, forcing them to re-login (admin only)
// @Tags         Admin - Session Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Param        request      body     admin.TerminateSessionsRequest true    "Termination reason"
// @Success      200          {object} map[string]interface{}                 "Sessions terminated successfully"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id}/sessions [delete]
// @Security     Bearer
// @ID           adminTerminateUserSessions

// HandleResetUserPassword godoc
// @Summary      Reset user password
// @Description  Administratively reset a user's password (admin only)
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Param        request      body     admin.ResetPasswordRequest true        "New password details"
// @Success      200          {object} map[string]interface{}                 "Password reset successfully"
// @Failure      400          {object} map[string]interface{}                 "Invalid password requirements"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id}/password [post]
// @Security     Bearer
// @ID           adminResetUserPassword

// HandleGetUserAuditLog godoc
// @Summary      Get user audit log
// @Description  Get detailed audit log for a specific user including all actions performed (admin only)
// @Tags         Admin - Audit
// @Accept       json
// @Produce      json
// @Param        user_id      path     string  true   "User ID"
// @Param        limit        query    int     false  "Maximum number of log entries"        default(100) maximum(500)
// @Param        start_date   query    string  false  "Start date for log entries (RFC3339)" format(date-time)
// @Param        end_date     query    string  false  "End date for log entries (RFC3339)"   format(date-time)
// @Success      200          {object} map[string]interface{}                 "Audit log entries"
// @Failure      401          {object} map[string]interface{}                 "Unauthorized - no valid token"
// @Failure      403          {object} map[string]interface{}                 "Forbidden - admin access required"
// @Failure      404          {object} map[string]interface{}                 "User not found"
// @Failure      500          {object} map[string]interface{}                 "Internal server error"
// @Router       /admin/users/{user_id}/audit [get]
// @Security     Bearer
// @ID           adminGetUserAuditLog
