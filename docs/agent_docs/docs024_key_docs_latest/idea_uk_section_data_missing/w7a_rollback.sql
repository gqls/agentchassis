-- W7a rollback: remove the gate.
UPDATE content_components
SET html_template =
 replace(replace(html_template,
  '{{if .illustration_url}}<div class="brief-explanation__image-wrapper">',
  '<div class="brief-explanation__image-wrapper">'),
  E'{{.badge_label}}\n        </span>\n      </div>{{end}}',
  E'{{.badge_label}}\n        </span>\n      </div>'),
 updated_at = now()
WHERE function = 'brief-explanation' AND is_active = true
  AND html_template LIKE '%{{if .illustration_url}}%'
RETURNING function;
