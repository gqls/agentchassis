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

### Sizing the daily ceiling

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
