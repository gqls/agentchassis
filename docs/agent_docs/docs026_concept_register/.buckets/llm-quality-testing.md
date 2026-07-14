
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model quality assessment: local 70B comparable for some tasks
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** 009 test table dated 2026-03-24; 023 raw comparative transcripts
- **what:** Llama 3.3 70B (single H100, num_ctx 8192) scores 8-9/10 vs Claude for classification/content, 7/10 design; Mistral Small 3 CPU adequate only for low-stakes structured tasks (5/10 classification, 3/10 design). Evaluation criteria captured in 023: JSON parse w/o fences, exact field names, specific headlines, action-verb CTAs, no invented claims. ThunderCompute quirks: 2-GPU instances broken, num_ctx metadata bug, KEEP_ALIVE=-1.
- **sources:** 009#Model Quality Assessment, #ThunderCompute Notes; 023 full
- **relations:** fine-tuning path; LLM tiering
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### LLM reliability strategy for component generation (observability first, shrink the contract second)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** "Track 1 — Make rejection observable. Done in this iteration" (2026-05-11); tracks 2-3 open
- **what:** LLMs are structurally good but unreliable at exact schema↔template list reconciliation (bookkeeping, not creativity). Strategy: (1) pre-store validator writes structured rejections to agent_error_log — done; (2) move bookkeeping out of the LLM: inject the root section wrapper at store time, declare Tier D sub-schemas centrally in queryresolve, optionally derive schema keys from the template parser; (3) prompt/model tweaks only after patterns are visible. Explicitly rejected: silent auto-correction at the validator; accumulating hand-written components without addressing the prompt.
- **sources:** FOCUS_llm_reliability_for_component_generation.md (whole)
- **relations:** validation gates; Tension #1 (same "don't trust LLM formal labels" family); tiered field classification
- **verify-later:** agent_error_log rejection rows; whether 2a/2b landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Verification harness (build + ops)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.3 "Build side is easy … the ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work"
- **what:** Build-check / test-runner / validator / canary / rollback expressed as actions/adapters, checking output against ground truth. The build side reuses existing validate→regenerate; the ops side (canary, infra rollback, detection) is the thinnest, most-new part.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** toolchain validator (self-dev pipeline); lifecycle map
- **verify-later:** go build/vet/test runner; canary/rollback adapters (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### LLM model governance: aliases, per-step model choice, llm_call_log
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** v2/027 regex-replaces all dated model names with aliases ("only the alias resolver in code needs updating"); 040 upgrades planners to sonnet with rationale and creates llm_call_log.
- **what:** Conventions for model management across ~90 agent definitions: model aliases (claude-sonnet-4-5 not dated strings) resolved in code; deliberate per-step model tiering (haiku for cheap classification, sonnet for high-leverage planning, opus for plan_site and tool recreation); llm_call_log capturing calls for cost analysis and training data. 067 (filename: "not_yet_implemented") prepared extended-thinking budget_tokens for classifier/planner gated on a Go patch.
- **sources:** sql_for_agents_v2/027_replace_claude_model_names.sql; 040_optimise_which_llms.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** finetuning flywheel (llm_call_log flywheel columns in 085); ai_endpoint_health
- **verify-later:** alias resolver; whether extended thinking was ever enabled

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel D — Claude vs local-model quality eval (replay harness)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.4d "paused"; partial results table "Case 1 … 27 min … not a practical substrate for production-scale replay-eval"; §2.4d-comparison methodology "run after eval completes"
- **what:** A replay-not-rerun eval: pull 20 diverse stored production prompts (`DISTINCT ON (orchestration_id)`) from `llm_call_log`, POST each to a local Ollama model, and compare against the stored Claude response across three levels (structural jq checks, Claude-as-judge, manual review). Target agent was `page-content-writer/iter_0_generate_content`. Stalled on shared-adapter CPU contention, prompting the dedicated `ollama-eval` pod (24Gi/28Gi) and the GPU-substrate argument.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4d, #2.4d-comparison, #14 (Eval and replay methodology)
- **relations:** provides the ROI justification and the eval gate for promoting fine-tuned adapters; blocks enabling model swap
- **verify-later:** ollama-eval deployment; llm_call_log; results.jsonl runner

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model quality assessment: local 70B comparable for some tasks
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** 009 test table dated 2026-03-24; 023 raw comparative transcripts
- **what:** Llama 3.3 70B (single H100, num_ctx 8192) scores 8-9/10 vs Claude for classification/content, 7/10 design; Mistral Small 3 CPU adequate only for low-stakes structured tasks (5/10 classification, 3/10 design). Evaluation criteria captured in 023: JSON parse w/o fences, exact field names, specific headlines, action-verb CTAs, no invented claims. ThunderCompute quirks: 2-GPU instances broken, num_ctx metadata bug, KEEP_ALIVE=-1.
- **sources:** 009#Model Quality Assessment, #ThunderCompute Notes; 023 full
- **relations:** fine-tuning path; LLM tiering
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### LLM reliability strategy for component generation (observability first, shrink the contract second)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** "Track 1 — Make rejection observable. Done in this iteration" (2026-05-11); tracks 2-3 open
- **what:** LLMs are structurally good but unreliable at exact schema↔template list reconciliation (bookkeeping, not creativity). Strategy: (1) pre-store validator writes structured rejections to agent_error_log — done; (2) move bookkeeping out of the LLM: inject the root section wrapper at store time, declare Tier D sub-schemas centrally in queryresolve, optionally derive schema keys from the template parser; (3) prompt/model tweaks only after patterns are visible. Explicitly rejected: silent auto-correction at the validator; accumulating hand-written components without addressing the prompt.
- **sources:** FOCUS_llm_reliability_for_component_generation.md (whole)
- **relations:** validation gates; Tension #1 (same "don't trust LLM formal labels" family); tiered field classification
- **verify-later:** agent_error_log rejection rows; whether 2a/2b landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Verification harness (build + ops)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.3 "Build side is easy … the ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work"
- **what:** Build-check / test-runner / validator / canary / rollback expressed as actions/adapters, checking output against ground truth. The build side reuses existing validate→regenerate; the ops side (canary, infra rollback, detection) is the thinnest, most-new part.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** toolchain validator (self-dev pipeline); lifecycle map
- **verify-later:** go build/vet/test runner; canary/rollback adapters (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### LLM model governance: aliases, per-step model choice, llm_call_log
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** v2/027 regex-replaces all dated model names with aliases ("only the alias resolver in code needs updating"); 040 upgrades planners to sonnet with rationale and creates llm_call_log.
- **what:** Conventions for model management across ~90 agent definitions: model aliases (claude-sonnet-4-5 not dated strings) resolved in code; deliberate per-step model tiering (haiku for cheap classification, sonnet for high-leverage planning, opus for plan_site and tool recreation); llm_call_log capturing calls for cost analysis and training data. 067 (filename: "not_yet_implemented") prepared extended-thinking budget_tokens for classifier/planner gated on a Go patch.
- **sources:** sql_for_agents_v2/027_replace_claude_model_names.sql; 040_optimise_which_llms.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** finetuning flywheel (llm_call_log flywheel columns in 085); ai_endpoint_health
- **verify-later:** alias resolver; whether extended thinking was ever enabled

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel D — Claude vs local-model quality eval (replay harness)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.4d "paused"; partial results table "Case 1 … 27 min … not a practical substrate for production-scale replay-eval"; §2.4d-comparison methodology "run after eval completes"
- **what:** A replay-not-rerun eval: pull 20 diverse stored production prompts (`DISTINCT ON (orchestration_id)`) from `llm_call_log`, POST each to a local Ollama model, and compare against the stored Claude response across three levels (structural jq checks, Claude-as-judge, manual review). Target agent was `page-content-writer/iter_0_generate_content`. Stalled on shared-adapter CPU contention, prompting the dedicated `ollama-eval` pod (24Gi/28Gi) and the GPU-substrate argument.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4d, #2.4d-comparison, #14 (Eval and replay methodology)
- **relations:** provides the ROI justification and the eval gate for promoting fine-tuned adapters; blocks enabling model swap
- **verify-later:** ollama-eval deployment; llm_call_log; results.jsonl runner
