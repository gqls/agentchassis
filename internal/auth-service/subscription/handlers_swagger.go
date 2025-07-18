package subscription

// NOTE: This file contains swagger annotations for the subscription handlers.
// Run `swag init` to generate the swagger documentation.
// All types are defined in their respective files.

// HandleGetSubscription godoc
// @Summary      Get current subscription
// @Description  Returns the current user's subscription details including tier, status, and expiry dates
// @Tags         Subscription
// @Accept       json
// @Produce      json
// @Success      200 {object} subscription.Subscription "Subscription retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      404 {object} map[string]interface{} "Subscription not found"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /subscription [get]
// @Security     Bearer
// @ID           getSubscription

// HandleGetUsageStats godoc
// @Summary      Get usage statistics
// @Description  Returns usage statistics for the current billing period including personas, projects, and content counts
// @Tags         Subscription
// @Accept       json
// @Produce      json
// @Success      200 {object} subscription.UsageStats "Usage statistics retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Failed to get usage stats"
// @Router       /subscription/usage [get]
// @Security     Bearer
// @ID           getUsageStats

// HandleCheckQuota godoc
// @Summary      Check resource quota
// @Description  Checks if the user has available quota for a specific resource type
// @Tags         Subscription
// @Accept       json
// @Produce      json
// @Param        resource query string true "Resource type to check" Enums(personas,projects,content)
// @Success      200 {object} subscription.QuotaCheckResponse "Quota check result"
// @Failure      400 {object} map[string]interface{} "Resource parameter required"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /subscription/check-quota [get]
// @Security     Bearer
// @ID           checkQuota

// Admin endpoints

// HandleCreateSubscription godoc
// @Summary      Create subscription
// @Description  Creates a new subscription for a user (admin only)
// @Tags         Admin - Subscription
// @Accept       json
// @Produce      json
// @Param        request body subscription.CreateSubscriptionRequest true "Subscription creation details"
// @Success      201 {object} subscription.Subscription "Subscription created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      409 {object} map[string]interface{} "User already has a subscription"
// @Failure      500 {object} map[string]interface{} "Failed to create subscription"
// @Router       /admin/subscriptions [post]
// @Security     Bearer
// @ID           adminCreateSubscription

// HandleUpdateSubscription godoc
// @Summary      Update subscription
// @Description  Updates an existing subscription tier or payment method (admin only)
// @Tags         Admin - Subscription
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Param        request body subscription.UpdateSubscriptionRequest true "Subscription update details"
// @Success      200 {object} subscription.Subscription "Subscription updated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request body"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      404 {object} map[string]interface{} "Subscription not found"
// @Failure      500 {object} map[string]interface{} "Failed to update subscription"
// @Router       /admin/subscriptions/{user_id} [put]
// @Security     Bearer
// @ID           adminUpdateSubscription

// HandleListSubscriptions godoc
// @Summary      List subscriptions
// @Description  Lists all subscriptions with pagination and filtering options (admin only)
// @Tags         Admin - Subscription
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(50) minimum(1) maximum(200)
// @Param        status query string false "Filter by status" Enums(active,trialing,past_due,canceled,expired)
// @Param        tier query string false "Filter by tier" Enums(free,basic,premium,enterprise)
// @Success      200 {object} subscription.SubscriptionListResponse "List of subscriptions retrieved successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure      500 {object} map[string]interface{} "Failed to retrieve subscriptions"
// @Router       /admin/subscriptions [get]
// @Security     Bearer
// @ID           adminListSubscriptions
