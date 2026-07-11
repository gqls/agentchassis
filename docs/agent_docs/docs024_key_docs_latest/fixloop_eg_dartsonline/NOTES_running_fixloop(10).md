# RUNNING NOTES — Diagnosis→Fix Loop (v1)

Chronological; newest entries appended under DISCUSSION LOG; decisions
promoted to DECISIONS with rationale. Continues NOTES_running_fixloop(9).md
(entries up to 2026-07-07 are there and are not repeated).

## 2026-07-09 — new thread opened; pilot pre-check run FIRST

### Turn 1 — context loaded, pre-checks executed, symptom did not survive intact

Inherited state read: RUNBOOK_diagnosis_fix_loop(9).md, NOTES_running_fixloop(9).md,
HANDOFF_fixloop_thread(8).md, z_bundles/BUNDLE_fixloop_F0.md.

**Bundle confirmed deficient as the handoff warned.** BUNDLE_fixloop_F0.md
(199,579B) carries `## Schema — _none provided_`, `## Database capabilities —
_none provided_`, `## Runtime evidence — _not available in the thin slice_`.
It was generated without `-psql`, so it is code+docs only. For this pilot the
DB half is where the answer lives. Regenerate before any loop run that is
supposed to consume it.

**The three ★ pre-check queries were run against the live cluster** (psql via
`kubectl exec -n ai-persona-system postgres-clients-0`). One needed
correction: `site_work_items` has no `attempts` column — it is `attempt_count`.
The runbook's pre-check SQL is wrong on that column and is corrected in
RUNBOOK(10).

**The pre-check did not sharpen the symptom. It dissolved it, and then opened
a bigger one.** Findings, in the order they arrived:

1. dartsonline has exactly **one** guide-ish page row: `guides-index`,
   `page_type='section-index'`, `build_status='planned'`, `in_header=true`.
   There are **no** `page_type='guide'` rows at all.
2. gamesdesign.co.uk has 5 × `guide` + 1 × `guides-index` (`section-index`),
   **all `deployed`**. The differential is real.
3. Widening to the whole site: dartsonline's page matrix is
   `content` (3) and `landing` (2) → **deployed**; `blog-post` (4),
   `entity-directory` (2), `entity-page` (2), `section-index` (1), `tool` (1)
   → **all `planned`**. So this was never a guides bug. **Ten of fifteen
   pages were never built**, and guides is simply the one that got a nav link.
4. `blog-post` is *in* the hypothesised routing table and still was not built
   → the standing hypothesis was already in trouble at this point.
5. The `needs_page` work items for those pages **all exist and all completed**
   (`status='complete'`, `handler_agent='page-build-handler'`,
   `attempt_count=0`). Twenty-three of them. The system marked as complete the
   construction of pages it never constructed.
6. The `result` payloads split the set cleanly. The 5 deployed pages carry
   `{"response":{"deploy_result":{... "files":["/about.html"] ...}}}`.
   The 10 unbuilt pages carry `{"response":{"site_record":{...}}}` or a bare
   design-tokens blob (`{"spacing":…,"typography":…}`) — **no `deploy_result`
   field at all**.
7. `pages.sections` is the discriminator, and it is an **exact partition, 5 v 10,
   no exceptions**: `jsonb_array_length(sections) > 0` ⟺ `deployed`;
   `= 0` ⟺ `planned`. gamesdesign's guide and section-index pages each carry
   `sections` (2 apiece) and are deployed — **through the same
   `page-build-handler`**. The handler is not the discriminator. Sections are.
8. A name mismatch sits alongside it: the imagery flow emitted
   `page_rerender:guide-barrel-weight` … `guide-steel-tip-vs-soft-tip` items
   (slugs derived from hero assets `hero_guide_barrel_weight` etc.), but the
   plan and `pages` name those rows `barrel-weight`, `beginners`,
   `flight-shapes`, `steel-tip-vs-soft-tip`. Those rerenders targeted pages
   that do not exist under that name, no-op'd, and completed.

### Turn 1 (cont.) — mechanism found in code; standing hypothesis REFUTED

Read `reconcile_site_plan_action.go`, `load_work_item_actions.go`,
`populate_nav_tables_action.go`, and the **live** `page-build-handler`
workflow JSON out of `agent_definitions`.

**Cause B — the silent success.** The live `page-build-handler` workflow
(v1, active) runs `plan_sections` → `check_has_ready_sections`, a `conditional`
with `condition: "section_plan.ready_count > 0"` and
`else_step: "complete_error"`. `complete_error` is **`action: complete_workflow`**
— a *success* terminal — with
`success_message: "Content writer skipped — page has no sections defined"` and
`output_fields: ["page_content","site_record"]`. That output_fields list is
*exactly* the shape observed in the 10 unbuilt items' `result` (finding 6).
The dispatch loop then stamps `status='complete'`. The page row is never
touched: `updated_at == created_at` (12:39:05) on all ten.

**The platform already knows.** `load_work_item_actions.go:750-756`, a comment
on the completion guard, says in as many words: *"the dispatch loop calls
complete_work_item on every successful handler saga, and page-build-handler's
complete_error is a SUCCESS-labelled complete_workflow"* — and names the
remedy: *"mark_no_sections for a sectionless page with no sibling layout"*
would flag the item `needs_human_review`. **`mark_no_sections` does not exist.**
It is absent from the live workflow's 18 steps and appears nowhere in the repo
except that one comment (`grep -rn mark_no_sections` → 1 hit, the comment).
The guard faithfully preserves a flag that nothing ever sets.

**Cause C — nav grounded in the wrong column.**
`populate_nav_tables_action.go:242-243` selects
`FROM pages WHERE site_id = $1 AND status IN ('active','deployed','pending')`.
`pages.status` is a *lifecycle* column defaulting to `'active'` on insert.
`build_status` — the actual build state — is **never consulted**. So
`guides-index` (`build_status='planned'`, `status='active'`, `in_header=true`,
`page_type='section-index'`, which is not in the `neverPrimaryTypes` set
`{blog-post, tool, entity-page}`) is published straight into the primary nav.
That, precisely, is "the system linked to something it never built".

**Cause A — the planner under-populates sections.** `build-site-planner` wrote
15 plan pages into `site_plan_pages` but authored `sections` for only the 5
`content`/`landing` ones. Everything downstream is a consequence.

**THE STANDING HYPOTHESIS IS REFUTED — and it named the wrong file.** The
routing table is real but it is in **`WriteBuildItemsAction`**
(`load_work_item_actions.go:218-228`), *not* in `reconcile_site_plan`.
`guide` and `section-index` are indeed absent from it and `tool`/`entity-page`
are indeed commented out. But absence from that map **does not drop a page**:
`:239` defaults `handlerAgent = "page-build-handler"`, and `:283` logs
`"Unknown page_type, using page-build-handler"` before falling through. What
*does* drop pages is the separate `unavailableBuilders` map (`:233-237` —
`tool`, `entity-directory`, `entity-page`) which hits `continue // Skip — don't
create a dispatch work item`. And `reconcile_site_plan_action.go:213-217`
hardcodes `handler_agent='page-build-handler'` for **every** plan page with no
type switch whatsoever — so it has no `unavailableBuilders` guard and re-emits
items for the very types the other path deliberately skips.

So the guides nav link is caused by **B + C**, not by the routing table. The
hypothesis was directionally right that "nav outlives a dropped page" and
wrong about every mechanism it named. Evidence over assumption, exactly as the
intake instructed.

**Fourth finding (new, unlooked-for):** the two intake paths disagree.
`WriteBuildItemsAction` skips `tool`/`entity-directory`/`entity-page`;
`reconcile_site_plan` emits `needs_page` for them anyway. dartsonline's
`shop-index`/`brands-index` (`entity-directory`) have completed `needs_page`
items that could never have built. gamesdesign's items are `needs_content_page`
(the `availableBuilders` path); dartsonline's are `needs_page` (reconcile).
**That is the build-route variable the runbook told us to establish by
evidence rather than assume. Established.**

### Turn 1 — consequence for the pilot (a judgement the owner must rule on)

Loop-worthiness criterion 3 says a task is loop material only if it is *not*
answerable by one or two direct queries. This took roughly a dozen queries and
four file reads — more than two, but it never required a hypothesis-driven
re-scoping loop. **Candidate 3 has now gone the way of candidates 1 and 2: the
cheap pre-check answered it.** Three for three. That is not a failure of pilot
selection; it is the pre-check doctrine doing its job, and it says something
structural about this platform — its bugs are mostly *legible* to someone with
schema access and grep.

But the conclusion this time is the opposite of "reject the pilot", because
**we now hold ground truth with citations across all three tiers.** A loop
cannot be graded on a bug whose answer nobody knows. It can be graded exactly
here. RECOMMENDATION: promote this from *discovery* pilot to **known-answer
benchmark** — the loop runs on the original symptom string, blind, and we score
its output against the findings recorded above. See PLAN_fixloop_pilot.md §3.

### Turn 2 — owner rulings; F0.1a LANDED (diagnosis_artifacts is live)

**Owner ruled: the F1 edit plan targets the PLATFORM**, not dartsonline's data.
Confirms the 2026-07-09 decision below and closes the only question turn 1 left
open. Owner also directed: proceed with the migration.

**Route taken.** Checked convention before inventing one.
`scripts/migration/run-migrations.sh` is **empty (0 bytes)** — there is no
migration runner and no `schema_migrations` version table in `clients_db`
(only `migration_backups`). Migrations are hand-applied via psql, and the
tools chat's live `doc_plans`/`doc_notes` still sit at `0NN_` in their own
workstream directory. So: same pattern, same directory, `0NN_` prefix,
`BEGIN`/`COMMIT`, `COMMENT ON`, an explicit manual-rollback footer, and a
companion pre-flight script.

**Pre-flight gates (`verify_before_migration_diagnosis_artifacts.sql`) — all
clean** on clients_db: no `diagnosis_artifacts` table; no collision on the three
index names (checked separately, because index names are schema-global and a
clean table check does not cover them); `site_work_items.pipeline` carries **no
CHECK constraint** and its live values are `build` (3758), `content` (24),
`design` (13), `maintenance` (2) — so F0.1c's `'diagnose'` namespace is free and
needs no schema change.

**Applied `0NN_diagnosis_artifacts.sql`.** Shape: `correlation_id` (text, *not*
uuid — `ExecutionContext.CorrelationID` is a string and the chassis does not
guarantee uuid form), `orchestration_id`, `iteration` (CHECK ≥ 1), `kind`
(CHECK ∈ {bundle, iteration_note}), `body`, nullable `site_id` (anchorless runs),
`metadata` jsonb, `source_agent`, `created_by`, `created_at`, plus the retention
knob `expires_at` + `pinned`. Three indexes: the `(correlation_id, iteration,
kind)` read path; a **partial unique** on `(correlation_id, iteration) WHERE
kind='bundle'`; a partial expiry index for the future sweep.

*Why the unique index is partial:* exactly one bundle per (run, iteration), so a
workflow step retry upserts rather than duplicates — but per-**step** notes mean
several `iteration_note` rows share one iteration, so notes must not be covered.

**F0.1a criteria — all seven pass.**
1. Applies clean. 2. **Idempotent**: re-applying the file is a no-op (`IF NOT
EXISTS` throughout; only harmless NOTICEs). 3. Round-trip: one row of each kind
inserts and reads back with `metadata` intact. 4. The `kind` CHECK **rejects** a
third kind (`'verdict'`). 5. The `iteration` CHECK **rejects** `0`. 6. The
partial unique **rejects** a second `bundle` for the same (run, iteration)…
7. …while **allowing** a second `iteration_note` for that same iteration.
Plus, verified because the Go write depends on it: `ON CONFLICT
(correlation_id, iteration) WHERE kind='bundle' DO UPDATE` correctly infers the
partial index and **replaces** the body, leaving one row. That exact conflict
clause is what F0.1b must use. Self-test rows deleted; table is empty.

