// FILE: internal/core-manager/admin/dashboard_handlers.go
package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DashboardHandlers provides admin dashboard endpoints
type DashboardHandlers struct {
	clientsDB   *pgxpool.Pool
	templatesDB *pgxpool.Pool
	authDB      *sql.DB // For accessing auth database
	logger      *zap.Logger
}

// NewDashboardHandlers creates new dashboard handlers
func NewDashboardHandlers(clientsDB, templatesDB *pgxpool.Pool, authDB *sql.DB, logger *zap.Logger) *DashboardHandlers {
	return &DashboardHandlers{
		clientsDB:   clientsDB,
		templatesDB: templatesDB,
		authDB:      authDB,
		logger:      logger,
	}
}

// DashboardMetrics represents overall system metrics
type DashboardMetrics struct {
	Overview       OverviewMetrics     `json:"overview"`
	UserMetrics    UserMetrics         `json:"user_metrics"`
	AgentMetrics   AgentMetrics        `json:"agent_metrics"`
	UsageMetrics   UsageMetrics        `json:"usage_metrics"`
	SystemHealth   SystemHealthMetrics `json:"system_health"`
	RecentActivity []ActivityEntry     `json:"recent_activity"`
}

type OverviewMetrics struct {
	TotalClients        int     `json:"total_clients"`
	TotalUsers          int     `json:"total_users"`
	ActiveUsers30Days   int     `json:"active_users_30_days"`
	TotalAgentInstances int     `json:"total_agent_instances"`
	TotalWorkflows      int     `json:"total_workflows"`
	SuccessRate         float64 `json:"success_rate"`
	TotalRevenue        float64 `json:"total_revenue_mtd"`
}

type UserMetrics struct {
	UsersByTier      map[string]int `json:"users_by_tier"`
	NewUsersToday    int            `json:"new_users_today"`
	NewUsersThisWeek int            `json:"new_users_this_week"`
	ChurnRate        float64        `json:"churn_rate_monthly"`
}

type AgentMetrics struct {
	AgentsByType        map[string]int `json:"agents_by_type"`
	MostUsedAgents      []AgentUsage   `json:"most_used_agents"`
	AverageResponseTime float64        `json:"avg_response_time_ms"`
}

type AgentUsage struct {
	AgentType  string `json:"agent_type"`
	UsageCount int    `json:"usage_count"`
}

type UsageMetrics struct {
	TotalFuelConsumed int64            `json:"total_fuel_consumed"`
	FuelByAgentType   map[string]int64 `json:"fuel_by_agent_type"`
	APICallsToday     int              `json:"api_calls_today"`
	StorageUsedGB     float64          `json:"storage_used_gb"`
}

type SystemHealthMetrics struct {
	DatabaseStatus  string  `json:"database_status"`
	KafkaStatus     string  `json:"kafka_status"`
	AverageLatency  float64 `json:"average_latency_ms"`
	ErrorRate       float64 `json:"error_rate_percent"`
	ActiveWorkflows int     `json:"active_workflows"`
	QueueDepth      int     `json:"queue_depth"`
}

type ActivityEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id,omitempty"`
	ClientID    string    `json:"client_id,omitempty"`
}

// HandleGetDashboard returns comprehensive dashboard metrics
func (h *DashboardHandlers) HandleGetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	metrics := &DashboardMetrics{
		Overview:       h.getOverviewMetrics(ctx),
		UserMetrics:    h.getUserMetrics(ctx),
		AgentMetrics:   h.getAgentMetrics(ctx),
		UsageMetrics:   h.getUsageMetrics(ctx),
		SystemHealth:   h.getSystemHealth(ctx),
		RecentActivity: h.getRecentActivity(ctx),
	}

	c.JSON(http.StatusOK, metrics)
}

