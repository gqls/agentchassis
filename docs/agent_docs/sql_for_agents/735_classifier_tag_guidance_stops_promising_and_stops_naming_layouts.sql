-- 735_classifier_tag_guidance_stops_promising_and_stops_naming_layouts.sql
--
-- bugs_open/445, Phase 0. Two surgical edits to domain-research-classifier's
-- `classify_and_extract` prompt_template. Config, so live on COMMIT — no roll.
--
-- WHY (both edits, one cause):
--
-- 1. THE FALSE PROMISE. The prompt told the model:
--
--      "If no existing tag fits this site well, coin a new one using the same
--       style — an unmatched tag will trigger a library-growth review work item
--       rather than silently fail."
--
--    It did not. `queueLayoutCandidateReview` fired only when the TOTAL score
--    was zero across the whole library, and the category/description/scheme
--    bonuses are added to that total independently of any tag matching — so a
--    coined tag matching nothing produced no signal whatsoever.
--    [MEASURED 2026-09-03] 188 of 216 distinct terms the classifier emitted
--    across 33 sites (87%) matched no layout, and exactly TWO
--    needs_new_layout_candidate items exist across 63,007 work items ever
--    written, both from the degenerate no-tags-at-all arm.
--
--    The code half is fixed in commit 76db94fc7 (weak-fit arm added), which is
--    INERT UNTIL THE NEXT CHASSIS ROLL. So this migration must not simply
--    restate the promise in the present tense — during the window between this
--    applying and that rolling it would still be false. The replacement text
--    asserts NO mechanism at all: it gives the instruction and drops the causal
--    claim, so it is accurate before the roll, after the roll, and if the roll
--    is delayed.
--
-- 2. THE LAYOUT NAMES IN THE EXAMPLES. The site-shape examples included
--    "magazine-grid", "affiliate-hub" and "docs-sidebar" — the layout
--    taxonomy's OWN row names. An LLM given layout names as examples of "shape"
--    echoes them back: 12 live sites carry a layout name in `industry_tags`.
--    They contribute exactly nothing to the score (verified on all three
--    scoring paths: a layout's own name is absent from its `industry_tags`, its
--    `category`, and its `description` — magazine-grid's description opens
--    "Publication layout with featured article…"). So they are pure noise that
--    displaces real tags from a capped 4-10 item list.
--
-- WHY THIS IS WORTH MORE TODAY THAN IT WAS YESTERDAY. Migration 734
-- (portfolio_positioning, applied 2026-09-03 11:39:14Z) added `layout_taxonomy`
-- to this step's `input_fields`. Until then the taxonomy was fetched by
-- `read_layout_taxonomy` and DROPPED at the template boundary, so the model was
-- rendered a literal `null` where the tag list should be and `<no value>` for
-- the layout count — verified in llm_call_log.prompt_rendered. The model was
-- obeying an instruction to match against an empty list. With a real list now
-- rendering, "prefer a tag from the list" becomes a genuine instruction for the
-- first time, and coining becomes a real last resort rather than the only path.
--
-- THE FORM-OVER-INDUSTRY SENTENCE is the load-bearing addition, and it is its
-- own sentence deliberately. 9 of 18 layouts have ZERO sites emitting any of
-- their tags, because their tags name an INDUSTRY (wellness, bakery, artisan;
-- law, consultancy; boxing, martial-arts) while the classifier emits form and
-- platform words. A tag that names the site's form can reach a layout; one that
-- names its vertical usually cannot.
--
-- SAFETY: both edits are exact-string replacements guarded by a count check. If
-- either anchor is absent (another lane rewrote the prompt first), this ABORTS
-- rather than silently no-opping — a prompt migration that matched nothing is
-- indistinguishable from one that worked.

BEGIN;

