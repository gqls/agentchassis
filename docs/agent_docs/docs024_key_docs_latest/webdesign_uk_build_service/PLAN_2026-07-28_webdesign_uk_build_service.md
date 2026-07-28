# PLAN — webdesign.uk: selling web design and build, using the framework

**Started 2026-07-28**, on the owner's direction. **This is a proposal to react to,
not a settled design.** Nothing is built. Where I am arguing rather than recording
the owner's words, it says so.

This executes a direction already on the record: `webdesign_couk/PLAN_2026-07-27b_buying_design.md`
§8 recorded the owner's next build as *"a website creation form using our system…
we stand up a copy chassis in its own cluster with its own database"*. That section
said **recorded now, not started**. This is the start.

---

## 1. The ask, restated

> Minimal page, little more than a chat box: *enter your domain name here*. From
> that we can, if you choose, create a website. A better optional route is the
> briefing questionnaire — email, telephone, the rest. Then payment. Then the
> build. Then the result goes to a subdomain of one of my domains and the link is
> sent the next day or so. We have no system to directly call the chassis and we
> should probably set up a dedicated cluster so the existing one doesn't get
> hacked. Keep it in line with tools-api, idea.uk and relojistas. Human-curated to
> start with.

Domain: **webdesign.uk**, deliberately separate from webdesign.co.uk so the two do
not interfere. **DNS is not pointed yet** (owner, 2026-07-28) — confirmed against
the wire the same day: `dig webdesign.uk A` and `NS` both return nothing, while
webdesign.co.uk resolves to Cloudflare.

---

## 2. What already exists — measured, not assumed

Every line here was checked today. The full commands and their output are in
`NOTES_webdesign_uk_build_service.md`; the ones worth reusing are in the RUNBOOK.

| capability | state | evidence |
|---|---|---|
| **Paid funnel on a VM** | **Built and has taken real money.** idea.uk: stdlib-only Go binary, Stripe behind a `Provider` interface, webhook = source of truth, HMAC-verified, `FakeProvider` for end-to-end tests, price passed **per call** not held on the provider. | `idea.uk/golang_files/billing.go`, `store.go`, `service.go`; first sale 2026-07-27 |
| **Public HTTP tool endpoint** | **Built, one worked example.** gin + own Postgres, CORS → rate-limit → input-cap middleware, request logging. Zero Kafka, zero chassis coupling. | `internal/tools-api/api/server.go:32-58` |
| **Static site build + deploy** | **Live for 12 sites.** chassis → git-adapter → repo `sites` → GitHub Action → B2. Per-site override to `vm-sites` for VM-hosted sites. | `platform/orchestration/actions/git_deployer_actions.go`, `git_repo_resolution_test.go` |
| **VM-hosted site with its own backend** | **Live twice.** idea.uk and relojistas.com: nginx + certbot, site-engine under systemd, box pulls its own folder from `vm-sites`. | `sites.github_repo='vm-sites'` for both |
| **A way to call the chassis from outside** | **Does not exist for untrusted callers.** core-manager has `POST /api/v1/admin/pipelines/:name/trigger` — behind `AuthMiddleware` + `AdminOnly()`, served inside the cluster behind api-gateway. tools-api holds no Kafka client at all. | `internal/core-manager/api/server.go:108-227`; `cmd/tools-api/main.go` |
| **Automated "domain in → site out"** | **Does not exist.** New sites are created by hand-authored SQL seeding the `sites` row and its `site_specs` aspects, then trigger scripts publish to Kafka. | `oufe/SEED_2026-07-25_oufe_site_and_specs.sql`; oufe has 12 aspects seeded on day one |
| **Per-call model cost** | **Measurable.** `llm_call_log` holds 45,205 rows, 2026-03-25 → 2026-07-28. | live query, 2026-07-28 |
| **Fleet size** | **32 site rows, 14 deployed, 17 `pool-*` shells.** See §12 — this matters. | live query, 2026-07-28 |

**The single most important line in that table** is the one that says the
chassis-calling seam does not exist. The owner already knew. What it means for
phasing is in §9: **it is not needed to start selling.**

The second most important is the last one but two: **site creation is currently a
human writing SQL.** That is not a gap to close before launch — it is precisely
the "human-curated/managed to start with" the owner asked for, and it already
works. oufe.com went from seed to 8 deployed pages in three days; fundamentallyai.com
13 pages in five.

---

## 3. The four seams

Everything in this product is one of four things. Keeping them apart is what stops
this becoming "a big new system".

