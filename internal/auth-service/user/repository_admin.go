// FILE: internal/auth-service/user/repository_admin.go
package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ListUsersParams contains parameters for listing users
type ListUsersParams struct {
	Offset    int
	Limit     int
	Email     string
	ClientID  string
	Role      string
	Tier      string
	IsActive  *bool
	SortBy    string
	SortOrder string
}

// UserStats contains statistics about a user
type UserStats struct {
	TotalProjects int        `json:"total_projects"`
	TotalPersonas int        `json:"total_personas"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	AccountAge    string     `json:"account_age"`
	TotalLogins   int        `json:"total_logins"`
}

// UserActivity represents a user activity log entry
type UserActivity struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// AdminUpdateRequest contains fields that can be updated by admin
type AdminUpdateRequest struct {
	Role             *string
	SubscriptionTier *string
	IsActive         *bool
	EmailVerified    *bool
}

// ListUsers returns a paginated list of users with optional filtering
func (r *Repository) ListUsers(ctx context.Context, params ListUsersParams) ([]User, int, error) {
	// Build the query dynamically
	query := `
		SELECT u.id, u.email, u.password_hash, u.role, u.client_id, 
		       u.subscription_tier, u.is_active, u.email_verified,
		       u.created_at, u.updated_at, u.last_login_at
		FROM users u
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users u WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Add filters
	var conditions []string

	if params.Email != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("u.email ILIKE $%d", argCount))
		args = append(args, "%"+params.Email+"%")
	}

	if params.ClientID != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("u.client_id = $%d", argCount))
		args = append(args, params.ClientID)
	}

	if params.Role != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("u.role = $%d", argCount))
		args = append(args, params.Role)
	}

	if params.Tier != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("u.subscription_tier = $%d", argCount))
		args = append(args, params.Tier)
	}

	if params.IsActive != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("u.is_active = $%d", argCount))
		args = append(args, *params.IsActive)
	}

	// Add conditions to queries
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Get total count
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Add sorting
	validSortColumns := map[string]bool{
		"email": true, "created_at": true, "updated_at": true,
		"last_login_at": true, "role": true, "subscription_tier": true,
	}

	sortBy := "created_at"
	if validSortColumns[params.SortBy] {
		sortBy = params.SortBy
	}

	sortOrder := "DESC"
	if strings.ToUpper(params.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY u.%s %s", sortBy, sortOrder)

	// Add pagination
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, params.Limit)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, params.Offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.Role,
			&user.ClientID, &user.SubscriptionTier, &user.IsActive,
			&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan user row", zap.Error(err))
			continue
		}

		// Don't load profile and permissions for list view (performance)
		users = append(users, user)
	}

	return users, totalCount, nil
}

// GetUserStats retrieves statistics for a user
func (r *Repository) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	stats := &UserStats{}

	// Get basic user info for last login and account age
	var createdAt time.Time
	var lastLogin sql.NullTime
	err := r.db.QueryRowContext(ctx,
		"SELECT created_at, last_login_at FROM users WHERE id = $1",
		userID,
	).Scan(&createdAt, &lastLogin)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		stats.LastLoginAt = &lastLogin.Time
	}

	// Calculate account age
	age := time.Since(createdAt)
	if age.Hours() < 24 {
		stats.AccountAge = fmt.Sprintf("%d hours", int(age.Hours()))
	} else if age.Hours() < 24*30 {
		stats.AccountAge = fmt.Sprintf("%d days", int(age.Hours()/24))
	} else {
		stats.AccountAge = fmt.Sprintf("%d months", int(age.Hours()/(24*30)))
	}

	// Get project count
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM projects WHERE owner_id = $1 AND is_active = true",
		userID,
	).Scan(&stats.TotalProjects)

	if err != nil && err != sql.ErrNoRows {
		r.logger.Warn("Failed to get project count", zap.Error(err))
	}

	// Note: Persona count would require access to clients DB
	// For now, we'll leave it at 0

	return stats, nil
}

// AdminUpdateUser updates user fields that only admins can change
func (r *Repository) AdminUpdateUser(ctx context.Context, userID string, req *AdminUpdateRequest) error {
	var setClauses []string
	var args []interface{}
	argCount := 0

	if req.Role != nil {
		argCount++
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argCount))
		args = append(args, *req.Role)
	}

	if req.SubscriptionTier != nil {
		argCount++
		setClauses = append(setClauses, fmt.Sprintf("subscription_tier = $%d", argCount))
		args = append(args, *req.SubscriptionTier)
	}

	if req.IsActive != nil {
		argCount++
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *req.IsActive)
	}

	if req.EmailVerified != nil {
		argCount++
		setClauses = append(setClauses, fmt.Sprintf("email_verified = $%d", argCount))
		args = append(args, *req.EmailVerified)
	}

	if len(setClauses) == 0 {
		return nil // Nothing to update
	}

	argCount++
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())

	argCount++
	args = append(args, userID)

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		argCount,
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetUserActivityLog retrieves activity logs for a user
func (r *Repository) GetUserActivityLog(ctx context.Context, userID string, limit, offset int) ([]UserActivity, error) {
	// First, ensure activity log table exists
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS user_activity_logs (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			action VARCHAR(100) NOT NULL,
			details TEXT,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			INDEX idx_activity_user_created (user_id, created_at DESC),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`

	_, err := r.db.ExecContext(ctx, createTableQuery)
	if err != nil {
		r.logger.Warn("Failed to ensure activity log table", zap.Error(err))
	}

	// Get activity logs
	query := `
		SELECT id, user_id, action, COALESCE(details, ''), 
		       COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM user_activity_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}
	defer rows.Close()

	var activities []UserActivity
	for rows.Next() {
		var activity UserActivity
		err := rows.Scan(
			&activity.ID, &activity.UserID, &activity.Action,
			&activity.Details, &activity.IPAddress, &activity.UserAgent,
			&activity.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan activity row", zap.Error(err))
			continue
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// GrantPermission grants a permission to a user
func (r *Repository) GrantPermission(ctx context.Context, userID, permissionName string) error {
	// First get the permission ID
	var permissionID string
	err := r.db.QueryRowContext(ctx,
		"SELECT id FROM permissions WHERE name = ?",
		permissionName,
	).Scan(&permissionID)

	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Grant the permission
	query := `
		INSERT INTO user_permissions (user_id, permission_id, granted_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE granted_at = VALUES(granted_at)
	`

	_, err = r.db.ExecContext(ctx, query, userID, permissionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	return nil
}

// RevokePermission revokes a permission from a user
func (r *Repository) RevokePermission(ctx context.Context, userID, permissionName string) error {
	query := `
		DELETE up FROM user_permissions up
		JOIN permissions p ON up.permission_id = p.id
		WHERE up.user_id = ? AND p.name = ?
	`

	_, err := r.db.ExecContext(ctx, query, userID, permissionName)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	return nil
}

// LogUserActivity logs a user action
func (r *Repository) LogUserActivity(ctx context.Context, activity *UserActivity) error {
	query := `
		INSERT INTO user_activity_logs 
		(id, user_id, action, details, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		activity.ID, activity.UserID, activity.Action,
		activity.Details, activity.IPAddress, activity.UserAgent,
		activity.CreatedAt,
	)

	return err
}
