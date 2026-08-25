# HANDOFF — vigilant designer + offer analyser (2026-08-25)

**COLD-START = this file + `bugs_open/395` + register `CLM-024` + `features_open/030` §10 + `features_open/034`.**
**This supersedes `HANDOFF_2026-08-24b_continue_here.md`** (still correct on history; out of date on
everything about v2(d)'s state, which is now LIVE).

> **Re-run every liveness claim here before acting.** This branch takes hundreds of commits a day.
> Verify against `git archive <resolved-sha>` — never the working tree (another lane is often
> mid-edit and it may not compile: on 2026-08-24 `platform/livespec/livespec.go` was dirty and broke
> `cmd/config-key-audit`'s test build) and never the moving name `HEAD` (HEAD moved through at least
> six commits *inside* the last session here).

## The one-line state

> **v2(d) is LIVE, exercised, and it caught a false green on its first run — which is now
> `bugs_open/395`.** A finding can carry a refute-only machine-checkable half of its own acceptance
> test; `verify_acceptance_predicates` decides whether one may be stored. Nothing automated READS a
> stored predicate yet, and the consumer that should is 395's fix candidate 1.

## What is DONE — do not re-take any of this

| | state |
|---|---|
| the gate action + exported `EvaluateAcceptancePredicate` | **LIVE** on the chassis that rolled 2026-08-24 (~22:0xZ), capability-probed PRESENT on both replicas with a positive and a negative control |
| migration `601_offer_analyser_acceptance_predicates_HOLD.sql` | **APPLIED BY HAND** 2026-08-24 22:0xZ, all guards passed, re-read independently afterwards, ledger row written |
| council | **APPROVED**, corr `ef482d1c-b36d-40c0-a40c-772656116016` (14 seats, none high; round 1 killed by a roll and re-fired on the same trail). Objections acted on in `ccb35e74d` |
| first live run (webdesign.co.uk, corr `4caba084`) | `checked 4 · kept 3 · rejected 0 · subjects_loaded 137` — three predicates written UNPROMPTED, all refuting; one finding left bare, so the **silence arm fired first time** |
| the passthrough on the shared write | **proven at the artefact** — all three predicates reached `site_work_items.spec` intact, through `findingsFromList`'s hand-written map path |
| `bugs_open/395` | **FILED** — an item closes `complete` while its own criterion is unmet, with a machine verdict |
| `016b` §9 | the transferable pattern added (a status records that the HANDLER succeeded) |
| register `CLM-024`, `LANDMINES` (`pages.in_header`, roll-kills-council recurrence), `WRONG_CALLS` | all written |
| the optional-key budget cron's literal | mine recorded, and a **sweep** of two other lanes' drift, applied and verified at the live configmap |
| `SUMMARY_2026-08-24_the_tests_become_checkable.md` | written (new file, ninth in the series) |

## What the next session should do

### 1. `bugs_open/395` — the completion-time consumer. This is the work.

The producer half is live and its first run produced a machine verdict that a closed item's criterion
is false. **Nothing reads it.** 395 §4 lists the fix candidates ordered by what makes the bad state
unrepresentable; candidate 1 (evaluate the predicate at completion, refuse or flag) is the one this
lane has been pointing at since `complete_work_item_no_change.go`'s comment was written.

⚠ **Three things to carry into that design, all learned the hard way:**
- It belongs **beside `handlerReportedFailure` / the `noChangeGates` roster**, not in a verifier:
  `verifyBeforeComplete`'s `VerifyTarget` carries the SPEC, not the RESULT, so a verifier grades the
  row's PREVIOUS value while appearing to work (213's own comment records this).
- **Opt-in per `item_type`, unsafe default OFF** (2026-08-02 ruling). A refused completion is a live
  behaviour change on handlers other lanes own.
- **There is no negative control yet.** All three live predicates refute; no row exists where a
  predicate is satisfied after its fix. A gate that has only ever seen failures cannot be
  distinguished from one that refuses everything — so *manufacture* the control (fix one page's meta
  by hand, or re-run after a real repair) before calling it proven.

### 2. Watch the second and third runs — the numbers that are still n=1

Everything about adoption rests on ONE run. Re-read it as runs accumulate:

```sql
SELECT s.domain, wi.created_at, wi.status,
       wi.spec->'acceptance_predicate'->>'type'                AS pred,
       wi.spec->'acceptance_predicate'->>'verdict_at_emission' AS verdict,
       wi.spec->'acceptance_predicate_rejected'                AS refused
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
 WHERE wi.spec->>'audit_source'='offer-analysis' AND wi.created_at > '2026-08-24 22:00:00+00'
 ORDER BY wi.created_at DESC;
```

⚠ **`rejected: 0` means the REFUSAL arm has never fired in production** — CLM-023's residual in a new
place. It is proven by mutation-tested units, not in the wild. **Never quote a clean run as evidence
the refusal works**, and note that a run where the model writes no predicate passes the gate
trivially. That mistake cost this lane two days on 537.

### 3. The rest of the v2 batch — write `602` for (a)+(b)+(c), config-only

