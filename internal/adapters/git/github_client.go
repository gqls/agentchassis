package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"strings"
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

// CommitToRepo is the main function that orchestrates the commit.
//
// It carries WRITES (data.Files) and REMOVALS (data.Deletions) through the same
// tree/commit/ref machinery, because in the Git Data API a removal simply IS a
// tree entry with a null sha. Doing it here rather than in a bespoke Contents-API
// call is what gives a retraction the ref-race retry below, the domain-prefixing
// convention, and atomicity with any writes in the same request — see the type
// comment on GitCommitData.Deletions.
func (c *GitHubClient) CommitToRepo(ctx context.Context, data GitCommitData) (string, error) {
	if data.RepoName == "" || (len(data.Files) == 0 && len(data.Deletions) == 0) {
		return "", fmt.Errorf("repo_name and at least one of files/deletions are required")
	}

	// Deletion paths are validated BEFORE anything is prefixed, and writes are
	// not. That asymmetry is deliberate: a malformed write path creates a junk
	// file, which is visible and recoverable by writing again; a malformed
	// DELETE path destroys someone else's artefact, and `../other-site.com/
	// index.html` under a domain prefix is a real traversal out of the caller's
	// own directory. The primitive refuses rather than trusting its callers,
	// because the chassis-side caller is not the only possible one.
	for _, p := range data.Deletions {
		if err := validateDeletionPath(p); err != nil {
			return "", err
		}
	}

	// Files go to {domain}/{filename} — a SITE-CONTENT convention. A branch-
	// targeted commit with no domain (the fix-implementer writing platform
	// code) uses paths repo-relative, unprefixed.
	if data.Domain != "" {
		prefixedFiles := make(map[string]interface{})
		for path, content := range data.Files {
			prefixedPath := data.Domain + "/" + path
			prefixedFiles[prefixedPath] = content
		}
		data.Files = prefixedFiles

		// Deletions take the IDENTICAL prefix from the IDENTICAL field. One
		// implementation, so a retraction cannot address a path outside the
		// domain a publish of the same request would have written to.
		prefixedDeletions := make([]string, 0, len(data.Deletions))
		for _, path := range data.Deletions {
			prefixedDeletions = append(prefixedDeletions, data.Domain+"/"+path)
		}
		data.Deletions = prefixedDeletions
	}

	c.log.Info("Committing to repo",
		zap.String("repo_name", data.RepoName),
		zap.String("branch", data.Branch),
		zap.Any("DEBUGaa: data", data),
	)

	// 1. Create or Get the Repo
	repo, err := c.createOrGetRepo(ctx, data.RepoName)
	if err != nil {
		return "", fmt.Errorf("failed to create/get repo: %w", err)
	}

	// Target branch: explicit request or the repo default.
	branch := data.Branch
	if branch == "" {
		branch = repo.DefaultBranch
	}

	// 3. Create a "Blob" for each file — ONCE, before the retry loop below.
	// Blobs are content-addressed: their SHAs stay valid whatever the branch
	// head moves to, so only the tree/commit/ref steps need re-basing when a
	// concurrent committer wins the ref race.
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
		sha := blobSHA
		treeEntries = append(treeEntries, TreeEntry{
			Path: path,
			Mode: "100644", // file
			Type: "blob",
			SHA:  &sha,
		})
	}

	// 2/4/5/6. Read the branch head, build tree+commit on it, move the ref —
	// retried as one unit on a non-fast-forward (bugs_open/120, owner ruling
	// 2026-07-28: same-repo deploys serialise). GitHub's 422 on updateRef IS
	// the serialisation point: every site shares one repo and one branch, and
	// with more than one adapter replica (or the chassis worker pool driving
	// concurrent deploys) two commits race read-head→update-ref; the loser
	// used to surface the 422 and fail its orchestration outright. Now the
	// loser re-reads the winner's head and re-bases — commits queue one
	// behind another at the API's own consistency check, the same optimistic
	// CAS-and-retry idiom as UpdateStateWithVersion. Distinct pages are
	// distinct files, and a same-file race is last-writer-wins, which is
	// already re-render semantics. Also covers a HUMAN push landing between
	// read and update — the collision class bugs_open/120 documents.
	const maxRefRaceRetries = 4
	for attempt := 1; ; attempt++ {
		// Latest commit SHA from the target branch; a just-created empty repo
		// has no head yet, so fall back to the base tree.
		latestSHA, err := c.getLatestCommitSHA(ctx, repo.Owner.Login, repo.Name, branch)
		if err != nil {
			latestSHA, err = c.getBaseTreeSHA(ctx, repo.Owner.Login, repo.Name, branch)
			if err != nil {
				return "", fmt.Errorf("failed to get latest commit/base tree for branch %q: %w", branch, err)
			}
		}

		// Removals are resolved INSIDE the loop, unlike blobs. A blob's sha is
		// content-addressed and stays valid whatever the head moves to; a
		// removal is a statement ABOUT the head, so a re-base must re-ask. If
		// the winner of a ref race removed the same path, this attempt now sees
		// it absent and drops it rather than asking GitHub to delete something
		// the new base tree no longer holds.
		//
		// THE FILTER IS REQUIRED, NOT DEFENSIVE, and this was probed rather
		// than assumed (gqls/sites, 2026-08-02, POST /git/trees against the
		// live master tree — a tree object is created unreferenced, so the
		// probe moves no ref and fires no workflow):
		//
		//	null sha, path PRESENT  -> 201, new tree sha
		//	null sha, path ABSENT   -> 422 GitRPC::BadObjectState
		//
		// So without this, re-running a repair FAILS. Being absent is the state
		// the caller asked for, so it is reported as a success.
		//
		// Existence is checked PER PATH rather than from one recursive tree
		// listing. `GET /git/trees/{sha}?recursive=1` carries a `truncated`
		// flag, and a truncated listing would report present files as ABSENT —
		// a real retraction would then skip silently and report success.
		// MEASURED 2026-08-02: the sites repo is 1,847 entries and
		// `truncated: false`, i.e. nowhere near the limit today — so this is
		// headroom, not a live hazard, and it is not the main reason. The main
		// reason is that a per-path 404 is also what makes the present/absent
		// split reportable at all.
		entries := treeEntries
		var absent []string
		if len(data.Deletions) > 0 {
			present, missing, err := c.partitionExistingPaths(ctx, repo.Owner.Login, repo.Name, branch, data.Deletions)
			if err != nil {
				return "", fmt.Errorf("failed to resolve deletion paths on branch %q: %w", branch, err)
			}
			absent = missing
			// Copy: treeEntries is reused by the next attempt and must not
			// accumulate one attempt's removals.
			entries = make([]TreeEntry, 0, len(treeEntries)+len(present))
			entries = append(entries, treeEntries...)
			for _, p := range present {
				entries = append(entries, TreeEntry{
					Path: p,
					Mode: "100644",
					Type: "blob",
					SHA:  nil, // null => remove this path from the tree
				})
			}
		}

		// Nothing to write and every requested removal already gone. Return
		// success WITHOUT committing: an empty commit would still push, fire
		// the deploy workflow and purge a Cloudflare zone for no change at all.
		if len(entries) == 0 {
			c.log.Info("CommitToRepo: no-op — every requested deletion is already absent",
				zap.String("repo", repo.Name),
				zap.String("branch", branch),
				zap.Strings("absent", absent))
			return repo.HTMLURL, nil
		}

		newTreeSHA, err := c.createTree(ctx, repo.Owner.Login, repo.Name, latestSHA, entries)
		if err != nil {
			return "", fmt.Errorf("failed to create tree: %w", err)
		}

		newCommitSHA, err := c.createCommit(ctx, repo.Owner.Login, repo.Name, data.CommitMessage, newTreeSHA, latestSHA)
		if err != nil {
			return "", fmt.Errorf("failed to create commit: %w", err)
		}

		err = c.updateRef(ctx, repo.Owner.Login, repo.Name, branch, newCommitSHA)
		if err == nil {
			return repo.HTMLURL, nil
		}
		if !isNonFastForward(err) || attempt >= maxRefRaceRetries {
			return "", fmt.Errorf("failed to update ref for branch %q: %w", branch, err)
		}

		c.log.Warn("REF_RACE_RETRY: non-fast-forward — a concurrent commit won; re-basing on the new head",
			zap.String("repo", repo.Name),
			zap.String("branch", branch),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", maxRefRaceRetries))
		// Brief, growing pause so the winner's ref settles before re-reading.
		time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
	}
}

