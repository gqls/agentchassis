// FILE: platform/orchestration/actions/discovery_checks/check_content_duplication.go
//
// The permanent form of the census bugs_open/151 ran by hand: "is this site
// saying the same things all over itself?" Owner-requested 2026-07-30 after that
// census, run against vonc.com, found the site's own about page rendering every
// section twice in public.
//
// THE SPLIT, AND WHY IT IS THE WHOLE DESIGN (owner ruling, 2026-07-31)
// -------------------------------------------------------------------
// The census finds two defects that are indistinguishable in its output and
// opposites underneath:
//
//	IN-REMIT  same page, IDENTICAL text     -> deterministic. Keep the earliest
//	                                           row, drop the rest. No judgement.
//	RESIDUE   near-duplicate / cross-page   -> requires rewriting prose, which
//	                                           requires deciding what a page is
//	                                           FOR. That is judgement.
//
// Treating them as one problem fails whichever way you pick it. A handler
// cautious enough for the residue cannot fix the in-remit case; one confident
// enough for the in-remit case has been pointed at possibly-good copy on the
// strength of a similarity score, and a similarity score says nothing about
// meaning. That is not hypothetical: on 2026-07-29 an LLM rewrite on
// gamesdesign.co.uk turned "built FOR designers" into "built BY a designer" and
// generated supporting detail to justify the invention (bugs_open/149 C1).
//
// So the residue is NOT dispatched. It becomes one capability_gap item — this
// platform's existing durable record of "found work I have no handler for"
// (remit.go) — which queues it for a decision instead of guessing at it. The
// structural fix for the residue is bugs_open/151 candidate 1 (assign facts to
// sections at plan time), owned by the brochure_component_library lane.
//
// THE DISCRIMINATOR IS CONTENT IDENTITY, NOT SLOT REPETITION
// ---------------------------------------------------------
// Measured fleet-wide for bugs_open/156: 17 duplicate (page_id, slot_name)
// groups exist, and 11 of them are LEGITIMATE — repeated slots carrying
// different content (generic-text-block x2-3 on five sites). A unique index on
// (page_id, slot_name) would break 11 real pages. Only content-identical groups
// are defects, so this check compares normalised text and never slot names.
//
// WHAT THE SIMILARITY NUMBER IS AND IS NOT
// ----------------------------------------
// Token-set Jaccard, used ONLY to size the residue, never to dispatch. It is a
// screen, not a verdict. 151's central measurement is why: two sections there
// asserted the identical six facts while being 18% textually similar, so any
// similarity threshold alone reports them clean. The fact-overlap half below is
// the stronger signal — but it has a floor, measured on vonc: with only 4
// approved facts, no section can reach a 3-fact overlap and the fact half is
// blind. A site under that floor gets the textual screen and an explicit note
// that the fact half could not discriminate. Neither half is trusted alone.
package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() {
	Register(&ContentDuplicationCheck{})
	RegisterVerifier("content_duplication", VerifyContentDuplicationResolved)
}

type ContentDuplicationCheck struct{}

func (c *ContentDuplicationCheck) Name() string { return "content_duplication" }

// jaccardResidueThreshold is deliberately loose. It sizes a population for a
// human to look at; it never causes an edit. Tightening it makes the residue
// report smaller, not the site cleaner.
const jaccardResidueThreshold = 0.55

// minFactsForFactCensus is the floor below which the fact-overlap half cannot
// discriminate. With 4 facts (vonc) a "shares 3+ facts" rule is nearly
// unsatisfiable; 151's own site had 9. Below this we say so rather than
// reporting a clean fact census, because a blind check and a clean site look
// identical in the output.
const minFactsForFactCensus = 6

type siteSection struct {
	ComponentID uuid.UUID
	PageID      uuid.UUID
	PageName    string
	SlotName    string
	Position    int
	Text        string
	Raw         string // content_data::text as read from jsonb — canonical, see datahelpers.SectionIdentityKey
	Tokens      map[string]struct{}
	FactIDs     []string
}

