# Handoff: Tool Pipeline — End-to-End Working + Cross-Linking

## Context

We're building a system that intelligently plans and builds multipage websites given domain names. Sites get interactive tools (calculators, converters, quizzes etc.) suggested and deployed automatically by an agent pipeline. This session took the tool pipeline from "novel tools fail at tool-deployer" to "fully autonomous end-to-end with novel generation, content, companion guides, nav entries, and cross-linking planned."

Prior handoff: `/mnt/transcripts/` or project knowledge `032_tool_handoff_novel_tool_generation.md`

Key docs (updated this session — v3 versions are current):
- `012b_tool_lifecycle_guide_v3.md` — full pipeline, all agents, work items, quality tiers
- `010b_tool_library_guide_v3.md` — library inventory, deployment, templates, both fork and generation paths
- `001g_development_guide_new_agents_v8.md` — agent creation guidelines, updated bug history

---

## What's Deployed and Working

### SQL changes (live in database)

1. **070_tool_suggester_routing.sql** — tool-suggester's `create_items_loop` now has a `check_is_library` conditional:
   - `tool_component_id != null` → `handler_agent: tool-deployer` (library fork)
   - `tool_component_id == null` → `handler_agent: tool-generator` (novel LLM generation)
   - Verified: `SELECT default_config->'workflow'->'steps'->'create_items_loop'->'config'->'sub_workflow'->'start_step' FROM agent_definitions WHERE type = 'tool-suggester'` returns `"check_is_library"`

2. **content_components check constraint** — temporarily widened to include `'tool-generator'`:
   ```sql
   CHECK (created_from IS NULL OR created_from IN ('manual', 'generated', 'adopted', 'tool', 'forked', 'tool-generator'))
   ```
   Once the Go code with `'generated'` deploys, tighten back to remove `'tool-generator'`.

3. **Password checkers deactivated** — all `tool-password-entropy` components set `is_active = false` (library original + all forks). This unblocked tool evaluation for the 3 deployed sites.

### Go changes (deployed)

1. **`create_tool_component_action.go`** — major update:
   - Strips markdown fences via `datahelpers.StripCodeFences(htmlContent)`
   - Sets `created_from = 'generated'` (was `'tool-generator'` which hit constraint — **but the currently running code may still have 'tool-generator'**, hence the constraint workaround above)
   - Page component at position 2 with `slot_name`, `rendered_html`, `build_status = 'deployed'`
   - Creates `needs_content_page` work item for tool page (hero/intro/CTA)
   - Creates companion guide page at `/guides/{function}-guide.html` + work item
   - Creates nav entry via fixed `addToolToNav`

2. **`addToolToNav`** (in create_tool_component_action.go) — fixed wrong column names:
   - `site_nav_groups`: uses `group_key`, `group_label`, `group_type`, `position` (was using non-existent `label`, `sort_order`, `in_header`, `in_footer`)
   - `site_nav_items`: uses `site_id`, `group_id`, `label`, `url`, `page_id`, `item_type`, `position`, `status` with `NOT EXISTS` guard (was using `sort_order`, missing `site_id`)

3. **`content_search.go`** — extracted `StripCodeFences()` as exported function from `tryParseJSON`. `tryParseJSON` now calls it.

4. **`deep_search.go`** — replaced inline fence stripping in Pattern 1 and Pattern 2 with `StripCodeFences()` calls.

5. **`check_missing_tools.go`** — removed the `deployedToolCount > 0 → skip` early return. Now uses tiered cooldowns:
   - 0 tools: 7-day cooldown (needs tools urgently)
   - 1+ tools: 30-day cooldown (periodic re-evaluation as content grows)
   - Tool-suggester already loads existing tools and avoids duplicates, so re-evaluation is safe.

### Verified end-to-end (tested with gaswholesalers.com)

Full pipeline proven:
```
missing_tools check → evaluate_tools → tool-suggester (LLM evaluation + routing)
→ tool-generator (LLM generates HTML) → create_tool_component (component, page,
  page_component with rendered_html, nav entry, content work items, companion guide)
→ page-build-handler (writes hero/intro/CTA + guide content)
→ rerender-pages → page-rerender → deployed to live site
```

