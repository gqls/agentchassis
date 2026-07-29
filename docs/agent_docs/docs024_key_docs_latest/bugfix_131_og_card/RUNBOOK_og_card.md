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

## Deploy a header/static asset to the live site — TWO ROUTES, ASK THE DB FIRST

```sql
SELECT domain, github_repo FROM sites WHERE domain = '<domain>';
```
**This query is not optional.** `github_repo='vm-sites'` → the site is served by nginx on a
VM from `gqls/vm-sites` (`deploy-to-vm.yml`). Empty → the B2 route, `gqls/sites` +
`deploy-to-b2.yml` + a Cloudflare Worker. **Both repos contain a `<domain>/` folder for some
VM sites**, so writing to the wrong one succeeds, runs a green workflow, and changes nothing
a visitor sees. Measured 2026-07-29: relojistas.com and idea.uk are `vm-sites`;
gaswholesalers.com and leopardessconsulting.co.uk are empty (B2). Substitute the right repo
below.

```bash
SHA=$(gh api "repos/gqls/sites/contents/<domain>/<path>" --jq '.sha' 2>/dev/null || true)
python3 -c "import base64,json;print(json.dumps({'message':'<msg>','content':base64.b64encode(open('<local>','rb').read()).decode(),**({'sha':'$SHA'} if '$SHA' else {})}))" \
  | gh api -X PUT "repos/gqls/sites/contents/<domain>/<path>" --input -
```
- **A local error after the PUT does not mean the PUT failed** — a bad `--jq` output
  expression exits 1 AFTER the write lands; a 409 on retry means it already landed. Check
  `repos/gqls/sites/commits?path=…` before retrying.
- **A green workflow run is not a live change if you picked the wrong repo.** The B2 workflow
  syncs `b2://portfolio-sites/<domain>` and purges Cloudflare, and reports success whether or
  not anything serves from there. If the wire still shows old bytes, re-check `github_repo`
  before theorising about caches — that mistake cost this lane an hour and a wrong inference
  about a non-existent lagging origin (NOTES entry (6), refuted).
- Verify by byte size on the wire with a cache-buster; poll, don't assume.

## Council verdict for this lane's code fix

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='bfd73f71-ad77-45b0-a1a2-433cc8dabc1e' AND kind='council_report' ORDER BY created_at;
```
A fleet roll KILLS an in-flight council round (orchestration stays EXECUTING_STEP, later
FAILED) — compare the audit trail's last change against chassis pod `startTime` before
diagnosing a hang, and resubmit with `RESUBMIT_CORR=<corr>` (same trail), not a fresh
submission.

## Before rolling the chassis: who is mid-council?

```sql
SELECT left(orchestration_id::text,8) AS orch,
       collected_data->'input_data'->>'submitter' AS submitter,
       current_step, updated_at
  FROM orchestration_states
 WHERE status='EXECUTING_STEP'
   AND (current_step LIKE 'review%' OR current_step LIKE 'gate%' OR current_step LIKE 'council%')
   AND updated_at > now() - interval '15 min'
 ORDER BY updated_at DESC;
```
- **The `submitter` field names the lane you would be damaging** — it comes from the
  submission JSON, so it is how you find out whose round it is without guessing.
- **Run this ADJACENT to the roll, not minutes before it.** A round can start in the gap: this
  lane's own wait returned "clear", and 26 seconds later a fresh round was at `review_honesty`.
  A stale all-clear is not an all-clear.
- **`updated_at > now() - interval '<n> min'` is doing real work here** — rounds killed by a
  previous roll sit at EXECUTING_STEP indefinitely, so without the recency bound you wait
  forever on corpses. Distinguish: a live round's `updated_at` advances between checks.
- **Do NOT count rounds with `WHERE current_step LIKE 'review%'` and call it a census of
  council traffic.** `current_step` moves as the round progresses, so a finished round drops
  out of the filter — the query answers "what is at a review step now", never "how many ran".
