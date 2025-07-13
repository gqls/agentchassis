package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Repository handles user data access
type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRepository creates a new user repository
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user with profile
func (r *Repository) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:               uuid.New().String(),
		Email:            strings.ToLower(req.Email),
		PasswordHash:     string(hashedPassword),
		Role:             "user", // Default role
		ClientID:         req.ClientID,
		SubscriptionTier: "free", // Default tier
		IsActive:         true,
		EmailVerified:    false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Insert user
	query := `
        INSERT INTO users (id, email, password_hash, role, client_id, subscription_tier, 
                          is_active, email_verified, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err = tx.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Role, user.ClientID,
		user.SubscriptionTier, user.IsActive, user.EmailVerified,
		user.CreatedAt, user.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("user with email %s already exists", req.Email)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create user profile
	profile := &UserProfile{
		UserID:    user.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Company:   req.Company,
	}

	profileQuery := `
        INSERT INTO user_profiles (user_id, first_name, last_name, company)
        VALUES (?, ?, ?, ?)
    `

	_, err = tx.ExecContext(ctx, profileQuery,
		profile.UserID, profile.FirstName, profile.LastName, profile.Company)

	if err != nil {
		return nil, fmt.Errorf("failed to create user profile: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	user.Profile = profile
	r.logger.Info("User created successfully", zap.String("user_id", user.ID))

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(email)

	var user User
	query := `
        SELECT id, email, password_hash, role, client_id, subscription_tier, 
               is_active, email_verified, created_at, updated_at, last_login_at
        FROM users
        WHERE email = ? AND is_active = true
    `

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.ClientID,
		&user.SubscriptionTier, &user.IsActive, &user.EmailVerified,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load profile
	profile, err := r.getUserProfile(ctx, user.ID)
	if err == nil {
		user.Profile = profile
	}

	// Load permissions
	permissions, err := r.getUserPermissions(ctx, user.ID)
	if err == nil {
		user.Permissions = permissions
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	query := `
        SELECT id, email, password_hash, role, client_id, subscription_tier, 
               is_active, email_verified, created_at, updated_at, last_login_at
        FROM users
        WHERE id = ? AND is_active = true
    `

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.ClientID,
		&user.SubscriptionTier, &user.IsActive, &user.EmailVerified,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load profile
	profile, err := r.getUserProfile(ctx, user.ID)
	if err == nil {
		user.Profile = profile
	}

	// Load permissions
	permissions, err := r.getUserPermissions(ctx, user.ID)
	if err == nil {
		user.Permissions = permissions
	}

	return &user, nil
}

// getUserProfile loads user profile
func (r *Repository) getUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	var profile UserProfile
	query := `
        SELECT user_id, first_name, last_name, company, phone, avatar_url, preferences
        FROM user_profiles
        WHERE user_id = ?
    `

	var preferences sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID, &profile.FirstName, &profile.LastName,
		&profile.Company, &profile.Phone, &profile.AvatarURL, &preferences)

	if err != nil {
		return nil, err
	}

	if preferences.Valid {
		json.Unmarshal([]byte(preferences.String), &profile.Preferences)
	}

	return &profile, nil
}

// getUserPermissions loads user permissions
func (r *Repository) getUserPermissions(ctx context.Context, userID string) ([]string, error) {
	query := `
        SELECT p.name
        FROM permissions p
        JOIN user_permissions up ON p.id = up.permission_id
        WHERE up.user_id = ?
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// UpdateUser updates user information
func (r *Repository) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update user table
	userQuery := `UPDATE users SET updated_at = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, userQuery, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Update profile
	var setClauses []string
	var args []interface{}

	if req.FirstName != nil {
		setClauses = append(setClauses, "first_name = ?")
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		setClauses = append(setClauses, "last_name = ?")
		args = append(args, *req.LastName)
	}
	if req.Company != nil {
		setClauses = append(setClauses, "company = ?")
		args = append(args, *req.Company)
	}
	if req.Phone != nil {
		setClauses = append(setClauses, "phone = ?")
		args = append(args, *req.Phone)
	}
	if req.Preferences != nil {
		prefsJSON, _ := json.Marshal(req.Preferences)
		setClauses = append(setClauses, "preferences = ?")
		args = append(args, string(prefsJSON))
	}

	if len(setClauses) > 0 {
		args = append(args, userID)
		profileQuery := fmt.Sprintf(
			"UPDATE user_profiles SET %s WHERE user_id = ?",
			strings.Join(setClauses, ", "),
		)

		_, err = tx.ExecContext(ctx, profileQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ValidatePassword checks if the provided password matches
func (r *Repository) ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword updates user password
func (r *Repository) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.ExecContext(ctx, query, string(hashedPassword), time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *Repository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// UpdateUserTier updates subscription tier
func (r *Repository) UpdateUserTier(ctx context.Context, userID, tier string) error {
	query := `UPDATE users SET subscription_tier = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, tier, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user tier: %w", err)
	}
	return nil
}

// DeleteUser soft deletes a user
func (r *Repository) DeleteUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET is_active = false, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
