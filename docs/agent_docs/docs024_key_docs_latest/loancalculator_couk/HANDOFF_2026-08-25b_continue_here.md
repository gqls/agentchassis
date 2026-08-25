# HANDOFF — loancalculator.co.uk · 385's CAUSE IS ESTABLISHED; the fix and one owner decision are what remain (2026-08-25, evening)

> Supersedes `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/HANDOFF_2026-08-25_continue_here.md`.
> That file's one open item — 385's cause — is **answered, first-hand, at every joint**:
> `bugs_open/385_HANDOFF_2026-08-24_a_rebuild_appends_an_unlinked_copy_of_the_locked_section_it_just_repositioned.md`
> **§5c** is the single source. Investigation record: NOTES `## 2026-08-25 (second
> session)`. Owner prose: `README_where_we_are.md` (same date). New LANDMINES entry:
> *"A locked interactive row whose `build_status` is `'deployed'` ARMS a silent
> duplicate…"* (verifier dispatched, corr `3d20ad68`). Pattern: 016b §9 (2026-08-25,
> the N−1-matchers entry).

```
site        loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
pages       28 active · locks 11 held (12 rows incl. archived standard-calc's)
harness     ✅ toolgolden --selftest green 2026-08-25 (this session, before any measurement)
golden      acceptance/GOLDEN_2026-08-24_post_385_repair_tool_values.json — still current
385         CAUSE ESTABLISHED (§5c) · damage repaired 08-24 · writer STILL LIVE and
            STILL ARMED on exactly ONE row fleet-wide: the victim's own locked row
```

## The cause, in three lines (the full chain with evidence is §5c — read it before coding)

`save_page_sections_action.go`'s **Layer 2 interactive-preservation** block preloads
stored rows `WHERE build_status='deployed' AND <interactive>` and pairs them to the
incoming set by **exact slot-name string only** (`:551-558`) — no identity arm, though its
own SELECT fetches `component_id` and `cc.function`. On the **build arm** incoming names
are the plan's FUNCTION names, so a positionally-named locked calculator reads as
"dropped" and gets a verbatim copy appended (`ComponentID:''`, RFC_046 opt-in OFF); the
lock guard — working correctly — cannot pair the copy, and it lands as the byte-identical
`NULL`-component orphan. `[MEASURED 2026-08-25]` the armed set (locked + `'deployed'` +
interactive) is **1 row fleet-wide: `tool-loan-vs-savings`/`tool-2`, the victim itself.**

## Next actions, in order

1. ~~The code fix~~ **DONE 2026-08-25, same session: commit `a799579fd`**
   (`matchPreservedSectionIdx` — identity → slot exact → slot kebab → function, with
   one-row-claims-one-section consumption; six unit tests mutation-verified RED on
   slot-only, wiring scan, whole-action pin; the 357 lane's four Layer 2 provenance
   tests pass unchanged — their own comments anticipated this as "the identity
   round"). `verify-head-builds.sh` OK at `a799579fd`. **Council verdict OWED:**
   correlation `ece638fb-ec12-4a73-a1d1-c80d96d22ff4`, committed with
   `Council-Submitted:`. **VERDICT ARC COMPLETE, same session:** round 1 REVISE
   (reuse_agent's gate — a third hand-mirrored copy of the arms is the drift class;
   correct) → revised: the relation extracted to `datahelpers/slot_pairing.go`
   (register **LOCK-009**), all three matchers now adapters, wiring scans in both
   packages; commits `b9d0f02be` + `3552e674b`, `verify-head-builds` OK → **round 2
   APPROVED** (2 advisories, none high; dispositions recorded in LOCK-009). Nothing
   further owed to the gate.
2. **The fix is INERT until an image rolls.** "Has it shipped?" is a query: chassis
   `build provenance` stamp, then `git merge-base --is-ancestor a799579fd <stamp>`.
   Until that passes, the one armed row can still duplicate on a build-arm rebuild.
3. **Owner decision (asked in README):** data-side disarm today — flip the one armed
   row's `build_status` `'deployed'` → `'approved'` (matches its ten serving locked
   siblings). Removes the only armed instance without waiting for a roll; touches a
   human-locked row, so it is not a session's call.
4. **Verification of the fix is BUILD-ARM ONLY** (bug §9): the rerender arm structurally
   cannot reproduce this (its incoming names ARE slot names). Do not rebuild
   `tool-loan-vs-savings` through `needs_page` until the fix is live in the stamp (or
   the disarm has been applied) — that exact dispatch is the reproduction.

## Standing cautions (carried)

- `toolgolden.py --selftest` green FIRST, or nothing measured afterwards is quotable.
- Prove a deploy at the artefact; per SERVICE, not per fleet.
- ⚠ `UPDATE page_components SET position` does NOT touch `updated_at`.
- ⚠ Before any repair of 385's shape, check `pages.sections` for a stale sixth entry.
- A single sample mid-wave proves nothing (08-24's false "publish failed").
- Hand-filed / un-parked work items must be `triaged`; the dispatcher cannot see `detected`.
- `retract_page_deployment` refuses active pages and its default selection takes
  `tool-standard-calc` — explicit `page_ids` always.
- Query runs BY CORRELATION, never `now()`-interval; collected_data can purge in ~2h.
- Before any planner run: the four cautions in `HANDOFF_2026-08-23_continue_here.md`.
