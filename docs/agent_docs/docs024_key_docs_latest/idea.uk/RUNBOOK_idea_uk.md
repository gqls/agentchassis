# idea.uk — runbook

The pieces:

| File | Role |
|---|---|
| `idea_method_runner.py` | The engine — runs the method (multi-lens generate → cut → web-verify → score → rank). Used by both front doors. |
| `idea_service.py` | The service — request-then-confirm intake, Stripe billing, fulfilment, internal endpoint. |
| `idea_uk_fakedoor.html` | The static landing/intake page (deploy to S3). Form posts to the service. |
| `test_idea_flow.py` | End-to-end state-machine test (FakeProvider + stubbed engine). |
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
pip install fastapi python-multipart anthropic httpx
python test_idea_flow.py        # expect: ALL 20 CHECKS PASSED
```

Uses the FakeProvider (no Stripe) and a stubbed engine (no LLM spend). Validates
the full state machine, idempotency, capacity, auth, and both front doors.

## Run it for real, locally (FakeProvider, real engine)

```bash
export ANTHROPIC_API_KEY=...           # real — the engine will spend
export INTERNAL_API_KEY=$(openssl rand -hex 16)
export AUTO_DELIVER=false
uvicorn idea_service:app --port 8080
# Internal run against one of our own domains (no billing):
curl -s localhost:8080/internal/run -H "X-Internal-Key: $INTERNAL_API_KEY" \
  -H 'content-type: application/json' \
  -d '{"domain":"agritec.uk","audience":"UK small farmers","assets":"curate scheme docs"}'
```

(No Stripe keys → FakeProvider; the external order flow still works locally:
`/request` → `/confirm` → visit the returned `/order/success?...&fake=1` URL to
simulate payment and trigger a real engine run.)

## Go-live checklist

1. **Build & deploy** the container: `docker build -t idea-svc . && docker run
   -p 8080:8080 -v idea_data:/data --env-file .env idea-svc`. Put it behind TLS.
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
- The sqlite db holds orders + subscribers + processed webhook events. Back it up
  (it's the record of who paid and what's owed). Mount `/data` on a volume.
- Sale-readiness: the engine takes assets as data and the billing sits behind a
  provider interface, so idea.uk remains a separable unit (PLAN_idea_uk.md §2).
