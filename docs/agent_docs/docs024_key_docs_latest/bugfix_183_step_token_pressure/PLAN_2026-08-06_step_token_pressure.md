# PLAN 2026-08-06 — bugs_open/183: the fleet gets a standing cap-headroom check

**Bug:** `bugs_open/183_HANDOFF_2026-08-03_classifier_output_cap_is_the_fleets_only_6000_and_now_blocks_adoption.md`
**Lane opened:** 2026-08-06, this directory. No prior lane for 183 existed (checked
`ls -lat docs024_key_docs_latest/` and transcript symbol-heat immediately before creating).

## Why this bug, and why it is still valid

- Ownership checked three ways 2026-08-06 morning: `who-owns.py` (no owning
  workstream), open `site_work_items` (zero rows), and live-transcript symbol grep.
  The only heat on `domain-research-classifier` is an adjacent Firecrawl-audit lane
  which itself recorded "domain-research-classifier has NO owner". 186 and 189 were
  considered and rejected: 186's fix site (`store/instances.go`) is being read by an
  active session right now; 189's fix site (`matchLockedRow`) is inside the active
  204 lane's council argument.
- Still valid, re-verified live: cap now 32000 (unshadowed — root block NULL), last
  run 2026-08-04 succeeded at 4295 tokens. The ACUTE symptom is gone. The bug is
  OPEN because the CLASS exposure is untouched: nothing anywhere warns when a step's
  output distribution approaches its configured cap. 183's step sat at p95 = 92.5%
  of cap for months before it burned three attempts per site.

## The decision: candidate 4, generalised, as a driven online check

183's fix candidates, and what this plan does with each:

1. Cap normalisation — **already applied** (6000 → 32000, live, unshadowed).
2. No salvage for this step — **respected**; we add no `repairTruncatedJSON` here.
   (Trailing-section absence would silently drop `design_intent.palette`, the part
   the pipeline actually reads. 183's reasoning stands.)
3. Split the one four-document step into four bounded generations — **recorded as
   the make-unrepresentable fix for THIS agent, not taken in this lane.** Reasons:
   it restructures a live 15-step agent mid-flight while adoption lanes are using
   it; at cap 32000 the observed max (6590) has ~5x headroom; and with a standing
   pressure check (below), regrowth toward the cap is announced long before it can
   burn a site again. If the check ever flags this step, candidate 3 is the answer
   — that condition is written into the bug file.
4. A check on caps — **this plan, but distribution-shaped, not a static floor.**
   183 asked for an offline floor lint ("no spec step below the 8000 mode"). An
   offline lint is a script nobody runs (the estate's own repeated finding), and a
   static floor needs a judgement call about document classes. The distribution
   form subsumes it: a step whose p95/peak approaches its cap is flagged whatever
   its cap is, and a step at 4000 that never goes near it is left alone.

**The vehicle: a new CTE-only `scheduled_tasks` row cloned from FIX-058
(`council-seat-token-pressure`, `104_TASK_seat_token_pressure_v1.sql`).**
That instrument is exactly this check, proven live (ran 2026-08-06 03:21, notes on
07-31/08-01/08-04), scoped to `step_name LIKE 'review_%'` only. Precedent alignment:

- RFC_006: live config is guarded by a **scheduled** check, never a commit-time
  hook — at commit time the config is unapplied, and caps change in the DB with no
  commit at all.
- RFC_012 second sitting (owner, 2026-08-06): standing checks should be **online
  within the framework** where the framework allows. `fire_message=false` means the
  `pre_query` IS the work: no Kafka, no orchestration, no LLM, no credits.
- Delivery is `doc_notes` (the proven surface). NOT `site_work_items
  status='detected'` — that status has no promoter and no cadence (bugs_open/083:
  98 rows parked since July).

## Design of `fleet-step-token-pressure`

One `scheduled_tasks` row, `fire_message=false`, `interval_seconds=21600` (6h,
matching FIX-058), `pre_query` computing per **(step_name, cap)** over a 14-day
`llm_call_log` window. (step_name, cap) and not (agent_type, step_name, cap),
for FIX-058's own reason: `agent_type` was relabelled 2026-07-26 (`generic` → the
resolved type), so an agent_type key splits one population in two at the relabel
line — today's 14-day window still reaches back past it. Keying on (step_name, cap)
merges the halves correctly; the note prints the distinct agent_types observed as
attribution labels, never as the unit of measurement:

- **Truncation count from `error_message ILIKE '%stop_reason=max_tokens%' OR
  ILIKE '%done_reason=length%'`** — never from `output_tokens >= max_tokens`
  (truncated first attempts log `output_tokens = NULL`; the token comparison found
  4 where the truth was 94). A truncated call scores `frac = 1.0`.
- **Current-cap discipline via OBSERVED caps, not a definitions join**: a group is
  reported only at the cap of the most recent call for that (agent_type,
  step_name). This sidesteps bugs_open/009 (a root `ai_service` block makes
  step-level `max_tokens` dead config — a jsonb path read into definitions would
  report shadowed values as live) because `llm_call_log.max_tokens` records what
  the code actually resolved. Cost, stated honestly: a cap raised in config with no
  call since is invisible until the next call — acceptable for an advisory
  leading-indicator; the note names the cap it measured.
- **Exclusion: `step_name NOT LIKE 'review_%'`** — those pairs are FIX-058's, same
  thresholds, already announced; double-announcing them would be alert fatigue and
  a second copy of the rule. The two tasks partition the fleet exactly.
- **Thresholds copied from FIX-058, one deliberate change**: NEAR-MISS peak ≥ 95%,
  PRESSURE p95 ≥ 85%, any truncation flags — but **n ≥ 5**, not FIX-058's n ≥ 20.
  Reason, measured: 183's own step had 11 calls in the last 21 days; at n ≥ 20 the
  check cannot flag the very case it exists for. (FIX-058's n ≥ 20 is right for
  council seats at hundreds of calls per fortnight; fleet steps are sparser.)
- **Event, not heartbeat**: `subject_key` carries an md5 digest of the flagged set;
  insert skipped when a note with that key exists (30-day look-back), so a
  persisting condition announces once and an escalation re-announces.
- **agent_type trap**: rows before 2026-07-26 ~15:00 log `agent_type='generic'`
  — and a 14-day window from today still reaches to 07-23, INSIDE the relabelled
  era (my own first draft of this plan claimed the window was clear; it is not,
  until 08-09). Handled structurally, not by waiting: the key is (step_name, cap),
  agent_type is display-only, and the note prints n.

**Territory, from the prior-art sweep (2026-08-06):** `bugs_closed/138` is CLOSED
(owner, 2026-08-03), its lane owes nothing, and its own `apply-seat-length-budget.py`
excluded `domain-research-classifier` with a printed reason — the boundary is drawn
short of this agent. Generalising FIX-058 beyond `review_%` is claimed by nobody.
NOT touched here: any council seat's cap or prompt (owner-ruled, 138), the
`Degraded` gating rule, `repairTruncatedJSON` semantics (fixed+verified 08-03).
FIX-058's open register question (should the near-miss threshold scale with cap
size?) stays open — this check inherits the thresholds as they are and the register
entry points at that question rather than silently answering it.

