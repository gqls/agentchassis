# HANDOFF 2026-08-11 — continue here: four rulings landed, £149 queue model — SUPERSEDES 08-10b

**Start here cold. Supersedes `HANDOFF_2026-08-10b_continue_here.md`** (its §0
state and §1 landmines still hold; its §3 decision list is now ANSWERED).
Read order: this file → `PLAN_2026-08-10_ai_site_selling_automation.md` §1/§1b
(the rulings) → `HANDOFF_2026-08-10_start_here.md` for deep evidence (§3.5
carries a correction — read it) → NOTES tail for verifications.

## 0. State in one paragraph

The Customers surface is LIVE end to end and council-APPROVED (round 1,
unanimous; corr `371f8b7d-0835-4879-b48f-ad0176bf2058`): migration 375
customer columns on `clients`, `/api/v1/admin/customers` CRUD (ADM-011), and
the admin-dashboard Customers tab — all pod-verified on roll v1.0.1283. The
Anthropic spend-cap outage of 08-10 is over. The owner has now ruled the four
open decisions (PLAN §1b, authoritative). What follows is the work those
rulings create, roughly in order.

## 1. The rulings, compressed (PLAN §1b is the full text — do not re-open)

1. **Subscription service**: build out `auth-service/subscription` (already
   in the build; has stripe_customer_id/stripe_subscription_id columns, no
   SDK, no webhook, `status=active` means "a row exists" —
   `stripe/001commentary.md:44-46`). Set it up even if unused at first.
   Vouchers belong here.
2. **£149 all-in per site — supersedes £1,200/£75-deposit entirely.**
   Visible queue, 3–4 slots to start, rough wait-time note, submissions
   CLOSE when full. Voucher codes the owner hands out: £10 all-in and £55
   variants. No refunds. One set of changes (later agent-driven).
3. **Chat into the framework**: framework-originated (seeded, built,
   rendered like everything else), may still be deployed on the box. Bot
   knowledge comes from the framework's actual capabilities — retires the
   hand-maintained `systemPromptFacts` constant (`box/chat-service/chat.go:26-47`).
   Later: chat sells palette/layout/logo-upload services independent of a build.
4. **Customer domains: theirs.** Deliverable = private preview + downloadable
   ZIP they host themselves. Paid optional extras: we host, or manual domain
   transfer; else recommend a third-party host. Explicit up front about what
   they get. "We run a nameserver" idea is open-not-ruled; owner will not
   risk breaking customer email — NS delegation takes ALL records (MX
   included), so evaluate instead: (a) hand them two records to add, or
   (b) Cloudflare zone onboarding, which imports existing records pre-cutover.

## 2. Work list this creates (proposed order, cheapest-real-first)

0. ~~Pin the £1,200 copy first~~ — **DONE 2026-08-11**:
   `snapshot_2026-08-11_gbp1200_offer/` (sites/pages/page_components/
   site_specs JSONL, counts verified 1/6/22/31; see its README for what it
   does and does not hold, and why restore is deliberate, not one-click).
