// FILE: platform/orchestration/actions/retract_page_deployment_action.go
//
// UNPUBLISH — the page-level half of `bugs_open/098`.
//
// THE GAP THIS CLOSES, stated the way `bugs_closed/125` stated it: *the platform
// can publish a page but has no implemented way to unpublish one.* Archiving a
// page (`pages.status='archived'`) correctly removes it from every derivation
// AND from re-rendering — so the last HTML it ever rendered is frozen and keeps
// being served, and no repair path can reach it because every repair path skips
// archived rows on purpose. The live instance the bug was filed on:
// robot-hands.com/learning-center/index.html serves 200 with a listing written
// 2026-07-03 that links to a page serving 404.
//
// Retraction is possible AT ALL because the deploy chain is already reconciling
// on the far side: gqls/sites' deploy-to-b2 workflow runs
// `b2 sync --delete … && purge the Cloudflare zone` on every push. So removing
// the file from the repo removes it from B2 and from the edge. Nothing was ever
// missing downstream — only the ability to remove the file, which is
// git-adapter `delete_file` (added with this).
//
// WHY THIS IS AN ACTION AND NOT A STEP IN AN ARCHIVE HANDLER: nothing in this
// codebase archives a page. Checked across Go and all four frontends — there is
// no writer of `status='archived'` at all; archiving is a hand-run SQL
// operation (052's containment route R6, and the relojistas `contacto` repair).
// So there is no existing seam to hook, and the honest shape is a callable
// platform capability that an operator run, a work-item handler, or a future
// archive path can all reach.
//
// ────────────────────────────────────────────────────────────────────────────
// THE GUARDS ARE THE DESIGN. This deletes live files, so what it CANNOT do
// matters more than what it does.
//
//  1. Paths are only ever derived from `pages.url` through the shared
//     datahelpers.PageFilePathFromURL — the same function the publish side has
//     used since bugs_closed/125. So a retraction can only ever name a file a
//     publish could have written: never an asset, never a stylesheet, never
//     .github/. A page whose url does not designate a file of its own (a
//     fragment, a query, an off-origin url) is DECLINED, never sanitised —
//     125's live case is idea.uk's "/tools.html#audience-check", where
//     /tools.html is a DIFFERENT page's canonical url.
//
//  2. The action re-reads the page rows itself. A caller names pages; it never
//     names files. There is no config path by which a caller-supplied string
//     reaches the adapter as a deletion.
//
//  3. A page is REFUSED if its derived file path is also derived by an `active`
//     page on the same site. Two urls can collide onto one file ("/foo/" and
//     "/foo/index.html"), and retracting the archived one would delete the live
//     one's artefact. Measured fleet-wide 2026-08-02: 0 such collisions exist,
//     so this guard is INERT today — it ships because "no collision exists" and
//     "a collision cannot cause damage" are different properties, and RFC_010 is
//     open on precisely this class.
//
//  4. Only pages the platform says must not be served are eligible:
//     `status <> 'active'`. Never "delete every file with no matching page row"
//     — the repo legitimately holds files `pages` does not model (assets,
//     sitemap.xml, robots.txt, the 404 page).
//
// ────────────────────────────────────────────────────────────────────────────
// WHAT THIS DELIBERATELY DOES NOT DO: it does not clear `pages.deployed_at`.
// 098's candidate 2 asks for that, and the census says the premise behind it is
// wrong: of 49 non-test references, three read the column as HISTORY rather
// than as a boolean — reconcile_superseded_reviews_action.go:98
// (`deployed_at > GREATEST(wi.created_at, …)`), maintenance_actions.go:725
// (findStalePages), and page_admin_handlers.go:101 (shown to a human). Clearing
// it changes what all three return, which is a change to what a shared column
// GUARANTEES — architecture scope by the owner ruling of 2026-07-29 §1, and
// exactly the ground bugs_closed/124 drew a REJECTED verdict on. It is recorded
// on the bug as measured-and-deferred, with the census.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RetractPageDeploymentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"site_id", "site_id_field", "domain_field", "page_ids_field", "repo_name", "dry_run"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("retract_page_deployment", RetractPageDeploymentInputSpec)
}

// retractionCandidate is one page considered for retraction, and the reason if
// it was not retracted. Refusals are RETURNED, not swallowed: a page the
// platform declines to retract is a page that keeps serving, and an operator
// reading a green result must be able to see which ones those are.
type retractionCandidate struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
	Refused  string `json:"refused,omitempty"`
}

