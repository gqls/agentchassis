# EXTRACTION U04 — docs024_key_docs_latest/idea.uk (idea.uk product + chassis site-build/design thread)
Extracted 2026-07-13. Files in scope: 289. Concepts found: 62.

Unit character: this directory carries TWO intertwined streams. (1) **idea.uk the product** — a
standalone, LIVE, earning £29 "verified AI product idea" report service (Go binary on a Hetzner VM,
Stripe, SES email), built outside the chassis but designed to fold into it. (2) **The chassis
site-build thread** — rebuilding idea.uk's front site through the agentchassis pipeline, which used
idea.uk as the test case that exposed a series of framework-level design gaps (structured
design_intent, scheme-aware layout matching, the section-contrast / scheme-to-components gap,
coordinator result extraction). Per the taxonomy rule, platform concepts observed on idea.uk are
extracted as platform concepts with the idea.uk docs as provenance.

Family notes (for the audit): `running_notes(45..63)` is an append-only journal — (63) verified a
superset by prefix/content checks, earlier numbers carry no unique content (family-delta).
`DESIGN_PIPELINE_two_track_investigation(2)..(5)` are byte-identical (md5); base and (1) differ only
by the later-superseded "decision options" section. `TODO_chassis_and_idea_uk(1).md` ==
`README_002_TODO_chassis_and_idea_uk(2).md` (md5-identical). `running_notes_2(5)==(6)`.
`golang_files/*.orig*` are code-version backups of the live files (headers of live files scanned;
origs are family-delta). `nginx/_iso/*` are training-checkpoint isolation-test artefacts
(binaries/generated). Secrets note: `golang_files/README_email.md` contains what appears to be a
real AWS console password and account id pasted verbatim — flagged for the operator, not extracted.

## Coverage
| file | treatment |
|---|---|
| 001_component_flow.md | full |
| 016_debugging_guide_v2_32(1).md | family-latest |
| 019_pcw_prompt_item_fields.sql | full |
| BUGS_idea_uk(2).md | family-delta |
| BUGS_idea_uk(3).md | family-delta |
| BUGS_idea_uk(4).md | family-latest |
| BUGS_idea_uk.md | family-delta |
| BUNDLE_1_idea_uk_golive.md | header-scan |
| BUNDLE_2_chassis_idea_engine_workflow.md | header-scan |
| CONSOLIDATION_where_it_all_fits.md | full |
| CONTEXT_FOR_NEXT_CHAT.md | header-scan |
| CONTEXT_PACK_idea_uk_golive.md | header-scan |
| DESIGN_PIPELINE_two_track_investigation(1).md | family-delta |
| DESIGN_PIPELINE_two_track_investigation(2).md | family-delta |
| DESIGN_PIPELINE_two_track_investigation(3).md | family-delta |
| DESIGN_PIPELINE_two_track_investigation(4).md | family-delta |
| DESIGN_PIPELINE_two_track_investigation(5).md | family-latest |
| DESIGN_PIPELINE_two_track_investigation.md | family-delta |
| DEVELOPMENT_RUNBOOK(1).md | family-delta |
| DEVELOPMENT_RUNBOOK(2).md | family-delta |
| DEVELOPMENT_RUNBOOK(3).md | family-latest |
| DEVELOPMENT_RUNBOOK.md | family-delta |
| EMAIL_identity_in_site_spec(2).md | family-delta |
| EMAIL_identity_in_site_spec(3).md | family-delta |
| EMAIL_identity_in_site_spec(5).md | family-latest |
| EMAIL_identity_in_site_spec.md | family-delta |
| HANDOFF(11).md | family-delta |
| HANDOFF(12).md | family-delta |
| HANDOFF(13).md | family-latest |
| HANDOFF(5).md | family-delta |
| HANDOFF(7).md | family-delta |
| HANDOFF(8).md | family-delta |
| HANDOFF(9).md | family-delta |
| HANDOFF.md | family-delta |
| HANDOFF_chassis_site.md | full |
| HANDOFF_scheme_to_components(1).md | full |
| KEY_DOC_idea_method_prompt.md | header-scan |
| KEY_DOC_idea_uk_mission.txt | full |
| LIABILITY_AND_TERMS(2).md | family-latest |
| LIABILITY_AND_TERMS.md | family-delta |
| PARALLEL_engine_deployment_and_layer5.md | full |
| PLAN_idea_uk(1).md | family-delta |
| PLAN_idea_uk(2).md | family-delta |
| PLAN_idea_uk(3).md | family-latest |
| PLAN_idea_uk.md | family-delta |
| PLAN_stripe_billing_integration(1).md | family-delta |
| PLAN_stripe_billing_integration(3).md | family-latest |
| README_001_todo_list.md | full |
| README_002_TODO_chassis_and_idea_uk(2).md | family-delta |
| README_002_TODO_chassis_and_idea_uk.md | family-delta |
| README_assemble_bundle_idea_missing_sections.md | full |
| README_claude_conversation.md | full |
| README_stripe.md | full |
| REPORT_scheme_does_not_reach_components.md | full |
| RUNBOOK_idea_uk(2).md | family-delta |
| RUNBOOK_idea_uk(2)golang.md | family-delta |
| RUNBOOK_idea_uk(3).md | family-delta |
| RUNBOOK_idea_uk(4).md | family-delta |
| RUNBOOK_idea_uk(5).md | family-delta |
| RUNBOOK_idea_uk(6).md | family-delta |
| RUNBOOK_idea_uk(7).md | family-delta |
| RUNBOOK_idea_uk(8).md | family-delta |
| RUNBOOK_idea_uk(9).md | family-latest |
| RUNBOOK_idea_uk.md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(1).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(10).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(11).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(12).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(13).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(14).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(15).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(16).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(17).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(18).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(19).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(2).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(20).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(21).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(22).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(23).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(24).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md | family-latest |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(3).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(4).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(5).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(6).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(7).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(8).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy(9).md | family-delta |
| RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md | family-delta |
| RUNBOOK_idea_uk_vm_cutover.md | full |
| TODO_chassis_and_idea_uk(1).md | full |
| UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md | family-latest |
| UPDATE_FOCUS_design_adoption_workplan_2026-06-19.md | family-delta |
| docubundle_idea_golive/package_idea_uk_golive(1).sh | family-delta |
| docubundle_idea_golive/package_idea_uk_golive(2).sh | header-scan |
| docubundle_idea_golive/package_idea_uk_golive.sh | family-delta |
| docubundle_idea_golive/package_module/output_contexts/idea-uk-golive_context.txt | skipped-generated |
| docubundle_idea_within_chassis/package_chassis_idea_engine(1).sh | family-delta |
| docubundle_idea_within_chassis/package_chassis_idea_engine(2).sh | family-delta |
| docubundle_idea_within_chassis/package_chassis_idea_engine(3).sh | header-scan |
| docubundle_idea_within_chassis/package_chassis_idea_engine.sh | family-delta |
| docubundle_idea_within_chassis/package_module/output_contexts/agent_definition_types.sql | header-scan |
| docubundle_idea_within_chassis/package_module/output_contexts/production_chassis-idea-engine_context.txt | skipped-generated |
| golang_files/001trigger.sh | header-scan |
| golang_files/Dockerfile | header-scan |
| golang_files/README_confirm_order.md | full |
| golang_files/README_email.md | header-scan |
| golang_files/README_setup_SETUP.md | full |
| golang_files/audience_check(2).go.orig2 | family-delta |
| golang_files/audience_check(3).go.orig3 | family-delta |
| golang_files/audience_check.go | header-scan |
| golang_files/audience_check.go.orig1 | family-delta |
| golang_files/billing.go | header-scan |
| golang_files/deploy/idea.env.example.orig1 | family-delta |
| golang_files/engine(6).go.orig6 | family-delta |
| golang_files/engine(7).go.orig7 | family-delta |
| golang_files/engine(8).go.orig8 | family-delta |
| golang_files/engine(9).go.orig9 | family-delta |
| golang_files/engine.go | header-scan |
| golang_files/engine.go.orig1 | family-delta |
| golang_files/engine.go.orig2 | family-delta |
| golang_files/engine.go.orig3 | family-delta |
| golang_files/engine.go.orig4 | family-delta |
| golang_files/engine.go.orig5 | family-delta |
| golang_files/engine.goorig5 | family-delta |
| golang_files/go.mod | header-scan |
| golang_files/idea | skipped-binary |
| golang_files/idea.env.example | header-scan |
| golang_files/main(12).gof.orig12 | family-delta |
| golang_files/main(18).go.orig18 | family-delta |
| golang_files/main.go | header-scan |
| golang_files/page(7).html.orig5 | family-delta |
| golang_files/page(8).html | skipped-generated |
| golang_files/page.html | skipped-generated |
| golang_files/page.html.orig1 | family-delta |
| golang_files/page.html.orig2 | family-delta |
| golang_files/page.html.orig3 | family-delta |
| golang_files/page.html.orig4 | family-delta |
| golang_files/privacy_preview.html | skipped-generated |
| golang_files/privacy_preview.html.orig1 | family-delta |
| golang_files/prompts(2).go.orig2 | family-delta |
| golang_files/prompts(3).go.orig3 | family-delta |
| golang_files/prompts(4).go.orig4 | family-delta |
| golang_files/prompts.go | header-scan |
| golang_files/prompts.go.orig1 | family-delta |
| golang_files/service(10).go.orig10 | family-delta |
| golang_files/service(12).go.orig12 | family-delta |
| golang_files/service(16).go.orig16 | family-delta |
| golang_files/service(17).go.orig17 | family-delta |
| golang_files/service(18).go.orig18 | family-delta |
| golang_files/service(19).go.orig19 | family-delta |
| golang_files/service(21).go.orig21 | family-delta |
| golang_files/service(22).go.orig22 | family-delta |
| golang_files/service(8).go.orig8 | family-delta |
| golang_files/service(9).go.orig9 | family-delta |
| golang_files/service.go | header-scan |
| golang_files/service.go.orig1 | family-delta |
| golang_files/service.go.orig2 | family-delta |
| golang_files/service.go.orig3 | family-delta |
| golang_files/service.go.orig4 | family-delta |
| golang_files/service.go.orig5 | family-delta |
| golang_files/service.go.orig6 | family-delta |
| golang_files/service.go.orig7 | family-delta |
| golang_files/service_test(1).go.orig1 | family-delta |
| golang_files/service_test(2).go.orig2 | family-delta |
| golang_files/service_test(3).go.orig3 | family-delta |
| golang_files/service_test(4).go.orig4 | family-delta |
| golang_files/service_test(5).go.orig5 | family-delta |
| golang_files/service_test(6).go.orig6 | family-delta |
| golang_files/service_test(7).go.orig7 | family-delta |
| golang_files/service_test.go | header-scan |
| golang_files/service_test.go.orig0 | family-delta |
| golang_files/setup.sh | header-scan |
| golang_files/store(3).go.orig3 | family-delta |
| golang_files/store.go | header-scan |
| golang_files/terms_preview.html | skipped-generated |
| idea_uk_architecture_and_deployment(1).md | family-delta |
| idea_uk_architecture_and_deployment(2).md | family-delta |
| idea_uk_architecture_and_deployment(3).md | family-delta |
| idea_uk_architecture_and_deployment(4).md | family-delta |
| idea_uk_architecture_and_deployment(6).md | family-latest |
| idea_uk_architecture_and_deployment.md | family-delta |
| idea_uk_fakedoor(1).html | family-delta |
| idea_uk_fakedoor(2).html | family-delta |
| idea_uk_fakedoor(3).html | family-delta |
| idea_uk_fakedoor(4).html | family-delta |
| idea_uk_fakedoor(5).html | family-delta |
| idea_uk_fakedoor(7).html | family-delta |
| idea_uk_fakedoor(8).html | family-delta |
| idea_uk_fakedoor(9).html | header-scan |
| idea_uk_fakedoor.html | family-delta |
| idea_uk_method_v0(1).md | family-delta |
| idea_uk_method_v0(2).md | family-delta |
| idea_uk_method_v0(3).md | family-latest |
| idea_uk_method_v0.md | family-delta |
| idea_uk_open_discussion.md | full |
| idea_uk_testrun_v0(1).md | family-delta |
| idea_uk_testrun_v0(2).md | header-scan |
| idea_uk_testrun_v0.md | family-delta |
| idea_uk_testrun_v2.md | header-scan |
| leopardess_uk_index.html | header-scan |
| migration_domain_research_classifier_structured_design_intent.sql | full |
| migration_layouts_scheme_and_light_tool_portal.sql | full |
| nginx/PERSISTENCE_design(1).md | full |
| nginx/PERSISTENCE_design.md | family-delta |
| nginx/PLAN_checkpoint_and_artefact_upload_b2(1).md | full |
| nginx/README.md | header-scan |
| nginx/README.md.orig1 | family-delta |
| nginx/README_get_b2_details.md | full |
| nginx/README_setup_box.md | full |
| nginx/VM_LAUNCH_PLAN.md | full |
| nginx/_iso/adapter.tar.gz | skipped-binary |
| nginx/_iso/adapter_out/adapter_config.json | skipped-generated |
| nginx/_iso/adapter_out/adapter_model.safetensors | skipped-binary |
| nginx/_iso/adapter_out/checkpoints/checkpoint-50/optimizer.pt | skipped-binary |
| nginx/_iso/adapter_out/manifest.json | skipped-generated |
| nginx/_iso/ckpt-0.tar.gz | skipped-binary |
| nginx/_iso/ckpt_src/checkpoint-50/config.json | skipped-generated |
| nginx/_iso/ckpt_src/checkpoint-50/optimizer.pt | skipped-binary |
| nginx/_iso/ckpt_src/checkpoint-50/trainer_state.json | skipped-generated |
| nginx/_iso/resume_dst/checkpoint-50/optimizer.pt | skipped-binary |
| nginx/isolation_test_phase_a.py | header-scan |
| nginx/setup.sh.orig1 | family-delta |
| nginx/setup.sh.orig2 | family-delta |
| nginx/setup.sh.orig3 | header-scan |
| old_golang_files/engine.go | family-delta |
| one_sentence_description.md | full |
| page(6).html | skipped-generated |
| python_files/Dockerfile | family-delta |
| python_files/Dockerfile(1) | family-delta |
| python_files/env(1).example | family-delta |
| python_files/env.example | family-delta |
| python_files/idea_method_prompt.md | header-scan |
| python_files/idea_method_runner(1).py | family-delta |
| python_files/idea_method_runner(2).py | family-delta |
| python_files/idea_method_runner(3).py | family-delta |
| python_files/idea_method_runner(4).py | family-delta |
| python_files/idea_method_runner(5).py | family-delta |
| python_files/idea_method_runner(6).py | family-delta |
| python_files/idea_method_runner.py | header-scan |
| python_files/idea_service(1).py | family-delta |
| python_files/idea_service(2).py | family-delta |
| python_files/idea_service(3).py | family-delta |
| python_files/idea_service.py | header-scan |
| python_files/test_idea_flow(1).py | family-delta |
| python_files/test_idea_flow.py | header-scan |
| reresolve_idea_uk_01_backup_and_inspect.sql | header-scan |
| reresolve_idea_uk_01b_inspect_only.sql | header-scan |
| reresolve_idea_uk_02_detach_and_clear(1).sql | header-scan |
| reresolve_idea_uk_02_detach_and_clear.sql | header-scan |
| reresolve_idea_uk_02b_check_state.sql | header-scan |
| reresolve_idea_uk_03_trigger(1).sh | header-scan |
| reresolve_idea_uk_03_trigger.sh | header-scan |
| reresolve_idea_uk_04_verify.sql | header-scan |
| reresolve_idea_uk_05_render.sh | header-scan |
| resolveLayoutByTags_weighted.go.patch.txt | header-scan |
| running_notes(45).md | family-delta |
| running_notes(46).md | family-delta |
| running_notes(47).md | family-delta |
| running_notes(48).md | family-delta |
| running_notes(49).md | family-delta |
| running_notes(50).md | family-delta |
| running_notes(51).md | family-delta |
| running_notes(52).md | family-delta |
| running_notes(53).md | family-delta |
| running_notes(54).md | family-delta |
| running_notes(55).md | family-delta |
| running_notes(56).md | family-delta |
| running_notes(57).md | family-delta |
| running_notes(58).md | family-delta |
| running_notes(59).md | family-delta |
| running_notes(60).md | family-delta |
| running_notes(61).md | family-delta |
| running_notes(62).md | family-delta |
| running_notes(63).md | family-latest |
| running_notes_2(1).md | family-delta |
| running_notes_2(2).md | family-delta |
| running_notes_2(3).md | family-delta |
| running_notes_2(4).md | family-delta |
| running_notes_2(5).md | family-delta |
| running_notes_2(6).md | family-latest |
| running_notes_2.md | family-delta |
| running_notes_checkpoint_tt.md | full |
| sample_paylink_email.html | skipped-generated |
| sample_report_email(1).html | skipped-generated |
| sample_report_email(2).html | skipped-generated |
| sample_report_email(3).html | skipped-generated |
| sample_report_email.html | skipped-generated |

