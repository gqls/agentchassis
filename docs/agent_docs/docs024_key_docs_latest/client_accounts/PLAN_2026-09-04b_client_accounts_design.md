# PLAN — 2026-09-04b — client accounts: the design, led from here

**This lane leads from here** (owner, 2026-09-04). It supersedes nothing; it builds on two files
and neither should be re-derived:

- `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/PLAN_2026-09-04_preliminary_customer_accounts_for_the_client_accounts_thread.md`
  — the pre-plan written *for* this thread by the delivery lane. **Accepted as the base.** Its
  phasing is adopted; its §6 "what I would NOT do" is adopted verbatim.
- `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_058_the_platform_stores_one_identity_where_the_business_has_at_least_three.md`
  — **OWNER RULING 2026-09-03: Option C, four identities, plus multi-valued contacts (not
  deferred) and a fifth selling-party identity (deferred).** This is decided. Do not re-open it.

`PLAN_2026-09-04_client_accounts.md` in this directory is the survey that opened the lane and stays
as the evidence base. **Owner scope ruling 2026-09-04:** "client accounts" means an account **with
us** (not third-party hosting accounts), and paid hosting is to be **costed, not built**.

This file records what the pre-plan did not have, where I differ, and what the owner is owed.

---

## 1. Corroboration first — two lanes measured this independently and agree

The delivery lane and this one measured the ownership chain separately, hours apart, and got the
same answer: **the chain is a schema with no data in it.** That agreement is worth more than either
measurement alone, so it is recorded rather than merged away.

⚠ **The three counts differ by date and all three are correct.** RFC_058 §2 measured
`sites.network_id` set on **33 of 54** `[MEASURED 2026-09-02]`; both lanes measured **42 of 60**
`[MEASURED 2026-09-04]`. Nothing regressed — the estate grew by six sites in two days. This is the
census-staleness rule in miniature: **quote the date with the number, or the next reader reconciles
two true figures as a contradiction.**

## 2. What I am adding to the pre-plan — five things it did not have

### 2a. Phase 0 is a PRODUCER change, not only archaeology — and the producer half is the durable half

The pre-plan scopes Phase 0 as "every existing site attributed, or explicitly marked as ours …
partly an archaeology exercise". True, and incomplete: **backfilling attribution decays the moment
the next site is built, because the build path cannot express it.**

`EnsureSiteRecordAction` (`platform/orchestration/actions/site_db_actions.go:178`) — the action that
creates every site row — **takes no network parameter at all.** It calls `getDefaultNetworkID`
unconditionally, and on failure falls back to the hardcoded literal
`00000000-0000-0000-0000-000000000002` (`:182`, and again at `:1031` in
`createPlaceholderSiteRecord`). `createDefaultNetwork` (`:1085`) inserts `slug='default'` with
`ON CONFLICT (slug) DO UPDATE`, so it can only ever produce one row.
`[MEASURED 2026-09-04]` `grep -rn "INSERT INTO networks" --include=*.go .` → **one hit**, that
function.

So Phase 0 has two halves and they must land together or the first is wasted:

1. **the producer** — `EnsureSiteRecordAction` learns to take a network from the order (the
   collector already threads `order_reference` / `customer_email` through `build_queue.direction`,
   so the carrier exists);
2. **the backfill** — the 42 attributed to the right party, the 18 with no network at all treated as
   archaeology, and everything of ours explicitly marked ours rather than left to be inferred.

**Order matters and it is the opposite of the intuitive one:** ship the producer first. A backfill
against a producer that still writes `default` is a census that is stale the day it is taken.

### 2b. The token machinery is structurally PER-SITE, and an account page is not

The pre-plan's Phase 1 — a read-only page at a token, no login — is the strongest idea in it and I
am adopting it. But `customer_access_tokens.site_id` is **`uuid NOT NULL REFERENCES sites(id)`**
(`docs/agent_docs/sql_for_agents/511_handover_state_and_customer_tokens.sql`). Every token is bound
to exactly one site.

So "your sites, your files, when hosting ends, what you have paid for" — a page about a *customer* —
**cannot be keyed on the existing token as it stands.** Three ways out, and the choice is not free:

- **(i) accept one page per site.** Zero schema change. Correct for a one-site customer, which is
  every customer we have. Wrong the first time somebody buys a second site.
- **(ii) widen the token to key on a client** (`site_id` becomes nullable, `client_id` added,
  a CHECK requiring exactly one). Small, and it makes the account the entity.
