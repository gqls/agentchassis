// FILE: internal/auth-service/subscription/handlers.go
package subscription

import (
	"go.uber.org/zap"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handlers wraps the subscription service for HTTP handling
type Handlers struct {
	service *Service
}

// NewHandlers creates new subscription handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// QuotaCheckResponse for quota verification results
type QuotaCheckResponse struct {
	HasQuota     bool   `json:"has_quota" example:"true"`
	Resource     string `json:"resource" example:"personas"`
	CurrentUsage int    `json:"current_usage,omitempty" example:"12"`
	MaxAllowed   int    `json:"max_allowed,omitempty" example:"50"`
	Remaining    int    `json:"remaining,omitempty" example:"38"`
}

// HandleGetSubscription returns the current user's subscription
func (h *Handlers) HandleGetSubscription(c *gin.Context) {
	userID := c.GetString("user_id")

	subscription, err := h.service.GetSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// HandleGetUsageStats returns usage statistics
func (h *Handlers) HandleGetUsageStats(c *gin.Context) {
	userID := c.GetString("user_id")

	stats, err := h.service.GetUsageStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// HandleCheckQuota checks if user has quota for a resource
func (h *Handlers) HandleCheckQuota(c *gin.Context) {
	userID := c.GetString("user_id")
	resource := c.Query("resource")

	if resource == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource parameter required"})
		return
	}

	hasQuota, err := h.service.CheckQuota(c.Request.Context(), userID, resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := QuotaCheckResponse{
		HasQuota: hasQuota,
		Resource: resource,
	}

	c.JSON(http.StatusOK, response)
}

// AdminHandlers for admin operations
type AdminHandlers struct {
	service *Service
	logger  *zap.Logger
}

// NewAdminHandlers creates admin handlers
func NewAdminHandlers(service *Service, logger *zap.Logger) *AdminHandlers {
	return &AdminHandlers{service: service, logger: logger}
}

// SubscriptionListResponse for paginated subscription lists
type SubscriptionListResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
	TotalCount    int            `json:"total_count" example:"156"`
	Page          int            `json:"page" example:"1"`
	Limit         int            `json:"limit" example:"50"`
	TotalPages    int            `json:"total_pages,omitempty" example:"4"`
}

// HandleCreateSubscription creates a subscription (admin only)
func (h *AdminHandlers) HandleCreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.service.CreateSubscription(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// HandleUpdateSubscription updates a subscription (admin only)
func (h *AdminHandlers) HandleUpdateSubscription(c *gin.Context) {
	userID := c.Param("user_id")

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.service.UpdateSubscription(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// HandleListSubscriptions lists all subscriptions with filtering (admin only)
func (h *AdminHandlers) HandleListSubscriptions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit > 200 {
		limit = 200
	}
	if page < 1 {
		page = 1
	}

	params := ListSubscriptionsParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
		Status: c.Query("status"),
		Tier:   c.Query("tier"),
	}

	subscriptions, total, err := h.service.repo.ListAll(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscriptions"})
		return
	}

	totalPages := (total + limit - 1) / limit

	response := SubscriptionListResponse{
		Subscriptions: subscriptions,
		TotalCount:    total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}

	c.JSON(http.StatusOK, response)
}
