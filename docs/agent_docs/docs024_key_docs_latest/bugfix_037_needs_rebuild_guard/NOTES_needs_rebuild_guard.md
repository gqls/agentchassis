# NOTES — bugs_open/037 needs_rebuild guard (append-only, newest at the bottom)

## 2026-07-21 — diagnosis

- Read 037, `/bugs_closed/001`, `/bugs_open/050`, `/038`, `/039`, `/051` and the memory
  `replan-clobbers-built-pages`.
- Grepped every setter of `build_status='needs_rebuild'`: `v3_site_actions.go:644`,
  `maintenance_actions.go` `flagPagesForRebuild`, `store_generated_component_action.go`
  `markPagesForRebuild`, `discovery_checks/check_unresolved_sections.go`. **All preserve
  `pages.sections`; none means "recompose".** The last two flag `needs_rebuild` *so the page renders
  a component its existing sections already name* — recomposition would defeat the very reason for the
  flag. This refutes the handoff's "For leaving it as-is" argument (that `needs_rebuild` is fix step
  4's explicit redesign intent). Decision: candidate 2.
- Live DB: 26 active `needs_rebuild` pages; 19 with a real composition (protected case), 5 empty (2
  `section-index` awaiting composition with 0 components; 3 robot-hands `tool` rendered-elsewhere with
  1 component — 050's case).

## 2026-07-21 — the file was a live minefield (misstep recorded)

- `v3_site_actions.go` was **modified in the working tree** by another session: uncommitted work for
  `/bugs_open/040` (partial-build `pageSectionShortfall`), `/041` (kebab lookup helpers in
  `component_validation.go`), and **`/050`** (deployed-empty gate in Pass B/B2). The whole `actions`
  package was dirty across dozens of files but **compiled** as a whole.
- **Key constraint discovered:** the working-tree `v3_site_actions.go` called `sectionLookupValueSet`
  etc. defined only in the *also-uncommitted* `component_validation.go`, so committing
  `v3_site_actions.go` alone would have (a) swept the 040/041/050 passenger and (b) **broken the HEAD
  build**. There was no clean file-by-file commit path.
- **MISSTEP:** I tried to prove the tests discriminating by neutralising the predicate **in place** in
  the shared file. The edit silently did not persist — a concurrent **owner sweep commit**
  (`fe2ba5e52`, `v1.0.1146`) rewrote/committed the file underneath me mid-experiment. Lesson: never
  run a throwaway in-place edit on a file under active concurrent commit; use an isolated
  `git worktree` (which is what finally proved discrimination). Logged to `WRONG_CALLS.md`.
- **What the sweep did:** the owner's `git add`-style sweep took the entire `actions` WIP — including
  **my** `realisedPageCompositionIsPreserved` change — into `v1.0.1146` and rolled it fleet-wide. So
  the fix landed via someone else's commit. Nothing lost (forward-only); attribution is the sweep's.
  My tests were NOT in that snapshot, so I committed them separately (`9864fab37`).

## 2026-07-21 — verification

- Unit tests: 4 new, all pass against HEAD; proven discriminating in an isolated worktree (three
  membership tests FAIL without the `needs_rebuild` term; `NeedsRebuildEmptyPageIsStillComposable`
  FAILS if `realisedPageIsBuilt` is naively widened instead of adding a separate predicate).
- Live: whole fleet on `v1.0.1146`; `strings /app/agent-chassis | grep -c
  realisedPageCompositionIsPreserved` = 1 in `agent-chassis-55bbccfdbc-xrkv6` (positive control
  `reconcilePlanWithRealised` = 2, negative control = 0). **Fix is live.**
- **NOT done:** a live re-plan to watch a real `needs_rebuild` page's `pages.sections` survive
  (mutates a site + ~30 min + spend) — [UNVERIFIED — live re-plan], pending owner go.

## 2026-07-21 — landmines for the next thread

- Do not widen `realisedPageIsBuilt` for this; it drives 050's empty-gate. Membership uses the
  separate `realisedPageCompositionIsPreserved`. The two questions are different and a
  `needs_rebuild` empty page answers them differently.
- `/tmp` is a 16 GB tmpfs and was **95% full** — `go test` failed ENOSPC until I set
  `TMPDIR`/`GOTMPDIR` to a repo-local dir.
