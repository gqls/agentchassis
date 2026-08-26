-- SQL_2026-08-26d — the offer-shape approval-language ban, ARMED.
-- Gap + proof: NOTES 2026-08-26; probe artefacts in the session scratchpad; the ban
-- object carries its own proof summary. Facts and writer_block untouched.
BEGIN;
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{banned_claims}', (c.data->'banned_claims') || $ban$[{"pattern": "\\byou(?:'ll| will| can| may| could)(?:(?: then| also| still)?(?: get to| be able to))? (?:approve|sign[- ]?off|sign it off)\\b|\\bfor your (?:approval|sign[- ]?off)\\b|\\b(?:awaiting|pending|once) your approval\\b|\\bsend (?:it|the site|the design)(?: to you)? for approval\\b", "reason": "OFFER-SHAPE approval-language ban (owner rulings 2026-08-18 retired flow + 2026-08-26 internal-only edit step): an internal owner edit/review step now EXISTS and must never leak into copy - as far as the customer is concerned the product is one-shot with no approval stage. Bans PROMISE shapes (you can/will be able to approve, for your approval, sign off) while denials (there is no approval stage) pass, per the 2026-08-19 round-of-changes narrowing precedent. Proven 2026-08-26 with a both-halves claimscan probe set: baseline 0/9 (the gap), candidate 5 blocks incl. the known evader 'You will be able to approve the site once you have seen it', 4 denials/live-copy passes."}]$ban$::jsonb) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-26d: offer-shape approval-language ban armed (internal edit step must never leak); claimscan both-halves proven.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;
DO $chk$
DECLARE d jsonb; prev jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF jsonb_array_length(d->'banned_claims') <> jsonb_array_length(prev->'banned_claims')+1 THEN RAISE EXCEPTION 'not exactly one ban added'; END IF;
  IF NOT EXISTS (SELECT 1 FROM jsonb_array_elements(d->'banned_claims') b WHERE b->>'pattern' LIKE '%be able to%approve%') THEN RAISE EXCEPTION 'new ban absent'; END IF;
  IF d->'facts' <> prev->'facts' THEN RAISE EXCEPTION 'facts moved'; END IF;
  IF d->>'writer_block' <> prev->>'writer_block' THEN RAISE EXCEPTION 'writer_block moved'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: one ban armed';
END $chk$;
COMMIT;
