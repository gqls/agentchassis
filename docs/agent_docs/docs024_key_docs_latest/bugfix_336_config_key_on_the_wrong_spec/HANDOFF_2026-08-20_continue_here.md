# HANDOFF — `bugs_open/336` (config key on the wrong action spec) and the 494 re-arm. START HERE. Written 2026-08-20 14:40Z.

**One-line state: the fix is committed, council-submitted, LIVE on `v1.0.1319`, and 494 is re-armed with
zero validation failures. Exactly ONE thing is unfinished, and it is a waiting game, not work.**

Full case file: `bugs_open/336_HANDOFF_2026-08-20_deploy_result_field_is_declared_on_the_wrong_actions_spec_so_arming_it_hard_fails_every_workflow_that_stamps_a_page.md`

## What happened, in four sentences

`update_page_status` reads `config["deploy_result_field"]` and its spec sets `StrictConfig: true`, but
the key was DECLARED on `RenderComponentInputSpec` — the other spec that `v3_site_actions.go`'s `init()`
registers, for an action that never reads it. When migration 494 armed the key on live
`update_page_status` steps at 06:49:49Z, the strict branch of `validation.checkStepConfigKeys` hard-failed
**every workflow that stamps a page**: 8 items across 4 item types, 123 `page_rerender` queued fleet-wide
and none draining. I restored service at 07:22:40Z with the owning lane's own snapshot-backed rollback,
proved recovery with a completed rerender rather than an absence of failures, then fixed the declaration.
The fleet rolled at 10:18Z and someone re-armed 494 at 14:27:34Z.

## Done, with the evidence

| | |
|---|---|
| fix commit | **`daaa7541b`** — key moved to `UpdatePageStatusInputSpec`, removed from `RenderComponentInputSpec`, two key-counting comments corrected, plus `platform/orchestration/actions/update_page_status_config_contract_test.go` (3 tests) |
| verified | at that commit in a clean worktree: `go build ./platform/...` rc=0; the 3 tests PASS and **all 3 FAILED against the pre-fix state**; `go test` green on `actions`, `validation`, `datahelpers` |
| council | `bc2f4b0e-45db-49c8-9f45-6af74a344cce` SUBMITTED (fix commit carries no trailer — see "loose ends") |
| live | `v1.0.1319`, image label revision `447f3a8a84428061059abdeeb4bb1d524941dbb7`, and `git merge-base --is-ancestor daaa7541b 447f3a8a8` TRUE; control `759eea9d6` correctly NOT an ancestor; `git-adapter` + `core-manager` same revision |
| 494 | re-armed 14:27:34Z (by another session; my run refused with `already applied`) |
| post-arm | **0** items carrying `unrecognised config keys` since 07:22:40Z; **122** rerenders completed between the roll and the arming |

## THE ONE OPEN ITEM — and why its zeros mean nothing yet

`pages.content_hash` is non-null on **0** rows and `agent_error_log` has **0**
`DEPLOY_EVIDENCE_UNREADABLE` rows since the arming. **Neither is evidence.** `page_rerender` queued
fleet-wide is **0** and **0** rerenders have completed since 14:27:34Z — there has been no demand through
the armed path at all. Do not report this as a pass.

