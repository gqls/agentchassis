https://claude.ai/chat/fbdaef1b-bb4c-45dd-88e5-34349bfe27bf

# Granular Site Editing — Architecture Report

## The Editing Spectrum

User edit requests range from "change one word" to "rewrite the whole page in a different format." The system needs to handle all of these through the same maintenance infrastructure, routing each to the right mechanism.

| Scope | Example | Mechanism | LLM needed? |
|-------|---------|-----------|-------------|
| Word/phrase | "Change 'Get Started' to 'Book a Call'" | Direct HTML patch | No |
| Field value | "Update the phone number" | Template re-render with new data | No |
| Section rewrite | "Rewrite the hero for SEO" | Content-writer on one section | Yes |
| Component swap | "Replace testimonials with portfolio" | Swap slot_name, re-render | Partial |
| Page rewrite | "Rewrite use-cases in PAS format" | page-rebuild with content_direction | Yes |
| Multi-page | "Refresh all stale pages" | maintenance-triage → page-rebuild | Yes |

## What Exists Today

**Data layer** — already supports granular editing:

- `page_components` stores each section separately with `rendered_html`, `slot_name`, `position`
- `content_components` holds templates with `input_schema` (the structured data each template needs)
- `page_components.content_data` is documented as "override/customization" — designed for exactly this purpose
- `page_components.content_snapshot` stores approved values for comparison and rollback
- `page_components.schema_mode` can be `strict` or `flexible` per section

**Component labeling** (new) — injects `data-pc-id`, `data-slot`, `data-position` into each `<section>` tag. This lets users and agents identify which DB row produced which visible section.

**`content_direction`** (new) — JSONB column on `pages` for page-level instructions that flow to the content-writer prompt.

**Existing agents and actions:**

- `page-rebuild` — rebuilds selected pages using existing brief + style
- `page-content-writer` — generates content per section (template or LLM)
- `CompilePageSectionsAction` — assembles sections into full page with header/footer
- `SavePageSectionsAction` — parses page HTML back into `page_components` rows
- `content-reviewer` — HITL review with HTML editing capability
- `render_component` — renders a template with data (Go template engine)
- `git_commit` + deployer — commits and deploys to Cloudflare

## What's Needed: One New Agent, Two New Actions

### section-editor agent

A single agent that handles the full range of section-level edits. It examines the edit instruction and decides whether to use LLM generation or algorithmic replacement.

**Workflow:**

```
load_edit_context → classify_edit → [branch] → render_updated_section →
  reassemble_page → commit_page → deploy_page → complete
```

**Step detail:**

1. `load_edit_context` (new action) — Given `site_id` + `page_component_id` (from the `data-pc-id` attribute), loads:
    - The target `page_component` row (current rendered_html, slot_name, position)
    - The matching `content_component` template and `input_schema`
    - The page record (for reassembly context)
    - All sibling `page_components` for the same page (for reassembly)

2. `classify_edit` (conditional step) — Examines the edit instruction to route:
    - **Field update** — instruction targets a specific schema field (e.g. "change headline to X"). Route to `apply_field_edit`.
    - **Component swap** — instruction says "replace with portfolio-showcase" or similar. Route to `apply_component_swap`.
    - **Content rewrite** — instruction is open-ended ("rewrite for SEO", "use PAS format"). Route to `generate_new_content` (LLM path).
    - **Direct HTML patch** — instruction targets specific text ("change 'Get Started' to 'Book a Call'"). Route to `apply_html_patch`.

3a. `apply_field_edit` (new action, algorithmic) — Parses the current `content_data` or extracts field values from rendered HTML via the template schema. Updates the specified field. Re-renders the template with updated data. No LLM.

3b. `apply_component_swap` (existing actions reused) — Changes the `page_component.slot_name` and `component_id` to the new component type. If the new component's `input_schema` differs, generates content for the new schema via LLM. If compatible, maps existing data to new template.

3c. `generate_new_content` (existing — calls content-writer) — Sends the section to the content-writer with the edit instruction as `content_direction`. The content-writer generates new content respecting the instruction, the template schema, and the site brief.

