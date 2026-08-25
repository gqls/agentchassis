# HANDOFF 2026-08-25 — dispatch throughput / whole-architecture scale review — CONTINUE HERE

> **SUPERSEDED 2026-08-25 (same day, session 2): read `HANDOFF_2026-08-25b_continue_here.md` instead.** The mechanism this file states ("each row is a single-flight slot", "stamp trap dead", "per-site serialisation") was REFUTED at the artefact hours after it was cut — NOTES 2026-08-25 §5. The council/queue state below is also stale (round 2 returned REVISE; round 3 submitted with the correction).

**Read first, in this order:** this file → `NOTES_dispatch_throughput.md` (same dir; the
dated evidence trail) → `RESEARCH_2026-08-18_throughput_to_thousands_of_domains.md`
(the scale review + decision table + owner rulings §10). The original
`PLAN_2026-08-18_dispatch_throughput.md` is the dispatch-specific slice (its §3.2/§3.3
carry CORRECTED blocks — trust those, not the original text). `RUNBOOK_…` has every
query. All paths relative to `docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/`.

## What this lane is

Owner-directed (session "throughput", 2026-08-18→25): research + execute the path to
**several thousand hosted domains** with **promotion bursts up to 50 signups/day**
(a welcome maximum; expect a fraction). The whole-architecture scale review the
site_delivery lane seeded is DONE and lives here (RESEARCH doc). The owner answered the
decision table 2026-08-21 (RESEARCH §10 + NOTES 08-21 entry — read both; the chat text
is mirrored in `README_where_we_are.md`).

## State: what is LIVE and PROVEN

1. **Dispatch N=2 is LIVE since 2026-08-24 ~10:38Z** — migrations
   `docs/agent_docs/sql_for_agents/582/583/584_dispatch_sibling_*.sql`, applied in order,
   artefact-verified: sibling row `build-pipeline-trigger-2` fires and stamps ITS OWN
   `scheduled_tasks` row; two `call_dispatch` turns observed concurrently; zero errored
   steps. **Rollback = `UPDATE scheduled_tasks SET enabled=false WHERE
   name='build-pipeline-trigger-2';`** (instant; config is live immediately).
   Register: WDS-002 (`docs/agent_docs/docs026_concept_register/register/work-dispatch.md`)
   carries the mechanism + the census instruction.
2. **The stamp trap had THREE writers, not the PLAN's two** (trigger `notify_scheduler` +
   `notify_scheduler_idle` + loop `notify_scheduler`). All parameterised
   (`WHERE name = $1`, `params: ["input_data.task_name"]`, task_name mapped through
   `call_dispatch.input_mapping`). Fleet-wide WHOLE-TEXT census 2026-08-25: **0**
   hardcoded stamps anywhere (positive control: census sees substep text — 6 configs
   match `spawn_handler`), 0 in `pre_query`, `$1` form in exactly the 2 edited agents.

## COUNCIL — round 2 PENDING, first thing to check

