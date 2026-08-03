# NOTES — bugs_open/091 candidate 1

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-02 23:45 — picking the bug, and how much of `bugs_open/` is actually taken

Swept all 55 files in `bugs_open/`. `scripts/who-owns.py` returns "OWNED or
recently active" for almost everything (it is deliberately conservative), so the
discriminating check was the second one: **grep the 19 live `.jsonl` session
transcripts for `bugs_open/NNN`** — `who-owns` reads COMMITS and is blind to a
session that is mid-fix with nothing committed yet.

Only three files had no owning workstream AND no mention in any live transcript:
**080** (residual data repair, site-specific), **091**, **158** (item 1 is
explicitly an RFC, item 3 needs an owner decision — not closeable this session).
091 it is: its remaining candidate is a change to a helper every detector in the
fleet calls, which is the "robust and applicable to the framework" shape.

Rejected along the way, with reasons worth keeping:

* **093** — looked ideal (council escalated it twice). It is **blocked on
  `bugs_open/083`**, not on code: the fix shipped in v1.0.1172 and has never
  executed, because the only thing that runs it is `improvement-sweep`, disabled
  since 2026-05-02. A bug can be "fixed, live, and still open" for want of a
  cadence, and the file says so — read to the end before starting.
* **122** (WCAG) and **087** (page-rebuild section plan) — both actively owned
  (dartsonline_traffic, gemini_content_provider). 087 in particular reads as
  unowned at the top and has three dated sections of another lane's work below.
* **134** — real but latent (the agent has never run) and one instance fleet-wide.

## 2026-08-02 23:55 — the measurement that resized the bug

The file rates this Medium and says "a delay, not a loss". Measured against the
live DB it is worse: `evidence-freshness` is **enabled** and ran at 18:36:07Z
today, and **four of the five open `stale_evidence` items name the wrong facts**
(table in the PLAN; query in the RUNBOOK). leopardess's item names
`C4-orchestration-state-records` while the fact that is actually drifting is
`C4-agent-definitions-catalogue` — a different fact entirely. vonc's item
describes drift that no longer exists.

Three query missteps, all recorded in the RUNBOOK rather than here because they
will recur: `orchestration_states` has no `agent_type` column (it is
`owner_agent_type`); `jsonb_array_elements` on a scalar aborts the whole
statement, so a missing type guard reads as "no rows" rather than as an error;
and the table is retention-clocked, so exactly ONE `evidence-freshness` run
exists at any moment and the comparison is only possible on the same day.

## 2026-08-03 00:30 — the design turn: `DO UPDATE` would have re-created the bug

First sketch followed the bug file literally: `ON CONFLICT … DO UPDATE SET spec,
summary, updated_at`. Writing the test for it is what caught the problem.
**`DO UPDATE` affects a row, so `RowsAffected()` returns 1** — and
`insertWorkItem` returns `rows > 0`, which `work_item_created` is set from. The
literal fix would have made `work_item_created` start reporting `true` for a
write that created nothing: **091's own defect, arriving inside 091's fix.**

So: a separate `UPDATE` in the conflict branch (the shared INSERT stays
byte-identical for ~20 callers), and a three-state outcome instead of a bool.

Second turn, same shape: `refreshOnConflict` was going to be a **field on
`workItem`**, exactly as the bug file proposes. But then a caller can set the
field and still call `insertWorkItem`, whose single bool cannot express a refresh
— a silent wrong answer, at the one call site most likely to be copied. Made it a
**parameter of a new `writeWorkItem`** instead, so `insertWorkItem` cannot receive
it and the mistake does not compile. This is a deliberate deviation from the
filed candidate and the reason is written into the code.

## 2026-08-03 00:50 — MISSTEP: I widened a shared statement and broke twenty tests I had never opened

Added `parent_item_id` to the shared INSERT unconditionally. `go build` was
clean. **The package test suite then failed 20 tests across 8 files** —
`save_sections_prune_floor_test.go`, `tool_render_path_test.go`,
`page_role_upsert_test.go`, `nav_rebuild_request_test.go` and others — because
sqlmock matches the **argument count**, and every one of those expectations lists
16 args positionally. None of those tests has anything to do with this bug, and
several are in lanes other sessions are working right now (175, 178).

The fix was not to edit twenty files. It was to notice what the failure was
telling me: **a shared statement widened for one caller charges every caller.**
`parent_item_id` is now appended as `$17` only when a parent is actually set, so
a caller that never asked for it sends the identical statement it always sent,
and the only test files touched are the three whose behaviour genuinely moved.

