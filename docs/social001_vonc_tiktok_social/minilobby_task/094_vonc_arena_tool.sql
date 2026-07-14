-- 094_vonc_arena_tool.sql — vonc.com: make the Arena a real destination
-- Created 2026-07-14. Run AFTER 091-093. Owner decision 2026-07-14: build the
-- Arena as a real page/tool (it existed only as copy — CTAs sold "the Arena"
-- with no page anywhere; a previous session parked the retarget "until the
-- real arena exists").
--
-- Platform constraint: no backend / user accounts. Arena v1 is a CLIENT-SIDE
-- tool page like the Gauntlet — the only pattern the platform supports.
-- Community features (rooms, duels, live votes) stay v3. v1 spec drawn from
-- 002e_concept_spark(6).md §Arena Mechanics.
--
-- What this migration does:
--   1. site_plan_pages row  tool-arena (role tool, in_header) on the current
--      vonc plan — the page is now PROMISED, so incomplete_page_group flags
--      it (needs_human_review) until it deploys: that finding is expected and
--      is the live proof the new check works.
--   2. add_tool work item -> tool-generator (the Gauntlet-class zero-Go path:
--      generate_tool_html LLM step -> create_tool_component, which inserts
--      the pages row itself and enforces the tool-doc sentinel header).
--   3. doc_plans PLAN row for the tool (130-pilot format) so acceptance
--      checkers have a contract.
--
-- Known gaps to work around (runbook):
--   TP-002 — tool-generator never enqueues the final render/deploy: dispatch
--            it manually after save_tool completes.
--   TP-004 — reconcile_site_plan's tool route is commented out: do NOT let a
--            needs_page item route tool-arena to page-build-handler.
--   TL-001 — a generic full page rebuild clobbers the widget row: future
--            edits via the section-editor targeted path ONLY.
--   After any vonc reconcile_site_plan, park the re-emitted
--   needs_page:provocation item back to 'detected'.
--
-- Reversal: DELETE the three inserts by the literal keys below (no existing
-- rows are modified, so no backup tables).

BEGIN;

-- ── 1. Plan page (current vonc plan; url matches the tool page convention) ──
INSERT INTO site_plan_pages (plan_id, name, role, slug, url, parent_section,
                             in_header, in_footer, nav_order, title, nav_label, meta_description)
VALUES
 ('77493277-f510-47ea-aa27-8fca415743d6', 'tool-arena', 'tool', 'arena',
  '/tools/arena/index.html', 'tools',
  true, false, 45, 'The Arena', 'Arena',
  'The competitive mode: today''s provocation, file your take, react to the takes on the floor — Genius, Delusional, Suspicious, Based or Cursed — and watch remix chains form.');

-- ── 2. Tool generation work item (novel path -> tool-generator) ─────────────
-- Spec fields are the ones tool-generator's generate_tool_html prompt reads:
-- name / function / description / complexity.
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'human', 'build', 'add_tool',
    'medium', 'Build the Arena — competitive provocation/take/reaction tool (v1, client-side)',
    '{
      "name": "The Arena",
      "function": "tool-arena-interface",
      "complexity": "medium",
      "description": "The Arena is Spark''s competitive mode, v1 as a fully self-contained client-side experience (no fetch calls, no backend). Four elements, in order: (1) TODAY''S PROVOCATION — a bold prompt displayed prominently at the top (embed 5 sample provocations in JS and pick one by day-of-date so the page changes daily, e.g. \"AI art is just expensive plagiarism — defend or destroy\"). (2) FILE YOUR TAKE — a textarea + submit button; store the visitor''s take in localStorage keyed by the provocation, show it back with a ''Your take is on the floor'' state, allow re-filing. (3) THE FLOOR — 4-6 embedded sample takes for today''s provocation, each with the five Arena Reactions as tappable chips: Genius / Delusional / Suspicious / Based / Cursed. Tapping increments a visible count (persist counts in localStorage); one reaction per take, switchable. (4) REMIX CHAIN — a simple visual (nested/indented cards or a connected line) showing one take remixed twice, illustrating chain reactions with shared credit. Tone: combative, analytical — ''sharpest response wins''. Include a short footer note that duels and live rooms are coming, linking the words ''the Gauntlet'' to /tools/gauntlet/index.html for the full daily game."
    }'::jsonb,
    100, 'tool-generator', 'triaged', 'human', 'add_tool_novel:tool-arena-interface'
);

