# PLAN — Trust Ledger Contract

**Status:** contract specification. The artifact named throughout `MASTER_autonomous_build_and_operate` as "the master knob" that governs both how conservatively a capability is produced (cascade floor) and whether its output applies without a human (gate). Referenced by maintenance (§5.7 of the onboarding agent specs) and central to the bidirectional ratchet. Now given concrete shape.

---

## 1. Purpose

For each capability the system can perform — write a Go action, provision an nginx, run a rollback, propose a schema migration — the trust ledger records the **current trust level** at which it operates. Cascade routers read it to set the production tier; gate policy engines read it to decide whether output applies autonomously or needs human confirmation. Mutations are gated by confirm-not-initiate and driven by drift signal from the maintenance agent.

---

## 2. State, not history

The trust ledger is **mutable state** — the current trust level per capability per tenant. It is **derived from** the immutable `decision_log` (the source of truth for what happened) but it is what's *used* at runtime. Two artifacts because they have different access patterns:

| Artifact | Property | Use |
|---|---|---|
| `decision_log` | Immutable, append-only | Audit, evidence, reconstruction |
| `trust_ledger` | Mutable state | The active "what trust level is this capability at right now" |

Every mutation to the ledger has a corresponding `decision_log` entry showing what evidence justified the change.

---

## 3. The table — `trust_ledger`

One row per (tenant, capability).

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | |
| `tenant_id` | uuid FK | Scope. Trust is built **per-tenant** by their use — a capability may be `autonomous` for tenant A and `confirm_every` for tenant B. |
| `capability_id` | text | Stable identifier (`write_go_action`, `provision_nginx`, `rollback_to_known_good`, `propose_schema_migration`, etc.) |
| `trust_level` | text + CHECK | `confirm_every` / `confirm_exceptions` / `notify` / `autonomous` |
| `gate_policy` | jsonb | Derived gate-policy details (the cheap-floor / expensive-escalation rules from `MASTER` §7 instantiated for this capability + level). |
| `evidence_summary` | jsonb | Compact stats: recent success rate, drift count, last drift event id, sustained-no-drift days, last verification outcome, etc. — populated by the maintenance agent. |
| `last_changed_at` | timestamptz | When the trust level last changed. |
| `last_change_direction` | text + CHECK, nullable | `up` (graduation) / `down` (de-graduation) / `manual_set` |
| `last_change_decision_id` | uuid FK, nullable | `decision_log` entry that justified the change. |
| `created_at` | timestamptz | |
| `updated_at` | timestamptz | |
| `deleted_at` | timestamptz, nullable | Soft delete (chassis convention) — e.g. when a capability is retired for a tenant. |

Indexes: `(tenant_id, capability_id) WHERE deleted_at IS NULL` unique; `(tenant_id, trust_level)` for "which capabilities are at level X."

*(Convention: enumerated columns are `text` + `CHECK`, not native enums, per `FOCUS_schema_verification_findings` §1.)*

---

## 4. The ceiling lives on the capability, not the ledger

The **ceiling** — the maximum trust level a capability can reach, determined by verifiability + containment per `MASTER` §6 — is a property of the **capability itself, not the tenant**. A capability whose factors limit it to `notify` cannot be `autonomous` for anyone, regardless of evidence.

So the ceiling lives in the **capabilities catalog** (a sibling table — see `PLAN_capabilities_catalog_contract`; decided as a sibling, not an extension of `agent_definitions`, because the existing `capabilities` column holds free descriptive tags, not trust-units), not on the ledger row:

```
capabilities:
  capability_id (PK)
  display_name
  ceiling                       enum
  verifiability                 enum  (strong/medium/weak)
  containment                   enum  (strong/medium/weak)
  cascade_tier_hints            jsonb
  ...
```

The ledger references it by `capability_id`. Mutations to the ceiling — e.g., a new validator raises a capability's verifiability and therefore its ceiling — happen on the capability row, not on the ledger; existing ledger entries respect the new ceiling on next read.

---

## 5. Mutation flow (graduation, de-graduation)

