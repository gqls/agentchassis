# Handoff — idea.uk contact-form spam + IP block list

> **CORRECTED 2026-07-14.** The original version of this handoff was wrong on three load-bearing
> points and would have sent the next session hunting in the wrong database. What it said, and what
> is actually true:
>
> | It said | Actually |
> |---|---|
> | The `/request` handler is "the chassis Go process" | It is the **standalone `idea` binary** (`module idea`, stdlib-only, zero deps) at `docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/`. Nothing to do with the chassis. |
> | The submissions are "almost certainly into a store we can query", and an earlier 0-row Postgres search was "a search miss, not a 'no data' result" | **The 0 rows were correct.** idea.uk has **no database**. Orders are a JSON file: `/var/lib/idea/orders.json` (`store.go:3-5`, `setup.sh:150`, `HANDOFF(13).md:8`). |
> | nginx already serves static pages and proxies the reserved tool paths | That is the **future** state. Today nginx proxies *everything* to `:8080`. The cutover has not happened — see `../idea_uk_vm_site/RUNBOOK_idea_uk_vm_site.md`. |
>
> **`spam_read.sql` is therefore void.** It searches `clients_db` for orders that live in a file on a
> VM. It will return the same orchestration-plumbing tables and find `ord_1783948426211007948` in
> none of them. **Discard it — do not extend it with more `ILIKE` needles.**
>
> The one thing the original got right, and it is the important one:
> *"Spammers rotate IPs, so a honeypot + timing check usually outperforms an IP list."*

## The problem
idea.uk's report-request form (the `/request` intake) is receiving spam: fake submissions with test
data and generated order ids, e.g.

    New idea.uk report request.
    Requester: test (test@test.com)
    Business: test
    Audience: test
    Notes: test
    Order id: ord_1783948426211007948

The owner wants (a) the existing spam removed and (b) an IP block list started.

## Where the data actually is
`/var/lib/idea/orders.json` on the Hetzner box `116.203.204.115`. A JSON file, no DB, no driver.
`store.go:3-5` says so outright: *"A JSON-file store (stdlib only) so the service runs standalone with
no DB driver. Production should swap in the chassis Postgres behind the same small method set"* — the
Postgres swap is explicitly **future**, not done.

Read it with SSH, not SQL:
```bash
ssh root@116.203.204.115 "python3 -c \"import json;d=json.load(open('/var/lib/idea/orders.json'));print(len(d))\""
```

## Two facts that constrain the whole fix

**1. The stored orders carry no IP.** The `Order` struct (`store.go:17-30`) is:
`ID, Name, Email, Domain, Audience, Assets, Status, Report, ReportHTML, ProviderSessionID, CreatedAt,
UpdatedAt`. **No `IP`, no `UserAgent`, no `Referer`.** So an IP block list **cannot be seeded
retroactively from the existing spam**. The only historical source of attacker IPs is the nginx access
log — `grep 'POST /request' /var/log/nginx/access.log*`, correlated on timestamp against `created_at`.

**2. nginx cannot currently see the real client IP.** idea.uk sits behind Cloudflare
(`HANDOFF(13).md:8`), but `setup.sh` never sets `set_real_ip_from <CF ranges>` + `real_ip_header
CF-Connecting-IP`. So `$binary_remote_addr` is a **Cloudflare edge IP**. Consequences:
- the existing `limit_req` zone (`setup.sh:86, 226, 299` — 10r/s burst 20) buckets **all** of
  Cloudflare's traffic per-edge rather than per-visitor, so it is far weaker than it looks;
- any nginx `geo`/`map` deny or fail2ban jail would ban **Cloudflare, not the spammer**.

This exact trap is already documented for the *other* box (`traffic_probe_runbook(12).md:314-316`) and
was never back-ported. **Fix it before attempting any IP-based blocking.** First confirm whether the
DNS record is actually proxied (orange) or DNS-only (grey) — that decides both whether this is live and
whether Cloudflare WAF/Turnstile is reachable as the blocking layer.

*(The Go app itself is unaffected: `clientIP()` takes the first XFF entry, which Cloudflare sets
correctly. App-level per-IP limiting would work today.)*

## What already exists and is simply unwired

| Thing | Where | Wired to |
|---|---|---|
| Per-IP sliding-window rate limiter (3/hr **and** 20/day, in-memory) | `audience_check.go:31-95` | **only** `/audience-check` |
| `clientIP()` — XFF-first, falls back to `RemoteAddr` | `audience_check.go:100-113` | **never called by `handleRequest`** |
| nginx forwards `X-Real-IP` / `X-Forwarded-For` | `setup.sh:229-230, 302-303` | live |
| nginx `geo`/`map` **allow**-list machinery (`WHITELIST_IPS`) | `setup.sh:73-88` | live — trivially invertible to a deny-list |
| fail2ban, installed and enabled | `setup.sh:126, 351-359` | **`[sshd]` jail only** — nothing watching HTTP |

`handleRequest` (`service.go:301-310`) has **no rate limit, no honeypot, no validation beyond
presence**, and discards `ParseForm`'s error. That is precisely why `test/test/test/test@test.com`
sails through.

## The order of work

1. **Restore the real client IP in nginx** (`set_real_ip_from` + `real_ip_header CF-Connecting-IP`).
   Nothing IP-based works until this is done.
2. **Harden `/request`.** Honeypot field + minimum time-to-submit + email-format check + length caps;
   reuse the existing `rateLimiter`; add an `IP` field to `Order` so a blocklist can be seeded *going
   forward*. Honeypot + timing beats an IP list.
3. **Remove the existing spam.** Signature: every field literally `test`, `test@test.com`. It is a
   JSON file, so there is no guarded `DELETE` — and `Store` has **no `Delete` method** (`persist`,
   `Save`, `Get`, `Update`, `ActiveCount`, `MarkEventSeen`, `AddSubscriber`). Stop the service, back
   up `orders.json`, filter, restart — reading out exactly what *would* be removed first.
4. **Only then** consider an IP block list, if the logs show the spam actually clusters on a few
   addresses. It may well not.

**Shipping:** the tool has no CI. Build and scp:
```bash
cd docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .
scp idea root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 'chmod 755 /opt/idea/idea.new && mv -f /opt/idea/idea.new /opt/idea/idea && systemctl restart idea'
```

## Working rules (unchanged)
Go not Python; British English; schema-first before any SQL; a 0-row result is not decisive until the
query itself is cleared — **but note that in this case it *was* decisive, and the original handoff
talked itself out of a correct answer**; structural over quick patches; reuse existing
functions before writing new ones; `logger.Info` not `logger.Debug`.

## Not part of this workstream
The chassis colour/imagery/page work is tracked in `HANDOFF_claude_code_continue.md`. The VM cutover
(which changes how `/request` is reached) is tracked in `../idea_uk_vm_site/`.

**Related but separate — worth its own thread.** Chassis-generated sites emit a contact form that posts
into a void: `apply_gap_plan_action.go:465` emits a `contact-form` section whose stored HTML is
`<form class="contact-form" action="/contact" method="POST">`, and generated sites are **static**, so
`POST /contact` resolves to nothing and **every submission is silently lost, fleet-wide**. idea.uk has a
deployed `/contact.html`. That is a *dead form* problem, not a spam problem. Whatever contact backend is
eventually built should be born with the honeypot and rate limit this handoff retrofits onto `/request`.
