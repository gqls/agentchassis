-- RETIREMENT STEP 1 of 3, CANARY SITE ONLY: archive finetuning.uk's password-entropy page.
--
-- Owner decision 1 (2026-08-25): the tool "can disappear everywhere" from these three sites,
-- but the shared library component `tool-password-entropy` STAYS is_active=true and available
-- to new sites. This touches ONE page row on ONE site. It does NOT touch the library component.
--
-- WHY ARCHIVE IS ITS OWN STEP, AND WHY IT COMES FIRST:
--   * `retract_page_deployment` (the file-removal half) REFUSES any page something editorial
--     still links to — by design, so a retraction cannot create a dead link. 13 rows across 11
--     active pages on this site still carry href="/tools/password-entropy.html".
--   * Those references cannot be re-resolved while the destination is still valid:
--     applyCTARecompute's KEEP #2 returns early for a stored destination in `validPages`.
--   * `loadResolverPageSet` (resolve_internal_links_action.go:964) selects
--     `status NOT IN ('deleted','archived')`. So ARCHIVING is what drops the page out of
--     validPages and unblocks the re-resolution. It is the key that turns both locks.
--
-- SAFETY: archiving does NOT unpublish. The action's own header: archiving "removes it from
-- every derivation AND from re-rendering — so the last HTML it ever rendered is frozen and keeps
-- being served". So the page still serves 200 after this; there is no 404 window. Fully
-- reversible by setting status back to 'active'.
--
-- No LLM involved.

BEGIN;

UPDATE pages
SET status = 'archived', updated_at = now()
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'   -- finetuning.uk
  AND url = '/tools/password-entropy.html'
  AND status = 'active';

DO $$
DECLARE
  n_archived int; n_active_left int; n_library int;
BEGIN
  SELECT count(*) INTO n_archived FROM pages
   WHERE site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
     AND url='/tools/password-entropy.html' AND status='archived';
  IF n_archived <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 archived row on finetuning.uk, found %', n_archived;
  END IF;

  -- The OTHER two sites must be untouched at this point: this is a canary.
  SELECT count(*) INTO n_active_left FROM pages
   WHERE url='/tools/password-entropy.html' AND status='active';
  IF n_active_left <> 2 THEN
    RAISE EXCEPTION 'expected 2 sites still active (canary scope), found %', n_active_left;
  END IF;

  -- Owner decision 1's explicit carve-out: the shared library component STAYS.
  -- ⚠ Asserted by COUNT OF ACTIVE MATCHES, not by an exact name. There is NO row named
  -- `tool-password-entropy` — the handoff says there is, and a guard written to its wording
  -- would have found 0 and falsely aborted this transaction. The live picture (2026-08-26):
  --   tool-password-entropy_pre_037                      is_active = t   <- the library row
  --   tool-password-entropy-{4 per-site variants}        is_active = f
  --   password-entropy_pre_037                           is_active = f
  -- The invariant that actually matters is that exactly one stays active and this run does
  -- not change that. (This UPDATE touches `pages` only, so it cannot reach these rows —
  -- the assertion is here to catch a concurrent session, not this statement.)
  SELECT count(*) INTO n_library FROM content_components
   WHERE name ILIKE '%password-entropy%' AND is_active;
  IF n_library <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 active password-entropy library component, found %', n_library;
  END IF;

  RAISE NOTICE 'canary archived: 1 finetuning.uk page, 2 sites still active, library component intact';
END $$;

COMMIT;
