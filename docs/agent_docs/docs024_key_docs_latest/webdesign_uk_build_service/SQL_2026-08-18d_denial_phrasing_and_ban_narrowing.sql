-- SQL_2026-08-18d: two fixes from the afternoon rewrite failures, all of one
-- family (denials phrased with banned tokens block the page):
--   how-it-works 11:53Z: banned_claim "refund" on "There's no refund once
--     payment's made" (the bare-'no' trap the writer_block already warns of);
--   faq 12:06Z: banned_claim "rounds of changes" on "no rounds of changes
--     afterwards" (same trap, different ban);
--   how-it-works 12:16Z: banned_claim "before you pay" on "nothing is shown
--     to you before you pay" - MY OWN 18b ban, over-broad: it matched a
--     correct payment-first DENIAL. The same mistake the refunds ban makes,
--     made while citing it as the cautionary precedent.
-- Fix 1: narrow the 18b ban to the promise shape (a preview offered before
--   payment), which the old copy actually used and the new copy never will.
-- Fix 2: one consolidated writer_block rule with sanctioned denial
--   phrasings, so the writer stops reaching for bare-'no' + banned-token.

BEGIN;

DO $fix$
DECLARE
  spec_id uuid; wb text; bans jsonb; n int;
BEGIN
  SELECT id, data->>'writer_block', data->'banned_claims'
    INTO spec_id, wb, bans
    FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND aspect='evidence_base' AND is_current;

  -- Fix 1: replace the over-broad pattern in place
  SELECT count(*) INTO n FROM jsonb_array_elements(bans) b
   WHERE b->>'pattern' = '\bbefore you pay\b';
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly one \bbefore you pay\b ban, found % (already fixed or concurrently edited)', n;
  END IF;
  SELECT jsonb_agg(
    CASE WHEN b->>'pattern' = '\bbefore you pay\b' THEN
      jsonb_build_object(
        'reason', 'RETIRED FLOW (owner directive 2026-08-18): a PREVIEW offered before payment is the retired promise. Narrowed 2026-08-18 (was \bbefore you pay\b, which blocked the correct denial "nothing is shown to you before you pay" - how-it-works, 12:16Z).',
        'pattern', '\bpreview[^.]{0,40}before you pay\b')
    ELSE b END)
    INTO bans FROM jsonb_array_elements(bans) b;

  -- Fix 2: the consolidated denial rule, appended to the wire
  wb := wb || E'\n\n' ||
    'A DENIAL THAT USES A BANNED TOKEN STILL BLOCKS THE PAGE. The gate''s negation guard recognises ''do not'', ''never'' and ''cannot'' in the same clause; it does NOT recognise a bare ''no'' or ''nothing'', so a sentence like ''there''s no refund'' or ''no rounds of changes'' reads to the gate as making the claim, and the page is refused (three rewrites were refused on exactly this, 2026-08-18). When stating what is not offered, either use ''do not'' or ''never'' in the same clause as the loaded word, or avoid the word entirely. Sanctioned phrasings, use these: ''We do not offer refunds.'' ''We do not revise the site after it is built.'' ''Nothing is shown until you have paid.'' Never: ''no refund'', ''no revisions'', ''no rounds of changes'', or any bare ''no'' immediately in front of a word the ban list carries.';

  UPDATE site_specs
     SET data = jsonb_set(jsonb_set(data, '{banned_claims}', bans), '{writer_block}', to_jsonb(wb)),
         updated_at = now()
   WHERE id = spec_id;

  -- Read back
  SELECT count(*) INTO n FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') b
   WHERE s.id = spec_id AND b->>'pattern' = '\bbefore you pay\b';
  IF n <> 0 THEN RAISE EXCEPTION 'over-broad ban survived the write'; END IF;
  SELECT count(*) INTO n FROM site_specs s
   WHERE s.id = spec_id AND position('A DENIAL THAT USES A BANNED TOKEN' in s.data->>'writer_block') > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'denial rule missing after write'; END IF;
  RAISE NOTICE '18d applied';
END $fix$;

UPDATE site_work_items SET status='triaged'
 WHERE id IN ('f853f532-ef9f-4951-9d45-27ed0757ae85','8d969047-88a3-4384-8376-6699135e67c7')
   AND status='needs_human_review';

COMMIT;
