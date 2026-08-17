-- 453 — held-pair-canary-escalation: a time limit on "waiting for a human to
--       canary this pair", and a named owner per type
--
-- OWNER DECISION 2026-08-17, put in bugfix_277's README_where_we_are.md and
-- answered: "Name an owner per type and set a time limit."
--
-- ============================================================================
-- WHAT IS BEING FIXED
-- ============================================================================
-- `detected-item-promoter` (SCH-026, seed 430 + 444) deliberately refuses to
-- promote an (item_type, handler_agent) pair that has never completed one --
-- every new type's first dispatch must be a human-run canary, so a brand-new
-- handler cannot misfire fleet-wide unwatched. That rule is right and stays.
--
-- The gap is that NOTHING DRIVES THE HUMAN. A held pair sits at `detected`,
-- which is an unclaimable state nobody reads, and no clock runs on it. So a new
-- finding type can stall indefinitely -- which is `bugs_open/083`'s own disease
-- (findings stranded in a queue with no consumer) reproduced one step later,
-- inside the mechanism built to cure it.
--
-- [MEASURED 2026-08-17] held today:
--   page_component_status_drift -> component-template-fixer   4 rows, 7 DAYS, 0 lifetime completes
--   placeholder_contact         -> page-build-handler         1 row,  1 day,  0 lifetime completes
--
-- ============================================================================
-- THE MECHANISM, AND WHY IT PARKS RATHER THAN RAISING SOMETHING NEW
-- ============================================================================
-- Pure-SQL scheduled task on the SCH-006 pattern (`fire_message=false`, the
-- pre_query IS the worker), daily. It does NOT create new work items. It moves
-- the held rows themselves from `detected` to `needs_human_review`, carrying
-- their facts -- which is exactly what CQ-023's router does for findings a
-- machine must not touch, and is the estate's existing "a human must decide"
-- channel rather than a new one.
--
-- Three reasons parking beats filing a new item:
--   * no dedup problem (`idx_swi_dedup` is per-SITE and a held PAIR is
--     fleet-wide, so a new row would need a synthetic key and would re-raise);
--   * the row stops being `detected`, so the promoter no longer reconsiders it
--     -- the decision is explicitly a human's now;
--   * the finding keeps its own identity, site and page instead of being
--     described second-hand by a tracking item.
-- Reversible: set the rows back to `detected` and clear the result key.
--
-- ============================================================================
-- N = 3 DAYS, AND WHY
-- ============================================================================
-- The promoter ticks every 15 minutes, so a pair is either canaried promptly or
-- it is forgotten; there is no middle case where 3 days is too soon. 3 days
-- clears a weekend-length gap without being noisy, and the type appears rarely
-- (2 pairs held in the mechanism's lifetime). Both pairs above are already past
-- it -- 7 days and (at the next daily tick) 2 -- which is the point: they ARE
-- forgotten. Change the literal in ONE place below if it proves wrong.
--
-- ============================================================================
-- THE OWNER MAP IS NOT A GATE -- THIS IS THE LOAD-BEARING DESIGN CHOICE
-- ============================================================================
-- Migration 444's council round produced exactly this warning from the guardian
-- seat about ITS allow-list: a hard-coded enumeration means anything not named
-- silently stops being handled, with no signal beyond a pile that looks like
-- normal backlog. So this file does NOT let the map decide who escalates.
-- EVERY pair past the limit escalates. The map only enriches the row with an
-- owner; an unmapped pair escalates with owner `(UNASSIGNED - claim this)`,
-- i.e. an unowned type gets LOUDER, not quieter. Adding an owner later is a
-- one-line change and costs nothing if forgotten.
--
-- The two owners named, from evidence rather than invention:
--   placeholder_contact -> `bugs_open/201` lane
--       (docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch). That
--       lane last touched check_placeholder_contact.go (2026-08-05, commit
--       fixing three checks that dispatched page-content-writer directly) and
--       owns the dispatch path of its handler, page-build-handler.
--   page_component_status_drift -> DELIBERATELY UNASSIGNED.
--       check_page_component_status_drift.go was added 2026-07-10 and has NEVER
--       been touched since; no lane doc claims it. Naming a plausible owner here
--       would be a fabrication that reads as a decision. The absence is the
--       finding, and the escalated row says so in words.

BEGIN;

INSERT INTO scheduled_tasks (name, description, interval_seconds, enabled, fire_message, target_agent_type, pre_query)
VALUES (
  'held-pair-canary-escalation',
  'Owner decision 2026-08-17 (bugs_open/083). A (item_type, handler_agent) pair held by SCH-026''s known-good rule for more than 3 days is escalated from detected to needs_human_review, carrying its named owner and what the human must do. The owner map enriches; it never gates.',
  86400,
  true,
  false,
  'generic',
  $Q$
    WITH held AS (
        SELECT wi.id, wi.item_type, wi.handler_agent, wi.created_at
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          -- held BY THE CANARY RULE specifically: the pair has never completed.
          -- A pair held for any other reason (inactive handler, a 444 door) is
          -- not this task's business and is left alone.
          AND NOT EXISTS (
            SELECT 1 FROM site_work_items done
            WHERE done.item_type = wi.item_type
              AND done.handler_agent = wi.handler_agent
              AND done.status = 'complete'
          )
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
    ),
    overdue AS (
        -- the CLOCK runs on the pair, not the row: a pair forgotten for 3 days
        -- escalates all of its rows together, so the human sees the whole case.
        SELECT h.*, p.oldest, (now()::date - p.oldest::date) AS days_waiting
        FROM held h
        JOIN (
            SELECT item_type, handler_agent, min(created_at) AS oldest
            FROM held GROUP BY 1, 2
            HAVING min(created_at) < now() - interval '3 days'
        ) p ON p.item_type = h.item_type AND p.handler_agent = h.handler_agent
    ),
    owners (item_type, owner) AS (
        VALUES
          ('placeholder_contact',
           'bugs_open/201 lane — docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch'),
          ('page_component_status_drift',
           '(UNASSIGNED - claim this) check_page_component_status_drift.go added 2026-07-10, never touched since, no lane doc claims it')
    ),
    escalated AS (
        UPDATE site_work_items wi
        SET status = 'needs_human_review',
            resolution_path = 'auto:held_pair_escalated',
            result = COALESCE(wi.result, '{}'::jsonb) || jsonb_build_object(
                'held_pair_escalation', jsonb_build_object(
                    'at',            to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
                    'reason',        'held by the detected-item-promoter known-good rule (SCH-026): this (item_type, handler_agent) pair has never completed one, so the promoter will not dispatch it until a human runs one by hand',
                    'days_waiting',  o.days_waiting,
                    'limit_days',    3,
                    'pair',          o.item_type || ' -> ' || o.handler_agent,
                    'owner',         COALESCE(ow.owner, '(UNASSIGNED - claim this) no owner named for this item_type in migration 453'),
                    'what_to_do',    'Promote ONE row of this pair by hand and watch it: UPDATE site_work_items SET status=''triaged'', pipeline=''build'', triaged_at=now(), spec=jsonb_set(COALESCE(spec,''{}''::jsonb),''{original_pipeline}'',to_jsonb(pipeline)) WHERE id=''<one id>''. If it completes, the pair becomes known-good and the promoter takes the rest automatically. If it fails, that is the finding — file it.',
                    'escalated_by',  'held-pair-canary-escalation (migration 453, owner decision 2026-08-17)'
                )
            ),
            updated_at = now()
        FROM overdue o
        LEFT JOIN owners ow ON ow.item_type = o.item_type
        WHERE wi.id = o.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent
    )
    SELECT COUNT(*)::text AS escalated,
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs
    FROM escalated
    WHERE (SELECT COUNT(*) FROM escalated) > 0
  $Q$
)
ON CONFLICT (name) DO UPDATE
SET description      = EXCLUDED.description,
    interval_seconds = EXCLUDED.interval_seconds,
    enabled          = EXCLUDED.enabled,
    fire_message     = EXCLUDED.fire_message,
    pre_query        = EXCLUDED.pre_query,
    updated_at       = now();

-- ============================================================================
-- Verification. RAISE, not SELECT -- a plain SELECT cannot stop the COMMIT.
-- ============================================================================
DO $$
DECLARE
    n_rows       int;
    q            text;
    n_held       int;
    n_overdue    int;
    n_unmapped   int;
    n_notheld    int;
BEGIN
    SELECT count(*) INTO n_rows FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF n_rows <> 1 THEN
        RAISE EXCEPTION '453: expected exactly 1 held-pair-canary-escalation row, found %', n_rows;
    END IF;

    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF q NOT LIKE '%interval ''3 days''%' THEN
        RAISE EXCEPTION '453: the 3-day limit is not in the live pre_query';
    END IF;
    IF q NOT LIKE '%UNASSIGNED - claim this%' THEN
        RAISE EXCEPTION '453: the unmapped-owner fallback is missing — an unowned pair would escalate with a NULL owner instead of a loud one';
    END IF;

    -- the population, measured now, so the NOTICE states what the first tick will do
    SELECT count(*) INTO n_held
      FROM site_work_items wi
      WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'')<>''
        AND NOT EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type=wi.item_type
                          AND d.handler_agent=wi.handler_agent AND d.status='complete')
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type=wi.handler_agent
                      AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL);

    SELECT count(*) INTO n_overdue
      FROM site_work_items wi
      WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'')<>''
        AND NOT EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type=wi.item_type
                          AND d.handler_agent=wi.handler_agent AND d.status='complete')
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type=wi.handler_agent
                      AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
        AND (wi.item_type, wi.handler_agent) IN (
            SELECT h.item_type, h.handler_agent FROM site_work_items h
            WHERE h.status='detected' AND COALESCE(h.handler_agent,'')<>''
              AND NOT EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type=h.item_type
                                AND d.handler_agent=h.handler_agent AND d.status='complete')
            GROUP BY 1,2 HAVING min(h.created_at) < now() - interval '3 days');

    -- POSITIVE CONTROL: the task must have something to do on its first tick,
    -- else "it works" is untested and the NOTICE below is vacuous.
    IF n_overdue = 0 THEN
        RAISE EXCEPTION '453: POSITIVE CONTROL FAILED — no pair is past the 3-day limit, so the first tick would be a no-op and this task would ship unexercised. Expected the two pairs measured 2026-08-17. Re-measure before applying.';
    END IF;

    -- NEGATIVE CONTROL, computed as an INTERSECTION of two independently-built
    -- sets so it can actually come out non-zero. (An earlier draft of this block
    -- asserted `EXISTS(X) AND NOT EXISTS(X) = 0`, which is a tautology and could
    -- never fail — the very trap this file's header is about, and the same one
    -- 430's author caught in their own verify block before applying.)
    --
    -- Set A: rows the escalation would move (the overdue held set).
    -- Set B: rows whose pair HAS completes, i.e. promotable — the promoter's
    --        business, never this task's.
    -- If the `NOT EXISTS` in the held CTE were inverted, A would BECOME B and
    -- this count would jump from 0 to the whole promotable pile.
    SELECT count(*) INTO n_notheld
      FROM site_work_items wi
      WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'')<>''
        -- in A: pair has no completes AND is past the limit
        AND NOT EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type=wi.item_type
                          AND d.handler_agent=wi.handler_agent AND d.status='complete')
        AND (wi.item_type, wi.handler_agent) IN (
            SELECT h.item_type, h.handler_agent FROM site_work_items h
            WHERE h.status='detected' AND COALESCE(h.handler_agent,'')<>''
            GROUP BY 1,2 HAVING min(h.created_at) < now() - interval '3 days')
        -- and simultaneously in B: the pair has at least one complete
        AND EXISTS (SELECT 1 FROM site_work_items b WHERE b.item_type=wi.item_type
                      AND b.handler_agent=wi.handler_agent AND b.status='complete');
    IF n_notheld <> 0 THEN
        RAISE EXCEPTION '453: NEGATIVE CONTROL FAILED — % row(s) are in BOTH the escalation set and the promotable pile, so the held predicate is inverted or the pair test is wrong', n_notheld;
    END IF;

    SELECT count(DISTINCT wi.item_type) INTO n_unmapped
      FROM site_work_items wi
      WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'')<>''
        AND NOT EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type=wi.item_type
                          AND d.handler_agent=wi.handler_agent AND d.status='complete')
        AND wi.item_type NOT IN ('placeholder_contact','page_component_status_drift');

    RAISE NOTICE '453: held-pair-canary-escalation live (daily, fire_message=false, limit 3 days). Held by the canary rule now: % rows, of which % are past the limit and will escalate on the first tick. Item types with no owner mapping: % (they escalate LOUDER, with an UNASSIGNED owner — the map never gates).',
        n_held, n_overdue, n_unmapped;
END $$;

COMMIT;
