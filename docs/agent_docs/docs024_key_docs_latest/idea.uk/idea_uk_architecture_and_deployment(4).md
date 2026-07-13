# idea.uk — architecture, hosting, and deployment guide

Plain-language map of what we built, how it runs, how it connects to Stripe, and
how to deploy it — on idea.uk, on other sites, and inside the agent chassis.

Read top to bottom once; after that, the section you need is self-contained.

---

## 1. What we built (the pieces)

Five things, in one folder (`idea-go/`):

| File | Plain English |
|---|---|
| `engine.go` + `prompts.go` | **The brain.** Runs the ideation method: challenge the audience → generate ideas across four lenses → cut them against the free alternative → verify survivors with web search → score → rank. Talks to Anthropic (and optionally OpenAI) directly over HTTP. |
| `store.go` | **The memory.** Saves orders, paid/unpaid state, and which Stripe events we've already handled. Currently a JSON file; swappable for Postgres. |
| `billing.go` | **The till.** Creates Stripe payment links and verifies Stripe's payment confirmations. |
| `service.go` | **The front desk.** A small web server: takes report requests, lets you confirm or decline them, takes the Stripe payment confirmation, runs the brain, delivers the report. |
| `main.go` | **The switch.** Starts the web server, or runs the brain once from the command line with no server and no billing. |

Plus the customer-facing page (`idea_uk_fakedoor.html`) and the test
(`service_test.go`).

There is **one engine**, used two ways:

- **Internal** — you run the brain for your own domains, no payment. (`idea internal …`)
- **External** — a stranger requests a report on idea.uk, pays, and gets it.

The engine takes the business details as input data, so the *same* engine serves
your domains and a paying stranger. That's deliberate — it keeps idea.uk a
self-contained thing you could sell later.

---

## 2. The shape of it (diagram)

```
                                  ┌─────────────────────────────────────────┐
   visitor's browser              │            idea.uk SERVICE                │
   ┌───────────────┐  POST /request │  (small always-on container)            │
   │ idea.uk page  │───────────────▶│                                         │
   │ (static, on   │                │  service.go ── front desk               │
   │  Backblaze B2)│◀───────────────│     │                                   │
   └───────────────┘  "we'll reply" │     ├─ store.go    (orders, JSON/PG)     │
          │                         │     ├─ billing.go  (Stripe)              │
          │ pay on Stripe page      │     └─ engine.go   ── the brain          │
          ▼                         │            │                            │
   ┌───────────────┐                │            ├─▶ Anthropic API (generate,  │
   │ Stripe Checkout│  webhook       │            │   verify, score)           │
   │ (hosted by    │───────────────▶│  /stripe/webhook  (← source of truth)   │
   │  Stripe)      │                │            └─▶ OpenAI API (the "cut",    │
   └───────────────┘                │                only if key is set)       │
                                    │                                         │
                                    │  delivers report by email (or holds     │
                                    │  it for you to review first)            │
                                    └─────────────────────────────────────────┘
```

Two halves with very different hosting needs — which is the next section.

---

## 3. Hosting: what is "serverless" and what isn't

This matters because the word "serverless" only applies to one half.

**The page is serverless.** `idea_uk_fakedoor.html` is a static file. It deploys
exactly like all your other sites: git → GitHub Actions → Backblaze B2 (S3).
No server, nothing to keep running, no per-request cost. Same pipeline you
already operate.

