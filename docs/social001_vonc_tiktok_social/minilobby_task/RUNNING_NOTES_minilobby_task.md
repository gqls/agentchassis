# RUNNING NOTES — vonc.com / Spark: mini-lobby trim session

Chronological log. Companion to `VERDICT_minilobby_trim_method.md` (the method verdict, outcome
tables, and backlog) and `HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md` (the inbound handoff).
Successor to the old environment's `RUNNING_NOTES_vonc_v2.md`, which lived in
`/mnt/user-data/outputs/` and did not travel; this one lives in the repo.

---

## 2026-07-09 — session log (condensed)

1. Read the bundle (`/tmp/bundle_minilobby_trim.md`) + scoped Go sources; settled the method —
   template edit by direct SQL (with `component_versions` snapshot), propagated by the
   **section-editor** (`content_edit`), loader by direct SQL + asset-renderer. Full reasoning in
   VERDICT §0–§4.
2. Key discovery: provocation-card and lobby-grid templates are **rendered artifacts stored as
   source** — `rendered_html == html_template minus '<no value>'`, exact by byte count. VERDICT §2.
3. `trim_minilobby.sql` committed: template 10300 → 6618, loader 4879 → 3365. Section-editor's first
   ever production run re-rendered the instance 10040 → 6488 (md5 == prediction). Asset-renderer
   redeployed snippets.js (3 snippets, `data.lobby` gone). All acceptance checks passed; browser
   confirmed by user (screenshot, 19:10).
4. Defect surfaced: `apply_section_edit` left the row `build_status='approved'` — the only such row
   in 578 — invisible to every discovery check. Repaired by hand (two-statement transaction to dodge
   the auto-lock trigger). VERDICT §8.1.
5. Cosmetic follow-up: `.pc-container` max-width 1200px → 820px (`centre_provocation_card.sql` +
   second section-editor run; 6459 bytes live, prediction held again). The `approved` defect recurred
   exactly as predicted; repaired again.
6. User's screenshot exposed `brief-explanation`'s `<img src="">` — root-caused to never-deployed
   illustration assets (`filename`/`storage_path` NULL, `/assets/images/` 404). Deferred to the
   imagery workstream. VERDICT §8.5a.
7. **Section-editor fix implemented** (user-approved): Go patch to `UpdatePageStatusAction`
   (optional `page_component_id_field`, mirrors the deploy mark onto the named page_component) +
   one-line `coordinator.go` dataRefKeys registration + workflow-row config addition (inert until
   chassis ships) + **dropped the `auto_lock_on_deploy` trigger**. VERDICT §8.0.
   Awaiting: user's chassis rebuild + push.

---

## 2026-07-10 — section-editor fix VERIFIED on deployed chassis

Chassis `v1.0.1102` (commit `8f9fe537`) deployed. Ran the no-op section-edit
(`086_section_edit_provocation-card_vonc.sh`, correlation `78d283bf-82d1-4705-ac3a-599a6786a0b4`):

- Row landed at **`deployed` automatically** — no manual repair, first time since the defect was found.
- `rendered_html` unchanged (6459), `content_data` unchanged, no `schema_mode`, no lock.
- Spawned-pod zap log (pod `agent-section-editor-45e02cd2-dzs28`) shows
  `UpdatePageStatusAction: Updated page` → 5 ms → `Marked page_component deployed` for `a757434e-…`.
- Table-wide sweep: **0 `approved` rows, 0 locked rows** (592 deployed / 22 pending).
- Live index still healthy: `pc-card` 0, six components, both `data-runtime-fill` markers, `820px` rule served.

The manual `approved → deployed` repair is **retired**. Operational note for log-hunting: the
section-editor's actions log in the **spawned `agent-section-editor-*` pod**, not the main
`agent-chassis` pod — grepping `-l app=agent-chassis` only finds the routing envelope.

`Categories:` (milestone, fix)

---

## 2026-07-10 — runtime-fill guard added to two discovery checks (loaded gun defused)

While mapping the discovery-check / fixer architecture for the "make it generic" pass, found that
**two registered checks would dismantle vonc's central mechanism the moment the improvement loop is
switched on.** Neither is currently enabled in any agent's `checks` array, so nothing has fired yet
(`0` rows with `item_key LIKE 'component_template_corrupted:%'`).

**The landmine.** Both checks match `html_template LIKE '%<no value>%'` with no exemption for
`data-runtime-fill`:

| check | routes to | would have done |
|---|---|---|
| `component_template_corrupted` | `component-creator` (`needs_component_regeneration`) | regenerate the shells with real `{{.field}}` slots + input_schema → next content pass bakes build-time copy in → mini-lobby plausibly returns, copy-flash before the loader runs, marker likely lost |
| `validate_component_standards` → `broken_template_slots` | `component-template-fixer` (`repair_template_slots`) | snapshot to `component_versions`, then bail `needs_regeneration` (no `</no>` tags) — non-destructive but churns a work item **and a version row every pass, forever** |

