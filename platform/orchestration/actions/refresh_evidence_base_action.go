// FILE: platform/orchestration/actions/refresh_evidence_base_action.go
//
// RefreshEvidenceBaseAction — V4 of the claims-verification layer
// (SPEC_claims_verification §5 V4): freshness for live-verifiable facts.
//
// A fact sourced from a SQL query goes stale the moment the underlying data
// moves. Published copy saying "2,767 records verified" stays true only while
// the database agrees. This action re-runs every `source.sql` fact in a site's
// evidence base, records the live value, and:
//
//   - updates `value` and `verified_at` on the fact (mechanical — the query IS
//     the source of truth for a sql-sourced fact);
//   - raises a `stale_evidence` work item when a live value moves outside the
//     fact's tolerance, INCLUDING when the site is now UNDER-claiming (the copy
//     says 2,767 and the database says 3,104 — not a lie, but worth knowing);
//   - regenerates the writer whitelist (`writer_block`, consumed by the
//     page-content-writer prompt — V2) so the numbers the writer is permitted
//     to assert can never quietly rot.
//
// Truth decisions stay human. This action updates a fact's NUMBER from its own
// declared source; it never invents a fact, never removes one, never edits
// published copy, and never rewrites the human-authored WORDS of the whitelist
// (see "Who owns what" below). Drift terminates at human review.
//
// ── Who owns what in the whitelist ──────────────────────────────────────────
//
// Humans own the words; the machine owns the numbers. Each fact may carry a
// `writer_line` — human-authored phrasing, with `{value}` where the number
// goes, and any caveat the audit demands ("This is a CATALOGUE count — never
// present it as a running fleet"). Regeneration substitutes fresh values into
// those lines; it never composes prose of its own beyond the three structural
// section headers. A fact with no `writer_line` is simply omitted from the
// whitelist — silence is the safe default for a fact nobody has phrased.
//
// Regeneration is opt-in per site via `writer_block_managed: true`. Where it is
// absent or false the hand-written `writer_block` is left exactly as it is, and
// any drift is reported in the work item so a human can update it.
//
// ── RFC_025: the other two source kinds ─────────────────────────────────────
//
// `sql` is not the only fact source this pass touches. `citation` facts
// (V5, SPEC_V5_researched_citations) are re-fetched and their quote re-checked
// — see refreshCitationFact. RFC_025 (ratified 2026-08-12, bugs_open/161's
// generalisable fix) adds two more, both opt-in and additive to today's
// behaviour for any fact that does not name them:
//
//   - `artifact` facts carrying the optional `artifact_check` key are re-proved
//     against their named stored artefact (today: a page_components row's
//     rendered_html) — see refreshArtifactCheckFact. An `artifact` fact with
//     no `artifact_check` is checked for presence only, exactly as before.
//   - `attested_by` facts get a staleness NUDGE on a ~180-day cadence — see
//     checkAttestationStaleness. A human's word cannot be re-proved by design,
//     so this only raises a `stale_attestation` item asking someone to re-look;
//     it never flags a mismatch, because it never checks one.
//
// ── SQL safety ──────────────────────────────────────────────────────────────
//
// These queries live in a data column, so they are treated as untrusted input
// even though the evidence base is human-owned and pinned:
//   - single statement only (no `;` other than a trailing one);
//   - must begin with SELECT — no CTEs, no DDL/DML verbs anywhere;
//   - executed inside a READ ONLY transaction with a statement timeout;
//   - must return exactly one row, one column, numeric.
// A query failing any of these is skipped and reported, never executed blind.
//
// Config:
//
//	"refresh_evidence_base": {
//	    "action": "refresh_evidence_base",
//	    "config": {
//	        "site_id": "site_record.site_id",   // optional — omit to sweep
//	                                            // every site with an evidence base
//	        "dry_run": false                    // optional — report, write nothing
//	    }
//	}
//
// Registration: "refresh_evidence_base" in registry.go (Category "site").

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RefreshEvidenceBaseInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"site_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("refresh_evidence_base", RefreshEvidenceBaseInputSpec)
}

// evidenceFactRefresh is one fact's outcome in a refresh pass.
type evidenceFactRefresh struct {
	FactID    string   `json:"fact_id"`
	Claim     string   `json:"claim"`
	StoredVal *float64 `json:"stored_value,omitempty"`
	LiveVal   *float64 `json:"live_value,omitempty"`
	Tolerance string   `json:"tolerance,omitempty"`
	Direction string   `json:"direction,omitempty"` // "up" | "down" | "unchanged"
	// Outcome: fresh | updated | drifted | error | attestation_due.
	// attestation_due (RFC_025 stage 1) is deliberately its OWN outcome rather
	// than reusing "drifted" — an attested_by fact going stale is not a detected
	// mismatch, it is silence past a cadence, and conflating the two would make
	// a human reading a stale_evidence item think something was found WRONG
	// when nobody has re-checked it at all.
	Outcome    string `json:"outcome"`
	Detail     string `json:"detail,omitempty"`
	VerifiedAt string `json:"verified_at,omitempty"`
	// EncodedByTools names the tool subject keys whose criteria fence declares
	// this fact (refresh_evidence_fact_drift.go), stamped only on drifted
	// entries and only when a declaration exists — omitempty, so a site with no
	// declarations marshals byte-identically to before Piece 3 existed.
	EncodedByTools []string `json:"encoded_by_tools,omitempty"`
}

