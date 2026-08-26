// FILE: platform/orchestration/actions/cta_label_audit.go
//
// bugs_open/399. The framework computes each CTA's real destination, writes the
// destination's TITLE into the same jsonb object as the copy, and hands that
// title to the content writer specifically so it can "write CTA copy FOR the
// actual destination instead of guessing one"
// (resolve_internal_links_action.go:335-339). Nothing ever checked whether it
// did. This pass asks, once, before the row is persisted.
//
// WHY A CHECK AND NOT A BETTER PROMPT, measured rather than assumed. The writer
// prompt ALREADY carries the instruction this enforces — migration 476 arms
// stampCTADestinationGuidance, 477 made it actually reach the prompt on
// 2026-08-20, and the sentence reads "Destination (fixed): <title>. Write this
// CTA's text to name or clearly promise this destination; never promise a
// different one." [MEASURED 2026-08-26] 781 of 2,297 page-content-writer
// prompts over three days carry that literal, and of the CTA pairs written
// SINCE the pipe went live, 155 of 1,060 (14.6%) still contradict their
// destination. That number could have come out near zero and did not.
// Prompt text is not a control.
//
// WHY IT RECORDS AND DOES NOT REFUSE OR REPAIR. Both were designed and both
// were rejected on measurement, not taste:
//
//   - REPAIR is nearly unreachable. [MEASURED 2026-08-26] of 186 mismatched
//     pairs live, the copy names exactly one other page in 13, names two or
//     more (RFC_047: refuse, never guess) in 78, and names no page at all in
//     95. An automatic repoint would reach 7% while inheriting bugs_open/248's
//     clobber — a repair turned a correct /contact.html into a wrong link on
//     2026-08-24.
//   - REFUSAL has nowhere to go. At 14.6% it would fail roughly one CTA write
//     in seven (~29 sections/day at the 2026-08-24/25 rate of ~200), nothing
//     could auto-satisfy it, and on a page that re-authors itself several times
//     a day one cosmetic mismatch becomes an indefinitely withheld refresh —
//     worse for the owner than the mislabelled button.
//
// WHY A RECORD AND NOT A WORK ITEM: bugs_closed/023's closure says it in terms
// — "more detection makes the invisible pile bigger" — and 78
// cta_names_unknown_destination plus 70 unresolved_cta rows already sit at
// needs_human_review with no handler. The same precedent writeContentDataLinkLog
// cites, one file along.
//
// ⚠ THE DELIVERABLE IS THE RATE, NOT THE ROW. Nobody should read 155
// individual records; somebody must notice if 14.6% becomes 30%, or falls to 2%
// after a prompt change. The reading obligation and its query live in the lane
// RUNBOOK and the register entry, because a record nobody reads is the exact
// class bugs_open/410 was filed for on 2026-08-26.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ctaLabelAuditErrorCode is its own code beside CONTENT_DATA_LINK_AUDIT for the
// reason that file gives for being distinct from the markup repair's: it
// answers a different question. Those rows say which stored links resolve to no
// page; these say which buttons resolve to a REAL page that is not the one
// their words name. Folding them together would silently change the population
// every existing query on the other code counts.
//
// Registered in architecture_review/finding_code_registry.json with disposition
// "instrumented" — deliberately NOT "consumed", which asserts a reader file and
// a sink table that do not exist yet.
const ctaLabelAuditErrorCode = "CTA_LABEL_MISMATCH"

// ctaLabelAuditConfigKey arms the pass. Opt-in with the UNSAFE DEFAULT OFF, per
// the owner ruling of 2026-08-02 (RFC_010/RFC_022): new authority on a shared
// seam ships as a field a reviewer of the CALLER can see. Unset or a non-bool
// value means today's behaviour, byte for byte.
const ctaLabelAuditConfigKey = "audit_cta_label_agreement"

// ctaLabelFinding is one button whose copy and destination disagree, or whose
// copy names two pages equally well. Both sides are carried because "the words
// say X and the link goes to Y" is only actionable with X and Y in hand — and
// because the two lanes that own the remedy need different halves of it:
// bugs_open/391 needs the destination (its ranking chose it), and
// cta_target_content_pass needs the label (its copy pass rewrites it).
type ctaLabelFinding struct {
	Slot        string `json:"slot"`
	Component   string `json:"component"`
	URLField    string `json:"url_field"`
	LabelField  string `json:"label_field"`
	Label       string `json:"label"`
	Destination string `json:"destination"`
	// TargetTitle is what the resolver believed the destination was, in the
	// operator's language. It is NOT an input to the judgement (see
	// datahelpers/cta_label_agreement.go for why comparing against it would be a
	// third predicate); it is here so a human reading the record can see the
	// contradiction the way the bug reported it.
	TargetTitle string `json:"target_title"`
	NamedURL    string `json:"named_url,omitempty"`
	NamedTitle  string `json:"named_title,omitempty"`
	Verdict     string `json:"verdict"`
	Ambiguous   bool   `json:"ambiguous,omitempty"`
}

