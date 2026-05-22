# HANDOFF — Tier D directory-builder progress + page-content-writer fabrication finding

## Date: 2026-05-11
## Session focus: directory-builder Tier D continuation, validation observability, end-to-end re-adoption of gamesdesign.co.uk
## Status: paused mid-investigation of a content fabrication path

---

## Headline findings

**1. tool-list and guide-list are both regenerated to Tier D shape**

- tool-list: hand-written via migration 041 in an earlier session (canonical Tier D reference). Schema declares 8 top-level fields including `card_link_label`, with `items` (array, source=`query.pages_where_type:tool`) and 4 sub-schema fields.
- guide-list: LLM-regenerated this session via migration 042 + work item dispatch. Attempt 1 failed validation (orphan `read_guide_label`), attempt 2 succeeded. Stored at 18:40:51 with quality_score=100, schema_template_synced=true, items source = `query.pages_where_type:guide`.

The retry-and-converge pattern held for guide-list (same as tool-list earlier this session). Two for two on the LLM producing structurally-good output that needs one retry to fix a single bookkeeping orphan.

**2. The new validation observability is working as designed**

The agent_error_log writer deployed this session captured guide-list's attempt-1 failure exactly as intended:
- severity = `warning`
- error_code = `component_validation_orphan_schema_field`
- context.orphan_schema_fields = `["read_guide_label"]`
- context.unknown_template_vars = `[]`
- template_variable_count = 11, schema_field_count = 10
- function = guide-list

In one SQL row we get the entire shape of the failure. No kubectl log forensics required.

**3. The directory-builder end-to-end verification on gamesdesign.co.uk needs revisiting**

Earlier this session the deployed tools.html showed 6 tool cards with real titles/URLs from the pages table — looked like queryresolve had populated content_data. On guide-list regen, we expected the same path to deliver real (1 guide) entries.

What we found instead: gamesdesign's index page has a guide-list section with **6 fabricated guide entries**, not from queryresolve and not corresponding to any rows in the pages table. These came from the `page-content-writer` agent at 18:48:29, AFTER guide-list regen. The page-content-writer LLM was given a prompt like "Write content for the custom section of GamesDesign.co.uk" and produced the full items array, fabricating 6 plausible-looking guide entries with fake URLs (`/guides/pseudo-random-distribution.html`, etc.). None of these URLs match real pages.

Equivalent llm_call_log entries show page-content-writer also ran for **tool-list on the tools page** at 15:39:26 — producing the 6 tool entries that we celebrated as "queryresolve working end-to-end." Those tools were also LLM-fabricated. They happened to match real pages because the LLM had context from the site adoption that included the real tools, but they're not the result of queryresolve hitting pages_where_type:tool.

**This means the directory-builder verification we recorded earlier in the session was misread.** Queryresolve may or may not be wired into the actual content path. page-content-writer is running independently, fabricating content for sections, and overwriting (or being asked to fill) array fields like `items` with LLM-generated lists.

This is now the next investigation. Architecturally the resolver was supposed to forbid fabrication. page-content-writer is silently doing it.

---

## What got deployed this session

### Validation observability (rejection logger)
- `store_generated_component_action.go` — added `recordValidationRejection` helper at end of file. On every pre-store validation rejection, writes a structured row to `agent_error_log` with parsed orphan/unknown field names, severity classification (warning for bookkeeping mismatches, error for structural), and full context as JSONB.
- Two regex constants: `orphanSchemaFieldPattern`, `unknownTemplateVarPattern`.
- Removed the temporary `DEBUG_PRESTORE_SCORE` log line.
- Build deployed and confirmed in binary via `strings | grep recordValidationRejection`.

### Documentation
- `026_component_regeneration_flow.md` extended from 396 to 596 lines. New section "Pre-store validation, error logging, and LLM reliability" covers: validation gate, parse-tree extractor, agent_error_log records with example queries, why we don't auto-correct, three-track decomposition strategy (move bookkeeping out of LLM), LLM model choice rationale, success criteria, operational notes, file locations, migration history table (038-041).
- `FOCUS_llm_reliability_for_component_generation.md` — standalone strategy doc with three tracks (rejection observability done, root-section wrapper injection next, derived schema fields longer-term).
- Migration 041 (already applied in earlier session): tool-list hand-written.
- Migration 042 (applied this session): guide-list regen work item + backup table.

---

## End-to-end re-adoption walkthrough

User process: delete `sites/gamesdesign.co.uk/` directory in github repo, delete the backblaze bucket, run cleanup SQL deleting from sites/style_collections/css_themes/palettes/typography_sets/layouts/etc. by source_domain, then trigger separated adoption via kcat publishing to `system.agent.generic.requests` with action=orchestrate, agent_type=site-adoption-orchestrator, target_url=https://gamedesign.uk, destination_domain=gamesdesign.co.uk.

Run identifiers: CORRELATION_ID=`3be334ae-288d-48fa-b7b5-f67832c31336`, ORCHESTRATION_ID=`f4446ec6-3465-4a4a-a930-a000ae2f469c`. Adoption finished in ~5 minutes.

