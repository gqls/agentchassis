-- ROLLBACK for 735. Restores the original two passages verbatim.
--
-- NOTE the asymmetry, deliberately: rolling this back restores a sentence that
-- is FALSE until chassis commit 76db94fc7 rolls, and unhelpful afterwards. It
-- exists because a migration without a rollback is not reviewable, not because
-- reverting is advisable.
BEGIN;
DO $$
DECLARE
    v_id uuid; v_prompt text; v_new text;
    v_cur_promise text := 'The library currently has {{.layout_taxonomy.layout_count}} active layouts. Prefer a tag from the list above whenever one describes this site — each match is an overlap point for the layout matcher. Coin a new tag only when nothing in the list fits.

Prefer tags that name the site''s FORM over tags that name its industry. "long-form", "tool-portal", "content-hub", "comparison" describe how a site is SHAPED and can be matched against a layout; "fintech", "veterinary", "uk-farming" describe its subject and usually cannot. Include vertical tags as well, but never instead.';
    v_old_promise text := 'The library currently has {{.layout_taxonomy.layout_count}} active layouts. If no existing tag fits this site well, coin a new one using the same style — an unmatched tag will trigger a library-growth review work item rather than silently fail.';
    v_cur_shapes text := '- site-shape tags (e.g. "interactive-platform", "tool-portal", "content-hub", "social-network", "provocation-platform", "long-form", "comparison")';
    v_old_shapes text := '- site-shape tags (e.g. "interactive-platform", "tool-portal", "social-network", "provocation-platform", "magazine-grid", "affiliate-hub", "docs-sidebar")';
BEGIN
    SELECT id, default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
      INTO v_id, v_prompt FROM agent_definitions
     WHERE type='domain-research-classifier' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_id IS NULL THEN RAISE EXCEPTION '735 ROLLBACK: no live classifier row.'; END IF;
    IF position(v_cur_promise in v_prompt) = 0 THEN
        RAISE EXCEPTION '735 ROLLBACK: 735 text not present — nothing to roll back, or another lane has since edited. Aborting.';
    END IF;
    v_new := replace(v_prompt, v_cur_promise, v_old_promise);
    v_new := replace(v_new,    v_cur_shapes,  v_old_shapes);
    IF v_new = v_prompt THEN RAISE EXCEPTION '735 ROLLBACK: no change produced. Aborting.'; END IF;
    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             '{workflow,steps,classify_and_extract,config,prompt_template}', to_jsonb(v_new), false),
           updated_at = NOW()
     WHERE id = v_id;
    RAISE NOTICE '735 ROLLBACK: original tag guidance restored.';
END $$;
COMMIT;
