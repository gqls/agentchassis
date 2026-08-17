# PLAN 2026-08-17 — delivery architecture decisions (snapshot of the planning session)

> **What this file is**: a verbatim snapshot of the 2026-08-17 planning
> session's working document (owner + session, five discussion rounds),
> taken at the moment the owner approved it. The OWNER DECISIONS section is
> the authoritative record; earlier sections deliberately retain superseded
> drafts with strike-through markers, because the trajectory of the
> reasoning is part of the record. `PLAN_2026-08-14_site_delivery_and_editor.md`
> still governs Phases 3–6 MECHANICS wherever this file does not explicitly
> supersede it (it supersedes: Part 2d's "ownership reverts to the ZIP"
> conclusion — ownership is now domain + their-own-Netlify + ZIP; and the
> single-subscription pricing sketch).

# Delivery architecture rethink — hosting ownership, domains, repos, forms, editing

## Context

Phase 2 (the publish seam) is proven in production. Before more code, the owner
has reopened the delivery-architecture questions: per-customer hosting accounts,
domain-per-site (owner is a Nominet registrar), where the customer's site lives
(repo model), forms/email without unpaid ongoing work, and the editing story.
High anticipated load; everything must be automatic. Scale review (own
cluster(s)) deferred until after the working site.

## Facts established by exploration (2026-08-16)

**Nominet/EPP — much further along than assumed:**
- TAG `DESIGNCONSULT` live; EPP LOGIN **proven from the cluster** (result 1000,
  egress node .26); IP allowlist cleared (5 cluster IPs added 2026-08-11).
- A working ~150-line stdlib Python EPP client EXISTS and is register-deployed
  (VMB-015: NS changes, dry-run default, host:create retry, IPv4-pinned) —
  `idea_uk_vm_site/box/nominet-epp-ns-change.py`.
- A SECOND TAG for customer domains is already APPLIED FOR (pending Nominet) —
  the separation the domain-per-site idea needs was anticipated.
- What does NOT exist: domain REGISTRATION (acquisition) via EPP — only NS
  repointing. Registration = new EPP verbs on the proven client. `domain:create`
  is the missing piece, not the plumbing.

**Payments — one-off only today:**
- PAY-009 deployed keyless: `mode=payment` HARDCODED (stripe.go:46); provider
  interface has only CreateCheckout+ParseWebhook. No subscription machinery is
  wired anywhere (PAY-007 scaffold "stamps status=active without payment" — a
  recorded trap, deprecate-or-unify ruled for after first sale).
- £10/mo domain retention therefore needs either Stripe Payment Links
  (near-zero code, link in delivery email) or the PAY-005-designed subscription
  methods. PAY-002 is the ruled shape for recurring.
