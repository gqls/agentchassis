// FILE: internal/agents/contentcreator/agent.go
package contentcreator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/database"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/memory"
	"github.com/gqls/agentchassis/platform/observability"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	AgentType     = "content-creator"
	RequestTopic  = "system.agent.content-creator.process"
	ResponseTopic = "system.responses.content-creator"
	ConsumerGroup = "content-creator-agent-group"
)

// RequestPayload defines the enhanced data structure for content generation
type RequestPayload struct {
	Action string `json:"action"` // e.g., "generate_blog_post", "generate_product_description", "generate_social_media"
	Data   struct {
		// Core content specifications
		Topic       string   `json:"topic"`              // Main topic or title
		Prompt      string   `json:"prompt,omitempty"`   // Additional prompt/instructions
		Keywords    []string `json:"keywords,omitempty"` // SEO keywords to include
		ContentType string   `json:"content_type"`       // blog_post, product_description, social_media, email, etc.

		// Style and format
		Length string `json:"length,omitempty"` // short, medium, long
		Style  string `json:"style,omitempty"`  // informative, persuasive, casual, professional
		Format string `json:"format,omitempty"` // markdown, html, plain_text
		Tone   string `json:"tone,omitempty"`   // friendly, formal, conversational, authoritative

		// Additional context
		Context        map[string]interface{} `json:"context,omitempty"`         // Additional context for the LLM
		TargetAudience string                 `json:"target_audience,omitempty"` // Who this content is for

		// Platform-specific
		Platform  string `json:"platform,omitempty"`   // For social media: twitter, linkedin, facebook
		MaxLength int    `json:"max_length,omitempty"` // Character/word limit
	} `json:"data"`
}

