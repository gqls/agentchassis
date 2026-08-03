// FILE: internal/agents/contentcreator/claims_guard.go
//
// The CLAIMS GUARD for free-standing generated text (bugs_open/123).
//
// WHY THIS FILE EXISTS. content-creator produces prose that belongs to no site
// and no page, so every enforcement surface the claims layer has is structurally
// out of reach: the deploy gate, the post-deploy sweep and the claims floor all
// resolve their evidence base by `site_id`. The floor says so itself, in its own
// scope boundary — `save_sections_claims_guard.go:81-82`:
//
//	"bugs_open/123's content-creator path has no site and no page row, so this
//	 seam cannot reach it at all."
//
// The consequence was measured, not assumed: a live generation on 2026-07-27
// published "Industry data shows that large language models experience
// hallucination rates between 3% and 10% depending on the task" — no source,
// invented range — and nothing in the pipeline objected, because nothing looked.
//
// IT RECORDS AND ANNOTATES. IT NEVER REFUSES. Three reasons, in order of weight:
//
//  1. The council's architecture seat has already ruled (council round
//     `7478233b`, recorded in LANDMINES.md's `DeliverReply` entry) that changing
//     these services' CALLER-OBSERVABLE FAILURE BEHAVIOUR "is an RFC moment".
//     Returning an error where prose was returned before is that same class of
//     change. Doing it here, under a different bug number, would be smuggling in
//     what that seat vetoed.
//  2. Both reply-delivery sites in this package are owned by the live
//     `bugs_open/158` lane ("contentcreator ×2"). This file touches neither
//     `sendSuccessResponse` nor `sendErrorResponse`, and does not adopt
//     `DeliverReply`.
//  3. The claims floor REFUSES a banned claim, and the symmetry is tempting —
//     but the floor sits on a PERSISTENCE seam, where refusing means "the
//     existing content stays". Here refusing means "the caller gets an error
//     instead of prose". Same severity vocabulary, different affordance.
//
// The floor's severity vocabulary is preserved so the record still says what a
// refusing seam WOULD have refused: banned claim → `error`, attributed-uncited
// statistic → `warning`. The stored `outcome` is always "recorded", written
// explicitly rather than inferred from the counts — the floor's own reasoning at
// save_sections_claims_guard.go:325-331: "blockers > 0 means it was refused" is
// true today and would silently stop being true the moment anyone adds a lever.
//
// If content-creator is ever put back on a publishing path, promoting `error`
// findings to a refusal is a deliberate decision for the RFC track — with the
// 158 lane's delivery work landed first — and not something this change
// pre-empts.

package contentcreator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// claimsGuardErrorCode is DELIBERATELY distinct from the gate's, the floor's and
// the link-repair one, for the reason recorded at
// save_sections_claims_guard.go:97-104: a shared code makes "which path caught
// this" unanswerable, and every query already written against the other codes
// must keep returning exactly what it returned before.
const claimsGuardErrorCode = "CONTENT_CREATOR_CLAIMS_DETAIL"

// claimsGuardToggles are the levers, resolved from the request.
type claimsGuardToggles struct {
	CheckClaims       bool
	CheckUncitedStats bool
}

// togglesFromRequest resolves the two levers.
//
// `check_claims` DEFAULTS ON, matching the name and default of the floor's lever
// (save_sections_claims_guard.go:200). It scans against the fleet-wide banned
// set, which is a set of statements no site may make about itself, and since
// this seam only records and annotates it grants no new authority over any
// caller — so ON by default does not collide with the owner ruling of
// 2026-08-02.
//
// `check_uncited_stats` DEFAULTS OFF, and that IS the owner ruling of
// 2026-08-02: new authority on a shared seam ships as an opt-in field with the
// unsafe default OFF. It is a heuristic over prose, and the caller who turns it
// on is the one who decided to accept its false positives.
func togglesFromRequest(req RequestPayload) claimsGuardToggles {
	t := claimsGuardToggles{CheckClaims: true}
	if req.Data.CheckClaims != nil {
		t.CheckClaims = *req.Data.CheckClaims
	}
	t.CheckUncitedStats = req.Data.CheckUncitedStats
	return t
}