**The service is NOT serverless — and can't really be.** The engine is a
*minutes-long* job: it makes several large LLM calls and live web searches per
report. That rules out the usual "serverless function" (those are built for
sub-second to a few-minutes bursts, and billing webhooks need a stable address
that's always listening). So the service is a **small always-on container** —
one cheap box that's always up. Options, cheapest-first:

- A small VM (Hetzner/DigitalOcean/Fly.io/Railway) running the container.
- A single pod in your existing `ai-persona-system` Kubernetes namespace.

It needs: ~256–512MB RAM, outbound HTTPS (to Anthropic/OpenAI/Stripe), one
inbound HTTPS port (for the page's form posts and Stripe's webhook), and a small
disk for the JSON store (or a Postgres connection).

So the honest one-liner: **static page on B2 (serverless) + one small persistent
service (not serverless)**. idea.uk is "static front + small back end," unlike
your pure-static content sites.

---

## 4. The Stripe connection (step by step)

The rule that keeps this safe: **the browser is never trusted about payment —
only Stripe's signed webhook is.** Someone hitting the "success" URL by hand
proves nothing; the money is real only when Stripe tells the service so, with a
signature we verify.

The sequence:

```
1. Visitor fills the request form        → POST /request        (no money yet)
2. You review it, then confirm           → POST /confirm        (operator action)
3. Service asks Stripe for a pay link     → Stripe API (secret key)
   and emails it to the customer
4. Customer pays on Stripe's own page     → (Stripe hosts this; we never see cards)
5. Stripe sends a signed "paid" event     → POST /stripe/webhook (SOURCE OF TRUTH)
6. Service verifies the signature,
   marks the order paid, runs the engine,
   delivers the report (or holds for review)
```

What you set up in Stripe, once:

- Get your **secret key** (`sk_live_…`) → service env `STRIPE_SECRET_KEY`.
- Create a **webhook endpoint** in the Stripe dashboard pointing at
  `https://idea.uk/stripe/webhook`, subscribed to `checkout.session.completed`.
  Stripe gives you a **signing secret** (`whsec_…`) → env `STRIPE_WEBHOOK_SECRET`.

That's it. The service builds the checkout, Stripe handles the card form and PCI,
the webhook confirms it. No card data ever touches your box. If both Stripe env
vars are missing, the service quietly uses a **Fake** payment provider for local
testing (no real money, no keys).

---

## 5. The safety gate you should know about

`AUTO_DELIVER` defaults to **off**. When off, a paid report is generated and then
**held for you to review** before it's emailed — you get the draft, you send it.
This honours the page's "we reply / refund if it's not useful" promise during the
early phase. Turn it on (`AUTO_DELIVER=true`) only once you trust the output.

There's also a **capacity limit** (`MAX_ACTIVE_ORDERS`, default 8): once that many
orders are in flight, `/confirm` refuses new ones, so you can't oversell what you
can deliver in 72 hours. `/capacity` reports the current state.

---

## 6. How to deploy idea.uk (the checklist)

1. **Build the container:** `docker build -t idea-svc .` (the Dockerfile is a Go
   multi-stage build; the binary is tiny).
2. **Run it** on your small box, with environment from `.env` (see `.env.example`):
   `ANTHROPIC_API_KEY`, `INTERNAL_API_KEY` (a random secret you choose),
   `OPERATOR_EMAIL`, `PUBLIC_BASE_URL=https://idea.uk`, and SMTP\_\* if you want
   real emails (without SMTP it writes reports to files so you can test).
   Mount a volume for the JSON store.
3. **Stripe:** set `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET`, and point a
   Stripe webhook at `https://idea.uk/stripe/webhook` (section 4).
4. **Optional, stronger:** set `OPENAI_API_KEY` (+ `OPENAI_CRITIQUE_MODEL`) to run
   the "cut" step on a different vendor (section 9 explains why).
5. **Deploy the page:** put `idea_uk_fakedoor.html` through your normal git →
   Actions → B2 pipeline. Make sure its form can reach the service — either serve
   the service under the same domain at `/request`, `/stripe/webhook` (a path
   proxy), or set `ALLOWED_ORIGINS` to the page's origin for cross-origin posts.
6. **Keep `AUTO_DELIVER=false`** until you've reviewed a few real reports.

Day-to-day after that: a request emails you; you run `/confirm` or `/decline`;
on payment the draft lands in your inbox; you read it and send it.

---

## 7. How to apply it to OTHER domains

Two genuinely different shapes — pick per domain.

**Shape A — the site IS the service (like idea.uk).** The domain is a thin static
page fronting the engine. To stand up another one (say `ideas-for-vets.uk`):
deploy its own static page through the normal pipeline, and either run a second
service instance or — better — point its page at the **one** idea.uk service.
Many pages, one engine. This is the cheap way to test the same product for
different audiences.

**Shape B — a "request a report" panel on an ordinary content site.** A normal
site you build gets a static page (like the fakedoor's request form) that posts
to the central service. The heavy engine stays central; the site just collects
the request. This is the only sensible way to put the ideation product on a
content site, because (next paragraph) it can't be a normal embedded tool.

**Why it can't be a normal "tool".** Your existing tools (`deploy_tool_to_site`)
are self-contained HTML/JS widgets forked into a site and rendered statically — a
VAT calculator runs entirely in the visitor's browser. The ideation engine is the
opposite: server-side, minutes long, costs real money per run, and needs payment.
So it does **not** go into `content_components` as a forked tool. It lives in the
central service; sites only ever *link to* it. Worth keeping that line clear so
nobody tries to fork it like a calculator.

---

## 8. Setting it up inside the chassis (workflows + actions)

Right now the Go service is **standalone** (its own server, its own engine). That
was the right MVP and it's sale-ready. The framework-native version is a
*second* way to run the same method — as an orchestrator agent — and you mostly
**reuse actions you already have** rather than porting `engine.go`.

Looking at the action registry, the method maps almost entirely onto existing
actions:

| Method step | Existing action(s) to reuse |
|---|---|
| Frame + challenge audience | `execute_llm_prompt` |
| Generate across four lenses | `execute_llm_prompt` |
| Cut (different model) | `execute_llm_prompt` (with a different model/provider configured for the step) |
| Verify with web search | `web_search` (and/or `scrape_web`, `firecrawl_*`) |
| Score + rank | `execute_llm_prompt` |
| Operator confirm / review gate | `request_human_input` / `create_approval_request` / `await_approval` / `process_approval_decision` (your HITL actions) |
| Notify operator | `send_notification` |
| Persist run + result | `store_result` / `write_my_state` / `store_memory` |
| Read/declare on a site | `read_site_spec` / `write_site_spec` |

So the chassis version is: **one new agent definition** (an
`idea-orchestrator`, every agent is an orchestrator) **+ one workflow** that
sequences the steps above, **+ at most a thin new action or two** only where an
existing one doesn't fit. Per your conventions: keep the workflow simple, put any
real logic in Go action code, keep workflow variable names matching what the
actions expect, and where a step needs heavier work (e.g. the four-lens
generation) **spawn a sub-agent with its own workflow** (`spawn_agent` /
`spawn_group`) rather than nesting subworkflows in SQL — children reply on the
parent's responses topic.

The billing half (Stripe, request-then-confirm, the customer-facing flow) is
**not** a natural fit for an internal agent workflow — that's a product/payment
concern. Keep that in the standalone service. The chassis agent is for running
the **method** internally across your own domains (the "internal" path), on a
schedule or on demand, writing results back to `site_specs` or a results table.

> I have **not** written this agent/workflow SQL yet, on purpose: it needs a pass
> over the real `agent_definitions` / workflow schema and the exact
> `execute_llm_prompt` / HITL action contracts first (check schema before SQL,
> reuse before rebuild). When you want it, we read those together and wire it to
> match — it should be a small amount of SQL plus maybe one or two Go actions.

### What "apply it via a site spec" looks like

Your `site_specs` model already has the mechanism. A domain that should offer the
ideation product would carry it as an item in its spec — most naturally in
`site_plan` (a page/feature) with a `status`:

- `blocked` — if the capability/agent isn't deployed yet (your `feasibility-recheck`
  task promotes it to `planned` once it is).
- `planned` → built — the build pipeline creates the static "request a report"
  page (Shape B above) and links it to the central service.

So "apply this tool to a new domain" = the classifier/planner writes an ideation
feature into that site's spec; because the engine is central, the per-site build
is just the request page + link, not a forked component. For an idea.uk-style
domain (Shape A), the spec's `classification`/`identity` says "this domain *is*
the ideation service," and the page it builds is the landing/intake page.

---

## 9. Running it — and the OpenAI question

### Running the engine once (no server, no billing)

You already did this:

```bash
cd idea-go
go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
```

That runs the real brain against the real Anthropic API and prints the report.
It cost real API tokens. (Your agritec run worked — advisor audience, scheme-diff
alerts, BNG cross-check — so the Go engine is producing real reports.)

### Did that use OpenAI? How to be sure.

The **cut** step (step 3, the ruthless filter) uses OpenAI **only if
`OPENAI_API_KEY` is set in the same shell**; otherwise it uses Anthropic
(Claude Sonnet). Nothing else in the pipeline uses OpenAI.

I just added a line so it tells you. You'll now see, on stderr, one of:

```
[cut] cross-vendor: OpenAI (gpt-4o)
[cut] same-vendor: Anthropic (claude-sonnet-4-6)
```

To check whether your key is set: `echo $OPENAI_API_KEY` (blank = not set).
To use OpenAI for the cut:

```bash
export OPENAI_API_KEY=sk-...
export OPENAI_CRITIQUE_MODEL=gpt-4o      # or whatever current model you prefer
go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
# watch for: [cut] cross-vendor: OpenAI (gpt-4o)
```

So: if the key was already exported when you ran it, you **already** used OpenAI
for the cut — the new log line will confirm it on the next run. If it wasn't, the
cut ran on Claude Sonnet, which is still a *different model* from the generator,
just the same vendor.

**Why bother with OpenAI here?** The whole point of the cut is to have something
*other than the generator* judge the ideas, so the method doesn't rubber-stamp
its own output. A different vendor is the strongest version of that. Running the
same domain once with the key and once without, and comparing, is the cleanest
test of whether cross-vendor critique actually changes the verdicts.

### Running the test (this is separate from the engine)

The test checks the **plumbing** — the request → confirm → pay → deliver state
machine, the auth gates, idempotency, the capacity limit. It uses a **fake**
payment provider and a **stubbed** engine, so it makes **no API calls, costs
nothing, and needs no keys**:

```bash
cd idea-go
GOPROXY=off GOTOOLCHAIN=local go test ./...
# expect: ok  idea  (PASS)
```

Add `-v` to see each check:

```bash
GOPROXY=off GOTOOLCHAIN=local go test -v ./...
```

Key point that resolves the confusion: **the test does not touch OpenAI or
Anthropic at all.** It deliberately swaps the real engine for a stub so it can
test the money/flow logic fast and free. OpenAI only ever comes in when you run
the *real* engine (`go run . internal …`, or a real paid order through the
service). The two are independent.

`GOPROXY=off GOTOOLCHAIN=local` just tells Go "don't go to the internet" — the
code is stdlib-only so it builds offline. On a normal networked machine you can
drop those and plain `go test ./...` / `go run .` work too.

---

## 10. Status — what's real today vs what's next

Real and working now:
- The Go engine produces real reports against live APIs (your agritec run).
- The service plumbing is built and passes its test (flow, billing logic,
  idempotency, capacity, auth).
- Stripe code is written (checkout + signed webhook); not yet pointed at a real
  account.
- The page exists; not yet deployed.

Not built yet (and deliberately so):
- The chassis-native agent + workflow (section 8) — needs a schema pass first.
- Live deployment (container hosted, Stripe keys in, page on B2).
- `AUTO_DELIVER` stays off until real reports are reviewed.

Smallest useful next steps, any order:
1. Run agritec once with `OPENAI_API_KEY` set and compare to without — settles the
   cross-vendor question and you'll see the new `[cut]` log line.
2. Stand up the container on one small box with Stripe test keys and walk one
   request → confirm → pay → deliver through end-to-end with real (test-mode)
   Stripe.
3. When you want it inside the chassis, we read the `agent_definitions` /
   workflow schema and the `execute_llm_prompt` + HITL action contracts together,
   and wire the `idea-orchestrator` agent to reuse them.

---

## Update — 2026-06-05 (decisions since first deploy)

The service is now **live on a Hetzner box** (CX-class x86, Nuremberg) behind nginx
+ Let's Encrypt, on FakeProvider (no Stripe keys yet). Decisions and changes since
this doc was first written:

**Page serving — embedded in the binary, not hosted separately.** The landing page
lives in the module as `page.html` and is compiled in with `//go:embed`; the service
serves it at `/`. The earlier "page on B2" idea is dropped for the single-box product:
one self-contained artefact matches the chassis "ship the binary" model, keeps nginx
to just TLS + proxy, and means the deployer ships one thing. nginx proxies everything
to the Go service; there is no separate static host. Editing the page means a rebuild.

**Server-rendered pages share one wrapper.** `a.page(title, body)` produces a full
brand-styled document (cream/ink/rust, Fraunces + IBM Plex, header bar, footer with
the contact email). The post-submit pages (request received, newsletter, order
success/cancel) and the policy pages all render through it. `writeHTML` is reserved
for fragments injected into an already-styled page (the taster result). See debug
guide §11 for the missing-page / missing-design failure modes this caused.

**Policy pages.** `/terms`, `/refund-policy`, and `/privacy` are served from string
constants (`termsBody` / `refundBody` / `privacyBody`) through the wrapper, with a
`{{EMAIL}}` token filled at serve time. Plain-language drafts; **a UK solicitor must
review before taking real payments** (runbook A6). Trading name is **idea.uk, no
address**. The terms now state plainly that reports are **AI-generated and can be
wrong (hallucination, invented facts/figures/sources, staleness)**, that the customer
must verify everything and take their own advice, and that use of the report is the
customer's responsibility, not ours. Privacy policy is UK-GDPR-shaped, names Stripe
and Anthropic as processors, flags the international transfer (Anthropic, US), states
no cookies/tracking, gives ICO recourse, and keeps security wording measured (low
risk appetite). Open items left as bracketed placeholders: hosting/email provider
names, transfer safeguards, and the data-retention period.

**Config.** `CONTACT_EMAIL` (public support address; falls back to `OPERATOR_EMAIL`)
and `MONTH_SLOTS` (header capacity phrase; falls back to "a limited number of") are
read from the env and templated into the page at startup. `REPORT_PRICE_GBP=29`.

**Copy.** Landing page and all server-rendered pages rewritten into plain,
matter-of-fact language — no LLM-ish vocabulary (honest/gate/deck/asset), no "X, not
Y" framing, and the method is described in user terms, not AI terms. The taster footer
spells out free-audience-check vs the £29 report in plain benefits.

**Redeploy** (any of the above, since the page is embedded): rebuild amd64, scp to
`/opt/idea/idea.new`, `mv -f` over `/opt/idea/idea`, `systemctl restart idea`.

---

## Correction — 2026-06-05 (email: Clook + leopardess.uk, not SES)

Supersedes the SES-for-sending line in the previous update. Email decisions:
- **Operator domain: leopardess.uk** — one neutral brand for all sites' system/transactional/
  support mail and the address replies come back to (NOT a bulk-sender for the lead-gen long
  tail; high-volume sites should use their own sending domain).
- **Provider: Clook (UK)** for BOTH sending and receiving, to start. SES London remains the
  documented swap (the app speaks plain SMTP; switching is env-only). Reason to start on
  Clook: low volume, one UK provider, no AWS account; move idea.uk to SES later if volume/
  deliverability warrant.
- **idea.uk sends only as leopardess.uk** (e.g. idea-uk@leopardess.uk), and the idea.uk pages
  will say "by leopardess.uk" so the sender isn't confusing. No idea.uk DKIM needed.
- **Inbound: Clook catch-all** on leopardess.uk -> forwards to a Gmail inbox
  (aaa@designconsultancy.co.uk). Per-site address = deterministic encoding of the domain
  (dots -> dashes): idea.uk -> idea-uk@leopardess.uk.
- **DNS: hosted at Clook** (cPanel zone). SPF + DMARC(p=none) present; DKIM to confirm via
  cPanel Email Deliverability.
- Framework structure for this: see EMAIL_identity_in_site_spec.md — a new `email` site_specs
  aspect (no DDL) + global operator config, structured for a future per-domain email-
  provisioner agent (not implemented; the catch-all makes it unnecessary now).
- Privacy policy updated to name Clook (UK) + Google (Gmail) instead of SES.

## Update — 2026-06-06 (email working both ways; the 465 detail; exact env)

> **Superseded on the port detail — see the 2026-06-10 update below.** From the live box,
> outbound 465 is blocked by Hetzner; the working submission port is **587**. The "passed
> SPF/DKIM/DMARC at Gmail" line below was not borne out by the delivery report — treat the
> 2026-06-10 update as current.


The email setup from the correction above is now tested and working. Specifics that matter:

- **Outbound — confirmed.** A send as `system@leopardess.uk` reached Gmail passing SPF, DKIM
  (aligned `d=leopardess.uk`, `s=default`), and DMARC, in ~3s. Clook routes outbound via
  **MailChannels**; the leopardess.uk SPF authorises it. cPanel Email Deliverability shows
  DKIM/SPF/DMARC/PTR all valid for **leopardess.uk** (the .uk, not the .co.uk twin — both are
  on the account and it is easy to edit the wrong one).
- **The SMTP port detail.** cPanel → Email Accounts → Connect Devices lists outgoing
  **`mail.leopardess.uk`** on **SMTP 465 (SSL/TLS) only** — no 587. Port 465 is *implicit TLS*,
  which Go's `smtp.SendMail` does not do (it does STARTTLS). So `service.go` has a `smtpSend`
  helper: **port 465 → `tls.Dial` + `smtp.NewClient` + Auth/Mail/Rcpt/Data/Quit**; any other
  port → `smtp.SendMail` (STARTTLS). This is why `SMTP_PORT=465` works where the stock library
  would have failed.
- **Inbound — confirmed.** leopardess.uk **Default Address (catch-all)** → "Forward to Email
  Address" → `aaa@designconsultancy.co.uk`. A test from an *external* sender arrived. (A test
  from the same Gmail appears to vanish — Gmail dedupes its own message by Message-ID; always
  test inbound from a different sender.) The cPanel "Forward all email for a domain" feature is
  NOT the mechanism — for a local mail domain the Default Address governs the catch-all.
