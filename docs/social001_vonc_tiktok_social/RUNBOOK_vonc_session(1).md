# Runbook — vonc.com build session

## What was diagnosed and fixed

### 1. `write_site_spec` — spec_data string rejection (FIXED, deployed)

**File:** `platform/orchestration/actions/site_spec_actions.go`

**Symptom:** `persist_mission_brief` and `persist_roadmap_brief` steps failed with
"spec_data must be a JSON object, got string" for any submission that passes
`mission_brief` / `roadmap_brief` as plain text strings.

**Cause:** `WriteSiteSpecAction` did a hard type assertion to
`map[string]interface{}` on `spec_data` with no string coercion path.
The domain-submitter workflow resolves `"spec_data": "input_data.mission_brief"`,
which is a plain string in the submission payload.

**Fix:** Added coercion block between the nil check and the type assertion:
- JSON string → parse to object (matches LLM output paths)
- Plain string → wrap as `{"text": value}` (matches brief fields)
- Objects pass through unchanged

This aligns with how the classifier prompt reads the stored value:
`{{.site_specs.specs.mission_brief.text}}`.

**Also noted:** The submission script `004_submit_vonc_trigger.sh` builds
`MESSAGE_BODY` via python3 but never sends it — the `kcat` call mid-script
uses a hardcoded inline `<<JSON` body. Both carry the same payload for vonc.com
so the run was unaffected, but the script should be tidied for future use.

---

### 2. `gauntlet-interface` component — broken template slots (FIXED in DB)

**Symptom:** `/tools/gauntlet/index.html` deployed with literal placeholder
strings (`eyebrow_label`, `challenge_title`, etc.) and `</no>` closing tags
visible in the HTML.

**Cause:** The `html_template` in `content_components` for `gauntlet-interface`
contained `<no value>FIELDNAME</no>` throughout — a Go text/template render
artifact stored as source. `render_component` rendered the template verbatim;
`RenderTemplate`'s `<no value>` cleanup pass stripped the prefix, leaving the
literal field name and `</no>` tag.

**DB fix applied:**
```sql
-- Step 1: replace <no value>FIELD</no> → {{\.FIELD}}
UPDATE content_components
SET html_template = regexp_replace(html_template, '<no value>([^<]+)</no>', '{{\.\1}}', 'g'),
    updated_at = now()
WHERE id = '5da50747-7936-4b8f-a66d-c1ea98919c75';

-- Step 2: fix backslash artifact → {{.FIELD}}
UPDATE content_components
SET html_template = regexp_replace(html_template, '\{\{\\\.',  '{{.', 'g'),
    updated_at = now()
WHERE id = '5da50747-7936-4b8f-a66d-c1ea98919c75';
```

Verified: `{{.eyebrow_label}}` correct, 0 remaining `<no value>` artifacts.

**Rerender queued:** `manual-rerender-gauntlet-fix-<uuid>` inserted into
`site_work_items` for tool-gauntlet.

---

### 3. `archetype-result-card` component — template needs regeneration (IN PROGRESS)

**Symptom:** 30 `<no value>` artifacts in `html_template`, zero `</no>` closing
tags. Quality score 30. Quality issues: "0 template variables".

**Cause:** Different failure mode from gauntlet — the template was rendered
against an empty context and the output (with all `{{.field}}` variables
resolved to `<no value>`, then cleaned to empty by `RenderTemplate`) was stored
back as the source. Field names are gone; no string repair is possible.

**Schema:** Intact. 30 fields with full LLM guidance in `input_schema`.
Component ID: `2c7678fb-9940-428d-8b78-62e2510f6dbe`.

**Work item raised:**
```sql
INSERT INTO site_work_items (..., item_type = 'needs_component_regeneration',
    handler_agent = 'component-creator', status = 'triaged', ...)
```
`component-creator` will call the LLM with `input_data`, `site_record`, and
`site_specs` (design_intent, content_direction, identity all present for vonc.com).
`StoreGeneratedComponentAction` will validate and reject if `<no value>` appears.
After regeneration, a rerender of tool-gauntlet will be needed.

---

### 4. Index page brochure-blue appearance (RERENDER QUEUED)

**Symptom:** Deployed index looked generic blue/brochure rather than dark Spark theme.

**Cause:** The first `page_rerender_index` fired at 01:22, before the fresh
dark-theme `page_components` were saved at 02:41. A second rerender
(`page_rerender:index`) claimed at 02:23 and completed at 02:43 — this one
should have assembled the correct components. The DB `rendered_html` for all
six index sections contains correct dark-theme content.

**Status:** Manual rerender queued (`manual-rerender-index-confirm-<uuid>`).
Once it completes, check the deployed HTML to confirm dark theme is live.

---

### 5. `needs_page:index` original build failure

**Symptom:** `needs_page:index` work item from 17:13 failed with "Claim timed
out (attempts exhausted)" — no page-content-writer ran for index in the initial
build window.

