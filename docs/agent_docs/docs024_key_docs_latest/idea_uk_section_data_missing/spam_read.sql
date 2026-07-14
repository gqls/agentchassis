-- Spam read (read-only, schema-first): find the table holding contact-form / report-request
-- submissions for idea.uk, and whether it captures an IP address (needed for a block list).
-- The sample spam: order id like 'ord_1783948426211007948', requester test@test.com, and a
-- report-request shape (Requester/Business/Audience/Notes) — so search by those columns/values.

-- 1. Candidate tables: anything whose name suggests submissions/orders/leads/contacts:
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND (table_name ILIKE '%order%' OR table_name ILIKE '%submission%'
       OR table_name ILIKE '%lead%'   OR table_name ILIKE '%contact%'
       OR table_name ILIKE '%request%' OR table_name ILIKE '%form%'
       OR table_name ILIKE '%report%')
ORDER BY table_name;

-- 2. Which table actually contains that order id string? (Adjust the table name in the
--    UNION list below to the candidates from query 1 if they differ. This checks the two
--    most likely names; extend as needed.)
--    Run per-candidate rather than guessing — replace <t> with each candidate:
--      SELECT '<t>' AS tbl, count(*) FROM <t> WHERE (to_jsonb(<t>.*)::text) LIKE '%ord_1783948426211007948%';

-- 3. Once the table is known, inspect its schema for an IP/user-agent column and the
--    spam-identifying fields (email, created_at). Replace <t>:
--      \d <t>
