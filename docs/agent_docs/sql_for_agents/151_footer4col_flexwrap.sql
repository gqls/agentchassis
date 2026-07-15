-- 151_footer4col_flexwrap.sql — fix the DURABLE source of a mobile overflow.
-- DB-only, data fix to a shared content_component template.
--
-- (Number 151: 149 and 150 were taken by another workstream that landed while
-- the travelling-docs turns were in flight — there are two 149s in the ledger.)
--
-- WHY. Tier-4 acceptance found vonc.com's footer overflowing on mobile: the
-- widest offender is `div.footer-legal` (506px at a 390px viewport). Its rule is
-- `display:flex; gap:2rem;` with the DEFAULT `flex-wrap:nowrap`, so six legal
-- links refuse to wrap and spill the viewport. The rule lives in the
-- content_component template `footer-4-column` (id 09034086-a581-4bba-a5b4-
-- 760d863bb2df), which is the DURABLE source: site_components.rendered_html is a
-- rendered artifact regenerated from this template, so a patch there is wiped by
-- the next refresh_site_components (observed 2026-07-14/15). The fix must be here.
--
-- BLAST RADIUS (intended): footer-4-column is shared by 8 sites, so this fixes
-- the same footer overflow on all of them. It is a genuine template defect —
-- the sibling `site-footer` template already carries flex-wrap; this one never
-- did. Adding flex-wrap:wrap + justify-content:center is safe on desktop (the
-- row already fits, so nothing wraps) and only takes effect where it must.
--
-- Proven before shipping: injecting exactly this CSS into the live vonc page in
-- real headless Chromium took .footer-legal 506px -> 326px and document overflow
-- 58px -> 0 (T16 ProveFix). Deploy is via rerender; verified by re-running Tier-4
-- acceptance (mobile-fit@mobile must pass).

BEGIN;

DO $$
DECLARE
    cid      uuid := '09034086-a581-4bba-a5b4-760d863bb2df';
    tmpl     text;
    old_rule text := E'.footer-legal {\n    display: flex;\n    gap: 2rem;\n}';
    new_rule text := E'.footer-legal {\n    display: flex;\n    gap: 2rem;\n    flex-wrap: wrap;\n    justify-content: center;\n}';
    n        int;
BEGIN
    SELECT html_template INTO tmpl FROM content_components WHERE id = cid;
    IF tmpl IS NULL THEN
        RAISE EXCEPTION '151: content_component % not found', cid;
    END IF;

    -- Guard: the rule must be present EXACTLY once and not already fixed.
    n := (length(tmpl) - length(replace(tmpl, old_rule, ''))) / length(old_rule);
    IF n <> 1 THEN
        RAISE EXCEPTION '151: expected the .footer-legal rule exactly once, found % — template drifted, inspect', n;
    END IF;
    IF strpos(tmpl, 'flex-wrap') > 0 THEN
        RAISE EXCEPTION '151: template already contains flex-wrap — already applied?';
    END IF;

    -- Rollback record: the full pre-edit template, kept in a doc_note.
    -- (subject_type is constrained to tool|pipeline; use pipeline for a
    -- component-template backup, with the component in subject_key.)
    INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
    VALUES ('pipeline', 'footer-4-column',
        E'## 151 backup — footer-4-column html_template BEFORE flex-wrap fix\n' ||
        E'Restore by setting content_components.id=' || cid || E' html_template to the block below.\n\n' ||
        E'```html\n' || tmpl || E'\n```\n' ||
        E'Categories: migration, backup',
        '["migration","backup"]'::jsonb, 'human', '151_footer4col_flexwrap');

    UPDATE content_components
    SET html_template = replace(tmpl, old_rule, new_rule),
        updated_at = now()
    WHERE id = cid;

    RAISE NOTICE '151: footer-4-column .footer-legal now wraps (% -> % chars)',
        length(tmpl), length(replace(tmpl, old_rule, new_rule));
END $$;

-- Pipeline note.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 151: footer-4-column template — .footer-legal wraps on mobile
Observed: Tier-4 acceptance on vonc.com found div.footer-legal overflowing (506px at 390px); the shared footer-4-column content_component template set display:flex + gap:2rem with the default flex-wrap:nowrap, so 6 legal links spilled the viewport on 8 sites.
Root cause: a genuine template defect at the durable layer. Earlier fixer patches hit site_components.rendered_html (a rendered artifact) and were wiped by refresh_site_components — the fix must be in content_components.html_template.
Fix: added flex-wrap:wrap + justify-content:center to .footer-legal in footer-4-column (id 09034086). Safe on desktop (already fits); wraps only where it must. Full pre-edit template backed up in a doc_note.
Verified: this exact CSS took .footer-legal 506->326px and overflow 58->0 in live Chromium before shipping; deploy via rerender, confirmed by re-running Tier-4 acceptance (mobile-fit@mobile).
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE tmpl text;
BEGIN
    SELECT html_template INTO tmpl FROM content_components WHERE id = '09034086-a581-4bba-a5b4-760d863bb2df';
    IF strpos(tmpl, 'flex-wrap: wrap') = 0 THEN
        RAISE EXCEPTION '151: post-check failed — flex-wrap not present after update';
    END IF;
    RAISE NOTICE '151: post-check OK';
END $$;

COMMIT;
