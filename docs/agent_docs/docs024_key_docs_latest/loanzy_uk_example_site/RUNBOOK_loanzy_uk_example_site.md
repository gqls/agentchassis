# RUNBOOK — loanzy.uk

Commands that were hard to get right, with the gotcha attached. Add here, not to scrollback.

## Is anything built on loanzy.uk? (the question that started this lane)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT id, domain, created_at FROM sites WHERE domain ~* 'loan|lend|borrow|credit' ORDER BY created_at;"
```
`[MEASURED 2026-08-18]` five rows, **none of them loanzy**: `loancalculator.co.uk`,
`loanandmortgagecalculator.co.uk`, `loancash.co.uk` (01 Aug), `lendzy.co.uk` (02 Aug),
`pool-mortgages-lending.internal`.

⚠ **`loanzy` and `lendzy` are one letter apart and both are loan brandables.** The owner
remembered "an FCA-rules site built alongside loancash" and that is **`lendzy.co.uk`**
(*"Know the Rules Before You Borrow"*, built the day after `loancash.co.uk`, L10's shadow —
`portfolio_positioning/MISSION_2026-08-02_lendzy_shadow.md`, deliberately no register row).
Search the family, not the name you were given, or you will confirm the wrong domain.

## Live state of the domain (no site, but fully wired)

```bash
curl -s -o /dev/null -w "%{http_code} %{size_download}B\n" --max-time 20 https://loanzy.uk/
dig +short NS loanzy.uk; dig +short A loanzy.uk
```
`[MEASURED 2026-08-18]` apex **404, 9 bytes**; www identical; NS `alexis`/`leah.ns.cloudflare.com`;
A `104.21.70.112` / `172.67.223.19` (Cloudflare proxy — orange).

**A 404 of 9 bytes is the router speaking, not a broken domain.** It is
`portfolio-sites-router` saying it has no site under this hostname. Compare the failure this
domain had on 2026-08-09: a **timeout**, because the NS were delegated to Cloudflare while the
zone did not exist, so requests reached the placeholder origin `199.59.243.228`, which accepts
nothing (`domains_cloudflare_rollout/NOTES`, and LANDMINES).

## Cloudflare state

```bash
T=$(cat ~/.config/cloudflare/token)
curl -s -4 -H "Authorization: Bearer $T" "https://api.cloudflare.com/client/v4/zones?name=loanzy.uk"
source ~/.cloudflare/404-token.env
curl -s -4 -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" \
  "https://api.cloudflare.com/client/v4/zones/18c86604a6066bdb717e11ff28effb48/workers/routes"
