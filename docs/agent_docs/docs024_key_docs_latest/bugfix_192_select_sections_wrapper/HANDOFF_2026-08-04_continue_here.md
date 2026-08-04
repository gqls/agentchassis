# HANDOFF 2026-08-04 — `bugs_open/192`, continue here

**Read this first, then `NOTES_select_sections_wrapper.md` bottom-up.** The lane is
`bugfix_192_select_sections_wrapper`. Everything below is measured unless marked.

---

## State in one paragraph

The bug is **diagnosed, fixed at source, council-APPROVED, and the Go half is LIVE on
chassis `v1.0.1250` (both replicas pod-verified)**. What remains is the closing sequence:
confirm the wrapper is gone on a fresh run, retire the temporary config shim, re-check that
`bugs_open/178` is not regressed, and move the file to `bugs_closed/`. Nothing is blocked
and nothing is at risk; the fleet is building pages again.

## What the bug was

`page-build-handler`'s `load_current_section_content` step (shipped by `bugs_open/178`,
commit `08d0515f3`, seed `299`) declares **`output_field: section_plan`** — reusing the key
`plan_sections` writes, deliberately, so no downstream `input_mapping` needed changing. But
the action returned a **wrapper** `{section_plan, applied, reason|matched}` on *every* return
path, including all eight it documented as "pass-throughs". `coordinator.go:1859-61`
(`storeActionResult`) stores a return value **wholesale** under `output_field`, so the flat
plan was demoted one level on **every page build in every mode**.

One cause, both consumer routes: `select_sections`' path 2 died directly, and path 1 died
because `resolve_links`' `input_mapping` reads `"sections?"` from the same flat path, so the
resolver was handed nothing and returned null. `ExtractFieldsAction` then **omitted the key
and returned success**, and `process_sections_loop` failed two steps later naming a missing
key. Every page build in the fleet failed 08:20→09:01Z.

## What shipped

| # | change | where | state |
|---|---|---|---|
| 1 | the action returns **the plan itself** on every path; bookkeeping to the log, and on the applied path only to `edit_live_meta` **inside** the plan | `load_current_section_content_action.go` | **LIVE** v1.0.1250 |
| 2 | `extract_fields` opt-in **`required`** list — a target resolving on no path fails the step naming the field, paths tried, and keys in scope. **Default OFF** | `v3_site_actions.go#ExtractFieldsAction` | **LIVE** |
| 3 | loop path-miss error lists **keys present at the failing level** | `loop_actions.go#getNestedValueForLoop` | **LIVE** |
| 4 | **second instance** found by census: `enrich_fingerprint_with_css` returned a status stub on both early-outs, overwriting a real `design_fingerprint` | `enrich_fingerprint_with_css_action.go` | **LIVE** |
| 5 | `required` cross-checks configured targets; a typo fails as a *step-config* error | `v3_site_actions.go` | **LIVE** |
| 6 | seed: third fallback path (**temporary shim**) + the `required` opt-in | `sql_for_agents/308_*.sql` | **applied + ledger-recorded** |

Tests: `load_current_section_content_action_test.go` (rewritten — the old one asserted the
wrapper and so **passed on the broken code**), `extract_fields_required_test.go` (new, 5
cases), `enrich_fingerprint_shape_test.go` (new, 3 cases). **All mutation-proven**: reverting
each fix fails the cases with the intended message. `go test ./platform/orchestration/...`
is green.

Council: **APPROVED round 2**, correlation `7afbf531-5ddd-484e-88c8-091994a0f51f`,
5 advisories none high-severity. Four acted on; the fifth (a guard on `storeActionResult`
itself) is architecture scope and is routed into **RFC_012**, not dropped.

## THE REMAINING SEQUENCE — do these in order

### 1. Confirm the wrapper is gone at source (V3) — IN FLIGHT

Work item `c67c4e86` (gaswholesalers.com / who-we-serve) was set `triaged` at 10:38:04Z to
induce a fresh `page-build-handler` run. Watch for it:

```sql
SELECT left(orchestration_id::text,8), owner_agent_type, status, current_step,
       (collected_data->'section_plan') ? 'applied'        AS wrapper_present,
       (collected_data->'section_plan') ? 'sections_ready' AS flat_ok, created_at
FROM orchestration_states
WHERE created_at > '2026-08-04 10:29' AND owner_agent_type IN ('page-build-handler','page-content-writer')
ORDER BY created_at DESC LIMIT 5;
```

**Required: `wrapper_present = f` AND `flat_ok = t`.** That is the whole point of the fix,
and it is falsifiable — `t/…` means the old binary still serves or the fix regressed.

> If that item does not dispatch, any `failed` `needs_page`/`content_rewrite` row will do —
> set `status='triaged'` (the dispatcher only takes `triaged`/`approved`, so a `failed` row
> never retries itself). Check `sites.locked_at IS NULL` and `approval_mode='auto'` first,
> and **do not dispatch within ~300s of a chassis pod restart** — the spawn is silently
> dropped. `5816c2b7` (webdesign.uk index) belongs to a **blocked lane** — see §5.

### 2. Retire the shim — `311_retire_the_192_wrapper_shim_path.sql` is WRITTEN, NOT APPLIED

**Only after step 1 passes.** The file is ready at
`docs/agent_docs/sql_for_agents/311_retire_the_192_wrapper_shim_path.sql`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/311_retire_the_192_wrapper_shim_path.sql
```

Then **record the number the same minute**, or it stays pending for ever and the next
`--apply` replays it (this is a standing landmine, and I tripped it on 308):

```bash
./scripts/migration/run-migrations.sh --record-only 311_retire_the_192_wrapper_shim_path.sql \
  --note '<what you verified>'