- **Two Google Workspace features are admin-gated** on the receiving account
  (designconsultancy.co.uk is Workspace): "Check mail from other accounts" (POP fetch) and
  "Send mail as" via external SMTP are disabled at the admin level. Consequence: *personal*
  replies go out as aaa@designconsultancy.co.uk, not idea-uk@ — cosmetic only. idea.uk's
  *automated* mail is unaffected (the Go service sends it via Clook as idea-uk@).
- **Exact env** (`/etc/idea/idea.env`): `SMTP_HOST=mail.leopardess.uk`, `SMTP_PORT=465`,
  `SMTP_USER=system@leopardess.uk`, `SMTP_PASS=<mailbox password>`,
  `SMTP_FROM=idea-uk@leopardess.uk`, `SMTP_FROM_NAME=idea.uk`,
  `SMTP_REPLY_TO=idea-uk@leopardess.uk`, `CONTACT_EMAIL=idea-uk@leopardess.uk`,
  `OPERATOR_EMAIL=idea-uk@leopardess.uk`.
- **One thing still unproven at the app level:** the service authenticates as `system@` but
  sends From `idea-uk@` (different local part). DKIM/SPF/DMARC pass regardless (they key on the
  domain), so it should work; if a send ever errors "sender not allowed", set
  `SMTP_FROM=system@leopardess.uk` and keep `SMTP_REPLY_TO=idea-uk@` so replies still route.
