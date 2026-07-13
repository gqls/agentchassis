# FOCUS — Schema Verification Findings

**Status:** verification of the contract set's reuse claims against the **live schemas** in the project (`schemas_some`, `schemas_all`, `bk_agent_definitions_backup.sql`). Required by `FOCUS_contract_set_review` §2.5. Inspected `site_work_items` (which `config_work_items` claims to mirror) and `agent_definitions` (which the capabilities catalog claims to extend). The contracts are corrected to match reality, not the reverse.

The headline: the reuse claims were directionally right but wrong in detail, and `site_work_items` already contains machinery the work-items contract reinvented.

---

## 1. Chassis conventions — apply across ALL contracts

The live schema follows conventions the contracts did not match. These apply to every table in the set:

- **Enumerated values are `text` + a `CHECK` constraint, not native Postgres enums.** `agent_definitions.status` is `text` with `CHECK (status IN ('active','experimental','deprecated','demo','template'))`; `agent_category` likewise; `site_work_items.status` is plain `text`. → Every column the contracts call "enum" becomes `text` + `CHECK`.
- **Versioning is `version integer` + `previous_version_id uuid` (self-FK), with a unique `(type, version)`.** → the active-config schema's `supersedes` becomes `previous_version_id`.
- **Soft delete is `deleted_at timestamptz`, not a `status = archived`.** → drop `archived` from status vocabularies; use `deleted_at`.
- **Timestamps:** `timestamp with time zone`, default `now()`. (Contracts already match.)
- **JSONB** for structured/flexible payloads, default `'{}'` or `'[]'`. (Match.)
- **snake_case** columns. (Match.)

---

## 2. `site_work_items` — real shape vs the `config_work_items` contract

Real columns (abridged): `id, site_id (NOT NULL, FK→sites), source, item_type, severity, summary, spec (jsonb), page_id, component_id, entity_id, ..., priority (integer default 100), handler_agent, status (text default 'detected'), created_by, handled_by, approved_by, claimed_by, depends_on (uuid[]), parent_item_id, related_item_ids (uuid[]), batch_id, attempt_count, max_attempts, result (jsonb), error, item_key, created_at, triaged_at, claimed_at, completed_at, approval_mode (text default 'auto'), updated_at, pipeline (text default 'build')`.

Status vocabulary (from indexes): `detected, triaged, approved, claimed, complete, verified, rejected, wont_fix, failed, unresolved, blocked`.

**Findings — the contract was wrong on:**

| Contract assumed | Reality | Correction |
|---|---|---|
| `kind` | `item_type` | Rename to `item_type`. |
| `payload` | `spec` (+ `result` for outcome) | Rename to `spec`; add `result`. |
| `priority` enum (high/normal/low) | `priority` **integer** (default 100, lower = sooner) | Integer. |
| `status` enum (pending/in_progress/confirmed/...) | `status` **text**, richer lifecycle | Text; reuse the existing vocabulary (§2.1). |
| single `parent_work_item_id` | `depends_on` (uuid[]) **and** `parent_item_id` | Adopt both. |
| (none) | `attempt_count` / `max_attempts` | Adopt the retry machinery. |
| custom "supersede" rule | `item_key` + unique dedup index | Reuse dedup (§2.2). |
| **invented confirm-not-initiate gating** | **`approval_mode` already exists** | **Reuse it (§2.3).** |
| speculated `pipeline` discriminator | **`pipeline` already exists** | It's real; config = `pipeline='config'`. |

**Blocker for direct reuse:** `site_work_items.site_id` is `NOT NULL` with a FK to `sites`. Config work is **tenant-scoped, not site-scoped**, and the orchestrator/chassis is well-tested (resist changing it). → A **parallel `config_work_items` table** scoped by `tenant_id`, mirroring the real shape, is the pragmatic choice — not a modification of the live site table.

### 2.1 Reuse the existing status lifecycle

The config confirmation flow maps cleanly onto the existing vocabulary rather than needing a new one:

```
detected  (agent emitted the proposal)
  → triaged   (surfaced to the tenant)
  → approved  (human confirmed)        | rejected
  → claimed   (confirmer applying)
  → complete  (applied)                | failed (apply error)
```

`deferred`/`expired` map to `wont_fix`/`unresolved` or a small config-specific addition; this is the one place a new value may be warranted.

### 2.2 `item_key` dedup replaces the hand-rolled "supersede" rule

