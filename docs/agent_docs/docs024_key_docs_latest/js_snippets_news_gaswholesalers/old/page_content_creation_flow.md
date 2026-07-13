# Page Content-Creation Flow

How a page goes from a `pages` row (sections list) to populated, deployed
HTML with `page_components` rows carrying `content_data` and
`component_id`. Traced from chassis source 2026-05-20 while diagnosing the
FAQ empty-items bug. Marks clearly what is verified in code vs what the
isolated build test is designed to pin down.

## Entry: how a build is initiated

`page-build-handler` is the agent that builds a page. It is reached two
ways:

1. **Via work item** (normal path). A `site_work_items` row with
   `handler_agent = 'page-build-handler'` is claimed by the
   build-dispatch-loop, which spawns the handler with the item's `spec`
   as `input_data.spec`. Observed spec shape (from live data):
   ```json
   {
     "reason": "not_built",
     "plan_id": "<uuid>",
     "page_name": "contact",
     "page_role": "content"
   }
   ```
   Note what the spec does NOT contain: the sections list. The handler
   reads sections from the `pages` row, not the work item.

2. **Direct orchestrate** (manual/test). A Kafka message to
   `system.agent.generic.requests` with `action=orchestrate`,
   `config.agent_type=page-build-handler`, and `input_data.spec.page_name`
   (or `page_id`). Same contract — the handler still loads the page row.

## The page-build-handler workflow (step by step)

`start_step = ensure_site_record`. The relevant chain:

```
ensure_site_record
  → load_page_record       (load the pages row: sections, title, page_type)
  → check_page_found
  → plan_sections          (resolve each section's data readiness)
  → check_has_ready_sections
  → spawn_content_writer    (generate content for ready sections)
  → [content writer runs: per-section generation + render]
  → compile / assemble      (CompilePageSectionsAction → sections_metadata)
  → deploy_page             (git commit)
  → save_page_sections      (persist page_components: html, content_data, component_id)
  → ...
```

### load_page_record — establishes the section list

`LoadPageRecordAction` looks up the page by `site_id` +
`page_name` (preferred) or `page_id`, and returns the full record
including parsed `sections`. Config:

```json
"config": {
  "page_id":   "input_data.spec.page_id",
  "site_id":   "site_record.site_id",
  "page_name": "input_data.spec.page_name"
}
```

It also has contract-compliant fallbacks for `page_name`
(`input_data.spec.page_name` → `input_data.spec.page.name` →
`current_page.name` → `page_record.name`).

**This is the authority for what sections the page has.** Whatever is in
`pages.sections` drives the rest of the build. For the FAQ page that was
`["hero", "generic-text-block", "faq", "call_to_action"]`.

### plan_sections — resolves data readiness per section

`PlanSectionsAction` reads each section's component `input_schema` and
triages each field by its `source`:

| Field `source` | Handling |
|---|---|
| `llm` | Added to `llm_fields` — the content writer must generate it. Section stays "ready". |
| `query.*` | Resolved now via the `queryresolve` package (SQL). Empty result → empty slice. |
| `renderer` / `static` | Deferred to render time — always "available". |
| `site_specs.*`, `pages.*`, `site_assets.*`, `config.*` | Resolved against current specs/pages/assets. Missing + required → `on_missing` rule (defer / skip / fallback). |

Section status becomes `ready` / `deferred` / `skipped`. Only `ready`
sections are passed to the content writer.

For the `faq` component, `input_schema.fields.questions` is
`type: array, items: {question, answer}, min_items: 3, source: llm,
required: true`. Because `source = llm`, plan_sections marks `questions`
as an LLM field and the section as **ready**. So the FAQ section is
correctly identified as "writer must generate a `questions` array." The
triage is not the bug.

`plan_sections` also has a `sectionDescription` resolver that already
looks for per-section briefs in the plan under three shapes:
`page.section_descriptions[sectionType]`,
`page.section_types[].description`, and a fall-back to `page.purpose`.
**The consumer for richer plans exists; the planner just doesn't emit any
of these shapes yet** (it emits bare section-name strings). See
`site_planner_depth_and_freshness_concerns.md`.

### content writer — generates content for ready sections

The writer is spawned per the workflow and processes ready sections (a
`process_sections_loop`). For each section it builds a prompt and calls
the LLM, then the result is consumed by `extractResponseContent`.

**Verified:** `extractResponseContent` returns the LLM result as a flat
**string** — it walks the response for `*_content` / `*_result` / `result`
/ `content` keys and returns the first string it finds. There is no
observed branch in this helper that parses the result into a structured
array of objects.

**The open question (what the test pins down):** for a component whose
schema needs a structured field (FAQ's `questions` = array of
`{question, answer}`), something must turn the LLM output into that array
and place it in the section's `content_data` under the key `questions`.
Whether the content-writer agent_definition's workflow (its prompt + a
parse step) does this is not visible in the Go action source — it lives
in the agent_definition config. The render and save layers (below) both
faithfully carry whatever `content_data` they're handed, so if
`content_data.questions` is absent here, it is absent everywhere
downstream and the accordion renders empty. The isolated faq-only build
test is designed to determine whether the writer produces
`content_data.questions` at all.

### render — binds content_data into the template

`RenderComponentAction` builds a `RenderContext` whose `ContentData` map
starts with site-level defaults (company_name, nav, cta, etc.) and then
**merges the section's `content_data` on top** (section values win):

```go
// Now merge the actual section content_data on top — these take priority
for key, value := range contentData {
    renderCtx.ContentData[key] = value
}
```

