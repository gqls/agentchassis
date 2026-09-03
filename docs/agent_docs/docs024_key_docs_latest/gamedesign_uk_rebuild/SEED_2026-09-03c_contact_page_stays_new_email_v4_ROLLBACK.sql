\set ON_ERROR_STOP on
-- ROLLBACK for SEED_2026-09-03c. Restores the 09-03b state: no contact page, no contact form,
-- no email address anywhere, and the address back to gamedesign@contactforsales.com.
--
-- ⚠ ONLY run this if the owner REVERSES the midday ruling. It re-imposes the exact banned claims
-- that blocked the homepage build at 11:03:32Z (work item ac76ec54), so expect that failure again.
--
-- Mechanism: site_specs is VERSIONED (is_current + superseded_at), so v3 is still on disk. This
-- supersedes v4 and re-flips the immediately-previous row per aspect. It does NOT delete v4.
BEGIN;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='mission_brief' AND is_current
     AND position('There is a contact page and it stays' in data->>'text') > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'ROLLBACK REFUSED: v4 mission_brief is not current (found %) — nothing to roll back', n; END IF;
END $g$;

-- mission_brief: supersede v4, restore the newest non-current row
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='mission_brief' AND is_current;
UPDATE site_specs SET is_current=true, superseded_at=NULL
 WHERE id = (SELECT id FROM site_specs
              WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='mission_brief' AND NOT is_current
              ORDER BY created_at DESC LIMIT 1);

-- evidence_base: same
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current;
UPDATE site_specs SET is_current=true, superseded_at=NULL
 WHERE id = (SELECT id FROM site_specs
              WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND NOT is_current
              ORDER BY created_at DESC LIMIT 1);

-- the address (updated in place by 09-03c, so reversed in place here)
UPDATE site_specs SET data = jsonb_set(data,'{email}','"gamedesign@contactforsales.com"'::jsonb)
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='submission' AND is_current;
UPDATE site_specs SET data = jsonb_set(data,'{contact,contact_email}','"gamedesign@contactforsales.com"'::jsonb)
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='briefing' AND is_current;

DO $v$
DECLARE bans int;
BEGIN
  SELECT jsonb_array_length(data->'banned_claims') INTO bans FROM site_specs
   WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND aspect='evidence_base' AND is_current;
  IF bans <> 18 THEN RAISE EXCEPTION 'ROLLBACK FAILED: expected 18 banned_claims restored, found %', bans; END IF;
  RAISE NOTICE 'ROLLBACK OK: 18 banned claims restored, contact page banned again';
END $v$;

COMMIT;