// siteRefreshResult is one site's outcome.
type siteRefreshResult struct {
	SiteID          string `json:"site_id"`
	Domain          string `json:"domain"`
	FactsChecked    int    `json:"facts_checked"`
	FactsUpdated    int    `json:"facts_updated"`
	Drifted         int    `json:"drifted"`
	Errors          int    `json:"errors"`
	WriterBlock     string `json:"writer_block_action"` // regenerated | unmanaged | unchanged
	WorkItemCreated bool   `json:"work_item_created"`
	// WorkItemRefreshed is a SEPARATE field rather than a widening of
	// WorkItemCreated. A refresh is not a creation, and the whole subject of
	// bugs_open/091 is a reported field that meant something other than its name.
	WorkItemRefreshed bool                  `json:"work_item_refreshed"`
	Facts             []evidenceFactRefresh `json:"facts"`

	// AttestationsDue counts attested_by facts RFC_025 stage 1's staleness nudge
	// flagged this pass. Kept separate from Drifted/WorkItemCreated (below)
	// rather than folded into them: an attestation going stale is not the same
	// FINDING as a value drifting, and a reader must be able to tell "a human's
	// word needs re-dating" from "a machine-checked value moved" without
	// re-deriving it from the outcome strings inside Facts.
	AttestationsDue              int  `json:"attestations_due"`
	AttestationWorkItemCreated   bool `json:"attestation_work_item_created"`
	AttestationWorkItemRefreshed bool `json:"attestation_work_item_refreshed"`

	// ArtifactCheckDrifted counts artifact_check facts that outcome=="drifted"
	// this pass — a SUBSET of Drifted (which also counts pre-existing sql/
	// citation drift), tracked separately for one reason only: GATING the
	// stale_evidence raise below without changing when it fires for the
	// pre-existing sql/citation branch (a council architecture objection,
	// 2026-08-12 — decoupling that raise from `changed` for the EXISTING
	// citation branch was flagged as a behaviour change outside what RFC_025
	// was ratified to cover; this field is how that stays scoped to the NEW
	// mechanism only). Not surfaced in the item body itself — createStale
	// EvidenceItem scans res.Facts by Outcome=="drifted" directly, which
	// already includes these facts regardless of this counter.
	ArtifactCheckDrifted int `json:"artifact_check_drifted"`

	// FactDrift is Piece 3's fan-out (refresh_evidence_fact_drift.go): one
	// entry per (fact, declaring tool) that is owed something this pass, with
	// the route it took and the honest write outcome. FactDeclarationsUnresolved
	// lists "<subject_key>:<fact_id>" declarations the register cannot resolve —
	// inert by rule (PBP-037), surfaced so a typo is visible. Both omitempty:
	// the no-op site's JSON does not change.
	FactDrift                  []factDriftEmission `json:"fact_drift,omitempty"`
	FactDeclarationsUnresolved []string            `json:"fact_declarations_unresolved,omitempty"`
}

func RefreshEvidenceBaseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "refresh_evidence_base"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RefreshEvidenceBaseInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	dryRun := configBoolOrDefault(params.StepConfig.Config, "dry_run", false)

	// Target sites: the one named, or every site with a current evidence base.
	siteIDs, err := resolveEvidenceSites(ctx, params.DB, inputs.Get("site_id"), logger)
	if err != nil {
		return nil, err
	}
	if len(siteIDs) == 0 {
		logger.Info("refresh_evidence_base: no sites with an evidence base — nothing to do")
		return map[string]interface{}{"sites_checked": 0, "results": []interface{}{}}, nil
	}

	results := make([]siteRefreshResult, 0, len(siteIDs))
	totalDrift := 0
	for _, siteID := range siteIDs {
		res, err := refreshOneSiteEvidence(ctx, params, siteID, dryRun, logger)
		if err != nil {
			// One site's failure must not abort the sweep.
			logger.Warn("refresh_evidence_base: site refresh failed",
				zap.String("site_id", siteID.String()), zap.Error(err))
			continue
		}
		totalDrift += res.Drifted
		results = append(results, *res)
	}

	logger.Info("refresh_evidence_base: complete",
		zap.Int("sites_checked", len(results)),
		zap.Int("total_drifted", totalDrift),
		zap.Bool("dry_run", dryRun))

	return map[string]interface{}{
		"sites_checked": len(results),
		"total_drifted": totalDrift,
		"dry_run":       dryRun,
		"results":       results,
	}, nil
}

