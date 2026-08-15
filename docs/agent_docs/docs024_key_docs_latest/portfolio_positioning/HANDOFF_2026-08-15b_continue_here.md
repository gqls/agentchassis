# HANDOFF — B3c + B3d live and verified; B3e/B3f then Phase C pilot next — 2026-08-15 evening, continue here

Supersedes `HANDOFF_2026-08-15_continue_here.md` (correct on its own history; this file
carries everything a fresh chat needs). Owner rulings in force are unchanged from that
file's §0 (P9 six decisions, pilot = remortgagecalculator.uk M4, build order M→B→I,
B8/B9/I10 HOLD, bug 270 hands-off, copy-voice work lives in session "copy quality two
stage" `79d969f9-…`).

## 1. What is DONE and PROVEN (this session, all verified at the artefact)

**B4 — supervised researcher runs: CLOSED** (third session). All three kinds proven:
register at close 25 active entities / 33 cited claims (mortgage 2/3, savings 13/15,
health 10/15). Migrations 423/424/428 shaped the researcher config. Evidence: NOTES.

**B3c — publish leg kind-aware: DONE, LIVE, COUNCIL-APPROVED.** Migration
`sql_for_agents/429_directory_publish_trigger_kind_aware_fan_out.sql` (+ROLLBACK),
commit `0af2c21f9`:
- `model-directory-trigger` finds due (site, kind) PAIRS across all six kinds — per-kind
  opt-in (`content_features.<spec_key>`), per-kind deployed component, per-kind
  publishable claims (mirrors QueryDirectoryEntries: active entities, is_current+found).
  `ORDER BY random() LIMIT 12` replaced `ORDER BY domain LIMIT 5` (deterministic
  starvation). The kind→spec_key→components→commit-message mapping is a SQL VALUES list
  in LOCKSTEP with Go's `directoryPublishProfiles` — **adding a kind is now SEVEN
  places** (LANDMINES entry + DIR-001 both updated; verifier armed).
- `model-directory-publisher` is ONE render→commit pair; `kind` REQUIRED by its
  input_contract (call_agent's ValidateInputContract fails a kind-less call loudly —
  the 2026-07-26 silent-model-default class, closed upstream of the action). Per-kind
  commit messages via `commit_message_field`; 427's no-error_step posture kept.
- First run verified: 3 publisher orchestrations, per-kind entity counts 44/40/8
  (DIFFER — the 07-26 collapse check), correct per-kind files, served JSON fresh at
  each run's completion second, finance kinds correctly self-gated to zero rows.
- **Council corr `a7c99b84-…`: APPROVED round 1** (18:24Z, 5 advisories none-high, two
  dispositioned with live evidence in NOTES — both were sketch-visibility artefacts).
  098 credits commit `0af2c21f9` automatically.
- Seed `model_directory_pipeline/SEED_directory_publish_trigger.sql` synced to live;
  FINDING_2026-08-10 banner updated; cross-lane notice
  `model_directory_pipeline/CONTRIB_2026-08-15_b3c_publish_leg_kind_aware.md` (includes
  the rerender_queued=0 observation left with that lane).

**B3d — evaluate_directory_features wired: DONE, LIVE.** Migration
`sql_for_agents/432_wire_evaluate_directory_features_b3d.sql` (+ surgical-inverse
ROLLBACK — deliberately NOT restore-from-backup; see file header):
- improvement-loop: `enrich_directory_features` between `enrich_news_feed` and
  `load_audit_state`, news edges re-pointed on BOTH success/error paths (291's
  reach-the-audit property preserved). site_id = `site_record.site_id`.
- domain-research-classifier: same step between `write_classification_spec` and
  `write_content_direction_spec` (greenfield flag at plan time). site_id =
  `input_data.site_id`.
- No-match/no-spec = NO WRITE, so the fleet outside the three finance verticals is
  untouched by construction.
- **Council corr `47785bb5-ca66-4aed-819f-2bd29277b80d`: SUBMITTED, VERDICT UNREAD —
  read it first thing**:
  `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
   correlation_id='47785bb5-ca66-4aed-819f-2bd29277b80d' AND kind='council_report' ORDER BY 1;`
  (Report body: same table, column `body`. If REVISE: fix and resubmit with
  `RESUBMIT_CORR=47785bb5-…`.)

## 2. Fresh findings a new session must know

- **`improvement-sweep` is DISABLED fleet-wide** (`enabled=f`, last fired 2026-08-14
  16:34Z, interval 900s) — so B3d's improvement-loop consumer is wired but UNDRIVEN.
  Not ours to re-enable unilaterally: find who disabled it (git/NOTES grep or ask the
  owner) before flipping. Until then, absence of enrichment rows is NOT a 432 failure.
  The classifier consumer self-proves on the Phase C pilot build.
