-- 350_pin_design_intent_palette_for_the_three_unpinned_sites.sql
--
-- bugs_open/122, propagation leg. PRECONDITION for re-rendering stylesheets.
--
-- WHY THIS EXISTS. Migration 338 repointed component and layout inks at the
-- renderer's new legible-ink slots. Those slots are DEFINED only by the engine's
-- step 12, which runs during a stylesheet render — so until each site's
-- styles.css regenerates, every reference 338 added resolves to its fallback and
-- renders exactly as today. The fix is applied and inert until a CSS re-render.
--
-- The only live agent holding render_css_from_spec is webdesign-agent, and its
-- step graph routes every entry through analyze_design (execute_llm_prompt)
-- before generate_css — there is no path that skips it. analyze_design invents a
-- fresh palette on any site lacking a STRUCTURED design_intent.palette block:
-- the recorded fleet mechanism from 2026-07-17 on robot-hands (four CSS rewrites
-- in one day, one of which put a LIGHT background on a dark site, live).
--
-- 9 of the 12 sites this propagation touches are already pinned — most by
-- domain-research-classifier in normal operation, so a pin is the platform's
-- default posture, not an intervention. These are the three that are not.
--
-- Pattern and precedent: robot_hands/SQL_2026-07-17_r1b_design_intent_palette_pin.sql
-- (proven — the next run reproduced the pinned values exactly). Same shape here:
-- supersede the current design_intent row, keep every existing field, ADD the
-- structured palette block.
--
-- ============================================================================
-- WHERE THE VALUES COME FROM, AND WHY THEY INCLUDE FAILING COLOURS
-- ============================================================================
-- Every reference_value below was read from the site's SERVED stylesheet on
-- 2026-08-09 (curl of /assets/css/styles.css), NOT from a DB copy — palettes,
-- css_themes.color_palette and style_collections.color_palette are three
-- separate stores that drift, and the served file is what a visitor actually got.
--
-- These pins deliberately preserve TODAY'S LOOK, including the accents that
-- currently fail contrast (gaswholesalers #E8A020 on #F4F1EB, finetuning
-- #C8873A). That is correct and is the whole point: 338 does not change any
-- FILL, it routes INKS through --color-accent-ink / --color-primary-ink, which
-- the renderer derives per-ground at render time. Pinning the fill is what stops
-- the LLM repainting the site; the ink companion is what fixes the contrast.
-- A pin that "corrected" the accent here would be a second, unreviewed change
-- riding along inside a safety measure.
--
-- The guidance text therefore says KEEP, not IMPROVE.
--
-- NOTE (not fixed here, deliberately): --color-primary-hover is #1e3a8a on all
-- three sites, which is almost certainly a shared default rather than three
-- independent choices, and ai-agent-orchestration's --color-primary (#0D1117) is
-- byte-identical to its --color-surface. Both are preserved as-is. They are
-- pre-existing oddities and a palette pin is not the place to decide them.
--
-- ============================================================================
-- BACKUP — take before applying, outside the transaction
-- ============================================================================
-- COPY (SELECT id,site_id,aspect,data,is_current FROM site_specs
--        WHERE aspect='design_intent' AND is_current
--          AND site_id IN ('2a8ebf9c-20a2-4c39-b191-840b012371da',
--                          '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
--                          '5fe15466-4e2e-4ff2-981e-98c1b7074002')) TO STDOUT
--
-- ROLLBACK is at the foot of this file.

\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    sites  jsonb := '[
      {"id":"2a8ebf9c-20a2-4c39-b191-840b012371da","domain":"ai-agent-orchestration.com","scheme":"dark",
       "character":"Dark technical console. Near-black blue-tinted ground, light high-contrast text, a single warm amber accent. Observed scheme on 2026-08-09: DARK (background #080B10).",
       "values":{"primary":"#0D1117","primary_hover":"#1e3a8a","primary_text":"#ffffff","secondary":"#161B22","accent":"#F0A500","background":"#080B10","surface":"#0D1117","text":"#E6EDF3","text_muted":"#8B949E","border":"#21262D"}},
      {"id":"1368e337-dd1d-4799-bbb3-8221a1b79bcc","domain":"finetuning.uk","scheme":"light",
       "character":"Warm light editorial. Off-white paper ground, deep navy-black ink, a muted bronze accent. Observed scheme on 2026-08-09: LIGHT (background #F5F3EF).",
       "values":{"primary":"#1A1A2E","primary_hover":"#1e3a8a","primary_text":"#ffffff","secondary":"#E8E4DC","accent":"#C8873A","background":"#F5F3EF","surface":"#FFFFFF","text":"#1A1A2E","text_muted":"#6B6860","border":"#D4CFC6"}},
      {"id":"5fe15466-4e2e-4ff2-981e-98c1b7074002","domain":"gaswholesalers.com","scheme":"light",
       "character":"Warm light commercial brochure. Cream ground, near-black text, amber/ochre accents. Observed scheme on 2026-08-09: LIGHT (background #F4F1EB).",
       "values":{"primary":"#1A1A2E","primary_hover":"#1e3a8a","primary_text":"#ffffff","secondary":"#C8880A","accent":"#E8A020","background":"#F4F1EB","surface":"#FFFFFF","text":"#1C1C1C","text_muted":"#5C5C5C","border":"#D6CFC2"}}
    ]'::jsonb;
    rec      jsonb;
    sid      uuid;
    dom      text;
    pre_cnt  int;
    ins_cnt  int;
    got_bg   text;
    want_bg  text;
    guidance text;
