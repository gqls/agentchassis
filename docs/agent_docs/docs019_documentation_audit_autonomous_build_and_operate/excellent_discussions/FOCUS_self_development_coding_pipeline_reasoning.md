# FOCUS — Self-Development Coding Pipeline: Reasoning Checkpoint

**Status:** exploratory. Nothing decided. This is a record of the reasoning so far so we can branch in different directions and revert if a branch doesn't work out.

**Goal under discussion:** use AI agents to develop the platform itself — searching the code and the design/handoff docs to solve problems and build workflows and actions — rather than (or alongside) the website-build and data-collection pipelines the chassis runs today.

---

## 1. The vision as stated

- One or more agents per area of the project, where an area is roughly one focus document plus the slice of code that document governs.
- Every agent holds the overall project objectives plus the local objectives of the part currently in focus.
- Docs possibly arranged hierarchically and read root → leaf: root = mission and guidelines, leaf = the specific part being worked on.
- Relevant agents contribute to each development step, checking that a change (a) doesn't break their area, (b) fits their objectives, (c) couldn't be done better.

---

## 2. Things that look low-regret regardless of how the open questions resolve

These held up across the discussion and don't depend on the coordination model we pick.

- **Scoped, per-area context is necessary, not optional.** ~4–5MB of source is roughly 1–1.5M tokens (code runs ~3–4 chars/token), so the whole codebase plus docs does not fit in a single 1M context window. An agent holding "everything" isn't on the table. One agent owning one area's docs + code slice is the only shape the token budget allows.
- **Hierarchical docs root → leaf map onto a prompt-composition layering**, not a new invention. This is close to the existing prompt-composition pattern (baseline context + per-area overrides). Root carries mission/guidelines; nested layers carry local objectives.
- **A toolchain validator is needed.** Whatever else we do, a coding agent needs ground-truth feedback: `go build`, `go vet`, `go test`, SQL/migration dry-run, returning structured pass/fail + errors. This is distinct from the contract-style validators used for components today.
- **A repo read/search capability is needed.** Today's STEP ZERO (grep `agent_definitions`, grep `registry.go`, grep for funcs/types) is a manual human step. For an agent to do it, something with the working tree checked out has to search and read files.
- **Deploy must be branch + gate, never hot-apply.** This pipeline modifies the platform that runs the pipeline. A bad change cannot be allowed to break the agents mid-flight. The existing HITL gating + github-actions → backblaze deploy give the pieces; the difference is the gate becomes mandatory and the blast radius is the platform itself.
- **Rework, not rebuild, and not a fork.** The case for a second system or a chassis fork is weak; it would double maintenance and break the reuse principle. If we proceed on the chassis, it's as a new coding domain reusing the existing spine.

---

## 3. What already exists in the chassis and would transfer

Grounded in the design docs (026 component regeneration flow, 004 improvement loop, the prompt-composition pattern, chassis action context):

- **Write → validate → regenerate-with-targeted-feedback loop.** The iteration engine already exists, including loop primitives (loop expansion, `continue_on_error`, `skipToNextLoopIteration`, iteration-output propagation). A coding loop (write → compile → read errors → fix → repeat) is the same shape. The audit-pass cap gives bounded iteration so it doesn't churn forever.
- **"Broken new output never overwrites existing state"** (026). Exactly the right safety rule for code: a change that fails validation must not replace working code.
- **Version history** (`component_versions`, `change_source`) → change history for code edits.
- **Locks** (permanent / timed / review), **optimistic locking** on state with retry/backoff.
- **HITL gating** (`needs_human_review`) → the code-review/merge gate.
- **git adapter → github-actions → backblaze** → the publish step.
- **Dependency-graph reasoning already present.** 026 chooses to UPDATE a component in place rather than INSERT-and-relink, specifically to keep foreign keys resolving. That is the same reasoning as "changing this function signature breaks its callers." The mental model is already there.

## 4. What is genuinely net-new for code

