# RUNBOOK — feature builder (the owner's tasks)

*Human tasks only — agent work lives in `PLAN_feature_builder.md`. Tasks are
ordered; each names its verification. Last updated 2026-07-19.*

## Standing summary

> **UPDATED 2026-07-25 — A1–A5 *and* B1–B4 are now DONE.** This section used to
> end "the one thing left: the implementer has never run." It has run, and
> finished: plan `c379f7b7`, orch `af286d2c`, six gated stages, **PR #3 merged
> into `main` at 09:19:16Z**. The B-task instructions below are kept because they
> are the procedure for the NEXT build, not because anything in them is pending.
> Chassis is now **v1.0.1158** (pod-verified), not v1.0.1132.

Code is live and pod-verified, all three agents are seeded and active at
`image_tag` v1.0.1158, the designer has converged twice (unanimous on run 5, and
after three council rounds on the B4 plan), and the implementer has completed a
full six-stage build. The original F1.2 pilot is CLOSED as superseded (another
thread hand-fixed the target 60s before our run; its approved plan must NOT be
applied).

**What is left is in `PLAN_feature_builder.md` → "Next steps":** a second build
on a target we pick, the delta-2 council decision, and handing `bugs_open/066`
to whoever owns deploys. A1–A5 and B1–B4 are kept below for provenance and as
the runbook for run number two.

**Before ANY implementer fire, run the 066 census** — a chassis roll does not
reach spawned agent pods, and the symptom looks exactly like an agent defect:
```sql
SELECT COALESCE(image_tag,'(null)') tag, count(*) FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND image_repository LIKE '%agent-chassis%' GROUP BY 1 ORDER BY 2 DESC;
-- every row should be the tag now on the agent-chassis Deployment
```

> **CORRECTED 2026-07-27 (bugfix-066 thread) — use `scripts/check-agent-image-drift.sh`
> instead, and read this census differently.** 066 is fixed in code (`c0d7c3a71`) but
> **INERT until a chassis roll past v1.0.1174**, so:
> - **Until the roll:** the instruction above stands exactly as written.
> - **After the roll:** the spawn path takes its image from the RUNNING chassis pod
>   (`platform/orchestration/actions/agent_image.go`), so this census stops being an
>   exposure measure — a stale row is then a bookkeeping defect, not a stale pod, and
>   reading it as exposure will produce false alarms where it used to produce false comfort.
> - **The check that tells them apart** is `scripts/check-agent-image-drift.sh` (or
>   `make check-agent-image-drift`): it prints what the Deployment runs, what the rows say,
>   and what spawned pods are actually running, as three separate answers.
> - **Confirm which world you are in** before trusting either, with a pod-grep of a string
>   the fix created:
>   `strings /app/agent-chassis | grep -c "bugs_open/066: agent_definitions.image_tag trails"`
>
> Also corrected: this workstream's `[INFERRED]` note that the sync lives only in
> `deploy-100-bootstrap-agents` (`makefile:518`) and *not* in `deploy-agents`. It is in
> **both** — `deploy-agents` calls `update-agent-images-v2` at `makefile:1028`. The
> inference was correctly marked `[INFERRED]`, and marking it is what made it cheap to
> falsify; the rows went stale anyway, for a different and better reason (a sync is a
> property of one deploy *path*, not of the system). See `bugs_open/066` and `WRONG_CALLS.md`.

## A1 — build + deploy the chassis image ☑ DONE (via v1.0.1132)

v1.0.1131 was built from committed HEAD (`c19b5d097`) and pushed; a
concurrent thread's rollout of **v1.0.1132** (which includes the same
commits) landed first. Verified against the RUNNING POD 2026-07-17:
`strings /app/agent-chassis | grep -c feature_stage_route` → 3. Done.

Verify against the RUNNING POD, never git, never the tag:

```
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "feature_stage_route"'   # expect >= 1
```

Mind the ~300s no-dispatch window after the pods (re)start.

## A2 — apply the three seeds (AFTER A1 verifies) ☑ DONE

Applied 2026-07-17 (owner-approved in-session) to
`postgres-clients-0`/`clients_db`, in order, `INSERT 0 1` each. Verified:
three active rows; step counts 22/22/3; designer chain editquality →
bug_historian → reuse_agent → guidelines → guardian; 5 council
review_fields; NO root `ai_service` key on any (MDL-039 guard).

## A3 — decide the designer's council roster is current ☐

