# PLAN — generalising the mini-lobby session's fixes to all domain work

**Created:** 2026-07-10 · **Status of the originating task:** CLOSED (see `VERDICT_minilobby_trim_method.md` §7)
**Companions:** `RUNNING_NOTES_minilobby_task.md` (chronological) · `RUNBOOK_minilobby_task.md` (operational)

---

## §1 The question this answers

The mini-lobby trim was one section on one site. But almost nothing that went wrong was about vonc.com.
This plan states which findings are **fleet-generic**, where each one's **detection** now lives (a discovery
check) and where its **repair** lives (a fixer), and what is still open.

---

## §2 The generalisation test — four rules

Applied to every fix in this session. They are the difference between "we fixed vonc" and "the fleet cannot
regress this way again".

1. **Fix the writer, not the row.** We hand-repaired `build_status` twice before fixing
   `UpdatePageStatusAction`. A row repair is evidence you have found a bug, never the fix for it. If the
   remedy is a `psql` statement scoped to a `site_id`, it is not yet a fix.
2. **Detect by contract, not by name.** The runtime-fill guard keys on the `data-runtime-fill` marker — the
   author's declaration of intent — never on `provocation-card` / `lobby-grid`. A check that names a component
   is a site-specific check wearing a fleet-shaped coat.
3. **Surface; never silently skip.** Where a check must not act (a runtime-fill shell, a `pending` row awaiting
   rebuild) it emits a **Finding**, not silence. Findings are informational in the action's return value and
   raise nothing. Silence is the failure mode this codebase keeps paying for — `complete_error` reporting
   success on zero sections; three enabled check names with no implementation; a live section carrying a status
   no reader honours.
4. **Every detection needs a fixer, or a written reason there is none.** A check that raises a work item no
   agent handles is a leak. A check that deliberately does not raise one must say why, in the check.

A fifth, inherited and reconfirmed twice this session: **verify by artifact, never by item status.**
`md5(rendered_html) == md5(replace(html_template,'<no value>',''))` is a verification. `status = COMPLETED` is not.

`Categories:` (decision)

---

## §3 Classification of everything found

