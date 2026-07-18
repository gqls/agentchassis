// FILE: platform/orchestration/actions/diagnose_read_repo_files_action.go
//
// F1.1b(c): fetch the CURRENT bodies of the approved plan's files so
// sketch_to_files rewrites real code instead of reconstructing files it has
// never seen — whole-file output from symbol snippets would be hallucination
// by construction.
//
// Uses the GitHub CONTENTS API with the raw media type (no git binary in the
// chassis image; no base64 size quirks), authenticated by GITHUB_READ_TOKEN
// from the pod env — which exists ONLY because fix-implementer is on the
// isRepoCloningAgent spawn gate. Read-only by construction: the token is the
// read token, and this action only GETs.
//
// Rules: a 'modify' file that 404s is a HARD error (the plan names a file
// that does not exist at the ref — implementing it would fabricate context);
// an 'add' file is EXPECTED to be absent and is skipped (nothing to read).
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseReadRepoFilesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"plan_field", "repo_owner", "repo_name", "ref", "ref_field", "max_file_bytes"},
	Defaults: map[string]interface{}{
		"plan_field":     "plan_row.body",
		"repo_owner":     "gqls",
		"repo_name":      "agentchassis",
		"ref":            "main",
		"max_file_bytes": 400_000, // a plan-named file bigger than this is not a "constrained" edit target
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_read_repo_files", DiagnoseReadRepoFilesInputSpec)
}

// DiagnoseReadRepoFilesAction reads the plan's modify/add files at a ref.
func DiagnoseReadRepoFilesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_read_repo_files"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	token := os.Getenv("GITHUB_READ_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_READ_TOKEN not in env — is this agent on the isRepoCloningAgent spawn gate?")
	}

	planRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(config, "plan_field", "plan_row.body"))
	if planRaw == nil {
		return nil, fmt.Errorf("plan missing at plan_field")
	}
	pb, err := planBytes(planRaw)
	if err != nil {
		return nil, fmt.Errorf("plan not serialisable: %w", err)
	}
	var plan planLite
	if err := json.Unmarshal(pb, &plan); err != nil {
		return nil, fmt.Errorf("plan does not match the fix-plan schema: %w", err)
	}

	// modify → must exist; add → expected absent; config_change → not a repo file.
	type target struct {
		path     string
		mustRead bool
	}
	var targets []target
	for _, e := range plan.Edits {
		op := strings.ToLower(strings.TrimSpace(e.Operation))
		switch op {
		case "modify":
			targets = append(targets, target{strings.TrimSpace(e.File), true})
		case "add":
			targets = append(targets, target{strings.TrimSpace(e.File), false})
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("plan has no modify/add edits — nothing to read")
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].path < targets[j].path })

	owner := datahelpers.GetStringField(config, "repo_owner", "gqls")
	repo := datahelpers.GetStringField(config, "repo_name", "agentchassis")
	ref := datahelpers.GetStringField(config, "ref", "main")
	// Stage-loop seam (delta 2): a routed per-stage ref wins when configured —
	// stage 1 reads the base ref, later stages read the feat/* branch so they
	// see earlier stages' commits.
	//
	// CONFIGURED-BUT-EMPTY IS AN ERROR, not a fallback (council-gate review
	// 5a65ec4c, bug-historian): silently keeping the literal here means stage N
	// reads `main` instead of the branch carrying stage N-1's commits, and the
	// implementer then rewrites files from the wrong tree with no error — the
	// platform's worst-known failure shape. Not configuring ref_field at all
	// remains the single-plan path and keeps the literal.
	if rf := datahelpers.GetStringField(config, "ref_field", ""); rf != "" {
		v := strings.TrimSpace(datahelpers.ExtractNestedFieldString(params.CollectedData, rf))
		if v == "" {
			return nil, fmt.Errorf("ref_field %q is configured but resolved empty — refusing to fall back to %q, which would read the wrong tree", rf, ref)
		}
		ref = v
	}
	maxBytes := datahelpers.GetIntField(config, "max_file_bytes", 400_000)
	client := &http.Client{Timeout: 20 * time.Second}

	files := map[string]interface{}{}
	var rendered strings.Builder
	for _, tg := range targets {
		body, status, err := fetchRepoFileRaw(ctx, client, token, owner, repo, tg.path, ref)
		switch {
		case err != nil:
			return nil, fmt.Errorf("read %s@%s: %w", tg.path, ref, err)
		case status == http.StatusNotFound && !tg.mustRead:
			// an 'add' target — absent as expected
			continue
		case status == http.StatusNotFound:
			return nil, fmt.Errorf("plan names %s (modify) but it does not exist at %s/%s@%s — refusing to fabricate context", tg.path, owner, repo, ref)
		case status != http.StatusOK:
			return nil, fmt.Errorf("read %s@%s: HTTP %d", tg.path, ref, status)
		case len(body) > maxBytes:
			return nil, fmt.Errorf("%s is %d bytes (> max_file_bytes %d) — too large for a constrained whole-file rewrite", tg.path, len(body), maxBytes)
		}
		files[tg.path] = string(body)
		fmt.Fprintf(&rendered, "=== %s (current, @%s) ===\n%s\n\n", tg.path, ref, string(body))
	}

	logger.Info("diagnose_read_repo_files: fetched",
		zap.Int("files", len(files)),
		zap.String("ref", ref),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"files":    files,             // path → current body (originals for prepare's no-op check)
		"rendered": rendered.String(), // prompt-ready text for sketch_to_files
		"count":    len(files),
	}, nil
}

// fetchRepoFileRaw GETs one file via the contents API with the raw media type.
func fetchRepoFileRaw(ctx context.Context, client *http.Client, token, owner, repo, path, ref string) ([]byte, int, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		owner, repo, escapePathSegments(path), url.QueryEscape(ref))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // hard 2MB backstop
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// escapePathSegments escapes each path segment but keeps the '/' separators.
func escapePathSegments(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return strings.Join(segs, "/")
}
