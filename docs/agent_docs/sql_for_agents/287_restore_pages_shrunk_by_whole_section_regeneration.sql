-- 287 — restore the pages whose content was lost to whole-section regenerations (bugs_open/178)
--
-- Owner-directed 2026-08-02: "please restore all the pages that have been reduced".
-- DATA only. Every restore sources its bytes from `page_component_history` BY ROW ID
-- — no content is pasted into this file, so what is restored is exactly what the
-- writer itself snapshotted immediately before overwriting.
--
-- ── DISPOSITIONS, one per candidate page (all six from 178's table) ───────
-- RESTORED here:
--   robot-hands /how-to-specify-a-gripper.html   generic-text-block
--       ecb4b420 (4,439) restored, PLUS the tool anchor the 93f2a3b7 item was
--       legitimately raised to add, inserted into the workpiece paragraph. The
--       merge keeps both truths: the old reference prose and the new crosslink.
--   fundamentallyai /tools/review-council-simulator.html   tool slot
--       content_data was NULLed 07-30 20:13; live rendered_html (32,876) still
--       serves the working tool — the 012 class, page alive by luck. f321055b
--       restored to content_data. ⚠ DO NOT RERENDER THIS PAGE until the
--       restored shape is checked against the renderer: the live render is
--       currently the only known-good copy, and a rerender regenerates FROM
--       content_data. No rerender item is created for it here, deliberately.
--   vetcomparison /about.html   faq + about-content
--       Shrunk by the 08-01 08:13 write: faq 7,206→2,197 (-70%), about-content
--       3,043→1,882 (-38%). 623a2888 + 3a20c2bc restored. Trade-off accepted
--       and stated: three small writes today (10:47–11:05) reshaped these slots
--       (possibly CTA fixes); the restore discards those edits. If a misdirected
--       CTA re-appears, discovery re-detects it — the reverse (lost prose) had
--       no detector, which is why it sat unnoticed since 08-01.
-- NOT restored, deliberately:
--   gamesdesign /tools/bayesian-ranking.html — NOT DAMAGE. Current hero/guide
--       fields are clean and purposeful; the larger old blob was a site-context
--       param dump. That shrink was a legitimate rewrite (07-29).
--   vonc /about.html — 12→6 slots is bugs_closed/156's md5-duplicate removal.
--   relojistas /glosario/index.html — a DELETED slot holding DefinedTermSet
--       JSON-LD (b0e119a4, 2,816). Confirmed absent site-wide (0 hits in
--       page_components AND site_components), so it was lost not moved — but
--       history records no slot_name and inventing one may not be referenced by
--       the template. Recorded in bugs_open/178; needs the 178 fix lane or an
--       owner call, not a blind INSERT.

BEGIN;

-- ── STEP 1 — PRE-FLIGHT: pages still in the state the dispositions were made on ──
-- All five booleans must be TRUE; anything else means a writer moved a page
-- under us — ROLLBACK and re-verify that page before proceeding.
SELECT
  (SELECT length(content_data::text) FROM page_components
    WHERE page_id='5a385981-c2fd-4edb-bc4d-927b93177281' AND slot_name='generic-text-block') = 1806
      AS robothands_as_measured,
  (SELECT content_data IS NULL FROM page_components pc
    WHERE pc.page_id=(SELECT id FROM pages WHERE url='/tools/review-council-simulator.html'
                      AND site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com'))
      AND pc.slot_name='tool-review-council-simulator')
      AS fundamentallyai_still_nulled,
  (SELECT length(content_data::text) FROM page_components pc
    WHERE pc.page_id=(SELECT id FROM pages WHERE url='/about.html'
                      AND site_id=(SELECT id FROM sites WHERE domain='vetcomparison.uk'))
      AND pc.slot_name='faq') = 2197
      AS vet_faq_as_measured,
  (SELECT count(*) FROM page_component_history
    WHERE (id::text LIKE 'ecb4b420%' AND length(content_data::text)=4439)
       OR (id::text LIKE 'f321055b%' AND length(content_data::text)=32444)
       OR (id::text LIKE '623a2888%' AND length(content_data::text)=7206)
       OR (id::text LIKE '3a20c2bc%' AND length(content_data::text)=3043)) = 4
      AS all_four_snapshots_present,
  (SELECT count(*) FROM site_work_items
    WHERE status IN ('triaged','approved','claimed')
      AND page_id IN ('5a385981-c2fd-4edb-bc4d-927b93177281',
                      (SELECT id FROM pages WHERE url='/about.html' AND site_id=(SELECT id FROM sites WHERE domain='vetcomparison.uk')))) = 0
      AS no_queued_writer_against_targets;

-- ── STEP 2 — robot-hands: restore + keep the anchor (the merge) ───────────
UPDATE page_components pc
SET content_data = jsonb_set(
      h.content_data, '{content}',
      to_jsonb(replace(h.content_data->>'content',
        'actual mass to 240g.</p>',
        'actual mass to 240g. Use our <a href="/tools/gripper-safety-factor-calculator/index.html">Gripper Safety Factor Calculator</a> to determine a defensible multiplier based on part fragility, acceleration profiles, and surface condition.</p>'))),
    updated_at = now()
FROM page_component_history h
WHERE h.id::text LIKE 'ecb4b420%'
  AND pc.page_id='5a385981-c2fd-4edb-bc4d-927b93177281'
  AND pc.slot_name='generic-text-block'
  AND length(pc.content_data::text) = 1806;          -- refuses a moved row

-- ── STEP 3 — fundamentallyai: restore the durable source ──────────────────
UPDATE page_components pc
SET content_data = h.content_data,
    updated_at = now()
FROM page_component_history h
WHERE h.id::text LIKE 'f321055b%'
  AND pc.page_id=(SELECT id FROM pages WHERE url='/tools/review-council-simulator.html'
                  AND site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com'))
  AND pc.slot_name='tool-review-council-simulator'
  AND pc.content_data IS NULL;                        -- refuses a moved row

-- ── STEP 4 — vetcomparison: restore faq + about-content ──────────────────
UPDATE page_components pc
SET content_data = h.content_data, updated_at = now()
FROM page_component_history h
WHERE h.id::text LIKE '623a2888%'
  AND pc.page_id=(SELECT id FROM pages WHERE url='/about.html'
                  AND site_id=(SELECT id FROM sites WHERE domain='vetcomparison.uk'))
  AND pc.slot_name='faq';

UPDATE page_components pc
SET content_data = h.content_data, updated_at = now()
FROM page_component_history h
WHERE h.id::text LIKE '3a20c2bc%'
  AND pc.page_id=(SELECT id FROM pages WHERE url='/about.html'
                  AND site_id=(SELECT id FROM sites WHERE domain='vetcomparison.uk'))
  AND pc.slot_name='about-content';

-- ── STEP 5 — queue rerenders for the two pages whose RENDER must catch up ──
-- (fundamentallyai deliberately excluded — see header.) Platform's own path:
-- page_rerender items, claimed by the now-fair dispatcher.
INSERT INTO site_work_items (site_id, source, item_type, severity, summary,
                             page_id, priority, handler_agent, status, created_by,
                             item_key, pipeline, spec)
SELECT p.site_id, 'triage-287', 'page_rerender', 'high',
       'Rerender after content restore (bugs_open/178, migration 287)',
       p.id, 10, 'page-rerender', 'triaged', 'triage-287',
       'restore_287:' || p.id, 'build',
       jsonb_build_object('domain', s.domain, 'reason', 'section_data_resolved',
                          'page_id', p.id::text,
                          'filename', regexp_replace(p.url, '^.*/', ''),
                          'page_name', regexp_replace(regexp_replace(p.url,'^/',''),'\.html$',''))
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.id='5a385981-c2fd-4edb-bc4d-927b93177281'
   OR (p.url='/about.html' AND s.domain='vetcomparison.uk')
ON CONFLICT DO NOTHING;

-- ── STEP 6 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- Expect: robot-hands ~4630 (4439 + anchor sentence), fundamentallyai 32444,
-- vet faq 7206, vet about-content 3043, and 2 rerender items created.
SELECT 'robot-hands gtb' AS what, length(content_data::text) AS len FROM page_components
 WHERE page_id='5a385981-c2fd-4edb-bc4d-927b93177281' AND slot_name='generic-text-block'
UNION ALL
SELECT 'fai tool slot', length(content_data::text) FROM page_components pc
 WHERE pc.page_id=(SELECT id FROM pages WHERE url='/tools/review-council-simulator.html'
                   AND site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com'))
   AND pc.slot_name='tool-review-council-simulator'
UNION ALL
SELECT 'vet faq', length(content_data::text) FROM page_components pc
 WHERE pc.page_id=(SELECT id FROM pages WHERE url='/about.html' AND site_id=(SELECT id FROM sites WHERE domain='vetcomparison.uk'))
   AND pc.slot_name='faq'
UNION ALL
SELECT 'vet about-content', length(content_data::text) FROM page_components pc
 WHERE pc.page_id=(SELECT id FROM pages WHERE url='/about.html' AND site_id=(SELECT id FROM sites WHERE domain='vetcomparison.uk'))
   AND pc.slot_name='about-content'
UNION ALL
SELECT 'rerender items', count(*)::int FROM site_work_items WHERE item_key LIKE 'restore_287:%';

COMMIT;

-- ── ROLLBACK ── the pre-restore states are themselves snapshotted: the shrunken
-- versions are in page_component_history rows written by the same overwrites
-- that did the damage; and this file's step-1 lengths (1806/NULL/2197/1882)
-- identify them. Restoring "back" would mean re-applying the damage — unlikely
-- to be wanted; listed for completeness.
--
-- ── VERIFY AT THE ARTEFACT (after the rerenders complete) ──
--   robot-hands: rendered generic-text-block should hold BOTH the ISO 9409-1
--     integration paragraph AND the gripper-safety-factor-calculator anchor:
--     SELECT rendered_html LIKE '%ISO 9409-1%' AND rendered_html LIKE '%gripper-safety-factor-calculator%'
--     FROM page_components WHERE page_id='5a385981-...' AND slot_name='generic-text-block';
--   vetcomparison: faq rendered length should roughly triple.
--   fundamentallyai: NOTHING should change on the live page (no rerender queued).
