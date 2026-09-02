# 443 — A page built from the FALLBACK tier cannot carry per-section subjects, so repeated component types write the same section three times

**Filed** 2026-09-02, finetuning_uk_service lane. **Status: OPEN.**
**Live damage:** `https://finetuning.uk/playground.html`, built today, serves three
`generic-text-block` sections whose `h2`s are "What you actually do in the hour",
"What you do in the hour" and "What you do in the hour". Same subject, three times, on a page
the owner will read.

## 1. Symptom

A hand-made page (a `pages` row + a `needs_content_page` item, the documented birth path for a
site with no `site_plans` row) builds cleanly — 6 sections, `ready_count=6`, `COMPLETED`, HTTP 200
— and the repeated component types in its layout all write **the same section**. The brief named
one distinct subject per section; the writer never saw them.

Served headings, `[MEASURED 2026-09-02]`:

| # | component | h2 |
|---|---|---|
| 1 | `hero` | The playground: an hour with your own model |
| 2 | `generic-text-block` | **What you actually do in the hour** |
| 3 | `generic-text-block` | **What you do in the hour** |
| 4 | `generic-text-block` | **What you do in the hour** |
| 5 | `faq` | Questions people ask before booking |
| 6 | `call-to-action` | Booking the hour starts with a conversation. |

> ## ⚠ ADDED 2026-09-02 (later) — `playground` is the WEAKEST case. Verify against `your-own-model`.
>
> I filed this from the page I had just built and did not check its siblings first. **All three
> pages this site has built through tier 3 repeat, and the two OLDER ones repeat VERBATIM:**
>
> | page | the three `generic-text-block` h2s | copy written |
> |---|---|---|
> | `your-own-model.html` | **"How it works" × 3, identical** | 2026-08-27 |
> | `technical-details.html` | **"The model and its licence" × 3, identical** | 2026-08-27 |
> | `playground.html` | "What you actually do in the hour" / "What you do in the hour" / "What you do in the hour" | 2026-09-02 |
>
> Three independent builds, three weeks apart, same tier, same 6-slot layout, same result. That is
> the control this file was filed without, and it goes the RIGHT way — but a reader should know it
> was run afterwards, not before. **`your-own-model` is the better verification target**: verbatim
> identical headings cannot be argued away as stylistic variation.
>
> `your-own-model` is the £99 front-door page and has served three identical headings since
> **2026-08-27** (`page_components.created_at`/`updated_at`, all six rows).
> ⚠ **Do NOT connect this to the owner's "very AI sounding" verdict** — that verdict is dated
> 2026-08-25 and this copy postdates it. Independent facts; the temptation to join them is strong
> and the dates refuse it.

## 2. Root cause — a deliberate tier gate, with an unintended consequence

`load_page_sections_from_spec_action.go:511-515` publishes the per-section scoping arrays
**only from the authoritative tier**:

```go
// section_subjects: same rule as section_facts — authoritative tier only,
// aligned or absent, never guessed against a fallback tier's list.
if specSource == "site_plan_tables" && len(specSectionSubjects) == len(specSections) {
    result["section_subjects"] = specSectionSubjects
}
```

`specSource` is `site_plan_tables` only for tier 1. The resolver's tiers are:

1. `site_plan_sections` for the current plan → **the only tier that can carry subjects**
2. the `site_specs.site_plan` aspect
3. `pages.sections`
4. same-role sibling layout

**The reasoning behind the gate is sound** — index-aligning a subject list against a *different*
tier's section list would be a guess, and a wrong alignment is worse than none. That is not what is
being disputed here.

**The consequence is that on a site with no `site_plans` row, per-section subjects are
structurally unreachable.** Every page on such a site resolves at tier 2, 3 or 4, so
`section_subjects` is never populated, so every repeated component type in the layout receives the
identical page-level brief. **Three `generic-text-block` slots therefore produce three attempts at
the same section — this is the predicted output, not a bad roll of the writer.**

> **ADDED 2026-09-02 (later) — the SAME gate publishes `section_facts`, and the original filing
> underplayed it.** Lines 508-510 are the `section_facts` arm and 511-515 the `section_subjects`
> arm, both conditioned on `specSource == "site_plan_tables"`. So a plan-less page cannot scope
> **facts** per section either. **A fix that carries only subjects fixes half of this.**

## 3. Why the brief did not save it

The brief in `spec.suggestion` DID name one subject per section, in order, count-matched to the
layout. It reaches the writer as page-level prose. Nothing splits it per slot, and the writer has
no per-slot scoping to bind it to — so the same page-level text is in front of it for slot 2,
slot 3 and slot 4, and it writes the most obvious section each time.

## 4. Scope — who else is exposed

Any hand-made page on any site with no current `site_plans` row, whose layout repeats a component
type. `[MEASURED 2026-09-02]` finetuning.uk has **0** `site_plans` rows (control: dartsonline.com,
robot-hands.com and loancalculator.co.uk carry 5, 5 and 4), and its three most recent pages —
`your-own-model`, `technical-details`, `playground` — all resolved at tier 3.
**⚠ This count is a census and goes stale by ADDITION: re-run before quoting.**

**The fleet census is now done** `[MEASURED 2026-09-02]`. **25 of 59 sites have no current
`site_plans` row — but 19 of those are `pool-*.internal` rows with ZERO deployed pages**, and
counting them would triple the apparent blast radius. The real exposure is **6 sites, 186 deployed
pages**:

