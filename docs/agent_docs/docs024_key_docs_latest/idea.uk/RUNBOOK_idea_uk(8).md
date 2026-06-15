# idea.uk — runbook

The pieces:

| File | Role |
|---|---|
| `idea-go/engine.go` + `prompts.go` | The engine — runs the method (multi-lens generate → cut → web-verify → score → rank). Used by both front doors. |
| `idea-go/service.go` (+ `store.go`, `billing.go`, `main.go`) | The service — request-then-confirm intake, Stripe billing, fulfilment, internal endpoint. |
| `idea_uk_fakedoor.html` | The static landing/intake page (deploy to S3). Form posts to the service. |
| `idea-go/service_test.go` | End-to-end state-machine test (FakeProvider + stubbed engine). |
| `Dockerfile`, `.env.example` | Deployment. |

## Architecture (note: not edge-only)

idea.uk is **static page + small always-on service**, unlike the day-pass chat
domains (static + synchronous edge worker). The method is a minutes-long
multi-LLM + web-search job, so it runs as a background task in the service, not
in an edge worker. Topology:

```
visitor → idea_uk_fakedoor.html (S3)  ── POST /request ──▶  idea_service (container)
                                                              │  engine (Anthropic + web search,
Stripe ── webhook ──▶ /stripe/webhook ───────────────────────┤  optional OpenAI cross-vendor cut)
                                                              ▼  deliver (SMTP) or hold for review
```

## Flow (request-then-confirm)

1. Visitor submits the form → `POST /request` (free). Order = `requested`.
   Operator is emailed.
2. Operator reviews and either:
   - `POST /confirm {order_id}` (header `X-Internal-Key`) → creates Stripe
     Checkout, emails the customer the pay link. Order = `awaiting_payment`.
     **Refused with 409 if at capacity** (see MAX_ACTIVE_ORDERS).
   - `POST /decline {order_id, reason}` → polite decline email, no charge.
3. Customer pays → Stripe fires `checkout.session.completed` →
   `POST /stripe/webhook` (verified, idempotent) → order = `paid`, fulfilment
   runs in the background.
4. Fulfilment runs the engine. With `AUTO_DELIVER=false` (default) the report is
   held (`awaiting_review`) and the operator is emailed the draft to review
   before sending. With `AUTO_DELIVER=true` it emails the customer directly.

## Local end-to-end test (no keys needed)

```bash
cd idea-go
GOPROXY=off GOTOOLCHAIN=local go test ./...   # expect: PASS (19 checks)
```

Uses the FakeProvider (no Stripe) and a stubbed engine (no LLM spend). Validates
the full state machine, idempotency, capacity, auth, and both front doors.

## Run it for real, locally (FakeProvider, real engine)

```bash
cd idea-go
export ANTHROPIC_API_KEY=...           # real — the engine will spend
export INTERNAL_API_KEY=$(openssl rand -hex 16)
export AUTO_DELIVER=false
GOPROXY=off GOTOOLCHAIN=local go run .   # starts the service on :8080
# Or run the engine directly against one of our own domains (no billing, no server):
go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
# Internal HTTP endpoint:
curl -s localhost:8080/internal/run -H "X-Internal-Key: $INTERNAL_API_KEY" \
  -H 'content-type: application/json' \
  -d '{"domain":"agritec.uk","audience":"UK small farmers","assets":"curate scheme docs"}'
```

(No Stripe keys → FakeProvider; the external order flow still works locally:
`/request` → `/confirm` → visit the returned `/order/success?...&fake=1` URL to
simulate payment and trigger a real engine run.)

## Go-live checklist

1. **Build & deploy** the container: `docker build -t idea-svc . && docker run
   -p 8080:8080 -v idea_data:/data --env-file .env idea-svc  (Go multi-stage build)`. Put it behind TLS.
2. **Stripe**: create a webhook endpoint → `PUBLIC_BASE_URL/stripe/webhook`,
   event `checkout.session.completed`; copy the signing secret to
   `STRIPE_WEBHOOK_SECRET`; set `STRIPE_SECRET_KEY`. Pin the `stripe` lib version.
