# Register — onboarding-config

22 concepts, consolidated from 50 raw extractions (25 unique blocks, each present
twice in the source cluster file due to mechanical duplication in the input) across
units U01, U15, U16, U17a, U18, U20, U21.

### ONB-001 — Domain submission tiers and mission/roadmap briefs (domain-submitter entry point)
- **status:** deployed
- **status-evidence:** "The domain-submitter agent is the entry point for all new site builds … Three tiers of domain submission" (WM/001_development_guide(0), unit U17a); current workflow with persist steps documented again in the numbered core docs (001(5), unit U01) — same concept confirmed independently by two doc sets.
- **what:** domain-submitter is the entry point for new site builds: it creates the site record, persists to site_specs, and emits the first needs_domain_research work item. Three tiers of submission: domain-only; domain+objective hint; domain+mission/roadmap (structured JSON for machine consumers, plus plain-text mission_brief/roadmap_brief that the classifier/planner actually read — briefs must be plain text parseable by small models). The mission/roadmap tier supports any pre-planned site (e.g. vonc.com/Spark), bypassing the classifier's domain-discovery step; persist steps skip gracefully via error_step when fields are absent.
- **sources:** 001(5)#Domain Submission, 007#Mission-Driven Sites (unit U01); WM/001_development_guide(0).md#domain-submission-trigger-script-reference, WM/007_adoption_pipeline_v3.md#mission-driven-sites (unit U17a)
- **relations:** classifier weighting of inputs (028); vonc/Spark pattern; adoption modes; build_queue domain queue (parallel/older entry path — reconcile)
- **verify-later:** domain-submitter agent definition; site_specs mission/roadmap aspects

### ONB-002 — build_queue domain queue with direction spectrum
- **status:** partial
- **status-evidence:** P1 (marked "a bit out of date but still has merit"); P2 depends on it for POST /sites; seed_build_queue named in 032/other docs as real
- **what:** build_queue rows (domain, direction jsonb, status, batch, priority); direction spans null → objective hint → full brief (skip research+briefing) → adopt_from → fork_from (specs pre-populated). seed_build_queue takes N, ensures site records, writes initial specs, inserts the appropriate first work item; pacing by batch size. Initial chain: needs_domain_research → needs_briefing → needs_site_plan with spec outputs per handler.
- **sources:** P1#Domain Queue, #Initial Build
- **relations:** public API POST /sites; domain-submitter (newer entry path — reconcile)
- **verify-later:** build_queue table + seed action exist; relation to domain-submitter

