// FILE: cmd/brief-negation-check/specclaims.go
//
// THE SECOND DETECTOR: a spec that ORDERS A PAGE TO SAY something the page gate
// would refuse (bugs_open/414).
//
// WHY IT IS HERE AND NOT A NEW SERVICE. This binary is already the estate's only
// mechanism that reads site_specs CONTENT and judges it, and it already owns the
// three hard parts: the visible surface derived from live agent prompts rather
// than a hardcoded list, a census of current specs, and filing/close-out that
// satisfies both work-item landmines. A second CronJob would have duplicated all
// of it, plus four makefile edits and five deployment files, to ask a sibling
// question about the same text. The binary's NAME says "negation"; its remit is
// "an instruction in a spec, in the text a generator actually reads", and this is
// the second instruction class in that remit.
//
// WHAT 414 WAS. A shadow experiment planted a tripwire in one site's
// content_direction: "include the exact phrase: checked against the FCA handbook,
// rule by rule". The writer obeyed — that is what the tripwire tested — and the
// sentence was served on a finance site as an unverifiable claim of regulatory
// diligence for 24 days. Then the estate's own audit fleet read the served copy
// back and filed work to REINFORCE it as the site's differentiator. Nothing in
// the platform looked at the instruction; every claims check looks at output.
//
// THE PROPAGATION IS THE REASON THIS IS A DAILY CHECK AND NOT A ONE-OFF QUERY.
// Stripping the plant from content_direction on 2026-08-26 did NOT retract it:
// `domain-strategist` had already restated it, in its own prose, in the
// `strategy` aspect on 2026-08-12 — a different aspect, read by different agents
// (build-site-planner, webdesign-agent), which no reader of the fixed row would
// think to check. An instruction, once in a spec, is copied by the agents that
// read specs.
//
// ---------------------------------------------------------------------------
// WHAT IT SCANS WITH, AND THE MEASUREMENT THAT DECIDED IT
// ---------------------------------------------------------------------------
// It scans with the PRACTICE FAMILY ONLY (datahelpers.ScanPracticeClaims):
// first-person claims about what this business does, including the documentary
// diligence pattern P6 that 414 added. It does NOT scan with the fleet-wide
// refusing set, and that is measured, not preference:
//
//	the fleet-wide + regulated patterns over all 522 current spec rows
//	[MEASURED 2026-08-26] ....... 21 hits, effectively every one false
//
// Fifteen of those are the estate's OWN honesty instructions — "Never invent a
// person, company, scheme, study, quotation, or statistic" — matching the
// `never (wrong|inaccurate|invents?…)` pattern; the negation guard cannot save
// them, because the match STARTS at "never" and the guard only looks backwards.
// Worse, `evidence_base` rows store each site's `banned_claims` AS DATA, quoting
// the forbidden sentence verbatim in the reason field, so a generic spec scan
// convicts every site's own immune system, daily, for ever. This is also the
// census this file's sibling header records as tried and withdrawn within hours
// on 2026-08-19. Do not re-run it.
//
// The practice family over the same text, by contrast:
//
//	current spec rows (532) ................................ 0 hits  [2026-08-27]
//	ALL spec rows including superseded history (2,782) ..... 2 hits  [2026-08-27]
//
// and those two are exactly the two hops of the observed chain — the original
// plant in content_direction (61ef7033…) and the strategist's copy of it in
// strategy (96eaff0b…). So this detector would have caught the plant on the day
// it was made AND the propagation ten days later, with nothing else to read.
// That is the demand control: a check whose only clean result is "we fixed it
// this morning" has proved nothing, so the positive case is pinned in history.
//
// STATED RESIDUAL. A spec that mandates a COMPLETENESS claim ("make the page say
// everything here is checked") is not in the practice family and is missed here.
// The page gate catches it at blocker severity on the output side, which is the
// half that stops it being served — but the instruction would survive in the
// spec, refusing every rebuild until a human reads the refusal. If that happens,
// the fix is a third arm here, not a widening of what a spec is scanned with.
//
// SEVERITY AND HANDLER. Filed at `needs_human_review` with NO handler agent, and
// that is deliberate to the point of being the whole design: an automated handler
// rewriting a spec off a spec-content finding is precisely how the audit fleet
// canonised the marker in the first place. A person decides what a brief should
// say.
package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const (
	// specClaimItemType is classified in
	// platform/orchestration/actions/discovery_checks/verifier_coverage_test.go
	// (the map AND the liveItemTypes ratchet, same commit — the build fails
	// otherwise), named for its sibling brief_supplies_negation.
	specClaimItemType = "spec_supplies_claim"

	// evidenceBaseAspect is excluded from the scan: its content IS the ban list,
	// so scanning it convicts the register for quoting what it forbids. It is
	// still CENSUSED, because it carries the operating-history attestation that
	// exempts the family.
	evidenceBaseAspect = "evidence_base"
)

