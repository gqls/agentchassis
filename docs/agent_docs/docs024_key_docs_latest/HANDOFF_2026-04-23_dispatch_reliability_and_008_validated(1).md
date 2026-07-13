# HANDOFF — 2026-04-23 — gamesdesign.co.uk cascade validation + dispatch reliability findings

## Session result in one line

Migration 008 (dynamic taxonomy in classifier) validated end-to-end on a natural adoption of gamesdesign.co.uk. `tool-portal-dark` selected via `library_match`, dark palette and Segoe UI typography preserved from adopted gamedesign.uk. The classifier → strategy → briefing → site-planner → composition chain now works as intended.

## What was deployed today (in main)

| File | Change | Status |
|---|---|---|
| `platform/orchestration/actions/work_items_common.go` | NEW. Centralised `workItemTerminalStatuses` const + `sqlInList()` helper. The single source of truth for what counts as "terminal" so idx_swi_dedup and ON CONFLICT WHERE clauses can't drift apart. | DEPLOYED |
| `platform/orchestration/actions/load_work_item_actions.go` | Fix A: ON CONFLICT predicate uses `sqlInList(workItemTerminalStatuses)` instead of inline string. Fix C: `queryPagesForBuild` widened from `[]string{"planned"}` to `[]string{"planned","needs_rebuild"}`, and removed early-return-on-zero-pages so site-level inserts (composition/design/rerender) always run. | DEPLOYED |
| Migration 007 — seed `tool-portal-dark` + `social-lobby` layouts with proper tags | | APPLIED |
| Migration 008 — classifier dynamic taxonomy | | APPLIED |
| `read_layout_taxonomy` action + registry entry | Returns layouts.{categories, industry_tags, layout_count, captured_at} so the classifier prompt can know what taxonomy it's classifying into | DEPLOYED |

## What was NOT deployed (still in /mnt/user-data/outputs/, ready to land)

| File | Purpose |
|---|---|
| `apply_adoption_plan_action.go` | Adoption→classifier handoff: replaces direct needs_composition+needs_design queue from adoption with a single needs_domain_research item. Reduced 1006→942 lines. Doc 028's "classifier is the strategic brain" change. |
| `007_adoption_pipeline_v4.md` | Doc aligned with classifier-first routing |
| `012_idx_swi_dedup_include_unresolved.sql` | Re-applies migration 010 predicate (adds `unresolved` to idx_swi_dedup exclusion list). Pair with the deployed Go const. Apply AFTER Go has been live for a bit. |

## What was validated (ran end-to-end successfully on a clean cascade)

Triggered fresh adoption of `gamedesign.uk` → `gamesdesign.co.uk` at 16:38. Cascade ran (with manual dispatcher pokes — see "Real bugs surfaced" below) through:

