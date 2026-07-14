
<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirm-not-initiate + the single central confirmer (one path to active)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_contract_set_review §2.1 "Resolution applied: a single central confirmer"; all contracts remain "contract specification" status.
- **what:** Agents propose (status=proposed rows + a work item); a human confirms; ONE component applies uniformly — flip to active, set last_verified_at/verified_by, deprecate the prior version, write the decision-log entry, emit the in-band change event — so confirm-not-initiate is a status-transition rule enforced in one auditable place, not a discipline reimplemented per agent. Hardening from the edge-case pass: the apply is one DB transaction with the change event in an outbox (crash-consistent, retry-safe), idempotent (re-applying an active version is a no-op), one live proposal per target extending down to layer rows (a new proposal replaces the proposed row; expiring a work item deprecates its row), and work items reference proposed rows by identity not pinned version.
- **sources:** FOCUS_contract_set_review.md#2.1,#2.3; FOCUS_pre_build_edge_cases(1).md#1.2,#1.3,#12; PLAN_config_work_items_contract(3).md#4
- **relations:** config_work_items; decision log; change layer in_band guard
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### config_work_items contract (mirror of site_work_items, tenant-scoped)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_config_work_items_contract(3) "contract specification"; corrected against the real site_work_items shape (FOCUS_schema_verification_findings §2).
- **what:** The shared queue every onboarding/maintenance agent emits into and the tenant reads: a parallel table (site_id NOT NULL blocks direct reuse) mirroring the verified shape — item_type/spec/result naming, integer priority, the real status lifecycle (detected→triaged→approved|rejected→claimed→complete|failed), reuse of approval_mode (the pre-existing confirm-not-initiate field; config defaults manual, 'auto' only for graduated capabilities), item_key unique-partial dedup (one live item per target), depends_on/parent_item_id, retry machinery. Batch confirmation for the initial onboarding flood (approval granularity adapts; apply still honours dependency order). Explicit scope: gates config, not deliverables.
- **sources:** PLAN_config_work_items_contract(3).md; FOCUS_schema_verification_findings.md#2; FOCUS_pre_build_edge_cases(1).md#2.1,#15
- **relations:** central confirmer; two-gated-paths; site_work_items (the reuse source)
- **verify-later:** table existence; site_work_items approval_mode semantics in code