- **Design note:** EMAIL_identity_in_site_spec.md now prefers a **specific forwarder per
  published site** over a server catch-all (only forwards addresses that exist; no backscatter;
  matches the future email-provisioner). The catch-all is a manual-phase convenience.

leopardess.uk also now has a one-page identity site (leopardess_uk_index.html), live on Clook.

## Update — 2026-06-10 (email: 587 not 465; mailer made async/bounded; MailChannels content filter)

Current email state from testing on the live box. Supersedes the port detail in the 2026-06-06
update.

**Outbound port is 587, not 465.** A TCP port sweep *from the box* to `mail.leopardess.uk`
(62.182.23.30):

```
25 blocked · 465 blocked · 2525 blocked · 587 OPEN · 80 open · 443 open
```

So Hetzner blocks outbound 25/465/2525 but leaves **587 (submission)** open — this was never a
blanket SMTP block. The cPanel "Connect Devices" page advertises 465, which is what misled the
06-06 note, but 465 is unreachable from the box (the send hung on the TCP connect timeout, ~2 min,
then failed). Set **`SMTP_PORT=587`**. The code's `smtpSend` takes the `smtp.SendMail` STARTTLS
path for any port != 465 — correct for a 587 submission port. The 465 implicit-TLS branch stays
for hosts that need it, but is unused here.

