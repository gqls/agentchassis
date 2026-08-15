// FILE: platform/orchestration/actions/write_audit_findings_retraction.go
//
// D1 HALF TWO (bugs_open/213): the design audit closes its own tickets when it
// stops reporting the defect — SITE-scoped, and only after N CONSECUTIVE
// silences.
//
// Half one (completion gate 1b, complete_work_item_no_change.go) stopped
// `dark_section_audit` items being stamped `complete` by a handler that
// reported it changed nothing. It could only refuse a false green; it left
// those items cycling to `failed` through a handler that provably cannot
// repair them, with no path back out. THIS is the path back out: when the
// defect is actually fixed — by a human, by a template change, by anything —
// the audit stops reporting it, and the ticket closes on that observation
// instead of sitting `failed` for ever.
//
// EVERY CLAUSE BELOW IS A CONSTRAINT THAT WAS MEASURED, NOT A PREFERENCE.
//
// ── WHY SITE-SCOPED AND NOT PAGE-SCOPED ─────────────────────────────────────
// WII-016 (the render audit's retraction) scopes to the pages the adapter
// SUCCESSFULLY MEASURED, which is the right shape and is unavailable here.
// This producer's page identity is `spec.page_name`, and that field is FREE
// LLM PROSE: the live values include `all`, `all pages`, `index`, `global`,
// `sitewide`, `index / about`, and a comma-joined list of three slugs.
// [MEASURED 2026-08-15] two of the eight rows filed on 2026-08-14 carry `all`
// and `all pages` — two spellings of one idea, on the same day. It cannot be
// resolved to a page, so a page-scoped rule would be inventing its own scope.
// The honest unit of observation this producer can support is the SITE.
//
// ── WHY THE SILENCE IS SITE-LEVEL, NOT PER-KEY ──────────────────────────────
// The item_key is `{audit_source}_{item_type}_{page_name}_{site_id}`, so it
// inherits the free prose above AND the audit_source literal. [MEASURED
// 2026-08-15] that literal was renamed `design-audit` → `visual-design-audit`
// between 08-12 and 08-13, and gamesdesign.co.uk therefore holds TWO rows for
// the same defect on the same page under two different keys. A per-key silence
// rule would have read the rename as fifteen defects being fixed at once.
// So the question asked here is about the SITE: did this producer look at this
// site and report NO dark-section defect at all? Any dark-section finding,
// under any spelling, on any page, keeps every one of that site's tickets
// alive. That is the conservative direction, and it is the one WII-016's
// header argues for — erring this way costs a ticket staying open another
// week; erring the other way closes a live defect.
//
// ── WHY N = 3 ───────────────────────────────────────────────────────────────
// Arithmetic, not a safety margin. [MEASURED 2026-08-13] the audit re-reported
// a KNOWN-UNREPAIRED colour defect on 7 of 7 post-closure re-visits across 4
// sites. Zero misses in seven trials does NOT bound the per-run miss rate at
// zero: the 95% upper bound is 1 - 0.05^(1/7) ≈ 0.35. A rule firing on one
// silence would therefore close a live defect up to ~35% of the time it ran;
// on two, ~12%; on three, ~4.2%. Three is the first N under 5%. If that
// measurement is ever re-run with more trials, THIS is the number to move, and
// the bound — not the observed count — is what moves it.
//
// ── WHY `triaged` AND `claimed` ROWS ARE EXCLUDED ───────────────────────────
// Two reasons, and the second is the load-bearing one because it is mechanical.
// (a) Both mean the pipeline is actively carrying the item — a handler is about
// to run or is running — and an audit's silence recorded mid-repair is racing
// it. (b) [MEASURED 2026-08-15, enumerated not assumed] exactly ONE live reader
// of `site_work_items.updated_at` exists — the `stale-work-item-reaper`
// scheduled task, enabled, hourly, whose predicate is
// `status='triaged' AND pipeline='build' AND updated_at < NOW() - INTERVAL '48
// hours' AND claimed_at IS NULL`. (The other two tasks that mention
// updated_at read `sites.updated_at` and `orchestration_states.updated_at`.)
// A silence-streak write bumps `updated_at` — UNAVOIDABLY, because
// `trg_site_work_items_updated_at` fires BEFORE UPDATE on every write to this
// table, so no careful column list can prevent it — and migration 237's own
// header names this exact hazard: "anything that touches the row for unrelated
// reasons bumps updated_at and postpones reaping, so a genuinely wedged item
// that is periodically written to would never be parked". With the sweep at
// 900s and the reaper at 48h, this file would have made that permanent for
// every `triaged` dark-section row. The status exclusion is the protection;
// there is no other one available.
//
// ── WHY IT IS SCOPED TO THIS RUN'S audit_source ─────────────────────────────
// Six live agents call write_audit_findings, and any of them could emit a
// `dark_section` category — the item_type does not name its producer. WII-016's
// guardian objection is exactly this ("adopting retraction on a CO-FILED
// item_type silently closes the other producer's findings"), and the answer
// there was a census. Here it is structural: a candidate is only in scope when
// its own `spec.audit_source` equals the source of the run doing the observing.
// A producer's silence says nothing about another producer's findings.
// [MEASURED 2026-08-15] all 23 dark_section_audit rows ever filed carry
// audit_source ∈ {design-audit, visual-design-audit}, both from
// visual-design-auditor via design-audit-agent — but that census sees only
// producers that have FILED, so the structural guard stands rather than the
// count. ⚠ Its cost, stated: because the source is part of the scope, renaming
// the literal again resets every streak and orphans the old rows from
// retraction. That is the safe direction and it is not free.
//
// OPT-IN, WITH THE UNSAFE DEFAULT OFF, per the owner's 2026-08-02 shared-seam
// ruling. An item_type absent from silenceRetractionGates takes a map miss and
// this file does nothing for it. The unsafe side really is the default: for a
// type whose auditor legitimately reports intermittently, three silences mean
// nothing at all.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// silenceRetractionRule declares, for ONE item_type, that this producer's
// sustained silence about the SITE is evidence the defect is gone.
type silenceRetractionRule struct {
	// Why the type opts in: the measurement that licenses reading silence as
	// evidence at all. Surfaced into the retraction reason, so whoever reads a
	// closed row sees the basis rather than the verb.
	Why string

	// ConsecutiveSilences is N — how many successive runs of this producer must
	// report nothing about the site before its open rows close. Never 1 unless
	// the observation is a MEASUREMENT rather than an LLM's report.
	ConsecutiveSilences int
}

