package subscription

// NOTE: This file contains swagger annotations for the subscription handlers.
// Run `swag init` to generate the swagger documentation.

// HandleGetSubscription godoc
// @Summary      Get current subscription
// @Description  Returns the current user's subscription details
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Success      200 {object} Subscription "Subscription retrieved successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      404 {object} gin.H "Subscription not found"
// @Router       /api/v1/subscription [get]
// @Security     BearerAuth
// @ID           getSubscription

// HandleGetUsageStats godoc
// @Summary      Get usage statistics
// @Description  Returns usage statistics for the current billing period
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Success      200 {object} UsageStats "Usage statistics retrieved successfully"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      500 {object} gin.H "Failed to get usage stats"
// @Router       /api/v1/subscription/usage [get]
// @Security     BearerAuth
// @ID           getUsageStats

// HandleCheckQuota godoc
// @Summary      Check resource quota
// @Description  Checks if the user has quota for a specific resource
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        resource query string true "Resource type to check" Enums(personas, projects, content)
// @Success      200 {object} map[string]interface{} "Quota check result"
// @Failure      400 {object} gin.H "Resource parameter required"
// @Failure      401 {object} gin.H "Unauthorized"
// @Failure      500 {object} gin.H "Internal server error"
// @Router       /api/v1/subscription/check-quota [get]
// @Security     BearerAuth
// @ID           checkQuota
