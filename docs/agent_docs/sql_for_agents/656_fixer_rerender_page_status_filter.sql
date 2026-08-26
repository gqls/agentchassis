-- 656_fixer_rerender_page_status_filter.sql
--
-- bugs_open/404 fix candidate 3, which is independent of the rest of that bug and
-- independently correct: `component-template-fixer`'s `create_rerender` step files
-- one `page_rerender` item per page carrying a fixed component, and its query has
-- NO PAGE-STATUS FILTER AT ALL.
--
-- WHAT THAT COSTS. It files re-renders for ARCHIVED pages — re-rendering and
-- re-publishing pages the platform has retired, and making a retraction
-- self-undoing (delete the file, the next template fix republishes it). That is
-- `bugs_open/098`'s mechanism at a seam 098's own sweep did not reach.
-- `[MEASURED 2026-08-25]` of 60 live `tool-cta` instances, **16 sit on archived
-- pages**; that query would file re-renders for all 16.
--
-- NO `_HOLD`, and the reasoning is worth stating because the sibling migration in
-- this same bug DOES need one. The fixer executes SQL held in its own workflow
-- config, so this needs no new binary and an early apply cannot break anything —
-- unlike a checks-array name, which fails a whole step against an older image.
-- The Go half of bugs_open/404 (the livespec vocabulary) needs no migration
-- either. The two legs are genuinely independent.
--
-- ⚠⚠ ORDERING: APPLY THIS **BEFORE** THE BINARY ROLLS. NOT "at merge", which is
-- what the first draft said and which the council's editquality seat correctly
-- called insufficient.
--
-- The Go half adds a livespec declaration asserting this query CONTAINS
-- `p.status = 'active'`. The estate's normal shipping order is binary-then-config
-- (the recorded landmine "a config-backed fix ships in TWO halves"), and under
-- that order the 07:00 `live-declaration-drift-check` exits 1 the morning after
-- the roll, naming the missing fragment.
--
-- That finding would be TRUE — the live query really would not carry the filter —
-- so it is not a false alarm, and it is exactly the kind of true-but-expected
-- alarm that teaches people to ignore an auditor. The window is entirely
-- avoidable: this file needs NO binary, so applying it first closes it.
--
--   apply 656  ->  the declaration is already true  ->  roll the binary  ->  silence
--
-- If it does fire because the roll went first, the remedy is to apply this file,
-- not to touch the declaration.
--
-- WHY `p.status = 'active'` AND NOT A BUILD-STATUS TEST, AND WHY THAT SPELLING —
-- MEASURED, NOT ASSERTED. The council's debug_historian seat objected at HIGH
-- that this literal was CHOSEN without enumerating what `pages.status` actually
-- holds, which is the estate's standing requirement before any status-scoped
-- query ("a pages query that filters on status may be filtering on NOTHING — two
-- of the four spellings in live use exist"). The seat was right that it had not
-- been done. Done now, and it settles three things:
--
--   1. `[MEASURED 2026-08-26]` live `pages.status` holds exactly TWO values —
--      active 948, archived 68. No third value, no NULLs. So this filter excludes
--      something real and is not a no-op.
--
--   2. `[MEASURED 2026-08-26]` `datahelpers.PageWantedLivePredicateFor("p")`
--      renders **byte-identically** to the literal written below —
--      `p.status = 'active'` — so the livespec fragment, the parity test and this
--      migration are all asserting one string. Had they differed, either this
--      file or that test would have been wrong and only production would have
--      found out.
--
--   3. `[MEASURED 2026-08-26]` the containment this buys, across ALL components
--      rather than just tool-cta: **31 archived pages carrying 111 non-owned
--      component instances** stop being eligible for a template-fix re-render.
--      (The bug file's figure — 16 of 60 tool-cta instances — is the same defect
--      seen through one component.)
--
-- It is the LIFECYCLE axis: "does the platform still want this page served".
-- `build_status` answers a different question (has it ever shipped), and archiving
-- deliberately leaves the build columns untouched — which is why an archived page
-- keeps serving until it is retracted.
--
-- HOW TO CHECK IT WORKED, and it is not the item count:
--   SELECT default_config #>> '{workflow,steps,create_rerender,config,query}'
--     FROM agent_definitions WHERE type='component-template-fixer' AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- then, on the next template fix, no `page_rerender` item should name a page whose
-- `pages.status` is 'archived'.

BEGIN;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'component-template-fixer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL THEN
    RAISE EXCEPTION '656: no live component-template-fixer create_rerender query — refusing to guess';
  END IF;

  -- ⚠ IDEMPOTENCY BEFORE THE SNAPSHOT, NOT AFTER.
  -- A replay that snapshots first writes a row labelled 'pre-update' which by then
  -- holds the POST-update state — and that is the artefact someone restores FROM.
  -- LANDMINES records the trap; migrations 526 and 541 both have this the wrong
  -- way round, which is how it keeps being copied.
  IF q LIKE '%p.status = ''active''%' THEN
    RAISE NOTICE '656: already applied — no snapshot taken, no update made';
    RETURN;
  END IF;

  -- Anchor on the LIVE text, which has already drifted from migration 460's file:
  -- it GAINED the rebuild_policy clause and LOST the filename spec key
  -- (`[MEASURED 2026-08-26]`). That drift is exactly why the fix is anchored here
  -- and declared in livespec rather than asserted against 460. If it has moved
  -- again, refuse rather than write into a shape nobody has read.
  IF q NOT LIKE '%AND p.rebuild_policy IS DISTINCT FROM ''owned''%' THEN
    RAISE EXCEPTION '656: the live create_rerender query has moved — re-anchor before applying. Live text begins: %', left(q, 240);
  END IF;

  PERFORM snapshot_agent('component-template-fixer',
    '656 pre-image: page-status filter on create_rerender');

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,create_rerender,config,query}',
           to_jsonb(replace(q,
             'AND p.rebuild_policy IS DISTINCT FROM ''owned''',
             'AND p.rebuild_policy IS DISTINCT FROM ''owned'' AND p.status = ''active'''))),
         updated_at = NOW()
   WHERE type = 'component-template-fixer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

