# RUNBOOK — mini-lobby task and the fleet fixes it produced

**Revised:** 2026-07-10 · **Companions:** `PLAN_generalise_fixes_to_fleet.md` (classification, decisions) ·
`RUNNING_NOTES_minilobby_task.md` (chronological) · `VERDICT_minilobby_trim_method.md` (method + outcome)

---

## §0 YOU ARE HERE

| thread | state |
|---|---|
| Mini-lobby trim (vonc index) | **CLOSED** — live, browser-verified, six sections, lobby-grid untouched |
| provocation-card centring (820px) | **CLOSED** — live |
| section-editor `approved` defect | **CLOSED** — fixed in chassis `v1.0.1102`, verified end-to-end 2026-07-10 |
| `auto_lock_on_deploy` trigger | **CLOSED** — dropped, migration `009`, reversal saved |
| Runtime-fill guards (template checks ×2) | **SHIPPED** `v1.0.1103`, confirmed in the `v1.0.1104` binary |
| `repair_page_component_status` fixer | **SHIPPED** `v1.0.1103`, confirmed in `v1.0.1104` |
| `page_component_status_drift` check | **SHIPPED** `v1.0.1104`; §3 checks pass |
| Enable decision | **DONE 2026-07-10** — `page_component_status_drift`, `sectionless_pages`, `component_template_corrupted` added to completeness-discovery-agent; first pass run on vonc; guard fired correctly (3 emit / 2 refuse) |
| Runtime-fill guard #3 (`check_empty_sections`) | **SHIPPED `v1.0.1105` + PROVEN** — re-ran discovery on vonc; the rejected shell items were NOT re-raised (they sit outside the partial dedup index, so absence is positive proof); both guards logged their exemptions by name in the spawned pod |
| 3 × `needs_component_regeneration` + 3 × `empty_section` (genuinely corrupted components) + 1 × `needs_rerender` | **CLOSED 2026-07-10** — manually dispatched via `087` (improvement-sweep still off). All 3 components regenerated; all 3 pages rebuilt with real content; the 2 runtime-fill shells (`provocation-card`, `lobby-grid` on index) **correctly rejected** by the guard, not rebuilt; index reassembled. Live-verified: `/archetypes.html`, `/about.html`, `/index.html` all HTTP 200 with rebuilt sections present and shells intact (`data-runtime-fill="true"`). See RUNNING_NOTES for the dispatch mechanics. |
| `brief-explanation` `<img src="">` | **CLOSED 2026-07-11** — root cause was a stale render predating the section-imagery resolver, not an undeployed asset (imagery workstream had already committed all 16 vonc assets to git). Fixed with one `page_rerender` item (`reason: image_landed` → light re-render, no LLM); src now `/assets/images/illustration-gauntlet-cta.jpg`, live-verified, shells md5-pristine. See PLAN §3 #4 for the `undeployed_assets`-never-fired answer (design-discovery-agent has never run on vonc) |
| `provocations.json` dead `lobby` key | **CLOSED 2026-07-11** — dropped in sites repo commit `c244ddc`, live-verified (keys: `generated_at/today/arena/archive`) |
| `page_components.build_status` CHECK constraint (PLAN §4) | **CLOSED 2026-07-11** — migration `049_page_components_build_status_check.sql` applied + negative-tested; invented statuses now fail loudly at write time |
| Stale loader comment | **OPEN, cosmetic** — provocation-card-loader's header in `js_snippets` (and the bundled snippets.js) still says "daily provocation + mini-lobby"; the mini-lobby fill is gone. Fix = update `js_snippets.description`/comment + re-run `083_trigger-asset-renderer-vonc.sh` |
| Archetype hub (Approach A) | **CLOSED 2026-07-12** — 8 entity pages built + grid refilled; `088_archetype_entity_pages.sql` + running-notes 2026-07-12 entry. Grid is build-time **query-resolved** (`pages_where_type:entity-page`); the old `entity_page` source was kebab-unrepresentable. All 8 icons now consumed (page-scope hero imagery rows) |
| Design-discovery survey findings (vonc) | **8 of 16 CLOSED** by the hub build (undeployed_asset ×8). Still at `detected`: deactivated chrome pointers ×3, stale-chrome rerenders ×3, hardcoded colours ×1, evaluate_tools ×1 |
| `needs_page:provocation` trap | **STANDING** — `reconcile_site_plan` re-emits it (triaged, dispatchable!) every run while `blog/provocation.html` sits at `planned`. It belongs to the Spark pipeline. **Park it to `detected` after every reconcile of vonc**, or build/re-status the page |
| content-block-about static stat labels | **FIXED 2026-07-12 (fleet), migration `090`** — `stat_1/2/3_label` + `cta_label` were `source='static'` on shared component `4e448d51` (13 pages / 5 sites), forcing 'Clients Served'/'Satisfaction Rate'/'Awards Won'/'Learn More About Us' on every render; the writer could only fill VALUES, so ALL sites showed crossed pairs ('500+ Models / Clients Served'). Flipped `static→llm` (fallbacks kept). Business sites safe — labels persisted in their content_data, live HTML untouched until they rebuild. vonc's 8 archetype pages re-authored + rerendered |