Results on gaswholesalers.com:
- **Fuel Cost Estimator** — created with old Go code, missing rendered_html/nav/guide (manually patched)
- **Gas Unit Converter** — created with new Go code, all pieces present:
  - Component: 8,682 chars, `created_from = 'tool-generator'`
  - Page component: position 2, slot_name set, rendered_html, build_status = deployed
  - Tool page: `/tools/tool-gas-unit-converter.html` — deployed
  - Companion guide: `/guides/tool-gas-unit-converter-guide.html` — deployed
  - Nav entry: Tools → Gas Unit Converter
  - Content work items: both complete
  - Live: https://gaswholesalers.com/tools/tool-gas-unit-converter.html

### Autonomous operation verified

After deploying `check_missing_tools.go` (tiered cooldowns) and deactivating password checkers, the system automatically:
- Evaluated all 3 deployed sites (ai-agent-orchestration.com, finetuning.uk, leopardessconsulting.co.uk)
- Suggested appropriate tools per site based on industry/audience
- Generated and deployed them without manual intervention

Results:
- **ai-agent-orchestration.com**: AI Agent ROI Estimator (novel, deployed)
- **finetuning.uk**: AI Readiness Quiz (novel) + AI Time Savings Estimator (library fork) — both deployed
- **leopardessconsulting.co.uk**: evaluation was triaged at session end, should complete on next dispatch cycle

---

## What's Written But Not Yet Deployed

### 071: Tool cross-linking (3 files)

**Problem:** Tools get created with nav entries and companion guides, but existing content pages don't reference them. A services page should say "Use our ROI estimator to see what this would save you" — but nothing creates those references.

**Solution:** Tool-suggester's LLM already sees all pages and understands the site. Add `related_pages` to its output schema, then create `content_rewrite` work items for each related page.

Files:

1. **`create_tool_cross_link_items.go`** — Go action that reads `evaluation.result.suggestions`, iterates over each suggestion's `related_pages` array, creates `content_rewrite` work items for `page-build-handler`. Each item has specific guidance to integrate inline, not add a new section. Dedup via `item_key = tool_crosslink:{function}:{page}:{site}`. Priority 110.

2. **`071_tool_cross_linking.sql`** — Two SQL updates to tool-suggester:
   - Adds `related_pages` field to the LLM output JSON schema + instruction 6 telling the LLM to pick 1-3 pages whose topic connects to the tool
   - Adds `create_cross_links` step after `create_items_loop` → calls the Go action → then `complete`
   - Uses `regexp_replace` on the prompt — **verify the regex actually matches** after applying, the prompt text has heavy escaping

3. **`registry_addition.go.snippet`** — Registry entry for `create_tool_cross_link_items` action in `registry.go`

**Risk with 071:** The SQL migration uses `regexp_replace` on the prompt string which has quadruple-escaped JSON. Test with:
```sql
-- After applying, verify:
SELECT default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template' LIKE '%related_pages%' as has_related_pages,
       default_config->'workflow'->'steps'->'create_cross_links'->>'action' as cross_links_action,
       default_config->'workflow'->'steps'->'create_items_loop'->>'next_step' as items_loop_next
FROM agent_definitions WHERE type = 'tool-suggester';
-- Should return: true, 'create_tool_cross_link_items', 'create_cross_links'
```

If the regex doesn't match, extract the full prompt, edit it manually, and `jsonb_set` the whole thing back. The key additions to the prompt are:
- In the JSON schema, after `"complexity": "simple"`, add `"related_pages": ["page-name-1", "page-name-2"]`
- Before the rejected_tools instruction, add: "For each suggestion, include related_pages: a list of 1-3 existing page names where a contextual reference would help visitors"

---

## Database Schema Reminders

```
agent_definitions: type (not name), default_config (jsonb)
site_nav_groups: id, site_id, group_key, group_label, group_type, position
  UNIQUE: (site_id, group_key)
site_nav_items: id, site_id, group_id, label, url, page_id, item_type, position, status
  NO unique constraint (use NOT EXISTS for dedup)
content_components.created_from: CHECK allows manual, generated, adopted, tool, forked
  (+ tool-generator temporarily until Go deploys with 'generated')
```