## Concepts

### Five-layer platform stack (chassis → idea engine → idea.uk → vertical tools → tool-rich sites → VM backend deploy)
- **category:** business-strategy
- **status-signal:** partial
- **status-evidence:** "Where it all fits" map dated 2026-06-04: Layer 0 EXISTS, Layer 1 BUILT, Layers 2–3 IN PROGRESS, Layers 4–5 FUTURE ("Thunder adapter is the seed").
- **what:** A consolidation model presenting the whole enterprise as one stack: the chassis builds sites (L0); the idea engine decides what's worth building (L1); idea.uk sells that externally (L2); recommended tools get built for real, chassis-native (L3); the engine becomes a planning input so any domain gets a tool-rich site (L4 — "the original problem statement"); and automated backend deployment onto VMs closes the last gap (L5). Each layer is a customer of the one below.
- **sources:** idea.uk/CONSOLIDATION_where_it_all_fits.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md
- **relations:** Layer-5 persistent-service wrapper; SFI26 Diff Alerts; chassis-native idea engine (Phase D); Thunder adapter (docs033/035).
- **verify-later:** existence of any service-deployer agent; site_plan aspects carrying blocked/planned tool items; thunder-adapter actions.

### Differentiator framework — payable idea = hard-to-reproduce asset × current AI capability
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** PLAN_idea_uk §3/§5 (framework in use); method encodes it as the core principle; testruns v0/v2 applied it across 8 domain runs.
- **what:** The AI model is never the differentiator (everyone has the same models); the defensible unit is an asset × capability aimed at an audience that will pay, doing something a free model with a good prompt cannot. Honest moat verdict: the durable advantages are **currency** (a maintained capability watchlist beating models' self-knowledge), **verification with evidence**, and the **build bridge** (we can build the idea, not just describe it) — a process/freshness/integration advantage, not a static asset. Includes the brand-fit corollary (treat the product collection as separate from the domain portfolio; match deliberately).
- **sources:** idea.uk/PLAN_idea_uk(3).md#5; idea.uk/idea_uk_method_v0(3).md; idea.uk/running_notes(63).md (2026-05-27 arc)
- **relations:** ideation method; capability watchlist; five-layer stack; paid multi-domain chat plan (§10 of that doc).
- **verify-later:** whether the capability watchlist exists as a recurring workflow anywhere in scheduled_tasks/agent_definitions.

### Sale-readiness / separability discipline (assets as data, minimal identifiable dependency set)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** PLAN_idea_uk §2 rule "keep our asset list as input data, never built into the method"; RUNBOOK_idea_uk Notes: "the engine takes assets as data and the billing sits behind a provider interface, so idea.uk remains a separable unit".
- **what:** idea.uk is built to be sold as a working unit: business assets are always passed in as data (so the same engine serves internal domains and strangers), the set of workflows/actions it uses is kept identifiable and minimal, and billing sits behind a provider interface. The standalone Go service honours this (stdlib-only, file store, FakeProvider fallback).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/RUNBOOK_idea_uk(9).md; idea.uk/idea_uk_architecture_and_deployment(6).md#1
- **relations:** provider abstraction (payments); engine Go port.
- **verify-later:** golang_files/engine.go input contract; billing.go Provider interface.

### idea.uk as an instance of the paid multi-domain chat plan (day-pass lineage)
- **category:** business-strategy
- **status-signal:** superseded
- **status-evidence:** PLAN_idea_uk §2 "idea.uk is itself an instance of the paid multi-domain chat"; the built product ended up a report service, not a chat domain — the worker/paywall/day-pass reuse never happened in the shipped form.
- **what:** idea.uk originated as one configured domain of a planned "simple paid multi-domain chat" product (edge worker + paywall + day-pass), with the ideation method as its bound tool. The 2026-05-27 running-notes arc covers day-pass economics, per-domain monetisation by domain type, and serverless-edge vs central-nginx topology. The shipped idea.uk deliberately diverged: it is NOT edge-shaped (minutes-long background job → always-on service).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/running_notes(63).md (2026-05-27, "Pivot to simple paid multi-domain chat", "Topology note: idea.uk is NOT pure-static/edge")
- **relations:** PLAN_simple_paid_multidomain_chat.md (outside this unit); hosting split concept below.
- **verify-later:** whether the chat/day-pass product exists anywhere else in docs/ (other units).

### Voluntary pay and "free goes" rejected → free taster + paid report
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** idea_uk_open_discussion §5 (2026-05-28): "probably not a good idea in this form… Drop voluntary pay and the multi-free-go idea. The taster is the better hook."
- **what:** Voluntary-pay ("pay if satisfied") and N-free-goes monetisation were analysed and rejected (abuse risk, no demand signal, trivially circumvented). Replaced by the pattern that shipped: a free, cheap (~£0.02) audience-check taster as proof-of-value plus a £29 full report with refund guarantee.
- **sources:** idea.uk/idea_uk_open_discussion.md#5; idea.uk/running_notes(63).md ("Day-pass collapses payment complexity", CHECKPOINT 2026-05-28 §4)
- **relations:** audience-check taster endpoint; pricing decisions.
- **verify-later:** n/a (business decision).

### Unit economics, pricing, and sourcing decisions (incl. self-hosting deferred)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** running_notes SESSION DECISIONS LOG 2026-06-11; open_discussion §§1–2, 6 with verified May-2026 pricing; £29 live and proven with a real card 2026-06-14.
- **what:** Per-run engine cost ~£0.40–0.60 (verify step dominates; optimisable to ~£0.20–0.30 via Haiku scoring + prompt caching); Stripe UK fees 1.5%+£0.20, break-even ~£0.72, worst-case refund cost ~£1.43; price settled at **pay-per-idea, cost-plus, £29 flat** (not B2B SaaS for the ideation product itself). Self-hosted LLMs analysed and deferred ("a 2027 decision, not a 2026 one") — commercial frontier models win at this volume, and open-weight models are weakest exactly at the cut step's ruthlessness.
- **sources:** idea.uk/idea_uk_open_discussion.md#1-2,6; idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md (pricing checkpoint)
- **relations:** Stripe webhook pattern; engine model line-up.
- **verify-later:** REPORT_PRICE_GBP env on the live box.

### Ideation method v0→v3 (staged, multi-model, web-verified pipeline)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Method doc carries v1/v2/v3 patches with dated rationale; engine validated end-to-end live 2026-06-04; live product runs it per paid order.
- **what:** The runnable method: (1) frame AND challenge the audience; (2) generate candidates across four lenses — demand, generalist-failure, frontier, outcome — plus the original asset×capability sweep; (3) cut each candidate against the *specific named free substitute* with a different model, incl. the seller-bundles-support-free check; (4) verify survivors with web research, evidence attached; (5) score; (6) rank and split test-now vs consider. Version history is itself conceptual: v1 added Durability + the specific-substitute cut; v2 added multi-lens generation + audience-fit challenge (single-lens generation diagnosed as "narrow — supply-side only"); v3 added the Risk column. Dogfooding rule: if the method can't find an advancing candidate for idea.uk itself, it isn't good enough.
- **sources:** idea.uk/idea_uk_method_v0(3).md; idea.uk/idea_uk_testrun_v2.md; idea.uk/KEY_DOC_idea_method_prompt.md; idea.uk/running_notes(63).md (method-run checkpoints)
- **relations:** operator-risk column; cross-vendor critique; capability watchlist; engine implementations.
- **verify-later:** golang_files/prompts.go step prompts vs the method doc.

### Operator-risk column: hazard scored separately from fitness, with gates
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Debugging-guide item 23 (2026-05-28) documents the addition end-to-end; A1 acceptance notes the live report carries "Operator risk: N/5" with auto-flags.
- **what:** A sixth scored dimension, Risk to the operator (1–5, 5 safest), scoring the CONSEQUENCE of being wrong, not the probability. Deliberately **not** added to the fitness sum and **not** in the Def≥3∧Will≥3 gate: Risk=1 (regulated professions) is dropped automatically into a visible "Dropped for operator risk" section; Risk≤2 advances but flagged "needs liability work before building" with the cheapest_test forced to demand PII + solicitor-reviewed T&Cs first; Risk breaks ties toward safer builds. Generalisable lesson: when a scoring system recommends actions to an operator who carries downstream exposure, hazard must be a separate scored dimension — fitness sums cannot see it. First real effect: paused the SFI single-farm assessment.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (item 23); idea.uk/idea_uk_method_v0(3).md (Risk rubric); idea.uk/LIABILITY_AND_TERMS(2).md (header)
- **relations:** liability framework; SFI26 Diff Alerts swap; ideation method.
- **verify-later:** engine.go `scored` struct + riskNote; idea_method_runner.py parity.

### Capability watchlist + real-world event-window watchlist
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** PLAN_idea_uk §8: "the capability watchlist runs as its own recurring research workflow" (stance accepted); no evidence either watchlist was ever built as a workflow.
- **what:** Two maintained lists that feed re-runs of ideation: (1) AI capabilities worth using now, grouped by what specialism does that generalists don't — the mechanism for being early to ideas a new capability just unlocked, and the single strongest durable advantage; (2) real-world event windows per domain (scheme deadlines, regulation changes — e.g. SFI26 Window 1), because timing changes what's actionable. The capability menu v1 ships inside the method/prompts; the *recurring maintenance workflows* remain unbuilt.
- **sources:** idea.uk/idea_uk_method_v0(3).md (capability list v1); idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md ("Watchlist should track scheme/event windows")
- **relations:** differentiator framework (currency moat); scheduler-and-tasks (would host the recurrence).
- **verify-later:** scheduled_tasks / agent_definitions for any watchlist workflow.

### Cross-vendor critique (the cut step on a different vendor)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Architecture doc §9: `[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic` stderr line added; env-switched via OPENAI_API_KEY.
- **what:** The method's quality gate (the cut) is run by a different model from the generator — ideally a different **vendor** — so the method isn't one model marking its own work. Implemented as an optional OpenAI branch on the cut step (OPENAI_API_KEY + OPENAI_CRITIQUE_MODEL); same-vendor fallback still uses a different model (Sonnet vs Opus). Cross-vendor comparison flagged as an untested open experiment.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#9; idea.uk/idea_uk_method_v0(3).md (diff-model markers); idea.uk/RUNBOOK_idea_uk(9).md (go-live step 5)
- **relations:** ideation method; multi-model ensemble moat claim.
- **verify-later:** engine.go cut-step branch.

### Engine implementations: single-shot prompt → Python runner → Go engine (with LLM feature upgrade)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** "Ported the idea.uk tooling from Python to Go (platform is Go throughout)" (running notes 2026-05-28ff); A1 DONE 2026-05-28 with live validation 2026-06-04; python files retained as superseded parity copies.
- **what:** Three coexisting expressions of the method: a paste-anywhere single-shot prompt (weakest — one model marks its own work), the Python `idea_method_runner.py`/`idea_service.py` originals (superseded), and the shipped Go engine (`engine.go`+`prompts.go`, stdlib-only, no SDKs, offline `GOPROXY=off` build). The A1 upgrade set the model line-up (Opus for generate/verify, Sonnet for cut/score, all env-overridable) and added extended thinking per step (off for brainstorm breadth), prompt caching on static system blocks, `web_search_20260209` + code-execution filtering on verify, and a `WEB_SEARCH_MAX_USES` budget (raised 6→12 after a quota-exhausted run left premises "provisional").
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1; idea.uk/golang_files/engine.go (header); idea.uk/python_files/idea_method_runner.py (header); idea.uk/RUNBOOK_idea_uk.md (base — Python era, family-delta)
- **relations:** LLM API shape disciplines (the three bugs found during validation); ideation method.
- **verify-later:** engine.go callClaudeOpts / usesAdaptiveThinking.

### idea.uk service: request-then-confirm flow, REVIEW_BEFORE_PAY, AUTO_DELIVER, capacity cap
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Live and earning; full flow proven end-to-end with a real card 2026-06-14 ("LIVE BUG RESOLVED: paid + report delivered end-to-end").
- **what:** The order state machine: visitor `/request` (free) → operator confirm/decline → pay → fulfil. Two switchable shapes: charge-first (engine runs after payment; AUTO_DELIVER=false holds the report for operator review) and **review-before-pay** (default from 2026-06-11: confirm runs the engine first, operator reviews the draft, `/approve` sends the pay link — money is only taken after the operator has seen the deliverable). `MAX_ACTIVE_ORDERS` caps in-flight orders so capacity can't be oversold; `/capacity` exposes it. Orders live in a JSON file store (`/var/lib/idea/orders.json`) — deliberately no DB on the exposed box.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (flow + 2026-06-11 update); idea.uk/golang_files/service.go (header); idea.uk/idea_uk_architecture_and_deployment(6).md#5
- **relations:** Stripe webhook truth; B2 dead-drop persistence design (future DB); liability framework (operator review as mitigation).
- **verify-later:** service.go state machine + service_test.go (19+ checks).

### Free audience-check taster endpoint
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** A2 DONE 2026-05-28 with acceptance ticks; live on the page; taster now logs the result (2026-06-11 checkpoint b).
- **what:** `/audience-check` — the method's step 1 (audience challenge + 2–3 alternative audiences) exposed as a free, no-auth, ~£0.02/run, ~10s taster: the conversion hook that replaced voluntary-pay. Per-IP sliding-window rate limiting (3/h, 20/day) with Retry-After; XSS-escaped HTML fragment for direct innerHTML insertion; TASTER_ENABLED kill switch; each run logs business/audience/result as market intelligence.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A2; idea.uk/golang_files/audience_check.go (header); idea.uk/idea_uk_open_discussion.md#5
- **relations:** voluntary-pay rejection; ideation method step 1.
- **verify-later:** audience_check.go limiter + tests.

### Click-through operator approval links (HMAC per-order tokens)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Checkpoint 2026-06-11 (f)/(g): built, then "click-through confirmed working by user".
- **what:** Request/review emails carry links to a page with Confirm/Approve/Decline buttons. The link carries an HMAC(order id, INTERNAL_API_KEY) token authorising that one order only; the link opens a **safe GET page** (mail-scanner prefetch can't trigger anything) and the action fires only on a button POST; actions stay gated by order status so a token can't double-fire. Curl + X-Internal-Key remains the fallback.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (2026-06-11 update); idea.uk/running_notes(63).md (checkpoints f–g)
- **relations:** request-then-confirm flow; hitl (same shape as approval flows).
- **verify-later:** service.go token mint/verify.

### Fake-door → intent-capture-first launch discipline
- **category:** NEW:idea-product
- **status-signal:** superseded
- **status-evidence:** PLAN §7 step 4 ("intent capture first, no payment"); superseded by the live request-then-confirm flow with real Stripe (the fakedoor page became the embedded landing page).
- **what:** Launch pattern: a static landing page offering the report at a flat price, capturing intent without charging ("we reply within 24h with a confirmed slot + payment link, or a polite decline") — deliberately avoiding charge-then-fail refund overhead — with a visible monthly slot count to throttle demand to manual capacity. Also prescribed as a parallel track for the strongest single-domain candidate (agritec SFI26 checker). The page evolved into the embedded `page.html` of the live service.
- **sources:** idea.uk/PLAN_idea_uk(3).md#7; idea.uk/running_notes(63).md ("Built the idea.uk fake-door page", "Fake-door modified to intent-capture-only"); idea.uk/idea_uk_fakedoor(9).html (deployment notes header)
- **relations:** request-then-confirm flow (its successor); demand-test philosophy in the method's cheapest_test.
- **verify-later:** n/a (historical).

### Deliverable quality standards for reports and product emails
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** BUGS_idea_uk 2026-06-11 entries all marked "Fixed this build" with standing "for future builds" rules.
- **what:** Standing rules distilled from report-email review: every customer-facing string in plain English for a non-technical owner (jargon/acronyms treated as defects); every standalone deliverable opens with a one-paragraph plain summary of what it is; rejected options always say what the thing was and why it died; deliverables get a deliberate professional design distinct from marketing surfaces (the £29 report email: navy/gold/serif "sheet" look, unlike the landing page); illustrative examples must not leak into generated output (audience-anchored generation). Transport rule: any HTML email must be base64/quoted-printable encoded (the SMTP 998-octet line-fold corrupted raw HTML mid-tag).
- **sources:** idea.uk/BUGS_idea_uk(4).md; idea.uk/RUNBOOK_idea_uk(9).md ("HTML emails are base64-encoded")
- **relations:** content-quality (platform analogue); transactional email realities.
- **verify-later:** service.go b64Body; report HTML renderer.

### Chassis-native idea engine (Phase D `idea-orchestrator`)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK Phase D "Not started yet… needs a schema pass first"; architecture doc §8 declines to write the SQL until the action contracts are read.
- **what:** The second way to run the method: as a chassis agent + workflow reusing existing actions almost 1:1 (execute_llm_prompt for frame/generate/cut/score, web_search for verify, HITL actions for the operator gate, store_result/write_site_spec for persistence) — for running the method internally across our own domains on schedule (the Layer-4 planning input). The billing half deliberately stays in the standalone service ("a product/payment concern, not an agent workflow"). Bundle 2 packages exactly this port task.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#8; idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase D); idea.uk/BUNDLE_2_chassis_idea_engine_workflow.md
- **relations:** five-layer stack (L4); development-guide conventions (every agent an orchestrator; spawn sub-agents).
- **verify-later:** agent_definitions for any idea-orchestrator; the docubundle context file.

### Multi-tenant branded intake pages on one central engine (white-label Option C)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** open_discussion §7: Option C "RECOMMENDED… Want me to do this in the next round?" — never built.
- **what:** Other sites offer the ideation product via their own branded static request page (built through the normal pipeline, own price/copy) POSTing to the central service with a tenant_id; per-tenant Stripe branding; iframe and CNAME/reverse-proxy options analysed and rejected. Needs ~100–200 lines (tenant field on Order, tenants config, tenant-aware /request). Shape A (site IS the service) vs Shape B (request panel on a content site) hosting split defined in the architecture doc; a forked-component tool is explicitly the wrong model for a server-side paid engine — sites only ever *link* to it.
- **sources:** idea.uk/idea_uk_open_discussion.md#7; idea.uk/idea_uk_architecture_and_deployment(6).md#7
- **relations:** tool-library boundary (why the engine is not a content_component); site_plan blocked/planned mechanism.
- **verify-later:** service.go for any tenant handling (expect none).

### Real-door streaming progress page + programmatic refund endpoint (Phase A3/A4)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK A3/A4 have outputs+acceptance defined but no DONE mark; refunds confirmed manual-only in the Stripe section ("There is no refund code").
- **what:** A3: post-payment page polls `/status/{order_id}` and renders live engine progress ("generating… cutting… verifying claim 1 of N"), report renders in-browser — the "real door" UX (option (a) of the real-door analysis; the honest 72h email model shipped instead). A4: operator-gated `/refund` calling Stripe POST /v1/refunds and marking the order refunded — refunds today are manual dashboard clicks and the app doesn't record them.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A3-A4; idea.uk/idea_uk_open_discussion.md#3-4; idea.uk/RUNBOOK_idea_uk(9).md (Refunds — manual)
- **relations:** request-then-confirm flow; Stripe pattern.
- **verify-later:** service.go routes (expect no /status, /refund).

### SFI26 Diff Alerts (first vertical tool) — replacing the single-farm assessment
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** "Tool swapped 2026-05-28… paused on liability grounds"; Phase C fully specified (C1–C5) with no build evidence; the base DEVELOPMENT_RUNBOOK still carries the original single-farm Phase C (family-delta capture of the abandoned product).
- **what:** The first Layer-3 vertical tool: a subscription digest for UK farm advisors summarising what changed in Defra/RPA SFI26 guidance, from a versioned scraped corpus, with every change cited to source+version, weekly, operator-reviewed for 8 issues before auto-send. Scored 19/25 with Risk 4. It replaced the **SFI26 single-farm assessment** (abandoned/backlogged: Risk 2 — a wrong number could cost a farmer £5–50k), the first product decision the Risk column changed. Chassis-native by design (recurring, per-user state, scheduled), the opposite plumbing to standalone idea.uk.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase C + swap note); idea.uk/DEVELOPMENT_RUNBOOK.md (base — original single-farm Phase C); idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 3)
- **relations:** operator-risk column; liability framework (SFI T&Cs draft); vet-med-pricing (sibling scraping shape).
- **verify-later:** any SFI corpus/agent in the repo or DB (expect none).

