# RUNBOOK — Experience Loop (pilot: vonc Spark game)

*Drafted 2026-07-17, session "vonc4", from a four-way machinery survey (acceptance
ladder, council pattern, build/render side, claims verification) + live-DB checks.
Companion to PLAN_experience_loop.md (the why) and RUNNING_NOTES_experience_loop.md
(the log). This is the execution doc: phases, tasks, exact touchpoints, proofs.
Nothing below has been built unless RUNNING_NOTES says so.*

---

## 0. Operating rules (read before every phase)

Distilled from repo-root CLAUDE.md + HANDOFF §4; they are not optional.

- **Concurrency**: many sessions share tree/branch/cluster/tag-sequence. Re-run
  `git status`/`git log` before trusting anything; commit per task with explicit
  pathspec; forward-only. At drafting time another session had uncommitted WIP in
  `platform/orchestration/actions/create_rerender_items_action.go` (a T2.1
  touchpoint) and an untracked probe test in `internal/adapters/browserrunner/`
  (the T5.1 image) — re-check both before editing those areas.
- **Deploy discipline** (UPDATED 2026-07-17, build default inverted fleet-wide):
  `make build-agent-chassis IMAGE_TAG=<next>` now builds from **committed
  HEAD** (git-archive; structurally cannot bundle WIP; `REF=<ref>` pins,
  `-tree` is the deliberate WIP escape hatch). Commit your task, then build.
  Bump the tag every build; `kubectl set image` on agent-chassis +
  business-intel + vet-intel (container `agent` on the intel pair) AND
  `UPDATE agent_definitions SET image_tag='<next>'` (spawned Jobs read the DB tag).
  Verify IN-POD by binary string, never by git/tag:
  `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<new symbol>"'`.
- **Ordering**: image → seed → fire (a workflow naming an unregistered action
  fails at that step; old workflow on new binary is harmless). No orchestration
  dispatch within ~300s of a chassis pod (re)start. Snapshot spawned-pod logs
  before completion (job-cleanup ~5min); read the FIRST error, later 25P02 etc.
  are cascade.
- **Discovery verdicts**: judge a run by the **completeness-discovery-agent row**
  in `orchestration_states` (the generic wrapper COMPLETES when the child FAILS);
  zero new items ≠ clean.
- **Migrations**: numbered SQL in `docs/agent_docs/sql_for_agents/`; sequence ran
  through **161** at drafting (151 and 157 both collided across sessions). Claim
  the next free number by `ls | sort` AT EXECUTION TIME; never reuse.
- **Model config gotcha (MDL-039)**: a root-level `ai_service` in
  `default_config` SHADOWS every step-level `ai_service`
  (`ai_actions.go:146-193`). New agent defs: NO root `ai_service`; configure
  per-step.
- **Work items**: build `item_key` via `workItemKey(itemType, target)`
  (`work_items_common.go:84`) so prefix == item_type. `idx_swi_dedup`'s status
  set and `workItemTerminalStatuses` (`work_items_common.go:29-44`) move in
  lockstep, same deploy window — this plan adds NO new statuses, so no change;
  if that ever changes, 42P10 fleet-wide is the failure mode.

## 1. Ground truth (verified 2026-07-17; re-verify only if something contradicts)

