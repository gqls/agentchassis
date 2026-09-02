-- 683_content_listing_rerender_after_roll_HOLD.sql
--
-- THE PROPAGATION HALF OF bugs_open/425. Held back from the runner deliberately:
-- it is correct ONLY after the chassis image carrying commit f57f5ad1f has rolled.
--
-- ══ WHY THIS FILE EXISTS ══════════════════════════════════════════════════════
-- Migration 682 guarded content-listing's per-item slots. A template edited by
-- SQL ships NOTHING: page_components.rendered_html holds what was rendered when
-- it was rendered, so 682 changes only what the NEXT render produces. The
-- council returned REVISE on exactly this (correlation 84b51f16, high-severity
-- objections from render_guardian AND debug_historian), and they were right —
-- "the next render picks it up" is prose, not a propagation step, and there is a
-- LANDMINE for it: "a template edited by SQL ships NOTHING — the
-- template_changed fan-out lives in component-template-fixer".
--
-- ══ THE PRECONDITION, AND IT IS NOT OPTIONAL ══════════════════════════════════
-- ⚠ DO NOT APPLY UNTIL THE CHASSIS CARRYING f57f5ad1f IS LIVE. SQL cannot check
-- this, so a human must, and it is one command against the running pod:
--
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--     | grep -m1 'build provenance'
--   git merge-base --is-ancestor f57f5ad1f <the sha that stamp names>   # exit 0 = shipped
--
-- (An EMPTY grep result means "not in range" — the line is a STARTUP line and
-- scrolls — not "unstamped". Fall back to the binary probe, with controls.)
--
-- Applied EARLY, this re-resolves against the OLD resolver: the guarded slots
-- collapse (682 is live) but no deck arrives, and 14 renders are spent for half
-- the fix. That is not harmful, it is wasteful — and it produces a page that
-- looks fixed and is not, which is worse.
--
-- ══ WHY THE REASON STRING IS THE WHOLE FILE ═══════════════════════════════════
-- page-rerender branches on spec.reason ALONE (check_rerender_mode). Only
--   image_landed | section_data_resolved | cta_links_stale | template_changed | literal_markdown
-- route to rerender_page_sections, which RE-RUNS the query.* resolvers. Anything
-- else — including no reason at all — routes to rerender_single_page, "simple
-- concatenation", which re-ships the stored `articles` array byte for byte. So a
-- rerender can COMPLETE, stamp a fresh deployed_at, and change nothing.
-- 'section_data_resolved' is the reason that re-runs the resolver, which is what
-- the Go half needs; it also re-renders from the template, which is what 682
-- needs. One reason serves both halves.
--
-- ⚠ AND WHEN YOU VERIFY: a COMPLETED page_rerender row is NOT evidence. Read
-- spec->>'reason' on it. That is bugs_open/384's own filing error.
--
-- ══ PRE-FLIGHT, BOTH ALREADY RUN ══════════════════════════════════════════════
-- 1. The section-shrink guard will NOT refuse. LANDMINES records this component
--    family as its classic trigger, and the refusal's own error text invites you
--    to lower a FLEET-WIDE guard to land one page — never do that; the unblock is
--    always the content gap. [MEASURED 2026-09-02] pages with an empty
--    meta_description, per site carrying content-listing: boxingonline 0 of 7
--    (mean deck 123 chars), dartsonline 0 of 23, garden-tools 0 of 5,
--    homegarden 0 of 4, idea.uk 0 of 9, robot-hands 0 of 8. Every card can be
--    filled, and the change ADDS ~123 chars per card. The direction is the
--    opposite of a shrink.
-- 2. STY-048: the section branch escalates the WHOLE page to the writer if any
--    section lacks a required source:"llm" field. Check per page before applying:
--      SELECT p.name, cc.name, f.key
--        FROM page_components pc
--        JOIN pages p ON p.id = pc.page_id
--        JOIN content_components cc ON cc.id = pc.component_id,
--             jsonb_each(cc.input_schema->'fields') f
--       WHERE p.id IN (SELECT page_id FROM page_components
--                       WHERE component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551')
--         AND f.value->>'source' = 'llm' AND (f.value->>'required')::boolean
--         AND NOT (pc.content_data ? f.key);
--
-- ══ SCOPE ═════════════════════════════════════════════════════════════════════
-- Files one page_rerender per PAGE carrying a content-listing instance — 14
-- instances across 6 sites as of 2026-09-02 (homegarden.uk 6, boxingonline.com 2,
-- dartsonline.com 2, garden-tools.uk 2, idea.uk 1, robot-hands.com 1), which is
-- fewer than 14 pages only if a page carries two. Every one of them renders four
-- empty slots today, so every one benefits.
--
-- ⚠ APPLYING THIS IS A DISPATCH DECISION, NOT A SCHEMA CHANGE. Other lanes own
-- these sites. To narrow it to one, add to the WHERE:
--     AND s.domain = 'boxingonline.com'
-- Nothing here writes to a page; it files triaged work items the normal
-- page-rerender handler picks up, so the queue's own ordering and guards apply.
--
-- Reversible: 683_..._ROLLBACK.sql deletes the items by batch_id while they are
-- still unstarted.

BEGIN;

-- GUARD 1: 682 must be applied, or there is nothing to propagate.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM content_components
         WHERE id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
           AND html_template LIKE '%{{if or .date .read_time}}<div class="article-card__meta">%'
    ) THEN
        RAISE EXCEPTION
            'migration 682 is not applied to content-listing — there is no template change to '
            'propagate. Apply 682 first, then this file, and only after the roll.';
    END IF;
