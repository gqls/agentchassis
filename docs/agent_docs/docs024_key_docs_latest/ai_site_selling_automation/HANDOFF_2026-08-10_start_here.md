# HANDOFF — 2026-08-10 — selling AI-built sites, fully automated: start here

**Status in one paragraph.** Nothing of the automated hookup is built, but far
more of it is real than the ask assumes. The chat intake is LIVE (a visitor can
talk to webdesign.uk's contact page today — verified reaching real visitors
2026-08-10 ~17:00, see §3.1). The build pipeline is LIVE (the 7-agent
work-item path builds whole sites from one Kafka message). The delivery layer is
LIVE (`<slug>.ugg2.com` serves with zero config; a real domain onboards onto
Cloudflare by API, proven 2026-08-09). The admin API even has client CRUD
endpoints already — only the front end lacks them. What does NOT exist is the
wiring between these proven pieces: the chat writes its transcripts to flat
files on an isolated VM with no path to the chassis, no customer record is
created anywhere, no payment gates anything, and the dispatch mechanism that
would auto-trigger builds currently no-ops non-deterministically
(`bugs_open/239`). The work is **integration across a deliberately-maintained
isolation boundary, plus one platform bug**, not greenfield construction.

**Read this file cold, then the two source plans it builds on** (do not
duplicate them):
- `../webdesign_uk_build_service/PLAN_2026-07-28_webdesign_uk_build_service.md`
  §9 — the owner-ruled business phases P0–P7. **This workstream ≈ P4 ("automate
  the trigger") + P5 ("automate the seeding") + the admin FE + the client DB.**
- `../webdesign_uk_build_service/PLAN_2026-08-04_webdesign_uk_vm_hosting.md` —
  the VM/chat architecture and the §2a trust-class ruling.
- Naming trap: the VM plan's "Phase 5/6" (chat box / DNS cutover) are NOT the
  business plan's "P5/P6" (seeding / isolation). This file uses **P-numbers for
  business phases** and spells out the VM phases in words.

---

## 1. The owner's ask, verbatim (2026-08-10)

> "yes I want this site or another site to sell ai generated websites that host
> them as a subdomain on ugg2.com as planned earlier. Please research the docs
> to see what we've discussed here and provide a handoff so I can explore this
> next stage from a fresh chat. I do want the automated hookup next after we've
> cut over this part. Completely automated and I'll need an admin front end to
> manage the client's site builds. We'll need a client db too. Please prepare a
> handoff so another thread can think about it clearly."

Amended same day: **"it won't only be ugg2.com there will be other domains
too"** — see §3.4 for the three delivery shapes that implies.

Decomposed, with where each half already stands:

| ask | what exists | what is missing |
|---|---|---|
| visitor chats → site built automatically | chat live (§3.1); build pipeline live (§3.2) | the hookup (§4.1) + reliable dispatch (§5.1) |
| goes live on a subdomain / other domains | ugg2 wildcard proven; API zone-create proven (§3.4) | slug allocation, registrar-side NS for real domains |
| admin front end to manage client builds | `/api/v1/admin/clients` endpoints live server-side (§3.5) | the FE tab (bolt-on precedent exists) |
| a client DB | `clients→networks→sites` FK chain real (§3.6) | every customer-shaped column; any billing linkage |
| "completely automated" | ONB-019 answers briefing autonomously (§3.7) | payment gate, dispatch reliability, review policy (§7) |

Timing: the owner gated this on "after we've cut over this part" — the
webdesign.uk DNS cutover (VM plan Phase 6) is **still pending** owner review as
of 2026-08-10 (`HANDOFF_2026-08-10c_continue_here.md` §1).

## 2. Prior art you MUST read before designing anything

The design corpus for this exact product already exists, unbuilt, in the
concept register — all status `aspirational` unless noted:

- **SAAS-001/SAAS-002** (`docs026_concept_register/register/saas-isolation-architecture.md:11-23`)
  — the isolation architecture ("an anonymous, internet-triggered,
  token-spending build pipeline **must not run on core**"; boundary is
  "strictly one-directional, async, egress-from-core-only") and the
  conversational intake design (briefing-agent dialogue → job lane).
- **BIZ-009/BIZ-014** (`register/business-strategy.md:73-79,114-121`) —
  build-as-a-service ("itself shipped with its own chat box (explicit
  recursion)"; flags "cost/abuse exposure from anonymous builds, need for
  accounts/billing/quota gating") and the entitlement seams ("**never calling
  Stripe directly** — always through a pluggable billing-adapter interface";
  reuse `clients→networks→sites` for ownership).