-- VERIFY AS A RAISE ON THE FINAL STATE. A verify block made of SELECTs cannot stop
-- the COMMIT — ON_ERROR_STOP does not fire on a non-empty result — so the
-- migration would report success having written the wrong thing.
--
-- Every check is POSITIVE-FORM: it asserts what must be present, so a key that has
-- gone missing fails loudly rather than NULL-ing quietly into a pass.
DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'component-template-fixer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  n := (length(q) - length(replace(q, 'p.status = ''active''', ''))) / length('p.status = ''active''');
  IF n <> 1 THEN
    RAISE EXCEPTION '656 verify: the page-status filter appears % time(s), expected exactly 1 — a replace() that matched twice would double the predicate', n;
  END IF;

  -- The replace() rewrites a whole query string, so the things it must NOT have
  -- lost are checked one by one. A rewrite that dropped any of these would leave
  -- the filter present and the query broken, and the count above would not see it.
  IF q NOT LIKE '%''reason'',''template_changed''%' THEN
    RAISE EXCEPTION '656 verify: the reason stamp was lost by the rewrite — items would file with no reason and route to ASSEMBLE (bugs_open/404)';
  END IF;
  IF q NOT LIKE '%rebuild_policy IS DISTINCT FROM ''owned''%' THEN
    RAISE EXCEPTION '656 verify: the owned-page guard was lost by the rewrite';
  END IF;
  IF q NOT LIKE '%NOT EXISTS%' THEN
    RAISE EXCEPTION '656 verify: the dedup NOT EXISTS was lost by the rewrite — every fix would re-file the same items';
  END IF;
  IF q NOT LIKE '%pc.component_id = $1%' THEN
    RAISE EXCEPTION '656 verify: the component scoping was lost by the rewrite — every page on the site would be filed';
  END IF;
END $$;

COMMIT;
