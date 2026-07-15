-- ============================================================
-- FIX: the hidden clone-template row renders as an empty row with a lone dot.
-- Cause: the component CSS sets `.provocations-archive__item { display: grid; }`
-- (an AUTHOR rule), which overrides the `hidden` attribute's UA-stylesheet
-- `display: none`. Fix: add a more specific author rule hiding the template.
-- Anchored REPLACE on the unique base-rule opener (marker-anchoring lesson);
-- guarded for idempotency. Apply to the TEMPLATE (future renders) and the
-- CURRENT page instance, then rerender the page (085 script) to redeploy.
-- ============================================================

-- template (component 70d6662a)
UPDATE content_components
SET html_template = REPLACE(html_template,
      '.provocations-archive-list-section .provocations-archive__item {',
      '.provocations-archive-list-section .provocations-archive__item[data-archive-template] { display: none; }

  .provocations-archive-list-section .provocations-archive__item {'),
    updated_at = NOW()
WHERE id = '70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37'
  AND html_template LIKE '%.provocations-archive-list-section .provocations-archive__item {%'
  AND html_template NOT LIKE '%[data-archive-template] { display: none; }%'
RETURNING (html_template LIKE '%[data-archive-template] { display: none; }%') AS template_fixed;

-- current page instance (page e4b3b195, provocations-index)
UPDATE page_components
SET rendered_html = REPLACE(rendered_html,
      '.provocations-archive-list-section .provocations-archive__item {',
      '.provocations-archive-list-section .provocations-archive__item[data-archive-template] { display: none; }

  .provocations-archive-list-section .provocations-archive__item {'),
    updated_at = NOW()
WHERE component_id = '70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37'
  AND page_id = 'e4b3b195-919f-45ad-854e-201d3e846ea8'
  AND rendered_html LIKE '%.provocations-archive-list-section .provocations-archive__item {%'
  AND rendered_html NOT LIKE '%[data-archive-template] { display: none; }%'
RETURNING (rendered_html LIKE '%[data-archive-template] { display: none; }%') AS instance_fixed;

-- Expect one row from each, both flags = t. A 0-row result means already-fixed
-- or anchor mismatch — check the stored text before concluding.
