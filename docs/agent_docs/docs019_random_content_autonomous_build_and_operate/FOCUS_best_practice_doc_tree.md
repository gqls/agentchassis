# FOCUS — Best-Practice Doc Tree (Cross-Cutting Concerns): Optimal Structure

**Status:** exploratory design. Greenfield — this proposes an optimal structure, not a migration of the current docs. Companion to `FOCUS_self_development_coding_pipeline_reasoning.md` (coordination positions A/B/C, the mediator, and the doc-vs-agent sort for concerns). Routing model discussed separately.

---

## 1. The two axes (recap)

Two orthogonal structures, kept separate and cross-referenced, never merged:

- **Objective tree (vertical):** mission → branch → leaf. Answers *should we do this, does it serve the goal here.* Read downward to the area in focus.
- **Concern tree (horizontal):** the best-practice / cross-cutting standards. Answers *is it done well.* Consulted sideways for any change, regardless of where in the objective tree it sits.

This document structures the concern tree.

---

## 2. Optimal unit: the atomic standard, not the document

The smallest addressable unit is a single standard (one rule), not a consolidated doc. Documents are **generated views** over the atoms. One source of truth; the constitution, the per-concern handbooks, and the machine-readable manifest are all produced from the same nodes, so nothing is maintained twice and nothing can drift between a "doc copy" and an "agent copy."

### 2.1 The standard atom

Each standard is a small node: structured frontmatter + a body split into separable parts.

Frontmatter:

| Field | Purpose |
|---|---|
| `id` | Stable identifier that never changes, e.g. `ARCH-0012`. References survive title/content edits. |
| `title` | Human label. |
| `concern` | One of the fixed top-level concerns (§2.3). |
| `scope` | `constitution` \| `domain` \| `leaf` — the load tier. |
| `applies_to` | Change types this governs: `go_action`, `sql_migration`, `workflow_json`, `agent_definition`, `adapter`, `frontend_component`, … |
| `kind` | `rule` (normative, enforced) \| `reference` (descriptive, not enforced). |
| `severity` | For rules only: `blocker` \| `should` \| `advisory`. |
| `status` | `proposed` \| `active` \| `deprecated` \| `archived`. |
| `version` | Increments on change; old version retained. |
| `supersedes` | `id` of the standard this replaces (for lineage). |
| `owner` | Human/team accountable. Locked content (see §4). |
| `check` | Name of the validator action that enforces it deterministically, or null if judgement-only. |
| `related` | `id`s of related standards. |

Body, in three separable parts so prompts pull only what they need:

- **Rule** — terse, imperative, prompt-loadable. The thing a builder/concern agent is given.
- **Rationale** — why it exists. Loaded only when an agent is reasoning about a tradeoff or a human is reviewing/changing the standard. Keeps prompt payloads tight.
- **Examples** — good/bad pairs. Loaded for generation guidance, optional.

### 2.2 Scope tiers (the root→leaf gradient)

- `constitution` — universal, non-negotiable, always loaded. Small set. E.g. reuse/search before creating, check schema before SQL, workflows simple with complexity in Go, every agent is an orchestrator, reply to the caller's topic, no `logger.Debug`, don't rename variables silently.
- `domain` — applies when a change touches that concern. E.g. the adapter response-envelope contract for `adapter` changes.
- `leaf` — narrow specifics loaded only when directly relevant. E.g. the exact `ActionInputs` method list, the `agent_definitions` required fields.

Tier is metadata on the atom, not folder depth — the same concern holds standards at all three tiers.

### 2.3 The concern set (top-level horizontal axis)

A small, deliberately stable set (changing this taxonomy is rare and human-decided — §4):

- architecture-and-boundaries
- messaging-and-contracts
- data-and-schema
- workflow-authoring
- code-style-and-api
- testing-and-validation
- observability
- model-and-llm-usage
- security-and-secrets
- performance-and-cost

### 2.4 Generated views (one source, many surfaces)

Nothing below is hand-maintained; all are produced from the atoms:

- **Constitution** = all `{scope: constitution, status: active}`. The always-on baseline.
- **Concern handbook** = all `{concern: X, status: active}`, grouped by scope. The human-readable "architecture standards" document.
- **Change-type bundle** = all `{status: active, applies_to ∋ T}`. This is exactly what the mediator composes for a change of type `T` — so the doc tree and the routing table are the same index queried differently.
- **Manifest** = the frontmatter of all active atoms, flattened. The machine lookup the mediator and concern agents consume.

