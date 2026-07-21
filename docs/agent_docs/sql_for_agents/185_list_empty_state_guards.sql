-- 185_list_empty_state_guards.sql — 2026-07-21, cta_link_integrity (bugs_open/054)
--
-- Five active library components render a query-sourced list with
-- {{range .items}} and NO empty-state. On a site whose entities/games/guides/
-- tools have not populated yet, the query resolves to an empty slice and the
-- section renders a blank container with no explanatory copy.
--
-- Why the empty render is reachable (this is the load-bearing fact):
--   plan_sections_action.go:1288-1321 — for a source:query.* field the resolver
--   sets resolvedData[field]=value and `continue`s whenever value != nil. An
--   empty slice is NOT nil in Go, so the required/on_missing/min_items branch at
--   :1333-1432 NEVER runs for a query array. So items {required:true,min_items:1}
--   — which all five carry — is SILENTLY IGNORED, and an empty list is handed
--   straight to the template. (The comment at :1285 claims on_missing applies on
--   empty; the code does not implement it. That deeper gap is bugs_open/054
--   fix-candidate-2 and is deliberately OUT of scope here — we fix the render,
--   not the resolver.)
--
-- This mirrors the two news components (latest-news, news-listing) which already
-- carry the guard — and, like news-listing's loading_text, the empty-state copy
-- is a source:llm field (empty_state_text) so it is authored in the site's own
-- language, with a hardcoded English fallback only for when the LLM field is
-- absent (bugs_open/026 — no hardcoded English on a non-English site).
--
-- Purely additive: when items IS present nothing changes; the {{else}} arm is
-- only reached on an empty list. Config change, LIVE IMMEDIATELY, no image roll;
-- takes effect per page at that page's next render.
--
-- Targeted by (function, is_active) not name: two of these are *_pre_037 rows
-- (game-list, guide-list) that are the SOLE active row for their function
-- (RUNBOOK R10) — function is the stable identity and each resolves to exactly
-- one active row. NOT deleted, NOT renamed; html_template edited in place.
--
-- ROLLBACK: bak_054_list_components_20260721 holds the five full rows.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE bak_054_list_components_20260721 AS
SELECT * FROM content_components
WHERE is_active AND function IN ('archetype-grid','game-list','guide-list','tool-cta','tool-list');

DO $mig$
DECLARE
  changed int;
  n int;
