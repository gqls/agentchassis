# FUTURE — `needs_section_data` Handler

> **SUPERSEDED 2026-05-06 by `FOCUS_directory_builder_and_list_components.md`.**
>
> When this doc was written I had assumed `needs_section_data` items meant "fetch list data for a section." Looking at the source of `plan_sections_action.go` and the contract in doc 003, that's wrong. Two things:
>
> 1. `needs_section_data` items are emitted with `status='needs_human_review'` directly. They mean "couldn't resolve component or required field — get a human to look." Not async dispatch.
> 2. The actual mechanism for list/directory section data is `query.{name}` source resolution per doc 003, handled by a `directory-builder` agent per doc 002 line 776 — both already decided in the project, never built.
>
> See the new focus doc for the correct architecture. Keeping this file for historical context only.

---

(original content below)

---

**Status:** Not built. 41 items system-wide currently stuck at `needs_human_review` because nothing dispatches them.
**First observed:** 2026-05-02 (one item completed via unknown manual route; nothing since).
**Surfaced clearly:** 2026-05-06 during gamesdesign Phase 1 verification.

---

## What is `needs_section_data`?

When `plan_sections` (in `page-build-handler`) processes a page, each section in the page's `sections` array gets resolved against the `content_components` registry. Most sections need only **content fields** (the LLM produces hero text, CTA copy, etc. via `call_content_writer`). Some sections need **derived data**: a list of cards drawn from other pages on the site, a directory of entities, a feed of news items.

For the latter type, `plan_sections` (or possibly a downstream step — verify in source) emits a `needs_section_data` work item with this shape:

```
item_type:   needs_section_data
item_key:    section_data_<page_name>_<component_function>_<site_id>
spec:        { site_id, page_name, section_name (= component function),
               component_id, page_record, ... }
status:      triaged
handler_agent: (empty)
```

The work item key encodes the (page, component, site) triple so dedupe works naturally.

The handler's job is to **produce the data the section needs to render**, write it into wherever the renderer reads from (likely `page_components.content_data` keyed by slot or section name), and mark the work item complete.

---

## What's wrong now

1. **No agent claims `needs_section_data` items.** The `handler_agent` column is empty — neither `plan_sections` nor any subsequent step populates it. Without a handler, dispatch has nothing to route to.
2. **Items eventually flip to `needs_human_review`**, presumably by a stuck-task reaper that detects a triaged item with no claim activity for some threshold. Once at HITL, they sit forever.
3. **The single completed item from 2026-05-02** (`section_data_provocations_provocation-feed_e1e22a7d-...`) has `claimed_by = NULL`. Was probably marked complete by some manual or one-off path, not via handler dispatch.

Net effect: any page with a directory/list/grid section cannot finish building cleanly. Across six sites in production today, this affects 41 items.

---

## What the handler should do

For each `needs_section_data` item, the handler agent needs to:

### 1. Read the work item spec

`spec.site_id`, `spec.page_name`, `spec.section_name`. From that, look up the component:

```sql
SELECT id, function, input_schema, html_template
FROM content_components
WHERE function = $section_name AND is_active = true
```

The `input_schema` is authoritative for what fields the renderer expects.

### 2. Determine the data source

The handler needs to know **how to populate the list/grid/feed**. There are at least three flavours:

**(a) Site-internal directory/listing.**
For a `tool-list` section, the data is "all `page_type = tool` pages on this site" — name, slug, url, title, meta_description, maybe a representative icon/image.

For a `guide-list`, "all `page_type = blog_post` pages where `parent_section = guides`" or similar.

For a `game-list`, "all `page_type = entity_page` (or `game`) pages on this site."

The query target is the `pages` table on the same `site_id`, filtered by `page_type` (and optionally `parent_section`).

**(b) Cross-site / external feed.**
A section like `news-listing` might draw from external news feeds (we have a `news-feed-pipeline` per `006_news_feed_pipeline_v2.md`). The handler would need to know which feed to pull from based on classification/industry tags.

**(c) Static/site-spec data.**
Some sections need data that lives in `site_specs` (e.g. team members, services, departments). Read from `site_specs.aspect = '<aspect>'` and project to the schema's expected shape.

The handler decides which flavour applies based on the section's `function` value or a routing hint in the spec.

### 3. Project to schema-shaped output

The component's `input_schema.fields` says what the renderer wants. For a `tool-list`, the schema might define fields like:

```
items: [{ title, slug, url, description, icon }]
heading, subheading, eyebrow (LLM-driven content)
cta_label, cta_url (Tier B labels with fallback)
```

