# Adoption Pipeline — Handoff Document

Session date: 2026-03-31 (continuation of 2026-03-30)

---

## What Was Built

### Site Adoption Agent — end-to-end pipeline

A dedicated `site-adoption-agent` that takes over existing websites. Triggered with a domain, crawls the site, classifies structure, extracts writing style, and creates everything needed for the build pipeline.

**Workflow flow:**
```
ensure_site_record → crawl_site (Firecrawl v2)
  → format_crawl_for_analysis (Go: 500-char summaries per page for LLM)
  → check_crawl_content (conditional)
  → analyze_site (LLM 1: classify structure from summaries)
  → select_content (Go: pick 2-3 prose-heavy pages from crawl)
  → derive_content_direction (LLM 2: extract detailed writing style guide)
  → apply_plan (Go: write specs, create pages, create work items)
  → complete
```

**Two-stage processing principle:** LLM gets lightweight summaries for reasoning (classification, style analysis). Go code handles content extraction from full crawl data — no token limits, no cost, deterministic.

### Deployed Go Actions

All in `platform/orchestration/actions/`:

| Action | Registry key | Purpose |
|--------|-------------|---------|
| `ApplyAdoptionPlanAction` | `apply_adoption_plan` | Writes specs, pages, work items from LLM analysis. Uses `datahelpers.UnwrapDeep` for LLM envelope. Repairs truncated JSON. Extracts content from crawl via `buildCrawlPageIndex`. Formats content_direction via `datahelpers.FormatContentDirection`. |
| `FormatCrawlForAnalysisAction` | `format_crawl_for_analysis` | Produces ~500 char summaries per crawled page. Configurable via `summary_chars_per_page`. |
| `LoadExistingContentAction` | `load_existing_content` | Loads adoption page content from `research_results` when work item has `mode: "recreate"`. Non-blocking — returns `{has_existing: false}` on any failure. |
| `SelectRepresentativeContentAction` | `select_representative_content` | Picks homepage + 2 longest prose-heavy pages from crawl for writing style analysis. Scores pages by prose density, skips tables/UI markup. |

### Deployed Go Utility (ready but may need deploying)

In `platform/orchestration/datahelpers/`:

| File | Purpose |
|------|---------|
| `format_content_direction.go` | `FormatContentDirection()` — walks any content_direction spec structure and produces a single readable text block. Used by `apply_adoption_plan` and should be hooked into `write_site_spec`. |

### Firecrawl Fix

`internal/adapters/webscrape/providers/firecrawl.go` — the Crawl function's `formats` field was moved into `scrapeOptions` object for Firecrawl v2 API compatibility. The Scrape function was already correct.

### SQL Applied

- Site-adoption-agent definition with full workflow
- `ai_service` in analyze_site step config (shared pods need this)
- `input_fields` for analyze_site and apply_plan steps
- `max_tokens: 32000` for analyze_site
- Format step using `format_crawl_for_analysis` with `summary_chars_per_page: 500`
- Lighter LLM prompt (structure only, no existing_content)
- `scrape_config` with `formats: ["markdown", "rawHtml"]` and `only_main_content: false`

### Trigger Script

`trigger-adopt-site.sh` — Usage: `./trigger-adopt-site.sh <domain> [url]`

---

## What Was Tested

| Site | Specs | Pages | Items | Status |
|------|-------|-------|-------|--------|
| gamedesign.uk | 3 (identity, design, structure) | 10 | 12 | Adopted successfully. Pages deleted during testing. Needs re-adoption. |
| robot-hands.com | 3 (identity, design, structure) | 2 | 4 | Adopted successfully. Data intact. |

The LLM correctly identified: company names, industries, page types (landing, tool, tool-index, blog-index, blog-post, game-list), section types, interactive features (calculators, simulators, games), and design characteristics.

---

## Files Ready to Deploy (not yet deployed)

### Go files in `/mnt/user-data/outputs/`:

| File | Destination | Status |
|------|-------------|--------|
| `apply_adoption_plan_action.go` | `platform/orchestration/actions/` | Latest version — uses `datahelpers.FormatContentDirection`, generic interactive features, rawHtml storage |
| `select_representative_content_action.go` | `platform/orchestration/actions/` + registry | New action |
| `format_content_direction.go` | `platform/orchestration/datahelpers/` | New utility |
| `load_existing_content_action.go` | `platform/orchestration/actions/` + registry | May be deployed already — check |
| `write_site_spec_patch.go` | Patch to `platform/orchestration/actions/site_spec_actions.go` | 5-line addition: auto-formats content_direction when `write_site_spec` runs with `aspect == "content_direction"` |

### SQL files in `/mnt/user-data/outputs/`:

| File | Purpose | Status |
|------|---------|--------|
| `add_content_direction_to_adoption.sql` | Adds select_content + derive_content_direction steps to adoption workflow. Uses `$prompt$` dollar-quoting. | Ready to run |
| `update_classifier_content_direction.sql` | Enriches domain-research-classifier prompt for richer content_direction output. Bumps max_tokens to 6000. Updates model to claude-sonnet-4-6. | Ready to run |
| `update_content_writer_one_field.sql` | Replaces 4 hardcoded content_direction template vars with 1 `formatted` field. | Ready to run |
| `wire_existing_content.sql` | Adds load_existing_content step to page-build-handler. Passes existing_content? and build_mode? to content writer. | May be applied already — check |

### Model update (run after all SQL):

```sql
UPDATE agent_definitions
SET default_config = (replace(default_config::text, 'claude-sonnet-4-5', 'claude-sonnet-4-6'))::jsonb
WHERE default_config::text LIKE '%claude-sonnet-4-5%'
  AND deleted_at IS NULL;
```

