-- Travelling-doc PLAN for the first repaired webdesign.co.uk tool.
--
-- This is the bridging step the workstream PLAN §2 describes: these 63 tools
-- were ported as HTML, never born through the generator, so none has a
-- doc_plans row and the acceptance ladder has never been able to see one.
-- Writing the PLAN is what makes the repair durable — without it the tool is
-- invisible to Tier 2/3/4 again the moment we look away.

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO doc_plans (subject_type, subject_key, body, source, source_agent, created_by, is_current)
VALUES ('tool', 'insight-injector', $plan$
# PLAN — Insight Injector (webdesign.co.uk)

## What it is for
Someone about to point an AI site-builder at their business gets generic copy,
because they gave it a generic brief. This tool turns three concrete things —
the business name, one checkable fact, one real customer story — plus a
banned-word list into a prompt that constrains the model. The visible lesson is
that specifics are what generic copy cannot fake.

## What it produces
A single prompt, ready to paste into v0, Lovable or Bolt. It embeds the three
inputs verbatim, names a tone, and forbids fifteen filler words
(delve, testament, tapestry, bustling, innovative, synergy, landscape, embark,
seamless, elevate, unlock, tailored, comprehensive, cutting-edge, premier).
Nothing is sent anywhere: the whole tool runs in the page.

## Delivery mechanism
Static page with inline JS. `generatePrompt()` reads `#biz-name`, `#biz-fact`,
`#biz-story`, `#biz-tone` and writes the prompt into `#output`. `window.onload`
renders the banned words into `#banned-tags` as `.banned-tag` spans.
`copyPrompt()` copies via the clipboard API and refuses to copy the placeholder.

## Deliberate decisions — do not "fix" these
- **Empty inputs are allowed.** Blank fields become bracketed placeholders
  (`[Business Name]`, `[Insert Hard Fact Here]`) rather than blocking
  generation, so the shape of the prompt can be seen before the user has their
  material to hand. Do not add required-field validation.
- **The banned list is hard-coded and not user-editable.** It is the tool's
  editorial opinion; making it configurable would turn a stance into a
  text box.
- **Two-column layout is load-bearing** — the output copy says "fill out the
  insights on the left". If the layout is ever made single-column, that
  sentence must change with it.

## Repair history
Restored 2026-07-29. The port had dropped the entire left-hand panel: the CSS
for it (`.tool-layout`, `.controls-panel`, `.input-group`, `.banned-words-box`,
`.banned-tag`) and all of the JS survived, but the markup did not, so the page
threw `TypeError: Cannot set properties of null (setting 'innerHTML')` on load
and offered a visitor exactly one control (Copy). The panel was rebuilt to the
contract the surviving CSS and JS already described — no redesign.

```criteria
{ "profiles": ["desktop","mobile"],
  "container": ".tool-layout",
  "checks": [
    {"id":"boots","type":"selector_exists","selector":".tool-layout"},
    {"id":"console","type":"no_console_errors"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"inputs-present","type":"selector_exists","selector":"#biz-name"},
    {"id":"banned-words-render","type":"selector_exists","selector":".banned-tag"},
    {"id":"generates-with-inputs","type":"interaction",
     "steps":[{"action":"fill","selector":"#biz-name","value":"Pemberton Joinery"},
              {"action":"fill","selector":"#biz-fact","value":"4,000 staircases since 1988"},
              {"action":"click","selector":".btn"}],
     "expect":{"selector":"#output","text_matches":"Pemberton Joinery"}},
    {"id":"embeds-the-ban-list","type":"interaction",
     "steps":[{"action":"click","selector":".btn"}],
     "expect":{"selector":"#output","text_matches":"delve"}}
  ]
}
```
$plan$, 'webdesign_tools_repair', 'webdesign_couk_thread', 'webdesign_couk_thread', true)
ON CONFLICT (subject_type, subject_key) WHERE is_current DO NOTHING;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('tool', 'insight-injector',
'Repaired 2026-07-29. Browser census scored it BROKEN: on load it threw
"TypeError: Cannot set properties of null (setting ''innerHTML'')" from
window.onload writing into #banned-tags, and a visitor had one control (Copy)
on the whole page. Cause: the port dropped the left-hand input panel — its CSS
and its JS both survived intact, only the markup was missing, so the tool was
addressing five elements (#biz-name, #biz-fact, #biz-story, #biz-tone,
#banned-tags) that no longer existed. Fix: rebuilt the panel to the contract the
surviving CSS and JS already described, adding nothing new. Verified in headless
chromium: 15 banned-word tags render, and a generated prompt carries the
business name, the hard fact, the customer story and the ban list
(tagsRendered=15, hasName/hasFact/hasStory/hasBanned all true).',
 '["repair","acceptance-fail-fixed","port-gap"]'::jsonb,
 'webdesign_tools_repair', 'webdesign_couk_thread');

DO $verify$
DECLARE v_plan int; v_note int; v_crit boolean;
BEGIN
    SELECT count(*) INTO v_plan FROM doc_plans
     WHERE subject_type='tool' AND subject_key='insight-injector' AND is_current;
    IF v_plan <> 1 THEN RAISE EXCEPTION 'expected 1 current PLAN, found %', v_plan; END IF;

    SELECT body LIKE '%```criteria%' INTO v_crit FROM doc_plans
     WHERE subject_type='tool' AND subject_key='insight-injector' AND is_current;
    IF NOT v_crit THEN RAISE EXCEPTION 'PLAN carries no criteria fence — the ladder cannot test it'; END IF;

    SELECT count(*) INTO v_note FROM doc_notes
     WHERE subject_type='tool' AND subject_key='insight-injector';
    IF v_note < 1 THEN RAISE EXCEPTION 'no NOTES entry recorded'; END IF;

    RAISE NOTICE 'insight-injector: PLAN with criteria fence + repair note recorded';
END $verify$;

COMMIT;
