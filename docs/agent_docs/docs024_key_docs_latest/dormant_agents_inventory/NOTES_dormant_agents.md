# NOTES — dormant-agents inventory (append-only, newest at the bottom)

## 2026-07-21 — build

Reproduced 044's fingerprint measurement live before building. Numbers moved
since 044 was filed (2026-07-20): 156/122/57 → **155/123/77**. The never-count
grew because more agents were seeded and the 55-day orchestration window keeps
advancing. Method itself is unchanged and reproduces exactly.

**Retention window matters.** `orchestration_states` holds ~106k rows, oldest
**2026-05-28** (~55 days). So "never observed" is really "never observed in the
RETAINED history." An agent that ran only before the window reads as
never-observed. This is not a bug in the method — it is why the ~34 legacy agents
(2025-08 → 2026-02) show as never-run: they genuinely have not run in the
retained window. Stated honestly in the report ("since <oldest>").

**MISSTEP I nearly made, then caught — feature-designer.** 044 lists
`feature-designer` as a validation case that "correctly detects as run." Live, it
is in the never-observed set (age 3.9d). I stopped to investigate before trusting
the method, because a false positive is the detector's whole failure mode.
Finding: feature-designer's 3 unique steps (`check_spec_approved`,
`load_council_report`, `load_spec`) appear **nowhere** in `orchestration_states`
— not as top-level keys, not even by full-text `LIKE` over `workflow_plan::text`.
So feature-designer's **own workflow has never run as an orchestration**. The
"designer half PROVEN 2026-07-18 (run 8e837814)" was the council approving a
plan, executed through other machinery (councils log as `agent_type='generic'`),
NOT the feature-designer agent's own workflow firing. So flagging it is
**correct**, and 044's use of it as a positive control was imprecise. The solid
positive controls remain fix-proposer / page-build-handler / section-editor.
> Lesson baked into the report: "observed" = "a unique top-level workflow step
> seen in retained history"; it can miss council/subtree execution, so the report
> is a triage inventory, not a verdict. Recorded in WRONG_CALLS.

**`orchestration_name` rejected as the mirrored-agent second signal.** 044
[INFERRED] it might close the blind spot. Live, every recent run's
`orchestration_name` is `''` or `generic-orchestrate-<timestamp>` — it does not
name the agent. So the blind spot (32 agents, incl. council-gate) stays
unmeasurable; the detector lists them and never flags them. No false positives
possible there by construction (they are not in `fingerprints`).

**Routing/status choices, verified against the code:**
- `site_work_items.site_id` is `NOT NULL` with a FK to `sites` → must anchor to a
  real site row. Used `triageSystemSiteID` (system.internal), which exists.
- claim/dispatch only touches `status IN ('triaged','approved')`
  (`claim_work_item_action.go:102`). `pipeline='maintenance'` + `status='dormant'`
  is never claimed.
- triage escalates only `status='failed'` and surfaces `deferred`/`capability_gap`
  — `status='dormant'` is invisible to it, so a dormant item is NOT mistaken for a
  code bug (which a `failed` item would be — that is why I did NOT reuse
  silent-check's `status='failed'`).
- `idx_swi_dedup` excludes the terminal set (complete/verified/rejected/wont_fix/
  **failed**/unresolved/cancelled). `dormant` is NOT excluded → `ON CONFLICT DO
  NOTHING` on `(site_id,item_key)` dedups cleanly. `complete` (close-out) IS
  excluded → a re-emission after close is allowed (matches triage/silent-check).
- No CHECK constraint on `site_work_items.status` (`pg_constraint` contype='c' →
  0 rows), so the bespoke `dormant` value is accepted.

**Timing.** Gather query (the only one scanning all 106k orchestration rows for
distinct step keys): **311 ms**. Seed timeout is 120 s. Comfortable.

**Seed assumptions verified live:** `diagnose-orchestrator` exists (infra-field
source), `snapshot_agent()` exists, `diagnosis-dormant-agents` does not yet exist,
unique constraint `agent_definitions_type_version_key` on `(type,version)` exists
(so `ON CONFLICT (type,version)` in the seed resolves).

**Tests:** 6 pure-function tests (item key stability, age-floor boundary at `>=`
not `>`, youngest-first emit order + no-mutation, spec JSON shape incl. the
owner_agent_type caveat, report honesty/completeness incl. retention date,
duplicate-active-rows flagged). All pass. Package builds clean on the shared tree.

**State:** code + seed + docs written and committed. Action is INERT until a
chassis image carrying it is rolled. 044 stays OPEN until then.