- **Validator changes kind:** from contract checks (`template_closed`, schema/template sync, cross-site contamination) to a toolchain validator. The loop around it exists; the checker is new. This is the most important new piece — it closes the loop with ground truth instead of an LLM's opinion of correctness.
- **Repo read/search** (the file-navigation gap above).
- **Edits against existing files rather than from-scratch artifacts.** Component-creator regenerates a whole template; for a Go file in a module, targeted edits/diffs are safer than whole-file regeneration. Behaviorally different from today, though not enormous.
- **Shared-repo serialization** — see the open question below; this is the deepest structural difference.

---

## 5. The open question (the live disagreement)

**How do area-owning agents coordinate a change that touches more than one area?**

Two positions are on the table. This is unresolved.

### Position A — work-item coordination (async, ownership-serialized)

The shape used by the website pipeline: one owning agent edits its area; cross-area impacts become side-effect work items for the owners of those areas; a gate merges. Serialized through ownership.

- **For:** matches the existing infrastructure and the chassis's stated preference for low coupling and no agent-to-agent coordination calls; keeps logs clear and responsibilities separable; avoids N-way coupling where every agent weighs in on every change.
- **Against (the reservation raised):** the website model works because sites are **independent** — they share no state, so reconciling "in the next cycle" is lossless. **Code changes are not independent.** They touch a single shared mutable repo and are often tightly interrelated. Reconciling tightly-coupled changes asynchronously across cycles may be too slow and may lose the context that made the changes related in the first place.

### Position B — direct communication between responsible agents (synchronous negotiation)

The user's current lean: because changes are closely related, the agents responsible for the affected areas should communicate directly about a change as it's being made, rather than discovering impacts after the fact through a queue.

- **For:** models tight interdependence directly; captures cross-cutting concerns at the moment of change rather than a cycle later; better fits a codebase where clean separation of responsibility may not actually hold.
- **Against / tensions to resolve:** pushes against the chassis's "minimal overlap / no agent calls another agent to coordinate" contract and the messaging model (every agent is an orchestrator; agents reply to the caller's responses topic). It reintroduces the coupling those rules were written to prevent, and risks the many-agents-review-every-step pattern with its log-clarity and termination concerns. Needs a defined protocol: who initiates, who must be consulted, how disagreement resolves, how it terminates.

### The deeper point underneath both

If changes really are closely related across areas, that strains the original premise of "agents with distinct responsibilities overlapping as little as possible." Either the **decomposition** (how we draw area boundaries) needs rethinking so changes are more local, or the **coordination** needs to be richer than work-items. Position B is essentially the bet that clean decomposition isn't achievable for this codebase and coordination has to carry the weight instead.

A hybrid is possible and not yet explored: ownership + a branch, with a synchronous consult step only for changes flagged as cross-area, falling back to work-items for the rest.

---

## 6. Branch points we can pick up from here

1. **Coordination model:** Position A (work-items) vs Position B (direct inter-agent communication) vs a hybrid. *This is the one the user wants to explore next.* If B, the next work is defining the communication protocol (initiation, mandatory consultees, conflict resolution, termination) and reconciling it with the existing messaging/orchestrator contracts.
2. **Decomposition:** how area boundaries are drawn, and whether they can be drawn so changes are mostly local (which would make A viable) or not (which pushes toward B).
3. **Context assembly:** stuff-the-area-slice vs RAG over the codebase. This is the same retrieval question discussed earlier (Files API is storage, not retrieval; 1M window changes the threshold but 4–5MB still exceeds it). Independent of the coordination model.
4. **Own chassis vs off-the-shelf coding-agent framework.** Currently leaning chassis (website design is itself codegen; the existing spine transfers). Revertible.

---

## 7. One-line state

We agree on the low-regret pieces (section 2) and on rework-not-rebuild. The unresolved crux is **section 5: how related changes coordinate** — and the next direction is to explore direct communication between responsible agents (Position B), with section 6 listing the other threads we can return to.
