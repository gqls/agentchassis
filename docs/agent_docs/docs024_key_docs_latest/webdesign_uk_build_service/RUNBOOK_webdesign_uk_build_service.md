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

### Measured state, 2026-08-02 ~23:20 UTC — three of the five rows above have moved

| domain | NS | in front | serving |
|---|---|---|---|
| `webdesign.uk` | Cloudflare | **proxied, origin `192.0.2.1`** | **302 → `https://webdesign.co.uk/`** (Page Rule) |
| `www.webdesign.uk` | — | proxied, `192.0.2.1` | **302**, same rule |
| `ugg2.com` | Cloudflare | **proxied** (was grey on 07-31) | Worker→B2, `objectKey: ugg2.com/index.html` |
| `*.ugg2.com` | Cloudflare | **proxied — NEW** | Worker→B2, key per host |
| `idea.uk` | **Cloudflare** (was Hetzner) | **proxied, SSL `strict`** | **200**; origin cert `CN=idea.uk` (Google Trust WE1) |

**`idea.uk`'s origin is now firewalled** — `116.203.204.115:80` and `:443` are both
`FILTERED` from a random internet host, while the site serves 200 through
Cloudflare. That is Option B of the idea.uk handoff, taken and evidently done
right. Not done by this lane.

> ### ⚠ LANDMINE — your stub resolver's cache makes a live site look DEAD, and `dig @1.1.1.1` will not show you
>
> **This cost me a false alarm on 2026-08-02: I told the owner idea.uk was down. It
> was serving 200 the whole time.**
>
> After a domain moves to new nameservers, `systemd-resolved` (127.0.0.53) can hold
> the **pre-migration** answer for a long time. If the old origin has since been
> firewalled — which is exactly what a well-executed migration does — then every
> request from this machine goes to an IP that now silently drops packets, and you
> get a **total timeout with no HTTP status at all**. There is no error to read.
>
> The trap is that the obvious cross-check **agrees with you**: `curl` hangs, and a
> `dig` looks fine — because `dig +short A x @1.1.1.1` **bypasses the system
> resolver entirely**. Two tools, two different resolvers, and the disagreement is
> invisible unless you deliberately ask both:
>
> ```bash
> getent ahosts idea.uk | awk '{print $1}' | sort -u   # what curl ACTUALLY uses
> dig +short A idea.uk                                  # system resolver
> dig +short A idea.uk @1.1.1.1                         # authoritative-ish
> # 08-02: system said 116.203.204.115 (old, firewalled); 1.1.1.1 said the CF anycast pair
> ```
>
> **Before reporting any site down, pin the address and retry:**
> `curl --resolve host:443:<ip> https://host/` against an IP from `@1.1.1.1`. If
> that returns 200, the site is up and your resolver is stale —
> `sudo resolvectl flush-caches`.
>
> **The generalisation worth carrying:** a timeout is not evidence about the
> server. It is evidence about the *path*, and the path starts at your own
> resolver. `curl` hanging where `openssl s_client -connect <ip>` succeeds is the
> signature — that pair differ in exactly one step, name resolution.

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

> **DONE 2026-08-02 — and step 2 above was already done before I got there.**
> The Worker route `*.ugg2.com/*` → `portfolio-sites-router` **already existed**
> (id `7796166a490d487bbad8583f24e3c7b6`); only the DNS record was missing. I had
> told the owner both were missing. **`dig` cannot tell a missing DNS record from a
> missing Worker route — both give you an empty answer.** Ask the API which one it
> actually is before naming a cause:
> `GET /zones/{zone}/workers/routes`.
>
> Created: `*.ugg2.com` A → `199.59.243.228`, proxied. Proven with two unseen
> subdomains — see PLAN §6a "RESOLVED". Claim (B) holds; route (ii) is dead work.

## Cloudflare over the API — the token, and what it can and cannot do

**Token: `~/.config/cloudflare/token`** (mode 0600, expires **2026-09-30**). Never
echo it; read it inline. Every call below is `-H "Authorization: Bearer $TOKEN"`.

```bash
TOKEN=$(tr -d '\n\r' < ~/.config/cloudflare/token)
curl -sS https://api.cloudflare.com/client/v4/user/tokens/verify \
  -H "Authorization: Bearer $TOKEN"          # status:active + expires_on
```

**It reaches 36 zones** — the whole estate, not just this lane's two. There is no
per-zone fence, so **name the zone id explicitly in every call** rather than
looping over `/zones`.

**Measured permissions, 2026-08-02** — the gaps are the useful part:

| endpoint | result |
|---|---|
| `/zones`, `/zones/{z}/dns_records` (GET/POST/PATCH) | ✅ works |
| `/zones/{z}/pagerules` (GET/POST) | ✅ works |
| `/zones/{z}/workers/routes` (GET) | ✅ works |
| `/zones/{z}/settings/{ssl,always_use_https}` | ✅ works |
| `/zones/{z}/rulesets*` — **Redirect Rules** | ❌ `code 10000 Authentication error` |

**So the modern Redirect Rules API is NOT available to this token, but Page Rules
are — and a Page Rule "Forwarding URL" does the same 302.** That is the workaround;
it is why the redirect below was created as a Page Rule and not as a Redirect Rule.
Free plan allows 3 Page Rules per zone. If you need Rulesets, the token needs
*Zone → Config Rules/Transform Rules → Edit* adding.

**Gotcha:** `comment` on a DNS record is capped at **100 characters** — over it you
get `code 9313` and the record is **not created**. It reads like a warning; it is a
hard failure.

### The 302 + holding DNS, as actually applied (2026-08-02)

```bash
Z=746f81e606d259495b40e40a5316afb7        # webdesign.uk
# 1. redirect FIRST — closes the exposure before DNS is touched
curl -sS -X POST "https://api.cloudflare.com/client/v4/zones/$Z/pagerules" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data '{
    "targets":[{"target":"url","constraint":{"operator":"matches","value":"*webdesign.uk/*"}}],
    "actions":[{"id":"forwarding_url","value":{"url":"https://webdesign.co.uk/","status_code":302}}],
    "status":"active","priority":1}'
# 2. then point the origin somewhere unroutable
curl -sS -X PATCH ".../zones/$Z/dns_records/3f0570fb2f0f45b9979b61779745e8fa" \
  ... --data '{"type":"A","name":"webdesign.uk","content":"192.0.2.1","proxied":true,"ttl":1}'
```