// silenceRetractionGates is the opt-in roster. Absent item_type → inert.
//
// The bar for an entry: you can state what the producer's silence MEASURES,
// and you have bounded its per-run miss rate. Without the second, N is a guess.
var silenceRetractionGates = map[string]silenceRetractionRule{
	"dark_section_audit": {
		Why: "the design audit re-reported a known-unrepaired dark-section defect on 7 of 7 " +
			"post-closure re-visits (measured 2026-08-13), which bounds its per-run miss rate at " +
			"~35% and puts three consecutive silences under 5%",
		ConsecutiveSilences: 3,
	},
}

// workItemInFlightStatuses are the statuses at which the build pipeline is
// actively carrying an item, and at which this file must not write.
//
// It sits with its three siblings in work_items_common.go rather than beside
// them only because it is this seam's, not the work-item vocabulary's — the
// comparison that matters is with workItemClosedStatuses:
//
//	closed    (never retract):        complete verified rejected wont_fix cancelled
//	in-flight (never touch, here):    triaged claimed
//
// Everything else — detected, blocked, failed, unresolved, deferred,
// needs_human_review — can accrue a silence observation. See the header for
// why `triaged` in particular is not negotiable (the stale reaper keys on
// updated_at, and the table's trigger bumps it on every write).
var workItemInFlightStatuses = map[string]bool{
	"triaged": true,
	"claimed": true,
}