3d. `apply_html_patch` (new action, algorithmic) — Simple string replacement in the rendered HTML. For targeted text changes where the user knows exactly what they want.

4. `render_updated_section` — Re-renders the component template with the new data. Updates `page_components.rendered_html`.

5. `reassemble_page` (new action) — Loads ALL `page_components` for this page ordered by position. Concatenates their `rendered_html`. Injects header + footer + head (reusing `InjectHeader`, `InjectFooter`, `InjectHead` from `component_library.go`). Produces the full page HTML. This is the inverse of `SavePageSectionsAction` — it builds a page from stored sections rather than parsing a page into sections.

6. `commit_page` + `deploy_page` — existing `git_commit` and deployer actions, unchanged.

### New actions needed

**`load_edit_context`** — DB reads only. Loads the target section, its template, the page, and all siblings. Returns everything needed for any edit path.

**`reassemble_page`** — Algorithmic. Reads all `page_components` for a page from DB, concatenates in position order, injects site-level components (header/footer/head). Produces complete HTML. This action is also useful for the section-rewriter described in the v2 maintenance architecture plan — it's the same reassembly logic.

**`apply_field_edit`** — Algorithmic. Takes a field name + new value, updates the component's data, re-renders the template. Uses the existing `render_component` logic.

**`apply_html_patch`** — Algorithmic. Simple find-and-replace in rendered HTML. For when the user says exactly what to change.

The content generation path (3c) reuses the existing content-writer agent — no new LLM action needed.

## How Edits Flow Through the System

### User-initiated edit (manual)

The user inspects the page, identifies the section via `data-pc-id`, and submits an edit request:

```sql
-- Example: swap social-proof for portfolio-showcase on index page
INSERT INTO maintenance_queue (site_id, task_type, priority, reason, payload)
VALUES (
  '4851f6fc-...', 'section_edit', 3, 'user_request',
  '{
    "page_component_id": "3f542adb-...",
    "edit_type": "component_swap",
    "new_component": "portfolio-showcase",
    "content_direction": {
      "instruction": "Showcase actual sites built by the platform. No fabricated testimonials.",
      "projects": ["leopardessconsulting.co.uk", "other-domain.co.uk"]
    }
  }'
);
```

Maintenance-triage (or a direct trigger) claims the task and dispatches section-editor.

### Automated edit (from maintenance scan)

The scan detects stale content or other issues and inserts a task with `edit_type: "section_rewrite"` and appropriate direction. Same flow, no human trigger needed.

### Page-level rewrite with content_direction

For broader changes (rewrite use-cases in PAS format), the existing page-rebuild agent handles this:

```sql
UPDATE pages SET
  content_direction = '{
    "format": "problem-agitate-solution",
    "instruction": "Each use case: present a real problem, agitate the pain, suggest what COULD be done. Not solved — suggestions only.",
    "use_cases": ["automated website generation", "content refresh", "report generation"]
  }',
  build_status = 'needs_rebuild'
WHERE name = 'use-cases' AND site_id = '4851f6fc-...';
```

Then trigger page-rebuild. The content-writer receives `content_direction` alongside the brief and uses it to shape the output.

## content_direction Flow

`content_direction` needs to reach the content-writer's prompt. The flow:

1. `get_pages_to_build` action already queries page records — add `content_direction` to its SELECT
2. The page record flows as `current_page` into the content-writer's loop
3. The content-writer's prompt template checks for `current_page.content_direction` and includes it when present
4. The LLM uses it as the primary guide for that page's content

Change needed: one line added to the `get_pages_to_build` query (add `content_direction` to the SELECT), and a conditional block in the content-writer prompt.

## Implementation Priority

1. **Apply now** (ready): portfolio-showcase component, content_direction column, component labeling
2. **Small change** (affects one query + one prompt): content_direction flowing through to content-writer
3. **New agent** (section-editor): the main new work — handles all section-level edits
4. **New action** (reassemble_page): needed by section-editor and future section-rewriter

The section-editor agent and reassemble_page action are the only significant new code. Everything else builds on existing infrastructure. The `page_components` table was designed for this — the `content_data`, `content_snapshot`, and `schema_mode` columns exist specifically to support granular editing and rollback.