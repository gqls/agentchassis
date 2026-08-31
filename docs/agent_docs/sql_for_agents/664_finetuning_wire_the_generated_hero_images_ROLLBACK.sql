-- ROLLBACK for 664: remove the hero_url this migration set, leaving `careers` (which was
-- wired by the pipeline itself, not by 664) untouched.
BEGIN;
UPDATE page_components pc
   SET content_data = pc.content_data - 'hero_url', updated_at = now()
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain = 'finetuning.uk'
   AND p.name IN ('about','approach','case-studies','contact','model-approach-selector',
                  'services','tool-ai-readiness-checker','use-cases');
COMMIT;