// silenceState is what a candidate row carries about its own streak, under
// result.retraction.
type silenceState struct {
	SilentRuns  int    `json:"silent_runs"`
	AuditSource string `json:"audit_source"`
	LastSilent  string `json:"last_silent_at,omitempty"`
}

// readSilenceState pulls result.retraction off a candidate.
//
// It deliberately does NOT re-check the stored audit_source against this run's,
// and the omission is the interesting part. A row only ever receives a streak
// write after passing the spec.audit_source scope check in the caller, and
// spec is immutable once filed — so the stored source is always this run's by
// construction, and a second check here would be a branch no test could reach.
// It was written, and removed when the mutation matrix showed deleting it
// changed no test's verdict: untested defence reads as protection and is not.
// The source is still WRITTEN into the row, as provenance for whoever reads it.
// ⚠ If a future change lets anything else write result.retraction, or widens
// the caller's scope check, that invariant is gone and this needs the check back.
func readSilenceState(result []byte) silenceState {
	if len(result) == 0 {
		return silenceState{}
	}
	var wrapper struct {
		Retraction silenceState `json:"retraction"`
	}
	if err := json.Unmarshal(result, &wrapper); err != nil {
		// An unreadable result is not evidence of a streak. Start again — the
		// error direction is "retract later than we could have".
		return silenceState{}
	}
	return wrapper.Retraction
}

// specAuditSource reads spec.audit_source off a candidate row.
func specAuditSource(spec []byte) string {
	if len(spec) == 0 {
		return ""
	}
	var s struct {
		AuditSource string `json:"audit_source"`
	}
	if err := json.Unmarshal(spec, &s); err != nil {
		return ""
	}
	return s.AuditSource
}

// silenceRetractionResult is the reported outcome, one entry per gated type.
// Every field is reported, including the zeros: "nothing retracted" and "the
// producer spoke, so nothing could have been" are different facts, and a single
// `retracted` number cannot tell them apart.
type silenceRetractionResult struct {
	Silent          bool `json:"silent"`
	Candidates      int  `json:"candidates"`
	StreaksBumped   int  `json:"streaks_bumped"`
	StreaksReset    int  `json:"streaks_reset"`
	Retracted       int  `json:"retracted"`
	RetractedPark   int  `json:"retracted_parked"`
	SkippedInFlight int  `json:"skipped_in_flight"`
}

// retractSilentAuditFindings runs the roster for one site, in one transaction.
//
// `observedItemTypes` MUST be built from every finding this run classified,
// BEFORE the blocked / dedup / insert filters — see the shared helper's header
// for why. `shapeRecognised` is the availability guard: it is false when the
// audit's reply could not be parsed into a findings list at all, and a run that
// did not understand its own instrument has observed NOTHING, which is a
// different fact from "observed nothing wrong".
func retractSilentAuditFindings(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	auditSource string,
	batchID uuid.UUID,
	observedItemTypes map[string]bool,
	shapeRecognised bool,
	logger *zap.Logger,
) (map[string]interface{}, error) {
	if !shapeRecognised || len(silenceRetractionGates) == 0 {
		return nil, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("retraction: begin: %w", err)
	}
	defer tx.Rollback()

	out := map[string]interface{}{}
	for itemType, rule := range silenceRetractionGates {
		res, rErr := retractOneSilentType(ctx, tx, siteID, auditSource, batchID,
			itemType, rule, !observedItemTypes[itemType], logger)
		if rErr != nil {
			return nil, rErr
		}
		out[itemType] = res
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("retraction: commit: %w", err)
	}
	return out, nil
}