**Cold start.** A new `(tenant, capability)` pair has no ledger row. The first time a capability is invoked for a tenant, the row is created at **`confirm_every`** — the most conservative level, bounded by the capability's ceiling — never at an inherited or higher default. Trust is earned per tenant from the most cautious starting point.

Every change to `trust_level` is gated.

1. **Maintenance agent (or a dedicated ratchet-evaluator) proposes** a change based on accumulated evidence in `evidence_summary` and recent `decision_log` entries:
   - Sustained no-drift + verification success → graduate up.
   - Repeated drift / repeated correction → de-graduate down.
2. **Proposal becomes a `config_work_items` row** with `kind = confirm_proposal`, `target_table = trust_ledger`, payload containing the proposed new level + the evidence.
3. **Human confirms** via the work-items interface, with `resolution_premise`.
4. **On confirmation**, the ledger row is updated: new `trust_level`, `last_changed_at`, `last_change_direction`, and `last_change_decision_id` pointing to a freshly-written `decision_log` entry that carries the resolution premise and the evidence summary at the time.
5. **De-graduations may be auto-applied** when evidence is severe (e.g., a verification failure rate above a threshold), with the work-item raised for *notification* rather than confirmation. This is the bidirectional ratchet's safety property — losing trust shouldn't wait on a human; gaining trust should. Specifically:
   - Graduation up: always confirm-not-initiate.
   - De-graduation down: may auto-apply with notification, especially on severe regression.

This asymmetry matters: the ratchet is bidirectional, but the two directions aren't symmetrically governed. Tightening is safer than loosening — losing trust is reversible; falsely gaining trust is what allows mistakes to apply unsupervised.

---

## 6. Read patterns

- **Cascade router** (per `MASTER` §3): for a task touching capability C in tenant T, read `trust_ledger WHERE tenant_id = T AND capability_id = C` → use the `trust_level` to floor the cascade tier (low trust forces tier 1 reuse; high trust permits tier 2 generate+verify or beyond).
- **Gate policy engine**: same read → `trust_level` decides `confirm_every | confirm_exceptions | notify | autonomous` at the gate.
- **Maintenance agent's evaluation loop**: reads its own `evidence_summary` + recent `decision_log` entries to decide whether to propose a graduation/de-graduation.
- **Tenant-facing dashboard**: read all entries for a tenant, joined with the capability catalog, to display "where each capability sits and why."

---

## 7. Relationship to other contracts

- **`decision_log`** is the source of evidence; the ledger is its summary state. Every ledger mutation logs a corresponding decision-log entry. Reading old ledger states is done via decision-log replay, not by storing old ledger rows.
- **`config_work_items`** is the gate. Graduations flow through it (confirm-not-initiate). Severe de-graduations may bypass confirmation but still emit a work-item for notification.
- **Capability catalog (`capabilities` or extended `agent_definitions`)** holds the ceiling and the static verifiability/containment factors.

---

## 8. Open

- **Capability catalog placement** — extend `agent_definitions` (since most capabilities map to agents) or a sibling `capabilities` table. Lean: sibling table, because some capabilities aren't single agents (e.g., "propose a schema migration" involves several).
- **Auto-apply thresholds for de-graduation.** What evidence severity skips the confirm gate? Likely tenant-configurable, with safe defaults.
- **Evidence summary refresh cadence** — recomputed every drift event, or on a schedule, or both. Probably both: drift triggers immediate update; periodic sweep catches what triggers missed.
- **Trust transferability across tenants.** If a capability is autonomous for 100 tenants without incident, does that affect the *starting* trust level for tenant 101? Lean no for safety (each tenant builds their own trust), but cohort-derived ceilings or warning signals could be useful. Deferred.

---

## 9. One-line state

The trust ledger is per-tenant mutable state holding `trust_level` and `gate_policy` per capability, derived from but separate from the immutable `decision_log`, with the ceiling living on the capability catalog (not the ledger), graduations gated by confirm-not-initiate and de-graduations allowed to auto-apply with notification — the bidirectional ratchet's asymmetric safety property, made operational.
