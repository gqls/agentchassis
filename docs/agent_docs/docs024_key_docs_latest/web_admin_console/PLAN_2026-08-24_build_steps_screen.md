# PLAN — the build-steps screen ("Builds" tab in the admin console)

**Status: PLANNED, not built.** Owner asked 2026-08-24. This is the feature the whole
admin-console workstream exists for: *"follow and contribute to the steps of each website
build."* "Contribute" already has screens (specs, components, work items — and the owner is
now logged in and using them at `admin.apis.uk`, as `uk@websy.uk`). "Follow" is this.

## 0. The gap, measured

The API serves the build steps; the SPA never asks. `grep -c workflow
frontends/admin-dashboard/src/App.tsx` → **0**. The three routes
(`internal/core-manager/api/server.go:183-185`):

| route | handler | returns |
|---|---|---|
| `GET /api/v1/admin/workflows` | `HandleListWorkflows` | `{workflows: [...], count}` — rows of `correlation_id, client_id, status, current_step, created_at, updated_at, error` |
| `GET /api/v1/admin/workflows/:correlation_id` | `HandleGetWorkflow` | full row incl. `awaited_steps`, `collected_data`, `initial_request_data`, `final_result` |
| `POST /api/v1/admin/workflows/:correlation_id/resume` | `HandleResumeWorkflow` | the operator action |

List filters (`WorkflowListRequest`, `system_handlers.go:152`): `status`, `client_id`,
`start_date`, `end_date`, `limit` (default 50), `offset`. **There is NO site/domain filter** —
see §2, it is the one backend change this plan needs.

## 1. The screen, smallest honest version

A **Builds** tab in the existing nav (Sites · All Items · Pipelines · Customers · **Builds**),
plus a **"Builds" button on each site card** — the second entry point is the one the owner
actually described ("each website build", i.e. arrive from the site).

1. **List view**: table of orchestrations — when, status, `current_step`, correlation id
   (truncated, click-to-copy), coloured like the existing `STATUS_COLORS`. Filters: status
   dropdown, date range, and the site (once §2 lands). Poll every ~10s while the tab is
   visible — builds move on minute timescales; websockets are not needed for v1.
2. **Detail view**: the step timeline. Derive the step list from `collected_data`'s keys in
   arrival order plus `current_step` and `awaited_steps`; show each step's output size and a
   collapsible JSON viewer (the `JSONEditor` component at `App.tsx:231` already exists —
   reuse read-only). `final_result` and `initial_request_data` as top/bottom cards.
3. **Resume button** on non-terminal rows, with a confirm dialogue naming the correlation id.

## 2. The one backend change: a site filter

`orchestration_states` has no site column; site identity lives inside the JSON. Add one
optional query param to `HandleListWorkflows` — `site_id` — implemented as
`initial_request_data->>'site_id' = $n OR collected_data->'input_data'->>'site_id' = $n`
(verify the actual key paths against live rows before coding; do not trust this line).
Small, additive, admin-only — normal council gate, no RFC shape.

## 3. ⚠ The landmine this screen MUST build around, or it will lie

**`bugs_open/099`: a FAILED step can show COMPLETED with the `error` column NULL — the truth
lives in `collected_data`'s `__step_error` keys.** A screen that renders `status` + `error`
verbatim will show a green build that silently discarded its design. So:

- the detail view must surface any `__step_error` entries **prominently**, whatever `status` says;
- the list view should mark rows whose `collected_data` contains `__step_error` (needs the
  backend filter change to expose a boolean cheaply — `collected_data ? '__step_error'` shape,
  again verify the real key spelling against live rows first);
- never render "error: —" from the NULL column as "no error".

## 4. Order of work

1. Verify the JSON key paths (§2, §3) against live `orchestration_states` rows — 10 minutes,
   sizes everything.
2. Backend: `site_id` filter + step-error boolean on the list route. Council gate (internal/).
3. Frontend: Builds tab + detail view + resume. No new packages; follow `App.tsx`'s existing
   fetch/`Badge`/`STATUS_COLORS` conventions.
4. Wire the site-card button.
5. Frontend image rebuild + deploy (frontends build from their own context, not `git archive`).

## 5. Risks

- `ADM-002` records bugs/mock data in parts of the admin API, predating the 07-13 freeze —
  re-verify the three workflow handlers actually return live data before styling anything.
- 369+ orchestrations exist for some sites; default `limit=50` + offset paging is fine, but the
  site filter must be server-side or the screen will fetch pages of other sites' rows.
- The resume action is real and mutating — keep the confirm dialogue, and log-colour it like
  the existing retry buttons.

---

