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
	"hash/fnv"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
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

	// FactsHistoryRecorded counts the superseded readings retained this pass
	// (bugs_open/386). Reported separately from FactsUpdated because they answer
	// different questions and can legitimately disagree: every armed fact whose
	// value moves increments both, while an UNARMED fact moving increments only
	// FactsUpdated. So a run showing updates and zero history is the honest signal
	// that nothing is armed yet — not a silent failure of the retention.
	FactsHistoryRecorded int `json:"facts_history_recorded"`

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

	// MisplacedArtifactChecks names facts carrying artifact_check at the TOP
	// LEVEL of the fact instead of inside `source`, where the code reads it. Such
	// a check is INERT and was, until 2026-08-25, silently so. omitempty: a
	// register with none marshals exactly as before.
	MisplacedArtifactChecks []string `json:"misplaced_artifact_checks,omitempty"`

	// FactBindingSuggestions is Phase 4's count of NON-declaring tools on this
	// site whose script text carries one or more of the site's own registered
	// values — an adoption signal, not a defect count. omitempty, so a site with
	// nothing to suggest marshals exactly as before.
	FactBindingSuggestions int `json:"fact_binding_suggestions,omitempty"`

	// EvidenceConsumerPagesQueued (bugs_open/427) counts the page_rerender
	// items queued because this write changed the register a
	// `query.upcoming_events`-consuming component declared — the propagation
	// half of "dates get corrected, not just added": a human ruling on a
	// stale_evidence item edits the register by hand, and that edit alone
	// does not re-render anything without this. omitempty: a site with no
	// such consumer (every site today) marshals exactly as before.
	EvidenceConsumerPagesQueued int `json:"evidence_consumer_pages_queued,omitempty"`

	// InvalidBannedClaimPatterns (RFC_060 §1e/§3e) names every per-site
	// banned_claims pattern that fails to compile as a regex this pass — a
	// silent no-op guard (claims.go:348's fallback) with nothing else in the
	// estate positioned to notice it. omitempty: a clean register (the fleet
	// today) marshals exactly as before.
	InvalidBannedClaimPatterns []invalidBannedClaimPattern `json:"invalid_banned_claim_patterns,omitempty"`
	// InvalidBannedClaimWorkItemsCreated is how many of the above got a NEW
	// work item this pass (a pattern already holding an open item counts
	// toward InvalidBannedClaimPatterns but not here — see
	// createInvalidBannedClaimPatternItems).
	InvalidBannedClaimWorkItemsCreated int `json:"invalid_banned_claim_work_items_created,omitempty"`
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

// invalidBannedClaimPattern names one per-site banned_claims pattern that
// fails to compile as a regex.
//
// RFC_060 §1e/§3e (2026-09-02): claims.go:348 falls back SILENTLY when a
// per-site pattern does not compile — no logger, no error path — degrading
// the guard to a literal match of its own source text rather than refusing
// the write or telling anyone. TestEveryGlobalPatternIsAValidRegex pins the
// FLEET-WIDE set (authored in Go); it cannot see a pattern arriving as DATA.
// The admin door counts patterns without compiling one, so the guard is
// armed, listed, counted, and INERT, with every count-based check passing.
type invalidBannedClaimPattern struct {
	Index   int    `json:"index"`
	Pattern string `json:"pattern"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"compile_error"`
}

// checkBannedClaimPatterns re-runs EXACTLY the compile claims.go:348 performs
// (case-insensitive, same prefix) over a site's raw banned_claims and reports
// every pattern that does not compile — pure, no DB, so the finding is
// unit-testable without a fixture site.
func checkBannedClaimPatterns(bannedClaimsRaw interface{}) []invalidBannedClaimPattern {
	list, ok := bannedClaimsRaw.([]interface{})
	if !ok {
		return nil
	}
	var invalid []invalidBannedClaimPattern
	for i, entry := range list {
		bc, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		pattern := datahelpers.GetStringField(bc, "pattern", "")
		if pattern == "" {
			continue
		}
		if _, err := regexp.Compile("(?i)" + pattern); err != nil {
			invalid = append(invalid, invalidBannedClaimPattern{
				Index:   i,
				Pattern: pattern,
				Reason:  datahelpers.GetStringField(bc, "reason", ""),
				Error:   err.Error(),
			})
		}
	}
	return invalid
}

