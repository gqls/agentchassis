# HANDOFF 2026-09-02 — the playground page is LIVE, and building it found a structural defect in how plan-less sites scope their sections. Start here.

**COLD-START for the finetuning.uk service lane.** Supersedes
`HANDOFF_2026-08-30_continue_here.md` (still the reference for the copy handover, the hero-image
delivery and the bug list — nothing in it is retracted). Technical log:
`NOTES_finetuning_uk_service.md`, entry 2026-09-02. Owner prose: `README_where_we_are.md`.

## What changed today

**The playground booking page is built and serving** — owner decision 4, the last of the seven
that was still unbuilt. `https://finetuning.uk/playground.html`, HTTP 200 (verified with an
invented-URL control that 404s, so the 200 is not a parked-domain artefact), 37 KB, six sections:
hero · 3 × generic-text-block · faq · call-to-action. It cites only registered facts —
`ft-booking-hours`, `ft-playground-hour`, `ft-deletion-window`, `ft-retention-default`,
`ft-price-99` — all five current in `evidence_base`.

**It has a visible defect and the owner has been told:** the three text blocks all wrote the same
section. Served `h2`s are "What you actually do in the hour", "What you do in the hour", "What you
do in the hour". **Do not hand-edit them** — see below.

## ⚠ THE THING THAT WILL COST YOU TIME IF YOU DO NOT READ IT

**This site has NO `site_plans` row, and that is now known to disable two separate mechanisms.**

1. **Section scoping.** `load_page_sections_from_spec_action.go:511-515` publishes
   `section_subjects` **only** when `specSource == "site_plan_tables"` (tier 1). This site resolves
   at tier 3 (`pages.sections`). So per-section subjects are structurally unreachable, every
   repeated component type gets the identical page-level brief, and it writes the same section
   each time. **This is why the page repeats — not a bad roll of the writer, and not something a
   better brief can fix.** Filed as `bugs_open/443` with fix candidates ordered.
2. **Hero imagery.** `site_plan_imagery.plan_id` hangs off the same table, so the `bugs_open/114`
   lane's route-1 hero delivery (IMG-078) cannot reach this site either. Reported to that lane.

`[MEASURED 2026-09-02]` **6 real sites / 186 deployed pages** have no current plan row —
finetuning.uk (52), ai-agent-orchestration.com (44), gaswholesalers.com (32), loancash.co.uk (30),
cookly.uk (15), lampenkap.com (13). ⚠ **25 of 59 sites screen as plan-less but 19 are
`pool-*.internal` with zero deployed pages** — count those and you triple the blast radius.

**The general finding matters more than either symptom: the plan tables are becoming the tier where
capability lives, and six real sites are not in them.**

## Do NOT do this

- **Do not hand-edit the three repeated blocks.** It violates the framework ruling
  (2026-08-04, no hand-authored content), and it papers over 443 on the one page that demonstrates
  it. The page is factually right and structurally sound; the repetition is a platform defect with
  a bug number.
- **Do not re-dispatch the build hoping for a better roll.** The subjects cannot reach the writer
  through tier 3. Same input, same output.
- **Do not give this site a `site_plans` row just to fix one page.** That changes the birth path
  for all 52 deployed pages. If it is the right fix it is 443's fix candidate 1's territory and
  belongs in a reviewed change, not a lane patch.

## Building a page on this site — the recipe that actually works

`pages` row at `build_status='planned'` + a `needs_content_page` item for `page-build-handler`
(status `triaged`, priority 40, item_key `gap_plan_new_<page>_<site_id>`; brief in
`spec.suggestion`). **And the part that is easy to miss:**

**Put the layout in `pages.sections`. Not in the work item's spec.** A `sections` array on the
item is read by nothing; the build completes at `mark_no_ready_sections` with 0 components, a 404,
and an EMPTY `agent_error_log`. Nothing errors, so there is no failure to find. Before dispatching,
assert the layout is readable from a tier that will serve — the three-tier query is in
`LANDMINES.md` under "A hand-made page whose `sections` is `[]`…", with the instruction to run it
against an already-built page as a control.

**And expect to wait.** The dispatcher takes ONE site per tick ordered by each site's oldest
pending item. Today's build sat ~27 minutes behind four other sites. That is fairness working, not
a defect — measure your POSITION, not your eligibility, and do not include the selector's
`status='claimed'` exclusion when you do (it hides exactly the sites ahead of you). Both traps are
in `LANDMINES.md`; the false "stale claim is blocking the site" conclusion is in `WRONG_CALLS.md`.

## Next session, in order

1. **`bugs_open/443`** — the one query left is how many of those 186 pages actually repeat a
   component type. That converts exposure into damage and sizes the fix. Consider a `090` run: the
   claim is cross-cutting and today's filing declared first-hand verification in place of one.
2. **Stripe — his, and last.** Unchanged. All seven owner decisions are now built except this.
3. **`bugs_open/398`** — stays OPEN until robot-hands.com and gaswholesalers.com serve
   `--color-cta-bg-ink`. Fix is committed and inert until the chassis roll; probe the binary with a
   present- and an absent-control on the day.
4. **Optional, owner's call:** the datasets page. Six datasets are built
   (`datasets/`, `PROVENANCE.md`); the page that presents them was deliberately not built while
   the copy register was unresolved. Copy quality now sits with `copy_quality_two_stage`.

## Still true from 2026-08-30

Terms and privacy pages published and locked · the 9 hero images delivered and wired · copy quality
handed to `copy_quality_two_stage` · the three voice datasets remain small (26/13/16) and would
benefit from more owner emails and articles if he offers them.