**Order matters and the reason is not cosmetic.** Redirect first, DNS second: at no
point is `webdesign.uk` a hostname that neither redirects nor resolves. Do it the
other way and there is a window where the domain is simply broken. Doing the
redirect first also closes the idea.uk exposure *immediately*, since a forwarding
Page Rule is answered at the edge and the origin is never contacted.

`*webdesign.uk/*` covers the apex, `www`, and every path in one rule — verified on
`/` and `/checkout`.

## Ordering the box: the SSH key and the two DNS questions (2026-08-04)

### The key to create (owner's admin key — paste the .pub into the order form)

```bash
ssh-keygen -t ed25519 -a 100 \
  -C "owner webdesign-box admin 2026-08" \
  -f ~/.ssh/webdesign_box_ed25519
```

- **ed25519**, not RSA — modern default, small, fast; Mythic accept it.
- **Set a passphrase** — this is a human admin key to a box that will hold
  customer transcripts and (later) Stripe test config.
- Paste **only the `.pub`** into the order form; the private half never leaves
  your machine. Then:

```
# ~/.ssh/config
Host webdesign-box
  HostName <ip-or-name Mythic gives you>
  User root
  IdentityFile ~/.ssh/webdesign_box_ed25519
  IdentitiesOnly yes
```

**This is ONE of THREE keys in the design — do not conflate them:**

| key | created by | created when | lives |
|---|---|---|---|
| admin login (this one) | owner, now | at order | owner's machine |
| GitHub **deploy key** for `vm-sites` pulls | `provision-pullsync.sh` **on the box** | Phase 2 — the script pauses and asks you to add it on GitHub | `/var/lib/sitesync/.ssh/` |
| (optional) a key for Claude sessions to reach the box | later, if wanted | when delegating box work | this machine |

Do **not** pre-create the deploy key — the script generates it on the box so the
private half never transits anywhere (and its `GIT_SSH_COMMAND` handling exists
because of the `bugs_open/016` HOME trap; leave it to the script).

### Forward DNS: give the box an infrastructure name, never the service name

The order form asks for a hostname. Pick an infrastructure label
(`webdesign-box1`, or under any zone you own). **Never point `webdesign.uk` (or
any service domain) at the box** — with the tunnel there is no inbound to reach,
service DNS is written by `cloudflared` at cutover (Phase 6), and a service
record aimed at an origin is exactly the exposure class the 07-31 incident came
from. The box's own name is a label for you and for `known_hosts`; nothing in
the serving path uses it.

### Reverse DNS: not needed — set it to the hostname if the form offers it

Reverse DNS (PTR) matters for **outbound SMTP deliverability**, and this box
sends no mail: enquiries are the customer's own mail client today, and any
future notification goes over an HTTPS email API, not port 25. Nothing in the
web path, the tunnel, GitHub pulls, Anthropic or Stripe calls ever consults the
PTR. If Mythic's form offers reverse DNS, set it to match the hostname —
tidiness, zero cost — and move on.

## Phase 4 — the chat service: build, deploy, verify