func (c *ContentDuplicationCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	sections, err := loadSiteSections(dctx)
	if err != nil {
		return nil, fmt.Errorf("load sections: %w", err)
	}
	if len(sections) == 0 {
		return &CheckResult{}, nil
	}

	facts, err := loadEvidenceFactIDs(dctx)
	if err != nil {
		// A missing or unreadable evidence base is not a reason to skip the
		// textual half. Record it and carry on with a blind fact census.
		dctx.Logger.Warn("content_duplication: evidence base unreadable, fact census skipped",
			zap.Error(err))
		facts = nil
	}
	assignFacts(sections, facts)

	inRemit := findIdenticalSamePage(sections)
	residue := findResidue(sections, facts)

	result := &CheckResult{}

	// ---- IN-REMIT: one dispatchable item per affected page ----
	for _, grp := range inRemit {
		redundant := grp.redundantComponentIDs()
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":                 "content_duplication",
			"shape":                 "identical_same_page",
			"page_id":               grp.PageID.String(),
			"page_name":             grp.PageName,
			"keep_component_id":     grp.Keep.ComponentID.String(),
			"remove_component_ids":  redundant,
			"duplicate_group_count": len(grp.Groups),
			"remedy":                "delete the later rows, renumber positions, re-render the page",
		})
		pageID := grp.PageID
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "content_duplication",
			"shape":     "identical_same_page",
			"page_name": grp.PageName,
			"groups":    len(grp.Groups),
			"redundant": len(redundant),
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   &pageID,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "content_duplication",
			Severity: "high", // it is visible to every visitor on a live page
			Summary: fmt.Sprintf(
				"Page %s renders %d section(s) twice or more — %d redundant component row(s) with identical content",
				grp.PageName, len(grp.Groups), len(redundant)),
			SpecJSON:     string(specJSON),
			Priority:     20,
			HandlerAgent: "deduplicate-sections",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "content_duplication:" + grp.PageName,
			BatchID:      dctx.BatchID,
		})
	}

	// ---- RESIDUE: exactly one capability_gap for the whole site ----
	if residue.total() > 0 {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":                "content_duplication",
			"shape":                "residue_needs_rewrite",
			"cross_page_identical": residue.CrossPageIdentical,
			"near_duplicate_pairs": residue.NearDuplicatePairs,
			"fact_overlap_pairs":   residue.FactOverlapPairs,
			"fact_census_blind":    residue.FactCensusBlind,
			"fact_pool_size":       len(facts),
			"examples":             residue.Examples,
			"jaccard_threshold":    jaccardResidueThreshold,
			"why_no_handler":       "rewriting sibling copy requires deciding what each page is for; a similarity score is not an instruction about meaning. Structural fix: bugs_open/151 candidate 1 (assign facts to sections at plan time), owned by brochure_component_library.",
			"do_not_auto_rewrite":  true,
		})
		blind := ""
		if residue.FactCensusBlind {
			blind = fmt.Sprintf(" (fact census BLIND — only %d approved facts, below the %d needed to discriminate)",
				len(facts), minFactsForFactCensus)
		}
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":             "content_duplication",
			"shape":             "residue_needs_rewrite",
			"near_duplicates":   residue.NearDuplicatePairs,
			"cross_page_exact":  residue.CrossPageIdentical,
			"fact_overlaps":     residue.FactOverlapPairs,
			"fact_census_blind": residue.FactCensusBlind,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "capability_gap",
			Severity: "medium",
			Summary: fmt.Sprintf(
				"Sibling sections restate each other and no handler can safely rewrite them: %d near-duplicate pair(s), %d cross-page identical, %d sharing 3+ approved facts%s",
				residue.NearDuplicatePairs, residue.CrossPageIdentical, residue.FactOverlapPairs, blind),
			SpecJSON:     string(specJSON),
			Priority:     150,
			HandlerAgent: "", // deliberately none — see remit.go
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "capability_gap:content_duplication_rewrite",
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// in-remit: identical content on one page
// ---------------------------------------------------------------------------

type identicalPageGroup struct {
	PageID   uuid.UUID
	PageName string
	Keep     siteSection
	Groups   [][]siteSection // one entry per distinct duplicated text
}

func (g identicalPageGroup) redundantComponentIDs() []string {
	var out []string
	for _, grp := range g.Groups {
		for _, s := range grp[1:] { // grp is position-sorted; keep the first
			out = append(out, s.ComponentID.String())
		}
	}
	sort.Strings(out)
	return out
}

