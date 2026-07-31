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
| **Paid funnel on a VM** | **Built, and a real Stripe transaction has run end to end.** idea.uk: stdlib-only Go binary, Stripe behind a `Provider` interface, webhook = source of truth, HMAC-verified, `FakeProvider` for end-to-end tests, price passed **per call** not held on the provider. **See §2a — the buyer was the owner.** | `idea.uk/golang_files/billing.go`, `store.go`, `service.go` |
| **Public HTTP tool endpoint** | **Built, one worked example.** gin + own Postgres, CORS → rate-limit → input-cap middleware, request logging. Zero Kafka, zero chassis coupling. | `internal/tools-api/api/server.go:32-58` |
| **Static site build + deploy** | **Live for 12 sites.** chassis → git-adapter → repo `sites` → GitHub Action → B2. Per-site override to `vm-sites` for VM-hosted sites. | `platform/orchestration/actions/git_deployer_actions.go`, `git_repo_resolution_test.go` |
| **VM-hosted site with its own backend** | **Live twice.** idea.uk and relojistas.com: nginx + certbot, site-engine under systemd, box pulls its own folder from `vm-sites`. | `sites.github_repo='vm-sites'` for both |
| **A way to call the chassis from outside** | **Does not exist for untrusted callers.** core-manager has `POST /api/v1/admin/pipelines/:name/trigger` — behind `AuthMiddleware` + `AdminOnly()`, served inside the cluster behind api-gateway. tools-api holds no Kafka client at all. | `internal/core-manager/api/server.go:108-227`; `cmd/tools-api/main.go` |
| **Automated "domain in → site out"** | **Does not exist.** New sites are created by hand-authored SQL seeding the `sites` row and its `site_specs` aspects, then trigger scripts publish to Kafka. | `oufe/SEED_2026-07-25_oufe_site_and_specs.sql`; oufe has 12 aspects seeded on day one |
| **Per-call model cost** | **Measurable.** `llm_call_log` holds 45,205 rows, 2026-03-25 → 2026-07-28. | live query, 2026-07-28 |
| **Fleet size** | **32 site rows, 14 deployed, 17 `pool-*` shells.** §12 — the owner has since ruled the "thousand sites" claim stands as forward-looking. | live query, 2026-07-28 |

### 2a. The correction that should change how this is read

> **CORRECTED 2026-07-28, hours after this file was written.** The first version
> of §2 said idea.uk *"has taken real money"* and *"survived a real sale"*, citing
> the 27 July first sale. **The buyer was the owner.** Genuine external buyers:
> **still zero** — the one order that looked external was a test, and the thread
> that inferred otherwise recorded its own correction
> (`idea_uk_vm_site/HANDOFF_RESUME…:17`, `RUNNING_NOTES…:2764`). Caught by the
> workstream memory index prompting a check of the source docs, not by me
> re-reading my own draft.

What survives and what does not:

- **The payment code is proven.** A real Stripe transaction ran end to end —
  checkout, webhook, signature verification, order state, delivery. That is a
  claim about *the code*, and the buyer's identity does not touch it. Copying
  `billing.go` remains the right move.
- **Nothing here evidences demand.** It never did; the phrasing implied it.

**And this is the most useful fact in the whole document, because it points the
other way from the obvious plan.** idea.uk is complete, verified, working, live,
and has sold nothing to a stranger. Its own workstream states the lesson in the
imperative: *a product can be complete, verified and working and still never have
completed a transaction — ask that of any site declared done.*

So the argument for P1 as a demand test is not tidiness, it is the estate's most
recent expensive lesson. **Build the shop, point real traffic at it, and find out
whether anyone types a domain in — before building the engine behind it.** The
failure mode to avoid is not a broken build pipeline; it is a beautiful one nobody
buys.

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

> **DECIDED (owner, 2026-07-28): ship the boundary now, take the isolation
> decision at P3.** So §4.1's VM-front + outbound-pull arrangement is a design
> constraint from here, not an option, and the choice between (a)–(d) is deferred
> to the point where the first paid build against scraped third-party content is
> about to run. Deferring costs nothing because the poller points at whichever
> cluster exists — that property is the reason to design it this way round, and it
> is now load-bearing. **Do not erode it**: any later shortcut that has the box
> dialling into the cluster, or holding a cluster credential, gives back the whole
> security argument.

Two things follow that should not be lost between now and P3:

