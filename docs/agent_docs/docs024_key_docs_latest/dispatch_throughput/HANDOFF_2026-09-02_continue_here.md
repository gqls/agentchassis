# HANDOFF 2026-09-02 — dispatch throughput / D4 spend governor — CONTINUE HERE (supersedes HANDOFF_2026-08-30)

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` 2026-08-31 + 2026-09-02
entries (the D4 arc: shedding ruling → stage A three council rounds → stage B three rounds →
budget set; plus the 384/426/314 cross-lane threads) → `PLAN_2026-08-18_dispatch_throughput.md`
§"D4 — LLM SPEND GOVERNOR" + §"stage B DESIGN CORRECTION" → `SUMMARY_2026-09-02` (the arc read
aloud) → `README_where_we_are.md` (owner's log — the last three entries are the current
conversation). Paths relative to `docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## The one-paragraph state (2026-09-02 ~21:15Z)

**D4 is BUILT, APPROVED, BUDGETED — and one blocked step from armed.** Stage A (migrations
671/672/673: meter, class map, config, state, 120s task, `governor_admits()` canonical
predicate, `governor_withheld_now` view) is live, inert, council-APPROVED (corr `80df0963`).
Stage B (Go: renderer + loader flag + claim backstop, commits `dec5ad61b`→`c0a18f37d`→
`6a84e3dc1`; config: `674_..._HOLD.sql`) is council-APPROVED (corr `8f4bb57d`, r3, all
advisories built in). **The owner set the budget: $2,000/month (18:26Z, verbatim in NOTES)**
— thresholds L1 $1,400 / L2 $1,700 / L3 $1,900; September MTD ~$244 at set time; the state
task recomputes against it every 120s (proven at its own tick 18:31Z). **UPDATE ~21:45Z, same evening: the token was refreshed and NEXT steps 0–5 ALL EXECUTED
CLEAN** — both replicas capability-probed (present+absent controls), **674 applied 21:27:14Z**
(new selector md5 `fcbe8821`), the 657-VERIFY lockstep done and green in the same sitting,
the 10-minute canary clean (11 fires, 0 governor refusals, 0 withheld while disabled), the
suffix dropped and the ledger updated. **ALL that remains of D4 is the owner's two acts:
raise the console cap above \$2,000, then `UPDATE governor_config SET enabled=true WHERE
id=1;`** — then observe the first shed cycle and unlock option C (NEXT-7). ⚠ small find at
the rename: number 674 is SHARED with another lane's `674_farmer_cull_...` — resolve by slug.
The NEXT list below is kept for the record of HOW it was done.

## NEXT (the exact sequence; steps 2–5 are written in 674's own header too)

0. **Kubeconfig refresh** — owner downloads from the Rackspace Spot console (the
   [[kubeconfig-token-expires-every-3-days]] memory; decode-check command inside it).
1. **Stamp check, BOTH replicas** (a roll is not evidence the code shipped):
   `kubectl -n ai-persona-system logs <pod> --tail=400 | grep -m1 'build provenance'` per pod,
   then `git merge-base --is-ancestor dec5ad61b <stamp>` — BOTH must pass. Startup line
   scrolls: absent = "not in range", fall back to the known-sha binary probe WITH both
   controls (memory: prove-a-deploy-at-the-artefact).
2. **Apply 674**: `kubectl ... psql -v ON_ERROR_STOP=1 -f - <
   docs/agent_docs/sql_for_agents/674_d4_stage_b_enable_selector_and_flags_HOLD.sql`.
   Refusal-first (md5 `d29807313` exact + anchor-count + function-exists + one-row censuses +
   governor-still-disabled); snapshots; selector gains `AND governor_admits(wi.item_type)`;
   two step flags (bare jsonb true). Chained apply→rollback dry-run was green 2026-09-02.
3. **⚠ 657-VERIFY LOCKSTEP (r3 guardian — do it in the same sitting):** the apply CHANGES the
   selector md5 that `657_selector_ranks_sites_by_loadable_work_VERIFY.sql` pins. Add the new
   md5 (read it back post-apply) to that file's accepted list, run it green, commit both files
   together — or the daily contract check reads as selector drift.
4. **Canary (r3 editquality): watch the first ~10 fires** — loops load, claims proceed,
   ZERO `spend_governor_shed` refusals (governor disabled = byte-identical behaviour).
   Meters: RUNBOOK §Concurrency + the floor with its THREE controls.
5. **Drop 674's suffix + `--record-only` in one motion** (bugs_closed/150; the runner refuses
   to record a sidecar, so rename first; both paths on the commit; ls-tree at HEAD).
6. **The owner's "enable"** (`UPDATE governor_config SET enabled=true`) — THE deliberate act.
   Before it, remind them: **console cap must sit ABOVE $2,000** (~$2,500) or the account
   wall still fires first. At enable, September MTD vs $1,400/1,700/1,900 decides the level
   immediately — at ~$99/day measured burn, L1 lands ~day 14 (this is a THROTTLE by owner
   choice, stated to them plainly 09-02 and accepted; README carries the plain version).
