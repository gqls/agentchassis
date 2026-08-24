// FILE: platform/orchestration/actions/work_item_failure_ladder.go   (NEW FILE)
//
// THE WORK-ITEM FAILURE-WRITE CONTRACT — one answer to "what happens when a
// work item fails", for every failure path in the fleet.
//
// bugs_open/307, and the owner ruling of 2026-08-18 it implements: "a transient
// blip should return the item to queued." The bug's §7 asks for exactly one
// mechanism rather than three patches, because the defect IS the disagreement:
// four writers of a work item's terminal status, four different guarantees.
//
// WHAT WAS WRONG, in the order it bites:
//
//  1. NO DELAY BETWEEN ATTEMPTS. FailWorkItemAction's ladder re-triaged an item
//     that claim_work_item could re-claim on the next tick, so during the
//     2026-08-17 git-adapter outage 88 of 100 items burned all three attempts
//     inside a 2h43m window. Retry-without-delay is equivalent to no retry for
//     any outage longer than a few dispatch cycles.
//  2. NO LADDER AT ALL on the other writer. UpdateWorkItemStatusAction's
//     `status: 'failed'` steps (5 live agents, 17 steps) wrote `failed`
//     directly with attempt_count+1 and no CASE, so ONE failure was terminal
//     regardless of max_attempts=3 — in fair weather as much as in an outage.
//     Measured 2026-08-19: 141 of 270 failed rows in 14 days died before
//     exhausting their budget, 139 of them at attempt_count=1. That is the
//     larger half of this bug and it has nothing to do with the outage.
//  3. NO TERMINAL-DECISION GUARD, where both sibling writers on the SUCCESS
//     path have one (CompleteWorkItemAction, failUnverifiedCompletion), with a
//     comment explaining why. A handler that deliberately wrote
//     needs_human_review or wont_fix could have that decision replaced by
//     triaged/failed on the way out — and `triaged` means re-dispatched to be
//     refused again. Invisible until 2026-08-19, because both paths wrote the
//     same word; 9 live `wont_fix` rows now make an overwrite a visible wrong
//     answer.
//
// WHY THE OBVIOUS FIX IS THE DANGEROUS ONE (bug §3). The pre-existing
// count-free branch (isAIUnavailable) retries FOREVER without consuming an
// attempt. Tolerable for LLM credit exhaustion. Not for a 404 on a branch: a
// genuinely deleted repo produces byte-identical error text to a GitHub blip
// (compare bugs_open/131, where sites.github_repo picks the wrong repo and the
// same seam yields a permanent 404). So a release here NEVER consumes an
// attempt AND NEVER omits a cooldown, and it is capped — see releaseCap.
//
// WHY THE TRANSIENT SIGNAL IS A BURST, NOT A STRING. Neither classifier in the
// estate matches the outage's error: MatchedTransientFailure and
// MatchedPermanentFailure both return "" for `github API request failed with
// status: 404 Not Found`, so RetryDisposition says (false, "") —
// unclassified-terminal. Correctly: a 404 alone is not evidence of an outage.
// What IS evidence is the SHAPE — many items, across different domains and
// different agent types, failing with the SAME error inside minutes. A deleted
// repo fails alone. Measured over 7 days (2026-08-19): with the conjunction
// below, exactly three signatures fire and all three were real infrastructure
// outages (Kafka partition-leader, LLM usage-limit, this git 404); ZERO
// single-item permanent faults fire, including the ones that failed repeatedly
// (`page X is rebuild_protected`, `save_page_sections: SEC…`, claims-floor
// blocks) — they are single-domain and single-agent-type by nature, which is
// the conjunction doing its job.
//
// POSTURE: ARMED, WITH ENV DISARMS (the WDS-018 precedent from
// bugs_closed/291 — "the owner has ruled against default-OFF switches that rot
// unexercised"). This implements an explicit owner ruling rather than granting
// caller-licensed authority, so it is not the WII-019 opt-in-default-OFF shape.
// Each new behaviour disarms INDEPENDENTLY, because they fail in different
// ways: the guard could refuse a write, the backoff could read as a stuck
// queue, the release could hold an item. One switch for three behaviours would
// force an operator to give up all three to quiet one.
//
// WHAT THIS DELIBERATELY DOES NOT TOUCH:
//   - FailWorkItemAction's status_override branch. That is bugs_open/033's D2,
//     with its own owner ruling of 2026-07-25 ("the framework should resolve
//     these, not park them"). Changing its attempt semantics from inside this
//     patch would quietly overrule that deferral.
//   - The decision-status writes of update_work_item_status
//     (needs_human_review, wont_fix, unresolved…). Those ARE deliberate
//     decisions; only the `failed` arm is a failure.
//
// SEE ALSO: register WII-0xx (this contract) · RFC_018/SCH-024 (reaper_policies,
// whose numbers this adopts as its second consumer) · RSH-007 (RetryDisposition,
// permanent-before-transient — the sequencing is the contract) · WII-003 (the
// guard pattern, one writer over) · WII-018 (any write bumps updated_at, so this
// writes ONCE per failure event and never on a cadence).

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"go.uber.org/zap"

	perrors "github.com/gqls/agentchassis/platform/errors"
)

