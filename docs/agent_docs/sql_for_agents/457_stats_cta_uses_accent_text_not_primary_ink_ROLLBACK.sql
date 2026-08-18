-- 457_stats_cta_uses_accent_text_not_primary_ink_ROLLBACK.sql
--
-- Restores `.stats-cta` to the `--color-primary-ink` foreground that 456 left it
-- with. ⚠ THAT STATE IS A MEASURED 1.61:1 FAILURE on ai-agent-orchestration.com
-- (#768eb2 on #F0A500). Roll back only to isolate a worse regression elsewhere,
-- and say which.
--
-- To return to the PRE-456 state for this declaration instead, run 456's own
-- rollback afterwards — it restores the bare `var(--color-primary, #1a1a2e)`, which
-- on an amber fill was legible.
--
-- Templates only. Placements already re-rendered keep the accent-text html until
-- re-rendered again.

BEGIN;

UPDATE content_components
SET html_template = regexp_replace(
      html_template,
      '(background:\s*var\(--color-accent[^;]*\);\s*)color:\s*var\(--color-accent-text, var\(--color-primary,(\s*)(#[0-9a-fA-F]{3,8})\)\)',
      '\1color: var(--color-primary-ink,var(--color-primary,\2\3))',
      'g'),
    updated_at = now()
WHERE name = 'system-stats'
  AND html_template LIKE '%color: var(--color-accent-text, var(--color-primary%';

DO $$
DECLARE
  reverted int;
  bare_ink int;
BEGIN
  SELECT count(*) INTO reverted FROM content_components
   WHERE name = 'system-stats'
     AND html_template LIKE '%color: var(--color-accent-text, var(--color-primary%';
  IF reverted <> 0 THEN
    RAISE EXCEPTION 'rollback 457: % system-stats row(s) still carry the accent-text rule', reverted;
  END IF;

  SELECT count(*) INTO bare_ink FROM content_components
   WHERE html_template ~ 'var\(\s*--color-(primary|accent)-ink\s*\)';
  IF bare_ink <> 0 THEN
    RAISE EXCEPTION 'rollback 457: % row(s) left a BARE ink reference behind', bare_ink;
  END IF;

  RAISE NOTICE 'rollback 457 OK: .stats-cta restored to --color-primary-ink. THE 1.61:1 FAILURE IS BACK; re-render to propagate.';
END $$;

COMMIT;
