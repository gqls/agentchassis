# RUNBOOK — 417 logo text policy

Every command here was needed and had a gotcha. Gotchas are attached, not appended.

## Is the exemplar fix (669) actually applied? Ask the ROW, not a migration tracker
There is no `schema_migrations_agents` table — that guess costs a round trip. And
`schema_migrations` has no `version` column. Read the artefact instead:
```sql
SELECT CASE WHEN default_config::text LIKE '%no text outside the wordmark itself%'
              THEN 'LICENCE STILL PRESENT (669 NOT applied)'
            WHEN default_config::text LIKE '%a text-free mark%'
              THEN '669 APPLIED'
            ELSE 'NEITHER — drifted' END, updated_at
FROM agent_definitions
WHERE type='build-site-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## The census — count the LICENCE, never the prohibition, and never a literal
⚠ The exemplar phrase itself matches `no text`, so "does the prompt forbid text?" scores the
contradictory prompts as SAFE. **And counting the licence by literal is still a literal** — the
model rewords it ("other than" vs "outside"). The binding census is a human read of the concept.
The mechanical pre-filter that is *honest about being a floor*:
```sql
SELECT s.domain, spi.id, spi.created_at, left(spi.prompt,200)
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id=spi.plan_id
LEFT JOIN sites s ON s.id=sp.site_id
WHERE spi.kind='logo' AND sp.is_current
  AND spi.prompt ILIKE '%wordmark%' AND spi.prompt NOT LIKE '%migration 6%';
```
⚠ `site_plan_imagery` has **no `updated_at`** column — naming it errors the whole query.

## Post-roll: did the guard REACH the generation? (the only decisive check)
```sql
SELECT a.id, s.domain, a.created_at,
       (a.origin_prompt LIKE '%Render a text-free mark%') AS text_free,
       (a.origin_prompt LIKE '%the exact wordmark%')      AS wordmark
FROM assets a JOIN sites s ON s.id=a.site_id
WHERE a.asset_key='logo' AND a.created_at > '<the roll>'
ORDER BY a.created_at DESC;
```
**Neither true = the guard was UNREACHED on that path** — check `kind` arrival first (two legacy
parents map no `kind`; LANDMINES). ⚠ `assets` has **no `asset_role`** column; it is `asset_type`,
and the useful key is `asset_key='logo'`.

## Rehearse a migration before applying it
`sed 's/^COMMIT;$/ROLLBACK;/' <file> | psql` — proves the row count inside the real transaction
without committing. The `DO`/`RAISE EXCEPTION` guard is what makes the count assertable; a verify
block of bare `SELECT`s cannot stop the `COMMIT`.

## Council submission — the schema, which the dry run will teach you for free
`DRY_RUN=1 097_TRIGGER_council_review_v1.sh <file>` spends nothing. Types that bit me:
- `plan.edits[].operation` ∈ `modify|add|remove|config_change` — **`create` is invalid, use `add`**.
- `plan.risks` must be a **STRING** (prose block).
- `plan.grounded_in` must be an **ARRAY of strings** — do not "fix" it to match `risks`.
- an edit's `sketch` must not be **comment-only** (every non-blank line starting `//`/`--`/`#` is refused).
One command answers all of it at once:
`python3 -c "import json;d=json.load(open(F));print({k:type(v).__name__ for k,v in d['plan'].items()})"`

## Verify the change against committed HEAD, isolated from other sessions' dirty files
```
./scripts/verify-head-builds.sh --with <each file> --test
```
⚠ It prints `FAILED` and **exits 0**, so `&&` chaining will not catch it. HEAD is independently
red in ~23 places. **Run the bare control** (`./scripts/verify-head-builds.sh --test`) and diff
the FAIL sets — the claim you can actually make is "every failure in my set is in the control's".

## Fetch a generated asset's BYTES and LOOK at it — no storage key in your session
The census proves the instruction arrived; only the eye proves it was obeyed, and the eye needs
the file. Three routes, in order of preference.

**1. The site — TRY THIS FIRST, and do not use `publish_target` to decide whether to bother.**

