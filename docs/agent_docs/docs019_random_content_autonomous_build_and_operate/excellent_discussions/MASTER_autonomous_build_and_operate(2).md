# MASTER — Autonomous Build & Operate: Folding the Threads Together

**Status:** synthesis spine, built over several turns. Sections 6–8 are deliberately stubs to detail next.
**Example target:** vonc.com (a social platform — the Forge / spark concepts are the *what*; this plan is the *how*).
**Base:** the existing framework is kept and extended, not replaced.

**Composes these companion FOCUS docs** (this is the umbrella over them):
`FOCUS_self_development_coding_pipeline_reasoning` (coordination A/B/C, mediator), `FOCUS_best_practice_doc_tree` (atomic standards), `FOCUS_mediator_routing_model`, `FOCUS_doc_tree_adoption`, `FOCUS_standards_curation_and_governance` (curators, coordinator, advocate, authority), `FOCUS_context_authored_derived_change` (authored / derived / change-layer).

---

## 1. Thesis: the bottleneck is trust, not capability

The technical pieces are proven — actions/workflows authored via chat; nginx, GPUs, Prometheus, Grafana provisioned from code; remote hosting stood up. The blocker is **LLM-response uncertainty**, which makes removing the human impractical, not impossible. So the whole problem is a **reliability problem**: bound the uncertainty at each step enough to progressively remove the human. Everything discussed across the prior docs is already apparatus for this; the synthesis names them as one toolkit and composes them across the full lifecycle.

---

## 2. Automation is a ratchet, not a switch

- Each **capability** (create an action, provision an nginx, reshard a DB) has its own trust level: `confirm-every → confirm-exceptions → notify → autonomous`.
- A capability **graduates** when evidence shows it reliable at its current level — the evidence-driven, confirm-not-initiate logic from the curation doc, applied to capabilities instead of standards.
- "Fully automated" = the **union of individually-graduated capabilities**. Nothing is globally automated; nothing moves until it earns the move.
- The **trust ledger** (§5, new) records each capability's level, gate policy, and the evidence behind it.

---

## 3. The reliability cascade (how any unit of work is produced)

Applied to every sub-task, in descending order of reliability:

1. **Known-good reuse** — adapt a proven solution from the library (§5). Highest reliability; the default. Makes reuse-before-recreate load-bearing for reliability, not just tidiness.
2. **Generate + verify** — no proven solution, but deterministic verification exists → generate, verify against ground truth.
3. **Compete + judge** — verification weaker → generate N candidates in a controlled sandbox, select by the strongest available evaluator; gate by HITL-confirm if stakes warrant ("generatively competing in a small controlled way").
4. **HITL** — verification weak and stakes high, or inherently human work → human does/decides.

**Feedback to the ratchet:** a generated solution that verifies and recurs becomes a candidate known-good; reliable repetition graduates its gate (§2). This is a per-task reliability router — sibling to the standards routing model.

---

## 4. Build vs operate asymmetry

- **Build** (actions, workflows, components/pages, agent definitions): isolatable — branches, sandboxes, "broken output never overwrites," per-site independence. **Competition is safe.** Ratchet moves faster.
- **Operate** (provisioning, scaling, sharding, incident response): live, stateful, hard to A/B. **Competition is risky.** Leans on known-good + canary + rollback + tighter HITL. Ratchet moves slower.
- So the cascade's mix shifts by domain: build tilts to generate/compete; ops tilts to reuse/known-good with conservative gates.

---

## 5. Documentation model for automation (the connective tissue)

Four layers — two existing, two new — plus governance across all of them.

- **(existing) Standards tree** — prescriptive: how things should be done. (`FOCUS_best_practice_doc_tree`)
- **(existing) Context substrate** — authored / derived / change-layer; the live grounding assembled per step. (`FOCUS_context_authored_derived_change`)
- **(NEW) Known-good solution library** — proven solutions captured as reusable, parameterised templates. The artifacts→docs arrow made concrete: derived success promoted to authored-reusable. Indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. This is the substrate the cascade's tier-1 reuse draws on.
- **(NEW) Trust ledger** — per-capability automation level, gate policy, and supporting evidence. The governance memory that drives the ratchet (§2).
- **Governance** (curators / coordinator / advocate, confirm-not-initiate) sits across all four as the confirmation and arbitration layer. (`FOCUS_standards_curation_and_governance`)

---

## 6. Lifecycle map — by verifiability and containment

