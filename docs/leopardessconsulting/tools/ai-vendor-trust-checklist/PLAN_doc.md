# PLAN — ai-vendor-trust-checklist (leopardessconsulting.co.uk)

## What it is

Twelve yes/no items an outsider can actually check about an AI vendor's data-handling
posture, grouped as certifications, data handling, governance and oversight, and
transparency. Ticking items updates a score, a denominator and one of three
plain-English verdict tiers. Every unticked item shows a one-line note explaining why
it matters. It is a conversation-starter, not a certification, and says so.

Built as the first end-to-end exercise of the S0-S7 staged build ladder
(`features_open/027`, lane `staged_component_build`).

## How it is delivered

`content_components` row rendered as a page section on `/tools/ai-vendor-trust-checklist.html`,
with the interactive logic in a separate static asset at
`/tools/assets/tool-ai-vendor-trust-checklist.js`. Entirely client-side: no fetch, no
framework, no LLM call, no server round trip. Recomputes on every `change`.

## The contract

Template fields (12, all supplied by hand in `content_data`): `section_aria_label`,
`badge_label`, `headline`, `subheadline`, `list_panel_title`, `results_panel_title`,
`score_label`, `cta_url`, `cta_label`, `reset_label`, `disclaimer_text`,
`source_link_label`.

Element ids the JS depends on: `#vtc-c1`..`#vtc-c12` (each also carrying
`data-vtc-item`), `#vtc-na-sector`, `#vtc-reset`, `#vtc-score-count`,
`#vtc-score-total`, `#vtc-meter-fill`, `#vtc-verdict-box`, `#vtc-verdict`,
`#vtc-verdict-detail`, `#vtc-gaps`.

Tier thresholds are RATIOS, not counts, because the denominator becomes 11 when the
sector certification is marked not applicable: `>= 0.75` strong, `>= 0.4` middle,
below that gaps. 9/12 and 9/11 both read strong; 4/12 and 4/11 both read gaps.

## Hazards and their answers

1. **The template/JS id contract spans two files and no platform check can see it.**
   `deploy_tool_action` calls `datahelpers.OrphanElementRefs`, which finds
   `getElementById`/`querySelector` references in the page and asserts the ids exist.
   For this tool it returns nil and passes, because the JS is an external asset so the
   template contains no references. *Answered:* check J in `render_harness.go` parses
   the JS and asserts every id it queries exists in the rendered markup, plus that the
   `data-component` value the JS looks for matches the one the template emits. This is
   the same defect class that left `llm-cost-calculator` pointing at
   `bayesian-ranking-hero-tool.js`, one level deeper.

2. **A wrong `<script src>` is invisible until a human clicks something.**
   *Answered:* check F asserts exactly one script tag with the exact expected path, and
   the mutant `wrong-script-filename` proves the check can fail.

3. **The N/A control is an `input[type=checkbox]` inside the same section and could be
   counted as a thirteenth item.** *Answered:* the JS selects only
   `input[type="checkbox"][data-vtc-item]`, and check D asserts the N/A box carries no
   `data-vtc-item`. Mutant `na-becomes-an-item`.

4. **A stale tick on the sector item could be counted after it leaves the
   denominator.** *Answered:* the JS clears AND disables that box when N/A is on.

5. **An empty denominator would make the ratio NaN and still produce a
   confident-looking verdict.** *Answered:* guarded explicitly in `recompute()`. It
   cannot happen with the shipped markup, which is why it is handled rather than
   assumed.

6. **A page-level placement is dropped by the next re-render while the work item
   reports `complete`.** *Answered, and the answer differs from the brochure lane's:*
   this site's `site_plans` row has ZERO `site_plan_sections` rows, so the lane's S4
   query cannot be the gate here. The mechanism that actually protects the other four
   tool pages is `pages.rebuild_policy='owned'`, which `save_page_sections_action.go`
   treats as a hard refusal ("Refusing to overwrite") and `reconcile_site_plan_action.go`
   routes to a review item rather than the generic builder. This page is created
   `rebuild_policy='owned'`.

7. **A checkbox smaller than 24x24 fails the S6 visible-area gate and is a genuine
   touch-target defect.** *Answered:* the boxes are `24px`, set in pixels rather than
   rem so a site-level root font-size cannot silently shrink them below the threshold
   the gate measures.

8. **Interaction checks share ONE browser page and accumulate state.**
   `evaluateOnPage` iterates every check against a single page per profile, so ticks
   left by one interaction are still there for the next. *Answered:* every interaction
   in the fence below begins by clicking `#vtc-reset`, which makes each claim
   self-contained. The reset control exists partly because the gate needed it, and it
   is useful to the visitor as well.

## Deliberate decisions — do not re-fix

- **The twelve items, their notes and the three tier texts are hardcoded in the
  template, not exposed as `source:"llm"` fields.** The checklist content is the
  component's identity; a vendor-trust checklist whose items an LLM can rewrite is a
  prompt, not a checklist. It also keeps the required-llm-field surface small, which is
  what the render escalation guard tests.
