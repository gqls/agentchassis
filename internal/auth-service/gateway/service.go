package gateway

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gqls/ai-persona-system/platform/config"
	"go.uber.org/zap"
)

// Service handles API gateway functionality
type Service struct {
	coreManagerURL *url.URL
	httpClient     *http.Client
	logger         *zap.Logger
}

// NewService creates a new gateway service
func NewService(cfg *config.ServiceConfig, logger *zap.Logger) *Service {
	coreManagerURL, _ := url.Parse(cfg.Custom["core_manager_url"].(string))

	return &Service{
		coreManagerURL: coreManagerURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetCoreManagerURL returns the core manager URL
func (s *Service) GetCoreManagerURL() *url.URL {
	return s.coreManagerURL
}