### ONB-003 — Three-layer onboarding/config model (mechanical / conventions / intent)
- **status:** aspirational
- **status-evidence:** Framed identically in two independent docs — a "shared preamble" principles synthesis ("The config has three layers with different derivability: mechanical (discovered + probed), conventions (inferred or doc-sourced — confirmed), intent (elicited)", unit U15) and a dedicated plan doc explicitly marked "Status: plan" (PLAN_onboarding_config_derivation, unit U16) — no implementation claimed in either.
- **what:** Onboarding/config treated as three separate problems with different confirmation mechanisms and derivability: the mechanical layer is discoverable (inspected + probed; low stakes, confirmable by reality, climbs the trust ratchet fastest); conventions are inferred or doc-sourced (a strong draft but weak authority — code shows what it does, not what it should do — so inferred-then-confirmed even in docs-authoritative mode, since hallucinated conventions would manufacture fake drift); intent and standards are elicited, never derivable from source (the tenant is the source, and capturing it is the tool's distinctive value), captured progressively rather than as an upfront tax.
- **sources:** NOTES_running_synthesis_principles(59) §Onboarding and config (unit U15); PLAN_onboarding_config_derivation.md#1; 001_onboarding_discussion.txt (unit U16)
- **relations:** Trust ratchet & capability ceiling model; doc claim-verification convention; the five onboarding agents; docs-authoritative decision
- **verify-later:** n/a (plan/framing only)

### ONB-004 — Progressive onboarding — a ramp, never "done"
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §2 — plan.
- **what:** A tenant gets value from the mechanical layer alone (fresh code context, signatures, reuse search, schema) before any intent is captured; conventions and intent fill incrementally and the tool deepens as they arrive. Onboarding tracks the repo forever — active-with-pending is the steady state, and leaf-level intent is captured just-in-time during use rather than as a setup tax.
- **sources:** PLAN_onboarding_config_derivation.md#2; PLAN_onboarding_agent_specs(6).md#3.7,#4.3
- **relations:** intent-elicitation agent; config-maintenance agent
- **verify-later:** n/a (plan)

### ONB-005 — Config as a maintained artifact (the wizard is the first pass; the lifecycle is the deliverable)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §3 — plan.
- **what:** The derived config drifts as the repo changes, so it gets the standards' own upkeep machinery: periodic re-derivation with divergence flagging, confirm-not-initiate on proposed changes, and per-entry provenance (discovered/inferred/supplied) determining trust and change authority. "Onboarding as a first-class deliverable" means this lifecycle, not a good setup script.
- **sources:** PLAN_onboarding_config_derivation.md#3; 001_onboarding_discussion.txt
- **relations:** config-maintenance agent; active-config provenance shape
- **verify-later:** n/a (plan)

### ONB-006 — Inference quality scales with codebase quality — surface uncertainty
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §4 — named tension, design mitigation.
- **what:** On a messy repo, convention inference confidently drafts the repo's bad habits, and confirming that codifies the mess — so the more a tenant needs the tool, the less their repo can teach it. Mitigation: surface inconsistency as questions to resolve, never a silent majority pick; inconsistency found during onboarding is itself valuable output ("your conventions aren't actually conventions").
- **sources:** PLAN_onboarding_config_derivation.md#4; 001_onboarding_discussion.txt
- **relations:** conventions agent; docs-authoritative mode
- **verify-later:** n/a (plan)

### ONB-007 — Docs-authoritative conventions for our own repo (the free drift audit)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §5 "Ours (decided): docs-authoritative" — the decision is recorded; the audit has not run.
- **what:** Source-of-truth for conventions is chosen per tenant by doc availability; for our repo, 001/003/the naming FOCUS are authoritative and code is read only to find disagreements. Each disagreement is recorded, not silently resolved — the set is a free audit of where the codebase drifted from its own documented standards, the drift detector's first run, on us. Our own onboarding is the template, not a special case.
- **sources:** PLAN_onboarding_config_derivation.md#5,#7; 001_onboarding_discussion.txt
- **relations:** conventions agent; drift audit three-bucket output
- **verify-later:** whether any drift audit ran

### ONB-008 — Conventions agent (extract-cite-confirm, then audit)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1 — spec only.
- **what:** Owns the conventions layer: extracts discrete convention atoms from the standards docs (each citing its exact doc span — extraction is inferred-then-confirmed, because auditing code against an invented convention manufactures fake drift, the one failure that would discredit the audit), gets the set human-confirmed BEFORE any audit, then checks code and records disagreements with location/convention/tier/confidence and a default disposition (code-drifted, doc-drifted, or legitimate exception — human confirms). Accepted exceptions are remembered so audits become incremental.
- **sources:** PLAN_onboarding_agent_specs(6).md#1
- **relations:** three checking tiers; docs-authoritative decision; check_*.go validators
- **verify-later:** n/a (spec)

### ONB-009 — Three checking tiers + three-bucket audit output (coverage honesty)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 — spec; the pattern recurs in maintenance §5.5.
- **what:** Conventions (and drift) are checked at three tiers: deterministic (static check settles it → violations), heuristic proxy (a measurable indicator flags candidates, not violations — "where to look, not what's wrong"; an optional LLM pass is still only a candidate flag, never a verdict), judgement-only (no proxy → reported as a coverage gap). The audit reports three numbers, never one — a clean tier-1 count beside many unchecked tier-3 conventions is a partial audit with known limits, and must say so. Companion role split: un-auditable conventions still serve as generation guidance (an atom can be audited, guiding, or both).
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#1.6; #5.5
- **relations:** conventions agent; config-maintenance drift tiers; LLM-as-candidate principle
- **verify-later:** n/a (spec)

### ONB-010 — Convention coverage IS capability reliability
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 closing — spec insight.
- **what:** When a bundle capability rests on a manual convention (log-correlation needs orchestration_id in every log line), the capability is only as reliable as the convention's coverage, not its existence. For any capability-bearing convention the audit reports how completely it is followed, and gaps surface as fixable (add the missing log statements) rather than hard limits — even on our own codebase, where the structure exists but coverage is unverified.
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#2.9
- **relations:** codebase-conditional capabilities; runtime evidence by orchestration_id
- **verify-later:** an orchestration_id logging coverage scan

### ONB-011 — Stack-discovery agent (inspect → interpret → declared probe plan → probe → confirm)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2 — spec only.
- **what:** Owns the mechanical layer: read-only inspection emits facts; interpretation ("this Makefile target is probably the test command") emits proposals with confidence — the subtle point being that interpretation has uncertainty even at the mechanical layer; a declared probe plan (the security contract, kept even for our own use as audit) precedes sandboxed probes; probe results update confidence. A failing build is useful output, candidate-only interpreted, never fixed by this agent. The output document carries per-entry source/confidence/probe-result with uncertainties listed separately. Also records the structural facts bundle capabilities depend on (§2.9).
- **sources:** PLAN_onboarding_agent_specs(6).md#2
- **relations:** confirmation by reality; sandboxing envelope; codebase-conditional capabilities
- **verify-later:** n/a (spec)

### ONB-012 — Confirmation by reality (the mechanical layer climbs the ratchet first)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.8; contract-set review §3.1 records the reconciliation.
- **what:** The mechanical layer can be confirmed by observation (the probed command actually works) — the strongest confirmation any config layer carries — so stack-discovery is the natural first capability to graduate past confirm_every. Reconciled with the gate: probe success is initially strong evidence inside the work-item gate (near-rubber-stamp, human still activates); only after trust-ledger graduation does probe success auto-activate. The gate is the starting position; graduation relaxes it — not a bypass.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.8; FOCUS_contract_set_review.md#3.1; PLAN_active_config_schema(3).md#5
- **relations:** trust ledger; confirm-not-initiate
- **verify-later:** n/a (design)

### ONB-013 — Sandboxed probing — the tenant-code security envelope
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.6: "gating for the service phase: no tenant code runs until sandboxing is solid."
- **what:** The first agent that may execute tenant code does so inside an ephemeral sandbox: repo mounted read-only, restricted network, time limit, no persistent state; the emitted probe plan is the contract the sandbox approves/restricts/denies per command. The Tier-C security concern made concrete; same gate applies to Phase-2 verification running tenant code.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.6; PLAN_context_assembly_tool_and_service(2).md#6
- **relations:** stack-discovery; service phase
- **verify-later:** n/a (unbuilt)

### ONB-014 — Intent-elicitation agent (progressive, value-returning interview)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §3 — spec only; reuse target (briefing_questionnaire column) verified real in FOCUS_schema_verification_findings §3.
- **what:** Captures the why-chain, per-node priority profiles and direction-of-travel via an interview that interleaves proposal-confirmation (where evidence exists — low friction, anchoring risk mitigated by citing the evidence so proposals are contestable) with free elicitation (blank page, unavoidable). Every exchange returns value (the captured piece changes the next bundle/mediation); the interview is not finite — leaf intent is captured just-in-time in the flow of work. A descendant of the briefing questionnaire / intake orchestrator, pointed at a codebase. Capture and use are separate roles (the user-rep advocate consumes what this captures). Open: detecting rubber-stamping.
- **sources:** PLAN_onboarding_agent_specs(6).md#3; FOCUS_schema_verification_findings.md#3
- **relations:** onboarding orchestrator; objectives table; user-rep advocate (hitl category)
- **verify-later:** n/a (spec)

### ONB-015 — Onboarding orchestrator (dependency-graph flow; active-with-pending)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §4 — spec only.
- **what:** Coordinates the three layer agents: stack-discovery first (both others depend on its mechanical config), conventions and intent in parallel (independent of each other) — sequencing follows dependencies, not policy. Routes all proposals through confirm-not-initiate; surfaces a compact onboarding-state artifact (per-layer confirmed/partial/blocked, pending, drift-audit counts); a blocked layer doesn't stop the others; a tenant walking away pauses cleanly. Terminal state is active-with-pending, handing over to maintenance — never "fully done".
- **sources:** PLAN_onboarding_agent_specs(6).md#4; FOCUS_onboarding_system_view_check.md#1,#3.4
- **relations:** the three layer agents; config-maintenance handoff; work-items queue
- **verify-later:** n/a (spec)

### ONB-016 — Config-maintenance agent (drift detection as the trust ratchet's signal source)
- **status:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §5 — spec only.
- **what:** After baseline, detects drift across all three layers, event-driven (change-layer diffs) plus a periodic sweep, dispatching to the layer agents for re-checks rather than reimplementing them; targeted re-validation (implicated-only recheck) instead of full sweeps. Drift evidence uses the same three tiers; surfacing is prioritised to avoid alert fatigue (high-impact deterministic first, heuristic in paced batches, freshness nudges background). Its deeper role: sustained no-drift is graduation evidence and repeated drift is de-graduation evidence — without this agent the bidirectional ratchet has nothing to act on at the right timescale.
- **sources:** PLAN_onboarding_agent_specs(6).md#5; FOCUS_onboarding_system_view_check.md#2
- **relations:** trust ledger; change-layer integration; published-reasoning gap detection
- **verify-later:** n/a (spec)

### ONB-017 — Active-config schema (four tables, computed-on-read effective values)
- **status:** aspirational
- **status-evidence:** PLAN_active_config_schema(3) "Status: contract specification"; corrected to chassis conventions after schema verification.
- **what:** The load-bearing contract: tenant_configs (scope-holder row per tenant, created directly at init — not a gate violation), mechanical_config (one JSONB row, per-field embedded provenance), standards (flat concern atoms with scope constitution/domain/leaf, applies_to change types, rule/rationale/check/check_kind), objectives (nested why-chain nodes with priority_profile, direction_of_travel, standing_concerns). A common provenance shape (source/source_ref/confidence/status/last_verified_at/verified_by/freshness_until/version/previous_version_id/deleted_at) across all layers so consumers reason uniformly. Effective priority profile is computed at read time by walking root→node (store authored differences, compute effective on read); acyclicity must be enforced on write AND the walk bounded, since a human can confirm a cycle. The constitution is a view over standards WHERE scope=constitution, not a table. Two atom trees deliberately kept distinct: flat concern tree vs nested objective tree.
- **sources:** PLAN_active_config_schema(3).md; FOCUS_onboarding_system_view_check.md#3.1,#3.7; FOCUS_pre_build_edge_cases(1).md#1.1
- **relations:** all six contracts hang off it; bundle authored layer reads it
- **verify-later:** whether any of the four tables exist in clients_db

### ONB-018 — Governed vocabularies and the hand-authored first constitution (prerequisites)
- **status:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §6 — named prerequisites, "currently assumed, not called out"; a thin_slice_constitution.md flat file exists and rides in every bundle.
- **what:** The concern taxonomy (standards.concern) and priority dimensions are fixed vocabularies the conventions/intent agents classify INTO, so they must be authored before those agents run. The first constitution is hand-written from 001/003 + working preferences (the tool that would help write it doesn't exist yet); the thin-slice flat-file constitution is its interim form, later becoming standards rows with scope=constitution. Also: "us" is a real tenant row, not a sentinel, so single-tenant exercises the multi-tenant code path.
- **sources:** FOCUS_pre_build_edge_cases(1).md#6; PLAN_active_config_schema(3).md#1,#3.1; tasks/gameslink bundles (constitution section present)
- **relations:** active-config schema; thin-slice-first
- **verify-later:** thin_slice_constitution.md content vs standards rows

### ONB-019 — build-briefing-agent (spec-reading briefing)
- **status:** deployed
- **status-evidence:** 050 definition, "Distinct from existing briefing-agent (v1) which... receives questionnaire directly as input. This version reads from site_specs."
- **what:** Handler for needs_briefing: answers the briefing questionnaire autonomously from site_specs identity + classification (no human), writes aspect "briefing", creates needs_site_plan. Marks the shift from HITL-driven briefing to spec-derived config; explicitly positioned as the successor generation to the early briefing-agent (ONB-020) and distinct from the per-builder questionnaire path (ONB-022).
- **sources:** 050_build_briefing_agent.sql
- **relations:** v1 briefing-agent (superseded for this path, ONB-020); build-site-planner downstream
- **verify-later:** briefing aspect shape

### ONB-020 — Briefing agent (early industry-brief / clarifying-question stage, pre-questionnaire)
- **status:** partial
- **status-evidence:** docs004 SQL (README.021/023, unit U20) shows it live as the first pipeline stage generating a brief JSON; docs005/006 (unit U21) shows a later revision that "sits before chief-strategist" with a human-approval pause, then evolving into the briefing-agent + per-builder briefing_questionnaire architecture — both eras superseded by the current onboarding-config/config-derivation direction.
- **stage2-verified (2026-07-14):** superseded → partial — Claimed successor 'onboarding-config PLAN_onboarding' has zero non-doc Go/SQL hits (grep -rln onboarding --include=*.go only matches unrelated companies_house_vertical_profiles.go). Meanwhile the era-2 descendant mechanisms are live: request_human_input wired in platform/orchestration/actions/registry.go and hitl_re...
- **what:** The original briefing-agent concept, documented across two eras before being superseded. Era 1 (docs004): an LLM turns domain+objective into a comprehensive structured brief JSON — industry inference with confidence, audience demographics/psychographics, brand tone/personality/voice examples, value proposition/key messages/USPs, recommended sections, theme recommendation with semantic tags, content guidelines, monetisation model and ad zones. Era 2 (docs005/006): an agent inserted before the strategist takes raw user input (domain, rough objective), asks clarifying questions, and outputs a structured brief (audience, tone, USPs, competitors, key messages) with a human-approval pause — this era evolved into the briefing-agent plus per-builder briefing_questionnaire with interactive (HITL) and auto (LLM-infer) modes.
- **sources:** docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md, #README.023.specialist_site_architects.md (unit U20); docs005_briefing_agent_domain_authority/README.0130.briefing_agent.md, docs006_workflow_builder/011_working_landing_page_builder.md#Briefing-Agent, #003_current_state_of_agents.sql#3-BRIEFING-AGENT (unit U21)
- **relations:** site classifier; questionnaire pattern; per-builder briefing questionnaires (ONB-022, direct successor); build-briefing-agent (ONB-019, later spec-reading successor); intake orchestrator; successor: onboarding-config PLAN_onboarding docs
- **verify-later:** briefing-agent row in agent_definitions; whether brief JSON shape survives in site_specs; current intake workflow

### ONB-021 — Intake orchestrator with two HITL gates and per-group briefing questionnaires
- **status:** partial
- **status-evidence:** 029.intake_and_groups.sql implements schema (briefing_questionnaire column), site-classifier, intake-orchestrator group, landing/content builder groups; Go actions written (request_human_input with skip conditions, fetch_group_questionnaire); registry additions still listed as "needed".
- **what:** A two-stage front door: classify project (site_type + recommended group) → HITL-1 confirm type → fetch the *target group's* briefing questionnaire (stored in agent_group_definitions, keeping the briefing agent generic) → execute questionnaire (LLM-inferred or human-answered) → HITL-2 review brief → spawn_group dynamically dispatches the chosen builder. HITL points have skip conditions (hitl_mode=auto) for automated runs. Likely the same underlying agent concept as the hitl category's "intake-orchestrator" entry (see register/hitl.md HITL-006), documented via a different source file (029.intake_and_groups.sql vs 002_intake_orchestrator.sql) — kept as a separate entry here since it is not certain they are the exact same migration generation.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; #025.agent_group_discussion#the-intake-orchestrator; #028.agent_group_selection_and_workflow.md
- **relations:** await_approval mechanism (reused); successor: onboarding-config PLAN_onboarding / config derivation; HITL-006 (hitl register, likely same agent)
- **verify-later:** intake-orchestrator group row; request_human_input/fetch_group_questionnaire in registry; briefing_questionnaire column

### ONB-022 — Per-builder briefing questionnaires
- **status:** deployed
- **status-evidence:** docs006/002 full questionnaire JSON on landing-page-builder and content-site-builder definitions; docs007/001 contrasts landing (10 conversion fields) vs brochure (15+ corporate fields).
- **stage2-verified (2026-07-14):** superseded → deployed — briefing_questionnaire column and fetch_agent_questionnaire action are live: platform/orchestration/actions/fetch_agent_questionnaire.go:141,236 selects briefing_questionnaire; registered in registry.go:543 and actioncheck/local_actions.go; also read in spawn_group.go:349,449,458,494. Claimed successor 'onboarding-c...
- **what:** Each builder agent definition carries a `briefing_questionnaire` JSONB (sections of typed questions — brand, value proposition, conversion, social proof for landing; company, services, leadership, case studies for brochure). `fetch_agent_questionnaire` retrieves the correct questionnaire for the chosen builder, and the briefing agent fills it via HITL or LLM inference.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Step-2; docs007_brochure_builder/001_brochure_builder_plan.md#Questionnaire-Differences; docs006_workflow_builder/003_current_state_of_agents.sql
- **relations:** briefing agent (ONB-020); site classifier; reviewed_brief; intake orchestrator (ONB-021, group-level variant of questionnaire storage)
- **verify-later:** briefing_questionnaire values in agent_definitions; fetch_agent_questionnaire action in Go