// scanGeneratedText is the pure seam — no database, no logging, no config. It is
// where the behaviour is tested (claims_guard_test.go), the same split the claims
// floor uses for `scanSectionClaims`.
//
// THE SCAN SEMANTICS ARE NOT REIMPLEMENTED HERE AND MUST NEVER BE. This calls
// the identical engine the deploy gate, the claims floor and cmd/claimscan call.
// A second implementation of a claims scan is `bugs_open/093`'s shape exactly:
// "one call site of a shared judgement gets the rigorous fix; the sibling stays
// heuristic".
//
// The evidence base passed to ScanAllBannedClaims is nil, and that is correct
// rather than a shortcut: "a nil eb means 'this site has no register', not 'do
// not scan'" (claims_global.go:242-243). A site-less producer is the case that
// nil-safety was built for.
func scanGeneratedText(text, format string, t claimsGuardToggles) (banned, uncited []datahelpers.ClaimFinding) {
	if text == "" {
		return nil, nil
	}

	// Format-aware, because ExtractAssertionText is an HTML parser and this
	// agent emits markdown and plain text as well (agent.go's Format field).
	// Passing markdown to the HTML extractor returns ONE fused block for the
	// whole document — measured, see claims_plaintext.go's header.
	blocks := datahelpers.AssertionBlocks(text, format)
	if len(blocks) == 0 {
		return nil, nil
	}

	if t.CheckClaims {
		banned = datahelpers.ScanAllBannedClaims(blocks, nil)
	}
	if t.CheckUncitedStats {
		// The WHOLE document's blocks: this scan's unit is the document, and a
		// per-fragment call over-reports. See claims_attributed.go's header.
		uncited = datahelpers.ScanAttributedUncitedStats(blocks)
	}
	return banned, uncited
}

// recordClaimFindings writes the durable half.
//
// A pod log line is not a record — the floor's rule
// (save_sections_claims_guard.go:281-284), and it applies harder here: this
// producer appears in no per-agent usage query at all (it writes nothing to
// `llm_call_log`), so a pod log is the only trace there would otherwise be, and
// pod logs roll.
//
// `site_id` is NULL. That is not a workaround: the column is nullable and
// NULL-site rows are the established shape on this table (283 of 484 rows in the
// week to 2026-08-03). A producer with no site is precisely what this bug is
// about, and a record that had to invent a site would be a false one.
//
// Best-effort throughout: a failure to record must never fail a generation whose
// content is acceptable, and must never mask the finding it describes.
func (a *Agent) recordClaimFindings(
	ctx context.Context,
	headers map[string]string,
	req RequestPayload,
	banned, uncited []datahelpers.ClaimFinding,
) {
	if a.db == nil {
		// The DB is optional in NewAgent (it degrades to "no memory support"),
		// so this is a real branch, not a defensive one.
		a.logger.Warn("claims guard: no database handle — findings are in this log line only",
			zap.Int("banned", len(banned)),
			zap.Int("uncited_stats", len(uncited)))
		return
	}

	severity := "warning"
	if len(banned) > 0 {
		severity = "error"
	}

	contextJSON, err := json.Marshal(map[string]interface{}{
		"outcome":        "recorded", // never "refused" on this seam — see the file header
		"topic":          req.Data.Topic,
		"content_type":   req.Data.ContentType,
		"format":         req.Data.Format,
		"correlation_id": headers["correlation_id"],
		"request_id":     headers["request_id"],
		"banned":         banned,
		"uncited_stats":  uncited,
	})
	if err != nil {
		a.logger.Warn("claims guard: failed to marshal finding context", zap.Error(err))
		return
	}

	if _, err := a.db.Exec(ctx, `
		INSERT INTO agent_error_log
		    (site_id, agent_type, step_name, action, error_message, error_code, severity, context)
		VALUES (NULL, $1, 'generate_content', $2, $3, $4, $5, $6::jsonb)`,
		AgentType, req.Action,
		fmt.Sprintf("Claims guard recorded: %d banned claim(s), %d attributed-uncited statistic(s) on generated %s",
			len(banned), len(uncited), req.Data.ContentType),
		claimsGuardErrorCode, severity, string(contextJSON),
	); err != nil {
		a.logger.Warn("claims guard: failed to write finding record", zap.Error(err))
	}
}
