# PLAN — Change-Layer Integration Contract

**Status:** contract specification. The mechanism by which diffs reach the maintenance agent. Referenced in §5.4 of the onboarding agent specs and §3.6 of the system-view check, now given concrete shape. Closes the final contract gap before implementation.

---

## 1. Purpose

The maintenance agent is event-driven on changes plus periodic on a sweep. This contract defines:

- **What a change event looks like** (the shape every source emits in).
- **Where events come from** (the mechanism options, plus the **in-band** case for the tool's own changes).
- **How events become maintenance triggers** (the filter, computed from the mechanical config).
- **Storage as first-class events** so the change layer itself is auditable.

---

## 2. The table — `change_events`

First-class storage so events are reproducible, retryable, and auditable. Not transient. (An append-only event record: no `version`/`previous_version_id` or `deleted_at` — old events are aged out by the retention policy, §9, not soft-deleted.)

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | |
| `tenant_id` | uuid FK | Scope. |
| `source` | text + CHECK | `git_webhook` / `polling` / `in_band` / `periodic_sweep` / `manual` |
| `commit_id` | text, nullable | If from git. Used for at-least-once dedup. |
| `commit_message` | text, nullable | |
| `commit_author` | text, nullable | |
| `files_changed` | jsonb | `[{path, change_type: add|modify|delete, lines_added?, lines_removed?}]` |
| `event_ref` | jsonb, nullable | Source-specific reference: webhook payload id, polling cycle id, or for `in_band` the originating agent + work-item id. |
| `received_at` | timestamptz | |
| `processed_at` | timestamptz, nullable | When the trigger filter ran. |
| `triggers_fired` | jsonb, nullable | `[{trigger_kind, dispatched_to_agent, work_item_id?}]`. `trigger_kind` is a controlled vocabulary (§4): `conventions_reextraction` / `mechanical_reprobe` / `schema_check` / `code_audit_refresh` / `intent_revalidation` / `freshness_check`. |
| `processing_error` | text, nullable | If processing failed (for retry / observability). |

*(Convention: the `source` column is `text` + `CHECK`, not a native enum, per `FOCUS_schema_verification_findings` §1.)*

Indexes: `(tenant_id, received_at)`; `(tenant_id) WHERE processed_at IS NULL` for the unprocessed queue; `(tenant_id, commit_id)` for dedup.

---

## 3. The four event sources

Events arrive through different mechanisms depending on tenant setup. The contract specifies the **shape** (§2) without forcing one mechanism — implementation per tenant.

- **`git_webhook`** — repo hosting sends a webhook on push. Lowest-latency; preferred where available. Listener service receives, normalises into a `change_events` row, fires the filter.
- **`polling`** — periodic git fetch; diff against the last known head. Fallback for repos without webhook support; tunable interval (tenant-configurable). Also useful as a backstop for missed webhooks.
- **`in_band`** — when the tool itself causes a change (the bundle builder applies a confirmed code change, an agent commits a doc update, a layer agent writes to the active config), the originating component emits a change event directly. This **closes the loop on self-modification** — without it, the tool's own changes evade the drift detector and the decision log. The event's `event_ref` carries the originating agent and the work-item that authorised the change.
- **`periodic_sweep`** — the maintenance agent's scheduled sweep emits a synthetic event so sweep-driven re-checks have the same first-class trail as commit-driven ones. Catches what other sources missed (webhook drops, polling lag, freshness expiries, drift in untracked surfaces).

`manual` is reserved for human-injected events (an operator telling the system "consider these files changed" — useful for testing, recovery, or external-system bridging).

---

## 4. The trigger filter — computed from the mechanical config

The filter maps `files_changed` to the set of maintenance triggers to fire. It is **derived from the mechanical config, not stored**.

The mechanical config already knows `code_paths.actions`, `code_paths.workflows`, `code_paths.migrations`, `doc_paths.root`, the Makefile path, etc. The filter reads these and classifies each changed path:

| Path matches | Trigger kind | Dispatched to |
|---|---|---|
| `doc_paths.root/*` (the standards docs) | `conventions_reextraction` | Conventions agent (§1) |
| `Makefile`, `go.mod`, CI config | `mechanical_reprobe` | Stack-discovery agent (§2) |
| `migrations/*.sql` | `schema_check` | Maintenance + schema validator |
| `code_paths.actions/*`, other code paths | `code_audit_refresh` | Conventions agent (running drift audit) |
| `objectives` table updates (not file-based) | `intent_revalidation` | Intent-elicitation agent in "do these still hold" mode |
| (no file change; scheduled) | `freshness_check` | Maintenance agent itself |

**Compute-on-read applied to routing.** When the mechanical config changes (docs move to a new path; a new code root is added), the filter updates automatically on the next event because it reads the current config. No separate maintained mapping to keep in sync — the same compute-on-read principle that keeps the priority profile fresh.

A single change event can fire multiple triggers (a commit touching both docs and code fires `conventions_reextraction` and `code_audit_refresh`); the `triggers_fired` field records all of them.

---

## 5. Processing flow

```
[event arrives] → change_events row inserted (processed_at NULL)
                → trigger filter runs against current mechanical_config
                → triggers dispatched to maintenance agent
                → maintenance creates config_work_items as needed
                → processed_at set; triggers_fired recorded
```

- **At-least-once delivery.** Webhooks and polling can both deliver the same commit twice. Dedup by `(tenant_id, commit_id)`.
- **Retry on failure.** A processing error leaves `processed_at NULL` and records `processing_error`; the listener retries with backoff.
- **No event silently dropped.** Even if no triggers fire (e.g., commit touched only files outside the tracked surface), the event row remains as `processed_at = <ts>, triggers_fired = []` — an explicit "this commit produced no maintenance work" record.

---

## 6. In-band emission — what counts as "in-band"

Concretely:
- The bundle builder applying a confirmed code change → emits `in_band` with the work-item that authorised it.
- A layer agent confirming a status transition in `standards`/`objectives`/`mechanical_config` (the §5 flip from `proposed` to `active` of the active-config schema) → emits `in_band`.
- A scheduled refresh that rewrites the constitution view → does not emit (compute-on-read, no underlying change).

The rule: **state changes emit; computed-view refreshes don't.** Anything that mutates an authoritative row emits an in-band event so maintenance sees its own effects.

**Guard — a confirmer apply does not re-trigger maintenance on the target it just confirmed.** When the central confirmer flips a layer row to `active` it emits an `in_band` event whose `event_ref` marks it as a confirmer apply. The trigger filter recognises this and produces `triggers_fired = []` for that target — the entry was just human-reviewed and confirmed (`last_verified_at` is now), so re-running conventions-reextraction or the drift audit on it would be a wasted cycle and could re-surface a freshly-approved entry. The event is still recorded (for audit and self-modification visibility). This guard is specific to the confirmer's own apply of an already-reviewed change; in-band events from **generation** (the bundle builder committing code) are *not* exempt and trigger normally, because generated code has not been reviewed by the maintenance lens.

---

## 7. Cross-repo and multi-surface considerations

Some tenants will have standards docs in a different repo from code. Each tracked surface gets its own integration — separate webhooks or polling — but events from all surfaces land in the same `change_events` table for the tenant. The trigger filter handles this naturally because it operates on paths (which include the surface), not on assumptions about a single repo.

For our own setup, code and docs live in the same repo; one integration suffices.

---

## 8. Relationship to other contracts

- **`config_work_items`** — maintenance triggers may create work-items (drift candidates for tenant resolution). The `change_events.triggers_fired` carries the work-item ids when created.
- **`decision_log`** — significant maintenance decisions arising from change events (a cluster of triggers, a bulk re-evaluation, a graduated audit) emit decision-log entries. The `change_events.id` becomes part of the decision's `subject_ref`.
- **`mechanical_config`** — the source of the trigger filter. Compute-on-read keeps the filter automatically current.
- **`trust_ledger`** — drift detected via change events feeds the ledger's evidence summary (§5.7 of maintenance feeding §3 of the trust ledger contract).

---

## 9. Open

- **Webhook delivery reliability** — at-least-once with dedup is the spec; specific provider semantics (GitHub, GitLab, Bitbucket, self-hosted) handled per integration.
- **Polling interval** — tenant-configurable with safe defaults. Trade-off: shorter interval = fresher detection + more network cost.
- **What in-band events are required vs optional** — the §6 rule covers the common cases; edge cases (e.g., a config-maintenance auto-apply that bypassed work-items) need a clear emit policy. Lean: anything that flips a row's status emits, period.
- **Bulk events** — a re-onboarding or a large code import produces many file changes in one commit; the filter may fire many triggers at once. Coalescing (one work-item per kind rather than one per file) is probably right but worth pinning when load shows.

---

## 10. One-line state

Change events arrive from four sources (git webhooks, polling, in-band, periodic sweep), land as `change_events` rows for first-class auditability, and pass through a trigger filter that is **computed from the mechanical config** (compute-on-read applied to routing) to fan out as typed triggers dispatched to the maintenance agent. In-band emission from the tool's own state changes closes the loop on self-modification so the drift detector and decision log see the tool's own effects.
