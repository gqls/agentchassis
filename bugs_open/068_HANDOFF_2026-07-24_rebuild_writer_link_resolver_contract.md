# 068 — every single-page REBUILD dies at resolve_links: optional mapping feeds a REQUIRED contract field, and only one of the writer's two callers supplies it

**Filed:** 2026-07-24 · **By:** about_page_commercial workstream (pilot rebuild was the third victim)
**Status:** OPEN → fix candidate B applied 2026-07-24 (see below), pending behavioural verification via the armed finetuning.uk pilot
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

**The Go action is already tolerant:** `ResolveInternalLinksAction`
(`platform/orchestration/actions/resolve_internal_links_action.go:132-135,151`) resolves
`sections` via a nil-safe path — missing → empty loop → returns empty `sections_ready`,
no error. The contract is the ONLY fatal element.

## Fix candidates

- **A. Supply section_plan from page-rebuild's dispatch** — REJECTED: the rebuild flow has
  no section plan at dispatch time (child plans its own sections); would mean inventing one.
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

## Verification

- [x] Contract row updated & re-read (`required: [site_id]`, `optional: [+sections]`).
- [ ] **Behavioural**: re-fire the armed pilot (`about_page_commercial/p1_trigger_rebuild.sh`,
  finetuning.uk about page is still `needs_rebuild`) → child must pass `resolve_links` and
  the page must deploy. Verify the LIVE page per the workstream RUNBOOK (component marker +
  template-created phrase + gated-off line absent).
- [ ] Confirm the build-handler path is unaffected: next routine writer child COMPLETED
  with `section_plan` supplied.

## Transferable pattern

016b §9: "An optional input_mapping (`field?`) feeding a REQUIRED contract field is a
latent per-caller fatality" (added 2026-07-24). Census query for other instances is in the
§9 entry. Related-but-distinct: `bugs_open/029` (phantom tool links — resolver OUTPUT
quality), `bugs_closed/054` query-list contract (empty-but-present list defeating
required; this case is the absent-key sibling).
