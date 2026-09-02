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

## A site has a logo asset but the header still shows TEXT — decide it, then fix it
The header emits an image only when `HeaderConfig.LogoURL != ""`
(`multipage_actions.go:1635`); otherwise it falls through to `<span class="logo-text">`.
`LogoURL` is resolved in `render_site_components_action.go:513-528` by a join that needs **three**
things at once — a **current** plan, a `site_plan_imagery` row with `scope='site' AND kind='logo'`,
and an `assets` row whose `asset_key` equals that row's `key` **and** `status='active'`.

**Run this BEFORE dispatching anything — it says whether a re-render can possibly help:**
```sql
SELECT s.domain,
       (SELECT count(*) FROM site_plan_imagery spi
          JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
         WHERE sp.site_id=s.id AND spi.scope='site' AND spi.kind='logo') AS plan_logo_rows,
       (SELECT a.asset_key FROM site_plan_imagery spi
          JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current=true
          JOIN assets a ON a.site_id=sp.site_id AND a.asset_key=spi.key AND a.status='active'
         WHERE sp.site_id=s.id AND spi.scope='site' AND spi.kind='logo'
         ORDER BY spi.ordering LIMIT 1) AS resolves_to,
       coalesce(nullif(s.logo_url,''),'(empty)') AS legacy_fallback
FROM sites s WHERE s.domain = '<domain>';
```
`resolves_to = 'logo'` ⇒ a chrome re-render will produce an `<img>`. **`resolves_to` NULL with
`plan_logo_rows = 0` ⇒ a re-render changes NOTHING** — the site owns a logo asset the current plan
never asked for, and the fix is a plan row, not a render. Measured 2026-09-02: of **34** sites with
an active logo asset, **29** headers show the image and **5** do not — and of those 5, two
(ai-agent-orchestration.com, cookly.uk) have zero plan rows AND an empty legacy `sites.logo_url`,
so they are unreachable by re-render.

**Two more pre-flight checks, both of which stopped a dispatch on 2026-09-02:**
```sql
-- 1. LOCKED chrome is refused by any forced render, by design (bugs_open/069).
SELECT slot_name, locked_at, lock_type FROM site_components sc JOIN sites s ON s.id=sc.site_id
 WHERE s.domain='<domain>' AND sc.slot_name IN ('header','footer','head');
-- loanandmortgagecalculator.co.uk: all three PERMANENT since 2026-08-05. Do not force; ask.
-- 2. Open work on the site (CLAUDE.md's dispatch rule).
SELECT item_type, status, count(*) FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE s.domain='<domain>' AND w.status NOT IN ('complete','cancelled','rejected')
 GROUP BY 1,2 ORDER BY 1;
```

**Dispatch — `rerender-chrome` (seed 351), the narrow tool.** Two steps, no LLM, no page fan-out;
it renders header/footer/head into `site_components` and stamps the digest. Use the publish lib so
the receipt is asserted (`kcat -P` exits 0 having sent nothing):
```bash
source scripts/kafka-publish-lib.sh
CORR=$(cat /proc/sys/kernel/random/uuid)
PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"rerender-chrome"},"input_data":{"site_id":"%s","domain":"%s"}}' "$SITE_ID" "$DOMAIN")
CORR=$(cat /proc/sys/kernel/random/uuid); REQ=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid); MSG=$(cat /proc/sys/kernel/random/uuid)
kafka_publish_checked --topic system.agent.generic.requests --payload "$PAYLOAD" --correlation "$CORR" \
  --header "request_id=$REQ" --header "message_id=$MSG" --header "orchestration_id=$ORCH" \
  --header "orchestration_name=rerender-chrome-<domain>-$(date +%H%M%S)" \
  --header "step_name=start" --header "client_id=system" \
  --header "message_type=request" --header "action=orchestrate" \
  --header "from_agent_type=user" --header "from_agent_id=cli" \
  --header "responses_topic=system.generic.responses"
```
> **⚠ SEND THE FULL HEADER SET. A partial one is refused with NO trace at all, and I published a
> partial recipe here before finding that out.** Measured 2026-09-02, same payload and topic each
> time: `action`+`message_type`+`from_agent_*`+`request_id`+`responses_topic` → **no row after 50
> minutes**; the same **plus `client_id`** → **still no row**; the same plus **`message_id`,
> `orchestration_id`, `orchestration_name`, `step_name=start`** → **row in ~10 s, COMPLETED**.
> ⚠ I added those last four together and did **not** isolate which is load-bearing `[UNMEASURED]`
> — `orchestration_id` is the strongest suspect (NOT NULL primary key of `orchestration_states`,
> minted by the caller), but do not quote that as established. `client_id` alone is **not**
> sufficient. The canonical twelve live in working code, not in a doc:
> `platform/orchestration/actions/dispatch_verifiers.go:154-168`.
> Take `client_id` from the target's own history:
> ```sql
> SELECT client_id, count(*) FROM orchestration_states
>  WHERE collected_data->'input_data'->>'domain' = '<domain>' GROUP BY 1 ORDER BY 2 DESC;
> ```