3. **Static page**: deploy `idea_uk_fakedoor.html` to S3. The form posts to
   `/request`; ensure it reaches the service (same origin via a path proxy, or
   set `ALLOWED_ORIGINS` to the page origin for CORS). Replace `CONTACT_EMAIL`.
4. **Capacity**: set `MAX_ACTIVE_ORDERS` to what you can review+deliver in 72h.
   `/capacity` returns `{open, active, max}` if you want the page to show "full".
5. **Cross-vendor critique** (optional, stronger moat): set `OPENAI_API_KEY` and
   `OPENAI_CRITIQUE_MODEL` (current model) — the cut step then runs on OpenAI.
6. **AUTO_DELIVER stays false** until you trust the output; review each draft and
   send manually (the operator email contains the draft).

## Notes

- Source of truth is the webhook, never the browser redirect.
- Webhooks are idempotent (dedup table). Safe to receive duplicates.
- Engine failure flags the order `failed` and emails the operator to refund.
- The JSON store (IDEA_DB_PATH) holds orders + subscribers + processed webhook events. Back it up
  (it's the record of who paid and what's owed). Mount `/data` on a volume.
- Sale-readiness: the engine takes assets as data and the billing sits behind a
  provider interface, so idea.uk remains a separable unit (PLAN_idea_uk.md §2).

## Status & deployment (2026-06-10)

The "Go-live checklist" above describes the original Docker/S3 plan. What's actually live differs
and is the current truth:

- **Deployment:** idea.uk runs as a **single Go binary under systemd on a Hetzner box**
  (Nuremberg, IPv4 116.203.204.115 — confirm), behind nginx + Let's Encrypt — **not** Docker on
  a container host, and the landing page is **embedded in the binary** (`//go:embed page.html`),
  not a separate file on S3. Redeploy = `&&`-chained build → scp → `mv -f` → `systemctl restart`.
  Env at `/etc/idea/idea.env` (systemd `EnvironmentFile`). See `idea_uk_architecture_and_deployment.md`.
- **Billing:** Stripe Checkout — a single £29 payment per report. Full setup (keys, restricted-key
  permission, webhook) is in the **Stripe billing — setup** section below. With no Stripe keys set, the
  app falls back to a FakeProvider for local testing.
- **Email:** the mailer is deployed; the service→Clook submission works on **port 587** (Hetzner
  blocks 25/465 from the box). One open item: MailChannels is content-filtering outbound — a
  buyer-path test to a real Gmail is outstanding before email is proven end-to-end. Detail in
  HANDOFF.md and the arch doc's 2026-06-10 update.
- **`makeDeliver` is now async** (sends in a goroutine) — mail never blocks the HTTP request path.

## Status & operating update (2026-06-11)

Supersedes the older Flow/Email/AUTO_DELIVER notes above where they differ.

**Order flow now has a switch — `REVIEW_BEFORE_PAY` (default on).**
- On (review-before-pay): `/request` → operator `/confirm` **runs the engine** and holds the draft
  (status `running` → `awaiting_review`; the draft is emailed to the operator). Operator reviews,
  then `/approve` sends the buyer the pay link, or `/decline` (no charge). Buyer pays → the
  already-generated report is delivered. No money is taken until the operator approves.
- Off: the original charge-first flow (`/confirm` sends the pay link; the engine runs after payment;
  `AUTO_DELIVER` decides review-vs-auto-send). Use it as a fallback if engine cost ever spikes.

**Click-through (easiest, new 2026-06-11):** the request and review emails now contain a link to a
page with Confirm / Approve / Decline buttons — just click. The link carries a per-order token
(HMAC of the order id under `INTERNAL_API_KEY`), so no key is needed and it authorises that one order
only. The link opens a **safe GET page** (a mail scanner pre-fetching it can't trigger anything); the
action fires only when you click a button (a POST carrying the token). Actions stay gated by order
status, so a token can't, e.g., confirm twice. The curl commands below still work as a fallback.

**Operator commands (on the box):**
```
KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_..."}'   # runs engine (review-before-pay)
curl -s localhost:8080/approve -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_..."}'   # bills the buyer
curl -s localhost:8080/decline -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_...","reason":"..."}'
```

**Email = AWS SES (London `eu-west-2`), STARTTLS 587.** Clook/MailChannels content-filtered
leopardess.uk's outbound, so sending moved to SES (From stays `idea-uk@leopardess.uk` with
leopardess.uk DKIM; pay-link reaches the inbox, DKIM/DMARC/SPF pass). Gotcha: SES `SMTP_USER` is the
`AKIA…` access-key-id, NOT the `ses-smtp-user.…` IAM name. Reports go out as **multipart HTML email**
(styled HTML + plain-text fallback); short notifications stay plain.

