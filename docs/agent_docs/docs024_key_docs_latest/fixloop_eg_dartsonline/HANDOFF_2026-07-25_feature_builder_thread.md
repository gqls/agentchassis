# HANDOFF — resume the FEATURE BUILDER thread

**Filed 2026-07-25** from the "fixloop feature builder 2" session.
**Supersedes `HANDOFF_2026-07-21_feature_builder_thread.md`**, which is now wrong
on the one thing that mattered: it says the implementer "has still never executed
once". **It has. It finished, and its pull request has been merged.**

Everything in this file was re-verified against the live cluster, the live DB and
GitHub on 2026-07-25 — not carried forward from the 07-21 file. Where a claim is
inference rather than measurement it is marked `[INFERRED]`.

## Read these first, in this order

1. This file, top to bottom.
2. `SUMMARY_feature_builder_2026-07-25.md` — plain-language read-out, 3 minutes.
3. `PLAN_feature_builder.md` — architecture as built + status table.
4. `NOTES_running_feature_builder.md` **turns 16–17** and the 2026-07-25 B4 entry.
5. `README_feature_builder_where_we_are.md` — owner's plain-prose log (bottom).
6. For the B4 run itself, the operating thread's own log is richer than ours:
   `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md`
   (2026-07-23 → 07-25) and `SUMMARY_2026-07-25_gauntlet_dead_cta.md`.

Repo rules override everything: `/CLAUDE.md` — **read it fresh from disk.**
Parent design: `DESIGN_feature_builder_and_council_gate.md` §1;
`DESIGN_stage_loop_delta2.md`; schema `SCHEMA_staged_plan_v1.md`.

## State in one paragraph

**The feature builder is PROVEN END TO END.** On 2026-07-25 the designer's
approved staged plan (`c379f7b7`) was executed by `feature-implementer` in a
single clean run: six gated stages, six commits, a passing derived test gate, and
one pull request — **PR #3, merged into `main` by the owner at 09:19:16Z**. The
milestone the last three handoffs were all pointed at (RUNBOOK **B4**) is CLOSED.
The build was NOT done by this thread: the `gauntlet_dead_cta` workstream drove
it, on its own target, and contributed the closure note. What remains here is
smaller and different in kind — a stale council trail, one open bug from the
shakeout, and the question of what the tool gets pointed at next.

## What actually happened (verified 2026-07-25)

**The designer.** Round 6 on the tools-api spec (work item
`9ed684bc-864a-4aa1-b17a-7ed061e08f2a`, corr `c379f7b7`) was **APPROVED after 3
council rounds** — the designer's first-ever convergence on a *revise* loop
rather than a first-pass approval. Rounds 1–5 each died on a real defect and each
produced a durable fix (see the shakeout ledger below).

**The implementer.** Round 8 (orch `af286d2c`, fired 08:37:19Z) ran the whole
plan: stages s1–s6 = scaffold → dockerfile+kustomize → JSON error helper + CORS →
rate limiting + input caps → rounds table + `/round` → position/defend via
`platform/aiservice`. Each stage committed separately on `feat/c379f7b7`, each
commit message carrying the plan correlation and the line *"Human review terminal
— do not merge without review."* `awaited_requests`: **8/8 processed, 0 expired**
— including all six `stage_commit` awaits, the exact class that killed rounds 6
and 7.

**The PR, and the gate.**
- `https://github.com/gqls/agentchassis/pull/3` — created 09:13:48Z,
  18 files, **+880 / −0**, six commits, branch `feat/c379f7b7`.
- **State: MERGED.** Merged 09:19:16Z by `gqls` into `main`; merge commit
  `c02d56b9a8ea542a79c653d5e2a171c4963131ae`. `gh api compare c02d56b9a...main`
  → `identical, ahead:0, behind:0` — i.e. that merge commit **is** the tip of
  `main`. The owner's one hard gate has been walked.
- Merged paths: `cmd/tools-api/main.go`, `internal/tools-api/{api,config,db,
  handlers,httperr,middleware,store}`, `build/docker/backend/tools-api.dockerfile`,
  `deployments/kustomize/services/tools-api/**`, and
  `docs/agent_docs/sql_for_agents/198_tools_api_gauntlet_rounds.sql`.

> **Note on the other thread's summary.** `SUMMARY_2026-07-25_gauntlet_dead_cta.md`
> says PR #3 "is open ... awaiting the owner's review". That was **true when it
> was committed** (09:17:06Z) — the merge landed two minutes later. It is not an
> error, it is a doc that aged in 120 seconds. Do not "correct" it; it is a
> point-in-time record. This file carries the later state.

## Traps this leaves behind — read before you touch anything