Two lessons, the second more useful than the first:

1. `go build` says nothing about a shared SQL statement's blast radius. The test
   suite is the only thing that measures it, and it must be the WHOLE package.
2. **The size of a test breakage is a measurement of the seam, not an obstacle to
   it.** My first instinct was "update the twenty expectations"; that would have
   shipped a real widening and made three other lanes' files carry my change.

## 2026-08-03 01:10 — the guards are MUTATION-PROVEN, and one of them found a blind spot

Four mutations, each run and each confirmed to fail the suite (a guard that is
merely present is not a guard — a test that passes with the rule removed was
testing nothing):

| mutation | test that caught it |
|---|---|
| policy check removed → default policy refreshes | `DropOnConflict_IssuesNoUpdate`, `InsertWorkItem_CannotRefresh` |
| held-status clause dropped from the refresh predicate | `RefreshStatement_GuardsTerminalAndHeldRows` |
| a refresh reports `Inserted: true` | `RefreshOnConflict_UpdatesTheOpenItem` |
| `recurrenceExpected` cleared on gap-plan items | `GapPlanWorkItem_IsRecurrenceExpected` |

**The fourth one exposed a blind spot in the harness that is worth more than the
mutation was.** Clearing `recurrenceExpected` failed the direct unit test but
**not** the three behavioural `applyNewPage`/`applyRetypeExisting` tests — and it
should have, because with the flag off `insertWorkItem` issues an anti-churn
`SELECT COUNT(*)` those mocks do not expect.

Cause, read rather than guessed: `load_work_item_actions.go:1250` is
`if err == nil && terminalCount > 0` — **the probe's error is swallowed by
design**. So an unexpected query returns an error, the error is discarded, and
the test passes. Consequence, stated plainly because it will mislead somebody
else: **no behavioural sqlmock test in this package can detect a change to
`recurrenceExpected`.** The gap-plan adoption is therefore covered by a direct
assertion on the built `workItem`, not by the behavioural tests, which are blind
to precisely the thing most likely to go wrong in that adoption. Filed as a
landmine and named in the submission's own `risks` block.

This is the [[a-mutation-that-passes-may-have-hit-a-guard-in-series]] shape with
a twist: the mutation passed not because a second guard caught it, but because
the *observer* was deaf. "No test failed" was not evidence.

## 2026-08-03 01:25 — submitted to the council

`SUBMISSION_CORR = 8e7357ae-9f8d-49bf-81c0-669d9a97a205`, 7 edits. The `risks`
block carries the measurements rather than asking the reviewers to take them —
including the harness blind spot above, and the one judgement I actively want
challenged: `needs_human_review` is deliberately NOT in the held list, so an item
a human is mid-way through reading can change under them. It is a queue, not a
claim, and the alternative is leaving the record false — but that is a judgement,
not a measurement, and it is stated as one.

## 2026-08-03 01:55 — council APPROVED (13 reviewers, 0 unreadable, 6 advisory objections)

`decided_by: approved with 6 advisory objection(s) — none high-severity`. Four were
checkable, so they were checked rather than banked. **Read the report by
CORRELATION, not by `doc_notes … ORDER BY created_at DESC LIMIT 1`** — that
returned another lane's REVISE verdict for a completely different submission
(`4a7f0877`, the unpublish primitive), which for a few seconds read as *my*
verdict. The runbook query is `diagnosis_artifacts WHERE correlation_id=…
AND kind='council_report'`.

### Acted on

**1. "Is the negative test vacuous?" — three seats independently (editquality,
guardian, debug_historian).** All three cited the landmine I filed myself: a test
asserting a query is NOT issued passes vacuously against `insertWorkItem`, because
the anti-churn probe swallows the mock's error. **They were right to ask and the
answer is no — proven by mutation, not by reading.** Forcing the default policy
down the refresh path fails the test with:

```
writeWorkItem: refresh failed for stale_evidence:…: all expectations were
already fulfilled, call to Query '…UPDATE site_work_items…'
```

i.e. `refreshOpenWorkItem` **propagates** its error (only `sql.ErrNoRows` means
"nothing matched"), so an unexpected query surfaces as a returned error and the
`t.Fatalf` fires. The distinction from the probe is exactly the `err == nil` test
one swallows and the other does not. Written into the test as a comment with the
mutation recorded, so the next reader does not have to re-derive it.

**2. "The silent no-op depends on one caller remembering to check `Recorded()`"
— bug_historian.** Sustained, and it is 091's own shape one level down: a shared
mechanism whose correctness depends on each adopter remembering something. The
`Warn` moved INTO `refreshOpenWorkItem`, so the fail-loud surface belongs to the
writer. A caller may add its own; it can no longer be the only one.