// isNonFastForward recognises GitHub's refusal to move a ref backwards or
// sideways — the 422 whose body says "Update is not a fast forward". It is
// the ONLY updateRef failure that re-basing can cure; everything else
// (auth, rate limit, 5xx) must keep failing loudly.
func isNonFastForward(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "fast forward")
}

// CreateBranch creates a new branch from FromBranch's head (default: the
// repo's default branch). F1.1b(c): the fix-implementer's first git step.
// Idempotent on re-runs: if the branch already exists, its current head is
// returned with created=false rather than an error — a re-fired implementer
// run must not die on its own leftovers.
func (c *GitHubClient) CreateBranch(ctx context.Context, data GitCreateBranchData) (sha string, created bool, err error) {
	if data.RepoName == "" || data.Branch == "" {
		return "", false, fmt.Errorf("repo_name and branch are required")
	}

	// getRepo, NOT createOrGetRepo: a typo'd repo name must fail loudly, not
	// auto-create an empty repo (that behaviour belongs to the site-content flow).
	repo, err := c.getRepo(ctx, data.RepoName)
	if err != nil {
		return "", false, fmt.Errorf("failed to get repo: %w", err)
	}

	from := data.FromBranch
	if from == "" {
		from = repo.DefaultBranch
	}
	baseSHA, err := c.getLatestCommitSHA(ctx, repo.Owner.Login, repo.Name, from)
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve head of %q: %w", from, err)
	}

	url := fmt.Sprintf("%s/repos/%s/%s/git/refs", c.apiBase, repo.Owner.Login, repo.Name)
	body := map[string]string{
		"ref": "refs/heads/" + data.Branch,
		"sha": baseSHA,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))

	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := c.sendGitHubRequest(req, &ref); err != nil {
		// 422 "Reference already exists" → return the existing head.
		if strings.Contains(err.Error(), "already exists") {
			existing, gerr := c.getLatestCommitSHA(ctx, repo.Owner.Login, repo.Name, data.Branch)
			if gerr != nil {
				return "", false, fmt.Errorf("branch exists but head unreadable: %w", gerr)
			}
			c.log.Info("Branch already exists, returning existing head",
				zap.String("branch", data.Branch), zap.String("sha", existing))
			return existing, false, nil
		}
		return "", false, fmt.Errorf("failed to create branch %q: %w", data.Branch, err)
	}

	c.log.Info("Branch created",
		zap.String("branch", data.Branch),
		zap.String("from", from),
		zap.String("sha", ref.Object.SHA))
	return ref.Object.SHA, true, nil
}

