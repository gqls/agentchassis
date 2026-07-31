# RUNBOOK — webdesign.uk build service

Commands that were hard to get right, with the gotcha attached. When one changes,
change it **here**.

---

## Grounding queries

DB access (from CLAUDE.md):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

**Gotcha:** there is no `postgres` role — `psql -U postgres` fails with
`FATAL: role "postgres" does not exist`. Use `clients_user` for everything,
including `\l`.

### Fleet size (the §12 figure — re-run before quoting any scale claim)

```sql
SELECT count(*) total,
       count(*) FILTER (WHERE status='deployed') deployed,
       count(*) FILTER (WHERE domain LIKE 'pool-%') pools
  FROM sites;
```

**Gotcha:** `count(*)` alone answers a different question — 17 of the 32 rows are
empty `pool-*.internal` shells and 1 is `system.internal`. A bare total overstates
the fleet by more than 2×.

### What a new site actually needs seeded

```sql
SELECT aspect, is_current, created_at::date FROM site_specs
 WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com')
 ORDER BY created_at;
```

Read `oufe/SEED_2026-07-25_oufe_site_and_specs.sql` alongside it — its preamble
explains *why* three of those aspects must exist **before the first page is
written** (`evidence_base` gates the whole claims layer and silently no-ops if
absent; a missing site email makes the hallucinated-email check fail open;
a missing `imagery_style_guide` makes `content_hero` generate unstyled).

### Agent definitions

```sql
SELECT type, category, status FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
 ORDER BY type;
```

**Gotcha:** the column is `display_name`, **not** `name` — `SELECT type, name`
fails. Run `\d agent_definitions` first, as CLAUDE.md says.

### Model cost of a build (do this before pricing anything)

```sql
\d llm_call_log
SELECT count(*), min(created_at)::date, max(created_at)::date FROM llm_call_log;
-- 45,205 rows, 2026-03-25 → 2026-07-28  (checked 2026-07-28)
```

**Gotcha, inherited from idea.uk:** a per-run figure is only a **floor** unless
every call in the run is logged. `EVIDENCE_2026-07-27_ai_unit_economics.md`
records exactly this — usage logging was gated on cache activity, so calls whose
system prompt fell under the cacheable minimum (512 tokens on Opus 5, 1024 on
Sonnet 5) logged nothing, and 3 of 5 calls were invisible. Check the denominator
before quoting a total.

### Which model is the fleet actually on

```sql
SELECT model, count(*), max(created_at)::timestamp(0) latest FROM llm_call_log
 WHERE created_at > now() - interval '4 days' GROUP BY 1 ORDER BY 3 DESC;
```

**Gotcha:** run this before acting on any statement of the form "we're on model
X". On 2026-07-29 the answer was 1,468 Sonnet 5 and **zero Fable 5** while the
session itself ran Fable — the phrase covered the session and an intention, not
the fleet. DB model config is live on write, so the distinction decides whether
the pre-flight checks below are still owed.

### Live-probe a model from inside the cluster — no local Anthropic client needed

This shell has no `ANTHROPIC_API_KEY` and no `ant` CLI. Any pod that calls
Anthropic already carries the key as an env var — reuse it rather than
provisioning a separate credential for a one-off check:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')

# write the request body in ONE exec call — a separate exec to `cat` then a
# second to `wc` can read 0 bytes; combine write+verify in one shell.
kubectl -n ai-persona-system exec -i "$POD" -- sh -c 'cat > /tmp/probe.json && wc -c /tmp/probe.json' < request.json

kubectl -n ai-persona-system exec "$POD" -- sh -c '
wget -q -O /tmp/resp.json -S \
  --header="x-api-key: $ANTHROPIC_API_KEY" \
  --header="anthropic-version: 2023-06-01" \
  --header="Content-Type: application/json" \
  --post-file=/tmp/probe.json -T 120 \
  https://api.anthropic.com/v1/messages 2>&1
cat /tmp/resp.json
'

# clean up — do not leave request/response files on a production pod
kubectl -n ai-persona-system exec "$POD" -- sh -c 'rm -f /tmp/probe.json /tmp/resp.json'
```

**Gotcha:** the pod's image has `wget` (BusyBox) and **no `curl`, no `jq`** —
confirmed via `which curl; which jq; which wget`, don't assume. BusyBox `wget`
takes `--post-file` (not `--data @file`) and `-S` prints response headers to
stderr. `-O -` to stdout also works, but stderr/stdout interleave oddly under
`kubectl exec`; writing to a file and `cat`-ing it back is cleaner.

**This is a spend action against the org's real Anthropic account** — treat it
with the same care as any other live cluster action, not as read-only
diagnostics.

### Fable-5 pre-flight (PLAN §7b) — in this order

1. **Org data retention ≥ 30 days.** A ZDR org gets `400 invalid_request_error`
   on *every* Fable request, and the error names the request, not the setting.
2. **Grep the chassis LLM call layer** for params Fable rejects with a 400:

```bash
grep -rn "temperature\|top_p\|top_k\|budget_tokens\|\"thinking\"" \
  --include="*.go" platform/ internal/ | head -40
