# 012 — Explicit per-page redesign intent (`recompose_pages` in the re-plan spec)

**Filed:** 2026-07-22, owner-approved follow-on to `/bugs_closed/037`. **Class:** planner
capability. **Status:** code **LIVE on v1.0.1149**, **VERIFIED END-TO-END** on the dartsonline test
site (same-page A/B proof below), and **COUNCIL-APPROVED** (advisory, retrospective, 2026-07-24 —
`Council-Reviewed: 3d025222-31a9-47a3-982e-197a77cce002`, 3 advisory objections none high-severity).
Remaining open: operator ergonomics, and the drop-loud follow-up the council recommended (below) —
neither blocks use.

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
3. ~~Live verification once rolled.~~ **DONE — proven live on dartsonline, see below.**
4. **Drop semantics — now with a council recommendation (upgraded 2026-07-24).** A recompose page the
   LLM then omits is dropped from the plan with no signal. The council's `bug_historian` seat
   (medium) rightly flags this as a miniature of the very silent-loss class 037 closed. **Recommended
   next step:** make the drop *loud* — after convergence, if a `recompose_pages` name is absent from
   the final plan, raise a `site_work_item` (or at minimum an error-level log distinguishable from
   ordinary absence). Optionally also the "redesign but never delete" variant (union the page back
   with empty sections). This is a small follow-up code change (+ image roll); **owner call on whether
   to build it now.** Until then, a caller should only name pages they expect the LLM to re-propose.

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

## VERIFIED LIVE — 2026-07-22, dartsonline.com (a clean same-page A/B)

Two `needs_site_plan` re-plans on the dartsonline test site, on chassis v1.0.1149:

- **Run 1** (`recompose_pages:["contact"]`) proved the plumbing — the orchestration's
  `input_data.spec` came through as `{"recompose_pages":["contact"]}` — and proved the guard still
  protects **unnamed** pages: `index` and `shipping-returns` were **preserved** (kept their realised
  composition) even though the LLM proposed *different* compositions for them. It was inconclusive on
  the release itself only because the LLM coincidentally re-proposed `contact`'s exact realised
  composition (the coin-flip `/bugs_closed/037` warned about), and the `validate_plan` step's
  ephemeral pod logs were already gone.

- **Run 2** (`recompose_pages:["index","shipping-returns"]`) closed it. The **same two pages** that
  were preserved-when-unnamed in run 1 were **released-when-named** in run 2 — they took the LLM's
  divergent composition, while the three control pages held:

  | page | run 1 (unnamed) | run 2 (named) |
  |---|---|---|
  | `index` | PRESERVED `[hero, product-grid, category-listing, features, cta, testimonials]` | **RELEASED** → `[hero, product-grid, info-card-grid, category-listing, cta, testimonials]` (LLM: dropped `features`, added `info-card-grid`) |
  | `shipping-returns` | PRESERVED `[generic-text-block]` | **RELEASED** → `[hero, generic-text-block, faq]` (LLM added `hero`+`faq`) |
  | `about` / `new-arrivals` / `contact` | preserved | preserved (unchanged) |

  Same page, opposite outcome, one variable (membership in `recompose_pages`) — the genericity-proof
  shape `/bugs_closed/001` used. Corroborated by run 2's own `site_plan_sections` (plan `0fb05b75`,
  `is_current`). Runs' spawned rebuilds were cancelled/left to settle; `index` and `shipping-returns`
  on the dartsonline TEST site now carry their recomposed layouts (restorable from the pre-run values
  recorded in the workstream NOTES if ever wanted).

## COUNCIL REVIEW — APPROVED, 2026-07-24 (advisory, retrospective)

`Council-Reviewed: 3d025222-31a9-47a3-982e-197a77cce002`. Submitted the combined 037+012 change
(one coherent task, same function). Verdict **APPROVED** — 13 seats, 5 abstained, **3 advisory
objections, none high-severity**; every guardian (incl. the hard-veto seat) approved. Reviewers
praised the separate-predicate design ("mirrors the lesson embedded in bugs_closed/050 itself") and
the reuse discipline (ExtractNestedField(input_data.spec) reuse "is the STEP ZERO discipline working
as intended"). (First run wedged on a bug-003 spawn-loss at `gate_bug_historian`; resubmitted, flowed
through in ~10 min.)

**Advisory objections + responses:**

- **editquality (medium) — "does swapping the predicate drop adoption-locked pages?"** No. The two
  membership sites are `if noCurrentPlan || realisedPageCompositionIsPreserved(rm)` — the
  adoption-locked term (now `noCurrentPlanFlag`, post-051 rename) is a *separate* OR'd term and was
  never folded into `realisedPageIsBuilt`. Swapping only the second term leaves adoption-locked
  coverage intact. No regression. (The seat flagged it because edit 1's sketch showed the predicate in
  isolation; the call-site line carries the `||` term.)
- **prior_art_librarian (medium) — "only 2 of 4 needs_rebuild setters were evidenced."** The other two
  also preserve `pages.sections`: `UpdatePageStatusAction` (`v3_site_actions.go:644` — refused
  0-component/partial deploy: sets `needs_rebuild`, clears `built_from_plan_version`, **keeps
  sections**) and `flagPagesForRebuild` (`maintenance_actions.go` — image/maintenance rebuild, **keeps
  sections**). All four preserve sections; the universal claim holds. (I checked all four when filing
  037 — see `/bugs_closed/037`; the submission only quoted the two clearest.)
- **bug_historian (medium) — FOLLOW-UP worth tracking.** The recompose escape hatch reopens, in
  miniature, the silent-loss class 037 closed: a recompose-named page the LLM omits is dropped with no
  signal. Recommendation: after convergence, if a `recompose_pages` name is absent from the final
  plan, raise a `site_work_item` (or at least an error-level log) so the drop is *loud*. This sharpens
  the parked drop-vs-keep choice below into a concrete, better option: **keep-or-surface, don't
  silently drop.** → tracked as open item 4 (upgraded).
- **guardian / bug_historian (low):** a `recompose_pages` name matching no realised page silently
  no-ops — worth a name-not-found log/metric; and confirm no *other* convergence path reads
  `build_status` for preserve-vs-recompose (only `reconcilePlanWithRealised` + its truncation do;
  `decideEmit` reads it for emit/skip, not preservation). Minor diligence, noted.

None unwind the live change; all are advisory improvements. The medium follow-up (loud-signal on a
recompose drop) is the one genuinely worth doing next.

## Grounded in

- `/bugs_closed/037` — the guard this opts out of; its "redesign route" section names this feature.
- `/bugs_closed/001` fix step 4 — the deferred explicit-intent design this implements.
- `/bugs_closed/050` — why "empty the sections" is not a viable redesign route for a deployed page.
- Spec-flow map (2026-07-22): dispatch-loop `call_handler` → `input_data.spec`; `ExtractNestedField`
  supports the dotted path `input_data.spec.recompose_pages`.
