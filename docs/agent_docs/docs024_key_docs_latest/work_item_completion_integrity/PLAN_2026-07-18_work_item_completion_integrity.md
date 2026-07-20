# PLAN — work-item completion integrity (bugs_closed/017)

**Started:** 2026-07-18 (session "bugfix thread2"). **Status: CLOSED 2026-07-20** — live in
v1.0.1139, verified against the running pod with discriminating strings; case moved to
`/bugs_closed/017_…`.

## The problem, in one line

`site_work_items` could be stamped `status='complete'` while storing, in the same
UPDATE, the error proving the work never happened.

## Design

Two layers, deliberately separated.

**Layer 1 — the instance.** `fix_forced_text_colors` was written (handler +
`ActionInputSpec`) but never added to `GlobalActionRegistry`. The workflow validator
(`platform/validation/workflow.go:69,80`) classifies an unregistered action as REMOTE
and demands a `topic`, so every dispatch died with "requires a topic". Fix: register it.

**Layer 2 — the class (the actual point).** `CompleteWorkItemAction` read the DELIVERY
envelope (`result.response_status`, set to `'complete'` by the coordinator on ANY reply)
instead of the saga's verdict (`result.response.status`). Fix: `handlerReportedFailure`
blocks completion on an explicit failure verdict and routes into the EXISTING attempt
machinery via a generalised `failUnverifiedCompletion`.

**Layer 3 — recurrence.** `registry_parity_test.go` fails the build when any action
registers an `ActionInputSpec` with no registry entry, with a `dormantActions` allowlist
that turns "orphaned by accident" into "dormant by decision".

### Decisions and their reasons

| Decision | Why |
|---|---|
| Guard keys on an explicit failure verdict, NOT on `response.error` being present | Handlers legitimately carry a non-fatal error string beside a success. Pinned by a test so a later widening to "any error present" fails. |
| Guard runs BEFORE the per-item-type verifier | A saga that failed outright is not worth verifying — and the verifier is opt-in per item_type, which is exactly how this class escaped (`hardcoded_section_colors` has no verifier). |
| Unknown verdict COMPLETES (does not block) | A novel status is not evidence of failure. Inverting would trade a silent pass for retries of real work. |
| …but is recorded to `agent_error_log`, not just `zap.Warn` | Council objection (bug_historian, both rounds). Pod logs are ephemeral and do not survive rollouts, so a new failure dialect would leave no queryable trace — the same silent-failure shape as 017 itself. |
| Anti-drift as a TEST, not a startup panic | Fails the build gate, not the fleet. |
| Deleted `actioncheck.LocalActions` rather than updating it | It was dead (zero live references) AND actively misleading — see the correction below. |
| Data correction to `failed`, not re-queued | Owner's call after discussion. `failed` is terminal for `idx_swi_dedup` exactly as `complete` is, so it releases the dedup slot and discovery can re-file fresh — without dispatching 49 unverified colour edits at 5 live sites. |

> **CORRECTION to the originating brief (`bugs_open/017`), 2026-07-18.** The filed report
> blamed Defect 1 on **drift between two hand-maintained rosters** — `registry.go` versus
> a DEPRECATED `actioncheck/local_actions.go`. **That was wrong.**
> `actioncheck.IsLocalAction` (`actioncheck.go:20`) delegates to a checker `registry.go`
> installs at `init`; the `LocalActions` map's own lookup was commented out
> (`local_actions.go:185-188`) and the map had **zero live references repo-wide**. It was
> dead, not drifting. There was only ever ONE live list. The
> `batch_webscrape_action.go` comment instructing authors to "register in TWO places"
> seeded the belief and survived long enough to misdirect the diagnosis of the bug it
> caused. **Caught by:** grepping the symbol for live references instead of trusting the
> "DEPRECATED" header. Deleted the map, the comment, and two live guide docs repeating it.

> **CORRECTION to the brief's scale, 2026-07-18.** The report recorded 2 affected items.
> The live sweep found **54**, across 6 sites and 4 item types, back to May 2026.
> It also missed a *second cause* reaching the same symptom: a seed naming
> `render_js_snippets` where the registry has `render_js_snippets_for_site` — a typo, not
> an unregistered action. Only the Layer-2 guard covers both.

## Phasing

1. ~~Register the action~~ — landed in HEAD via `06376bcbf` (another session swept my
   working tree mid-task; nothing lost, per the git rules).
2. ~~Guard + generalised failure path~~ — `c82b2872c`.
3. ~~Parity test + dead-map deletion + doc purge~~ — `c82b2872c`.
4. ~~Unknown-verdict → `agent_error_log`~~ — `c80fffc83` (council round 2).
5. ~~Data correction, 54 rows~~ — applied, reversible via `result._correction`.
6. ~~Ship a chassis image, then verify against the RUNNING pod~~ — **v1.0.1139, done
   2026-07-20.** Verified with symbols that cannot exist unless the change shipped (the
   registry entry's own Description text; the guard's error message), plus positive
   controls. The obvious grep — the action name — was a false pass: that string predated
   the fix. See NOTES misstep 3 and 016b §9.

## Open questions for the owner

- None outstanding. Residual monitoring only: the guard's *blocking* path has not yet
  fired in production (nothing has failed since deploy). It will surface as
  `site_work_items.error LIKE 'completion blocked: handler saga reported failure%'`, or
  `agent_error_log.error_code='UNKNOWN_HANDLER_VERDICT'` for an unfamiliar verdict.