The review's "newer supersedes pending" rule (`FOCUS_contract_set_review` §2.3) is better served by the existing mechanism: a unique partial index on `(scope, item_key)` for non-terminal statuses already prevents two live items for the same target. Set `item_key` to the target identity (e.g., `standards:<id>` or `mechanical_config:build.command`); a second proposal for the same target collides and is reconciled, rather than two pending items coexisting. Less custom logic, reuses proven machinery.

### 2.3 `approval_mode` is the confirm-not-initiate mechanism — it already exists

`site_work_items.approval_mode` (text, default `'auto'`) already encodes whether a human gates the item. Build items default to `'auto'` (no gate); **config items default to a manual mode** (gate required), which is confirm-not-initiate expressed in the existing field. The "central confirmer" and "status-gated transition" from the review map onto this: the confirmer acts when `approval_mode` requires it and a human has set `status = approved`. No new gating machinery — reuse `approval_mode`.

---

## 3. `agent_definitions` — relevant to the capabilities catalog

Real columns (relevant): `id, type, display_name, description, category, default_config (jsonb), is_active, capabilities (jsonb default '[]'), ..., version (integer), previous_version_id (uuid, self-FK), task_workflow / orchestrator_workflow / orchestration_workflow, delegation_preferences, agent_category (text CHECK strategist/executor/analyst/integrator/coordinator/specialist), status (text CHECK active/experimental/deprecated/demo/template), domain_tags (jsonb), briefing_questionnaire (jsonb), usage_count, is_snapshot, input_contract (jsonb), output_contract (jsonb), deleted_at`.

**Findings:**

- **`capabilities` jsonb already exists** on `agent_definitions` — a per-agent list of what the agent can do. The trust catalog is therefore **not** an extension of this column; it is a **sibling table** holding per-capability trust attributes (ceiling, verifiability, containment), referenced by the agents' `capabilities` list. Reason: capabilities aren't 1:1 with agents (some span several), and the ceiling is a property of the capability, not the agent.
- **`input_contract` / `output_contract` (jsonb) exist** — the agent contracts the specs referenced are real fields.
- **`briefing_questionnaire` (jsonb) exists** on `agent_definitions` — confirms the intent-elicitation agent's reuse of the briefing questionnaire is real and already lives in the schema.
- **`agent_category` CHECK set** (strategist/executor/analyst/integrator/coordinator/specialist) is the existing agent taxonomy — onboarding agents should declare one of these.
- Versioning (`version` + `previous_version_id`), `deleted_at`, and `status` CHECK confirm the §1 conventions.

---

## 4. Verdict on the reuse claims

- **"`config_work_items` mirrors `site_work_items`"** — true in spirit, wrong in detail. Corrected column names, integer priority, text status with the real vocabulary, and — most important — reuse of the existing `approval_mode` (gating), `pipeline` (discriminator), `depends_on`, retry, and `item_key` dedup that the contract had reinvented or omitted. Parallel table scoped by `tenant_id` (site_id NOT NULL blocks direct reuse).
- **"capabilities catalog extends `agent_definitions`"** — partly. The `capabilities` jsonb exists but the trust catalog is a sibling table referenced by it. Corrected.
- **Cross-cutting:** contracts used native enums, `supersedes`, and `status=archived`; chassis uses text+CHECK, `previous_version_id`, and `deleted_at`. Corrected across the set (§1).

---

## 5. Corrections: applied vs pending

- **`config_work_items` contract** — revised this pass (§2 findings applied).
- **Capabilities catalog** — written this pass against §3 (sibling table, chassis conventions).
- **`active_config_schema`** — convention note added this pass; a full pass to swap enum→text+CHECK, `supersedes`→`previous_version_id`, and `archived`→`deleted_at` still to apply.
- **`decision_log`, `trust_ledger`, `change_layer`** — same convention pass pending (enum→text+CHECK etc.). Mechanically simple; flagged here so it isn't forgotten.

---

## 6. One-line state

The live schemas confirm the reuse direction but correct the detail: `config_work_items` should mirror `site_work_items` properly (item_type/spec/result, integer priority, text status with the existing lifecycle, and reuse of `approval_mode`, `pipeline`, `depends_on`, retry, and `item_key` dedup) as a parallel tenant-scoped table; the capabilities catalog is a sibling of `agent_definitions`, not an extension; and across all contracts the chassis convention is text+CHECK (not enums), `version`+`previous_version_id` (not supersedes), and `deleted_at` (not status=archived).
