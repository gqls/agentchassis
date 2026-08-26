# HANDOFF 2026-08-26 — dispatch throughput — CONTINUE HERE (supersedes HANDOFF_2026-08-25b)

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` 2026-08-26 entries
(ruling B applied; ~2h + afternoon reads; VERIFY 6/7 zombie-tail narrowing; bugs_open/413 filed;
658 prepared + APPROVED) → `SUMMARY_2026-08-26_dispatch_throughput.md` (the lane's first summary
— the whole arc, written to be read aloud) → `README_where_we_are.md` (owner's plain-prose log).
The superseded `HANDOFF_2026-08-25b_continue_here.md` still holds the fuller 584/N=2 history.
Paths relative to `docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## The one-paragraph state (evening 2026-08-26)

**Ruling B is live and working.** Migration 637 (08:51Z): sibling retired (row kept as rollback),
`build-pipeline-trigger` at `interval_seconds=30` (~60s fires). Day-one measurements: cadence
p50 60s; 22 distinct sites/2h (pair co-picked one site 94%); lost claims 11.4% vs 58–60% under
the pair (post-outage residue 10.7%); afternoon claims 265–278/h against the ~300/h ceiling;
demand deep (1,200+ triaged, ~30 sites). **Phase 3 (batch 5→8) is PREPARED and council-APPROVED
but NOT applied** — `sql_for_agents/658_dispatch_phase3_batch_8_HOLD.sql` (+`_ROLLBACK`), gated
on tomorrow's 24h read. **bugs_open/413** (selector ranks sites by oldest-row AGE, loader serves
by PRIORITY → a pinned worst-priority old row starves younger sites positionally, invisible to
aggregates) is filed by this lane; the FIX is owned by the "bugs_open/413" session (their
migration 657), coordinated below. A chassis roll at ~20:26Z caused **no dispatch gap**
(measured, NOTES ~20:5x).

## TOMORROW 2026-08-27 — the ordered protocol (owner-sequenced)

1. **09:00Z SHARP — the 24h post-B read.** All queries in `RUNBOOK` §"Concurrency meters" +
   §"Per-site starvation floor". Window: post-B ≥ 2026-08-26 08:55Z. Metrics: fire cadence
   p50/p90/max + sibling frozen; lost-claim share + anatomy; loops + distinct sites/hour;
   claims/h + completions/h with arrivals AND open depth beside; wait-to-claim p50/p90;
   **per-site floor + pinned/victim census (REQUIRED — quote the WORST site)**; daily 584
   VERIFY (7 assertions; 3–20 min under load, never a 2-min timeout; zombie-tail NOTICEs are
   benign). Caveats to hold beside: the 08-25 23:47→08:55 credit outage (pre-B side), the
   ~20:26Z chassis roll (inside the window, no gap measured), `orchestration_states` retention
   ~24–27h (a late read loses the morning — say plainly what was lost, never narrow silently).
   Record: NOTES (figures+windows) → README (plain prose) → NOT a new SUMMARY unless surprised.
2. **GATE:** cadence p50 ≤ ~65s · lost claims well below 58–60% · 0 true double-handles ·
   VERIFY green. **FAIL → hold 658, tell the 657 session immediately, investigate.**
3. **PASS → ~09:30Z hand-apply Phase 3:**
   `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/658_dispatch_phase3_batch_8_HOLD.sql`
   Re-run the one-row census first (RUNBOOK/NOTES ~20:5x — the guardian advisory check).
   Verify at the artefact (RUNBOOK §"Phase 3 apply"): both knobs read jsonb number 8; post-apply
   loops show loaded up to 8 and claim_result keys past _4. **Honest expectation: ~+7%**
   (overhead amortisation) — cap binds (80.3% of turns, 22/29 sites ≥8) but turns lengthen
   ~60%; do NOT quote ~1.5×. Watch collected_data sizes (~×1.6 → ~7.3MB max tail vs 8MiB warn).
4. **~11:30Z — 2h Phase-3 read** (same meters + floor + collected_data watch).
5. **≥12:00Z — the 657 selector fix applies** (their session pings the stamped time; cut all
   windows on it). Their selector reads `max_items` LIVE — no lockstep on our side.
6. Next day: 24h Phase-3 window with the floor; hand the 657 session their post-fix floor
   numbers (our 09:00Z read is their pre-fix baseline).

## Council state (all closed)