Source lives in `box/chat-service/` (this lane's own directory in the repo).
Adapted from `idea.uk/golang_files/` (the proven reference: raw stdlib
`net/http` calling Anthropic directly, no SDK dependency — the box has no Go
toolchain, so this is cross-compiled and shipped as a static binary).

### Build and deploy sequence

```bash
cd docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/box/chat-service
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o webdesign-chat .

cd ..  # box/
scp -i ~/.ssh/webdesign_box_ed25519 chat-service/webdesign-chat webdesign-chat.service webdesign.uk.nginx \
  root@webdesign.vs.mythic-beasts.com:/root/

ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com <<'REMOTE'
install -m 755 /root/webdesign-chat /usr/local/bin/webdesign-chat
mkdir -p /var/lib/webdesign-chat && chown www-data:www-data /var/lib/webdesign-chat
install -m 644 /root/webdesign-chat.service /etc/systemd/system/webdesign-chat.service
install -m 644 /root/webdesign.uk.nginx /etc/nginx/sites-available/webdesign.uk
nginx -t && systemctl daemon-reload && systemctl reload nginx
systemctl enable --now webdesign-chat.service
REMOTE
```

**Re-deploying a new binary** (after a code change): same `scp` + `install` +
`systemctl restart webdesign-chat.service` — no need to re-touch nginx or the
systemd unit unless those files changed.

### The chat service env file

`/etc/webdesign-chat.env` (600, root:root) — created once by hand (never
committed; see HANDOFF_2026-08-09 §3 for the key-creation walkthrough), holds:

```
ANTHROPIC_API_KEY=sk-ant-...          # scoped Workspace key, workspace "webdesign-uk-chat"
                                      # ⚠ a workspace key is NOT a separate BUDGET — see
                                      # "Moving the chat to a different API budget" below
CONTACT_EMAIL=webdesign@contactforsales.com
CONTACT_PHONE=+44 (0) 7934 524 911
```

Both `CONTACT_EMAIL`/`CONTACT_PHONE` are load-bearing, not decorative: `main.go`
refuses to start without at least one — the fail-closed message (turn cap and
daily ceiling both route to it) needs somewhere real to point. **Proven live
2026-08-09**: the service correctly refused to boot 3 times in a row
(`journalctl -u webdesign-chat`) until the contact lines were added — the
startup guard is not theoretical.

Optional overrides (defaults are engineering judgement, not owner-confirmed
figures — see "sizing the daily ceiling" below):
```
MAX_TURNS_PER_CONVERSATION=20   # default in code
DAILY_SPEND_CEILING_USD=10.00   # default in code
```

### Moving the chat to a different API budget (the key swap)

**Run `box/swap-chat-api-key.sh`. Do not hand-edit the env file.**

```bash
box/swap-chat-api-key.sh --status   # read-only, no key needed — which key is live?
box/swap-chat-api-key.sh --check    # prompt for a key, TEST it against the API, write nothing
box/swap-chat-api-key.sh            # prompt, test, back up, write, restart, verify

# ...or take the key from a file holding a bare key or one `NAME=value` line:
box/swap-chat-api-key.sh --check --from-file ~/.config/anthropic/<file>
box/swap-chat-api-key.sh         --from-file ~/.config/anthropic/<file>
```

**Prefer `--from-file` over hand-piping the file in.** The key usually already lives in a
file, and the alternative is building a `cut | tr | ssh` pipeline at the prompt — which
puts the extraction outside anything reviewed and is exactly the shape that ends up in a
shell history or a transcript. `--from-file` strips a `NAME=` prefix if present, prints
only the fingerprint, and keeps every line that touches a key inside one audited file.
`[MEASURED 2026-09-02]` the swap was run this way: `API_KEY=` prefix stripped, preflight
200, one line rewritten, `RUNNING process: cd3e51a196a7`.

**The thing to understand first: a different KEY is not a different BUDGET.** The
Anthropic usage limit belongs to the ORGANISATION, not to the key. `[MEASURED
2026-08-31]` the chat's key (`c3358af6406c`) is already a different key from the fleet's
(`79eafe5d414e`, cluster secret `personae-default-secrets`), and the fleet still spent
the chat dark twice — `anthropic 400 ... "You have reached your specified API usage
limits"` on 2026-08-27 and 2026-08-30. **So a swap only buys anything if the new key is
on a different ACCOUNT.** The script's preflight will happily accept a valid key whose
account is already capped; it says so explicitly (`400` naming "usage limits") rather
than letting you believe you have fixed something.

**Keys are never printed, here or anywhere. The currency is a fingerprint** — twelve hex
of sha256 over the key:

```bash
printf %s "<the key>" | sha256sum | cut -c1-12
```

Equal digests mean the same key, and nothing secret is revealed, so fingerprints are
safe in chat, docs and commit messages. The box, the script and the service's own
journal all speak that one number. **This is how a session that may never read a key can
still answer "is the chat on the separate budget?"**

**Why not just edit the file?** All three hand steps fail quietly:

| hand step | how it fails silently |
|---|---|
| edit the key | a mistyped key does NOT stop the service — `main.go` only checks it is non-EMPTY. systemd says `active`, `/health` says 200, and every visitor gets the fail-closed contact line: **the same symptom as a usage-limit outage, from a different cause** |
| restart | forget it and NOTHING changed — systemd reads `EnvironmentFile` once, at start. Worse, every check that reads the FILE then agrees with the file |
| read the journal | a failed restart leaves the shopfront with no chat at all, and the old value is gone |

The script preflights (one token) and **writes nothing on a non-200**; backs up
`/etc/webdesign-chat.env` and **restores it automatically** if the unit does not come
back; rewrites exactly one line or refuses; and then asks the RUNNING PROCESS what it
holds rather than re-reading the file.

> **The running-process check needs the fingerprint line, added to `main.go` 2026-08-31.
> A binary older than that reports `RUNNING process: unknown` and the script falls back
> to the env file — which is exactly the weaker check the whole section is warning
> about.** Roll it with `make box-release` (from committed HEAD) and the journal answers
> for itself:
> ```bash
> ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
>   "journalctl -u webdesign-chat --no-pager -o cat | grep 'api key fingerprint' | tail -1"
> ```

**The last step is the owner's and it is the only end-to-end proof**: send one message
through the chat on the live site, then look at the NEW account's usage in the Console.
The service answering proves the key works; **the usage appearing against the new
account is what proves the BUDGET moved.**

⚠ **The second-site recipe below inherits whatever this file holds** — it `grep`s
`ANTHROPIC_API_KEY` out of `/etc/webdesign-chat.env` — so after a swap every new site's
chat silently joins the new account's budget. That is usually what you want; say which
budget you are joining when you provision, because otherwise a `grep` is making the
decision.

### Sizing the daily ceiling

> **SET TO $1.50/day on 2026-09-02 (owner's instruction), and the REASON is the shape to
> keep, not the number.** The guards must NEST: the chat's own daily ceiling × 30 has to
> sit **under** the account's monthly cap, or the inner brake is decorative and the outer
> one bites first. After the budget swap the chat sat on an account capped at **$55/month**
> while its own ceiling was **$10/day ≈ $300/month** — so the account limit would always
> have bitten first, and **an account limit failing closed presents as the "contact us
> directly" line**, i.e. the exact outage the swap was done to prevent, re-created from
> the other end. $1.50/day ≈ $45/month clears the cap with room.
>
> It is still enormously generous against real traffic: `[MEASURED 2026-09-02]` the chat
> spends **$0.0036/day** and **$0.286 across its entire life** since launch, so $1.50 is
> ~1,500 conversations a day.
>
> ```bash
> # env var only — no rebuild. Validate as main.go will (it log.Fatalf's on an
> # unparseable or <= 0 value, so a typo takes the chat DOWN, not degraded):
> #   back up -> write ONE line -> assert the key's fingerprint is unchanged ->
> #   restart -> restore automatically if the unit does not come back.
> # Then read the RUNNING service's own statement, never the file:
> ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
>   "journalctl -u webdesign-chat --since '-2 min' -o cat | grep 'sitechat on'"
> # expect: sitechat on 127.0.0.1:8081 (max_turns=20, daily_ceiling=$1.50)
> ```
>
> ⚠ **Do NOT verify a budget change on the Console's billing page.** It rounds to cents
> and the chat spends fractions of a penny a day — it printed `$0.00` after the swap and
> read as a failed change (`WRONG_CALLS.md` 2026-09-02). Use the key's **`Last used`**
> timestamp, the Analytics/Usage view (tokens), or the box's own ledger:
> `/var/lib/webdesign-chat/state.json` → `daily_spend_usd`.



Haiku 4.5: $1.00/$5.00 per MTok (input/output). Measured live on the first real
call through the tunnel (2026-08-09): 372 input + 11 output tokens = **$0.000427**
for one turn. A full conversation of several turns is comfortably under a cent.
The code default of **$10.00/day** is a wide engineering margin (≈2,000+ turns)
— tighten it once real traffic volume is known; it's an env var, no redeploy of
code needed, only a `systemctl restart`.

### Two-network CF-Connecting-IP proof

**Partial proof done 2026-08-09; full two-network proof still owed.** Fired one
real message through `https://preview.webdesign.uk/api/chat` and read back
`client_ip` from `transcripts.jsonl` on the box:
`2a02:c7c:f61f:ac00:f819:c606:416b:1535`. Independently confirmed via
`curl -6 https://api64.ipify.org` from the SAME machine: **identical value.**
This proves the header is carrying a genuine externally-observed address, not
a stuck constant, not the tunnel's own loopback, not cloudflared's address —
any of those failure modes would have produced a mismatch here, and none did.

**What's still owed**: a SECOND request from a genuinely different network
(e.g. a phone on mobile data) must log a DIFFERENT `client_ip` — this is the
`bugs_open/139`-shaped check (`count(DISTINCT ip) > 1`), and one network alone
cannot rule out every failure mode (e.g. if `clientIP()` somehow always echoed
the FIRST caller's own address back for every visitor, one tester would still
see a match). Two curls from two networks, five minutes, next time someone
has a second connection handy:
```bash
# from network A
curl -s -X POST https://preview.webdesign.uk/api/chat -H "Content-Type: application/json" -d '{"message":"a"}'
# from network B (phone on mobile data, different wifi, anything else)
curl -s -X POST https://preview.webdesign.uk/api/chat -H "Content-Type: application/json" -d '{"message":"b"}'
# then on the box:
grep -o '"client_ip":"[^"]*"' /var/lib/webdesign-chat/transcripts.jsonl | sort -u
# must show TWO distinct values, not one
```

### Verify after any deploy

```bash
curl -s https://preview.webdesign.uk/health   # {"status":"ok"}
curl -s -X POST https://preview.webdesign.uk/api/chat -H "Content-Type: application/json" -d '{"message":"hello"}'
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'tail -3 /var/lib/webdesign-chat/requests.jsonl /var/lib/webdesign-chat/transcripts.jsonl'
```

### Running the test suite (including the mutation proofs)

```bash
cd box/chat-service && go build ./... && go vet ./... && gofmt -l . && go test ./... -v
```

`chat_test.go` proves the turn cap and spend ceiling gates by MUTATING them
(neutralizing the condition, confirming the test fails, reverting) — not just
by observing the tests pass. See the file's own comments for why a passing
test alone isn't proof (memory: "a mutation that PASSES usually hit a guard in
series" — these two gates were checked in isolation, each proven to be the
thing that stops the call, not a downstream guard catching it anyway).

## Pushing a hand-verified page edit to `vm-sites` without the dispatch mechanism

Used 2026-08-09/10 to recover from a production incident and to add
`chat-input-box` — `bugs_open/239` found the normal work-item drive-loop
dispatch can silently no-op or misfire, so for a small, deterministic,
already-verified content change (no LLM needed), pushing directly is safer
than routing through it.

```bash
cd /home/ant/projects/vm-sites
git pull --ff-only origin main            # fast-forward only; this repo is shared, never force
# edit the ONE file — e.g. via the Edit tool, reading it first
git add webdesign.uk/contact.html         # only if it's a new path; existing files don't need add
git commit webdesign.uk/contact.html -m "..."   # explicit pathspec — same discipline as agentchassis itself
git push origin main
```

**Before committing, verify what you're about to push is actually correct** —
don't trust your own edit blindly:
- If restoring known-good content: diff against the last good git commit for
  that file (`gh api repos/gqls/vm-sites/commits --paginate -q '.[] |
  select(.commit.message | test("<filename>")) | "\(.sha) \(.commit.author.date)"'`
  to find it), and check the specific fields that matter (e.g. `grep -o
  'hero-contact\.jpg\|hero\.jpg'` for the image-binding landmine) — not just a
  whole-file byte match, since Cloudflare's own edge rewriting (email
  obfuscation) means the SERVED page will never byte-match the committed
  source exactly, and that's expected, not a bug.