func findIdenticalSamePage(sections []siteSection) []identicalPageGroup {
	byPage := map[uuid.UUID][]siteSection{}
	for _, s := range sections {
		if len(s.Text) < 80 {
			continue // too short to be a meaningful duplicate
		}
		byPage[s.PageID] = append(byPage[s.PageID], s)
	}

	var out []identicalPageGroup
	for pageID, secs := range byPage {
		byText := map[string][]siteSection{}
		for _, s := range secs {
			// Identity = same slot + byte-identical blob. The >=80-char floor
			// above uses the normalised PROSE (is this a meaningful section at
			// all); identity uses the BYTES (is there anything to decide). See
			// datahelpers.SectionIdentityKey — both halves must use this.
			k := datahelpers.SectionIdentityKey(s.SlotName, s.Raw)
			byText[k] = append(byText[k], s)
		}
		var groups [][]siteSection
		for _, grp := range byText {
			if len(grp) < 2 {
				continue
			}
			sort.Slice(grp, func(i, j int) bool { return grp[i].Position < grp[j].Position })
			groups = append(groups, grp)
		}
		if len(groups) == 0 {
			continue
		}
		sort.Slice(groups, func(i, j int) bool { return groups[i][0].Position < groups[j][0].Position })
		out = append(out, identicalPageGroup{
			PageID:   pageID,
			PageName: secs[0].PageName,
			Keep:     groups[0][0],
			Groups:   groups,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PageName < out[j].PageName })
	return out
}

// ---------------------------------------------------------------------------
// residue: everything that needs judgement
// ---------------------------------------------------------------------------

type residueReport struct {
	CrossPageIdentical int
	NearDuplicatePairs int
	FactOverlapPairs   int
	FactCensusBlind    bool
	Examples           []string
}

func (r residueReport) total() int {
	return r.CrossPageIdentical + r.NearDuplicatePairs + r.FactOverlapPairs
}

func findResidue(sections []siteSection, facts []string) residueReport {
	rep := residueReport{FactCensusBlind: len(facts) < minFactsForFactCensus}

	var usable []siteSection
	for _, s := range sections {
		if len(s.Text) >= 120 {
			usable = append(usable, s)
		}
	}

	for i := 0; i < len(usable); i++ {
		for j := i + 1; j < len(usable); j++ {
			a, b := usable[i], usable[j]
			if a.PageID == b.PageID {
				continue // same-page identical is in-remit; same-page near-dupes
				// are the residue's weakest signal and are omitted deliberately
				// to keep this report about the cross-page failure 151 describes.
			}
			switch {
			case a.Text == b.Text:
				rep.CrossPageIdentical++
				rep.addExample(a, b, "identical")
			case jaccard(a.Tokens, b.Tokens) >= jaccardResidueThreshold:
				rep.NearDuplicatePairs++
				rep.addExample(a, b, "near-duplicate")
			}
			if !rep.FactCensusBlind && sharedFactCount(a, b) >= 3 {
				rep.FactOverlapPairs++
				rep.addExample(a, b, "shares 3+ facts")
			}
		}
	}
	return rep
}

func (r *residueReport) addExample(a, b siteSection, why string) {
	if len(r.Examples) >= 12 {
		return // a report, not a dump; the counts above are the measurement
	}
	r.Examples = append(r.Examples, fmt.Sprintf("%s/%s <-> %s/%s (%s)",
		a.PageName, a.SlotName, b.PageName, b.SlotName, why))
}

func sharedFactCount(a, b siteSection) int {
	set := map[string]struct{}{}
	for _, f := range a.FactIDs {
		set[f] = struct{}{}
	}
	n := 0
	for _, f := range b.FactIDs {
		if _, ok := set[f]; ok {
			n++
		}
	}
	return n
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	small, large := a, b
	if len(b) < len(a) {
		small, large = b, a
	}
	for t := range small {
		if _, ok := large[t]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

// ---------------------------------------------------------------------------
// loading and normalising
// ---------------------------------------------------------------------------

func loadSiteSections(dctx DiscoveryCheckContext) ([]siteSection, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.id, p.id, p.name, COALESCE(pc.slot_name, ''), pc.position,
		       COALESCE(pc.content_data::text, '{}')
		FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.content_data IS NOT NULL
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []siteSection
	for rows.Next() {
		var s siteSection
		var raw string
		if err := rows.Scan(&s.ComponentID, &s.PageID, &s.PageName, &s.SlotName, &s.Position, &raw); err != nil {
			return nil, err
		}
		s.Text = datahelpers.NormaliseSectionText(raw)
		s.Raw = raw
		s.Tokens = datahelpers.SectionTokenSet(s.Text)
		out = append(out, s)
	}
	return out, rows.Err()
}

func loadEvidenceFactIDs(dctx DiscoveryCheckContext) ([]string, error) {
	var raw sql.NullString
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data::text FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current
	`, dctx.SiteID).Scan(&raw)
	if err == sql.ErrNoRows || !raw.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var eb struct {
		Facts []struct {
			ID           string        `json:"id"`
			Value        interface{}   `json:"value"`
			Claim        string        `json:"claim"`
			ContextTerms []string      `json:"context_terms"`
			_            []interface{} `json:"-"`
		} `json:"facts"`
	}
	if err := json.Unmarshal([]byte(raw.String), &eb); err != nil {
		return nil, err
	}
	var ids []string
	for _, f := range eb.Facts {
		if f.ID != "" {
			ids = append(ids, f.ID)
		}
	}
	factDetail = map[string]factMatcher{}
	for _, f := range eb.Facts {
		factDetail[f.ID] = factMatcher{Value: fmt.Sprintf("%v", f.Value), Terms: f.ContextTerms, Claim: f.Claim}
	}
	return ids, nil
}

type factMatcher struct {
	Value string
	Terms []string
	Claim string
}

// factDetail is populated by loadEvidenceFactIDs for the current run. A check
// instance is used by one site at a time within a run (registry.go dispatches
// sequentially), so this is safe; it is a cache, not shared state across sites.
var factDetail = map[string]factMatcher{}

// assignFacts marks which approved facts each section asserts. A section asserts
// a fact when the fact's VALUE appears as a standalone token AND at least one of
// its context terms is present. The value alone is far too loose — "3" matches
// almost any prose — which is the trap the hand-run census hit first.
func assignFacts(sections []siteSection, facts []string) {
	if len(facts) == 0 {
		return
	}
	for i := range sections {
		for _, id := range facts {
			m, ok := factDetail[id]
			if !ok || m.Value == "" {
				continue
			}
			if !containsStandalone(sections[i].Text, m.Value) {
				continue
			}
			if len(m.Terms) > 0 && !anyTermPresent(sections[i].Text, m.Terms) {
				continue
			}
			sections[i].FactIDs = append(sections[i].FactIDs, id)
		}
	}
}

func containsStandalone(text, value string) bool {
	idx := strings.Index(text, value)
	for idx >= 0 {
		beforeOK := idx == 0 || !isNumberish(text[idx-1])
		after := idx + len(value)
		afterOK := after >= len(text) || !isNumberish(text[after])
		if beforeOK && afterOK {
			return true
		}
		next := strings.Index(text[idx+1:], value)
		if next < 0 {
			return false
		}
		idx = idx + 1 + next
	}
	return false
}

func isNumberish(b byte) bool { return (b >= '0' && b <= '9') || b == '.' }

func anyTermPresent(text string, terms []string) bool {
	for _, t := range terms {
		if t != "" && strings.Contains(text, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// verifier
// ---------------------------------------------------------------------------

// VerifyContentDuplicationResolved re-runs the in-remit predicate for the item's
// page and reports resolved only when no content-identical group remains.
//
// Unlike most verifiers in this package, rows DISAPPEARING is the expected shape
// of the fix here — the remedy is deletion — so a reduced row count is not the
// ambiguous signal it is for empty_section. What IS ambiguous is the page itself
// vanishing, and that is an error rather than a success.
func VerifyContentDuplicationResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	if target.PageID == nil {
		return VerifyResult{}, fmt.Errorf("content_duplication item %s has no page_id; cannot locate the defect", target.ItemID)
	}

	var exists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM pages WHERE id = $1)`, *target.PageID,
	).Scan(&exists); err != nil {
		return VerifyResult{}, fmt.Errorf("page existence check: %w", err)
	}
	if !exists {
		// AMBIGUOUS, and never reported as success: a deleted page is equally the
		// signature of a fix and of this platform's most repeated failure, a
		// rebuild removing content. Failing open records the unknown.
		return VerifyResult{}, fmt.Errorf("page %s no longer exists — cannot distinguish a fix from content loss", *target.PageID)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(pc.content_data::text, '{}')
		FROM page_components pc
		WHERE pc.page_id = $1 AND pc.content_data IS NOT NULL
	`, *target.PageID)
	if err != nil {
		return VerifyResult{}, err
	}
	defer rows.Close()

	counts := map[string]int{}
	total := 0
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return VerifyResult{}, err
		}
		text := datahelpers.NormaliseSectionText(raw)
		if len(text) < 80 {
			continue
		}
		counts[text]++
		total++
	}
	if err := rows.Err(); err != nil {
		return VerifyResult{}, err
	}

	dupeGroups, redundant := 0, 0
	for _, n := range counts {
		if n > 1 {
			dupeGroups++
			redundant += n - 1
		}
	}
	if dupeGroups > 0 {
		return VerifyResult{
			Resolved: false,
			Detail: fmt.Sprintf("still %d content-identical group(s) on this page, %d redundant section(s) of %d compared",
				dupeGroups, redundant, total),
		}, nil
	}
	return VerifyResult{
		Resolved: true,
		Detail:   fmt.Sprintf("no content-identical sections remain (%d sections compared)", total),
	}, nil
}
