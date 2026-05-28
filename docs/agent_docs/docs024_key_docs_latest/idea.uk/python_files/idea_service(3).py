#!/usr/bin/env python3
"""
idea_service.py
───────────────
Service layer for idea.uk: turns the ideation method (idea_method_runner.run)
into a working tool with two front doors and one-off billing.

  • EXTERNAL (idea.uk customers): request → operator confirms or declines →
    pay (Stripe one-off) → run → deliver.   (request-then-confirm flow)
  • INTERNAL (our own domains): authenticated /internal/run, no billing.

Both call the SAME engine; assets are passed in as data, so the engine never
embeds our assets (keeps idea.uk sale-ready — PLAN_idea_uk.md §2).

Why request-then-confirm (matches the landing page): it lets the operator screen
a request BEFORE taking money — declining politely where the method would return
"no candidate advances" — which protects the 72h/refund promise and quality
during the manual early-access phase.

Billing follows PLAN_stripe_billing_integration.md principles, lightweight:
  • WEBHOOK is the source of truth — payment never trusted from the browser;
  • provider behind an interface (Stripe first, swappable);
  • idempotent webhooks (dedup on event id).

AUTO_DELIVER defaults OFF: a paid run is held for operator review before sending.

Run:
  pip install fastapi uvicorn python-multipart "stripe>=9" anthropic
  export ANTHROPIC_API_KEY=...                 # engine
  export STRIPE_SECRET_KEY=sk_...               # omit → Fake provider (testing)
  export STRIPE_WEBHOOK_SECRET=whsec_...
  export PUBLIC_BASE_URL=https://idea.uk
  export INTERNAL_API_KEY=$(openssl rand -hex 16)   # internal + operator actions
  export OPERATOR_EMAIL=you@idea.uk
  uvicorn idea_service:app --port 8080

Stripe calls use long-stable patterns; pin the stripe lib version and check
current docs before going live.
"""

import os
import json
import sqlite3
import smtplib
import logging
from email.message import EmailMessage
from datetime import datetime, timezone
from typing import Protocol, Optional

from fastapi import FastAPI, Request, BackgroundTasks, Header, HTTPException, Form
from fastapi.responses import HTMLResponse, JSONResponse

from idea_method_runner import run as run_method     # the engine, imported not duplicated

log = logging.getLogger("idea_service")
logging.basicConfig(level=logging.INFO)

# ── CONFIG ───────────────────────────────────────────────────────────────────
PRICE_GBP        = int(os.environ.get("REPORT_PRICE_GBP", "199"))
AUTO_DELIVER     = os.environ.get("AUTO_DELIVER", "false").lower() == "true"
PUBLIC_BASE_URL  = os.environ.get("PUBLIC_BASE_URL", "http://localhost:8080")
INTERNAL_API_KEY = os.environ.get("INTERNAL_API_KEY", "")
OPERATOR_EMAIL   = os.environ.get("OPERATOR_EMAIL", "ops@idea.uk")
DB_PATH          = os.environ.get("IDEA_DB_PATH", "idea_orders.db")
# Capacity: max orders in flight (awaiting_payment..awaiting_review) before we
# stop confirming new ones — protects the 72h delivery promise during the manual
# review phase. Raise once AUTO_DELIVER is on and trusted.
MAX_ACTIVE       = int(os.environ.get("MAX_ACTIVE_ORDERS", "8"))
# CORS: origins allowed to call the service (the static page's origin). Comma-sep.
ALLOWED_ORIGINS  = [o.strip() for o in os.environ.get(
    "ALLOWED_ORIGINS", PUBLIC_BASE_URL).split(",") if o.strip()]


# ── BILLING PROVIDER (interface + Stripe + Fake) ─────────────────────────────
class CheckoutResult:
    def __init__(self, provider_session_id: str, url: str):
        self.provider_session_id, self.url = provider_session_id, url


class WebhookEvent:
    """Normalised event — the service speaks this, not Stripe."""
    def __init__(self, event_id, type, order_id, paid):
        self.event_id, self.type, self.order_id, self.paid = event_id, type, order_id, paid


class BillingProvider(Protocol):
    def create_checkout(self, order_id: str, email: str) -> CheckoutResult: ...
    def parse_webhook(self, payload: bytes, sig_header: str) -> WebhookEvent: ...


