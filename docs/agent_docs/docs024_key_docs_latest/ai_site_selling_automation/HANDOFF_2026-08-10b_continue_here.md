# HANDOFF 2026-08-10b — continue here: first increment LIVE; next is transcripts ingestion

**Start here cold.** Read `HANDOFF_2026-08-10_start_here.md` for the full
evidence base (note its §3.5 now carries a dated CORRECTION — read it), then
`PLAN_2026-08-10_ai_site_selling_automation.md` for decisions + design, then
the NOTES tail for tonight's verifications. This file is the short bridge.

## 0. State in one paragraph

The owner ruled three things at session start (PLAN §1 — do not re-open):
customer identity lives on `clients → networks → sites` (BIZ-014); the
automation target is the £1,200 tier's INTAKE (a human reviews the rendered
preview and releases); design + ungated build is sanctioned while the
webdesign.uk Phase 6 cutover awaits owner review. On those rulings the first
increment shipped and is **LIVE on roll v1.0.1283, pod-verified both
replicas, both services**: migration 375 (email/phone/tier/customer_status/
notes on `clients`), new `/api/v1/admin/customers` CRUD (ADM-011), and the
admin-dashboard Customers tab. Council submission for the platform-code half:
corr `371f8b7d-0835-4879-b48f-ad0176bf2058` (fired ~20:30, verdict likely
landed by the time you read this — check before doing anything to that code).
The Anthropic account spend cap that stopped the whole fleet 14:51→~20:00
today has LIFTED (chat box answers correctly again).

## 1. Landmines this lane created or confirmed (read before touching)

- **Two client stores.** `/admin/clients` reads `clients_info` (lazily
  created tenant side-table, empty), NEVER the `clients` table. Customers →
  `/admin/customers`. Full entry in LANDMINES.md (synced to doc_notes).
- `fe6b99d05` (the platform commit) predates its council corr and carries no
  trailer — 098 will list it un-reviewed forever; the NOTES entry + verdict
  note are the audit trail. Do not "fix" this with an amend (forward-only).
- The migration runner's `--apply` is blocked by four older pending files
  with pre-state mismatches (353/358/359/361 — other threads') — apply new
  migrations out-of-band (`psql -f` via kubectl) + `--record-only`, as 375 was.
- The admin FE has never typechecked (15 pre-existing tsc errors in App.tsx);
  verification bar is `vite build` in a node:20 container on a COPY.

## 2. Next work, in order (from PLAN §3)

1. **Check the council verdict** for `371f8b7d` (query in NOTES). APPROVED →
   nothing to do (trailer impossible, recorded). REVISE/REJECTED → treat the
   objections as the next task on that code.
2. **Build transcripts → `site_chat_turns`** — fully designed in PLAN §2.3:
   B2 sink (box pushes, sibling lane's turf — hand them the box-side asks
   listed there: per-turn uuid, domain field, log cap rejections), spawn-free
   k8s CronJob puller (precedent: the eight check CronJobs), UUIDv5 pairing
   for idempotency, hash IPs at ingest. Ships with the 046 DDL as the next
   free migration + CHAT-001 register update in the same commit.
3. **Trigger seam (P4)** stays DESIGN-ONLY until BOTH: Phase 6 cutover
   reviewed (owner) AND `bugs_open/239` fixed (owned by its own session —
   check `scripts/who-owns.py 239` + live sessions before touching).
4. Payment gating waits on the owner decisions below.

## 3. Decisions the OWNER still needs to make (handoff §7 minus tonight's three)

1. **Which Stripe surface grows** (§7.2): port idea.uk's proven 2-method
   Provider (real £29 payment, HMAC-verified webhook as sole truth) into the
   buildable repo, or finish `auth-service/subscription`'s half-built columns
   (no SDK, no webhook, `active` = "a row exists"). Related: does live Stripe
   precede or follow automation? (P3's gate — written terms still owed.)
2. **Refund mechanics** (§7.7): manual-dashboard-only refunds (by design, no
   refund code anywhere) — acceptable at automation volume, or bring the
   PAY-003 entitlement seams forward?
3. **Where the chat lives long-term** (§7.5): stays on the VM (proven,
   per-site, facts are a hand-edited Go constant — cannot scale per-customer)
   vs CHAT-002's edge-worker (per-domain at scale). Bears on BIZ-009's
   recursion: every sold site ships its own chat box.
4. **Domain ownership & handover** (§7.6): who registers a customer's real
   domain, in whose account, what transfers on decline/refund. The
   delete-objects un-publish lever only works for domains we control. Nominet
   EPP TAG + three registrar keys still owed (domains rollout lane).
5. **Phase 6 cutover review** (not this lane's decision to make, but its
   gate): the webdesign.uk DNS cutover still awaits owner review in the
   sibling lane; P4 trigger work here is gated on it by the owner's own
   sequencing.

Settled tonight, for the record: identity shape, automation tier, interim
scope (PLAN §1); the spend cap (lifted); §7.4 review policy is settled in
principle by the tier ruling (human reviews the rendered preview — specifics
when the trigger builds).

## 4. Falsifiers

Re-run, don't trust: the council verdict (query, not this file); the cap
(RUNBOOK curl — it can be re-hit any day the fleet spends hard); 239's
ownership (live session listing changes hourly); the sibling lane's handoffs
(they were mid-work tonight — `webdesign_uk_build_service/` newest files win).
