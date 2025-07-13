// FILE: internal/auth-service/subscription/handlers.go
package subscription

import (
	"net/http"

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

	c.JSON(http.StatusOK, gin.H{
		"has_quota": hasQuota,
		"resource":  resource,
	})
}

// AdminHandlers for admin operations
type AdminHandlers struct {
	service *Service
}

// NewAdminHandlers creates admin handlers
func NewAdminHandlers(service *Service) *AdminHandlers {
	return &AdminHandlers{service: service}
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
