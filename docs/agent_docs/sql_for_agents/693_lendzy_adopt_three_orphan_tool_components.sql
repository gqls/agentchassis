-- 693_lendzy_adopt_three_orphan_tool_components.sql
--
-- lendzy.co.uk: three tool pages serve HTTP 200 and are recorded as never built.
-- Each carries a SINGLE page_components row, written 2026-08-02 by the original
-- shadow build, with component_id NULL and slot_name 'section'. resolveComponent
-- (rerender_page_sections_action.go:361) resolves a section by component_id then
-- by slot_name; the id is empty and no component is named, functioned or typed
-- 'section', so the slot enters UnresolvedSlots, rerenderResolution.fatal()
-- (:650) fails the step, the page never reaches build_status='deployed', and
-- UpdatePageStatusAction (v3_site_actions.go:1082) only stamps deployed_at
-- inside that branch. The 2026-08-02 artefact keeps serving, so the artefact and
-- the record disagree while both read correctly.
--
-- [MEASURED 2026-09-02] Fleet-wide, the active pages on which NO component row
-- carries a component_id are exactly these three, and all three are unstamped.
--
-- WHY ADOPTION AND NOT REGENERATION (owner instruction 2026-09-02: "keep the
-- tools if they're working"). create_tool_component builds a component from
-- LLM-generated HTML; running it here would replace three working calculators.
-- Instead the stored, live HTML is promoted to html_template. This is lossless:
-- [MEASURED 2026-09-02] all three stored bodies contain ZERO '{{' bindings, so
-- the rendered form IS the template and there is nothing to bind. That would NOT
-- be true of a data-bound component, and the zero-bindings measurement is the
-- licence for this migration rather than a convenience.
--
-- Shape copied from the six healthy siblings, measured the same day:
--   component_level='tool', section_type NULL, forked_from NULL, is_active,
--   render_mode='template', category='interactive', suitable_site_types='[]',
--   name = '<function>-lendzy-co-uk' (CLC-020 scoped storage identity).
-- created_from is 'adopted', NOT the siblings' 'generated' — these rows are
-- adopted from live HTML and the column exists to say so (chk_created_from_valid
-- permits it).
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/
-- 090 diagnosis: intake 1ff4c475-6977-4631-b641-993735429186,
--                run 89a84ad3-5668-44b3-a089-f9d6c0df7cbb
-- Rollback: 693_..._ROLLBACK.sql   Verify: 693_..._VERIFY.sql

BEGIN;

-- ---------------------------------------------------------------------------
-- GUARD 1. RFC_036 §9.3: if a LIBRARY tool already claims one of these
-- functions, the new row must be born a FORK, and a bare INSERT dies on
-- idx_cc_tool_function_unique (23505). True today (0 rows); another session
-- could make it false between writing and applying, so ASSERT rather than
-- assume.
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
   WHERE component_level = 'tool' AND forked_from IS NULL AND is_active
     AND function IN ('tool-price-cap-checker',
                      'tool-true-cost-calculator',
                      'tool-complaint-deadline-calculator');
  IF n <> 0 THEN
    RAISE EXCEPTION '693 ABORT: % library tool row(s) already claim one of these functions — this migration must fork, not insert (RFC_036 9.3)', n;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- GUARD 2. The three page_components rows must still be unrepaired. If another
-- session has already pointed one at a component, this migration must not
-- overwrite that decision.
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM pages p JOIN page_components pc ON pc.page_id = p.id
   WHERE p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND p.name IN ('tool-price-cap-checker',
                    'tool-true-cost-calculator',
                    'tool-complaint-deadline-calculator')
     AND pc.component_id IS NULL;
  IF n <> 3 THEN
    RAISE EXCEPTION '693 ABORT: expected 3 unrepaired page_components rows, found % — someone has already acted; re-read before applying', n;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- GUARD 3. The stored HTML must still carry no template bindings. This is the
-- measurement the whole approach rests on; if it has changed, adoption is no
-- longer lossless and this migration is wrong.
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM pages p JOIN page_components pc ON pc.page_id = p.id
   WHERE p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND p.name IN ('tool-price-cap-checker',
                    'tool-true-cost-calculator',
                    'tool-complaint-deadline-calculator')
     AND pc.rendered_html LIKE '%{{%';
  IF n <> 0 THEN
    RAISE EXCEPTION '693 ABORT: % stored body/bodies now contain template bindings — promoting rendered_html to html_template is no longer lossless', n;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- THE CHANGE. One component per tool, built FROM the page's own live HTML.
-- ---------------------------------------------------------------------------
INSERT INTO content_components
    (name, display_name, description, html_template, function, section_type,
     component_level, category, render_mode, created_from, is_active,
     forked_from, suitable_site_types, suitable_page_types)
SELECT
    p.name || '-lendzy-co-uk',
    d.display_name,
    d.description,
    pc.rendered_html,
    p.name,
    NULL,
    'tool',
    'interactive',
    'template',
    'adopted',
    true,
    NULL,
    '[]'::jsonb,
    '[]'::jsonb
  FROM pages p
  JOIN page_components pc ON pc.page_id = p.id
  JOIN (VALUES
      ('tool-price-cap-checker',
       'Price cap checker',
       'Checks a high-cost short-term credit balance against the FCA cost cap, adopted 2026-09-02 from the live 2026-08-02 build.'),
      ('tool-true-cost-calculator',
       'True cost calculator',
       'Totals the full cost of a short-term loan including interest and charges, adopted 2026-09-02 from the live 2026-08-02 build.'),
      ('tool-complaint-deadline-calculator',
       'Complaint deadline calculator',
       'Works out the deadline for referring a credit complaint to the Financial Ombudsman Service, adopted 2026-09-02 from the live 2026-08-02 build.')
   ) AS d(page_name, display_name, description) ON d.page_name = p.name
 WHERE p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND pc.component_id IS NULL;

-- Repoint each page_components row at its new component, and correct slot_name
-- from the generic 'section' to the tool function — the shape the six healthy
-- siblings carry, and what makes the slot_name fallback route resolve too.
UPDATE page_components pc
   SET component_id = cc.id,
       slot_name    = p.name,
       updated_at   = NOW()
  FROM pages p, content_components cc
 WHERE pc.page_id = p.id
   AND p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND cc.function = p.name
   AND cc.name = p.name || '-lendzy-co-uk'
   AND pc.component_id IS NULL;

-- ---------------------------------------------------------------------------
-- VERIFY, as DO/RAISE. A verify block of bare SELECTs CANNOT stop the COMMIT —
-- ON_ERROR_STOP ignores a non-empty result set. This block can.
-- ---------------------------------------------------------------------------
DO $$
DECLARE comps int; linked int; orphans int; bound int;
BEGIN
  SELECT count(*) INTO comps FROM content_components
   WHERE name IN ('tool-price-cap-checker-lendzy-co-uk',
                  'tool-true-cost-calculator-lendzy-co-uk',
                  'tool-complaint-deadline-calculator-lendzy-co-uk')
     AND component_level = 'tool' AND is_active AND forked_from IS NULL;
  IF comps <> 3 THEN
    RAISE EXCEPTION '693 VERIFY: expected 3 adopted tool components, found %', comps;
  END IF;

  SELECT count(*) INTO linked
    FROM pages p JOIN page_components pc ON pc.page_id = p.id
    JOIN content_components cc ON cc.id = pc.component_id
   WHERE p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND p.name IN ('tool-price-cap-checker',
                    'tool-true-cost-calculator',
                    'tool-complaint-deadline-calculator')
     AND cc.function = p.name AND pc.slot_name = p.name;
  IF linked <> 3 THEN
    RAISE EXCEPTION '693 VERIFY: expected 3 repointed page_components rows, found %', linked;
  END IF;

  -- The founding query of this lane, re-run: no active page anywhere may be
  -- left with every component row unidentified.
  SELECT count(*) INTO orphans FROM (
      SELECT p.id FROM pages p JOIN page_components pc ON pc.page_id = p.id
       WHERE p.status = 'active'
       GROUP BY p.id
      HAVING count(*) FILTER (WHERE pc.component_id IS NOT NULL) = 0
  ) q;
  IF orphans <> 0 THEN
    RAISE EXCEPTION '693 VERIFY: % active page(s) still have no identified component on any row', orphans;
  END IF;

  -- The adopted templates must remain binding-free, or the renderer is being
  -- handed something this migration never measured.
  SELECT count(*) INTO bound FROM content_components
   WHERE created_from = 'adopted' AND html_template LIKE '%{{%'
     AND name LIKE '%-lendzy-co-uk';
  IF bound <> 0 THEN
    RAISE EXCEPTION '693 VERIFY: % adopted template(s) contain bindings', bound;
  END IF;

  RAISE NOTICE '693 OK: 3 components adopted, 3 page_components repointed, 0 orphan pages fleet-wide';
END $$;

COMMIT;
