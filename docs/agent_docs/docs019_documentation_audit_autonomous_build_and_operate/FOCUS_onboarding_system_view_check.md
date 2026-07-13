# FOCUS — Onboarding Agents: System-View Check

**Status:** review across §1–§5 of `PLAN_onboarding_agent_specs.md`. Asks: does the data flow cleanly, are the contracts between agents right, is anything missing for the five to fit into the broader machine. Findings drive the recommended order in §5 — contracts before code.

---

## 1. The data flow as specced

**Onboarding (orchestrator-driven):**
tenant arrives → orchestrator (§4) → stack-discovery (§2) runs → tenant confirms mechanical → conventions (§1) and intent (§3) run in parallel → tenant confirms each → orchestrator marks active-with-pending → maintenance (§5) takes over.

**Steady state (maintenance-driven):**
change-layer events / periodic sweep → maintenance dispatches to layer agents for re-checks → drift candidates produced → tenant via confirm-not-initiate → updated config → drift signal also feeds the trust ledger.

---

## 2. What is clean

- **Sequence follows dependencies, not policy.** Stack-discovery first because the other two depend on its mechanical config; conventions and intent in parallel because they don't depend on each other. The dependency graph *is* the flow plan.
- **Distinct responsibilities, minimal overlap.** Each agent has a single domain; cross-cutting concerns (the change layer, confirm-not-initiate) are shared, not duplicated.
- **Confirm-not-initiate is consistent.** Every agent proposes; every gate is human-confirmed.
- **Reuse is real.** Maintenance dispatches to the layer agents for re-checks rather than reimplementing extraction/inspection/elicitation. The orchestrator is an instance of the existing orchestrator pattern, not new infrastructure.
- **Capture vs use is separated cleanly.** Intent-elicitation (§3) captures intent; the user-rep advocate (from the salience doc) uses it at runtime. Different agents, different lifecycles.
- **The three-tier evidence model recurs.** Deterministic / heuristic / judgement-only appears in conventions (§1.9) and in maintenance (§5.5) — same pattern, applied to checking and to drift respectively.
- **Cross-references to the broader machine are stated.** Intent → advocate (§3.8); maintenance → trust ledger (§5.7); orchestrator → same code for our use and tenants (§4.7). The seams exist in concept.

---

## 3. Gaps and contracts to settle

### 3.1 The active-config schema (load-bearing)

**This is the most important gap.** Each layer agent has its own output shape — a mechanical-config document (§2.2), convention atoms (§1.2, with the doc-tree FOCUS atom shape), and objective-tree atoms with priority profiles and direction-of-travel (§3.2, with the salience-doc shape). The **"active config"** that the orchestrator outputs (§4.2) and the maintenance agent consumes (§5.2) is the **union** of these three — but its shape is implied, not stated. Every downstream consumer — the bundle builder, the user-rep advocate at runtime, the mediator, the decision-point checkers — reads this artifact, and they all need a stable shape to read.

**Needs:** a versioned schema for the active config holding `{ mechanical_config, convention_atoms[], objective_tree[], provenance_metadata }`, with field-level provenance (source / confidence / last_verified / verified_by) consistent across layers.

### 3.2 Storage decisions

Where each artifact lives:

| Artifact | Likely store |
|---|---|
| Mechanical config | Postgres table (or one row per tenant in a `configs` table) |
| Convention atoms (concern tree, flat) | `standards` table |
| Objective-tree atoms (why-chain, nested) | `objectives` table |
| Drift candidates / audit history | Work-items table (§3.3) or a `drift_findings` table |
| Trust ledger | TBD (§3.5) |
| Published-reasoning log | TBD (§3.5) |

Most map naturally to Postgres tables consistent with the existing chassis. The table shapes follow from §3.1.

### 3.3 Pending-confirmation queue / post-onboarding interface

During onboarding, the orchestrator routes proposals to the tenant. After onboarding, maintenance surfaces drift candidates "via confirm-not-initiate" — but the routing surface isn't specced. Options:

- (a) Each agent surfaces directly to a tenant UI.
- (b) **All agents emit work-items into a shared queue; a single config interface reads the queue.**