// resolveEvidenceSites returns the target site list: the named site, or every
// site holding a current evidence_base spec.
func resolveEvidenceSites(ctx context.Context, db *sql.DB, siteIDStr string, logger *zap.Logger) ([]uuid.UUID, error) {
	if siteIDStr != "" {
		id, err := uuid.Parse(siteIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
		}
		return []uuid.UUID{id}, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT site_id FROM site_specs
		WHERE aspect = 'evidence_base' AND is_current = true
	`)
	if err != nil {
		return nil, fmt.Errorf("list evidence-base sites: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			logger.Warn("refresh_evidence_base: site id scan failed", zap.Error(err))
			continue
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// shouldRaiseStaleEvidence decides whether a pass's drift, if any, should
// raise a stale_evidence work item — split out as a pure function (no DB, no
// side effect) specifically so the gating decision itself is unit-testable
// in isolation, per a council objection (2026-08-12) that the equivalent
// inline condition was verified only by manual code-reading, not a test.
//
// The rule, stated plainly: raise if something drifted AND (the register was
// also rewritten this pass OR the drift came from the NEW artifact_check
// mechanism specifically). The second disjunct is deliberately narrow — it
// must NOT let a pre-existing citation-only drift (changed=false,
// ArtifactCheckDrifted=0) raise, because that would be a behaviour change
// for an existing caller outside what RFC_025 was ratified to cover (see the
// call site's comment for the full reasoning).
func shouldRaiseStaleEvidence(res *siteRefreshResult, changed bool) bool {
	return res.Drifted > 0 && (changed || res.ArtifactCheckDrifted > 0)
}

func refreshOneSiteEvidence(
	ctx context.Context, params ActionParams, siteID uuid.UUID, dryRun bool, logger *zap.Logger,
) (*siteRefreshResult, error) {
	db := params.DB

	// The row id read here is the compare-and-swap token for the write below.
	var specRowID uuid.UUID
	var rawJSON []byte
	var pinned sql.NullBool
	err := db.QueryRowContext(ctx, `
		SELECT id, data, pinned FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, siteID).Scan(&specRowID, &rawJSON, &pinned)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("site has no current evidence_base")
	}
	if err != nil {
		return nil, fmt.Errorf("read evidence_base: %w", err)
	}

	// Work on the generic map so unknown keys (audit_doc, banned_claims,
	// allowed_entities, anything a future revision adds) survive the rewrite
	// untouched — this action must never drop a human's data.
	var eb map[string]interface{}
	if err := json.Unmarshal(rawJSON, &eb); err != nil {
		return nil, fmt.Errorf("unmarshal evidence_base: %w", err)
	}

	domain := loadSiteDomain(ctx, db, siteID, logger)
	res := &siteRefreshResult{SiteID: siteID.String(), Domain: domain, WriterBlock: "unchanged"}

	factsRaw, _ := eb["facts"].([]interface{})
	today := currentDateString(ctx, db)

	changed := false
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		// V5: citation facts are re-verified by re-fetching their source and
		// matching the stored verbatim quote (SPEC_V5_researched_citations §3c).
		if src, ok := fact["source"].(map[string]interface{}); ok {
			if _, has := src["citation"]; has {
				entry := refreshCitationFact(ctx, fact, today)
				if entry != nil {
					res.FactsChecked++
					switch entry.Outcome {
					case "drifted":
						res.Drifted++
					case "error":
						res.Errors++
					}
					if entry.Outcome == "updated" || (entry.Outcome == "fresh" && entry.VerifiedAt == today) {
						res.FactsUpdated++
						changed = true
					}
					res.Facts = append(res.Facts, *entry)
				}
				continue
			}
			// RFC_025 stage 2: an `artifact` fact carrying the optional
			// `artifact_check` key is re-proved against its named stored
			// artefact, the same shape refreshCitationFact re-proves a citation
			// against its URL. Read straight off this untyped map — deliberately
			// NOT a field on EvidenceSource/EvidenceFact (RFC_025 §9 Q2, the
			// ratified constraint: the whole point of this design over the
			// rejected typed-field alternative is to avoid touching an exported
			// symbol other packages depend on).
			if _, has := src["artifact_check"]; has {
				entry := refreshArtifactCheckFact(ctx, db, siteID, fact, today)
				if entry != nil {
					res.FactsChecked++
					switch entry.Outcome {
					case "drifted":
						res.Drifted++
						// Counted separately from res.Drifted (which also
						// reports it, for the item body's sake — see
						// createStaleEvidenceItem, which scans res.Facts by
						// Outcome, not this counter). This one exists purely
						// to GATE the raise below independently of `changed`
						// — see the architecture council objection recorded
						// at the gating site.
						res.ArtifactCheckDrifted++
					case "error":
						res.Errors++
					}
					if entry.Outcome == "fresh" && entry.VerifiedAt == today {
						res.FactsUpdated++
						changed = true
					}
					res.Facts = append(res.Facts, *entry)
				}
				continue
			}
		}

		query := factSQLSource(fact)
		if query == "" {
			// artifact/attested facts with no artifact_check are checked for
			// presence, not re-proved. RFC_025 stage 1: the attested_by subset
			// gets a staleness NUDGE instead of a check — a human's word cannot
			// be re-proved by design (claims.go's EvidenceSource doc comment),
			// so this only turns long silence into a queue for a human, never a
			// pass/fail verdict on the claim itself.
			if src, ok := fact["source"].(map[string]interface{}); ok {
				if _, has := src["attested_by"]; has {
					if entry := checkAttestationStaleness(fact, today); entry != nil {
						res.FactsChecked++
						res.AttestationsDue++
						res.Facts = append(res.Facts, *entry)
					}
				}
			}
			continue
		}

		res.FactsChecked++
		entry := evidenceFactRefresh{
			FactID:    datahelpers.GetStringField(fact, "id", ""),
			Claim:     datahelpers.GetStringField(fact, "claim", ""),
			Tolerance: datahelpers.GetStringField(fact, "tolerance", "exact"),
		}
		if v, ok := numericField(fact["value"]); ok {
			stored := v
			entry.StoredVal = &stored
		}

		live, err := runEvidenceQuery(ctx, db, query)
		if err != nil {
			entry.Outcome = "error"
			entry.Detail = err.Error()
			res.Errors++
			res.Facts = append(res.Facts, entry)
			logger.Warn("refresh_evidence_base: fact query failed",
				zap.String("fact_id", entry.FactID), zap.Error(err))
			continue
		}
		liveVal := live
		entry.LiveVal = &liveVal

		switch {
		case entry.StoredVal == nil:
			entry.Outcome = "updated"
			entry.Detail = "fact had no stored value; recorded the live one"
		case math.Abs(*entry.StoredVal-live) < 1e-9:
			entry.Outcome = "fresh"
			entry.Direction = "unchanged"
		default:
			entry.Direction = "up"
			if live < *entry.StoredVal {
				entry.Direction = "down"
			}
			if evidenceValueWithinTolerance(*entry.StoredVal, live, entry.Tolerance) {
				entry.Outcome = "updated"
				entry.Detail = fmt.Sprintf("live value moved %s within tolerance %s",
					entry.Direction, entry.Tolerance)
			} else {
				entry.Outcome = "drifted"
				res.Drifted++
				entry.Detail = fmt.Sprintf(
					"live value %.0f is outside tolerance %q of the published %.0f (moved %s) — published copy may need a human ruling",
					live, entry.Tolerance, *entry.StoredVal, entry.Direction)
			}
		}

		// The query is the source of truth for a sql-sourced fact, so the
		// register's number tracks it either way. Drift is reported, not
		// suppressed — the work item is what a human acts on.
		if entry.Outcome != "fresh" {
			fact["value"] = live
			fact["verified_at"] = today
			entry.VerifiedAt = today
			res.FactsUpdated++
			changed = true
		} else if datahelpers.GetStringField(fact, "verified_at", "") != today {
			fact["verified_at"] = today
			entry.VerifiedAt = today
			changed = true
		}

		res.Facts = append(res.Facts, entry)
	}

	if res.FactsChecked == 0 {
		return res, nil // no live-verifiable facts — nothing to do, no noise
	}

	// ── Whitelist regeneration (V2's writer_block) ──
	managed := false
	if b, ok := eb["writer_block_managed"].(bool); ok {
		managed = b
	}
	if managed {
		if block := composeWriterBlock(eb); block != "" {
			if existing, _ := eb["writer_block"].(string); existing != block {
				eb["writer_block"] = block
				changed = true
				res.WriterBlock = "regenerated"
			}
		}
	} else {
		res.WriterBlock = "unmanaged"
	}

	// Piece 3 (refresh_evidence_fact_drift.go): plan the per-tool fan-out for
	// any fact a criteria fence on this site declares. Planned BEFORE the dry-run
	// return so a dry run REPORTS it; written after the existing raises below so
	// it can never change when they fire. Read-only here.
	factDrift := planSiteFactDrift(ctx, db, siteID, specRowID, eb, res, dryRun, logger)
	res.FactDrift = factDrift.Emissions

	if dryRun {
		return res, nil // report only — write nothing, raise nothing
	}

	if changed {
		if err := writeRefreshedEvidenceBase(ctx, db, siteID, specRowID, eb, pinned, res, logger); err != nil {
			return nil, err
		}
	}

	// The stale_evidence raise stays gated behind `changed`, EXACTLY as it
	// always was, for the pre-existing sql/citation branch — a council
	// architecture objection (2026-08-12) on an earlier version of this
	// change found that decoupling this raise from `changed` altered
	// runtime behaviour for the existing citation-fact caller (every site,
	// not just new RFC_025 facts), which RFC_025's ratification never
	// covered. So: a citation-only drift with nothing else auto-updating in
	// the register still does NOT raise here — that is unchanged, ungated,
	// pre-existing behaviour, and fixing it (if it should be fixed) is a
	// separate, separately-reviewed change, not bundled into this one.
	//
	// The ONE new case this must still cover: an artifact_check-only fact
	// that drifts with nothing else in the register changing (no live VALUE
	// to re-sync, same shape as a citation drift) — RFC_025 was ratified
	// specifically to let this new mechanism raise a finding, so
	// res.ArtifactCheckDrifted (incremented ONLY for the new mechanism, see
	// its own field comment) opens the gate for that case alone, leaving the
	// old citation/sql behaviour's gate untouched.
	if shouldRaiseStaleEvidence(res, changed) {
		// bugs_open/091 candidate 2. This used to read `else { res.WorkItemCreated
		// = true }`, so work_item_created meant "no error", not "a record exists".
		// The item is keyed per SITE, and insertWorkItem dedups on that key while
		// any non-terminal row is open — so when a SECOND, different fact drifts
		// while an earlier item is still open, nothing is written and the run
		// reported the write anyway. The only trace of the truth was a log line
		// carrying inserted=false, in a pod that was replaced four minutes later.
		//
		// Candidate 1 (2026-08-02) then closed the door the report had been left
		// pointing at: the write now uses refreshOnConflict, so the open item's
		// description is brought up to date instead of the finding being dropped.
		// The two reported fields stay distinct — a refresh is not a creation.
		write, err := createStaleEvidenceItem(ctx, db, siteID, domain, res, params.AgentType, logger)
		if err != nil {
			logger.Warn("refresh_evidence_base: failed to create stale_evidence item", zap.Error(err))
		}
		res.WorkItemCreated = write.Inserted
		res.WorkItemRefreshed = write.Refreshed
		if err == nil && !write.Recorded() {
			logger.Warn("refresh_evidence_base: drift found but NO work item written or refreshed — "+
				"an open stale_evidence item already holds this site's key and could not be updated "+
				"(it went terminal, or a handler holds it), so its spec still describes the EARLIER "+
				"drift (bugs_open/091)",
				zap.String("site_id", siteID.String()),
				zap.String("domain", domain),
				zap.Int("drifted", res.Drifted))
		}
	}

	if res.AttestationsDue > 0 {
		// RFC_025 stage 1. Modelled on the stale_evidence raise directly above,
		// but a SEPARATE item_type (stale_attestation): this is a cadence nudge,
		// not a detected mismatch, and folding it into stale_evidence would tell
		// a human "drift was found" about a fact nobody has re-checked at all.
		write, err := createStaleAttestationItem(ctx, db, siteID, domain, res, params.AgentType, logger)
		if err != nil {
			logger.Warn("refresh_evidence_base: failed to create stale_attestation item", zap.Error(err))
		}
		res.AttestationWorkItemCreated = write.Inserted
		res.AttestationWorkItemRefreshed = write.Refreshed
	}

	// Piece 3's write half — its own items, its own keys, after (and
	// independent of) the two raises above. See refresh_evidence_fact_drift.go.
	if len(factDrift.Emissions) > 0 {
		writeFactDriftItems(ctx, db, siteID, domain, &factDrift, params.AgentType, logger)
		res.FactDrift = factDrift.Emissions
	}

	return res, nil
}