For a runtime-fill shell the literal `<no value>` **is the mechanism, not the defect**: `RenderTemplate`
strips it, leaving exactly the empty shell the loader fills. The cross-site guard already in
`component_template_corrupted` does not help — it only suppresses a *duplicate* open item, never the
first one.

**The fix (both files).** Keep the shells matching the query, but exclude them from **work-item
emission** and record a `Findings` entry instead (findings are informational in the action's return
value; they raise nothing). A bare `AND html_template NOT LIKE '%data-runtime-fill%'` in the SQL was
rejected as the wrong shape: it would make the shells vanish from the check entirely, and a silent
skip is the exact failure mode this codebase keeps paying for (`complete_error`, the three phantom
check names below).

**Verified against the live DB**, both queries run verbatim. vonc splits 3 emit / 2 skip:

```
archetype-grid           60 artifacts  runtime_fill=f  -> EMIT
game-master-explanation  32 artifacts  runtime_fill=f  -> EMIT
platform-comparison      30 artifacts  runtime_fill=f  -> EMIT
lobby-grid               37 artifacts  runtime_fill=t  -> finding only
provocation-card         13 artifacts  runtime_fill=t  -> finding only
```

Fleet-wide the check now emits 8 regenerations (vonc 3, finetuning.uk 2, leopardessconsulting 2,
robot-hands 1) and skips 2. The guard is **not** over-broad: `provocations-archive-list` is
runtime-fill *and* healthy (8 real slots, 0 `<no value>`), so it never matched the predicate at all —
only components that are **both** runtime-fill **and** rendered-artifact-shaped are affected.

Needs a chassis rebuild + push to take effect. Neither check is enabled, so there is no urgency —
but **both must ship before the improvement loop is switched on.**
*(Superseded later the same day: both guards shipped in `v1.0.1103` — see the closing entry below.)*

`Categories:` (fix, root-cause, decision)

---

## 2026-07-10 — generalisation pass: new check + new fixer; thread CLOSED

Answered "how do we make these fixes fleet-generic rather than vonc-specific". Full classification of all
thirteen findings in `PLAN_generalise_fixes_to_fleet.md` §3; operational position in
`RUNBOOK_minilobby_task.md` §0. **Of thirteen findings exactly one is site-specific** (the dead `lobby` key in
`provocations.json`). The mini-lobby trim was a keyhole onto fleet-wide defects.

**Four rules extracted** (PLAN §2): fix the writer not the row; detect by contract not by name; surface never
silently skip; every detection needs a fixer or a written reason there is none.

**New detection — `check_page_component_status_drift.go`.** `page_components.build_status` has **no CHECK
constraint** (verified: the only constraint on the table is on `lock_type`). It is free text, and essentially
every discovery check filters `pc.build_status = 'deployed'` — so any other value removes a live section from
the entire audit surface. That is precisely how `apply_section_edit`'s `'approved'` hid provocation-card. The
check finds page_components on a **deployed** page whose status isn't `deployed`, and splits them:

- status outside the known vocabulary (`deployed`/`pending`/`removed`/`needs_rebuild`) → **emit** work item
  `page_component_status_drift` → `component-template-fixer`. **0 rows today** — a pure regression guard.
- `pending` on a deployed page → **finding only**. 19 such rows across 5 sites, all with `open_items = 0`:
  invisible to every check and with nothing scheduled to fix them. Their repair is a rebuild, not a status
  flip — flipping them to `deployed` would *hide* real staleness. Surfaced, not acted on.

Proved it catches the original bug with a read-only CTE simulating `provocation-card` back at `'approved'`:
one item emitted, routed correctly, `has_html = t`.

**New repair — `repair_page_component_status`** fix_type in `fix_component_template_action.go`, sitting beside
`align_slot_name` (the same class: page_component *metadata* repair, never `rendered_html`). Two refusal guards:
the parent page must itself be `deployed`, and the component must carry non-empty `rendered_html` (positive
evidence, mirroring `pageHasComponents`). It refuses to flip `pending`/`removed`/`needs_rebuild` — those are
honest states. Routing verified: `handler_agent` dispatches freely (no item_type allowlist) and the fixer's
`apply_fix` step resolves `fix_type` from `input_data.spec.fix_type`, which the check supplies.

**Also corrected** `fix_component_template`'s header, which claimed `remove_element` defers page-component
content to the section editor — implying it touches page components at all. It operates on `site_components`
only. That sentence is what mis-planned this entire task on 2026-07-09; the header now groups every fix_type by
the table it actually writes.

