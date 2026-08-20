# HANDOFF — 2026-08-20, fresh chat starts here: **steps 1–4 are ALL LIVE + PROVEN.** Only step 5 remains, and its precondition is measured UNMET — it is now a costed list of ~13 items, not a flip.

**Supersedes `docs/agent_docs/docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-18b_continue_here.md`**,
whose §5 checklist items 3 and 4(a)/(b) are all DONE. That file stays readable for the step-1→4
evidence trail; **do not take its step-5 plan at face value** — §5.4(c) there carries the same
correction as §4 below, and the retention argument it originally gave is unsound.

**Read in this order:** this file → NOTES `## 2026-08-20 (morning)` and `## 2026-08-19 (night)` →
`bugs_open/330` (090 CONFIRMED) → the CONTRIB at
`docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/CONTRIB_2026-08-20_from_staged_component_build_commit_sha_resolves_by_guess.md`.

## 1. What is true now (measured 2026-08-20 ~06:50Z; every figure has its query in the RUNBOOK)

- **Live build `v1.0.1317`**, both chassis pods up 2026-08-19 22:26Z, **one shared digest**
  `sha256:64783665…` (no partial roll).
- **Steps 1, 2, 3 and 4 are all live and proven.** Step 4 was verified at the **artefact**, not the
  tag: `current_page_name` and the `renderContextStepContractRenames` symbol present on **both**
  pods, present-control (`build_render_context`) ok, absent-control (`current_page_name_NOTREAL`)
  ok. The `build provenance` line had already scrolled at 8 h — expected, and why the capability
  probe is the right instrument (RUNBOOK, "Verifying a roll when the provenance line has scrolled").
- **Step 4's done-condition is MET.** pcw/`current_page` conflict rows after the roll = **0**;
  its last-ever row is `2026-08-19 22:24:39Z`, **two minutes before the roll**. Demand control: 3
  pcw runs post-roll; pre-roll rate 34 rows / 11 runs = **3.1 per run**, so ~9 were expected.
  ⚠ **Detection floor, stated so nobody over-reads it: 3 runs rules out the old every-run
  behaviour, and cannot detect a residual rarer than ~1 in 3.** No new `*.current_page` candidate
  set appeared.
- **`bugs_open/330`** — filed 08-19, **090 CONFIRMED first iteration**. Nine tools on
  webdesign.co.uk cross-linked to one unrelated tool's pages. The loop cited the layer I had only
  inferred: `strategy0Resolved` in `ExtractActionInputs` withholds from the whole-tree search only
  the fields Strategy 0 **actually resolved**, so *declared-but-empty* and *never-asked* are
  recorded identically. **Its scope-widening half is unanswered and still owed** (see §5).
- **Nothing is owed on any council review.** Prune `ae0dfb93`, tie-break `96ac93e6`, gate
  `07468ec0` (r2), step 4 `f3716ebe` (r2) — all APPROVED and all shipped.

## 2. Step 5 — what it is, and why it is not a one-commit flip

Step 5 flips `findFieldRecursive` (`platform/orchestration/datahelpers/unified_extractor.go`) from
Phase 1 ("conflicting candidates resolve to the stable shallowest winner + WARN") to Phase 2
("conflicting candidates resolve to **nothing**"). Flip sites are named in the header of
`platform/orchestration/datahelpers/unified_extractor_search_test.go`.

**The flip's own precondition, quoted from `findFieldRecursive`'s comment:** *"zero conflict WARNs
observed over the window, **or every observed field/caller pair given an explicit mapping first**."*

**It is UNMET.** 19 field/caller pairs have been logged since the instrument shipped on 2026-08-16
(that is the instrument's start date, **not** a retention floor — `agent_error_log` holds rows back
to 07-20). Steps 1–4 killed six. **Thirteen remain.**

## 3. THE STEP-5 WORK LIST (census 2026-08-20 06:55Z — re-run before trusting it)

`runs_24h` is the **demand control**: it is what separates *fixed* from *not yet provoked*. A high
count beside a quiet class means **armed, not fixed**.

