-- 0NN_supersede_xp_curve_plan_selectors.sql — correct the invented selectors.
-- DRAFT 2026-07-09. Renumber 0NN. DB-only. No snapshot_agent (data tables).
--
-- WHY: the first machine-written PLAN (source='tool-generator', 2026-07-09
-- 13:07:50) named two ids that do NOT exist in the component HTML:
--   #xpTableBody  -> no such id; the real element is a bare <tbody> inside #tableWrap
--   #statsStrip   -> the real id is #statRow
-- Verified ids in html_template: baseXP, curveHint, curveType, formulaBox,
-- growthFactor, growthHint, growthLabel, maxLevel, statRow, tableWrap, xpChart.
-- The composer obeyed "never invent a selector" for the control it ACTS on
-- (#curveType, real) and broke it for the thing it ASSERTS on. The durable
-- remedy is Tier-2 static validation of criteria selectors against
-- html_template at write time — not a sterner prompt. This migration repairs
-- the one live PLAN and records the correction as a NOTES entry.
--
-- Supersede pattern (mirrors write_doc_plan + site_specs): flip the current
-- row (is_current=false, superseded_at=now()), insert the new body as current.
-- idx_doc_plans_current (partial unique on (subject_type,subject_key) WHERE
-- is_current) enforces exactly one current row, so the order matters.

-- Clear any poisoned session state before starting. If the previous attempt
-- raised inside a DO block, psql sits in an ABORTED transaction (prompt shows
-- `clients_db=!#`) and silently ignores EVERY subsequent command — including
-- BEGIN. On a clean session this line just warns "no transaction in progress".
ROLLBACK;

-- Prefer:  psql "$CLIENTS_DB_URL" -f drafts/0NN_supersede_xp_curve_plan_selectors.sql
-- (or \i <path> inside psql). Pasting long files into the prompt mangles
-- comment lines and dollar-quoted bodies.

BEGIN;

-- Guard 1: the assumption behind the new interaction check — a <table> lives
-- inside #tableWrap. If this fails, do NOT proceed: rename the check id to
-- `curve-switch-EDIT` (checkers skip ids ending -EDIT) and re-run.
--
-- NB: written with strpos/substr, NOT a regex. Postgres ARE caps bounded
-- repetition at 255 (RE_DUP_MAX), so `.{0,300}` raises
-- "invalid regular expression: invalid repetition count(s)".
DO $$
DECLARE h text; p int;
BEGIN
    SELECT html_template INTO h
    FROM content_components
    WHERE function = 'tool-xp-curve-designer'
      AND component_level = 'tool'
      AND is_active
    LIMIT 1;

    IF h IS NULL THEN
        RAISE EXCEPTION 'active tool component tool-xp-curve-designer not found';
    END IF;

    p := strpos(h, 'id="tableWrap"');
    IF p = 0 THEN
        RAISE EXCEPTION 'id="tableWrap" not present in html_template';
    END IF;

    IF strpos(substr(h, p, 400), '<table') = 0 THEN
        RAISE EXCEPTION 'no <table> within 400 chars after id="tableWrap" — do not assert "#tableWrap tr"; rename the check to curve-switch-EDIT and re-run';
    END IF;
END $$;

-- Guard 2: the ids we are switching TO must exist; the ones we remove must not.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM content_components
    WHERE function = 'tool-xp-curve-designer' AND is_active
      AND html_template LIKE '%id="statRow"%'
      AND html_template LIKE '%id="tableWrap"%'
      AND html_template NOT LIKE '%xpTableBody%'
      AND html_template NOT LIKE '%statsStrip%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'selector premises not met (found %) — re-check the ids before superseding', n;
    END IF;
END $$;

-- 1) Retire the current PLAN.
UPDATE doc_plans
SET is_current = false,
    superseded_at = now(),
    updated_at = now()
WHERE subject_type = 'tool'
  AND subject_key = 'tool-xp-curve-designer'
  AND is_current;

