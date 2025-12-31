Looking at the existing patterns, I can see:

- **assets** - site_id scoping, origin tracking (generated vs uploaded), purpose/type
- **relationships** - flexible entity linking via source/target entity types
- **page_components** - content_data inline, links to research_results for provenance
- **research_results** - research findings that *inform* content creation

What's missing: **content_items** - the actual reusable text content that:
- Can be referenced by page_components (instead of embedded content_data)
- Can be shared across sites (library content)
- Tracks origin like assets doHere's how this fits together:

**The new architecture:**

```
┌─────────────────────────────────────────────────────────────────┐
│                         CONTENT LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  content_items                                                   │
│  - Reusable text content ("what to say")                        │
│  - site_id=NULL for library content                             │
│  - origin tracking (which agent, from which research)           │
│  - content_type: headline, service_description, testimonial...  │
└──────────────────────────┬──────────────────────────────────────┘
                           │ content_item_id
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                         LAYOUT LAYER                             │
├─────────────────────────────────────────────────────────────────┤
│  page_components                                                 │
│  - Instances on pages ("how to show it")                        │
│  - References content_items OR has inline content_data          │
│  - component_id → content_components (templates)                │
│  - rendered_html = final output                                 │
└──────────────────────────┬──────────────────────────────────────┘
                           │ page_id
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                       STRUCTURE LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  pages                                                           │
│  - Metadata, navigation, workflow                               │
│  - sections = planning reference                                │
│  - build_status = workflow state                                │
└─────────────────────────────────────────────────────────────────┘
```

**Key features of content_items:**

| Feature | How it works |
|---------|--------------|
| Reuse within site | Same content_item referenced by multiple page_components |
| Reuse across clients | `is_library=true`, `site_id=NULL` |
| Industry collections | `industry_vertical='consulting'`, `library_tags=['professional-services']` |
| Origin tracking | `origin_agent`, `origin_research_id` - like assets |
| Content override | page_components.content_data merges/overrides content_item |
| Versioning | `version` column + can use `origin_content_id` for derived content |

**Workflow change:**

```
Before: content writer → page_components.content_data (inline)
After:  content writer → content_items → page_components.content_item_id
```

This means the same tagline can appear in hero, footer, and meta description without duplication, and library content can seed new sites.

Should I also update the pages revert migration to remove the extra columns, keeping only the minimal set we discussed?