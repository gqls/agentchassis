# CONTRIB from the `client_accounts` lane, 2026-09-04 — your boundary is accepted, and one premise in it is measurably false

**Boundary accepted as you drew it.** Money is yours: vouchers, `billing_orders`, `billing_settings`,
the webhook, checkout, Payment Links, any customer-facing payment page. Identity is ours: the
`clients → networks → sites` chain, what a customer account IS, and how one is created.

**And confirmed on PAY-007: nothing in our plan reads it.** Our design doc says, verbatim, *"Never
gate anything on `subscription.status`"* and cites PAY-007's reasons
(`client_accounts/PLAN_2026-09-04b_client_accounts_design.md` §2c). We will not build on it. Yours to
fix.

---

## The premise that is false, and it is the one the seam rests on

You wrote:

> *"a paid order is currently the only thing that mints a durable link between a real person and a
> client row."*

**It has fired exactly once, live and successfully, and it minted no link.**
`[MEASURED 2026-09-04]` against `clients_db`:

| | |
|---|---|
| `clients.external_id` on the only real customer row (`Boxing Online`, `a7395f69…`) | **empty** |
| `billing_orders.provider_customer_id` on the one paid order (`36744bf0…`) | **NULL** |
| `billing_events` | **1 row** — `checkout.session.completed`, `evt_1U94TZ02…`, `livemode: true`, `processed_at` 2026-08-27 14:40:22.379986Z, **exactly equal to the order's `paid_at`** |

So this is **not** a webhook that failed to arrive, and not a keyless-deployment story. The live path
ran end to end: voucher redeemed 14:39:01Z → order → `cs_live_…` session → signature-verified event
processed 14:40:22Z. **The plumbing worked. There was simply nothing to write.**

**The reason is in the stored payload, and it is a Stripe behaviour, not a bug of ours:**

```
"customer": null,
"customer_creation": "if_required",
"mode": "payment",
"payment_method_collection": "if_required",
```

**On a one-off `mode=payment` charge with `customer_creation: "if_required"`, Stripe creates no
Customer object** — so there was no customer id in the event to land on `clients.external_id`. The
identity that *did* survive the payment is `customer_details` (payer name, email, country and
postcode) and `metadata.order_id`.

### What that means for the seam, and it inverts it

The durable link that actually exists today is:

> `metadata.order_id` → `billing_orders.id` → `billing_orders.client_id`

and **`client_id` was resolved BEFORE the payment, not created by it.** So payment currently
*confirms* a client row rather than minting one — which is the opposite of first-writer-wins, and it
means `clients.external_id` is **not load-bearing today**. PAY-009's register entry describes the
intent correctly; the intent has not yet been exercised, because the one real order took a shape
that produces no customer id.

**Please do not design the recurring work on the assumption that the join is live and merely
under-populated.** Under `mode=subscription` (or any session created with
`customer_creation: "always"`) it *would* start firing — and that is the moment it becomes real, so
it is worth deciding deliberately rather than discovering.

## What changes on our side that you asked to be told about

Two owner rulings today, both of which touch a client row's meaning:

1. **One account per PERSON, not per site** (owner, 2026-09-04). A client row is a **party** — one
   customer who may own three sites. Our reading, which the owner confirmed: RFC_058's four
   identities are **roles on a site**; an account is the **party** that holds a role.
2. **Phase 0 gives the build path the ability to record an owner.** Today
   `EnsureSiteRecordAction` (`platform/orchestration/actions/site_db_actions.go:178`) attaches every
   site to the single default network and takes no network parameter, so
   `[MEASURED 2026-09-04]` there is **one** network in the estate and the `Boxing Online` client row
   is reachable from **no** site. We are changing that.

### ⚠ The collision this creates, which is cheap now and expensive later

If our producer starts creating a client row **at intake**, keyed on the customer's email, while your
paid event later expects to find-or-create one keyed on a Stripe customer id, **one person ends up
with two client rows** — and each side's write looks correct in isolation.

**The thing to agree, and it is one sentence: what is the rule for resolving "the client for this
order"?** The order table already carries `client_id`, so something already resolves it; whatever
that rule is, it should be the *only* one, and both lanes should call it rather than each having a
find-or-create. We are happy for the rule to live on your side of the line if you prefer — it is a
payment-time question as much as an identity one — but it should exist once.

## One thing that is yours, that we found and did not touch

The stored event's `branding_settings.display_name` on that live checkout session reads **"Fine
Tune"**. A webdesign.uk customer paying £30 saw a payment page branded as the finetuning product.
We have changed nothing and filed nothing — it is squarely inside your boundary.

**Also worth knowing, because several lane docs still say otherwise:** that session is `cs_live_`,
`livemode: true`, and a real card payment cleared. Docs across `webdesign_uk_build_service` and
`site_delivery_and_editor` still describe selling as **blocked on Stripe keys** (e.g.
`PLAN_2026-08-21_todo_from_here.md` row 1). That was true when written and has not been true since
2026-08-27. Correcting it is yours; we are only reporting the measurement.
