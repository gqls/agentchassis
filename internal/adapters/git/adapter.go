package git

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// ---
// 1. Configuration
// ---

type Config struct {
	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroup    string
	GitHubToken   string
	GitHubOrg     string // Optional: to create repos in an organization
	GitHubAPIBase string
}

func LoadConfig() Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal("KAFKA_BROKERS env var not set")
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "system.adapter.git.requests"
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN env var not set")
	}

	return Config{
		KafkaBrokers:  strings.Split(brokers, ","),
		KafkaTopic:    topic,
		KafkaGroup:    "git-adapter-group",
		GitHubToken:   token,
		GitHubOrg:     os.Getenv("GITHUB_ORG"), // e.g., "my-company"
		GitHubAPIBase: "https://api.github.com",
	}
}

// ---
// 2. Main Service
// ---

func main() {
	cfg := LoadConfig()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting git-adapter-service",
		zap.Strings("brokers", cfg.KafkaBrokers),
		zap.String("topic", cfg.KafkaTopic),
	)

	// Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0 // Or your Kafka version
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	// HTTP client for GitHub API
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Producer for sending responses
	producer, err := sarama.NewSyncProducer(cfg.KafkaBrokers, saramaConfig)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	defer producer.Close()

	// Create the handler
	handler := &ConsumerGroupHandler{
		cfg:        cfg,
		logger:     logger,
		httpClient: httpClient,
		producer:   producer,
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroup, saramaConfig)
	if err != nil {
		logger.Fatal("Failed to create consumer group", zap.Error(err))
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{cfg.KafkaTopic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				logger.Error("Error from consumer", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Wait for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	case <-sigterm:
		logger.Info("Interrupt signal received, shutting down")
	}

	cancel()
	wg.Wait()
	if err := client.Close(); err != nil {
		logger.Error("Failed to close consumer client", zap.Error(err))
	}
}

// ---
// 3. Kafka Consumer
// ---

// ConsumerGroupHandler implements the sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	cfg        Config
	logger     *zap.Logger
	httpClient *http.Client
	producer   sarama.SyncProducer
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// This is the core message processing loop
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.logger.Info("Received message",
			zap.String("topic", message.Topic),
			zap.Int64("offset", message.Offset),
		)

		var req AdapterRequest
		if err := json.Unmarshal(message.Value, &req); err != nil {
			h.logger.Error("Failed to unmarshal request", zap.Error(err), zap.ByteString("value", message.Value))
			session.MarkMessage(message, "") // Mark as processed
			continue
		}

		// Get response topic from headers
		responsesTopic := req.Headers.ResponsesTopic
		if responsesTopic == "" {
			h.logger.Error("No responses_topic in headers, discarding message", zap.String("request_id", req.Headers.RequestID))
			session.MarkMessage(message, "")
			continue
		}

		// Handle the request and get a response payload
		responsePayload := h.handleRequest(req)

		// Send the response
		if err := h.sendResponse(responsesTopic, req.Headers, responsePayload); err != nil {
			h.logger.Error("Failed to send response", zap.Error(err))
		}

		// Mark message as processed
		session.MarkMessage(message, "")
	}
	return nil
}

// handleRequest routes the message to the correct git function
func (h *ConsumerGroupHandler) handleRequest(req AdapterRequest) interface{} {
	switch req.Body.Action {
	case "commit":
		repoURL, err := h.commitToRepo(req.Body.Data)
		if err != nil {
			h.logger.Error("Failed to process git commit", zap.Error(err), zap.String("request_id", req.Headers.RequestID))
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		}
		return map[string]interface{}{
			"success":  true,
			"repo_url": repoURL,
		}
	default:
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("unknown action: %s", req.Body.Action),
		}
	}
}