```

   **Gotcha:** a clean grep of Go source is not the whole answer — model params
   also live in `agent_definitions.default_config` as live DB rows, so check
   both. (Precedent: all 16 council seats set `max_tokens=8000` from config.)
3. **Then** measure one real Fable build from `llm_call_log`.

## DNS

```bash
dig +short webdesign.uk A
dig +short webdesign.uk NS
```

~~Empty is **expected** as of 2026-07-28 — the owner has not pointed it yet. Empty
is therefore not evidence about registration either way.~~

> **CORRECTED 2026-07-31 — `webdesign.uk` IS pointed, and has been for some time.**
> The line above was carried forward from the owner's 07-28 remark *"I haven't
> pointed the dns yet"* and then repeated in the PLAN and in memory **for three
> days without one `dig`**. It cost nothing this time only because the truth turned
> out to be *more* convenient. Logged in `WRONG_CALLS.md`.
> **A statement about infrastructure is checkable in one second. Check it.**

### Measured state, 2026-07-31 ~19:20 UTC

**Never quote this table without re-running it** — it is a snapshot of someone
else's control panel, not a repo fact.

| domain | NS | A | in front | serving |
|---|---|---|---|---|
| `webdesign.uk` | Cloudflare (`alexis`/`leah`) | `172.67.223.216`, `104.21.54.51` (+AAAA) | **Cloudflare, proxied** | **JSON 404 from the Worker→B2 static path** |
| `ugg2.com` | Cloudflare (`alexis`/`leah`) | `199.59.243.228` | **none — grey/DNS-only** | nothing; port 80 **times out** (registrar parking IP) |
| `webdesign.co.uk` | — | — | Cloudflare, proxied | **200**, `x-amz-*` headers ⇒ B2 origin |
| `idea.uk` | **Hetzner** | `116.203.204.115` (the VM, bare) | **NONE** | `server: nginx/1.28.3 (Ubuntu)` |
| `relojistas.com` | Cloudflare | `172.67.199.16`, `104.21.34.62` | **Cloudflare, proxied** | `server: cloudflare` |

```bash
# the whole table, re-runnable
for d in webdesign.uk ugg2.com idea.uk relojistas.com; do
  echo "=== $d ==="; dig +short NS $d; dig +short A $d
  curl -sS -o /dev/null -D- -m 12 "https://$d" 2>&1 | grep -iE '^server|cf-ray'
done
```

**Gotcha — `dig` alone cannot tell you whether Cloudflare is in the path, and the
A record is the *least* reliable way to ask.** Cloudflare NS with a non-Cloudflare
A record means **grey/DNS-only**: delegated but unproxied, so no WAF, no
Turnstile, no rate limiting, and the origin IP is public. `ugg2.com` is in exactly
that state right now. The discriminator is the **response header** — `cf-ray`
present ⇒ proxied; `server: nginx/...` ⇒ you are talking to the origin directly.

### The domain→B2 object key — answered by the 404, not by reading anything

```bash
curl -sS https://webdesign.uk
# {"error":"B2 returned error","objectKey":"webdesign.uk/index.html","status":404, …}
```

**The key is the hostname verbatim, `<host>/index.html`.** This resolves PLAN
§6(i)'s `[UNVERIFIED]` and opens §6a route (iii).

**Gotcha:** `webdesign.uk` has **no row in `sites`** (checked the same minute), and
the Worker still built a key for it. That is what makes the mapping look
host-derived rather than allow-listed — but it is **one sample**, and the Worker's
source is **not in this repo** (`grep -rn 'objectKey\|B2 returned error'` → no
match). Settle it with the `test.ugg2.com` probe in PLAN §6a before designing on it.

**Second gotcha, security-shaped:** that JSON is served to anyone who visits *any*
unpopulated domain on the account. It names the origin technology and the
object-key convention. Harmless alone — bucket keys are not credentials — but it
is a free disclosure with a one-object fix (publish a holding page) or a
Cloudflare redirect rule.

### What to create, when P1 has a box

**Do none of this before the box exists.** A Cloudflare Tunnel writes its own DNS
record and will overwrite anything staged now.

| host | type | value | proxy | why |
|---|---|---|---|---|
| `webdesign.uk` | (tunnel-managed) | — | orange | replaces the current Worker→B2 record |
| `www.webdesign.uk` | CNAME | `webdesign.uk` | orange | people type it |
| `webdesign.uk` | TXT | SPF | — | order confirmations; **without it they land in spam** |
| `webdesign.uk` | TXT `_dmarc` | DMARC | — | as above |
| `webdesign.uk` | MX | a real mailbox | — | customers **reply** to confirmations |
| `*.ugg2.com` | A/CNAME | Worker path (route iii) or VM (route ii) | **orange, required** | previews; grey ⇒ cert warning on the trust page |
| `ugg2.com` | A/CNAME | holding page | orange | people trim the URL |
| `ugg2.com` | MX | `0 .` (null MX) | — | it sends no mail; stops spoofing |

**Gotcha:** a proxied wildcard is covered by Universal SSL for **one label only** —
`acme.ugg2.com` yes, `a.b.ugg2.com` no. Keep preview slugs single-label.

## Deploy path facts worth not re-deriving

- Default artefact repo is `sites` (→ GitHub Action → B2). VM-hosted sites
  override to `vm-sites` via `sites.github_repo`; resolution is
  `resolveGitRepoNameDB` and `git_repo_resolution_test.go` documents the bug it
  prevents.
- Only `idea.uk` and `relojistas.com` are on `vm-sites`. Only `relojistas.com`
  has a non-empty `deploy_config`.

## Not yet written

Everything operational — provisioning the box, the engine build, Stripe test
mode, the preview vhost. Those arrive with P1; this file gets them then, not
before.

---

## Cloudflare: a 302 holding redirect for a domain with no site yet

**Why 302 and not 301.** A 301 is cached by browsers close to permanently and is
**not reliably flushable** — you cannot make a visitor forget it. webdesign.uk's
redirect is temporary by definition (it is deleted the day P1 ships), so a 301
would strand returning visitors on webdesign.co.uk after launch. Cloudflare
defaults this field to 301; **change it.**

### Steps (Cloudflare dashboard)

1. Select the **`webdesign.uk`** zone → **Rules** → **Redirect Rules** →
   **Create rule**.
2. Name it something that says it is temporary, e.g. `holding — delete at P1`.
3. **If incoming requests match** → **All incoming requests** (simplest; covers
   apex and any hostname in the zone). Use a custom expression
   `(http.host eq "webdesign.uk")` only if `www` needs different treatment.
4. **Then… Take action** → **URL redirect**
   - Type **Static**
   - URL `https://webdesign.co.uk/`
   - Status code **302 (Found)** ← the field that defaults to 301
   - Preserve query string **off** (the target is a different site)