END $$;

INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    page_id, priority, handler_agent, status, created_by,
    spec, item_key, batch_id
)
SELECT DISTINCT
    p.site_id,
    'bugs_open/425',
    'build',
    'page_rerender',
    'medium',
    'content-listing card slots guarded (mig 682) and the list-item projection now '
      || 'carries a display title + excerpt (f57f5ad1f) — re-resolve so the stored '
      || 'articles array picks both up',
    p.id,
    80,
    'page-rerender',
    'triaged',
    'bugs_open/425',
    jsonb_build_object(
        'reason',    'section_data_resolved',
        'page_name', p.name
    ),
    -- The canonical key shape, matching actions.PageRerenderItemKey exactly:
    -- page_rerender_<page name>_<site id>_<reason>. A second spelling would mean
    -- two items for one page, because idx_swi_dedup is (site_id, item_key).
    'page_rerender_' || p.name || '_' || p.site_id || '_section_data_resolved',
    '00000000-0000-0000-0000-000000000683'::uuid
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE pc.component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
  AND p.status IN ('active', 'deployed')
ON CONFLICT DO NOTHING;

-- VERIFY. DO/RAISE, not SELECTs: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    filed        int;
    wrong_reason int;
    targets      int;
BEGIN
    SELECT count(*) INTO filed
      FROM site_work_items WHERE batch_id = '00000000-0000-0000-0000-000000000683';

    SELECT count(DISTINCT p.id) INTO targets
      FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE pc.component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
       AND p.status IN ('active','deployed');

    -- The load-bearing assertion: an item whose reason is not one the
    -- check_rerender_mode conditional recognises takes assemble mode and cannot
    -- pick this change up. Filing one would be worse than filing none, because
    -- it completes and looks like success.
    SELECT count(*) INTO wrong_reason
      FROM site_work_items
     WHERE batch_id = '00000000-0000-0000-0000-000000000683'
       AND COALESCE(spec->>'reason','') <> 'section_data_resolved';

    IF wrong_reason > 0 THEN
        RAISE EXCEPTION 'ABORT: % item(s) filed with a reason that routes to assemble mode — '
                        'they would complete and change nothing', wrong_reason;
    END IF;

    IF filed = 0 AND targets > 0 THEN
        RAISE EXCEPTION 'ABORT: % target page(s) but 0 items filed — every insert hit '
                        'ON CONFLICT, which means open items already exist for these pages. '
                        'Read them before re-running.', targets;
    END IF;

    IF EXISTS (SELECT 1 FROM site_work_items
                WHERE batch_id = '00000000-0000-0000-0000-000000000683'
                  AND (spec->>'page_name' IS NULL OR spec->>'page_name' = '')) THEN
        RAISE EXCEPTION 'ABORT: an item is missing spec.page_name, which the section branch requires';
    END IF;

    RAISE NOTICE '683: filed % page_rerender item(s) for % target page(s), all '
                 'reason=section_data_resolved with a page_name', filed, targets;
END $$;

COMMIT;
