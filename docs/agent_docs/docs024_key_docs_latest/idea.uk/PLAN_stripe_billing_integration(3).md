# PLAN — Stripe Billing Integration

Plan for making billing actually work for the website build/host/chat product.
Companion to `PLAN_isolated_chat_environment.md` (§13 commercial model). The
verdict that motivates this plan: the auth service has a subscription **scaffold**
(tables, models, tier constants, admin CRUD, JWT carrying `client_id` + `tier`)
but **no working payment integration** — no Stripe SDK, no checkout creation, no
webhooks, mock usage stats, and a DB-dialect inconsistency suggesting parts were
never run. So payment processing, verified entitlement, working quota,
client-scoping, a one-off path, and site-shaped tiers are all genuinely new.
Self-contained; verified findings are in the Appendix.

> **idea.uk reference (sibling product, now LIVE — added 2026-06-13).** A separate, smaller product on
> the same platform — idea.uk, a £29 one-off "verified AI ideas" report — already has a *working, live*
> Stripe integration built on the principles this plan sets out, so it's a concrete reference for the
> one-off ($5 build) path here. What it confirms:
> - **Webhook is the only source of truth** (§2): entitlement/fulfilment is granted only on a
>   signature-verified `checkout.session.completed`; the success redirect just shows a page. Proven end
>   to end (request → checkout → webhook → deliver).
> - **Provider behind an interface** (§4): a `Provider` with `CreateCheckout` + `ParseWebhook`, a
>   `FakeProvider` for local/testing, no Stripe symbols past the provider package — same shape as the
>   abstraction here.
> - **One-off via Checkout `mode=payment` with inline `price_data`** and `metadata[order_id]` — the same
>   mechanism proposed for the $5 build (just `metadata[client_id]` + `intent` instead).
> - **Least-privilege key** (§9): a **restricted** key scoped to `Checkout Sessions: Write` only; secret
>   + webhook signing secret from env, never in code. Worth adopting here from day one.
> - **Refunds**: currently **manual in the Stripe dashboard** — no `/refund` endpoint and the webhook
>   ignores refund events (the order stays delivered). See §11's refund decision; idea.uk is the
>   real-world data point that manual is fine at low volume, and that a `charge.refunded` handler is the
>   next step when automation is wanted.
> - **Webhook secret must be byte-exact** (reliability, §9): in idea.uk's go-live a single stray
>   character appended to the signing secret on a paste failed *every* signature check (HTTP 400) and
>   stalled the order at `awaiting_payment` until corrected; resending the event then recovered the paid
>   order. Worth a format-check on the secret (no trailing whitespace/quotes) wherever it's configured,
>   and remember a 200 from the endpoint only means "no signature error", not "processed".
>
> idea.uk's own *operational* setup (the live webhook destination, the env file, the test→live steps)
> lives in its runbook (`RUNBOOK_idea_uk.md` → "Stripe billing — setup"), not here — this plan stays the
> chassis build/host/chat billing design.

> **Schema caveat:** this plan is written from the auth subscription *Go models*,
> not the auth DB migrations or the full auth schema. Every DDL below is
> **PROPOSED** — verify against the live auth migrations (and confirm the auth DB
> engine; the `?` vs `$1` mix must resolve to one dialect) before applying.

---

## TL;DR

- **Truth lives in one place: Stripe webhooks.** Nothing grants entitlement
  except a verified webhook event. `status = active` today means "a row exists",
  which is unsafe to gate on; that is the first thing to fix.
- **Build it behind a provider interface from day one.** Nothing calls Stripe
  yet, so there is zero retrofit cost to making billing pluggable. Stripe is
  implementation #1.
- **Two planes, one bridge.** The auth service owns billing truth (subscriptions,
  credits, webhooks). The chassis *gates* on a **cached `client_entitlements`
  table**, fed one-directionally from auth. Build-submission gate reuses the
  existing `approval_mode` hold; maintenance gate is a join-filter on the
  heartbeat selection queries.
- **Two charge shapes:** recurring (maintenance/tier subscription, per `client`)
  and one-off (the **$5 build**, a credit — not a subscription). The scaffold
  only models recurring; the one-off credit path is new.
