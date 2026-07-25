-- SQL_p5_chrome_forks.sql — webdesign.co.uk, phase 5
--
-- Three per-site chrome components (head, header, footer) plus explicit
-- site_components rows binding them to this site's slots.
--
-- WHY FORK RATHER THAN REUSE. render_site_components_action.go resolves a slot
-- by an explicit site_components.component_id if one exists, and otherwise
-- falls back to a global `WHERE function=$1 ORDER BY name LIMIT 1` lookup.
-- There are five active `site-header` rows fleet-wide; which one that lookup
-- returns is an accident of naming. Binding explicitly is the difference
-- between a header we chose and a header we got.
--
-- Forking is also the only real protection for chrome. bugs_open/069 is still
-- open — site_components has the same lock gap page_components had — so a
-- re-render can regenerate chrome at any time. It regenerating from OUR
-- template is fine; the danger is only ever it regenerating from someone else's.
--
-- SEARCH. The header carries the search pill and the carried 63-line engine
-- from website-design.com, shipped as this component's js_content. That
-- publishes to /tools/assets/site-header.js — the route fixed in bugs 018/041
-- (closed, live v1.0.1146, verified end-to-end on idea.uk). It only loads
-- because the template itself carries the <script src> tag; the assembler does
-- not inject one.
--
-- EVERY ANCHOR IS GATED. bugs_open/049 and 053 are both "chrome rendered a link
-- to something that isn't there": an empty legal group once filled a footer with
-- every in_footer page on the site. So each link is wrapped in an {{if}} and the
-- footer deliberately does not render a legal group at all.
--
-- Dollar-quoted throughout ($tmpl$ … $tmpl$) so the embedded HTML, CSS and JS
-- need no escaping. Doubling quotes across 400 lines of CSS is how a stray
-- character reaches production.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. HEAD
-- ---------------------------------------------------------------------------
-- assemblePage regex-replaces <title> from pages.title and injects the meta
-- description into the FIRST empty content="" in the head. Both placeholders
-- must therefore exist and be empty here.
INSERT INTO content_components (name, function, component_level, render_mode, html_template, input_schema, is_active, created_at, updated_at)
VALUES (
  'webdesign.co.uk Document Head',
  'webdesign-couk-head',
  'site',
  'template',
  $head$<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title></title>
<meta name="description" content="">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=Fira+Code:wght@400;600&display=swap">
<link rel="stylesheet" href="/assets/css/styles.css">
<link rel="stylesheet" href="/assets/css/port-compat.css">
<link rel="icon" href="/favicon.ico">$head$,
  '{"fields": {}}'::jsonb,
  true, NOW(), NOW()
);