DO $$
DECLARE
    v_id            uuid;
    v_prompt        text;
    v_new           text;
    v_anchor_promise text := 'The library currently has {{.layout_taxonomy.layout_count}} active layouts. If no existing tag fits this site well, coin a new one using the same style — an unmatched tag will trigger a library-growth review work item rather than silently fail.';
    v_anchor_shapes  text := '- site-shape tags (e.g. "interactive-platform", "tool-portal", "social-network", "provocation-platform", "magazine-grid", "affiliate-hub", "docs-sidebar")';
    v_repl_promise   text := 'The library currently has {{.layout_taxonomy.layout_count}} active layouts. Prefer a tag from the list above whenever one describes this site — each match is an overlap point for the layout matcher. Coin a new tag only when nothing in the list fits.

Prefer tags that name the site''s FORM over tags that name its industry. "long-form", "tool-portal", "content-hub", "comparison" describe how a site is SHAPED and can be matched against a layout; "fintech", "veterinary", "uk-farming" describe its subject and usually cannot. Include vertical tags as well, but never instead.';
    v_repl_shapes    text := '- site-shape tags (e.g. "interactive-platform", "tool-portal", "content-hub", "social-network", "provocation-platform", "long-form", "comparison")';
BEGIN
    SELECT id, default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
      INTO v_id, v_prompt
      FROM agent_definitions
     WHERE type = 'domain-research-classifier'
       AND is_active
       AND COALESCE(is_snapshot, false) = false
       AND deleted_at IS NULL;

    IF v_id IS NULL THEN
        RAISE EXCEPTION '735: no live domain-research-classifier row found. Aborting.';
    END IF;

    -- Guard 1: both anchors must be present exactly once.
    IF position(v_anchor_promise in v_prompt) = 0 THEN
        RAISE EXCEPTION '735: the false-promise sentence is not present verbatim. Another lane has rewritten this prompt; re-derive the anchor from the live row before re-running. Aborting rather than no-opping.';
    END IF;
    IF position(v_anchor_shapes in v_prompt) = 0 THEN
        RAISE EXCEPTION '735: the site-shape example line is not present verbatim. Another lane has rewritten this prompt; re-derive the anchor. Aborting rather than no-opping.';
    END IF;

    v_new := replace(v_prompt, v_anchor_promise, v_repl_promise);
    v_new := replace(v_new,    v_anchor_shapes,  v_repl_shapes);

    -- Guard 2: the replacement must actually have changed the text, and must
    -- have removed the promise. A `replace` that silently matched nothing
    -- leaves the string identical, which reads exactly like success.
    IF v_new = v_prompt THEN
        RAISE EXCEPTION '735: replacement produced an identical prompt. Aborting.';
    END IF;
    IF position('will trigger a library-growth review work item' in v_new) > 0 THEN
        RAISE EXCEPTION '735: the promise survived the replacement. Aborting.';
    END IF;
    IF position('"magazine-grid"' in v_new) > 0 OR position('"affiliate-hub"' in v_new) > 0 THEN
        RAISE EXCEPTION '735: a layout name survived in the shape examples. Aborting.';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(
               default_config,
               '{workflow,steps,classify_and_extract,config,prompt_template}',
               to_jsonb(v_new),
               false
           ),
           updated_at = NOW()
     WHERE id = v_id;

    RAISE NOTICE '735: classifier tag guidance updated (prompt %s -> %s chars)',
                 length(v_prompt), length(v_new);
END $$;

-- Post-check: a DO block that RAISEs, not a bare SELECT. ON_ERROR_STOP ignores a
-- non-empty result set, so a verify block made of SELECTs cannot stop the COMMIT.
DO $$
DECLARE v_prompt text;
BEGIN
    SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
      INTO v_prompt
      FROM agent_definitions
     WHERE type = 'domain-research-classifier' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('will trigger a library-growth review work item' in v_prompt) > 0 THEN
        RAISE EXCEPTION '735 VERIFY: the false promise is still live.';
    END IF;
    IF position('Prefer tags that name the site''s FORM' in v_prompt) = 0 THEN
        RAISE EXCEPTION '735 VERIFY: the form-over-industry guidance did not land.';
    END IF;
    IF position('"magazine-grid"' in v_prompt) > 0 THEN
        RAISE EXCEPTION '735 VERIFY: a layout name is still offered as a shape example.';
    END IF;
    RAISE NOTICE '735 VERIFY: OK — promise removed, form guidance present, layout names gone.';
END $$;

COMMIT;
