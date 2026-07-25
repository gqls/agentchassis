-- ============================================================================
-- PREPARED — Recovery waterfall & priming simulator for oufe.com
-- NOT YET APPLIED. Written 2026-07-25.
--
-- WAIT FOR THE SITE PLAN BEFORE RUNNING THIS.
--   bugs_open/001 and /050: re-planning clobbers pages, and the guard keys on
--   build_status='deployed'. A tool page inserted BEFORE build-site-planner has
--   run can be reconciled away. Run this only once the plan exists and the P1
--   pages are built:
--     SELECT name, page_type, build_status FROM pages
--      WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com');
--
-- HOW TO LOAD THE HTML
--   The component body lives in tool/tool-recovery-waterfall.html, NOT inline
--   here — a 15KB template pasted into SQL is unreviewable and every edit would
--   have to be made twice. Load it with psql's \set, which does no escaping and
--   handles the quotes and backslashes in the JS correctly:
--
--     \set tmpl `cat docs/agent_docs/docs024_key_docs_latest/oufe/tool/tool-recovery-waterfall.html`
--
--   then use :'tmpl' where marked. Run from the repo root.
--
-- THREE CONSTRAINTS THIS FILE HONOURS
--   1. JS is INLINE in html_template and js_content stays NULL. The js_content
--      lane publishes /tools/assets/<function>.js and then injects no <script>
--      tag — published-but-inert, the bugs_open/041 class. Live tool components
--      all follow the inline convention; do not "tidy" the JS out.
--   2. IDENTITY MUST AGREE across three places or acceptance resolves the wrong
--      URL and silently tests nothing (landmine G6):
--        pages.name = content_components.function = doc_plans.subject_key
--        = 'tool-recovery-waterfall'
--   3. Tool pages are rebuild_policy='owned' so the generic page builder never
--      rewrites them (save_page_sections_action.go:139-157 hard-refuses).
-- ============================================================================

\set ON_ERROR_STOP on
-- \set tmpl `cat docs/agent_docs/docs024_key_docs_latest/oufe/tool/tool-recovery-waterfall.html`

BEGIN;

-- 1. the component -----------------------------------------------------------
INSERT INTO content_components (
  name, function, display_name, description, category,
  component_level, render_mode, html_template, js_content,
  input_schema, is_active, suitable_page_types
)
VALUES (
  'tool-recovery-waterfall-oufe-com',
  'tool-recovery-waterfall',
  'Recovery waterfall and priming simulator',
  'Client-side model of a strict priority waterfall. The reader sets an enterprise value, a super-senior new-money tranche, three classes of claim and a senior write-down, and sees recovery per class in money and per cent, which class the value breaks in, and exactly how much each existing class gives up to the new money. Deterministic arithmetic, no back end, nothing transmitted.',
  'finance',
  'tool',
  'template',
  :'tmpl',
  NULL,            -- deliberate: see constraint 1 in the header
  NULL,
  true,
  '["tool"]'::jsonb
);

-- 2. the page ----------------------------------------------------------------
INSERT INTO pages (
  site_id, name, url, title, page_type, status, build_status,
  nav_label, nav_order, in_header, in_footer, rebuild_policy, meta_description
)
SELECT
  s.id,
  'tool-recovery-waterfall',
  '/tools/tool-recovery-waterfall.html',
  'Recovery waterfall and priming simulator | OUFE',
  'tool',
  'active',
  'pending',
  'Recovery waterfall',
  40,
  false,   -- keep the P1 header short; the tool is reached from the dossier
  true,
  'owned',
  'An interactive model of how value is distributed in a restructuring: set an enterprise value and a capital structure, and see which class the value breaks in and what rescue finance costs the classes beneath it.'
FROM sites s WHERE s.domain = 'oufe.com';

-- 3. bind component to page --------------------------------------------------
INSERT INTO page_components (page_id, component_id, position, slot_name, build_status)
SELECT p.id, c.id, 1, 'main', 'pending'
FROM pages p
JOIN sites s ON s.id = p.site_id
CROSS JOIN content_components c
WHERE s.domain = 'oufe.com'
  AND p.name = 'tool-recovery-waterfall'
  AND c.function = 'tool-recovery-waterfall'
  AND c.name = 'tool-recovery-waterfall-oufe-com';
-- NOTE: slot_name must NOT be NULL. A null slot renders nothing while the job
-- still reports COMPLETED (idea.uk delivery trap).