// ── Kill switches. Armed when unset (WDS-018 idiom: `== ""` means armed). ────
const (
	envDisableDecisionGuard    = "DISABLE_WORK_ITEM_DECISION_GUARD"
	envDisableRetryBackoff     = "DISABLE_WORK_ITEM_RETRY_BACKOFF"
	envDisableTransientRelease = "DISABLE_WORK_ITEM_TRANSIENT_RELEASE"
	envDisableBurstRelease     = "DISABLE_WORK_ITEM_BURST_RELEASE"
	// bugs_open/345 candidate 2. A fleet-wide disarm for the repeat rule, in
	// this file's existing idiom, so it can be switched off without a config
	// migration on every opted-in dispatcher. It disarms only; it can never
	// arm the rule, because arming is per-item-type and lives in step config.
	envDisableRepeatTermination = "DISABLE_WORK_ITEM_REPEAT_TERMINATION"
)

// ── Tunables. Defaults are the measured ones; env overrides are for an
// operator mid-incident, not for callers (a per-caller knob on a shared seam is
// how a contract becomes seventeen contracts).
const (
	envBurstWindowMinutes   = "WORK_ITEM_BURST_WINDOW_MINUTES"
	envBurstMinErrors       = "WORK_ITEM_BURST_MIN_ERRORS"
	envBurstCooldownMinutes = "WORK_ITEM_BURST_COOLDOWN_MINUTES"
	envTransientReleaseCap  = "WORK_ITEM_TRANSIENT_RELEASE_CAP"

	defaultBurstWindowMinutes   = 10
	defaultBurstMinErrors       = 10
	defaultBurstCooldownMinutes = 15
	defaultTransientReleaseCap  = 10

	// The distinct-consumer minimums are NOT env-tunable, and that is
	// deliberate: they are the whole discriminator. Dropping either to 1
	// turns "the fleet is failing" into "something failed twice", which is
	// exactly the single-item permanent fault that must still exhaust its
	// attempts. Measured 2026-08-19: with both at 2, zero single-item faults
	// fired over 7 days.
	burstMinDistinctDomains    = 2
	burstMinDistinctAgentTypes = 2

	// Fallback when no reaper_policies row applies. 30 minutes × attempt
	// gives 30m then 60m on a max_attempts=3 item, i.e. ~1.5h of span for
	// three attempts — enough to outlive the median outage, and far inside
	// the stale-work-item-reaper's 48h ceiling on a triaged row.
	defaultBackoffMinutes = 30

	// A policy row is operator-set, so it is trusted but not unbounded: a
	// fat-fingered backoff_minutes must not push a row past the 48h reaper.
	maxBackoffMinutes = 720
)

// retryAfterColumnPresent lets the binary survive arriving BEFORE its own
// migration. site_work_items.retry_after is added by migration
// 502_site_work_items_retry_after.sql; on a shared tree ANOTHER lane's roll can
// ship this code at any moment, and a failure path that errors on every write
// would be far worse than the bug it fixes. So: assume present, and on the one
// error Postgres raises for a missing column (42703) fall back permanently to
// the pre-migration statement shape and log it loudly.
var retryAfterColumnPresent atomic.Bool

func init() { retryAfterColumnPresent.Store(true) }

// workItemDecisionStatuses is the guard list for the FAILURE path — the
// statuses that record a DECISION rather than an outcome.
//
// IT IS NOT THE COMPLETION PATH'S LIST, AND THE DIFFERENCE IS THE POINT.
// CompleteWorkItemAction guards with
//
//	needs_human_review failed unresolved rejected wont_fix verified blocked
//
// i.e. `failed` and `unresolved` are protected there. Here they must NOT be:
// overwriting `failed` IS the retry ladder (an item that failed and is being
// retried moves failed → triaged), and it is what the admin retry door at
// site_admin_handlers.go expects. Copying the sibling list verbatim would stop
// fail_work_item re-triaging a failed row — the exact naive move the §third-gap
// contribution to bugs_open/307 warned against, from the thread that found the
// gap and deliberately did not close it.
//
// `cancelled` is here and is absent from the sibling lists — a gap named in the
// council submission rather than fixed silently on the completion path. 400
// rows in 14 days carry it, all owner dispositions; re-triaging one would undo
// a human's decision.
var workItemDecisionStatuses = []string{
	"needs_human_review",
	"wont_fix",
	"rejected",
	"verified",
	"blocked",
	"cancelled",
	// `deferred` added 2026-08-20 answering the council's bug_historian seat
	// (round 2, medium, advisory): these lists are stated as exhaustive, so an
	// omission is a claim — and it was one. `deferred` is the estate's canonical
	// PARKING state (LANDMINES: "`deferred` is the ONLY parking state — a
	// `blocked` one un-parks itself within 600s"), and a park is a decision by
	// exactly the definition this list uses. [MEASURED 2026-08-20] 344 live rows
	// carry it — capability_gap items born deferred, plus migration 389's
	// parked-with-provenance rows — against ZERO for `blocked`, which was in the
	// list from the start. Same class, opposite treatment, and the inconsistency
	// ran in the wrong direction.
	//
	// Unreachable in practice either way, since claim takes only
	// triaged/approved so no handler saga can fail on a deferred row — which is
	// precisely why it belongs here: this list is defence in depth, and `blocked`
	// earns its place on identical reasoning.
	"deferred",
}

