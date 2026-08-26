-- 650: customer_access_tokens learns to CARRY the ZIP link's presigned URL, so
-- /d/<token> can be pure DB -> 302 with no credentials anywhere standing.
--
-- WHY (DECISION_2026-08-21b, resolved by the owner 2026-08-21): a presigned URL
-- is capped at 7 days by SigV4 and the customer's window is 30 advertised / 42
-- served, so the CUSTOMER holds a token of OURS and the presign lives server-side
-- against it. zip-deliverer (the storage-enabled spawned agent, DGH-011) produces
-- the presign; it is stored HERE at token-mint time; /d/ redeems the token and
-- 302s to the stored URL while it is fresh. No standing service touches object
-- storage (bugs_open/245 stands).
--
-- Nullable ON PURPOSE: only zip_download rows carry a stored URL; confirm rows
-- never do. The freshness comparison is stored_url_expires_at vs now() at the
-- /d/ handler; a stale row renders the honest "being refreshed" page rather than
-- 302ing to a link that 403s as SignatureDoesNotMatch (the LANDMINES trap: that
-- error reads as broken credentials, not an expired link).
--
-- Safe to apply BEFORE the code rolls (columns unused until then). Guarded:
-- refuses a double apply loudly rather than no-opping.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'customer_access_tokens'
       AND column_name = 'stored_url'
  ) THEN
    RAISE EXCEPTION '650: already applied (customer_access_tokens.stored_url exists)';
  END IF;
END $$;

ALTER TABLE customer_access_tokens
  ADD COLUMN stored_url            text,
  ADD COLUMN stored_url_expires_at timestamptz;

COMMENT ON COLUMN customer_access_tokens.stored_url IS
  'zip_download rows only: the presigned object-store URL this token redeems to. Written at mint (delivery.MintZipToken) and by the refresher; NULL on every other purpose. The URL itself expires (SigV4 <=7 days) independently of the token — see stored_url_expires_at.';
COMMENT ON COLUMN customer_access_tokens.stored_url_expires_at IS
  'When stored_url stops working at the object store. /d/ compares this to now(): fresh -> 302, stale -> honest refresh page + persisted error row (never a redirect to a 403).';

-- Verify block (induced): both columns present, both nullable.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_name='customer_access_tokens'
     AND column_name IN ('stored_url','stored_url_expires_at')
     AND is_nullable='YES';
  IF n <> 2 THEN
    RAISE EXCEPTION '650 verify failed: expected 2 nullable columns, found %', n;
  END IF;
END $$;
