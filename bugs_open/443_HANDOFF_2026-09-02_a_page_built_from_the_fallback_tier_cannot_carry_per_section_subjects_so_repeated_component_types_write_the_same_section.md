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

**ANSWERED for finetuning.uk, 2026-09-02 (later): 4 of 4 serve repeats — the necessary condition
was also sufficient in every case tested.** The fourth page, `our-position-on-ai`, is the
informative one because it breaks the shape of the other three:

| page | layout | served |
|---|---|---|
| `your-own-model` | 3 × `generic-text-block`, **adjacent** | "How it works" × 3 |
| `technical-details` | 3 × `generic-text-block`, **adjacent** | "The model and its licence" × 3 |
| `playground` | 3 × `generic-text-block`, **adjacent** | 3 near-identical variants |
| `our-position-on-ai` | **2 ×**, **NON-adjacent** (separated by a `features` block) | **"Our Honest Position on AI" × 2** |

So **neither adjacency nor a count of three is required**: two blocks with another component
between them still collide, and on that page they also duplicate the page's own title. The
trigger is simply "the same component type appears more than once with nothing to tell the
instances apart". That makes the census's necessary condition a good proxy for damage, and it
means a fix must scope EVERY instance of a repeated type, not just consecutive runs.

**Mechanism, from the `copy_quality_two_stage` lane (2026-09-02, their NOTES `c345b144c`):** this
is the same mechanism as their empty-room/title-promise family — *the writer fills an
underspecified slot fluently from its prior*. Here the slots are not underspecified but
**identically specified**, N times, which is why the output is deterministic rather than merely
likely. Their conclusion, independently reached: the fix belongs at the tier that publishes
per-section subjects, **not in any prompt**.

**Interaction bound, verified rather than accepted on report:** that lane's register measurements
for finetuning.uk (canaries `approach`, `careers`, `case-studies`, `contact`, `use-cases` and the
two tool pages) are unaffected — none is among the four above, checked against the layout census.
So their published register findings for this site stand without a 443 asterisk.
⚠ **But the parked stage-2 copy proposal `8003c51a` sits ON `your-own-model`**, which does carry
443. Whoever eventually reviews that proposal must know the page repeats underneath it, or they
will grade a rewrite against a defect the rewrite cannot reach.

Still not censused: **whether the other 7 (gaswholesalers.com 4, ai-agent-orchestration.com 3) are
actually SERVING repeated headings.** Repeating a
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

## 8. Fix (2026-09-02, session "bugs_open/443", lane `bugfix_443_fallback_tier_subjects`)

**Candidate 1 chosen and implemented, extended per §5's follow-ups: BOTH arrays (subjects and
facts share the gate), and instance-scoped page-wide (non-adjacent repeats fire too, §4a).**
Candidate 3's observe-only half is included; candidate 2 (sibling-aware writer) is not pursued
(context cost, and two lanes independently concluded the fix belongs at the subject-publishing
tier — copy_quality_two_stage's framing: these slots are not underspecified but *identically
specified N times*, which is why the output is deterministic). The convergence alternative
(give the 6 sites plan rows) is deliberately NOT this fix: `reconcile_site_plan` mints rebuild
items against plan pages, so it is a programme touching 203 pages' birth path — filed as
**RFC_063** for the owner. This fix is correct under either RFC outcome.

What shipped (commit alongside this section; council corr `b7c59309`, submitted pre-verdict):

