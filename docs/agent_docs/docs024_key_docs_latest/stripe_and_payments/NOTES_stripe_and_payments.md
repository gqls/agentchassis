# NOTES — stripe and payments

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-09-04 — lane opened; the sweep, and what it found

**Instruction (owner, this session):** take responsibility for all Stripe and related
transactions, including voucher creation; check which other lanes touch it and take it
from them.

### Lanes found touching payments, and what was done about each

| lane | what it touches | disposition |
|---|---|---|
| `site_delivery_and_editor` | minted both vouchers; owns the delivery letter that quotes prices | **handed over**, confirmed by its session; keeps site/delivery/editor |
| `webdesign_uk_build_service` | built the PAY-009 surface, owns `trial_checkout.sh`, the go-live checklist | absorbed (no live session to notify) |
| `client_accounts` | reads the billing tables; owns identity | boundary agreed — money mine, identity theirs |
| `finetuning_uk_service` | "Stripe his, and last" | acknowledged; will bring one-off booking + £99 charges here when real |
| `idea_uk_vm_site` | a **separate** Stripe account, separate codebase, on a VM | claimed; awaiting their reply on what is current |
| order-intake (P4/PAY-010) | the paid gate reading `billing_orders` | paid gate mine, brief/build half theirs |

### Measurements taken today (all `[MEASURED 2026-09-04]`)

- `vouchers` = **2**. `WD-9FAB5-2NVNF` (redeemed 08-27), `WD-KN3WU-9PZN4` (£30, expires
  2026-09-18, unredeemed).
- `billing_orders` = **1**, paid, £30, `BR-9AUZ59`, paid_at 2026-08-27 14:40:22Z.
- `billing_events` = **1**.
- `billing_settings.payment_timing` = `upfront` (since 2026-08-17).
- `scheduled_tasks.order-intake-collect` = **enabled**, 900s, last fired 14:32:18Z.
- `build_queue`: `boxingonline.com` seeded 2026-08-31 — the collector's one release.
- `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET` present in `personae-platform-secrets`
  **and** in terraform `047-base-configs` (so they survive a release — the 2026-08-26
  wipe is durably fixed). Key names read; **values never read into this session.**
- `POST https://webdesign.uk/stripe/webhook` → **400** (keyed refusal, the correct
  reading). Control: apex → 200.

### Finding 1 — the voucher minted today is sound, and I proved it rather than assuming

The delivery lane minted `WD-KN3WU-9PZN4` **by raw SQL**, not through the admin API
(it holds no admin JWT). It said so unprompted, which is the right call. But a voucher
that honours the guards rather than being checked by them is exactly the thing to verify
before the owner spends a trial on it.

The redemption predicate is `repository.go:123` —
`WHERE code = $1 AND redeemed_at IS NULL AND expires_at > now()`.

Ran that exact statement in a transaction and **rolled back**, with two controls:

| case | expected | got |
|---|---|---|
| `WD-KN3WU-9PZN4` | 1 row | **1 row**, £30 |
| `WD-XXXXX-XXXXX` (does not exist) | 0 rows | 0 rows |
| `WD-9FAB5-2NVNF` (already redeemed) | 0 rows | 0 rows |
| post-rollback re-read | still unredeemed | still unredeemed |

**The voucher will redeem for £30.** The controls matter: without them a "1 row" result
proves only that the statement ran, not that it discriminates.

### Finding 2 — I asserted something false to another lane, and it was refuted the same hour

In my handover message to `client_accounts` I wrote that *"a paid order is currently the
only thing that mints a durable link between a real person and a client row"*, citing
PAY-009's `clients.external_id` first-writer-wins design.

**That is wrong, and the register entry describing the intent is what misled me.** The
`client_accounts` lane refuted it with measurements; I re-verified all of them myself
rather than accept the report:

- `clients.external_id` on the only real customer row (Boxing Online, `a7395f69…`) is
  **EMPTY**.
- `billing_orders.provider_customer_id` on the one paid order is **NULL**.
- The stored webhook payload says why: `"customer": null`,
  `"customer_creation": "if_required"`, `"mode": "payment"`. **Stripe creates no Customer
  object for a one-off payment in that shape**, so there was never a customer id to land.

The plumbing worked perfectly — voucher redeemed 14:39:01Z, `cs_live_` session,
signature-verified event processed 14:40:22Z. There was simply nothing to write.

**So the seam is inverted from how PAY-009 describes it.** The durable link today is
`metadata.order_id → billing_orders.id → billing_orders.client_id`, and that `client_id`
was resolved **before** the payment. Payment currently *confirms* a client row; it does
not mint one. `external_id` is not load-bearing.

**The cheap check I skipped:** I quoted a register entry's stated intent as if it were
behaviour. `LANDMINES.md` already carries this exact trap — *"a concept-register STATUS
line is a snapshot that outlives its truth"* — and it fired on me within an hour of the
SessionStart hook showing it to me. The check was one query against the column the entry
names. **A register entry describes what was designed; only the column says what happened.**

This matters beyond the correction: `external_id` **starts firing** the moment anything
moves to `mode=subscription` or `customer_creation: "always"` — i.e. exactly when the
£10/month rental becomes real. That is a decision to take deliberately, not discover.

### Finding 3 — the live checkout page is branded as the wrong product

Chasing Finding 2 through the stored payload, the same object carries:

```
"branding_settings": { "display_name": "Fine Tune", ... "button_color": "#0074d4" }
```