⚠ **Latency vs DROP — and the standing advice points the wrong way when it is a drop.** The
estate's rule is *"a missing `orchestration_states` row is latency, not a drop — do not re-fire"*,
which is right for a queued message and **wrong for a refused one**, where waiting is infinite.
(For this agent the 25–36-minute framing is itself wrong: a well-formed dispatch lands in **~10
seconds** — measured three times, plus another lane's `af0857d2` which completed in one second.)
Tell them apart with a demand control before you decide:
```sql
-- Are OTHER top-level dispatches landing? These come straight off the topic.
SELECT count(*), max(created_at) FROM orchestration_states
 WHERE parent_orchestration_id IS NULL AND created_at > now() - interval '60 minutes';
```
**154 in the hour my two were missing** ⇒ the consumer is healthy, so mine were never accepted and
a re-fire duplicates nothing. A genuinely queued message sits behind a *stalled* consumer, and that
control reads zero or stale.

**⚠ THE HALF THAT IS EASY TO MISS: `rerender-chrome` does NOT touch deployed pages.** Its own
`complete` step says so — *"No page assembly, no deploys — served pages are untouched until their
own next rerender."* So the header in `site_components` gains its `<img>` and **the live site does
not change.** Propagate with `page-rerender` in **assemble mode**, and mind the two traps recorded
by the leopardess lane: the assemble branch needs **`page_id`, not `page_name`** (`page_name` alone
errors `page_id not found in input`, 29/29 pages), and it must carry **NO `spec.reason`** — sending
`reason=section_data_resolved` runs `rerender_page_sections`, whose pre-check escalates the whole
page to the content-writer on any missing `source:"llm"` field.

**Verify at the artefact, in this order** — the middle one is the one people skip:
```sql
SELECT slot_name, updated_at, length(rendered_html) AS len,
       (rendered_html LIKE '%logo-img%') AS has_img,
       (rendered_html_digest = md5(rendered_html)) AS digest_ok
FROM site_components sc JOIN sites s ON s.id=sc.site_id WHERE s.domain='<domain>';
```
then `curl -s https://<domain>/index.html | grep -o 'logo-img\|logo-text'` for the SERVED page.
**The component row and the served page are different facts** and the whole point of the
propagation step is the gap between them.

### The 5 text-header sites, resolved `[MEASURED 2026-09-02 20:20Z]` — only ONE was re-renderable

A re-render fixes exactly one shape of this. Establish which shape you have **before** dispatching.

| site | why the header shows text | does a re-render fix it? |
|---|---|---|
| **websitepromotion.co.uk** | header rendered 17:30Z with `render_inputs.plan_logo=""`, 30 min BEFORE the logo existed; the 18:01Z **page** re-render re-used the stored header | **YES — done**, header 2362→2633 B, `logo-img` present, digest ok |
| **webdesign.co.uk** | its header is a **bespoke component** (`webdesign.co.uk Site Header`, `ad6033ae`), and that template contains **no `logo_url`, no `logo-img`, no `logo-text` at all** — 4,133 chars with no logo slot | **NO.** Re-rendered 20:16Z: ran cleanly, `has_img` still false, length unchanged at 4142. Needs a template change or a switch to `header-theme-chrome` — a design decision |
| **ai-agent-orchestration.com** | zero `site_plan_imagery` rows for `scope=site/kind=logo` on the current plan, and `sites.logo_url` empty | **NO** — nothing for the renderer to resolve; needs a plan row |
| **cookly.uk** | same as above | **NO** |
| **loanandmortgagecalculator.co.uk** | all three chrome slots **permanently locked** since 2026-08-05 (`lock_type='permanent'`) | **NO** — forced renders are refused on locked slots by design; that is a human decision to lift |

**The lesson worth carrying: "the header shows text" is one symptom over at least four causes**, and
only the first is a render problem. `render_inputs.plan_logo` being non-empty does **not** mean the
header can show a logo — webdesign.co.uk recorded a real logo digest as an input and its template
had nowhere to put it. **Check the component's template for a logo branch before blaming the data:**
```sql
SELECT cc.name, (cc.html_template LIKE '%logo-img%') AS has_img_branch
FROM site_components sc JOIN sites s ON s.id=sc.site_id
JOIN content_components cc ON cc.id=sc.component_id
WHERE s.domain='<domain>' AND sc.slot_name='header';
```