- Standing blocker: Stripe keys (owner) AND the webhook URL is edge-blocked
  (apex 302s; register `preview.webdesign.uk/stripe/webhook` or add an edge
  exception — webdesign_uk lane's item).

**Netlify — the matrix already contains the automatable ownership path:**
- PLAN 2026-08-14 row 2: **"Netlify via customer OAuth connect — fully
  automated (customer clicks Connect once); REAL ownership — site created in
  THEIR account; per-customer tiers, no shared ceiling."** The CUSTOMER does
  their own signup (their captcha, their email), clicks Connect once, and our
  automation deploys into their account via OAuth. ToS-clean, no transfer step,
  billing lands on them directly. This was already parked as the future paid
  "hosted ownership" extra (PLAN:163).
- Mass-creating accounts with temporary emails remains ToS-hostile at every
  provider; no provider has an account-creation API.

**Repo mechanics — the git-adapter is NOT the vehicle for customer repos:**
- `sites.github_repo` is a BARE name; owner comes from adapter env
  (GITHUB_ORG=gqls); "org/name" produces a malformed API path, no validation.
- The adapter auto-creates missing repos PUBLIC (recorded trap); `create_repo`
  is NOT reachable from workflows (3-verb allowlist + live RFC_011 guardian
  veto on widening that shared vocabulary).
- One GitHub Actions runner deployment PER REPO is the established cost of the
  monorepo deploy chain (gqls/sites + gqls/vm-sites each have one).
- The publish seam is the right extension point: Publisher interface + For()
  confirmed one-const-one-case per backend, no caller changes; but Deps carries
  only ObjectStore — a `github` backend needs a direct GitHub API client (new
  Deps field), not the Kafka adapter.

**Forms — an owner ruling already exists, plus fresh parallel research:**
- Fleet truth: 11 sites' contact forms are `mailto:` BY RULING; the repair
  seam (LNK-032) rewrites dead actions to mailto and refuses without a real
  address. Owner's own README: "the contact form opens an email; a form that
  posts to a server is a paid extra" — needs one sentence in the offer.
- Formspree + Basin are already the NAMED third-party options in webdesign.uk's
  evidence base (measured: no affiliate programme at either).
- Fresh RESEARCH (2026-08-15, dartsonline lane, owner-commissioned): recommends
  a minimal enquiry endpoint on tools-api (deployed, has DB+router, currently
  gauntlet-only routes) → table → email notify → repoint the shared component
  opt-in per site. Also: NO spam protection exists anywhere in the estate.
- `platform/mailer` is real, sanctioned, live only on the gripper path.

**Serving/DNS at scale:**
- Zone-per-domain is the current pattern, ALL MANUAL (zone-create never yet
  exercised via API); no Cloudflare client in Go anywhere; 1,200 req/5min
  account API limit; UNPUBLISHED zone cap flagged ("thousands of zones may need
  a CF support ticket mid-run"); "Cloudflare for SaaS"/custom hostnames: zero
  prior thinking in the estate.
- Worker serves any hostname prefix with zero per-site config (proven);
  per-new-domain manual steps: zone create → A+www records → 2 worker routes →
  NS repoint (EPP client exists for that) → poll active.

## Draft recommended shape (for discussion)

1. **Hosting default: ours** (B2+worker, near-zero marginal cost, fully
   automatic). **Ownership = domain + ZIP (+ repo mirror later)**.
2. **"Your own hosting account" = the OAuth connect path** (Netlify first),
   offered as an extra: customer signs up themselves (ToS-clean), clicks
   Connect once, our seam deploys into their account via a `netlify-oauth`
   publisher backend. Their account, their billing, their bandwidth ceiling.
   No temporary-email account minting — that pattern risks bulk suspension.
3. **Domain-per-site**: framework picks a .uk name, registers via EPP
   (`domain:create` added to the proven client, under the NEW customer tag once
   Nominet grants it), zone+routes via a small CF API client (to build), NS set
   at registration. Delivery email carries a £10/mo retention payment link;
   lapse policy = NS repoint/park at year end (registration year is sunk, ~£4).
4. **Repo deliverable**: defer to a later phase as a paid/power-user extra via
   a `github` publisher backend (direct API client, dedicated org, private
   repos, collaborator invite). ZIP remains the ownership artefact in v1.
5. **Forms v1**: keep the existing ruling (mailto + one offer sentence).
   The paid forms tier later = tools-api enquiry endpoint (per the fresh
   research) with spam protection designed in. Do NOT silently adopt a
   third-party relay the evidence base doesn't name.
6. **Editing**: own editor (Phases 5–6) unchanged; repo mirror is the
   power-user path later.
7. **Build order**: Phase 3 ZIP (invariant) → handover+email (Phase 4, now
   carrying the domain-retention link) → auth (5) → editor (6). Domain
   programme runs as its own phase alongside once the second TAG lands.

## TODO additions (owner-requested, to record in the lane)

1. Domain-per-site programme (above) — blocked on: second Nominet TAG grant;
   Stripe keys; webhook exposure.
2. Whole-architecture scale review incl. own cluster(s) — AFTER working site.
   Concrete agenda seeded: CF zone cap (unpublished), zones-vs-CF-for-SaaS,
   runner-per-repo model, single monorepo build throughput, spawned-pod-per-
   reconciler-tick economics, cluster capacity.
3. Busy-site payment thread — traffic visibility on our infra = upsell
   trigger; OAuth-connected accounts put bandwidth billing on the customer
   directly.

## Hosting operations — what "we host" commits the owner to (discussion 2)

**Serving cost ≈ zero** (B2 storage pennies per site; B2→CF egress free via
Bandwidth Alliance; CF free plan unmetered). The REAL costs are four
operational surfaces:

1. **Metering + billing**: to charge for traffic you must measure it
   per-hostname. Build: the worker writes per-hostname counters (Workers
   Analytics Engine — designed for this, generous free tier; the estate
   already has a visit-beacon pattern, TRF-006) → scheduled sweep reads
   aggregates → threshold crossing files an upsell/notice item → payment
   link. One-time build + one sweep; human time only on disputes.
2. **Abuse / shared fate**: thousands of customer sites under ONE CF account
   and ONE B2 bucket = host-of-record obligations (abuse contact, takedown
   process, ToS that permit acting) AND blast-radius risk: one phishing page
   can flag the shared account. Mitigation the estate uniquely has: the
   framework builds all content and edits flow through the editor's
   structured fields + locks — content control is automatable (claims sweeps
   already exist). Real but manageable; the isolation architecture below is
   the structural half.
3. **Support posture**: £149 cannot fund phone support. Email + the chat-box
   pattern (sibling lane's build) only; status page; SLA language in the
   offer copy. Support time is priced INTO the subscription (below), not
   given away.
4. **DNS/TLS architecture at scale** — "how do I isolate accounts in
   Cloudflare": CF has NO sub-accounts on self-serve. Three real options:
   - **A. Zone-per-domain, one account** (current pattern): free, full
     per-zone analytics/SSL; but unpublished zone cap (~1k region cited),
     1,200 req/5min shared API limit, one account-wide token (recorded
     landmine: no per-zone fence), zero isolation — account suspension takes
     every customer down.
   - **B. Own authoritative DNS + Cloudflare for SaaS** (the at-scale shape
     when WE are the registrar): being registrar means WE must host DNS
     anyway — zone-per-domain IS that today. At scale, replace with a boring
     self-hosted authoritative DNS pair (template zone file per domain,
     fully automatable, pennies) + CF for SaaS custom hostnames on ONE zone
     ($0.10/hostname/mo past 100, automated TLS). Shrinks token blast
     radius, kills zone sprawl and API-limit pressure. Zero prior estate
     thinking — scale-review item.
   - **C. Account sharding / CF Tenants** (partner platform for managing
     many customer accounts): caps shared-fate per shard; operationally
     heavier; Tenants needs partner status. Scale-review item.
   **Trajectory: A now (proven, free, fine for tens–hundreds) → B near
   ~500–1,000 domains → C only if partner status earns its keep. The scale
   review owns the trigger.**

~~**Pricing reframe**: ONE combined "keep your site live" subscription~~ —
**SUPERSEDED by OWNER DECISION 1 below: fees stay SEPARATE** (domain £10/mo;
hosting deliberately expensive, offered beside free Netlify-connect).
(The open-questions list that stood here is answered in full by the
OWNER DECISIONS section below.)

## The Netlify-connect timing question (discussion 3)

> **Timing SUPERSEDED by OWNER DECISION 2b below** (connect moves into the
> request-phase build-wait window; delivery-email link becomes the repeat,
> not the first ask). The flow mechanics and pressure-valve reasoning here
> still stand.

Where the "set up your Netlify account" step goes:
- **At request/before payment: NO** — third-party account creation at the top
  of the funnel kills conversion on a "most competitive offer" product, and
  incomplete signups generate pre-revenue support.
- **Delivery email + a standing button (recommended)**: the sale completes on
  OUR hosting with zero customer steps (site already live on its domain — the
  wow moment is unconditional). The delivery email offers "put it in your own
  Netlify account" as self-serve; the same button lives permanently in the
  customer's editor/account page. A customer who never clicks costs us a
  quiet static site; a customer who clicks removes themselves from our ops
  surface voluntarily.
- **The pressure valve**: when metering flags a busy site, the upsell email
  offers BOTH the paid tier with us AND free connect-to-your-own-Netlify —
  the customers who cost the most are exactly the ones invited to take their
  hosting home. Ours-default + OAuth-option is therefore MORE trouble-free
  than Netlify-default: no customer action ever gates delivery, and load
  exports itself.

OAuth flow shape (netlify-oauth publisher backend, later phase): one-time
Netlify OAuth app registration (ours) → customer clicks Connect → signs
up/logs in at Netlify, authorises → we hold a per-site token (encrypted) →
seam creates the site in THEIR account, zip-deploys, sets the custom domain,
we flip DNS (we are the registrar — one EPP/zone change). Token kept so
editor changes keep flowing (the framework-keeps-write-access ruling); if
they revoke, their site keeps serving but stops receiving updates — clean
degradation, their choice, wording in the delivery email.

Composition with Option B: not either/or. Ours-default + OAuth-option now;
the A→B DNS/SaaS migration decision applies only to whatever REMAINS on our
infra at scale-review time.

## OWNER DECISIONS (2026-08-17, discussion 4 — supersede earlier drafts)

1. **Fees SEPARATE**: domain "keep this domain — £10/mo" · hosting priced
   deliberately HIGH (owner does not want the hosting business), offered next
   to free Netlify-connect and possibly other third-party options.
2. **Free-rider policy: NO free custom-domain serving.** Site always visible
   on the preview subdomain (our brand); the CUSTOM domain serves the
   choose-a-home page from day one until they pick Netlify / paid hosting /
   ZIP-elsewhere. (Owner chose this over courtesy-window-then-park.)
2b. **Netlify connect moves INTO the request phase (owner, discussion 5):**
   the request-CONFIRMATION page + email (the build-wait window — "usually
   ready the next day") carry "set up where your site will live: connect
   your free Netlify account". NOT on the request form itself (form stays
   zero-friction), and SKIPPABLE (a stalled signup never blocks a build or
   sale — the link repeats in the approval + delivery emails). Most sites
   are then born directly into the customer's account at delivery; the
   choose-a-home page becomes the fallback for non-completers only. Token
   hygiene: OAuth tokens for requests that never convert expire after N days.
3. **Account surface: v1 = delivery email carries every link** (ZIP, Netlify
   connect, hosting payment link, domain subscription link, Stripe hosted
   customer portal); **Phase 6 editor login home becomes the account hub**
   (Edit / Domain / Hosting / Billing). No standalone account page.
4. **Own authoritative DNS: GO** — as part of the DOMAIN programme (not the
   scale review): every registered domain needs DNS regardless of hosting
   home; three zone templates (ours / netlify / choose-a-home page); EPP sets
   NS at registration; first customer domains may ride zone-per-domain until
   the pair is live (migration = one proven EPP call each). Our-hosted sites
   then need CF for SaaS custom hostnames for TLS/CDN (small count if the
   expensive tier stays small).
5. **Deferred by owner (recorded, do not build)**: newsfeeds/editorial
   pieces as a paid-hosting perk.

## Build order (recommended, given the decisions above)

**Phase 3 — ZIP deliverable.** Unchanged by every decision above, and now
load-bearing for all three doors (it is what a Netlify deploy uploads AND
what "take it elsewhere" hands over). Build next.
⚠ Do NOT reuse b2worker's whole-buffer upload for the ZIP's own output
(B2 411s a non-seekable body; a site ZIP is a different size class — stream
with known length or multipart). Runs in a spawned storage-enabled pod.
Reuse `publish.S3Source`/`ObjectStore` for listing+reading.

**Phase 4 — handover + the emails.** `sites.handed_over_at` (single reader:
the editor gate) + the delivery email carrying every link: ZIP, Netlify
connect, hosting payment link, domain subscription link, Stripe hosted
portal. Also the request-confirmation email's connect prompt (decision 2b).
Uses `platform/mailer` (sanctioned, live on the gripper path).

**Phase 4b — netlify-oauth publisher backend.** One const + one case on the
proven `publish.For()` seam (needs a new `Deps` field for the OAuth/HTTP
client — `Deps` carries only `ObjectStore` today). Per-site encrypted token;
revocation degrades to "serves but stops updating".

**Phase 5–6 — customer auth + editor** (unchanged plan; editor home becomes
the account hub per decision 3; cross-tenant probe is the acceptance).

**Domain programme (parallel track, gated on the second Nominet TAG).**
`domain:create` added to the proven EPP client (VMB-015,
`idea_uk_vm_site/box/nominet-epp-ns-change.py`) + own DNS pair with three
zone templates (ours / netlify / choose-a-home) + choose-a-home page +
£10/mo retention via Stripe Payment Link. Note `mode=payment` is hardcoded
in `stripe.go:46` — recurring needs the Payment Link route or the
PAY-005-designed methods; PAY-007's scaffold must NOT be trusted (stamps
active without payment).

**Blocked on the owner, gating first revenue:** Stripe keys + the webhook
edge exception (apex 302s and Stripe treats 3xx as failed delivery);
second Nominet TAG grant (domain programme only).

## Deliverables for THIS execution (owner-requested, 2026-08-17)

All in `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/`
unless stated. One pathspec commit per coherent group.

1. **Snapshot this plan into the repo** as
   `PLAN_2026-08-17_delivery_architecture_decisions.md` — verbatim copy of
   this file (it lives outside the repo today, so the summary has nothing
   durable to point at otherwise). Header states it is a snapshot of the
   2026-08-17 planning session, and that `PLAN_2026-08-14` still governs
   Phases 3–6 mechanics where this file does not supersede it.
2. **Record the decisions where the work happens:**
   - `NOTES_…md` — append a 2026-08-17 entry: the five owner decisions +
     2b, each with its REASON, and the facts that changed my earlier advice
     (Netlify-OAuth was already in the 08-14 matrix; EPP login already
     proven + second TAG applied for; `mode=payment` hardcoded; forms
     already ruled mailto). Marked as superseding the 08-14 PLAN's Part 2d
     "ownership reverts to the ZIP" conclusion — ownership is now
     domain + their-own-Netlify + ZIP.
   - `README_where_we_are.md` — append the owner's plain-prose account of
     the decisions (no jargon; what a customer will experience).
   - **TODO/backlog** items to record in NOTES (owner-requested):
     (a) domain-per-site programme, (b) whole-architecture scale review incl.
     own cluster(s) — AFTER the working site, (c) busy-site payment thread,
     (d) DEFERRED: newsfeeds/editorial as a paid-hosting perk.
3. **`SUMMARY_2026-08-17_site_delivery_and_editor.md`** — NEW file (never
   overwrite 08-16), five headings, plain prose to be read aloud, pointing
   at the plan snapshot (1) for the full reasoning. Justified as an
   inflection: the delivery architecture changed materially (fees split,
   no free custom-domain serving, own DNS, Netlify connect at request
   phase), so "where we are now" reads nothing like 08-16's.
4. **`HANDOFF_2026-08-17_phase3_zip.md`** — cold-start for a fresh chat:
   Phase 3 scope + the carried hazards (no whole-buffer upload; spawned
   storage-enabled pod; reuse `publish.S3Source`; `unzip -l` count ==
   object count acceptance; presigned 200-in-expiry/403-after), the
   decisions that now bear on it (ZIP is what Netlify uploads AND what
   "take it elsewhere" hands over), council+register obligations, and the
   falsifiers. Supersede pointer added at the top of
   `HANDOFF_2026-08-16_continue_here.md` rather than deleting it.
5. **Memory**: update the lane's `MEMORY_workstreams.md` line to point at
   the new handoff and carry the one-line decision state.

Register/council: no platform code changes in this batch (docs only), so no
council run is owed. DGH-008 needs no status edit. When Phase 3 code lands,
it takes its own council round + register entry per the PLAN roll-up.

**The v1.0.1305 roll (2026-08-17 14:43Z) is a NO-OP for this lane, checked
not assumed**: this lane has no inert code awaiting a roll (`b4981634d`
shipped in 1304 and was proven there), and the canary is undisturbed —
`noted.co.uk` still carries `published_hash` from 2026-08-16 16:01:40Z, i.e.
the reconciler has correctly found no drift across the roll. Nothing to
update on that account; the deliverables above are unchanged. (Worth one
line in the NOTES entry so the next session does not re-verify it.)

## Verification (once decisions land)

Each phase keeps its existing acceptance (ZIP: unzip -l count == object count,
presigned expiry proof; handover: single reader gate; editor: cross-tenant
probe). Domain programme acceptance: a real .uk registered end-to-end on the
customer tag, resolving through CF to the served site, retention link
completing a test-mode subscription.