// auditSectionCTALabels is the whole decision, extracted pure so it is testable
// without a database — the shape shouldRefuseDeadURLControls and
// auditSectionContentDataLinks both use.
//
// schemaOf maps a component id to its parsed input_schema. The pairing of a url
// field to its LABEL field is read from the schema and never guessed: measured
// on the live fleet 2026-08-26, cta_target_title pairs with cta_text,
// secondary_cta_target_title with bare secondary_cta, and
// cta_primary_target_title with cta_primary_label — three different rules, so
// any string-manipulation shortcut is wrong on two of them.
//
// A pair is judged only when the label, the destination AND the target title are
// all present. The title's presence is the SCOPE TEST: it marks a destination
// the resolver actually wrote, which is the population bugs_open/399 is about.
// A CTA-bearing component the resolver never covered (system-stats, tool-cta,
// featured-content, tool-list … 5, 5, 4 and 3 rows respectively as of
// 2026-08-26) carries no title, is skipped here, and remains unresolved_cta /
// bugs_open/389 territory. Do not read this pass as covering them.
func auditSectionCTALabels(sections []SectionData, schemaOf map[string]map[string]interface{},
	candidates []datahelpers.LabelMatchCandidate, pageName, pageURL string) []ctaLabelFinding {
	if len(candidates) == 0 {
		return nil
	}
	var out []ctaLabelFinding
	for i := range sections {
		s := &sections[i]
		if len(s.ContentData) == 0 {
			continue
		}
		schema := schemaOf[s.ComponentID]
		if len(schema) == 0 {
			continue
		}
		for _, cf := range datahelpers.DeriveCTAURLFields(schema) {
			if cf.LabelField == "" {
				continue
			}
			label, _ := s.ContentData[cf.LabelField].(string)
			destination, _ := s.ContentData[cf.URLField].(string)
			title, _ := s.ContentData[ctaTargetTitleField(cf.URLField)].(string)
			if label == "" || destination == "" || title == "" {
				continue
			}
			// SCOPE IS THIS CALLER'S, and it is deliberately the same remit the
			// misdirected-CTA detector applies: internal page links only. A
			// tel:/mailto: destination is bugs_open/299's class and already has
			// its own live check (check_cta_nonpage), which asks this same
			// question through the same predicate with a different filter.
			// Judging one here would double-file it.
			if scope := datahelpers.ClassifyLinkScope(destination); scope != datahelpers.LinkScopePage {
				continue
			}
			j := datahelpers.JudgeCTALabel(label, destination, candidates, pageName, pageURL)
			if j.Verdict == datahelpers.CTALabelAgrees {
				continue
			}
			if j.Verdict == datahelpers.CTALabelNoOpinion && !j.Ambiguous {
				continue // generic, unmatched or self-naming copy: no signal
			}
			out = append(out, ctaLabelFinding{
				Slot:        s.ComponentName,
				Component:   s.ComponentName,
				URLField:    cf.URLField,
				LabelField:  cf.LabelField,
				Label:       label,
				Destination: destination,
				TargetTitle: title,
				NamedURL:    j.Named.URL,
				NamedTitle:  j.Named.Title,
				Verdict:     j.Verdict.String(),
				Ambiguous:   j.Ambiguous,
			})
		}
	}
	return out
}

// countCTALabelFindings splits the findings by arm for the log line and the
// record: a contradiction is a defect, an ambiguity is a button nobody can
// decide. They must not be summed — RFC_047 §10 ruled the second belongs to an
// agent that knows the site's premise, so a count that merges them would
// misreport the size of the actionable half.
func countCTALabelFindings(findings []ctaLabelFinding) (contradicts, ambiguous int) {
	for _, f := range findings {
		if f.Ambiguous {
			ambiguous++
			continue
		}
		contradicts++
	}
	return contradicts, ambiguous
}