1. **The shopfront** — webdesign.uk itself: the page, the domain box, the
   questionnaire, the order record. Public. Untrusted input.
2. **Payment** — Stripe Checkout, webhook as source of truth.
3. **The build trigger** — how a paid order becomes a build in the chassis.
4. **Delivery** — how the built site reaches a URL the customer can open.

Seams 1 and 2 are **solved by copying idea.uk**. Seam 4 is **solved by copying
relojistas/idea.uk's pull-sync**. Only seam 3 is new, and §9 defers it.

---

## 4. The trust boundary — and the dedicated cluster

> **[ARGUED — this is the part I most want challenged, and it changes the cost of
> the whole project.]**

The owner's instinct is right about the risk and, I think, one step ahead of
itself about the remedy. Two separate things are being asked of a dedicated
cluster, and only one of them actually needs one.

### 4.1 What "don't let my cluster get hacked" really requires

The property wanted is: *nothing reachable from the public internet can reach the
production chassis or its database.* A second cluster is one way to get that. The
estate already has a cheaper one that the owner himself chose three days ago.

The island (tools-api, Mythic Beasts VDS) is built on exactly this rule — *"the
production cluster appears NOWHERE in this path"*, *"NOTHING on this box holds any
production credential"* — and when the VM-estate plan asked whether framework
control should invert that, the owner ruled: **pull-only, outbound from the box**
(`vm_estate/PLAN_2026-07-25…` §5.1, 2026-07-25).

Mirror it and the answer here falls out:

```
  PUBLIC                          |            PRIVATE
  webdesign.uk VM                 |            production cluster
  ├── minimal page + domain box   |            ├── chassis, 14 services
  ├── questionnaire               |            ├── clients_db
  ├── Stripe checkout + webhook   |            └── scheduler
  └── orders, on the box          |
             ▲                    |                     │
             └────────────────────┴─── outbound poll ───┘
              cluster fetches paid orders; box never dials in
```

The box holds **no cluster credential** and has **no inbound path**. If it is
fully compromised, the attacker has the box, the orders on it, and a token that
lets them *answer* a poll — they cannot reach `clients_db`, cannot publish to
Kafka, cannot spawn an agent. That is the security property, and it costs one
outbound HTTP call, not a second Kubernetes cluster.

### 4.2 The argument FOR a dedicated cluster that I think is the real one

There is a genuine isolation argument here and it is **not** "hacked". It is this:

**A build for a customer consumes attacker-supplied content.** We scrape a domain
the customer typed, feed that HTML into LLM prompts, and write generated
components into `clients_db` — the same database serving 14 live sites. Prompt
injection in a scraped page is, in that arrangement, an input to agents that hold
write access to production content. That is a real blast-radius argument, and it
is much stronger than the network one because a firewall does not touch it.

If the owner wants isolation, **this is the reason to state** — in the plan, in
the concept register, and in whatever review this goes through. It also changes
what the isolation has to be: the thing that must be separated is **the database
and the agents that write to it**, not the ingress.

Which opens cheaper options than a cluster, in ascending cost:

| option | isolates | cost |
|---|---|---|
| (a) Separate `network_id` for customer sites | rows, by convention | ~zero — the column exists on `sites` today. **Strength [UNVERIFIED] — needs a read of what actually enforces it** |
| (b) Separate database, same cluster | the data | one Postgres statefulset, one connection string, migration duplication |
| (c) Separate namespace + database | data + pod-level | above, plus a second fleet deploy |
| (d) Dedicated cluster | everything | second of everything: k8s, Kafka, Postgres, images, migrations, monitoring, kubeconfig |

**Recommendation:** ship the boundary in §4.1 on day one (it is nearly free and
needed under every option), and make the isolation choice at the point where the
first *paid build against scraped third-party content* is about to run — which
§9 puts at P3, deliberately after money is already coming in. Do not build (d)
before P1, because P1 through P3 do not need it and it would delay the only thing
that tells us whether any of this sells.

**If the owner prefers (d) regardless, it does not invalidate anything below** —
the pull design points the poller at whichever cluster exists, so the phases are
unchanged and the second cluster becomes a P6 swap rather than a P0 prerequisite.
That is the main reason to design it this way round.

### 4.3 One asset nobody has documented

The cluster runs a **`wireguard` deployment** (`linuxserver/wireguard:latest`,
`deployments/kustomize/services/wireguard/`). It appears in no docs024 workstream.
**[UNVERIFIED — purpose unknown to me.]** If it is already a working private
transport between the cluster and the VM estate, it may be a better answer than
polling over public HTTPS for the *return* leg. Worth one session's read before
P4, not before.

