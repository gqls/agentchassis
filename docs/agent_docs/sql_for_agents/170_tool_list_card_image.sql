-- 170_tool_list_card_image.sql — give `tool-list` cards an image slot.
-- DB-only, data fix to a shared content_component template. Style follows 151.
--
-- WHY. Imagery phase I3/F3: the imagery loop now generates a per-page
-- "content hero" for TOOL pages and derives an 800x450 card from it
-- (check_content_image_missing gained a `tool` surface, commit 8b804bc27).
-- The card is already being handed to this component: `tool-list` sources its
-- items from `query.pages_where_type:tool`, and that resolver's shared
-- pageImageProjection returns an `image` key per row (card first, plan hero
-- fallback, "" when neither exists). The template simply never rendered it —
-- the tool directory is text-only cards while the imagery sits unused.
--
-- BLAST RADIUS (intended): tool-list is a global library component (no
-- site_id) used on 5 live pages across 3 sites (gamesdesign.co.uk, idea.uk,
-- robot-hands.com). The image is rendered ONLY inside {{if .image}}, so a site
-- whose tool pages have no imagery yet renders exactly as it does today —
-- this is inert until images land, not a visual change forced on every site.
-- The existing tl-card-icon is kept for imageless cards, so nothing loses its
-- current treatment.
--
-- Geometry: cards are 800x450 (16:9, storage.ImagePurposes["card"]), so the
-- media box uses aspect-ratio 16/9 + object-fit:cover and cannot letterbox or
-- stretch whatever it is given.
--
-- Rollback: the full pre-edit template is written to a doc_note below.

BEGIN;

DO $$
DECLARE
    cid       uuid := 'a68b52b7-61c5-4797-a701-8e8643684f75';  -- tool-list
    tmpl      text;
    schm      jsonb;
    old_card  text := E'      <article class="tl-card">\n        <div class="tl-card-icon" aria-hidden="true">';
    new_card  text := E'      <article class="tl-card">\n        {{if .image}}<div class="tl-card-media"><img src="{{.image}}" alt="{{.title}}" loading="lazy"></div>{{end}}\n        <div class="tl-card-icon" aria-hidden="true">';
    old_css   text := E'.tool-list-section .tl-card-icon {';
    new_css   text := E'.tool-list-section .tl-card-media { margin: -1.75rem -1.75rem 0.25rem; border-radius: var(--border-radius, 0.5rem) var(--border-radius, 0.5rem) 0 0; overflow: hidden; aspect-ratio: 16 / 9; background: var(--color-surface); }' || E'\n' ||
                       E'.tool-list-section .tl-card-media img { width: 100%; height: 100%; object-fit: cover; display: block; }' || E'\n' ||
                       E'.tool-list-section .tl-card-icon {';
    n         int;
BEGIN
    SELECT html_template, input_schema INTO tmpl, schm
      FROM content_components WHERE id = cid;
    IF tmpl IS NULL THEN
        RAISE EXCEPTION '170: content_component % not found', cid;
    END IF;

    -- Guards: each anchor must appear EXACTLY once, and the edit must not
    -- already be applied (re-running must be a no-op failure, not a double
    -- insert of the media block).
    IF strpos(tmpl, 'tl-card-media') > 0 THEN
        RAISE EXCEPTION '170: template already contains tl-card-media — already applied?';
    END IF;
    n := (length(tmpl) - length(replace(tmpl, old_card, ''))) / length(old_card);
    IF n <> 1 THEN
        RAISE EXCEPTION '170: expected the tl-card article opener exactly once, found % — template drifted, inspect', n;
    END IF;
    n := (length(tmpl) - length(replace(tmpl, old_css, ''))) / length(old_css);
    IF n <> 1 THEN
        RAISE EXCEPTION '170: expected the .tl-card-icon rule exactly once, found % — template drifted, inspect', n;
    END IF;

    -- Rollback record: the full pre-edit template.
    INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
    VALUES ('pipeline', 'tool-list',
        E'## 170 backup — tool-list html_template BEFORE the card image slot\n' ||
        E'Restore by setting content_components.id=' || cid || E' html_template to the block below.\n\n' ||
        E'```html\n' || tmpl || E'\n```\n' ||
        E'Categories: migration, backup',
        '["migration","backup"]'::jsonb, 'human', '170_tool_list_card_image');

    -- Version snapshot alongside the doc_note (component_versions is the
    -- machine-readable trail the regeneration paths already consult).
    INSERT INTO component_versions (component_id, version_number, html_template, input_schema,
                                    change_description, changed_by, change_source)
    SELECT cid,
           COALESCE((SELECT MAX(version_number) FROM component_versions WHERE component_id = cid), 0) + 1,
           tmpl, schm,
           '170: pre-edit snapshot before adding the tl-card-media image slot (imagery I3/F3)',
           '170_tool_list_card_image', 'migration';

    UPDATE content_components
    SET html_template = replace(replace(tmpl, old_css, new_css), old_card, new_card),
        -- Declare `image` on the item shape. The resolver already returns it;
        -- declaring it keeps the schema an honest description of the data the
        -- template consumes.
        input_schema = jsonb_set(schm, '{fields,items,items,image}',
                                 '{"type": "image"}'::jsonb, true),
        updated_at = now()
    WHERE id = cid;

    RAISE NOTICE '170: tool-list gained tl-card-media (% -> % chars)',
        length(tmpl), length(replace(replace(tmpl, old_css, new_css), old_card, new_card));
END $$;

-- Pipeline note.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 170: tool-list cards render their image
Observed: the imagery loop derives 800x450 cards for tool pages (I3/F3) and query.pages_where_type already returns an `image` per row, but the tool-list template had no img at all — the tool directory stayed text-only while the imagery went unused.
Root cause: missing template slot, not missing data. content-listing (the reference surface) already rendered {{.image}}; tool-list never did.
Fix: added a {{if .image}}-guarded .tl-card-media block (16/9, object-fit:cover) above the existing icon, plus its CSS, and declared `image` on the item schema. Global component: 5 live pages / 3 sites. Inert where tool pages have no imagery — the existing icon treatment is unchanged for imageless cards.
Verify: re-render a page carrying tool-list on a site with tool cards and confirm /assets/images/card-*.jpg refs appear in the served HTML.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

COMMIT;

-- Post-check.
DO $$
DECLARE tmpl text; schm jsonb;
BEGIN
    SELECT html_template, input_schema INTO tmpl, schm
      FROM content_components WHERE id = 'a68b52b7-61c5-4797-a701-8e8643684f75';
    IF strpos(tmpl, 'tl-card-media') = 0 THEN
        RAISE EXCEPTION '170: post-check failed — tl-card-media not present after update';
    END IF;
    IF strpos(tmpl, '{{if .image}}') = 0 THEN
        RAISE EXCEPTION '170: post-check failed — image guard not present';
    END IF;
    IF schm #>> '{fields,items,items,image,type}' IS DISTINCT FROM 'image' THEN
        RAISE EXCEPTION '170: post-check failed — items.image not declared in input_schema';
    END IF;
    RAISE NOTICE '170: post-check OK';
END $$;