**The demand is already on its way; do not force it.** The parallel `webdesign_tool_rebuilds` session is
mid-rebuild (#22 `tool-oklch-picker`); every rebuild retires a slot and re-renders its page, which is a
`git_commit` → `update_status` pass on a page that lane owns. That is the content-neutral instrument the
315 lane wanted and explicitly declined to fabricate. Check with:

```sql
SELECT name, left(content_hash,12) AS hash, deployed_at::timestamp(0)
FROM pages WHERE content_hash IS NOT NULL AND content_hash <> '' ORDER BY deployed_at DESC LIMIT 5;

SELECT error_code, count(*) FROM agent_error_log
WHERE occurred_at > '2026-08-20 14:27:34+00' AND error_code LIKE 'DEPLOY_EVIDENCE%' GROUP BY 1;

SELECT count(*) FROM site_work_items WHERE error LIKE '%unrecognised config keys%'
  AND updated_at > '2026-08-20 14:27:34+00';   -- must stay 0; if it moves, disarm 494 first, ask later
```

- **a hash appears** → this is what `bugs_open/315` exists for; tell the 315 lane, and 336 can close.
- **`DEPLOY_EVIDENCE_UNREADABLE` instead** → now a REAL defect in the resolution path (both halves and the
  declaration are live). First thing to read is the field name per agent: three definitions carry
  `deploy_result_field`, and `section-editor`'s deploy step outputs **`git_result`**, not `deploy_result`.
- **`unrecognised config keys` returns non-zero** → the fix did not take on some path. Disarm immediately
  with `docs/agent_docs/sql_for_agents/494_stamp_reads_deploy_evidence_HOLD_ROLLBACK.sql` (snapshot-backed,
  verified, one command) and then diagnose. Restoring the publish path beats understanding it.

## Loose ends, none blocking

1. **The fix commit carries no council trailer.** I committed before submitting because the file was
   being edited by another session and my first copy of this fix had already been lost to their commit
   (see the case file); forward-only forbids an amend. The correlation is recorded in `bugs_open/336`, so
   the `098` report will list `daaa7541b` as un-reviewed. If the verdict is APPROVED, that is a reporting
   artefact, not a coverage gap.
2. **Read the verdict**: `SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%bc2f4b0e%' ORDER BY created_at DESC LIMIT 1;`
   If it comes back REVISE, the objections arrive with the reviewers' own checks answered — revise and
   resubmit with `RESUBMIT_CORR=bc2f4b0e-45db-49c8-9f45-6af74a344cce`.
3. **The general guard is NOT built.** 336's class — an action reading a config key its own spec does not
   declare — is invisible to every current instrument. `cmd/config-key-audit --suspicious-keys` only
   inspects key NAMES for documentation punctuation. My test covers `update_page_status` alone. A
   fleet-wide read-vs-declared mode belongs in that tool, where the source-scanning machinery lives, and
   it must follow reads through helpers (`shouldRefuseDeadURLControls(config, …)`) rather than grepping
   for `config["`.
4. **Sibling-key check is INCONCLUSIVE, not clean.** Two keys on `RenderComponentInputSpec`
   (`refuse_dead_url_controls`, `refuse_mistyped_llm_fields`) returned zero read-sites to a crude grep
   `[UNVERIFIED]`. Almost certainly a grep artefact — the spec's own comment says the first is read
   through a helper — and in any case that spec is warning-only, no `StrictConfig`, so a misplaced key
   there is inert rather than an outage. Settle it with (3), not with another grep.

## Do NOT pick up

**The `webdesign_tool_rebuilds` tool queue.** A second session has owned it since ~07:14Z and is at
**20 of 63** with #22 in flight; its cold-start is
`docs/agent_docs/docs024_key_docs_latest/webdesign_tool_rebuilds/HANDOFF_2026-08-19_continue_here.md`
(they refreshed the header at 09:5xZ). Filing a rebuild there would collide on the `add_tool` item key
and duplicate their brief-writing. My own contributions to that lane up to #18 are in its NOTES.

## Two traps this incident produced, both already in LANDMINES.md

- **A config key declared on the wrong action's spec passes every instrument** — the literal is in the
  binary from the reader and its `zap.String` calls whichever spec lists it, `git log -S` matches the zap
  call and so names the reader's commit, and an arming precondition written about the reader is satisfied
  too. Footprint: `RegisterActionInputSpec`, `ConfigKeys`, `StrictConfig`, `UnknownConfigKeys`.
- **A chassis pod retains ~2 minutes of logs** and `--since=6h` is silently clamped by rotation, so
  grepping for a line your code emits is only valid within about a minute of the event. If the branch
  writes a ROW, assert the row instead — it has no shelf life. (Entry refined with today's re-measurement.)
