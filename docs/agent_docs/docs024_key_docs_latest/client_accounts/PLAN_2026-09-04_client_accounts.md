# PLAN — 2026-09-04 — client accounts

Lane opened 2026-09-04 at the owner's direction: *"find out everything we've planned and
discussed about setting up client accounts. This lane can have the responsibility for it."*

This first file is the **survey**, not yet a design. It records what is already built, what
is already ruled, what is designed-and-unbuilt, and the questions nobody has answered — each
with its source path, because several lanes have talked about this from different angles and
none of them owns it.

> **⚠ SCOPE AMBIGUITY, stated up front and not yet resolved by the owner.** "Client accounts"
> has two readings in this estate and both were live in the last 24 hours:
>
> **(A) An account WITH US** — a customer logs in to something we run and sees their site,
> their domain, their hosting and their billing. Designed as Phases 5–6 of the delivery
> architecture; nothing is built.
>
> **(B) An account with a THIRD-PARTY HOST**, set up for the customer as part of handover —
> the Netlify question the owner raised on 2026-09-04 after performing the handoff himself.
> Costed in
> `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/OPTIONS_2026-09-04_what_it_would_take_to_set_up_hosting_accounts_for_customers.md`.
>
> They are one subject — *what account does a customer have, with whom, and what can they do
> with it* — but they are different builds with different owners. §7 asks the owner to pick.

---

## 1. The canonical shape, already ruled — and the fact that it holds no data

**Owner ruling 2026-08-10** (`ai_site_selling_automation/PLAN_2026-08-10_ai_site_selling_automation.md`
§1.1): customer identity lives on the existing **`clients → networks → sites`** FK chain.
Customer-shaped columns were ADDED to `clients`; `sites` gains **no** ownership columns; the
`owner_id`-on-sites shape and the abandoned `site_ownership` junction (register **ADM-008**)
both lose. Stripe linkage rides the existing `clients.external_id`.

Shipped as migration `docs/agent_docs/sql_for_agents/375_clients_customer_identity_columns.sql`
— `clients` gained `email`, `phone`, `tier`, `customer_status`, `notes`.

**`[MEASURED 2026-09-04]` the chain is structurally present and carries no ownership at all:**