func retractOneSilentType(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	auditSource string,
	batchID uuid.UUID,
	itemType string,
	rule silenceRetractionRule,
	silent bool,
	logger *zap.Logger,
) (silenceRetractionResult, error) {
	res := silenceRetractionResult{Silent: silent}

	candidates, err := loadAuditRetractionCandidates(ctx, tx, siteID, itemType)
	if err != nil {
		return res, err
	}

	// Pass one decides; pass two writes. Nothing is written while the decision
	// is still being formed, so the streak a row is retracted on is the streak
	// that was read, and a partial failure cannot leave half the site advanced.
	type plan struct {
		cand   auditRetractionCandidate
		streak int
	}
	var toBump, toReset, toRetract []plan

	// Every candidate gets a verdict, including the ones this run may not judge:
	// decide() below is then a pure lookup, and a row missing from this map is a
	// bug rather than a silently-kept row.
	verdicts := make(map[uuid.UUID]retractionVerdict, len(candidates))

	for _, c := range candidates {
		if workItemInFlightStatuses[c.Status] {
			res.SkippedInFlight++
			verdicts[c.ID] = retractionOutOfScope
			continue
		}
		if specAuditSource(c.Spec) != auditSource {
			// Another producer's finding. This run's silence is not evidence
			// about it — that is the co-filing trap, refused structurally.
			verdicts[c.ID] = retractionOutOfScope
			continue
		}
		res.Candidates++

		state := readSilenceState(c.Result)
		if !silent {
			// The producer spoke about this site, so the streak is broken. Only
			// write when there is something to clear.
			if state.SilentRuns > 0 {
				toReset = append(toReset, plan{cand: c})
			}
			verdicts[c.ID] = retractionStillFailing
			continue
		}
		next := state.SilentRuns + 1
		if next >= rule.ConsecutiveSilences {
			toRetract = append(toRetract, plan{cand: c, streak: next})
			verdicts[c.ID] = retractionResolved
		} else {
			toBump = append(toBump, plan{cand: c, streak: next})
			// Silent, but not yet enough of them. Not a retraction and not a
			// re-observation of the defect — the streak is the whole record.
			verdicts[c.ID] = retractionStillFailing
		}
	}

	for _, p := range toReset {
		if _, err := tx.ExecContext(ctx, `
			UPDATE site_work_items
			   SET result = COALESCE(result, '{}'::jsonb) - 'retraction'
			 WHERE id = $1`, p.cand.ID); err != nil {
			return res, fmt.Errorf("reset silence streak %s: %w", p.cand.ID, err)
		}
		res.StreaksReset++
	}

	for _, p := range toBump {
		if _, err := tx.ExecContext(ctx, `
			UPDATE site_work_items
			   SET result = COALESCE(result, '{}'::jsonb) || jsonb_build_object(
			                  'retraction', jsonb_build_object(
			                      'silent_runs',    $2::int,
			                      'audit_source',   $3::text,
			                      'last_silent_at', now()::text))
			 WHERE id = $1`, p.cand.ID, p.streak, auditSource); err != nil {
			return res, fmt.Errorf("bump silence streak %s: %w", p.cand.ID, err)
		}
		res.StreaksBumped++
	}

	streaks := map[uuid.UUID]int{}
	for _, p := range toRetract {
		streaks[p.cand.ID] = p.streak
	}

	retracted, parked, err := retractResolvedAuditFindings(ctx, tx, siteID, auditSource, batchID, itemType,
		// Availability is the caller's shapeRecognised, already checked above:
		// reaching here means this run understood its instrument.
		true, candidates,
		func(c auditRetractionCandidate) (retractionVerdict, string) {
			if verdicts[c.ID] != retractionResolved {
				return verdicts[c.ID], ""
			}
			return retractionResolved, fmt.Sprintf(
				"%s re-audited this site on %d consecutive runs and reported no %s finding "+
					"on any page (%s)", auditSource, streaks[c.ID], itemType, rule.Why)
		}, logger)
	if err != nil {
		return res, err
	}
	res.Retracted = retracted
	res.RetractedPark = parked

	if res.Retracted > 0 || res.StreaksBumped > 0 {
		logger.Info("write_audit_findings: silence retraction",
			zap.String("item_type", itemType),
			zap.String("audit_source", auditSource),
			zap.Bool("silent", silent),
			zap.Int("candidates", res.Candidates),
			zap.Int("streaks_bumped", res.StreaksBumped),
			zap.Int("retracted", res.Retracted),
			zap.Int("retracted_parked", res.RetractedPark))
	}
	return res, nil
}
