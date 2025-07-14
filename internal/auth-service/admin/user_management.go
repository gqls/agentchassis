// FILE: internal/auth-service/admin/user_management.go
package admin

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/auth-service/user"
	"go.uber.org/zap"
)

// UserManagementHandlers provides enhanced user management functionality
type UserManagementHandlers struct {
	userRepo *user.Repository
	db       *sql.DB
	logger   *zap.Logger
}

// NewUserManagementHandlers creates new user management handlers
func NewUserManagementHandlers(userRepo *user.Repository, db *sql.DB, logger *zap.Logger) *UserManagementHandlers {
	return &UserManagementHandlers{
		userRepo: userRepo,
		db:       db,
		logger:   logger,
	}
}

// BulkUserOperation represents a bulk operation on users
type BulkUserOperation struct {
	UserIDs   []string               `json:"user_ids" binding:"required"`
	Operation string                 `json:"operation" binding:"required,oneof=activate deactivate delete upgrade_tier"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Reason    string                 `json:"reason"`
}

// UserExportRequest for exporting user data
type UserExportRequest struct {
	Format  string      `json:"format" binding:"required,oneof=csv json"`
	Filters UserFilters `json:"filters"`
	Fields  []string    `json:"fields"`
}

type UserFilters struct {
	ClientID         string     `json:"client_id"`
	SubscriptionTier string     `json:"subscription_tier"`
	Role             string     `json:"role"`
	IsActive         *bool      `json:"is_active"`
	CreatedAfter     *time.Time `json:"created_after"`
	CreatedBefore    *time.Time `json:"created_before"`
}

// UserImportResult tracks the result of a bulk import
type UserImportResult struct {
	TotalProcessed int      `json:"total_processed"`
	Successful     int      `json:"successful"`
	Failed         int      `json:"failed"`
	Errors         []string `json:"errors,omitempty"`
	UserIDs        []string `json:"created_user_ids"`
}

// HandleBulkUserOperation performs bulk operations on multiple users
func (h *UserManagementHandlers) HandleBulkUserOperation(c *gin.Context) {
	var req BulkUserOperation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the bulk operation
	h.logger.Info("Performing bulk user operation",
		zap.String("operation", req.Operation),
		zap.Int("user_count", len(req.UserIDs)),
		zap.String("admin_id", c.GetString("user_id")))

	results := map[string]interface{}{
		"operation": req.Operation,
		"total":     len(req.UserIDs),
		"succeeded": 0,
		"failed":    0,
		"errors":    []string{},
	}

	for _, userID := range req.UserIDs {
		var err error

		switch req.Operation {
		case "activate":
			err = h.activateUser(c.Request.Context(), userID)
		case "deactivate":
			err = h.deactivateUser(c.Request.Context(), userID, req.Reason)
		case "delete":
			err = h.deleteUser(c.Request.Context(), userID)
		case "upgrade_tier":
			if tier, ok := req.Params["tier"].(string); ok {
				err = h.upgradeTier(c.Request.Context(), userID, tier)
			} else {
				err = fmt.Errorf("tier parameter required for upgrade_tier operation")
			}
		}

		if err != nil {
			results["failed"] = results["failed"].(int) + 1
			errors := results["errors"].([]string)
			results["errors"] = append(errors, fmt.Sprintf("User %s: %v", userID, err))
		} else {
			results["succeeded"] = results["succeeded"].(int) + 1
		}
	}

	// Log activity
	h.logBulkOperation(c.Request.Context(), c.GetString("user_id"), req)

	c.JSON(http.StatusOK, results)
}

// HandleExportUsers exports user data in various formats
func (h *UserManagementHandlers) HandleExportUsers(c *gin.Context) {
	var req UserExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build query with filters
	query, args := h.buildUserExportQuery(req.Filters, req.Fields)

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("Failed to export users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export users"})
		return
	}
	defer rows.Close()

	switch req.Format {
	case "csv":
		h.exportAsCSV(c, rows, req.Fields)
	case "json":
		h.exportAsJSON(c, rows, req.Fields)
	}
}

// HandleImportUsers imports users from CSV
func (h *UserManagementHandlers) HandleImportUsers(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()

	clientID := c.PostForm("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id required"})
		return
	}

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV file"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have header and at least one row"})
		return
	}

	result := UserImportResult{
		TotalProcessed: len(records) - 1, // Excluding header
		Errors:         []string{},
		UserIDs:        []string{},
	}

	// Process each row
	headers := records[0]
	for i, row := range records[1:] {
		userReq := h.parseCSVRow(headers, row, clientID)

		createdUser, err := h.userRepo.CreateUser(c.Request.Context(), userReq)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+2, err))
		} else {
			result.Successful++
			result.UserIDs = append(result.UserIDs, createdUser.ID)
		}
	}

	c.JSON(http.StatusOK, result)
}

// HandleGetUserSessions returns active sessions for a user
func (h *UserManagementHandlers) HandleGetUserSessions(c *gin.Context) {
	userID := c.Param("user_id")

	sessions, err := h.getUserSessions(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// HandleTerminateUserSessions terminates all sessions for a user
func (h *UserManagementHandlers) HandleTerminateUserSessions(c *gin.Context) {
	userID := c.Param("user_id")

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	// Invalidate all tokens for the user
	query := `DELETE FROM auth_tokens WHERE user_id = ?`
	result, err := h.db.ExecContext(c.Request.Context(), query, userID)
	if err != nil {
		h.logger.Error("Failed to terminate sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to terminate sessions"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Log the action
	h.logUserActivity(c.Request.Context(), userID, "sessions_terminated", map[string]interface{}{
		"terminated_by": c.GetString("user_id"),
		"reason":        req.Reason,
		"count":         rowsAffected,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":           "All sessions terminated",
		"sessions_affected": rowsAffected,
	})
}

// HandleResetUserPassword resets a user's password
func (h *UserManagementHandlers) HandleResetUserPassword(c *gin.Context) {
	userID := c.Param("user_id")

	var req struct {
		NewPassword      string `json:"new_password" binding:"required,min=8"`
		RequireChange    bool   `json:"require_change"`
		NotifyUser       bool   `json:"notify_user"`
		NotificationNote string `json:"notification_note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update password
	err := h.userRepo.UpdatePassword(c.Request.Context(), userID, req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to reset password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	// Terminate existing sessions
	h.db.ExecContext(c.Request.Context(), "DELETE FROM auth_tokens WHERE user_id = ?", userID)

	// Log the action
	h.logUserActivity(c.Request.Context(), userID, "password_reset", map[string]interface{}{
		"reset_by":       c.GetString("user_id"),
		"require_change": req.RequireChange,
		"notified":       req.NotifyUser,
	})

	// TODO: Send notification if requested

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
		"user_id": userID,
	})
}

