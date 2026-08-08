# HANDOFF — bugfix 222 (fabrication gate negation-blindness), 2026-08-08 ~21:30 BST

> **UPDATE 2026-08-08 ~23:15 BST — steps 1-7 below are DONE. Only step 8
> (reading the council verdict) is still open.** Fix is live and pod-verified
> on chassis v1.0.1269 (both replicas). The mortgagecalculator lane has been
> told, 016b is annotated, the bug file's status line is updated in place
> (stays in `bugs_open/` per owner direction — do not move it). Council round
> 2 is in flight (`aa2d0d62-4aba-480e-aedc-8be264d53b01`, `RESUBMIT_CORR`
> already used) after the owner's credit top-up cleared round 1's fleet-wide
> block. Full detail in `NOTES_fabrication_negation.md`'s final entry. If you
> are picking this up: just check the verdict query below and act on it —
> nothing else in this workstream is open.
> ```sql
> SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
>  WHERE correlation_id='aa2d0d62-4aba-480e-aedc-8be264d53b01' AND kind='council_report'
>  ORDER BY created_at;
> ```
> APPROVED → commit `Council-Reviewed: aa2d0d62-4aba-480e-aedc-8be264d53b01`
> (docs-only, no code change). REVISE → read objections, fix, resubmit with
> the same `RESUBMIT_CORR`. Original (now largely historical) handoff below.

Written because this session's context is large (extensive research + a fable
planning pass + full implementation + mutation testing) and the remaining work
is a build/deploy/verify cycle plus a council resubmission — cleaner to pick up
fresh. Read this file, then `NOTES_fabrication_negation.md` (full detail, newest
at the bottom) if you need the "why", then act.

## What this bug is

`bugs_open/222`: the fabrication gate's declaration tier (`check_tool_fabrication_action.go`)
convicted a DENIAL of fabrication ("no fabricated data — starts empty") as a
DECLARATION of it, discarding a correct tool recreation to human review.

## What is DONE (code, tests, docs — all committed, verified against clean HEAD)

- Fix implemented: `NegationGuard` extracted from the existing claims-layer
  negation guard (CLM-017, `datahelpers/claims.go`) as an exported,
  parameterised primitive; the fabrication gate uses it with its own cue
  vocabulary. Registered as CLM-020.
- Reproduction-first tests written and confirmed red against unfixed code,
  then green after the fix.
- Full mutation-and-control triple run and recorded.
- Zero regression to the claims layer, proven mechanically (its suite passes
  with no test file touched).
- Commits, in order: `f8cbaf551` (the fix), `06241a516` (docs), `0f753107d`
  (council-blocked note). All on branch `087_towards_multiple_domains`.
- Council submitted round 1: `SUBMISSION_CORR=aa2d0d62-4aba-480e-aedc-8be264d53b01`.
  **Blocked, not reviewed** — fleet-wide Anthropic credit exhaustion
  (`step review_editquality failed ... "Your credit balance is too low"`),
  the same outage another lane (bug 220 / finetuning lane) hit the same
  evening. Nothing wrong with the submission itself.

## What is NOT done — this is the actual next step

**The fix has never been built into a chassis image.** Verified at the
artefact, not inferred: both running `agent-chassis` pods are 89 minutes old
(started ≈19:58 BST); my fix committed at 20:35:33, 37 minutes later. A
positive-control grep (`declared synthetic/fake data`, an unchanged string)
confirms the gate exists and the grep methodology is sound; the new strings
(`negated declaration ignored`, `NegationGuard`) return **zero** on both
replicas. `makefile`'s `IMAGE_TAG` is still `v1.0.1268`, unbumped for this fix.

**Do this, in order:**

1. Bump `IMAGE_TAG` in `makefile` (currently `v1.0.1268`).
2. `make build-agent-chassis` — builds from committed HEAD, so this fix (already
   committed) is included. Check the build output for "leaving out N
   uncommitted changes" — should not mention any of this fix's files (it's
   committed), but other sessions' current WIP may show up; that's normal and
   not your concern.
