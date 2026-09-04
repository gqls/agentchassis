# PLAN 2026-09-04 — Stripe and payments, as one owned lane

**Lane opened 2026-09-04 on the owner's instruction:** this thread is responsible for all
Stripe and related transactions across the estate — vouchers, orders, checkout, the
webhook, Payment Links, and any customer-facing payment page. Other lanes route payment
work here rather than doing it.

> Why a lane at all: until today the payment surface was built and operated by whichever
> lane needed it that week — the webdesign.uk build service built it, the delivery lane
> minted the vouchers, idea.uk runs a second, unrelated Stripe account on a VM. Nothing
> was wrong with any of that individually. What it produced is an estate where **no single
> document says what can take money and what state it is in**, and where a voucher can
> exist in the live table for forty minutes without any doc explaining it (which is what
> happened this afternoon, and is how this lane found its first fact).

## 1. What this lane owns

**Owned outright:**
- `internal/auth-service/billing/` — the whole package (provider, stripe, models,
  repository, service, handlers) and its wiring in `cmd/auth-service/main.go`.
- The four clients_db billing tables (migration 391): `vouchers`, `billing_orders`,
  `billing_events`, `billing_settings`.
- The Stripe accounts themselves: keys, restricted-key scopes, webhook endpoints,
  Payment Links, the dashboard.
- Any customer-facing payment page: `/pay/success`, `/pay/cancel`, and whatever
  replaces or joins them.
- `platform/orchestration/actions/collect_external_orders_action.go` **only where it
  reads payment state** — the paid gate. The brief/build half stays with the order-intake
  lanes.
- idea.uk's separate Stripe surface on the VM (`idea-go`), which is a different account
  and a different codebase but is still money.
- `internal/auth-service/subscription/` — the PAY-007 scaffold, because it is the thing
  that most looks like a payment system and is not one (see §4).

**Explicitly NOT owned, and where it goes instead:**
- Prices as *facts and copy* — the £149, the £10/month, the £59.99, finetuning's
  £3.15/£6.65. **Setting a price is the owner's; publishing it is the owning site's lane.**
  This lane owns only what the machinery charges. See §5, which is where that gets
  uncomfortable.
- Client identity, the clients→networks→sites chain, what a customer account *is* —
  the `client_accounts` lane.
- Site delivery, the editor, the delivery email as a product — `site_delivery_and_editor`.
- Refunds as a *business decision* — the owner's, and by his 2026-08-11 ruling they are
  manual, dashboard-only and unadvertised. This lane's job is to not model them.

## 2. Design position, inherited and not up for re-litigation

These are owner rulings already made. They are recorded here so this lane does not
re-open them and does not accidentally contradict them in code.

| ruling | date | what it means for code |
|---|---|---|
| £149 all-in list price | 2026-08-11 | `ListPricePence`; the amount is computed server-side, never taken from the client |
| Voucher variants are £10 / £30 / £55 **only** | 2026-08-11, £30 added 2026-08-26 | `RuledVoucherPences`; `Service.CreateVoucher` refuses anything else. A £0 voucher is wrong twice — not a ruled amount, and it would skip the payment path a trial exists to exercise |
| Vouchers are single-use, named, expiring | 2026-08-11 | the invariant IS the `UPDATE … WHERE redeemed_at IS NULL … RETURNING`; one transaction can move it, a raced second gets zero rows |
| **No `refunded` order status** | 2026-08-11 | `models.go` says it outright: refunds are manual and unadvertised, *"code must not model them"*. Statuses are `created \| paid` and nothing else |
| Payment timing = `upfront` | 2026-08-17 | pay before build, no customer preview first. `billing_settings.payment_timing` |
| Buy-out £59.99, rental £10/month | 2026-08-26 | superseded an earlier £200; fact sources legitimately still quote £200, so a price sweep must scope to writer-visible surfaces, never whole-row `data::text` |
| Webhook is the truth, never the redirect | inherited from PAY-001 | a browser landing on `/pay/success` proves nothing; entitlement moves only on a signature-verified event |
| Payment joins to a brief by **reference**, never by the brief | 2026-08-26 | `billing_orders.external_reference` carries `BR-XXXXXX`; briefs change, references do not |

## 3. Where the money actually is — the state this lane inherits

All figures `[MEASURED 2026-09-04]` unless marked. Re-measure before quoting; a count
goes stale by addition and reads as current for ever.

- **One real payment has ever been taken by the cluster**: order `36744bf0`, £30,
  reference `BR-9AUZ59`, paid 2026-08-27 14:40:22Z. **It was the owner's own.**
- **Two vouchers exist ever.** `WD-9FAB5-2NVNF` (redeemed, that order) and
  `WD-KN3WU-9PZN4` (live, £30, expires 2026-09-18, unredeemed — trial run 2).
- **One `billing_events` row.** One webhook has ever been processed for real.
- **Genuine external buyers, estate-wide, across every site: ZERO.** idea.uk's lane
  records "first organic Stripe webhook" as an open residual; the webdesign.uk side has
  taken exactly the one owner payment above. This is the single most important number in
  this document and the easiest one to lose sight of, because every surface below it works.
