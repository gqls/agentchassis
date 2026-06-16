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

## 6. Lifecycle map — TO DETAIL

The full set of capabilities to automate, across build and operate. For each: current state, verification strength, applicable cascade tier, target trust level. Domains to enumerate:
- **Build:** actions (Go), workflows (SQL), site components/pages, agent definitions.
- **Provision:** hosting, nginx, GPUs, DNS/TLS, storage.
- **Observe:** Prometheus, Grafana, logging, alerting.
- **Secure:** secrets, access, hardening, audit.
- **Scale:** sharding, replication, capacity, cost.
- **Operate:** incident detection, diagnosis, remediation, rollback.
- **Meta:** the development loop itself (the system improving its own build/ops).

---

## 7. Control loop — TO DETAIL

How "build and run vonc.com" decomposes through the existing framework: orchestrators → mediator → curators/advocates → cascade-routed execution → verification → trust-ledger update → ops feedback. To be built out.

---

## 8. Adoption path — TO DETAIL

From today (HITL via chat) to progressively autonomous, reusing the base. Per-capability ratchet rollout, ordered by where reliability is easiest to earn first. To be built out.

---

## 9. One-line state

The bottleneck is trust, not capability. Automation is a per-capability ratchet driven by a reliability cascade (reuse → generate+verify → compete+judge → HITL), supported by four documentation layers (standards, context substrate, known-good library [new], trust ledger [new]) and the existing governance. Build tilts to competition, ops to known-good. Lifecycle map, control loop, and adoption path to be detailed over coming turns.
