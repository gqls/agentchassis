# HANDOFF 2026-08-11b — continue here: billing LIVE (keyless), copy migration is next — SUPERSEDES HANDOFF_2026-08-11

**Start here cold. Supersedes `HANDOFF_2026-08-11_continue_here.md`** (its §1
rulings compression, §5 landmines and §6 falsifiers all still hold — read them;
its work list is superseded by §2 below). Read order: this file →
`PLAN_2026-08-10_ai_site_selling_automation.md` §1b/§1c (the rulings, do not
re-open) → NOTES tail (08-11 night through post-roll) for the evidence.

## 0. State in one paragraph

Every product decision is ruled (PLAN §1b/§1c). The £149 payment surface
(**PAY-009**) is BUILT and **DEPLOYED KEYLESS** on the 2026-08-11 evening roll:
migration 391's four tables live in clients_db; `internal/auth-service/billing`
(Stripe checkout + HMAC webhook-as-truth on the idea.uk PAY-001 shape, no SDK);
admin routes 401 behind auth, webhook 503s awaiting keys — verified at the pod
(provenance stamp bb5348642, startup mount line, wget probes; NOTES post-roll
entry). The council trail on it stands at four REVISE rounds, **stopped by
design** — three real findings landed (commits b9bea5e1d, 895029d24); do NOT
resubmit unless a new real defect appears (NOTES close-of-session entry, and
the memory lesson `after-the-fact-council-review-of-shipped-code-loops`). The
live webdesign.uk site still sells the retired £1,200 offer — fixing that is
the top work item and was deferred twice only because the sibling lane's live
session is testing rerender locks on that exact site.

## 1. OWNER DECISIONS OPEN (the ask list, in order of when they bite)

1. **Stripe keys — bites when you want to take a first payment.** Create in
   the Stripe dashboard: a **restricted secret key** (Checkout Sessions:Write
   only — the PAY-001 least-privilege shape idea.uk proved) and the **webhook
   signing secret** for endpoint `POST /api/v1/billing/webhooks/stripe`, event
   `checkout.session.completed`. Land both in `personae-platform-secrets` as
   `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET`, then restart auth-service
   (or ride the next roll). Test-mode keys first, per PAY-001: test and live
   are separate accounts with separate webhook secrets.
2. **Webhook public exposure — bites with the keys.** Stripe must reach the
   webhook from the internet; nothing routes into the cluster today. Options:
   (a) **proxy from the webdesign.uk box** over the existing tunnel — the box
   already has public HTTPS; smallest new surface; needs the sibling lane
   (RECOMMENDED as the evaluation default, not yet agreed with them);
   (b) an Ingress + TLS cert on a hostname → ingress-nginx NodePorts
   30080/30443 (controller exists, zero Ingress resources defined — first use
   on this cluster; DNS + cert management falls to us);
   (c) a Cloudflare tunnel into the cluster (new moving part, but no inbound
   ports). Decide when the keys are being made; whoever implements writes it
   up first.
3. **Unify-or-deprecate the old subscription scaffold — bites at first real
   sale.** Two mechanisms now describe "this client paid": the unwired
   `subscriptions` table (auth MySQL; `status=active` means only "a row
   exists" — PAY-007) and the real webhook-verified `billing_orders` path
   (PAY-009). The council's reuse seat flagged the coexistence; recorded
   recommendation: **deprecate the scaffold's create/update surface once the
   first £149 sale has gone through**; any future recurring product builds on
   the PAY-002 plan against clients_db, not the scaffold.
4. **Payment timing switch — yours to flip, no build needed.** Live default is
   `after_approval` (ruled: "while we test"). Flip to `upfront` later via
   `PUT /api/v1/admin/billing/settings {"payment_timing":"upfront"}` (admin
   JWT) or one SQL UPDATE on `billing_settings`.
5. **Standing external asks, unchanged**: Nominet second-tag application
   (pending Nominet); the three registrar API keys (dynadot/porkbun/spaceship,
   later, per you); Phase 6 cutover review (sibling lane — still the gate for
   the trigger seam here).

## 2. Work list (next session, in order)

1. **Copy/FAQ migration on webdesign.uk to the £149 no-frills model** — top
   item, deferred twice for live-collision reasons, not difficulty. The live
   site still says "£1,200 is the total price", "you only pay if you like
   it", and "we handle the setup" (FAQ component
   `edfecdf2-c25a-4bbd-90c1-c26e644d86cf`, `/faq.html`); the chat bot quoted
   £1,200 on 08-10. Before starting: **check the webdesign lane's live
   session state** (their transcript was active 19:04 on 08-11; grep live
   `.jsonl` mtimes — timestamps inside are UTC, mtimes local, see memory) and
   their NOTES for a reply to our 08-11 coordination ask. Then: through the
   framework only; update `evidence_base`; banned-claims-style sweep for
   "£1,200"/"£75 deposit"/"before any money changes hands" (bugfix-161
   lesson: correcting the source does not arm detection); re-verify the chat
   box survived (their `pages.sections` landmine); includes the hosting setup
   page (UK S3-compatible + CDN walkthrough) and the affiliate block
   (Lovable, Durable, …). The £1,200 copy is archived in
   `snapshot_2026-08-11_gbp1200_offer/` if ever needed.
