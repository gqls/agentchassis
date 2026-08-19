-- ROLLBACK for 490 — restores internal-linker's pre-490 state exactly:
-- output_format back to "array", the LIMIT 15 query, the bare
-- {{range .candidate_pages}} template, and REMOVES fail_on_non_numeric.
-- NOTE this restores bugs_open/313's defect (the dead branch) by design — a
-- rollback restores the prior state, it does not pick a better one.
-- Sidecar: never applied by the runner; run by hand only.

BEGIN;

SELECT snapshot_agent('internal-linker',
  '490_ROLLBACK: pre-rollback');

DO $$
DECLARE ofmt text; tmpl text; flag text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_candidate_pages,config,output_format}',
         default_config#>>'{workflow,steps,plan_links,config,prompt_template}',
         default_config#>>'{workflow,steps,check_candidates,config,fail_on_non_numeric}'
    INTO ofmt, tmpl, flag
    FROM agent_definitions WHERE id = '93cffe67-baf4-4fb1-bec9-ba546fb24a54';

  IF ofmt IS DISTINCT FROM 'object' THEN
    RAISE EXCEPTION '490_ROLLBACK: output_format is %, expected object — 490 is not applied here; refusing', ofmt;
  END IF;
  IF position('{{range .candidate_pages.rows}}' in tmpl) = 0 THEN
    RAISE EXCEPTION '490_ROLLBACK: template does not carry the 490 form; refusing';
  END IF;
  IF flag IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '490_ROLLBACK: fail_on_non_numeric is % — not 490''s state; refusing', flag;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config =
       jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,load_candidate_pages,config,output_format}',
             to_jsonb('array'::text)
           ),
           '{workflow,steps,load_candidate_pages,config,query}',
           to_jsonb($oldq$SELECT p.name, p.url, p.title, p.page_type, LEFT(string_agg(pc.rendered_html, ' '), 800) as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name != $2 AND p.status = 'active' AND p.page_type IN ('content', 'service', 'landing', 'tool') GROUP BY p.name, p.url, p.title, p.page_type HAVING COUNT(pc.id) > 0 ORDER BY p.name LIMIT 15$oldq$::text)
         ),
         '{workflow,steps,plan_links,config,prompt_template}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,plan_links,config,prompt_template}',
             '{{range .candidate_pages.rows}}',
             '{{range .candidate_pages}}'
           )
         )
       ) #- '{workflow,steps,check_candidates,config,fail_on_non_numeric}',
       updated_at = now()
 WHERE id = '93cffe67-baf4-4fb1-bec9-ba546fb24a54'
   AND type = 'internal-linker'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE ofmt text; q text; tmpl text; flag text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_candidate_pages,config,output_format}',
         default_config#>>'{workflow,steps,load_candidate_pages,config,query}',
         default_config#>>'{workflow,steps,plan_links,config,prompt_template}',
         default_config#>>'{workflow,steps,check_candidates,config,fail_on_non_numeric}'
    INTO ofmt, q, tmpl, flag
    FROM agent_definitions WHERE id = '93cffe67-baf4-4fb1-bec9-ba546fb24a54';

  IF ofmt IS DISTINCT FROM 'array' THEN
    RAISE EXCEPTION '490_ROLLBACK: output_format is % after rollback, expected array', ofmt;
  END IF;
  IF position('LIMIT 15' in q) = 0 THEN
    RAISE EXCEPTION '490_ROLLBACK: LIMIT 15 absent after rollback';
  END IF;
  IF position('{{range .candidate_pages}}' in tmpl) = 0
     OR position('.candidate_pages.rows' in tmpl) > 0 THEN
    RAISE EXCEPTION '490_ROLLBACK: template not restored to the bare form';
  END IF;
  IF flag IS NOT NULL THEN
    RAISE EXCEPTION '490_ROLLBACK: fail_on_non_numeric still present after rollback';
  END IF;
END $$;

COMMIT;
