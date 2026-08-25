-- 614_tool_cta_cards_render_the_listing_image_ROLLBACK.sql
--
-- Restores tool-cta's html_template and input_schema from the pre-image that
-- 614 copied to content_components_bak_20260825_toolcta_image.
--
-- WHEN TO RUN THIS. The visual change is fleet-wide (10 domains, 44 active
-- pages) and the owner may simply not want the thumbnails. The other trigger is
-- LOAD: rendering .image puts tool-cta inside the page-image consumer set, so
-- card landings on those sites now file page_rerender items. If that rate is
-- materially above what the lane recorded, roll this back and bring the number
-- to the owner — the query is in 603's header.
--
-- WHAT THIS DOES NOT UNDO, deliberately:
--   * the `template_changed` page_rerender items already filed (D3). They are
--     ordinary re-render requests and complete on their own; after this
--     rollback they simply re-render the pre-image template, which is correct.
--   * the 29 card assets derived for loancalculator.co.uk and
--     loanandmortgagecalculator.co.uk. Those are real listing crops and are
--     used by every image-rendering consumer, not just tool-cta.
--   * the `image` key inside stored content_data arrays. It was there before
--     614 and will be there after — 614 never wrote item data.
--
-- AFTER ROLLING BACK, the pages still serve the thumbnailed HTML until
-- something re-renders them. File a fan-out the same way D3 did, or the
-- rollback is invisible on the deployed artefact.

BEGIN;

DO $$
DECLARE n_bak int; n_live int;
BEGIN
    SELECT count(*) INTO n_bak FROM content_components_bak_20260825_toolcta_image WHERE is_active;
    IF n_bak <> 1 THEN
        RAISE EXCEPTION '614 rollback: expected exactly 1 active row in the pre-image table, found % — refusing to guess', n_bak;
    END IF;
    SELECT count(*) INTO n_live FROM content_components WHERE name = 'tool-cta' AND is_active;
    IF n_live <> 1 THEN
        RAISE EXCEPTION '614 rollback: expected exactly 1 active live tool-cta row, found %', n_live;
    END IF;
END $$;

UPDATE content_components c
   SET html_template = b.html_template,
       input_schema  = b.input_schema,
       updated_at    = NOW()
  FROM content_components_bak_20260825_toolcta_image b
 WHERE c.name = 'tool-cta' AND c.is_active AND b.is_active;

DO $$
DECLARE tmpl text; sch jsonb;
BEGIN
    SELECT html_template, input_schema INTO tmpl, sch
      FROM content_components WHERE name = 'tool-cta' AND is_active;
    IF tmpl ~* '\.image\y' THEN
        RAISE EXCEPTION '614 rollback verify: tool-cta still renders .image — the restore did not take';
    END IF;
    IF sch #> '{fields,items,items}' ? 'image' THEN
        RAISE EXCEPTION '614 rollback verify: input_schema still declares the image item key';
    END IF;
    IF tmpl NOT LIKE '%{{.title}}%' OR tmpl NOT LIKE '%{{.nav_label}}%' THEN
        RAISE EXCEPTION '614 rollback verify: the restored template is missing pre-existing card fields';
    END IF;
END $$;

COMMIT;
