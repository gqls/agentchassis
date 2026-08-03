# CONTINUE HERE — the work-item conflict-refresh lane (091 closed, 184 in HEAD)

> **DONE 2026-08-03 ~19:30Z — both bugs are now CLOSED and LIVE.** 184 shipped on
> **v1.0.1243** via the owner's release flow (`make release redeploy-agents
> ENVIRONMENT=production REGION=uk001` — the whole fleet on one tag; that is the
> procedure, not a single-service `kubectl apply -k`). Pod-grep passed on both
> replicas: markers 1/1/1, controls 2/1, negative control 0. 184
> (three_more_detectors slug) moved to `bugs_closed/` with the closure block. The
> §"Owed but NOT done" objection below is ANSWERED — exhaustively, in
> `NOTES_workitem_conflict_refresh.md` (2026-08-03 ~19:30 entry): every
> `claimed_by` writer enumerated, none can claim `needs_human_review`.
> **The ONE remaining action is calendar-bound: §3 below, 2026-08-23.**

**Written 2026-08-03 ~11:00Z.** This is the *current state and next action*. The full
account is in `NOTES_workitem_conflict_refresh.md` (technical log),
`README_where_we_are.md` (plain prose), `SUMMARY_2026-08-03_*.md` (the milestone
read-out). Read this file first; it is short on purpose.

---

## TL;DR

| | state |
|---|---|
| `bugs_closed/091` | ✅ **DONE.** Fixed, live on **v1.0.1237**, refinements live on **v1.0.1238**, verified against real dropped findings, council `8e7357ae` **APPROVED 13-0**. Nothing owed. |
| `bugs_open/184` (this lane's) | **FIXED IN HEAD, NOT LIVE.** Commits `4c3a968cc` + `600bd99a8`. Council `d6cda33d` **APPROVED**, 3 advisory objections, all acted on. |

**The one thing that is genuinely owed: a roll, then a pod-grep, then close 184.**

---

## Correlations and commits

```
SUBMISSION_CORR (091)  8e7357ae-9f8d-49bf-81c0-669d9a97a205   APPROVED 13-0, 6 advisory
SUBMISSION_CORR (184)  d6cda33d-1e5a-4ea3-8ddc-98f0a6848e35   APPROVED, 3 advisory, all acted on
```

Read a verdict **by correlation**, never `doc_notes … ORDER BY created_at DESC LIMIT 1`
— that returns whichever lane wrote last and it briefly convinced me another lane's
REVISE was mine:

```sql
SELECT metadata->>'decision', metadata->>'unreadable', metadata->>'reviewers'
  FROM diagnosis_artifacts
 WHERE correlation_id='d6cda33d-1e5a-4ea3-8ddc-98f0a6848e35' AND kind='council_report';
-- full objections are in `body` (JSON), same row
```

Key commits: `6468a2746` (091 fix) · `7f85873b9` (091 council round) · `4b2cfe047`
(091 closed + 184 filed) · `4c3a968cc` (184 fix) · `ef3d04e2f` (184 docs).

## ⚠ `184` IS AN AMBIGUOUS NUMBER

Two unrelated bugs. **Resolve by slug, and `git log` the FILE PATH, not the number.**

- **this lane's:** `184_HANDOFF_2026-08-03_three_more_detectors_key_per_site_over_a_per_item_finding.md`
- the mortgagecalculator lane's: `184_…llm_markdown_reaches_the_page_as_literal_asterisks.md` (`905895069`)

## What is owed on 184, in order

### 1. The roll, then the pod-grep (do not skip the controls)

> **⚠ CORRECTED 2026-08-03, by the council's `debug_historian` seat, before anyone ran
> it.** This block first named `'the bugs_closed/091 class'` as the marker. **That is a
> Go COMMENT and never reaches the binary** — it would have greped 0 for ever and read
> as "the fix did not ship" on a perfectly good image. And the underlying problem was
> worse: the 184 change was three identifiers, so it added **no string literal at all**
> and `strings` could not discriminate it. Fixed at source (`600bd99a8`): each emitter
> now logs its own outcome with a distinct literal, which these three previously did not
> do at all — the row was the only evidence they had run. **Verify a negative control
> against your own source, and a positive one by grepping a binary you actually built.**

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== $p"
  kubectl -n ai-persona-system exec "$p" -- sh -c "
    strings /app/agent-chassis | grep -c 'HITL citation-reject item written'       # NEW: 0 today, 1 after
    strings /app/agent-chassis | grep -c 'HITL directory-reject item written'      # NEW: 0 today, 1 after
    strings /app/agent-chassis | grep -c 'HITL directory-freshness item written'   # NEW: 0 today, 1 after
    strings /app/agent-chassis | grep -c 'refreshed the open work item'            # CONTROL: 2 on v1.0.1238
    strings /app/agent-chassis | grep -c 'FINDING NOT RECORDED'                    # CONTROL: 1 on v1.0.1238
  "
done
```

Every replica, not `logs deploy/…`. A roll is not evidence your fix shipped
(`bugs_open/153`) — the image can predate your commit, which is exactly what happened
between v1.0.1234 and v1.0.1237 here.

### 2. Then close 184 — but read this first

**Do NOT expect a live behavioural verification the way 091 got one.** 091 was
verifiable immediately because four sites were *already* dropping findings, so
re-arming its sweep was a real experiment. For 184 there is nothing to catch today:

