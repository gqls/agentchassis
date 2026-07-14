# EXTRACTION U17a — docs019 _archive/excellent_discussions + _archive/working/main
Extracted 2026-07-13 (recovered from a sub-agent of U17 that completed before the parent hit the session limit). Part of U17 (docs019_documentation_audit_autonomous_build_and_operate/_archive/ + go_files/) — covers two sub-trees: excellent_discussions (exploratory MASTER/FOCUS reasoning docs on autonomous build-and-operate) and working/main (archived earlier drafts of 001/007/016/028/029/030/031/033/ARCHITECTURAL_TENSIONS/FOCUS_dispatch_diagnostic/FOCUS_interactive_content_generation, plus the April 2026 blog handoff).

## Coverage
(Common path prefix below abbreviated as `ED/` = `docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/excellent_discussions/` and `WM/` = `docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/`.)

| path | treatment |
|---|---|
| ED/MASTER_autonomous_build_and_operate(4).md | family-latest |
| ED/MASTER_autonomous_build_and_operate(1)-(3).md, .md | family-delta |
| ED/FOCUS_salience_and_multi_author_mediation(4).md | family-latest |
| ED/FOCUS_salience_and_multi_author_mediation(1)-(2).md, .md | family-delta |
| ED/FOCUS_self_development_coding_pipeline_reasoning(1).md | family-latest |
| ED/FOCUS_self_development_coding_pipeline_reasoning.md | family-delta |
| ED/FOCUS_best_practice_doc_tree(1).md | family-latest |
| ED/FOCUS_best_practice_doc_tree.md | family-delta |
| ED/FOCUS_standards_curation_and_governance(1).md | family-latest |
| ED/FOCUS_standards_curation_and_governance.md | family-delta |
| ED/FOCUS_context_authored_derived_change.md | full |
| ED/FOCUS_doc_tree_adoption.md | full |
| ED/FOCUS_mediator_routing_model.md | full |
| ED/102_blog_handoff-2026-04-10.md | full |
| WM/001_development_guide(0).md | full |
| WM/007_adoption_pipeline_v3.md | full |
| WM/016_debugging_guide_v2_44.md | family-latest |
| WM/016_debugging_guide.md, 016_debugging_guide_v2_18_ through v2_43.md (~35 versions) | family-delta |
| WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md | family-latest |
| WM/016_debugging_guide_addendum_adopted_tools_no_widget.md, (1).md | family-delta |
| WM/028_platform_mission_and_pipeline_direction.md | full |
| WM/029_site_plan_and_reconciler(1).md | full |
| WM/030_phase1_plan_and_reconciler(4).md | family-latest |
| WM/030_phase1_plan_and_reconciler(1).md | family-delta |
| WM/031_locks(2).md | family-latest |
| WM/031_locks(1).md | family-delta |
| WM/033_thunder_adapter_design.md | full |
| WM/ARCHITECTURAL_TENSIONS(2).md | family-latest |
| WM/ARCHITECTURAL_TENSIONS(1).md, .md | family-delta |
| WM/FOCUS_dispatch_diagnostic(3).md | family-latest |
| WM/FOCUS_dispatch_diagnostic(1)-(2).md, .md | family-delta |
| WM/FOCUS_interactive_content_generation(3).md | family-latest |
| WM/FOCUS_interactive_content_generation(1)-(2).md, .md | family-delta |

## Concepts

### Autonomous Build-and-Operate — the trust-not-capability thesis
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) header "Status: synthesis spine, built over several turns. Sections 6–8 are deliberately stubs to detail next"; §1 "the technical pieces are proven … the blocker is LLM-response uncertainty"
- **what:** Umbrella vision: everything already built is apparatus for a single reliability problem — bound LLM uncertainty at each step enough to progressively remove the human. The whole plan targets building/operating a real site (vonc.com) autonomously by composing the companion FOCUS mechanisms into one toolkit across the full lifecycle.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#1, ED/MASTER_autonomous_build_and_operate(4).md#9
- **relations:** umbrella over all the salience/standards/mediator/context FOCUS concepts below; vonc
- **verify-later:** none (discussion doc)

### Automation ratchet (per-capability trust levels)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §2 "Automation is a ratchet, not a switch"; a capability graduates `confirm-every → confirm-exceptions → notify → autonomous` "when evidence shows it reliable"
- **what:** Automation is not global; each capability (create action, provision nginx, reshard DB) carries its own trust level and graduates only on evidence. "Fully automated" is the union of individually-graduated capabilities. A trust ledger records each capability's level, gate policy, and supporting evidence.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#2, ED/MASTER_autonomous_build_and_operate(4).md#8.1
- **relations:** trust ledger; reliability cascade; bidirectional ratchet
- **verify-later:** none

### Reliability cascade (reuse → generate+verify → compete+judge → HITL)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §3 "Applied to every sub-task, in descending order of reliability"
- **what:** A per-task router for producing any unit of work in descending reliability order: known-good reuse, then generate+deterministic-verify, then compete-N-and-judge in a sandbox, then HITL. A verified recurring generated solution becomes candidate known-good and graduates its gate — feedback into the ratchet.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#3, ED/MASTER_autonomous_build_and_operate(4).md#7.2
- **relations:** cascade router; known-good library; multi-author generation
- **verify-later:** none

### Build-vs-operate asymmetry
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §4 "Build … isolatable … Competition is safe. Operate … live, stateful … Competition is risky"
- **what:** Build work (actions, workflows, components, agent defs) is branchable/sandboxable so competition is safe and the ratchet moves fast; operate work (provisioning, scaling, incident response) is live and stateful so it leans on known-good + canary + rollback + tighter HITL. The cascade's tier mix shifts by domain.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#4, ED/MASTER_autonomous_build_and_operate(4).md#7.6
- **relations:** lifecycle map (Tier A/B/C); reliability cascade
- **verify-later:** none

### Lifecycle map by verifiability + containment (Tier A/B/C)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §6.1–6.2 tables; "Ceiling is separate from current maturity"
- **what:** Every capability's autonomy ceiling is set by two independently-failing factors — verifiability (can we tell against ground truth it's correct) and containment (how bad/reversible if wrong). Tier A (Go actions, SQL, component-structural, observability, rollback) reaches autonomy; Tier C (security, sharding, replication, live remediation, meta-loop) stays gated regardless of agent capability. Ceiling ≠ maturity drives where to invest.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.1, ED/MASTER_autonomous_build_and_operate(4).md#6.2, ED/MASTER_autonomous_build_and_operate(4).md#6.4
- **relations:** build-vs-operate asymmetry; verification harness
- **verify-later:** none

### Four-layer documentation model for automation
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §5 "Four layers — two existing, two new — plus governance"
- **what:** Automation's connective tissue is four doc layers plus governance across them: the existing standards tree (prescriptive), the existing authored/derived context substrate, a NEW known-good solution library, and a NEW trust ledger; governance (curators/coordinator/advocate) sits across all four.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5
- **relations:** atomic standard; authored-vs-derived context; known-good library; trust ledger; standards curation
- **verify-later:** none

### Known-good solution library
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §5 "(NEW) … proven solutions captured as reusable, parameterised templates"; §6.3 "reuse pgvector + a parameterised-solutions table"
- **what:** A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. It is the substrate the cascade's tier-1 reuse draws on — derived success promoted to authored-reusable, gated multi-instance to avoid codifying a fluke.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** reliability cascade; authored-vs-derived context (artifacts→docs arrow)
- **verify-later:** pgvector; a parameterised-solutions table (proposed)

### Trust ledger + gate-policy engine
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7.2 "The trust ledger is the master knob"; §8.2 "Stand up the ledger (a table)"
- **what:** A per-capability store of automation level, gate policy, and supporting evidence, plus a small evaluator mapping (capability, trust level, stakes) → gate. It is the master knob: it governs both how conservatively a thing is produced and whether the result applies without a human.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** automation ratchet; cascade router; governance decision package
- **verify-later:** trust-ledger table (proposed)

### Cascade router
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §6.3 "Cascade router — the per-task decision (reuse / generate / compete / HITL)"; §8.4 "the loop's least-bounded step"
- **what:** An action/agent that picks a cascade tier per leaf task from three inputs — the capability's verifiability/containment, its trust-ledger entry, and the task's stakes. Named as the loop's least-bounded step, so conservative-by-default and ledger-gated.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#8.4
- **relations:** reliability cascade; trust ledger; verification harness
- **verify-later:** none

