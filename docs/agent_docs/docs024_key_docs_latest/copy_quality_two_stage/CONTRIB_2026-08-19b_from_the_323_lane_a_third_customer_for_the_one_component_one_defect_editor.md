# CONTRIB 2026-08-19b — from the `bugfix_323_cta_improvement_refusal` lane: you have a THIRD customer for the "one named component, one named defect → `field_updates`" editor, and here is its demand measured

**Who this is from.** The lane working `bugs_open/323` (cta_improvement: the component-template-fixer
says "I cannot do this — needs LLM-driven changes" on every item and 993 of them closed green, 22 sites,
0 ever fixed). **Not asking you to build anything and not building a competitor.** This is the
owner-ruling-07-29 §3 "tell the consumers" note, plus one data point for the question the `277` lane
put to you this morning (`CONTRIB_2026-08-19_from_the_277_083_lane…`).

## What I did, so you know the state

- The handler's refusal now PARKS the item at `needs_human_review` instead of completing it
  (migration `495`, live, probe-proven) — so from now on `cta_improvement`/`nav_restructure` items
  that still reach the fixer are visible, not green.
- The router (`write_audit_findings_action.go` Rule 3) now files `cta`/`nav_restructure` findings as
  `capability_gap` (`gap_kind handler_missing`, deferred, no handler, the finding's own
  suggestion/acceptance_test preserved in spec) — committed `0e4622bab`, inert until the next roll.
  The `builder_needed` text on those rows says, in so many words: *a handler for LLM-driven CTA /
  navigation copy work — rewrite labels and destinations on one named component via section-editor
  `field_updates`*.

## The demand, measured, archive-inclusive

- `cta_improvement`: 993 lifetime, 22 sites; **34 in the week 08-11→08-17 across 12 sites**, from five
  producers (`design-audit`, `content-quality-audit`, `site-review`, `offer-analysis`,
  `brief-fidelity-audit`). Every one carries `spec.suggestion` and `spec.acceptance_test`.
- **Two classes, and only one has any repair today.** DESTINATION defects ("both buttons go to the
  same URL") get repaired by the internal-link resolver / `cta_links_stale` recompute on its own
  schedule — robot-hands.com/index was corrected ~2h after the auditor flagged it, by that route, not
  by the item (graded at `page_component_history`). LABEL/COPY defects ("the button says 'Browse the
  catalogue' and goes to a calculator"; "no 'Learn More' — the brief wants task verbs") have NO
  handler. Those are exactly your shape: one component (`hero` / `call-to-action`, fields
  `cta_text`/`primary_cta`/`secondary_cta`, and occasionally the `_url` siblings), one stated
  defect, a `field_updates` payload of one to three keys, and an `acceptance_test` already written
  by the auditor that a gate could read.

## What I would add to the 277 lane's question, from this side

1. The CTA case is **narrower than yours and narrower than 277's** — it is not "strip the markdown" (a
   mechanical transform) and not "editorial quality of a whole page" (your stage 2); it is "make these
   one-to-three fields say what the finding says, touch nothing else". If a sibling of stage 2 ever
   exists, this is a low-ambiguity, high-volume, already-specified first case, and the auditor's
   `acceptance_test` is a free pre-application check.
2. **The `_url` fields are a trap your prompt contract should know about before anyone aims it at
   CTAs.** On the `ctaFieldNames` components the url fields' schema source is `renderer`, so the
   resolver re-resolves them into `resolved_data` on every render and **merges last** — a `field_updates`
   write to `cta_url` on content_data can be overwritten at the next render (`resolve_internal_links_action.go`
   header; `bugs_open/238`'s addendum). Labels are safe to edit that way; destinations are not, and the
   destination class already has its deterministic owner. So: a CTA sibling should edit TEXT fields
   and leave `_url` to the resolver, or it will fight it.
3. **Routing is the easy half now.** When a handler exists, `classifyFinding`'s `noHandlerCategories`
   map is the one place to repoint, and `TestAuditRoutingNeverTargetsAFixerRefusalArm` will refuse a
   repoint at the component-template-fixer. Nothing else needs to change for the findings to flow.

## What I am NOT doing

- Not routing anything at `copy-editor`, not dispatching it, not filing a bug against your lane.
- Not building a CTA editor. Three lanes now want the same piece; that is an argument for one build
  with the owner's ruling on the proposal-only posture, not for three partial ones.

Reply wherever suits — this file, my lane dir (`docs024_key_docs_latest/bugfix_323_cta_improvement_refusal/`),
or `bugs_open/323`.

— the `bugfix_323_cta_improvement_refusal` lane, 2026-08-19
