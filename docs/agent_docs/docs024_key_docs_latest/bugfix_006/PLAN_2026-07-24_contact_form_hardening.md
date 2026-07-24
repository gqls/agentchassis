# PLAN — standing protection against dead contact forms (bugfix 006 §B follow-on)

**Created 2026-07-24.** A decision brief: what it would take to make sure generated contact forms
never silently go dead again, and how to wire it into the platform's own checker / handler /
improvement-loop machinery. Ends with the choices the owner needs to make.

Parent record: `bugs_open/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md` §B.

---

## 1. Where we are (already done & live)

The original defect: fleet-wide, generated contact forms rendered fine and delivered nothing —
`form_action` came from per-component `content_data` written by the content LLM (`#contact`, `""`),
which POSTs to the current URL = 405/404 on a static host, so the message was silently lost. 10 of 11
live sites.

Shipped and verified live (chassis **v1.0.1150**):
- **Render-seam fix** — `sanitiseFormAction` / `sanitiseFormActionStrings` / `deliverableFormAction`
  in `component_library.go` convert a non-delivering `form_action` to a `mailto:` built from the
  site's real address, on both render branches. **Refuses to fabricate** an address it does not have,
  and refuses the synthesised `info@<own-domain>` fallback (commit `efe634b37`).
- **Unit tests** — `component_library_form_action_test.go` (5 tests, both branches, fault-injected).
- **Discovery check** — `check_contact_form_undeliverable.go`, ENABLED on `completeness-discovery-agent`
  (`sql_for_agents/190`). Reads DEPLOYED `rendered_html`, flags contact forms that still POST nowhere,
  routes to `needs_human_review` with **no handler**.
- **Addresses** — all 4 formerly address-less sites now have a real `sites.email`.

**Proven working live 2026-07-23:** `fundamentallyai.com` re-rendered and self-converted `#contact` →
a real `mailto:`; the check produced its first `needs_human_review` item.

**What is still missing** is exactly what this plan is about: the render fix and the check are both in
place, but nothing runs the check on a **cadence**, nothing **auto-repairs** the deployed pages, and
the **check itself has no test**. So a dead form can still be *produced* by a new path or *reappear*
on a page that has not re-rendered, and no standing mechanism would catch or fix it.

---

## 2. How the ask maps onto machinery that already exists

The platform already has a "checker → handler → loop" structure; the request maps onto it 1:1.

- **Checker** = a *discovery check* — a `check_*.go` under
  `platform/orchestration/actions/discovery_checks/` implementing `Name()` + `Run()`, self-registered
  via `init(){ Register(&X{}) }`. `Run()` returns `CheckResult{ WorkItems []WorkItemSpec }`; the runner
  inserts them into `site_work_items` (`discovery_checks.go:143-159`). A check runs only when its
  `Name()` is listed in a discovery agent's `default_config.workflow.steps.run_checks.config.checks`
  array. **We have this — it is enabled.**

- **Handler** = the `handler_agent` column on the work item. There is **no central item_type→handler
  table**; whoever writes the row stamps `handler_agent`, and the dispatch loop spawns that agent.
  Dispatch = `LoadWorkItemsAction` (`load_work_item_actions.go:550-590`, selects only
  `status IN ('triaged','approved')`) → `ClaimWorkItemAction` (`claim_work_item_action.go:88-165`,
  atomic claim; **handler-existence gate**: a `handler_agent` not present in `agent_definitions` is set
  `status='blocked'`) → spawn `handler_agent`.
  - A row with `status='needs_human_review'` is **never selected** by the loader, so it never
    dispatches — it just sits. `HandlerAgent=''` is a second, independent block.
  - The only promoter of discovery output into `'triaged'` is `TriageDetectedItemsAction`
    (`triage_detect_items_action.go:88-101`), which promotes **only `status='detected'`** and rewrites
    `pipeline → 'build'`.
  - **Our check emits the parked shape on purpose** (`check_contact_form_undeliverable.go:155-171`:
    `HandlerAgent:""`, `Status:"needs_human_review"`). That was correct while sites were address-less.

- **The loop** = the `improvement-loop` agent (`sql_for_agents/054_improvement_loop.sql`): an
  orchestrator that runs on ONE site per run —
  `ensure_site → discovery agents (quality/design/completeness) → LLM audit (design-audit + site-review)
  → triage_findings → if findings: insert a re-render item (handler_agent=rerender-pages) → dispatch →
  re-render → complete`. Discovery findings **are** `site_work_items` rows; there is no separate
  findings table. This is literally the "checker + handler" loop the request names.

