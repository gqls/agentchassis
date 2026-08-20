# HANDOFF — 2026-08-20 (rev. ~17:15Z), fresh chat starts here: **steps 1–4 LIVE + PROVEN. Step 5 is the only work left, and it is now TIERED: 2 hard blockers, 1 silent-loss blocker, ~10 record-a-decision items.** The lane cannot close yet.

> **⚠ THIS FILE NOW CONSOLIDATES TWO.** `HANDOFF_2026-08-18b_continue_here.md` was still being
> updated by a parallel session of this lane until ~10:17Z today (its audit results are folded in
> below and credited). **That file is now bannered SUPERSEDED — read only this one.** Two files
> both saying "fresh chat starts here" existed for ~9 h; that was my doing, and §7 records it.

**Read in this order:** this file → NOTES `## 2026-08-20 (evening)`, `(morning)`, `## 2026-08-19 (night)`
→ `bugs_open/330` (090 CONFIRMED; **§9 carries the sizing audit + its two dated corrections**) →
`bugs_open/334` (bdl/`commit_sha`) → the CONTRIB at
`docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/CONTRIB_2026-08-20_from_staged_component_build_commit_sha_resolves_by_guess.md`.

## 1. What is true now (measured 2026-08-20 ~17:00Z)

- **Live build `v1.0.1320`**, both pods up 16:09Z. **Nothing of this lane is in it** — steps 1–4
  all shipped in `v1.0.1317` and step 5 is unbuilt. The 1320 roll is a **non-event here**; do not
  manufacture a verification task for it.
- **Steps 1, 2, 3, 4 are all LIVE and PROVEN.** Step 4 verified twice independently: at the
  artefact on both pods (capability probe `current_page_name` + present- and absent-controls) and
  by the parallel session (stamp `2d13d530d`, 3/3 trees filing `current_page_name`). pcw/
  `current_page`'s last-ever conflict row is **2026-08-19 22:24:39Z — two minutes before the roll.**
- **Step 5 is NOT built.** `findFieldRecursive` still carries only the Phase-1 warn
  (`grep -c "conflicting = true"` → 1; no Phase-2 arm). **`bugs_open/334`'s candidate 1 is NOT
  built either** — no live step config wires `commit_sha` (only `code-indexer/index_symbols` and
  `site-work-orchestrator/build_items_loop` mention it, neither being the caller).
- **Nothing owed on any council review.** Prune `ae0dfb93`, tie-break `96ac93e6`, gate `07468ec0`,
  step 4 `f3716ebe` — all APPROVED, all shipped.
- **`bugs_open/330`** — 090 CONFIRMED. **Its owed sizing audit is DONE** (parallel session, 10:12–10:17Z):
  451 plain Strategy-0 wires fleet-wide / 309 pairs / 83 agents; stripped-probe runtime sample of
  the 8 high-demand agents → **4 genuinely rescue-prone wires** (the naive `LIKE` probe said 10 and
  over-counted by matching `agent_config`/`__raw_message__`/`retry_payload`, which the search skips
  via `isInfrastructureKey`). Of the 4, **pbh `page_id`+`page_name` produce AGREEING candidates**
  (0 genuine disagreements / 40 runs) so they are **invisible to step 5's conflict flip** and matter
  only to 330 candidate 2. Wrong-value population on the sampled slice = **exactly 330's own wire**.
  Unsampled remainder: 269 pairs / 75 agents; the corrected probe is reusable (RUNBOOK).