**3. "Parameterise the status lists, do not interpolate" — constitution.**
Sustained. `sqlInList` exists only because `insertWorkItem`'s `ON CONFLICT … WHERE`
predicate MUST be literal for partial-index inference to resolve `idx_swi_dedup`.
The refresh is a plain `WHERE` and has no such constraint, so the exemption never
extended to it — I had simply copied the sibling. Now `status <> ALL($5::text[])`,
with the literal built the way `depends_on` already is. PREPAREd **and EXECUTEd**
against the live schema (0 rows, as expected) — a PREPARE alone would not have
proven the array literal binds.

**4. "Log when a refresh lands on a `needs_human_review` row" — guardian, echoed
by architecture.** Done. This is the one judgement I flagged as wanting challenge,
and the seats' answer was "not blocking, but make it observable rather than
documented". It now says so every time it happens.

**5. "'368, still growing' contradicts your own drop from 380" —
prior_art_librarian.** Caught a real inconsistency in my submission. Re-measured:
368 open, **50 raised in 7 days, 2 in 24h**. The queue is accumulating at ~7/day
and the net fall is drain outpacing intake. "Growing" was wrong about the count and
right about the intake; the correction is in the bug file, marked.

### Not acted on, deliberately

**tooling_provenance (low): no `doc_notes` row capturing the decision.** It is
covered — the two `LANDMINES.md` entries are synced into `doc_notes` by
`landmines-sync.py --apply` (run), which is the ruled system of record (D10). Writing
a second hand-authored row is explicitly forbidden by that ruling.

**guidelines seat flagged a GUIDELINE GAP rather than objecting:** the documented
work-item dedup rule says "use DELETE+INSERT, not ON CONFLICT", which the entire
existing helper contradicts. Not this change's to resolve — but it is the rule the
`a5b70424` guidelines seat cited against `apply_gap_plan`, so the rule and the
platform have been out of step for a while. Worth someone's RFC; noted, not taken.

## 2026-08-03 10:09 — round 2 live on v1.0.1238, and a NUMBER COLLISION on 184

All four council refinements shipped. Both replicas: `FINDING NOT RECORDED` → 1,
`the row is in the HUMAN-REVIEW queue` → 1, `refreshed the open work item` → 2
(it gained the human-review arm), negative control still 0. Nothing owed on 091.

**`184` is now an ambiguous number.** I filed
`184_…three_more_detectors_key_per_site_over_a_per_item_finding.md` at 09:59Z; the
`mortgagecalculator_couk_adoption` lane filed
`184_…llm_markdown_reaches_the_page_as_literal_asterisks.md` at 10:26Z (`905895069`).
Both stay — numbers are never reassigned and the repo already carries six such
pairs. Flagged in my file's header. **Resolve by slug; `git log` the FILE PATH.**
Worth noting *how* this happens: both of us read "next free number" off the same
directory listing within half an hour, and there is no allocation step. The
convention absorbs it, but the cost lands on every later reader of a commit message
that says "184".

## 2026-08-03 10:40 — 184: the measurement inverted what I had filed

This is the entry worth keeping from the 184 work.

**When I filed 184 I implied the three siblings were dropping findings now.** I had
not measured them — the file says so (`unmeasured on these three — that is the first
job`), which is the only reason the claim was not a WRONG_CALLS entry. Measuring:

- **`stale_directory_claim`: the daily sweep checks ZERO claims.** Not broken —
  `loadDueDirectoryClaims` selects on `verified_at < now() - staleness_days`, and of
  97 current claims **none is due until 2026-08-23**. Read the predicate before
  concluding a silent mechanism is broken ([[zero-adoption-means-read-the-mechanism]]).
- **And the conclusion INVERTS the usual one.** A dated exposure is normally an
  argument to defer. Here it is the argument to fix now: the batch lands in three
  weeks, the July row will still be holding the key (nothing works
  `needs_human_review` — 033, 368 parked), and the only run that can exercise this
  for real would be lost. **"Wait for the symptom" is wrong when the symptom is
  scheduled and singular.**
- **`directory_citation_unverified` is [UNMEASURED AND UNRECOVERABLE].** Same 15
  rejects listed since 07-24 across every weekly sweep since. Whether a later sweep
  found a different set cannot be established: 24h retention, and a rejected
  *candidate* never reaches `directory_claims`, so a dropped finding leaves no trace
  in any table. Marked rather than guessed — and it is the argument for the refresh,
  since it is what makes the next one observable.