// CreatePullRequest opens a PR from Head into Base (default: the repo's
// default branch). F1.1b(c): the fix loop's HUMAN TERMINAL — the platform
// creates PRs and never merges them.
func (c *GitHubClient) CreatePullRequest(ctx context.Context, data GitCreatePRData) (htmlURL string, number int, err error) {
	if data.RepoName == "" || data.Title == "" || data.Head == "" {
		return "", 0, fmt.Errorf("repo_name, title and head are required")
	}

	repo, err := c.getRepo(ctx, data.RepoName)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get repo: %w", err)
	}

	base := data.Base
	if base == "" {
		base = repo.DefaultBranch
	}

	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.apiBase, repo.Owner.Login, repo.Name)
	body := map[string]interface{}{
		"title": data.Title,
		"body":  data.Body,
		"head":  data.Head,
		"base":  base,
		"draft": data.Draft,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))

	var pr struct {
		HTMLURL string `json:"html_url"`
		Number  int    `json:"number"`
	}
	if err := c.sendGitHubRequest(req, &pr); err != nil {
		return "", 0, fmt.Errorf("failed to create pull request %q -> %q: %w", data.Head, base, err)
	}

	c.log.Info("Pull request created",
		zap.String("url", pr.HTMLURL),
		zap.Int("number", pr.Number),
		zap.String("head", data.Head),
		zap.String("base", base))
	return pr.HTMLURL, pr.Number, nil
}