### Verification harness (build + ops)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.3 "Build side is easy … the ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work"
- **what:** Build-check / test-runner / validator / canary / rollback expressed as actions/adapters, checking output against ground truth. The build side reuses existing validate→regenerate; the ops side (canary, infra rollback, detection) is the thinnest, most-new part.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** toolchain validator (self-dev pipeline); lifecycle map
- **verify-later:** go build/vet/test runner; canary/rollback adapters (proposed)

### Autonomous control loop (route-produce-verify-gate-apply-feedback)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7 "the new machinery wraps each leaf task … the orchestrator's decompose-and-dispatch is reused unchanged"; §7.7
- **what:** The orchestrator's decompose-and-dispatch is reused unchanged; new machinery wraps each leaf: route (cascade), produce, verify (harness), gate (trust ledger level), apply→derived-state, feed back. Ops re-enters the same loop, triggered by derived state instead of a build goal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#7.6, ED/MASTER_autonomous_build_and_operate(4).md#7.7
- **relations:** cascade router; verification harness; trust ledger; mediator
- **verify-later:** existing orchestrator spawn/work-items machinery

### Bidirectional ratchet (trust can be lost)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7.5 "Trust can be lost, not only gained … the safety property that lets the ratchet advance at all"
- **what:** Feedback is two-directional: success accrues evidence toward graduation; repeated/severe failure drops the trust level, tightens the gate and raises the cascade floor. A regressing capability is automatically pulled back under supervision.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** automation ratchet; trust ledger
- **verify-later:** none

### Cross-cluster / multi-cluster dispatch (design reference within MASTER)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.6 "already designed and partly built (FOCUS_multi_cluster_dispatch_mvp): a dispatch_agent action + remote-job-spawner consuming system.dispatch.requests"
- **what:** Provisioning can target a third-party-hosted Kubernetes cluster with a `dispatch_agent` action and `remote-job-spawner`; remote agents reply on the same Kafka via `parent_responses_topic`; remote DB access uses a VPN tunnel + local PgBouncer at the same in-cluster DNS name. Forces one refinement: infrastructure-attributed failures are quarantined from the trust signal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.6, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** references FOCUS_multi_cluster_dispatch_mvp; trust ledger; verification harness
- **verify-later:** dispatch_agent action; remote-job-spawner; system.dispatch.requests; agent_dispatch_log

### Adoption path — vertical-slice dogfooding of the ratchet
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §8.1 "Vertical slice, not horizontal layer … Dogfood the ratchet … First capability = writing Go actions"
- **what:** Walk one capability (writing Go actions) end-to-end through route→produce→verify→gate→feedback before generalising; each new machinery piece starts at `confirm-every` and graduates on evidence. Phases 1–2 double as "improve my current chat workflow"; 3–6 are the leap to autonomy.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#8.1, ED/MASTER_autonomous_build_and_operate(4).md#8.2, ED/MASTER_autonomous_build_and_operate(4).md#8.5
- **relations:** automation ratchet; self-development coding pipeline
- **verify-later:** none

### Salience over presence (context bundle)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §1 "an LLM loses the bigger picture not because the text left the window … but because local detail is more salient mid-reasoning"
- **what:** The reframe underpinning the whole salience thread: attention follows the concrete and immediate, so the lever is salience at the moment of decision, not mere presence in a task-scoped context bundle.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#1, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** authored-vs-derived context; step-type-aware composition; checker model
- **verify-later:** none

### Four axes governing a development step
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §2 table (Purpose/How-well/Where-heading/What-is); "The dynamic axis was the gap"
- **what:** A dev step is governed by four axes — Purpose (why-chain, vertical), How-well (concern tree, horizontal), Where-heading (direction-of-travel, dynamic), What-is (code+state, local). The dynamic trajectory axis was the missing one: a snapshot says where things are, not where they're heading.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#2, ED/FOCUS_salience_and_multi_author_mediation(4).md#3
- **relations:** why-chain; direction-of-travel; concern tree
- **verify-later:** none

### Why-chain (objective-tree traversal)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a traversal of the existing objective tree … Stable, low-churn, human-owned"
- **what:** The purpose axis rendered as a root-to-node path over the existing objective tree. Turned into a *question* at decision/gate points ("does this serve [why-chain]?") — described as the strongest, cheapest anti-drift mechanism.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** four axes; priority profile; objective tree
- **verify-later:** existing objective/agent tree

### Direction-of-travel (trajectory layer)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a vector, not a reason … the authored vector laid over the derived change-layer … freshness-stamped"
- **what:** A fast-churn dynamic layer capturing current heading, settled-don't-relitigate decisions, deliberately-temporary states, and what's in flux. Proposed by the system from recent diffs but only human confirmation makes it authored-by-record; kept thin, pointer-rich, freshness-stamped, surfaced flagged-stale rather than silently trusted.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.6
- **relations:** why-chain; authored-vs-derived context; priority profile
- **verify-later:** none

### Step-type-aware prompt composition (altitude-aware)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §4 "framing/routing → full why-chain + direction; generation → collapse to a one-line tether … depth is a virtue, not only a failure mode"
- **what:** Prompt composition made altitude-aware: framing/routing gets the full why-chain + direction; generation collapses to a one-line tether; conformance is local; fitness-check and gate get full why-chain.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** salience over presence; why-chain; prompt-composition pattern
- **verify-later:** none

### Checker model (single-axis parallel checkers)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §5 "run several budgets, each narrow, each fully salient on one axis … curators and the advocate already are"
- **what:** Because one attention budget can't hold detail and breadth at once, run several narrow single-axis checkers fired *at decision points*, returning terse verdicts that are reconciled. Parallelism produces verdicts, not decisions — arbitration stays singular.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** curators/advocate; multi-author generation; mediator
- **verify-later:** none

### Multi-author generation (every concern authors a full solution)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §6 "each perspective is an author, not a guardrail … generative competition along the concern axis"
- **what:** Instead of guardrails, each implicated concern authors its own maximally-on-axis solution; disagreements become worked demonstrations, not complaints. Reuses cascade tier-3/mediator/advocate but competes N attempts at *different* objectives. Bounded by routing (~2–4 implicated concerns) and counter-proposals-on-deltas.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#6, ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** reliability cascade; mediator as multi-objective optimiser; N-round convergence
- **verify-later:** none

### Mediator as multi-objective optimiser
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §7 "'Right' = a requirement-relative, defensible balance … not pick, not merge"; "a heuristic floor … full mediation"
- **what:** The mediator finds the requirement-relative balance point among conflicting dimensions using the priority profile, with authored solutions as the extremes that bound the space. Priority is not global; provenance informs weighting, not deference. A cheap heuristic floor settles the uncontested majority but must emit a decision + provenance and be auto-flagged for re-mediation when the why-chain no longer matches its baked-in assumptions.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#7, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** priority profile; multi-author generation; N-round convergence; drift detection
- **verify-later:** none

### Priority profile (order not weights; sealed constraints)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §9 "Representation: an order, not numeric weights … A node stores only its differences from what it inherits … computed on demand"
- **what:** Requirement-relative priority among dimensions (security/speed/simplicity/generality/functionality/cost) lives on the objective-tree node as an *order* (with sealed/constraint flags), stored as differences-from-inherited and computed on read. Sealed constraints are ancestor-wins legal floors; a change triggers targeted re-validation of descendants holding conflicting overrides. The open crux (§9.7) is choosing the entry node.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#9, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.7
- **relations:** why-chain; mediator; drift detection; direction-of-travel
- **verify-later:** none

### N-round convergence (author/checker modes)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §8 "replaces that one hard step with several easier ones … non-convergence is the escalation signal"
- **what:** Replaces single-round reconciliation of N whole solutions with rounds where each active concern reacts to a candidate (satisfied → checker mode; still-needs → author mode), so the active set shrinks. Non-convergence isolates the one genuine value-laden tradeoff to escalate; bounded by an audit-pass-style cap; a concern can withdraw or be dismissed.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#8, ED/FOCUS_salience_and_multi_author_mediation(4).md#12
- **relations:** multi-author generation; checker model; mediator; candidate ownership
- **verify-later:** none

### N-round candidate ownership (owned by no concern; rival-base)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §12 "Resolution: the candidate is owned by no concern … seeded deliberately from the profile … The honest limitation: this is path-dependent"
- **what:** The per-round candidate is a shared artifact plus a change log, seeded from known-good or the highest-priority dimension's solution, changed only by premised targeted proposals the mediator adjudicates. Path-dependent and biased toward the seed; rival-base proposals are the gated escape from seed bias.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#12, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** N-round convergence; mediator reasoning trace; known-good library
- **verify-later:** none

