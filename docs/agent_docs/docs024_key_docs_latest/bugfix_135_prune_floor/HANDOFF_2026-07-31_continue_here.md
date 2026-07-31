# HANDOFF — bugfix_135_prune_floor, 2026-07-31 evening

**Cold start for this lane. Read this first; everything else in this directory is
detail.**

## State in one line

`bugs_closed/135` is **DONE** — fixed, live on chassis **v1.0.1218**, both branches
induced in production, closed and moved. The lane's remaining work is **not** 135:
it is `bugs_open/165`, filed at the council's direction and **unowned**.

## What shipped

| thing | where |
|---|---|
| The rule (pure counts, no SQL, reusable) | `platform/orchestration/actions/prune_floor.go` |
| Its tests (11, every branch incl. the refusal text) | `platform/orchestration/actions/prune_floor_test.go` |
| The one wired call site + cohort SQL + durable refusal | `code_symbols_actions.go` (`codeSymbolPruneCohorts`, `recordPruneRefusal`) |
| Read-side: "the index is not at one commit" | `diagnose_code_lookup_action.go` (`mixedCommitNote`), wired into BOTH readers |
| Register | **CTXA-025**, `docs026_concept_register/register/context-assembly.md` |
| Landmine | `LANDMINES.md`, "A green indexing run proves NOTHING about the prune floor" (synced to `doc_notes`) |

Commits: `10524a03c` (fix) · `a1225f9e3` (round-1 verdict) · `fefc79888` (close) ·
`c016c8bbf` (renumber + 016b §10 rows).

## What was proven, and how — do not re-derive this

Production pod-grep on both replicas, one `strings` pass each, **positive and
negative control in the same exec**. Then:

- **Refusal:** 400 synthetic `interface` rows at a bogus `commit_sha` →
  `prune_status=refused_floor`, `pruned=0`, all 5,392 rows intact, durable
  `doc_notes` row, `prune_floor_from_config: false`.
- **Pass:** cleaned up, 5 rows left above the floor → `pruned=5`, index restored
  **exactly** to baseline (4,992 rows, 592 paths, one commit, 0 leftovers).

**The finding that matters, and it contradicts the design rationale I wrote:**

```
kind=interface   8%  (33 of 433)   below -> REFUSED
distinct paths  60%  (592 of 992)  PASSED
```

The whole-repo signal I had documented as *stronger* is the one that would have let
the delete through. **Per-class cohorts are the load-bearing part.** Corrected in
CTXA-025, NOTES (11) and `WRONG_CALLS.md`.

## Council — read this before doing anything with the correlation

`SUBMISSION_CORR = 14239fa4-552f-4821-abaf-ea15ccee4ea5`. Two rounds, **both
REVISE**, and no `Council-Reviewed:` trailer has been claimed anywhere (correctly —
never claim one on a verdict you have not read as approved). Commits carry
`Council-Submitted:`, which asserts nothing and is resolved by `098` at report time.

- **Round 1** — `reuse_agent` HIGH: the plan never showed it had checked council
  `18fe4035` / migration 243's reader-side fix. **Answered** (four greps; it was
  body-coverage, the freshness half is a different lineage, nothing in the tree
  measures commit spread, and it was never considered). See NOTES (9).
- **Round 2** — `bug_historian` HIGH, on **scope not code**: a rigorous guard here
  while three identical unguarded deletes stay live. **Answered by filing
  `bugs_open/165`**, not by widening the patch (CLAUDE.md's seam ruling, `124`'s
  REJECTED verdict, two of the three files being edited by other lanes that day,
  and cohorts that would have been guesses). `guardian`'s medium answered by
  measurement; `debug_historian`'s medium (a pre-roll image grep cited as if it
  were a pod grep) was fair and is now moot.

**If you resubmit (round 3):** the code needs no change. What round 2 lacked is
now true — production pod-grep, both branches induced, and the sibling objection
filed. Use `RESUBMIT_CORR=14239fa4-552f-4821-abaf-ea15ccee4ea5`. This is optional:
the gate is advisory, the fix is proven live, and the outstanding objection is
about *other* call sites.

## The actual next work: `bugs_open/165`

Three sites still delete-what-they-did-not-see with no completeness check:

1. `save_page_sections_action.go:532` — `page_components`. **Start here.** It has
   lost real customer-facing content twice (016b §9 cases 1–2; case files 001, 037,
   038, `bugs_closed/058`). Its `pageComponentAgentWritableSQL` guard is an
   **authority** check ("may I delete this row?"), **not a completeness** check — a
   writer that returned two sections instead of twelve passes it perfectly.
2. `populate_nav_tables_action.go:147,150` — `site_nav_items` / `site_nav_groups`.
3. `site_db_actions.go:1474` — `link_registry`.

**Reuse `evaluatePruneFloor`; do NOT copy 135's cohorts.** 135's were defensible
only because they were chosen *after* reading the live distribution. Guessing is how
you build a guard that fires on legitimate edits and gets deleted by the first
person it blocks. Before starting, check the tree for in-flight WIP on those files —
`who-owns.py` reads commits and cannot see a session mid-fix.

## Traps this lane hit (full accounts in NOTES + WRONG_CALLS)

- **`go test ./platform/orchestration/actions/` failing is probably not your fault
  or HEAD's.** Twice it was another session's uncommitted refactor in the shared
  tree. Verify with `git archive HEAD` + only your own files overlaid; a compile
  error names the file it *fails* in, not the file that *changed*.
- **`go build ./...` fails at HEAD** for an unrelated reason (two packages in one
  directory under `traffic_probe/deploy_setup/working_dir`). Build
  `./platform/... ./internal/... ./pkg/...`.
- **`ls bugs_* | sort -n | tail -1` is the max, not the next free.** I filed a
  duplicate `162`; renumbered to `165`.
- **A roll kills an in-flight council round.** Sequence submissions and deploys.
- **`docker push` is blocked by the auto-mode classifier** — the owner runs it, or
  you wait for another lane's build from HEAD (which is what happened here).