- **The isolation question must actually be asked at P3**, not silently skipped
  because money is flowing by then. The trigger is specific: *the first paid build
  that scrapes a domain we do not own*. Whoever runs that build owes the decision.
- **§4.2 is the reason to state when it is asked** — blast radius of attacker-
  supplied content reaching agents with write access to the DB serving 14 live
  sites. Not "hacked". The wrong reason picks the wrong remedy.

### 4.3 One asset nobody has documented

The cluster runs a **`wireguard` deployment** (`linuxserver/wireguard:latest`,
`deployments/kustomize/services/wireguard/`). It appears in no docs024 workstream.
**[UNVERIFIED — purpose unknown to me.]** If it is already a working private
transport between the cluster and the VM estate, it may be a better answer than
polling over public HTTPS for the *return* leg. Worth one session's read before
P4, not before.

---

## 5. The product — the chat box, and the one change I want to argue for

### 5.1 The chat box — OWNER RULING: a real LLM chat

> **DECIDED (owner, 2026-07-28): a real LLM chat, not a stepped form.**
> I argued for a stepped conversational form — same feel, a fraction of the
> exposure, since the job is to capture one domain name. **Overruled, and on
> reflection the ruling buys something my version could not:** see §5.3. A real
> conversation can *conduct the briefing*, which is exactly the material the
> optional questionnaire would otherwise have to collect from a form nobody fills
> in. The two rulings work together better than either does alone.

The argument I made, kept because it is the risk register for what now has to be
built: a real chat is unbounded input, unbounded turns, cost per message from
strangers, prompt-injection surface, and no natural end.

**So the ruling moves work earlier, and that is the main planning consequence.**
A fake door with a stepped form costs nothing to run and could ship bare. A fake
door with a real LLM chat **spends money on every visitor, including hostile
ones**, so the controls are no longer P2 polish — they ship with P1 or P1 does not
go public:

| control | why | reuse |
|---|---|---|
| Per-IP rate limit | the faucet | `middleware.RateLimitMiddleware` (tools-api) — **but see the landmine below; do not reuse its keying** |
| Request body cap | oversize prompts | `middleware.InputCapMiddleware` |
| **Turn cap per session** | a chat has no natural end; a form does | new, trivial |
| **Per-day global spend ceiling** | the only control that bounds total loss | new; `llm_call_log` makes it measurable |
| Cheap model for the chat itself | the chat is intake, not the product | the build is where the money should go |
| Request log from deploy #1 | `bugs_open/083`: the island had no denominator | `gin.Logger()` |
| Prompt-injection containment | the transcript becomes build input (§4.2) | treat the transcript as **data, never instructions** |

That last row is the one to think hardest about, and it is not a middleware
setting. The chat transcript flows into the brief, which flows into the build. A
visitor who types *"ignore your instructions and…"* is writing into a document
that later agents read. **The transcript must enter the build as quoted customer
statements in a named field, never as free prose spliced into a prompt.**

> **LANDMINE — a "per-IP" limiter behind Cloudflare is usually one global bucket.**
> Added 2026-07-29 from `bugs_open/139`, which is the same architecture we are
> about to build (Caddy/nginx + Cloudflare in front of a Go service). The island's
> per-IP limiter keyed on a **constant**: `client_ip_hash` was
> `sha256("172.18.0.1")` — the docker gateway — in **83 of 83 rows**, so every
> visitor on earth shared one bucket and the stored column had never
> distinguished anybody. Two things make this expensive to find:
> - **The real address is in `CF-Connecting-IP` only.** Caddy overwrites
>   `X-Forwarded-For`; Cloudflare strips `X-Real-IP`. `platform/httpguard`'s
>   rightmost-XFF fallback lands on the same constant — **it reads as a fix and
>   is not one.**
> - **One test machine cannot tell a constant from a working key.** Your own
>   requests get one value either way, so the limiter appears to work. The
>   discriminating check is `count(DISTINCT <ip key>) > 1` **from two different
>   networks**.
>
> This bites the spend ceiling directly: a global bucket means one visitor can
> exhaust the day's budget for everyone, and the per-IP control we are relying on
> to bound §8's faucet would be decorative.

idea.uk still supplies the page shape: one embedded `page.html` compiled into the
binary, 46KB, no framework, no build step. The chat is a panel on it, not a
rewrite of it.

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

### 5.3 The questionnaire — OWNER RULING: stays optional

