-- 636_aiao_two_tool_templates_token_collisions.sql
--
-- ai-agent-orchestration.com — the last 4 firm contrast failures that are this lane's
-- to fix, on the two tool pages whose stored HTML ALREADY carried the ink token. That
-- is why they survived migrations 456/469/625: **they are not the 456 defect.**
--
-- ⚠ DIAGNOSED BY ASKING THE BROWSER WHICH DECLARATION WINS, not by grepping CSS.
-- `getComputedStyle` plus a CSSOM walk over `document.styleSheets` matching each
-- failing element (2026-08-26). This site's served stylesheet has lied about headings
-- before, and two of the five rules below would never have been found by a template
-- grep for the 456 pattern.
--
-- ── THE ROOT, AND IT IS THE SAME ONE AS 456 WEARING A DIFFERENT FACE ────────────
-- On this site `--color-primary` and `--color-surface` are THE SAME VALUE (`#0D1117`,
-- probed live). So ANY rule that pairs those two tokens collapses to 1.00:1 — and it
-- reads as perfectly sensible CSS, because on every other site in the fleet a label in
-- `--color-surface` on a `--color-primary` fill is exactly right.
--
-- ── THE FIVE DECLARATIONS, each measured against its own ground ─────────────────
--
--   tool-automation-savings-estimator-ai-agent-orchestration-com
--     .method-details summary   color:var(--color-primary)    1.00 -> ink        5.66
--     #…-calculate-savings-button color:var(--color-surface)  1.00 -> primary-text 18.92
--     .cta-link                 color:var(--color-secondary)  1.09 -> accent-ink  9.09
--     .result-value             color:var(--color-primary)    1.00 -> ink        5.66
--   tool-build-vs-buy-analyzer-ai-agent-orchestration-com
--     .bvb-btn-primary          color:var(--color-surface)    1.00 -> primary-text 18.92
--
-- ⚠ `.result-value` IS NOT ONE OF THE AUDIT'S 17. It sits in the result panel, which
-- is hidden until the visitor runs the calculator, so `render_audit.py` — which reads
-- the page as loaded — cannot see it. It is the same 1.00:1 defect and a user who
-- actually uses the tool would meet it. **Found by censusing the template for the
-- collapsing PAIR rather than by fixing what the instrument listed.** Anything the
-- audit reports is a lower bound on a page with conditional UI.
--
-- ⚠ `.cta-link` IS AN OVERRIDE, NOT AN OMISSION. The template already carries
-- `a { color: var(--color-accent-ink, var(--color-accent)) }`, which is legible
-- (9.09:1). `.cta-link` has higher specificity and replaces it with
-- `--color-secondary` (`#161B22`) — a near-black on a near-black ground. So the fix
-- is convergence on the token the sibling rule already uses, not a new choice.
--
-- SCOPE: two components, both `-ai-agent-orchestration-com` forks placed on exactly
-- ONE site each (measured 2026-08-26). No fleet blast radius, unlike `contact-form`
-- (20 sites), which is deliberately NOT touched here.
--
-- TWO-LEVEL FALLBACKS THROUGHOUT, per the standing rule: where the companion token is
-- absent the declaration degrades to today's exact colour rather than dropping. A bare
-- `var(--color-primary-ink)` would make the whole declaration vanish.
--
-- BOTH PAGES ARE `rebuild_policy='generic'`, so unlike `625`'s owned tool pages the
-- ordinary propagation works: a page-scoped `template_changed` rerender. This file
-- changes TEMPLATES ONLY and files nothing.
--
-- ROLLBACK: 636_..._ROLLBACK.sql (byte-exact from migration_backups)

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '636_aiao_two_tool_templates_token_collisions', 'content_components', cc.id::text,
       jsonb_build_object('html_template', cc.html_template), 'pre-636 template for ' || cc.name
FROM content_components cc
WHERE cc.id IN ('5e3a4ca5-9792-4b96-87d6-e4e984792bab','fd8b92ea-3cbe-4bfd-a9ee-6af7da0fa0a2');

