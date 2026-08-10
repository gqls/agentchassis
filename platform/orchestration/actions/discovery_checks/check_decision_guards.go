// FILE: platform/orchestration/actions/discovery_checks/check_decision_guards.go
//
// RFC_015 stage 2: decision GUARDS. A decision record (doc_notes, categories
// ? 'decision-record' — see decision_guard.go on why NOT bare 'decision') may
// carry a fenced ```guard block asserting an OUTCOME on a
// page. This check evaluates every guard for the site and files a
// decision_regression work item when one stops being true — the decision
// pins the WHAT; how the page achieves it stays free to improve.
//
//	```guard
//	{"page": "index", "assert": "contains", "pattern": "href=\"/tools.html#audience-check\""}
//	```
//
// assert: "contains" | "not_contains" — case-insensitive substring over the
// page's STORED assembly (chrome + page_components.rendered_html in position
// order). Stored, not served, deliberately: every automated writer writes
// stored content and deploys assemble from it, so a regression lands here
// first; stored-vs-served drift is a different failure class with its own
// checks. (The known landmine "phantom check reads STORED html" is about
// claiming the SERVED page is fine — this check never claims that.)
//
// A decision with no guard fence, a malformed fence, or a page that does not
// exist checks nothing — guards are opt-in per decision, and this check is
// inert on sites with no decision rows.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

func init() {
	Register(&DecisionGuardsCheck{})
	// The completion verifier. Without one, an item is stamped 'complete' on the
	// handler's own word — and decision_regression is filed at
	// needs_human_review, so "the handler" is a person asserting they fixed it.
	// This lane owes this registration: shipping the item_type on 2026-08-08
	// with no classification left verifier_coverage_test RED at HEAD for two
	// days, and the bugfix_236 lane classified it in passing so the package
	// suite could run at all, explicitly leaving the verifier decision here.
	//
	// It is buildable precisely because the guard predicate is re-runnable: the
	// same assertion over the same stored assembly. That is not true of most
	// item types, which is why so few carry verifiers.
	RegisterVerifier("decision_regression", VerifyDecisionRegressionResolved)
}

type DecisionGuardsCheck struct{}

func (c *DecisionGuardsCheck) Name() string { return "decision_guards" }

var guardFenceRe = regexp.MustCompile("(?s)```guard\\s*\\n(.*?)```")

type decisionGuard struct {
	Page    string `json:"page"`
	Assert  string `json:"assert"`
	Pattern string `json:"pattern"`
}

// storedPageAssembly is the single definition of "the page, as stored" for this
// check AND its verifier. Extracted so the two cannot drift: a verifier whose
// predicate differs from its check's is worse than no verifier, because it
// retracts findings the check would still make (or refuses to retract ones it
// would not).
const storedPageAssemblySQL = `
	SELECT COALESCE(pg.rendered_header,'') || COALESCE(pg.rendered_footer,'') ||
	       COALESCE((SELECT string_agg(COALESCE(pc.rendered_html,''), '' ORDER BY pc.position)
	                 FROM page_components pc WHERE pc.page_id = pg.id), '')
	FROM pages pg
	WHERE pg.site_id = $1 AND pg.name = $2
`

// decisionGuardViolated is the predicate itself, shared for the same reason.
// Case-insensitive substring; an unrecognised assert verb is NOT a violation
// (an unknown verb must not manufacture findings — the fence author gets a
// silent no-op, which is the conservative direction for a vocabulary that may
// grow).
func decisionGuardViolated(assembled, assert, pattern string) bool {
	found := strings.Contains(strings.ToLower(assembled), strings.ToLower(pattern))
	switch assert {
	case "contains":
		return !found
	case "not_contains":
		return found
	default:
		return false
	}
}

