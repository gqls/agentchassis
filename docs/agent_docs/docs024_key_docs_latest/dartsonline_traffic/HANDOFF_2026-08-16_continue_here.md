# HANDOFF — dartsonline guides, privacy, footer, consent. START HERE. Written 2026-08-16.

Cut because the originating chat's token load got high, not because anything is wrong.
**Nothing is on fire.** Everything below is either done-and-verified, or a clean next step.

Owning lane: `dartsonline_traffic`. Sibling lane spun out of this work:
`docs024_key_docs_latest/inline_guide_imagery/`.

---

## 0. The one-paragraph version

The owner asked for (a) an affiliate programme recommendation and (b) more imagery in the
guides. Both turned into something bigger. The imagery ask exposed a **live defect that had
silently deleted 4 of the 8 guide figures** and gone unnoticed for nine days; those are
restored and a durable design is written but unbuilt. The affiliate ask exposed a
**fleet-wide compliance gap** — 11 live sites run analytics with no cookie consent and 8 of
them have no privacy policy at all. Privacy copy for dartsonline is owner-approved but the
page is **not yet created**.

---

## 1. DONE and verified at the artefact (not at a status)

| thing | state | evidence |
|---|---|---|
| 4 guide figures restored | **LIVE**, survived the 08-16 chassis roll | `curl` each page → exactly 1 in-body `<img>` on flight-shapes, tungsten-guide, steel-tip-vs-soft-tip, beginners `[MEASURED 2026-08-16]` |
| Shipping & Returns out of the footer | **LIVE on 24 of 25 pages** | see §2 for the one remaining, which is expected |
| privacy copy | **owner-APPROVED**, committed, **NOT published** | `DRAFT_2026-08-15_privacy_copy_for_owner_approval.md` |
| bug 114 correction | committed | foot of `bugs_open/114` |
| 2 landmines | committed + synced to `doc_notes` | see §5 |
| durable-imagery design | written, unbuilt | `inline_guide_imagery/PLAN_2026-08-14_…md` incl. **§9 owner steer** |
| contact form + consent research | written | `RESEARCH_2026-08-15_contact_form_and_consent_banner.md` |

**Commits (all pathspec, this work only):** `450dffa52` `bf6e25b27` `5e796b125` `47c161f32`
`d70d7b089` `a40fe5209` `35ec187bb` `1c29a5558` `008b4de28` `9ec646a9a` `5e567796d`.

---

## 2. The ONE loose end from the footer job, and why it is probably fine

`/shipping-returns.html` **still shows the Shipping & Returns link in its own footer**
`[MEASURED 2026-08-16]`. All 24 other pages are clean. Two readings, and I did not settle it:

- it is the page's own self-link and simply was not in the redeploy batch, **or**
- the assemble-only rerender I filed for it completed against a page whose chrome was
  regenerated before the nav rebuild landed.

**It is very likely moot, because the owner should probably retire this page anyway** — see
§4 decision 3. Do not spend a cycle on it before asking. If you do fix it: one assemble-only
`page_rerender` (spec `{domain, page_id, page_name, filename}`, **no `reason` key**).

---

## 3. What is NOT done, in the order I would do it

### 3.1 Create the dartsonline privacy page (copy is approved, page does not exist)
`/privacy.html` → **404** `[MEASURED]`. The approved wording, the controller identity and
the "what it deliberately does not say" reasoning are all in
`DRAFT_2026-08-15_privacy_copy_for_owner_approval.md`.

> ⚠ **CORRECTED 2026-08-16 — I RELAYED A CLAIM AND IT IS WRONG. There IS a framework route,
> and THIS LANE HAS ALREADY USED IT.** The earlier text here said *"there is no framework
> path that adds one content page on demand"*. I took that from the `noted.co.uk` lane's
> commit message (`fb317132a`) and repeated it to the owner without checking it — exactly the
> relayed-claim failure `WRONG_CALLS.md` exists for. Verified from source 2026-08-16:
>
> - **`apply_gap_plan` has a `new_page` approach whose own header says it "creates page
>   record + needs_content_page work item"** (`apply_gap_plan_action.go:8`, branch at `:140`,
>   implementation `applyNewPage` at `:289`). So `needs_content_page` needing an existing page
>   is true *of that item type alone* and false *of the chain* — `new_page` creates the page
>   first and then files it. That is the sentence I misread.
> - The action runs in the live **`content-gap-planner`** agent `[MEASURED]`, fed by
>   `needs_content_planning` items which are **actively draining — 27 `complete`, newest
>   2026-08-15** `[MEASURED]`.
> - **The precedent is on this very site**: `SQL_2026-07-29n_news_page.sql` in this directory
>   raised a planner item with `'approach','new_page'` asserted in the spec, and
>   `/news/index.html` exists and serves 200 today. So the route is proven here, not just in
>   principle.
>
> **Caveat to carry:** `approach` is read from the *plan* the LLM produces
> (`apply_gap_plan_action.go:127`), so the planner could choose `add_to_page` or
> `not_actionable` instead. Asserting `approach` in the item spec is what 29n did and it
> worked — treat it as a strong steer, not a guarantee, and check which branch actually ran.
>
> The adoption-route mirror below remains a valid fallback, but it is **no longer the first
> choice** — prefer the planner route, which is the framework doing its own job.