### Published reasoning as substrate + drift detection
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §11 "Every decision publishes its reasoning, not just its outcome … Drift is the gap between a decision's logged premise and the current premise"
- **what:** Every decision publishes the *premise* it rested on, not just the outcome, upgrading the trace into reusable reasoning and training signal. This is what makes drift detection possible: a detector flags where a standard's stated premise no longer matches the premises recent decisions were actually made on.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#11, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** priority profile; mediator trace; confirm-not-initiate
- **verify-later:** none

### Self-development coding pipeline — coordination positions A/B/C
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) header "Status: exploratory. Nothing decided"; §5 "the live disagreement … How do area-owning agents coordinate a change that touches more than one area?"
- **what:** Use AI agents to develop the platform itself (one+ agent per area = one focus doc + its code slice). Unresolved crux: Position A work-item/ownership-serialized, Position B synchronous inter-agent negotiation, Position C a mediated go-between; current lean is C in a spawn-fresh variant (a per-change orchestrator that spawns current area-owners on demand, dissolving the ephemeral-worker staleness problem).
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#5, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#6
- **relations:** mediator routing model; toolchain validator; MASTER control loop; spawn-fresh coordinator
- **verify-later:** existing spawn machinery; coordinator agent_category

### Toolchain validator + repo read/search (net-new for code)
- **category:** tool-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) §4 "Validator changes kind: from contract checks … to a toolchain validator … the most important new piece"
- **what:** Low-regret net-new pieces for a self-coding pipeline: a toolchain validator giving ground-truth `go build/vet/test` + SQL dry-run pass/fail, a repo read/search capability (automating today's manual STEP ZERO), edits-against-existing-files rather than whole-file regeneration, and shared-repo serialization. The write→validate→regenerate loop, "broken output never overwrites," locks, HITL gating, and git→actions→backblaze deploy all transfer.
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#3, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#4, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2
- **relations:** verification harness; STEP ZERO; self-dev coordination positions
- **verify-later:** existing loop primitives; component_versions; needs_human_review

### Objective tree vs concern tree (two orthogonal axes)
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** FOCUS_best_practice_doc_tree(1) §1 "Two orthogonal structures, kept separate and cross-referenced, never merged"
- **what:** Two doc structures kept separate: the vertical objective tree (mission→branch→leaf) read downward, and the horizontal concern tree (best-practice cross-cutting standards) consulted sideways for any change. Each objective node carries `standing_concerns` linking to the standards it always pulls.
- **sources:** ED/FOCUS_best_practice_doc_tree(1).md#1, ED/FOCUS_best_practice_doc_tree(1).md#2.5
- **relations:** atomic standard; why-chain; concern curators
- **verify-later:** agent_definitions.domain_tags (proposed storage)

### Atomic standard (generated-views doc tree)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** FOCUS_best_practice_doc_tree(1) §2 "Optimal unit: the atomic standard, not the document … Documents are generated views over the atoms"
- **what:** The smallest addressable unit is one rule-atom with structured frontmatter (id, concern, scope, applies_to, kind, severity, status, version, supersedes, owner, check, related) and a body split into rule/rationale/examples. Constitution, per-concern handbooks, change-type bundles, and the machine manifest are all *generated views* over one source, so nothing drifts between a doc copy and an agent copy.
- **sources:** ED/FOCUS_best_practice_doc_tree(1).md#2, ED/FOCUS_best_practice_doc_tree(1).md#4, ED/FOCUS_best_practice_doc_tree(1).md#5
- **relations:** mediator routing model (the atom fields are the routing table); doc-tree adoption; concern curators
- **verify-later:** proposed `standards` table

### Standards curation & governance — concern curators
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §2 "Ownership maps to the concern, not to a node in the agent tree, and the set of owners is flat (one per top-level concern, ~8–10)"
- **what:** One curator agent per top-level concern owns that concern's atoms — reusing the auditor pattern and doubling as the routing advisor. A curator does vigilance + drafting + mechanical health but holds no authority over a rule's *meaning*. Ownership is flat and horizontal, deliberately not tied to the volatile agent tree.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#2, ED/FOCUS_standards_curation_and_governance(1).md#3, ED/FOCUS_standards_curation_and_governance(1).md#6
- **relations:** atomic standard; confirm-not-initiate; coordinator role; user-representative advocate
- **verify-later:** none

### Confirm-not-initiate HITL model (decision package)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §5 "Rules are not authored by humans from scratch and are not changed unilaterally by agents. The decision's reasoning is agent-led; the human confirms"
- **what:** Agents carry the analysis and produce a decision package (summary, tradeoffs/impact, genuine choices incl. reject and defer); the human confirms an informed choice, with rigor scaled to stakes. Acceptance writes a new version and deprecates the old (never deletes).
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#5, ED/FOCUS_best_practice_doc_tree(1).md#4.2, ED/MASTER_autonomous_build_and_operate(4).md#7.2
- **relations:** concern curators; coordinator; governance gate; deprecate-not-delete
- **verify-later:** none

### Coordinator role (arbitrates and frames)
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §7 "A thin coordinator layer above the curators … Resolved: the coordinator both arbitrates and frames"
- **what:** A thin layer above the peer curators owning what belongs to no single concern: the concern taxonomy, the `applies_to` vocabulary, cross-concern conflicts, and packaging cross-concern decisions for human confirmation. Both arbitrates and frames, checked by a user-aligned advocate inside the framing process.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#7, ED/FOCUS_standards_curation_and_governance(1).md#8
- **relations:** concern curators; user-representative advocate; decision authority
- **verify-later:** none

### User-representative advocate (intent + conflict triage)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §8 "A standing agent … representing the user sits inside the framing process as the check on the coordinator's framing power"
- **what:** A standing advocate for the user's intent that checks the coordinator's framing. Its signature job is triaging claimed conflicts before any reach the user: dissolve illusory or already-reconciled tensions, ask a clarifying question when unsure a conflict is real, and escalate only genuine tradeoffs.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#8, ED/FOCUS_standards_curation_and_governance(1).md#8.2, ED/FOCUS_standards_curation_and_governance(1).md#8.4
- **relations:** coordinator; decision authority; concern curators
- **verify-later:** intake-orchestrator; briefing-agent

### Decision authority (co-equal voices, abstention, creator veto)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §9 "Default: co-equal voices in the frame … Abstention fallback: the advocate decides, bounded … Creator override / veto"
- **what:** When advocate and curator genuinely disagree, both cases go to the human as co-equal voices. If the user declines to choose, the advocate decides bounded by codified intent — but high-stakes abstention and blocker conflicts escalate to the creator, who holds veto or optional final choice. Three distinct human roles: user, confirmer, creator.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#9
- **relations:** user-representative advocate; coordinator; confirm-not-initiate
- **verify-later:** none

### Mediator routing model (change → consultees)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS_mediator_routing_model.md#1 "the doc tree's metadata is the routing table … routing is matching a change descriptor against those tags"
- **what:** Routing reduces a change to a descriptor `{change_types, areas, touched_subsystems}` (paths→types via globs), queries the manifest for matching active standards, and acts on each by its own fields (run `check` validator / compose `reference` into prompt / consult concern agent / spawn area-owner). Runs a cheap tier always and an expensive tier on trigger; runs twice per change (pre from intent, post from diff).
- **sources:** ED/FOCUS_mediator_routing_model.md#2, ED/FOCUS_mediator_routing_model.md#5, ED/FOCUS_mediator_routing_model.md#6, ED/FOCUS_mediator_routing_model.md#7
- **relations:** atomic standard (fields are the routing table); self-dev Position C; concern curators
- **verify-later:** proposed change classifier; path→area glob map

### Authored vs derived context (one substrate, change layer between)
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** FOCUS_context_authored_derived_change.md#2 "The distinction that does matter: authored vs derived"; §3 "The change layer between"
- **what:** Everything grounding a reasoning step is retrievable evidence; the load-bearing split is authored (owned, maintainable, can drift/be-wrong) vs derived (emitted by the system running, no owner). A third change layer (diffs of code/logs) is derived-but-narrative and is the natural audit/learning surface. Two staleness modes get two fixes: keep authored thin+pointer-rich; fetch derived at reasoning time.
- **sources:** ED/FOCUS_context_authored_derived_change.md#2, ED/FOCUS_context_authored_derived_change.md#3, ED/FOCUS_context_authored_derived_change.md#4, ED/FOCUS_context_authored_derived_change.md#5
- **relations:** four-layer doc model; salience over presence; known-good library (artifacts→docs)
- **verify-later:** none

### Doc-tree adoption plan (constitution + tag/embedding retrieval)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_doc_tree_adoption.md header "actionable plan … without committing to the atomic rewrite, the mediator, or the routing build"; §1 "the corpus does not fit in context (~200 files, ~6.7MB, ~1.0–1.7M tokens)"
- **what:** First path to value from the doc-tree design against the current setup: Phase 1 write a tiny constitution, Phase 2 tag existing docs by concern/`applies_to` into a manifest, Phase 3 make the retrieval split real (tag-based deterministic selection for rules; existing nomic/pgvector/ollama RAG for the broad corpus), Phase 4 atomic extraction deferred/evidence-driven.
- **sources:** ED/FOCUS_doc_tree_adoption.md#4, ED/FOCUS_doc_tree_adoption.md#2, ED/FOCUS_doc_tree_adoption.md#5
- **relations:** atomic standard; mediator routing; RAG actions (existing stack)
- **verify-later:** rag_actions/nomic prefixes; proposed doc_index/standards table

### Development Guide (agent-build daily reference)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** 001(0) consolidation note "This is the canonical 001_development_guide. It supersedes the prior copy"; archive copy with live successor in docs024_key_docs_latest
- **what:** The consolidated practical reference for building/debugging/maintaining agents: core design principles (agents own their domain, callers pass raw data, workflows simple with complexity in Go, actions are the unit of work, spawn-before-call, reply-to-caller's-topic), a new-agent checklist, migration guide, and 20+ lessons-learned bug entries.
- **sources:** WM/001_development_guide(0).md#core-design-principles, WM/001_development_guide(0).md#checklist-for-new-specialist-agent, WM/001_development_guide(0).md#summary-of-rules-for-the-dev-guide
- **relations:** superseded by docs024 live 001; STEP ZERO; wrapper-orchestrator; loop mechanisms
- **verify-later:** platform/orchestration/actions/*

### STEP ZERO — reuse-before-create discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(0) "Pre-Flight: Does This Already Exist? … Real example: We built asset-deploy-agent … The existing asset-deployer already did the same thing. Three hours wasted"
- **what:** The mandatory pre-flight before creating any agent/action/function: search `agent_definitions`, the action registry, Go funcs, gate functions and workflows for every noun in the proposed name, document what was found, and prefer patching an existing thing. Includes the canonical field-path resolution rule (use `datahelpers.ExtractNestedField*`, don't add another).
- **sources:** WM/001_development_guide(0).md#pre-flight-does-this-already-exist-step-zero, WM/001_development_guide(0).md#field-path-resolution-use-the-canonical-functions, WM/001_development_guide(0).md#reuse-before-creating
- **relations:** Development Guide; STEP-ZERO-for-standards (curation); reliability cascade (reuse tier)
- **verify-later:** registry.go; datahelpers package; isStorageEnabledAgent

### Wrapper-orchestrator pattern (pod lifecycle)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) "Every pod-running agent needs a parent that spawned it … the rule we violated when we first wrote site-adoption-agent"; canonical minimal wrapper med-export-orchestrator
- **what:** Agents get a dedicated Kubernetes Job pod only when reached via `spawn_agent`→`call_agent`; substantive work reached via the generic entry point runs in-chassis with interleaved logs and blocks a shared pod slot. The fix is a tiny wrapper orchestrator (spawn→call→complete) so real work runs in its own pod.
- **sources:** WM/001_development_guide(0).md#every-pod-running-agent-needs-a-parent-that-spawned-it, WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics, WM/007_adoption_pipeline_v3.md#the-adoption-agent
- **relations:** generic entry point vs job topics; site-adoption-orchestrator; agent = row in agent_definitions
- **verify-later:** SpawnAgentAction; spawnAgentKubernetesJobFromDefinition; setupAgentTopics

### Kafka topic model (generic entry point vs per-spawn job topics)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) "system.agent.generic.requests — the generic entry point … job.<stable-identity>.requests — per-spawn dedicated topics"
- **what:** Two distinct patterns: the shared generic entry point (consumed by long-lived chassis replicas that run the workflow in-process via `config.agent_type`) for anything starting a workflow from outside a spawn tree, versus per-spawn `job.<stable-identity>.requests` topics for agent-to-agent traffic inside a workflow; plus per-type fixed topics for long-lived adapters.
- **sources:** WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics, WM/001_development_guide(0).md#agent-message-structure
- **relations:** wrapper-orchestrator; scheduled tasks target_topic; adapters
- **verify-later:** KAFKA_TOPIC(S) env; createTopics; MessageProcessor.extractGroupInfo

### Standardized input extraction (ActionInputSpec, ? optional, field collisions)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(0) "The three layers … input_mapping / input_fields / ActionInputSpec"; "Field name collisions … runs a nested-source loop late in its resolution chain"
- **what:** Three layers move data into an action: caller `input_mapping` (with `?`-optional destination keys), action `input_fields`, and `ExtractActionInputs(spec)` with a documented resolution chain. The nested-source loop iterates required AND optional fields, so names like `site_id`/`content_data`/`domain` can silently resolve from the wrong nested source — prefer collision-free names.
- **sources:** WM/001_development_guide(0).md#standardized-input-extraction, WM/030_phase1_plan_and_reconciler(4).md#note-on-the-target_site_id-input-field-name, WM/016_debugging_guide_v2_44.md#0
- **relations:** dispatch loop input_mapping; ? suffix; target_site_id convention
- **verify-later:** datahelpers/action_inputs.go; ResolveInputMapping; coordinator.go resolveInputMapping

### Stale orchestration sweeper/reaper
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 001(0) "Timeout handling uses in-process goroutines. These die when pods restart … This is the #1 cause of pipeline stalls"
- **what:** In-process timeout goroutines die on pod restart, stranding orchestrations in AWAITING_RESPONSES. A periodic DB sweep (every 60s, `FOR UPDATE SKIP LOCKED`) classifies each expired awaited request: child completed (synthesize the lost response), child failed (forward), or no child/still-running (retry up to 3 then fail parent). The `stale-orchestration-reaper` scheduled task also fails 24h-stale orchestrations.
- **sources:** WM/001_development_guide(0).md#stale-orchestration-sweeper, WM/016_debugging_guide_v2_44.md#4, WM/007_adoption_pipeline_v3.md#known-issue-zombie-dispatch-loop-pods
- **relations:** timeout chain; work item lifecycle; awaited_requests
- **verify-later:** orchestration_states; awaited_requests; scheduled_tasks stale-orchestration-reaper

### LLM infrastructure (model aliases, call logging, Ollama, RAG)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 001(0) "Implementation Status: LLM Optimization … Deployed and verified: 081/082 migrations, logging … Not yet deployed: ollama.go, rag_actions.go"
- **what:** Cross-cutting LLM infra: short model aliases resolved via `model_aliases.go`; fire-and-forget `llm_call_log` doubling as fine-tune training data; an Ollama CPU adapter serving nomic embeddings and quantized 7B classification; and `rag_lookup`/`rag_index` actions over a shared `knowledge_base` pgvector table with trigram fallback.
- **sources:** WM/001_development_guide(0).md#llm-infrastructure, WM/001_development_guide(0).md#implementation-status-llm-optimization
- **relations:** fine-tuning flywheel; LLM tiering; doc-tree adoption (RAG); Thunder adapter
- **verify-later:** llm_call_log; knowledge_base; migrations 081/082; ai_actions.go createAIClient

### Fine-tuning flywheel (call-log → LoRA → GGUF → Ollama)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** 001(0) "The training data pipeline: LLM call logging → export → LoRA fine-tune on GPU → GGUF export → load into Ollama → update agent definition to provider: ollama"; "Not yet built (future work)"
- **what:** A path to replace short-output classification/extraction agents with local fine-tuned models: accumulate 200+ successful `llm_call_log` examples, export Alpaca/ChatML, LoRA fine-tune on GPU (unsloth), export GGUF into Ollama, flip the agent definition to `provider: ollama`, then A/B test against Claude.
- **sources:** WM/001_development_guide(0).md#fine-tuning-path, WM/001_development_guide(0).md#implementation-status-llm-optimization, WM/033_thunder_adapter_design.md#tldr
- **relations:** LLM infrastructure; Thunder adapter; LLM tiering
- **verify-later:** training_data_export.sql; model_lifecycle.training_runs; unsloth

### Loop mechanisms (workflow expansion, dispatch loop, ErrLoopExpansionHandled)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) Appendix C "Loops are not Go for-loops — they are dynamic workflow expansion. At runtime, the loop step injects N × M steps"
- **what:** A loop step resolves a collection, then `handleLoopExpansion` injects `{loop}_iter_{N}_{substep}` steps into the workflow plan plus a `_complete` aggregator; `setLoopVariable` sets the current item and propagates prior-substep outputs. The canonical use is the dispatch loop (claim→spawn→call→mark). `ErrLoopExpansionHandled` is a sentinel fixing a race where a fast child response would otherwise skip remaining iterations.
- **sources:** WM/001_development_guide(0).md#appendix-c-loop-mechanisms, WM/001_development_guide(0).md#the-dispatch-loop-pattern, WM/001_development_guide(0).md#the-race-condition-and-errloopexpansionhandled
- **relations:** dispatch loop state machine; dynamic dispatch
- **verify-later:** loop_actions.go; loop_expansion_handler.go; coordinator.go continueExecution

### Domain submission & mission-driven sites (three tiers)
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 001(0) "The domain-submitter agent is the entry point for all new site builds … Three tiers of domain submission"
- **what:** `domain-submitter` is the entry point: it creates the site record, persists to `site_specs`, and emits the first `needs_domain_research` item. Three tiers: domain-only, domain+objective, and domain+`mission`/`roadmap`. Mission/roadmap aspects support any pre-planned site (e.g. vonc.com/Spark), bypassing the classifier's domain-discovery.
- **sources:** WM/001_development_guide(0).md#domain-submission-trigger-script-reference, WM/007_adoption_pipeline_v3.md#mission-driven-sites
- **relations:** classifier strategic brain; adoption modes; vonc
- **verify-later:** domain-submitter agent_definition; site_specs aspects mission/roadmap

### Work item lifecycle (blocking, unblocking, unresolved)
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** 001(0) "The unresolved mechanism … Located in load_work_item_actions.go, in insertWorkItem"; 102_blog_handoff "Unresolved status mechanism … line ~893"
- **what:** Items get blocked three ways (missing handler agent → auto-unblocked; spec status blocked → manual; manual block). The unresolved mechanism suppresses a re-emitted item if the newest terminal item with the same `item_key` is <3h old, and creates a visible-but-undispatched `unresolved` item (attempt_count 0) if 2+ prior terminal items exist.
- **sources:** WM/001_development_guide(0).md#work-item-lifecycle-blocking-unblocking-and-unresolved, ED/102_blog_handoff-2026-04-10.md#unresolved-status-mechanism
- **relations:** dispatch loop; feasibility-recheck; dedup index idx_swi_dedup
- **verify-later:** load_work_item_actions.go insertWorkItem; site_work_items status enum

### Website platform mission (best site per domain, one pipeline)
- **category:** business-strategy
- **status-signal:** partial
- **status-evidence:** 028 header "Living document. Second revision (2026-04-22)"; "produce the best possible website for each … with minimal human input"
- **what:** The anchoring *why*: given any domain, produce the best possible site end-to-end through one agent graph, where "best" = most useful to probable visitors (measured by engagement) and best revenue via whatever model genuinely fits. Commercial viability ≠ a brochure "business site"; defaulting to consultancy/services/contact when the signal is absent is an explicit failure mode to counter.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-mission, WM/028_platform_mission_and_pipeline_direction.md#commercial-viability-is-not-the-same-as-a-business-site, WM/028_platform_mission_and_pipeline_direction.md#failure-modes-we-want-to-eliminate
- **relations:** classifier strategic brain; fidelity dial; interactive content generation
- **verify-later:** site_specs classification aspect; domain-research-classifier

### Classifier as strategic brain (always runs full)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "The classifier … runs on every site entering the pipeline, and it always does its full job … Adoption does not shortcut it"; Phase 1 current, Phases 2–5 not implemented
- **what:** The `domain-research-classifier` decides what a site *should be* on every site; adoption/operator-mission are weighted inputs, not bypasses. It is not constrained to current capability — best-version items it can't build yet are marked `blocked` for `feasibility-recheck`. Silent override is the failure mode.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-classifier-is-the-strategic-brain-it-always-runs-in-full, WM/028_platform_mission_and_pipeline_direction.md#input-sources-and-their-weight, WM/028_platform_mission_and_pipeline_direction.md#phased-implementation
- **relations:** website mission; fidelity dial; spec-has-status; adoption pipeline
- **verify-later:** domain-research-classifier agent_definition; migration 006

### Fidelity dial (locked/high/medium/low)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "Fidelity … Five values, with high as the default when adoption evidence is present"; "depends on per-item status on specs … Phase 3"
- **what:** A dial controlling how much aspirational extension reaches the first build vs how faithfully it matches the strongest evidence (usually adoption): `locked` (exact, no promotion), `high`, `medium`, `low`; no-adoption reinterprets it as a confidence tolerance. Currently only implicit `high` (Phase 1).
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#fidelity-controlling-how-much-aspiration-reaches-the-first-build
- **relations:** spec-has-status; classifier strategic brain; adoption faithfulness locks
- **verify-later:** proposed adoption_meta/build_policy spec aspect

### Spec has per-item status — one spec, not two
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "one spec, not two. Items have status (deployed / planned / blocked) … It is not fully implemented yet … planned to be implemented in Phase 2"
- **what:** The dream is the full spec; the build is its non-blocked subset. Per-item status makes the dream-vs-build distinction mechanical — the build pipeline builds only `deployed`, `feasibility-recheck` promotes `blocked→planned`. Each spec row records source/source_agent/source_item_id for provenance; agents read-and-extend, never silently overwrite.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-spec-has-status-deployed-planned-blocked, WM/028_platform_mission_and_pipeline_direction.md#who-writes-what-who-doesnt-override
- **relations:** references doc 021; fidelity dial; feasibility-recheck
- **verify-later:** site_specs is_current/superseded_at; feasibility-recheck task

### Adoption pipeline & backend capability tiers (three-layer infra)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 007 "Phase 1 — Adoption pipeline (current)"; Phases 4–5 marked planned/future
- **what:** The platform runs in three layers (Layer 1 core factory; Layer 2 client delivery via S3 + config-driven site-api-router + client Postgres; Layer 3 framework builder), with five backend capability tiers from static+JS up to full platform. Adoption is a one-off capture, not a permanent state.
- **sources:** WM/007_adoption_pipeline_v3.md#infrastructure-separation, WM/007_adoption_pipeline_v3.md#backend-capability-tiers, WM/007_adoption_pipeline_v3.md#site-adoption, WM/007_adoption_pipeline_v3.md#principles
- **relations:** site adoption agent; design fingerprint; component selector/creator; site-api-router
- **verify-later:** site-adoption-orchestrator; site-api-router; vetcomparison export path

### Site adoption agent (crawl → fingerprint → classify → apply plan)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 "site-adoption-agent workflow (runs in the spawned pod): 16 steps … apply_adoption_plan → complete"
- **what:** A thin `site-adoption-orchestrator` wrapper spawns `site-adoption-agent` to run a 16-step workflow: firecrawl crawl, Go design-fingerprint extraction, LLM classification/archetype/content-direction/design-intent, and `apply_adoption_plan` writing specs, pages, and work items. Separates `target_url` (crawled) from `destination_domain` (built).
- **sources:** WM/007_adoption_pipeline_v3.md#the-adoption-agent, WM/007_adoption_pipeline_v3.md#three-stage-processing-go-extracts-design-llm-classifies-go-extracts-content, WM/007_adoption_pipeline_v3.md#running-an-adoption-what-to-expect-and-what-to-watch
- **relations:** wrapper-orchestrator; design fingerprint; canonicalisation (page identity); interactive parse-stage gap
- **verify-later:** apply_adoption_plan_action.go; extract_design_fingerprint_action.go; firecrawl_crawl

### Design fingerprint & design_reference vs design_intent
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 007 "Design Fingerprint Pipeline (added 2026-04-12)"; principle 7 "Design reference is history, design intent is direction"
- **what:** A Go extractor (`extract_design_fingerprint`, goquery) parses crawled rawHTML/external CSS into a fingerprint with a `suggested_mapping`; an LLM (`generate_design_intent`) turns it into a semantic brief. `design_reference` is an immutable historical record; `design_intent` is forward-looking direction — evolution happens by updating intent, never reference.
- **sources:** WM/007_adoption_pipeline_v3.md#design-fingerprint-pipeline-added-2026-04-12, WM/007_adoption_pipeline_v3.md#design-evolution-lifecycle, WM/FOCUS_interactive_content_generation(3).md#adoption-captures-content-and-extracts-structured-design-data
- **relations:** site adoption agent; interactive parse-stage; webdesign-agent three-way priority
- **verify-later:** enrich_fingerprint_with_css_action.go; site_specs design_reference/design_intent

### Component selector + creator (section_type vs function)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 007 "Phase 3 — Component selector, patterns, and research" (planned); "The separation: Planner decides WHAT section types … Component selector decides WHICH template"
- **what:** Splits the planner's conflated role: the planner picks section_types, a Go component-selector scores templates by metadata with a fallback to `needs_new_component`, and a `component-creator` agent LLM-generates a template from the full component contract when none fits. `function` currently does two jobs (page-role identifier + template choice); `section_type` separates them.
- **sources:** WM/007_adoption_pipeline_v3.md#component-selector-and-creator, WM/007_adoption_pipeline_v3.md#component-creation-contracts, WM/FOCUS_interactive_content_generation(3).md#components-more-broadly
- **relations:** interactive content generators; site plan sections; tool/game library model
- **verify-later:** content_components metadata columns; component-creator agent; plan_sections

### Human direction channels + lock lifecycle + audit-pass cap
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** 007 "Humans influence sites at three levels … Lock types: permanent / timed / review"; "The 3-pass audit cap prevents unbounded improvement cycles"
- **what:** Humans steer sites via three channels. HITL-requested content is protected by lock types (permanent, timed-90d, review-creates-HITL-item-on-expiry). A 3-pass audit cap bounds improvement cycles and resets on time/direction-change/major-rebuild/manual.
- **sources:** WM/007_adoption_pipeline_v3.md#human-direction, WM/007_adoption_pipeline_v3.md#hitl-requested-content-and-lock-lifecycle, WM/007_adoption_pipeline_v3.md#audit-pass-cap-reset
- **relations:** locks (031); improvement loop; direction spec
- **verify-later:** site_specs direction aspect; lock_type/lock_expires_at; last_audit_reset_at

### Site plan as declarative artefact + reconciler
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 029(1) "The shape that fixes this is the same pattern Kubernetes uses: a declarative artefact … plus a reconciler … Phase 0 lands today"
- **what:** Fixes the duplicate-pages bug where two surfaces (adoption + site-planner) both wrote `pages` rows without a shared identity space. The planner writes a declarative desired-state plan; a deterministic Go reconciler (`reconcile_site_plan`, no LLM) walks desired-vs-realised and emits `needs_page:<name>` for the diff only.
- **sources:** WM/029_site_plan_and_reconciler(1).md#why-this-exists, WM/029_site_plan_and_reconciler(1).md#phase-1-plan-as-declarative-artefact-reconciler-emits-work, WM/030_phase1_plan_and_reconciler(4).md#plan-builder-cascade-replaces-todays-site-planner-emit-and-queue
- **relations:** CanonicalisePage; plan-domain schema; LLM tiering; drift auditors
- **verify-later:** reconcile_site_plan action; site_plan_structure/pages; pages.built_from_plan_version

### CanonicalisePage + role validator (deterministic page identity)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 029(1) Phase 0 "A single canonicalisation helper in datahelpers/ … called from both surfaces"; 030(4) Q3 role validator (Go)
- **what:** A single `datahelpers/page_canonical.go` helper maps a `(role, slug, parent_section)` descriptor to a canonical `(name, url, page_type)` triple, called from both adoption and planner surfaces. Phase 1 extends it with `ParentSection` and adds a role-validator that corrects LLM role mislabels deterministically before persisting.
- **sources:** WM/029_site_plan_and_reconciler(1).md#fix, WM/030_phase1_plan_and_reconciler(4).md#q3-url-paths-canonicalisepage-phase-0-helper-extended-linknav-agents-own-drift, WM/016_debugging_guide_v2_44.md#adoption-faithfulness
- **relations:** site plan reconciler; architectural tension #1/#2; adoption faithfulness strip bug
- **verify-later:** datahelpers/page_canonical.go; ValidateRoles; CanonicalisePage

### Plan-domain schema + directive cascade + brief assembly
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030(4) Q1 "separate site_plans schema, not site_specs aspects … four plan-domain tables, all row-shaped for scale"
- **what:** Phase 1 rejects reusing `site_specs` aspects in favour of normalised plan tables (`site_plans`, `site_plan_pages`, `site_plan_sections`, `site_plan_directives`) row-shaped for 1000+ page scale. Guidance lives in `site_plan_directives` at site/page/section scope; a Go brief renderer (`datahelpers/page_brief.go`) walks the cascade and applies single- vs multi-valued cardinality.
- **sources:** WM/030_phase1_plan_and_reconciler(4).md#q1-plan-storage-separate-site_plans-schema-not-site_specs-aspects, WM/030_phase1_plan_and_reconciler(4).md#directive-cascade-and-brief-assembly, WM/030_phase1_plan_and_reconciler(4).md#what-stays-in-site_specs
- **relations:** site plan reconciler; lock transfer; strategic-vs-plan-time naming split
- **verify-later:** site_plan_directives; datahelpers/page_brief.go; write_site_plan action

### LLM tiering (large/medium/small/none → Opus/Sonnet/local-70B/Go)
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** 029(1) "Every action that calls an LLM declares its tier … flip medium from Sonnet to local. No action code touched"
- **what:** A cross-cutting `llm_tier` annotation on each LLM call site that the chassis maps to an endpoint via flippable config: Opus for strategy, Sonnet→local-70B for plan partials/audits, Haiku→local for slot-fills, Go for reconciler/validation.
- **sources:** WM/029_site_plan_and_reconciler(1).md#llm-tier-per-call-site, WM/029_site_plan_and_reconciler(1).md#affiliate-product-listings-same-pattern-applied-at-scale
- **relations:** LLM infrastructure; Thunder/local models; reliability cascade
- **verify-later:** llm_tier config → endpoint map (proposed)

### Locks — HITL durability across the platform
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_locks(2) "This doc is the canonical reference for lock semantics"; "Tech debt: lock-model coherence (target model) … Status (2026-05-19): the lock model has accreted"
- **what:** Two per-row lock patterns protect human-edited data: Pattern A (`locked_at`+`locked_by`, dominant) and legacy Pattern B (`pinned` boolean on site_specs, don't use for new tables). Every writer must read lock state before writing and preserve it when superseding. A coherence cleanup to three orthogonal columns under the invariant permanent⟺human is recorded as deferred tech debt.
- **sources:** WM/031_locks(2).md#the-two-patterns-in-use, WM/031_locks(2).md#lock-transfer-across-rebuilds, WM/031_locks(2).md#tech-debt-lock-model-coherence-target-model, WM/030_phase1_plan_and_reconciler(4).md#lock-transfer-across-plan-rebuilds
- **relations:** human direction/lock lifecycle (007); adoption faithfulness via locks; site plan directives
- **verify-later:** migration 053; check_component_lock.go; FOCUS_adoption_faithfulness_via_locks.md; PLAN_lock_coherence.md

### Thunder adapter (GPU provisioning, reaper, cost caps, credential boundary)
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** 033 header "Proposal for routing all Thunder Compute interactions through a long-running cluster adapter"; debugging guide v2_44 shows Phases 2–6 progressing
- **what:** A single long-lived `thunder-adapter` Deployment that holds the Thunder API key/B2 creds/SSH keypair store, provisions ephemeral GPU VMs via Kafka actions, and preserves a credential boundary: VMs get only ephemeral SSH keys + hours-expiring presigned URLs. Defence-in-depth: Thunder hard 12h uptime cap + a 15-min reaper + a daily cost cap.
- **sources:** WM/033_thunder_adapter_design.md#tldr, WM/033_thunder_adapter_design.md#preventing-indefinite-running-gpus-defence-in-depth, WM/033_thunder_adapter_design.md#new-schema, WM/016_debugging_guide_v2_44.md#9
- **relations:** fine-tuning flywheel; adapters pattern; multicluster provisioning
- **verify-later:** thunder_instances; thunder_budget_state; model_lifecycle.training_runs; system.adapter.thunder.requests

### Architectural tensions catalogue (infer-and-repair; multi-owner page identity)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** ARCH_TENSIONS(2) "An entry graduates from 'observed' to 'resolved' only when the resolution principle is actually enforced in code"
- **what:** A living catalogue naming genre-level design tensions that keep generating incidents. Tension #1: trusting LLM free-text structure as truth then repairing with starved heuristics vs deriving structure deterministically from the LLM's reliable signals. Tension #2: page identity re-derived in multiple stages that undo each other.
- **sources:** WM/ARCHITECTURAL_TENSIONS(2).md#tension-1-trusting-llm-free-text-structure-as-truth-infer-and-repair-vs-deriving-structure-deterministically, WM/ARCHITECTURAL_TENSIONS(2).md#tension-2-page-identity-is-derived-in-multiple-places-that-can-undo-each-other
- **relations:** CanonicalisePage; adoption faithfulness strip; site plan reconciler
- **verify-later:** ValidateRoles/nestedRoleFromURL; CanonicalisePage; normaliseRole vs normalisePageType

### Dispatch loop & detected→triaged→claimed state machine
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** FOCUS_dispatch_diagnostic(3) "detected is a valid intermediate state, not a bug"
- **what:** Discovery emits `detected` → design-audit-agent runs auditors then `triage_detected_items` → `triaged` → dispatch claims → `claimed` → handler → `complete`/`failed`. The dispatch chain is `scheduled_tasks`(30s) → `build-pipeline-trigger` (one site per tick) → `build-dispatch-loop`. A `NOT EXISTS`-on-claimed clause is an absolute per-site blocker.
- **sources:** WM/FOCUS_dispatch_diagnostic(3).md#tldr, WM/FOCUS_dispatch_diagnostic(3).md#q3-the-dispatcher-architecture-one-site-per-tick-not-exists-blocked-researched-2026-05-15, WM/FOCUS_dispatch_diagnostic(3).md#q4-what-is-the-pipeline-field-actually-for-surfaced-2026-05-15
- **relations:** work item lifecycle; loop mechanisms; claimed-item-timeout; triage_detected_items
- **verify-later:** build-pipeline-trigger find_dispatchable_site SQL; idx_swi_handler/idx_swi_site_pending; triage_detected_items registry.go:722

### Interactive content generation (four-stage parse/assess/generate/integrate)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** FOCUS_interactive_content_generation(3) "Tools — most mature … Games — nothing yet"; sequencing "locked in 2026-05-14"
- **what:** A map for building tools/games/news/other interactive types via a four-stage pattern (parse the source, assess what's producible, generate the artefact, integrate into a page). Tools are most mature but missing a parse stage; games have nothing. Capability assessment is a spec-lifecycle property marking each element `producible_now`/`producible_simpler`/`blocked`.
- **sources:** WM/FOCUS_interactive_content_generation(3).md#the-four-stage-pattern, WM/FOCUS_interactive_content_generation(3).md#whats-working-today, WM/FOCUS_interactive_content_generation(3).md#capability-assessment
- **relations:** tool pipeline; component creator; spec-has-status; adoption parse-stage
- **verify-later:** tool-suggester/deployer/generator/improver/auditor; page_types vocabulary

### Adoption parse-stage for interactive logic (interactive_reference/intent)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** FOCUS_interactive_content_generation(3) "1. Path C — Parse stage in adoption … Smaller piece of work than I first thought — closer to a couple of days"
- **what:** The prioritised gap: adoption captures markdown/design but not interactive JS. Closing it reuses the proven design-extraction shape: add `<script>`/`<canvas>` selectors to goquery, fetch `<script src>` via existing firecrawl_scrape, and add an LLM step producing `interactive_reference`/`interactive_intent` site_specs aspects.
- **sources:** WM/FOCUS_interactive_content_generation(3).md#where-the-parsing-capability-work-would-slot-in, WM/FOCUS_interactive_content_generation(3).md#sequencing-agreed-order, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** design fingerprint; interactive content generation; tool-recreation-handler misroute
- **verify-later:** extract_design_fingerprint_action.go; firecrawl_scrape; proposed extract_interactive_fingerprint

### Debugging guide & assumption-checklist methodology
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** 016 v2_44 §0 "Most defects in recent sessions … came from acting on unverified assumptions"; archive copy, live successor in docs024
- **what:** The canonical symptom→cause→fix guide, fronted by a 23-item assumption checklist. Covers pod health, work-item/orchestration/scheduled-task/error-log queries, timeout chain, and ~50 specific failure patterns.
- **sources:** WM/016_debugging_guide_v2_44.md#0, WM/016_debugging_guide_v2_44.md#9, WM/016_debugging_guide_v2_44.md#7
- **relations:** superseded by docs024 live 016; architectural tensions; agent = row in agent_definitions
- **verify-later:** orchestration_states.error_preview; agent_error_log; llm_call_log

### Agent = row in agent_definitions (workflow model)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.0 "An agent is not a Go type, class, or file. It is a row in agent_definitions whose default_config.workflow is a declarative graph of steps"
- **what:** Agents live in the database, not the Go source. A workflow is a step graph threaded by dotted-path reads from a shared data bag; "every agent is an orchestrator" is literal. `spawn_agent`+`call_agent` are a pair. Traps: the description can contradict the config; `agent_definitions` may be read from more than one DB.
- **sources:** WM/016_debugging_guide_v2_44.md#6.0, WM/016_debugging_guide_v2_44.md#6, WM/001_development_guide(0).md#agent-message-structure
- **relations:** wrapper-orchestrator; kafka topic model; loop mechanisms
- **verify-later:** agent_definitions (templates_db vs clients_db); default_config.workflow; SpawnAgentAction

### Snapshots and revert (snapshot_agent/revert_agent)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.1 "the convention is to call snapshot_agent('<agent-type>') first … revert_agent finds the most recent unrestored snapshot"
- **what:** Before patching an agent's `default_config`, call `snapshot_agent(type, reason)`; roll back with `revert_agent(type)`. Snapshots are rows in `agent_definitions_backup` kept as an audit trail. A legacy pre-migration pattern stored snapshots in `agent_definitions` itself (is_snapshot/version+1000), the source of several patch/revert footguns.
- **sources:** WM/016_debugging_guide_v2_44.md#6.1, WM/016_debugging_guide_v2_44.md#9
- **relations:** deprecate-not-delete; component_versions history; debugging guide
- **verify-later:** snapshot_agent/revert_agent functions; agent_definitions_backup

### LLM step config shadowing bug (per-object resolution)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §6.6 "a top-level ai_service shadows step-level overrides … Tracked in FOCUS_step_level_llm_config_ignored.md"
- **what:** `ExecuteLLMPromptAction` resolves the `ai_service` object once, taking the first match wholesale even if it lacks `max_tokens`, so a top-level ai_service silently shadows step-level model/max_tokens overrides and `max_tokens` falls back to a hardcoded 2048. Temperature has only one read path and isn't logged.
- **sources:** WM/016_debugging_guide_v2_44.md#6.6, WM/016_debugging_guide_v2_44.md#7
- **relations:** LLM infrastructure; LLM tiering; llm_call_log
- **verify-later:** ExecuteLLMPromptAction; AnthropicClient.GenerateText 2048; FOCUS_step_level_llm_config_ignored.md

### claimed-item-timeout & timeout chain
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §7 "Three timeouts interact and must be ordered correctly … claim_timeout > call_handler timeout > workflow timeout"
- **what:** A `claimed-item-timeout` scheduled task resets long-claimed items; three timeouts must stay ordered, else two handlers run one item. A two-phase reset (15-min evidence-based, 40-min blind) is used; the evidence check can produce false-positive completions.
- **sources:** WM/016_debugging_guide_v2_44.md#7, WM/016_debugging_guide_v2_44.md#9, WM/007_adoption_pipeline_v3.md#implementation-fixes-schema-notes-from-028j-handoff
- **relations:** dispatch loop; stale orchestration sweeper; work item lifecycle
- **verify-later:** claimed-item-timeout pre_query; scheduled_tasks; idle_timeout_seconds

### Adoption faithfulness — WriteSitePlanAction identity strip
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §9 "The corruption is in WriteSitePlanAction, not the LLM … Fix direction (not yet applied)"
- **what:** Even after a faithful adoption, `WriteSitePlanAction`'s `ValidateRoles`+`CanonicalisePage` interaction permanently strips identity for `content`/`blog_post` page_types: `ValidateRoles` derives a slug that strips `tool-/guide-/game-`/`-index`, and `CanonicalisePage` only re-adds prefixes for tool/game/guide roles — so mistyped section-index hubs flatten. Root cause is the wrong `page_type`; clean fix is upstream at adoption time.
- **sources:** WM/016_debugging_guide_v2_44.md#adoption-faithfulness-llm-convergence-are-faithful-writesiteplanaction-strips-identity-for-content-blog_post-types, WM/016_debugging_guide_v2_44.md#0, WM/ARCHITECTURAL_TENSIONS(2).md#tension-2-page-identity-is-derived-in-multiple-places-that-can-undo-each-other
- **relations:** CanonicalisePage; architectural tension #1/#2; locks (adoption_locked); FOCUS_adoption_faithfulness_via_locks
- **verify-later:** WriteSitePlanAction; datahelpers/page_canonical.go ValidateRoles/normaliseSlug; analyze_site page_type

### Tool page missing widget (M1 clobber vs M2 misroute)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 016 addendum(4) "RESOLVED 2026-05-26 → b1 … key the feature map by the canonical name in buildPageFeatureMap"; companion PLAN_tool_widget_clobber.md
- **what:** A `page_type='tool'` page rendering a description but no widget has two causes needing different fixes: M1 clobber (`SavePageSectionsAction` deletes page_components and its content-regression guard can't see a script-heavy widget) vs M2 never-generated (adoption recreate has no parse stage). For gamesdesign, root cause was a misroute: `buildPageFeatureMap` keys by raw page name while the route looks up canonicalised (`tool-`-prefixed) names.
- **sources:** WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#root-cause-m2-corrected-after-verification, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#diagnostic-recipe-read-only-30-seconds, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** CanonicalisePage; adoption parse-stage; site plan reconciler; interactive content
- **verify-later:** buildPageFeatureMap; tool-recreation-handler; SavePageSectionsAction; PLAN_tool_widget_clobber.md

### Blog-listing / orphan-page routing session handoff
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** 102_blog_handoff header "Session Handoff — April 10 2026"; "Ready to Deploy (files generated, not yet applied)"
- **what:** A dated operational handoff fixing blog-listing rendering (slot-name mismatch, empty-schema CSS-only template, missing article links) and reclassifying orphan pages into three routes (blog-post→rerender, nav-flags→nav-drift→nav-updater, no-nav→needs_internal_links→internal-linker). Documents self-hosted GitHub Actions runner deploy, the page-build-handler `error_step`-placement fix (46 validation crashes), the dedup pattern, and a future Mistral-Small-on-CPU internal-linker.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, ED/102_blog_handoff-2026-04-10.md#ready-to-deploy-files-generated-not-yet-applied, ED/102_blog_handoff-2026-04-10.md#remaining-unresolved-groups-not-yet-addressed
- **relations:** work item lifecycle/unresolved; deployment-github; nav sync; link management
- **verify-later:** rebuild_blog_listing_action.go; check_orphan_pages.go; github-actions-runner; nav-updater/internal-linker

### Deployment-GitHub / self-hosted runner + deploy path
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 102_blog_handoff "Self-hosted GitHub Actions runner — deployed and running … Runner v2.333.1, pod in ai-persona-system namespace"
- **what:** The publish path: agents commit generated site files via a git adapter; a self-hosted GitHub Actions runner runs the sites-repo workflow which `b2 sync`s to Backblaze. `needs_rerender` is the terminal build item that assembles pages and triggers deployment.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, WM/001_development_guide(0).md#every-pipeline-must-end-with-assembly-and-deployment, WM/007_adoption_pipeline_v3.md#data-flow-between-layers
- **relations:** blog-listing handoff; storage architecture (B2/S3); site plan reconciler terminal items
- **verify-later:** git-adapter; github-actions-runner dockerfile; needs_rerender handler

### Nav sync & config-driven page deactivation
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 001(0) "Nav Sync: Config-Driven Page Deactivation … deactivate pages not in the current plan"; deactivate_stale_pages config flag
- **what:** Header/footer nav displayed stale pages because `SyncPagesToDBAction`'s `ON CONFLICT` only overwrote matching names and nav queries didn't filter `build_status`. Fix: nav getters add `AND build_status = 'deployed'`, and a new-build flow deactivates pages absent from the current plan gated by `deactivate_stale_pages: true`.
- **sources:** WM/001_development_guide(0).md#nav-sync-config-driven-page-deactivation, ED/102_blog_handoff-2026-04-10.md#a-check_orphan_pagesgo-new-routing-logic
- **relations:** site plan reconciler nav auditor; link management; blog-listing handoff
- **verify-later:** SyncPagesToDBAction; GetHeaderNavFromPages/GetFooterNavFromPages; site_nav_items

### Adapter & service deployment debugging (rescued/dropped section)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** 016 base "Adapter & Service Deployment Issues … Rescued from an earlier guide revision … dropped from the main line"; absent from v2_44
- **what:** A family-delta section present in the base 016 but dropped from the v2_x main line: diagnosing adapter deployment failures (ImagePullBackOff/`insufficient_scope`, immediate crashes from `args:` replacing the whole CMD, `Unknown Topic Or Partition` on first message) and a deployment-essentials checklist. Built from the thunder-adapter Phase 2 debugging.
- **sources:** WM/016_debugging_guide.md#adapter-service-deployment-issues, WM/016_debugging_guide.md#imagepullbackoff-insufficient_scope-authorization-failed
- **relations:** dropped from 016 v2_44 (superseded main line); Thunder adapter; deployment-github
- **verify-later:** kustomize base/deployment.yaml; docker-hub-creds secret; ai-persona-app service account

### Extended thinking configuration
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(0) "Extended Thinking Configuration … When budget_tokens is set … the client adds {thinking: {type: enabled, budget_tokens: N}}"
- **what:** Setting `budget_tokens` in an LLM step's `ai_service` config enables Anthropic extended thinking: temperature is removed (API requirement), response parsing skips thinking blocks, latency rises 30–90s.
- **sources:** WM/001_development_guide(0).md#extended-thinking-configuration
- **relations:** LLM infrastructure; model aliases
- **verify-later:** platform/aiservice/anthropic.go thinking block

### QueryDatabaseAction parameterised queries & schema-drift discipline
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(0) Appendix A #1 "New query_database usage MUST use $1 placeholders with 'params' array. Never embed values via {{.field}} in SQL — SQL injection risk"; #18 "Schema column renames — always check the live schema"
- **what:** `QueryDatabaseAction` supports `$1` placeholders via a `params` config array (never Go-template interpolation into SQL); the live DB is the source of truth for column names (dumps drift), and best-effort/fire-and-forget writes silently no-op on schema mismatch. Includes the `to_jsonb('...'::text)` cast rule for updating prompt templates.
- **sources:** WM/001_development_guide(0).md#1-querydatabaseaction-doesnt-support-parameterised-queries, WM/001_development_guide(0).md#18-schema-column-renames-always-check-the-live-schema, WM/001_development_guide(0).md#15-postgresql-to_jsonb-fails-with-could-not-determine-polymorphic-type
- **relations:** debugging guide schema reminders; snapshots/revert; LLM call logging
- **verify-later:** QueryDatabaseAction; site_specs/site_work_items/component_versions schemas

## Scope-handling notes
excellent_discussions families are purely additive — latest versions contain every earlier section; no abandoned-idea deltas there (they are exploratory/aspirational by nature, "nothing decided", "synthesis spine"). The one genuine family-delta dropped concept found across both sub-trees is the "Adapter & Service Deployment Issues" section in base `016_debugging_guide.md`, absent from `016_debugging_guide_v2_44.md` (captured above). The working/main docs are archive copies whose live successors sit in `docs024_key_docs_latest` (001/016/007/028/029/030/031/033 all tagged superseded/partial per their in-doc phase status vs the live docs already captured in earlier units). This unit overlaps material also touched by U01/U02/U09/U10/U16 (live docs024 + docs019 design/plans) — consolidation should de-duplicate accordingly, retaining this unit's unique value: the MASTER/FOCUS reasoning-architecture concepts (salience, mediator, standards curation, authored/derived context) which have no other extraction unit covering them, since the rest of docs019/_archive/excellent_discussions was not otherwise in scope.