- **PAY-002..007** (`register/payments.md`) — the chassis-wide billing design.
  Every DDL is PROPOSED; PAY-007 is `partial` and its measured state is the
  warning: "`status = active` … only reflects 'a row exists', not 'payment
  cleared'" (`payments.md:60`, corroborated `stripe/001commentary.md:44-46`).
- **CHAT-001..009** (`register/site-chatbot.md`) — the edge-worker chatbot
  designs. CHAT-009 names the **"job lane"** — "long-running submissions like
  'build me a site', ack + status + deliver" (`site-chatbot.md:78`). CHAT-008
  is the deliberately-cheap monetisation counter-proposal: flat day-pass via
  stateless signed `{domain, expiry}` token, synchronous Stripe guest
  checkout, **no accounts, no webhook on the critical path** (`:70`).
- **The worked example is literally this workstream**:
  `docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(4).md`
  §12 — "The chat is no longer a Q&A box; it is the **intake + orchestration
  front-end to the whole build platform, offered as a service**" (`:412`);
  "The build must run on the satellite, not core" (`:416`); "Two `sites`
  populations … Core owns the portfolio; the satellite owns customer SaaS
  sites" (`:427`). Newest copy:
  `docs024_key_docs_latest/tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md`.

**Register health warning**: every register file above carries
"covers-through: 2026-07-13 · extraction freeze" — and it is already wrong on
the thing that matters most here. VMB-010/CTS-049 ("component declares it
needs a backend") read `partial`, but the live measurement says **not built**:
"no `semantic_tags` column on `site_components` … 0 active agent definitions
matching `%requires-backend%`"
(`webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md:984`,
2026-08-04). The whole webdesign.uk lane (post-freeze) has **no register entry
at all**. Treat register status fields as hypotheses; re-measure.

## 3. What is provably real and reusable (each with its trap)

### 3.1 The chat intake — LIVE

`chat-input-box` component on webdesign.uk's contact page (position 2,
`page_components` row `fc70ab85…`), loader in `js_snippets`, bundled into
`/assets/js/snippets.js`. Backend: a hand-written Go service,
`webdesign_uk_build_service/box/chat-service/` (`main.go:80` —
`mux.HandleFunc("/api/chat", …)`, port 8081), on the Mythic Beasts VM behind
nginx + a cloudflared tunnel. Model `claude-haiku-4-5` (`claude.go:19`),
measured ~$0.0004/turn. Five abuse gates from first commit: per-IP limit keyed
on `CF-Connecting-IP` (nginx `limit_req_zone $http_cf_connecting_ip`), hard
turn cap (default 20), daily spend ceiling (default $10) **failing closed to
contact details**, request log with tokens+cost, structured transcripts.

Verified 2026-08-10 ~17:00 (this handoff, first-hand): an un-cache-busted fetch
of `snippets.js` — what a real visitor gets — now contains
`chat-input-box-loader` (the 08-10c cache gap has closed), and `contact.html`
serves the box.

Traps: **(a)** `pages.sections` is not durable — a `page-build-handler` rebuild
silently removed the chat box once already
(`HANDOFF_2026-08-10b_continue_here.md` §3). **(b)** the bot's facts
(`chat.go:26-47` `systemPromptFacts`) are a hand-maintained Go constant with
**no code link to `evidence_base`** — "THIS FILE MUST BE UPDATED BY HAND or the
bot will state stale facts". A per-customer bot cannot reuse this design as-is.
**(c)** Cloudflare edge-caches `.js` for up to 4h and the available token has
**no Cache Purge permission** — this already produced one owner-visible "I
submitted and nothing happened" failure (`HANDOFF_2026-08-10c` §0).

### 3.2 The build pipeline — LIVE, and it is the ONLY intake door now

The live flow, evidenced from `site_work_items.handler_agent` +
`site_specs.created_by` (`retired_agents/HANDOFF_2026-08-02_continue_here.md` §2):