// sendResponse sends a JSON payload back to the specified topic
func (h *ConsumerGroupHandler) sendResponse(topic string, headers AdapterHeaders, payload interface{}) error {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal response payload: %w", err)
	}

	// Create a minimal response envelope
	responseMsg := map[string]interface{}{
		"headers": map[string]string{
			"correlation_id":   headers.CorrelationID,
			"orchestration_id": headers.OrchestrationID,
			"request_id":       headers.RequestID,
			"message_type":     "response",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		},
		"body": json.RawMessage(responseBody), // Embed the payload as raw JSON
	}

	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal full response: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(headers.RequestID), // Use RequestID as key
		Value: sarama.ByteString(responseBytes),
	}

	_, _, err = h.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send response message: %w", err)
	}

	h.logger.Info("Sent response", zap.String("topic", topic), zap.String("request_id", headers.RequestID))
	return nil
}

// ---
// 4. GitHub API Logic
// ---

// This function is the core of your "commit" action.
// It's complex because the GitHub API requires multiple steps to commit files.
func (h *ConsumerGroupHandler) commitToRepo(data json.RawMessage) (string, error) {
	// 1. Parse the 'data' payload from the body
	var commitData struct {
		RepoName      string            `json:"repo_name"`
		Files         map[string]string `json:"files"`
		CommitMessage string            `json:"commit_message"`
	}
	if err := json.Unmarshal(data, &commitData); err != nil {
		return "", fmt.Errorf("failed to parse commit data: %w", err)
	}

	if commitData.RepoName == "" || len(commitData.Files) == 0 {
		return "", fmt.Errorf("repo_name and files are required")
	}

	// 2. Create or Get the Repo
	repo, err := h.createOrGetRepo(commitData.RepoName)
	if err != nil {
		return "", fmt.Errorf("failed to create/get repo: %w", err)
	}

	// 3. Get the latest commit SHA from the default branch
	latestSHA, err := h.getLatestCommitSHA(repo.Owner.Login, repo.Name, repo.DefaultBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit: %w", err)
	}

	// 4. Create a "Blob" for each file
	var treeEntries []TreeEntry
	for path, content := range commitData.Files {
		blobSHA, err := h.createBlob(repo.Owner.Login, repo.Name, content)
		if err != nil {
			return "", fmt.Errorf("failed to create blob for %s: %w", path, err)
		}
		treeEntries = append(treeEntries, TreeEntry{
			Path: path,
			Mode: "100644", // file
			Type: "blob",
			SHA:  blobSHA,
		})
	}

	// 5. Create a new "Tree" from the blobs
	newTreeSHA, err := h.createTree(repo.Owner.Login, repo.Name, latestSHA, treeEntries)
	if err != nil {
		return "", fmt.Errorf("failed to create tree: %w", err)
	}

	// 6. Create the new "Commit"
	newCommitSHA, err := h.createCommit(repo.Owner.Login, repo.Name, commitData.CommitMessage, newTreeSHA, latestSHA)
	if err != nil {
		return "", fmt.Errorf("failed to create commit: %w", err)
	}

	// 7. Update the "Ref" (e.g., move 'main' branch to point to the new commit)
	if err := h.updateRef(repo.Owner.Login, repo.Name, repo.DefaultBranch, newCommitSHA); err != nil {
		return "", fmt.Errorf("failed to update ref: %w", err)
	}

	return repo.HTMLURL, nil
}

// --- GitHub API Helper Functions ---
// (These would be in a separate file in a larger service)

func (h *ConsumerGroupHandler) createOrGetRepo(repoName string) (*GitHubRepo, error) {
	// First, try to get the repo
	url := fmt.Sprintf("%s/repos/%s/%s", h.GitHubAPIBase, h.getRepoOwner(), repoName)
	req, _ := http.NewRequest("GET", url, nil)

	repo := &GitHubRepo{}
	if err := h.sendGitHubRequest(req, &repo); err == nil {
		// Got it successfully
		h.logger.Info("Found existing repo", zap.String("repo", repoName))
		return repo, nil
	}

	// Not found, so let's create it
	h.logger.Info("Repo not found, creating...", zap.String("repo", repoName))

	createURL := h.GitHubAPIBase + "/user/repos"
	if h.cfg.GitHubOrg != "" {
		createURL = fmt.Sprintf("%s/orgs/%s/repos", h.GitHubAPIBase, h.cfg.GitHubOrg)
	}

	body := map[string]interface{}{
		"name":      repoName,
		"private":   false, // Or true, as you wish
		"auto_init": true,  // Create with a README
	}
	jsonBody, _ := json.Marshal(body)
	req, _ = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonBody))

	if err := h.sendGitHubRequest(req, &repo); err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}
	return repo, nil
}

