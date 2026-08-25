# DECISION 2026-08-25 — discretionary refunds: an operator CAPABILITY, never a term

**Owner, 2026-08-25 (evening), verbatim:**
> I will need the ability to refund a person's money if they're not happy with the site.
> I don't offer the service - it's not part of the terms, but I might well offer it readily
> if a site is not up to standard.

## 1. What this does NOT change — and why it must not

The public position stays exactly as attested: `no_refund` ("We do not offer refunds"),
the refund ban patterns, the writer_block's sanctioned phrasing. **Nothing about
discretion may enter the register.** The chat bot renders every fact's `claim` verbatim
into its prompt (`renderSystemPrompt`), and the claims gate reads the bans: a fact or
writer line saying "the owner may refund at discretion" would have customers told, within
five minutes, that refunds are possible. "Unadvertised" is enforced by exactly those two
readers, so the capability lives on the operator side only: code, console, this file.

## 2. What it DOES change — a prior ruling, narrowly

`internal/auth-service/billing/models.go:38` (PAY-009): *"Order statuses. Deliberately no
'refunded' — refunds are manual and unadvertised (owner ruling 2026-08-11); code must not
model them."* Statuses are `created | paid` only. Today's message keeps *unadvertised* and
asks for *the ability* — and an ability the platform cannot see is the defect class this
estate keeps paying for ("operators must remember X"): a refund issued in Stripe's
dashboard leaves `billing_orders.status='paid'`, the delivery path live, tokens valid, and
the customer holding the site and the ZIP. So the "code must not model them" clause is
what this decision supersedes, **pending the owner's choice below**.

## 3. Measured 2026-08-25 (the ground the options stand on)

- Refund code: **none**. 3 Go files mention "refund", all in comments.
- Readers of `billing_orders` outside `internal/auth-service/billing/`: **0** (the intake
  gate that will read paid state is future work, PAY-009 "consumers to notify"). So paid
  state gates nothing today; the only consequence chain a refund must unwind is delivery.
- Delivery/retraction: `retract_asset_files` exists (git adapter; 11 references) and
  `internal/core-manager/handlers/delivery.go:51` describes retracting a live site on
  schedule; the weekly chase / retraction JOB is not built (HANDOFF_2026-08-21 §6).
- Stripe: keys absent, webhook answers 503 (honest keyless state); `customer_access_tokens`
  **0**; nothing has ever been paid. **Not a launch gate** — it becomes required the day the
  Stripe keys land, alongside the second-click page and the delivery email.
- The judgement point already exists: the owner reviews brief + rendered site BEFORE the
  delivery email (`DECISION_2026-08-21e`, via the work-item queue). "Not up to standard"
  is decided there — so the cleanest refund is **refund-instead-of-deliver**: nothing was
  handed over, nothing to retract, one Stripe refund plus a cancelled handover.

## 4. Options, costed

**A — webhook-as-truth (recommended).** The owner refunds in Stripe's own dashboard (the
ability exists the day keys land, no build). auth-service handles one more event type,
`charge.refunded`, exactly as it handles `checkout.session.completed` (PAY-001's proven
pattern: fulfilment and un-fulfilment both follow signed events, never a button click):
`billing_orders.status='refunded'` + a `billing_events` row, and a consumer that cancels
the site's delivery (revoke `customer_access_tokens`, cancel handover/chase items, run
`retract_asset_files` if the site was already delivered). Supersedes the 08-11 clause
narrowly: code models a refund as an INBOUND fact and never offers one. Cost: a status,
an event branch, a consumer, tests; one council round (`internal/` scope).

**B — console button.** A "Refund" action on the order in `admin.apis.uk` → auth-service →
Stripe refunds API, plus A's bookkeeping. "Readily" suggests this eventually; it is A
plus an outbound call and a UI, and it is the surface where a mistaken click costs money,
so it wants the confirm-dialogue treatment the Terminate button got.

**C — keep 2026-08-11 literally.** Manual in Stripe, platform unaware; an operator must
also hand-cancel tokens, handover and items. Cheapest to build, and the failure mode is
silent: a refunded customer keeps a live site until someone remembers.

## 5. Where it lives

Payments surface = PAY-009 (`ai_site_selling_automation` lane built it); delivery/handover =
`site_delivery_and_editor` (joint with this lane). Whichever option is chosen gets its own
plan there and a council round; register the mechanism (a PAY entry) when it ships; tell
PAY-009's listed consumers what changed about the guarantee ("a paid order can become
refunded; delivery must check it").

## 6. Ruling — owner, 2026-08-25: **Option A**

Stripe dashboard + webhook. The owner refunds in Stripe; auth-service handles
`charge.refunded` as it handles `checkout.session.completed`; `billing_orders` gains a
`refunded` status; a consumer cancels delivery (tokens, handover/chase items, retraction
if already delivered). Refunds stay unadvertised: nothing enters the register or the copy.
This supersedes the 2026-08-11 "code must not model them" clause narrowly (inbound fact,
never an offered action). Build: its own plan + council round in the payments/delivery
lane, before Stripe keys land; not a launch gate. Option B (console button) stays open as
a later convenience on top of A.