UPDATE content_components
SET html_template = replace(replace(replace(
      html_template,
      'color: var(--color-primary);',   'color: var(--color-primary-ink, var(--color-primary));'),
      'color: var(--color-surface);',   'color: var(--color-primary-text, var(--color-surface));'),
      'color: var(--color-secondary);', 'color: var(--color-accent-ink, var(--color-secondary));'),
    updated_at = now()
WHERE id IN ('5e3a4ca5-9792-4b96-87d6-e4e984792bab','fd8b92ea-3cbe-4bfd-a9ee-6af7da0fa0a2');

DO $$
DECLARE b int; n int; tpl text;
BEGIN
  SELECT count(*) INTO b FROM migration_backups
   WHERE migration_name='636_aiao_two_tool_templates_token_collisions';
  IF b <> 2 THEN RAISE EXCEPTION '636: expected 2 backup rows, wrote %', b; END IF;

  -- No collapsing declaration may survive in either template.
  SELECT count(*) INTO n FROM content_components
   WHERE id IN ('5e3a4ca5-9792-4b96-87d6-e4e984792bab','fd8b92ea-3cbe-4bfd-a9ee-6af7da0fa0a2')
     AND html_template ~ 'color:\s*var\(--color-(primary|surface|secondary)\)\s*;';
  IF n <> 0 THEN
    RAISE EXCEPTION '636: % template(s) still carry a bare collapsing colour token', n;
  END IF;

  -- All five repointed declarations present: 2 ink (summary + result-value), 2
  -- primary-text (the two buttons), 1 accent-ink (.cta-link).
  SELECT html_template INTO tpl FROM content_components WHERE id='5e3a4ca5-9792-4b96-87d6-e4e984792bab';
  IF (SELECT count(*) FROM regexp_matches(tpl,'var\(--color-primary-ink, var\(--color-primary\)\)','g')) <> 2 THEN
    RAISE EXCEPTION '636: expected 2 ink repoints in the savings estimator (summary AND the hidden .result-value)';
  END IF;
  IF tpl !~ 'var\(--color-primary-text, var\(--color-surface\)\)' THEN
    RAISE EXCEPTION '636: the savings-estimator button label was not repointed';
  END IF;
  IF tpl !~ 'var\(--color-accent-ink, var\(--color-secondary\)\)' THEN
    RAISE EXCEPTION '636: .cta-link was not repointed to the token its sibling a-rule already uses';
  END IF;

  SELECT html_template INTO tpl FROM content_components WHERE id='fd8b92ea-3cbe-4bfd-a9ee-6af7da0fa0a2';
  IF tpl !~ 'var\(--color-primary-text, var\(--color-surface\)\)' THEN
    RAISE EXCEPTION '636: .bvb-btn-primary was not repointed';
  END IF;

  -- Never a bare companion reference: the disabled path emits nothing and a bare
  -- var() drops the whole declaration.
  SELECT count(*) INTO n FROM content_components
   WHERE id IN ('5e3a4ca5-9792-4b96-87d6-e4e984792bab','fd8b92ea-3cbe-4bfd-a9ee-6af7da0fa0a2')
     AND html_template ~ 'var\(\s*--color-(primary-ink|accent-ink|primary-text)\s*\)';
  IF n <> 0 THEN RAISE EXCEPTION '636: % template(s) carry a BARE companion reference', n; END IF;

  -- Scope: nothing outside these two ids was written in this transaction.
  SELECT count(*) INTO n FROM content_components
   WHERE updated_at >= now() - interval '1 second'
     AND id NOT IN ('5e3a4ca5-9792-4b96-87d6-e4e984792bab','fd8b92ea-3cbe-4bfd-a9ee-6af7da0fa0a2');
  IF n <> 0 THEN RAISE EXCEPTION '636: % component(s) outside the two targets were written; scope escaped', n; END IF;

  RAISE NOTICE '636 OK: 5 collapsing declarations repointed across 2 aiao-only templates (incl. one the audit CANNOT see). Propagate with a page-scoped template_changed rerender, then verify over HTTP.';
END $$;

COMMIT;
