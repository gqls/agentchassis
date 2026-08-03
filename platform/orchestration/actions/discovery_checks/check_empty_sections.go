// FILE: platform/orchestration/actions/discovery_checks/check_empty_sections.go
//
// Detects page sections with empty or near-empty rendered HTML.
// Creates empty_section work items routed to page-build-handler.
//
// CHANGES:
//   - HandlerAgent: "page-build-handler" (was "page-content-writer")
//   - SQL excludes blog/blog-index pages (handled by check_empty_blog.go)
//   - Skips locked components (locked_at IS NULL filter)
//     Human-locked components are intentionally managed — don't flag them.
//   - Runtime-fill guard (2026-07-10): a section marked data-runtime-fill is
//     DELIBERATELY empty at build time — a browser-side loader fills it (same
//     exemption as sectionHasVisibleContent in rerender_single_page_action.go).
//     Its empty <h2> etc. are the shell the loader populates; routing it to
//     page-build-handler would bake build-time copy into the shell. First live
//     catch: vonc.com's provocation-card and lobby-grid, flagged as
//     'empty_heading' the first pass after enabling the new checks. Reported as
//     findings, never emitted.
//   - Retraction (2026-08-03): this check can now CLOSE empty_section items it
//     positively re-observes as fixed, via CheckResult.Resolved. It is the first
//     adopter of the RFC_010 seam that a real site exercises. Read
//     findResolvedEmptySections' comment before changing anything about it —
//     three separate refusals are load-bearing, and the obvious simplification
//     ("close what this run did not find") is the one the seam exists to forbid.
//
// Registration: automatic via init() → Register(&EmptySectionsCheck{})

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() {
	Register(&EmptySectionsCheck{})
	RegisterVerifier("empty_section", VerifyEmptySectionResolved)
}

type EmptySectionsCheck struct{}

func (c *EmptySectionsCheck) Name() string { return "empty_sections" }

func (c *EmptySectionsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	sections, err := findEmptySections(dctx)
	if err != nil {
		return nil, err
	}

	result := &CheckResult{}

	// ⚠ THE EARLY RETURN THAT USED TO LIVE HERE WAS THE WHOLE BUG IN THE OBVIOUS
	// ADOPTION. It read `if len(sections) == 0 { return &CheckResult{}, nil }`,
	// which is correct while a check can only ever FILE — nothing found, nothing
	// to say. The moment a check can also RETRACT it is exactly backwards: a site
	// with zero empty sections is the site whose stale empty_section items most
	// need closing, and it is the only site the early return can fire on. Bolting
	// the retraction onto the end of this function while leaving the guard in
	// place produces a mechanism that is inert precisely where it is needed and
	// looks fine in every test that has a finding. Kept as a comment because the
	// shape recurs: any monotonic check adopting Resolved has one of these.
	if len(sections) > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":      "empty_sections",
			"count":      len(sections),
			"components": sections,
		})
	}

	for _, section := range sections {
		// Runtime-fill shell: deliberately empty at build time; a loader fills it
		// in the browser. Surface it, never send it to the writer.
		if section.IsRuntimeFill {
			dctx.Logger.Info("empty_sections: runtime-fill shell exempt from rebuild",
				zap.String("page", section.PageName),
				zap.String("slot", section.SlotName),
				zap.String("pattern", section.EmptyPattern))
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     "empty_sections",
				"page":      section.PageName,
				"slot_name": section.SlotName,
				"skipped":   true,
				"reason":    "data-runtime-fill shell — empty by design, filled client-side",
			})
			continue
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":              "empty_sections",
			"component_id":       section.ComponentID,
			"page_id":            section.PageID,
			"page_name":          section.PageName,
			"slot_name":          section.SlotName,
			"component_function": section.ComponentFunction,
			"empty_pattern":      section.EmptyPattern,
		})

		var pageIDPtr *uuid.UUID
		if parsed, err := uuid.Parse(section.PageID); err == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "empty_section",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Empty section '%s' on page %s", section.SlotName, section.PageName),
			SpecJSON:     string(specJSON),
			PageID:       pageIDPtr,
			Priority:     100,
			HandlerAgent: "page-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("empty_section:%s:%s", section.PageID, section.SlotName),
			BatchID:      dctx.BatchID,
		})
	}

	// ── Retraction: say what was positively observed HEALTHY ────────────────
	//
	// A FAILED OBSERVATION MUST NEVER SUPPRESS A FINDING. Warn and carry on with
	// an empty Resolved rather than returning the error: the runner's `continue`
	// on error would drop this run's INSERTS too, so propagating a retraction
	// fault would trade a missed closure for a missed defect. Degrading to "did
	// not retract" leaves every item open, which is the safe direction.
	resolved, rerr := findResolvedEmptySections(dctx)
	if rerr != nil {
		dctx.Logger.Warn("empty_sections: retraction observation failed — findings unaffected, nothing retracted",
			zap.Error(rerr))
	} else {
		result.Resolved = append(result.Resolved, resolved...)
	}

	return result, nil
}

