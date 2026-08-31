# OWNER REVIEW 2026-08-31 — boxingonline.com: what he found, and what each finding actually is

**Site:** `d2aa5206-73bc-4707-a69c-2702c1eb9152`, order BR-9AUZ59, the first paid customer
build. Planned and built 2026-08-31, reviewed by the owner in chat the same evening.
**This file records his critique, what I measured against each item, and which lane now holds
it.** Everything marked `[MEASURED 2026-08-31]` was checked against the live DB or the served
site at `https://boxingonline.ugg2.com` that day, query inline. Where I did not settle a
mechanism I say so rather than inferring one.

Pipeline ownership stays with this lane's other session; I own the critique write-ups. Split
agreed cross-session the same evening.

---

## 0. The one thing that was urgent, and its honest status

His personal address `aaa@designconsultancy.co.uk` was published on the site: the footer
`Contact` block of **every** page, plus four places on `/contact.html`. He asked for it off
immediately.

**Done in the database, verified in-transaction:** `sites.email` → NULL (that column is what the
footer chrome is assembled from); the `contact-info` component's `email` key dropped and its
prose rewritten to point at the on-page form; `contact-form` and `contact-info` rendered HTML
cleared; the index chrome payload's email blanked. No replacement address was invented
(`bugs_open/140`). A whole-site rerender was fired and **completed** (correlation
`3f604312-d5ad-4ad2-8930-d74f66591940`, receipt asserted via `kafka-publish-lib.sh`, rc=0).
Post-rerender: **0 component rows and 0 page-chrome rows carry the string.**

**Not yet true of what the public sees.** The public copy is a b2worker mirror re-published by
an hourly reconciler, so the served pages lag the origin by up to an hour. Measured immediately
after the rerender, with a control string that must be present: `/` email=1 (control 8),
`/contact.html` email=4 (control 9). The other session owns that seam and is verifying the
re-mirror end to end. **Do not report this as closed until the served page is clean** — the
database being clean is not the same fact.

> **CLOSED 2026-08-31 16:23Z — and my first two "clean" readings were both wrong, which is the
> part worth keeping.**
>
> **Final state, two independent sweeps agreeing.** All 19 deployed pages enumerated from
> `pages WHERE deployed_at IS NOT NULL`, cache-busted, each with a must-be-present control:
> **19/19 `email=0`, every control non-zero** (3–11 hits), plus a positive control proving the
> grep finds strings that ARE present ("Get in touch" on /contact.html → 4). All four stored
> sources re-verified in the same minute: `site_components=0 page_components=0 specs=0
> siterow=0`. The delivery lane's independent sweep adds an invented-URL catch-all control
> (404, so the domain is not blanket-200ing). Nothing re-populates it: the identity spec's
> contact block is null, so the fill-if-empty sync has nothing to copy, and the order seed does
> not re-run.
>
> **Two false "clean" readings on the way, both the same failure — an incomplete enumeration
> reading as a result.**
>
> 1. I censused `sites`, `page_components` and `site_specs`, asserted clean in-transaction, and
>    told the owner the database was clean. The footer lives in a **fourth** table,
>    `site_components`, which I never enumerated. `pages.rendered_footer` is NULL site-wide, so
>    the footer looked as though it had no store at all; it is assembled at deploy from that row.
>    It kept publishing the address for ~40 minutes behind three passing checks.
> 2. A single-page watcher then reported `/about.html` clean and would have closed the case. A
>    full sweep ten seconds later found **6 of 19 pages still carrying it.** The one-page probe
>    was the optimistic one.
>
> **And a plausible story absorbed a true positive.** For ~40 minutes a correct reader-side probe
> kept returning `email=1` with a good control, and two sessions independently explained it as
> publish-mirror latency — which made it feel corroborated rather than checked. The tell we both
> walked past: `last-modified` / `x-amz-version-id` showed the object **freshly written and still
> dirty**, i.e. a current publish of a dirty source, the opposite of lag. Logged in
> `WRONG_CALLS.md`.
>
> **Two real defects fell out, both in `bugs_open/420`:** `refresh_site_components:true` refreshed
> `head` and `header` and **skipped `footer`**; and site-level chrome does not participate in a
> page's content hash, so a chrome-only change leaves every page looking unchanged — a whole-site
> rerender no-ops on all of them **and reports success**. Targeted per-page rerenders were the
> only thing that moved it.
>
> **Two residuals, on the pre-delivery list, NOT closed:** the footer row is a surgical
> `regexp_replace` edit and must be re-rendered properly from `content_data` before handover; and
> the contact page is now a form with `form_action "#contact"` that submits nowhere, on a page the
> brief never asked for — delete-or-wire is with the owner.