// allAgentsConfigSQL reads every live agent's config, one row per agent, so the
// visible surface is the union across the fleet and each finding can name the
// agents that would read it.
//
// ⚠ THE WRITER-ONLY SURFACE WOULD HAVE MISSED 414's SURVIVING HALF. The sibling
// detector derives its surface from page-content-writer alone, correctly — a
// define-by-negation phrase matters because the WRITER reads it. A planted
// instruction does not need the writer: the copy that outlived the 08-26 fix sat
// in `strategy`, which the writer never reads and build-site-planner does.
// [MEASURED 2026-08-27, union over live agent prompts: identity ×12 agents,
// content_direction ×5, classification ×4, evidence_base ×3, design_intent ×3,
// mission_brief ×2, site_archetype ×2, strategy ×2, roadmap_brief ×2, and one
// each for briefing, design, design_reference, webdesign.]
const allAgentsConfigSQL = `SELECT COALESCE(jsonb_agg(jsonb_build_object(
                                'type', type, 'config', default_config::text)), '[]'::jsonb)
                              FROM agent_definitions
                             WHERE is_active AND COALESCE(is_snapshot,false) = false
                               AND deleted_at IS NULL`

type agentConfig struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

// fleetSurface returns the spec paths any live agent reads, mapped to the agents
// that read them, plus the sorted aspect list for the census.
func fleetSurface(raw string) (map[string][]string, []string, error) {
	var agents []agentConfig
	if err := json.Unmarshal([]byte(raw), &agents); err != nil {
		return nil, nil, fmt.Errorf("agent config census would not decode: %w", err)
	}
	readers := map[string]map[string]bool{}
	for _, a := range agents {
		for _, p := range visibleSurface(a.Config) {
			if readers[p] == nil {
				readers[p] = map[string]bool{}
			}
			readers[p][a.Type] = true
		}
	}
	surface := map[string][]string{}
	aspectSet := map[string]bool{}
	var aspects []string
	for p, set := range readers {
		names := make([]string, 0, len(set))
		for n := range set {
			names = append(names, n)
		}
		sort.Strings(names)
		surface[p] = names
		a := strings.Split(p, ".")[0]
		if !aspectSet[a] {
			aspectSet[a] = true
			aspects = append(aspects, a)
		}
	}
	sort.Strings(aspects)
	return surface, aspects, nil
}

// specClaim is one claim a spec hands to a generator.
type specClaim struct {
	Aspect   string   `json:"aspect"`
	Field    string   `json:"field"`
	Readers  []string `json:"read_by"`
	Matched  string   `json:"matched"`
	Snippet  string   `json:"snippet"`
	Reason   string   `json:"reason"`
	Mandated bool     `json:"mandated"`
}

// specClaimAssessment is one site's answer.
type specClaimAssessment struct {
	Domain   string      `json:"domain"`
	SiteID   string      `json:"site_id"`
	Claims   []specClaim `json:"claims"`
	Attested bool        `json:"operating_history_attested"`
	Scanned  int         `json:"scanned_fields"`
}

func (a specClaimAssessment) Finding() bool { return len(a.Claims) > 0 }

