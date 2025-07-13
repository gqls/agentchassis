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

// Template Management Handlers

func (h *HTTPHandler) HandleListTemplates(c *gin.Context) {
	h.proxyToCoreManager(c, "/templates")
}

func (h *HTTPHandler) HandleCreateTemplate(c *gin.Context) {
	h.proxyToCoreManager(c, "/templates")
}

func (h *HTTPHandler) HandleGetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	h.proxyToCoreManager(c, "/templates/"+templateID)
}

func (h *HTTPHandler) HandleUpdateTemplate(c *gin.Context) {
	templateID := c.Param("id")
	h.proxyToCoreManager(c, "/templates/"+templateID)
}

func (h *HTTPHandler) HandleDeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	h.proxyToCoreManager(c, "/templates/"+templateID)
}

// Instance Management Handlers

func (h *HTTPHandler) HandleCreateInstance(c *gin.Context) {
	h.proxyToCoreManager(c, "/personas/instances")
}

func (h *HTTPHandler) HandleListInstances(c *gin.Context) {
	h.proxyToCoreManager(c, "/personas/instances")
}

func (h *HTTPHandler) HandleGetInstance(c *gin.Context) {
	instanceID := c.Param("id")
	h.proxyToCoreManager(c, "/personas/instances/"+instanceID)
}

func (h *HTTPHandler) HandleUpdateInstance(c *gin.Context) {
	instanceID := c.Param("id")
	h.proxyToCoreManager(c, "/personas/instances/"+instanceID)
}

func (h *HTTPHandler) HandleDeleteInstance(c *gin.Context) {
	instanceID := c.Param("id")
	h.proxyToCoreManager(c, "/personas/instances/"+instanceID)
}

// Project Management Handlers

func (h *HTTPHandler) HandleListProjects(c *gin.Context) {
	h.proxyToCoreManager(c, "/projects")
}

func (h *HTTPHandler) HandleCreateProject(c *gin.Context) {
	h.proxyToCoreManager(c, "/projects")
}

func (h *HTTPHandler) HandleGetProject(c *gin.Context) {
	projectID := c.Param("id")
	h.proxyToCoreManager(c, "/projects/"+projectID)
}

func (h *HTTPHandler) HandleUpdateProject(c *gin.Context) {
	projectID := c.Param("id")
	h.proxyToCoreManager(c, "/projects/"+projectID)
}

func (h *HTTPHandler) HandleDeleteProject(c *gin.Context) {
	projectID := c.Param("id")
	h.proxyToCoreManager(c, "/projects/"+projectID)
}

// proxyToCoreManager is a helper method
func (h *HTTPHandler) proxyToCoreManager(c *gin.Context, path string) {
	c.Params = append(c.Params, gin.Param{Key: "path", Value: path})
	h.ProxyToCoreManager(c)
}

// HandleTemplateRoutes handles all template-related routes
func (h *HTTPHandler) HandleTemplateRoutes(c *gin.Context) {
	// Extract the path after /api/v1/templates
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1/templates")

	// Check if it's a specific template ID
	if path != "" && path != "/" {
		templateID := strings.TrimPrefix(path, "/")
		c.Params = append(c.Params, gin.Param{Key: "id", Value: templateID})
	}

	// Proxy to core manager
	h.ProxyToCoreManager(c)
}

// HandleInstanceRoutes handles all instance-related routes
func (h *HTTPHandler) HandleInstanceRoutes(c *gin.Context) {
	// Extract the path after /api/v1/personas/instances
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1/personas/instances")

	// Check if it's a specific instance ID
	if path != "" && path != "/" {
		instanceID := strings.TrimPrefix(path, "/")
		c.Params = append(c.Params, gin.Param{Key: "id", Value: instanceID})
	}

	// Proxy to core manager
	h.ProxyToCoreManager(c)
}

// HandleWebSocket handles WebSocket connections
func (h *HTTPHandler) HandleWebSocket(c *gin.Context) {
	wsProxy := NewWebSocketProxy(h.service.coreManagerURL, h.logger)
	wsProxy.ProxyWebSocket(c)
}