### idea.uk standalone service page-serving and deploy gotchas
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Debugging guide §11 added for the idea.uk service; each gotcha tied to a fixed live incident.
- **what:** The operational failure catalogue for the single-binary service: every served path needs an explicit mux handler (bare 404s on linked pages); `writeHTML` fragments vs the `a.page()` full-page brand wrapper (navigation targets must wrap; injected fragments must not); startup templating of `CONTACT_EMAIL`/`MONTH_SLOTS` placeholders; systemd EnvironmentFile keeps inline comments (crash-loop + nginx 502); certbot failure made non-fatal in setup.sh; replace a running binary by scp-to-temp + `mv -f` (text-file-busy); Let's Encrypt rejects placeholder emails.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md#11; idea.uk/golang_files/README_setup_SETUP.md; idea.uk/BUGS_idea_uk(4).md (mobile safe-area padding)
- **relations:** setup.sh; VM launch plan.
- **verify-later:** service.go routes() vs page.html hrefs.

### Stripe integration pattern: webhook as the only source of truth
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** Live £29 payments proven end-to-end 2026-06-14 (incl. resolving the stray-character webhook-secret incident); full setup documented from the real dashboards.
- **what:** The reference payments pattern proven by idea.uk: entitlement/fulfilment granted only on a signature-verified `checkout.session.completed` (browser redirects prove nothing); webhook handling idempotent via an event-dedup table; a **restricted** API key scoped to Checkout Sessions:Write only; test and live are separate accounts with separate webhook destinations and secrets ("a sandbox webhook does not cover live"); the signing secret must be byte-exact (one pasted stray character 400'd every event and stalled a paid order — recovered by resending the event); Stripe keeps its fee on refunds; no SDK — raw HTTP + HMAC verify.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (Stripe billing — setup + troubleshooting); idea.uk/PLAN_stripe_billing_integration(3).md (idea.uk reference block); idea.uk/golang_files/billing.go (header)
- **relations:** platform billing plan (adopts these principles); request-then-confirm flow.
- **verify-later:** billing.go webhook verify (HMAC-SHA256 over timestamp+body, constant-time compare).

### Platform Stripe billing integration plan (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** "the auth service has a subscription scaffold… but no working payment integration — no Stripe SDK, no checkout creation, no webhooks"; every DDL marked PROPOSED.
- **what:** The chassis-wide billing design for the build/host/chat product: truth lives in the auth DB mutated only by verified webhooks; the chassis gates on a one-way-fed `client_entitlements` cache (Kafka entitlement-changed events + reconciliation sweep) because the maintenance heartbeat can't call auth per site; two charge shapes — recurring tier subscription per client and a one-off **$5 build credit** (Checkout mode=payment, consumed via the atomic-claim idiom); build-submission gate reuses the `approval_mode` hold; provider interface from day one. idea.uk is the cited working reference for the one-off path.
- **sources:** idea.uk/PLAN_stripe_billing_integration(3).md; idea.uk/RUNBOOK_idea_uk(9).md (reference implementation)
- **relations:** Stripe webhook pattern; admin-dashboard-and-api (auth service); scheduler heartbeats.
- **verify-later:** auth service repo subscriptions tables; any client_entitlements table (expect absent).

### Operator email identity: leopardess.uk + deterministic per-site addresses + email spec aspect
- **category:** NEW:email-infrastructure
- **status-signal:** partial
- **status-evidence:** "Status: design, not yet implemented in the chassis. idea.uk… carries these values in its env"; the identity scheme is live for idea.uk (idea-uk@leopardess.uk), the aspect/provisioner are design-only.
- **what:** One neutral operator domain (leopardess.uk — also given a one-page identity site) fronts all sites' transactional/support mail. Per-site address = deterministic encoding (lowercase, dots→dashes, @operator_domain), resolved by matching against the known-domain set, never by reversing; collisions detected at assignment and stored overrides win. A new `site_specs` aspect `email` (no DDL) carries per-site identity/status/provider; a future `email-provisioner` agent flips provisioned=false→true (same provision-and-write-back shape as model-trainer/Thunder). Refined 2026-06-06: prefer a **specific forwarder per published site** over a server catch-all (no backscatter; only forward addresses that exist).
- **sources:** idea.uk/EMAIL_identity_in_site_spec(5).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 correction); idea.uk/leopardess_uk_index.html
- **relations:** transactional sending realities; site-spec aspect model (021); feasibility-recheck promotion mechanism.
- **verify-later:** site_specs DISTINCT aspect for 'email' (expect absent); 021 doc aspect list.

