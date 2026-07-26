-- 219b_content_reviewer_declares_sections_metadata.sql
-- Council round 5 (submission 569241fb), MEDIUM: 219 declared
-- require_sections_metadata on page-build-handler only, leaving content-reviewer
-- — same `validate_page_content` action, same DEFAULT html_field, i.e. the same
-- page_content path — undeclared, on the grounds that it is dormant.
--
-- The objection is right and the reasoning was weak: **dormancy today is not
-- closure.** If content-reviewer is ever re-enabled over section-built pages it
-- inherits exactly the silent-skip ambiguity the round-1 fix exists to remove,
-- and nothing would force the declaration to be added at that moment. Relying on
-- a currently-quiet caller staying quiet is the same shape as guarding one call
-- site of a generic mechanism.
--
-- Declaring it now costs nothing (the agent has zero runs in the retained
-- orchestration window, so there is no behaviour to change today) and removes the
-- latent gap rather than documenting it. The finding it enables is a WARNING, so
-- even when content-reviewer does run it cannot block anything.
--
-- Measured before writing (the round-4 lesson — check the premise, do not assert
-- it): active, non-snapshot agents with a step whose action is
-- validate_page_content are exactly page-build-handler, tool-recreation-handler
-- and content-reviewer. tool-recreation-handler stays UNDECLARED deliberately —
-- its html_field is completeness_check.clean_html, a tool blob with no sections,
-- so a warning there would be a false positive every time.
--
-- Idempotent.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('content-reviewer', '219b: pre-update (require_sections_metadata)');

DO $d$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='content-reviewer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,validate_content}' IS NOT NULL;
    IF n < 1 THEN
        RAISE EXCEPTION '219b: content-reviewer has no validate_content step — re-derive';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             '{workflow,steps,validate_content,config,require_sections_metadata}',
             'true'::jsonb, true),
           updated_at = now()
     WHERE type='content-reviewer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    -- Assert the whole intended end state, not just this write: every active
    -- caller that validates over the page_content path declares it, and the tool
    -- caller does not.
    SELECT count(*) INTO n
    FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') e(k,v)
    WHERE COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL AND ad.is_active
      AND v->>'action' = 'validate_page_content'
      AND COALESCE(v->'config'->>'html_field','page_content.response.page_html') LIKE 'page_content.%'
      AND COALESCE(v->'config'->>'require_sections_metadata','false') <> 'true';
    IF n <> 0 THEN
        RAISE EXCEPTION '219b: % page_content-path caller(s) still undeclared — the silent-skip gap persists there', n;
    END IF;

    RAISE NOTICE '219b: every active page_content-path caller of validate_page_content now declares require_sections_metadata';
END $d$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('219b_content_reviewer_declares_sections_metadata.sql',
        'Council round 5 (569241fb, medium): dormancy is not closure. content-reviewer uses the same validate_page_content action over the same page_content path as page-build-handler and was left undeclared because it has zero runs; if re-enabled it would inherit the silent-skip ambiguity with nothing forcing the declaration. Now declared. tool-recreation-handler stays undeclared deliberately (html_field is a tool blob, no sections). The post-condition asserts the whole end state, so a future caller added without the declaration fails this check on replay.')
ON CONFLICT (filename) DO NOTHING;

COMMIT;