**Service→Clook submission works.** A test message authenticated (`dovecot_plain`) as
`system@leopardess.uk`, From `idea-uk@leopardess.uk`, sender IP = the box — accepted by Clook,
confirmed in the cPanel delivery report, with no error in `journalctl -u idea`.

**Mailer code change (`service.go`, deployed):**
- `makeDeliver` now wraps the send in a goroutine. It was running `smtpSend` **inline on the HTTP
  request path**, so the failed 465 connect froze the visitor "thanks" page for ~2 minutes. Now
  the request returns immediately and the send result is logged. (Failures are logged, not
  surfaced — the webhook/store is the source of truth for fulfilment; these mails are
  notifications.)
- `smtpSend`'s 465 path now uses `tls.DialWithDialer(&net.Dialer{Timeout: 10s}, …)` plus a 30s
  `conn.SetDeadline`, so a network problem fails fast instead of hanging. Added the `net` import.
- Verified to build + `go vet` clean (Go 1.22).

**The open item — MailChannels content filter.** Clook relays outbound via MailChannels, which
spam-filters. The operator "NEW REQUEST" notification — a *forward* (`idea-uk@leopardess.uk` is a
local address with no mailbox, so the catch-all forwards it to the Gmail) — was rejected
`550 5.7.1 [CS] Message blocked`; MailChannels Insights shows **"Blocked (Spam Content)"**. The
likely triggers are in that body: it opens with a literal `From: name <email>` line (which reads
as an embedded/forged header) and contains raw `POST /confirm {"order_id":…}` JSON. The buyer
confirmation body, by contrast, is clean prose plus a payment link.

**Next (decides the fix):** test the buyer path — a confirmation to a real external Gmail (a
normal send, not a forward). If it delivers, only the operator notification needs tidying
(relabel `From:` → `Requester:`, drop the raw JSON). If the buyer confirmation is *also* blocked,
MailChannels is filtering legitimate outbound and the move is to ask Clook to relax outbound
filtering for the account or to send transactional mail through a dedicated provider over 443 —
the more dependable route for a transactional service. Clicking "Not Spam" in MailChannels
Insights trains the filter on the false positive.

**Lesson for the framework:** a cloud box typically can't use outbound 25/465 (provider blocks),
and a shared-host relay (MailChannels) content-filters outbound — both bite transactional mail.
Plan the framework's transactional sending accordingly (dedicated sender or a relay that won't
filter), and keep transactional message bodies clean (no embedded header-like lines, no raw JSON).
