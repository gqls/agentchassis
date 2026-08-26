-- RETIREMENT steps 1+2, MOVED TO A SITE THAT IS ACTUALLY BEING DISPATCHED.
--
-- WHY THE CANARY MOVED. The first canary went to finetuning.uk, chosen as structurally
-- simplest (0 site_nav_items rows, both nav flags false). Sound on the criteria I used —
-- and I never asked the one that mattered: IS THAT SITE BEING SERVICED? It is not.
-- [MEASURED 2026-08-26 15:23Z] claims in the last 3h: leopardess 72, aiao 111,
-- finetuning.uk ZERO (last claim 05:09:30). finetuning is a starvation victim of
-- bugs_open/413 (selector ranks sites by their oldest row's AGE, loader serves by
-- PRIORITY, so a pinned old row freezes its site's age and starves younger sites).
-- The finetuning archive STAYS as it is -- archived pages keep serving, so it is safe
-- parked, and its relink is load position 1 whenever 413 clears.
--
-- So this repeats the SAME test on a site whose dispatcher is demonstrably alive.
--
-- THE HYPOTHESIS UNDER TEST, unchanged and still falsifiable at the served bytes:
-- archiving drops the page out of loadResolverPageSet (`status NOT IN ('deleted','archived')`),
-- so applyCTARecompute's KEEP #2 stops returning early, KEEP #3 cannot catch a relative
-- /tools/... path, and control reaches the positional pick that the nav_order 1 -> 900
-- demotion already made correct. If WRONG, /careers.html completes and still serves
-- href="/tools/password-entropy.html" twice.
--
-- Archiving does NOT unpublish (verified on finetuning: archived page still returns 200),
-- so there is no 404 window and no dead inbound link. Fully reversible.
-- No LLM in either half.

BEGIN;

-- Step 1: archive
UPDATE pages
SET status = 'archived', updated_at = now()
WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'   -- leopardessconsulting.co.uk
  AND url = '/tools/password-entropy.html'
  AND status = 'active';

-- Step 2: one relink, on the page carrying TWO references so both fields are tested.
-- priority 30 puts it at load position 1 on this site (best existing is 60).
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
VALUES (
  '4851f6fc-71cf-4160-a270-e03d6d3e0732',
  'c5e3f65d-90e6-4c22-9944-49cc9b601786',
  'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
  'page_rerender', 'page-rerender', 'detected', 'high', 30,
  'Retirement step 2 canary (moved to a serviced site): re-resolve the two CTA hrefs on leopardess/careers that point at the now-archived password-entropy tool',
  'retire_relink:c5e3f65d-90e6-4c22-9944-49cc9b601786',
  jsonb_build_object(
    'reason',    'cta_links_stale',
    'check',     'misdirected_cta',
    'page_id',   'c5e3f65d-90e6-4c22-9944-49cc9b601786',
    'page_name', 'careers',
    'fix',       'The CTA destination /tools/password-entropy.html is now archived and must not be linked. Recompute the CTA targets so each href points at a live, relevant tool.',
    'original_pipeline', 'build'
  )
);

DO $$
DECLARE n_arch int; n_active int; n_lib int; n_item int;
BEGIN
  SELECT count(*) INTO n_arch FROM pages
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
     AND url='/tools/password-entropy.html' AND status='archived';
  IF n_arch <> 1 THEN
    RAISE EXCEPTION 'expected 1 archived leopardess row, found %', n_arch;
  END IF;

  -- Two down, one to go: only ai-agent-orchestration.com should still be active.
  SELECT count(*) INTO n_active FROM pages
   WHERE url='/tools/password-entropy.html' AND status='active';
  IF n_active <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 site still active (aiao), found %', n_active;
  END IF;

  -- Owner decision 1: the shared library component STAYS. Asserted by count of active
  -- matches, NOT by exact name -- there is no row called `tool-password-entropy`.
  SELECT count(*) INTO n_lib FROM content_components
   WHERE name ILIKE '%password-entropy%' AND is_active;
  IF n_lib <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 active library component, found %', n_lib;
  END IF;

  SELECT count(*) INTO n_item FROM site_work_items
   WHERE item_key='retire_relink:c5e3f65d-90e6-4c22-9944-49cc9b601786';
  IF n_item <> 1 THEN
    RAISE EXCEPTION 'canary relink not created (% rows)', n_item;
  END IF;

  RAISE NOTICE 'leopardess archived, canary relink queued at priority 30; 1 site still active';
END $$;

COMMIT;
