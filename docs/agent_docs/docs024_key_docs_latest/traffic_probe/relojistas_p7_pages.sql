-- ============================================================================
-- relojistas.com — P7 content pages: 4 Guías + 8 Glosario terms
-- Written 2026-07-19. Site ecf15e75-a966-4900-bcb0-1c85f689dbfd,
-- current plan f12ab433-f7c4-4209-8c21-dcfdaed43078.
--
-- APPLY AFTER relojistas_evidence_base.sql (the cite-or-omit fence must exist
-- BEFORE any page is written, or the writer generates unfenced and we would be
-- reviewing fabrications instead of preventing them).
--
-- ---------------------------------------------------------------------------
-- WHY THESE PAGE TYPES (decided by reading working sites, not by preference)
--
--   Listings are a COMPONENT behaviour: a component declares
--   "source": "query.pages_where_type:<t>" and PlanSectionsAction resolves it.
--   Only two library components list pages by type in a way we can use:
--     guide-list      → query.pages_where_type:GUIDE        (limit 12)
--     archetype-grid  → query.pages_where_type:ENTITY-PAGE  (limit 8)
--
--   So guides MUST be page_type='guide' (NOT 'blog-post' — guide-list would
--   not see a blog-post) and glossary terms MUST be 'entity-page'.
--
--   Working model copied wholesale from gamesdesign.co.uk, which already runs
--   this exact shape in production:
--     guides-index : section-index, sections ["hero","guide-list"]
--     guide pages  : page_type 'guide', /guides/<slug>/index.html,
--                    sections ["hero","generic-text-block"]
--   and from vonc.com for the entity side:
--     entity pages : page_type 'entity-page', sections
--                    ["hero","content-block-about","call-to-action"]
--
-- ---------------------------------------------------------------------------
-- WHY 8 GLOSSARY TERMS AND NOT THE ~12 AGREED
--
--   archetype-grid's items field carries "limit": 8. PlanSectionsAction reads
--   the field's own limit, so a 9th term would be authored, published and then
--   silently NOT listed on the Glosario index — unreachable except by URL.
--
--   Rejected alternatives, recorded so nobody re-litigates them blind:
--     - raise the limit in place → archetype-grid is a SHARED library component
--       with no site_id; vonc.com's archetypes page renders from the same row.
--       Editing it to suit relojistas changes another live site.
--     - fork it to a glossary-grid (content_components.forked_from exists for
--       this) → viable, and the right move if the glossary is meant to grow.
--       Not done here: it adds an untested template + CSS surface for four
--       extra terms, and "8 well-sourced terms" satisfies the agreed
--       "small and real" better than 12 terms with a broken index does.
--   If the glossary should grow past 8, fork the component FIRST, then add
--   terms — not the other way round.
--
-- ---------------------------------------------------------------------------
-- REPURPOSE, NOT DELETE
--   `articulo` and `glosario-entrada` are the two stray planner pages (see the
--   correction in the running notes — they are NOT templates). Rather than
--   DELETE against a live plan, both are renamed in place into real content.
--   articulo → guia-mantenimiento deliberately: /guias/mantenimiento is one of
--   the three phantom homepage links, so it becomes a real destination instead
--   of being deleted.
--
--   Plan row and page row MUST move together — reconcile_site_plan diffs
--   plan-vs-realised on `name`, so renaming one and not the other orphans both.
-- ============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Give the two index pages their sections.
--    This is the whole reason they were stuck: sections=[] makes
--    PlanSectionsAction return "no sections to plan" and the handler routes the
--    page to needs_human_review. Same defect, same fix, as /noticias.
-- ---------------------------------------------------------------------------
UPDATE pages SET sections = '["hero","guide-list"]'::jsonb, updated_at = now()
 WHERE site_id = 'ecf15e75-a966-4900-bcb0-1c85f689dbfd' AND name = 'guias-index';

UPDATE pages SET sections = '["hero","archetype-grid"]'::jsonb, updated_at = now()
 WHERE site_id = 'ecf15e75-a966-4900-bcb0-1c85f689dbfd' AND name = 'glosario-index';

-- ---------------------------------------------------------------------------
-- 2. Repurpose the two stray pages (plan row + page row, in lockstep).
-- ---------------------------------------------------------------------------
UPDATE site_plan_pages SET
    name = 'guia-mantenimiento', role = 'guide', slug = 'mantenimiento',
    url = '/guias/mantenimiento/index.html', parent_section = 'guias',
    title = 'Cuidado y mantenimiento de tu reloj',
    nav_label = 'Mantenimiento', nav_order = 20,
    meta_description = 'Cómo cuidar un reloj mecánico en el día a día: manipulación, agua, imanes y cuándo acudir a un servicio oficial.'
 WHERE plan_id = 'f12ab433-f7c4-4209-8c21-dcfdaed43078' AND name = 'articulo';

