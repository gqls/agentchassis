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
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--          -o jsonpath='{.items[0].metadata.name}')
--   for sym in ListItemExcerpt resolvePagesWhereType ListItemTitleXYZNOTREAL; do
--     kubectl -n ai-persona-system exec "$POD" -- grep -aqs -- "$sym" /proc/1/exe \
--       && echo "$sym PRESENT" || echo "$sym absent"
--   done
--   # ListItemExcerpt must be PRESENT; resolvePagesWhereType is the positive
--   # control (a probe that finds nothing must not read as an answer); the
--   # invented symbol must be absent, and it CONTAINS ListItemTitle, so its
--   # absence also proves the grep is not matching loosely.
--
-- ⚠ NOT `git merge-base`. The first cut of this header said to resolve the
-- build-provenance stamp and test ancestry, and the council's debug_historian
-- seat objected at HIGH severity — correctly, and against this estate's own
-- lore: deploy state is verified at the RUNNING POD's binary, never at git. A
-- same-tag rebuild serves the node's cached image while merge-base still says
-- yes. Probe the CAPABILITY, not the commit.
-- ⚠ AND PROBE EVERY POD, not one. Two pods of one ReplicaSet are not guaranteed
-- identical for exactly the cached-image reason above; probing one of two left a
-- hole in this bug's own evidence that could not be closed afterwards, because
-- the unprobed pod was gone by the time it mattered.
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
-- 'template_changed' is the reason THE OWNING MECHANISM USES, and that is why
-- this file uses it too — see the reuse note below. Both it and
-- 'section_data_resolved' take the same branch, so either would re-resolve; the
-- tie-break is dedup, not behaviour.
--
-- ══ AND THE SCOPED PATH REALLY DOES RE-RESOLVE — CITED, NOT ASSUMED ═══════════
-- The council's render_guardian seat objected (round 2) that the plan leaned on a
-- landmine saying ASSEMBLE mode re-affirms a stale query.* snapshot, and then
-- ASSUMED the opposite of scoped mode without citing code. Fair, and now checked:
--
--   rerender_page_sections_action.go, file header, MECHANISM:
--     "newSourceResolver + planSection (plan_sections_action.go) rebuild each
--      section's resolved_data (queryresolve for query.*, the page-aware
--      ensureAssets for the hero)"
--     "...resolved_data merged last so it wins, matching RenderComponentAction"
--   and the call itself, same file:
--     resolver := newSourceResolver(siteID, params.DB, logger, pageName)
--
-- `articles` is source: query.blog_posts, i.e. a RESOLVED field, so it is rebuilt
-- freshly and wins over the stored copy. That is the whole premise of this file
-- and it is a citation now rather than an inference.
--
-- ⚠ AND WHEN YOU VERIFY: a COMPLETED page_rerender row is NOT evidence. Read
-- spec->>'reason' on it. That is bugs_open/384's own filing error.
--
-- ══ "BUT THE DISCRIMINATOR IS component_id, NOT reason" — NOT ON THIS PATH ═════
-- A correction landed in the webdesign lane's NOTES on 2026-09-02 (found by
-- experience_loop, relayed via the boxingonline session) saying the rerender
-- discriminator is component_id and that section_data_resolved "degrades to
-- assemble-only without one". That is TRUE and it is about the PRODUCER, not the
-- consumer — read this before concluding the file below is inert.
--
-- platform/livespec/rerender_reasons.go marks section_data_resolved
-- ComponentScoped with StampAlways=false, so create_rerender_items does not
-- STAMP the reason onto an item it files without a component_id: the item
-- arrives carrying NO reason, and then degrades. The degrade is a property of
-- the reason never reaching spec, not of the reason being insufficient.
--
-- THIS FILE WRITES spec.reason ITSELF, so the stamping step never runs, and the
-- consumer side reads nothing else. Verified at the LIVE agent config, both
-- steps, 2026-09-02:
--
--   page-rerender.check_rerender_mode.config.condition  -- verbatim:
--     input_data.spec.reason == 'image_landed' OR ... == 'section_data_resolved'
--     OR ... == 'cta_links_stale' OR ... == 'template_changed'
--     OR ... == 'literal_markdown'
--   -> spec.reason ALONE. component_id does not appear.
--
--   page-rerender.rerender_sections.config  -- verbatim keys:
--     reason, page_name, target_site_id, strip_literal_markdown,
--     record_dead_url_controls
--   -> "Re-render ALL sections from stored content_data + fresh resolved
--      fields". No component_id, so nothing needs scoping BY one.
--
-- Re-run both if you doubt it (config is live and mutable):
--   SELECT jsonb_pretty(default_config->'workflow'->'steps'->'check_rerender_mode')
--     FROM agent_definitions WHERE type='page-rerender' AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
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
-- ══ SCOPE, AND WHY THIS IS A COPY OF AN EXISTING MECHANISM ═══════════════════
-- ⚠ THIS IS NOT A NEW FAN-OUT. Three council seats (reuse_agent,
-- prior_art_librarian, guardian) objected — correctly — that the first cut of
-- this file hand-rolled a competing propagation while the platform already owns
-- one, and that the author quoted the landmine naming it and then bypassed it.
--
-- The owner is `component-template-fixer`, step `create_rerender`. Its SQL is
-- reproduced below almost verbatim, and copying it fixed three real defects the
-- hand-rolled version had:
--   * it excludes OWNED pages (`rebuild_policy IS DISTINCT FROM 'owned'`) —
--     bugs_open/301; the first cut would have filed rerenders for pages the save
--     path refuses, burning the dispatch;
--   * it dedups with NOT EXISTS on an OPEN item, not `ON CONFLICT DO NOTHING` —
--     which is the documented convention for site_work_items, and which the
--     `guidelines` seat flagged; ON CONFLICT can also silently leave a STALE row
--     under the same key untouched;
--   * it uses `p.status = 'active'`. [MEASURED 2026-09-02] `pages.status` holds
--     exactly two live values fleet-wide — `active` (1,082) and `archived` (90).
--     **`'deployed'` NEVER OCCURS**, and the first cut's `IN ('active','deployed')`
--     carried a dead literal. debug_historian called this before I measured it.
--
-- WHY A FILE AT ALL, THEN: that step runs inside the fixer's workflow, triggered
-- by `fix_result.component_id` — a template fix made BY THE AGENT. A template
-- edited by SQL never reaches it, which is precisely what the landmine says. So
-- this file is a HAND-RUN INSTANCE OF THAT STEP, not a rival to it. If the fixer
-- ever gains a trigger for SQL-applied template edits, delete this file rather
-- than maintaining both.
--
-- Using the SAME reason string is part of the reuse and is load-bearing: the
-- fixer's own NOT EXISTS dedup filters on `spec->>'reason' = 'template_changed'`,
-- so a row filed here under a different reason would be INVISIBLE to it and the
-- next real template fix would file a second, duplicate rerender for the page.
--
-- Targets: pages carrying a content-listing instance — 14 across 6 sites as of
-- 2026-09-02, all status='active' and rebuild_policy='generic'. The count is
-- derived at apply time, not baked in; it was 13 a few hours before it was 14.
--
-- ⚠ APPLYING THIS IS A DISPATCH DECISION, NOT A SCHEMA CHANGE. Other lanes own
-- these sites. To narrow to one, add to the WHERE:  AND s.domain = '<domain>'
--
-- Reversible: 683_..._ROLLBACK.sql deletes this batch while it is still unstarted.

