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
