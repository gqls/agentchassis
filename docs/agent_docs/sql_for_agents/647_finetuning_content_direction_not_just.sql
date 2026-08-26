-- 647 — finish 646: the three ', not just' demonstrations it did not reach
--
-- 646 cleared all 7 'rather than' and left 3 'not just' — found by RE-COUNTING the
-- live spec after applying, not by trusting the migration's own verify block, which
-- only asserted on 'rather than'. A verify block can only refuse what it was told to
-- look for.
--
-- All three are instructions demonstrating the construction they teach against, and
-- ', not just' is on the owner's OWN named list (the identity spec's [0]/[4]/[5] were
-- reseeded on 2026-08-24 for carrying exactly it). Form only; no rule loses force.
-- Rollback: 647_..._ROLLBACK.sql
BEGIN;
UPDATE site_specs ss SET is_current = false, superseded_at = now()
  FROM sites s WHERE s.id = ss.site_id AND s.domain = 'finetuning.uk'
   AND ss.aspect = 'content_direction' AND ss.is_current;
INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, notes)
SELECT ss.site_id, ss.aspect, (replace(replace(replace(ss.data::text, 'Headings should be outcome-oriented or problem-framing — not just noun labels.', 'Headings should be outcome-oriented or problem-framing.'), 'describe the problem solved and the mechanism — not just a vague positive outcome.', 'describe the problem solved and the mechanism.'), 'Reinforce vendor neutrality as a throughline across all pages, not just the homepage', 'Reinforce vendor neutrality as a throughline on every page'))::jsonb,
       'owner_ruling', 'claude-finetuning-uk-lane', true,
       'Finishes 646: the three ", not just" demonstrations, form-only.'
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
 WHERE s.domain = 'finetuning.uk' AND ss.aspect = 'content_direction'
   AND ss.superseded_at IS NOT NULL ORDER BY ss.superseded_at DESC LIMIT 1;
DO $$
DECLARE n_nj int; n_rt int; n_keys int; len_after int;
BEGIN
  SELECT (length(ss.data::text) - length(replace(ss.data::text,'not just',''))) / length('not just'),
         (length(ss.data::text) - length(replace(ss.data::text,'rather than',''))) / length('rather than'),
         (SELECT count(*) FROM jsonb_object_keys(ss.data)), length(ss.data->>'formatted')
    INTO n_nj, n_rt, n_keys, len_after
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='finetuning.uk' AND ss.aspect='content_direction' AND ss.is_current;
  IF n_nj <> 0 THEN RAISE EXCEPTION '647: % "not just" remain', n_nj; END IF;
  IF n_rt <> 0 THEN RAISE EXCEPTION '647: % "rather than" reappeared — wrong base row', n_rt; END IF;
  IF n_keys < 8 THEN RAISE EXCEPTION '647: only % keys survived', n_keys; END IF;
  IF len_after < 10000 THEN RAISE EXCEPTION '647: formatted is % chars', len_after; END IF;
  RAISE NOTICE '647 OK: 0 "not just", 0 "rather than", % keys, formatted % chars', n_keys, len_after;
END $$;
COMMIT;