- **(iii) two token kinds** — per-site links stay as they are, a new client-scoped purpose is added
  beside them.

**This is the same fork as pre-plan question 3 ("one account per person, or per site?") wearing a
different coat** — which is why it belongs in the decision list rather than in a build.

Either way it costs a migration, **and that is by design**: `purpose` is a closed CHECK vocabulary
(`[MEASURED 2026-09-04]` live: `CHECK (purpose = ANY (ARRAY['zip_download','confirm_transfer']))`)
precisely so a fourth kind of customer link is visible in the ledger rather than appearing in a Go
constant. `editor_session` is the reserved third and does not exist yet.

### 2c. There is no public route into the cluster — which prices Phase 3 and vindicates Phase 1

`[MEASURED 2026-09-04]` `kubectl -n ai-persona-system get ingress` → **No resources found.** The
only non-ClusterIP service in the namespace is `wireguard` (NodePort).

So a customer login page cannot be served from the cluster at all. It has to live on the box and
call in over the tunnel — which is exactly what the Phase 5 design in
`site_delivery_and_editor/PLAN_2026-08-14_site_delivery_and_editor.md` already says, and it is
**a whole new public surface with its own hardening, not a route on something we already run.**

Two corollaries worth stating so nobody offers them as shortcuts:

- **`auth-service` is not the shortcut it looks like.** It runs (3/3 pods, 399 days) and already
  exposes `/api/v1/auth/register|login|refresh` (`cmd/auth-service/main.go:185`). It is unreachable
  from outside, it is platform-user space, and the Phase 5 design rejected it on purpose. Its
  subscription scaffold is worse than useless for entitlement: register **PAY-007** — it stamps
  `status = active` with **no payment**, and `GetUsageStats` returns mock zeros so `CheckQuota`
  always passes. **Never gate anything on `subscription.status`.**
- **`frontends/user-portal/` is a persona-era stub.** `src/App.tsx` is **0 bytes**. Say so before
  anyone cites its existence as prior art.

This makes the pre-plan's Phase-1-before-Phase-3 argument stronger than it states it: the case is
not only that a login is expensive in support, it is that **we do not have a front door to put one
behind.**

### 2d. A customer's details live in FOUR unjoined stores, and none of them is the ruled one

Beyond the pre-plan's table, the stores that actually hold a paying customer today:

1. `build_queue.direction.customer_email` / `customer_name` / `order_reference` — the payer, written
   by `collect_external_orders_action.go`, and its durable home per `bugs_open/420`'s contract split.
2. `billing_orders.client_id` + `external_reference` — the payment.
3. `sites.email` / `company_name` — the **published** contact only, written by
   `seedCustomerIdentity` (`seed_build_queue_action.go:317`) and only from
   `direction.published_contact`. Never the payer.
4. `site_deliveries.delivered_to` — who a delivery actually went to. **New, in flight in another
   session's tree** (`docs/agent_docs/sql_for_agents/778_…sql`, written by `platform/delivery.Claim`).
   It exists because the obvious column was populated and wrong: idea.uk carries
   `idea.uk@contactforsales.com` while the delivery went to `aaa@designconsultancy.co.uk`.

`clients` — the store the owner ruled canonical on 2026-08-10 — is written by **none** of these. Its
only writer is `POST /admin/customers` (`internal/core-manager/admin/customer_handlers.go:190`).

**Why this matters to the design rather than being a tidy-up:** store 4 was created *this week*
because a reader trusted store 3. Every additional unjoined store is another place a future reader
will confidently read the wrong identity. **The account is the thing that joins them, and until it
does, adding stores is the failure mode, not the progress.**

### 2e. Nothing enforces hosting expiry — so "we host for six weeks" is currently false in the safe direction

`[MEASURED 2026-09-04]` every Go reference to `live_link_expires_at` is a write at handover, a
follow-up-email eligibility predicate, or a test. **Nothing stops serving.** Serving is unbounded,
exactly as migration 511's header warned it would be until a retraction job existed.

This splits the owner's "cost it first" question cleanly, and the split is the useful part:

- **keeping a paying customer's site up past 30 days costs approximately nothing to build** — it is
  the current behaviour, by omission;
- **all of the engineering is in the other half** — stopping for the people who do not pay. That is
  where the once-only discipline, the grace period, and the risk of unpublishing somebody's business
  over a billing hiccup all live. It is a subsystem, not a column.

## 3. Where I differ from the pre-plan: the account is a PARTY, and identity is a ROLE

