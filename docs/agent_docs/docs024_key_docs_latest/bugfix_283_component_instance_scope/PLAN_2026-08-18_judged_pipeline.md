# PLAN 2026-08-18 — the judged pipeline for the 25 (283's last quarter)

Designs the LLM ("judged") half of RFC_034's hybrid: the 25 components whose scripts genuinely
declare into global scope — 23 LMC calculators + `tool-archetype-clash-calculator` (vonc.com) +
`tool-bayesian-ranking` (gamesdesign.co.uk). The mechanical three-quarters is DONE and live
(case file §13.7). This is the design only; the build is the next task and goes through the
council gate on the lane correlation (`RESUBMIT_CORR=07635a2f-…`).

Every figure in §1 was re-measured this session (2026-08-18, evening) against the live DB and
the working tree at `0b185bad2` — pods verified running that commit by the RUNBOOK §1 digest
chain first.

---

## 1. The facts the design stands on (all verified this session)

| fact | value | evidence |
|---|---|---|
| judged components | 25 rows, 26 placements (each on its own page; no page places one twice) | corpus join re-run live |
| **all 23 LMC calculators sit on `rebuild_policy='owned'` pages** | 24 placements (`mortgages-repayment` on 2 pages) | same query |
| the 2 non-LMC tools sit on **generic** pages | vonc.com 28,737 B; gamesdesign.co.uk 10,520 B | same query |
| `component_level` of the placed 25 | **22 `section` + 3 `tool`** | live query |
| `template_changed` rerender path **excludes owned pages** | mig 462: `p.rebuild_policy IS DISTINCT FROM 'owned'` in the fixer's `create_rerender` | read live workflow config |
| the owned-page delivery path binds the token | `section_editor_actions.go:850` and `:948` call `BindSingleSectionInstanceToken` (occurrence 0) | read at the arm |
| occurrence 0 is correct for every placement | no page places any of the 25 twice (26 distinct component×page pairs) | placement query |
| `execute_llm_prompt` hard-refuses capped completions | `ai_actions.go:409–517`: refusal unless `tolerate_truncation` opted in per step | read |
| the whole-template write guard exists and is comparative | `component_write_guard.go` — collapse ratio + ends-mid-token, calibrated on live history | read |
| `GateConvertedTemplate` renders two instances through the REAL layer and detector | `component_instance_conversion.go:220` | read |
| the mechanical arm already routes the 25 | returns `fixed:false, action:"needs_script_scoping"`, writes nothing | `fix_component_template_action.go:1081` |
| `conditional` supports string `==` | `conditional_branch_action.go:227` | read |
| batch seed shape | `item_type='instance_scope_conversion'`, `handler_agent='component-template-fixer'` | live items query |
| oracle addresses tools by literal `#id` selectors | e.g. `#amount` → post-conversion `#c-loans-standard-calc-amount`; one prefix per tool | `oracle.py:133` + `InstanceToken` |

Two of these change the design materially, so they are called out:

- **Delivery must be split.** The proven `template_changed` rerender covers only the 2 generic
  pages. The 23 LMC conversions deliver via the section-editor (`apply_section_edit`), the
  sanctioned owned-page path — tool-improver has delivered 42 tool fixes through exactly that
  step shape.
- **The judged writer must be fan-out-intended, not page-scoped.** 22 of the 25 are
  `component_level='section'`; tool-improver's fenced write (`sharedComponentWriteCheck`)
  REFUSES a non-tool component placed on >1 page — which `mortgages-repayment` is. The write
  belongs beside `scope_component_instance` in `fix_component_template_action.go`, declared
  fan-out-intended in `component_template_writer_coverage_test.go` with this line as the cited
  reason.

## 2. The decision: extend `component-template-fixer`, self-routed — not tool-improver

The judged pipeline is a **branch of the existing 283 programme agent**, triggered by the
mechanical arm's own refusal. Reasons, in order of weight:

1. **The router already exists.** `apply_fix`'s `needs_script_scoping` refusal
   *is* the classification. ~~it fired correctly on all 25 during the batch~~
   > **CORRECTED 2026-08-19: FALSE — the 25 were never seeded; the batch excluded them.** The
   > classification came from `cmd/instanceaudit` (the detector), not from production runs of
   > the converter. Written in the confident voice with no marker — exactly the
   > `WRONG_CALLS.md` shape (logged there). The claim is now TESTED instead: the build's
   > fixtures run the real refusal on real judged-pool bytes. No second
   classifier, no second seed shape: the judged batch is seeded exactly like the mechanical one
   (`instance_scope_conversion` items, one per component ROW), and the fixer routes each item
   mechanically or judged by looking, not by being told.
2. **One audit trail.** Snapshots land in `component_versions` under sibling
   `change_source` values (`scope_component_instance` / `scope_component_instance_judged`);
   reconciliation queries from §13.5/§13.7 carry over unchanged.
3. **tool-improver's write contract is wrong for this** (the fence fact above), and its
   subject is a page's finding, not a component fan-out.

## 3. The pipeline, per component