Steps, in order:
1. Register the approved copy in `evidence_base` as `supplied_copy.privacy`, **deriving** the
   new spec row from the live one (`data || {...}`) so existing `banned_claims`/facts carry
   across untouched. Pattern: `noted_rebuild/apply_privacy_copy.py` — **it is noted-specific
   and needs parameterising**, do not run it as-is.
2. Create the page via **`needs_content_planning` → `content-gap-planner`**, `approach:
   new_page` asserted in the spec, modelled on `SQL_2026-07-29n_news_page.sql`. File it at
   `status='triaged'` (`detected` does not drain on this site — §5). Fallback only if the
   planner refuses: mirror the adoption path (`apply_adoption_plan_action.go:541`,
   `INSERT … ON CONFLICT (site_id, name)`), taking every value from the current plan so no
   identity is hand-rolled (`bugs_open/080`).
   Convention from 7 other sites `[MEASURED]`: `name='privacy'`, `url='/privacy.html'`,
   `page_type='content'`, `sections=["generic-text-block"]`, `in_header=false`, `in_footer=true`.
   ⚠ `page_type` is a routing key, not a label — verify the row before it builds
   (`bugs_open/081`: a deployed mistyped page has no repair path).
3. `nav_drift` → nav-updater for the footer link, **then §5's landmine: verify the SERVED
   bytes on every page, not `pages.rendered_footer`.**
4. Verify the served page carries the copy **verbatim**.

### 3.2 The component hierarchy feature plan — owner asked for it, NOT written
Owner 2026-08-15: *"I would like to implement the component hierarchy - maybe write a plan
and put it in the features_open directory for future development."*

**Not done: three Fable agents died on model/session limits.** `035_FEATURE_component_hierarchy.md`
is the reserved slot (confirm it is still free). The owner specifically wanted **Fable** for
the design work — do not silently substitute a model; ask.

The brief is worth reusing verbatim; its load-bearing measurement is:
**`page_components.parent_instance_id` exists and is used by ZERO of 1580 rows fleet-wide**
`[MEASURED 2026-08-15]`. So composition is real in the schema and has never once been
exercised — adopting it is build-and-prove, not wiring, which **inverts the phasing and
probably the architecture-scope call** in the existing plan. Full context + the agent roster
he named: `inline_guide_imagery/PLAN_2026-08-14_…md` §9.

### 3.3 Consent banner — 11 sites, and half of it is the owner's
See the research doc. Two separable layers: configure Consent Mode in `GTM-PQ3WCTBD` (one
change, 11 sites, **needs the Google account — owner action**), and the banner UI in shared
chrome (**stored artefact**, so every site must be re-rendered AND redeployed — §5).

### 3.4 Contact form — one shared component on 11 sites
`tools-api` is deployed and a sound host but has **no enquiry endpoint** (its whole route
table is vonc-gauntlet). **Read whatever serves idea.uk's `/request` and `/audience-check`
first** — they are the only working form endpoints in the estate and I never traced them.
That is the single most useful unknown in this handoff.

---

## 4. Owner decisions outstanding

1. **File the consent gap as a bug?** It is durable and biting 11 live sites, which is
   exactly `bugs_open/`'s question — but the remedy is legal/business, so I did not file it
   unilaterally.
2. **Which model for the component-hierarchy plan**, given Fable keeps hitting limits.
3. **Retire `/shipping-returns.html` entirely?** He asked only for the footer link, which is
   done. The page is still live and indexed at **200**, still claiming shipping/returns on a
   site whose own identity spec says it holds no stock — and an affiliate reviewer reads
   exactly these pages.
4. **Postcode for the privacy page.** Controller given as *Fine Tuning, Fleetside, West
   Molesey, East Surrey*; no postcode supplied. Left exactly as given rather than guessed.

**Already decided, do not re-ask:** grip photo goes next to the grip section (done); the
contrast fix may go fleet-wide (he told that thread himself); publish the privacy copy now
and rewrite when the affiliate networks supply their wording; business identity as above.

---

## 5. Traps this work paid for — read before touching chrome or guide imagery

Both are in `LANDMINES.md` (committed + synced), but the two that will bite soonest:

- **An image inside guide prose is destroyed by the next body rewrite, silently.**
  `article-body` has ONE llm-owned `content` field, so figure and words are the same
  overwritable blob. **The 4 restorations in §1 are a stopgap and the next rewrite undoes
  them.** Never conclude "never wired" from a present-tense scan — `page_component_history`
  distinguishes *never placed* from *placed and overwritten*, and that mistake already cost
  me a wrong bug filing.
- **A `nav_drift` rebuild refreshes stored chrome on EVERY page and redeploys only SOME.**
  Measured here: item `complete`, **0 of 25** pages stale in the DB, **19 of 25 SERVED pages
  still stale**. The 6 that had redeployed were the homepage and the two index pages — i.e.
  exactly what a person spot-checks. Grade at the served bytes over the whole sitemap.

Also live and relevant: `detected` work items **do not drain** on this site (mine sat 10
hours); file at `status='triaged'` instead. And `bugs_open/274` (~15,000 completed workflows
whose results never reach their parents) bears on any design that waits on a child agent.

---

## 6. Affiliate recommendation, so it is not lost

- **Target Darts via Webgains** — 8% commission, 30-day cookie, apply inside a free Webgains
  account. Best fit for the guides' subject matter.
- **Darts Corner via Adtraction** — 13,000+ products, good for comparison guides; commission
  not published pre-application; a second network account.
- **Blocker the owner accepted:** privacy page first (§3.1). The consent gap (§3.3) is the
  one I would also fix before applying.