Site row: gamesdesign.co.uk site_id=`859d7ad5-0f22-4ba1-8efd-cd59e8fb042f`.

Pages created (11 in the first batch at 15:16, 5 more at 16:31:50):
- index (page_type=index)
- tools (page_type=content)
- guides-index (page_type=blog_index)
- games (page_type=content)
- 6 tool pages (page_type=tool)
- guide-rng-design (page_type=blog_post)
- tools-index (page_type=entity_directory, 16:31)
- guide-template, about, tool-lanchester-combat-calculator, tool-loot-probability-calculator (16:31 batch)

Critical: **NO pages have page_type='guide'**. Guides are blog_post. The LLM-regenerated guide-list queries `pages_where_type:guide` — wrong query name for this site. Whether this matters depends on whether queryresolve actually fires or whether page-content-writer overrides.

Build pipeline: dispatched in priority order. The `tools` page got deployed at 15:39 with tool-list section containing the 6 LLM-fabricated tool cards. index, games, guides-index, and three tool detail pages reached `deployed`. Several others still `planned`.

Initial guide-list block: a `needs_section_data` work item with `status=needs_human_review` fired for the index page demanding `guide_2_url` through `guide_6_url` from `site_specs.guides.guide_N_url` — the pre-Tier-D shape. This work item was the gating reason guide-list couldn't resolve via the OLD shape. Migration 042 worked around this by regenerating the component, not by resolving the work item.

After guide-list regen (18:40:51): the regen action emitted a follow-up `needs_rerender` work item (`4f2fa6bf-...`). The dispatch loop CLAIMED it (later than we initially saw). The index page page_components row for guide-list was created at 18:51:14 with 6 LLM-fabricated guide entries (different topics from gamedesign.uk's actual guides, which are p2p-architecture, skinner-box, rng-design, fairness-in-rng, economy-basics).

---

## The page-content-writer fabrication path

This is the open question. Three observations from llm_call_log:

**a) tool-list on tools page (15:39:26):**
- Agent: page-content-writer
- Step: `process_sections_loop_iter_1_generate_content`
- Prompt: "Write content for the custom section of Tools & Calculators - GameDesign.uk."
- Response: full JSON including eyebrow_label, section_heading, section_intro, items array with 6 tool entries (PRD Calculator, XP Curve Builder, TTK Calculator, Drop Rate Simulator, Economy Faucet & Sink Balancer, Probability Distribution Visualiser).
- The 6 tools "happen to" match what the actual `tool` page_type rows contain. They're plausible content for the gamedesign source.

**b) Same agent, earlier same day (15:21:52):**
- Step: `process_sections_loop_iter_1_generate_content`
- Prompt: "Write content for the custom section of GameDesign.uk | The Utility Engine for Game Developers."
- Response: tool-list shape but with `"items": []` and `"card_link_label": "Launch tool"`.

**c) guide-list on index (18:48:29):**
- Same agent, same step naming pattern, iter_2.
- Prompt: "Write content for the custom section of GamesDesign.co.uk — Design with Authority."
- Response: full guide-list JSON with 6 fabricated guides.

**Pattern:** page-content-writer is invoked per section per page during the build flow, given a prompt with company context + section type, and produces JSON content for the entire section including the items array. It does NOT consult queryresolve. It does NOT consult the pages table. It generates from LLM context.

**Open question 1:** Does plan_sections's queryresolve invocation ever happen, OR does page-content-writer always run instead? Or do BOTH run with one overwriting the other?

**Open question 2:** Where in the code is page-content-writer wired into the build flow? Looking at the workflow steps `process_sections_loop_iter_N_generate_content` suggests it's called from a loop driver in the page-build-handler workflow. Need to find this loop and understand what it does for `items`-typed fields.

**Open question 3:** Why did tool-list get plausible items while guide-list got fabricated ones? Possible reasons:
- The adoption captured the gamedesign.uk source text including tool names in some site spec — page-content-writer's LLM had access to "PRD Calculator", "TTK Calculator" etc. via context.
- The adoption captured limited guide info, so the LLM invented plausible guide topics that match the technical voice.
- Either way, the LLM-fabricated tools happening to match real pages is coincidence, not architecture.

**Architectural intent:** queryresolve was built specifically to remove fabrication from items fields. If page-content-writer is running AFTER plan_sections and overwriting query.* fields with LLM output, the architecture is being subverted. If page-content-writer is supposed to run only for non-query.* fields (eyebrow_label, section_heading, etc.) but accidentally runs for items too, that's a bug in the loop driver.

---

## What to do next

### Immediately

Find where page-content-writer is invoked and understand its scope:

```sql
-- All page-content-writer calls in this session, grouped by step name
SELECT 
    DISTINCT step_name, 
    COUNT(*) AS call_count,
    MIN(created_at) AS first_called,
    MAX(created_at) AS last_called
FROM llm_call_log
WHERE agent_type = 'page-content-writer'
  AND created_at > '2026-05-11 14:00:00'
GROUP BY step_name
ORDER BY first_called;
```

