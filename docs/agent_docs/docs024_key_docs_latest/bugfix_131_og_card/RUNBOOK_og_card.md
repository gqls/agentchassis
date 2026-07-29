# RUNBOOK — og-card lane (bugs_open/131 slug)

The commands that were hard to get right, with their gotchas attached. Created 2026-07-29
(session relojistas-5); the derive/verify recipes date from 07-28.

## Queue a brand-head derivation (card + favicon from the site's logo)

```sql
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec,
       handler_agent, status, created_by, priority, pipeline, item_key, triaged_at)
SELECT id, 'discovery', 'needs_brand_head_assets', 'medium', '<summary>',
       '{"mode": "brand_head"}'::jsonb, 'asset-deployer', 'triaged', '<who>', 70,
       'build', 'needs_brand_head_assets:og_card', now()
  FROM sites WHERE domain = '<domain>';
```
- **`status='triaged'` AND `pipeline='build'` are load-bearing** — default `detected` is never
  dispatched. One site per 120s tick. No dispatch within ~300s of a chassis pod restart.
- On the pre-`e9e345464` binary the favicon comes out squashed (non-proportional resize) and
  locked rows do NOT protect the git artefact — derive only after v1.0.1199+ is live.

## Verify a card (all four steps — the first three passed on WRONG cards)

```bash
img=$(curl -s https://<site>/ | grep -o 'property="og:image" content="[^"]*"' | sed 's/.*content="//;s/"//')
curl -s -o card.png -w "%{http_code}\n" "$img"
file card.png          # must say PNG 1200 x 630 — a 404 page saves happily as card.png
# then: Read card.png — LOOK at it. Dimensions and MIME are a status code, not the artefact.
```

## assets.url — the only safe form

Write **path-style HTTPS**: `https://s3.us-east-005.backblazeb2.com/<bucket>/<key>`
(signature/query optional and ignored — the chassis strips it and re-signs).
- **A bare `s3://bucket/key` BREAKS derivation**: `presignedURLToS3URI` parses the path as
  `/bucket/key`, so the real bucket (in the host position) is lost and the first key segment
  is eaten as the bucket → wrong key → `NoSuchKey`.
- A deployed web path (`/assets/images/logo.jpg`) also breaks it — that was idea.uk.
- Bucket: `personae-prod-uk001-images`; key convention `images/system/<yyyymmdd>/<uuid>.png`.

## Put exact bytes into S3 "through the chassis" (owner ruling — no creds in-session)

One-off Job + binaryData ConfigMap, mirroring the database-backup cronjob's upload pattern:
```bash
kubectl -n ai-persona-system create configmap <name> --from-file=logo.png=<local path>
kubectl apply -f ingest_job.yaml   # template: relojistas-5 scratchpad ingest_job.yaml, or
                                   # rebuild: alpine:3.20 + apk add aws-cli; secretKeyRef
                                   # B2_APPLICATION_KEY_ID / B2_APPLICATION_KEY from
                                   # personae-storage-secrets; aws s3 cp --endpoint-url
                                   # https://s3.us-east-005.backblazeb2.com --content-type image/png
kubectl -n ai-persona-system logs job/<name>   # expect the upload line + s3 ls line + INGEST-OK
kubectl -n ai-persona-system delete job/<name> configmap/<name>
```
The `aws s3 ls` inside the job is the in-bucket verification — byte count must match the local
file exactly.

## Generate a logo via the adapter WITHOUT installing anything

The full `image-build-handler` auto-deploys after store — it cannot be used when owner
sign-off must precede install. Publish straight to the adapter instead (no DB writes, no git):
```bash
# payload: {"headers":{"correlation_id":"<uuid>","client_id":"system","request_id":"<uuid>"},
#           "body":{"action":"generate_image","data":{"prompt":"...","negative_prompt":"...",
#                   "kind":"logo","provider_hint":"banana","aspect_ratio":"3:2"},
#                   "reply_to_topic":"<your scratch topic>"}}
# ONE LINE (kcat -P splits stdin on newlines); topics auto-create — probe first.
cat payload.json | kubectl -n kafka run -i --rm "kcat-$(date +%s)" --image=edenhill/kcat:1.7.1 \
  --restart=Never -- kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.image-generator.requests
# response arrives on reply_to_topic with image_uri (s3://) + image_url (7-day presigned):
kubectl -n kafka run -i --rm "kcat-c-$(date +%s)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t <your scratch topic> -o beginning -e -q
```
- `provider_hint:"banana"` pins gemini — it renders TEXT correctly; SDXL garbled idea.uk's
  original ("IBTA"). `client_id` becomes the S3 key's second segment (`images/system/…`).
- The adapter stores the provider's JPEG bytes under a `.png` key — `image.Decode` sniffs
  content, so derivation is unaffected. Do not "fix" the extension.
- Curl the presigned `image_url` and **Read the PNG** before any row is touched.

## Deploy a header/static asset to the live site (gqls/sites contents API)

```bash
SHA=$(gh api "repos/gqls/sites/contents/<domain>/<path>" --jq '.sha' 2>/dev/null || true)
python3 -c "import base64,json;print(json.dumps({'message':'<msg>','content':base64.b64encode(open('<local>','rb').read()).decode(),**({'sha':'$SHA'} if '$SHA' else {})}))" \
  | gh api -X PUT "repos/gqls/sites/contents/<domain>/<path>" --input -
```
- **A local error after the PUT does not mean the PUT failed** — a bad `--jq` output
  expression exits 1 AFTER the write lands; a 409 on retry means it already landed. Check
  `repos/gqls/sites/commits?path=…` before retrying.
- The "Deploy to B2" workflow syncs `b2://portfolio-sites/<domain>` and purges Cloudflare —
  but the live edge serves via an **intermediate origin that pulls from B2 on its own
  cadence** (nginx-style etag), so the URL can serve stale bytes long after a green run.
  Verify by size on the wire; poll, don't assume.

## Council verdict for this lane's code fix

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='bfd73f71-ad77-45b0-a1a2-433cc8dabc1e' AND kind='council_report' ORDER BY created_at;
```
A fleet roll KILLS an in-flight council round (orchestration stays EXECUTING_STEP forever) —
compare the audit trail's last change against chassis pod `startTime` before diagnosing a
hang, and resubmit with `RESUBMIT_CORR=<corr>` (same trail), not a fresh submission.