> **DECIDED (owner, 2026-07-28): optional, as originally described.**
> I argued it should gate the paid build. Overruled. **The argument is kept below
> in full, because the risk it describes does not go away when the gate does — it
> moves, and §5.3a is where it moves to.** Recording the reasoning rather than
> deleting it is what lets a later thread re-open this with the evidence intact
> rather than re-deriving it.

The argument, as put:

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

So the questionnaire's real job was never lead capture — it is the **fabrication
control**, and its minimum fields are exactly the ones the platform otherwise
invents: legal business name, contact email, telephone, address, services, and an
explicit *"claims you are happy for us to make"* box.

### 5.3a Where the control moves, now that it is not a gate

**A gate is one way to stop a fabricated telephone number reaching a customer's
live site. It is not the only way, and it is not the strongest one.** Ranked by
what makes the bad state *unrepresentable* rather than merely discouraged:

1. **Emit no contact block at all unless the details were supplied.** Not a
   placeholder that could ship — **absent**. This is the structural fix, it is
   cheap, and it removes the failure entirely rather than reducing its odds. Note
   `bugs_open/063` is precisely a *check* failing open; a field that is never
   generated cannot fail open.
2. **Seed `evidence_base` before the first page is written.** The platform already
   has this mechanism and the oufe seed's own preamble explains it: *the entire
   claims layer is gated on the PRESENCE of this aspect — `loadEvidenceBase`
   returns nil and every lane silently no-ops*
   (`validate_page_content.go:727-746`). For a customer site the evidence base is
   assembled from the two sources we actually have — **what they told the chat**
   and **what we read on their existing site** — each attributed. That turns "do
   not invent" from an instruction into a data structure.
3. **The human pre-release review**, which P3 has anyway.

**The conditional that follows, and it is the one to carry forward:** the
questionnaire can stay optional *while a human reviews every site before release*.
P5 is the phase that automates release. **When release stops passing a human, this
decision must be re-opened** — either the gate returns, or controls 1 and 2 must
be demonstrably doing the work on their own. Flagged here so the phase that
removes the backstop is the phase that notices.

**And the ruling on §5.1 helps more than it looks.** A real LLM chat can ask for
the telephone number, the address and the services *conversationally*, in the
moment someone is already engaged — which is a far better collection mechanism
than an optional form after the fact. The optional questionnaire is much less
likely to be the empty path than it would have been behind a stepped form.

The page still gets something true and unusual to say, and it survives the ruling
intact: *we will not write a word about your business that you have not told us or
we have not read on your site.*

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

> **DECIDED (owner, 2026-07-29): a different, shorter domain — to be supplied.**
> So the previews do **not** live under `webdesign.uk`. The mechanism above is
> unchanged and the domain name is a placeholder wherever it appears in this
> workstream; substitute it once the owner names it. **Not blocking** — nothing
> before P3 needs the name, and the wildcard-cert and vhost work is identical
> whatever it turns out to be.
>
> Two properties the chosen domain must have, because §7a now leans on this host:
> it carries the **guarantee mechanism** (a refund is "the preview comes down"),
> and a customer will read the URL before they trust it — so it should look like
> somewhere a real deliverable lives, and it must be a zone we control DNS for
> (DNS-01 wildcard issuance).

---

## 7. Price, and what we are allowed to say about cost

The pricing input is measurable and should be measured, not recalled. idea.uk has
now walked this twice and the second lap is the instructive one:

1. `EVIDENCE_2026-07-27_ai_unit_economics.md` measured **$0.641** and said plainly
   it was a **floor** covering two of five calls, not a total.
2. **Once all five calls logged, the honest answer turned out to be a *range*, not
   a number: `~$1.20–$1.45 depending on length`.** Output tokens are **~92% of
   spend**, so cost tracks the **length of the artefact** — 26,264 characters
   against the previous run's 13,227 roughly doubled it
   (`idea_uk_vm_site/HANDOFF_RESUME…:42-43`, `RUNNING_NOTES…:3106`).

~~**That second finding transfers directly and is the one to design around.** A
website is a far more variable artefact than a report: a 5-page brochure site and
a 40-page site are not the same product at the same cost. **So the price cannot be
a single number unless the deliverable is capped** — which argues for a fixed page
count in the offer, or tiers, rather than "we'll build you a website".~~