// visibleFieldPaths reports whether a leaf path inside an aspect is on the
// surface, and who reads it. A bare-aspect reference ("{{.site_specs.specs.strategy}}")
// injects the whole document, so every leaf under it is visible.
func visibleFieldPaths(surface map[string][]string, aspect, leaf string) ([]string, bool) {
	if r, ok := surface[aspect+"."+leaf]; ok {
		return r, true
	}
	if r, ok := surface[aspect]; ok {
		return r, true
	}
	return nil, false
}

// assessSpecClaims is the predicate — pure, and the part the tests hold.
//
// It walks the string leaves of each visible aspect (one level of nesting, which
// is what spec documents actually are), splits each with the PLAIN-TEXT
// assertion splitter — never the HTML one, and never by joining the document:
// joining fuses the end of one sentence to the start of the next and every
// pattern in this layer is windowed, so a joined scan reports sentences the
// document does not contain — and scans the blocks with the practice family.
func assessSpecClaims(rows []siteSpecs, surface map[string][]string) []specClaimAssessment {
	type acc struct {
		a  specClaimAssessment
		eb *datahelpers.EvidenceBase
	}
	bySite := map[string]*acc{}
	order := []string{}

	// First pass: the attestation, which exempts the whole family for a site.
	for _, r := range rows {
		if r.Aspect != evidenceBaseAspect {
			continue
		}
		b, err := json.Marshal(r.Data)
		if err != nil {
			continue
		}
		eb, err := datahelpers.ParseEvidenceBase(b)
		if err != nil {
			continue
		}
		if bySite[r.SiteID] == nil {
			bySite[r.SiteID] = &acc{a: specClaimAssessment{Domain: r.Domain, SiteID: r.SiteID}}
			order = append(order, r.SiteID)
		}
		bySite[r.SiteID].eb = eb
	}

	for _, r := range rows {
		if r.Aspect == evidenceBaseAspect {
			continue
		}
		if bySite[r.SiteID] == nil {
			bySite[r.SiteID] = &acc{a: specClaimAssessment{Domain: r.Domain, SiteID: r.SiteID}}
			order = append(order, r.SiteID)
		}
		cur := bySite[r.SiteID]
		cur.a.Attested = cur.eb.OperatingHistoryAttested()

		for key, v := range r.Data {
			readers, visible := visibleFieldPaths(surface, r.Aspect, key)
			if !visible {
				continue
			}
			for field, text := range stringLeaves(key, v) {
				cur.a.Scanned++
				for _, block := range blockSplitRe.Split(text, -1) {
					blocks := datahelpers.SplitPlainAssertionText(block)
					for _, f := range datahelpers.ScanPracticeClaims(blocks, cur.eb) {
						cur.a.Claims = append(cur.a.Claims, specClaim{
							Aspect:   r.Aspect,
							Field:    field,
							Readers:  readers,
							Matched:  f.Matched,
							Snippet:  f.Snippet,
							Reason:   f.Reason,
							Mandated: mandateRe.MatchString(block),
						})
					}
				}
			}
		}
	}

	out := make([]specClaimAssessment, 0, len(order))
	for _, id := range order {
		a := bySite[id].a
		sort.Slice(a.Claims, func(i, j int) bool {
			if a.Claims[i].Field != a.Claims[j].Field {
				return a.Claims[i].Field < a.Claims[j].Field
			}
			return a.Claims[i].Matched < a.Claims[j].Matched
		})
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Domain < out[j].Domain })
	return out
}

