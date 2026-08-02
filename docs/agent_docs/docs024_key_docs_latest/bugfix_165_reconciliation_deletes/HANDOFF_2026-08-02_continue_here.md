# HANDOFF — 2026-08-02 evening — bugfix_165_reconciliation_deletes

**This lane is FINISHED.** `bugs_closed/165`, all four call sites guarded, live on
v1.0.1229, everything actionable from here is done. This file exists so the next
session does not re-derive any of it, and so the two things that were *raised
rather than actioned* are not lost.

Read this file, then stop unless one of the "still open" items below is your task.

---

## 1. What the lane was, in one paragraph

Four places in the platform reconcile by deleting everything they previously wrote
and re-inserting what the current run produced. Correct — until the run saw only
*part* of the corpus, at which point the delete removes the remainder and the
outcome is **absence**, indistinguishable from "there was never anything there".
`bugs_closed/135` built the shared decision rule (`prune_floor.go`, register
**CTXA-025**) and wired one call site. This lane converted the other three.

## 2. Final state — all four guarded and live

| site | table | live | refusal branch | pass branch |
|---|---|---|---|---|
| — | `code_symbols` | v1.0.1218 | induced | induced |
| A | `page_components` | v1.0.1223 | **induced in prod** 07-31 (`planned sections 35% (7 of 20)`, 7 rows byte-identical) | induced |
| B | `site_nav_items` | v1.0.1228 | **induced in prod** 08-02 (`nav items 33% (8 of 24)`, 8 rows byte-identical, 16 synthetics intact) | **a genuine production run** (`site-adoption-agent` on loancash.co.uk, both cohorts 100%) |
| C | `link_registry` | v1.0.1228 | offline mutation-proof only — **structurally un-inducible** | un-inducible |

Plus the shared **refusal-aftermath fix** (corr `22cdef56`, APPROVED, live
v1.0.1229): `Reason()` now takes a **required** caller-supplied "what happens next"
clause. It previously carried 135's aftermath to all four consumers, where it was
false for three — telling their operators to wait for a tidy-up that never comes.

## 3. Still open — and none of it is this lane's

1. **Site C's live induction is MOOT, not pending.** `link_registry` has never held
   a row; its only carrier, `multipage-website-builder`, was **retired 2026-08-02**
   (see `docs024_key_docs_latest/retired_agents/`). The floor is inert by
   construction and arms itself only if that agent is revived. If it is, run
   `RUNBOOK_reconciliation_deletes.md` § R-B2 — it transfers.
2. **`bugs_open/173`** — a refusal on one page aborts a whole multi-page build.
   Filed by the B/C lane; this lane contributed the census (four loops across four
   agents; 9 of 20 live loops set `continue_on_error`, **none** wraps a
   floor-guarded action, so nothing is being swallowed today).
3. **Two of six consumers route on error rather than failing**
   (`page-build-handler`, `tool-recreation-handler`), so a refusal there is
   recorded while the pipeline reports success. **Content is protected in both by
   construction** — the floor returns before the DELETE and the work item is
   written before the error — so this is a *visibility* question, not data loss.
   Never measured empirically; the engine rule was read instead
   (`coordinator.go:908`, `:3350-3363`, `loop_error_handler.go:71-89`).

## 4. THE BIG OPEN QUESTION, raised with the owner and not acted on

Retiring `multipage-website-builder` exposed something larger. Using `site_specs`
(**no retention job**, back to 2026-02-25, 1,874 rows, 36 sites), the only
`recommended_builder` ever recorded is **`pageflow-builder`** — 1,216 rows, 14
sites. The other five builders in the menu have **never been chosen**.

And new sites *are* being created — 25 in 30 days — while `intake-orchestrator`,
`site-classifier` and `build-briefing-agent` show no recent runs at all. So
whatever builds sites now, it is **not** the intake→classify→spawn-a-builder path
that menu belongs to.

> **CORRECTED — see §10. This section's conclusion was WRONG.** The intake path is
> not superseded; it was re-plumbed, and its research/brief stages ran today. The
> "no recent runs" behind this paragraph came from `orchestration_states`, the
> 24h-reaped table this very handoff warns about in §5. What IS superseded is
> narrower and better evidenced. Read §10 instead of this paragraph.

## 5. Landmines this lane produced (all in `LANDMINES.md`, synced to `doc_notes`)

- **`orchestration_states` keeps terminal rows ~24 HOURS**, and `min(created_at)`
  over the whole table says *20 days* because `CANCELLED`/`RUNNING`/`INITIALIZED`
  are not reaped. **This one invalidated a claim I had already published in four
  documents.** Any absence claim from that table must be bounded *per status* and,
  if durable, re-sourced from a table with no retention job.
