# PLAN — Decision Log Contract (Published Reasoning + Inputs)

**Status:** contract specification. The published-reasoning log named in `MASTER_autonomous_build_and_operate` and referenced by the maintenance agent (§5.4, §5.5, §5.7) and throughout the running notes, now given concrete shape. Extends the published-reasoning idea with the user's input-snapshot requirement: log the *computed inputs that were in hand* at the decision, not only the premise and outcome.

---

## 1. Purpose — the freshness-vs-retrospect resolution

The active config is **computed on read** (the effective priority profile by walking root to leaf, the constitution as a view over `standards` with scope=constitution, the change-type bundle as a manifest query). This keeps it fresh — an ancestor change is reflected everywhere below it on next read, with no stored derivation to go stale.

The cost is that historical reconstructability disappears: nothing on disk records what the system *had in hand* at a past decision. Was the priority profile that drove this decision X or Y? Did the agent see the constitution as it stood that day, or the version after the law was added?

**Resolution:** compute on read for freshness; **log the result at point of use** for retrospect. The active value stays fresh because nothing stale is stored as authoritative. The log is a historical record — it wasn't authoritative when written and isn't now. It is audit material, not a source. Combined with the versioned atoms the active-config schema already supports, the log gives full historical reconstructability without poisoning the present.

---

## 2. The table — `decision_log`

**Immutable.** Append-only; never updated, never deleted. Archival only.

| Field | Type | Purpose |
|---|---|---|
| `id` | uuid PK | Identifier. |
| `tenant_id` | uuid FK | Scope. |
| `decision_kind` | enum | `bundle_assembly` / `confirm_proposal` / `drift_disposition` / `routing_choice` / `gate_decision` / `intent_capture` / `convention_extraction` / `mediator_arbitration` (etc. — extensible) |
| `subject_ref` | jsonb | What was being decided: `{kind, id}` — references a task, a work-item, an atom, a change. |
| `inputs_used` | jsonb | The active-config slice in hand at the decision. **The retrospect anchor.** Default compact form (§3); optional full snapshot for high-stakes decisions (§4). |
| `premise` | text | The reasoning premise. "Chose X because for this node security was a constraint not a weighted objective, and the candidate met it at acceptable speed cost." Foundational — see running notes. |
| `decision` | jsonb | What was decided. Shape varies by `decision_kind`. |
| `decided_by` | text | Agent identifier (`stack_discovery`, `mediator`, `bundle_builder`, etc.) or user id. |
| `decided_at` | timestamptz | When. |
| `parent_decision_id` | uuid FK, nullable | Chains (a mediator arbitration may be the parent of several routing choices). |

Indexes: `(tenant_id, decided_at)` for time-range queries; `(tenant_id, decision_kind, decided_at)` for kind-scoped audits; `(subject_ref->>'id', subject_ref->>'kind')` for "decisions about this thing."

**Why immutable.** A log that can be edited cannot be trusted to reconstruct past state. Immutability is what makes retrospective audit possible at all.

---

## 3. `inputs_used` — the compact form (default)

Most decisions log inputs compactly — enough to reconstruct, not the full data. Combined with versioned atoms, this gives full reconstructability without storing the merged result.

```jsonc
{
  "active_config_at": "<timestamp>",            // logical "as of" time

  "standards_used": [
    { "id": "<uuid>", "version": 4 },           // the exact atom versions
    { "id": "<uuid>", "version": 2 },
    ...
  ],

  "objectives_walk": [                          // root → node path
    { "id": "<root-uuid>", "version": 3 },
    { "id": "<branch-uuid>", "version": 1 },
    { "id": "<leaf-uuid>", "version": 7 }
  ],

  "mechanical_config_version": 12,              // whatever versioning is applied

  "merged_hashes": {                            // content hashes of the computed views
    "constitution":         "sha256:...",
    "effective_priority":   "sha256:...",
    "applicable_standards": "sha256:..."
  }
}
```

To reconstruct the historical state: look up each atom by `id` and `version` in `standards` / `objectives` / `mechanical_config` (which retain old versions via `supersedes`), walk and merge per the schema's effective-on-read rules, and verify against `merged_hashes`. The hash makes tampering detectable.

