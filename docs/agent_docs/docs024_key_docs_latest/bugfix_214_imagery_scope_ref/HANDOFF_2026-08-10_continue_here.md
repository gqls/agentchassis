# HANDOFF — bugfix 214, imagery `scope_ref` · cold start for a fresh chat

**Written 2026-08-10 ~22:20Z.** Read this first, then `NOTES` for the evidence trail and
`README_where_we_are` for the plain-prose account.

---

## One paragraph

`WriteSitePlanAction` canonicalised page identity for `site_plan_pages.name` and
`site_plan_sections.page_name` but copied the planner's raw map key verbatim into
`site_plan_imagery.scope_ref` — same function, ~60 lines apart. So `about` became
`about-index` for the pages and stayed `about` for the imagery, and the imagery row named
a page its own plan did not contain. Nothing looked broken: consumers select the plan row
by `scope_ref` then join the asset by `asset_key`, so the failure mode is an asset that was
planned, generated, deployed, **paid for** and referenced by nothing. **Fixed at the write
path, live on `v1.0.1283`, and the 9 rescuable live rows are repaired.**

## STATUS: essentially DONE. Three things outstanding, none blocking

| # | outstanding | why it is not done |
|---|---|---|
| 1 | **Council verdict** | Round 1 died on the fleet credit outage. **Resubmitted** and RUNNING at time of writing (`gate_tooling_provenance`). Read it and act on REVISE/REJECTED — the code is already on shared HEAD. |
| 2 | **The rewrite arm has not been OBSERVED firing in production** | Two induced runs were non-discriminating (below). Not a defect — it needs a plan containing a renamed page. |
| 3 | **`mortgagecalculator.co.uk` `tools-index`** | Names a page existing under no spelling. Needs a human decision; deliberately left. |

**Do NOT re-apply `sql_for_agents/373`** — already applied, and its guard will abort
(it requires exactly 1 unresolvable row to remain, and that is now the state).

## What is proven, and how

- **Live on `v1.0.1283`, BOTH replicas** — binary grepped: 4 added literals present,
  pre-existing control present, **fabricated negative control 0**.
- **Backfill applied**: 5 page-scope + 4 section-scope repaired; census **10 → 1**; the
  `DO`/`RAISE` guard fired its success notice.
- **Repair proven at the CONSUMER's own join**, run verbatim from
  `plan_sections_action.go` — `gamesdesign about-index → hero_about`,
  `contact-index → hero_contact`, `fundamentallyai news-index → hero_news`, and all four
  gamesdesign icons through the `LIKE` join. **All returned nothing beforehand.**
- **The new code path executes in production**: the induced run's `plan_written` return
  carries `imagery_refs_canonicalised` / `imagery_refs_unresolved` /
  `imagery_duplicates_merged`. Those keys **cannot exist without this change**.
- **Wiring is mutation-proven**: delete the resolution block from `WriteSitePlanAction` and
  `write_site_plan_imagery_wiring_test.go` fails on both arms. The 16 unit tests do **not**
  catch that mutation — measured, and the reason the wiring suite exists separately.

## What is NOT proven — read this before writing "verified" anywhere

- **The rewrite arm firing in production.** Both induced pool-site runs returned
  `imagery_refs_canonicalised: 0`, because the planner emitted only `content`-role pages
  and `CanonicalisePage` renames none of those. **That zero would have been zero on the old
  binary too**, so it is not evidence about the fix. It IS evidence of no regression.
  - **To close it:** watch the first replan of a site that has `-index` pages —
    `gamesdesign.co.uk`, `dartsonline.com` and `robot-hands.com` all qualify. Then
    `plan_written.imagery_refs_canonicalised` should be **> 0**, and RUNBOOK R3 should
    return no row whose page part is a raw alias of a page that plan contains.
  - **Do not force this by replanning a customer site** to get the number. That rewrites a
    live site's plan for a test, which is not this lane's to do.
- **`mortgagecalculator.co.uk`'s repaired refs are not "working"** — its assets do not
  exist yet (`asset_exists=f` on all 7). The reference is correct; nothing is visible until
  imagery is generated.

## Two traps this lane paid for — both already written into the corpus

1. **The census predicate belongs to the READER.** Resolving `scope_ref` against
   `site_plan_pages` (the writer's table) gives **22**; against `pages.name` (what all ten
   consumers join) gives **10**. Both are legitimate — 22 is plan-consistency, 10 is live
   damage — but quoting 22 as damage overstates it, and a backfill built on 22 would have
   **repointed twelve working heroes**. In `WRONG_CALLS.md` and the LANDMINES entry.
2. **A green test suite that cannot see its own fix deleted.** Proven by mutation, not
   assumed. In `WRONG_CALLS.md`.

## Residue I created and cleaned up — check this if something looks odd

To get a production run I dispatched `build-site-planner` at **two pool sites**
(`pool-ai-agents.internal`, `pool-energy-utilities.internal`; `status='pool'`, nothing
serves them) rather than at a customer site. Each run wrote a plan, 5 pages and a queue of
work items — **41 open items in total, including 24 `needs_imagery`**, each of which would
have triggered a paid image generation on a throwaway site.

**All 41 cancelled**, verified by a `DO`/`RAISE` guard asserting 0 remaining. The plans and
(undeployed) pages remain and are harmless — they do **not** pollute the R1 census, because
the planner's own `sync_pages` step created matching `pages` rows, so their refs resolve
(checked: census still 1).

> **The general lesson, worth carrying:** inducing a run on a "safe" site is not free —
> the pipeline queues work behind it. Check `site_work_items` after any induced dispatch.

## Key facts and identifiers

- Commits: `c21af5eda` (fix) · `c90212df6` (gofmt) · `b689e9122` (docs) · `a4cf6a195`
  (council death) · `6abd186ee` (reconciliation) · `477a38eef` (roll + backfill evidence).
- Council trail correlation: **`46a50b4c-f00d-4492-b7fd-ce5dc2023480`** (use this, not the
  run id). Round 2 run `adba954d-599a-4913-98dc-c65fee1bb095`, orch
  `8a54fbc4-c376-4638-a60c-527df468daf7`.
- Register: **IMG-070** (`register/imagery.md`), index row landed via another session's
  commit `4451b2a0a`.
- Code: `platform/orchestration/actions/write_site_plan_imagery_scope.go` (+ two test
  files); wiring is block **2c** of `WriteSitePlanAction`.
- **Not done deliberately:** exporting `datahelpers.NormaliseSlug`. `page_canonical.go`
  carried another session's uncommitted `FlatURLs` work and a pathspec commit cannot exclude
  a same-file passenger. Routed through the exported `CanonicalisePage` instead, coupling
  pinned by `TestNormalisePageKey_MatchesTheNormalisationValidateRolesApplies`. **If that
  work has since landed, switching to a proper exported `NormaliseSlug` is a tidy
  follow-up** — the test will keep you honest either way.
- **Architecture-scope, NOT taken:** RFC_016 §1's move of imagery inside the section entry.
  That is the only real fix for ordinal semantics; this lane deliberately validates ordinals
  and never rewrites them (see IMG-070's landmine note for the three reasons).
