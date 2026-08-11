-- FILE: docs/agent_docs/sql_for_agents/378_bugfix_238_restore_finetuning_index_structural_urls.sql
--
-- bugs_open/238 — restore the 11 resolver-sourced URL keys a content
-- regeneration dropped from finetuning.uk /index.html's case-studies-grid, and
-- seed the two site_specs aspects that make six of them resolve properly from
-- now on.
--
-- THE CODE FIX IS THE FIX. This file repairs one row's data. The write path is
-- fixed by the plan-time structural-key carry in plan_sections_action.go, which
-- is inert until the chassis image rolls. Applying this without that fix leaves
-- the page correct until the next CONTENT REGENERATION of this section (a
-- tone_shift / content_rewrite item), which would drop the same 11 keys again.
-- A no-LLM rerender (reason=section_data_resolved / image_landed) MERGES and
-- cannot lose them, which is why 378a queues that shape and not a rebuild.
--
-- WHAT WENT, AND WHY IT IS NOT A PASTE-BACK. The 2026-08-09 15:18Z tone_shift
-- run rewrote which case studies the five cards describe as well as dropping
-- the keys, so the historical card→image pairing no longer holds. Assignment
-- below is by SUBJECT, and each of the five assets is used exactly once:
--
--   card | new card subject (live copy)                        | asset
--   -----+-----------------------------------------------------+-------------------------
--     1  | quote-request automation, facilities provider        | case-study-facilities
--     2  | years of documents searchable in plain English       | case-study-legal-rag
--     3  | pulling company filings together                     | case-study-financial-data
--     4  | a team of agents handling exceptions, logistics ops  | case-study-logistics-strategy
--     5  | keeping client data off third-party servers          | case-study-private-ai
--
-- Corroborated independently by the regenerated alt texts, which the same run
-- wrote and which this file does not touch: card2's alt is "indexed documents
-- connected by thin linking lines" and card5's is "a secure enclosed network
-- with data held inside" — i.e. the alts already describe the assets assigned
-- here. Cards 2 and 4 are the JUDGEMENT CALLS and were put to the owner as
-- such (approved 2026-08-10): card 2's client changed (legal services →
-- professional services) while its subject stayed document search, and card 4's
-- subject is new (multi-agent exception handling) while its client is the
-- logistics one. If either reads wrong on the page, the fix is a new image, not
-- a re-shuffle — every other pairing is forced.
--
-- THE LINK FIELDS ARE NOT RESTORED TO THEIR OLD VALUES. All five historical
-- targets (/case-studies/<slug>.html) return HTTP 404 today and no such rows
-- exist in `pages`; restoring them would trade five vanished links for five
-- dead ones. Owner decision (2026-08-10): point all five at /case-studies.html,
-- which is active and serves 200. cta_link_url is restored to its genuine
-- stored value, /contact.html — an active page, and NOT the phantom-CTA default
-- of bugs_open/203 (that defect is a resolver FABRICATING /contact.html for an
-- unknown destination; this is a stored value being put back).
--
-- WHY THE FIVE IMAGE URLS STAY IN content_data RATHER THAN BECOMING RESOLVABLE.
-- All five card image fields declare ONE source, `site_assets.image`, so a
-- resolver cannot map five distinct assets through them; and the resolver only
-- ever reaches assets via site_plan_imagery joins (this site has none) — it
-- never looks an asset up by literal key. Re-sourcing the schema is out: the
-- component has no site_id and is shared by four pages across three sites. A
-- literal-asset-key resolver arm is the real answer and is named for the owner,
-- not built here. content_data IS the render source (PBP-014); the carry fix is
-- what makes that durable.
--
-- EXPECTED OUTCOME: the row goes from 47 keys to 58 (every field the template
-- references), with 11 non-empty URL values. The DO block ABORTS on any other
-- outcome — in both directions: still 47 means the UPDATE was inert, and a key
-- count that is not 58 means the template contract is not met. A verify block
-- of bare SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty
-- result), so the guard is a RAISE.
--
-- Rollback: 378_bugfix_238_restore_finetuning_index_structural_urls_ROLLBACK.sql

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Pre-flight. Read this before letting the transaction commit.
--    locked_at MUST be NULL: a locked row belongs to a human and neither
--    save_page_sections nor this file may overwrite it.
-- ---------------------------------------------------------------------------
\echo '=== BEFORE: the damaged row ==='
SELECT pc.id,
       pc.slot_name,
       pc.locked_at,
       (SELECT count(*) FROM jsonb_object_keys(pc.content_data)) AS key_count,
       (SELECT count(*) FROM jsonb_object_keys(pc.content_data) k WHERE k LIKE '%\_url') AS url_key_count
  FROM page_components pc
 WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

