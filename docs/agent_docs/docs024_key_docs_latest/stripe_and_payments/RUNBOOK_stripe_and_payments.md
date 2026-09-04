# RUNBOOK — stripe and payments

Every command that was hard to get right, with its gotcha attached. Change it HERE.

## 0. Standing rules for this lane

- **Never read a Stripe key value into a session.** Read key *names* from the secret;
  probe from inside the pod so the value stays where it belongs. (MEMORY:
  `never-extract-keys-probe-from-the-pod`.)
- **The webhook is the truth.** A browser landing on a success page proves nothing about
  payment. Never gate anything on a redirect.
- **The amount is computed server-side, always.** Never accept a price from a client.
- **Read the live row before editing any price that lives in config.** Prices live in an
  agent definition's `body_template` as well as in this lane's tables (§7).

## 1. Is the payment surface actually live?

The single most useful probe. **400 = keyed and working. 503 = the keys are gone.**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X POST https://webdesign.uk/stripe/webhook \
  -H 'Content-Type: application/json' -d '{}'          # expect 400
curl -s -o /dev/null -w "%{http_code}\n" https://webdesign.uk/   # control: expect 200
```

⚠ **Always run the apex control.** A 000/timeout on the webhook alone cannot distinguish
"keys gone" from "domain down" — and the domain has been parked before.

⚠ **503 after a release is the known failure**: `047-base-configs` re-applies an
authoritative `data` map and REMOVES hand-patched keys. Both Stripe keys are in that map
as REQUIRED variables (fixed 2026-08-26), so a release missing them now fails loudly by
name instead of reverting silently. Verify at the **consumer** (this 400), never at the
secret listing.

```bash
grep -n 'STRIPE' deployments/terraform/environments/production/uk001/047-base-configs/main.tf
kubectl -n ai-persona-system get secret personae-platform-secrets \
  -o jsonpath='{.data}' | tr ',' '\n' | grep -oE '"[A-Z_]+"'   # NAMES only, never values
```

## 2. The state of the money, in one block

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db \
  -c "SELECT code, drops_price_to_pence, recipient_name, expires_at, redeemed_at FROM vouchers ORDER BY created_at DESC;" \
  -c "SELECT id, amount_pence, status, external_reference, paid_at FROM billing_orders ORDER BY created_at DESC;" \
  -c "SELECT * FROM billing_settings;" \
  -c "SELECT count(*) FROM billing_events;"
```

⚠ Column names that will bite you: it is `drops_price_to_pence`, **not** `amount_pence`,
on `vouchers` (the voucher does not have a price, it drops one). `scheduled_tasks` uses
`name`, **not** `task_name`, and has `interval_seconds`, not `schedule`.

## 3. Verify a voucher will actually redeem — before the owner spends a trial on it

Needed whenever a voucher was minted **by SQL** rather than through the admin API, because
then `Service.CreateVoucher`'s guards were honoured rather than applied. The redemption
predicate is `repository.go:123`.

**Run it and ROLL BACK. Always include both controls** — a bare "1 row" proves the
statement ran, not that it discriminates.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
BEGIN;
UPDATE vouchers SET redeemed_at = now(), redeemed_by_client_id = '<a real client uuid>'
 WHERE code = '<CODE>' AND redeemed_at IS NULL AND expires_at > now()
 RETURNING id, drops_price_to_pence;                      -- expect 1 row
UPDATE vouchers SET redeemed_at = now(), redeemed_by_client_id = '<same uuid>'
 WHERE code = 'WD-XXXXX-XXXXX' AND redeemed_at IS NULL AND expires_at > now()
 RETURNING id;                                            -- CONTROL: expect 0
UPDATE vouchers SET redeemed_at = now(), redeemed_by_client_id = '<same uuid>'
 WHERE code = '<an already-redeemed code>' AND redeemed_at IS NULL AND expires_at > now()
 RETURNING id;                                            -- CONTROL: expect 0