// workItemCompletionGuardStatuses is the guard list for a COMPLETION write, and
// it is deliberately NOT the list above. The two sit together so the difference
// is visible rather than discovered — the same idiom work_items_common.go uses
// for its four status vocabularies.
//
//	failure path (above):  needs_human_review wont_fix rejected verified blocked cancelled
//	completion path (here): … PLUS failed and unresolved
//
// WHY THE EXTRA TWO, and it is the whole reason these cannot be one constant.
// On the FAILURE path, overwriting `failed` IS the retry ladder (failed →
// triaged) and overwriting `unresolved` is how a revalidator re-opens a
// given-up item; guarding them there would break the ladder this file exists to
// build. On the COMPLETION path the same two must be protected, because
// stamping `complete` over a row that already failed or was given up is exactly
// the silent-overwrite defect WII-003 closed on CompleteWorkItemAction — the
// class this change is otherwise fixing.
//
// This was caught by the council's editquality seat (corr
// 4cdec68b-fa17-436d-8e25-8c422ee6c8c5, round 1, gating): the first version
// reused the failure list here and its rationale claimed it was "the same guard
// its two siblings already have", which was false in exactly the two entries
// that mattered. Recorded rather than quietly corrected, because a REVISE round
// that finds a real defect is the cheapest place to be wrong.
//
// `cancelled` is present here and absent from the two inlined sibling copies
// (load_work_item_actions.go, complete_work_item_verification.go). That is a
// deliberate addition, not a transcription slip: [MEASURED 2026-08-19] 400 rows
// in 14 days carry it, every one an owner disposition, and stamping `complete`
// over one would undo a human's decision.
//
// ANSWERING THE COUNCIL'S reuse_agent SEAT (round 2, advisory): it asked
// whether the siblings' list is a named Go constant this could import rather
// than re-declare. It is NOT — it is two independently-written INLINE literals
// that happen to match, at load_work_item_actions.go:1032 and
// complete_work_item_verification.go:429, with no named constant for it
// anywhere in the package (verified 2026-08-20). So there was nothing to
// import, and this is the FIRST named version of that vocabulary rather than a
// fifth copy of it. The siblings are NOT edited here —
// converging their inlined copies onto this constant changes THEIR behaviour and
// belongs in its own change.
var workItemCompletionGuardStatuses = []string{
	"needs_human_review",
	"failed",
	"unresolved",
	"rejected",
	"wont_fix",
	"verified",
	"blocked",
	"cancelled",
	// See the failure list above: a park is a decision on both paths.
	"deferred",
}

// jsonbIntFragment reads an integer out of jsonb without trusting its type. A
// bare `(spec->>'k')::int` raises 22P02 on any non-numeric value ever written
// there by anything, and this is the fleet's failure path — it must not be the
// thing that breaks.
const jsonbIntFragment = `COALESCE(CASE WHEN jsonb_typeof(%s) = 'number' THEN (%s)::int END, 0)`

// normSigFragment is the burst signature: an error message with its variable
// parts removed, so two reports of one outage collapse to one string.
//
// IT IS APPLIED IN SQL ON BOTH SIDES ON PURPOSE. Normalising the incoming
// message in Go and the stored rows in SQL would be two implementations of one
// rule — the drift class bugs_closed/034 closed. Passing the raw message as a
// bound parameter and normalising it with the SAME expression makes drift
// unrepresentable.
//
// Order matters: UUIDs first (they contain digit runs), then quoted strings
// (branch names, topics), then bare digit runs (status codes, counts, ports).
// Truncated to 120 chars because a long Go error accumulates per-call detail
// after the part that identifies the fault.
const normSigFragment = `left(regexp_replace(regexp_replace(regexp_replace(%s,` +
	`'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}','<uuid>','g'),` +
	`'"[^"]*"','<q>','g'),'[0-9]+','<n>','g'),120)`

