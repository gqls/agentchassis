# PLAN — 2026-08-10 — AI-site-selling automation (webdesign.uk next stage)

This is the working plan for the lane briefed in
`HANDOFF_2026-08-10_start_here.md` (same directory). That handoff is the
evidence base — what is live, what is absent, with proofs. This file holds the
decisions and the phasing; it does not restate the evidence. The two source
plans it builds on are in `../webdesign_uk_build_service/`
(`PLAN_2026-07-28_…` §9 business phases; `PLAN_2026-08-04_…` VM/chat + trust
classes).

## 1. Owner rulings — 2026-08-10, this session (do not re-open)

Put to the owner directly at session start; three answers, all "recommended"
options taken:

1. **Customer identity lives on `clients` (BIZ-014's shape).** Reuse the live
   `clients → networks → sites` FK chain; customer-shaped columns are ADDED to
   `clients` (email, phone, tier, status, Stripe mapping via the existing
   `external_id`). `sites` gains **no** ownership columns; ownership flows
   through the chain. This retires §7.1 of the handoff: the `owner_id`-on-sites
   TL;DR shape and the abandoned ADM-008 junction both LOSE.
   *Reason:* the admin API already speaks `/clients`; `external_id` is the
   documented Stripe key; least new machinery.
2. **The automation target is the £1,200 done-for-you tier's INTAKE.**
   Automate chat → seeded specs → triggered build → preview link; a human
   (the owner) reviews the rendered preview and releases. The £19 self-serve
   tier remains future direction, not current scope. This retires §7.3.
   *Consequence:* review machinery survives — the design keeps a
   `checkpoint_for_review`-shaped gate before anything customer-visible goes
   live, and the payment gate can stay simple (deposit before build is a
   policy question for P-payments, not a blocker for intake automation).
3. **Scope while the webdesign.uk Phase 6 cutover is pending owner review:
   design + ungated build.** Build what touches neither webdesign.uk nor the
   auto-trigger: the admin Customers tab (endpoints exist), the client-DB
   columns (per ruling 1), and the transcripts→`site_chat_turns` ingestion
   design. Trigger-seam (P4) work stays design-only for now.

Inherited settled rulings (handoff §6) are inherited, not re-argued here:
trust classes (customer deliverables to B2 via the worker, never on a box),
framework-only builds, static-first + skip-list, isolation deferred to P3 with
its named trigger, pricing (£1,200 / £75 deposit / 2 rounds / 14 days),
P0–P7 sequencing.

## 2. Design

### 2.1 Client DB — extend `clients` [PROPOSED, DDL to be written after schema read]

Columns to add (all nullable, no default-behaviour change for the two
placeholder rows): contact email + phone, `tier` (text; 'done_for_you' now),
`customer_status` (distinct name — `sites.status` is load-bearing for CORS,
`subscription.status` already means "a row exists"; do not overload a third
`status`), and Stripe linkage **via the existing `external_id`** (documented
natural key) — no new stripe column unless a second id is genuinely needed.
Trap owned: the base DDL for clients/networks/sites lives only in
`docs/_archive/…/sql_for_tables/002_…`, NOT in `platform/database/migrations/`
— so the ALTER migration must be self-contained (IF NOT EXISTS guards) and
must not assume a prior migration created the table. Migration-runner practice
applies (dry-run this session, scope the dir).
Council: `platform/database/migrations/` is council scope. The gate is DOWN
fleet-wide as of 14:51Z today (Anthropic account cap — see NOTES); submission
happens when it can survive, and the commit that ships the migration says so.

### 2.2 Admin "Customers" tab [BUILDABLE NOW]

FE-only, against the live `/api/v1/admin` client endpoints
(`POST /clients`, `GET /clients`, `GET /clients/:client_id/usage`). Mount as
one more `view ===` branch per the `PipelinesPage.tsx` precedent. Out of
council scope (`frontends/`). First cut shows what the API already returns;
new columns from 2.1 flow in later.

### 2.3 Chat transcripts → `site_chat_turns` [DESIGN NOW, BUILD NEXT]

Finish the `046_site_chat_turns.sql` design, not invent one. Shape is ruled by
SAAS-001: **core pulls from the box**, one-directional, egress-from-core-only;
"the chat service POSTs into the cluster" is a non-conforming shape and will
not be built. Turns the JSONL dead-end into queryable demand data and is the
rehearsal for the trigger seam. Detail lands here after this session's scout
report (producer format vs table DDL vs available pull mechanism).

### 2.4 Trigger seam (P4) [DESIGN ONLY until cutover reviewed + 239 fixed]

Intake record → `082`-shaped Kafka envelope with a real `client_id` replacing
`demo_client`; every new field added to every hop's `input_mapping`
(allow-list, not passthrough — the migration-274 lesson). Architecture-scope:
council run + concept-register entry in the shipping commit.
**Hard blocker: `bugs_open/239`** (dispatch non-deterministically no-ops) —
being worked by a live session as of tonight; this lane does not touch it,
only tracks it. Nothing auto-triggered is trustworthy until it is fixed.

### 2.5 Seeding automation (P5) — later

Chat brief → `site_specs` aspects. ONB-019 (`build-briefing-agent`) is the
existing autonomous half. Not designed in this pass.

## 3. Sequencing, with gates named

| step | gated on | state |
|---|---|---|
| standing five + rulings recorded | — | this commit |
| admin Customers tab (2.2) | — | next |
| clients columns migration (2.1) | council reachable (advisory) | DDL drafted, ship when gate can run |
| transcripts ingestion (2.3) | design ready | design this session |
| trigger seam (2.4) | Phase 6 cutover reviewed AND 239 fixed | design only |
| seeding (2.5) | trigger seam | not started |
| payment gating | §7.2/§7.7 owner decisions (deferred) | not started |
| customer go-live | IMP-053 seated (site reachability) | platform-side, tracked |

## 4. Standing constraints this lane must not violate

- Isolation boundary: egress-from-core-only (SAAS-001). No box→core writes.
- Do not touch the `webdesign_uk_build_service` lane's box, contact page, or
  Phase 6 cutover — that lane has its own live session. Contribute into their
  NOTES if this lane finds something theirs.
- Customer sites are static/Tier-1 by construction; the P3 isolation trigger
  ("first paid build that scrapes a domain we do not own") transfers to this
  lane intact.
- `sites.status='deployed'` drives tools-api CORS; `sites.status` DDL default
  is `'active'` — neither gets overloaded with customer lifecycle meaning.
