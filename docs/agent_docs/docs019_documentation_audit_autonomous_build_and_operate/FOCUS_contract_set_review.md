# FOCUS — Contract Set: Final Review

**Status:** cross-contract review of the five contracts, read against each other and against the agents that use them — not each in isolation. Findings by severity. Fixes applied in this pass are marked; recommendations and judgement calls are flagged for decision.

**The five contracts:** `PLAN_active_config_schema`, `PLAN_config_work_items_contract`, `PLAN_decision_log_contract`, `PLAN_trust_ledger_contract`, `PLAN_change_layer_integration_contract`.

---

## 1. Self-contradictions (must fix — fixed in this pass)

### 1.1 `target_ref` shape doesn't fit `mechanical_config`

The work-items contract defines `target_ref` as `{id, version}` and indexes on `target_ref->>'id'`. That works for `standards` and `objectives` (atom tables with per-row `id` + `version`), but **`mechanical_config` has no atom-level id** — the active-config schema deliberately stores it as one JSONB row per tenant. A work-item proposing a change to, say, the build command has nothing to point `{id, version}` at.

**Fix applied:** `target_ref` is now **polymorphic** — `{table:'standards'|'objectives', id, version}` for atom tables, `{table:'mechanical_config', tenant_id, json_path}` for the mechanical layer. Index adjusted accordingly. This preserves the active-config schema's storage decision rather than forcing the mechanical layer into atom rows.

### 1.2 `modified` is used as a status but isn't in the status enum

The work-items status enum is `pending / in_progress / confirmed / rejected / deferred / expired`, but the lifecycle diagram and a bullet treat `modified` as a terminal status. The `resolution` enum already carries `accept / reject / modify / defer`, so the modify/accept distinction belongs there, not in `status`.

**Fix applied:** a modification is `status = confirmed` with `resolution = modify` and the edited content in `resolution_payload`. The lifecycle diagram and bullets reconciled. `status` stays the lifecycle; `resolution` stays the outcome.

---

## 2. Gaps visible only in combination

### 2.1 The confirmer — who applies a confirmed work-item (resolved in this pass)

Three contracts hedge: the active-config schema and the work-items contract both say "the source agent **or a dedicated confirmer service**" applies the confirmed change. Left as a choice, this is a real ambiguity, and the wrong choice weakens confirm-not-initiate: if every agent reimplements the apply logic (flip `proposed`→`active`, deprecate prior, write the decision-log entry, emit the in-band change event), there are five code paths to `active` and five chances to get the gate wrong.