(b) is cleaner and reuses the existing work-items pattern: every agent emits work-items in a standard shape (kind, source agent, payload, status); the UI/CLI consumes them; tenant confirmations close them; agents read the resolutions. This also means the tenant's interaction surface stays singular — they don't deal with one interface for onboarding and another for ongoing maintenance.

### 3.4 Maintenance ↔ orchestrator handoff and re-onboarding

Open in §4.9 and §5.10. Needs a defined **active-with-pending** state that triggers maintenance to take over. Lean:

- Orchestrator marks the config as active when the mechanical layer is confirmed, the convention set is confirmed, and at least top-of-tree intent is captured.
- Maintenance starts watching from that point.
- **Re-onboarding** (a tenant restructures, or adopts different doc standards) re-enters the orchestrator only on explicit tenant request or detection of major structural change; otherwise everything is incremental via maintenance.

### 3.5 Trust ledger and published-reasoning log

Both are referenced by maintenance (§5.4, §5.5, §5.7) but not specced anywhere — they belong to the broader machine (`MASTER_autonomous_build_and_operate`), not strictly to onboarding. Two routes:

- (a) Spec them now as adjacent infrastructure.
- (b) **Settle the contract shape now (what entries look like, what mutations are allowed); defer implementation to the broader build.**

Recommend (b). The contracts are what the maintenance agent depends on; the implementation can come later. Without the contracts, maintenance has open ends.

### 3.6 Change-layer integration

The maintenance agent uses diffs as triggers (§5.4) but the mechanism isn't named — git hooks, watchers, polling, push events. For our chassis: probably a git-event listener that reads commits and triggers the maintenance agent. For tenants: similar mechanism scoped per tenant, depending on how the tenant's repo is connected.

### 3.7 The two atom trees

Convention atoms and objective-tree atoms are both "atoms" but live in conceptually distinct trees — the **horizontal concern tree** from `FOCUS_best_practice_doc_tree` and the **vertical objective tree** from `FOCUS_salience_and_multi_author_mediation`. They share fields (id, rule/why, provenance) but have different relationships (concerns are flat with `applies_to` tags; objectives nest with parent links). Decision: two tables (`standards`, `objectives`) sharing common metadata fields. Worth being explicit in the schema so the two trees don't get confused at the storage layer.

---

## 4. Integration with the broader machine

The onboarding agents produce the active config that the rest of the system reads. The seams:

- **Active config → bundle builder.** The bundle builder (`PLAN_context_assembly_tool_and_service`) composes task-scoped bundles from the active config. Needs §3.1.
- **Active config → user-rep advocate, mediator, decision-point checkers.** These read the priority profile, why-chain, and convention atoms at runtime decisions. Same §3.1 dependency.
- **Maintenance drift signal → trust ledger.** The §5.7 feedback loop. Needs §3.5 contract.
- **Every gate-confirmation decision → published-reasoning log.** Foundational from the running notes: every decision publishes its premise. This includes the onboarding orchestrator's gates and the maintenance agent's drift confirmations. Needs §3.5 contract.

The agents and the broader machine are cleanly **referenced** in concept but the **shapes** at the seams are not yet defined.

---

## 5. Recommended order before implementation

Each step unblocks the next. In dependency order:

1. **Define the active-config schema (§3.1).** The load-bearing contract. Lock the field set; storage decisions follow.
2. **Define the work-items shape for pending confirmations (§3.3).** The shared queue all agents emit into.
3. **Settle the trust-ledger and published-reasoning-log contracts (§3.5).** Implementation can lag; the contracts cannot, because maintenance depends on them.
4. **Specify the change-layer integration (§3.6).** How diffs reach the maintenance agent.
5. **Then build.** Phase 0/1 of `PLAN_context_assembly_tool_and_service` against now-stable contracts.

Without §3.1 in particular, the agents can be implemented individually but they will quietly disagree on the shape of the artifact they all touch, and the consumers downstream (bundle builder, advocate, mediator) will be built against guesses.

---

## 6. One-line state

The five onboarding agents have distinct responsibilities, clean reuse, and a sensible flow, but they share four undefined contracts — the active-config schema, the pending-confirmation queue, the trust-ledger / reasoning-log contracts, and the change-layer integration. The active-config schema is the load-bearing one; the others fall into place once it is defined. Settling these is the prerequisite to coherent implementation.
