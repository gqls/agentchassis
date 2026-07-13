# Runbook — apply & verify the differentiators / item-fields fix

Companion to `plan_pcw_item_fields_fix.md` (the why/what). This is the command set, keyed to
the plan's numbered steps. British English; commands assume `kubectl` context is already
pointed at the right cluster.

## Assumptions / placeholders (confirm before running)

- `<WRITER_DEPLOY>` / `<WRITER_POD>` — the page-content-writer workload name. Discover it in
  step 0; it is not hard-coded here because it varies.
- The **rebuild trigger** (step 3). Two options are given; confirm which your orchestrator
  uses — the DB flag, or an explicit message/CLI/API kick (the same path that produced the
  original ~17-minute build).
- `<DEPLOYED_URL>` — the public/preview URL the B2 deploy serves for idea.uk. The examples use
  `https://idea.uk`; substitute if you verify against a preview bucket/CDN URL instead.
- Whether your CI runs migrations on deploy, or you apply them by hand (step 2). The commands
  apply by hand.
- Known facts used below: namespace `ai-persona-system` (DB + agents), `kafka`; DB pod
  `postgres-clients-0`, db `clients_db`, user `clients_user`; idea.uk `site_id`
  `97ed2f64-65ca-4b67-8a98-dfd8195a0d3a`, page `url = '/index.html'`; Kafka cluster
  `personae-kafka-cluster`.

## 0. Pre-flight / context

```bash
# Pods in both namespaces
kubectl -n ai-persona-system get pods
kubectl -n kafka get pods

# Find the page-content-writer workload
kubectl -n ai-persona-system get deploy | grep -i -E 'writer|content'
kubectl -n ai-persona-system get pods   | grep -i -E 'writer|content'

# Interactive psql shell
kubectl exec -it -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db
```

Convenience function for one-off queries / heredocs (use in this shell session):

```bash
psqlc() { kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 "$@"; }
# e.g. psqlc -c "select now();"  or  psqlc <<'SQL' ... SQL
```

## 1. Repo / deploy (plan §1)

```bash
# From repo root, after dropping the two patched files into place:
gofmt -w path/to/plan_sections_action.go path/to/v3_site_actions.go
go build ./...

# Commit + push (triggers GitHub Actions → Backblaze B2)
git add path/to/plan_sections_action.go \
        path/to/v3_site_actions.go \
        path/to/migrations/019_pcw_prompt_item_fields.sql \
        path/to/migrations/019_pcw_prompt_item_fields_down.sql
git commit -m "page-content-writer: declare array item fields in prompt + render-time reconciler"
git push origin <branch>

# Watch the Action (GitHub CLI)
gh run watch --exit-status
```

Confirm the new image is the one running:

```bash
kubectl -n ai-persona-system get deploy -o wide | grep -i -E 'writer|content'   # check IMAGE/tag
kubectl -n ai-persona-system get pod <WRITER_POD> -o jsonpath='{.spec.containers[*].image}{"\n"}'

# If the pipeline doesn't auto-restart on a new tag:
kubectl -n ai-persona-system rollout restart deploy/<WRITER_DEPLOY>
kubectl -n ai-persona-system rollout status  deploy/<WRITER_DEPLOY>
```

## 2. Database — apply the migration (plan §2)

```bash
# Apply via stdin (dollar-quoted DO block pipes fine)
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < 019_pcw_prompt_item_fields.sql
# Expect: NOTICE "prompt patched"  (or "already applied, skipping" on a re-run)
```

Verify both markers present:

```bash
psqlc <<'SQL'
SELECT
  position('{{if .item_fields}}'   IN (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')) AS wtw_marker,
  position('{{if $f.item_fields}}' IN (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')) AS out_marker
FROM agent_definitions WHERE type = 'page-content-writer';
SQL
# both > 0
```

Eyeball the rendered prompt blocks:

```bash
psqlc -At <<'SQL'
SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
FROM agent_definitions WHERE type = 'page-content-writer';
SQL
```

## 3. Regenerate idea.uk index (plan §3)

Snapshot the current (broken) keys first, to diff against:

```bash
psqlc <<'SQL'
SELECT pc.content_data -> 'features' -> 0 AS first_item_before
FROM page_components pc
JOIN pages p           ON p.id  = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = '97ed2f64-65ca-4b67-8a98-dfd8195a0d3a'
  AND p.url = '/index.html' AND cc."function" = 'differentiators';
SQL
# expect title/body
```