- **`bugs_open/334`** — filed for bdl/`commit_sha`. Its 090 came back **UNVERIFIABLE at the
  iteration cap** = no information (016b §9's standing reading), so the mechanism rests on two
  sessions' first-hand verification. Two corrections landed there: the onset is the **486/487
  traffic batch**, not an adapter roll; and `Deprecated: commit_sha_field` is a **LIVE Strategy-3
  bridge** which **cannot** stop the conflict (it runs *after* the search and is gated on a missing
  value) — so candidate 1 must be the Strategy-0 `commit_sha?` form.

## 2. THE KEY ADVANCE (2026-08-20 evening): step 5's list is TIERED, not flat

Every conflicting field was checked against the `ActionInputSpec` that declares it. **All but two
are `Optional`.** That splits the ~13 pairs into three tiers with completely different costs, and
it is the design step 5 needed:

### Tier A — HARD BLOCKERS: the field is `Required`, so the flip turns a guess into a FAILURE (2)

| pair | action | spec | note |
|---|---|---|---|
| **tool-generator / `function`** | `create_tool_component` | **Required** (`create_tool_component_action.go:44`) | Confirmed by elimination — `create_tool_component` is the only action in tool-generator's 10 steps that takes `function`. Wired by migration 211 (`config.function = input_data.spec.function`); the conflict fires only when that path is empty, and then the search rescues it. **Post-flip the action fails hard.** Arguably correct (better than building a tool under another tool's function name) but it WILL surface as failures — decide deliberately, do not discover it |
| **site-review-agent / `audit_source`** | `write_audit_findings` | **Required** (`write_audit_findings_action.go:43`) | Agent has **0 runs in 24 h** and no trace in `orchestration_states` at all, so low risk today — but see the §6 caveat: that is not the same as retired |

### Tier B — SILENT-LOSS BLOCKER: `Optional`, but something downstream depends on the value (1)

**bdl / `commit_sha`** — `Optional` in `CompleteWorkItemInputSpec`, written to `result.commit_sha`
(`load_work_item_actions.go:937`). Absence is "handled" in the sense that Optional means no error —
**and that is exactly the danger**: the field silently stops being recorded, and `bugs_open/315`'s
page stamping depends on it. **This is the one pair that genuinely needs a config wire before the
flip.** 442 conflict rows and growing (last 15:23Z). `bugs_open/334` candidate 1. A CONTRIB asks
the 315 lane for the path that is correct *by their lights* — **do not pick one from the shape.**

### Tier C — RECORD A DECISION, no migration needed (~10)

`tool-generator`/`reason`,`related_pages`,`description` · `component-creator`/`description`,`site_type` ·
`page-build-handler`/`page_type`,`sections` · `generic`/`summary`,`page_id` · `page-rerender`/`current_page` ·
`rerender-pages`/`reason`

All `Optional`, and for the spec-array shape **absence IS the correct answer** — the audit's own
words for 330: *"absence is the fix"*. So the precondition's "give every pair an explicit mapping"
is satisfied here by a **recorded judgement that nothing is the right value**, not by config. That
is a paragraph per pair, not a migration per pair. **This is why the remaining work is much smaller
than the flat count of 13 suggests.**

⚠ The one in Tier C to actually look at is **`page-rerender`/`current_page`** (78 rows, **627
runs/24 h** and quiet): step 4 deliberately did not rename the stored template key, so this route
survives by design. Confirm absence is right there rather than assuming it.

## 3. Recommended order of work

1. **Tier C first** — it is ~10 paragraphs of recorded judgement and it shrinks the list fastest.
   Write them into `bugs_open/330` §9 or a new step-5 design doc, one per pair, each naming the
   action and why absence is correct.
2. **Tier A** — two decisions. For `tool-generator/function`, measure how often
   `input_data.spec.function` is actually empty (if never, the flip is free; if sometimes, it is a
   deliberate new failure mode and needs the owner or the council). For `audit_source`, the agent
   is dormant — a safe-by-inspection note is probably enough, with §6's caveat stated.
3. **Tier B last, and it is the true gate** — blocked on the 315 lane's answer. Chase it, or take
   their instruction. This is the only item that can silently lose data.
4. **Then** flip, council-gated, and retire the read-side tolerance in the same commit — using the
   §4 reasons below, **not** the retention argument.

## 4. Retiring the read-side tolerance: the plan's REASON was wrong, the conclusion holds

Do **not** repeat "the step-4 roll has outlived `orchestration_states`' ~24 h retention". Rows from
**2026-07-19** are still in the table, and the tolerance's second call site is
`mergeIntoRenderContext` — the RE-RENDER restore — where stored component `content_data` **never
expires** (20 live `page_components` rows across 12 sites hold `current_page` as a string; 17 on
`deployed` pages). Cite these instead, both one query:
1. **Zero NON-TERMINAL pre-roll orchestrations** — all 2,476 pre-roll rows are
   COMPLETED/CANCELLED/FAILED, so none can be resumed into the build-side call site.