| # | Fact | Where |
|---|---|---|
| G1 | `doc_plans`/`doc_notes` CHECK constraints allow subject_type `('tool','pipeline')` only | live `\d doc_plans` / `\d doc_notes`; DDL `sql_for_agents/125_doc_plans_and_notes.sql` |
| G2 | One-current-per-subject enforced by partial unique `idx_doc_plans_current (subject_type, subject_key) WHERE is_current` | same |
| G3 | Doc actions: `write_doc_plan` (supersede+insert, one tx), `append_doc_note`, `load_doc_context` (extracts the ```criteria fence → `criteria_json`) | `platform/orchestration/actions/{write_doc_plan,append_doc_note,load_doc_context}_action.go`; registry.go:1654-1672 |
| G4 | tool-generator writes a PLAN at birth (steps `compose_plan → write_plan → index_plan`, `error_step: complete`) | `sql_for_agents/131_tool_generator_plan_writing.sql` |
| G5 | tool-acceptance-agent = `ensure_site_record → load_docs → request_run → judge`; skip (never fake-pass) when criteria empty | `sql_for_agents/145_tool_acceptance_agent.sql`; `tool_acceptance_actions.go` |
| G6 | URL resolution in `RequestBrowserRunAction` is `pages.name = <function>` (`:127`); criteria load keys on `content_components.function`; travelling docs key on `subject_key = function`. The arena page rename to `tool-arena` broke the middle link while `function` stayed `tool-arena-interface` — re-keying doc_plans ALONE does not reconnect the sweep | `tool_acceptance_actions.go`; live doc_plans rows |
| G7 | `tool_acceptance_due` is a Go discovery check (7-day cooldown on `source='tool-acceptance'` notes), enabled by appending to the discovery agent's `checks` array | `discovery_checks/check_tool_acceptance_due.go`; `146_enable_tool_acceptance_due.sql` |
| G8 | Discovery checks self-register in `init()` via `discovery_checks/registry.go`; runner reads the `checks` array from step config; unknown names logged+skipped | `discovery_checks.go:37`; `registry.go` |
| G9 | browser-runner step actions are exactly `fill|click|select`; **no goto/navigate**; each URL = fresh browser, no shared state. Journeys are a real extension. Separate image (`cmd/browser-runner-adapter`) | `internal/adapters/browserrunner/run_checks_action.go` (`Do` :685, `Execute` :241) |
| G10 | Failure scopes today: `tool|chrome|unknown` (`run_checks_action.go:83`); routing split in `extractRunResults` (:357) + `JudgeAcceptanceResultsAction` (:467) → improve_tool vs responsive_fix | same files |
| G11 | Tier-2 static checks + built-in shell checks live in `evaluateStaticCriteria` (`check_tool_acceptance.go:329`, shell checks :390-398) — the dead-control check's home | that file |
| G12 | No page-ownership marker exists on any table; TL-001 protection = heuristic guards in `SavePageSectionsAction` (:335 content-regression, :373 interactivity-regression) + a MANUAL park step after each reconcile | survey; `094_vonc_arena_tool.sql:29,103` |
| G13 | `ReconcileSitePlanAction` emits `needs_page:<name>` → page-build-handler for EVERY unbuilt plan page incl. tool-owned; the tool/game-role exclusion exists only in `check_incomplete_page_group.go` (TP-004) | `reconcile_site_plan_action.go` (`decideEmit` :293) |
| G14 | No code re-keys travelling docs on rename; no automated renamer exists (arena rename was manual). Drift root cause: `create_tool_component_action.go:207-208` hardcodes `/tools/{function}.html` instead of `datahelpers.CanonicalisePage` (`page_canonical.go:106`) | survey |
| G15 | Council pattern = ONE agent workflow: sequential `execute_llm_prompt` reviewer steps → deterministic Go `diagnose_council_decide` (verdicts `approve|object|veto`, `hard_veto_from`, `max_rounds: 3`, round state = `diagnosis_artifacts` rows `kind='council_report'` counted by orchestration_id) → conditional router (approve→complete / veto→reframe-or-escalate / object→repropose) | `fixloop_eg_dartsonline/0NN_fix_proposer_v6_bug_historian.sql`; `diagnose_council_decide_action.go` |
| G16 | New-agent seeding convention: copy `image_repository/image_tag/command/resources/topics/health_config/env_vars` FROM an existing def; `snapshot_agent()` before update; upsert `ON CONFLICT (type, version)` | fix-proposer SQL :81-93 |
| G17 | Trigger convention: durable intake work item (inert status, `ON CONFLICT DO NOTHING`) + kcat `orchestrate` envelope on the SAME correlation_id; coverage probe refuses on open items touching the target unless FORCE=1 | `fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh` |
| G18 | Claims V1a (build gate, `validate_page_content.go` Check 8) + V1b (discovery check `unverified_claims`) are committed but in NO deployed image; shared engine `datahelpers/claims.go`; operator CLI `cmd/claimscan`. UPDATE 2026-07-17: `unverified_claims` is already pre-enabled in quality-discovery-agent's checks array (activates when the image ships), and vonc's `evidence_base` now EXISTS (migration 166, T2.4 done) | claims RUNNING_NOTES; survey; live checks-array query |
| G19 | `/data/provocations.json` daily emitter is NOT BUILT — live file is a hand-committed sample; design in `docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`; JSON shape fixed by the client loader (`today/lobby/arena` keys); model action `render_news_section_action.go` | survey |
| G20 | vonc publishes via git-adapter commits to `{domain}/{filename}` in the site repo; `DeployToHostingAction` is a stub; verify = `curl https://vonc.com/...` | `git_deployer_actions.go`; `github_client.go:59-70` |
| G21 | `item_type` is free-form (no CHECK, no central enum); routing = emitter names `handler_agent`; two-strike + `unresolved` semantics come free via `insertWorkItem` | `load_work_item_actions.go:511,987` |
| G22 | Runtime-fill shells (provocation-card, lobby-grid) legitimately render blank/`#` BEFORE client-side hydration; the phantom-links check already carries a runtime-fill guard (`9752bc68d`) — any dead-control check must reuse it | `check_phantom_internal_links.go` |
| G23 | The three misdirected_cta `unresolved` triplets on vonc are EXPECTED (parked copy decisions), and each qualifying discovery pass adds a duplicate triplet (unresolved sits outside idx_swi_dedup — flagged to fixloop, not ours) | HANDOFF §1 |

## 2. Decisions in force + checkpoints

**Decisions (owner accepted PLAN §7 defaults, 2026-07-17):**
- **D1** Gauntlet MVP = minimal-real playable round (client-side scoring, no
  leaderboard, zero fabricated numbers). Demote to coming-soon ONLY if the
  feasibility critic proves minimal-real can't be honest and small.
- **D2** Provocation detail = static daily-emitted pages. The daily emitter is
  unbuilt (G19) — the council MUST scope it explicitly in the MVP cut (in-MVP,
  or MVP builds detail pages for the sample feed and the emitter is round 2).
- **D3** Pilot fully autonomous on vonc; artifact-verified checkpoints in
  RUNNING_NOTES, no approval gates. One decision menu per round boundary at most.

**Checkpoints (each = a RUNNING_NOTES entry with the verifying artifact):**
- **CP1** Guard rails live: reconcile on vonc routes tool-owned pages to
  needs_human_review (no handler); a forced generic section-save against a
  marked page REFUSES; in-pod string proof of the new binary.
- **CP2** EXPERIENCE_PLAN `vonc-spark-game` is_current in doc_plans with all five
  sections; council trail present (≥1 `council_report` artifact + doc_notes per
  round); converged (not escalated) or escalation package written.
- **CP3** MVP built + deployed: `curl` proofs — arena page fetches the live feed;
  detail pages resolve from archive links; gauntlet playable; `claimscan` exit 0
  on vonc export.
- **CP4** One full journey GREEN end-to-end via the ladder on live vonc, AND an
  induced `data`-scope failure produced a `needs_experience_replan` item routed
  to the planner (drill subject, then cleaned up).
- **CP5** Continuity: two consecutive scheduled sweeps (no manual trigger) each
  produced/refreshed experience acceptance verdicts.

## 3. Phase map

| Phase | What | Depends on | Deploys |
|---|---|---|---|
| 1 | Foundations: `experience` subject_type migration | — | DB only |
| 2 | Guard rails 1–4 (ownership marker, rename re-key, dead-control, claims lane) | 1 | chassis image roll A + DB |
| 3 | experience-planner + challenge council; run to convergence on `vonc-spark-game` | 1 (2 strongly preferred) | DB only (G15: all existing actions) |
| 4 | MVP build round: arena re-attach + rebuild, detail pages, gauntlet minimal-real, claims-clean | 2, 3 | site content + maybe small chassis bits |
| 5 | Journey acceptance: browser-runner journeys, `data`/`plan-gap` scopes, `needs_experience_replan`, `experience_acceptance_due` sweep | 2 (image A), parts can start parallel to 3/4 | browser-runner image + chassis image roll B + DB |
| 6 | Feature rounds off the LATER list | 5 GREEN | per round |

Rationale: guard rails first because Phase 4 rebuilds the exact pages the current
machinery likes to clobber (G12/G13); the council before the build because
contract-first is the point; Phase 5's Go can be developed in parallel but its
sweep goes live only after MVP pages exist to verify.

---

## Phase 1 — Foundations (DB migration, live immediately)

**T1.1 — migration `<next-free>_doc_subjects_experience.sql`** (162 at drafting;
re-check): on BOTH `doc_plans` and `doc_notes`, drop and re-add the
`*_subject_type_check` CHECK as `IN ('tool','pipeline','experience')` (G1).
Additive; no code change; existing rows untouched. Apply via psql; record in the
migration ledger per house convention.
**Verify**: `INSERT` a throwaway `doc_notes` row with `subject_type='experience'`,
then delete it; `\d doc_plans` shows the new constraint.

**T1.2 — subject conventions** (document in the migration header): experience
subject_key = kebab experience name (`vonc-spark-game`); drills use suffix
`-drill` and are superseded/cleaned after use (T5.4).

## Phase 2 — Guard rails (chassis image roll A)

All Go lands as ONE ref-built image (plus the already-committed claims V1a/V1b
and `21e74808e`, which ride the same roll — G18, HANDOFF §0).

**T2.1 — Guard rail 1: page-ownership marker.**
- Migration: `ALTER TABLE pages ADD COLUMN rebuild_policy text NOT NULL DEFAULT
  'generic' CHECK (rebuild_policy IN ('generic','owned'))`. Seed `'owned'` for
  vonc's `tool-arena`, `provocations`, `gauntlet` pages and every
  `page_type='tool'` page fleet-wide.
- Go refuse points (G12/G13):
  1. `ReconcileSitePlanAction.decideEmit` (`reconcile_site_plan_action.go:293`):
     plan pages whose realised row is `owned` (or role tool/game, mirroring
     `check_incomplete_page_group.go`) emit `needs_human_review` with NO handler
     instead of `needs_page` → page-build-handler. **This retires the manual
     park step for `needs_page:provocation`.**
  2. `SavePageSectionsAction`: hard REFUSE (structured error + doc_note) when the
     target page is `owned` — the generic re-plan path must not touch it. The
     section-editor's `apply_section_edit` path stays allowed.
  3. Explicitly DO NOT block `page_rerender`/`AssemblePageAction` — re-assembling
     existing `page_components` is how owned pages deploy.
- Check `create_rerender_items_action.go` dirty state before touching anything
  near it (§0).

**T2.2 — Guard rail 2: rename re-keys the travelling docs.**
- New datahelper `RekeyTravellingDocs(tx, subjectType, oldKey, newKey)`: UPDATE
  both tables' `subject_key`, then append a doc_note under the NEW key recording
  the re-key (category `rekey`). Beware G2: if a current row already exists under
  newKey, supersede it first (write_doc_plan semantics), never violate
  `idx_doc_plans_current`.
- Call sites: `reconcilePlanWithRealised` Pass B (`v3_site_actions.go:4640` — the
  only automated place identity snaps today) + expose as a registered action
  (`rekey_travelling_docs`) so manual renames stop being silent.
- Root-cause fix (G14): `create_tool_component_action.go:207-208` → derive
  name/url via `datahelpers.CanonicalisePage`, killing future function/slug
  drift at birth (TL-003 class).

**T2.3 — Guard rail 3: dead-control check (Tier-2 lane).**
- Shared helper (near `datahelpers`): scan HTML for `<a href="#">`/empty-href/
  `<button>` with no handler; MUST reuse the runtime-fill guard (G22) so
  pre-hydration shells don't false-positive.
- Wire twice: (a) built-in shell check in `evaluateStaticCriteria`
  (`check_tool_acceptance.go:390-398`) for tool pages under acceptance; (b) new
  discovery check `dead_controls` (registry `init()`, G8) over deployed pages,
  emitting `needs_human_review` items (no auto-fixer yet — first landing is
  detect-and-surface). Enable via `jsonb_set` append AFTER image roll A (G7/G8
  ordering). The post-hydration (Tier-4) variant is T5.1's job.

**T2.4 — Guard rail 4: claims lane for vonc.**
- Create vonc's `evidence_base` site_specs row (pattern: leopardess rev 4, G18):
  facts = only what is real (likely near-empty at first), banned patterns for the
  known fabrications (`12,847`, `94,210`, competitor/leaderboard-name classes).
- AFTER image roll A: append `"unverified_claims"` to a vonc-covering discovery
  agent's `checks` array (V1b activation); V1a activates by itself once the code
  ships. Run `cmd/claimscan` against a vonc component export for the baseline
  finding list (expect gauntlet hits — they get stripped in T4.4).

**T2.5 — Build, roll, prove (CP1).**
Commit per task as T2.x complete → `make build-agent-chassis-ref REF=HEAD
IMAGE_TAG=<next>` → roll per §0 → in-pod string check on a new symbol (e.g.
`rebuild_policy`) → trigger a completeness discovery on vonc (075 script; §0
300s rule) → verify: owned pages routed to needs_human_review; forced
section-save refused; write CP1 to RUNNING_NOTES with artifacts.

## Phase 3 — Experience planner + challenge council (DB-only)

**T3.1 — agent definition `experience-planner`** (one workflow = planner AND
council, on the fix-proposer v6 template, G15/G16). New numbered seed in
`sql_for_agents/`. No root `ai_service` (§0 MDL-039). Steps:

1. `ensure_site_record`
2. `load_context` — `query_database`: site pages (+`rebuild_policy`), components,
   nav, feed URLs, current tool doc_plans, open work items for the site
3. `compose_plan` — `execute_llm_prompt`, per-step ai_service claude-sonnet-5,
   max_tokens 16000, temperature 0.2; prompt mandates the five EXPERIENCE_PLAN
   sections (T3.2) incl. a ```criteria fence in the runner grammar
4. `persist_plan` — `write_doc_plan` `subject_type='experience'`,
   `subject_key_field='input_data.experience_key'`, `plan_source='experience-planner'`
5. `review_journeys` / `review_feasibility` / `review_honesty` / `review_mvp` —
   four sequential `execute_llm_prompt` steps (sonnet-5, max_tokens 3000,
   temperature 0.0, strict JSON verdict `{reviewer, verdict: approve|object|veto,
   objections[], missing[], checks[], notes}`), lenses per PLAN §3-B. Include the
   `load_schema_hint` trick (G15) so any `checks[]` SQL is column-accurate.
   Veto rights: journeys/feasibility/honesty may veto; **`hard_veto_from:
   ['review_honesty']`** (fabrication is the cardinal rule); MVP referee is
   advisory (`approve|object` only — the bug-historian precedent, G15).
6. `council_decide` — `diagnose_council_decide` verbatim, `max_rounds: 3`
7. Router conditionals exactly as fix-proposer v6: approved→`append_note`+
   `complete`; revise→`repropose` (compose with objections injected, loops to 4);
   veto→one `reframe` then `escalate` (persists `kind='escalation'` artifact +
   `needs_human_review` item). Every round: `append_doc_note` (subject =
   experience key, category `experience-council`).

**Verify before firing**: every action name in the workflow exists in
`registry.go` on the RUNNING image (G3, G15) — this is what makes Phase 3
DB-only; if anything is missing it becomes part of image roll B.

**T3.2 — EXPERIENCE_PLAN body contract** (enforced by compose prompt + the
journey critic): §1 Journeys (each step: page, control selector, action,
observable outcome); §2 Promise ledger (CTA copy → what the destination must
deliver); §3 Data contracts (`/data/provocations.json` shape per G19 incl. the
reserved `arena` key; who writes it, when; client-side-only constraint); §4 MVP
cut + LATER list (hard rule: absent or labelled coming-soon, never simulated; no
dead controls; no unregistered numbers — D1/D2 constraints stated); §5 per-journey
```criteria``` in the Tier-2/Tier-4 grammar (T5.1's `journey` type; selectors
real, anchor rule).

**T3.3 — trigger script** `0NN_TRIGGER_experience_plan.sh` on the 090 template
(G17): durable intake item `item_type='needs_experience_plan'` (inert status,
key `needs_experience_plan:vonc-spark-game`, dedup via ON CONFLICT DO NOTHING) +
kcat orchestrate envelope → `experience-planner`, shared correlation_id; coverage
probe over open vonc items (FORCE=1 override). Beware kcat -P line-splitting
(travelling-docs trap): pass the envelope as a single line.

**T3.4 — run to convergence on `vonc-spark-game` → CP2.** Judge by the
experience-planner orchestration row; expect the council to sharpen D2's emitter
scoping and D1's minimal-real shape. On escalation: the escalation artifact IS
the round-boundary decision menu (D3) — surface it and stop the phase.

## Phase 4 — MVP build round (contract-first)

Inputs: the converged EXPERIENCE_PLAN. Build ONLY the MVP cut. Order below runs
data → surfaces so nothing ships pointing at absent data.

**T4.1 — data contract first.** Make `/data/provocations.json` conform to §3 of
the spec (incl. populating the reserved `arena` key). If the council put the
daily emitter IN the MVP: build it per
`PLAN_spark_provocation_pipeline.md` (clone the news-feed pipeline; model action
`render_news_section_action.go`; scheduled daily; publishes via git-adapter,
G20). If round-2: hand-refresh the sample to spec and the emitter goes top of
LATER.

**T4.2 — Arena: re-attach, then let the ladder fix it.** Deliberate sequence
(HANDOFF §3c "cheap win", done properly with T2.2 machinery):
1. Supersede the current `tool-arena-interface` PLAN with experience-derived
   criteria (incl. "fetches `/data/provocations.json`" as an interaction/journey
   criterion) via `write_doc_plan`.
2. One tx: `UPDATE content_components SET function='tool-arena' WHERE
   function='tool-arena-interface'` + `rekey_travelling_docs('tool',
   'tool-arena-interface','tool-arena')`; check `js_snippets.applies_to` and nav
   rows for the old key (the "stale nav phantom" class). `pages.name` is already
   `tool-arena` — this closes all three key links (G6).
3. Let `tool_acceptance_due` sweep it: expected FAILING verdict → `improve_tool`
   → tool-improver rebuilds against the criteria → re-verify GREEN. That is the
   pilot's showcase: the existing ladder pulling a real rebuild. Fallback after
   two strikes: recreate via the 086 bundle_recreation pattern with the spec as
   input.

**T4.3 — Provocation detail pages (D2).** Un-park `needs_page:provocation` ONLY
by superseding it with spec'd work: emit per-provocation static pages
(`/provocations/<slug>/index.html`) from the feed per the spec's data contract —
new small render action modelled on G19's emitter (or the emitter itself if
in-MVP), publishing via git-adapter. Point the archive loader's `href` fields at
the real slugs (the loader hydrates from the same feed, G22). The provocations
index page stays runtime-fill and `owned` (T2.1).

**T4.4 — Gauntlet minimal-real (D1).** Recreate `gauntlet-interface` per the
spec: a playable timed round against today's provocation (data from T4.1),
client-side scoring, real CTAs; STRIP every fabricated number; any surviving
quantitative copy must have an `evidence_base` fact or derive from the live feed
(T2.4). Rebuild path: tool-recreation/tool-generator with explicit spec so it is
born with PLAN + criteria (G4) and canonical identity (T2.2 root-cause fix);
mark the page `owned`.

**T4.5 — claims-clean gate → CP3.** `claimscan` on vonc export exit 0; `curl`
artifact checks per CP3; RUNNING_NOTES entry.

## Phase 5 — Journey acceptance + self-heal

Go work here can be developed in parallel with Phases 3–4 but goes live after
image roll A. Two images: browser-runner-adapter AND chassis (roll B).

**T5.1 — browser-runner journeys** (`internal/adapters/browserrunner/`,
separate image — check the untracked probe-test WIP first, §0):
- New check type `journey`: `steps[]` gains `goto` (path) alongside
  `fill|click|select`; a click that navigates waits for load and CONTINUES in
  the same page context (today each URL is an isolated fresh browser, G9);
  optional per-step `expect`, final `expect` required. Same profiles machinery;
  screenshots-on-failure (P3) inherited.
- Tier-2 side (`check_tool_acceptance.go`): journey criteria are anchor-checked
  only on their starting page; deeper steps are Tier-4's job (state this in the
  check, don't fake static coverage).
- Post-hydration dead-control assertion (closing T2.3's Tier-4 gap): after
  settle, interactive elements matching the criteria's container must not be
  `href="#"`/no-op.

**T5.2 — scopes + routing** (chassis, G10): add `ScopeData`/`ScopePlanGap`
consts (`run_checks_action.go:83`); journey criteria may carry `scope_hint:
'data'` on data-dependent steps (feed fetch fails / data-driven expect empty →
`data`; step whose page/control has no owner in the spec → `plan-gap`); bucket
them in `extractRunResults` (:357); route in `JudgeAcceptanceResultsAction`
(:467): both → **`needs_experience_replan`** items, `handler_agent=
'experience-planner'`, key `workItemKey("needs_experience_replan",
<experience_key>)`, spec carries the failing journey + verdict context. Two-
strike/unresolved semantics come free (G21). No new statuses → no dedup-index
change (§0).

**T5.3 — `experience_acceptance_due` sweep + acceptance variant**:
- New discovery check (registry init, G8): due = current `doc_plans`
  `subject_type='experience'` with criteria, no `source='experience-acceptance'`
  note in 7 days, no open `experience_acceptance_run` item → emit item,
  `handler_agent='experience-acceptance-agent'`.
- `experience-acceptance-agent` def = tool-acceptance workflow with
  `subject_type='experience'` in `load_docs` config and start-URLs taken from
  the criteria doc (small `RequestBrowserRunAction` tweak: read `urls` from
  criteria before the `pages.name` fallback — also unblocks any tool whose page
  name drifts, the G6 class).
- Enable via `jsonb_set` append AFTER images roll (image → seed → fire).

**T5.4 — proofs.** Real journeys GREEN on live vonc (CP4 first half). Induced
failure WITHOUT touching the live site: drill subject `vonc-spark-game-drill`
with one criteria step expecting an element the feed cannot satisfy → sweep →
verify `needs_experience_replan` created + claimed by the planner → cancel the
drill item + supersede the drill plan (T1.2). CP4 entry. Then leave the sweep
alone for two scheduled cycles → CP5.

## Phase 6 — Feature rounds

After CP5: loop B→D per round on the LATER list. Known candidates: the daily
emitter (if round-2), quiz "Get Your Full Report" empty href (product decision),
gauntlet/quiz `needs_rebuild` investigation, 096-style label unlock +
content pass on archetype grids (HANDOFF §1 leftovers), imagery/copy polish.
One decision menu per round boundary at most (D3).

## 7. PLAN §6 inventory → tasks

| PLAN §6 piece | Task(s) |
|---|---|
| experience-planner + EXPERIENCE_PLAN schema | T1.1, T3.1, T3.2 |
| Challenge-council workflow | T3.1 (same workflow), T3.3 |
| Tier-4 journey runs | T5.1 |
| `needs_experience_replan` + data/plan-gap scopes | T5.2 |
| Page-ownership rebuild guard | T2.1 |
| Doc re-keying on rename | T2.2 |
| Dead-control Tier-2 check | T2.3 (+T5.1 Tier-4 variant) |
| (implicit) claims lane for vonc | T2.4 |
| (implicit, D2) detail-page emitter | T4.1/T4.3 |

## 8a. Execution state (kept current; newest change last)

| Task | State | Evidence |
|---|---|---|
| T1.1 subject_type migration | **DONE 2026-07-17** — landed as **163** (162 was taken mid-flight by another session's toolgen-rerender tail) | `163_doc_subjects_experience.sql` applied+ledgered; commit `378054bad`; probe insert/delete verified |
| T2.1 ownership marker | **DONE (code) 2026-07-17** — migration **164** applied (38 pages owned: 36 tool + vonc provocations-index/provocation); reconcile emits `owned_page_review`, save_page_sections hard-refuses | commit `fb89f1071`; Go awaits image roll A |
| T2.2 rename re-key | **DONE (code) 2026-07-17** — `RekeyTravellingDocs` + `rename_tool_identity` action + CanonicalisePage at tool birth | commit `aabd38161`; awaits image roll A |
| T2.3 dead-control check | **DONE (code) 2026-07-17** — helper+tests, Tier-2 shell check, `dead_controls` discovery check; enable SQL **165 written, NOT applied** (image-first) | commit `f2824a713` |
| T2.4 claims lane | **DONE 2026-07-17** — migration **166** applied (evidence_base: facts empty, 9 banned patterns); claimscan baseline = **14 findings** incl. 3 previously unknown ('14,203 Happy Customers' about+index, '10K+ Players Scored' index, mangled about stat labels) → T4 strip list | commit `c437682a6`; baseline output in session log |
| T2.5 image roll A + CP1 | **IN PROGRESS** — target tag v1.0.1132 (cluster at 1130, 1131 possibly claimed) | — |
| T3–T5 | not started | — |

## 8. Standing landmines while executing

G22 (runtime-fill shells: leave blank fields alone); G23 (unresolved triplets are
expected; they duplicate per discovery pass); G13 until T2.1 ships (park
`needs_page:provocation` manually after every reconcile); browser-runner and
chassis are SEPARATE images — a chassis roll does not ship T5.1; the
council/fix-proposer pattern is live but was "not yet exercised end-to-end" as of
2026-07-17 — treat the first council run as also a shakedown of that pattern;
`claimscan`'s baseline on vonc will list the gauntlet numbers until T4.4 — that
is signal, not noise.
