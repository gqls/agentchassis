-- 716_loanzy_add_cca_s77_agreement_copy_fact.sql
--
-- Appends ONE fact to loanzy.uk's register: CCA 1974 s.77, the debtor's right
-- to a copy of the executed fixed-sum credit agreement (+ statement of account)
-- for a £1 fee. Handed by the copy_quality lane from the owner-ruled get-help
-- evidence pass (their item bd03c2b3); quote verified through the production
-- matcher with a control before writing.
--
-- THE SECOND FACT THEY OFFERED IS DELIBERATELY NOT REGISTERED, and the reason
-- is a new finding for the register programme: National Debtline's provenance
-- (a charity run by the Money Advice Trust; free) is TRUE — but BOTH natural
-- hosts (nationaldebtline.org, moneyadvicetrust.org) serve full content to
-- curl while serving NOTHING to the production fetcher's User-Agent
-- ("Mozilla/5.0"): even the single word "free" fails QuoteFoundInText there.
-- That is a THIRD unregistrable-host signature — distinct from the Cloudflare
-- challenge page (maps.org.uk class), it passes every curl-side control (real
-- title, full size, quotes present) and only the write-time production probe
-- (cmd/fcaquotecheck) catches it. A citation there would be citation_lost
-- every day for ever. RFC_060's rule applies: do not add facts you cannot
-- re-check. Relayed to the FCA-mirror design (lendzy lane).
--
-- Also corrected en route (relay-vs-page): the page says National Debtline is
-- "a charity run by the Money Advice Trust", not "a debt advice service run
-- by"; the "We never charge for our support" sentence is not on the about-us
-- page at all. Reported to the copy lane for their bd03c2b3 evidence note.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
-- Rollback: 716_..._ROLLBACK.sql
\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int; nf int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '716 ABORT: expected exactly 1 current loanzy register, found %', n; END IF;
  SELECT count(*) INTO nf FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND s.aspect='evidence_base' AND s.is_current
     AND f->>'id'='LAW-CCA-77';
  IF nf <> 0 THEN RAISE EXCEPTION '716 ABORT: LAW-CCA-77 already registered'; END IF;
END $$;

WITH cur AS (
  UPDATE site_specs SET is_current=false, superseded_at=now()
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current
   RETURNING data
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT
  '55213ded-03ec-40f7-8fc1-169de05e05c8',
  'evidence_base',
  jsonb_set(cur.data, '{facts}', (cur.data->'facts') || '[
    {"id": "LAW-CCA-77", "kind": "policy", "rule": "Consumer Credit Act 1974 s.77", "claim": "Under a regulated fixed-sum credit agreement, the creditor must give the debtor a copy of the executed agreement and a statement of account within the prescribed period of receiving a written request and a 1 pound fee.", "writer_line": "your right to request a copy of your credit agreement (Consumer Credit Act 1974, s.77)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/77", "quote": "shall give the debtor a copy of the executed agreement", "title": "Consumer Credit Act 1974 - Section 77", "publisher": "legislation.gov.uk", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}
  ]'::jsonb),
  'manual', NULL, 'loanzy_uk_example_site lane (migration 716)', true, true,
  'One fact appended: CCA 1974 s.77 (agreement-copy right, 1 pound fee) from the copy lane''s get-help evidence pass, production-matcher-verified. The National Debtline provenance fact deliberately NOT registered: both natural hosts serve nothing to the production fetcher UA while passing every curl-side check - the third unregistrable-host signature, recorded in the migration header and relayed to the FCA-mirror design. banned_claims untouched.'
FROM cur;

DO $$
DECLARE nfacts int; nb int; ncur int; nnew int;
BEGIN
  SELECT count(*) INTO ncur FROM site_specs
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF ncur <> 1 THEN RAISE EXCEPTION '716 VERIFY: expected exactly 1 current row, found %', ncur; END IF;
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims') INTO nfacts, nb
   FROM site_specs WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF nfacts <> 4 THEN RAISE EXCEPTION '716 VERIFY: expected 4 facts (3 carried + 1 new), found %', nfacts; END IF;
  IF nb <> 5 THEN RAISE EXCEPTION '716 VERIFY: banned_claims were LOST - expected 5, found %', nb; END IF;
  SELECT count(*) INTO nnew FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND s.aspect='evidence_base' AND s.is_current
     AND f->>'id'='LAW-CCA-77' AND length(f->'source'->'citation'->>'quote') > 20;
  IF nnew <> 1 THEN RAISE EXCEPTION '716 VERIFY: LAW-CCA-77 absent or thinly quoted'; END IF;
  RAISE NOTICE '716 OK: loanzy facts = 4 (banned_claims intact at 5)';
END $$;
COMMIT;
