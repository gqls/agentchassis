# 200 — The mobile hamburger renders as ONE bar on every layout in the library: `inline-flex` lays the three spans in a row

**Filed** 2026-08-04 (late evening), by the bugfix_188 close-out thread. The
observation is the brochure component library lane's, from the 2026-08-03 contact
sheet look (recorded in `bugs_closed/188` §3, TL-035:351, and the vigilant
CONTRIB §7 — "may be a genuine page defect rather than a capture artefact").
It is a genuine page defect, and it is the whole layout library, not one page.
**Status** OPEN.

---

## 1. Symptom

On every generated site at mobile width, the header's menu toggle draws as **one
long thin bar** instead of the three-bar hamburger. Visitors do not read a single
2px line as "menu"; on a phone the site's navigation is effectively invisible
(the button still works — the defect is purely visual, but it is the visual that
tells a visitor the button exists).

Evidence, at the artefact:

- `tool-review-council-simulator` mobile render from acceptance run `25c44133`
  (2026-08-04, **landing state** — so this is the page as served, not a
  driven-state artefact; the same defect appears in the 08-03 driven-state
  render, so it is not interaction-dependent either). The header shows the logo
  and one wide white bar, nothing else.
- The bar measures **~217 device px wide at 3× ≈ 72 CSS px** — see §2 for why
  that number is the fingerprint.

## 2. Root cause — CONFIRMED first-hand; the arithmetic is the fingerprint

The markup is correct everywhere it was checked — three spans, the standard
pattern:

```html
<button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
    <span></span><span></span><span></span>
</button>
```

The library CSS is not. Every layout's `css_template` carries (desktop):

```css
.mobile-menu-toggle { display: none; ... min-width: 44px; min-height: 44px; ... }
.mobile-menu-toggle span { display: block; width: 22px; height: 2px; ... margin: 5px 0; }
```

and in the mobile media query:

```css
.mobile-menu-toggle { display: inline-flex; }
```

**`inline-flex` with no `flex-direction` makes the button a row flex container.**
The three spans become flex items laid out **side by side**: three 22-24px × 2px
bars with only vertical margins fuse into one ~66-72px × 2px line. On
fundamentallyai.com the page's own chrome `<style>` block overrides the span
width to 24px, and the observed bar is ~72 CSS px — **exactly 3 × 24**. The
observed width being precisely three span-widths is what confirms the layout
mechanism over any alternative (paint glitch, missing spans, font fallback).

