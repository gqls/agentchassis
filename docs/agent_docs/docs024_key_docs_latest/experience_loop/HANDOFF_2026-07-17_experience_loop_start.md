# HANDOFF — session "vonc3" close-out: link-integrity DONE, 42P10 incident FIXED, Experience Loop PROPOSED
*2026-07-17. Read top-to-bottom; this is the bootstrap for the next chat. The next
chat's expected front is the Experience Loop pilot (§3) — §1 and §2 are finished
context you should NOT redo.*

---

## 0. Bootstrap facts (verified at writing)

- **Live site**: vonc.com, site_id `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`.
- **DB**: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
- **Deployed chassis**: `v1.0.1128` (deployment AND all agent_definitions — verified).
  Commit `21e74808e` (three EXISTS/UPDATE status-filter alignments, non-urgent)
  is committed but NOT yet in a deployed image — rides the next roll.
- **Repo state**: branch `085_debug_and_feature_loops`; multiple concurrent
  sessions commit to it — re-run `git status`/`git log` before trusting any
  snapshot; commit per task with explicit pathspec (repo-root CLAUDE.md, read it).
- **Discovery is manual-trigger**:
  `scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh vonc.com completeness`.
  Never dispatch within ~300s of a chassis pod restart (spawn silently dropped).
- Judge a discovery run by `orchestration_states.status` of the
  **completeness-discovery-agent row** (the generic wrapper COMPLETES even when
  the child FAILS); zero new work items ≠ clean run.

## 1. FINISHED — link-integrity / recompute broadening (do not redo)

Full story: `docs/social001_vonc_tiktok_social/minilobby_task/`
(HANDOFF_link_integrity_arena_2026-07-16.md with 07-16/07-17 addendum →
RUNNING_NOTES_minilobby_task.md, entries dated 2026-07-16 afternoon + 2026-07-17).

- `ctaFieldNames` broadened (hero, call-to-action, archetype-grid,
  archetype-combinations, gauntlet-cta, content-block-about); migration **098**
  flipped the five `site_specs.*`-sourced url fields → `renderer`; all live and
  artifact-verified (the archetypes `/contact.html` button now → Gauntlet).
- End state on vonc: three `unresolved` misdirected_cta triplets
  (about/archetypes/index) = copy-vs-destination decisions correctly parked for
  the future content pass. NOT failures. Note: each qualifying discovery pass
  currently adds a duplicate unresolved triplet (unresolved sits outside
  idx_swi_dedup; flagged to fixloop workstream, not fixed).
- Remaining from that thread (lower priority): gauntlet/quiz `needs_rebuild`
  flag investigation; 096-style label unlock for archetype-grid /
  archetype-combinations THEN the content pass; quiz "Get Your Full Report"
  empty href (product decision).

## 2. FINISHED — the 42P10 fleet-wide insert breakage (do not redo; know the lesson)

Migration 157 (claims session) added `'cancelled'` to `idx_swi_dedup`'s excluded
statuses; Go's `workItemTerminalStatuses` (interpolated into insertWorkItem's
`ON CONFLICT ... WHERE`) stayed at six → arbiter inference failed → **every keyed
work-item insert failed SQLSTATE 42P10** and discovery run_checks transactions
aborted fleet-wide, invisibly (zero items looks clean). Fixed: `5e2711997`
(cancelled added; live in v1.0.1127+) + `21e74808e` (stale hardcoded copies
aligned; awaiting next image). Debugging entry: 016b_debugging_guide_8, entry
"split-contract-drift" (2026-07-17). Memory: `dedup-index-go-list-lockstep`.
**Rule: any migration touching idx_swi_dedup moves in lockstep with
workItemTerminalStatuses, same deploy window.**

## 3. THE NEXT FRONT — Experience Loop (proposed, nothing built)

**Read first**: `docs/agent_docs/docs024_key_docs_latest/experience_loop/PLAN_experience_loop.md`
(committed f9af5a330). Prompted by owner screenshots of the vonc Spark game.

### 3a. The diagnosis (artifact-verified 2026-07-17 — do not re-derive)

| Surface | Fact | Class |
|---|---|---|
| `/provocations/index.html` | archive items runtime-filled into a template with `href="#"`; per-provocation detail pages never planned/built (parked `needs_page:provocation`) | journey dead-end |
| `/tools/arena/index.html` | current `tool-arena-interface` component (23,353B, stored 07-14 17:02) has ONE script, four localStorage refs, **NO fetch of `/data/provocations.json`** (feed itself live, 200/5.6KB) → "Loading… DAY 0" forever. Its doc_plans PLAN (WITH acceptance criteria) is orphaned under old subject_key `tool-arena-interface` — page renamed to `tool-arena` (TL-003 reconcile), so `tool_acceptance_due` NEVER covers the live page | broken artifact + orphaned criteria |
| `/tools/gauntlet/index.html` | mock by construction: both CTAs `href="#"`; `gauntlet-interface.js` (3.9KB) does only strikethrough/timer/counters; FABRICATED stats (12,847 competitors, named leaderboard) violate the anti-fabrication rule; NO doc_plan (predates travelling docs) | mock shipped as product |

