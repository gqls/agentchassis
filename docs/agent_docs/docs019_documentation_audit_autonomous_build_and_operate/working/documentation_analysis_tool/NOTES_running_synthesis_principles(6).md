# NOTES — Running Synthesis: Framings and Principles

**Status:** running. A list of synthesis points and design framings that have crystallised across the conversation, captured as reusable lenses for further design work, not as a final theory. Added to as new ones surface.

Organised by theme. Each item is a principle with a short rationale, brief enough to be a reusable lens rather than a paragraph of philosophy.

---

## Context substrate

- **Documentation is code.** Standards, intent, and trajectory are operational inputs to code generation, not passive reference. Versioned, drift-tested, composed deterministically.
- **Authored vs derived.** Two epistemic categories: owner + lifecycle (can be wrong, needs maintenance) vs no-owner true-right-now readout (can't be wrong, only current or superseded). Source code sits on the line — authored, but also ground truth.
- **The change layer between.** Diffs of code and logs are derived (auto-generated, true by being actual) but narrative (tell the story of what changed) — the natural audit and learning surface.
- **References, not copies.** A chat full of pasted code holds copies; a chat that fetches the file when it needs it holds references. Authored layers point at derived rather than paraphrase, so they don't drift when reality moves.
- **Salience over presence.** LLMs lose the bigger picture because local detail is more salient mid-reasoning, not because text fell out of the window. The lever is salience at the moment of decision.
- **Right altitude at the right moment.** Depth is a virtue, not only a failure mode. Blanket-mission-everywhere degrades implementation steps; selective surfacing serves both correctness and bigger-picture fidelity.
- **Two staleness modes, two fixes.** Authored drift: keep authored thin and pointer-rich. Derived snapshot-staleness: fetch at reasoning time, not paste-time.

---

## Trust, reliability, and the ratchet

- **Bottleneck is trust, not capability.** Reframes "make it autonomous" as a reliability problem: bound uncertainty per step to progressively remove the human.
- **Automation is a per-capability ratchet, not a switch.**
- **Bidirectional ratchet.** Trust can be lost, not only gained. This is the safety property that lets the ratchet advance at all — graduating a capability isn't an irreversible bet.
- **Trust ledger as the master knob.** One control surface governing both how conservatively something is produced (cascade floor) and whether it applies without a human (gate).
- **Verifiability + containment.** Two factors; both needed to climb; the weaker holds the ceiling. Verifiability = ground truth; containment = blast radius / reversibility.
- **Ceiling vs maturity.** High ceiling + low maturity = invest (it converts to ratchet progress). Low ceiling regardless of maturity = invest in good HITL ergonomics, not in removing the gate.
- **Reliability cascade per task:** reuse → generate+verify → compete+judge → HITL. Highest-reliability tier first. **Reuse-before-recreate is load-bearing for reliability, not just tidiness.**

---

## "Right" as balance, not a single answer

- **"Right" is a requirement-relative balance** among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge.
- **Authored solutions are extremes that bound the space.** The mediator finds the requirement's point inside it.
- **Priority is not global.** It comes from the why-chain's priority profile, modulated temporally by direction-of-travel.
- **Order, not numeric weights.** Avoid false precision; external events arrive as "X now outranks Y," not as weight tweaks.
- **Drop-out as mode change.** A satisfied concern demotes from author to checker (passively verifying); re-promotes if a later change breaks it. **Unifies the checker and multi-author models as two modes of one process.**
- **Convergence separates resolvable tensions from the irreducible one.** Non-convergence is the escalation signal — it identifies the genuine tradeoff that needs a judgement, with everything else settled around it.
- **Multi-author surfaces tradeoffs vividly; it cannot resolve value-laden conflicts.** Option-generation engine, not decision engine. Decision still lands in the authority model.
- **Candidate owned by no concern.** Shared artifact + change log, seeded deliberately from the profile's dominant dimension. Rounds = premised diffs the mediator adjudicates. Path-dependent — converges near the seed; rival-base proposals are the gated escape for structurally different solutions.

---

## Governance and HITL

- **Confirm-not-initiate.** Decision reasoning is agent-led; the human confirms via a decision package (summary + tradeoffs + explicit choices). The hard job becomes "agent explains well enough that a human can confirm responsibly," not "human authors from scratch."
- **Every decision publishes its reasoning, not just its outcome.** Foundational. Drift detection, the N-round process, and the trace all stand on it. "Chose X because [premise]" is reusable; "chose X" teaches nothing later.
- **Drift detection = gap between a decision's logged premise and the current premise.** Possible only because reasoning is published.
- **Heuristic decisions emit decision + provenance.** Information availability, not information presence. A heuristic that decides invisibly hides drift; a legible one lets later judgement start from "here's what was already decided and on what basis."
- **Provenance informs weighting, not deference.** Knowing which concern authored a solution shapes how to read it; authority comes from the requirement-relative priority, not the loudest voice.
- **Two precedence directions in inheritance.** Normal entries are child-wins (local refinement); sealed constraints are ancestor-wins (legal floors, mission non-negotiables). Sealing prevents a leaf's prior relaxation defeating a new law.
- **Three resolutions to a disagreement: code drifted / doc drifted / legitimate exception.** The default presumption is set by the source-of-truth choice (docs-authoritative → default = code drifted), but the human can pick any. **Default ≠ only.** This is where docs-authoritative earns its keep without becoming rigid.
- **The user-rep is a different *kind* of advocate.** Curators advocate for concerns (correctness, consistency); the user-rep advocates for intent (fidelity to what the user wants). The coordinator arbitrates between them; the user-rep is the check on the coordinator's framing power.
- **"Sufficiently adopted" via both self-judge and mediator-judge**, with the residual recorded if the mediator dismisses an unsatisfied concern. Override doesn't erase; it logs.

---

## Build vs operate

- **The two routes unify at the context layer, not at "website is code."** Both run on thin authored + thick live derived; vertical (infra vs web) changes which slices, not the machinery.
- **Build vs operate asymmetry.** Build is isolatable (competition safe); operate is live/stateful (competition risky, lean on known-good + canary + rollback).
- **Rollback can be trusted autonomously earlier than roll-forward**, because returning to a known-good state is the inherently safe direction.
- **Cross-cluster: boundary made invisible by design** (same Kafka talk-back, same DNS for remote DB) → theory barely expands. Mostly new capabilities, not new theory.
- **Defect-vs-partition disambiguation.** Quarantine infrastructure-attributed failures from the trust signal, or transient flakiness will de-graduate capabilities that are actually fine.

---

## Onboarding and config

- **The config has three layers with different derivability:** mechanical (discovered + probed), conventions (inferred or doc-sourced — confirmed), intent (elicited). They are not one onboarding problem; they are three uneven ones.
- **Progressive onboarding.** Mechanical value first, ramp not cliff; never "done" — config tracks the repo.
- **Config is a maintained artifact** with the standards' lifecycle. Drift detection, confirm-not-initiate, per-entry provenance (discovered / inferred / supplied).
- **Inference quality scales with codebase quality.** The messier the repo, the less it teaches; the more a tenant needs the tool, the less their repo can teach it. Surface uncertainty rather than silently picking; inconsistency detected is itself output.
- **Extraction is inferred-then-confirmed even in docs-authoritative mode.** The docs are the source, not extraction's authority. Hallucinated conventions would manufacture drift.
- **Coverage honesty.** Don't imply a clean audit when half the conventions weren't checkable. Report what was checked and what couldn't be — three buckets (deterministic violations, heuristic candidates, unchecked judgement), not one number.
- **Conventions have two roles: audited and guiding.** Some atoms appear in the drift audit; some load into generation context; some both. Un-auditable conventions are not lost — they contribute on the input side rather than the output side.
- **LLM as judgement-checker is itself only a candidate flag, never a verdict.** The same LLM-uncertainty problem the wider cascade exists for. Don't create a weaker authority channel by trusting it as a verdict — narrow where to look, leave the verdict to a human.
- **Exceptions feedback.** Marking a disagreement as a legitimate exception is recorded against location + convention, so the audit doesn't re-flag — it becomes incremental.
- **Our own setup is the template, not a special case.** The docs-vs-code disagreement list from our own onboarding is the drift detector's first run, on us — a free audit of where the code drifted from its own documented standards.
- **Inspection is fact, interpretation is proposal.** Reading a file's contents is fact; deciding what it *means* (which Makefile target is "the test command"; that a directory called `actions/` holds the actions in this codebase's sense) is interpretation, which is inferred-then-confirmed even at the mechanical layer. The pattern applies where it is least expected.
- **Probing is declared.** No execution without a plan emitted first, with rationale. Necessary for sandboxing tenants (the security contract); useful as audit for our own use (what did the agent actually do).
- **Confirmation by reality.** The mechanical layer can be confirmed by probe success — the recorded commands actually running is the strongest confirmation any config layer can carry. This is why the mechanical layer climbs the ratchet fastest of the three: conventions and intent cannot be confirmed by reality and depend on human confirmation as the only available authority.
- **The interview is progressive, never finite.** Up-front elicitation captures the top of the tree; leaves are captured just-in-time as work happens. The intent layer fills out with usage rather than as a setup tax. Initial onboarding stops asking when the top is covered; the rest fills as the tool is used.
- **Value-return per exchange.** Each elicited piece changes the next interaction immediately (capturing the mission tethers the next generation to it; capturing an area's priority order is what the mediator uses in the next tradeoff). The tenant should never feel they have given several answers and gotten nothing back.
- **Proposal-confirmation vs free elicitation.** Lean proposal-confirmation where evidence exists (low friction); free elicitation where it does not. Pure either is wrong — free is the blank-page problem; pure proposal-confirmation anchors the tenant on the agent's guess. Mitigation: proposals cite their evidence so a wrong inference is contestable.
- **Capture vs use are separate roles, separate agents.** The intent-elicitation agent captures intent (onboarding); the user-rep advocate uses intent (runtime decision). Same concept, two roles. Keeps elicitation unpressured by live decisions and advocacy undistracted by capture.
- **The dependency graph is the flow plan.** Sequencing follows dependencies, not process preference: run in series where outputs feed downstream, run in parallel where they don't. Reading the dependencies *is* designing the flow.
- **Onboarding never reaches "fully done."** Active-with-pending is the steady state. Maintenance takes over from "enough to use" rather than "complete." This is the progressive principle applied to the lifecycle, not just to the interview.
- **The orchestrator coordinates and surfaces; it does not extract.** Same boundary applies to any layer-coordinating orchestrator: dispatch to the specialists, route gates to the human, surface state — but do not start doing the specialists' jobs, or boundaries blur and the orchestrator inflates.
- **Same code for our use and for tenants.** Only the entry point and isolation context differ. The engine/config split principle applied to the orchestrator itself.
- **Maintenance is where the ratchet gets its long-run signal.** Drift detection is not a separate upkeep function from the bidirectional trust ratchet — it is the source of evidence the ratchet acts on. Without something playing this role, the ratchet has nothing to move on. Immediate verification pass/fail is the wrong timescale for trust; sustained no-drift and repeated-drift are the right ones.
- **Drift triggers follow the change layer.** Diffs to docs/code/objective-tree nodes are the natural event source; the periodic sweep catches what events missed. No separate sensor — the change layer is the source.
- **Targeted re-validation, not full sweeps.** When something changes high in a tree, check only the descendants whose entries reference or override the changed dimension. Efficient maintenance is implicated-only-recheck — the priority-profile maintenance pattern generalised across the config.
- **Surface drift in priority order, not all at once.** Over-surfacing causes alert fatigue and the signal is ignored; under-surfacing lets drift accumulate. Critical drift immediately; low-impact drift periodically; the tenant's attention is earned, not consumed.
- **Contracts are the bottleneck for coherent implementation.** Agents specced individually can each look fine while quietly disagreeing on the shape of the artifact they all produce or consume. A system-view check is what surfaces these — and the shared contracts have to be settled before implementation, not as you go. The load-bearing one is usually the artifact every agent touches (here: the active config).
- **Compute on read for freshness; log the result at point of use for retrospect.** Derived values (the effective priority profile, the constitution view, the applicable-standards bundle) are computed fresh on each read so ancestor changes propagate automatically without stored derivations going stale. But the result *at the moment it was used* is recorded in the decision log so the historical state is reconstructable. The active value is always fresh; the log is audit, not source — it wasn't authoritative when written and isn't authoritative now. Combined with versioned atoms, this gives full retrospective reconstructability without poisoning the present.

---

## Process and adoption

- **Two goals on one ramp.** The verification harness and bundle that improve today's workflow are literally the first rungs of the autonomy ratchet — no fork between "fix the workflow" and "automate."
- **Walk one capability through the whole machine before generalising.** Vertical slice first; horizontal expansion on proven rails.
- **Dogfood the ratchet.** Each new piece is itself adopted under the trust model it implements: starts at confirm-every, graduates on evidence.
- **Continuous with today, not a replacement.** The current chat workflow is the starting trust level; adoption progressively automates it.
- **Design for the optimal, deliver incrementally.** Early phases expose the interfaces the cascade / checkers / mediator will plug into, so reaching the optimal is adding components, not rebuilding. The seams are the way to get to optimal sooner.