2. **Queue + submission gate** (PLAN §2.6, designed): derived occupancy from
   open work items, owner-settable capacity (3–4) + `queue_paused`, gate at
   every intake door, non-binding wait note. Platform code → council +
   register (submit BEFORE the shipping commit — see the memory lesson).
3. **Admin FE voucher screen** (handoff item 3b): issue/list vouchers, read
   orders, flip payment timing — against `/api/v1/admin/billing/*` on
   auth-service directly (NOT the core-manager proxy); PipelinesPage bolt-on
   precedent. First real voucher POST doubles as post-roll recipe step 4.
4. **ZIP delivery** (work item 4 of the old list, unchanged): pull the
   completed site's B2 prefix, zip, store as a deliverable asset, link it.
   Pairs with the what-you-get copy being explicit.
5. **Transcripts → `site_chat_turns`** (PLAN §2.3, designed) — unchanged.
6. **Trigger seam (P4)** — still design-only, BUT: `bugs_open/239` gained a
   fail-closed fix commit from the sibling lane (`a097e3e26`, 08-11). Check
   whether it is fixed-AND-LIVE (their bug file + a pod provenance check)
   before treating dispatch as trustworthy; the Phase 6 cutover review
   remains the other gate.

## 3. What is live vs inert (verified post-roll, 08-11 evening)

- **LIVE**: migration 391's four tables (clients_db); the billing routes on
  auth-service (admin 401s, webhook 503s); the payment-timing switch
  (`after_approval`); migration 375's customer columns; ADM-011 customers
  API + tab (since v1.0.1283).
- **INERT / OWED**: any actual payment (keyless by design); post-roll recipe
  **step 4** (real voucher POST — needs an admin JWT; first dashboard use
  covers it; **recipe gotcha: core-manager's image has wget, not curl** —
  RUNBOOK examples use curl, swap accordingly); admin FE voucher screen;
  webhook public exposure.

## 4. Landmines for this work (beyond 08-11 §5, all still holding)

- **The council gate loops on shipped-code submissions** — fresh panels
  re-gate on unverifiable-from-tier every round (4× REVISE here). Submit
  BEFORE or ALONGSIDE the shipping commit; cap after-the-fact reviews at ~2
  rounds once real findings are landed. Full story: NOTES close-of-session +
  the memory lesson.
- **`2>/dev/null` on psql/kubectl turned two real errors into silent empty
  results this session** (a wrong column name read as "no ledger row"; a
  missing curl read as "no response"). Let stderr through on any check whose
  empty result you might believe.
- The 097 trigger piped through `tail`/`grep` buffers until EOF — run it to a
  file when backgrounding (RUNBOOK).
- kcat delivery failure vs queue latency: the RUNBOOK's "when a missing row
  IS a dropped dispatch" section — kafka broker 0 was 0/1 NotReady on 08-11.
- Voucher variants are HARD-CODED to £10/£55 (`RuledVoucherPences`) per the
  ruling — a new variant is a ruling first, then a one-line change.
- `billing_orders.status` has NO refunded state on purpose (owner ruling:
  refunds manual, unadvertised). Do not add one in passing.

## 5. Falsifiers (re-check before trusting this file)

A newer handoff in this directory; the sibling lane's session state (live
right now?) and any reply in their NOTES; `bugs_open/239`'s fixed-AND-LIVE
state; the chat bot's current price line (one wget, RUNBOOK); whether the
Stripe keys have appeared in `personae-platform-secrets`
(`kubectl -n ai-persona-system get secret personae-platform-secrets -o
jsonpath='{.data}' | python3 -c "import json,sys;
print(list(json.load(sys.stdin)))"`).