---

## 5. The product — the chat box, and the one change I want to argue for

### 5.1 Make the "chat box" a stepped form, not a chat

[ARGUED.] A real chat box implies an LLM conversation: unbounded input, unbounded
turns, prompt injection, cost per message, no natural end. For the job it actually
does — *capture one domain name* — a single field and a button is better UX and a
fraction of the risk surface.

The *feel* the owner wants (one question at a time, conversational, not a wall of
form fields) is achievable with a stepped form that asks one thing per screen.
**Recommendation: conversational-styled stepped form for v1.** idea.uk already
proves the shape: one embedded `page.html` compiled into the binary, 46KB, no
framework, no build step.

### 5.2 What "create a website from a domain" actually means

The domain splits the product in two, and they are not equally safe:

- **The domain has a live site.** We scrape it, and everything we generate can be
  *grounded in what we read*. Strong demo. We already have the machinery:
  `web-scrape-adapter`, `analyser-adapter`, `batch_webscrape_action`, and
  `cmd/webdesignport` ported 97 pages of two whole sites this way.
- **The domain is parked or unregistered.** We have a string. Anything we generate
  about the business is invention. This is the weak, dangerous half.

**Design consequence:** the free artefact is generated only from what we read, and
where we read nothing we say so rather than filling it in.

### 5.3 The questionnaire must be required, not optional

> **[ARGUED — this is a change to the owner's flow, and the evidence for it is
> ours.]**

The owner has the questionnaire as "an optional better route". I think it has to
be the **gate on the paid build**, for a reason this repo has already paid for
twice:

- `bugs_open/063`: the hallucinated-email check **fails open** when a site has no
  contact email — *a fabricated address reached production for hours* that way.
- `bugs_open/043`: an entire workstream about generated copy inventing
  quantitative claims. This project has shipped invented statistics **twice**.
- The oufe seed's own preamble bans five specific numbers because they were model
  recollections, and notes that the claims scanner is **effectively inert** on
  that site's vocabulary.

A site we generate for *a real named business* that invents its telephone number,
its address, its accreditations or its years in business is not a defect — it is
that business's liability, published under their name, sold to them by us.

So:

- **Free teaser from a domain alone: fine.** It is visibly ours, visibly a demo,
  built only from what we read, and it leaves contact details blank **by design**
  rather than by accident.
- **Paid build: no questionnaire, no build.** The questionnaire's real job is not
  lead capture — it is the **fabrication control**. Its minimum fields are exactly
  the ones the platform otherwise invents: legal business name, contact email,
  telephone, address, services, and an explicit *"claims you are happy for us to
  make"* box.

This is not a smaller product. It is the same funnel with the honest gate in it,
and it gives the page something true and unusual to say: *we will not write a word
about your business that you have not told us or we have not read on your site.*

### 5.4 What the free teaser should contain

Proposal, ordered by cost:

1. **An honest read of the existing site** — near-free, and we have real
   machinery: browser-runner checks (`no_horizontal_overflow`, `no_console_errors`,
   `page_status_ok`), and **`scripts/render_audit.py`** (register VIZ-010) — a
   headless-Chromium render witness that composites alpha through transparent
   ancestors to find the effective background, applies 4.5:1 / 3.0:1, and reports
   images that failed to load. On fundamentallyai.com it found **101 AA failures
   across 5 pages in about two minutes**, on a site where every page said
   `deployed` and ~50 discovery checks had objected to nothing. That is the
   teaser, almost verbatim. It is **run by hand only — wired into nothing**, so
   P2 is largely a matter of calling it.
   This is T2 from the buying-design plan, and it is the cheapest strong move.

   > **CORRECTED before this file was two hours old.** This paragraph first cited
   > `cmd/contrastscan`. That tool **does not exist** — it was built and deleted
   > the same day (2026-07-28) as a duplicate of the Python one above, and the
   > register records why: *"the prior-art grep was `--include=*.go` and the prior
   > art is Python."* Caught by `pattern-check.py`'s `new-capability-surface` on
   > the commit, not by me. The cheap check that would have caught it is `ls cmd/`.
2. **A proposed site plan** — pages and sections, a palette, a direction. A
   handful of model calls.
3. **One rendered section** in the proposed style — the conversion lever, and the
   biggest embarrassment risk if it comes out badly.

