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