1. **The merged code is NOT in your working tree.** This repo's checkout is on
   branch `086_experience_loop`; there is **no local `main` branch**, and the
   local `origin/main` ref is stale (`998c0b312`). `cmd/tools-api` and
   `internal/tools-api` do not exist on disk. `git fetch` before you go looking
   for them, and do **not** conclude the merge didn't happen because a local grep
   is empty — that is the absence-without-search failure this workstream has
   already logged once.
2. **Migration 198 is UNAPPLIED, and probably not for `clients_db`.** It is not
   in `schema_migrations` (checked; the ledger has 199, 200, 201×2, 202×2 and no
   198). The gauntlet thread's deploy target pivoted to the **ISLAND VM**, so 198
   belongs to the island's own Postgres, not the cluster DB, and the PR's cluster
   kustomize manifests are **not** the deployment path. That call is the island
   session's — see `infra/island/RUNBOOK_island.md` and the gauntlet docs.
3. **Migration numbering has collided across threads.** `201_`, `202_`, `203_`,
   `205_` and `206_` each exist twice under different slugs from concurrent
   workstreams. The ledger keys on **filename**, so they coexist without error —
   but a bare number is now ambiguous in conversation. Re-check 198 is still free
   at apply time and quote slugs, not numbers.
4. **The tools-api work item is stale.** `9ed684bc-864a-4aa1-b17a-7ed061e08f2a`
   is still `status='needs_human_review'` although its PR is merged. It belongs
   to the gauntlet thread — flag it to them, don't close it from here.

## The B4 shakeout fix ledger (all of this came out of the first fire)

The first fire was expected to find defects and it found six. Every one is now a
durable record — this is the milestone's real yield, more than the PR itself.

| # | Defect | State |
|---|---|---|
| `bugs_closed/065` | `formatGeneratedGo` rejected the commit-entry shape its only caller sends | CLOSED, live |
| `bugs_open/066` | Spawned agent pods pin `agent_definitions.image_tag`; a chassis roll never reaches them | **OPEN** — see below |
| `bugs_closed/067` (+addendum) | Designer repropose/compose caps made every revise cycle fatal | CLOSED, live (migrations 201/202) |
| `bugs_open/071` | `agent-job-cleanup` deleted live `job.*` topics every 10 min — its "is anything running" guard label matched no pod, ever | Fix LIVE + **tick-proven**; residuals shipped `bc1f12718`; case still open |
| mig `199_implementer_module_path_rule` | Module path pinned for generated code | applied 07-24 |
| mig `200_implementer_plan_paths_are_law` | Rule-8 example was seeding a path deviation | applied 07-24 |

**071 is the important one.** Two implementer runs died on 07-24 in a way that
looked like implementer defects and were not: a housekeeping cron was deleting
the Kafka topics the running agents were replying on. The B4 run crossed three
cleanup ticks (08:50/09:00/09:10) and each one took the KEEP branch — that is the
fix's behavioural proof, not a green status.

**066 status, measured today:** all **173** active agent rows with a chassis
image are at `v1.0.1158` = the deployed tag, so the gap is not currently biting.
But the bug is *not* fixed: the `UPDATE agent_definitions SET image_tag` lives in
the `deploy-100-bootstrap-agents` target (`makefile:518`), **not** in
`deploy-agents`. `[INFERRED]` — whoever rolled 1158 happened to run a target that
syncs. **Re-run the census before any implementer fire**, don't assume:

> **CORRECTED 2026-07-27 (bugfix-066 thread): the `[INFERRED]` half is FALSE.** The UPDATE
> is in **both** targets — `deploy-agents` calls `update-agent-images-v2` at `makefile:1028`.
> So "whoever rolled 1158 happened to run a target that syncs" is not the explanation;
> the ordinary roll path syncs. The rows went four tags stale on 07-24 anyway, and the real
> reason is better: **a deploy-time sync is a property of one deploy PATH, not of the
> system** — `kubectl apply -k`, `kubectl set image` and `rollout undo` all move the cluster
> without it. Marking the claim `[INFERRED]` is what made it cheap to falsify — this is the
> marker earning its place, not a criticism of the note. 066 is now fixed in code
> (`c0d7c3a71`, INERT until a roll past v1.0.1174); the census below is superseded by
> `scripts/check-agent-image-drift.sh` — see the correction in `RUNBOOK_feature_builder.md`.
```sql
SELECT COALESCE(image_tag,'(null)') tag, count(*) FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND image_repository LIKE '%agent-chassis%' GROUP BY 1 ORDER BY 2 DESC;
```

## What is live right now (all pod/DB-verified 2026-07-25)

