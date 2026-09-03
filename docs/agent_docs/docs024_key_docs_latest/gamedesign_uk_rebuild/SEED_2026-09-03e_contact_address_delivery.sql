\set ON_ERROR_STOP on
-- SEED_2026-09-03e — deliver the owner's new contact address to the three components that
-- still carry the old one.
--
-- Owner 2026-09-03 midday: the contact page stays, the address is gamedesignuk@contactforsales.com.
-- SEED_2026-09-03c updated the SPECS, but about and contact were 'deployed' (not needs_rebuild),
-- so the 14:15Z re-plan preserved them and their copy was never regenerated. Measured at the
-- served bytes 2026-09-03 15:45Z: https://gamedesign.uk/contact.html still shows the OLD address.
--
-- ⚠ The address is baked into LLM-WRITTEN PROSE (hero-contact.subheadline, generic-text-block.content),
-- NOT a resolved field, so no re-render can fix it — 454's resolved-data fix does not reach a literal
-- inside a written sentence. The copy itself has to change.
--
-- Route: section_edit -> section-editor (apply_section_edit), edit_type=content_edit with
-- field_updates. That is a LIVE dispatching path (158 completed fleet-wide in the 3 days to
-- 2026-09-03); content_rewrite is NOT — all 182 recent rows are record-mode/deferred under
-- RFC_056's circuit breaker, so filing one there would have parked silently.
--
-- ⚠ This is a MECHANICAL substitution of an owner-specified literal, not authored copy: every
-- other character of the framework's own prose is preserved byte-for-byte (verified by asserting
-- exactly one occurrence per field, and by diffing length: NEW is 2 chars longer than OLD).
-- CLAUDE.md's "the framework writes the content, not you" is respected — asking the model to
-- rewrite the paragraph would change prose the owner did not ask to change and could reintroduce
-- a banned shape.
--
-- Apply: psql -f THIS FILE ONLY.
BEGIN;


-- about / generic-text-block (position 3), field 'content'
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, page_id, component_id, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'gamedesign_uk_rebuild lane', 'build', 'section_edit', 'medium',
  'Replace the superseded contact address on about/generic-text-block with gamedesignuk@contactforsales.com (owner ruling 2026-09-03)',
  ($j${"edit_type": "content_edit", "page_name": "about", "slot_name": "generic-text-block", "reason": "owner 2026-09-03: contact address changed to gamedesignuk@contactforsales.com; mechanical substitution, all other prose preserved"}$j$::jsonb || jsonb_build_object('field_updates', $f${"content": "<p>gamedesign.uk is a publication about the practice of game design, written for people who are already doing the work. The audience is design leads and principal designers, narrative designers, producers who run design teams, and the people who hire and manage them. Nobody reading it needs a game design document (GDD) explained, or a definition of the difference between systems design and level design, because that ground has already been covered elsewhere and covering it again here would waste the reader's time.</p><h3>What the site covers</h3><p>The <a href=\"/articles/index.html\">articles</a> look at how design decisions get made and unmade inside a studio: where a balance pass stalls, why a spec that reads clearly to design still needs a conversation with engineering before it means anything to a producer, what actually changes when a senior designer is promoted into a lead role. Each piece is written to advance an argument.</p><h3>Who it's for, and who it isn't</h3><p>This is written for people inside or alongside a studio. A piece that would sit equally comfortably on a general product management blog is treated as off-brief here, whatever its quality, because it doesn't require any game design knowledge to read.</p><h3>What it doesn't claim</h3><p>The site does not publish named case studies or testimonials, and where a pattern is described, it's described in general terms. Studio context varies too much for a single example to hold as a universal, and treating it that way would tell the reader less than it appears to.</p><p>Read the <a href=\"/index.html\">practice pages</a> for a wider view of the site, or get in touch through the <a href=\"/contact.html\">contact page</a> or at gamedesignuk@contactforsales.com.</p>"}$f$::jsonb)),
  'af3dd9bc-ca75-495c-838d-a9ce6882442e', '8d81e665-3ee0-443d-a873-690268c15fbb', 50, 'section-editor', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03', 'section_edit_contact_addr_about_generic-text-block')
