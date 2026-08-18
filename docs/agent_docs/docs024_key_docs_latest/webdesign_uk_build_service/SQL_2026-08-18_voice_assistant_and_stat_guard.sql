-- SQL_2026-08-18 — two writer_block changes, one supersede.
--
-- (1) THE OWNER'S VOICE BRIEF, 2026-08-18: "we're more like a helpful assistant than a
--     marketing bot. It's not statement, statement, statement, but rather statement,
--     statement, here's what you can do, or here's what's next step by step, or here's a
--     list of people that can help." His worked example: after the ZIP, say you will need
--     to host it and name free hosting (Netlify), then show how.
--     The register ALREADY holds the raw material: third_party_options names six services
--     by category. What was missing is the instruction to USE them as help rather than
--     list them as exclusions.
--
-- (2) THE STAT GUARD, and this is why the home page could not rebuild. Measured
--     2026-08-18 09:57 from agent_error_log (context.issues), the run's only failure:
--       type unregistered_stat | value "1 day" | location brief-explanation.stat_2_value
--       "a figure published in a stat field matches no evidence_base fact value"
--     The writer turned the HEDGED fact build_duration ("usually ready the next day") into
--     a BARE STAT ("1 day"). The gate is correct to refuse it: a stat field publishes a
--     figure as fact, and "usually next day" is not the number 1. The fix is to keep hedged
--     facts out of stat fields, NOT to attest "1 day" as a number — attesting it would
--     convert the owner's hedge into a promise, which is the opposite of what he asked for
--     ("I don't really want to promise anything").
--
-- Facts are UNCHANGED here; only writer_block moves. Appended by anchor, never retyped.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain = 'webdesign.uk' AND ss.aspect = 'evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{writer_block}', to_jsonb(
      c.data->>'writer_block' || E'\n\n'
      || 'HOW TO BE USEFUL RATHER THAN PROMOTIONAL (owner brief, 2026-08-18). Write like a helpful assistant, not a marketing bot. Do not stack statement on statement. State the thing, then tell the reader what they can do about it: the next step, the short list of what to do in order, or the people and services that can do it for them. A page that only asserts is doing half the job. Worked example the owner gave, and the shape to copy: the customer gets a ZIP of the finished site, so the next sentence tells them they will need to host it somewhere, names free hosting they can use, and shows what to do. The register already carries six named third-party services by category in third_party_options; use them as HELP where the reader would otherwise be stuck, not merely as a list of things not included. Never invent a service that is not in that list, and never promise an outcome on a third party behalf.'
      || E'\n\n'
      || 'STAT AND FIGURE FIELDS ARE FOR ATTESTED NUMBERS ONLY. A stat, metric, counter or figure field publishes a number as a hard fact, and the claims gate refuses any figure there that does not match an evidence_base fact value. The build duration is HEDGED ("usually ready the next day") and is not a number: never render it as a stat, a counter, or a bare figure such as "1 day" or "24 hours". Say it in prose, with the hedge intact. The same rule holds for anything else stated with "usually", "about", "roughly" or "a few": if the fact hedges, the page hedges, and it never becomes a digit in a stat box. (This instruction exists because a home page rebuild was refused on exactly that: unregistered_stat, "1 day", brief-explanation.stat_2_value, 2026-08-18.)'
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
       'SQL_2026-08-18: assistant voice (next steps and named help, not statement-stacking) + stat-field guard that keeps hedged facts out of figure fields. Facts unchanged.',
       true, 'webdesign_uk_build_service lane, owner brief 2026-08-18', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; wb text; n int;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts');
  IF n <> 15 THEN RAISE EXCEPTION 'fact count changed: % (expected 15) - facts must be untouched here', n; END IF;
  wb := d->>'writer_block';
  IF position('helpful assistant, not a marketing bot' in wb) = 0 THEN RAISE EXCEPTION 'voice brief did not land'; END IF;
  IF position('STAT AND FIGURE FIELDS ARE FOR ATTESTED NUMBERS ONLY' in wb) = 0 THEN RAISE EXCEPTION 'stat guard did not land'; END IF;
  -- the pre-existing wire must survive the append
  IF position('pays before the site is built' in wb) = 0 THEN RAISE EXCEPTION 'the payment sentence was lost'; END IF;
  IF length(wb) < 8000 THEN RAISE EXCEPTION 'writer_block shorter than expected: %', length(wb); END IF;
END $$;

COMMIT;
