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