- If adding new hand-rendered content: render the component's
  `content_components.html_template` by substituting its `content_data`
  values in place of each `{{.field}}`, HTML-escaping as you go
  (`&`→`&amp;`, `<`→`&lt;`, `>`→`&gt;`, and `"`→`&#34;` inside attribute
  values) — this only works safely when the template has no `{{if}}`/`{{range}}`
  blocks; check the template first.

**After pushing**, force the box to pick it up rather than waiting up to 5
minutes for the timer:
```bash
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'systemctl start sitesync.service'
```
Then verify at the served artefact (`curl` the page, grep for what changed),
and re-run `verify_served_site.sh` in full — a targeted check can miss a
regression elsewhere on the same page.

**If a write to this path is refused by the permission classifier**: this
happened after a prior incident on the same file in the same session. Try
the plain git workflow above once (not `gh api -X PUT`, not a raw `cp`) — if
still refused, stop and ask rather than trying further mechanisms; see
`HANDOFF_2026-08-10_continue_here.md` §2 for the full account.

### Activating live facts on the chat bot (CHAT-010, after the core-manager image rolls)

The endpoint (`GET /api/v1/site-facts/:domain`) ships in core-manager
**v1.0.1292**. It 404s on any earlier image. The box binary already carries the
consumer (deployed 2026-08-12, running in legacy/compiled-in mode). To switch
the bot to live DB facts, once v1.0.1292 is rolled:

1. **Set the shared token on core-manager** (its env; whole-fleet release owns
   this, or set it directly and restart core-manager for a one-off):
   `SITE_FACTS_TOKEN=<a long random string>` — with it UNSET the endpoint
   fail-closes (401 to everyone), so this must be set for the endpoint to work.
2. **Set the box unit's env** (`/etc/webdesign-chat.env`, 600 root:root):
   ```
   FACTS_URL=http://<core-manager ClusterIP>:8088/api/v1/site-facts/webdesign.uk
   FACTS_TOKEN=<the same random string as step 1>
   ```
   The ClusterIP is stable per-service but not forever — resolve it fresh:
   `kubectl -n ai-persona-system get svc core-manager -o jsonpath='{.spec.clusterIP}'`.
   > **UPDATED 2026-08-15: use the DNS name — it is live config now, not an
   > aspiration.** The box's system resolver routes `*.cluster.local` to
   > kube-dns over the tunnel (wg0 PostUp: `resolvectl dns wg0 10.21.0.10`
   > + routing domain `~cluster.local`; runtime-applied AND durable in
   > `/etc/wireguard/wg0.conf`). `FACTS_URL` uses
   > `core-manager.ai-persona-system.svc.cluster.local:8088` since 17:25Z
   > (journal: `fetched 15 facts` + `live mode` on the named URL), and the
   > nginx `/stripe/webhook` upstream names auth-service the same way. The
   > earlier claim here that the name "also works" was true only of
   > `dig @10.21.0.10` — the SYSTEM resolver (what the Go binary uses) had no
   > route to kube-dns until the resolvectl config landed. Probe note: the
   > relay authenticates via the `X-Facts-Token` header, NOT
   > `Authorization: Bearer` (facts.go:109) — a Bearer probe 401s and reads
   > exactly like a dead relay.
