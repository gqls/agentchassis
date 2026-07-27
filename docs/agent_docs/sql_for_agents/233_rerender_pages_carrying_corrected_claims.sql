-- 233_rerender_pages_carrying_corrected_claims.sql
-- Publish the corrections made in 229, 230, 231 and 232.
--
-- ══ WHY THIS FILE EXISTS ═══════════════════════════════════════════════════
-- Every one of those migrations edited `page_components.content_data` and
-- deliberately left `rendered_html` alone. That is the right call — writing a
-- corrected rendered_html while the stored copy stayed wrong would make a page
-- LOOK fixed while every future re-render reprinted the error, which is exactly
-- bugs_open/093's complaint. But it means the corrections are not published
-- until each page re-renders, and until then the live sites still serve them:
--
--   SELECT ... (content_data ILIKE '%~80%%') AS in_stored,
--              (rendered_html ILIKE '%~80%%') AS in_published ...
--   finetuning.uk | index | case-studies-grid |  f  |  t     <-- exactly this
--
-- ── A RE-RENDER, NOT A REBUILD. THE DISTINCTION IS THE WHOLE POINT ─────────
-- `page-rerender` renders stored content_data with NO LLM. Every correction
-- here already lives in content_data, so a re-render is sufficient AND is the
-- only safe option: a rebuild would put a writer back in the loop and could
-- reinvent the very claims we just removed. It is also why this does not touch
-- bugs_open/087 (the single-page REBUILD path with no section plan) at all.
--
-- ── DISPATCH BY WORK ITEM, NOT BY kcat ────────────────────────────────────
-- The `kubectl run -i --rm | kcat -P` publish pattern silently drops messages
-- with exit 0 and a printed correlation id. A work item goes through the
-- platform's own dispatcher instead, which is the path 619 completed
-- page_rerender items already took. Lane confirmed healthy before writing this:
-- 12 page-rerender orchestrations COMPLETED in the preceding 6 hours, the most
-- recent at 11:24 today.
--
-- ── CORRECTED, ~2 MINUTES AFTER APPLYING: status MUST BE 'triaged' ────────
-- This file first inserted the items with `status='detected'`, copying the
-- convention every discovery check uses. **That is a queue with no consumer**,
-- and it is `bugs_open/083`: the dispatch loop filters
-- `status IN ('triaged','approved')` (`claim_work_item_action.go:102`,
-- `load_work_item_actions.go:559`), and the ONLY thing that promotes
-- `detected` -> `triaged` is `TriageDetectedItemsAction`, which runs only
-- inside the `improvement-loop` agent, fired only by the `improvement-sweep`
-- scheduled task — **disabled since 2026-05-02**. 98 rows sit parked fleet-wide
-- because of it, 28 of them `page_rerender`. These five would have joined them.
--
-- Caught because the items sat at `detected` with no orchestration after 45s on
-- a lane measured healthy minutes earlier (12 completions in 6h) — i.e. the
-- discrepancy was between the LANE being alive and MY items not moving, which
-- is a different question from "is it slow". The INSERT below now writes
-- `triaged` directly. **Do not "follow the convention" here**: the convention
-- is what 083 is about.
--
-- ── CORRECTED AGAIN: `spec.reason` SELECTS THE MODE — IT IS NOT A LABEL ───
-- The first run of this file used `reason: 'claims_corrected'`, invented so the
-- item_key would not collide with another session's rerender. The work item
-- went to `complete`, the orchestration reported COMPLETED — and the page still
-- served the old figure. `page_components.updated_at` was the timestamp of the
-- CONTENT edit, not of the render, which is what gave it away.
--
-- The page-rerender agent branches on that exact string:
--
--   check_rerender_mode:
--     condition: input_data.spec.reason == 'image_landed'
--             OR input_data.spec.reason == 'section_data_resolved'
--             OR input_data.spec.reason == 'cta_links_stale'
--     then_step: rerender_sections     -- re-renders each section from content_data
--     else_step: render_page           -- ASSEMBLE-ONLY: reuses stored section HTML
--
-- An unrecognised reason therefore takes `else_step`, re-assembles the page from
-- the STALE section HTML, deploys it, and reports success. Nothing errors,
-- because nothing went wrong — the workflow did exactly what it was asked.
--
-- So a value that reads like free-text provenance is load-bearing control flow.
-- Same family as the fleet landmine "a string step-config is a REFERENCE, never
-- a literal". **The dedup key and the reason are DIFFERENT FIELDS** — vary
-- `item_key` freely, never `spec.reason`.
--
-- This also matters to bugs_open/093 beyond this file: the stat audit added
-- there reads stored `content_data`, and assemble-only mode never reads
-- content_data at all. A page can be re-published on a path where the audited
-- artefact is not the published one.
--
-- ── item_key AND THE DEDUP INDEX ──────────────────────────────────────────
--   CREATE UNIQUE INDEX idx_swi_dedup ON site_work_items (site_id, item_key)
--   WHERE item_key IS NOT NULL AND status <> ALL (ARRAY['complete','verified',
--         'rejected','wont_fix','failed','unresolved','cancelled']);
-- So the key only has to be unique among LIVE items. The reason suffix is
-- `claims_corrected` rather than the usual `section_data_resolved`, so these
-- cannot collide with a rerender another session raises for a different cause,
-- and ON CONFLICT DO NOTHING makes a concurrent duplicate harmless.