// stringLeaves flattens one spec value to field-path → text, one level of
// nesting plus arrays of strings. Deliberately not arbitrarily deep: spec
// documents are shallow, and an unbounded walk would scan machine fields
// (ids, urls, timestamps) whose text is not prose anybody reads.
func stringLeaves(prefix string, v interface{}) map[string]string {
	out := map[string]string{}
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) != "" {
			out[prefix] = t
		}
	case []interface{}:
		for i, e := range t {
			if s, ok := e.(string); ok && strings.TrimSpace(s) != "" {
				out[fmt.Sprintf("%s[%d]", prefix, i)] = s
			}
		}
	case map[string]interface{}:
		for k, e := range t {
			switch s := e.(type) {
			case string:
				if strings.TrimSpace(s) != "" {
					out[prefix+"."+k] = s
				}
			case []interface{}:
				for i, el := range s {
					if str, ok := el.(string); ok && strings.TrimSpace(str) != "" {
						out[fmt.Sprintf("%s.%s[%d]", prefix, k, i)] = str
					}
				}
			}
		}
	}
	return out
}

// specClaimKey is keyed on the FINDING, not the site — the sibling's reasoning
// exactly: correcting one claim and leaving another files a NEW item describing
// what is actually there, and the old one closes because it no longer describes
// the spec. That keeps both landmines satisfied at once (never silently drop a
// finding; never rewrite an open row daily, which bumps updated_at and makes it
// unreapable for ever — bugs_closed/213).
func specClaimKey(siteID string, claims []specClaim) string {
	parts := make([]string, 0, len(claims))
	for _, c := range claims {
		parts = append(parts, c.Aspect+"|"+c.Field+"|"+strings.TrimSpace(c.Matched))
	}
	sort.Strings(parts)
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return fmt.Sprintf("spec-claims:%s:%x", siteID, sum[:6])
}