The designer seed mirrors the fix-proposer's **v8 roster (5 seats: +
reuse-agent, + guidelines)** as of 2026-07-17 evening. The roster moved twice
in one day — re-check it is still current at apply time; if it moved again,
mirror the edits BEFORE applying (same 4-edit shape as v6→v7→v8; see the
seed's header note).

## A4 — create + approve the F1.2 pilot spec ☑ DONE

Work item `db066cac-c647-44bf-a3ca-e04416405b28` created (site anchor
`System (internal)` `eac60db8-…`), then APPROVED by **aaa** 2026-07-17. The
designer's spec gate computes `approved`. Ready to fire.

## A5 — fire the designer; grade the plan ☑ DONE (5 runs)

Five fires on the F1.2 spec, each surfacing a real defect before the next:
runs 1–2 refused at validation (prompt steer; then an over-strict rule of our
own — seed-only plans need no image_deploy); run 3 sat the full council for
the first time; run 4 escalated correctly after 3 rounds; **run 5
(`8e837814`) was APPROVED unanimously, 5/5, zero objections.** Grading and
evidence in `NOTES_running_feature_builder.md` turns 7–12.

---

# B tasks — ALL DONE 2026-07-25; kept as the procedure for the next build

## B1 — choose a fresh pilot target ☑ DONE (tools-api, by the gauntlet thread)

The F1.2 pilot is spent. Pick a NEW capability that is:
- small, real, and genuinely wanted;
- of a known-good shape, so the plan can be graded rather than guessed at;
- **not being touched by another thread** — this bit is not optional. Check
  BOTH before choosing (the F1.2 collision cost a whole pilot):
```sql
SELECT id, item_type, status, left(summary,80) FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected')
  AND (summary ILIKE '%<target>%' OR spec::text ILIKE '%<target>%');
```
```
git log --since="3 days ago" --oneline -- <paths the target would touch>
```

Good shape for a first implementer exercise: a change with **at least one Go
file edit and one new file**, so the stage loop, the per-stage allowlist, the
build gate AND the derived test gate all get exercised. A seed-only target
would leave most of the machinery untested.

## B2 — write + approve the spec ☑ DONE (item `9ed684bc-864a-4aa1-b17a-7ed061e08f2a`)

```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity,
  summary, spec, priority, handler_agent, status, created_by, item_key, max_attempts)
VALUES ('eac60db8-b032-432b-b36d-76f37632045d', 'manual', 'maintenance',
  'capability_gap', 'medium', '<one-line summary>',
  '{"builder_needed": "feature-builder",
    "goal": "<what exists after this that did not before>",
    "code_pointers": [{"path": "platform/...", "why": "..."}],
    "deployment_reality": "<what is deployed vs merely committed — VERIFY BY POD GREP>",
    "seed_shape_rule": "if this edits an existing live agent row, the seed MUST be a surgical jsonb_set, never a whole-column replacement"}'::jsonb,
  40, '', 'needs_human_review', '<you>', '<unique-item-key>', 1)
RETURNING id;
```
(Site anchor `eac60db8…` is `System (internal)` — the name is not
`system.internal`.) Then approve it, by name:
```sql
UPDATE site_work_items SET spec = spec ||
  '{"owner_approval": {"approved_by": "<name>", "date": "YYYY-MM-DD"}}'
WHERE id = '<work_item_id>';
```

## B3 — fire the designer, grade, let the council approve ☑ DONE (round 6, corr `c379f7b7`, approved after 3 council rounds)

```
./0NN_TRIGGER_feature_designer_v1.sh <work_item_id>   # SAVE the FEATURE_CORR
```
Watch: `SELECT kind, metadata->>'decision' FROM diagnosis_artifacts WHERE
correlation_id='<corr>' ORDER BY created_at;` — proceed only on `approved`.

## B4 — THE FIRST IMPLEMENTER FIRE ☑ DONE 2026-07-25 ← the milestone, CLOSED

```
FEATURE_CORR=<uuid> ./0NN_TRIGGER_feature_implementer_v1.sh
```
**Via the orchestrator only** (the trigger already targets
`feature-implementer-orchestrator`) — fired directly it dies with no read
token, the fix loop's proven 2026-07-13 failure.

Expect: `feat/<short-corr>`, one commit per stage, green gate per stage, a
derived `go test` end gate, then ONE PR whose body carries the post-merge
checklist as a task list. A red gate = NO PR, branch + log left for
inspection: that is the hand-off working, not a failure of it.

> **RESULT 2026-07-25.** It took **8 fires**. Rounds 1–5 each hit a real defect
> (max_tokens refusal at s2, output-shape refusal at s1, a path deviation our own
> rule-8 example had seeded, designer caps that made every revise cycle fatal);
> rounds 6–7 died on `bugs_open/071` — an unrelated cleanup cron deleting the
> live `job.*` topics the agents were replying on — which looked exactly like an
> implementer defect and was not. Round 8 ran clean: six gated stage commits,
> `test_gate` PASS, **PR #3** (18 files, +880/−0), merged by the owner six
> minutes later. Full record in `gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md`.

Expect the same on the next fire: watch it closely and expect to find defects —
that has been the pattern of every first fire so far, and it is the point.

## Parked / recurring

- Credits: every designer/council/implementer run spends them — the go is
  yours per run.
- Roster drift: if the council gains/loses seats, the designer's
  `council_decide.review_fields` is now the ONLY list to update (patch 017
  made the reviser read the artifact).
- Delete stale `feat/*` branches before re-firing — the loop refuses them
  loudly (E4), by design.
- Delta 2's council-gate resubmission (`RESUBMIT_CORR=5a65ec4c-686c-40c7-813e-7c7fce03a779`)
  once the four open objections in the PLAN are answered.
- If pointer-curation proves too costly, the upgrade path is a dedicated-pod
  designer with repo read — a Go change; ask for it as a feature through this
  very tool.