// refreshCitationFact re-verifies one citation fact in place and returns its
// refresh entry. Outcomes and what they mean:
//
//	fresh    — quote still present at the source; accessed/verified_at bumped.
//	drifted  — the source no longer supports the claim (quote gone on a 200),
//	           OR the citation has aged past its staleness_days policy. Either
//	           way the published claim needs a human ruling.
//	error    — the source could not be checked (network, 403, PDF). Unknown is
//	           not loss: reported, never treated as drift.
//
// Facts marked "reverifiable": false are never fetched — they age by policy
// only, and re-attesting them is a human act.
func refreshCitationFact(ctx context.Context, fact map[string]interface{}, today string) *evidenceFactRefresh {
	entry := &evidenceFactRefresh{
		FactID:    datahelpers.GetStringField(fact, "id", ""),
		Claim:     datahelpers.GetStringField(fact, "claim", ""),
		Tolerance: "citation",
	}
	src, _ := fact["source"].(map[string]interface{})
	cit, err := datahelpers.ParseCitation(src)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = err.Error()
		return entry
	}

	stalenessDays := 0.0
	if v, ok := numericField(fact["staleness_days"]); ok {
		stalenessDays = v
	}
	verifiedAt := datahelpers.GetStringField(fact, "verified_at", "")
	now, _ := parseFlexibleDate(today)

	if datahelpers.GetBoolField(fact, "reverifiable", true) == false {
		if citationDateStale(cit.Published, verifiedAt, stalenessDays, now) {
			entry.Outcome = "drifted"
			entry.Detail = "citation is past its staleness_days policy and is marked reverifiable:false — re-attest it by hand or retire the claim"
		} else {
			entry.Outcome = "fresh"
			entry.Detail = "reverifiable:false — aged by policy only, not re-fetched"
		}
		return entry
	}

	outcome := verifyCitationLive(ctx, cit)
	switch {
	case outcome.FailClass == "fetch_error":
		entry.Outcome = "error"
		entry.Detail = outcome.FailDetail
	case !outcome.Found:
		entry.Outcome = "drifted"
		entry.Detail = "citation_lost: " + outcome.FailDetail
	case citationDateStale(cit.Published, verifiedAt, stalenessDays, now):
		entry.Outcome = "drifted"
		entry.Detail = "quote still present, but the citation is past its staleness_days policy — the source itself has aged; re-research the figure"
	default:
		entry.Outcome = "fresh"
		fact["verified_at"] = today
		if citRaw, ok := src["citation"].(map[string]interface{}); ok {
			citRaw["accessed"] = today
		}
		entry.VerifiedAt = today
	}
	return entry
}