```
082_submit_domain_unified.sh  →  ONE Kafka message (system.agent.generic.requests)
  → domain-submitter (files a WORK ITEM, spawns nothing)
  → domain-research-classifier → domain-strategist → vertical-exemplar-researcher
  → site-design-planner → build-briefing-agent → build-site-planner (the builder)
```

The old orchestrator door (`intake-orchestrator` + `site-classifier`) was
**retired 2026-08-02** — soft-deleted with backup + paired restore SQL
(`retired_agents/` §9; DB re-verified 2026-08-10: `deleted_at='2026-08-02'` on
both, non-snapshot). The register's "two intake doors, consolidation open"
(`docs026_concept_register/006_VERIFICATION_stage2.md:348`) is **stale** —
the queue door is the only door. Do not resurrect the old shape; its useful
idea (HITL gates with `skip_if: input_data.hitl_mode == auto`) survives in the
generic actions (`hitl_request_human_input.go`, and the newer
`checkpoint_for_review_action.go` which "replaces the request_human_input →
suspended orchestration pattern").

Traps for the programmatic trigger (all read from the script + migrations):
- **`call_agent`'s `input_mapping` is an allow-list, not a passthrough** —
  any new field (customer id, transcript id, slug) is silently dropped at
  every spawn boundary unless added to every hop's mapping (this already
  caused a silent regression, migration 274).
- **`client_id` is hardcoded `demo_client`** in `082` — this is the exact seam
  where a real customer id must enter, and today it doesn't.
- `--fidelity` is recorded but wired to nothing (except `locked`).
- `hitl_mode=auto` does not merely skip gates — it **synthesises** an
  `auto_confirmed` result from classification defaults and
  warns-and-continues when none exist ("downstream steps may fail",
  `hitl_request_human_input.go:68-71`): an under-specified build, silently.
- **ONB-019 `build-briefing-agent` is `deployed`** and answers the briefing
  questionnaire **autonomously from site_specs, no human**
  (`register/onboarding-config.md`) — the automatable half of briefing
  already ships; the policy question (§7) is whether it should.

### 3.3 The delivery layer — LIVE (ugg2), and it is shared infrastructure