| query | result |
|---|---|
| `SELECT count(*) FROM clients` | **3** — `Default Client`, `System Scheduler`, and one real customer row `Boxing Online` (created 2026-08-27, email `aaa@designconsultancy.co.uk`, i.e. the owner's own) |
| `SELECT count(*) FROM networks` | **1** — `00000000-…-0002`, owned by `Default Client` |
| `SELECT count(*), count(network_id) FROM sites` | **60 sites, 42 with a network** |
| boxingonline.com's `network_id` | `00000000-…-0002` — the **default** network, not a Boxing Online one |

So every framework-built site hangs off one shared network, and the `Boxing Online` client row
is **not joined to boxingonline.com by anything**. The chain the owner ruled canonical is a
schema, not a record.

**Why:** `createDefaultNetwork` (`platform/orchestration/actions/site_db_actions.go:1085`) is
the only writer of `networks` in the tree, and it creates exactly one row, `slug='default'`.
Nothing anywhere creates a network per customer. `[MEASURED 2026-09-04]`
`grep -rn "INSERT INTO networks" --include=*.go .` → one hit, that function.

**This is the single largest gap and it is cheap to state:** whatever else "client accounts"
means, a customer→site link has to exist, and today it does not.

## 2. What a customer's identity actually IS today — four stores, no join

`[MEASURED 2026-09-04]` a paying customer's details are spread across four places, and no
column joins them:

1. **`build_queue.direction.customer_email` / `customer_name` / `order_reference`** — written by
   `collect_external_orders` (`platform/orchestration/actions/collect_external_orders_action.go`),
   the durable home of the *payer's* address by `bugs_open/420`'s contract split.
2. **`billing_orders.client_id` + `external_reference`** (`internal/auth-service/billing/models.go`,
   migrations `391`, `659`) — the payment side. `[MEASURED 2026-09-04]` 1 paid order, 1 voucher.
3. **`sites.email` / `sites.company_name`** — the **PUBLISHED** contact only, since `bugs_open/420`.
   Written by `seedCustomerIdentity` (`platform/orchestration/actions/seed_build_queue_action.go:317`)
   and only from `direction.published_contact`, never from the payer's address.
4. **`site_deliveries.delivered_to`** — who a delivered site was actually sent to. ~~**NEW and in
   flight in another session's tree**~~ **CORRECTED 2026-09-04, same day: COMMITTED (`698b144fa`) and
   APPLIED.** `[MEASURED 2026-09-04]` the table exists live and carries one backfilled row
   (`recorded_by = 'backfill-orchestration-states-778'`, written 14:16:57Z). Its Go half rides the
   v1.0.1361 cut and had not rolled at the time of writing. Caught by the fleet roll notice, hours
   after writing — **a peer lane's in-flight work is the fastest-staling fact you can put in a
   document** (`docs/agent_docs/sql_for_agents/778_…sql`, written by `platform/delivery.Claim`). Exists precisely because `sites.email` is populated and wrong for
   this purpose: idea.uk carries `idea.uk@contactforsales.com` while the delivery went to
   `aaa@designconsultancy.co.uk`.

`clients` — the store the owner ruled canonical — is written by **none** of these. Its only
writer is the admin endpoint `POST /admin/customers`
(`internal/core-manager/admin/customer_handlers.go:190`).

## 3. What IS built (customer-facing), verified

- **`customer_access_tokens`** (`docs/agent_docs/sql_for_agents/511_handover_state_and_customer_tokens.sql`)
  — one hashed, expiring, optionally single-use token per customer-facing link. sha256 hex;
  the plaintext lives only in the email. `purpose` is a **CLOSED** vocabulary enforced by a
  CHECK, so a new kind of customer link costs a migration on purpose.
  `[MEASURED 2026-09-04]` live constraint = `CHECK (purpose = ANY (ARRAY['zip_download','confirm_transfer']))`;
  live rows = 1 of each, both minted by the 2026-09-03 idea.uk delivery rehearsal.
  **`editor_session` is the reserved third purpose and does not exist yet.**
- **`sites.handed_over_at` / `live_link_expires_at` / `transfer_confirmed_at`** (same migration).
  `[MEASURED 2026-09-04]` **1 of 60** sites has ever been handed over (idea.uk, 2026-09-03).
- **The links host** — `links.webdesign.uk`, `/c/<token>` (confirm transfer) and `/d/<token>`
  (zip download), on the box, outside the cluster's auth middleware.
- **`LiveLinkWindow = 6*7*24h` = 42 days** (`platform/delivery/handover.go:104`), with
  `AdvertisedLiveWindowDays` = 30 as a *named deliberate margin* (`handover.go:106`,
  `prepare.go:218`). The 30-day cut is a policy, not a limit — which is what makes §7's paid
  hosting question an ordinary value change rather than a build.
- **The billing surface** — register **PAY-009**: vouchers, one-off `billing_orders`,
  webhook-as-truth, `£149` list price. Deployed **keyless**: no `STRIPE_SECRET_KEY` yet, so the
  webhook 503s and nothing can actually be charged.

## 4. What is DESIGNED and unbuilt — Phases 5–6 (reading A)

Owned today by the `site_delivery_and_editor` lane. Sources:
`docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/PLAN_2026-08-14_site_delivery_and_editor.md`
(§Phases 5–6) and
`…/site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md` (§OWNER DECISIONS).

- **Phase 5 — customer auth: cluster-issued magic-link tokens.** Explicitly **NOT** auth-service
  (that is platform-user space) and **NOT** box-local accounts (that would fork identity off the
  `clients` chain). A third `customer_access_tokens.purpose = 'editor_session'`, exchanged at a
  new `POST /api/v1/editor/session-exchange` on core-manager (cloned from `sitefacts.go`: static
  header token, constant-time compare, fail-closed, ClusterIP+WireGuard only). Gate:
  `handed_over_at IS NOT NULL`. `"customer"` joins `humanLockSources` in `lock_policy.go`.
- **Phase 6 — the editor**, a box service on `edit.webdesign.uk` holding no cluster or provider
  credential, writing through the existing component-edit transaction with
  `locked_by='customer'`. **Sharpest risk in the whole plan, in its own words:** tenant scoping
  must be structural, proven by a cross-tenant probe (session A asking for site B's component
  ids must 404 every time).
- **Owner decision 3, 2026-08-17:** *"Account surface: v1 = the delivery email carries every
  link"* (ZIP, Netlify connect, hosting payment link, domain subscription link, Stripe hosted
  portal); **Phase 6's editor login home later becomes the account hub** (Edit / Domain /
  Hosting / Billing). **No standalone account page.** So a login page is explicitly NOT the v1.