// ============================================================================
// RFC_025 stage 2 — artifact_check
// ============================================================================
//
// docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_025_artifact_sourced_facts_are_trusted_once_registered.md
// (RATIFIED 2026-08-12; §9 Q2 is the constraint this file follows: the key is
// read off the untyped `source` map, never added as a field on the exported
// EvidenceSource/EvidenceFact structs in datahelpers/claims.go).
//
// The motivating case (bugs_open/161): gamesdesign.co.uk's `gd-trials` fact
// cited "the figure is hard-coded in the shipped drop-rate tool JavaScript" for
// a claim the artefact did not support — an `artifact` source is, today,
// checked for presence in the register and never re-proved against the thing
// it cites. artifact_check closes that for the facts whose author opts in.
//
// Two safety properties added 2026-08-12 after council review, both closing
// gaps that would have reproduced RFC_025's own motivating failure INSIDE its
// fix: (1) parseArtifactCheck refuses a bare-numeric pattern (no surrounding
// regex context at all) — the platform's own documented landmine, "grepping a
// tool for 10000 matches 100000", is exactly a false PASS this mechanism could
// otherwise produce; (2) refreshArtifactCheckFact resolves component_id
// scoped to the fact's OWN site (joined through pages), refusing rather than
// silently matching a component belonging to a different site.

// artifactCheckSpec is source.artifact_check, parsed. Unexported and local to
// this file: unlike Citation (datahelpers/citations.go), nothing outside
// refreshOneSiteEvidence needs to know this shape today, so it stays out of
// datahelpers rather than growing a second exported symbol for a shape with
// one reader.
type artifactCheckSpec struct {
	ComponentID   string
	Pattern       string
	MustBePresent bool
}

// bareNumericPattern matches an artifact_check.pattern that is NOTHING but
// digits — no surrounding regex structure, no word/context anchoring at all.
// Refused by parseArtifactCheck: this is the exact shape of the platform's own
// documented landmine ("grepping a tool for 10000 matches 100000", named in
// bugs_open/161's own "Traps this cost me" section — the motivating bug for
// this whole mechanism). A pattern this bare would let `10000` "verify" a fact
// against a component whose only 10000-shaped substring is `100000`, silently
// reporting fresh when the fact has actually drifted — reproducing, inside the
// fix, the precise failure mode RFC_025 exists to close. Any pattern with
// SOME surrounding structure (a word boundary, punctuation, surrounding code
// context, as RFC_025's own worked example `Math\.min\(val,\s*10000\)` uses)
// clears this — the guard only refuses the maximally naive case.
var bareNumericPattern = regexp.MustCompile(`^[0-9]+$`)

// parseArtifactCheck reads source.artifact_check off the untyped map. Returns
// an error for anything that would make the check unrunnable — a missing
// component_id or pattern, or a pattern too bare to trust — so the caller can
// fail CLOSED (RFC_017) rather than silently skip or silently mismatch.
func parseArtifactCheck(src map[string]interface{}) (*artifactCheckSpec, error) {
	raw, ok := src["artifact_check"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("artifact_check is present but not an object")
	}
	spec := &artifactCheckSpec{
		ComponentID:   strings.TrimSpace(datahelpers.GetStringField(raw, "component_id", "")),
		Pattern:       datahelpers.GetStringField(raw, "pattern", ""),
		MustBePresent: datahelpers.GetBoolField(raw, "must_be_present", true),
	}
	var missing []string
	if spec.ComponentID == "" {
		missing = append(missing, "component_id")
	}
	if spec.Pattern == "" {
		missing = append(missing, "pattern")
	}
	if len(missing) > 0 {
		return spec, fmt.Errorf("artifact_check missing required field(s): %s", strings.Join(missing, ", "))
	}
	if bareNumericPattern.MatchString(spec.Pattern) {
		return spec, fmt.Errorf(
			"artifact_check.pattern %q is bare digits with no surrounding context — this can substring-match a "+
				"larger number (e.g. %q would match inside %q) and silently report a drifted fact as fresh; "+
				"add surrounding context, e.g. word boundaries (\\b%s\\b) or the actual code shape "+
				"(as in \"Math\\.min\\(val,\\s*%s\\)\")", spec.Pattern, spec.Pattern, spec.Pattern+"0", spec.Pattern, spec.Pattern)
	}
	return spec, nil
}

// refreshArtifactCheckFact re-proves one artifact_check fact against its named
// stored artefact — today, a page_components row's rendered_html, addressed by
// component_id. Outcomes:
//
//	fresh    — the pattern's presence matches must_be_present; verified_at bumped.
//	drifted  — it does not: the artefact changed under a published claim about it,
//	           or (must_be_present:false) the thing the claim says is ABSENT is
//	           now present. Either way the published claim needs a human ruling —
//	           the same shape as a citation losing its quote.
//	error    — the check could not run at all: component_id does not resolve, the
//	           pattern does not compile, or the read failed. RFC_017: a check that
//	           cannot run must never be reported as a pass.
//
// db is passed explicitly (unlike refreshCitationFact, which only touches the
// network) because the artefact this stage checks lives in Postgres, not on
// the open web. siteID scopes the component_id lookup to the fact's OWN site
// (council objection, 2026-08-12, raised independently by four reviewers:
// page_components has no site column of its own, so an unscoped `WHERE id =
// $1` would let a fact "verify" against another site's component — a false
// PASS on the exact re-verification machinery this RFC exists to strengthen).
// Joined through pages, the one table that actually carries site_id.
func refreshArtifactCheckFact(ctx context.Context, db *sql.DB, siteID uuid.UUID, fact map[string]interface{}, today string) *evidenceFactRefresh {
	entry := &evidenceFactRefresh{
		FactID:    datahelpers.GetStringField(fact, "id", ""),
		Claim:     datahelpers.GetStringField(fact, "claim", ""),
		Tolerance: "artifact_check",
	}
	src, _ := fact["source"].(map[string]interface{})
	spec, err := parseArtifactCheck(src)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = err.Error()
		return entry
	}

	componentID, err := uuid.Parse(spec.ComponentID)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = fmt.Sprintf("artifact_check.component_id %q is not a valid id: %v", spec.ComponentID, err)
		return entry
	}

	re, err := regexp.Compile(spec.Pattern)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = fmt.Sprintf("artifact_check.pattern %q does not compile: %v", spec.Pattern, err)
		return entry
	}

	var rendered sql.NullString
	err = db.QueryRowContext(ctx, `
		SELECT pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.id = $1 AND p.site_id = $2
	`, componentID, siteID).Scan(&rendered)
	switch {
	case err == sql.ErrNoRows:
		entry.Outcome = "error"
		entry.Detail = fmt.Sprintf(
			"artifact_check.component_id %s does not resolve to a page_components row on THIS site (%s) — "+
				"either it does not exist, or it belongs to a different site and is refused rather than trusted",
			componentID, siteID)
		return entry
	case err != nil:
		entry.Outcome = "error"
		entry.Detail = fmt.Sprintf("artifact_check: reading component %s failed: %v", componentID, err)
		return entry
	}

	found := re.MatchString(rendered.String)
	switch {
	case found == spec.MustBePresent:
		entry.Outcome = "fresh"
		fact["verified_at"] = today
		entry.VerifiedAt = today
	case spec.MustBePresent:
		entry.Outcome = "drifted"
		entry.Detail = fmt.Sprintf(
			"artifact_check: pattern %q no longer found in component %s — the artefact this fact cites may have changed; the published claim needs a human ruling",
			spec.Pattern, componentID)
	default:
		entry.Outcome = "drifted"
		entry.Detail = fmt.Sprintf(
			"artifact_check: pattern %q is now PRESENT in component %s, but the fact asserts it must be absent — the published claim needs a human ruling",
			spec.Pattern, componentID)
	}
	return entry
}

