# 336 — `deploy_result_field` is declared on `render_component`'s spec, not `update_page_status`'s, so arming it (migration 494) makes EVERY workflow that stamps a page fail validation

**Filed 2026-08-20 07:25Z by the `webdesign_tool_rebuilds` lane, which found it as a blocked
served-page grade. Status: OPEN. CAUSE IDENTIFIED, one-line fix. SERVICE RESTORED by running the
owning lane's own rollback — the fleet is NOT broken as you read this, but the defect is still at HEAD
and re-arming 494 will reproduce it.**

Owner of the code and the migration: the `bugfix_315_deployed_at_without_publication` lane
(`docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/`). This file is
a report to them, not a takeover.

## Symptom

Every work item whose workflow contains a `update_page_status` step died at validation:

```
WORKFLOW_INVALID: Invalid workflow configuration (caused by: step 'update_status'
(action 'update_page_status') has unrecognised config keys [deploy_result_field] —
this action declares its config contract as complete, so an unknown key is a
definition error, not a no-op) (code: WORKFLOW_INVALID)
```

`[MEASURED 2026-08-20 07:20Z]` 8 items carried it — 4 `page_rerender`, 2 `needs_content_page`, 1
`section_edit` (6 left `failed`, 2 bounced back to `triaged`) — with **123 `page_rerender` items queued
fleet-wide and none draining**. The three affected agent definitions were `page-rerender`,
`report-builder` and `section-editor`, i.e. the fleet's page-publishing path.

## Cause — the key is declared on the wrong action's spec

`platform/orchestration/actions/v3_site_actions.go` registers TWO specs in one `init()`:

```go
func init() {
	datahelpers.RegisterActionInputSpec("update_page_status", UpdatePageStatusInputSpec)
	datahelpers.RegisterActionInputSpec("render_component", RenderComponentInputSpec)
}
```

- **`UpdatePageStatusInputSpec.ConfigKeys` (v3_site_actions.go:550-556) is exactly five keys** —
  `status`, `page_id_field`, `site_id_field`, `page_name_field`, `page_component_id_field` — and the
  spec sets **`StrictConfig: true`** (:637), with a comment asserting a census where "every key they
  set is one of the five ConfigKeys above. Zero unrecognised."
- **`deploy_result_field` is declared in `RenderComponentInputSpec.ConfigKeys` instead** (:674), with
  the `bugs_open/315 / RFC_038` comment attached to it — inside the spec for the OTHER action, which
  never reads it.
- **The reader is in `UpdatePageStatusAction`** (:982): `if field, _ := config["deploy_result_field"].(string); …`.

So the action that reads the key does not declare it, and the action that declares it does not read it.
Because the reader's spec is `StrictConfig: true`, the validator's strict branch
(`platform/validation/workflow.go:188-193`, via `datahelpers.UnknownConfigKeys`) turns the key into a
**hard error for the whole workflow** the moment any live step sets it.

## Why the arming precondition did not protect it

The 315 lane recorded, correctly, that 494 must not be armed until a build carrying `f0dd97c71` had
rolled. **That precondition was MET and the arming still broke the fleet:** `v1.0.1317` (pods started
2026-08-19 22:26Z, stamp `2d13d530d`) has both `086f9b7b7` and `f0dd97c71` as ancestors. The condition
was about the READER shipping; the defect is in the DECLARATION, which neither commit put on the right
spec. `[MEASURED]` 494 was applied at **06:49:49Z**; the first failure was **07:01:50Z**, twelve
minutes later, when the first item was claimed.

**A binary probe cannot see this and will mislead you**: `grep -aq "deploy_result_field" /proc/1/exe`
on the chassis returns PRESENT — because the literal is in the binary three times over (the reader, two
`zap.String` calls, and the wrong spec). Presence of the literal is not membership of the right list.
Likewise `git log -S'"deploy_result_field",'` names `086f9b7b7` and looks like the declaration's
commit; that match is the `zap.String("deploy_result_field", field)` call. Check the LIST, at the
commit, inside the right spec.

## The fix (one line, and it needs a roll)

Move `"deploy_result_field"` from `RenderComponentInputSpec.ConfigKeys` into
`UpdatePageStatusInputSpec.ConfigKeys`, keeping the RFC_038 comment with it, and correct the "five
ConfigKeys" census comment to six. Then rebuild + roll, then re-arm 494.

Do NOT "fix" it by relaxing `StrictConfig` on `update_page_status`: the strictness worked exactly as
designed — it caught a definition error at validation instead of letting the stamp silently not read
its evidence, which is the failure mode `bugs_open/234` and this action's three retired keys already
paid for.

**Worth considering as the durable guard:** a test that every key a registered action READS
(`config["…"]` in its own function) appears in ITS OWN spec, and that no spec declares a key its action
never reads. This defect is invisible to a per-spec review because both halves are in the same file,
forty lines apart, and each looks right on its own.

## What was done to restore service (2026-08-20 07:22:40Z)

Ran the owning lane's own rollback verbatim —
`docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD_ROLLBACK.sql` — which snapshots
all three definitions and removes the key. Verified: all three read
`default_config::text LIKE '%deploy_result_field%'` = false at 07:22:40Z. Re-arming is one command
(`494_…_HOLD.sql`) once the declaration moves and a build carrying it has rolled.

**Service proven restored with DEMAND, not with an absence** (07:25:35Z claim → **07:25:59Z complete**
on `tools-index`, then `learn-index` at 07:26:24Z — the first completions since the arming). This
matters because "no failures" would have been worthless evidence: the hour before the incident also
shows zero rerender completions, simply because nothing was queued.

**The six items that failed on the defect are handled.** Two were already re-filed by the platform
itself — a `failed` row releases the `idx_swi_dedup` slot, so `page_rerender_tool-shadow-stacker_…`
and the `index` rerender each had a live replacement, and attempting to flip them back gave
`duplicate key value violates unique constraint "idx_swi_dedup"`. **That refusal is the signal the
work is already queued**, not an obstacle. The other four had no replacement and were flipped
`failed → triaged` with `claimed_at` cleared and the reason written into `error` (07:29Z): the three
`needs_content_page` guide-page writes for this lane's tools (`9cb5d4e5`, `e291e4ea`, `35972e9b`) and
the `section_edit` template-fix delivery `126c586a`, which belongs to the `bugs_open/283` lane and is
a pure re-render from the fixed template by their own description. A status flip creates no new row,
so no two-strike attempt accrues.

⚠ **One instrument to distrust while cleaning this up: `WHERE error LIKE '%deploy_result_field%'` is
NOT a failure census.** The `error` column keeps its text after a row later succeeds, so that filter
reported "2 new failures since the disarm" when both rows were `complete` — they were the two items
that had bounced to `triaged` and then ran fine. Filter on `status`, and use `error` only to attribute.

## How to verify a fix

1. The declaration is in the right list: `git show <commit>:platform/orchestration/actions/v3_site_actions.go`
   → `deploy_result_field` inside `UpdatePageStatusInputSpec.ConfigKeys`, absent from
   `RenderComponentInputSpec.ConfigKeys`.
2. The build carrying it is live, by ANCESTRY of the pod's stamp, not by tag and not by a literal grep.
3. Re-arm 494, then watch a real `page_rerender` complete — and check `pages.content_hash` becomes
   non-zero, which is what 315 exists for.
4. **Demand control**: assert a rerender COMPLETED after the re-arm. "No failures" is not evidence
   while the queue is empty — the hour before this incident shows zero completions too, because
   nothing was queued.