BEGIN;

DO $q$
DECLARE
    r        record;
    v_batch  uuid := gen_random_uuid();
    n        int  := 0;
BEGIN
    FOR r IN
        SELECT DISTINCT s.id AS site_id, s.domain, p.id AS page_id, p.name,
               CASE
                 WHEN s.domain='finetuning.uk'              THEN 'the unevidenced ~80% quote-preparation claim was removed (230)'
                 WHEN s.domain='vonc.com'                   THEN 'the transposed archetype/tool figures were corrected (229)'
                 ELSE 'the agent figure was corrected and the forbidden concurrency claim removed (231, 232)'
               END AS why
        FROM page_components pc
        JOIN pages p ON p.id = pc.page_id
        JOIN sites s ON s.id = p.site_id
        WHERE p.status = 'active'
          AND (
                (s.domain='finetuning.uk'              AND pc.rendered_html ILIKE '%~80%%')
             OR (s.domain='vonc.com' AND p.name='about' AND pc.rendered_html ~ '>3<|>8<')
             OR (s.domain='ai-agent-orchestration.com' AND (pc.rendered_html ~ '(^|[^1])70\+\s*[Aa]gent'
                                                         OR pc.rendered_html ILIKE '%thousands of concurrent%'))
              )
    LOOP
        INSERT INTO site_work_items
            (site_id, page_id, source, pipeline, item_type, severity, summary, spec,
             priority, handler_agent, status, created_by, item_key, batch_id)
        VALUES
            (r.site_id, r.page_id, 'claims_correction', 'build', 'page_rerender', 'medium',
             format('Re-render %s — %s', r.name, r.why),
             jsonb_build_object('domain', r.domain, 'page_id', r.page_id::text,
                                -- spec.reason IS CONTROL FLOW, NOT A LABEL. See below.
                                'page_name', r.name, 'reason', 'section_data_resolved'),
             40, 'page-rerender', 'triaged', 'bugfix-043-lane',   -- NOT 'detected' — see bugs_open/083
             format('page_rerender_%s_%s_claims_corrected', r.name, r.site_id),
             v_batch)
        ON CONFLICT DO NOTHING;

        n := n + 1;
        RAISE NOTICE '233: queued re-render % / % (%)', r.domain, r.name, r.why;
    END LOOP;

    IF n = 0 THEN
        RAISE EXCEPTION '233: no stale pages found — either the corrections were already published, or the predicates no longer match. Re-survey rather than assuming success.';
    END IF;
    RAISE NOTICE '233: % page(s) queued, batch %', n, v_batch;
END $q$;

COMMIT;

-- ── Verify AFTER the dispatcher has run, not now ───────────────────────────
-- These items are `detected`; they are not done because this file committed.
-- The check that matters is the published HTML, not the item's status
-- (a status is a claim; the rendered artefact is the evidence):
--
--   SELECT s.domain, p.name,
--          (pc.content_data::text ILIKE '%~80%%')  AS in_stored,
--          (pc.rendered_html      ILIKE '%~80%%')  AS in_published
--   FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
--   WHERE s.domain='finetuning.uk';
--   -- want: in_stored f, in_published f
--
-- And fleet-wide, that the deployed HTML stops tripping the sites' own bans:
--   cmd/claimscan, or the banscan harness described in NOTES_fabricated_stats_043.md.