### Transactional email sending realities (587-only, relay filtering, SES + per-domain DKIM)
- **category:** NEW:email-infrastructure
- **status-signal:** deployed
- **status-evidence:** "DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels" (2026-06-11); SES live in production same day; EMAIL doc header codifies the lesson for the future provisioner.
- **what:** Hard-won operational truths now standing framework guidance: cloud boxes can't use outbound SMTP 25/465 (Hetzner leaves only 587 submission open — the cPanel UI advertising 465 misleads); Go's smtp.SendMail does STARTTLS not implicit-TLS, so a 465 path needs a tls.Dial branch; shared-host relays (Clook→MailChannels) content-filter legitimate transactional mail (a `From:`-like line + raw JSON in a body triggered "Spam Content"); therefore transactional sending needs a **dedicated sender (AWS SES eu-west-2) with the operator domain's own DKIM**, bodies kept clean, and the mailer async/bounded so a hung send can't freeze the request path. Gotcha: SES SMTP_USER is the AKIA access-key-id, not the IAM user name (535s otherwise). Chronology: Clook both-ways → catch-all/Default-Address fixes → MailChannels blocks → SES.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05/06/10/11 updates); idea.uk/EMAIL_identity_in_site_spec(5).md (2026-06-11 header + operational note); idea.uk/running_notes(63).md (email checkpoints)
- **relations:** operator email identity; deliverable quality standards (clean bodies).
- **verify-later:** /etc/idea/idea.env SMTP block (on the box); smtpSend in service.go.

