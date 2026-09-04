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
4. `site_deliveries.delivered_to` — who a delivery actually went to.
   ~~**New, in flight in another session's tree**~~ **CORRECTED 2026-09-04, same day: COMMITTED
   (`698b144fa`) AND APPLIED** — `[MEASURED 2026-09-04]` the table is live with one backfilled row
   (`recorded_by = 'backfill-orchestration-states-778'`, 14:16:57Z); the Go half rides the v1.0.1361
   cut, unrolled at the time of writing. (`docs/agent_docs/sql_for_agents/778_…sql`, written by
   `platform/delivery.Claim`.) It exists because the obvious column was populated and wrong: idea.uk
   carries `idea.uk@contactforsales.com` while the delivery went to the customer's real address.
   **The staleness is the point of the footnote:** this store was described here as unbuilt and was
   live within hours, which is exactly why §2d argues that the count of stores is the thing to
   watch.

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
  `phone` and `external_id`. Contacts move out of it into rows; it keeps identity.

  > **CORRECTED 2026-09-04, mine, same day:** this line originally read *"`external_id` (the Stripe
  > key)"*. **False as a statement about data.** PAY-009's register entry describes the *intent* —
  > Stripe's customer id landing there from the paid event, first writer wins — and I repeated it as
  > though the join were live. `[MEASURED 2026-09-04]` `clients.external_id` is **empty** on the only
  > real customer row, `billing_orders.provider_customer_id` is **NULL**, and the one paid order's
  > webhook event carries `"customer": null` because a one-off `mode=payment` charge with
  > `customer_creation: "if_required"` makes no Stripe Customer. **The column is unused by Stripe
  > today.** Caught by querying it in order to answer the `stripe` lane, which had quoted the same
  > entry the same way — so two of us read a design intent as a live join off one register entry.
  > The cheap check: **before describing a register entry's mechanism as live, select the column.**
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
  and migration `778`. ~~uncommitted in the shared tree~~ **CORRECTED 2026-09-04: committed
  (`698b144fa`) and applied; the Go half rides the v1.0.1361 cut.** Still theirs — stay out of those
  files.
- **`bugfix_417_420` owns RFC_058.** Any identity/role schema is a contribution to that RFC, not a
  parallel design.
- **`finetuning_uk_service` is about to need the same thing** —
  `DESIGN_2026-09-03_examples_catalogue_shape.md` §3 says outright *"The estate does not have a
  customer account system today for framework sites … So this is the first real one, and it is the
  largest piece here."* **Not yet told.** They should be, before they build a second one.

---

# OWNER RULINGS 2026-09-04 — four answers, and two of them close §5's biggest fork

Put to the owner directly by this lane on the day it took the lead. Quoted where the option text is
quoted; the consequences below are this lane's reading and are marked as such.

## Ruling 1 — scope: **an account WITH US.** Not third-party hosting accounts.