| site | deployed pages |
|---|---|
| finetuning.uk | 52 |
| ai-agent-orchestration.com | 44 |
| gaswholesalers.com | 32 |
| loancash.co.uk | 30 |
| cookly.uk | 15 |
| lampenkap.com | 13 |

> **CORRECTED 2026-09-02, same day — the 186 is too narrow and the better figure is 203.**
> `build_status='deployed'` (my predicate) is a strict subset of `deployed_at IS NOT NULL` (the
> `bugs_open/114` lane's, derived independently). The 17-page gap is entirely
> **`build_status='needs_rebuild'`** — gaswholesalers.com 9, finetuning.uk 5,
> ai-agent-orchestration.com 3: pages that HAVE deployed and are flagged to rebuild. For any
> defect that bites *at render*, those 17 are the highest-value pages in the cohort, not a
> rounding error — and my predicate silently dropped exactly them. **Use 203.** The lesson
> generalises past this bug: `build_status='deployed'` answers "in the deployed state now",
> `deployed_at IS NOT NULL` answers "has ever deployed", and a rebuild-pending page is invisible
> to the first while being the most likely thing to re-render next.

**CENSUSED 2026-09-02 (later) — and it DEFLATES this bug, which is worth saying plainly.**
Exposure is 203 pages; **actual damage is 11.** Pages whose layout repeats a component type:

| site | pages repeating a type | deployed pages |
|---|---|---|
| finetuning.uk | 4 | 53 |
| gaswholesalers.com | 4 | 31 |
| ai-agent-orchestration.com | 3 | 40 |
| loancash.co.uk | **0** | 26 |
| cookly.uk | **0** | 9 |
| lampenkap.com | **0** | 7 |

So three sites carry all of it and three carry none. **Size the fix against 11 pages, not 203** —
though note the fix is framework-wide either way, because the 192 unaffected pages are unaffected
only for as long as nobody adds a repeated slot to one.

⚠ **The obvious census query drops pages SILENTLY, and I nearly published the wrong number a
second time.** `CROSS JOIN LATERAL (SELECT count(*) FROM jsonb_array_elements_text(p.sections)
GROUP BY …)` yields no rows for an EMPTY `sections` array, so the page disappears from the result
instead of counting as zero. **37 of the 203 have empty `pages.sections`** and vanished that way —
the totals above sum to 166, not 203, and that gap IS the artefact. Use `LEFT JOIN LATERAL` or
`COALESCE(jsonb_array_length(...),0)`.
Those 37 are independently worth someone's attention: an empty-`sections` page is exactly what
no-ops at `mark_no_ready_sections` if anything ever rebuilds it (see the LANDMINES entry
"A hand-made page whose `sections` is `[]`…").

Still not censused: **whether all 11 are actually SERVING repeated headings.** Repeating a
component type is the necessary condition; three of the 11 are confirmed serving duplicates
(finetuning.uk's three, read off the live pages). The other 8 — gaswholesalers.com 4,
ai-agent-orchestration.com 3, finetuning.uk 1 — have the layout but nobody has read their served
headings. **Read them at the served page, not from `rendered_html`.** That is the one query left,
and it is the difference between 11 pages damaged and 3.

⚠ **The same tier boundary bites a second, unrelated mechanism.** `site_plan_imagery.plan_id` hangs
off `site_plans` too, so the `bugs_open/114` lane's route-1 hero delivery (IMG-078) cannot reach
these 6 sites either — reported to that lane 2026-09-02. **The plan tables are becoming the tier
where capability lives, and 6 real sites are not in them.** A fix for 443 that only widens the
subject gate leaves the general problem standing; treat the shared geometry as the more important
finding.

## 5. Fix candidates, ordered by what closes the door

1. **Let the fallback tiers carry subjects when the alignment is a FACT, not a guess.** The gate
   exists because alignment across tiers is unverifiable — but a subject list written *against*
   `pages.sections`, stored beside it and validated as the same length, is aligned by
   construction. A `pages.section_subjects` column (or a `content_direction.section_subjects` key)
   checked `len(...) == len(specSections)` on the *same* tier makes the bad state unrepresentable
   rather than merely unlikely.
2. **Make the writer aware of its siblings.** Pass the already-written sections of the same page
   into each subsequent slot's context, so slot 3 can see that slot 2 already covered the subject.
   Fixes the symptom for every cause, including ones not yet found, but costs context per section.
3. **Refuse the layout instead of the subjects.** Have `plan_sections` warn (or defer) when a
   layout repeats a component type and no `section_subjects` are available — it cannot write the
   page correctly, and today it does so silently.
4. Weakest: **operators must avoid repeating component types on plan-less sites.** Listed only to
   be rejected — "operators must remember X" is a defect, not a fix.

## 6. Verification for the fixing thread

Rebuild a plan-less page whose layout repeats a component type and assert the served `h2`s are
**distinct**. The control that makes it a real test: the same assertion against a tier-1 page,
which must pass both before and after.

## 7. Diagnosis provenance — first-hand, declared

**No `090` run.** Per the 2026-07-31 ruling this file states plainly what was substituted: the
deciding arm was read directly (`load_page_sections_from_spec_action.go:511-515`, the
`specSource == "site_plan_tables"` condition — not a grep hit, the condition itself); the page's
actual resolved tier was confirmed by measuring all three tiers for it and for two control pages
that read differently (`services`/`about` return tier2=1, `playground`/`your-own-model` return
0/0/6); and the predicted symptom was then read off the **served** page, not the DB. The claim is
cross-cutting, so a `090` run remains worth its cost to a later thread — it would be a genuine
independent check, not a formality.