**Why no `090` diagnosis run was filed — the substitution, stated per the owner
ruling of 2026-07-31:** the mechanism is directly observed end to end, with a
disconfirmable measurement at each link: (a) the rendered pixel artefact on two
independent runs (08-03 driven, 08-04 landing); (b) the served CSS of **four**
live sites read first-hand (fundamentallyai.com `styles.css:425`, finetuning.uk,
idea.uk, webdesign.co.uk — all carry the identical rule); (c) the arithmetic
identity above; (d) a full census of the live source (§3, could have come out
otherwise and didn't); (e) two negative controls — the repo's two fallback
headers use `display: block` and render correctly, and both stored `site-header`
components carry `flex-direction: column`, which is why *component-rendered*
headers are not affected. There is no code-path question left for the loop to
answer; the deciding evidence is pixels and served CSS, both already fetched.

## 3. Blast radius — the whole layout library, since birth

- **Live census 2026-08-04:** `SELECT count(*), count(*) FILTER (WHERE
  css_template LIKE '%mobile-menu-toggle { display: inline-flex%') FROM layouts;`
  → **18 / 18**. Every row.
- The seed files match (`docs/agent_docs/docs024_key_docs_latest/layouts/`,
  17 layouts across 16 `layout_*.sql` files, all carrying the rule; tracked and
  clean at HEAD, unchanged since April — `c6ce21d65 "css composition templates"`,
  moved by `4d5c263cd`). The defect is congenital to the library, not drift.
  (Trap for the next reader: a plain `git log -- <the docs024 layouts dir>`
  prints nothing — follow the FILE, `git log --follow -- .../layout_04_...sql`,
  which shows the docs020→docs024 move. A quiet git log is not silence.)
- Served and confirmed on four live sites' `styles.css` (list above). Any site
  whose stylesheet was rendered from any of the 18 layouts serves it.
- **What is NOT affected:** headers rendered from the stored `site-header`
  components (`header-theme-chrome`, `header-leopardess` — both set
  `flex-direction: column`) and the two Go fallback headers
  (`component_library.go:2002`, `multipage_actions.go:1556` — both use
  `display: block`). The defect lives in exactly one place: the layout
  `css_template` chrome rules, composed into `styles.css` by
  `render_css_composition_loader.go:106`.

## 4. Why nothing catches it

No acceptance check asserts toggle geometry — the fences assert what tools
compute, `no-horizontal-overflow`, console errors. A 2px-tall bar overflows
nothing and errors nowhere. This is precisely the "faults only an eye catches"
class TL-035 was built for, and it worked: the defect was found by a human
reading the first contact sheet, and confirmed on the first landing-state
render after `bugs_closed/188` fixed the camera's timing.

## 5. Fix candidates, ordered by what closes the door

1. **Fix the library rule in the seed files and re-run the seed driver.** Add
   `flex-direction: column; justify-content: center; align-items: center;` to
   the toggle's mobile rule (or to the base rule — inert while `display: none`)
   in all 16 `layout_*.sql` files, then
   `003_layouts_seed_driver.sql` (`ON CONFLICT (name) DO UPDATE` refreshes
   `css_template` — the driver's own header says re-running is the intended
   refresh path). Then rerender `styles.css` per site. Closes the door for
   every current and future site; no Go change, no image roll — but note the
   17 copies stay 17 copies, so:
2. **The follow-on worth an owner thought, not decided here:** the toggle/chrome
   rules are duplicated across all 17 layouts (this bug is the proof that copies
   drift together only until they don't). Whether header chrome CSS should be
   composed from one shared block rather than pasted per layout is a
   design-composition question for that lane (DES-038's territory), not a patch.
3. **A durable guard:** a `pattern-check` (or seed-lint) rule — any
   `.mobile-menu-toggle` (or hamburger-class) rule that sets `display:*flex`
   must set `flex-direction: column` in the same file. Cheap, and it makes the
   regression unrepresentable at the seam where it was introduced.
4. **Do nothing on the page side** — the button works (the tap target is the
   44px button, wired by the chrome script). Rejected: a control a visitor
   cannot recognise is not a control.

Candidates 1+3 together are the recommendation. Component-rendered headers need
no change.

## 6. How to verify

- After the seed refresh + a site's `styles.css` rerender: fetch the mobile
  render of any acceptance run (or `curl .../assets/css/styles.css` and check
  the rule), and the header crop must show **three separate bars** (~22×2px
  each, ~5px gaps), not one.
- Negative direction: the desktop render must still show NO toggle
  (`display: none` outside the media query must survive the edit).
- The census query in §3 must return 18 / **0**.

## 7. Provenance

- Observation: brochure component library lane, 2026-08-03 contact sheet
  (commit `1f375991f`'s sheet; recorded in `bugs_closed/188` §3 as "may be a
  genuine page defect", explicitly out of 188's scope).
- Mechanism, census and this file: bugfix_188 close-out thread, 2026-08-04,
  first-hand (served CSS + rendered pixels + live `layouts` census + seed
  files at HEAD). Queue checked (no open item mentions the hamburger);
  `bugs_open/` + `bugs_closed/` grepped (041 is idea.uk's *dead menu JS*,
  a different defect on a different seam; 117/170 chrome bugs are about
  chrome *storage*, not the layout CSS).
- Not routed at a lane: `who-owns` shows the layout seeds untouched since
  April and no thread active on them. The natural owners are the
  design-composition lane (the library) and whoever runs the next site
  rerender sweep.

---

## 8. FIX RECORD 2026-08-05 — source fixed and guarded; ONE step remains, and it is permission-gated

**Done, verified:**

1. **All 18 live `layouts` rows fixed** — seed
   `314_hamburger_toggle_gets_flex_direction_column.sql`, a surgical `replace()`
   (exact-hit count pre-measured at 1 per row) with a DO/RAISE verify that was
   **induced against the pre-fix state first** (raised "18 rows still carry…").
   Applied 2026-08-05, `UPDATE 18`, guard passed.
   **The seed driver was deliberately NOT run** — five rows carry 2026-07-02
   live-only changes the seed files never got, and `tool-portal-light` has no
   seed in the layouts dir at all. A driver refresh would have silently
   reverted all five. This is now a LANDMINES entry ("the layout seed driver's
   're-running is safe' header is FALSE for six of eighteen layouts").
2. **All 17 layouts in the 16 seed files fixed** (commit `123d65198`) so a
   future driver run cannot regress the live fix.
3. **Guard armed**: `check_flexless_hamburger` in `scripts/pattern-check.py` —
   fires on any changed .css/.sql/.go/.html where a menu-toggle/hamburger rule
   sets `display:*flex` without `flex-direction` in the same block. Induced
   through the real hook path (planted staged file → fired at the right line;
   fixed seeds → silent).

**Remaining — the 15 live sites still SERVE the old stylesheet.** Survey
2026-08-05 (curl, all 20 external domains): 15 serve `styles.css` with the bare
rule exactly once (ai-agent-orchestration.com, dartsonline.com, finetuning.uk,
fundamentallyai.com, gamesdesign.co.uk, gaswholesalers.com, idea.uk,
leopardessconsulting.co.uk, mortgagecalculator.co.uk, oufe.com, relojistas.com,
robot-hands.com, vetcomparison.uk, vonc.com, webdesign.co.uk — all whole files,
13–26 KB, no clobbered fragments); the three loan* sites 404 on
`/assets/css/styles.css` (different asset layout — nothing to patch on this
path); lendzy.co.uk and webdesign.uk do not resolve/serve.

**Why the pipeline is NOT the way to apply it:** regenerating a stylesheet
through the design pass re-rolls the palette (`deploy_stylesheet_direct.sh`'s
own header, measured: `analyze_design`'s fresh `color_scheme` WINS over the
palette row — the colour-churn landmine). The correct application is the same
surgical replacement on each site's SERVED file, published via the brochure
lane's proven direct git-adapter path. All 15 patched files are prepared and
byte-checked (+70 bytes each, exactly one rule changed).

**The publish itself was blocked by the session's permission classifier**
(an outward-facing write to 15 production site repos — reasonably a human
call). Operator: run the prepared batch —
`bash <scratchpad>/deploy_bug200_all.sh` (deploys all 15 via Kafka → git-adapter,
waits 90 s, then diffs every SERVED file against its patched copy). Re-runnable;
per-site verify is `curl -s https://<domain>/assets/css/styles.css | diff -
<patched file>`. If the scratchpad is gone, the procedure is: fetch served
styles.css, apply the §2 replacement (assert exactly 1 occurrence, +70 bytes),
publish via `brochure_component_library/scripts/deploy_stylesheet_direct.sh`
with an honest commit message, verify served == patched.

**Close when:** all 15 serve the patched file AND one mobile render (any
acceptance run) shows three bars — then §6's negative check (desktop still
hides the toggle) and move this file to `bugs_closed/`.