-- 2) Insert the corrected PLAN as current.
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by, notes)
VALUES ('tool', 'tool-xp-curve-designer', $plan$# PLAN — tool-xp-curve-designer

## Aim
Model and compare level-progression XP curves (linear, quadratic, exponential):
tune base XP and growth, and read per-level costs in a table beside a
cumulative-XP canvas chart. Entirely client-side.

## Source spec
Function `tool-xp-curve-designer` on gamesdesign.co.uk; inputs: curve type, base
XP, growth factor, max level; outputs: formula display, stats row, per-level
table with bar, cumulative canvas chart.

## Behaviour contract
- **Inputs:** `#curveType` (linear/quadratic/exponential), `#baseXP`
  (10–100000), `#growthFactor` (clamped per curve type; label and hint in
  `#growthLabel` / `#growthHint`), `#maxLevel` (5–100).
- **Formulas:** linear `base + growth×(n−1)`; quadratic `base × n^growth`;
  exponential `base × growth^(n−1)`. Values rounded to integers.
- **Growth limits:** linear 0–100000 step 10; quadratic 1–10 step 0.1;
  exponential 1–3 step 0.01. Defaults switch on curve change (`#curveHint`).
- **On every `input`/`change`:** `#formulaBox` shows the formula; `#statRow` the
  summary stats; the table inside `#tableWrap` renders one row per level
  (transition, per-level XP, cumulative XP, proportional bar); `#xpChart`
  redraws on 2D canvas (area, line, dots ≤25 levels, grid).
- **Resize:** canvas redraws via a debounced `resize` listener (100 ms).
- **No external libraries, no network calls.**

## Acceptance criteria
```criteria
{ "profiles": ["desktop","mobile"],
  "checks": [
    {"id":"boots","type":"selector_exists","selector":".tool-container"},
    {"id":"console","type":"no_console_errors"},
    {"id":"asset","type":"asset_loads","path":"/tools/assets/tool-xp-curve-designer.js"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"curve-switch","type":"interaction",
      "steps":[{"action":"select","selector":"#curveType","value":"exponential"}],
      "expect":{"selector":"#tableWrap tr"}}
  ] }
```

## Delivery mechanism
Path 1 — component inline script, extracted to
/tools/assets/tool-xp-curve-designer.js on rerender.

## Dependencies
None.

## Deliberate decisions — do not re-fix
- The growth-factor input swaps min/max/step/default per curve type: one reused
  field, not three — v1 kept simple by design.
- Raw 2D canvas, not a charting library — no external dependencies.
- Chart dots suppressed above 25 levels: a deliberate v1 threshold.
- XP rounds to integers; fractional XP is out of scope for v1 — the improvement
  loop iterates it against the criteria above.

## Correction log
- v1 (machine-written, 2026-07-09) asserted a table-body id and described a
  stats-strip id that the component never defines. Real: the table lives inside
  `#tableWrap` (its `<tbody>` carries no id) and the stats element is `#statRow`.
- Validating criteria selectors against the component HTML at write time is
  Tier-2's job, not the composer's good intentions.
$plan$, 'migration', 'supersede-fix-selectors',
'v1 named non-existent ids (#xpTableBody, #statsStrip); corrected to #tableWrap / #statRow. Composer invented the assert-side selector; Tier-2 static validation is the remedy.');

-- 3) Record the correction in the tool's NOTES stream (this is what NOTES are for).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('tool', 'tool-xp-curve-designer', $note$## Corrected invented selectors in the first auto-written PLAN
Observed: the machine-written PLAN asserted `#xpTableBody tr` and described `#statsStrip`; neither id exists in the component HTML.
Root cause: the composer copied a real selector for the control it acts on (`#curveType`) but invented ids for the elements it asserts on. The prompt already forbids invention, so prompting is not the fix.
Fix: PLAN superseded — the interaction check now expects `#tableWrap tr`, and the behaviour contract names `#statRow`.
Verified: ids listed from `html_template` (baseXP, curveHint, curveType, formulaBox, growthFactor, growthHint, growthLabel, maxLevel, statRow, tableWrap, xpChart); migration guards assert `#tableWrap` contains a `<table>` and that the removed ids are absent.
Categories: criteria, docs, acceptance-fail
$note$, '["criteria","docs"]'::jsonb, 'migration', 'supersede-fix-selectors');

-- Guard 3: exactly one current PLAN, and it carries the corrected selectors.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM doc_plans
    WHERE subject_type = 'tool' AND subject_key = 'tool-xp-curve-designer' AND is_current
      AND body LIKE '%#tableWrap tr%'
      AND body NOT LIKE '%xpTableBody%'
      AND body LIKE '%```criteria%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'corrected PLAN not current (found %)', n;
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT is_current, superseded_at IS NOT NULL AS retired, length(body) AS len,
--          body LIKE '%```criteria%' AS has_fence, source, created_at
--   FROM doc_plans WHERE subject_type='tool' AND subject_key='tool-xp-curve-designer'
--   ORDER BY created_at;                       -- expect: v1 retired, v2 current
--
--   SELECT categories, left(body,80) AS head, created_at
--   FROM doc_notes WHERE subject_key='tool-xp-curve-designer' ORDER BY created_at DESC LIMIT 3;
--
-- Note: the tool's NOTES stream now opens with a docs correction — written by
-- hand. The first MACHINE-written note still awaits a successful fix or
-- recreation (Task-4 proof).