- **658 APPROVED round 1**, corr `95099f95-b32b-4656-bfe1-28f140de3717` ("1 advisory, none
  high-severity"). All three advisory checks DONE ~20:5xZ with evidence in NOTES: wrong-path
  concern covered mechanically by the refusal-first NULL check; build-dispatch-loop has exactly
  ONE row in any state (the four duplicate-active types are content-creator,
  content-creator-contact, chief-strategist, site-component-architect); knob census — only
  site-work-orchestrator also uses `load_work_items`, deliberately untouched. Commit
  `2c419c32d` carries `Council-Submitted:` (auto-credits); the NOTES/handoff commit carries
  `Council-Reviewed:`.
- 637 APPROVED corr `69a04e0a` (resubmission after the credit-outage-killed round); 582/584
  APPROVED r3 corr `db9b7cbf`. Nothing owed on any.

## Coordination

- **"bugs_open/413" session owns the 413 FIX** (owner-routed; who-owns names this lane only
  because we FILED it — human routing supersedes). Their 657: selector ranks sites by
  min(created_at) over each site's top-K by the LOADER's ordering, K read LIVE from
  `load_items.max_items` (COALESCE 1); candidate 2 (age floor) goes to the owner as policy.
  Contract: they apply ≥12:00Z 2026-08-27 **on this lane's all-clear after the ~11:30Z read**,
  and ping with the stamped time. Their build is COMMITTED (`5f1401a85`: 657_HOLD/_ROLLBACK/
  _VERIFY + an AST ordering-contract test, mutation-proven) with council in flight
  (corr `ecf2e542` — before giving the all-clear, check their verdict landed and any REVISE was
  acted on). Their dry-run evidence, dated 20:46Z: rolled-back-transaction apply green; census
  28 eligible/15 pinned; same-instant divergence OLD→webdesign.co.uk (the rank-135 pin),
  NEW→vetcomparison.uk. The fire-gate gap is filed as **bugs_open/415** (gate:
  `status='triaged' AND pipeline='build'` vs selector: triaged+approved, no pipeline filter —
  an approved-only backlog never fires). Note: their "413 FIX BUILT" NOTES section rode commit
  `32b4c89c7` as a declared same-file passenger — expected, not an anomaly.
- **391 lane** (bugfix_389_cta_relevance) handed over the 413 symptom, PARKED, nothing owed;
  their CONTRIB file sits in this dir (priority is inert BETWEEN sites — with corrections).

## Traps for this lane (new first; 25b's all still stand)

- **658's rollback restores 5 EXPLICITLY — never `#-` the keys**: Go defaults are 50/20
  (`load_work_item_actions.go:699`, `loop_actions.go:52-55`); deletion = batch 50/20, 10×.
  Same reason the knob values must be bare jsonb NUMBERS (a quoted "8" silently runs 50/20).
- **Do NOT re-run `sql_for_agents/051_build_dispatch_loop.sql`** — it replaces the entire loop
  workflow with the ancient one-item config, not merely a knob.
- **already-applied refusal BEFORE snapshot_agent** in any agent_definitions migration
  (LANDMINES 2026-08-26 replay-decoy; 526/541/633 have it backwards; 658/its rollback are the
  corrected pattern). Verify by value read-back, never `updated_at` (degenerate) and never by
  joining `snapshot_agent()`'s return (it returns the SOURCE row id).
- **A stale-reaped orchestration's `updated_at` is the REAP stamp** — raw [created_at,updated_at]
  overlap counts zombie tails; VERIFY 6/7 excludes-and-NOTICEs that shape (LANDMINES 2026-08-26).
- **The aggregate meters cannot see 413's starvation** — the damage is an absence; the per-site
  floor is the meter. A pin census is a dated SNAPSHOT (pins clear dynamically).
- Any by-name UPDATE on `scheduled_tasks` 'build-pipeline-trigger' misses the disabled sibling
  (the rollback path must not desync — VERIFY 1/7 spans disabled rows). `213_...sql` sits
  untracked with exactly that shape; leave it.
- `orchestration_states` retains ~24–27h — baselines the moment you need them.
- Full VERIFY runtime is DB-load-bound: 2m49s quiet, 10–20+ min under concurrent jsonb scans.

## OWED after tomorrow (the standing queue, unchanged order)

D4 LLM spend governor (first BUILD item; gate for interval ≤25 = option C; the 08-25/26 credit
outage is its measured case — confirm at-cap shedding policy with the owner first) · deploy
batching (D8 interim) · clients-first lane (D2) · Batch API (D6) · D16 retention (the ~24–27h
window keeps destroying pre/post comparisons — design consideration, not just cost) · per-class
maintenance LLM cost (RESEARCH §6) · DNS plan B (with the domain-programme lane; check pickup).

## Session hygiene

Pathspec commits; grep LANDMINES for symbols you touch; who-owns before routing at any bug;
re-read CLAUDE.md from disk before acting on multi-session rules; the workstreams memory points
here (COLD-START = this file).