| Thing | Where | State |
|---|---|---|
| chassis image | deploy `agent-chassis`, pod `agent-chassis-54fff9df8b-966hj` | **v1.0.1158**; `feature_stage_route`=3, `formatGeneratedGo`=2 |
| `IMAGE_TAG` | `makefile:16` | `v1.0.1158` |
| `feature-designer` | agent_definitions | LIVE, active, `image_tag` v1.0.1158; council still **5 seats**; reviewer seats on sonnet-5 (`ee31c3632`); patches 016/017/018/022 + migrations 201/202 applied |
| `feature-implementer` | agent_definitions | LIVE, active, v1.0.1158 — **now proven** |
| `feature-implementer-orchestrator` | agent_definitions | LIVE, active, v1.0.1158 — proven |
| stage loop + 3 shared seams | `feature_stage_route_action.go` + `diagnose_{read_repo_files,prepare_fix_commit,build_gate}` | LIVE and **exercised end to end** |
| travelling action subjects | doc_plans/doc_notes `subject_type='action'` | LIVE (migration 184) |
| machine-built feature | `origin/main` | **MERGED** (PR #3, `c02d56b9a`) |

## The delta-2 council trail — UNCHANGED and still open

Correlation `5a65ec4c`. **Round 2 = REVISE (2026-07-21). Nothing has run since** —
verified today: zero orchestration rows with that `fix_correlation_id` after the
round-2 run. No `Council-Reviewed` trailer exists anywhere for it, correctly (the
trailer is earned by APPROVED only).

The round-3 shopping list still stands, and is still genuine design work:
1. Read `LoopAction` (`loop_actions.go`) and give the real reuse answer — could
   the stage loop have been a `loop` config, or does emit-stage-as-a-plan +
   branch/ref threading genuinely need a bespoke action? The gate caught me
   asserting "no loop-controller exists" without looking; a generic
   `loop`/`loop_complete`/`conditional_route` does exist (`registry.go:47/53/73`).
   Logged in `WRONG_CALLS.md`.
2. Decide `githubBranchExists` (`feature_stage_route_action.go:375`): add a read
   verb to the git-adapter (`internal/adapters/git/adapter.go:337-345` has
   commit/create_repo/delete_repo/create_branch/create_pull_request — no read
   verb) and route through it, or justify the direct read-only GET explicitly.
3. Cheap closes: migration 184 as its own declared edit; attach the registry +
   verb search **output**, not an assertion; name the owning pipeline on the
   config_change edit; show a pre-state needle-gate in PATCH_018.

**The argument for spending a round 3 is weaker now than it was on the 21st.**
The code the trail reviews has since run a full six-stage build in production and
opened a merged PR. The gate is advisory, its one high-severity find was fixed
back on the 18th (`9c94cc842`), and the surviving objections are design flags on
code that is now proven in use. Closing the trail is a tidiness and
review-coverage decision, not a risk decision. **Owner's call, one run's credits.**

## What is actually next

B1–B4 are all done. The open questions are new ones:

1. **Point the tool at a second target.** One successful build is not a capability
   — the run needed 6 designer rounds and 8 implementer fires, and most of the
   failures were environmental (071, 066) rather than design. A second build,
   chosen by us rather than inherited, is what tells us whether the shakeout
   fixes generalised. RUNBOOK B1's selection criteria still apply, especially
   "not being touched by another thread".
2. **Decide the delta-2 trail** (above) — round 3, or accept advisory-REVISE and
   record it as closed-unapproved.
3. **Hand 066 back to whoever owns deploys.** It is the one shakeout bug still
   able to bite silently, and its symptom is indistinguishable from an agent
   defect. That was the whole cost of rounds 6–7.
4. **Follow-up on the merged PR belongs to the island/gauntlet threads,** not
   here: image build, island deployment, migration 198 to the island DB,
   smoke-testing tools.apis.uk. Do not start it from this thread — check
   `scripts/who-owns.py` first.

## Do NOT do these

- Do **not** put a `Council-Reviewed: 5a65ec4c` trailer on anything — the verdict
  is REVISE. It is earned by an APPROVED round only.
- Do **not** apply the plan from run `8e837814` (the dead F1.2 pilot) — it would
  regress a working hand-fix. Item `db066cac` is closed.
- Do **not** fire `feature-implementer` directly — only via
  `feature-implementer-orchestrator`, or it dies with no read token.
- Do **not** re-apply `0NN_feature_designer.sql` wholesale — it would revert
  patches 016/017/018/022 and migrations 201/202. Diff against the LIVE row;
  edits are surgical `jsonb_set`.
- Do **not** deploy the tools-api PR using its own cluster kustomize manifests
  without checking with the island thread — the exposure decision moved.
- Do **not** spend a designer/council/implementer run without the owner's
  per-run go.

## Coordination notes

This repo and cluster are worked by many threads concurrently. Since the 07-21
handoff: the chassis rolled 1144 → 1158, the feature-builder's own milestone was
completed by a *different* workstream, four bugs were filed and two closed
against our machinery, and five migration numbers collided. Your session-start
`git status` is a snapshot — re-run it. Commit narrowly with explicit pathspecs.
Check `site_work_items` and `scripts/who-owns.py` before routing work at anything.
