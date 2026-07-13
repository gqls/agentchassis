// FILE: internal/auth-service/gateway/handlers.go
package gateway

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPHandler handles gateway HTTP requests
type HTTPHandler struct {
	service *Service
	logger  *zap.Logger
}

// NewHTTPHandler creates a new gateway HTTP handler
func NewHTTPHandler(service *Service, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		logger:  logger,
	}
}

// ProxyToCoreManager proxies requests to the core manager service
func (h *HTTPHandler) ProxyToCoreManager(c *gin.Context) {
	// Build target URL
	targetPath := c.Param("path")
	if targetPath == "" {
		targetPath = strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
	}

	targetURL := h.service.coreManagerURL.ResolveReference(&url.URL{
		Path:     "/api/v1" + targetPath,
		RawQuery: c.Request.URL.RawQuery,
	})

	h.logger.Debug("Proxying request",
		zap.String("method", c.Request.Method),
		zap.String("target", targetURL.String()))

	// Create new request
	req, err := http.NewRequest(c.Request.Method, targetURL.String(), c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to create proxy request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add user context headers
	req.Header.Set("X-User-ID", c.GetString("user_id"))
	req.Header.Set("X-Client-ID", c.GetString("client_id"))
	req.Header.Set("X-User-Role", c.GetString("user_role"))
	req.Header.Set("X-User-Tier", c.GetString("user_tier"))
	req.Header.Set("X-User-Email", c.GetString("user_email"))

	// Add permissions
	if perms, exists := c.Get("user_permissions"); exists {
		if permissions, ok := perms.([]string); ok {
			req.Header.Set("X-User-Permissions", strings.Join(permissions, ","))
		}
	}

	// Execute request
	resp, err := h.service.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Proxy request failed", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status code
	c.Status(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		h.logger.Error("Failed to copy response body", zap.Error(err))
	}
}

// HandleTemplateRoutes handles all template-related routes
func (h *HTTPHandler) HandleTemplateRoutes(c *gin.Context) {
	h.ProxyToCoreManager(c)
}

// HandleInstanceRoutes handles all instance-related routes
func (h *HTTPHandler) HandleInstanceRoutes(c *gin.Context) {
	h.ProxyToCoreManager(c)
}

// HandleAdminRoutes provides a generic proxy for all admin routes destined for core-manager
func (h *HTTPHandler) HandleAdminRoutes(c *gin.Context) {
	h.ProxyToCoreManager(c)
}

// HandleWebSocket handles WebSocket connections
func (h *HTTPHandler) HandleWebSocket(c *gin.Context) {
	wsProxy := NewWebSocketProxy(h.service.coreManagerURL, h.logger)
	wsProxy.ProxyWebSocket(c)
}