> **DECIDED (owner, 2026-07-29): full sites, at a high, quality-based price —
> the cap-or-tiers recommendation above is SUPERSEDED** (kept struck-through, per
> the working-docs rules). And the ruling dissolves the concern it answered: at a
> high price point, model-cost variance of a few dollars is margin noise. The
> length finding stays true — it just stops being a *pricing* constraint and
> becomes a *margin* line-item. **The genuinely scarce, genuinely variable cost is
> the owner's attention** (bugfixes, spec changes), and §7a is where that gets
> priced.

Model prices as of **2026-07-29** (from the claude-api skill, not memory — the
fleet-relevant rows):

| model | input $/MTok | output $/MTok | note |
|---|---:|---:|---|
| claude-fable-5 | 10.00 | 50.00 | **2× Opus 5; the build model per §7b** |
| claude-opus-5 | 5.00 | 25.00 | |
| claude-sonnet-5 | 2.00 intro | 10.00 intro | **intro ends 2026-08-31 → $3/$15** |
| claude-haiku-4-5 | 1.00 | 5.00 | the chat/intake tier |

Applied here:

- Measure a full build from `llm_call_log` before quoting any cost or margin —
  the price is quality-based, but the **margin claim** still needs the
  measurement, and P0 owns it (now on Fable 5 per §7b).
- Any published figure carries its **measurement date**. The fleet's dominant
  model is Sonnet 5 (**1,468 of the last 4 days' 1,900 calls**, measured
  2026-07-29), so the current cost baseline rises ~50% on that share when the
  introductory rate ends **2026-08-31**.

## 7a. The offer, ruled and sharpened (owner, 2026-07-29)

> **DECIDED (owner, 2026-07-29):** full sites · high quality-based price ·
> full money-back guarantee, **acceptance-gated on the preview** · corrections
> carry a fee, **boundary: customer changes paid, our defects free**. The
> mechanics below are how the pieces were sharpened in discussion; the rulings
> are the owner's.

**Productised, not bespoke.** The owner's framing — *"they are buying what we are
offering rather than something they get elsewhere"* — is the idea.uk shape: the
page sells what our system builds, reviewed by a human, at a stated price. Not
"we'll build whatever you want". This makes the page copy honest and simple, and
it is what lets a fee-per-change model exist at all: changes are priced because
the deliverable was defined.

**The guarantee and the fee model are one mechanism, hinged on acceptance:**

```
   pay → build → PREVIEW on our subdomain → customer reviews
                        │
        ┌───────────────┴───────────────┐
   refund (full)                    ACCEPT
   preview comes down          handover; guarantee ends
   nothing delivered           corrections now exist:
                               our defects → free
                               their changes → paid
```

- **Before acceptance:** the guarantee is the only instrument. No paid
  corrections exist yet — revisions during preview are part of getting to
  acceptance or the money comes back. Refund = the preview comes down; the
  customer keeps nothing.
- **After acceptance:** the guarantee has ended and the fee model starts. A
  defect we caused — broken link, rendering bug, an invented fact — is fixed
  free, indefinitely. A change they want — different copy, new colour, another
  page — is paid work.
- **Why defects-free is load-bearing, not generosity:** the whole positioning
  (webdesign.co.uk's buying-design section, this product's pitch) is *we tell
  the truth about AI builds*. Charging a customer to fix our own broken link is
  the cheapest attack a competitor could quote. It also makes §5.3a's
  fabrication controls **directly revenue-protective**: every invented detail
  that ships is a free defect fix we owe — the no-invented-contact-block rule
  now defends margin, not just reputation.
- **The preview subdomain is therefore not just delivery — it is the guarantee
  mechanism.** §6's choice of host carries this weight.

**Terms must exist in writing before the first sale.** The guarantee, the
acceptance event, and the fee boundary are all contractual claims; idea.uk
already walked this (`idea.uk/LIABILITY_AND_TERMS.md` and its `terms_preview.html`
are the precedent). One flag per the legal rail: whether the buyer is a business
or a consumer changes what cancellation rights attach **[UNVERIFIED — needs a
primary source before any terms page ships; do not write it from memory]**.

## 7b. Model roles — builds on Fable 5 (owner direction, 2026-07-29)

> **DECIDED (owner, 2026-07-29): the paid builds are planned on `claude-fable-5`.**
> Checked against the skill the same day: Fable's stated sweet spot is exactly
> this work — the most demanding long-horizon agentic tasks, first-shot
> implementation of well-specified systems, self-verification. At $10/$50 per
> MTok it is 2× Opus 5 and ~5× the fleet's current Sonnet 5 intro rate — small
> against a high-priced site, and the measurement (P0) will say precisely how
> small.