// bannedClaimPatternItemKey identifies ONE bad pattern, not the site — the
// council-argued fix to the sibling stale_evidence shape (bugs_open/091): a
// key scoped to the site alone would let a SECOND, different bad pattern hide
// behind an already-open item for the first (measured there: four of five
// open items named the wrong fact). fnv64a of the pattern text, mirroring
// citationFactID's existing convention in this package.
func bannedClaimPatternItemKey(siteID uuid.UUID, pattern string) string {
	h := fnv.New64a()
	h.Write([]byte(pattern))
	return fmt.Sprintf("invalid_banned_claim_pattern:%s:%x", siteID.String(), h.Sum64())
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

	// RFC_060 §1e/§3e: computed here (pure, no DB) so a dry run REPORTS it —
	// the write, gated below with everything else, is what a dry run must not
	// do. Independent of the facts loop below; banned_claims carries no
	// dependency on fact state.
	res.InvalidBannedClaimPatterns = checkBannedClaimPatterns(eb["banned_claims"])

	factsRaw, _ := eb["facts"].([]interface{})
	today := currentDateString(ctx, db)

	changed := false
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		// ⚠ A MISPLACED artifact_check IS SILENTLY INERT, AND THAT IS THE SAME
		// DEFECT AS THE ONE PHASE 1 CLOSED FOR CRITERIA FENCES, ONE TABLE OVER.
		//
		// RFC_025 §2.2 puts artifact_check INSIDE `source`, as a sibling of
		// citation/artifact/attested_by — i.e. as a verification mechanism for a
		// source kind. A perfectly reasonable author instead puts it at the TOP
		// LEVEL of the fact, because it describes the FACT rather than the source;
		// arguably the better data model, and the one that is not implemented.
		// Found 2026-08-25 on agritec.uk: four fences, correct contents, correct
		// patterns, correct subject_key, live in the register — and read by
		// nothing, with no signal anywhere that they were inert. The author had
		// tested the patterns and reported the fence "live".
		//
		// So: say so. This does not guess and does not relocate the key — the
		// register is human-owned and a consumer must not rewrite a human's row
		// (CLM-001). It reports, which is all a reader needs to fix it in seconds.
		if _, misplaced := fact["artifact_check"]; misplaced {
			if src, ok := fact["source"].(map[string]interface{}); !ok || src["artifact_check"] == nil {
				res.MisplacedArtifactChecks = append(res.MisplacedArtifactChecks,
					datahelpers.GetStringField(fact, "id", "(unnamed fact)"))
				logger.Warn("refresh_evidence_base: a fact carries artifact_check at the TOP LEVEL, where nothing reads it — "+
					"RFC_025 places it INSIDE source, beside citation/artifact/attested_by. The check is INERT until it moves.",
					zap.String("site_id", siteID.String()),
					zap.String("fact_id", datahelpers.GetStringField(fact, "id", "")))
			}
		}

		// RFC_025 STAGE 2b (bugs_open/288 §5.6) — artifact_check is now reachable
		// for EVERY source kind, not only for a fact whose sole source is `artifact`.
		//
		// THE DEFECT THIS CLOSES: the citation arm below `continue`s before the
		// artifact_check test seventeen lines further down, and every fact holding a
		// legislated figure IS a citation fact (185 of 294 current facts, measured
		// 2026-08-24). So an artifact_check written beside an SDLT band was never
		// evaluated — the one reader in the estate on the raw-bytes surface could not
		// be pointed at the class of fact it was needed for.
		//
		// A PRE-PASS, deliberately, rather than a restructure of the arms below: the
		// existing branches and their `continue` are untouched, so every path taken
		// today is taken identically. It fires only for a fact carrying
		// artifact_check AND another source — a combination that exists ZERO times on
		// the live fleet (0 of 294 facts carry artifact_check at all, against 185
		// carrying citation as the control), so the blast radius today is provably
		// nil, and the no-op test pins it.
		if src, ok := fact["source"].(map[string]interface{}); ok {
			if _, hasAC := src["artifact_check"]; hasAC && factHasNonArtifactSource(fact, src) {
				if entry := refreshArtifactCheckFact(ctx, db, siteID, fact, today, false); entry != nil {
					res.FactsChecked++
					switch entry.Outcome {
					case "drifted":
						res.Drifted++
						res.ArtifactCheckDrifted++
					case "error":
						res.Errors++
					}
					// Deliberately NOT setting `changed` — see ownsVerifiedAt.
					res.Facts = append(res.Facts, *entry)
				}
			}
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
			// `&& !factHasNonArtifactSource` is load-bearing, not defensive. Without
			// it a fact carrying attested_by (or source.sql) PLUS an artifact_check
			// is checked TWICE — once by the stage-2b pre-pass above as a secondary,
			// then again here as a primary — appending two entries under one FactID
			// and bumping verified_at from the second. The citation arm is safe only
			// because its `continue` happens to intervene. Found by asking what an
			// end-to-end test of the loop would do, which is also how the pre-pass
			// itself came to be pinned.
			if _, has := src["artifact_check"]; has && !factHasNonArtifactSource(fact, src) {
				entry := refreshArtifactCheckFact(ctx, db, siteID, fact, today, true)
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
			// Record the OUTGOING reading before it is lost. Until this existed,
			// every refresh discarded the value the register had just held, which
			// is what turned each already-deployed page rendering it into an
			// "unregistered claim" overnight (bugs_open/386). Ordered before the
			// overwrite because after it the old reading is unrecoverable from
			// this map — the superseded site_specs row keeps it, but nothing in
			// the scan path reads that.
			if recordFactHistory(fact, entry.StoredVal,
				datahelpers.GetStringField(fact, "verified_at", "")) {
				res.FactsHistoryRecorded++
			}
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
	factDrift := planSiteFactDrift(ctx, db, siteID, eb, res, dryRun, logger)
	res.FactDrift = factDrift.Emissions

	// Phase 4 (bugs_open/288) — ADOPTION. Measured 2026-08-24: 1 of 132 current
	// tool PLANs declares anything, and 178 tool pages sit on sites that have a
	// register, so a mechanism gated on hand-authoring covers ~0% of the fleet
	// for ever. Propose the bindings that are already visible: probe each
	// NON-declaring tool's script text for this site's own registered values and
	// file a paste-ready suggestion. Read-only here; the write is below and is
	// suppressed on a dry run.
	factSuggestions := planFactBindingSuggestions(ctx, db, siteID, eb, factDrift.declaringSubjects, logger)
	res.FactBindingSuggestions = len(factSuggestions)

	if dryRun {
		return res, nil // report only — write nothing, raise nothing
	}

	if len(res.InvalidBannedClaimPatterns) > 0 {
		created, err := createInvalidBannedClaimPatternItems(
			ctx, db, siteID, domain, res.InvalidBannedClaimPatterns, params.AgentType, logger)
		if err != nil {
			logger.Warn("refresh_evidence_base: invalid_banned_claim_pattern write failed", zap.Error(err))
		}
		res.InvalidBannedClaimWorkItemsCreated = created
	}

	writeFactBindingSuggestions(ctx, db, siteID, factSuggestions, dryRun, logger)

	if changed {
		if err := writeRefreshedEvidenceBase(ctx, db, siteID, specRowID, eb, pinned, res, logger); err != nil {
			return nil, err
		}
		// bugs_open/427: the register just changed (a fact re-synced, a
		// human's edit re-read on the next pass, or a new event fact's daily
		// re-verification). Tell any page that consumes it via
		// query.upcoming_events. This sits AFTER writeRefreshedEvidenceBase
		// has returned successfully, and reads the OUTCOME it already
		// recorded rather than re-deriving one — it introduces no new gate.
		// "skipped_concurrent_edit" is that function's PRE-EXISTING sentinel
		// (set at its CAS-miss branch, :1471, not added by this change): a
		// concurrent human edit won the race, the write was REFUSED, and
		// `eb`/`res` describe a register that was never persisted — so there
		// is nothing new for a consumer to re-resolve, and querying
		// ConsumerPages here would be correct-but-pointless work on every
		// contested write. (Council REVISE 08f56b7e, guardian: asked where
		// this sits relative to the CAS guard and whether the sentinel was
		// new — answered here rather than left implicit.)
		if res.WriterBlock != "skipped_concurrent_edit" {
			res.EvidenceConsumerPagesQueued = queueEvidenceBasePageRerenders(ctx, db, siteID, logger)
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
	ComponentID string
	// SubjectKey addresses the artefact by the TOOL that owns it instead of by a
	// page_components row id (RFC_025 stage 2b, bugs_open/288 §5.6). Mutually
	// exclusive with ComponentID.
	//
	// WHY: a component_id dies. bugs_closed/225's own component (55682bc8-…) no
	// longer exists — the page was decomposed into prose-0 / tool-1 / prose-2 —
	// so an artifact_check written when that bug was filed would today resolve
	// to nothing and fail closed for ever, which reads exactly like a check
	// working. A subject key survives decomposition because it is resolved
	// through the platform's own name rule every pass.
	SubjectKey    string
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
		SubjectKey:    strings.TrimSpace(datahelpers.GetStringField(raw, "subject_key", "")),
		Pattern:       datahelpers.GetStringField(raw, "pattern", ""),
		MustBePresent: datahelpers.GetBoolField(raw, "must_be_present", true),
	}
	// Exactly one address. Both is ambiguous and neither is unrunnable, and both
	// fail CLOSED rather than picking a winner — RFC_017, and the same posture
	// the ErrNoRows arm below already takes.
	if spec.ComponentID != "" && spec.SubjectKey != "" {
		return spec, fmt.Errorf(
			"artifact_check carries BOTH component_id (%s) and subject_key (%s) — exactly one address, "+
				"or the check silently proves a different artefact than its author meant",
			spec.ComponentID, spec.SubjectKey)
	}
	var missing []string
	if spec.ComponentID == "" && spec.SubjectKey == "" {
		missing = append(missing, "component_id or subject_key")
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

// factHasNonArtifactSource reports whether this fact already has a PRIMARY
// verification mechanism, i.e. one of the arms after the pre-pass will handle
// it. Only then is an artifact_check a SECONDARY check that must not own
// verified_at. A fact whose only mechanism is artifact_check keeps taking the
// original branch, unchanged.
func factHasNonArtifactSource(fact map[string]interface{}, src map[string]interface{}) bool {
	if _, has := src["citation"]; has {
		return true
	}
	if _, has := src["attested_by"]; has {
		return true
	}
	return factSQLSource(fact) != ""
}

// artifactCheckSubjectSurfaceQuery collects the stored HTML of every active
// component on every page THIS SITE has for a tool subject key.
//
// Same join family as the fact-drift fan-out's factDriftIndexQuery, and for the
// same reason: it reuses discovery_checks.ToolSubjectKeyExpr (the platform's own
// name rule, which Tier 4 already uses to find a tool's URL) rather than the
// acceptance ladder's toolEligibilityWhere. Measured 2026-08-16 and unchanged
// since: that predicate's sole-component clause admits NEITHER of the two SDLT
// tools this mechanism exists for, so addressing through it would produce a
// check that can never run on the artefacts that motivated it — and would read
// as a clean pass.
//
// The archived-AND-never-deployed exclusion is the AUDIT predicate, not the
// liveness one: an archived page that was deployed is still being served, and
// judging only p.status='active' pages is the mistake the council's
// debug_historian seat caught in the fan-out.
var artifactCheckSubjectSurfaceQuery = `
	SELECT COALESCE(string_agg(COALESCE(pc.rendered_html, ''), E'\n' ORDER BY p.name, pc.position), '')
	FROM pages p
	JOIN page_components pc ON pc.page_id = p.id
	JOIN content_components cc ON cc.id = pc.component_id AND cc.is_active = true
	WHERE p.site_id = $1
	  AND ` + discovery_checks.ToolSubjectKeyExpr + ` = $2
	  AND NOT (p.status = 'archived' AND ` + datahelpers.NeverDeployedPagePredicateFor("p") + `)`

// resolveArtifactCheckSurface returns the stored bytes the pattern is matched
// against, plus a human-readable name for the address (for the drift detail),
// or an error that the caller turns into outcome=error — never a pass. RFC_017:
// a check that cannot run must not be reported as one that ran.
func resolveArtifactCheckSurface(ctx context.Context, db *sql.DB, siteID uuid.UUID, spec *artifactCheckSpec) (string, string, error) {
	if spec.SubjectKey != "" {
		var surface string
		err := db.QueryRowContext(ctx, artifactCheckSubjectSurfaceQuery, siteID, spec.SubjectKey).Scan(&surface)
		if err != nil && err != sql.ErrNoRows {
			return "", "", fmt.Errorf("artifact_check: reading the surface for tool %q failed: %v", spec.SubjectKey, err)
		}
		// An EMPTY surface is an error, not an absence. Both readings are
		// "the pattern is not there", and only one of them is about the artefact:
		// a renamed page or a deactivated component must not be reported as a
		// drifted claim, and must certainly not be reported as fresh.
		if strings.TrimSpace(surface) == "" {
			return "", "", fmt.Errorf(
				"artifact_check.subject_key %q resolves to no stored component HTML on THIS site (%s) — "+
					"the page may have been renamed, or its components deactivated; refused rather than read as absent",
				spec.SubjectKey, siteID)
		}
		return surface, fmt.Sprintf("tool %q", spec.SubjectKey), nil
	}

	componentID, err := uuid.Parse(spec.ComponentID)
	if err != nil {
		return "", "", fmt.Errorf("artifact_check.component_id %q is not a valid id: %v", spec.ComponentID, err)
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
		return "", "", fmt.Errorf(
			"artifact_check.component_id %s does not resolve to a page_components row on THIS site (%s) — "+
				"either it does not exist, or it belongs to a different site and is refused rather than trusted",
			componentID, siteID)
	case err != nil:
		return "", "", fmt.Errorf("artifact_check: reading component %s failed: %v", componentID, err)
	}
	return rendered.String, fmt.Sprintf("component %s", componentID), nil
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
// ownsVerifiedAt distinguishes the two callers. TRUE for a fact whose ONLY
// verification mechanism is this one (the original, sole-reachable case —
// behaviour unchanged): it owns fact["verified_at"]. FALSE when this runs as a
// SECONDARY check beside a citation/sql/attested primary (stage 2b): the primary
// arm owns verified_at and `changed`, and a PASSING artifact check must not bump
// the date on a fact whose citation was just lost, nor flip `changed` and
// thereby open shouldRaiseStaleEvidence for unrelated citation drift. Leaving
// that gate's semantics alone for existing callers is exactly what RFC_025's
// ratification was scoped to guarantee.
func refreshArtifactCheckFact(ctx context.Context, db *sql.DB, siteID uuid.UUID, fact map[string]interface{}, today string, ownsVerifiedAt bool) *evidenceFactRefresh {
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

	re, err := regexp.Compile(spec.Pattern)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = fmt.Sprintf("artifact_check.pattern %q does not compile: %v", spec.Pattern, err)
		return entry
	}

	surface, addr, err := resolveArtifactCheckSurface(ctx, db, siteID, spec)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = err.Error()
		return entry
	}

	found := re.MatchString(surface)
	switch {
	case found == spec.MustBePresent:
		entry.Outcome = "fresh"
		if ownsVerifiedAt {
			fact["verified_at"] = today
			entry.VerifiedAt = today
		}
	case spec.MustBePresent:
		entry.Outcome = "drifted"
		entry.Detail = fmt.Sprintf(
			"artifact_check: pattern %q no longer found in %s — the artefact this fact cites may have changed; the published claim needs a human ruling",
			spec.Pattern, addr)
	default:
		entry.Outcome = "drifted"
		entry.Detail = fmt.Sprintf(
			"artifact_check: pattern %q is now PRESENT in %s, but the fact asserts it must be absent — the published claim needs a human ruling",
			spec.Pattern, addr)
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

// recordFactHistory appends a fact's OUTGOING reading to its history, in place
// on the raw map, and reports whether it wrote anything (bugs_open/386).
//
// Why the raw map rather than the typed EvidenceFact: this whole action works on
// map[string]interface{} precisely so unknown keys survive the rewrite, and
// round-tripping one fact through the struct to add a field would drop every key
// the struct does not know about — the exact data loss the comment at the top of
// refreshOneSite exists to prevent.
//
// Four refusals, each closing a way this could quietly widen what the scan
// accepts:
//   - unarmed facts are skipped, so retention is opt-in per fact and the unsafe
//     side is off by default;
//   - a fact with no outgoing value records nothing — there is no former reading
//     to remember, and writing a zero would invent one;
//   - a reading identical to the newest retained entry is not duplicated, so a
//     same-day second refresh cannot inflate the history;
//   - the array is trimmed to the cap from the FRONT, dropping the oldest.
func recordFactHistory(fact map[string]interface{}, outgoing *float64, outgoingVerifiedAt string) bool {
	if fact == nil || outgoing == nil {
		return false
	}
	if !datahelpers.GetBoolField(fact, "retain_history", false) {
		return false
	}

	existing, _ := fact["history"].([]interface{})

	// A repeat of the newest entry is not a new reading. Without this, two
	// refreshes on one day (measured: 2026-08-16 has two) would each append, and
	// the cap would then be reached in half the calendar time it claims to cover.
	if n := len(existing); n > 0 {
		if last, ok := existing[n-1].(map[string]interface{}); ok {
			if v, ok := numericField(last["value"]); ok && math.Abs(v-*outgoing) < 1e-9 {
				return false
			}
		}
	}

	entry := map[string]interface{}{"value": *outgoing}
	if outgoingVerifiedAt != "" {
		entry["verified_at"] = outgoingVerifiedAt
	}
	updated := append(existing, entry)
	if len(updated) > datahelpers.FactHistoryMaxEntries {
		updated = updated[len(updated)-datahelpers.FactHistoryMaxEntries:]
	}
	fact["history"] = updated
	return true
}

// composeWriterBlock rebuilds the V2 writer whitelist from the register.
// Structure (the three headers) is the machine's; every claim sentence is the
// human's `writer_line`, with `{value}` replaced by the current number. A fact
// without a writer_line is omitted — never auto-phrased.
func composeWriterBlock(eb map[string]interface{}) string {
	factsRaw, _ := eb["facts"].([]interface{})
	var numbers, events, capabilities []string

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
		} else if eventDate := strings.TrimSpace(datahelpers.GetStringField(fact, "event_date", "")); eventDate != "" {
			// bugs_open/427: a dated event fact (news_feed_ingestion's
			// registration writes event_date/venue/participants/broadcaster
			// alongside a citation, kind="entity" — no numeric `value`, so
			// without this branch it would fall to the CAPABILITIES arm below
			// and ship any {event_date}/{venue}/{participants}/{broadcaster}
			// token in its writer_line UNSUBSTITUTED, literal braces and all,
			// into the writer's prompt — the CAPABILITIES arm only ever
			// existed to carry a token-free line verbatim.
			events = append(events, "- "+substituteEventTokens(line, fact, eventDate))
		} else {
			capabilities = append(capabilities, "- "+line)
		}
	}

	if len(numbers) == 0 && len(events) == 0 && len(capabilities) == 0 {
		return "" // nothing phrased — leave the existing block alone
	}

	var b strings.Builder
	if len(numbers) > 0 {
		b.WriteString("NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):\n")
		b.WriteString(strings.Join(numbers, "\n"))
	}
	if len(events) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("SCHEDULED EVENTS (state only as dated below; a blank field is \"TBC\", never invent one):\n")
		b.WriteString(strings.Join(events, "\n"))
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

	// Human-owned guidance, carried VERBATIM through every regeneration. This is
	// NEGATIVE/PROHIBITIVE guidance ONLY by contract (NEVER-STATE lists, bans) —
	// never positive writing instructions or claims about our own accuracy: the
	// field is durable across regeneration BY DESIGN, so anything in it ships
	// for ever (compliance seat, council 0de22385 r1). It does not widen the
	// trust surface — an unmanaged hand-written writer_block is already
	// verbatim-into-prompt and unvalidated — and wherever this field takes
	// effect its text lands inside the composed writer_block, which is on
	// brief-negation-check's runtime-derived writer-visible surface.
	// DURABILITY, verified against the typed-struct landmine ("parsing
	// evidence_base through EvidenceBase and writing it back DELETES unlisted
	// fields"): NO write path round-trips that struct — all nine
	// ParseEvidenceBase callers are readers or validation guards (the admin
	// handler stores the client's raw bytes), and the two real write paths
	// (this action's raw-map marshal; write_site_spec's siteSpecDeepMerge)
	// preserve unknown keys, pinned by round-trip tests. The clincher is
	// precedent: writer_block itself is equally unlisted in the struct, and no
	// site has ever lost one to a struct write — the struct is the SCANNER'S
	// read model, and the landmine stands guard over anyone changing that.
	// This is
	// the carry that lets a site with a hand-written NEVER-STATE list adopt
	// writer_block_managed at all: without it, the first regeneration DELETES
	// the hand-written half — the recurring estate defect ("a generator that
	// rebuilds a row from its source silently reverts every hand edit", the
	// bugs_open/288 lane's landmine, one table over) that kept 13 of 19
	// writer_block sites unmanaged as of 2026-08-25 (bugs_open/387: a stand-in
	// token hand-typed into one of those unmanaged blocks reached the public).
	// Opt-in with the unsafe side OFF: an absent or empty key produces
	// byte-identical output to before this existed (pinned by test). Ordering
	// note: the nothing-phrased early return above still wins — guidance alone
	// never causes regeneration to replace a hand-written block.
	if g := strings.TrimSpace(datahelpers.GetStringField(eb, "writer_block_guidance", "")); g != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(g)
	}
	return b.String()
}

// substituteEventTokens fills a dated-event writer_line's {event_date}/{venue}/
// {participants}/{broadcaster} tokens (bugs_open/427). Unlike {value}, a token
// with nothing stated renders as "TBC" rather than being left as a bare brace:
// {value} always has a caller-supplied number by construction, but an event's
// venue or broadcaster is routinely unstated at announcement time, and a raw
// "{venue}" reaching published copy would be a worse defect than the honest
// placeholder. participants joins with ", " rather than " vs " — a boxing
// fight reads fine either way, but the estate's other verticals (a launch
// panel, a hearing with several parties) are not always exactly two names.
func substituteEventTokens(line string, fact map[string]interface{}, eventDate string) string {
	tbc := func(s string) string {
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
		return "TBC"
	}
	line = strings.ReplaceAll(line, "{event_date}", tbc(eventDate))
	line = strings.ReplaceAll(line, "{venue}", tbc(datahelpers.GetStringField(fact, "venue", "")))
	line = strings.ReplaceAll(line, "{broadcaster}", tbc(datahelpers.GetStringField(fact, "broadcaster", "")))
	participants := datahelpers.ExtractStringListHelper(fact["participants"])
	line = strings.ReplaceAll(line, "{participants}", tbc(strings.Join(participants, ", ")))
	return line
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

// queueEvidenceBasePageRerenders tells every page that consumes
// query.upcoming_events (bugs_open/427) that the register just changed,
// exactly the way queueNewsPageRerenders does for query.latest_news/
// query.news_archive — same shared consumer lookup (RFC_052), same
// page_rerender shape, same reason (section_data_resolved), best-effort
// after the register write has already committed. See that function's
// comment for why the page-status/owned-page/template-filter predicates all
// live in the shared queryresolve.ConsumerPages lookup rather than here.
func queueEvidenceBasePageRerenders(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) int {
	if db == nil {
		return 0
	}
	pages, err := queryresolve.ConsumerPages(ctx, db, siteID, queryresolve.DepEvidenceBase, logger)
	if err != nil {
		logger.Warn("queueEvidenceBasePageRerenders: consumer lookup failed", zap.Error(err))
		return 0
	}
	if len(pages) == 0 {
		return 0
	}

	batchID := uuid.New()
	queued := 0
	for _, page := range pages {
		spec := fmt.Sprintf(
			`{"reason":"section_data_resolved","page_name":%q,"page_id":%q,"domain":%q}`,
			page.Name, page.ID.String(), page.Domain)
		itemKey := pageRerenderItemKey(page.Name, siteID, "section_data_resolved")

		inserted, err := insertPageRerenderItem(ctx, db, siteID, page.ID,
			"refresh_evidence_base", "low",
			fmt.Sprintf("Re-render %s — evidence base changed (dated event facts)", page.Name),
			spec, itemKey, batchID)
		if err != nil {
			logger.Warn("queueEvidenceBasePageRerenders: insert failed",
				zap.String("page", page.Name), zap.Error(err))
			continue
		}
		if inserted {
			queued++
		}
	}

	if queued > 0 {
		logger.Info("queueEvidenceBasePageRerenders: queued scoped re-renders",
			zap.Int("queued", queued), zap.Int("consumer_pages", len(pages)))
	}
	return queued
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

// createInvalidBannedClaimPatternItems raises one work item PER invalid
// pattern (RFC_060 §1e/§3e), not one per site — deliberately NOT modelled on
// createStaleEvidenceItem's per-site key. dropOnConflict (insertWorkItem,
// ON CONFLICT ... DO NOTHING), never refreshOnConflict: a daily re-write of
// an open item bumps updated_at, and the stale-item reaper that keys on it
// would then never reap this row (bugs_closed/213) — the finding is a
// standing defect until a human fixes the pattern, not a value that needs
// its description kept current. Returns the count actually inserted (a
// DO-NOTHING conflict on an already-open item for the SAME pattern is not a
// failure — it means yesterday's finding is still open, which is correct).
func createInvalidBannedClaimPatternItems(
	ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string,
	invalid []invalidBannedClaimPattern, agentType string, logger *zap.Logger,
) (int, error) {
	inserted := 0
	for _, bad := range invalid {
		specJSON, err := json.Marshal(map[string]interface{}{
			"check":         "banned_claim_pattern_compile",
			"domain":        domain,
			"pattern":       bad.Pattern,
			"reason":        bad.Reason,
			"compile_error": bad.Error,
			"fix": "This banned_claims pattern does not compile as a regex and has silently degraded to a " +
				"literal match of its own source text (claims.go:348) — it is armed, listed and counted, but " +
				"very likely matches nothing on the live site. Fix the regex (common cause: an unescaped " +
				"paren or bracket) and re-save the evidence_base; the pattern will compile and start " +
				"scanning for real on the next refresh.",
		})
		if err != nil {
			logger.Warn("refresh_evidence_base: failed to marshal invalid_banned_claim_pattern spec",
				zap.String("site_id", siteID.String()), zap.Error(err))
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return inserted, fmt.Errorf("begin tx: %w", err)
		}

		write, err := writeWorkItem(ctx, tx, workItem{
			siteID:       siteID,
			source:       "scheduled",
			pipeline:     "content",
			itemType:     "invalid_banned_claim_pattern",
			severity:     "medium",
			summary:      fmt.Sprintf("Banned-claim pattern does not compile (%s): %q", domain, bad.Pattern),
			spec:         string(specJSON),
			priority:     35,
			handlerAgent: "human-review",
			status:       "needs_human_review",
			createdBy:    agentType,
			itemKey:      bannedClaimPatternItemKey(siteID, bad.Pattern),
		}, dropOnConflict, logger)
		if err != nil {
			tx.Rollback()
			logger.Warn("refresh_evidence_base: failed to create invalid_banned_claim_pattern item",
				zap.String("site_id", siteID.String()), zap.Error(err))
			continue
		}
		if err := tx.Commit(); err != nil {
			logger.Warn("refresh_evidence_base: failed to commit invalid_banned_claim_pattern item",
				zap.String("site_id", siteID.String()), zap.Error(err))
			continue
		}
		if write.Inserted {
			inserted++
		}
	}
	if inserted > 0 {
		logger.Warn("refresh_evidence_base: invalid banned_claims pattern(s) found — silently degraded to literal match",
			zap.String("site_id", siteID.String()),
			zap.String("domain", domain),
			zap.Int("invalid_patterns", len(invalid)),
			zap.Int("items_inserted", inserted))
	}
	return inserted, nil
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