// getOverviewMetrics collects high-level system metrics
func (h *DashboardHandlers) getOverviewMetrics(ctx context.Context) OverviewMetrics {
	metrics := OverviewMetrics{}

	// Count total clients
	var totalClients int
	err := h.clientsDB.QueryRow(ctx, `
		SELECT COUNT(DISTINCT schema_name) 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'client_%'
	`).Scan(&totalClients)
	if err != nil {
		h.logger.Error("Failed to count clients", zap.Error(err))
	}
	metrics.TotalClients = totalClients

	// Count total users from auth DB
	err = h.authDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE is_active = true
	`).Scan(&metrics.TotalUsers)
	if err != nil {
		h.logger.Error("Failed to count users", zap.Error(err))
	}

	// Count active users in last 30 days
	err = h.authDB.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT user_id) 
		FROM user_activity_logs 
		WHERE created_at > NOW() - INTERVAL 30 DAY
	`).Scan(&metrics.ActiveUsers30Days)
	if err != nil {
		h.logger.Error("Failed to count active users", zap.Error(err))
	}

	// Count total agent instances across all clients
	rows, err := h.clientsDB.Query(ctx, `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'client_%'
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var schemaName string
			if err := rows.Scan(&schemaName); err == nil {
				var count int
				query := fmt.Sprintf("SELECT COUNT(*) FROM %s.agent_instances WHERE is_active = true", schemaName)
				h.clientsDB.QueryRow(ctx, query).Scan(&count)
				metrics.TotalAgentInstances += count
			}
		}
	}

	// Count total workflows and calculate success rate
	var successCount int
	err = h.clientsDB.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'COMPLETED') as success
		FROM orchestrator_state
		WHERE created_at > NOW() - INTERVAL '30 days'
	`).Scan(&metrics.TotalWorkflows, &successCount)
	if err == nil && metrics.TotalWorkflows > 0 {
		metrics.SuccessRate = float64(successCount) / float64(metrics.TotalWorkflows) * 100
	}

	// Calculate total revenue (simplified - would need proper billing integration)
	err = h.authDB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(
			CASE 
				WHEN st.name = 'basic' THEN st.price_monthly
				WHEN st.name = 'premium' THEN st.price_monthly
				WHEN st.name = 'enterprise' THEN st.price_monthly
				ELSE 0
			END
		), 0) as total_revenue
		FROM users u
		JOIN subscriptions s ON u.id = s.user_id
		JOIN subscription_tiers st ON s.tier = st.name
		WHERE u.is_active = true AND s.status = 'active'
	`).Scan(&metrics.TotalRevenue)
	if err != nil {
		h.logger.Error("Failed to calculate revenue", zap.Error(err))
	}

	return metrics
}

// getUserMetrics collects user-related metrics
func (h *DashboardHandlers) getUserMetrics(ctx context.Context) UserMetrics {
	metrics := UserMetrics{
		UsersByTier: make(map[string]int),
	}

	// Count users by subscription tier
	rows, err := h.authDB.QueryContext(ctx, `
		SELECT subscription_tier, COUNT(*) 
		FROM users 
		WHERE is_active = true 
		GROUP BY subscription_tier
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tier string
			var count int
			if err := rows.Scan(&tier, &count); err == nil {
				metrics.UsersByTier[tier] = count
			}
		}
	}

	// Count new users today
	err = h.authDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users 
		WHERE DATE(created_at) = CURDATE()
	`).Scan(&metrics.NewUsersToday)

	// Count new users this week
	err = h.authDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users 
		WHERE created_at > NOW() - INTERVAL 7 DAY
	`).Scan(&metrics.NewUsersThisWeek)

	// Calculate churn rate (simplified)
	var totalLastMonth, churnedThisMonth int
	h.authDB.QueryRowContext(ctx, `
		SELECT 
			(SELECT COUNT(*) FROM users WHERE created_at < DATE_SUB(NOW(), INTERVAL 1 MONTH)) as total_last_month,
			(SELECT COUNT(*) FROM users WHERE is_active = false AND updated_at > DATE_SUB(NOW(), INTERVAL 1 MONTH)) as churned
	`).Scan(&totalLastMonth, &churnedThisMonth)

	if totalLastMonth > 0 {
		metrics.ChurnRate = float64(churnedThisMonth) / float64(totalLastMonth) * 100
	}

	return metrics
}

