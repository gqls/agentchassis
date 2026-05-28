#!/usr/bin/env python3
"""
test_idea_flow.py — end-to-end check of the idea.uk service state machine.

Runs the whole flow with NO real Stripe (FakeProvider) and NO real LLM calls
(engine stubbed): request → confirm → pay(webhook) → fulfil → deliver, plus
decline, webhook idempotency, capacity throttle, and the internal path.

  pip install fastapi python-multipart anthropic httpx
  python test_idea_flow.py
"""
import os
import tempfile

# Env MUST be set before importing the service (module-level config reads it).
os.environ.setdefault("ANTHROPIC_API_KEY", "dummy-not-used-engine-is-stubbed")
os.environ["INTERNAL_API_KEY"] = "testkey"
os.environ["AUTO_DELIVER"] = "true"            # so a paid run ends 'delivered'
os.environ["PUBLIC_BASE_URL"] = "http://test"
os.environ["MAX_ACTIVE_ORDERS"] = "2"          # to exercise the capacity gate
os.environ["IDEA_DB_PATH"] = tempfile.mktemp(suffix=".db")
os.environ.pop("STRIPE_SECRET_KEY", None)      # force FakeProvider
os.environ.pop("STRIPE_WEBHOOK_SECRET", None)

import json
import idea_service
from fastapi.testclient import TestClient

# Stub the engine and capture deliveries instead of sending/writing.
idea_service.run_method = lambda d, a, s: f"# Stub report for {d}\nAudience: {a}\n"
SENT = []
idea_service.deliver = lambda to, subj, body: SENT.append((to, subj, body))

c = TestClient(idea_service.app)
OP = {"X-Internal-Key": "testkey"}
passed = 0


def ok(label, cond):
    global passed
    assert cond, f"FAIL: {label}"
    passed += 1
    print(f"  ok  {label}")


def make_request(email, business="acme.co.uk", audience="aud"):
    r = c.post("/request", data={"name": "Sam", "email": email,
                                 "business": business, "audience": audience,
                                 "notes": "n"})
    assert r.status_code == 200, r.text
    # recover the order id from the operator notification email
    subj = SENT[-1][1]
    return subj.split("NEW REQUEST ")[1].split(" ")[0]


print("health & capacity")
ok("health ok", c.get("/health").json()["ok"] is True)
cap = c.get("/capacity").json()
ok("capacity open initially", cap["open"] and cap["active"] == 0)

print("happy path: request -> confirm -> pay -> deliver")
SENT.clear()
oid = make_request("buyer@x.com")
ok("order is 'requested'", idea_service.get_order(oid)["status"] == "requested")
ok("operator notified of request", any("NEW REQUEST" in s[1] for s in SENT))

ok("confirm without key 401", c.post("/confirm", json={"order_id": oid}).status_code == 401)
SENT.clear()
r = c.post("/confirm", json={"order_id": oid}, headers=OP)
ok("confirm ok -> awaiting_payment", r.status_code == 200 and r.json()["status"] == "awaiting_payment")
ok("customer emailed a pay link", any("checkout.url" not in s[2] and "/order/success" in s[2] for s in SENT))

# Simulate Stripe firing the paid webhook.
SENT.clear()
evt = {"event_id": "evt_1", "type": "checkout.session.completed",
       "order_id": oid, "paid": True}
r = c.post("/stripe/webhook", content=json.dumps(evt))
ok("webhook accepted", r.json()["status"] == "accepted")
ok("order delivered (AUTO_DELIVER)", idea_service.get_order(oid)["status"] == "delivered")
ok("report emailed to buyer", any(s[0] == "buyer@x.com" and "Stub report" in s[2] for s in SENT))

print("webhook idempotency")
r = c.post("/stripe/webhook", content=json.dumps(evt))
ok("duplicate webhook ignored", r.json()["status"] == "duplicate_ignored")

print("decline path")
SENT.clear()
oid2 = make_request("decline@x.com")
r = c.post("/decline", json={"order_id": oid2, "reason": "no edge"}, headers=OP)
ok("declined", r.status_code == 200 and idea_service.get_order(oid2)["status"] == "declined")
ok("customer emailed decline", any(s[0] == "decline@x.com" for s in SENT))

print("capacity throttle (MAX_ACTIVE_ORDERS=2)")
a = make_request("a@x.com"); b = make_request("b@x.com"); d = make_request("d@x.com")
ok("confirm #1 ok", c.post("/confirm", json={"order_id": a}, headers=OP).status_code == 200)
ok("confirm #2 ok", c.post("/confirm", json={"order_id": b}, headers=OP).status_code == 200)
r = c.post("/confirm", json={"order_id": d}, headers=OP)
ok("confirm #3 blocked at capacity", r.status_code == 409)
ok("capacity now closed", c.get("/capacity").json()["open"] is False)

print("internal path (no billing)")
r = c.post("/internal/run", json={"domain": "ours.com", "audience": "x", "assets": "y"}, headers=OP)
ok("internal run returns report", r.status_code == 200 and "Stub report" in r.json()["report"])
ok("internal run needs key", c.post("/internal/run", json={"domain": "d", "audience": "a", "assets": "s"}).status_code == 401)

print("subscribe")
ok("subscribe ok", c.post("/subscribe", data={"email": "sub@x.com"}).status_code == 200)

print(f"\nALL {passed} CHECKS PASSED")