**Two things found but not acted on.** (a) Eight checks are registered in Go and enabled in no agent —
including `sectionless_pages`, the exact detector for the ten silent `complete_error` builds. (b) Three
*enabled* names have no Go implementation (`missing_content`, `orphan_nav`, `stale_pages`, all in
`maintenance-triage`); the runner does `logger.Warn("Unknown discovery check — not registered")`, so this is
noisy rather than silent — one fewer defect than feared. Enable decision in PLAN §5.

**Proposed, not applied:** a CHECK constraint on `page_components.build_status` (PLAN §4). It is the root
enabler of the whole class — it would have turned a silent invisibility into a loud write failure.

All four Go files build and vet clean. None is enabled, so the push is safe in isolation.

`Categories:` (milestone, decision, next-task)

---

## 2026-07-10 (later) — v1.0.1103 shipped the guards + fixer mid-session; position reconciled

While the generalisation pass was being written, the user committed and deployed chassis **`v1.0.1103`**
(commit `49d67e82`; the `component_template_corrupted` check itself had arrived in `21130d94`). Verified on
the running pod:

- **In the deployed binary:** both runtime-fill guards (`check_component_template_corrupted.go`,
  `check_component_standards.go` / `broken_template_slots`) and the `repair_page_component_status` fixer with
  the corrected `fix_component_template` header.
- **Safety confirmed:** `component_template_corrupted` / `validate_component_standards` are enabled in **no**
  agent's `checks` array, and no `needs_component_regeneration` item has ever been raised against
  `provocation-card` or `lobby-grid`. The guard landed before the check was ever switched on — the ordering
  was luck, not design; the check had been sitting loaded in the same working tree.
- **Still NOT shipped:** `check_page_component_status_drift.go` remains **untracked** — it needs a commit and
  the next chassis push before it can be enabled.

Current position is maintained in `RUNBOOK_minilobby_task.md` §0; the fleet classification and open decisions
(enable list, CHECK constraint) in `PLAN_generalise_fixes_to_fleet.md` §3–§5. This log is chronological and
closed at this entry unless the thread reopens.

`Categories:` (milestone)

---

## 2026-07-10 (evening) — v1.0.1104 deployed: ALL code from this thread is now in production

Chassis `v1.0.1104` deployed. Learned in passing: **builds are from the local filesystem via the Makefile;
commits are decoupled** (user's discretion; image tag now hand-recorded in commit messages) — so deployed
contents were verified against the running pod, not git history:

- binary strings: `page_component_status_drift` ×7, `repair_page_component_status` ×4, guard log line ×1 ✅
- RUNBOOK §3 query 1: **0** components on deployed pages with an unknown status (regression guard quiet) ✅
- RUNBOOK §3 query 2: vonc guard split 3 emit / 2 skip, unchanged ✅
- template checks still enabled nowhere; 0 regeneration items against the shells ✅

Every check, fixer, guard and writer-fix this thread produced is now live. The one remaining action is the
PLAN §5 enable decision (which checks to switch on, in which discovery agent). The proposed `build_status`
CHECK constraint (PLAN §4) also remains open.

`Categories:` (milestone)

---

## 2026-07-10 (evening, cont.) — three checks ENABLED; first pass ran; a THIRD unguarded check caught live

User approved enabling `page_component_status_drift`, `sectionless_pages` and `component_template_corrupted`.
All three added to **completeness-discovery-agent**'s checks array (backup
`_completeness_agentdef_backup_20260710`). Context: the `improvement-sweep` scheduled task (→ improvement-loop,
180 s) has been **disabled since 2026-05-02**, so discovery only runs when triggered by hand
(`scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh <domain> completeness`).

**First pass on vonc.com** (correlation `e0207aad-…`), results by check:

- `component_template_corrupted` — **the guard's first live firing, and it worked**: emitted exactly 3
  regeneration items (archetype-grid, game-master-explanation, platform-comparison), refused both runtime-fill
  shells.
- `page_component_status_drift` — 0 emissions, as designed.
- `sectionless_pages` — 0 (vonc has none).
- **`empty_sections` — a THIRD unguarded check, caught in the act.** Enabled all along, it flagged
  provocation-card and lobby-grid as `empty_heading` (`<h2 class="pc-headline"></h2>` is the shell) and raised
  `empty_section` items routed to **page-build-handler** — a full LLM rebuild that would bake copy into the
  shells. This check never fired on them before because until yesterday the shells' pages predated the check's
  enablement/the sections carried enough incidental text… no: it simply had not been run against vonc since the
  shells went in (discovery has been manual-only since May).