class StripeProvider:
    def __init__(self, secret_key, webhook_secret):
        import stripe
        self.stripe = stripe
        self.stripe.api_key = secret_key
        self.webhook_secret = webhook_secret

    def create_checkout(self, order_id, email):
        s = self.stripe.checkout.Session.create(
            mode="payment",
            customer_email=email or None,
            line_items=[{
                "price_data": {
                    "currency": "gbp",
                    "unit_amount": PRICE_GBP * 100,
                    "product_data": {
                        "name": "idea.uk — verified AI opportunity report",
                        "description": "Ranked, web-verified candidate ideas for your business.",
                    },
                },
                "quantity": 1,
            }],
            metadata={"order_id": order_id},
            success_url=f"{PUBLIC_BASE_URL}/order/success?o={order_id}",
            cancel_url=f"{PUBLIC_BASE_URL}/order/cancel?o={order_id}",
        )
        return CheckoutResult(s["id"], s["url"])

    def parse_webhook(self, payload, sig_header):
        e = self.stripe.Webhook.construct_event(payload, sig_header, self.webhook_secret)
        obj = e["data"]["object"]
        paid = (e["type"] == "checkout.session.completed"
                and obj.get("payment_status") == "paid")
        return WebhookEvent(e["id"], e["type"],
                            (obj.get("metadata") or {}).get("order_id"), paid)


class FakeProvider:
    """No Stripe keys → local testing only. NEVER use in production."""
    def create_checkout(self, order_id, email):
        return CheckoutResult(f"fake_{order_id}",
                              f"{PUBLIC_BASE_URL}/order/success?o={order_id}&fake=1")

    def parse_webhook(self, payload, sig_header):
        b = json.loads(payload or b"{}")
        return WebhookEvent(b.get("event_id", "evt_fake"),
                            b.get("type", "checkout.session.completed"),
                            b.get("order_id"), bool(b.get("paid", True)))


def make_provider() -> BillingProvider:
    sk, wh = os.environ.get("STRIPE_SECRET_KEY"), os.environ.get("STRIPE_WEBHOOK_SECRET")
    if sk and wh:
        return StripeProvider(sk, wh)
    log.warning("No Stripe keys — using FakeProvider (local/testing only).")
    return FakeProvider()


provider: BillingProvider = make_provider()


# ── STORE (sqlite) ───────────────────────────────────────────────────────────
# Order lifecycle: requested → (declined | awaiting_payment) → paid → running
#                  → (awaiting_review | delivered | failed)
def db():
    c = sqlite3.connect(DB_PATH)
    c.row_factory = sqlite3.Row
    return c


def init_db():
    with db() as c:
        c.execute("""CREATE TABLE IF NOT EXISTS orders (
            id TEXT PRIMARY KEY, name TEXT, email TEXT,
            domain TEXT, audience TEXT, assets TEXT,
            status TEXT, report TEXT, provider_session_id TEXT,
            created_at TEXT, updated_at TEXT)""")
        c.execute("""CREATE TABLE IF NOT EXISTS webhook_events (
            event_id TEXT PRIMARY KEY, processed_at TEXT)""")
        c.execute("""CREATE TABLE IF NOT EXISTS subscribers (
            email TEXT PRIMARY KEY, created_at TEXT)""")


def active_count() -> int:
    """Orders occupying a fulfilment slot (committed work not yet finished)."""
    with db() as c:
        return c.execute(
            "SELECT COUNT(*) FROM orders WHERE status IN "
            "('awaiting_payment','paid','running','awaiting_review')").fetchone()[0]


def now():
    return datetime.now(timezone.utc).isoformat()


def new_id():
    return "ord_" + os.urandom(8).hex()


def save_order(o):
    with db() as c:
        c.execute("""INSERT INTO orders
            (id,name,email,domain,audience,assets,status,report,provider_session_id,created_at,updated_at)
            VALUES (?,?,?,?,?,?,?,?,?,?,?)""",
            (o["id"], o["name"], o["email"], o["domain"], o["audience"], o["assets"],
             o["status"], o.get("report"), o.get("provider_session_id"),
             o["created_at"], o["updated_at"]))


def get_order(oid):
    with db() as c:
        r = c.execute("SELECT * FROM orders WHERE id=?", (oid,)).fetchone()
        return dict(r) if r else None


def update_order(oid, **f):
    f["updated_at"] = now()
    sets = ", ".join(f"{k}=?" for k in f)
    with db() as c:
        c.execute(f"UPDATE orders SET {sets} WHERE id=?", (*f.values(), oid))


def event_seen(event_id) -> bool:
    with db() as c:
        try:
            c.execute("INSERT INTO webhook_events VALUES (?,?)", (event_id, now()))
            return False
        except sqlite3.IntegrityError:
            return True


