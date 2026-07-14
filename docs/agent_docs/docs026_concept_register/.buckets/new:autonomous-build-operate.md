
<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### trust ledger (bidirectional trust ratchet, per-tenant per-capability)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Now given concrete shape." — a design document, no implementation claimed. A later/fuller version of this exact contract exists live at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_trust_ledger_contract(4).md (outside this unit's scope).
- **what:** A `trust_ledger` table (one row per tenant×capability) holding a mutable `trust_level` (`confirm_every`/`confirm_exceptions`/`notify`/`autonomous`) that cascade routers read to floor the production tier and gate-policy engines read to decide autonomy — derived from, but distinct from, the immutable append-only `decision_log`. The capability's **ceiling** (max reachable trust, set by verifiability × containment) lives on a separate capability catalog, not the ledger row, so it's a property of the capability, not the tenant. Mutation is asymmetric: graduation (trust up) is always confirm-not-initiate via a `config_work_items` proposal; de-graduation (trust down) may auto-apply with notification on severe evidence — "losing trust is reversible; falsely gaining trust is what allows mistakes to apply unsupervised."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_trust_ledger_contract(1).md
- **relations:** change-layer integration contract (below, feeds the ledger's evidence_summary); governance/HITL principles (below); the fuller live PLAN_trust_ledger_contract(4).md (not in this unit's scope)
- **verify-later:** any `trust_ledger` or `capabilities` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### change-layer integration contract (change_events, trigger filter, in-band emission)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Closes the final contract gap before implementation." A fuller live version exists at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_change_layer_integration_contract(4).md (outside this unit's scope).
- **what:** Defines how code/doc diffs reach a maintenance agent: a first-class `change_events` table (source ∈ git_webhook/polling/in_band/periodic_sweep/manual, at-least-once with commit_id dedup) feeding a **trigger filter computed from the mechanical config** (not stored — so it self-updates when doc/code paths move) that fans out into typed triggers (`conventions_reextraction`, `schema_check`, `code_audit_refresh`, `reuse_index_refresh`, `intent_revalidation`, `freshness_check`). The `in_band` source is the mechanism that "closes the loop on self-modification" — when the tool's own bundle-builder or a layer agent applies a confirmed change, it emits its own change event so the drift detector doesn't go blind to its own effects; a scoped guard prevents a just-confirmed entry from re-triggering on itself while still letting genuine downstream effects (e.g. auditing existing code against a newly confirmed convention) fire.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(3).md (family-latest in this unit), PLAN_change_layer_integration_contract(1).md (delta-checked, no drops)
- **relations:** trust ledger (above); reuse-check retrieval pipeline (below, reuse_index_refresh trigger)
- **verify-later:** any `change_events` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### context substrate model (authored vs derived, salience over presence)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "a list of synthesis points... captured as reusable lenses for further design work, not as a final theory."
- **what:** A framing for how LLM context should be built: documentation/standards/intent are operational inputs to generation, not passive reference, and split into two epistemic categories — **authored** (has an owner and lifecycle, can be wrong, needs maintenance) vs **derived** (no-owner true-right-now readout, can only be current or superseded; source code sits on this line). The **change layer** (diffs) is derived-but-narrative — the natural audit/learning surface. Authored layers should hold **references, not copies** of derived material so they don't drift when reality moves. Two staleness modes need two different fixes: authored drift is fixed by keeping authored content thin and pointer-rich; derived snapshot-staleness is fixed by fetching at reasoning time, not paste-time. LLMs lose the big picture from **salience, not window size** — local detail crowds out context mid-reasoning even when the text is still "in the window."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Context substrate
- **relations:** contextkit toolchain (above); flat-file constitution (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### mediator model for competing design concerns ("right" as requirement-relative balance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document, framed as an unbuilt lens.
- **what:** A model for resolving conflicting design dimensions (fast/secure/generic/simple/functional) as a requirement-relative balance rather than a pick-one-winner or naive-merge: authored solutions are treated as extremes that bound the solution space, and a mediator finds the point inside it that the requirement's priority profile dictates (ordered priority, not numeric weights, since real-world priority shifts arrive as "X now outranks Y"). A satisfied concern demotes from active author to passive checker (re-promoting if a later change breaks it) — unifying "checker" and "multi-author" as two modes of one process. Non-convergence among concerns is treated as the genuine escalation signal, isolating the one real tradeoff that needs human judgement from everything else that settles on its own. Multi-author surfaces tradeoffs vividly but cannot resolve value-laden conflicts — it's an option-generation engine, not a decision engine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §"Right" as balance, not a single answer
- **relations:** governance/HITL principles (below)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### governance/HITL principles (confirm-not-initiate, decision publishing, sealed inheritance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document; principles, not shipped mechanisms.
- **what:** A cluster of governance rules for an autonomous build-and-operate system: **confirm-not-initiate** (agent-led reasoning, human confirms via a decision package, never authors from scratch); **every decision publishes its reasoning**, since drift detection is only possible because premises are logged and can be compared to the current premise; **two precedence directions in inheritance** — normal entries are child-wins (local refinement) but sealed constraints are ancestor-wins (legal floors, mission non-negotiables), so a leaf can't defeat a new law by prior relaxation; **three resolutions to a doc/code disagreement** (code drifted / doc drifted / legitimate exception) with a configurable default presumption that the human can always override; **one path to a privileged state transition** (e.g. `proposed → active`) routed through a single central confirmer rather than reimplemented per producer, so confirm-not-initiate is airtight in one place; **newer supersedes pending** — a fresh proposal for an already-pending target expires the older one rather than blocking on staleness.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Governance and HITL
- **relations:** trust ledger (above); change-layer integration contract (above); mediator model (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### autonomous-system building-block hardening checklist
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed as "edge cases caught before building" in the running-synthesis notes — a pre-implementation checklist, not a verified implementation.
- **what:** A catalog of structural safety patterns for building the autonomous-operate machinery itself, distilled from design review: self-referential structures (trees, version chains) need a cycle guard on write plus a detect-and-fail walk; a multi-step apply (writes + an emitted event) must be all-or-nothing via one transaction with an outbox, or a mid-crash leaves a live row with no log/event; assembling from several tables needs one consistent point-in-time snapshot; "at most one live X per target" must be enforced at every layer down to the underlying row, not just the queue; bulk operations need bulk confirmation (per-item confirm doesn't scale to an onboarding flood); transient/infrastructure failures must be filtered out before they're allowed to lower a capability's trust; derived indexes/caches go stale silently and need a freshness stamp; recovery must not depend on the thing being recovered (the rollback path can't route through the agents it's rolling back); blast radius caps the trust ceiling regardless of verifiability, because self-modification is a residual risk that's managed (conservative early trust, human-in-the-loop, external rollback), not solved.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Building discipline
- **relations:** trust ledger (above); change-layer integration contract (above)
- **verify-later:** none (a design checklist, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### trust ledger (bidirectional trust ratchet, per-tenant per-capability)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Now given concrete shape." — a design document, no implementation claimed. A later/fuller version of this exact contract exists live at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_trust_ledger_contract(4).md (outside this unit's scope).
- **what:** A `trust_ledger` table (one row per tenant×capability) holding a mutable `trust_level` (`confirm_every`/`confirm_exceptions`/`notify`/`autonomous`) that cascade routers read to floor the production tier and gate-policy engines read to decide autonomy — derived from, but distinct from, the immutable append-only `decision_log`. The capability's **ceiling** (max reachable trust, set by verifiability × containment) lives on a separate capability catalog, not the ledger row, so it's a property of the capability, not the tenant. Mutation is asymmetric: graduation (trust up) is always confirm-not-initiate via a `config_work_items` proposal; de-graduation (trust down) may auto-apply with notification on severe evidence — "losing trust is reversible; falsely gaining trust is what allows mistakes to apply unsupervised."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_trust_ledger_contract(1).md
- **relations:** change-layer integration contract (below, feeds the ledger's evidence_summary); governance/HITL principles (below); the fuller live PLAN_trust_ledger_contract(4).md (not in this unit's scope)
- **verify-later:** any `trust_ledger` or `capabilities` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### change-layer integration contract (change_events, trigger filter, in-band emission)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Closes the final contract gap before implementation." A fuller live version exists at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_change_layer_integration_contract(4).md (outside this unit's scope).
- **what:** Defines how code/doc diffs reach a maintenance agent: a first-class `change_events` table (source ∈ git_webhook/polling/in_band/periodic_sweep/manual, at-least-once with commit_id dedup) feeding a **trigger filter computed from the mechanical config** (not stored — so it self-updates when doc/code paths move) that fans out into typed triggers (`conventions_reextraction`, `schema_check`, `code_audit_refresh`, `reuse_index_refresh`, `intent_revalidation`, `freshness_check`). The `in_band` source is the mechanism that "closes the loop on self-modification" — when the tool's own bundle-builder or a layer agent applies a confirmed change, it emits its own change event so the drift detector doesn't go blind to its own effects; a scoped guard prevents a just-confirmed entry from re-triggering on itself while still letting genuine downstream effects (e.g. auditing existing code against a newly confirmed convention) fire.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(3).md (family-latest in this unit), PLAN_change_layer_integration_contract(1).md (delta-checked, no drops)
- **relations:** trust ledger (above); reuse-check retrieval pipeline (below, reuse_index_refresh trigger)
- **verify-later:** any `change_events` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### context substrate model (authored vs derived, salience over presence)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "a list of synthesis points... captured as reusable lenses for further design work, not as a final theory."
- **what:** A framing for how LLM context should be built: documentation/standards/intent are operational inputs to generation, not passive reference, and split into two epistemic categories — **authored** (has an owner and lifecycle, can be wrong, needs maintenance) vs **derived** (no-owner true-right-now readout, can only be current or superseded; source code sits on this line). The **change layer** (diffs) is derived-but-narrative — the natural audit/learning surface. Authored layers should hold **references, not copies** of derived material so they don't drift when reality moves. Two staleness modes need two different fixes: authored drift is fixed by keeping authored content thin and pointer-rich; derived snapshot-staleness is fixed by fetching at reasoning time, not paste-time. LLMs lose the big picture from **salience, not window size** — local detail crowds out context mid-reasoning even when the text is still "in the window."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Context substrate
- **relations:** contextkit toolchain (above); flat-file constitution (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### mediator model for competing design concerns ("right" as requirement-relative balance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document, framed as an unbuilt lens.
- **what:** A model for resolving conflicting design dimensions (fast/secure/generic/simple/functional) as a requirement-relative balance rather than a pick-one-winner or naive-merge: authored solutions are treated as extremes that bound the solution space, and a mediator finds the point inside it that the requirement's priority profile dictates (ordered priority, not numeric weights, since real-world priority shifts arrive as "X now outranks Y"). A satisfied concern demotes from active author to passive checker (re-promoting if a later change breaks it) — unifying "checker" and "multi-author" as two modes of one process. Non-convergence among concerns is treated as the genuine escalation signal, isolating the one real tradeoff that needs human judgement from everything else that settles on its own. Multi-author surfaces tradeoffs vividly but cannot resolve value-laden conflicts — it's an option-generation engine, not a decision engine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §"Right" as balance, not a single answer
- **relations:** governance/HITL principles (below)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### governance/HITL principles (confirm-not-initiate, decision publishing, sealed inheritance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document; principles, not shipped mechanisms.
- **what:** A cluster of governance rules for an autonomous build-and-operate system: **confirm-not-initiate** (agent-led reasoning, human confirms via a decision package, never authors from scratch); **every decision publishes its reasoning**, since drift detection is only possible because premises are logged and can be compared to the current premise; **two precedence directions in inheritance** — normal entries are child-wins (local refinement) but sealed constraints are ancestor-wins (legal floors, mission non-negotiables), so a leaf can't defeat a new law by prior relaxation; **three resolutions to a doc/code disagreement** (code drifted / doc drifted / legitimate exception) with a configurable default presumption that the human can always override; **one path to a privileged state transition** (e.g. `proposed → active`) routed through a single central confirmer rather than reimplemented per producer, so confirm-not-initiate is airtight in one place; **newer supersedes pending** — a fresh proposal for an already-pending target expires the older one rather than blocking on staleness.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Governance and HITL
- **relations:** trust ledger (above); change-layer integration contract (above); mediator model (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### autonomous-system building-block hardening checklist
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed as "edge cases caught before building" in the running-synthesis notes — a pre-implementation checklist, not a verified implementation.
- **what:** A catalog of structural safety patterns for building the autonomous-operate machinery itself, distilled from design review: self-referential structures (trees, version chains) need a cycle guard on write plus a detect-and-fail walk; a multi-step apply (writes + an emitted event) must be all-or-nothing via one transaction with an outbox, or a mid-crash leaves a live row with no log/event; assembling from several tables needs one consistent point-in-time snapshot; "at most one live X per target" must be enforced at every layer down to the underlying row, not just the queue; bulk operations need bulk confirmation (per-item confirm doesn't scale to an onboarding flood); transient/infrastructure failures must be filtered out before they're allowed to lower a capability's trust; derived indexes/caches go stale silently and need a freshness stamp; recovery must not depend on the thing being recovered (the rollback path can't route through the agents it's rolling back); blast radius caps the trust ceiling regardless of verifiability, because self-modification is a residual risk that's managed (conservative early trust, human-in-the-loop, external rollback), not solved.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Building discipline
- **relations:** trust ledger (above); change-layer integration contract (above)
- **verify-later:** none (a design checklist, not a built artifact)