`scripts/cloudflare/worker.js` (`portfolio-sites-router`): object key is
literally `` `${hostname}${path}` `` into the hardcoded `portfolio-sites` B2
bucket (`worker.js:2,33`), `/index.html` default, no config table, no KV.
**Publishing a customer site = uploading `<slug>.ugg2.com/index.html` to the
bucket. Un-publishing (the refund lever) = deleting the objects.** The
`*.ugg2.com/*` route + wildcard DNS were proven 2026-08-02: two
never-before-seen subdomains served on first request
(`PLAN_2026-07-28_webdesign_uk_build_service.md:495-521` — "No box, no
certbot, no DNS-01, no scoped token, no renewal timer").

Traps: **(a)** the wildcard `*` DNS record and the route live **only in live
Cloudflare** — not in this repo, not in Terraform. **(b)** the same worker
fronts **all 38 active zones** (`bugs_open/236` §census) — a bad worker deploy
takes down the whole portfolio, and deploying it without its two B2 secret
bindings "strips the worker's credentials and takes every site down"
(`scripts/cloudflare/README.md:29-31`). **(c)** Universal SSL covers one label
deep — `<slug>.ugg2.com` yes, `x.y.ugg2.com` no.

### 3.4 "Other domains too" — the three delivery shapes (owner amendment)

1. **`<slug>.ugg2.com`** — zero config, proven. The preview/default shape.
2. **A real customer domain** — onboards onto Cloudflare with the standard
   template: zone + apex A placeholder + www CNAME + both
   `portfolio-sites-router` routes. **API zone-create is proven as of
   2026-08-09** (cookly.uk: HTTP 200 zone create, all four records/routes by
   API — `domains_cloudflare_rollout/NOTES_domains_cloudflare_rollout.md:68-100`).
   The remaining manual gap is **registrar-side NS delegation** (the
   customer's registrar, or ours — Nominet EPP TAG + three registrar keys
   still owed, `PLAN_2026-08-03_domains_cloudflare_rollout.md`). Landmine
   from the same NOTES file: a domain can `curl` 200 while serving the
   registrar's **parking page** — `dig +short NS` is the truth, not HTTP.
3. **Subdomains of other owned parents** — same worker; needs that parent's
   own wildcard DNS + route created once (only ugg2.com's exists today).

The domains rollout lane's **static-first + skip-list rule** applies to every
new zone: static template by default; a domain leaves the template only when
it will host a framework-backed dynamic service, and "the skip-list GROWS as
services launch" (`PLAN_2026-08-03…:31-37`).

### 3.5 The admin front end — smaller than it looks

`frontends/admin-dashboard/`: 3 source files, no router, one 2,499-line
`App.tsx`, navigation is one `useState` string; `PipelinesPage.tsx` (402
lines, self-contained, mounted as one more `view ===` branch) is the proven
bolt-on-tab precedent. **Zero customer concept in the FE — but NOT in the
API**: core-manager already serves `POST /clients`, `GET /clients`,
`GET /clients/:client_id/usage` under `/api/v1/admin`
(`internal/core-manager/api/server.go:138-140`, handlers in
`internal/core-manager/admin/`). A "Customers" tab is FE-only work against
endpoints that already exist.

### 3.6 The client DB — the skeleton is real, the flesh is absent

`clients → networks → sites` FK chain confirmed live (both FKs `ON DELETE
CASCADE`; 2026-08-10: 2 placeholder client rows — "Default Client", "System
Scheduler" — 1 network, 39 sites). `clients` has exactly
`id / external_id / name / settings(jsonb) / created_at / updated_at` — **no
email, no billing, no tier, no Stripe id**. `clients.external_id` is
documented as "the natural place to map a Stripe [customer]"
(`stripe/001commentary.md:28`). Note today the customer's contact lands on
the **site** row, not a client: `sites.email/phone` exist
(`sql_for_tables/011_sites_table.sql`) and `082 --email` writes them there.

Traps: **(a)** the DDL for these three tables lives in
`docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql`
— `platform/database/migrations/` does not contain them; there is **no
ordered, replayable migration story** for exactly the tables this workstream
must extend. **(b)** `sites.status` DDL default is `'active'`, and tools-api's
CORS allowlist is DB-driven off `status='deployed'`
(`internal/tools-api/store/sites.go:33-35`) — a second "zero config once
deployed" surface, and a second consumer of that column to not break.
**(c)** `site_chat_turns` **is already designed**
(`sql_for_tables/046_site_chat_turns.sql`: session, question/answer, refusal,
tokens, cost, `client_ip_hash`, idempotent PK) — the DB destination the chat
service's JSONL currently bypasses. Treat "chat transcripts into the DB" as
finishing this design, not inventing one.

### 3.7 Payments — one proven pattern, one half-built surface, a decision between them

- **idea.uk's pattern is proven with real money** (£29 live payment
  2026-06-14, `register/payments.md:11`): 2-method Provider interface
  (`docs/…/idea.uk/golang_files/billing.go:28-34` — `CreateCheckout` /
  `ParseWebhook`), raw net/http + hand-rolled HMAC verify (no SDK),
  **signature-verified `checkout.session.completed` as the sole source of
  truth** (`billing.go:104`), idempotent via event dedup, restricted API key
  (Checkout Sessions:Write only — no refund capability), orders in a flat
  JSON file. Refunds are manual-dashboard-only by design. Caveats: the
  proof has not been repeated since idea.uk's Cloudflare front-door cutover
  (2026-08-04 — the first organic webhook through the new door is still
  owed); and idea.uk's code is a nested Go module under `docs/`, outside
  `go build ./...` and the image pipeline — the same structural position as
  the chat service.
- **`internal/auth-service/subscription/`** is in the build and has
  `stripe_customer_id`/`stripe_subscription_id` columns — but "no Stripe SDK
  … no webhook handler anywhere … `status = active` means 'a row exists'"
  (`stripe/001commentary.md:44-46`).
- The entitlement design when needed is small: "entitlement state hanging off
  clients/networks, a billing adapter keyed on `external_id`, a
  `pending_entitlement` reuse of `approval_mode` at build-submission, and a
  join-filter on three selection queries at maintenance. All four are seams
  in existing structures" (`001commentary.md:32`; PAY-003).

### 3.8 Adjacent proven pieces worth knowing exist

- `platform/mailer` (PUB-003): built, 8 tests, council-approved, **zero
  importers** (re-confirmed 2026-08-09 via `bugs_open/228:334`); explicitly
  excludes retries/queueing/bounce (`mailer.go:36-37`) — relevant to any
  "email the customer their link" step, and its adoption story is the
  cautionary tale that porting ≠ landing.
- tools-api (`cmd/tools-api` + `internal/tools-api/`, runtime on the island
  VM via docker compose + tunnel): one hardcoded product group
  (`/api/v1/tools/gauntlet`), one Anthropic key, no generic form/booking
  endpoint — but reusable middleware (rate limit, input cap, LLM call log)
  and the DB-driven CORS above. The named-but-unbuilt "contact forms POST to
  tools-api → platform/mailer" design (`bugs_open/228:330-337`,
  "architecture-scope, not a bug fix") is the closest thing to a shared
  customer-facing backend pattern, still unowned.
- `site_unreachable` discovery check (IMP-053): **built, not yet live**
  (rides next chassis roll; driver config HELD) — closes bug 236's "a live
  site can serve 522 and nothing notices". A selling pipeline should not
  go-live customers before this is seated.

## 4. What is confirmed ABSENT (with the proof)

1. **The chat→build hookup.** `grep` of `box/chat-service/*.go` for
   kafka/postgres/pgx/work_item: **nothing**. The service writes
   `requests.jsonl`/`transcripts.jsonl` to local disk on a tunnel-only VM.
   Its system prompt collects exactly the two facts a build needs — "Ask what
   business the visitor runs and what domain they'd want the site on. Do not
   ask for anything else." (`chat.go:47`) — and then they go nowhere.
2. **Backend generation for customer sites** — three independent lines:
   DYN-001 tier 2 aspirational; `site-api-router` has **zero code** (grep of
   all go/yaml/sql/sh/tf: only markdown hits, and the register itself flags
   it "verify-later"); the "component declares it needs a backend" gate has
   no column to hang on (§2's live measurement). Customer sites are
   **static/Tier-1 by construction** — which the trust-class ruling (§6)
   turns into a feature.
3. **Any customer/billing column in `clients_db`** (§3.6) and **any customer
   concept in the admin FE** (§3.5 — API exists, FE doesn't).
4. **Site ownership**: ADM-008 `site_ownership` junction table —
   "**abandoned** … Never created; admin API sidesteps it"; "`sites` has no
   ownership columns" (`register/admin-dashboard-and-api.md:67-73`). The
   user-scoped public API half (ADM-007/PUB-001) was never built.
5. **Confirmed-delivery contact forms, estate-wide**: every framework site
   ships a `mailto:` handoff (bug 228's class fix made it an *honest*
   mailto, not a delivery mechanism). The webdesign lane's own README calls
   the consequence out: "a form that posts to a server is a paid extra".
6. **Register coverage of everything post-2026-07-13** — the chat service,
   the VM hosting pattern, ugg2 preview delivery and this workstream itself
   have no register entries. Per the standing routing rule
   (`PLAN_2026-07-28…:1016-1030`) the intake/trigger seam is
   architecture-scope and owes its register entry **in the same commit that
   ships it**.

## 5. Blockers and constraints, ranked

1. **`bugs_open/239` — dispatch unreliability — blocks "completely
   automated" outright.** An `orchestrate` kcat envelope non-deterministically
   falls back to `owner_agent_type='generic'`, runs **nothing**, and reports
   `COMPLETED`. Byte-identical resends behave differently. The build queue is
   starved; the webdesign lane hand-drove every stage. An automated selling
   pipeline built on this dispatch would silently sell nothing. **Do not
   bisect it against production** — doing so once regenerated a live page's
   hero (the file's own warning). Root cause not yet found at Go level; the
   pointer is workflow *selection*/caching, not `extractGroupInfo`.
2. **The isolation boundary is a design constraint, not an obstacle.**
   SAAS-001: one-directional, async, **egress-from-core-only**. The hookup
   must not give the box synchronous or write access into core. A
   pull-from-core shape (core polls/ingests the box's transcripts —
   the `site_chat_turns` header describes exactly this: edge → sink →
   Layer-1 puller) conforms; "the chat service POSTs into the cluster" does
   not.
3. **`bugs_open/240`** (Kafka metadata storm): no per-job topics — IMP-053
   was deliberately built spawn-free for this reason. Constrains any
   per-customer build agent design.
4. **Prompt-injection scope** (`PLAN_2026-07-28…` §4.2, owner-ruled): "A
   build for a customer consumes attacker-supplied content … agents that
   hold write access to production content". Boundary shipped now; the
   full isolation decision is **deferred to P3 with a named trigger — "the
   first paid build that scrapes a domain we do not own"**. The automation
   thread inherits that trigger; it must not quietly cross it.
5. **Sequencing**: webdesign.uk's own DNS cutover (VM Phase 6) is still
   pending owner review, and the owner gated this workstream on it. Phase 6
   step 4 requires the ugg2 path be untouched by the cutover.
6. **Registrar credentials** (real customer domains): Nominet EPP TAG +
   dynadot/porkbun/spaceship keys still owed (domains rollout lane).
7. **Demand reality check**: idea.uk — complete, verified, working — has
   **zero genuine external buyers** to date; its landmine reads "a product
   can be complete, verified and working and still never have completed a
   transaction". webdesign.uk's stated phase goal is enough traffic to find
   customer-handling bugs, not volume. Build automation in the order that
   demand evidence arrives, not ahead of it.

## 6. Settled rulings — do NOT re-open (cite them, inherit them)

- **Trust classes** (`PLAN_2026-08-04…` §2a, owner 2026-08-04): money-live
  keeps its own box; our product sites share the new box; **customer
  deliverables go to B2 via the worker, never on a box** — "customer sites
  do not consume box capacity"; scale in customers is scale in objects.
- **Framework-only builds** (owner ruling 2026-08-04, CLAUDE.md): every site
  goes through the framework; hand-building is a bug to file, not a shortcut.
- **Static-first + skip-list** for every new zone (§3.4).
- **Isolation decision deferred to P3 with its named trigger** (§5.4) —
  neither pre-build the satellite nor ignore the trigger.
- **Pricing** (owner, 2026-08-10, `PLAN_2026-08-04…` §7): £1,200 total, no
  VAT, **£75 non-refundable deposit**, 2 revision rounds, decline within 14
  days → £1,125 back. The copy-the-site-and-refund worst case was discussed
  and deliberately NOT engineered around — low volume is how the owner finds
  out if the fear is real. No traffic limiter yet: with Stripe still
  test-mode, the human fulfilment step IS the limiter.
- **P0–P7 sequencing** (`PLAN_2026-07-28…` §9): P3 manual fulfilment is "a
  shipping business with zero new platform code"; P4 automate the trigger;
  P5 automate the seeding ("the largest single piece"); P6 isolation; P7
  ownership/handover. Council routing for P4/P5: architecture-scope, one
  run per coherent task, register entry in the shipping commit, **measure
  blast-radius claims before submitting** (the 124 lesson).

## 7. Open decisions for the owner (named, not resolved here)

1. **Where customer identity lives.** The corpus contradicts itself:
   BIZ-014 says reuse `clients→networks→sites` (no new ownership columns);
   the isolated-chat plan's TL;DR still says "honour now: `owner_id` on
   sites"; ADM-008's junction table is abandoned. One shape must win before
   any schema work.
2. **Which Stripe surface grows**: port idea.uk's proven 2-method Provider
   into the buildable repo (the `platform/mailer` lift-and-land precedent —
   noting mailer's zero importers as the cautionary half), or finish
   `auth-service/subscription`'s half-built columns. Related: does live
   Stripe (P3's gate — written terms still owed by the owner) precede or
   follow the automation work?
3. **Which tier is the automation target.** The £1,200 done-for-you tier
   keeps a human by design (2 revision rounds) — automating its *intake*
   (chat → seeded specs → triggered build → preview link) still leaves a
   human releasing the build. The £19 all-in self-serve tier (owner,
   2026-08-10, "future direction, not current scope") is the only shape
   that is honestly "completely automated" end-to-end. Naming which one P4/P5
   serves changes how much review machinery survives.
4. **How much human review survives, and reviewing what.** ONB-019 already
   answers briefings autonomously; `hitl_mode=auto` exists but synthesises
   defaults silently; `checkpoint_for_review_action` is the modern gate. The
   platform's own doctrine (WDS-004, CLAUDE.md) is positive evidence over
   status — a review step, if kept, should look at the rendered preview, not
   a `complete` flag.
5. **Where the chat lives long-term**: stays on the VM (current, proven,
   per-site) vs CHAT-002's edge-worker shape (per-domain at scale, no box in
   the serving path). Bears on the recursion in BIZ-009 — every sold site
   ships its own chat box.
6. **Domain ownership and handover** (P7, now sharper given "other domains
   too"): who registers a customer's real domain, in whose account, and what
   transfers on decline/refund. The delete-objects lever (§3.3) only works
   for domains we control.
7. **Refund mechanics**: deposit £75 in; on decline, £1,125 back —
   manual-dashboard refunds only (no refund code exists anywhere, by
   design). Acceptable at automation volume, or does the entitlement work
   (PAY-003) come forward?

## 8. Recommended sequencing (evidence-ranked, cheapest-real-first)

1. **Admin "Customers" tab + client-DB columns.** Server endpoints exist;
   FE bolt-on precedent exists; zero platform risk; unblocks nothing but
   informs everything. Decide open question 1 first (it names the schema).
2. **Chat transcripts → `site_chat_turns`** (finish the 046 design).
   One-directional core-pulls-from-box, so isolation-safe; turns the JSONL
   dead-end into queryable demand data; the first integration across the
   boundary and the rehearsal for the trigger.
3. **Fix or reliably route around `bugs_open/239` before ANY auto-trigger
   work.** Nothing downstream is trustworthy until dispatch is. (Check the
   queue first — a diagnosis may already be filed.)
4. **The trigger seam (P4)**: transcript/intake record → `082`-shaped Kafka
   envelope, with a real `client_id` replacing `demo_client` and every new
   field added to every hop's `input_mapping`. Architecture-scope: council
   run + concept-register entry in the shipping commit. This is the first
   platform-code change of the workstream.
5. **Seeding automation (P5)**: brief → `site_specs` aspects — "the largest
   single piece". ONB-019 is the existing autonomous half; the missing half
   is chat-brief → specs.
6. **Payment gating** per open questions 2/7, scoped by which tier won
   question 3.
7. Throughout: customer sites **static/Tier-1 only**; the P3 isolation
   trigger stands; go-live for customers waits for IMP-053 (site
   reachability) to be seated.

## 9. Citations not already inline

- Owner ask + amendments: webdesign session transcript 2026-08-10 (this
  handoff quotes it verbatim in §1); pricing conversation ibid.
- Retirement evidence: `retired_agents/HANDOFF_2026-08-02_continue_here.md`,
  `RESTORE_intake_path_orphans.sql` (restore is a pair — both or neither).
- Live-DB checks (2026-08-10, this handoff): `agent_definitions` rows for
  intake-orchestrator/site-classifier (`deleted_at=2026-08-02`);
  `clients` count=2 + column list; networks=1, sites=39; webdesign.uk row
  `github_repo='vm-sites'`, `deploy_config={"target":"vm","capabilities":["backend"]}`.
- Chat delivery check (2026-08-10 ~17:00): un-cache-busted
  `curl https://preview.webdesign.uk/assets/js/snippets.js | grep -c
  chat-input-box-loader` → 1; `contact.html | grep -c chat-input-box` → 32.
- Bug files: `bugs_open/239_…orchestrate_dispatch_falls_back…`,
  `bugs_open/236_…522…` (owned by `bugfix_236_site_availability/`),
  `bugs_open/228_…contact_block…`, `bugs_open/240` (via IMP-053's note).
- Lane state: `webdesign_uk_build_service/HANDOFF_2026-08-10b/c_continue_here.md`,
  `SUMMARY_2026-08-10_webdesign_uk_build_service.md`.
- Domains: `domains_cloudflare_rollout/PLAN_2026-08-03…`, `NOTES_…` (cookly.uk
  zone-create, 2026-08-09).

**What would falsify the big claims here**: 239 fixed and dispatch proven
reliable (re-run its repro, not this file's summary); a register un-freeze
(re-check `covers-through` banners); the sibling webdesign session shipping
newer handoffs than 08-10c (it was live while this was written — re-read that
directory before trusting §3.1's snapshot).
