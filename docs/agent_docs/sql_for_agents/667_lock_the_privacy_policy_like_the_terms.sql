-- 667 — protect the privacy policy the way the terms page is protected
--
-- OWNER 2026-08-31: "yes, go ahead and lock the privacy policy pages."
--
-- WHY. Migration 666 put three data-handling commitments into the privacy policy — retention,
-- deletion-on-request, and where documents physically sit during training. `[MEASURED 2026-08-31]`
-- that page was `rebuild_policy='generic'` with NO component lock, so a future page-build could
-- silently reword or drop them. `terms` has carried both protections since 2026-07-21
-- (`182_legal_pages`, permanent) and its commitments are safe; this closes the asymmetry.
--
-- SCOPE: exactly one page. The site has two legal pages, `terms` (already protected) and
-- `privacy-policy`. A looser name match also catches `tool-privacy-redactor` and its guide, which
-- are TOOLS and must not be locked — hence the explicit name, not a pattern.
--
-- BOTH protections, because they stop different things:
--   rebuild_policy='owned'   — the framework will not REGENERATE the page (the LLM-rewrite risk)
--   component lock, permanent — nothing agent-side WRITES the row
--
-- ⚠ `locked_at` is set to now() and `locked_by` names THIS migration, deliberately. This is a NEW
-- lock, not a restoration: inventing an older provenance would misrepresent when the page became
-- protected, and lock-age reviews read that timestamp. (Contrast 665, which RESTORED terms'
-- original 2026-07-21 stamp because that lock already existed.)
--
-- ⚠ CONSEQUENCE, stated because it is the cost: this page can no longer be updated by the normal
-- path. Future edits need the same unlock/edit/relock 665 used — and the unlock must clear ALL FOUR
-- lock columns, because clearing only `lock_type` leaves `locked_at` set and
-- `classifyComponentLock` then treats the row as HARD while every column you can see reads unlocked.
--
-- Rollback: 667_..._ROLLBACK.sql

BEGIN;

UPDATE pages p SET rebuild_policy = 'owned', updated_at = now()
  FROM sites s WHERE s.id = p.site_id AND s.domain = 'finetuning.uk' AND p.name = 'privacy-policy';

UPDATE page_components pc
   SET locked_at = now(), locked_by = '667_legal_pages', lock_type = 'permanent', lock_expires_at = NULL
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain = 'finetuning.uk'
   AND p.name = 'privacy-policy' AND pc.build_status <> 'removed';

DO $$
DECLARE writable bool; pol text; lt text; n_tools int; has_all bool;
BEGIN
    SELECT (pc.locked_at IS NULL OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW())),
           COALESCE(p.rebuild_policy,'generic'), pc.lock_type,
           (pc.content_data->>'content' LIKE '%rented GPU machine%'
            AND pc.content_data->>'content' LIKE '%30 days after we hand it over%'
            AND pc.content_data->>'content' LIKE '%within a week%')
      INTO writable, pol, lt, has_all
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='privacy-policy' AND pc.build_status <> 'removed';

    IF writable THEN RAISE EXCEPTION '667: privacy-policy is still agent-writable'; END IF;
    IF pol <> 'owned' THEN RAISE EXCEPTION '667: rebuild_policy is %, want owned', pol; END IF;
    IF lt <> 'permanent' THEN RAISE EXCEPTION '667: lock_type is %, want permanent', lt; END IF;
    -- Locking a page that had LOST the commitments would protect the wrong thing.
    IF NOT has_all THEN RAISE EXCEPTION '667: the three commitments are not present — refusing to lock a page that does not carry them'; END IF;

    -- The tool pages a looser match would have caught must be untouched.
    SELECT count(*) INTO n_tools FROM pages p JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name LIKE '%privacy-redactor%' AND COALESCE(p.rebuild_policy,'generic')='owned';
    IF n_tools <> 0 THEN RAISE EXCEPTION '667: % privacy-redactor TOOL page(s) were locked — the match was too loose', n_tools; END IF;

    RAISE NOTICE '667 OK: privacy-policy is owned + permanently locked, carries all three commitments, and no tool page was touched';
END $$;

COMMIT;