// ============================================================================
// RFC_025 stage 1 — attestation staleness nudge
// ============================================================================
//
// attested_by facts are a human's word, by design never machine-checkable
// (datahelpers/claims.go's EvidenceSource doc comment). This does not check
// anything — it turns long silence into a queue for a human, the same shape
// stale_evidence already uses for sql facts and staleness_days already uses
// for citations, just with no re-proof step at the end of it.

// attestationStaleDays is the cadence RFC_025 §2.1 names (~180 days). It
// governs the THRESHOLD, not the poll interval — this runs on the same daily
// evidence-freshness sweep as everything else in this file, and only actually
// flags a fact once it has gone this long without a human re-dating it. That
// mirrors how citationDateStale's staleness_days already works inside the same
// daily pass; no second scheduled_task is needed for a threshold this file
// already knows how to apply on a cadence it already runs.
const attestationStaleDays = 180

// checkAttestationStaleness reports whether an attested_by fact is due for a
// human to re-look at it, anchored on the fact's own verified_at — the last
// time a human dated it. A fact with no usable verified_at is treated as due
// immediately: an undated attestation is not evidence of freshness, it is the
// absence of the one signal this check has.
func checkAttestationStaleness(fact map[string]interface{}, today string) *evidenceFactRefresh {
	now, ok := parseFlexibleDate(today)
	if !ok {
		return nil // no usable clock this pass — nothing to age from
	}
	verifiedAt := datahelpers.GetStringField(fact, "verified_at", "")
	anchor, ok := parseFlexibleDate(verifiedAt)
	due := !ok || now.Sub(anchor) > attestationStaleDays*24*time.Hour
	if !due {
		return nil
	}

	attestedBy := ""
	if src, ok := fact["source"].(map[string]interface{}); ok {
		attestedBy = datahelpers.GetStringField(src, "attested_by", "")
	}
	whenPhrase := "was never dated"
	if verifiedAt != "" {
		whenPhrase = "was last dated " + verifiedAt
	}
	return &evidenceFactRefresh{
		FactID:     datahelpers.GetStringField(fact, "id", ""),
		Claim:      datahelpers.GetStringField(fact, "claim", ""),
		Tolerance:  "attestation",
		Outcome:    "attestation_due",
		VerifiedAt: "", // deliberately NOT bumped — a nudge is not a re-attestation
		Detail: fmt.Sprintf(
			"attested_by fact (%q) %s — more than %d days is the cadence this platform nudges on; a human should re-look and re-date it, or retire the claim",
			attestedBy, whenPhrase, attestationStaleDays),
	}
}

// factSQLSource returns the fact's SQL source, if it has one.
func factSQLSource(fact map[string]interface{}) string {
	src, ok := fact["source"].(map[string]interface{})
	if !ok {
		return ""
	}
	return strings.TrimSpace(datahelpers.GetStringField(src, "sql", ""))
}

func numericField(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

var selectOnlyRe = regexp.MustCompile(`(?is)^\s*select\s`)

// forbiddenVerbRe catches write verbs anywhere in the statement, so a nested
// or trailing write cannot ride along inside an otherwise-SELECT query.
var forbiddenVerbRe = regexp.MustCompile(`(?is)\b(insert|update|delete|drop|alter|truncate|grant|revoke|create|copy|vacuum|call|do|merge)\b`)

// validateEvidenceQuery applies the safety rules in the file header and
// returns the statement to execute. Pure and side-effect free, so the guard is
// testable on its own — the part that must never be got wrong is the part that
// never touches a database.
func validateEvidenceQuery(query string) (string, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(query), ";")
	if strings.Contains(trimmed, ";") {
		return "", fmt.Errorf("refused: query contains multiple statements")
	}
	if !selectOnlyRe.MatchString(trimmed) {
		return "", fmt.Errorf("refused: query must begin with SELECT")
	}
	if forbiddenVerbRe.MatchString(trimmed) {
		return "", fmt.Errorf("refused: query contains a data-modifying keyword")
	}
	return trimmed, nil
}

// runEvidenceQuery validates one evidence query, executes it read-only, and
// returns its single numeric scalar.
func runEvidenceQuery(ctx context.Context, db *sql.DB, query string) (float64, error) {
	trimmed, err := validateEvidenceQuery(query)
	if err != nil {
		return 0, err
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, fmt.Errorf("begin read-only tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SET LOCAL statement_timeout = '15s'`); err != nil {
		return 0, fmt.Errorf("set statement timeout: %w", err)
	}

	var raw interface{}
	if err := tx.QueryRowContext(ctx, trimmed).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("query returned no rows")
		}
		return 0, fmt.Errorf("query failed: %w", err)
	}

	switch v := raw.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case []byte:
		var f float64
		if _, err := fmt.Sscanf(string(v), "%g", &f); err != nil {
			return 0, fmt.Errorf("non-numeric result %q", string(v))
		}
		return f, nil
	case string:
		var f float64
		if _, err := fmt.Sscanf(v, "%g", &f); err != nil {
			return 0, fmt.Errorf("non-numeric result %q", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported result type %T", raw)
	}
}

