-- 085h_READY_place_hero_carousel_on_index.sql   *** PREPARED, NOT RUN ***
--
-- The owner said "I only see one carousel". He is right, and the reason is
-- sharper than scarcity: THE HOMEPAGE HAS NONE.
--
--   SELECT p.name, cc.function FROM page_components pc
--     JOIN content_components cc ON cc.id=pc.component_id
--     JOIN pages p ON p.id=pc.page_id
--    WHERE p.site_id='199733a8-...' AND cc.function LIKE '%carousel%';
--   -> capabilities | hero-card-carousel
--      multi-agent-review-council | swipeable-insight-carousel
--      self-correction-leopardessconsulting | swipeable-insight-carousel
--
-- The mission brief opens its visual direction with "an auto-advancing hero
-- carousel of a few cards (each with a title, one short teaser line, and a
-- 'read more' link, no more copy than that)". The homepage instead runs a
-- static hero straight into a stat band. So this is not a taste question — it
-- is a brief requirement that was never placed.
--
-- NOT RUN, for one reason and one only: until the corrected styles.css is
-- published, this component's card links render dark-on-dark like everything
-- else. Placing a carousel whose "read more" is invisible makes the page worse,
-- not better. RUN THIS AFTER the stylesheet lands, then re-render index.
--
-- Placement is position 2 — after the hero, before the stat band — rather than
-- REPLACING the hero. The brief's phrase is "hero carousel", but replacing the
-- homepage's opening statement is an owner decision about what the site leads
-- with, and this file will not make it silently. Position 2 satisfies "a
-- variety of section types down the page" and gives the homepage a carousel;
-- if the owner wants it as the hero proper, delete the hero row instead.
--
-- Cards reuse the four the capabilities page already carries, which are
-- grounded, owner-approved copy with images that now resolve. `autoplay` is
-- OFF, per the owner's standing ruling that movement is opt-in.

BEGIN;

-- shift everything from position 2 down
UPDATE page_components
   SET position = position + 1, updated_at = NOW()
 WHERE page_id = (SELECT id FROM pages
                   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='index')
   AND position >= 2;

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, created_at, updated_at)
SELECT (SELECT id FROM pages WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='index'),
       (SELECT id FROM content_components WHERE function='hero-card-carousel' AND is_active LIMIT 1),
       'hero-card-carousel', 2,
       (SELECT pc.content_data
          FROM page_components pc
          JOIN content_components cc ON cc.id = pc.component_id
          JOIN pages p ON p.id = pc.page_id
         WHERE p.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
           AND p.name='capabilities' AND cc.function='hero-card-carousel'),
       NOW(), NOW();

-- and re-render
INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT p.site_id, 'page_rerender',
       'page_rerender_index_199733a8_hero_carousel_placement_' || to_char(NOW(),'YYYYMMDDHH24MI'),
       'triaged', 'build',
       'Republish index: place the hero card carousel the brief asks for',
       jsonb_build_object('domain','fundamentallyai.com','page_id',p.id::text,
                          'page_name','index','filename','index.html',
                          'reason','section_data_resolved'),
       'page-rerender','operator:brochure_component_library','operator:brochure_component_library',
       0, 3, NOW(), NOW()
  FROM pages p
 WHERE p.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND p.name='index';

COMMIT;

-- Then, always:
--   scripts/render_audit.py https://fundamentallyai.com/index.html
