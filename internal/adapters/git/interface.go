package git

import (
	"encoding/json"
)

// AdapterRequest matches the message sent by the agent
type AdapterRequest struct {
	Headers AdapterHeaders `json:"headers"`
	Body    AdapterBody    `json:"body"`
}

// AdapterHeaders matches the agent's header structure
type AdapterHeaders struct {
	// Core identifiers
	CorrelationID   string `json:"correlation_id"`
	OrchestrationID string `json:"orchestration_id"`
	RequestID       string `json:"request_id"`
	ClientID        string `json:"client_id"`

	// Parent context (critical for orchestration)
	ParentOrchestrationID string `json:"parent_orchestration_id,omitempty"`
	ParentRequestID       string `json:"parent_request_id,omitempty"`

	// Step context
	StepID   string `json:"step_id,omitempty"`
	StepName string `json:"step_name,omitempty"`

	// Response routing
	ResponsesTopic string `json:"responses_topic"`
}

// AdapterBody is the agent's body structure
type AdapterBody struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"` // The specific payload for the action
}

// GitCommitData is the expected structure of the 'data' field for a 'commit' action
type GitCommitData struct {
	RepoName      string            `json:"repo_name"`
	Domain        string            `json:"domain"`
	Files         map[string]string `json:"files"`
	CommitMessage string            `json:"commit_message"`
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
