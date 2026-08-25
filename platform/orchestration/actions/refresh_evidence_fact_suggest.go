// FILE: platform/orchestration/actions/refresh_evidence_fact_suggest.go
//
// bugs_open/288 Phase 4 — ADOPTION. The mechanism cannot protect what nobody
// declares, and measured 2026-08-24 exactly ONE of 132 current tool PLANs
// declares anything — and that one only because this lane asked another lane to
// hand-seed it. 178 tool pages sit on sites that have a register. A mechanism
// gated on hand-authoring covers ~0% of the fleet, for ever.
//
// So: the machine proposes the binding, a human decides it. For each
// register-bearing site this finds tool subjects that do NOT declare, probes
// their script text for the site's own registered values, and files one doc_note
// per tool carrying a paste-ready `"facts": [...]` fragment.
//
// ── WHY THIS IS NOT THE SCAN PLAN_2026-08-09 §3 REFUSED ────────────────────
//
// That §3 refused "a static scan of tool JavaScript for UNREGISTERED numbers",
// on the ground that every constant — 12, 100, 0.01, a viewport width — would
// need whitelisting, and the false-positive rate would train people to ignore
// the alarm. The direction here is inverted, and that is the whole difference:
//
//	that scan: unbounded target set (every number in the file), needs a whitelist,
//	           output is an ALARM, a wrong one costs credibility
//	this one:  finite per-site target set (the register's own values, already
//	           vouched for by a human with a citation), no whitelist possible or
//	           needed, output is a SUGGESTION, a wrong one costs one ignorable note
//
// It is also floored at the measured distinctiveness threshold and matched
// against script text only, so it inherits every guard the probe carries.
//
// ── A doc_note, AND WHY THAT IS NOT ROUTING ROUND A BROKEN QUEUE TWICE ─────
//
// The council's architecture seat objected (medium, advisory, corr 67643b47) that
// Phase 1's doc_note surface was already the SECOND bespoke durable surface
// invented to avoid the dead needs_human_review queue, and that each such bypass
// makes the eventual triage fix harder to land. That objection is recorded in
// bugs_open/288 §6 and stands. It applies less here for a reason worth stating:
// a suggestion is not a finding. There is nothing wrong with the tool; nobody is
// owed a fix; the note is an INPUT to whoever next edits that PLAN, and doc_notes
// are already loaded into every builder and improver context by load_doc_context.
// A work item would be claiming work is owed, which would be false.
//
// ── WHAT A SUGGESTION PROVES, WHICH IS CO-OCCURRENCE AND NOT ROLE ──────────
//
// That the site's register carries value V and this tool's code contains V as a
// number in its own right. NOT that the tool uses V for the thing the fact
// describes. A human confirms the binding; the machine only notices it. And this
// leg only finds tools that AGREE with the register — a tool wrong from birth
// (bugs_closed/225's original shape) is invisible to it, and is reachable only
// through a human declaration or Piece 4's oracle. Say so in the note.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// factSuggestToolsQuery lists every tool subject on the site WITH its stored
// component HTML — no doc_plans join, because a tool with no PLAN at all is
// exactly the population most worth suggesting to (TL-032: the recreation
// handler writes no PLAN, so those tools have no criteria and no Tier 2/4).
//
// Same subject-key rule and same audit page predicate as the fan-out, so a
// suggestion is addressed the way a declaration will be resolved.
// ⚠ ONE ROW PER COMPONENT, NOT A string_agg — same reason as pageSurfaceQuery:
// an unbalanced <script> in one partial fragment would leak the tokenizer's
// inScript state into the next component's PROSE, and a suggestion built on that
// would propose a binding for a figure the tool only MENTIONS. Extraction is per
// fragment in buildFactProbeSurface. (Council debug_historian, high, corr 041b3026.)
var factSuggestToolsQuery = `
	SELECT ` + discovery_checks.ToolSubjectKeyExpr + ` AS subject_key,
	       p.name,
	       COALESCE(pc.rendered_html, '') AS fragment
	FROM pages p
	JOIN page_components pc ON pc.page_id = p.id
	JOIN content_components cc ON cc.id = pc.component_id AND cc.is_active = true
	WHERE p.site_id = $1
	  AND (p.page_type = 'tool' OR cc.component_level = 'tool')
	  AND NOT (p.status = 'archived' AND ` + datahelpers.NeverDeployedPagePredicateFor("p") + `)
	ORDER BY 1, 2, pc.position`

type factBindingSuggestion struct {
	SubjectKey string
	PageName   string
	FactIDs    []string // UNAMBIGUOUS bindings only — these go in the paste-ready fragment
	Detail     []string // one line per binding, for the note body
	Ambiguous  []string // one line per value that >1 fact shares; reported, never proposed
}

