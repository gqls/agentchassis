// FILE: platform/orchestration/actions/refresh_evidence_fact_drift.go
//
// Piece 3 of PLAN_2026-08-09_facts_into_tool_acceptance.md (bugs_open/225's
// class fix): the daily evidence sweep fans a moved fact out to the TOOLS that
// declare they encode it, instead of stopping at one per-site stale_evidence
// item nobody handles.
//
// The one-line problem it closes: the evidence register guards COPY, not CODE.
// A fact constrains what a page may SAY; nothing told the calculators that
// encode a fact when the fact moved. mortgagecalculator's SDLT tool ran a rule
// that expired 2025-03-31 for sixteen months while every check passed
// (bugs_open/225). The register on that site now carries the current figures,
// re-verified daily; this file is what makes a change to one of them reach the
// tool the same day.
//
// How the join works (and what it deliberately is NOT):
//
//   - The declaration lives in the tool's criteria fence, fence-level `facts`
//     (criteria_facts.go). doc_plans has NO site_id, so ids are resolved against
//     the register of the site being swept — the driven page's site, never the
//     PLAN (PBP-037's rule; plan §5.1). An id the site's register does not carry
//     is inert and reported, never fatal, never an item.
//   - PLAN ↔ page resolution is the platform's OWN name rule, the one Tier 4
//     already uses to find a tool's URL (tool_acceptance_actions.go:
//     `name IN ($2, 'tool-' || $2)`), plus a tool-level component whose
//     function is the subject key. It is deliberately NOT the acceptance ladder's
//     eligibility predicate (discovery_checks/tool_eligibility.go): measured
//     2026-08-15, that predicate's sole-component clause admits NEITHER of the
//     two tools this mechanism was built for (mortgagecalculator `tool-stamp-duty`
//     carries two components, loanandmortgagecalculator `mortgages-stamp-duty`
//     three after its B2 decomposition) — a fan-out keyed on it would be a check
//     that can never fire on the tools that need it. A tool need not be
//     acceptance-eligible to encode a fact.
//
// The split that must not collapse (plan §2 Piece 3, and TL-040/TL-042):
//
//   - VALUE DRIFT — the fact's number is not the number the tool was last told
//     about — routes to `improve_tool` (tool-improver) ONLY when the tool is a
//     real fork (a tool-level component owns the code) AND its fence does not
//     say `no_auto_fix`. Otherwise it routes to a human as `fact_drift_review`:
//     tool-improver rewrites the SHARED template of a non-fork (it did, fleet-
//     wide, on 2026-08-05 and 2026-08-14 — bugs_open/281), and a fence author
//     who wrote no_auto_fix has said a human decides what may change on a page
//     quoting tax law.
//   - EVIDENCE DRIFT — the fact's citation was lost, its artefact changed, or a
//     sql fact left tolerance, with no value move — is ALWAYS a human's:
//     "GOV.UK returned 404" is not evidence the number moved, and pointing a
//     rewriter at a calculator on that evidence is bugs_open/126's failure with
//     arithmetic as the target.
//   - FETCH ERROR — the source could not be checked at all — fans out to
//     NOBODY. CLM-008: unknown is not loss; a per-tool item on every 403 day
//     trains people to ignore the queue. Reported in the result, not filed.
//
// Baseline for "value drift": the value the tool was last TOLD. First the most
// recent fact_drift item for (fact, tool) — its spec.fact.new_value; else the
// same fact's value in the previous evidence_base row for the site (a human
// supersede or an in-place UPDATE both leave the older row behind); else no
// baseline and no fire — a register with one row ever, or a fact newly
// declared, is silent. Piece 4 (an RFC, not this file) owns "is the current
// number RIGHT"; this file only owns "did it MOVE, and who encodes it".
//
// What this file never does: touch `changed`, res.Drifted,
// res.ArtifactCheckDrifted or shouldRaiseStaleEvidence — the RFC_025 council
// history (refresh_evidence_base_action.go, the comment above that gate) is
// that the pre-existing citation raise must not change for existing callers.
// A citation-only drift with nothing else changed still raises no site item;
// the per-tool item is what fires now, on its own trigger and its own keys.
//
// Work items (owner ruling 2026-08-02 §1 — producer set and key shape are named
// here and in the concept register, CLM-022):
//
//	improve_tool        key fact_drift:<fact_id>:<subject_key>:<site_id>
//	                    producer refresh_evidence_base (third producer after
//	                    check_tool_health / check_tool_acceptance), handler
//	                    tool-improver, status detected, priority 30
//	fact_drift_review   key fact_drift_review:<fact_id>:<subject_key>:<site_id>
//	                    NEW type, sole producer refresh_evidence_base, handler
//	                    human-review, status needs_human_review, priority 35
//
// Two key prefixes rather than one because refreshOnConflict rewrites an open
// item's summary/spec but never its type or handler — an open review item must
// not absorb a later improve_tool finding.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// factDriftTool is one (page, PLAN) pair on the swept site whose fence declares
// facts. ForkComponentID is set iff a tool-level component whose function is the
// subject key lives on the page — the only case tool-improver may be pointed at.
type factDriftTool struct {
	SubjectKey      string
	PageID          string
	PageName        string
	PageURL         string
	BuildStatus     string
	ForkComponentID string
	NoAutoFix       bool
	NoAutoFixReason string
	DeclaredFacts   []string
	Criteria        string
}