| role | model | why |
|---|---|---|
| Chat/intake (P1) | cheap + fast — `claude-haiku-4-5` tier | the chat is intake, not the product; it faces strangers and its cost is bounded by §5.1's controls |
| **Paid build** | **`claude-fable-5`** | the product; quality-based price funds the premium model |
| Pre-release review | human (P3) | the honesty gate; automation of it is P5's question |

**The fleet is NOT on Fable — measured, not assumed** (2026-07-29,
`llm_call_log` last 4 days): claude-sonnet-5 1,468 · claude-sonnet-4-6 311 ·
mistral-small3.1 85 · gemini-pro-latest 33 · **claude-fable-5 0**. So "builds on
Fable 5" is a change to make deliberately, and P0 gains three items **in this
order** — note DB model config is live immediately (CLAUDE.md), so (i) and (ii)
come before any lane is pointed at Fable:

1. ~~Verify org data retention ≥ 30 days.~~ **DONE 2026-07-31 — PASSES.** Not
   greppable from the tree, so verified the only way available: a live probe
   call to `claude-fable-5` from inside the cluster (`agent-chassis` pod, using
   the org's own `ANTHROPIC_API_KEY`, via `wget` — no client on the pod).
   **`HTTP 200`, not the retention-configured `400 invalid_request_error`** —
   the org's retention already meets the requirement. Full request/response and
   the exact numbers are in §7c below.
2. ~~Grep the chassis LLM call layer for temperature/top_p/top_k/budget_tokens/
   thinking config.~~ **DONE 2026-07-29 — CLEAN.** Read `platform/aiservice/
   anthropic.go:89-113`: `temperature` is **already unconditionally dropped** —
   never sent to Anthropic at all (a standing guard, predating this plan, for
   the same 400 on Opus 4.7+). `budget_tokens` → `thinking:{enabled,...}` is
   sent **only if** `ai_service.budget_tokens` is set in an agent's config; no
   `top_p`/`top_k` are sent by this client at all (the `top_k` hit in
   `rag_actions.go` is retrieval-k, an unrelated concept, confirmed by reading
   the call site). Checked which live agents set `budget_tokens`:
   `SELECT type FROM agent_definitions WHERE default_config::text LIKE
   '%budget_tokens%'` → **`council-gate`, `fix-proposer` only** — the D9/D10
   mirror pair (§ landmines below), neither a candidate for this build lane.
   **Consequence: pointing a fresh agent (no `budget_tokens` in its own config)
   at `claude-fable-5` is safe at the code layer today** — nothing to change in
   `anthropic.go`, and the omitted-`thinking` behaviour (adaptive, always on)
   is exactly what Fable expects. Confirmed against the `claude-api` skill the
   same day: omit or `{type:"adaptive"}` → runs; `{type:"disabled"}` → 400
   (Fable-specific — Opus 4.8/4.7 accept `disabled`); `{enabled,budget_tokens}`
   → 400; any non-default `temperature`/`top_p`/`top_k` → 400.
3. **Measure one real Fable-5 build** end to end from `llm_call_log` — that
   number, dated, is the pricing input §7 waits on. Still open; (1) and (2) no
   longer block it.

Two more Fable properties the build lane must absorb: **minutes-long turns are
normal** (timeouts and progress handling, not a hung-lane diagnosis), and
**`stop_reason: "refusal"` must be handled** before reading content — a safety
classifier can decline mid-build and the lane needs to treat that as a state,
not a crash.
- The durable story needs no arithmetic at all, and idea.uk's version is the
  model: *a bespoke site, built end to end by an AI pipeline, with a human
  reviewing it before it goes out.*

## 7c. First live Fable-5 measurement — a FLOOR, not a build cost

> **Read the heading before the numbers.** idea.uk's own workstream has already
> been burned once by exactly this: a real, dated, correctly-measured figure
> ($0.641) got repeated elsewhere as *"the cost of a report"* when it covered 2
> of 5 calls. **This is smaller still — one short call, not one page, not one
> site.** Treat it as a unit-price confirmation, not a pricing input.

**What was done, 2026-07-31.** No Anthropic client exists in this shell and no
chassis lane is wired to Fable yet (P4/P5), so the only honest way to answer
"does this org's retention pass, and what does Fable actually cost" was a live
probe: `kubectl exec` into a running `agent-chassis` pod (which already carries
`ANTHROPIC_API_KEY` as it must, to do its job), and `wget` — the only HTTP
client the pod's minimal image has — a single request straight to
`api.anthropic.com/v1/messages`, `model: "claude-fable-5"`, no `thinking` key
(omit → adaptive), asking for a short piece of representative website copy (an
"About Us" paragraph, 120–150 words). Request and response files were removed
from the pod's `/tmp` immediately after.

**Result: `HTTP 200`.** Retention passes (§7b item 1, now closed). Measured:

| | tokens | rate | cost |
|---|---:|---:|---:|
| input | 69 | $10/MTok | $0.000690 |
| output (incl. 25 thinking) | 282 | $50/MTok | $0.014100 |
| **total, this one call** | | | **$0.01479** |

Also confirmed live, not just from the skill doc: `stop_reason: "end_turn"`,
`stop_details: null` (matches the documented *null-unless-refusal* shape);
`thinking` block present with empty `"thinking":""` text (matches the
documented default `display:"omitted"` — the reasoning happened and was billed,
the text is withheld); the exact $10/$50 per-MTok rate the skill quoted, live
against a real invoice line rather than a cached table.

**What this is NOT.** One short paragraph is not one page, and one page is not
one site. A real build (§7 P5) runs classification, planning, several pages of
content, imagery direction, and a verification pass — each its own call, each
with its own input/output split, several likely far larger than 282 output
tokens. **Do not multiply this number by anything to get a build price.**

**What it does settle:** the retention blocker (§7b item 1) is closed, the
$50/MTok output rate — the dominant cost driver, same as idea.uk's finding — is
confirmed live rather than assumed, and the request shape (no `thinking` key,
Fable runs adaptive and returns an empty-text thinking block by default) is
proven against the real API rather than the skill's description of it. **The
one remaining P0 item is unchanged: a full build, wired through P4/P5, measured
end to end from `llm_call_log`.**

---

## 8. Risks that are day-one, not polish

Four, and only the last is exotic:

1. ~~SSRF — checked 2026-07-29, and it is NOT covered.~~ **BUILT AND SHIPPED
   2026-07-31 — `platform/fetchguard`** (register `DBI-025`, `bugs_open/159`).
   A sibling package to `httpguard`, covering the mirror direction. Its
   `NewClient(cfg)` returns an `*http.Client` whose `Transport.DialContext`
   resolves the target itself and refuses to dial any address that is not
   publicly routable — checked at the *specific resolved address* about to be
   dialed, never a pre-resolved hostname, which is what closes the DNS-rebinding
   TOCTOU gap a naive "check then connect" design leaves open. Redirects re-dial
   through the same transport, so a redirect to a private target is caught by
   the identical check automatically. `LimitedRead` caps response size and
   reports truncation explicitly.
   **Already wired into a live, in-production call site** —
   `internal/adapters/webscrape/adapter.go`'s `downloadImage`, which turned out
   to be a real, shipped SSRF hole: it fetches image URLs taken straight from a
   *scraped page's own content*, attacker-influenced by construction, with no
   checks at all. Filed as `bugs_open/159` because it was already live across
   10 measured agent types, not something this product would have introduced.
   **What this means for P1/P2:** the guard exists and is proven (own test
   suite: refuses a real loopback listener, a redirect to one, and a literal
   metadata-shaped IP; passes a real public-shaped target where the
   environment allows testing it). Any new fetch this product adds — the
   domain-intake teaser reading a customer's site, in particular — should
   construct its client via `fetchguard.NewClient`, not a bare `&http.Client{}`.
   **Explicitly NOT covered**: a headless browser navigating a URL (Playwright,
   as `browser-runner-adapter` uses) is a different fetch surface this
   Go-transport guard cannot see — flagged in the register entry, not silently
   assumed handled.
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

- **P0 — DONE, except the full build measurement.** Owner rules §11 (round two
  done 2026-07-29). **All three Fable-5 checks closed:** call-layer grep clean
  (§7b), data retention verified via a live probe — `HTTP 200`, not the
  retention 400 — and unit pricing confirmed live at $0.01479 for one short
  call (§7c, explicitly marked a floor, not a build cost). **SSRF built and
  shipped** — `platform/fetchguard` (§8 item 1), wired into a real production
  hole it found along the way (`bugs_open/159`), submitted to council
  (`Council-Submitted: 41bbaca4-25f1-45da-a2c1-28a246a5d07a`). Only remaining
  P0 item: **one real build, wired through P4/P5, measured end to end** from
  `llm_call_log` — not answerable until those phases exist. *Everything else
  P0 was meant to produce — cost floor, retention, price direction, an SSRF
  guard reusable fleet-wide — is in hand.*
- **P1 — The shopfront, with nothing behind it.** webdesign.uk on a VM: the
  minimal page, the **LLM chat**, the questionnaire, Stripe in **test mode**,
  orders stored on the box. Nothing builds. **This is the fake-door idea.uk
  already ran** (`idea.uk/idea_uk_fakedoor.html`) and it answers the only question
  that matters — does anyone type a domain in and go through with it.
  > **Resized by the §5.1 ruling.** A fake door with a stepped form costs nothing
  > to run and could have gone up bare. **A fake door with a real LLM chat spends
  > money on every visitor**, so §5.1's control table — per-IP limit, turn cap,
  > per-day spend ceiling, request log, transcript-as-data — is **part of P1**,
  > not P2 polish. P1 is therefore a bigger phase than it first looked, and it is
  > the phase that must not be rushed to "just get it up".
- **P2 — The free teaser.** §5.4 item 1 first (near-free, real machinery), then 2.
  Item 3 only once the cost from P0 is known. Ship the remaining guards from §8
  with it — SSRF above all, since P2 is where a user-supplied URL first reaches a
  fetcher.
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

## 11. Decisions taken, and what is still open

**Ruled by the owner, 2026-07-28** — all three recorded where they bite, not only
here:

1. ~~**The trust boundary (§4).**~~ **DECIDED: boundary now, isolation at P3.**
   §4.1 is a design constraint from here. The trigger for the deferred decision is
   the first paid build that scrapes a domain we do not own.
2. ~~**The questionnaire as a gate (§5.3).**~~ **DECIDED: stays optional.** My
   argument is kept in §5.3 and the control it was carrying moves to §5.3a — emit
   no contact block unless supplied; seed `evidence_base` from chat + scrape.
   **Re-open at P5**, which is when automated release removes the human backstop.
3. ~~**The chat box (§5.1).**~~ **DECIDED: a real LLM chat.** Resizes P1: the
   spend and abuse controls ship *with* the fake door.

**Ruled by the owner, 2026-07-29** — the second round:

4. ~~**The offer shape.**~~ **DECIDED: full sites, high quality-based price, full
   money-back guarantee, corrections carry a fee.** Mechanics in §7a: the
   guarantee is **acceptance-gated on the preview**, and the fee boundary is
   **customer changes paid, our defects free**. Supersedes §7's cap-or-tiers
   recommendation.
5. ~~**The preview host (§6).**~~ **DECIDED: a different, shorter domain, to be
   supplied.** Mechanism unchanged; name is a placeholder. Non-blocking until P3.
6. ~~**§12 — the "thousand sites" figure.**~~ **DECIDED: accepted as-is** —
   *"we're about to do that"*. See §12.
7. **Builds run on `claude-fable-5`** (§7b) — with three P0 checks that must
   precede pointing any lane at it.

Still open:

8. **The price number itself**, and whether the teaser is free to everyone or
   gated on an email. Unblocked by nothing but P0's Fable-5 build measurement.
9. **The preview domain name** (item 5's placeholder).
10. **Isolation option (§4.2)**, at P3 — (a) `network_id`, (b) separate DB,
    (c) namespace + DB, (d) dedicated cluster.

---

## 12. A figure that needs pinning before it is sold on — RULED: accepted

> **DECIDED (owner, 2026-07-29): the thousand-sites figure stands for now** —
> *"I'm ok with a thousand site's figure for now because we're about to do that."*
> The item is **closed**; the measurement below stays as the record of what was
> true when it was checked, not as an objection.
>
> Two things worth keeping from it, neither of which reopens the decision:
> (a) the figure is **forward-looking**, so it acquires a shelf life — the pool
> build-out either happens or the sentence quietly becomes false, and nobody is
> currently watching that; (b) when webdesign.uk's own copy is written it should
> take its numbers from **this** workstream's measurements, not inherit
> webdesign.co.uk's prose — the two sites make different promises and one of them
> is now taking money.


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