- **Reuse:** the subscription scaffold, tier model, JWT propagation, the
  `clients→networks→sites` ownership joined to auth by `client_id`, the
  `approval_mode` hold, and the atomic-claim idiom for credit consumption.

---

## 1. Starting point — what exists vs what's missing (verified)

Reusable scaffold (keep):
- `subscriptions` table/model (`user_id`, `tier`, `status`, dates,
  `payment_method`, `stripe_customer_id`, `stripe_subscription_id`),
  `subscription_tiers` table, status/tier constants, admin CRUD handlers.
- JWT carries `client_id` + `tier`; `client_id` joins auth to the chassis
  (`networks.client_id → clients.id`, `clients.external_id` back-reference).

Missing / broken (build):
- No Stripe SDK; `CreateSubscription` is a bare insert stamping `status=active`
  with no payment. No checkout creation. **No webhook handler.**
- `GetUsageStats` returns mock zeros → `CheckQuota` always passes.
- `repository.go` mixes `?` and `$1` placeholders → never exercised against one
  DB. `Create` doesn't persist the `stripe_*` columns it selects.
- Subscription is per-`user`; the product needs per-`client`. Tiers are
  personas/projects/content; the product needs sites/builds/maintenance/chat.

---

## 2. Design principles

1. **Webhook is the only source of truth.** Client-side success redirects never
   grant entitlement (they are spoofable and unreliable). Only a
   signature-verified webhook mutates entitlement state.
2. **Provider behind an interface.** No Stripe symbol leaks past the provider
   package. `provider` + `provider_customer_id` + `provider_subscription_id`
   generalise the `stripe_*` columns.
3. **Reuse, don't rebuild.** Extend the subscription scaffold; reuse
   `approval_mode` for the build gate; reuse `clients`/`networks` for ownership;
   reuse the atomic `UPDATE ... WHERE ... RETURNING` idiom (as in
   `claim_work_item`) for credit consumption.
4. **Credentials parameterised.** All Stripe keys + webhook secret from
   config/secrets, never hardcoded — consistent with the sell-the-unit
   requirement.
5. **Idempotent and reconcilable.** Every webhook is dedup'd by event id; a
   periodic reconciliation sweep catches missed events.

---

## 3. Architecture — where truth lives, where gates check

```
                 Stripe
                   │  (signature-verified webhooks)
                   ▼
        Auth service  ── owns billing TRUTH ──┐
          subscriptions / credits / events    │ emits entitlement-changed event
          (writes only via webhooks)          ▼
                                   chassis: client_entitlements (CACHE)
                                     read by BOTH gates ▲
                                                        │
     build-submission gate ───────────────────────────┤  (per build, low volume)
       approval_mode 'pending_entitlement' hold        │
     maintenance-run gate ─────────────────────────────┘  (heartbeat join, high volume)
       filter selection queries by maintenance_active
```

- **Truth = auth DB**, mutated only by webhooks. Source of record for
  subscriptions and one-off credits.
- **Gate reads = chassis `client_entitlements` cache**, fed one-directionally
  from auth (entitlement-changed event on Kafka, applied by a small consumer;
  reconciliation sweep as backstop). The cache is a local, SQL-joinable table —
  required because the maintenance heartbeat filters thousands of sites per tick
  and cannot make a per-site call to auth.
- **JWT** carries `client_id` for identity and a coarse `tier` hint, but gates
  trust the **cache** (verified), not the token's tier.
- **Isolation fit:** this is the same "core publishes, satellite caches" pattern
  from the isolation plan — a chat/SaaS satellite holds its own
  `client_entitlements` cache and never calls back into auth synchronously.

Phasing note: the **build-submission** gate is low-volume (one check per build)
and *may* start by calling an auth entitlement endpoint synchronously; the
**maintenance** gate needs the cache from the start. Converge both on the cache.

---

## 4. Provider abstraction (Stripe first)

A `billing` provider interface; Stripe is one implementation. Sketch, not final:

