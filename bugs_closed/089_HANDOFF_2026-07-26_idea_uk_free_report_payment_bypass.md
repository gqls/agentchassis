# 089 — idea.uk: a buyer can take the £29 report without paying

**Filed** 2026-07-26, from the idea.uk VM-site workstream.
**Class** authorisation / test-affordance reachable in production.
**Status** **CLOSED 2026-07-26 — fixed, deployed and proven live against a real
order.** Binary deployed 18:29 UTC (rollback kept at `/opt/idea/idea.prev-2026-07-25`).

**The live proof, on the deployed binary, against a genuine order in the exact
vulnerable state** (`ord_1785090638951163875`, `awaiting_payment`, a £29 Stripe
`cs_live_` session outstanding and its report already stored — i.e. everything
the attack needs to pay off):

```
$ curl -s 'https://idea.uk/order/success?o=ord_1785090638951163875&fake=1'   # from the public internet
HTTP 200
order status after the attack: awaiting_payment      # pre-fix this read: delivered
Jul 26 18:45:44 idea1 idea[109722]: orderSuccess: refused fake=1 payment shortcut
  under *main.StripeProvider (order "ord_1785090638951163875", ip 2a02:c7c:...)
```

Not a seeded fixture and not a test double: the real funnel, mid-purchase, hit
the way a buyer could hit it.
**Scope** the idea.uk tool only (`docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/`),
a standalone stdlib-only Go module. Not chassis code, not fleet-wide.

## Symptom

`GET https://idea.uk/order/success?o=<order_id>&fake=1` moves an order from
`awaiting_payment` straight to `paid`, and then delivers the report — with no
money having changed hands and no Stripe involvement at all.

Verified against the code path deployed on the box (binary of 2026-07-25 15:11).

## Root cause

`service.go:orderSuccess` honoured a local-test shortcut on the strength of a
**query parameter alone**:

```go
func (a *App) orderSuccess(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("fake") != "" { // FakeProvider local-test shortcut
		id := r.URL.Query().Get("o")
		if o, ok := a.store.Get(id); ok && o.Status == "awaiting_payment" {
			a.store.Update(id, func(o *Order) { o.Status = "paid" })
			...
```

The shortcut exists for `FakeProvider`, whose `CreateCheckout` returns exactly
that URL (`billing.go:134`) so a local run can complete a purchase with no
Stripe keys. `billing.go:129` even labels the type *"local/testing only — NEVER
in production"*. But the handler never asked which provider was configured, so
the affordance was equally available under `StripeProvider`.

Confirmed live: `curl -s http://127.0.0.1:8080/health` on the box →
`{"auto_deliver":false,"ok":true,"price_gbp":29,"provider":"*main.StripeProvider"}`.

## Why it is reachable — the part that makes it a real defect, not a theoretical one

The order id is not a secret from the buyer. `StripeProvider.CreateCheckout`
(`billing.go:53-54`) hands Stripe **both** URLs with the id embedded:

```go
form.Set("success_url", fmt.Sprintf("%s/order/success?o=%s", p.publicBaseURL, orderID))
form.Set("cancel_url",  fmt.Sprintf("%s/order/cancel?o=%s",  p.publicBaseURL, orderID))
```

So a buyer who opens the pay link and **cancels** is redirected to a URL
containing their own order id. Appending `&fake=1` to the sibling success path
then delivers the report. No guessing, no interception — one visible URL and one
query parameter.

Blast radius is bounded by the funnel: it needs an order the operator has
already approved (`awaiting_payment`). It is not mass-exploitable — ids are
`ord_<UnixNano>`, so enumeration is not practical — but every genuine buyer is
one URL edit away from not paying. Nothing has been lost to it: of the 72 orders
on the box, 3 are `delivered` and none reached `delivered` by this path.

## The fix (committed, awaiting deploy)

Gate the shortcut on the provider actually being `FakeProvider`, and log a
refusal when it is attempted under a real one:

```go
_, providerIsFake := a.provider.(*FakeProvider)
if r.URL.Query().Get("fake") != "" && !providerIsFake {
	log.Printf("orderSuccess: refused fake=1 payment shortcut under %T (order %q, ip %s)",
		a.provider, r.URL.Query().Get("o"), clientIP(r))
}
if providerIsFake && r.URL.Query().Get("fake") != "" {
	...
```

The page itself still renders 200 for everyone — Stripe redirects real buyers
here, so it must.

`payment_bypass_test.go`, three cases:

1. **`TestFakePaymentShortcutRefusedUnderStripe`** — seeds an `awaiting_payment`
   order with a stored report, hits the bypass URL, asserts the status is still
   `awaiting_payment` and that nothing was emailed to the buyer.
2. **`TestFakePaymentShortcutStillWorksUnderFakeProvider`** — the legitimate
   local path still delivers, so the shortcut keeps its only purpose.
3. **`TestOrderSuccessPageStillRendersUnderStripe`** — the refused case is a
   no-op, not an error: still 200.

**The failing branch was induced, not assumed.** Test 1 was run against the
pre-fix `service.go` (`git show HEAD:…/service.go` into a scratch copy of the
module) and fails with:

```
payment_bypass_test.go:75: PAYMENT BYPASS: status moved to "delivered" under
StripeProvider; want it held at awaiting_payment
```

`delivered`, not merely `paid` — under the box's `ReviewBeforePay` config the
stored report is emailed to the buyer in the same request. Post-fix all three
pass; `go vet` clean; full suite green; `GOOS=linux GOARCH=amd64 go build` ok.

## How to verify once deployed

On the box, with a real order sitting in `awaiting_payment` (or seed one while
the service is **stopped** — see the trap below):

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  'http://127.0.0.1:8080/order/success?o=<id>&fake=1'   # expect 200
python3 -c "import json;d=json.load(open('/var/lib/idea/orders.json'));print(d['orders']['<id>']['status'])"
# expect: awaiting_payment  (pre-fix this printed: delivered)
journalctl -u idea --since '-5 min' | grep 'refused fake=1'   # expect the refusal line
```

**Trap:** never edit `/var/lib/idea/orders.json` under a running service — it is
read once at startup and rewritten wholesale from memory on every order change.
`systemctl stop idea` → edit → `systemctl start idea`.

## The transferable pattern (also added to 016b §9)

**A test affordance that is selected by request data rather than by build or
configuration is a production authorisation hole.** The guard belongs on the
thing that cannot be supplied by the caller — here, the type of the configured
provider. A comment saying "NEVER in production" documents the intent; it does
not implement it. Grep any codebase for shortcut branches keyed on a query
parameter, header, or body field whose only purpose is to skip a real-world
step: payment, auth, rate limiting, email sending.

## Related

- `bugs_open/006` C — idea.uk infra errors (same box, unrelated cause).
- Workstream docs: `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/`
  (`HANDOFF_RESUME_idea_uk_vm_site.md` is the entry point; `RUNNING_NOTES` §X.19
  is the log entry for this find).
