# Context for the next chat — what to attach

Two bodies of work run through this project: **idea.uk** (the small paid product we've been
building — its files live in *this chat's outputs* and are NOT in your chassis project, so
you must download and add them) and the **chassis platform** (the agent site-builder — those
files are already in your project). Attach the subset for the task, not everything; the
context window fills fast.

---

## Start here (attach first, whichever track)

- **HANDOFF.md** — current state + the exact next steps.
- **running_notes.md** — the full cross-session journal (read the tail first).

---

## Track A — continuing idea.uk (Stripe go-live / finish the email test)

idea.uk is one self-contained Go binary with file-based persistence. **No SQL schema or
table-content files are needed** — persistence is `store.go`.

### The live code — download from `outputs/idea-go/` (these are not in your chassis project)
- `service.go` — App/Config/NewApp, all HTTP handlers, the `a.page()` wrapper, `writeHTML`,
  the mailer incl. `smtpSend` (the 465 implicit-TLS path), policy-page constants
- `engine.go`, `prompts.go` — the ideation method
- `audience_check.go` — the free taster
- `store.go` — persistence (idea.uk's own; no external DB)
- `billing.go` — Stripe / Fake provider
- `main.go` — config loading
- `service_test.go` — the test
- `page.html` — the embedded landing page (must sit beside the .go files for `go:embed`)
- `go.mod`
- `deploy/setup.sh`, `deploy/idea.env.example`, `deploy/README.md`

### Docs — download from `outputs/` — essential
- `idea_uk_architecture_and_deployment.md` — architecture, hosting, deploy, email live-state
- `EMAIL_identity_in_site_spec.md` — the email design (per-site forwarders)
- `LIABILITY_AND_TERMS.md` — the legal posture (solicitor review still pending)
- `016_debugging_guide_v2_32.md` — chassis pipeline guide; §11 = idea.uk page-serving + deploy
- `PLAN_stripe_billing_integration.md` — for the next likely task: taking real money

### Docs — supporting (attach if you want the full backstory)
- `idea_uk_method_v0.md`, `idea_method_prompt.md` — the method in detail
- `RUNBOOK_idea_uk.md`, `DEVELOPMENT_RUNBOOK.md` — the phased plan (A–D)
- `CONSOLIDATION_where_it_all_fits.md`, `PARALLEL_engine_deployment_and_layer5.md`,
  `PERSISTENCE_design.md`, `VM_LAUNCH_PLAN.md`, `PLAN_idea_uk.md`
- `idea_uk_testrun_v2.md` — a sample engine run
- `terms_preview.html`, `refund_policy_preview.html`, `privacy_preview.html` — preview copies
  (the same content is embedded in `service.go`)
- `idea_uk_fakedoor.html`, `leopardess_uk_index.html` — the public pages

### Skip (superseded / noise)
- `016_debugging_guide_v2_30.md`, `_v2_31.md` (keep only **v2_32**)
- `idea_uk_testrun_v0.md` (keep v2)
- `idea_uk_open_discussion.md` (only if you want the early discussion)
- `PLAN_isolated_chat_environment.md`, `PLAN_simple_paid_multidomain_chat.md`,
  `FOCUS_site_chatbot_edge_worker_and_context_pack.md` — earlier explorations, not idea.uk

---

## Track B — chassis platform work (the agent site-builder)

These are already in your project (the mounted set). Attach the **index plus the area you're
working on**, not the whole tree.

### Master index (attach first)
- `000_documentation_index.md`

### Schemas (.sql) — for the "check the schema before SQL" rule
- `002_intake_orchestrator.sql`, `003_site_classifier.sql`, `018_briefing_questionnaire.sql`,
  `agent_definition_types.sql`, `006_unify_prices_schema.sql`, `019_model_lifecycle_schema.sql`,
  `021_model_swap_and_rollback.sql`, `101_switch_to_haiku.sql`
- `flywheel_B_step0_checks.sql`, `flywheel_B_step1b_pgvector_smoke.sql`,
  `flywheel_D_step0_discovery.sql`, `flywheel_D_target_selection.sql`
- `schemas_all`, `schemas_some` — full schema dumps (the quickest schema reference)

### Table content / seed data
- `initial_messages__without_current_ids_`
- `initial_vet_practice_check_message`

### Core orchestration (Go)
- `registry.go` (+ `registry_go.txt`), `coordinator.go`, `queryresolve.go`,
  `action_inputs.go`, `spawn_actions.go`, `safe_unmarshal.go`, `timeout_helpers.go`,
  `sql_helpers.go`
- `production_agent-chassis-actions-current_context.txt` — current actions context

### Site builder / content (Go)
- Site actions: `v3_site_actions.go`, `write_site_plan_action.go`, `site_db_actions.go`,
  `site_spec_actions.go`
- Content/extract: `content_search.go`, `deep_search.go`, `unified_extractor.go`,
  `file_extractor.go`, `data_helpers.go`, `format_content_direction.go`, `nav_labels.go`,
  `page_canonical.go`, `page_role_validator.go`, `web_architecture_helpers.go`,
  `imagery_helpers.go`, `debug_collected_data.go`, `duplicate_logger.go`
- Integrity checks: the **`check_*.go` set (~24 files)** — they travel together; e.g.
  `check_integrity.go`, `check_broken_nav_links.go`, `check_orphan_pages.go`,
  `check_unlinked_components.go`, `check_unresolved_sections.go`, `check_missing_structure.go`,
  `check_tool_health.go`, `check_undeployed_assets.go`, … (attach the whole set when working
  on the integrity/triage pass)

### Infra / config (yaml + make)
- `rbac.yaml`, `service.yaml`, `thunder-adapter.yaml`, `kustomization.yaml`,
  `deployment.yaml`, `makefile.txt`

### Docs spine (numbered) — attach the ones for your area
- Foundations: `001_development_guide`, `002_system_architecture`,
  `002_README-flywheel...evaluation_pipeline`, `003_contracts_and_standards`,
  `004_improvement_loop`, `005_tool_pipeline`
- Pipeline/infra: `007_adoption_pipeline_v4` (+ `.patch`), `008_vet_med_pricing_pipeline`,
  `009_model_infrastructure`, `010_scheduler_and_tasks`, `011_database_and_infrastructure`,
  `012_admin_dashboard`, `013_content_governance`, `014_site_snapshots_and_revert`,
  `015_batch_processing_architecture_v2`
- Content/tooling/design: `017_companies_house_enrichment`, `018_canine_biology`,
  `019_tool_library`, `020_tool_lifecycle`, `021_site_spec_and_classifier`,
  `022_dynamic_applications`, `023_llm_quality_testing`, `024_link_management_v2`,
  `025_palette_layout_typography_migration`, `026_component_regeneration_flow`,
  `027_design_and_site_planner_v2`, `029_site_plan_and_reconciler`,
  `030_phase1_plan_and_reconciler`
- Platform/ops: `028_platform_mission_and_pipeline_direction`, `031_locks` (+ variants),
  `032_storage_architecture_and_credentials`, `033_thunder_adapter_design`,
  `106_claude_anthropic_skill`

### Debug / triage docs
- `016_debugging_guide.md` (+ `_v2_27`, `_addendum...`),
  `105_dispatch-pipeline-failures-report_v4.md`, `HANDOFF-pipeline-triage-april-2026.md`
- The dated `HANDOFF_*`, `FOCUS_*`, `ANALYSIS_*`, `ASSESSMENT_*`, `ARCHITECTURAL_TENSIONS`,
  `STATUS_*`, `TODO_*`, `P1/P2/P3` notes — attach only the ones touching your task

---

## Rule of thumb

- idea.uk task → the `idea-go/` code + HANDOFF + running_notes + architecture + email +
  Stripe plan. That's enough.
- chassis task → `000_documentation_index.md` + the schemas/Go/docs for that one area.
- Add more only when a specific question needs it. Carrying the whole tree wastes context and
  makes answers worse, not better.

---

## To trim this to the exact set

Tell me which the next chat is — **(a) idea.uk go-live** (Stripe + finishing the email test)
or **(b) chassis platform work** (and which area) — and I'll cut this down to the minimal
file list for that, so you're not carrying noise.