- **All visitor-facing copy lives in the template, including the verdict tiers, which
  the JS reads from `data-` attributes on `#vtc-verdict-box`.** The JS holds logic and
  no prose, so copy cannot drift between two files.
- **The static markup is deliberately the JS's own zero state** (0 of 12, lowest tier,
  verdict text equal to `data-low-label`). If the script never runs the panel is honest
  rather than blank. Check K asserts this equality, so the two cannot drift.
- **No em-dashes**, matching the fleet norm the brochure lane established. Check L
  enforces it.
- Invalid or partial input is not possible here (there is nothing to type), so unlike
  `smart-contrast` there is no input-validation behaviour to mistake for a dead tool.

## Deferrals — what the check vocabulary cannot express

Recorded as missing check types rather than worked around by driving the tool's own
functions, per the gate-authoring rule.

- **"There are exactly twelve checkboxes" is NOT expressible.** `selector_count` has no
  expected-value field in `criteriaCheck` and is evaluated identically to
  `selector_exists` (`if n := page.Count(sel); n > 0`), so a fence asserting twelve
  would pass with one. Asserted at S2 instead, where the harness counts them.
- **"The sector checkbox is disabled" is not expressible** — there is no attribute or
  property assertion. Covered indirectly: with N/A on, `#vtc-score-total` reads 11.
- **"The meter bar is 75% wide" is not expressible** — no style or computed-value
  assertion.

## Acceptance criteria

```criteria
{ "profiles": ["desktop", "mobile"],
  "container": ".tool-ai-vendor-trust-checklist-section",
  "checks": [
    {"id": "status", "type": "page_status_ok"},
    {"id": "boots", "type": "selector_exists",
     "selector": "[data-component=\"tool-ai-vendor-trust-checklist\"]"},
    {"id": "verdict-has-area", "type": "has_visible_area", "selector": "#vtc-verdict"},
    {"id": "first-box-is-a-real-target", "type": "has_visible_area",
     "selector": "#vtc-c1", "min_width": 20, "min_height": 20},
    {"id": "mobile-fit", "type": "no_horizontal_overflow", "profiles": ["mobile"]},
    {"id": "claim-one-tick-counts-one", "type": "interaction",
     "steps": [{"action": "click", "selector": "#vtc-reset"},
               {"action": "click", "selector": "#vtc-c1"}],
     "expect": {"selector": "#vtc-score-count", "text_matches": "^\\s*1\\s*$"}},
    {"id": "claim-nine-of-twelve-is-strong", "type": "interaction",
     "steps": [{"action": "click", "selector": "#vtc-reset"},
               {"action": "click", "selector": "#vtc-c1"},
               {"action": "click", "selector": "#vtc-c2"},
               {"action": "click", "selector": "#vtc-c3"},
               {"action": "click", "selector": "#vtc-c4"},
               {"action": "click", "selector": "#vtc-c5"},
               {"action": "click", "selector": "#vtc-c6"},
               {"action": "click", "selector": "#vtc-c7"},
               {"action": "click", "selector": "#vtc-c8"},
               {"action": "click", "selector": "#vtc-c9"}],
     "expect": {"selector": "#vtc-verdict", "text_matches": "Strong footing"}},
    {"id": "claim-five-of-twelve-is-the-middle-tier", "type": "interaction",
     "steps": [{"action": "click", "selector": "#vtc-reset"},
               {"action": "click", "selector": "#vtc-c1"},
               {"action": "click", "selector": "#vtc-c2"},
               {"action": "click", "selector": "#vtc-c3"},
               {"action": "click", "selector": "#vtc-c4"},
               {"action": "click", "selector": "#vtc-c5"}],
     "expect": {"selector": "#vtc-verdict", "text_matches": "ask about the gaps"}},
    {"id": "claim-reset-returns-to-the-lowest-tier", "type": "interaction",
     "steps": [{"action": "click", "selector": "#vtc-reset"},
               {"action": "click", "selector": "#vtc-c1"},
               {"action": "click", "selector": "#vtc-c2"},
               {"action": "click", "selector": "#vtc-c3"},
               {"action": "click", "selector": "#vtc-reset"}],
     "expect": {"selector": "#vtc-verdict", "text_matches": "Significant gaps"}},
    {"id": "claim-not-applicable-shrinks-the-denominator", "type": "interaction",
     "steps": [{"action": "click", "selector": "#vtc-reset"},
               {"action": "click", "selector": "#vtc-na-sector"}],
     "expect": {"selector": "#vtc-score-total", "text_matches": "^\\s*11\\s*$"}},
    {"id": "console", "type": "no_console_errors"}
  ]
}
```

Every interaction asserts a TERMINAL value (the score, the denominator, or the verdict
the visitor reads), never a waypoint such as "a class changed". A
`no_horizontal_overflow` failure attributed OUTSIDE the tool container is site chrome
and is not this tool's defect.

---

## ADDENDUM 2026-07-30 — what building it up the ladder actually found

Appended, not edited in: the sections above are what was believed BEFORE the tool
was placed and driven, and that ordering is the useful part.

### S6 result: 18 pass, 3 fail, 0 unexpected skips

