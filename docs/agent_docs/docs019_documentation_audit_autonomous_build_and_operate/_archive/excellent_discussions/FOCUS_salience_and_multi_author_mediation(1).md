# FOCUS — Salience, Multi-Author Generation, and the Mediator as Multi-Objective Optimiser

**Status:** exploratory, active. Continues the thread after `HANDOFF_2026-05-25`. Companions: `FOCUS_context_authored_derived_change` (authored/derived/change), `FOCUS_standards_curation_and_governance` (curators, coordinator, advocate, authority), `FOCUS_mediator_routing_model` (who's implicated), `MASTER_autonomous_build_and_operate` (the cascade, control loop).

---

## 1. Origin: salience, not presence

The thread began on documentation/code-analysis for prompting — a task-scoped **context bundle** (thin authored frame + tag-selected standards + in-scope code full-text wrapped in a signature-level neighbourhood map + fresh slices of derived state + pointers to everything else; code analysis = signatures/types, call graph, reuse search, schema extraction).

The reframe that reorganised everything: an LLM loses the bigger picture **not because the text left the window** (the bundle keeps it present) **but because local detail is more *salient*** mid-reasoning. Attention follows the concrete and immediate. So the lever is **salience at the moment of decision**, not presence somewhere up top.

---

## 2. The four axes a development step is governed by

| Axis | Kind | Source |
|---|---|---|
| **Purpose** — why this exists | vertical | the why-chain (objective tree) |
| **How-well** — standards | horizontal | the concern tree |
| **Where-heading** — trajectory | **dynamic** | direction-of-travel |
| **What-is** — code + live state | local | the bundle |

The **dynamic axis was the gap.** A snapshot of code + objectives says where things *are*, not where they're *heading* — the bigger-picture context a human holds and an LLM discards.

---

## 3. Representation and currency of the new layers

- **Why-chain** = a traversal of the existing objective tree (each node: a thin "why" + parent link; the chain is the root-to-node path). Stable, low-churn, **human-owned** (confirm-not-initiate). Reuses the tree; not a new structure.
- **Direction-of-travel** = a *vector*, not a reason: current heading, settled-don't-relitigate decisions, deliberately-temporary states, and what's still in flux. **Fast-churn**, so kept as a *separate attribute* from the why (different change rates). It is the **authored vector laid over the derived change-layer** — so the system can *propose* it from recent diffs (authored-by-inference) but only **human confirmation** makes it authored-by-record, because intent (settled decisions, deliberate-temporary) isn't recoverable from diffs. Confirmation is **low-ceremony** (a wrong trajectory note is far lower-stakes than a wrong rule; heavy gating would just make it stale by friction).
- **Both staleness modes apply to direction-of-travel** (drifts from reality *and* snapshot of a moving target) → keep it thin + pointer-rich *and* **freshness-stamped**. A note past its window is surfaced **flagged as possibly stale**, never silently trusted — a stale trajectory presented as current is worse than none (false confidence).
- Ownership reuses governance: area owners own their node's why and confirm its direction; the coordinator owns mission-level why and cross-area trajectory coherence.

---

## 4. Surfacing at the moments that matter

- **Salience over presence → turn the why-chain into a *question* at decision/gate points** ("does this serve [why-chain]; does it relitigate anything in [direction-of-travel]?"). Interrogation forces the altitude shift that mere inclusion does not. Strongest, cheapest anti-drift mechanism.
- **Step-type-aware composition** (prompt-composition made altitude-aware): framing/routing → full why-chain + direction; generation → collapse to a **one-line tether** (depth needs local salience); conformance → local, why-chain irrelevant; **fitness** check (does the verified artifact serve the objective — distinct from "does it build") → full why-chain; gate → full why-chain + direction as first-class frame.
- **Within a long step**, re-assert the one-line tether at transition points (sub-task done, about to commit) — light, or it becomes noise.
- Selective ≠ compromise: **depth is a virtue, not only a failure mode.** Blanketing the mission over every step would degrade the implementation steps where depth is what you want. Goal = **right altitude at the right moment.**

---

## 5. Parallelism for salience (the checker model)

- A single attention budget can't hold detail and breadth at once → **run several budgets, each narrow, each fully salient on one axis.** This is what curators and the advocate **already are** — each holds one concern at full salience *because it isn't doing the implementation.*
- The builder is **allowed to tunnel**; that loss is *covered* by processes that can't lose their axis. Tunnel vision becomes division of labour.
- **Key distinction:** *parallel holding* (continuous, expensive, mostly wasted — you don't need the mission burning a process while a loop body is laid out) vs *check at the decision* (single-axis checkers fired at decision points, concurrently, returning terse verdicts, then reconciled). Want the latter.
- **Decision (user):** lean to **frequent cheap checks, accept some latency (not too bad)**; cheap single-purpose evaluations by default, escalate to a full agent only on a flag (cascade).
- **Stresses:** detecting decision points — structural ones (commit, gate, approach-choice) are easy; **implicit mid-stream decisions** are the hard case (builder raising its hand asks the tunnel-visioned process to notice it's at a junction; fallback = fixed cadence, every N steps / sub-task boundary, accepting waste). More salient voices = more conflict → **parallelism produces verdicts, not decisions**; arbitration stays singular (mediator/authority model).

---

## 6. The multi-author alternative (every concern authors a full solution)

A *what-if* that **inverts where correctness comes from**: each perspective is an **author**, not a guardrail. Purpose writes the purpose-optimal solution, security the security-optimal one, etc. → N solutions, each **maximally salient on its axis by construction** (the only thing in the room when it was written).

- **Salience dissolves structurally** rather than being managed — a stronger guarantee than any amount of checking, because it removes the competition for attention.
- **Disagreements become demonstrations, not complaints** — a worked alternative ("here's what prioritising purpose looks like in code") is richer than an objection. Better substrate for getting it right.
- Structurally = **generative competition along the *concern* axis** (not the attempt axis): standard best-of-N competes N attempts at one objective; this competes N attempts at *different* objectives, and selection becomes reconciliation across concerns. Reuses cascade tier-3, mediator, advocate, verification.

**Costs / tensions:**
- **Reconciliation changes kind.** N whole solutions may differ *structurally* (different decompositions/data flows), often **unmergeable** ("take purpose's control flow and schema's tables" may be incoherent). Risk: the mediator now holds N solutions + all concerns to judge — **the salience overload relocated upward.**
- **Mitigations:** most concerns **abstain** on most tasks → routing bounds the fan-out to the *implicated* set (~2–4), not the whole concern set; and **counter-proposals on deltas**, not whole rewrites (keeps the demonstration value, bounds reconciliation to where solutions diverge).
- **Deep tension:** "getting it right" assumes a single right; for value-laden tradeoffs **there isn't one.** Multi-author is an excellent **option-generation / tradeoff-surfacing** engine and a **poor decision engine** — it makes the tradeoff vivid, it cannot resolve it.
- **Relationship to §5:** multi-author sits **on top of** the checker model, **selectively** — cheap checks as the frequent floor, multi-author as the deep ceiling for high-stakes or checker-flagged conflict. Cascade again.

---

## 7. The mediator as multi-objective optimiser

- **"Right" = a requirement-relative, defensible *balance*** among conflicting dimensions (fast / secure / generic / simple / functional) — **not pick, not merge.** The authored solutions are the **extremes that bound the space**; the mediator finds the point inside it the requirement wants.
- **Priority is not global.** A fixed ranking is wrong exactly where judgement is needed. The ranking is **requirement-relative and comes from the why-chain's priority profile** (an internal tool and a public auth endpoint have different right answers). **Direction-of-travel can override temporarily** ("simple for now, hardening later").
- **"All areas represented"** refined: every implicated concern is **invited and none silently omitted**, but a concern may **abstain** ("no strong position"), and **abstention is information** — it marks which dimensions are in genuine tension vs slack.
- **"All info even when heuristics decided":** two layers — a **heuristic floor** (fast, settles the *uncontested* majority via standards / priority profile / known-good) and **full mediation** (rare, contested). The crucial rule: a heuristic must emit a **decision + provenance (a legible trace)**, not a silent fait accompli. **Information availability, not information presence** — the full picture is recoverable on demand when judgement is invoked, without being forced into context when it isn't (else the salience overload returns).
- **Provenance:** **knowing-but-disciplined.** The mediator knows which concern authored which solution (that's the signal that makes tension legible), but provenance informs **weighting, not deference** — authority comes from the requirement-relative priority profile applied to a known provenance, **not from the loudest/highest-status concern.**
- **Tension:** a codified heuristic encodes **yesterday's balance.** If requirements shift (internal → public), the heuristic settles the tradeoff the old way, fast and invisibly — exactly when it should be reopened. So legibility is also the **invalidation hook**: a heuristic carries the priority assumptions it baked in, and is **auto-flagged for re-mediation** when the current why-chain/direction no longer matches them. The cheap floor stays cheap but knows when it's no longer entitled to decide alone.

---

## 8. N-round convergence

The single-round multi-author approach (§6) requires the mediator to combine N complete solutions in one step, which is the reconciliation cost that can overload it. The N-round approach replaces that one hard step with several easier ones.

- **Mechanism.** Round one produces solutions. The process then produces a current candidate, and each still-active concern reacts to that specific candidate: either it is satisfied, or it states what it still needs. The candidate changes across rounds. Because a satisfied concern stops participating, the set of active concerns shrinks as the process continues, so the work per round decreases rather than staying at N.
- **Drop-out is a mode change, not removal.** A concern that is satisfied stops proposing changes (author mode) but continues to cheaply verify that each new candidate still satisfies it (checker mode), and returns to author mode if a later change breaks it. This makes the checker model (§5) and the multi-author model (§6) two modes of the same concern within one process: concerns begin as authors, become checkers as they are accommodated, and become authors again if violated. The expensive author mode is occupied only by unsatisfied concerns and shrinks over time; the cheap checker mode is the steady state.
- **Convergence separates resolvable tensions from the irreducible one.** Most apparent conflicts are unexplored space and dissolve once concerns see a candidate. The rounds remove these. If the process does not converge — accommodating one concern repeatedly breaks another — that residual is the genuine value-laden tradeoff that no synthesis resolves, now isolated with everything else settled around it. So **non-convergence is the escalation signal**, and what it escalates is precise: one identified tradeoff needing a judgement, not a vague disagreement.
- **Termination.** Bound the rounds with a cap (the audit-pass-cap pattern). Converged before the cap: all concerns are in checker mode, done. Cap reached while still oscillating: the residual is the irreducible tradeoff, escalated to the mediator (using the priority profile, §9) or, if the profile does not decide it, to a human via the co-equal-voices decision package. Reaching the cap is itself diagnostic, not only a safety stop.
- **Who judges that a concern is satisfied.** Both the concern and the mediator. A concern may withdraw itself (it knows its own satisfaction). The mediator may also dismiss a concern it judges sufficiently accommodated (this prevents a concern holding the process open indefinitely). A concern dismissed by the mediator while still unsatisfied has its residual objection recorded, not erased, so a wrong dismissal is visible at the gate. Concerns advise, the mediator arbitrates, residual disagreement escalates — the authority model, applied per round.
- **Cost.** Rounds times concerns is many calls, so N-round is the deepest level of the cascade, reached only when single-round multi-author has not settled a high-stakes contested decision. The round count is adaptive: most decisions converge in one or two rounds; only genuinely hard ones run deep, bounded by the cap.

---

## 9. Priority-profile representation

The priority among dimensions (fast / secure / generic / simple / functional) is requirement-relative and lives on the objective-tree node, read off the why-chain.

- **Weights and constraints, not a strict total order.** A strict ranking is wrong where judgement is needed. Each node carries a small structured profile that distinguishes **hard constraints** (a floor that gates — "must be secure") from **weighted objectives** (traded against each other — faster versus simpler).
- **Inheritance with override.** A node inherits its parent's profile and overrides specific dimensions (an area raises security from weighted to constraint for a public-data leaf). Effective profile = the merge down the chain, child winning on conflicts. This is the same override pattern as prompt composition, reused.
- **"Sometimes a security measure is too strict" has a precise meaning:** an explicit, recorded judgement to relax a constraint to a weighted objective for this node — a profile override with provenance, not an ad-hoc exception.
- **Direction-of-travel modulates the profile temporarily.** "Deprioritise hardening for now" is a time-bounded, freshness-stamped reweight over the static profile. Effective priority = the static profile (from the why-chain) modulated by the trajectory override (time-bounded, expiring).

---

## 10. Mediator reasoning and the trace

- **Reasoning.** The mediator identifies where the authored solutions diverge (tensions) and where they agree (slack). For each tension it applies the effective profile: a constraint bounds the outcome; a weighted set is traded according to the weights. Where the profile does not decide, that is the residual that escalates (§8 termination).
- **The trace** records, per tension: that it existed; how it was resolved (which rule, weight, or known-good solution, or a human decision); the priority assumptions the resolution depended on (the invalidation hook, §7); and any objections dismissed but recorded (§8). This one artifact is simultaneously the gate decision package, the provenance a later re-opening reads, and — in the N-round case — the accumulated record of changes and drop-outs.

---

## 11. Published reasoning as substrate, and drift detection

This is foundational, not an addition: it is what the trace, the N-round process, and the drift detector all stand on.

- **Every decision publishes its reasoning, not just its outcome.** "Chose X" teaches nothing later; "chose X because for this node security was a constraint not a weighted objective, and the candidate met it at acceptable speed cost" is reusable. When a human decides, the human explains the premise so later reasoning can learn the basis the decision rested on, not only the decision. This upgrades the trace from a record into reusable reasoning and training signal: later reasoning can ask what premise a decision rested on and whether it still holds.
- **Drift detection.** After several document updates, a detector flags where a standard's or objective's stated premise no longer matches the premises that recent decisions were actually made on. Drift is the gap between a decision's logged premise and the current premise. This is possible only because reasoning is published — without logged premises there is nothing to compare. The detector raises candidates for human review (it does not edit standards), through the confirm-not-initiate path.

---

## 12. Candidate ownership in N-round (resolving the §8 probe)

The question was: who produces the current candidate each round? If the mediator synthesises it, the mediator again holds everything, now per round. If the candidate is simply the previous round's best solution carried forward, the mediator stays light but the result is overly influenced by whichever concern authored the first solution.

**Resolution: the candidate is owned by no concern.** It is a shared artifact plus a log of changes. Each round is a set of proposed changes against the current candidate, not a re-synthesis of it.

- **The base candidate is seeded deliberately** — from the known-good solution, or from the highest-priority dimension's solution per the profile — not from an arbitrary first author. The profile already says which dimension dominates for this node, so the starting point is chosen for a stated reason rather than by accident.
- **Each later round, active concerns propose targeted changes** with their reasoning, not rewrites. The mediator's per-round job is to adjudicate a small set of proposed changes against the current candidate (accept, reject, or mark as conflicting), each adjudication published with its reasoning. The mediator holds the current candidate plus this round's proposed changes — bounded, and shrinking as concerns drop out.
- **Drop-out and re-promotion fall out directly:** a concern drops to checker mode when it proposes no changes in a round; it returns to author mode when an accepted change from another concern breaks its check, which is a cheap local check because the changes are small.
- **The change log is the trace** (§10): the base, every proposed change with its premise, every accept/reject with the mediator's reasoning. It serves as decision package, published reasoning, and drift-detector input at once.

**The honest limitation: this is path-dependent.** The order in which changes are proposed and accepted affects where the candidate ends up, and it converges to a balance near the seed. It is good at refining toward a balance close to the starting point and poor at discovering a structurally different solution that would need a different decomposition rather than a series of changes. Seeding from the profile removes the *arbitrariness* but introduces a *bias toward the seed's dimension* — the base solution's structural choices persist unless explicitly changed away.

**Mitigation — rival-base proposals.** A concern that cannot reach an acceptable result by proposing changes may instead propose a different starting structure. This resets the candidate, so it is expensive and is gated: allowed only in early rounds, or when a concern's check keeps failing across rounds (a signal the seed was wrong). Rare and bounded, but it keeps structurally different solutions reachable rather than permanently excluded by the seed choice.

**Open question:** how available rival-base proposals should be — early rounds only, or available throughout but costly. Too available and the process may not converge and pays for repeated resets; too restricted and the seed's bias becomes near-permanent and structurally different correct solutions are lost.

---

## 13. One-line state

Salience is lost to local detail, not to absent text, so it is addressed by single-axis processes (checkers) convening at decision points, with multi-author generation as the deeper level that removes the competition for attention by making each concern an author. N-round convergence replaces one hard reconciliation with several easier ones; concerns move between author and checker modes and the active set shrinks; non-convergence isolates the one genuine tradeoff to escalate. The mediator finds a requirement-relative balance using a priority profile of constraints and weighted objectives from the why-chain, modulated by direction-of-travel. Every decision publishes its premise, which makes the trace reusable and makes drift detection possible. The round candidate is owned by no concern — a shared artifact changed by premised proposals, seeded deliberately from the profile, with rival-base proposals as a gated escape from seed bias. Open: how available rival-base proposals should be.
