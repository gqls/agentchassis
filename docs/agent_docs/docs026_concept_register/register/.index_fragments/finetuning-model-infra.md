| FTW-001 | Finetuning flywheel four-lane programme (A/B/C/D) | partial | The A/B/C/D flywheel: export, RAG, LoRA training, Claude-vs-local eval | finetuning-flywheel.md |
| FTW-002 | Three compounding improvement channels | partial | RAG, LoRA, and prompt-variant A/B as three independent compounding quality levers | finetuning-flywheel.md |
| FTW-003 | Fine-tuning path (log→export→LoRA→GGUF→Ollama→swap) | partial | Full pipeline shape; last-mile production wiring explicitly still outstanding | finetuning-flywheel.md |
| FTW-004 | llm_call_log as ops visibility + training-data capture | deployed | Every LLM call logged fire-and-forget with flywheel columns for training exports | finetuning-flywheel.md |
| FTW-005 | Training-data export as chassis agent + action | deployed | training_data_export action/worker/orchestrator, v1→v3.2 evolution | finetuning-flywheel.md |
| FTW-006 | training_exports Postgres schema | deployed | Named, versioned snapshot datasets in Postgres, not ephemeral JSONL files | finetuning-flywheel.md |
| FTW-007 | ChatML export format with metadata sidecar | deployed | Row shape: chat messages + metadata JSONB back-linking to llm_call_log | finetuning-flywheel.md |
| FTW-008 | Response cleaning + SFT negative-example exclusion | deployed | Strip fences, exclude prose edge-cases from SFT, reserve them for future DPO | finetuning-flywheel.md |
| FTW-009 | `<no value>` contamination + iter_1 filter floor | partial | Prompt-builder bug polluted iter_0 data; fix-date filter never recorded | finetuning-flywheel.md |
| FTW-010 | Dataset profile/schema heterogeneity iter_0 | deployed | One training slice spans 3 component JSON schemas; anchors max_seq choice | finetuning-flywheel.md |
| FTW-011 | Flywheel C training pipeline (scripts 00-03) | deployed | Manual Unsloth QLoRA scripts with a smoke-gates-full discipline | finetuning-flywheel.md |
| FTW-012 | Base-model decision: Llama 3.3 70B QLoRA | deployed | 70B chosen for available hardware; 8B ablation planned but never run | finetuning-flywheel.md |
| FTW-013 | iter_0 baseline training run | deployed | Real cost/time/loss anchors: ~9h, ~$20, final_loss 0.27 | finetuning-flywheel.md |
| FTW-014 | GPU environment version pinning (cu124 stack) | deployed | Narrow torch/transformers/flash-attn pin set required for training to work | finetuning-flywheel.md |
| FTW-015 | Snapshot economics: setup script beats VM snapshots | deployed | Thunder snapshots uneconomic below ~18 runs/month; use idempotent setup script | finetuning-flywheel.md |
| FTW-016 | GPU training performance model | deployed | Smoke rate ≠ steady state; FA2 lever; O(N²) attention cost of seq length | finetuning-flywheel.md |
| FTW-017 | fp16 adapter save decision | aspirational | One-line fix to halve adapter size (fp32→fp16); agreed but never shipped | finetuning-flywheel.md |
| FTW-018 | Flywheel D replay-eval methodology + CPU-eval-pod evolution | partial | Replay stored prompts against candidate model; CPU attempt superseded by GPU | finetuning-flywheel.md |
| FTW-019 | Three-level evaluation pipeline (L1/L2/L3) | deployed | Structural checks, Claude-as-judge, and spot-check folded into one report | finetuning-flywheel.md |
| FTW-020 | Held-out eval set v1 | deployed | Canonical 20-case (of 50) comparison set reused across all iterations | finetuning-flywheel.md |
| FTW-021 | Claude-as-judge anonymised A/B + self-recognition bias | deployed | Judge design controls for position and self-recognition bias, empirically observed | finetuning-flywheel.md |
| FTW-022 | iter_0 verdict: shippable for low-stakes | deployed | Matches Claude on JSON/schema; voice fidelity is the main iter_1 lever | finetuning-flywheel.md |
| FTW-023 | Fine-tuning candidate selection/prioritisation | aspirational | High-volume structured-JSON agents ranked as swap candidates | finetuning-flywheel.md |
| FTW-024 | model_lifecycle schema | deployed | training_runs/artefacts/evaluations/deployable_adapters lifecycle namespace | finetuning-flywheel.md |
| FTW-025 | Eval gate before promotion | partial | Human deployment_decision required; also the integrity boundary for uploads | finetuning-flywheel.md |
| FTW-026 | Flywheel C phase-2 automation architecture (evolution) | superseded | Three generations: HTTP job server → SSH-exec → adapter dispatch | finetuning-flywheel.md |
| FTW-027 | model-trainer orchestration chain / Phase 5 kickoff | deployed | Orchestrator spawns data-preparer→provisioner→launcher over Kafka/saga | finetuning-flywheel.md |
| FTW-028 | training-launcher real workflow | deployed | presign→manifest→detached SSH launch→mark_running, full path ~26s | finetuning-flywheel.md |
| FTW-029 | setsid detached launch + false-success gap | deployed | ssh_exec returns immediately; exit_code 0 doesn't prove the training started | finetuning-flywheel.md |
| FTW-030 | run.sh launch chain + RUN_SH markers | deployed | Grep-able marker protocol; RUN_SH_DONE now implies durable upload | finetuning-flywheel.md |
| FTW-031 | Scripts bundle in B2 as deployment unit | deployed | Re-uploading bundle.tar.gz IS the deploy; must stay flat, no DB change | finetuning-flywheel.md |
| FTW-032 | Checkpoint & final-adapter durability via presigned PUT manifest | partial | Hostile-VM-safe upload via pre-minted URLs; O(K²) loop retired via batch presign | finetuning-flywheel.md |
| FTW-033 | CheckpointUploader trainer callback | deployed | Best-effort checkpoint upload, hard-gated final adapter upload | finetuning-flywheel.md |
| FTW-034 | Resume path | partial | Relaunch auto-resumes from highest B2 checkpoint; unproven in prod | finetuning-flywheel.md |
| FTW-035 | Monitor enablement gate: DONE must mean durable | partial | Monitor stays disabled until upload path proven, to avoid destroying adapters | finetuning-flywheel.md |
| FTW-036 | knowledge_base RAG store + Flywheel B verification | deployed | Lane-B deployment/verification record; full mechanism lives in RAGK-001 | finetuning-flywheel.md |
| FTW-037 | Nomic task prefixes load-bearing | deployed | Missing search_query/search_document prefixes silently broke ranking 5x | finetuning-flywheel.md |
| FTW-038 | Thunder instance lifecycle: reaper + training monitor | deployed | SQL-migration-level summary of the reaper/monitor pair as a unit | finetuning-flywheel.md |
| FTW-039 | LLM fallback extraction doubling as training data (med pricing) | deployed | CPU Mistral price-extraction fallback incidentally accrues fine-tune data | finetuning-flywheel.md |
| FTW-040 | Thunder adapter (credential-boundary GPU provisioning) | deployed | Ephemeral credential-free VMs, spend/uptime/concurrency caps, 15-min reaper | finetuning-flywheel.md |
| FTW-041 | Text LoRA — veterinary knowledge extractor | aspirational | Deferred recipe to fine-tune a local 7-8B knowledge-extractor model | finetuning-flywheel.md |
| MDL-001 | Model aliases and the model selection strategy | deployed | Short aliases resolved in code; sonnet/haiku/opus/ollama per-step defaults | model-infrastructure.md |
| MDL-002 | Agent model-assignment upgrade sweeps (migration 081) | deployed | Targeted UPDATEs upgrade agent model refs; stale claude-3.x replaced globally | model-infrastructure.md |
| MDL-003 | Ollama adapter (CPU embeddings + local classification) | partial | CPU provider for embeddings + small-model classification, same AIService interface | model-infrastructure.md |
| MDL-004 | RAG pipeline deployment bundle | deployed | The original rollout PR that added ollama-adapter + rag actions + migrations | model-infrastructure.md |
| MDL-005 | ai_endpoint_health: multi-endpoint model routing / GPU scheduler | deployed | Single table is the GPU scheduler; healthy→flow, unhealthy→back-to-triage | model-infrastructure.md |
| MDL-006 | Model swap/snapshot/revert control plane (migration 083) | deployed | snapshot/swap/revert functions; agent_definitions is the routing control plane | model-infrastructure.md |
| MDL-007 | LLM tiering + cluster-then-slot-fill scaling pattern | aspirational | Per-call-site llm_tier annotation mapped to endpoint via flippable config | model-infrastructure.md |
| MDL-008 | Ollama CPU adapter operational envelope | deployed | Recreate strategy, load timeouts, memory headroom rule, measured throughput | model-infrastructure.md |
| MDL-009 | thunder-training-monitor | partial | Per-instance SSH probe classifying ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN | model-infrastructure.md |
| MDL-010 | Monitor/reaper responsibility split | deployed | Reaper is a dependency-free cost backstop; monitor depends on adapter+SSH | model-infrastructure.md |
| MDL-011 | Thunder Compute adapter (provision/decommission lifecycle) | deployed | Kafka adapter wrapping Thunder API; provision/decommission mechanics + API gotchas | model-infrastructure.md |
| MDL-012 | thunder-reaper scheduled task + uptime deadline | deployed | 15-min task decommissions instances past max_uptime_hours; deadline is ours | model-infrastructure.md |
| MDL-013 | Thunder spend gating (DB-side check) | deployed | thunder_provision_check view enforces a rolling daily spend cap before create | model-infrastructure.md |
| MDL-014 | Orphan-sweep for stale thunder_instances rows | aspirational | Agreed but unbuilt design to reconcile DB rows against Thunder's live list | model-infrastructure.md |
| MDL-015 | Adapter-managed SSH access (ed25519 in k8s Secrets) | deployed | Per-provision keypair, deterministic Secret naming, ubuntu not root login | model-infrastructure.md |
| MDL-016 | Thunder Prototyping vs Production mode economics | partial | Prototyping proven fine for inference; untested for long training runs | model-infrastructure.md |
| MDL-017 | Thunder checkpoint upload + O(K²) batch-presign retirement | partial | Presigned PUT manifest for durability; batch presign fixed an O(K²) slowdown | model-infrastructure.md |
| MDL-018 | Anthropic client temperature parameter removed unconditionally | superseded | Opus 4.7+ rejects any non-default temperature; client omits it on every call | model-infrastructure.md |
| MDL-019 | GPU/AI-endpoint scheduling design evolution (superseded) | superseded | Historical four-option debate resolved into the single ai_endpoint_health table | model-infrastructure.md |
| MDL-020 | agent_definitions backup naming convention | superseded | _preNNN suffix ties a backup to its guarding migration; never-drop rule | model-infrastructure.md |
| MDL-021 | Code-context retrieval infrastructure (analyser adapter) | deployed | In-cluster code indexing into a pgvector code_symbols table; found stale | model-infrastructure.md |
| MDL-022 | LLM step config shadowing bug | partial | Top-level ai_service silently shadows step-level model/max_tokens overrides | model-infrastructure.md |
| MDL-023 | Extended thinking configuration | deployed | budget_tokens enables Anthropic extended thinking; strips temperature | model-infrastructure.md |
| MDL-024 | Static vs dynamic agent deployment + GPU cost strategy | aspirational | Dynamic spawned GPU workers claimed as 95% cheaper than static GPU deployment | model-infrastructure.md |
| MDL-025 | Model-tiering by task ("the 3B problem") | aspirational | Frontier models for synthesis only; tiny specialised models for extraction | model-infrastructure.md |
| MDL-026 | Self-hosted LLM inference (vLLM/GPU at scale) | aspirational | Plan to serve 7B models via vLLM continuous batching to escape API cost | model-infrastructure.md |
| MDL-027 | RAG best practices (superseded v1) | superseded | Older doc superseded by the v2 best-practices doctrine (see RAGK-004) | model-infrastructure.md |
| MDL-028 | Model quality assessment & per-agent model assignment | partial | Benchmark data behind the Claude/Llama70B/Mistral per-agent routing table | model-infrastructure.md |
| MDL-029 | Flywheel C — LoRA fine-tuning path & iter0 adapter output | deployed | Training pipeline + the first closed-out adapter artefact (828MB) | model-infrastructure.md |
| MDL-030 | Flywheel C Phase 2 — HTTP-job-server automation | abandoned | First-generation automation design, superseded before being built | model-infrastructure.md |
| MDL-031 | Phase 5 training-launcher + model-trainer chain | deployed | Real launcher (migration 102) driven by the model-trainer orchestrator | model-infrastructure.md |
| MDL-032 | setsid detached launch command | deployed | Training launch backgrounded via setsid so ssh_exec returns immediately | model-infrastructure.md |
| MDL-033 | run.sh RUN_SH markers + set -e durability hard-gate | deployed | Marker protocol lets DONE imply "trained AND uploaded" | model-infrastructure.md |
| MDL-034 | iter0 pre-trigger + Phase B/C/D deploy runbooks | deployed | Operational runbooks staging the launcher and checkpoint-upload rollout | model-infrastructure.md |
| MDL-035 | Per-workflow-step model routing (data-sovereignty) | deployed | Three-tier ai_service resolution lets any step stay in-cluster | model-infrastructure.md |
| MDL-036 | Text-provider wiring reality (two providers end-to-end) | deployed | Only Anthropic and Ollama actually work end-to-end for text | model-infrastructure.md |
| MDL-037 | llama3.3:70b trained but never used for inference | partial | Completed training run never wired to any agent's production ai_service | model-infrastructure.md |
| RAGK-001 | RAG knowledge_base (shared pgvector store) | deployed | Shared table, vector(768), ivfflat+trigram fallback, SHA256 dedup | rag-knowledge-base.md |
| RAGK-002 | rag_lookup action (vector search + trigram fallback) | partial | Embeds query, cosine search within collection, trigram fallback when Ollama down | rag-knowledge-base.md |
| RAGK-003 | rag_index action (chunk, embed, dedup, store) | partial | Chunks + SHA256 dedup + embed + insert; stores even if embedding fails | rag-knowledge-base.md |
| RAGK-004 | RAG best practices — filter-first, quality gating, token budget | aspirational | Filter by metadata before ranking; 20-30% context budget; nomic task prefixes | rag-knowledge-base.md |
| RAGK-005 | knowledge-indexer agent (deferred) | aspirational | Deliberately unbuilt owning-agent; actions suffice until a use case demands it | rag-knowledge-base.md |
| RAGK-006 | Concept-document RAG for content writers (v2+) | aspirational | Deferred design to ingest full concept docs into knowledge_base for v2+ | rag-knowledge-base.md |
| LQT-001 | Model quality assessment: local 70B comparable for some tasks | deployed | Llama 70B near-parity with Claude on classification/content; Mistral weak | llm-quality-testing.md |
| LQT-002 | LLM reliability strategy for component generation | partial | Observability first, then shrink the LLM's bookkeeping contract | llm-quality-testing.md |
| LQT-003 | Verification harness (build + ops) | partial | Build-check/validator/canary/rollback; ops side is the thinnest part | llm-quality-testing.md |
| LQT-004 | LLM model governance: aliases, per-step model choice, llm_call_log | deployed | Conventions tying aliases, tiering, and call logging together across ~90 agents | llm-quality-testing.md |
| LQT-005 | Flywheel D — Claude vs local-model quality eval (replay harness) | partial | Quality-testing-category anchor for the paused Flywheel-D eval lane | llm-quality-testing.md |
| LCO-001 | Temperature/max_tokens logging gap in llm_call_log | partial | Columns exist but the Go writer never populates them from the actual API call | llm-call-observability.md |
| LCO-002 | Per-field LLM config resolution fallback chain | aspirational | Proposes lifting temperature to the same multi-level fallback max_tokens has | llm-call-observability.md |
| LCO-003 | Possibility-A-vs-B diagnostic method for silent LLM config failures | partial | Cheapest observability fix first, then audit-query to localise the real bug | llm-call-observability.md |
| LCO-004 | Default temperature hardening (chassis-level fallback ~0.4) | aspirational | Proposed ~0.4 default once the read path is proven, gated on earlier steps | llm-call-observability.md |