**Reconnaissance for F0.1b, done while here.** `DiagnoseAssembleBundleAction`
(line 107) already has `params.DB` and `params.ExecutionContext` in hand, and
returns its map at line 219 — the write-through goes immediately before that
return, no signature change, no workflow-shape change. The iteration number is
derivable without new plumbing: `diagnose_route` writes
`route.diagnose_state` (`pkg/diagnose/advance.go:23`, json tag `iteration`), and
assemble runs *before* route each pass, so
`iteration = route.diagnose_state.iteration + 1`, defaulting to **1** when
`diagnose_state` is absent (the genuine first pass). Note the trap
`diagnose_route` documents at its `state_field` default: a bare `diagnose_state`
never exists at top level — it must be read at `route.diagnose_state`.

### Turn 3 — F0.1b written, built, tested; NOT YET DEPLOYED

**The write-through lives inside `DiagnoseAssembleBundleAction`**, immediately
before its existing return. No signature change, no workflow-shape change, no
new step — exactly the placement Q-A chose to keep off the tools chat's active
`emit → persist_note → complete` surface.

Four config knobs added to the InputSpec (all optional, all defaulted):
`persist_bundle` (true), `iteration_field` (`route.diagnose_state.iteration`),
`site_id_field` (`input_data.site_id`), `bundle_retention_days` (30; ≤0 keeps
forever). Nothing existing changed shape.

**The read-only contract is preserved and the degradation rule is enforced.**
The loop's read-only contract concerns the *system under diagnosis*; this writes
only our own artifacts table. Even so, persistence is observability, and
observability must never cost a diagnosis — so a nil DB handle, a missing
correlation_id, a marshal failure or a failed INSERT each log a warning and the
action returns its bundle normally. There is no error path back to the loop.

**Reuse caught before recreate.** I had written a `nullIfEmpty` helper; the
package already has one at `helpers.go:133` with an identical signature. Deleted
mine. (Precisely the "we already have one of these" case the future reuse
reviewer is meant to catch — worth noting that it was a grep, not a memory.)

**A wart the verification found, not the compiler.** The first `DO UPDATE` set
refreshed only `body`/`metadata`/`expires_at`. Executing the real SQL showed a
retry that supplied `orch-9` leaving `orchestration_id` NULL on the stored row —
the row would then misreport which orchestration last wrote it. `DO UPDATE` now
also refreshes `orchestration_id`, `site_id` and `source_agent`. Found by
running the statement, not by reading it.

**Verification actually performed** (not "it compiles"):
- `gofmt` clean; `go build ./platform/orchestration/actions/` OK; whole-repo
  `go build ./...` OK apart from a **pre-existing** stray-package clash under
  `docs/.../traffic_probe/deploy_setup/working_dir` (two `package` names in one
  dir), untouched by this change.
- `go test ./pkg/diagnose/... ./platform/orchestration/actions/...` — all pass.
- **New unit test** `diagnose_assemble_iteration_test.go` pins `assembleIteration`:
  first pass → 1 (never 0, which the CHECK rejects); `float64(1)` → 2 (JSON
  decodes the LoopState, so the value is float64, not int); int → works too; and
  two cases locking down **the trap** — a bare `diagnose_state` path does *not*
  resolve and silently collapses every bundle to iteration 1 (where the partial
  unique index would upsert them into a single row), while the same data at the
  correct `route.diagnose_state.iteration` path resolves. The test exists so
  nobody "simplifies" the default field path.
- **The exact production SQL executed against the live table** via
  `PREPARE`/`EXECUTE` with typed parameters — because the risk here is in the
  SQL, not the Go. Confirmed: `ON CONFLICT … WHERE kind='bundle'` infers the
  partial index; a retry of iteration 1 **replaces** (3 inserts → 2 rows, then
  2 → 1); NULL `site_id` works for an anchorless run; `make_interval(days => 30)`
  gives a 30-day expiry and retention `0` yields NULL (keeps forever). Test rows
  deleted; `diagnosis_artifacts` is empty.

**⚠ NOT LIVE IN THE CLUSTER.** The `agent-chassis` pod runs a built binary.
This change is in the repo and green, but no bundle row will appear until the
image is rebuilt and rolled out. F0.1b is *code-complete*, not *deployed*. The
first real end-to-end proof (rows appearing for an actual run) arrives with the
benchmark run, and that ordering is fine — but do not read "verified" as "live".

**Unrelated working-tree changes observed, not mine, not touched:**
`platform/orchestration/actions/plan_sections_action.go` (+80 lines) and the
untracked `actions/discovery_checks/check_image_source_unsatisfiable.go`. Some
other session's work in progress. Flagged so a future commit does not sweep them
in by accident. (`plan_sections` is, note, the very step upstream of
`check_has_ready_sections` in the pilot diagnosis — worth a look before F1
edits that area.)

### Turn 4 — F0.1c LANDED; Q-B's "null-site" was impossible; reuse found instead

**Q-B, as decided on 2026-07-07, could not be built.** It said site-less code
bugs "ride null-site in the new `pipeline='diagnose'` namespace". Reading the
schema before writing the code killed that in two independent ways:

1. `site_work_items.site_id` is **NOT NULL**. The runbook assumed nullable.
2. Even with the constraint dropped it would not work: `LoadWorkItemsAction`
   parses `site_id` as a **required uuid** and its query is
   `WHERE wi.site_id = $1`. The relay's loader is site-anchored *by
   construction*, and `site_id = NULL` matches nothing regardless. A NULL-site
   item would sit in the table forever, invisible.

**The platform had already solved it — we just hadn't looked.** `system.internal`
(`eac60db8-b032-432b-b36d-76f37632045d`, `sites.status='system'`) is an existing
pseudo-site holding platform-wide work: the `maintenance`-pipeline
`component_quality_scan` items live there today, alongside 45
`needs_component_regeneration` rows. So: anchor `needs_diagnosis` to
`system.internal`. No migration, no constraint change, no second mechanism.
Reuse before recreate — and this is the second time in two slices that grep beat
memory (the first being `nullIfEmpty`).

**A hazard found while doing it, worth more than the slice.** The live
`build-dispatch-loop`'s `load_items` step is configured with **only**
`{site_id, max_items}` — it has **no `item_pipeline` filter**. So a work item of
*any* pipeline, parked on a real site, is claimed by that site's next build
dispatch and handed to whatever `handler_agent` it names. Two consequences:
- **Every** `needs_diagnosis` item anchors to `system.internal`, even when the
  bug is about a real site. The site under diagnosis travels in
  `spec.site_id` / `spec.runtime_site`, never in the item's own `site_id`
  column. This keeps the diagnose namespace physically off every per-site
  dispatch loop **without editing a workflow the builder thread owns** — the
  collision rule holds.
- Belt and braces: the item is written with `status='detected'`, and the loader
  takes only `status IN ('triaged','approved')`. Nothing existing can claim it
  even if a loop is one day pointed at `system.internal`.

**Deliverable:** `090_TRIGGER_needs_diagnosis_v1.sh` (085–089 were taken). It
writes the durable intake record, then fires the 084 envelope on the **same
correlation_id**, so the work item, the `diagnosis_artifacts` bundles, and the
terminal `doc_notes` row all join on one key. `DISPATCH=0` records without
firing. 084 is *not* replaced — Q-B kept the bare manual trigger for ad-hoc runs
with no intake record.

**F0.1c criteria — all pass** (`bash -n` clean; run with `DISPATCH=0`):
1. Item inserts with `pipeline='diagnose'`, `item_type='needs_diagnosis'`,
   `handler_agent='diagnose-orchestrator'`, anchored to `system.internal`;
   `spec` carries symptom, ref, target `runtime_site`, `subject_type`/
   `subject_key` and the `correlation_id`.