# ── DELIVERY (smtp, with file/log fallback) ──────────────────────────────────
def deliver(to_email, subject, body):
    host = os.environ.get("SMTP_HOST")
    if not host:
        path = f"delivered_{to_email.replace('@','_at_')}_{int(datetime.now().timestamp())}.md"
        with open(path, "w") as fh:
            fh.write(body)
        log.info("SMTP not configured — wrote %s", path)
        return
    m = EmailMessage()
    m["From"] = os.environ.get("SMTP_FROM", OPERATOR_EMAIL)
    m["To"], m["Subject"] = to_email, subject
    m.set_content(body)
    with smtplib.SMTP(host, int(os.environ.get("SMTP_PORT", "587"))) as s:
        s.starttls()
        if os.environ.get("SMTP_USER"):
            s.login(os.environ["SMTP_USER"], os.environ["SMTP_PASS"])
        s.send_message(m)
    log.info("Emailed %s", to_email)


# ── FULFILMENT ───────────────────────────────────────────────────────────────
def fulfil(oid):
    o = get_order(oid)
    if not o or o["status"] not in ("paid", "running"):
        log.warning("fulfil: %s not fulfillable (%s)", oid, o and o["status"])
        return
    update_order(oid, status="running")
    try:
        report = run_method(o["domain"], o["audience"], o["assets"] or "")
    except Exception as e:
        log.exception("engine failed %s", oid)
        update_order(oid, status="failed")
        deliver(OPERATOR_EMAIL, f"[idea.uk] RUN FAILED {oid}",
                f"Order {oid} engine error: {e}\nReview & refund per the 72h promise.")
        return
    if AUTO_DELIVER:
        update_order(oid, status="delivered", report=report)
        deliver(o["email"], "Your idea.uk report", report)
    else:
        update_order(oid, status="awaiting_review", report=report)
        deliver(OPERATOR_EMAIL, f"[idea.uk] REVIEW {oid} ({o['domain']})",
                f"Paid order ready for review before sending to {o['email']}.\n\n"
                f"--- DRAFT REPORT ---\n\n{report}")


# ── APP ──────────────────────────────────────────────────────────────────────
app = FastAPI(title="idea.uk service")

from fastapi.middleware.cors import CORSMiddleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=ALLOWED_ORIGINS,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

init_db()


def require_operator(key):
    if not INTERNAL_API_KEY or key != INTERNAL_API_KEY:
        raise HTTPException(401, "unauthorised")


@app.get("/health")
def health():
    return {"ok": True, "auto_deliver": AUTO_DELIVER, "price_gbp": PRICE_GBP,
            "provider": type(provider).__name__}


@app.get("/capacity")
def capacity():
    """Public: whether we can take on new confirmed work right now. The static
    page can fetch this to show a 'currently full' notice."""
    a = active_count()
    return {"open": a < MAX_ACTIVE, "active": a, "max": MAX_ACTIVE}


@app.post("/subscribe", response_class=HTMLResponse)
async def subscribe(email: str = Form(...)):
    """Lower-commitment list capture (the page's updates strip)."""
    with db() as c:
        try:
            c.execute("INSERT INTO subscribers VALUES (?,?)", (email, now()))
        except sqlite3.IntegrityError:
            pass
    return "<h1>Thanks</h1><p>We'll send the next worked example.</p>"


@app.post("/request", response_class=HTMLResponse)
async def request_report(
    name: str = Form(...), email: str = Form(...),
    business: str = Form(...), audience: str = Form(...),
    notes: str = Form(default=""),
):
    """Public intake (form-encoded, matches the landing page). No payment.
    Stores the request and notifies the operator to confirm or decline."""
    oid = new_id()
    save_order({
        "id": oid, "name": name, "email": email,
        "domain": business, "audience": audience, "assets": notes,
        "status": "requested", "report": None, "provider_session_id": None,
        "created_at": now(), "updated_at": now(),
    })
    deliver(OPERATOR_EMAIL, f"[idea.uk] NEW REQUEST {oid} ({business})",
            f"From: {name} <{email}>\nBusiness: {business}\nAudience: {audience}\n"
            f"Notes: {notes}\n\nConfirm:  POST /confirm  {{'order_id':'{oid}'}}\n"
            f"Decline:  POST /decline  {{'order_id':'{oid}','reason':'...'}}")
    return ("<h1>Request received</h1><p>We'll reply within 24 hours with a "
            "confirmed slot or a polite decline. No charge yet.</p>")


