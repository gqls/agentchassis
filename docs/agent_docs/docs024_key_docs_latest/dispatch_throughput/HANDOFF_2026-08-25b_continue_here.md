# HANDOFF 2026-08-25b — dispatch throughput — CONTINUE HERE (supersedes HANDOFF_2026-08-25)

> **UPDATE 2026-08-26 — OWNER RULED B and it is APPLIED (migration 637, 08:51Z):** sibling
> DISABLED (row kept for rollback), `build-pipeline-trigger` at `interval_seconds=30` (~60 s
> spaced fires). C (interval 25) is gated on the D4 governor — **mechanically**: the 584 VERIFY
> (now 7 assertions) RAISEs on a second enabled row or interval < 30. Council for 637:
> **APPROVED 2026-08-26 ~10:4x** on corr `69a04e0a` (2 advisories = the ROLLBACK file's absence
> from the edits array; it exists, shipped in `a5fd1651e` — disposition in NOTES; nothing owed).
> ⚠ The FIRST 637 round died `complete_invalid` on an **account-credit outage, 2026-08-25
> 23:47Z → ~08:55Z** (~9 h, every LLM call 400'd "credit balance too low"; recovered ~09:00Z) —
> the 24h post-B read must hold that window beside it (LLM-bearing handlers dead, mostly the
> pre-B side), and D4's case is now measured, not hypothetical.
> OWED 1 below is therefore DONE; the new queue head is the **24h post-B read** (cadence p50
> ~60 s; lost-claim share from ~59% → single digits; distinct sites per hour; wait p50/p90 with
> arrivals held beside) — meters in RUNBOOK; ~30-min AND ~2h reads in NOTES 2026-08-26 (cadence p50 60 s; 22 distinct sites/2h; lost claims 11.4%, post-outage 10.7%; batch cap BINDS at avg 4.66/5). After that:
> Phase 3 batch-only (5→8 both knobs), then D4. Full ruling + evidence: NOTES 2026-08-26.
> **ADDED ~15:4xZ: the 24h read MUST include the per-site starvation floor (RUNBOOK §"Per-site
> starvation floor") — `bugs_open/413`, found via the 391 lane's handover: the selector ranks
> sites by oldest-row AGE while the loader serves by PRIORITY, so one old worst-priority row
> pins its site and starves every younger site of trigger dispatch (finetuning.uk: 73 eligible,
> 10+ h, zero loops), invisibly to every aggregate meter. 090 run `250188a7` in flight at cut
> time; verdict + Phase 3 interaction notes in the bug file.**

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` 2026-08-25 session-2
entry (§1–§7 — the day the lane's central claim fell) → `README_where_we_are.md` 2026-08-25
(the owner's plain-prose version + options A–D). The PLAN/STARTER/RESEARCH all carry
**CORRECTED 2026-08-25 blocks — trust those over the original text.** `RUNBOOK` §"Concurrency
meters that actually measure concurrency" has every query this handoff cites. Paths relative to
`docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## The one-paragraph state

**The scheduler is fire-and-forget and always was (since `892a289e9` 2026-03-17):** `runTick` →
`fireTrigger` → `stampCompleted` sets BOTH row stamps at fire, so a `scheduled_tasks` row was
never a single-flight slot, `max_concurrent`/`timeout_seconds`/`countInFlight`/the per-row guard
are inert for every `fire_message` row (40 enabled fleet-wide), and the agents' `notify_scheduler`
stamps are inert — 582/583 were hygiene, not a fix. What 584 (N=2, live since 08-24) actually does:
a second fire ~1 s after the first every ~90 s, which **co-picks the same site 94%** of the time
(the selector can't see a claim that lands p50 17.7 s later) → **~+10–15% claims/h, not 2×**, two
handlers on the deep site, 39% of claim attempts lost as cheap bounced claims. **Safety held:
0 double-handles in 2,579 handlers / 2,502 items; handler fail rate LOWER with a same-site partner
(1.55% vs 3.85%).** Full evidence: NOTES 2026-08-25 §5; `bugs_open/398`; LANDMINES 2026-08-25 ×2;
`WRONG_CALLS.md` 2026-08-25. Rollback unchanged: `UPDATE scheduled_tasks SET enabled=false WHERE
name='build-pipeline-trigger-2';`

## COUNCIL — round 3 APPROVED 2026-08-25 ~19:50Z (nothing owed here)

- Correlation **`db9b7cbf-7b94-471a-a4cf-26a6679fa47f`** (rounds 1–3 accumulate on it).
  R1 REVISE (guardian: 4th-stamp fear → answered by whole-text census). R2 REVISE (guardian:
  induction owed → answered on the natural population; editquality: 2 turns ≠ 2 sites → answered,
  and measuring it is what refuted the mechanism). **R3 submitted 2026-08-25 ~16:4xZ with the
  corrected mechanism stated plainly** + the VERIFY file as edit 5 (`operation: add` — 'create'
  is refused client-side). Submission JSON: scratchpad `council_582_584_r3.json` (regenerable
  from NOTES §5 + the r2 JSON).
- Check: `SELECT created_at, left(body,400) FROM doc_notes WHERE categories ? 'council-gate' AND
  body LIKE '%db9b7cbf%' ORDER BY created_at DESC LIMIT 1;` — ⚠ the header says "(round 1)" on
  EVERY round; count reports, don't read the label. Full report: `diagnosis_artifacts`
  kind='council_report', same correlation, latest.
- **R3 verdict: APPROVED — "4 advisory objection(s), none high-severity", 7 abstained.** The two
  actionable advisories were acted on the same evening (VERIFY liveness re-keyed off the
  unreliable `owner_agent_type` column; assertion 6 added — a THIRD sibling row RAISEs;
  mutation-proved). `dc76d1c30`/`e80561f04` carry `Council-Submitted:` (auto-credited); the
  advisory-fix commit carries `Council-Reviewed: db9b7cbf-…` (verdict read in full). Nothing owed.

## OWED — the ordered queue

1. **OWNER DECISION (blocks the lever's final shape): options A–D in README 2026-08-25.**
   A = keep sibling (+10–15%, safe). B = retire sibling, `interval_seconds` 60→30 on the original
   (60 s cadence, ~1.5×, spaced fires → distinct sites) — **recommended**. C = interval 25 (30 s
   cadence, ~3×) — hold for the D4 governor per the owner's own caution. D = keep sibling + teach
   `find_dispatchable_site` to skip sites with a live dispatch turn (restores per-site
   serialisation; `idx_orch_site_active` exists for exactly that shape; agent-config change =
   council scope). **Do not flip anything without the ruling** — N=2 was owner-authorised by name.
   **DEADLINE 2026-09-01 (set per council r3 architecture advisory): if unruled by then,
   re-present A–D to the owner — provisional-with-no-deadline becomes permanent by default.**
2. **Until that ruling: run the VERIFY daily** —
   `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
   -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql`
   (7 assertions since 637's lockstep edit; 6/7 narrowed 2026-08-26 to exclude-and-NOTICE
   zombie-tail pairs — a stale-reaped handler's updated_at is the reap stamp, so raw overlap
   false-RAISEs on a benign successor re-claim; NOTES 2026-08-26). This is the monitoring commitment made to the guardian.
3. **Phase 3 is now batch-ONLY**: `load_items.max_items` 5→8 + `process_item.max_iterations` 5→8
   together (one without the other is a silent no-op). The "timeout 300→600 lockstep" half is a
   NO-OP on this binary (timeout_seconds is inert — bug 398); record D3 as satisfied-vacuously,
   don't ship a dead knob. Only after the owner's lever ruling — batch interacts with the co-pick.
4. **D4 LLM spend governor** — unchanged, first BUILD item, gate for more concurrency (N>2 or
   interval ≤25). Confirm the at-cap shedding policy with the owner before building.
5. Then the standing queue: deploy batching (D8 interim), clients-first lane (D2), Batch API (D6),
   D16 retention proposal (note: `orchestration_states` retention ~24–27 h destroyed the pre/post
   comparison here — a D16 design consideration, not just a cost one), per-class maintenance LLM
   cost (RESEARCH §6).
6. DNS plan B: with the domain-programme lane (pointer left 08-21); check they picked it up.

## Traps for this lane (new ones first)

- **Any by-name UPDATE on `scheduled_tasks` `build-pipeline-trigger` misses the sibling** —
  LANDMINES 2026-08-25; `213_dispatch_gate_matches_dispatcher.sql` sits UNTRACKED (since 08-12,
  never applied) and has exactly that shape. Agent-side edits reach both rows; row-side don't.
- **The per-minute distinct-sites CLAIMS meter is not a concurrency meter** (pre-change control:
  27.7% of minutes ≥2 sites). Use the orchestration-level meters in RUNBOOK.
- **`orchestration_states` retains ~24–27 h** — take baselines the moment you need them.
- Do NOT add sibling #3 (unchanged), and do NOT re-file the 090 (max_tokens class, `bugs_open/183`).
- The claim census/attribution gaps, `bugs_open/029` ownership, build-cost n=1 — unchanged from
  the superseded handoff; read it if you need them.

## Session hygiene

Pathspec commits; grep LANDMINES for symbols you touch (two new entries are ABOUT this lane);
who-owns before touching any bug; workstreams memory points here (COLD-START).