| # | Finding | Class | Scope | Detection | Repair | Status |
|---|---|---|---|---|---|---|
| 1 | `apply_section_edit` left `build_status='approved'`; every discovery check filters `= 'deployed'`, so a live section vanished from the whole audit surface | silent invisibility | **generic** — any section edit, any site | **NEW** `page_component_status_drift` | **NEW** `repair_page_component_status` fix_type (`component-template-fixer`) | Writer fixed + **verified live** (v1.0.1102). Fixer **shipped** (v1.0.1103). Check written, **untracked — needs commit + push** |
| 2 | `auto_lock_on_deploy` stamped `schema_mode='strict'` for a stillborn subsystem (snapshot columns never created; 0 Go readers; fired once ever) | dead machinery | **generic** — all 10 sites are `first_deploy` | n/a (schema-level) | migration `009_drop_auto_lock_on_deploy.sql` | **DONE, live** |
| 3 | `provocation-card` / `lobby-grid` templates are rendered artifacts (`<no value>`, 0 slots, 0 schema fields) | corrupted template | **generic pattern**; the *runtime-fill exemption* is the generic insight | `component_template_corrupted` + `broken_template_slots`, **now runtime-fill-guarded** | `component-creator` regeneration; runtime-fill shells excluded (finding only) | Guards **shipped** (v1.0.1103); checks still enabled nowhere |
| 4 | `brief-explanation` ships `<img src="">` — illustration assets exist but `filename`/`storage_path` NULL, nothing under `/assets/images/` | undeployed asset | **generic** | `undeployed_assets` **exists and is enabled** — but has never fired for vonc (0 items). Worth its own look | `asset-deployer` | **OPEN** → imagery workstream |
| 5 | 19 `page_components` are `pending` on `deployed` pages with **zero open work items** — invisible to every check, nothing scheduled to fix them | stuck backlog | **generic** — 5 sites | reported as findings by the new drift check | rebuild (a status flip would *hide* the staleness) | **OPEN, now surfaced** |
| 6 | 8 checks registered in Go, enabled in no agent — incl. `sectionless_pages`, the exact detector for the ten silent `complete_error` builds | wiring gap | **generic** | n/a (meta) | enable in a discovery agent's `checks` array | **OPEN — decision** (§5) |
| 7 | 3 enabled check names have no Go implementation (`missing_content`, `orphan_nav`, `stale_pages`, all `maintenance-triage`) | config/code drift | **generic** | runner already `logger.Warn`s "Unknown discovery check" — honest, not silent | implement or delete the names | **OPEN** |
| 8 | `page_components.build_status` has **no CHECK constraint** — free text. This is the root enabler of #1 | missing invariant | **generic** | n/a | proposed migration (§4) | **OPEN — needs approval** |
| 9 | `fix_component_template`'s header claimed `remove_element` defers page-component content to the section editor, implying it touches page components. It operates on `site_components` only | misleading doc | **generic** — it mis-planned this very task | n/a | header rewritten | **DONE** |
| 10 | `commit_from` listed in `coordinator.go` as "Used by update_page_status"; the action never reads it | dead config | generic | n/a | n/a | documented |
| 11 | `js_snippets` has no `updated_at` column (the handoff's schema notes implied one) | schema doc drift | generic | n/a | n/a | documented |
| 12 | `last_built_at` never written by build or rerender | dead column | generic | n/a | decide: write or drop | **OPEN** |
| 13 | `provocations.json` still ships a dead `lobby` key | stale data | **site-specific (vonc)** | n/a | drop when Phase-3 emits the file | **OPEN, trivial** |
| 14 | **`empty_sections` (enabled all along) flags runtime-fill shells as `empty_heading`** and routes them to page-build-handler — a full LLM rebuild of the shell. Caught live on the first enabled-checks pass, 2026-07-10 | same landmine, 3rd instance | **generic** — any runtime-fill site | the check itself; now runtime-fill-guarded (finding only) | items rejected by hand; guard ships next build | Guard **written**, needs build. The landmine existed in **three** checks: two guarded by code reading, this one caught by running the loop |

**Read of the table:** of thirteen findings, exactly one (#13) is site-specific. The mini-lobby trim was a
keyhole onto fleet-wide defects.

`Categories:` (root-cause, decision)

---

## §4 Proposed structural fix for #8 (not applied — needs approval)

`build_status` being free text is what let `apply_section_edit` invent `approved` and remove a live section from
the audit surface unnoticed. A CHECK constraint turns that from a silent invisibility into a loud write failure.

```sql
-- NOT YET APPLIED. 'approved' is retained: apply_section_edit legitimately writes it
-- as a pre-deploy state, and update_page_status now advances it to 'deployed'.
-- Verify no other value exists first:
--   SELECT DISTINCT build_status FROM page_components;   -- expect deployed, pending
ALTER TABLE page_components
  ADD CONSTRAINT page_components_build_status_check
  CHECK (build_status IN ('deployed','pending','approved','removed','needs_rebuild'));
```

Risk: any writer using a value outside the set starts erroring. That is the point, but it should land with the
chassis, not before it. The drift check (#1) then covers the *residual* case the constraint cannot — an
`approved` row whose deploy step failed and never advanced.

Same argument applies to `pages.build_status`. Not surveyed here.

---

## §5 Open decision: which registered-but-disabled checks to enable

**DECIDED + APPLIED 2026-07-10:** `page_component_status_drift`, `sectionless_pages` and
`component_template_corrupted` are enabled in **completeness-discovery-agent** (backup
`_completeness_agentdef_backup_20260710`). First pass run manually on vonc.com — the runtime-fill guard fired
correctly on its first live use (3 regenerations emitted, both shells refused, 0 drift emissions). It also
exposed finding #14: the long-enabled `empty_sections` check carried the same landmine; guard written, ships
next build. Note the improvement-sweep scheduler has been disabled since 2026-05-02, so discovery runs only by
manual trigger (`075_trigger_discovery.sh <domain> completeness`) until it is re-enabled.

Remaining registered-but-disabled candidates:

| check | notes |
|---|---|
| `validate_component_standards` | guard shipped — safe to enable; several sub-checks, survey emissions first |
| `cross_site_contamination` | unknown | survey first |
| `unrendered_templates` | unknown | survey first |
| `phantom_internal_links` | unknown | survey first |
| `backend_unreachable` / `tool_recreation_needed` | unknown | survey first |

Enabling SQL is in `RUNBOOK_minilobby_task.md` §4.

---

## §6 What shipped, and what is waiting on a push

**Live now (DB):** `009_drop_auto_lock_on_deploy.sql`; the section-editor workflow-row config
(`page_component_id_field`, inert until the binary caught up — it now has).

**Live now (chassis `v1.0.1102`):** `UpdatePageStatusAction`'s `page_component_id_field`; `coordinator.go`'s
`dataRefKeys` registration. Verified end-to-end 2026-07-10.

**Live now (chassis `v1.0.1103`, commit `49d67e82`):**

- `discovery_checks/check_component_template_corrupted.go` — runtime-fill guard (emission skip + finding)
- `discovery_checks/check_component_standards.go` — same guard on `broken_template_slots`
- `fix_component_template_action.go` — **new fixer** `repair_page_component_status`, + corrected header

Confirmed safe on the running pod: `component_template_corrupted` and `validate_component_standards` are
**still enabled in no agent**, and no `needs_component_regeneration` item has ever been raised against
`provocation-card` or `lobby-grid`. The guard shipped before it was ever needed.

**Live now (chassis `v1.0.1104`):**

- `discovery_checks/check_page_component_status_drift.go` — **new check**, string-verified in the running
  binary 2026-07-10. RUNBOOK §3 verification passed (0 emissions, 3/2 guard split).

Everything this plan produced has shipped. The single remaining action is the §5 enable decision.

Build practice note: images are built from the local filesystem via the Makefile; commits are decoupled and at
the user's discretion (image tag now recorded in commit messages by hand). Verify deployed contents against the
pod, not git history.

`Categories:` (next-task)