### Liability framework: risk-tiered mitigations, disclaimers, PII, and draft T&Cs
- **category:** NEW:legal-liability
- **status-signal:** partial
- **status-evidence:** /terms, /refund-policy, /privacy live on the service (2026-06-05) with AI-can-be-wrong wording; solicitor review explicitly still pending (A6 open); PII quote a kickoff item, no policy on file evidenced.
- **what:** The full liability posture for AI-analysis products: negligent misstatement (Hedley Byrne) named as the real exposure route; disclaimers must be conspicuous and *proximate* (in the report itself, top-of-report box, not just site footer); every claim cited + date-stamped, versioned corpus as audit trail, 6-year input/output retention; operator review of early deliveries; generous visible refunds; PII insurance with the AI-assisted-human-reviewed framing disclosed honestly; limited company for payment. Draft starter T&Cs for both idea.uk (information-not-advice, liability capped at fee) and the sharper SFI product (verify-before-acting obligations, exclusions list). Policy pages served from string constants through the brand wrapper with an {{EMAIL}} token; UK-GDPR privacy naming Stripe/Anthropic as processors. The mission's "never verdicts → opinion+evidence+questions" framing is the same posture applied to site content.
- **sources:** idea.uk/LIABILITY_AND_TERMS(2).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 update); idea.uk/DEVELOPMENT_RUNBOOK(3).md#A6-A7
- **relations:** operator-risk column (what triggers "needs liability work"); idea.uk mission; SFI Diff Alerts.
- **verify-later:** live /terms content; any PII policy record.

### Layer-5 gap = a persistent-service wrapper on already-deployed Thunder plumbing
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "The hard plumbing for Layer 5 already exists and is deployed in production… The remaining gap is a persistent-service wrapper — modest, and largely assembling existing pieces" (2026-06-04); no service-deployer built.
- **what:** The honest reassessment of automated backend deployment: provisioning, ssh_exec, presigned-B2 file transfer, and decommission all exist (Thunder adapter, verified in production), but they're built for **ephemeral** training VMs (18h cap, 15-min reaper, credential-free). A persistent service is the exact opposite shape, so the gap is: persistent-mode provisioning (reaper exemption), credential delivery to the box, DNS+TLS wiring, a `service_instances` table (sibling of thunder_instances), and a parameterised setup script — a `service-deployer` orchestrator modelled on model-trainer, with idea.uk as first consumer. Two distinct things kept clear: deploying the engine binary to a VM (infrastructure) vs expressing the engine as chassis actions (Phase D) — complementary, not alternatives.
- **sources:** idea.uk/PARALLEL_engine_deployment_and_layer5.md; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 5)
- **relations:** Thunder adapter (docs033); model-trainer pattern; Path A/setup.sh; 007 box recipe + site_api_routes.
- **verify-later:** thunder_instances table; absence of service_instances; cmd/thunder-adapter actions.

### Path A manual VM deploy — setup.sh as the future service-deployer payload
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** idea.uk LIVE on the Hetzner box 2026-06-05 via this path; setup.sh iterated through real incidents (certbot abort, env comments).
- **what:** "Do it by hand once, and capture the steps as the automation artefact": a single idempotent, non-interactive, parameterised `setup.sh` that converges a fresh Ubuntu box to nginx+TLS+ufw+fail2ban+unattended-upgrades+hardened systemd unit+binary — deliberately written so the chassis service-deployer can later `ssh_exec` the same file (MODE=update = binary swap; re-run = rebuild; anti-lockout guard on SSH password disable). The single-binary model: landing page `go:embed`ded, env in /etc/idea/idea.env, atomic mv-based redeploys.
- **sources:** idea.uk/nginx/README.md; idea.uk/nginx/setup.sh.orig3 (header); idea.uk/nginx/README_setup_box.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md (Path A)
- **relations:** Layer-5 wrapper (Path B); page-serving gotchas; VM launch plan.
- **verify-later:** the live box's drift vs setup.sh (the doc's own rule: fold tweaks back in).

### VM launch plan — dedicated hardened box, prior OVH reverse-proxy files audited
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** Box provisioned 2026-06-04 (Hetzner CX, Nuremberg) following this plan; the year-old files' concrete bugs "all catalogued in the doc".
- **what:** Infrastructure-track decisions: a **dedicated** VM for idea.uk rather than the existing shared OVH multi-domain reverse proxy (blast-radius isolation; the proxy only knows how to reach k8s); reuse of the prior Terraform/nginx/fail2ban/logrotate/prometheus patterns with their specific year-old bugs fixed; secrets confirmed clean before reuse; VM sizing grounded in the engine being I/O-bound (1 vCPU / 512MB–1GB); search-grounded provider comparison (Hetzner vs Oracle vs spot).
- **sources:** idea.uk/nginx/VM_LAUNCH_PLAN.md; idea.uk/running_notes(63).md (2026-06-04 infra checkpoints)
- **relations:** Path A; 007 adoption-pipeline box recipe; Layer-5 wrapper.
- **verify-later:** the OVH proxy box's role for content sites (51.89.148.216 → k8s NodePort).

### B2 dead-drop persistence: one-way flow from the exposed box into the framework DB
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "Persistence decisions LOCKED (updated PERSISTENCE_design.md §10)" 2026-06-04 — design settled, but the live service still runs on orders.json only; no idea-ingest agent or idea_orders table evidenced.
- **what:** Standard tiered/DMZ design for internet-facing satellites: the exposed idea.uk box holds NO core-DB credentials and no network path to the cluster; it keeps a local operational store (JSON now; SQLite analysed — would break the stdlib-only property) and writes immutable terminal-event records to a scoped write-only B2 prefix (the dead-drop, reusing Thunder's presigned pattern); a scheduled in-cluster `idea-ingest` agent polls B2 and idempotently INSERTs into framework Postgres — the system of record. Kafka topic / narrow HTTPS ingest / direct PG all rejected (each is an inbound path in).
- **sources:** idea.uk/nginx/PERSISTENCE_design(1).md; idea.uk/running_notes(63).md (persistence checkpoints)
- **relations:** storage-architecture (B2, presigned URLs); scheduler-and-tasks (the ingest schedule); checkpoint-upload plan (same threat model).
- **verify-later:** any idea-events B2 prefix or idea_orders table (expect absent).

### VM cutover: nginx front door with reserved tool paths (staging-in-place via DNS)
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** Runbook delivered 2026-06-21; "gated on P0 + the site review… deliberate, not done" (TODO P1, 2026-06-26).
- **what:** The go-live mechanism for a chassis-built front site on a VM that already hosts a live paid tool: because idea.uk's DNS (Cloudflare) points at the VM while the chassis deploys to B2, **every chassis build is invisible at the live domain — safe staging-in-place** — and cutover is purely an nginx change: static root for general pages, `location` proxies for the reserved tool paths (/request /confirm /approve /decline /stripe/webhook /internal/* /order/* /op /health /capacity + policy pages), `try_files … =404` so a missed tool path fails loudly, no body rewrites on the webhook location (signature integrity), prove the webhook through nginx BEFORE cutover, rollback = restore one server block. Named biggest risk: reserved-path completeness. Monorepo stays authoritative; the VM is just one more consumer (pull-sync from B2/git or a path-conditional Action push).
- **sources:** idea.uk/RUNBOOK_idea_uk_vm_cutover.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 2); idea.uk/TODO_chassis_and_idea_uk(1).md#P1
- **relations:** scheme-to-components P0 (the gate); deployment-github (monorepo → Actions → B2); hybrid build_approach/hosting_trajectory classifier fields.
- **verify-later:** live nginx config on the box; whether cutover has since happened.

### Two-stage base+override design pipeline (site-design-planner + webdesign-agent)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Investigation 2026-06-20 "confirmed against agent_definitions… and the deployed render code"; routing table shows needs_composition→site-design-planner, needs_design→webdesign-agent with an explicit depends_on.
- **what:** Design is deliberately split (027 §2 — ordering was the reason): Stage 1 `site-design-planner` (deterministic, no LLM) resolves layout/typography/palette and installs them (css_themes 3-FK row + style_collections + sites.style_collection_id + a resolved_composition spec) — renders nothing; Stage 2 `webdesign-agent` produces an LLM design overlay, renders the layout template over the installed composition base per a fixed merge-authority rule (LLM wins core palette slots + typography; composition wins layout/structure tokens/specialised slots), and is the sole styles.css deployer (git_commit → Actions → B2). `emit_design_items` queues both from one step with needs_design gated on composition. The 2026-06-20 correction: this is NOT a shared-responsibility bug — the overlay was designed as *optional and partial* (025 §5).
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md; idea.uk/HANDOFF(13).md (three-layer model); idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md
- **relations:** mandatory-full overlay bug; resolved_composition pointer; palette cascade; 002 architecture doc (rewritten 2026-06-22 to match).
- **verify-later:** render_css_composition_helpers.go buildPaletteMap/buildTypographyMap; emit_design_items_action.go.

### resolved_composition pointer spec + install_site_composition semantics
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Verified live on idea.uk twice (dark install, then re-resolve); "resolved_composition is a *pointer* — it carries palette_id/name/source, not the colour values".
- **what:** The composition install contract: css_themes row created with all three FKs but empty css_content (webdesign-agent fills at render); style_collections points at the theme; `sites.style_collection_id` set only if NULL — install **errors rather than overwrites** an existing composition ("re-resolve not supported; clear it manually"), which is why re-resolving requires an explicit detach; the old resolved_composition spec is superseded and a new one inserted as the lineage/decision record (`lineage.{palette_source, typography_source, layout_source}`). Renderer resolution is strict: missing/NULL composition parts hard-error ("migration gaps are audit events, not silent fallbacks"), with a loud emergency fallback to standard-brochure.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (install + loader sections); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A: "a bare needs_composition requeue no-ops")
- **relations:** composition re-resolve procedure; two-stage pipeline.
- **verify-later:** install_site_composition_action.go; render_css_composition_loader.go.

### Palette/typography resolution cascade + the dead-slot bug and fingerprint fallback hardening
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Cascade live and proven; dead slot "CONFIRMED why (2026-06-19, from resolve_composition_palette.go)"; hardening "DELIVERED" as code 2026-06-19/20 but "READY… NOT YET APPLIED" in the backlog (needs image rebuild + roll).
- **what:** Palette source cascade: design_reference → mission → design_intent.palette.reference_values → layout seed → archetype default (typography analogous; palettes always site-specific, layouts a shared curated library). The bug: cascade slot 1 reads `design_reference.palette.reference_values`, a key the adoption fingerprint never writes (it stores suggested_mapping/css_variables/colors) — so slot 1 was dead and adopted references never drove the palette; the delivered hardening points slot 1 at the fingerprint's real keys as a fallback after design_intent. Under the current LLM-wins merge the composition palette mostly doesn't paint anyway (it fixes lineage + rare-gap fallback) — the painting lever is the classifier fix feeding the LLM.
- **sources:** idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md#3; idea.uk/HANDOFF(13).md (cascade + backlog item 3); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (problem 1)
- **relations:** structured design_intent migration; two-stage pipeline; adoption generate_design_intent.
- **verify-later:** resolve_composition_reference_helpers.go deployed or not; extractPaletteSignal/extractTypographySignal.

### Structured design_intent from the classifier (palette + typography reference_values)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Migration applied 2026-06-20; "Palette migration proven on a real build" 2026-06-21 — fresh idea.uk resolved `palette_source=design_intent_values`, parchment, "no invented blue".
- **what:** Root cause of generic-looking fresh builds: the classifier wrote design_intent colours as **prose** (hex buried in colour_mood sentences) while every consumer — the composition cascade and the analyze_design prompt — reads **structured** `design_intent.palette.reference_values` (8 slots) + `typography.reference_values`. The migration edits the classifier's classify_and_extract schema and adds a MANDATORY-fields bullet (all 8 slots as hex; style_direction must agree with the palette; never default to blue-and-grey), applied via snapshot_agent backup + exact-anchor replace() with a RAISE self-check. This single change is what makes both design stages agree (base = parchment, overlay starts from parchment).
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A + checklist); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md ("Direct consequence" section)
- **relations:** two-stage pipeline; mandatory-full overlay bug (this is precondition for its fix); prompt-migration discipline.
- **verify-later:** live classifier prompt contains the palette block; site 1244516d design_intent spec.

