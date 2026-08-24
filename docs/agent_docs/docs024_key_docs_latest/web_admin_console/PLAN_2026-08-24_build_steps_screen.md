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
(`1c6f3424-9d05-4a18-963b-72541bc19dca`), submitted 2026-08-22 12:18. **`[MEASURED 2026-08-24
11:31Z]`, re-run after the first version of this table was shown to be a stale snapshot (see the
correction box below):**

| item_type | status | created | completed | took |
|---|---|---|---|---|
| `needs_domain_research` | complete | 12:18:28 | 12:26:30 | 8m |
| `site_unreachable` | complete | 12:20:38 | **16:21:44** | self-resolved once the page served |
| `needs_vertical_research` | complete | 12:26:28 | 12:37:26 | 11m |
| `needs_strategy` | complete | 12:37:25 | 12:43:31 | 6m |
| `needs_briefing` | complete | 12:43:29 | 12:49:24 | 6m |
| `needs_site_plan` | complete | 12:49:22 | 12:55:57 | 7m |
| `needs_composition` | complete | 12:55:55 | 13:00:44 | 5m |
| `needs_page` | complete | 12:55:54 | 13:06:06 | 10m |
| `needs_design` | complete | 12:55:55 | 13:15:43 | 20m |
| `needs_imagery` ×2 | complete | 12:55:55 | 13:18–13:20 | ~23m |
| `needs_rerender` | complete | 12:55:54 | 13:25:00 | 29m |

**The whole fresh-build cascade ran in ~67 minutes**, 12:18 → 13:25, each stage handing to the
next within seconds of the previous completing. That is the screen: a dozen rows, in order, with
durations. `/admin/workflows` cannot produce it.

**Two design constraints this corrected read adds, which the stale one hid:**

1. **The chain is a minority of the rows.** That site now has **70** work items `[MEASURED
   2026-08-24]`, of which **25** are `page_divergence_overwritten` (a rebuild overwriting
   hand-patched bytes, filed once per component per rebuild) and several more are `needs_imagery`
   fan-out. A screen that lists a site's work items chronologically shows the build buried in
   repeat-noise. **The pipeline stages must be an explicit ordered vocabulary the screen knows
   about**, not "whatever rows this site has".
2. **Terminal-looking rows are not the end of the build.** `site_unreachable` sat `detected` for
   ~4 hours and then went `complete` on its own at 16:21 when the page began serving. A screen
   showing red-until-someone-acts would have shown a false alarm for four hours.

**`[UNMEASURED]` how completely the early stages survive on OTHER sites** — the fresh-build types
are rare fleet-wide (`needs_domain_research` **7** rows all-time as of 2026-08-24,
`needs_strategy` 5, `needs_briefing` 5, `needs_site_plan` 4), because full fresh builds are rare,
not because rows are reaped. `apis.uk` is one site's chain, not proof the shape generalises.

> **CORRECTED 2026-08-24 11:31Z — the first version of this table was a two-day-old snapshot
> presented as current, and the `apis_uk_bees_homepage` lane caught it.** It showed three rows and
> said the build was "sitting at `needs_vertical_research` triaged". Every cell was true **when I
> ran the query — at ~12:26 on 2026-08-22**. `needs_vertical_research` completed **11 minutes
> later** and the cascade finished the same afternoon.
>
> **The measurement did not go wrong; re-publishing it did.** This session began on 08-22 and the
> date rolled to 08-24 mid-session. I carried the result forward and — this is the part that was
> never measured at all — added *"about two days now, not progressing"*, an inference generated by
> subtracting the row's timestamp from today's date. **A snapshot of STATE expires; the inference
> I bolted on was pure arithmetic on a stale row**, and it is what turned a harmless old figure
> into a false claim about a live build.
>
> **The check: re-run any state query before quoting it in a document written on a later day**,
> and never compute an age from a stored timestamp without re-reading the row. The lane's own
> memory index already carried this rule — *"a `[MEASURED]` claim about STATE expires while a
> DATED EVENT does not"* — so this was a known rule, violated, not a gap. Second wrong call of
> this session; both in `WRONG_CALLS.md`.

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

### 6g. A one-line addition worth more than the screen: the silent-overwrite counter

From correspondence with the `apis_uk_bees_homepage` lane, 2026-08-24, who lived it.

**The trap the console itself opens.** The SPA already exposes `regenerate`, `restore-section`
and `lock` on the same screen (`App.tsx`, the `/sites/:id/pages/:page/components/...` calls). An
operator who hand-patches a component through that UI and does **not** lock it has armed a silent
loss: the next rebuild overwrites their bytes, archives the old copy to `page_component_history`,
and files a `page_divergence_overwritten` work item. Nothing in the UI says this happened.

