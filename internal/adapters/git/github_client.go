package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// GitHubClient handles all interactions with the GitHub API
type GitHubClient struct {
	httpClient *http.Client
	token      string
	org        string
	apiBase    string
	log        *zap.Logger
	userLogin  string
}

// NewGitHubClient creates and initializes a GitHub client
func NewGitHubClient(token, org, apiBase string, log *zap.Logger) (*GitHubClient, error) {
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}

	c := &GitHubClient{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		token:      token,
		org:        org,
		apiBase:    apiBase,
		log:        log,
	}

	// Get the authenticated user's login, so we know who the owner is
	if org == "" {
		login, err := c.getAuthenticatedUserLogin(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get authenticated user: %w", err)
		}
		c.userLogin = login
		log.Info("Authenticated as user", zap.String("login", login))
	} else {
		log.Info("Operating in organization mode", zap.String("org", org))
	}

	return c, nil
}

// CommitToRepo is the main function that orchestrates the commit
func (c *GitHubClient) CommitToRepo(ctx context.Context, data GitCommitData) (string, error) {
	if data.RepoName == "" || len(data.Files) == 0 {
		return "", fmt.Errorf("repo_name and files are required")
	}

	// Files now go to {domain}/{filename}
	prefixedFiles := make(map[string]interface{})
	for path, content := range data.Files {
		prefixedPath := data.Domain + "/" + path
		prefixedFiles[prefixedPath] = content
	}
	data.Files = prefixedFiles

	c.log.Info("Committing to repo",
		zap.String("repo_name", data.RepoName),
		zap.Any("DEBUGaa: data", data),
	)

	// 1. Create or Get the Repo
	repo, err := c.createOrGetRepo(ctx, data.RepoName)
	if err != nil {
		return "", fmt.Errorf("failed to create/get repo: %w", err)
	}

	// 2. Get the latest commit SHA from the default branch
	latestSHA, err := c.getLatestCommitSHA(ctx, repo.Owner.Login, repo.Name, repo.DefaultBranch)
	if err != nil {
		// This might fail if the repo was *just* created and is empty
		// Let's try to get the base tree SHA
		latestSHA, err = c.getBaseTreeSHA(ctx, repo.Owner.Login, repo.Name, repo.DefaultBranch)
		if err != nil {
			return "", fmt.Errorf("failed to get latest commit/base tree: %w", err)
		}
	}

	// 3. Create a "Blob" for each file
	var treeEntries []TreeEntry
	for path, fileData := range data.Files {
		var content, encoding string

		switch v := fileData.(type) {
		case string:
			// Legacy format: just a content string
			content = v
			encoding = "utf-8"
		case map[string]interface{}:
			// New format: {content, encoding}
			if c, ok := v["content"].(string); ok {
				content = c
			}
			if e, ok := v["encoding"].(string); ok {
				encoding = e
			}
			if encoding == "" {
				encoding = "utf-8"
			}
		default:
			return "", fmt.Errorf("invalid file data type for %s: %T", path, fileData)
		}

		blobSHA, err := c.createBlob(ctx, repo.Owner.Login, repo.Name, content, encoding)
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

	// 4. Create a new "Tree" from the blobs
	newTreeSHA, err := c.createTree(ctx, repo.Owner.Login, repo.Name, latestSHA, treeEntries)
	if err != nil {
		return "", fmt.Errorf("failed to create tree: %w", err)
	}

	// 5. Create the new "Commit"
	newCommitSHA, err := c.createCommit(ctx, repo.Owner.Login, repo.Name, data.CommitMessage, newTreeSHA, latestSHA)
	if err != nil {
		return "", fmt.Errorf("failed to create commit: %w", err)
	}

	// 6. Update the "Ref" (e.g., move 'main' branch to point to the new commit)
	if err := c.updateRef(ctx, repo.Owner.Login, repo.Name, repo.DefaultBranch, newCommitSHA); err != nil {
		return "", fmt.Errorf("failed to update ref: %w", err)
	}

	return repo.HTMLURL, nil
}

// --- GitHub API Helper Functions ---

func (c *GitHubClient) createOrGetRepo(ctx context.Context, repoName string) (*GitHubRepo, error) {
	owner := c.getRepoOwner()
	url := fmt.Sprintf("%s/repos/%s/%s", c.apiBase, owner, repoName)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	c.log.Info("In createOrGetRepo github_client.go GitHub repo",
		zap.String("url", url),
		zap.String("owner", owner),
	)

	repo := &GitHubRepo{}
	if err := c.sendGitHubRequest(req, &repo); err == nil {
		c.log.Info("Found existing repo", zap.String("repo", repoName))
		return repo, nil
	}

	c.log.Info("Repo not found, creating...", zap.String("repo", repoName))

	createURL := c.apiBase + "/user/repos"
	if c.org != "" {
		createURL = fmt.Sprintf("%s/orgs/%s/repos", c.apiBase, c.org)
	}

	body := map[string]interface{}{
		"name":      repoName,
		"private":   false,
		"auto_init": true, // Create with a README
	}
	jsonBody, _ := json.Marshal(body)
	req, _ = http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewBuffer(jsonBody))

	if err := c.sendGitHubRequest(req, &repo); err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}
	// GitHub API repo creation can be slow, add a small delay
	time.Sleep(1 * time.Second)
	return repo, nil
}

