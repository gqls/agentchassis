# HANDOFF — dartsonline.com. START HERE. Written 2026-08-31.

Supersedes `HANDOFF_2026-08-16_continue_here.md` (still accurate about the traps in its §5; its
§3 work is all done). Owning lane: `dartsonline_traffic`. Sibling lanes this one depends on:
`inline_guide_imagery` (design, unbuilt), `news_editorial` (owns `features_open/035`), and the
`apis.uk` lane (owns fleet traffic).

**Nothing is on fire.** Everything below is done-and-verified, queued, or a stated open question.

---

## 0. The one-paragraph version

The site was rejected by Webgains for insufficient traffic; that lane of work is now measurable
rather than speculative. Since 08-16 the site has gained a privacy page, a sitemap, 4+ new guides,
working card images, correct contrast, and a real traffic baseline. Four framework defects were
found and filed along the way, three of them fixed at the framework rather than on this site. The
two live threads are **imagery** (four bad heroes regenerating now; guides need per-section images
and the mechanism is another lane's, unbuilt) and **affiliates** (apply where the gate is not
traffic).

---

## 1. DONE and verified at the artefact

| thing | state | evidence |
|---|---|---|
| privacy page | **LIVE**, copy verbatim | 16/16 approved blocks matched, whole copy contiguous; owner removed the affiliate-independence sentence 08-20 |
| `/shipping-returns.html` | **RETIRED** | archived + retracted (`sites@2af7c17dd`), serves 404 |
| sitemap + robots | **LIVE** | `/sitemap.xml` 200; **37 URLs** as of 2026-08-31 (was 23 on 08-20) |
| stylesheet | **RESTORED** | was 164 bytes for 2 days (`bugs_open/198` clobber); recovered from git, seeded into `css_themes` so a patch run cannot re-clobber |
| contrast | **0 real failures** | all 23 pages, desktop *and* mobile. The 15 remaining rows are the probe's `overImage` placeholders, not measurements |
| card images | **12/12 on the homepage** | fixed at the framework by the `384` lane, not hand-patched here |
| privacy footer link | **23/23 served pages** | graded at the bytes, not at the item status |
| traffic | **measurable** | Cloudflare `Zone → Analytics → Read` live on 2 tokens |

---

## 2. The traffic baseline — and the trap inside it

**30 days to 2026-08-24: 5,631 page views, ~188/day.** But classified by client:

| class | share | /day |
|---|---|---|
| real browsers | 33.6% | 63 |
| **our own tooling** (curl + headless) | **28.5%** | 53.5 |
| unattributed | 36.2% | 68 |
| **search crawlers** | **1.5%** | **2.8** |

**⚠ Do not quote a rise without classifying it.** Total page views 2.6×'d across the window and
**none of it was growth** — human went 57.6 → 68.4/day (+19%) while our own tooling went 18.5 →
88.6 (4.8×). The `apis.uk` lane reproduced the contamination independently at 27.1% on their site,
with an unworked control site at 2.4% — so it is what an actively-worked site looks like, not a
dartsonline quirk. **Quote the pair, never the percentage alone.**

**The number that actually matters: GoogleBot made 54 page views in 30 days across 24 URLs.**
Search is barely visiting. That, not the page-view total, is what the affiliate applications turn on.

**Not established:** whether the sitemap moved crawling. 2.72/day before it shipped vs 3.00/day in
the 5 days after, two of those days zero. Five days cannot separate that from noise — **re-measure
at 30 days**, i.e. around 2026-09-19.

---

## 3. IMAGERY — the live thread, and where it actually stands

### 3.1 Four hallucinated heroes, regenerating now
Owner, 2026-08-31: the heroes "hallucinated what darts and dartboards and other objects look like".

**The measurement narrows this from "loads of pages" to four**, and the discriminator is mechanical:

> **An active `assets` row ⇒ the image is current and good. No asset row AND a served image ⇒ a
> July SDXL-era leftover file, and those are the bad ones.** `SELECT ... FROM assets WHERE
> url='/assets/images/<file>'` answers it without opening the image.
>
> ⚠ **The second clause is load-bearing and I left it out at first.** The `news_editorial` lane ran
> this rule on their own pages 2026-08-31 and their `insights-index` returned zero rows — which my
> original wording flags as orphaned. They controlled it instead of trusting the rule: the served
> page carries one `<img>` and it is the logo. **No row and no image is consistent, not orphaned.**
> Run the rule and then look at the page before calling anything a leftover.