```
ensure_site_record
  → apply_fix                     (existing scope_component_instance arm, unchanged behaviour;
                                   result EXTENDED to carry: ids-converted template, the
                                   inline-handler inventory (on*= attrs found), onload count)
  → check_scope_route  [NEW]      conditional: fix_result.action == needs_script_scoping
        │ else → check_needs_rerender          (mechanical fixed:true, or terminal refusal)
        ▼ then
  → scope_script_llm   [NEW]      execute_llm_prompt, claude-sonnet-5, max_tokens 32000,
                                   NO tolerate_truncation (capped completion = step failure)
  → apply_judged_write [NEW]      action fix_component_template, fix_type
                                   scope_component_instance_judged — gate + snapshot + write
                                   fused in ONE Go action (§4); on refusal → fail_work_item
                                   needs_human_review, nothing written
  → check_needs_rerender          (existing conditional, unchanged)
  → create_rerender               (existing: template_changed, generic pages only)
  → create_section_edit_delivery [NEW]  query_database INSERT of section_edit items for
                                   OWNED placements (config copied from tool-improver's proven
                                   create_rerender_item; NOT EXISTS dedup like create_rerender)
  → compose_note → append_note → complete
```

**The LLM's brief is deliberately narrow.** It receives the **ids-converted** template (the
deterministic pass has already renamed all ids, `getElementById` strings, `label for=`, CSS
`#id`, `data-*` — the surfaces it is proven on) and is asked to do ONLY the judged part:

- wrap the inline script in an IIFE (`(function(){ 'use strict'; … })();`);
- convert each inline `on*=` handler (inventory supplied) to `addEventListener` inside the
  IIFE, removing the attribute from the markup;
- replace `window.onload =` with a `DOMContentLoaded` listener (or direct invocation at IIFE
  end when the script only wires handlers);
- preserve the `/* tool-doc */` header comment inside the script;
- change NOTHING else — no id renames, no markup edits beyond removing `on*=` attributes, no
  reformatting;
- output the full template, no fences, no commentary.

Rationale: the judged surface is exactly the script + its handler wiring (RFC_034 §3: all 20
inline handlers and all 8 `window.onload`s live inside these 25). Everything mechanical stays
mechanical, and "changed nothing else" becomes *checkable* (§4.2).

## 4. The gate (`scope_component_instance_judged`) — checks in order, each refusing loudly

The action re-loads the row and **re-derives the ids-converted baseline B itself** by re-running
`ConvertTemplateToInstanceScope` (deterministic ⇒ identical) — it never trusts workflow-carried
bytes. If the gate on B now says the script is ALREADY scoped (corpus drift), it converges: write
B via the mechanical contract and skip the LLM output entirely. Otherwise, for LLM output T:

1. **`GateConvertedTemplate(function, T)` must return fully clean** — two instances rendered
   through the real layer: 0 duplicate ids, 0 unrendered/empty tokens, 0 unscoped scripts,
   ≤1 onload. This is the §2.1 rule made mechanical for the judged half too.
2. **Markup parity outside the script**: strip `<script>…</script>` bodies from B and T;
   `removeInlineHandlers(markup(B))` must equal `markup(T)` after whitespace normalisation
   (collapse runs to one space). Catches the LLM "improving" markup, dropping a section, or
   renaming an id. *Known risk: false refusals from cosmetic reformatting — acceptable, because
   the failure direction is refuse-to-human, never write; relax only on evidence.*
3. **Id-set parity**: the set of `id=` values in T == in B. (Also a truncation tell — a cut
   template loses ids.)
4. **The existing comparative write guard** (collapse ratio, ends-mid-token) against the
   CURRENT row — the `bugs_open/012` class, second layer behind the LLM step's own
   capped-completion refusal.
5. Snapshot to `component_versions` (`change_source='scope_component_instance_judged'`), write,
   return `fixed:true` + counts, naming the delivery each page will take.

Any check fails → `fixed:false`, named check in the reason, **nothing written**, item →
`needs_human_review`. No automatic retry in v1: 25 components; a handful of refusals is triage,
not a queue. (A single gate-fed retry is a v2 option if refusals cluster.)

## 5. Verification and sequencing

**Order: LMC canary → LMC batch → the 2 generic tools.**

1. **Before the first conversion (the two owed steps, case file §12.5.3):**
   - **b2_verify rebaseline** — conversion ends the byte-identical property *by design*
     (RFC_034 §5.1). The LMC lane is told via CONTRIB (§6) BEFORE the canary, with a veto
     window.
   - **oracle lockstep mechanism agreed**: per converted tool, `#id` → `#c-<function>-id` in
     `oracle.py` — mechanical, one prefix per tool, shipped in the same commit as that tool's
     conversion verification, with oracle runs before (PASS 170) and after (PASS 170 restored)
     plus the `--mutate expectation` control in the same session
     (two-clean-runs / mutation-control practice).
2. **Canary: `loans-standard-calc`** (2,469 B, single page, oracle-covered, the worked example
   at `oracle.py:126`). Full chain: item → judged branch → gate → write → section_edit delivery
   → served page shows prefixed ids, 0 unrendered tokens → oracle selectors moved → PASS 170.
   **The canary also proves the one [UNVERIFIED] link in this design**: that a `section_edit`
   item with empty `field_updates` re-renders the slot from the converted template and deploys
   it (tool-improver's 42 completes make this likely; it has not been watched end-to-end on an
   LMC owned page).