// ResponsePayload defines the enhanced response format
type ResponsePayload struct {
	Success       bool                   `json:"success"`
	GeneratedText string                 `json:"generated_text"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Agent is the content writing specialist with memory support
type Agent struct {
	ctx           context.Context
	logger        *zap.Logger
	consumer      *kafka.Consumer
	producer      kafka.Producer
	aiClient      aiservice.AIService
	fuelManager   *governance.FuelManager
	memoryService *memory.Service
	db            *pgxpool.Pool
	configLoader  *config.AgentConfigLoader
	metrics       *observability.SafeAIMetrics
	metricsConfig struct {
		Enabled        bool
		FailSilently   bool
		DetailedErrors bool
	}
}

// NewAgent creates a new content creator agent with optional memory support
func NewAgent(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error) {
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, RequestTopic, ConsumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Initialize AI client from custom config
	aiConfig, ok := cfg.Custom["ai_service"].(map[string]interface{})
	if !ok {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("ai_service configuration missing in custom config")
	}

	aiClient, err := aiservice.NewAnthropicClient(ctx, aiConfig)
	if err != nil {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	agent := &Agent{
		ctx:          ctx,
		logger:       logger,
		consumer:     consumer,
		producer:     producer,
		aiClient:     aiClient,
		fuelManager:  governance.NewFuelManager(),
		configLoader: config.NewAgentConfigLoader(logger),
		metrics:      observability.NewSafeAIMetrics(logger),
	}

	// Initialize database and memory service if configured
	if cfg.Infrastructure.ClientsDatabase.Host != "" {
		db, err := database.NewPostgresConnection(ctx, cfg.Infrastructure.ClientsDatabase, logger)
		if err != nil {
			logger.Warn("Failed to connect to database, running without memory support", zap.Error(err))
		} else {
			agent.db = db
			agent.memoryService = memory.NewService(db, aiClient, logger)
			logger.Info("Memory service initialized successfully")
		}
	}

	// Load metrics config
	if metricsConfig, ok := cfg.Custom["metrics"].(map[string]interface{}); ok {
		if enabled, ok := metricsConfig["enabled"].(bool); ok {
			agent.metricsConfig.Enabled = enabled
		} else {
			agent.metricsConfig.Enabled = true // default to enabled
		}

		if failSilently, ok := metricsConfig["fail_silently"].(bool); ok {
			agent.metricsConfig.FailSilently = failSilently
		} else {
			agent.metricsConfig.FailSilently = true // default to fail silently
		}
	} else {
		// Default configuration
		agent.metricsConfig.Enabled = true
		agent.metricsConfig.FailSilently = true
	}

	return agent, nil
}

// Run starts the agent's main loop
func (a *Agent) Run() error {
	a.logger.Info("Content Creator Agent is running and waiting for tasks...")

	for {
		select {
		case <-a.ctx.Done():
			a.consumer.Close()
			a.producer.Close()
			if a.db != nil {
				a.db.Close()
			}
			return nil
		default:
			msg, err := a.consumer.FetchMessage(a.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				observability.SystemErrors.WithLabelValues(AgentType, "fetch_message").Inc()
				time.Sleep(1 * time.Second)
				continue
			}

			observability.KafkaMessagesConsumed.WithLabelValues(msg.Topic, ConsumerGroup).Inc()
			go a.handleMessage(msg)
		}
	}
}

// handleMessage processes a single content generation request with memory support
func (a *Agent) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("request_id", headers["request_id"]),
		zap.String("client_id", headers["client_id"]),
		zap.String("agent_instance_id", headers["agent_instance_id"]),
	)

	observability.AgentTasksReceived.WithLabelValues(AgentType, headers["action"]).Inc()
	startTime := time.Now()
	defer func() {
		observability.AgentProcessingDuration.WithLabelValues(AgentType, headers["action"]).
			Observe(time.Since(startTime).Seconds())
	}()

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request payload", zap.Error(err))
		a.sendErrorResponse(headers, errors.ValidationError("payload", "invalid JSON"))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Log request details
	l.Info("Processing content creation request",
		zap.String("action", req.Action),
		zap.String("content_type", req.Data.ContentType),
		zap.String("topic", req.Data.Topic),
		zap.String("style", req.Data.Style),
		zap.String("length", req.Data.Length),
	)

	// --- Fuel Check & Deduction ---
	initialFuel, err := governance.GetFuelFromHeader(headers)
	if err != nil {
		l.Error("Missing or invalid fuel header", zap.Error(err))
		a.sendErrorResponse(headers, errors.New(errors.ErrInternal, "Missing or invalid fuel header").WithCause(err).Build())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Determine cost based on content length and model
	actionCost := a.determineFuelCost(req)
	if !a.fuelManager.HasEnoughFuel(initialFuel, a.getFuelActionName(req)) {
		l.Warn("Insufficient fuel for operation",
			zap.Int("current_fuel", initialFuel),
			zap.Int("required_fuel", actionCost))
		a.sendErrorResponse(headers, errors.InsufficientFuel(actionCost, initialFuel, req.Action))
		observability.FuelExhausted.WithLabelValues(AgentType, req.Action, headers["client_id"]).Inc()
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	remainingFuel := a.fuelManager.DeductFuel(initialFuel, a.getFuelActionName(req))
	governance.SetFuelHeader(headers, remainingFuel)
	observability.FuelConsumed.WithLabelValues(AgentType, req.Action, headers["client_id"]).Add(float64(actionCost))

	// --- Load Agent Configuration ---
	var agentConfig *models.AgentConfig
	if a.db != nil && headers["client_id"] != "" && headers["agent_instance_id"] != "" {
		agentConfig, err = a.configLoader.LoadFromDatabase(a.ctx, a.db, headers["client_id"],
			headers["agent_instance_id"], AgentType)
		if err != nil {
			l.Warn("Failed to load agent configuration, using defaults", zap.Error(err))
			agentConfig = a.configLoader.GetDefaultConfig(headers["agent_instance_id"], AgentType)
		}
	}

	// --- Memory Retrieval ---
	var relevantMemories []models.MemoryEntry
	if a.shouldUseMemory(agentConfig) {
		agentID, _ := uuid.Parse(headers["agent_instance_id"])
		query := a.buildMemoryQuery(req)

		memories, err := a.memoryService.RetrieveRelevantMemories(
			a.ctx,
			agentID,
			query,
			a.getMemoryRetrievalCount(agentConfig),
		)
		if err != nil {
			l.Warn("Failed to retrieve memories", zap.Error(err))
		} else {
			relevantMemories = memories
			l.Info("Retrieved relevant memories", zap.Int("count", len(memories)))
		}
	}

	// --- Build Prompt with Context ---
	prompt := a.buildEnhancedContentPrompt(req, relevantMemories, agentConfig)

	// --- Generate Content ---
	options := a.getGenerationOptions(req, agentConfig)
	l.Info("Calling AI service for content generation",
		zap.Any("options", options),
		zap.Int("memory_context_count", len(relevantMemories)))

	result, err := a.aiClient.GenerateText(a.ctx, prompt, options)
	if err != nil {
		l.Error("AI content generation failed", zap.Error(err))
		a.sendErrorResponse(headers, errors.New(errors.ErrAIServiceError, "Failed to generate content").WithCause(err).Build())
		observability.AgentTasksProcessed.WithLabelValues(AgentType, req.Action, "ai_error").Inc()
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// --- Token Counting & Metrics ---
	tokensUsed := a.estimateTokens(result)
	estimatedCost := a.estimateCost(tokensUsed, options["model"].(string))
	a.metrics.RecordAIServiceMetrics(a.aiClient, "text_generation", tokensUsed, "output")

	// --- Store Memory ---
	if a.shouldStoreMemory(agentConfig, req) {
		agentID, _ := uuid.Parse(headers["agent_instance_id"])
		memoryEntry := models.MemoryEntry{
			Type:    "generated_content",
			Content: a.createMemoryContent(req, result),
			Metadata: map[string]interface{}{
				"topic":           req.Data.Topic,
				"content_type":    req.Data.ContentType,
				"keywords":        req.Data.Keywords,
				"style":           req.Data.Style,
				"length":          req.Data.Length,
				"platform":        req.Data.Platform,
				"word_count":      len(strings.Fields(result)),
				"character_count": len(result),
				"request_id":      headers["request_id"],
			},
			Timestamp: time.Now(),
		}

		if err := a.memoryService.StoreMemory(a.ctx, agentID, memoryEntry); err != nil {
			l.Warn("Failed to store memory", zap.Error(err))
		} else {
			observability.MemoriesStored.WithLabelValues(AgentType, "generated_content").Inc()
		}
	}

	// --- Prepare Response ---
	responsePayload := ResponsePayload{
		Success:       true,
		GeneratedText: result,
		Metadata: map[string]interface{}{
			"tokens_used":     tokensUsed,
			"estimated_cost":  estimatedCost,
			"remaining_fuel":  remainingFuel,
			"content_type":    req.Data.ContentType,
			"word_count":      len(strings.Fields(result)),
			"character_count": len(result),
			"memories_used":   len(relevantMemories),
			"model_used":      options["model"],
			"generation_time": time.Since(startTime).Seconds(),
		},
	}

	a.sendSuccessResponse(headers, responsePayload)
	observability.AgentTasksProcessed.WithLabelValues(AgentType, req.Action, "success").Inc()

	// Commit message
	a.consumer.CommitMessages(context.Background(), msg)
	l.Info("Content generation completed successfully",
		zap.Int("tokens_used", tokensUsed),
		zap.Float64("generation_time", time.Since(startTime).Seconds()))

	if a.metricsConfig.Enabled {
		observability.RecordAIServiceMetricsSafe(a.logger, a.aiClient, "text_generation", tokensUsed, "output")
	}
}

// buildEnhancedContentPrompt creates a sophisticated prompt based on request type and memories
func (a *Agent) buildEnhancedContentPrompt(req RequestPayload, memories []models.MemoryEntry, config *models.AgentConfig) string {
	var prompt strings.Builder

	// Add role and context
	prompt.WriteString("You are a professional content creator with expertise in writing ")
	prompt.WriteString(a.getContentTypeDescription(req.Data.ContentType))
	prompt.WriteString(".\n\n")

	// Add memory context if available
	if len(memories) > 0 {
		prompt.WriteString("## Previous Related Content (for context and consistency):\n")
		for i, memory := range memories {
			metadata := memory.Metadata
			prompt.WriteString(fmt.Sprintf("%d. Topic: %v, Style: %v\n", i+1,
				metadata["topic"], metadata["style"]))
			// Include a snippet of previous content
			contentPreview := memory.Content
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			prompt.WriteString(fmt.Sprintf("   Preview: %s\n\n", contentPreview))
		}
	}

	// Main instruction
	prompt.WriteString("## Task:\n")
	prompt.WriteString(fmt.Sprintf("Create %s content with the following specifications:\n\n", req.Data.ContentType))

	// Core requirements
	prompt.WriteString(fmt.Sprintf("**Topic**: %s\n", req.Data.Topic))
	if req.Data.Prompt != "" {
		prompt.WriteString(fmt.Sprintf("**Additional Instructions**: %s\n", req.Data.Prompt))
	}

	// Content specifications
	if req.Data.ContentType != "" {
		prompt.WriteString(fmt.Sprintf("**Content Type**: %s\n", req.Data.ContentType))
	}
	if req.Data.Style != "" {
		prompt.WriteString(fmt.Sprintf("**Writing Style**: %s\n", req.Data.Style))
	}
	if req.Data.Tone != "" {
		prompt.WriteString(fmt.Sprintf("**Tone**: %s\n", req.Data.Tone))
	}
	if req.Data.Length != "" {
		prompt.WriteString(fmt.Sprintf("**Length**: %s\n", a.getLengthDescription(req.Data.Length)))
	}
	if req.Data.Format != "" {
		prompt.WriteString(fmt.Sprintf("**Format**: %s\n", req.Data.Format))
	}

	// Keywords
	if len(req.Data.Keywords) > 0 {
		prompt.WriteString(fmt.Sprintf("**Keywords to Include**: %s\n", strings.Join(req.Data.Keywords, ", ")))
	}

	// Audience
	if req.Data.TargetAudience != "" {
		prompt.WriteString(fmt.Sprintf("**Target Audience**: %s\n", req.Data.TargetAudience))
	}

	// Platform-specific requirements
	if req.Data.Platform != "" {
		prompt.WriteString(fmt.Sprintf("**Platform**: %s\n", req.Data.Platform))
		prompt.WriteString(a.getPlatformGuidelines(req.Data.Platform))
	}

	// Length constraints
	if req.Data.MaxLength > 0 {
		prompt.WriteString(fmt.Sprintf("**Maximum Length**: %d characters\n", req.Data.MaxLength))
	}

	// Additional context
	if len(req.Data.Context) > 0 {
		contextBytes, _ := json.MarshalIndent(req.Data.Context, "", "  ")
		prompt.WriteString(fmt.Sprintf("\n**Additional Context**:\n%s\n", string(contextBytes)))
	}

	// Content-type specific instructions
	prompt.WriteString("\n## Specific Requirements:\n")
	prompt.WriteString(a.getContentTypeInstructions(req.Data.ContentType, req.Data.Style))

	// Final instruction
	prompt.WriteString("\n## Output:\nGenerate the content now, ensuring it meets all specified requirements:")

	return prompt.String()
}

// Helper methods for prompt building
func (a *Agent) getContentTypeDescription(contentType string) string {
	descriptions := map[string]string{
		"blog_post":           "engaging blog posts that inform and captivate readers",
		"product_description": "compelling product descriptions that highlight benefits and drive conversions",
		"social_media":        "viral-worthy social media content that sparks engagement",
		"email":               "persuasive emails that get opened and drive action",
		"landing_page":        "high-converting landing page copy that captures leads",
		"press_release":       "newsworthy press releases that get media attention",
		"technical_doc":       "clear technical documentation that simplifies complex topics",
	}

	if desc, ok := descriptions[contentType]; ok {
		return desc
	}
	return "various types of content"
}

func (a *Agent) getContentTypeInstructions(contentType, style string) string {
	instructions := map[string]string{
		"blog_post": `- Start with a compelling hook that grabs attention
- Use clear headers and subheadings for structure
- Include relevant examples and data points
- End with a strong conclusion and call-to-action
- Optimize for SEO without sacrificing readability`,

		"product_description": `- Lead with the primary benefit
- Use sensory language to help customers visualize the product
- Address common pain points and objections
- Include social proof elements where relevant
- End with a clear value proposition`,

		"social_media": `- Start with an attention-grabbing first line
- Keep it concise and scannable
- Include relevant hashtags naturally
- Use emojis appropriately for the platform
- Include a clear call-to-action`,

		"email": `- Write a compelling subject line (include this first)
- Open with personalization and relevance
- Keep paragraphs short and scannable
- Use bullet points for key benefits
- Include one clear call-to-action`,
	}

	if inst, ok := instructions[contentType]; ok {
		return inst
	}
	return "- Follow best practices for this content type\n- Ensure clarity and engagement"
}

func (a *Agent) getLengthDescription(length string) string {
	descriptions := map[string]string{
		"short":  "Brief and concise (100-300 words)",
		"medium": "Moderate length with good detail (300-800 words)",
		"long":   "Comprehensive and in-depth (800-1500 words)",
	}

	if desc, ok := descriptions[length]; ok {
		return desc
	}
	return length
}

func (a *Agent) getPlatformGuidelines(platform string) string {
	guidelines := map[string]string{
		"twitter":   "- Maximum 280 characters\n- Use 1-2 relevant hashtags\n- Make it retweetable\n",
		"linkedin":  "- Professional tone\n- 1300 character limit for best visibility\n- Include relevant industry hashtags\n",
		"facebook":  "- Conversational and engaging\n- Optimal length 40-80 characters\n- Ask questions to boost engagement\n",
		"instagram": "- Maximum 2200 characters\n- Include up to 30 hashtags\n- Use line breaks for readability\n",
	}

	if guide, ok := guidelines[platform]; ok {
		return guide
	}
	return ""
}

// Memory-related helper methods
func (a *Agent) shouldUseMemory(config *models.AgentConfig) bool {
	return a.memoryService != nil && config != nil && config.MemoryConfig.Enabled
}

func (a *Agent) shouldStoreMemory(config *models.AgentConfig, req RequestPayload) bool {
	if !a.shouldUseMemory(config) {
		return false
	}

	// Store all generated content by default, but could add logic to filter
	// based on content type, quality score, etc.
	return config.MemoryConfig.AutoStore
}

func (a *Agent) getMemoryRetrievalCount(config *models.AgentConfig) int {
	if config != nil && config.MemoryConfig.RetrievalCount > 0 {
		return config.MemoryConfig.RetrievalCount
	}
	return 5 // default
}

func (a *Agent) buildMemoryQuery(req RequestPayload) string {
	// Build a query that will help find relevant past content
	parts := []string{req.Data.Topic}

	if req.Data.ContentType != "" {
		parts = append(parts, req.Data.ContentType)
	}

	if len(req.Data.Keywords) > 0 {
		parts = append(parts, strings.Join(req.Data.Keywords, " "))
	}

	if req.Data.TargetAudience != "" {
		parts = append(parts, req.Data.TargetAudience)
	}

	return strings.Join(parts, " ")
}

func (a *Agent) createMemoryContent(req RequestPayload, generatedText string) string {
	// Create a structured memory entry that captures both request and response
	memory := map[string]interface{}{
		"request": map[string]interface{}{
			"topic":        req.Data.Topic,
			"content_type": req.Data.ContentType,
			"style":        req.Data.Style,
			"keywords":     req.Data.Keywords,
		},
		"response": map[string]interface{}{
			"preview": generatedText,
			"length":  len(generatedText),
		},
	}

	memoryJSON, _ := json.Marshal(memory)
	return string(memoryJSON)
}

// Cost and options methods
func (a *Agent) determineFuelCost(req RequestPayload) int {
	// Base cost on content length and complexity
	baseCost := a.fuelManager.GetCost("ai_text_generate_claude_sonnet")

	// Adjust based on length
	switch req.Data.Length {
	case "long":
		baseCost = int(float64(baseCost) * 2.0)
	case "medium":
		baseCost = int(float64(baseCost) * 1.5)
	}

	// Add cost for complex content types
	complexTypes := map[string]bool{
		"technical_doc": true,
		"landing_page":  true,
		"press_release": true,
	}

	if complexTypes[req.Data.ContentType] {
		baseCost = int(float64(baseCost) * 1.3)
	}

	return baseCost
}

func (a *Agent) getFuelActionName(req RequestPayload) string {
	// Map to standard fuel action names
	switch req.Data.Length {
	case "long":
		return "ai_text_generate_claude_opus" // Use more powerful model for long content
	default:
		return "ai_text_generate_claude_sonnet"
	}
}

func (a *Agent) getGenerationOptions(req RequestPayload, config *models.AgentConfig) map[string]interface{} {
	options := make(map[string]interface{})

	// Set defaults
	options["temperature"] = 0.7
	options["max_tokens"] = 2000
	options["model"] = "claude-3-5-sonnet-20241022"

	// Override with agent config if available
	if config != nil && config.CoreLogic != nil {
		if temp, ok := config.CoreLogic["temperature"].(float64); ok {
			options["temperature"] = temp
		}
		if maxTokens, ok := config.CoreLogic["max_tokens"].(float64); ok {
			options["max_tokens"] = int(maxTokens)
		}
		if model, ok := config.CoreLogic["model"].(string); ok {
			options["model"] = model
		}
	}

	// Adjust based on content type and style
	switch req.Data.ContentType {
	case "technical_doc":
		options["temperature"] = 0.3 // Lower temperature for accuracy
	case "social_media":
		options["temperature"] = 0.9 // Higher temperature for creativity
	}

	switch req.Data.Style {
	case "creative", "casual":
		options["temperature"] = options["temperature"].(float64) + 0.1
	case "formal", "professional":
		options["temperature"] = options["temperature"].(float64) - 0.1
	}

	// Adjust max tokens based on length
	switch req.Data.Length {
	case "short":
		options["max_tokens"] = 500
	case "long":
		options["max_tokens"] = 4000
		options["model"] = "claude-3-opus-20240229" // Use larger model for long content
	}

	// Platform-specific limits
	if req.Data.Platform == "twitter" {
		options["max_tokens"] = 100 // Very short for tweets
	}

	return options
}

func (a *Agent) estimateTokens(text string) int {
	// Rough estimation: 1 token ≈ 4 characters
	return len(text) / 4
}

func (a *Agent) estimateCost(tokens int, model string) float64 {
	// Rough cost estimation based on model
	costPerThousand := map[string]float64{
		"claude-3-haiku-20240307":    0.00025,
		"claude-3-5-sonnet-20241022": 0.003,
		"claude-3-opus-20240229":     0.015,
	}

	rate, ok := costPerThousand[model]
	if !ok {
		rate = 0.003 // default
	}

	return float64(tokens) / 1000.0 * rate
}

// Response methods
func (a *Agent) sendSuccessResponse(headers map[string]string, payload ResponsePayload) {
	responseBytes, _ := json.Marshal(payload)
	responseHeaders := a.createResponseHeaders(headers)

	if err := a.producer.Produce(a.ctx, ResponseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to produce response message", zap.Error(err))
		observability.SystemErrors.WithLabelValues(AgentType, "produce_response").Inc()
	} else {
		observability.KafkaMessagesProduced.WithLabelValues(ResponseTopic).Inc()
	}
}

func (a *Agent) sendErrorResponse(headers map[string]string, domainErr *errors.DomainError) {
	responseHeaders := a.createResponseHeaders(headers)
	domainErr.TraceID = headers["correlation_id"]

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   domainErr,
		"agent":   AgentType,
	}
	errorBytes, _ := json.Marshal(errorResponse)

	errorTopic := fmt.Sprintf("system.errors.%s", AgentType)
	if err := a.producer.Produce(a.ctx, errorTopic, responseHeaders,
		[]byte(headers["correlation_id"]), errorBytes); err != nil {
		a.logger.Error("Failed to produce error response", zap.Error(err))
		observability.SystemErrors.WithLabelValues(AgentType, "produce_error_response").Inc()
	} else {
		observability.KafkaMessagesProduced.WithLabelValues(errorTopic).Inc()
	}
}

func (a *Agent) createResponseHeaders(originalHeaders map[string]string) map[string]string {
	respHeaders := map[string]string{
		"correlation_id": originalHeaders["correlation_id"],
		"causation_id":   originalHeaders["request_id"],
		"request_id":     uuid.NewString(),
		"client_id":      originalHeaders["client_id"],
		"agent_type":     AgentType,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}

	// Propagate fuel budget
	if fuel, ok := originalHeaders[governance.FuelHeader]; ok {
		respHeaders[governance.FuelHeader] = fuel
	}

	return respHeaders
}

// StartHealthServer starts a simple HTTP server for health checks
func (a *Agent) StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"agent":  AgentType,
		})
	})

	go func() {
		a.logger.Info("Starting health server", zap.String("port", port))
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			a.logger.Error("Health server failed", zap.Error(err))
		}
	}()
}