DO $$
DECLARE
    v_locked timestamptz;
    v_keys   int;
BEGIN
    SELECT pc.locked_at, (SELECT count(*) FROM jsonb_object_keys(pc.content_data))
      INTO v_locked, v_keys
      FROM page_components pc
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

    IF NOT FOUND THEN
        RAISE EXCEPTION '238: target page_components row not found — another session may have rebuilt the page; re-derive the row id before applying';
    END IF;
    IF v_locked IS NOT NULL THEN
        RAISE EXCEPTION '238: row is LOCKED (locked_at=%) — a human owns this content; do not overwrite', v_locked;
    END IF;
    IF v_keys = 58 THEN
        RAISE EXCEPTION '238: already applied — the row already carries all 58 keys';
    END IF;
    IF v_keys <> 47 THEN
        RAISE EXCEPTION '238: unexpected key count % (expected 47) — the row changed since this file was written; re-measure the diff before applying', v_keys;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 2. Belt-and-braces snapshot. The archive trigger (IMP-052 / bugs_open/229)
--    fires on UPDATE of rendered_html; this row's rendered_html is untouched
--    here, so take the content_data snapshot explicitly rather than assume it.
-- ---------------------------------------------------------------------------
-- NOTE the naming trap: page_component_history.component_id is a FK to
-- page_components(id) — the ROW — not to content_components(id), even though
-- page_components.component_id means the latter. It is ON DELETE SET NULL,
-- which is why every historic row here reads NULL: save_page_sections archives
-- the row and then DELETEs it, nulling the reference on the way out.
INSERT INTO page_component_history (component_id, page_id, site_id, content_data, source, slot_name, position, op)
SELECT pc.id,
       pc.page_id,
       p.site_id,
       pc.content_data,
       'manual_repair_238',
       pc.slot_name,
       pc.position,
       -- `op` is constrained to 'overwrite' | 'delete' | NULL (pch_op_check);
       -- 'UPDATE' fails the constraint and aborts the whole transaction.
       'overwrite'
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
 WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

-- ---------------------------------------------------------------------------
-- 3. The repair. `||` merges, so every LLM-written key the regeneration
--    produced (titles, excerpts, alts, stats — the rewrite that was correct)
--    is preserved untouched; only the 11 missing keys are added.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
   SET content_data = pc.content_data || jsonb_build_object(
           'card1_image_url', '/assets/images/case-study-facilities.jpg',
           'card2_image_url', '/assets/images/case-study-legal-rag.jpg',
           'card3_image_url', '/assets/images/case-study-financial-data.jpg',
           'card4_image_url', '/assets/images/case-study-logistics-strategy.jpg',
           'card5_image_url', '/assets/images/case-study-private-ai.jpg',
           'card1_link_url',  '/case-studies.html',
           'card2_link_url',  '/case-studies.html',
           'card3_link_url',  '/case-studies.html',
           'card4_link_url',  '/case-studies.html',
           'card5_link_url',  '/case-studies.html',
           'cta_link_url',    '/contact.html'
       ),
       updated_at = now()
 WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

-- ---------------------------------------------------------------------------
-- 4. Seed the two site_specs aspects the link fields declare as their source.
--    Neither aspect exists on this site — and neither has EVER existed on any
--    site fleet-wide, so `site_specs.case_studies.*` and `site_specs.pages.*`
--    have never resolved anywhere. Seeding them here means a future plan
--    resolves the six link fields live instead of relying on the carry.
--    UNIQUE (site_id, aspect) WHERE is_current — no existing row to supersede.
--
--    NOTE the name collision that is NOT one: `pages` is also a top-level
--    source type reading the real pages table (ensurePages). The schema field
--    declares `site_specs.pages.contact_url`, which routes to this ASPECT.
-- ---------------------------------------------------------------------------
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
VALUES
  ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'case_studies',
   jsonb_build_object(
       'card1_url', '/case-studies.html',
       'card2_url', '/case-studies.html',
       'card3_url', '/case-studies.html',
       'card4_url', '/case-studies.html',
       'card5_url', '/case-studies.html'
   ),
   'manual_repair', 'bugfix-238',
   'bugs_open/238: case-studies-grid declares site_specs.case_studies.cardN_url; the aspect never existed, so the five card links resolved to nothing and were silently dropped. Per-case-study pages do not exist (the historical /case-studies/<slug>.html targets 404), so all five point at the index. Replace with per-card URLs when individual case-study pages are built.'),
  ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'pages',
   jsonb_build_object('contact_url', '/contact.html'),
   'manual_repair', 'bugfix-238',
   'bugs_open/238: case-studies-grid declares site_specs.pages.contact_url; the aspect never existed, so the section CTA button was silently dropped (the template gates it). /contact.html is a live page on this site — this is a stored value restored, not the bugs_open/203 phantom default.');