```go
type Provider interface {
    // Map a client to a provider customer (idempotent).
    EnsureCustomer(ctx, clientID, email string) (providerCustomerID string, err error)

    // Recurring: maintenance / tier subscription.
    CreateSubscriptionCheckout(ctx, in SubscriptionCheckoutInput) (CheckoutSession, error)

    // One-off: the $5 build. metadata carries client_id + intent (e.g. site_id/spec ref).
    CreateOneOffCheckout(ctx, in OneOffCheckoutInput) (CheckoutSession, error)

    // Self-service management/cancel.
    CreatePortalSession(ctx, providerCustomerID, returnURL string) (url string, err error)

    CancelSubscription(ctx, providerSubscriptionID string) error

    // Verify signature and normalise to our event shape.
    ParseWebhook(payload []byte, signature string) (Event, error)
}

// Normalised events — provider-agnostic, what our handlers switch on.
type Event struct {
    Type     EventType // OneOffPaid | SubscriptionActivated | SubscriptionUpdated |
                        // SubscriptionCanceled | InvoicePaid | PaymentFailed
    ClientID string
    Tier     string
    Status   string
    PeriodEnd *time.Time
    ProviderCustomerID, ProviderSubscriptionID string
    Metadata map[string]string // e.g. {"intent":"site_build","site_id":...}
    ProviderEventID string      // for idempotency
}
```

Keep the `CheckoutSession` type already in the models; drop the Stripe-specific
example URLs from its doc. The handlers/gates speak only `Event` and `Provider`.

---

## 5. Data model changes (PROPOSED — schema-check first)

Auth DB (extend the scaffold):

- **Generalise provider columns** on `subscriptions`: add `provider` (default
  `'stripe'`), rename/alias `stripe_customer_id`/`stripe_subscription_id` to
  `provider_customer_id`/`provider_subscription_id` (or add the generic columns
  and backfill). Add `current_period_end timestamptz` (needed to expire
  entitlement) and `client_id uuid` (scope to client, not just user — see §7).
- **Fix the dialect**: all queries to one engine's placeholders (`$N` if
  Postgres). `Create` must persist the provider columns it reads.
- **Real usage** to replace mock `GetUsageStats`: count sites/builds per client
  (cross-plane — see §7), or drop `CheckQuota`'s resource model in favour of the
  entitlement model below.

New tables (auth DB):

```sql
-- PROPOSED — verify auth schema/engine before applying

-- One-off credits ledger (the $5 build, first-site-free grant).
CREATE TABLE billing_credits (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id     uuid NOT NULL,
    kind          text NOT NULL,              -- 'site_build'
    granted       integer NOT NULL,           -- +1 on purchase/free grant
    consumed      integer NOT NULL DEFAULT 0, -- +1 when a build is released
    source        text NOT NULL,              -- 'stripe_oneoff' | 'free_tier' | 'manual'
    provider_ref  text,                       -- checkout/session/payment id
    created_at    timestamptz NOT NULL DEFAULT now()
);
-- balance = sum(granted) - sum(consumed) per (client_id, kind)

-- Webhook idempotency + audit.
CREATE TABLE billing_events (
    provider          text NOT NULL,
    provider_event_id text NOT NULL,
    type              text NOT NULL,
    payload           jsonb,
    processed_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, provider_event_id)   -- dedup
);
```

Chassis DB (the gate cache):

```sql
-- PROPOSED — verify chassis schema before applying
CREATE TABLE client_entitlements (
    client_id          uuid PRIMARY KEY REFERENCES clients(id) ON DELETE CASCADE,
    tier               text,
    subscription_status text,                  -- verified: active|trialing|past_due|canceled|expired
    maintenance_active boolean NOT NULL DEFAULT false,
    build_credits      integer NOT NULL DEFAULT 0,  -- one-off builds remaining
    builds_per_period  integer,                -- tier allowance, NULL = none, -1 = unlimited
    builds_used_period integer NOT NULL DEFAULT 0,
    valid_until        timestamptz,            -- current_period_end mirror
    updated_at         timestamptz NOT NULL DEFAULT now()
);
```

Tier reframe (auth `subscription_tiers`): replace persona/project/content limits
with `max_sites`, `builds_per_period`, `maintenance_included bool`,
`chat_included bool`. New rows, same table shape (minus the wrong columns).

---

## 6. Stripe integration pieces

- **Customer**: `EnsureCustomer` on first checkout; store `provider_customer_id`
  on the subscription/client. Map `clients.external_id` ↔ provider customer.
- **Checkout — subscription mode**: for maintenance/tier. Returns a redirect URL.
- **Checkout — payment mode**: for the **$5 build**. `metadata` carries
  `client_id` + `intent=site_build` (+ optional `site_id`/spec ref) so the
  webhook can grant the right credit.
