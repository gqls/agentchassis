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

## 2. Stack-Discovery Agent (mechanical layer)

### 2.1 Responsibility and boundary

Owns the **mechanical layer** of the config (`PLAN_onboarding_config_derivation` §1). Detects what is in the repo so the rest of the system can rely on a stable shape: languages, layout, build and test commands, migrations, dependencies, CI. Produces a structured **mechanical-config document** with provenance and confidence per entry.

**Does not:** extract conventions (the conventions agent), elicit intent (the intent agent), modify the repo, or execute anything outside its declared probe plan. Reports what *is*, not what *should be*.

### 2.2 Inputs and outputs

**Inputs:** the repo (path/ref); for tenants, the sandbox configuration that gates probes (§2.6); for repeated runs, the previously confirmed mechanical config so drift can be flagged rather than overwritten.

**Output: the mechanical-config document** — a stable-schema artifact the rest of the system consumes. Illustrative shape:

```
schema_version: 1
detected_at: <ts>

languages:
  - { name: go, version: "1.22", source: inspection(go.mod), confidence: high }
  - { name: sql, dialect: postgres, source: inspection(migrations/), confidence: high }

module: { name: <module-path>, source: inspection(go.mod) }

layout:
  code_paths: { actions: actions/, adapters: adapters/, workflows: workflows/, migrations: migrations/ }
  doc_paths:  { root: ., index: 000_documentation_index.md }
  source: inspection(walk + pattern)
  confidence: medium

build:
  command: "go build ./..."
  source: probe + inspection(Makefile)
  last_probe: { ran_at: <ts>, exit_code: 0, duration_ms: 4200 }
  confidence: high

test:
  command: "go test ./..."
  source: probe + inspection(Makefile)
  last_probe: { ran_at: <ts>, exit_code: 0, passed: N, failed: 0 }
  confidence: high

migrations: { directory: migrations/, pattern: "NNN_*.sql", runner: <supplied>, ... }

uncertainties:
  - "Multiple Makefile targets could be the test command; chose `test`."
  - "Migrations runner not declared; needs supplied entry."
```

Stable-schema properties:
- Every entry carries `source` (inspection / probe / inferred / supplied), `confidence` (high / medium / low), and a probe result where applicable.
- Uncertainties are listed **separately** — they do not pollute confident entries.
- The schema is part of the engine, not per-tenant; what varies per tenant is what populates the fields.

### 2.3 Steps

1. **Inspect** (read-only). Read `go.mod`, Makefile, `.github/workflows/`, walk the tree, classify files by extension, find migration folders and doc roots. Emits inspection **facts**.
2. **Interpret** (proposes). Inspection facts → mechanical-config **proposals** (e.g., "this Makefile target named `test` is probably the test command"). Marked `confidence: medium` until probed.
3. **Emit probe plan** (declared). List the commands the agent intends to run, with rationale and expected duration. For tenants: the plan goes to the sandbox for approval. For our use: logged and approved.
4. **Probe** (executes, sandboxed for tenants). Run the planned commands; capture exit codes, stdout/stderr, duration; nothing else.
5. **Update confidence and propose.** Probe success raises confidence; probe failure is recorded as a fact and triggers diagnosis-as-candidate (§2.7). Emit the mechanical config with provenance and the uncertainties list.
6. **Human confirms / corrects** (confirm-not-initiate). The config is **active** only after confirmation.

### 2.4 Reuse

- **`go/packages`, `go/ast`** — Go module and layout inspection.
- **Existing build/test commands** — Makefile targets; the agent does not invent commands, it discovers and probes them.
- **The sandbox infrastructure** (Phase 3 of `PLAN_context_assembly_tool_and_service`) — probes run inside it for tenants.
- **Orchestrator/spawn pattern** — the inspect → plan → probe → confirm flow is a simple workflow with Go actions per step (complexity in actions, per the standing rules).

### 2.5 The subtle point — inspection is fact, interpretation is proposal

A file's contents are facts. The exit code of a probe is a fact. Deciding which Makefile target is "the test command", or that a directory called `actions/` holds the actions in this codebase's sense, is **interpretation** — and interpretation has uncertainty, even at the mechanical layer. So inspection results are recorded as facts; interpretations are recorded as proposals with confidence and source, and become facts only on confirmation (by probe success or by human). This is the inferred-then-confirmed pattern from the conventions agent, applied where it is least expected.

### 2.6 Sandboxing — the security envelope for tenants

This is the first agent that may execute tenant code (via probes). For tenants, probes run inside an ephemeral sandbox: the repo mounted read-only, no or restricted network, a time limit, no persistent state across probes. The probe plan emitted in step 3 is the security contract — the sandbox approves, restricts, or denies individual commands; the agent runs only what is approved. This is the Tier-C security concern made concrete and is **gating for the service phase**: no tenant code runs until sandboxing is solid. For our own use, sandbox detail is deferred but the **declared-probe-plan discipline is kept** — it is also useful as audit (what did the agent actually do).

### 2.7 Failure handling — a non-building repo

A probe that fails is itself useful output, not a failure of the agent. The mechanical config still records the attempted command and the result. The interpretation (broken code, missing deps, wrong command, transient flake) is **candidate-only, never a verdict** — consistent with the LLM-as-candidate principle (§1.9). A failing build during onboarding is surfaced as a finding: the tenant's repo is not currently in a buildable state, here is what was tried, here is what it returned. The agent does not attempt to fix it. Fixing is a maintenance task for the eventual build/verification harness.

### 2.8 Confirmation by reality (why this agent climbs the ratchet fastest)

The mechanical layer has a property the conventions and intent layers do not: it can be **confirmed by reality**. Inspection-and-interpretation propose; probe success confirms; the recorded commands actually working is the strongest confirmation any layer of the config can carry. This is why the mechanical layer can be near-automatic and high-confidence while the conventions and intent layers require human confirmation as the only available authority.

### 2.9 Dependencies

This is the foundation. No dependencies on other onboarding agents. The conventions agent (§1) and the intent-elicitation agent both depend on the mechanical config this agent produces (where the code is, how to parse it). Stack-discovery runs first in the onboarding sequence.

### 2.10 Open

- The exact mechanical-config schema (the shape in §2.2 is illustrative, not final).
- Multi-language tenants — adding language-specific inspector modules (Python, Node) follows the same pattern but is deferred until needed.
- Whether interpretive proposals (e.g., "this command *is* the test command") need human confirmation even when the probe succeeds, or whether a successful probe is sufficient. Leaning: probe success confirms the *command works*; identifying it as the test command specifically may still want a one-time human nod, because the same command might run a subset of tests.

---

*(Further agents — intent-elicitation, config-maintenance, onboarding orchestrator — to be added as sections below.)*