// HandleGetUserAuditLog returns audit log for a user
func (h *UserManagementHandlers) HandleGetUserAuditLog(c *gin.Context) {
	userID := c.Param("user_id")

	// Parse query params
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	startDate := time.Now().AddDate(0, -1, 0) // Default: last month
	if s := c.Query("start_date"); s != "" {
		startDate, _ = time.Parse(time.RFC3339, s)
	}

	endDate := time.Now()
	if e := c.Query("end_date"); e != "" {
		endDate, _ = time.Parse(time.RFC3339, e)
	}

	// Get audit logs
	logs, err := h.getUserAuditLog(c.Request.Context(), userID, startDate, endDate, limit)
	if err != nil {
		h.logger.Error("Failed to get audit log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"logs":    logs,
		"count":   len(logs),
		"period": gin.H{
			"start": startDate,
			"end":   endDate,
		},
	})
}

// Helper methods

func (h *UserManagementHandlers) activateUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET is_active = true, updated_at = NOW() WHERE id = ?`
	_, err := h.db.ExecContext(ctx, query, userID)
	return err
}

func (h *UserManagementHandlers) deactivateUser(ctx context.Context, userID, reason string) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = ?`
	_, err := h.db.ExecContext(ctx, query, userID)

	if err == nil {
		// Terminate all sessions
		h.db.ExecContext(ctx, "DELETE FROM auth_tokens WHERE user_id = ?", userID)

		// Log the deactivation
		h.logUserActivity(ctx, userID, "account_deactivated", map[string]interface{}{
			"reason": reason,
		})
	}

	return err
}

func (h *UserManagementHandlers) deleteUser(ctx context.Context, userID string) error {
	// Soft delete
	return h.userRepo.DeleteUser(ctx, userID)
}

func (h *UserManagementHandlers) upgradeTier(ctx context.Context, userID, tier string) error {
	return h.userRepo.UpdateUserTier(ctx, userID, tier)
}

