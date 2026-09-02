-- 721_page_scope_hero_components_declare_their_image_field_ROLLBACK.sql
--
-- Removes the background_image field 721 added to the six page-scope hero
-- components, restoring them to declaring no image field.
--
-- ⚠ WHAT ROLLING BACK RESTORES: the defect. Those pages go back to rendering
-- the SITE-WIDE homepage hero while their own generated page-scope asset sits
-- orphaned — and back to passing every "does the page have a hero image" check
-- while doing it, because a wrong image is indistinguishable from a right one
-- to a presence check.
--
-- ══ WHY A KEY DELETION AND NOT A CAPTURED LITERAL ════════════════════════════
-- 721 adds one key with jsonb_set(..., create=true). Deleting that key is its
-- EXACT inverse and is safer than restoring a literal snapshot: a literal would
-- also clobber any OTHER field another session has added to these schemas since
-- 721 applied, silently. The guard below refuses if the field is not the one
-- 721 wrote, so this cannot delete somebody else's differently-shaped field.

BEGIN;

DO $$
DECLARE n int; mismatched int;
BEGIN
    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active
       AND function IN ('hero-tool','hero-about','hero-contact',
                        'hero-services','hero-case-studies','hero-use-cases')
       AND input_schema->'fields' ? 'background_image';
    IF n = 0 THEN
        RAISE EXCEPTION 'REFUSING: none of the six carries background_image — 721 is not '
                        'applied, or has already been rolled back';
    END IF;

    -- Refuse to delete a field that is not the one 721 wrote.
    SELECT count(*) INTO mismatched
      FROM content_components
     WHERE is_active
       AND function IN ('hero-tool','hero-about','hero-contact',
                        'hero-services','hero-case-studies','hero-use-cases')
       AND input_schema->'fields' ? 'background_image'
       AND (input_schema->'fields'->'background_image'->>'source' IS DISTINCT FROM 'site_assets.hero'
            OR input_schema->'fields'->'background_image'->>'type' IS DISTINCT FROM 'image');
    IF mismatched > 0 THEN
        RAISE EXCEPTION 'REFUSING: % component(s) carry a background_image that is NOT the one '
                        '721 wrote — another session has edited it. Read it before deleting.',
                        mismatched;
    END IF;
END $$;

UPDATE content_components
   SET input_schema = input_schema #- '{fields,background_image}',
       updated_at = now()
 WHERE is_active
   AND function IN ('hero-tool','hero-about','hero-contact',
                    'hero-services','hero-case-studies','hero-use-cases');

DO $$
DECLARE remaining int; lost int;
BEGIN
    SELECT count(*) INTO remaining FROM content_components
     WHERE is_active AND function IN ('hero-tool','hero-about','hero-contact',
                                      'hero-services','hero-case-studies','hero-use-cases')
       AND input_schema->'fields' ? 'background_image';
    SELECT count(*) INTO lost FROM content_components
     WHERE is_active AND function IN ('hero-tool','hero-about','hero-contact',
                                      'hero-services','hero-case-studies','hero-use-cases')
       AND NOT (input_schema->'fields' ? 'headline');
    IF remaining > 0 THEN
        RAISE EXCEPTION 'ABORT: % still carry background_image', remaining;
    END IF;
    IF lost > 0 THEN
        RAISE EXCEPTION 'ABORT: % lost their headline field — the delete hit the wrong path', lost;
    END IF;
    RAISE NOTICE '721 rolled back: background_image removed, headline fields intact';
END $$;

COMMIT;