**The measured case.** On `apis.uk`, that fired **25 times over two days** — same lane, same
cause, every one correct — and stayed invisible until somebody read the work-item queue by hand.
`[MEASURED 2026-08-24 11:31Z]` all 25 now `complete`, one shared `updated_at` of `11:29:48`, i.e.
resolved in a single sweep once noticed. **25 of that site's 70 work items were this.**

**The fix is a pair, and knowing only half of it is what cost them two days** — the work item's
own fix text names both: *re-declare it in `content_data` **or lock the component (058)** — do not
paste it back into `rendered_html`, which only re-arms this same loss.* They had re-declared and
not locked. `save_page_sections_action.go:460` documents the locked-slot path (locked copy kept,
incoming discarded), and **43 components across 6 other lanes** already use `lock_type='permanent'`
— established practice, not a new mechanism.

**What the screen should do, and it needs NO backend work**: beside the regenerate button, show
the component's `page_divergence_overwritten` count from `site_work_items`. The rows already
exist and already carry `site_id`; this is a `GROUP BY` on data the admin API can already reach.
An unlocked component with a non-zero count is the exact state that is about to lose someone's
edit again.

> This is the highest value-per-line item in this plan. The build-steps view tells the owner what
> happened; this tells him what is **about to** go wrong, and it is cheaper than any other item
> here. If only one thing gets built, consider making it this.

### 6h. The console's spec editor makes a WISH and a CONTROL look identical

This is the "contribute" half of the owner's ask, and it has the same shape as 6g: the UI offers
an action whose effect is far weaker than it appears, and says nothing about it.

**What I verified myself** `[MEASURED 2026-08-24]`:

- `PATCH /admin/sites/:site_id/specs/:aspect` takes the aspect as a **free string** — the only
  validation is `if aspect == ""` (`site_admin_handlers.go:197-199`); there is no allow-list. It
  supersedes-then-inserts into `site_specs`. The SPA reaches it from `SpecEditor` via
  `/sites/${siteId}/specs/${selectedSpec.aspect}`, and lists every aspect from `GET .../specs`.
- Live aspects include both kinds side by side: `content_direction` (**94** rows),
  `briefing` (64), `strategy` (59), `mission_brief` (21), `roadmap_brief` (8) — **and**
  `evidence_base` (**313**).
- The enforced prohibitions are **not a table** — `banned_claims` does not exist in `clients_db`.
  They are a **key inside the `evidence_base` spec's JSON**: `{"facts": [], "banned_claims": []}`
  (`platform/orchestration/actions/evidence_citations.go:222`;
  `datahelpers/claims.go:263` types it as `BannedClaims []BannedClaim`).

**What the `apis_uk_bees_homepage` lane measured, and it refuted a hypothesis of mine** — I had
guessed the gap was a write-time-only checker. It is not: `ScanDeployedClaims`
(`discovery_checks/check_unverified_claims.go`) reads **deployed** content, its finding carries
`Source: rendered_html | content_data`, and its header says it exists precisely because the
build-time gate is insufficient. **The scanner was working the whole time.** Their site served
an `<h1>` of "A page about bees" for two days — the exact string its own `roadmap_brief` calls
unacceptable — and **0 of the 38 `banned_claims` patterns matched it**. The rule and the breach
sat in the same database with nothing to join them.

**So the finding for this console:** `roadmap_brief`, `content_direction`, `briefing`,
`mission_brief` and `writer_block` are **prompt text — instructions to a writer, enforced by
nothing**. `evidence_base.banned_claims` is the only enforced layer. The console edits both
through the same editor, with the same gesture and the same success toast. **An owner who types
"never say X" into `content_direction` has written a wish; the identical sentence added as a
pattern under `evidence_base.banned_claims` is a control.** Nothing on the screen distinguishes
them, and the wish is the more natural place to type it.

This is CLAUDE.md's own owner ruling of 2026-08-02 §2 in a second domain — *"a comment is not a
control on a tree this many sessions share"* — and the same remedy applies: the enforcing
artefact must be the thing the operator edits, not a doc the operator hopes someone reads.

**Cheapest honest version, no backend work:** in `SpecEditor`, label the prompt-text aspects as
advisory and `evidence_base` as enforced, and when an edit to a prompt-text aspect contains a
prohibition ("never", "must not", "do not say", "unacceptable"), prompt for the matching
`banned_claims` pattern in the same save. That lane's own prescribed check is the same pair:
**when you write "must never X" into a brief, add the matching `banned_claims` pattern in the
same edit**, then prove the pair fires on the forbidden string and stays silent on the live page.

~~⚠ **`[UNVERIFIED]`** whether `SpecEditor` renders `evidence_base` usefully (313 rows, large JSON),
and whether a hand-edited `banned_claims` entry is picked up without a re-seed. Check both before
building anything on this.~~ **BOTH RESOLVED 2026-08-24** — raised by the `apis_uk_bees_homepage`
lane, re-verified here. And one of them turned up a hazard that outranks everything else in this
plan.