- `stale_directory_claim` — the daily sweep checks **0** claims. Measured: 97 current
  claims, **none due until 2026-08-23**. Not broken; `loadDueDirectoryClaims` selects
  on `verified_at < now() - staleness_days`.
- `directory_citation_unverified` — weekly sweep; its row has listed the same 15
  rejects since 07-24.
- `citation_unverified` — driven by content pipelines, no schedule.

So the honest bar for 184 is **shipped + pod-grepped + unit-proven**, with the real
behavioural proof deferred to the date below. Say that in the closure rather than
implying a live run confirmed it.

### 3. 📅 2026-08-23 — the run that actually proves `stale_directory_claim`

The directory freshness batch falls due. Require:

```sql
-- the row filed 2026-07-25 must MOVE, and its count must change
SELECT updated_at, summary, jsonb_array_length(spec->'flipped') AS listed
  FROM site_work_items WHERE item_type='stale_directory_claim';
```

`updated_at` must move off 07-27 and the summary count must stop reading
"1 claim(s)". **`orchestration_states` keeps ~24h**, so check the same day or the
run's own report is gone.

## What this lane built (so you do not rebuild it)

**BATCH-005**, `docs026_concept_register/register/batch-processing.md` — the seam, its
two landmines, and the open review question. In short:

`writeWorkItem(ctx, tx, item, refreshOnConflict, logger)` refreshes an open row's
`summary`/`spec` instead of discarding the finding. `insertWorkItem` is unchanged and
delegates with `dropOnConflict`. Four call sites now opt in: `refresh_evidence_base`
(091) and the three in 184.

Three design points that are load-bearing and non-obvious:

1. **`ON CONFLICT … DO UPDATE` re-creates the bug it fixes** — it affects a row, so
   `RowsAffected()` is 1 and any "created" field starts lying. Hence a separate
   statement and a three-state `workItemWrite{Inserted, Refreshed}`.
2. **The policy is a PARAMETER, not a `workItem` field** — so a caller cannot set it
   and still call `insertWorkItem`, whose bool cannot express a refresh.
3. **The UPDATE's predicate is the concurrency guard** — terminal rows are never
   resurrected, handler-held rows (`workItemHeldStatuses`) are never mutated, and the
   unlocked gap between the two statements is lost safely in both directions.

## Landmines you will hit in this code

- **No behavioural sqlmock test in `platform/orchestration/actions` can see a change to
  `recurrenceExpected`** — the anti-churn probe discards its own error
  (`if err == nil && terminalCount > 0`), so an unexpected query is absorbed. Assert
  the built `workItem` directly. (`LANDMINES.md`, 2026-08-03.)
- **A test named for a call site must CALL it.** My first 184 test drove
  `writeWorkItem` directly, passed, read well, and was worthless. Mutate the thing the
  test is *named* for — and only trust a mutation against a **confirmed-green**
  baseline (mine was already red, which made "failed" and "did nothing" identical).
  `WRONG_CALLS.md`, 2026-08-03.
- **Widening the shared INSERT charges every caller.** Adding `parent_item_id`
  unconditionally failed 20 tests in 8 files in other lanes (sqlmock matches arg
  count). It is appended as `$17` only when set. Do not "fix" that by editing the
  twenty.
- **`workItemHeldStatuses` lives in `load_work_item_actions.go`, not beside its
  siblings in `work_items_common.go`** — only because that file had another session's
  uncommitted work when this landed. Move it when that file is clean.
- **`git mv` + a one-sided pathspec ships a COPY.** Name both paths on the commit and
  verify with `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep <n>`.

## The open question somebody should eventually rule on

`needs_human_review` is deliberately **not** in `workItemHeldStatuses`, so a refresh
can rewrite an item under a human who is reading it. Today that is safe because
`bugs_open/033` establishes the queue has no working surface — there is no reader.
**The moment a HITL surface ships, that stops being true and `needs_human_review`
should probably become a held status.** Recorded on BATCH-005; 091's guardian and
architecture seats both named it; neither blocked on it.

## Not this lane's, but found and left for someone

The `guidelines` seat on `8e7357ae` flagged a **guideline gap** rather than objecting:
the documented work-item dedup rule says "use DELETE+INSERT, not ON CONFLICT", which
the entire shared helper has contradicted for a long time. It is the rule the
`a5b70424` seat cited against `apply_gap_plan`. Worth an RFC; not taken here.

## Owed but NOT done — one objection I recorded rather than answered

`prior_art_librarian` (edit 3, medium) on `d6cda33d`: the argument that a refresh may
safely rewrite a `needs_human_review` row leans on `bugs_open/033`'s "no reader works
that queue" — **an absence claim sourced from another bug file, not from a live query in
this submission.** It is measured in this lane only as volume (368 parked, 50 raised in
7 days). What was NOT checked is whether `claimed_by` / `handled_by` show any activity
on those rows. One query, and it either confirms the downgrade or overturns the one
judgement both councils flagged:

```sql
SELECT count(*) FILTER (WHERE claimed_by IS NOT NULL) AS ever_claimed,
       count(*) FILTER (WHERE handled_by IS NOT NULL) AS ever_handled,
       max(claimed_at) AS newest_claim
  FROM site_work_items WHERE status='needs_human_review';
```