```

> **Why 311 and not 309:** the number was claimed from the **LEDGER**, not from `ls`.
> `schema_migrations`' highest was 308 (mine); 309 and 310 already existed **on disk** from
> other lanes, unrecorded. A number on disk is not a number taken.
> **Why removal is by VALUE not by index:** seed **309** (another lane —
> `page_content_writer_plans_its_own_sections`, unapplied as of writing) **appends a fourth
> path** `section_plan.sections_ready` to the same array. It appends idempotently and does
> **not** touch `required`, so it is compatible — but whether it lands before or after 311
> is unknowable from here, and an index-based delete would remove whichever element
> happened to sit there. 311 filters on the literal and asserts outcomes, not a length.

### 3. Re-check `bugs_open/178` is not regressed (V4)

Dispatch an `edit_live` `content_rewrite` and assert the writer child carries
`input_data.section_plan.sections_ready[0].existing_content_html` **at the FLAT path**, then
178's own check: `page_components.content_data` length grows by roughly the inserted anchor,
not a wholesale replacement. **This was blocked by the outage and is now unblocked** — items
`18bc832c` and `9e9ec430` (vetcomparison) both ran to `complete` at 09:05/09:2xZ and are
ready-made subjects. The 178 lane has been notified in their bug file.

### 4. Close it

Bar is **fixed AND live**. After steps 1–3: move
`bugs_open/192_HANDOFF_2026-08-04_select_sections_fallback_dies_on_a_null_link_resolution.md`
→ `bugs_closed/`, replace the STATUS banner with the closing evidence, and add the `016b`
**§10** index row. Note `git log` the FILE PATH, not the number.

### 5. Tell the webdesign.uk lane it is unblocked

`webdesign_uk_build_service` appended to `192` saying it is **blocked** and applying no
workaround — its shopfront landing page could not build. Items `5816c2b7` and `4f981a3d`
(both webdesign.uk / index) are still `failed` and **will not retry themselves**. I did not
re-dispatch them: it builds a live public page on another lane's site. Tell them, or let
them fire it.

## Things that will bite you

- **`orchestration_states` has no `id` and no `agent_type`** — it is `orchestration_id` and
  `owner_agent_type`. Schema before SQL.
- **Compare a run's `created_at` against a config change, never `updated_at`.** A failed
  run's `updated_at` is when it *died*. One run looked like a counter-example to the fix
  until I noticed it started two seconds *before* the seed committed.
- **Enumerate jsonb keys; do not read the path you expect.** `->'x'->'y'` returns NULL for
  "absent" and for "moved one level" alike. `jsonb_object_keys` distinguishes them, and that
  is what cracked this bug.
- **`/tmp` is a 16G tmpfs and was at 98–100%** (14G of unreaped session scratchpads, 92
  dirs). Go links through `/tmp`, so `make build-*` fails with
  `mapping output file failed: no space left on device` while `df /` shows ~185G free.
  Landmine filed. Do **not** sweep `/tmp/claude-1000/*` blind — many dirs belong to live
  sessions.
- **Pod-grep needs one `grep` per `exec`**; batching several into one `sh -c` times out on a
  binary this size. The image has no `strings` — use `grep -ac` directly, and pick **long**
  literals (Go compiles short ones to immediate comparisons that never reach rodata).
- **A negative control was not constructible here**, and that is recorded: every message I
  changed I *extended*, so each old literal is a prefix of its replacement and greps 1 on
  both old and new binaries. If you want the cheapest deploy proof next time, **replace** a
  string rather than extend it.

## Open, not mine, deliberately not chased

- **The overnight `process_sections_loop_iter_N_generate_content` failures** (~38 runs,
  21:00–01:00 on 08-03). `192`'s filing counted these as this bug; they are a **different**
  step, reachable only *after* `select_sections` succeeds. Still undiagnosed, nobody on it.
- **`research-agent` has never made an LLM call in the 4.5 months `llm_call_log` retains**
  (0 rows under any spelling — the relabelling trap was checked — against 18,590 for
  `page-content-writer`), yet it is spawned on **every** page build by
  `page-content-writer.spawn_research_agent`. Stated precisely: it has never made an LLM
  call, **not** that it has never run. Not filed — I have not diagnosed it and would be
  filing a symptom.
- **`RFC_012`** now carries the architectural half, with the working detector specified and
  the finding that **both naive detector versions return 0 on the bug that motivated them**.
  The `bug_historian` seat recommends a human treat that addendum as a required follow-on.

## Paper trail

`PLAN_`/`RUNBOOK_`/`NOTES_`/`README_where_we_are`/`SUMMARY_2026-08-04_` in this directory ·
`WRONG_CALLS.md` (3 entries: the filing's pooled failure modes, the path-read that could not
see the shape change, and my own right-number-wrong-reason census) · `LANDMINES.md` (2
entries: `output_field` replaces-not-annotates, and the tmpfs build failure) · `016b` §9 ·
register **WFA-009** · `RFC_012` addendum · notices appended to `bugs_open/178` and
`bugs_open/087`.

Commits: `a0e3ecee8` (claim+diagnose) → `2b9d84072` (fix) → `c1bf10f7e` (docs) →
`8db1135cf` (status) → `f94ee4996` (notices) → `e950e5a76` (runbook+summary) →
`5f569bb8e` (016b) → `0e518cfc9` (2nd proof) → `9abef444e` (revise r1) → `18d03ce0e`
(census correction) → `4228f5b98` (2nd-instance tests) → `e6e6d7ae5` (approved) →
`c5b36b962` (pod-verified).
