-- SQL_2026-08-21 — OWNER RULING 2026-08-21: "we only sell .co.uk and .uk tlds
-- for now."
--
-- WHAT THIS SETTLES. The joint handoff's STILL OPEN item 6 ("Which TLDs do we
-- actually sell?") has been open since 2026-08-19 and was flagged there as
-- unowned AND load-bearing. It is now owned. This attests the scope at the
-- register, which is the wire: the chat bot reads facts[] live from the
-- site-facts relay and refreshes every 5 minutes, so this reaches the only live
-- customer-facing surface without a build, a deploy or a page rebuild.
--
-- WHY IT NEEDED A FACT AND NOT NOTHING. The bot ASKS about the domain: its
-- conduct says "Ask what the site is for and what domain they would want it on"
-- (box/chat-service/facts.go, promptConduct). Asked "Can I have a .com domain
-- for my site, or do you only do .co.uk?" at 2026-08-21 it answered *"the sites
-- are built on domains we provide, and right now those are on .uk domains"* —
-- which is roughly right and entirely UNGROUNDED. No fact attested it, so the
-- model was improvising a commercial term, which is the one thing this register
-- exists to stop (governing_rule: every commercial term must trace to a fact).
-- Improvising the right answer is not the same as being unable to improvise the
-- wrong one.
--
-- WHAT THE FACT DOES NOT SAY, deliberately:
--   * It does not restate £10 or £200. Those are attested once each, in
--     domain_rent_monthly and domain_buy_once. A second copy is a second thing
--     to move when a price moves, and this lane has already been bitten by
--     duplicated copies of the same term (submission / content_direction).
--   * It does not say "for now" in the claim. The owner's "for now" is recorded
--     in source.attested_by, where it belongs: on a page, "for now" invites
--     "when will you add .com?", which is a question nobody may answer because
--     no pre-sales service is included. The scope is stated flatly and can be
--     re-attested when it changes, which is how every other term here works.
--   * It does not describe the .uk transfer-out mechanism. That is a separate,
--     still-open question, and it got SHARPER today rather than softer: see the
--     note at the bottom of this file.
--
-- MEASURED, not asserted (cmd/claimscan, the same engine as the deploy gate,
-- against the live register exported 2026-08-21):
--   * the new claim, the new writer_line, a natural page paragraph and a
--     bot-style answer built from them: 0 findings across 4.
--   * with three CONTROLS in the same run, all three of which fired, so the
--     scanner was demonstrably live on this corpus: "Your site will be ready
--     the next day" (armed next-day ban), "We handle the domain setup for you"
--     (retired-offer ban), and the live writer_block's own turnaround
--     instruction. A must-pass-only probe set cannot tell a clean scan from a
--     dead one.
--
-- writer_block is edited BY ANCHOR in two places, and the verify block proves
-- EXACTLY those two edits by reconstructing the new text from the old.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{facts}', (c.data->'facts') || jsonb_build_array(
        jsonb_build_object(
          'id',   'domain_tlds_offered',
          'kind', 'attestation',
          'claim', 'The domain endings on offer are .co.uk and .uk. Those are the only ones we register, rent or sell. We do not supply .com or any other ending. A customer who wants a different ending registers it with a registrar themselves and puts the delivered site files on it: the files work under any domain.',
          'writer_line', 'The domains we provide are .co.uk and .uk. We do not supply .com or any other ending. If you want one, register it yourself and put your site files on it.',
          'verified_at', '2026-08-21',
          'source', jsonb_build_object(
            'attested_by', 'owner, 2026-08-21, chat ruling: "we only sell .co.uk and .uk tlds for now". The "for now" is recorded here rather than in the claim: the scope is stated flatly on the page and re-attested if it changes.'),
          'context_terms', jsonb_build_array('domain','co.uk','uk','com','ending','tld','registrar')
        ))),
      '{writer_block}',
      to_jsonb(
        replace(
          replace(c.data->>'writer_block',
            'The ZIP means they can equally host it anywhere themselves.',
            'The domains we provide are .co.uk and .uk, and we do not supply .com or any other ending: anyone who wants a different ending registers it with a registrar themselves and puts the site files on it. The ZIP means they can equally host it anywhere themselves.'),
          'for a one-off £200 and then transferred freely, that there is no approval stage',
          'for a one-off £200 and then transferred freely, that the domain endings we provide are .co.uk and .uk and no others, that there is no approval stage')
      )
    ) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
 'SQL_2026-08-21: owner ruling 2026-08-21 - we only sell .co.uk and .uk TLDs for now. New fact domain_tlds_offered; writer_block gains the scope in the domain paragraph and in the may-state list. Closes joint handoff STILL OPEN item 6.',
 true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; fact jsonb; wb text; pwb text; expect text;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  wb := d->>'writer_block'; pwb := prev->>'writer_block';

  -- Bans must not move. Two lanes write this row, so never assert an absolute
  -- count: compare against the row this transaction superseded.
  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims'
    THEN RAISE EXCEPTION 'banned_claims changed, they must not'; END IF;

  -- Facts: exactly one APPENDED, none of the existing ones touched.
  IF jsonb_array_length(d->'facts') <> jsonb_array_length(prev->'facts') + 1
    THEN RAISE EXCEPTION 'expected exactly one new fact, count went % -> %',
      jsonb_array_length(prev->'facts'), jsonb_array_length(d->'facts'); END IF;
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'facts') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'facts')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e IS DISTINCT FROM b.e;
  IF n <> 0 THEN RAISE EXCEPTION '% existing fact(s) changed, none may', n; END IF;

  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='domain_tlds_offered';
  IF fact IS NULL THEN RAISE EXCEPTION 'domain_tlds_offered did not land'; END IF;
  IF position('.co.uk and .uk' in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'the claim does not name both endings'; END IF;
  IF position('We do not supply .com' in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'the claim lost its denial, and a bare "no" would not survive the negation guard'; END IF;
  IF fact->>'writer_line' IS NULL OR fact->>'writer_line' = ''
    THEN RAISE EXCEPTION 'no writer_line, so the page writer gets nothing'; END IF;

  -- writer_block: EXACTLY the two intended anchor edits and nothing else.
  -- Reconstructing the new text from the old is the only guard that can see an
  -- unintended third edit; asserting the two new substrings are present cannot.
  expect := replace(
              replace(pwb,
                'The ZIP means they can equally host it anywhere themselves.',
                'The domains we provide are .co.uk and .uk, and we do not supply .com or any other ending: anyone who wants a different ending registers it with a registrar themselves and puts the site files on it. The ZIP means they can equally host it anywhere themselves.'),
              'for a one-off £200 and then transferred freely, that there is no approval stage',
              'for a one-off £200 and then transferred freely, that the domain endings we provide are .co.uk and .uk and no others, that there is no approval stage');
  IF expect IS DISTINCT FROM wb
    THEN RAISE EXCEPTION 'writer_block is not the old text plus exactly the two named edits'; END IF;
  IF wb = pwb THEN RAISE EXCEPTION 'writer_block did not change, so the writer may not state the scope'; END IF;

  -- Each anchor existed exactly once before, so neither edit fired twice.
  n := (length(pwb) - length(replace(pwb, 'The ZIP means they can equally host it anywhere themselves.', ''))) / length('The ZIP means they can equally host it anywhere themselves.');
  IF n <> 1 THEN RAISE EXCEPTION 'domain-paragraph anchor occurred % times in the old writer_block, expected 1', n; END IF;
  n := (length(pwb) - length(replace(pwb, 'for a one-off £200 and then transferred freely, that there is no approval stage', ''))) / length('for a one-off £200 and then transferred freely, that there is no approval stage');
  IF n <> 1 THEN RAISE EXCEPTION 'may-state anchor occurred % times in the old writer_block, expected 1', n; END IF;
END $$;

COMMIT;

-- ─────────────────────────────────────────────────────────────────────────────
-- ⚠ WHAT THIS RULING SHARPENS, AND IT IS NOW THE ONLY CASE RATHER THAN ONE OF
--   SEVERAL. Both endings we sell are Nominet endings. SQL_2026-08-19g recorded
--   a mechanical wrinkle "rather than writing it into copy": for a .uk domain
--   the transfer out is executed by the LOSING registrar changing the IPS TAG,
--   so the final action is ours however the commercial terms are worded. That
--   was recorded when the TLD scope was unknown and the wrinkle might have
--   applied to some sales. With the scope closed to .co.uk and .uk it applies to
--   EVERY domain we sell, and it sits against two attested facts:
--     * domain_buy_once  — "arranging the transfer with their new registrar is
--       theirs to do, and no support time is included in the price"
--     * no_presales_service — "nobody's time is included", absolute, and the
--       owner resolved 2026-08-19's collision in its favour rather than around it
--   NOT ENCODED, NOT PAPERED OVER, and no copy changed for it: this is an owner
--   question, exactly like the one 19g escalated. The unverified half is who can
--   actually execute the TAG change (the registrant through Nominet's own online
--   services, or only us as the losing registrar) — nobody in this lane has
--   checked, and it decides whether "the transfer is theirs to do" is mechanically
--   true or merely commercially stated. See the handoff, and NOTES 2026-08-21.
