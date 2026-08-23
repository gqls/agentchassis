# RUNBOOK — apis.uk bees home page

Commands for this workstream, each with the gotcha that made it hard to get right.
**Status markers:** `[PROVEN]` = run here, output seen. `[UNPROVEN]` = written from
another workstream's runbook, not executed on this domain.

DB shell used throughout:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

> **CORRECTION to the oufe runbook, which says the agent definitions live in the
> templates DB.** They do NOT. `agent_definitions` is in **`clients_db`**
> (**216** rows as of 2026-08-22); `templates_db` has a table of the same name with
> **8** rows as of 2026-08-22, and querying it returns zero rows for every agent
> type you actually want — which reads exactly like "that agent does not exist"
> rather than "wrong database". [PROVEN]

---

## 1. Read the apis.uk zone WITHOUT changing it

The whole point of this workstream is not disturbing `tools.apis.uk`, so every
zone command here is read-only. Token: `~/.config/cloudflare/portfoliotoken`.

```bash
T=$(tr -d '\n' < ~/.config/cloudflare/portfoliotoken)
ZID=$(curl -s -H "Authorization: Bearer $T" \
  "https://api.cloudflare.com/client/v4/zones?name=apis.uk" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['result'][0]['id'])")

# DNS records
curl -s -H "Authorization: Bearer $T" \
  "https://api.cloudflare.com/client/v4/zones/$ZID/dns_records?per_page=200" \
  | python3 -c "
import json,sys
for r in json.load(sys.stdin)['result']:
    print(f\"{r['type']:8} {r['name']:28} -> {r['content'][:60]:62} proxied={r['proxied']}\")"

# Worker routes — THE ROUTE IS THE HAZARD, NOT THE DNS
curl -s -H "Authorization: Bearer $T" \
  "https://api.cloudflare.com/client/v4/zones/$ZID/workers/routes" \
  | python3 -c "
import json,sys
for r in json.load(sys.stdin)['result']: print(f\"{r['pattern']:32} -> {r.get('script')}\")"
```
[PROVEN 2026-08-22] Zone `a8c1ac6111424c218cb9e9368ed0586f`. **4** DNS records and
**2** worker routes as of 2026-08-22 — see PLAN §1 for the table.

**Gotchas**
- **A worker route beats DNS.** The apex CNAME still points at the tunnel, and
  that record is now *vestigial*: the route `apis.uk/*` intercepts at the edge
  before the origin is consulted. Reading the DNS alone tells you the opposite of
  what is happening, which is how three standing documents came to say a DNS swap
  was needed when it is not.
- **Never add `*.apis.uk/*` as a worker route.** It would match `tools.apis.uk`,
  intercept the live API and serve a B2 404. No DNS record changes, nothing looks
  wrong, and the API is dead. See LANDMINES.
- `scripts/cloudflare/add_www_redirect.sh` is a per-zone sweep. **Never run it
  unnamed against this zone**; if it must run, name the zone explicitly and read
  its classification output first.

## 2. Identify which server answers a hostname [PROVEN 2026-08-22]

Do not reason about this from config — ask the hostnames, they identify themselves:

```bash
curl -sS https://apis.uk/                    # -> "Not found"  = the WORKER (worker.js:91)
curl -sSI https://www.apis.uk/               # -> 301 to apex = the WORKER (worker.js:23)
curl -sS https://zzqq-probe-test.apis.uk/    # -> 0 bytes     = island probe :8082
curl -sS https://tools.apis.uk/              # -> 0 bytes     = island Caddy :8081
```

**Gotcha — the one that has already burned two sessions.** `https://tools.apis.uk/`
returns **404 at the root and always has**; that is the Caddy `handle { respond 404 }`
arm, not a dead API. `WRONG_CALLS.md` records a session curling the root, reading
404, and concluding the API was down. **Test the real endpoint instead:**
```bash
curl -sS -o /dev/null -w "%{http_code}\n" -X POST \
  https://tools.apis.uk/api/v1/tools/gauntlet/round \
  -H 'Origin: https://vonc.com' -H 'Content-Type: application/json' -d '{}'
```
[PROVEN 2026-08-22] → **200**. That is the tools-api liveness check; use it as the
before/after control for anything touching this zone.