**The defect behind it is real and will fire again.** Order intake writes the ORDERING email
into `sites.email` (council-approved as the identity store for delivery), and the footer
assembles the PUBLISHED contact from the same column. Two contracts on one column. On order 2
it publishes whatever address that customer happened to pay with. The other session is filing it.

---

## 1. "The copy tells the user what the site is doing rather than talking to the user"

His headline complaint, and he quoted three passages, e.g. from `/about.html`:

> **How we cover it** — We write the way a knowledgeable fan talks… A preview that says a fight
> 'could be great' tells the reader nothing, so we'd rather name the styles, the records and the
> stakes…

**What I found:** those sentences are a paraphrase of **our own research spec**. The
`vertical_landscape` aspect holds a `lessons.avoid[]` list written as instructions to the build,
including *"Vague fight previews that say 'this could be a great fight' — every preview must
contain specific analysis of styles, records, and what's at stake"*, *"Letting opinion drift into
fact"*, *"Stale calendar entries — a wrong fight date actively harms readers"*. Each maps onto a
sentence he quoted. **The instruction sheet was rendered as page copy.**

This is the copy lane's own headline mechanism — *demonstrations govern, instructions don't* —
in a new variant: a rule ABOUT the writing, placed in the writer's context, comes out AS the
writing. **Filed:** `copy_quality_two_stage/CONTRIB_2026-08-31_from_the_first_paid_build_the_page_DESCRIBES_the_editorial_policy_instead_of_doing_it.md`.

**What I did NOT establish:** whether the writer saw `vertical_landscape` directly, saw it via
`strategy`, or converged independently. `[UNVERIFIED]`. The three have different fixes, and the
copy lane's replay harness settles it more cheaply than a diagnosis run.

## 2. `/articles/index.html` is "explanations about what we're doing" with no benefit to the reader

Confirmed: 3,114 characters of body copy, **zero articles**, four headed sections of editorial
policy (*"What's in the mix"*, *"Keeping it accurate"*…). He quoted the last one.

**Why there are no articles:** `bugs_open/419` — the site planner emitted the `article`
(blog-post) page with **zero** planned sections, so the six article slots the paid brief
promised never built. Then the link validator rewrote the dead editorial links to point at
existing pages, so the page renders clean with no visible defect. Every page on the site
validated `valid=true, issues=0`.

**The part worth noticing:** emptiness did not present as emptiness. It presented as a
well-written, in-voice, fully-validated page. Two separate honesty mechanisms tidied the absence
away. **The mechanism behind 419 is deliberately NOT settled** — the diagnosis run came back
UNVERIFIABLE, so 419 stands as symptom and census only and no lane should quote a cause yet.

## 3. The quiz "links to a guide, not the quiz" — and the guide is more prominent than the tool

He first read it as a broken link, then found the tool and sharpened it correctly: **the guide is
more prominent than the tool.** It is, and structurally:

- The home page's editorial block is titled **"Latest from the ring"** with the subtitle *"Fresh
  news, previews and results from around the boxing world"*. Its four items are the four
  auto-generated tool guides, displayed under their literal titles — *"Understanding Boxing Quiz
  — Test Your Knowledge | Guide"*. All four have an empty image field.
- The four **tool** pages have `in_header = true` but **`nav_label` NULL** and `nav_order` 200,
  so they cannot render in the navigation at all.

So the usable thing is reachable only from a card grid below the fold; the explainer leads the
page. The listing filled with guides because zero articles existed — 419 again, downstream. The
other session expects it to partly self-correct once the six articles land, and will re-measure;
if the listing still prefers guides afterwards that is a fresh defect.

**Filed:** `experience_loop/CONTRIB_2026-08-31_from_the_first_paid_build_four_experience_defects_that_every_check_passed.md`,
with a proposed durable check: *a listing's items must belong to the content class its heading
promises* — worth having regardless of 419.

## 4. The countdown guide is "way too long" and says one interesting thing

Confirmed and fair. `article-body` is 4,415 characters across six headed sections. The one real
insight — a card advertised at 9pm may not see the main event ring walk until past midnight — is
in the second paragraph, and the remaining five sections restate it. The other three guides are
the same shape (5,200 / 4,706 / 6,158 characters).

**His question: do we have a quality auditor that might have picked this up?** **Yes, and it
never ran on this site.** `content-quality-auditor` has been active since 2026-03-06 and has 49
COMPLETED / 25 FAILED runs fleet-wide, the most recent at 15:19Z the same day — and **zero** of
them touched boxingonline. It is not in the new-build path. That is a capability we maintain and
did not receive on the build that most needed it. (The 34% failure rate is a separate question I
have not looked into.)