<!-- SOURCE: U16_docs019_design_plans.md -->
### Decision log (immutable; premise vs rule_trace; inputs_used)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_decision_log_contract(2) "contract specification".
- **what:** The published-reasoning log given shape: append-only, one row per decision, carrying either a human-readable premise (judgement decisions) or a rule_trace (mechanical ones — exactly one of the two, so mechanical steps don't produce noise premises), plus inputs_used: the active-config slice in hand at decision time (compact atom id+version references + merged-view hashes by default; full snapshot inline for high-stakes kinds). Resolves freshness-vs-retrospect: compute on read for freshness, log at point of use for reconstruction. Write discipline: every decision logs, the entry precedes the apply, logging is not itself a logged decision. Read patterns: drift detection (premise vs current profile), heuristic invalidation, trust-ledger evidence, retrospective audit, compliance review. Open seam flagged: bundle assemblies would dominate a reasoning log by volume — bundle provenance may belong as the consuming decision's inputs_used instead.
- **sources:** PLAN_decision_log_contract(2).md; FOCUS_pre_build_edge_cases(1).md#4.4,#11; FOCUS_whole_plan_review.md#2.2
- **relations:** bundle provenance; trust ledger; work-item resolutions feed premises
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Trust ledger + bidirectional ratchet (asymmetric by design)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_trust_ledger_contract(4) "contract specification".
- **what:** Mutable state, one row per (tenant, capability): trust_level ∈ confirm_every/confirm_exceptions/notify/autonomous plus derived gate_policy and evidence_summary; derived from but separate from the immutable decision log (different access patterns). Cold start is always confirm_every (trust earned per tenant; no cross-tenant inheritance — deferred deliberately). Graduation up is always confirm-not-initiate; de-graduation down may auto-apply with notification on severe evidence — losing trust shouldn't wait on a human, gaining it should — but de-graduation evidence must first pass the defect-vs-partition filter so a flaky test or infra blip can't drop a capability and trigger a confirmation flood. Cascade routers and gate policy engines read it at runtime.
- **sources:** PLAN_trust_ledger_contract(4).md; FOCUS_pre_build_edge_cases(1).md#2.2; FOCUS_whole_plan_review.md#2.1
- **relations:** capabilities catalog (the ceiling); maintenance agent (the evidence source); outcome-record gap
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Capabilities catalog: the ceiling lives on the capability (blast radius caps trust)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_capabilities_catalog_contract(1) "contract specification. The sixth contract, closing the trust ledger's open dependency."
- **what:** A sibling table to agent_definitions (its existing capabilities jsonb holds free descriptive kebab tags for discovery, deliberately left alone; catalog capability_ids are snake_case dispatch keys — a recorded naming decision) holding per-capability ceiling, verifiability and containment. The ceiling is a judgement over the two factors (the weaker holds it); stored for cheap reads but the factors are authoritative — a factor change triggers a gated ceiling re-proposal. Capabilities aren't 1:1 with agents; the operation→capability mapping is declared at the action level. Seeding principle made explicit: the more a capability can break — especially chassis-editing ones — the lower its ceiling, regardless of verifiability; never fully autonomous for chassis-touching capabilities.
- **sources:** PLAN_capabilities_catalog_contract(1).md; FOCUS_pre_build_edge_cases(1).md#13; FOCUS_whole_plan_review.md#1.4
- **relations:** trust ledger; recursive self-improvement risk; cascade router
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Change-layer integration (change_events; in_band closes the self-modification loop)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_change_layer_integration_contract(4) "contract specification. … Closes the final contract gap before implementation."
- **what:** Change events from four sources (git_webhook, polling, in_band, periodic_sweep, + manual) land as first-class change_events rows (at-least-once with commit dedup; no event silently dropped — triggers_fired=[] is an explicit record). The trigger filter mapping changed paths to typed maintenance triggers is computed from the mechanical config, not stored (compute-on-read applied to routing). in_band emission — the tool's own applies emit events — is what keeps self-modification visible to the drift detector and decision log; rule: state changes emit, computed-view refreshes don't. Guard: a confirmer apply doesn't re-trigger maintenance on the entry just confirmed, but genuine downstream effects (audit code against a newly-active convention) still fire, and generation-origin in_band events are never exempt. reuse_index_refresh is its own trigger because a stale reuse index fails silently.
- **sources:** PLAN_change_layer_integration_contract(4).md; FOCUS_whole_plan_review.md#1.2; FOCUS_pre_build_edge_cases(1).md#4.1,#2.4
- **relations:** config-maintenance agent; central confirmer; reuse-search freshness
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Two gated paths: config changes vs deliverables
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §3 "a conceptual conflation to fix"; restated in the trust-ledger and work-items contracts.
- **what:** Changes to the tool's knowledge of the codebase (standards/objectives/mechanical) flow config_work_items → central confirmer → active config. The tool's outputs (generated code; edits to workflows/agent definitions — deliverables even though they're DB rows) flow cascade → trust-ledger gate → apply+commit+in_band event. The decision log spans both; the gates are not the same gate, and there are correspondingly two gated-mutation mechanisms (config confirmer; ledger ratchet-evaluator with asymmetric de-graduation).
- **sources:** FOCUS_pre_build_edge_cases(1).md#3; PLAN_trust_ledger_contract(4).md#1; PLAN_config_work_items_contract(3).md#5; FOCUS_whole_plan_review.md#2.1
- **relations:** trust ledger; config_work_items
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### The outcome-record gap (the loop runs on outcomes nobody sources)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §9 "real gap … You can log every input and decision and still have no feedback signal."
- **what:** The contracts log decisions and inputs, but nothing records whether a deliverable succeeded — verification pass/fail, reverted, human-corrected, accepted-as-is — the raw signal evidence_summary must aggregate for the ratchet to move. Companion gap (§10): "the bundle helped" has no defined metric (candidate signals: fewer correction rounds, fewer convention violations, less manual context-gathering); both needed before Phase 2.
- **sources:** FOCUS_pre_build_edge_cases(1).md#9,#10
- **relations:** trust ledger; thin-slice premise test
- **verify-later:** n/a (gap)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Thin vertical slice before the six-contract infrastructure
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** FOCUS_pre_build_edge_cases §8 recommended it; the tool plan's status note shows it happened ("a thin slice of Phase 1 is built … deliberately ahead of [the contracts] to test the core thesis first").
- **what:** The whole design rests on one unproven premise — an assembled bundle beats paste-and-rot — and none of the six contracts test it. So: hand-write a minimal flat-file constitution, build analyser+schema extractor, assemble ONE bundle for ONE real task, paste it by hand; only if it visibly helps build the infrastructure. This sequencing was followed: the thin-slice harness shipped and was used on real bugs while the contracts stayed specifications.
- **sources:** FOCUS_pre_build_edge_cases(1).md#8,#16; PLAN_context_assembly_tool_and_service(2).md status
- **relations:** contextkit harness; the six contracts (their build deliberately deferred)
- **verify-later:** n/a (executed strategy)

<!-- SOURCE: U16_docs019_design_plans.md -->
### External rollback (the self-hosting trap) + recursive self-improvement as residual risk
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §5, §14 — stated rules/risks, no implementation claimed.
- **what:** The tool runs on the chassis it modifies; a bad change could break the chassis badly enough that the tool can't run to fix it. Rule: rollback to known-good must be runnable externally, with no dependency on the agents/orchestrator being rolled back. And a self-improvement that passes verification can still degrade the tool's judgement gradually — not fully solvable; managed by conservative early trust, the human gate, external rollback and low ceilings for chassis-touching capabilities, and named as an accepted residual risk rather than assumed closed.
- **sources:** FOCUS_pre_build_edge_cases(1).md#5,#13,#14
- **relations:** capabilities-catalog ceilings; human gate; fork isolation
- **verify-later:** existence of any external rollback path

<!-- SOURCE: U16_docs019_design_plans.md -->
### Morality review as a configured, layered standard (not a baked-in view)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** MAPPING_tool_to_actions_and_agents(2) — design discussion; review contributors "(none yet)" in the thin-slice column.
- **what:** Distinct from liability ("will this get us sued" vs "is this right"): a build-time review contributor applying a layered standard held in the active config — an operator-chosen recognised base source (ASA/CAP Code, CMA guidance; OECD/UNESCO/NIST for the AI angle), operator values layered above it, jurisdiction/current-focus overlays later. Two altitudes: per-output, and a vertical-level gate at intake (should we build this site/industry at all). Contested cases route to HITL; the tool applies the configured standard and flags — it is not the moral authority.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md (morality review section)
- **relations:** build-time review contributors; active-config standards layer; council compliance eye
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Contributors vs checkers (build-path reviews ≠ improvement-loop monitors)
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** MAPPING(2): "Checkers are a different concept — not these" — a settled terminology/architecture distinction (reuse overlap flagged to investigate).
- **what:** Context contributors assemble bundle slices (code/data/runtime/standards); build-time review contributors (reuse, near-duplicate, liability, morality, correctness) review a PROPOSED change before it ships, raising concerns that revise or HITL-gate; improvement-loop checkers (the check_*.go family) continuously monitor DEPLOYED sites against plan/spec in the operate layer. Two layers restated: the website-builder builds sites; the context tool builds reliable changes to the builder.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md; PLAN_workflows_and_actions_migration(19).md (group-agent reviews)
- **relations:** council pattern (the reviews' fix-loop descendant); improvement-loop category
- **verify-later:** whether build-time reviews reuse check_*.go logic (flagged open)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirm-not-initiate + the single central confirmer (one path to active)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_contract_set_review §2.1 "Resolution applied: a single central confirmer"; all contracts remain "contract specification" status.
- **what:** Agents propose (status=proposed rows + a work item); a human confirms; ONE component applies uniformly — flip to active, set last_verified_at/verified_by, deprecate the prior version, write the decision-log entry, emit the in-band change event — so confirm-not-initiate is a status-transition rule enforced in one auditable place, not a discipline reimplemented per agent. Hardening from the edge-case pass: the apply is one DB transaction with the change event in an outbox (crash-consistent, retry-safe), idempotent (re-applying an active version is a no-op), one live proposal per target extending down to layer rows (a new proposal replaces the proposed row; expiring a work item deprecates its row), and work items reference proposed rows by identity not pinned version.
- **sources:** FOCUS_contract_set_review.md#2.1,#2.3; FOCUS_pre_build_edge_cases(1).md#1.2,#1.3,#12; PLAN_config_work_items_contract(3).md#4
- **relations:** config_work_items; decision log; change layer in_band guard
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### config_work_items contract (mirror of site_work_items, tenant-scoped)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_config_work_items_contract(3) "contract specification"; corrected against the real site_work_items shape (FOCUS_schema_verification_findings §2).
- **what:** The shared queue every onboarding/maintenance agent emits into and the tenant reads: a parallel table (site_id NOT NULL blocks direct reuse) mirroring the verified shape — item_type/spec/result naming, integer priority, the real status lifecycle (detected→triaged→approved|rejected→claimed→complete|failed), reuse of approval_mode (the pre-existing confirm-not-initiate field; config defaults manual, 'auto' only for graduated capabilities), item_key unique-partial dedup (one live item per target), depends_on/parent_item_id, retry machinery. Batch confirmation for the initial onboarding flood (approval granularity adapts; apply still honours dependency order). Explicit scope: gates config, not deliverables.
- **sources:** PLAN_config_work_items_contract(3).md; FOCUS_schema_verification_findings.md#2; FOCUS_pre_build_edge_cases(1).md#2.1,#15
- **relations:** central confirmer; two-gated-paths; site_work_items (the reuse source)
- **verify-later:** table existence; site_work_items approval_mode semantics in code

<!-- SOURCE: U16_docs019_design_plans.md -->
### Decision log (immutable; premise vs rule_trace; inputs_used)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_decision_log_contract(2) "contract specification".
- **what:** The published-reasoning log given shape: append-only, one row per decision, carrying either a human-readable premise (judgement decisions) or a rule_trace (mechanical ones — exactly one of the two, so mechanical steps don't produce noise premises), plus inputs_used: the active-config slice in hand at decision time (compact atom id+version references + merged-view hashes by default; full snapshot inline for high-stakes kinds). Resolves freshness-vs-retrospect: compute on read for freshness, log at point of use for reconstruction. Write discipline: every decision logs, the entry precedes the apply, logging is not itself a logged decision. Read patterns: drift detection (premise vs current profile), heuristic invalidation, trust-ledger evidence, retrospective audit, compliance review. Open seam flagged: bundle assemblies would dominate a reasoning log by volume — bundle provenance may belong as the consuming decision's inputs_used instead.
- **sources:** PLAN_decision_log_contract(2).md; FOCUS_pre_build_edge_cases(1).md#4.4,#11; FOCUS_whole_plan_review.md#2.2
- **relations:** bundle provenance; trust ledger; work-item resolutions feed premises
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Trust ledger + bidirectional ratchet (asymmetric by design)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_trust_ledger_contract(4) "contract specification".
- **what:** Mutable state, one row per (tenant, capability): trust_level ∈ confirm_every/confirm_exceptions/notify/autonomous plus derived gate_policy and evidence_summary; derived from but separate from the immutable decision log (different access patterns). Cold start is always confirm_every (trust earned per tenant; no cross-tenant inheritance — deferred deliberately). Graduation up is always confirm-not-initiate; de-graduation down may auto-apply with notification on severe evidence — losing trust shouldn't wait on a human, gaining it should — but de-graduation evidence must first pass the defect-vs-partition filter so a flaky test or infra blip can't drop a capability and trigger a confirmation flood. Cascade routers and gate policy engines read it at runtime.
- **sources:** PLAN_trust_ledger_contract(4).md; FOCUS_pre_build_edge_cases(1).md#2.2; FOCUS_whole_plan_review.md#2.1
- **relations:** capabilities catalog (the ceiling); maintenance agent (the evidence source); outcome-record gap
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Capabilities catalog: the ceiling lives on the capability (blast radius caps trust)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_capabilities_catalog_contract(1) "contract specification. The sixth contract, closing the trust ledger's open dependency."
- **what:** A sibling table to agent_definitions (its existing capabilities jsonb holds free descriptive kebab tags for discovery, deliberately left alone; catalog capability_ids are snake_case dispatch keys — a recorded naming decision) holding per-capability ceiling, verifiability and containment. The ceiling is a judgement over the two factors (the weaker holds it); stored for cheap reads but the factors are authoritative — a factor change triggers a gated ceiling re-proposal. Capabilities aren't 1:1 with agents; the operation→capability mapping is declared at the action level. Seeding principle made explicit: the more a capability can break — especially chassis-editing ones — the lower its ceiling, regardless of verifiability; never fully autonomous for chassis-touching capabilities.
- **sources:** PLAN_capabilities_catalog_contract(1).md; FOCUS_pre_build_edge_cases(1).md#13; FOCUS_whole_plan_review.md#1.4
- **relations:** trust ledger; recursive self-improvement risk; cascade router
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Change-layer integration (change_events; in_band closes the self-modification loop)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_change_layer_integration_contract(4) "contract specification. … Closes the final contract gap before implementation."
- **what:** Change events from four sources (git_webhook, polling, in_band, periodic_sweep, + manual) land as first-class change_events rows (at-least-once with commit dedup; no event silently dropped — triggers_fired=[] is an explicit record). The trigger filter mapping changed paths to typed maintenance triggers is computed from the mechanical config, not stored (compute-on-read applied to routing). in_band emission — the tool's own applies emit events — is what keeps self-modification visible to the drift detector and decision log; rule: state changes emit, computed-view refreshes don't. Guard: a confirmer apply doesn't re-trigger maintenance on the entry just confirmed, but genuine downstream effects (audit code against a newly-active convention) still fire, and generation-origin in_band events are never exempt. reuse_index_refresh is its own trigger because a stale reuse index fails silently.
- **sources:** PLAN_change_layer_integration_contract(4).md; FOCUS_whole_plan_review.md#1.2; FOCUS_pre_build_edge_cases(1).md#4.1,#2.4
- **relations:** config-maintenance agent; central confirmer; reuse-search freshness
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Two gated paths: config changes vs deliverables
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §3 "a conceptual conflation to fix"; restated in the trust-ledger and work-items contracts.
- **what:** Changes to the tool's knowledge of the codebase (standards/objectives/mechanical) flow config_work_items → central confirmer → active config. The tool's outputs (generated code; edits to workflows/agent definitions — deliverables even though they're DB rows) flow cascade → trust-ledger gate → apply+commit+in_band event. The decision log spans both; the gates are not the same gate, and there are correspondingly two gated-mutation mechanisms (config confirmer; ledger ratchet-evaluator with asymmetric de-graduation).
- **sources:** FOCUS_pre_build_edge_cases(1).md#3; PLAN_trust_ledger_contract(4).md#1; PLAN_config_work_items_contract(3).md#5; FOCUS_whole_plan_review.md#2.1
- **relations:** trust ledger; config_work_items
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### The outcome-record gap (the loop runs on outcomes nobody sources)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §9 "real gap … You can log every input and decision and still have no feedback signal."
- **what:** The contracts log decisions and inputs, but nothing records whether a deliverable succeeded — verification pass/fail, reverted, human-corrected, accepted-as-is — the raw signal evidence_summary must aggregate for the ratchet to move. Companion gap (§10): "the bundle helped" has no defined metric (candidate signals: fewer correction rounds, fewer convention violations, less manual context-gathering); both needed before Phase 2.
- **sources:** FOCUS_pre_build_edge_cases(1).md#9,#10
- **relations:** trust ledger; thin-slice premise test
- **verify-later:** n/a (gap)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Thin vertical slice before the six-contract infrastructure
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** FOCUS_pre_build_edge_cases §8 recommended it; the tool plan's status note shows it happened ("a thin slice of Phase 1 is built … deliberately ahead of [the contracts] to test the core thesis first").
- **what:** The whole design rests on one unproven premise — an assembled bundle beats paste-and-rot — and none of the six contracts test it. So: hand-write a minimal flat-file constitution, build analyser+schema extractor, assemble ONE bundle for ONE real task, paste it by hand; only if it visibly helps build the infrastructure. This sequencing was followed: the thin-slice harness shipped and was used on real bugs while the contracts stayed specifications.
- **sources:** FOCUS_pre_build_edge_cases(1).md#8,#16; PLAN_context_assembly_tool_and_service(2).md status
- **relations:** contextkit harness; the six contracts (their build deliberately deferred)
- **verify-later:** n/a (executed strategy)

<!-- SOURCE: U16_docs019_design_plans.md -->
### External rollback (the self-hosting trap) + recursive self-improvement as residual risk
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §5, §14 — stated rules/risks, no implementation claimed.
- **what:** The tool runs on the chassis it modifies; a bad change could break the chassis badly enough that the tool can't run to fix it. Rule: rollback to known-good must be runnable externally, with no dependency on the agents/orchestrator being rolled back. And a self-improvement that passes verification can still degrade the tool's judgement gradually — not fully solvable; managed by conservative early trust, the human gate, external rollback and low ceilings for chassis-touching capabilities, and named as an accepted residual risk rather than assumed closed.
- **sources:** FOCUS_pre_build_edge_cases(1).md#5,#13,#14
- **relations:** capabilities-catalog ceilings; human gate; fork isolation
- **verify-later:** existence of any external rollback path

<!-- SOURCE: U16_docs019_design_plans.md -->
### Morality review as a configured, layered standard (not a baked-in view)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** MAPPING_tool_to_actions_and_agents(2) — design discussion; review contributors "(none yet)" in the thin-slice column.
- **what:** Distinct from liability ("will this get us sued" vs "is this right"): a build-time review contributor applying a layered standard held in the active config — an operator-chosen recognised base source (ASA/CAP Code, CMA guidance; OECD/UNESCO/NIST for the AI angle), operator values layered above it, jurisdiction/current-focus overlays later. Two altitudes: per-output, and a vertical-level gate at intake (should we build this site/industry at all). Contested cases route to HITL; the tool applies the configured standard and flags — it is not the moral authority.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md (morality review section)
- **relations:** build-time review contributors; active-config standards layer; council compliance eye
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Contributors vs checkers (build-path reviews ≠ improvement-loop monitors)
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** MAPPING(2): "Checkers are a different concept — not these" — a settled terminology/architecture distinction (reuse overlap flagged to investigate).
- **what:** Context contributors assemble bundle slices (code/data/runtime/standards); build-time review contributors (reuse, near-duplicate, liability, morality, correctness) review a PROPOSED change before it ships, raising concerns that revise or HITL-gate; improvement-loop checkers (the check_*.go family) continuously monitor DEPLOYED sites against plan/spec in the operate layer. Two layers restated: the website-builder builds sites; the context tool builds reliable changes to the builder.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md; PLAN_workflows_and_actions_migration(19).md (group-agent reviews)
- **relations:** council pattern (the reviews' fix-loop descendant); improvement-loop category
- **verify-later:** whether build-time reviews reuse check_*.go logic (flagged open)