**This thread's engineering is COMPLETE as of `v1.0.1105` (2026-07-10).** Every check, guard, fixer and
writer-fix is deployed and artifact-verified. The earlier caution about re-running completeness discovery on
vonc is **lifted**. The 7 legitimate work items were manually dispatched and **all reached terminal state
2026-07-10** (5 complete, 2 shells correctly rejected) — live-verified. What remains is separate:
the imagery-workstream `<img src="">`, the dead `lobby` key, and the PLAN §4 CHECK-constraint proposal.

**Dispatch note (2026-07-10):** manual dispatch via `087` needed **five passes** because page-build handlers
intermittently left items stuck at `claimed` without spawning (survived across the `v1.0.1107` deploy). The
recovery each time was to reset the stuck row (`status='triaged', claimed_by=NULL, claimed_at=NULL`) and re-run
`087`. Two `empty_section` items shared one page (`/about.html`, page `a28abcd7`); the first page rebuild
regenerated **all** its sections, so the second item was redundant and was closed by artifact
(`build_status='deployed'`, real `rendered_html`) rather than by a second handler run — a page rebuild is
whole-page, so co-page `empty_section` items are duplicates.

**Build practice (2026-07-10):** images are built from the **local filesystem via the Makefile**; commits are
at the user's discretion and unrelated to builds — the source of truth for code is local. The Makefile image
tag is now being added to commit messages by hand. Do not infer a deployed binary's contents from git history;
verify against the running pod (`kubectl exec <pod> -- grep -ac '<symbol>' /app/agent-chassis`).

---

## §1 Key ids

| thing | id |
|---|---|
| site vonc.com | `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74` |
| page: index | `b4d24f8e-fccd-49df-9dad-aa56a0b20a68` |
| page_component: provocation-card @ index | `a757434e-ab8a-4d2d-bfee-0fb6932f140e` |
| component: provocation-card | `6163ff14-9f94-4962-aa19-d2718eabdeb1` |
| component: lobby-grid | `9304f14d-e19b-4ce1-b3fd-f6a315aec6ed` |
| component: provocations-archive-list (the healthy reference) | `70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37` |

Backups taken 2026-07-09: `_vonc_pc_backup_20260709`, `_vonc_cc_pcard_backup_20260709`,
`_vonc_snippet_backup_20260709`, `_section_editor_agentdef_backup_20260709`.
Template snapshots in `component_versions`: `change_source` `minilobby_trim_20260709`, `pc_container_centre_20260709`.

---

## §2 Commands

Database: `kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db`

| script (in `minilobby_task/`) | what it does |
|---|---|
| `trim_minilobby.sql` | backups + `component_versions` snapshot + template + loader UPDATE, self-verifying |
| `centre_provocation_card.sql` | `.pc-container` max-width 1200px → 820px |
| `086_section_edit_provocation-card_vonc.sh` | section-editor `content_edit` — re-render one section from its template, reassemble, deploy |
| `auto_lock_on_deploy.FUNCTION_BACKUP.sql` | reversal for the dropped trigger |

Elsewhere: `docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql`;
`scripts/initial_messages/210_vonc_trigger/083_trigger-asset-renderer-vonc.sh` (rebuild `snippets.js`);
`.../083_rerender-index-vonc.sh` (assemble-only rerender + deploy).

**Log-hunting gotcha:** the section-editor's actions log in the **spawned `agent-section-editor-*` pod**,
not the main `agent-chassis` pod. `kubectl logs -l app=agent-chassis` shows only the routing envelope.

---

## §3 Verifying the new code — PASSED against `v1.0.1104`, 2026-07-10