## 5. "Not enough imagery … why didn't we use infographics"

**Every served page carries exactly one `<img>`, and it is the logo.** Four hero images exist but
sit behind text as CSS backgrounds. The eight imagery jobs all completed successfully — they
produced 1 logo, 4 heroes, 3 icons, which is the entire set the planner asked for. Zero
illustrations, zero infographics.

**The fleet number is the finding.** Across every site plan ever written: hero 359 rows / 29
sites, icon 196 / 25, logo 45 / 28, illustration 19 / 5, and **infographic — 1 row, on 1 site.**

The capability is fully built and wired: generation config, provider routing, plan admission,
and a reader that maps section-scope infographics into the render. **Nothing ever asks for one.**

And he asked for exactly this before: `inline_guide_imagery/PLAN_2026-08-14` opens *"The ask
(owner, 2026-08-13): guide/blog articles should carry explanatory imagery inside the article body
— between paragraphs and beside them"* and still says **"Status: design, nothing implemented."**
Seventeen days later it shipped to a paying customer without it.

His instinct that infographics should replace much of the explanatory copy is well supported by
the pages themselves: the countdown guide's real content is one timeline; the weight-class guide's
is one table of divisions and limits. **Filed:**
`editorial_design_uplift/CONTRIB_2026-08-31_the_infographic_kind_has_ONE_row_fleet_wide…md`, and a
pointer into `inline_guide_imagery`.

## 6. The tools make the reader supply the data

> "That fighter comparison tool requires the user to input all the details… we should make the
> comparisons just from the name and include all that information from our research instead."

Confirmed: **18 manual inputs** (two free-text names plus wins, losses, draws, KOs, reach,
height, age and a form string, twice over) and **no fighter data ships with the page at all** —
searching the payload for Usyk, Fury, Joshua, Canelo or Inoue returns zero. The weight-class
finder is the same shape: enter a weight, not a name.

The tool-suggester's own recorded reasoning is thoughtful about relevance and **never once asks
whether we hold data that would let the tool answer anything**. A pure client-side calculator is
what you get when data availability is not a selection criterion. Worth stating plainly: there is
currently no fact on this site that the visitor did not bring with them.

## 7. No biographies, no editorials, no directories (where to watch, suppliers)

Correct, and the sharper point is that **our own research recommended most of them.** The
`vertical_landscape` research (confidence 0.88; ringtv.com, boxingscene.com, skysports.com/boxing
crawled and analysed) explicitly recommends fighter profile/tag pages as a structural layer,
event pages as first-class objects, and a repeatable fight-time/how-to-watch format. The plan
carries none of them.

## 8. "Is every site supposed to be the best in its vertical?" — his direct question

**The research knew. The strategy knew. The plan did not, and the phrase reaches nothing.**

- The research is genuinely good. It names a differentiation opportunity that is better than
  anything on the built site: own the *"what to watch this weekend"* moment — a weekly curated
  viewing guide across every broadcaster with fight-time conversions and honest one-line
  previews, *"like a mate who's already done the research so you don't have to."*
- The `strategy` spec carried it: *"A weekly 'what to watch' curated guide article… becomes the
  signature piece — the thing readers bookmark the site for"*, plus fighter entity pages and the
  magazine grid.
- **The plan is six pages** — index, about, contact, articles-index, fight calendar, and the one
  article page that never built. No weekly guide, no fighter pages, no magazine grid.
- **The phrase itself never arrives.** 0 of 10 current specs on this site contain "best in
  class"; `strategy` has no `benchmark` key. `vertical_landscape` is read by exactly two agents
  fleet-wide (the strategist and its own writer), and in the Go code it appears only as an
  existence check — we confirm the research row is present, then build without opening it.

The copy lane wrote `PLAN_2026-08-25_best_in_class_propagation.md` six days ago in answer to his
own 08-25 ruling, and measured then that **0 of 51 sites** carried the standard. This is site 54
and the first paid one, and the measurement still holds. **That plan is the standing fix and it
has not shipped.** It is waiting on his go.

---

## What needs a decision from him

1. **Fix-before-delivery, or deliver-then-improve, for THIS site.** Most of what he raised is
   fleet-shaped and belongs in other lanes' machinery. But boxingonline is a paid deliverable
   about to be emailed. Only he can rule on the sequencing, and the other session's rehearsal
   plan depends on it.
2. **The go on best-in-class propagation** (§8) — the plan exists, is costed, and answers the
   question he asked tonight.
3. **What contact address the site should publish**, now that his own is off it (§0). Nothing
   will be invented in the meantime.
