# PLAN 2026-08-19 — bugs_open/313 + 298: the internal linker has never made a link

**Lane:** `bugfix_313_internal_linker` · **Bugs:** `bugs_open/313` (the dead branch),
`bugs_open/298` (the LIMIT 15 cap, moot until 313 is fixed) · **Session:** named `bugs_open/313`,
owner-directed 2026-08-19 ("prepare a plan … preferring a robust solution … applicable to the
framework as a whole in preference to the individual case").

## What we are fixing

`internal-linker`'s `load_candidate_pages` declares `output_format: "array"` (bare slice, no
`count` key) while `check_candidates` tests `candidate_pages.count > 0` — a path that cannot
resolve, so the numeric arm of `evaluateSingleComparison` returns `false, nil` and every run since
the agent's creation (2026-04-12) has exited at `complete_no_candidates`. `plan_links` — the only
LLM step — has therefore **never run**: zero `llm_call_log` rows in all history, re-verified
2026-08-19. 298's alphabetical `LIMIT 15` on the same query has consequently never shaped a link,
and goes live the moment 313 is fixed (8 of 26 sites exceed it today; worst 69).

**Diagnosis basis (owner ruling 2026-07-31):** 313 already carries a CONFIRMED diagnosis-loop
verdict (`RUN_CORRELATION_ID=c4aa3559-86b1-4356-a28b-c71dfa661465`, first iteration, 2026-08-18).
A second 090 run on the same mechanism would be a duplicate round; substituted instead:
**first-hand re-verification today** of (a) the live config (still `array` / `.count > 0`, row
`93cffe67`, queried 2026-08-19), (b) the cited code paths (`database_actions.go:129-145`,
`conditional_branch_action.go:275-284, 397-412` — unchanged since the verdict), and (c) the
counters (llm_call_log still 0; 20 open work items; diagnosis queue empty — no other thread on it).
`who-owns.py`: both bugs were filed by the `bugfix_275_silent_row_caps` lane, which **closed
2026-08-19** with these tickets explicitly handed off as unowned.

## The change set

### 1. Migration `490_internal_linker_candidates_object_uncapped_fail_loud.sql` (+ ROLLBACK sidecar)

Config-only, **live on apply**, id-scoped to `93cffe67-baf4-4fb1-bec9-ba546fb24a54`, in the
snapshot → DO/RAISE pre-gate → jsonb_set → DO/RAISE verify shape migration 484 used. Four edits,
one row:

1. `load_candidate_pages.config.output_format`: `"array"` → `"object"` — the producer then emits
   `rows`/`count`/`columns`, making `candidate_pages.count > 0` resolvable, matching the working
   `load_target_page`/`check_target_found` pair in the same workflow (313 fix candidate 1).