- **`git diff | grep '^-[^-]'` cannot see a deleted markdown bullet** (`-- **x**`
  is two hyphens). Gate on `git diff --numstat`. Two sibling spellings on the same
  entry: grepping a diff for a symbol counts **context** lines; and `git log -S` is
  occurrence-COUNT based, so it misses an edit that preserves the count — use `-G`.

## 6. Reusable technique, worth more than the fix

- **A mutation suite needs at least one mutation it should SURVIVE.** Mutating the
  guarded thing proves the test *notices*; only an unrelated mutation proves it is
  *specific*. Added after a council seat objected — it is the control I had not
  thought to run.
- **Run mutations in a `git archive HEAD` tree, and compile each mutant before
  trusting its failure.** On a concurrently-edited package a mutant that breaks the
  build reads exactly like one correctly caught. (If the isolated baseline fails,
  suspect your archive: `doc_subjects_common_test.go` reads
  `docs/agent_docs/sql_for_agents`.)
- **Prefer the mechanism to the sample, and the record to the manufactured.**
  Reading the coordinator resolved three `[INFERRED]` consumers in one pass and
  surfaced a hazard no induction would have looked for; one query found a real
  production run that made an expensive synthetic pass-branch unnecessary.
- **Measure the cohorts before choosing them.** The bug file's own suggested
  partition was refuted by data **twice** — for A because the cohorts were too
  SMALL (998 of 1,009 `(page_id, slot_name)` groups hold one row), for B because
  the key is not STABLE (the classifier re-homes pages between groups). **The
  general test: an AUTHORED table ratchets, a DERIVED table self-heals.**

## 7. Cold-start pointers

- Case: `bugs_closed/165`, `bugs_closed/135`, `bugs_closed/092`
- Register: **CTXA-025** in `docs026_concept_register/register/context-assembly.md`
- Lane docs: this directory — `PLAN`, `RUNBOOK` (§ R-B2 is the live-induction
  recipe), `NOTES` §1–13, `README_where_we_are`, `SUMMARY_2026-08-01`
- Retirement: `docs024_key_docs_latest/retired_agents/`
- Councils: `a54172b6` (A), `c69e935a` (B+C), `22cdef56` (aftermath) — all APPROVED

## 8. Two process errors worth inheriting

1. **I cited a bug number without resolving it.** "Blocked on `bugs_open/092`" went
   into three durable documents; 092 had closed the same day the pointer was
   written, and closing it did not unblock anything. `who-owns.py <n>` costs 0.3s.
   **Citing a bug number is acting on it.**
2. **A deferral names a destination and nobody checks the destination accepted
   it.** 092 said the link-registry question was 165's; 165 said it was 092's.
   Neither owned it and both closed. **Write the item into the OTHER case's file,
   not only your own.**

---

## 9. Addendum from the sites B+C lane — post-`v1.0.1229` (appended, nothing above edited)

**Everything this lane owed is live and verified.** `v1.0.1229`, both replicas,
positive + pipeline + negative controls each time. §14 and §15 of
`NOTES_reconciliation_deletes.md` carry the tables and the exact grep strings.

Three commits after the roll, all comment/string only, all builds+tests green:

| commit | what |
|---|---|
| `cdbe27325` | 9 stale `bugs_open/` pointers in the B+C sources; the `092` ones carried a wrong *answer*, replaced with the measured one |
| `e1300a81c` | the third sibling, `save_sections_prune_floor.go:280` — site A's file, found by a **failing** negative control |
| `f6913a1fa` | this lane's own header posed a decided question as open — `multipage-website-builder` was retired 7h before the comment was written |

**Nothing is owed. Two things are left for others, neither this lane's:**

1. **`bugs_open/173`** — per-substep error routing. Still OPEN and **UNOWNED**. It
   is latent (nothing fails because of it today) but it is what pushed site C into a
   workaround that four council seats rejected, and §CONTRIBUTION shows it would buy
   three of site A's loops as well as C's.
2. **The stale-citation class, measured and deliberately not swept.** 107 of 139 bug
   numbers cited as `bugs_open/NNN` in Go source now live in `bugs_closed/`; ~100 of
   those sit in *string literals* that reach a human (log lines, work-item remedy
   text). Reproduction and the reason a mechanical rewrite is **unsafe** — numbers
   naming two unrelated cases can legitimately exist in both directories — are in
   NOTES §15. The durable fix is a `pattern-check.py` rule firing at edit time, not
   a sweep.

