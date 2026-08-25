-- 608: LEDGER BACKFILL — noted.co.uk header CTA override key (already live).
--
-- The write this file records was executed by hand (psql jsonb_set) on
-- 2026-08-17 and is LIVE; it was edit 2 of the council-APPROVED plan
-- 89f3331e-57f4-4f8f-8f58-de6222d17337 (verdict 2026-08-18). Three of that
-- approval's own advisory objections (editquality, guardian, debug_historian)
-- said the same thing: an ad-hoc UPDATE recorded only in NOTES prose has no
-- ledger entry, cannot be replayed, and sets a precedent. This file is the
-- disposition of those advisories: the same write, idempotent, in the ledger.
--
-- Running it against the current database is a NO-OP by design (jsonb_set to
-- the value the row already holds). It exists so "what changed
-- sites.content_data and when, and under what approval" is answerable from
-- the ledger like every other config write on this estate.
--
-- Mechanism (approved plan): render_site_components_action.go applies
-- sites.content_data->>'header_cta_url' / 'header_cta_text' AFTER the header
-- CTA derivation, gated by the SAME ChromeLinkPolicy, so absence = derived
-- behaviour and a stale override degrades instead of shipping a dead button.

DO $$
DECLARE
  n int;
BEGIN
  UPDATE sites
     SET content_data = jsonb_set(COALESCE(content_data, '{}'::jsonb),
                                  '{header_cta_url}',
                                  '"/tools/write/index.html"'::jsonb)
   WHERE domain = 'noted.co.uk';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 noted.co.uk sites row, updated %', n;
  END IF;
END $$;

-- Verify (induced-failure style: this block RAISES rather than SELECTing,
-- because a verify block of bare SELECTs cannot stop a COMMIT):
DO $$
BEGIN
  IF (SELECT content_data->>'header_cta_url' FROM sites WHERE domain='noted.co.uk')
     IS DISTINCT FROM '/tools/write/index.html' THEN
    RAISE EXCEPTION 'header_cta_url is not the approved value';
  END IF;
END $$;
