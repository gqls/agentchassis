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
