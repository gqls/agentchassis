-- p4_20 — point /report.html's CTAs at the request form instead of /contact.html
--
-- WHY (measured 2026-07-28, not guessed):
--   The form anchor `id="request-a-report"` sits at byte 31,326 of 43,224 — 72%
--   down the page. Worse, the hero's primary button, the most prominent control
--   on the page and the one a reader hits first, rendered:
--
--       <a href="/contact.html">Request a Verified Idea Report</a>
--
--   ...so the visitor who DID act on the obvious call to action was taken away
--   from the purchase entirely. The section immediately above the form
--   (call-to-action, position 4) rendered no link at all — a dead label.
--
-- MECHANISM: `cta_url` and `primary_cta_url` are `source: renderer` with NO
--   fallback in the component input_schema. content_data carried the CTA *text*
--   but no URL, so the renderer's hardcoded /contact.html default filled the gap.
--   Same shape as the home-page CTA fix (p3_05-07). Because these fields are
--   source=renderer they are NOT recomputed on a section_data_resolved rerender,
--   so a value set in content_data holds — which is exactly why that fix worked
--   and why this one will.
--
-- SAFETY:
--   * Both rows are UNLOCKED (checked: locked_at IS NULL), so save_page_sections
--     will not discard the rerender — the p4_08 trap in reverse.
--   * No section on this page has NULL content_data (checked), so the rerender
--     cannot escalate to the LLM content writer and rewrite live copy.
--   * The hero template is gated `{{if and .cta_text .cta_url}}`, so if this
--     value were ever cleared the anchor disappears rather than resurrecting the
--     phantom /contact.html link. Correct-or-absent (LNK-005).
--   * secondary_cta_url is deliberately left unset on both: those labels then
--     render nothing, which is correct, rather than inventing a destination.
--
-- ROLLBACK: re-run with '#request-a-report' replaced by removing the key:
--   content_data = content_data - 'cta_url'   (and - 'primary_cta_url')

\set page_id '41333d74-0c5a-4e12-b942-50ba4df793e6'

-- BEFORE
SELECT position, slot_name,
       content_data->>'cta_url'         AS cta_url,
       content_data->>'primary_cta_url' AS primary_cta_url
FROM page_components
WHERE page_id = :'page_id' AND position IN (1, 4)
ORDER BY position;

BEGIN;

-- hero (position 1): the early, prominent path to the form
UPDATE page_components
SET content_data = jsonb_set(content_data, '{cta_url}', '"#request-a-report"'::jsonb, true),
    updated_at   = now()
WHERE page_id = :'page_id'
  AND position = 1
  AND locked_at IS NULL;

-- call-to-action (position 4): sits directly above the form and rendered a
-- dead label; give it the same anchor so the CTA is live wherever the reader is.
UPDATE page_components
SET content_data = jsonb_set(content_data, '{primary_cta_url}', '"#request-a-report"'::jsonb, true),
    updated_at   = now()
WHERE page_id = :'page_id'
  AND position = 4
  AND locked_at IS NULL;

COMMIT;

-- AFTER — both must now show #request-a-report
SELECT position, slot_name,
       content_data->>'cta_url'         AS cta_url,
       content_data->>'primary_cta_url' AS primary_cta_url
FROM page_components
WHERE page_id = :'page_id' AND position IN (1, 4)
ORDER BY position;