3. Push + deploy. Check whether this needs to go through a fleet-wide
   `make release` rather than a one-service deploy — memory flags "releases
   are WHOLE-FLEET, owner runs make release" as a standing note; if unsure,
   ask rather than assume a one-service `deploy-agent-chassis` is the right
   scope for a shared cluster.
4. **Re-verify at the pod, both replicas**, before believing anything shipped:
   ```bash
   kubectl -n ai-persona-system get pods -l app=agent-chassis
   kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c "NegationGuard"'
   ```
   Expect ≥1 on both. If 0, the build/deploy did not carry the commit — check
   `IMAGE_TAG` was actually bumped and the pods actually restarted on the new
   tag (`kubectl get pods -o jsonpath='{.items[*].spec.containers[*].image}'`).
5. **Behavioural check**: re-run the portfolio recreation item (or dispatch a
   fresh `tool-recreation-handler` run against a similar fixture) and read
   `check_fabrication` in `orchestration_states.collected_data` **same day**
   (it purges) — expect `fabricated:false` where the only match is a denial.
6. Tell `mortgagecalculator_couk_adoption` (their `HANDOFF_2026-08-08b_continue_here.md`)
   the bug-222 comment-style workaround clause can come out now it's fixed.
7. **Only after step 4 confirms live**: annotate `016b_debugging_guide_8_consolidated.md`
   §9's existing entry for this bug with the mechanism-level rule ("a
   proximity detector needs its negation guard applied at the negatable
   token's position, not the match start") — deferred deliberately until
   live, per `PLAN_2026-08-08_negation_aware_declaration_tier.md` §10.
8. **Re-check the council verdict.** Check whether the fleet-wide Anthropic
   outage has cleared (any council/LLM call succeeded recently anywhere in
   the fleet). If so, resubmit with `RESUBMIT_CORR=aa2d0d62-4aba-480e-aedc-8be264d53b01`
   so the trail accumulates (same script:
   `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh`,
   same submission JSON, add `RESUBMIT_CORR` env var per the script's own
   header). Read the real verdict before writing any `Council-Reviewed:`
   trailer on a future commit — never write it unread.

## Files that matter

- `PLAN_2026-08-08_negation_aware_declaration_tier.md` — the design, alternatives
  rejected and why, full test list, risk register.
- `NOTES_fabrication_negation.md` — the full session log, append-only, including
  two fixture-authoring mistakes caught by running tests red-first, the
  mutation-triple results (with one informative deviation from the plan's
  prediction, explained), and the live same-file-passenger hit on `WRONG_CALLS.md`.
- `RUNBOOK_fabrication_negation.md` — every command, including the clean-archive
  isolation technique for testing against a dirty shared tree, and the exact
  mutation commands.
- `submission_222_r1.json` — the council submission, reusable for resubmission.
- `bugs_open/222_HANDOFF_2026-08-08_fabrication_declaration_tier_convicts_the_denial.md`
  — the bug file itself, already updated with a "fix landed" section.

## Landmines hit this session (already logged where they belong)

- A pathspec commit still takes a same-file passenger: my `WRONG_CALLS.md`
  addition got swept into another session's unrelated commit (`d53d04786`)
  because we edited the same file concurrently in the shared tree. Nothing
  lost — confirmed present at `HEAD:docs/…/WRONG_CALLS.md`. Don't re-add it.
- A plan's illustrative code is a sketch, not a diff to paste: the plan's
  nested if/else Tier-A cascade hit a real Go scoping error on `suppressed`
  going out of scope with no `else` to hold it — rewritten as a `for` loop.
- Two of my own test fixtures didn't reproduce what they were meant to on
  first write (one matched nothing, one matched an unrelated second regex hit
  in a different clause) — caught only by requiring every new test to fail
  red against unfixed code before trusting it.
