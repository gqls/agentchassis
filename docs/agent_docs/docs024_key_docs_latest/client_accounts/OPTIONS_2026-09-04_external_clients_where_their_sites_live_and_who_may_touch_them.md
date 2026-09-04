# OPTIONS — external clients: where their sites live, and who may touch them

**Written 2026-09-04 at the owner's request.** He raised three things and asked for possibilities to
be considered before anything is designed, *"even if we don't build these parts yet."* So this
proposes nothing and builds nothing. The measurements are the durable part.

> 1. *"If we build the clients domains, we will have their details in our sites directory and we
>    could if we choose be responsible for their maintenance and growth. So it is not just in
>    cloudflare and b2."*
> 2. *"I am thinking a duplicate cluster for external clients rather than using the remote job
>    spawner."*
> 3. *"We should be careful to manage the various customers of each single site cleanly. The agency
>    may want to hand over the site to their client etc or give their client access."*

---

## 1. The finding that changes the isolation argument — and it is not cost, and not abuse

The owner's point 1 is already true, and further than he may realise. A customer's site is not just
files in Cloudflare and B2: it is a row in `sites`, with pages, components, specs and an evidence
register, and **the framework keeps write access by his own ruling**. So we are *already* maintaining
and growing customer sites.

**The problem is that we cannot do it selectively.**

`[MEASURED 2026-09-04]` of **60 enabled scheduled tasks**, **34** carry a `pre_query` and **13 select
`FROM sites`**. **Not one filters on tenancy.** Two matched a `network_id|client_id|customer|tier`
regex and neither is a filter: `evidence-register-absence` matched the word *"compliance-tiered"*
inside a prose comment, and `zip-link-refresh` matched the table name `customer_access_tokens`.

The thirteen: `build-pipeline-trigger`, `claims-audit-rotation`, `evidence-register-absence`,
`improvement-sweep`, `meta-description-backfill`, `report-request-pull`,
`site-discovery-rotation-{availability,completeness,design,quality}`, `sitemap-refresh-rotation`,
`site-publish-reconciler`, `site-render-audit-rotation`.

**So the real argument for isolating customers is not hosting cost and not abuse. It is that every
fleet-wide sweep, and every change any lane ships this afternoon, currently reaches a paying
customer's live site — and nothing in the estate can express "not that one".** That is a correctness
argument, and it is much stronger than the ones previously written down.

It is also the same gap as Phase 0 wearing its most expensive coat: **the reason nothing can filter
on tenancy is that tenancy is not recorded.** `[MEASURED 2026-09-04]` one network, 42 of 60 sites
funnelling through it.

## 2. The tension between his first two points, stated plainly

They pull in opposite directions and it is better to say so than to design around it quietly.

- **Point 1 wants customer sites IN our directory** — so we can see them, maintain them, grow them,
  and bill for it.
- **Point 2 wants customers on a SEPARATE CLUSTER.**

A separate cluster with its own database **removes the single directory**. You cannot have one
`sites` table and two clusters that both own their data.

**Three ways out, and they are genuinely different products:**

- **(i) Split compute, share data.** One database, agents run elsewhere. *This is the remote job
  spawner he is rejecting* — and §3A says he is right to.
- **(ii) Split everything, keep a roll-up.** The customer instance owns its data; a periodic
  read-only export gives us one directory for billing and oversight. **But then maintenance cannot
  run from here** — our sweeps have nothing to write to.
- **(iii) Split everything, and maintenance runs THERE.** The customer instance runs its own
  improvement loop over its own sites. *"We are responsible for their maintenance and growth"* stays
  true; it simply happens on their instance, and our directory becomes a commercial roll-up rather
  than an operational store.

**(iii) is what a real multi-tenant product looks like** and it is the honest destination if point 2
is taken seriously. It is also the most work, because it means the customer instance is not a copy of
the cluster — it is a second *operator*.

## 3. Four shapes, with what each actually isolates

