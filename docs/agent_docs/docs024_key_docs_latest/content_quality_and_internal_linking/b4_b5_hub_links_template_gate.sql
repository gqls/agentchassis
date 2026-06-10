-- B4/B5 — gate the "Browse All X" button on a resolved cta_url.
--
-- The list templates render the anchor ungated (href="{{.cta_url}}"), so an
-- unresolved cta_url produced href="". With cta_url now sourced from
-- query.section_index_for (b4_b5_hub_links_schema.sql), gating gives the
-- correct-or-absent guarantee: a real hub renders the button, no hub renders
-- nothing. For gamesdesign all three hubs exist so the buttons render; the gate
-- is the safety net for sites lacking a hub.
--
-- Snapshot: covered by content_components_bak_hubfix_0610 from
-- b4_b5_hub_links_schema.sql (full-row snapshot taken before any edit). Run this
-- after that file. Each replace must report UPDATE 1.

-- tool-list
UPDATE content_components
SET html_template = replace(html_template,
      '<a class="tl-cta-btn" href="{{.cta_url}}">{{.cta_label}}</a>',
      '{{if .cta_url}}<a class="tl-cta-btn" href="{{.cta_url}}">{{.cta_label}}</a>{{end}}'),
    updated_at = now()
WHERE name = 'tool-list';

-- game-list_pre_037
UPDATE content_components
SET html_template = replace(html_template,
      '<a class="gl-cta-btn" href="{{.cta_url}}">{{.cta_label}}</a>',
      '{{if .cta_url}}<a class="gl-cta-btn" href="{{.cta_url}}">{{.cta_label}}</a>{{end}}'),
    updated_at = now()
WHERE name = 'game-list_pre_037';

-- guide-list_pre_037
UPDATE content_components
SET html_template = replace(html_template,
      '<a class="guide-list-cta-btn" href="{{.cta_url}}">{{.cta_button_label}}</a>',
      '{{if .cta_url}}<a class="guide-list-cta-btn" href="{{.cta_url}}">{{.cta_button_label}}</a>{{end}}'),
    updated_at = now()
WHERE name = 'guide-list_pre_037';

-- Verify — each template now gates the Browse-All anchor.
SELECT name,
       (html_template LIKE '%{{if .cta_url}}<a%') AS browse_all_gated
FROM content_components
WHERE name IN ('tool-list', 'game-list_pre_037', 'guide-list_pre_037')
ORDER BY name;