### The mandatory-full overlay bug + improver-not-rewriter fix (and the superseded rewrite options)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** "direction settled 2026-06-20, not started" (runbook open items); superseded options (a)/(b)/(c) "kept for history" in the investigation doc.
- **what:** The merge rule assumed the LLM overlay would be *optional and partial* (asserting only genuine brand identity), but the analyze_design prompt mandates a full 8-slot color_scheme with "be distinctive" framing — so the LLM repaints every fresh build, the 028-forbidden silent override. Fix v1 (no contract change): show the LLM the established palette, require it to keep it as the foundation and change slots only with a reason, diff the result against the composition base and write an audit record (slot, old→new, reason) per build; v2 (deferred, evidence-driven): cap core-slot changes per refine + optional denylist. Explicitly supersedes the earlier single-owner options: (a) LLM-owns-core / slim the planner, (b) flip the merge so structured composition wins, (c) collapse to one design agent — rejected because the base+partial-overlay split is intentional.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (CORRECTION + fix sections; superseded options); idea.uk/DESIGN_PIPELINE_two_track_investigation(1).md (the pre-correction "decision options" — family-delta); idea.uk/HANDOFF(13).md (backlog item 4)
- **relations:** structured design_intent (precondition); design docs 025/027/028.
- **verify-later:** analyze_design prompt in webdesign-agent def; any design-audit table.

