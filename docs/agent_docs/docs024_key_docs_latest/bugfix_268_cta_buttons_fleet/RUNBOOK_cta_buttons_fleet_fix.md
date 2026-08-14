# RUNBOOK — bugfix_268_cta_buttons_fleet

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Coordination (rerun before routing work — every check below is a snapshot)

```bash
python3 scripts/who-owns.py 268
# who-owns is blind to uncommitted sessions — grep live transcripts too:
cd ~/.claude/projects/-home-ant-projects-agentchassis && \
  for f in $(ls -t *.jsonl | head -12); do \
    echo "$f hits=$(grep -c 'bugfix_268\|cta_buttons_fleet' "$f")"; done
# Gotcha: a hit can be just the SessionStart LANDMINES banner (bugs_open/ footprint
# matches 268's filename) — read the hit contexts before concluding a lane is on it.
```

## What is the chassis running (per SERVICE, at the artefact)

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image,STARTED:.status.startTime'
kubectl -n ai-persona-system logs <pod> --tail=5000 | grep -m1 'build provenance'
# Gotcha: `logs -l app=... | grep 'build provenance'` can match a GIANT unrelated
# line (agents log doc_notes payloads containing those words) and blow the output
# to MB — name the pod, and take the JSON line with git_commit in it.
git merge-base --is-ancestor <your-commit> <the stamp>   # did my fix ship?
# No orchestration dispatch within ~300s of a chassis pod (re)start.
```

## Fleet census (the §2 measurement — labels without URLs)

Full query: `bugs_open/268_HANDOFF_2026-08-12_...md` §2. Baseline 2026-08-12
~20:45Z: 216 components / 19 sites carry a label with no URL; 214 render zero anchors.

## Invariant diff (the one check of six that sees this damage)

```sql
SELECT p.name, pc.slot_name,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'href="','g')) AS links
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND p.status='active' ORDER BY p.name, pc.position;
```
Take it BEFORE and AFTER any regeneration, as a matched pair.

## webdesign.uk locks (expect 8; leave the site out of sweeps)

```sql
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND pc.lock_type='permanent' AND pc.slot_name IN ('hero','call-to-action');
```

## 090 diagnosis run

```bash
SLUG=<slug> RUNTIME_SITE=webdesign.uk SITE_ID=1fcfa4f3-ec80-4010-878b-b971cd46711f \
SEED_SCOPE='platform/orchestration/actions/plan_sections_action.go:carryStored,platform/orchestration/actions/save_page_sections_action.go,platform/orchestration/actions/rerender_page_sections_action.go' \
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh '<symptom>'
```
- Symptom text: no `"` or `\`; mechanism as a QUESTION; no counts; point at
  `page_component_history` windows 2026-08-12 16:37–17:23Z and 20:20–20:45Z;
  state that live webdesign.uk rows were repaired (17:23, ~20:44) and locked
  (20:46) so the loop does not read them as refutation.
- REF defaults to the current branch and must exist on origin — push first.
- Coverage check will hit our own open items → read findings, then FORCE=1.
- Budget ~30 min publish→run-start under load. Find the run by payload, not id:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';
```
- A `failed` work-item status is not always failed work (spawn→call handshake
  race) — check `diagnosis_artifacts` by correlation before believing it.