func fileSpecClaimFinding(db *sql.DB, a specClaimAssessment) (bool, error) {
	mandated := 0
	for _, c := range a.Claims {
		if c.Mandated {
			mandated++
		}
	}
	spec, _ := json.Marshal(map[string]interface{}{
		"check":          "spec_supplies_claim",
		"domain":         a.Domain,
		"claims":         a.Claims,
		"mandated_count": mandated,
		"scanned_fields": a.Scanned,
		"fix": "Human decision, and it is about the SPEC, not the page: a generator reads this text and " +
			"will say what it is told. Remove the claim from the spec (write the WHOLE aspect object, " +
			"never a patch — bugs_open/327), then check whether it has already been COPIED into another " +
			"aspect by an agent that reads specs: that is what kept bugs_open/414 alive after its source " +
			"was 'fixed'. Do NOT resolve this by substantiating the claim — the audit fleet already tried " +
			"that once and filed work to reinforce a planted tripwire.",
	})
	res, err := db.Exec(`
		INSERT INTO site_work_items
		  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
		   handler_agent, status, max_attempts, created_by, item_key)
		VALUES ($1, 'discovery', 'content', $2, 'high', $3, $4::jsonb, 45,
		        '', 'needs_human_review', 1, $5, $6)
		ON CONFLICT DO NOTHING`,
		a.SiteID, specClaimItemType,
		fmt.Sprintf("%s's specs hand a generator %d unverifiable claim(s) about the business (%d ordered onto pages)",
			a.Domain, len(a.Claims), mandated),
		string(spec), createdBy, specClaimKey(a.SiteID, a.Claims))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// closeAnsweredSpecClaims closes only on a POSITIVE re-observation. A site
// missing from the census, or one whose specs would not decode, closes nothing:
// an absence of evidence is not evidence of a fix.
func closeAnsweredSpecClaims(db *sql.DB, assessments []specClaimAssessment) ([]string, error) {
	rows, err := db.Query(`SELECT item_key, site_id::text FROM site_work_items
	                        WHERE item_type = $1 AND created_by = $2
	                          AND status NOT IN ('complete','cancelled','rejected')`,
		specClaimItemType, createdBy)
	if err != nil {
		return nil, err
	}
	type openItem struct{ key, site string }
	var open []openItem
	for rows.Next() {
		var k, sid string
		if err := rows.Scan(&k, &sid); err == nil {
			open = append(open, openItem{k, sid})
		}
	}
	rows.Close()

	current := map[string]string{}
	for _, a := range assessments {
		if a.Finding() {
			current[a.SiteID] = specClaimKey(a.SiteID, a.Claims)
		} else {
			current[a.SiteID] = ""
		}
	}

	var closed []string
	for _, o := range open {
		cur, examined := current[o.site]
		if !examined || cur == o.key {
			continue
		}
		reason := `{"closed_by":"brief-negation-check","reason":"re-examined: no visible spec field hands a generator an unverifiable claim about the business"}`
		if cur != "" {
			reason = `{"closed_by":"brief-negation-check","reason":"re-examined: the spec's claims have CHANGED, so this item no longer describes them — a fresh item carries the current set"}`
		}
		if _, err := db.Exec(`UPDATE site_work_items
		                         SET status = 'complete', updated_at = now(),
		                             result = COALESCE(result,'{}'::jsonb) || $2::jsonb
		                       WHERE item_key = $1 AND item_type = $3 AND created_by = $4
		                         AND status NOT IN ('complete','cancelled','rejected')`,
			o.key, reason, specClaimItemType, createdBy); err != nil {
			return closed, err
		}
		closed = append(closed, o.key)
	}
	return closed, nil
}

// renderSpecClaims is appended to the one report this binary writes, so a single
// doc_notes row per run still says everything both detectors saw.
func renderSpecClaims(assessments []specClaimAssessment, aspects []string, filed, closed []string) string {
	var b strings.Builder
	b.WriteString("\n\n---\n\n# spec_supplies_claim — specs that hand a generator an unverifiable claim (bugs_open/414)\n\n")
	b.WriteString("Surface: every aspect referenced by ANY live agent's prompt — not the writer's alone, " +
		"because 414's surviving copy sat in `strategy`, which the writer never reads:\n  ")
	b.WriteString(strings.Join(aspects, ", ") + "\n")
	b.WriteString("\nScanned with the PRACTICE family only (`ScanPracticeClaims`). The fleet-wide refusing " +
		"set is deliberately NOT used here: over 522 current spec rows it produced 21 hits, effectively " +
		"all false — 15 of them the estate's own \"never invent a person, company, scheme\" honesty " +
		"instructions, plus every site's `banned_claims` records quoting the phrases they forbid " +
		"[MEASURED 2026-08-26]. `evidence_base` is censused for its attestation and never scanned.\n")

	findings := 0
	for _, a := range assessments {
		if a.Finding() {
			findings++
		}
	}
	fmt.Fprintf(&b, "\n%d of %d sites hand a generator at least one such claim.\n", findings, len(assessments))
	for _, a := range assessments {
		if !a.Finding() {
			continue
		}
		b.WriteString("\n## " + a.Domain + "\n")
		for _, c := range a.Claims {
			tag := "claim"
			if c.Mandated {
				tag = "**MANDATED onto pages**"
			}
			fmt.Fprintf(&b, "- %s [`%s.%s`, read by %s] %q\n",
				tag, c.Aspect, c.Field, strings.Join(c.Readers, "/"),
				datahelpers.TruncateString(c.Snippet, 200))
		}
	}
	if len(filed) > 0 {
		b.WriteString("\nFiled: " + strings.Join(filed, ", ") + "\n")
	}
	if len(closed) > 0 {
		b.WriteString("Closed (re-examined and clean): " + strings.Join(closed, ", ") + "\n")
	}
	b.WriteString("\n### What a clean result here does and does not mean\n")
	b.WriteString("- It means no CURRENT spec visible to any generator carries a first-person claim about " +
		"the business that no register can adjudicate. Over spec history (2,782 rows) this predicate finds " +
		"exactly the two hops of 414's chain — the plant, and the strategist's copy of it ten days later " +
		"[MEASURED 2026-08-27] — so the zero is a measured zero, not a blind one.\n")
	b.WriteString("- It does NOT cover a spec mandating a COMPLETENESS claim (\"everything here is checked\"). " +
		"That shape is refused on the OUTPUT side at blocker severity, so it cannot be served — but the " +
		"instruction would survive here and refuse every rebuild until a human reads the refusal.\n")
	return b.String()
}
