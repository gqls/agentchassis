// internal/adapters/webscrape/providers/interfaces.go
package providers

import (
	"context"
	"net/http"

	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// ScrapingProvider interface for different scraping providers
type ScrapingProvider interface {
	Name() string
	IsAvailable() bool
	Scrape(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error)
	Crawl(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error)
	Map(ctx context.Context, url string, config map[string]interface{}) (map[string]interface{}, error)
	ExtractStructured(ctx context.Context, url string, schema map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)
}

// BaseProvider contains common functionality for all providers
type BaseProvider struct {
	httpClient    *http.Client
	storageClient storage.Client
	logger        *zap.Logger
}

func NewBaseProvider(httpClient *http.Client, storageClient storage.Client, logger *zap.Logger) *BaseProvider {
	return &BaseProvider{
		httpClient:    httpClient,
		storageClient: storageClient,
		logger:        logger,
	}
}
