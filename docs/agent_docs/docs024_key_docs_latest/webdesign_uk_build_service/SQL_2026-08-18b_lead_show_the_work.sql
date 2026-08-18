-- SQL_2026-08-18b — the page LEAD, ruled by the owner 2026-08-18: proposal F,
-- "show the work, promise nothing".
--
-- THE ONE JUDGEMENT IN THIS FILE. The lead says "see real sites built with this system, and
-- the prompt that made each one" — and those example sites DO NOT EXIST YET. The owner:
-- "The site examples will happen once the system is working and I use it." So encoding the
-- lead naively would have writers point at a gallery that is not there: copy promising a
-- destination that does not exist is exactly bugs_open/299's class, on the page that sells
-- the product. The instruction below therefore carries its own precondition — show only what
-- exists, and note the one example that ALREADY does: this site was built by the system it
-- sells (attested by fact ai_built). When the owner adds his own domains, they slot into the
-- same instruction with no further change.
--
-- Also lands the superlative ban the owner asked for on 2026-08-17 ("don't claim anything
-- untenable"), because a cheaper, less-service sibling brand is coming FROM HIM: any
-- market-wide superlative is falsified by his own next launch. Patterns are deliberately
-- narrow — banned_claims are BLOCKERS, so a loose pattern stops a page for no reason.
--
-- Facts UNCHANGED. writer_block appended by anchor; banned_claims appended, never rewritten.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{writer_block}', to_jsonb(
        c.data->>'writer_block' || E'\n\n'
        || 'HOW THE SITE LEADS (owner ruling, 2026-08-18). Lead by showing the work, not by promising an outcome. The strongest thing this offer has is that the result is inspectable. Name the real sites built by this system, which fact any_site_type_examples attests, and say plainly that this site was built the same way, which fact ai_built attests. Never promise what THEIR site will contain: every build differs and a specification is a hostage. ONE PRECONDITION, absolute: reference only what actually exists and is reachable. The named example sites exist and may be named; a GALLERY showing each prompt beside its result does NOT exist yet, so do not write "see our examples", do not link an examples or portfolio page, and do not imply a showcase, until one is live. When the owner adds prompt-and-result pairs, they belong here in that shape.'
        || E'\n\n'
        || 'NEVER CLAIM A PLACE IN THE MARKET. No "cheapest", no "best value", no "no one does less", no "the simplest offer anywhere", no comparison against other providers or tools by name or implication. Two reasons, both binding: a cheaper and even less serviced sibling brand is planned by the same owner, so any market-wide superlative is falsified by his own next launch; and a claim about someone else capability cannot be verified by this platform checker and dates badly. Absolute statements about THIS product are fine and are the whole voice: the price, that it is the total, that no changes are included, that there are no refunds, that the files are theirs. State those flatly and let the reader draw the comparison.'
      )),
      '{banned_claims}',
      COALESCE(c.data->'banned_claims','[]'::jsonb) || jsonb_build_array(
        jsonb_build_object(
          'pattern', '\b(cheapest|lowest priced?|best value|unbeatable)\b',
          'reason', 'MARKET SUPERLATIVE (owner ruling 2026-08-17): a cheaper, less-serviced sibling brand is planned by the same owner, so any market-wide superlative is falsified by his own next launch. Absolute claims about this product are fine; comparative ones are not.'),
        jsonb_build_object(
          'pattern', '\bno ?(-|\s)?one (does|offers|builds) (it )?(cheaper|less|simpler|for less)\b',
          'reason', 'MARKET SUPERLATIVE (owner ruling 2026-08-17): same reason - the owner will undercut this himself.')
      )
    ) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
  'SQL_2026-08-18b: the LEAD (show the work, promise nothing) with its examples-must-exist precondition, plus two narrow market-superlative bans. Facts unchanged.',
  true,'webdesign_uk_build_service lane, owner ruling 2026-08-18', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; wb text; n int; nb int;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  -- NOT a fixed count: another lane edits this row IN PLACE (7 facts appeared between my
  -- 10:23 supersede and this one, and hosting_and_domain_not_included was RETIRED by them
  -- when hosting/domain became paid options). Assert instead that nothing is lost by THIS write.
  SELECT count(*) INTO n FROM unnest(ARRAY['price_total','price_is_total_no_vat','ai_built','build_duration','no_changes_included','no_lock_in','taking_it_further','yours_to_change','queue_limited','contact','third_party_options','payment_upfront','no_refund','delivery_preview_and_zip','one_shot_no_approval','starter_site_initial_copy','hosted_under_our_domain','domain_rent_monthly','domain_buy_once','any_site_type_examples','no_presales_service']) k
   WHERE (SELECT count(*) FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'=k) <> 1;
  IF n <> 0 THEN RAISE EXCEPTION '% known fact(s) lost or duplicated by this write', n; END IF;
  wb := d->>'writer_block';
  IF position('Lead by showing the work' in wb)=0 THEN RAISE EXCEPTION 'lead instruction did not land'; END IF;
  IF position('NEVER CLAIM A PLACE IN THE MARKET' in wb)=0 THEN RAISE EXCEPTION 'superlative rule did not land'; END IF;
  IF position('reference only what actually exists and is reachable' in wb)=0 THEN RAISE EXCEPTION 'the examples precondition did not land - without it the writers point at a gallery that does not exist'; END IF;
  -- earlier wires must survive
  IF position('pays before the site is built' in wb)=0 THEN RAISE EXCEPTION 'payment sentence lost'; END IF;
  IF position('helpful assistant, not a marketing bot' in wb)=0 THEN RAISE EXCEPTION 'voice brief lost'; END IF;
  IF position('STAT AND FIGURE FIELDS' in wb)=0 THEN RAISE EXCEPTION 'stat guard lost'; END IF;
  SELECT count(*) INTO nb FROM jsonb_array_elements(d->'banned_claims');
  IF nb < 31 THEN RAISE EXCEPTION 'banned_claims shrank to % - the existing bans must be preserved', nb; END IF;
END $$;

COMMIT;