5. Deploy.

### The DNS half — do NOT skip it

A Redirect Rule only fires if the request reaches Cloudflare, so the hostname
**must have a proxied (orange) record**. But it must not keep pointing at a real
box:

| record | value | proxy |
|---|---|---|
| `webdesign.uk` A | **`192.0.2.1`** | orange |
| `www` A | `192.0.2.1` | orange |

**`192.0.2.1` is TEST-NET-1 — reserved for documentation and never routable.** The
redirect is served at Cloudflare's edge and the origin is never contacted, so the
address only has to exist for the record to be proxiable.

**Gotcha, and this is the whole reason for the dummy address:** leaving the record
on a real box means the redirect is the *only* thing preventing that box being
served under this hostname. Disable the rule, mis-scope its expression, or hit a
path it does not match, and the origin answers again. On 2026-07-31 that origin was
idea.uk's live shop, which serves under **any** hostname (no `Host` validation) and
would take a real payment. **Fail closed: point it at an address that cannot
answer.**

### Verify from outside

```bash
curl -sS -o /dev/null -D- -m 15 https://webdesign.uk | grep -iE '^HTTP|^location'
# want: HTTP/2 302   and   location: https://webdesign.co.uk/
```

**Gotcha:** test in a private window or with `curl`, never in a browser you have
already loaded the domain in — if you visited it while a 301 was live, your own
browser will keep redirecting and you will "verify" a rule that is not running.

**Fallback if Redirect Rules are unavailable on the plan:** a **Page Rule** with
*Forwarding URL* → **302** does the same job. Older mechanism, smaller free quota.

## Cloudflare: wildcard previews on ugg2.com

**Measured 2026-07-31 21:58 — the apex is wired to the bucket and the wildcard is
NOT there:**

```bash
$ curl -sS https://ugg2.com
{"error":"B2 returned error","objectKey":"ugg2.com/index.html","status":404, ...}
$ dig +short A test.ugg2.com    # -> empty
$ dig +short A acme.ugg2.com    # -> empty
```

The apex reaching the Worker proves the **route** works for `ugg2.com`; it says
nothing about subdomains. A proxied wildcard record would return CF anycast
addresses for *any* subdomain, so **empty means the record does not exist**.

Two things are needed, and they are separate systems — either alone looks broken:

1. **DNS:** a record with name `*` (A or CNAME, same target as the apex),
   **Proxied (orange)**. Grey ⇒ certificate warnings on every preview.
2. **Worker route:** add `*.ugg2.com/*` alongside the existing `ugg2.com/*`. A
   route for the apex does **not** match subdomains.

Then the test that settles PLAN §6a claim (B):

```bash
# upload an object at key  test.ugg2.com/index.html  in the bucket, then:
curl -sS -o /dev/null -D- https://test.ugg2.com/ | head -3     # want 200
curl -sS https://test.ugg2.com/ | head -c 120                  # want your file
```

**Gotcha:** if it 404s, read the `objectKey` in the JSON body before assuming the
wildcard failed — it tells you exactly which key was looked for, so a key/filename
mismatch is distinguishable from a routing miss. Those are different failures with
different fixes.

**Gotcha:** Universal SSL covers **one label** — `acme.ugg2.com` is covered,
`a.b.ugg2.com` is not. Keep preview slugs single-label.