7. **After one real or induced shed is observed: option C unlocks** (trigger interval ≤25s)
   — its own migration editing VERIFY 2/7's lever in lockstep, its own council round.

## D4 reference card

- Tables/objects (all migration-created, ledger-recorded): `governor_model_prices` (verified
  rates; cache writes at the 1h rate — err high), `governor_work_class_map` (20 types; class
  + llm_bearing; unmapped ⇒ maintenance+bearing at enforcement), `governor_config` (id=1;
  enabled=false; budget 2000; 70/85/95), `governor_state` (computed_at IS the heartbeat —
  stale means task-not-running, never "fine"), `governor_spend_mtd` (the meter),
  `governor_admits(item_type)` (THE predicate — never re-spell it; Go tests enforce),
  `governor_withheld_now` (the shed-event view: withheld vs stuck in one query),
  scheduled task `spend-governor-state` (120s; advisory lock + FOR UPDATE old-read;
  doc_notes row on level CHANGE only).
- Register **AGOV-013**; travelling design `doc_plans('pipeline','spend-governor')` + decision
  doc_notes `1032c8f4`; owner rulings verbatim in NOTES (shed order 08-31; budget 09-02).
- **STANDING GATE (r3 architecture):** a SECOND consumer opting into `honour_spend_governor`
  is architecture-scope — its own review round, never routine config. In 674's header + AGOV-013.
- Rollbacks: 674 (restores md5 `d29807313`, DELETES flags — correct for opt-in bools, unlike
  658's knobs), 675 (refuses while any config references the function), 671 (refuses while
  enabled). All guarded, all dry-run-proven.

## Traps (new since 08-30; the 08-30 + 08-26 handoffs' all still stand)

- **Council sketches: paste guard/verify text VERBATIM** — three r-rounds in one corr drew
  objections to checks that existed but were summarised. A summary of a guard reads as its
  absence.
- **A 100%/0% split against one predicate = you measured a DEFINITION** (memory:
  a-uniform-split-means-you-measured-a-definition) — fetch the predicate; a colleague's
  summary (or a pre_query's FIRST SCREEN) is a hypothesis about the query. Cost two lanes
  four errors on 09-02.
- **The `detected` layer with empty handler_agent is DESIGNED-permanent flags** (promoter's
  own text) — never meter it as parked work; the RUNBOOK §"Promotion-layer meters" carries the
  double-corrected meter (handler-BEARING detected age + the promoter's held_detail).
- **`unresolved` is TERMINAL and can be minted at BIRTH** (384's find) — the scale-baselines
  `work_items_open` figure overcounts; the floor's eligible-count is the actionable figure.
- **Hand-apply is finished only when `--record-only` has run** (the 09-02 seven-unrecorded
  find; 426's daily drift check now watches fleet-wide). _HOLD lifecycle = apply → drop
  suffix → record (bugs_closed/150).
- Timer/wait discipline: long background sleeps get reaped — persistent Monitor for >1h
  waits; `date -d` parses LOCAL time.
- Pre-roll turn-time baselines contain 408's 4h wedges (their fix rides the same roll as
  stage B's Go); post-roll tail changes are NOT the governor's (it is inert).

## Coordination state

- **All cross-lane threads CLOSED**: 413 (bug closed on our verdicts), 414, 314/426 (their
  daily ledger-drift check live; candidate 2 = probe vocabulary stays THEIR open half),
  384 (promoter resolved as designed; their two-strike chain sits with 389/bugfix_308 —
  FYI-only: their self-resume prediction window is ~2026-09-03 21:30Z; if leopardess's
  rerender completions jump unaided, their mechanism is confirmed — nothing owed from us).
- 415's fix (migration 688, wider fire gate) may apply any time — no lockstep owed (VERIFY
  1/7 pins parity, not text); hold its apply time beside cadence reads (more cheap no-op
  fires expected).
- Ordering decisions ROUTE TO THIS LANE (owner 09-02); the owner's PROVISIONAL no-reorder
  ruling (verbatim in NOTES) declined candidates 2+3; mechanical revisit trigger = the pin
  census age tail.

## The standing queue after D4 (unchanged order)

Option C (post-enable, see NEXT-7) · deploy batching (D8 interim) · clients-first lane (D2 —
untouched by the no-reorder ruling) · Batch API (D6) · D16 retention (destroyed the 08-28
read) · per-class maintenance LLM cost (RESEARCH §6) · DNS plan B (domain-programme lane).
Daily habit: 584 VERIFY (zombie NOTICEs benign, TWO reaper spellings; runtime is DB-load-bound).

## Session hygiene

Pathspec commits; grep LANDMINES for symbols you touch; who-owns before routing at any bug;
re-read CLAUDE.md from disk before acting on multi-session rules; the workstreams memory
points here (COLD-START = this file).
