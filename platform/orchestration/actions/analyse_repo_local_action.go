// FILE: platform/orchestration/actions/analyse_repo_local_action.go
//
// DRAFT for the agent-chassis repo (module github.com/gqls/agentchassis).
// Does NOT compile in the contextkit container — built/deployed in your env.
//
// analyse_repo_local REPLACES request_repo_analysis in the diagnose-agent
// workflow's analyse_repo step. request_repo_analysis sends a Kafka request to
// the analyser adapter, which fetches + parses in ITS OWN pod and returns
// metadata (line spans, not bodies) — leaving the diagnose-agent pod with NO
// on-disk checkout, so repo_analysis.root is empty and diagnose_assemble_bundle's
// ReadSymbolBody has no file to slice (the "repo root not found" failure).
//
// This action does the fetch-and-parse IN-PROCESS, in the diagnose-agent's own
// pod, so the source is local and stays local:
//
//   1. Resolve owner/repo/ref (same config keys as request_repo_analysis), and
//      PIN ref to the commit code_symbols was indexed at, so the path:Symbol
//      entries lookup_code_symbols seeds from the index resolve in the fetched
//      tree (and the diagnosis reflects the indexed state).
//   2. Fetch the repo source to a LOCAL temp dir via the shared read-only tarball
//      fetcher (reposource.GitHubSource — one GET /repos/{o}/{r}/tarball/{ref},
//      no git binary; GITHUB_READ_TOKEN injected to diagnose-agent pods only).
//   3. analysis.AnalyseWithExclude(dir, exclude_patterns) IN-PROCESS -> Output,
//      whose Root IS that local dir. §7C.1: excludes default ["docs/"] so
//      archived code copies under docs/ never enter the index/bundle.
//   4. Return the Output (root = the local checkout) + commit_sha as repo_analysis.
//
// Result SHAPE is what the DOWNSTREAM diagnose consumers actually read: the
// analysis.Output fields are at the TOP LEVEL (so "repo_analysis.root" resolves
// and decoding "repo_analysis" as analysis.Output yields .files), with commit_sha
// (+ owner/repo/ref) added alongside. So diagnose_assemble_bundle (repo_root_field
// "repo_analysis.root", analysis_field "repo_analysis") and diagnose_route (call
// graph from "repo_analysis") read it UNCHANGED. This is NOT the analyser
// adapter's wrapped AnalyseResult shape ({..., output: Output}); the diagnose
// workflow never runs index_code_symbols (that lives in the separate code-indexer
// agent and still uses request_repo_analysis), so there is no shape conflict.
//
// The checkout is created ONCE — analyse_repo is the workflow's start_step (run
// once); the loop returns to load_runtime, NOT here — and is DELIBERATELY NOT
// cleaned up: it must persist for the life of the (ephemeral, per-orchestration)
// diagnose-agent pod so every iteration's ReadSymbolBody + call-graph re-scope
// reads from repo_analysis.root. Pod teardown reclaims the temp dir.
//
// SELF-CONTAINED at runtime: no analyser-adapter round-trip. (lookup_code_symbols
// still reads the code_symbols index for retrieval seeding — that index is the
// separate code-indexer's job; "self-contained" means no runtime analyser call,
// NOT independence from the shared index.)
//
// READ-ONLY: GETs a tarball, walks files, runs ONE read-only SELECT on
// code_symbols for the pinned commit. Writes nothing; triggers nothing.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gqls/agentchassis/internal/analysis"
	"github.com/gqls/agentchassis/internal/reposource"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// githubReadTokenEnv is the read-only, repo-scoped GitHub token. spawn_actions.go
// injects it into diagnose-agent pods ONLY (isRepoCloningAgent), via a secretKeyRef
// to personae-platform-secrets — the spawning chassis pod never holds it, and no
// other agent type receives it. Same credential the analyser adapter uses.
const githubReadTokenEnv = "GITHUB_READ_TOKEN"

// AnalyseRepoLocalInputSpec declares the action's contract. The identity keys
// mirror request_repo_analysis (owner/repo/ref/language via *_field paths or
// literals) so the workflow step is a near-drop-in swap; pin_to_index_commit is
// the only addition.
var AnalyseRepoLocalInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"owner_field", "repo_field", "ref_field", "language_field",
		"owner", "repo", "ref", "language",
		"github_api_base_field", "github_api_base",
		"pin_to_index_commit",
		"exclude_patterns",
	},
	Defaults: map[string]interface{}{
		"owner_field":         "input_data.owner",
		"repo_field":          "input_data.repo",
		"ref_field":           "input_data.ref",
		"language_field":      "input_data.language",
		"pin_to_index_commit": true,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("analyse_repo_local", AnalyseRepoLocalInputSpec)
}