**Cost of 1–3: [UNMEASURED].** Do not guess it. `llm_call_log` has 45k rows and
makes this a query, not an estimate — measure one real run before setting a price
or a free-tier cap (§9 P0).

---

## 6. Delivery — the preview subdomain

The customer's built site must land somewhere they can open. Two routes:

- **(i) B2 + wildcard DNS**, the path the 12 static sites use. Requires working
  out the domain→bucket mapping and touching the production static path.
  **[UNVERIFIED — I have not read how a domain maps to a bucket prefix.]**
- **(ii) The same VM serves it.** nginx with a regex `server_name`
  (`~^(?<slug>.+)\.preview\.webdesign\.uk$`), `root /var/www/preview/$slug`, a
  wildcard certificate by DNS-01, and the box pulling `/preview/<slug>/` from the
  sites repo on the systemd timer **that idea.uk and relojistas already run**.

**Recommendation: (ii).** It reuses a proven mechanism, needs no change to the
static path 12 live sites depend on, and keeps the whole customer-facing surface
on one box we can reason about. Note the pages are built with **root-relative
links**, so a path prefix (`/preview/<slug>/`) would break them — the subdomain is
not cosmetic, it is required.

---

## 7. Price, and what we are allowed to say about cost

The pricing input is measurable and should be measured, not recalled. The
canonical discipline is already written: `idea_uk_vm_site/EVIDENCE_2026-07-27_ai_unit_economics.md`
— *the measured figure was a FLOOR covering two of five calls, and presenting it
as a total would have been false*.

Applied here:

- Measure a full build from `llm_call_log` before quoting any cost or margin.
- Any published figure carries its **measurement date**. `claude-sonnet-5` leaves
  its introductory rate on **2026-08-31** — a 50% rise on that half of the bill,
  five weeks out.
- The durable story needs no arithmetic at all, and idea.uk's version is the
  model: *a bespoke site, built end to end by an AI pipeline, with a human
  reviewing it before it goes out.*

---

## 8. Risks that are day-one, not polish

Four, and only the last is exotic:

1. **SSRF.** The product's core interaction is *"type a URL and we will fetch
   it"*. Handed to a scraper or a headless browser unguarded, that fetches
   `169.254.169.254`, `10.x`, `localhost`, and anything else. **Requires:**
   scheme allow-list, public-IP-only resolution checked *after* DNS resolution,
   redirect capping, response size cap. The consolidation programme has
   `platform/httpguard` (item A3) — **check whether it already does this before
   writing a second one.**
2. **The spend faucet.** A public endpoint that makes model calls is free money
   spent by strangers. **Requires:** per-IP rate limit, per-day global ceiling,
   and a cheap-first ordering so the expensive step happens last. tools-api's
   `RateLimitMiddleware` and `InputCapMiddleware` already exist.
3. **PII.** The questionnaire collects email, telephone and address for real
   businesses, on a box the vm-estate table records as having **no backups**.
   That is a data-protection duty with a retention policy attached, and it needs
   deciding before the first form is filled in, not after.
4. **Prompt injection into shared state** — §4.2. The reason isolation is worth
   money.

And one operational rule taken from `bugs_open/083`: **log every request from the
first deploy.** The island ran without a request log and consequently had no
denominator — its 503 rate could not be honestly quoted at all.

---

## 9. Phasing — each phase independently useful, independently stoppable

- **P0 — Decide and measure. No code.** Owner rules §11. Measure one full site
  build's model cost from `llm_call_log`. Read `platform/httpguard` to see whether
  risk 1 is already solved. *Output: a price, a cost, and the isolation ruling.*
- **P1 — The shopfront, with nothing behind it.** webdesign.uk on a VM: the
  minimal page, the domain box, the questionnaire, Stripe in **test mode**, orders
  stored on the box. Nothing builds. **This is the fake-door idea.uk already ran**
  (`idea.uk/idea_uk_fakedoor.html`) and it answers the only question that matters
  — does anyone type a domain in and go through with it.
- **P2 — The free teaser.** §5.4 item 1 first (near-free, real machinery), then 2.
  Item 3 only once the cost from P0 is known. Ship the guards from §8 with it, not
  after.
- **P3 — Manual fulfilment. Real money.** Stripe live. A paid order is picked up
  by a human, who seeds the site row and specs the way oufe was seeded, triggers
  the existing build, reviews it, releases it to
  `<slug>.preview.webdesign.uk`, and sends the link. **This is a shipping business
  with zero new platform code**, and the "next day or so" on the page is honest
  because it is a human's day.
