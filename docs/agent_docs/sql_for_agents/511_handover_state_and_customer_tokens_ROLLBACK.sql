-- 511 ROLLBACK — run by hand, deliberately. The runner never picks this up.
--
-- ⚠ DROPPING customer_access_tokens DESTROYS EVERY LIVE CUSTOMER LINK. A token's
-- plaintext exists only in an email that has already been sent, so there is no
-- way to reconstruct the rows: every download link and every confirm-transfer
-- link in a customer's inbox becomes dead, silently, from their side. Check
-- before you run it:
--   SELECT purpose, count(*) FROM customer_access_tokens
--    WHERE revoked_at IS NULL AND expires_at > now() GROUP BY 1;
-- and refuse yourself if that is non-zero unless you mean it.
--
-- Dropping the sites columns loses the handover, expiry and confirmation stamps.
-- handed_over_at in particular is the Phase 5 gate; losing it re-opens the
-- editor to sites that were never delivered.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM customer_access_tokens
   WHERE revoked_at IS NULL AND expires_at > now();
  IF n > 0 THEN
    RAISE EXCEPTION '511 ROLLBACK REFUSED: % live customer token(s) would be destroyed and cannot be reconstructed. Revoke them deliberately first if you really mean this.', n;
  END IF;
  SELECT count(*) INTO n FROM sites WHERE handed_over_at IS NOT NULL;
  IF n > 0 THEN
    RAISE EXCEPTION '511 ROLLBACK REFUSED: % site(s) are stamped handed_over_at; dropping the column loses the Phase 5 gate for them.', n;
  END IF;
END $$;

DROP TABLE IF EXISTS customer_access_tokens;
ALTER TABLE sites
  DROP COLUMN IF EXISTS handed_over_at,
  DROP COLUMN IF EXISTS live_link_expires_at,
  DROP COLUMN IF EXISTS transfer_confirmed_at;

COMMIT;