2. `load_candidate_pages.config.query`: drop `LIMIT 15` (298 fix candidate 2 — the 275-approved
   shape: the dominant column `content_sample` is already bounded at `LEFT(…, 800)`, worst site 69
   rows ≈ 60 kB, well within the model's context) **and mark the truncation** the LEFT() performs
   (`[…truncated]`, migration 446's worked remedy). `ORDER BY p.name` stays — with no cap it is
   deterministic presentation, no longer a cut.
3. `plan_links.config.prompt_template`: `{{range .candidate_pages}}` → `{{range
   .candidate_pages.rows}}` — **the mandatory second half**: ranging the new map would yield
   `columns`/`count`/`rows` in key order, a broken prompt (313's stated trap).
4. `check_candidates.config.fail_on_non_numeric`: `true` — opts this routing condition into the
   new loud-failure guard (§3). Inert until the chassis rolls (unknown config keys are silently
   ignored at execution — bugs_open/101), and inert afterwards unless the path stops resolving
   again, i.e. purely a tripwire against re-drift.

The seed `101_internal_linker.sql` is deliberately **not** edited: it is applied history and the
drift guard checks file content; the migration supersedes it and says so in its header.

### 2. Framework half A — offline audit `config-key-audit --array-producer-conditions` (closes the class at config time)

313 fix candidate 2, the half that stops the next one. New mode on the existing audit binary
(WFA-004/006/007/013 precedent), `cmd/config-key-audit/conditionalshape.go` + test + wrapper
`scripts/audit-array-producer-conditions.sh` (same live-export query as the sibling scripts):
for every `conditional` step — walked with `validation.WalkSteps`, so nested `sub_workflow`/
`substeps` steps are covered (the 32%-undercount landmine) — take each dotted left-side path
`F.k…` in its condition; if the step producing `output_field: F` in the same workflow is a
`query_database` whose **effective** `output_format` is `array` (explicit *or defaulted* —
`database_actions.go:25` defaults to array) and `k` is not a numeric index, that is a finding:
the path can never resolve. Exit 1 on findings. Census 2026-08-18 says the live fleet has exactly
one — `internal-linker` — so the audit lands with a clean fleet the moment 490 applies, and the
motivating case is pinned as a fixture (pre-fix shape must flag; post-fix must not).

### 3. Framework half B — opt-in `fail_on_non_numeric` on `conditional` (closes the class at run time)

313 fix candidate 3, shipped in the estate-sanctioned shape rather than as a blanket behaviour
change: a bool step-config key, **default OFF, absent key preserves behaviour byte-for-byte**.
When set, a numeric comparison (`>`, `>=`, `<`, `<=`) whose left or right side fails `ToFloat64`
**fails the step** with an error naming the field, instead of silently returning false and taking
`else_step`. Deliberately scoped to the numeric arms only: for `==`/`!=`/truthy, nil is a
legitimate operand (`target_page.page_id != null` is exactly a nil probe); for a numeric operator
nil is never legitimate — it is always a config/data defect, and 313 is the measured cost of
routing on it silently.

- Read with `datahelpers.GetBoolFieldLoud` (WFA-010 — this is an author-facing key, its intended
  use).
- Public signature `evaluateStringCondition(expr, data, logger)` unchanged (two external test
  files call it); a `strict` variant threads the flag.
- An `ActionInputSpec` for `conditional` is registered declaring its full key set — measured
  fleet-wide 2026-08-19: live conditional steps carry **exactly** `condition`/`then_step`/
  `else_step` (145 uses each, top-level and nested), so the enumeration is complete —
  plus the new key. No `CheckConfig` runtime enforcement (adoption stays "one line away";
  turning validation on for 145 live steps is not this bug's blast radius).
- **Scope argued per the standing rulings, WFA-009's worked precedent:** owner 2026-07-29 §1 — no
  shared guarantee changes (absent key ⇒ identical behaviour); owner 2026-08-02 §2 — new authority
  on a shared seam as an opt-in field, unsafe default OFF; RFC_022 — all three conditions hold
  and the consumers are **enumerated, not asserted**: exactly one config names the key
  (`internal-linker.check_candidates`, via 490). Optional-key budget: `conditional` moves from
  unregistered to 1 optional key against N=10. Registered in the concept register (WFA-018/019)
  **in the same commit** (2026-07-28 ruling, condition 2).

### 4. What we are deliberately NOT doing

- **Not** teaching `resolveFieldValue` to synthesise `.count` for arrays (313 fix candidate 4):
  it changes what every conditional in the fleet evaluates, would flip any other
  currently-always-false condition of this shape to true in agents nobody is watching, and the
  census says the affected set is exactly one — the shared change buys nothing this bug needs.
- **Not** ranking candidates by relevance (298 candidate 3): with no cap the model sees every
  candidate, so ranking is prompt-ordering cosmetics; revisit only if a site's candidate set
  outgrows the payload budget.
- **Not** wiring the new audit into the daily CronJob in this change: the cron image carries a
  Python mirror (WFA-006 landmine) and its own parity contract; the wrapper script + register
  entry make the check runnable and discoverable now, cron wiring is named as a follow-up
  decision rather than smuggled in.

## Order of operations

1. Standing docs (this file) — done at start. 2. Write migration + Go + tests; `go build`/`go
test` on the affected packages, then the archive-tree build check. 3. Council submission (097)
before/alongside commit; expect the gate to admit it via the `platform/` files (the config-only
half alone would be refused — bugs_open/314). 4. Commit per task with explicit pathspec +
`Council-Submitted:` trailer. 5. Apply 490 by direct psql (ON_ERROR_STOP, the scoped-runner trap),
record in `schema_migrations`, re-read live config. 6. Verify at the artefact (below). 7. Update
bug files, register, WRONG_CALLS/LANDMINES as earned.

## Verification (the disconfirming pairs, both already half-established)

- **313:** `llm_call_log` for `agent_type='internal-linker'` gains its **first row in all
  history** on the next run with candidates (the "before" arm is zero, all-history, re-verified
  today — it cannot be faked); its `prompt_rendered` lists page names under `## Candidate Pages`
  (not `rows`/`count`/`columns` — the broken-prompt trap surfacing as evidence); a completed
  `orchestration_states` row with non-empty `collected_data->'candidate_pages'->'rows'` no longer
  ends at `complete_no_candidates`.
- **298:** the census (query's own predicate) returns **0 sites over the cap** because there is no
  cap; on a site with >15 candidates, a page sorting past position 15 appears in the rendered
  prompt.
- **Traffic:** 57 completed link jobs in 25 days ≈ >2/day, so a natural run should arrive within
  hours; verification reads the durable channels (`llm_call_log`,
  `orchestration_states.collected_data`), never the 15–90 s pod log.
- **Go halves after the next roll:** binary stamp ancestry check per service, then the audit run
  against the live export (clean fleet expected) and an induced `fail_on_non_numeric` failure in
  a test workflow — not before the roll (Go is inert until then).
