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

Still not censused: **how many of those 186 pages actually repeat a component type** — that is what
converts exposure into damage, and it is the one query left. A page whose layout has no repeated
type is exposed but unharmed.

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
