-- 456_ink_repoint_ai_agent_orchestration_templates.sql
--
-- bugs_open/122 family — repoint FOREGROUND `--color-primary` declarations to the
-- legible companion, in the 12 shared templates that ai-agent-orchestration.com
-- actually renders. Same one-line shape as migration 415, wider selector set.
--
-- WHY. ai-agent-orchestration.com serves 44 firm contrast failures across 4 pages
-- (render_audit.py, over_image excluded, 2026-08-17), 14 of them at **1.00:1** —
-- text painted in exactly its own background colour. The mechanism:
--
--   .differentiator-item h3 { ... color: var(--color-primary, #1a1a2e); }
--
-- is byte-identical in `content_components.html_template` and in the served
-- `page_components.rendered_html`, so the template is the source. That site's
-- palette makes the declaration fatal:
--
--   --color-primary   #0D1117
--   --color-surface   #0D1117    <- IDENTICAL
--   --color-background #080B10
--
-- so the heading is drawn in the surface colour, on the surface. Only 2 of 23
-- sites carrying a design_intent palette are degenerate this way
-- (ai-agent-orchestration.com and oufe.com, the latter UNAUDITED).
--
-- ⚠ THE `#1a1a2e` FALLBACK IS PRESENT IN THE SOURCE AND NEVER APPLIED, because the
-- variable is set. A grep of the stylesheet reads "dark navy heading"; the browser
-- paints invisible. Every colour quoted here is a getComputedStyle value.
--
-- WHY NOT FIX THE PALETTE INSTEAD. Considered and rejected on measurement:
-- `--color-primary` is dual-role on this site — **37 foreground (`color`) uses and
-- 24 `background` uses** in rendered component CSS. Lightening it so headings read
-- would put light fills under the white/near-white labels that sit on them, i.e.
-- trade 20 failures for a fresh set. That is render_audit.py's own defect family 2
-- ("a token used in two roles — correct in one place, invisible in the other").
-- The ink companion exists precisely so the FILL can stay dark while the
-- FOREGROUND becomes legible; this migration uses it for what it is for.
--
-- SHAPE: the documented TWO-LEVEL fallback, never bare. Where the companion is
-- absent — an un-migrated renderer, or the `legible_ink_enabled` kill-switch, which
-- emits nothing — the declaration degrades to today's exact colour rather than
-- dropping. A bare `var(--color-primary-ink)` would make the whole declaration
-- vanish. This migration keeps the fleet-wide bare-reference count at 0.
--
-- SAFE ON LIGHT SITES BY CONSTRUCTION. These 12 templates are shared (
-- differentiators-section is placed on 15 sites, contact-info on 11, tool-guide-intro
-- on 9, content-block-about on 7, system-stats on 5). On a site whose primary
-- already clears the contrast floor the derivation returns the smallest sufficient
-- change, so the companion is at or near the brand colour; where no companion is
-- emitted the fallback is today's value. Neither branch can darken a light site.
--
-- ⚠ `--color-primary-text` IS A DIFFERENT TOKEN AND IS NOT TOUCHED. It is the label
-- colour that sits ON a primary fill (3 occurrences in this set, as
-- `var(--color-primary-text, var(--color-white,#fff))`). Both patterns below require
-- `,` or `)` immediately after `--color-primary`, so `-text` cannot match.
--
-- SCOPE: 12 rows, named explicitly rather than selected by a live subquery so the
-- set cannot shift under a concurrent placement. **144 further templates fleet-wide
-- carry the same bare-foreground defect** (156 of 294 total; only 4 mention the ink
-- companion at all) — that pass belongs to the bugfix_122 lane and is deliberately
-- NOT taken here. See CONTRIB_2026-08-17_ink_is_live_and_the_site_is_still_broken.md.
--
-- DOES NOT RE-RENDER. Rendered placements keep the OLD html until re-rendered,
-- exactly as with 338, 368 and 415.
--
-- ROLLBACK: 456_ink_repoint_ai_agent_orchestration_templates_ROLLBACK.sql

BEGIN;

-- Form 1: no fallback           color: var(--color-primary)
-- Form 2: hex fallback          color: var(--color-primary, #1a1a2e) / #1e40af
-- Whitespace after `color:` and after the comma is preserved via backreferences,
-- so a template's own formatting is not reflowed.
UPDATE content_components
SET html_template = regexp_replace(
      regexp_replace(
        html_template,
        'color:(\s*)var\(--color-primary\)',
        'color:\1var(--color-primary-ink,var(--color-primary))',
        'g'),
      'color:(\s*)var\(--color-primary,(\s*)(#[0-9a-fA-F]{3,8})\)',
      'color:\1var(--color-primary-ink,var(--color-primary,\2\3))',
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
  AND html_template ~ 'color:\s*var\(--color-primary[,)]';

-- Guard the exact post-condition. A verify block of bare SELECTs cannot stop a
-- COMMIT (ON_ERROR_STOP ignores a non-empty result), so this must RAISE.
DO $$
DECLARE
  named        int;
  leftover     int;
  repointed    int;
  bare_ink     int;
  primary_text int;
BEGIN
  SELECT count(*) INTO named FROM content_components
   WHERE name IN (
      'AI Model Directory (full listing)','contact-info','content-block-about',
      'differentiators-section','system-stats',
      'tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
      'tool-ai-agent-roi-estimator-ai-agent-orchestration-com',
      'tool-ai-readiness-quiz-finetuning-uk-ai-agent-orchestration-com',
      'tool-automation-savings-estimator-finetuning-uk-ai-agent-orchestration-com',
      'tool-build-vs-buy-analyzer-ai-agent-orchestration-com','tool-guide-intro',
      'tool-llm-cost-calculator-ai-agent-orchestration-com');
  IF named <> 12 THEN
    RAISE EXCEPTION '456: expected 12 named templates, found % — the name set has drifted, ABORTING rather than half-applying', named;
  END IF;

  -- No bare FOREGROUND primary may remain in the named set. This is what catches a
  -- whitespace or fallback variant the two patterns above did not anticipate.
  SELECT count(*) INTO leftover FROM content_components
   WHERE name IN (
      'AI Model Directory (full listing)','contact-info','content-block-about',
      'differentiators-section','system-stats',
      'tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
      'tool-ai-agent-roi-estimator-ai-agent-orchestration-com',
      'tool-ai-readiness-quiz-finetuning-uk-ai-agent-orchestration-com',
      'tool-automation-savings-estimator-finetuning-uk-ai-agent-orchestration-com',
      'tool-build-vs-buy-analyzer-ai-agent-orchestration-com','tool-guide-intro',
      'tool-llm-cost-calculator-ai-agent-orchestration-com')
     AND html_template ~ 'color:\s*var\(--color-primary[,)]';
  IF leftover <> 0 THEN
    RAISE EXCEPTION '456: % named template(s) still carry a bare foreground --color-primary', leftover;
  END IF;

  SELECT count(*) INTO repointed FROM content_components
   WHERE html_template LIKE '%var(--color-primary-ink,var(--color-primary%';
  IF repointed < 12 THEN
    RAISE EXCEPTION '456: expected at least 12 rows carrying the repointed rule, found %', repointed;
  END IF;

  -- The kill-switch safety invariant: never introduce a BARE ink reference, because
  -- the disabled path emits nothing and a bare var() drops the whole declaration.
  SELECT count(*) INTO bare_ink FROM content_components
   WHERE html_template ~ 'var\(\s*--color-(primary|accent)-ink\s*\)';
  IF bare_ink <> 0 THEN
    RAISE EXCEPTION '456: % content_component(s) now carry a BARE ink reference; the two-level fallback is mandatory', bare_ink;
  END IF;

  -- --color-primary-text must be untouched: it is the label ON a primary fill.
  SELECT count(*) INTO primary_text FROM content_components
   WHERE html_template LIKE '%--color-primary-ink,var(--color-primary-text%';
  IF primary_text <> 0 THEN
    RAISE EXCEPTION '456: % row(s) wrapped --color-primary-text in an ink fallback; that token is a different role and must not be repointed', primary_text;
  END IF;

  RAISE NOTICE '456 OK: 12 templates repointed to --color-primary-ink (two-level fallback). Rendered placements keep the OLD html until re-rendered. 144 further templates fleet-wide still carry the defect — bugfix_122 lane.';
END $$;

COMMIT;