> **CORRECTED 2026-09-02 19:45, same session, and I had written the wrong rule here four hours
> earlier.** The original text said *"Only sites with `publish_project` set are served"* and sent
> readers to the bucket. **That is false.** `publish_target`/`publish_project` govern **mirroring to
> a SECOND hostname** (`platform/publish/publisher.go`: "copies the tree under a second hostname
> prefix … served by the existing `*.ugg2.com` worker"); the site's own domain serves whenever its
> DNS points at the worker, which is independent of that column. **Measured 19:45: five domains —
> websitepromotion.co.uk, designblog.co.uk, seotools.co.uk, advertise.co.uk, gamedesign.uk — all
> have `publish_target` EMPTY and all return 200 on `/index.html` with a 404 invented-path
> control.** I generalised a serving rule from a column that does not control serving, on a sample
> of one domain that happened to be failing for an unrelated reason.

```bash
curl -sS -o out.png -w '%{http_code} %{size_download} %{content_type}\n' \
     "https://<domain>/assets/images/logo.png"
curl -sS -o /dev/null -w '%{http_code}\n' "https://<domain>/zzz-not-real-control.png"   # must NOT be 200
```
⚠ **A domain can be repointed mid-session.** advertise.co.uk returned 404 with a stranger's Drupal
markup at ~17:00 and served our own site at 19:45 — so "not ours" is a reading with a shelf life of
hours, not a property. **Re-probe before repeating it.** Always run the invented-path control (a
parked domain 200s every path).
⚠ Overriding `Host:` against the `*.ugg2.com` worker does not work — Cloudflare 403s it.

**2. Otherwise, from the bucket, THROUGH A POD** (owner 2026-08-23: never read a key into the
session). The B2 **native** API needs no SigV4, so BusyBox `wget` is enough — the S3 endpoint
would need signing and openssl, which the image does not have:
```bash
POD=$(kubectl -n ai-persona-system get pods -l app=image-generator-adapter \
        -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- sh -c '
BASIC=$(printf "%s:%s" "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY" | base64 | tr -d "\n")
R=$(wget -q -O - --header="Authorization: Basic $BASIC" \
      "https://api.backblazeb2.com/b2api/v2/b2_authorize_account" 2>/dev/null)
TOK=$(printf "%s" "$R" | tr "," "\n" | grep authorizationToken | sed "s/.*authorizationToken[\": ]*//; s/\".*//")
DL=$(printf "%s" "$R" | tr "," "\n" | grep -m1 downloadUrl   | sed "s/.*downloadUrl[\": ]*//; s/\".*//" | sed "s|\\\\/|/|g")
wget -q -O /tmp/a.bin --header="Authorization: $TOK" "$DL/file/$IMAGE_BUCKET/<key>"
base64 /tmp/a.bin | tr -d "\n"; rm -f /tmp/a.bin' 2>/dev/null | base64 -d > out.bin
```
`<key>` is `storage_path` with the `s3://<bucket>/` prefix stripped. Then `Read` the file.
- ⚠ **Use `/b2api/v2/`, not `v3`** — v3 nests `downloadUrl` under `apiInfo.storageApi` and the
  flat parse above silently yields an empty string, which then 401s and looks like a bad key.
- ⚠ **Always run an invented-object control in the same exec** — a wrong key and a private bucket
  both return failure, and only the control tells you the recipe works.
- ⚠ **Never print `$R`, `$TOK` or the key**; print `${#TOK}` if you need to prove it parsed.
- Clean up `/tmp` in the pod afterwards — it is a production pod.

**3. ⚠ THE EXTENSION LIES. Read the magic bytes before you trust the format.**
`dynamic_adapter.go:717` hard-codes `.png` in the key and `:726` hard-codes `image/png` as the
upload Content-Type, while `:492` DISCARDS the provider's real MIME into `_`. **12 of 12 logo
source objects sampled 2026-09-02 (spanning 2026-08-10 → 09-02) are JPEG** (`ffd8ffe0`), stored
under `.png` keys and served as `image/png`:
```bash
head -c 4 out.bin | od -An -tx1     # 89504e47 = PNG, ffd8ffe0 = JPEG
file out.bin                        # also gives dimensions and whether alpha is present
```
⚠ **ASSERT THE BYTE COUNT — the base64-through-`kubectl exec` pipe truncates silently.** Hit
2026-09-02: 776,498 of 1,081,560 bytes arrived, and `file` and PIL both still reported
`PNG 1024x1024 RGBA` correctly off the header — the loss only surfaced on *pixel* access. Echo the
size from inside the pod (`echo "POD_BYTES=$(wc -c < /tmp/a.bin)" >&2`) and compare with the local
file before believing any measurement over the pixels.
This is why `assets.mime_type` cannot be backfilled from the extension — see `bugs_open/433`.
