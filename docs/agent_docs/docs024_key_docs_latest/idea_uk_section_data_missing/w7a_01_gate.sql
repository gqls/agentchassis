-- W7a: gate brief-explanation's image wrapper. Needed regardless of the defer question:
-- whenever the section renders with the field skipped, the ungated <img> would emit
-- src="" (a broken image). Go templates treat "" and missing as false, so the gate
-- covers both. Inert for the DROP itself (deferral happens before render) — the section
-- returns when the Go batch deploys correct plan_sections behaviour.
-- Gate check first, then the two replaces, then verify booleans.
SELECT function,
 (html_template LIKE '%<div class="brief-explanation__image-wrapper">%')            AS n1_open,
 position(E'{{.badge_label}}\n        </span>\n      </div>' in html_template) > 0  AS n2_close,
 (html_template LIKE '%{{if .illustration_url}}%')                                   AS already_gated  -- expect f
FROM content_components
WHERE function = 'brief-explanation' AND is_active = true AND forked_from IS NULL;

UPDATE content_components
SET html_template =
 replace(replace(html_template,
  '<div class="brief-explanation__image-wrapper">',
  '{{if .illustration_url}}<div class="brief-explanation__image-wrapper">'),
  E'{{.badge_label}}\n        </span>\n      </div>',
  E'{{.badge_label}}\n        </span>\n      </div>{{end}}'),
 updated_at = now()
WHERE function = 'brief-explanation' AND is_active = true AND forked_from IS NULL
  AND html_template NOT LIKE '%{{if .illustration_url}}%'
RETURNING function,
          (html_template LIKE '%{{if .illustration_url}}<div class="brief-explanation__image-wrapper">%') AS gated_open,   -- expect t
          (html_template LIKE E'%</div>{{end}}%')                                                          AS gated_close;  -- expect t
-- Expect: gate row t/t/f, then UPDATE 1 with t/t.
