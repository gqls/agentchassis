# Running Notes — guide-skinner-box build + adoption "sectionless page" root

**Session date:** 2026-06-08 (continues HANDOFF_2026-06-06)
**Site:** gamesdesign.co.uk (adoption clone of gamedesign.uk). Go agent-chassis, Kafka saga agents, Postgres `clients_db`.
**Scope:** Close the open `guide-skinner-box` build bug (resolve the readiness fork left open in the handoff, apply + verify the unblock), then trace the same failure to its adoption-general root (a page reaching build with zero sections) and decide the structural fix — without shipping code until the root is confirmed.

This is a thinking/decisions log — dead-ends and corrections included — not a clean writeup. Durable technical findings belong in `016_debugging_guide`; this captures *how/why*.

---

## Part 1 — Resolving the readiness fork (the handoff's open question)

**The fork (from HANDOFF_2026-06-06 / CONTEXT_PACK):** the prescribed fix wrote BOTH `site_plan_sections` rows AND `pages.sections`. It was unknown which (if either) the build actually reads for readiness. The handoff said: if the page still completes ~90s/empty after the rows are present, then `site_plan_sections`+`pages.sections` are NOT the source `plan_sections` reads, and the next root is to pull the actual `load_page_sections_from_spec` + `plan_sections` code. That code "was not on disk last session."

**Found it.** The code is in `production_page-build-debug_context.txt` (the chassis dump), not the loose `.go` files. Read end-to-end:

- **`page-build-handler` workflow** (from `agent_definitions`): `ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections → plan_sections → check_has_ready_sections → spawn_content_writer → call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender_agent → deploy_page → complete`.
  - `load_spec_sections` = action `load_page_sections_from_spec`, config `{site_id, page_name, page_sections_fallback: page_record.sections}`, output `spec_sections`.
  - `plan_sections` config `{sections: spec_sections.sections, …}`, output `section_plan`.
  - `check_has_ready_sections` condition `section_plan.ready_count > 0` → ELSE `complete_error`.
  - `complete_error` is a `complete_workflow` (SUCCESS) with message *"Content writer skipped — page has no sections defined."* — the silent-success smell.

- **`load_page_sections_from_spec_action.go`** — the decisive one. In order:
  1. Query `site_specs` aspect **`site_plan`**, `is_current=true`. For gamesdesign.co.uk that aspect **does not exist** (aspect list: briefing, classification, content_direction, design_intent, design_reference, identity, resolved_composition, site_archetype, strategy, structure). So `planDataJSON` is NULL → `specSections` stays empty. When it IS present it also *syncs* the result into `pages.sections`.
  2. **Fallback to `pages.sections`** — first via `page_sections_fallback` (= `page_record.sections`), then a direct `SELECT sections FROM pages WHERE site_id=$1 AND name=$2`.
  3. Still empty → returns `{sections: [], count: 0, source: "none"}`.

- **`plan_sections_action.go`** — `ready_count = len(ready)`. Each section name is resolved: path-1 direct function/name lookup (`loadComponentSchemas` → `planSection`), path-2 section_type selector, path-3 not-found → `needs_new_component`. Empty `sectionNames` → early return `ready_count: 0, reason: "no sections to plan"`.

**Fork resolved — decisively.** The field the build gates on is **`pages.sections`** (read by `load_page_sections_from_spec`'s fallback, since the `site_plan` aspect is absent here). **`site_plan_sections` is read nowhere on the build path.** So in the prescribed SQL, the load-bearing statement is the `pages.sections` UPDATE; the `site_plan_sections` INSERT is *plan hygiene*, not what unblocks the build. `hero` + `generic-text-block` resolve via path-1 (they're the exact pair the working sibling `guide-rng-design` builds with), so they go `ready`, `ready_count=2`, and the content writer is spawned.

**Corrections to the handoff's mental model (logged):**
- The handoff's causal story — *"the content writer died on claim-timeout → content never written → no sections"* — is the **wrong mechanism**. The content writer writes `page_components` (via `save_page_sections`); it does **not** write `site_plan_sections` or `pages.sections`. Those are populated upstream (plan + sync). The dead `needs_content_page` was a *downstream symptom*, not the cause of the missing sections.
- `site_plan_sections` being load-bearing for readiness: **false**. It is not on the build read path.

---

## Part 2 — The immediate fix (applied + verified)

**Verification query first (decisive comparison):**
```
 page_name        | section_rows | comps
 guide-rng-design |            2 | {hero,generic-text-block}
 (guide-skinner-box: no row)
```
Confirms skinner-box has **zero** `site_plan_sections` rows in the current plan while the sibling has two → the gap is at **plan emission**, not content-write time.

**Applied (snapshots first; site_id resolved via domain subquery per standing rule):**
- `pages_bak_skbox` (page row), `sps_bak_skbox` (section rows) snapshots — `sps_bak_skbox` returned `SELECT 0`, independently confirming zero section rows.
- `(A)` `UPDATE pages SET sections='["hero","generic-text-block"]'` → `UPDATE 1` — the load-bearing fix.
- `(B)` INSERT 2 `site_plan_sections` rows mirroring rng-design → `INSERT 0 2` — plan hygiene.
- `(C)` re-issue `needs_content_page` (`mode=recreate, source=adoption`), item_key `…-retry2` → `INSERT 0 1`.

**Awaiting verification (the confirmation, not a fork):** retry item `triaged → claimed`; runs **well past 90s** (writer ~1200s; >90s ⇒ it spawned, since the empty path completes in ~90s); `page_components` non-empty; `pages.build_status='deployed'`; skinner-box card gains its description. If it *still* completes ~90s/empty with `pages.sections` now set, that contradicts the code path above and we pull the live `agent_error_log` for that item — but the code says it won't.