// getRepo fetches a repo WITHOUT creating it when absent — branch/PR
// operations must fail loudly on a bad repo name.
func (c *GitHubClient) getRepo(ctx context.Context, repoName string) (*GitHubRepo, error) {
	owner := c.getRepoOwner()
	url := fmt.Sprintf("%s/repos/%s/%s", c.apiBase, owner, repoName)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	repo := &GitHubRepo{}
	if err := c.sendGitHubRequest(req, repo); err != nil {
		return nil, fmt.Errorf("repo %s/%s not accessible: %w", owner, repoName, err)
	}
	return repo, nil
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

// validateDeletionPath refuses anything that is not a plain repo-relative file
// path. It rejects rather than sanitises, for the reason
// datahelpers.PageFilePathFromURL gives about the publish side: silently
// rewriting a path that was meant to address one thing so that it addresses
// another is worse than declining it, and a deletion is where that costs most.
func validateDeletionPath(p string) error {
	if strings.TrimSpace(p) != p || p == "" {
		return fmt.Errorf("deletion path %q is empty or padded", p)
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("deletion path %q must be repo-relative, not absolute", p)
	}
	if strings.Contains(p, `\`) {
		return fmt.Errorf("deletion path %q contains a backslash; GitHub tree paths take it literally", p)
	}
	// path.Clean collapses "..", "//" and "./". If cleaning would CHANGE the
	// path, the path was not what it appeared to be — including the traversal
	// case, which is the one that matters.
	if cleaned := path.Clean(p); cleaned != p {
		return fmt.Errorf("deletion path %q is not canonical (would clean to %q)", p, cleaned)
	}
	if p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return fmt.Errorf("deletion path %q escapes the repository root", p)
	}
	return nil
}

// partitionExistingPaths splits paths into those the branch currently holds and
// those it does not, preserving the caller's order in both. It is the guard that
// makes a retraction idempotent — GitHub answers a null-sha entry for a path its
// base tree does not hold with 422 GitRPC::BadObjectState (probed on the live
// repo; see the call site), so without this a second run of a repair fails.
//
// A non-404 failure is returned as an error rather than being treated as
// "absent": guessing absent on a 500 or a rate limit would silently skip a
// removal the caller asked for and report success, which is the failure mode
// this whole bug is about.
func (c *GitHubClient) partitionExistingPaths(ctx context.Context, owner, repo, branch string, paths []string) (present, absent []string, err error) {
	for _, p := range paths {
		exists, err := c.pathExists(ctx, owner, repo, branch, p)
		if err != nil {
			return nil, nil, fmt.Errorf("existence check for %q: %w", p, err)
		}
		if exists {
			present = append(present, p)
		} else {
			absent = append(absent, p)
		}
	}
	return present, absent, nil
}

// pathExists reports whether ref holds path. 404 is the only status that means
// "no"; every other non-2xx is an error.
func (c *GitHubClient) pathExists(ctx context.Context, owner, repo, ref, path string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		c.apiBase, owner, repo, pathEscapeSegments(path), neturl.QueryEscape(ref))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return false, nil
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return true, nil
	default:
		return false, fmt.Errorf("github contents API returned %s for %q", resp.Status, path)
	}
}

// pathEscapeSegments escapes each path segment but keeps the "/" separators —
// neturl.PathEscape would turn them into %2F and address a single file whose
// name contains slashes, which is not a thing.
func pathEscapeSegments(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = neturl.PathEscape(s)
	}
	return strings.Join(segs, "/")
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
