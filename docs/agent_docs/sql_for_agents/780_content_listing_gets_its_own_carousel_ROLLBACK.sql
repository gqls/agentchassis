-- 780_content_listing_gets_its_own_carousel_ROLLBACK.sql
--
-- Undoes all four parts of 780: the template fork, the schema field, the applies_to entry,
-- and the boxingonline instance flag. Restores the pre-780 template byte-for-byte.
--
-- ⚠ WHAT THIS DOES NOT UNDO: any page already RENDERED as a carousel keeps its stored HTML
-- until it re-renders. This restores the SOURCE, not the artefacts built from it.

BEGIN;

DO $mig$
DECLARE n int; comp_id uuid;
BEGIN
    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name='content-listing' AND input_schema->'fields' ? 'carousel';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: 780 is not applied (content-listing declares no carousel field)';
    END IF;

    SELECT id INTO comp_id FROM content_components WHERE is_active AND name='content-listing';

    UPDATE content_components
       SET input_schema = input_schema #- '{fields,carousel}',
           html_template = $ct$<section class="section section--articles">
      <div class="container">
        {{if or .section_title .section_subtitle}}<div class="section__header">
          {{if .section_title}}<h2 class="section__title">{{.section_title}}</h2>{{end}}
          {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
        </div>{{end}}
        <div class="article-grid grid grid--3">
          {{range .articles}}
          <article class="article-card hover-lift">
            {{if .image}}<div class="article-card__image">
              <img src="{{.image}}" alt="{{.title}}" loading="lazy">
              {{if .category}}<span class="article-card__category">{{.category}}</span>{{end}}
            </div>{{end}}
            <div class="article-card__content">
              <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
              {{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}
              {{if or .date .read_time}}<div class="article-card__meta">
                {{if .date}}<span class="article-card__date">{{.date}}</span>{{end}}
                {{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}
              </div>{{end}}
            </div>
          </article>
          {{end}}
        </div>
        {{if .show_load_more}}
        <div class="section__actions">
          <button class="button button--secondary">{{.load_more_text}}</button>
        </div>
        {{end}}
      </div>
    </section>$ct$,
           updated_at = now()
     WHERE id = comp_id;

    UPDATE js_snippets
       SET applies_to = (SELECT jsonb_agg(e) FROM jsonb_array_elements(applies_to) e
                          WHERE e <> '"content-listing"'::jsonb)
     WHERE is_active AND name='hero-card-carousel';
END $mig$;

UPDATE page_components pc
   SET content_data = pc.content_data - 'carousel',
       updated_at = now()
  FROM content_components cc
 WHERE cc.id = pc.component_id AND cc.name = 'content-listing'
   AND pc.content_data ? 'carousel';

DO $$
DECLARE n int; tpl text;
BEGIN
    SELECT html_template INTO tpl FROM content_components WHERE is_active AND name='content-listing';
    IF position('data-hcc' in tpl) > 0 THEN
        RAISE EXCEPTION 'ABORT: carousel markup survives in the restored template';
    END IF;
    IF length(tpl) <> 1526 THEN
        RAISE EXCEPTION 'ABORT: restored template is % chars, expected the pre-780 1526', length(tpl);
    END IF;

    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name='content-listing' AND input_schema->'fields' ? 'carousel';
    IF n <> 0 THEN RAISE EXCEPTION 'ABORT: the carousel field survives'; END IF;

    SELECT count(*) INTO n FROM js_snippets
     WHERE is_active AND name='hero-card-carousel' AND applies_to @> '["content-listing"]'::jsonb;
    IF n <> 0 THEN RAISE EXCEPTION 'ABORT: applies_to still names content-listing'; END IF;

    -- The ORIGINAL two entries must survive — a bad jsonb_agg could empty the array.
    SELECT count(*) INTO n FROM js_snippets
     WHERE is_active AND name='hero-card-carousel'
       AND applies_to @> '["hero-card-carousel","info-card-grid"]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: applies_to lost its original entries — the filter removed too much';
    END IF;

    SELECT count(*) INTO n FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
     WHERE cc.name='content-listing' AND pc.content_data ? 'carousel';
    IF n <> 0 THEN RAISE EXCEPTION 'ABORT: % instances still carry the carousel key', n; END IF;

    RAISE NOTICE '780 ROLLBACK: template restored to 1526 chars, field dropped, applies_to '
                 'restored to its original two entries, instance flags cleared';
END $$;

COMMIT;
