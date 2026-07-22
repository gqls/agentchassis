-- 055_seed_allowlist_VERIFY.sql — post-seed verification (read-only).
-- PASS when: key_present = t, n = 4, and the array contains
-- leopardessconsulting.co.uk (the domain that was actually blocking).
SELECT domain,
       content_data ? 'allowed_reference_domains'                         AS key_present,
       jsonb_array_length(content_data->'allowed_reference_domains')      AS n,
       content_data->'allowed_reference_domains' @> '["leopardessconsulting.co.uk"]'::jsonb
                                                                          AS has_leopardess,
       content_data->'allowed_reference_domains'                         AS domains
FROM sites WHERE domain = 'fundamentallyai.com';
