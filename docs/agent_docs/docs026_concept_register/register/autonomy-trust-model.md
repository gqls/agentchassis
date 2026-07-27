# Register — autonomy-trust-model

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

2 concepts, consolidated from 4 raw extractions across unit U15 (each of the 2 distinct blocks
appeared byte-identically twice within the cluster input file — treated as duplicate copies of
one extraction, not independent corroboration). Both are pure design framing from a shared
preamble repeated across several running-synthesis documents, with no implementation evidence
anywhere in this cluster's source material.

### ATM-001 — Trust ratchet & capability ceiling model
- **status:** aspirational
- **status-evidence:** "Bottleneck is trust, not capability... Automation is a per-capability ratchet, not a switch. Bidirectional ratchet." (NOTES_running_synthesis_principles(59) §Trust, reliability, and the ratchet) — no implementation evidence anywhere in this file set; the same core design is given fuller concrete shape as contract specifications under autonomy-governance (AGOV-004) and autonomous-build-operate (ABO-001).
- **what:** A design framework (never implemented, purely a framing document across all five families' shared preamble) for autonomous build/operate systems: trust is per-(tenant, capability), starts at the most conservative level, and moves on a bidirectional ratchet (losable, not just gainable) governed by a "trust ledger." A capability's ceiling is set by verifiability (can ground truth confirm it) × containment (blast radius), independent of how mature/trusted it currently is; the reliability cascade for any task is reuse → generate+verify → compete+judge → HITL, highest-reliability tier first; de-graduation (tightening) may auto-apply on severe evidence, but graduation (loosening) is always confirm-not-initiate — the core safety asymmetry.
- **sources:** NOTES_running_synthesis_principles(59), NOTES_running_synthesis_v2(36).md, v3(32), v4(39) — shared §Trust/§Build-vs-operate preamble (identical across all four)
- **relations:** governance/HITL confirm-not-initiate model (autonomous-build-operate register, ABO-005); trust ledger + bidirectional ratchet (autonomy-governance register, AGOV-004); trust ledger — earlier draft (autonomous-build-operate register, ABO-001); requirement-mediation model (ATM-002)
- **verify-later:** no known code implements a "trust ledger" or per-capability ceiling table — treat as pure design framing pending stage-2 verification

### ATM-002 — Requirement-mediation model ("right" as balance)
- **status:** aspirational
- **status-evidence:** "'Right' is a requirement-relative balance among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge." (principles(59) §"Right" as balance, not a single answer).
- **what:** A design framing for resolving competing quality dimensions in generated artifacts: authored solutions are treated as extremes that bound a solution space, a mediator finds the requirement-relative point inside it, priority is ordered (not numerically weighted) and modulated by direction-of-travel, and a satisfied concern demotes from "author" to passive "checker" (re-promoting if a later change breaks it) — unifying single-author and multi-author review as two modes of one process. Multi-author deliberation surfaces tradeoffs but cannot itself resolve value-laden conflicts; those still land with a human/authority model.
- **sources:** NOTES_running_synthesis_principles(59) §"Right" as balance (shared preamble across all four non-fixloop families)
- **relations:** trust ratchet & capability ceiling model (ATM-001); governance/HITL confirm-not-initiate model (autonomous-build-operate register, ABO-005); mediator model for competing design concerns (autonomous-build-operate register, ABO-004); mediator as multi-objective optimiser (reasoning register, RSN-009)
- **verify-later:** none — pure design framing
