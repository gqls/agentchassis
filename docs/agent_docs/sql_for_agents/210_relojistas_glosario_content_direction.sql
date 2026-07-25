-- 210_relojistas_glosario_content_direction.sql
--
-- bugs_open/028 (page_build_noop_reports_complete_and_deploys_borrowed_components),
-- residual "candidate 3" — the borrowed hero on relojistas.com glossary pages.
--
-- WHAT THE MECHANISM ACTUALLY IS (corrects the original filing):
--   There is no "borrowing fallback". Nothing in save/deploy copies a sibling or
--   site-level component. Each hero is GENERATED fresh by page-content-writer.
--   Its only page-subject signal is `{{.current_page.title}}`, and on these pages
--   that title is a single bare term ("Tourbillon", "Calibre"). With no
--   content_direction, no meta_description and an empty content_brief, the model
--   has essentially no page subject and writes a site-level hero paraphrased from
--   the site's value_proposition. 6 of 8 glossary pages fell that way; 2 did not.
--   It is non-deterministic because nothing in the prompt REQUIRES the subject.
--
--   Fleet survey (2026-07-25): every other site's leaf-page heroes are on-topic —
--   their titles are descriptive sentences, so the thin signal is still enough.
--   relojistas' glosario is the only bare-title page set in the fleet, which is
--   why the damage is contained to it.
--
-- THE FIX: give these pages a real per-page subject via `pages.content_direction`
--   — the lever built and proven for exactly this by bugs_closed/025 (wired
--   v1.0.1146; reaches the writer as `.current_page.content_direction`, consumed
--   by the "## Page-Specific Content Direction (for THIS page - follow closely)"
--   block of page-content-writer's prompt). No code change, no fleet blast
--   radius, existing machinery.
--
-- Applied to all 8 glossary entity-pages, not just the 6 broken ones: the two
-- that came out correct did so by luck, and the steering is what stops them
-- regressing on any future rebuild.
--
-- ROLLBACK:
--   UPDATE pages p SET content_direction = NULL
--   FROM sites s
--   WHERE s.id = p.site_id AND s.domain = 'relojistas.com'
--     AND p.page_type = 'entity-page' AND p.name LIKE 'glosario-%';

BEGIN;

-- Precondition guard. All 8 are NULL as of filing; if a concurrent session has
-- set one meanwhile, fail loudly rather than clobber their steering.
DO $$
DECLARE
    n_total   int;
    n_notnull int;
BEGIN
    SELECT count(*), count(*) FILTER (WHERE p.content_direction IS NOT NULL)
      INTO n_total, n_notnull
      FROM pages p JOIN sites s ON s.id = p.site_id
     WHERE s.domain = 'relojistas.com'
       AND p.page_type = 'entity-page'
       AND p.name LIKE 'glosario-%';

    IF n_total <> 8 THEN
        RAISE EXCEPTION '210: expected 8 glosario entity-pages, found %', n_total;
    END IF;
    IF n_notnull <> 0 THEN
        RAISE EXCEPTION '210: % of % glosario pages already carry content_direction — another session may have set it; inspect before re-running', n_notnull, n_total;
    END IF;
END $$;

UPDATE pages p
SET content_direction = jsonb_build_object(
        'instruction', format(
            'Esta página del glosario define el término «%s». Toda la copia de la página debe tratar exclusivamente de ese término: qué es, cómo funciona y qué implica para quien usa un reloj. El titular del hero debe nombrar el término.',
            p.title),
        'format', 'Titular de una sola frase que nombre el término y diga qué es.',
        'examples', jsonb_build_array(
            format('%s: qué es y por qué importa en un reloj', p.title)),
        'avoid', jsonb_build_array(
            'Titulares sobre el sitio o el portal en general, por ejemplo «Relojería en español: noticias, guías y glosario» o «Relojería en español, sin rodeos».',
            'Repetir la propuesta de valor de Relojistas en lugar de explicar el término.',
            'Cualquier titular que no mencione el término de esta página.')
    ),
    updated_at = now()
FROM sites s
WHERE s.id = p.site_id
  AND s.domain = 'relojistas.com'
  AND p.page_type = 'entity-page'
  AND p.name LIKE 'glosario-%';

-- Post-condition guard: all 8 steered, and each instruction names its own term.
DO $$
DECLARE
    n_set     int;
    n_named   int;
BEGIN
    SELECT count(*) FILTER (WHERE p.content_direction IS NOT NULL),
           count(*) FILTER (WHERE p.content_direction->>'instruction' LIKE '%' || p.title || '%')
      INTO n_set, n_named
      FROM pages p JOIN sites s ON s.id = p.site_id
     WHERE s.domain = 'relojistas.com'
       AND p.page_type = 'entity-page'
       AND p.name LIKE 'glosario-%';

    IF n_set <> 8 THEN
        RAISE EXCEPTION '210: expected 8 steered pages, got %', n_set;
    END IF;
    IF n_named <> 8 THEN
        RAISE EXCEPTION '210: expected 8 instructions naming their own term, got %', n_named;
    END IF;
END $$;

COMMIT;
