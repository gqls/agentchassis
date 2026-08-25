# HANDOFF 2026-08-25 — `bugs_open/345`: closeable; what it hands on is the NEXT defect, not a residual

**Cold-start for a fresh session.** Read this, then the bug file's last four sections (dated 08-24
night → 08-25 morning). Everything here was measured by the session that wrote it; re-measure
before acting — every count below is dated and the log arm rots in ~31 days.

## Where we are, in one paragraph

345 filed a defect with two harms: *the writer never learns why it was refused* and *one item burned
52 generations*. Both are now fixed, live and measured. The feedback channel (candidate 1, WII-026)
is demand-proven. The repeat-termination mechanism (candidate 2, **WII-029**) is live on chassis
`v1.0.1336+`, opted in for `needs_new_component` by migration `607` (applied 2026-08-24 20:23:00Z),
and **has fired four times**, each leaving a durable `result.terminated_on_repeat` marker. The
rejection message now names every offending field, deterministically. Council: mechanism APPROVED
(`f1f1fc37`); completion round 1 REVISE (`cf086b8d`, gating objection answered by the four firings
themselves), round 2 resubmitted — read the verdict before writing `Council-Reviewed:` anywhere.

## The decisions the owner has (as of 2026-08-25 morning)

1. **Close 345.** Every piece is fixed-and-live-and-measured; the file says so with evidence. Nothing
   in the original defect is open. Recommended: close, and open a new file for item 2 below.
2. **Orphan-only rejections: keep blocking, or make non-blocking?** `recordValidationRejection`
   scores an orphan-only schema/template mismatch as **`warning`** (a declared-but-unrendered field
   wastes content tokens; it does not break a page), yet the store gate at
   `store_generated_component_action.go:324` **blocks** on it. [MEASURED 2026-08-25, all history]
   **9 items died with every rejection orphan-only (21 rejections); 1 completed.** All four
   candidate-2 firings are this class (`*_label` fields on article/category grids — the writer
   renders the CTA label as static text and declares a field for it anyway). **Naming the field to
   the writer does NOT fix it** (measured, n=2, both refuting). Options: (a) store + record the
   warning; (b) auto-drop the orphaned fields at the gate — lossless for rendering; (c) leave as is —
   candidate 2 now caps the cost at one wasted retry per item, but the pages stay hollow. This
   changes what a shared gate guarantees, so it is the owner's call and probably an RFC (2026-07-29
   §1), not a bug patch.
3. **Register the opt-in key via `RegisterActionInputSpec`** — on BOTH failure writers
   (`update_work_item_status` has **no spec at all**), so the RFC_022 optional-key budget can see
   it. Small; a follow-up, not a blocker.
4. **Options struct for `applyWorkItemFailureLadder`** — 8 positional params; the 8th broke two
   sibling sqlmock files. Low urgency; do it before the 9th.

## Evidence you can re-run

```sql
-- firings (durable)
SELECT left(id::text,8), status, attempt_count||'/'||max_attempts, updated_at
  FROM site_work_items WHERE result ? 'terminated_on_repeat' ORDER BY updated_at;
-- proxy (blind to last-attempt firings; ~7-day archiver shelf life)
SELECT count(*) FROM site_work_items WHERE item_type='needs_new_component' AND status='failed'
   AND attempt_count < max_attempts AND updated_at > '2026-08-24 20:23:00+00';
-- the opt-in, read from the live row (a top-level walk does NOT see this step)
SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'->'steps'->'mark_failed'->'config'
  FROM agent_definitions WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- the asymmetry (decision 2)
WITH per AS (SELECT work_item_id wid, bool_and(error_code='component_validation_orphan_schema_field') orphan_only, count(*) n
  FROM agent_error_log WHERE error_code LIKE 'component_validation%' AND work_item_id IS NOT NULL GROUP BY 1)
SELECT w.status, per.orphan_only, count(*), sum(per.n) FROM per JOIN site_work_items w ON w.id=per.wid GROUP BY 1,2;
```
```bash
# liveness, per pod, with control — probe LITERALS/FUNCTION NAMES, never a Go field name (LANDMINES entry, 2026-08-25)
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do n=${p#pod/}
  for s in terminated_on_repeat describeSchemaTemplateMismatch stop_on_repeat_failure zz_no_such_symbol_345; do
    kubectl -n ai-persona-system exec "$n" -- grep -aq "$s" /proc/1/exe && echo "$n PRESENT $s" || echo "$n absent $s"; done; done
# council verdict
# SELECT current_step,status FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id'='<corr>';
```

## How a firing must be READ (or the census lies)

- A firing means *"feedback was shown and did not help"* **only** when the item's terminal error
  carries one of the 3 producer validation codes (`component_validation_rejected` /
  `_orphan_schema_field` / `_unknown_template_var`) — the prompt renders for those alone (mig 563).
  On any other class it means *"this class has no feedback channel"*.
- The `Handler failed` fallback literal on `mark_failed` is a latent false-match channel (two
  DIFFERENT failures both falling back to it look identical). **0 rows** carry it as of 607's apply;
  607's NOTICE prints the live count every time it is (re)applied.
- Attempts spanning the `v1.0.1336` roll straddle two wordings of the mismatch message and will not
  match — one transition attempt per in-flight item, self-resolving.

## Traps this lane hit (all in `LANDMINES.md` / `WRONG_CALLS.md`, dated 08-23 → 08-25)

- A `git mv` committed by one pathspec leaves BOTH names at HEAD (and `524`'s `_HOLD` twin is still
  the live instance — not ours to fix).
- A demand control that shares the instrument's CHANNEL but not its PREDICATE validates nothing (the
  first tripwire); a `-run` filter is the same error in test clothing.
- A mutation that PASSES probably hit a guard in series (the helper's sort behind the producer's).
- A mutant that does not COMPILE proves nothing.
- Probing a binary for a struct field name reads absent while the feature is live.
- "The message is deficient" explains a failure; it is not evidence that fixing it produces success.
- Register ids: re-read the file's TAIL immediately before taking the next number (WII-027 collided).

## Files

- Bug: `bugs_open/345_HANDOFF_2026-08-21_a_rejected_component_is_regenerated_from_identical_inputs_so_the_writer_never_learns_why_and_one_item_burned_52_generations.md`
- Register: `docs/agent_docs/docs026_concept_register/register/work-item-integrity.md` — **WII-026** (feedback channel), **WII-029** (repeat termination + marker; filed as 027 in commit `980973491`, renumbered)
- Code: `platform/orchestration/actions/work_item_failure_ladder.go` (+`_test.go`), `compute_component_quality.go`, `store_generated_component_action.go` (`describeSchemaTemplateMismatch`), `store_generated_mismatch_message_test.go`
- Config: `docs/agent_docs/sql_for_agents/607_*.sql` (+`_ROLLBACK`)
- Council: `f1f1fc37` (mechanism, APPROVED r1) · `cf086b8d` (completion, REVISE r1 → resubmitted; corr in the follow-up commit)
- Sibling lane docs: this directory's `CONTRIB_2026-08-24_…` (the 326 lane's commit-message correction); `bugfix_337_token_cap/` (truncation class, deliberately unbuilt, tripwire in the bug file)
