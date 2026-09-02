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
-- REVISED 2026-09-02 after council round 1 returned REVISE on a GATING objection
-- from bug_historian/editquality, and the objection was RIGHT:
--
--   "The edit fixes the resolveComponent input (component_id/slot_name) but
--    includes no mechanism to actually trigger a rerender afterward.
--    UpdatePageStatusAction only stamps deployed_at when a rerender succeeds and
--    reaches the deployed branch ... and no edit causes it to fire."
--
-- Round 1 named this as risk (5) and left it as follow-up prose. That is not the
-- same as shipping it: as submitted, this migration satisfied NONE of its own
-- stated acceptance criteria on its own. Section 4 below now files the three
-- rerenders in the same transaction, so the change drives itself.
--
-- The council's own read-only check also CORRECTED round 1's framing. It found
-- page_rerender items ARE filed against these pages periodically (a batch of 20+
-- on 2026-09-01 10:45, source 'rerender-pages'), so "needs_rebuild has no
-- consumer" was too strong for this site — the items are filed and then FAIL.
-- Filing our own is therefore belt-and-braces rather than the only route, and it
-- is still correct to file: relying on someone else's scheduled batch to notice
-- is how a repair sits inert for a fortnight and reads as done.
--
-- ROUND 3 (2026-09-02), answering all four round-2 findings with evidence:
--
-- (a) prior_art_librarian HIGH — the rerender INSERT omitted the FIRST-CLASS
--     page_id column (set only inside spec) and, worse, the NOT-NULL created_by
--     column entirely: as written, round 2 would have ERRORED at apply. Both
--     columns are now set, matching the live producer's rows
--     [MEASURED 2026-09-02: all 7,481 'rerender-pages' items carry page_id
--     non-NULL, created_by='rerender-pages']; the verify now asserts
--     page_id IS NOT NULL on the queued rows.
--
-- (b) prior_art_librarian MEDIUM — the named prior class EXISTS and is
--     bugs_open/357 ("a whole tool page is stored in a slot that claims to be a
--     hero component", OWNED, active session, 090 run 63d4d1a7 pending). Same
--     family — a whole working tool stored under a FALSE component identity,
--     one row per page — different arm: 357's rows carry a WRONG component_id
--     (the shared hero), so a regeneration would swap a working 16KB tool for a
--     2KB title band, and their park is load-bearing; lendzy's rows carry a
--     NULL component_id, so nothing resolves and the page can never deploy.
--     357's §3 hazard CANNOT fire here, structurally: adoption makes the
--     declared template BYTE-IDENTICAL to the stored tool, so the regeneration
--     that is their disaster is our no-op. (And adoption may be exactly their
--     fix shape — CONTRIB'd to their session, not decided for them.)
--
-- (c) bug_historian — "will resolveComponent actually resolve the new rows, or
--     silently drop them at the template guard?" ANSWERED WITH THE PRODUCTION
--     FUNCTION, not inspection: a one-shot in-package test ran all three
--     adopted bodies through toolTemplateValid (the guard
--     loadComponentSchemasByID applies to component_level='tool') — all three
--     PASS, and a >100-char truncated control is REJECTED in the same run.
--     (The first control was VACUOUS — 31 chars, under the guard's 100-char
--     stub floor, so it passed while proving nothing; the lane NOTES log it.)
--
-- (d) editquality's 'missing' — how the 47 unbuilt_internal_link items clear:
--     they carry a REGISTERED verifier (check_phantom_internal_links.go:451,
--     RegisterVerifier("unbuilt_internal_link", VerifyUnbuiltInternalLinkResolved))
--     whose shipped/unbuilt judgement is datahelpers.NeverDeployedPagePredicate —
--     the moment these pages stamp, revalidation resolves the items without any
--     link being edited. No edit here retracts them by hand, deliberately:
--     hand-closing 47 rows the verifier exists to judge would blind the very
--     mechanism that filed them.
--
-- POST-CONDITIONS FOR THE OPERATOR (the migration cannot observe its async
-- outcome — run these after the three rerenders reach 'complete'):
--   1. SELECT name, build_status, deployed_at FROM pages
--       WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
--         AND name IN ('tool-price-cap-checker','tool-true-cost-calculator',
--                      'tool-complaint-deadline-calculator');
--      -- all three: build_status='deployed', deployed_at NOT NULL
--   2. At the ARTEFACT, with the invented-URL control: the three URLs serve 200
--      and their <input> counts are UNCHANGED (3 / 1 / 2) — RUNBOOK §2.
--   3. curl https://lendzy.co.uk/sitemap.xml | grep -c '<loc>'  -- expect 30
--   4. The 47: SELECT count(*) FROM site_work_items
--       WHERE site_id='8ff093d5-...' AND item_type='unbuilt_internal_link'
--         AND status='needs_human_review';  -- falling to 0 as revalidation runs
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
-- 4. DRIVE THE REBUILD. Repairing the input is inert on its own: deployed_at is
-- only written inside UpdatePageStatusAction's `newStatus == "deployed"` branch,
-- which is reached only by a rerender that SUCCEEDS. So file one page_rerender
-- per page, in the same transaction as the repair, in the exact shape the
-- 'rerender-pages' producer uses (measured on this site's own rows 2026-09-02:
-- handler_agent 'page-rerender', priority 80, status 'triaged' — the status
-- idx_swi_handler indexes for pickup — and a spec carrying domain, page_id,
-- filename and page_name).
--
-- idx_swi_dedup is UNIQUE on (site_id, item_key) EXCLUDING terminal statuses, so
-- re-filing a key whose previous rows are 'complete' is permitted; a LIVE row
-- with the same key would collide, and the guard below turns that collision into
-- an abort with a readable reason instead of a 23505.
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND item_key IN ('page_rerender_tool-price-cap-checker_8ff093d5-1f19-453b-9439-a10379bbcd76_assemble',
                      'page_rerender_tool-true-cost-calculator_8ff093d5-1f19-453b-9439-a10379bbcd76_assemble',
                      'page_rerender_tool-complaint-deadline-calculator_8ff093d5-1f19-453b-9439-a10379bbcd76_assemble')
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION '693 ABORT: % live page_rerender item(s) already queued for these pages — a rebuild is already pending, do not stack another', n;
  END IF;
END $$;

INSERT INTO site_work_items
    (site_id, item_type, item_key, status, handler_agent, priority, source, summary,
     page_id, created_by, spec)
SELECT
    p.site_id,
    'page_rerender',
    'page_rerender_' || p.name || '_' || p.site_id || '_assemble',
    'triaged',
    'page-rerender',
    80,
    'lendzy_co_uk lane (migration 693)',
    'Rerender page: ' || p.name,
    p.id,
    'lendzy_co_uk lane (migration 693)',
    jsonb_build_object(
        'domain',    'lendzy.co.uk',
        'page_id',   p.id::text,
        'filename',  ltrim(p.url, '/'),
        'page_name', p.name,
        'reason',    'component adopted by migration 693 — first rerender since the component_id was NULL')
  FROM pages p
 WHERE p.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND p.name IN ('tool-price-cap-checker',
                  'tool-true-cost-calculator',
                  'tool-complaint-deadline-calculator');

-- ---------------------------------------------------------------------------
-- VERIFY, as DO/RAISE. A verify block of bare SELECTs CANNOT stop the COMMIT —
-- ON_ERROR_STOP ignores a non-empty result set. This block can.
-- ---------------------------------------------------------------------------
DO $$
DECLARE comps int; linked int; orphans int; bound int; queued int;
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

  -- The rebuild must actually be queued, or this migration is inert and its own
  -- acceptance criteria (deployed_at, sitemap 30, the 47 items) cannot be met.
  SELECT count(*) INTO queued FROM site_work_items
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND item_type = 'page_rerender'
     AND source = 'lendzy_co_uk lane (migration 693)'
     AND status = 'triaged' AND page_id IS NOT NULL;
  IF queued <> 3 THEN
    RAISE EXCEPTION '693 VERIFY: expected 3 queued rerenders, found % — the repair would be inert', queued;
  END IF;

  RAISE NOTICE '693 OK: 3 components adopted, 3 page_components repointed, 3 rerenders queued, 0 orphan pages fleet-wide';
END $$;

COMMIT;