// failureLadderOutcome is what the caller reports. Every field answers a
// question a dispatch loop's own output previously could not: was this the last
// attempt, was the write refused, and if the item is going back, when can it
// actually be claimed.
type failureLadderOutcome struct {
	NewStatus     string
	AttemptsLeft  int
	Released      bool   // transient: returned to the queue, attempt NOT consumed
	ReleaseReason string // burst | ai_unavailable | transient_classified
	Skipped       bool   // the guard refused: a deliberate decision was already recorded
	SkipReason    string
	BackoffMins   int // 0 = claimable immediately
	// bugs_open/345 candidate 2: the budget was ended EARLY because this
	// failure was byte-identical to the one already on the row, i.e. the
	// feedback did not change the outcome. Reported rather than inferred from
	// NewStatus, because "failed at the ceiling" and "failed because retrying
	// is pointless" are different facts and only one of them is a cost saving.
	TerminatedOnRepeat bool
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func envArmed(key string) bool { return os.Getenv(key) == "" }

// applyWorkItemFailureLadder is THE failure-path write. Both failure writers
// route through it; nothing else may write a failure status to site_work_items.
//
// resultMerge is optional jsonb merged into result (update_work_item_status
// carries its orchestration fingerprint there; fail_work_item passes nil).
func applyWorkItemFailureLadder(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	itemID uuid.UUID,
	errorMsg string,
	agentType string,
	resultMerge []byte,
	// bugs_open/345 candidate 2: the CALLING STEP's config, read only for the
	// repeat-termination opt-in (stopOnRepeatFailure). nil is a valid, complete
	// answer meaning "not opted in" — every caller that has no opinion passes
	// nil and gets byte-identical behaviour. It is deliberately the whole step
	// config rather than a pre-extracted bool so the opt-in stays visible where
	// a reviewer of the CALLER reads it, which is the RFC_010 §2 ruling.
	stepConfig map[string]interface{},
) (failureLadderOutcome, error) {
	var out failureLadderOutcome
	if db == nil {
		return out, fmt.Errorf("database connection required")
	}

	// ── Read the row's own state first, then decide in Go. The alternative —
	// encoding eligibility into the UPDATE's WHERE — cannot tell "the guard
	// refused" from "not eligible for release" when it reports 0 rows, and
	// those two need opposite follow-ups. The write still carries the guard,
	// so a concurrent writer cannot be clobbered between the two statements.
	var (
		curStatus            string
		attemptCount         int
		maxAttempts          int
		itemType             string
		transientReleasesRun int
		// bugs_open/345 candidate 2: the failure ALREADY on the row, so the
		// incoming one can be recognised as a repeat of it. NullString because
		// the column is NULL on every item that has not failed yet.
		priorError sql.NullString
	)
	readQ := fmt.Sprintf(`
		SELECT status, attempt_count, max_attempts, item_type, %s, error
		FROM site_work_items WHERE id = $1`,
		fmt.Sprintf(jsonbIntFragment, "spec->'transient_releases'", "spec->>'transient_releases'"))
	if err := db.QueryRowContext(ctx, readQ, itemID).
		Scan(&curStatus, &attemptCount, &maxAttempts, &itemType, &transientReleasesRun, &priorError); err != nil {
		if err == sql.ErrNoRows {
			return out, fmt.Errorf("work item %s not found", itemID)
		}
		return out, fmt.Errorf("failed to read work item state: %w", err)
	}

	// ── Is this a transient failure we should not charge an attempt for?
	//
	// The three classifiers are LAYERED, not merged. isAIUnavailable and
	// RetryDisposition's needle sets disagree in BOTH directions —
	// isAIUnavailable alone matches EOF, status 401/402, "credit",
	// "api key", "authentication", "billing", "model not found";
	// TransientErrorNeedles alone matches "temporary", "service unavailable",
	// "bad gateway". Swapping one for the other silently drops live coverage
	// either way, so both are asked. RetryDisposition is asked via its
	// sequenced form, never MatchedTransientFailure directly: the
	// permanent-before-transient ordering is the contract (RSH-007), and it is
	// what keeps e.g. "pq: invalid connection" terminal.
	release, reason := "", ""
	if envArmed(envDisableTransientRelease) && maxAttempts > 1 && transientReleasesRun < envInt(envTransientReleaseCap, defaultTransientReleaseCap) {
		if envArmed(envDisableBurstRelease) && detectFailureBurst(ctx, db, logger, errorMsg) {
			release, reason = "burst", "burst"
		} else if isAIUnavailable(fmt.Errorf("%s", errorMsg)) {
			release, reason = "ai", "ai_unavailable"
		} else if recoverable, matched := perrors.RetryDisposition(fmt.Errorf("%s", errorMsg)); recoverable {
			release, reason = "classified", "transient_classified:"+matched
		}
	}

	guard := ""
	if envArmed(envDisableDecisionGuard) {
		guard = " AND status NOT IN (" + sqlInList(workItemDecisionStatuses) + ")"
	}

	if release != "" {
		cooldown := envInt(envBurstCooldownMinutes, defaultBurstCooldownMinutes)
		newStatus, attemptsLeft, affected, err := writeTransientRelease(
			ctx, db, itemID, errorMsg, reason, cooldown, resultMerge, guard)
		if err != nil {
			return out, err
		}
		if affected == 0 {
			// The guard refused: something recorded a decision under us.
			out.Skipped, out.SkipReason = true, "decision_status_preserved:"+curStatus
			logger.Info("work item failure ladder: transient release skipped — a deliberate status is already recorded",
				zap.String("item_id", itemID.String()), zap.String("status", curStatus))
			return out, nil
		}
		out.NewStatus, out.AttemptsLeft = newStatus, attemptsLeft
		out.Released, out.ReleaseReason, out.BackoffMins = true, reason, cooldown
		logger.Info("work item failure ladder: transient — returned to the queue with a cooldown, attempt NOT consumed",
			zap.String("item_id", itemID.String()),
			zap.String("reason", reason),
			zap.Int("cooldown_minutes", cooldown),
			zap.Int("releases_so_far", transientReleasesRun+1),
			zap.Int("release_cap", envInt(envTransientReleaseCap, defaultTransientReleaseCap)))
		return out, nil
	}

	// ── bugs_open/345 candidate 2: has the feedback demonstrably failed to
	// change anything? If so, stop paying for the rest of the budget.
	//
	// REACHED ONLY AFTER the transient arm has declined this failure, and that
	// ordering is the contract, not a convenience. A burst, an AI outage or a
	// classified-transient fault repeats identically BY DEFINITION, and it
	// repeats because the world is broken, not because the writer failed to
	// learn. Terminating those would convert every outage into a wave of
	// permanently failed items. So: release first, and only classify a repeat
	// among the failures nobody has excused.
	//
	// WHY THIS EXISTS. 345's evidence was 99 rejections in which EVERY item with
	// more than one had exactly one distinct reason (count(DISTINCT
	// md5(error_message)) = 1), from 2 repeats up to 52 on a single item. That
	// was the disconfirming test for "a stochastic writer will eventually differ"
	// and it failed to disconfirm. Candidate 1 fixed the cause by feeding the
	// refusal back, and it works — item ceea0c07 on 2026-08-22 was refused at
	// 12:18:43Z, re-dispatched at 12:51 carrying the text, and completed at
	// 12:53:07. This is the OTHER half: when the feedback has been supplied and
	// the answer is byte-identical anyway, the remaining attempts cannot help.
	//
	// ⚠ NOT the bug file's candidate 2 as written. That said "park on the FIRST
	// rejection", whose premise ("it cannot succeed unchanged") was true only
	// while the retry was blind. Implementing it now would disable the mechanism
	// that just started working. The rule here can never fire on a first
	// failure: it needs a prior recorded failure to compare against.
	//
	// EXACT EQUALITY, deliberately, and NOT normSigFragment. That signature
	// exists for burst detection and strips quoted strings, so two DIFFERENT
	// invented field names — "currency_symbol" and "ctas.primary_url", both real
	// examples from this bug — collapse to one signature and a writer that was
	// genuinely converging would be killed. The bug's own evidence used exact
	// message identity, so that is the evidence-grounded threshold, and it errs
	// toward false negatives (a volatile substring makes two "same" failures
	// differ and the rule simply stays silent), which is the safe direction.
	//
	// Compared in Go, not SQL: because the comparison is exact there is no
	// normalisation to drift, so the "applied in SQL on both sides on purpose"
	// rule above does not bind here — and Go is what makes it unit-testable.
	// ⚠ GATED ON A NON-BLANK RECORDED FAILURE, **NOT** ON attempt_count — and
	// that is this bug's own round-2 lesson, applied in the mirror position.
	// The council refuted an `attempt_count > 0` gate on the READER half (corr
	// 67b07528) with the measurement that the 52-rejection item was 52 DISTINCT
	// orchestrations while attempt_count reached only 3, because the transient
	// arm above returns an item to the queue with `error` WRITTEN and the
	// attempt deliberately NOT consumed. An attempt gate here would exempt
	// exactly those uncounted re-dispatches — the majority population — from the
	// rule meant to stop them.
	//
	// A first-ever failure is excluded anyway, and by the honest discriminator:
	// a fresh item's `error` is NULL, so there is nothing to have repeated. I
	// wrote the attempt gate first and MUTATION-TESTING removed it: deleting it
	// broke no test, which said it was redundant, and reading why said it was
	// also wrong.
	repeatTermination := false
	if envArmed(envDisableRepeatTermination) && stopOnRepeatFailure(stepConfig, itemType) {
		prev := strings.TrimSpace(priorError.String)
		if prev != "" && prev == strings.TrimSpace(errorMsg) {
			repeatTermination = true
		}
	}

	// ── The counting ladder. Backoff minutes come from reaper_policies
	// (RFC_018's invitation: "adopt reaper_policies for its numbers first,
	// executor second"), scaled linearly by attempt exactly as its first
	// consumer does (20 × (retry_count+1) in reap_stale_collection_tasks).
	backoff := 0
	// !repeatTermination: a backoff stamps "not claimable before T" on a row we
	// are about to make terminal. Harmless but incoherent, and it would make a
	// terminated row look like a scheduled retry to every reader of retry_after.
	if envArmed(envDisableRetryBackoff) && attemptCount+1 < maxAttempts && !repeatTermination {
		backoff = backoffMinutesFor(ctx, db, logger, itemType) * (attemptCount + 1)
		if backoff > maxBackoffMinutes {
			backoff = maxBackoffMinutes
		}
	}

	if repeatTermination {
		// Durable firing marker (bugs_open/345, Fable review F2b). Without
		// this, TerminatedOnRepeat lives only in the in-memory outcome and
		// the Info log below — which scrolls — so "has it ever fired?" is
		// unanswerable from the DB, and even the attempt_count<max_attempts
		// proxy expires when the archiver drains terminal rows (~7 days,
		// WII-024). Merged in Go rather than bound as a new SQL parameter:
		// a seventh positional arg is exactly the sqlmock arity trap that
		// broke two sibling test files when the sixth was added, and the
		// statement text must not vary with the flag (the 42P18 pin).
		m := map[string]interface{}{}
		if len(resultMerge) > 0 {
			// Best-effort: an unparseable caller merge must not turn a
			// terminal write into an error; the marker then rides alone.
			_ = json.Unmarshal(resultMerge, &m)
		}
		m["terminated_on_repeat"] = true
		if b, jerr := json.Marshal(m); jerr == nil {
			resultMerge = b
		}
	}

	newStatus, attemptsLeft, affected, err := writeCountingLadder(
		ctx, db, itemID, errorMsg, agentType, backoff, resultMerge, guard, repeatTermination)
	if err != nil {
		return out, err
	}
	if affected == 0 {
		out.Skipped, out.SkipReason = true, "decision_status_preserved:"+curStatus
		logger.Info("work item failure ladder: skipped — a deliberate status is already recorded, not overwriting",
			zap.String("item_id", itemID.String()), zap.String("status", curStatus))
		return out, nil
	}
	out.NewStatus, out.AttemptsLeft, out.BackoffMins = newStatus, attemptsLeft, backoff
	out.TerminatedOnRepeat = repeatTermination
	if repeatTermination {
		// Info, not Warn: this is the mechanism working. It is logged loudly
		// because it is the ONLY place the saving is visible — the row just
		// shows 'failed', exactly as it would at the ceiling, and 345's whole
		// difficulty was that a wasteful loop looked identical to a normal one.
		logger.Info("work item failure ladder: TERMINATED ON REPEAT — this failure is byte-identical to the one already recorded, so the feedback did not change the outcome and the remaining attempts cannot either",
			zap.String("item_id", itemID.String()),
			zap.String("item_type", itemType),
			zap.String("new_status", newStatus),
			zap.Int("attempts_that_would_have_run", maxAttempts-attemptCount-1))
		return out, nil
	}
	logger.Info("work item failure ladder: attempt counted",
		zap.String("item_id", itemID.String()),
		zap.String("new_status", newStatus),
		zap.Int("attempts_left", attemptsLeft),
		zap.Int("backoff_minutes", backoff))
	return out, nil
}

// writeCountingLadder is the ladder UPDATE: increment, terminal at the ceiling,
// otherwise back to the queue with a not-claimable-before stamp.
//
// claimed_by/claimed_at are cleared on re-triage. Today's ladder leaves them
// set, which makes a re-triaged row invisible to stale-work-item-reaper (whose
// predicate requires claimed_at IS NULL) — i.e. immortal in exactly the way
// bugs_closed/213's landmine describes. Clearing them is the honest state and
// matches the existing AI-unavailable branch. Stated in the submission as a
// behaviour change: retried rows become reapable at 48h.
func writeCountingLadder(ctx context.Context, db *sql.DB, itemID uuid.UUID,
	errorMsg, agentType string, backoffMins int, resultMerge []byte, guard string,
	terminateNow bool) (string, int, int64, error) {

	q, args := countingLadderStatement(itemID, errorMsg, agentType, backoffMins, resultMerge,
		guard, retryAfterColumnPresent.Load(), terminateNow)

	var newStatus string
	var attemptsLeft int
	err := db.QueryRowContext(ctx, q, args...).
		Scan(&newStatus, &attemptsLeft)
	if err == sql.ErrNoRows {
		return "", 0, 0, nil // the guard refused
	}
	if err != nil {
		if noteMissingRetryAfterColumn(err) {
			return writeCountingLadder(ctx, db, itemID, errorMsg, agentType, backoffMins, resultMerge, guard, terminateNow)
		}
		return "", 0, 0, fmt.Errorf("failed to apply work item failure ladder: %w", err)
	}
	return newStatus, attemptsLeft, 1, nil
}

// countingLadderStatement assembles the ladder UPDATE and its bind list
// TOGETHER, so no variant can reference a parameter the other side does not
// supply — or bind one the text never references.
//
// Why this is a hard rule and not tidiness: the first cut built the
// retry_after fragment by Go-side branch and bound five parameters
// unconditionally. On the TERMINAL attempt computeLadder passes backoff 0, the
// fragment collapsed to `retry_after = NULL,` and $4 vanished from the text
// while staying in the bind list — SQLSTATE 42P18 ("could not determine data
// type of parameter $4") at PREPARE time, so the terminal write NEVER ran:
// max_attempts was unreachable through this ladder in production
// (bugs_open/307 §9, live 2026-08-21, one canary + one natural fail_work_item
// hit). Fifteen sqlmock tests passed throughout — a mock matches text and
// never types a placeholder — which is why
// TestLadderStatementPlaceholdersMatchBindList audits the assembled variants
// instead of mocking them. The handled_by COALESCE comment below records the
// same discipline applied to $3; $4 was the one parameter that missed it.
//
// The zero-backoff special case lives IN the CASE ($4::int <= 0 → NULL) so the
// statement text is identical whichever value is bound.
func countingLadderStatement(itemID uuid.UUID, errorMsg, agentType string, backoffMins int,
	resultMerge []byte, guard string, withRetryAfter bool, terminateNow bool) (string, []interface{}) {

	// $4 is ALWAYS bound and ALWAYS referenced, whichever way it goes — the
	// discipline this function's header exists to enforce. bugs_open/345
	// candidate 2 adds it: when the incoming failure is byte-identical to the
	// one already on the row, the remaining attempts cannot help, so the item
	// goes terminal now. Expressed as "force the ceiling" rather than as a new
	// status: idx_swi_dedup and workItemTerminalStatuses are ONE contract, and
	// inventing a status here is a fleet-wide 42P10.
	args := []interface{}{itemID, errorMsg, agentType, terminateNow}

	retryAfter := ""
	if withRetryAfter {
		// A terminated row must not carry a not-claimable-before stamp: it is
		// not being retried, and retry_after is how every other reader tells
		// "scheduled" from "done".
		retryAfter = `retry_after = CASE WHEN attempt_count + 1 >= max_attempts OR $4::boolean OR $5::int <= 0 THEN NULL
			                 ELSE NOW() + make_interval(mins => $5::int) END,`
		args = append(args, backoffMins)
	}
	args = append(args, jsonbTextOrNil(resultMerge))
	resultParam := fmt.Sprintf("$%d", len(args))

	q := `
		UPDATE site_work_items
		SET attempt_count = attempt_count + 1,
		    -- COALESCE, not a bare assignment: bugs_closed/040 candidate 2 exists
		    -- because a literal-less error step left site_work_items.error EMPTY
		    -- while agent_error_log held the real message. An empty incoming
		    -- message must never blank an error a previous attempt recorded.
		    error = COALESCE(NULLIF($2, ''), error),
		    -- handled_by is written only when the caller HAS an identity, and
		    -- COALESCE (rather than a Go-side branch) keeps every bind
		    -- parameter referenced whichever way it goes.
		    --
		    -- update_work_item_status has never written this column, and
		    -- a NULL handled_by is consequently the documented tell that
		    -- separates its writes from fail_work_item's — bugs_open/307 §2.2
		    -- attributes the attempt-1 deaths with exactly that, and other
		    -- lanes' censuses use it too. Writing '' here would populate the
		    -- column without identifying anybody and silently invalidate all of
		    -- those queries mid-stream. The identity is not lost: that action
		    -- already stamps result.completed_by_step.
		    handled_by = COALESCE(NULLIF($3, ''), handled_by),
		    status = CASE WHEN attempt_count + 1 >= max_attempts OR $4::boolean THEN 'failed' ELSE 'triaged' END,
		    ` + retryAfter + `
		    claimed_by = NULL,
		    claimed_at = NULL,
		    result = COALESCE(result, '{}'::jsonb) || COALESCE(` + resultParam + `::jsonb, '{}'::jsonb),
		    updated_at = NOW()
		WHERE id = $1` + guard + `
		RETURNING status, max_attempts - attempt_count`
	return q, args
}

// writeTransientRelease returns the item to the queue WITHOUT consuming an
// attempt, but never without a cooldown, and records the release on the row so
// the next one can be capped. One write per failure event — never on a cadence
// (WII-018: any write bumps updated_at, and a periodic write to an open row
// makes it permanently unreapable).
func writeTransientRelease(ctx context.Context, db *sql.DB, itemID uuid.UUID,
	errorMsg, reason string, cooldownMins int, resultMerge []byte, guard string) (string, int, int64, error) {

	q, args := transientReleaseStatement(itemID, errorMsg, reason, cooldownMins, resultMerge,
		guard, retryAfterColumnPresent.Load())

	var newStatus string
	var attemptsLeft int
	err := db.QueryRowContext(ctx, q, args...).
		Scan(&newStatus, &attemptsLeft)
	if err == sql.ErrNoRows {
		return "", 0, 0, nil
	}
	if err != nil {
		if noteMissingRetryAfterColumn(err) {
			return writeTransientRelease(ctx, db, itemID, errorMsg, reason, cooldownMins, resultMerge, guard)
		}
		return "", 0, 0, fmt.Errorf("failed to release work item after transient failure: %w", err)
	}
	return newStatus, attemptsLeft, 1, nil
}

// transientReleaseStatement — statement and bind list assembled together, for
// the reason countingLadderStatement documents. The column-absent variant of
// the first cut had the same latent 42P18 (cooldownMins bound as an
// unreferenced $4), which also made noteMissingRetryAfterColumn's recovery
// unreachable: the error it waits for (undefined column) can never be produced
// by a statement that no longer names the column.
func transientReleaseStatement(itemID uuid.UUID, errorMsg, reason string, cooldownMins int,
	resultMerge []byte, guard string, withRetryAfter bool) (string, []interface{}) {

	// bugs_open/345: the stored text is the error ITSELF, with no
	// "transient (<reason>): " prefix in front of it any more. Two reasons, and
	// the second is the load-bearing one:
	//
	//  1. The prefix ate the FRONT of the reader's 2,000-char cap
	//     (load_work_item_actions.go), and the front is the end that cap's own
	//     comment says it preserves — "the guard names the offending field
	//     first".
	//  2. It made two reports of ONE failure textually different depending on
	//     which arm wrote them, so a repeat could not be recognised as a repeat.
	//     stopOnRepeatFailure below compares this column against the incoming
	//     message; with the prefix in place the comparison could never match and
	//     the rule would be silently inert.
	//
	// Nothing is lost: the reason is already recorded structurally, in
	// spec->'last_transient_release'->>'reason' below, which is where every
	// query should have been reading it from anyway.
	args := []interface{}{itemID, errorMsg, reason}

	retryAfter := ""
	if withRetryAfter {
		retryAfter = `retry_after = NOW() + make_interval(mins => $4::int),`
		args = append(args, cooldownMins)
	}
	args = append(args, jsonbTextOrNil(resultMerge))
	resultParam := fmt.Sprintf("$%d", len(args))

	q := `
		UPDATE site_work_items
		SET status = 'triaged',
		    -- COALESCE/NULLIF, matching the counting arm at writeCountingLadder.
		    -- The two arms were asymmetric: this one was a bare assignment, so an
		    -- empty incoming message could blank an error a previous attempt had
		    -- recorded — the very thing bugs_closed/040 candidate 2 exists to
		    -- prevent, prevented on one arm only. [MEASURED 2026-08-22: 0 live
		    -- rows carry the content-free form, so this is a hazard closed
		    -- before it fired, NOT a measured incident.]
		    error = COALESCE(NULLIF($2, ''), error),
		    claimed_by = NULL,
		    claimed_at = NULL,
		    handled_by = NULL,
		    ` + retryAfter + `
		    spec = COALESCE(spec, '{}'::jsonb) || jsonb_build_object(
		             'transient_releases', ` + fmt.Sprintf(jsonbIntFragment, "spec->'transient_releases'", "spec->>'transient_releases'") + ` + 1,
		             'last_transient_release', jsonb_build_object('reason', $3::text, 'at', NOW()::text)),
		    result = COALESCE(result, '{}'::jsonb) || COALESCE(` + resultParam + `::jsonb, '{}'::jsonb),
		    updated_at = NOW()
		WHERE id = $1` + guard + `
		RETURNING status, max_attempts - attempt_count`
	return q, args
}

// detectFailureBurst answers "is the fleet failing this way right now?".
//
// The row for THIS failure is already committed when we ask: routeToErrorStep
// writes agent_error_log and sets __step_error together, then advances to the
// error step — so the error step (this code) reads a log that already includes
// itself. Measured cost 0.36ms on idx_error_log_time.
//
// It keys on `domain`, not site_id: site_id is 89.8% NULL in this table over 7
// days, and 2025 of 2084 rows were NULL INSIDE the outage window itself, so a
// distinct-sites test would not have fired during the very outage this exists
// for. It keys on the normalised error_message, never error_code:
// agenterrors.ClassifyError labels the git 404 LLM_API_ERROR, which would both
// miss the burst and lump it in with unrelated LLM faults.
//
// Any error reading the log means "no burst" and the normal ladder runs. A
// detector that could block the failure write would be worse than no detector.
func detectFailureBurst(ctx context.Context, db *sql.DB, logger *zap.Logger, errorMsg string) bool {
	if strings.TrimSpace(errorMsg) == "" {
		return false
	}
	window := envInt(envBurstWindowMinutes, defaultBurstWindowMinutes)
	minErrors := envInt(envBurstMinErrors, defaultBurstMinErrors)

	q := fmt.Sprintf(`
		SELECT count(*), count(DISTINCT domain), count(DISTINCT agent_type)
		FROM agent_error_log
		WHERE occurred_at >= NOW() - make_interval(mins => $1)
		  AND %s = %s`,
		fmt.Sprintf(normSigFragment, "error_message"),
		fmt.Sprintf(normSigFragment, "$2::text"))

	var errs, domains, types int
	if err := db.QueryRowContext(ctx, q, window, errorMsg).Scan(&errs, &domains, &types); err != nil {
		logger.Warn("work item failure ladder: burst probe failed, treating as no burst",
			zap.Error(err))
		return false
	}

	fired := errs >= minErrors && domains >= burstMinDistinctDomains && types >= burstMinDistinctAgentTypes
	logger.Info("work item failure ladder: burst probe",
		zap.Bool("burst", fired),
		zap.Int("errors", errs), zap.Int("distinct_domains", domains),
		zap.Int("distinct_agent_types", types), zap.Int("window_minutes", window))
	return fired
}

// backoffMinutesFor reads the declared policy for this item type, falling back
// to the queue-wide '__default__' row and then to the code default. A policy
// row is how an operator retunes a queue without a build (RFC_018 owner
// decision 2026-08-08 #3: "a task type declares its own ceiling by inserting a
// row"), which is why this is a lookup and not a literal.
func backoffMinutesFor(ctx context.Context, db *sql.DB, logger *zap.Logger, itemType string) int {
	const q = `
		SELECT backoff_minutes FROM reaper_policies
		WHERE queue = 'site_work_items' AND item_type IN ($1, '__default__')
		ORDER BY (item_type = $1) DESC
		LIMIT 1`
	var mins int
	if err := db.QueryRowContext(ctx, q, itemType).Scan(&mins); err != nil {
		if err != sql.ErrNoRows {
			logger.Warn("work item failure ladder: reaper_policies lookup failed, using code default",
				zap.String("item_type", itemType), zap.Int("default_minutes", defaultBackoffMinutes),
				zap.Error(err))
		}
		return defaultBackoffMinutes
	}
	if mins <= 0 {
		return 0
	}
	return mins
}

// noteMissingRetryAfterColumn latches the pre-migration statement shape when
// the binary has arrived ahead of migration 502. Reported once, loudly: it
// means backoff is silently not in effect, which is exactly the "the fix looks
// live but does nothing" state this estate keeps getting bitten by.
func noteMissingRetryAfterColumn(err error) bool {
	if err == nil || !retryAfterColumnPresent.Load() {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "retry_after") &&
		(strings.Contains(msg, "does not exist") || strings.Contains(msg, "42703") ||
			strings.Contains(msg, "undefined column")) {
		retryAfterColumnPresent.Store(false)
		return true
	}
	return false
}

// jsonbTextOrNil renders an optional jsonb payload for a `$n::jsonb` parameter.
//
// The package already has nullableJSON (hitl_persistence.go) and this is NOT a
// duplicate of it: that one returns []byte, which lib/pq encodes as bytea and
// which therefore will not cast to jsonb. Changing it would touch a caller this
// change has no business touching, so the text form gets its own name.
func jsonbTextOrNil(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

// stopOnRepeatFailure reads the CALLER's opt-in for bugs_open/345 candidate 2.
//
// THE SHAPE IS AN OWNER RULING, not a preference (2026-08-02, RFC_010 §2): new
// authority on a shared seam ships as an OPT-IN FIELD whose UNSAFE default is
// OFF, because "a comment is not a control on a tree this many sessions share".
// Ending a retry budget early IS new authority, and this ladder is as shared as
// a seam gets — [MEASURED 2026-08-22] 799 items fleet-wide currently carry a
// non-blank error across ~30 item types. Absent or empty config therefore means
// today's behaviour, byte for byte.
//
// WHY IT IS A LIST OF ITEM TYPES AND NOT A BOOLEAN. The natural place to put a
// boolean would be the dispatcher's `mark_failed` step — but `component-creator`
// has no error_step of its own (read from the live agent_definitions row,
// 2026-08-22), so the item's failure is written by build-dispatch-loop's
// `mark_failed`, which is ONE step shared by every item type that dispatcher
// handles. A boolean there would silently arm the rule for all of them. Naming
// the types keeps the blast radius visible to a reviewer of the caller, which is
// the whole point of the ruling:
//
//	"mark_failed": {
//	    "action": "fail_work_item",
//	    "config": { "stop_on_repeat_failure_item_types": ["needs_new_component"] }
//	}
//
// The ladder already reads item_type for the backoff policy, so honouring it
// costs nothing extra.
//
// Unknown or misspelled type names are inert by construction: they simply never
// match. That is a deliberate silent-failure, and it is the safe direction — a
// typo disarms the rule rather than arming it for something unintended.
func stopOnRepeatFailure(stepConfig map[string]interface{}, itemType string) bool {
	if stepConfig == nil || itemType == "" {
		return false
	}
	raw, ok := stepConfig["stop_on_repeat_failure_item_types"]
	if !ok {
		return false
	}
	// []interface{} is what JSON config arrives as; []string is what a Go
	// caller or a test would pass. Both are accepted so the seam behaves the
	// same whichever side reaches it.
	switch v := raw.(type) {
	case []interface{}:
		for _, e := range v {
			if s, ok := e.(string); ok && strings.TrimSpace(s) == itemType {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if strings.TrimSpace(s) == itemType {
				return true
			}
		}
	}
	return false
}
