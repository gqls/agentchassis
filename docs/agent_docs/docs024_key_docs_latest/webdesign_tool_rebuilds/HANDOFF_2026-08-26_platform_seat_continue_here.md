# HANDOFF — webdesign tool rebuilds, THE PLATFORM SEAT. Written 2026-08-26 ~17:15Z.
# Supersedes HANDOFF_2026-08-25_platform_seat_continue_here.md (kept for its correction trail).

**TWO SEATS WORK THIS LANE.** This is the PLATFORM seat (session `webdesign-tool-rebuilds`,
WITH the trailing s). The GRIND seat (`webdesign-tool-rebuild`, no s) owns the rebuild queue and
its own fresh handoff: `HANDOFF_2026-08-26_continue_here.md` (Gate Zero, the 41-hold hazard, #44's
analysis, the two owed re-fixes). Do not file add_tool items from this seat; announce before
touching files the grind writes. Evidence for everything below: `NOTES_…` 2026-08-25→26 entries
(newest at bottom; note the 08-26 morning entries' HEADERS ran up to 3h fast — trust the commit
times, correction block in place).

## State — ALL LANDED, nothing mid-flight

- **Track 2 is DONE and VERIFIED AT THE ARTEFACT** (2026-08-26): both demand controls PASSED on
  webdesign.co.uk's 03:46Z sweep — ONE `capability_gap:tool_health_contract_rules` row, residue
  20 = the grind's exact remainder; 66 acceptance findings over 69 deployed slots, **0 of 46
  tombstones enumerated**. Council trail 21540c8e complete (REVISE→APPROVED→advisories acted).
- **Design rotation RE-ENABLED** (owner ruling, 2026-08-26 09:20:04Z, 10800s) and proven firing
  end-to-end in 40s. The improvement loop (owner-restarted 08-25 ~21:18Z) is the SECOND carrier —
  it dispatches design/completeness/acceptance discovery as children on its own site selection and
  writes NO rotation stamps. Fourteen threads were notified; `bugs_open/401` (the watchdog's 3/3
  miscount) stays OPEN — fix unowned, CONTRIB in `bugfix_230_discovery_driver/`.
- **The tombstone predicate is ONE constant** — `datahelpers.NotRemovedSQL` / `NotRemoved(alias)`,
  NULL-safe, adopted at ALL **15** spellings across **13** files (census as of 2026-08-26; it was
  "5" then "11" then "15", each revision from a wider instrument). Council trail **89dcc04a**:
  round-1 REVISE (reuse: shared const, not lockstep tests) → done as asked → **APPROVED** (3
  advisories, all dispositioned with measurements — NOTES ~13:20Z entry). Commits `18561ff05`,
  `d36faaeaf`, `18853ade6`, `5f35e066a`; guards = value test + comment-skipping negative scan over
  platform/orchestration AND internal/, both mutation-proven (beware: a sed mutation gofmt
  disagrees with passes VACUOUSLY — grep the file before trusting a mutation's "ok").
- **The anchor-class acceptance defect is CONFIRMED and CONTAINED, not fixed.** Diagnosis
  `91228c39` (corr `2b64e510`): acceptance criteria are the tool's own PLAN document
  (`doc_plans`, via `loadCurrentCriteria` — NOT `doc_context.criteria_json`, that's the step
  config's field name) naming bare ids; the renderer instance-prefixes every id
  (`ConvertTemplateToInstanceScope`, `InstanceToken = "c-"+function`); bare selectors can never
  match. 110/112 of the check's historical failures are this class; **32 completed regenerations
  fleet-wide are suspect**. Fix + cleanup + criteria-vs-renderer decision belong to
  **staged_component_build** (ACTIVE; evidence in their CONTRIB trail). Holds in place: noted's
  editor items, webdesign.uk's flagship pair, and the grind's **41 on webdesign.co.uk** (16:45Z:
  deferred, handler_agent KEPT, release condition IN-ROW, 15 other-site rows untouched). **The
  hold shape is the lane's reference — NOTES ~16:55Z. Never drain a deferred anchor row without
  reading its in-row condition; a hold that clears the handler is a deletion wearing a status.**

## What the NEXT session of this seat should do, in order

1. **Pod-verify the tombstone-constant work at the NEXT ROLL** (it is Go-source SQL, inert until
   images ship — affects agent-chassis AND core-manager): ancestry of `18853ade6` and `5f35e066a`
   in each service's own stamp (`git merge-base --is-ancestor <commit> <stamp>`), then the
   behavioural control: `SELECT COALESCE(build_status,'<NULL>'), count(*) FROM page_components
   GROUP BY 1;` (was deployed 2256 / removed 49 / pending 31 / approved 14 / NULL 0 on
   2026-08-26 — any future NULL rows must now appear in ALL populations, not just the assembler's).
2. **Watch the standing demand control**: webdesign's `capability_gap:tool_health_contract_rules`
   residue should SHRINK from 20 as the grind resumes rebuilding. If it grows or a SECOND open row
   appears for the site, that is a checker regression — read `check_tool_health.go` rules 16/17.
3. **Watch for the grind's Phase B ping** (first rich app built → this seat second-eyes the
   feature-list browser grade). Currently far out: their queue sits behind ~107 sweep items.
4. **Nothing else is this seat's**: the checker fix is staged_component_build's; the watchdog fix
   is the 230 lane's; the noted lane may bring an external-source-of-truth RFC (owner has directed
   fleet-wide — this seat contributes the checker-side evidence section if asked, builds NO
   tool_health-local waiver).

## Decisions that belong to the OWNER (restated 2026-08-26, asked-and-answered removed)

1. **The five Phase B rich-app reviews** — standing ruling: one at a time, at the served page.
   Not yet reachable (grind blocked); becomes live when they get there.
2. **The ~30-tool cross-mention backfill + the owned-page mention posture** — unchanged, unowned;
   until ruled, mentions on this site effectively do not happen.
3. **The "tool asserts something untrue about itself" class** → Tier-2 criterion or not, decided
   post-grind with the corpus.
- *Optional overrides, defaults otherwise fine:* ratify-or-reverse the 41-hold (one UPDATE
  reverses); let rebuild work jump the sweep backlog or let it drain (default: drain).

## Traps for this seat (lane-wide ones live in the grind's handoff + LANDMINES)

- **Ancestry-of-stamp is the ONLY load-bearing deploy proof; per SERVICE.** Probe traps recorded:
  literal-in-binary-anyway, exit-137≈1, linker dead-code-elimination, and NEW 08-26: a
  provenance-grep of chassis logs can match LANDMINE TEXT about provenance.
- **Check names and item types are DISJOINT vocabularies** for the tool family (LANDMINES
  2026-08-26): census by `spec->>'check'`, never `item_type='tool_health'` (0 rows for ever).
- **A config-gated check's first run is bounded below by its config's apply time** — zeros before
  that are vacuous (WRONG_CALLS 2026-08-26, told to a peer as coverage).
- **All-history `site_work_items` counts must UNION `site_work_items_archive`**; a claimed CHANGE
  needs one query string run twice; **read `now()` before writing a timestamp** — this seat's
  08-26 morning headers ran 3h fast by estimating elapsed time.
- **A lane's shipping mechanism is read from its FILINGS, not from what exists in its directory**
  (WRONG_CALLS 2026-08-26b — the replace_existing over-generalisation).
- Bug numbers collide (401 is uniquely ours today, but 410 now names two — CLAUDE.md list);
  resolve by slug, `git log` the file path.
