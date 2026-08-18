# 283 — CONTINUE HERE (2026-08-18, end of day). Mechanical programme COMPLETE and serving; next is the JUDGED pipeline design.

**The case file is the record** (`bugs_open/283_HANDOFF_…_element_ids_are_literal.md`, §13.1–13.8
for the execution arc). Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/` — the RUNBOOK's
§1 deploy-digest chain is the first thing to run in any session, and §11 has the corpus
classifier. Supersedes `283_CONTINUE_HERE_2026-08-17.md`.

---

## 1. State in one paragraph

Owner ruling (RFC_034): hybrid, LMC first, through the framework — **the mechanical
three-quarters is DONE, LIVE and RECONCILED.** 69 components converted and serving
instance-scoped ids (spot-checked at the served page on 4 domains: 0 unrendered tokens, 0
duplicate ids); 2 of the 69 since **deactivated by the webdesign rebuild lane** (13:52/14:13
UTC 08-18 — expected, pre-agreed via the dated pointer in their handoff; their templates still
carry the conversion). Council trail on correlation `07635a2f-3605-4e67-9a6d-7636b07f16ca` is
**closed: round 5 APPROVED** (13:50 UTC) — the arc ran REVISE → approved → approved → REVISE →
approved, and the round-4/5 exchange produced real value (the dual-active-row check, the
53-page side-effect sweep). **283 stays OPEN for exactly one reason: the judged pool — 25
components that genuinely declare into global scope, 23 of them the LMC calculators.**

## 2. What is LIVE (all digest- or artefact-verified)

| thing | state |
|---|---|
| `scope_component_instance` fix_type + gate (CLC-017) | in the running chassis since `v1.0.1307`; 70-item batch run through it |
| `template_changed` rerender reason + page-scoped fixer rerenders | migrations **460/461/462** applied (461 = the PREPARE-compile lesson; 462 = skip owned pages); snapshots in `agent_definitions_backup` |
| the conversion evidence | every converted row has a `component_versions` snapshot (`change_source='scope_component_instance'`); rerender ledger 53 complete + 15 cancelled + 4 failed = 72, all accounted (§13.7) |
| tripwire (CLC-016) | **RETIRED** after firing on schedule; manifests remain as the demand-control worked example |
| pattern-check `check_unscoped_component_render` | commit-time guard on new render call sites |

## 3. Do next, in order

1. **Design the judged pipeline for the 25** (the real remaining work — a design session):
   LLM rewrites each script (IIFE-wrap + rewire the 20 inline `on*=` handlers + replace the 8
   `window.onload`), through the framework; **`GateConvertedTemplate` is the acceptance gate**
   (it already refuses unscoped output); **byte-level truncation check on every rewrite**
   (`output_tokens == max_tokens` means CUT — `bugs_open/012`); LMC first because `oracle.py`'s
   **170 literal-id checks** are the independent witness. Judged list: RFC_034 §3a (23
   `loans-*`/`mortgages-*` + `tool-archetype-clash-calculator`, `tool-bayesian-ranking`).
2. **Before the first LMC conversion, two owed steps**: rebaseline `b2_verify`'s byte-identical
   check (converting ends that property — deliberate, not discovered mid-batch), and move
   `oracle.py`'s selectors in lockstep (one `c-<function>-` prefix per tool — the token rule was
   chosen to make this mechanical).
3. **The four parked residuals**, each with its route named in §13.7:
   `tool-process-automation-scorer` (rename the `ec2` id or judged pool);
   `tool-spawn-rate-balancer` (repair the `chartTitle` internal duplicate, then reconvert);
   the **two forked-function shrink refusals** (`tool-model-approach-selector`,
   `tool-llm-cost-calculator`) — one deliberate investigation: why does a fork's page plan only
   1 of 3 stored sections? (Suspect: sections referencing the OTHER fork's rows.)
4. **The companion tracking item** `item_key='rerender-reason-producers-283'` (parked
   `needs_human_review`): sibling producers of reason-less rerenders
   (`StoreGeneratedComponentAction.createRerenderWorkItem` first), the
   `create_rerender_items_action` consolidation, the dedup status-set alignment. Not this
   lane's to execute unilaterally — it is the round-4 bug_historian's tracked ask.
5. **RFC_032 remains open** (ComponentID unification — three context-builders, one name, two
   meanings). It gained urgency: the estate now has `InstanceID` fully adopted alongside it.

## 4. ⚠ Traps for the next session

- **Verify any "fresh build" at the DIGEST** (RUNBOOK §1). Two same-tag cache traps hit this
  lane in two days; the tag, restart time and local label all lie.
- **A converted template disappearing from a webdesign tool is the rebuild lane's regeneration,
  not a revert** — pre-agreed; conversions are idempotent and re-runnable after their rebuild.
- **Counts drift daily** (corpus 91→94→95; active 243→245). Derive, never paste — the
  batch-seed SQL and `cmd/instanceaudit` both derive at run time.
- **The two shrink-refused pages keep pre-conversion ids** until the fork investigation; their
  templates ARE converted, so any successful future rerender picks it up.
- **Judged rewrites are the `bugs_open/012` truncation class.** The gate catches unscoped
  output; it cannot catch a truncated-but-scoped fragment — the byte-level structural check is
  the only defence, per rewrite, no exceptions.
- **`{{.ComponentID}}` still means two different things** on two paths (LANDMINES). Untouched,
  by design — RFC_032's question, not a conversion bug.

## 5. Council practice for this lane (learned over five rounds)

One correlation (`07635a2f…`) carries the whole trail — always `RESUBMIT_CORR`, never a fresh
id. Config migrations under `sql_for_agents/` need `FORCE=1` (the path filter cannot see that
they are platform changes). Answer a REVISE with **measurements, not defence** — both REVISE
rounds found real things (a submission-sketch gap; the dual-active-row check and the
53-page sweep). The verify block of any migration that writes an embedded query must
**PREPARE-compile it** (LANDMINES, the 460→461 lesson).