The pre-plan's question 4 asks whether a customer account subsumes the RFC_058 identities or sits
beside them, and reads it as *"it is the ordering party"*. I think that is right and one step short,
and the missing step dissolves question 3 as well.

**RFC_058's four identities are ROLES ON A SITE.** A site has an ordering party, an operating party,
a published contact, a subject. **A customer account is a PARTY** — one organisation or person, who
may play the ordering role on three sites and the operating role on none of them.

They are different cardinalities, so they cannot be the same row:

> **parties** (who exists) — one row per person or organisation, with contacts as rows beneath it,
> each contact carrying its purpose and its own consent state (RFC_058's own direction of travel).
> **roles** (who is what, where) — one row per (site, role, party). Four roles today, a fifth
> deferred; under this shape the fifth is an **insert**, which is the whole reason the owner's
> deferral is cheap.
> **an account** is a party we have additionally given a way to *read* — and later, perhaps, to
> authenticate.

**What this buys, stated as claims that can be checked rather than as elegance:**

- **pre-plan question 3 answers itself.** One account per **party**. Someone with three sites is one
  party holding three ordering roles — not three customers, and not a decision anybody has to make
  again later.
- **the deferred fifth identity stays an insert**, which is the condition RFC_058 attaches to its own
  deferral being safe.
- **`clients` is not a fifth store.** It is the parties table, already carrying `name`, `email`,
  `phone`, `external_id` (the Stripe key). Contacts move out of it into rows; it keeps identity.
- **the pre-plan's §6 "do not add `sites.client_id`" survives intact** — and gains a reason. A
  direct site→client column would encode *one* role and silently name it "the owner", which is the
  conflation RFC_058 exists to end.

⚠ **The honest cost of this shape, stated so it is not discovered later:** it makes Phase 0 slightly
larger than "one network per client", because the network chain expresses exactly one relationship
and this shape has four. **Phase 0 should still ship on the network chain** — it is what exists, the
pre-plan is right that a second path to the same fact is the divergence that costs a census, and
attribution is needed before roles are. But it should ship knowing it is the *ordering* role being
recorded, and say so in the register entry, so the later role table subsumes it rather than
competing with it.

## 4. Phasing, as adopted

| phase | what | change from the pre-plan |
|---|---|---|
| **0** | make the relationship exist — producer first, then backfill | **producer half added (§2a); ship it first** |
| **1** | a read-only customer page at a token, no login | adopted; **the per-site/per-client token fork (§2b) must be decided first** |
| **2** | hosting as a renewable entitlement | **owner ruled: COST IT, do not build.** §2e says which half is free and which is the subsystem |
| **3** | authentication, if still wanted | adopted, and **priced up** by §2c — there is no front door |

## 5. Decisions the owner owes (the pre-plan's four, re-cut against what I found)

1. **Do customers log in at all, or is a token-addressed page enough?** Unchanged from the pre-plan,
   still the biggest fork, and §2c raises the price of "yes": there is no public route into the
   cluster, so a login is a new public surface on the box, not a route on something we run.
2. **One account per person, or per site?** §3 argues this answers itself if the account is a party
   — but it is his call, and it **decides §2b's token fork**, which is a migration either way.
3. **What does hosting expiry actually DO** — stop serving, show a holding page, or only stop
   renewing the domain? §2e: today it does nothing at all, so this is a decision about what to
   *build*, and only the non-payer half costs anything.
4. **Does the account subsume the RFC_058 identities or sit beside them?** §3's answer: the account
   is a **party**; the identities are **roles**. Worth him confirming, because getting it wrong
   duplicates the model — and duplicating the model is how we got four unjoined stores.

## 6. Ownership and coordination

- **This lane leads.** `site_delivery_and_editor` keeps the delivery half — handover stamp, tokens,
  the delivery email, `/c/` and `/d/` — and has offered to be called on for the token machinery
  Phase 1 needs (their §7). Take them up on it rather than re-implementing.
- **`bugfix_477_delivery_followup` is ACTIVE** and owns `platform/delivery/{delivery,handover}.go`
  and migration `778` right now, uncommitted in the shared tree. Stay out of those files.
- **`bugfix_417_420` owns RFC_058.** Any identity/role schema is a contribution to that RFC, not a
  parallel design.
- **`finetuning_uk_service` is about to need the same thing** —
  `DESIGN_2026-09-03_examples_catalogue_shape.md` §3 says outright *"The estate does not have a
  customer account system today for framework sites … So this is the first real one, and it is the
  largest piece here."* **Not yet told.** They should be, before they build a second one.
