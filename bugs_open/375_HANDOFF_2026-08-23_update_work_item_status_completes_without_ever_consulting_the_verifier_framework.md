# 375 — `update_work_item_status` stamps `complete` without ever consulting the verifier framework, so one of the three completion writers has no false-completion guard at all

**Filed 2026-08-23** by the `bugfix_367_router_remit` lane, found while tracing why a router's
wrong close was silent. **Status: OPEN, UNOWNED.** Deliberately NOT fixed inside `bugs_open/367`:
this changes what a **shared completion path** guarantees, which is architecture-scope under the
owner ruling of 2026-07-29, and `bugs_closed/124` drew a REJECTED verdict for exactly that shape
of change arriving inside a bug patch.

## 1. The defect in one paragraph

The platform has a **verifier framework** whose entire purpose is stated in
`platform/orchestration/actions/discovery_checks/verifier_coverage_test.go`:

> *"`CompleteWorkItemAction` consults a per-item_type verifier before stamping 'complete'. …
> That is the same class as `bugs_open/017` (a saga reporting success without touching the
> defect), one level up: 017 stops a saga that says it FAILED from being stamped complete; a
> **verifier is what stops one that says it SUCCEEDED but did nothing**."*

There are **three** writers of `complete` on `site_work_items`. `CompleteWorkItemAction` consults
the verifier. `UpdateWorkItemStatusAction` — `platform/orchestration/actions/v3_site_actions.go:6010`,
registered as action `update_work_item_status` (`registry.go:939`) — **does not**. Its `complete`
arm (`v3_site_actions.go:6290-6300`) carries only the *terminal-decision* guard:

```sql
UPDATE site_work_items
   SET status = $2, completed_at = NOW(), updated_at = NOW(),
       attempt_count = attempt_count + 1,
       result = COALESCE(result,'{}'::jsonb) || $3::jsonb,
       error  = COALESCE(NULLIF($4,''), error)
 WHERE id = $1
   AND status NOT IN (workItemCompletionGuardStatuses)
```

`GetVerifier` is never called on this path. The code's own comment says as much about what it *did*
add, and the omission is visible in it:

> *"The `complete` arm carries the terminal-decision guard `CompleteWorkItemAction` has (WII-003,
> load_work_item_actions.go) — this action is a **third writer of `complete`** and had no guard at
> all … Same defect, same remedy, one writer over."*

The remedy that was carried over was the terminal-decision guard (don't overwrite a row that
already failed or was given up). The **verifier** was not.

## 2. Why this matters — it is the reason `bugs_open/367` was silent

`bugs_open/367` is one instance. `required-fields-missing-handler`'s `close_stale` step uses
`update_work_item_status` with `status: complete`. A true finding — schema-required fields
genuinely empty on a component that genuinely exists — was stamped `complete`, `attempt_count 1`,
`error` NULL, and disappeared into the "actioned" bucket. A verifier for
`required_fields_missing` would have re-run the finding's own predicate and refused the completion.

367 was fixed at the router (migration `574`, live 2026-08-23) so that particular wrong close can
no longer be constructed. **That fixes one disposer's reasoning; it does not restore the guard.**
Any agent definition can call `update_work_item_status` with `status: complete` from DB config,
with no code change and no review — which is precisely the property `bugs_open/213` relies on when
it refuses to enumerate producers in code.

## 3. Evidence `[MEASURED 2026-08-23]`

- `grep -rn "RegisterVerifier\|GetVerifier" platform/orchestration/actions/` — **12+ registered
  verifiers** (`content_duplication`, `decision_regression`, `empty_section`,
  `unbuilt_internal_link`, `page_canonical_collision`, `literal_markdown`,
  `hardcoded_section_colors`, `orphan_element_refs`, `truncated_component`, `revenue_shape_cta`,
  `missing_conversion_path`, `dead_fragment_link`, `needs_brand_head_assets`, …) — and **no call
  to `GetVerifier` anywhere in `UpdateWorkItemStatusAction`**.