**Durability caveat (honest):** the build path reads `pages.sections`. `upsertPage` (sync) overwrites `pages.sections` from the plan-page's `sections` array on `ON CONFLICT … sections = EXCLUDED.sections`, defaulting to `'[]'` when absent. The sync reads the plan from `extractPagesFromPlan(CollectedData)`, **not** from `site_plan_sections`. So this manual fix holds for the current build but a future full re-plan / re-sync that re-derives skinner-box without sections would clobber it again. The durable fix is getting sections into the plan emission itself — see Part 4.

---

## Part 3 — Tracing the adoption-general root (where sections come from)

**Two writers populate the two tables; neither is the content writer.**
- `write_site_plan_action.go` (plan-builder terminal step) writes `site_plans` + `site_plan_pages` + **`site_plan_sections`** (per section within each page) from the LLM plan output, in one transaction. It does **not** emit work items (reconciler's job).
- `upsertPage` (sync, `site_db_actions.go`) writes the `pages` row including **`pages.sections`** from `page["sections"]` extracted from the in-memory plan in `CollectedData`. Empty/absent → `'[]'`.

**So the chain that produced the bug:** the plan for skinner-box carried **no sections** → `write_site_plan` wrote zero `site_plan_sections` rows AND the synced `pages.sections` became `'[]'` → page-build had no source → `ready_count=0` → silent `complete_error`. The 5 guides exist as `role=guide` (post-convergence), but at least skinner-box went into the plan as a **bare page** (no per-page sections).

**The decisive unknown (must confirm before any structural code):** *why* did the plan emit sections for the other guides but not skinner-box? Two live hypotheses, not yet separated:
- **H1 — transient/truncation:** the LLM plan output dropped sections for skinner-box specifically (one-off). Predicts: skinner-box uniquely zero among the 5 guides.
- **H2 — systematic path:** skinner-box entered the plan via a path that doesn't carry sections (e.g. promoted from the flat `structure.data.pages` name list, or an adoption page-discovery path), distinct from the LLM pages-with-sections path. Predicts: any guide added via that same path is also zero.

**Discriminating check (cheap, decides H1 vs H2):** count `section_rows` for **all** guide pages in the current plan, not just the two. If only skinner-box is zero → H1 (transient; the immediate fix + a convergence guard suffices). If a cluster is zero → H2 (a plan-emission path needs fixing at source).

---

## Part 4 — Structural options (noted; not yet chosen — pending Part 3 confirmation)

Framed against the guidelines (001 STEP ZERO reuse-first; 002 discovery→work-item→reconcile feedback loop; the FOCUS silent-completion doc). The unifying theme remains the handoff's: **"complete" is being used to mean "we stopped," not "the work succeeded."**

**Option S1 — Reuse the discovery-check pattern for a "sectionless planned page" guard.**
A page in the current plan with zero `site_plan_sections` (and/or empty `pages.sections`) is exactly an *algorithmic binary SQL check* — the shape the existing `DiscoveryCheck` registry (`discovery_checks/`) already handles (cf. `unlinked_site_components`). It would emit a work item into the same dispatch loop rather than introducing a new agent. Open sub-decision: what the guard *emits* — re-plan the page (preferred, addresses root) vs. re-issue `needs_content_page` with a default section set (what we did by hand). Reuse-first says: a check that flags + re-triggers, not bespoke code.

**Option S2 — Fix the silent-success at `check_has_ready_sections`.**
`complete_error` is a SUCCESS-labelled `complete_workflow`. A sectionless page should NOT terminate as `complete`. Per the FOCUS doc, route to a non-terminal/flagged state (`needs_investigation` / `needs_human_review`) so the sectionless page is visible and retriable rather than silently "done." This is the minimal correctness fix and stops new residue accruing while S1/source fix lands. (Workflow-def change, thin; no new Go.)

**Option S3 — Source fix at plan emission (only if H2).**
If a plan-emission path adds sectionless guides, fix that path so every page it adds carries sections (or is explicitly deferred), so re-plans/re-syncs are self-healing. Scope depends on which path — unknown until Part 3.

**Companion (cross-cutting, from FOCUS doc):** positive-evidence completion — small reusable helpers `pageHasComponents(ctx, db, pageID)` / `pageIsDeployed(ctx, db, pageID)`; complete only on explicit success OR positive DB evidence. This is the coherent mini-project the silent-completion modes (claim-timeout→complete, validate routing, reaper auto-complete) all belong to. Deliberately a *set*, not piecemeal.

**Leaning (provisional):** S2 (correctness, stops the bleed) + S1 (reuse the discovery loop to catch + retrigger the residue), with S3 only if Part 3 shows H2. No code until Part 3 confirms the root and we've run STEP ZERO (search `agent_definitions` / registry / discovery_checks for any existing "sectionless"/"empty page"/"missing sections" check before writing one).

---

## Part 5 — Convergence trace: the writers are correct; the gap is reconciliation

Verified result: skinner-box is built — `page_components`=2 (both with `component_id`), `build_status='deployed'`, `pages.sections=["hero","generic-text-block"]`. Positive evidence, not just a `complete` flag (the work item does show `complete`/no-error, which on its own is the ambiguous signal — the components + deployed are what confirm it).

Guide section-row census (post-fix, so skinner-box now shows 2 from our INSERT): all 5 guides at 2. Combined with the pre-fix snapshot (`sps_bak_skbox` = 0) and the earlier comparison (skinner-box absent, rng-design 2), the pre-fix state was: **4 guides with 2 sections, skinner-box alone at 0.** → **single-page residue, not a guide-category systematic bug.** H2-as-a-category is ruled out; **S3 (source fix for a whole emission path) is not indicated.**

Traced the one path that adds a page the LLM plan omitted — adoption-faithfulness convergence (`reconcilePlanWithRealised`, FOCUS_adoption_faithfulness_via_locks):
- Pass A unions locked realised pages missing from the LLM plan, via `normaliseRealisedToPlanPage(rm)`.
- `load_existing_pages` (line 29896) selects `sections::text`, so the realised map carries sections.
- **`normaliseRealisedToPlanPage` carries `sections` through** (37770–37800) — and its comment shows this was a deliberate fix: without it the union emits empty values and the sync's `<col> = EXCLUDED.<col>` clobbers the adopted page's real sections. So the union is **correct** and faithfully carries whatever `pages.sections` held.

**Conclusion (grounded, not assumed):** every page-emission writer is individually correct — `write_site_plan` (sections from LLM plan), `upsertPage` (sections from plan page, EXCLUDED-overwrite), and the convergence union (`normaliseRealisedToPlanPage`, carries sections). A page reaches build sectionless only when the value it faithfully carries is *already empty* — e.g. a page planned-but-never-built (so `pages.sections` was never populated), or one clobbered to `[]` by a pre-`normaliseRealisedToPlanPage`-fix build and then unioned. The exact sub-cause for skinner-box can't be fully reconstructed (we overwrote `sps_bak_skbox`'s zero state and the plan-version history isn't retained), but it doesn't change the fix surface.

**Therefore — corrected structural direction:**
- **Do NOT patch `normaliseRealisedToPlanPage` / the convergence union.** It is correct; patching it to "manufacture" sections would be a wrong fix (it has nothing to carry). Checking this before prescribing avoided that.
- The real gap is **reconciliation**: nothing repairs a page that is in the current plan with zero sections, and page-build then **silently completes it as success**. That is the durable bug, and it is the same family as the empty-guide-list and render short-circuit from the handoff.

**Revised option set (supersedes Part 4 leaning):**
- **S2 (correctness, minimal): `check_has_ready_sections` ELSE must not be a SUCCESS-labelled `complete_workflow`.** A sectionless page routes to a flagged/non-terminal state (`needs_human_review` reusing existing HITL machinery, or a `needs_investigation` class) — never `complete`. Thin workflow-def change; no new Go. Stops new residue going silently green.
- **S1 (recovery, reuse): a discovery check "planned page with zero sections" that flags + re-triggers**, emitting into the existing dispatch loop (reuse the `DiscoveryCheck` registry pattern per 002 — algorithmic binary SQL, the same shape as `unlinked_site_components`). Open sub-decision: re-trigger as `needs_content_page` with a role-default section set (what we did by hand for skinner-box), vs. re-plan the page. The role-default re-trigger is the smaller, reuse-first move and matches how the working guide siblings are shaped.
- **S3: not indicated** (no systematic emission path; ruled out by the census).
- **Companion (separate mini-project, FOCUS doc): positive-evidence completion** — `pageHasComponents` / `pageIsDeployed` helpers; complete only on explicit success OR positive DB evidence; fix the three silent-completion modes as a set. Tracked, not bundled into this task.

## Part 6 — STEP ZERO (reuse search) before any S1/S2 code

Searched the chassis dump (`production_page-build-debug_context.txt`): action registry, the `discovery_checks/` directory (the `DiscoveryCheck` registry pattern from 002), and adjacent section/page actions. Terms: section, sections, empty, empty page, missing, no/zero sections, planned page, orphan, unresolved, content_page, gap, structure.

**Existing surfaces found (the catalogue):**
- **`checkEmptyPageSections`** — sub-check inside `ComponentStandardsCheck` (`Name()="validate_component_standards"`, registered via `init()`). Query: pages with `build_status IN ('deployed','active')` and **zero** rendered `page_components` (`HAVING COUNT(pc.id)=0`). Emits `item_type='needs_content_page'`, `check='empty_page_sections'`, `handler_agent='page-content-writer'`, dedup `empty_page_<name>_<site>`. **This is the closest existing thing to S1 — it already targets "page with no rendered sections."** Two gaps explain why it never caught/repaired skinner-box:
  1. **Scope:** `build_status IN ('deployed','active')` only. Skinner-box sat at `planned` (build short-circuited before deploy) → invisible to it.
  2. **Recovery:** emits `needs_content_page` with **no sections in the spec** → page-build re-reads empty `pages.sections` → `ready_count=0` → silent complete. It would loop, not recover. (Routes to `page-content-writer`, not `page-build-handler` — possibly stale; either way the readiness short-circuit is the same.)
- **`EmptySectionsCheck`** (`empty_sections`) — pages whose `page_components` exist but have empty/near-empty `rendered_html`; requires component rows + `build_status='deployed'`. Doesn't see a zero-component page. Not a match.
- **`UnresolvedSectionsCheck`** (`unresolved_sections`) — iterates `jsonb_array_elements_text(p.sections)`; requires **non-empty** `sections` + `deployed`. Skips zero-section pages. Not a match.
- **`OrphanPagesCheck`**, **`MissingStructureCheck`** (named adjacent: nav-orphan / site-level structure-spec, not per-page sections). **`ReconcileSitePlanAction`** (`reconcile_site_plan`) — emits `needs_page` on plan↔realised drift, keyed at page level, not section level.
- **`load_page_sections_from_spec`** — the read-time resolver with the existing fallback chain (`site_specs.site_plan` → `pages.sections` → empty). Natural extension point: a final fallback.
- **`reconcile_section_data_action.go`** — handles `needs_section_data` (deferred, query-resolvable) only; not zero-section.

**0f — decision (reuse over create; no new agent, no new check struct):**
- **S1 = patch the existing `checkEmptyPageSections`**, not a new check: widen its predicate to also flag a *planned* page that is definitively unbuildable (empty `pages.sections`), and enrich the emitted item so the handler can repair rather than loop.
- **Recovery = patch the existing `load_page_sections_from_spec`** (small extension, mirrors the guide's "extend, don't recreate" examples): add a final fallback — when `site_specs.site_plan` and `pages.sections` are both empty, derive the section list from a **same-role sibling** in the current plan (faithful, data-driven, e.g. a guide copies another guide's `hero`+`generic-text-block`) and write it to `pages.sections`. This dissolves the dead-end at the exact point the build reads, so a sectionless page self-heals on the next build attempt.
- **S2 = patch the page-build-handler workflow** `check_has_ready_sections` ELSE so the residual genuinely-zero case (no sibling/role default exists) routes to a flagged/non-terminal state — never a success-labelled `complete_workflow`.
- All three are edits to existing surfaces. No new agent, no new action, no new discovery check.

**Live confirmations still owed (SQL for the user — 0a/0e + site enablement):**
```sql
-- 0a: any existing agent for sectionless/empty/discovery work
SELECT type, display_name, status, agent_category, substring(description,1,120) AS d
FROM agent_definitions
WHERE deleted_at IS NULL AND (type ILIKE '%section%' OR type ILIKE '%discover%'
  OR type ILIKE '%content%' OR description ILIKE '%empty%' OR description ILIKE '%section%')
ORDER BY type;

-- 0e: which agents wire the relevant actions/checks
SELECT type FROM agent_definitions WHERE deleted_at IS NULL
  AND (default_config::text ILIKE '%run_discovery_checks%'
    OR default_config::text ILIKE '%validate_component_standards%'
    OR default_config::text ILIKE '%empty_page_sections%');

-- Site enablement + history: is the empty_page_sections check even running here,
-- and has it ever fired for this site? (explains why skinner-box was never auto-caught)
SELECT item_type, status, count(*), max(created_at)
FROM site_work_items
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND (item_type IN ('needs_content_page','empty_section')
    OR spec->>'check' IN ('empty_page_sections','empty_sections','unresolved_sections'))
GROUP BY item_type, status ORDER BY item_type, status;
```

**Live confirmation results (user-run):**
- **0a:** the owning agent is **`completeness-discovery-agent`** (active, analyst) — description explicitly includes "empty sections." So `validate_component_standards`/`checkEmptyPageSections` is a real, active surface; extending it reaches production. (`page-build-handler` confirmed as the wrapper around the `page-content-writer` specialist.)
- **0e:** three discovery agents run checks (`completeness-`, `design-`, `quality-discovery-agent`); the OR-match doesn't say which enables `validate_component_standards` by name — resolved by the `checks`-array query below.
- **History:** **11** `needs_content_page` items, **all `complete`** (max = our 17:46 retry2). Ambiguous: the grouping doesn't show `spec->>'check'`, so we can't yet tell which (if any) came from `checkEmptyPageSections` vs manual rebuilds vs other. This distinguishes "check has a scope gap" from "check not enabled here" — different S1 edits.

**RESULT (decisive):** The 11 `needs_content_page` items are 9 adoption-run items (2026-06-05, one per page incl. skinner-box) + our 2 manual skinner-box rebuilds. **`check` is empty on all 11 → `checkEmptyPageSections` has never fired here.** Query (2): `validate_component_standards` (its wrapper) is **not enabled in any** discovery agent (`enables_vcs=f` for all three); only `empty_sections` (the empty-HTML check) is enabled, in `completeness-discovery-agent`. So the empty-page detector is **dormant code**, not a buggy check. Detection (S1) is therefore a separate, smaller wiring decision — and enabling the wrapper would also switch on its unrelated `checkUnwantedElements` sub-check, so it needs thought, not a toggle. **The recovery fix (sibling fallback in `load_page_sections_from_spec`) is the primary fix** — it removes the dead-end at read time regardless of trigger.

Note on "same-role sibling" (clarifying a question raised): a page's `sections` is the *layout skeleton* (component-type list, e.g. `["hero","generic-text-block"]`), not content. Same-role pages share that skeleton by design (all 5 guides use it). Content is written later by the content-writer from the page's *own* adoption crawl (recreate mode). So a sibling supplies layout only; no risk of borrowed text. Fallback must be last-resort, logged loudly, and absent a same-role sibling with sections → route to flagged state (S2), not silent complete.

**Queries that produced the above:**
```sql
-- (1) provenance of the 11 needs_content_page — did checkEmptyPageSections ever fire?
SELECT created_at, source, created_by, spec->>'check' AS check,
       spec->>'page_name' AS page, status, item_key
FROM site_work_items
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND item_type='needs_content_page'
ORDER BY created_at;

-- (2) which discovery agent's `checks` array enables the relevant checks
SELECT type,
       default_config::text ILIKE '%validate_component_standards%' AS enables_vcs,
       default_config::text ILIKE '%"empty_sections"%'            AS enables_empty_sections
FROM agent_definitions
WHERE type IN ('completeness-discovery-agent','design-discovery-agent','quality-discovery-agent')
  AND deleted_at IS NULL;
```
Mechanism confirmed in code: `run_discovery_checks` runs only the names listed in its step-config `"checks"` array (loops `registry.Get(name)`), so (2) is decisive for enablement.

## Part 7 — Recovery patch written (awaiting build/deploy): `load_page_sections_from_spec` sibling fallback

Patched the existing read-time resolver (no new action/agent). Added **fallback 2b** between the `pages.sections` fallback and the return: when both `site_specs.site_plan` and `pages.sections` are empty, query same-role siblings in the current plan (`site_plan_pages.role` + `site_plan_sections`), pick the **modal** layout (the ordered `component_name` list shared by the most siblings; deterministic tie-break), copy it onto the page, and persist to `pages.sections` via the same guarded `IS DISTINCT FROM` UPDATE the action already uses for the site_specs sync.

Why a sibling is valid: `sections` is a layout skeleton (component-type list), not content; same-role pages share a layout by design; per-page content is still written later by the content-writer from the page's own source. So only the skeleton is borrowed.

Design choices: modal layout (one odd sibling can't skew it); WARN log on every synthesis (named page + sibling — greppable, never silent); writes only `pages.sections` (the build-read field, same field step 1 syncs) — deliberately does NOT touch `site_plan_sections` (keeps the action's responsibility narrow); no usable sibling → unchanged, returns `source:"none"` (the case S2 will later flag).

Conformance: no new imports (reuses context/json/fmt/uuid/datahelpers/zap); no var/field renames; new `source` value `"same_role_sibling"` only (nothing downstream reads `source` for control — `plan_sections` consumes `sections`); no `logger.Debug`. Brace/paren/bracket balance verified (61/61, 83/83, 42/42).

**Before build:** confirm fresh `\d site_plan_pages` (has `role`) and `\d site_plan_sections` (`page_name, ordering, component_name`) — SQL is embedded in Go, written off the dump not a live schema.
**Verify after deploy:** blank a test guide's `pages.sections` (snapshot first), queue `needs_content_page`, watch logs for the `SYNTHESISED layout from same-role sibling` WARN + the page building normally.
File: `load_page_sections_from_spec_action.go` (outputs) → place at `./platform/orchestration/actions/`.

## Part 8 — S1 written: durability via detection + retrigger (`check_sectionless_pages`)

Durability question resolved: 2b only helps a page that *gets a build attempt*. The failure that created skinner-box was a build that **died** (claim-timeout, marked complete) and was never retried — 2b can't reach that. So S1 (detect + retrigger) is needed for "doesn't happen again," and it's the piece that closes the never-retried gap.

Key enabling discovery (`insertWorkItem`, line 46319): a built-in **two-strike rule** — an `item_key` with ≥2 terminal attempts in 7 days is inserted as `unresolved` (visible), and anything <3h after a terminal item is suppressed; live dups blocked by `ON CONFLICT (site_id,item_key) WHERE status NOT IN terminal`. So S1 needs no anti-churn logic of its own and does **not** depend on S2 to avoid looping.

Wrote `discovery_checks/check_sectionless_pages.go` (new check, the framework's sanctioned extension = "add a file"; self-registers via `init()`). Chose a **dedicated** check over extending `checkEmptyPageSections` because the latter is dormant (not in any agent's `checks` array), bundled with an unrelated sub-check (`checkUnwantedElements`), scoped to `deployed/active` only, and routes to `page-content-writer` (which bypasses the section-loading step). The dedicated check:
- Detects current-plan pages with `pages.sections` NULL/`[]` **that have a same-role sibling with sections** — exactly the set 2b can repair, so detection+fallback is a closed, churn-free loop. No-sibling pages are out of scope (→ S2).
- Emits `needs_content_page` (recreate) to **`page-build-handler`** (the workflow that runs `load_spec_sections`→`plan_sections` where 2b lives), item_key `sectionless_page:<name>:<site>`, status `detected` (discovery convention).
- Conformance: imports match `check_empty_sections.go` (json/fmt/uuid/zap); no new struct in registry; brace/paren/bracket balance verified (22/22, 27/27, 5/5).

**Enablement** (config, not code): add `"sectionless_pages"` to completeness-discovery-agent `default_config {workflow,steps,run_checks,config,checks}`. Inspect-first (confirms step name `run_checks`), snapshot, append idempotently (`|| '["sectionless_pages"]'`, guarded by `? 'sectionless_pages'`). Full steps + verify + rollback in `RUNBOOK_section_sectionless_durability.md`.

**The durability stack (how it doesn't happen again):**
1. 2b (read-time) — a triggered build self-heals from a sibling. *(written)*
2. S1 (discovery) — finds stuck sectionless pages and retriggers; with 2b they build; two-strike auto-flags the unfixable. *(written)*
3. S2 (workflow) — no-sibling case → flagged state, not silent SUCCESS. *(next, optional for this gap; cleanliness)*
4. FOCUS positive-evidence completion + 3 silent-completion modes — stops residue being created at all. *(separate mini-project, `FOCUS_page_build_handler_silent_completion.md`)*

**FOCUS doc clarification (answering a question raised):** nothing missing — the doc is `FOCUS_page_build_handler_silent_completion.md`, already in the project; it's the home for layer-4 above, not a new doc I need.

## Part 9 — Re-verified `checkEmptyPageSections` claims; S2 written; roadmap

**Re-check (prompted by skepticism — "all four wrong at once sounds odd"). Read the actual code; all four hold, with evidence:**
- *Dormant:* `ComponentStandardsCheck.Run` (line 11850) DOES call `checkEmptyPageSections` — so it's wired in, not uncalled. But the wrapper `validate_component_standards` is in no agent's `checks` array (query 2: `enables_vcs=f` ×3) → it never runs. "Dormant" = wrapper not enabled.
- *Bundled:* understated — bundled with **six** other sub-checks under one registered check (unlinked components, slot mismatch, missing metadata, missing asset refs, nav layout, unwanted elements). Enabling the wrapper enables all.
- *Deployed-only:* line 12222 `build_status IN ('deployed','active')` — excludes `planned` (skinner-box's state).
- *Wrong handler:* CONFIRMED with direct evidence. `page-content-writer` def (line 51541) is a *task* specialist: iterates `input_data.section_plan.sections_ready` (caller must supply the plan), ends at `compile_page` → `page_content`; **no save_page_sections, no update_status, no deploy** — persistence/deploy live only in the page-build-handler wrapper. So routing a discovery item straight to it (no section_plan, no persistence) cannot build/deploy a page. Corroboration: sibling `check_empty_sections.go` carries `HandlerAgent: "page-build-handler" (was "page-content-writer")` — the same handler was fixed there but not here.
- **Unifying explanation:** `checkEmptyPageSections` is stale, never-enabled code **partially superseded** by `EmptySectionsCheck` (which got the handler fix + enablement). Coherent single cause, not four coincidences. Reinforces the dedicated-check decision (different detection condition + half-abandoned).

**Runbook:** the uploaded `RUNBOOK_phase_b_c_d_deploy_5__.md` is the **checkpoint/adapter (model-training)** runbook (launcher def, bundle.tar.gz, B2, model-trainer, resume, monitor) — a different subsystem. Page-build deploy steps don't belong there; kept in standalone `RUNBOOK_section_sectionless_durability.md` (now 2b + S1 + S2).

**S2 written** (page-build-handler workflow def; config, no Go). Reuses the existing `mark_needs_review` pattern (`fail_work_item` + `status_override`): added step `mark_no_sections` (status_override `needs_human_review`, clear message), repointed `check_has_ready_sections.else_step` from `complete_error` → `mark_no_sections`. After 2b+S1 the only page reaching `ready_count=0` is one with no same-role sibling; S2 flags it instead of the SUCCESS-labelled `complete_error`.
- `needs_human_review` is NOT in `workItemTerminalStatuses` (complete/failed/verified/rejected/wont_fix/unresolved) → non-terminal, so the flagged item's `item_key` still blocks the dedup ON CONFLICT → **S1 won't re-trigger a flagged page** (the flag holds the slot). No loop.
- **Caveat to verify:** does the flag stick, or does workflow completion / the reaper overwrite `needs_human_review` → `complete`? `mark_needs_review` relies on the same assumption; if S2's flag is clobbered, so is the existing validation-failure flag — that's a documented silent-completion mode, fix belongs with the FOCUS work.
- Value of S2 even though the two-strike rule already flags `unresolved` after 2 tries: S2 flags on the FIRST no-sibling attempt (no wasted retries, no false "complete" events).

**What then — the roadmap after S2:**
1. **FOCUS positive-evidence completion (the root, biggest remaining structural item).** `FOCUS_page_build_handler_silent_completion.md`: fix the three silent-completion modes (claim-timeout→complete, validate_content routing, reaper auto-complete) + add `pageHasComponents`/`pageIsDeployed` and complete only on explicit success OR positive DB evidence. This is what stops residue being *created*; once done, 2b/S1/S2 become rarely-fired safety nets. Also resolves S2's clobber caveat.
2. **Content quality (gamesdesign, high user-visible value).** Hero CTAs wrong site-wide (every hero → /contact.html + /services.html; /services.html phantom); guide copy tool-flavoured; polish batch (empty "Browse All" hrefs, brand-suffix in card titles, empty footer). Separate from build mechanics.
3. **Render-off-diff debt.** page-rerender skips planned-but-unrendered sections on a `deployed` page (current workaround: `build_status='needs_rebuild'` reset). Proper fix: drive render off a planned-vs-rendered `page_component` diff.
4. **Tools/games behavioural QA loop** (`PLAN_tools_games_behavioral_qa_loop.md`) — standalone mini-project.

Recommended next: (1) FOCUS positive-evidence — it's the source of this whole class and unblocks the S2 caveat.

## Part 10 — FOCUS silent-completion: inventory before code (modes re-checked against current code)

Started the FOCUS positive-evidence work by inventorying the actual completion paths (FOCUS doc's own step 1: "there may be more than three"; and the doc is April, code has moved since).

**Mode 2 (validate_content → complete) — appears ALREADY FIXED in current code.** `ValidatePageContentAction` returns `fmt.Errorf("content validation failed: %d blockers, %d errors")` on blockers (line 32818). The current page-build-handler def routes `validate_content` `error_step → mark_needs_review` → `fail_work_item` status_override `needs_human_review`. So a validation failure is flagged, not completed. (Only weakness noted in-code: thin error detail, line 32844.) To confirm: no OTHER agent/workflow still routes validate failures to complete.

**Mode 3 (claim-timeout → complete) — correct machinery already exists, reaper just bypasses it.** `FailWorkItemAction` default branch (line 46235): `attempt_count+1`, then `'failed'` if `>= max_attempts` else `'triaged'` (retry). That IS the correct claim-timeout behaviour. The claim-timeout reaper marks `complete` directly instead of routing through this. Fix = make the reaper reset via fail_work_item-style logic. Reuse, not new logic.

**Mode 1 (reaper auto-complete on lost response) — needs the reaper code.** Produces `"Auto-completed: work verified done despite lost response"`; completes on weak evidence. Fix = positive-evidence guard.

**Architectural finding / boundary:** work-item completion on success is driven by the **dispatch loop** reacting to the handler saga (the page-build-handler def's terminal steps use `complete_workflow` = saga end, NOT `complete_work_item`). The reaper(s) are a separate background process. **Neither the dispatch-loop completion logic nor the reaper is in `production_page-build-debug_context.txt`** — confirmed: the error strings appear only in data rows/comments; `dispatch_actions.go` has only `DispatchAgentAction`; `CompleteWorkItemAction` exists but the page-build workflow doesn't call it for terminal steps. So modes 1 & 3 (and the S2 clobber question) live in the **build-dispatch-loop** service, not on disk here.

**S2 clobber question resolves to the same code.** `fail_work_item` (status_override) sets `needs_human_review` and commits before `complete_workflow` runs. Whether the dispatch loop overwrites that on saga success decides if the S2 flag sticks. Since `mark_needs_review` is in production use, normal completion almost certainly preserves it; the real threat is the **reaper** (mode 1) auto-completing a flagged item after a lost response. So fixing mode 1 also protects S2.

**Decision: don't write the mode-1/3 patches or the positive-evidence helpers yet** — they must integrate with the dispatch-loop/reaper completion site, which isn't available. Requested from the user (fresh from chassis): the build-dispatch-loop code — (a) where it marks an item `complete` on handler saga response, and (b) the reaper(s) producing "Claim timed out…" (mode 3) and "Auto-completed…" (mode 1).

**Deliverable available now (read-only safety net; FOCUS step 5):** a positive-evidence monitor — `complete` items whose page has no `page_components` with non-null `component_id` and non-empty `rendered_html`. Catches residue from all three modes; complements S1 (which only repairs the sectionless subset).

## Part 11 — FOCUS modes mostly ALREADY FIXED; one real gap left (`complete_work_item` clobber)

User supplied the build-dispatch-loop def, the `claimed-item-timeout` reaper (scheduled_task pre_query), `claim_work_item_action.go`, `load_work_item_actions.go` (complete/fail), `validate_page_content.go`. Re-assessed each FOCUS mode against current code:

- **Mode 1 (reaper lost-response auto-complete): FIXED.** The `claimed-item-timeout` reaper now auto-completes ONLY with positive evidence — `completed_by_evidence` CTE requires `page_components` (`component_id` + non-empty `rendered_html` + `updated_at > claimed_at`) for `needs_content_page`, `pages.deployed_at > claimed_at` for `page_rerender`, head-slot update for `needs_design`. Exactly the FOCUS prescription. Comment cites a prior false-positive bug (gamesdesign homepage, build_status='deployed' with zero components) already hardened against.
- **Mode 3 (claim-timeout → complete): FIXED.** Reaper's `reset` CTE: stuck claimed items (>40 min, no evidence) → `attempt_count+1`, `triaged` (or `failed` at max). "Claim timed out — handler pod likely died" now accompanies a RESET, not a complete.
- **Mode 2 (validate_content → complete): FIXED.** `ValidatePageContentAction` returns `fmt.Errorf` on blockers (validate_page_content.go:272); page-build-handler routes `error_step → mark_needs_review` → `fail_work_item` status_override `needs_human_review`.
- **Positive-evidence helpers:** effectively implemented inline in the reaper SQL; no separate Go helper needed there.
- **Monitor query = 0 rows** → no silent-completion residue on the site now. Consistent with the above.

**So the FOCUS positive-evidence mini-project is ~done.** Was about to rebuild it — checking first avoided that (reuse/verify discipline).

**The one real remaining gap (surfaced this turn): `complete_work_item` is unguarded.** `CompleteWorkItemAction` (load_work_item_actions.go:751) does `UPDATE … SET status='complete' … WHERE id=$1` with NO status guard. The dispatch loop's `mark_complete` step runs it after a successful handler saga, and page-build-handler's `complete_error` is a SUCCESS-labelled `complete_workflow`. Consequences:
  - (a) **Clobber:** a handler-set `needs_human_review` (existing `mark_needs_review`, and S2's `mark_no_sections`) is overwritten back to `complete` by the dispatch loop. So both the existing HITL flag AND S2 may be ineffective under the dispatch loop. (Smoking gun: a `complete` item whose `error` still says "needs human review".)
  - (b) **Dispatch-level silent completion:** ANY page-build-handler path ending at `complete_error` (deploy fail, save fail, content-writer skip, no sections) returns saga-success → dispatch marks the item `complete`. The reaper can't catch these (item already `complete`, not `claimed`); only the monitor query can.

**Proposed Fix A (small, surgical, reuse):** guard `CompleteWorkItemAction`'s UPDATE so it does not overwrite a deliberate flagged/terminal status:
`… WHERE id=$1 AND status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')`, and return `completed = rows>0` (log when skipped). This is case (a) of the FOCUS rule (handler returned explicit success) — positive-evidence not needed here; just stop clobbering deliberate flags. Makes both `mark_needs_review` and S2 effective.
**Fix B (deeper, lower urgency given monitor=0):** `complete_error` shouldn't look like success for genuine-failure paths (deploy/save fail) — flag the item or stop success-labelling. Defer.

**Verify-first before writing Fix A (don't conclude the clobber from logic alone):**
```sql
-- smoking gun: handler-set needs_human_review overwritten to complete?
SELECT id, item_type, status, handled_by, LEFT(error,90) AS err, completed_at
FROM site_work_items WHERE error ILIKE '%needs human review%'
ORDER BY updated_at DESC LIMIT 50;

SELECT status, count(*) FROM site_work_items
WHERE error ILIKE '%needs human review%' OR error ILIKE '%no sections%'
GROUP BY status ORDER BY count DESC;
```
If flagged items show `status='complete'` → clobber confirmed → apply Fix A (and it's also required for S2 to work). If they show `status='needs_human_review'` → flag survives, mechanism differs, investigate before changing anything.

## Part 12 — Fix A written (`complete_work_item` guard); FOCUS doc updated

Smoking-gun queries returned **0 rows** — but that's "no test case," not "no clobber": `mark_needs_review` has never fired (zero validation failures), and the "no sections" message is a workflow `success_message` not stored in `site_work_items.error`, so neither query could have surfaced the relevant case. Wrong artifact chosen.

**Clobber confirmed by inference from data we already hold:** the 2026-06-06 skinner-box retry ran against a sectionless page → `check_has_ready_sections` else → `complete_error` → ended `status='complete'` (not failed/triaged) with zero `page_components`. That proves `complete_error → dispatch mark_complete` fires. Combined with `complete_work_item` being an unconditional `UPDATE … WHERE id=$1`, a flag set by `fail_work_item` immediately before `complete_error` is overwritten. So the clobber holds, and **Fix A is a prerequisite for S2** (and makes the existing `mark_needs_review` effective).

**Fix A written** (`load_work_item_actions.go` → `CompleteWorkItemAction`): `UPDATE … WHERE id=$1 AND status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')`; returns `completed = rows>0`; logs when skipped. Provably safe — no-op on the normal `claimed → complete` path (claimed isn't in the list), only preserves deliberate flags. No new imports; `_, err =` → `res, err :=` to read RowsAffected; balance verified (195/195, 246/246, 74/74). Dispatch loop's `mark_complete → done` is unconditional, so returning `completed:false` doesn't break the loop. This is the "handler returned explicit success" case → no positive-evidence needed here, just stop clobbering.

**FOCUS doc updated** (`FOCUS_page_build_handler_silent_completion.md`): status changed to "modes 1–3 resolved; one residual gap"; added a 2026-06-09 section documenting each mode's current fix (reaper positive-evidence + reset; validate→mark_needs_review; monitor=0), the residual `complete_work_item` gap, Fix A (applied), and Fix B (`complete_error` semantics, deferred). Original April characterisation retained below for history.

**Net state of the silent-completion work:** modes 1–3 already fixed in prod; Fix A closes the clobber + makes S2 viable; Fix B (complete_error-as-success on genuine-failure paths) deferred, low urgency while monitor=0. Definitive confirmation of the clobber available via controlled test (build a no-sibling sectionless page, watch the item land `needs_human_review` not `complete`) — runbook §4c.

## Part 13 — Handoff + updated collector for next task (content quality / hero CTAs)

Produced `HANDOFF_2026-06-09_sections_durability_and_content_quality.md` (state, deploy-pending 2b/S1/S2/FixA, next task = gamesdesign content quality led by site-wide hero CTAs, parked items, key refs).

Updated the collector → `package_content_quality_debug.sh` (was `package_page_build_debug.sh`): re-scoped to the CTA/content path (added `build_render_context`/`prepare_link_context`/`render_component`/`compile_page_sections`/`execute_llm_prompt` — names flagged "verify"; kept site_spec/validate/contentcreator/render/dispatch); **added the thunder-style DOCS_FILES copy block** (the page-build script lacked it) with find-by-basename fallback; re-aimed live SQL at hero CTA schema, `*_index_url` specs, phantom-`/services.html` detection, sample hero HTML, list-hub card titles/Browse-All hrefs, footer/contact, plus a slim durability-verify block (skinner-box built, monitor=0, needs_human_review items, sectionless_pages enabled). `bash -n` clean.

**Pending (the user's first step):** confirm the latest doc filenames/versions (`001/002/003`, `016_debugging_guide_v2_NN`, `CATALOGUE_gamesdesign_post_sync_fix_defects(N)`, component/hero docs, `019/020`). `DOCS_FILES` is a proposed set flagged for replacement once confirmed.

## Open / next
1. **skinner-box: closed** (verified, Part 5).
2. **Root: settled** — single-page sectionless residue; all page-emission writers correct (do not patch the convergence union). Durable gap = no reconciliation of a zero-section planned page + silent-success completion.
3. **STEP ZERO done (Part 6):** no new agent/check needed. Reuse surfaces identified — extend `checkEmptyPageSections` (S1), `load_page_sections_from_spec` (recovery fallback), and the page-build-handler workflow `check_has_ready_sections` (S2). **Decision point (user):** run the three live-confirmation queries above, then approve the three patches? No code until then.
4. Positive-evidence completion + the three silent-completion modes: separate mini-project (FOCUS doc), not bundled here.
5. Durability watch: the manual `pages.sections` for skinner-box survives until a future full re-plan/re-sync of the site (sync re-derives from the plan, not from `site_plan_sections`). S1 would make this self-healing.

## Key references (this session)
- `load_page_sections_from_spec_action.go`, `plan_sections_action.go`, `site_db_actions.go` (`upsertPage`), `write_site_plan_action.go` — all in `production_page-build-debug_context.txt`.
- page-build-handler workflow def (same dump).
- Guidelines: 001 (STEP ZERO, reuse, work-item lifecycle, guard-every-skip), 002 (QA Layer 0 = plan_sections, site lifecycle, discovery→reconcile loop), 003 (component v2 schema, content validation, orchestrator boundaries).
- `FOCUS_page_build_handler_silent_completion.md` (the three silent-completion modes + positive-evidence fix).
