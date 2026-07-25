# RUNBOOK — gauntlet_dead_cta (every command that was hard to get right, with its gotcha)

## 1. Fire the feature-designer / implementer

```bash
# designer (work item is capability_gap:tools-api-gauntlet-debate, id 9ed684bc-…)
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh 9ed684bc-864a-4aa1-b17a-7ed061e08f2a
# implementer — ONLY via its orchestrator trigger, ONLY after council 'approved':
FEATURE_CORR=<approved-corr> ./docs/.../0NN_TRIGGER_feature_implementer_v1.sh
```
- **GOTCHA (fatal, repeated):** fire ONCE and wait. Ingest latency under load is
  MINUTES (observed 9 min). A refire creates a second implementer that races branch
  creation; both mutually die on E4. Treat a missing orchestration row as QUEUED for
  ≥10 min. Patient watcher: `scratchpad/fire_impl6.sh` shape.
- **GOTCHA:** E4 hard-refuses if `feat/<corr8>` exists — `git push origin --delete feat/<corr8>` first.
- **GOTCHA (bugs_open/071, fixed 2026-07-25 but know the signature):** a
  stage_commit await that expires while the COMMIT LANDED on the branch =
  produced-but-never-consumed response. Killer was `agent-job-cleanup` deleting
  live `job.*` topics every 10 min (guard label matched zero pods always). If it
  recurs: `kubectl -n ai-persona-system logs job/agent-job-cleanup-<tick>` must
  say "Live spawned workload … keeping", NOT "No running spawned pods".
- **GOTCHA (bugs_open/066):** the implementer runs in a SPAWNED pod using
  `agent_definitions.image_tag` — NOT the chassis deployment's image. After any
  chassis roll that the implementer needs:
  `UPDATE agent_definitions SET image_tag='<tag>' WHERE type IN ('feature-implementer','feature-implementer-orchestrator') …` (snapshot first).

## 2. Watch a run (never trust the wrapper; UTC/BST trap)

```sql
-- the implementer's OWN orchestration (wrapper completes early/independently):
SELECT orchestration_id, status, current_step FROM orchestration_states
WHERE owner_agent_type='feature-implementer' ORDER BY created_at DESC LIMIT 1;
-- designer verdicts by corr:
SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<CORR>' ORDER BY created_at;
```
- DB times are UTC; local is BST (+1). A "10:45" orch at your 11:45 is the same moment.
- Refusal reason: `collected_data->>'__step_error'` on the agent's orchestration.
- Gate log: `collected_data->'stage_gate'->>'log'` (tail it; PASS/FAIL at the end).

## 3. Council gate submission (platform Go changes — the norm)

```bash
./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>   # save SUBMISSION_CORR
RESUBMIT_CORR=<corr> ./docs/.../097_TRIGGER_council_review_v1.sh <same.json>  # same trail
```
- **GOTCHA:** `complete_invalid` with no council_report = an infra death (e.g.
  Anthropic endpoint timeout at a seat), NOT a judgement. Resubmit on the same trail.
- Verdict: doc_notes (categories ? 'council-gate') or diagnosis_artifacts by corr.

## 4. Owned-page content delivery (the P4 front-end path, proven 07-22)

- Template/js/schema edits: dollar-quoted UPDATE on `content_components` (backup first).
- Deliver: section-editor `content_edit` via the 086-style DIRECT orchestrator envelope
  — `scripts/deliver_gauntlet_section_edit.sh`. The bare 049b `action=orchestrate`
  envelope has silently failed to ingest here; don't use it.
- **GOTCHA:** `apply_section_edit` does NOT republish `js_content`. Follow with the
  assemble-only rerender: `scripts/republish_gauntlet_js.sh`.
- **GOTCHA:** section-editor leaves `pc.build_status='approved'` — set back to
  `'deployed'` (an assemble path may drop non-deployed components).
- Verify live cache-busted, matching strings the change CREATED (never generic CSS).

## 5. tools-api build & deploy (after the PR merges — next stages)

```bash
# NO makefile edit needed — pattern rule covers any service with a dockerfile:
make build-tools-api-ref            # builds committed HEAD via build/docker/backend/tools-api.dockerfile
docker push docker.io/aqls/tools-api:$(IMAGE_TAG)   # confirm target name from the kustomize base
kubectl apply -k deployments/kustomize/services/tools-api/overlays/production
kubectl -n ai-persona-system rollout status deployment/tools-api
```
- **Migration number:** the PR carries `198_tools_api_gauntlet_rounds.sql` — RE-CHECK
  198 is still free in ledger+dir at APPLY time (renumber the applied copy if taken;
  ledger the actual filename). Apply AFTER the image is live (image-first-then-seed).
- Smoke (in-cluster first, then via bastion once P3 lands):
  `kubectl -n ai-persona-system run curl-smoke --rm -i --image=curlimages/curl -- \
   curl -s -X POST http://tools-api:<port>/api/v1/tools/gauntlet/round -H 'Origin: https://vonc.com' -d '{}'`

## 6. Experience re-plan re-fire (ONLY after the API answers a smoke POST)

```bash
./docs/agent_docs/sql_for_agents/092_TRIGGER_experience_plan.sh vonc.com vonc-spark-game "the Spark daily-provocation game"
```
- FIRST: carry the liveness evidence into the compose decisions block (the 197
  channel) — a small migration replacing the D1-REVISED text's API description with
  "LIVE, verified <date>: <curl output snippet>" so the feasibility seat can ground it.
- Accept only `approved` + `abstained:0` + `reviewers:5`; the current `is_current`
  plan is REJECTED-do-not-build until then.

## 7. Bastion / exposure (P3 — owner tasks then mine)

See `infra/README_bastion_exposure.md`. Owner: subdomain on apis.uk, bastion VM,
WireGuard peering. Mine: cloudflared + Caddyfile (fill <SUB>/<WG_*>/<PORT>),
NetworkPolicy after the service exists; smoke from outside + a denied-origin CORS check.