**Q2 — does a hand-edited ban take effect without a re-seed? YES, immediately.** `loadEvidenceBase`
(`platform/orchestration/actions/validate_page_content.go:1281`) is a plain live read of
`site_specs … aspect='evidence_base' AND is_current = true`. No cache, no seed step, no rebuild;
`is_current` resolves at call time. **So "add the ban in the same save" is genuinely available
from the console** and the §6h remedy keeps its shape.

**Q1 — sizes, and a correction to my own figure above.** ~~`evidence_base` (**313** rows)~~ —
**313 is ALL rows including superseded history; only 19 are current** `[MEASURED 2026-08-24]`:

| | |
|---|---|
| rows with a **current** `evidence_base` | **19** |
| largest current | **95,730 bytes** |
| average current | **17,850 bytes** |

> **That is the same error I logged twice today, a third time.** I quoted a row count from a
> supersede-then-insert table as if it were a population of live registers. `site_specs` keeps
> history — every aspect count in §6h above (`content_direction` 94, `briefing` 64, …) is a
> history count, not 94 live content directions. The counts still make the point they were cited
> for (both kinds of aspect are in live use, edited through one editor) but **no figure in §6h
> should be read as "how many sites have this".** Filter `is_current` before quoting any
> `site_specs` count. Caught by the same lane, not by me.

So a naive JSON textarea is ~17KB typical, ~94KB worst case. The hard element is not the array:
`apis.uk`'s current register holds 40 bans, 41 allowed entities **and a 3,733-character
`writer_block` — one multi-paragraph prose field inside the JSON.** A generic key/value editor
makes that field close to unusable.

#### ⚠⚠ THE HAZARD: the console can switch a site's claims checking OFF, silently, with a 200

This is the sharpest thing in this document and it is live today — it needs no new screen.

`HandleUpdateSiteSpec` (`site_admin_handlers.go:203-238`) binds `data` as `json.RawMessage` and,
in one transaction, **supersedes the current row and inserts the new one**. It validates
**nothing about the shape** — only that the envelope is JSON.

Now the receiving end, and it is quieter than "malformed JSON gets logged":

- `ParseEvidenceBase` (`datahelpers/claims.go:278`) returns an **error** only on `json.Unmarshal`
  failure — the case a `jsonb` column would reject anyway.
- **It returns `(nil, nil)` — no error — when the parsed object has no facts, no banned claims
  and no regulated attestation** (`claims.go:291`).
- So valid JSON of the **wrong shape** — a misspelt `bannedClaims`, a pasted fragment, an object
  nested one level too deep — unmarshals cleanly into an empty struct and comes back `(nil, nil)`.
- `loadEvidenceBase` then never reaches its warn branch (`err == nil`), and returns **nil with no
  log line at all**. `UnrecognisedKinds()`/`AliasedKinds()` are nil-safe (`claims.go:230-233`,
  with a test asserting it), so there is no panic either.

**Net effect: one save of well-formed-but-wrong JSON supersedes a good register, returns 200 and
a success toast, and every claims lane on that site silently no-ops from then on. The site then
reports clean — because nothing is checking.** The good register is already `is_current = false`
by the time anyone could notice, recoverable only from history.

**What the editor owes, before it is trusted with this aspect** — none of it needs backend work
beyond the handler:
1. **Parse-and-reload on save**, not just a 200: re-read the row through `ParseEvidenceBase` and
   fail the request if it comes back nil for an aspect that previously parsed non-nil.
2. **Show the counts back** — "40 banned claims, 12 facts saved" — because a zero there is the
   whole signal, and it is the one number a wrong-shape save cannot fake.
3. **Refuse to supersede a non-empty register with an empty one** without an explicit confirm.
4. The aspect is a free string with **no allow-list**, so a typo creates a fourteenth aspect that
   nothing reads (13 distinct current aspects on `apis.uk` alone) — the same silent-no-op class
   one level up. An allow-list, or a warning on an unknown aspect, closes it.

**This is `bugs_open/105`'s own lesson repeating one layer out**: that fix made an unrecognised
fact *kind* announce itself rather than behave as a default. An unrecognised evidence_base
*shape* still does not announce itself at all.

### 6i. Falsifiers for everything above

- All counts are a 7-day window read on **2026-08-24** and move daily; `orchestration_states`
  took ~50 new rows during the ~20 minutes these queries ran (4,353 → 4,410 across passes).
- 6d's "the build is a work-item chain" is read from `082_submit_domain_unified.sh`'s header and
  one live site's rows. It has **not** been checked against the handler code for each stage.
- 6e's terminate-500 is inferred from `to_regclass` + the code path; **the endpoint was not
  called** (it mutates, and a live workflow is not mine to terminate).
- 6b's 67 = 67 identity holds for this window; a future nested writer would break it. The test
  that detects that is the `'"__step_error":'` strpos count diverging from the `?` count.