3. **Batch: the remaining 22 LMC calculators**, monitored drain like the mechanical batch;
   oracle run per delivered tool (or per small wave), selectors moved in lockstep.
   `mortgages-repayment` (2 pages → 2 section_edit items) gets a deliberate look.
4. **Last: the 2 generic tools** — no oracle exists there; verification is
   `cmd/instanceaudit` on the written template, `DetectInstanceCollisions` + 0 unrendered
   tokens on the served page after the `template_changed` rerender, and a manual functional
   click-through. They run last so the pipeline has 23 proven passes behind it.

## 6. Cross-lane obligations (owner ruling 2026-07-29 §3: consumers are TOLD, not measured)

- **LMC lane** — CONTRIB filed in their directory this session: b2_verify's byte-identity ends
  at conversion; oracle selector lockstep; delivery via section-editor touches their owned
  pages; canary named; veto window before it runs.
- **webdesign rebuild lane** — unaffected by this half (their 2 regenerated components were
  mechanical-pool; none of the 25 is theirs), no action.
- **copy_quality_two_stage lane** — uses section-editor on LMC homepage *links*; ids do not
  collide with that surface. No action, named here so the check is visible.

## 7. Explicitly out of scope

- `tool-process-automation-scorer` (`ec2` hex-ambiguous id): route is **rename the id first,
  then the mechanical path** — not judged; the deterministic prepare would hit the same refusal.
- `tool-spawn-rate-balancer` (`chartTitle` internal duplicate): repair, then mechanical.
- The two forked-function shrink refusals (investigation owed, separate).
- `rerender-reason-producers-283` companion item (bug_historian's tracked ask, not this lane's
  to execute).
- RFC_032 (`ComponentID` unification) — sequencing note stands: it gained urgency, but the
  judged conversion does not move it.

## 8. Build checklist (next session)

1. Go: extend `fixScopeComponentInstance`'s judged-refusal result (ids-converted template +
   handler inventory); new `scope_component_instance_judged` arm (gate §4); declare it
   fan-out-intended in `component_template_writer_coverage_test.go` (cite §1); helpers
   `removeInlineHandlers` / markup-strip + tests — **mutation-tested in both directions**
   (a truncated T must refuse; a markup-edited T must refuse; the real converted shape must
   pass). Fixtures = live row bytes (the `data-*` lesson).
2. Config migration on `component-template-fixer` (numbered, snapshot to
   `agent_definitions_backup`, PREPARE-compile the new `create_section_edit_delivery` query in
   the verify block — the 460→461 lesson; `FORCE=1` on the council path filter for
   `sql_for_agents/`).
3. `scripts/pattern-check.py` run; optional-key parity check if any new spec key enters the
   registry (RFC_022 N=10 discipline — expected: none; `component_id` is reused).
4. Council: submit build on `RESUBMIT_CORR=07635a2f-3605-4e67-9a6d-7636b07f16ca`; commit with
   `Council-Submitted:` if the verdict has not landed.
5. Image: bump `IMAGE_TAG`, build from committed HEAD, verify at the DIGEST (RUNBOOK §1) —
   two same-tag traps hit this lane in two days.
6. Then §5's sequence: owed steps → canary → batch → generic pair.

---

## 9. ADDENDUM 2026-08-19 (build session) — what building it changed, and the defect the build found

The pipeline above is BUILT (Go arm, gate, workflow migration `sql_for_agents/486_…_HOLD.sql`).
Three deltas against §3–§4 as designed:

1. **The gate gained a fourth mechanical check — the binding detector** (`UnprefixedBindings`,
   `component_instance_bindings.go`): no declared id may survive bare in the script, no
   concatenated prefix unprefixed, no composition hazard. Added because reading the canary's
   REAL bytes exposed that the deterministic converter had shipped exactly that defect on 32 of
   the 69 mechanical conversions — **`bugs_open/324`**, the batch's blind spot, 14 rows serving
   broken. The converter itself gained pass 5 (same file), and the mechanical arm now reports
   what it could not place.
2. **A repair sub-programme precedes the judged sequence**: fix_type
   `repair_instance_scope_bindings` + seed `487_…_HOLD.sql` repairs the 27 mechanically
   repairable rows first (serving-broken at priority 30); its 5 refusals join the judged pool,
   so the judged queue is 25 + 5 = 30 rows, LMC still first among the LMC-affected.
3. **Refusal routing is a conditional, not an error path**: `apply_judged_write` returns
   `fixed:false` with the failing checks named (structured, kept on the state), and a
   `check_judged_result` conditional routes anything not `fixed:true` to `fail_work_item →
   needs_human_review`. An action error (DB down) takes the same road via `error_step`.

Everything else stands as designed, including the sequencing in §5 — with the repair batch
inserted before the canary, because 4 of the live-broken placements are on the judged rows'
own domains and the canary's oracle baseline must be taken on a repaired estate.