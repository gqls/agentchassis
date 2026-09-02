-- 686_article_body_hero_image_capability.sql
--
-- WHAT: give `article-body` the ability to display an image. It gains one
-- optional image field (`hero_image_url`, sourced from the existing
-- `site_assets.hero` role) plus its alt text, and a template block that renders
-- them ONLY when the URL resolves.
--
-- WHY: measured 2026-09-02 at boxingonline.ugg2.com (first paid site, second
-- owner review). Six `content_hero` images exist, one per article, all deployed
-- and serving HTTP 200 — and not one of the six article pages references its own
-- hero, as <img> or as a CSS background. Every /blog/ page is exactly ONE
-- `article-body` row, and `article-body` has a single field (`content`) and a
-- template whose only interpolation is `{{.content}}`: no <img>, no <figure>, no
-- background-image. It cannot display an image BY CONSTRUCTION, whatever the
-- planner requests and whatever assets exist. This is a consumption gap, not a
-- generation gap — the pictures were already made.
--
-- Its own writer guidance already says imagery "belong[s] to the component
-- system, not inside this text". This migration is the half of that contract
-- that was never built.
--
-- NOTHING NEW IS INVENTED. Every mechanism this leans on already exists and is
-- exercised by live components:
--   * `plan_sections_action.go:463-476` already looks up
--     `imageryplan.ContentHeroKey(pageName)` in `assets` and binds it to
--     `r.assets["hero"]` — its own comment says this is "what makes the article
--     page show the same image family as its listing card", and notes the
--     "convention with no plan row". So NO site_plan_imagery row is required.
--   * `site_assets.hero` as a field source is precedented by the live `hero`
--     component (`background_image`).
--   * optional-image-plus-alt is precedented by `content-block-about`
--     (`image_src` = site_assets.image + `image_alt` = llm, both required:false).
--   * `rerender_page_sections_action.go:464` builds the resolver WITH the page
--     name, so existing pages pick the field up on their next rerender; no
--     re-plan and no backfill are needed.
--
-- BLAST RADIUS: 297 `article-body` instances across 30 sites [MEASURED
-- 2026-09-02]. Every one of them renders UNCHANGED except for inert CSS — see
-- the equivalence note below.
--
-- EQUIVALENCE, stated precisely rather than as "byte-identical": an instance
-- with no resolvable hero renders markup that is byte-for-byte what it renders
-- today, because the entire new block sits inside `{{if .hero_image_url}}`. The
-- ONLY difference in the emitted bytes is 176 characters of added CSS inside the
-- existing <style> block, whose two selectors match no element that such a page
-- contains. Proven locally against the real stored content_data of a live
-- article before submission (three renders: no-hero equivalence, hero present,
-- hero present with alt absent).
--
-- ALT TEXT IS GUARDED, deliberately: `alt="{{if .hero_image_alt}}...{{end}}"`
-- rather than a bare `{{.hero_image_alt}}`. A bare reference to a key that is
-- absent from content_data emits the literal `<no value>` into the attribute,
-- and an absent optional field is the NORMAL case here (the writer has never
-- been asked for this field on any existing page). `alt=""` is the correct
-- fallback for an image whose meaning the adjacent heading already carries.
--
-- SAFETY: guarded on the template's exact md5 so a concurrent edit by another
-- session aborts this rather than silently stacking on top of it, and idempotent
-- (re-running after success is a no-op, not a double insert). Verification uses
-- DO/RAISE, not SELECTs — a verify block of SELECTs cannot stop the COMMIT.

BEGIN;

DO $$
DECLARE
    v_id          uuid;
    v_md5         text;
    v_len         int;
    v_already     boolean;
    v_markup_anchors int;
    v_style_anchors  int;
    k_expected_md5 constant text := '002cbcd9cada6a37bf4a5158fd1e5f22';