**The judgement 091 flagged came out the other way here, and that is the point of
making it per-site.** 091's least-certain call was that a refresh can rewrite a row
under a human mid-read. For these three that concern is *weaker*: `bugs_open/033`
establishes the queue has no working surface, so there is no reader — and what the
old behaviour protects is a description that is already false. Had I switched all
four in one go at 02:00 without measuring, I would have got the right answer for the
wrong reason, and would not have known which.

## 2026-08-03 10:55 — MISSTEP: a test named for three call sites that never called them

Full entry in `WRONG_CALLS.md`. In brief: the first `hitl_refresh_adoption_test.go`
called `writeWorkItem` directly with `refreshOnConflict` and asserted the outcome —
i.e. it re-tested the helper 091 had already proven, while its name and header
claimed it covered the emitters. Reverting a call site did not fail it.

**Two things nearly hid it.** The suite was already red from an unmatched
`ExpectCommit`, so "mutation fails" and "mutation does nothing" looked identical —
a mutation result is only evidence against a **confirmed-green** baseline. And the
test *read* well: a table of the three item types, their keys, their spec fields, a
paragraph on why the defect is invisible. All accurate; none of it executed.

Rewritten to drive `createCitationFailuresItem`,
`createDirectoryCitationFailuresItem` and `createStaleDirectoryClaimItem` directly,
then mutation-proven one call site at a time — revert one, exactly one test fails.

## 2026-08-03 ~12:00 — v1.0.1239 built and proven; the deploy is blocked on a permission

**The pod-grep on v1.0.1238 came back exactly as the handoff predicted**, which is
worth stating because it is the discriminating result and not merely a negative:

```
citation-reject   (NEW): 0    directory-reject (NEW): 0    dir-freshness (NEW): 0
refreshed open wi (CTL): 2    FINDING NOT RECORDED (CTL): 1
```

Both replicas, identical. The two controls reading their predicted values is what
makes the three zeros mean "184 did not ship" rather than "the grep was wrong" — a
bare zero cannot tell those apart (`bugs_open/153`).

**Then the same grep against the image I had just built, BEFORE pushing it**, which
is the check the `debug_historian` seat asked for and the one that would have caught
the false marker in the first draft of the handoff:

```
citation-reject (NEW): 1   directory-reject (NEW): 1   dir-freshness (NEW): 1
refreshed open wi (CTL): 2   FINDING NOT RECORDED (CTL): 1
'the bugs_closed/091 class' (NEGATIVE CONTROL): 0
```

The negative control is the retired marker — a Go **comment**, which never reaches a
binary. It reads 0 in an image that definitely contains the fix, which is exactly
why it was useless as evidence and why `600bd99a8` replaced it.

`v1.0.1239` is built and **pushed** (digest `sha256:a1ed28ec…50e6a5`). The overlay is
updated to `newTag: v1.0.1239`. **`kubectl apply -k` was refused by this session's
permission classifier**, so the roll itself is owed to a session that can run it.
Deliberately NOT done via `make deploy-agents`: that target seds *every* agent overlay
to `$(IMAGE_TAG)` and applies all fourteen, and only `agent-chassis` exists at 1239 —
it would have put thirteen services into ImagePullBackOff to roll one.

## 2026-08-03 ~12:00 — the `prior_art_librarian` objection is ANSWERED, and the handoff's own query would have misread it

The handoff listed this as "owed but NOT done". **It was in fact already answered in
`evidence_citations.go:421-437`** by the fix commit itself; the handoff's "owed"
section is stale. Re-verified here first-hand rather than taken from the comment.