ROLLBACK;
SELECT code, redeemed_at FROM vouchers WHERE code = '<CODE>';   -- must still be NULL
SQL
```

## 4. Mint a voucher and create an order (the sanctioned path)

`docs024_key_docs_latest/webdesign_uk_build_service/trial_checkout.sh` — **the owner runs
this himself**, in his own terminal. Both routes are admin-JWT-gated by design and the JWT
is his; the password never leaves his process.

```
./trial_checkout.sh <client_id> <BR-reference> [1000|3000|5500] [recipient] [email]
```

It port-forwards auth-service, logs him in, mints a 14-day single-use voucher at a **ruled**
amount (the API refuses anything else), creates the order carrying the BR- reference and
the voucher, and prints the Stripe checkout URL.

⚠ The amounts are **1000 / 3000 / 5500 only**. £30 exists specifically so the owner's
end-to-end trials exercise the real voucher-and-payment path. A £0 voucher is refused, and
would be wrong anyway — it would skip the path the trial exists to test.

## 5. Read what a real payment actually contained

The stored webhook payload is the only record of what the customer saw. It answers
questions no code read can — branding, customer creation, the landing URLs.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -c \
"SELECT jsonb_pretty(jsonb_build_object(
   'mode',              payload->'data'->'object'->'mode',
   'customer',          payload->'data'->'object'->'customer',
   'customer_creation', payload->'data'->'object'->'customer_creation',
   'livemode',          payload->'livemode',
   'metadata',          payload->'data'->'object'->'metadata',
   'success_url',       payload->'data'->'object'->'success_url',
   'branding_settings', payload->'data'->'object'->'branding_settings')) FROM billing_events;"
```

⚠ **`customer` is null on a one-off `mode=payment` charge** with
`customer_creation: "if_required"` — Stripe creates no Customer object, so
`clients.external_id` and `billing_orders.provider_customer_id` stay empty. That is
correct behaviour, **not** a broken webhook. It changes the day anything moves to
`mode=subscription` or `customer_creation: "always"`.

## 6. Order intake — is the paid gate running?

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT name, enabled, interval_seconds, last_triggered_at FROM scheduled_tasks WHERE name='order-intake-collect';" \
  -c "SELECT domain, status, created_at FROM build_queue ORDER BY created_at DESC LIMIT 5;"
```

A brief is released into `build_queue` only when a **paid** `billing_orders` row carries
its `BR-` reference in `external_reference`. The reference is the join, never the brief
itself (owner ruling 2026-08-26 — briefs change, references do not).

## 7. ⚠ Prices are NOT only in this lane's tables

The £10/month rental and £59.99 buy-out are quoted in the delivery email's
`body_template`, which is a step config on an **agent definition**. A price census that
greps the live *specs* will miss them by construction — migration `726` had to correct a
superseded £200 there for exactly this reason.

**Before changing any published price: read the live row, not a repo copy.** As of
2026-09-04 that row is contended — migration `776` moved it at 12:05:25Z today, session
`475` is editing the same jsonb path, and `bugs_open/477` has an unapplied second letter
with its own `body_template`. Tell both before touching a figure.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -c \
"SELECT jsonb_path_query(default_config, '\$.workflow.steps[*].config.body_template')
 FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

## 8. Things that are NOT in the repository and will never fail a test

Keep this list; it is the lane's blind spot by construction.

- **The Stripe account's branding** (`display_name`, logo, colours). As of 2026-09-04 the
  live account's display name is **"Fine Tune"** — wrong for webdesign.uk customers.
- **Payment Links and the Customer Portal** — dashboard objects. `[UNMEASURED]` whether
  the rental/buy-out links exist at all.
- **Restricted-key scopes.** Guidance given 2026-08-26: Checkout Sessions = **Write** is
  the only one required; Charges/Refunds = **Read** (dashboard refunds use no API key; the
  webhook uses the signing secret). Widen precisely on a named permissions error, never
  pre-emptively.
- **Webhook endpoint registration** (URL + subscribed events). Stripe does not validate
  reachability at creation time, so an endpoint can be registered long before it works.