**Stripe billing:** described in full in the **Stripe billing — setup** section below.

**Logging / records.**
- Free taster: each run logs business, stated audience, and the result (carried audience,
  willingness, alternatives) — `journalctl -u idea | grep "free taster"`.
- Paid reports: stored per order in `/var/lib/idea/orders.json` (`report` plain + `report_html`),
  with status/email/timestamps — the full record of paid tasks.

**Planned: optional PDF of the report.** To make the £29 spend more tangible we intend to also
deliver the report as a PDF (attachment or link). Constraint: the service is stdlib-only and builds
offline (`GOPROXY=off`), so a PDF generator means either relaxing that to vendor a library or
rendering the existing report HTML to PDF. Not built yet — see HANDOFF backlog.

## Troubleshooting: a confirmed order's draft didn't arrive (2026-06-11)

After Confirm, `fulfil` runs in the background: engine (minutes) → store draft → email it to
`OPERATOR_EMAIL`. If the draft doesn't arrive:

1. Status + whether the draft was generated and stored:
   ```
   ssh root@116.203.204.115 "python3 -c \"import json;o=json.load(open('/var/lib/idea/orders.json'))['orders']['<ORDER_ID>'];print('status:',o['status'],'| report stored:',bool(o.get('report')),'| html:',bool(o.get('report_html')))\""
   ```
   - `awaiting_review` + report stored → engine OK, draft SAVED (not lost); the **email** is the problem.
   - `running` (and it's been >10 min) → engine hung, or the service was restarted mid-run (a restart
     kills the in-flight goroutine — don't redeploy while a run is in progress).
   - `failed` → engine error; a RUN FAILED email should have gone to OPERATOR_EMAIL (check spam).
2. Watch the run live (needs the 2026-06-11 logging): `journalctl -u idea -f`, then confirm a test order.
   Expect: `fulfil: <id> running engine` → `engine done (N chars)` → `draft ready, emailing review to
   <addr>` → `email to <addr> sent` (or `email to <addr> failed: <err>`).
3. The draft is in `orders.json` (`report`/`report_html`) even if the email failed — recover it from
   there if needed. The draft goes to OPERATOR_EMAIL, not the requester.

## Stripe billing — setup

idea.uk takes a single £29 payment per report through Stripe Checkout (hosted). The app uses Stripe
when both `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET` are set; with neither it falls back to a
FakeProvider for local testing. No code change is needed to switch Stripe on — `orderSuccess` already
handles a real redirect (no `fake` param → it just shows "Payment received"; the webhook does delivery).

**Where the keys go:** both values live in **`/etc/idea/idea.env`** on the box (the systemd
`EnvironmentFile`). After editing it, run `systemctl restart idea`. Keep the keys to this file only —
never in chat, the repo, or the page. The two values the app reads:
- `STRIPE_SECRET_KEY` — the key the app calls Stripe with (use a restricted key — see below).
- `STRIPE_WEBHOOK_SECRET` — the signing secret of the webhook endpoint (`whsec_…`, from the
  destination's details page).

**Webhook destination (test — created 2026-06-13):**
- Name: `idea-uk-webtool-ideas`
- Destination ID: `we_1TiHnv08YuzM2cqfHjQSbdF2`
- Endpoint URL: `https://idea.uk/stripe/webhook`
- API version: `2025-04-30.basil`
- Listening to: `checkout.session.completed` (1 event)
- Signing secret: reveal/roll it on the destination's details page → copy into `STRIPE_WEBHOOK_SECRET`.

That is the **test** destination. Going live needs a **separate** destination created in live mode,
with its own signing secret.

**Dashboard links (the home page doesn't link straight to these):**
- Test — **sandbox** account `acct_1RNfPY08YuzM2cqf`:
  - Webhooks: `https://dashboard.stripe.com/acct_1RNfPY08YuzM2cqf/test/workbench/webhooks`
- Live — **real** account `acct_1RNfPL02nQ76FNif` (note: a *different* account id from the sandbox):
  - API keys: `https://dashboard.stripe.com/acct_1RNfPL02nQ76FNif/apikeys`
  - Webhooks: `https://dashboard.stripe.com/acct_1RNfPL02nQ76FNif/workbench/webhooks`

The test webhook lives in the sandbox account; the live keys and the live webhook live in the real
account — so they're managed in two different places, and a sandbox webhook does **not** cover live.

### Which API key (use a restricted key)
The app reads exactly two Stripe env vars: `STRIPE_SECRET_KEY` (the bearer token for API calls) and
`STRIPE_WEBHOOK_SECRET` (the endpoint's signing secret). It does NOT use a publishable key (Checkout is
hosted — the publishable key is for client-side JS we don't run) or an organisation key. So of the keys
Stripe offers, you only need to supply one secret-type key plus the webhook secret.
- **Use a Restricted API key (`rk_…`) as `STRIPE_SECRET_KEY`**, not the full secret key — least
  privilege, so a leak can't drain the account. It's a bearer token like any key, so there's no code
  change (the app just puts it in the `Authorization: Bearer` header).
- Minimal permission the integration needs: **Checkout Sessions → Write**. The app creates a Checkout
  Session with inline `price_data`; it reads no customers, lists nothing, and does refunds by hand. The
  webhook needs **no** API permission — it only verifies a signature locally.
- **In the restricted-key permission list: set `Checkout Sessions` → Write, and leave everything else
  `None`** — including Payment Intents, Charges and Refunds, Customers, Products, Prices, and Webhook
  Endpoints. (The session is created with inline pricing, the buyer email is passed inline, and webhooks
  are verified locally, so none of those scopes are needed.)
- If a checkout create ever returns a permissions error, Stripe names the exact missing scope — grant
  that and retry. (If we later add API-driven refunds, add **Charges and Refunds → Write**.)
- Create the restricted key in the **same mode** you're working in: a test/sandbox `rk_test_…` for the
  test pass, a live `rk_live_…` for live.

### Which webhook events
Select ONE event: **`checkout.session.completed`**. It's the only event the app handles (it marks the
order paid and triggers delivery). Ignore the long list of `account.*`, `v2.core.*`, and "Connected
accounts" events — those are for Stripe **Connect** platforms managing other people's accounts, which we
are not. Set **Event destination scope = "Your account"** (not Connected accounts). The Workbench flow
asks for events first, then the destination type, then the endpoint URL
(`https://idea.uk/stripe/webhook`). (Optional, only if we add refund recording later: also add
`charge.refunded`.)

### Sandbox vs live keys
A Stripe **sandbox / test mode gives TEST keys only** (`rk_test_`/`sk_test_`, and a test webhook
secret). You **cannot** get live keys from a sandbox. Live keys come from **live mode in the real
account**, where you also create a **separate** live webhook endpoint with its own signing secret. So
the sandbox you're in now is exactly right for the test pass below — you don't need live keys yet; get
them when you flip to live (step B).

### A. Prove the real Stripe path in TEST mode (no real money)
1. Stripe dashboard, **test mode** → Developers → Webhooks → Add endpoint: URL
   `https://idea.uk/stripe/webhook`, event `checkout.session.completed`. Copy its signing secret (`whsec_…`).
2. `/etc/idea/idea.env`: `STRIPE_SECRET_KEY=<test restricted key, rk_test_…>`,
   `STRIPE_WEBHOOK_SECRET=whsec_…` (the test endpoint's secret) → `systemctl restart idea`. Startup log
   must stop saying "No Stripe keys".
3. Request → confirm → approve → the pay-link email now carries a real Stripe link → pay with test card
   `4242 4242 4242 4242` (any future expiry/CVC) → land on "Payment received" → the webhook shows 200 in
   the dashboard → report delivered.

### B. Go live and bill yourself
4. Switch the dashboard to **live mode** → create a SEPARATE webhook endpoint (same URL + event) → copy
   its **live** signing secret. Get the live secret key from Developers → API keys (live).
5. `/etc/idea/idea.env`: `STRIPE_SECRET_KEY=<live restricted key, rk_live_…>`,
   `STRIPE_WEBHOOK_SECRET=whsec_…` (the live endpoint's secret) → restart.
6. Put one real £29 through on your own card; confirm the same chain works end to end.

### C. Refunds — manual, in the dashboard
- There is **no refund code**: no `/refund` endpoint, and the webhook ignores refund events. The refund
  promise (the /refund-policy page, the terms, and the 14-day line in emails) is fulfilled by **you**
  issuing the refund in the Stripe dashboard when a customer emails. The app does **not** record it (the
  order stays `delivered`).
- To test: dashboard → open the payment → Refund → full amount → confirm it processes (money returns
  over a few days).
- Optional, not built: a `charge.refunded` webhook handler could mark the order `refunded` — ask if wanted.

### Gotchas
- Live keys move **real money** — keep them in the box env only, never paste them anywhere.
- The webhook signing secret **differs** between test and live and must match the keys you've set, or the
  signature check fails and the order never moves past payment.
- Stripe does **not** return its per-transaction fee on a refund (~60–65p on £29), so a self-test refund
  costs that.

## HTML emails are base64-encoded (2026-06-11)

All multipart (HTML) emails — report, review draft, delivered report, pay-link — have their parts
base64-encoded and wrapped at 76 characters, so no line hits the SMTP 998-octet limit. Before this, a
long unbroken HTML line was being folded mid-tag by a mail server and showed up as literal `< p …>` in
the email. If an email ever shows literal tags again, that's the symptom — check the transfer encoding.

## Troubleshooting: paid but no report (the webhook)

Delivery after payment is driven entirely by the Stripe **webhook** (the success page just says
"Payment received"). So "paid but no report" almost always means the webhook didn't fire or didn't
verify — especially right after going live, because the live webhook + live signing secret are separate
from the sandbox ones.

1. **Stripe dashboard, LIVE mode** (`acct_1RNfPL02nQ76FNif`) → Workbench → Webhooks → the
   `https://idea.uk/stripe/webhook` endpoint → recent deliveries for your payment:
   - **No live endpoint at all** → that's it: create one in live mode, copy its signing secret into
     `STRIPE_WEBHOOK_SECRET`, restart. (The test endpoint is in the sandbox account; it does not cover live.)
   - **Response 400** → idea.uk rejected the signature → `STRIPE_WEBHOOK_SECRET` doesn't match this live
     endpoint's secret. Causes seen: the test/sandbox secret left in place; the secret copied from a
     *different* endpoint; or a hidden character (trailing space / `\r`) in the env value. To fix:
     re-reveal the signing secret on **this** endpoint's details page, then on the box check the value is
     clean — `grep '^STRIPE_WEBHOOK_SECRET=' /etc/idea/idea.env | cat -A` should end `…fG5c0$` with
     nothing between the secret and the `$` (no space, no `^M`, no quotes). Confirm the running process
     actually has it: `PID=$(systemctl show -p MainPID --value idea); tr '\0' '\n' < /proc/$PID/environ |
     grep STRIPE_WEBHOOK_SECRET`. Restart after any change.
   - **Response 200** → accepted → the issue is downstream; check the box (step 2).
2. **On the box:**
   - Order status: `awaiting_payment` still → the webhook didn't complete (matches a 400 / missing
     endpoint). `delivered` → it was sent; check the buyer's spam.
   - `journalctl -u idea | grep -iE 'stripe webhook|email to'` → look for `REJECTED (signature/parse)`
     (secret mismatch), `order … paid → delivering`, and `email to <buyer> sent`.
3. **Recover the already-paid order:** once the live endpoint + secret are right and the service is
   restarted, **resend** the `checkout.session.completed` event. In the Workbench, Resend is under the
   endpoint's **Event deliveries** tab (not Overview) → click the failed delivery → **Resend** (or open
   the event itself → Resend). Stripe also auto-retries failed deliveries with backoff for ~3 days, so a
   corrected secret will eventually deliver on its own — resending is just faster. On success the handler
   runs `awaiting_payment → paid → deliverReport → report sent`. The report is already stored in
   `orders.json`, so nothing is lost.