- **Reconsideration, not decided** (`…/site_delivery_and_editor/NOTE_2026-08-25_from_webdesign_uk_lane_editor_reconsideration.md`):
  the owner asked whether customers should edit **earlier** — during the hosted preview window,
  before handover — and possibly **by voice**. Nothing decided; do not build. It moves the
  attested terms (`no changes are included`) and the register together with the code.

## 5. What is designed and unbuilt — the wider billing/entitlement layer

Register `docs/agent_docs/docs026_concept_register/register/payments.md`:
**PAY-002/003/004/005/006** are all `aspirational` — the `client_entitlements` cache, the
build-submission and maintenance-run entitlement gates, the credit ledger. **PAY-007** is the
warning: the auth-service `subscription` scaffold stamps `status = active` with **no payment**,
`GetUsageStats` returns mock zeros so `CheckQuota` always passes, and `repository.go` mixes `?`
and `$1` placeholders. **Do not gate anything on `subscription.status`.**

`BIZ-014` (`…/register/business-strategy.md`) is the strategic frame: operator-primary at
scale, vendor-optional per domain, with ownership riding the `clients→networks→sites` chain.

## 6. The hard constraint everything must respect

`[MEASURED 2026-09-04]` `kubectl -n ai-persona-system get ingress` → **No resources found**;
the only non-ClusterIP service is the WireGuard NodePort. **There is no public route into the
cluster.** Every customer-facing surface therefore lives on the box (chat, `links.webdesign.uk`)
and calls a token-gated cluster endpoint over WireGuard — the CHAT-010 facts-relay pattern.
This is why Phase 5 puts the session exchange behind a box service and not a public login page,
and it is the same problem that keeps the Stripe webhook unexposed.

`auth-service` runs (3/3 pods) and already has `/api/v1/auth/register|login|refresh` plus a
users/projects/subscriptions surface — but it is unreachable from outside and is
platform-user space, which the Phase 5 design rejected on purpose. `frontends/user-portal/`
exists and `src/App.tsx` is **0 bytes** — a stub from the persona era, not a starting point.

## 7. Questions for the owner (none of these is an engineering blocker)

1. **Which reading?** (A) an account with us, (B) hosting accounts elsewhere, or both in one
   lane. §0 above.
2. **Option C** from the OPTIONS doc — *keep hosting it ourselves and charge for it*. The
   engineering is small (`live_link_expires_at` is already per-site); the cost is a recurring
   obligation to keep the lights on and answer when they do not. Stated there as his call.
3. **Does a customer→site link get created at intake, or at handover?** Whichever, something
   must create a per-customer `network` row, because nothing does today (§1).
4. **Does the "one set of changes included" position move?** A customer editor and the attested
   terms cannot both stand as they are (§4, last bullet).

## 8. Ownership, and who NOT to compete with

- `site_delivery_and_editor` — **ACTIVE, 149 commits/14d.** Owns Phases 3–6, the delivery
  email, the links host, and the OPTIONS doc. Any Phase 5/6 build is theirs unless the owner
  moves it. Contribute by `CONTRIB_` note, do not fork.
- `bugfix_477_delivery_followup` — **ACTIVE.** Owns `platform/delivery/{delivery,handover}.go`
  and migration `778` (`site_deliveries`) — ~~uncommitted in the shared tree~~ **committed and
  applied 2026-09-04 (`698b144fa`)**; they still own those files. Stay out
  of those files.
- `web_admin_console` — owns the operator-side Customers screen that ADM-011 feeds.
