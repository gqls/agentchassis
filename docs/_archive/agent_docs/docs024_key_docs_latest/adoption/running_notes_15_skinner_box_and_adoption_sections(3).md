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