- **Webhooks** (the core): one signed endpoint. Verify signature, dedup on
  `provider_event_id`, then apply:

| Stripe event | Normalised | Effect (auth truth) |
|---|---|---|
| `checkout.session.completed` (mode=payment) | `OneOffPaid` | `billing_credits` +1 `site_build` for client; emit entitlement-changed |
| `checkout.session.completed` (mode=subscription) | — | link `provider_subscription_id`; await subscription events |
| `customer.subscription.created`/`updated` | `SubscriptionActivated/Updated` | upsert `status`, `tier`, `current_period_end`; emit |
| `customer.subscription.deleted` | `SubscriptionCanceled` | `status=canceled`; emit |
| `invoice.paid` | `InvoicePaid` | ensure `status=active`, extend `period_end`; emit |
| `invoice.payment_failed` | `PaymentFailed` | `status=past_due`; emit |

  Every effect that changes entitlement **emits an entitlement-changed event**
  that the chassis consumer applies to `client_entitlements`.
- **Customer portal** (optional, cheap): `CreatePortalSession` for self-service
  cancel/update — offloads subscription management to Stripe.
- **Reconciliation sweep** (backstop): periodic job pulls subscription status
  from the provider for clients whose `valid_until` is near/past, to catch missed
  webhooks. Belt-and-braces for the "truth" guarantee.

---

## 7. Product flows end-to-end

**$5 one-off build (next-day).**
1. Chat collects domain + spec (briefing). Before the expensive build, the
   intake creates a **payment-mode checkout** (`intent=site_build`, `client_id`).
2. The first expensive work item is held: `approval_mode='pending_entitlement'`
   (the build does not spend).
3. User pays → `checkout.session.completed` webhook → `billing_credits` +1 →
   entitlement-changed → chassis cache `build_credits` +1.
4. Gate sees a credit, **atomically consumes it** (`UPDATE client_entitlements
   SET build_credits = build_credits - 1 WHERE build_credits > 0 RETURNING …`),
   releases the held item → build proceeds.
5. Build runs on the cheap/batched tier (`saas_cheap`), delivered next-day.

**First-site-free.** On client creation (or first build request), grant one
`free_tier` credit (no checkout). Flow above from step 4. Subsequent builds need
payment. Identity (email-verified user/client) is required so the free grant is
per-real-client, not per-request — free is a cost/abuse magnet.

**Maintenance subscription (recurring).** Subscription checkout → webhooks set
`status=active` + `period_end` → cache `maintenance_active=true`,
`valid_until=period_end`. The **maintenance gate** (heartbeat selection queries)
filters to sites whose client has `maintenance_active=true AND valid_until > now()`.
On `payment_failed`/`deleted`, cache flips false → maintenance work stops being
selected for those sites. (This is the cost-control valve at scale — it pays off
in the pure-operator case before anything is sold.)

**Per-domain sell-on.** Re-parent the site's `network_id` to the buyer's
network/client; the new client's entitlement governs maintenance going forward;
cancel/transfer the seller's subscription as appropriate. Credentials swap per
the separability plan.

---

## 8. The two gates wired

