-- W2b step 0 (read-only): reuse-before-create — does an updated_at trigger function
-- already exist, and do any tables already auto-update?
SELECT p.proname
FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = 'public' AND p.proname ~* '(updated_at|touch|set_updated)';

SELECT c.relname AS table_name, t.tgname AS trigger_name, p.proname AS function_name
FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_proc  p ON p.oid = t.tgfoid
WHERE NOT t.tgisinternal
  AND p.proname ~* '(updated_at|touch|set_updated)'
ORDER BY c.relname;

-- Side-reads for W3 prep:
-- (a) idea.uk's sites.status (it was missing from the 0.4 active-sites count — why?)
SELECT domain, status FROM sites WHERE domain = 'idea.uk';
-- (b) per-layout site counts WITHOUT the status filter:
SELECT l.name AS layout, l.scheme, s.status, count(*) AS sites
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct ON ct.id = sc.css_theme_id
JOIN layouts l ON l.id = ct.layout_id
GROUP BY l.name, l.scheme, s.status
ORDER BY l.name, s.status;
