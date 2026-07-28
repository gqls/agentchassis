-- p4_22 — put maxlength on the request form so the over-length error cannot happen
--
-- The server-side half shipped in the 8th deploy: an over-length submission now
-- gets a styled page naming the field and the counts, instead of 52 bytes of
-- text/plain. This is the preventive half — the browser simply stops accepting
-- more, so a visitor never loses their place at all.
--
-- Limits mirror the Go exactly (service.go overLongField). If they ever diverge,
-- the browser silently truncates at a length the server would have accepted, or
-- the server rejects something the browser allowed — so they are stated here as
-- one list and must be changed together.
--
--     name       200     email    254
--     business   500     audience 2000     notes 4000
--
-- SAFE TO EDIT DIRECTLY: this component is forked to idea.uk — 1 site, 1
-- instance (checked) — so there is no backward-compatibility concern for other
-- sites, which is the trap a shared-template edit carries.
--
-- maxlength is advisory in the browser and trivially bypassed, which is why the
-- server-side check stays exactly as it is. This removes the common case; it is
-- not a validation boundary.
--
-- NOT touched: company_url (the honeypot — capping it would tell a bot something)
-- and _elapsed (hidden timing field).

\set comp '8a88fcd4-83fe-4f7a-bed1-f270e6edee53'

-- BEFORE: expect 0
SELECT count(*) AS maxlength_attrs_before
FROM (SELECT regexp_matches(html_template, 'maxlength="[0-9]+"', 'g') FROM content_components WHERE id = :'comp') t;

BEGIN;

UPDATE content_components SET html_template =
  regexp_replace(
  regexp_replace(
  regexp_replace(
  regexp_replace(
  regexp_replace(html_template,
    '(<input[^>]*name="name")',      '\1 maxlength="200"'),
    '(<input[^>]*name="email")',     '\1 maxlength="254"'),
    '(<input[^>]*name="business")',  '\1 maxlength="500"'),
    '(<input[^>]*name="audience")',  '\1 maxlength="2000"'),
    '(<textarea[^>]*name="notes")',  '\1 maxlength="4000"'),
  updated_at = now()
WHERE id = :'comp';

COMMIT;

-- AFTER: expect 5, one per real field
SELECT count(*) AS maxlength_attrs_after
FROM (SELECT regexp_matches(html_template, 'maxlength="[0-9]+"', 'g') FROM content_components WHERE id = :'comp') t;

-- and each one on the right field
SELECT (regexp_matches(html_template, '(name="[a-z_]+"[^>]*maxlength="[0-9]+")', 'g'))[1] AS field_and_limit
FROM content_components WHERE id = :'comp';
