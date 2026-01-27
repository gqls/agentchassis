// FILE: platform/orchestration/actions/deploy_image_asset_action.go
// Deploys an image asset: downloads from storage, optimizes, commits to git

package actions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// DeployImageAssetAction downloads an image from storage, optimizes it, and commits via git
// Config:
//   - purpose: image purpose (hero, logo) - determines output path and optimization settings
//   - uri_field: path to storage URI in collected_data (default: {purpose}_uri or {purpose}_result.image_uri)
//   - domain_field: path to domain (default: site_record.domain)
func DeployImageAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "deploy_image_asset"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get purpose (hero, logo, etc.)
	purpose := "hero"
	if p, ok := config["purpose"].(string); ok && p != "" {
		purpose = p
	}

	// Find the storage URI
	storageURI := findStorageURI(params.CollectedData, config, purpose, logger)
	if storageURI == "" {
		logger.Warn("No storage URI found for image asset",
			zap.String("purpose", purpose),
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return map[string]interface{}{
			"deployed": false,
			"skipped":  true,
			"reason":   "no storage URI found for " + purpose,
		}, nil
	}

	// Get domain
	domain := findDomain(params.CollectedData, config)
	if domain == "" {
		return nil, fmt.Errorf("domain not found")
	}

	// Get storage client - it's an interface, need to type assert to *S3Client for image methods
	if params.StorageClient == nil {
		return nil, fmt.Errorf("storage client not available")
	}

	s3Client, ok := params.StorageClient.(*storage.S3Client)
	if !ok {
		return nil, fmt.Errorf("storage client is not S3Client type")
	}

	logger.Info("Deploying image asset",
		zap.String("purpose", purpose),
		zap.String("storage_uri", storageURI),
		zap.String("domain", domain))

	// Download, optimize, and prepare paths using storage package
	processed, err := s3Client.DownloadOptimizeAndPrepare(ctx, storageURI, purpose, logger)
	if err != nil {
		logger.Error("Failed to process image", zap.Error(err))
		return map[string]interface{}{
			"deployed": false,
			"error":    err.Error(),
		}, nil
	}

	logger.Info("Image processed",
		zap.Int("size_bytes", len(processed.Data)),
		zap.String("content_type", processed.ContentType),
		zap.String("output_path", processed.Paths.FilePath))

	// Build files map for git commit (base64 encoded for binary)
	filesMap := map[string]interface{}{
		processed.Paths.FilePath: map[string]interface{}{
			"content":  base64.StdEncoding.EncodeToString(processed.Data),
			"encoding": "base64",
		},
	}

	// Send to git adapter
	result, err := sendGitCommitRequest(ctx, params, domain, filesMap, purpose, logger)
	if err != nil {
		return nil, err
	}

	// Add image URL to result
	result["image_url"] = processed.Paths.RelativeURL
	result["output_path"] = processed.Paths.FilePath
	result["size_bytes"] = len(processed.Data)

	return result, nil
}

// findStorageURI looks for the storage URI in multiple locations
func findStorageURI(collectedData map[string]interface{}, config map[string]interface{}, purpose string, logger *zap.Logger) string {
	// Priority 1: Config-specified field
	if uriField, ok := config["uri_field"].(string); ok && uriField != "" {
		if uri := datahelpers.ExtractNestedFieldString(collectedData, uriField); uri != "" {
			return uri
		}
	}

	// Priority 2: {purpose}_uri (set by StoreAssetAction)
	if uri := datahelpers.ExtractNestedFieldString(collectedData, purpose+"_uri"); uri != "" {
		return uri
	}

	// Priority 3: {purpose}_result.image_uri pattern
	resultField := purpose + "_result.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, resultField); uri != "" {
		return uri
	}

	// Priority 4: {purpose}_stored.asset_url if it's a storage URI
	storedField := purpose + "_stored.asset_url"
	if url := datahelpers.ExtractNestedFieldString(collectedData, storedField); storage.IsS3URI(url) {
		return url
	}

	return ""
}

// findDomain extracts domain from collected data
func findDomain(collectedData map[string]interface{}, config map[string]interface{}) string {
	// Try config-specified field
	domainField := "site_record.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}

	if domain := datahelpers.ExtractNestedFieldString(collectedData, domainField); domain != "" {
		return domain
	}

	// Fallback to input_data.domain
	return datahelpers.ExtractNestedFieldString(collectedData, "input_data.domain")
}

// sendGitCommitRequest sends the commit request to git-adapter
func sendGitCommitRequest(ctx context.Context, params ActionParams, domain string, filesMap map[string]interface{}, purpose string, logger *zap.Logger) (map[string]interface{}, error) {
	if params.Producer == nil {
		return nil, fmt.Errorf("kafka producer not available")
	}

	// Get execution context
	correlationID := params.ExecutionContext.CorrelationID
	orchestrationID := params.ExecutionContext.OrchestrationID
	clientID := params.ExecutionContext.ClientID
	responsesTopic := params.ExecutionContext.ResponsesTopic

	// Fallback for responses topic
	if responsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			responsesTopic = alt
		}
	}
	if responsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			responsesTopic = parent
		}
	}

	newRequestID := uuid.New().String()

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":    correlationID,
			"orchestration_id":  orchestrationID,
			"client_id":         clientID,
			"step_name":         params.ExecutionContext.StepName,
			"request_id":        newRequestID,
			"message_type":      "request",
			"sender_agent_type": params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":   orchestrationID,
			"sender_pod_name":   params.ExecutionContext.Sender.PodName,
			"responses_topic":   responsesTopic,
		},
		"body": map[string]interface{}{
			"action": "commit",
			"data": map[string]interface{}{
				"repo_name":      "sites",
				"domain":         domain,
				"files":          filesMap,
				"commit_message": fmt.Sprintf("Deploy %s image for %s", purpose, domain),
			},
		},
	}

	// Serialize and send
	requestBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal git request: %w", err)
	}

	adapterTopic := "system.adapter.git.requests"
	headers := map[string]string{
		"correlation_id":   correlationID,
		"orchestration_id": orchestrationID,
		"request_id":       newRequestID,
		"message_type":     "request",
	}

	err = params.Producer.Produce(ctx, adapterTopic, headers, []byte(correlationID), requestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send git commit request: %w", err)
	}

	logger.Info("Image deploy request sent to git adapter",
		zap.String("request_id", newRequestID),
		zap.String("domain", domain))

	return map[string]interface{}{
		"deployed":       true,
		"await_response": true,
		"request_id":     newRequestID,
		"purpose":        purpose,
		"domain":         domain,
	}, nil
}
