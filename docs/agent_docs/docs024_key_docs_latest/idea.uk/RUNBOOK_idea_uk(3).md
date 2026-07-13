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
- **Billing:** still **FakeProvider** — no Stripe keys yet. Stripe is the next step
  (`PLAN_stripe_billing_integration.md`) and the thing standing between live and earning.
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

**Stripe go-live (test mode).** Code is ready; switches to Stripe when both `STRIPE_SECRET_KEY` and
`STRIPE_WEBHOOK_SECRET` are set. Create a webhook endpoint at `https://idea.uk/stripe/webhook`
(event `checkout.session.completed`), copy its signing secret to `STRIPE_WEBHOOK_SECRET`, set
`STRIPE_SECRET_KEY=sk_test_…`, restart. The publishable key isn't used server-side. Test card
`4242 4242 4242 4242`. Startup log stops saying "No Stripe keys" once live.

**Logging / records.**
- Free taster: each run logs business, stated audience, and the result (carried audience,
  willingness, alternatives) — `journalctl -u idea | grep "free taster"`.
- Paid reports: stored per order in `/var/lib/idea/orders.json` (`report` plain + `report_html`),
  with status/email/timestamps — the full record of paid tasks.

**Planned: optional PDF of the report.** To make the £29 spend more tangible we intend to also
deliver the report as a PDF (attachment or link). Constraint: the service is stdlib-only and builds
offline (`GOPROXY=off`), so a PDF generator means either relaxing that to vendor a library or
rendering the existing report HTML to PDF. Not built yet — see HANDOFF backlog.