type emptySectionFinding struct {
	ComponentID       string `json:"component_id"`
	PageID            string `json:"page_id"`
	PageName          string `json:"page_name"`
	SlotName          string `json:"slot_name"`
	ComponentFunction string `json:"component_function"`
	HTMLLength        int    `json:"html_length"`
	EmptyPattern      string `json:"empty_pattern"`
	IsRuntimeFill     bool   `json:"is_runtime_fill"`
}

func findEmptySections(dctx DiscoveryCheckContext) ([]emptySectionFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.id, pc.page_id, p.name, COALESCE(pc.slot_name, ''),
		       COALESCE(cc.function, pc.slot_name, 'unknown'),
		       LENGTH(COALESCE(pc.rendered_html, '')),
		       CASE
		           WHEN pc.rendered_html IS NULL THEN 'null_html'
		           WHEN TRIM(pc.rendered_html) = '' THEN 'empty_html'
		           WHEN LENGTH(pc.rendered_html) < 50 THEN 'minimal_html'
		           WHEN pc.rendered_html ~* '<(h[1-6])[^>]*>\s*</\1>' THEN 'empty_heading'
		           WHEN pc.rendered_html ~* 'class="section[^"]*">\s*</div>' THEN 'empty_container'
		           ELSE 'near_empty'
		       END,
		       COALESCE(pc.rendered_html, '') LIKE '%data-runtime-fill%'
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components cc ON pc.component_id = cc.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')
		  AND COALESCE(cc.function, '') NOT IN ('header', 'footer', 'head-seo')
		  AND NOT (COALESCE(p.suppressed_sections, '[]'::jsonb) ? COALESCE(pc.slot_name, ''))
		  AND p.name NOT IN ('blog')
		  AND COALESCE(p.page_type, '') NOT IN ('blog-index')
		  AND (
		      pc.rendered_html IS NULL
		      OR TRIM(pc.rendered_html) = ''
		      OR LENGTH(pc.rendered_html) < 50
		      OR pc.rendered_html ~* '<(h[1-6])[^>]*>\s*</\1>'
		  )
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("empty_sections query failed: %w", err)
	}
	defer rows.Close()

	var findings []emptySectionFinding
	for rows.Next() {
		var f emptySectionFinding
		if err := rows.Scan(&f.ComponentID, &f.PageID, &f.PageName, &f.SlotName,
			&f.ComponentFunction, &f.HTMLLength, &f.EmptyPattern, &f.IsRuntimeFill); err != nil {
			dctx.Logger.Warn("Failed to scan empty section", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// ============================================================================
// Retraction (RFC_010 Decision 1 — CheckResult.Resolved)
// ============================================================================

// findResolvedEmptySections names the open empty_section findings this run has
// POSITIVELY RE-OBSERVED to be fixed. It is this check's adoption of the
// retraction seam (RFC_010, owner ruling 2026-08-02 "Decision 1: option 1"),
// which shipped with one adopter that is enabled by zero agents — so until a
// check like this one uses it, `items_resolved: 0` means "nobody adopted it",
// not "it is broken".
//
// ── IT ENUMERATES FROM THE ITEM SIDE, AND THAT IS THE SAFETY PROPERTY ───────
//
// The tempting implementation is "close every open empty_section item that is
// not in this run's findings". That is retraction by ABSENCE and it is the one
// thing the seam forbids: a check that errored, or that was silently blinded by
// a bad predicate, returns an empty finding set indistinguishable from a
// healthy site. Measured against the live queue 2026-08-03, the absence rule
// would have closed ELEVEN items it has no evidence about — ten whose slot
// holds no deployed component at all, and one mixed slot (below).
//
// So this walks the slots that empty_section items actually name for this site,
// and asks the database what is in each one NOW. Every retraction is a
// statement about a row that was read, never about a row that was missing.
//
// ── THE THREE REFUSALS, each of which is a real population ─────────────────
//
//   - NO DEPLOYED COMPONENT IN THE SLOT → no retraction. Absence here is
//     equally "fixed by removing the section" and "a rebuild silently deleted
//     it", which is this platform's most repeated failure (bugs_open/012, /021,
//     /032). VerifyEmptySectionResolved already refuses this exact case for the
//     completion gate; refusing it here too keeps one answer to one question.
//     10 of 47 open items sit in this state today.
//   - MIXED SLOT → no retraction. bugs_open/156 means (page_id, slot_name) is
//     not unique: one live page holds THREE deployed components in one slot. If
//     any of them still renders empty, the finding may still be true of that
//     one, so the whole slot is left alone. 1 of 47 sits here.
//   - STILL EMPTY → no retraction, obviously. 19 of 47.
//
// Leaving 17 items across 6 sites that are retractable on evidence.
//
// ── IT CARRIES NO STATUS VOCABULARY, DELIBERATELY ──────────────────────────
//
// There is no `status NOT IN (...)` filter here, and adding one would be a
// mistake worth naming. This package already holds TWO hand-rolled copies of
// that list and they ALREADY DISAGREE — check_truncated_component.go:163
// includes 'cancelled', check_component_template_corrupted.go:141 omits it.
// A third copy in a function whose job is CLOSING items is how a status list
// drift becomes a wrongly-closed defect. resolveWorkItems owns the predicate
// (workItemClosedStatuses); this function owns the observation. Reading a few
// already-closed rows costs a no-op UPDATE and nothing else — measured, the
// worst site names 14 distinct slots in its whole history.
//
// A malformed or absent spec page_id needs no guard for the same reason: it
// matches no component, so it yields no observation, so it retracts nothing.
func findResolvedEmptySections(dctx DiscoveryCheckContext) ([]ResolvedFinding, error) {
	// page_id is compared AS TEXT rather than cast to uuid: a cast would raise
	// on a malformed spec value and take the whole check down with it, and the
	// set is bounded at a few dozen keys per site, so the index is not the point.
	// page_id comes from the FIRST-CLASS COLUMN first, the spec only as a
	// fallback. Raised by the council's editquality seat (medium, correlation
	// 97923026) against the standing landmine about handlers that read one of
	// site_work_items' first-class columns while the creator populates the other.
	// The filing half above sets BOTH — WorkItemSpec.PageID and spec.page_id —
	// so today they cannot disagree, and they do not: measured over all 58
	// empty_section rows ever, 0 have a NULL column, 0 disagree, and that holds
	// for all 47 open ones. The COALESCE is here anyway because "measured empty"
	// is not "unreachable" — this lane was already wrong once in exactly that way
	// — and because preferring the column costs nothing while removing the whole
	// class: a future filing change that sets the column and forgets the spec key
	// would otherwise silently stop retracting, with no signal.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT k.item_key, pc.id, COALESCE(pc.rendered_html, '')
		FROM (
		    SELECT DISTINCT item_key,
		           COALESCE(page_id::text, spec->>'page_id') AS page_id,
		           COALESCE(spec->>'slot_name','')           AS slot_name
		      FROM site_work_items
		     WHERE site_id = $1
		       AND item_type = 'empty_section'
		       AND item_key IS NOT NULL
		) k
		LEFT JOIN page_components pc
		       ON pc.page_id::text = k.page_id
		      AND COALESCE(pc.slot_name, '') = k.slot_name
		      AND pc.build_status = 'deployed'
		ORDER BY k.item_key
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("empty_sections retraction query failed: %w", err)
	}
	defer rows.Close()

	type slotObservation struct {
		deployed int
		healthy  int
		detail   string
	}
	observed := map[string]*slotObservation{}
	var order []string

	for rows.Next() {
		var itemKey string
		var componentID sql.NullString
		var html string
		if err := rows.Scan(&itemKey, &componentID, &html); err != nil {
			dctx.Logger.Warn("empty_sections: failed to scan retraction row", zap.Error(err))
			continue
		}
		obs, seen := observed[itemKey]
		if !seen {
			obs = &slotObservation{}
			observed[itemKey] = obs
			order = append(order, itemKey)
		}
		if !componentID.Valid {
			continue // LEFT JOIN miss: the slot holds no deployed component
		}
		obs.deployed++
		// The SAME verdict function the completion gate uses. One predicate, two
		// consumers — the alternative is a second copy of "what counts as empty"
		// drifting against the first, which is the defect class RFC_009 was
		// written about in this very lane.
		if v := emptySectionVerdict(html); v.Resolved {
			obs.healthy++
			obs.detail = v.Detail
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("empty_sections retraction scan failed: %w", err)
	}

	var out []ResolvedFinding
	for _, itemKey := range order {
		obs := observed[itemKey]
		if obs.deployed == 0 || obs.healthy != obs.deployed {
			continue
		}
		reason := fmt.Sprintf(
			"re-observed healthy: all %d deployed component(s) in this slot render content (%s)",
			obs.deployed, obs.detail)
		// ItemKey, never AllOfType. The unit observed is one slot on one page,
		// which is exactly the unit the key addresses; a health probe answering a
		// whole item type at once (backend_unreachable) is a different shape and
		// this is not it.
		out = append(out, ResolvedFinding{
			ItemType: "empty_section",
			ItemKey:  itemKey,
			Reason:   reason,
		})
	}
	return out, nil
}

// ============================================================================
// Completion-time verifier (item_type "empty_section")
// ============================================================================

// emptyHeadingRe mirrors the SQL detection pattern '<(h[1-6])[^>]*>\s*</\1>'
// in findEmptySections. Go's RE2 has no backreferences, so the closing tag
// matches any h1-h6 — marginally broader than the SQL, and anything the
// broader form catches is still an empty heading shell.
var emptyHeadingRe = regexp.MustCompile(`(?i)<h[1-6][^>]*>\s*</h[1-6]>`)

// VerifyEmptySectionResolved re-runs the empty-section predicate for the
// single component named in the item spec. Registered for item_type
// "empty_section" so CompleteWorkItemAction can refuse to stamp 'complete'
// while the section still renders empty.
func VerifyEmptySectionResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	componentID, _ := target.Spec["component_id"].(string)
	if componentID == "" {
		return VerifyResult{}, fmt.Errorf("empty_section spec has no component_id")
	}
	if _, err := uuid.Parse(componentID); err != nil {
		return VerifyResult{}, fmt.Errorf("empty_section spec component_id %q: %w", componentID, err)
	}

	var html sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT rendered_html FROM page_components WHERE id = $1`, componentID,
	).Scan(&html)
	if err == sql.ErrNoRows {
		// AMBIGUOUS — never report success here (bugs_open/032).
		//
		// A missing page_components row is equally the signature of a genuine fix
		// (or a deliberate removal) and of this platform's most repeated failure:
		// a rebuild silently deleting the component. The two are indistinguishable
		// from this query, so reporting Resolved:true recorded content-loss
		// incidents as verified fixes — the exact blind spot that caused the
		// incidents this gate exists to catch (bugs_open/012, /021).
		//
		// An error rather than Resolved:false because the gate's documented policy
		// is to fail OPEN on verifier error (complete_work_item_verification.go:14):
		// the item still completes and nothing wedges, but result._verification
		// records that verification could not be made. A false success becomes a
		// visible unknown. Resolved:false would burn an attempt and, at
		// max_attempts, strand a legitimately-removed component's item in 'failed'.
		//
		// Deliberately left open: if the page still EXPECTS this component (a
		// plan_sections entry, a slot reference), absence is not ambiguous at all
		// — it is deletion, and Resolved:false would be the honest answer. That is
		// a better verdict and a bigger change; it belongs to the
		// empty_sections_loop_integrity thread, and this floor does not preclude it.
		return VerifyResult{}, fmt.Errorf(
			"cannot verify: component %s no longer exists (genuinely fixed or silently deleted — indistinguishable here)",
			componentID)
	}
	if err != nil {
		return VerifyResult{}, err
	}

	return emptySectionVerdict(html.String), nil
}

// emptySectionVerdict classifies rendered HTML with the same patterns as
// findEmptySections' SQL WHERE clause (null/empty, minimal, empty-heading)
// plus the data-runtime-fill exemption. Pure — unit tested.
func emptySectionVerdict(html string) VerifyResult {
	if strings.TrimSpace(html) == "" {
		return VerifyResult{Resolved: false, Detail: "rendered_html is still null/empty"}
	}
	if strings.Contains(html, "data-runtime-fill") {
		return VerifyResult{Resolved: true, Detail: "runtime-fill shell — empty by design, filled client-side"}
	}
	if len(html) < 50 {
		return VerifyResult{Resolved: false, Detail: "rendered_html still minimal (<50 chars)"}
	}
	if emptyHeadingRe.MatchString(html) {
		return VerifyResult{Resolved: false, Detail: "empty heading shell still present in rendered_html"}
	}
	return VerifyResult{Resolved: true, Detail: "section has rendered content"}
}
