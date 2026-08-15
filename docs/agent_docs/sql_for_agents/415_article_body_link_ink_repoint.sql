-- 415_article_body_link_ink_repoint.sql
--
-- bugs_open/122 — repoint the shared `article-body` component's in-prose LINK
-- colour from the raw brand colour to its legible companion.
--
-- WHY NOW, AND WHY NOT BEFORE. `.article-body__content a` paints links with the
-- RAW `--color-primary`. On a dark site that is a near-black navy on a near-black
-- ground: dartsonline.com measured **1.11:1** against `--color-background` and
-- **1.06:1** against `--color-surface` on 2026-08-15 — invisible, and the symptom
-- the owner reported twice (2026-08-12 with a screenshot, and again 2026-08-15
-- after the ink canary landed).
--
-- This exact one-line repoint was DELIBERATELY WITHHELD until 2026-08-14. Before
-- the derivation repair, `--color-primary-ink` resolved to `--color-text` on every
-- site in the fleet, so this edit would have written BODY TEXT into 97 placements
-- across 20 sites — a de-branding that `render_audit.py` would have scored a clean
-- pass, because it measures contrast and has no opinion about brand. After the
-- repair (commits 12cf55015 + 8ad05d01a + d4bbbf645, council d60aab29 APPROVED,
-- live from v1.0.1301) the companion is a lightness-shifted BRAND colour at the
-- framework default target of 5.0. Same edit, opposite outcome.
--
-- OWNER AUTHORISATION 2026-08-15: "please go ahead with the unreadable links in
-- the screenshot - it only needs to go to 5.00 - the framework default." This
-- migration adds no target of its own: it points at `--color-primary-ink`, which
-- the renderer derives at `inkMinContrast` (5.0). Nothing here pins a number, so
-- if the framework default ever moves, these links follow it automatically.
--
-- EXPECTED EFFECT, dartsonline.com (worst-of-four grounds 5.122 by construction):
--   #1A1F2E -> #94a0c2   1.11:1 -> 7.00:1 on background, 1.06:1 -> 5.93:1 on surface
--
-- SHAPE: the documented TWO-LEVEL fallback, never bare. If the companion is ever
-- absent — including via the `legible_ink_enabled` kill-switch, which emits
-- nothing — the declaration degrades to today's colour rather than dropping. A
-- bare `var(--color-primary-ink)` would make the whole declaration vanish. The
-- fleet-wide bare-reference count was 0 on 2026-08-15 and this keeps it at 0.
--
-- SCOPE: one row. `content_components.article-body` is the single source; it is
-- rendered into 97 `page_components` across 20 sites, which keep the OLD html
-- until re-rendered. This migration does NOT re-render them — that is a separate,
-- per-(component, site) step, exactly as with migrations 338 and 368.
--
-- ROLLBACK: docs/agent_docs/sql_for_agents/415_article_body_link_ink_repoint_ROLLBACK.sql
--   (restores the bare `var(--color-primary,#1e40af)` form; re-render to propagate)

BEGIN;

UPDATE content_components
SET html_template = replace(
      html_template,
      '.article-body__content a{color:var(--color-primary,#1e40af)',
      '.article-body__content a{color:var(--color-primary-ink,var(--color-primary,#1e40af))'
    ),
    updated_at = now()
WHERE name = 'article-body'
  AND html_template LIKE '%.article-body__content a{color:var(--color-primary,#1e40af)%';

-- Guard the exact post-condition. A verify block of bare SELECTs cannot stop a
-- COMMIT (ON_ERROR_STOP ignores a non-empty result), so this must RAISE.
DO $$
DECLARE
  repointed int;
  leftover  int;
  bare_ink  int;
BEGIN
  SELECT count(*) INTO repointed FROM content_components
   WHERE name = 'article-body'
     AND html_template LIKE '%.article-body__content a{color:var(--color-primary-ink,var(--color-primary,#1e40af))%';
  IF repointed <> 1 THEN
    RAISE EXCEPTION '415: expected exactly 1 article-body row carrying the repointed rule, found %', repointed;
  END IF;

  SELECT count(*) INTO leftover FROM content_components
   WHERE html_template LIKE '%.article-body__content a{color:var(--color-primary,#1e40af)%';
  IF leftover <> 0 THEN
    RAISE EXCEPTION '415: % row(s) still carry the RAW --color-primary link rule', leftover;
  END IF;

  -- The kill-switch safety invariant: never introduce a BARE ink reference,
  -- because the disabled path emits nothing and a bare var() drops the whole
  -- declaration instead of degrading.
  SELECT count(*) INTO bare_ink FROM content_components
   WHERE html_template ~ 'var\(\s*--color-(primary|accent)-ink\s*\)';
  IF bare_ink <> 0 THEN
    RAISE EXCEPTION '415: % content_component(s) now carry a BARE ink reference; the two-level fallback is mandatory', bare_ink;
  END IF;

  RAISE NOTICE '415 OK: article-body links repointed to --color-primary-ink (two-level fallback); 97 rendered placements across 20 sites still need a re-render';
END $$;

COMMIT;