`"livemode": true`. **A webdesign.uk customer paying for a website is shown a payment page
headed "Fine Tune"** — a different product on a different domain. The owner saw it himself
on 2026-08-27 and it did not register as a defect, which is unsurprising: he knows both
products are his.

Reported by the `client_accounts` lane as an aside; verified here from the payload rather
than taken on report. It is a Stripe **account** setting, so it is one dashboard field and
costs nothing to fix — but nothing in the codebase will ever tell you about it, because
the account's branding is not in the codebase.

**Both of this payment page's customer-facing defects are invisible to every check we
have**: the page is branded wrong, and the page it returns to 404s. Neither is reachable
from any test, because both live outside the repository.

### Finding 4 — `/pay/success` and `/pay/cancel` are still 404, eight days on

`stripe.go:55-56` mints them as every checkout's landing. Confirmed in the stored payload
of the real payment (`success_url` = `https://webdesign.uk/pay/success?o=36744bf0…`), and
confirmed live today: **both 404**. Found 2026-08-27 by the trial run, unfixed since.
Nothing loses money — the webhook is the truth — but a customer who has just paid real
money lands on a bare 404 branded as another company.

### Carried in from the delivery lane (do not re-litigate)

- £149 list; voucher variants £10/£30/£55 **only**; £30 added 2026-08-26 specifically so
  the owner's own trials exercise the real voucher path. A £0 code is wrong twice over.
- **No `refunded` status by ruling** — `models.go`: *"code must not model them"*.
- Domain prices (£10/mo, £59.99) live in the delivery email's `body_template`, which is a
  step config on an **agent definition** — outside any census of "every £200 in the live
  specs". Migration `726` corrected a superseded £200 there for that reason.
- ⚠ **Live collision**: migration `776` moved that `body_template` today at 12:05:25Z.
  Session `475` is taking `{{instructions_link}}` in the same jsonb path. `bugs_open/477`
  has `774`/`775` written-and-unapplied, including a **second** letter with its own
  `body_template`. Two letters will soon quote this lane's prices.

### Stale claims to correct elsewhere (found, not yet actioned)

Several lane docs still say selling is blocked on Stripe keys — e.g.
`webdesign_uk_build_service/PLAN_2026-08-21_todo_from_here.md` row 1. True when written;
**false since 2026-08-27**, when a real card payment cleared on a live key. Reported by
`client_accounts`; correcting it is this lane's.

### What I could not measure, and why

I tried to enumerate the Stripe account's **Payment Links** (the £10/month rental, the
Customer Portal, the £59.99 buy-out) by calling `GET /v1/payment_links` from inside the
auth-service pod — the sanctioned shape, so the key is never read into the session. The
call was **refused by this session's own permission classifier**. I did not work around it.

So whether those links exist is **`[UNMEASURED]`**. This is load-bearing: the delivery
letter promises the rental link "arrives in your delivery email". If it does not exist,
the letter promises something the machinery cannot do.

---

## 2026-09-04 (later) — the branding question answered: per-checkout override exists

Owner asked, on reading Finding 3: *"That is the name of the account — can we have sub
names? or set up a different name into the same account?"*

**Yes, and it needs no new account.** Stripe supports a per-Checkout-Session branding
override — `branding_settings` on session create, which overrides the Dashboard defaults
for that one session. Sub-fields: `display_name`, `logo`, `icon`, `background_color`,
`button_color`, `font_family`, `border_style`. **Any field omitted falls back to the
Dashboard value**, so overriding only the name is legitimate and leaves everything else
alone.

Docs: https://docs.stripe.com/payments/checkout/customization/appearance?payment-ui=stripe-hosted

**This fits our code exactly.** `stripe.go` builds a raw `url.Values` form POST (no SDK),
and the parameter is form-encoded — so it is **one `form.Set` line**:

```go
form.Set("branding_settings[display_name]", brandName)
```

**The API version is already new enough, and I can prove it rather than assume it.** We
send no `Stripe-Version` header, so Stripe uses the account's default version. The webhook
payload we stored on 2026-08-27 **already contains a `branding_settings` object** — an API
version that did not know the field could not have returned it. So no version bump is
needed and no pin has to change.

### Two limits that matter, both from the docs

1. **Invoices and receipts still use the Dashboard branding.** The docs say so explicitly.
   So a per-session override fixes the *checkout page* and not the customer's receipt. The
   account-level name is therefore still a real decision, not one the override retires.
2. **Card-network rules want the business name to be accurate and consistent**, and Stripe
   warns that inconsistency raises chargeback risk. So the checkout name, the receipt name
   and the **bank statement descriptor** should agree.

⚠ **We currently set no statement descriptor at all** (`stripe.go` sets none), so the
buyer's bank statement shows the account default. If we brand checkout per-product, that
becomes a third name in play. `payment_intent_data[statement_descriptor_suffix]` is the
lever, and it is the same one-line shape.

### The alternative, for completeness

A **separate Stripe account per brand** — which is what idea.uk already does. It gives
total separation of branding, payouts and reporting, at the cost of a second key pair, a
second webhook endpoint and secret, and separate reconciliation. Given that keys dying at
release has already bitten this estate once (2026-08-26), doubling the key surface is a
real cost. **Per-session branding is the cheaper and more reversible answer** unless the
owner wants the payouts separated.

**Not implemented yet — the names themselves are the owner's call**, and the account-level
name is entangled with receipts, so it is not a decision to take by inference.