1. site-adoption-agent: crawl → fingerprint → fetch CSS → enrich → analyze → classify_archetype → derive_content_direction → apply_adoption_plan → generate_design_intent → write design_intent
2. domain-research-classifier (with dynamic taxonomy from migration 008): emitted `category=interactive`, `industry_tags=["interactive-platform","tool-portal-dark","game-design","developer-tools","systems-design","browser-tools","calculator-platform","game-dev-education"]`, confidence 0.97
3. domain-strategist: 10317-char strategy spec written
4. build-briefing-agent: 2701-char briefing spec written
5. build-site-planner: 9821-char site_plan spec written, write_build_items queued 18 items (composition, design, logo, hero, 9 content pages, rerender, 3 capability_gap)
6. site-design-planner: layout matcher picked `tool-portal-dark` (4-tag overlap, library_match), palette via design_intent_values (#121212 bg, #00bcd4 accent), typography via fingerprint_font_family_match (Segoe UI). Composition installed, sites.style_collection_id pointer updated.
7. webdesign-agent: analyze_design LLM ran (5335 char response), design_spec merged into content_data top-level (color_scheme, typography, spacing keys), proceeded through generate_css and deploy_css to check_should_fork.

The actual values in the design_spec match the adopted source: `text=#e0e0e0`, `accent=#00bcd4`, `primary=#00bcd4`, `surface=#1e1e1e`, `secondary=#1a1a2e`, `background=#121212`, `text_muted=#a0a0a0`, font `'Segoe UI', Roboto, Helvetica, Arial, sans-serif`.

## Real bugs surfaced (NEW — found today, not previously tracked)

### Bug 1 — Dispatcher orchestrations stall in AWAITING_RESPONSES
**Severity:** Blocker for autonomous cascade completion.
**Symptom:** build-dispatch-loop orchestrations sit at `process_item_iter_N_call_handler` indefinitely. Work items remain `claimed` forever. Each dispatcher invocation handled 1-3 items before stalling.
**Evidence:** Multiple orchestration_states rows with `since_activity_s > 600`, while collected_data.handler_result shows `response_received_at` was populated for some iterations (response data IS arriving) but the state machine doesn't advance.
**Suspected causes (need investigation):**
- (a) **Kafka consumer reconnect failures.** Logs from agent-build-dispatch-loop pods showed: `Failed to fetch message: failed to dial: ... dial tcp 10.20.161.228:9092: i/o timeout` repeatedly to broker 0. Kafka cluster itself was healthy (`kubectl -n kafka get pods` showed all 3 brokers Running 28d, no restarts). Looks like a per-pod transient connectivity issue — possibly CNI, possibly Kafka client connection pool bug.
- (b) **mark_complete transition not firing even when response received.** For ttk-calculator and games pages, response was processed (collected_data complete) but work item stayed `claimed`. May or may not be the same root cause as (a).

**Likely fix shape:** the Kafka consumer client needs to detect a dead connection and reconnect, OR the orchestrator runtime needs a watchdog that progresses past mark_complete even when response delivery is partial.

### Bug 2 — No claim-timeout / orchestration-timeout cleanup
**Severity:** Compounds Bug 1.
**Symptom:** When a dispatcher pod dies or stalls, work items stay `claimed` forever. New dispatchers respect the claim and don't re-pick. So accumulated stuck claims block further progress.
**Fix shape:**
- Periodic sweeper that releases items with `status='claimed' AND updated_at < NOW() - INTERVAL '1 hour'` back to `triaged`
- AWAITING_RESPONSES orchestrations past their timeout_seconds should be force-marked FAILED so concurrency_group is freed

### Bug 3 — `build-pipeline-trigger` scheduled task isn't site-targeted
**Severity:** Causes Bug 1 to be much worse than it could be, because no automated mechanism re-fires the dispatcher at sites with open work.
**Symptom:** Scheduler-driven dispatcher invocations all default to `system.internal` site_id (visible in `orchestration_states.collected_data.input_data` for build-dispatch-loop owner). They find nothing dispatchable and complete clean. Meanwhile real sites with 9+ open items wait indefinitely.
**Fix shape:** Add a `pre_query` to the `build-pipeline-trigger` row in `scheduled_tasks` that selects (site_id, domain) per site with open work items in `pipeline='build'`, so the scheduler fires one dispatcher per site:
```sql
UPDATE scheduled_tasks
SET pre_query = $$
    SELECT DISTINCT s.id::text AS site_id, s.domain
    FROM sites s
    JOIN site_work_items wi ON wi.site_id = s.id
    WHERE wi.status IN ('triaged','approved')
      AND wi.pipeline = 'build'
      AND wi.attempt_count < wi.max_attempts
    LIMIT 10
$$
WHERE name = 'build-pipeline-trigger';
```
Verify the column actually exists on scheduled_tasks (it does — confirmed today).

### Bug 4 — Planner invents hero images ignoring site_archetype
**Severity:** Content quality, not blocker.
**Symptom:** site_archetype.design.imagery for gamedesign.uk says `imagery: "minimal icons/diagrams, no decorative photography"`. The planner's site_plan still produced lavish hero_home/hero_about/hero_tools/hero_contact/hero_prototypes prompts. Pages got rendered with hero-image backgrounds that aren't part of the adopted design.
**Fix shape:** site-planner prompt needs a constraint section reading site_archetype.design.imagery. If it says "none" or "minimal", set `needs_images=false` and don't generate hero prompts.

### Bug 5 — Planner invents pages that duplicate adopted ones
**Severity:** Content quality.
**Symptom:** Adoption queued `adoption_page_games_*` (the source's actual /games page). Planner separately queued `needs_page:prototypes`. Both get built. Two pages, near-identical purpose, different names, different commits.
**Fix shape:** site-planner should be aware of pages adoption already queued. Either dedupe by URL, or the planner should READ existing pages from sync_pages_to_db output and only queue NEW pages.

### Bug 6 — Content writer leaves empty fields in component templates
**Severity:** Causes Bug 7 (validator flags page).
**Symptom:** `tool-list` component on the index page rendered with empty `<span class="tl-eyebrow"></span>`, empty `href=""`, empty `data-filter=""` etc. The writer produced content for the hero and CTA but didn't fill the tool-list.
**Probable cause:** Component template has fields the writer didn't know about, or the writer is generic and doesn't have per-component context. The page-build-handler workflow has a `plan_sections → check_has_ready_sections → spawn_content_writer` flow — sections without ready data are deferred, but evidently the tool-list component was treated as ready while having empty fields.

### Bug 7 — Validator false-positives on adopted content
**Severity:** Blocks page deployment.
**Symptom:** `validate_page_content` flagged the adopted tools page with "1 blockers, 0 errors". Likely tripped on cross-site contamination heuristic seeing "gamedesign.uk" (the source domain) in the page's adopted content. For adopted pages, references to the source domain are expected and shouldn't trigger this.
**Fix shape:** validate_page_content needs an "adopted-from" domain whitelist when running on `mode=recreate` items.

## Other items still tracked (carried forward)

| # | Item | Status |
|---|---|---|
| 7 | item_key cascade_run_id scoping (so re-runs don't trigger two-strike on completed prior items) | Deferred |
| 8 | install_site_composition deliberate-reinstall flag (currently refuses to overwrite existing pointer) | Deferred |
| 9 | item_key naming standardisation (`needs_composition` vs `strategy_{domain}` vs `needs_page:{name}` vs `capability_gap:{type}:{page}`) | Deferred |
| 10 | Content pages should `depends_on` composition + design (currently parallel; renders against null collection mid-cascade) | Deferred |
| 11 | Duplicate `image-generator` agent_definition row | Deactivated today (Aug 2025 row → is_active=false; Oct 2025 row stays) |
| 18 | **Site delete should also delete library rows (css_themes, palettes, style_collections, typography_sets) where source_domain matches and the rows aren't referenced by any surviving site.** Currently FK is SET NULL on these, leaving orphans that can be picked up by the matcher for future sites. Either (a) CASCADE on the source_site_id FK, OR (b) an explicit cleanup step in whatever delete-site action exists. The semantic "library rows persist for re-use" only makes sense if rows are actually being re-used; otherwise they're just clutter. | New, low priority |
| 19 | **Discover agents should skip sites that have no current adoption / are in deleted-or-stub state.** Today's queue had 80+ `needs_content_planning` items for gamesdesign.co.uk, mostly with `[stale: triaged 48h+]` markers. These were quality-discovery / content-discovery agents running on a sites row whose adoption had failed or been deleted, repeatedly finding "no content / no tagline / no portfolio" and queuing remediation work for a site that doesn't really exist. The discover side keeps generating noise at scale. Add a precondition: discover agents should skip site_ids where (a) status='deleted' or 'archived', (b) the site has no `is_current=true` site_specs aspect='identity' or similar adoption marker, OR (c) a recent adoption is in flight. | New, medium priority |
| 20 | **Investigate whether site-adoption-orchestrator creates a new sites row for a destination_domain that already has one.** Couldn't confirm in today's session (deleted evidence before checking) but the queue carrying gamesdesign work items after multiple delete attempts is suspicious. Worth checking on next adoption run: query `SELECT id, created_at FROM sites WHERE domain = 'gamesdesign.co.uk'` BEFORE and AFTER triggering adoption. If two rows exist after, adoption is duplicating. Decision then needed: refuse adoption when destination exists (require explicit delete first), or reuse existing site_id (treat as refresh). The duplicate-creation behaviour is the worst of the three options — it leaves orphan work items pointing at a now-stale row, while a new cascade runs against a different row, with both thinking they're authoritative. | New, medium priority |

## Two-strike rule — FINAL DECISION

We discussed weakening it (count only `failed`, not `complete + failed`). **Decided NOT to weaken.** The rule's job is to break loops between discover agents and fix agents. If discover-A keeps re-finding an issue after fix-B reports `complete`, the loop runs forever unless `complete` counts toward the strike budget. The proper fix for re-cascades hitting two-strike on yesterday's `complete` items is **item_key cascade_run_id scoping** (#7 above), not weakening the rule.

The patch deployed today (work_items_common.go + load_work_item_actions.go) reflects this — Fix B was REVERTED. Comments in work_items_common.go explain why.

## Cleanup performed today (queue state)

Before handoff, ran cleanup of:
- gamesdesign.co.uk site fully deleted (FK chain: site_work_items → maintenance_queue → content_feed_items → page_component_history → link_registry.target_site_id → sites)
- All `unresolved` items with `[stale:` or `[unresolved after%` summary prefixes — these were two-strike-exhausted casualties from today's cascade and accumulated 48h debt
- All `failed` items with "Claim timed out" or "timed out" in error — dispatcher-stall casualties, not real handler failures
- `claimed` items unclaimed if claimed_at > 1h ago — they were stuck on dead pods anyway

### Library row cleanup (new pattern for "bad cascade" recovery)

Beyond the site delete, the cascade also created multiple library rows
(css_themes, palettes, style_collections) — one set per resolve attempt.
For gamesdesign these were all wrong-decision artefacts (brochure-formal
fallbacks + the eventual tool-portal-dark that we don't trust because the
cascade was a mess). Cleared them with:

```sql
BEGIN;
-- Delete in reverse-FK order: style_collections has outgoing FKs to the others
DELETE FROM style_collections WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM css_themes        WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM palettes          WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM typography_sets   WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM layouts
  WHERE source_domain = 'gamesdesign.co.uk'
    AND name NOT IN ('tool-portal-dark','social-lobby','brochure-formal',
                     'brochure-bold','technical-precise','affiliate-hub');
-- The NOT IN guard protects seeded library layouts in case any source
-- pointer got corrupted during cascade. For gamesdesign this was a no-op
-- (no fork-derived layouts), but it's the safe pattern.

-- Verify zero remaining
SELECT 'css_themes' AS tbl, COUNT(*) FROM css_themes WHERE source_domain = 'gamesdesign.co.uk'
UNION ALL SELECT 'palettes',          COUNT(*) FROM palettes          WHERE source_domain = 'gamesdesign.co.uk'
UNION ALL SELECT 'style_collections', COUNT(*) FROM style_collections WHERE source_domain = 'gamesdesign.co.uk'
UNION ALL SELECT 'typography_sets',   COUNT(*) FROM typography_sets   WHERE source_domain = 'gamesdesign.co.uk'
UNION ALL SELECT 'layouts',           COUNT(*) FROM layouts           WHERE source_domain = 'gamesdesign.co.uk';
COMMIT;
```

Counts cleared for gamesdesign:
- 4 css_themes (one per cascade resolve attempt)
- 7 palettes (multiple per attempt because re-resolve creates new palette rows even when re-using values)
- 4 style_collections (one per install)
- 0 typography_sets, 0 layouts (none were forked from this site)

This pattern should be used whenever a cascade goes badly enough that
the produced library artefacts shouldn't influence future site matchers.

## What to do next session (in priority order)

### Priority 1 — Fix Bug 1 (dispatcher mark_complete stall) and Bug 2 (claim/orchestration timeouts)

These two together are the single biggest blocker to autonomous site building. Without them, every site requires manual dispatcher pokes. Specific investigation steps:

1. Read the build-dispatch-loop Go agent code. Look at how it consumes responses from spawned handler pods and how mark_complete fires. The stall is at `process_item_iter_N_call_handler` step name — find that step transition logic.
2. Look at Kafka consumer client — is there a reconnect path? Does it handle broker connection drop?
3. Check whether `complete_work_item` action returns success even when the UPDATE affects 0 rows (e.g. when item is no longer in `claimed` state). If yes, that's why the dispatcher state advances locally but DB stays out of sync.
4. Consider a periodic claim-release sweeper as a safety net (item 8 has its own merit, but a sweeper helps even before that).

### Priority 2 — Apply migration 012 + deploy adoption→classifier handoff

Quick wins. Both ready in /mnt/user-data/outputs/. Migration 012 re-tightens the idx_swi_dedup predicate now that the Go const is the source of truth. apply_adoption_plan_action.go simplifies adoption to route through the classifier instead of bypassing it.

### Priority 3 — Fix Bug 3 (scheduler pre_query)

Add the pre_query SQL to build-pipeline-trigger. Once dispatcher reliability is fixed, this completes the loop: scheduler iterates over sites with open work, fires one dispatcher per site, dispatchers actually run to completion.

### Priority 4 — Re-run gamesdesign.co.uk

Once Priority 1+3 are in place, fresh adoption of gamedesign.uk → gamesdesign.co.uk should run autonomously without manual pokes. That's the real validation. Trigger via the standard adopt-separated kcat command:
```
SOURCE_URL=https://gamedesign.uk
DEST_DOMAIN=gamesdesign.co.uk
```

### Priority 5 — Content quality bugs (4-7)

Fix planner imagery + page-dedup respect for adoption. Fix validator's adopted-content whitelist. Fix content writer's empty-field problem. These are quality-of-output bugs, not flow-blockers.

## Useful queries for next session

```sql
-- Site-by-site queue health
SELECT s.domain, swi.status, COUNT(*) AS n
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
WHERE swi.status NOT IN ('complete','verified','rejected','wont_fix')
GROUP BY s.domain, swi.status
ORDER BY s.domain, swi.status;

-- Stuck claims across all sites
SELECT s.domain, swi.item_type, swi.item_key, swi.claimed_by,
       EXTRACT(EPOCH FROM (NOW() - swi.claimed_at))::int / 60 AS claim_age_min
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
WHERE swi.status = 'claimed'
  AND swi.claimed_at < NOW() - INTERVAL '30 minutes'
ORDER BY claim_age_min DESC;

-- Orchestrations stuck in AWAITING_RESPONSES
SELECT owner_agent_type, COUNT(*),
       MAX(EXTRACT(EPOCH FROM (NOW() - last_activity))::int / 60) AS oldest_stuck_min
FROM orchestration_states
WHERE status = 'AWAITING_RESPONSES'
  AND last_activity < NOW() - INTERVAL '15 minutes'
GROUP BY owner_agent_type
ORDER BY 3 DESC;

-- Validate the deployed Go patch is in effect (look for the centralised predicate)
-- Should see workItemTerminalStatuses being used:
-- (a) In ON CONFLICT WHERE in insertWorkItem (line 1003-1017 area)
-- (b) In composition-lookup SELECT (line 401-415)
-- (c) Via fmt.Sprintf with sqlInList helper

-- Manual dispatcher trigger script (existing) at:
-- /mnt/user-data/outputs/trigger-dispatch-gamesdesign.sh
-- (parameterise the site_id/domain for any site needing a poke)
```

## Files referenced in this session that exist in /mnt/user-data/outputs/

- `work_items_common.go` (DEPLOYED, kept for reference)
- `load_work_item_actions.go` (DEPLOYED, kept for reference)
- `apply_adoption_plan_action.go` (NOT YET DEPLOYED)
- `007_adoption_pipeline_v4.md`
- `012_idx_swi_dedup_include_unresolved.sql` (NOT YET APPLIED)
- `trigger-dispatch-gamesdesign.sh`
- `read_layout_taxonomy_action.go` (DEPLOYED, kept for reference)

## End-of-session state

- gamesdesign.co.uk site row + all FK-chained data: DELETED
- Library rows (css_themes, palettes, style_collections) sourced from gamesdesign.co.uk: DELETED
- Git directory `sites/gamesdesign.co.uk/`: deleted earlier in session
- Cleanup of stale unresolved/failed items across all sites: PENDING (run cleanup SQL block at top of next session if desired)
- LLM credit refill: pending user action

Ready for fresh adoption trigger when next session resumes.