ON CONFLICT DO NOTHING;

-- contact / generic-text-block (position 2), field 'content'
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, page_id, component_id, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'gamedesign_uk_rebuild lane', 'build', 'section_edit', 'medium',
  'Replace the superseded contact address on contact/generic-text-block with gamedesignuk@contactforsales.com (owner ruling 2026-09-03)',
  ($j${"edit_type": "content_edit", "page_name": "contact", "slot_name": "generic-text-block", "reason": "owner 2026-09-03: contact address changed to gamedesignuk@contactforsales.com; mechanical substitution, all other prose preserved"}$j$::jsonb || jsonb_build_object('field_updates', $f${"content": "<p>This site examines how game design actually happens inside studios: pipeline decisions, sign-off structures, the places where a spec breaks down before it reaches engineering. If something written here matches or contradicts your own experience, that is worth hearing about.</p><h3>Corrections and disagreements</h3><p>Generalisations about studio practice are qualified where possible, but no generalisation survives contact with every studio. If a piece describes a pattern that does not hold where you work, or gets a mechanism wrong, say so. That kind of correction does more for the next piece than any amount of editing done in isolation.</p><h3>Pitching or contributing</h3><p>Design leads, principal designers, and narrative designers with a position worth arguing can write in with what they have in mind. The bar is the same as for anything on the <a href=\"/articles/index.html\">articles</a> pages: the piece should go somewhere.</p><h3>Everything else</h3><p>Questions about the site itself, or the thinking behind it, can go to the same address. There is more on that thinking on the <a href=\"/about.html\">about</a> page.</p><p><strong>Email:</strong> gamedesignuk@contactforsales.com</p>"}$f$::jsonb)),
  '2e4e8700-c19e-42b2-af77-5dc28130fa93', '8d81e665-3ee0-443d-a873-690268c15fbb', 50, 'section-editor', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03', 'section_edit_contact_addr_contact_generic-text-block')
ON CONFLICT DO NOTHING;

-- contact / hero-contact (position 1), field 'subheadline'
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, page_id, component_id, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'gamedesign_uk_rebuild lane', 'build', 'section_edit', 'medium',
  'Replace the superseded contact address on contact/hero-contact with gamedesignuk@contactforsales.com (owner ruling 2026-09-03)',
  ($j${"edit_type": "content_edit", "page_name": "contact", "slot_name": "hero-contact", "reason": "owner 2026-09-03: contact address changed to gamedesignuk@contactforsales.com; mechanical substitution, all other prose preserved"}$j$::jsonb || jsonb_build_object('field_updates', $f${"subheadline": "For questions about an article, a proposal to write one, or anything else related to the practice of game design, write to gamedesignuk@contactforsales.com."}$f$::jsonb)),
  '2e4e8700-c19e-42b2-af77-5dc28130fa93', '231098ee-085d-4005-ae50-cc9863d3af9d', 50, 'section-editor', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03', 'section_edit_contact_addr_contact_hero-contact')
ON CONFLICT DO NOTHING;


DO $v$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND item_type='section_edit'
     AND item_key LIKE 'section_edit_contact_addr_%'
     AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled');
  IF n <> 3 THEN RAISE EXCEPTION '09-03e: expected 3 live section_edit rows, found %', n; END IF;
  RAISE NOTICE '09-03e OK: 3 section_edit items filed to section-editor';
END $v$;

COMMIT;

SELECT item_key, status, spec->>'page_name' AS page, spec->>'slot_name' AS slot,
       (spec->'field_updates') IS NOT NULL AS has_updates
  FROM site_work_items
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND item_key LIKE 'section_edit_contact_addr_%';