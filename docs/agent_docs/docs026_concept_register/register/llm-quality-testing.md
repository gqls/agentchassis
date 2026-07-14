# Register — llm-quality-testing

5 concepts, consolidated from 5 raw extractions across units U01, U02, U17a,
U18, U24b. No exact duplicates within this category's raw material, though
LQT-001 overlaps substantially with a model-infrastructure raw block ("Model
quality assessment & per-agent model assignment", from the same underlying
benchmark) and LQT-005 overlaps heavily with several finetuning-flywheel
entries covering the same Flywheel-D methodology in more granular detail —
both left in their originally-tagged category per the per-category assignment
rule, cross-referenced here.

### LQT-001 — Model quality assessment: local 70B comparable for some tasks
- **status:** deployed
- **status-evidence:** dated test table (2026-03-24) plus raw comparative transcripts corroborating the scores.
- **what:** Llama 3.3 70B (single H100, num_ctx 8192) scores 8-9/10 vs Claude for classification/content, 7/10 for design; Mistral Small 3 on CPU is adequate only for low-stakes structured tasks (5/10 classification, 3/10 design). Evaluation criteria: JSON parses without fences, exact field names, specific headlines, action-verb CTAs, no invented claims. The same benchmark also drove a recommended per-agent assignment (strategist/webdesign/planner→Claude, classifier/content-writer/triage→Llama 70B GPU, briefing→Mistral CPU) with a ~95%-at-2000-domains cost projection ($910-990 vs $15-30k all-Claude). ThunderCompute quirks noted: 2-GPU instances broken, num_ctx metadata bug, KEEP_ALIVE=-1.
- **sources:** 009_model_infrastructure.md#Model Quality Assessment, #ThunderCompute Notes; 023 (full comparative transcripts); old/older1/020c_gpu_and_model_infrastructure_v3.md#model-quality-assessment, #cost-projection
- **relations:** Fine-tuning path (finetuning-flywheel); LLM tiering (model-infrastructure); model swap/snapshot/revert (model-infrastructure, the deployment mechanism for any resulting assignment); "Model quality assessment & per-agent model assignment" (model-infrastructure, same benchmark data)
- **verify-later:** —

### LQT-002 — LLM reliability strategy for component generation (observability first, shrink the contract second)
- **status:** partial
- **status-evidence:** "Track 1 — Make rejection observable. Done in this iteration" (2026-05-11); tracks 2-3 open.
- **what:** LLMs are structurally good but unreliable at exact schema-to-template-list reconciliation (bookkeeping, not creativity). Strategy: (1) pre-store validator writes structured rejections to agent_error_log — done; (2) move bookkeeping out of the LLM: inject the root section wrapper at store time, declare Tier D sub-schemas centrally in queryresolve, optionally derive schema keys from the template parser; (3) prompt/model tweaks only after patterns are visible. Explicitly rejected: silent auto-correction at the validator, and accumulating hand-written components without addressing the prompt.
- **sources:** FOCUS_llm_reliability_for_component_generation.md (whole)
- **relations:** validation gates; tiered field classification; LLM step config shadowing bug (model-infrastructure, same "don't trust LLM formal labels" family)
- **verify-later:** agent_error_log rejection rows; whether tracks 2a/2b landed

### LQT-003 — Verification harness (build + ops)
- **status:** partial
- **status-evidence:** "Build side is easy … the ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work."
- **what:** Build-check / test-runner / validator / canary / rollback expressed as actions/adapters, checking output against ground truth. The build side reuses existing validate→regenerate machinery; the ops side (canary, infra rollback, detection) is the thinnest, most-net-new part of the design.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, #8.2
- **relations:** toolchain validator (self-dev pipeline); lifecycle map
- **verify-later:** go build/vet/test runner; canary/rollback adapters (proposed)

### LQT-004 — LLM model governance: aliases, per-step model choice, llm_call_log
- **status:** deployed
- **status-evidence:** a SQL migration regex-replaces all dated model names with aliases ("only the alias resolver in code needs updating"); a later migration upgrades planners to sonnet with rationale and creates llm_call_log.
- **what:** Conventions for model management across ~90 agent definitions: model aliases (`claude-sonnet-4-5`, not dated strings) resolved in code; deliberate per-step model tiering (haiku for cheap classification, sonnet for high-leverage planning, opus for plan_site and tool recreation); llm_call_log capturing calls for cost analysis and training data. One migration (filename flagged "not_yet_implemented") prepared extended-thinking budget_tokens for classifier/planner, gated on a Go patch.
- **sources:** sql_for_agents_v2/027_replace_claude_model_names.sql; 040_optimise_which_llms.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** finetuning flywheel (llm_call_log flywheel columns); ai_endpoint_health (model-infrastructure); Model aliases and the model selection strategy (model-infrastructure); Extended thinking configuration (model-infrastructure)
- **verify-later:** alias resolver; whether extended thinking was ever enabled

### LQT-005 — Flywheel D — Claude vs local-model quality eval (replay harness)
- **status:** partial
- **status-evidence:** "paused" per the FOCUS doc's own status table; partial results recorded ("Case 1 … 27 min … not a practical substrate for production-scale replay-eval"); a comparison methodology section notes "run after eval completes".
- **what:** A replay-not-rerun eval: pull 20 diverse stored production prompts (`DISTINCT ON (orchestration_id)`) from `llm_call_log`, POST each to a local Ollama model, and compare against the stored Claude response across three levels (structural jq checks, Claude-as-judge, manual review). Target agent was `page-content-writer/iter_0_generate_content`. Stalled on shared-adapter CPU contention, prompting a dedicated `ollama-eval` pod (24Gi/28Gi) and the GPU-substrate argument that eventually superseded the CPU attempt. The finetuning-flywheel register carries this same methodology in much greater granularity (replay-eval methodology, the three-level L1/L2/L3 pipeline, the held-out eval set, Claude-as-judge bias handling, and the iter_0 verdict) — read those entries for the full mechanics; this entry is the quality-testing-category anchor for the same lane.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4d, #2.4d-comparison, #14 (Eval and replay methodology)
- **relations:** provides the ROI justification and the eval gate for promoting fine-tuned adapters; blocks enabling model swap; Flywheel D replay-eval methodology, Three-level evaluation pipeline, Held-out eval set v1, Claude-as-judge, iter_0 verdict (all finetuning-flywheel)
- **verify-later:** ollama-eval deployment; llm_call_log; results.jsonl runner
