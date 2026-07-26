# 068 — every single-page REBUILD dies at resolve_links: optional mapping feeds a REQUIRED contract field, and only one of the writer's two callers supplies it

**Filed:** 2026-07-24 · **By:** about_page_commercial workstream (pilot rebuild was the third victim)
**Status:** **CLOSED 2026-07-26** — fix candidate B (config) is live and now **behaviourally
verified on the failing branch**: a rebuild-shaped writer child, dispatched with no `section_plan`,
passed `resolve_links` for the first time. See "Verification" below.
**Two of this file's own claims were wrong and are corrected in place below** — the mechanism
(`bugs_open/086`) and the reason candidate A was rejected (`bugs_open/087`). Neither changes the
fix; both change where the next thread should look.
**Diagnosis loop:** 090 corr `38cffebf-d01a-4922-9f39-e2deb5930e0d` — verdict **UNVERIFIABLE (iteration-cap)**: it CONFIRMED the downstream mechanism from the runtime log but could not fetch the two callers' dispatch inputs (its data queries errored); the missing upstream evidence was then closed BY HAND at runtime level (below). The loop's "still needed" statement named exactly the right evidence.

## Symptom

`page-rebuild` (seed 039) dispatches its per-page writer; the child fails:

```
step resolve_links failed: ... contract violation for agent 'internal-link-resolver':
missing required fields: [sections]
Provided fields: [page_name page_type site_id]
```

Occurrences: 2026-07-16 ×2 (agent_error_log, same `build_pages_loop_iter_0_write_page_content` shape — someone hit this eight days ago and did not file it), 2026-07-24 (orchestration corr `7a820803-8ce3-455a-8732-258638e6d976`, the about-commercial pilot).

## Root cause (runtime-evidenced)

Three facts compose:

1. **The writer's mapping is optional.** `page-content-writer` → `resolve_links` step:
   `"sections?": "input_data.section_plan.sections_ready"` — the `?` means "omit when absent".
2. **The target contract requires it.** `agent_definitions.input_contract` for
   `internal-link-resolver`: `required: [site_id, sections]`. An omitted optional mapping
   meeting a required contract field fails at EXTRACTION, before the call — which also
   **bypasses the step's `error_step: select_sections`** (the author's clear intent was
   that a failed link-resolve is non-fatal; extraction failures never reach that routing).

   > **CORRECTED 2026-07-26 — the second sentence is wrong, and it is the load-bearing one.**
   > Extraction failures *do* reach the routing: `executeStep`'s error goes straight to
   > `routeToErrorStepOrFail` (`platform/orchestration/coordinator.go:869`). What is missing is
   > the field it reads. `convertToWorkflowPlan` (`platform/messaging/processor.go`) builds
   > `models.Step` field by field and never named `error_step`, so **no persisted plan has ever
   > carried a step-level `error_step`** — 0 of 14,209 plan steps over three days, against 1,828
   > carrying the `config.error_step` twin (which survives only because the whole config map is
   > copied). Fleet-wide, 55 declared handlers across 19 agents are inert, and all 32
   > `resolve_links` failures ever logged are `severity='fatal'` — 30 of them resolver timeouts,
   > the exact case this step declares survivable. Filed as **`bugs_open/086`**, Go fix committed
   > `dca5649b3`, inert until an image roll.
   >
   > **What caught it:** reading `orchestration_states.workflow_plan` instead of
   > `agent_definitions`. The original claim was inferred from the definition and the outcome,
   > never checked against the plan the coordinator actually executes.
3. **Only one caller supplies section_plan.** Runtime comparison of the writer children's
   `initial_request_data->input_data` keys (orchestration_states):
   - FAILED child `d6e737fc` (caller **page-rebuild**): `current_page, db_sync, hero_url,
     logo_url, reviewed_brief, site_plan, site_record, style_collection` — **no section_plan**
   - OK child 15:35Z same day (caller **page-build-handler**): `current_page, domain,
     existing_content, `**`section_plan`**`, site_id, site_plan, site_record`

   Writer children (signature `compile_page`+`resolve_links`) over 8 days: 32 COMPLETED
   (build-handler path), 3 FAILED (the rebuild attempts). The rebuild flow has no section
   plan at dispatch time — this generation of the writer selects its own sections
   (`select_sections`/`process_sections_loop` are child steps), so the rebuild caller
   CANNOT sensibly supply one.

   > **CORRECTED 2026-07-26 — the last clause is wrong.** `select_sections` selects nothing:
   > it is an `extract_fields` reading `sections_ready` from either the resolver's output or
   > `input_data.section_plan.sections_ready`. On the rebuild path both are empty, so the writer
   > has no section plan and no way to build one — it dies at `process_sections_loop` instead
   > (proven live today, `bugs_open/087`). The rebuild caller not only *can* supply a plan, it is
   > the only party that can — via `plan_sections`, the action `page-build-handler` already runs.
   > So candidate A below was rejected on an unchecked reading of a step name.

**The Go action is already tolerant:** `ResolveInternalLinksAction`
(`platform/orchestration/actions/resolve_internal_links_action.go:132-135,151`) resolves
`sections` via a nil-safe path — missing → empty loop → returns empty `sections_ready`,
no error. The contract is the ONLY fatal element.

