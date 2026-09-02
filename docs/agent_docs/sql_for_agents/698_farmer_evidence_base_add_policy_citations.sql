-- 698_farmer_evidence_base_add_policy_citations.sql
--
-- Adds the POLICY half to farmerinsurance.uk's evidence register. Unlike
-- lendzy (695) and loanzy (697), farmer ALREADY has a current evidence_base —
-- news-machinery entity facts (Aon/USI etc.) — so this migration SUPERSEDES
-- (never edits) the current row with a successor carrying the existing facts
-- UNCHANGED plus FOUR policy citations, appended. Owner instruction 2026-09-02
-- via the lendzy lane; method = lendzy RUNBOOK §8; RFC_060 §4.
--
-- Farmer is INSURANCE, not consumer credit, so the governing sources are
-- ICOBS / DISP / statute rather than CONC — the method held unchanged (the
-- insurance-shaped answer the lendzy lane asked about): ICOBS 8.1.1 R (claims
-- handled promptly and fairly - the site's "What your insurer is required to
-- do" claim), DISP 1.6.2 R (eight-week final response - the site's complaints
-- clock claim), DISP 3.6.6 R (an accepted ombudsman determination is final and
-- binding - the site's "binding once you accept" claim), and the Employers'
-- Liability (Compulsory Insurance) Regulations 1998 reg 3 (the £5 million
-- minimum - the site's "law sets a minimum level of cover of £5 million").
--
-- EVERY QUOTE VERIFIED THROUGH THE PRODUCTION MATCHER (cmd/fcaquotecheck) on
-- 2026-09-02, each with a deliberately-absent control returning false in the
-- same run. Every URL confirmed by <title>, never by status (the handbook
-- 200s every path - LANDMINES.md).
--
-- The `rule` fields are HUMAN-VERIFIED (checked against each rule's own
-- heading 2026-09-02) - the handbook has no rule-level URL, so the daily
-- refresher keeps the QUOTES honest, not the rule ids (RFC_060 §3d/Q6).
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
-- Rollback: 698_..._ROLLBACK.sql
\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int; npolicy int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '99cae989-2413-430d-b026-59dfeeb638c0' AND aspect = 'evidence_base' AND is_current;
  IF n <> 1 THEN
    RAISE EXCEPTION '698 ABORT: expected exactly 1 current farmer evidence_base row, found % - re-read before writing', n;
  END IF;
  SELECT count(*) INTO npolicy FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '99cae989-2413-430d-b026-59dfeeb638c0' AND s.aspect='evidence_base' AND s.is_current
     AND f->>'kind' = 'policy';
  IF npolicy <> 0 THEN
    RAISE EXCEPTION '698 ABORT: farmer already has % policy fact(s) - another session got here first', npolicy;
  END IF;
END $$;

-- Supersede-and-merge: successor = current row's facts || the four new policy facts.
WITH cur AS (
  UPDATE site_specs SET is_current = false, superseded_at = now()
   WHERE site_id = '99cae989-2413-430d-b026-59dfeeb638c0' AND aspect = 'evidence_base' AND is_current
   RETURNING data
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT
  '99cae989-2413-430d-b026-59dfeeb638c0',
  'evidence_base',
  jsonb_set(cur.data, '{facts}', (cur.data->'facts') || '[
    {"id": "FCA-ICOBS-8-1-1", "kind": "policy", "rule": "ICOBS 8.1.1", "claim": "An FCA rule requires an insurer to handle claims promptly and fairly, and not to unreasonably reject a claim.", "writer_line": "Insurers regulated by the Financial Conduct Authority (FCA) must handle claims promptly, fairly, and without imposing unreasonable barriers, under the FCA''s rules on claims handling.", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/ICOBS/8/1.html", "quote": "handle claims promptly and fairly", "title": "FCA Handbook - ICOBS 8.1 Insurers: general", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "FCA-DISP-1-6-2", "kind": "policy", "rule": "DISP 1.6.2", "claim": "An FCA rule requires a firm to send a final response to a complaint by the end of eight weeks after receiving it, or explain why it cannot yet do so.", "writer_line": "they have eight weeks to issue a final response, or explain why they need longer", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/DISP/1/6.html", "quote": "by the end of eight weeks after its receipt of the complaint", "title": "FCA Handbook - DISP 1.6 Complaints time limit rules", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "FCA-DISP-3-6-6", "kind": "policy", "rule": "DISP 3.6.6", "claim": "An FCA rule makes an ombudsman determination final and binding on both parties if the complainant accepts it within the time limit.", "writer_line": "its decisions are binding on the firm once you accept them", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/DISP/3/6.html", "quote": "accepts the determination within that time limit, it is final and binding on both parties", "title": "FCA Handbook - DISP 3.6 Determination by the Ombudsman", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "LAW-ELCI-1998-R3", "kind": "policy", "rule": "Employers'' Liability (Compulsory Insurance) Regulations 1998, reg 3", "claim": "The law requires employers'' liability insurance cover of not less than 5 million pounds in respect of a claim arising from any one occurrence.", "writer_line": "The law sets a minimum level of cover of £5 million", "source": {"citation": {"url": "https://www.legislation.gov.uk/uksi/1998/2573/regulation/3", "quote": "not less than £5 million", "title": "The Employers'' Liability (Compulsory Insurance) Regulations 1998 - Regulation 3", "publisher": "legislation.gov.uk", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}
  ]'::jsonb),
  'manual',
  NULL,
  'loanzy_uk_example_site lane (migration 698)',
  true,
  true,
  'Policy half added to the existing news-entity register: ICOBS 8.1.1, DISP 1.6.2, DISP 3.6.6, ELCI Regs 1998 reg 3 - the four regulatory claims farmer''s served pages state. Existing entity facts carried forward UNCHANGED (supersede-and-merge, never edit). Quotes verified via cmd/fcaquotecheck with negative controls, 2026-09-02; handbook URLs title-confirmed.'
FROM cur;

DO $$
DECLARE nfacts int; npolicy int; nentity int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = '99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base' AND is_current;
  SELECT count(*) INTO npolicy FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '99cae989-2413-430d-b026-59dfeeb638c0' AND s.aspect='evidence_base' AND s.is_current
     AND f->>'kind'='policy' AND length(f->'source'->'citation'->>'quote') > 15;
  SELECT count(*) INTO nentity FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '99cae989-2413-430d-b026-59dfeeb638c0' AND s.aspect='evidence_base' AND s.is_current
     AND f->>'kind'='entity';
  IF npolicy <> 4 THEN RAISE EXCEPTION '698 VERIFY: expected 4 substantively-quoted policy facts, found %', npolicy; END IF;
  IF nentity < 1 THEN RAISE EXCEPTION '698 VERIFY: the pre-existing entity facts were LOST (found %)', nentity; END IF;
  IF nfacts <> npolicy + nentity THEN RAISE EXCEPTION '698 VERIFY: fact count % != policy % + entity %', nfacts, npolicy, nentity; END IF;
  RAISE NOTICE '698 OK: farmer register now carries % facts (% entity carried forward + 4 policy)', nfacts, nentity;
END $$;
COMMIT;
