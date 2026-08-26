-- SQL_2026-08-25g — OWNER, 2026-08-25 (night): (1) editing guidance loses the named
-- third party: "Your own code editor (IDE) or your favourite AI tool are good for
-- this." (served pages edited directly the same hour, owner-sanctioned; this file
-- makes REBUILDS produce the same wording). (2) THE HOME PAGE CARRIES ALL THE
-- POINTS: index gains a short-bullets section (generic-text-block, right after the
-- hero) ending with a link to /what-you-get.html.
--
-- Pieces: third_party_options drops its Editing/VS Code sentence (back to SIX named
-- services); allowed_entities drops Visual Studio Code; writer_block: seven->six x3,
-- the editing list line, the how-to-edit paragraph wording, + the HOME-POINTS
-- paragraph; content_direction / mission_brief / submission share one VS Code
-- sentence, replaced by one needle each; the PLAN gains the index section (orderings
-- shifted) and pages.sections gains the slot.

BEGIN;

-- ── 1. evidence_base: third_party_options + allowed_entities + writer_block ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(
        jsonb_set(c.data, '{facts}', (
          SELECT jsonb_agg(
            CASE WHEN f->>'id'='third_party_options' THEN
              jsonb_set(jsonb_set(f,
                '{claim}', to_jsonb(replace(f->>'claim',
                  ' Editing the files: Visual Studio Code, a free editor that opens and edits the site files, which are plain HTML and CSS that any text editor can open.', ''))),
                '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                  $s$; Visual Studio Code REMOVED by owner wording 2026-08-25 night ("Your own code editor (IDE) or your favourite AI tool are good for this") - editing guidance is generic now, no named third party$s$))
            ELSE f END ORDER BY ord)
          FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
        )),
        '{allowed_entities}', (c.data->'allowed_entities') - 'Visual Studio Code'
      ),
      '{writer_block}',
      to_jsonb(
        replace(replace(replace(replace(replace(
          c.data->>'writer_block',
          'DO name the seven services',
          'DO name the six services'),
          'The seven and their groups, exactly and no others: For hosting the files: Cloudflare Pages and Netlify. For seeing who visits: Fathom Analytics and Plausible. For making a contact form work: Formspree and Basin. For editing the files: Visual Studio Code.',
          'The six and their groups, exactly and no others: For hosting the files: Cloudflare Pages and Netlify. For seeing who visits: Fathom Analytics and Plausible. For making a contact form work: Formspree and Basin. For editing, name no third party: say, in the owner''s wording, your own code editor (IDE) or your favourite AI tool are good for this.'),
          'The register already carries seven named third-party services by category',
          'The register already carries six named third-party services by category'),
          'the seven named third-party services (Cloudflare Pages, Netlify, Fathom Analytics, Plausible, Formspree, Basin, Visual Studio Code)',
          'the six named third-party services (Cloudflare Pages, Netlify, Fathom Analytics, Plausible, Formspree, Basin)'),
          'the files are plain HTML and CSS, so any text editor opens them; Visual Studio Code is free and made for exactly this, per third_party_options; and anyone the customer hires can take the files on, per taking_it_further.',
          'the files are plain HTML and CSS, so any text editor opens them; say, in the owner''s wording: your own code editor (IDE) or your favourite AI tool are good for this; and anyone the customer hires can take the files on, per taking_it_further.')
        || $wb$

THE HOME PAGE CARRIES ALL THE POINTS (owner ruling, 2026-08-25). The index has a short bullet section (the generic-text-block slot, straight after the hero). One short bullet each, plain words, one line apiece: who this is for; the ZIP of files that is yours to host, edit and maintain; that we are not a hosting company, and we can help you set it up on free hosting like Netlify; the 30 days the link stays live; the kinds of site that can be asked for, and that we do not build online shops that take payment; and how the site is edited afterwards: your own code editor (IDE) or your favourite AI tool are good for this. End the section with one link to /what-you-get.html, labelled plainly (for example: the full list of what you get). No selling, no softening: the bullets are the terms said small.$wb$
      )
    ) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25g: editing guidance generic (VS Code removed; six named services again); HOME-POINTS writer_block paragraph for the new index bullets section.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 2. content_direction (both copies), mission_brief, submission: one shared sentence ──
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='content_direction' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    replace(c.data::text,
      'Visual Studio Code is free, and any web developer can take them on.',
      'your own code editor (IDE) or your favourite AI tool are good for this, and any web developer can take them on.')::jsonb AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'content_direction', r.newdata, 'owner-ruling',
  'SQL_2026-08-25g: editing wording generic (both copies).', true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='mission_brief' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{text}', to_jsonb(replace(c.data->>'text',
      'Visual Studio Code is free, and any web developer can take them on.',
      'your own code editor (IDE) or your favourite AI tool are good for this, and any web developer can take them on.'))) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'mission_brief', r.newdata, 'owner-ruling',
  'SQL_2026-08-25g: editing wording generic.', true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='submission' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{mission_brief,text}', to_jsonb(replace(c.data->'mission_brief'->>'text',
      'Visual Studio Code is free, and any web developer can take them on.',
      'your own code editor (IDE) or your favourite AI tool are good for this, and any web developer can take them on.'))) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'submission', r.newdata, 'owner-ruling',
  'SQL_2026-08-25g: editing wording generic (embedded mission copy kept equal).', true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- ── 3. the PLAN: index gains the bullets section straight after the hero ──