BEGIN;

-- GUARD: 682 must be applied, or there is nothing to propagate.
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
    priority, handler_agent, status, created_by, spec, batch_id
)
SELECT DISTINCT
    p.site_id, 'side_effect', 'build', 'page_rerender', 'low',
    'Rerender page after template fix (bugs_open/425, mig 682 + f57f5ad1f): ' || p.name,
    80, 'page-rerender', 'triaged', 'bugs_open/425',
    jsonb_build_object(
        'reason',    'template_changed',
        'page_id',   p.id::text,
        'page_name', p.name,
        'domain',    s.domain
    ),
    '00000000-0000-0000-0000-000000000683'::uuid
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE pc.component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
  AND p.rebuild_policy IS DISTINCT FROM 'owned'
  AND p.status = 'active'
  AND NOT EXISTS (
        SELECT 1 FROM site_work_items w
         WHERE w.site_id = p.site_id
           AND w.item_type = 'page_rerender'
           AND w.spec->>'page_id' = p.id::text
           AND w.spec->>'reason'  = 'template_changed'
           AND w.status IN ('detected','triaged','claimed')
      );

-- VERIFY. DO/RAISE, not SELECTs: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT.
--
-- ⚠ THE COUNTS ARE DELIBERATELY COMPUTED BY DIFFERENT PREDICATES. The first cut
-- derived `targets` with the INSERT's own WHERE clause, so a wrong predicate made
-- targets and filed agree AT ZERO and the guard could never fire — a check that
-- cannot detect its own predicate being wrong (debug_historian, round 2). Here
-- `carrying` has NO status/policy/dedup filter, so a predicate that selects
-- nothing shows up as carrying > 0 with eligible = 0.
DO $$
DECLARE
    carrying int;   -- pages with the component, unfiltered
    eligible int;   -- ...after status + ownership
    already  int;   -- ...that already have an open template_changed rerender
    filed    int;