### 2.5 Linkage to the objective tree

Concern tree and objective tree stay separate but cross-referenced. Each objective-tree area node carries `standing_concerns: [standard_ids | concern_names]` — the standards it always pulls (the adapter area always pulls messaging-and-contracts). For a given change the mediator unions: constitution + the area's standing concerns + the change-type's `applies_to` matches. (Natural storage hook for standing concerns if this lives on agents: `agent_definitions.domain_tags`.)

---

## 3. Why atomic-not-consolidated is the optimal trade here

The current corpus is consolidated (57 → ~26) for human maintainability, which is the right call when humans are the only readers. Once agents compose context per change and a mediator routes by concern, the constraints invert: you need independently-loadable, addressable, individually-versioned units, and you need to load only the relevant slice within a token budget. Atoms give that directly; large consolidated docs force all-or-nothing loading and copy-drift across docs.

The cost of atomic — that no single file reads as "the standards" for a human — is paid back by generated views (§2.4): humans read the generated handbooks, agents read the atoms and manifest, and both come from one source. So the readability the consolidation was protecting is preserved as an output, not the storage format.

---

## 4. Updating the tree (governance and lifecycle)

How standards and the tree change over time.

### 4.1 Ownership and locking
Standards are human-owned and locked (the `permanent` lock pattern). Agents never silently rewrite a standard. Spawn-fresh agents always read the current `active` version, so there is no stale-standard-baked-into-a-long-running-process problem.

### 4.2 Proposal → gate flow
Agents *propose*, humans *decide*. A concern agent (or any agent) that repeatedly hits a pattern raises a `propose_standard_change` item (new standard, edit, severity change, or deprecation) with rationale and observed evidence. A human gates it. Acceptance writes a new `version` and flips the old to `deprecated` (never deletes).

### 4.3 Deprecate, don't delete
Status lifecycle: `proposed → active → deprecated → archived`. `supersedes` records lineage; a `referenced-by` reverse index (derived from `related` and from objective-tree `standing_concerns`) shows what points at a standard before you deprecate it, so deprecation doesn't silently break references.

### 4.4 STEP ZERO for standards
Before adding a standard, the same does-this-already-exist discipline used for agents/actions: search active standards in the same `concern` + overlapping `applies_to` for a duplicate or a contradiction. A new standard that contradicts an active one in the same scope is a conflict to resolve, not a second copy to add. This is automatable as a pre-insert check an agent runs.

### 4.5 Evidence-driven changes
Changes are driven by observed signal, not opinion:
- Validator/auditor logs show repeated violations of a `should` → propose elevating to `blocker`.
- A doc-layer concern is repeatedly underweighted in practice → propose promoting it to a concern agent (the doc→agent promotion from the companion doc).
- A `check`-less rule that keeps being violated → propose writing a validator action for it (judgement-only → deterministic).

### 4.6 Drift / sync checks
Because rules are atomic and single-source, copy-drift is largely designed out. The remaining drift is rule↔enforcement: a periodic check that every `check:` points at a validator action that still exists, and flags `blocker` rules with no `check` (candidates for automation). 

### 4.7 Freshness cadence
Standards get a review rhythm rather than indefinite assumed-correct status — mirrors the audit-pass-cap rhythm: a periodic sweep surfaces `active` standards untouched for N days for a human freshness check. Prevents both churn and silent staleness.

### 4.8 Two change rates
- **Adding/editing a standard** — common, semi-automated via propose→gate.
- **Changing the concern taxonomy (§2.3)** — rare, deliberate, fully human. The top-level concern set is a slow-moving spine; treat changes to it as schema changes, not routine edits.

---

## 5. Deferred / open

- Physical layout of atoms (one file per standard vs a small number of structured data files, e.g. rows in a `standards` table generated into views). A table makes the manifest and views trivially queryable and versioned alongside the other operational data; flat files are easier to diff in git and review in PRs. Leaning toward a `standards` table as source with generated file views into the repo for human review and locking — but open.
- Exact `applies_to` vocabulary — needs to be finalized against the real set of change types the routing model classifies into (§ routing, next).

---

## 6. One-line state

Optimal concern tree = atomic standards with routing metadata, documents as generated views, one human-owned/locked/versioned source, updated via agent-proposes/human-gates with deprecate-not-delete and evidence-driven promotion. The atom's `applies_to` / `scope` / `severity` / `check` fields are the routing table — so the routing model (next) is mostly matching a change descriptor against these tags.