UPDATE pages SET
    name = 'guia-mantenimiento', url = '/guias/mantenimiento/index.html',
    page_type = 'guide', title = 'Cuidado y mantenimiento de tu reloj',
    nav_label = 'Mantenimiento', nav_order = 20,
    in_header = false, in_footer = false,
    sections = '["hero","generic-text-block"]'::jsonb,
    meta_description = 'Cómo cuidar un reloj mecánico en el día a día: manipulación, agua, imanes y cuándo acudir a un servicio oficial.',
    build_status = 'planned', built_from_plan_version = NULL, updated_at = now()
 WHERE site_id = 'ecf15e75-a966-4900-bcb0-1c85f689dbfd' AND name = 'articulo';

UPDATE site_plan_pages SET
    name = 'glosario-tourbillon', role = 'entity-page', slug = 'tourbillon',
    url = '/glosario/tourbillon.html', parent_section = 'glosario',
    title = 'Tourbillon', nav_label = 'Tourbillon', nav_order = 40,
    meta_description = 'Qué es un tourbillon: el escape alojado en una jaula giratoria, y por qué se asocia a la alta relojería.'
 WHERE plan_id = 'f12ab433-f7c4-4209-8c21-dcfdaed43078' AND name = 'glosario-entrada';

UPDATE pages SET
    name = 'glosario-tourbillon', url = '/glosario/tourbillon.html',
    page_type = 'entity-page', title = 'Tourbillon',
    nav_label = 'Tourbillon', nav_order = 40,
    in_header = false, in_footer = false,
    sections = '["hero","content-block-about"]'::jsonb,
    meta_description = 'Qué es un tourbillon: el escape alojado en una jaula giratoria, y por qué se asocia a la alta relojería.',
    build_status = 'planned', built_from_plan_version = NULL, updated_at = now()
 WHERE site_id = 'ecf15e75-a966-4900-bcb0-1c85f689dbfd' AND name = 'glosario-entrada';

-- ---------------------------------------------------------------------------
-- 3. Three more Guías.
-- ---------------------------------------------------------------------------
INSERT INTO site_plan_pages (plan_id, name, role, slug, url, parent_section, in_header, in_footer, nav_order, title, nav_label, meta_description)
VALUES
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','guia-relojes-de-buceo','guide','relojes-de-buceo','/guias/relojes-de-buceo/index.html','guias',false,false,21,
  'Cómo leer las especificaciones de un reloj de buceo','Relojes de buceo',
  'Qué significan las cifras de un reloj de buceo y cómo interpretarlas sin generalizar de un modelo a todos.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','guia-complicaciones','guide','complicaciones','/guias/complicaciones/index.html','guias',false,false,22,
  'Complicaciones: qué son y cuáles importan','Complicaciones',
  'Cronógrafo, reserva de marcha, tourbillon y horas saltantes explicados con ejemplos reales de la actualidad relojera.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','guia-ediciones-limitadas','guide','ediciones-limitadas','/guias/ediciones-limitadas/index.html','guias',false,false,23,
  'Ediciones limitadas y coleccionismo','Ediciones limitadas',
  'Qué hace especial a una edición limitada y qué conviene mirar antes de interesarse por una.')
ON CONFLICT (plan_id, name) DO NOTHING;

INSERT INTO pages (site_id, name, url, title, page_type, status, nav_label, nav_order, in_header, in_footer, build_status, sections, meta_description)
VALUES
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','guia-relojes-de-buceo','/guias/relojes-de-buceo/index.html','Cómo leer las especificaciones de un reloj de buceo','guide','active','Relojes de buceo',21,false,false,'planned','["hero","generic-text-block"]'::jsonb,
  'Qué significan las cifras de un reloj de buceo y cómo interpretarlas sin generalizar de un modelo a todos.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','guia-complicaciones','/guias/complicaciones/index.html','Complicaciones: qué son y cuáles importan','guide','active','Complicaciones',22,false,false,'planned','["hero","generic-text-block"]'::jsonb,
  'Cronógrafo, reserva de marcha, tourbillon y horas saltantes explicados con ejemplos reales de la actualidad relojera.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','guia-ediciones-limitadas','/guias/ediciones-limitadas/index.html','Ediciones limitadas y coleccionismo','guide','active','Ediciones limitadas',23,false,false,'planned','["hero","generic-text-block"]'::jsonb,
  'Qué hace especial a una edición limitada y qué conviene mirar antes de interesarse por una.')
