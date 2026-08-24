-- 578 — re-type the mislabelled tool rows (bugs_open/357 phase 3, RFC_046)
--
-- ⚠⚠ HELD, AND HELD HARDER THAN 577. This is the only file in this lane that
-- changes what a LIVE SITE is built from. Do not apply it from a runner, and do
-- not apply it because it is "next". Every precondition below must be TRUE and
-- CHECKED on the day, not assumed from this comment.
--
-- ============================================================================
-- PRECONDITIONS — each is measurable, none is assumable
-- ============================================================================
--
--  1. Phases 0 and 2 are COMMITTED, BUILT, ROLLED and verified at the artefact.
--     Not "merged": rolled. `make build-*` builds from HEAD and a same-tag
--     rebuild serves the node's cached binary, so verify the running service.
--
--  2. The stamp is LIVE AND READABLE — the owner's own words, and the condition
--     on which the repair was authorised in principle on 2026-08-22. Prove it on
--     the shape this migration creates, not in general:
--        SELECT count(*) FROM page_components pc
--         JOIN content_components cc ON cc.id = pc.component_id
--        WHERE cc.function = 'adopted-fragment' AND pc.component_version_id IS NOT NULL;
--     A post-roll, organically adopted row with a non-NULL stamp must exist.
--     Zero means adoption has never actually completed in production and this
--     migration would be creating rows of a shape nothing has demonstrated.
--
--  3. Migration 577 has been applied and `adopt_unidentified_fragments` is ARMED
--     on the carriers that rebuild these pages. If the producer is not fixed, a
--     repaired row is re-mislabelled by the next rebuild and this is wasted work:
--     the population is SELF-RENEWING (12 of the 22 rows were born on the single
--     day 2026-08-23).
--
--  4. A canary rebuild has run on a page adopted at BIRTH by phase 2, and the
--     conservation loop preserved it: bytes identical, component still
--     adopted-fragment, row count unchanged.
--
--  6. SCOPE, corrected 2026-08-24 on the owner's instruction: rebuild_policy='owned'
--     pages ARE included. The first draft skipped them on my mistaken reading that
--     'owned' meant a person had claimed the page. It does not — it means the page
--     belongs to a tool/widget, it is set in code, and 172 of 704 pages carry it.
--     They are also the only rows phase 2 can never repair, because the save is
--     refused before adoption runs, so this migration is their sole route.
--
--  5. RE-CENSUS ON THE DAY. Do not trust the number 22 — it was true on
--     2026-08-23 and the population mints daily. This migration therefore selects
--     its targets by PREDICATE, never by a pasted id list, and prints what it
--     matched before it writes.
--
-- ============================================================================
-- WHAT IT DOES, AND THE THREE THINGS IT REFUSES TO TOUCH
-- ============================================================================
--
-- Sets, per row: component_id -> adopted-fragment, content_data ->
-- {"body": <the stored bytes>}, component_version_id -> the {{.body}} version.
--
-- LEAVES ALONE: slot_name, position, rendered_html, rendered_html_digest, and
-- pages.sections. That is not caution, it is the entire safety argument. Layer 2
-- matches stored rows to incoming ones on slot-name equality and nothing else, so
-- renaming a slot makes the next rebuild miss the match, take the re-append arm,
-- and leave the page with the tool AND a freshly generated hero band — visible on
-- four live sites, and invisible to any check that only asks "is the tool still
-- there?". Because no name moves here, that landmine is never armed.
--
-- The bytes are untouched, so the page serves exactly what it served before. What
-- changes is what the row CLAIMS: today it says "I am the shared hero component"
-- (so the schema check demands a headline, the router parks it as no_content_data,
-- and any repair that gave it content_data would regenerate a 2KB title band over
-- a 16KB tool); afterwards it says "these bytes are my content", which is true and
-- regenerable.
--
-- rendered_html_digest stays valid and becomes HONEST: the row genuinely is
-- reproducible from its content_data once body holds the bytes.
--
-- AND IT CANNOT TRIP THE DIVERGENCE GUARDS — verified 2026-08-24 at their source,
-- prompted by the bugs_open/283 lane, because this was an unexamined risk here.
-- Both readers of rendered_html_digest compare the stored digest against the
-- STORED bytes, never against a fresh render:
--   page_component_divergence.go:68   pc.rendered_html_digest <> md5(pc.rendered_html)
--   site_component_divergence.go:70/120  SELECT rendered_html_digest, md5(rendered_html)
--   livespec.go:203/215                OLD.rendered_html_digest = md5(OLD.rendered_html)
-- What they detect is a HAND-EDITED row: bytes changed without the digest
-- following. This migration moves NEITHER, so the pair stays in lockstep and the
-- guards see nothing — which is the property that makes a byte-preserving re-type
-- invisible to them, rather than something they would report as tampering.
--
-- ============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 0. Backup EVERY affected row first, whole. The bug file's own verification
--    standard is the artefact, not the item, so the diff base has to be real.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS page_components_backup_357_20260823 AS
SELECT * FROM page_components WHERE false;

