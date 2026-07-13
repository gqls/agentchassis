-- pilot_PLAN_tool-archetype-taster-quiz.sql — Task 2A pilot seed. 2026-07-07.
-- Hand-seeds the FIRST tool PLAN into doc_plans (source='human'). Later
-- write_doc_plan calls supersede it cleanly. Concrete facts come from the
-- tool-generator's own invariants + 037; unknowns are marked EDIT: — fill or
-- leave (an EDIT marker in prose is honest; the criteria placeholders are
-- flagged in the block's "note" field so checkers skip them).
--
-- Known: function pairs with archetype-result-card (037); a condensed
-- 3-question taster of the Daily Gauntlet; currently is_active = f on BOTH
-- instance rows (duplicates = per-instance rows; docs key by function).

INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by, notes)
VALUES ('tool', 'tool-archetype-taster-quiz', $plan$# PLAN — tool-archetype-taster-quiz

## Aim
A condensed 3-question taster of the Daily Gauntlet experience: a visitor
answers three questions and receives an archetype result, funnelling interest
toward the full Gauntlet. EDIT: extend with the product framing you want on
record (the candidates list truncates the original description at 60 chars).

## Source spec
EDIT: paste the site-spec / roadmap slice this derives from (vonc/Spark —
the Gauntlet feature family; see 037's candidate list).

## Behaviour contract
Three questions presented in sequence; selecting an answer advances; after the
third answer an archetype result renders (delivered with/alongside
`archetype-result-card`). All logic client-side; no external calls (generator
invariant). EDIT: states/inputs/outputs in detail once the component html is
in front of you (question data source, result mapping, reset behaviour).

## Acceptance criteria
```criteria
{ "profiles": ["desktop", "mobile"],
  "note": "quiz-flow-EDIT uses PLACEHOLDER selectors — set them from the component html before Tier-3/4 consume this block; checkers skip checks whose id ends in -EDIT",
  "checks": [
    {"id": "boots",      "type": "selector_exists",        "selector": ".tool-container"},
    {"id": "console",    "type": "no_console_errors"},
    {"id": "asset",      "type": "asset_loads",            "path": "/tools/assets/tool-archetype-taster-quiz.js"},
    {"id": "status",     "type": "page_status_ok"},
    {"id": "mobile-fit", "type": "no_horizontal_overflow", "profiles": ["mobile"]},
    {"id": "quiz-flow-EDIT", "type": "interaction",
      "steps": [{"action": "click", "selector": "EDIT-q1-answer"},
                 {"action": "click", "selector": "EDIT-q2-answer"},
                 {"action": "click", "selector": "EDIT-q3-answer"}],
      "expect": {"selector": "EDIT-result-element", "text_matches": "\\S"}}
  ] }
```

## Delivery mechanism
Path 1 (component inline `<script>` → `/tools/assets/tool-archetype-taster-quiz.js`,
extracted on rerender) — per the generator's structure rules. EDIT/VERIFY: this
tool predates parts of the current pipeline; confirm the asset exists in the
deployed repo before Tier-2 relies on the `asset` check.

## Dependencies
`archetype-result-card` (result rendering pairing — 037); the Gauntlet feature
family on vonc/Spark. Status note: both instance rows are currently
`is_active = false` — reactivation is a product decision, not a fix.

## Deliberate decisions — do not re-fix
- Exactly THREE questions — the taster deliberately under-delivers relative to
  the Daily Gauntlet; do not "improve" it by adding questions.
- EDIT: add any further intentional choices (e.g. no persistence of answers;
  result phrasing tone) so later passes don't fight them.
$plan$, 'human', 'pilot', 'pilot seed — road-tests the PLAN format (Task 2A, 2026-07-07)');

-- Verify (expect: one current row, fence intact):
SELECT subject_key, is_current, body LIKE '%```criteria%' AS has_fence,
       length(body) AS body_len, created_at
FROM doc_plans
WHERE subject_type = 'tool' AND subject_key = 'tool-archetype-taster-quiz';