-- ---------------------------------------------------------------------------
-- 2. HEADER
-- ---------------------------------------------------------------------------
INSERT INTO content_components (name, function, component_level, render_mode, html_template, input_schema, js_content, is_active, created_at, updated_at)
VALUES (
  'webdesign.co.uk Site Header',
  'webdesign-couk-header',
  'site',
  'template',
  $hdr$<style>
.wd-header {
  position: sticky; top: 0; z-index: 1000;
  background: #ffffff;
  border-bottom: 1px solid var(--border, #edece9);
  box-shadow: 0 2px 8px rgba(43,43,43,0.04);
}
.wd-header-inner {
  max-width: 1200px; margin: 0 auto; padding: 0.9rem 2rem;
  display: flex; align-items: center; justify-content: space-between; gap: 1.5rem;
}
.wd-wordmark {
  font-family: 'Inter', system-ui, sans-serif; font-weight: 800; font-size: 1.15rem;
  letter-spacing: -0.02em; color: var(--text, #2b2b2b); text-decoration: none; white-space: nowrap;
}
.wd-wordmark .tld {
  font-family: 'Fira Code', ui-monospace, monospace; font-weight: 600;
  color: var(--primary, #5c6b5d);
}
.wd-header-right { display: flex; align-items: center; gap: 1.25rem; }
.wd-nav { display: flex; align-items: center; gap: 0.35rem; list-style: none; margin: 0; padding: 0; }
.wd-nav a {
  display: block; padding: 0.45rem 0.8rem; border-radius: 8px;
  color: var(--text, #2b2b2b); text-decoration: none; font-size: 0.95rem; font-weight: 500;
  transition: background 0.15s ease, color 0.15s ease;
}
.wd-nav a:hover { background: var(--surface, #f3f1ec); color: var(--primary, #5c6b5d); }
.wd-search { position: relative; }
.wd-search input {
  width: 230px; padding: 0.5rem 1rem; border-radius: 99px;
  border: 1px solid var(--border, #edece9); background: var(--bg, #f9f8f6);
  font-family: 'Inter', system-ui, sans-serif; font-size: 0.9rem; color: var(--text, #2b2b2b);
  transition: border-color 0.15s ease, background 0.15s ease;
}
.wd-search input::placeholder { color: var(--text-dim, #717171); }
.wd-search input:focus {
  outline: none; border-color: var(--primary, #5c6b5d); background: #ffffff;
  box-shadow: 0 0 0 3px rgba(92,107,93,0.12);
}
.wd-results {
  display: none; position: absolute; top: calc(100% + 0.5rem); right: 0;
  width: 330px; max-height: 60vh; overflow-y: auto;
  background: #ffffff; border: 1px solid var(--border, #edece9); border-radius: 12px;
  box-shadow: 0 12px 32px rgba(43,43,43,0.08); z-index: 1100; padding: 0.35rem;
}
.wd-results a {
  display: block; padding: 0.6rem 0.75rem; border-radius: 8px;
  text-decoration: none; color: var(--text, #2b2b2b);
}
.wd-results a:hover { background: var(--surface, #f3f1ec); }
.wd-results .rc {
  display: block; font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 0.68rem; letter-spacing: 0.06em; text-transform: uppercase;
  color: var(--primary, #5c6b5d); margin-bottom: 0.1rem;
}
.wd-results .rt { font-weight: 600; font-size: 0.92rem; }
.wd-results .empty { padding: 0.9rem; color: var(--text-dim, #717171); font-size: 0.9rem; }
.wd-burger { display: none; background: none; border: 0; cursor: pointer; padding: 0.4rem; color: var(--text, #2b2b2b); font-size: 1.3rem; }
@media (max-width: 860px) {
  .wd-header-inner { flex-wrap: wrap; gap: 0.75rem; }
  .wd-burger { display: block; }
  .wd-nav { display: none; width: 100%; flex-direction: column; align-items: stretch; }
  .wd-nav.open { display: flex; }
  .wd-search input { width: 100%; }
  .wd-search { flex: 1 1 100%; order: 3; }
}
</style>
<header class="wd-header">
  <div class="wd-header-inner">
    <a class="wd-wordmark" href="/index.html">webdesign<span class="tld">.co.uk</span></a>
    <button class="wd-burger" id="wdBurger" aria-label="Menu" aria-expanded="false">&#9776;</button>
    <div class="wd-header-right">
      <nav><ul class="wd-nav" id="wdNav">
        {{range .categories}}{{if .url}}<li><a href="{{.url}}">{{.name}}</a></li>{{end}}{{end}}
      </ul></nav>
      <div class="wd-search">
        <input type="search" id="globalSearch" placeholder="Search tools and articles" autocomplete="off" aria-label="Search tools and articles">
        <div class="wd-results" id="searchResults"></div>
      </div>
    </div>
  </div>
</header>
<script src="/tools/assets/webdesign-couk-header.js"></script>$hdr$,
  $schema${
    "fields": {
      "categories": {
        "type": "array",
        "source": "nav",
        "required": false,
        "llm_guidance": "Primary navigation. Each item needs name and url. Exactly three: Tools, Learn, About."
      }
    }
  }$schema$::jsonb,
  $js$/* Global search + mobile nav for webdesign.co.uk.
   Carried from website-design.com/assets/js/search.js, with three changes:
   it no longer throws when its elements are absent (chrome renders on every
   page, including any that lack the widget), results are escaped rather than
   interpolated raw, and the keyword match is token-based instead of a substring
   test on the joined string — the old one matched "ss" inside "css". */
(function () {
  'use strict';

  var input = document.getElementById('globalSearch');
  var results = document.getElementById('searchResults');
  var burger = document.getElementById('wdBurger');
  var nav = document.getElementById('wdNav');

  if (burger && nav) {
    burger.addEventListener('click', function () {
      var open = nav.classList.toggle('open');
      burger.setAttribute('aria-expanded', open ? 'true' : 'false');
    });
  }

  if (!input || !results) return;

  var index = [];
  fetch('/search.json')
    .then(function (r) { return r.json(); })
    .then(function (data) { index = Array.isArray(data) ? data : []; })
    .catch(function () { /* search degrades to nothing; the site still works */ });

  function esc(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }

  function matches(item, q) {
    if (String(item.title || '').toLowerCase().indexOf(q) !== -1) return true;
    if (String(item.category || '').toLowerCase().indexOf(q) !== -1) return true;
    var kw = String(item.keywords || '').toLowerCase().split(/\s+/);
    for (var i = 0; i < kw.length; i++) {
      if (kw[i].indexOf(q) === 0) return true;
    }
    return false;
  }

  function render(list) {
    if (!list.length) {
      results.innerHTML = '<div class="empty">No results found.</div>';
      results.style.display = 'block';
      return;
    }
    var html = '';
    for (var i = 0; i < list.length && i < 25; i++) {
      var it = list[i];
      html += '<a href="' + esc(it.url) + '">' +
              '<span class="rc">' + esc(it.category) + '</span>' +
              '<span class="rt">' + esc(it.title) + '</span></a>';
    }
    results.innerHTML = html;
    results.style.display = 'block';
  }

  input.addEventListener('input', function (e) {
    var q = String(e.target.value || '').trim().toLowerCase();
    if (!q) { results.style.display = 'none'; return; }
    render(index.filter(function (item) { return matches(item, q); }));
  });

  input.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') { results.style.display = 'none'; input.blur(); }
  });

  document.addEventListener('click', function (e) {
    if (!input.contains(e.target) && !results.contains(e.target)) {
      results.style.display = 'none';
    }
  });
})();
$js$,
  true, NOW(), NOW()
);

-- ---------------------------------------------------------------------------
-- 3. FOOTER
-- ---------------------------------------------------------------------------
-- No newsletter, no legal group, no invented links. The site collects nothing,
-- so a footer offering to take an email address would be a lie in furniture.
INSERT INTO content_components (name, function, component_level, render_mode, html_template, input_schema, is_active, created_at, updated_at)
VALUES (
  'webdesign.co.uk Site Footer',
  'webdesign-couk-footer',
  'site',
  'template',
  $ftr$<style>
.wd-footer {
  border-top: 1px solid var(--border, #edece9);
  margin-top: 4rem; padding: 3rem 2rem;
  text-align: center; color: var(--text-dim, #717171);
  font-family: 'Inter', system-ui, sans-serif; font-size: 0.875rem;
}
.wd-footer-mark {
  font-family: 'Fira Code', ui-monospace, monospace; font-weight: 600;
  color: var(--text, #2b2b2b); font-size: 0.95rem; display: block; margin-bottom: 0.6rem;
}
.wd-footer-links { display: flex; gap: 1.25rem; justify-content: center; flex-wrap: wrap; margin: 0 0 0.9rem; padding: 0; list-style: none; }
.wd-footer-links a { color: var(--text-dim, #717171); text-decoration: none; }
.wd-footer-links a:hover { color: var(--primary, #5c6b5d); }
.wd-footer-note { margin: 0; }
</style>
<footer class="wd-footer">
  <span class="wd-footer-mark">webdesign.co.uk</span>
  <ul class="wd-footer-links">
    {{range .categories}}{{if .url}}<li><a href="{{.url}}">{{.name}}</a></li>{{end}}{{end}}
  </ul>
  <p class="wd-footer-note">Everything here runs in your browser. Nothing is uploaded, nothing is stored, no account is needed.</p>
  {{if .copyright}}<p class="wd-footer-note">{{.copyright}}</p>{{end}}
</footer>$ftr$,
  $schema${
    "fields": {
      "categories": { "type": "array", "source": "nav", "required": false },
      "copyright":  { "type": "string", "source": "site", "required": false }
    }
  }$schema$::jsonb,
  true, NOW(), NOW()
);

-- ---------------------------------------------------------------------------
-- 4. Bind them to this site's slots, explicitly.
-- ---------------------------------------------------------------------------
INSERT INTO site_components (site_id, slot_name, component_id, build_status, created_at, updated_at)
SELECT s.id, v.slot, c.id, 'pending', NOW(), NOW()
FROM sites s
CROSS JOIN (VALUES
    ('head',   'webdesign-couk-head'),
    ('header', 'webdesign-couk-header'),
    ('footer', 'webdesign-couk-footer')
) AS v(slot, fn)
JOIN content_components c ON c.function = v.fn AND c.is_active
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (
      SELECT 1 FROM site_components sc
       WHERE sc.site_id = s.id AND sc.slot_name = v.slot
  );

-- Any slot that already had a row (from an earlier render) is repointed.
UPDATE site_components sc
   SET component_id = c.id, updated_at = NOW()
  FROM sites s, content_components c
 WHERE sc.site_id = s.id
   AND s.domain = 'webdesign.co.uk'
   AND c.is_active
   AND c.function = 'webdesign-couk-' || sc.slot_name
   AND sc.component_id IS DISTINCT FROM c.id;

-- ---------------------------------------------------------------------------
-- 5. Brand fields the renderer reads.
-- ---------------------------------------------------------------------------
UPDATE sites
   SET company_name = 'webdesign.co.uk',
       updated_at = NOW()
 WHERE domain = 'webdesign.co.uk';

DO $verify$
DECLARE v_site uuid; n int; head_tmpl text;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT count(*) INTO n FROM site_components
     WHERE site_id = v_site AND component_id IS NOT NULL
       AND slot_name IN ('head','header','footer');
    IF n <> 3 THEN
        RAISE EXCEPTION 'expected 3 bound chrome slots, found %', n;
    END IF;

    -- The two placeholders assemblePage rewrites must exist and be EMPTY.
    SELECT c.html_template INTO head_tmpl
      FROM site_components sc JOIN content_components c ON c.id = sc.component_id
     WHERE sc.site_id = v_site AND sc.slot_name = 'head';
    IF head_tmpl NOT LIKE '%<title></title>%' THEN
        RAISE EXCEPTION 'head has no empty <title> for assemblePage to fill';
    END IF;
    IF head_tmpl NOT LIKE '%content=""%' THEN
        RAISE EXCEPTION 'head has no empty content="" for the meta description';
    END IF;

    -- The header must load its own JS: the assembler injects no script tag.
    IF NOT EXISTS (
        SELECT 1 FROM site_components sc JOIN content_components c ON c.id = sc.component_id
         WHERE sc.site_id = v_site AND sc.slot_name = 'header'
           AND c.html_template LIKE '%/tools/assets/webdesign-couk-header.js%'
           AND COALESCE(c.js_content,'') <> ''
    ) THEN
        RAISE EXCEPTION 'header must carry BOTH js_content and its own <script src> tag';
    END IF;

    RAISE NOTICE 'chrome forks bound: head, header (with search), footer';
END
$verify$;

COMMIT;
