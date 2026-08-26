# HANDOFF — webdesign tool rebuilds, THE PLATFORM SEAT. Written 2026-08-25 ~19:50Z.
# ⚠ SUPERSEDED 2026-08-26 by HANDOFF_2026-08-26_platform_seat_continue_here.md — start THERE.
# This file is kept for its correction trail (the "sweeps are scheduled" premise and its fix).

**TWO SEATS WORK THIS LANE. This file is the PLATFORM seat's thread** (session name
`webdesign-tool-rebuilds`, WITH the trailing s — ran Phase A #16–#28, TL-048, bugs_closed/362,
Track 2). **The GRIND seat** (`webdesign-tool-rebuild`, no s — currently at ~41/63, ~5 tools/day)
owns the rebuild queue, the add_tool dedup key, and the lane's START-HERE:
`HANDOFF_2026-08-25_continue_here.md` (theirs). Do not file add_tool items from this seat; announce
before touching any file the grind writes. Evidence for everything below: `NOTES_…` (newest at
bottom, entries 2026-08-25); the read-aloud account:
`SUMMARY_2026-08-25_two_thirds_done_and_the_platform_pays_us_back.md`.

## State of this seat's work — ALL LANDED, nothing mid-flight

- **Track 2 is DONE and LIVE**: contract rules 16/17 in `tool_health` (forks→improve_tool; ported
  residue→ONE `capability_gap` per site) + the tombstone slot filter. Council trail
  REVISE→APPROVED (`21540c8e`, r2 15:29Z 08-25); LIVE on **v1.0.1339** (stamp `a7459a44b`, both
  commits ancestors). The round-2 reuse advisory was acted on post-approval: the filter is
  CENTRALISED in `toolEligibilityWhere` (commit `f44451494`, functionally identical to what 1339
  runs; rides the next roll).
- **bugs_closed/362** (link repair in all three tool writers): both rounds approved, live since
  1337, census clean. **bugs_closed/331/103-successor work**: TL-048 (composed-only tool meta)
  approved + live since 1321.
- **Council coverage**: every platform commit from this seat carries a `Council-Submitted` trailer
  on an APPROVED correlation — 098 credits them; nothing owed.

## What the NEXT session of this seat should do, in order

1. **Run the two Track-2 demand controls once a discovery sweep has fired post-1339** (the sweeps
   are scheduled; do not force one):
   - `SELECT summary, spec->>'population', spec->>'residue' FROM site_work_items WHERE
     item_key='capability_gap:tool_health_contract_rules' AND site_id='6b49db8e-…';` → expect ONE
     open row for webdesign whose residue ≈ the grind's not-yet-rebuilt count (shrinking); other
     sites appear as their sweeps run.
   - `tool_acceptance` findings on webdesign must NOT enumerate the 41+ tombstoned instances.
   - If the capability_gap is MISSING after a sweep demonstrably ran on a site with ported inline
     handlers, that is the picker-not-running shape: check `auditContractRules` fired (no log
     marker exists — judge by the gap row) and read the sweep's findings for `contract_rules_16_17`.

   > **CORRECTED 2026-08-25 ~20:20Z, next session of this seat: "the sweeps are scheduled" is
   > FALSE for the sweep these controls need.** `tool_health` and `tool_acceptance` are carried
   > ONLY by `design-discovery-agent`, whose rotation task `site-discovery-rotation-design` has
   > been `enabled=false` since 2026-08-11 12:43Z (the slow-ramp pause; migration 395 re-enabled
   > quality only, and design's re-enable never happened). The quality sweep that fired on
   > webdesign 19:47Z post-roll runs nine checks, none of them these. **The controls are blocked
   > on an owner decision — re-enable the design rotation (395's foot UPDATE) or hand-fire one
   > design-discovery run on webdesign — not on waiting.** The watchdog line "rotation tasks
   > enabled: 3/3" is blind to this (counts the availability task): `bugs_open/401`, LANDMINES
   > 2026-08-25, WRONG_CALLS 2026-08-25c, NOTES ~20:20Z entry. Caught by running the control and
   > asking which agent had produced the sweep.
   >
   > **RESOLVED 2026-08-26: step 1 is DONE — both controls PASSED.** Owner ruled and the rotation
   > was re-enabled 09:20Z (proven firing end-to-end); independently, the owner's improvement-loop
   > restart (08-25 ~21:18Z) had already carried a design sweep to webdesign at 03:46Z, and that
   > sweep satisfied both controls: ONE `capability_gap:tool_health_contract_rules` row, residue
   > 20 = the grind's exact remainder; 66 acceptance findings over 69 deployed slots, 0 of 46
   > tombstones enumerated. Evidence: NOTES 2026-08-26 ~10:00Z. Nothing further owed on step 1.
2. **Watch for the grind's Phase B ping** — the standing arrangement (NOTES 2026-08-25 ~12:00Z):
   when their FIRST rich app builds (mind-map/meme/logic-architect/micro-CMS/pasteboard), this seat
   second-eyes the feature-list browser grade. They ping with the app and feature list.
3. **Nothing else is this seat's**: the ~30-tool crosslink backfill is unowned pending the owner;
   the owned-page mention design question is the 333/353 lanes'; the grind is the grind's.

## Decisions that belong to the OWNER (unchanged in substance, restated current)

1. **The five Phase B rich-app reviews** — your standing ruling: one at a time, seen at the served
   page by you. The grind will reach them within days at current pace. The only scheduled human step.
2. **The ~30-tool cross-mention backfill** (tools built before the related_pages fix; PLUS the
   parked-by-design owned-page mentions) — unowned by agreement; the 333 lane's register carries the
   measured shape (every topically-correct pick parks; writable pages are the shopfront only). The
   real decision underneath: SHOULD owned pages ever receive a woven-in mention (via the proven
   section_edit route), or is the shopfront-only posture intended? Until ruled, mentions on this
   site effectively do not happen.
3. **The "tool asserts something untrue about itself" class** (10+ sightings) — still human-caught
   per brief; decide post-grind whether it becomes a Tier-2 criterion, with the full corpus in hand.

## Traps for this seat (the lane-wide ones live in the grind's handoff + LANDMINES)

- **Ancestry-of-stamp is the ONLY load-bearing deploy proof.** Three probe traps now recorded:
  literal-in-binary-anyway (file paths), exit 137 ≈ exit 1 (and the ABSENT control is structurally
  the slow one), and NEW 08-25: **the linker dead-code-eliminates unreachable functions, so a
  literal can be absent from a binary that carries its commit** (LANDMINES, the 375 lane's proof).
- **A claimed CHANGE requires one instrument run twice** — two sessions' counts differing by one is
  a predicate diff until the same query STRING produced both (the hyphen incident, WRONG_CALLS
  2026-08-25 ×2: mine and 333's).
- **All-history counts on `site_work_items` must UNION `site_work_items_archive`** (rolling window).
- **`cd "$dir" || exit 1` in every compound scratch command**; a day-old scratch path is a prophecy
  (WRONG_CALLS 2026-08-24). Use `scripts/verify-head-builds.sh`, never hand-rolled archives.
- **Bug files MOVE when fixed-and-live** — resolve `bugs_open/` vs `bugs_closed/` before appending;
  and bug NUMBERS collide (362 names two unrelated bugs) — resolve by slug.