// evidenceValueWithinTolerance mirrors the scan-side tolerance semantics in
// datahelpers/claims.go, from the register's point of view: does the live value
// still sit within what the published claim tolerates?
//
//   - exact: any movement is drift (the copy names a specific number)
//   - gte:   the copy says "more than N"; growth is fine, a FALL below the
//     published figure is drift (the copy would be overclaiming)
//   - approx_pct:N: movement beyond N percent is drift
func evidenceValueWithinTolerance(stored, live float64, tolerance string) bool {
	switch {
	case tolerance == "gte":
		return live >= stored
	case strings.HasPrefix(tolerance, "approx_pct:"):
		var pct float64
		if _, err := fmt.Sscanf(strings.TrimPrefix(tolerance, "approx_pct:"), "%g", &pct); err != nil {
			return false
		}
		if stored == 0 {
			return live == 0
		}
		return math.Abs(live-stored)/math.Abs(stored)*100 <= pct
	default: // exact
		return math.Abs(live-stored) < 1e-9
	}
}

// composeWriterBlock rebuilds the V2 writer whitelist from the register.
// Structure (the three headers) is the machine's; every claim sentence is the
// human's `writer_line`, with `{value}` replaced by the current number. A fact
// without a writer_line is omitted — never auto-phrased.
func composeWriterBlock(eb map[string]interface{}) string {
	factsRaw, _ := eb["facts"].([]interface{})
	var numbers, capabilities []string

	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		line := strings.TrimSpace(datahelpers.GetStringField(fact, "writer_line", ""))
		if line == "" {
			continue
		}
		if v, ok := numericField(fact["value"]); ok {
			line = strings.ReplaceAll(line, "{value}", formatEvidenceNumber(v))
			numbers = append(numbers, "- "+line)
		} else {
			capabilities = append(capabilities, "- "+line)
		}
	}

	if len(numbers) == 0 && len(capabilities) == 0 {
		return "" // nothing phrased — leave the existing block alone
	}

	var b strings.Builder
	if len(numbers) > 0 {
		b.WriteString("NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):\n")
		b.WriteString(strings.Join(numbers, "\n"))
	}
	if len(capabilities) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("CAPABILITIES (assert without inventing numbers):\n")
		b.WriteString(strings.Join(capabilities, "\n"))
	}
	if ents, ok := eb["allowed_entities"].([]interface{}); ok && len(ents) > 0 {
		names := make([]string, 0, len(ents))
		for _, e := range ents {
			if s, ok := e.(string); ok && s != "" {
				names = append(names, s)
			}
		}
		if len(names) > 0 {
			b.WriteString("\n\nNAMED ENTITIES you may assert relationships with: ")
			b.WriteString(strings.Join(names, ", "))
			b.WriteString(".")
		}
	}
	return b.String()
}

// formatEvidenceNumber renders a fact value the way published copy states it:
// whole numbers with thousands separators, fractions left alone.
func formatEvidenceNumber(v float64) string {
	if v != math.Trunc(v) {
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", v), "0"), ".")
	}
	s := fmt.Sprintf("%.0f", math.Abs(v))
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if v < 0 {
		return "-" + string(out)
	}
	return string(out)
}

