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

## 8. Open points (in discussion now)

- **N-round generative multi-author:** multiple comparison rounds; solutions evolve; a solution **drops out** when the concern (or the mediator) judges its concern **sufficiently adopted**. Emphasis on getting it right over avoiding trouble.
- **Priority-profile representation** on the objective-tree node (weights vs ranks, hard constraints vs soft objectives, inheritance, direction-of-travel override).
- **Mediator reasoning + the trace** it records (which doubles as the gate decision package and the heuristic-invalidation provenance).

---

## 9. One-line state

Salience is lost to local detail, not to absent text → fix by *salient single-axis processes* convening at decision points (checkers), with *multi-author generation* as the deep ceiling that dissolves salience structurally by making each concern an author. The mediator is a multi-objective optimiser finding a requirement-relative balance, with priority from the why-chain, heuristics settling the uncontested majority *legibly*, and full mediation reserved for genuine tradeoffs. Open: N-round convergence with drop-out, priority-profile representation, and the mediator's reasoning trace.