v2(d) shipped alone because it needed Go and its migration had to be held for a roll; the others do
not. The batch's real goal — ONE re-proof — is still available: write **602** for v2(a) + v2(b) +
v2(c), apply it, then fire **one** run. Traps unchanged, in `features_open/030` §10:
- **v2(a)** bounded head-of-hero excerpt per page — ⚠ GROWS the surface; re-run the truncation check
  on webdesign.co.uk afterwards (the `__truncated`-absent-at-104-pages baseline is v1's). **It also
  widens what a predicate can address** — body-text shapes are excluded today precisely because the
  surface carries no content.
- **v2(b)** attribution in the `why` clauses — partly done by 537 (points only); re-read the live
  prompt first.
- **v2(c)** `primary_model` in the degraded arm's field list — LATENT, no live instance; must not be
  the reason to open the batch, and do not fix it by letting the model *infer* one.

### 4. `features_open/034` — claims audit over `site_specs` prose

Owner-approved 2026-08-14, still not designed. Today's work checks whether a page matches what we
said about it; 034 asks whether what we said was true.

### 5. Coverage — 5 of 28 sites, and the denominator moves

`[MEASURED 2026-08-24]` five sites carry `offer_ordering`, out of **28** live sites all of which have
pages. The last summary said "five of twenty-three": coverage flat, estate grew by five. ⚠ **Do not
carry a site count forward from any document** — re-run it.

## Watch-outs this lane has now paid for

- **⚠ CONFIRMING THAT THE PROMOTER YOU THOUGHT OF IS OFF IS NOT A SAFETY ARGUMENT.** I checked
  `improvement-sweep.enabled = f` and fired a single dispatch believing it could not change live
  pages. `build-pipeline-trigger` → `build-dispatch-loop` promoted the findings **31 seconds** later
  and `page-build-handler` rebuilt and deployed the index page. `fire-offer-analyser.sh`'s header is
  true of the SCRIPT and silent about what else is running. **The cheap check is the inverse query —
  `SELECT name, enabled FROM scheduled_tasks WHERE enabled` — because a `WHERE` clause naming your
  own suspicion cannot discover a second cause.** Full entry in `WRONG_CALLS.md`. ⚠ Its entry was
  **swept into another lane's commit `d7528dc53`**, so `git log` on my commits will not find it.
- **⚠ `pages.in_header` IS NOT THE RENDERED NAV** — 13 rows vs 7 served destinations on leopardess. A
  column-based nav check "found" a false green that `curl` refuted. Nav is out of the predicate
  vocabulary for this reason, stated in the action's own file header. `pages.rendered_header` is not
  the escape route: `''` on all 35 active pages of robot-hands.com. `LANDMINES.md`.
- **⚠ A ROLL KILLS AN IN-FLIGHT COUNCIL, and a lone casualty is the EXPECTED shape.** Round 1 froze
  at `review_guardian` 18:30:18Z with new pods at 18:32Z. I went looking for other frozen runs as
  corroboration and found **one: mine** — which reads as "the roll was not the cause". The pod-age
  comparison is the check; a peer census is not a second opinion. Appended to the landmine.
- **⚠ `run-migrations.sh --record-only` REFUSES a `_HOLD` file** as an uppercase sidecar, so the
  supported path cannot record the one class of migration that is ordering-critical. Recorded by hand
  INSERT, matching 575/526/417. The runner's `SIDECAR_RE` predates the `_HOLD` convention; a one-line
  narrowing would fix it and nobody owns it.
- **⚠ MIGRATION NUMBER COLLISION: there are TWO 601s.** `601_claims_auditor_page_text_strips_per_component.sql`
  (bugs_open/380 lane, applied 19:00Z) and mine, `601_offer_analyser_acceptance_predicates_HOLD.sql`.
  Both are in the ledger, which keys on FILENAME, so nothing is lost — but **resolve by slug, never by
  number** (597 is double-used too). Do NOT renumber mine: it is applied and recorded under this
  filename, and renumbering would make it look pending and get it replayed.
- **⚠ `site_work_items` has no `audit_source` COLUMN** — it is `spec->>'audit_source'`, and the column
  form ERRORS rather than returning zero, so behind a `2>/dev/null` it reads as "no findings".
- **⚠ `llm_call_log.agent_type` is the DISPATCH context** — a hand-fired run lands under `'generic'`.
  Key on `step_name='run_offer_analysis'`.
- **⚠ `orchestration_states` has no `agent_type`** — it is `owner_agent_type`, and the wrong name
  errors. A zero from the corrected query still is not evidence: terminal rows are reaped in ~24-48h.
- **⚠ psql prints UTC, your shell prints BST** — always toward alarm. Make the DATABASE subtract.
  (This session lost thirteen hours to a `now() - interval '30 minutes'` filter written before an
  overnight gap: the four items it "could not find" were there all along.)

## Residuals, stated plainly

1. **The refusal arm has never fired** (`rejected: 0`). Units only. See §2.
2. **Nothing automated reads `acceptance_predicate`.** That is `bugs_open/395` and it is §1.
3. **The truncation asymmetry, unmeasured:** the model authors predicates against a meta description
   truncated at **160 chars** in the offer surface; the evaluator reads the FULL column. A needle past
   char 160 is visible to the gate and was not visible to its author. No live instance yet.
4. **The original code commit `7b875b08f` carries no council trailer** (the token had expired when it
   was made), so it lists as un-reviewed in `098` for ever; `ccb35e74d` carries
   `Council-Reviewed: ef482d1c-…`. Forward-only forbids the amend.

## Who owns what nearby

The **leopardess lane** holds five of this lane's findings at `needs_human_review` pending an owner
design report — **coordinate before firing B4 at that site, and note the promotion trap above: filing
findings there would dispatch handlers at work they are holding.** `bugs_open/333` belongs to the 301
lane. `copy_quality_two_stage` + the LMC lane still work loanandmortgagecalculator.co.uk. The
**`bugfix_308_cta_destination_provenance`** lane has routed the undecidable-CTA question to this agent
(`CONTRIB_2026-08-24`, owner ruling in `RFC_047` §10); its own read is *"after your v2 batch"* and
nothing is blocked on us.