// RetractPageDeploymentAction removes the deployed artefacts of pages that must
// no longer be served, via the git-adapter's delete_file verb.
//
// Selection is either an explicit page-id list (page_ids_field) or, failing
// that, every non-active page on the site that still carries a deployed_at
// stamp — the exact population 098 measures (13 fleet-wide on 2026-08-02).
func RetractPageDeploymentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "retract_page_deployment"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("retract_page_deployment requires a database handle")
	}

	siteID, err := resolveRetractionSiteID(params, config)
	if err != nil {
		return nil, err
	}

	var domain string
	if err := params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("resolve domain for site %s: %w", siteID, err)
	}
	if strings.TrimSpace(domain) == "" {
		return nil, fmt.Errorf("site %s has no domain; a retraction cannot be scoped", siteID)
	}

	pageIDs := extractRetractionPageIDs(params, config)
	candidates, err := loadRetractionCandidates(ctx, params.DB, siteID, pageIDs)
	if err != nil {
		return nil, err
	}

	// Guard 3: the file paths an ACTIVE page on this site owns are off limits,
	// whatever any other row says. Built once, from the same derivation.
	activePaths, err := loadActivePageFilePaths(ctx, params.DB, siteID)
	if err != nil {
		return nil, err
	}

	var paths []string
	seen := map[string]bool{}
	for i := range candidates {
		c := &candidates[i]

		// Guard 4, restated at the row: only a page the platform says must not
		// be served. loadRetractionCandidates already filters, but an explicit
		// page-id list is caller input and gets checked again here.
		if c.Status == "active" {
			c.Refused = "page is active — retracting a live page is not what archiving means"
			continue
		}

		fp, ok := datahelpers.PageFilePathFromURL(c.URL)
		if !ok {
			// Guard 1. Declined, never guessed: a page whose url does not name
			// a file of its own has no artefact this action can identify, and
			// falling back to its NAME is what published 125's orphans.
			c.Refused = "url does not designate a file of its own (fragment, query, or off-origin) — no path can be derived safely"
			continue
		}
		c.FilePath = fp

		if owner, clash := activePaths[fp]; clash {
			c.Refused = fmt.Sprintf("file path is also owned by ACTIVE page %q — retracting it would delete a live page's artefact", owner)
			continue
		}
		if seen[fp] {
			c.Refused = "duplicate of another candidate's file path in this batch"
			continue
		}
		seen[fp] = true
		paths = append(paths, fp)
	}
	sort.Strings(paths)

	result := map[string]interface{}{
		"site_id":    siteID.String(),
		"domain":     domain,
		"considered": len(candidates),
		"candidates": candidates,
		"paths":      paths,
		"retracted":  len(paths),
	}

	if dry, _ := config["dry_run"].(bool); dry {
		result["dry_run"] = true
		result["await_response"] = false
		logger.Info("retract_page_deployment: dry run",
			zap.String("domain", domain), zap.Strings("paths", paths))
		return result, nil
	}

	if len(paths) == 0 {
		// Nothing to do is a real, common, correct outcome — say so rather than
		// dispatching an empty request the adapter would have to reject.
		result["await_response"] = false
		result["status"] = "nothing_to_retract"
		logger.Info("retract_page_deployment: nothing to retract",
			zap.String("domain", domain), zap.Int("considered", len(candidates)))
		return result, nil
	}

	requestID, err := sendGitDeleteFileRequest(ctx, params, config, domain, paths, logger)
	if err != nil {
		return nil, err
	}
	result["request_id"] = requestID
	result["await_response"] = true
	return result, nil
}

// resolveRetractionSiteID takes the site from config, from a configured field,
// or from the usual collected-data locations. A retraction with no site is
// refused rather than defaulted: an unscoped deletion in a repo that holds every
// site is the one mistake this action must not be able to make.
func resolveRetractionSiteID(params ActionParams, config map[string]interface{}) (uuid.UUID, error) {
	if raw, ok := config["site_id"].(string); ok && raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("config site_id %q is not a uuid: %w", raw, err)
		}
		return id, nil
	}
	fields := []string{"site_record.id", "input_data.site_id", "site_id"}
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		fields = append([]string{f}, fields...)
	}
	for _, f := range fields {
		if v := datahelpers.ExtractNestedFieldString(params.CollectedData, f); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				continue
			}
			return id, nil
		}
	}
	return uuid.Nil, fmt.Errorf("no site_id resolved — a retraction must be scoped to one site")
}