// getAgentMetrics collects agent-related metrics
func (h *DashboardHandlers) getAgentMetrics(ctx context.Context) AgentMetrics {
	metrics := AgentMetrics{
		AgentsByType: make(map[string]int),
	}

	// Count agents by type
	rows, err := h.clientsDB.Query(ctx, `
		SELECT type, COUNT(*) 
		FROM agent_definitions 
		WHERE is_active = true 
		GROUP BY type
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var agentType string
			var count int
			if err := rows.Scan(&agentType, &count); err == nil {
				metrics.AgentsByType[agentType] = count
			}
		}
	}

	// Get most used agents (from all client schemas)
	// This is simplified - in production you'd aggregate across all client schemas
	mostUsedQuery := `
		SELECT 
			ad.type,
			COUNT(*) as usage_count
		FROM orchestrator_state os
		JOIN agent_definitions ad ON true
		WHERE os.created_at > NOW() - INTERVAL '7 days'
		GROUP BY ad.type
		ORDER BY usage_count DESC
		LIMIT 5
	`

	rows, err = h.clientsDB.Query(ctx, mostUsedQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var usage AgentUsage
			if err := rows.Scan(&usage.AgentType, &usage.UsageCount); err == nil {
				metrics.MostUsedAgents = append(metrics.MostUsedAgents, usage)
			}
		}
	}

	// Calculate average response time (simplified)
	// In production, this would come from your metrics system
	metrics.AverageResponseTime = 245.7 // Mock value

	return metrics
}

// getUsageMetrics collects resource usage metrics
func (h *DashboardHandlers) getUsageMetrics(ctx context.Context) UsageMetrics {
	metrics := UsageMetrics{
		FuelByAgentType: make(map[string]int64),
	}

	// Get total fuel consumed across all clients
	// This would need to aggregate from all client schemas
	metrics.TotalFuelConsumed = 125000 // Mock value

	// Fuel by agent type (mock data)
	metrics.FuelByAgentType = map[string]int64{
		"copywriter":      45000,
		"researcher":      35000,
		"reasoning":       25000,
		"image-generator": 20000,
	}

	// API calls today (mock)
	metrics.APICallsToday = 8543

	// Storage used (would query actual storage metrics)
	metrics.StorageUsedGB = 45.7

	return metrics
}

// getSystemHealth checks current system health
func (h *DashboardHandlers) getSystemHealth(ctx context.Context) SystemHealthMetrics {
	metrics := SystemHealthMetrics{}

	// Check database status
	if err := h.clientsDB.Ping(ctx); err != nil {
		metrics.DatabaseStatus = "unhealthy"
	} else {
		metrics.DatabaseStatus = "healthy"
	}

	// Kafka status (simplified)
	metrics.KafkaStatus = "healthy"

	// Get active workflows count
	h.clientsDB.QueryRow(ctx, `
		SELECT COUNT(*) FROM orchestrator_state 
		WHERE status IN ('RUNNING', 'AWAITING_RESPONSES', 'PAUSED_FOR_HUMAN_INPUT')
	`).Scan(&metrics.ActiveWorkflows)

	// Mock metrics - in production these would come from Prometheus
	metrics.AverageLatency = 123.5
	metrics.ErrorRate = 0.02
	metrics.QueueDepth = 42

	return metrics
}

// getRecentActivity returns recent system activity
func (h *DashboardHandlers) getRecentActivity(ctx context.Context) []ActivityEntry {
	activities := []ActivityEntry{}

	// Get recent user registrations
	rows, err := h.authDB.QueryContext(ctx, `
		SELECT created_at, 'user_registration' as type, email, client_id
		FROM users
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var activity ActivityEntry
			var email string
			rows.Scan(&activity.Timestamp, &activity.Type, &email, &activity.ClientID)
			activity.Description = fmt.Sprintf("New user registered: %s", email)
			activities = append(activities, activity)
		}
	}

	// Get recent workflow completions
	rows, err = h.clientsDB.Query(ctx, `
		SELECT updated_at, status, correlation_id, client_id
		FROM orchestrator_state
		WHERE status IN ('COMPLETED', 'FAILED')
		ORDER BY updated_at DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var activity ActivityEntry
			var status, correlationID string
			rows.Scan(&activity.Timestamp, &status, &correlationID, &activity.ClientID)
			activity.Type = "workflow_" + strings.ToLower(status)
			activity.Description = fmt.Sprintf("Workflow %s: %s", status, correlationID[:8])
			activities = append(activities, activity)
		}
	}

	// Sort by timestamp
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	// Return top 20
	if len(activities) > 20 {
		activities = activities[:20]
	}

	return activities
}

// HandleGetSystemLogs returns recent system logs
func (h *DashboardHandlers) HandleGetSystemLogs(c *gin.Context) {
	service := c.Query("service")
	level := c.Query("level")
	limit := 100

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// In production, this would query your centralized logging system
	// For now, return mock data
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Service:   "auth-service",
			Level:     "info",
			Message:   "User login successful",
			Metadata: map[string]interface{}{
				"user_id": "123",
				"ip":      "192.168.1.1",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
		"query": gin.H{
			"service": service,
			"level":   level,
			"limit":   limit,
		},
	})
}

type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
