-- p4_08_unlock_derived_hub_listing.sql — CORRECTION to p4_03: never lock a section that derives.
--
-- WHAT I GOT WRONG. p4_03 locked the guides hub's `guide-list` section, with this comment:
--     "The hub's listing copy (the items themselves stay query-resolved and must NOT be frozen —
--      the lock protects the surrounding copy, not the derived list)."
-- **That is false, and the code says so.** A lock is applied to the `page_components` ROW. It cannot
-- distinguish authored copy from derived items inside that row, because both live in the same row's
-- content_data/rendered_html. `SavePageSectionsAction` (save_page_sections_action.go:487-534) loads
-- actively-locked rows, holds them out of the rebuild DELETE, and re-attaches them verbatim —
-- "Human-locked rows must survive the rebuild with copy AND row identity", logging
-- "preserving human-locked section over rebuilt copy (bugs_open/058)". So the locked row's rendered
-- HTML stands and the freshly-resolved one is discarded.
--
-- HOW IT SURFACED, and it was luck. The final verification sweep showed the live guides hub
-- listing ONE card (Patents) when two guides exist. The copyright guide shipped at ~09:22, after
-- the hub's last render at ~08:57, so it was simply not yet re-resolved — but the lock applied at
-- ~09:33 would have made that permanent: every future guide would have been written, deployed, and
-- silently never listed, with the hub reporting a successful render each time. The
-- self-populating listing that is increment 1's whole reusable contribution would have been dead
-- on arrival, and nothing would have alerted anyone.
--
-- THE RULE, which is more general than this page:
--   **A section whose schema has ANY `query.*` source must never be locked.** Locking freezes the
--   derivation, and a frozen derivation is indistinguishable from a working one until the data it
--   was supposed to track has moved on.
-- Locks are for AUTHORED sections only. The patents guide, the copyright guide and the patent-check
-- tool page have no query-sourced fields, so p4_03's locks on those 8 sections are correct and stay.
--
-- The hub's copy does not need the lock anyway: after p4_04 its `cta_url` / `cta_button_label` /
-- `eyebrow_label` are content_data-driven, and a `section_data_resolved` rerender does not invoke
-- the content writer while content_data is non-empty. That is the right protection for a section
-- that must keep deriving — guard the inputs, not the output.

\set ON_ERROR_STOP on

BEGIN;

-- Guard: this must only ever unlock sections that actually derive something.
DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/index.html'
    AND pc.slot_name = 'guide-list'
    AND EXISTS (
      SELECT 1 FROM jsonb_each(cc.input_schema->'fields') AS f(k,v)
      WHERE v->>'source' LIKE 'query.%'
    );
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: expected exactly 1 query-sourced guide-list section on the hub, found % — re-ground before unlocking.', n;
  END IF;
END
$guard$;

UPDATE page_components pc
SET locked_at = NULL,
    locked_by = NULL,
    lock_type = NULL,
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/index.html'
  AND pc.slot_name = 'guide-list';

INSERT INTO doc_notes (subject_type, subject_key, body, categories, created_at)
VALUES (
  'pipeline', 'component_locks',
  'RULE (idea.uk p4_08, 2026-07-25): never apply a page_components lock to a section whose component '
  || 'schema has any query.* sourced field. Locks are ROW-granular — SavePageSectionsAction '
  || '(save_page_sections_action.go:487-534) holds locked rows out of the rebuild and re-attaches them '
  || 'verbatim, so the derivation freezes while every render still reports success. idea.uk''s guides '
  || 'hub was locked this way for ~10 minutes and would have silently stopped listing new guides '
  || 'forever. Protect a deriving section by making its authored fields content_data-driven instead '
  || '(see p4_04), not by locking it.',
  '["component-locks","derived-fields","idea.uk","do-not-lock-derived"]'::jsonb,
  now()
);

COMMIT;

-- Locks after the correction: authored pages locked, deriving hub free.
SELECT p.url, pc.slot_name, COALESCE(pc.lock_type,'(none)') AS lock_type,
       EXISTS (SELECT 1 FROM jsonb_each(cc.input_schema->'fields') AS f(k,v) WHERE v->>'source' LIKE 'query.%') AS derives
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url IN ('/guides/index.html','/guides/patents/index.html','/guides/copyright/index.html',
                '/tools/patent-check/index.html','/tools.html')
ORDER BY p.url, pc.position;