## Fix candidates

- **A. Supply section_plan from page-rebuild's dispatch** — REJECTED: the rebuild flow has
  no section plan at dispatch time (child plans its own sections); would mean inventing one.
  > **CORRECTED 2026-07-26:** rejected on a false premise (see above). It is not "inventing" a
  > plan — `plan_sections` derives one from the page's own section list, which is exactly what the
  > working caller does. A is now candidate A of `bugs_open/087`. B was still the right *first*
  > move (it is config-only, live immediately, and unblocks the step in front of you); A/087 is
  > the rest of the same path.
- **B. Move `sections` from required to optional in internal-link-resolver's
  input_contract** — APPLIED 2026-07-24 (seed `docs/agent_docs/sql_for_agents/203_link_resolver_sections_optional.sql`,
  targeted jsonb UPDATE; guarded, idempotent). Build-handler path unchanged (always
  supplies sections). Rebuild path: resolver no-ops harmlessly (nil-safe action), page
  proceeds — links on rebuilt pages resolve via the later section-level machinery, or not
  at all, which is the degradation the `error_step` design already accepted.
  Revert = the original contract verbatim, carried as a REVERT statement in seed 203.

**Census of the class** (query in the §9 entry): 3 instances fleet-wide — this one, plus
`index-orchestrator → code-indexer` (`repo?`, `owner?`), latent-only because
index-orchestrator is one of the never-observed dormant agents (`bugs_open/044`
inventory's territory; not filed separately).
- **C. Route extraction-time contract violations to error_step** (coordinator Go change) —
  the structural fix for the whole CLASS (an extraction failure silently escalates to
  step-fatal even where the author declared a non-fatal intent). Bigger blast radius,
  needs an image roll + council; NOT taken here. Candidate for a platform thread.
  > **CORRECTED 2026-07-26 — C describes a change that is already in the code.** The coordinator
  > routes extraction failures like any other step failure; the class defect is one layer earlier
  > (the plan converter drops the field). Taken as `bugs_open/086`, fixed in `dca5649b3`, council
  > corr `88ef6d08-ca87-4c90-b682-f85f1e6036f1`, inert until the roll.

## Verification

- [x] Contract row updated & re-read (`required: [site_id]`, `optional: [+sections]`).
- [x] **Still live 2026-07-26**, two days and several config sweeps later:
  `required: [site_id]`, `optional: [page_type, page_name, sections]`.
- [x] **Behavioural, 2026-07-26 — the failing branch, not a green happy path.**
  Re-armed finetuning.uk `/about` (`build_status → needs_rebuild`) and fired
  `about_page_commercial/p1_trigger_rebuild.sh` with `SEND=1`, correlation
  `7becd532-6c2b-43ec-b4d4-08a4727ddeb7`. The writer child
  (`bbcc1186-381c-42ea-8a09-0a35da69bac6`) was the rebuild shape —
  `initial_request_data->'input_data' ? 'section_plan'` = **false** — and it **passed
  `resolve_links`**, reaching `process_sections_loop`. No new
  `contract violation … 'internal-link-resolver'` row in `agent_error_log`; the last one in
  the fleet remains 2026-07-24 15:21, before the fix.
- [x] Build-handler path unaffected: in the same ten minutes, two `page-content-writer`
  children **with** `section_plan` ran, one `COMPLETED`, and no contract violation.
- [x] Live page unharmed: `https://finetuning.uk/about.html` still carries
  `about-commercial-block` (6 matches) and `Built by` (1), with `available to acquire` at 0.
  `build_status` restored to `deployed` as found.

**The rebuild still does not finish** — the same child then died at `process_sections_loop`
because nothing supplies it a section plan. That is a *different* defect, filed as
`bugs_open/087` with fix candidates. 068 is closed on its own defect: the contract violation is
gone, proven on the branch that used to hit it.

**No `Council-Reviewed:` trailer**: fix B was a config seed applied 2026-07-24 without a council
run, and a verdict cannot be back-dated onto a commit that predates it.

## Transferable pattern

016b §9: "An optional input_mapping (`field?`) feeding a REQUIRED contract field is a
latent per-caller fatality" (added 2026-07-24). Census query for other instances is in the
§9 entry. Related-but-distinct: `bugs_open/029` (phantom tool links — resolver OUTPUT
quality), `bugs_closed/054` query-list contract (empty-but-present list defeating
required; this case is the absent-key sibling).

**Added 2026-07-26, from the two corrections above** — both are the same habit, and it is the
expensive one in this repo: *a claim about runtime behaviour, inferred from a definition or a
step's name, written in the same voice as a measurement.*

- **A field in an agent definition is only live if the plan converter copies it.** Check
  `orchestration_states.workflow_plan`, not `agent_definitions` (`bugs_open/086`).
- **A step name is not a specification.** `select_sections` selects nothing; it is an
  `extract_fields` over two sources that are both empty on the path in question
  (`bugs_open/087`). Open the step config before reasoning from the name.