### A. Coupled multi-cluster dispatch — the remote job spawner. **The owner is right to reject it.**

One control plane, one Postgres, one Kafka; agents merely *run* elsewhere (register **MCL-001**;
`remote-job-spawner` is `[MEASURED 2026-09-04]` live and idle in prod, `cluster_id: uk_001`).

**It isolates compute, and compute is not what is at risk.** A bad agent still writes the shared
`sites` row, a bad migration still lands on the shared database, and §1's thirteen sweeps still reach
every customer. It moves where the CPU burns, not what can be damaged.

Its own gaps confirm it was never built for tenancy: **MCL-008** — the Kafka cluster has no
`spec.kafka.authorization`, so everything connects as `User:ANONYMOUS` with full access. A tenancy
boundary on top of an authorisation-free bus is not a boundary.

### B. Duplicate cluster, own everything — SAAS-001's "Y-copy"

Separate cluster, Kafka, Postgres, storage; a curated `agent_definitions` seed. Register status
**aspirational**; nothing stood up. **Real isolation** — this is the shape the owner is describing.

**The cost is not the servers. It is keeping in step, and it is measurable:**

| `[MEASURED 2026-09-04]` | |
|---|---|
| migrations, 30 days | **318 → 778 = 460**, ~15/day |
| migrations, 7 days | **663 → 778 = 115**, ~16/day |
| commits, 7 days | **2,956**, ~420/day |

**A second instance either tracks that or diverges from it.** Tracking means every change is deployed
twice and every migration applied twice, for ever. Diverging means customers run older software than
we do, silently, with no channel that says so. **Neither is stated anywhere as the price, and it is
the whole price.**

### C. One cluster, tenancy in the data model

Record tenancy (Phase 0), then require every sweep to honour it — plus a customer flag that fails
closed. **Cheapest by far, and needs no new infrastructure.**

**But §1 is the evidence against it:** thirteen sweeps exist today and zero are tenancy-aware, so
this is not "add a filter", it is "add a filter to thirteen things and to everything anyone writes
afterwards". And the estate's own record says that discipline fails silently — RFC_058 §5.2 documents
the fill-only-if-empty guards **inverting into a refill vector** the moment "empty" changed meaning,
at two separate layers, with nothing failing.

**A convention that thirteen existing and N future writers must remember is the failure mode this
estate has already proved twice.** If C is chosen, the filter must be structural — a view, or a
column with a NOT NULL default that forces every new query to say which population it means — never a
documented rule.

### D. A second cluster as a RELEASE CHANNEL, not a tenant boundary — *not previously written down*

**Reframe: the second cluster is not "theirs", it is "stable".** Our portfolio sites — the ones we
experiment on — stay on the current cluster and run the tip. Customer sites live on an instance that
only ever takes **released** versions.

**Why this is worth considering above the others:**

- It answers §1 exactly. The risk is not that customers share a cluster with us; it is that they
  share it with *this afternoon's change*. A release channel removes precisely that.
- **Keeping in step stops being a cost and becomes the feature.** B's 16 migrations a day is a
  burden; D's deliberate lag is the product. The estate already has the machinery — `make release`,
  image tags, a migration runner.
- It matches the instinct behind the owner's own standing rules: *"a better product beats a faster
  promise"*, and every-site-through-the-framework. Experimenting on people who paid is the thing to
  avoid, and tenancy is only a proxy for it.

**What it does not solve, stated honestly:** it is still two of everything to operate; the
"which database" question of §2 is unchanged; and a release channel needs something the estate does
not have today — **a definition of "released"**. Fleet releases are whole-fleet by owner ruling and
there is no notion of a version customers are pinned to.

## 4. Managing several customers on one site — the agency cases

This is the most concrete of the three questions and the least speculative. It is also the one where
the party/role model earns its keep, so it is worth working the cases rather than asserting.

**Recall the model** (`CONTRACT_2026-09-04_resolving_the_ordering_party.md`): a **party** is who
exists; an **identity** is a role that party holds on a site; RFC_058's four roles are *ordering ·
operating · published contact · subject*.