Component facts for the rebuild: `content_components.function='tool-arena-interface'`
is_active, 23KB template, `js_content` empty; `component_versions` has history
(join on component_id; js column lives on content_components, not cv). The
gauntlet component is `gauntlet-interface` (25KB tpl / 3.9KB js).

### 3b. The proposal in one paragraph

Lift travelling docs one level: an **experience-planner** writes an
EXPERIENCE_PLAN (`doc_plans`, `subject_type='experience'`, pilot key
`vonc-spark-game`) — journeys (page→control→action→observable), a promise ledger
(CTA copy → what the destination must deliver), data contracts, and an MVP cut
with the hard rule "absent or coming-soon, never simulated". A **challenge
council** (journey-completeness / feasibility / honesty / MVP-referee critics,
reusing the concept-register council pattern) attacks and revises it to
convergence. Build is contract-first (criteria attached at birth, as
tool-generator already does). Acceptance = the existing Tier-2/Tier-4 ladder
extended to multi-page journeys, with new failure scopes `data` and `plan-gap`
routing `needs_experience_replan` items back to the planner. Then feature
rounds. New-build list is PLAN §6 (small); everything else is reuse.

### 3c. Guard rails to build regardless of the pilot (PLAN §4)

1. **Page-ownership marker** — tool-owned pages make generic rebuild/rerender
   REFUSE (mechanises TL-001; the arena clobber is the proof).
2. **Rename re-keys travelling docs** (or sweep resolves aliases) — the orphaned
   arena PLAN is the proof. Cheap immediate win available: UPDATE the existing
   doc_plans row's subject_key to the live tool key so the acceptance sweep
   picks the Arena up again — but ONLY use it deliberately; on the current
   broken page it will (correctly) produce a failing verdict + improve_tool.
3. **Tier-2 dead-control check** — `href="#"`/no-op interactive elements on a
   deployed page = acceptance failure.
4. **Quantitative copy must trace to a data contract** (claims-verification
   machinery; the gauntlet numbers are its jurisdiction).

### 3d. Owner decisions pending (defaults proposed in PLAN §7)

1. Gauntlet in MVP: minimal-real playable round (default) vs honest demotion.
2. Provocation detail pages: static daily-emitted pages (default) vs
   client-side detail rendering.
3. Fully autonomous pilot on vonc (default yes, artifact-verified checkpoints).

**The owner has NOT yet answered these.** First move of the next chat: get the
three answers (or confirm defaults), then draft the RUNBOOK for phase order —
suggested: guard rails 1+2 first (they protect everything else), then
experience-planner + council on the Spark game, then the MVP build round.

### 3e. Machinery you will reuse (state of each, 2026-07-17)

- **Travelling docs**: `doc_plans`/`doc_notes` live; 11 plan rows; tools write
  PLANs at birth. See travelling_docs/STATUS_2026-07-16_where_we_are.md.
- **Acceptance ladder**: Tier-4 desktop+mobile, interactions (P2), overflow
  (P1), screenshots-on-failure (P3, proven v1.0.1125). `tool_acceptance_due`
  sweep is continuous. browser-runner-adapter is a separate image (was being
  extended by another session — check before assuming).
- **Council pattern**: concept-register council live in prod (3 reviewers,
  fix-proposer v6).
- **Claims verification**: V0+V1 built, evidence_base live (check deploy state
  before relying).
- **Fixloop / diagnose loop**: complete, first real case confirmed 2026-07-16.

## 4. Landmines for the next chat (hard-won, current)

- Multiple sessions share tree/branch/cluster/tag-sequence: expect commits and
  image rolls you didn't make (this session's Go was swept into other sessions'
  commits twice; v1.0.1125/1126/1128 were rolled by others mid-task). Verify
  deployed code IN-POD by binary string (`grep -ac '<literal>' /app/agent-chassis`),
  never by git or tag.
- `make build-agent-chassis-ref REF=<sha> IMAGE_TAG=<next>` builds from a
  committed ref, no WIP — use it, bump the tag, then `kubectl set image` on
  agent-chassis + business-intel + vet-intel (container name `agent` on the
  intel pair) AND `UPDATE agent_definitions SET image_tag=...` (spawned Jobs
  read the DB tag, effective instantly).
- Spawned-pod logs vanish fast (job-cleanup ~5min): snapshot logs BEFORE the
  pod completes; filter for the FIRST error (later 25P02/etc are cascade).
- TL-001 still un-mechanised: nothing stops a generic rebuild clobbering
  `/tools/arena/index.html` again until guard-rail 1 exists.
- `provocation-card`/`lobby-grid` are runtime-fill shells — leave blank fields
  alone. `needs_page:provocation` stays parked until the experience spec gives
  it a definition (that IS pilot work now).
- The three misdirected_cta `unresolved` triplets are EXPECTED state, not bugs.

## 5. This session's commits (for archaeology)

`5e2711997` (42P10 Go fix) · `21e74808e` (filter alignments) · `2cbfede90`
(link-integrity close-out docs) · `f9af5a330` (Experience Loop PLAN). The
broadened-recompute Go itself was swept into other sessions' commits
`e10c656f3`/`d076c3c8e` on 07-16; migration 098 + notes swept into `d076c3c8e`/
`6880c669e`. Memory files updated: `vonc-spark-workstream` (item 8),
`dedup-index-go-list-lockstep` (new), `experience-loop-workstream` (new).