2. **`buildRerenderBaseData` writes the NEW key fresh** from its `pageName` argument, and the
   tolerance's first branch `continue`s whenever `current_page_name` is present — so those 20 stored
   rows never reach the second branch.
```sql
SELECT count(*) FROM orchestration_states WHERE created_at < '2026-08-19 22:26:25Z'
  AND status NOT IN ('COMPLETED','FAILED','CANCELLED');            -- must be 0
SELECT jsonb_typeof(content_data->'current_page'), count(*) FROM page_components
 WHERE content_data ? 'current_page' GROUP BY 1;                    -- know this number
```

## 5. The instrument's permanent blind spot (unchanged, and it bounds what the flip can promise)

A conflict row requires the candidates to **differ** (`reflect.DeepEqual`). A tree with ONE match —
or several that agree — substitutes silently: no WARN, no row, and the value can still be wrong.
**So "zero conflict WARNs" can never establish the search is safe**, only that the conflicting
subset is empty. The audit above put a *measured floor* under this (4 rescue-prone wires on the
sampled slice, 1 carrying a wrong value) but 269 pairs / 75 agents are unsampled. Step 5's design
must state this rather than inherit the precondition's "or" branch as if it were sufficient.

## 6. ⚠ Correction to this lane's own demand control (new, 2026-08-20 evening)

The demand-control join I introduced (`orchestration_states.owner_agent_type`, RUNBOOK) is sound
for **"is this agent running NOW?"** and nothing more. It cannot size a class historically:
`site-review-agent` has **58 rows in `agent_error_log`** and **zero trace in `orchestration_states`
by any column** (`owner_agent_type`, `workflow_plan`, `execution_metadata`) — its runs aged out.
`agent_error_log` retains from 07-20; `orchestration_states` keeps ~24 h of COMPLETED rows. So:
- `runs_24h = 0` **does** license "cannot fire right now".
- `runs_24h = 0` does **NOT** license "retired", "never ran", or any statement about the period when
  the class was firing. For that, use the error log's own history.
My earlier table marked two rows "agent idle" — the conclusion held, the stated reason was too
strong. The LANDMINES entry has been amended.

## 7. Traps carried forward

- **Two "continue here" handoffs coexisted for ~9 h** because I wrote a new dated file at 07:59
  while a parallel session kept updating the old one. **Before creating a dated successor, grep the
  lane for other `*_continue_here.md` and check `git log` on them for edits newer than your read.**
  Consolidated here; 08-18b bannered.
- **097's `plan` is an OBJECT** (`summary`/`edits`/`grounded_in`/`risks`), not an array. A schema
  refusal is CLIENT-side — no round spent. A *published* run must never be re-triggered.
- **Council scope widened 2026-08-19** to appliable migrations under `docs/agent_docs/sql_for_agents/`
  (`bugs_open/314`), `_HOLD.sql` included — so a Tier A/B migration is now council-gateable.
  Scope is single-sourced in `scripts/council-scope.sh`; `DRY_RUN=1` tests admission free.
- **A mutation that breaks the BUILD proves nothing.** Mutate to a no-op (`if true { return fields }`).
- **The instrument stores candidate PATHS, never VALUES.** Judge a class at
  `orchestration_states.collected_data` (RUNBOOK four-step method).
- **Config wiring is at `config.<field>`, NOT `config.params.<field>`** — probing the wrong key
  returns NULL and reads as "not configured" (cost a false root cause, `WRONG_CALLS.md` 08-19).
  If config has nothing, the asker is an **action input spec**.
- **A top-level `jsonb_each(steps)` census misses sub-workflow/loop substeps** — bit this lane
  three times now, incl. 334's `mark_complete`. Use the whole-config text search as the ceiling.
- **`grep -aq` exits 1 on no match**, so `&& echo` prints nothing and the shell says
  `command terminated with exit code 1` — that IS the absent-control passing.

## 8. Session-start checklist

1. `git log --oneline -10`; re-read this file from disk. **~270 commits landed here in one day** —
   assume anything you remember is stale.
2. Nothing owed on reviews; nothing owed on rolls.
3. Re-run the census (RUNBOOK) **with the demand control read per §6's limits**, before trusting
   any row of §2.
4. Start at §3 step 1 (Tier C). Tier B is the real gate and is blocked externally.
5. **Do not flip anything** until every §2 row is killed, wired, or has a recorded decision.
