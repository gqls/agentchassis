-- SQL_2026-08-18d — narrow the refund ban from the BARE TOKEN to PROMISE SHAPES.
--
-- ⚠ HELD, NOT APPLIED. Four page rewrites (index, what-you-get, faq, how-it-works)
-- were in flight against this register at 11:50Z, driven by the session that now
-- runs both lanes. Editing banned_claims under a running rewrite is the in-place
-- collision the joint handoff warns about. Apply this only when those four are
-- settled, and re-read the register first: another lane edits this row IN PLACE.
--
-- WHY. The live ban is `\brefunds?\b|\brefundable\b|\bmoney.back\b` — the bare word.
-- The joint handoff calls the resulting failures "known coin-flip failures the gate
-- correctly catches (just re-triage the item)". MEASURED 2026-08-18, they are not a
-- coin flip and the gate is not correct: the ban blocks the DENIAL of the very claim
-- it exists to ban.
--
--   Run: platform/orchestration/datahelpers.ScanBannedClaims, live pattern,
--   12 natural ways to state the owner's no-refunds position.
--   Result: 8 of 12 BLOCKED. Only "we do not offer refunds" / "we don't offer
--   refunds" survive.  Blocked included:
--       "Refunds are not available."          "There are no refunds."
--       "Refunds are not offered once ..."    "No refunds."
--       "The price is non-refundable."        "Do you offer refunds? No."
--
-- THE MECHANISM (claims.go, NegationGuard.NegatedAt): the guard scans BACKWARDS
-- from the matched token, within the clause, for a cue. So a cue that FOLLOWS the
-- token never suppresses — "refunds are NOT available" reads as a refund promise —
-- and bare "no"/"non-" are excluded from the cue vocabulary on purpose
-- (documented, pinned by TestBareNoIsAKnownResidualOfTheSharedGuard). That
-- exclusion is a deliberate FLEET-WIDE decision and is NOT touched here: the fix
-- belongs in this site's own pattern, not in a shared guard every site depends on.
--
-- This is the same lesson the register's own writer NOTE already half-records, and
-- the same one bugs_open/161's landmine states for number-bearing claims ("make the
-- pattern require the NUMBER, not the bare phrase"). A policy word has no number,
-- so the promise SHAPE is the equivalent handle.
--
-- WHAT IT COSTS TODAY. The disclosure is squeezed out of the copy: the served index
-- carried "We do not offer refunds" at 10:22Z and carries ZERO occurrences of
-- "refund" now (cache-busted fetch, 2026-08-18 11:42Z). A writer has two survivable
-- phrasings out of twelve; each miss costs a failed rebuild. Whether the home page
-- MUST carry the disclosure is an owner call (consumer-rights) and is NOT decided
-- here — this only makes it sayable.
--
-- VERIFIED BOTH DIRECTIONS before writing this file:
--   * 24 hand-written cases: 0 failures — every denial allowed, all 12 promise
--     shapes still blocked (money-back, full refund, refund is available,
--     refundable deposit, we will refund you, request a refund, ...).
--   * 26 REAL corpus lines (every refund-bearing component in the fleet, 7 sites,
--     nobody wrote them for this test): 0 newly blocked. The 5 retired £1,200-model
--     promises on this site ("walk away and get a full refund", "Full refund until
--     you accept") stay blocked under the new pattern. The 11 now-allowed lines are
--     all OTHER sites' consumer-rights prose ("a refund of the interest and charges"
--     from the Ombudsman guides) — never a promise by the site itself.
--
-- Facts UNCHANGED. Ban COUNT unchanged — one pattern replaced in place, by exact
-- match on the old pattern, every other ban carried through untouched.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{banned_claims}', (
      SELECT jsonb_agg(
               CASE WHEN b->>'pattern' = '\brefunds?\b|\brefundable\b|\bmoney.back\b'
                    THEN jsonb_build_object(
                      'pattern', '\bmoney[ -]?back\b|\b(full|100%|complete|partial|unconditional|no[- ]quibble|guaranteed)\s+refunds?\b|\brefunds?\s+(are|is|will be|can be|may be)\s+(available|offered|given|issued|provided|possible|guaranteed)\b|\b(we|you|customers?|clients?)\s+(will |can |may |shall |always )*(get|receive|claim|request|obtain)\s+(a |an |your |the |full )*refunds?\b|\bwe\s+(will |can |do |shall |always )*(offer|give|issue|provide|process|honour|honor)\s+(a |an |full )*refunds?\b|\b(request|claim|apply for)\s+(a |an |your |the |full )*refunds?\b|\brefunds?\s+(you|your money|in full)\b|(^|[^-\w])refundable\b',
                      'reason', 'RETIRED OFFER (owner ruling 2026-08-11, PLAN 1c.4): no refund is offered, so no page may PROMISE one. Narrowed 2026-08-18 from the bare word to promise shapes: the bare token also blocked the DENIAL (8 of 12 natural phrasings, measured), because the negation guard only scans BACKWARDS within the clause - so "refunds are not available", "there are no refunds" and "non-refundable" all read as promises. You may now state the position in any natural wording, and you may point at the full terms. What stays banned is offering, giving, guaranteeing or inviting a refund.')
                    ELSE b END
               ORDER BY ord)
      FROM jsonb_array_elements(c.data->'banned_claims') WITH ORDINALITY AS t(b, ord)
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'lane-fix',
  'SQL_2026-08-18d: refund ban narrowed from the bare token to promise shapes, so the no-refunds position can be STATED. Facts unchanged, ban count unchanged.',
  true,'webdesign_uk_build_service lane, measured 2026-08-18', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; nb_old int; nb_new int; n int;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;

  -- Nothing lost: no fixed fact count (another lane edits this row in place), so
  -- assert the known ids each appear exactly once.
  SELECT count(*) INTO n FROM unnest(ARRAY['price_total','price_is_total_no_vat','ai_built','build_duration','no_changes_included','no_lock_in','taking_it_further','yours_to_change','queue_limited','contact','third_party_options','payment_upfront','no_refund','delivery_preview_and_zip','one_shot_no_approval','starter_site_initial_copy','hosted_under_our_domain','domain_rent_monthly','domain_buy_once','any_site_type_examples','no_presales_service']) k
   WHERE (SELECT count(*) FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'=k) <> 1;
  IF n <> 0 THEN RAISE EXCEPTION '% known fact(s) lost or duplicated by this write', n; END IF;

  -- The bare-token ban must be GONE and the promise-shape ban PRESENT.
  SELECT count(*) INTO nb_old FROM jsonb_array_elements(d->'banned_claims') b
   WHERE b->>'pattern' = '\brefunds?\b|\brefundable\b|\bmoney.back\b';
  IF nb_old <> 0 THEN RAISE EXCEPTION 'the bare-token refund ban is still present - the exact-match replace did not fire (has another lane rewritten this pattern?)'; END IF;

  SELECT count(*) INTO nb_new FROM jsonb_array_elements(d->'banned_claims') b
   WHERE b->>'pattern' LIKE '%money[ -]?back%' AND b->>'pattern' LIKE '%refundable%';
  IF nb_new <> 1 THEN RAISE EXCEPTION 'expected exactly 1 promise-shape refund ban, found %', nb_new; END IF;

  -- Every other ban carried through. 31 was the count before this write; this
  -- replaces one in place, so the total must not move.
  SELECT count(*) INTO nb_new FROM jsonb_array_elements(d->'banned_claims');
  IF nb_new < 31 THEN RAISE EXCEPTION 'banned_claims shrank to % - the other bans must be preserved', nb_new; END IF;
END $$;

COMMIT;