**Cause:** Pod handling the claim died (deploy-mid-work-item or handler timeout).
No error in `agent_error_log` for this orchestration beyond the two
`write_site_spec` failures — confirms pod death rather than logic error. A later
`page_rerender:index` item triggered a successful build.

**Actionable:** None for this run. For future submissions, the build pipeline
should be stable without mid-deploy interference.

---

### 6. `provocations/index.html` 404

**Cause:** Stale `provocations.html` from a previous run exists at repo root; the
current plan uses `/provocations/index.html` (section-index URL). The old file
was not from this submission.

**Action required:** Remove `provocations.html` from the repo root before the
next submission or trigger a rerender of `provocations-index` to overwrite.
The `provocations-index` page_components are current (built in this run).

---

## Code changes produced this session

All files are in `/mnt/user-data/outputs/` (or previously delivered):

| File | Change |
|---|---|
| `site_spec_actions.go` | String coercion for `spec_data` in `WriteSiteSpecAction` |
| `check_component_standards.go` | Added `checkBrokenTemplateSlots` sub-check |
| `fix_component_template_action.go` | Added `repair_template_slots` fix type + `repairNoValueSlots` helper; distinguishes repairable (`</no>` present) from unrepairable (needs regeneration) |

---

## Confirmed state (as of 2026-06-23 ~13:22)

- `archetype-result-card`: regenerated cleanly. quality_score=100, template_variable_count=28, no_value_count=0.
- `tool-gauntlet`: rerendered at 13:12-13:13 after template fix and regeneration.
- `index`: rerendered at 13:11-13:12.
- All other pages: rerendered in the 13:xx batch.
- `page_rerender_provocation`: claimed 13:21, completion pending at time of check.

The three manual rerender items we inserted all failed with `page_id not found in input`.
The `page-rerender` action requires `page_id` (UUID), `domain`, `filename`, and `page_name`
in the spec. Our inserts only had `page_name`. They were superseded by the pipeline-generated
batch which ran correctly.

## Correct spec shape for manual rerenders

```sql
-- Step 1: get page_id and url
SELECT id, name, url FROM pages
WHERE site_id = '<site_id>' ORDER BY name;

-- Step 2: insert with full spec
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, pipeline
) VALUES (
    '<site_id>', 'manual', 'page_rerender', 'medium', 'Manual rerender of <page>',
    jsonb_build_object(
        'domain',    '<domain>',
        'page_id',   '<id from pages>',
        'page_name', '<name>',
        'filename',  '<url without leading slash>'
    ),
    50, 'page-rerender', 'triaged', 'manual',
    'manual-rerender-<page>-' || gen_random_uuid(), 'build'
);
```

Example from pipeline (provocations-index):
```json
{"domain": "vonc.com", "page_id": "e4b3b195-919f-45ad-854e-201d3e846ea8",
 "filename": "provocations/index.html", "page_name": "provocations-index"}
```

## Remaining steps

**1. Visual check**
Visit vonc.com — index, gauntlet, archetypes. Confirm dark theme, Spark branding,
no placeholder strings visible.

**2. Remove stale `provocations.html` from repo root**

This file predates the current submission. Leaving it causes confusion about
the site structure. Either delete it directly from the git repo or trigger a
fresh page build for `provocations-index` which will overwrite the path.

**6. Unresolved CTA work items — review when site has more pages**
Two `unresolved_cta` items are parked at `needs_human_review`:
- `unresolved_cta_index_hero` — no section-index hub to link to from the
  hero CTA (the resolver found no eligible content hub)
- `unresolved_cta_archetypes_*` — same

These will resolve naturally once section-index pages (provocations-index,
gauntlet hub) are active and the resolver can find them. Not blocking.

**7. `needs_section_data` items — manual input or defer**
- `section_data_archetypes_archetype-combinations` — needs combination data
- `section_data_about_content-block-about` — needs about content

These require human input or a content pass. Not blocking the build.

---

## Structural findings to carry forward

**Two distinct broken-template failure modes exist in the component library:**

- **Mode A** (`<no value>FIELD</no>`): Template generated with correct `{{.field}}`
  syntax, then rendered and re-stored with field names preserved as fallback text.
  Repairable by string substitution. `repair_template_slots` handles this.

- **Mode B** (bare `<no value>`): Template generated correctly, rendered against
  empty context, cleaned output stored. Field names lost. Requires regeneration.
  `repair_template_slots` now detects this and returns `action: needs_regeneration`
  rather than attempting a futile repair.

**`StoreGeneratedComponentAction` already rejects both modes** at the gate (line 304).
The `checkBrokenTemplateSlots` discovery check now surfaces any that slipped
through the gate (pre-dating the check or via paths that bypass it).

**The `needs_page:index` pod-death pattern** is worth watching. If it recurs on
the next submission, consider whether the `page-build-handler` timeout
(currently 1800s in the workflow) is sufficient for the content-writer call,
or whether the pod health check initial delay needs tuning for heavier pages.