ON CONFLICT (site_id, name) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 4. Seven more Glosario terms (8 total with tourbillon = archetype-grid's cap).
--    Every term is one the site's OWN corpus actually uses — we document the
--    vocabulary our readers meet, we do not invent a dictionary.
-- ---------------------------------------------------------------------------
INSERT INTO site_plan_pages (plan_id, name, role, slug, url, parent_section, in_header, in_footer, nav_order, title, nav_label, meta_description)
VALUES
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-cronografo','entity-page','cronografo','/glosario/cronografo.html','glosario',false,false,41,'Cronógrafo','Cronógrafo','Qué es un cronógrafo y en qué se diferencia de un cronómetro.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-reserva-de-marcha','entity-page','reserva-de-marcha','/glosario/reserva-de-marcha.html','glosario',false,false,42,'Reserva de marcha','Reserva de marcha','Cuánto tiempo sigue funcionando un reloj mecánico sin volver a darle cuerda.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-movimiento-automatico','entity-page','movimiento-automatico','/glosario/movimiento-automatico.html','glosario',false,false,43,'Movimiento automático','Movimiento automático','Cómo se da cuerda sola una máquina automática mediante el rotor.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-calibre','entity-page','calibre','/glosario/calibre.html','glosario',false,false,44,'Calibre','Calibre','Qué designa el calibre de un reloj y por qué se cita en las fichas técnicas.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-complicacion','entity-page','complicacion','/glosario/complicacion.html','glosario',false,false,45,'Complicación','Complicación','Cualquier función de un reloj más allá de dar la hora.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-hermeticidad','entity-page','hermeticidad','/glosario/hermeticidad.html','glosario',false,false,46,'Hermeticidad','Hermeticidad','Qué expresa la resistencia al agua de un reloj y por qué no equivale a poder bucear.'),
 ('f12ab433-f7c4-4209-8c21-dcfdaed43078','glosario-horas-saltantes','entity-page','horas-saltantes','/glosario/horas-saltantes.html','glosario',false,false,47,'Horas saltantes','Horas saltantes','Una indicación de la hora que salta en una ventana en lugar de usar agujas.')
ON CONFLICT (plan_id, name) DO NOTHING;

INSERT INTO pages (site_id, name, url, title, page_type, status, nav_label, nav_order, in_header, in_footer, build_status, sections, meta_description)
VALUES
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-cronografo','/glosario/cronografo.html','Cronógrafo','entity-page','active','Cronógrafo',41,false,false,'planned','["hero","content-block-about"]'::jsonb,'Qué es un cronógrafo y en qué se diferencia de un cronómetro.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-reserva-de-marcha','/glosario/reserva-de-marcha.html','Reserva de marcha','entity-page','active','Reserva de marcha',42,false,false,'planned','["hero","content-block-about"]'::jsonb,'Cuánto tiempo sigue funcionando un reloj mecánico sin volver a darle cuerda.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-movimiento-automatico','/glosario/movimiento-automatico.html','Movimiento automático','entity-page','active','Movimiento automático',43,false,false,'planned','["hero","content-block-about"]'::jsonb,'Cómo se da cuerda sola una máquina automática mediante el rotor.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-calibre','/glosario/calibre.html','Calibre','entity-page','active','Calibre',44,false,false,'planned','["hero","content-block-about"]'::jsonb,'Qué designa el calibre de un reloj y por qué se cita en las fichas técnicas.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-complicacion','/glosario/complicacion.html','Complicación','entity-page','active','Complicación',45,false,false,'planned','["hero","content-block-about"]'::jsonb,'Cualquier función de un reloj más allá de dar la hora.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-hermeticidad','/glosario/hermeticidad.html','Hermeticidad','entity-page','active','Hermeticidad',46,false,false,'planned','["hero","content-block-about"]'::jsonb,'Qué expresa la resistencia al agua de un reloj y por qué no equivale a poder bucear.'),
 ('ecf15e75-a966-4900-bcb0-1c85f689dbfd','glosario-horas-saltantes','/glosario/horas-saltantes.html','Horas saltantes','entity-page','active','Horas saltantes',47,false,false,'planned','["hero","content-block-about"]'::jsonb,'Una indicación de la hora que salta en una ventana en lugar de usar agujas.')
ON CONFLICT (site_id, name) DO NOTHING;

COMMIT;

-- ----------------------------------------------------------------------------
-- Verify:
--   SELECT name, url, page_type, build_status,
--          jsonb_array_length(sections) AS secs
--     FROM pages
--    WHERE site_id='ecf15e75-a966-4900-bcb0-1c85f689dbfd'
--    ORDER BY nav_order NULLS LAST, name;
--   -- expect 19 rows; guias-index and glosario-index now secs=2;
--   -- 4 page_type='guide'; 8 page_type='entity-page'; NO 'articulo',
--   -- NO 'glosario-entrada' (both renamed, not deleted).
--
-- Then re-queue the two stuck index build items and let the dispatch loop run.
-- ----------------------------------------------------------------------------
