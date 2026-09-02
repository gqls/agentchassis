# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-09-02)

**Supersedes `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/HANDOFF_2026-08-21_continue_here.md`**
(and its 08-24 update block). Nothing in that file needs reading except its §5 traps, which still hold.

**The owner asked for two things and BOTH need re-aiming. Read §1 and §2 before doing either.**
Everything below was measured **2026-09-02**, against a site that six other lanes have been working
while this one was away — 46 pages (42 deployed), unlocked, 0 items armed by us.

---

## 1. ⚠ "Fix the dead calculators" — THEY ARE ALREADY FIXED. Do not start here.

On 2026-08-26 this lane reported two tool pages loading a 404 calculator script. **Both are
resolved.** Measured today:

| script | today | referenced by any page? |
|---|---|---|
| `/tools/assets/tool-equity-release.js` | **200** | no — page now uses inline JS |
| `/tools/assets/tool-btl-investor.js` | still 404 | **no — 0 live, 0 in stored `page_components`** |
| `/tools/assets/mortgage-lender-directory-listing.js` | 200 | yes (1 page) |
| `/assets/js/snippets.js` | 200 | yes (all 42) |

Both tool pages now carry **~6 KB of inline JS** with real inputs and buttons instead of an external
file. Another lane did this (19 `improve_tool` complete, 10 `audit_tool` complete, 6
`acceptance_run` complete since 08-26).

**Item `d5131e25` was closed `complete` on 08-27. Item `a7c5d5ab` (btl-investor) is STILL `detected`
and is STALE** — its condition is gone (0 references anywhere). Closing it is safe and is the only
"fix" left on this front. ⚠ This lane has now hit a stale-open-item three times on this site
(`unbuilt_internal_link` 08-21, this one) — **a parked item is not evidence its condition survives;
re-probe before acting.**

**What IS unresolved on the tools:** two `improve_tool` items are `failed` (08-27) —
`tool-deposit-tracker` and `tool-remortgage-savings`, both *Tier-4 acceptance:
calculate-shows-results@desktop*. **Both failed on INFRASTRUCTURE, not on the tool:**
`step load_tool failed: … query_database: query param path 'input_da…'`. So those two tools'
acceptance is **unverified**, not failed. That is the real open question here.

## 2. ⚠ "Wire up the existing images" — WIRING IS NOT THE BLOCKER. Doing it will change nothing.

The obvious job — point each tool page's hero at its own `content-hero-tool-*.jpg` — **has already
been done on three pages and made no difference.** This is the finding to start from:

| page | `content_data.background_image` | hero renders a background? |
|---|---|---|
| `tool-equity-release` | `…/content-hero-tool-equity-release.jpg` | **NO** |
| `tool-overpayment` | `…/content-hero-tool-overpayment.jpg` | **NO** |
| `tool-simple` | `…/content-hero-tool-simple.jpg` | **NO** |
| `tool-affordability`, `tool-repayment`, +6 more | `…/hero.jpg` | **NO** |
| **`tool-*-guide` × 8** (same component, same field) | `…/hero.jpg` | **YES** |

**Same `hero` component, same field, same value shape — renders on guide pages, not on tool pages.**
So the data is not the problem and the template is not the problem: the template *does* emit it
(`{{if or .hero_url .background_image}}background-image: … url('{{or .hero_url .background_image}}')`,
verified in `content_components.html_template`). The value is in `content_data` and **absent from
`rendered_html`** — so it is not reaching the render context on this page type.

**Note the field's source: `background_image` is `source: site_assets.hero`** — a resolver-populated
field, not an LLM one. Precedent worth reading first: MEMORY `the-framework-writes-the-content-not-you`
records a wasted run from asking the writer to set a resolver-owned URL. **Check who populates the
field at render time before editing anything.**

**Start here:** diff the render path for a `tool-*` page against a `tool-*-guide` page — same
component, one renders, one does not. That is a narrow, reproducible handle.