## 6. CORRECTIONS — measured 2026-08-24, before any code was written

Added by the session the owner sent to correspond with this lane rather than build a second
console. These are §4 step 1 ("verify the JSON key paths — 10 minutes, sizes everything") done.
All figures `[MEASURED 2026-08-24]` against `clients_db`, window = last 7 days unless stated.

### 6a. ⚠ §2's premise is FALSE — `orchestration_states` HAS a `site_id` column, and it is indexed

§2 says *"`orchestration_states` has no site column; site identity lives inside the JSON"* and
proposes a JSON-extraction filter. **It has one.** `\d orchestration_states` lists `site_id uuid`
as the last column, and `pg_indexes` shows **three** indexes on it — `idx_orch_site`,
`idx_orch_site_id`, `idx_orch_site_active`.

The two candidate paths §2 offers, measured against the column:

| expression | rows non-NULL of 4,410 |
|---|---|
| `site_id` (the column) | **2,136** |
| `initial_request_data->>'site_id'` | **0** — this path does not exist |
| `collected_data->'input_data'->>'site_id'` | **2,136** — the same set as the column |

So the backend change is smaller and safer than §2 thinks: add `site_id` to the SELECT list and
`AND site_id = $n` to the WHERE. **No JSON extraction, and it hits an existing index.** Half of
§2's proposed predicate (`initial_request_data`) would have matched nothing and the `OR` would
have hidden that.

### 6b. §3's `collected_data ? '__step_error'` is CORRECT AND EXACT — do NOT widen it to a text scan

§3 hedges (*"verify the real key spelling"*) and suggests the shape `collected_data ?
'__step_error'`. The spelling is right and **the top-level key test is exact**:

| test, COMPLETED rows (4,359) | count |
|---|---|
| `collected_data ? '__step_error'` (top-level key) | **67** |
| `strpos(collected_data::text,'"__step_error":')>0` (real key, any depth) | **67** |
| `strpos(collected_data::text,'__step_error')>0` (bare literal, any depth) | **176** |

The key never appears nested — 67 = 67. **The extra 109 rows are workflow CONFIGURATION naming
the field, not errors**: `"note_body_field": "__step_error.message"` in an `append_doc_note`
step's config. A substring test would mark **109** clean builds as failed. This is the estate's
"prompt text scores as the behaviour it describes" trap; the jsonb key operator is immune to it
because it asks the parser, not the text.

> **I got this wrong first and it is worth recording why.** My initial pass reported the
> top-level test "misses 109 of 176" and called it a grep-approximating-a-parser defect. Both
> halves were wrong, in opposite directions. Two separate errors compounded:
> 1. **`LIKE '%__step_error%'` is not a literal search.** In SQL `LIKE`, `_` is a single-character
>    wildcard, so that pattern means "any two characters followed by `step_error`". It returned
>    **315**. Re-run with `strpos`, the honest count is **176**. A pattern language silently
>    reinterpreting the very characters that make this key distinctive is the whole trap.
> 2. **Then I assumed the 176−67 gap was nesting** without looking at one. Extracting 320
>    characters around the literal settled it in one query: the gap is config text.
>
> The check that would have caught both at once, and the one to use: **read one matching row
> before believing any count derived from a pattern.**

### 6c. `bugs_open/099` quantified — the landmine §3 warns about, in numbers

`[MEASURED 2026-08-24]` Of **4,359** COMPLETED orchestrations in 7 days, the `error` column is
non-NULL on **0** of them, while **67** carry a top-level `__step_error`. FAILED rows (33) set
the column on all 33. **So `status` + `error` rendered verbatim would show 67 green builds in a
week that had a step fail.** §3 is right to make this blocking; this is its size.

### 6d. ⚠ THE BIGGER ONE — `/admin/workflows` does not return "builds", and §0 implies it does

§0 says *"The API serves the build steps; the SPA never asks."* The first half does not hold, and
this changes the screen's shape rather than its styling:

- **`execution_path` is empty on 100% of rows** (0 of 4,410 have a non-empty array). There is no
  stored step sequence to render. §1's plan to derive the timeline from `collected_data` keys
  plus `current_step`/`awaited_steps` is therefore not one option among several — it is the only
  one available, and the "steps" it shows are reconstructed, not recorded.
- **Only 2,136 of 4,410 rows have a site at all** (6a). The rest cannot appear under any site.
- **`orchestration_name` is empty on ~40%** of rows; the bulk of the named remainder are
  `generic-orchestrate-*` / `generic-process-*` machinery, not builds.
- **A site's `site_id`-tagged orchestrations are mostly its periodic sweeps, not its build.** For
  `agritec.uk` (`0a538b4a-…`) all 8 tagged rows are availability/quality/completeness discovery
  and render-audit runs.

