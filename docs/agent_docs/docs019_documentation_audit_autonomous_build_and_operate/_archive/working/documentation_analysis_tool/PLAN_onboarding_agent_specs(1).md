# PLAN — Onboarding Agent Specs

**Status:** per-agent expansion of `PLAN_onboarding_config_derivation.md` §6. One section per onboarding agent, built up over turns. The onboarding plan is the overview; this is the detail.

---

## 1. Conventions Agent (docs-authoritative mode → drift audit)

### 1.1 Responsibility and boundary

Owns the **conventions layer** of the config. In docs-authoritative mode (our own onboarding, decided): extracts the conventions from the standards docs (the authoritative source) and audits the code for **disagreements** — places where the code violates a documented convention. Produces two things: the confirmed convention set (the conventions layer of the config) and the drift audit (the disagreements).

**Does not:** fix code (a maintenance task), change standards (a human does, via confirm-not-initiate), elicit intent/why-chain (the intent agent), or derive the mechanical layer (the stack-discovery agent). It detects and reports; it does not resolve. Distinct responsibility, minimal overlap.

### 1.2 Inputs and outputs

**Inputs:** the standards doc set (001, 003, the naming FOCUS — paths/ids); the repo (path/ref); the mechanical config from the stack-discovery agent (where the code is, how to parse it — §1.7); the source-of-truth mode (`docs-authoritative`).

**Outputs:**
- **Convention set** — discrete conventions extracted as atoms, each with: an id, the rule, the **exact doc citation** (provenance = doc-sourced), and a `checkable` flag (deterministic / judgement).
- **Drift audit** — disagreements, each with: code location (file, symbol, line), the convention violated (atom id + citation), how detected (which check), confidence, and a **default disposition** (code-drifted, since docs are authoritative) — recorded, not resolved.

### 1.3 Steps

1. **Extract conventions from each doc** → discrete checkable atoms, each citing the exact source span; classify each as deterministically-checkable or judgement. (LLM action, or a sub-agent per doc.) This is also the Phase-2 tagging work from the adoption plan — the extracted conventions are the atoms.
2. **Confirm the convention set** (§1.5) — proposed conventions with citations go to a human; only confirmed conventions become checks. This precedes any audit.
3. **Check the code** — for each confirmed checkable convention, map it to an existing `check_*.go` validator or a new check, and run it over the code. The checking model has three tiers depending on the convention's nature (see §1.9): deterministic check, heuristic proxy that flags candidates, or judgement-only (not audited). Tiers produce different output classes — violations, candidates, or unchecked — not one mixed result.
4. **Record disagreements** — each with location, convention, detection method, **tier**, confidence, and the default disposition. Recorded, never auto-resolved.
5. **Emit the drift audit and route dispositions** — each disagreement carries three possible resolutions: code drifted (fix the code), doc drifted (update the standard), or legitimate exception (record it). Docs-authoritative sets the *default* presumption (code drifted), not the only one. The human confirms the disposition (confirm-not-initiate); the agent proposes, never fixes or rewrites.

### 1.4 Reuse

- **`check_*.go` validators** — the existing checking layer (component standards, naming, missing structure, broken links, etc.); the conventions agent's checks are these plus new ones for uncovered conventions.
- **Go analysis** — `go/ast`, `go/packages`, gopls for naming, signatures, the response-envelope shape, `logger.Debug` usage, variable-name checks.
- **Atomic-standards format** — extracted conventions are atoms (doc-tree FOCUS), so this step doubles as standards tagging.
- **confirm-not-initiate** — disposition and convention confirmation.
- Workflow stays simple; the extraction (LLM) and the checks (AST) are Go/LLM actions — complexity in actions, per the standing rules.

### 1.5 The subtle point: extraction is inferred-then-confirmed

Docs-authoritative means the docs are the source, not that extraction is infallible. The agent can miss a convention (false negative) or invent one (false positive). Auditing code against an invented convention manufactures drift that isn't real — the one failure that would discredit the audit. Mitigation: every extracted convention **cites its exact doc span**, so it is traceable and verifiable, and the convention set is **human-confirmed before** the audit runs. Extraction proposes; a human confirms; only then does the code get checked.