```
`[MEASURED 2026-08-18]` zone `18c86604a6066bdb717e11ff28effb48`, **active**; routes
`loanzy.uk/*` **and** `*.loanzy.uk/*` → `portfolio-sites-router`.

⚠ **Neither token reads DNS.** `dns_records` returns `success: false` on the main token and on
the 404 token (documented in `domains_cloudflare_rollout/NOTES` for the latter). Use `dig` —
and note `success:false` prints an empty result set that looks exactly like "no records".

⚠ **Pin `curl -4`.** At least one v6-sourced call to this API was refused naming the v6 temp
address; SLAAC addresses rotate daily so v6 can never be stably allow-listed.

## The build, when the prompt lands (Phase 2)

Seed from the prompt alone, then dispatch — image first, then seeds; a seed naming an
unregistered action fails at runtime:

```bash
./scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh loanzy.uk \
  --email … --mission-file MISSION_<date>_loanzy_uk.md
```
No `--roadmap-file` flag exists (webdesign lane, HANDOFF_2026-08-06). Do not dispatch within
~300s of a chassis pod restart — the spawn is silently dropped. Verify the item **landed**
(`needs_domain_research` moving `triaged → claimed → complete`), never exit 0.

## Verifying the pair afterwards

Verify at the **served page**, cache-busted — not at item status, and not at
`rendered_html` in the DB. `complete` is not proof the work happened.

## The dispatch that was actually used (2026-08-18)

```bash
bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh loanzy.uk
```
**`bash <path>`, not `./<path>`** — the script is mode 644 in the tree and `./` fails with
"Permission denied". No `--mission`, `--email`, `--phone` or seed: the domain string is the
whole input, which is the point of this lane.

Then verify it LANDED (never trust exit 0 — `kcat -P` can publish nothing and exit clean):

```sql
SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='<corr>'::uuid;
SELECT id, domain, status, build_status FROM sites WHERE domain='loanzy.uk';
SELECT w.item_type, w.status, w.handler_agent FROM site_work_items w
  JOIN sites s ON s.id=w.site_id WHERE s.domain='loanzy.uk' ORDER BY w.created_at;
```
⚠ Qualify `status` in the join — both tables have one, and an unqualified reference aborts the
whole `psql` invocation with "column reference is ambiguous", which looks like the query found
nothing.

## Watching the cascade

```sql
SELECT ss.aspect, ss.source_agent, ss.is_current, ss.created_at FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='loanzy.uk' ORDER BY ss.created_at;
```
The strategy spec is where the framework's own answer to "what is this domain for" first
becomes readable — that is the artefact to judge before any page deploys.

## `garden-tools.uk` — zone setup for the next clean-domain test (2026-08-23, DONE)

Followed `portfolio_positioning/RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md`
§"PROVEN RECIPE" exactly. Token `~/.config/cloudflare/portfoliotoken` (All zones: Zone/DNS/
Workers-Routes edit, no expiry), account `13044f178ae0b156961065f55c8fada8`.

Zone `82d90228c20877e2b3fc8470c2bc73d1` · NS `alexis`/`leah.ns.cloudflare.com` (owner set them
at the registrar) · one proxied apex `A → 192.0.2.1` · route `garden-tools.uk/*` →
`portfolio-sites-router` · `www` record + route added by `scripts/cloudflare/add_www_redirect.sh
--apply garden-tools.uk`.

**Timings, measured — useful for sizing the next domain:** zone `pending → active` **60s** after
`PUT /zones/<id>/activation_check`; Universal SSL issued **~90s** after that; `www` 301 live
**~20s** after its record was added.

⚠ **The apex returns an EMPTY body for two unrelated reasons during setup, and they look
identical.** Before activation, Cloudflare's nameservers serve the raw `192.0.2.1`
(TEST-NET-1, unroutable) — nothing connects, so HTTP, HTTPS and TLS all fail together. After
activation but before the certificate, the proxy is live but the handshake fails with **alert 40
/ "no peer certificate available"**. Separate them by asking the layers individually rather than
re-running the same request:
```sh
dig +short @alexis.ns.cloudflare.com <domain> A     # proxy IPs (104.21.x/172.67.x) = activated
echo | openssl s_client -connect <proxy-ip>:443 -servername <domain> 2>&1 | grep -E "subject=|alert"
curl -s --resolve <domain>:443:<proxy-ip> https://<domain>/ | head -c 200
```
⚠ **Use `--resolve` against a proxy IP throughout.** A local resolver caches the pre-activation
`192.0.2.1` for its TTL, so `curl https://<domain>/` keeps failing long after the zone is
correct — a stale cache and a broken zone are indistinguishable from the client.

**Ready state confirmed:** apex serves the router's **9-byte `Not found`** and
`https://www.garden-tools.uk/` → **301 → `https://garden-tools.uk/`**. That 9-byte 404 is the
positive signal to require before dispatching any build: the route is live and the bucket is
empty. `sites` and `site_work_items` remain at **0 rows** — nothing is built.

## "Why hasn't my build started?" — read the selector, don't watch your row (added 2026-08-23)

A submitted item sits at `triaged` with `claimed_at` NULL for tens of minutes on a healthy estate.
That looks identical to `bugs_open/327`'s dispatch drop and to a stalled queue, and **your own row
cannot tell the three apart**. Ask the selector instead.

```sh
# 1. Did it land at all? (a row exists = not the 327 drop)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT item_type, status, claimed_at FROM site_work_items w JOIN sites s ON s.id=w.site_id
   WHERE s.domain='<domain>';"

# 2. Is the fleet dispatching at all, or is it just not you?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT max(claimed_at) AS last_claim_fleetwide FROM site_work_items;"

# 3. THE ONE THAT ANSWERS IT — how many sites are ahead of you in the FIFO?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT count(DISTINCT wi.site_id) AS sites_ahead
    FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE s.locked_at IS NULL AND wi.status IN ('triaged','approved')
     AND wi.attempt_count < wi.max_attempts
     AND (wi.retry_after IS NULL OR wi.retry_after <= NOW())
     AND wi.created_at < '<your item created_at>'
     AND NOT EXISTS (SELECT 1 FROM site_work_items a
                      WHERE a.site_id=wi.site_id AND a.status='claimed');"
```

**Multiply `sites_ahead` by ~90s and that is your wait.** The upstream picker is
`build-pipeline-trigger`, one tick per ~90s, and its `find_dispatchable_site` step orders
`wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1` — **FIFO by item age; `priority` only
breaks ties within an identical timestamp.** Measured 2026-08-23: a priority-**5**
`needs_domain_research` waited **24m52s** behind 64 priority-**110** `content_rewrite` items,
purely because they were created earlier. Read the live selector before quoting this — it is config,
so it can change without a deploy:

```sh
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c \
 "SELECT jsonb_pretty(default_config->'workflow'->'steps'->'find_dispatchable_site')
    FROM agent_definitions WHERE type='build-pipeline-trigger'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

⚠ **A site with ANY `claimed` item is invisible to the selector** (`NOT EXISTS … status='claimed'`),
so a build is a strictly sequential walk of its own items no matter how parallel the fleet is — and
a stuck `claimed` item halts that site's whole build while looking like nothing is wrong.

## Collateral check that reports UNCHANGED/CHANGED instead of eight md5s to eyeball (added 2026-08-23)

Do not diff hashes by eye, and **do not use a baseline from another lane's notes without re-pinning
it first** — all eight moved under us on 2026-08-20 (`bugs_open/283`) while still reading as the
current pin. Pin your own values immediately BEFORE dispatch, then:

```sql
WITH baseline(id8, html_md5, schema_md5) AS (VALUES
  ('7d8b0503','<html>','<schema>'), ... )
SELECT b.id8, cc.function,
       CASE WHEN md5(cc.html_template)=b.html_md5 THEN 'UNCHANGED' ELSE '*** HTML CHANGED ***' END AS html,
       CASE WHEN md5(cc.input_schema::text)=b.schema_md5 THEN 'UNCHANGED' ELSE '*** SCHEMA CHANGED ***' END AS schema
FROM baseline b JOIN content_components cc ON left(cc.id::text,8)=b.id8 ORDER BY cc.function;
```

**Prove the instrument discriminates before you trust a clean result**: this one returned
`*** HTML CHANGED ***` on all eight against the stale 08-19 pins the same morning it returned
`UNCHANGED` on all eight against the fresh ones. A check that has never once come out the other way
is not evidence. Full harness: `after_test.sh` (this session's scratchpad; promote it here if it
survives a second run).

## ⚠ `kubectl`'s own error lines are LOCAL time; the database is UTC (added 2026-08-23)

While waiting on a `retry_after` I read a `kubectl` klog line stamped `19:29:41` and briefly
concluded a 19:20:32 retry had already fired. It had not — the DB said **18:29:50 UTC**. Client-side
`kubectl`/klog diagnostics print in the machine's local zone (**BST = UTC+1** here); everything in
`site_work_items`, `orchestration_states` and `date -u` is UTC.

**One hour is exactly the size of error that looks plausible**, so it does not announce itself: a
retry window that has not opened reads as one that has closed, and the natural next action is to
conclude the step failed again and act on it.

**Always stamp your own comparison** — put `date -u` in the same command as the query, and compare
UTC to UTC:

```sh
date -u; kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT item_type, status, attempt_count, retry_after FROM site_work_items w
   JOIN sites s ON s.id=w.site_id WHERE s.domain='<domain>';"
```

Never take the time from a `kubectl` error line, a pod log prefix, or a shell prompt. This bites
hardest in the BST half of the year, when the offset is non-zero but small enough to be mistaken for
clock skew rather than a zone.

## After the 2026-08-26 credit top-up: re-fire the outage's terminal residue

The credit outage (23:46:29Z 08-25 → top-up) permanently failed work items that exhausted
`max_attempts` against the dead API. **They do NOT self-heal** — the 1,400 `triaged` rows drain on
their own, these do not. `[MEASURED 08:56Z]` 21 items, 20 carrying "credit balance" in `error`,
bleed ~4/hour while the outage continues. Found by the 391 lane, verified here at the artefact.

```sql
-- 1. LOOK FIRST (the list, not the count):
SELECT wi.id, s.domain, wi.item_type, wi.updated_at, left(wi.error, 80)
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
 WHERE wi.status = 'failed' AND wi.updated_at > '2026-08-26 00:00:00+00'
   AND wi.error ~* 'credit balance|connection reset'
 ORDER BY wi.updated_at;

-- 2. Re-fire exactly those (bound the window at the top-up time; the error
--    predicate is what keeps a genuinely-broken item from riding along):
UPDATE site_work_items
   SET status = 'triaged', attempt_count = 0, error = NULL, retry_after = NULL, updated_at = now()
 WHERE status = 'failed'
   AND updated_at BETWEEN '2026-08-26 00:00:00+00' AND '<top-up time>'
   AND error ~* 'credit balance|connection reset';
```
The 1 non-matching failure (of 21) is judged on its own error text, by hand. The `needs_diagnosis`
row re-runs its 090 diagnosis on re-fire — expected.

## The verdict queue (record-mode rows) — read, release, and the promises around it

The acceptance council's model seats file VERDICTS: `deferred`, `handler_agent ''`,
`spec.filing_mode='record'`. Nothing dispatches them. ~~The seat's own silence-retraction clears the
ones that stop reproducing~~ — **CORRECTED 2026-08-26: that was true only for `dark_section_audit`**
(the gates map's single entry); the default gate for record rows (`recordModeSilenceRule`) is built
and rides the next roll — **until then, this queue is the WHOLE lifecycle: rows leave it only by the
release UPDATE below or a hand `cancelled`.** This query is the interim
release interface RFC_056's verdict note promised (a real surface is owed before rows accumulate
fleet-wide):

```sql
-- The queue, per site per seat:
SELECT s.domain, wi.spec->>'audit_source' AS seat, wi.item_type AS routed_as,
       count(*) AS verdicts, max(wi.created_at)::timestamp(0) AS newest
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
 WHERE wi.spec->>'filing_mode' = 'record'
 GROUP BY 1,2,3 ORDER BY 1, 4 DESC;

-- Read one site's verdicts in full:
SELECT wi.item_type, wi.severity, wi.summary, wi.spec->>'suggestion' AS suggestion
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
 WHERE s.domain = '<domain>' AND wi.spec->>'filing_mode' = 'record'
 ORDER BY wi.severity, wi.created_at;

-- Release ONE verdict into real work (deliberate, per row — never bulk):
UPDATE site_work_items
   SET status = spec->>'routed_status', handler_agent = spec->>'routed_handler', updated_at = now()
 WHERE id = '<id>' AND status = 'deferred' AND spec->>'filing_mode' = 'record';
```
⚠ A released row is claimed by the dispatcher within ~60s and its handler REGENERATES the page —
release is the moment the opinion becomes a rewrite, which is why it is a human verb.

## Hand-filing a framework work item that must actually RUN (learned 2026-08-31, the farmer logo)
A row filed at `detected` with `handler_agent` empty is a FLAG, permanently — the promoter's
`scored` CTE excludes it before any door is evaluated (`COALESCE(wi.handler_agent,'')<>''`),
and no held_detail line ever names it, so it "waits" silently for ever. To make a hand-filed
item promotable, mirror what the producing pipeline sets AT FILING:
```sql
UPDATE site_work_items SET handler_agent='<the handler>', pipeline='build', updated_at=now()
 WHERE id='<row>' AND status='detected';
```
(pipeline must be in build/content/design — door 1; the pair must have history — known-good
door; check the doors yourself in scheduled_tasks.pre_query, name='detected-item-promoter'.)
Gotcha that cost an hour: I copied the SHAPE of a completed row's spec but its handler_agent
was set by its producer at filing, not by promotion — the flag-only exclusion is invisible
because held_detail reports every REFUSED row and this one was never scored.

## Retiring a page CLUSTER whose CTAs other pages advertise (learned 2026-08-31, the farmer tool cull)
1. `pages.status='archived'` on the cluster; cancel moot review items.
2. Fire `216_TRIGGER_page_retraction.sh` — expect REFUSALS for every page a live page links;
   the referrers are in the RETRACTION_AUDIT row: `agent_error_log.context->'editorial_inbound'`.
3. **Do NOT file plain rerenders for the referrers — they cannot help.** The guard reads
   content_data (a rerender faithfully reproduces it), and the CTA recompute inside
   page-rerender only runs for **`spec.reason='cta_links_stale'`**. File the misdirected-cta
   check's own shape: item_type page_rerender, item_key `misdirected_cta:<page>:<site_id>`,
   spec `{"check":"misdirected_cta","reason":"cta_links_stale","page_name":…,"page_id":…}`,
   handler page-rerender, pipeline build, status detected.
4a. THE SPEC LAYER: census `site_specs WHERE is_current AND data::text ~* <tool names>` — the suggestion inventory (aspect `tools`) is cull work (supersede, never edit in place); prose specs go to the copy machinery; analyser artefacts re-derive via their lane. Without this, a future generation pass re-mints the culled content.
4. Verify the residue is ZERO before re-firing: count live components whose content_data
   LIKE any archived url — the recompute can only rewrite fields in ctaFieldNames; a LIST
   component (e.g. guide-list) holding the url needs its own treatment.
5. Re-fire 216 with the refused ids. Acceptance stays two-part: 404 now AND still-404 after
   the next ~08:0x/20:0x refresh with zero fresh page_rerender rows for the retired pages.

## Growth posture — the owner's "this site is ready" switch (WDS-020, built 2026-09-02)
Go is inert until the next roll; the key is read from sites.settings, so setting it costs
nothing before the roll and starts working the moment the binary lands.
```sql
-- hold growth on a site (tool chain files held; audit growth is already record-mode):
UPDATE sites SET settings = jsonb_set(COALESCE(settings,'{}'::jsonb),
  '{maintenance_profile,growth_posture}', '"hold"') WHERE domain='<domain>';
-- reopen:
UPDATE sites SET settings = settings #- '{maintenance_profile,growth_posture}' WHERE domain='<domain>';
-- what is held, and the release verb (recipe is also stamped on each row's spec):
SELECT id, item_type, summary FROM site_work_items
 WHERE spec->>'growth_held'='true' AND status='deferred';
```
Worked example: apis.uk holds since 2026-09-03 (site 1c6f3424, owner-named first). Gotcha: the hold FILES items (deferred, handler-less) rather than skipping them — an empty
result above on a held site means the machinery has not WANTED to grow it, not that the
switch failed. `source='owner-request'` rows bypass by design.
