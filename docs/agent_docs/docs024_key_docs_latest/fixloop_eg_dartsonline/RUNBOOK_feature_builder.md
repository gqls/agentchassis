# RUNBOOK — feature builder (the owner's tasks)

*Human tasks only — agent work lives in `PLAN_feature_builder.md`. Tasks are
ordered; each names its verification. Last updated 2026-07-17.*

## Standing summary

Delta 1 + 2 code is committed (`4b3d50f4c`, `c19b5d097`) but INERT until an
image ships; the three seeds are DRAFT FILES, deliberately never executed by
the loop — applying them is yours, after the image (that ordering is the very
discipline the builder encodes). Then the F1.2 pilot runs the whole chain.

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

## A5 — fire the designer; grade before approving further ☐

```
./0NN_TRIGGER_feature_designer_v1.sh <work_item_id>    # SAVE FEATURE_CORR
```

Grade its staged plan against the hand-written reference
(`SCHEMA_staged_plan_v1.md` §6) BEFORE moving on: right stages, right files,
seed + checklist present, image before seed. The council must land
`approved`. Escalations/rejections park in `diagnosis_artifacts`
kind=escalation — read, decide, re-fire or hand-build.

## A6 — fire the implementer; review the PR ☐

```
FEATURE_CORR=<uuid> ./0NN_TRIGGER_feature_implementer_v1.sh
```

Expect: `feat/<short-corr>` with one commit per stage, green gates, ONE PR.
A red gate = no PR, branch + log left — that is the hand-off working, not a
failure of it. Review, merge (or don't), then walk the PR's checklist IN
ORDER: image → apply the v2 fix-implementer seed → verify → delete stale
`fix/*` branches.

## A7 — close the loop on the pilot ☐

After the checklist: fire the fix-implementer once with an explicit ref on a
known approved correlation and confirm it reads/branches from it. Then the
feature builder's first feature is LIVE — record the grade in
`NOTES_running_feature_builder.md` and update the PLAN's status table.

## Parked / recurring

- Credits: every designer/council/implementer run spends them — the go is
  yours per run.
- Roster drift: if the council gains/loses seats, A3 applies to any future
  designer re-seed.
- If pointer-curation (A4-style specs) proves too costly, the upgrade path is
  a dedicated-pod designer with repo read — a Go change; ask for it as a
  feature through this very tool.
