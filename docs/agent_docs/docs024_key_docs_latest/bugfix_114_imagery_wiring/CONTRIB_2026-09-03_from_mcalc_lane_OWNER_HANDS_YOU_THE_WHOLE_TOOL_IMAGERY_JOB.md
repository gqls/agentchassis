# CONTRIB 2026-09-03, from `mortgagecalculator_couk_adoption` — the owner has handed you the WHOLE tool-imagery job for this site. Here is everything, including the half you cannot see from your side.

**This is a handover, not a request.** The owner's instruction (2026-09-03): *"hand that lane the
whole job."* So mortgagecalculator.co.uk's tool imagery is yours — the 18 tool pages, both
mechanisms, and the decision about spend. **We are not working it and will not dispatch anything at
it.** Everything below is measured; where it is not, it says so.

## 1. What you already have from us

`bugs_open/114` carries our CONTRIB of 2026-09-02: **`hero-tool` declares no image-typed field**, so
`plan_sections_action.go:2846`'s `sectionHasImageField` gate is false, the per-page hero is never
written into `resolved_data`, and the template falls through to the site-wide default. **69
instances, 0 with a per-page image, against `hero`'s 632/72.** 54 of the 69 have an active
`ContentHeroKey` asset going unshowable. Four-line fix, mirrors what `hero` already declares.
That analysis stands unchanged.

## 2. What has changed since, and it changes your job here

**Migration `701` landed overnight** (owner-applied ~22:00Z 09-02; `bugs_closed/357`, population 0).
It retyped this site's 11 adopted rows to `component_level='tool'` with proper per-tool components.

**That removes the blocker we told the owner about.** On 09-02 we said ten of the fourteen
image-less tool pages *had nowhere to put a picture* — their `hero` slot held the whole calculator,
and rendering a hero into it would have destroyed the tool. **That is no longer true.** The
calculator now lives in its own `tool`-level component, and the `hero`/`hero-tool` slot is free.
**The composition change you would need is now safe**, and it was not 24 hours ago.

⚠ **One live consequence you must know before you re-render anything here.** 701 created every
adopted component with an instance-scoped template (`{{.InstanceID}}-`), but preserved the existing
rendered bytes. **A re-render therefore rewrites every element id on the page.** Five of the ten
re-rendered this morning (09-03 08:46–08:49) and their ids went `amt` → `c-tool-simple-amt`. We
checked: **0 dangling JS bindings on all ten**, so the tools still work — the converter rewrites
bindings with the ids. But it silently invalidated five acceptance fences (`bugs_open/441`), and
**any page you re-render for imagery will do the same to that tool's fence.** Not a reason to avoid
re-rendering; a reason to tell us which pages you touch, so we re-point the fences after.

## 3. The state of the imagery on this site, measured

- **All 18 tool pages have an active asset at exactly their `ContentHeroKey`**
  (`'content_hero_'||replace(page_name,'-','_')`) — generated, deployed, serving 200. Verified row
  by row 09-02. **Nothing needs generating.** This is a wiring job, not a spend decision.
- **14 of 18 display no image.** The four that do (`btl-investor`, `credit-health-check`,
  `overpayment-priority`, `rate-stress-test`) carry the generator's four-section shape, and their
  own content hero appears via `tool-guide-intro`'s `guide_image_url` (`source: site_assets.image`)
  — **so the per-page resolver demonstrably works on this site**; `hero-tool` just never asks it.
- **Their `hero-tool` band shows `/assets/images/hero.jpg`**, the site default, on all four. That is
  §1's defect.
- `tool-cta` renders 6 distinct card thumbnails and is fine. ⚠ We briefly recorded it as broken —
  that was our extractor reading only the first `<img>`. It is not a defect; don't re-investigate it.

## 4. Two traps that cost us time here

1. **`LIKE '%background-image%'` matches the CSS PROPERTY NAME in a `<style>` block.** It reported
   every `hero-tool` row as carrying an image; they carry a solid colour. Test the value
   (`rendered_html ~ 'url\(''?/assets'`), not the property.
2. **Do not hand-set `content_data.background_image` per page.** It works — one
   leopardessconsulting page does exactly that and is the only `hero-tool` instance fleet-wide
   showing a per-page image — and it is hand-setting a resolver-owned URL (MEMORY
   `the-framework-writes-the-content-not-you`). It also hides §1 from the next reader.

## 5. What we think the job is, offered as a view and not a spec

Give the ten newly-freed tool pages the composition the four healthy ones already have — a
`hero-tool` band ahead of the calculator — and land §1's schema fix so that band resolves each
page's own `content_hero_*` rather than the site default. Then the same change serves all 18 and the
other 20 sites. **But the sequencing, the spend and whether a hero belongs above a calculator at all
are yours to decide now, not ours.**

⚠ **The adopted calculators already carry their own `<h1>` and description** (`<div
class="tool-page"><div class="tool-header"><h1>…`). A hero band above that duplicates the title
unless the tool's own header goes. Worth a look before you plan the section.

## 6. Where to reach us / what we still hold

We keep the **tools as product** — acceptance fences, Tier-4 verification, `bugs_open/441`, `448`
and `449`. If your work re-renders a tool page, that is fine; just say which, and we will re-point
its fence.

Evidence and full working:
`docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`
`## 2026-09-02 (b)`, `(c)`, `(d)` and the 09-03 entry ·
`HANDOFF_2026-09-02b_continue_here.md` §3 · our CONTRIB in `bugs_open/114`.

— `mortgagecalculator_couk_adoption`
