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

## 3. Intent-Elicitation Agent (intent layer)

### 3.1 Responsibility and boundary

Owns the **intent layer** of the config — the elicited tier: the why-chain (objective-tree purpose), the priority profile per node, the direction-of-travel for areas in active work, and the standards where they do not already exist. Largely not derivable from source code; the tenant is the source.

**Conducts a progressive, value-returning interview** to capture intent in small exchanges that grow with usage rather than asked up front.

**Does not:** extract conventions (conventions agent), inspect the mechanical layer (stack-discovery agent), make decisions on the tenant's behalf, or edit code. Captures and structures intent; the tenant authorises.

### 3.2 Inputs and outputs

**Inputs:**
- The mechanical config (anchors questions to specific areas — "you have an `actions/` directory; what is its purpose?").
- Any existing docs (constitution, README, mission statements) that allow **proposal-confirmation** rather than blank-page questioning.
- Live tenant interaction.

**Outputs:**
- The **why-chain** — objective-tree node atoms with stable ids, each carrying a one-line `why`.
- **Priority profile** per node (ordered dimensions, sealed constraints called out).
- **Direction-of-travel** notes for areas in active work — settled-not-to-relitigate decisions, deliberately-temporary states, freshness-stamped (per `FOCUS_salience_and_multi_author_mediation` §9).
- All entries carry `source: supplied(human)` provenance.
- A **coverage report** — what has been captured, what is still pending. Never claims complete.

### 3.3 Steps (the progressive interview)

1. **Bootstrap from existing material.** If a constitution / mission doc / README exists, propose a mission and a small set of top-level areas from it; tenant confirms or edits. If not, ask one question: "in one sentence, what is the purpose of this codebase?"
2. **Per top-level area** (proposed from the mechanical layer where possible): one-sentence purpose + a short exchange on priority profile (rank a few dimensions, flag any sealed constraints).
3. **For areas in active work** (tenant indicates which): a brief prompt for direction-of-travel — any settled decisions, anything deliberately temporary, anything still in flux.
4. **Mark the rest as pending.** The intent layer is **incomplete by design** at this stage. Coverage report shows what is missing.
5. **From this point, leaf-level intent is just-in-time** (§3.7) — captured as the tool helps with tasks, not as an upfront tax.

### 3.4 Reuse

- **The briefing questionnaire and intake orchestrator** — the same structured-elicitation pattern, pointed at a codebase's intent rather than a website brief. Descendant, not new.
- **Atomic standards format** — captured intent is stored as objective-tree node atoms with the `why`, `priority_profile`, and `direction_of_travel` fields.
- **LLM-driven dialogue** for adaptation and proposal generation.
- **Confirm-not-initiate** for every captured piece — the agent proposes; the tenant authorises.

### 3.5 The subtle point — proposal-confirmation vs free elicitation

The interview **interleaves two modes**:
- **Proposal-confirmation** — where the agent has evidence (existing docs, the mechanical layer, prior answers, code patterns), it proposes and the tenant confirms or edits. Low friction.
- **Free elicitation** — where it has no evidence, it asks a direct question. Higher friction but unavoidable.

Pure free elicitation is the blank-page problem; pure proposal-confirmation **anchors** the tenant on the agent's guess. The agent leans proposal-confirmation when evidence exists, free elicitation when it does not.

**Anchoring risk in proposal-confirmation:** a tenant rubber-stamps a poorly-guessed proposal because it is easier than authoring. **Mitigation:** every proposal cites its evidence ("based on your README's mission statement, you have these 5 main areas — does this look right?"), so the tenant can see *why* the proposal was made and is more likely to spot a wrong inference. The proposal is contestable, not implicit.

### 3.6 Value-return per exchange (answering "how much to ask before it feels worth it")