func (h *ConsumerGroupHandler) getLatestCommitSHA(owner, repo, branch string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/ref/heads/%s", h.GitHubAPIBase, owner, repo, branch)
	req, _ := http.NewRequest("GET", url, nil)

	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}

	if err := h.sendGitHubRequest(req, &ref); err != nil {
		return "", err
	}
	return ref.Object.SHA, nil
}

func (h *ConsumerGroupHandler) createBlob(owner, repo, content string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/blobs", h.GitHubAPIBase, owner, repo)
	body := map[string]string{
		"content":  content,
		"encoding": "utf-8",
	}
	// Note: GitHub also supports base64 encoding if you are committing binary files
	// body["encoding"] = "base64"
	// body["content"] = base64.StdEncoding.EncodeToString([]byte(content))

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

	var blob struct {
		SHA string `json:"sha"`
	}
	if err := h.sendGitHubRequest(req, &blob); err != nil {
		return "", err
	}
	return blob.SHA, nil
}

func (h *ConsumerGroupHandler) createTree(owner, repo, baseTreeSHA string, entries []TreeEntry) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/trees", h.GitHubAPIBase, owner, repo)
	body := map[string]interface{}{
		"base_tree": baseTreeSHA,
		"tree":      entries,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

	var tree struct {
		SHA string `json:"sha"`
	}
	if err := h.sendGitHubRequest(req, &tree); err != nil {
		return "", err
	}
	return tree.SHA, nil
}

func (h *ConsumerGroupHandler) createCommit(owner, repo, message, treeSHA, parentSHA string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/commits", h.GitHubAPIBase, owner, repo)
	body := map[string]interface{}{
		"message": message,
		"tree":    treeSHA,
		"parents": []string{parentSHA},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := h.sendGitHubRequest(req, &commit); err != nil {
		return "", err
	}
	return commit.SHA, nil
}

func (h *ConsumerGroupHandler) updateRef(owner, repo, branch, commitSHA string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", h.GitHubAPIBase, owner, repo, branch)
	body := map[string]string{
		"sha":   commitSHA,
		"force": false, // Set to true if you want to force-push
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))

	return h.sendGitHubRequest(req, nil)
}

func (h *ConsumerGroupHandler) sendGitHubRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Authorization", "Bearer "+h.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// ... (add error body parsing here for better debugging)
		return fmt.Errorf("github API request failed with status: %s", resp.Status)
	}

	if v == nil {
		return nil // We don't need to decode the response body
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (h *ConsumerGroupHandler) getRepoOwner() string {
	if h.cfg.GitHubOrg != "" {
		return h.cfg.GitHubOrg
	}
	// If no org is specified, it will create a repo for the authenticated user.
	// We need the user's login. This is a simple way, but a better way
	// is to call the /user endpoint once on startup.
	url := fmt.Sprintf("%s/user", h.GitHubAPIBase)
	req, _ := http.NewRequest("GET", url, nil)
	var user struct {
		Login string `json:"login"`
	}
	if err := h.sendGitHubRequest(req, &user); err == nil {
		return user.Login
	}
	return "" // Fallback, but should be handled better
}

// ---
// 5. Shared Structs (for readability)
// ---

// AdapterRequest matches the message sent by the agent
type AdapterRequest struct {
	Headers AdapterHeaders `json:"headers"`
	Body    AdapterBody    `json:"body"`
}

type AdapterHeaders struct {
	CorrelationID   string `json:"correlation_id"`
	OrchestrationID string `json:"orchestration_id"`
	RequestID       string `json:"request_id"`
	ResponsesTopic  string `json:"responses_topic"`
}

type AdapterBody struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"` // The specific payload for the action
}

// GitHubRepo is a partial struct for GitHub's API response
type GitHubRepo struct {
	Name          string `json:"name"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// TreeEntry is for building the Git tree
type TreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
}