| case | what changes | free under the model? |
|---|---|---|
| Agency orders for its client | ordering = agency · subject = the client's business · published contact = the client's details | **yes** — this is exactly what four roles are for, and it is the case RFC_058 was raised on |
| Agency hands the site over | the **operating** role moves agency → client | **yes** — close one role row, open another |
| Agency gives the client access without handing over | both parties hold roles on the site at once | **yes** — roles are rows, so two is not special |
| Client leaves and takes the site | operating moves; agency's access must END | **yes, IF access derives from roles** — see the principle below |
| Agency with 20 clients logs in | sees 20 sites; each client sees 1 | **yes** — a party sees the sites it holds a role on |

**What is NOT free, and is genuinely new:**

**(a) Access is not identity, and RFC_058 does not cover it.** RFC_058 answers *who is involved*. It
does not answer *who may read or change what*. That is authorisation and it is a new axis.

> **The principle worth adopting now, because it is nearly free now and expensive later: ACCESS IS
> DERIVED FROM ROLES, NEVER GRANTED SEPARATELY.** If access is its own grant table, then revocation
> has two places to happen and one will eventually be missed — an agency that handed over a site two
> years ago still has a login that works. If access is a *query over live roles*, ending the role
> ends the access, in one write, with nothing to remember.

**(b) The ordering party is HISTORICAL and must never be rewritten.** When an agency hands over, it
is tempting to repoint `billing_orders.client_id` at the new owner. **That falsifies who paid**, and
it breaks refunds, disputes and the £149 record. Handover changes the *operating* role; the order is
a fact about the past and stays put.

**(c) Whose subscription is it after a handover?** If the agency pays the £10/month domain rental and
then hands over, does the client inherit it, start their own, or does the agency keep paying? Three
different products. **Nobody has asked the owner and it is not this lane's to decide** — it touches
the `stripe` lane's half directly.

**(d) The cross-tenant risk gets sharper, not milder.** The Phase 6 design already names tenant
scoping as *"the sharpest risk of the whole plan"*, with a cross-tenant probe as its acceptance —
session A asking for site B's components must 404 every time. **With agencies, the probe's premise
weakens: session A now legitimately holds site B.** The failure mode changes from *a stranger sees
your site* to *a client sees their agency's other clients*, which is worse commercially and harder to
test, because both parties have valid sessions over overlapping data. Any acceptance test written for
this must include a **delegated** session, not just an unrelated one.

## 5. What is decidable now, and what should wait

**Decidable now, and needed under every option above:**

- **Phase 0 — record tenancy.** A, B, C and D all require knowing which sites are whose. It is also
  §1's fix and §4's prerequisite. **Nothing is decidable without it and it is unblocked.**
- **Roles as rows, access derived from roles** (§4a). Cheap now, and it is what makes handover and
  delegation ordinary rather than special.
- **Ordering party is immutable** (§4b). Costs nothing to adopt and prevents a class of falsified
  record.

**Should wait, and why:**

- **The cluster question** — the owner has already parked the whole-architecture scale review until
  after the first working site, and nothing here argues for un-parking it. What §3 adds is that the
  choice should be made against **the keep-in-step figures (§3B) and the release-channel reframe
  (§3D)**, neither of which was on the table when it was parked.
- **Which database owns customer sites** (§2 i/ii/iii). This is the decision that actually determines
  the product, and it should not be taken as a side effect of choosing a cluster topology.
- **Subscription-on-handover** (§4c) — a question for the owner, jointly with the `stripe` lane.

## 6. The one-line reading, if only one line is read

> **Isolating customers is not about hosting cost or abuse. It is that thirteen fleet sweeps and
> every change we ship reach a paying customer's live site, and nothing can express "not that one" —
> because tenancy is not recorded. Record it first: every option above needs it, and it is the only
> part that is unblocked today.**