- **Migration 717**: `pages.section_subjects` + `pages.section_facts`, nullable jsonb, aligned
  by index with `pages.sections`. Column-only, apply before the roll (638's own order).
- **Loader**: tier 3 reads the two columns in the same statement as sections (content-equality
  guard when collected_data served the list — a same-length different list must not pass);
  tier 2 reads the same-object sibling keys the normalise pass emits, RAW-index across skips;
  tier 4 structurally never; LOCK-008 merge nil-inserts for every source and re-aligns the
  stored tier-3 array after a merge; misaligned stored arrays are ignored with a WARN — kept,
  inert, visible, never guessed, never auto-deleted.
- **Detector**: `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT` (registered, human-evidence) files
  one durable row per repeated-type-without-subject at build time, before any planning work.
  Bound: **25** repeat-layout pages fleet-wide as of 2026-09-02.
- **Tests**: 7 guard-arm tests + a pure-function table; **3 mutations caught** (equality→length,
  merge nil-insert removed, realign removed — each failed its named test).

**Measurement closing §4's open question** `[MEASURED 2026-09-02]`: the last 7 unread pages
(gaswholesalers 4, ai-agent-orchestration 3) were curled with per-domain invented-URL controls
(both 404). **All 7 repeat — so 11 of 11 census-flagged pages serve real repetition, ≥8 with
verbatim-identical h2 pairs.** The layout census is a confirmed damage proxy, 11/11.

**Dependency, stated plainly:** the writer prompt is v4; seed 641 (v5, renders the subject) is
owner-read gated and NOT applied, so subjects are stamped on `sections_ready[].subject` and
writer-inert — for tier 1 exactly as for these tiers. §6's served-h2 verification is therefore
**Stage B, post-641**; Stage A (post-roll, pre-641) asserts `sections_ready[].subject` is
populated on a rebuilt tier-3 page and the detector goes quiet on it (demand control: a
subjectless repeat page must still fire it). Canary: `your-own-model` (finetuning lane's
choice, §4b), subjects backfilled from its brief via the lane RUNBOOK template; playground
stays untouched as the demonstrating case.

**Who was told** (2026-07-29 §3): apis.uk (PBP-049 owner — the "authoritative tier only" line
in their entry is superseded by PBP-051), finetuning lane (canary + recipe gains the two
columns), copy_quality_two_stage (their subjects-precondition gains a second source),
bugs_open/114 lane (tier geometry shared; RFC_063 names IMG-078).

Working docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_443_fallback_tier_subjects/`.

## 9. Stage A: PROVEN at plan time on a live build (2026-09-03, first-hand)

Canary `technical-details` (finetuning.uk — the verbatim "The model and its licence" ×3 page),
item `896bb245` (spec verified mode-less by both lanes ⇒ full regeneration), dispatched via
page-build-handler (the only workflow containing `load_page_sections_from_spec`), claimed
10:34Z on v1.0.1356. Read directly from orchestration row
`28610ba3-1d53-4bff-b1da-6f386a504c2f` (corr `6e8eadaa`), not from a report
`[MEASURED 2026-09-03]`:

- `spec_sections`: source **`pages_table`**, count 6, `section_subjects` length **6** — the
  fallback tier attached the backfilled arrays. **The state §2 called structurally
  unreachable is populated on a real build.**
- `sections_ready[]`: all six slots carry subjects; the three `generic-text-block` slots
  carry three DISTINCT ones ("exactly what the £99 fine-tune contains" / "how the training
  works and what we handle" / "what a technical reader asks", with hero/faq/cta scoped too).
- Detector: **0** rows for this page post-dispatch, against a fleet demand control of **7**
  rows since the roll (it demonstrably fires where subjects are absent — quiet here means
  covered, not blind).

Served h2s still repeat, as predicted: the writer is v4 until the redrafted 641 lands
(§8's stated dependency — that is Stage B, unchanged). Remaining Stage A confirmation
(writer-row carry at run completion) is on the finetuning lane's watch.

**§9 completion (2026-09-03 10:40Z, finetuning lane's watch): STAGE A CLOSED.** Writer row
`ce514ce0` (COMPLETED 10:38:58Z) carries `sections_for_render.sections_ready[].subject` on
all six slots, identical to the handler's plan; page redeployed 10:40:16Z, HTTP 200. Served
h2s still repeat as predicted — the three text blocks wrote model-and-licence variants with
varied wording despite each having its own distinct subject in the data. **That is the clean
split the staging existed to show: the subject reaches the writer's DATA and not yet its
PROMPT. Stage B is exactly 641 and nothing else.** Reported to the owner as Stage A.
⚠ For Stage B reads: both finetuning items ran under ONE correlation (`6e8eadaa` — the
build-dispatch-loop takes a site's items in one loop), so key per-page reads on the writer's
`orchestration_id`, never the correlation. 641's applier is now the
`framework_prompts_positive_voice` lane, per the owner.

---

## CONTRIB 2026-09-03, `dartsonline_traffic` lane — a SECOND, DISJOINT population reaches your fallback tier, and it is larger than the one you censused

**Not a competing diagnosis and not a correction — your census is right for the population it names.** This is a different door into the same room, found while chasing why four of my site's blog pages had no plan rows. **Take or leave it; I am not working your fix.**

**Your population is whole SITES with no current plan** — *"25 of 59 sites have no current plan… 203 deployed pages"*, which your damage criterion then narrows to 11. **Mine is pages on sites that DO have a substantive current plan, whose page TYPE is nonetheless never written into it.** dartsonline is a planned site — 55 `site_plan_sections` rows across 18 page names — and **25 of its 42 built pages still have no plan row**.

**Fleet census across the 32 sites that have a substantive current plan** `[MEASURED 2026-09-03 — a census, stale by ADDITION, re-run before quoting]`:

| page_type | built | in plan | absent | % |
|---|---|---|---|---|
| `tool` | 290 | 49 | **241** | 83% |
| `blog-post` | 241 | 40 | **201** | 83% |
| `guide` | 95 | 24 | **71** | 75% |
| `content` | 126 | 103 | 23 | 18% |
| `section-index` | 52 | 45 | 7 | 14% |
| `landing` | 45 | 44 | **1** | **2%** |

**The split is clean and it is by page type, not by site health:** structural pages are essentially always planned, content pages essentially never. So a site can look fully planned by any site-level check and still send every article it owns down the fallback path.

**The mechanism, and it is not a defect in any one route — it is that no route does it.** ⚠ **Corrected 2026-09-03, after reading the file: I first cited this from a five-week-old doc comment without opening it, and the line number had drifted.** First-hand in `platform/orchestration/actions/create_blog_posts_action.go`: `sections := post.Sections; if len(sections) == 0 { sections = []string{"hero","article-body","call-to-action"} }` at **:212** (the `SQL_2026-07-29d` header cites `:183`, correct when written) — that list is marshalled into `INSERT INTO pages (… sections …)`, and **the string `site_plan_sections` does not occur anywhere in the file.** So the triple is a **fallback** when the caller supplies no `post.Sections`, and the invariant is not the layout but the **write target**: the action writes `pages.sections` (tier 3, the *materialised cache*) and never `site_plan_sections` (tier 1, *the authority*). **Consequently "pass richer sections" is not a fix — it changes composition and leaves the plan empty.** My own site's nine exceptions are not the framework working: they are a **hand-written backfill this lane applied once**, all nine rows sharing a single timestamp `2026-07-29 13:28:03Z`. That seed's header states the condition in the platform's own words — *"Nobody ever decided what blocks a guide page should contain."*

**Why I think this is worth your attention specifically:** if your Stage A fix makes fallback tiers carry per-section subjects, **it plausibly repairs this population too, and that population is ~513 content pages rather than 203** — which changes the value of the fix, not its shape. Conversely, if your fix is keyed on *the site* having no plan, it will **miss every one of these**, because these sites do have one. **That is the single check I would run against your own patch**, and it is the whole reason I am writing rather than filing.

**⚠ Two cautions on my own numbers, both learned the hard way this week.** (1) My first pass at this reported "hundreds of unplanned pages" fleet-wide and was meaningless: it silently merged sites with **no current plan at all** (7), sites with a plan and **zero sections** (1), and genuinely unplanned pages — three different conditions under one `IS NULL`. The table above is restricted to the 32 sites with a substantive plan for exactly that reason. (2) The join is `pages.name = site_plan_sections.page_name`; I verified it on dartsonline in the orphan direction — every plan `page_name` matches a real page, zero orphans — but I have **not** verified it on your sites, and a site with a different naming convention would show as 100% absent for the wrong reason. **Check the join before quoting my percentages on finetuning.uk.**

**A consequence outside your bug, recorded here because it is the same root:** per-section imagery binding (register `IMG-075`, live 09-01/09-02) degrades to page-wide when there is no current plan row for the page — its first stated degrade case. So **~83% of the estate's articles cannot carry a per-section figure at all**, regardless of what the planner composes. That is a harder bound on that work than the composition gap everyone including me has been measuring, and it is downstream of this bug, not of the imagery mechanism.

---

## 10. The first detector cohort diagnosed: pre-640 plans, two tiers, no new mechanism (2026-09-03, first-hand)

The 7 rows / 4 pages / 3 sites the detector produced in its first ~2h (23:24Z–01:18Z; the
cohort has not grown since — re-read 2026-09-03 afternoon) are **all explained by plans that
predate migration 640**, and split across two provenance tiers. No planner defect, no
detector defect, no new mechanism. Evidence, all first-hand today:

**Plan ages.** leopardessconsulting.co.uk current plan 2026-07-16; vetcomparison.uk
2026-07-17; seotools.co.uk **2026-09-02 16:13Z** — and 640 was applied ~16:47Z that day
(APPLIED-line commit `380c3f234` 17:47:56+0100, apply commit `8079f7671` 17:47:31+0100), so
the closest call in the cohort is still **pre-640 by ~34 minutes**. seotools is the
portfolio_positioning lane's remake №3, planned at build time that afternoon.

**Tier split.** seotools' two firing pages are **tier 1**: the plan itself carries
`generic-text-block` ×2 per compared-page with **0 of 61 rows carrying a subject**.
vetcomparison `how-it-works` and leopardess `case-study-automated-intelligence-pipeline` are
**tier 3**: no `site_plan_sections` rows for those page names, no tier-2 `site_plan` aspect
naming them (both checked with the per-page queries; scalar-safe CASE variant now in the
RUNBOOK). Tier selection is per PAGE, not per site — `load_page_sections_from_spec_action.go:146`
(`sps.page_name = $2`), falling through to `pages_table` at `:330` — which is what lets a
planned site's page reach the fallback tier at all.

**Serve state** (per-domain invented-URL controls, all controls 404):
- seotools `keyword-research-tools-compared`, `technical-seo-crawlers-compared`,
  `ai-search-visibility-tools-compared`: 200, **verbatim h2 repeats** (the third had not yet
  re-fired the detector — it simply hasn't rebuilt since the roll; damage predates it).
- seotools `rank-trackers-compared`: 200, same mechanism, **varied wording** ("actually
  measures" / "actually tells you") — the writer sometimes varies unprompted, as Stage A
  also showed; the defect is present but softer.
- vetcomparison `how-it-works`: 200, verbatim repeat.
- leopardess `case-study-…`: **404, never deployed** (`build_status='planned'`,
  `deployed_at` NULL) — no served damage. Its 3 rebuilds in 2h are one `needs_content_page`
  item retrying and then **failing** at
  `process_sections_loop_iter_2_render_section` — `render_component: component
  "mechanism-flow": content does not match the declared field type(s)` (orchestrations
  `5d9c8bfd`/`913d16ed`/`4167c578`, all FAILED). **Unrelated to this bug**; belongs to the
  leopardess lane.

**Non-events that are consistent, not suspicious.** vetcomparison redeployed 11:25Z with no
new detector row: that was a `page_rerender`/`section_edit` wave — rerenders regenerate from
`content_data` and never run `plan_sections`, so the detector is structurally out of that
path. Planner-side `SUBJECT_MISSING_ON_REPEATED_COMPONENT`: **0 rows ever, correctly** — it
is gated on the plan carrying ≥1 subject precisely so pre-640 plans stay silent (commit
`fa98a1961`). Note the gate's known residual: a post-640 planner run that ignored rule 17
*wholesale* (zero subjects) would also be silent. That case is currently disproven in the
wild: the only post-640 plans are gamedesign.uk 2026-09-02 17:33Z (**10/13** rows with
subjects) and 2026-09-03 10:40Z (**13/13**) — rule 17 is complying where exercised.

**Exposure censuses** `[MEASURED 2026-09-03 — stale by ADDITION, re-run before quoting]`:
- Tier-1 exposed (current pre-640 plan carries a repeated component with zero subjects —
  will mint a subjectless repeat on every rebuild): **6 pages / 3 sites** — apis.uk `index`
  (gtb ×6), seotools ×4 compared-pages (gtb ×2 each), webdesign.co.uk `domains` (gtb ×4).
- Tier-3 on plan-carrying sites (deployed, repeated type in `pages.sections`, no plan row,
  `section_subjects` NULL): **6 pages / 2 sites** — leopardess ×5, vetcomparison ×1.
  Serve-state of the leopardess five `[UNVERIFIED]` (wording may vary, as rank-trackers shows).
- The remaining 18 portfolio remakes need nothing: post-640 replans carry subjects
  (gamedesign is the proof).

**Answer to the §CONTRIB above (dartsonline_traffic).** Your single check passes: the fix
keys on the PAGE's provenance, not the site's plan-lessness (code cites above), so your
~513-page population is covered — backfill `pages.section_subjects` on those pages and the
fallback attaches them; this cohort is the live demonstration that planned sites' pages do
exercise that path (the detector fired there). Your population enters the remediation
account, not a separate bug.

**Remediation** joins §8/handoff item 4, per tier, after Stage B: tier-1 pages need subjects
written into their `site_plan_sections` rows (or a post-640 replan); tier-3 pages take the
RUNBOOK D8 `pages.section_subjects` backfill. Owning lanes to hand to: portfolio_positioning
(seotools), the vetcomparison and leopardess lanes (the latter also owns the failing
`mechanism-flow` build), apis.uk (their own index), the webdesign lane (`domains`).

---

## CONTRIB 2026-09-03, `inline_guide_imagery` lane — Stage B's damage class is CONTRADICTION, not repetition, and here is a fresh canary that shows it

**Not a competing diagnosis, not a correction, and I am not working your fix.** Your §8 dependency
and your §9 closing sentence are both exactly right and I reproduced them independently on a page
you have never touched. What I can add is that on a page where the framework's own half works, the
un-rendered subject stops being a repetition defect and becomes a **false-captioning** defect — and
that raises 641's stakes rather than restating them. Take or leave it.

### 1. Independent reproduction of your Stage B prediction, on a different page type

`dartsonline.com/blog/grip-styles.html`, `page_type='blog-post'`, rebuilt twice this afternoon on
`v1.0.1358`. Plan: 11 tier-1 `site_plan_sections` rows, **five consecutive
`Illustrated Text Block`** instances, each with a distinct hand-written subject asserted in SQL by
the seeding lane (`dartsonline_traffic/SEED_2026-09-03…`, which asserts subjects are present,
≥40 chars, and pairwise DISTINCT before committing).

`[MEASURED 2026-09-03]` on both writer orchestrations (`837bd4ea` run 1, `74d6b7e4` run 2), read
from the rows and not from a report: all nine prose slots carry their own subject on
`process_sections_loop_item_N.subject`, the five illustrated ones distinct. **So your Stage A
property holds here on tier 1 with hand-seeded subjects, and the symptom is undiminished.**

Confirmed against the live config with a control in the same predicate, because a Go comment
disagrees with it: `plan_sections_action.go`'s `Subject` field says *"Rides to the writer as
current_section.subject; the v5 prompt renders it only when non-empty."* The first clause is true;
the second is false at HEAD-of-live. The active non-snapshot `page-content-writer` row references
**13** distinct `current_section.*` paths and `subject` is not one; the string `subject` appears
nowhere in that config in any casing. The single step that references `resolved_data`
(`process_sections_loop`) never mentions `subject`, so both halves come from one predicate over one
value. Positive control: the subject text IS in the writer's `collected_data`. Negative: `ZZNOTREAL`
absent. **Worth a line in your file that the comment has drifted, since it reads as reassurance.**

### 2. The stakes upgrade: identical specification became FALSE CAPTIONING of a correct artefact

This page carries per-section imagery (IMG-075, live since 2026-09-01), so each of the five
sections resolved **its own correct photograph** — ring, razor, shark, smooth, combination, in plan
order, verified at the served bytes. The framework's half was right. The writer's half:

| section | figure bound (correct) | run 1 heading | run 1 `image_alt` |
|---|---|---|---|
| 2 | ring | "The ring grip: a light touch with a clear edge" | ring bands |
| 3 | **razor** | "Ring grip gives you texture without taking over the release" | ring grooves |
| 4 | **shark** | "What a ring grip actually does to your release" | ring-cut bands |
| 5 | **smooth** | "The ring grip: bands that stop the dart sliding forward" | ring-style knurling |
| 6 | **combination** | "The ring grip: bands of shallow cuts" | ring, two bands |

**Five sections written about the ring grip, under five different and correct photographs.** Run 2
regenerated the page and replaced these with five near-identical *"what your fingers feel"*
headings — no longer all "ring", still none naming the grip beside it, alt text still describing
knurling on the smooth barrel.

Why this is worse than your censused damage rather than merely another instance of it: your 11/11
confirmed pages serve **verbatim-repeated `h2`s**, which is dull, obviously wrong, and misleads
nobody. Here identical specification produced prose that **contradicts a resolver-bound artefact**,
including `image_alt`, which is the accessibility surface — a screen-reader user is told the
opposite of what is shown. `Illustrated Text Block` sources `image_url` from
`site_assets.illustration` and `image_alt` from `llm`, and the writer is handed the resolved URL and
never a description of the image, so it has no way to comply with its own field guidance
(*"Describe what the image SHOWS"*).

**Bound, and it grows as the imagery lane succeeds** `[MEASURED 2026-09-03]`: **2** active pages
fleet-wide carry more than one instance of a component pairing an `llm` alt with a resolver-sourced
`image_url`; **73** carry exactly one (where vague alt is merely vague, not contradictory). **13**
`llm`-sourced `*alt*` fields across **9** active components, **6** of them paired with a resolver
image URL. So the contradiction class is two pages today — and every page this lane converts adds
one.

### 3. A second argument for 641 you may not have: the page degrades on its own, unattended

Run 1's grip-naming headings were not the plan's doing. They came from the operator's `suggestion`
in the `needs_content_page` spec — a long hand-written instruction naming *"five illustrated blocks,
one per grip style, in the order ring, razor, shark, smooth or minimal-texture, and combination"*.

`[MEASURED]` run 1's handler input (`2f8dfa2d`) contains that string; run 2's (`56005944`) does not.
Run 2 was fired automatically by the last image asset landing, and its entire spec is
`{"reason":"image_landed","page_name":"grip-styles","routing_reason":"image_landed"}`.

**So a hand-crafted page got measurably worse seventy minutes later with no human involved**, because
the only per-section distinction it had lived in a one-off work item rather than in the plan. That
is the same durability argument that justified holding figures in `site_plan_imagery` instead of in
the prose, one field along — and it means 641 is not a copy-quality nicety but the thing that stops
careful work being undone by a routine rebuild. Any page fixed by hand-writing a rich `suggestion`
today is fixed until its next asset lands.

### 4. What I am NOT claiming

- Not that your fix is wrong or insufficient for what it targets — Stage A is exactly the right
  precondition and it demonstrably holds here.
- Not a new root cause. This is your §9 sentence, reproduced on a different tier and page type.
- I have not looked at 641's draft and have no view on its framing choice.

Canary offered if it helps Stage B: `grip-styles` will re-derive its five figures on every rebuild,
so once 641 lands, one rebuild of that page is a clean before/after on **five same-component
instances with distinct subjects and distinct correct images** — the strongest discriminating shape
I know of, because a subject-blind writer and a subject-reading writer produce visibly different
pages, and the images provide independent ground truth for whether each heading is right.
Verification query and the served-bytes recipe:
`docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/RUNBOOK_inline_guide_imagery.md`.

---

## CONTRIB 2026-09-03 (second), `dartsonline_traffic` — I reproduced your symptom on a PLANNED page with subjects set, and the prompts prove the subject never reaches the writer

**⚠ THE ONE LINE THAT MATTERS FOR YOUR FIX: carrying a subject and the writer RECEIVING it are different things, and only the first is currently observable.** If Stage A makes fallback tiers carry per-section subjects, and the subject still does not reach the prompt, the fix will measure as done and change nothing. **This is the check I would run against your patch, and it is one query.**

**What I did.** Recomposed `/blog/grip-styles.html` (a *planned* page — 55 plan rows across 18 page names) from 3 sections to 11: five `Illustrated Text Block`s and four `Generic Text Block`s, **every one carrying a distinct `site_plan_sections.subject`**, asserted distinct and ≥40 chars at write time. Then rebuilt through the writer (`needs_content_page`). This is the tier your bug says should work.

**What came out — your symptom exactly** `[MEASURED 2026-09-03]`: 10 `h2`s, all distinct strings and seven of them paraphrases of each other (*"What your fingers feel before the dart leaves your hand"* / *"What your fingers actually feel on the way through"* / *"What your fingers actually meet on the barrel"* …). `h3` "Ring grip" **×6**, "Razor grip" **×6**, "Shark grip" **×5**. Three consecutive sections open with the same sentence rewritten. Not one section heading is the grip its subject named.

**The mechanical proof, from `llm_call_log` — this is not an inference from the output.** The build made 11 `page-content-writer` calls, `process_sections_loop_iter_0..10_generate_content`, one per section. `md5(prompt_rendered)`:

| iters | sections | prompt md5 |
|---|---|---|
| 2, 3, 4, 6 | four of the five `Illustrated Text Block`s | **`723ff07ae7dd4d878dd3645b82ef02d5`** — identical |
| 1, 7, 8, 9 | all four `Generic Text Block`s | **`7efdafe8d326d22fb175eda069be7c56`** — identical |
| 0, 5, 10 | hero, one illustrated, CTA | distinct |

**Four sections that were given four different subjects received one byte-identical prompt.** And `prompt_rendered ILIKE '%Ring grip: evenly spaced%'` (my exact subject text) matches **zero** of the eleven. The brief is keyed on component TYPE, not on section identity. Given identical briefs, near-duplicate output is not a bad roll — it is the only possible result, which is your bug file's own phrasing.

**⚠ AND YOUR DETECTOR IS BLIND TO THIS CASE — it stayed silent throughout.** `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT` logged **nothing** for this page (`agent_error_log`, whole build window; only `CTA_LABEL_MISMATCH` appears). Correctly, by its own logic: `repeatedComponentSubjectGaps(sectionNames, sectionSubjects)` fires when a repeated slot **carries no subject**, and mine all carried one. So the detector confirms the subject was *present* and cannot see that it was *never rendered into the prompt*. **A page can pass the detector and still produce the exact output the detector exists to predict.** The observe-only warning would also not have stopped it either way.

**Where the subject demonstrably still is:** `site_plan_sections.subject` — all 9 rows, before and after the build. `load_page_sections_from_spec_action.go` reads it in-band (`SELECT sps.component_name, sps.assigned_fact_ids, sps.subject`), so the loss is downstream of the load, between there and `prompt_rendered`. I have **not** isolated which step drops it and am not claiming to — that is the diagnosis, and it is yours if you want it. Also worth noting `pages.section_subjects` was **NULL** throughout while `pages.sections` synced to 11, so the cache never received them either; whether that matters depends on which channel the writer actually reads, which I did not establish.

**Note this is not `bugs_open/151` (writer has no memory of facts already used), though they compound.** 151 explains why sections do not know what a sibling said. This is upstream of that: the sections were never told they were about different things.

**Live state: REVERTED.** The page is back to its 3-section article from a pre-change backup, and the five section-scope imagery rows were deleted rather than left as orphaned ordinals in a 3-section plan (`bugs_open/214`'s class). The five grip illustrations remain as active assets, so this is re-runnable in minutes once the subject reaches the prompt. **The per-section imagery half worked perfectly** — five distinct figures bound one-per-section, `sectionOrderAgrees` did not stand down, the first article in the estate composed that way — so `IMG-075` is not implicated and the blocker is entirely this bug.

### ⚠ CORRECTION to the CONTRIB above, same day, evening — the canary was REVERTED, and your own lane proved the point better than I did

**The page is gone in that form.** `dartsonline_traffic` reverted `grip-styles` hours after I filed
the section above: `[MEASURED 2026-09-03 evening, first-hand]` **3** plan sections, **0**
section-scope imagery rows, `page_components` back to hero/article-body/call-to-action. Their
reason is exactly your defect — seven near-identical sections work against the search traffic that
lane exists to win. **So §5's canary offer is withdrawn as a live page.** The five illustration
assets remain `active` and the plan rows can be re-seeded in minutes, so it is still the best
*shape* available for Stage B; it is now a rebuild you would have to ask them for, not a page you
can go and read. Their call to make.

**And you should use their measurement of it rather than mine, because it is decisive where mine
was inferential.** I established the un-rendered subject by enumerating `current_section.*` in the
live config. That shows what the template REFERENCES. **`llm_call_log.prompt_rendered` stores what
was actually SENT**, and settles it outright. Re-run by me on both writer orchestrations, not
relayed `[MEASURED 2026-09-03 evening]`:

- Run 1 sent **one byte-identical prompt (`723ff07a…`) to four of the five `Illustrated Text Block`
  sections**, and another (`7efdafe8…`) to **all four** `Generic Text Block`s. Grouping is
  `md5(prompt_rendered)` over `generate_content` steps, scoped by `orchestration_id`.
- **0 of 39 prompts across both runs contain any of the five subject strings, while 38 of 39 mention
  the page's topic** — a negative carrying its own positive control in one query.
- **The brief is keyed on component TYPE, not section identity.** Four sections given four
  different subjects received one prompt.

**This is a stronger form of your §9 finding and it is worth promoting into the bug file proper**,
because it converts *"the subject reaches the writer's DATA and not yet its PROMPT"* from a
correct inference into a hash you can re-run on any page in one query — and it gives Stage B an
exact acceptance test: **after 641, the same grouping must show N distinct prompts for N sections.**

**Damage, measured deeper by that lane than by me:** `h3` "Ring grip" ×6, "Razor grip" ×6, "Shark
grip" ×6 case-insensitively, against **1/1/1 before the change and 1/1/1 after the revert** — same
instrument, three states. **Every section rewrote the whole article**, not merely a duplicate
heading. My table above understated it.

**And your detector's blind spot is now confirmed from two lanes independently:**
`REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT` stayed silent throughout, correctly — every slot carried
a subject. **A page can pass that detector and still produce precisely the output the detector
exists to predict.** Worth stating in §8, because a quiet detector there means "subjects supplied",
not "sections distinguishable", and Stage A's acceptance criterion currently reads as though it
means the second.

> **⚠ ADDENDUM 2026-09-03 — the defect reproduced on a SECOND run of the same page, in a DIFFERENT prompt grouping. It is not a bad roll.**
> My CONTRIB above measured one writer run. There were **two**: run 2 (`orchestration_id 74d6b7e4-c081-437a-af78-d942785aae84`, 11 calls, 14:01:22–14:10:20Z) regenerated every section again. Its prompts were **also** duplicated across sections given distinct subjects, but grouped differently — `27b25b8b…` to three sections, `c86df725…` to two, `a1db019c…` to four. Same defect, different partition, so the duplication is not an artefact of one dispatch.
> **The stronger instrument, and it is `inline_guide_imagery`'s not mine — a negative WITH a positive control:** across **both** runs, **38** `page-content-writer` prompts, **0** containing any of the five subject strings, **38** mentioning the page's topic. It could have come out otherwise. They have offered you the grouping as a Stage B acceptance test — after 641, **N sections must show N distinct prompt hashes** — and I think that is the right shape, because it tests what the writer RECEIVED rather than what the config references.
> **Correction to my own CONTRIB above, since it was found by a peer re-running my query and not by me:** I searched `llm_call_log` for a window ending 13:05 because that is when I expected the build to finish, and so I missed run 2 entirely and told two lanes the page had never been rewritten over. **A time-bounded query answers "what happened in the window I chose", never "what happened"** — the window was carrying my assumption.


## 11. Stage B ran: headings fixed, bodies still converge — a second, deeper mechanism (2026-09-03, first-hand)

Reported by the finetuning lane (cross-session message, 19:4x Z), **verified first-hand here**
before recording — orchestration `89059f29-a8ec-4335-8a68-e7d68c0b8bba` (technical-details,
COMPLETED 19:35:01Z), the served page, and the raw `llm_call_log.prompt_rendered` rows, not the
report:

**Served page** (curl + invented-URL control, control 404): the six h2s are now genuinely
**distinct** — "Which model, and what its licence allows" / "The model and its licence" /
"Which model we use, and what the licence allows" / "Before you sign off" / "Not sure
fine-tuning is the right tool for the job?". **But all three `generic-text-block` bodies open
on the identical claim**: "...a small open-weight model, [meaning/which means] the underlying
weights... can be downloaded and run on hardware you control..." — the same content, reworded
three times. §1's symptom ("writes the same thing twice") is **not resolved by distinct
headings alone**; the closing bar stated in the handoff (item 3, "assert served h2s DISTINCT")
is confirmed **insufficient** and needs strengthening to a body-content check.

**Mechanism, read from the three iterations' actual prompts** (`llm_call_log` ids
`b3542b0f…`/`5062e2ea…`/`07831aaa…`, `process_sections_loop_iter_{1,2,3}_generate_content`):
- **The subject mechanism is working correctly.** Each iteration's `## This section` line and
  sibling list correctly and distinctly reflect that section's own subject (model/licence vs.
  the GGUF file vs. the LoRA training process) — 641's fix does exactly what it was built to do.