Trigger the rebuild — **Option A** (DB flag; if your scheduler watches `idx_pages_needs_build`):

```bash
psqlc <<'SQL'
UPDATE pages SET build_status = 'needs_rebuild', updated_at = now()
WHERE site_id = '97ed2f64-65ca-4b67-8a98-dfd8195a0d3a' AND url = '/index.html';
SQL
```

**Option B** — if regeneration is driven by a message/CLI/API, use your existing trigger for
this page (the same one behind the original build). Confirm the entry point before relying on
Option A alone. If you go via Kafka and need to inspect topics on the Strimzi cluster:

```bash
kubectl exec -i -n kafka personae-kafka-cluster-combined-pool-prod-0 -- \
  bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```

Watch it build (logs in §4). When done, confirm the stored keys flipped:

```bash
psqlc <<'SQL'
SELECT pc.build_status,
       jsonb_object_keys(pc.content_data -> 'features' -> 0) AS item_keys,
       pc.content_data -> 'features' -> 0 AS first_item_after
FROM page_components pc
JOIN pages p           ON p.id  = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = '97ed2f64-65ca-4b67-8a98-dfd8195a0d3a'
  AND p.url = '/index.html' AND cc."function" = 'differentiators';
SQL
# expect keys name / description
```

Check the deployed HTML:

```bash
curl -s https://idea.uk/index.html | grep -A3 'class="differentiator-item"' | head -40
# expect non-empty <h3>/<p>
```

## 4. Logs (plan §4)

```bash
WRITER=$(kubectl -n ai-persona-system get pods -o name | grep -i -E 'writer|content' | head -1)
echo "$WRITER"

# Follow during the rebuild
kubectl -n ai-persona-system logs -f "$WRITER"

# Reconciler signal
kubectl -n ai-persona-system logs "$WRITER" --since=30m | grep -i 'reconcileGeneratedItemKeys'

# Narrow to this section
kubectl -n ai-persona-system logs "$WRITER" --since=30m | grep -i -E 'differentiators|current_section|features'
```

If there are multiple writer replicas, sweep them:

```bash
for p in $(kubectl -n ai-persona-system get pods -o name | grep -i -E 'writer|content'); do
  echo "== $p =="
  kubectl -n ai-persona-system logs "$p" --since=30m | grep -i 'reconcileGeneratedItemKeys'
done
```

Read:
- WARN `remapped`/`normalised` — model still drifted, caught by the net.
- ERROR `unrecoverable` — an item field with no synonym; investigate that component/field.
- no reconcile lines for differentiators — the prompt change alone produced the right keys.

(No `logger.Debug` is used, so these surface.)

## 5. Decision toggle (plan §5)

Non-fatal (current) vs fatal is a code choice, not a command: in
`reconcileGeneratedItemKeys` (v3_site_actions.go), to hard-fail, return an error from the
`!remapped` branch and propagate it in `RenderComponentAction`. Flag it and I'll patch.

## Rollback

**Go** — redeploy the previous image:

```bash
kubectl -n ai-persona-system rollout undo deploy/<WRITER_DEPLOY>
kubectl -n ai-persona-system rollout status deploy/<WRITER_DEPLOY>
```

**Prompt** — run the paired down-migration (idempotent; skips if already at the old form):

```bash
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < 019_pcw_prompt_item_fields_down.sql
# Expect: NOTICE "prompt reverted"
```

Note: rolling back the prompt does not unwind already-regenerated `content_data`. With the Go
reconciler still deployed, mislabelled keys are repaired at render time regardless, so a prompt-
only rollback is low-risk; a full rollback also redeploys the previous image.

## Follow-on (not this fix)

- **services-grid** — same schema; benefits automatically. Spot-check when first built:
  ```bash
  psqlc <<'SQL'
  SELECT cc."function", jsonb_object_keys(pc.content_data -> 'features' -> 0) AS item_keys
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE cc."function" = 'services-grid' LIMIT 5;
  SQL
  ```
- **info-card-grid** — confirm the `<no value>` baked-in template before any content work:
  ```bash
  psqlc -At <<'SQL'
  SELECT html_template FROM content_components WHERE "function" = 'info-card-grid';
  SQL
  ```
- **idea.uk parked gaps** (handoff): empty hero + CTA buttons, contact form posting to
  `#contact`, thin nav/footer — main-chat items, after this lands.