All four Go changes are in the running binary (string-verified in the pod: `page_component_status_drift` ×7,
`repair_page_component_status` ×4, the guard's log line ×1). Query 1 returned 0 (regression guard quiet);
query 2 returned the 3-emit/2-skip split. Everything remains inert until enabled. Re-run this section after
any future chassis deploy.

```bash
# 1. The drift check emits nothing today (regression guard) and reports 19 pending rows.
#    Confirm the underlying state is unchanged:
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT COUNT(*) AS should_be_0
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.build_status = 'deployed'
    AND COALESCE(pc.build_status,'') NOT IN ('deployed','pending','removed','needs_rebuild');"

# 2. The runtime-fill guard: vonc must split 3 emit / 2 skip.
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT DISTINCT cc.function,
         cc.html_template LIKE '%data-runtime-fill%' AS runtime_fill_skipped
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  JOIN content_components cc ON cc.id=pc.component_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND cc.is_active AND cc.html_template LIKE '%<no value>%' ORDER BY 1;"
# expect: archetype-grid f | game-master-explanation f | lobby-grid t
#         platform-comparison f | provocation-card t
```

**Fixer smoke test (safe, reversible).** Drive `repair_page_component_status` end to end by manufacturing the
exact bug it exists for, then letting the fixer undo it:

```sql
-- arrange: reintroduce the 2026-07-09 defect on one row
UPDATE page_components SET build_status = 'approved'
WHERE id = 'a757434e-ab8a-4d2d-bfee-0fb6932f140e';
```

Then enable `page_component_status_drift` (§4), run the discovery agent for vonc, and confirm one work item
`page_component_status_drift:a757434e-…` routed to `component-template-fixer`, and that the row returns to
`deployed`. If anything stalls: `UPDATE page_components SET build_status='deployed' WHERE id='a757434e-…';`
(no lock to clear — the trigger is gone).

---

## §4 Enabling a discovery check

Checks are registered in Go by `init()` but only run when named in a discovery agent's config array.
As of `v1.0.1104` every check in this document is in the running binary and safe to enable, including
`page_component_status_drift`. (Before `v1.0.1103`, the two template checks would have handed vonc's
runtime-fill shells to `component-creator` / `component-template-fixer`.)

```sql
-- Back up the row first; the name must match a Go Name(), else the runner
-- logs "Unknown discovery check — not registered" and skips it.
CREATE TABLE IF NOT EXISTS _agentdef_backup_<yyyymmdd> AS
  SELECT * FROM agent_definitions WHERE type = 'completeness-discovery-agent';

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config #> '{workflow,steps,run_checks,config,checks}') || '"page_component_status_drift"'::jsonb
    )
WHERE type = 'completeness-discovery-agent';

-- verify
SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
FROM agent_definitions WHERE type = 'completeness-discovery-agent';
```

Registered check names (Go `Name()` values) are logged on every run under `registered`. Currently registered but
enabled nowhere: `component_template_corrupted`, `sectionless_pages`, `validate_component_standards`,
`cross_site_contamination`, `unrendered_templates`, `phantom_internal_links`, `backend_unreachable`,
`tool_recreation_needed`, and the new `page_component_status_drift`.

Enabled but **not implemented** (runner warns each pass): `missing_content`, `orphan_nav`, `stale_pages` —
all in `maintenance-triage`. Implement or delete the names.

---

## §5 Standing rules earned here

- **A `complete` work item proves nothing.** Verify by artifact: DB row, `curl`, browser. `md5(rendered_html) ==
  md5(replace(html_template,'<no value>',''))` is a verification.
- **`<no value>` in an `html_template` is not always a defect.** In a `data-runtime-fill` shell it *is* the
  mechanism: `RenderTemplate` strips it, leaving the empty shell the loader fills. Guard on the marker, never on
  the component name.
- **`page_components.build_status` is free text.** Any value other than `deployed` removes a live section from
  every discovery check. Fix the writer; the drift check is the backstop.
- **`rerender_single_page` is assemble-only.** A template-only edit will not appear through it. Use the
  section-editor (`apply_section_edit`) for one section; `rerender_page_sections` re-renders *all* of them and
  will pick up any sibling whose template post-dates its instance.
- **`js_snippets` has no `updated_at`.** Direct SQL is its sanctioned writer; `render_js_snippets_for_site` only
  reads.
- **Never reuse a dated backup name** — `CREATE TABLE IF NOT EXISTS` silently no-ops against an old one.
