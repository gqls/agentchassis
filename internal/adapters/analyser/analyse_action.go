// FILE: internal/adapters/analyser/analyse_action.go
//
// DRAFT for the agent-chassis repo (module github.com/gqls/agentchassis).
// Does not compile in the contextkit container — built/deployed in your env.
//
// analyse action handler for the analyser adapter. Mirrors the thunder
// adapter's action-handler shape (NewXAction(deps) + Execute(ctx, req) →
// result, error).
//
// Flow:
//   1. Validate the request (owner, repo; language defaults to "go").
//   2. Fetch the repo source at ref via the read-only GitHub source fetcher.
//   3. Walk it with analysis.Analyse — the layer-1 library (Go-only today;
//      a non-Go parser slots in here behind the same action — that's the
//      polyglot drop-in the adapter exists for).
//   4. Return the structural summary + the resolved commit SHA.
//
// Stateless and idempotent: each call re-fetches and re-parses; the temp
// checkout is always cleaned up. No DB; no secret beyond the read-only token
// the fetcher holds.

package analyser

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"

	// Layer-1 walk, moved from contextkit into the chassis (see analyse.go).
	"github.com/gqls/agentchassis/internal/analysis"
)

// repoSource is the seam the handler depends on (GitHubSource implements it).
// Keeps the handler testable and lets a non-GitHub source slot in later.
type repoSource interface {
	FetchToDir(ctx context.Context, owner, repo, ref string) (dir, commitSHA string, err error)
}

// AnalyseRequest is the 'data' payload for the 'analyse' action.
type AnalyseRequest struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	Ref      string `json:"ref"`      // branch, tag, or SHA; defaults to HEAD
	Language string `json:"language"` // defaults to "go"; others not yet wired
}

// AnalyseResult is returned in the response body's data field. CommitSHA is the
// exact commit the source was read at — code_symbols.commit_sha stamps it.
type AnalyseResult struct {
	Owner     string          `json:"owner"`
	Repo      string          `json:"repo"`
	Ref       string          `json:"ref"`
	CommitSHA string          `json:"commit_sha"`
	Language  string          `json:"language"`
	Output    analysis.Output `json:"output"`
}

// AnalyseAction holds the handler's dependencies.
type AnalyseAction struct {
	source repoSource
	logger *zap.Logger
}

// NewAnalyseAction builds the handler.
func NewAnalyseAction(source repoSource, logger *zap.Logger) *AnalyseAction {
	return &AnalyseAction{source: source, logger: logger.Named("analyse")}
}

// Execute runs the fetch-and-parse flow. Returns an error for unrecoverable
// problems (bad request, unsupported language, fetch failure); the dispatcher
// maps that to an error ResponseMessage.
func (a *AnalyseAction) Execute(ctx context.Context, req AnalyseRequest) (*AnalyseResult, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, fmt.Errorf("analyse: owner and repo are required")
	}
	lang := req.Language
	if lang == "" {
		lang = "go"
	}
	if lang != "go" {
		// Polyglot is the adapter's purpose, but only Go is wired today. A JS
		// parser slots in here behind the same action without changing callers.
		return nil, fmt.Errorf("analyse: language %q not yet supported (only 'go')", lang)
	}

	dir, commitSHA, err := a.source.FetchToDir(ctx, req.Owner, req.Repo, req.Ref)
	if err != nil {
		return nil, fmt.Errorf("analyse: fetch source: %w", err)
	}
	defer os.RemoveAll(dir)

	out, err := analysis.Analyse(dir)
	if err != nil {
		return nil, fmt.Errorf("analyse: walk %s/%s@%s: %w", req.Owner, req.Repo, req.Ref, err)
	}

	a.logger.Info("analysed repo",
		zap.String("owner", req.Owner),
		zap.String("repo", req.Repo),
		zap.String("ref", req.Ref),
		zap.String("commit_sha", commitSHA),
		zap.Int("file_count", out.FileCount),
	)

	return &AnalyseResult{
		Owner:     req.Owner,
		Repo:      req.Repo,
		Ref:       req.Ref,
		CommitSHA: commitSHA,
		Language:  lang,
		Output:    out,
	}, nil
}
