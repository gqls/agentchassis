# Mapping — the context tool as workflows, agents and checkers

What the thin-slice CLIs graduate into, and where everything we've discussed fits. Kept short, to build against.

---

## Two layers (so the bigger picture isn't lost)

- **The website-builder (the chassis).** The original problem: build a sophisticated multi-page site from a domain, using many specialist agents — adoption, planner, content-writer, research, imagery, rerender, deploy. This is the *subject* — the system the tool helps develop, and a product in its own right. Not replaced by anything here.
- **The context tool (what we've been building).** A development layer: given a task against a codebase, assemble the right context (intent + code + data + runtime + standards), help produce the change, and verify it before it ships. It is itself a workflow of spawned agents and checkers, and is generic/multi-tenant so it can serve other codebases later.

One line: the builder builds sites; the context tool builds reliable changes to the builder (and to any codebase). Different layers, both multi-agent.

---

## The context tool as a workflow

Each row is a distinct responsibility → a separate spawned agent/container (so logs and responsibilities stay distinct, replying on the caller's topic, as the platform already runs). The thin-slice CLIs are the first prototypes of three of these.

| Responsibility | Thin-slice today | Graduates into | Optional HITL stop |
|---|---|---|---|
| Index the codebase | `analyser.go` | **indexer action** + a refresh trigger on change | — |
| Find what the task touches | (manual `-scope`) | **target-resolution agent** (code + which tables + which docs) | **confirm the scope** |
| Assemble code context | `assembler.go` | **code-context action** (in-scope + call-graph neighbourhood + wiring-include) | — |
| Assemble data context | `dbcontext.go` | **data-context action** (schema + multipass row sizing) | — |
| Assemble runtime evidence | (placeholder) | **runtime-evidence agent** (run trace/logs/errors by `orchestration_id`) | — |
| Attach rules + relevant docs | constitution + manual `-doc` | **standards/docs agent** (always-on rules + *matched* guidelines, retrieved) | — |
| Compose the bundle | the assembler's output | **bundle orchestrator** (the bundle-shape contract) | **confirm the bundle** before generation |
| Produce the change | (the chat) | the generation step (frontier model) | — |
| Check the change | (none yet) | **checkers + paired fixers** (below) | **confirm before merge/deploy** |
| Govern + learn | designed, not built | decision log, trust ledger, change layer | (graduations gated) |

---

## Where documentation and guidelines fit (your catch)

The analyser reads **code only** — it never includes docs. The assembler includes the **constitution** (always-on rules) and any doc named with **`-doc`**; it does not *find* the relevant guidelines for a task. So today the documentation half is filled by hand: constitution for the rules, `-doc` for named docs, the project's attached knowledge for the rest.

Selecting the *right* guidelines for a task is the **standards/docs agent** above — matching + retrieval (the embeddings layer). It's deferred, not designed-away: it's a named slot in the bundle (the standards/reference section), currently filled manually. This is the same gap as the wiring-include for code (the call graph misses `registry.go`; manual `-doc` misses the relevant guideline) — both are "the tool can include it, but can't yet *find* it."

---

## The agents coming from different angles

Two kinds, and they're why separate containers help:

- **Context contributors** — code, data, runtime, standards. Each assembles one slice of the bundle from its own angle.
- **Checkers (with paired fixers)** — each reviews a proposed change from one angle, and these are the improvement-loop pieces:
  - **near-duplicate guardrail** (pre-merge; clone-check against the index),
  - **reuse check** (is there existing code to build on — retrieval, before generation),
  - **liability review** (does the output create legal exposure),
  - **morality check** (is this the right thing to put into the world, beyond mere legality — manipulation, exploitation of vulnerable users, misinformation, harmful claims, and whether a vertical should be served at all),
  - **correctness / standards conformance** (does it meet the constitution + matched standards).
  A checker raises a finding; a fixer acts on it. Both are ordinary agents/actions in the loop.

### The morality check — a configured standard, not a baked-in view

Morality is a distinct angle from liability: liability asks "will this get us sued or fined", morality asks "is this right", and the two overlap but aren't the same (some legal things are still wrong). It is deliberately **not** a single moral view hard-coded into the tool. Instead it is a checker that applies a **layered, configured standard** held in the active-config (a concern in the standards taxonomy, with provenance and effective-priority computed on read):

1. **A chosen base source** — a recognised, pluralistic framework the operator selects rather than the tool imposing one. For a system that builds marketing/business sites the concrete, authoritative candidates are advertising and consumer-protection standards (e.g. the UK ASA/CAP Code, CMA guidance: not misleading, substantiating claims, protecting the vulnerable, social responsibility), and for the AI-generation angle the recognised AI-ethics frameworks (e.g. OECD AI Principles, UNESCO, NIST AI RMF). The operator picks the source(s); the tool doesn't pick for them.
2. **Operator input and judgement** — the operator's own values layered on top, with priority over the base where they conflict.
3. **Jurisdiction and current-focus layers (later)** — current government guidelines and a "current focus" overlay, slotted in as higher-priority sources as they change. The standards layer already supports prioritised, dated, governed sources, so this is config, not new structure.

Two altitudes: **per-output** (is this specific content/feature moral) and a **vertical-level gate** (should we build this site/industry at all — a check at intake on the domain/objective). And because moral calls are often contested, this checker routes hard cases to a **HITL stop** (the operator's judgement) rather than auto-fixing — it surfaces the concern and the reasoning; the human decides. The tool applies the configured standard and flags; it is not the moral authority.

(The website-builder layer has its own "different angles" — research, content, design, imagery — same orchestration pattern, lower layer.)

---

## Generic, and with a front-end

- **Generic** comes from two things already designed: per-tenant **active-config** (rules/standards/objectives aren't hard-wired), and **pluggable per-language analysers** — Go now via `go/ast`, others later through the same bundle contract (the way deploy already has "github now, other adapters later"). Codebase-specific knowledge stays behind the analyser adapter and the tenant config.
- **Front-end**: the admin side shows the resolved scope, the assembled bundle, the HITL decisions, and the checker results — extending the existing admin dashboard. Public/self-serve is the later extension.

---

## What the thin-slice CLIs are

Prototypes of the indexer, code-context, and data-context actions — nothing more. A trial run tests an action's *logic*; what it teaches (the wiring-include rule, the scope-resolution signals, whether live rows are needed) lands as a feature of the action that ships. The trial's output is throwaway; the rule it teaches is durable. Build each trial piece in this shape so it graduates rather than ossifies as a script.