⚠ **Do NOT size this by the images.** 14 deployed pages have no content image and **all 14 are tool
pages** (28 with / 14 without, measured 09-02 — *identical* to 08-26). 12 `content_hero` images and
6 tool `card` images exist and serve 200 (115–142 KB) and are referenced by nothing: that is
**`bugs_open/114_HANDOFF_2026-07-27_generated_imagery_is_deployed_and_never_referenced.md`**, OPEN
since 2026-07-27, filed off the owner's identical complaint ("there is not enough imagery on the
site").

## 3. The measurement trap that nearly cost this handoff a false alarm

Re-running the image audit with a **different extractor** (dropping CSS `url()`) produced
"6 pages with images, 36 without" against last week's 28/14 — a large fake regression. With the
**same** extractor it is 28/14, unchanged. **When you re-measure to compare, re-run the ORIGINAL
query, not a new one that answers the same question differently.** Distinct images referenced did
rise 9 → 12.

Likewise `snippets.js` returned one 404 among 13 probes. It is fine (12× 200). **A single probe is
not evidence** — this lane's third transient false-negative in two weeks.

## 4. State, measured 2026-09-02

- **46 pages, 42 deployed.** Site **unlocked**. **0 items armed by this lane.**
- Open queue: **58 `needs_human_review`, 48 `detected`.** The `detected` rows include the 38 filed
  by the 08-26 design-discovery run (7 `needs_imagery`, 12 `undeployed_asset`, 6 `improve_tool`,
  6 `audit_tool`, 1 `deactivated_component`, 1 `needs_rerender`).
- **12 `needs_imagery` items still DEFERRED since 2026-08-02** — 7 page heroes, 4 section icons,
  1 infographic, never generated. Per the 307 lane's CONTRIB (2026-08-26), **this lane's own
  15-second auto-defer backstop parked them.** Re-arming them is real image-generation spend and is
  the owner's call.
- ⚠ **11 of the 12 `undeployed_asset` items are MISLABELLED and 1 is a false positive** — all 12
  URLs serve 200. Evidence filed into
  `bugs_closed/142_HANDOFF_2026-07-29_undeployed_asset_detector_cannot_see_missing_artefacts.md`
  (the `logo` one is defect #2 of that closed bug, still live because the fix enumerated two
  purposes instead of testing the property). **Do not act on those 12 as written.**
- `site-discovery-rotation-design` is **enabled** now (10,800 s). The 08-18 handoff recorded it
  disabled and "the owner's separate call"; that changed elsewhere, not here.

## 5. Unread/unactioned inbound

- `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/CONTRIB_2026-08-26_from_copy_quality_your_brief_carries_a_fossil_of_the_OLD_house_voice.md`
  — this site is **the only one of 31** whose brief carries pasted old house-voice text that a
  canonical rewrite cannot reach. They offer to fold it into a fleet sweep if we say so. **Needs an
  answer.**
- `…/CONTRIB_2026-08-24_from_288_I_ADDED_AN_ARTIFACT_CHECK_TO_YOUR_REGISTER.md` — unread.
- `…/CONTRIB_2026-08-21_from_the_307_lane_your_item_was_flagged_and_then_overwritten.md` — actioned
  (it corrected `bugs_open/348`; `bugs_open/344` owns that mechanism).

## 6. Still open from earlier arcs

`/scorecard-simulator.html` is still the site's one dead internal link — the writer reliably
mistypes `mechanism-flow.steps[].branches`, which is `bugs_closed/260`'s **writer** half, owned by
`copy_quality_two_stage` (case handed over 2026-08-21, reproducible on demand). Owner decisions
outstanding: the 13 `fact_drift_review` items; a business email for the contact page (item
`07bc64cd`); and whether the contact page's `<title>`/footer label should stop saying "Contact".

## 7. Files of record

`NOTES_mortgagecalculator_couk.md` `## 2026-08-21`, `## 2026-08-24`, `## 2026-08-26`, `## 2026-09-02` ·
`README_where_we_are.md` (owner's log) 2026-08-21, 2026-08-24 ·
`bugs_open/348…` (mechanism corrected; `bugs_open/344` owns it) ·
`bugs_closed/142…` (evidence appended 08-26) · `WRONG_CALLS.md` 08-21 ×2, 08-24.