func (t factDriftTool) isFork() bool { return t.ForkComponentID != "" }

// factDriftIndex is the site's declarations, keyed by fact id, plus the ids the
// register could not resolve ("<subject_key>:<fact_id>").
type factDriftIndex struct {
	byFact     map[string][]factDriftTool
	unresolved []string
	issues     []string // malformed declarations, "<subject_key>: <issue>"
}

// factDriftIndexQuery resolves every current tool PLAN whose fence mentions
// `facts` to the active pages on THIS site that carry the tool. See the file
// header for why this is the name rule and not the ladder predicate. The LIKE
// is a cheap pre-filter; the fence is parsed in Go.
const factDriftIndexQuery = `
		SELECT p.id::text, p.name, COALESCE(p.url, ''), COALESCE(p.build_status, ''),
		       dp.subject_key, dp.body,
		       COALESCE((
		           SELECT cc.id::text FROM page_components pc
		           JOIN content_components cc ON cc.id = pc.component_id
		           WHERE pc.page_id = p.id AND cc.is_active = true AND cc.component_level = 'tool'
		             AND cc.function IN (dp.subject_key, 'tool-' || dp.subject_key)
		           ORDER BY pc.position LIMIT 1
		       ), '') AS fork_component_id
		FROM pages p
		JOIN doc_plans dp
		  ON dp.subject_type = 'tool' AND dp.is_current = true
		 AND (p.name = dp.subject_key OR p.name = 'tool-' || dp.subject_key
		      OR EXISTS (
		          SELECT 1 FROM page_components pc2
		          JOIN content_components cc2 ON cc2.id = pc2.component_id
		          WHERE pc2.page_id = p.id AND cc2.is_active = true AND cc2.component_level = 'tool'
		            AND cc2.function = dp.subject_key))
		WHERE p.site_id = $1 AND p.status = 'active' AND dp.body LIKE '%"facts"%'
		ORDER BY dp.subject_key, p.name`

