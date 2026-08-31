-- 665 — state the owner's four service commitments in the terms, under an unlock/relock
--
-- OWNER 2026-08-31: "please unlock the terms pages, make the edits and lock them again. that's
-- probably the safest route." His four terms decisions (2026-08-26) are registered as facts but
-- appear on NO page: `[MEASURED 2026-08-31]` terms.html and privacy-policy.html mention none of
-- "30 days", "within a week", "9am", "GPU" or "playground".
--
-- THE TEXT IS NOT WRITTEN HERE. Every sentence is the `writer_line` of a registered, owner-attested
-- fact in `evidence_base.facts[]` — ft-data-location, ft-retention-default, ft-deletion-window,
-- ft-playground-hour, ft-booking-hours — reproduced verbatim. Nothing is composed.
--
-- ⚠ WHY A DIRECT EDIT AND NOT A REBUILD. A page-build regenerates the whole page through the
-- writer. On this page that risks an INVENTED CLAUSE, and the precedent is real: a rebuild of this
-- exact kind fabricated a contact email on another site (render_news_section_html.go:39-56). On a
-- page stating what we promise about deleting customer data, a fabricated commitment is a promise
-- nobody made. `rebuild_policy='owned'` is therefore left IN PLACE — the page keeps its protection
-- from regeneration; this is a targeted write underneath it.
--
-- ⚠ THE UNLOCK MUST CLEAR ALL FOUR COLUMNS. Writability is decided by `locked_at` alone; clearing
-- `lock_type`/`locked_by` only makes the row read as unlocked while `classifyComponentLock` treats
-- a set `locked_at` with no type as HARD (LANDMINES, "Clearing lock_type/locked_by does NOT unlock
-- a page_components row"). And the RELOCK restores the ORIGINAL `locked_at` (2026-07-21
-- 09:21:45.96136+00, `182_legal_pages`, permanent) rather than now() — that timestamp is the lock's
-- provenance and any lock-age review reads it.
--
-- Placed before "Intellectual property" so the service terms sit with the commercial sections.
-- Rollback: 665_..._ROLLBACK.sql

BEGIN;

-- 1. UNLOCK: all four columns, or it is not unlocked.
UPDATE page_components pc
   SET locked_at = NULL, locked_by = NULL, lock_type = NULL, lock_expires_at = NULL
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain='finetuning.uk' AND p.name='terms';

DO $$
DECLARE writable bool;
BEGIN
    SELECT (pc.locked_at IS NULL OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()))
      INTO writable FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='terms';
    IF NOT writable THEN RAISE EXCEPTION '665: terms did not become agent-writable — the unlock did not take'; END IF;
END $$;

-- 2. EDIT.
UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{content}',
                        to_jsonb(replace(pc.content_data->>'content', '<h2>Intellectual property</h2>', '<h2>The fine-tuning service</h2><p>Where you buy our &pound;99 fine-tuning service through this website, the following applies in addition to the sections above.</p><p><strong>Where your documents go.</strong> During training your documents sit on a rented GPU machine; for the handover they sit in our storage. That is the whole of it.</p><p><strong>How long we keep them.</strong> We keep your documents and your model for 30 days after we hand it over, then delete them.</p><p><strong>Deletion on request.</strong> Ask us to delete your documents and your model and we will, within a week.</p><p><strong>The included hour.</strong> One hour on the playground is included, to be used within 30 days of handover; more hours can be bought.</p><p><strong>Booking that hour.</strong> You pick the hour: 9am to 5pm UK time, Monday to Friday. Outside those hours we can usually arrange something.</p><h2>Intellectual property</h2>'))),
       updated_at = now()
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain='finetuning.uk' AND p.name='terms'
   AND pc.content_data->>'content' LIKE '%<h2>Intellectual property</h2>%'
   AND pc.content_data->>'content' NOT LIKE '%The fine-tuning service%';

-- 3. RELOCK with the ORIGINAL provenance, not now().
UPDATE page_components pc
   SET locked_at = TIMESTAMPTZ '2026-07-21 09:21:45.96136+00',
       locked_by = '182_legal_pages', lock_type = 'permanent', lock_expires_at = NULL
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain='finetuning.uk' AND p.name='terms';

DO $$
DECLARE writable bool; n_terms int; has_all bool; lk timestamptz; lb text; lt text;
BEGIN
    SELECT (pc.locked_at IS NULL OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW())),
           pc.locked_at, pc.locked_by, pc.lock_type
      INTO writable, lk, lb, lt
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='terms';

    IF writable THEN RAISE EXCEPTION '665: terms is STILL agent-writable after the relock'; END IF;
    IF lk <> TIMESTAMPTZ '2026-07-21 09:21:45.96136+00' OR lb <> '182_legal_pages' OR lt <> 'permanent' THEN
        RAISE EXCEPTION '665: lock provenance not restored (locked_at=%, by=%, type=%)', lk, lb, lt;
    END IF;

    SELECT count(*), bool_and(
             pc.content_data->>'content' LIKE '%30 days after we hand it over%' AND
             pc.content_data->>'content' LIKE '%within a week%' AND
             pc.content_data->>'content' LIKE '%rented GPU machine%' AND
             pc.content_data->>'content' LIKE '%9am to 5pm UK time%' AND
             pc.content_data->>'content' LIKE '%One hour on the playground is included%')
      INTO n_terms, has_all
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='terms';
    IF n_terms <> 1 THEN RAISE EXCEPTION '665: expected 1 terms component, found %', n_terms; END IF;
    IF NOT has_all THEN RAISE EXCEPTION '665: not all five commitments are present in the content'; END IF;

    RAISE NOTICE '665 OK: five commitments stated, lock restored to its original 2026-07-21 provenance, page not agent-writable';
END $$;

COMMIT;
