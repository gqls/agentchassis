-- SQL_2026-08-02_retire_stale_brand_head_undeployed_items.sql
--
-- Retires the 11 stale `undeployed_asset` work items carrying a brand-head
-- purpose (favicon / og_card). Owner-authorised 2026-08-02, following council
-- `abd9b119` round 2, whose gating objection these items were the evidence for.
--
-- ─── WHY THESE ITEMS ARE FALSE, MEASURED NOT ASSUMED (2026-08-02) ───
--
-- Each says "Asset '<purpose>' generated but not deployed to site". All four
-- artefacts they name serve HTTP 200 on the wire:
--
--   dartsonline.com/assets/images/favicon.png   200
--   dartsonline.com/assets/images/og-card.png   200
--   robot-hands.com/assets/images/favicon.png   200
--   robot-hands.com/assets/images/og-card.png   200
--
-- And the CURRENT check would raise none of them:
--   * `check_undeployed_assets` excludes brand-head purposes from its generic
--     half (`AND NOT (COALESCE(a.purpose,'') = ANY($2::text[]))`) since
--     bugs_closed/142.
--   * Its brand-head half raises only when NO asset row exists. dartsonline's
--     rows record the published paths exactly (`rowAtPublishedPath`), so it is
--     silent. robot-hands' rows record `/assets/images/input-data.asset-key.jpg`
--     — an unresolved template literal — which hits the check's deliberate
--     third state: observed as a `brandHeadProvenanceNote`, never filed,
--     because claiming "never generated" would be a FALSE claim.
--
-- So these are findings from a predicate that no longer exists. They are not
-- repaired by re-derivation (nothing needs deriving — the artefacts are live)
-- and must not be dispatched: until chassis v1.0.1229 they would have deployed
-- an arbitrary image to `og_card.png`, and after bugs_closed/168 unified the
-- path derivation they would have written to the REAL `og-card.png`, replacing
-- a live social card. That clobber is now refused in code, so this cleanup is
-- hygiene rather than mitigation — but leaving eleven false claims in a
-- dispatchable queue is its own defect.
--
-- ─── WHAT IS NOT FIXED HERE, AND MUST NOT BE READ AS FIXED ───
--
-- robot-hands.com's two asset rows still carry the unresolved template literal
-- in `assets.url`. That is a REAL defect and it is `bugs_open/152`'s (the
-- asset-URL rewrite lane). Cancelling these items does not touch it, and the
-- reason string below points at 152 so the signal is not lost with the item.
--
-- ─── WHY `cancelled` AND NOT `complete` ───
--
-- `complete` asserts the work was done; nothing did any work. `cancelled` with
-- a reason is the existing precedent for retiring a finding that no longer
-- holds — there are already 3 rows carrying `spec->>'reason' = 'stale'`.
-- NOTE `idx_swi_dedup` excludes `cancelled`, so this does not block a future
-- genuine re-detection of the same item_key. That is deliberate: if the
-- condition ever becomes true again, the check must be free to say so.

\set ON_ERROR_STOP on

-- ── 1. DRY RUN — read this before running the transaction below ──
SELECT swi.id, s.domain, swi.status, swi.spec->>'purpose' AS purpose,
       swi.item_key, swi.created_at::date
  FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
 WHERE swi.item_type = 'undeployed_asset'
   AND swi.spec->>'purpose' IN ('og_card','favicon')
   AND swi.status NOT IN ('complete','cancelled','rejected','verified','wont_fix')
 ORDER BY s.domain, swi.created_at;

-- ── 2. APPLY ──
-- Guarded: the transaction aborts if the affected count is not what was
-- measured, so a concurrent change by another session cannot be swept up
-- silently. 11 rows at 2026-08-02 19:4x (2 detected dartsonline, 9 unresolved
-- robot-hands).
BEGIN;

DO $repair$
DECLARE
    expected CONSTANT int := 11;
    affected int;
BEGIN
    UPDATE site_work_items swi
       SET status = 'cancelled',
           updated_at = NOW(),
           resolution_path = 'stale_predicate',
           error = NULL,
           spec = COALESCE(swi.spec,'{}'::jsonb) || jsonb_build_object(
               'reason', 'stale',
               'cancelled_at', NOW()::text,
               'cancelled_by', 'bugfix_168_deployed_asset_path lane (owner-authorised)',
               'cancel_detail',
                 'Finding is false: the named artefact serves HTTP 200. Raised by a '
              || 'predicate that no longer exists - check_undeployed_assets has excluded '
              || 'brand-head purposes from its generic half since bugs_closed/142, and its '
              || 'brand-head half is silent when an asset row exists. Do NOT re-dispatch: '
              || 'deploy_image_asset now REFUSES a brand-head purpose (bugs_closed/168), '
              || 'because after the path unification this item would have overwritten the '
              || 'live og-card.png. robot-hands.com assets.url still holds the unresolved '
              || 'template literal /assets/images/input-data.asset-key.jpg - that defect '
              || 'is REAL and belongs to bugs_open/152; it is not fixed here.')
     WHERE swi.item_type = 'undeployed_asset'
       AND swi.spec->>'purpose' IN ('og_card','favicon')
       AND swi.status NOT IN ('complete','cancelled','rejected','verified','wont_fix');

    GET DIAGNOSTICS affected = ROW_COUNT;

    -- The guard: another session may have touched these between the dry run
    -- and now. A surprise count means the population moved, and a data repair
    -- that silently repairs a different set than the one you read is the whole
    -- failure mode this file exists to avoid.
    IF affected <> expected THEN
        RAISE EXCEPTION 'row-count guard failed: expected % rows, updated % - '
                        're-run the dry run and re-read the population before applying',
                        expected, affected;
    END IF;

    RAISE NOTICE 'OK: % stale brand-head undeployed_asset items cancelled', affected;
END
$repair$;

COMMIT;

-- ── 3. VERIFY ──
SELECT status, count(*) AS n
  FROM site_work_items
 WHERE item_type = 'undeployed_asset' AND spec->>'purpose' IN ('og_card','favicon')
 GROUP BY 1 ORDER BY 1;

-- Expect: ZERO rows in any non-terminal status.
SELECT count(*) AS still_open_must_be_zero
  FROM site_work_items
 WHERE item_type = 'undeployed_asset'
   AND spec->>'purpose' IN ('og_card','favicon')
   AND status NOT IN ('complete','cancelled','rejected','verified','wont_fix');