// loadFactDriftIndex runs the one join query for a site and parses each fence.
// registerFactIDs is the set of ids the site's register carries; declared ids
// outside it are recorded as unresolved. Returns nil (no error) when no PLAN on
// the site declares anything — the no-op path costs exactly this one SELECT.
func loadFactDriftIndex(ctx context.Context, db *sql.DB, siteID uuid.UUID, registerFactIDs map[string]bool) (*factDriftIndex, error) {
	rows, err := db.QueryContext(ctx, factDriftIndexQuery, siteID)
	if err != nil {
		return nil, fmt.Errorf("fact drift index: %w", err)
	}
	defer rows.Close()

	idx := &factDriftIndex{byFact: map[string][]factDriftTool{}}
	any := false
	for rows.Next() {
		var t factDriftTool
		var body string
		if err := rows.Scan(&t.PageID, &t.PageName, &t.PageURL, &t.BuildStatus, &t.SubjectKey, &body, &t.ForkComponentID); err != nil {
			return nil, fmt.Errorf("fact drift index scan: %w", err)
		}
		t.Criteria = extractCriteriaBlock(body)
		ids, issues := parseCriteriaFacts(t.Criteria)
		for _, is := range issues {
			idx.issues = append(idx.issues, t.SubjectKey+": "+is)
		}
		if len(ids) == 0 {
			continue
		}
		t.NoAutoFix, t.NoAutoFixReason = parseNoAutoFix(t.Criteria)
		for _, id := range ids {
			if !registerFactIDs[id] {
				idx.unresolved = append(idx.unresolved, t.SubjectKey+":"+id)
				continue
			}
			t.DeclaredFacts = append(t.DeclaredFacts, id)
		}
		if len(t.DeclaredFacts) == 0 {
			continue
		}
		any = true
		for _, id := range t.DeclaredFacts {
			idx.byFact[id] = append(idx.byFact[id], t)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fact drift index rows: %w", err)
	}
	if !any && len(idx.unresolved) == 0 && len(idx.issues) == 0 {
		return nil, nil
	}
	return idx, nil
}

// factDriftBaselines holds "the value the tool was last told" per (fact, tool)
// and the previous register row's value per fact. See the file header.
type factDriftBaselines struct {
	lastItem    map[string]float64 // key: factID + "|" + subjectKey
	previousRow map[string]float64 // key: factID
}

func baselineKey(factID, subjectKey string) string { return factID + "|" + subjectKey }

// baselineFor answers with the most specific baseline available, or nil.
func (b factDriftBaselines) baselineFor(factID, subjectKey string) *float64 {
	if v, ok := b.lastItem[baselineKey(factID, subjectKey)]; ok {
		return &v
	}
	if v, ok := b.previousRow[factID]; ok {
		return &v
	}
	return nil
}

const factDriftLastItemQuery = `
		SELECT spec->'fact'->>'id', COALESCE(spec->>'subject_key', ''), spec->'fact'->>'new_value'
		FROM site_work_items
		WHERE site_id = $1 AND spec->>'check' = 'fact_drift'
		ORDER BY created_at DESC`

const factDriftPreviousRowQuery = `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND id <> $2
		ORDER BY created_at DESC LIMIT 1`

// loadFactDriftBaselines reads both baselines. Only called when the site has at
// least one resolvable declaration, so sites without declarations pay nothing.
// The previous row is read as a raw map — never through the typed EvidenceBase
// (plan landmine §5.2: the typed struct does not model citation/writer_line/unit).
func loadFactDriftBaselines(ctx context.Context, db *sql.DB, siteID, currentRowID uuid.UUID) (factDriftBaselines, error) {
	b := factDriftBaselines{lastItem: map[string]float64{}, previousRow: map[string]float64{}}

	rows, err := db.QueryContext(ctx, factDriftLastItemQuery, siteID)
	if err != nil {
		return b, fmt.Errorf("fact drift baselines (items): %w", err)
	}
	for rows.Next() {
		var factID, subjectKey, newVal sql.NullString
		if err := rows.Scan(&factID, &subjectKey, &newVal); err != nil {
			rows.Close()
			return b, fmt.Errorf("fact drift baselines scan: %w", err)
		}
		if !factID.Valid || !newVal.Valid {
			continue
		}
		k := baselineKey(factID.String, subjectKey.String)
		if _, seen := b.lastItem[k]; seen {
			continue // ORDER BY created_at DESC: first row per key is the latest
		}
		var f float64
		if _, err := fmt.Sscanf(strings.TrimSpace(newVal.String), "%g", &f); err == nil {
			b.lastItem[k] = f
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return b, fmt.Errorf("fact drift baselines rows: %w", err)
	}

	var raw []byte
	err = db.QueryRowContext(ctx, factDriftPreviousRowQuery, siteID, currentRowID).Scan(&raw)
	switch {
	case err == sql.ErrNoRows:
		return b, nil
	case err != nil:
		return b, fmt.Errorf("fact drift baselines (previous row): %w", err)
	}
	var prev map[string]interface{}
	if err := json.Unmarshal(raw, &prev); err != nil {
		return b, nil // an unreadable previous row is "no baseline", not a failure
	}
	facts, _ := prev["facts"].([]interface{})
	for _, fr := range facts {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		id := datahelpers.GetStringField(fact, "id", "")
		if id == "" {
			continue
		}
		if v, ok := numericField(fact["value"]); ok {
			b.previousRow[id] = v
		}
	}
	return b, nil
}

// factDriftEmission is one planned or written per-(fact, tool) finding.
type factDriftEmission struct {
	FactID      string   `json:"fact_id"`
	SubjectKey  string   `json:"subject_key"`
	PageID      string   `json:"page_id,omitempty"`
	PageName    string   `json:"page_name,omitempty"`
	ComponentID string   `json:"component_id,omitempty"`
	Kind        string   `json:"kind"`             // value_drift | evidence_drift | skipped_unknown
	Route       string   `json:"route"`            // improve_tool | fact_drift_review | none
	Reason      string   `json:"reason,omitempty"` // not_a_fork | no_auto_fix | evidence_drift | fetch_error
	OldValue    *float64 `json:"old_value,omitempty"`
	NewValue    *float64 `json:"new_value,omitempty"`
	Detail      string   `json:"detail,omitempty"`
	ItemKey     string   `json:"item_key,omitempty"`
	Outcome     string   `json:"outcome"` // planned | dry_run | inserted | refreshed | dropped | error | none
}

// classifyFactDrift decides, for ONE (fact, tool) pair, whether anything is
// owed to the tool and where it goes. Pure: no DB, no side effects. entry may
// be nil (a fact this pass produced no refresh entry for). Returns ok=false when
// nothing is owed.
func classifyFactDrift(entry *evidenceFactRefresh, fact map[string]interface{}, tool factDriftTool, baseline *float64, siteID string) (factDriftEmission, bool) {
	em := factDriftEmission{
		FactID:     datahelpers.GetStringField(fact, "id", ""),
		SubjectKey: tool.SubjectKey,
		PageID:     tool.PageID,
		PageName:   tool.PageName,
		Route:      "none",
		Outcome:    "planned",
	}
	if tool.isFork() {
		em.ComponentID = tool.ForkComponentID
	}
	newVal, hasVal := numericField(fact["value"])

	switch {
	case hasVal && baseline != nil && math.Abs(newVal-*baseline) > 1e-9:
		old := *baseline
		nv := newVal
		em.Kind = "value_drift"
		em.OldValue = &old
		em.NewValue = &nv
		switch {
		case !tool.isFork():
			em.Route, em.Reason = "fact_drift_review", "not_a_fork"
		case tool.NoAutoFix:
			em.Route, em.Reason = "fact_drift_review", "no_auto_fix"
		default:
			em.Route = "improve_tool"
		}
		em.Detail = fmt.Sprintf("registered value moved %s → %s; %s declares it encodes this fact",
			formatEvidenceNumber(old), formatEvidenceNumber(nv), tool.SubjectKey)
	case entry != nil && entry.Outcome == "drifted":
		em.Kind = "evidence_drift"
		em.Route, em.Reason = "fact_drift_review", "evidence_drift"
		em.Detail = entry.Detail
		if hasVal {
			nv := newVal
			em.NewValue = &nv
		}
	case entry != nil && entry.Outcome == "error":
		em.Kind = "skipped_unknown"
		em.Reason = "fetch_error"
		em.Outcome = "none"
		em.Detail = "source could not be checked this pass — unknown is not loss (CLM-008); no item filed"
		return em, true
	default:
		return em, false
	}

	switch em.Route {
	case "improve_tool":
		em.ItemKey = fmt.Sprintf("fact_drift:%s:%s:%s", em.FactID, tool.SubjectKey, siteID)
	case "fact_drift_review":
		em.ItemKey = fmt.Sprintf("fact_drift_review:%s:%s:%s", em.FactID, tool.SubjectKey, siteID)
	}
	return em, true
}

// planFactDriftFanOut walks every declared fact on the site and returns the
// emissions owed, in stable order. Pure. It also stamps EncodedByTools onto
// the drifted entries in res.Facts so the site's stale_evidence item (if the
// existing gate raises one) names the tools — that field is omitempty, so a
// site with no declarations marshals byte-identically to before.
func planFactDriftFanOut(res *siteRefreshResult, factsByID map[string]map[string]interface{}, idx *factDriftIndex, base factDriftBaselines, siteID string) []factDriftEmission {
	if idx == nil {
		return nil
	}
	entryIdx := map[string]int{}
	for i, e := range res.Facts {
		entryIdx[e.FactID] = i // last entry per id wins
	}
	factIDs := make([]string, 0, len(idx.byFact))
	for id := range idx.byFact {
		factIDs = append(factIDs, id)
	}
	sort.Strings(factIDs)

	var out []factDriftEmission
	for _, id := range factIDs {
		fact := factsByID[id]
		if fact == nil {
			continue
		}
		var entry *evidenceFactRefresh
		if i, ok := entryIdx[id]; ok {
			entry = &res.Facts[i]
		}
		tools := idx.byFact[id]
		if entry != nil && entry.Outcome == "drifted" {
			for _, t := range tools {
				entry.EncodedByTools = append(entry.EncodedByTools, t.SubjectKey)
			}
		}
		for _, t := range tools {
			em, ok := classifyFactDrift(entry, fact, t, base.baselineFor(id, t.SubjectKey), siteID)
			if !ok {
				continue
			}
			out = append(out, em)
		}
	}
	return out
}

// factDriftPlan is what planSiteFactDrift hands back: the emissions (also
// surfaced on the result as res.FactDrift) plus the lookups the writer needs.
type factDriftPlan struct {
	Emissions  []factDriftEmission
	factsByID  map[string]map[string]interface{}
	toolsByKey map[string]factDriftTool // key: subjectKey + "|" + pageID
}

// planSiteFactDrift is the glue refreshOneSiteEvidence calls once per site,
// BEFORE the dry-run return so a dry run reports what it would file. Read-only.
// Any DB failure here is logged and yields no emissions — the sweep's existing
// work must not be aborted by the fan-out.
func planSiteFactDrift(ctx context.Context, db *sql.DB, siteID, specRowID uuid.UUID, eb map[string]interface{}, res *siteRefreshResult, dryRun bool, logger *zap.Logger) factDriftPlan {
	plan := factDriftPlan{toolsByKey: map[string]factDriftTool{}}
	factsRaw, _ := eb["facts"].([]interface{})
	factsByID := make(map[string]map[string]interface{}, len(factsRaw))
	registerIDs := make(map[string]bool, len(factsRaw))
	for _, fr := range factsRaw {
		if fact, ok := fr.(map[string]interface{}); ok {
			if id := datahelpers.GetStringField(fact, "id", ""); id != "" {
				factsByID[id] = fact
				registerIDs[id] = true
			}
		}
	}

	plan.factsByID = factsByID

	idx, err := loadFactDriftIndex(ctx, db, siteID, registerIDs)
	if err != nil {
		logger.Warn("refresh_evidence_base: fact-drift index failed — no fan-out this pass",
			zap.String("site_id", siteID.String()), zap.Error(err))
		return plan
	}
	if idx == nil {
		return plan
	}
	for _, tools := range idx.byFact {
		for _, t := range tools {
			plan.toolsByKey[t.SubjectKey+"|"+t.PageID] = t
		}
	}
	res.FactDeclarationsUnresolved = idx.unresolved
	if len(idx.unresolved) > 0 {
		logger.Info("refresh_evidence_base: criteria fences declare fact ids this site's register does not carry — inert (PBP-037)",
			zap.String("site_id", siteID.String()), zap.Strings("unresolved", idx.unresolved))
	}
	for _, is := range idx.issues {
		logger.Warn("refresh_evidence_base: malformed facts declaration ignored", zap.String("site_id", siteID.String()), zap.String("issue", is))
	}
	if len(idx.byFact) == 0 {
		return plan
	}

	base, err := loadFactDriftBaselines(ctx, db, siteID, specRowID)
	if err != nil {
		logger.Warn("refresh_evidence_base: fact-drift baselines failed — no fan-out this pass",
			zap.String("site_id", siteID.String()), zap.Error(err))
		return plan
	}
	ems := planFactDriftFanOut(res, factsByID, idx, base, siteID.String())
	if dryRun {
		for i := range ems {
			if ems[i].Route != "none" {
				ems[i].Outcome = "dry_run"
			}
		}
	}
	plan.Emissions = ems
	return plan
}

// factDriftSourceURL pulls the citation URL off a fact for the item body.
func factDriftSourceURL(fact map[string]interface{}) string {
	src, _ := fact["source"].(map[string]interface{})
	if src == nil {
		return ""
	}
	if cit, ok := src["citation"].(map[string]interface{}); ok {
		return datahelpers.GetStringField(cit, "url", "")
	}
	return ""
}

// factDriftIssueText is the human-readable payload. tool-improver's prompt
// reads {{.input_data.issue}}, so everything the fixer needs is IN the text.
func factDriftIssueText(fact map[string]interface{}, em factDriftEmission, tool factDriftTool) string {
	var b strings.Builder
	unit := datahelpers.GetStringField(fact, "unit", "")
	claim := datahelpers.GetStringField(fact, "claim", "")
	switch em.Kind {
	case "value_drift":
		fmt.Fprintf(&b, "Registered fact %q (%s) moved from %s to %s%s.", em.FactID, claim,
			formatEvidenceNumber(*em.OldValue), formatEvidenceNumber(*em.NewValue), unitSuffix(unit))
	default:
		fmt.Fprintf(&b, "Registered fact %q (%s): %s.", em.FactID, claim, em.Detail)
	}
	fmt.Fprintf(&b, " Tool %q (page %s) declares it encodes this fact (fence-level facts, PLAN %s).",
		tool.SubjectKey, tool.PageName, tool.SubjectKey)
	if em.Kind == "value_drift" {
		b.WriteString(" Change ONLY what encodes this fact; the register's wording is authoritative.")
	} else {
		b.WriteString(" Do NOT change the tool's numbers on this evidence — a lost citation is not a moved figure; a human rules first.")
	}
	if wl := datahelpers.GetStringField(fact, "writer_line", ""); wl != "" && em.NewValue != nil {
		fmt.Fprintf(&b, " Register wording: %q.", strings.ReplaceAll(wl, "{value}", formatEvidenceNumber(*em.NewValue)))
	}
	if u := factDriftSourceURL(fact); u != "" {
		fmt.Fprintf(&b, " Source: %s (verified %s).", u, datahelpers.GetStringField(fact, "verified_at", "?"))
	}
	if em.Reason == "no_auto_fix" {
		fmt.Fprintf(&b, " Routed to a human because the fence says no_auto_fix (%s).", reasonOrUnstated(tool.NoAutoFixReason))
	}
	if em.Reason == "not_a_fork" {
		b.WriteString(" Routed to a human because no tool-level component owns this page's code (a ported or decomposed tool) — " +
			"re-file needs_tool_recreation, whose builder reads the register (CLM-021); never improve_tool, whose component_id would be the shared wrapper.")
	}
	return b.String()
}

func unitSuffix(unit string) string {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "":
		return ""
	case "gbp":
		return " GBP"
	case "percent", "%":
		return " percent"
	default:
		return " " + unit
	}
}

// writeFactDriftItems files one work item per emission with a route, each in
// its own transaction, and records the honest outcome on the emission.
func writeFactDriftItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string, plan *factDriftPlan, agentType string, logger *zap.Logger) {
	ems := plan.Emissions
	factsByID, toolsByKey := plan.factsByID, plan.toolsByKey
	for i := range ems {
		em := &ems[i]
		if em.Route != "improve_tool" && em.Route != "fact_drift_review" {
			continue
		}
		fact := factsByID[em.FactID]
		tool := toolsByKey[em.SubjectKey+"|"+em.PageID]
		if fact == nil {
			em.Outcome = "error"
			em.Detail += " (fact vanished before write)"
			continue
		}
		spec := map[string]interface{}{
			"check":       "fact_drift",
			"kind":        em.Kind,
			"reason":      em.Reason,
			"domain":      domain,
			"subject_key": em.SubjectKey,
			"page_id":     em.PageID,
			"page_name":   em.PageName,
			"page_url":    tool.PageURL,
			"fact": map[string]interface{}{
				"id":          em.FactID,
				"claim":       datahelpers.GetStringField(fact, "claim", ""),
				"old_value":   em.OldValue,
				"new_value":   em.NewValue,
				"unit":        datahelpers.GetStringField(fact, "unit", ""),
				"writer_line": datahelpers.GetStringField(fact, "writer_line", ""),
				"source_url":  factDriftSourceURL(fact),
				"verified_at": datahelpers.GetStringField(fact, "verified_at", ""),
			},
			"issue": factDriftIssueText(fact, *em, tool),
		}
		if em.ComponentID != "" {
			spec["component_id"] = em.ComponentID
		}
		if em.Route == "improve_tool" && strings.TrimSpace(tool.Criteria) != "" {
			spec["acceptance_test"] = json.RawMessage(tool.Criteria)
		}
		if em.Route == "fact_drift_review" {
			spec["route_hint"] = "human decides; for a non-fork re-file needs_tool_recreation (builder reads the register, CLM-021); for no_auto_fix re-verify the tool against the fact's citation, then rebuild"
		}
		specJSON, err := json.Marshal(spec)
		if err != nil {
			em.Outcome = "error"
			continue
		}

		var pageUUID, compUUID *uuid.UUID
		if pu, perr := uuid.Parse(em.PageID); perr == nil {
			pageUUID = &pu
		}
		if cu, cerr := uuid.Parse(em.ComponentID); cerr == nil {
			compUUID = &cu
		}

		item := workItem{
			siteID:      siteID,
			source:      "scheduled",
			pipeline:    "build",
			spec:        string(specJSON),
			pageID:      pageUUID,
			componentID: compUUID,
			createdBy:   agentType,
			itemKey:     em.ItemKey,
		}
		switch em.Route {
		case "improve_tool":
			item.itemType = "improve_tool"
			item.severity = "high"
			item.priority = 30
			item.handlerAgent = "tool-improver"
			item.status = "detected"
			item.summary = fmt.Sprintf("Registered fact %s moved (%s → %s): tool %s encodes it",
				em.FactID, formatEvidenceNumber(*em.OldValue), formatEvidenceNumber(*em.NewValue), em.SubjectKey)
		default:
			item.itemType = "fact_drift_review"
			item.severity = "medium"
			item.priority = 35
			item.handlerAgent = "human-review"
			item.status = "needs_human_review"
			item.summary = fmt.Sprintf("Fact drift for tool %s: %s %s (%s)", em.SubjectKey, em.FactID, em.Kind, em.Reason)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			em.Outcome = "error"
			logger.Warn("refresh_evidence_base: fact-drift item begin tx failed", zap.Error(err))
			continue
		}
		write, err := writeWorkItem(ctx, tx, item, refreshOnConflict, logger)
		if err != nil {
			tx.Rollback()
			em.Outcome = "error"
			logger.Warn("refresh_evidence_base: fact-drift item write failed", zap.String("item_key", em.ItemKey), zap.Error(err))
			continue
		}
		if err := tx.Commit(); err != nil {
			em.Outcome = "error"
			continue
		}
		switch {
		case write.Inserted:
			em.Outcome = "inserted"
		case write.Refreshed:
			em.Outcome = "refreshed"
		default:
			em.Outcome = "dropped"
		}
		logger.Warn("refresh_evidence_base: fact drift fanned out to a tool",
			zap.String("site_id", siteID.String()), zap.String("item_type", item.itemType),
			zap.String("item_key", em.ItemKey), zap.String("outcome", em.Outcome))
	}
}