- **Build-submission** — reuse `approval_mode`. Add a `pending_entitlement` hold
  state (mirrors `hitl`'s `pending_review`): the first expensive item parks until
  the entitlement check passes (a credit consumed, or an active tier with
  `builds_per_period` allowance remaining). On pass, release to the normal flow;
  the dispatch loop is unchanged. Consumption is atomic (above) to avoid double
  spend on retries/races.
- **Maintenance-run** — filter the heartbeat selection queries
  (`build-pipeline-trigger`, `improvement-loop`, `content-feed-trigger`) with a
  join to `client_entitlements` requiring `maintenance_active`. Single insertion
  point per trigger; filter at selection, not per item.

---

## 9. Reliability & security

- **Signature verification** on the webhook endpoint; reject unsigned/invalid.
- **Idempotency** via `billing_events` PK on `(provider, provider_event_id)`.
- **No trust in redirects** — success/cancel URLs are UX only; entitlement
  changes only on webhooks.
- **Out-of-order / missed events** — handlers are upserts keyed on current state;
  the reconciliation sweep repairs drift.
- **Secrets** — Stripe secret key + webhook signing secret from secret store, not
  code; per-instance for sold units.
- **Test mode** — Stripe test keys + CLI webhook forwarding + test clocks for
  subscription lifecycle in CI/staging before any live key.
- **Abuse** — moderate the brief and rate-limit intake (anonymous "type a domain,
  get a site" is a spam/cost magnet); require verified identity for the free
  grant.

---

## 10. Build order (structural first)

1. **Provider interface + normalised `Event`** — the contract everything depends
   on. No Stripe symbol outside the provider package.
2. **Schema** — generalise provider columns + dialect fix on `subscriptions`;
   add `billing_credits`, `billing_events` (auth); add `client_entitlements`
   (chassis). All after schema-check.
3. **Webhook endpoint** — signature verify + dedup + the event→effect table
   (§6). This establishes verified truth; nothing else is safe until it exists.
4. **Entitlement-changed event → chassis cache consumer** + reconciliation sweep.
   Now gates have something trustworthy to read.
5. **Maintenance gate** — join-filter the three heartbeat queries on the cache.
   (Earns its keep immediately as a cost valve, pre-sales.)
6. **One-off credit + checkout (payment mode)** and the **`pending_entitlement`**
   hold on build-submission. This lights up the $5 product.
7. **First-site-free** grant + identity requirement.
8. **Subscription checkout + portal** for recurring maintenance.
9. **Tier reframe** to site/build/maintenance/chat nouns.

Steps 1–4 are the spine: provider seam, schema, verified webhooks, cached
entitlement. Everything user-facing hangs off them.

---

## 11. Open decisions

- **Auth DB engine** (Postgres vs MySQL) — the `?`/`$1` mix must resolve; it
  determines the dialect for all the above.
- **Entitlement scope**: per-`client` (recommended; sites are per-client) vs
  keeping per-`user`. If multiple users per client, per-client is needed; decide
  how a user's action resolves to the client's entitlement (JWT `client_id`).
- **Bridge transport**: Kafka entitlement-changed event (idiomatic, isolation-
  friendly) vs a synchronous auth endpoint for the build gate only. Maintenance
  gate needs the cache regardless.
- **Where the $5 checkout is created**: chat/intake on the satellite calling
  auth, vs auth-hosted checkout. Affects the isolation boundary.
- **Tier vs pure metered**: fixed tiers (free/basic/…) vs per-build metered
  pricing. The credit model supports both; the tier table is the fixed-plan side.
- **Refund/chargeback handling**: revoke credits / suspend maintenance on
  `charge.refunded`/dispute — design later, note the hook. (idea.uk currently does
  refunds **manually in the dashboard** with no `charge.refunded` handler — fine at
  low volume; that handler is the first automation step when wanted.)

---

## Appendix — verified findings (for self-containment)

From reading `internal/auth-service/subscription/{models,repository,service,
handlers}.go`:

- `service.go` imports only `uuid`, `zap`, `time`, `context`, `fmt` — **no Stripe
  SDK**. `CreateSubscription` inserts a row with `status=active`, no payment.
  `Cancel`/`Update` are bare DB writes. **No webhook handler exists.**
- `GetUsageStats` returns hardcoded zeros ("returning mock data"), so
  `CheckQuota` (personas/projects/content) always passes.
- `repository.go`: `GetByUserID`/`Create`/`Update`/`Cancel`/`GetTier` use `?`
  placeholders; `ListAll` uses `$1` — cannot all run on one engine. `Create`
  omits the `stripe_*` columns it `SELECT`s elsewhere.
- `models.go`: `Subscription` keyed on `user_id`; `stripe_customer_id`/
  `stripe_subscription_id` present; tiers `free/basic/premium/enterprise` with
  persona/project/content limits; `CheckoutSession` type present but never
  constructed; status `active/trialing/past_due/canceled/expired`.
- `handlers.go`: get/usage/quota (user-scoped) + admin create/update/list. No
  checkout or webhook endpoints.
- JWT `Claims`: `user_id`, `email`, `client_id`, `role`, `tier`, `permissions`.
- Chassis ownership: `clients → networks → sites`; join key `client_id`;
  `clients.external_id` back-reference.