```bash
# Find the workflow definition that invokes page-content-writer
grep -rn "page-content-writer\|page_content_writer" /mnt/project/production_agent-chassis-full_context.txt | head -20
```

Then inspect the agent_definitions row:

```sql
SELECT name, prompt_template, workflow
FROM agent_definitions
WHERE name = 'page-content-writer';
```

The prompt_template will tell us EXACTLY what fields it's asked to produce, including whether it's instructed about items arrays.

### Then

Three possible outcomes once we understand the wiring:

1. **page-content-writer is supposed to skip query.* fields but doesn't** — fix the loop driver to detect query.* source on a field and skip LLM generation for that field. Queryresolve becomes the sole producer of items arrays.

2. **page-content-writer's prompt is too broad — it asks for the full section JSON regardless of source.** Fix the prompt to instruct it to leave `items: []` empty and let queryresolve fill it. Risk: LLMs ignore "leave blank" instructions.

3. **page-content-writer is the intended provider for items in all cases and queryresolve was never wired in.** Then the directory-builder slice was never functional and we need to wire queryresolve into the actual code path. Worst-case scenario.

Whichever it is, this is the blocker before we proceed with broadening Tier D to game-list / blog-listing. There's no point regenerating more components if the items they're supposed to query are silently being fabricated by another agent.

### Deferred (carried forward from earlier in session)

- The `error` column on `site_work_items` isn't cleared when a retry succeeds. Minor cosmetic — the row reports status=complete with attempt_count=1 and an error string from attempt 1's rejection. Confused us briefly. Should set `SET error=NULL` on success transition.
- Cosmetic: separated-adoption title interpolation uses source domain ("GameDesign.uk") rather than destination domain ("GamesDesign.co.uk"). Minor.
- Dead BEM CSS / theme variables on gamesdesign (D-priority).
- 41 stuck `needs_section_data` items across 6 other sites — most are "component not found" cases that should have been `needs_component:*`.

---

## Key identifiers (for next session)

- Component-creator agent_definitions id: `23720180-7a39-4e3d-92e1-ebdbf95b57f4`
- system.internal site_id: `eac60db8-b032-432b-b36d-76f37632045d`
- gamesdesign.co.uk site_id: `859d7ad5-0f22-4ba1-8efd-cd59e8fb042f`
- tool-list component_id: `a68b52b7-61c5-4797-a701-8e8643684f75` (hand-written, migration 041)
- guide-list component_id: `9d5e461a-8981-4ecc-b236-05895edfc15d` (LLM-regenerated, migration 042 + attempt 2)
- Index page_id on gamesdesign: `43744ce2-5731-4a9b-95de-7cfbd6b3d866`
- Tools page_id on gamesdesign: `1e5dfbe0-c136-4742-9f4d-4ba12bfb22d4`
- guide-list rerender work_item_id (was claimed and processed): `4f2fa6bf-acf3-4de8-96a4-aae30c5b40ad`

---

## Files staged in /mnt/user-data/outputs/ this session

Active:
- `042_regen_guide_list_tier_d.sql` (APPLIED — guide-list regen work item + backup)
- `store_generated_component_action.go` + `.diff` (DEPLOYED — rejection logger)
- `026_component_regeneration_flow.md` (amended, 596 lines, references v8 of contracts doc)
- `FOCUS_llm_reliability_for_component_generation.md` (strategy doc)

Historical from earlier sessions still relevant:
- `041_handwrite_tool_list_tier_d.sql` (hand-written tool-list reference)
- `compute_component_quality.go` + `.diff` (parse-tree validator, deployed earlier)
- `queryresolve/queryresolve.go` — the resolver we thought was being used; status uncertain after this session's finding

---

## Operational lessons reinforced this session

1. **Validation observability landed cleanly.** Two real rejections this session (one from attempt-1 of guide-list, and an earlier one we saw cleared by retry success). Both produced clean structured records in agent_error_log. Pattern detection across sessions is now SQL-queryable. The work item state-machine has a cosmetic bug (error column not cleared on retry success) but that doesn't impair the error_log records.

2. **The retry-and-converge pattern holds on Tier B label fields.** tool-list missed `card_link_label`; guide-list missed `read_guide_label`. Both passed on attempt 2. The retry budget of 3 is calibrated correctly for this class of failure. Decomposition (moving label fields to central registry) would prevent the failures entirely; until then, retry handles them.

3. **End-to-end verification needs ground-truth checks beyond "did the SQL row update."** The deployed tools.html LOOKED right because the items happened to match real tool pages. They were actually LLM-fabricated content that coincided with real data. Looking at content_data values without confirming where they came from misses this. Future verifications should confirm: (a) the resolver fired, (b) the values match the pages table, (c) no other agent overwrote between resolution and render.

4. **page-content-writer is a major surface we hadn't been tracking.** It's been running this whole time, generating section content per section per page. The directory-builder design assumed queryresolve was the path. There's a gap between design and implementation here that we need to close.
