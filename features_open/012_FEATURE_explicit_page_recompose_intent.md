# 012 — Explicit per-page redesign intent (`recompose_pages` in the re-plan spec)

**Filed:** 2026-07-22, owner-approved follow-on to `/bugs_closed/037`. **Class:** planner
capability. **Status:** code **LIVE on v1.0.1149** (2026-07-22 — commit `385eb0b26`; verified in the
running pod: `strings /app/agent-chassis | grep -c recomposePagesFromSpec` = 2). Remaining open:
operator ergonomics (a friendlier way to set `recompose_pages`), an optional end-to-end live check,
and the drop-vs-keep design choice below.

> **Note on the spec-read path.** The one link unit tests can't cover is the live
> `input_data.spec.recompose_pages` extraction. It uses the SAME accessor an existing production
> action already relies on (`update_site_spec_from_item_action.go:74` reads `input_data.spec`), so the
> plumbing is proven-in-use, not new. A full live re-plan would only re-confirm known-good plumbing —
> hence it's optional, not a blocker.

## Why

`/bugs_closed/037` closed a defect: a re-plan silently discarded a `needs_rebuild` page's built
composition. The fix preserves the composition of every `deployed` **and** `needs_rebuild` page. That
is correct, but it removes the (accidental, coin-flip) route by which a page used to get redesigned —
so a **deliberate** single-page redesign needs an explicit signal. This is bug 001's long-deferred
"fix step 4": *"gate a deliberate rebuild behind explicit intent (a per-page `rebuild:true` in the
`needs_site_plan` spec)."* The owner ruled (2026-07-22) to build it.

Without it, the only way to redesign a preserved page is to hand-empty its `pages.sections` and
re-plan — a fiddly, undocumented dance, and for a `deployed` page it collides with `/bugs_closed/050`
(deployed + empty sections is treated as "rendered elsewhere", so the LLM is *forbidden* from
composing it). `recompose_pages` is the clean, explicit alternative.

## What was built (chassis code — committed `385eb0b26`)

`ValidateSitePlanAction` (`platform/orchestration/actions/v3_site_actions.go`) now reads an optional
list from the `needs_site_plan` trigger spec:

```
input_data.spec.recompose_pages = ["index", "contact"]
```

and **pre-filters** those named realised pages out of `existingPages` *before* convergence. Effect:
`reconcilePlanWithRealised` (and the truncation must-keep, which reads the same slice) treat a named
page as **from-scratch** — the LLM's proposed composition governs it; it may be redesigned or, if the
LLM omits it, dropped. Unnamed pages keep the full `/bugs_closed/037` protection.

Design notes:
- **No signature change** to `reconcilePlanWithRealised` (a hotly-contended function). The whole
  behaviour is one pre-filter at the call site plus two small helpers
  (`recomposePagesFromSpec`, `filterOutRecomposePages`).
- **No workflow/DB change needed.** The dispatch loop already forwards the entire work-item spec to
  `input_data.spec` (confirmed: `build-dispatch-loop` `call_handler` maps `spec:current_item.spec`;
  `update_site_spec_from_item_action.go:74` is the existing precedent for reading `input_data.spec`).
  So a `recompose_pages` key placed in a `needs_site_plan` item's `spec` JSONB arrives unmodified.
- **Explicit intent overrides `/bugs_closed/050`'s silent-injection guard** deliberately: a recompose
  page is removed before Pass B/B2 run, so an owner who names a sectionless tool page *can* have it
  composed. 050 protects against *silent* injection; this is the opposite — named, intended.
- Absent field ⇒ `nil` ⇒ every ordinary re-plan is byte-for-byte unchanged.

Tests (`v3_site_reconcile_test.go`): `TestRecompose_ReadsSpecList`,
`TestRecompose_FilterReleasesOnlyNamedPages`,
`TestRecompose_EndToEnd_NamedPageIsRedesignedPeerIsPreserved` (a named page is redesigned to the LLM's
composition while an unnamed peer is snapped back). Isolated-worktree build confirms HEAD + these
files compiles clean.

## What is still OPEN

1. ~~It is inert until the chassis image rolls.~~ **DONE — live on v1.0.1149 (2026-07-22), symbol
   verified in the running pod.**
2. **Operator ergonomics — how you actually set it.** Today you would emit a `needs_site_plan` item
   with `spec = '{"recompose_pages":["index"]}'` by hand (see the RUNBOOK). A nicer trigger (a small
   script, or an admin-dashboard action) is optional polish, not required for the capability to work.
3. **Live verification** once rolled: emit a re-plan with `recompose_pages:["<a deployed page>"]`,
   assert that page takes the LLM's new composition while its peers keep theirs.
4. **Drop semantics.** A recompose page the LLM then omits is dropped from the plan. That is the
   honest meaning of "recompose from scratch", but if the owner wants "redesign but never delete", a
   follow-up could union such a page back with empty sections instead. Recorded, not built.

## How to use it (once live)

```sql
-- redesign just the homepage of <site>, preserving everything else
INSERT INTO site_work_items (id, site_id, item_type, status, spec, handler_agent, created_at)
SELECT gen_random_uuid(), s.id, 'needs_site_plan', 'detected',
       '{"recompose_pages":["index"]}'::jsonb, 'build-site-planner', NOW()
FROM sites s WHERE s.domain = '<domain>';
```
(Shape mirrors `docs/.../idea_uk_vm_site/sql/p1_01_replan_emit.sql`; the only addition is the
`recompose_pages` key in `spec`.)

## Grounded in

- `/bugs_closed/037` — the guard this opts out of; its "redesign route" section names this feature.
- `/bugs_closed/001` fix step 4 — the deferred explicit-intent design this implements.
- `/bugs_closed/050` — why "empty the sections" is not a viable redesign route for a deployed page.
- Spec-flow map (2026-07-22): dispatch-loop `call_handler` → `input_data.spec`; `ExtractNestedField`
  supports the dotted path `input_data.spec.recompose_pages`.