// planFactBindingSuggestions probes every non-declaring tool on the site against
// the register's probeable values. Read-only; the caller writes.
func planFactBindingSuggestions(ctx context.Context, db *sql.DB, siteID uuid.UUID, eb map[string]interface{}, declaring map[string]bool, logger *zap.Logger) []factBindingSuggestion {
	factsRaw, _ := eb["facts"].([]interface{})
	if len(factsRaw) == 0 {
		return nil
	}
	type probeFact struct {
		id  string
		val float64
	}
	var probeable []probeFact
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		id := datahelpers.GetStringField(fact, "id", "")
		v, has := numericField(fact["value"])
		if id == "" || !has {
			continue
		}
		if _, ok := factValueLiterals(v); !ok {
			continue // below the measured distinctiveness floor — refused, not guessed
		}
		probeable = append(probeable, probeFact{id: id, val: v})
	}
	if len(probeable) == 0 {
		return nil
	}
	// ⚠ TWO FACTS THAT SHARE A VALUE CANNOT BE TOLD APART BY A VALUE PROBE, and
	// proposing both creates a binding a human can never reconcile.
	//
	// FOUND IN PRODUCTION ON THE FIRST REAL SWEEP (2026-08-25, agritec.uk): two
	// facts, `CIT-3f1b219f15ec6a39` and `CIT-86c4010f7cdf820d`, both asserting the
	// SFI26 annual agreement cap of £100,000, were BOTH proposed for the one
	// `100000` in the tool's script. The `not_probed` constant's own comment
	// already claimed "refused: no value, below the floor, or ambiguous" — the
	// ambiguity arm did not exist. A comment promising a guard that was never
	// written, which is the exact class this lane spent two days closing.
	//
	// Reported rather than silently dropped: the collision is usually a REGISTER
	// duplicate and the owner should see it. But it stays out of the paste-ready
	// fragment, because pasting it declares two facts for one constant and every
	// later pass then owes a reconciliation nobody can perform.
	sharedValue := map[string]int{}
	for _, pf := range probeable {
		sharedValue[formatEvidenceNumber(pf.val)]++
	}

	rows, err := db.QueryContext(ctx, factSuggestToolsQuery, siteID)
	if err != nil {
		logger.Warn("refresh_evidence_base: fact-binding suggestion query failed — no suggestions this pass",
			zap.String("site_id", siteID.String()), zap.Error(err))
		return nil
	}
	defer rows.Close()

	// Collect fragments per (subject, page), then extract each SEPARATELY.
	type toolKey struct{ subject, page string }
	order := []toolKey{}
	frags := map[toolKey][]string{}
	for rows.Next() {
		var subjectKey, pageName, fragment string
		if err := rows.Scan(&subjectKey, &pageName, &fragment); err != nil {
			logger.Warn("refresh_evidence_base: fact-binding suggestion scan failed", zap.Error(err))
			return nil
		}
		if subjectKey == "" || declaring[subjectKey] {
			continue // already declares: this is an adoption lever, not a re-audit
		}
		k := toolKey{subjectKey, pageName}
		if _, seen := frags[k]; !seen {
			order = append(order, k)
		}
		frags[k] = append(frags[k], fragment)
	}
	if err := rows.Err(); err != nil {
		logger.Warn("refresh_evidence_base: fact-binding suggestion rows failed", zap.Error(err))
		return nil
	}

	var out []factBindingSuggestion
	for _, k := range order {
		surface := buildFactProbeSurface(frags[k])
		if strings.TrimSpace(surface.ScriptText) == "" {
			continue // no code to look at; silence is the honest answer
		}
		s := factBindingSuggestion{SubjectKey: k.subject, PageName: k.page}
		ambiguous := map[string][]string{} // display value -> fact ids that share it
		for _, pf := range probeable {
			lits, _ := factValueLiterals(pf.val)
			for _, lit := range lits {
				if valueOccursGuarded(surface.ScriptText, lit) {
					disp := formatEvidenceNumber(pf.val)
					if sharedValue[disp] > 1 {
						ambiguous[disp] = append(ambiguous[disp], pf.id)
						break
					}
					s.FactIDs = append(s.FactIDs, pf.id)
					s.Detail = append(s.Detail, fmt.Sprintf("`%s` = %s, present in the script as %q",
						pf.id, disp, lit))
					break
				}
			}
		}
		ambigVals := make([]string, 0, len(ambiguous))
		for k := range ambiguous {
			ambigVals = append(ambigVals, k)
		}
		sort.Strings(ambigVals) // stable order: the note must not churn between passes
		for _, disp := range ambigVals {
			ids := ambiguous[disp]
			sort.Strings(ids)
			s.Ambiguous = append(s.Ambiguous, fmt.Sprintf(
				"%s is carried by %d facts on this site (%s) — this probe cannot tell which one the tool uses, "+
					"so none of them is proposed. Usually this means the register has duplicates; de-duplicate, "+
					"or bind by hand with a contextual artifact_check.",
				disp, len(ids), strings.Join(ids, ", ")))
		}
		if len(s.FactIDs) > 0 || len(s.Ambiguous) > 0 {
			sort.Strings(s.FactIDs)
			out = append(out, s)
		}
	}
	return out
}

