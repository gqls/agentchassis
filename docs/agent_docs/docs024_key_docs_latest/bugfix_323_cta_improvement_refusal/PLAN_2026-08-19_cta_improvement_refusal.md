# PLAN — `bugs_open/323`, the fixer says "I cannot do this" and the item completes anyway

Design, phasing, decisions and their reasons. Corrections to the originating brief live here,
marked as corrections. Evidence and queries: `NOTES_…` / `RUNBOOK_…` beside this file.

---

## The brief, and what it turned out to be

**As filed** (`bugs_open/323`, by the 302 lane, 2026-08-19): 468 `cta_improvement` items closed
`complete` while the handler's payload says *"fix_type requires LLM-driven changes, not
programmatic"*; four fix candidates, ordered; and a stated unknown — whether the CTA work happens by
another route. It also argues that gate 1b cannot express this handler's report because `fixed` is a
boolean, and that `spacing_fix` and `cta_improvement` report "the same field, same value, opposite
meanings; only the `reason` separates them".

**As measured** (every figure in NOTES, every query in RUNBOOK):

> **CORRECTION 1 to the brief — the handler's vocabulary ALREADY separates refusal from no-op, and
> it is not the `reason` prose.** Every refusal arm in `fix_component_template_action.go` (13 sites)
> returns `action: "needs_review"`; every idempotent no-op arm returns no `action` key. Measured over
> the handler's entire history: 470 rows carry the flag (468 cta + 2 responsive "missing slot_name"
> refusals), 299 idempotent no-ops carry none, **zero overlap either way**. The discriminator exists
> and is machine-readable; what is missing is a reader. The comment at
> `fix_component_template_action.go:58` claiming the flag "stops the dispatch loop recording the
> work item as done" is false — nothing reads it (Go grep + live workflow-condition census).

> **CORRECTION 2 — "468 items" understates it; 993 lifetime, 0 ever fixed.** The other 525 carry a
> FOREIGN payload in `result` (441 webdesign-agent design-token blobs Mar–May, 12 spawn records per
> `bugs_closed/287`, ~70 triage decisions), so they are invisible to a `fixed=false` filter but are
> the same non-event. The handler has NEVER reported `fixed=true` on this type.

> **CORRECTION 3 — the "another route" question splits by defect class, and the answer is
> different for each.** DESTINATION defects ("both buttons go to X") have a live deterministic
> repair — the internal-link resolver / `cta_links_stale` recompute — that runs regardless of the
> item; robot-hands.com/index was corrected ~2h after filing by that route (graded at
> `page_component_history`). LABEL/COPY defects ("say 'Compare', not 'Learn More'") have no handler
> at all. So the honest statement is still the bug's: "closed with the handler saying it did not do
> the work" — and for the copy class, nobody else did either.

**Scope decision (stated, not silent).** The bug is a bug in TWO places and a gap in a third:
(a) the handler's refusal is unheard (completion path); (b) the producer routes a category at a
handler whose dispatch table refuses it by design (routing); (c) no handler exists for LLM-driven
CTA/nav copy work (capability). This lane fixes (a) and (b) and makes (b) unrepresentable; it does
NOT build (c) — see "What this lane does not do".

## The decision: three layers, smallest blast radius first, each an existing estate pattern

### Layer 1 — the handler honours its own refusal flag (config migration; LIVE on apply)

`component-template-fixer`'s workflow gains one conditional and one park step:

```
apply_fix → check_needs_rerender
              ├─ fixed == true      → create_rerender → compose_note → append_note → complete   (unchanged)
              └─ else               → check_refused                                            (NEW edge; was compose_note)
                                        ├─ fix_result.action == 'needs_review' → park_refused → compose_note → …
                                        └─ else                                → compose_note → …   (unchanged path)
park_refused = fail_work_item { status_override: needs_human_review, error_message: <literal naming 323>, error_step: compose_note }
```

- **Pattern, not invention:** `page-build-handler`'s `mark_needs_review` / `mark_writer_skipped`
  ("park the work item visibly instead of letting the dispatch loop stamp it complete"), and the
  283 lane's pending `486_HOLD` adds `judged_refusal` to THIS SAME AGENT with the same
  `fail_work_item → needs_human_review` shape, keyed on THIS SAME `fix_result.action` field.