Correlation `dc952633-4bc9-4395-a12f-4e13118f6540`, desktop + mobile, real clicks.
All five behavioural claims passed on both profiles, and `no_console_errors`
passed on both. The one skip was the deliberate profile-scoping of `mobile-fit`.

**The three failures are `bugs_open/157`, a platform bug, NOT defects in this
tool.** `has_visible_area` reports 0 for any axis whose rendered size is a whole
number, because `VisibleArea` asserts `.(float64)` on a value that
`playwright-go` returns as `int` when it is integral. `#vtc-c1` is exactly
`24px x 24px`, so it measured `0x0`; `#vtc-verdict` measured `0x94` on mobile —
only the integral axis read 0, which is the observation that identifies the bug.

Screenshots of the live URL at 1366x1500 and 390x900 show the checkbox rendering
as a normal ~24px control and the verdict box at full column width on both
profiles, and the `interaction` checks in the same run successfully CLICKED
`#vtc-c1` (Playwright enforces visibility before clicking).

**DO NOT make the checkbox size fractional to turn the gate green.** The `24px`
is deliberate (WCAG 2.2 target size, in pixels so a site root font-size cannot
shrink it), and this page was the cleanest reproducer of 157.

> **RESOLVED 2026-07-31 18:08 UTC — the platform bug is fixed and this tool needs
> NO change.** `bugs_closed/157`, live in `browser-runner-adapter` **`v1.0.1216`**.
> Re-run `bce1da22-6b47-4fef-bef7-7ef62b488ab4` against the same page:
> **21 passed / 0 failed / 1 skipped** (the skip is still `mobile-fit`'s deliberate
> profile scoping). `#vtc-c1` now measures **24x24** on desktop and mobile, and
> `#vtc-verdict` **386x47** / **143x94** — so the tool's own numbers were right all
> along and the checkbox stays at `24px`. The `improve_tool` work item this
> manufactured (`975c3be4-a310-4d7c-aece-f837478d084d`) has been **cancelled as a
> false positive**; if you find another ticket blaming this tool for a `0x0`
> measurement, it predates `v1.0.1216` and should be cancelled too, not acted on.

### S3 correction: the JS asset has a native mechanism, and it closes a defect class

The handoff that scoped this tool prescribed committing the JS through the
git-adapter by hand. That works and was done first, but it duplicates something
the platform already does: `rerender_single_page_action.collectJSAssets` reads
`content_components.js_content` and emits `tools/assets/{function}.js` as part of
the page's own commit. So `js_content` is set on this component, and the asset
path is **derived from `function`** rather than typed into the template.

That matters beyond tidiness: it is why a `<script src>` mismatch — the live
`llm-cost-calculator` defect — becomes structurally impossible here rather than
merely checked for. Both copies were verified identical by md5 (file, DB row,
served asset).

### S4 correction: two mechanisms, and they are in tension

The lane's S4 gate ("present in `site_plan_sections`, not only `page_components`")
does not apply on this site: leopardess's `site_plans` row has **zero**
`site_plan_sections` rows, and the mechanism actually protecting the other four
tool pages is `pages.rebuild_policy='owned'` — a hard refusal in
`save_page_sections_action.go`, not a resolution source.

**And it blocks the initial render.** `page-rerender`'s `save_sections` step IS
the generic section save that the ownership guard refuses, so the first attempt
failed with *"page … is rebuild_policy=owned … Refusing to overwrite."* The order
is forced: render with `generic`, then flip to `owned`. Verified as a red/green
pair — with `owned` set, the same render refuses and the served page stays
byte-identical (md5 `c44ee464c88172680248cadb6cc6c225` before and after).

### The naming contract nothing states in one place

Three values must be equal or the acceptance ladder quietly does nothing:

    doc_plans.subject_key  ==  pages.name  ==  content_components.function
                                            ==  tool-ai-vendor-trust-checklist

`load_docs` keys on `input_data.spec.function`; a mismatch yields an empty fence
and `request_browser_run` SKIPS with `needs_criteria` — honest, but it is not a
failure either, so it reads as a clean run that asserted nothing. And
`request_browser_run` resolves the URL with `name IN ($2, 'tool-' || $2)`, so a
page named `function` MINUS the prefix matches neither. This page was renamed
from `ai-vendor-trust-checklist` to `tool-ai-vendor-trust-checklist` for that
reason; the URL is unaffected because the deployed filename derives from
`pages.url`.

Measured 2026-07-30: **6 of 22 hosted tools fleet-wide** are unresolvable this
way, across finetuning.uk, fundamentallyai.com, gamesdesign.co.uk,
leopardessconsulting.co.uk and vonc.com — including this site's own
`ai-agent-roi-estimator` and `llm-cost-calculator`. They cannot be
acceptance-tested at all until renamed.

### One more deliberate decision

`disclaimer_text` is `source:"static"` with the full text as its `fallback`, not
`source:"llm"`. It is the paragraph that says the checklist is not a guarantee
("a process, not a promise"). A regeneration that softened it would change what
the tool claims, so it is deliberately outside the regenerable set. Same
reasoning as the twelve items.
