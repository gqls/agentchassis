-- 229_vonc_about_swapped_stat_values.sql
-- vonc.com /about publishes two figures with their values transposed.
-- Found 2026-07-27 by the register seeded in 228, on its first pass.
--
-- ── THE DEFECT ─────────────────────────────────────────────────────────────
-- Four components across vonc publish the archetype and tool counts. Three
-- agree with the database; one has the two values swapped:
--
--   page  pos  component            stat_1                   stat_2
--   about   2  content-block-about  Archetypes         = 3    Tools Live         = 8  <-- WRONG
--   about   6  gauntlet-cta         Tools live         = 3    Archetypes in play = 8
--   index   3  gauntlet-cta         Archetypes         = 8    Tools Live         = 3
--   index   4  brief-explanation    Interactive tools live = 3  Written guides live = 2
--
-- Ground truth, re-derived 2026-07-27 (NOT read off the site — that copy is
-- what may be wrong):
--
--   SELECT page_type, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
--   WHERE s.domain='vonc.com' AND p.status='active' GROUP BY 1;
--     entity-page 8  (the eight archetypes)  |  tool 3  |  blog-post 2
--
-- So: 8 archetypes, 3 tools. The about page says 3 archetypes and 8 tools.
-- Both labels are correct and both values are correct — they are on the wrong
-- fields. Hence a swap, not a rewrite: no new figure is introduced by this
-- file, which is why it can be a mechanical correction rather than an
-- editorial one.
--
-- Display order is preserved (Archetypes first), matching index/gauntlet-cta.
--
-- ── SCOPE ──────────────────────────────────────────────────────────────────
-- This edits page_components.content_data ONLY. It does NOT touch
-- rendered_html, so the live page keeps serving the wrong figures until the
-- page is re-rendered. That is deliberate and is the honest state to leave:
-- writing a corrected rendered_html here would make the page LOOK fixed while
-- content_data — the thing every future re-render reads — stayed wrong, which
-- is the inverse of bugs_open/093's whole complaint.
--
-- Verify after a re-render, not after this file.

BEGIN;

DO $fix$
DECLARE
    v_id     uuid;
    v_before jsonb;
    n        int;
BEGIN
    SELECT pc.id, pc.content_data INTO v_id, v_before
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    JOIN sites s ON s.id = p.site_id
    LEFT JOIN content_components cc ON cc.id = pc.component_id
    WHERE s.domain = 'vonc.com'
      AND p.name = 'about'
      AND COALESCE(cc.name,'') = 'content-block-about'
      AND pc.content_data->>'stat_1_label' = 'Archetypes'
      AND pc.content_data->>'stat_1_value' = '3'
      AND pc.content_data->>'stat_2_label' = 'Tools Live'
      AND pc.content_data->>'stat_2_value' = '8';

    -- No row means the state has already moved — another session may have
    -- fixed it, or a rebuild may have rewritten the component. Either way this
    -- file must not guess: fail loudly rather than patch something else.
    IF v_id IS NULL THEN
        RAISE EXCEPTION '229: the exact swapped state (Archetypes=3 / Tools Live=8 on vonc about/content-block-about) was not found — re-survey before forcing this';
    END IF;

    UPDATE page_components
    SET content_data = jsonb_set(
            jsonb_set(content_data, '{stat_1_value}', '"8"'::jsonb),
            '{stat_2_value}', '"3"'::jsonb),
        updated_at = now()
    WHERE id = v_id
      AND content_data ? 'stat_1_value'
      AND content_data ? 'stat_2_value';

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN
        RAISE EXCEPTION '229: expected to update exactly 1 component, updated %', n;
    END IF;

    RAISE NOTICE '229: vonc about/content-block-about % — Archetypes 3->8, Tools Live 8->3', v_id;
END $fix$;

-- ── Post-condition: all four components now agree with the database ────────
DO $post$
DECLARE
    bad int;
BEGIN
    SELECT count(*) INTO bad
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    JOIN sites s ON s.id = p.site_id,
    LATERAL (VALUES ('stat_1_label','stat_1_value'), ('stat_2_label','stat_2_value')) f(lk, vk)
    WHERE s.domain = 'vonc.com'
      AND pc.content_data ? f.vk
      AND (
            (pc.content_data->>f.lk ILIKE '%archetype%' AND pc.content_data->>f.vk <> '8')
         OR (pc.content_data->>f.lk ILIKE '%tool%'      AND pc.content_data->>f.vk <> '3')
          );
    IF bad <> 0 THEN
        RAISE EXCEPTION '229: % archetype/tool stat field(s) still disagree with the live counts (8 archetypes, 3 tools)', bad;
    END IF;
    RAISE NOTICE '229 OK: every archetype/tool stat field on vonc now matches the database';
END $post$;

COMMIT;
