# 184 — three more HITL detectors key their work item per SITE while finding per ITEM, so every finding after the first is dropped

**Filed:** 2026-08-03 by the bugs-sweep lane, while closing `bugs_closed/091`.
**Severity:** Medium, and **unmeasured on these three** — that is the first job below.
091's instance was Medium on paper and turned out to have **4 of 5 live records naming
the wrong facts**, so do not assume these are quieter without looking.
**Class:** the `091` class exactly — key coarser than finding, dedup drops the finding,
the durable record goes on describing the first thing it ever saw.
**Status:** OPEN, not started. The remedy already exists and is live; the work is
one judgement per call site, not new machinery.

Grepped `bugs_open/` and `bugs_closed/` for the mechanism before filing: the only prior
member is `091` (closed 2026-08-03). No existing bug covers these three.

---

## Why this is a separate file rather than 091 staying open

`091` is fixed, live and verified on its own instance (`stale_evidence`), and its
`bugs_closed/` entry has to mean "this is not biting prod". These three are **different
instances of the same class**, found by enumerating call sites while fixing it. Leaving
091 open for them would have hidden a closed defect behind an unstarted one.

## The three

All HITL-terminal (`handler_agent='human-review'`, so the only consumer is a person),
all carrying a LIST in `spec` that will differ next run, all keyed so the list cannot
grow:

| file:line | item_key | what the spec holds | key granularity |
|---|---|---|---|
| `platform/orchestration/actions/evidence_citations.go:426` | `citation_unverified:<siteID>` | `rejected[]` — candidate claims that failed live verification | per SITE |
| `platform/orchestration/actions/directory_claims.go:333` | `directory_citation_unverified` | failures from the model-directory sweep | **constant** (one row for the whole directory site) |
| `platform/orchestration/actions/directory_claims.go:575` | `stale_directory_claim` | `flipped[]` — claims whose verification status changed | **constant** |

The third is the sharpest: its summary is
`fmt.Sprintf("Model directory freshness: %d claim(s) changed verification status", len(flipped))`
— **it states a count of the findings it is about to drop.** A human reading a stale row
sees a number that was true once.

## The remedy exists, is live, and is one line per site

`writeWorkItem(ctx, tx, item, refreshOnConflict, logger)` — **BATCH-005**, live on
`v1.0.1237`, verified by induced-from-real-data run 2026-08-03 08:56Z
(`bugs_closed/091`). It refreshes the open row's `summary`/`spec` instead of discarding
the finding, keeps the row count at exactly one, skips terminal and handler-held rows,
and reports `Refreshed` distinctly from `Inserted` so no caller can report a refresh as
a creation.

**So this is NOT a build task. It is three judgements**, and that is deliberately not
mechanical:

## What each one needs decided before switching, and by whom

1. **Is a silently-CHANGING record better than a silently-WRONG one, here?**
   091 answered yes for `stale_evidence` on measured evidence: its records were wrong on
   4 of 5 sites and the reader was being sent to the wrong fact. `needs_human_review` is
   deliberately NOT in `workItemHeldStatuses` (it is a queue, not a claim — nothing can
   hold one while `bugs_open/033` has no working surface), so a refresh **can** rewrite
   an item under a human mid-read. That was the judgement 091's council was asked to
   challenge and did not block on; it is **not** automatically transferable, and the
   answer may differ for the model directory, which has a different reader.
2. **Measure first — the same query 091 used.** Compare what each open item SAYS against
   what the last run FOUND. For the two directory items the site is
   `directorySystemSiteID`, so the whole directory shares ONE row:
   ```sql
   SELECT swi.item_type, swi.updated_at, swi.summary, swi.spec
     FROM site_work_items swi
    WHERE swi.item_type IN ('citation_unverified','directory_citation_unverified','stale_directory_claim')
      AND swi.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
   ```
   Then run the emitting sweep and diff. **`orchestration_states` keeps ~24h**, so the
   run and the comparison have to happen the same day (091's RUNBOOK, gotcha 1).
3. **Would a per-item key be better than a refresh, for these?** 091 refused that shape
   (candidate 3) because one row per finding fills `needs_human_review`, which
   `bugs_open/033`'s owner ruling forbids — measured 2026-08-03: **368 open, 50 raised in
   the last 7 days**. The two DIRECTORY items may be different: they are keyed
   *constant*, i.e. one row for the entire directory rather than one per site, which is
   coarser than anything 091 dealt with. A per-site key there might be right where a
   per-finding key is not.

## LANDMINES for whoever takes this

- **The happy path is identical under both policies.** An insert that succeeds behaves
  the same whether or not the key is granular enough, so this defect is invisible in
  every test, every green build, and every log line except one. Do not conclude "fine"
  from a clean run — compare the RECORD against the FINDING.
- **`refreshOnConflict` is reachable only through `writeWorkItem`.** It is a parameter,
  not a `workItem` field, precisely so a caller cannot set it and still call
  `insertWorkItem`, whose single bool cannot express a refresh. If you find yourself
  wanting a field, read `conflictPolicy`'s comment first.
- **No behavioural sqlmock test in `platform/orchestration/actions` can see a change to
  `recurrenceExpected`** — the anti-churn probe swallows its own error, so an unexpected
  query is absorbed. Assert the built `workItem` directly. (`LANDMINES.md`, 2026-08-03.)
- **`directoryS­ystemSiteID` means the two directory rows are not per-site at all.**
  Anything you reason about "one row per site" is wrong for those two; check what the
  site column actually holds before designing a key.

## Related

- `bugs_closed/091` — the parent case, the mechanism, the live verification, and the
  three deviations from its own filed candidate that are worth reading before reusing it.
- **BATCH-005** in `docs026_concept_register/register/batch-processing.md` — the seam,
  its two landmines, and the open review question about `needs_human_review`.
- `bugs_open/033` — why one row per finding is not available as a fix here.
- `docs024_key_docs_latest/bugfix_091_workitem_conflict_refresh/` — PLAN, RUNBOOK
  (the measurement queries, with their three gotchas), NOTES, README.