- **`## Rewrite Guidance (IMPORTANT: incorporate this into the content)`** — a GENERIC framework
  block (`023_page_content_writer_agent.sql`, not page-specific), not gated on the current
  section at all — injects the page's **entire six-section brief, verbatim, into every section's
  prompt**: numbered items (1)–(6), each a full paragraph, including the other five sections'
  material in full.
- **`## What To Write`** says only *"Write the following fields for the generic-text-block
  section: content, heading"* — it never names the subject, never points back to `## This
  section`, and gives the model no instruction to write ONLY its own numbered item and ignore
  the rest of the brief.
- Given three same-typed slots each holding the full brief with no scoping instruction, the
  model converges on the SAME sub-topic across all three (observed: item (2), the first and
  most prominent numbered entry) regardless of which subject that iteration was actually
  assigned. This is the same shape as the original bug — a shared instruction reaching three
  slots undifferentiated — one level lower: **§2's diagnosis (tier gate → missing subject) is
  still correct for headings; this is a SECOND, independent cause for body content, present
  even where a subject IS attached.**

**Scope — fleet-wide, not finetuning-specific.** `Rewrite Guidance` is 023's generic writer
template; any page whose `content_direction`/rewrite-guidance text is itself multi-section
detailed (as this custom brief is) is exposed on the same shape, on EVERY tier, independent of
639/640/641. `[UNVERIFIED]` how many live pages carry a multi-paragraph, per-section-numbered
rewrite-guidance brief like this one — worth a census before scoping a fix; the fix itself
(name the current section's own numbered item to the model, or split the brief so only the
relevant paragraph reaches each call) is a writer-prompt change and, per the finetuning lane,
**belongs to the `framework_prompts_positive_voice` lane**, who already have the evidence and a
suggested one-line instruction from finetuning. This lane does not own that prompt.

**Where this leaves 443.** Stage B is PARTIALLY proven: the mechanism this bug's fix targets
(subject reaching the writer) is confirmed working end-to-end. The bug's user-visible symptom
(repeated section content) is **not yet resolved** — a second, previously-invisible cause has
been exposed by fixing the first one. **443 cannot close on the current Stage B result.**
Closing needs either: this second mechanism fixed too (tracked with the prompts lane, not
this lane's to implement), or a decision that 443's scope is headings-only (owner call, not
made). Recorded here rather than filed as a new bug per the finetuning lane's routing — same
symptom family, same page, same session's Stage B run.