-- (plan_id, page_name, ordering) is UNIQUE (idx_site_plan_sections_key), so a bare
-- +1 shift collides mid-statement (caught by the trial run). Two-step offset:
UPDATE site_plan_sections SET ordering = ordering + 100
 WHERE plan_id='6a3e6d1b-f149-44ce-9911-a031d4a6d222' AND page_name='index' AND ordering >= 1;
UPDATE site_plan_sections SET ordering = ordering - 99
 WHERE plan_id='6a3e6d1b-f149-44ce-9911-a031d4a6d222' AND page_name='index' AND ordering >= 100;
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name, palette_id, layout_id, typography_set_id)
SELECT '6a3e6d1b-f149-44ce-9911-a031d4a6d222', 'index', 1, 'generic-text-block', sib.palette_id, sib.layout_id, sib.typography_set_id
FROM site_plan_sections sib
WHERE sib.plan_id='6a3e6d1b-f149-44ce-9911-a031d4a6d222' AND sib.page_name='what-you-get' AND sib.component_name='generic-text-block';

UPDATE pages p SET sections = '["hero","generic-text-block","brief-explanation","chat-input-box","call-to-action"]'::jsonb
FROM sites s WHERE s.id=p.site_id AND s.domain='webdesign.uk' AND p.name='index';

-- ── GUARDS ──
DO $chk$
DECLARE wb text; pwb text; d jsonb; prev jsonb; n int; cd jsonb; mb text; sub text;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF strpos(prev::text,'Visual Studio Code')=0 THEN RAISE EXCEPTION 'control: VS Code already absent from evidence_base'; END IF;
  -- VS Code must be gone from every WRITER-VISIBLE surface (claims, writer_lines,
  -- writer_block). Source attributions keep the history and legitimately name it.
  PERFORM 1 FROM jsonb_array_elements(d->'facts') AS t(x)
   WHERE strpos(COALESCE(x->>'claim',''),'Visual Studio Code')>0
      OR strpos(COALESCE(x->>'writer_line',''),'Visual Studio Code')>0;
  IF FOUND THEN RAISE EXCEPTION 'a fact claim/writer_line still names Visual Studio Code'; END IF;
  IF strpos(d->>'writer_block','Visual Studio Code')>0 THEN RAISE EXCEPTION 'writer_block still names Visual Studio Code'; END IF;
  wb := d->>'writer_block'; pwb := prev->>'writer_block';
  IF strpos(wb,'THE HOME PAGE CARRIES ALL THE POINTS')=0 THEN RAISE EXCEPTION 'home-points paragraph missing'; END IF;
  IF strpos(wb,'seven')>0 THEN RAISE EXCEPTION 'a seven survives in writer_block'; END IF;
  IF strpos(wb,'your own code editor (IDE) or your favourite AI tool')=0 THEN RAISE EXCEPTION 'owner wording missing from writer_block'; END IF;
  IF (length(wb)-length(replace(wb,'—',''))) <> (length(pwb)-length(replace(pwb,'—',''))) THEN RAISE EXCEPTION 'em-dash count moved'; END IF;
  SELECT count(*) INTO n FROM (
    SELECT x->>'id' FROM jsonb_array_elements(prev->'facts') AS t(x)
    EXCEPT SELECT x->>'id' FROM jsonb_array_elements(d->'facts') AS t(x)) q;
  IF n <> 0 THEN RAISE EXCEPTION 'a fact vanished'; END IF;
  IF d->'banned_claims' <> prev->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed'; END IF;
  IF NOT (d->'allowed_entities' @> '["Netlify"]' AND NOT d->'allowed_entities' @> '["Visual Studio Code"]')
    THEN RAISE EXCEPTION 'allowed_entities wrong'; END IF;

  SELECT ss.data INTO cd FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='content_direction' AND ss.is_current;
  IF strpos(cd::text,'Visual Studio Code')>0 THEN RAISE EXCEPTION 'content_direction still names VS Code'; END IF;
  SELECT ss.data->>'text' INTO mb FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='mission_brief' AND ss.is_current;
  SELECT ss.data->'mission_brief'->>'text' INTO sub FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.aspect='submission' AND ss.is_current;
  IF strpos(mb,'Visual Studio Code')>0 OR sub <> mb THEN RAISE EXCEPTION 'mission/submission wrong or diverged'; END IF;

  SELECT count(*) INTO n FROM site_plan_sections WHERE plan_id='6a3e6d1b-f149-44ce-9911-a031d4a6d222' AND page_name='index';
  IF n <> 5 THEN RAISE EXCEPTION 'index plan sections: expected 5, got %', n; END IF;
  PERFORM 1 FROM site_plan_sections WHERE plan_id='6a3e6d1b-f149-44ce-9911-a031d4a6d222' AND page_name='index' AND ordering=1 AND component_name='generic-text-block';
  IF NOT FOUND THEN RAISE EXCEPTION 'new index section not at ordering 1'; END IF;
  SELECT count(DISTINCT ordering) INTO n FROM site_plan_sections WHERE plan_id='6a3e6d1b-f149-44ce-9911-a031d4a6d222' AND page_name='index';
  IF n <> 5 THEN RAISE EXCEPTION 'index orderings collide'; END IF;
  PERFORM 1 FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='webdesign.uk' AND p.name='index' AND jsonb_array_length(p.sections)=5 AND p.sections->>1='generic-text-block';
  IF NOT FOUND THEN RAISE EXCEPTION 'pages.sections not updated'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: VS Code gone everywhere, six services restored, home-points section planned';
END $chk$;

COMMIT;