BEGIN
  -- (1) archetype-grid -----------------------------------------------------------------
  IF (SELECT count(*) FROM content_components WHERE is_active AND function='archetype-grid') <> 1 THEN
    RAISE EXCEPTION 'archetype-grid: expected exactly 1 active row, schema drifted — re-measure';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM content_components
                 WHERE is_active AND function='archetype-grid'
                   AND position($o1${{range .items}}
      <article class="ag-card" role="listitem">
        <div class="ag-card-badge">{{.nav_label}}</div>
        <div class="ag-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="9"/><path d="M12 8v4l3 3"/></svg>
        </div>
        <h3 class="ag-card-title">{{.title}}</h3>
        <p class="ag-card-desc">{{.meta_description}}</p>
        <a class="ag-card-link" href="{{.url}}">{{.title}} <span aria-hidden="true">&rarr;</span></a>
      </article>
      {{end}}$o1$ in html_template) > 0
                   AND html_template !~ '\{\{ *if [^}]*items *\}\}') THEN
    RAISE EXCEPTION 'archetype-grid: range block not in expected UNGATED form — re-derive the needle';
  END IF;
  UPDATE content_components
     SET input_schema = jsonb_set(input_schema, '{fields,empty_state_text}', '{"type":"text","source":"llm","required":false,"llm_guidance":"Short friendly one-line message shown when this section currently has no items, in the language of the site (Spanish example: Pronto habra mas). Say more are on the way; under 12 words. bugs_open/054."}'::jsonb, true),
         html_template = replace(html_template, $o1${{range .items}}
      <article class="ag-card" role="listitem">
        <div class="ag-card-badge">{{.nav_label}}</div>
        <div class="ag-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="9"/><path d="M12 8v4l3 3"/></svg>
        </div>
        <h3 class="ag-card-title">{{.title}}</h3>
        <p class="ag-card-desc">{{.meta_description}}</p>
        <a class="ag-card-link" href="{{.url}}">{{.title}} <span aria-hidden="true">&rarr;</span></a>
      </article>
      {{end}}$o1$, $n1${{if .items}}{{range .items}}
      <article class="ag-card" role="listitem">
        <div class="ag-card-badge">{{.nav_label}}</div>
        <div class="ag-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="9"/><path d="M12 8v4l3 3"/></svg>
        </div>
        <h3 class="ag-card-title">{{.title}}</h3>
        <p class="ag-card-desc">{{.meta_description}}</p>
        <a class="ag-card-link" href="{{.url}}">{{.title}} <span aria-hidden="true">&rarr;</span></a>
      </article>
      {{end}}{{else}}<p class="ag-empty">{{if .empty_state_text}}{{.empty_state_text}}{{else}}More coming soon.{{end}}</p>{{end}}$n1$),
         updated_at = now()
   WHERE is_active AND function='archetype-grid';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'archetype-grid: update touched % rows', changed; END IF;

  -- (2) game-list -----------------------------------------------------------------
  IF (SELECT count(*) FROM content_components WHERE is_active AND function='game-list') <> 1 THEN
    RAISE EXCEPTION 'game-list: expected exactly 1 active row, schema drifted — re-measure';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM content_components
                 WHERE is_active AND function='game-list'
                   AND position($o2${{range .items}}
      <article class="gl-card">
        <div class="gl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M8 5v14l11-7z"/></svg>
        </div>
        <h3 class="gl-card-title">{{.title}}</h3>
        <p class="gl-card-desc">{{.meta_description}}</p>
        <a class="gl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}$o2$ in html_template) > 0
                   AND html_template !~ '\{\{ *if [^}]*items *\}\}') THEN
    RAISE EXCEPTION 'game-list: range block not in expected UNGATED form — re-derive the needle';
  END IF;
  UPDATE content_components
     SET input_schema = jsonb_set(input_schema, '{fields,empty_state_text}', '{"type":"text","source":"llm","required":false,"llm_guidance":"Short friendly one-line message shown when this section currently has no items, in the language of the site (Spanish example: Pronto habra mas). Say more are on the way; under 12 words. bugs_open/054."}'::jsonb, true),
         html_template = replace(html_template, $o2${{range .items}}
      <article class="gl-card">
        <div class="gl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M8 5v14l11-7z"/></svg>
        </div>
        <h3 class="gl-card-title">{{.title}}</h3>
        <p class="gl-card-desc">{{.meta_description}}</p>
        <a class="gl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}$o2$, $n2${{if .items}}{{range .items}}
      <article class="gl-card">
        <div class="gl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M8 5v14l11-7z"/></svg>
        </div>
        <h3 class="gl-card-title">{{.title}}</h3>
        <p class="gl-card-desc">{{.meta_description}}</p>
        <a class="gl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}{{else}}<p class="gl-empty">{{if .empty_state_text}}{{.empty_state_text}}{{else}}More games coming soon.{{end}}</p>{{end}}$n2$),
         updated_at = now()
   WHERE is_active AND function='game-list';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'game-list: update touched % rows', changed; END IF;

  -- (3) guide-list -----------------------------------------------------------------
  IF (SELECT count(*) FROM content_components WHERE is_active AND function='guide-list') <> 1 THEN
    RAISE EXCEPTION 'guide-list: expected exactly 1 active row, schema drifted — re-measure';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM content_components
                 WHERE is_active AND function='guide-list'
                   AND position($o3${{range .items}}
      <a class="guide-card" href="{{.url}}" aria-label="{{.title}}">
        <span class="guide-card-badge">{{.nav_label}}</span>
        <h3 class="guide-card-title">{{.title}}</h3>
        <p class="guide-card-desc">{{.meta_description}}</p>
        <span class="guide-card-link-label">Read guide</span>
      </a>
      {{end}}$o3$ in html_template) > 0
                   AND html_template !~ '\{\{ *if [^}]*items *\}\}') THEN
    RAISE EXCEPTION 'guide-list: range block not in expected UNGATED form — re-derive the needle';
  END IF;
  UPDATE content_components
     SET input_schema = jsonb_set(input_schema, '{fields,empty_state_text}', '{"type":"text","source":"llm","required":false,"llm_guidance":"Short friendly one-line message shown when this section currently has no items, in the language of the site (Spanish example: Pronto habra mas). Say more are on the way; under 12 words. bugs_open/054."}'::jsonb, true),
         html_template = replace(html_template, $o3${{range .items}}
      <a class="guide-card" href="{{.url}}" aria-label="{{.title}}">
        <span class="guide-card-badge">{{.nav_label}}</span>
        <h3 class="guide-card-title">{{.title}}</h3>
        <p class="guide-card-desc">{{.meta_description}}</p>
        <span class="guide-card-link-label">Read guide</span>
      </a>
      {{end}}$o3$, $n3${{if .items}}{{range .items}}
      <a class="guide-card" href="{{.url}}" aria-label="{{.title}}">
        <span class="guide-card-badge">{{.nav_label}}</span>
        <h3 class="guide-card-title">{{.title}}</h3>
        <p class="guide-card-desc">{{.meta_description}}</p>
        <span class="guide-card-link-label">Read guide</span>
      </a>
      {{end}}{{else}}<p class="guide-list-empty">{{if .empty_state_text}}{{.empty_state_text}}{{else}}More guides coming soon.{{end}}</p>{{end}}$n3$),
         updated_at = now()
   WHERE is_active AND function='guide-list';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'guide-list: update touched % rows', changed; END IF;

  -- (4) tool-cta -----------------------------------------------------------------
  IF (SELECT count(*) FROM content_components WHERE is_active AND function='tool-cta') <> 1 THEN
    RAISE EXCEPTION 'tool-cta: expected exactly 1 active row, schema drifted — re-measure';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM content_components
                 WHERE is_active AND function='tool-cta'
                   AND position($o4${{range .items}}
        <a href="{{.url}}" class="tool-cta-card">
          <h3 class="tool-cta-card-title">{{.title}}</h3>
          <p class="tool-cta-card-desc">{{.meta_description}}</p>
          <span class="tool-cta-card-link">{{.nav_label}}</span>
        </a>
        {{end}}$o4$ in html_template) > 0
                   AND html_template !~ '\{\{ *if [^}]*items *\}\}') THEN
    RAISE EXCEPTION 'tool-cta: range block not in expected UNGATED form — re-derive the needle';
  END IF;
  UPDATE content_components
     SET input_schema = jsonb_set(input_schema, '{fields,empty_state_text}', '{"type":"text","source":"llm","required":false,"llm_guidance":"Short friendly one-line message shown when this section currently has no items, in the language of the site (Spanish example: Pronto habra mas). Say more are on the way; under 12 words. bugs_open/054."}'::jsonb, true),
         html_template = replace(html_template, $o4${{range .items}}
        <a href="{{.url}}" class="tool-cta-card">
          <h3 class="tool-cta-card-title">{{.title}}</h3>
          <p class="tool-cta-card-desc">{{.meta_description}}</p>
          <span class="tool-cta-card-link">{{.nav_label}}</span>
        </a>
        {{end}}$o4$, $n4${{if .items}}{{range .items}}
        <a href="{{.url}}" class="tool-cta-card">
          <h3 class="tool-cta-card-title">{{.title}}</h3>
          <p class="tool-cta-card-desc">{{.meta_description}}</p>
          <span class="tool-cta-card-link">{{.nav_label}}</span>
        </a>
        {{end}}{{else}}<p class="tool-cta-empty">{{if .empty_state_text}}{{.empty_state_text}}{{else}}More tools coming soon.{{end}}</p>{{end}}$n4$),
         updated_at = now()
   WHERE is_active AND function='tool-cta';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'tool-cta: update touched % rows', changed; END IF;

  -- (5) tool-list -----------------------------------------------------------------
  IF (SELECT count(*) FROM content_components WHERE is_active AND function='tool-list') <> 1 THEN
    RAISE EXCEPTION 'tool-list: expected exactly 1 active row, schema drifted — re-measure';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM content_components
                 WHERE is_active AND function='tool-list'
                   AND position($o5${{range .items}}
      <article class="tl-card">
        {{if .image}}<div class="tl-card-media"><img src="{{.image}}" alt="{{.title}}" loading="lazy"></div>{{end}}
        <div class="tl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
        </div>
        <h3 class="tl-card-title">{{.title}}</h3>
        <p class="tl-card-desc">{{.meta_description}}</p>
        <a class="tl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}$o5$ in html_template) > 0
                   AND html_template !~ '\{\{ *if [^}]*items *\}\}') THEN
    RAISE EXCEPTION 'tool-list: range block not in expected UNGATED form — re-derive the needle';
  END IF;
  UPDATE content_components
     SET input_schema = jsonb_set(input_schema, '{fields,empty_state_text}', '{"type":"text","source":"llm","required":false,"llm_guidance":"Short friendly one-line message shown when this section currently has no items, in the language of the site (Spanish example: Pronto habra mas). Say more are on the way; under 12 words. bugs_open/054."}'::jsonb, true),
         html_template = replace(html_template, $o5${{range .items}}
      <article class="tl-card">
        {{if .image}}<div class="tl-card-media"><img src="{{.image}}" alt="{{.title}}" loading="lazy"></div>{{end}}
        <div class="tl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
        </div>
        <h3 class="tl-card-title">{{.title}}</h3>
        <p class="tl-card-desc">{{.meta_description}}</p>
        <a class="tl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}$o5$, $n5${{if .items}}{{range .items}}
      <article class="tl-card">
        {{if .image}}<div class="tl-card-media"><img src="{{.image}}" alt="{{.title}}" loading="lazy"></div>{{end}}
        <div class="tl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
        </div>
        <h3 class="tl-card-title">{{.title}}</h3>
        <p class="tl-card-desc">{{.meta_description}}</p>
        <a class="tl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}{{else}}<p class="tl-empty">{{if .empty_state_text}}{{.empty_state_text}}{{else}}More tools coming soon.{{end}}</p>{{end}}$n5$),
         updated_at = now()
   WHERE is_active AND function='tool-list';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'tool-list: update touched % rows', changed; END IF;

  -- post-conditions: all five now carry an items guard, the empty-state class,
  -- and the empty_state_text schema field.
  SELECT count(*) INTO n FROM content_components
   WHERE is_active AND function IN ('archetype-grid','game-list','guide-list','tool-cta','tool-list')
     AND html_template ~ '\{\{ *if [^}]*items *\}\}'
     AND (input_schema->'fields') ? 'empty_state_text';
  IF n <> 5 THEN RAISE EXCEPTION 'post-condition failed: % of 5 components guarded+fielded', n; END IF;

  -- every one still ranges over items exactly once (guard did not disturb the loop)
  SELECT count(*) INTO n FROM content_components
   WHERE is_active AND function IN ('archetype-grid','game-list','guide-list','tool-cta','tool-list')
     AND html_template LIKE '%{{range .items}}%';
  IF n <> 5 THEN RAISE EXCEPTION 'post-condition failed: % of 5 still range over items', n; END IF;
END $mig$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('185_list_empty_state_guards.sql',
        'bugs_open/054: wrap the 5 unguarded {{range .items}} list components (archetype-grid, game-list, guide-list, tool-cta, tool-list) in {{if .items}}...{{else}}<empty-state>{{end}}, matching the news pair. Empty-state copy is a new source:llm empty_state_text field (translatable) with an English template fallback. Config change, live immediately; purely additive (only the empty-list path changes). Root-cause resolver gap (min_items ignored for query arrays, plan_sections_action.go:1288-1321) left to 054 fix-candidate-2.');

COMMIT;

-- Post-apply, for the record:
--   SELECT function,
--          html_template ~ '\{\{ *if [^}]*items *\}\}' AS guarded,
--          (input_schema->'fields') ? 'empty_state_text'  AS has_field
--     FROM content_components
--    WHERE is_active AND function IN ('archetype-grid','game-list','guide-list','tool-cta','tool-list')
--    ORDER BY function;   -- expect guarded=t, has_field=t for all five
