package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketProxy handles WebSocket connections
type WebSocketProxy struct {
	targetURL *url.URL
	upgrader  websocket.Upgrader
	logger    *zap.Logger
}

// NewWebSocketProxy creates a new WebSocket proxy
func NewWebSocketProxy(targetURL *url.URL, logger *zap.Logger) *WebSocketProxy {
	return &WebSocketProxy{
		targetURL: targetURL,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Configure based on your security requirements
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: logger,
	}
}

// ProxyWebSocket proxies WebSocket connections
func (p *WebSocketProxy) ProxyWebSocket(c *gin.Context) {
	// Build target WebSocket URL
	targetURL := *p.targetURL
	targetURL.Scheme = "ws"
	if p.targetURL.Scheme == "https" {
		targetURL.Scheme = "wss"
	}
	targetURL.Path = c.Request.URL.Path
	targetURL.RawQuery = c.Request.URL.RawQuery

	p.logger.Debug("Proxying WebSocket connection",
		zap.String("target", targetURL.String()))

	// Connect to target
	targetHeader := http.Header{}
	for key, values := range c.Request.Header {
		if key == "Upgrade" || key == "Connection" || key == "Sec-Websocket-Key" ||
			key == "Sec-Websocket-Version" || key == "Sec-Websocket-Extensions" {
			continue
		}
		targetHeader[key] = values
	}

	// Add user context headers
	targetHeader.Set("X-User-ID", c.GetString("user_id"))
	targetHeader.Set("X-Client-ID", c.GetString("client_id"))

	targetConn, resp, err := websocket.DefaultDialer.Dial(targetURL.String(), targetHeader)
	if err != nil {
		p.logger.Error("Failed to connect to target WebSocket", zap.Error(err))
		if resp != nil {
			c.Status(resp.StatusCode)
		} else {
			c.Status(http.StatusBadGateway)
		}
		return
	}
	defer targetConn.Close()

	// Upgrade client connection
	clientConn, err := p.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		p.logger.Error("Failed to upgrade client connection", zap.Error(err))
		return
	}
	defer clientConn.Close()

	// Proxy messages
	errChan := make(chan error, 2)

	// Client to target
	go func() {
		for {
			messageType, data, err := clientConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			if err := targetConn.WriteMessage(messageType, data); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Target to client
	go func() {
		for {
			messageType, data, err := targetConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			if err := clientConn.WriteMessage(messageType, data); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Wait for error
	err = <-errChan
	p.logger.Debug("WebSocket proxy closed", zap.Error(err))
}

// ReverseProxy creates a standard HTTP reverse proxy
func (s *Service) ReverseProxy() *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(s.coreManagerURL)

	// Customize the director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Add user context from gin context if available
		if ginCtx, ok := req.Context().Value("gin_context").(*gin.Context); ok {
			req.Header.Set("X-User-ID", ginCtx.GetString("user_id"))
			req.Header.Set("X-Client-ID", ginCtx.GetString("client_id"))
			req.Header.Set("X-User-Role", ginCtx.GetString("user_role"))
			req.Header.Set("X-User-Tier", ginCtx.GetString("user_tier"))
		}
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		s.logger.Error("Reverse proxy error", zap.Error(err))
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "Service temporarily unavailable"}`))
	}

	return proxy
}
