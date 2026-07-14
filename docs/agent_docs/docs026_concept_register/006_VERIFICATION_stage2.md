# Stage 2 — Code/DB Verification Log

Started 2026-07-14. Each entry cross-checks a stage-1 *documentary signal* against
code, config, and deployment reality. Stage-1 status is what the docs claimed;
**verified status** is ground truth.

Method: for each concept, follow its `verify-later` pointers into the repo. Record
the exact file/line evidence. Where ground truth differs from the documentary
signal, correct the register entry and note why the docs misled.

---

## Batch 1 — priority tensions + suspected duplicates (2026-07-14)

### MCL-002 — Adjacent-cluster Phase 4a rollout: va001 second cluster
- **stage-1 signal:** deployed
- **verified status:** **aspirational** ⚠️ *(status corrected)*
- **evidence:**
  - `va001` appears in **zero** code, config, kubeconfig, or manifest files. Every
    hit is under `docs/_archive/` (running_notes prose).
  - Kubeconfigs present: `config_production_uk001`, `gpu_config_sanjose001`.
    There is **no** `config_production_va001`.
  - `remote-job-spawner` has kustomize overlays and terraform for **`uk_001` only**
    (`deployments/kustomize/services/remote-job-spawner/overlays/production/uk_001/`,
    `deployments/terraform/environments/production/uk001/services/agents/2260-remote-job-spawner/`).
    No va001 overlay exists.
- **why the docs misled:** `PLAN_isolated_chat_environment(5)` describes va001 in the
  **present tense** ("A second K8s cluster (va001…) *runs* remote-job-spawner"). It is a
  plan document narrating its own design as if built. Consolidation reasonably preferred
  the "more recent, more specific" evidence — but recency does not equal reality.
  **This is the canonical stage-2 failure mode: aspirational prose written in present tense.**

### MCL-001 — Multi-cluster dispatch contract (dispatch_agent + remote-job-spawner)
- **stage-1 signal:** partial
- **verified status:** **partial** ✓ *(confirmed, and made precise)*
- **evidence:**
  - `platform/orchestration/actions/dispatch_actions.go` — exists, 278 lines.
  - `cmd/remote-job-spawner/main.go` — exists, 591 lines.
  - `platform/orchestration/actions/registry.go:95` — `"dispatch_agent"` **is** registered.
    The docs listed "the registry patch is outstanding" — that gap is **now closed**.
  - Deployed to the primary cluster (kustomize + terraform, uk_001).
  - **Zero** `.sql` / `.yaml` / `.yml` / `.json` files reference `dispatch_agent` —
    **no workflow or agent definition invokes it.**
- **precise state:** the machinery is built, registered, and deployed — but nothing calls
  it, and (per MCL-002) there is no second cluster to dispatch *to*. It is a complete,
  live, unused code path. The owner's read ("aspirational, not deployed") is correct at
  the **system** level; "partial" is correct at the **component** level.

### MCL-003 — Cluster-filter gap in remote-job-spawner (Gap A)
- **stage-1 signal:** partial
- **verified status:** **partial** ✓ *(confirmed exactly)*
- **evidence:** `cmd/remote-job-spawner/main.go:202` — the `target_cluster` filter **exists**
  (`if targetCluster != "" && targetCluster != clusterID && targetCluster != "any"`).
  Line 203 logs the skip path at **`logger.Debug`**, not `Info`. The one-line
  observability fix is still outstanding, precisely as the register says.

### MCL-004 — Dispatch confirmation observability gap (agent_dispatch_log)
- **stage-1 signal:** aspirational
- **verified status:** **aspirational** ✓ *(confirmed)*
- **evidence:** `agent_dispatch_log` — zero hits across all `.sql` and `.go`. Table never created.

### MDL-029 / FTW-003 — LoRA fine-tuning path and the iter0 adapter
- **stage-1 signals:** MDL-029 deployed; FTW-003 partial ("last-mile wiring outstanding")
- **verified status:** **both correct — they describe different things.** Tension dissolves.
- **evidence — the artefact is real and closed out (MDL-029 = deployed ✓):**
  - `iter0_adapter_out/adapter_model.safetensors` — 828 MB, present on disk.
  - `iter0_adapter_out/manifest.json`: base `unsloth/Llama-3.3-70B-Instruct-bnb-4bit`,
    1,958 examples, 3 epochs, lora_r 16, final loss **0.266**, 25.1 h runtime
    (90,337 s), peak 44.2 GB VRAM, completed **2026-06-04T20:33:11Z**. A genuine,
    successfully-trained adapter.