BEGIN
    SELECT id, md5(html_template), length(html_template),
           html_template LIKE '%article-body__hero%',
           (length(html_template) - length(replace(html_template, '<div class="container">', ''))) / length('<div class="container">'),
           (length(html_template) - length(replace(html_template, '</style>', ''))) / length('</style>')
      INTO v_id, v_md5, v_len, v_already, v_markup_anchors, v_style_anchors
      FROM content_components
     WHERE name = 'article-body' AND is_active;

    IF v_id IS NULL THEN
        RAISE EXCEPTION '686: no active article-body component found — refusing';
    END IF;

    IF v_already THEN
        RAISE NOTICE '686: article-body already carries the hero block — nothing to do (idempotent no-op)';
        RETURN;
    END IF;

    IF v_md5 <> k_expected_md5 THEN
        RAISE EXCEPTION '686: article-body template has changed under us (md5 % , length %, expected %). Another session has edited it. Re-derive the anchors against the CURRENT template before applying.',
            v_md5, v_len, k_expected_md5;
    END IF;

    -- Anchor uniqueness. replace() replaces EVERY occurrence, so a second
    -- occurrence would inject the block twice and no later check would say why.
    -- (Counted in the SELECT above: PL/pgSQL's IF takes an expression, not a
    -- FROM clause — the first cut of this file got that wrong and the
    -- BEGIN/ROLLBACK rehearsal is what caught it.)
    IF v_markup_anchors <> 1 THEN
        RAISE EXCEPTION '686: markup anchor <div class="container"> occurs % times, expected exactly 1 — refusing', v_markup_anchors;
    END IF;

    IF v_style_anchors <> 1 THEN
        RAISE EXCEPTION '686: style anchor </style> occurs % times, expected exactly 1 — refusing', v_style_anchors;
    END IF;

    UPDATE content_components
       SET html_template = replace(
                             replace(html_template,
                               '<div class="container">',
                               '<div class="container">{{if .hero_image_url}}<figure class="article-body__hero"><img src="{{.hero_image_url}}" alt="{{if .hero_image_alt}}{{.hero_image_alt}}{{end}}" loading="lazy" /></figure>{{end}}'),
                             '</style>',
                             '.article-body-section .article-body__hero{margin:0 0 2rem}.article-body-section .article-body__hero img{width:100%;height:auto;display:block;border-radius:var(--radius-md,8px)}</style>'),
           input_schema = jsonb_set(
                            jsonb_set(input_schema,
                              '{fields,hero_image_url}',
                              jsonb_build_object(
                                'type', 'url',
                                'source', 'site_assets.hero',
                                'required', false,
                                'on_missing', 'skip_field'),
                              true),
                            '{fields,hero_image_alt}',
                            jsonb_build_object(
                              'type', 'text',
                              'source', 'llm',
                              'required', false,
                              'on_missing', 'skip_field',
                              'llm_guidance', 'One short factual sentence describing what the article''s header image shows, for screen-reader users. Describe only what is visibly in the image. Leave it out entirely if you have not been shown the image.'),
                            true),
           updated_at = now()
     WHERE id = v_id;

    RAISE NOTICE '686: article-body updated (id %)', v_id;
END $$;

-- Post-conditions. DO/RAISE so a failure actually aborts the transaction.
DO $$
DECLARE
    t text;
    s jsonb;
BEGIN
    SELECT html_template, input_schema INTO t, s
      FROM content_components WHERE name = 'article-body' AND is_active;

    IF t NOT LIKE '%{{if .hero_image_url}}%' THEN
        RAISE EXCEPTION '686 VERIFY: guarded hero block absent after update';
    END IF;
    IF (length(t) - length(replace(t, '{{.content}}', ''))) / length('{{.content}}') <> 1 THEN
        RAISE EXCEPTION '686 VERIFY: the content interpolation is no longer present exactly once — the article body itself is at risk';
    END IF;
    IF t LIKE '%<no value>%' THEN
        RAISE EXCEPTION '686 VERIFY: literal <no value> present in template';
    END IF;
    IF s->'fields'->'hero_image_url'->>'source' <> 'site_assets.hero' THEN
        RAISE EXCEPTION '686 VERIFY: hero_image_url source wrong or missing (got %)', s->'fields'->'hero_image_url'->>'source';
    END IF;
    IF s->'fields'->'hero_image_alt'->>'source' <> 'llm' THEN
        RAISE EXCEPTION '686 VERIFY: hero_image_alt source wrong or missing';
    END IF;
    IF (s->'fields'->'hero_image_url'->>'required')::boolean IS DISTINCT FROM false
       OR (s->'fields'->'hero_image_alt'->>'required')::boolean IS DISTINCT FROM false THEN
        RAISE EXCEPTION '686 VERIFY: a new field is required=true — every existing instance would become unsatisfiable';
    END IF;
    IF s->'fields'->'content'->>'source' <> 'llm' THEN
        RAISE EXCEPTION '686 VERIFY: the existing content field was disturbed';
    END IF;

    RAISE NOTICE '686 VERIFY: OK — guarded block present, content intact, both new fields optional';
END $$;

COMMIT;
