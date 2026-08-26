# HANDOFF — webdesign tool rebuilds, THE PLATFORM SEAT. Written 2026-08-26 ~17:15Z.
# UPDATED 2026-08-26 22:20Z — the v1.0.1345 roll landed and step 1 (pod-verify the tombstone
# constant) is DISCHARGED with evidence; the to-do list below is renumbered accordingly.
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
- **The tombstone predicate is ONE constant — and as of 22:13Z it is LIVE IN PRODUCTION**
  (v1.0.1345 roll, verified per service at the artefact; step 0 below has the evidence) —
  `datahelpers.NotRemovedSQL` / `NotRemoved(alias)`,
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

0. ~~**Pod-verify the tombstone-constant work at the NEXT ROLL**~~ — **DONE 2026-08-26 22:13Z**
   (NOTES 22:15Z entry has the full evidence). The roll was `v1.0.1345`, pods restarted
   20:23–20:25Z. Both services stamp `b34c24f4c` (core-manager: own startup line; agent-chassis:
   binary probe with negative control — its startup line rotates out of the log in **~1h50m**, far
   faster than the 08-11 "hours" figure). All four commits (`18561ff05`, `d36faaeaf`, `18853ade6`,
   `5f35e066a`) are ancestors of that stamp → **`datahelpers.NotRemovedSQL` is LIVE in
   agent-chassis AND core-manager.** Behavioural control passed: deployed 2340 / removed 55 /
   pending 40 / approved 23 / **NULL 0** as of 2026-08-26 22:13Z (baseline 2256/49/31/14/0 that
   morning — every population grew, NULL stayed empty). Any future NULL rows must appear in ALL
   populations; today there are none.
1. **Watch the standing demand control**: webdesign's `capability_gap:tool_health_contract_rules`
   residue should SHRINK from 20 as the grind rebuilds. **Status at 22:13Z: still ONE open row
   (`0aba0ca8`, `deferred`, "20 finding(s)") but `updated_at` is unmoved at 03:46:26Z — the checker
   has NOT re-swept, so "still 20" is a lagging meter, not a regression.** The grind hit 49/63 live
   at 19:47Z, so the true remainder is **14**; the next sweep (updated_at moves, or a new row
   closes/replaces this one) should read ~14. Only a FRESH sweep still saying 20, growth, or a
   SECOND open row for the site means checker regression — then read `check_tool_health.go` rules
   16/17. Working query (residue count is in `summary` TEXT — "N finding(s)…" — NOT a jsonb array;
   the first attempt at `result->'residue'` errors):
   `SELECT swi.id, swi.status, swi.updated_at, left(swi.summary,80) FROM site_work_items swi JOIN
   sites s ON s.id=swi.site_id WHERE s.domain LIKE 'webdesign.co%' AND swi.item_type=
   'capability_gap' AND swi.item_key LIKE '%tool_health%' AND swi.status NOT IN
   ('complete','cancelled','rejected');`
2. **Watch for the grind's Phase B ping** (first rich app built → this seat second-eyes the
   feature-list browser grade). **Nearer than this file said at 17:15Z**: the grind is at 49/63
   with 14 remaining, FIVE of which are the rich apps (owner: one at a time, at the served page).
   Read their `HANDOFF_2026-08-26_continue_here.md` for current queue position before assuming.
3. **Nothing else is this seat's**: the checker fix is staged_component_build's; the watchdog fix
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
  provenance-grep of chassis logs can match LANDMINE TEXT about provenance. Also measured 08-26
  22:12Z: the chassis startup line rotates out of the retained log in **~1h50m** — an empty grep
  there means "out of range", not "unstamped"; the binary probe (expected sha + bogus-sha control
  in the same breath, `$SHA` non-empty guarded) is the fallback with no shelf life.
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