The interview is structured so each small exchange **delivers something useful in return**:
- Capturing the mission → it appears as the top of the why-chain in every bundle from the next task onward; generations stay tethered to it.
- Capturing one area's purpose → bundles for tasks in that area immediately carry that purpose at the right altitude.
- Capturing one area's priority order → the mediator (when active) immediately uses it to settle tradeoffs there.

The tenant should never feel they have given several answers and gotten nothing back. Each captured piece changes the next interaction. This is what makes the interview earn its asks.

### 3.7 Just-in-time leaf-level intent during use

Leaf-level intent — the specific purpose of a feature, the rationale for a particular change — is **not asked up front**. As the tool helps with a specific task, it asks at that moment: "what is the purpose of this feature?" or "what is the why behind this change?" — one question, in the flow of work, with concrete context to ground the answer. The answer is captured to that leaf's node atom.

The intent layer fills out as the tool is used, not as a setup tax. This is the key answer to "how to keep the interview progressive": **the interview is not finite**. Initial onboarding captures the top of the tree; the leaves are captured as work happens. The tool gets better as the tenant uses it — usage and capture reinforce each other.

### 3.8 Relationship to the user-rep advocate

Both agents work with intent, in different roles:
- **The intent-elicitation agent captures** intent (onboarding role — this agent).
- **The user-rep advocate uses** intent (runtime role at decision points — `FOCUS_salience_and_multi_author_mediation` §8).

Same intent concept, two roles, two agents. The advocate reads the why-chain and direction-of-travel this agent populates; this agent gathers what the advocate later defends. Capture and use are kept distinct so the elicitation is not pressured by live decision-making and the advocacy is not distracted by capture.

### 3.9 Dependencies

Depends on the **stack-discovery agent's mechanical config** to anchor questions to specific code areas. Benefits from any existing docs for bootstrapping. The conventions agent's output is independent — they can run in parallel after stack-discovery completes.

### 3.10 Open

- **Detecting rubber-stamping.** Very fast confirmation of dense proposals is a signal the tenant is not engaging — flag for re-prompt or a small explicit push-back ("I want to check — does this actually fit, or should we look closer?"). How aggressively to do this is open.
- **Pacing of just-in-time leaf questions.** Too eager interrupts work; too sparse and the leaves never fill. Probably tied to how often the tool is invoked on a given area, plus an explicit "ask me about purpose for this task" opt-in.
- **Pivots.** How a tenant signals "we've changed direction here" and triggers re-elicitation — overlaps with the config-maintenance agent's job.

---

## 4. Onboarding Orchestrator

### 4.1 Responsibility and boundary

Coordinates the three layer agents (stack-discovery §2, conventions §1, intent-elicitation §3) into the onboarding flow and surfaces progress and proposals to the tenant. Owns the **onboarding state** — what is confirmed, what is pending, what is blocked. Reuses the existing orchestrator pattern (every agent is an orchestrator); this is the one that orchestrates the others.

**Does not:** do extraction / checking / elicitation itself (delegates to the layer agents). Does not modify the repo. Does not make policy decisions — every gate is confirm-not-initiate, routed to the tenant.

### 4.2 Inputs and outputs

**Inputs:** the repo (path/ref); the tenant identity (for the service); the sandbox configuration (for tenants); the **source-of-truth mode** for conventions (docs-authoritative vs code-inference — decided at start with the tenant); any previously confirmed config (for re-onboarding).

**Outputs:**
- The **active config** — the layered (mechanical + conventions + intent) configuration with provenance per entry.
- The **onboarding state** — a tenant-facing summary (§4.5).
- The **initial drift audit** from the conventions agent (in docs-authoritative mode).

### 4.3 Steps (the flow)