-- ── 3. Tool PLAN doc (130-pilot format; -EDIT checks are skipped by checkers) ──
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by, notes)
VALUES ('tool', 'tool-arena-interface', $plan$# PLAN — tool-arena-interface

## Aim
Make "the Arena" — sold by site copy since launch — a real destination: the
competitive provocation/take/reaction loop as a v1 client-side page, funnelling
into the Daily Gauntlet. Source: 002e_concept_spark(6).md §Arena Mechanics.

## Behaviour contract
Daily provocation (5 embedded samples, picked by date) at top; take-filing via
textarea persisted to localStorage (re-filing allowed); 4-6 sample takes each
carrying the five Arena Reactions (Genius/Delusional/Suspicious/Based/Cursed)
as chips with localStorage-persisted counts, one switchable reaction per take;
a static remix-chain visual (one take remixed twice). All client-side, no
fetch. Footer links "the Gauntlet" -> /tools/gauntlet/index.html.

## Acceptance criteria
```criteria
{ "profiles": ["desktop", "mobile"],
  "note": "reaction-flow-EDIT uses PLACEHOLDER selectors — set them from the generated component html before Tier-3/4 consume this block; checkers skip checks whose id ends in -EDIT",
  "checks": [
    {"id": "boots",      "type": "selector_exists",        "selector": ".tool-container"},
    {"id": "console",    "type": "no_console_errors"},
    {"id": "status",     "type": "page_status_ok"},
    {"id": "mobile-fit", "type": "no_horizontal_overflow", "profiles": ["mobile"]},
    {"id": "reaction-flow-EDIT", "type": "interaction",
      "steps": [{"action": "click", "selector": "EDIT-first-reaction-chip"}],
      "expect": {"selector": "EDIT-reaction-count", "text_matches": "[1-9]"}}
  ] }
```

## Deliberate decisions — do not re-fix
- Sample provocations/takes are EMBEDDED, not fetched from provocations.json —
  generator invariant (no fetch); wiring to the live provocations feed (which
  already carries an `arena` key) is a v2 task, not a defect.
- No usernames anywhere (Radical Anonymity is an Arena founding rule).
- Duels / rooms / live voting deliberately absent — v3 community features.
- TL-001: this page must NEVER receive a generic full page rebuild (it would
  clobber the widget row); edits go through the section-editor targeted path.
$plan$, 'human', '094_vonc_arena_tool', 'Arena v1 — owner decision 2026-07-14');

-- ── Verify ──────────────────────────────────────────────────────────────────
DO $$
DECLARE planrow INT; item INT; plandoc INT;
BEGIN
  SELECT COUNT(*) INTO planrow FROM site_plan_pages
  WHERE plan_id = '77493277-f510-47ea-aa27-8fca415743d6' AND name = 'tool-arena';
  SELECT COUNT(*) INTO item FROM site_work_items
  WHERE item_key = 'add_tool_novel:tool-arena-interface'
    AND status IN ('triaged', 'claimed', 'complete');
  SELECT COUNT(*) INTO plandoc FROM doc_plans
  WHERE subject_type = 'tool' AND subject_key = 'tool-arena-interface' AND is_current;
  IF planrow <> 1 OR item < 1 OR plandoc <> 1 THEN
    RAISE EXCEPTION 'verify failed: planrow=% item=% plandoc=% (want 1/1+/1)', planrow, item, plandoc;
  END IF;
  RAISE NOTICE 'verified: tool-arena planned, add_tool item queued, PLAN doc current';
END $$;

COMMIT;
