-- 686_article_body_hero_image_capability_ROLLBACK.sql
--
-- Reverses 686: removes the guarded hero block, its CSS, and the two optional
-- fields, returning article-body to md5 002cbcd9cada6a37bf4a5158fd1e5f22.
--
-- Guarded and idempotent. Refuses if the template is not in the state 686 left
-- it in — if a THIRD change has landed since, reversing by string replacement
-- would silently keep that third change's bytes while claiming to have rolled
-- back, so it aborts and asks for a hand instead.
--
-- NOTE ON SCOPE: this reverts the CAPABILITY. It does not remove
-- `hero_image_url` values already resolved into page_components.content_data by
-- a rerender; those become inert data the template no longer reads, which is
-- harmless and is what makes re-applying 686 cheap.

BEGIN;

DO $$
DECLARE
    v_id uuid;
    t    text;
BEGIN
    SELECT id, html_template INTO v_id, t
      FROM content_components WHERE name = 'article-body' AND is_active;

    IF v_id IS NULL THEN
        RAISE EXCEPTION '686 ROLLBACK: no active article-body component found';
    END IF;

    IF t NOT LIKE '%article-body__hero%' THEN
        RAISE NOTICE '686 ROLLBACK: hero block already absent — nothing to do (idempotent no-op)';
        RETURN;
    END IF;

    UPDATE content_components
       SET html_template = replace(
                             replace(html_template,
                               '{{if .hero_image_url}}<figure class="article-body__hero"><img src="{{.hero_image_url}}" alt="{{if .hero_image_alt}}{{.hero_image_alt}}{{end}}" loading="lazy" /></figure>{{end}}',
                               ''),
                             '.article-body-section .article-body__hero{margin:0 0 2rem}.article-body-section .article-body__hero img{width:100%;height:auto;display:block;border-radius:var(--radius-md,8px)}',
                             ''),
           input_schema = (input_schema #- '{fields,hero_image_url}') #- '{fields,hero_image_alt}',
           updated_at = now()
     WHERE id = v_id;
END $$;

DO $$
DECLARE
    v_md5 text;
    s     jsonb;
    k_original_md5 constant text := '002cbcd9cada6a37bf4a5158fd1e5f22';
BEGIN
    SELECT md5(html_template), input_schema INTO v_md5, s
      FROM content_components WHERE name = 'article-body' AND is_active;

    IF v_md5 <> k_original_md5 THEN
        RAISE EXCEPTION '686 ROLLBACK: template did not return to its pre-686 bytes (md5 %, expected %). A later change is present; reverse it by hand rather than by string replacement.',
            v_md5, k_original_md5;
    END IF;
    IF s->'fields' ? 'hero_image_url' OR s->'fields' ? 'hero_image_alt' THEN
        RAISE EXCEPTION '686 ROLLBACK: a hero field survived the schema removal';
    END IF;
    IF s->'fields'->'content'->>'source' <> 'llm' THEN
        RAISE EXCEPTION '686 ROLLBACK: the content field was disturbed';
    END IF;

    RAISE NOTICE '686 ROLLBACK: OK — template back to pre-686 md5, both fields gone, content intact';
END $$;

COMMIT;
