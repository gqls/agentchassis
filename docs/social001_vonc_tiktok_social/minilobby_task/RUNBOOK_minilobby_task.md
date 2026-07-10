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
| Runtime-fill guards (2 checks) | **SHIPPED** in `v1.0.1103` (`49d67e82`) — landed before the check was ever enabled |
| `repair_page_component_status` fixer | **SHIPPED** in `v1.0.1103` |
| `page_component_status_drift` check | **WRITTEN, untracked** — needs commit + chassis push |
| Enabling the 8 disabled checks | **OPEN — decision**, see PLAN §5 |
| `brief-explanation` `<img src="">` | **OPEN** → imagery workstream |
| `provocations.json` dead `lobby` key | **OPEN**, trivial |

**Next action:** commit + push `check_page_component_status_drift.go`, then §3 verification, then the §5 enable
decision. Nothing is currently at risk: as of 2026-07-10 the two template checks are enabled in no agent, and
no `needs_component_regeneration` item has ever been raised against `provocation-card` or `lobby-grid`.

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

## §3 Verifying the new code

Three of the four Go changes shipped in `v1.0.1103` (both runtime-fill guards + the
`repair_page_component_status` fixer). The fourth — `check_page_component_status_drift.go` — is untracked;
commit it and run this section again after its push. Everything here is inert until enabled, so this is a
smoke test, not a gate.

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
The runtime-fill guards shipped in `v1.0.1103`, so `component_template_corrupted` and
`validate_component_standards` are now safe to enable. (Before that build they would have handed vonc's two
runtime-fill shells to `component-creator` / `component-template-fixer`.) `page_component_status_drift` cannot
be enabled until its file is committed and pushed — the running binary does not contain it, and the runner
would log "Unknown discovery check" each pass.

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