---

## 4. `inputs_used` — the full snapshot form (high-stakes decisions)

For high-stakes decisions — gate passes, autonomy graduations, sealed-constraint changes, mediator arbitrations on contested cross-area changes — store the full merged result inline rather than relying on reconstruction:

```jsonc
{
  ...compact form above...,
  "full_snapshot": {
    "constitution_text":           "...",
    "effective_priority_profile":  { ... },
    "applicable_standards":        [ ... full atoms ... ],
    "objectives_path":             [ ... ]
  }
}
```

Trade: storage cost for guaranteed retrievability even if an atom is later corrupted, lost, or the schema migrates in a non-back-compatible way. Worth paying on the small fraction of decisions where audit is most important.

`decision_kind` determines which form is used; agents emit the appropriate one. The set of high-stakes kinds is configurable per tenant.

---

## 5. Read patterns

Concrete uses so the log earns its keep:

- **Drift detector** (per `FOCUS_salience_and_multi_author_mediation` §11): for each recent decision, compare `premise` against the *current* priority profile / standards / objectives. A gap is the canonical drift signal.
- **Heuristic-invalidation hook** (`MASTER` §7.5 and salience §9.4): when a profile or standard changes high in the tree, query `decision_log` for decisions whose `inputs_used` referenced atoms invalidated by the change → flag those decisions for re-mediation.
- **Trust-ledger signal** (per maintenance §5.7): sustained no-drift across recent decisions in a capability → reliability evidence (graduation candidate); repeated drift → regression evidence (de-graduation trigger).
- **Retrospective audit**: "did this decision consider the elements it should have?" → inspect `inputs_used`; verify the right standards and the right objective path were present. **This is the original requirement that prompted the input-snapshot extension.**
- **Compliance review**: for a legal-sourced sealed constraint added on date D, query decisions made after D in affected areas to confirm the constraint was in the `inputs_used`.

---

## 6. Write discipline

- **Every decision emits a log entry.** No silent decisions; no "we decided this but didn't log it." Agents that fail to log do not decide — logging is part of deciding.
- **The entry is written *before* the decision is applied.** If the apply step fails, the log still records the intent, which is itself information. (A separate `applied_at` field could record application separately if needed — open in §8.)
- **The `premise` is human-readable.** Not "rule_id=42"; a written reason. Reusable later by humans and by reasoning over the log.

---

## 7. Relationship to other contracts

- **The published-reasoning log** named in `MASTER` and the running notes **is this table.** Same artifact, given concrete shape. Term unified.
- **Work-items resolutions feed here.** When a tenant resolves a `config_work_items` row, the resolution premise (the human's stated reason) becomes the `premise` of a corresponding `decision_log` entry. Human decisions are first-class citizens in the log alongside agent decisions.
- **The active-config schema's versioned atoms are what makes the compact form work.** Without `version` and `supersedes` on layer rows, historical reconstruction would require full snapshots everywhere.

---

## 8. Open

- **Separate `applied_at` vs `decided_at`** to distinguish recorded-intent from applied-outcome, for cases where the apply step fails after logging. Likely useful; deferred until the apply machinery is concrete.
- **Retention policy.** The log grows unboundedly. Archive policy (move to cold storage after N months? compact full snapshots to compact form after M months?). Worth pinning before the log is heavily used.
- **The set of `decision_kind` values** is extensible; the initial set comes from the agents that currently emit — expanded as new decision points appear.
- **Whether to surface the log to the tenant** (as a "decision history") or keep it as audit-only. Probably surface, filtered, since retrospective visibility is part of why it exists.

---

## 9. One-line state

The decision log is the published-reasoning log given concrete shape: immutable, append-only, one row per decision, carrying the `premise` (the reasoning) and the `inputs_used` (the active-config slice in hand at decision time — compact references for routine decisions, full snapshots for high-stakes ones). Combined with versioned atoms in the layer tables, this resolves the freshness-vs-retrospect tension by **computing on read and logging at point of use** — the active value stays fresh; the log preserves what the system actually had in hand whenever it acted.