@app.post("/confirm")
async def confirm(request: Request, x_internal_key: str = Header(default="")):
    """Operator action: confirm a request → create the Stripe Checkout and email
    the customer the payment link. Payment happens only after this."""
    require_operator(x_internal_key)
    oid = (await request.json()).get("order_id")
    o = get_order(oid)
    if not o:
        raise HTTPException(404, "no such order")
    if o["status"] != "requested":
        raise HTTPException(409, f"order is {o['status']}, not 'requested'")
    if active_count() >= MAX_ACTIVE:
        # Don't over-commit beyond what manual review can deliver in 72h.
        raise HTTPException(409, {"error": "at_capacity", "active": active_count(),
                                  "max": MAX_ACTIVE,
                                  "hint": "wait for a slot to free, or raise MAX_ACTIVE_ORDERS"})
    checkout = provider.create_checkout(oid, o["email"])
    update_order(oid, status="awaiting_payment",
                 provider_session_id=checkout.provider_session_id)
    deliver(o["email"], "Your idea.uk report — confirmed, ready to pay",
            f"Hi {o['name']},\n\nWe can do a useful job on this. To start your "
            f"report (£{PRICE_GBP}), pay here:\n\n{checkout.url}\n\n"
            f"Delivered within 72 hours of payment. Full refund if we can't "
            f"find anything worth acting on.")
    return {"status": "awaiting_payment", "checkout_url": checkout.url}


@app.post("/decline")
async def decline(request: Request, x_internal_key: str = Header(default="")):
    """Operator action: politely decline a request (e.g. method would return
    'no candidate advances'). No charge was ever made."""
    require_operator(x_internal_key)
    body = await request.json()
    o = get_order(body.get("order_id"))
    if not o:
        raise HTTPException(404, "no such order")
    update_order(o["id"], status="declined")
    deliver(o["email"], "About your idea.uk request",
            f"Hi {o['name']},\n\nThanks for the request. Honestly, we don't think "
            f"we'd produce something worth £{PRICE_GBP} for this right now — "
            f"{body.get('reason', 'the differentiator is not strong enough yet')}. "
            f"No charge, and we'd rather say so than sell you a weak report.")
    return {"status": "declined"}


@app.post("/stripe/webhook")
async def stripe_webhook(request: Request, background: BackgroundTasks,
                         stripe_signature: str = Header(default="")):
    """Source of truth. Verify, dedup, and on a paid event for an
    awaiting_payment order, mark paid and fulfil in the background."""
    payload = await request.body()
    try:
        evt = provider.parse_webhook(payload, stripe_signature)
    except Exception as e:
        raise HTTPException(400, f"bad webhook: {e}")
    if event_seen(evt.event_id):
        return {"status": "duplicate_ignored"}
    if not (evt.paid and evt.order_id):
        return {"status": "ignored", "type": evt.type}
    o = get_order(evt.order_id)
    if not o:
        return {"status": "unknown_order"}
    if o["status"] in ("paid", "running", "awaiting_review", "delivered"):
        return {"status": "already_processed"}
    if o["status"] != "awaiting_payment":
        return {"status": f"unexpected_state:{o['status']}"}
    update_order(evt.order_id, status="paid")
    background.add_task(fulfil, evt.order_id)
    return {"status": "accepted"}


@app.post("/internal/run")
async def internal_run(request: Request, x_internal_key: str = Header(default="")):
    """Internal front door: run the engine for our own domains, no billing.
    Same engine as the paid path; assets passed in as data."""
    require_operator(x_internal_key)
    d = await request.json()
    for k in ("domain", "audience", "assets"):
        if not d.get(k):
            raise HTTPException(400, f"missing field: {k}")
    return JSONResponse({"report": run_method(d["domain"], d["audience"], d["assets"])})


@app.get("/order/success", response_class=HTMLResponse)
def order_success(o: str = "", fake: str = ""):
    if fake:                                   # FakeProvider local-test shortcut
        order = get_order(o)
        if order and order["status"] == "awaiting_payment":
            update_order(o, status="paid")
            fulfil(o)
    return ("<h1>Payment received</h1><p>Your report is being prepared and will "
            "arrive by email within 72 hours. If we can't find anything worth "
            "acting on, you're refunded.</p>")


@app.get("/order/cancel", response_class=HTMLResponse)
def order_cancel(o: str = ""):
    return "<h1>Order cancelled</h1><p>No charge was made.</p>"


# Internal CLI: python idea_service.py internal "<domain>" "<audience>" "<assets>"
if __name__ == "__main__":
    import sys
    if len(sys.argv) >= 5 and sys.argv[1] == "internal":
        print(run_method(sys.argv[2], sys.argv[3], sys.argv[4]))
    else:
        print(__doc__)