1. **Initialise.** Confirm the source-of-truth mode with the tenant (docs-authoritative for good-docs repos like ours, code-inference for doc-poor ones). Record the decision. Set up tenant isolation if applicable.
2. **Run stack-discovery.** Receive the proposed mechanical config + uncertainties; route to the tenant for confirmation; record the confirmed mechanical config. **Mechanical layer = active.**
3. **Run conventions and intent in parallel** (now that the mechanical layer is active — both depend on it; neither depends on the other, per §1.7 and §3.9):
   - **Conventions agent** → extracted conventions + drift audit + uncertainties; tenant confirms convention set (which precedes the audit being trusted, §1.5); **conventions layer = active**.
   - **Intent-elicitation agent** → top-of-tree progressive interview; captured intent atoms; **intent layer = partial-active** (always partial by design, §3.7).
4. **Mark active-with-pending.** Onboarding never reaches "fully done": active-with-pending is the steady state. The orchestrator hands ongoing drift management to the config-maintenance agent (§5, next).
5. **Tenant can use the tool from this point.** Just-in-time leaf-level intent capture (§3.7) and ongoing drift detection (§5) take over from the orchestrator's bounded role.

### 4.4 Coordination patterns

- **Sequencing follows dependencies, not policy.** Stack-discovery first because conventions and intent both depend on its mechanical config. Conventions and intent in parallel because they do not depend on each other. The dependency graph **is** the flow plan.
- **Gates.** Each layer agent emits proposals; the orchestrator routes them to the tenant via confirm-not-initiate. The tenant confirms in any order; nothing waits on confirmation of unrelated layers.
- **No cross-layer arbitration under normal flow.** The three layers don't typically conflict. If a contradiction surfaces (intent says "this area's purpose is X" but conventions extracted from docs say something incompatible), the orchestrator escalates it as a finding — rare in practice.

### 4.5 Onboarding state — what is surfaced to the tenant

A small structured artifact, clear enough that a tenant knows where they are without reading the full configs:

```
mechanical:  confirmed | partial(<reason>) | blocked(<reason>)
conventions: confirmed (N atoms, M extracted)
             drift audit: { deterministic: V violations, heuristic: C candidates, unchecked: U }
intent:      top_of_tree_captured: { areas: A, with_why: A, with_priority: B }
             coverage: <partial — see report>
pending:     [ list of proposals awaiting tenant confirmation ]
blocked:     [ list of things the orchestrator can't proceed with, with reasons ]
```

This is the tenant-facing summary. The full configs sit behind it.

### 4.6 Failure handling

- **A layer agent that can't complete does not stop the others.** Stack-discovery's probe fails (non-building repo) → mechanical layer is partial; conventions and intent still run with what was inspected. The orchestrator records the blocker and surfaces it.
- **A tenant who walks away mid-interview** → onboarding pauses; resumes on next use; nothing lost.
- **A layer-agent error (exception, timeout)** → retry once; if persistent, mark the layer as `blocked(<error>)` and surface to the tenant.

### 4.7 Reuse

- **Orchestrator/spawn pattern** — this *is* an orchestrator, running on the existing machinery. No new infrastructure.
- **Work-item / coordination model** — for tenants, the onboarding state and pending confirmations may persist as work-items (resumable across sessions); for our own use a simpler in-memory state may suffice. (Open, §4.9.)
- **Mediator pattern** — only needed if cross-layer arbitration is required (rare); under normal flow the orchestrator just sequences and surfaces.
- **Same code for our use and for tenants.** Only the entry point and the isolation context differ — consistent with the engine/config split principle from `PLAN_context_assembly_tool_and_service`.

### 4.8 Dependencies

Depends on the three layer agents (§1, §2, §3). No other dependencies; the orchestrator is the entry point for onboarding.

### 4.9 Open

- **Persistence shape.** Onboarding state as work-items (full coordination model, resumable) for tenants vs simpler in-memory structure for our own use.
- **Re-onboarding** after a major repo restructure — incremental refresh (existing confirmed entries stay until contradicted) vs full re-run. Incremental is the lean.
- **UI / API surface** — how a tenant actually interacts (CLI, web, API behind gateway). Implementation detail deferred to the tool-and-service plan.

---

*(Further agents — config-maintenance — to be added as section 5.)*