`OPTIONS_2026-09-04_what_it_would_take_to_set_up_hosting_accounts_for_customers.md` (the delivery
lane's Netlify question) is **not** this lane's subject. Its Option D — make the handoff good — is
already shipping there and stays theirs.

## Ruling 2 — paid hosting: **cost it first, do not build.**

So pre-plan Phase 2 becomes a costing deliverable, not a build. §2e already splits it and the split
is the finding: **keeping a paying site up costs nothing to build** (nothing enforces
`live_link_expires_at`, so it is the current behaviour by omission), and **the entire cost is the
stop-serving half** — once-only discipline, a grace period, and the risk of unpublishing a business
over a billing hiccup.

The costing must therefore price three things separately and not blend them:
(a) the money — marginal storage and delivery per site, which is **`[UNMEASURED]`** today and has no
figure anywhere in the tree; (b) the **obligation** — uptime and someone to answer, which the OPTIONS
doc already named as the real cost; (c) the **stop-serving subsystem**, which is the only real
engineering. Note before costing (b): detection partly exists —
`[MEASURED 2026-09-04]` `scheduled_tasks.site-discovery-rotation-availability` runs every 5 minutes
and probes the least-recently-checked deployed site for `site_unreachable` (`bugs_open/236`).

## Ruling 3 — **token page now, login later.** Login is a stated destination, not a maybe.

> *"Token page now, login later … so the token page is built as the fallback for people who never
> log in, not as a throwaway."*

**This is not the same answer as "token page only", and the difference is a design constraint, not a
mood.** Because a login is coming, the token page must be built so that **what the token grants and
what a password would later grant are the same thing.** Concretely, this lane's reading:

- the redemption step yields a **scoped read of an account**, and the token is one way to obtain it;
- a password later becomes a *second* way to obtain the same scope, not a second surface with its own
  notion of what a customer may see;
- so **the scope is the thing to get right now** — what an account can read — and authentication is
  deliberately left as the swappable half.

⚠ **The trap this ruling creates, recorded before anyone hits it:** "login later" invites building
the page against a token and *then* discovering the login needs a different shape. The guard is to
name the scope explicitly in the register entry when Phase 1 ships, so the Phase 3 session inherits a
contract rather than a page.

Phase 3's price is unchanged and stands: §2c — there is no public route into the cluster, so a login
is a new public surface on the box with password resets, verification and support attached.

## Ruling 4 — **one account per PERSON**, not per site.

> *"The account is the customer. Someone with three sites is one customer with three sites, not three
> customers."*

This confirms §3's party-vs-role reading from the owner's side, and it **settles §2b's token fork**.

**What it resolves, and how.** The account page is client-scoped; the existing links
(`zip_download`, `confirm_transfer`) are genuinely about a site and stay per-site. So the token table
must carry both shapes. Of §2b's three options:

- **(i) one page per site is now ruled out** — it contradicts this ruling directly.
- **(iii) a second token table is rejected by the pre-plan's own §6** — two paths to the same fact is
  the divergence that later costs a census, and migration 511's header says outright that the table
  exists *because* two customer links were the same mechanism and a third was coming. A second table
  would undo the convergence it was written to achieve.
- **(ii) is the shape:** `site_id` becomes nullable, `client_id` is added, and a **purpose-aware
  CHECK** requires exactly the right one per purpose — `zip_download` / `confirm_transfer` require
  `site_id`; the new account purpose requires `client_id`. One hashing rule, one expiry rule, one
  redemption path, one ledger.

⚠ **Do NOT make `site_id` plainly nullable and stop there.** That would silently weaken a NOT NULL
guarantee the two live purposes rely on, and nothing would fail — a `zip_download` row with a NULL
site would be accepted and only break at redemption. The CHECK is the load-bearing half of this
migration, and it should be **induced** in the migration's verify block the way 511 induced its own
two, not merely asserted to exist.

## What these four rulings unblock, in order

1. **Phase 0, producer half** — `EnsureSiteRecordAction` learns to take a network from the order.
   Ruling 4 makes this unambiguous: the network being attributed is the **customer's**, one per
   party. Blocked on nothing.
2. **Phase 0, backfill** — after (1), never before (§2a).
3. **Phase 1 migration** — the purpose-aware token widening above, plus the new account purpose.
   Blocked on nothing; council-scope (a migration IS the running system).
4. **Phase 1 page** — the scope contract first, the page second (ruling 3).
5. **Phase 2 costing** — independent of all the above; can run in parallel.

**Still owed by the owner, and not urgent:** what hosting expiry actually DOES (stop serving / holding
page / domain only). The costing in (5) will price all three rather than wait for it.

---

# OWNER SUGGESTION 2026-09-04 — *"we could have a separate cluster for them altogether"*

Recorded, not adopted, and **not a decision** — the owner raised it mid-plan. It has real prior art
in this estate, so the useful reply is to say which of three different things it would isolate,
because they have three different answers.

## The three things a "separate cluster" could isolate

**1. Where customer sites are SERVED. A separate cluster changes nothing here.** Sites do not serve
from the cluster at all — they serve from B2 `portfolio-sites` through one Cloudflare Worker across
38 zones plus the `*.ugg2.com` wildcard. The shared-fate risk at this layer is **one Cloudflare
account**: `PLAN_2026-08-17_delivery_architecture_decisions.md` §4.2 names it plainly — *"one
phishing page can flag the shared account"* — and §4.4 already costs the fix as an **A → B → C**
trajectory (zone-per-domain now → own authoritative DNS + CF for SaaS near ~500–1,000 domains →
account sharding / CF Tenants only if partner status earns its keep). **That is the isolation
question that actually bites first, and it is not a compute question.**

**2. Where customer BUILDS run. This is the real prior proposal, and it is closer than it looks.**
Register **SAAS-001** — the *"Y-copy"* isolated-satellite architecture: a cut-down copy of the whole
chassis on a separate cluster, decomposing *"don't let customer work interfere with core"* into
three threats (load, hack, bug) across a one-directional, egress-from-core-only boundary. Its own
escalation argument is exactly this lane's subject: *"an anonymous, internet-triggered,
token-spending build pipeline must not run on core."* **Status: aspirational; nothing stood up.**

But the plumbing is further along than the status suggests, and this is checkable:

- `[MEASURED 2026-09-04]` **`remote-job-spawner` is LIVE in the production cluster right now** —
  1/1, 187 days old, `cluster_id: uk_001`, consuming `system.dispatch.requests`, **and idle.**
- `dispatch_agent` is a registered action (`platform/orchestration/actions/dispatch_actions.go`),
  publishing to that topic with `target_cluster` in the message headers.
- `[MEASURED 2026-09-04]` `kubectl config get-contexts` shows **one** cluster,
  `uk001-prod-agent-chassis-cluster`. MCL-002's `va001` second cluster is not in this kubeconfig.

**So the receiving half of multi-cluster dispatch is deployed and has nothing to do. What is missing
is a second cluster and a reason — not the mechanism.** Known gaps before it could be trusted are
already enumerated: MCL-003 (the cluster filter logs at `Debug`, which does not appear in our logs,
so you cannot verify the filter works), MCL-004 (no consumer for `system.dispatch.responses` — a
failed dispatch is silent until timeout), MCL-008 (Kafka has **no** `spec.kafka.authorization`, so
everything connects as `User:ANONYMOUS` with full access — an authenticated cross-cluster user would
be unrestricted). **MCL-008 is the one that matters for a customer-facing satellite** and it is a
bigger change than the satellite itself.

**3. Where customer DATA lives. Here it cuts against today's finding.** §2d: a customer already
lives in four unjoined stores and nothing joins them — and the fourth was created *this week*
because a reader trusted the third. **Splitting the database across clusters before the join exists
adds a fifth store, not isolation.**

## The sequencing argument, which is the actual answer

**You cannot shard a relationship you have not recorded.** *"Which customer's data moves to the other
cluster?"* is precisely the question the estate cannot answer today: `[MEASURED 2026-09-04]` one
network, 42 of 60 sites funnelling through it, and the one real customer row reachable from no site.

So this suggestion **does not reorder the plan — it raises the value of Phase 0**, which is the same
work under every isolation story and is a prerequisite for all of them.

And the estate's own recorded caution points the same way. **BIZ-014**: the unit of blast-radius
isolation (the satellite/cluster) is *distinct* from the unit of separability-for-sale (the domain),
and **"operating thousands of domains does not require thousands of clusters."** That is not an
argument against a satellite; it is an argument against reaching for a cluster when the thing you
want is a boundary.

## What it would buy, and what it would cost

**Buy:** the strongest available answer to §4.2's abuse/shared-fate problem; the shape that makes
"sell the whole thing to a buyer" cheap later (BIZ-014's vendor-optional half); and SAAS-001's
stated goal of keeping anonymous internet-triggered spend off core.

**Cost, stated honestly:** a second cluster is a second everything to operate — and
`webdesign_uk_build_service/PLAN_2026-08-21_todo_from_here.md` records that the owner has already
**parked** the *"whole-architecture scale review incl. own cluster(s)"* until **after** the working
site. Adopting this now would be reversing his own parking decision, which is his to do but should
be done knowingly.

## This lane's recommendation

1. **Do not reorder the plan.** Phase 0 is unchanged and is the prerequisite either way.
2. **Fold it into the Phase 2 costing** (ruling 2), as a fourth line beside money / obligation /
   stop-serving: *what isolation would cost and buy*. It belongs there because it is an answer to
   the same operational-obligation question, and the costing is already the deliverable.
3. **Route the compute half to the scale review**, where the owner already put it — not into this
   lane. **Tell the `multicluster` lane** that a customer-facing satellite is being considered
   again, because MCL-003/004/008 are theirs and MCL-008 gates it.
