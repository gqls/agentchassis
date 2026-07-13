# FOCUS — Whole-Plan Review (second pass)

**Status:** a hard pass across the entire body of work, not just the contracts: the tool plan, the onboarding config plan, the five onboarding agent specs, the six contracts, the salience/mediation design, and the master plan — read against each other for cross-cutting inconsistencies after the schema verification and capabilities-catalog corrections. Like the contract-set review, findings are by severity; fixes applied are marked.

---

## 1. Inconsistencies and gaps to fix

### 1.1 The bundle shape is an undefined contract (the next load-bearing one)

The system-view check found the active-config schema was undefined though every agent touched it. The same hole exists one step downstream: the **bundle builder** (tool plan Phase 1 — the MVP, the earliest-value piece) reads the active config and produces a **bundle**, but the bundle's shape has no contract. Every consumer of the bundle — the generation step, the paste-into-chat path, later the cascade/mediator — needs a stable shape, exactly as the config consumers did.

**Recommendation:** spec the bundle shape before/at the start of Phase 1. It is the active-config schema's downstream twin and the next load-bearing contract. Not done this pass (it's a new contract, not a fix), but flagged as the top pre-Phase-1 item.

### 1.2 In-band confirmation events could make maintenance re-process what was just confirmed

The confirmer emits an `in_band` change event when it flips a layer row to `active` (change-layer §6: state changes emit). That event passes through the trigger filter, which could fire `conventions_reextraction` or `code_audit_refresh` on the very entry just confirmed — a redundant cycle, and at worst maintenance re-flagging a freshly-approved entry.

