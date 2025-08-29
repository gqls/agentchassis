package agentbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

type BootstrapRequest struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	ClientID  string `json:"client_id"`
}

type BootstrapResponse struct {
	Message string                 `json:"message"`
	Token   string                 `json:"token,omitempty"`  // Optional JWT for future calls
	Config  map[string]interface{} `json:"config,omitempty"` // Optional dynamic config
}

// Bootstrap registers the agent with the core manager
func (a *Agent) Bootstrap() error {
	bootstrapKey := os.Getenv("AGENT_BOOTSTRAP_KEY")
	if bootstrapKey == "" {
		a.logger.Warn("AGENT_BOOTSTRAP_KEY not set, skipping bootstrap")
		return nil // Not fatal - allows local development
	}

	coreManagerURL := os.Getenv("CORE_MANAGER_URL")
	if coreManagerURL == "" {
		coreManagerURL = "http://core-manager.ai-persona-system.svc.cluster.local:8088"
	}

	agentID := os.Getenv("AGENT_ID")
	agentType := os.Getenv("AGENT_TYPE")
	clientID := os.Getenv("CLIENT_ID")

	req := BootstrapRequest{
		AgentID:   agentID,
		AgentType: agentType,
		ClientID:  clientID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal bootstrap request: %w", err)
	}

	httpReq, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/v1/agents/bootstrap", coreManagerURL),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create bootstrap request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Bootstrap-Key", bootstrapKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send bootstrap request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bootstrap failed with status %d", resp.StatusCode)
	}

	var bootstrapResp BootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&bootstrapResp); err != nil {
		return fmt.Errorf("failed to decode bootstrap response: %w", err)
	}

	a.logger.Info("Agent bootstrap successful",
		zap.String("agent_id", agentID),
		zap.String("message", bootstrapResp.Message))

	// Store token if provided for future API calls
	if bootstrapResp.Token != "" {
		os.Setenv("AGENT_JWT_TOKEN", bootstrapResp.Token)
	}

	// Apply any dynamic configuration
	if bootstrapResp.Config != nil {
		// Apply config overrides
		for k, v := range bootstrapResp.Config {
			a.logger.Info("Applying bootstrap config",
				zap.String("key", k),
				zap.Any("value", v))
		}
	}

	// Store the agent configuration for later use
	if bootstrapResp.Config != nil {
		// Store agent type if returned
		if agentType, ok := bootstrapResp.Config["agent_type"].(string); ok {
			a.agentType = agentType
		}

		// Store topics if returned
		if topics, ok := bootstrapResp.Config["topics"].(map[string]interface{}); ok {
			if reqTopic, ok := topics["requests"].(string); ok {
				a.requestTopic = reqTopic
			}
			if respTopic, ok := topics["responses"].(string); ok {
				a.responseTopic = respTopic
			}
		}
	}

	return nil

	return nil
}