// writeFactBindingSuggestions files one doc_note per tool, 30-day cooldown per
// (subject, site) — the same scoping Phase 1's note uses, and for the same
// reason: one fleet-global PLAN resolves on many sites (measured 2026-08-24:
// 6 tool subjects do), so a note about site A's register must not silence the
// different suggestion site B would make.
//
// Writes nothing on a dry run.
func writeFactBindingSuggestions(ctx context.Context, db *sql.DB, siteID uuid.UUID, suggestions []factBindingSuggestion, dryRun bool, logger *zap.Logger) int {
	if dryRun || len(suggestions) == 0 {
		return 0
	}
	written := 0
	for _, s := range suggestions {
		var recent bool
		if err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM doc_notes
				WHERE subject_type='tool' AND subject_key=$1 AND site_id=$2
				  AND categories ? 'fact_binding_suggested'
				  AND created_at > NOW() - INTERVAL '30 days')
		`, s.SubjectKey, siteID).Scan(&recent); err != nil {
			logger.Warn("refresh_evidence_base: binding-suggestion cooldown check failed",
				zap.String("tool", s.SubjectKey), zap.Error(err))
			continue
		}
		if recent {
			continue
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories,
			                       source, source_agent, created_by)
			VALUES ('tool', $1, $2, $3, '["fact_binding_suggested"]'::jsonb,
			        'refresh_evidence_base', 'evidence-freshness', 'evidence-freshness')
		`, s.SubjectKey, siteID, factBindingSuggestionBody(s)); err != nil {
			logger.Warn("refresh_evidence_base: binding-suggestion note insert failed",
				zap.String("tool", s.SubjectKey), zap.Error(err))
			continue
		}
		written++
	}
	if written > 0 {
		logger.Info("refresh_evidence_base: filed fact_binding_suggested notes",
			zap.String("site_id", siteID.String()), zap.Int("tools", written))
	}
	return written
}

func factBindingSuggestionBody(s factBindingSuggestion) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## This tool appears to encode registered facts, and declares none — %s\n", s.SubjectKey)
	fmt.Fprintf(&b, "Observed: the daily evidence sweep found %d unambiguous registered value(s) in the script of `%s`:\n",
		len(s.FactIDs), s.PageName)
	for _, d := range s.Detail {
		fmt.Fprintf(&b, "  - %s\n", d)
	}
	b.WriteString("Root cause: not-applicable — nothing is wrong with this tool. It is not declared, so when one of " +
		"these figures changes in the register nothing tells this calculator, which is bugs_closed/225's class.\n")
	for _, a := range s.Ambiguous {
		fmt.Fprintf(&b, "  ⚠ AMBIGUOUS, NOT PROPOSED: %s\n", a)
	}
	if len(s.FactIDs) == 0 {
		b.WriteString("Fix: nothing is proposed — every match on this tool was ambiguous (see above). " +
			"Resolve the duplicate facts in the register, or bind by hand with a contextual artifact_check.\n")
	} else {
		b.WriteString("Fix: if these bindings are right, add them to the tool's PLAN criteria fence — paste-ready:\n\n")
		b.WriteString("      \"facts\": [\n")
		for i, id := range s.FactIDs {
			comma := ","
			if i == len(s.FactIDs)-1 {
				comma = ""
			}
			fmt.Fprintf(&b, "        %q%s\n", id, comma)
		}
		b.WriteString("      ]\n\n")
	}
	b.WriteString("  Install it through the lane's own fence installer; never hand-edit the doc_plans row. " +
		"If this tool has no PLAN at all, it needs one first (no PLAN means no criteria and no Tier 2/4).\n")
	b.WriteString("Verified: n/a — this is a SUGGESTION, and it proves co-occurrence, not role. The value is in the " +
		"register and the same number is in the code; that the tool uses it FOR the thing the fact describes is for " +
		"a human to confirm. ⚠ And this can only find tools that AGREE with the register: a calculator that has been " +
		"wrong since it was built carries no registered value to match, so it stays invisible here.\n")
	b.WriteString("Categories: fact_binding_suggested")
	return b.String()
}
