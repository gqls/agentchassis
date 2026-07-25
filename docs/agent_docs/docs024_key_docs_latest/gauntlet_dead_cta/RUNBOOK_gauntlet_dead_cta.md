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

## 5. tools-api build & deploy — ISLAND, not the cluster (as done 2026-07-25)

> The cluster kustomize steps that stood here are SUPERSEDED: the exposure home
> is the island VM (Route B1) — the PR's kustomize manifests are unused.
> Full as-built record: `infra/island/RUNBOOK_island.md` §ENGINE LANDED.

```bash
# build from the 086 branch (carries tools-api source + post-merge fixes);
# bump IMAGE_TAG in the makefile first, commit both:
make build-tools-api-ref
# ship WITHOUT registry creds on the island:
docker save docker.io/aqls/tools-api:<TAG> | gzip | \
  ssh root@toolsapisuk.vs.mythic-beasts.com 'gunzip | docker load'
# update image: tag in infra/island/docker-compose.yml, scp it to /opt/island/,
# then on the island: cd /opt/island && docker compose up -d tools-api
```
- **Migrations go to the ISLAND Postgres** (`docker exec -i island-postgres-1
  psql -U tools_api -d tools_api`), ledgered in its `island_migrations` table —
  NOT clients_db, NOT schema_migrations. The island also carries a minimal
  `sites` table (CORS lookup + FK target) seeded with real cluster site ids.
- **GOTCHA:** the island runs the code exactly as committed — a defect fixed
  in-repo is inert until you rebuild + save|load + `compose up -d` (same
  image-first rule as the cluster, different transport).
- **GOTCHA (Cloudflare):** origin 502 bodies are REPLACED by Cloudflare's own
  page — use 503 for "engine offline" so the JSON shape survives; verify error
  shapes from OUTSIDE, not just island-local.
- Smoke from outside:
  `curl -s -X POST https://tools.apis.uk/api/v1/tools/gauntlet/round -H 'Origin: https://vonc.com' -d '{}'`
  plus the matrix: denied origin → 403, missing round → 404, real-round
  /position with no key → 503, preflight OPTIONS → 204, `/anything` → 404.

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
