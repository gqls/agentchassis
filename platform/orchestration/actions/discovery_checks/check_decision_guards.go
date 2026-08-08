// FILE: platform/orchestration/actions/discovery_checks/check_decision_guards.go
//
// RFC_015 stage 2: decision GUARDS. A decision record (doc_notes, categories
// ? 'decision') may carry a fenced ```guard block asserting an OUTCOME on a
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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&DecisionGuardsCheck{}) }

type DecisionGuardsCheck struct{}

func (c *DecisionGuardsCheck) Name() string { return "decision_guards" }

var guardFenceRe = regexp.MustCompile("(?s)```guard\\s*\\n(.*?)```")

type decisionGuard struct {
	Page    string `json:"page"`
	Assert  string `json:"assert"`
	Pattern string `json:"pattern"`
}

func (c *DecisionGuardsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT subject_key, body
		FROM doc_notes
		WHERE site_id = $1
		  AND categories ? 'decision'
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
			SELECT COALESCE(pg.rendered_header,'') || COALESCE(pg.rendered_footer,'') ||
			       COALESCE((SELECT string_agg(COALESCE(pc.rendered_html,''), '' ORDER BY pc.position)
			                 FROM page_components pc WHERE pc.page_id = pg.id), '')
			FROM pages pg
			WHERE pg.site_id = $1 AND pg.name = $2
		`, dctx.SiteID, p.guard.Page).Scan(&assembled)
		if err != nil {
			// Page absent: not this check's finding — page-existence checks own it.
			continue
		}

		found := strings.Contains(strings.ToLower(assembled), strings.ToLower(p.guard.Pattern))
		violated := (p.guard.Assert == "contains" && !found) ||
			(p.guard.Assert == "not_contains" && found)
		if !violated {
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
			ItemKey:      fmt.Sprintf("decision_regression:%s:%s", dctx.SiteID, p.key),
			BatchID:      dctx.BatchID,
		})
	}
	return result, nil
}