- `required_fields_missing` is itself an **unverified type by declaration**:
  `verifier_coverage_test.go:237` lists it `{catMechanical, "carries page_id and component_id"}`.
- ⚠ **A count is owed here and this file does not have it.** How many live agent definitions
  reach `complete` through `update_work_item_status` rather than `complete_work_item` — and how
  many of those carry an item_type that *does* have a registered verifier — is the number that
  sizes this bug, and it was not measured. Start here:
  ```sql
  SELECT type, jsonb_path_query_array(default_config, '$.**.action') @> '["update_work_item_status"]' AS uses_uwis
  FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
  Then cross that against the registered verifier list. **Do not size this from `bugs_open/367`
  alone** — that is one router, found by accident.

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Consult the verifier on this path too**, behind an **opt-in step-config key whose unsafe
   default is OFF** — the shape the owner ruled for new authority on a shared seam
   (2026-08-02 §2: *"when a seam's widest branch is licensed by 'callers must all be X', make X a
   field with the unsafe default OFF"*). Arm it per step, so a reviewer of the **caller** sees the
   decision. Note `RFC_022`'s narrowing: an opt-in field with an unsafe default OFF that no live
   consumer names is **not** architecture-scope — but the moment a consumer names it, and
   certainly if the intent is to arm it fleet-wide, it is.
   ⚠ **`CQ-023` already warns of a live consequence:** a verifier registered for
   `required_fields_missing` *"would fail-closed the `converted` arm's completion"*. So arming is
   not free per type, and whoever arms one must read that type's close paths first.
2. **Make the two writers one.** Have `UpdateWorkItemStatusAction`'s `complete` arm delegate to
   `CompleteWorkItemAction`'s guarded path, the way `workItemHandlerRegisteredSQL` was unified in
   `bugs_closed/284` (owner ruling 2026-08-17) with a structural single-definition test to stop a
   fourth copy appearing. Cleanest, largest blast radius, and squarely architecture-scope.
3. **Do nothing, but record it honestly** — the verifier framework's coverage guard
   (`verifier_coverage_test.go`) currently reads as though registering a verifier protects a type.
   For any type completed via `update_work_item_status`, it does not. At minimum that test's header
   should say so, or its promise is overstated for an unmeasured share of the fleet.

## 5. How to verify a fix

Register a verifier for a type that is completed via `update_work_item_status`, make its predicate
return "still failing", drive one item through, and require the completion to be **refused** — then
the negative control in the same run: the same path with the verifier's predicate satisfied must
still complete. Assert at the item's status and at the verifier's own record, not at the saga's
report — the saga reporting success is the thing under test.

⚠ A mock's own bookkeeping cannot assert this negative: mutate the guard to prove it is load-bearing
(`LANDMINES.md`, *"a mock's own bookkeeping cannot assert a NEGATIVE"*).

## 6. Where the record lives

`docs/agent_docs/docs024_key_docs_latest/bugfix_367_router_remit/` (found here, in NOTES and
README_where_we_are). Related: `bugs_open/367` (the instance), `bugs_open/017` (a saga reporting
success without touching the defect), `bugs_open/213` (`GradesFunc` — why a verifier keyed on
item_type alone mis-grades a second producer's items, and why a code-side producer list is
refuted), `bugs_open/021` §INSTANCE 2 (the coverage guard), `bugs_closed/284` (the
single-definition precedent for unifying duplicate writers), register `CQ-023`, `WII-003`.

**No `090` diagnosis run.** Stated plainly per the owner ruling of 2026-07-31, because this file
asserts a structural property. The substitute is first-hand verification and it is direct: the
action was read end to end, its `complete` arm quoted above, and the absence of any `GetVerifier`
call on that path confirmed by grep over the package. **What is NOT established is the blast
radius** — §3 says so and gives the query. A thread taking this on should run that census before
choosing between the candidates, and should file `090` if it intends to assert a cause beyond
"the call is absent".
