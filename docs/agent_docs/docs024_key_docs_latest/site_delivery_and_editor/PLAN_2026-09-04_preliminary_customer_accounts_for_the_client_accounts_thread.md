# Preliminary plan — customer accounts on our own system

Written 2026-09-04 by `site_delivery_and_editor` at the owner's request, **for the new client-accounts
thread to own**. It is a starting point to argue with, not a design to implement: the measurements
are the durable part, the phasing is a proposal, and the decisions listed at the end are the owner's.

---

## 1. What exists today, measured rather than assumed

| thing | state |
|---|---|
| `clients` | **3 rows**: "Default Client", "System Scheduler", "Boxing Online". 11 columns including `tier` and `customer_status` — **both empty on all three** |
| `networks` | carries `client_id` |
| `sites` | **60 rows, 40 columns.** No `client_id` and no `customer_id` — it links only via `network_id` |
| `billing_orders` | **1 row.** Has `client_id`, `amount_pence`, `status`, `provider`, `paid_at` |
| `build_queue` | 3 rows; the ordering email lives in `direction->>'customer_email'` |
| `customer_access_tokens` | 2 rows, both from last night's rehearsal. Purpose-scoped, hashed, expiring |

### ⚠ The finding that sets the whole problem

**We cannot answer "which sites belong to this customer?" today.** The join path exists —
`clients → networks.client_id → sites.network_id` — and it is empty of meaning:

> `[MEASURED 2026-09-04]` 60 sites · 42 carry a `network_id` · **1 distinct network between them** ·
> 18 carry none at all.

So every linked site funnels through a single shared network. "Boxing Online" exists as a client row
with an email and **no reachable sites**. The structure is there; the data has never used it.

That is good news for a plan: nothing has to be undone. But it means **"keep track of them and their
builds" is not a reporting feature — the relationship it would report does not exist yet.**

---

## 2. What is already ruled, and must not be re-litigated

**`RFC_058` — OWNER RULING 2026-09-03: Option C, four identities.** *Ordering party · operating party
· published contact · subject.* Plus two additions he made himself:

- a fifth **selling-party** identity, which he **explicitly deferred**;
- **multi-valued contacts** — *"there may be more than one contact detail for any of these"* — which
  he **did not** defer.

⚠ **The two are not independent, and this is the trap that will cost the most if missed.** Multi-valued
contacts force a *relation*. A relation is also what makes the deferred fifth identity cheap: under a
relation a new identity is a **row**; under columns it is a **migration**. **A thread that shortcuts
contacts back into columns "for now" silently converts the owner's deferral into future migration
work.**

**`RFC_058` §5.2's trap, which an accounts thread will walk into first:** making `sites.email` a
*derived view* of the published contact means **every reader must learn which identity it is
reading**. A reader the census misses keeps serving the pre-ruling identity indefinitely, with
nothing failing.

**Owner ruling 2026-09-04 (today):** the delivery recipient goes in **its own delivery record**, not
as a column on `sites` — because `sites` is read by many things and 420's contract split governs
which address may live where. **Any accounts schema should inherit that instinct: customer PII lives
on narrow tables, not on `sites`.**

---

## 3. The three things "an account" means, which must not be conflated

**Identity** — who they are. *Already ruled: RFC_058 Option C. Do not invent a parallel model.*

**Authentication** — how they prove it. Not solved, and **not obviously needed in phase 1** (§4).

**Entitlement** — what they are owed: a build, hosting, a domain. `billing_orders` and the empty
`tier` / `customer_status` columns are the beginnings of this.

Most account systems fail by building authentication first, because it feels like the product. Here
the missing thing is **entitlement joined to identity**, and authentication is a way of reading it.

---

## 4. Proposed phasing

### Phase 0 — make the relationship exist (no new concepts)

Give a site an owner. Today that is a network with one row in it.

The cheapest honest form is to **use the chain that already exists** — one network per client, sites
pointed at their own — rather than adding `sites.client_id` beside it. Two paths to the same fact is
the divergence that later costs a census.

Deliverable: every existing site attributed, or explicitly marked as ours. **18 sites carry no
network at all**, so this is partly an archaeology exercise and should be scoped as one.

### Phase 1 — a customer-facing page with NO login

**We already ship customers a token-addressed page.** `/c/<token>` and `/d/<token>` are live, the
token machinery is built (hashed, purpose-scoped, expiring, single-use optional), and `776`'s email
already carries such links.

So the first "account" can be a **read-only page at a token**: your site, your files, when hosting
ends, what you have paid for. No password, no reset flow, no email verification — none of the forty
minutes the owner just spent on Netlify's signup.

**This is the phase that delivers most of what was asked for at a fraction of the cost**, and it is
worth doing even if real accounts follow, because it is the fallback for anyone who never logs in.

### Phase 2 — hosting as a renewable entitlement

Hosting is already **90% built and nobody has noticed**: `sites.live_link_expires_at` is a per-site
hosting expiry, set to `now() + LiveLinkWindow` (6 weeks) at handover. We serve 42 days and advertise
30, deliberately.

So "offering hosting" is mostly: make that column **renewable**, attach renewal to a payment, and
decide what happens at expiry. `billing_orders` already has `provider`, `provider_customer_id` and
`paid_at`, so recurring billing has somewhere to sit.

⚠ **The thing to get right: what expiry DOES.** Today it is a promise in an email. If it becomes a
switch that takes a paying customer's site down, it needs the once-only discipline the handover stamp
has, and a grace period, or a billing hiccup unpublishes somebody's business.

### Phase 3 — authentication, if still wanted

Only worth it when there is something a customer must *do* rather than *see*. `auth-service` exists
but is ours; a customer login is a separate surface with password resets, verification and support.

**Netlify's signup is the argument for delaying this**: the owner's own run took ~40 minutes, with an
invisible security check and a rejected good password. Every one of those is a support email we would
be signing up for.

---

## 5. Decisions the owner owes before this can be planned properly

1. **Do customers log in at all, or is a token-addressed page enough?** This is the biggest cost fork
   and phase 1 is designed to defer it.
2. **What does hosting expiry actually do** — stop serving, show a holding page, or only stop
   renewing the domain? The answer changes phase 2 from a column to a subsystem.
3. **One account per person, or per site?** Someone with three sites is a different product from
   three separate customers, and it decides whether the account or the site is the entity.
4. **Does a customer account subsume the RFC_058 identities, or sit beside them?** My reading: it
   *is* the ordering party, and the other three stay as they are. Worth stating explicitly, because
   getting it wrong duplicates the model.

## 6. What I would NOT do

**Do not add `sites.client_id`.** The chain exists; a second path to the same fact is the divergence
that later costs a census, and today's ruling on the delivery recipient points the same way.

**Do not build authentication first.** It is the most visible part and the least of the value.

**Do not model contacts as columns**, even temporarily. RFC_058's ruling makes that a debt, and the
owner's deferral of the fifth identity is only cheap under a relation.

## 7. Handover

Owned by the client-accounts thread from here. This lane holds the delivery half — handover stamp,
tokens, the delivery email, the customer-facing `/c/` and `/d/` pages — and will keep it working; the
token machinery in phase 1 is ours and we are happy to be called on for it.

Related: `RFC_058` · `bugs_open/420` (the contract split) · `bugs_open/475`, `476`, `477` (what the
customer currently receives) · `OPTIONS_2026-09-04_what_it_would_take_to_set_up_hosting_accounts_for_customers.md`
(the Netlify-vs-us question, which phase 2 partly answers).