func (c *GitHubClient) getLatestCommitSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/ref/heads/%s", c.apiBase, owner, repo, branch)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}

	if err := c.sendGitHubRequest(req, &ref); err != nil {
		return "", err
	}
	return ref.Object.SHA, nil
}

func (c *GitHubClient) getBaseTreeSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	// Fallback for newly created, empty repos
	url := fmt.Sprintf("%s/repos/%s/%s/branches/%s", c.apiBase, owner, repo, branch)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	var branchInfo struct {
		Commit struct {
			SHA    string `json:"sha"`
			Commit struct {
				Tree struct {
					SHA string `json:"sha"`
				} `json:"tree"`
			} `json:"commit"`
		} `json:"commit"`
	}

	if err := c.sendGitHubRequest(req, &branchInfo); err != nil {
		return "", fmt.Errorf("failed to get branch info for base tree: %w", err)
	}
	return branchInfo.Commit.Commit.Tree.SHA, nil
}

func (c *GitHubClient) createBlob(ctx context.Context, owner, repo, content, encoding string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/blobs", c.apiBase, owner, repo)

	// Default to utf-8 if not specified
	if encoding == "" {
		encoding = "utf-8"
	}

	body := map[string]string{
		"content":  content,
		"encoding": encoding,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))

	var blob struct {
		SHA string `json:"sha"`
	}
	if err := c.sendGitHubRequest(req, &blob); err != nil {
		return "", err
	}
	return blob.SHA, nil
}

func (c *GitHubClient) createTree(ctx context.Context, owner, repo, baseTreeSHA string, entries []TreeEntry) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/trees", c.apiBase, owner, repo)
	body := map[string]interface{}{
		"base_tree": baseTreeSHA,
		"tree":      entries,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))

	var tree struct {
		SHA string `json:"sha"`
	}
	if err := c.sendGitHubRequest(req, &tree); err != nil {
		return "", err
	}
	return tree.SHA, nil
}

func (c *GitHubClient) createCommit(ctx context.Context, owner, repo, message, treeSHA, parentSHA string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/commits", c.apiBase, owner, repo)
	body := map[string]interface{}{
		"message": message,
		"tree":    treeSHA,
		"parents": []string{parentSHA},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := c.sendGitHubRequest(req, &commit); err != nil {
		return "", err
	}
	return commit.SHA, nil
}

func (c *GitHubClient) updateRef(ctx context.Context, owner, repo, branch, commitSHA string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", c.apiBase, owner, repo, branch)
	body := map[string]interface{}{
		"sha":   commitSHA,
		"force": false, // Boolean, not string
	}

	c.log.Info("in updateRef ",
		zap.String("branch", branch),
		zap.String("commitSHA", commitSHA),
	)

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonBody))

	return c.sendGitHubRequest(req, nil)
}

func (c *GitHubClient) getAuthenticatedUserLogin(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/user", c.apiBase)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	var user struct {
		Login string `json:"login"`
	}
	if err := c.sendGitHubRequest(req, &user); err != nil {
		return "", err
	}
	return user.Login, nil
}

func (c *GitHubClient) sendGitHubRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read error body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.log.Error("in sendGithubRequest updateRef failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(bodyBytes)),
		)
		return fmt.Errorf("github API request failed with status: %s - %s", resp.Status, string(bodyBytes))
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *GitHubClient) getRepoOwner() string {
	if c.org != "" {
		return c.org
	}
	return c.userLogin
}