> **CORRECTED 2026-08-03 — the paragraph above ("The judgement 091 flagged came out
> the other way here") still says `bugs_open/033` establishes the queue "has no
> working surface, so there is no reader". That is FALSE as a general statement**, and
> the council was right to refuse it as an absence sourced from another bug file.
> Measured today: of **431** rows at `needs_human_review`, **86 have been claimed and
> 55 handled**, newest claim 2026-08-02. A surface exists.

**The handoff's supplied query cannot distinguish residue from a live reader, and
read naively it overturns the downgrade it was meant to test.** `claimed_by` persists
across status transitions, so "currently `needs_human_review` AND `claimed_by` not
null" says nothing about *when* the claim happened. The discriminating query:

```sql
SELECT count(*) FILTER (WHERE claimed_at IS NOT NULL AND updated_at > claimed_at) AS updated_after_claim,
       count(*) FILTER (WHERE claimed_at IS NOT NULL AND updated_at <= claimed_at) AS claim_is_newest
  FROM site_work_items WHERE status='needs_human_review';
-- 86 | 0
```

**Zero rows have the claim as their newest event.** Every claim precedes the row's
arrival at this status — it is a claimed→handled→parked history, not a live reader.

Two independent confirmations, one structural and one empirical:

- **Structural, and the stronger of the two:** `claim_work_item_action.go:102` and
  `load_work_item_actions.go:632` both claim on `status IN ('triaged','approved')`.
  `needs_human_review` is *unreachable* by the dispatch loop — not merely quiet.
- **Empirical:** of the four item types that opt into `refreshOnConflict` — 8 rows all
  told — **0 claimed, 0 handled**. Matches the comment's figure exactly.

So the downgrade is **confirmed**, on a narrower and better warrant than the one first
written. The `claimed_by`/`handled_by` values are `build-dispatch-loop` and
`page-build-handler` — both automated, neither a human surface.

**BATCH-005's open question stands unchanged**: if a HITL surface ever works these
four types, `needs_human_review` should become a held status. Nothing today does.

## 2026-08-03 ~19:30 — LIVE on v1.0.1243 via the owner's release flow; 184 CLOSED; and the self-review

**Deploy procedure corrected by the owner.** The single-overlay `kubectl apply -k` this
lane attempted (and the classifier blocked) is NOT the estate's procedure. Releases are
whole-fleet on one tag: `make release redeploy-agents ENVIRONMENT=production
REGION=uk001`, run by the owner. My stated reason for avoiding `make deploy-agents` —
"thirteen services into ImagePullBackOff" — was correct only for the fragmented state I
had created (one service built at 1239); `make release` avoids that state entirely by
building and pushing ALL backend images at the tag before any overlay moves. The 1239
image is an orphan in the registry: pushed, never deployed, superseded. Harmless.
My commit `aec57bcee` records 1239 in the makefile and overlay; the release flow's
1243 supersedes both on the same lines (uncommitted tree state, the norm here).

**Verification on v1.0.1243, both replicas:** three markers 1/1/1, controls 2 and 1,
retired-comment negative control 0 — identical to the pre-push image grep. `4c3a968cc`
and `600bd99a8` confirmed ancestors of built HEAD; the five relevant files untouched
between the unit-proof (clean `git archive`) and the build, so the proof carries.
**184 (three_more_detectors slug) moved to `bugs_closed/` with the closure block**,
including a visible correction: the bug file's own owed-list still instructed the
retired comment-marker as THE discriminating check — the handoff got corrected on
08-03 morning, the bug file did not. Fixed at closure.

**Self-review of the ~12:00 claims — two sharpened, none overturned:**

1. **The temporal query was presented one notch stronger than it is.** "0 rows have
   the claim as the newest event" proves the claim is never the LAST thing that
   happened; strictly it cannot exclude claimed-at-this-status-then-updated-again.
   The structural audit is the conclusive leg, and at noon it was SAMPLED, not
   exhaustive — two Go sites cited from a comment in `apply_gap_plan_action.go`.
2. **Now exhaustive.** Every `claimed_by` writer on `site_work_items`, from a bare
   repo-wide grep (my first grep filtered on `update|set|insert` and MISSED
   `claimed_by = $2` lines — the SET keyword sits on a different line; redone
   unfiltered) plus a fleet-wide `agent_definitions` config scan:
   `claim_work_item_action.go:99` and `load_work_item_actions.go:632` claim on
   `('triaged','approved')`; `report-dispatch-loop` config SQL claims on
   `'awaiting_report'`; `diagnose-dispatch-loop` on `'awaiting_diagnosis'`; every
   other site NULLs the column or writes `maintenance_queue` (a different table —
   `maintenance_actions.go:874` was a false hit). `needs_human_review` is
   unreachable by every claimer. The noon conclusion stands on better ground.

**A sharpening for BATCH-005, found by the audit:** `reporting`, `awaiting_report`
and `awaiting_diagnosis` are claimed-adjacent statuses NOT in `workItemHeldStatuses`
(`{claimed, diagnosing}`). Safe today only because no refresh-adopting item type
flows through the reports or diagnose pipelines — the same shape as the
`needs_human_review` open question, and it should be answered by the same ruling
when one comes. Recorded on the 184 closure.

**Still owed by the calendar, not by this lane's hands: 2026-08-23**, the first run
that can behaviourally prove `stale_directory_claim`. Same-day check required
(~24h `orchestration_states` retention).
