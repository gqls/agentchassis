# Register — reasoning

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

12 concepts, consolidated from 24 raw extractions across units U13 and U17a (each of the 12
distinct blocks appeared byte-identically twice within the cluster input file — treated as
duplicate copies of one extraction, not independent corroboration).

### RSN-001 — Chain-of-thought prompt pattern catalog
- **status:** unknown
- **status-evidence:** Presented as a curated list of externally-sourced/community prompts with no indication any is wired into an actual chassis agent's system prompt.
- **what:** A reference collection of five chain-of-thought prompting archetypes: (1) "Step Budget and Reflection" — scratchpad thinking with a numeric step budget and self-scored confidence driving continue/backtrack; (2) "Stream-of-Consciousness" — raw, marker-tagged unpolished reasoning trace; (3) "Panel of Experts" — simulated multi-domain-expert debate with per-claim correctness percentages; (4) "Enhanced Reasoning Protocol" — a two-stage consult-then-branch protocol; (5) classic baseline CoT ("Let's think step by step").
- **sources:** reasoning/001_chain_of_thought_prompts.md
- **relations:** n/a
- **verify-later:** whether any agent_definitions system prompt actually uses one of these five patterns

### RSN-002 — Salience over presence (context bundle)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §1 "an LLM loses the bigger picture not because the text left the window … but because local detail is more salient mid-reasoning".
- **what:** The reframe underpinning the whole salience thread: attention follows the concrete and immediate, so the lever is salience at the moment of decision, not mere presence in a task-scoped context bundle.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#1, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** authored-vs-derived context (context substrate model, autonomous-build-operate register ABO-003); step-type-aware composition; checker model
- **verify-later:** none

### RSN-003 — Four axes governing a development step
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §2 table (Purpose/How-well/Where-heading/What-is); "The dynamic axis was the gap".
- **what:** A dev step is governed by four axes — Purpose (why-chain, vertical), How-well (concern tree, horizontal), Where-heading (direction-of-travel, dynamic), What-is (code+state, local). The dynamic trajectory axis was the missing one: a snapshot says where things are, not where they're heading.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#2, ED/FOCUS_salience_and_multi_author_mediation(4).md#3
- **relations:** why-chain; direction-of-travel; concern tree
- **verify-later:** none

### RSN-004 — Why-chain (objective-tree traversal)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a traversal of the existing objective tree … Stable, low-churn, human-owned".
- **what:** The purpose axis rendered as a root-to-node path over the existing objective tree. Turned into a question at decision/gate points ("does this serve [why-chain]?") — described as the strongest, cheapest anti-drift mechanism.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** four axes governing a development step; priority profile; objective tree
- **verify-later:** existing objective/agent tree

### RSN-005 — Direction-of-travel (trajectory layer)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a vector, not a reason … the authored vector laid over the derived change-layer … freshness-stamped".
- **what:** A fast-churn dynamic layer capturing current heading, settled-don't-relitigate decisions, deliberately-temporary states, and what's in flux. Proposed by the system from recent diffs but only human confirmation makes it authored-by-record; kept thin, pointer-rich, freshness-stamped, surfaced flagged-stale rather than silently trusted.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.6
- **relations:** why-chain; authored-vs-derived context; priority profile
- **verify-later:** none

### RSN-006 — Step-type-aware prompt composition (altitude-aware)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §4 "framing/routing → full why-chain + direction; generation → collapse to a one-line tether … depth is a virtue, not only a failure mode".
- **what:** Prompt composition made altitude-aware: framing/routing gets the full why-chain + direction; generation collapses to a one-line tether; conformance is local; fitness-check and gate get full why-chain.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** salience over presence; why-chain; prompt-composition pattern
- **verify-later:** none

### RSN-007 — Checker model (single-axis parallel checkers)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §5 "run several budgets, each narrow, each fully salient on one axis … curators and the advocate already are".
- **what:** Because one attention budget can't hold detail and breadth at once, run several narrow single-axis checkers fired at decision points, returning terse verdicts that are reconciled. Parallelism produces verdicts, not decisions — arbitration stays singular.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** curators/advocate; multi-author generation; mediator as multi-objective optimiser
- **verify-later:** none

### RSN-008 — Multi-author generation (every concern authors a full solution)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §6 "each perspective is an author, not a guardrail … generative competition along the concern axis".
- **what:** Instead of guardrails, each implicated concern authors its own maximally-on-axis solution; disagreements become worked demonstrations, not complaints. Reuses cascade tier-3/mediator/advocate but competes N attempts at different objectives. Bounded by routing (~2-4 implicated concerns) and counter-proposals-on-deltas.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#6, ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** reliability cascade; mediator as multi-objective optimiser; N-round convergence
- **verify-later:** none

### RSN-009 — Mediator as multi-objective optimiser
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §7 "'Right' = a requirement-relative, defensible balance … not pick, not merge"; "a heuristic floor … full mediation".
- **what:** The mediator finds the requirement-relative balance point among conflicting dimensions using the priority profile, with authored solutions as the extremes that bound the space. Priority is not global; provenance informs weighting, not deference. A cheap heuristic floor settles the uncontested majority but must emit a decision + provenance and be auto-flagged for re-mediation when the why-chain no longer matches its baked-in assumptions.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#7, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** priority profile; multi-author generation; N-round convergence; drift detection; mediator model for competing design concerns (autonomous-build-operate register, ABO-004); requirement-mediation model (autonomy-trust-model register, ATM-002)
- **verify-later:** none

### RSN-010 — N-round convergence (author/checker modes)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §8 "replaces that one hard step with several easier ones … non-convergence is the escalation signal".
- **what:** Replaces single-round reconciliation of N whole solutions with rounds where each active concern reacts to a candidate (satisfied → checker mode; still-needs → author mode), so the active set shrinks. Non-convergence isolates the one genuine value-laden tradeoff to escalate; bounded by an audit-pass-style cap; a concern can withdraw or be dismissed.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#8, ED/FOCUS_salience_and_multi_author_mediation(4).md#12
- **relations:** multi-author generation; checker model; mediator as multi-objective optimiser; N-round candidate ownership
- **verify-later:** none

### RSN-011 — N-round candidate ownership (owned by no concern; rival-base)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §12 "Resolution: the candidate is owned by no concern … seeded deliberately from the profile … The honest limitation: this is path-dependent".
- **what:** The per-round candidate is a shared artifact plus a change log, seeded from known-good or the highest-priority dimension's solution, changed only by premised targeted proposals the mediator adjudicates. Path-dependent and biased toward the seed; rival-base proposals are the gated escape from seed bias.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#12, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** N-round convergence; mediator reasoning trace; known-good library
- **verify-later:** none

### RSN-012 — Self-development coding pipeline — coordination positions A/B/C
- **status:** aspirational
- **status-evidence:** FOCUS_self_development(1) header "Status: exploratory. Nothing decided"; §5 "the live disagreement … How do area-owning agents coordinate a change that touches more than one area?".
- **what:** Use AI agents to develop the platform itself (one+ agent per area = one focus doc + its code slice). Unresolved crux: Position A work-item/ownership-serialized, Position B synchronous inter-agent negotiation, Position C a mediated go-between; current lean is C in a spawn-fresh variant (a per-change orchestrator that spawns current area-owners on demand, dissolving the ephemeral-worker staleness problem).
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#5, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#6
- **relations:** mediator routing model; toolchain validator; MASTER control loop; spawn-fresh coordinator
- **verify-later:** existing spawn machinery; coordinator agent_category
