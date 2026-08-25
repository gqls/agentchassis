-- ROLLBACK for 609 — restores the ORIGINAL (false) comment verbatim.
--
-- ⚠ Rolling this back re-installs a comment that misdescribes the column in every clause.
-- The only reason to run this is to restore a byte-exact prior state; if you are here
-- because something broke, note that a COMMENT cannot break anything — look elsewhere.
\set ON_ERROR_STOP on
BEGIN;
COMMENT ON COLUMN content_components.usage_count IS
  'Times this component has been assigned to a page. Incremented by selector. Higher = more battle-tested.';
COMMIT;