In practice the freshness check (the entry's `last_verified_at` is now) and the drift detector (logged premise just matched current) prevent a *false* re-flag. But the wasted cycle is real and the loop should be closed explicitly.

**Fix applied (change-layer contract):** an `in_band` event whose `event_ref` identifies it as a **confirmer apply** does not fire re-check triggers for the target it just confirmed; it is recorded (for audit and self-modification visibility) but produces `triggers_fired = []` for that target. Self-modification from *generation* (the bundle builder committing code) still triggers normally — it's only the confirmer's own apply of an already-reviewed change that is exempt.

### 1.3 Tool-plan phasing is out of sync with the contract work

`PLAN_context_assembly_tool_and_service` predates the six contracts. Two corrections:
- **Phase 0 foundations** should now explicitly include implementing the active-config schema and the work-items / decision-log / trust-ledger / capabilities-catalog / change-layer contracts — they are the foundation the bundle builder and everything else stands on.
- **The onboarding split** needs to be explicit: onboarding *ourselves* (producing our own config) is a **Phase 0/1 prerequisite** — the bundle builder has nothing to read until our config exists — whereas onboarding *arbitrary tenants* (sandboxing, code-inference mode, arbitrary-repo config derivation) is **Phase 3 service** work. The five agent specs mostly describe the general case; which parts are needed now vs for the service should be tagged.

**Recommendation:** a phasing-sync edit to the tool plan. Flagged, not applied this pass (it's a revision of a plan, best done deliberately).

### 1.4 `capability_id` casing — conscious decision, not accident (note applied)

`capability_id` values (`write_go_action`) are snake_case. The naming convention (`FOCUS_naming_conventions_kebab_vs_snake`) makes this the right call — snake is for type values used as **dispatch/lookup keys** (the doc cites `site_work_items.item_type = needs_blog_posts` as the precedent), and `capability_id` is exactly that (the ledger key, the operation→capability map key). But the **sibling `agent_definitions.capabilities` tags are kebab** (`git-commit`, `web-search`) because they are descriptive metadata, not keys. The two coexist correctly under the convention's own rule. **Note applied** to the capabilities catalog so this is a recorded decision, not a latent inconsistency of the kind the naming doc was written to prevent.

---

## 2. Clarifications (no contradiction, but worth stating)

### 2.1 There are two gated-mutation mechanisms, not one

"One path to `active`" (the central confirmer) governs the **config-entry layer tables** (`standards`, `objectives`, `mechanical_config`). The **trust ledger** is a different domain (state, not config entries) with its **own** mutation flow — graduations gated via work-items, de-graduations allowed to auto-apply with notification. The "one confirmer" principle should not be read as covering the ledger. Two appliers, both using work-items + decision-log, but distinct: the config confirmer and the ledger's ratchet-evaluator. Worth stating so the asymmetric de-graduation isn't seen as a hole in the "one path" rule.

### 2.2 "Every decision emits a log entry" is intentionally high-volume

Every bundle assembly logs a `decision_log` entry (with `inputs_used`). That is one entry per task minimum — deliberately, because the retrospective audit ("did this decision consider the right elements") depends on it. The compact `inputs_used` form keeps each entry cheap, but the volume is real and the retention policy (decision-log §8, open) must be set with this in mind. Not a contradiction — a consequence to size for.

### 2.3 Mediation-session persistence is a future contract

The optimal machinery (multi-author, N-round convergence, the no-one-owns-it candidate + diff log from the salience doc §12) produces a runtime artifact — the diff log of a mediation session. The `decision_log` captures the *outcome* (`mediator_arbitration`), which is enough for now. Persisting the full session diff log is a Phase-4 contract, not needed before the mediator is built. Noted so it isn't mistaken for a current gap.

### 2.4 Cross-cluster is handled below the contract layer

The contracts are tenant-scoped and cluster-agnostic. The cross-cluster mechanics (remote DB via tunnel + PgBouncer at the same DNS, per the multi-cluster design) are boundary-invisible by design, so the contracts need no cluster-awareness — a tenant's tables are reached the same way regardless of cluster. Consistent; stated so no one adds cluster columns the design deliberately avoids.

### 2.5 Fallback/tolerant-reading is disallowed — applies to the whole plan

The naming doc reaffirms the standards rule that **fallbacks should not be relied upon** (tolerant readers were rejected for `page_type`). This applies across the plan: the bundle builder, the confirmer, the trigger filter should fail loudly on an unexpected shape, not normalise-and-continue. Anywhere the plan implies "tolerate and carry on," it conflicts with this standard. No specific violation found, but a lens to hold when building.

---

## 3. Affirmed consistent (checked, holds)

- **Priority-profile representation** — order + sealed/constraint flags in the salience doc §9.6 matches the `objectives.priority_profile` jsonb shape in the active-config schema §3.4.
- **Two precedence directions** (child-wins / sealed-ancestor-wins) — captured by the `sealed` flag plus compute-on-read merge.
- **Constitution** — a view over `standards WHERE scope = constitution`, not a separate table. Consistent across the bundle read patterns.
- **Cold-start + dogfooding** — onboarding's own layer agents run at `confirm_every` (cold start), consistent with "dogfood the ratchet."
- **Capture vs use** — intent-elicitation captures, the user-rep advocate uses; clean separation.
- **Capabilities-catalog seeding** — after the correction, seeding *defines* trust-units rather than extracting from the (kebab, descriptive) tag column. Consistent.
- **`item_type` snake-casing** — `config_work_items.item_type` values (`confirm_proposal`) are snake, matching the `site_work_items.item_type` precedent.

---

## 4. Recommended order (unchanged in spirit, sharpened)

1. **Spec the bundle shape** (§1.1) — the next load-bearing contract, before Phase 1.
2. **Phasing-sync the tool plan** (§1.3) — contracts into Phase 0; onboarding-ourselves vs onboarding-tenants split.
3. **Implement Phase 0 foundations** — the six contracts as real tables (verifying `site_work_items`/`agent_definitions` assumptions hold at DDL time, per schema-before-SQL), plus the constitution for our repo.
4. **Phase 1** — the Go analyser + bundle builder against the now-stable contracts.

---

## 5. One-line state

The whole plan is internally consistent after this pass, with one genuine gap (the bundle shape — the active-config schema's undefined downstream twin, and the top pre-Phase-1 item), one loop closed (confirmer in-band events no longer re-trigger maintenance on just-confirmed targets), one phasing-sync owed to the tool plan, and a recorded naming decision for `capability_id`. The clarifications (two mutation mechanisms, intentional log volume, deferred mediation-session persistence, cluster-agnostic contracts, no-fallbacks) are stated so they aren't mistaken for holes later.