- Correlation **`db9b7cbf-7b94-471a-a4cf-26a6679fa47f`** (both rounds accumulate on it).
- Round 1: **REVISE**, gating = guardian HIGH (feared a 4th stamp nested in
  sub_workflow/substeps invisible to a top-level walk). Round 2 submitted 2026-08-25
  with the census above + the apply-gating answer (583/584 DO/RAISE pre-flights assert
  their predecessors; 583's was INDUCED pre-apply) + the INSERT column-exclusion answer
  (id/stamps excluded by explicit column list) + reuse answer (max_concurrent is DEAD
  config — STARTER §2; per-task executions fix = the D9 fork, owner-deferred).
- **Check the verdict** (budget ~30 min from submission; find by payload not printed id):
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%db9b7cbf%'
  ORDER BY created_at DESC LIMIT 1;` — full report:
  `SELECT body FROM diagnosis_artifacts WHERE kind='council_report' AND correlation_id
  LIKE 'db9b7cbf%' ORDER BY created_at DESC LIMIT 1;`
- If APPROVED: nothing to do (098 auto-credits the commit `dc76d1c30` via its
  `Council-Submitted:` trailer — do NOT amend, forward-only). If REVISE again: answer
  with measurements on the same correlation (`RESUBMIT_CORR=db9b7cbf…`), submission JSON
  at scratchpad `council_582_584.json` is regenerable from NOTES if the scratchpad is gone.

## OWED — the ordered queue (owner rulings applied)

1. **24h N=2 throughput read** (earliest meaningful: 2026-08-25 ~11:00Z): completions/hr
   vs the pre-change 24h (**629**, baseline 2026-08-24 10:3x, backlog 141 across 6
   sites) — ⚠ HOLD BACKLOG DEPTH BESIDE IT (demand control; a quiet queue reads as a
   failed fix). Meter: per-MINUTE `count(DISTINCT site_id)` of claims — NEVER 5-min
   buckets. Expect 2 in busy minutes. Queries: RUNBOOK.
2. **Same-site double-pick induction** (PLAN §4 safety argument): two manual dispatches
   at one site → confirm wasted-spawn-not-double-handle at `claimed_by`/`attempt_count`
   + the handler artefact. Still OWED; recorded in NOTES 08-24.
3. **Phase 3: batch 5→8 + `timeout_seconds` 300→600 in ONE migration** (D3 lockstep
   RULED). Only after (1) reads sane. Move `load_items.max_items` AND
   `process_item.max_iterations` together (one without the other is a silent no-op).
4. **D4 LLM spend governor** — first BUILD item (promotion prerequisite; gate for N>2).
   Owner's at-cap policy (reading recorded, unconfirmed — confirm before building):
   shed own-domain build/improvement first, keep maintenance, protect client work;
   governor acts BEFORE the hard cap. Note: platform code ⇒ council + image roll.
5. Then per RESEARCH §10 rulings: deploy batching (one commit per site per turn — D8
   interim), clients-first priority lane (D2 RULED — service-order change, prepare with
   measured p90 table), Batch API classes (D6), D16 retention proposal, per-class
   maintenance LLM cost measurement (RESEARCH §6, the one figure owed before pricing).
6. **DNS plan B: RULED start-now** — belongs to the domain programme lane; pointer left
   in `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/NOTES_site_delivery_and_editor.md`
   (2026-08-21 entry). Do not build here; check they picked it up.

## Hard constraints / traps for this lane

- **Do NOT add sibling #3.** N>2 is gated on the adapter await/retry decision + D4
  governor — measured 0/5 failure at five concurrent items (chassis lane, 07-28).
- The 090 on the scheduler claim FAILED at its verdict step (`max_tokens=32000`, run
  `a16b82cd-…`) — **no loop verdict exists; do NOT re-file** (same cap would kill it;
  class owned by `bugs_open/183`). The claim's status: strong first-hand, loop-unverified
  — stated in RESEARCH §2.1; keep stating it wherever cited.
- `bugs_open/029` (wedged loop / hung spawns) is ACTIVELY OWNED — six commits 08-19;
  contribute into their file, never re-file. Its mechanism section was corrected
  2026-07-21 in-file (NOT what the old PLAN text says).
- The build-cost figure (loanzy.uk: ~213 items / 410 orch (21% FAILED) / ~$20 LLM /
  ~10.5h) is **n=1**; a second attribution attempt returned a well-formed ZERO (site not
  named in payloads) — treat per-site cost attribution as an open gap.
- Council-gate spend dominates fleet LLM ($200/wk incl. 271.9M cache-reads, 08-19
  7d window) and scales with CHANGE volume, not domains — never divide it by domains.
- Fleet was chassis v1.0.1332 (rolled 08-24 09:39Z) at last read. Baselines in NOTES are
  dated per the counts-carry-dates ruling — re-run before believing any of them.

## Session hygiene reminders (this tree)

Pathspec commits only; grep LANDMINES for any symbol you touch; who-owns before touching
any bug; the workstreams memory line for this lane points here (COLD-START).