### Scheme-aware weighted layout matcher + layouts.scheme + tool-portal-light
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Matcher code LIVE — merged … built into the chassis image and site-design-planner rolled. (Confirmed live 2026-06-25.)"; migration applied; re-resolve proved tool-portal-light selection end-to-end 2026-06-25.
- **what:** Replaces the tags-only, scheme-blind `resolveLayoutByTags` (exact-overlap count, alphabetical ties) that put light-editorial idea.uk on tool-portal-dark. New matcher (all in Go, ~17-row library fetched and scored transparently): scheme as a **near-hard constraint** (a light site won't land dark while any non-dark fits; mismatch queues the existing needs_new_layout_candidate HITL item), IDF-weighted tag rarity (specific beats generic), synonym normalisation to a controlled vocabulary, category + description keyword bonuses. Paired migration adds nullable `layouts.scheme` (light/dark/neutral; NULL degrades gracefully) and a new `tool-portal-light` layout — same structural class contract as its dark twin, light fallbacks, reads palette vars. Decision history: NO auto-layout-generation — a curated, varied library + scheme-aware matching is the lever; LLM-judge/pgvector deferred.
- **sources:** idea.uk/resolveLayoutByTags_weighted.go.patch.txt (header); idea.uk/migration_layouts_scheme_and_light_tool_portal.sql; idea.uk/HANDOFF(13).md (matcher rewrite)
- **relations:** deriveSchemeFromDesignIntent; composition re-resolve; scheme-to-components gap (the next layer down).
- **verify-later:** fork_theme_composition.go current resolveLayoutByTags; remaining NULL layouts.scheme rows (backlog).

### Composition re-resolve procedure (gated, file-based, backup-first)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Steps 1–6 all marked DONE with results (2026-06-22→25); "RE-RESOLVE SUCCEEDED: idea.uk now on tool-portal-light (scheme fix proven end-to-end)".
- **what:** The safe pattern for re-running composition on an already-built site (install refuses overwrites): ordered SQL FILES — backup+inspect (with four uniqueness checks that must all be 0), gated detach+clear (NULL style_collection_id; delete the site's own collection→theme→palette→typography chain only where source_site_id matches; supersede the old spec), state-check, kcat re-trigger of site-design-planner (`domain` required by ensure_site_record), verify. Two learned caveats now doctrine: run SQL as files never pasted (paste mangled \set/blank lines and left an open transaction); a standalone-orchestrated planner ends at install and emits NO needs_design — the styles.css render is a separate explicit webdesign-agent orchestration. Distinct from the adoption teardown (bulk delete by source_domain), which must NOT be used on a fresh site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (re-resolve section); idea.uk/reresolve_idea_uk_01_backup_and_inspect.sql (+02/02b/03/04/05 series); idea.uk/running_notes(63).md (xxx–jjj checkpoints)
- **relations:** install semantics; launch idioms; scheme-aware matcher validation.
- **verify-later:** bak_*_idea_20260625 tables; orchestration_states rows for the re-resolve correlations.

### The scheme-does-not-reach-components gap (P0 framework fix)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Investigation complete + report written 2026-06-26 ("Step 7 DONE — conclusion: do NOT rebuild yet; structural gap"); the fix itself deferred to a dedicated thread; composition+stylesheet layers verified correct, page layer still dark.
- **what:** The central framework finding of the thread: a site's scheme (light/dark) is decided at composition and reaches styles.css `:root`, but never reaches the components that render sections/header/footer. Components are drawn from a **dark-oriented library** by a one-active-component-per-function lookup (nothing light exists for hero/CTA/footer), self-style via inline CSS with their own class vocabulary (the layout's section rules don't apply — class-name mismatch), and hardcode dark treatments or set dark `--section-*` themselves — so a light-resolved site renders dark chrome over light content (only var-reading sections went parchment). Supporting facts: `is_dark_section` is loaded but never used in selection, unreliable, and conflates "intrinsically dark" with "should contrast the page"; no layout declares default header/footer and the planner never runs update_site_defaults; the hero navy-button bug (--accent-color vs --color-accent) was already fixed in the library — deployed pages were stale. Modelled as: scheme was treated as colour-only; the structural half was never plumbed.
- **sources:** idea.uk/REPORT_scheme_does_not_reach_components.md; idea.uk/HANDOFF_scheme_to_components(1).md; idea.uk/running_notes_2(6).md (lll–ooo); idea.uk/one_sentence_description.md
- **relations:** scheme-as-override thesis (the fix shape); section-contrast model; header/footer wiring; component class-contract question; scheme-aware matcher (upstream, done).
- **verify-later:** whether the dedicated thread landed (component templates de-hardcoded; light footer exists; update_site_defaults in build path).

### Scheme-as-override thesis + section-contrast model (base scheme + per-section contrast intent)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** REPORT §4–5: "the likely shape (to be validated by the investigations, not assumed)"; eight design questions "must be answered before any code".
- **what:** The design steer for the P0 fix: scheme is a **set of variable values** — an override layer supplied by composition/renderer and consumed by de-hardcoded components — never a duplication of components into *-light/*-dark variants (new functions only for genuine structural divergence). The model separates **site scheme (base)** from **per-section contrast intent** (a dark hero on a light site is legitimate, intentional contrast — "make everything light" is wrong), both applied as data at render time through the existing `--section-*` mechanism, making the renderer the single adaptation point. Eight design questions scoped (where scheme lives at render; who owns section darkness; the override mechanism; the gating class-vocabulary question Q4; is_dark_section's fate; header/footer; migration without breaking dark sites; an auditor guard), with nine investigations (A–I) and a provisional fix shape stated as hypothesis. Definition of done includes a scheme-coherence audit so it can't silently regress.
- **sources:** idea.uk/REPORT_scheme_does_not_reach_components.md#4-8; idea.uk/HANDOFF_scheme_to_components(1).md; idea.uk/TODO_chassis_and_idea_uk(1).md#P0
- **relations:** CSS colour-inheritance model (the vehicle); improvement-loop (the guard); scheme gap (the problem).
- **verify-later:** stage-2 code check of component templates for --section-* consumption.

### CSS colour inheritance / --section-* luminance model (inline style = override, not main CSS)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Documented in 003 ("CSS Colour Inheritance Model") and restated as renderer contract in the tool-portal-light css_template header; the dark components *bypassing* it is the observed live behaviour.
- **what:** The platform's component colour contract as evidenced here: styles.css sets body/heading/element rules through `var(--section-*, var(--color-*))`; the renderer appends `--section-*` defaults after rendering based on palette/background luminance; a dark callout section overrides `--section-*` on its own container (sanctioned); layouts MUST NOT declare `--section-*` defaults; renderer-managed surface classes must be surface-coloured; a component's inline `<style>` is an **optional override, not its main CSS** (user correction, checkpoint mmm). The scheme gap exists precisely because dark components violate this — hardcoding backgrounds and `--section-*` inline. Two parallel styling systems (layout class vocabulary vs component class vocabulary) is the structural tension to resolve (Q4).
- **sources:** idea.uk/running_notes_2(6).md (lll/mmm); idea.uk/REPORT_scheme_does_not_reach_components.md#2; idea.uk/migration_layouts_scheme_and_light_tool_portal.sql (renderer contract comment); idea.uk/001_component_flow.md
- **relations:** scheme-as-override thesis; styling-render-pipeline (036); 003 contracts doc.
- **verify-later:** the luminance-appender code location (report investigation G names finding it).

### Header/footer chrome wiring chain (and its live gaps)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Data findings 2026-06-26: all inspected layouts have NULL default header/footer ids; idea.uk's new style_collection inherited NULLs; site_components still point at the original dark, now-inactive site-header/site-footer.
- **what:** Site-level chrome flows down a chain: `layouts.default_{header,footer}_component_id` → install_site_composition copies onto style_collections → `update_site_defaults` copies onto site_components → renderAndStoreSiteComponent renders into site_components.rendered_html, with a hardcoded RenderFallbackHeader when unlinked. Live gaps: no layout declares defaults, site-design-planner never runs update_site_defaults, and header/footer are therefore never scheme-derived — a re-resolve leaves the old chrome in place. The library has light headers but NO light footer. Fix direction: layouts declare scheme-appropriate defaults + the build runs update_site_defaults, one adaptive header/footer per the override thesis.
- **sources:** idea.uk/running_notes_2(6).md (mmm data findings); idea.uk/REPORT_scheme_does_not_reach_components.md (facts + Q6/investigation F); idea.uk/001_component_flow.md
- **relations:** scheme gap; contracts-and-standards Site Component Linkage Contract (003).
- **verify-later:** update_site_defaults_action.go and its call sites; how the original build chose idea.uk's header.

### Section→component resolution: direct-function Path 1 vs scoring selector Path 2
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** plan_sections code read 2026-06-26: "Path 1 = components[sectionName] direct lookup… All current sites hit this path"; component_selector "SELECTs is_dark_section into the struct but NEVER uses it in scoring".
- **what:** How a planned section becomes a component: Path 1 matches the section name directly against `content_components.function` (one active component per function — uniqueness index), which all current sites hit; Path 2, the scoring `component_selector` (suitable_site_types 0.35 + page_types 0.15 + quality 0.3 + specificity + usage), only runs for section_type names that aren't functions — and is scheme-blind. Consequences: there is no place to pick a scheme-appropriate variant for current sites (making a scheme-aware selector necessary-but-insufficient), and layout-aware section selection is explicitly documented future work (027 §10). page-rerender re-assembles stored HTML without re-selecting; only page-build-handler re-runs plan_sections.
- **sources:** idea.uk/running_notes_2(6).md (mmm/nnn corrections); idea.uk/REPORT_scheme_does_not_reach_components.md#2; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Step 7 findings)
- **relations:** scheme gap; tool-library (component registry/matching); flag_page_image_rebuild as the rebuild trigger.
- **verify-later:** plan_sections_action.go Path-1 comment; component_selector.go scoring.

### Coordinator result-extraction contract (resolveResultSpec) and the silent-stub class
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Coordinator result-extraction fix — field-validated 2026-06-19 (idea.uk index built + deployed)"; git archaeology settled the cause to commit 06a8c6ef (14 Jan), unchanged since.
- **what:** The class of bug and its structural fix. Bug: workflow `complete` steps declaring singular `output_field` were never honoured (coordinator read plural `output_fields` only since 14 Jan), falling back to a working-state dump; when a big multi-section page's dump cleared the 900k `MaxResultSizeBytes` cap, `extractMinimalResult` returned a `status:"completed"` **stub** — silent false completion (gamesdesign) or, where the claimed-item evidence gate refuses 0-component pages, honest claim-timeout failure (idea.uk's empty index). **Size was the trigger; the singular key necessary-but-not-sufficient** (bucket audit: 100 plural steps safe, ~59 dump-bucket agents fine because small, 4 singular — only the writer breaks). Fix: centralised result contract in `result_spec.go` — singular→FLATTEN (named field's contents become the response body), plural→FIELDS (unchanged), `output`→MAPPING (previously silently dumped), none→dump; completion metadata via setIfAbsent; **oversize is now a loud error** routed to notifyParentOfFailure with a per-field size breakdown; stub removed; deprecated keys alias to result_from/multiple_output_fields/result_mapping. Retest doctrine: requeue the one failed page, do NOT re-adopt.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Empty index diagnosis + fix-direction 1 + retest + git section); idea.uk/running_notes(63).md (ff–qq checkpoints); idea.uk/README_001_todo_list.md
- **relations:** claimed-item evidence gate; debugging trap "inferring writers from readers"; MaxResultSizeBytes guardrail (do not raise).
- **verify-later:** platform/orchestration/result_spec.go + coordinator.go; the mode=flatten log on a writer run.

### Claimed-item timeout evidence gate (failed vs false-completed)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "the evidence-gate (migration_claimed_item_timeout_evidence_v2, live since 2026-06-04) refuses to complete a 0-component page… That is the gate working as intended."
- **what:** The dispatch gate that distinguishes the two failure signatures of the same coordinator bug: without it a stubbed page false-completes (gamesdesign); with it, a 0-component page's claim is reset and retried until attempts exhaust → an honest `failed`. Used here as diagnostic doctrine: don't conflate a silent stub with a genuine handler hang — read the parent's collected_data response to tell them apart.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (empty-index diagnosis); idea.uk/016_debugging_guide_v2_32(1).md (claimed-item sections)
- **relations:** coordinator result contract; work-item state machine.
- **verify-later:** the evidence-check migration in the chassis migrations.

### Launch idioms: orchestrate vs work-item insert (and what each trigger does NOT do)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Confirmed from the production trigger scripts" (2026-06-20 iii); the 081c finding (no hand-rolled wrappers) cited.
- **what:** Two production ways work starts: (1) static agents are orchestrated by producing one Kafka message to system.agent.generic.requests (action=orchestrate, config.agent_type, full header set) via a one-off kcat pod; (2) dynamic handlers (page-build-handler etc.) cannot be orchestrated directly — INSERT a `site_work_items` row (status='triaged') and the running build-dispatch-loop claims and spawns them. Key caveat learned on idea.uk: the content triggers (rerender-pages / page-rerender / page-rebuild) never re-resolve composition — palette changes must go through needs_composition/needs_design. Deploy topology is likewise two-path: Go changes ship in the chassis image (roll agents to the tag), site HTML ships via the sites monorepo → Actions → B2.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Launch idioms); idea.uk/running_notes_checkpoint_tt.md (deploy topology + trigger mechanism); idea.uk/HANDOFF(13).md
- **relations:** composition re-resolve; scheduler/dispatch loop; debugging guide kcat sections.
- **verify-later:** 082_submit_domain_unified.sh; build-dispatch-loop definition.

### Fresh vs adoption entry paths converge on one cascade (fresh = adoption minus the crawl)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Capability map table with per-row verdicts ("already in fresh"); resolved empirically 2026-06-14 — a fresh submit flowed end-to-end through dispatch without manual triage.
- **what:** Two entry agents — domain-submitter (fresh: {domain,email,mission_brief}) and site-adoption-orchestrator (adopt: crawl→fingerprint→archetype→seeds) — converge on needs_domain_research and share the whole cascade (classifier read-and-extends adopted seeds → strategist → briefing → planner → composition → design → pages → rerender). The capability map shows the only adoption capabilities fresh lacks (CSS fingerprint, interactive-feature detection, full archetype) are inherently crawl-products; a new "fresh-build" single-agent copy was rejected as premature — reuse the existing path. The unified trigger `082_submit_domain_unified.sh` picks the entry (--from ⇒ adopt) and gained `--mission-file` (used to ship idea.uk's mission). The richest "seed with the existing setup" is adoption pointed at the live site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (capability map + submission entry points); idea.uk/HANDOFF(13).md (pipeline graph)
- **relations:** Phase 0 read; fidelity dial; adoption teardown vs fresh detach.
- **verify-later:** 082 script; site-adoption-orchestrator definition.

### Phase 0 classifier-only positioning read
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** "Phase 0 result (2026-06-14 — ran, with the live site up)": faithful identity/classification/content_direction/design_intent specs, interactive-platform at ~0.91 confidence.
- **what:** Running just domain-research-classifier on a domain as a near-zero-cost positioning brief before committing to a build — its four spec aspects ARE the answer to "what does this site do for a stranger?". Caveats codified: a fresh read is NOT blank-slate (the classifier scrapes the live site up to 3 pages unless an adoption already ran); a generic name yields a generic name-only read, so hiding the live site removes signal not bias; a safe suppression trick exists (temporary blank nginx `location = /` — never touch DNS/nginx wholesale with a live Stripe webhook); the classifier's terminal needs_strategy item will flow into a full build if dispatch is running. Decision: leave the live site up.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 0); idea.uk/HANDOFF_chassis_site.md
- **relations:** standing-ambition mission (the read was faithful-but-backward-looking, which motivated it); fresh vs adoption.
- **verify-later:** idea.uk site 97ed2f64 spec rows (incl. the duplicate-spec tidy-up).

### Standing-ambition default in the mission aspect (aspiration has no generated home)
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** "Go action carrying a default in the mission aspect… owner to finalise the principle wording (draft in notes)" — still an open backlog item in every later list.
- **what:** Diagnosis: the classifier is a current-state tool; mission_brief/roadmap_brief are the designated aspirational slots but owner-supplied (nothing generates them), and strategy runs after the classifier — so with the slots empty, fresh builds describe what exists instead of leading the field. Fix (framework, not hand-seeding): domain-submitter always writes a mission_brief = a fixed platform standing-ambition principle (lead the vertical; most useful forward-looking content; build around the site's distinctive tools; surpass don't mirror), merged with any owner mission, in a Go action — no new aspect (aspect list checked: no vision/ambition; free-text), no reorder. Design deliberately excluded: ambition lifts content via the LLM readers, but site-design-planner is deterministic and doesn't read mission prose — design leadership is its own track.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Setting direction before the build); idea.uk/HANDOFF_chassis_site.md (the two decisions); idea.uk/running_notes(63).md (w/x checkpoints)
- **relations:** build-standard migration (the classifier-side sibling); mission-file submission; Phase 0 read.
- **verify-later:** domain-submitter def for any standing-ambition merge (expect absent).

### Build-standard classifier migration (best-in-class quality/fit, not scope)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "READY, PROVEN-CORRECT, NOT YET APPLIED… replace() simulated against the live prompt → lands cleanly" (2026-06-21); still unapplied in TODO P2 (2026-06-26).
- **what:** A migration prepending a "Build standard" block to the classifier prompt: every build should aim at best-in-class **quality and fit** for its vertical — explicitly not scope inflation. Its first draft failed on a multi-line anchor (whitespace mangling) and was fixed to a single-line anchor with a rollback proven clean — feeding the prompt-migration discipline. Test plan: fresh build first; confirm an adopted rebuild stays faithful rather than drifting.
- **sources:** idea.uk/HANDOFF(13).md (Migrations status); idea.uk/TODO_chassis_and_idea_uk(1).md#P2; idea.uk/running_notes(63).md (lll/mmm 06-20/21)
- **relations:** standing-ambition default; prompt-migration discipline.
- **verify-later:** whether migration_classifier_build_standard.sql was later applied (file itself lives outside this unit).

### idea.uk mission and identity (workshop of tools; never verdicts; warm-paper identity)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "idea.uk mission REFRAMED away from the £29 tool… shipped as a file" and submitted via --mission-file 2026-06-21; classifier "read the mission well" on site 1244516d.
- **what:** The site's genuinely site-specific concept: idea.uk = the place to take an idea seriously — a growing workshop of genuinely good tools (the main event; free + paid), tools labelled **private (in-browser, nothing sent)** vs **AI/hosted** with private leading; cutting-edge succinct research-grounded guides; a news section; **never verdicts** — perspective, evidence, and questions framed as opinion (the legal reframe); the £29 verified report demoted to one specialised flagship tool; later a build-and-bring-to-market service; preserve the warm-paper/ink/single-rust-accent/Fraunces+IBM-Plex editorial identity. Noted honestly: the privacy and latest-research promises are stated intent the chassis can't yet enforce.
- **sources:** idea.uk/KEY_DOC_idea_uk_mission.txt; idea.uk/HANDOFF(13).md (idea.uk current state); idea.uk/running_notes(63).md (nnn/ooo 06-21)
- **relations:** liability framework (never-verdicts); mission-file mechanism; standing-ambition.
- **verify-later:** site 1244516d mission spec content.

### Fidelity dial — documented but not wired
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "The trigger records --fidelity but it doesn't yet modulate the build… doc 028's explicit fidelity input and its build_policy/adoption_meta aspect and per-item status are not yet wired (fidelity is currently implicit high)."
- **what:** A planned locked/high/medium/low fidelity input governing how faithfully a build reproduces its source/spec, flowing into a build_policy aspect with per-item planned/deployed status. Today the unified trigger records the flag and nothing reads it — a clean doc-vs-reality gap flagged repeatedly in the handoffs.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (launch idioms); idea.uk/HANDOFF(13).md
- **relations:** doc 028 (design/adoption); fresh vs adoption convergence.
- **verify-later:** any consumer of the fidelity field.

### Array-item field contract for the page-content-writer (item_fields fix)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "Prompt migration is already applied… But until a chassis image carrying the Go change is live, {{if .item_fields}} is always false… the applied prompt is inert on its own" (checkpoint tt, 2026-06-21).
- **what:** Root-cause class behind empty rendered sections (the 7 blank differentiator cards): the writer's prompt listed an array field with its type but never its per-element shape, so the model guessed item keys (title/body) that render empty against templates reading name/description. Fix has three coupled parts: plan_sections populates `ItemFields` on each llm_field_spec (Go); the prompt migration renders the exact per-item field list in both What-To-Write and the JSON skeleton (019_pcw_prompt_item_fields.sql — idempotent by sentinel, no broken intermediate state either deploy order); a render-time reconciler in v3_site_actions.go. Deploy order matters: chassis image first, then trigger.
- **sources:** idea.uk/019_pcw_prompt_item_fields.sql; idea.uk/running_notes_checkpoint_tt.md; idea.uk/README_assemble_bundle_idea_missing_sections.md (the bundled problem statement)
- **relations:** section-data reconciler; coordinator contract (sibling contract-mismatch class); diagnosis-loop bundles (the assemble-bundle invocation).
- **verify-later:** plan_sections_action.go ItemFields; whether the chassis tag carrying it shipped.

### Section-data reconciler and the human-sourced-field boundary
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "reconcile_section_data IS wired (registry.go L914… 're-trigger pages whose deferred section data is now query-resolvable')" — correcting an earlier stale "built but unwired" note (rr, 2026-06-19).
- **what:** Deferred section data (needs_section_data) is re-triggered when it becomes *query-resolvable*; the boundary concept: **human-sourced** spec fields (e.g. pricing tiers from site_specs.pricing) are not query-resolvable, so the reconciler can never fill them — either capture the data into specs (the £29 into pricing) or the section shouldn't be on the page. The unresolved-CTA gating (render no button when no eligible destination page exists) is the same honest-degradation family, tied to the thin 4-page plan having no hub pages.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + empty-index content gaps); idea.uk/README_001_todo_list.md
- **relations:** item_fields fix; site-plan thinness; content-governance (pricing spec).
- **verify-later:** reconcile_section_data_action.go host wiring; idea.uk pricing spec.

### Content-rebuild de-tools tool pages (confirmed hazard)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "confirmed hazard, fix pending" (TODO P3, 2026-06-26).
- **what:** A needs_page / link_resolution_rebuild on a tool or game page regenerates the page from plan_sections, and the plan knows nothing about the interactive tool living in a section's rendered_html — so the tool is silently replaced with generated prose. Fix direction: route link maintenance through a preserve-sections re-render path, stamp source_item_id, add an interactivity-aware save guard. Flagged as a direct risk to idea.uk's post-P0 rebuild if tools land first.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md#P3; idea.uk/running_notes_2(6).md (backlog)
- **relations:** tool pipeline (005/016b/020/026 cross-refs); page-rerender vs page-build-handler distinction.
- **verify-later:** whether the preserve-sections path landed.

### LLM API shape disciplines (server-tool injection, per-model thinking shapes, long-call timeouts)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Items 24–27 added to the debugging guide from the 2026-06-04 validation run ("three API bugs found and fixed during validation").
- **what:** Standing disciplines from live API breakage: (24) a hosted server tool may auto-inject its documented dependency — declaring it yourself collides (web_search v2 injects code_execution; the 400 names the conflict); (25) the same capability has different wire formats across model generations — newer Opus-class models take adaptive thinking + output_config.effort while Sonnet 4.6 takes manual budget_tokens, so helpers must branch per model (and Opus also rejects non-default temperature/top_p); (26) long agentic calls (high effort + N searches) send no headers for minutes — size client timeouts for the worst-case step (180s→900s; streaming is the durable answer); (27) always confirm the current request shape from live docs before coding, especially after a model bump — remembered shapes are guesses and each failed round-trip costs real spend.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (items 24–27); idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1 (acceptance); idea.uk/running_notes(63).md (2026-06-04 checkpoints)
- **relations:** engine upgrade; llm-quality-testing; model-infrastructure.
- **verify-later:** engine.go usesAdaptiveThinking + client timeout.

### Prompt/agent-definition migration discipline (snapshot, anchor, sentinel, file-not-paste)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Practised across three real migrations in this unit; the classifier one applied 2026-06-20; the paste failure and jsonb_set no-op both bitten and documented.
- **what:** The safe way to edit live agent prompts and specs in SQL: snapshot_agent() backup inside the same transaction (is_snapshot at version+1000; runtime selection excludes snapshots); UPDATE guarded to the live row only; exact-anchor replace() with a self-check that RAISEs and rolls back if an anchor is missing; idempotency sentinels so re-runs no-op (blind replace would double-expand); single-line anchors (multi-line anchors broke on whitespace); run migration FILES via `kubectl … < file` — pasting into psql mangles \set/\echo/blank lines and once left an open transaction. Companion jsonb facts: jsonb_set into a missing parent silently no-ops — use `||` to add top-level keys; site_specs jsonb column is `data`; partial UNIQUE (site_id,aspect) WHERE is_current.
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql (header); idea.uk/019_pcw_prompt_item_fields.sql (idempotency note); idea.uk/HANDOFF(13).md (schemas + operating rules)
- **relations:** build-standard migration (anchor-bug case); snapshot/revert machinery (docs014/016 6.1).
- **verify-later:** snapshot_agent function; bak_* conventions.

### Running-notes journal + distilled HANDOFF discipline (memory-off cross-session state)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** running_notes.md grew to ~5,690 lines / checkpoints (a)…(kkk) then archived into running_notes_2.md ("Present this file at the END OF EVERY TURN"); HANDOFF refreshed per session and marked "canonical cold-start doc".
- **what:** The working method that produced this whole unit: memory is off, so an append-only checkpoint journal is the cross-session record (dated checkpoints with lettered ids, CORRECTION entries that supersede earlier readings, decision logs), paired with a distilled HANDOFF kept fresh (current state, strict user preferences, schemas, backlog) and per-thread cold-start briefs (HANDOFF_scheme_to_components) that name exactly what to attach and read in what order. Includes the archival pattern (part 1 frozen, part 2 carries a CARRY-OVER STATE header) and the checkpoint-tt pattern of appending a prepared block.
- **sources:** idea.uk/running_notes_2(6).md (header); idea.uk/HANDOFF(13).md (header); idea.uk/HANDOFF_scheme_to_components(1).md
- **relations:** docs037 travelling-docs conventions; bundle packagers.
- **verify-later:** n/a (process concept).

### Docubundle context packagers + curated attach-lists for fresh threads
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Packager scripts "fixed for the real (messy) repo layout" 2026-06-10; two bundles produced (go-live 596KB, chassis-engine 1.5MB context files present).
- **what:** Self-contained bash packagers that assemble a single context file for an AI assistant per task — the idea.uk go-live bundle (Go code + embedded page + deploy + the go-live docs; explicitly no live capture because there's no DB/k8s) and the chassis idea-engine bundle (the engine to port + the chassis framework to build it in, action catalogue for reuse-discovery). Copes with the messy folder by resolving docs to the newest "(N)" variant by mtime and dropping *.orig* backups and binaries. Complemented by hand-written attach-lists (BUNDLE_1/2, CONTEXT_PACK, CONTEXT_FOR_NEXT_CHAT) that spell out which files a fresh thread needs and warn the idea.uk files are NOT in the chassis project.
- **sources:** idea.uk/docubundle_idea_golive/package_idea_uk_golive.sh (header); idea.uk/docubundle_idea_within_chassis/package_chassis_idea_engine(3).sh (header); idea.uk/BUNDLE_1_idea_uk_golive.md; idea.uk/CONTEXT_FOR_NEXT_CHAT.md
- **relations:** diagnosis-loop bundle tooling (cmd/bundle in README_assemble_bundle); running-notes discipline.
- **verify-later:** n/a.

### Training checkpoint/adapter upload to B2 via pre-minted presigned single-object PUTs
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Status: Phase A (02_train) BUILT 2026-06-05, isolation test pending. Phases B–D not built." — with the isolation-test harness and its artefacts present in nginx/_iso/.
- **what:** Solves three coupled gaps of ephemeral Thunder training VMs (final adapter not durable; the monitor's DONE_OK path decommissions the disk; no checkpoints/resume on a ~24h run): the launcher pre-mints single-object, write-only presigned PUT URLs and hands them to the VM in a manifest — the hostile-VM threat model rejects standing scoped keys and callback endpoints (nothing on the box can mint or write beyond the fixed URL set). Explicit framing: this protects access, not artefact integrity — a malicious-but-valid adapter still needs the flywheel-D eval gate before promotion. Phase A validated box-free by isolation_test_phase_a.py (presign/PUT signature, tar round-trip, checkpoints/ exclusion, GET+extract byte-identical).
- **sources:** idea.uk/nginx/PLAN_checkpoint_and_artefact_upload_b2(1).md; idea.uk/nginx/isolation_test_phase_a.py (header); idea.uk/nginx/README_get_b2_details.md
- **relations:** Thunder adapter; B2 dead-drop (same one-way pattern); model-infrastructure eval gate.
- **verify-later:** 02_train upload hooks; thunder-training-monitor-worker decommission path.

### Hosting split: static-serverless front + small always-on backend ("static front + small back end")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Architecture doc §3 and the live topology (page embedded in the binary on the VM; B2 for everything chassis-built).
- **what:** The hosting taxonomy idea.uk established for the platform: pure-static content sites are serverless on B2; anything running a minutes-long multi-LLM job with a payment webhook cannot be serverless or edge-shaped — it needs a small always-on service with a stable inbound address. The classifier's `build_approach: hybrid` / `hosting_trajectory: needs_server` fields are the framework's slot for this distinction (noted as not yet confirmed in the classifier output). This is the hinge Layer 5 eventually automates, and why the engine can never be a forked client-side tool component.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#3; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 2 hosting reality); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (reuse section)
- **relations:** Layer-5 wrapper; VM cutover; tool-library boundary.
- **verify-later:** where build_approach/hosting_trajectory actually live (strategist? architecture doc concept only?).

### idea.uk chassis-site build state (two site rows; staging-only; gated go-live)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** Current position 2026-06-26: composition + stylesheet correct and verified (tool-portal-light + parchment, commit 05ef817, no LLM drift); pages still dark; rebuild + review + cutover gated on the P0 scheme thread.
- **what:** The concrete case-study state: first chassis run under site 97ed2f64 (2026-06-14: classifier→…→empty index → coordinator fix validated on it); old chassis site torn down and resubmitted fresh 2026-06-21 as site 1244516d with the mission file; re-resolved onto tool-portal-light 2026-06-25; deployed page defects catalogued (empty differentiators, unresolved CTAs, dead contact form, missing pricing spec, thin nav/footer, empty meta description, dark chrome). The live £29 VM tool is a separate stream, untouched and earning throughout — the safety property the whole thread leans on.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + current position); idea.uk/HANDOFF(13).md
- **relations:** every design/pipeline concept above (this is their test case); VM cutover.
- **verify-later:** current state of site 1244516d; whether the P0 thread completed and cutover happened.
