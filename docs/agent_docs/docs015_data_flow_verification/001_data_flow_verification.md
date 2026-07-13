# Research Agent Data Flow Verification

## Path Verification Matrix

This document verifies that all input/output paths match between workflow config and action implementations.

### Step 1: extract_topic

| Config | Value | Verified |
|--------|-------|----------|
| action | `extract_fields` | ✓ Action exists |
| output_field | `extracted` | ✓ |
| fields.topic | `["current_section.topic", ...]` | ✓ ExtractNestedField handles arrays |

**Output stored at:** `collected_data["extracted"]`
**Structure:** `{topic: string, company: string, industry: string}`

---

### Step 2: build_search_query

| Config | Value | Verified |
|--------|-------|----------|
| action | `execute_llm_prompt` | ✓ Action exists |
| input_fields | `["extracted"]` | ✓ Maps to collected_data["extracted"] |
| prompt_template | `{{.extracted.topic}}` | ✓ Template variable matches |
| output_field | `search_query` | ✓ |

**Output stored at:** `collected_data["search_query"]`
**Structure:** `{result: string, type: "text"}`

---

### Step 3: search_web

| Config | Value | Verified |
|--------|-------|----------|
| action | `web_search` | ✓ WebSearchAction exists |
| query_from | `search_query.result` | ✓ extractSearchQuery checks this path |
| output_field | `search_results` | ✓ |

**Input path:** `collected_data["search_query"]["result"]` → string
**Output stored at:** `collected_data["search_results"]`
**Adapter returns:** `{success: bool, results: [], query: string, total: int, provider: string}`

---

### Step 4: prepare_urls

| Config | Value | Verified |
|--------|-------|----------|
| action | `prepare_urls` | ✓ PrepareUrlsAction (new) |
| (implicit) | Reads from `search_results` | ✓ findResultsArray tries search_results.results |
| max_scrapes | `3` | ✓ Config option |
| output_field | `prepared_urls` | ✓ |

**Input path:** `collected_data["search_results"]["results"]` → array
**Output stored at:** `collected_data["prepared_urls"]`
**Structure:** `{urls_to_scrape: [{url, title, index}], scrape_count: int, snippet_context: string, ...}`

---

### Step 5: scrape_pages

| Config | Value | Verified |
|--------|-------|----------|
| action | `batch_webscrape` | ✓ BatchWebscrapeAction (new) |
| urls_field | `prepared_urls.urls_to_scrape` | ✓ Exact path to array |
| output_field | `scrape_results` | ✓ |

**Input path:** `collected_data["prepared_urls"]["urls_to_scrape"]` → array of {url, title, index}
**Output stored at:** `collected_data["scrape_results"]`
**Adapter returns:** `{success: bool, results: [], success_count: int, error_count: int, total_count: int}`

---

### Step 6: format_content

| Config | Value | Verified |
|--------|-------|----------|
| action | `format_research_content` | ✓ FormatResearchContentAction (new) |
| scrape_field | `scrape_results` | ✓ Finds results at scrape_results.results |
| snippets_field | `prepared_urls.snippet_context` | ✓ Exact path |
| output_field | `research_content` | ✓ |

**Input paths:**
- `collected_data["scrape_results"]["results"]` → array of scraped pages
- `collected_data["prepared_urls"]["snippet_context"]` → string

**Output stored at:** `collected_data["research_content"]`
**Structure:** `{research_text: string, sources: [], source_count: int, content_quality: string}`

---

### Step 7: synthesize

| Config | Value | Verified |
|--------|-------|----------|
| action | `execute_llm_prompt` | ✓ |
| input_fields | `["extracted", "research_content"]` | ✓ Both exist in collected_data |
| prompt_template | `{{.extracted.topic}}`, `{{.research_content.research_text}}` | ✓ Paths match |
| output_format | `json` | ✓ Parsed JSON returned |
| output_field | `synthesis` | ✓ |

**Input paths:**
- `collected_data["extracted"]["topic"]` → string
- `collected_data["research_content"]["research_text"]` → string

**Output stored at:** `collected_data["synthesis"]`
**Structure:** `{result: {summary: string, key_points: [], recommendations: [], confidence: float}, type: "json"}`

---

### Step 8: store_research

| Config | Value | Verified |
|--------|-------|----------|
| action | `insert_research_result` | ✓ Action exists |
| fields.topic | `extracted.topic` | ✓ collected_data["extracted"]["topic"] |
| fields.query | `search_query.result` | ✓ collected_data["search_query"]["result"] |
| fields.site_id | `site_record.site_id` | ✓ From input data |
| fields.sources | `research_content.sources` | ✓ collected_data["research_content"]["sources"] |
| fields.summary | `synthesis.result.summary` | ✓ collected_data["synthesis"]["result"]["summary"] |
| fields.findings | `synthesis.result` | ✓ Full synthesis object |
| output_field | `stored_research` | ✓ |

---

### Step 9: complete

| Config | Value | Verified |
|--------|-------|----------|
| action | `complete_workflow` | ✓ |
| output.summary | `synthesis.result.summary` | ✓ |
| output.key_points | `synthesis.result.key_points` | ✓ |
| output.source_count | `research_content.source_count` | ✓ |

---

## Response Header Verification

All adapter responses include required headers:

| Header | Required For | Present |
|--------|--------------|---------|
| correlation_id | All messages | ✓ |
| message_type | All messages | ✓ (= "response") |
| orchestration_id | Responses | ✓ |
| client_id | Responses (echoed) | ✓ |
| in_response_to_request_id | Responses | ✓ |
| in_response_to_step_name | Responses | ✓ |
| status | Responses | ✓ (= "complete" or "error_recoverable") |

---

## Files Created

1. **research_actions.go** - PrepareUrlsAction, FormatResearchContentAction, findResultsArray helper
2. **batch_webscrape_action.go** - BatchWebscrapeAction
3. **webscrape_batch_handler.go** - handleBatchScrape, sendBatchSuccessResponse, sendBatchErrorResponse
4. **research_agent_workflow_update.sql** - Updated workflow with verified paths
5. **webscrape_adapter_patch.go** - Instructions for patching adapter.go

---

## Registration Checklist

### Actions to register in registry.go:
```go
"prepare_urls":            PrepareUrlsAction,
"format_research_content": FormatResearchContentAction,
"batch_webscrape":         BatchWebscrapeAction,
```

### Actions to add to local_actions.go:
```go
"prepare_urls":            true,
"format_research_content": true,
"batch_webscrape":         true,
```

### Adapter modification:
Add `case "batch_scrape":` to handleMessage switch in webscrape adapter.