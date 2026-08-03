# 184 — three more HITL detectors key their work item per SITE while finding per ITEM, so every finding after the first is dropped

**Filed:** 2026-08-03 by the bugs-sweep lane, while closing `bugs_closed/091`.
> **⚠ `184` IS AN AMBIGUOUS NUMBER — two unrelated cases.** This one (per-site key over a
> per-item finding), and `184_HANDOFF_2026-08-03_llm_markdown_reaches_the_page_as_literal_asterisks.md`
> (the mortgagecalculator lane, filed 27 minutes later). Both stay: numbers are never
> reassigned, and the repo already carries 016/017/083/112/131/146 in the same state.
> **Resolve by SLUG, and `git log` the FILE PATH, not the number** — a commit message
> saying "184" is not evidence about which.
**Severity:** Medium, and **unmeasured on these three** — that is the first job below.
091's instance was Medium on paper and turned out to have **4 of 5 live records naming
the wrong facts**, so do not assume these are quieter without looking.
**Class:** the `091` class exactly — key coarser than finding, dedup drops the finding,
the durable record goes on describing the first thing it ever saw.
**Status:** **FIXED IN HEAD, NOT LIVE** — all three switched (`4c3a968cc`),
`Council-Submitted: d6cda33d-1e5a-4ea3-8ddc-98f0a6848e35`. Go code, so it is inert
until the next chassis image. **Do not close on the commit.**

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

---

## 2026-08-03 — FIXED IN HEAD, and the measurement corrected this file's own framing

All three now call `writeWorkItem(..., refreshOnConflict, ...)` (`4c3a968cc`). No new
mechanism — three identifiers plus a test per call site. Because all three already
discarded `insertWorkItem`'s boolean, they also inherit for free the not-recorded
warning that `091`'s council asked to be moved INTO the shared writer.

> ### ⚠ CORRECTION to this file, from measuring what it told the next reader to measure
>
> The header said "**unmeasured on these three** — that is the first job", and implied,
> by placing them in 091's class without qualification, that findings were being dropped
> now. **They are not, and the difference changes how this should be read.**
>
> **`stale_directory_claim` — the exposure is DATED, not continuous.**
> `model-directory-freshness` is enabled and runs daily, and it checked **zero** claims
> on 2026-08-03 09:32Z. Cause read from `loadDueDirectoryClaims`, not inferred: it
> selects on `verified_at < now() - staleness_days`, and of **97** current claims **none
> is due** — `min(verified_at + staleness)` = **2026-08-23**. All 134 claim rows are
> status `found`. So nothing is being dropped this week.
>
> **That is an argument for fixing it now, not for deferring it.** A batch falls due in
> three weeks; the row filed 2026-07-25 will still be holding the key when it does,
> because nothing works the `needs_human_review` queue (`bugs_open/033` — 368 parked, 50
> raised in 7 days). Its summary will still read "1 claim(s) changed verification
> status". **Waiting for the symptom here means waiting for the one run in three weeks
> that matters, and then losing it.**
>
> **`directory_citation_unverified` — [UNMEASURED, AND UNRECOVERABLE].** The live row has
> listed the same **15** rejected candidates since 2026-07-24, across every weekly
> `model-directory-discovery` sweep since (last 07-31). Whether that sweep found a
> different set **cannot now be established**: `orchestration_states` retains ~24h, and a
> rejected *candidate* never reaches `directory_claims`, so a dropped finding here leaves
> no trace in any table. Marked rather than guessed. It is the argument for the refresh —
> it is what makes the next one observable.
>
> **`citation_unverified`** — one row on oufe.com since 2026-07-26, one reject listed.
> Driven by content pipelines, so cadence is caller-dependent and was not measured.

### The judgement, made explicitly because this file demanded it be

Question 1 of "what each one needs decided" was whether a silently-CHANGING record beats
a silently-WRONG one here. **Yes, and the case is stronger than 091's, not weaker.** 091's
least-certain call was that a refresh can rewrite a row under a human mid-read.
`bugs_open/033` establishes that queue **has no working surface at all** — so there is no
reader to disturb, and what the current behaviour protects is a description that is
already false. Question 3 (a per-item key instead) stays refused for 091's reason: 033's
queue must not fill, and a refresh keeps the count at exactly one row.

### Tests — and the draft that was worthless

One test per call site, each **driving the real emitter**, each mutation-proven: revert
`evidence_citations` and only its test fails; revert `stale_directory_claim` and only its
test fails. Plus a counterfactual asserting the DEFAULT policy still drops, so the three
are known to be measuring something.

**The first draft called `writeWorkItem` directly and asserted the outcome.** It passed,
it read convincingly, and it was worthless — reverting a call site did not fail it,
because it never touched the call sites. Logged in `WRONG_CALLS.md`; the check that
removes the class is *mutate the thing the test is NAMED for*, against a
confirmed-green baseline.

### What is owed before this closes

1. **The roll**, then pod-grep. Discriminating marker for this change (0 today):
   ```bash
   strings /app/agent-chassis | grep -c 'the bugs_closed/091 class'
   ```
   Run it on **every** replica, with the 091 markers as positive controls
   (`refreshed the open work item` → 2, `FINDING NOT RECORDED` → 1 on v1.0.1238).
2. **Watch 2026-08-23**, when the directory freshness batch falls due — that is the first
   run that can exercise `stale_directory_claim` for real. Require the July row's
   `updated_at` to move and its summary count to change.
3. **The council verdict** on `d6cda33d`, and act on it.