**Resolution applied:** a **single central confirmer**. One component watches for resolved work-items and performs the apply uniformly: flip the layer row to `active`, set `last_verified_at`/`verified_by`, deprecate the prior version with `supersedes`, write the `decision_log` entry (carrying the human's `resolution_premise`), emit the `in_band` change event. **One path to `active`** makes confirm-not-initiate airtight and the apply logic auditable in one place. The work-items contract is updated to specify this.

### 2.2 The capabilities catalog — referenced, unspecced

The trust ledger references a `capabilities` catalog for the **ceiling** and the static verifiability/containment factors, but the catalog itself is only sketched (§4, §8 of that contract). This is the trust ledger's analog of what the active-config schema is for the layer agents — a referenced-but-undefined dependency.

**Recommendation (not yet applied):** spec the `capabilities` catalog as a small sixth contract before the trust ledger is implemented — `capability_id`, `display_name`, `ceiling`, `verifiability`, `containment`, `cascade_tier_hints`. Small, but the ledger has an open end without it.

### 2.3 Concurrent / competing proposals on the same target (rule added)

Nothing in the set says what happens when a second proposal arrives for a target that already has a pending work-item — e.g., a work-item proposes a new build command, and before it resolves, a Makefile commit makes the maintenance agent propose a different one. Two competing proposed versions, no rule for which wins.

**Rule added** to the work-items contract: a new proposal on a target with an unresolved work-item **supersedes** the pending one — the older work-item is marked `expired` with a note, and the newer proposal carries forward. This keeps the tenant from confirming a stale proposal. (The alternative — block the new proposal until the old resolves — risks the tenant confirming something already overtaken by reality, which is worse.)

### 2.4 Bootstrapping and cold-start (notes added)

Two unstated starting conditions:
- **First `tenant_configs` row.** The onboarding orchestrator must create the `tenant_configs` row (with `onboarding_state = not_started`) directly at initialisation — before any work-item gate exists for that tenant. This is not a confirm-not-initiate violation: `tenant_configs` is the **scope-holder**, not a config value. Noted in the active-config schema.
- **Cold-start trust level.** A new `(tenant, capability)` pair has no ledger row. The first time a capability is invoked for a tenant, the ledger row is created at **`confirm_every`** (the most conservative level, bounded by the capability's ceiling), not at any inherited or default-higher level. Noted in the trust ledger contract.

### 2.5 Unverified reuse claims (must check before implementation)

Several contracts claim to mirror or extend existing tables — `config_work_items` "mirrors `site_work_items` (same status enum, same chain structure)"; the capabilities catalog "extend `agent_definitions`." **These are assumptions; the actual schemas have not been checked.** Per the standing rule (schema-before-SQL), these claims must be verified against the live `site_work_items` and `agent_definitions` schemas before any DDL is written — the real shapes may differ from what the contracts assume, and the contracts should be corrected to match reality, not the reverse. Flagged in the affected contracts; resolved at implementation start by inspecting the schemas.

---

## 3. Clarifications (no contradiction, but ambiguous)

### 3.1 "Confirmation by reality" vs the work-item gate — graduated, not bypassed

The stack-discovery spec says the mechanical layer is "confirmed by reality" (probe success), while the active-config schema says **all** transitions to `active` go through the work-item gate. Read carelessly these conflict. They don't, and the resolution is worth stating because it unifies two mechanisms:

- **Initially**, mechanical entries go through the work-item gate like everything else (`confirm_every`). Probe success is *strong evidence* that makes the confirmation near-rubber-stamp, but the human still gates activation.
- **As the stack-discovery capability proves reliable**, it graduates via the trust ledger to auto-activate on probe success (`notify` level) — the mechanical layer is the natural **first capability to graduate** past `confirm_every`, precisely because it has the strongest verifiability.

So "confirmation by reality" is not a bypass of confirm-not-initiate; it is the *graduated state* the trust ledger eventually permits. The gate and the ratchet work together — confirm-not-initiate is the starting position, graduation is what relaxes it. Worth noting in both specs.

### 3.2 Idempotent apply

If a confirmed work-item's apply step runs twice (retry after a partial failure), it must not double-apply. The `version`/`supersedes` mechanism makes this natural — applying a version already `active` is a no-op — but it should be stated as a requirement on the central confirmer (§2.1), not left implicit.

### 3.3 Provenance shape is deliberately scoped

The common provenance shape (source / source_ref / confidence / status / last_verified_at / verified_by / freshness_until / version / supersedes) appears on config-entry tables (`standards`, `objectives`, embedded in `mechanical_config`) but **not** on `decision_log`, `trust_ledger`, `config_work_items`, or `change_events`. This is correct, not an omission: those four are events, history, state, and a queue respectively — not config entries — and carry their own appropriate shapes. Confirming it is intentional so a later reader doesn't "fix" the inconsistency by forcing provenance onto a log.

---

## 4. What's solid (affirmed)

- **One path to `active`** (once the confirmer is centralised, §2.1) makes confirm-not-initiate airtight rather than a discipline.
- **State vs history separation** (trust ledger / decision log) is clean and the access patterns justify two artifacts.
- **Compute-on-read + log-at-use** resolves the freshness-vs-audit tension coherently across the set.
- **Versioned atoms + decision log** give full retrospective reconstructability, which the compact `inputs_used` form depends on — the pieces fit.
- **Reuse of existing patterns** (work-items, agent_definitions) keeps this from being new infrastructure — *subject to §2.5 verification*.
- **The change-layer's in-band emission** closes the self-modification loop, and the "state changes emit, view refreshes don't" rule is a clean test.

---

## 5. Status after this pass

**Fixed in the contracts:** §1.1 (polymorphic target_ref), §1.2 (modified status), §2.1 (central confirmer), §2.3 (concurrent-proposal rule), §2.4 (bootstrapping + cold-start notes), §3.2 (idempotent apply on the confirmer).

**Recommended before implementation:** §2.2 (spec the capabilities catalog as a sixth contract), §2.5 (verify `site_work_items` and `agent_definitions` schemas against the contracts' assumptions), §3.1 and §3.3 (clarifying notes in the affected specs).

---

## 6. One-line state

The contract set is internally consistent after fixing two self-contradictions (the mechanical-layer `target_ref` and the `modified` status), consolidating the apply path into a single central confirmer (one route to `active`, airtight confirm-not-initiate), and adding rules for concurrent proposals and cold-start. Two items remain before code: spec the capabilities catalog (the trust ledger's open dependency) and verify the reuse claims against the live `site_work_items` and `agent_definitions` schemas per schema-before-SQL.
