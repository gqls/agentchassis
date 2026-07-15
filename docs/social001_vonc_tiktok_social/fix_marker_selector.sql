-- ============================================================
-- FIX: data-runtime-fill marker was injected into the inline scripts'
-- attribute selectors by the over-broad REPLACE, producing a malformed
-- selector  [data-component="X" data-runtime-fill="true"]  → SyntaxError →
-- the cosmetic hover/entrance IIFE dies (loaders in snippets.js unaffected).
-- Revert ONLY the in-selector copy (the one immediately followed by ']');
-- the <section> tag copy (followed by a space + more attributes) is KEPT so
-- the assembler still keeps the section. Applies to lobby-grid (this turn) and
-- provocation-card (marked last turn — same bug).
-- Backups already taken: _vonc_pc_backup_20260704, _vonc_cc_lobby_backup_20260704.
-- ============================================================

-- ── lobby-grid: template ──
UPDATE content_components
SET html_template = REPLACE(html_template,
      'data-component="lobby-grid" data-runtime-fill="true"]',
      'data-component="lobby-grid"]'),
    updated_at = NOW()
WHERE id = '9304f14d-e19b-4ce1-b3fd-f6a315aec6ed'
  AND html_template LIKE '%data-component="lobby-grid" data-runtime-fill="true"]%'
RETURNING (html_template LIKE '%data-component="lobby-grid" data-runtime-fill="true"]%') AS still_broken,
          (html_template LIKE '%data-component="lobby-grid" data-runtime-fill="true" %') AS section_marker_kept;

-- ── lobby-grid: current index instance ──
UPDATE page_components
SET rendered_html = REPLACE(rendered_html,
      'data-component="lobby-grid" data-runtime-fill="true"]',
      'data-component="lobby-grid"]'),
    updated_at = NOW()
WHERE component_id = '9304f14d-e19b-4ce1-b3fd-f6a315aec6ed'
  AND page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'
  AND rendered_html LIKE '%data-component="lobby-grid" data-runtime-fill="true"]%'
RETURNING (rendered_html LIKE '%data-component="lobby-grid" data-runtime-fill="true"]%') AS still_broken,
          (rendered_html LIKE '%data-component="lobby-grid" data-runtime-fill="true" %') AS section_marker_kept;

-- ── provocation-card: template ──
UPDATE content_components
SET html_template = REPLACE(html_template,
      'data-component="provocation-card" data-runtime-fill="true"]',
      'data-component="provocation-card"]'),
    updated_at = NOW()
WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'
  AND html_template LIKE '%data-component="provocation-card" data-runtime-fill="true"]%'
RETURNING (html_template LIKE '%data-component="provocation-card" data-runtime-fill="true"]%') AS still_broken,
          (html_template LIKE '%data-component="provocation-card" data-runtime-fill="true" %') AS section_marker_kept;

-- ── provocation-card: current index instance ──
UPDATE page_components
SET rendered_html = REPLACE(rendered_html,
      'data-component="provocation-card" data-runtime-fill="true"]',
      'data-component="provocation-card"]'),
    updated_at = NOW()
WHERE component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1'
  AND page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'
  AND rendered_html LIKE '%data-component="provocation-card" data-runtime-fill="true"]%'
RETURNING (rendered_html LIKE '%data-component="provocation-card" data-runtime-fill="true"]%') AS still_broken,
          (rendered_html LIKE '%data-component="provocation-card" data-runtime-fill="true" %') AS section_marker_kept;

-- Each row: still_broken should be f (malformed selector gone),
--           section_marker_kept should be t (the <section> tag keeps the marker).
