-- SQL_2026-08-19g — OWNER RULING 2026-08-19 (evening): the £200 buys the DOMAIN
-- ONLY, and the customer transfers it themselves.
--
-- THE QUESTION THIS SETTLES. This morning the owner ruled *"I don't want to be
-- their registrar if they buy the domain. They still have to buy it though so
-- they'll need to provide details about their registrar and we'll have to
-- document and probably hand hold the transfer process for them."* The
-- hand-holding half collided head-on with an attested fact: `no_presales_service`
-- says the price stays down because *"nobody's time is included … never offer the
-- owner's time"*. Two attested facts contradicting each other is not something a
-- session may paper over in copy, so both were left untouched and the collision
-- was put to the owner.
--
-- HIS RULING, 2026-08-19 evening: **£200 = domain only, they transfer it.**
-- "Keep nobody's-time-included absolute. We document the steps and hand over what
-- they need; the transfer itself is theirs to do." So:
--   * `no_presales_service` is UNCHANGED and stays absolute. It is not narrowed,
--     and there is no carve-out for the £200. This resolves the collision in its
--     favour rather than around it.
--   * `domain_buy_once` is re-attested: the domain must move to their own
--     registrar (we do not stay their registrar), we give them what they need,
--     and arranging it is theirs to do.
--
-- WHAT WAS WRONG WITH THE OLD WORDING, and it was wrong twice over. It read
-- *"the customer is then free to transfer it to their own registrar or host."*
--   1. "free to" is an OPTION. The owner's position is an OBLIGATION: we are not
--      staying on as their registrar, so it has to move.
--   2. It generated banned copy. At 15:56Z today the Website Brief Starter guide
--      was blocked on "whenever you like" in *"a one-off £200, after which you're
--      free to move it to your own registrar whenever you like"* - the writer
--      elaborating an optional-sounding freedom into an unbounded one. That is
--      the second page-blocking this fact's wording has caused today.
--
-- ⚠ ONE MECHANICAL WRINKLE, RECORDED RATHER THAN WRITTEN INTO COPY. For a `.uk`
-- domain the transfer is executed by the LOSING registrar changing the IPS TAG,
-- so the final action is ours, not the customer's, however the commercial terms
-- are worded. The new claim is written to survive that ("we give them what they
-- need to move it") without promising anyone's time and without describing a
-- mechanism nobody has documented yet. Which TLDs we actually sell is still
-- unowned and still blocks documenting this properly - see the handoff's STILL
-- OPEN item 5.
--
-- MEASURED: the new claim, the new writer_line and a natural paragraph built
-- from them scan clean against the live register (cmd/claimscan, 0 findings
-- across 3), so this wording cannot repeat what the old one did.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{facts}', (
      SELECT jsonb_agg(
        CASE WHEN f.elem->>'id' = 'domain_buy_once' THEN
          f.elem
            || jsonb_build_object(
                 'claim', 'Buying the domain outright is a one-off £200. The domain is then the customer''s, and it must be moved to their own registrar: we do not stay their registrar. We give them what they need to move it; arranging the transfer with their new registrar is theirs to do, and no support time is included in the price.',
                 'writer_line', 'Buying the domain is a one-off £200. It is then yours, and you move it to your own registrar; we give you what you need to do that.',
                 'verified_at', '2026-08-19',
                 'source', jsonb_build_object('attested_by', 'owner, 2026-08-19, chat ruling: the £200 buys the domain only, the customer transfers it, and nobody''s time is included'),
                 'context_terms', jsonb_build_array('domain','buy','one-off','transfer','registrar','move','own'))
        ELSE f.elem END ORDER BY f.ord)
        FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS f(elem, ord)
    )) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
 'SQL_2026-08-19g: owner ruling 2026-08-19 - the £200 buys the domain only and the customer transfers it. domain_buy_once re-attested; no_presales_service deliberately unchanged and still absolute.',
 true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; fact jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed, they must not'; END IF;
  IF prev->>'writer_block' IS DISTINCT FROM d->>'writer_block' THEN RAISE EXCEPTION 'writer_block changed, it must not'; END IF;
  IF jsonb_array_length(prev->'facts') <> jsonb_array_length(d->'facts') THEN RAISE EXCEPTION 'fact count moved'; END IF;

  -- EXACTLY ONE fact may differ, and it must be the one named.
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'facts') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'facts')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e IS DISTINCT FROM b.e;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 fact to change, got %', n; END IF;
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'facts') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'facts')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e IS DISTINCT FROM b.e AND b.e->>'id' <> 'domain_buy_once';
  IF n <> 0 THEN RAISE EXCEPTION 'a fact other than domain_buy_once changed'; END IF;

  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='domain_buy_once';
  IF (fact->>'value')::numeric <> 200 THEN RAISE EXCEPTION 'the £200 figure moved'; END IF;
  IF position('free to transfer' in fact->>'claim') > 0 THEN RAISE EXCEPTION 'the option wording survives in the claim'; END IF;
  IF position('free to transfer' in fact->>'writer_line') > 0 THEN RAISE EXCEPTION 'the option wording survives in the writer_line'; END IF;
  IF position('must be moved to their own registrar' in fact->>'claim') = 0 THEN RAISE EXCEPTION 'the obligation did not land'; END IF;

  -- no_presales_service must be untouched and still absolute: that is the ruling.
  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='no_presales_service';
  IF position('nobody''s time is included' in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'no_presales_service lost its absolute clause'; END IF;
END $$;

COMMIT;
