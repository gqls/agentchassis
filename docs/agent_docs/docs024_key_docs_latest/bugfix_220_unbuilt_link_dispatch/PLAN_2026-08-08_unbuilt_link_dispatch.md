# PLAN 2026-08-08 — bugfix 220: unbuilt_internal_link dispatch rebuilds the container and reads green

Bug: `bugs_open/220_HANDOFF_2026-08-08_unbuilt_link_dispatch_rebuilds_the_container_and_reads_green.md`.
Lane opened 2026-08-08 evening. The 206 lane filed the bug and explicitly left it
("Nothing pending here; the follow-up is the separate bug 220 dispatcher fix" — their
session's closing words). The 116 lane contributed the blast-radius census. Transcript
sweep run before starting: no session holds the fix-site symbols
(`load_page_record`, `availableBuilders`, the call_handler mapping) — the sessions that
score hits are the 206 lane (closed out), the fragments lane (reads the bug, works
`dead_fragment_link`), and the 201/RFC_017 lane (verification gate, committed, adjacent).

## Validity re-check (2026-08-08 ~18:50Z, all live)

- `build-dispatch-loop` `process_item.sub_workflow.steps.call_handler.input_mapping`
  still maps `"page_name?": "current_item.spec.page_name"` and has **no key reading
  `current_item.page_id`** — read from the live `agent_definitions` row.
- No verifier registered for `unbuilt_internal_link` (`verifier_coverage_test.go:160`
  still lists it in `itemTypesWithoutVerifiers`, category mechanical).
- The check still hardcodes `HandlerAgent: "page-build-handler"` for the arm
  (`check_phantom_internal_links.go:145`).
- The original demo instance is HEALED (vetcomparison `directory-index` deployed
  2026-08-08 17:02 by the 206 lane, through the framework) — the mechanism is intact;
  the next unbuilt link re-mints the loop.

## The defect, one level deeper than the bug file states it

The bug file's candidate 1 says "map `page_id?` and have `load_page_record` prefer it
over `page_name`". Reading `load_page_record_action.go:7-10` and `:174-187`: the action
resolves **page_name FIRST**, and falls to `page_id` only when the name is empty or a
non-page marker ("new page needed" etc.). For `unbuilt_internal_link` items
`spec.page_name` is always present and always a real page (the CONTAINER), so
forwarding the right id alone changes nothing — the name would still win. The fix must
give an explicitly-forwarded work-item id precedence, and per the OWNER RULING
2026-08-02 (RFC_010 §2: "new authority on a shared seam ships as an OPT-IN FIELD, not a
documented contract"), that precedence is an opt-in config field on the step, not a
global priority flip. A global flip would silently change `tool-recreation-handler`
(the only other live agent with a `load_page_record` step) and every future config that
supplies both keys with an LLM-adjacent `spec.page_id`.

## Decisions

1. **Candidate 1 (routing), realised as three small pieces:**
   - `build-dispatch-loop` mapping += `"page_id?": "current_item.page_id"`.
     PRECEDENT, not novelty: `site-work-orchestrator`'s own `call_handler` already maps
     `"page_id?": "current_fix_item.page_id"` — build-dispatch-loop is the outlier.
     `LoadWorkItemsAction` exposes the column already (`load_work_item_actions.go:776`),
     NULL column → key absent (never `""` — the 154 landmine rule).
   - `load_page_record` gains optional input `authoritative_page_id`. When it resolves
     and parses as a UUID, the lookup is by id and the name is ignored (with an info log
     naming the disagreement when the name would have resolved elsewhere). Malformed
     value → error (config defect, loud, matching `site_id`'s handling). Absent →
     behaviour byte-for-byte today's.
   - `page-build-handler`'s `load_page_record` step config +=
     `"authoritative_page_id": "input_data.page_id"`.
2. **Candidate 3 (verification): `VerifyUnbuiltInternalLinkResolved`**, registered for
   the item type, in `check_phantom_internal_links.go` next to the arm that mints it.
   Two disjuncts, per the `VerifyDeadFragmentLinkResolved` precedent and for the same
   reason (the item's own spec.fix names TWO remedies): resolved when the container no
   longer renders the href, OR when the target page (`target.PageID` = the item's
   `page_id` column = the TARGET) has `deployed_at IS NOT NULL`. Fail-closed (RFC_017
   default; no `FailOpenOnError`). Coverage map entry REMOVED, not edited
   (the literal_markdown precedent in the same file).
3. **Candidate 2 (check writes target's name into spec.page_name): NOT taken.**
   It closes one item type and leaves the mapping trap armed — the bug file says so
   itself; with candidate 1 landed it would be dead weight and would also LIE to any
   spec reader about which page contains the link.
4. **Candidate 4 (route by target page_type through availableBuilders): DEFERRED,
   recorded in the bug file.** Reasons: (a) `availableBuilders` lives inside
   `LoadWorkItemsAction` in package `actions`, which IMPORTS `discovery_checks` — the
   check cannot import it back; sharing it means a leaf-package refactor (the RFC_012
   landmine: a refactor is not a review of what it moves) touching the 206 lane's
   day-old machinery. (b) With the verifier live, a directory-target item now fails
   LOUDLY (attempts exhaust → `failed`, two-strike parks re-detections) instead of
   lying green — the silent-loop harm is gone; what remains is an efficiency gap on a
   rare page_type. (c) The 206 lane's `needs_directory` path is the designed remediation
   for those targets and already exists.

## Ordering / rollout

- Config live + old binary: mapping key forwarded but read by nobody (measured: zero
  live handlers dispatched by this loop read `input_data.page_id`); the
  `authoritative_page_id` config key is not in the old binary's InputSpec → ignored at
  runtime (`ExtractActionInputs` iterates DECLARED fields; `UnknownConfigKeys` is an
  offline audit, not a runtime rejection). So the migration is safe to apply
  immediately and the whole fix arms on the next fleet roll. NO `_HOLD` needed.
- New binary + config not yet applied: no `authoritative_page_id` in step config →
  unresolved → today's behaviour; verifier active (correct — it verifies outcomes, not
  routing).
- Migration `340_unbuilt_link_dispatch_authoritative_page_id.sql` + `_ROLLBACK` sidecar.

## Blast radius (measured, not argued)

- 116 lane census (bug file CONTRIB section): over all live items, 24 types,
  `unbuilt_internal_link` is the ONLY type whose `page_id` column disagrees with
  `spec.page_name`'s page; positive control 13/13 all-history rows disagree.
- Zero live `agent_definitions` rows map `current_item.page_id` in this loop today;
  `page-retraction` / `deduplicate-sections` read `input_data.page_id` but have zero
  work items ever and are not dispatched by this loop.
- Only `page-build-handler` and `tool-recreation-handler` carry `load_page_record`
  steps; only page-build-handler's step opts in.

## How this will be verified

1. `go test ./platform/orchestration/actions/...` (new verifier test + coverage guard).
2. After the fleet roll: pod-grep `authoritative_page_id` (positive) — note this Go
   change REMOVES no string, so the negative control per bugs_open/153 practice is
   unavailable; the deploy proof is the positive symbol + the tag + behaviour (3).
3. Behaviour: one-shot improvement loop over a site with a live link to a
   never-deployed page (bug file § How to verify). Success = the minted item's dispatch
   builds the page named by the item's `page_id`, and completion carries
   `result._verification.status='verified'`.