The handler produces an `items` array drawn from the data source, then merges with whatever the LLM already wrote (heading etc. come from `call_content_writer`).

### 4. Write to wherever the renderer reads

This is the part that needs investigation. Most likely target: `page_components.content_data` (a JSONB column keyed by slot/section). The page-build-handler already writes content from `call_content_writer` there. The section_data handler appends/merges its data for the relevant slot.

### 5. Mark the work item complete

Standard work-item completion: `status = 'complete'`, `completed_at = NOW()`, `claimed_by = '<agent_id>'`.

### 6. Trigger or signal that the page can resume building

If page-build-handler is blocked waiting for section_data, it needs to know it can proceed. Two patterns possible:
- **Polling**: page-build-handler iterates and detects all section_data items for this page are complete → continues.
- **Event-driven**: section_data handler emits a `section_data_ready` signal back to the parent page-build orchestration.

The current page-build-handler workflow may already have this loop; needs to be checked.

---

## Open design questions

1. **Is this a new dedicated agent, or a step inside `page-build-handler`?**
   - Dedicated agent: clean separation; dispatch routes by `item_type`.
   - Inline step: simpler; page-build-handler does it before returning to render. Less parallelism (page can't proceed until data is ready), but no cross-orchestration coordination needed.

   *Recommendation: dedicated agent.* It mirrors the `component-creator` pattern (plan-time emission, separate handler, completion signals upstream). Lets multiple section_data items on the same page resolve in parallel.

2. **How does the handler know what data source to use?**
   - Hardcoded mapping: `tool-list` → query pages table for tools.
   - Component-driven: the component's `input_schema` carries metadata declaring its data source (e.g. `"data_source": "pages.where(page_type=tool)"`).
   - LLM-routed: a small classifier decides per-component.

   *Recommendation: component-driven with a small declarative DSL.* Keeps the handler generic; new list components don't need handler code changes.

3. **What about ordering / sorting / limits?**
   - A site may have 50 tool pages but the index should show 8. Component schema needs to declare `limit`, `order_by`, etc.

4. **Cache invalidation.**
   - When a new tool page is added or deleted, the index page's `tool-list` section is now stale. Reconciler currently doesn't track this. Likely a follow-up to the reconciler — emit `needs_rerender` for pages whose list sections may have new contents.

5. **Empty-result handling.**
   - If a `tool-list` query returns zero rows (e.g. site has no tools yet), what should the section render? Empty list? Hidden? Placeholder? Probably component-controlled — schema declares `empty_state_text` as a Tier B field.

---

## Where to start when implementing

1. **Schema check**: read `page_components` table layout, see how `content_data` is structured. Confirm whether per-slot key exists.
2. **Read page-build-handler workflow** to find where `needs_section_data` is emitted, and what the parent step does after emission. Likely `plan_sections_action.go` or a sibling — see `019_tool_library.md` for action inventory.
3. **Inventory existing list components**: `tool-list`, `guide-list`, `game-list`, `news-listing`, `blog-listing`, `category-listing`, `content-listing`, `directory-listing`, `case-studies-list`, `services-grid`, `info-card-grid`, `product-grid`, `archetype-grid`, `departments-grid`, `filtered-result-grid`, `lobby-grid`, `featured-inventory`. ~17 components affected. Group by data-source flavour.
4. **Decide on the routing mechanism** (hardcoded / component-declared / LLM-routed) — see open question 2.
5. **Build a minimal handler** for the most common flavour (site-internal directory) covering 80% of cases. Defer external-feed and site-spec-data flavours.
6. **Update component-creator's prompt** so newly-created list components include the data-source declaration in their `input_schema`.

---

## Cleanup query for the 41 stuck items once the handler exists

```sql
-- Reset all stuck needs_section_data items so the new handler picks them up
UPDATE site_work_items
SET status = 'triaged',
    claimed_at = NULL,
    claimed_by = NULL,
    completed_at = NULL,
    attempt_count = 0,
    error = NULL,
    updated_at = NOW()
WHERE item_type = 'needs_section_data'
  AND status = 'needs_human_review';
```

Once the handler is consuming items from triaged, this releases the backlog.

---

## Affected sites today (snapshot 2026-05-06)

| domain | stuck items |
|--------|-------------|
| leopardessconsulting.co.uk | 17 |
| gaswholesalers.com | 6 |
| robot-hands.com | 6 |
| ai-agent-orchestration.com | 5 |
| finetuning.uk | 4 |
| vonc.com | 3 |

Plus the 6 we manually completed for gamesdesign.co.uk during the Phase 1 verification.