// AnalyseRepoLocalAction fetches a repo to a local temp dir and analyses it
// in-process, returning the analyser Output (with a real local root) + commit_sha
// under the step's output_field (the workflow sets it to "repo_analysis").
func AnalyseRepoLocalAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Resolve identity the SAME way request_repo_analysis does (config.X_field path
	// into collected_data, or a literal config.X) — drop-in config compatibility.
	owner := resolveRAGConfigField(config, "owner_field", "owner", params.CollectedData)
	repo := resolveRAGConfigField(config, "repo_field", "repo", params.CollectedData)
	ref := resolveRAGConfigField(config, "ref_field", "ref", params.CollectedData)
	language := resolveRAGConfigField(config, "language_field", "language", params.CollectedData)
	if language == "" {
		language = "go"
	}
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("analyse_repo_local: owner and repo are required (set config.owner/repo or *_field paths)")
	}
	if language != "go" {
		// Polyglot is the analyser adapter's job; the in-process path is Go-only
		// (it calls the Go analysis library directly). For another language, route
		// that step back through request_repo_analysis instead.
		return nil, fmt.Errorf("analyse_repo_local: language %q not supported in-process (only 'go'; use request_repo_analysis for others)", language)
	}
	if ref == "" {
		ref = "HEAD"
	}

	// Pin to the commit the code_symbols index was built at, so the path:Symbol
	// entries lookup_code_symbols seeds from the index resolve in the fetched tree
	// (and the diagnosis reflects the indexed state). Best-effort: if the index is
	// empty or the read fails, fall back to ref. code_symbols.repo is the COMPOSED
	// "owner/repo" label (index_code_symbols composes owner + "/" + repo).
	if pinToIndexCommit(config) && params.DB != nil {
		if sha := indexCommitSHA(ctx, params.DB, owner+"/"+repo); sha != "" {
			logger.Info("analyse_repo_local: pinning to code_symbols commit",
				zap.String("ref_requested", ref), zap.String("commit_sha", sha))
			ref = sha
		} else {
			logger.Info("analyse_repo_local: no index commit to pin to; using ref as-is",
				zap.String("ref", ref))
		}
	}

	// Read-only token, injected to diagnose-agent pods only (spawn_actions.go).
	token := os.Getenv(githubReadTokenEnv)
	if token == "" {
		return nil, fmt.Errorf("analyse_repo_local: %s not set — this action must run in a SPAWNED repo-cloning agent pod (spawn_actions.go isRepoCloningAgent injects the secretKeyRef; the shared chassis pod deliberately never holds the token). If this fired on an agent-chassis pod, the workflow was adopted in-place: trigger via a spawning orchestrator (e.g. index-orchestrator) instead", githubReadTokenEnv)
	}
	apiBase := resolveRAGConfigField(config, "github_api_base_field", "github_api_base", params.CollectedData)

	src, err := reposource.NewGitHubSource(token, apiBase, logger)
	if err != nil {
		return nil, fmt.Errorf("analyse_repo_local: build source: %w", err)
	}

	// Fetch source to a LOCAL temp dir (tarball; read-only). NOTE: deliberately
	// NOT os.RemoveAll(dir) — the checkout must OUTLIVE this action so every loop
	// iteration's ReadSymbolBody + call-graph re-scope reads from repo_analysis.root.
	// The diagnose-agent pod is ephemeral (one orchestration) and its teardown
	// reclaims the dir; cleaning up here would dangle root on the next iteration.
	dir, commitSHA, err := src.FetchToDir(ctx, owner, repo, ref)
	if err != nil {
		return nil, fmt.Errorf("analyse_repo_local: fetch %s/%s@%s: %w", owner, repo, ref, err)
	}

	// Resolve the fetched snapshot's own committer date (and full sha) — the
	// self-contained fact the read-time freshness verdict keys on (bugs_open/108
	// defect A: row updated_at says when the INDEXER ran, never how old the code
	// is). Resolved against the sha the tarball actually yielded, not the
	// requested ref — the ref can move between the two calls; the sha cannot.
	// Best-effort: a failure leaves commit_time absent, so the indexer stores
	// NULL and the freshness banner reads UNKNOWN — the honest degrade — rather
	// than a date that was never fetched.
	var commitTime time.Time
	if fullSHA, committedAt, ciErr := src.CommitInfo(ctx, owner, repo, commitSHA); ciErr != nil {
		logger.Warn("analyse_repo_local: could not resolve commit date (freshness banner will read UNKNOWN)",
			zap.String("commit_sha", commitSHA), zap.Error(ciErr))
	} else {
		commitSHA = fullSHA
		commitTime = committedAt
	}

	// Parse IN-PROCESS. out.Root == dir (a real local path) — the whole point: it
	// makes repo_analysis.root a checkout THIS pod can slice bodies from.
	// §7C.1: denylist-style excludes (substring match inside the analyser walk),
	// default ["docs/"] — the §7C reindex showed archived code copies under
	// docs/agent_docs/… entering the index. Config-overridable per step; the
	// default lives in Go (defaultAnalyseExcludePatterns), NOT in the Defaults
	// map, so there is a single source of truth. configStringSlice is REUSED
	// from diagnose_load_runtime_action.go (same package).
	excludes := configStringSlice(config, "exclude_patterns", defaultAnalyseExcludePatterns)
	logger.Info("analyse_repo_local: analysing with excludes", zap.Strings("exclude_patterns", excludes))
	out, err := analysis.AnalyseWithExclude(dir, excludes)
	if err != nil {
		return nil, fmt.Errorf("analyse_repo_local: analyse %s (%s/%s@%s): %w", dir, owner, repo, ref, err)
	}

	logger.Info("analyse_repo_local: analysed in-process",
		zap.String("owner", owner), zap.String("repo", repo), zap.String("ref", ref),
		zap.String("commit_sha", commitSHA), zap.String("root", out.Root),
		zap.Int("file_count", out.FileCount))

	// Shape: the analysis.Output fields at the TOP LEVEL (so repo_analysis.root
	// resolves and decoding repo_analysis -> analysis.Output yields .files for the
	// call graph + body spans), plus commit_sha and owner/repo/ref. owner+repo are
	// included so the COMPOSING lookup_code_symbols (post repo-label patch) can read
	// repo_analysis.owner + repo_analysis.repo; extra keys are ignored when the map
	// is decoded back into analysis.Output downstream.
	repoAnalysis, err := outputToMap(out)
	if err != nil {
		return nil, fmt.Errorf("analyse_repo_local: encode output: %w", err)
	}
	repoAnalysis["commit_sha"] = commitSHA
	repoAnalysis["owner"] = owner
	repoAnalysis["repo"] = repo
	repoAnalysis["ref"] = ref
	if !commitTime.IsZero() {
		repoAnalysis["commit_time"] = commitTime.UTC().Format(time.RFC3339)
	}
	return repoAnalysis, nil
}

