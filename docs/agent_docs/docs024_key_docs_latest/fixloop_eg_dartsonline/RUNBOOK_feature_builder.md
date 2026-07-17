# RUNBOOK — feature builder (the owner's tasks)

*Human tasks only — agent work lives in `PLAN_feature_builder.md`. Tasks are
ordered; each names its verification. Last updated 2026-07-17.*

## Standing summary

Delta 1 + 2 code is committed (`4b3d50f4c`, `c19b5d097`) but INERT until an
image ships; the three seeds are DRAFT FILES, deliberately never executed by
the loop — applying them is yours, after the image (that ordering is the very
discipline the builder encodes). Then the F1.2 pilot runs the whole chain.

## A1 — build + deploy the chassis image ☐

```
make build-agent-chassis          # committed HEAD (must include c19b5d097)
# bump IMAGE_TAG first (makefile ~line 16); then push-/deploy- as usual
```

Verify against the RUNNING POD, never git, never the tag:

```
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "feature_stage_route"'   # expect >= 1
```

Mind the ~300s no-dispatch window after the pods (re)start.

## A2 — apply the three seeds (AFTER A1 verifies) ☐

In this order, each via psql into `postgres-clients-0` / `clients_db`
(each snapshots the prior row itself; renumber 0NN when filing):

1. `0NN_feature_designer.sql`
2. `0NN_feature_implementer.sql`
3. `0NN_feature_implementer_orchestrator.sql`

Verify: `SELECT type, version, updated_at FROM agent_definitions WHERE type
LIKE 'feature-%' AND is_active;` → three rows.

## A3 — decide the designer's council roster is current ☐

The designer seed mirrors the fix-proposer's v7 roster (4 seats incl.
reuse-agent) as of 2026-07-17. If the concept-register thread has changed the
live roster since, mirror those edits into the designer seed BEFORE applying
(same 4-edit shape as v6→v7; see the seed's header note).

## A4 — create + approve the F1.2 pilot spec ☐

One work item, then one approval update (no new status values — approval
lives in spec jsonb):

```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity,
  summary, spec, priority, handler_agent, status, created_by, item_key, max_attempts)
SELECT id, 'owner', 'maintenance', 'capability_gap', 'medium',
  'F1.2: make fix-implementer read ref/base per-run inputs (live-set to a stale branch)',
  '{"builder_needed": "feature-builder",
    "goal": "ref and base_branch become per-run inputs of the fix-implementer workflow, defaulting to main, so re-fires cannot silently read/branch from a stale base",
    "code_pointers": [
      {"path": "platform/orchestration/actions/diagnose_read_repo_files_action.go", "why": "ref resolution — ref_field exists since c19b5d097; the fix-implementer workflow does not use it yet"},
      {"path": "platform/orchestration/actions/diagnose_prepare_fix_commit_action.go", "why": "base_branch resolution — branch_field/commit_message_field exist since c19b5d097"},
      {"path": "docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_fix_implementer.sql", "why": "the live workflow whose v2 seed must pass input_data.ref/base through (read_current_files, prepare, create_branch)"}
    ]}'::jsonb,
  40, '', 'needs_human_review', 'feature-builder-runbook', 'f12-ref-input-pilot', 1
FROM sites WHERE name = 'system.internal' LIMIT 1
RETURNING id;   -- SAVE this work_item_id
```

Approve it (your explicit act, by name):

```sql
UPDATE site_work_items SET spec = spec ||
  '{"owner_approval": {"approved_by": "<your name>", "date": "2026-07-17"}}'
WHERE id = '<work_item_id>';
```

(Adjust the sites anchor if `system.internal` is named differently — check
`SELECT id, name FROM sites WHERE name ILIKE '%system%';` first. Schema
first: `\d site_work_items` before running, per house rule.)

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
