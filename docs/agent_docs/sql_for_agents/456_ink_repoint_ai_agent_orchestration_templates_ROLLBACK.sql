-- 456_ink_repoint_ai_agent_orchestration_templates_ROLLBACK.sql
--
-- Reverses 456 by unwrapping the ink companion from the 12 named templates,
-- restoring the bare foreground `var(--color-primary…)` form.
--
-- ⚠ THIS RESTORES THE DEFECT. On ai-agent-orchestration.com the bare form is
-- invisible text (1.00:1, primary == surface). Roll back only if the repoint
-- caused a worse regression somewhere else, and say where.
--
-- ⚠ ROLLING BACK IS NOT ENOUGH ON ITS OWN. Placements already re-rendered against
-- the repointed template keep the ink-aware html until they are re-rendered again;
-- this file changes templates only, exactly as 456 did.
--
-- The two patterns are the inverse of 456's and are anchored on the full wrapped
-- form, so a declaration that carried an ink companion BEFORE 456 (there were 4
-- such templates fleet-wide, none of them in this named set) cannot be stripped by
-- accident — the name filter excludes them anyway.

BEGIN;

UPDATE content_components
SET html_template = regexp_replace(
      regexp_replace(
        html_template,
        'color:(\s*)var\(--color-primary-ink,var\(--color-primary\)\)',
        'color:\1var(--color-primary)',
        'g'),
      'color:(\s*)var\(--color-primary-ink,var\(--color-primary,(\s*)(#[0-9a-fA-F]{3,8})\)\)',
      'color:\1var(--color-primary,\2\3)',
      'g'),
    updated_at = now()
WHERE name IN (
      'AI Model Directory (full listing)',
      'contact-info',
      'content-block-about',
      'differentiators-section',
      'system-stats',
      'tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
      'tool-ai-agent-roi-estimator-ai-agent-orchestration-com',
      'tool-ai-readiness-quiz-finetuning-uk-ai-agent-orchestration-com',
      'tool-automation-savings-estimator-finetuning-uk-ai-agent-orchestration-com',
      'tool-build-vs-buy-analyzer-ai-agent-orchestration-com',
      'tool-guide-intro',
      'tool-llm-cost-calculator-ai-agent-orchestration-com')
  AND html_template LIKE '%var(--color-primary-ink,var(--color-primary%';

DO $$
DECLARE
  still_wrapped int;
  bare_ink      int;
BEGIN
  SELECT count(*) INTO still_wrapped FROM content_components
   WHERE name IN (
      'AI Model Directory (full listing)','contact-info','content-block-about',
      'differentiators-section','system-stats',
      'tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
      'tool-ai-agent-roi-estimator-ai-agent-orchestration-com',
      'tool-ai-readiness-quiz-finetuning-uk-ai-agent-orchestration-com',
      'tool-automation-savings-estimator-finetuning-uk-ai-agent-orchestration-com',
      'tool-build-vs-buy-analyzer-ai-agent-orchestration-com','tool-guide-intro',
      'tool-llm-cost-calculator-ai-agent-orchestration-com')
     AND html_template LIKE '%var(--color-primary-ink,var(--color-primary%';
  IF still_wrapped <> 0 THEN
    RAISE EXCEPTION 'rollback 456: % named template(s) still carry the ink wrapper', still_wrapped;
  END IF;

  SELECT count(*) INTO bare_ink FROM content_components
   WHERE html_template ~ 'var\(\s*--color-(primary|accent)-ink\s*\)';
  IF bare_ink <> 0 THEN
    RAISE EXCEPTION 'rollback 456: % row(s) left a BARE ink reference behind', bare_ink;
  END IF;

  RAISE NOTICE 'rollback 456 OK: 12 templates restored to the bare foreground form. THE CONTRAST DEFECT IS BACK; re-render to propagate.';
END $$;

COMMIT;