| # | agent / field | rows | last seen | runs_24h | read |
|---|---|---|---|---|---|
| — | pcw / `current_page` | 869 | 08-19 22:24 | 38 | ✅ **killed by step 4** (last row pre-dates the roll) |
| — | bdl / `current_page` | 4702 | 08-19 12:14 | 227 | ✅ killed by step 3 |
| — | bdl / `work_item_id` | 2466 | 08-18 18:02 | 227 | ✅ killed by step 1 |
| — | bdl / `result` | 326 | 08-17 16:29 | 227 | ✅ killed (migration 417) |
| — | pbh / `current_page` | 46 | 08-19 11:13 | 52 | ✅ killed by 306 cand 3 |
| **1** | **bdl / `commit_sha`** | **181** | **08-20 02:45** | 227 | 🔴 **LIVE + GROWING. Do this first** — §4 |
| **2** | **tool-generator / `reason`** | 22 | **08-20 06:55** | 12 | 🔴 LIVE (minutes ago) — spec-array shape |
| **3** | **tool-generator / `related_pages`** | 20 | **08-20 06:54** | 12 | 🔴 LIVE — **this is `bugs_open/330`** |
| 4 | tool-generator / `description` | 11 | 08-18 16:33 | 12 | 🟠 armed |
| 5 | tool-generator / `function` | 11 | 08-18 16:33 | 12 | 🟠 armed |
| 6 | component-creator / `description` | 35 | 08-18 17:58 | 1 | 🟠 armed, low traffic |
| 7 | component-creator / `site_type` | 35 | 08-18 17:58 | 1 | 🟠 armed, low traffic |
| 8 | **page-rerender / `current_page`** | 78 | 08-18 13:07 | **599** | 🟠 **armed — 599 runs/24 h and quiet.** Step 4 does NOT reach it (see §4) |
| 9 | page-build-handler / `page_type` | 37 | 08-18 12:07 | 52 | 🟠 armed |
| 10 | page-build-handler / `sections` | 36 | 08-17 19:34 | 52 | 🟠 armed |
| 11 | generic / `summary` | 3 | 08-17 18:29 | 212 | 🟠 armed — resolves into the agent's OWN config |
| 12 | generic / `page_id` | 2 | 08-17 12:40 | 212 | 🟠 armed |
| 13 | rerender-pages / `reason` | 2 | 08-18 14:58 | 0 | ⚪ agent idle — unfalsifiable from this table |
| 14 | site-review-agent / `audit_source` | 2 | 08-17 12:37 | 0 | ⚪ agent idle — unfalsifiable from this table |

**Four mechanical shapes, not fourteen problems:**
- **spec-array collision** (2,3,4,5,6,7,12) — an ordinary word recurring across an array of
  unrelated objects (`specs.tools.rejected_tools[N].reason`, `specs.identity.services[N].description`).
  Biggest group; likely one fix serves most of it.
- **loop-iteration echo** (1) — one action's result under N aliases across iterations.
- **stored `content_data`** (8) — a page NAME string inside stored component data.
- **config self-reference** (11) — the search reached into `config.workflow.steps.*`.

## 4. Three things that will bite whoever builds step 5

**(a) `bdl`/`commit_sha` is a COSTED REGRESSION, and it is not ours to guess.**
`CompleteWorkItemInputSpec` (`load_work_item_actions.go:56`) declares `commit_sha` Optional and
writes `result.commit_sha` (line 937). **No live step config wires it** — so it arrives via the
whole-tree search. Inside a multi-iteration loop the per-iteration values genuinely differ (checked
at `collected_data`, not inferred from the instrument), the unsuffixed `handler_result` alias tracks
the **latest** iteration and sorts first, so it wins — and **it is probably right today by luck**
`[INFERRED]`. Flip it to refusal and `result.commit_sha` **silently stops being written**, which
`bugs_open/315`'s page-stamping depends on. A CONTRIB is filed asking the 315 lane for the path
that is correct *by their lights*. **Do not pick a path from the shape — that is the guess this
workstream exists to stop.** Wait for their answer, or take their instruction.
Its arrival is also a worked example of the landmine: it appeared 08-19 20:40Z, **three minutes
after migrations 486/487** seeded the 283 batch. It is traffic, not a regression.