The template then reads `{{.headline}}`, `{{range .questions}}`, etc.
from this merged map. **Verified correct:** if `content_data` contains a
populated `questions` array, the FAQ template renders populated
accordion items. If `content_data.questions` is empty/absent, the
`{{range}}` produces empty shells. The render layer is faithful — it is
not the bug.

Output of the render for each section: `rendered_html`, `content_data`,
`component_id`, `component_function`/`component_name`.

### compile — gathers sections into metadata

`CompilePageSectionsAction` assembles the rendered sections into the page
HTML and produces a `sections_metadata` array, each entry carrying
`rendered_html` + the component metadata (`component_id`,
`component_function`, `content_data`). This metadata array is the
structured hand-off to the save step.

### save_page_sections — persists page_components

`SavePageSectionsAction` has two paths:

1. **Structured metadata path** (preferred): reads
   `sections_metadata`, and for each entry builds a `SectionData`
   `{ComponentName, ComponentID, HTML, Position, ContentData}` via
   `extractSectionsFromMetadata`. `content_data` and `component_id` are
   taken straight from the metadata entry.
2. **HTML-parsing fallback**: regex-extracts `<section>` blocks from the
   assembled HTML when no metadata is present (adopted sites / older
   pipelines). This path has no `content_data` and looks up
   `component_id` from the `data-component` attribute.

Both paths then run `enrichSectionsWithPlannedNames` and
`enrichSectionsWithComponentIDs` to fill `component_id` from
`content_components.function` matching, before the INSERT into
`page_components` (sets `rendered_html`, `slot_name`, `component_id`,
`content_data`, `content_brief`, `build_status`).

**Verified failure mode (matches the FAQ page):** when a section's
metadata is incomplete — no recovered `component_id`/`component_function`
and no `content_data` — the section saves with `ComponentName="section"`,
`component_id` NULL (orphaned), and empty `content_data`. The
`extractSectionFromMap` comment in the source documents exactly this:
unrecovered metadata → `enrichSectionsWithComponentIDs` skips the section
→ `page_components.component_id` NULL. The FAQ page's positions 2 and 3
are orphaned in precisely this way.

## The data contract at each hop

| Hop | Carries page identity as | Carries content as |
|---|---|---|
| Work item / trigger | `spec.page_name` / `spec.page_id` | (none — sections come from DB) |
| load_page_record | `page_record` (with `sections`) | — |
| plan_sections | section names + component info | `llm_fields` (to generate), `resolved_data` (already available) |
| content writer | per-section | **section `content_data`** ← the critical construction |
| RenderComponentAction | per-section | merges `content_data` → template |
| CompilePageSections | `sections_metadata[]` | `content_data` + `component_id` per entry |
| save_page_sections | `page_id` + position | INSERT `content_data` + `component_id` |

The single point where structured content (`questions` array) must be
*constructed* is the content writer. Every layer after it merely carries
or binds what it was given. Every layer before it (triage, page load) is
verified to set the FAQ section up correctly as "needs an LLM-generated
`questions` array." That narrows the cause to the writer's
structured-field generation, which the isolated test confirms.

## Creating a build target (test pages)

`pages.url` is `NOT NULL` — the INSERT must set it. Corrected form:

```sql
INSERT INTO pages (site_id, name, url, page_type, sections,
                   status, build_status, title, in_header, in_footer)
VALUES (
  '<site-id>',
  'faq-test',
  '/faq-test.html',
  'content',
  '["hero", "faq", "call_to_action"]'::jsonb,
  'active', 'pending', 'FAQ Build Test',
  false, false          -- keep test page out of nav/footer
)
RETURNING id, name, url, sections;
```

`page_type` must satisfy `chk_page_type_kebab_case`. `in_header` /
`in_footer` default true — set false for a throwaway so it doesn't enter
navigation (the `trg_invalidate_nav_on_page_change` trigger fires on
insert regardless, which is harmless).

Cleanup after the test: `DELETE FROM pages WHERE site_id = '<site-id>'
AND name = 'faq-test'` (cascades to its `page_components`).

## Trigger (direct orchestrate)

```json
{
  "action": "orchestrate",
  "config": {"agent_type": "page-build-handler"},
  "input_data": {
    "spec": {
      "page_name": "faq-test",
      "page_role": "content",
      "reason": "not_built"
    }
  }
}
```

Sent to `system.agent.generic.requests` with the standard header set
(correlation_id, orchestration_id, request_id, message_id,
message_type=request, client_id, action=orchestrate, sender_agent_type,
responses_topic, timestamp), matching the single-page rerender trigger
already in use this session.

## What the test reads out

After the build completes:

```sql
SELECT pc.position, cc.function, pc.component_id IS NOT NULL AS linked,
       jsonb_typeof(pc.content_data) AS cd_type,
       CASE WHEN pc.content_data ? 'questions'
            THEN jsonb_array_length(pc.content_data->'questions') END AS q_count,
       LEFT(pc.rendered_html, 120) AS html_start
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_components cc ON pc.component_id = cc.id
WHERE p.site_id = '<site-id>' AND p.name = 'faq-test'
ORDER BY pc.position;
```

- **faq populated** (`q_count >= 3`, accordion has text) → the writer CAN
  build structured fields when faq stands alone; the live faq page failed
  because of the competing `generic-text-block` (planner duplicate-surface
  cause).
- **faq empty** (`q_count` null/0, empty shells) even with faq alone →
  the writer's structured-field generation is the cause; fix is
  schema-aware generation (prompt the LLM for the declared structure,
  parse into `content_data[field]`), which then benefits every structured
  component, not just FAQ.

Either way the fix is provable by re-running the identical trigger, and
the live faq page is then repaired by the corrected pipeline rather than
by hand.
