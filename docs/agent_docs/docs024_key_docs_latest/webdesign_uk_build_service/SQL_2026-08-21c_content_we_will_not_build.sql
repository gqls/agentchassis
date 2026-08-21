-- SQL_2026-08-21c — OWNER RULING 2026-08-21: a content policy, and a remedy that
-- is deliberately NOT a refund.
--
-- HIS WORDS: "in our terms and conditions I'd like to add that we don't want to do
-- porn, violence, politics or otherwise distateful sites and if we get those briefs
-- rather than refunds, we reserve the right to change the brief and deliver a site
-- that is within the bounds of respectability within their genre of request."
--
-- WHY THIS GOES TO THE REGISTER AND NOT TO A TERMS PAGE. There IS no terms page:
-- the site has 8 pages and none of them is one (checked 2026-08-21), even though
-- writer_block already tells the writer to point at "the full terms". Building that
-- page is a framework job and a separate one. Meanwhile the register IS the wire:
-- facts[] reaches the chat bot live over the site-facts relay within 5 minutes, and
-- the chat bot is the ONLY intake this business has until Stripe lands. So a
-- customer describing a site we will not build gets the right answer today, from an
-- attested fact rather than from whatever the model would otherwise improvise. The
-- terms page, when it is built, states the same fact.
--
-- THE COLLISION THIS RULING CREATES, HANDLED RATHER THAN PAPERED OVER. `any_site_type`
-- has said since 2026-08-18 that the service "builds any sort of site", with the
-- writer_line "Not just business sites. Any sort of site can be asked for." That is
-- now false as an absolute. Its ORIGINAL point was breadth of PURPOSE (personal,
-- community, project sites, not only business ones) and not content, so the fix is
-- narrow: the claim keeps its meaning and the writer_line stops being unbounded. Two
-- attested facts contradicting each other is exactly what this lane escalated on
-- 2026-08-19 rather than smoothing over, and the same rule applies to a collision a
-- session would be CREATING.
--
-- WHAT THE REMEDY MUST NOT BECOME. "rather than refunds" is the owner ruling out a
-- refund, not mentioning one. `no_refund` is attested and the refund ban is armed and
-- broad, so this fact says what happens INSTEAD ("the order is not cancelled; we amend
-- the brief") and never names refunds at all. Restating the refund position here would
-- also have created a second copy of a term that already lives in one place.
--
-- MEASURED before writing (cmd/claimscan, same engine as the deploy gate, against the
-- live register): the new claim, the new writer_line, an FAQ-shaped paragraph, the
-- amended any_site_type line and a bot-style answer — 0 findings across 5, in a run
-- whose refund control ("a refund is available") fired. A must-pass-only set cannot
-- tell a clean scan from a dead one.
--
-- ⚠ FOUND IN THE SAME RUN, NOT FIXED HERE: "You will be able to approve the site once
-- you have seen it" scans CLEAN. `one_shot_no_approval` is attested and writer_block
-- forbids approval-stage copy, but NOTHING ENFORCES IT. That matters more from today,
-- because the owner has just asked for an internal approval step before the delivery
-- email, and internal steps leak into copy. A ban is the fix, but it must be an
-- OFFER-shape ban or it blocks the denial too (the 2026-08-19 `round of changes`
-- narrowing is the worked precedent). Recorded in NOTES and the handoff.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{facts}', (
        -- (a) narrow any_site_type's writer_line, (b) append the new fact.
        (SELECT jsonb_agg(
           CASE WHEN f.elem->>'id' = 'any_site_type' THEN
             f.elem || jsonb_build_object(
               'writer_line', 'Not just business sites. Personal, community and project sites are all fine, within the content limits.',
               'verified_at', '2026-08-21')
           ELSE f.elem END ORDER BY f.ord)
           FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS f(elem, ord))
        || jsonb_build_array(jsonb_build_object(
             'id',   'content_we_will_not_build',
             'kind', 'attestation',
             'claim', 'We do not build pornographic, violent or political sites, and we do not build sites that are otherwise distasteful. Where a brief asks for one, the order is not cancelled: we reserve the right to amend the brief ourselves and to build a site that stays within the bounds of respectability for whatever the customer does. The customer still receives a finished site and the files that go with it.',
             'writer_line', 'We do not build pornographic, violent or political sites, or anything otherwise distasteful. If a brief asks for one, we amend it and build something respectable in the same line of work.',
             'verified_at', '2026-08-21',
             'source', jsonb_build_object(
               'attested_by', 'owner, 2026-08-21, chat ruling: "we don''t want to do porn, violence, politics or otherwise distateful sites and if we get those briefs rather than refunds, we reserve the right to change the brief and deliver a site that is within the bounds of respectability within their genre of request"'),
             'context_terms', jsonb_build_array('brief','content','respectable','amend','adult','political','violent')
           )))),
      '{writer_block}',
      to_jsonb(
        replace(
          replace(c.data->>'writer_block',
            'Where a page invites a request, say that any sort of site can be asked for, not just a business site.',
            'Where a page invites a request, say that any sort of site can be asked for, not just a business site. There ARE limits and they are stated plainly rather than hinted at: we do not build pornographic, violent or political sites, or anything otherwise distasteful. Say what happens when a brief asks for one, because it is the unusual half and the reader will assume the opposite: the order is not cancelled and no money comes back into it. We reserve the right to amend the brief ourselves and build a site that stays within the bounds of respectability for whatever the customer actually does, and they still get the finished site and its files. Never write this as a threat or a warning, and never imply we vet or approve people. It is a limit on the WORK, stated once, in the same flat voice as the price.'),
          'that any sort of site can be asked for (no example sites are named yet)',
          'that any sort of site can be asked for within the content limits (no example sites are named yet), that we do not build pornographic, violent or political or otherwise distasteful sites and that such a brief is amended rather than cancelled')
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
 'SQL_2026-08-21c: owner ruling 2026-08-21 - content policy (no porn, violence, politics or otherwise distasteful) with the remedy being an amended brief, NOT a refund. New fact content_we_will_not_build; any_site_type writer_line narrowed because "any sort of site" is no longer absolute; writer_block gains the policy and the may-state entry.',
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

  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims'
    THEN RAISE EXCEPTION 'banned_claims changed, they must not (the approval-shape ban is a SEPARATE decision)'; END IF;

  -- Exactly one fact appended, and exactly one existing fact changed.
  IF jsonb_array_length(d->'facts') <> jsonb_array_length(prev->'facts') + 1
    THEN RAISE EXCEPTION 'expected one new fact, count went % -> %',
      jsonb_array_length(prev->'facts'), jsonb_array_length(d->'facts'); END IF;
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'facts') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'facts')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e IS DISTINCT FROM b.e;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 existing fact to change, got %', n; END IF;
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'facts') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'facts')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e IS DISTINCT FROM b.e AND b.e->>'id' <> 'any_site_type';
  IF n <> 0 THEN RAISE EXCEPTION 'a fact other than any_site_type changed'; END IF;

  -- The new fact must carry all four of the owner's exclusions AND the remedy.
  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='content_we_will_not_build';
  IF fact IS NULL THEN RAISE EXCEPTION 'content_we_will_not_build did not land'; END IF;
  IF position('pornographic' in fact->>'claim') = 0
     OR position('violent'      in fact->>'claim') = 0
     OR position('political'    in fact->>'claim') = 0
     OR position('distasteful'  in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'the claim does not carry all four exclusions the owner named'; END IF;
  IF position('not cancelled' in fact->>'claim') = 0 OR position('amend' in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'the claim states the exclusions but not the remedy, which is the half the owner was specific about'; END IF;
  -- "rather than refunds" is the owner RULING ONE OUT. Naming refunds here would
  -- both duplicate no_refund and walk into an armed ban.
  IF fact->>'claim' ILIKE '%refund%' OR fact->>'writer_line' ILIKE '%refund%'
    THEN RAISE EXCEPTION 'the content-policy fact names refunds; it must state the remedy instead'; END IF;

  -- any_site_type must no longer be unbounded, and must keep its original point.
  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='any_site_type';
  IF position('Any sort of site can be asked for' in fact->>'writer_line') > 0
    THEN RAISE EXCEPTION 'any_site_type still tells the writer any sort of site can be asked for, which now contradicts an attested fact'; END IF;
  IF position('Not just business sites' in fact->>'writer_line') = 0
    THEN RAISE EXCEPTION 'any_site_type lost its original point (breadth of PURPOSE, not content)'; END IF;

  -- writer_block: exactly the two intended edits, proven by reconstruction.
  expect := replace(
              replace(pwb,
                'Where a page invites a request, say that any sort of site can be asked for, not just a business site.',
                'Where a page invites a request, say that any sort of site can be asked for, not just a business site. There ARE limits and they are stated plainly rather than hinted at: we do not build pornographic, violent or political sites, or anything otherwise distasteful. Say what happens when a brief asks for one, because it is the unusual half and the reader will assume the opposite: the order is not cancelled and no money comes back into it. We reserve the right to amend the brief ourselves and build a site that stays within the bounds of respectability for whatever the customer actually does, and they still get the finished site and its files. Never write this as a threat or a warning, and never imply we vet or approve people. It is a limit on the WORK, stated once, in the same flat voice as the price.'),
              'that any sort of site can be asked for (no example sites are named yet)',
              'that any sort of site can be asked for within the content limits (no example sites are named yet), that we do not build pornographic, violent or political or otherwise distasteful sites and that such a brief is amended rather than cancelled');
  IF expect IS DISTINCT FROM wb
    THEN RAISE EXCEPTION 'writer_block is not the old text plus exactly the two named edits'; END IF;

  -- And the outcome, asserted independently of how it was reached.
  IF position('pornographic' in wb) = 0
    THEN RAISE EXCEPTION 'writer_block does not carry the content policy, so the writer may not state it'; END IF;
  IF position(fact->>'writer_line' in wb) > 0
    THEN RAISE EXCEPTION 'unexpected: any_site_type writer_line quoted verbatim in writer_block'; END IF;
END $$;

COMMIT;
