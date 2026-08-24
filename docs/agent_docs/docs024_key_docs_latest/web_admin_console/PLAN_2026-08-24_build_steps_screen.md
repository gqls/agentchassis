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