Organised purely by capability and verification strength (existence of adapters/actions/workflows is set aside — they have strong foundation docs and build quickly; the orchestrator is well-tested and should resist change — see §6.3).

### 6.1 Two factors set the ceiling

How far and fast a capability can climb the ratchet (its **ceiling**) is set by two factors that fail differently — a capability needs both, and is held back by whichever is weaker:

- **Verifiability** — can we tell, against ground truth, that the output is correct? *Strong* = compiles / tests / validates / health-checks. *Weak* = subjective ("is this page good") or adversarial ("is this secure" — can't prove a negative).
- **Containment** — if it is wrong, how bad and how reversible? *Strong* = sandbox / branch / fast rollback. *Weak* = live data loss, wide blast radius, slow rollback (DNS propagation), or compounding effects.

Ceiling is separate from **current maturity** (where a capability sits today). The two call for different investment — see §6.4.

### 6.2 The map

Grouped by ceiling. "Cascade tilt" = which tier of §3 dominates.

**Tier A — high ceiling (strong verifiability + strong containment): autonomy reachable, competition safe**

| Capability | Verifiability | Containment | Cascade tilt |
|---|---|---|---|
| Go actions | Strong (build/vet/test) | Strong (sandbox/branch) | generate+verify; cheap compete |
| SQL workflows | Med–Strong (parse/dry-run; semantic order softer) | Strong (sandbox) | generate+verify |
| Component structural validity | Strong (valid markup, scoped CSS, schema/template sync, renders) | Strong (per-site isolation, regenerate) | generate+verify |
| Observability config (Prometheus/Grafana/logging) | Strong (metric flows, dashboard renders, test alert fires) | Strong — failure loses *visibility*, not the system | reuse+verify |
| Rollback | Strong (returns to known-good state) | Strong — rollback *is* the containment | known-good |

Note on rollback: it can be trusted to run autonomously *earlier than roll-forward*, precisely because returning to a known-good state is inherently the safe direction.

**Tier B — medium ceiling (one factor weaker): supervised-autonomous**

| Capability | Verifiability | Containment | Cascade tilt |
|---|---|---|---|
| Agent definitions | Medium (behavioural verification needs running it) | Med–Strong | generate+verify (behavioural) |
| Greenfield provisioning (hosting/nginx/GPU) | Strong (`nginx -t`, health checks) | Medium (mutating live is lower) | reuse+verify+canary |
| DNS/TLS | Strong (resolves, cert validates) | Med–Weak (slow rollback via TTL/propagation) | reuse+canary |
| Capacity / autoscaling | Medium | Medium (cost + outage, reversible) | known-good + guardrails |
| Incident detection | Medium (false +/- rate measurable over time) | Strong (observation, no action) | reuse+verify |
| Diagnosis (as *suggestion*, no action) | Weak (confirmed only by the fix working) | Strong (reasoning, no action) | generate |
| Component quality (design/content) | Weak (subjective) | Strong (regenerate) | compete + outcome signal (engagement) or HITL |

**Tier C — low ceiling (weak verifiability and/or catastrophic blast radius): stays gated regardless of agent capability**

| Capability | Verifiability | Containment | Cascade tilt |
|---|---|---|---|
| Security (secrets/access/hardening) | Weak (adversarial; can't prove a negative) | Weak (catastrophic, reputational/legal) | known-good baseline + external scan + HITL |
| Sharding / reshard | Medium (mechanics) | Weak (live data, hard to reverse) | known-good runbook + HITL |
| Replication / consistency | Medium | Weak (data consistency) | known-good + HITL |
| Storage with persistent data | Strong (mechanics) | Weak (data loss) | known-good + HITL |
| Live remediation (roll-*forward* under pressure) | Medium (fix observable) | Variable–Weak (can worsen live state) | known-good runbook + HITL |
| Meta-loop (system changing its own dev/ops) | Weak (slow, confounded signal) | Weak (compounding across all future work) | HITL-heavy |

### 6.3 Reality check — implementing from the base

- **The orchestrator stays and is wrapped, not modified.** The new machinery is a pre-dispatch decision layer plus data stores, expressed as agents/actions the orchestrator already knows how to dispatch (reusing spawn, work-items, adapters, "every agent is an orchestrator"). This honours "the orchestrator should resist change."
- **Genuinely new, additive pieces** (at the edges of the existing base):
  1. **Verification harness** — build-check / test-runner / validator / canary / rollback as actions/adapters. Build side is easy (strong foundation docs); the **ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work**.
  2. **Known-good solution library** — a templates store + retrieval (reuse pgvector + a parameterised-solutions table). Feeds cascade tier-1.
  3. **Trust ledger + gate-policy engine** — a store plus a small evaluator mapping (capability, trust level, stakes) → gate. Drives where HITL fires.
  4. **Cascade router** — the per-task decision (reuse / generate / compete / HITL), an action or agent.
  5. **Sandbox / canary substrate** — build sandbox is easy (branch/isolated); ops canary is the hard, thin part.
- **Thinnest part of the base = the operate/ops side**, matching the read that the autonomous part is weakest. Most new building concentrates in the ops verification harness.

### 6.4 Ceiling vs maturity (where to invest)

- **High ceiling, low current maturity → invest; it converts directly to ratchet progress.** This is where "backend weak" sits — weak in *maturity*, strong in *ceiling* (actions/workflows are verifiable). Effort here pays off fast.
- **Low ceiling regardless of maturity → don't chase full autonomy; invest in known-good runbooks and good HITL ergonomics instead.** Security, sharding, live remediation. They stay gated; the goal is to make the gated path fast and safe, not to remove the gate.
- **Frontend:** high ceiling structurally ("nearly there"), with the quality half capped by subjectivity until an outcome signal (engagement) exists to judge it.

### 6.5 Sequencing (the critical path to "build the site from scratch")

1. **Build the ratchet machinery on Tier A first** (router + library + ledger + verification harness, on actions, workflows, component-structural, observability, rollback). Strongest verifiability + containment = the safest place to prove the whole reuse→generate→verify→compete→graduate→gate-down loop. Do not build autonomy on an unproven loop.
2. **In parallel, raise backend maturity** — high ceiling, currently weak; the best return on build effort.
3. **Then the autonomous control loop (§7)** — cascade-routing + trust-ledger gating that lets Tier A/B run with reducing supervision. This is the "full autonomous part": the weakest and largest investment, sequenced *after* the machinery is proven because it stands on it.
4. **Operate / scale / secure (Tier C) last and partial** — automate only the safe sub-parts (rollback, detection, diagnosis-as-suggestion); keep the dangerous parts (reshard, security, live remediation) on known-good runbooks + HITL, possibly indefinitely.

The critical path to building vonc.com from scratch runs 1 → 2 → 3. The weakest link (the autonomous control loop) sits *on* that path, so it gets the most investment — but it cannot be shortcut, because it depends on the machinery being proven on Tier A first. Operating at scale (4) follows building.

---

## 7. Control loop

How a goal becomes verified, trust-gated, applied work — and how the result feeds back. The orchestrator's decompose-and-dispatch is reused unchanged; the new machinery wraps each leaf task and the feedback paths.

### 7.1 Reused vs new

- **Reused unchanged:** orchestrator decompose-and-dispatch; spawn; work-items; adapters; the validate→regenerate loop; the mediator (Position C, spawn-fresh); the change layer as derived state.
- **New, wrapping the edges:** cascade router, verification harness, trust ledger + gate engine, known-good library, and the feedback paths back into the ledger and library.

### 7.2 The loop at a single leaf task (the unit)

A leaf task is one unit of work ("write the vote-tally action"). Each leaf runs:

1. **Route (how to produce it).** The cascade router (§3) picks a tier using three inputs: the capability's verifiability/containment (§6 map), its **trust-ledger** entry, and this task's stakes. Order: known-good reuse → generate+verify → compete+judge → HITL.
2. **Produce.** The chosen path runs via existing spawn machinery (agents spawned fresh, reply to caller's topic).
3. **Verify.** The verification harness checks against ground truth — build/vet/test/validate for build; canary/health-check for ops. Failure loops back to step 2 with the failure as feedback (existing validate→regenerate), bounded by the audit-pass cap.
4. **Gate (whether to apply without a human).** The trust ledger's level for this capability sets the gate: `autonomous` → apply; `notify` → apply + inform; `confirm-exceptions` → apply unless an anomaly, else gate; `confirm-every` → HITL, via a governance decision package (curators advise, advocate represents intent, coordinator arbitrates and frames, co-equal voices, abstention fallback, creator veto).
5. **Apply → derived state.** The applied result becomes new derived state, observable as a diff in the change layer.
6. **Feed back** (§7.5): the outcome updates the trust ledger's evidence, and a verified recurring solution becomes a candidate known-good.

**The trust ledger is the master knob.** It governs both *how conservatively* a thing is produced (it can floor the cascade at a safer tier for a low-trust capability) and *whether* the result applies without a human (the gate). The ratchet turns this one knob.

### 7.3 Decomposition above the leaf (reuses the orchestrator + objective tree)

"Build and run vonc.com" is decomposed by the existing orchestrator down the objective tree — site → front-end / back-end / data model / provisioning / observability / security → … → leaf tasks. Fractal, every agent an orchestrator. A **single-area** leaf runs §7.2 directly (the cheap path). A **cross-area** leaf engages the **mediator** (Position C): the routing model selects the applicable standards, the relevant **curators advise**, the **advocate** checks intent, the **coordinator arbitrates and frames**, and the mediator serialises the change on a branch through the gate.

### 7.4 Worked example — the build path

**Single-area leaf: "write the vote-tally action."**
Router queries the known-good library → if a proven counter/aggregation action exists, adapt it (tier 1). If not → generate+verify (tier 2: Go action, strong verifiability — build, vet, unit tests against fixtures). Verify passes. Gate: if "write Go action" has graduated to `notify`, apply to branch and log; if still `confirm-every`, emit a decision package for HITL. Result commits to the branch (derived state). Feedback: success recorded as evidence toward graduating the "write Go action" capability; if this vote-tally pattern recurs, it is captured as a candidate known-good.

**Cross-area leaf: "add a `reputation` field the vote-tally action writes and the thread-view component reads."**
Touches data-schema + back-end + front-end → the mediator engages. Routing selects the schema-migration standards and the component contract; the data-schema and front-end curators advise; the advocate confirms this serves the user's intent. The "schema migration" capability has weak containment (Tier C) so its gate is `confirm-every`: the coordinator frames a co-equal decision package and a human confirms. On confirmation, the mediator serialises the migration, the action change, and the component change on one branch through the gate. Each sub-change still ran its own §7.2 loop; the mediator coordinated their ordering and the single human confirmation.

### 7.5 Feedback and the ratchet (bidirectional)

- **Success** → evidence accrues toward graduating the capability's trust level; a verified, recurring, conformant solution is promoted to the known-good library (the artifacts→docs arrow, gated as in the curation model — multi-instance, human-confirmed, never a single success, to avoid codifying a fluke).
- **Failure** → iterate within the cap; on repeated or severe failure the ratchet moves **down** — trust level drops, the gate tightens, the cascade floor rises. **Trust can be lost, not only gained.** This bidirectionality is the safety property that lets the ratchet advance at all: a capability that regresses is automatically pulled back under supervision rather than continuing to run unwatched.

### 7.6 Ops re-enters the same loop (lighter)

The build loop terminates when a goal is built; the operate loop is continuous, triggered by **derived state** — logs, metrics, alerts — rather than by a goal. It is the *same* loop with two shifts: the trigger is an event/alert, and the cascade tilts to known-good + canary (§4). Within ops, the safe sub-capabilities climb independently: **rollback** can run `autonomous` early (it returns to known-good); **detection** is observation; **diagnosis** runs as a *suggestion* (no action); **live remediation** and **reshard/security** stay `confirm-every` on known-good runbooks. So "run vonc.com" is the control loop driven by derived state instead of by a build goal — the unification from §6: build and operate ride one loop, differing in trigger and cascade tilt. (Detailed ops capabilities deferred; build path is the critical path.)

### 7.7 Where uncertainty is bounded (the thesis, realised)

Each layer bounds a different uncertainty, which is §1 made concrete:
- **Cascade router** bounds *method risk* (don't generate what you can reuse; don't compete where you can't judge).
- **Verification harness** bounds *correctness risk* (ground truth, not the LLM's opinion).
- **Trust gate** bounds *application risk* (nothing applies above its earned trust level).
- **Governance** bounds *decision risk* (framed, contested, confirmed — not rubber-stamped).
- **Mediator** bounds *coordination risk* (cross-area changes serialised through one owner).
The loop is uncertainty-bounding at every step; full autonomy is what remains once every step's bound is tight enough, capability by capability.

---

## 8. Adoption path — TO DETAIL

From today (HITL via chat) to progressively autonomous, reusing the base. Per-capability ratchet rollout, ordered by where reliability is easiest to earn first. To be built out.

---

## 9. One-line state

The bottleneck is trust, not capability. Automation is a per-capability ratchet driven by a reliability cascade (reuse → generate+verify → compete+judge → HITL), supported by four documentation layers (standards, context substrate, known-good library [new], trust ledger [new]) and the existing governance. Build tilts to competition, ops to known-good. Lifecycle map, control loop, and adoption path to be detailed over coming turns.
