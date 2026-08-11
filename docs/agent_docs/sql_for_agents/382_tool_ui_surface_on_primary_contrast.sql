-- 382 — tool UI: retire the `color: var(--color-surface)` on
-- `background: var(--color-primary)` pairing (illegible where surface is dark)
--
-- WHY. The pairing assumes --color-surface is a pale shade that contrasts with
-- --color-primary. It is a variable, and on a dark-surface site it is dark too.
-- Found by the tool-acceptance vision pass's FIRST ever run (2026-08-11, run
-- 0ee53904…, bugs_open/243): dartsonline.com resolves primary #1A1F2E /
-- surface #1E2436 → contrast 1.06:1 against the WCAG AA floor of 4.5:1 — the
-- selected option chips and the submit CTA are invisible. Fleet census of the
-- idiom (both halves in one is_active html_template): 9 components / 7
-- functions live on 8 domains; 6 of the 8 are legible (pale surface), so the
-- idiom is unguarded rather than wrong. Second casualty:
-- mortgagecalculator.co.uk at 2.95:1 (#b59230 fill, #ffffff label).
--
-- THE FIX. Swap the FILL, keep the label: `background: var(--color-text)` under
-- the existing `color: var(--color-surface)`. text-on-surface is the pairing
-- the site itself guarantees (it is body text on a card); measured from each
-- affected site's served stylesheet on 2026-08-11 it runs 10.35:1 (worst,
-- mortgagecalculator) to 17.85:1 (best, vetcomparison) — every site ≥ AA.
-- Evidence + per-site table: staged_component_build NOTES
-- `## 2026-08-11 (parallel session)`; owner decision in chat the same day
-- ("fix the shared component").
--
-- Trade-off, stated: on the sites where the old pairing was legible the solid
-- fill loses its brand colour (e.g. vetcomparison blue → dark navy). On 4 of
-- the 6 healthy sites --color-text is within a shade of --color-primary, so
-- the change is barely visible; legibility everywhere is worth the brand
-- colour on the other two. The structural alternative — a --color-on-primary
-- token in the palette vocabulary — is a framework change with fleet-wide
-- stylesheet regeneration; this migration does not preclude it.
--
-- SCOPE. Exact-string replace of `background: var(--color-primary);` (with the
-- semicolon — the one color-mix() use of the token does not match), pinned to
-- the 9 measured component ids. 14 occurrences total (chips, CTAs, one
-- highlight card + the isi stage label/hover). Every occurrence sits inside a
-- <style> block (verified per template, 2026-08-11). The guard fails the file
-- if ANY active row fleet-wide still carries the pair after the update — so if
-- a 10th, unmeasured row has appeared since the census, the whole file rolls
-- back and the census must be re-run rather than silently half-fixing.
--
-- AFTER APPLY: templates are inert until each page re-renders. Re-render the 9
-- affected active pages via staged_component_build/scripts/RERENDER_page.sh
-- <site_id> <domain> <page_id> section_data_resolved  — the reason is
-- load-bearing (template changes ship ONLY on the rerender_sections path).
-- Canary two pages and diff the served HTML before doing the rest.
--
-- ROLLBACK: 382_tool_ui_surface_on_primary_contrast_ROLLBACK.sql (restores the
-- 9 templates byte-exact from migration_backups).

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '382_tool_ui_surface_on_primary_contrast.sql',
       'content_components', cc.id::text,
       jsonb_build_object('html_template', cc.html_template),
       'pre-382: carries surface-on-primary pair; function=' || cc.function
  FROM content_components cc
 WHERE cc.id IN (
   'c243e0e0-a466-4557-a4df-04c6ffa5acfd',  -- tool-automation-savings-estimator (canonical)
   '16a0bf97-b576-491b-a460-4268cf7f56fa',  -- tool-automation-savings-estimator (fork)
   'dc0f9f3a-ee94-4501-9c4e-0a14acb67193',  -- tool-automation-savings-estimator (fork)
   'cc3c7b15-e250-4c4e-b0ba-2c99c04652ee',  -- tool-bridging-compound
   '0776f252-f150-4f5e-9538-3a5a908be427',  -- tool-cma-obligation-checker
   '7c5660a4-f2bb-4a23-b37d-0084101f684c',  -- tool-fuel-cost-estimator
   'f5017854-4f4d-4c2d-8721-2a02d0e9989d',  -- tool-idea-stage-identifier
   '76d4e8cd-a4a9-4a82-8747-ec4a11e8287a',  -- tool-rate-scenarios
   'c1d9060a-5465-4b44-95da-3a45ab7b55bf'   -- tool-setup-builder
 );

UPDATE content_components
   SET html_template = replace(html_template,
                               'background: var(--color-primary);',
                               'background: var(--color-text);'),
       updated_at = now()
 WHERE id::text IN (SELECT target_id FROM migration_backups
                     WHERE migration_name = '382_tool_ui_surface_on_primary_contrast.sql');

DO $$
DECLARE
  backed     int;
  remaining  int;
  swapped    int;
  occurrences int;
BEGIN
  SELECT count(*) INTO backed FROM migration_backups
   WHERE migration_name = '382_tool_ui_surface_on_primary_contrast.sql';
  IF backed <> 9 THEN
    RAISE EXCEPTION '382 guard: expected 9 backup rows, found % — id list vs live table drifted, re-run the census', backed;
  END IF;

  -- The fleet-wide zero: no ACTIVE component may still carry the pair. This is
  -- deliberately broader than the 9 ids — a 10th, unmeasured carrier fails the
  -- file rather than surviving it.
  SELECT count(*) INTO remaining FROM content_components
   WHERE is_active
     AND html_template LIKE '%background: var(--color-primary);%'
     AND html_template LIKE '%color: var(--color-surface)%';
  IF remaining <> 0 THEN
    RAISE EXCEPTION '382 guard: % active row(s) still carry surface-on-primary after the update', remaining;
  END IF;

  -- Occurrence identity, not just row count: the census counted 14 pair sites
  -- across the 9 templates (3+1+1+3+2+1+1+1+1). Assert exactly 14 landed.
  SELECT count(*), COALESCE(sum(
           (length(cc.html_template)
            - length(replace(cc.html_template, 'background: var(--color-text);', '')))
           / length('background: var(--color-text);')), 0)
    INTO swapped, occurrences
    FROM content_components cc
   WHERE cc.id::text IN (SELECT target_id FROM migration_backups
                          WHERE migration_name = '382_tool_ui_surface_on_primary_contrast.sql');
  IF swapped <> 9 OR occurrences <> 14 THEN
    RAISE EXCEPTION '382 guard: expected 14 swapped occurrences across 9 rows, found % across %', occurrences, swapped;
  END IF;
END $$;

COMMIT;