- **Why the new edge hangs off `check_needs_rerender.else_step` and not `apply_fix.next_step`:**
  `486_HOLD` rewires `apply_fix.next_step` (→ `check_scope_route`) and guards on its pre-image
  value; touching it here would make 486 refuse to apply, or be clobbered by it. Neither migration
  touches the other's anchor, so they compose in either order. `[VERIFIED by reading 486: it
  references check_needs_rerender only as a destination.]`
- **Why `needs_human_review` and not `wont_fix`/`capability_gap` at the handler:** the refusal arms
  are heterogeneous — locked chrome (human must unlock), unidentifiable slot (spec defect), a
  fix_type with no capable handler. All are "a person decides next"; only the producer knows
  whether work should have been filed at all (that is Layer 2). `complete_work_item`'s guard
  preserves the status (`status NOT IN ('needs_human_review', …)`) — no change to the shared path.
- **Blast radius, measured, disconfirmable:** 470 historical rows would have parked, 0 legitimate
  completions blocked (the no-op arms never carry the flag). Covers all 13 refusal arms, not one.
- **Side effect, stated:** `complete_work_item` then skips the `result` write on a flagged row, so
  the handler's `reason` lives in the run's orchestration state and in the `doc_notes` 'fix' entry
  — `compose_note`'s prompt is amended to title refusals and quote `fix_result.reason`, so the
  reason is durably on a queryable surface.
- Migration `495_fixer_parks_refusals_as_needs_human_review.sql` (+ ROLLBACK): snapshot first,
  double-apply guard on `check_refused` presence, guards on the pre-image edge, verify block with
  `DO/RAISE` (a SELECT cannot stop a COMMIT).

### Layer 2 — the producer stops routing `cta`/`nav_restructure` at a handler that refuses them (Go; inert until roll)

`classifyFinding` Rule 3 (`componentCategories`) currently files `cta_improvement` /
`nav_restructure` at component-template-fixer. Replace with the estate's "found work I have no
handler for" record — `capability_gap`, `gap_kind = handler_missing`, `status deferred`, empty
handler, dedup `capability_gap:no_handler_for_audit_category:<cat>` (one open row per site per
category), the finding's own severity/description/suggestion/acceptance_test preserved in spec.

- **Pattern:** `bugs_closed/077` owner ruling (keep wide detection, split, queue the handler work),
  already used by this very router for unknown categories (`TestUnknownCategoryFilesACapabilityGapNotAMintedType`).
- **Why not `content_rewrite` (page-build-handler):** a whole-page regeneration for a button label is
  the wrong tool — it rewrites copy that was fine, refuses owned pages (`301`), and is the mechanism
  of `238`; and Rule 4's dedup key would swallow a CTA finding under any open content finding on the
  same page. **Why not keep routing at the fixer with Layer 1 parking each item:** that spends a
  spawn + an LLM note per finding to produce a `needs_human_review` row in a 983-row queue; the
  roadmap row is what the owner ruled this shape should be.
- **Consequence to name to the consumers (owner ruling 07-29 §3):** five auditor prompts describe
  `cta`/`nav_restructure` as "→ component fixer". They are not wrong to emit the category — the
  record is the point — but `brief-fidelity-auditor`'s routing hint becomes stale; note, do not
  churn prompts in this lane.
- Two existing tests drive `cta → cta_improvement` as the canonical "ungated producer path"
  (`TestWriteAuditFindings_UngatedProducerPathIsUnchanged`, `…DedupStillSuppresses`): re-point at
  `tone → tone_shift`; update `classifyEmittableItemTypes`.

### Layer 3 — make the bad state unrepresentable (Go test; inert until roll but bites at build)

Lift the fixer's by-design refusals out of the `switch` into a package-level
`fixTypesRefusedByDesign map[string]string` (fix_type → why), consulted by the dispatch default.
Then `TestAuditRoutingNeverTargetsAFixerRefusalArm`: drive the category universe through
`classifyFinding`; for every result whose handler is `component-template-fixer`, assert
`spec.fix_type` is not in that map. Mutation-proven: re-adding `"cta"` to `categoryToFixType` +
Rule 3 must fail it. This is the lockstep the estate already uses for routing↔roster
(`TestNoChangeRosterMatchesLiveRouting`) and routing↔existence (`handler_coverage_test.go`), one
rung further: routing↔capability.

## What this lane does NOT do, and why

- **Does not build the LLM CTA/nav copy editor.** The missing piece is "turn 'this component is
  wrong in this way' into a `field_updates` payload for `section-editor`" — identical to what the
  `277` and `301/083` lanes asked the `copy_quality_two_stage` lane for TODAY (CONTRIB 08-19). Three
  customers now. Building a new LLM write-route into live pages is an architecture-scope decision
  (and the copy-editor lane's proposal-only posture is the owner's to relax). This lane TELLS that
  lane it is a third customer and points at the 34/week demand; it does not start a competing build.
- **Does not rewrite history.** The 993 rows stay as they are; the 287-class foreign payloads are
  already documented.
- **Does not widen gate 1b** (bug candidate 3): with Layers 1–3 there is no consumer left for a
  boolean counter, and widening a shared contract for one consumer is the candidate the bug itself
  ranks last.

## Phasing

1. Diagnosis loop verdict (fired 15:55Z) — read before committing to Layer 2/3. ✅/❌ recorded in NOTES.
2. Layer 1 migration written + applied + verified (re-read live config; induce the refusal arm with
   a probe item if cheap) — config, live.
3. Layers 2+3 Go change + tests; `go test ./platform/orchestration/actions/...`; council submission;
   commit by pathspec with `Council-Submitted:`; rides the next chassis roll.
4. Tell the copy-editor lane (CONTRIB file in their dir); landmine + register + 016b §9 + WRONG_CALLS.
