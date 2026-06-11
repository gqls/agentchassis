# PLAN — Capabilities Catalog Contract

**Status:** contract specification. The sixth contract, closing the trust ledger's open dependency (`PLAN_trust_ledger_contract` §4, §8 — the ceiling and the verifiability/containment factors live here, not on the ledger). Built against the verified `agent_definitions` schema (`FOCUS_schema_verification_findings` §3) and following chassis conventions (text + CHECK, `version` + `previous_version_id`, `deleted_at`).

---

## 1. Purpose

The trust ledger holds **per-tenant** trust *level* for each capability. The **ceiling** — the maximum level a capability can ever reach — and the static factors that determine it (verifiability, containment) are properties of the **capability itself, not the tenant**. They need a home that is not the ledger and not duplicated per agent. That home is the capabilities catalog.

---

## 2. Why a sibling table, not an extension of `agent_definitions`

`agent_definitions` already has a `capabilities` jsonb column — a per-agent **list** of what each agent can do. The catalog is a different thing:

- **Capabilities aren't 1:1 with agents.** "Propose a schema migration" or "roll back to a known-good state" may involve several agents; a capability is a unit of *trust*, not a unit of *agent*.
- **The ceiling is a property of the capability**, set by its verifiability and containment, regardless of which agent exercises it.

So the catalog is a **sibling table** (`capabilities`), and the agents' existing `capabilities` jsonb **references catalog `capability_id`s**. The agent lists the capabilities it exercises; the catalog holds each capability's trust attributes. No duplication, and capabilities that span agents are represented once.

---

## 3. The table — `capabilities`

Follows chassis conventions throughout.

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | `gen_random_uuid()`. |
| `capability_id` | text | Stable identifier (`write_go_action`, `provision_nginx`, `rollback_to_known_good`, `propose_schema_migration`). Unique with `version`. |
| `display_name` | text | Human label. |
| `description` | text, nullable | What the capability does. |
| `ceiling` | text + CHECK | Max trust level: `confirm_every` / `confirm_exceptions` / `notify` / `autonomous`. The ledger respects this. |
| `verifiability` | text + CHECK | `strong` / `medium` / `weak` — how well the output can be ground-truth checked (per `MASTER` §6). |
| `containment` | text + CHECK | `strong` / `medium` / `weak` — blast radius / reversibility. |
| `ceiling_rationale` | text, nullable | Why the ceiling is what it is — the reasoning, since the ceiling is a judgement over the two factors. |
| `cascade_tier_hints` | jsonb | Hints for the cascade router (e.g., "reuse-search corpus", "competing-solutions count"). Default `'{}'`. |
| `category` | text + CHECK, nullable | Reuse the `agent_category` taxonomy where it applies: `strategist`/`executor`/`analyst`/`integrator`/`coordinator`/`specialist`. |
| `status` | text + CHECK | `active` / `experimental` / `deprecated` / `demo` / `template` — same vocabulary as `agent_definitions.status`. |
| `version` | integer | Default 1. |
| `previous_version_id` | uuid, nullable | Self-FK, lineage (chassis convention). |
| `usage_count` | integer | Default 0 — how often exercised (feeds evidence). |
| `created_at`, `updated_at` | timestamptz | Default `now()`. |
| `deleted_at` | timestamptz, nullable | Soft delete (chassis convention). |

Constraints/indexes: unique `(capability_id, version)`; index `(status) WHERE deleted_at IS NULL`; index `(ceiling)`.

---

## 4. The ceiling is computed from the factors (compute-on-read where it matters)

`ceiling` is a judgement over `verifiability` and `containment` (the weaker factor holds the ceiling, per `MASTER` §6). Two options, mirroring the priority-profile decision:

- **Store the ceiling** (a column, as above) — simple to read, but goes stale if the factors change (a new validator raises verifiability).
- **Compute the ceiling** from the factors on read — always fresh.

**Decision:** store the ceiling as the active value **but** treat the factors as authoritative — when `verifiability` or `containment` changes, a maintenance step recomputes and re-proposes the ceiling (confirm-not-initiate), with `ceiling_rationale` recording the judgement. This keeps reads cheap (the ledger reads one column) while keeping the factors the source of truth. The recompute is itself a config-work-item, so a factor change surfaces a proposed ceiling change for confirmation rather than silently moving it.

This matches the broader pattern: the *active* value is stored for cheap reads; a *change to its inputs* triggers a gated re-proposal (the targeted-re-validation pattern).

---

## 5. Relationship to other contracts

- **Trust ledger** references `capabilities.capability_id`. The ledger row carries the per-tenant `trust_level`; the catalog carries the `ceiling` it must respect. A ledger graduation cannot propose a level above the catalog ceiling.
- **`agent_definitions.capabilities` jsonb** references catalog `capability_id`s — the link from agents to the capabilities they exercise.
- **Cascade router** reads `cascade_tier_hints` alongside the ledger's `trust_level` to set the production tier.
- **Maintenance agent** writes `usage_count` and feeds the evidence that drives ledger graduation proposals; a factor change here triggers the §4 ceiling re-proposal.
- **`config_work_items`** carries both ceiling re-proposals (factor changed) and the agent↔capability linkage confirmations.

---

## 6. Seeding

The initial catalog is seeded from the capabilities the onboarding and tool agents exercise — write Go action, propose schema migration, assemble bundle, run verification, roll back, etc. — each given a ceiling from an initial verifiability/containment judgement (with rationale). This is a one-time authored seed, confirmed like any other config, and grows as new capabilities appear. Seeding is itself gated (the ceilings are judgements, so confirm-not-initiate applies).

---

## 7. Open

- **Capability granularity.** How fine-grained is a capability? "Write Go action" vs "write Go action in package X" vs "write any code." Too coarse and trust is blunt; too fine and the catalog explodes. Lean: capability granularity matches the unit at which verifiability/containment actually differ — if two things have the same factors, they're one capability.
- **Cross-tenant ceiling learning.** If a capability proves safe across many tenants, does that inform its ceiling? The ceiling is structural (factor-derived), so cohort evidence would adjust the *factors* (e.g., raise verifiability confidence), not bypass them. Deferred, consistent with the trust ledger's "no automatic cross-tenant trust" stance.
- **Relationship to `agent_definitions.capabilities` content.** Whether that jsonb currently holds free strings or structured refs determines the migration to make it reference catalog ids. Inspect before wiring (schema-before-SQL).

---

## 8. One-line state

The capabilities catalog is a sibling table of `agent_definitions` (not an extension of its `capabilities` jsonb), holding per-capability `ceiling` + `verifiability` + `containment` — properties of the capability, not the tenant — which the per-tenant trust ledger references and must respect. The ceiling is stored for cheap reads but the factors are authoritative: a factor change triggers a gated ceiling re-proposal. Built on chassis conventions (text+CHECK, version+previous_version_id, deleted_at), closing the trust ledger's open dependency.