3. `systemctl restart webdesign-chat` and check the log: it should print
   `facts: live mode, relay=…` then `facts: fetched N facts from relay`. If it
   prints a fatal about the relay being unreachable + no cache, the endpoint
   isn't rolled yet or the token/URL is wrong — the bot **refuses to start**
   rather than serve stale compiled facts (by design).

**Prove it end to end** — the whole point of the exercise:
```bash
# 1. change a fact in the DB (a harmless reversible one)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "UPDATE site_specs SET data = jsonb_set(data,'{facts}', <...edited facts...>) WHERE site_id='1fcfa4f3-...' AND aspect='evidence_base' AND is_current;"
# 2. wait up to 5 min (refresh timer) or restart the box unit, then ask the bot
curl -s -X POST https://preview.webdesign.uk/api/chat -H 'Content-Type: application/json' \
  -d '{"conversation_id":"","message":"what is the price?"}'
# the reply reflects the DB change WITHOUT rebuilding or redeploying the box binary.
```

**Tunnel health** (`wg show` is not enough — see the LANDMINE):
```bash
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'curl -s --max-time 10 http://10.21.0.10:53 >/dev/null; \
   curl -s --max-time 10 http://<core-manager clusterIP>:8088/health'
# a real /health body = the tunnel forwards; a timeout with a fresh wg handshake = ip_forward, check the pod.
```

## Restoring or rotating the facts-relay token (2026-08-13)

The token lives in TWO places that must hold the same value: `SITE_FACTS_TOKEN`
in `personae-platform-secrets` (cluster side) and `FACTS_TOKEN` in
`/etc/webdesign-chat.env` (box side). **The cluster copy is terraform-managed —
never `kubectl patch` it**: `make release` re-applies
`047-base-configs/main.tf` and deletes any key not declared there (that is
exactly how the 2026-08-13 outage happened). The value belongs in
`deployments/terraform/environments/production/uk001/047-base-configs/terraform.tfvars.secret`
as `site_facts_token = "..."` (gitignored); `main.tf` maps it into the secret.

```bash
# 1. set/replace site_facts_token in terraform.tfvars.secret, then:
cd deployments/terraform/environments/production/uk001/047-base-configs
terraform init -input=false && terraform apply -auto-approve -var-file=terraform.tfvars.secret
# gotcha: a bare "Error: Unauthorized" from plan/apply usually means the dir
# was never re-inited (run init) — only after that suspect the 3-day kubeconfig token.

# 2. env vars are read at pod start, so the new value needs a restart:
kubectl -n ai-persona-system rollout restart deploy/core-manager
kubectl -n ai-persona-system rollout status deploy/core-manager

# 3. only if the box-side value changed too: edit FACTS_TOKEN in
#    /etc/webdesign-chat.env on the box, then
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com 'systemctl restart webdesign-chat'

# 4. verify at the artefact (the refresh timer is 5 min):
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'journalctl -u webdesign-chat -n 20 --no-pager | grep facts'
# expect "facts: fetched 15 facts from relay" (+ "facts: live mode" after a bot restart)
```

## One binary, several sites: `sitechat` + `sitechat@<domain>` (2026-08-16, PLAN_2026-08-11 step 5)

The chat service is now a per-site parameterised binary, `/usr/local/bin/sitechat`,
built from `box/chat-service/` and run by two unit shapes on the box:

| unit | reads | data | for |
|---|---|---|---|
| `webdesign-chat.service` (kept, repointed) | `/etc/webdesign-chat.env` | `/var/lib/webdesign-chat` | webdesign.uk — proven store + journal history stay put |
| `sitechat@<domain>.service` (template) | `/etc/sitechat/<domain>.env` | `/var/lib/sitechat/<domain>` | every other site on this box |

Both run the SAME file. The old `/usr/local/bin/webdesign-chat` is kept as
`webdesign-chat.bak-20260815b` (rollback: repoint `ExecStart`, `daemon-reload`, restart).

> **⚠ CORRECTED 2026-08-18 — this paragraph used to say "`md5sum` on the box must
> equal the local build's", and that check CANNOT do the job it looks like it does.**
> Measured: rebuilding the **exact commit** behind the then-live binary
> (`84202f061`) produced md5 `65da9971` and 9381552 bytes, against the box's
> `f07fb146` and 9381544. Same source, different digest, eight bytes apart. These
> builds are **not byte-reproducible across build environments**, so an md5
> comparison proves only that the box holds the file you just pushed in THIS
> session. It cannot answer "which commit is the box running", which is the
> question that matters, and a mismatch looks identical to a failed deploy.
>
> **Ask the binary instead.** Since `434d2b64b` it stamps its own commit and logs
> the same line the backend fleet uses:
> ```bash
> ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
>   "journalctl -u webdesign-chat --since '-5 min' -o cat | grep 'build provenance'"
> ```
> `make box-verify` does both halves: md5 for "the file arrived", provenance for
> "the running service is this commit". A binary built outside `make box-build`
> says `unstamped` rather than printing an empty commit.