- **The periodic driver** = the `improvement-sweep` `scheduled_tasks` row
  (`sql_for_tables/020_scheduled_tasks.sql:76-96`): `interval 600s`, `target improvement-loop`,
  `pre_query` returns **one site per tick** (`ORDER BY last_built_at ASC … LIMIT 1`), rotating the whole
  fleet over time. **It is currently `enabled=false` (disabled since ~2026-05).** So `completeness-
  discovery-agent` — which has no trigger of its own and is only spawned inside `improvement-loop` —
  runs today only on an initial site build or a manual trigger. **Nothing runs periodically right now.**

---

## 3. The facts that decide the design (verified live 2026-07-24)

1. **All 10 affected sites now have a resolvable `sites.email`** (the 4 supplied + 6 that already had
   one; all `@contactforsales.com`, none is `info@<domain>`). So every one is *auto-healable* in
   principle — the render seam will convert its form the moment it re-renders.

2. **The two re-render paths read the address from DIFFERENT columns** — this is the crux:
   - **Full-writer path** (`needs_page` → `page-build-handler`; also `render_site_components` /
     `section_editor`) loads via `loadSiteDataFull` → `COALESCE(si.email,'')`
     (`render_site_components_action.go:315,328`) → `ctx.Email` from the **`sites.email` column**. This
     is the column the check's remedy names and the one the 4 addresses were set in. ✅ correct source.
   - **Light section-rerender path** (`page_rerender` → `page-rerender` → `rerender_page_sections`)
     builds `ctx.Email` from `buildRerenderBaseData` (`rerender_page_sections_action.go:496-533`), which
     `SELECT content_data FROM sites` (`:503`) and maps `content_data.email` → `ctx.Email`. It **never
     reads the `sites.email` column.** ❌ wrong source for our data.
   - Live proof this matters: of the 10 sites, `content_data.email` is **empty for 8** and **stale for
     idea.uk** (`idea-uk@leopardess.uk`, vs the correct `idea.uk@contactforsales.com` in `sites.email`).
     So a light `page_rerender` as-is would repair almost none of them, and would bake idea.uk's stale
     address into a mailto.

3. **The light path re-runs our fix; the full path does too.** `rerender_page_sections` renders each
   section via `RenderTemplate` → `contextToInterfaceMap` → `sanitiseFormAction`
   (`rerender_page_sections_action.go:342`). So either re-render executes the seam — the only question
   is which address column it reads (fact 2).

4. **The full rebuild risks a wall the light re-render does not.** A full-writer rebuild of an
   already-`deployed` page is documented to bounce to `needs_human_review` at attempt 0 (the reason the
   manual remediation of the 10 was deferred in the first place). The light `rerender_page_sections`
   path has **no deployed-status gate** — it re-renders in place — so it avoids that wall. It also does
   **not** regenerate copy via the LLM (the full writer does, exposing the content-regression guard).

5. **There is no CI in this repo** (no `.github/workflows`, no Makefile test target). Go tests run on
   `go test` / a build, not on a clock. So "run tests periodically" can only mean the live-data check
   (§ layer 2), unless we add a runner.

6. **Adding a check name is safe regardless of image timing** — the runner skips-not-errors an
   unregistered name (`discovery_checks.go:122-127`).

---

## 4. What it would take — three layers

### Layer 1 — Regression tests (recommended regardless; cheap, no behaviour change)
- **Detection test** for the check (sqlmock, modelled on
  `discovery_checks/check_empty_sections_test.go`): asserts the check flags `#contact` / empty /
  `#anchor` / `/contact`, and does NOT flag a real `mailto:` or `/request`. The check has **zero tests**
  today.
- **Render-path integration test** (modelled on `tool_render_path_test.go`, the end-to-end test written
  for bug 024): a component with `form_action="#contact"` put through the render path comes out
  `mailto:` for an addressed site and unchanged for an address-less one — catching the "seam not wired
  into a path" class that 024/006 keep producing.
- **Effort:** ~2 test files. **Caveat:** no CI → they guard "someone ran the suite/a build," not a
  clock. Optional add-on: a `make test-contact-forms` target + a one-line script if a runnable periodic
  hook is wanted.