// writeRefreshedEvidenceBase supersedes the current spec row and inserts the
// refreshed one, preserving `pinned` — same shape as write_site_spec's
// versioning, kept local so no other spec semantics (deep merge, formatting)
// apply to an evidence base.
//
// COMPARE-AND-SWAP, not blind overwrite. This action rewrites the whole `data`
// column from a copy read moments earlier, so a human editing the evidence base
// in between would otherwise have their edit silently dropped — the classic
// lost update, and the evidence base is exactly the kind of human-owned row
// where losing an edit is unacceptable. The supersede is therefore keyed on the
// row id we read; if it no longer matches (someone wrote first), we affect zero
// rows and abort WITHOUT writing. The refresh is safely re-runnable, so losing
// a pass costs one scheduled tick; losing a human's edit costs trust.
func writeRefreshedEvidenceBase(
	ctx context.Context, db *sql.DB, siteID, specRowID uuid.UUID,
	eb map[string]interface{}, pinned sql.NullBool,
	res *siteRefreshResult, logger *zap.Logger,
) error {
	newJSON, err := json.Marshal(eb)
	if err != nil {
		return fmt.Errorf("marshal refreshed evidence_base: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	sres, err := tx.ExecContext(ctx, `
		UPDATE site_specs SET is_current = false, superseded_at = now()
		WHERE id = $1 AND is_current = true
	`, specRowID)
	if err != nil {
		return fmt.Errorf("supersede evidence_base: %w", err)
	}
	affected, err := sres.RowsAffected()
	if err != nil {
		return fmt.Errorf("supersede rows affected: %w", err)
	}
	if affected != 1 {
		logger.Warn("refresh_evidence_base: evidence base changed under us — skipping write to avoid clobbering a concurrent edit",
			zap.String("site_id", siteID.String()),
			zap.String("read_row_id", specRowID.String()))
		res.WriterBlock = "skipped_concurrent_edit"
		return nil
	}

	notes := fmt.Sprintf(
		"V4 freshness pass: %d live-verifiable fact(s) checked, %d updated, %d drifted, whitelist %s",
		res.FactsChecked, res.FactsUpdated, res.Drifted, res.WriterBlock)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO site_specs
		    (site_id, aspect, data, source, notes, is_current, created_by, pinned)
		VALUES ($1, 'evidence_base', $2::jsonb, 'scheduled', $3, true, 'evidence-refresher', $4)
	`, siteID, string(newJSON), notes, pinned.Valid && pinned.Bool); err != nil {
		return fmt.Errorf("insert refreshed evidence_base: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit refreshed evidence_base: %w", err)
	}

	logger.Info("refresh_evidence_base: evidence base updated",
		zap.String("site_id", siteID.String()),
		zap.Int("facts_updated", res.FactsUpdated),
		zap.Int("drifted", res.Drifted),
		zap.String("writer_block", res.WriterBlock))
	return nil
}

// createStaleEvidenceItem raises drift for human ruling. HITL-terminal: the
// handler is the human-review pseudo-handler, matching claims_unverified —
// changing published copy because a number moved is a human decision (spec
// open question 3), even when the site is merely under-claiming.
// createStaleEvidenceItem returns what was actually WRITTEN, not merely whether
// the attempt errored (bugs_open/091). The item is keyed per site while the
// finding is per fact, so a second, DIFFERENT drift arriving while an earlier
// item is open used to dedup to nothing — the finding lost and the open record
// left describing the earlier drift. It now refreshes that record instead, and
// the outcome distinguishes the three cases (created / refreshed / neither),
// because a field named `work_item_created` must not be set from a refresh.
func createStaleEvidenceItem(
	ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string,
	res *siteRefreshResult, agentType string, logger *zap.Logger,
) (workItemWrite, error) {
	drifted := make([]evidenceFactRefresh, 0, res.Drifted)
	for _, f := range res.Facts {
		if f.Outcome == "drifted" {
			drifted = append(drifted, f)
		}
	}

	specJSON, err := json.Marshal(map[string]interface{}{
		"check":   "evidence_freshness",
		"domain":  domain,
		"drifted": drifted,
		"fix": "A live-verifiable fact has moved outside the tolerance its published wording allows. " +
			"Review the pages that state it: either update the copy to the new figure, re-word the claim " +
			"to a tolerance that survives movement (e.g. 'more than N'), or accept it. The register's " +
			"number has already been re-synced to the live query — the COPY is what needs a human ruling.",
	})
	if err != nil {
		return workItemWrite{}, fmt.Errorf("marshal stale_evidence spec: %w", err)
	}

	summary := fmt.Sprintf("Evidence freshness (%s): %d fact(s) drifted outside tolerance", domain, len(drifted))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return workItemWrite{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// refreshOnConflict, because this item is keyed per SITE and the finding is
	// per FACT. A second, different fact drifting while the first item is open is
	// not a duplicate — under the default policy it was dropped, and the durable
	// record went on describing drift that had since been re-synced. Measured
	// 2026-08-02: four of five open stale_evidence items named the wrong facts,
	// including one that named a completely different fact from the one that had
	// moved, and one describing drift that no longer existed at all.
	write, err := writeWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "scheduled",
		pipeline:     "content",
		itemType:     "stale_evidence",
		severity:     "medium",
		summary:      summary,
		spec:         string(specJSON),
		priority:     35,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    agentType,
		itemKey:      "stale_evidence:" + siteID.String(),
	}, refreshOnConflict, logger)
	if err != nil {
		return write, fmt.Errorf("insert stale_evidence item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return workItemWrite{}, fmt.Errorf("commit stale_evidence item: %w", err)
	}

	logger.Warn("refresh_evidence_base: evidence drift raised for human review",
		zap.String("site_id", siteID.String()),
		zap.Int("drifted", len(drifted)),
		zap.Bool("inserted", write.Inserted),
		zap.Bool("refreshed", write.Refreshed))
	return write, nil
}

// createStaleAttestationItem raises RFC_025 stage 1's staleness nudge for
// human ruling. Modelled directly on createStaleEvidenceItem immediately
// above — same keyed-per-site/refreshOnConflict shape, same honest
// created/refreshed reporting — but a DIFFERENT item_type (stale_attestation)
// and a different fix message, because this is not a detected defect: nothing
// has been found WRONG, a human's word has simply gone unrefreshed past the
// platform's cadence.
func createStaleAttestationItem(
	ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string,
	res *siteRefreshResult, agentType string, logger *zap.Logger,
) (workItemWrite, error) {
	due := make([]evidenceFactRefresh, 0, res.AttestationsDue)
	for _, f := range res.Facts {
		if f.Outcome == "attestation_due" {
			due = append(due, f)
		}
	}

	specJSON, err := json.Marshal(map[string]interface{}{
		"check":  "attestation_freshness",
		"domain": domain,
		"due":    due,
		"fix": fmt.Sprintf(
			"Each listed fact is sourced from a human's word (source.attested_by) — nothing can "+
				"re-prove it, by design (see EvidenceSource's doc comment in datahelpers/claims.go). "+
				"It has gone more than %d days since it was last dated. Re-look at the claim: confirm it "+
				"still holds and bump verified_at, reword it, or retire it. This is a NUDGE, not a "+
				"detected defect — no mismatch was found, because none can be, machine-side.",
			attestationStaleDays),
	})
	if err != nil {
		return workItemWrite{}, fmt.Errorf("marshal stale_attestation spec: %w", err)
	}

	summary := fmt.Sprintf("Attestation staleness (%s): %d attested fact(s) due for human re-look", domain, len(due))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return workItemWrite{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// refreshOnConflict for the same reason as stale_evidence above: keyed per
	// SITE, and a second pass finding a DIFFERENT (or additional) fact due must
	// bring the open item's list up to date rather than be silently dropped
	// while an earlier item sits open.
	write, err := writeWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "scheduled",
		pipeline:     "content",
		itemType:     "stale_attestation",
		severity:     "low",
		summary:      summary,
		spec:         string(specJSON),
		priority:     60,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    agentType,
		itemKey:      "stale_attestation:" + siteID.String(),
	}, refreshOnConflict, logger)
	if err != nil {
		return write, fmt.Errorf("insert stale_attestation item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return workItemWrite{}, fmt.Errorf("commit stale_attestation item: %w", err)
	}

	logger.Info("refresh_evidence_base: attestation staleness raised for human review",
		zap.String("site_id", siteID.String()),
		zap.Int("due", len(due)),
		zap.Bool("inserted", write.Inserted),
		zap.Bool("refreshed", write.Refreshed))
	return write, nil
}

func loadSiteDomain(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	var domain sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		logger.Warn("refresh_evidence_base: domain lookup failed", zap.Error(err))
		return ""
	}
	return domain.String
}

// currentDateString reads the date from the database so `verified_at` matches
// the clock that produced the values.
func currentDateString(ctx context.Context, db *sql.DB) string {
	var today string
	if err := db.QueryRowContext(ctx, `SELECT to_char(now(), 'YYYY-MM-DD')`).Scan(&today); err != nil {
		return ""
	}
	return today
}
