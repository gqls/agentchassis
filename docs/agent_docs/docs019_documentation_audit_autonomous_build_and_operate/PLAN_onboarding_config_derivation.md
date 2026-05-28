# PLAN — Onboarding: Deriving and Maintaining a Repo's Config

**Status:** plan, expanding the onboarding discussion. Companion to `PLAN_context_assembly_tool_and_service.md` (§6 there names onboarding as the hard service problem; this is its detail). The agents in §7 are headings to expand in later turns.

Onboarding is where "documentation is code" gets tested: the config is itself authored knowledge about a repo, so deriving it is the same authored-versus-derived problem as everything else, one level up.

---

## 1. The config is three layers with different derivability

Lumping these together makes onboarding look like one hard problem instead of three uneven ones.

- **Mechanical — discoverable.** Languages, package/module layout, how the repo builds and tests, where migrations live. Mostly already in the repo (go.mod, Makefile, CI config, directory structure). A scanner derives it by inspection; where it can't, it **probes** (run the build, see if it succeeds). Low-stakes, confirmable by observing whether verification actually runs. **The mechanical layer also records the structural facts the bundle's capabilities depend on** — whether behaviour is stored as data and where, whether a run-correlation key spans the telemetry tables, whether steps are named and logged, how logs are fetched (`PLAN_onboarding_agent_specs` §2.9). These are rich on our own codebase (so we build the optimal version against it first) and may be absent or partial on a tenant's, so each dependent capability **degrades where its fact is missing rather than breaking.** This is the engine/config split made concrete: the engine is generic; the config records what this codebase actually provides.
- **Conventions — inferred or doc-sourced.** Naming (kebab vs snake), the response-envelope shape, the orchestrator/handler boundary, "workflows simple, complexity in Go." Inferable from code by example, but code shows what it *does*, not what it's *supposed* to do — they differ exactly where the code is wrong or mid-refactor. So inference is a **strong draft, weak authority** → confirm. This is the authored-by-inference category: a likely convention reconstructed from instances, authoritative only on confirmation, because a convention is intent and intent isn't fully recoverable from instances.
- **Intent and standards — elicited.** Why-chain, priority profile, direction-of-travel, the standards themselves. Largely **not derivable from source** — you can't read "security outranks speed here" or "deliberately simple for now" off the code. Supplied by the tenant. Also the part that delivers most of the tool's distinctive value.

So onboarding is **three processes, not one**: automatic for the mechanical layer, inferred-then-confirmed for conventions, a structured interview for intent.

---

## 2. Progressive, not all-or-nothing

A tenant should get value from the mechanical layer alone — fresh code context, signatures, reuse search, schema — **before** any intent is captured, because that half already beats pasting into a chat. Conventions and intent are filled incrementally and the tool deepens as they arrive. A tool that demands a complete config before doing anything has a cliff at the start; this has a ramp. It also means onboarding is **never "done"** — it tracks the repo (§4).

---

## 3. The config is a maintained artifact (same lifecycle as the docs)

"Onboarding as a first-class deliverable" means the config is a maintained artifact with the standards' lifecycle, not a good setup wizard. The wizard is the first pass; the lifecycle is the deliverable.

- **Drift detection** — periodically re-derive the mechanical layer and re-check conventions against code; flag where reality diverged from the recorded config.
- **Confirm-not-initiate** — proposed config changes are human-confirmed.
- **Provenance per entry** — each config entry records whether it was discovered, inferred, or supplied, which sets how much to trust it and who may change it.

This reuses the upkeep machinery already designed for the standards (drift detector, confirm-not-initiate, provenance).

---

## 4. The tension: inference quality scales with codebase quality

On a clean, consistent repo, convention inference is strong and onboarding is fast. On a messy or inconsistent repo — most real repos — inference produces a confident draft of conventions that are actually the repo's bad habits, and confirming that draft codifies the mess as the standard. So **the more a tenant needs the tool (because their repo is inconsistent), the less their repo can teach it**, and the more the intent layer must be elicited rather than inferred.

