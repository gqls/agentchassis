# PLAN — bugs_open/165, completeness floors for the three remaining reconciliation deletes

**Started 2026-07-31 evening.** Lane opened to work `bugs_open/165`, which the
`bugfix_135_prune_floor` lane filed at the council gate's direction and left
unowned.

## The shape of the work

`bugs_closed/135` built the RULE (`prune_floor.go`, register **CTXA-025**) and
wired one call site. Three mechanistically identical deletes remain:

| site | file | table | state |
|---|---|---|---|
| **A** | `save_page_sections_action.go:532` | `page_components` | **DONE 2026-07-31** — `ecf738002`, awaiting verdict + roll |
| B | `populate_nav_tables_action.go:147,150` | `site_nav_items` / `site_nav_groups` | OPEN |
| C | `site_db_actions.go:1474` | `link_registry` | OPEN |

Order is by stakes, not by convenience: A is the only one of the three that has
actually lost customer-facing content.

## Decisions, and why

**1. Reuse the rule; measure the cohorts fresh. (Decided at the start, and it
turned out to be the whole of the difficulty.)** `evaluatePruneFloor` is used
unchanged. Nothing about the shared mechanism moves, so this is a normal
council-gate change rather than architecture scope (owner ruling 2026-07-29 §1:
the RFC trigger is a change to what the shared mechanism GUARANTEES). What is NOT
reusable is 135's cohort choice — its per-kind partition was defensible only
because it was chosen after reading the live distribution, and the same is
required here.

**2. Per-slot cohorts were REJECTED on the data — including the shape the bug
file itself proposed.** `165` suggested "plausibly one cohort per `slot_name`".
Measured: 998 of 1,009 `(page_id, slot_name)` groups hold exactly one row. Every
per-slot cohort would be 1 stored, so a legitimate single-section removal scores
that cohort 0% and refuses the save — 89 real shrinkages in 4.5 months, each one
blocked. The bug file's own warning ("a floor that fires on legitimate edits is
worse than useless") applied to its own suggestion.

**3. Two cohorts, in genuinely different units.** Rows (what this save inserts vs
what the DELETE would remove) and plan (vs `pages.sections`). The plan cohort
exists to break the **ratchet**: once a truncating writer has cut a page from
twelve rows to two, the row cohort reads 2/2 = 100% for ever and the damage
becomes the new baseline. `pages.sections` is written by seven other actions and
never by this one, so it is a real second opinion rather than an echo.

**4. The refusal fails the whole save, unlike 135's.** 135 refuses the prune and
lets the run continue — self-healing, because a later healthy run prunes what was
retained. Here the delete and the insert are one operation: refusing only the
delete and still inserting would write new sections alongside the old, which is
`bugs_open/156`'s duplicate-`page_components` defect. So the save is refused
outright, exactly as the three sibling regression guards in that file already do.
The cost of a false positive is therefore higher here than in 135, which is why
the false-positive rate was measured before the floor was chosen and not after.

> **CORRECTION 2026-07-31, made during the work, not after it.** The first plan
> denominator was the raw planned section count. It was wrong, and the way it was
> wrong is the useful part: it counted planned slots that an active lock makes
> unwritable. A perfect rebuild of `idea.uk/index.html` (6 planned, 4 locked)
> writes 2 sections and would have scored 2/6 = 33% — **refused**. That is a guard
> that blocks every rebuild of precisely the pages a human cared enough about to
> lock, which is how a guard earns being deleted by the first person it blocks.
> Caught by opening the six rows instead of trusting the count that said "live=2".
> Denominator is now `planned − suppressed − locked`; trips on 0 of 238 reachable
> pages. Pinned by `TestPageSectionFloorPassesAHealthyRebuildOfALockedPage`.

## Phasing

- [x] **A** — measure, choose cohorts, implement, test, submit, commit.
- [ ] **A** — read the verdict (`a54172b6-9756-4abc-a9e0-f173ad4de779`), then
      induce both branches in production after the next roll. A green run proves
      nothing: the floor is inert on healthy input by design.
- [ ] **B** — blocked on the `bugfix_149_nav_membership` lane, which had
      `populate_nav_tables_action.go` dirty in the tree on 2026-07-31. Its cohorts
      must be measured, not copied: nav rows are per-site, not per-page, and the
      landmine already recorded against that file (nav-updater deletes 16 rows
      across 7 domains that nothing recreates) suggests the denominator question
      there is different again.
- [ ] **C** — `link_registry`, lowest stakes, regeneratable.

## What "done" means for each site

The bar `135` was held to, restated: **a green run proves nothing.** Induce the
fault, watch the refusal fire with its numbers, confirm nothing was deleted, clear
the induction, confirm a normal run still prunes. Both branches, in production,
pod-verified.