1. **Copy/FAQ migration on webdesign.uk** — the live site contradicts the
   rulings everywhere: hero "£1,200 is the total price", "You see it
   finished … before any money changes hands", "you only pay if you like
   it", and the FAQ answer "What about the domain and hosting?" →
   "We handle the setup" (component `edfecdf2-c25a-4bbd-90c1-c26e644d86cf`,
   `/faq.html`; price copy spans at least index/faq/what-you-get + the chat
   bot's facts — it quoted £1,200 live on 08-10 evening). NOW UNBLOCKED —
   payment timing is ruled (§3.1); write to the no-frills-honest direction
   (§3.3), including the hosting/DNS story (ZIP download, third-party
   hosting setup page, optional paid extras). Through the framework only;
   update `evidence_base`; sweep retired claims (£1,200, £75 deposit) with
   a banned-claims-style check — correcting the register does NOT arm
   detection (bugfix 161 lesson). **Coordinate with the
   webdesign_uk_build_service lane** (their site, their chat service; check
   live sessions before touching). Includes the new **hosting setup page**
   (UK S3-compatible + CDN walkthrough) and the affiliate-links block
   (Lovable, Durable, …) — affiliate programme sign-ups are an owner task.
2. **Queue + submission gate design**: where the 3–4-slot queue lives
   (candidate: count open build work items for this product; surface the
   count + wait note on the site and in chat), what closes submissions,
   what reopens them. Small, but it is the demand limiter the £1,200-era
   "human fulfilment IS the limiter" assumption no longer provides.
3. **Subscription service + vouchers**: real Stripe SDK-or-raw-HTTP client,
   webhook handler with signature verify as sole source of truth (the
   proven idea.uk PATTERN — pattern, not code port), wired to
   `clients.external_id`; voucher table (code, discount-to amount, ?single-
   use, redeemed_by/at) validated at submission. Council + register entries
   — this is platform code and a new shared mechanism.
4. **ZIP delivery**: a completed site's objects live in B2 under
   `<host>/<path>` keys (worker §3.3 of start-here) — a "download your
   site" step is: pull prefix, zip, store as a deliverable asset, link it.
   Small; pairs with the what-you-get copy being explicit.
5. **Transcripts → `site_chat_turns`** (designed, PLAN §2.3) — unchanged,
   still next platform increment; now also feeds the queue/wait-time note
   and demand evidence for the £149 experiment.
6. **Chat-into-framework rebuild** (ruling 3) — the biggest piece; overlaps
   the sibling lane's Phase 5 chat service. Propose: new framework
   component + capability-sourced facts, deployed to the box; the
   hand-written service retires. Needs its own design pass + council run.
7. **Trigger seam (P4)** — still design-only until the Phase 6 cutover is
   owner-reviewed AND `bugs_open/239` is fixed (owned elsewhere; re-check
   `who-owns` + live sessions).

## 3. Questions — ANSWERED 2026-08-11 (second ruling batch, PLAN §1c) except Q4

1. ~~When is the £149 taken?~~ **After approval while we test the system,
   moving to up-front later — build payment timing as a switch.** Refunds
   exist behind the scenes only (manual dashboard), never offered visibly.
2. ~~Voucher shape?~~ **Single-use, nameable to the recipient, with an
   expiry** (code, drops-price-to £10/£55, recipient, expires_at,
   redeemed_at/by).
3. ~~Does £1,200 survive?~~ **No — off the table entirely** (owner: no time
   to make full, good websites right now). Its complete copy is ARCHIVED in
   `snapshot_2026-08-11_gbp1200_offer/` (this directory; row counts
   verified) for possible later revival. Copy direction for the rewrite:
   **no-frills, Ryanair-honest** — you get what you pay for, and we say so,
   including that the sites are AI-built.
4. ~~Queue semantics~~ **ANSWERED (owner, 2026-08-11 evening): the wait note
   is an approximation, NOTHING binding** — the queue may pause on software
   malfunction or scale trouble. Queue copy must promise nothing; a pause
   state is a first-class state, not an error.

New direction attached to the rulings: **hosting goes to affiliates /
third-parties** — recommend a UK-based S3-compatible store + Cloudflare (or a
UK equivalent; research which) and write customers a setup page; those who
don't want our offering get (affiliate) links to Lovable, Durable, etc.
**Differentiation comes from the example sites built from the owner's own
domains** — the portfolio is the sales proof.

## 4. Owner asks outstanding (external, cheap to nag)

- **Nominet**: ~~TAG name~~ **provided 2026-08-11: `DESIGNCONSULT`** —
  credentials file can be completed (domains lane NOTES has the
  contribution). **STILL OWED: the five cluster IPs on the EPP allowlist**
  (134.213.168.26/.37/.44/.54/.56) — the owner-supplied live list
  (5.65.164.9, 116.203.204.115, 151.226.83.138, 176.58.121.95) contains
  none of them, so cluster EPP is still blocked; suggest dropping the
  rotating office IP 151.226.83.138 when adding. Owner is applying for a
  SECOND tag for this venture (Channel Partner-shaped — only one
  Self-Managed tag is allowed per registrar); draft application body was
  provided in-session 2026-08-11. Registrar keys: later, per owner.
- Phase 6 cutover review (sibling lane) — still the gate for P4 here.

## 5. Landmines for this work (beyond 08-10b §1, which all still hold)

- `/admin/clients` reads `clients_info`, never lists customers → use
  `/admin/customers` (LANDMINES.md entry, synced).
- The chat service's `facts[]` is bookkeeping; `writer_block` is the wire
  (sibling lane's landmine) — relevant to any interim price fix in the bot.
- `pages.sections` is not durable — a rebuild silently removed the chat box
  once; any copy migration must re-verify the chat box survived.
- webdesign.uk is `github_repo='vm-sites'`, deploy target `vm` — its deploy
  path is NOT the B2 worker path; don't assume the ugg2 delivery shape.
- Migration runner `--apply` still blocked by other threads' broken pending
  files — out-of-band apply + `--record-only`, as 375 was.

## 6. Falsifiers

Re-check before trusting: live sessions on the sibling lane and on 239; the
queue of pending migrations (it changes); the chat bot's current price line
(one curl, RUNBOOK); whether a newer handoff than this exists in this
directory.
