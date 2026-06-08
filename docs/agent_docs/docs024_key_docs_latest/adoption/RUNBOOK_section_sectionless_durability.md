# RUNBOOK section — sectionless-page durability (2b + S1)

Drop-in for the deploy runbook. Covers two chassis changes that together make a
sectionless page self-healing rather than a silent dead-end:

- **2b** — `load_page_sections_from_spec_action.go`: read-time sibling fallback.
  Makes a build that runs on a sectionless page synthesise its layout from a
  same-role sibling and proceed.
- **S1** — `discovery_checks/check_sectionless_pages.go`: detection + retrigger.
  Finds plan pages with empty sections (that a sibling can fix) and re-issues a
  build, closing the gap where a page is left sectionless and never retried
  (e.g. a build that died on claim-timeout).

Both are Go chassis changes → rebuild the agent-chassis image and roll the pods
in `ai-persona-system`. They do not touch the site-deploy path (git → Backblaze).

---

## 0. Pre-deploy schema confirmation (standing rule: fresh `\d`)
The SQL embedded in both files reads these columns — confirm before building:
```sql
\d site_plan_pages      -- expect: plan_id, name, role
\d site_plan_sections   -- expect: plan_id, page_name, ordering, component_name
\d pages                -- expect: sections (jsonb), status, page_type
```
If `site_plan_pages.role` or the `site_plan_sections` column names differ, stop
and adjust the queries before building.

## 1. Place files
```
platform/orchestration/actions/load_page_sections_from_spec_action.go   (replace)
platform/orchestration/actions/discovery_checks/check_sectionless_pages.go (new)
```
The new check self-registers via `init()` (no registry.go edit needed).

## 2. Build + deploy chassis
- Build the agent-chassis image (your normal pipeline).
- Roll the pods:
  ```
  kubectl -n ai-persona-system rollout restart deploy/<agent-chassis-deploy>
  kubectl -n ai-persona-system rollout status  deploy/<agent-chassis-deploy>
  ```
- Confirm the new check registered (appears once a discovery agent runs — see §4):
  the `RunDiscoveryChecksAction: Running checks` log line lists `registered`
  names; `sectionless_pages` should be present.

## 3. Enable S1 in the completeness-discovery-agent (config, not code)
First inspect (also confirms the step is named `run_checks`; if this returns
NULL the path/step name differs and must be found before updating):
```sql
SELECT default_config #> '{workflow,steps,run_checks,config,checks}' AS checks
FROM agent_definitions
WHERE type='completeness-discovery-agent' AND deleted_at IS NULL;
```
Snapshot, then append idempotently (does not overwrite the existing array):
```sql
CREATE TABLE IF NOT EXISTS agent_def_bak_sectionless AS
SELECT * FROM agent_definitions
WHERE type='completeness-discovery-agent' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,run_checks,config,checks}',
    COALESCE(default_config #> '{workflow,steps,run_checks,config,checks}','[]'::jsonb)
      || '["sectionless_pages"]'::jsonb
)
WHERE type='completeness-discovery-agent' AND deleted_at IS NULL
  AND NOT (COALESCE(default_config #> '{workflow,steps,run_checks,config,checks}','[]'::jsonb)
           ? 'sectionless_pages');
```
Re-run the inspect SELECT to confirm `sectionless_pages` is now in the array.

## 4. Verify
**2b (sibling fallback)** — on a test site, blank one guide's sections, then
re-trigger and watch for the synthesis:
```sql
-- snapshot + blank a known-good test page (NOT a live one)
CREATE TABLE IF NOT EXISTS pages_bak_2btest AS
SELECT * FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') AND name='<test-guide>';
UPDATE pages SET sections='[]'::jsonb
WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') AND name='<test-guide>';
-- queue a recreate build for it (handler page-build-handler)
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, page_id, priority, handler_agent, status, created_by, spec, item_key)
SELECT s.id,'manual-rebuild','build','needs_content_page','medium','2b verify',
       p.id,90,'page-build-handler','triaged','manual-rebuild',
       jsonb_build_object('mode','recreate','source','adoption','page_name','<test-guide>','page_type','guide'),
       'needs_content_page:<test-guide>-2bverify'
FROM sites s JOIN pages p ON p.site_id=s.id AND p.name='<test-guide>'
WHERE s.domain='<domain>' ON CONFLICT DO NOTHING;
```
Pass: chassis logs `LoadPageSectionsFromSpec: SYNTHESISED layout from same-role sibling`
(WARN, naming the page + sibling); the build runs past ~90s; `pages.sections`
repopulates; `page_components` non-empty; `build_status='deployed'`.

**S1 (detection + retrigger)** — after §3, trigger the completeness-discovery
agent for the site (your normal trigger). Pass: the run logs `sectionless_pages`
under both `registered` and `checks`; any plan page with empty sections and a
same-role sibling gets a `sectionless_page:<name>:<site>` item, which
page-build-handler then builds (via 2b). Confirm no churn: a page with no usable
sibling is NOT emitted; a persistently-failing item is flagged `unresolved` by
the two-strike rule, not looped.

## 5. Rollback
- Code: redeploy the previous chassis image (revert both files). 2b and S1 are
  additive — reverting removes the fallback and the check; no data migration.
- Config: remove the array element (snapshot in `agent_def_bak_sectionless`):
  ```sql
  UPDATE agent_definitions
  SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config #> '{workflow,steps,run_checks,config,checks}') - 'sectionless_pages'
  )
  WHERE type='completeness-discovery-agent' AND deleted_at IS NULL;
  ```
- `pages.sections` rows that 2b/S1 already repaired are valid layouts and need no
  rollback.

## Notes / boundaries
- 2b and S1 only ever write `pages.sections` (the build-read field). Neither
  touches `site_plan_sections`; relational-plan hygiene stays a separate concern.
- A genuinely sectionless page with NO same-role sibling is intentionally out of
  scope (cannot be repaired from a sibling). Making that case visible instead of
  a silent SUCCESS completion is **S2** (page-build-handler `check_has_ready_sections`
  ELSE), tracked separately. The deeper "stop creating residue" work
  (positive-evidence completion + the three silent-completion modes) is tracked
  in `FOCUS_page_build_handler_silent_completion.md`.
