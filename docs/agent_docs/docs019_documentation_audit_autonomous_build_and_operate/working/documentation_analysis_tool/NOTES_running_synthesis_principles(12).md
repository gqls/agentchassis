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
- **Apparent contradictions can be graduated states, not bugs.** When two rules seem to conflict ("everything is gated" vs "the mechanical layer is confirmed by reality"), check whether one is the default starting position and the other is what a trust mechanism eventually permits. They may be the same rule at different maturity levels — the gate is the start; graduation relaxes it.
- **Cold start at the most conservative level.** A new (tenant, capability) pair starts at `confirm_every`, never inherited-higher. Trust is earned per tenant from the most cautious starting point.
- **A capability is a unit of trust, not a unit of agent.** Capabilities aren't 1:1 with agents — "propose a schema migration" or "roll back to known-good" may span several. So the capability (with its ceiling) lives in its own catalog, referenced by the agents that exercise it, rather than as an attribute of any one agent. Trust attaches to the capability; agents are how it's exercised.

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
- **One path to the privileged state.** When a transition must be gated (here: `proposed → active`), route it through a single component, not reimplemented in each producer. Multiple implementations of a gate are multiple chances to weaken it; one path (the central confirmer) makes confirm-not-initiate airtight and auditable in one place rather than a discipline spread across agents.
- **Newer supersedes pending.** When a new proposal arrives for a target that already has an unresolved one, the newer wins and the older expires. Better never to confirm something reality has overtaken than to block on a stale proposal. Prefer freshness over queue order when the world can move under a pending decision.
- **A system that records its own changes shouldn't re-investigate the ones it made on purpose.** The tool watches for changes so it stays honest about its own effects — but when it deliberately applies a change a human just approved, it should note that change without treating it as something new to re-check. Tell apart "I just did this on purpose, already reviewed" from "something changed, go look." The first needs no re-work; the second does. (Changes the tool generates but that haven't been reviewed still count as the second kind.)

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
- **When compute-on-read is too expensive, store the value but keep its inputs authoritative.** The middle path between "store (goes stale)" and "recompute every read (costly)": store the derived value for cheap reads (e.g. a capability's ceiling), but treat its inputs (verifiability, containment) as the source of truth, so a change to an input triggers a gated re-proposal of the stored value rather than silent drift. Same discipline as compute-on-read — the inputs are authoritative — differing only on whether the value is recomputed each read or stored and refreshed on input change.
- **State vs history are different artifacts.** State is mutable, derived from history, and is what's *used* at runtime (the trust ledger). History is immutable, append-only, and is what's *audited* (the decision log). Same source, different access patterns. Mixing them contaminates both — reading old state means walking history; reading history means walking history; the present is always the present.
- **The bidirectional ratchet is asymmetrically governed.** Tightening (de-graduation, lowering trust) may auto-apply with notification when evidence is severe. Loosening (graduation, raising trust) is always confirm-not-initiate. The asymmetry is the safety property: losing trust is reversible; falsely gaining trust permits mistakes to apply unsupervised. Make it easy to be cautious, hard to be incautious.
- **The ceiling is structural; the level is built.** A capability's ceiling (the max trust it can ever reach) is determined by its verifiability + containment — a property of the capability, not the tenant. The trust level a tenant has reached for it is built by use and lives on the per-tenant ledger row. Same capability, different levels for different tenants, but the same ceiling for all.
- **In-band events close the loop on self-modification.** When the tool itself causes a change (the bundle builder applies a confirmed code change; a layer agent flips a row to active), it emits a change event so the maintenance agent sees its own effects. Otherwise the tool's own changes evade the drift detector and the decision log — the system would become blind to itself. State changes emit; computed-view refreshes don't.
- **The trigger filter is computed from the mechanical config.** Compute-on-read applied to routing — when the mechanical config changes (docs move, new code root added), the filter updates automatically on next event. No separate maintained mapping to drift out of sync with reality.

---

## Process and adoption

- **Two goals on one ramp.** The verification harness and bundle that improve today's workflow are literally the first rungs of the autonomy ratchet — no fork between "fix the workflow" and "automate."
- **Walk one capability through the whole machine before generalising.** Vertical slice first; horizontal expansion on proven rails.
- **Dogfood the ratchet.** Each new piece is itself adopted under the trust model it implements: starts at confirm-every, graduates on evidence.
- **Continuous with today, not a replacement.** The current chat workflow is the starting trust level; adoption progressively automates it.
- **Design for the optimal, deliver incrementally.** Early phases expose the interfaces the cascade / checkers / mediator will plug into, so reaching the optimal is adding components, not rebuilding. The seams are the way to get to optimal sooner.
- **Review contracts against each other, not in isolation.** Each contract can read fine alone while quietly disagreeing with another — a reference shape that fits two tables but not the third, a status used in one place and absent from the enum. The cross-contract reading is the check that catches these; do it before code.
- **Verify reuse claims against live schemas (schema-before-SQL, applied to contracts).** A contract that claims to mirror or extend an existing table is an assumption until the real schema is inspected. Correct the contract to match reality, not the reverse. Design-level reuse claims need the same verification as the SQL would.
- **Checking reuse claims reveals reuse you didn't assume.** Verification isn't only validation — inspecting the real `site_work_items` turned up `approval_mode`, `pipeline`, `depends_on`, retry counters, and `item_key` dedup that the contract had reinvented or omitted. The act of checking often finds *more* to reuse than the claim assumed; budget the inspection as discovery, not a rubber stamp.
- **Reuse-before-recreate applies to control mechanisms, not just functions.** The confirm-not-initiate gate was being designed as new machinery when `approval_mode` already encoded exactly that in the existing table. Before building a gate, a queue, a status lifecycle, check whether the system already encodes it — control mechanisms hide in existing schemas as readily as helper functions do.
- **Design artifacts conform to the house style too.** The chassis uses text+CHECK (not native enums), `version`+`previous_version_id` (not `supersedes`), and `deleted_at` (not `status=archived`). A new contract must match these storage-layer conventions, the same way new code matches the code conventions — the conventions agent's remit extends to the schema layer.
- **Scope a shared shape to what it describes.** The common provenance shape belongs on config-entry tables, not on events, logs, state, or queues. Don't universalise a shared structure onto things it doesn't describe — confirming intentional scoping prevents a later "fix" that forces false uniformity.
- **Checking the real system can prove an assumption wrong, not just a name.** Looking at the real `agent_definitions.capabilities` showed it holds loose descriptive tags, not the references the contract assumed — so the whole idea of how two things connected was wrong, not just a label. Check the real thing before wiring anything to it; you may be wrong about the connection itself.
- **If one shared shape is undefined, look for its neighbour.** The bundle shape was missing in exactly the way the active-config shape had been — one step further along the same path. When you find one undefined shared artifact, check the things just upstream and downstream of it; the same gap tends to repeat.
- **Fail loudly; don't quietly tolerate the unexpected.** The house rule (from the naming work) is that readers should not paper over odd inputs and carry on — that just hides the real problem. New parts (the bundle builder, the confirmer, the trigger filter) should stop and surface an unexpected shape, not normalise it away.
- **A strong rule usually has a scope; don't stretch it over everything.** "One path to active" covers the config tables; the trust ledger is a separate domain with its own gated flow. The asymmetric de-graduation isn't a hole in the rule — it's outside the rule's scope. When a rule feels universal, name what it actually covers.
