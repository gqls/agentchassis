-- 415_article_body_link_ink_repoint_ROLLBACK.sql
--
-- HAND-RUN ONLY. The trailing _ROLLBACK is uppercase, so run-migrations.sh's
-- SIDECAR_RE excludes it from auto-apply while still listing it.
--
-- Reverts 415: puts `.article-body__content a` back on the RAW --color-primary.
--
-- ⚠ THIS RESTORES A KNOWN-INVISIBLE STATE. On every dark site the links return to
-- ~1.1:1 against the page ground — the defect the owner reported twice. Only run
-- this if the repoint caused something worse. If the problem is the COLOUR rather
-- than the mechanism, prefer the kill-switch, which needs no migration and no
-- re-render:
--     legible_ink_enabled: false                       (fleet)
--     legible_ink_disabled_site_ids: ["<site uuid>"]   (one site)
--   on the `render_css_from_spec` step config of `webdesign-agent`. With the
--   companion absent, this rule's two-level fallback degrades to --color-primary
--   on its own, i.e. exactly the pre-415 rendering, without touching this row.
--
-- After running this, the 97 rendered placements keep the REPOINTED html until
-- they are re-rendered per (component, site).

BEGIN;

UPDATE content_components
SET html_template = replace(
      html_template,
      '.article-body__content a{color:var(--color-primary-ink,var(--color-primary,#1e40af))',
      '.article-body__content a{color:var(--color-primary,#1e40af)'
    ),
    updated_at = now()
WHERE name = 'article-body'
  AND html_template LIKE '%.article-body__content a{color:var(--color-primary-ink,var(--color-primary,#1e40af))%';

DO $$
DECLARE reverted int;
BEGIN
  SELECT count(*) INTO reverted FROM content_components
   WHERE name = 'article-body'
     AND html_template LIKE '%.article-body__content a{color:var(--color-primary,#1e40af)%';
  IF reverted <> 1 THEN
    RAISE EXCEPTION '415 ROLLBACK: expected exactly 1 article-body row back on the raw rule, found %', reverted;
  END IF;
  RAISE NOTICE '415 ROLLBACK OK: article-body links back on raw --color-primary (KNOWN INVISIBLE on dark sites)';
END $$;

COMMIT;
