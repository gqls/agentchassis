# Baseline queries
1. Page inventory — what adoption decided each page is
   sqlSELECT name, page_type, status, build_status, created_at
   FROM pages
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   ORDER BY page_type, name;
   Key thing to check: any pages that should be game but got classified as content/tool/entity_page. This is the "missing vocabulary" signal we'll want to fix as part of Path A (games), but also worth knowing whether adoption noticed any game-like pages on the source.
2. Classification aspect content — what the classifier said the site contains
   sqlSELECT data
   FROM site_specs
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   AND aspect = 'classification'
   AND is_current = true;
   Look for: any mention of games as content type, any page_types enumeration, any signal that the classifier saw games even if it couldn't label them.
3. Strategy aspect — what the strategist proposed
   sqlSELECT data
   FROM site_specs
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   AND aspect = 'strategy'
   AND is_current = true;
   This tells us what the platform intends the site to be, beyond what was adopted directly. Useful contrast with what Path C will surface (parsed source interactivity).
4. Tool-list resolved vs game-list fabricated — side by side
   sqlSELECT p.name, pc.slot_name,
   pc.content_data->>'section_heading' AS heading,
   jsonb_array_length(COALESCE(pc.content_data->'items','[]'::jsonb)) AS resolved_items,
   pc.content_data->>'game1_title' AS fab_g1,
   pc.content_data->>'game2_title' AS fab_g2,
   pc.content_data->>'game3_title' AS fab_g3,
   pc.content_data->'items'->0->>'title' AS resolved_first_title,
   pc.content_data->'items'->0->>'url'   AS resolved_first_url
   FROM page_components pc
   JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   AND pc.slot_name IN ('tool-list','guide-list','game-list')
   ORDER BY pc.slot_name, p.name;
   The clearest before/after frame. After Path C + Path A, the fab_g* columns should be NULL (schema rewritten away) and resolved_items should be >0 for game-list too.
5. Components in use on this site — which of the 25 anti-pattern components are deployed
   sqlSELECT DISTINCT pc.slot_name, COUNT(*) AS used_count
   FROM page_components pc
   JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   GROUP BY pc.slot_name
   ORDER BY pc.slot_name;
   Tells us which other numbered-pattern components might be silently fabricating today. From the earlier audit: blog-listing, case-studies-grid, archetype-combinations, featured-inventory, product-details, provocation-feed all share the shape. If any are in use on this site, they're invisible fabrications waiting for Path E.
6. Corrected orchestration query — for future use
   sqlSELECT orchestration_id, owner_agent_type, current_step, status, updated_at
   FROM orchestration_states
   WHERE correlation_id = '1400b739-40b4-4627-80eb-9b76324c572a'
   ORDER BY created_at;
   Use orchestration_id not id everywhere we query this table.

Run those, paste the output, and I'll capture the baseline in a short note alongside the focus doc. Then we start on Path C.
Path C starts with me reading the two existing Go files closely:

extract_design_fingerprint_action.go (line ~62021, 250+ lines)
enrich_fingerprint_with_css_action.go (line ~38055, ~120 lines)

And the surrounding workflow that calls them (likely in the adoption orchestrator workflow JSON). Once I have those clear, I'll propose the concrete changes: new selectors, new fetch step, new aspect names, new LLM prompt. We'll work in step sizes that keep things reversible.