BEGIN
    FOR rec IN SELECT * FROM jsonb_array_elements(sites) LOOP
        sid := (rec->>'id')::uuid;
        dom := rec->>'domain';

        -- IDEMPOTENCY / precondition: exactly one current row, and not already pinned.
        SELECT count(*) INTO pre_cnt FROM site_specs
         WHERE site_id = sid AND aspect = 'design_intent' AND is_current;
        IF pre_cnt <> 1 THEN
            RAISE EXCEPTION '350/%: expected exactly 1 current design_intent row, found %. '
                'The supersede pattern INSERTs FROM the superseded row, so 0 rows would '
                'silently write nothing. STOP and read the site''s specs.', dom, pre_cnt;
        END IF;

        PERFORM 1 FROM site_specs
         WHERE site_id = sid AND aspect = 'design_intent' AND is_current
           AND data->'palette'->'reference_values' IS NOT NULL;
        IF FOUND THEN
            RAISE EXCEPTION '350/%: already pinned — another session got here first. STOP.', dom;
        END IF;

        guidance := format(
            'KEEP THESE VALUES. They are the site''s live colours as served on 2026-08-09 and '
            'are pinned to stop analyze_design re-inventing the palette on each run '
            '(robot-hands, 2026-07-17: four rewrites in one day, one shipped a light background '
            'onto a dark site). Observed scheme is %s and must not change. Do NOT "improve" the '
            'accent for contrast: legible ink variants are derived by the renderer as '
            '--color-accent-ink / --color-primary-ink (bugs_open/122, migration 338), and '
            'changing the fill here would undo that design. Adjust hues only within the stated '
            'scheme and only where the brand character clearly demands it.',
            upper(rec->>'scheme'));

        WITH old AS (
            UPDATE site_specs
               SET is_current = false, superseded_at = now()
             WHERE site_id = sid AND aspect = 'design_intent' AND is_current
            RETURNING site_id, data
        )
        INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
        SELECT site_id, 'design_intent',
               data || jsonb_build_object('palette', jsonb_build_object(
                   'character',        rec->>'character',
                   'reference_values', rec->'values',
                   'guidance',         guidance)),
               'manual-recovery',
               '350 palette pin (bugs_open/122 propagation): values read from the SERVED '
               'stylesheet 2026-08-09; pinned so the CSS re-render that delivers 338 cannot '
               'repaint the site. Accents deliberately preserved as-is including failing ones — '
               '338 fixes inks via the derived -ink companions, not by changing fills.',
               true,
               'bugfix-122-propagation'
          FROM old;
        GET DIAGNOSTICS ins_cnt = ROW_COUNT;
        IF ins_cnt <> 1 THEN
            RAISE EXCEPTION '350/%: expected 1 row inserted, got %', dom, ins_cnt;
        END IF;

        -- post-state: the pin is readable, and the pre-existing fields survived
        want_bg := rec->'values'->>'background';
        SELECT data->'palette'->'reference_values'->>'background' INTO got_bg
          FROM site_specs
         WHERE site_id = sid AND aspect = 'design_intent' AND is_current;
        IF got_bg IS DISTINCT FROM want_bg THEN
            RAISE EXCEPTION '350/%: pin not readable back (background=% want %)', dom, got_bg, want_bg;
        END IF;

        PERFORM 1 FROM site_specs
         WHERE site_id = sid AND aspect = 'design_intent' AND is_current
           AND data ? 'colour_mood' AND data ? 'style_direction';
        IF NOT FOUND THEN
            RAISE EXCEPTION '350/%: pre-existing design_intent fields were lost by the merge', dom;
        END IF;

        RAISE NOTICE '350: pinned % (% scheme, bg %)', dom, rec->>'scheme', want_bg;
    END LOOP;

    -- fleet assertion: all 12 propagation targets now pinned
    SELECT count(*) INTO pre_cnt
      FROM sites s
     WHERE s.domain IN ('gaswholesalers.com','gamesdesign.co.uk','robot-hands.com','idea.uk',
                        'finetuning.uk','dartsonline.com','vonc.com','ai-agent-orchestration.com',
                        'leopardessconsulting.co.uk','webdesign.co.uk','mortgagecalculator.co.uk',
                        'lendzy.co.uk')
       AND EXISTS (SELECT 1 FROM site_specs ss
                    WHERE ss.site_id = s.id AND ss.aspect = 'design_intent' AND ss.is_current
                      AND ss.data->'palette'->'reference_values' IS NOT NULL);
    IF pre_cnt <> 12 THEN
        RAISE EXCEPTION '350: expected all 12 propagation targets pinned after this, found % — '
            'do NOT dispatch the CSS re-render until this is 12', pre_cnt;
    END IF;
    RAISE NOTICE '350: all 12 propagation targets pinned';
END $$;

COMMIT;

-- ============================================================================
-- AFTER THIS: the CSS re-render dispatch. This file only removes the blocker.
-- Verify a canary FIRST (one site, served stylesheet diffed against its pin)
-- before dispatching the remaining eleven. bugs_open/122 handoff §2b.
-- ============================================================================
--
-- ROLLBACK — restore the superseded rows and drop the pinned ones.
-- BEGIN;
-- DELETE FROM site_specs
--  WHERE aspect='design_intent' AND is_current AND created_by='bugfix-122-propagation'
--    AND site_id IN ('2a8ebf9c-20a2-4c39-b191-840b012371da',
--                    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
--                    '5fe15466-4e2e-4ff2-981e-98c1b7074002');
-- UPDATE site_specs SET is_current=true, superseded_at=NULL
--  WHERE aspect='design_intent' AND NOT is_current AND superseded_at IS NOT NULL
--    AND site_id IN ('2a8ebf9c-20a2-4c39-b191-840b012371da',
--                    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
--                    '5fe15466-4e2e-4ff2-981e-98c1b7074002')
--    AND superseded_at > now() - interval '1 day';   -- scope to THIS migration's supersede
-- COMMIT;
