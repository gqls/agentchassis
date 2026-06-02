# PLAN — Active Config Schema (the load-bearing contract)

**Status:** contract specification. The schema every onboarding agent produces or consumes, and every downstream consumer reads. Settles the §3.1 gap from `FOCUS_onboarding_system_view_check`. Storage is Postgres, consistent with the existing chassis. The actual DDL is implementation; this doc defines the *shape* and the *semantics*.

---

## 1. What the active config is

The **logical union** of the three onboarding layers:
- **Mechanical** — from the stack-discovery agent (`PLAN_onboarding_agent_specs` §2).
- **Conventions (concern atoms, flat)** — from the conventions agent (§1).
- **Objectives (why-chain atoms, nested)** — from the intent-elicitation agent (§3).

Operationally it lives in **four Postgres tables**, joined per-tenant. There is no monolithic blob — consumers read whatever slice they need. "The active config for tenant T" is the join of these tables filtered by `tenant_id` and `status = active`.

**Prerequisite — the governed vocabularies are seeded first.** `standards.concern` classifies into a fixed concern taxonomy, and the priority profiles classify into a fixed set of dimensions. Both are governed vocabularies the conventions and intent agents sort *into*, so they must be authored/seeded before those agents run. They are not derived per tenant; they are the fixed sets everything else references.

---

## 2. The common provenance shape

Used **identically across all three layers** so downstream consumers (bundle builder, advocate, mediator, maintenance) can reason about provenance uniformly without per-layer branching.

> **Chassis conventions (verified — `FOCUS_schema_verification_findings` §1):** enumerated values are `text` with a `CHECK` constraint, **not** native Postgres enums; versioning is `version integer` + `previous_version_id uuid`; soft delete is `deleted_at timestamptz`, not a `status = archived`. The table below is corrected to these; the same applies to every other contract table.

| Field | Type | Purpose |
|---|---|---|
| `source` | text + CHECK | How the entry was obtained: `inspection` / `probe` / `inferred` / `supplied` / `doc_sourced` |
| `source_ref` | jsonb | Pointer to the source artifact: file path + line range, probe id, doc citation, etc. Shape varies by source. |
| `confidence` | text + CHECK | `high` / `medium` / `low`. Set by the producing agent. |
| `status` | text + CHECK | Lifecycle: `proposed` / `active` / `deprecated`. (Soft-delete via `deleted_at`, not an `archived` status.) |
| `last_verified_at` | timestamptz | When the entry was last confirmed or re-confirmed. |
| `verified_by` | text | User id or system process id that confirmed. |
| `freshness_until` | timestamptz, nullable | Optional expiry, after which re-confirmation is required (direction-of-travel uses this). |
| `version` | integer | Monotonic per-entry version. Bumped on edit. |
| `previous_version_id` | uuid, nullable | Self-FK to the prior version, for lineage (chassis convention; replaces the earlier `supersedes`). |
| `deleted_at` | timestamptz, nullable | Soft delete (chassis convention). |

These fields appear on every storable entry (mechanical fields, standards atoms, objectives atoms) — whether as table columns or as embedded JSONB metadata (mechanical layer; §4.2).

---

## 3. The four tables

### 3.1 `tenant_configs` — one row per tenant

Top-level metadata. The entry point for any "give me the config for T" query.

| Field | Type | Purpose |
|---|---|---|
| `tenant_id` | uuid PK | Identifier. For our own use, a **real tenant row** (not a sentinel or special case), so the single-tenant path exercises the same code as the multi-tenant one. |
| `schema_version` | int | Active schema version (for migrations). |
| `source_of_truth_mode` | enum | `docs_authoritative` / `code_inference` — set at onboarding (§4.3 of the agent specs). |
| `onboarding_state` | enum | `not_started` / `mechanical_active` / `active_with_pending` / `re_onboarding` |
| `created_at`, `updated_at` | timestamptz | Standard. |

### 3.2 `mechanical_config` — one row per tenant

The mechanical layer is a small, fixed-shape document. JSONB on a single row with **per-field embedded provenance** is cleaner than exploding into many columns.

| Field | Type | Purpose |
|---|---|---|
| `tenant_id` | uuid FK | One row per tenant. |
| `body` | jsonb | The mechanical-config document per §2.2 of the agent specs. Each entry inside carries its own `{ value, source, source_ref, confidence, last_verified_at, verified_by }`. |
| `uncertainties` | jsonb | List of uncertainties surfaced by stack-discovery (§2.2). |
| `last_full_scan_at` | timestamptz | Last time the agent did a complete re-inspection. |