### Layer 2 — Periodic fleet detection (so a dead form is caught even if nobody looks)
Nothing runs periodically today (§2, improvement-sweep disabled). Two ways:
- **(2a) New lightweight rotation task** *(recommended for this class)* — a `scheduled_tasks` row that
  rotates live sites (one per tick, mirroring the improvement-sweep `pre_query`) and calls
  `completeness-discovery-agent` **directly**, skipping the LLM audit + auto-dispatch steps. Standing
  fleet coverage for the discovery checks with no LLM cost. ~1 seed row; the scheduler already supports
  the pattern (and the no-op-slot fix from `bugs_open/048` means an empty tick won't starve the group).
- **(2b) Re-enable `improvement-sweep`** — `UPDATE scheduled_tasks SET enabled=true …`. Every live site
  then cycles the **whole** improvement-loop on the 600s rotation, incl. the two LLM audit steps
  (design-audit, site-review) and auto-dispatch of *every* enabled check's handler. Heavier and
  pricier; the recorded operational policy is to only re-enable once the other enabled checks' handler
  agents exist, so findings clear rather than accumulate. (Our undeliverable-form items are safe under
  it — being `needs_human_review`/no-handler, they don't churn.)

### Layer 3 — Auto-remediation handler (turn "park for a human" into "self-heal")
Gate the check on `sites.email` resolvability, inside the existing per-page loop
(`check_contact_form_undeliverable.go:134-172`): compute resolvability once per site
(`COALESCE(sites.email,'')` non-empty, contains `@`, not `info@<domain>` — the same guard
`deliverableFormAction` uses); then **resolvable → emit a re-render item that self-heals; address-less
→ keep the current `needs_human_review` row unchanged.** All 10 are currently resolvable, so all 10
would auto-heal; the human-review branch stays for any genuinely address-less site.

The check's own header argues *against* a handler ("picking one automatically would guess") — that
reasoning applies **only** to the address-less case. For a site with a resolvable `sites.email` the
render seam is deterministic, so an auto re-render is not a guess; it executes the fix the operator's
data already implies. Restricting the handler to the resolvable branch dissolves the objection.

Two coherent designs for the re-render item:

- **Option A — full rebuild (`needs_page` → `page-build-handler`).** Mirrors
  `flag_page_image_rebuild_action.go:131-160`. Reads the **correct** `sites.email` column already, no
  source change needed. **Costs:** regenerates page copy via the LLM (content-regression guard exposure)
  **and** risks the deployed-page "bounce to `needs_human_review` at attempt 0" wall (fact 4) — the very
  wall that made manual remediation heavy. So it may not actually heal the 10 *deployed* pages without
  extra handling.

- **Option B — light re-render (`page_rerender` → `page-rerender`) + a one-line source fix**
  *(recommended)*. Mirrors `check_misdirected_cta.go:296-329` (a discovery check that already emits a
  `page_rerender`). No LLM, in-place, no deployed-page bounce. **Requires** a small fix to
  `buildRerenderBaseData` (`rerender_page_sections_action.go:503`) so the light path reads
  `COALESCE(sites.email, content_data.email)` instead of `content_data.email` alone — otherwise it
  repairs almost none of our 10 (fact 2) and would bake idea.uk's stale address in. **That fix is not
  overhead — it closes a real latent inconsistency** (the light path reading stale `content_data.email`
  while the full path reads `sites.email`; idea.uk is live proof). Use a dedicated `reason`
  (e.g. `form_action_stale`) added to `check_rerender_mode` so it routes to `rerender_page_sections`
  without also firing the unrelated `cta_links_stale` CTA recompute.

**Note (per CLAUDE.md, updated 2026-07-24):** council approval is now reachable (~80%), so the Go
changes in Layer 3 (and Layer 1's tests are fine without) should go through the advisory council gate
before/alongside committing — `platform/` code, one run per coherent task. The `buildRerenderBaseData`
change touches a shared render path fleet-wide and particularly warrants it.

---

## 5. The choices the owner needs to make

**Tests (Layer 1):** proposed regardless — cheap safety net, no behaviour change. (Say if you'd rather
skip.)

**Decision 1 — the auto-remediation handler (Layer 3):**
- **Option B** — light re-render + one-line `buildRerenderBaseData` fix. Cheap, no LLM, avoids the
  deployed-page bounce, self-heals all 10, and fixes a real inconsistency. *Recommended.*
- **Option A** — full `needs_page` rebuild. Reads the right column already, but regenerates copy via
  the LLM and risks the deployed-page attempt-0 bounce.
- **Detection-only** — no handler; keep parking undeliverable forms at `needs_human_review` for a
  human. Lowest risk, no auto-behaviour change.

**Decision 2 — periodic fleet detection (Layer 2):**
- **New lightweight rotation task** — rotates sites, discovery-only, no LLM. *Recommended.*
- **Re-enable `improvement-sweep`** — the whole loop incl. 2 LLM audit steps + auto-dispatch; heavier
  and pricier, with the policy caveat above.
- **Skip periodic for now** — rely on the check firing on build / manual trigger, as today.

---

## 6. References
- Check: `platform/orchestration/actions/discovery_checks/check_contact_form_undeliverable.go`
- Render seam: `platform/orchestration/actions/component_library.go`
  (`contextToInterfaceMap`, `contextToMap`, `sanitiseFormAction`, `deliverableFormAction`)
- Dispatch: `load_work_item_actions.go:550-590`, `claim_work_item_action.go:88-165`,
  `triage_detect_items_action.go:88-101`
- Re-render: `create_rerender_items_action.go`, `rerender_page_sections_action.go` (esp. `:342`, `:496-533`),
  `check_misdirected_cta.go:296-329`, `flag_page_image_rebuild_action.go:131-160`
- Full-writer address source: `render_site_components_action.go:204,315,328`
- Improvement loop: `sql_for_agents/054_improvement_loop.sql`; sweep `sql_for_tables/020_scheduled_tasks.sql:76-96`
- Check enable convention: `sql_for_agents/190_enable_contact_form_undeliverable_check.sql`
- Test models: `discovery_checks/check_empty_sections_test.go`, `tool_render_path_test.go`,
  `component_library_form_action_test.go`

---

## 7. DECISIONS & BUILD (2026-07-24, same day)

**Owner decisions:** Layer 3 = **Option B** (light re-render + address-source fix). Layer 2 =
**skip periodic for now** (rely on build/manual-trigger discovery; revisit later). Layer 1 tests: yes.

**BUILT & COMMITTED `cc2cff79b` — inert until a chassis image roll.** What shipped:
- `buildRerenderBaseData` now reads `COALESCE(sites.email,'')` in its single-row query and applies it
  to `base["email"]`/`base["contact_email"]` AFTER the content_data merge — the canonical column wins,
  empty column falls back to content_data (no regression). Both render paths now agree on the source.
- `check_contact_form_undeliverable` gates on resolvability (`resolveSiteContact` +
  `contactAddressResolvable`, mirroring `deliverableFormAction` incl. the `info@<own-domain>` refusal;
  lockstep comments both sides). Resolvable → `page_rerender` item (handler `page-rerender`, status
  `detected`, pipeline `build`, reason `section_data_resolved`, key
  `contact_form_undeliverable_rerender:<page>`). Address-less → the unchanged `needs_human_review` row.
  Header routing paragraph updated to document both branches.
- Tests (all fault-injected, watched to fail): `TestContactAddressResolvable`,
  `TestContactFormUndeliverableRoutesByResolvability` (sqlmock end-to-end routing),
  `TestBuildRerenderBaseDataPrefersSitesEmailColumn` (the idea.uk stale-address case is a literal
  test case). actions-package tests verified via `git archive HEAD` + overlay (another session's WIP
  in `diagnose_dormant_agents_test.go` broke the shared tree's test build — not ours).

**Design points settled during build** (rationale in the submission JSON alongside this file):
- Reason = `section_data_resolved`, NOT a new `form_action_stale`: the routing conditional lives in
  the page-rerender agent's config, and a config-only reason would be clobberable by an agent re-seed
  (documented landmine); the shared reason is seed-stable, routes to the true template re-render, and
  does NOT fire the `cta_links_stale`-gated CTA recompute. Its `scoped` behaviour in
  `create_rerender_items` needs a `component_id` the spec doesn't carry, so it is not triggered.
- Distinct dedup key for the rerender item, so a site that gains an address later is not suppressed
  by a stale parked human-review row (`check_misdirected_cta` precedent for a check-owned key).
- `contactAddressResolvable` duplicates (not imports) the actions-package guard — package boundary;
  parity is pinned by tests on both sides.

**Council:** submitted alongside the commit (advisory, per the 2026-07-24 norm) —
`SUBMISSION_CORR=5d64be67-b9e8-47e8-8768-828a34093b08`, submission JSON
`submission_006B_autoheal_handler.json` (this directory). Verdict PENDING at time of writing; if
APPROVED, note the trailer in a follow-up commit.

**What this means operationally once the image rolls:** the next discovery cycle on each affected
site emits a `page_rerender` instead of parking — the 9 remaining `#contact` forms self-heal through
triage → dispatch → light re-render → `sanitiseFormAction` → mailto to `sites.email`. The deferred
"remediate the 10 deployed" follow-up (2) is thereby absorbed into the standing loop. Any future
address-less site still parks for a human — the honesty rule is untouched.
