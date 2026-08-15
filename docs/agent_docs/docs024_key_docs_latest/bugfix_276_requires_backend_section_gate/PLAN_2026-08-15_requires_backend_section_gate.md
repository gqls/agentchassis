# PLAN — bugs_open/276: gate `requires-backend` sections out of generic planning

## Origin and correction to the brief

`bugs_open/276` (filed 2026-08-14, council-directed) and the concept register (VMB-010) both
name exactly one call site to fix: `build-site-planner`'s `load_components` step. **That
undercounts the problem.** Fleet-wide enumeration of every `agent_definitions` step whose
query touches `content_components` (14 hits, all read and classified 2026-08-15) found THREE
"menu" queries an LLM planner reads to choose section components — and `build-site-planner`
is the minority path:

| agent type | step | dispatches (30d, re-measured 2026-08-15 14:03Z) | gate before this work |
|---|---|---|---|
| `content-gap-planner` | `load_available_components` | **131, most recent today** | none |
| `build-site-planner` | `load_components` | 2 | none (has an unrelated *tool*-level gate, migration 407) |
| `site-planner` | `load_available_components` | **0, ever** (unbounded check) | none |

Fixing only the named instance would leave the dominant real placement path (`content-gap-
planner`, 131 vs 2 runs) completely open. This plan fixes all three.

No live damage exists today — the only `intent-probe` placement is `relojistas.com`/`index`,
already backend-capable. This is a forward gate, not a repair.

## Decisions (stated, so nothing is a silent gap)

1. **Three separate migration files**, one per agent type — matches house style (406, 407,
   410 are each one-agent).
2. **Id-scoped UPDATE + byte-exact pre-state gate**, not the bare type-scoped `WHERE` in
   406/407's own UPDATE bodies. 406's own addendum records this as council-stated guidance
   for *future* migrations; the agent row for `content-gap-planner` has been edited twice by
   a concurrent session in the hours around this plan (migrations 413, 414 — confirmed to
   touch `plan_gaps`/the model field, not `load_available_components`) — exactly the
   collision risk id-scoping exists to neutralise. A byte-exact pre-state check turns any
   surprise into a loud abort, not a silent double-apply or an overwrite of someone else's
   concurrent edit.
3. **`content-gap-planner` gets a full capability-checked gate** (`$1` = `site_record.site_id`).
   Proof the binding resolves: this workflow is strictly linear `ensure_site_record` →
   `load_specs` (binds `site_id: site_record.site_id`) → `load_existing_pages` (same) →
   `load_available_components`. Two sibling steps in this exact chain already resolve it, and
   a direct measurement confirms **131/131** recent `orchestration_states` rows for this agent
   carry a non-null `collected_data#>>'{site_record,site_id}'`.
4. **`build-site-planner` gets the same clause**, reusing the `$1` already bound since
   migration 407 (no params change) — added as a top-level `AND` so it applies uniformly to
   both the section/element branch and the already-placed-tool branch.
5. **`site-planner` gets an unconditional exclusion, no capability check.** Zero dispatches
   ever, and no `ensure_site_record`-equivalent step exists anywhere in its (short, 4-step)
   workflow — there is no proven site-id binding to use. A params path that resolves to nil
   hard-fails the step (`database_actions.go` — query param path resolves to nil is fatal),
   so guessing an unproven binding on dead code risks planting an outage-class landmine for
   zero present benefit. Decisive call: strip the tag unconditionally instead — closes the
   door completely, costs nothing, and a future reviver of this agent does the binding
   legwork properly and earns the opt-in like the other two.
6. **No architecture RFC.** Two live owner rulings already cover this shape: an opt-in tag
   whose producer/consumer set is named in the register entry doesn't need one
   (2026-08-02 §1), and new authority on a shared seam as an opt-in field with the unsafe
   side defaulting OFF is the *prescribed* remedy (2026-08-02 §2). Both hold here, exactly as
   they already held for the tool-side gate (406).
7. **VMB-010's "audit check" (Later item) stays deferred, out of scope.** Nothing live to
   audit today (checked); it's Go discovery-check work (image rebuild) outside this
   config-only window; the council's actual ask in 276 was gate-parity with the tool half.
8. **Council submission: `FORCE=1`, not a pseudo-edit anchor.** The scope pre-filter needs
   ≥1 edit under `platform/`/`internal/`/`pkg/`; this fix is genuinely config-only (no Go
   bytes change). Naming a Go file as a "config_change" edit whose sketch is really just an
   observation is the exact shape ("a fix plan proposes changes, not observations") that got
   a real submission refused server-side on 2026-08-14 (round 2, corr c78ed496). Direct
   precedent exists for the honest alternative: migration 406 itself (this bug's sibling, the
   tool-side gate) was submitted `FORCE=1` with the stated reason "config ships as a docs-path
   migration" and drew a full three-round council review ending APPROVED. Follow that
   precedent exactly.

## Migration numbers

Claimed live 2026-08-15 ~14:05Z after re-checking (410–417 all taken by concurrent sessions,
including a same-number collision at 415 between two other threads): **418, 419, 420.**
Re-verify immediately before each file write — this tree has shown new claims roughly every
few minutes today.

- `418_content_gap_planner_requires_backend_gate.sql` (+ `_ROLLBACK.sql`)
- `419_build_site_planner_requires_backend_gate.sql` (+ `_ROLLBACK.sql`)
- `420_site_planner_requires_backend_gate.sql` (+ `_ROLLBACK.sql`)

## Verification plan

Disagreeing-pair proof per migration, run directly via `psql` against the read-only current
rows before `apply`, then re-run against the live rows after:
- 418/419: compare OLD vs NEW query output for a backend-capable site
  (`relojistas.com` = `ecf15e75-a966-4900-bcb0-1c85f689dbfd`) and a static control
  (`gamesdesign.co.uk` = `e33263f4-74f8-494f-b191-546845dbbddf`). Backend site: identical
  row sets. Static site: NEW is missing exactly one row (`intent-probe`), nothing else
  differs.
- 420: OLD vs NEW, no site id — NEW missing exactly `intent-probe`, unconditionally.
- Mutation-test each post-apply verify `DO` block against the *current ungated* row first
  (proves the check can actually fail), per house practice.

## Docs / records to update on completion

- `bugs_open/276` — TAKEN UP section, naming this workstream and the three-call-site
  correction, closure bar = all three applied + verified + council resolved.
- `docs/agent_docs/docs026_concept_register/register/vm-backend-sites.md`, VMB-010 — append
  (never rewrite) a dated status block.
