-- 055_seed_allowlist.sql — seed fundamentallyai.com's approved cross-reference
-- allowlist (bugs_open/055). Guarded, idempotent, backup + RETURNING post-state.
--
-- WHAT: writes sites.content_data->'allowed_reference_domains' for
--   fundamentallyai.com = the four owner-approved first-party domains it markets
--   as case studies (all OUR OWN sites — cannot smuggle a third party past the
--   contamination guard, which only checks the five hardcoded known domains).
--
-- WHEN: run ONLY after the validate_page_content.go allowlist code is LIVE on the
--   running chassis pod (verify by pod-grep for loadAllowedReferenceDomains —
--   never the image tag). Seeding before the code is live is harmless (the key
--   sits inert) but pointless.
--
-- WHY here, not the schema-migration ledger: this is environment-specific SITE
--   DATA for one row, not a schema change — it does not belong in the numbered
--   migration sequence. It gets the SAME backup / pre-state / idempotency /
--   verify discipline a migration would (council debug_historian, corr
--   03908b72, round 1).
--
-- Run:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < 055_seed_allowlist.sql

\set ON_ERROR_STOP on
BEGIN;

-- 0. BACKUP the exact pre-state of the row's content_data (session-local; also
--    print it so it lands in the run log for out-of-band restore).
CREATE TEMP TABLE _bak_055 AS
  SELECT id, domain, content_data
  FROM sites WHERE domain = 'fundamentallyai.com';
SELECT 'PRE-STATE BACKUP' AS marker, domain,
       content_data ? 'allowed_reference_domains' AS key_already_present
FROM _bak_055;

-- 1. GUARDED, IDEMPOTENT UPDATE — writes only when the key is absent, so a
--    re-run never clobbers a hand-tuned list (to change the list later: run the
--    ROLLBACK sidecar first, then re-seed). RETURNING proves the post-state; an
--    empty result means the guard held (key already present) — expected on re-run.
UPDATE sites
SET content_data = jsonb_set(
        COALESCE(content_data, '{}'::jsonb),
        '{allowed_reference_domains}',
        '["leopardessconsulting.co.uk","finetuning.uk","idea.uk","relojistas.com"]'::jsonb,
        true /* create_if_missing */)
WHERE domain = 'fundamentallyai.com'
  -- type guard: jsonb_set with an object path only makes sense on an object
  -- (some sites store content_data as a JSON array — e.g. leopardess). null is
  -- fine, COALESCE turns it into '{}'. fundamentallyai's is an object.
  AND (content_data IS NULL OR jsonb_typeof(content_data) = 'object')
  AND NOT (content_data ? 'allowed_reference_domains')
RETURNING 'UPDATED' AS marker, domain,
          content_data->'allowed_reference_domains' AS allowed_reference_domains;

-- 2. POST-CONDITION (in-txn): the key is present with the four approved domains.
SELECT 'POST-STATE' AS marker, domain,
       content_data ? 'allowed_reference_domains' AS key_present,
       jsonb_array_length(content_data->'allowed_reference_domains') AS n,
       content_data->'allowed_reference_domains' AS allowed_reference_domains
FROM sites WHERE domain = 'fundamentallyai.com';

COMMIT;
