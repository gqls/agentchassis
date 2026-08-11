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

1. **Copy/FAQ migration on webdesign.uk** — the live site contradicts ruling
   2 and 4 everywhere: hero "£1,200 is the total price", "You see it
   finished … before any money changes hands", "you only pay if you like
   it", and the FAQ answer "What about the domain and hosting?" →
   "We handle the setup" (component `edfecdf2-c25a-4bbd-90c1-c26e644d86cf`,
   `/faq.html`; price copy spans at least index/faq/what-you-get + the chat
   bot's facts — it quoted £1,200 live on 08-10 evening). Blocked on §3
   question 1 (payment timing) — the new copy cannot be written until that
   is answered. Through the framework only; update `evidence_base`; sweep
   retired claims (£1,200, £75 deposit) with a banned-claims-style check —
   correcting the register does NOT arm detection (bugfix 161 lesson).
   **Coordinate with the webdesign_uk_build_service lane** (their site,
   their chat service; check live sessions before touching).
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

## 3. Questions the owner has NOT answered (ask before building the thing they gate)

1. **When is the £149 taken?** At submission (gates the queue, kills spam,
   fits "no refunds") or at preview-acceptance (fits the live "only pay if
   you like it" promise)? Gates work item 1 AND 3.
2. **Are vouchers single-use / expiring / tracked to a person?** Gates the
   voucher table shape.
3. **Does the £1,200 done-for-you tier survive anywhere** (e.g. as the paid
   "we host / we transfer / we do it all" tier), or is £149 the only
   product now? The copy rewrite reads completely differently each way.
4. Queue semantics: per-product queue or per-fleet? What does "roughly how
   long" promise — a count, a date, or nothing binding?

## 4. Owner asks outstanding (external, cheap to nag)

- **Nominet**: the TAG name (username matching the password already
  provided) + add the five fixed cluster IPs to the EPP allowlist:
  134.213.168.26, .37, .44, .54, .56 (office IP rotates — already went
  stale once). Registrar keys (dynadot/porkbun/spaceship): later, per owner.
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