- **P4 — Automate the trigger.** The pull action + a scheduled task. **First
  platform-code change; first council submission; first concept-register entry.**
- **P5 — Automate the seeding.** Brief → `site_specs` aspects. This is where "use
  the framework directly" actually happens, and it is the largest single piece.
- **P6 — Execute the isolation ruling** (§4.2), before volume.
- **P7 — Ownership and handover.** What the customer actually gets: the domain
  pointing at it, the code, who can edit it later. Note this is *pillar 6 of the
  buying-design section* — we would be selling advice on lock-in, so we had better
  be exemplary at it.

The shape of that list is the argument: **P1–P3 are a business, and none of them
touch `platform/`.**

---

## 10. Council and register routing

- **This document cannot go through the council gate.** Scope is `platform/`,
  `internal/`, `pkg/`; docs are refused client-side and never spend credits.
- **P4 and P5 must**, one run per coherent task.
- The intake/trigger seam is a **shared mechanism**, so under the owner's ruling
  of 2026-07-28 it is **architecture-scope** even though it is additive. That
  means: register it in `docs026_concept_register` **in the same commit that ships
  it**, state any real ordering constraint in the commit message, and **measure the
  blast-radius claim before submitting** — 124 was rejected for asking the council
  to do the measuring.
- Expect the isolation decision (§4.2) to need **a human ruling, not a verdict**:
  the architecture seat has fired **0 times ever** while counted in the 16-seat
  roster.

---

## 11. Open for the owner

1. **The trust boundary (§4).** Ship the VM-front + outbound-pull boundary now and
   take the isolation decision at P3 — or stand up the dedicated cluster first?
   My recommendation is the former; the latter is unchanged in cost if deferred.
2. **Isolation option (§4.2)** — (a) `network_id`, (b) separate DB, (c) namespace
   + DB, (d) dedicated cluster. And is the reason the blast radius of scraped
   content, as I have argued, or something else?
3. **The questionnaire as a gate (§5.3)** — required before any paid build, or
   genuinely optional as you first described it?
4. **The chat box (§5.1)** — stepped conversational form, or do you specifically
   want a real LLM chat?
5. **Price**, and whether the teaser is free to everyone or gated on an email.
6. **The preview host (§6)** — `*.preview.webdesign.uk` on the VM, or a subdomain
   of a different domain of yours?
7. **§12 — the "thousand sites" figure.** Needs a decision, and it is the one
   thing here with a deadline attached, because it is already in outward copy.

---

## 12. A figure that needs pinning before it is sold on

The buyer-track positioning rests on a claim of scale: *"we run one of these
systems in production across about a thousand sites"*
(`webdesign_couk/README_where_we_are.md:405`, and again in
`SUMMARY_2026-07-28_what_the_news_feed_taught_us.md:21`).

**Measured today: the `sites` table holds 32 rows — 14 deployed, 17 `pool-*`
shells, 1 `system.internal`.**

Tracing it back, ~1,000 appears in the architecture threads as a **target**
("reach 1,000 sites", "at 1,600 domains this couples site count to binary size") —
a scale premise for arguing about per-site Go actions, which is a legitimate use.
It has since drifted into outward-facing prose as a **present-tense claim**.

It may well be true of a different noun — **domains owned** is plausibly in that
range, and is a different sentence. **The figure needs pinning to its noun before
it is published**, and webdesign.uk makes that urgent rather than academic:
this is the exact class of claim — a number that feels like common knowledge — that
this project has shipped wrongly twice, and we are about to build a business whose
entire pitch is *we are the ones who tell you the truth about AI web builds*. A
buyer who checks, and this audience checks, would find it.

**Not a criticism of the positioning, which I think is right.** The honest version
is at least as strong: *fourteen production sites, built and run by the system, on
the record including where it went wrong.*

---

## Related

- `webdesign_couk/PLAN_2026-07-27b_buying_design.md` §8 — where this was recorded.
- `idea_uk_vm_site/` — the precedent that has taken money. `PLAN`, and
  `EVIDENCE_2026-07-27_ai_unit_economics.md` for the cost discipline.
- `vm_estate/PLAN_2026-07-25_framework_controlled_vm_estate.md` — the estate this
  box joins, and the pull-only ruling.
- `gauntlet_dead_cta/infra/island/RUNBOOK_island.md` — the trust boundary as built.
- `oufe/SEED_2026-07-25_oufe_site_and_specs.sql` — what "seed a new site" means today.
