\set ON_ERROR_STOP on
-- SEED_2026-09-03c — OWNER RULING 2026-09-03 (midday), REVERSING the 09:45Z ruling in SEED_2026-09-03b.
--
--   "ok leave the contact page, the email can be gamedesignuk@contactforsales.com"
--
-- 09-03b had banned the contact page, the contact form and EVERY email address as shapes, and
-- wrote a CONTACT paragraph into the writer block forbidding all three. That combination BLOCKED
-- the homepage build at 11:03:32Z: the writer wrote "…the contact page is there for it" and the
-- banned-claim gate refused it (work item ac76ec54, needs_human_review). The gate was right; the
-- instruction behind it is what changed.
--
-- This seed reverses exactly that, and nothing else:
--   1. mission_brief   v4 — the no-contact sentence replaced by one that KEEPS the page and names
--                           the new address. Built by anchored replace() on the live v3 text, NOT
--                           re-pasted, so no other sentence can drift.
--   2. evidence_base   v4 — the TWO banned claims added by 09-03b removed (any-email; contact
--                           form/page). The THIRD claim it added (human-masthead) STAYS — the AI
--                           authorship ruling is untouched. writer_block's CONTACT paragraph
--                           replaced in place.
--   3. submission.email + briefing.contact.contact_email → gamedesignuk@contactforsales.com.
--
-- Every step is guarded: anchors must appear EXACTLY once and the banned-claim count must fall by
-- exactly 2, or the transaction aborts. Guards are DO/RAISE, never bare SELECTs — per LANDMINES,
-- ON_ERROR_STOP does not stop a COMMIT on a non-empty result set.
--
-- Apply: psql -f THIS FILE ONLY.  Rollback companion: SEED_2026-09-03c_..._ROLLBACK.sql
BEGIN;


-- ---------------------------------------------------------------- 0. PRE-FLIGHT GUARDS
DO $g$
DECLARE
  mission_hits int; wb_hits int; ban_count int; email_ban int; contact_ban int;
BEGIN
  SELECT count(*) INTO mission_hits FROM site_specs
   WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='mission_brief' AND is_current
     AND position('There is no contact page, no contact form and no email address anywhere on the site' in data->>'text') > 0;
  IF mission_hits <> 1 THEN
    RAISE EXCEPTION '09-03c REFUSED: expected 1 current mission_brief carrying the no-contact sentence, found %', mission_hits;
  END IF;

  SELECT count(*) INTO wb_hits FROM site_specs
   WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current
     AND position('CONTACT (owner ruling 2026-09-03): there is no contact page' in data->>'writer_block') > 0;
  IF wb_hits <> 1 THEN
    RAISE EXCEPTION '09-03c REFUSED: expected 1 current evidence_base whose writer_block carries the CONTACT paragraph, found %', wb_hits;
  END IF;

  SELECT jsonb_array_length(data->'banned_claims') INTO ban_count
    FROM site_specs WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current;
  IF ban_count IS NULL OR ban_count < 3 THEN
    RAISE EXCEPTION '09-03c REFUSED: banned_claims count is %, expected >= 3', ban_count;
  END IF;

  SELECT count(*) INTO email_ban FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') e
   WHERE s.site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND s.aspect='evidence_base' AND s.is_current
     AND e->>'pattern' = '[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}';
  SELECT count(*) INTO contact_ban FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') e
   WHERE s.site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND s.aspect='evidence_base' AND s.is_current
     AND e->>'pattern' = 'contact (form|page)|(fill in|submit) (the|this|a) form';
  IF email_ban <> 1 OR contact_ban <> 1 THEN
    RAISE EXCEPTION '09-03c REFUSED: expected exactly 1 email-ban and 1 contact-ban, found % and %', email_ban, contact_ban;
  END IF;
END $g$;

-- ---------------------------------------------------------------- 1. mission_brief v4
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='mission_brief' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT '8f17eb73-fc74-4718-8371-b3125bc4e414', 'mission_brief',
  jsonb_build_object('text', replace(prev.data->>'text',
    'There is no contact page, no contact form and no email address anywhere on the site; a reader who wants the tools goes to the sister site, and that is the only outward link the site needs to offer.',
    'There is a contact page and it stays. The address is gamedesignuk@contactforsales.com, and the contact page may name it plainly; do not invent any other address, phone number or postal address. A reader who wants the tools still goes to the sister site gamesdesign.co.uk.'
  )),
  'manual',
  'v4 2026-09-03 midday, owner reversing the 09:45Z no-contact ruling: the contact page STAYS and the address is gamedesignuk@contactforsales.com. Built by anchored replace() on v3; every other sentence byte-identical. The AI-authorship ruling is untouched.',
  true, true, 'gamedesign-uk-lane-2026-09-03'
FROM site_specs prev
WHERE prev.site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND prev.aspect='mission_brief'
ORDER BY prev.created_at DESC LIMIT 1;

