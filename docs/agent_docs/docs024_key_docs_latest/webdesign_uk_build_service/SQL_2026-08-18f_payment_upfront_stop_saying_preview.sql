-- SQL_2026-08-18f — the LAST retired-vocabulary leftover in evidence_base:
-- fact `payment_upfront` still calls the post-payment link a "preview link".
--
-- The owner ruled 2026-08-18 (recorded by the site_delivery_and_editor lane in the
-- joint handoff) that the post-payment link is NEVER called a "preview". Fact
-- `delivery_live_link_and_zip` already words it correctly ("a link to their site
-- already live"); `payment_upfront` contradicts it in the same register. The claim
-- is what the chat bot renders verbatim and what writers are handed, so this is
-- live customer-facing wording, not an internal note.
--
-- LEFT ALONE DELIBERATELY, and checked one by one: `price_total` and
-- `build_duration` mention £1,200 ONLY inside source.attested_by, as provenance
-- ("supersedes the £1,200 price attested by the owner on 2026-08-03"), and the
-- writer_block mentions it only to say the deposit and fourteen-day window were
-- retired with it. That is the audit trail of what superseded what. Stripping it
-- would destroy the record while changing no customer-facing word.
--
-- Guarded the way this register requires: another lane edits this row IN PLACE, so
-- assert against the row THIS transaction supersedes rather than any fixed count.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{facts}', (
      SELECT jsonb_agg(
               CASE WHEN f->>'id' = 'payment_upfront'
                    THEN jsonb_set(f, '{claim}', to_jsonb(
                      'Payment is taken before the site is built. The customer does not see the site before paying; the link to their live site and the files come after payment.'::text))
                    ELSE f END ORDER BY ord)
      FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
  'SQL_2026-08-18f: payment_upfront stops calling the post-payment link a preview. Facts and bans otherwise unchanged.',
  true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  -- nothing lost by THIS write
  SELECT count(*) INTO n FROM (
    (SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f)
    UNION ALL
    (SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f)) x;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) differ - this write changes one claim only', n; END IF;
  IF jsonb_array_length(prev->'banned_claims') <> jsonb_array_length(d->'banned_claims')
    THEN RAISE EXCEPTION 'banned_claims count moved'; END IF;
  IF (prev->>'writer_block') IS DISTINCT FROM (d->>'writer_block')
    THEN RAISE EXCEPTION 'writer_block changed - it must not'; END IF;

  -- the fix landed, and no FACT CLAIM anywhere says "preview link" any more
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts') f
   WHERE f->>'claim' ~* 'preview link';
  IF n <> 0 THEN RAISE EXCEPTION '% fact claim(s) still say "preview link"', n; END IF;
END $$;

COMMIT;
