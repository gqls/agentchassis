# Proposal: Modular Page Assembly Architecture

**Status**: Draft for discussion  
**Created**: 2026-02-06  
**Context**: Current system uses regex injection for header/footer which is fragile. This proposal outlines a cleaner slot-based assembly approach.

---

## Problem Statement

Current page assembly:
1. Builds sections → concatenates → wraps in page structure
2. Uses regex to find and remove existing header/footer
3. Injects new header/footer via regex

Issues:
- Regex can't understand document structure
- Components have varying structures (styles inside/outside tags)
- Mixed responsibilities in single agents
- Difficult to debug
- Not reusable for other workflows

---

## Proposed Architecture: Slot-Based Assembly

### Core Principle

Treat a page as **slots to be filled**, not a string to be mutated:

```
Page Structure:
├── doctype
├── html_open
├── head_slot          ← site-level, rendered once
├── body_open
├── header_slot        ← site-level, rendered once
├── sections[]         ← page-specific, each rendered independently
├── footer_slot        ← site-level, rendered once
├── body_close
└── html_close
```

Assembly = concatenation, not injection.

---

## Specialist Agents (Each Independently Reusable)

### 1. site-planner
**Responsibility**: Plan pages and sections  
**Input**: Brief, domain  
**Output**: Structured site plan (pages, sections, navigation structure)  
**Dependencies**: None  
**Reusable for**: Any workflow needing site structure

### 2. site-component-renderer
**Responsibility**: Render site-level components (header, footer, head)  
**Input**: Site plan, style collection, navigation  
**Output**: Rendered HTML stored in `site_components` table  
**Dependencies**: Navigation structure (all pages known)  
**Reusable for**: Rerender workflows, style updates

### 3. section-content-writer
**Responsibility**: Generate content for ONE section  
**Input**: Section type, schema, brief context  
**Output**: Validated JSON content (NOT HTML)  
**Dependencies**: Brief data  
**Validation**: Against section schema before storing  
**Reusable for**: Any content generation workflow

### 4. link-manager
**Responsibility**: Manage internal links and navigation  
**Input**: All pages, section content  
**Output**: Link mappings, navigation HTML  
**Dependencies**: All pages must exist  
**Reusable for**: Nav updates, sitemap generation, link validation

### 5. page-assembler
**Responsibility**: Pure assembly (no generation)  
**Input**: Pre-rendered header/footer/head, section content (JSON)  
**Output**: Complete HTML page  
**Dependencies**: All content pre-generated  
**Reusable for**: Rerender, template changes

### 6. meta-manager
**Responsibility**: Page-level meta (title, description, OG tags)  
**Input**: Page content, site info  
**Output**: Head meta elements  
**Dependencies**: Page content exists  
**Reusable for**: SEO updates, social sharing fixes

### 7. site-finalizer
**Responsibility**: Final site-wide operations  
**Input**: All assembled pages  
**Output**: Sitemap, robots.txt, deployment trigger  
**Dependencies**: All pages deployed  
**Reusable for**: Post-deploy hooks

---

## Data Flow

```
Phase 1: Planning
  site-planner → site_plan (JSON)
                 └→ pages table populated

Phase 2: Site-Level Components  
  site-component-renderer → site_components table
                            ├── header (rendered HTML)
                            ├── footer (rendered HTML)
                            └── head (rendered HTML)

Phase 3: Content Generation (parallelizable per section)
  section-content-writer → page_sections table
                           └── content (validated JSON, NOT HTML)

Phase 4: Link Resolution (after all content exists)
  link-manager → navigation HTML
                 └→ internal link mappings

Phase 5: Page Assembly (per page)
  page-assembler → concatenates slots
                   └→ complete HTML (no regex)

Phase 6: Finalization
  site-finalizer → sitemap.xml, deployment
```

---

## Storage Changes

### Current
- `page_components`: Stores rendered HTML per section
- Header/footer injected at runtime

### Proposed
- `page_sections`: Stores validated JSON content per section
- `site_components`: Stores pre-rendered header/footer/head
- Rendering happens at assembly time from JSON

**Benefit**: JSON storage allows re-rendering with different templates without regenerating content.

---

## Validation at Ingestion

```go
// Current (hope for the best)
content := llm.Generate(prompt)
rendered := renderTemplate(template, content)

// Proposed (validate first)
content := llm.Generate(prompt)
validated, err := validateAgainstSchema(content, sectionSchema)
if err != nil {
    validated = retryOrFallback(sectionSchema)
}
storeSection(pageId, sectionType, validated) // JSON
// Render only at assembly time
```

---

## Invalidation Rules

| Change | What to Re-run |
|--------|----------------|
| Brief changes | section-content-writer (affected sections) |
| Navigation changes | site-component-renderer (header/footer), link-manager |
| Style collection changes | site-component-renderer, page-assembler |
| Template changes | page-assembler only (content unchanged) |
| New page added | link-manager, site-component-renderer (nav) |

---

## Open Questions

1. **Section storage granularity**: One JSON blob per section, or structured further?  I don't know let's discuss - probably start with one json blob but design for extension into more granular storage

2. **Link resolution timing**: During content generation, or as post-process? I think initial link generation should work (no intentional bad links, if there are new links then content should be created) but a link checker agent would be wanted too and potentially a sophisticated post process link agent or set of agents that can sweep any site would be a great next stage.

3. **Agent orchestration**: Central orchestrator spawns all, or agents spawn their dependencies? agents should spawn their own dependencies, taking care to keep the paths to data clear and fault free

4. **Backward compatibility**: Migration path from current system? no, we don't need migration

5. **Error handling**: If one section fails validation, what happens to the page? we flag it and carry on and a rerender agent that we already have as a nascent "maintenance" agent can fix things

6. **Caching**: How long are site_components valid? TTL or event-based invalidation? we should store site components for reuse or rejigging in other sites if they work well, if not they can be further down a priority or classification list, some will be good for some designs of site (CTA for a sales site eg AIDA, long text sections for long form copy type site) we can discuss this

7. **Preview mode**: Can we assemble without deploying for review? yes. but each agent can have hitl if we enable it. I think we have this already. default is that the flow is primarily automatic but we will probably make it more and more hitl as people will want more control over everything. 

---

## Next Steps

1. Finish fixing current system (regex improvements)
2. Design detailed schemas for section content
3. Prototype section-content-writer with validation
4. Prototype page-assembler with slot concatenation
5. Migrate one page type as proof of concept
6. Gradually migrate remaining functionality

---

## Related Files

- Current page-content-writer: `agent_definitions` table
- Current pageflow-builder: `agent_definitions` table
- Inject functions: `platform/orchestration/actions/render_actions.go`
- Component templates: `content_components` table