**What a "website build" actually is on this platform: a chain of `site_work_items`**, and it is
documented in `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh`:
`needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → (cascade)
needs_composition → needs_design → needs_content_page ×N → rerender`.

Live worked example, and the shape the owner described — `apis.uk`
(`1c6f3424-9d05-4a18-963b-72541bc19dca`), submitted 2026-08-22 12:18:

| item_type | status | created | updated |
|---|---|---|---|
| `needs_domain_research` | complete | 12:18:28 | 12:26:30 |
| `site_unreachable` | detected | 12:20:38 | — |
| `needs_vertical_research` | triaged | 12:26:28 | — |

That reads as "where is my build, what is it waiting on" in three rows. `/admin/workflows`
cannot produce it. **`[UNMEASURED]` how completely the early stages survive** — the fresh-build
types are rare fleet-wide (`needs_domain_research` **7** rows all-time as of 2026-08-24,
`needs_strategy` 5, `needs_briefing` 5, `needs_site_plan` 4), because full fresh builds are rare,
not because rows are reaped. Worth confirming before designing around it.

**Suggested consequence, for this lane to decide, not me:** the Builds screen is probably
`site_work_items` grouped by site and ordered by the pipeline above, with `/admin/workflows`
as a drill-down for one stage's orchestration — rather than a workflow list with a site filter.
That inverts §1 and §2. I have NOT acted on it.

### 6e. Two live admin-API bugs any workflow screen will hit

- **`POST /admin/workflows/:id/resume` with `action:"terminate"` 500s.**
  `internal/core-manager/admin/system_handlers.go:578` UPDATEs **`orchestrator_state`**, which
  does not exist — `SELECT to_regclass('public.orchestrator_state')` → NULL; every other query
  in the file uses `orchestration_states`. The `resume` arm does not touch it, so only terminate
  is affected. This is register `ADM-002`'s **B2**, still live.
- **`ADM-002`'s B1 (MySQL `CURDATE()` in the dashboard query) is in UNREACHABLE code.**
  `NewDashboardHandlers` has **zero callers repo-wide** and `DashboardHandlers` has **zero
  references** outside its own file; no route mounts it (`server.go:101` notes the SPA and auth
  proxy are served by nginx). Also worth knowing before anyone "fixes" it: that query runs
  against `h.authDB`, which **is** MySQL, so `CURDATE()`/`INTERVAL 7 DAY` would have been correct
  syntax there anyway. The register entry reads as though a live dashboard is broken; it is
  neither live nor, in that respect, broken. Correct the entry rather than the code.

### 6f. Two smaller corrections to the older lane docs

- **`RUNBOOK_web_admin_console.md` and `PLAN_2026-08-22` name the auth host wrongly.** They say
  identity lives in an external MySQL at `catalogu_vectordb_chassis:3306`. That is the **database
  name**, not the host. `[MEASURED 2026-08-24]` from `cm/personae-prod-config`:
  `AUTH_DB_HOST=rs17.uk-noc.com`, `AUTH_DB_PORT=3306`, `AUTH_DB_NAME=catalogu_vectordb_chassis`,
  `AUTH_DB_USER=catalogu_personae`. (The `[UNVERIFIED]` admin-account question those docs carry
  is separately resolved — the owner has logged in as `uk@websy.uk`.)
- **auth-service's own admin routes are unreachable through the console's gateway.**
  auth-service serves `GET/PUT/DELETE /api/v1/admin/users…`, `/api/v1/admin/subscriptions`
  (its startup route dump), but `frontends/admin-dashboard/nginx.conf` sends **only**
  `/api/v1/auth/` to auth-service and **everything else under `/api/v1/`** to core-manager,
  which has no `/admin/users` route. So a user-admin screen is not buildable without an nginx
  location change. Not needed for the build-steps screen — recorded so nobody plans one blind.

### 6g. Falsifiers for everything above

- All counts are a 7-day window read on **2026-08-24** and move daily; `orchestration_states`
  took ~50 new rows during the ~20 minutes these queries ran (4,353 → 4,410 across passes).
- 6d's "the build is a work-item chain" is read from `082_submit_domain_unified.sh`'s header and
  one live site's rows. It has **not** been checked against the handler code for each stage.
- 6e's terminate-500 is inferred from `to_regclass` + the code path; **the endpoint was not
  called** (it mutates, and a live workflow is not mine to terminate).
- 6b's 67 = 67 identity holds for this window; a future nested writer would break it. The test
  that detects that is the `'"__step_error":'` strpos count diverging from the `?` count.