## 3. Seed and submit [PROVEN 2026-08-22]

```bash
cat docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/SEED_2026-08-22_apis_uk_site_and_specs.sql \
 | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
   psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

bash ./scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh apis.uk \
  --email apis-uk@contactforsales.com \
  --mission-file docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/MISSION_BRIEF_apis_uk_2026-08-22.md
```
Submitted 2026-08-22 12:18Z. `CORRELATION_ID=ba7a9c24-aea3-4fd0-9def-7e1d6f1cf891`,
`ORCHESTRATION_ID=59161d99-391e-4b93-a714-9e0a33ac2bfa`.

**Gotchas**
- The script is **not executable** in this tree — `./082_…` gives *Permission
  denied*. Invoke it as `bash ./082_…`.
- **Validate the seed's JSON before applying it**, and validate the *decoded*
  regexes, not just that the JSON parses. See §4 — this caught three real defects.
- The seed is one transaction with `\set ON_ERROR_STOP on`, so a failure part-way
  rolls the whole thing back. It did exactly that on the first attempt (a missing
  `FROM sites` clause on the third insert): `SELECT count(*) FROM sites WHERE
  domain='apis.uk'` returned **0** afterwards, which is the behaviour you want.
- `\set ON_ERROR_STOP on` inside the file is not the same as `-v ON_ERROR_STOP=1`
  on the command line. Both are set here; keep both.

## 4. Validate the evidence base BEFORE applying it [PROVEN 2026-08-22]

This is the step worth copying to other site seeds, because it found defects that
every cheaper check passed.

```bash
python3 - <<'PY'
import re, json
src = open('…/SEED_2026-08-22_apis_uk_site_and_specs.sql').read()
d = json.loads(re.search(r'\$eb\$(.*?)\$eb\$', src, re.S).group(1))
pats = {b['pattern'] for b in d['banned_claims']}
for s in ["bees pollinate 75% of crops", "two million flowers", "a worker lives six weeks"]:
    print(("CAUGHT " if [p for p in pats if re.search(p, s, re.I)] else "MISSED "), s)
PY
```

**Gotchas — all three were real, and all three passed a plain `jq -e .`**
- **`\\\\.` in the file decodes to a regex meaning "literal backslash", not
  "decimal point".** The seed's own safety net — *"an invalid regex degrades to a
  literal substring, so a typo never silently drops a ban"* — **does not catch
  this**, because the pattern is perfectly valid, just wrong. A valid-but-wrong
  regex never fires and reports clean for ever. In the SQL file you need `\\.`
  (backslash backslash dot), which JSON decodes to `\.`.
- **Digit-adjacent patterns do not see a magnitude WORD.** `[0-9][0-9,]* ?flowers`
  misses `2 million flowers` — which is the exact form the most famous bee
  statistic takes. Ban the magnitude words separately, in digit, spelled-out and
  bare-plural forms.
- **`live` does not match `lives`.** `live (for )?[0-9]` misses `lives for 6 weeks`.
  Use `lives?`.
- **RE-RUNNABLE NOW: `check_evidence_base.py <domain>`** (in this directory). It asserts
  three things and **exit 2 means check 3 did not run** — see below.
- **Assert the clean cases too, not only the catches.** A ban list that matches
  everything is useless in the other direction; the suite here holds **8** ordinary
  bee sentences that must stay clean as of 2026-08-22.

## 5. Monitor the build

```sql
SELECT status, current_step, error FROM orchestration_states
 WHERE correlation_id='ba7a9c24-aea3-4fd0-9def-7e1d6f1cf891'::uuid;

SELECT aspect, source_agent, is_current, created_at FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='apis.uk' ORDER BY created_at;

SELECT item_type, wi.status, handler_agent, LEFT(summary,60)
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
 WHERE s.domain='apis.uk' ORDER BY priority;

SELECT name, page_type, build_status, jsonb_array_length(sections) AS n_sections
  FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='apis.uk');
```

