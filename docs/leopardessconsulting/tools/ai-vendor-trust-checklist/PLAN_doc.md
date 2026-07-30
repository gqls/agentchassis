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