-- 4. travelling PLAN with acceptance criteria --------------------------------
UPDATE doc_plans SET is_current = false, superseded_at = now()
WHERE subject_type = 'tool' AND subject_key = 'tool-recovery-waterfall' AND is_current;

INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by, is_current, pinned)
VALUES (
  'tool', 'tool-recovery-waterfall',
  $plan$# PLAN — Recovery waterfall and priming simulator (oufe.com)

## What it is for
A reader who has never modelled a capital structure should be able to see, in
under a minute, the single most important fact about a restructuring: value is
distributed in strict order of priority, so there is a point in the stack where
it runs out, and everything below that point recovers nothing. Moving one slider
should make that visible.

## What it computes
A strict priority waterfall. For each class in order, recovery = min(claim,
value remaining), and the remainder passes down. Inputs: enterprise value, a
super-senior new-money tranche (with an include/exclude toggle), senior, junior
and subordinated claims, and a percentage write-down applied to the senior claim.

Two derived readouts carry the teaching:
- **Where the value breaks** — the first class that is only partly paid, or if
  none is partly paid, the first that recovers nothing.
- **What the new money costs** — the model is run a second time with the
  super-senior tranche removed, and the difference is reported per class. This
  is the point of the tool: priming is not an abstraction, it is a measurable
  transfer out of the existing stack.

## Honesty contract
Every number displayed is computed from the reader's own inputs by the
arithmetic above. Nothing is fetched, nothing is assumed about any real company,
and there is no LLM in the render path. Defaults are round illustrative figures
and are labelled as such in the interface. No real capital structure is
pre-loaded: the site's evidence register currently holds zero verified facts,
and a tool that shipped a named company's figures as defaults would be asserting
them (see docs024/oufe/PLAN §C2, and the 043 spec-poisoning precedent).

The tool carries its own proximate disclaimer, in the tool rather than in the
site footer: it is educational scenario modelling, not a valuation, not a
forecast and not advice, and it explicitly names what it does NOT model
(security, guarantees, structural subordination, intercompany claims,
contingent liabilities).

## Privacy
Everything runs in the reader's browser. No input is transmitted or persisted.
That is what makes the "bring your own numbers" idea honest at this stage: it is
private because there is no server, not because we promise not to look.

## Delivery
Inline — all CSS and JS ship inside html_template, render_mode='template',
js_content NULL. The js_content lane publishes the file and injects no script
tag (bugs_open/041 class), so extracting the JS would silently kill the tool.

## Acceptance criteria

```criteria
{ "profiles": ["desktop","mobile"],
  "container": ".tool-container",
  "checks": [
    {"id":"boots","type":"selector_exists","selector":".tool-container"},
    {"id":"console","type":"no_console_errors"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"high-ev-covers-all","type":"interaction",
     "steps":[{"action":"fill","selector":"#rw-ev","value":"15000"}],
     "expect":{"selector":"#rw-break-class","text_matches":"All classes covered"}},
    {"id":"zero-ev-breaks-at-top","type":"interaction",
     "steps":[{"action":"fill","selector":"#rw-ev","value":"0"}],
     "expect":{"selector":"#rw-break-class","text_matches":"New money"}}
  ]
}
```

The two interaction checks assert the arithmetic actually runs, at opposite ends
of the range: with enough value every claim is covered, and with none the break
is at the very top of the stack. A tool whose JS failed to execute would keep
its initial verdict text and fail both. Deliberately NO `asset_loads` check —
that criterion asserts a JS-extraction path that was designed and never built,
and failed every tool on its first sweep (TL-016).

Selectors are copied from the shipped template, not invented.
$plan$,
  'manual', 'oufe-workstream-2026-07-25', true, true
);

COMMIT;

-- ---------------------------------------------------------------------------
-- AFTER APPLYING
--   1. Render and deploy the page (do NOT re-run build-site-planner just to add
--      a page — re-planning is how built pages get clobbered, bugs_open/001/050).
--   2. Verify the JS actually reached the live page — the whole point of the
--      inline convention:
--        curl -s https://oufe.com/tools/tool-recovery-waterfall.html \
--          | grep -c "rw-break-class"
--   3. Let the tool_acceptance_due sweep raise an acceptance_run item, or
--      dispatch tool-acceptance-agent directly. Read the verdict in doc_notes:
--        SELECT body FROM doc_notes WHERE subject_key='tool-recovery-waterfall'
--         ORDER BY created_at DESC LIMIT 1;
--      A Tier-2 static pass is NOT evidence the tool works — only the headless
--      run is.
-- ---------------------------------------------------------------------------
