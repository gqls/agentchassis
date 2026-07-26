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

## 2026-07-22 — live sweep + a substrate finding that changes the emission story

Chassis v1.0.1149 shipped with the action (pod-grep: `diagnose_dormant_agents`=4,
matching the silent-check control). Applied the seed, fired the trigger
(`TRIGGER_diagnosis_dormant_agents_v1.sh`, orch `4d8c433d…`). It **ran end-to-end
in ~3s** (dispatch was fast, not the ~29min queue) and wrote a dry-run report to
`doc_notes` (categories `dormant-agents`). The detector works.

**BUT the report exposed a substrate problem I had underweighted.** The report's
window read "since **2026-07-13**" — a **9-day** window, not the ~55 days I
measured on 07-21. And the never-count jumped 77 → **103**. Cause, run down live:

- `orchestration_states` went from **106,530 rows (oldest 05-28)** on 07-20 to
  **1,737 rows (oldest 07-13, 94% in the last 36h)** now.
- The `database-cleanup` scheduled task (hourly) does
  `DELETE FROM orchestration_states WHERE status IN ('COMPLETED','FAILED') AND
  updated_at < now()-INTERVAL '24 hours'`. So the real retention is **24h**; the
  55-day window on 07-20 was a transient (cleanup wasn't pruning then).
- Proof it over-flags: **`fix-proposer` is now flagged never-observed** — its
  unique step `check_confirmed` appears **nowhere** (its runs were pruned). A
  known-live agent, a false positive, purely from retention.

**No durable fallback exists.** `agent_definitions.usage_count` is **0 for every
agent** (162/162), including fix-proposer/section-editor/generic — the column is
unmaintained. `orchestration_state_audit` is also ~2 days. So there is NO durable
"has this agent ever run" signal — which is the exact lifetime question `044`
wants (section-editor had 3 lifetime runs when declared missing). Filed as
**bugs_open/060**.

**What I changed in response (committed):**
- **Window guard** in the action: when live (`!dry_run`) AND the observation
  window `< age_floor`, emit NOTHING and print a loud "WINDOW TOO SHORT" banner
  naming the prune + the dead usage_count. Prevents a future `dry_run=false` from
  flooding false positives. New unit test `…WindowTooShortBanner`.
- **Report reworded**: "never observed" is now explicitly a RECENT-ACTIVITY
  signal over the retained window, not lifetime; states the window depth and that
  fix-proposer-class agents are false positives.
- **Seed flip-warning**: DO NOT flip dry_run until the guarded build (>v1.0.1149)
  is live AND 060 is addressed. The LIVE binary (1149) has no guard — its only
  protection is dry_run=true, which I left in place.

**Net:** the detector is a sound, honest, windowed **inventory report** today
(discoverability — the core `044` value, LIVE and proven). Its **emission** is
correctly gated off until a durable run record exists (060). I did NOT build 060
(a hot-path write with the same generic-attribution crux as owner_agent_type) —
surfaced it for an owner call.

MISSTEP logged (WRONG_CALLS 2026-07-22): on 07-21 I treated the observed 55-day
`orchestration_states` window as a stable retention policy and reasoned from it;
it was a transient. The cheap check I skipped: read the cleanup job's retention
(`SELECT pre_query FROM scheduled_tasks WHERE name='database-cleanup'`) before
trusting `min(created_at)` as "the window".

## 2026-07-24 — 060 designed and built (uncommitted)

A continuation session designed the durable fix end-to-end
(`PLAN_2026-07-24_durable_run_record.md`): a new `agent_run_stats` table keyed
on the RESOLVED real agent type (not a revived `usage_count` — a dedicated
table stays isolated from the versioned `agent_definitions` row), a writer
wired into `processor.go`'s `executeWorkflow`, a defensive fallback in
`ai_actions.go` for the one execution-driving reader of `owner_agent_type`, and
the detector rewired to read the new table. Diagnosed the `owner_agent_type`
"generic" problem down to `determineOwnerAgentType` (`coordinator.go`)
preferring `Sender.AgentType`/env over the real resolved type, and mapped every
reader of `owner_agent_type` for safety before committing to the change. Left
uncommitted at end of session — code written, compiled, tested, but not landed.

## 2026-07-26 — adopted, shipped, live-verified; 060 and 044 both CLOSED

Found the 07-24 work sitting uncommitted while picking up 044. Read every
line, re-compiled and re-ran the tests against fresh `HEAD` (which had moved —
`coordinator.go`/`context.go`'s half of the `RunAgentType` plumbing had already
landed separately as a same-file passenger in `3af7b9d8d`, and a different
session's `034` fix had already carefully committed only its own 12-line hunk
out of `processor.go`, deliberately leaving the rest of that file's diff — my
diff — uncommitted rather than sweep it in wholesale; see that session's own
landmine note). Composed cleanly on top of both. Corrected 060's original
"verify against `section-editor`" bar — the shipped design is forward-only, no
backfill, and `section-editor` only runs a few times a year, so gated
verification on a frequent agent (`fix-proposer`/`page-build-handler`) instead.

Committed (`baf887a8e`), submitted to council review (advisory), applied
migration 203 by hand + `--record-only` (safe, additive), bumped `IMAGE_TAG` to
v1.0.1167, built from committed `HEAD`, pushed, and rolled — deliberately via
the single-service kustomize overlay + a direct `docker push`, **not**
`make deploy-agents`, which is fleet-wide and would have repointed every other
agent service at a tag only `agent-chassis` was ever built at.

**Live evidence, in order:**
- Pod-grep the new binary: `RecordAgentRun`=3, `agent_run_stats`=13,
  `run_agent_type`=2 (up from a 2-only baseline pre-writer).
- `agent_run_stats` filled with real traffic within ~2 minutes of the roll:
  `council-gate` (run_count=1) and `endpoint-health-checker` (run_count=2).
  `council-gate` is the exact agent the OLD method could never measure (its 099
  roster mirror has no unique step key) — the strongest available proof the
  resolution fix works, stronger than waiting on `section-editor`.
- Fresh generic-dispatched `orchestration_states` rows: `owner_agent_type` now
  names the real agent (`council-gate`, `endpoint-health-checker`), not
  `generic`.
- Live sweep (unchanged seed, `dry_run` still `true`): report correctly names
  "durable run record" as the method, states a 0.0-day tracking window
  honestly, and correctly excludes both agents above from the never-run list.

**Two message drops during verification, root-caused, not from this fix.**
`postgres-clients-0` crash-looped repeatedly right after the roll — a separate,
already-filed issue (`bugs_open/082`, a 1s liveness probe under CPU contention,
found independently the same day by another workstream). Chassis logs showed
the exact mechanism: the message was consumed, `FindByType` failed with
`SQLSTATE 08P01` mid-lookup, the processor fell back to an invalid synthetic
workflow, and errored out before ever writing an `orchestration_states` row —
so "message consumed, no row" without any visible error. Retried once
`postgres-clients-0` had held stable for 5 consecutive checks; both retries
completed cleanly. Diagnosed from the pod's own logs, not inferred — see
`bugs_closed/060`'s `§CLOSED` for the exact log lines.

**Council review submitted 3 times** under one correlation
(`2d2748e8-8a60-45a2-8cce-68148af9076e`); the first two runs stalled mid-review
on the same `082` instability (idle audit trail well past any normal review
duration, no error). Advisory only per CLAUDE.md — did not block closing
either bug. Will reconcile a `Council-Reviewed` trailer against commit
`baf887a8e` later if a verdict lands (no trailer was added at commit time;
forward-only precludes amending).

Both `060` and `044` moved to `bugs_closed/`.