**The one thing I would tell the next session.** This lane's characteristic failure
was not a wrong measurement — every number checked out. It was **citations asserting
a status that had already changed**: `092` closed before I pointed at it, `165`
closed while I was writing about it, the retirement decided before I called it open.
Three instances in one session, the last two *after* writing the WRONG_CALLS entry
about the first. Knowing the class did not help. The only thing that worked was
querying the live row — `git ls-tree` for a bug path, `agent_definitions` for an
agent — in the same breath as writing the sentence.


---

## 10. §4 ANSWERED, and §4 was WRONG — the build path, measured 2026-08-02 evening

Asked to check whether the intake path is superseded, before retiring the other
builders. **It is not superseded. It was re-plumbed**, from an
orchestrator-spawns-a-builder model to a work-item model.

### What is ALIVE (all of it ran on 2026-08-02)

Evidence is `site_work_items.handler_agent` and `site_specs.created_by` — both
**durable, no reaper**, unlike `orchestration_states`.

```
domain-submitter            (entry point: ensure_site_record -> persist mission/roadmap/contact
                             -> create_research_item -> complete.  It spawns NOTHING; it files a WORK ITEM)
   -> domain-research-classifier   items 2, specs 58/15 sites, last 2026-08-02
   -> domain-strategist            items 1,           last 2026-08-02
   -> vertical-exemplar-researcher items 1,           last 2026-08-02
   -> site-design-planner          items 1, specs 12, last 2026-08-02
   -> build-briefing-agent         items 1, specs 15, last 2026-08-02
   -> build-site-planner           items 1,           last 2026-08-02
        steps: read_specs -> plan_site -> validate_plan -> reconcile_site_plan
               -> write_site_plan -> sync_pages -> populate_nav -> emit_design -> emit_imagery
```

**`build-site-planner` is the builder now.** It does the work the `*-builder`
agents were shaped for, and it is reached by a work item rather than by a spawn.

### What IS superseded — narrower, and now properly evidenced

| thing | evidence |
|---|---|
| `intake-orchestrator` | **0** work items, **0** specs, no `scheduled_tasks` row, no agent config spawns it |
| `site-classifier` | **0** work items, **0** specs |
| **the entire `%-builder` menu** | **`SELECT ... FROM site_work_items WHERE handler_agent LIKE '%-builder'` returns ZERO rows, all history.** No work item has *ever* named a builder |

So `multipage-website-builder`'s retirement was right, and the remaining five are in
the *same* position — but on better evidence than was used for it: nothing routes
to any of them, ever.

### The finding worth keeping: a vestigial field, faithfully maintained

`domain-research-classifier`'s prompt **mandates** `recommended_builder` and pins it
(*"should always be `pageflow-builder` for now"*). It has written that value into
**1,216 spec rows across 14 sites**, and **nothing consumes it** — no work item is
ever routed to a builder. An LLM is being instructed, on every classification, to
fill in a field for a consumer that no longer exists.

**Before retiring `pageflow-builder`, fix that prompt** — otherwise the classifier
keeps emitting a name that resolves to nothing. The other four
(`content-site-builder`, `landing-page-builder`, `report-builder`,
`website-builder`) are not named anywhere and can go on the same evidence as
`multipage-website-builder`, using `retired_agents/` as the pattern.

### How §4 got it wrong, which is the transferable part

§4 said "intake-orchestrator, site-classifier and build-briefing-agent show no
recent runs". Two of those three were right; **`build-briefing-agent` ran today**
and so did `domain-research-classifier`. The claim came from
`orchestration_states` — the table §5 of this very file warns keeps terminal rows
~24 hours. **I filed that landmine and then reasoned from the same table hours
later, in the same session.** Knowing the class does not protect you; changing the
source does. The durable answer took two queries against tables with no reaper.

### 9a. Docs added after this handoff was written (2026-08-02, evening)

- **`SUMMARY_2026-08-02_all_four_guarded_and_the_lane_closes.md`** — the milestone
  read-out, and the **last** in this lane's series. A NEW file, not an edit of
  `SUMMARY_2026-08-01_site_a_live_and_proven.md`; the series is the record. Start
  here if you want the whole story in plain prose.
- **`RUNBOOK` R-V1** — prove a fix reached the running binary, plus the three check
  spellings that lied to us (an `-E` pattern one character short reading as "not
  shipped"; a negative control matching 20 unrelated queries; `IS NULL` on a
  `NOT NULL DEFAULT ''` column). All three fail toward *"your change is not
  there"*, which is why they survive a careful reader.
- **`RUNBOOK` R-V2** — check a bug citation *before* writing it, and why the
  fleet-wide instance must not be swept by script.
- `NOTES` §14 and §15 — the post-roll verification tables and the measurement of
  the stale-citation class.

Nothing above section 9 has been edited at any point.