**Containment, same hour:** the two shell `empty_section` items set to `rejected` with an explanatory `error`;
the same runtime-fill guard (emission skip + finding, identical pattern) added to `check_empty_sections.go`;
builds/vets clean. **Not yet in the deployed binary.**

**Dedup caveat learned:** `idx_swi_dedup` is partial — it only covers non-terminal rows, and the two-strike
rule only counts `complete`/`failed`. So a completeness pass on the current binary **will re-raise** the two
rejected shell items. Do not re-run completeness discovery on vonc until the guarded build ships (or re-reject
after).

Legitimate items now open from the pass: 3 × `needs_component_regeneration` + 3 × `empty_section` (the three
genuinely corrupted components, on their own pages) + 1 × `needs_rerender` (missing_structure; harmless —
the rerender path is assemble-only and carries the runtime-fill exemption). Nothing dispatches them while the
improvement loop stays off.

Score for the day: the runtime-fill landmine existed in **three** checks; two were guarded pre-emptively by
code reading, the third was caught by running the loop against the live site within minutes of enabling it.

`Categories:` (milestone, root-cause, fix)

---

## 2026-07-10 (night) — v1.0.1105: third guard PROVEN; thread engineering COMPLETE

Chassis `v1.0.1105` deployed carrying the `check_empty_sections` guard (string-verified in the pod binary).
Re-ran completeness discovery on vonc (correlation `a62610c3-…`) — deliberately, because the two rejected
shell items sit **outside** the partial dedup index: on the old binary this pass would have re-raised them,
so their absence is positive proof, not a dedup artifact.

Result: **zero new work items**; the only shell `empty_section` rows remain the two `rejected` ones from the
first pass. The spawned pod logged all four exemptions by name:

```
empty_sections: runtime-fill shell exempt from rebuild              | provocation-card
empty_sections: runtime-fill shell exempt from rebuild              | lobby-grid
component_template_corrupted: runtime-fill shell not auto-regenerable | lobby-grid
component_template_corrupted: runtime-fill shell not auto-regenerable | provocation-card
```

The caution against re-running completeness discovery on vonc is **lifted**. Everything this thread produced —
section-editor writer fix, trigger drop, three runtime-fill guards, drift check, status fixer — is deployed
and artifact-verified. Remaining items are operational, tracked in `RUNBOOK_minilobby_task.md` §0.

`Categories:` (milestone)

---

## 2026-07-09 — the dropped trigger, for the record (verbatim from the session Q&A)

`auto_lock_on_deploy` — a Postgres trigger on the `page_components` table (trigger name
`trigger_auto_lock_on_deploy`, plus its backing function `auto_lock_on_deploy()`), dropped from the
live `clients_db` earlier this session, with your approval via the question I asked.

**What it did:** `BEFORE UPDATE` on `page_components` — whenever a row's `build_status` changed *to*
`'deployed'`, and the owning site had `strict_mode_trigger = 'first_deploy'`, it stamped that row
`schema_mode = 'strict'`, `locked_at = now()`, `locked_by = 'first_deploy'`.

**Why it went:** the strict-mode subsystem it belonged to was stillborn —

- No Go code reads `schema_mode` or `strict_mode_trigger` (the former exists only as a struct field,
  never consulted).
- The `schema_snapshot` / `content_snapshot` columns its own design (file
  `008_page_components_and_schema_mode.sql`) says should accompany the lock were **never created** in
  production — the sibling functions `lock_section_to_strict` / `unlock_section_for_redesign` would
  error if ever called.
- It had fired **exactly once** in the system's history (a gaswholesalers.com tool page, 2026-04-03),
  producing one `strict` row that nothing consumed.

**Why it mattered right then:** the section-editor fix we shipped makes `UpdatePageStatusAction`
UPDATE `page_components.build_status` to `'deployed'` — exactly the transition that fires this
trigger. Left in place, every section-edit would have locked its row to `strict`, making edited
sections the only locked rows anywhere, for a feature nothing honours. (It also nearly bit me
manually: my first `approved → deployed` repair had to be a two-statement transaction specifically to
undo the lock the trigger applied.)

**Reversibility:** the exact function body is saved at
`docs/social001_vonc_tiktok_social/minilobby_task/auto_lock_on_deploy.FUNCTION_BACKUP.sql`, and the
drop itself is a documented migration at
`docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql` — re-running the backup file plus
one `CREATE TRIGGER` restores it. The drop also normalised that single legacy gaswholesalers row back
to `schema_mode = NULL`.

What I deliberately *didn't* drop: the orphaned `lock_section_to_strict` /
`unlock_section_for_redesign` functions and the `schema_mode` / `strict_mode_trigger` columns —
harmless, and column drops are a bigger decision than a trigger drop.

`Categories:` (decision, fix, root-cause)