### 1.6 Coverage honesty, confidence, exceptions

- **Coverage honesty.** Many conventions are not deterministically checkable. The audit reports which conventions were checked and which could not be, so its completeness is known — it must not imply a clean result when half the conventions were judgement-only and unchecked.
- **Confidence.** Static checks have false positives (a naming check flags a legitimate exception). Each disagreement carries a confidence; the disposition step (§1.3.5) handles false positives.
- **Exceptions feedback.** When a human marks a disagreement as a legitimate exception, that is recorded against the code location + convention, so the next audit does not re-flag it. The audit becomes incremental — it learns accepted exceptions. (Same provenance/lifecycle as the standards.)

### 1.7 Dependencies

Assumes the **mechanical layer** from the stack-discovery agent: where the code is, the languages present, how to parse/build it. So stack-discovery runs first. The conventions agent is the second onboarding step, not the first.

### 1.8 Open

- Whether the docs-vs-code disagreement audit is a separate reporting concern or stays inside this agent's output.
- Mode switch to code-inference-then-confirm for doc-poor tenants (the harder, more common case) — deferred; this spec is docs-authoritative.

### 1.9 Handling judgement (non-deterministic) conventions

Many conventions cannot be settled by a static check. The audit handles this with three tiers, plus a role split that keeps un-auditable conventions useful.

**Three checking tiers:**

1. **Deterministic** — a static check settles it. Naming, `logger.Debug` usage, response-envelope shape, schema-before-SQL adherence. Existing `check_*.go` + go/ast. Output: **violations**, high confidence.
2. **Heuristic proxy** — a measurable indicator flags *candidates*, not violations. "Workflows simple" → workflow step count, nesting depth, conditional logic presence. "Reuse before recreate" → signature/embedding similarity between a new function and existing ones. The proxy says *where to look*, not *what's wrong*. Output: **candidates**, low confidence; route to human review (or an optional LLM-judgement pass — see below).
3. **Judgement-only** — no checkable proxy. "Keep responses pragmatic," much architectural taste. Output: **unchecked**, reported as a coverage gap, not a violation.

**LLM-judgement on heuristic candidates (optional).** An LLM can read the candidate and the convention and add an opinion. It is *still only a candidate flag*, never a verdict. An LLM judging convention adherence has the same uncertainty as an LLM writing code, so it gets the same treatment used everywhere else: a flag for human attention, not a settled outcome. The human is the verdict; the LLM narrows where to look. This keeps the audit's authority intact and avoids creating a weaker authority channel.

**Audit vs guidance — the role split.** Some conventions belong in **generation context**, not in the audit. Tier-3 conventions ("keep pragmatic") can't be audited but can shape what gets generated when loaded into the builder's prompt. A convention atom can play up to two roles: **audited** (appears in the drift audit) and **guiding** (loaded during generation). Some atoms play one, some the other, some both. This stops un-auditable conventions from being lost — they contribute on the input side rather than the output side.

**Three-bucket audit output (the operational form of coverage honesty, §1.6):**
- *Deterministic violations* (count + locations).
- *Heuristic candidates* (count + locations, flagged for review).
- *Unchecked judgement-only conventions* (count, reported as coverage gap).

The audit reports three numbers, not one. A clean tier-1 count alongside many tier-3 unchecked is not a clean audit; it is a partial audit with known limits.

**Concrete examples (our repo):**
- "Workflows simple, complexity in Go" — heuristic: workflow step count / depth. Candidate review: "is this complexity legitimate or should it move to Go?"
- "Reuse before recreate" — heuristic: similarity search between new and existing functions. Candidate review: "should this be a reuse of X?"
- "Don't change variable names silently" — heuristic at diff time (detect renames). Candidate review for past commits without explicit acknowledgement.
- "Every agent is an orchestrator" — partly checkable (structure) + judgement (does it genuinely orchestrate). Heuristic on structure, candidate-review on the rest.
- "Keep responses pragmatic" — guidance only; goes into generation context, not the audit.

---

*(Further agents — stack-discovery, intent-elicitation, config-maintenance, onboarding orchestrator — to be added as sections below.)*
