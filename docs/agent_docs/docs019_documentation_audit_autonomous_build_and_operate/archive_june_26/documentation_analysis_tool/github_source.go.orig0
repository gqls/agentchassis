// FILE: internal/adapters/analyser/github_source.go
//
// DRAFT for the agent-chassis repo (module github.com/gqls/agentchassis).
// Does not compile in the contextkit container — built/deployed in your env.
//
// Read-only GitHub source fetcher for the analyser adapter. Fetches a repo at a
// ref as a tarball (one API call) and extracts it to a temp directory for
// parsing. Mirrors the auth/header pattern of the git adapter's GitHubClient,
// but holds a SEPARATE, READ-ONLY, repo-scoped credential — least privilege:
// the analyser only reads source, it never writes (the git adapter keeps the
// write credential for commits).
//
// Why a tarball rather than a clone or per-blob fetch: GET /repos/{owner}/
// {repo}/tarball/{ref} returns the whole tree in a single request — no git
// binary in the pod, no rate-limit storm from per-blob GETs. GitHub wraps the
// archive in a single top-level dir "{owner}-{repo}-{sha}", so the exact commit
// SHA is recovered from it, which is what code_symbols.commit_sha needs.

package analyser

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// GitHubSource fetches repository source over the GitHub API. Read-only.
type GitHubSource struct {
	token      string
	apiBase    string
	httpClient *http.Client
	log        *zap.Logger
}

// NewGitHubSource builds the fetcher. The token should be a fine-grained,
// repo-scoped, READ-ONLY (contents: read) credential — not the git adapter's
// write token.
func NewGitHubSource(token, apiBase string, log *zap.Logger) (*GitHubSource, error) {
	if token == "" {
		return nil, fmt.Errorf("analyser source: read-only GitHub token is required")
	}
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	return &GitHubSource{
		token:      token,
		apiBase:    strings.TrimRight(apiBase, "/"),
		httpClient: &http.Client{Timeout: 120 * time.Second},
		log:        log.Named("github_source"),
	}, nil
}

// FetchToDir downloads owner/repo at ref as a tarball and extracts it to a new
// temp directory. Returns the directory and the resolved commit SHA (parsed
// from the archive's top-level dir). The caller MUST os.RemoveAll(dir) when
// done. Read-only: only GETs the tarball endpoint.
func (g *GitHubSource) FetchToDir(ctx context.Context, owner, repo, ref string) (dir, commitSHA string, err error) {
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner and repo are required")
	}
	if ref == "" {
		ref = "HEAD"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", g.apiBase, owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("build tarball request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch tarball: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", "", fmt.Errorf("github tarball %s/%s@%s returned %d: %s",
			owner, repo, ref, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	tmp, err := os.MkdirTemp("", "analyser-src-*")
	if err != nil {
		return "", "", fmt.Errorf("create temp dir: %w", err)
	}

	commitSHA, err = extractTarGz(resp.Body, tmp)
	if err != nil {
		os.RemoveAll(tmp)
		return "", "", fmt.Errorf("extract tarball: %w", err)
	}

	g.log.Info("fetched repo source",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.String("ref", ref),
		zap.String("commit_sha", commitSHA),
		zap.String("dir", tmp),
	)
	return tmp, commitSHA, nil
}

// extractTarGz unpacks a gzipped tar stream into destDir, stripping the single
// top-level directory GitHub wraps the archive in ("{owner}-{repo}-{sha}/").
// Returns the SHA parsed from that top-level dir name. Guards against path
// traversal (entries resolving outside destDir are rejected) and caps per-file
// size defensively.
func extractTarGz(r io.Reader, destDir string) (commitSHA string, err error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	cleanDest := filepath.Clean(destDir)
	topPrefix := ""
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar next: %w", err)
		}

		// First entry establishes the top-level dir ("{owner}-{repo}-{sha}/").
		name := hdr.Name
		if topPrefix == "" {
			if i := strings.IndexByte(name, '/'); i >= 0 {
				topPrefix = name[:i+1]
				top := strings.TrimSuffix(topPrefix, "/")
				if dash := strings.LastIndexByte(top, '-'); dash >= 0 {
					commitSHA = top[dash+1:]
				}
			}
		}
		rel := strings.TrimPrefix(name, topPrefix)
		if rel == "" {
			continue
		}

		target := filepath.Join(cleanDest, rel)
		// Path-traversal guard.
		if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			return "", fmt.Errorf("tar entry escapes destination: %s", name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, io.LimitReader(tr, 64<<20)); err != nil {
				f.Close()
				return "", err
			}
			f.Close()
		default:
			// Skip symlinks and other entry types — not needed for Go source,
			// and safer (a symlink could point outside destDir).
		}
	}
	return commitSHA, nil
}
