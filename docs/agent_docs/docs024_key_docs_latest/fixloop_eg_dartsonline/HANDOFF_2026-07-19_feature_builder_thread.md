# HANDOFF — resume the FEATURE BUILDER thread

**Filed 2026-07-19** from the "fixloop feature builder" session. Cold-start for
a NEW chat. Supersedes `HANDOFF_2026-07-17_feature_builder_thread.md` (that one
described a thing not yet built; this one describes a thing that runs).

## Read these first, in this order

1. `SUMMARY_feature_builder_2026-07-19.md` — plain-language, 3 minutes.
2. `PLAN_feature_builder.md` — architecture as built, status table, open
   council objections, backlog.
3. `RUNBOOK_feature_builder.md` — **your tasks are B1–B4**; A1–A5 are done.
4. `NOTES_running_feature_builder.md` — turn-by-turn record, 13 turns.
   Turns 7–13 are the live-run history and are where the lessons are.

Repo rules that override everything: `/CLAUDE.md` (read it fresh — it grew
twice during this thread). Parent design:
`DESIGN_feature_builder_and_council_gate.md` §1; schema:
`SCHEMA_staged_plan_v1.md`; delta-2 design: `DESIGN_stage_loop_delta2.md`.

## State in one paragraph

The feature builder is BUILT and LIVE. Its designing half is proven: five live
runs, the last one **approved unanimously by a five-seat council**. Its
building half — `feature-implementer`, the stage loop — is seeded, live, and
**has never executed once**. That is the entire remaining gap. Everything in
the RUNBOOK's B tasks exists to close it.

## What is live right now

| Thing | Where | State |
|---|---|---|
| `feature-designer` | agent_definitions | LIVE, 5-seat council, 5 runs, 1 unanimous approval |
| `feature-implementer` | agent_definitions | LIVE, **never fired** |
| `feature-implementer-orchestrator` | agent_definitions | LIVE, never fired |
| staged-v1 validation | `diagnose_persist_fix_plan_action.go` | live (image v1.0.1132+) |
| `feature_stage_route` + seams | `feature_stage_route_action.go` etc. | live, **never executed** |
| Triggers | `0NN_TRIGGER_feature_{designer,implementer}_v1.sh` | designer proven; implementer unused |

Commits that matter: `4b3d50f4c` (validation), `c19b5d097` (stage loop),
`62018e272` (D4 refinement), `9c94cc842` (fail-loud fix), patches `016`/`017`
on the designer row.

## The five things this thread learned the hard way

1. **A first fire always finds a defect.** Five designer runs, five real
   defects — two prompt-steer gaps, one over-strict validation rule of our
   own, one severed feedback loop, one high-severity fail-loud bug found by
   the council gate. Expect the implementer's first fire to do the same. That
   is the point of firing it, not a reason to delay.
2. **Verify against the RUNNING POD, never git.** `base_branch_field` was
   committed but absent from the live binary; a seed using it would have been
   silently inert. Pod-grep, always.
3. **The row moves under you.** The F1.2 target was hand-fixed by another
   thread 60 seconds before our run started, making our approved plan a
   regression if applied. Check `site_work_items` AND `git log` before
   choosing a target, and re-check immediately before firing.
4. **Editing an existing live agent row must be a surgical `jsonb_set`.**
   Never a whole-column `default_config` replacement — these rows are
   co-edited and your view of them is always partial. The council rejected
   exactly that shape, and it matches the concept-register thread's own rule.
5. **`{{.X.result}}` in a prompt template renders `<no value>` silently.**
   (bug 016.) Config dot-paths keep `.result`; templates must not. Our reviser
   was blind for three runs because of it. Patch 017 now has the reviser read
   the `council_report` artifact, so roster growth can't reopen the wound.

## Do NOT do these

- Do **not** apply the approved plan from run `8e837814` — F1.2 was fixed by
  hand; applying it would regress a working fix. Work item `db066cac` is
  closed with the reason in its spec.
- Do **not** fire `feature-implementer` directly — only via
  `feature-implementer-orchestrator`, or it dies with no read token.
- Do **not** re-apply `0NN_feature_designer.sql` wholesale without first
  diffing against the LIVE row: patches 016 and 017 are surgical and a full
  re-apply would revert them unless the seed file is current (it is, as of
  this handoff — but check).
- Do **not** spend credits without the owner's per-run go.

## Open items, with owners

**Owner's call:**
- B1: choose a fresh pilot target (the old one is spent). Ideal shape: at
  least one Go edit AND one new file, so the stage loop, allowlist, build gate
  and derived test gate all get exercised.
- Whether to resubmit delta 2 to the council gate
  (`RESUBMIT_CORR=5a65ec4c-686c-40c7-813e-7c7fce03a779`) before the first
  implementer fire.

**Ours, when work resumes** — the four open council-gate objections on delta 2
(full text in `PLAN_feature_builder.md`): declare `registry.go` as its own
edit; mitigate `expected_symbols`' false-reject via designer-prompt guidance;
write a real answer to the reuse-agent's `site_work_items`-sequencing
question (our view: no — that's queueing, this is in-run state); add
travelling PLAN+NOTES subjects for the three shared actions.

## Coordination notes for whoever picks this up

Several threads work this repo and cluster at once. During this thread alone:
three commits were swept into other sessions' commits (nothing lost —
forward-only), the council roster grew 3→5→13 seats, the makefile build
default inverted, the coverage report was fixed mid-use, and the pilot target
was completed by someone else. None of that is unusual here. Commit narrowly
and immediately, re-read `git status` before acting on it, and check the queue
before dispatching anything.
