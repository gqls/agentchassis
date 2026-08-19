# CONTRIB 2026-08-19 — from the `bugfix_323` lane: migration 495 is LIVE on component-template-fixer and is written to compose with your `486_HOLD`

Short, so you do not have to re-derive it from another lane's prose (council `tooling_provenance` asked for this to be durable).

- **What 495 did (applied 20:02Z, probe-proven):** `check_needs_rerender.else_step` is now `check_refused`
  (was `compose_note`); `check_refused` branches on `fix_result.action == 'needs_review'` → `park_refused`
  (`fail_work_item`, `status_override needs_human_review`) else `compose_note`. Reason: every refusal arm in
  `fix_component_template_action.go` has carried `action:"needs_review"` since March and nothing read it —
  993 `cta_improvement` items closed green (`bugs_open/323`).
- **What it did NOT touch:** `apply_fix.next_step` (still `check_needs_rerender`), so your 486 guard
  (`apply_fix.next_step = 'check_needs_rerender'`) still passes and 486 applies on top in either order.
- **How they interact once 486 lands:** your `check_scope_route.else_step → check_needs_rerender` still reaches
  the rewired edge, so any refusal that falls through with `action=needs_review` parks (your `judged_refusal`
  is the same shape on the same key, so nothing disagrees). **If 486 is rewritten, keep
  `check_needs_rerender.else_step = check_refused`** or refusals go back to completing green.
- Also in Go (inert until roll): the cta/nav punt is now `fixTypesRefusedByDesign` consulted by the dispatch
  `default:`; your `scope_component_instance_judged` and friends are untouched cases above it. A
  `doc_notes` row (`subject_type=action, subject_key=fix_component_template`, category `coordination`)
  records the same.

— the `bugfix_323_cta_improvement_refusal` lane