func (h *UserManagementHandlers) buildUserExportQuery(filters UserFilters, fields []string) (string, []interface{}) {
	// Default fields if none specified
	if len(fields) == 0 {
		fields = []string{"id", "email", "role", "client_id", "subscription_tier", "created_at"}
	}

	// Build SELECT clause
	selectFields := []string{}
	for _, field := range fields {
		// Validate field names to prevent SQL injection
		if isValidField(field) {
			selectFields = append(selectFields, field)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM users WHERE 1=1", strings.Join(selectFields, ", "))
	args := []interface{}{}
	argCount := 0

	// Add filters
	if filters.ClientID != "" {
		argCount++
		query += fmt.Sprintf(" AND client_id = ?")
		args = append(args, filters.ClientID)
	}

	if filters.SubscriptionTier != "" {
		argCount++
		query += fmt.Sprintf(" AND subscription_tier = ?")
		args = append(args, filters.SubscriptionTier)
	}

	if filters.Role != "" {
		argCount++
		query += fmt.Sprintf(" AND role = ?")
		args = append(args, filters.Role)
	}

	if filters.IsActive != nil {
		argCount++
		query += fmt.Sprintf(" AND is_active = ?")
		args = append(args, *filters.IsActive)
	}

	if filters.CreatedAfter != nil {
		argCount++
		query += fmt.Sprintf(" AND created_at > ?")
		args = append(args, *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		argCount++
		query += fmt.Sprintf(" AND created_at < ?")
		args = append(args, *filters.CreatedBefore)
	}

	query += " ORDER BY created_at DESC"

	return query, args
}

func (h *UserManagementHandlers) exportAsCSV(c *gin.Context, rows *sql.Rows, fields []string) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=users_export.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write headers
	writer.Write(fields)

	// Write data
	for rows.Next() {
		values := make([]interface{}, len(fields))
		valuePtrs := make([]interface{}, len(fields))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		record := make([]string, len(fields))
		for i, v := range values {
			record[i] = fmt.Sprintf("%v", v)
		}
		writer.Write(record)
	}
}

func (h *UserManagementHandlers) exportAsJSON(c *gin.Context, rows *sql.Rows, fields []string) {
	users := []map[string]interface{}{}

	for rows.Next() {
		values := make([]interface{}, len(fields))
		valuePtrs := make([]interface{}, len(fields))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		user := make(map[string]interface{})
		for i, field := range fields {
			user[field] = values[i]
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"count":       len(users),
		"exported_at": time.Now(),
	})
}

func (h *UserManagementHandlers) parseCSVRow(headers []string, row []string, clientID string) *user.CreateUserRequest {
	req := &user.CreateUserRequest{
		ClientID: clientID,
	}

	for i, header := range headers {
		if i < len(row) {
			switch strings.ToLower(header) {
			case "email":
				req.Email = row[i]
			case "password":
				req.Password = row[i]
			case "first_name", "firstname":
				req.FirstName = row[i]
			case "last_name", "lastname":
				req.LastName = row[i]
			case "company":
				req.Company = row[i]
			}
		}
	}

	// Generate password if not provided
	if req.Password == "" {
		req.Password = generateRandomPassword()
	}

	return req
}

func (h *UserManagementHandlers) getUserSessions(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	sessions := []map[string]interface{}{}

	query := `
        SELECT id, token_hash, expires_at, created_at
        FROM auth_tokens
        WHERE user_id = ? AND expires_at > NOW()
        ORDER BY created_at DESC
    `

	rows, err := h.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var session struct {
			ID        string
			TokenHash string
			ExpiresAt time.Time
			CreatedAt time.Time
		}

		if err := rows.Scan(&session.ID, &session.TokenHash, &session.ExpiresAt, &session.CreatedAt); err != nil {
			continue
		}

		sessions = append(sessions, map[string]interface{}{
			"id":         session.ID,
			"expires_at": session.ExpiresAt,
			"created_at": session.CreatedAt,
			"is_active":  true,
		})
	}

	return sessions, nil
}

func (h *UserManagementHandlers) getUserAuditLog(ctx context.Context, userID string, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	logs := []map[string]interface{}{}

	query := `
        SELECT id, action, details, ip_address, user_agent, created_at
        FROM user_activity_logs
        WHERE user_id = ? AND created_at BETWEEN ? AND ?
        ORDER BY created_at DESC
        LIMIT ?
    `

	rows, err := h.db.QueryContext(ctx, query, userID, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var log user.UserActivity
		if err := rows.Scan(&log.ID, &log.Action, &log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt); err != nil {
			continue
		}

		logs = append(logs, map[string]interface{}{
			"id":         log.ID,
			"action":     log.Action,
			"details":    log.Details,
			"ip_address": log.IPAddress,
			"user_agent": log.UserAgent,
			"created_at": log.CreatedAt,
		})
	}

	return logs, nil
}

func (h *UserManagementHandlers) logUserActivity(ctx context.Context, userID, action string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)

	activity := &user.UserActivity{
		ID:        uuid.NewString(),
		UserID:    userID,
		Action:    action,
		Details:   string(detailsJSON),
		CreatedAt: time.Now(),
	}

	h.userRepo.LogUserActivity(ctx, activity)
}

func (h *UserManagementHandlers) logBulkOperation(ctx context.Context, adminID string, operation BulkUserOperation) {
	details := map[string]interface{}{
		"operation":  operation.Operation,
		"user_count": len(operation.UserIDs),
		"reason":     operation.Reason,
		"params":     operation.Params,
	}

	h.logUserActivity(ctx, adminID, "bulk_user_operation", details)
}

func isValidField(field string) bool {
	allowedFields := map[string]bool{
		"id":                true,
		"email":             true,
		"role":              true,
		"client_id":         true,
		"subscription_tier": true,
		"is_active":         true,
		"created_at":        true,
		"updated_at":        true,
		"last_login_at":     true,
	}
	return allowedFields[field]
}

func generateRandomPassword() string {
	const (
		length  = 16
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		// Fallback to less secure random generator if crypto/rand fails
		return "fallbackPassword123"
	}

	for i := 0; i < length; i++ {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