Mitigation: the tool **surfaces its uncertainty**. When inference finds inconsistency ("functions are named three different ways here"), it presents that as a question to resolve, not a silent majority pick. **Inconsistency detected during onboarding is itself valuable output** — it tells the tenant where their own conventions aren't actually conventions.

---

## 5. Decision: source-of-truth for conventions

**General rule:** choose per tenant by doc availability — docs-authoritative where good docs exist, code-inferred-then-confirmed where they don't. This choice is itself a recorded onboarding decision, an area of documentation and decision-making in any existing or adopted codebase/documentation system.

**Ours (decided):** **docs-authoritative.** 001, 003, and the naming FOCUS doc are the source of conventions; the code is used **only to find disagreements**. We've already looked at these conventions, so this gives a baseline and makes onboarding easier. Where the conventions extracted from the docs disagree with what the code does, the disagreement is **recorded, not resolved silently** — that set of disagreements is a free audit of where our codebase has drifted from its own documented standards. It is the drift detector's first run, on us, before any tenant sees it.

---

## 6. Onboarding as a set of agents (one per problem)

Each sub-problem becomes an agent (every agent an orchestrator; reuses the curator/advocate/coordinator patterns and spawn machinery — agents-per-responsibility, not new infrastructure). Headings to expand in later turns.

- **Onboarding orchestrator** — runs the three layers progressively, assembles the config, routes confirmations (confirm-not-initiate). Delivers mechanical value first, then deepens.
- **Stack-discovery agent (mechanical)** — scans the repo, proposes the mechanical config, and **probes** to confirm (runs the build/test). Confirmed by reality, low-stakes.
- **Conventions agent** — in docs-authoritative mode (ours), extracts conventions from the standards docs and runs the code **only to find disagreements**; surfaces inconsistency as questions. Has a mode switch to code-inference-then-confirm for doc-poor tenants. The source-of-truth choice (§5) is its key parameter.
- **Intent-elicitation agent** — a descendant of the briefing questionnaire / intake orchestrator, pointed at a codebase's intent rather than a website brief. Structured interview → why-chain, priority profile, direction-of-travel.
- **Config-maintenance agent** — ongoing drift detection over the config (§3): re-derive mechanical, re-check conventions, flag divergence for confirm-not-initiate.
- **Cross-cutting principle (all agents):** surface uncertainty rather than silently picking; record provenance on every entry.

(The docs-vs-code disagreement audit is produced by the conventions agent in docs-authoritative mode; whether it's a separate reporting concern is open.)

---

## 7. Our own onboarding as the template

Not special-cased — it is the template, run on us first:
- **Mechanical:** scan — Go, Postgres, the Makefile, the migrations, the chassis layout.
- **Conventions:** extract from the existing standards docs (001, 003, naming) per §5; run code only to surface disagreements (the drift audit).
- **Intent:** capture via the constitution and the objective-tree why-chain (the adoption-plan work).

Doing our own onboarding this way does double duty: it produces our config **and** exercises and shapes the onboarding we'll later offer tenants, with the advantage that we can check the output against our own knowledge of the repo. The first deliverable beyond the config is the docs-vs-code disagreement list — the drift audit.

---

## 8. Open / next

- **Expand each agent in §6** into a spec (the user's "address each as a potential agent").
- **Mode-switch criteria** for the conventions agent — when docs are "good enough" to be authoritative versus falling back to inference.
- **Intent-elicitation depth** — how much interview is enough for useful initial value before the ramp continues.

---

## 9. One-line state

The config has three layers — mechanical (discovered + probed), conventions (doc-sourced for us, inferred-then-confirmed otherwise), intent (elicited) — onboarded progressively (mechanical value first), maintained as a first-class artifact with drift detection, confirm-not-initiate, and per-entry provenance. The harder a repo's state, the less it can teach, so uncertainty is surfaced rather than silently resolved. For our own onboarding, the standards docs are authoritative and code is used to find disagreements, which doubles as a drift audit. Each sub-problem maps to an agent (§6), to be expanded next.