**Gotchas**
- **An absent orchestration row means QUEUED, not dropped.** Dispatch latency of
  16–30 minutes is normal under fleet load. Do not resubmit on that evidence.
- **The page count is the check that "home page only" held.** More than one row in
  `pages` means the roadmap_brief did not bind — read it back before blaming the
  planner, and check `is_current`/`pinned` survived.
- A `validate_content` blocker reason is **not recoverable from the DB** — watch
  the chassis log live during `validate_page_content` or you will not learn why.
- New sites are **not enrolled in the discovery/audit sweeps**. "No discovery
  items" on apis.uk means invisible, not healthy.

## 6. Publish check — the artefact, not the status

The worker serves B2 object key `<hostname><path>`, so the home page is
`b2://portfolio-sites/apis.uk/index.html`.

```bash
b2 ls b2://portfolio-sites/apis.uk/
curl -sI "https://apis.uk/?cachebust=$(date +%s)"
```
[PROVEN 2026-08-22, as the BEFORE control] bucket prefix empty, apex serves 404
`Not found`. That is the disconfirming baseline: if the apex still says exactly
that after a deploy, nothing shipped.

**Gotcha** — an unbusted GET tells you what a cache thinks, not what is deployed.
**And re-run the §2 tools-api probe after any publish**, because "the bees page went
live" and "the API still answers" are independent facts.


## 7. The evidence-base test suite, and the five gaps it found [PROVEN 2026-08-23]

```bash
python3 docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/check_evidence_base.py apis.uk
# exit 0 = all three checks passed · exit 1 = a gap or false positive · exit 2 = check 3 could not RUN
```

Three assertions, and the third is the one people leave out:

1. every FORBIDDEN sentence is caught — no gaps (**23** as of 2026-08-23);
2. every PERMITTED sentence is clean — no false positives (**12**);
3. **the site's own LIVE PROSE is clean** — the list does not block the copy we want
   (**8,907** chars of the framework's own bee prose, zero hits).

**Why it exists: inspecting a ban list tells you what you MEANT, not what the regexes DO.**
All five defects below passed `jq -e .` and looked correct on the page. Every one was found
by asserting on SENTENCES:

| gap | why it was invisible |
|---|---|
| `\\.` decoding to "literal backslash", not "decimal point" | a **VALID** regex, so evidence_base's own stated safety net ("an invalid regex degrades to a literal substring") never fired. Would have read clean for ever. |
| `2 million flowers` — the most repeated bee statistic there is | a magnitude WORD sat between the number and the noun; every digit-adjacent pattern needs them adjacent |
| `colonies fell by 40%` | the decline pattern required the decline word AFTER the figure |
| `the population has halved since 1990` | `halved` carries its own magnitude and offers no number to anchor on |
| `sign up to our newsletter` | slipped between `sign up` and `our` in the commercial pattern |

**Gotchas**
- **A failed page fetch is FATAL (exit 2), deliberately.** The first cut returned empty on
  any exception and still printed `ALL CONSISTENT` — and Cloudflare 403s urllib's default
  User-Agent, so check 3 silently did not run on the very site it was written for. **A check
  that quietly does not run is worse than no check**: it produces a pass that outlives the
  blindness. `--skip-live` allows it deliberately (e.g. before first deploy).
- **`$?` after a pipe is the PIPE's exit code.** Verifying the above with
  `python3 check.py … | tail -3; echo $?` reported **0** while the script was correctly
  printing `INCOMPLETE` and returning 2. Measure with `cmd >/dev/null 2>&1; echo $?`, or
  `${PIPESTATUS[0]}`.
- **The two lists ARE the specification.** Extend `FORBIDDEN` the moment you think of a
  sentence the page must never contain; that is cheaper than reasoning about regexes.
- Definitional numbers are deliberately PERMITTED (`six-sided cells`, `six legs`, `one
  queen`) via a `writer_block` carve-out — a number that is part of what a thing IS is not
  a measurement. The test asserts both directions so the carve-out and the bans cannot drift
  apart.