## Edits (all committed, pathspec, one task)

1. `docs/.../bugfix_183_step_token_pressure/SQL_2026-08-06_fleet_step_token_pressure_task.sql`
   — the seed, apply/remove instructions in-header, modelled on 104_TASK.
2. `docs/.../bugfix_183_step_token_pressure/RUNBOOK_step_token_pressure.md` — the
   report queries for humans (pointing at the pre_query for thresholds, never
   re-encoding them — the 099/102 drift rule).
3. Concept-register entry (new reusable mechanism — another workstream could need
   this and not know it exists), **same commit as the seed** per the seam ruling.
4. `bugs_open/183` updated: candidate 4 state, candidate 3's trigger condition,
   the corrected fleet measurement.
5. WRONG_CALLS.md: this session's own first-pass headroom query inherited the
   output_tokens-NULL blindness (details in NOTES).

## Council

Submitted to the council gate before/alongside the commit per CLAUDE.md. Note the
gate's stated footprint is `platform/`, `internal/`, `pkg/` and refuses docs paths
client-side; this change is live-config + docs. If the trigger refuses, that is
recorded here rather than gamed, and review falls to the RFC_006 precedent (the
owner sees the seed, the register entry, and the first note).

## Verification (the disconfirmable kind)

1. **The check must flag the case it exists for**: run the pre_query with its
   window pinned over 2026-08-02 — `classify_and_extract@6000` MUST appear
   (5 truncations + p95 over threshold). A check that cannot flag its own
   motivating case is inert (the 5d1df2777 lesson).
2. Run live: enumerate today's flagged set and record it in NOTES. Known from the
   corrected hand-run: `extract_and_reconcile@2048` (vet-practice-verifier,
   63/133 truncated), `llm_audit@4000` (tool-auditor), `suggest_tools@3000`,
   `compose/recompose@16000` (experience-planner), others — the note is the report.
3. **Negative control**: a step with comfortable headroom (e.g.
   `stage_implement@32000`, p95 24%) must NOT appear.
4. After seeding: confirm `last_triggered_at` advances within one interval and the
   first note lands in `doc_notes`. A missing row after 6h means the job did not
   run — which is a different finding from "nothing is wrong".

## Decisions log

- 2026-08-06: bug picked, validity re-verified, design chosen (this file).
