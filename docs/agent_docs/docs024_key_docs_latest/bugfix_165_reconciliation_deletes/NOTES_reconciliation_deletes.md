# NOTES — bugs_open/165 completeness floors

Append-only, newest at the bottom. The missteps are the point.

---

## 1. Picking the lane up (2026-07-31 ~19:30 BST)

Handed the `bugfix_135_prune_floor` handoff. 135 itself is closed and live
(v1.0.1218); the remaining work is `bugs_open/165`, unowned.

Ownership check before touching anything, because `who-owns.py` reads commits and
cannot see a session mid-fix:

- `save_page_sections_action.go` — clean in the tree. **Free.**
- `site_db_actions.go` — clean. Free.
- `populate_nav_tables_action.go` — **dirty** (the `bugfix_149_nav_membership`
  lane). Site B is live territory; leave it.
- Grepped the live `.jsonl` transcripts for the target's code symbols. The 135
  lane's own session (`f0fe4678`) had 102 hits but its last entry was 18:22Z —
  it filed 165 and finished. Nobody else is on it.

Started with site A per the handoff.

## 2. Measuring before choosing cohorts

`165` is explicit that the cohorts must not be guessed, so nothing was written
until the distribution was read. Four things came out of it:

- **Pages are SMALL.** Of 409 pages with agent-writable rows: 178 have 1, 51 have
  2, 98 have 3, max 8. A per-class partition of a 3-row page is not a partition.
- **`slot_name` is 1:1 with a row** — 998 of 1,009 groups hold exactly one. This
  kills the per-slot cohort the bug file itself proposed. 89 legitimate shrinkages
  in 4.5 months would each have been refused by it.
- **`component_id` is the row count in a costume** — 365 of 409 pages have as many
  distinct components as rows. No independent signal.
- **The existing guards are blind on a third of the corpus.** The content-regression
  and interactivity guards both scope to `build_status='deployed'`, and **142 of
  426** pages with components have no deployed row.

Historic false-positive rate for the row cohort, from `page_component_history`
consecutive overwrite pairs: 2,620 transitions, 89 shrank, **4 below 0.5**
(0.15%), 1 below 0.34, 0 below 0.25. [PROXY — the "after" is the next event's
snapshot, which excludes empty-`rendered_html` rows and includes locked rows.]

## 3. THE MISTAKE THAT MATTERS: I read a count as damage

The plan-side cohort simulation said 3 pages would trip. One was
`idea.uk/index.html`: planned 6, "live 2". I wrote that down as evidence the
cohort was **finding real damage** — a homepage stripped to two sections is
exactly the defect this guard exists for.

It was not damage. Opening the rows instead of trusting the count:

```
hero · brief-explanation · tool-list · call-to-action · latest-news · info-card-grid
```

Six components for six planned sections. **Four of them are locked.** My
`live` count applied the agent-writable predicate, so it counted 2 — correctly —
and I read "2" as "the page has 2 sections" when it means "2 sections this save
may touch".

The consequence was in the design, not the prose: my plan denominator was the raw
planned count, so a **perfect** rebuild of that page (6 sections handed over, 4
swallowed by locks, 2 written) scores 2/6 = 33% and is **REFUSED**. A guard that
blocks every rebuild of precisely the pages a human cared enough about to lock.
That is not a false positive to tune away later — it is the failure mode `165`
warns about in terms ("a guard that cries wolf gets deleted by the first person it
blocks"), aimed at the most curated pages on the estate.

Fixed: denominator is `planned − suppressed − locked`. Re-measured: **0 trips on
238 reachable pages** (the 2 remaining are `rebuild_policy='owned'`, refused ~370
lines earlier). Pinned by `TestPageSectionFloorPassesAHealthyRebuildOfALockedPage`,
and mutation M1 confirms that test fails the moment the locked term is removed.

**What would have caught it cheaply:** the same thing that did — reading the six
rows rather than the one number. A count that has a predicate in it answers the
question the predicate encodes, not the question you asked it.

## 4. Smaller wrong turns, recorded because they cost time

- **The step-level consumer census returned 0, then 3, when the answer is 6.**
  `step->>'action_type'` is the wrong key (it is `action`), and even corrected the
  top-level census misses `pageflow-builder` and `page-rebuild`, which nest the
  step inside a loop. The `086/087` landmine says exactly this and I hit it anyway.
  Matching the literal key text `"action": "save_page_sections"` gives 6, which
  agrees with the claims guard's own comment ("six live agents persist sections
  through here") — a cross-check I should have looked for first.
- **First mutation-test pass was invalid and looked fine.** The backup path was
  wrong, `cp` silently failed to restore, and M2/M3/M4 each ran on top of the
  previous mutant. Every one "failed as expected", which is exactly what a valid
  run looks like. Re-ran from a fresh baseline each time; results held, but they
  had not been *evidence* until then.
- **The shared tree would not build,** and it was not mine:
  `discovery_checks/check_empty_sections.go:249: undefined: datahelpers`, another
  session's in-flight edit. The 135 handoff warns about exactly this. Worked around
  with a 12MB `git archive HEAD go.mod go.sum platform internal pkg` tree rather
  than touching another session's file.
- **`/tmp` (shared 16G tmpfs) hit 100% mid-session** and commands began failing
  with ENOSPC. A full `git archive HEAD` needs ~350MB and died half-way. Asked the
  owner rather than deleting other sessions' scratch unilaterally; cleared 74
  scratch dirs belonging to sessions idle >6h (~4.2GB) on his say-so.

## 5. What shipped

`ecf738002` — `save_sections_prune_floor.go` (+17 tests), wired into
`save_page_sections_action.go`. Council `a54172b6-9756-4abc-a9e0-f173ad4de779`,
committed with `Council-Submitted:` since the verdict had not landed.

Not yet proven live: the code is inert until the next chassis roll, and **a green
run proves nothing** — the floor is inert on healthy input by design. Both branches
still need inducing in production, per `165`'s own verification bar.