---

## Key Architecture Decisions

### Adoption is one-off data capture
Crawl data stored in `research_results` (transient reference). Only clean, forward-looking data goes in `site_specs`. The `adopted_from` URL in the identity spec is the only permanent trace.

### Two-stage processing
LLM for reasoning (classification, style analysis). Go for extraction (content, rawHtml, URL matching). Never send full page content through an LLM just to get it back.

### Generic interactive features
No separate `needs_tool_page` item type. All pages use `needs_content_page`. Pages with interactive features carry them in the work item spec under `interactive_features`. rawHtml stored for all pages when available.

### Content direction as formatted text
The content_direction spec stores both structured data (for programmatic access) and a `formatted` string field (for LLM prompts). `FormatContentDirection()` in datahelpers walks whatever structure the LLM produced and generates readable text. The content writer reads one template variable: `{{.site_specs.specs.content_direction.formatted}}`. No template changes needed regardless of how the spec evolves.

### Overwrite protection for owned domains
When adopting a domain you host (gamedesign.uk, robot-hands.com), the crawl data in `research_results` is the canonical source. Every rebuild reads from there, never re-crawls. If a partial build overwrites the live site, the original content is preserved in the DB.

### Component discovery (not yet built)
When plan_sections encounters a section type with no matching component template, the system should create a `needs_new_component` work item rather than forcing content into existing templates. The component library grows through adoption. This is being worked on in a separate chat.

---

## Debugging History — Issues Found and Fixed

1. **Firecrawl v2 "Unrecognized key"** — `formats` at top level of /crawl payload. Fixed: wrapped in `scrapeOptions`.
2. **format_research_content couldn't parse crawl** — Expected batch format, got crawl format. Fixed: created `format_crawl_for_analysis`.
3. **ai_service not found** — Shared pods need `ai_service` in step config, not top-level. Fixed.
4. **Wrong model name** — Used versioned string instead of alias. Fixed to `claude-sonnet-4-6`.
5. **LLM returned all "Unknown"** — `input_fields` missing from analyze_site step. Template data was empty. Fixed.
6. **LLM result wrapped in envelope** — `{"type":"text","result":"..."}`. Fixed: uses `datahelpers.UnwrapDeep`.
7. **JSON truncated by max_tokens** — 10 pages + features exceeded token limit. Fixed: bumped to 32000, added `repairTruncatedJSON`, made prompt lighter.
8. **Single quotes in SQL prompt** — Broke jsonb_set string. Fixed: use `$prompt$` dollar-quoting in DO blocks.

---

## Pending Work — What Needs Doing Next

### Immediate (get adoption end-to-end):

1. **Deploy the Go files** listed above (format_content_direction.go, select_representative_content, updated apply_adoption_plan, write_site_spec patch)
2. **Run the SQL files** (add_content_direction, update_classifier, update_content_writer_one_field, model update)
3. **Re-adopt gamedesign.uk** with the content direction pipeline
4. **Verify content_direction spec** has structured data + formatted text
5. **Enable dispatch loop** — let it process the 12 work items
6. **Handle missing component templates** — plan_sections currently passes unknown sections through as "ready" with no component. The content writer tries to generate content but render_component has no template. Need the component selection system (being built in another chat) or a fallback path.

### Short-term:

7. **Content writer recreate mode** — wire_existing_content.sql may need re-checking. The content writer prompt has an existing_content conditional block (already injected). The load_existing_content step needs to be in page-build-handler.
8. **Interactive element builder** — degradation loop: try direct recreation → described recreation → simplified equivalent → content fallback. Uses content-reviewer's eval pattern. rawHtml is in research_results.
9. **Try mortgagecalculator.co.uk** — the real test case for compliance-heavy content direction.
10. **Robot-hands.com build** — already adopted, 4 work items queued.

### Medium-term:

11. **Component discovery** — `needs_new_component` work items when plan_sections has no matching template. Component-creator handler generates template from reference content.
12. **Component selector** — section_type metadata, scoring, quality feedback.
13. **Lock types** — `lock_type` and `lock_expires_at` on page_components and site_components.
14. **Dashboard direction panel** — post-build HITL.

---

## Key Files in the Project Knowledge

| File | What it covers |
|------|---------------|
| `026_infrastructure_layers_and_adoption.md` | Full architecture: 3 infra layers, 5 backend tiers, adoption pipeline, HITL direction, component discovery |
| `009d_improvement_loop_v4.md` | Improvement loop with adopted site handling, lock types, audit reset |
| `001g_development_guide_new_agents_v7.md` | Dev guidelines — check this before writing any code |
| `015_consolidated_site_spec_classifier_architecture.md` | Who writes which specs, content strategy framework |

---

## Kubernetes Reference

```bash
kubectl -n ai-persona-system get pods       # main namespace
kubectl -n kafka get pods                    # kafka cluster
kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50
```

Deployment: git push → GitHub Actions → S3.

---

## Important Patterns to Remember

- `input_fields` in step config is only for `execute_llm_prompt` — Go actions read from `params.CollectedData` directly
- `ai_service` for shared-pod agents goes in step config, not top-level default_config
- Model aliases: `claude-sonnet-4-6` (standard), `claude-haiku-4-5` (orchestration/routing only)
- `?` suffix on input_mapping fields that are item-type-specific
- `pipeline` column on work items: `build`, `maintenance`, `marketing`
- `error_step` for graceful degradation — non-blocking steps should fall through
- Don't use `logger.Debug` — won't show in logs
- Use `$prompt$` dollar-quoting for SQL with single quotes in prompts
- Every `execute_llm_prompt` needs `api_key_env_var: "ANTHROPIC_API_KEY"` in ai_service