The `body` is the structured document — `languages`, `module`, `layout`, `build`, `test`, `migrations`, `ci`, etc. Per-field provenance lives inside `body`, not as outer columns, because the structure is nested and known.

### 3.3 `standards` — one row per convention atom (concern tree, flat)

Per `FOCUS_best_practice_doc_tree` §2.1. Multi-tenant: filter by `tenant_id`.

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | Stable identifier. |
| `tenant_id` | uuid FK | Scope. |
| `concern` | text | The owning concern (singular — `architecture-and-boundaries`, `messaging-and-contracts`, etc.). The ownership key. |
| `scope` | enum | `constitution` / `domain` / `leaf` |
| `applies_to` | text[] | Change types this governs. Plural. |
| `kind` | enum | `rule` (normative) / `reference` (descriptive) |
| `severity` | enum, nullable | For rules: `blocker` / `should` / `advisory`. |
| `title` | text | Human label. |
| `rule` | text | Terse, prompt-loadable. |
| `rationale` | text, nullable | Loaded only for tradeoff reasoning or human review. |
| `examples` | jsonb, nullable | Good/bad pairs, optional. |
| `check` | text, nullable | Name of the deterministic validator action; null if judgement-only or heuristic. |
| `check_kind` | enum | `deterministic` / `heuristic` / `judgement_only` — drives the three-tier audit output (§1.9). |
| `related` | uuid[], nullable | Related atom ids. |
| `owner` | text | Human/team accountable. |
| *...provenance fields per §2...* | | `source` is typically `doc_sourced` for atoms from existing docs, `supplied` for human-authored ones. `source_ref` cites the doc span (e.g., `{doc: "001_development_guide.md", lines: [42, 58]}`). |

Indexes: `(tenant_id, status, concern)`, `(tenant_id, status)` GIN on `applies_to`.

### 3.4 `objectives` — one row per why-chain node (objective tree, nested)

Per `FOCUS_salience_and_multi_author_mediation` §9. Multi-tenant: filter by `tenant_id`.

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | Stable identifier. |
| `tenant_id` | uuid FK | Scope. |
| `parent_id` | uuid FK, nullable | Tree structure. NULL for the root (mission). |
| `name` | text | Short label. |
| `why` | text | One-line purpose. |
| `priority_profile` | jsonb | The profile per §9.6 of the salience doc: ordered objectives + sealed constraints + dimension entries. Shape: `{ order: [<dim>, <dim>, ...], constraints: [{dim, kind: sealed\|relaxed-from-sealed, premise}], dimension_entries: [{dim, severity, source_event}] }`. Inherited at read time. |
| `direction_of_travel` | jsonb, nullable | `{ current_heading, settled_not_relitigate[], deliberate_temporary[] }` with freshness on each. |
| `standing_concerns` | uuid[], nullable | `standards.id` entries this node always pulls (the linkage from §2.5 of the doc-tree FOCUS). |
| *...provenance fields per §2...* | | `source: supplied` for human-authored; `source_ref` may cite an intake interview turn or a confirmed prior version. |

Indexes: `(tenant_id, parent_id)`, `(tenant_id, status)`.

**Effective priority profile** is *computed at read time* by walking from root to the node and merging entries per §9.1 of the salience doc — not stored. This is the "store authored differences, compute effective on read" rule that keeps ancestor changes automatically reflected below.

**Acyclicity is required.** The read walks `parent_id` root→leaf, and `previous_version_id` chains and `related` links are also self-referential — a cycle would loop forever. Reject any write/confirm that would create a cycle, and make the walk detect-and-fail (bounded) as a backstop. A human can confirm a cycle by mistake, so this cannot rely on confirmation alone.

---

## 4. Read patterns (what consumers actually do)

Concrete examples so the schema can be validated against real use.

- **Bundle builder, for a task touching `actions/`:**
  - `tenant_configs` for metadata.
  - `mechanical_config.body.layout.code_paths` to locate the actions directory.
  - `standards WHERE tenant_id = T AND status = active AND (scope = 'constitution' OR 'go_action' = ANY(applies_to))` — constitution + matched standards.
  - `objectives` walked root → area-node for the why-chain, plus the node's `direction_of_travel` and `priority_profile`.
