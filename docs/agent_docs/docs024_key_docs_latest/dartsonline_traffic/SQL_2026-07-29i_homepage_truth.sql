-- The homepage is the last page carrying shop language, and the most-visited one.
-- Live now: hero says "we've got the barrels, shafts, flights, and boards to match how
-- you play" and a section is literally headed "What We Stock". We have none of it.
-- Same treatment as about/shipping-returns: an explicit page_spec.purpose (read into the
-- content brief by save_page_sections_action.go:462-466) plus a rebuild against the
-- corrected content_direction.
BEGIN;
UPDATE pages
SET title = 'Darts Online | Spec-First Darts Buying Guides & News',
    meta_description = 'Spec-first darts buying guides: barrel weight, tungsten '
      || 'percentage, shaft length and flight shape, explained by what each one changes '
      || 'about your throw. Plus darts news from the PDC circuit.',
    page_spec = COALESCE(page_spec,'{}'::jsonb) || jsonb_build_object(
      'purpose',
      'The front door of a darts publication, not a shop window. Lead with what a player '
      || 'gets here: guides that explain what barrel weight, tungsten percentage, shaft '
      || 'length and flight shape actually change about the throw, plus darts news. Point '
      || 'readers at the guides and the news, which are the things this site has. '
      || 'FORBIDDEN, and currently present on this page: "we''ve got the barrels, shafts, '
      || 'flights and boards", a section headed "What We Stock", and any suggestion of a '
      || 'catalogue, range, stock or checkout. This site holds no stock and sells nothing; '
      || 'that independence is the reason to trust the advice, so say it rather than hide it.'
    ),
    updated_at = now()
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND name='index';

INSERT INTO site_work_items
  (site_id, item_type, item_key, status, pipeline, priority, handler_agent, source, spec, created_by, summary)
VALUES ('5fe8785b-223d-41a3-88ee-c07187622381','needs_page',
        'truth_reset:index:5fe8785b-223d-41a3-88ee-c07187622381','triaged','build',40,
        'page-build-handler','dartsonline-traffic-workstream',
        jsonb_build_object('reason','identity_corrected',
          'plan_id','0fb05b75-04f4-4f4c-8890-c34d6a71012c',
          'page_name','index','page_role','landing',
          'note','Homepage still says "What We Stock". Rewrite against corrected specs + new page_spec.purpose.'),
        'dartsonline-traffic-workstream',
        'Rebuild homepage — hero and a section heading still claim stock');
COMMIT;
SELECT name, title, left(page_spec->>'purpose',60) AS purpose FROM pages
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND name='index';
