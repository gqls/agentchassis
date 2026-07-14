
<!-- SOURCE: U15_docs019_running_notes.md -->
### Trust ratchet & capability ceiling model
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "Bottleneck is trust, not capability... Automation is a per-capability ratchet, not a switch. Bidirectional ratchet." (NOTES_running_synthesis_principles(59) §Trust, reliability, and the ratchet — no implementation evidence anywhere in this file set).
- **what:** A design framework (never implemented, purely a framing document across all five families' shared preamble) for autonomous build/operate systems: trust is per-(tenant, capability), starts at the most conservative level, and moves on a bidirectional ratchet (losable, not just gainable) governed by a "trust ledger." A capability's ceiling is set by verifiability (can ground truth confirm it) × containment (blast radius), independent of how mature/trusted it currently is; the reliability cascade for any task is reuse → generate+verify → compete+judge → HITL, highest-reliability tier first; de-graduation (tightening) may auto-apply on severe evidence, but graduation (loosening) is always confirm-not-initiate — the core safety asymmetry.
- **sources:** NOTES_running_synthesis_principles(59), NOTES_running_synthesis_v2(36).md, v3(32), v4(39) — shared §Trust/§Build-vs-operate preamble (identical across all four).
- **relations:** Governance/HITL confirm-not-initiate model; onboarding/config three-layer model; requirement-mediation model.
- **verify-later:** No known code implements a "trust ledger" or per-capability ceiling table — treat as pure design framing pending stage-2 verification.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Requirement-mediation model ("right" as balance)
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "'Right' is a requirement-relative balance among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge." (principles(59) §"Right" as balance, not a single answer).
- **what:** A design framing for resolving competing quality dimensions in generated artifacts: authored solutions are treated as extremes that bound a solution space, a mediator finds the requirement-relative point inside it, priority is ordered (not numerically weighted) and modulated by direction-of-travel, and a satisfied concern demotes from "author" to passive "checker" (re-promoting if a later change breaks it) — unifying single-author and multi-author review as two modes of one process. Multi-author deliberation surfaces tradeoffs but cannot itself resolve value-laden conflicts; those still land with a human/authority model.
- **sources:** NOTES_running_synthesis_principles(59) §"Right" as balance (shared preamble across all four non-fixloop families).
- **relations:** Trust ratchet & capability ceiling model; governance/HITL confirm-not-initiate model.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Trust ratchet & capability ceiling model
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "Bottleneck is trust, not capability... Automation is a per-capability ratchet, not a switch. Bidirectional ratchet." (NOTES_running_synthesis_principles(59) §Trust, reliability, and the ratchet — no implementation evidence anywhere in this file set).
- **what:** A design framework (never implemented, purely a framing document across all five families' shared preamble) for autonomous build/operate systems: trust is per-(tenant, capability), starts at the most conservative level, and moves on a bidirectional ratchet (losable, not just gainable) governed by a "trust ledger." A capability's ceiling is set by verifiability (can ground truth confirm it) × containment (blast radius), independent of how mature/trusted it currently is; the reliability cascade for any task is reuse → generate+verify → compete+judge → HITL, highest-reliability tier first; de-graduation (tightening) may auto-apply on severe evidence, but graduation (loosening) is always confirm-not-initiate — the core safety asymmetry.
- **sources:** NOTES_running_synthesis_principles(59), NOTES_running_synthesis_v2(36).md, v3(32), v4(39) — shared §Trust/§Build-vs-operate preamble (identical across all four).
- **relations:** Governance/HITL confirm-not-initiate model; onboarding/config three-layer model; requirement-mediation model.
- **verify-later:** No known code implements a "trust ledger" or per-capability ceiling table — treat as pure design framing pending stage-2 verification.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Requirement-mediation model ("right" as balance)
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "'Right' is a requirement-relative balance among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge." (principles(59) §"Right" as balance, not a single answer).
- **what:** A design framing for resolving competing quality dimensions in generated artifacts: authored solutions are treated as extremes that bound a solution space, a mediator finds the requirement-relative point inside it, priority is ordered (not numerically weighted) and modulated by direction-of-travel, and a satisfied concern demotes from "author" to passive "checker" (re-promoting if a later change breaks it) — unifying single-author and multi-author review as two modes of one process. Multi-author deliberation surfaces tradeoffs but cannot itself resolve value-laden conflicts; those still land with a human/authority model.
- **sources:** NOTES_running_synthesis_principles(59) §"Right" as balance (shared preamble across all four non-fixloop families).
- **relations:** Trust ratchet & capability ceiling model; governance/HITL confirm-not-initiate model.
