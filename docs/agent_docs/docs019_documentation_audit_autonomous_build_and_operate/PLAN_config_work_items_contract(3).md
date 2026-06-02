# PLAN — Config Work-Items Contract

**Status:** contract specification. The shared queue every onboarding and maintenance agent emits into; the tenant interface reads from; agents react to resolutions. Settles the §3.3 gap from `FOCUS_onboarding_system_view_check`. Mirrors the existing `site_work_items` pattern from the chassis — reuse, not new infrastructure.

---

## 1. Purpose

Every agent that proposes a change to the active config (`PLAN_active_config_schema`) emits a work-item. The tenant has one place to look at pending decisions across onboarding, drift, and re-onboarding. Confirmations close work-items; the originating agent reads the resolution and applies the change (or doesn't, on rejection).

This is the operational implementation of confirm-not-initiate: agents do not write `status = active` directly to layer tables — they propose, the work-item carries the proposal to the tenant, and the resolution drives the status transition.

---

## 2. The table — `config_work_items`

A **parallel table** mirroring the real `site_work_items` shape (verified in `FOCUS_schema_verification_findings` §2). Not a modification of `site_work_items` — that table's `site_id` is `NOT NULL` with a FK to `sites`, and config work is tenant-scoped, not site-scoped; the live site table is well-tested and not to be disturbed. Column names, the integer `priority`, the text `status` lifecycle, and the `approval_mode` / `pipeline` / `depends_on` / `item_key` / retry machinery are taken from the real table rather than invented.

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | `gen_random_uuid()`. |
| `tenant_id` | uuid FK | Scope (the config analog of `site_work_items.site_id`). |
| `pipeline` | text | `'config'`. The discriminator already exists on `site_work_items`; kept so the two could converge later. |
| `source` | text | Origin of the item (which trigger / sweep / onboarding step). |
| `item_type` | text | `confirm_proposal` / `resolve_drift` / `complete_interview_step` / `confirm_disposition` / `confirm_re_onboarding`. (Was `kind` — corrected to the real column name.) |
| `severity` | text | `low` / `medium` / `high`. Default `'medium'`. |
| `summary` | text | Tenant-facing one-line summary. |
| `spec` | jsonb | Proposal content + evidence + suggested disposition. (Was `payload` — corrected.) Default `'{}'`. |
| `target_table` | text, nullable | `standards` / `objectives` / `mechanical_config` / null. |
| `target_ref` | jsonb | **Polymorphic** reference. Atom tables: `{table, id, version}`. Mechanical layer: `{table:'mechanical_config', tenant_id, json_path}`. |
| `priority` | integer | Default 100; lower = surfaced sooner. (Was an enum — corrected to integer.) |
| `handler_agent` | text, nullable | The agent that will apply on confirmation (the central confirmer). |
| `status` | text | Reuses the real lifecycle vocabulary — see §3. Default `'detected'`. |
| `approval_mode` | text | **The confirm-not-initiate field, already in `site_work_items`.** Config items default to a manual mode (gate required); `'auto'` only for graduated capabilities. |
| `created_by` | text | The emitting agent. |
| `approved_by` | text, nullable | The confirming human. |
| `claimed_by` | text, nullable | The confirmer applying it. |
| `depends_on` | uuid[], nullable | Dependencies (real column). |
| `parent_item_id` | uuid, nullable | Hierarchy / chains (real column name — was `parent_work_item_id`). |
| `attempt_count` / `max_attempts` | integer | Retry machinery (real columns). Defaults 0 / 3. |
| `result` | jsonb | Outcome, incl. the human's `resolution_premise` (feeds the decision log). Default `'{}'`. |
| `error` | text, nullable | Apply error if failed. |
| `item_key` | text, nullable | Target identity for dedup (e.g. `standards:<id>`) — see §4. |
| `created_at` / `updated_at` | timestamptz | Default `now()`. |
| `triaged_at` / `claimed_at` / `completed_at` | timestamptz, nullable | Lifecycle timestamps. |

Indexes mirror `site_work_items`: unique partial on `(tenant_id, item_key)` for non-terminal statuses (the dedup, §4); `(tenant_id, status)`; `(tenant_id, priority) WHERE status IN ('detected','triaged')`; GIN on `depends_on`.

---

## 3. Lifecycle (reusing the real `site_work_items` vocabulary)

The config confirmation flow maps onto the existing status vocabulary rather than a new one:

```
detected  (agent emitted the proposal)
  → triaged   (surfaced to the tenant)
  → approved  (human confirmed)     | rejected (declined)
  → claimed   (confirmer applying)
  → complete  (applied)             | failed   (apply error → retry up to max_attempts)
```

- **`detected`** — emitted by an agent; a `proposed` row exists in the layer table.
- **`triaged`** — surfaced to the tenant in the interface.
- **`approved`** — human confirmed. `approval_mode` decides whether this requires a human (config default) or is automatic (`'auto'`, for graduated capabilities only). The `result` carries the human's `resolution_premise`, and whether the approval was as-proposed or with edits (edited content also in `result`/`spec`).
- **`rejected`** — declined; not applied. Repeated rejection of similar items is a drift signal maintenance picks up.
- **`claimed` → `complete`** — the central confirmer claims the approved item, applies it (flips the layer row to `active`, etc.), and marks it complete. `failed` on apply error, retried up to `max_attempts`.
- **`unresolved`** — the config analog of deferred/expired (reuses an existing terminal value rather than adding one).

(There is no separate `modified` status — an edited approval is `status = approved` with the edited content carried in `result`/`spec`. Status is the lifecycle; the edit is data.)

---

## 4. Relationship to the layer tables

The work-items table is the gate between proposals and active config:

1. A layer agent writes a row in `standards` / `objectives` / `mechanical_config` with `status = proposed`.
2. The agent emits a `config_work_items` row (`status = detected`) referencing that proposed row via `target_table` + `target_ref`, with `item_key` set to the target identity.
3. The tenant interface surfaces triaged items (priority then age) and captures the tenant's decision (`approved` / `rejected`), with the premise written to `result`.
4. On `approved`, a **single central confirmer** claims the item (`claimed`), applies it — flips the layer row's `status` from `proposed` to `active` (using the edited content if the approval carried edits), sets `last_verified_at` and `verified_by`, deprecates any prior active version with `previous_version_id` set on the new row, writes the `decision_log` entry carrying the premise, emits the `in_band` change event — and marks the item `complete` (`failed` on error, retried up to `max_attempts`).
5. The work-item stays as historical record (status preserved; soft-delete via timestamp if ever needed, never hard-deleted).

**The apply must be all-or-nothing.** The flip-to-active, the deprecation of the prior version, the decision-log write, and the work-item completion go in **one database transaction**; the `in_band` change event is written to an **outbox in that same transaction** and published by a relay. So a crash mid-apply leaves a consistent state — never a live row with no log entry, never an applied change with no event for maintenance to see — and a retry completes cleanly.

**One confirmer, one path to `active`.** The apply logic lives in a single component, not reimplemented per agent — confirm-not-initiate enforced in one auditable place, one route a layer row takes to `active`. The apply is **idempotent**: applying a version already `active` is a no-op, so a retry cannot double-apply.

**Concurrent proposals on the same target — one live proposal per target.** The unique partial index on `(tenant_id, item_key)` over non-terminal statuses means a second proposal for the same target (same `item_key`) cannot create a coexisting live work-item — it collides. Crucially the same "one live per target" rule applies **down at the layer table**: the newer proposal **replaces** the existing `proposed` row rather than adding a second, and when a work-item is expired/superseded its `proposed` layer row is resolved (deprecated) — otherwise proposed rows orphan. (This corrects the earlier framing that relied on the queue dedup alone; the dedup has to reach the layer rows too.)

This is the only path a layer row gets to `status = active`. Confirm-not-initiate is a status-transition rule plus `approval_mode`, enforced by the gate, not a process discipline.

---

## 5. Surfacing order — priority and age

The tenant interface orders triaged items by `priority` (integer; lower sooner) then `created_at`. The maintenance agent (§5.6 of the agent specs) sets priorities so high-impact drift surfaces before low-impact freshness nudges, avoiding alert fatigue. Time-bounded items (a freshness-expiring direction-of-travel note) carry a short expiry handled via a sweep that moves them to `unresolved`; long-lived items have none.

**Batch confirmation for the initial flood.** Cold-start trust is `confirm_every`, and the first onboarding of a repo produces a large batch of proposals at once (every convention, every objective node, the drift findings). Per-item confirmation there is a wall. So the interface supports **confirming a whole reviewed batch together** for the initial onboarding, while keeping per-item confirmation for the steady trickle of later drift. The gate is unchanged — what adapts is the granularity of the confirmation, not whether there is one.

**Scope: this gates config, not deliverables.** `config_work_items` gates changes to the tool's own knowledge of the codebase (standards, objectives, mechanical facts). The tool's actual outputs — generated code, and edits to workflows/agent definitions — are a **different gated path** (the capability's trust level + verification + the cascade), not this queue. Editing a workflow is a deliverable, not a config change; don't route it through the confirmer.

---

## 6. Reuse — verified against the live schema

The real `site_work_items` schema has been inspected (`FOCUS_schema_verification_findings` §2). This table mirrors it properly: `item_type`, `spec`/`result`, integer `priority`, the text `status` lifecycle (`detected → triaged → approved → claimed → complete`), and — importantly — it **reuses the existing `approval_mode` (the confirm-not-initiate field), `pipeline`, `depends_on`, `attempt_count`/`max_attempts`, and `item_key` dedup** rather than inventing equivalents. It is a parallel table (not a modification of `site_work_items`, whose `site_id` is `NOT NULL` with a FK to `sites`); the two could converge under a unified `work_items` table with `pipeline` discriminating, deferred until both surfaces are mature.

---

## 7. Open

- **Edit-on-approval richness.** How large an edit on approval before it's effectively a new proposal? Lean: small edits ride in `result`/`spec`; large edits → `rejected` + a fresh agent-drafted proposal. Pin when the UI is built.
- **Chained items.** Re-onboarding fans out many items; a sealed-constraint change chains to descendant re-validations via `depends_on` / `parent_item_id`. Default: children linked but independently actionable; resolving the parent unblocks blocked-on-parent children.
- **`unresolved` semantics.** Reusing `unresolved` for defer/expire is pragmatic but may want a dedicated config value if the two cases need distinguishing in the UI.
- **Convergence with `site_work_items`** under one table — deferred.

---

## 8. One-line state

Every config change proposal flows through `config_work_items`, a parallel table mirroring the verified `site_work_items` shape: the emitting agent writes `status = detected` with `target_ref` at a `proposed` layer row and `item_key` for dedup; the tenant approves or rejects (with premise in `result`); a single central confirmer claims and applies, flipping the layer row to `active`. Confirm-not-initiate is the existing `approval_mode` field plus the status-gated transition, enforced by the gate — not new machinery and not just a discipline.
