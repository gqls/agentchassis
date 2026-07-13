-- W3c step 0 (read-only): the five hero-* variant templates + idea.uk's component inventory.

-- 0.1 The variant templates (hazard class: declare dark --section-*, paint no matching
--     background; all 26 stored renders imageless). Paste back in full — the fix needles
--     derive from these, per-element, same discipline as footer/CTA/hero.
SELECT function, is_dark_section, created_from, length(html_template) AS len, html_template
FROM content_components
WHERE function IN ('hero-about','hero-case-studies','hero-contact','hero-services','hero-use-cases')
  AND is_active = true AND forked_from IS NULL
ORDER BY function;

-- 0.2 Which component functions does each idea.uk page actually use? This GATES W6:
--     any hazard-class or unconverted band-class function appearing here must be fixed
--     BEFORE the rebuild, or the rebuilt page carries the bug (e.g. an invisible
--     white-on-parchment hero-about).
SELECT p.name AS page, pc.slot_name, cc.function, cc.is_dark_section, cc.is_active
FROM pages p
JOIN page_components pc ON pc.page_id = p.id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
ORDER BY p.name, pc.slot_name;