// extractRetractionPageIDs reads an optional explicit page list. Empty means
// "every non-active page on the site that is still stamped as deployed".
func extractRetractionPageIDs(params ActionParams, config map[string]interface{}) []string {
	field, _ := config["page_ids_field"].(string)
	if field == "" {
		return nil
	}
	raw := datahelpers.ExtractNestedField(params.CollectedData, field)
	var out []string
	switch v := raw.(type) {
	case []string:
		out = append(out, v...)
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case string:
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func loadRetractionCandidates(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageIDs []string) ([]retractionCandidate, error) {
	// The default population is 098's own measurement query: not active, and
	// still carrying a deploy stamp. `deployed_at IS NOT NULL` is used here as
	// the bug documents it — "a deploy happened once", i.e. an artefact may
	// exist — NOT as proof it is fetchable. Whether it is actually still there
	// is settled by the adapter's per-path existence check, which reports an
	// absent path as a success rather than a failure.
	query := `
		SELECT id::text, COALESCE(name,''), COALESCE(url,''), COALESCE(status,'')
		  FROM pages
		 WHERE site_id = $1
		   AND COALESCE(status,'') <> 'active'
		   AND deployed_at IS NOT NULL`
	args := []interface{}{siteID}
	if len(pageIDs) > 0 {
		// An explicit list still may not reach outside the site or pick up an
		// active page — it narrows the population, it never widens it.
		query = `
			SELECT id::text, COALESCE(name,''), COALESCE(url,''), COALESCE(status,'')
			  FROM pages
			 WHERE site_id = $1
			   AND id::text = ANY($2)`
		args = append(args, pageIDs)
	}
	query += ` ORDER BY url`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load retraction candidates: %w", err)
	}
	defer rows.Close()

	var out []retractionCandidate
	for rows.Next() {
		var c retractionCandidate
		if err := rows.Scan(&c.PageID, &c.PageName, &c.URL, &c.Status); err != nil {
			return nil, fmt.Errorf("scan retraction candidate: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// loadActivePageFilePaths maps every file path an ACTIVE page on this site
// derives, to that page's name. The derivation is the shared one, so this
// answers the question that matters — "would deleting this file remove a live
// page's artefact?" — rather than the weaker url-equality question.
func loadActivePageFilePaths(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(name,''), COALESCE(url,'')
		  FROM pages
		 WHERE site_id = $1 AND COALESCE(status,'') = 'active'`, siteID)
	if err != nil {
		return nil, fmt.Errorf("load active page paths: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return nil, fmt.Errorf("scan active page: %w", err)
		}
		if fp, ok := datahelpers.PageFilePathFromURL(url); ok {
			out[fp] = name
		}
	}
	return out, rows.Err()
}

// sendGitDeleteFileRequest dispatches the retraction to the git-adapter. The
// envelope mirrors GitCommitAction's — same topic, same headers, same
// deploy-target resolution — because a retraction must land in exactly the repo
// the publish landed in.
//
// Branch is NOT set. gqls/sites has no `main` (it carries `master`, its default,
// and `750start`) while sites.github_branch says 'main' for most rows, and the
// B2 deploy workflow triggers on `master`. Deploys work because the commit path
// leaves the branch unset and CommitToRepo falls back to the repo default;
// a retraction that helpfully passed the column would target a branch that does
// not exist.
func sendGitDeleteFileRequest(ctx context.Context, params ActionParams, config map[string]interface{}, domain string, paths []string, logger *zap.Logger) (string, error) {
	if params.Producer == nil {
		return "", fmt.Errorf("kafka producer not available")
	}

	correlationID := getExecutionContextValue(params, "correlation_id", params.ExecutionContext.CorrelationID)
	orchestrationID := getExecutionContextValue(params, "orchestration_id", params.ExecutionContext.OrchestrationID)
	clientID := getExecutionContextValue(params, "client_id", params.ExecutionContext.ClientID)
	responsesTopic := getExecutionContextValue(params, "responses_topic", params.ExecutionContext.ResponsesTopic)
	if responsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			responsesTopic = alt
		}
	}
	if responsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			responsesTopic = parent
		}
	}

	repoName := resolveGitRepoNameDB(ctx, params.DB, config, params.CollectedData, domain, logger)
	newRequestID := uuid.New().String()

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":         correlationID,
			"orchestration_id":       orchestrationID,
			"client_id":              clientID,
			"step_name":              params.ExecutionContext.StepName,
			"step_id":                params.ExecutionContext.StepID,
			"request_id":             newRequestID,
			"message_type":           "request",
			"sender_agent_type":      params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":        orchestrationID,
			"sender_pod_name":        params.ExecutionContext.Sender.PodName,
			"responses_topic":        responsesTopic,
			"parent_responses_topic": responsesTopic,
			"timestamp":              time.Now().UTC().Format(time.RFC3339),
			"action":                 "delete_file",
		},
		"body": map[string]interface{}{
			"action": "delete_file",
			"data": map[string]interface{}{
				"repo_name": repoName,
				"domain":    domain,
				"paths":     paths,
				"commit_message": fmt.Sprintf("Retract %d retired page(s) from %s (bugs_open/098)",
					len(paths), domain),
			},
			"reply_to_topic":         responsesTopic,
			"parent_responses_topic": responsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   orchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
				"domain":                domain,
				"paths_count":           len(paths),
			},
			"request_context": map[string]interface{}{
				"correlation_id":   correlationID,
				"orchestration_id": orchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return "", fmt.Errorf("marshal delete_file request: %w", err)
	}
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string, len(rawHeaders))
	for k, v := range rawHeaders {
		headers[k] = fmt.Sprintf("%v", v)
	}

	if err := params.Producer.ProduceWithValidation(ctx, "system.adapter.git.requests",
		headers, []byte(correlationID), messageBytes); err != nil {
		return "", fmt.Errorf("send delete_file to git-adapter: %w", err)
	}

	logger.Info("retract_page_deployment: delete_file sent, awaiting response",
		zap.String("repo_name", repoName),
		zap.String("domain", domain),
		zap.Strings("paths", paths),
		zap.String("request_id", newRequestID))

	return newRequestID, nil
}