-- ---------------------------------------------------------------------------
-- 5. Verify, and ABORT unless the row now satisfies the template contract.
--    58 keys is the template's full field list, measured from html_template.
-- ---------------------------------------------------------------------------
\echo '=== AFTER: key counts and the restored values ==='
SELECT (SELECT count(*) FROM jsonb_object_keys(pc.content_data)) AS key_count,
       pc.content_data->>'card1_image_url' AS card1_image_url,
       pc.content_data->>'card4_image_url' AS card4_image_url,
       pc.content_data->>'cta_link_url'    AS cta_link_url
  FROM page_components pc
 WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

DO $$
DECLARE
    v_keys      int;
    v_empty     int;
    v_bad_asset int;
    v_aspects   int;
BEGIN
    SELECT (SELECT count(*) FROM jsonb_object_keys(pc.content_data))
      INTO v_keys
      FROM page_components pc
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

    IF v_keys <> 58 THEN
        RAISE EXCEPTION '238: key count is % after repair, expected 58 — aborting', v_keys;
    END IF;

    -- every URL key must be present AND non-empty: an empty string satisfies
    -- `?` and still renders src="" — the very defect being repaired.
    SELECT count(*) INTO v_empty
      FROM page_components pc,
           LATERAL jsonb_each_text(pc.content_data) kv
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7'
       AND kv.key LIKE '%\_url'
       AND COALESCE(btrim(kv.value), '') = '';
    IF v_empty > 0 THEN
        RAISE EXCEPTION '238: % URL key(s) are present but empty — aborting', v_empty;
    END IF;

    -- every restored image path must name an ACTIVE asset of this site, so a
    -- typo cannot commit: this is the check that would have caught a paste-back
    -- pointing at a file that does not exist.
    SELECT count(*) INTO v_bad_asset
      FROM page_components pc,
           LATERAL jsonb_each_text(pc.content_data) kv
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7'
       AND kv.key LIKE 'card%\_image\_url'
       AND NOT EXISTS (
             SELECT 1 FROM assets a
              WHERE a.site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
                AND a.status = 'active'
                AND kv.value = '/assets/images/' || a.asset_key || '.jpg');
    IF v_bad_asset > 0 THEN
        RAISE EXCEPTION '238: % card image URL(s) name no active asset on this site — aborting', v_bad_asset;
    END IF;

    SELECT count(*) INTO v_aspects
      FROM site_specs
     WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
       AND is_current AND aspect IN ('case_studies', 'pages');
    IF v_aspects <> 2 THEN
        RAISE EXCEPTION '238: expected 2 seeded site_specs aspects, found % — aborting', v_aspects;
    END IF;

    RAISE NOTICE '238: repair verified — 58 keys, no empty URL, 5 image URLs resolve to active assets, 2 aspects seeded';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- NEXT: the page must be RE-RENDERED for any of this to reach the artefact.
-- 378a_bugfix_238_rerender_finetuning_index.sql queues the no-LLM shape.
-- Authority is the served page, never this row:
--   curl -s https://finetuning.uk/index.html > /tmp/ft.html
--   grep -c 'csg-card-image" src=""'          /tmp/ft.html   # want 0
--   grep -c 'src="/assets/images/case-study-' /tmp/ft.html   # want 5
--   grep -c '<a class="csg-card-link" href="' /tmp/ft.html   # want 5
--   grep -c '<a class="csg-cta-btn" href="'   /tmp/ft.html   # want 1
-- Grep the ANCHOR, never the bare class name — the class appears in the
-- component's own <style> block, so a class grep returns hits with zero links.
-- ---------------------------------------------------------------------------