- **evidence — it never reached inference (FTW-003 = partial ✓):**
  - `deployments/kustomize/services/ollama-adapter/base/deployment.yaml:34` pulls only
    **`nomic-embed-text`** and **`mistral-small3.1`** — stock models. No Modelfile, no GGUF
    import, no adapter mount.
  - Zero references to `lora` / `adapter_model` / `iter0` anywhere in the Go source. The
    only hit is an aspirational comment in `vet_med_price_scrape_action.go:187`
    ("collects training data for *future* LoRA fine-tuning").
- **resolution:** the adapter was **trained but never served**. The `log→export→LoRA→GGUF→
  Ollama→swap` pipeline is complete through `LoRA` and stops dead there. Consistent with the
  owner's "late testing" read. **Recommend clarifying MDL-029's summary** to say the adapter
  is a closed-out *training artefact*, not a serving model — as written, "deployed" invites
  the misreading that it is in production.

### ADM-007 / ADM-008 / PUB-001 — public API + site_ownership *(duplicate cluster)*
- **stage-1 signals:** ADM-007 aspirational; ADM-008 abandoned; PUB-001 aspirational
- **verified status:** all three **confirmed unbuilt** ✓
- **evidence:** zero hits for `site_ownership` (any `.sql`/`.go`) and zero hits for
  `api/v1/sites` (any `.go`).
- **duplicate ruling:** **PUB-001 is a genuine duplicate** — it is exactly ADM-007
  (the `/api/v1/sites/*` endpoint plan) plus ADM-008 (the `site_ownership` junction table),
  merged into one entry by a different consolidator cluster. Recommend: keep ADM-007 and
  ADM-008 (the finer-grained pair, correctly split by status — the endpoints are
  *aspirational*, the table is *abandoned*), and retire PUB-001 to a pointer.
  The `public-api.md` category holds only this one concept and can be folded into
  `admin-dashboard-and-api.md`.

---

## Scope finding: the `deployed` bucket is not safe to trust

The single status error found in batch 1 (**MCL-002**) came from the **deployed** bucket —
not from the partial/unknown bucket that stage 2 was scoped to verify. Every partial,
aspirational, and abandoned signal checked in batch 1 held up exactly.

That is the opposite of the assumed risk profile. The mechanism is clear and general:
**a plan document that narrates its own design in the present tense reads as evidence of
deployment.** Consolidation's tie-break rule — prefer the most recent, most specific
evidence — actively *selects for* this failure, because the polished present-tense plan is
usually the newest document in the family.

Implication: the 871 `deployed` concepts are the bucket most likely to contain
false positives, and false positives there are the *expensive* kind — they are what a
stage-3 council agent would confidently assert as built. The partial/unknown bucket is
comparatively safe: it is already flagged as uncertain, so a wrong signal there is
self-limiting.

**Recommendation:** widen stage 2 from *"verify the 314 partial/unknown"* to
*"verify the 314 partial/unknown **and** sweep the 871 deployed for present-tense-plan
false positives."* The deployed sweep is cheap per concept — most have a named file, table,
or endpoint in `verify-later`, and existence is a one-grep check.

---

## Batch 2 — full-corpus workflow sweep (2026-07-14)

Ran the widened scope from batch 1's recommendation as a 145-agent workflow: one agent per
category-file work unit (68 "deep" units covering all 314 partial/unknown concepts, 77
"sweep" units covering all 871 deployed concepts in ≤45-concept chunks), each following
`verify-later` pointers into the live repo. Every proposed status correction then went
through an **independent adversarial re-check** — a fresh agent, blind to the first agent's
reasoning beyond its evidence claim, instructed to try to refute it — before being accepted.
347 agents total, 1,189 verdicts, ~44 minutes, 11.0M tokens, 3,101 tool calls.

**Result: 105 corrections confirmed, 97 proposed corrections overturned by the adversarial
pass.** The near-1:1 confirm/overturn ratio shows the adversarial gate is doing real work,
not rubber-stamping — see the overturned-corrections note below.

### Discovery: a 7th status was needed — `convention`

