# PLAN — Config Work-Items Contract

**Status:** contract specification. The shared queue every onboarding and maintenance agent emits into; the tenant interface reads from; agents react to resolutions. Settles the §3.3 gap from `FOCUS_onboarding_system_view_check`. Mirrors the existing `site_work_items` pattern from the chassis — reuse, not new infrastructure.

---

## 1. Purpose

Every agent that proposes a change to the active config (`PLAN_active_config_schema`) emits a work-item. The tenant has one place to look at pending decisions across onboarding, drift, and re-onboarding. Confirmations close work-items; the originating agent reads the resolution and applies the change (or doesn't, on rejection).

This is the operational implementation of confirm-not-initiate: agents do not write `status = active` directly to layer tables — they propose, the work-item carries the proposal to the tenant, and the resolution drives the status transition.

---

## 2. The table — `config_work_items`

Parallel to `site_work_items` (same shape pattern). Could be merged with a `pipeline = 'config'` value in a unified work-items table later if useful; kept parallel for now to keep config and site workflows from mixing.

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | Identifier. |
| `tenant_id` | uuid FK | Scope. |
| `kind` | enum | `confirm_proposal` / `resolve_drift` / `complete_interview_step` / `confirm_disposition` (for drift-audit findings) / `confirm_re_onboarding` |
| `source_agent` | enum | Which agent emitted it: `stack_discovery` / `conventions` / `intent_elicitation` / `orchestrator` / `maintenance` |
| `target_table` | enum, nullable | `standards` / `objectives` / `mechanical_config` / null if generic. |
| `target_ref` | jsonb | **Polymorphic** reference to the proposed entry. Atom tables: `{table:'standards'\|'objectives', id, version}`. Mechanical layer (one JSONB row, no atom id): `{table:'mechanical_config', tenant_id, json_path}` (e.g. `json_path:'build.command'`). |
| `payload` | jsonb | Proposal content + evidence + suggested disposition. Agent-specific shape inside. |
| `summary` | text | Tenant-facing one-line summary — what to confirm, why. |
| `status` | enum | `pending` / `in_progress` / `confirmed` / `rejected` / `deferred` / `expired` |
| `priority` | enum | `high` / `normal` / `low`. Drives surfacing order. |
| `parent_work_item_id` | uuid FK, nullable | Chains (re-onboarding fans out many items; drift on a sealed constraint may chain to descendant re-validations). |
| `created_at` | timestamptz | |
| `expires_at` | timestamptz, nullable | For time-bounded items (e.g., a direction-of-travel re-confirmation that auto-defers if not acted on). |
| `resolved_at` | timestamptz, nullable | |
| `resolved_by` | text, nullable | User id. |
| `resolution` | enum, nullable | `accept` / `reject` / `modify` / `defer` |
| `resolution_payload` | jsonb, nullable | Modified content if the tenant edited the proposal before accepting. |
| `resolution_premise` | text, nullable | The human's stated reason for the decision — feeds the decision log (next contract). |

Indexes: `(tenant_id, status, priority, created_at)` for the tenant interface; `(target_table, (target_ref->>'id'))` for agents watching atom-table proposals, plus `(target_table, (target_ref->>'json_path'))` for mechanical-layer proposals.

---

## 3. Lifecycle

```
[agent emits] → pending → in_progress → { confirmed | rejected } → [confirmer applies]
                                       → deferred → [returns to pending later]
                                       → expired  → [confirmer re-emits or drops]
```

- **`pending`** — emitted, awaiting tenant.
- **`in_progress`** — tenant is actively reviewing (UI state).
- **`confirmed`** — accepted. The `resolution` field distinguishes `accept` (as proposed) from `modify` (accepted with edits in `resolution_payload`); both apply, the latter applies the edited version.
- **`rejected`** — declined; the change is not applied. Repeated rejection of similar items is itself a drift signal the maintenance agent picks up.
- **`deferred`** — explicitly postponed; returns to `pending` after a timeout or trigger.
- **`expired`** — `expires_at` reached without resolution, or superseded by a newer proposal on the same target (§4). The confirmer decides whether to re-emit or drop.

(`modified` is **not** a status — a modification is `status = confirmed` with `resolution = modify`. Status is the lifecycle; `resolution` is the outcome.)

---

## 4. Relationship to the layer tables

The work-items table is the gate between proposals and active config:

1. A layer agent writes a row in `standards` / `objectives` / `mechanical_config` with `status = proposed`.
2. The agent emits a `config_work_items` row referencing that proposed row via `target_table` + `target_ref`.
3. The tenant interface reads pending work-items (priority + age order), surfaces them, captures the tenant's resolution.
4. On `confirmed` (whether `resolution = accept` or `modify`), a **single central confirmer** reads the resolution and applies it: flips the layer row's `status` from `proposed` to `active` (using `resolution_payload` if `modify`), sets `last_verified_at` and `verified_by`, deprecates any prior active version with `supersedes` set on the new row, writes the `decision_log` entry carrying the human's `resolution_premise`, and emits the `in_band` change event.
5. The work-item itself stays as historical record (status preserved; do not delete).

**One confirmer, one path to `active`.** The apply logic lives in a single component, not reimplemented per agent — so confirm-not-initiate is enforced in one auditable place and there is exactly one route a layer row takes to `active`. The confirmer's apply must be **idempotent**: applying a version already `active` is a no-op, so a retry after partial failure cannot double-apply.

**Concurrent proposals on the same target.** If a new proposal arrives for a target that already has an unresolved work-item (e.g. a Makefile commit makes maintenance propose a new build command while an earlier proposal is still pending), the **newer proposal supersedes the older** — the older work-item is marked `expired` with a note, and the newer carries forward. This prevents the tenant from confirming a proposal already overtaken by reality. (Blocking the new proposal until the old resolves is rejected — it risks confirming something stale.)

This is the only path a layer row gets to `status = active`. The schema enforces confirm-not-initiate by making this gate the single transition route.

---

## 5. Surfacing order — priority and age

The tenant interface orders pending items by `priority` first, then by `created_at` (older first within a priority). The maintenance agent (§5.6 of the agent specs) sets priorities so high-impact drift surfaces before low-impact freshness nudges, avoiding alert fatigue.

`expires_at` lets agents emit time-bounded items (a freshness-expiring direction-of-travel note gets a short expiry; an architectural drift candidate may have none).

---

## 6. Reuse

This is **intended to mirror the existing `site_work_items` shape** — the same lifecycle pattern and chain structure — so it is not new infrastructure. **This is an assumption to verify:** the actual `site_work_items` schema must be inspected before DDL (schema-before-SQL), and this contract corrected to match the real shape where it differs, not the reverse. The two tables stay separate for now (config workflows differ from site workflows enough that mixing would muddle logs); they could converge under a unified `work_items` table with a `pipeline` field later if useful.

---

## 7. Open

- **Modification flow.** When a tenant chooses `modify`, how rich an edit is allowed before it's effectively a new proposal? Probably: small edits via `resolution_payload`; large edits cause the original proposal to be `rejected` and a new one to be drafted by the agent based on the tenant's correction. The boundary is fuzzy and worth pinning when the UI is built.
- **Chained items.** A re-onboarding triggers many items; a sealed-constraint change triggers descendant re-validations. The chain needs surfacing semantics — does resolving the parent auto-defer the children, are children blocked until the parent resolves, can the tenant resolve in any order? Defaults: children are linked but independently actionable; resolving the parent unblocks any blocked-on-parent children.
- **Whether to merge with `site_work_items`** under a unified `work_items` table with `pipeline` discriminator — deferred until both surfaces are mature enough to compare.

---

## 8. One-line state

Every config change proposal flows through `config_work_items`: agent emits with `status = pending` and `target_ref` pointing at a `status = proposed` row in the layer table; tenant resolves with confirm/reject/modify/defer plus a `resolution_premise`; on confirmation the source agent flips the layer row to `active`. The table mirrors the existing `site_work_items` pattern; confirm-not-initiate becomes a status-transition rule enforced by the gate, not just a process discipline.
