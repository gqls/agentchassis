# RUNNING NOTES — Experience Loop

*Append-only. Newest entry last. Companion to PLAN_experience_loop.md and
RUNBOOK_experience_loop.md. House rule: every entry states what was done, what
was verified (and how), and what it changes for the next actor.*

---

## 2026-07-17 — Workstream ACTIVE: defaults accepted, RUNBOOK drafted (session "vonc4")

**Owner input**: "defaults accepted" — resolving PLAN §7 / HANDOFF §3d:

1. **Gauntlet in MVP = minimal-real** — a playable timed round against the daily
   provocation, client-side scoring, no leaderboard; every fabricated number
   stripped. Demotion to coming-soon only if the feasibility critic proves
   minimal-real can't be honest and small.
2. **Provocation detail pages = static daily-emitted pages** — consistent with
   the daily JSON emitter design. NOTE (verified this session): the daily
   emitter itself is NOT BUILT — `/data/provocations.json` is a hand-committed
   static sample; the pipeline exists only as
   `docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`.
   The council must scope this dependency in the MVP cut.
3. **Pilot fully autonomous on vonc** — artifact-verified checkpoints recorded
   here (CP1–CP5, defined in RUNBOOK §2), no approval gates.

**Done this session**: machinery survey (4 read-only exploration passes over the
acceptance ladder, council pattern, build/render side, claims verification) +
live-DB schema checks; RUNBOOK_experience_loop.md drafted from the findings;
PLAN header flipped PROPOSED → ACTIVE.

**Load-bearing facts established (each verified against code or live DB, refs in RUNBOOK §1):**

- `doc_plans`/`doc_notes` CHECK constraints allow only `('tool','pipeline')` —
  `subject_type='experience'` requires a migration on BOTH tables (RUNBOOK T1.1).
- browser-runner has NO navigation step action; each URL runs in a fresh browser.
  Tier-4 journeys are a genuine extension, not configuration (RUNBOOK T5.1).
- The arena's orphaning is a THREE-way key mismatch: criteria load by
  `content_components.function`, URL resolution by `pages.name = function`,
  travelling docs by `subject_key = function`. The page rename to `tool-arena`
  broke the middle link; re-keying doc_plans alone would NOT reconnect the sweep
  (RUNBOOK T4.2).
- `ReconcileSitePlanAction` emits `needs_page:<name>` → page-build-handler for
  EVERY plan page including tool-owned ones — the tool-role exclusion that
  `check_incomplete_page_group.go` already encodes (TP-004) is missing there.
  That asymmetry is guard rail 1's primary edit (RUNBOOK T2.1).
- No page-ownership marker column exists anywhere; TL-001 protection today is
  heuristic guards in `SavePageSectionsAction` + a manual park step.
- No code re-keys travelling docs on rename; no automated renamer exists either
  (the arena rename was manual). Root cause of the drift class:
  `create_tool_component_action.go:207` hardcodes `/tools/{function}.html`
  instead of calling `CanonicalisePage` (RUNBOOK T2.2).
- Claims verification V1a/V1b code is committed but NOT in any deployed image;
  vonc has NO evidence_base row, so the gauntlet's fabricated numbers are
  currently outside every enforcement lane (RUNBOOK T2.4).
- Migration sequence (`docs/agent_docs/sql_for_agents/`) runs through 161 with
  known collisions at 151 and 157; next free number at this snapshot = 162 —
  RE-CHECK at execution time.
- Concurrent-session state at drafting time: uncommitted WIP by another session
  in `platform/orchestration/actions/create_rerender_items_action.go` (a
  guard-rail-1 touchpoint) and an untracked live-probe test in
  `internal/adapters/browserrunner/` — re-check both before editing.

**Next actor starts at**: RUNBOOK §3 phase table — Phase 1 (foundations
migration) then Phase 2 (guard rails).