- **User-rep advocate at a decision point in area A:**
  - `objectives` walk root → A. Reads `why`, effective `priority_profile`, `direction_of_travel`.
- **Mediator arbitrating in area A:**
  - Effective `priority_profile` for A.
  - `standards` matching the change's `applies_to` for the convention atoms in play.
- **Maintenance agent — drift sweep:**
  - All entries `WHERE freshness_until < now()` across tables.
  - `mechanical_config` re-inspection diffed against current `body`.
  - `standards WHERE source = 'doc_sourced'` whose `source_ref` documents have changed since `last_verified_at`.

---

## 5. Write patterns and confirm-not-initiate

Agents propose; the human confirms; layer agents apply.

- **Proposing** — a layer agent writes a row with `status = proposed` (or updates an existing row's `version` and `status = proposed`).
- **Pending queue** — a row in the (next-contract) work-items table references the proposed entry. The tenant interface reads work-items, surfaces the proposal, awaits confirmation.
- **Confirmation** — the work-item is closed; the layer agent (or a dedicated confirmer service) writes `status = active`, sets `last_verified_at` and `verified_by`. The previously-active version, if any, moves to `status = deprecated` with `supersedes` set on the new row.
- **No agent writes `active` directly** — this is the operational meaning of confirm-not-initiate. Status transitions are gated, and the apply is performed by a single central confirmer (see `PLAN_config_work_items_contract` §4), so there is exactly one code path to `active`.

**One live proposal per target.** There is at most one `proposed` row per target at a time. A new proposal for a target that already has one **replaces** it (does not add a second), and expiring/superseding the work-item also resolves (deprecates) its proposed row — otherwise proposed rows orphan and accumulate. The "at most one live thing per target" rule that the work-items queue has extends down to these layer rows.

**Consistent-snapshot reads.** A consumer that reads across `standards`, `objectives`, and `mechanical_config` (the bundle builder especially) must read them at one point-in-time / one transaction, so a confirmation landing mid-read can't produce a mix of pre- and post-change rows that never coherently existed.

**Bootstrapping.** The `tenant_configs` row is created directly by the onboarding orchestrator at initialisation (`onboarding_state = not_started`), before any work-item gate exists for that tenant. This is not a confirm-not-initiate violation: `tenant_configs` is the scope-holder, not a config value. Every config *value* (mechanical entry, standard, objective) still goes through the gate.

**"Confirmation by reality" is a graduated state, not a bypass.** The stack-discovery agent calls the mechanical layer "confirmed by reality" (probe success). Initially this still goes through the work-item gate — probe success is strong evidence that makes the confirmation near-rubber-stamp, but the human gates activation. Only once the stack-discovery capability graduates via the trust ledger (it has the strongest verifiability, so it is the natural first to graduate past `confirm_every`) does probe success auto-activate without a per-entry gate. The gate is the starting position; graduation relaxes it.

---

## 6. Schema versioning

`tenant_configs.schema_version` records the active schema. Migrations bring older configs forward. Reading an older config is supported by migrating its data on read (or up-front). Schema changes are themselves human-confirmed.

---

## 7. Open

- **Exact provenance enums** — the `source` set (`inspection`, `probe`, `inferred`, `supplied`, `doc_sourced`) may need additions as more agents come online. Versioned with the schema.
- **JSONB shapes inside `mechanical_config.body`, `priority_profile`, `direction_of_travel`** — defined here at the level of "what fields and what they mean," but the precise JSON schemas are implementation detail. Worth pinning when the bundle builder is being built, because that consumer reads them most heavily.
- **Multi-tenant isolation at the row-level vs schema-per-tenant** — for the service, row-level with strict `tenant_id` filtering is the lean; schema-per-tenant only if isolation requirements force it. Deferred to the service-phase decisions.

---

## 8. One-line state

The active config is the logical union of mechanical_config + standards + objectives, scoped per tenant via tenant_configs, with a common provenance shape across all three layers, status-gated transitions enforcing confirm-not-initiate, and effective values (especially the priority profile) computed at read time so ancestor changes propagate automatically. Storage is per-tenant rows in Postgres; consumers read the slices they need. This is the load-bearing contract; the other three (work-items, trust ledger / reasoning log, change-layer integration) hang off it.
