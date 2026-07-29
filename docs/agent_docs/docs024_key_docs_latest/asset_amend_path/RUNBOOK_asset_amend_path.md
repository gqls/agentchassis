# RUNBOOK — the asset-amend path

Commands that had to be got right, with their gotchas attached. Update HERE when one changes.

## Amend an asset (the whole thing, one command)

```bash
./scripts/amend-asset.sh <domain> <asset_key> <file> [--purpose <p>] [--note "<why>"] [--dry-run]
# e.g.
./scripts/amend-asset.sh relojistas.com logo corrected-logo.png \
  --note "bugs_open/131: stored logo was a two-up spec sheet; cropped to the light variant"
```

- Stages the bytes (BYTEA) AND queues the triaged work item in one transaction.
- **Gotcha:** the dedup key is `amend_asset:<asset_key>` — a second amend for the same key
  while one is in flight fails on `idx_swi_dedup`. That is the dedup working; wait.
- **Gotcha:** a LOCKED assets row refuses by design. Clearing a lock is a separate,
  deliberate step — never scripted into the amend.
- Dispatch tick is ~120s; the ingest itself is seconds.

## Watch it

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT w.status, w.attempt_count, left(coalesce(w.error,'-'),80),
       w.result->'response'->'ingest_result' AS ingest_result
  FROM site_work_items w
 WHERE w.item_key='amend_asset:<asset_key>'
   AND w.site_id=(SELECT id FROM sites WHERE domain='<domain>')
 ORDER BY w.created_at DESC LIMIT 1;"
```

`ingest_result` carries `presigned_url`, `storage_path`, `width`/`height`, `bytes`.

## Verify — the part that is not optional

```bash
# 1. the object is really there (no S3 creds needed — the URL is the proof)
curl -s -o /tmp/amended.png -w "%{http_code} %{content_type}\n" "<presigned_url from ingest_result>"
# 2. LOOK AT IT. Read /tmp/amended.png as an image.
#    Every mechanical signal was green on 131's spec-sheet card. The eye is the check.
```

Staging-side state:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT status, left(coalesce(error,'-'),60), consumed_at, octet_length(content) AS bytes
  FROM asset_ingest_staging ORDER BY created_at DESC LIMIT 5;"
```

Assets-row state (the amendment history lives in `alterations`):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT asset_key, origin_type, origin_model, storage_path IS NOT NULL AS has_sp,
       jsonb_array_length(COALESCE(alterations,'[]'::jsonb)) AS amendments
  FROM assets WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') AND asset_key='<key>' AND status='active';"
```

## The failure branch (run it once per deploy of the action — verify-the-failing-branch)

```bash
# corrupt the sha deliberately: stage normally with --dry-run, edit the sha in the SQL, pipe it in,
# then queue the item by hand. Expect: staging status='failed', error mentions sha256, assets row untouched.
```

## Migrations (ordering is load-bearing)

- `265_asset_ingest_staging.sql` — the table. Inert; apply any time.
- `266_asset_deployer_ingest_mode.sql` — the live mode. **ONLY after the image roll**, proven by:

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "ingest_staged_asset"'
# ≥1 on BOTH pods, plus a positive control (grep a string you know is absent → 0)
```

Apply with the standard migration practice: dry-run first, scope the directory (`--apply` takes
EVERY pending file in scope).

## Downstream: getting the amended logo onto the SITE (header) and into the card

The ingest fixes the SOURCE OF TRUTH (S3 + assets row). The site header and og-card are
downstream artefacts, re-derived from it:

```sql
-- header file (writes /assets/images/logo.png via deploy_image_asset):
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec, handler_agent,
                             status, created_by, priority, pipeline, item_key, triaged_at)
SELECT id, 'operator', 'undeployed_asset', 'medium', 'Deploy amended logo to site',
       jsonb_build_object('s3_uri', '<s3_uri from ingest_result>', 'purpose', 'logo',
                          'asset_key', 'logo', 'domain', domain),
       'asset-deployer', 'triaged', 'operator', 70, 'build', 'deploy_amended:logo', now()
  FROM sites WHERE domain='<domain>';

-- favicon + og-card (brand_head derivation — only with chassis ≥ the roll carrying e9e345464,
-- the favicon-aspect + lock-honour fix):
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec, handler_agent,
                             status, created_by, priority, pipeline, item_key, triaged_at)
SELECT id, 'operator', 'needs_brand_head_assets', 'medium', 'Re-derive favicon + og-card from amended logo',
       '{"mode":"brand_head"}'::jsonb, 'asset-deployer', 'triaged', 'operator', 70, 'build',
       'needs_brand_head_assets:og_card', now()
  FROM sites WHERE domain='<domain>';
```

**Gotcha:** the header `<img>` src on older sites is `/assets/images/logo.jpg` (hero-purpose
deploy geometry) while a logo-purpose deploy writes `logo.png` — check what the served page
actually references before assuming the deploy landed in the right filename.