---

## Cleanup Items

1. **Tighten created_from constraint** once Go code with `'generated'` is confirmed deployed:
   ```sql
   ALTER TABLE content_components DROP CONSTRAINT chk_created_from_valid;
   ALTER TABLE content_components ADD CONSTRAINT chk_created_from_valid 
     CHECK (created_from IS NULL OR created_from IN ('manual', 'generated', 'adopted', 'tool', 'forked'));
   ```

2. **Fuel Cost Estimator on gaswholesalers** — created with old code, manually patched (rendered_html, position, slot_name). Missing: companion guide, content work items, proper nav entry. Could either leave it (tool_health audit will improve it) or delete and let tool-suggester recreate it on the next 30-day evaluation.

3. **Leopardess tool evaluation** — was at `triaged` at session end. Check:
   ```sql
   SELECT wi.item_type, wi.summary, wi.handler_agent, wi.status
   FROM site_work_items wi
   WHERE wi.site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
     AND wi.item_type IN ('evaluate_tools', 'add_tool')
   ORDER BY wi.created_at DESC;
   ```

4. **Orphaned tool pages** — gaswholesalers has `tool-fuel-cost-estimator` page with only 1 page_component (no hero/CTA sections). The guide page for gas-unit-converter is deployed but only linked from the guide itself, not from the tool page. Cross-linking (071) addresses this for future tools.

---

## What to Test in Next Session

1. **Deploy 071 (cross-linking)** — apply SQL migration, deploy Go, verify:
   - Trigger evaluation for a test site
   - Check that `content_rewrite` items appear with `source = 'tool-suggester'`
   - Check that page-build-handler adds inline tool references

2. **Verify the prompt regex worked** — the `regexp_replace` in 071 operates on heavily escaped text. If it didn't match, the prompt won't have `related_pages` and the Go action will find empty arrays.

3. **Check all deployed sites have tools** — run the verification query:
   ```sql
   SELECT s.domain, COUNT(cc.id) as tool_count
   FROM sites s
   LEFT JOIN pages p ON p.site_id = s.id AND p.status = 'active'
   LEFT JOIN page_components pc ON pc.page_id = p.id
   LEFT JOIN content_components cc ON pc.component_id = cc.id 
     AND cc.component_level = 'tool' AND cc.is_active = true
   WHERE s.status = 'deployed'
   GROUP BY s.domain;
   ```

4. **Check gaswholesalers tools are live and working** — visit:
   - https://gaswholesalers.com/tools/tool-gas-unit-converter.html
   - https://gaswholesalers.com/tools/tool-fuel-cost-estimator.html
   - https://gaswholesalers.com/guides/tool-gas-unit-converter-guide.html

---

## Files Produced This Session

| File | Location | Status |
|------|----------|--------|
| `070_tool_suggester_routing.sql` | Applied to DB | ✅ Live |
| `create_tool_component_action.go` | Deployed | ✅ Live |
| `content_search.go` (StripCodeFences) | Deployed | ✅ Live |
| `deep_search.go` (StripCodeFences) | Deployed | ✅ Live |
| `check_missing_tools.go` (tiered cooldowns) | Deployed | ✅ Live |
| `012b_tool_lifecycle_guide_v3.md` | In outputs | 📝 Doc update |
| `010b_tool_library_guide_v3.md` | In outputs | 📝 Doc update |
| `001g_development_guide_new_agents_v8.md` | In outputs | 📝 Doc update |
| `071_tool_cross_linking.sql` | In outputs | ⏳ Not yet applied |
| `create_tool_cross_link_items.go` | In outputs | ⏳ Not yet deployed |
| `registry_addition.go.snippet` | In outputs | ⏳ Not yet deployed |

---

## Architecture Reminders

- Every agent is an orchestrator
- Workflows in SQL, complexity in Go
- Don't create sub-workflows — spawn sub-agents
- Agents respond to parent's response topic
- `logger.Info` not `logger.Debug`
- Check schemas before SQL
- `agent_definitions.type` not `name`
- Kubernetes: `-n ai-persona-system` for pods, `-n kafka` for Kafka