49 of the 105 corrections were the same shape: a concept tagged `deployed` in stage 1 that
is not a code/DB/infra artifact at all — it's a design doctrine, working practice, or
one-off analysis (e.g. "reuse before create", "epistemic tagging discipline", "honest-delta
disclosure in pitch docs"). Stage 1's six-value vocabulary had no slot for this, so
extraction defaulted them to `deployed` because the docs describe them as current,
established practice. That is a genuine documentary-signal error distinct from the
present-tense-plan failure mode found in batch 1: a stage-3 council agent reading `deployed`
would go hunting for code that was never meant to exist.

The adversarial pass took this correction seriously rather than rubber-stamping it — in
several cases (BIZ-004, BIZ-008, BIZ-010, and others in the overturned list) it found real
enforcing code the first-pass agent missed and overturned the `convention` proposal back to
`deployed`. Added `convention` to the status vocabulary in `README.md`.

Full list of the 49 `deployed → convention` reclassifications:

<details>
<summary>49 concepts reclassified deployed → convention (click to expand)</summary>

| ID | Category | Concept idea (not a code artifact) |
|---|---|---|
| BIZ-013 | business-strategy | Honest-delta disclosure is a pitch-writing/documentation discipline (a comparison table in a doc), not a code/infra artifact; stage1 itself lists verify-later: n/a. |
| BLD-001 | build-pipeline | This is a documented methodology/finding (census route method), not a built artifact claim; no code to check — status-evidence is a doc citation of a one-off analysis, correctly reclassed as process not code-deployed. |
| CTXE-004 | context-engineering-principles | Explicitly a standing doctrine ('n/a — doctrine' in verify-later). No named file/service/table to check; 'deployed' status label is a misnomer for a methodology principle. |
| CTXE-005 | context-engineering-principles | Bundle-size working rule/doctrine, verify-later says 'n/a (doctrine)'; no concrete artifact named (bundle size limit is a heuristic, not code). |
| DES-055 | design-composition | Three-per-row grid rule is a content-authoring convention encoded in design_intent.layout_preference prose, not a standalone code/db artifact. |
| DES-060 | design-composition | Hazard/band-class split is an audit classification report from a specific run, not a standing code/db object to verify by grep. |
| DEV-001 | development-guide | reuse-before-create discipline (STEP ZERO), a working practice not a single artifact |
| DEV-009 | development-guide | agent-vs-infrastructure design test/heuristic, not an artifact |
| DEV-018 | development-guide | manual work-item crafting discipline, a process not a code artifact |
| DEV-019 | development-guide | file itself notes verify-later 'n/a (convention, not code)' |
| DEV-030 | development-guide | a norm/discipline, not a code artifact |
| DEV-040 | development-guide | deploy-verification discipline, not a built artifact |
| DEV-041 | development-guide | operational workaround/timing rule; file itself says 'n/a — process/design record' |
| DEV-045 | development-guide | port/copy-drift procedure, not a code artifact itself |
| DEV-049 | development-guide | schema-before-SQL discipline; verify-later explicitly '—' |
| DEV-083 | development-guide | ops reference/procedure description, not a single artifact |
| DEV-084 | development-guide | recurring compliance-audit practice, not a code artifact |
| DOC-004 | documentation-system | idea.uk running-notes journal discipline; verify-later itself says 'n/a — documentary practice' |
| DOC-007 | documentation-system | packaging workflow provenance note; verify-later says 'n/a', records a practice not a build claim |
| DOC-008 | documentation-system | epistemic tagging discipline, a documentation practice; verify-later says 'n/a (convention)' |
| DOC-009 | documentation-system | BUNDLE/HANDOFF/PLAN/RUNBOOK cold-start doc practice, not a claimed chassis code build |
| DOC-024 | documentation-system | prose→structured→enforced graduation rule is a design/philosophy statement; no lock mechanism found or claimed |
| DOC-025 | documentation-system | purely conceptual framing (plan=enforced/pipeline=compiled/notes=log); verify-later says 'n/a (conceptual framing)' |
| DOC-031 | documentation-system | handoff-document discipline (travelling_docs thread); verify-later says 'n/a (working practice)' |
| DOC-035 | documentation-system | single-source-of-truth doc relocation practice; no code artifact |
| DOC-036 | documentation-system | documentation-consolidation methodology; verify-later says 'a documentation-process note' |
| DOC-037 | documentation-system | house convention of pre-flight SQL scripts; verify-later says 'n/a — process/design record' |
| DOC-040 | documentation-system | doc claim-verification/dated-claim convention, a documentation-authoring rule not code |
| DOC-043 | documentation-system | classify-don't-merge principle; verify-later says 'n/a (principle)' |
| DOC-044 | documentation-system | design principle from DESIGN_doc_drift_classifier (itself unbuilt, DOC-041); 'n/a (principle)' |
| DOC-053 | documentation-system | multi-thread working-practice description; verify-later says 'n/a (practice)' |
| FIX-007 | fix-loop | benchmark methodology/doctrine, no concrete artifact named beyond a rubric in a doc |
| FIX-008 | fix-loop | narrative history of pilot-candidate selection (a process record), not a code/infra claim |
| FIX-009 | fix-loop | benchmarking discipline/rule about symptom wording and seed_scope omission |
| FIX-031 | fix-loop | governing design principle ('PR is the human terminal'), not itself a checkable artifact |
| FIX-047 | fix-loop | five-criterion intake doctrine — explicitly 'n/a (doctrine)' per its own verify-later field |
| FIX-048 | fix-loop | cross-cutting design pattern description ('gates between every LLM step'); the individual gates it describes are independently confirmed real |
| FIX-049 | fix-loop | recorded design decision/value proposition narrative; no code artifact claimed |
| FTW-015 | finetuning-flywheel | economic decision (no snapshots); verify-later explicitly says 'none (economic decision)' |
| FTW-016 | finetuning-flywheel | mental/performance model narrative, not a built artifact |
| FTW-022 | finetuning-flywheel | evaluative verdict/judgment call, not a built artifact |
| HITL-010 | hitl | hitl_respond.sh: 0 hits anywhere in repo; a documented operational runbook, not a built artifact (underlying awaited_requests mechanism is real but this named script isn't) |
| IMG-063 | imagery | runbook operating model / division-of-labour doctrine (humans-at-taste-layer), no built artifact claimed |
| INVD-002 | investigation-discipline | verify-later explicitly says 'n/a — process/design record, no separate code artifact' |
| MCL-012 | multicluster | documented operational discipline (re-derive facts each session), not a testable code artifact |
| OPD-002 | operating-doctrine | parallel-thread boundary/handoff convention — a documented working-agreement, not code/infra |
| SPEC-005 | site-spec-and-classifier | design rule/heuristic ('agents with install side-effects must queue re-run work item'), not a single built artifact |
| SQLC-001 | sql-change-management | change-management methodology/pattern (needle-gate SQL editing), not a single built artifact |
| TL-012 | tool-lifecycle | empirical argument ('checks passing != working'), not a built artifact |

</details>

### The 56 real status flips

These are genuine ground-truth corrections — not taxonomy mismatches, but cases where the
documentary signal and the code disagreed about whether something is live, dead, or
half-built:

<details>
<summary>56 concepts with a corrected deployed/partial/aspirational/superseded/unknown status (click to expand)</summary>

| ID | Category | Was | Now | Why |
|---|---|---|---|---|
| ADM-009 | admin-dashboard-and-api | partial | deployed | No literal site-admin-dashboard.jsx file exists, but frontends/admin-dashboard/src/App.tsx (2439 lines) implements the same Review Queue UI, wired to live /api/v1/admin — a mock prototype was superseded by a real API-wired dashboard. |
| ADO-009 | adoption-pipeline | unknown | deployed | EnsureSiteRecordAction does INSERT...ON CONFLICT(domain) DO UPDATE; sites.domain has a UNIQUE constraint — structurally prevents duplicate site rows for a domain. |
| ASG-003 | agent-spawning-and-groups | partial | deployed | agent_discovery.go is imported and called by discovery_actions.go/spawn_actions.go; registry.go registers discover_agents/discover_best_agents as live actions. |
| ASG-006 | agent-spawning-and-groups | partial | aspirational | evolution.go/performance.go (full EvolutionService) exist but have zero external Go callers and no registry entry — built, never wired in. |
| BATCH-001 | batch-processing | partial | aspirational | Zero hits anywhere in code for llm_batch_queue, llm_batch_agent_config, QueueLLMBatchAction, escalate_batch_item — a fully-specified design with no implementation. |
| BATCH-004 | batch-processing | unknown | aspirational | processing_tier column exists in one SQL migration; zero readers/writers anywhere in code. |
| BIP-001 | business-intelligence-platform | partial | deployed | business_intel schema + all named tables exist and are populated; live agent_definitions workflow drives them. |
| BIP-006 | business-intelligence-platform | partial | deployed | Live agent_definitions row for vet-batch-processor with the full described workflow, active in production. |
| BIP-009 | business-intelligence-platform | partial | deployed | Live agent_definitions row for vet-pipeline-orchestrator with the full deployed workflow JSON. |
| BLD-015 | build-pipeline | unknown | deployed | page-rebuild is actively called by maintenance-triage via a live rebuild_loop step, backed by real Go actions (PrepareRebuildDispatchesAction) and a registry handler. |
| CANB-002 | canine-biology | partial | aspirational | Training-run infra is real but generic (not vertical-specific); the vet-specific LoRA/RAG project itself was never executed. |
| CGV-006 | content-governance | unknown | partial | Dual-source COALESCE pattern is confirmed live in three call sites; a consolidation action exists but coverage is incomplete. |
| CGV-018 | content-governance | unknown | aspirational | content_items/content_item_id: DDL exists but zero Go readers/writers anywhere. |
| CGV-025 | content-governance | partial | aspirational | install_chat/uninstall task_type: zero hits in code outside a design doc; nothing enqueues or handles it. |
| CQ-007 | content-quality | unknown | partial | The exact defects (brand-suffix titles, tool-flavoured copy, empty descriptions) are documented as expected/known, matching a partial rather than deployed or unknown state. |
| CTXA-017 | context-assembly | partial | deployed | resolveCodeRepoLabel is a real shared resolver called from two independent call sites. |
| CTXE-006 | context-engineering-principles | partial | aspirational | resolve_targets/embed/fuse exist only as a standalone prototype Go module, never merged into the chassis. |
| CTXK-011 | contextkit-toolchain | partial | deployed | cmd/diagnose/main.go is a real CLI harness; platform code imports the matching pkg/diagnose package. |
| CTXK-012 | contextkit-toolchain | partial | deployed | pkg/diagnose/* files exist and are ahead of the contextkit docs copy — genuinely live, docs are stale. |
| DBI-008 | database-and-infrastructure | unknown | aspirational | sites.build_status is set to 'pending' at insert but has zero code paths that ever update it afterward — a dead field. |
| DES-047 | design-composition | partial | aspirational | The described extract_computed_styles/getComputedStyle Go action does not exist anywhere in the codebase. |
| DEV-034 | development-guide | partial | deployed | ValidateInputContract now checks both the top-level and input_data.spec.* paths, matching the documented fix exactly. |
| DEV-073 | development-guide | partial | aspirational | Only the config-provided-agents SpawnGroupAction is registered; the DB-lookup variant described in docs was never wired into the registry. |
| DGH-003 | deployment-github | partial | aspirational | resolveGitRepoName exists but has zero callers anywhere in the repo — dead code. |
| ENT-001 | entity-data | partial | aspirational | None of site_entities/site_entity_relationships/entity_sources/entity_sync_log exist as tables anywhere — design only. |
| FIX-034 | fix-loop | partial | deployed | fixloop_digest_action.go + its registry entry are live, and a later commit postdates the doc's "awaiting next image" snapshot — the gap has since closed. |
| FIX-042 | fix-loop | unknown | aspirational | bug_records and guideline-amendment/guideline-gap handling: zero hits anywhere in code. |
| HITL-017 | hitl | partial | aspirational | No hitl_handler.go or any matching file exists; no /api/v1/hitl or hitl/respond route exists in code. |
| IMG-006 | imagery | unknown | aspirational | The current planner prompt has zero references to site_archetype and always requests images — the archetype-gated behaviour was never implemented. |
| IMP-010 | improvement-loop | unknown | deployed | The live pre_query at HEAD already implements the documented ordering fix, predating the handoffs that questioned it. |
| LNK-019 | link-management | partial | aspirational | None of link-crawler/link-validator/link-registry-sync/redirect-manager/affiliate-link-manager exist in code outside docs. |
| MCL-002 | multicluster | deployed | aspirational | va001/config_production_va001 has zero hits in code/config/deployment — a present-tense-plan doc was misread as a completion record. |
| MCL-010 | multicluster | partial | aspirational | Zero Go-code implements the viability/RTT classification logic; the numbers are explicitly labelled as projections pending a smoke test that itself depends on the (non-existent) second cluster. |
| MDL-003 | model-infrastructure | partial | deployed | ollama.go, a live "ollama" case in createAIClient, and full kustomize/deployment manifests all exist and match. |
| NEWS-004 | news-feed-pipeline | deployed | aspirational | loadNewsItems has no ROW_NUMBER/PARTITION BY logic anywhere — a plain ORDER BY with LIMIT, no source-diversity interleaving exists. |
| NEWS-017 | news-feed-pipeline | partial | deployed | check_orphan_pages.go implements exactly the documented 3-route classification end-to-end. |
| PAY-008 | payments | partial | deployed | ReviewBeforePay config flag and its default-on behaviour are confirmed live in the actual service code. |
| PBP-005 | page-build-pipeline | partial | deployed | reconcileGeneratedItemKeys is defined and actually called inside RenderComponentAction, gated correctly. |
| PLAN-032 | site-plan-and-reconciler | unknown | deployed | emit_design_items_action.go now exists, explicitly built to close this exact gap per its own header comment. |
| PLAN-038 | site-plan-and-reconciler | partial | superseded | load_page_sections_from_spec_action.go now reads site_plan_sections as authoritative — the older approach was replaced, not left half-done. |
| RAGK-002 | rag-knowledge-base | partial | deployed | RAGLookupAction (vector search + trigram fallback) is fully implemented and registered as a live action. |
| SCH-013 | scheduler-and-tasks | partial | superseded | The documented gap (no work-item-level reaper) was real when written but has since been replaced by a working mechanism. |
| SPEC-013 | site-spec-and-classifier | unknown | deployed | spec-updater's apply_update step has a matching, existing Go action (update_site_spec_from_item). |
| SQ-002 | site-quality | unknown | partial | render_site_components exists and is wired into the main build workflow, but coverage/quality gaps remain. |
| STY-041 | styling-render-pipeline | partial | deployed | The three proposed assemble_* actions now all exist in the registry exactly as specified. |
| STY-046 | styling-render-pipeline | partial | superseded | generate_css's action was changed from an LLM prompt to a deterministic render_css_from_spec Go function — the bug class was fixed by replacement, not patched. |
| SYS-004 | system-architecture | deployed | partial | No sweeper.go or StaleOrchestrationSweeper exists; a thinner real mechanism (cleanupExpiredAwaitedRequests) exists but lacks the documented SKIP LOCKED claiming and A/B/C classification logic. |
| SYS-012 | system-architecture | unknown | partial | The per-pod-unique responses-group design is confirmed in code, partially matching the fuller documented scheme. |
| SYS-060 | system-architecture | unknown | partial | fuel.go's cost/enforcement logic is genuinely called and enforced, but only in one agent, not system-wide. |
| SYS-081 | system-architecture | unknown | deployed | Optimistic-locking CAS update, retry loop, and failure handling are all confirmed live in state.go/coordinator.go. |
| SYS-082 | system-architecture | unknown | deployed | The documented max-3-retries-with-incremented-retry_version scheme is implemented exactly as described. |
| TL-001 | tool-lifecycle | partial | deployed | The "interactivity regression blocked" guard is a real, committed safeguard in save_page_sections_action.go. |
| TL-015 | tool-lifecycle | partial | deployed | browser-runner-adapter has a full cmd/, kustomize overlay, Dockerfile, and adapter package — completely built and deployed. |
| TLIB-005 | tool-library | partial | deployed | The fix motivated by this closed incident (a field-contract guard) is confirmed live in store_generated_component_action.go. |
| WDS-004 | work-dispatch | partial | deployed | All three previously-open failure modes are now confirmed fixed in code. |
| WDS-005 | work-dispatch | unknown | partial | The claim/orchestration-timeout cleanup bug is resolved in coordinator.go, but related gaps remain. |

</details>

### On the 97 overturned corrections

The adversarial pass is not a rubber stamp — spot-checks show it doing real work in both
directions:

- **Catching over-eager downgrades.** Several `deployed → convention` proposals in the
  business-strategy category were overturned with concrete counter-evidence: BIZ-004's
  doctrine is actually encoded as executable scoring/gating logic in `engine.go`; BIZ-008's
  pricing convention drives a real Stripe charge (`billing.go:49`, a live £29 transaction on
  2026-06-14); BIZ-010 likewise had real enforcing code the first-pass agent missed.
- **Preferring precision over a coarser match.** Several `partial → deployed` proposals in
  business-intelligence-platform were overturned back to `partial` because the register's own
  vocabulary ("partial — live but with known quality gaps") was the more precise fit — the
  workflows are live, but the team's own `status='experimental'` label on the underlying
  agent_definitions rows says the gap is real, not closed.

Full detail for all 97 overturned proposals — including the ones this summary doesn't cover
by name — is in the workflow journal (not duplicated here to keep this file navigable);
the pattern above generalizes to the rest.

### Net effect on the register

Status distribution, before → after batch 2:

| status | before | after |
|---|---|---|
| deployed | 871 | 847 |
| partial | 274 | 246 |
| aspirational | 271 | 290 |
| superseded | 99 | 102 |
| abandoned | 72 | 72 |
| unknown | 40 | 21 |
| convention | 0 | 49 |

All 105 corrections were applied directly to their `register/<category>.md` entries (status
line updated, `stage2-verified` provenance note added inline) and to
`register/000_concept_index.md` (Status column updated). The remaining ~1,522 concepts keep
their stage-1 documentary status, now implicitly higher-confidence since batch 2's sweep
found no further errors among them beyond what's listed above.

---

## Batch 3 — superseded/abandoned sweep, closing out stage 2 (2026-07-14)

Extended `stage2_workflow.js` with two more prompt modes and ran the last unswept portion of
the register: all 102 `superseded` + 72 `abandoned` concepts (174 total, 73 work units, 100
agents, 2.7M tokens, ~7.6 minutes). Same verify → adversarial-recheck pipeline as batch 2.

- **`superseded` mode** hunts for the failure where a claimed replacement doesn't actually
  exist — meaning the "old" mechanism is quietly still the live one.
- **`abandoned` mode** hunts for ideas quietly resurrected since the abandonment note was
  written.

**Result: 18 corrections confirmed, 9 overturned by the adversarial pass.** ~10.3% error
rate — consistent with batch 2's ~9% (105/1,185), suggesting the register's error rate is
fairly uniform across status buckets once you look past the `deployed`-bucket present-tense
problem batch 1 found.

| ID | Category | Was | Now | Why |
|---|---|---|---|---|
| ADM-010 | admin-dashboard-and-api | superseded | partial | Auth/Projects/Subscriptions routes are genuinely gone (proxied externally), but Templates and Instances routes are still live with full CRUD handlers — the bundled "v1 persona REST surface" concept is only half-superseded. |
| AME-002 | agent-memory-and-evolution | abandoned | partial | `agent_variants` table was never built, but the sibling `is_snapshot` mechanism from the same design is fully live (5 call sites, migration 021) — one half survived, one didn't. |
| DES-009 | design-composition | superseded | partial | `theme_tags` is genuinely gone, but `css_snippets`/`js_snippets` are still actively wired and queried — the claimed successor only replaced one of four bundled components. |
| DOC-064 | documentation-system | abandoned | partial | The doc survives byte-identical under a sibling live project folder (`adoption/docubundle/`) that stage 1's search never reached — a **search-scope gap**, not a genuine abandonment. |
| FLW-003 | flows-and-narrative | superseded | abandoned | Neither the old mechanism nor its claimed replacement (PERS-001, itself abandoned) was ever built — a pivot between two unbuilt designs, not a live supersession. |
| HITL-006 | hitl | superseded | partial | The claimed replacement (domain-submitter) is real and wired, but the old intake-orchestrator was never deactivated anywhere in the repo — a new parallel entry point, not a proven decommission. |
| IMP-026 | improvement-loop | abandoned | deployed | The exact three-step algorithm the docs said vanished is fully implemented and wired (`write_audit_findings_action.go:583-687`) — only the documentation of it lapsed, not the code. |
| ONB-020 | onboarding-config | superseded | partial | Claimed successor has zero code implementation; the era-2 descendant mechanisms it was meant to replace are still live and wired. |
| ONB-022 | onboarding-config | superseded | deployed | `briefing_questionnaire`/`fetch_agent_questionnaire` are fully live (registry + 3 call sites); the claimed successor has no matching implementation at all. |
| SCH-007 | scheduler-and-tasks | abandoned | deployed | The exact ownership-gap fix the docs said was never built (`last_completed_at` set by the scheduler on both paths) is implemented in `cmd/scheduler/main.go`. |
| SCH-008 | scheduler-and-tasks | abandoned | deployed | The exact starvation-prevention logic (timeout escape hatch, `countInFlight`) is implemented and wired into the main scheduler loop. |
| SCH-009 | scheduler-and-tasks | abandoned | deployed | The docs' claim "scheduler does not currently read fire_message" is now false — the code reads and branches on it directly. |
| SNAP-003 | site-snapshots-and-revert | superseded | partial | The old snapshot function is genuinely gone, but the claimed `needs_snapshot`→snapshot-agent replacement has zero Go implementation — live code instead uses a simpler `is_current`/`superseded_at` flag pattern not described in either the old or claimed-new design. |
| SPEC-015 | site-spec-and-classifier | superseded | partial | Claimed replacement `intake-orchestrator-v2` was never built (0 SQL hits; its own proposal doc says the old one "must not be disrupted" and still routes builds); the original intake-orchestrator remains the live mechanism. |
| STY-036 | styling-render-pipeline | superseded | partial | `aggregate_webpage` is still registered and enabled (registry.go:371), used in a non-backup SQL file modified after the claimed supersession — and its claimed successor is itself dead code, so nothing was cleanly replaced. |
| STY-039 | styling-render-pipeline | superseded | partial | `assemble_multipage_site` is still registered, enabled, and used live in non-backup SQL — the old batch-assembly path coexists with the newer per-page loop rather than having been removed. |
| SYS-025 | system-architecture | superseded | partial | Stage 1 itself flagged a mixed classification; the QA architecture doc was merged (not dropped) and the underlying three-layer mechanism is confirmed live via two agent-definition SQL files. |
| SYS-026 | system-architecture | superseded | deployed | The `domain`→`pipeline` column rename is confirmed complete and wired into live action code, not a pending migration. |

**A new failure-mode class, distinct from batch 2's present-tense-plan finding: search-scope
gaps.** DOC-064 was tagged `abandoned` because extraction's search for its file was scoped to
one doc subtree (`idea.uk/`) and never reached a sibling live project folder
(`adoption/docubundle/`) holding a byte-identical copy. Present-tense-plan errors (batch 1/2)
come from *misreading* evidence that was found; this is a different mechanism — evidence that
was simply never *found* because the search boundary was too narrow. Worth keeping in mind for
any future extraction pass: single-subtree searches can miss cross-cutting duplicates.

**Cluster observation:** SCH-007/008/009 — all three scheduler-and-tasks concepts corrected in
this batch — are the same shape: a documented gap, a real fix landing in `cmd/scheduler/main.go`,
and the docs simply never being updated to reflect it. Three independent confirmations of the
same underlying drift (code moved on, docs didn't) rather than three unrelated findings.

All 18 corrections applied directly to their register entries and the master index (two —
SYS-025/026 — needed a manual patch after the apply script's block-capture regex missed an
edge case: an italic "merged from N findings" annotation line immediately after the header
broke the bullet-only capture pattern).

### Final status distribution (stage 2 complete)

| status | stage 1 | after batch 2 | after batch 3 (final) |
|---|---|---|---|
| deployed | 871 | 847 | 853 |
| partial | 274 | 246 | 257 |
| aspirational | 271 | 290 | 290 |
| superseded | 99 | 102 | 90 |
| abandoned | 72 | 72 | 67 |
| unknown | 40 | 21 | 21 |
| convention | 0 | 49 | 49 |

Total corrections across all three batches: **1 (MCL-002, batch 1, hand-verified) + 105
(batch 2) + 18 (batch 3) = 124 corrections** out of 1,627 concepts (~7.6%), plus one duplicate
resolved (PUB-001 → pointer entry, not a status change).

### Stage 2 status: COMPLETE

Every concept in the register (1,627) has now been checked against the live codebase at least
once — all 314 originally-scoped partial/unknown, all 871 deployed (added after batch 1's
scope finding), and all 174 superseded/abandoned (this batch). Both priority items from the
original handoff (multi-cluster dispatch, LoRA/Thunder) and all three suspected duplicates are
resolved. The register is ready for **stage 3** (expert council agents per concept area) — see
`PLAN_concept_register.md` §Stage 3 for the design, which is now grounded in the live fix-loop
mechanism.