**The whole scp/systemctl recipe below is superseded by `make box-release`** (added
2026-08-18 at the owner's request). It builds from committed HEAD via `git archive`
— the hand recipe built from the working tree and could ship another session's WIP.

### Per-site parameters (the env file)

```
ANTHROPIC_API_KEY=...                       # required
SITE_DOMAIN=<domain>                        # required in live mode; cross-checked vs the relay's `domain`
SITE_DESCRIPTION="<owner-confirmed phrase>" # required in live mode; renders "intake assistant for <domain>, <desc>."
FACTS_URL=http://core-manager.ai-persona-system.svc.cluster.local:8088/api/v1/site-facts/<domain>
FACTS_TOKEN=...                             # SITE_FACTS_TOKEN — same value fleet-wide, terraform-owned; grep/cut it out of /etc/webdesign-chat.env, NEVER `source` that file
CONTACT_EMAIL=... / CONTACT_PHONE=...       # at least one; what the four controls fail closed TO
BIND_ADDR=127.0.0.1:<port>                  # loopback ONLY for new instances (see below)
MAX_TURNS_PER_CONVERSATION / DAILY_SPEND_CEILING_USD   # optional engineering defaults
```

`SITE_DESCRIPTION` is **owner copy**, like the contact line — the bot introduces
itself with it. `sites.company_name/tagline/email/phone` are the operator's starting
point, not the answer. **The binary refuses to start live mode without SITE_DOMAIN and
SITE_DESCRIPTION** — no default, so an instance can never fall back to another
site's identity. It also refuses when the relay's `domain` field ≠ SITE_DOMAIN
(proven live 2026-08-16: noted's identity + webdesign's URL → `refusing another
site's facts`, no listener) and on zero facts.

### Rolling the shared binary (all instances on the box)

```bash
cd docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/box/chat-service
GOPROXY=off GOTOOLCHAIN=local go test . -count=1 && \
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o sitechat .
scp -i ~/.ssh/webdesign_box_ed25519 sitechat root@webdesign.vs.mythic-beasts.com:/root/
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'install -m 755 /root/sitechat /usr/local/bin/sitechat && systemctl restart webdesign-chat.service "sitechat@*.service" 2>/dev/null; \
   sleep 2; journalctl -u webdesign-chat -n 3 --no-pager -o short-iso; ss -ltnp | grep sitechat'
```
Expected journal: `facts: fetched N facts from relay` → `facts: live mode, site=<domain>, …`
→ `sitechat on 127.0.0.1:<port>`. **`ss` must show `127.0.0.1:` binds only** — the pre-08-16
build bound `*:8081` (noted's nginx names it as the pattern not to copy); if you ever see
`*:` or `0.0.0.0:` here, BIND_ADDR is missing from that unit/env.

### Provisioning a second site's instance (answering a `backend_provision` item)

The chassis raises ONE `backend_provision` work item per (tool, site) when
`deploy_tool_to_site` forks a `requires-backend` tool (TL-043; inert until the chassis
roll after `51c33f482`). Its `spec` carries `domain`, `facts_relay_url`, the token
secret's NAME, `facts_count_at_deploy`, `contact_fact_at_deploy` (a snapshot — re-read
the live fact) and a pointer here. To answer it:

```bash
D=<domain>; PORT=<unused loopback port, e.g. 8084>
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com <<REMOTE
set -e
install -d -m 700 /etc/sitechat; install -d -o www-data -g www-data -m 700 /var/lib/sitechat/$D
TOKEN=\$(grep '^FACTS_TOKEN=' /etc/webdesign-chat.env | cut -d= -f2-)
KEY=\$(grep '^ANTHROPIC_API_KEY=' /etc/webdesign-chat.env | cut -d= -f2-)
cat > /etc/sitechat/$D.env <<EOF2
ANTHROPIC_API_KEY=\$KEY
SITE_DOMAIN=$D
SITE_DESCRIPTION="<OWNER-CONFIRMED PHRASE>"
FACTS_URL=http://core-manager.ai-persona-system.svc.cluster.local:8088/api/v1/site-facts/$D
FACTS_TOKEN=\$TOKEN
CONTACT_EMAIL=<owner-confirmed>
BIND_ADDR=127.0.0.1:$PORT
EOF2
chmod 600 /etc/sitechat/$D.env
systemctl enable --now sitechat@$D.service; sleep 2
systemctl is-active sitechat@$D.service; journalctl -u sitechat@$D -n 3 --no-pager -o short-iso
REMOTE
```
Then add to that site's nginx vhost (`/etc/nginx/sites-available/$D`, mirroring
webdesign.uk's `location /api/chat` block incl. its `limit_req` zone) a
`proxy_pass http://127.0.0.1:$PORT;`, `nginx -t && systemctl reload nginx`, and
prove Journey A through the public edge: `curl -X POST https://<host>/api/chat -d
'{"message":"What does this cost?"}'`. Only THEN mark the item complete. **Do not
pre-stage an env file with placeholder copy** — a disabled unit someone starts later
would introduce the bot with the placeholder.

Gate before you start: `curl -s -o /dev/null -w '%{http_code}\n' -H "X-Facts-Token: $TOKEN"
<facts_relay_url>` **and** the body's `facts` length. **200 is not enough** — noted.co.uk
returns 200 with `facts: []` (measured 2026-08-16), and the binary refuses on zero
facts. Facts are attested by the site's owner/lane (register trail), never by the
provisioner.

### Transient proof runner (no unit, dummy key, nothing persisted)

For "does this site's instance come up" without enabling anything and with NO chance
of an LLM call:
```bash
ANTHROPIC_API_KEY=proof-not-a-real-key CONTACT_EMAIL=proof@example.invalid \
SITE_DOMAIN=<domain> SITE_DESCRIPTION="proof" FACTS_URL=<relay>/<domain> FACTS_TOKEN=$TOKEN \
DATA_DIR=/tmp/sitechat-proof BIND_ADDR=127.0.0.1:18082 timeout 6 /usr/local/bin/sitechat
```
Success = `fetched N facts` + `live mode, site=<domain>` + a `/health` 200 on 18082
within 3s. Used 2026-08-16 for relojistas.com (13 facts, came up), noted.co.uk (zero
facts, refused), and the mismatch case (refused).

## Editing `evidence_base` safely: the reconstruction guard, and proving it can fail (2026-08-21)

Every register edit here supersedes a row. The guard that matters is not "did the new
value land" (it always does) but "did ONLY the intended edits land". Assert it by
**reconstruction** — apply the same `replace()` chain to the superseded text and demand
equality:

```sql
expect := replace(replace(pwb, '<anchor 1 old>', '<anchor 1 new>'),
                          '<anchor 2 old>', '<anchor 2 new>');
IF expect IS DISTINCT FROM wb
  THEN RAISE EXCEPTION 'writer_block is not the old text plus exactly the named edits'; END IF;
```

**Gotcha:** asserting *"the new substring is present"* cannot see an unintended THIRD
edit riding along in the same transaction. Reconstruction can. Pair it with an
exactly-once check on each anchor in the OLD text, because `replace()` is happy to fire
twice and silently:

```sql
n := (length(pwb) - length(replace(pwb, '<anchor>', ''))) / length('<anchor>');
IF n <> 1 THEN RAISE EXCEPTION 'anchor occurred % times, expected 1', n; END IF;
```

**Gotcha, and it is the one that bites:** a reconstruction guard proves no *unintended*
edit. It cannot prove the intended ones mattered. Add OUTCOME guards next to it — the
retired string is gone, the attested string is present — and see the mutation recipe
below for why they are not redundant.

### Proving the guard can fail, before you apply anything

Rolled-back transactions against the real row. **A guard that has only seen the state it
was written for proves nothing.**

```bash
SP=<scratch dir>; F=<your SQL file>
# strip everything after COMMIT, append ROLLBACK, inject one mutation per variant
python3 - "$F" "$SP" <<'PY'
import sys; src=open(sys.argv[1]).read(); body=src.split("\nCOMMIT;")[0]
open(sys.argv[2]+"/clean.sql","w").write(body+"\nROLLBACK;\n")
# ... variants: an extra edit; the intended edit missing its anchor; a fact mutated
PY
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < $SP/clean.sql
```

**Read the clean run's output, do not just check for absence of ERROR.** It must print
`BEGIN` → `INSERT 0 1` → `DO` → `ROLLBACK`. A guard that passes because the INSERT
matched nothing looks identical at the grep.

**Isolating outcome guards.** Run them against the real PRE-fix row with the
reconstruction guard removed and every other outcome guard deleted (not commented, and
not neutered by rewriting the first line of a multi-line `IF` — a two-condition
`IF a OR b` neutered to `IF false OR b` still fires, which makes every variant report the
same earlier error and look like a working test). Delete the whole `IF … END IF;` block.

### claimscan probe sets: BOTH halves, or the run is worthless

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
  "SELECT ss.data::text FROM site_specs ss JOIN sites s ON s.id=ss.site_id
    WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current" > eb.json
# TSV: name <TAB> slot <TAB> base64(html) <TAB> page_type   (the 4th column matters)
go run ./cmd/claimscan -evidence eb.json -components probe.tsv
```

**Gotcha:** must-pass candidates alone cannot distinguish a clean scan from a dead one.
Put must-BLOCK controls in the same run. On 2026-08-21 that is what proved the live
`writer_block`'s own turnaround instruction was banned copy, and it is what caught a
replacement instruction that was itself banned (*"in one day"* — the negation guard
scans backwards only a short way, so a leading `Never` does not cover the third item of
a list).

**Scanning `writer_block` itself** is worth doing and nobody thinks to, because it is
prompt text rather than page text. Split it by paragraph so findings are locatable:

```bash
psql … -At -c "SELECT data->>'writer_block' FROM site_specs WHERE …" > wb.txt
python3 -c "
import base64,sys
for i,p in enumerate([x.strip() for x in open('wb.txt') if x.strip()]):
    print('wb_%02d\tbody\t%s\tservice' % (i, base64.b64encode(p.encode()).decode()))" > wb.tsv
go run ./cmd/claimscan -evidence eb.json -components wb.tsv
```

Expect legitimate hits: prohibitions quote their own banned tokens on purpose
(*"never write 'no refund'"*). Read each one rather than counting them.

## Transferring a sold domain out to the customer (2026-08-21) — UNTESTED, and say so

**Owner ruling 2026-08-21: a manual step per domain is acceptable for now.** Do not
automate this. Full mechanism, fees and the one open decision:
`DECISION_2026-08-21_domain_transfer_out_from_nominet.md`.

> **⚠ THIS PROCEDURE HAS NEVER BEEN RUN. Zero domains have been sold.** It is assembled
> from Nominet's published process and from this estate's existing EPP access, not from a
> completed transfer. **The first time it is used, correct it here** — and treat every
> step below as a claim to verify rather than a command to trust. The `[VERIFIED]` marks
> apply to the *published rules*, not to our having executed them.

### Before the first sale ever happens, prove the access still works

Do this once, well before a customer is waiting. All three can fail silently.

```bash
# 1. Credentials present? (existence only — never cat these)
ls -l ~/.config/nominet/epp-password ~/.config/nominet/credentials

# 2. The tag. `nominet-epp-ns-change.py`'s usage example says DESIGNCONSULT; the
#    domains_cloudflare_rollout RUNBOOK recorded "TAG still needed" on 2026-08-04.
#    RESOLVE IT FROM THE CREDENTIALS FILE, do not copy it from a doc example.
grep -o '^TAG=.*' ~/.config/nominet/credentials

# 3. A real login. See the gotcha below — a greeting proves nothing.
python3 docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/box/nominet-epp-ns-change.py \
  --tag <TAG> --domain <a domain we already hold> \
  --password-file ~/.config/nominet/epp-password        # dry-run by default
```

**Gotcha, and it has already cost this estate time:** Nominet serves the EPP **greeting
to any IP**, and login is refused unless your egress IP is allowlisted in Online Services
(Settings → EPP → IP addresses) *regardless of whether the credentials are right*. So a
successful TLS handshake and a 2,527-byte greeting are **not** a connectivity proof —
only a completed login is. **Pin to IPv4**: IPv6 gets a 94-byte brush-off where IPv4 gets
the full greeting, so an IPv6-first resolver makes a healthy path look broken.

**Gotcha:** the EPP password is 16 bytes on this machine and comes from a file or
`$NOMINET_EPP_PW`, **never argv** — argv is visible in `ps` to every process on the box.

### The sale itself: TWO Nominet operations, in this order

Changing the registrar and changing the owner are different things. Both are needed.
Doing them in the wrong order strands the domain.

1. **Registrant Transfer — change the recorded owner to the customer.**
   Nominet Online Services → **Registrant Transfer**. This one **cannot** be done over
   EPP or through our own systems; it is a registry operation. `[VERIFIED 2026-08-21 at
   Nominet's published fee schedule, NOT by having done it]` **£10+VAT** for a
   straightforward name change, **£20+VAT** change of type/company, **£35+VAT** where
   extra verification is required. Budget the fee into the £200; it is not a surprise.
2. **Release the tag to the customer's chosen registrar.**
   Ask the customer for their new registrar's tag first — *they* choose the registrar,
   which is the half the attested `domain_buy_once` fact says is theirs to do. Then
   change the IPS TAG to it. Free for us.

   **After step 1 the customer can do step 2 themselves** through their own Nominet
   account for ~£10+VAT. Offer that as the alternative; do not insist on it, and do not
   walk them through their registrar's side of it (`no_presales_service`).

### ⚠ From 9 FEBRUARY 2027 step 2 CHANGES COMPLETELY — rewrite this section that week

Nominet retires the IPS TAG transfer process and replaces it with a **Transfer
Authorisation Code**: we generate one and give it to the registrant, they hand it to the
gaining registrar, and the transfer completes immediately if the domain is unlocked.
Step 1 is unaffected. Formal notice 4 June 2026; transition 9 February 2027; portfolios
migrate to Dragon Domain Manager and Nominet moves to standard EPP at the same time.

**Check the detail by 2026-12-01** — https://registrars.nominet.uk/registry/dot-uk/faq/ —
because the useful consequence is a product change, not just a process one: a code can be
**pre-issued at handover and put in the delivery email**, which is the same "hand over
the thing rather than promise the action" shape as Phase 4's ZIP token.

### What to record after the first real transfer

The point of writing this before it is needed is that the first run corrects it. Record:
which fee actually applied; whether Nominet's UI matched these step names; how long each
step took; whether the customer self-served step 2 or we did it; and **anything that
succeeded while doing nothing**, which is this estate's most expensive failure shape.

---

## The wireguard egress fence (applied 2026-08-22) — what it is and how to verify it

The webdesign.uk box reaches the cluster as `peer_webdesignbox` on the **main** `wireguard`
deployment. That instance masquerades, and `allow-same-namespace` unions away
`database-access-policy`, so before this fence any peer could reach **every** service in
`ai-persona-system` including `postgres-clients:5432` — proven, with a closed-port control.

`deployments/kustomize/services/wireguard/base/networkpolicy-wireguard-egress.yaml` is a
deny-by-default **egress** policy on the wireguard pod. Egress-on-the-pod is the only
enforcement point that works: because of the masquerade, a policy keyed on a peer's
`10.13.13.x` address can never match, and would fail closed while looking like it worked.

**Allowlist, and it is evidence-based — every cluster upstream named anywhere in the box's
own config, nothing guessed:**

| destination | why | what breaks without it |
|---|---|---|
| kube-dns `:53` | both upstreams are proxied BY NAME | everything, presenting as "core-manager is down" |
| `core-manager:8088` | box nginx `location /c/` **and** the chat bot's facts relay (`box/chat-service/facts.go`) | `/c/`, and **the bot refuses to START** (its own stated design) |
| `auth-service:8081` | box nginx `/stripe/webhook` | **THE MONEY PATH** |
| `admin-dashboard:8080` | the owner's `laptop`/`phone` peers | his admin access over the VPN |

⚠ **Adding a peer that needs a new destination means adding it HERE.** Otherwise that peer
gets a timeout that looks exactly like the destination service being down.

### Verify the fence

Run from the wireguard pod — that is where every peer's traffic emerges after the
masquerade, so the pod's reach IS the peer's reach. **Both arms matter**: a bare `nc -z`
that is missing or busybox-limited reads the same as a blocked port, so the closed-port
control is what proves the instrument works.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=wireguard -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec -i "$POD" -c wireguard -- sh -s <<'PROBE'
probe(){ printf '  %-34s ' "$1"; if nc -z -w4 $2 $3 2>/dev/null; then echo "REACHABLE"; else echo "blocked/refused"; fi; }
printf '  %-34s ' "resolve core-manager"; (nslookup core-manager.ai-persona-system.svc.cluster.local >/dev/null 2>&1 && echo OK) || echo FAIL
probe "core-manager:8088   (MUST stay)"    10.21.127.41  8088
probe "auth-service:8081   (MUST stay)"    10.21.217.63  8081
probe "admin-dashboard:8080(MUST stay)"    10.21.171.225 8080
probe "postgres:5432       (MUST be gone)" 10.21.233.177 5432
probe "postgres:5433       (control)"      10.21.233.177 5433
PROBE
```

Expected: DNS `OK`; the three `MUST stay` rows `REACHABLE`; **postgres:5432
`blocked/refused`**; the control also `blocked/refused`. If the control ever reads
`REACHABLE`, the probe is meaningless — stop and fix the probe, not the policy.

⚠ The ClusterIPs above are **pinned literals**. Re-read them (`kubectl -n ai-persona-system
get svc`) if a service is ever recreated, or you will be probing nothing and calling it a
pass.

### Then verify the live services that actually ride the tunnel

A policy probe is not the product. These three go through the box and would each fail
differently if the allowlist were wrong:

```bash
# the bot — proves the facts relay still works (it fetches core-manager over the tunnel)
curl -sS -X POST https://preview.webdesign.uk/api/chat -H "Content-Type: application/json" \
  -d '{"message":"How long will my site take?"}'
# /c/ — still in-cluster only; ask core-manager directly, the box does not serve it on preview
CM=$(kubectl -n ai-persona-system get pods -l app=core-manager -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$CM" -- sh -c 'wget -q -S -O /dev/null http://127.0.0.1:8088/c/x 2>&1 | head -1'   # expect 200
# the site
curl -sS -o /dev/null -w '%{http_code}\n' https://preview.webdesign.uk/    # expect 200
```

⚠ **`https://preview.webdesign.uk/c/x` returns 404 and that is CORRECT, not a regression.**
The box's nginx `server_name` is `webdesign.uk www.webdesign.uk`, not `preview`, so `/c/` is
not served on the preview host at all. Do not "fix" this by widening the server_name — that
is exactly the public-exposure decision (D-A) that is still open.

### To roll it back

`kubectl -n ai-persona-system delete networkpolicy wireguard-egress-containment`. One
command, no pod restart, peers keep their tunnels. **This is the fence's best property: it
contains an internet-facing peer without re-keying anything, without touching the box, and
without any downtime on the money path** — unlike the separate-instance design in
`gauntlet_dead_cta/infra/wireguard_bastion.yaml`, which is still the better long-run shape
but needs a box-side cutover.