-- ---------------------------------------------------------------------------
-- 1. Preconditions, as refusals rather than as reading. A verify block made of
--    SELECTs cannot stop a COMMIT — ON_ERROR_STOP ignores a non-empty result —
--    so every check here RAISEs (the RFC_006 trap).
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    adopted_id   uuid;
    adopted_tpl  text;
    version_id   uuid;
    n_target     int;
    n_adopted    int;
    skipped      record;
BEGIN
    SELECT id, html_template INTO adopted_id, adopted_tpl
      FROM content_components
     WHERE function = 'adopted-fragment' AND is_active = true;
    IF adopted_id IS NULL THEN
        RAISE EXCEPTION 'migration 577 has not been applied: no active adopted-fragment component';
    END IF;
    IF btrim(adopted_tpl) <> '{{.body}}' THEN
        RAISE EXCEPTION 'the adopted-fragment template is %, not the identity function. Re-typing rows '
                        'onto a component that would NOT regenerate their bytes is the defect, not the fix.',
                        btrim(adopted_tpl);
    END IF;

    -- Precondition 2, enforced rather than trusted: adoption must have actually
    -- happened in production at least once.
    SELECT count(*) INTO n_adopted
      FROM page_components pc
     WHERE pc.component_id = adopted_id AND pc.component_version_id IS NOT NULL;
    IF n_adopted = 0 THEN
        RAISE EXCEPTION 'no organically adopted row with a provenance stamp exists yet. The owner authorised '
                        'this repair only AFTER the stamp is live and readable; zero here means it has never '
                        'completed in production, so this migration would be minting an unproven shape.';
    END IF;

    -- The version row the repaired rows will point at: get-or-create, keyed on the
    -- template TEXT exactly as resolveComponentVersionID keys it, so hand-repaired
    -- rows and render-stamped rows converge on ONE version rather than two.
    SELECT id INTO version_id
      FROM component_versions
     WHERE component_id = adopted_id AND html_template = adopted_tpl
     ORDER BY version_number DESC LIMIT 1;
    IF version_id IS NULL THEN
        INSERT INTO component_versions
            (component_id, version_number, html_template, input_schema, change_description, change_source)
        SELECT adopted_id,
               COALESCE(MAX(version_number), 0) + 1,
               adopted_tpl,
               (SELECT input_schema FROM content_components WHERE id = adopted_id),
               'observed at repair — bugs_open/357 phase 3 re-type, not an edit',
               'retype_357'
          FROM component_versions WHERE component_id = adopted_id
        RETURNING id INTO version_id;
    END IF;

    -- ------------------------------------------------------------------
    -- 2. The targets, by predicate. This is the bug file's own re-runnable
    --    population test — a row bound to `hero` whose stored bytes do not
    --    contain hero's own static template prefix — narrowed further to rows
    --    that actually hold an interactive fragment, so a merely drifted
    --    template can never be swept in.
    -- ------------------------------------------------------------------
    -- The interactivity test is the estate's OWN definition, mirrored from
    -- interactiveStructuralMarkers / interactiveControlMarkers
    -- (save_page_sections_action.go:1801, rendered into SQL by interactiveHTMLSQL).
    -- Not a fresh predicate: a second spelling of "is this a tool" is exactly the
    -- drift this lane exists to stop, and a narrower one silently drops rows.
    -- ⚠ Measured 2026-08-23 while writing this: an earlier draft tested only
    -- ILIKE '%tool-page%' AND excluded owned pages, and selected 16 of the 22 —
    -- silently dropping the six pages the bug was FILED about. Print, then write.
    CREATE TEMP TABLE candidates_357 ON COMMIT DROP AS
    SELECT pc.id, pc.page_id, pc.rendered_html,
           COALESCE(p.rebuild_policy, '') = 'owned' AS is_owned
      FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
      JOIN pages p ON p.id = pc.page_id
     WHERE cc.name = 'hero'
       AND position(left(cc.html_template, position('{{' in cc.html_template) - 1) in pc.rendered_html) = 0
       AND (
             (pc.rendered_html ILIKE '%<canvas%' OR pc.rendered_html ILIKE '%game-container%'
              OR pc.rendered_html ILIKE '%tool-page%' OR pc.rendered_html ILIKE '%data-tool%')
             OR (pc.rendered_html ILIKE '%<script%' AND (
                   pc.rendered_html ILIKE '%<input%' OR pc.rendered_html ILIKE '%<select%'
                OR pc.rendered_html ILIKE '%<textarea%' OR pc.rendered_html ILIKE '%oninput=%'
                OR pc.rendered_html ILIKE '%onchange=%' OR pc.rendered_html ILIKE '%onclick=%'))
           )
       -- Never touch a row that already declares itself something: an attribute is
       -- evidence about the bytes and outranks this repair.
       AND pc.rendered_html !~ 'data-component="[^"]+"';

    -- OWNED PAGES ARE INCLUDED (owner instruction, 2026-08-24), and they are the
    -- rows that need this migration MOST. Still listed, because a reader should see
    -- which rows are repaired under a guard rather than have to infer it.
    --
    -- CORRECTION: an earlier draft skipped them, describing rebuild_policy='owned'
    -- as "a human has claimed the page". THAT WAS WRONG. The guard's own words are
    -- that such a page "belongs to a tool/widget or is a runtime-fill shell"
    -- (save_page_sections_action.go:172), and the flag is set in code --
    -- create_report_page_action.go:176 writes it outright. Measured in that guard's
    -- own comment: 172 of 704 pages estate-wide are owned. It is a CATEGORY, not a
    -- claim, and nobody chose it page by page.
    --
    -- AND THE MISREADING INVERTED THE CONCLUSION. These six are the only rows phase
    -- 2 can NEVER heal: the owned-page guard returns at
    -- save_page_sections_action.go:186 and adoption runs at :397, so the save is
    -- refused two hundred lines before adoption is reached. No rebuild will ever
    -- type them correctly. A migration is their sole route -- they were the last
    -- rows I was willing to touch and they are the only ones that cannot fix
    -- themselves.
    --
    -- AND THEY ARE NOT VERBATIM PAGES, which is the one thing that WOULD make this
    -- unsafe -- checked rather than assumed [MEASURED 2026-08-24]. A page ships
    -- verbatim when THREE things hold: rebuild_policy='owned' AND exactly one
    -- component row AND that row carries content_data->>'deploy_mode'='verbatim'
    -- (the rule this file's own addendum records from the loancalculator_couk
    -- decompose lane). All six are owned with exactly one row -- two of three -- and
    -- every one reads deploy_mode = NONE. They are ASSEMBLED pages that work because
    -- assembly emits the single row's stored HTML, which is what the bug file says.
    --
    -- That distinction is load-bearing because the flip between verbatim and
    -- assembled IS THE ROW COUNT, not a flag: adding a row beside a verbatim one
    -- silently switches the page to assembly with the old full document still in the
    -- mix, producing a document nested inside a document. This migration adds no
    -- rows and sets no deploy_mode, so it cannot flip anything either way.
    --
    -- The three genuinely verbatim loancash pages are excluded STRUCTURALLY, not by
    -- luck: none is bound to the `hero` component at all, so the predicate's first
    -- clause rules them out. Verified 2026-08-24 rather than inferred from the
    -- earlier lane's exemption note.
    --
    -- WHY REPAIRING A ROW ON AN OWNED PAGE IS SAFE. The guard exists to stop the
    -- generic pipeline's DELETE-and-reinsert of page_components clobbering a tool
    -- page (the TL-001 shape). This migration does not do that: it UPDATEs three
    -- columns and never touches rendered_html, position or slot_name, so the
    -- operation the guard protects against is not the operation performed here.
    -- And because the pipeline refuses these pages, a row repaired here STAYS
    -- repaired -- there is no rebuild to undo it, which makes them the most durable
    -- targets in the population rather than the riskiest.
    FOR skipped IN
        SELECT p.name AS page, s.domain
          FROM candidates_357 c JOIN pages p ON p.id = c.page_id JOIN sites s ON s.id = p.site_id
         WHERE c.is_owned ORDER BY s.domain, p.name
    LOOP
        RAISE NOTICE 'INCLUDED (rebuild_policy=owned means it belongs to a tool, NOT human-claimed; '
                     'phase 2 can never reach it, so this is its only repair): %/%',
                     skipped.domain, skipped.page;
    END LOOP;

    CREATE TEMP TABLE targets_357 ON COMMIT DROP AS
    SELECT id, page_id, rendered_html FROM candidates_357;

    SELECT count(*) INTO n_target FROM targets_357;
    RAISE NOTICE 'bugs_open/357 phase 3: % candidate(s), % skipped as owned, % to repair on this run',
                 (SELECT count(*) FROM candidates_357),
                 (SELECT count(*) FROM candidates_357 WHERE is_owned),
                 n_target;
    IF n_target = 0 THEN
        RAISE EXCEPTION 'zero rows matched. Either the producer fix has already drained the population '
                        '(good — nothing to do, do not force this) or the predicate no longer describes it '
                        '(bad — re-derive it before writing anything).';
    END IF;

    -- 3. Backup, then re-type. Backup first, always.
    INSERT INTO page_components_backup_357_20260823
    SELECT pc.* FROM page_components pc JOIN targets_357 t ON t.id = pc.id;

    UPDATE page_components pc
       SET component_id         = adopted_id,
           content_data         = jsonb_build_object('body', pc.rendered_html),
           component_version_id = version_id
      FROM targets_357 t
     WHERE pc.id = t.id;

    -- 4. Prove the bytes did not move. This is the assertion the whole design
    --    rests on, and it is cheap: compare against the backup taken moments ago.
    IF EXISTS (
        SELECT 1
          FROM page_components pc
          JOIN page_components_backup_357_20260823 b ON b.id = pc.id
         WHERE md5(COALESCE(pc.rendered_html, '')) <> md5(COALESCE(b.rendered_html, ''))
            OR pc.slot_name IS DISTINCT FROM b.slot_name
            OR pc.position  IS DISTINCT FROM b.position
    ) THEN
        RAISE EXCEPTION 'a repaired row changed its bytes, slot_name or position. This repair is '
                        'byte-preserving by definition; rolling back.';
    END IF;

    -- 5. And prove the row is now regenerable, which is the point of the exercise:
    --    content_data.body must equal the stored bytes for every repaired row.
    IF EXISTS (
        SELECT 1 FROM page_components pc JOIN targets_357 t ON t.id = pc.id
         WHERE pc.content_data->>'body' IS DISTINCT FROM pc.rendered_html
    ) THEN
        RAISE EXCEPTION 'a repaired row is not reproducible from its own content_data';
    END IF;
END $$;

COMMIT;

-- ============================================================================
-- AFTERWARDS — verify at the ARTEFACT, not at this transaction
-- ============================================================================
-- A committed UPDATE is not a repaired page (bugs_closed/287). Then:
--
--   * curl each affected page and assert its own markup is still there —
--     class="tool-page", its controls, its <script>. Not "the item closed".
--   * re-run the bug file's population query: the repaired rows must be gone
--     from it.
--   * let ONE rebuild run on a repaired page and compare BEFORE/AFTER row counts
--     and per-row md5. A count that went UP by one is the carry-forward landmine
--     firing through a path this migration believes it never touched — and it is
--     invisible to any check that only asks whether the tool is still present.
--   * confirm the 9 false `required_fields_missing` items about `hero` on tool
--     pages stop being re-filed.
-- ============================================================================