**(b) Retiring the read-side tolerance: the plan's REASON was wrong, the conclusion holds.**
The old plan said the tolerance in `setRenderContextScalarsFromData` can go because the step-4 roll
will have outlived `orchestration_states`' ~24 h retention. **Do not repeat that.** Rows from
**2026-07-19** are still in the table, and the tolerance's second call site is
`mergeIntoRenderContext` — the RE-RENDER restore — where stored component `content_data` **never
expires**; **20 live `page_components` rows across 12 sites** hold `current_page` as a string today,
17 on `deployed` pages. Cite these two instead, both one query:
1. **Zero NON-TERMINAL pre-roll orchestrations** — all 2,476 are COMPLETED/CANCELLED/FAILED, so
   none can be resumed into the build-side call site.
2. **`buildRerenderBaseData` writes the NEW key fresh** from its `pageName` argument, and the
   tolerance's first branch `continue`s whenever `current_page_name` is present — so those 20
   stored rows never reach the second branch.
```sql
SELECT count(*) FROM orchestration_states WHERE created_at < '2026-08-19 22:26:25Z'
  AND status NOT IN ('COMPLETED','FAILED','CANCELLED');            -- must be 0
SELECT jsonb_typeof(content_data->'current_page'), count(*) FROM page_components
 WHERE content_data ? 'current_page' GROUP BY 1;                    -- know this number
```

**(c) The instrument UNDER-COUNTS its own class, permanently `[INFERRED, UNMEASURED]`.**
A conflict row requires the candidates to **differ** (`reflect.DeepEqual`). A tree holding ONE
match — or several that agree — substitutes silently: no WARN, no row, and the value can still be
wrong (`bugs_open/330` is the worked case). **So "zero conflict WARNs" can never establish that the
search is safe**, only that the conflicting subset is empty. Step 5's design should say this out
loud rather than inherit the precondition's "or" branch as though it were sufficient. The silent
population has **not** been measured; measuring it is a candidate task in its own right.

## 5. Also owed, smaller

- **`bugs_open/330`'s scope-widening audit.** The 090 loop explicitly declined it: *which other
  live steps declare a config input path whose value is currently supplied by the whole-tree search
  rather than by the path they named?* This sizes 330's fix candidate 2 (don't fall through for an
  explicitly-wired field), which is the only candidate that also closes the silent half in (c) —
  and it is **resolver-scope, i.e. ours**, not tool-generator's.
- **`bugs_open/307`** (outage-killed items) — converged design specced in the bug file, unbuilt.
  Not this lane's, but it is the nearest unowned coherent piece.

## 6. Traps carried forward (all still live)

- **097's `plan` is an OBJECT** (`summary`/`edits`/`grounded_in`/`risks`), not an array. A schema
  refusal is CLIENT-side — no round spent, safe to fix and re-run. A *published* run must never be
  re-triggered.
- **A mutation that breaks the BUILD proves nothing.** Mutate the body to a no-op
  (`if true { return fields }`) so the package still compiles and only the guarded behaviour changes.
- **The conflict instrument stores candidate PATHS, never VALUES.** "Are these the same page?" is
  answered only at `orchestration_states.collected_data` (RUNBOOK four-step method).
- **The demand-control join column is `owner_agent_type`** — `orchestration_states` has no
  `agent_type` column.
- **Config wiring lives at `config.<field>`, NOT `config.params.<field>`.** Probing the wrong key
  returns NULL and reads as "not configured" — that cost a false root cause on 08-19
  (`WRONG_CALLS.md`). If config has nothing, the asker is an **action input spec**, not a config key.
- **`grep -aq`'s exit 1 on no match** means an `&& echo` prints nothing and the shell reports
  `command terminated with exit code 1` — that IS the absent-control passing, not a failure.

## 7. Session-start checklist

1. `git log --oneline -10`; re-read this file from disk (it is co-edited).
2. Nothing owed on reviews (§1). Nothing owed on rolls — steps 1–4 are all live.
3. Re-run the §3 census and the demand-control join before trusting any row in it; classes move
   within hours on this tree.
4. Pick up step 5 at §3 item 1 — but **item 1 is blocked on the 315 lane's answer**, so if that has
   not landed, start with the spec-array shape (items 2–7), which is one mechanism and the largest
   group, and whose worked case (`bugs_open/330`) is already diagnosed and CONFIRMED.
5. Do **not** flip anything until every row in §3 is either killed or explicitly mapped — that is
   the precondition, and §4(c) is why "the window went quiet" will never be sufficient on its own.
