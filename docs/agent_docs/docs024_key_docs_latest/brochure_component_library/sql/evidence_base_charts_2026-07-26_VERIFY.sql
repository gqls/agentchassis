\set ON_ERROR_STOP on
-- VERIFY the chart layer on fundamentallyai.com's evidence_base.
--
-- Every check below is a rule the evidence-chart component depends on and that
-- nothing else enforces. Each returns ZERO ROWS when healthy; any row is a
-- defect, named. Run after the seed, and again after anything edits the
-- evidence_base (refresh_evidence_base rewrites values in place).
--
-- Deliberately NOT written as a re-run of the seed's own logic: a check that
-- shares its logic with the fix cannot falsify it, only agree with it — the
-- mistake this workstream made twice on 2026-07-25 (landmine L2).

\set site_id '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'

\echo '=== 0. shape: how many charts, points, facts ==='
SELECT jsonb_array_length(data->'charts') AS charts,
       jsonb_array_length(data->'facts')  AS facts,
       (SELECT count(*) FROM jsonb_array_elements(data->'charts') c,
                             jsonb_array_elements(c->'points') p) AS points
  FROM site_specs
 WHERE site_id = :'site_id'::uuid AND aspect = 'evidence_base' AND is_current;

\echo '=== 1. DEFECT if any row: a point references a fact that does not exist ==='
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT c->>'id' AS chart, p->>'fact_id' AS dangling_fact_id
  FROM eb, jsonb_array_elements(eb.data->'charts') c,
           jsonb_array_elements(c->'points') p
 WHERE NOT EXISTS (SELECT 1 FROM jsonb_array_elements(eb.data->'facts') f
                    WHERE f->>'id' = p->>'fact_id');

\echo '=== 2. DEFECT if any row: max_fact_id references a fact that does not exist ==='
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT c->>'id' AS chart, c->>'max_fact_id' AS dangling_max_fact_id
  FROM eb, jsonb_array_elements(eb.data->'charts') c
 WHERE c ? 'max_fact_id'
   AND NOT EXISTS (SELECT 1 FROM jsonb_array_elements(eb.data->'facts') f
                    WHERE f->>'id' = c->>'max_fact_id');

\echo '=== 3. DEFECT if any row: a charted fact whose value is not a JSON number ==='
\echo '    (html/template CSS-filters a string under printf %.4f -> ZgotmplZ -> dead bar)'
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current),
charted AS (
  SELECT DISTINCT p->>'fact_id' AS fid FROM eb, jsonb_array_elements(eb.data->'charts') c,
                                            jsonb_array_elements(c->'points') p
  UNION
  SELECT c->>'max_fact_id' FROM eb, jsonb_array_elements(eb.data->'charts') c WHERE c ? 'max_fact_id')
SELECT f->>'id' AS fact, jsonb_typeof(f->'value') AS value_type
  FROM eb, jsonb_array_elements(eb.data->'facts') f
  JOIN charted ON charted.fid = f->>'id'
 WHERE jsonb_typeof(f->'value') IS DISTINCT FROM 'number';

\echo '=== 4. DEFECT if any row: a charted fact carrying tolerance gte ==='
\echo '    (those facts say "state a FLOOR, never the exact number" — a bar states it exactly)'
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current),
charted AS (
  SELECT DISTINCT p->>'fact_id' AS fid FROM eb, jsonb_array_elements(eb.data->'charts') c,
                                            jsonb_array_elements(c->'points') p)
SELECT f->>'id' AS fact, f->>'tolerance' AS tolerance
  FROM eb, jsonb_array_elements(eb.data->'facts') f
  JOIN charted ON charted.fid = f->>'id'
 WHERE f->>'tolerance' = 'gte';

\echo '=== 5. DEFECT if any row: a SQL-sourced fact carrying a hand-written display ==='
\echo '    (refresh_evidence_base rewrites value + verified_at but never display -> silent drift)'
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT f->>'id' AS fact, f->>'display' AS display
  FROM eb, jsonb_array_elements(eb.data->'facts') f
 WHERE f->'source' ? 'sql' AND f ? 'display';

\echo '=== 6. DEFECT if any row: a point label missing every one of its fact context_terms ==='
\echo '    (the claims gate reads a 70-char window and needs a context term inside it,'
\echo '     or it reports our own charted figure as an unregistered number)'
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT c->>'id' AS chart, p->>'fact_id' AS fact, p->>'label' AS label
  FROM eb, jsonb_array_elements(eb.data->'charts') c,
           jsonb_array_elements(c->'points') p
 WHERE EXISTS (SELECT 1 FROM jsonb_array_elements(eb.data->'facts') f
                WHERE f->>'id' = p->>'fact_id' AND jsonb_array_length(COALESCE(f->'context_terms','[]'::jsonb)) > 0)
   AND NOT EXISTS (
       SELECT 1 FROM jsonb_array_elements(eb.data->'facts') f,
                     jsonb_array_elements_text(f->'context_terms') t
        WHERE f->>'id' = p->>'fact_id'
          AND lower(p->>'label') LIKE '%' || lower(t) || '%');

\echo '=== 7. DEFECT if any row: a value exceeding its chart denominator (bar overflows) ==='
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT c->>'id' AS chart, p->>'fact_id' AS fact,
       (SELECT (f->>'value')::numeric FROM jsonb_array_elements(eb.data->'facts') f
         WHERE f->>'id' = p->>'fact_id') AS value,
       COALESCE((c->>'max')::numeric,
                (SELECT (f->>'value')::numeric FROM jsonb_array_elements(eb.data->'facts') f
                  WHERE f->>'id' = c->>'max_fact_id')) AS denominator
  FROM eb, jsonb_array_elements(eb.data->'charts') c,
           jsonb_array_elements(c->'points') p
 WHERE (SELECT (f->>'value')::numeric FROM jsonb_array_elements(eb.data->'facts') f
         WHERE f->>'id' = p->>'fact_id')
     > COALESCE((c->>'max')::numeric,
                (SELECT (f->>'value')::numeric FROM jsonb_array_elements(eb.data->'facts') f
                  WHERE f->>'id' = c->>'max_fact_id'));

\echo '=== 8. DEFECT if any row: a chart definition containing a bare business figure ==='
\echo '    (prose fields must carry no digits — the register owns every number)'
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT c->>'id' AS chart, k AS prose_field, v AS text_with_digits
  FROM eb, jsonb_array_elements(eb.data->'charts') c,
           LATERAL (VALUES ('title', c->>'title'), ('caption', c->>'caption'),
                           ('source_note', c->>'source_note')) AS x(k, v)
 WHERE v ~ '[0-9]';

\echo '=== 9. INFORMATIONAL: what each chart will draw, in order ==='
WITH eb AS (SELECT data FROM site_specs
             WHERE site_id = :'site_id'::uuid AND aspect='evidence_base' AND is_current)
SELECT c->>'id' AS chart, COALESCE(c->>'pages','(all pages)') AS pages,
       p->>'label' AS label,
       (SELECT f->>'value' FROM jsonb_array_elements(eb.data->'facts') f
         WHERE f->>'id' = p->>'fact_id') AS value,
       COALESCE(c->>'unit','') AS unit,
       (SELECT f->>'verified_at' FROM jsonb_array_elements(eb.data->'facts') f
         WHERE f->>'id' = p->>'fact_id') AS verified
  FROM eb, jsonb_array_elements(eb.data->'charts') c,
           jsonb_array_elements(c->'points') p;