- **GOOD, verified by eye:** `hero.jpg` (real bristle board, correct wire spider and bull colours),
  `content-hero-grip-styles.jpg` (four genuinely distinct, correct grip patterns). All Banana.
- **BAD, no asset row:** `hero-home.jpg`, `hero-new-arrivals.jpg` (feathered flights — those are
  archery arrows — and blue board segments), `hero-guides.jpg` (garbled numbers), `hero-sale.jpg`.

> ### ⚠ CORRECTED 2026-08-31 (afternoon, second session) — THE DISCRIMINATOR ANSWERS PROVENANCE, NOT ACCURACY
>
> The rule above ("active `assets` row ⇒ the image is current and good") is sound for the
> question it was built for: *is this file a stale July leftover?* **It cannot certify that an
> image is CORRECT, and I proved that by using it.** After the 12:35Z run all four keys had
> active Banana rows updated in place — the rule says "good" — and **two of the four were still
> wrong when I opened them.** Read the second clause as "no row ⇒ stale", never the first as
> "row ⇒ accurate". *An image is only verified by looking at it.*
>
> **And `last-modified` on the served file cannot stand in for looking either.** After the run
> it read ~12:41–12:56Z on all four — and *also* 12:39Z on `hero.jpg` and
> `content-hero-grip-styles.jpg`, which were **not** regenerated. The header records the bucket
> sync, not the generation, so a census keyed on it calls six files regenerated when four were.

**Four `needs_imagery` items filed 2026-08-31 12:35Z** (`SEED_2026-08-31_regenerate_four_hallucinated_heroes.sql`,
`created_by='dartsonline-traffic-2026-08-31'`), at `triaged`, handler `image-build-handler`.
**Verify at the served file, not the item status** — and expect them to overwrite in place, because
`deploy_image_asset` derives the filename from `(asset_key, purpose)` and refuses a caller path.

**⚠ The expired-signed-URL asset rows on this site are a RED HERRING.** Their `url` is a signed S3
link that expired 7 days after generation, nothing references them, and regenerating them changes
nothing a visitor sees. Left alone deliberately.
*Corrected 2026-08-31 afternoon: it is **10** rows, not 8, and **three are Banana** (`hero_brands`,
`hero_shipping`, `hero_shop`) rather than all SDXL — so "the stability rows" under-names the set.
The "nothing references them" half is now verified at the artefact rather than inferred: **no served
page references a `backblazeb2` URL at all**; every hero on every page is a local
`/assets/images/*.jpg`. The expiry is inert, not latent.*

**Why regeneration gets the good model** — proven, not assumed: `bugs_closed/382` flipped the
missing-kind routing default to Banana on 08-24, and that commit is an ancestor of the running
adapter build (v1.0.1349 / `ef06af0e0`), with a control commit 5 ahead correctly *not* an ancestor.
Every spec also carries `kind:'hero'` explicitly, since the **absent** kind was 382's cause.

### 3.2 The guides need per-section images — mechanism is NOT ours and NOT built
Owner, 2026-08-31: guides want an accurate image per small section (ring grip, razor grip, shark
grip…). Measured today:

```
grip-styles     0 content images, 7 h2 + 6 h3   ← the owner's own example
board-setup     1 in-body illustration, 7 h2 + 3 h3
flight-shapes   1 in-body illustration, 8 h2 + 2 h3
beginners       1 in-body image,        7 h2 + 1 h3
```

**Do not splice figures into `article-body` to satisfy this.** Figure and prose share one llm-owned
field, so anything added today dies at the next body rewrite — this lane lost four figures exactly
that way in August.

**⚠ AND DO NOT WAIT FOR COMPOSITION EITHER.** The `news_editorial` lane answered this on 2026-08-31
and the answer is unambiguous: **P5 — the phase where a guide migrates onto parent/children — is
months away, not weeks.** Measured by them the same day: P1 is in progress with **nothing live**,
the council has returned REVISE **three times** on the wiring plan, and **0 of 2,249 `page_components`
carry a `parent_instance_id`; 0 of 386 `content_components` declare a slots block**. P5 sits behind
P2, P3 and P4, and 035 §5 gates it on P1–P3 holding "for real weeks" with §6.1's un-owned-page
question settled first — and these guides are precisely the un-owned pages that question is about.