- **Binary is v1.0.1303** (rolled 18:45Z by another session, both replicas). It
  postdates RFC_029's strict-marker commit (`1806371ef`, 14:07Z), so `kind!` strict
  mappings are PROBABLY available now — [UNVERIFIED at the pod; check build provenance
  before relying on it]. 429 deliberately used plain references + contract enforcement
  and needs no change.
- **Verify-query for future kind-aware publish runs** (one orchestration per pair now):
  `SELECT collected_data->'input_data'->>'kind', collected_data->'directory_render_result'->>'entity_count'
   FROM orchestration_states WHERE owner_agent_type='model-directory-publisher' ORDER BY created_at DESC;`
  Per-kind counts MUST differ; identical counts = the 07-26 defect back.
- **Council submission schema gotchas** (cost three rejected submits): `.plan.summary`
  required; operations are modify|add|remove|config_change ('create' refused);
  `.plan.risks` is a STRING not an array. FORCE=1 needed for config migrations under
  docs/ (411/429/432 precedent).
- `orchestration_states` payload searches (`collected_data->'input_data'->>…`) can run
  >120s — query `diagnosis_artifacts` by `correlation_id` instead for council verdicts
  (indexed, instant; columns: `body`, `metadata`, NOT `content`).

## 3. Next actions, in order

1. **Read the 432 verdict** (corr above). Disposition advisories in NOTES; handle REVISE
   if it comes.
2. **B3e** — build-site-planner prompt rule (206's replace()-idiom, snapshot first,
   anchors pre-checked unique) + the six directory component names into the planner
   vocabulary. Config-only. The six functions: model-directory, adoption-tracker,
   protocol-tracker, mortgage-lender-directory, savings-provider-directory,
   health-insurer-directory (+ their -listing twins).
3. **B3f** — enable the 6 directory checks AND the 5 Phase-A structural checks on
   completeness-discovery-agent (194/215 jsonb_set pattern). Binary precondition
   satisfied (v1.0.1303 ⊇ v1.0.1301). Directory checks self-gate on publishable data,
   and B4 populated all three finance kinds, so nothing blocks this now.
4. **Phase C pilot** — remortgagecalculator.uk (M4) end-to-end: mission-file from the
   register entry, pre-seeded specs, marker sentence, dispatch via
   `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh`, cost
   baseline from `llm_call_log`/`assets`, owner sign-off before Phase E. The B3d
   classifier wiring means this build should carry
   `content_features.mortgage_lender_directory` at plan time — CHECK IT on the built
   site's classification spec; that is 432's live proof.
5. Then Phase D decisions / Phase E waves per `PLAN_2026-08-12_fleet_buildout.md`.

## 4. Standing cautions (carried forward)

- The mortgagecalculator.co.uk voice review is ACTIVE IN ANOTHER THREAD (session "copy
  quality two stage") — do not duplicate from this lane.
- Bug 270 owned elsewhere — cite, hands off.
- `git stash` forbidden (hook-blocked); pathspec commits; forward-only; re-run
  `git status` before acting on it (this session watched migration numbers 430/431 get
  taken within 30 minutes).
- Migration files: next free number was 433 at session end — RE-CHECK, the queue moves.
- Landmine verifier was armed for the new VALUES-lockstep entry
  (`landmines-verify-dispatch.sh` run this session); its verdict lands in doc_notes
  (categories ? 'landmine-verification') — no action unless it objects.

## 5. Files of record

This dir: `PLAN_2026-08-12_fleet_buildout.md` (phase map) ·
`SUMMARY_2026-08-15_guardrails_live_directories_built.md` +
`SUMMARY_2026-08-15b_first_supervised_runs.md` (milestones) ·
`NOTES_portfolio_positioning.md` (evidence, newest at bottom — the two 2026-08-15
fourth-session entries cover B3c/B3d in full) · `README_where_we_are.md` ·
`COUNCIL_SUBMISSION_429_…json` / `COUNCIL_SUBMISSION_432_…json`.
Migrations: `sql_for_agents/429_…` + `432_…` (+ROLLBACKs), all applied.
Register: `docs026_concept_register/register/directory-pipeline.md` (DIR-001, updated).
Cross-lane: `model_directory_pipeline/CONTRIB_2026-08-15_b3c_publish_leg_kind_aware.md`.
Commits: `0af2c21f9` (B3c, council-approved corr `a7c99b84`), B3d commit follows this
file (corr `47785bb5`).