BEGIN
    SELECT count(DISTINCT p.id) INTO carrying
      FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE pc.component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551';

    SELECT count(DISTINCT p.id) INTO eligible
      FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE pc.component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
       AND p.rebuild_policy IS DISTINCT FROM 'owned'
       AND p.status = 'active';

    SELECT count(DISTINCT p.id) INTO already
      FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE pc.component_id = 'aa3e4b68-bcea-49ca-890a-c111acefa551'
       AND p.rebuild_policy IS DISTINCT FROM 'owned'
       AND p.status = 'active'
       AND EXISTS (SELECT 1 FROM site_work_items w
                    WHERE w.site_id = p.site_id AND w.item_type = 'page_rerender'
                      AND w.spec->>'page_id' = p.id::text
                      AND w.spec->>'reason' = 'template_changed'
                      AND w.status IN ('detected','triaged','claimed')
                      AND w.batch_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000683'::uuid);

    SELECT count(*) INTO filed
      FROM site_work_items WHERE batch_id = '00000000-0000-0000-0000-000000000683';

    IF carrying > 0 AND eligible = 0 THEN
        RAISE EXCEPTION 'ABORT: % page(s) carry the component but 0 are eligible — the status '
                        'or ownership predicate is selecting nothing. Read it before re-running.',
                        carrying;
    END IF;

    -- EXACT, not "more than zero": ON CONFLICT/NOT EXISTS can drop a SUBSET, and
    -- an un-filed page keeps the stale deck for ever with nothing surfaced
    -- (bug_historian, round 2; 016b §9 "err == nil is not work queued").
    IF filed <> eligible - already THEN
        RAISE EXCEPTION 'ABORT: filed % but expected % (eligible % minus % already queued) — '
                        'a subset was silently dropped', filed, eligible - already, eligible, already;
    END IF;

    IF EXISTS (SELECT 1 FROM site_work_items
                WHERE batch_id = '00000000-0000-0000-0000-000000000683'
                  AND (spec->>'reason' <> 'template_changed'
                       OR COALESCE(spec->>'page_name','') = ''
                       OR COALESCE(spec->>'page_id','')   = '')) THEN
        RAISE EXCEPTION 'ABORT: an item is missing reason/page_name/page_id, or carries a reason '
                        'that routes to assemble mode and would complete while changing nothing';
    END IF;

    RAISE NOTICE '683: filed % item(s); % page(s) carry the component, % eligible, % already queued',
                 filed, carrying, eligible, already;
END $$;

COMMIT;