// loadSectionInputSchemas reads the input_schema of every component on the page
// in ONE query. Returns an empty map on any failure: this pass is an
// instrument, and an instrument that can fail a save is a worse defect than the
// one it measures.
func loadSectionInputSchemas(ctx context.Context, db *sql.DB, sections []SectionData,
	logger *zap.Logger) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	if db == nil {
		return out
	}
	ids := make([]string, 0, len(sections))
	seen := map[string]bool{}
	for i := range sections {
		id := sections[i].ComponentID
		if id == "" || seen[id] {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return out
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id::text, input_schema FROM content_components WHERE id = ANY($1::uuid[])`,
		pqStringArray(ids))
	if err != nil {
		logger.Warn("cta label audit: component schema read failed; pass skipped", zap.Error(err))
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			continue
		}
		if schema := datahelpers.ParseInputSchemaValue(string(raw)); len(schema) > 0 {
			out[id] = schema
		}
	}
	return out
}

// auditCTALabelAgreement is the wired entry point: config gate, loads, judge,
// record. Best-effort throughout — every failure path returns quietly, because
// nothing here may fail a save whose content is otherwise correct.
func auditCTALabelAgreement(ctx context.Context, params ActionParams, siteID uuid.UUID,
	domain, pageName, pageURL string, sections []SectionData, logger *zap.Logger) {
	if params.DB == nil || len(sections) == 0 {
		return
	}
	if on, _ := params.StepConfig.Config[ctaLabelAuditConfigKey].(bool); !on {
		return
	}

	// The SHARED universe (LNK-036), not a private candidate list: the detector
	// that reports a wrong button and this pass that watches one being written
	// must answer "which pages may this label name?" from one place, or they
	// will disagree about the same button.
	candidates, err := datahelpers.LoadCTALabelUniverse(ctx, params.DB, siteID)
	if err != nil {
		logger.Warn("cta label audit: label universe unavailable; pass skipped",
			zap.String("page_name", pageName), zap.Error(err))
		return
	}
	schemaOf := loadSectionInputSchemas(ctx, params.DB, sections, logger)
	if len(schemaOf) == 0 {
		return
	}

	findings := auditSectionCTALabels(sections, schemaOf, candidates, pageName, pageURL)
	if len(findings) == 0 {
		return
	}
	contradicts, ambiguous := countCTALabelFindings(findings)

	// Distinctive compiled string: the pod-grep marker for this pass.
	logger.Info("SavePageSectionsAction: audited CTA label/destination agreement before persist",
		zap.String("page_name", pageName),
		zap.String("domain", domain),
		zap.Int("contradicts", contradicts),
		zap.Int("ambiguous", ambiguous))

	writeCTALabelAuditLog(ctx, params, siteID, domain, pageName, pageURL,
		findings, contradicts, ambiguous, logger)
}

// writeCTALabelAuditLog persists the audit on the success path, through the one
// shared writer (RFC_012 B retired the hand-copied INSERT class; do not
// reintroduce one here).
func writeCTALabelAuditLog(ctx context.Context, params ActionParams, siteID uuid.UUID,
	domain, pageName, pageURL string, findings []ctaLabelFinding,
	contradicts, ambiguous int, logger *zap.Logger) {
	if len(findings) == 0 {
		return
	}
	payload, err := json.Marshal(findings)
	if err != nil {
		logger.Warn("cta label audit: findings unserialisable; record skipped", zap.Error(err))
		return
	}
	if !LogActionEntry(ctx, params, agenterrors.Entry{
		SiteID:    siteID.String(),
		Domain:    domain,
		AgentType: params.AgentType,
		StepName:  saveSectionsStepName(params),
		Action:    "save_page_sections",
		ErrorMessage: fmt.Sprintf(
			"%d CTA button(s) on %s do not agree with their recorded destination: "+
				"%d whose copy names a different page, %d whose copy names two pages equally "+
				"(refused, not guessed — RFC_047); see context.findings",
			len(findings), pageName, contradicts, ambiguous),
		ErrorCode: ctaLabelAuditErrorCode,
		Severity:  "warning",
		Context: map[string]interface{}{
			"page_name":   pageName,
			"page_url":    pageURL,
			"contradicts": contradicts,
			"ambiguous":   ambiguous,
			"findings":    json.RawMessage(payload),
		},
	}, logger) {
		logger.Warn("cta label audit: durable record not written",
			zap.String("page_name", pageName))
	}
}