// pinToIndexCommit reads the optional bool config flag (default true). datahelpers
// exposes GetStringField/GetIntField in this codebase; there is no GetBoolField at
// time of writing, so coerce here. PRE-MERGE: if datahelpers gains GetBoolField,
// use it instead (grep: `grep -rn "func GetBoolField" platform/orchestration/datahelpers/`).
func pinToIndexCommit(config map[string]interface{}) bool {
	v, ok := config["pin_to_index_commit"]
	if !ok {
		return true // default on
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1" || t == "yes"
	default:
		return true
	}
}

// indexCommitSHA returns the dominant commit_sha recorded in code_symbols for the
// composed repo label ("owner/repo"), or "" if none. Dominant = the commit the
// bulk of the index was built at (robust to a partial re-index leaving a few rows
// on another commit). READ-ONLY single SELECT. A query error or no-rows is
// swallowed (best-effort pin) — the caller falls back to ref.
func indexCommitSHA(ctx context.Context, db *sql.DB, repoLabel string) string {
	var sha string
	err := db.QueryRowContext(ctx, `
		SELECT commit_sha
		FROM code_symbols
		WHERE repo = $1 AND commit_sha IS NOT NULL AND commit_sha <> ''
		GROUP BY commit_sha
		ORDER BY COUNT(*) DESC
		LIMIT 1`, repoLabel).Scan(&sha)
	if err != nil {
		return "" // no rows, or read failed — fall back to ref
	}
	return sha
}

// outputToMap renders analysis.Output to the map[string]interface{} shape
// collected_data uses (root/generated_at/file_count/files via the Output json
// tags), so commit_sha + owner/repo/ref can be added alongside and the result
// still decodes cleanly back into analysis.Output downstream (extra keys ignored).
func outputToMap(out analysis.Output) (map[string]interface{}, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// defaultAnalyseExcludePatterns backs the exclude_patterns config key (§7C.1).
// Denylist style, consistent with the bundle schema-section decision: new
// archive locations under docs/ never re-enter the index without a config edit.
var defaultAnalyseExcludePatterns = []string{"docs/"}