**The third option, which is this lane's to take and needs nothing from them:** the trap is not
composition's absence, it is that **ONE field owns both prose and figures**. A guide whose **h3
sections are separate `page_components` rows** — ordinary flat sections, no `parent_instance_id`,
no composition — has per-section durability **today**, because a rewrite then targets one section's
row and cannot take a sibling's figure with it. That is the same durability property composition
would give, one grain coarser. Composition's genuine addition over it is nesting a figure *inside*
a prose section; for `grip-styles`, where each h3 (Ring / Razor / Shark) wants exactly one image,
**finer sections are sufficient and nesting is not needed**.

> ### ✅ SUPERSEDED 2026-09-02 — THE MECHANISM SHIPPED. Do not take the flat-sections route, and do not wait.
>
> Everything above this box was true on 08-31 and is **out of date**. **`IMG-075` — per-section
> binding for section-scope imagery — went LIVE 2026-09-01 21:00 (chassis `v1.0.1351`),
> `Council-Reviewed: 2979c27f`, approved at round 3.** It does exactly what §3.2 says nobody could
> do yet: a page can carry **a different figure in each section**, re-derived from the plan on every
> build and every re-render — which is the durability property the flat-sections workaround was
> invented to approximate, obtained properly.
>
> **Verified at the artefact, not read off the register** `[MEASURED 2026-09-02]`, because a
> register status line is a snapshot that outlives its truth: binary probe on the running chassis
> pod returned `sectionRefForOrdinal` PRESENT, `sectionOrderAgrees` PRESENT, `PlanSectionsAction`
> PRESENT (must-be-present control), `sectionRefForOrdinalNOTREAL` ABSENT (must-be-absent control).
> **`sectionOrderAgrees` is a ROUND-2 symbol** — probing only round-1 symbols returns all-present on
> a binary carrying half the change, which the IMG-075 entry warns about in its own words.
>
> **A peer lane told us on 09-02 that "the resolver maps by KIND first-wins, so several illustrated
> sections on one page resolve to the same image" and that grip-styles "is precisely the case that
> limit blocks". That was true until 09-01 21:00 and is now stale** — it is the defect IMG-075 was
> written to remove, and dartsonline's own grip-styles is the worked example in its test
> (`TestPlanSections_SectionScopeIllustrationBindsToItsOwnSection`, fixtures
> `illustration-ring-grip.jpg` / `illustration-shark-grip.jpg`). Corrected back to them.
>
> **What is actually missing is the ASKING, and it is this lane's to do.** `[MEASURED 2026-09-02]`
> dartsonline has **zero** section-scope *illustration* rows — its 7 section-scope rows are all
> `kind='icon'` on `index:2`/`about:2`, and **icons cannot reach this branch at all** (they resolve
> by literal key; the per-section map holds kind keys only).
>
> > **⚠ CORRECTED same day — and the correction is the more useful fact.** I first wrote here, and
> > told a peer lane, that *"no page in the estate has more than one section-scope illustration
> > row"*, i.e. that the mechanism was live and **undriven**. **That figure was IMG-075's, stamped
> > `[MEASURED 2026-08-31]`, and I relayed it without its date.** Re-run `[MEASURED 2026-09-02]`:
> > **`apis.uk` `/index` carries SIX** — `index:1`–`index:6`, distinct keys, created
> > **2026-09-02 16:47:03Z**. Logged in `WRONG_CALLS.md`: a census goes stale by ADDITION and reads
> > as current for ever, which is why the entry carried the date I discarded.
> >
> > **⚠ AND THEN CORRECTED AGAIN, by `inline_guide_imagery`, because "driven" was also too strong.**
> > I wrote here that the mechanism was *live and DRIVEN*. **The honest state is ARMED AND
> > UNEXERCISED.** Rows seeded is the ASK; the branch running is the evidence, and they are not the
> > same thing. Verified here `[MEASURED 2026-09-02]`: every one of apis.uk `/index`'s six sections
> > still carries `page_components.updated_at = 2026-08-24 11:27:26Z` — *before* IMG-075 shipped and
> > long before the rows were seeded. **Nothing has re-resolved, so the new branch has still never
> > run anywhere.**
> >
> > The page IS capable, which I checked rather than assumed after briefly fearing otherwise: its
> > `slot_name` reads `generic-text-block` while the component behind `component_id` is
> > **`Illustrated Text Block`**, the only component in the estate declaring
> > `source: site_assets.illustration`. That plan-vs-rows spelling disagreement is precisely the
> > case IMG-075's round 2 was written to survive, and it holds.
> >
> > **⚠ So do not read apis.uk's next ordinary re-render as evidence either way** — its six figures
> > are already in `content_data`, so an assemble-only render emits identical bytes whether the
> > branch fired or did nothing. Evidence arrives at that page's next **`content_rewrite`**: the
> > exact event the mechanism exists to survive, on the exact page whose six values it was written
> > to protect.
>
> **The concrete blocker is the page shape, it is measured, and it is NOT confined to grip-styles —
> it is every content page on the site.** `[MEASURED 2026-09-02]` **all 22** content pages — 13
> `/blog/*` and 9 `/guides/tool-*` — have **exactly three** components, `hero` + `article-body` +
> `call-to-action`, and **zero** of them are illustration-capable. (`/guides/index.html` has two,
> `hero` + `content-listing`.) The whole of each guide is one `article-body`, which is §3.2's "one
> field owns prose and figures" in its purest form. **So this is not "re-plan grip-styles", it is
> "re-plan the site's content estate", and that is a materially bigger piece of work than §3.2
> assumed** — size it before promising it.
>
> The shape of the work per page: **re-plan into per-h3 sections on a component that sources
> `site_assets.illustration`** — exactly **two** exist estate-wide, `Illustrated Text Block` and
> `brief-explanation` `[MEASURED 2026-09-02]` — then seed `site_plan_imagery` rows at
> `scope='section'`, `kind='illustration'`, `scope_ref='<slot>:<ordinal>'`, one per figure.
> **grip-styles remains the right canary** (the owner named it, and IMG-075's own test fixtures are
> modelled on it), but do it as a canary, not as if it were the whole job.
>
> **⚠ AND CITE THE RIGHT REASON FOR DOING IT.** The justification is the owner's 08-31 ask — he wants
> an accurate image per small section, and 22 pages of banner-plus-prose is the complaint. It is
> **not** "give `inline_guide_imagery` a second driver to build confidence in their mechanism":
> that lane explicitly declined that as a reason, and they are right — it is a bad basis for
> touching 22 live pages. If the owner's ask ever goes away, so does the case for this work.
>
> **⚠ THE "LIVE 2026-09-01 21:00" DATE ABOVE IS ROUND 1 ONLY — corrected 09-02 by the owning lane.**
> `v1.0.1351` carried round 1. **Round 2's `sectionOrderAgrees` — the guard that stands the binding
> down on plan-vs-live disagreement, i.e. the thing that stops a silent MIS-bind — only rolled with
> `v1.0.1354`, pods up 2026-09-02 15:39:42 and 15:53:18.** So there was an unguarded window of about
> 18 hours. `inline_guide_imagery` probed at ~13:20 against `v1.0.1352` and correctly found it
> ABSENT; my probe found it PRESENT because it ran after the 15:39 roll. **Both readings were right;
> they are different binaries.** Nothing to redo — but do not cite "live since 09-01" for the guard.
>
> **THE PER-PAGE CENSUS, from `inline_guide_imagery` and re-derived independently here
> `[MEASURED 2026-09-02]` — they agree exactly:**
> - **9 of 13 `/blog/*` pages BIND today, and `grip-styles` is one of them** (plan 3 / live 3,
>   sequences matching): barrel-weight, beginners, board-setup, brand-comparison, flight-shapes,
>   **grip-styles**, shaft-length, steel-tip-vs-soft-tip, tungsten-guide.
> - **4 do NOT, and not because of drift — they have ZERO rows in the current plan** while carrying
>   three built components each: `barrel-shapes`, `checkout-chart`, `dart-balance`, `dart-points`.
>
> > **⚠ CORRECTED 2026-09-03 — I had the consequence right and the CAUSE wrong, twice over.**
> > I first wrote that those four were "exactly the four articles commissioned by
> > `SQL_2026-08-20_content_batch_week1.sql`" and that **that content route built four live pages
> > without writing plan sections**, implying our batch skipped a step other routes perform.
> > **No route performs it.** `[MEASURED 2026-09-03]` across the **32 sites that DO have a
> > substantive current plan**, being absent from the plan is the NORM for content page types and
> > the exception for structural ones:
> >
> > | page_type | built | in plan | absent |
> > |---|---|---|---|
> > | `blog-post` | 241 | 40 | **201 (83%)** |
> > | `tool` | 290 | 49 | **241 (83%)** |
> > | `guide` | 95 | 24 | **71 (75%)** |
> > | `content` | 126 | 103 | 23 (18%) |
> > | `section-index` | 52 | 45 | 7 (14%) |
> > | `landing` | 45 | 44 | **1 (2%)** |
> >
> > **So dartsonline is unusual in the OTHER direction, and the nine planned posts are the anomaly,
> > not the four unplanned ones.** All nine got their plan rows in a **single write at
> > 2026-07-29 13:28:03Z** — this lane's own `SQL_2026-07-29d_article_sections.sql`, a hand-written
> > backfill. Its header says why in its own words: *"Nobody ever decided what blocks a guide page
> > should contain."* Every article created since — **14** of them, the four from 08-20 plus ten
> > tool guides — has no plan rows, which is simply the estate default.
> >
> > **The real mechanism:** `create_blog_posts_action.go:183` encodes the canonical article layout
> > `["hero","article-body","call-to-action"]` **in the action**, and builds the page from it
> > without writing `site_plan_sections`. So the layout is real and the plan is empty.
> >
> > **Why this matters far beyond tidiness: every mechanism keyed on the plan is structurally
> > unavailable to those pages.** Per-section imagery binding degrades to page-wide when there is
> > no current plan row for the page (IMG-075's first stated degrade case) — so **~83% of the
> > estate's articles cannot carry a per-section figure at all**, whatever the planner composes.
> > That is a much harder bound on the imagery work than "the planner has never composed an article
> > out of illustrated blocks", and it is the thing to fix first.
> >
> > Adjacent and **actively owned — contribute, do not compete**: `bugs_open/443` (finetuning lane,
> > 8 commits in 2 days, Stage A closed 09-03) is the same root reached the other way — whole SITES
> > with no current plan (their census: 203 pages across 6 sites). **My population is disjoint from
> > theirs**: planned sites whose article page TYPES are unplanned. CONTRIB filed there.
>
> **⚠ SEQUENCE MATTERS, and getting it wrong looks like the mechanism failing.** The build path
> compares the plan against `pages.sections` (synced from the plan) so they agree during a rebuild
> and the binding engages; the **re-render** path compares against stored `page_components`. Between
> a recompose and the rebuild landing, a re-render therefore sees plan≠live and stands down —
> **correctly**, because the ordinals name a composition the page does not have yet. So:
> **recompose → seed rows → rebuild → verify → only then re-render freely.**
>
> **⚠ THE RISK HERE IS NARROWER THAN A RAW POPULATION COUNT SUGGESTS — and I recorded the raw count
> as if it settled the matter, which was wrong.** I first wrote here that only 9 active pages in the
> estate carry an illustration-capable component and 32 sites have none, and concluded *"there is no
> precedent to copy"*. **That reads a STOCK as if it were a verdict. Ask WHEN, and it inverts.**
> `[MEASURED 2026-09-02, re-derived here and agreeing with `inline_guide_imagery` exactly]`
> migration `644` taught the planner menus the word for an image on **2026-08-26 11:16Z**, and
> **6 of those 9 pages were composed AFTER it** — webdesign.uk 08-26 16:52 (five hours later),
> idea.uk/tools 08-28, lendzy 09-01, oufe 09-01, remortgagecalculator **09-02 12:45**, robot-hands
> **09-02 15:26**. **Two of them today.** The capability is not unused and ungated; it is being
> selected, and adoption is growing.
>
> **What IS missing is the arrangement, not the capability, and that is a much cheaper gap:**
>
> | | |
> |---|---|
> | pages with an illustrated section | **9** |
> | …`page_type='landing'` carrying **exactly one** (an accent) | **8** |
> | …`page_type='content'` carrying one | **1** (idea.uk `/tools.html`) |
> | …carrying **several** | **1** — apis.uk, and it was **hand-built**, pre-`644` |
> | …`page_type` `blog-post` **or** `guide` | **0** |
>
> So the planner has learned to reach for an illustrated block **once, on a landing page**, and has
> **never composed an article out of them**. **grip-styles is therefore asking the planner for
> something it demonstrably can do, in an arrangement and on a page type it has not been asked for**
> — a planning-behaviour question with a live and growing precedent to point at, not a missing
> capability and not the estate-wide recomposition that "1 of 442" implies. It is still a first, so
> the verification step stays; it is not the leap into the dark I first wrote down.
>
> **The transferable lesson, which cost a published diagnosis one lane over and a wrong emphasis
> here: a count of a population says nothing about whether that population is GROWING.** Two
> readings of one census an hour apart differed entirely on whether anyone had asked **when** the
> rows were created — and the timestamps were in the same table the whole time.

> **⚠ TWO THINGS THAT WILL MAKE THIS LOOK DONE WHEN IT IS NOT:**
> 1. **Grade it with a `content_rewrite`, NOT a page re-render.** Only `reason=image_landed` and
>    `reason=section_data_resolved` re-resolve; **every other re-render reason redeploys the stored
>    HTML unchanged**, so seeding rows and firing an ordinary re-render produces identical bytes
>    whether the binding engaged or did nothing. The save path is what this mechanism exists to
>    survive, so the save path is the test.
> 2. **dartsonline `index` currently STANDS DOWN** — `[MEASURED 2026-09-02]` its plan lists 6
>    sections and the live page has 4, and any disagreement stands the whole binding down by design
>    (`sectionOrderAgrees`, added in round 2 precisely because a mis-bound figure renders and
>    deploys looking correct). A re-plan the page has not been rebuilt from. Check the guide pages'
>    plan-vs-live agreement before seeding, or the rows will bind to nothing and read as inert.
>
> **Still true from the peer, and worth keeping:** a component field sourced `site_assets.image`
> does NOT give a section illustration — `imageRoleAliases` maps `image` → `hero` unconditionally,
> so it re-renders the page's own banner (`IMG-074`, and a `LANDMINES.md` entry). Use
> `site_assets.illustration`. That half of their message is confirmed, not stale.

Durable design for the imagery half remains
`inline_guide_imagery/PLAN_2026-08-14_durable_inline_guide_imagery.md` (plan-as-truth, unbuilt).
`grip-styles` is the natural canary; the other lane wants it at **P2**, not now.

**Accuracy is no longer the constraint** — current-setup generation is correct; only placement and
durability are open.

---

## 4. Affiliates — unchanged, and unblocked

Apply where the gate is **not** traffic: **Awin → Red Dragon** (10%, 60-day, £5 refundable deposit),
**Adtraction → Darts Corner** (free; owner logged in 08-24, waiting for a traffic figure),
**Paid On Results** (second route to Red Dragon). **Webgains → Target Darts** rejected the *network
account*, so it is a re-application once there is a number worth quoting. **Amazon UK** has no
traffic minimum but starts a 180-day / 3-sale clock — apply when there are readers to convert.
Privacy-page precondition is met.

---

## 5. Bugs this lane filed or contributed to

| bug | state |
|---|---|
| `bugs_open/384` — a landed card image never invalidates its listing | **fixed at the framework** (`requestPageListReresolve`). ⚠ my first mechanism was WRONG and is corrected in-file; read the correction block |
| `bugs_open/399` — CTA copy vs recorded destination never compared | **owned by another lane, in build.** Their measurement: the writer IS shown the destination and misdescribes it anyway, 155/1,060 = 14.6% |
| `bugs_open/410` — three seams fail toward the quiet default | filed; candidate 3 has a precedent guard in the same package |
| `bugs_open/198` — stylesheet clobber | dartsonline restored; 3 other sites restored by the owner |

---

## 6. Traps this lane paid for — read before touching anything

- **A `[MEASURED]` marker on a figure you were TOLD is a false claim.** I did this in 410 and had to
  correct it. Relayed numbers get attributed.
- **An inherited filter is the previous question still being answered** — it can only ever narrow,
  so the error is always an undercount and "it seems low" is the only symptom. Cost two wrong counts.
- **Cite the symbol, not the line number.** A line I cited in 410 expired within hours.
- **A nav rebuild refreshes stored chrome everywhere and republishes almost nothing** — 2 of 23 here.
  Fix with `docs/leopardessconsulting/scripts/reconcile_footer_nav.sh` (now has `--absent` and a
  three-valued predicate, both added by this lane).
- **`detected` items do not drain on this site.** File at `triaged`.
- **The B2 sync lands tens of seconds after the DB stamp** — a census taken straight after a deploy
  manufactures false negatives. I made this mistake twice.

---

## 7. What I would do next

1. ~~**Verify the four heroes at the served files.**~~ **DONE 2026-08-31 afternoon — and two of
   the four were still wrong.** Full account in `NOTES` (this date) and `README_where_we_are`.
   - **GOOD, unchanged since:** `hero-new-arrivals.jpg` (real flat four-wing flights at last),
     `hero-sale.jpg` (barrel macro, "24g" stamp).
   - **`hero-home.jpg`** had an all-red bull (the outer bull ring must be green). **Now FIXED
     and the best of the set** — measured 6,270 green px (5.9%) against 22 before, correct
     red/green doubles and trebles rings, striped flights, darts in the treble twenty.
   - **`hero-guides.jpg`** still carried a **feathered archery flight** — the owner's exact
     complaint — and a "7" where the board reads "1". **ALL FIXED after three passes**
     (13:25Z): flat moulded flight, bull correctly red-inside-green, rings alternating red and
     green (**6,454 green px, 7.8%**, against 2,867 in v1 and 0 in v2), and the number ring
     correct in **both** directions — 20,1,18,4,13,6,10,15,2,17 clockwise and 5,12,9,14,11
     anticlockwise. **All four heroes serve 200 on their live pages.**
   - ⚠ **Each re-roll trades an axis.** The three passes went feather+green → flight+no-green →
     flight+green. So **verify every axis after every roll, not just the one you were fixing**,
     and note the two 0%-green images (`new-arrivals`, `sale`) are *correct* — neither has a
     board in it, so a green floor applied across the set flags two good files.
   - **The cause was in our own stored descriptions, not the model.** `hero_home`'s prompt
     called the board "deep black and red" — two colours, neither green nor cream. `hero_guides`'
     said only "a single dart" and never said what a dart is. Both amended at the plan row.
   - **The structural gap, now closed:** every clause of this site's `imagery_style_guide`
     governed composition, palette or commercial claim, and **none said what the product looks
     like** — so re-rolling could never fix an anatomy error. Anatomy clauses added in
     `SEED_2026-08-31b`. ⚠ **They had to go in `kinds.hero` / `kinds.content_hero`, NOT the
     guide-level `avoid`:** `avoidForKind` returns the per-kind override *instead of* the
     guide-level list, so this site's heroes are governed by 111 characters while the
     652-character list is unreachable for them. A guide-level edit would have been inert and
     the re-roll would have credited it. Written up in `LANDMINES.md`.
   - **Every version's S3 object is recorded** in the headers of `SEED_2026-08-31b/c`, so any
     pass can be restored — each re-roll has traded one correctness axis for another.
2. **Re-measure crawler traffic around 2026-09-19** — the only honest read on whether the sitemap
   worked.
3. **Per-section guide images: take the flat-sections route** (§3.2) rather than waiting for
   composition — answered 2026-08-31, P5 is months away. The design question this lane still owes
   an answer on is whether any guide genuinely needs figure-*inside*-prose rather than
   figure-*between*-sections; if one does, that makes the guides a P2 consumer and `news_editorial`
   want to know.
4. **Apply to Awin and Paid On Results** — neither gates on traffic and both are one form.
5. **Fleet imagery:** other sites will carry the same orphaned-hero pattern. The discriminator in
   §3.1 is one query and should be run fleet-wide before anyone regenerates anything.
