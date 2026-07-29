-- SQL_t06_smart_contrast_pilot_plan.sql — webdesign.co.uk
--
-- THE PILOT PLAN: the first criteria fence on this estate that asserts a
-- tool's ACTUAL CLAIM rather than its liveness. smart-contrast promises a
-- correct WCAG contrast ratio; these checks assert it with known-answer
-- pairs, each WATCHED PASSING on the live page before being written here
-- (migration 148's rule: never author a criterion you have not watched pass):
--   #767676 on #ffffff -> "4.54 : 1"   (the AA boundary gray, 4.5388...)
--   #000000 on #ffffff -> "21.00 : 1"  (the theoretical maximum)
-- Selectors are read off the live DOM (#fgText, #bgText, #ratioDisplay,
-- section.ported-page), not invented — the composer's twice-recorded failure
-- class (TL-016).

\set ON_ERROR_STOP on
BEGIN;

INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by, is_current)
VALUES ('tool', 'smart-contrast', $plan$
# PLAN — smart-contrast (webdesign.co.uk)

## What it is

Enter a foreground and a background colour; get the WCAG contrast ratio, pass/fail badges for AA and AAA, and suggested nearest-passing fixes.

## How it is delivered

Entirely client-side; a ported single-page tool (section.ported-page). Recomputes on every input event.

## Deliberate decisions — do not re-fix

- Invalid hex is deliberately IGNORED (the display keeps the last valid ratio). An earlier probe scored this as "dead"; it is input validation working (NOTES, first census).
- The AA badge is EXPECTED to read score-fail for many inputs — that is the tool doing its job, not a broken state.

## Acceptance criteria

```criteria
{ "profiles": ["desktop", "mobile"],
  "container": "section.ported-page",
  "checks": [
    {"id": "boots", "type": "selector_exists", "selector": "section.ported-page"},
    {"id": "console", "type": "no_console_errors"},
    {"id": "status", "type": "page_status_ok"},
    {"id": "mobile-fit", "type": "no_horizontal_overflow", "profiles": ["mobile"]},
    {"id": "claim-aa-boundary", "type": "interaction",
     "steps": [{"action": "fill", "selector": "#fgText", "value": "#767676"},
               {"action": "fill", "selector": "#bgText", "value": "#ffffff"}],
     "expect": {"selector": "#ratioDisplay", "text_matches": "4\\.54"}},
    {"id": "claim-maximum", "type": "interaction",
     "steps": [{"action": "fill", "selector": "#fgText", "value": "#000000"},
               {"action": "fill", "selector": "#bgText", "value": "#ffffff"}],
     "expect": {"selector": "#ratioDisplay", "text_matches": "21\\.00"}}
  ]
}
```
$plan$, 'webdesign_tools_repair', 'webdesign_couk_thread', true)
ON CONFLICT (subject_type, subject_key) WHERE is_current DO NOTHING;

DO $verify$
DECLARE v boolean;
BEGIN
    SELECT body LIKE '%```criteria%' AND body LIKE '%claim-aa-boundary%' INTO v
      FROM doc_plans WHERE subject_type='tool' AND subject_key='smart-contrast' AND is_current;
    IF NOT COALESCE(v,false) THEN RAISE EXCEPTION 'pilot PLAN not written with claim checks'; END IF;
    RAISE NOTICE 'smart-contrast pilot PLAN written: 4 standard + 2 claim checks';
END $verify$;

COMMIT;
