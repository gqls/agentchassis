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

> ## ⚠ CORRECTED 2026-09-02, same day — the playground page is NOT the worst case, and I filed the bug before checking
>
> Prompted by the `bugs_open/443` session asking for unwritten context, I checked the sibling
> pages — which I should have done before filing. **All three pages this site built through the
> fallback tier repeat, and the two OLDER ones repeat VERBATIM:**
>
> - **`your-own-model.html` — "How it works" × 3, identical.** The £99 front door. Copy written
>   **2026-08-27** and unchanged since, so it has served that way for six days.
> - **`technical-details.html` — "The model and its licence" × 3, identical.**
> - `playground.html` — the same shape, wording varied. **The mildest of the three.**
>
> So the control went the right way and 443's root cause is stronger, not weaker — but verification
> belongs on `your-own-model`, where verbatim-identical headings cannot be read as style.
>
> ⚠ **Do NOT join this to the owner's "very AI sounding" verdict.** That verdict is dated
> **2026-08-25**; this copy postdates it. Independent facts, and the dates refuse the connection.
> The copy lane has been told directly, so a voice rewrite is not aimed at a defect it cannot reach.
>
> Scope, `[MEASURED 2026-09-02]`: **11 pages fleet-wide** repeat a component type — finetuning.uk 4,
> gaswholesalers.com 4, ai-agent-orchestration.com 3; loancash.co.uk, cookly.uk and lampenkap.com
> **zero**. Exposure is 203 pages, damage is 11. `bugs_open/443` is OWNED by the session named
> `bugs_open/443`, building a framework-wide fix through the council — **contribute to the bug file,
> do not compete.**

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

- **Do not hand-edit the repeated blocks on ANY of the three pages** (not just playground). It violates the framework ruling
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

## 443 IS FIXED — what this lane owes when the roll lands

The fix was designed, council-APPROVED (round 1, corr `b7c59309`) and committed (`dbb218a41`) by
the session named `bugs_open/443`, the same evening this bug was filed. Full account:
`bugs_open/443` §8 and `docs/agent_docs/docs024_key_docs_latest/bugfix_443_fallback_tier_subjects/`.
It adds `pages.section_subjects` and `pages.section_facts` — jsonb arrays stored beside
`pages.sections`, applied **only when aligned** (misaligned = ignored with a WARN, never guessed).

**State, all verified at the artefact 2026-09-02 — do NOT re-derive it from `schema_migrations`:**

| piece | state | how to check |
|---|---|---|
| migration `717` (the columns) | **applied** 20:11:05Z | `information_schema.columns` on `pages` |
| migration `639` (wires the handler) | **applied and LIVE** | the live `page-build-handler` row: `plan_sections.config.section_subjects` = `spec_sections.section_subjects` |
| rails commit `35905c547` | **rolled** | `section_subjects` in the chassis binary (born in that commit: 0 files at `35905c547^`, 5 at it) |
| migration `641` (writer prompt v5) | **HELD — owner read only** | not applied; gate 1 already clear |
| commit `dbb218a41` (the fallback half) | **NOT rolled** | `subjects_attached` / `facts_attached` = **0** in the binary |

⚠ **`schema_migrations` is SILENT on `_HOLD` migrations by construction** — they are applied by
hand and keep the `_HOLD` suffix for ever. I read that ledger, concluded `639` was unapplied, and
told the owner he was behind three gates when he was behind one. **Read the live
`agent_definitions` row.** (WRONG_CALLS 2026-09-02(c).)

⚠ **Probe the roll with `subjects_attached` or `facts_attached` — NOT `section_subjects`.** That
literal predates `dbb218a41` (it was born in the rails commit), so it returns a false positive for
the fallback-tier half. 641's own header carries the contaminated probe; it is correct for *its*
question and misleading for this one.

**What this lane owes, agreed with that session:** when `dbb218a41` rolls, **we run the backfill
and rebuild** for our four pages — we own the briefs, and each page's per-section subjects already
exist as prose inside its `spec.suggestion`. Write the subjects array beside the layout using the
template in that lane's RUNBOOK. **`0 rows updated` means MISALIGNED, not "already done" — fix the
array, do not proceed.**

**Two stages, and report them as two** (their framing, and it is the honest one):
- **Stage A** (post-roll, pre-`641`): proves `sections_ready[].subject` populates. **The served h2s
  will STILL repeat** — the writer prompt has not changed yet. This is success, not failure.
- **Stage B** (post-`641`): the distinct-h2s assertion, with a tier-1 page as control.

Canary: `your-own-model`, per this lane's recommendation — verbatim-identical headings make the
before/after unarguable.

**RFC_063** files the bigger question (should plan-less sites get plan tables at all) as an owner
decision, naming this site as the largest affected.

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