func (c *DecisionGuardsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT subject_key, body
		FROM doc_notes
		WHERE site_id = $1
		  AND categories ? 'decision-record'
	`, dctx.SiteID)
	if err != nil {
		dctx.Logger.Warn("DecisionGuardsCheck: decision query failed", zap.Error(err))
		return &CheckResult{}, nil
	}
	defer rows.Close()

	type pending struct {
		key   string
		guard decisionGuard
	}
	var guards []pending
	for rows.Next() {
		var key, body string
		if err := rows.Scan(&key, &body); err != nil {
			return &CheckResult{}, nil
		}
		m := guardFenceRe.FindStringSubmatch(body)
		if m == nil {
			continue
		}
		var g decisionGuard
		if json.Unmarshal([]byte(strings.TrimSpace(m[1])), &g) != nil ||
			g.Page == "" || g.Pattern == "" ||
			(g.Assert != "contains" && g.Assert != "not_contains") {
			dctx.Logger.Warn("DecisionGuardsCheck: malformed guard fence — decision steers but cannot guard",
				zap.String("decision", key))
			continue
		}
		guards = append(guards, pending{key: key, guard: g})
	}
	if len(guards) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}
	for _, p := range guards {
		var assembled string
		err := dctx.DB.QueryRowContext(dctx.Ctx, `
			`+storedPageAssemblySQL+`
		`, dctx.SiteID, p.guard.Page).Scan(&assembled)
		if err != nil {
			// Page absent: not this check's finding — page-existence checks own it.
			continue
		}

		if !decisionGuardViolated(assembled, p.guard.Assert, p.guard.Pattern) {
			continue
		}

		dctx.Logger.Warn("DecisionGuardsCheck: decision regression detected",
			zap.String("decision", p.key),
			zap.String("page", p.guard.Page),
			zap.String("assert", p.guard.Assert))

		specJSON := fmt.Sprintf(
			`{"check":"decision_guards","decision":%q,"page":%q,"assert":%q,"pattern":%q}`,
			p.key, p.guard.Page, p.guard.Assert, p.guard.Pattern)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":    "decision_guards",
			"decision": p.key,
			"page":     p.guard.Page,
			"assert":   p.guard.Assert,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "decision_regression",
			Severity: "high",
			Summary: fmt.Sprintf("Decision %s regressed: page %q fails %s %q — restore the outcome or supersede the decision by name",
				p.key, p.guard.Page, p.guard.Assert, p.guard.Pattern),
			SpecJSON:     specJSON,
			Priority:     80,
			HandlerAgent: "",
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			// KEYED BY PAGE AS WELL AS DECISION. A covers-fence may span many
			// pages — D-004 names nine — and one decision can carry a guard per
			// page. Keyed on decision alone, a violation on page A files the
			// item and a later violation on page B under the same decision
			// collides on item_key and is silently dropped by the dedup path:
			// the fleet's "idx_swi_dedup: KEY coarser than FINDING" failure
			// mode. Raised by the council's editquality seat on the first
			// RFC_015 round (2026-08-09), whose verdict this lane did not read
			// at the time because the round looked dropped.
			ItemKey: fmt.Sprintf("decision_regression:%s:%s:%s", dctx.SiteID, p.key, p.guard.Page),
			BatchID: dctx.BatchID,
		})
	}
	return result, nil
}

// VerifyDecisionRegressionResolved re-runs the guard the item was filed for and
// resolves only if the assertion holds again.
//
// FAIL-CLOSED on anything it cannot evaluate (RFC_017, owner ruling 2026-08-08):
// a malformed spec, a missing page, or a DB error all return an error, so
// CompleteWorkItemAction refuses the completion rather than taking a handler's
// word. The unsafe direction here is stamping a decision regression 'complete'
// while the decision is still regressed — that is precisely the silent-loss
// shape this whole mechanism exists to prevent, so it does not get the benefit
// of the doubt. No VerifierPolicy override: FailOpenOnError stays OFF.
//
// A missing page is deliberately an ERROR and not "resolved". A decision whose
// covered page has vanished is a real question for a human — the outcome the
// decision pins can hardly be true on a page that does not exist — and silently
// completing the item would erase the only record that the question arose.
func VerifyDecisionRegressionResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	page, _ := target.Spec["page"].(string)
	assert, _ := target.Spec["assert"].(string)
	pattern, _ := target.Spec["pattern"].(string)
	decision, _ := target.Spec["decision"].(string)

	if page == "" || assert == "" || pattern == "" {
		return VerifyResult{}, fmt.Errorf("decision_regression spec incomplete (decision=%q page=%q assert=%q pattern set=%t) — cannot re-run the guard",
			decision, page, assert, pattern != "")
	}
	if assert != "contains" && assert != "not_contains" {
		return VerifyResult{}, fmt.Errorf("decision_regression spec has unknown assert %q for decision %q — cannot re-run the guard", assert, decision)
	}

	var assembled string
	if err := db.QueryRowContext(ctx, storedPageAssemblySQL, target.SiteID, page).Scan(&assembled); err != nil {
		return VerifyResult{}, fmt.Errorf("re-assembling stored page %q for decision %q: %w", page, decision, err)
	}

	if decisionGuardViolated(assembled, assert, pattern) {
		return VerifyResult{
			Resolved: false,
			Detail: fmt.Sprintf("decision %s still regressed: page %q fails %s %q against %d bytes of stored assembly",
				decision, page, assert, pattern, len(assembled)),
		}, nil
	}
	return VerifyResult{
		Resolved: true,
		Detail: fmt.Sprintf("decision %s holds again: page %q satisfies %s %q (%d bytes of stored assembly)",
			decision, page, assert, pattern, len(assembled)),
	}, nil
}