2. **Not claimable**: the loader's *exact* query, run against `system.internal`,
   returns 0 rows. (0 rows is decisive here only because the preceding query
   proved the row exists — the "0 rows isn't decisive until the query is
   checked" rule, honoured.)
3. **Negative control**: flipping the row to `status='triaged'` makes it appear
   in that same loader query immediately. So non-claimability is caused by the
   status guard, not by a typo in the query — and the build-dispatch hazard
   above is demonstrated, not merely argued.
4. **Idempotent**: re-running the same `SLUG` with a different symptom while the
   intake is open yields `INSERT 0 0` and still one open intake — `ON CONFLICT
   DO NOTHING` pairing with `idx_swi_dedup (site_id, item_key) WHERE status NOT
   IN (terminal…)`.
Self-test rows deleted; `pipeline='diagnose'` is empty again.

**Still open, deliberately:** automatic dispatch of `pipeline='diagnose'` needs
its own pipeline-filtered loop (a new agent definition, or an `item_pipeline`
filter added to `build-dispatch-loop` — the latter touches the builder thread's
surface and should be *their* call). Until then this script is the dispatcher,
and that is the documented route F0.2 asked for. Flagged rather than quietly
built.

### Turn 5 — automatic dispatch built; owner's challenge exposed a design flaw

Owner is deploying `v1.0.1100` (`make release redeploy-agents ENVIRONMENT=production
REGION=uk001`) and asked for automatic dispatch of the `diagnose` pipeline here.

**Why the standard claim path cannot carry this namespace** (all read live, none
assumed). The relay cannot tell pipelines apart at the point it matters:
- `build-dispatch-loop`'s `load_items` has only `{site_id, max_items}` — **no
  `item_pipeline` filter**.
- `build-pipeline-trigger`'s `find_dispatchable_site` has **no pipeline filter
  either**, though its description says "a site with pending *build* items". Any
  site holding a `triaged`/`approved` item of any pipeline is selected.
- **This is how the `maintenance` pipeline gets dispatched at all — by accident.**
  So the tempting one-key fix (add `item_pipeline='build'` to
  `build-dispatch-loop`) would **orphan the maintenance pipeline**. Reported to
  the builder thread; deliberately not fixed here.
- `triage_detect_items` promotes `WHERE site_id=$1 AND status='detected'` — no
  pipeline filter — and **rewrites `pipeline` to `'build'`**. Its comment claims
  "the dispatch loop (which filters item_pipeline='build')". It does not. Same
  comment-vs-code family as the pilot bug. So `'detected'` is not safe parking.
- `claim_work_item` claims only `triaged|approved` — precisely the statuses that
  expose an item to the two unfiltered readers above.

**Design: two private statuses, inert by construction.** Queued =
`awaiting_diagnosis`; in-flight = `diagnosing`. Every sweep filters on explicit
status values, so unknown values are ignored by construction rather than by luck
of anchor-site choice. Deliverables: `0NN_diagnose_dispatch_loop.sql` —
`diagnose-dispatch-loop` agent (image columns copied from the `build-dispatch-loop`
donor per the seed gotcha; tag already `v1.0.1100`) + `diagnose-pipeline-trigger`
scheduled task, **shipped `enabled=false`**.

**★ OWNER'S CHALLENGE — "why are we setting claimed status; the handler will pick
that up when ready" — half wrong, half a real bug I had missed.**
- *The premise*: on this platform the **dispatcher** claims, not the handler.
  Only `build-dispatch-loop` (and now ours) reference `claim_work_item`;
  `page-build-handler` neither claims nor completes its own item — the dispatch
  loop calls `complete_work_item` on its behalf (`load_work_item_actions.go:750`).
  Verified before answering. Without an atomic move out of the queue state, the
  next 60s tick re-dispatches the same 26-minute LLM run.
- *The flaw it exposed*: I argued for a private, inert status and then moved the
  item into **`claimed`** — the one status that walks straight back into the
  sweep surface I had just escaped. Two live consequences:
  (a) `claimed-item-timeout` resets any 40-minute-old claim to **`triaged`**,
  handing a slow diagnosis to the build dispatcher — I had "fixed" this with
  `max_attempts=1`, a workaround, not a design;
  (b) `find_dispatchable_site` excludes any site holding a `claimed` item, so a
  70-minute diagnosis would have **blocked build dispatch for `system.internal`
  for its entire duration** — unasked-for cross-pipeline interference I had not
  spotted at all.
- *Fix*: the in-flight status is now **`diagnosing`**. `claimed_by`/`claimed_at`
  are still stamped for audit. Because `diagnosing` is (by design) invisible to
  `claimed-item-timeout`, the loop now **reaps its own dead runs**: a new
  `reap_stuck` first step fails any `diagnosing` row older than 75 minutes (past
  the workflow's own 4200s timeout). `max_attempts=1` is kept, but as semantics —
  a 26-minute LLM loop should not silently auto-retry — not as the safety net.
- The 15-minute auto-complete branch is safe either way: it is gated on
  `item_type IN (needs_content_page, page_rerender, needs_design)`.

**Verification (live, then cleaned up).** `snapshot_agent` taken before the
update — it writes to `agent_definitions_backup`, *not* `agent_definitions` with
`is_snapshot=true`, so my first check looked in the wrong table; the pre-update
workflow (`start_step=claim_item`) is captured there. Then:
- **Inertness matrix**: the item at `awaiting_diagnosis`, and again at
  `diagnosing`, scores **0 hits** against all six sweep predicates —
  `claim_work_item`, `load_work_items`, `triage_detect_items`,
  `feasibility-recheck`, `stale-work-item-reaper`, and (the one that matters)
  `claimed-item-timeout`.
- **Positive control**: our own claim predicate finds it (1 hit). Inertness only
  means something if *someone* can still claim it.
- `find_dispatchable_site`'s exact query returns robot-hands.com, not
  `system.internal`; and `blocks_system_internal` is **false** with our item
  in-flight — the interference is gone.
- The workflow's exact `claim_item` statement claims one row, returns the
  envelope with `target_site_id=dartsonline` (**not** the anchor), and a second
  tick returns 0 rows → `claimed.count=0` → `notify_scheduler`.
- `reap_stuck`'s exact statement turns a 76-minute-old `diagnosing` row into
  `failed`, `attempt_count=1`. Terminal, never `triaged`.

**Also fixed this turn:** 090 emitted `seed_scope` as a comma string.
`ExtractStringListHelper` accepts only `[]interface{}`/`[]string`, so it would
have parsed to nil and the seed silently ignored. Now emitted as a JSON array
(verified: `jsonb_typeof = array`). The auto path deliberately does **not**
forward `seed_scope` — `query_database` flattens every column to text — so
assemble falls back to `code_results`, which is the designed default. Documented,
not hidden.

### Turn 6 — ★ THE BENCHMARK RAN. Plumbing passed. The loop FAILED the rubric — instructively.

Fired 090 at 15:54:23 UTC, `REF=main`, `RUNTIME_SITE=dartsonline.com`,
`SUBJECT_TYPE=pipeline SUBJECT_KEY=build`, **no `SEED_SCOPE`** (blinding).
`correlation_id=4d43d002-671f-496f-a64a-c3bb8ffe35e2`. Ran 5 iterations,
15:58 → 16:22, verdict **CONFIRMED**, `stopped_by=confirmed`.

**Timeline gotcha, nearly a false alarm.** The first poll appeared to show a run
that had completed 50 minutes *before* it was fired. Cause: the host is on **BST
(+0100)** while the DB is UTC, so `stat`'s `16:54 +0100` is `15:54` UTC. Worse,
`orchestration_states.last_activity` is `timestamp WITHOUT time zone` while
`created_at` is `timestamptz` — so the `idle_s` arithmetic in 084/090's own
suggested query is silently wrong by the UTC offset. Recorded as a gotcha.

### F0 plumbing criteria — 3 of 4 pass, and the fourth isn't built yet
1. **Intake via the documented route — PASS.** Item written, dispatched, and
   closed on one `correlation_id`.
2. **Per-iteration bundles fetchable — PASS.** Five `kind='bundle'` rows
   (34.8k / 45.2k / 33.4k / 18.7k / 33.2k bytes), one per iteration, fetched back
   out to files by the documented query. **F0.1a+b work end to end in production.**
3. **Per-iteration notes — NOT MET, because F0.3 does not exist yet.** Only
   `kind='bundle'` rows were written. Not a failure of the run; a slice we have
   not built. The table already carries the `iteration_note` kind for it.
4. **Terminal note in doc_notes — PASS (unplanned bonus).** One row,
   `subject_type=pipeline`, `subject_key=build`, `categories=["diagnosis"]`,
   `source='diagnosis-loop'`, 3,267 bytes. **Q-F's integration is verified live**:
   our envelope's subject fields opened the tools chat's persist_note gate. Their
   3b threading has landed (`diagnose-orchestrator` now forwards
   `subject_type?`/`subject_key?`).

### The rubric — 0 of 4 musts. It confirmed a cause that does not explain the symptom.
| # | claim | result |
|---|---|---|
| 1 | `pages.sections` empty ⟺ not built (the exact 5/10 partition) | **partial** — proved `sections=[]` + `planned` for the five guide-ish pages; never contrasted the five *built* pages, so never established the ⟺ |
| 2 | `check_has_ready_sections` routes sectionless pages to `complete_error` | **not reached** |
| 3 | `complete_error` is a `complete_workflow` — a success terminal | **not reached** |
| 4 | nav selects on `pages.status`, not `build_status` | **not reached** |
| 5 | work-item `result` lacks `deploy_result` | abstained (neutral) |
| 6 | gamesdesign uses the same handler ⇒ handler isn't the discriminator | abstained — it never saw gamesdesign; `RUNTIME_SITE` scoped evidence to dartsonline alone |

**Refutation credit: PASSED.** The conclusion does not mention `reconcile` or a
`routing table` (checked directly). It even pulled `reconcile_site_plan_action.go`
into iteration 5's scope and declined to blame it. A citing loop that had
confirmed our false standing hypothesis would have been worse than useless; it
didn't.

**It emitted `status=CONFIRMED` anyway.** This is the benchmark's single most
important output: **cite-or-abstain does not prevent confirming the wrong cause.**
Every citation it gave is real and checks out. The cause it assembled from them
simply does not explain the symptom in the intake string — nothing in its
conclusion accounts for *why a nav link was published*. The loop needs a
"does the cited cause explain the reported symptom?" gate before it may say
CONFIRMED. Nothing in the current guards does that.

### Why it missed — two mechanical causes, both fixable, both verified
**(a) The answer is not in the corpus.** Causes B/C live in
`agent_definitions.default_config` — a **workflow JSON blob in the database**.
`code_symbols` indexes `.go` files only (verified: every indexed path ends `.go`).
So the loop's static tier *structurally cannot* reach the `check_has_ready_sections
→ complete_error` routing. It could have got there with a `data_request` against
`agent_definitions`, and nothing prompted it to.

**(b) Right file, wrong symbol.** Iteration 1's scope contained
`populate_nav_tables_action.go:isLegalPage` — a trivial helper. The nav bug is in
**`loadPagesForNav`** (same file, line 229; the `status IN ('active','deployed',
'pending')` query at line 243). `loadPagesForNav` never entered **any** bundle
(grepped all five). Symbol-granular retrieval found the right file and handed the
verdict the wrong function.

**And the pointer was in front of it, five times.** Every bundle's
`agent_error_log` section contains, verbatim:
`page-build-handler/complete_error (complete_workflow) fatal: …`.
The mechanism's exact name — *and the fact that it is a `complete_workflow`* — sat
in the runtime evidence of all five iterations. The loop never followed it. The
"FOLLOW what the evidence names" principle is stated in the runbook and is not
enforced anywhere in the engine.

### What the loop found that WE missed — a real, verified addition to the diagnosis
It established, and I confirmed independently, that the four "guide" pages are not
guides at all: `barrel-weight`, `beginners`, `flight-shapes`,
`steel-tip-vs-soft-tip` are `page_type='blog-post'` at **`/blog/*.html`**, with
`site_area_id IS NULL`, while `guides-index` is `section-index` at
`/guides/index.html`, also with `site_area_id IS NULL`. It cited
`queryresolve/section_index_for.go:resolveSectionIndexForType`, whose area-based
lookup would therefore resolve to the *blog* section-index, never to
`guides-index`. **So even had the pages built, they would never have appeared
under /guides/.** The nav link points at a section whose children were planned
into a different section. Our ground truth did not contain this. Credit where due:
the rubric should gain it as claim 9.

### Turn 7 — owner: "think hard about this change." Done — and it changed the diagnosis of the loop's failure.

Pressure-tested the three proposed engine fixes against the engine code and the
five persisted bundles before building anything. Two empirical findings:

**FINDING 1 — the verdict never sees the original symptom after iteration 2.**
Proved from the persisted bundles (the artifacts table paying for itself):
`grep 'Guides nav link'` → present in iter 1–2 (seed hypothesis = symptom),
**absent from iter 3, 4 and 5**. `diagnose_assemble_bundle` puts ONE hypothesis
at the top of the bundle — `route.hypothesis` (the revised one) once the loop has
iterated — and the verdict prompt says "judge whether the evidence confirms THE
HYPOTHESIS". By iteration 5 the hypothesis had drifted to "the four guide pages
exist as blog-posts, unbuilt, structurally disconnected" — which is TRUE, and the
evidence directly supports it. **The verdict behaved correctly per its contract.
The loop confirmed the right answer to the wrong question.** My turn-6 framing
("cite-or-abstain does not prevent confirming the wrong cause") was off: the
defect is *hypothesis drift with no symptom anchor*, which is sharper, smaller,
and mostly fixable in Go on our side of the collision boundary.

Crucial nuance: drift is a FEATURE — the verdict prompt's own worked example
celebrates following evidence to "a symbol the symptom could never have named".
The fix is to keep the symptom VISIBLE and require the terminal confirm to close
the loop back to it — anchor, don't clamp.

**FINDING 2 — the three-tier doctrine is not enforced.** The engine's ONLY
confirm guard is `len(Citations)==0 → Unverifiable` (advance.go:93–94). No
tier-coverage check exists anywhere in `Advance`. "Citations across all three
tiers" is runbook doctrine and prompt convention only. (This run happened to cite
all three tiers, so enforcement would not have changed this outcome — but a
future single-tier CONFIRMED would sail through today.)

**Ownership fact that shapes everything:** the verdict `prompt_template` lives in
the diagnose-agent workflow JSON in `agent_definitions` — the tools chat's
surface (fetch-first + coordinate; their 3b has landed so it is quieter now, but
the prompt itself declares "the output schema MUST stay in lockstep with
verdict_wire.go" — wire changes and prompt changes are one atomic, coordinated
change). Everything else below is Go-side, in our lane.

**The revised fix set (F0.4), with ownership and cost:**
- **F0.4a — symptom anchor** (Go, ours, ~5 lines): assemble always renders
  "## Original symptom" (from `input_data.symptom`) above "## Hypothesis under
  test". Restores the invariant; costs nothing.
- **F0.4b — follow-the-error-log enrichment** (Go, ours): parse
  `agent/step (action)` patterns out of the runtime evidence and inline the named
  workflow step's JSON from `agent_definitions` into the bundle (capped).
  Mechanically enforces "follow what the evidence names" for the commonest
  pointer, and bridges the Go-only corpus gap — cause B lived in exactly such a
  step, named verbatim in all five bundles.
- **F0.4c — same-file sibling signatures** (Go, ours): a signatures list for
  in-scope files, from the `repo_analysis` already in collected_data (parity with
  contextkit's Neighbourhood section, which the in-cluster assembler lost).
  `isLegalPage` vs `loadPagesForNav` was this gap.
- **F0.4d — symptom-closure gate** (wire = ours; prompt = coordinated): CONFIRMED
  must carry a `symptom_check` mapping the confirmed mechanism to the original
  symptom's observations, or the outcome downgrades (new partial status).
  Verdict_wire + its tests + prompt_template change atomically; snapshot_agent
  first; courtesy FYI to the tools chat.
- **F0.4e — tier-coverage guard** (Go, ours, pure `Advance`): Stop(confirmed)
  requires ≥1 static AND ≥1 state|runtime citation, else treat as
  Unverifiable-and-continue. Enforces the stated doctrine.

**Overfitting check (asked and answered):** none of a–e mentions guides or
dartsonline. (a) restores an invariant; (b) implements the runbook's own stated
principle; (c) restores parity the port dropped; (e) enforces stated doctrine.
(d) is the only genuinely new behaviour and it generalises ("a diagnosis must
explain the symptom" is not benchmark-specific).

**Risks named:** bundle growth from b+c (cap each section; degrade by truncation
— max_body_chars covers bodies only today); assemble is shared by every diagnose
run (new sections must be append-only and failure-degrading, same pattern as the
write-through); d changes what CONFIRMED means — and F1's fixer gates on
CONFIRMED, so d precedes building the fixer mechanism.

**Benchmark comparability:** run 2 must use the IDENTICAL symptom string and the
site's data must stay untouched (it has — we have not fixed dartsonline). Change
one variable cluster per run: run 2 = a+b+c+e (Go-side), measure; run 3 = d
after coordination. Also note our symptom string is itself compound (three
observations) — against worthiness criterion 4's letter — keep it unchanged for
comparability, but F0.4d's design should decompose symptoms into observations at
intake.

**F1 split (the reorder, resolved):** fixing the dartsonline platform bug
(mark_no_sections + nav build_status) needs no loop — the diagnosis is
human-confirmed and the site is broken now; it can proceed any time. Building the
F1 *fixer mechanism* waits for d, because a fixer keyed on today's CONFIRMED
would act on wrong-cause confirmations.

### Turn 8 — F0.4 a/b/c/e BUILT AND TESTED; awaiting image for run 2

**F0.4e — tier-coverage guard (pkg/diagnose).** The two duplicated item-24
coercion blocks (DecideStep + Advance's trail record) — plus a THIRD copy found
in `Run()`'s trail append — are now one shared `coerceVerdict()`, extended with
the new rule: CONFIRMED must carry ≥1 `static` AND ≥1 `state|runtime` citation
or it degrades to Unverifiable and the loop continues gathering. Refutation is
exempt on purpose (prompt rule-3 asymmetry: one contradicting log line
legitimately breaks a hypothesis; a confirm must show the mechanism AND its
occurrence). Five existing test fixtures confirmed on a single tier — including
the wire-format replay of the real gamesdesign path — and now correctly failed;
each was updated to carry both families, which IS the new contract. New
`step_tierguard_test.go` pins: four single-family confirm shapes all degrade;
static+state and static+runtime both stand; single-tier REFUTE still stands;
and Advance's trail records the COERCED outcome (decision/trail no-drift).

**F0.4a — symptom anchor (assemble).** New `symptom_field` config (default
`input_data.symptom`); the bundle now opens with "## Original symptom (the
question this diagnosis must answer)" whenever it differs from the hypothesis
under test — i.e. exactly from iteration 2 onwards, where run 1's bundles lost
it. Drift stays allowed; the question stays visible.

**F0.4b — follow-the-error-log enrichment (assemble).** A regex over the
runtime evidence extracts up to 4 distinct `agent/step (action)` refs (the
agent_error_log line shape); each named step's JSON is fetched from
`agent_definitions` (`ORDER BY version DESC LIMIT 1`, live rows only) and
inlined under "## Workflow step definitions named in the runtime evidence…",
capped at 8KB, every failure degrading to a log line. **Verified against the
live DB with run 1's actual ref**: the exact query returns page-build-handler's
`complete_error` JSON — `action: complete_workflow`, success_message included.
In run 2 the causal step is citable static evidence the moment the error log
names it, as it did in all five run-1 bundles.

**F0.4c — same-file sibling signatures (assemble).** For each `path:Symbol`
scope entry, the signatures of that file's OTHER functions are listed (6KB cap,
whole-file entries excluded) with an explicit "name these in next_scope" hint.
Unit test pins the exact run-1 gap: `isLegalPage` in scope ⇒ `loadPagesForNav`
listed as a sibling, the in-scope symbol itself excluded, unrelated files
excluded, cap leaves a truncation marker.

**Verification:** gofmt clean (two pre-existing dirty files in the package are
not mine and untouched); `go build` OK; full `pkg/diagnose` +
`actions` suites green; new tests: `TestTierGuard_*` (4),
`TestWorkflowRefsFromRuntime`, `TestSiblingSignatures`.

**NOT LIVE:** all four slices are chassis-binary changes. Run 2 needs an image
build + rollout (v1.0.1101), then fire 090 with the IDENTICAL symptom string,
no SEED_SCOPE, site data untouched. F0.4d (symptom-closure gate) is
deliberately NOT in this batch — one variable cluster per run.

### Turn 9 — v1.0.1101 deployed; run-2 attempt 1 LOST TO THE ROLLOUT; re-fired as r3

Owner deployed. Verified before firing: commit `16961a04 v1.0.1101 diagnosis
loop amend` carries all four F0.4 source changes (only the untracked
`step_tierguard_test.go` stayed local — test-only, no binary effect); all three
diagnose agent_definitions carry `image_tag=v1.0.1101`; no stale diagnose-agent
pod to be reused.

**Attempt 1 (`guides-nav-benchmark-r2`, corr `32c1e88f`) never started.** The
spawned diagnoser (confirmed on v1.0.1101) initialised and successfully
produced its init response at 18:35:30 — but the parent orchestration froze at
`spawn_diagnoser`/`AWAITING_RESPONSES` (`last_activity` 18:35:29, never moved).
`call_diagnoser` was never issued, so the job topic was never created — which
is why the idle diagnoser then logged "Unknown Topic Or Partition" against its
own request topic, plus broker dial timeouts. **Not a Kafka outage**: brokers
11 days up / 0 restarts, and the platform completed 90 orchestrations in the
same 40 minutes. **Cause: a race with the rollout.** The parent `agent-chassis`
pod restarted at ~18:32 (the v1.0.1101 rollout); the run was fired at 18:35
into its boot/consumer-rebalance window and the init response fell into the
gap. GOTCHA recorded: **after `make release redeploy-agents`, wait for the
chassis deployment to settle before firing a diagnosis** — a run dispatched
into the rebalance window dies silently at spawn, with no error row anywhere.

Cleanup: r2 intake marked `failed` with the cause in `error`; the stuck
orchestration rows left for the stale-orchestration-reaper; the idle pod reaps
at ~3600s. **Re-fired as `guides-nav-benchmark-r3`, corr
`dd1186b9-467d-4edd-9240-fa16a9a7d780`** — identical symptom string, no
SEED_SCOPE, deployments all fully ready this time. This remains "benchmark
run 2" for scoring; r3 is only the item slug.

(Noted with a smile: "orchestration stuck at spawn with no error anywhere" is
itself precisely diagnosis-loop material — runtime tier names the frozen step,
static tier names the missing error_step on `call_diagnoser` in
diagnose-orchestrator's workflow. A future intake candidate.)

### Turn 10 — ★ BENCHMARK RUN 2 SCORED: the F0.4 fixes moved the needle exactly where predicted

Run 2 (`guides-nav-benchmark-r3`, corr `dd1186b9`): 5 iterations, diagnosis in
~18.5 min (vs ~24 in run 1), verdict **CONFIRMED**, terminal note persisted to
doc_notes (id `5366fbd9`; my first check raced the emit→persist_note gap and
wrongly read 0 rows — the completion watch fires on emit, persist lands ~2 min
later; noted).

**F0.4 verified in production, each doing its designed job:**
- **b (enrichment)**: iteration 1's bundle already carried
  `page-build-handler / complete_error` JSON with the success_message — pulled
  in because the error log names it. **The verdict's static citation is sourced
  "(agent_definitions)" — it cited the new section.** Run 1's unreachable cause
  became run 2's quoted evidence.
- **a (anchor)**: absent on iterations 1–2 (hypothesis still = symptom),
  **present on iteration 5** alongside the drifted hypothesis. Anchor, not clamp.
- **c (siblings)**: section present from iteration 1.
- **e (tier guard)**: the confirm carries state+static+runtime and passes the
  guard legitimately — the static leg being the causal step itself.

**RUBRIC SCORE (run 1 → run 2):**
| # | must-claim | run 1 | run 2 |
|---|---|---|---|
| 1 | `sections=[]` ⟺ not built (exact 5/10 partition) | partial | **partial** — correctly names `sections=[]` as guides-index's cause; never establishes the site-wide partition; and asserts one FALSE side-claim (the four guide pages "suggesting those pages built" — all are `planned`) |
| 2 | `check_has_ready_sections` routes sectionless → `complete_error` | not reached | **partial** — mechanism described ("no-sections condition … completing via the complete_error step") but the conditional not named/cited |
| 3 | `complete_error` is a SUCCESS-labelled `complete_workflow` | not reached | **PASS, CITED** — [static] quote of the success_message from agent_definitions |
| 4 | nav selects on `pages.status`, not `build_status` | not reached | **FAIL — and actively dismissed**: the conclusion asserts "not a nav issue" |

Refutation credit: PASS again (no reconcile/routing-table blame). Net: from
0/4-with-a-drifted-confirm to **1 pass + 2 partial + 1 fail, with the confirm
now genuinely explaining the blank-page half of the symptom** (sections=[] →
complete_error succeeds → work item complete → build_status stays planned →
blank page). The heart of cause B is diagnosed and cited.

**The residue is a precision-guided case for F0.4d.** The symptom's other half
— *why was the nav link published* — is not merely unreached; the verdict
waves it off ("not a nav issue") while confirming. A symptom-closure gate that
required each clause of the intake symptom to be explained-or-marked-unexplained
would have forced either another iteration into nav territory or an honest
partial. Run 3's variable is now empirically motivated, not speculative.

**Also observed:** the loop still asserts beyond its evidence in the periphery
(the "those pages built" side-claim). The closure gate's design should require
side-claims to carry citations too, or be dropped from the conclusion.

### 2026-07-09 (turn 10) — decisions
- **Benchmark method validated end-to-end**: same symptom, one variable cluster,
  measurable delta (0/4 → 1+2 partials). Runs are ~19 min and ~£small; keep
  iterating.
- **Next slice = F0.4d** (symptom-closure gate), now with run-2 evidence for its
  design: decompose the intake symptom into clauses; CONFIRMED must map each
  clause to a cited mechanism or list it as UNEXPLAINED; any UNEXPLAINED clause
  ⇒ status downgrades (or one more iteration targeted at the residue).
  Coordination: verdict_wire (ours) + prompt_template (tools chat's surface,
  fetch-first, snapshot, FYI).
- F0.3 (per-iteration `iteration_note` rows) remains open; the anchor/enrichment
  work has not touched it.

### Turn 11 (2026-07-10) — F0.4d BUILT: the symptom-closure gate; prompt updated on the tools chat's surface

**WHAT THESE ARE (owner-requested plain-language paragraph, kept verbatim):**
The symptom-closure gate is a new rule in the diagnosis loop's engine that
stops it from declaring victory on a partial answer. Today the loop ends with
CONFIRMED whenever its *current* hypothesis is directly supported by cited
evidence — but benchmark run 2 showed a verdict can be genuinely well-cited and
still explain only half of what the user reported: it nailed why
/guides/index.html was blank, then waved away the other half of the report
("not a nav issue") — why a navigation link to that empty page was ever
published. The gate closes that hole: a CONFIRMED verdict must now carry a
`symptom_check` — a short checklist mapping each observation in the *original*
reported symptom to the confirmed mechanism ("explained, and here's how") or
honestly marking it `unexplained`. If any observation is unexplained, or the
checklist is missing, the engine refuses to accept the confirm and sends the
loop back to gather evidence for the residue; if iterations run out, it ends as
an honest "couldn't fully explain it" rather than a confident half-answer. The
tools chat is not software — it is a separate, parallel Claude conversation the
owner runs, whose workstream owns the tool-generation pipeline and the
travelling-docs infrastructure (the doc_plans/doc_notes tables, the
persist_note wiring that files each diagnosis's terminal note, and the
diagnose-agent workflow's tail). The two chats coordinate through documents in
the repo rather than a shared context, and the standing collision rule is that
neither edits the other's active surface without fetching its current state
first and leaving a note. That matters here because the verdict *prompt* lives
inside the diagnose-agent workflow JSON — the tools chat's territory — so this
change was done fetch-first, with a snapshot, and with a written FYI.

**BUILD RECORD:**
- **Domain + wire** (ours): `Verdict.SymptomCheck []SymptomCheck`
  (`{observation, explained, how}`) in `loop.go`; `VerdictWire` +
  `toVerdict()` carry it through; the doc-comment records honestly that the
  gate enforces coverage was DECLARED, not that the model's decomposition is
  complete — that judgement stays with the prompt and the human.
- **Engine** (ours): the gate lives in `coerceVerdict` — the SAME shared
  coercion as item-24 and the tier guard, so DecideStep's decision, Advance's
  trail, and Run()'s trail cannot drift. CONFIRMED with missing `symptom_check`
  → Unverifiable ("must map each observation…"); CONFIRMED with any
  `explained:false` → Unverifiable with the residue NAMED in NeededEvidence so
  the next iteration can chase it. REFUTED/UNVERIFIABLE are exempt — the gate
  stops premature victory, not investigation. `confirmConclusion` now renders a
  "Symptom coverage:" block so the human sees what the gate enforced.
- **Hard-require rationale**: a lazy model could bypass a presence-optional
  gate by omitting the field; requiring it mirrors cite-or-abstain's hard rule.
  All 6 standing confirm fixtures gained coverage via a shared `covered()` test
  helper; new `step_closuregate_test.go` pins: missing check degrades; partial
  coverage degrades AND names the residue; full coverage stands and renders;
  REFUTED/UNVERIFIABLE exempt; wire round-trip parses and degrades.
- **Prompt** (tools chat's surface, coordinated): fetch-first;
  `snapshot_agent` id `34f4afc8-de3c-45e6-a713-263ef19755c7`; added hard rule 8
  (decompose the ORIGINAL symptom; never dismiss an unexplained observation as
  out of scope; side-claims need grounding — the run-2 "those pages built"
  false side-claim is this line's origin) + the `symptom_check` schema entry.
  Write-back via dollar-quoted UPDATE ($F04DQ$ tag, collision-checked);
  verified after: length 9,497, rule 8 present, `{{.bundle.bundle}}` intact,
  verdict config keys untouched. **FYI filed for the tools chat**:
  `travelling_docs/FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md`.
- gofmt clean; build OK; full `pkg/diagnose` + `actions` suites green.

**DEPLOY STATE**: the prompt is LIVE now (workflow JSON is read per run); the
engine gate rides the next chassis image (needs a build past v1.0.1101). Until
that deploys, the model is asked for `symptom_check` but nothing enforces it —
old binaries ignore the unknown field harmlessly, so ordering is safe. **Run 3
(the F0.4d measurement) is blocked on the image.** Prediction to grade run 3
against: on the identical symptom, the loop either (a) chases the nav clause
into `populate_nav_tables_action.go` territory and confirms with full coverage,
or (b) ends UNVERIFIABLE-at-cap with the nav clause explicitly named
unexplained. Either is a pass for the gate; a CONFIRMED with the nav clause
absent from symptom_check is the failure mode to watch for (model games the
decomposition by omitting the awkward observation).

### Turn 12 (2026-07-10) — ★ RUN 3 SCORED: no more confident half-answers; the guards fired in production; and the benchmark found its third engine defect

v1.0.1102 deployed (owner). Pre-flight: diagnose definitions on 1102; gate
sources clean in HEAD; chassis pod 63 min old (past the rebalance window that
killed run-2 attempt 1); no stale diagnoser. Fired `guides-nav-benchmark-r4`,
corr `5120c0dc-2fee-4b33-bb9c-a1117868dec5`, identical symptom, no seed scope.
Spawned pod confirmed on v1.0.1102. **3 iterations, ~16 min, verdict
UNVERIFIABLE, stopped_by scope-not-narrowing.**

**The run, iteration by iteration (from the trail + persisted bundles):**
1. Model abstains, issues **4 data_requests** for the pages/work-item rows.
2. Bundle 2 (55.9KB) **carries the answers** — the guides-index
   `planned`/`sections=[]` row is in it. The model attempts **CONFIRMED**… and
   **the tier guard coerces it**: the trail's NeededEvidence carries F0.4e's
   exact text ("confirmed on one evidence family only; a confirm needs BOTH a
   static … AND a state/runtime citation"). **First production firing of any
   of the new guards.** The confirm was state-only; the engine demanded the
   mechanism too; the loop continued. (The closure gate sat layered BEHIND the
   tier guard and was never reached — first refusal wins. Its prompt half
   plausibly contributed to the model's caution; unmeasurable.)
3. Bundle 3 is **byte-identical in size to bundle 1 (37,529 B)** — the
   data_request answers are GONE. Blind again, the model re-issues 4 near-same
   requests; the scope-narrowing guard stops the loop.

**Final output — the honest version of run 2's overreach.** "NOT CONFIRMED …
still needed:" a four-item gap list that explicitly names the causal mechanism
as the leading suspect — "if [sections] empty or missing, the page-build-handler
complete_error step fires with 'Content writer skipped — page has no sections
defined', which would produce a blank page" — and ends "Hand to a human with
the full trail; do NOT auto-conclude."

**SCORE against the pre-registered run-3 criteria:** primary criterion **PASS**
— no confident half-answer exists to hand to a fixer; what would have shipped
as run 2's partial CONFIRMED now ends as an honest, precisely-scoped
abstention. The gaming failure mode (a confirm whose symptom_check omits the
nav clause) did not occur. The (b)-path nuance recorded honestly: the residue
named in still-needed is the missing DB evidence, not the nav clause verbatim —
because no confirm survived long enough for the closure gate to interrogate
its coverage.

**★ NEW ENGINE DEFECT — the benchmark's third: data_request answers are
ONE-SHOT.** They ride only the bundle immediately after the requesting verdict
and evaporate from every later one. So when a guard (rightly) refuses the
verdict that follows the enriched bundle, the fetched evidence is LOST; the
loop re-requests, looks like it is spinning, and trips scope-not-narrowing.
Run 3 ended UNVERIFIABLE not because the loop couldn't reach the answer but
because it couldn't KEEP it. Next slice (**F0.5**): persist data_request
results across iterations — accumulate them in LoopState (capped) or re-inline
prior answers in every subsequent gather. With that, run 3's shape converges:
iter-3 would have held the rows AND the tier guard's demand for a static
citation, pointing straight at a full-coverage confirm.

**The arc across three runs, for the record:** run 1 — wrong answer,
confidently (drifted confirm, 0/4 musts). Run 2 — half answer, confidently
(right mechanism for the blank page; nav clause dismissed). Run 3 — no
half-answer possible: over-reach refused by the guards, honest abstention with
the mechanism named as prime suspect and a hand-to-human instruction. Each run
cost ~20 min and found a real engine defect. The known-answer benchmark is
carrying this workstream.

### Turn 13 (2026-07-10) — F0.5 BUILT: data_request answers now persist across iterations

**The design found reuse instead of new machinery.** Three options were on the
table: (A) persist answers in `diagnosis_artifacts` and re-read them each
gather (durable, but needs a `kind` CHECK migration and a write path);
(B) accumulate answers in LoopState (rejected outright — answer TEXT riding
collected_data is exactly the cd-bloat class that caused the recorded 1.27MB
Kafka incident); (C) **re-run, don't store**. C won on a found fact: the
engine's spin guard ALREADY accumulates every issued request in
`LoopState.SeenRequests`, keyed by the raw trimmed SQL
(`loop.go:guardAfter` — `key := strings.TrimSpace(dr.SQL)`). The requests
already round-trip through state; only the FORWARDING was one-shot.

**The whole fix is one change in `diagnose_route_action.go`:** where the
continue branch forwarded only the current verdict's wire requests into
`route.data_requests`, it now forwards the UNION — current first (their `why`
preserved), then the prior `SeenRequests` keys, deduped, sorted for
determinism, capped at 12 (`maxForwardedDataRequests`), and each prior key
re-linted with `IsReadOnlySQL` before forwarding (state rides collected_data,
so its keys are treated as data; load_runtime re-lints again and its READ ONLY
transaction remains the real guarantee). `load_runtime` then re-runs the lot
every iteration under its existing per-request row/cost caps — answered
evidence persists for the price of a few bounded SELECTs. The spin guard is
untouched: it judges what the MODEL issues, never what the route forwards.

**Cost note:** ~12 SQL strings ≈ 3–4KB in collected_data (safe), and the
re-run results land in the BUNDLE (disk-persisted artifact), not in state.

**Tests** (`diagnose_route_datareq_test.go`, all pass): the run-3 hole (empty
current verdict still re-forwards prior requests, marked as re-runs);
current-request `why` preserved + dedupe against seen; cap honoured with
current first; **a write statement in state is never re-forwarded**; empty in
→ empty out. gofmt clean; both suites green.

**Predicted run-4 shape** (pre-registered): iteration 2's guard-refused
confirm no longer loses the rows — iteration 3's bundle carries the
guides-index row AND the tier guard's demand for a static citation AND (via
F0.4b) the `complete_error` step JSON. Expected outcome: a CONFIRMED with
static+state citations and a `symptom_check` covering both symptom clauses —
the first full-coverage confirm — or an honest abstention if the nav clause
still eludes it. **Run 4 is blocked on the next chassis image (post-1102).**

### Turn 14 (2026-07-10) — ★ RUN 4: the first FULL-COVERAGE CONFIRMED — and the gate's structural limit, on display in its own output

v1.0.1103 deployed. Pre-flight caught the chassis pod at 2 min old — the exact
run-2 killer — so the fire was DELAYED 300s by design (the gotcha earning its
keep). `guides-nav-benchmark-r5`, corr `5179a2ea-12d5-4215-9457-dda0f4e2c687`,
identical symptom, no seed scope. **CONFIRMED in 2 iterations, ~8 minutes —
the fastest run of the four — with the first `symptom_check` ever emitted:
five observations, all marked explained, rendered in the conclusion's new
"Symptom coverage:" block.** Trail: iteration 1 abstained and issued
data_requests; iteration 2 confirmed with state+static+runtime citations —
all three guards (citation, tier, closure) passed LEGITIMATELY. Terminal
doc_note persisted. Item closed.

**What it got RIGHT (scored against ground truth):**
- The blank-page causal chain is correct and cited end-to-end: `sections=[]`
  on guides-index → `complete_error` fires ("Content writer skipped — page has
  no sections defined", quoted from the F0.4b-inlined step definition) →
  work item completes → page never built → blank. Must-claim 3 cited for the
  second consecutive run (tier mislabeled runtime vs static; the quote is the
  step definition — minor).
- The nav-link clause is IN the coverage table with a real citation (the live
  `site_nav_items` row) — versus run 2's "not a nav issue" dismissal. The
  human now SEES what was and wasn't reached.
- Refutation credit passed a third time; the model even flagged its own
  contradiction mid-explanation ("However, the persisted sections=[]
  contradicts this default") instead of papering over it — the rule-8
  side-claim discipline visibly at work.

**What honest scoring must also say:**
1. **Must-claim 4 is still unreached, and the gate let a SHALLOW explanation
   through.** The nav clause's "explained" is "the nav row exists and is
   live" — which restates the observation rather than explaining WHY the
   system published a link to an unbuilt page (`loadPagesForNav` filtering on
   `status`, not `build_status`). The gate enforces that coverage is DECLARED
   (exactly what its doc-comment promised, no more); explanation DEPTH is
   beyond a structural check. The improvement over run 2 is real but specific:
   the shallow answer is now visible and auditable instead of silently absent.
2. **Grade inflation on the control clauses.** The gamesdesign/gaswholesalers
   entries are marked `explained: true` while their own text says "the bundle
   contains no data … unverifiable from this bundle." Under rule 8 those
   should arguably be `explained: false` — though marking them false would
   deadlock the loop (RUNTIME_SITE scopes evidence to dartsonline, so they are
   permanently unverifiable in-run). The truer fix is a third disposition —
   `context`/`not-applicable` — for comparative clauses that are not
   observations of the defect. Logged as an F0.6 candidate alongside:
   require each `explained:true` entry to reference at least one citation.
3. **A wrong side-trail survived into the static citations**: the
   `apply_gap_plan`/`defaultSectionsForPage` creation narrative (the pages
   were planner-created, not gap-plan-created). The model itself noticed the
   contradiction and hedged, and the FINAL causal statement rests on the right
   evidence — but two of the four static citations are decoration for the
   wrong path.
4. **F0.5 went unexercised.** The run won in 2 iterations, so there was no
   iteration 3 for re-forwarded requests to rescue. Its unit tests stand; its
   production proof waits for a run that needs it. (Why run 4 succeeded where
   run 3 was refused: iteration 2's verdict this time carried static citations
   alongside the state rows — tier-covered — where run 3's was state-only.
   Run-to-run retrieval variance, not a code change we can point to.)

**The four-run arc, complete:**
| run | outcome | character |
|---|---|---|
| 1 | CONFIRMED, wrong cause | drifted; answer unreachable (Go-only corpus) |
| 2 | CONFIRMED, half the symptom | right mechanism for the blank page; nav dismissed |
| 3 | UNVERIFIABLE, honest | over-reach refused by the tier guard; evidence lost (one-shot) |
| 4 | CONFIRMED, full declared coverage | right blank-page chain cited; nav covered shallowly but VISIBLY |

Each run found a real defect; four engine fixes shipped and verified against
the same symptom. The remaining blind spot is singular and precise: no run has
ever pulled `populate_nav_tables_action.go:loadPagesForNav` into scope —
retrieval satisfies the nav clause with the `site_nav_items` DATA row instead
of the nav-generation CODE. That, plus the F0.6 coverage refinements, is the
open frontier. F0 itself — intake, egress, observability, honest verdicts —
is functionally COMPLETE bar F0.3's per-iteration notes.

### Turn 15 (2026-07-10) — F0.6 + blind-spot fix + F1.1a, all built (owner: "do them all, your order")

Order chosen: F0.6 first (smallest; sharpens the CONFIRMED contract F1
consumes), blind spot second (investigate before fixing), F1.1a last. All Go
changes batch into ONE image so a single run 5 measures everything.

**F0.6 — coverage refinements (engine + prompt, both live where they can be):**
`SymptomCheck` gains `context bool` (comparative/background clauses exempt
from the accounting — the model is no longer forced to grade-inflate clauses
it structurally cannot verify) and `cites []int` (indices into the citations
array; an `explained:true` entry with no in-range index degrades — "an
explanation you cannot ground is not an explanation"). Gate order: missing
check → unexplained → uncited. Conclusions render three marks:
[explained]/[UNEXPLAINED]/[context]. Prompt rule 8 + schema updated
(fetch-first — surface unchanged since the morning write, one accumulated
trailing newline trimmed; snapshot 'pre-F0.6'; FYI addendum appended for the
tools chat). Three new gate tests incl. the exact run-4 grade-inflation case.
NOTE: run 4's verdict would NOT survive the new gate — that is the intent.

**Blind spot — investigated first, and the villain was OUR OWN CAP.**
`loadPagesForNav` IS in the corpus; `isLegalPage` WAS in scope in every run's
iteration 1; but `cap_hit=1` in every persisted bundle: `siblingSignatures`
rendered files in analysis order, so alphabetically-early giants
(apply_gap_plan, maintenance_actions) ate the whole 6KB before
populate_nav_tables_action.go got a line. Not retrieval; first-come-first-
served budgeting in F0.4c. Fix: **fair-share per file** (capChars/n, floor
600) with a per-file "+N more — put the bare file path in next_scope" marker.
Regression test reproduces the exact starvation shape (80-function giant +
nav file; the nav sibling must appear).

**F1.1a — the constrained edit plan (fix-proposer), first F1 slice:**
- Migration applied: `diagnosis_artifacts` kinds + `fix_plan`; iteration ≥ 0
  (0 = run-level artifact). Both constraints verified rejecting.
- New action `diagnose_persist_fix_plan` (registered): structural validation —
  summary/edits/rationale/sketch non-empty, operations allowlisted
  (modify|add|remove|config_change), repo-relative paths (no traversal, no
  absolute, no whitespace), ≤8 edits ("more is architecture change"),
  `grounded_in` quotes REQUIRED, 32KB cap — then persist kind='fix_plan'.
  UNLIKE the bundle write-through this FAILS the step on bad input: persisting
  a malformed plan would hand F1.1b garbage. 6 validator tests incl. traversal
  and rogue-operation rejection.
- New agent `fix-proposer` seeded (donor diagnose-orchestrator, 7 steps):
  load_diagnosis (by correlation_id) → **check_confirmed — refuses anything
  not CONFIRMED, the gate F1 waited for** → load final bundle → propose
  (LLM, hard rules: platform-not-site-data per the owner ruling; minimal;
  grounded; no new deps/DDL; config_change for workflow-JSON edits) →
  diagnose_persist_fix_plan → complete. **Writes NO code — no git token
  needed; F1.1b (branch + PR behind the spawn-gated token) is next.**

**DEPLOY STATE:** F0.6 prompt live now; fix-proposer agent live now (its
workflow runs on the deployed image). Go changes awaiting next image
(post-1103): F0.6 engine gate, fair-share siblings, `diagnose_persist_fix_plan`
(the proposer's persist step will fail with "unknown action" until then —
do not fire fix-proposer before the image lands). **Run 5** measures F0.6 +
fair-share on the benchmark; then fix-proposer fires on run 4/5's CONFIRMED
correlation for the first plan.

### Turn 16 (2026-07-10) — ★ RUN 5 SCORED (F0.6 + fair-share all verified); fix-proposer's first run failed usefully

**RUN 5** (`guides-nav-benchmark-r6`, corr `e08c5b01-01ef-42ad-80d0-b77c50ec9e84`,
v1.0.1104): **CONFIRMED in 2 iterations (~9.5 min) under the strictest gate
stack yet** — and run 4's verdict would not have survived it. Every F0.6
objective observed in the output:
- Both control clauses arrive **[context]** with honest "no data to verify"
  text — the run-4 grade-inflation defect is gone.
- **Fair-share worked end to end**: `loadPagesForNav` in iteration 1's sibling
  list (per-file "+N more" markers present), the file pulled into scope by
  iteration 2, and — first time in five runs — the verdict carries **static
  citations from nav-generation code**: `classifyPagesForNav`'s child-page
  skip and its `childPrefixes` list (which is why guide sub-pages are excluded
  from nav). New evidence angle too: `page_components` = 0 rows for the page.
- Coverage now six entries: four [explained] with cites, two [context].
  Must-claim 4's precise line (`loadPagesForNav`'s `status`-not-`build_status`
  filter) still not cited verbatim — right file, adjacent mechanism — scored
  PARTIAL, honestly. Run arc: wrong → half → honest abstention → covered →
  **covered under a stricter gate with nav code in evidence**.

**FIX-PROPOSER first firing — refused its own plan, for the right reason,
with two real defects found:**
1. The propose step's `max_tokens: 4000` truncated the plan mid-JSON-string
   (10 open braces, 9 closed; cut at "WHERE function LIKE"). Fixed live:
   snapshot `f9d90a2d`, max_tokens → 8000.
2. `diagnose_persist_fix_plan` assumed `proposal.result` is a parsed map;
   `execute_llm_prompt` stores a raw STRING when the model's JSON does not
   parse. Patched with the same map-or-string defence as decodeAnalysisOutput
   (`planBytes`: fences stripped, string passed to the validator) plus a
   truncation-specific error message; tests added; rides the next image.
   With max_tokens fixed, valid JSON arrives as a map, which the DEPLOYED
   binary already handles — so the retry needs no deploy.
The failure chain itself is a good sign: a truncated plan was REFUSED by the
validation gate rather than persisted — exactly the fail-closed behaviour
F1.1a promised. Retry in flight on the same correlation.

### Turn 17 (2026-07-10) — ★ THE FIRST FIX PLAN PERSISTED — pipeline proven; plan quality judged honestly

Attempt 3 (after the max_tokens-inside-ai_service fix): **fix plan persisted in
~80s** — kind='fix_plan' on correlation `e08c5b01`, 4 edits, 7 grounded_in
quotes, risks section present. The FULL CHAIN now exists end to end on one
correlation id: symptom in via 090 → 2 bundles → gated CONFIRMED with
symptom coverage → constrained edit plan, all in diagnosis_artifacts.

**Judged against ground truth (the fix we have known since turn 1):**
- **MISSES cause B**: no edit touches page-build-handler's missing
  `mark_no_sections` — the silent-success terminal the diagnosis itself cited.
- **MISSES cause C**: edit 4 says "no change is needed here for the nav" and
  proposes... a COMMENT. The `loadPagesForNav` status-vs-build_status fix
  remains unproposed.
- Edit 1 (defaultSectionsForPage: add a section-index case) is a defensible
  platform hardening but targets the GAP-PLAN creation path — the pages were
  planner-created (same apply_gap_plan mis-attribution as run 4's citations).
- Edits 2–3 are an "audit this" instruction and a literal "no code change
  required" — structurally valid, semantically no-ops.
- On the plus side: honest risks ("the reviewer MUST verify
  'section-listing' exists in content_components"), real grounding quotes,
  nothing destructive, platform-not-site respected.

**Verdict on F1.1a: the MACHINERY passed; the PLAN quality is limited by its
input** — the proposer sees only the diagnosis + final bundle, and run 5's
conclusion foregrounded sections=[]/page_components/classifyPagesForNav rather
than the complete_error chain runs 2/4 had front-and-centre. Garbage-bounded-in,
garbage-bounded-out: the fail-closed shape works; the diagnosis→plan
information channel needs widening (feed ALL coverage entries + the
complete_error citation explicitly, or all bundles not just the last).

**F1.1b refinements queued from this first plan:**
1. Validator: reject no-op edits (sketch containing "no code change
   required" / operation 'modify' with an audit-only sketch).
2. Proposer input: include the diagnosis's full citation list verbatim, not
   just the conclusion prose; consider all bundles.
3. The COUNCIL (F2) is visibly the missing organ: a guidelines/reuse reviewer
   would have flagged "comment-only edit" and "wrong creation path" instantly.

**Also this turn — the max_tokens dead-config find (platform-wide gotcha):**
`execute_llm_prompt` reads max_tokens ONLY from agent top-level config or
INSIDE step ai_service (ai_actions.go:252-256); step-config root is silently
ignored and the Anthropic client defaults to 2048. The diagnose-agent verdict
step ran capped at 2048 through ALL FIVE benchmark runs (survived on short
outputs; coverage-rich verdicts are near the ceiling). Both agents fixed live
with snapshots; runbook gotcha added; FYI addendum tells the tools chat to
grep their workflows for the pattern. Two truncated plans (~8KB ≈ 2048 tokens,
the tell) were correctly REFUSED by the F1.1a validation gate — fail-closed
worked on its first two real failures.

### Turn 18 (2026-07-10) — F1.1b(a)+(b) and F2.1 BUILT; F1.1b(c) designed

Owner: continue with F1* and F2 in my suggested order. Delivered:

**F1.1b(a) — no-op rejection in the validator** (Go, awaits image): explicit
phrases only ("no code change", "no change required/needed", "clarifying
note/comment", "comment-only") — over-blocking a real edit is worse than
letting the council catch a subtle no-op. The first plan's own two no-ops are
now test fixtures and are rejected.

**F1.1b(b) — proposer input widened** (DB, live): last TWO bundles
(string_agg, desc); prompt rules 6 (every cited mechanism gets a covering edit
or an explicit risks line saying why not — aimed squarely at the missed
complete_error) and 7 (every edit CHANGES something; observations belong in
risks). max_tokens correctly inside ai_service in the seed now too.

**F2.1 — the council, first slice** (workflow live; decision action awaits
image). Two reviewers as sequential LLM steps + a DETERMINISTIC Go decision —
deliberately not a third model opinion about two model opinions:
- review_editquality: real edits / minimality / right causal path / missing
  cited mechanisms (the run-e08c5b01 failures, made a standing reviewer).
- review_guardian: blast radius, architecture-change signals (Q-E v1 lives in
  its prompt), surface ownership — **holds the hard veto** (Q-D v1: flag in
  step config as hard_veto_from; column-vs-config placement stays open).
- diagnose_council_decide: ordered rules (hard veto → rejected; any veto →
  rejected; any objection → revise; all approve → approved), decided_by names
  the rule that fired, report persisted as kind='council_report' on the same
  correlation. Malformed reviewer output FAILS CLOSED (same stance as the plan
  validator; reuses planBytes' map-or-string defence + the truncation-aware
  error). 5 aggregation tests.
- Q-G v1 = role prompts + plan + diagnosis. plan_persisted.plan_json (string)
  added to the persist action's result so reviewer templates render the exact
  plan, not a Go map dump.

**F1.1b(c) — DESIGN recorded in the plan** (build next): separate
fix-implementer agent, gated on council approval; GITHUB_WRITE_TOKEN via a new
spawn gate mirroring isRepoCloningAgent (never on shared pods); sketches →
concrete diffs → constrained editor with the plan's file list as a hard
allowlist → branch → gofmt+build in a spawned golang-image Job (chassis image
has no toolchain) → PR carrying the Q-H package. Human terminal.

**State:** kind CHECK now includes 'council_report' (applied); fix-proposer v2
live (10 steps); suites green. **Everything Go rides the next image — do NOT
fire fix-proposer until it deploys** (council steps would fail as unknown
actions). After deploy: re-fire on `e08c5b01` → expect plan v2 (no no-ops,
complete_error covered or excused) + the first council report.

### Turn 19 (2026-07-10) — ★ THE COUNCIL'S FIRST LIVE RUN: plan v2 much better, council said REVISE with real objections

Re-fired fix-proposer v2 on `e08c5b01` (settle-hold honoured — pod was 3m37s).
Full chain ran in ~140s: plan → 2 reviewers → deterministic decision →
council_report persisted. **DECISION: revise** (by editquality's objection).

**Plan v2 vs v1 — the F1.1b input+prompt changes worked:**
- **Edit 2 now hits cause B directly**: `page-build-handler/complete_error`
  `action: complete_workflow` → `fail_workflow` — the success-labelled error
  terminal, the exact mechanism v1 ignored. Rule 6 (cover every cited
  mechanism) did its job. This IS the mark_no_sections fix in different dress.
- **No no-op edits** (rule 7 + the validator): 2 real edits, down from v1's
  4-with-2-noops.
- Still imperfect: edit 1 still targets the gap-plan creation path.

**The council caught exactly its designed quarry — it did NOT wave it through:**
- editquality (object): edit 1 targets the wrong causal path — the diagnosis
  established sections=[] arrives via applyNewPage's `ON CONFLICT DO UPDATE SET
  sections=EXCLUDED.sections`, not defaultSectionsForPage; flagged the Kafka
  "topic partition not found" line as a red herring; MISSING: the actual
  ON CONFLICT path. Every one of these is correct.
- guardian (object): edit 1 touches apply_gap_plan_action.go, a **shared
  platform file every site consumes** (blast radius — the guardian's whole
  reason to exist); edit 2 doesn't name the owning pipeline (surface
  ownership); and — the best catch — **"does fail_workflow trigger unbounded
  retry when sections is empty?"**, a genuine architecture-safety question a
  human reviewer would want answered before merge.
- Deterministic decision: two objections, no veto → **revise**, decided_by
  "objection from editquality". Auditable, no third LLM.

**Why this is the turn that matters:** for six runs the loop's failure mode was
overconfidence — wrong/half/shallow answers stated as fact. The council is the
first component that pushes BACK on the fixer with specific, correct, actionable
objections and a decision that is neither rubber-stamp (approved) nor
dead-end (rejected) but "revise — here's what's wrong". The organ the first
plan proved was missing now exists and demonstrably works on the first firing.

**Artifacts on `e08c5b01` now:** 2 bundle, 2 fix_plan (v1+v2), 1 council_report
— the full symptom→diagnosis→plan→review chain, one correlation id, all
fetchable. F2.1 DONE. Next: F1.1b(c) (branch+PR behind the write token) — but
note a revise decision means there's nothing to implement yet; a natural
follow-on is a REVISE LOOP (feed council objections back to a re-propose step,
cap 2) so the fixer converges before F1.1b(c) ever opens a PR.

### Turn 20 (2026-07-10) — F2.2 revise loop BUILT; docs brought current; milestone doc written

**Revise loop** (owner: build it). `diagnose_council_decide` now counts the
council_reports for the run (one per round — the durable counter, no workflow
loop-state threading) and returns `round` + `should_revise` (= decision=='revise'
AND round<max_rounds, default 2). A revise that runs out of rounds becomes
**'exhausted'** (terminal, named so a human sees the loop gave up rather than
silently approved). Workflow v3: council_decide → **check_revise** conditional →
`should_revise` ? **repropose** (feeds diagnosis + prior plan + BOTH reviewers'
objections back to the model) → persist_plan (re-validates) → review ×2 →
council_decide (loop). Cap makes it terminate; a fresh plan each round makes it
converge. Verified wiring: persist_plan→review_editquality→…→council_decide→
check_revise→{repropose→persist_plan | complete}. 5-case cap test; suites green.

**Docs**: this file, PLAN, RUNBOOK all refreshed to current state (RUNBOOK's
CURRENT POSITION was stale at 2026-07-09; now reflects the full F0–F2 arc).
New shareable milestone doc: `MILESTONE_diagnosis_fix_loop_2026-07-10.md`.

**Deploy**: workflow v3 live; the round-counting Go rides the next image — the
revise loop won't fire correctly until then (should_revise absent → check_revise
else → complete, i.e. it degrades to single-round, harmless). Re-fire on
`e08c5b01` after deploy to watch a plan converge across rounds.

### Turn 21 (2026-07-10) — Handoff: revise-loop demo fired; awaiting results

**Handoff document created:** `HANDOFF_turn21_2026-07-10.md` — comprehensive
state summary for the next chat: what's built, what's running, what's untracked,
gotchas, database queries for checkpoint, and F1.1b(c) design sketch. Start a
new chat by reading this first.

**Critical design fix applied during turn 20:** round count was scoped per
correlation (the diagnosis), but correlation accumulates council_reports across
proposer re-runs. Fixed to count per orchestration_id (per proposer run).
Without this fix, demo run would start at round 2 of 2 and exhaust without a
repropose.

**Demo run in flight:** `e08c5b01-01ef-42ad-80d0-b77c50ec9e84` (run 5's CONFIRMED),
max_rounds=3, settled 300s before firing (rebalance-window safety). Expected:
plan v2 → council → revise → repropose → plan v3 → council → {approved | revise
| exhausted}. This is the first time the full five-step chain (diagnosis →
bundle → plan → review → revise-iterate) runs end-to-end in production.

**Untracked files ready for commit:** diagnose_council_decide_action.go (new,
F2.1 + F2.2), diagnose_council_test.go (new, 5-case cap + decision tests),
MILESTONE doc (new, shareable narrative), NOTES/PLAN/RUNBOOK (updated turns 16–20),
0NN_fix_proposer.sql (v3 seed), FYI addendum. Snapshot discipline observed for
all agent updates.

**Next steps:** (1) check demo completion (SQL in handoff); (2) if approved,
proceed to F1.1b(c) (design ready, sketch in handoff); (3) if revise/exhausted,
diagnose via council objections; (4) if timeout, check orchestration_states
`__step_error`.

### Turn 22 (2026-07-10) — demo graded; deploy gap found; a VETO exposes the loop's real gap

**The turn-20 demo did NOT get a fair run — and the reason is a deploy gap, not
config.** Graded the flight: 2 council reports, 1 repropose, final result stored
as `{round:3, decision:exhausted, "revise cap reached (3 rounds)"}`. But the run
wrote only 2 reports of its own, and there were 3 on the correlation (one from a
PRIOR proposer run at 15:51). round=3 = the **correlation-wide** count, not the
per-orchestration count (=2). So the **deployed v1.0.1107 binary counts council
rounds per correlation** — it does NOT carry the orchestration_id-scoping fix.
That fix is in the source (`diagnose_council_decide_action.go:155-158`,
committed 0333302d this session) but never rode into the image. Consequence: a
proposer run on a correlation that already has review history starts mid-count
and burns rounds it never used. `max_rounds` was already 3 (verified: DB row,
JSON number, updated 16:17:12 before the 16:18:40 spawn) — the config was a red
herring. **Gotcha added:** deployed round-counting is per-correlation until the
next chassis image; for a fair run, clear prior `council_report` rows on the
fix_correlation_id first (orchestration_id scoping makes this unnecessary once
deployed).

**Clean re-fire (orch `8c770fd5`, via new `091_TRIGGER_fix_proposer_v1.sh`).**
Cleared the 3 stale reports on `e08c5b01` (bundles + fix_plans preserved), fired
a fresh run. Passed the CONFIRMED gate, drafted a plan, council decided in one
round: **guardian HARD VETO → `rejected`** (`decided_by: hard veto from
guardian`). Terminal by design (veto ≠ revise → no repropose). Notably the plan
was the closest yet to ground truth — edit 3 was `complete_error → fail_workflow`
(the known-answer cause B) — yet the guardian vetoed all three edits as *"an
architecture change dressed as a contained fix"* (each touches shared platform
surface: `defaultSectionsForPage`/`applyNewPage`/the page-build-handler terminal,
all fleet-wide), and named the safe alternative (scoped data fix + re-queue a
page_rerender, "zero blast radius, fully reversible").

**THE FINDING (matters more than the demo):** the veto is CORRECT, and it
exposes a structural gap in the loop, not a reviewer error. Fixing causes B/C
*properly* is a broad, cross-pipeline change — an **architecture-level** edit —
which is exactly what the fix-proposer's mandate ("the smallest set of edits…
MINIMAL… if you need more than a handful it's architecture change — say so and
keep to the safe core") is built to REFUSE. So on this bug the proposer is asked
to produce a constrained plan for a fix that cannot be constrained; the guardian
rightly vetoes it every time. A `rejected` today is a silent terminal — the run
completes, the veto (with its recommended remediation) is persisted, and nothing
consumes it. That dead-end is the loop's real gap.

**REJECTION-HANDLING DESIGN (proposed — F2.3 candidate).** `revise` and
`rejected` must feed back DIFFERENTLY. Plan:
1. Replace the single `check_revise` conditional with a decision router
   (`check_decision`) that branches on `council.decision`:
   `approved`→complete(→F1.1b(c) PR); `revise`(rounds left)→repropose (as now);
   `exhausted`→escalate; `rejected`→**reframe-once**→(still vetoed)→escalate.
2. A **reframe** step distinct from repropose. Repropose says "address these
   objections, same plan shape." Reframe says: "The council VETOED this as too
   broad / architecture-level / cross-pipeline. Produce EITHER (a) a strictly
   narrower remediation the guardian would accept — prefer the reviewer's own
   recommended alternative — OR (b) an explicit 'this needs architecture review'
   declaration plus the minimal safe interim step. Do NOT resubmit platform-wide
   edits." Cap at 1 so it can't thrash.
3. **Escalation becomes a first-class SUCCESS terminal**, not a silent complete:
   write an `escalation` artifact (new kind, or council_report flag) carrying the
   diagnosis conclusion, the plan(s), the blocking veto/objections, and the
   reviewer-recommended alternative — the human hand-off package (feeds
   F1.1b(c): the human-review PR can carry it). "Needs architecture review" is a
   legitimate correct output for an architecture-level bug, not a failure.
4. Do NOT auto-apply the guardian's scoped data-fix suggestion: it fixes ONE
   site and leaves causes B/C live everywhere, contradicting the 2026-07-09
   platform-fix decision. The human owns the platform-review-vs-interim call;
   the tool's job is to surface both, which (b) does.
Deferred implementation to after the deploy gap is closed (round-scoping must be
live first) and after the F1.1b(c) write step, since escalation shares its
human-hand-off surface.

**Re-fire for a multi-round revise cycle (B) launched** after clearing the veto
report — plan variability (plans differ run-to-run despite temp=0) means another
run may draw objections rather than a veto.

**B result (orch `aadd532a`) — the full 3-round cycle, and it CONVERGED on one
axis.** Clean start (0 prior reports), 3 plans, 3 council reports, all this
run's own; deployed binary counted correctly (round 1→2→3, exhausted at cap,
`decided_by: objection from guardian — revise cap reached (3 rounds)`). The arc
is the demo we wanted:
- R1: editquality object (3: high/med/med, 1 missing) + guardian object (3:
  high/high/med, 4 missing) → revise.
- R2: editquality object (2: low/high, 0 missing) + guardian object (3:
  high/med/med, 3 missing) → revise.
- R3: **editquality APPROVE (0 objections)**; guardian object (4, but its notes
  say explicitly: "None of these objections cross into architecture-change
  territory… Veto is not warranted. All four objections are containable by
  pre-deploy audit queries") → revise → exhausted.
All three plans kept the same 3-edit skeleton (defaultSectionsForPage /
applyNewPage / page-build-handler config_change) and refined within it. So the
revise loop demonstrably converges: one reviewer fully satisfied, the other down
from architecture anxieties to OPERATIONAL DEPLOY GATES — pre-deploy audit
queries (cross-platform section-index survey; build_status domain check;
confirm the real component function name for the 'section-list' placeholder)
and sequencing on the known Kafka partition fault.

**Second design insight (pairs with the rejection design above):** round-3
exhausted here is really "conditionally approved pending verifications the loop
never ran." The guardian kept asking for READ-ONLY queries the platform can
answer (schema type of pages.sections; content_components.function values;
cross-site section-index counts). The diagnosis loop already has a data_request
mechanism — the PROPOSER lacks one. F2.3 should therefore also consider a
**verify step**: when a council round's blocking objections are all
"containable by pre-deploy checks", run those read-only queries and feed the
results into the next repropose instead of burning a blind revise round. That
plus the decision router turns both observed dead-ends (rejected, exhausted)
into productive paths: veto→reframe-or-escalate; exhausted-on-verifiable-asks→
verify-and-continue (or attach the checklist to the escalation package).

### Turn 22 (cont.) — F2.3 BUILT (code + v4 seed; rides the next image)

**Go (all in `platform/orchestration/actions/`, suites green, gofmt clean):**
- `diagnose_council_decide_action.go`: cap mapping extracted to
  `applyCouncilCaps` (pure, tested directly — the old test mirrored the logic;
  now it exercises the real function, 8 cases). Adds `should_reframe` = first
  rejection with rounds left; rejected-count from council_report metadata
  (`metadata->>'decision'='rejected'`, orchestration-scoped); count failure
  fails CLOSED (reframe spent → escalate) — a DB error must not grant extra
  LLM rounds. `max_rounds` bounds TOTAL cycles: a reframe consumes a round.
- `diagnose_run_checks_action.go` (NEW): the verify step. Pure reuse of the
  diagnosis containment — `dataRequestsFromCollected` (same {sql,why} wire as
  verdict data_requests) + `runDataRequests` (lint → READ ONLY tx →
  statement_timeout → EXPLAIN gate → capped rows). `max_checks` cap (default 8)
  with dropped checks NAMED in output. No checks → harmless note.
- `diagnose_escalate_action.go` (NEW): persists kind='escalation' — reason +
  decided_by + round + diagnosis conclusion + final plan + both reviews.
  Council decision missing = hard error (an escalation needs its reason);
  everything else best-effort with absences named. Persist failure FAILS the
  step (loud, unlike the bundle write-through: here the artifact IS the
  outcome).
- Registry: both actions registered, `IsLocal: true`, category diagnose.

**SQL (`0NN_fix_proposer.sql` → v4, dry-run PASSED on live DB in a rolled-back
tx; step graph verified — 19 steps, defined == referenced):**
- kind CHECK gains 'escalation'.
- Router chain: council_decide → check_approved → check_rejected →
  check_reframe → check_revise; `complete` is now reachable ONLY via approved.
- run_checks between check_revise and repropose; repropose prompt gains the
  verification-results section ("settle objections with these; do not argue
  with the data").
- reframe step (veto → narrower plan or explicit needs-architecture-review +
  minimal interim; site-scoped interim allowed IF risks names the deferred
  structural fix — constitution: note a knowingly deferred structural fix).
- Both reviewer prompts: optional `checks:[{sql,why}]` field ("ask rather than
  assume"); guardian additionally told a veto should name the safest contained
  alternative in notes (it seeds the reframe).
- escalate → complete_escalated success terminal; max_rounds 3.

**DEPLOY ORDER (hard):** chassis image (> v1.0.1107, carries round-scoping fix
+ the two new actions) → apply v4 seed → fire. v4 on the old binary fails at
the first unknown action; old workflow on the new binary is harmless. The §1
constraint block alone is safe any time.

## DECISIONS (with rationale)

### 2026-07-09 (turn 6) — benchmark verdict and what it buys
- **The plumbing is proven in production** (fetchable bundles, intake, terminal
  note). F0 is functionally complete bar F0.3's per-iteration notes.
- **The loop is not yet fit for unattended diagnosis of this bug class**, and we
  now know precisely why, with evidence rather than intuition. Three fixes,
  ordered by value:
  1. **A symptom-explanation gate before CONFIRMED.** The verdict must state how
     the cited cause produces the reported symptom, or downgrade to a partial
     finding. Highest value: it converts a confident wrong answer into an honest
     one.
  2. **Widen the static tier past Go.** Workflow definitions in
     `agent_definitions.default_config` are load-bearing platform logic and are
     invisible to the corpus. Either index them, or teach the loop that a
     `handler/step` string in `agent_error_log` is a pointer into that table.
  3. **File-granular scope expansion.** When retrieval implicates a file, offer
     the verdict its sibling symbols — `isLegalPage` beat `loadPagesForNav` on a
     vector match and the answer was one function away.
- The known-answer benchmark **earned its keep on the first run**: a discovery
  pilot would have ended with a plausible CONFIRMED verdict, three real citations,
  and nobody any the wiser that it was wrong.

### 2026-07-09 (turn 5) — diagnose items use two private statuses; dispatcher claims
- `awaiting_diagnosis` (queued) and `diagnosing` (in-flight). Never `triaged`,
  `approved`, `detected` or `claimed`. Rationale: every sweep in the platform
  filters on explicit status values, so a private value is inert *by
  construction*. Anchor-site choice then stops carrying any safety weight.
- The dispatcher claims (platform pattern, verified). The loop owns its own reap
  because it opted out of `claimed-item-timeout`'s coverage.
- Automatic dispatch ships **disabled**. Enable only after the image is live and
  the benchmark's blinding is confirmed — otherwise the loop would claim the
  pilot item and run it before we have checked it cannot read the answer.

### 2026-07-09 (turn 4) — Q-B CORRECTED: system.internal anchor, not null-site
- Null-site is impossible (NOT NULL column; site-anchored loader). Anchor every
  `needs_diagnosis` item to the existing `system.internal` pseudo-site and carry
  the site under diagnosis in `spec`. Status starts at `detected` so no existing
  dispatch loop can claim it.
- **Why this beats dropping the constraint:** dropping NOT NULL would weaken an
  invariant the entire relay relies on, to serve one namespace, and would still
  need a new loader. The pseudo-site already exists and already carries
  platform-wide work. The narrower change is the right one.

### 2026-07-09 (turn 2) — F1 targets the platform (OWNER CONFIRMED)
- The missing `mark_no_sections` step and the nav column fix land in the
  platform. Fixing dartsonline's plan would fix one site; causes B and C are
  relay-level and fix every site by construction. Same reasoning that promoted
  the roadmap gap to builder item 6.

### 2026-07-09 — pilot reframed as a known-answer benchmark (proposed, owner to confirm)
- The dartsonline guides symptom is fully diagnosed by hand, with static /
  live-data / runtime citations. Running the loop on it can no longer *discover*
  anything; it can **verify the loop**. That is more valuable at F0 than a
  discovery run, because F0's five criteria are all about the plumbing
  (intake, fetchable bundles, per-iteration notes, a cited mechanism) and only
  the fourth is about the answer — which we can now mark objectively.
- Scoring rubric pre-registered in PLAN_fixloop_pilot.md §3 so the grading
  cannot drift to fit whatever the loop happens to emit.

### 2026-07-09 — the bug itself is a platform defect, not a dartsonline defect
- Causes B and C are relay-level (a workflow definition; a nav action). Fixing
  dartsonline's plan fixes one site. Same shape as the roadmap gap that became
  builder item 6. The fix belongs in the platform; the F1 edit plan should
  target `check_has_ready_sections`/`complete_error` and
  `populate_nav_tables_action.go`, not the site's data.

### Carried forward unchanged from (9)
Q-A diagnosis_artifacts table (kind ∈ bundle|iteration_note), write-through
inside assemble. Q-B intake = `needs_diagnosis` item, `pipeline='diagnose'`,
null-site allowed. Q-C separate fixer agent, isolated write token, constrained
edit plan, gofmt+build gate. Q-D flag-based hard_veto; guideline-gap =
side-task. Q-F own working-notes storage, terminal note only into the tools
chat's doc_notes. Q-E/Q-G/Q-H remain open (F2).

## OPEN QUESTIONS RAISED THIS TURN
- Why does `/guides/index.html` render **blank** rather than 404, given no
  build ever ran? Something is serving a shell. Unresolved; low priority for
  the diagnosis, relevant to the fix's verification step.
- Should `reconcile_site_plan` grow the `unavailableBuilders` guard, or should
  `WriteBuildItemsAction` lose it? The two paths must agree; which way is a
  design decision for the builder thread.
- Was `mark_no_sections` ever written and removed, or only ever intended?
  Git archaeology would settle it and would feed F3's learning record.
