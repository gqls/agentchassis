-- Reads to size Step D (dartsonline fixer run) AND the three page builds. Schema-first.

-- ── A. dartsonline: what the fresh build produced (the fixer's specimen) ──
\d pages
SELECT name, page_type, build_status, status
FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' ORDER BY name;

-- Its components carrying literal --section-* or hardcoded text colour (what the
-- re-aimed fixer will act on — the specimen's "before"):
SELECT p.name, pc.slot_name,
       (pc.rendered_html LIKE '%--section-text:%#%' OR pc.rendered_html LIKE '%--section-heading:%#%') AS literal_section_var,
       (pc.rendered_html ~ 'color:\s*#[0-9a-fA-F]{3,6}') AS literal_text_hex
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
ORDER BY p.name, pc.slot_name;

-- ── B. idea.uk planned pages: do stub rows exist, or must they be created? ──
-- (Decides whether "build" = emit needs_page for existing rows, or create rows first.)
SELECT name, page_type, build_status, status
FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='idea.uk')
ORDER BY name;

-- The nav/plan source that references news/guides/audience-check (why they 404):
SELECT aspect, source_agent, is_current,
       left(data::text, 200) AS data_head
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='idea.uk' AND aspect IN ('navigation','site_plan','composition')
ORDER BY ss.created_at DESC LIMIT 6;

-- ── C. the audience-check TOOL: how are existing tool pages shaped? ──
-- (idea.uk already deployed /tools/audience-check per earlier greps — is there a
--  page row + a tool component to copy the shape from?)
SELECT p.name, p.page_type, pc.slot_name, cc.function
FROM pages p
LEFT JOIN page_components pc ON pc.page_id=p.id
LEFT JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id=(SELECT id FROM sites WHERE domain='idea.uk')
  AND (p.name LIKE '%audience%' OR p.name LIKE '%tool%' OR p.page_type='tool')
ORDER BY p.name, pc.slot_name;
