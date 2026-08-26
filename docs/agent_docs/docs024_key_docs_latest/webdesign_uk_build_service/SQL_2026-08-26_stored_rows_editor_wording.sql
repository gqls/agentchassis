-- SQL_2026-08-26 — the owner's editor wording applied to the STORED rows.
--
-- WHY. The 2026-08-25 direct edits fixed the SERVED pages only; page_components
-- kept the old sentences, and the overnight discovery sweep's assemble-rerender of
-- how-it-works (vm-sites 5a48e20) resurrected "Visual Studio Code" from the stored
-- row. Assemble-type rerenders read rendered_html; template regenerations read
-- content_data — so BOTH fields get the same five sentence replacements the pages
-- got, and rendered_html_digest is recomputed (verified md5(rendered_html) on the
-- live rows). Rows not naming VS Code are untouched.

BEGIN;

UPDATE page_components pc SET
  rendered_html = replace(replace(replace(replace(replace(pc.rendered_html,
    '<strong>Visual Studio Code</strong> is free and built for exactly this kind of work.',
    'Your own code editor (IDE) or your favourite AI tool are good for this.'),
    'Visual Studio Code is free and built for exactly this job, and',
    'Your own code editor (IDE) or your favourite AI tool are good for this, and'),
    ', and Visual Studio Code is free and built for exactly this job.',
    '. Your own code editor (IDE) or your favourite AI tool are good for this.'),
    'and Visual Studio Code for editing the files.',
    'and your own code editor (IDE) or favourite AI tool for editing the files.'),
    'Visual Studio Code is free and built for exactly this job. If',
    'Your own code editor (IDE) or your favourite AI tool are good for this. If'),
  content_data = replace(replace(replace(replace(replace(pc.content_data::text,
    '<strong>Visual Studio Code</strong> is free and built for exactly this kind of work.',
    'Your own code editor (IDE) or your favourite AI tool are good for this.'),
    'Visual Studio Code is free and built for exactly this job, and',
    'Your own code editor (IDE) or your favourite AI tool are good for this, and'),
    ', and Visual Studio Code is free and built for exactly this job.',
    '. Your own code editor (IDE) or your favourite AI tool are good for this.'),
    'and Visual Studio Code for editing the files.',
    'and your own code editor (IDE) or favourite AI tool for editing the files.'),
    'Visual Studio Code is free and built for exactly this job. If',
    'Your own code editor (IDE) or your favourite AI tool are good for this. If')::jsonb
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain = 'webdesign.uk'
  AND (pc.rendered_html LIKE '%Visual Studio Code%' OR pc.content_data::text LIKE '%Visual Studio Code%');

UPDATE page_components pc SET rendered_html_digest = md5(pc.rendered_html)
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain = 'webdesign.uk'
  AND pc.rendered_html_digest IS NOT NULL
  AND pc.rendered_html_digest <> md5(pc.rendered_html);

DO $chk$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='webdesign.uk' AND (pc.rendered_html LIKE '%Visual Studio Code%' OR pc.content_data::text LIKE '%Visual Studio Code%');
  IF n <> 0 THEN RAISE EXCEPTION '% stored row(s) still name Visual Studio Code', n; END IF;
  SELECT count(*) INTO n FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='webdesign.uk' AND pc.rendered_html ILIKE '%our own code editor (IDE)%';
  IF n < 4 THEN RAISE EXCEPTION 'expected the new wording in >=4 rows, found %', n; END IF;
  SELECT count(*) INTO n FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='webdesign.uk' AND pc.rendered_html_digest IS NOT NULL AND pc.rendered_html_digest <> md5(pc.rendered_html);
  IF n <> 0 THEN RAISE EXCEPTION '% row(s) with a stale digest', n; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: stored rows carry the owner wording, digests fresh';
END $chk$;

COMMIT;
