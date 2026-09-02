-- 682_content_listing_collapse_empty_card_slots.sql
--
-- WHAT THIS FIXES (bugs_open/425, boxingonline.com, reported by the owner
-- 2026-09-02 as "the cards need better designs").
--
-- The `content-listing` component renders six article cards. Four of each
-- card's six content slots — category, excerpt, date, read_time — were rendered
-- with NO guard, so a key the producer never wrote became an EMPTY ELEMENT that
-- still occupied layout. The card read as image + one long headline + a run of
-- blank boxes. That is a data gap wearing the costume of a design fault, and
-- the design fault is what got reported.
--
-- This migration is the half that makes the component able to tell a missing
-- input from an intentional blank. The other half — the producer that was never
-- writing `excerpt` at all — is the Go change in the same commit
-- (queryresolve.ListItemTitle / ListItemExcerpt).
--
-- THE SECTION HEADER IS GUARDED THE SAME WAY, and it was found by the render
-- proof rather than by the report: both its children were already guarded, so
-- when a section carries neither a title nor a subtitle the WRAPPER still
-- rendered, empty, with its own margin. That is the reported defect one level
-- up, in the same component, and fixing only the cards would have left it.
--
-- WHY THE META ROW IS GUARDED AS A PAIR: `article-card__meta` is a flex row
-- with its own margin. Guarding only its two children leaves an empty flex
-- container still claiming vertical space, which is the same defect one level
-- up. `{{if or .date .read_time}}` collapses the row itself.
--
-- SCOPE. [MEASURED 2026-09-02] 13 live page_components across 6 sites use this
-- component (homegarden.uk 6, dartsonline.com 2, garden-tools.uk 2,
-- boxingonline.com 1, idea.uk 1, robot-hands.com 1). Every one of them gains a
-- deck and loses four empty elements. No slot that HAS a value renders
-- differently — this migration can only ever remove empty markup.
--
-- DELIBERATELY NOT IN SCOPE: the array-level empty state. `content-listing`
-- declares `on_missing: "skip_section"` for `articles`, and bugs_closed/054
-- fix-candidate 2 (live on chassis v1.0.1149) is what honours it, so an empty
-- listing drops the section before the template runs. Adding 054's
-- {{if .articles}}…{{else}}…{{end}} shape here would need a translatable
-- `empty_state_text` schema field for a branch that cannot be reached.
--
-- Config change: live immediately, no image roll. Reversible — see
-- 682_..._ROLLBACK.sql.

BEGIN;

-- DRIFT GUARD. Another session may have edited this template between the read
-- that produced the expected value and this apply. Abort rather than clobber.
-- A guard that checks DRIFT, not order (see CLAUDE.md's migration-runner note).
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
      FROM content_components
     WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
       AND name = 'content-listing'
       AND is_active
       AND html_template LIKE '%<span class="article-card__category">{{.category}}</span>%'
       AND html_template LIKE '%<p class="article-card__excerpt">{{.excerpt}}</p>%';

    IF n <> 1 THEN
        RAISE EXCEPTION
            'content-listing is not in the state this migration was written against '
            '(matched % rows, expected 1). Another session has edited the template, or '
            'this migration has already applied. Re-read it before re-running.', n;
    END IF;
END $$;

UPDATE content_components
   SET html_template = $tmpl$<section class="section section--articles">
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
    </section>$tmpl$,
       updated_at = now()
 WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';

-- VERIFY. A DO/RAISE, not a SELECT: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT (CLAUDE.md / RFC_006).
DO $$
DECLARE
    t text;
    unguarded text[] := '{}';
BEGIN
    SELECT html_template INTO t
      FROM content_components
     WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';

    -- Each test asserts the GUARDED form is present, never that the bare form is
    -- absent. The bare span is a SUBSTRING of the guarded one
    -- ({{if .category}}<span …>{{.category}}</span>{{end}}), so an "is the bare
    -- form still here?" test is true either way and reports every correct
    -- migration as a failure. It did, on the first run of this file — which is
    -- what the DO/RAISE is for: it aborted the transaction and the template was
    -- left untouched, where a block of SELECTs would have committed silently.
    IF t NOT LIKE '%{{if .category}}<span class="article-card__category">{{.category}}</span>{{end}}%' THEN
        unguarded := unguarded || 'category'::text;
    END IF;
    IF t NOT LIKE '%{{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}%' THEN
        unguarded := unguarded || 'excerpt'::text;
    END IF;
    IF t NOT LIKE '%{{if .date}}<span class="article-card__date">{{.date}}</span>{{end}}%' THEN
        unguarded := unguarded || 'date'::text;
    END IF;
    IF t NOT LIKE '%{{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}%' THEN
        unguarded := unguarded || 'read_time'::text;
    END IF;
    IF t NOT LIKE '%{{if or .date .read_time}}<div class="article-card__meta">%' THEN
        unguarded := unguarded || 'meta-row'::text;
    END IF;
    IF t NOT LIKE '%{{if or .section_title .section_subtitle}}<div class="section__header">%' THEN
        unguarded := unguarded || 'section-header'::text;
    END IF;

    IF array_length(unguarded, 1) IS NOT NULL THEN
        RAISE EXCEPTION 'content-listing still renders unguarded per-item slots: %',
            array_to_string(unguarded, ', ');
    END IF;

    -- The positive control: the slots must still be PRESENT, guarded. A
    -- migration that deleted them would pass every check above.
    IF t NOT LIKE '%article-card__excerpt%'
       OR t NOT LIKE '%article-card__category%'
       OR t NOT LIKE '%article-card__read-time%'
       OR t NOT LIKE '%article-card__date%' THEN
        RAISE EXCEPTION 'content-listing lost a card slot entirely — guarding was meant to '
                        'collapse empty slots, not remove them';
    END IF;

    RAISE NOTICE 'content-listing: four per-item slots guarded; meta row and section header each collapse as a pair';
END $$;

COMMIT;