-- ---------------------------------------------------------------- 2. evidence_base v4
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT '8f17eb73-fc74-4718-8371-b3125bc4e414', 'evidence_base',
  prev.data
  || jsonb_build_object(
       'banned_claims', (
         SELECT COALESCE(jsonb_agg(e), '[]'::jsonb)
           FROM jsonb_array_elements(prev.data->'banned_claims') e
          WHERE e->>'pattern' <> '[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}'
            AND e->>'pattern' <> 'contact (form|page)|(fill in|submit) (the|this|a) form'
       ),
       'writer_block', replace(prev.data->>'writer_block',
         'CONTACT (owner ruling 2026-09-03): there is no contact page, no contact form and no email address anywhere on the site. Do not write one, offer one, or refer to one. The only outward destination the site offers is the sister site gamesdesign.co.uk for tools.',
         'CONTACT (owner ruling 2026-09-03, REVISED the same day — this supersedes the earlier no-contact rule): the site HAS a contact page and it stays. The address is gamedesignuk@contactforsales.com. Name it plainly on the contact page, and refer to the contact page anywhere a reader would reasonably look for it. Do not invent any other email address, phone number or postal address, and do not claim a response time. The sister site gamesdesign.co.uk remains the destination for tools.'
       )
     ),
  'manual',
  'v4 2026-09-03 midday: the two 09-03b bans (any-email; contact form/page) REMOVED per owner ruling; the human-masthead ban RETAINED; writer_block CONTACT paragraph replaced in place. Unblocks work item ac76ec54.',
  true, true, 'gamedesign-uk-lane-2026-09-03'
FROM site_specs prev
WHERE prev.site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND prev.aspect='evidence_base'
ORDER BY prev.created_at DESC LIMIT 1;

-- ---------------------------------------------------------------- 3. the address itself
UPDATE site_specs
   SET data = jsonb_set(data, '{email}', '"gamedesignuk@contactforsales.com"'::jsonb)
 WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='submission' AND is_current
   AND data->>'email' = 'gamedesign@contactforsales.com';

UPDATE site_specs
   SET data = jsonb_set(data, '{contact,contact_email}', '"gamedesignuk@contactforsales.com"'::jsonb)
 WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='briefing' AND is_current
   AND data->'contact'->>'contact_email' = 'gamedesign@contactforsales.com';

-- ---------------------------------------------------------------- 4. POST-CONDITIONS
DO $v$
DECLARE
  ban_now int; still_banned int; mission_ok int; wb_ok int; sub_ok int; brief_ok int;
BEGIN
  SELECT jsonb_array_length(data->'banned_claims') INTO ban_now
    FROM site_specs WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current;

  SELECT count(*) INTO still_banned FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') e
   WHERE s.site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND s.aspect='evidence_base' AND s.is_current
     AND (e->>'pattern' = '[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}'
       OR e->>'pattern' = 'contact (form|page)|(fill in|submit) (the|this|a) form');
  IF still_banned <> 0 THEN
    RAISE EXCEPTION '09-03c FAILED: % contact/email ban(s) survived the filter', still_banned;
  END IF;

  -- the human-masthead ban from 09-03b must SURVIVE — this seed reverses two rules, not three
  IF (SELECT count(*) FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') e
       WHERE s.site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND s.aspect='evidence_base' AND s.is_current
         AND e->>'reason' LIKE '%author is an AI LLM%') <> 1 THEN
    RAISE EXCEPTION '09-03c FAILED: the AI-authorship ban was lost — this seed must not touch it';
  END IF;

  SELECT count(*) INTO mission_ok FROM site_specs
   WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='mission_brief' AND is_current
     AND position('There is a contact page and it stays' in data->>'text') > 0
     AND position('gamedesignuk@contactforsales.com' in data->>'text') > 0
     AND position('no contact page, no contact form' in data->>'text') = 0
     AND position('written by an AI' in data->>'text') > 0;   -- the untouched ruling still there
  IF mission_ok <> 1 THEN RAISE EXCEPTION '09-03c FAILED: mission_brief v4 did not land as expected'; END IF;

  SELECT count(*) INTO wb_ok FROM site_specs
   WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current
     AND position('the site HAS a contact page and it stays' in data->>'writer_block') > 0
     AND position('there is no contact page' in data->>'writer_block') = 0;
  IF wb_ok <> 1 THEN RAISE EXCEPTION '09-03c FAILED: writer_block CONTACT paragraph not replaced'; END IF;

  SELECT count(*) INTO sub_ok FROM site_specs WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='submission' AND is_current
     AND data->>'email' = 'gamedesignuk@contactforsales.com';
  SELECT count(*) INTO brief_ok FROM site_specs WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='briefing' AND is_current
     AND data->'contact'->>'contact_email' = 'gamedesignuk@contactforsales.com';
  IF sub_ok <> 1 OR brief_ok <> 1 THEN
    RAISE EXCEPTION '09-03c FAILED: address not updated (submission=%, briefing=%)', sub_ok, brief_ok;
  END IF;

  RAISE NOTICE '09-03c OK: banned_claims now %, contact page permitted, address = gamedesignuk@contactforsales.com', ban_now;
END $v$;

COMMIT;

SELECT aspect,
       created_at::timestamp(0),
       jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb)) AS bans,
       (data->>'text' ILIKE '%There is a contact page and it stays%') AS mission_v4
  FROM site_specs
 WHERE site_id = '8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND aspect IN ('mission_brief','evidence_base') AND is_current;