- The surface itself is **live and keyed**: `POST https://webdesign.uk/stripe/webhook`
  returns **400** (keyed refusal — the correct post-roll reading; 503 would mean the keys
  are gone again).
- The order-intake collector is **enabled** (900s, last fired 2026-09-04 14:32:18Z) and
  has released exactly one brief: `boxingonline.com`, seeded 2026-08-31.

## 4. The work, ranked by what closes the door

Ordered by what makes a bad state unrepresentable, not by size.

**(a) `/pay/success` and `/pay/cancel` return 404 to a customer who has just paid.**
Found 2026-08-27 by the trial run, still true 2026-09-04 (measured, both 404 on the live
domain). `stripe.go:55-56` mints them as every checkout's landing. Nothing loses money —
the webhook is the truth — but it is the worst possible moment to show someone a bare 404,
and it is **owed before ordering opens to anyone but the owner**. Must be framework-built
(the 2026-08-04 no-hand-built-HTML ruling applies; this is precisely the "however small,
however temporary" case it was written for).

**(b) The site promises a payment link the letter cannot contain.** *(Rewritten
2026-09-04 after measuring at the consumer — the original item asked whether the Payment
Links exist in Stripe, which is both unanswerable from here and the wrong question.)*

`[MEASURED 2026-09-04]` `domain_rent_url`, `domain_buy_url` and `stripe_portal_url` are all
**absent** from the live `delivery-email-sender` config, so `Links.DomainRent` is empty by
construction and the letter correctly says *"Reply to this email to arrange either."* The
machinery is honest. **The served FAQ is not**: it tells the customer the subscription link
will arrive *in their delivery email*. Fix by correcting the copy (cheap, and the letter is
the one that is right) or by building the link — owner's call, delivery lane's copy.

**(b-was) Whether the Payment Links exist in Stripe** stays `[UNMEASURED]`. `[UNMEASURED]` — the £10/month
rental subscription, the Customer Portal, and the £59.99 buy-out link are named in the
delivery letter and in the go-live checklist, and I have **not** confirmed any of them
exists in the Stripe account. My attempt to enumerate them from the auth-service pod was
refused by this session's own permission classifier, so this is an honest gap, not a
finding. **The delivery email already promises the rental link "arrives in your delivery
email".** If the link does not exist, that letter makes a promise the machinery cannot
keep — the exact defect shape the delivery lane spent this week finding by hand.

**(c) There is no recurring path at all.** `form.Set("mode", "payment")` is hardcoded in
`stripe.go:46`. The £10/month rental is a subscription by definition and cannot go through
this surface. Payment Links sidestep it for now, which is a legitimate answer — but it
should be a *decision*, recorded, not a thing that is true because nobody looked.

**(d) PAY-007 is a loaded gun.** The older `internal/auth-service/subscription/` package
still stamps `status = active` with no payment step, and `GetUsageStats` returns hardcoded
zeros so every quota check passes. Any future code that reads a subscription tier as
evidence of payment would be reading a lie. It is untouched scaffold sitting in the same
service as the real billing package, under a more obvious name. Ranked here because the
door it leaves open is opened by *a future reader being reasonable*.

**(e) No admin screen.** Vouchers and orders are admin-JWT API routes with no UI, so the
owner mints by running `trial_checkout.sh` in his own terminal. That is deliberate and
fine for trials; it does not survive the first week of real orders.

**(f) The intake gate does not create its own order.** Today the chat takes a brief and
the owner creates the order by hand. `HandleCreateOrder`'s own comment names this as the
next half. Until then, "customer pays" is not a path a customer can walk unaided.

## 5. The seam this lane cannot own alone — prices live in a letter

The domain prices (£10/month, £59.99) are **not only in this lane's surfaces**. They are
in the delivery email's `body_template`, which is a step config on an *agent definition* —
so a price census that greps the live specs will miss them by construction. Migration
`726` corrected a superseded £200 there for exactly this reason.

**This is live and contended right now.** `[MEASURED 2026-09-04]` migration `776` moved
that same `body_template` at 12:05:25Z today. Session `475` is taking the
`{{instructions_link}}` placeholder in the same jsonb path. `bugs_open/477` has migrations
`774`/`775` written and not applied, including a **second** customer letter with its own
`body_template` — so when that sender lands there will be two letters quoting this lane's
prices.

**Rule for this lane, adopted now:** a price change is this lane's decision and the
delivery lane's copy. Never edit a `body_template` from a repo copy — read the live row
first, because it moved today. Tell 475 and 477 before changing a figure.

## 6. How this lane works

- Standing five in this directory, updated as the work happens.
- Council gate before or alongside any commit touching `internal/auth-service/billing/`
  or an appliable migration — this is platform code and a live money path.
- **Anything that can take, move or refuse money gets the diagnosis loop before a durable
  root-cause claim**, not after. The cost of being confidently wrong here is measured in
  someone else's money.
- Register mechanisms in `docs026_concept_register/register/payments.md` — which is
  frozen at 2026-07-13 for everything except the entries later lanes added (PAY-009,
  PAY-010). Absence from that file is not evidence of absence.
