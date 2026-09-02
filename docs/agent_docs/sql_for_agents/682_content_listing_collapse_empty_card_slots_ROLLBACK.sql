-- 682_content_listing_collapse_empty_card_slots_ROLLBACK.sql
--
-- Restores content-listing's html_template to its pre-682 state: the four
-- per-item card slots rendered UNGUARDED.
--
-- Only run this if guarding the slots caused a problem. Note what it restores:
-- every card whose producer does not write category/excerpt/date/read_time goes
-- back to rendering four empty elements that still occupy layout — the defect
-- bugs_open/425 reports. It cannot restore data; it only un-hides the gap.

BEGIN;

DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
      FROM content_components
     WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
       AND html_template LIKE '%{{if or .date .read_time}}%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'content-listing does not carry the 682 template (matched % rows) — '
                        'nothing to roll back, or another session has edited it since', n;
    END IF;
END $$;

UPDATE content_components
   SET html_template = $tmpl$<section class="section section--articles">
      <div class="container">
        <div class="section__header">
          {{if .section_title}}<h2 class="section__title">{{.section_title}}</h2>{{end}}
          {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
        </div>
        <div class="article-grid grid grid--3">
          {{range .articles}}
          <article class="article-card hover-lift">
            {{if .image}}<div class="article-card__image">
              <img src="{{.image}}" alt="{{.title}}" loading="lazy">
              <span class="article-card__category">{{.category}}</span>
            </div>{{end}}
            <div class="article-card__content">
              <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
              <p class="article-card__excerpt">{{.excerpt}}</p>
              <div class="article-card__meta">
                <span class="article-card__date">{{.date}}</span>
                <span class="article-card__read-time">{{.read_time}}</span>
              </div>
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
    </section>$tmpl$,
       updated_at = now()
 WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';

COMMIT;
