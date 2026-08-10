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

### 2.1 Client DB — extend `clients` [DONE 2026-08-10: migration 375 LIVE]

Shipped first session: `email`, `phone`, `tier`, `customer_status`, `notes` —
all nullable; Stripe linkage stays on the existing `external_id`.
`customer_status` is deliberately NOT `status` (`sites.status` is load-bearing
for tools-api CORS; `subscription.status` already means "a row exists").
Blast radius measured before applying: zero Go readers of `FROM clients`,
zero positional INSERTs. Applied out-of-band + `--record-only` (older pending
files from other threads have pre-state mismatches and block the runner's
`--apply`). Commit `fe6b99d05`. Council submission owed — gate down
fleet-wide at ship time (Anthropic account cap, see NOTES).

### 2.2 Admin "Customers" surface [DONE 2026-08-10, inert until rolls]

**The handoff's "FE-only work" premise was false** — the existing `/clients`
endpoints read the `clients_info` TENANT store, not the ruled chain (see the
dated correction in the handoff §3.5, and LANDMINES.md). Shipped instead:
new `/api/v1/admin/customers` CRUD in core-manager backed by
`clients → networks → sites` (list + site counts, detail + sites, create,
partial PATCH), registered as **ADM-011**, plus the FE Customers tab on the
`PipelinesPage` bolt-on precedent (`fe6b99d05`, `a84d544d1`; vite build
verified in a container). Go half inert until the next core-manager roll; FE
half until the next admin-dashboard image build. The old `/clients` endpoints
are untouched — different population, do not merge without an owner ruling.

### 2.3 Chat transcripts → `site_chat_turns` [DESIGNED 2026-08-10, BUILD NEXT]

Finish the `046_site_chat_turns.sql` design, not invent one. Shape ruled by
SAAS-001: **core pulls from a sink**, one-directional, egress-from-core-only;
"the chat service POSTs into the cluster" will not be built. Design, from
this session's code-level scout of producer vs table vs precedents:

- **Sink = B2.** The box already has an outbound-only posture and the VM plan
  already commits to nightly dump→encrypt→push of exactly these files to the
  backups bucket (`PLAN_2026-08-04…:177`). Add a transcripts export push
  (box-side, sibling lane's turf — an ask, not our edit). Core never dials
  the box; today it structurally cannot (scout: no code path, tunnel-only).
- **Puller = a spawn-free k8s CronJob**, one stateless binary: read new B2
  objects → pair turns → upsert. NOT a scheduled_tasks→Kafka→chassis action:
  `bugs_open/239` makes dispatch untrustworthy and `bugs_open/240` punishes
  per-job topics — the same reasons IMP-053 was built spawn-free. Precedent:
  the eight existing check CronJobs under `deployments/kustomize/services/`.
  The isolated-chat plan's own words: "No Kafka, no chassis."
- **Pairing/idempotency.** Producer writes one JSONL line per *message*
  (`TranscriptEntry`: timestamp, conversation_id, raw client_ip, role,
  content); the table is one row per *turn* with an edge-supplied uuid PK.
  Until the producer supplies a turn uuid, derive one: UUIDv5 over
  (conversation_id, user-line timestamp) — deterministic, so re-ingest stays
  idempotent via `ON CONFLICT (id) DO NOTHING`.
- **Gaps to handle at ingest:** producer stores RAW IPs → hash with salt at
  ingest, never store raw (table comment is explicit, GDPR); no site_id or
  domain in the feed → puller config maps feed→`sites` row (single-tenant
  feed today); Claude-failure turns have a user line but no assistant line →
  insert with empty answer + `error_message` correlated from
  `requests.jsonl`; turn-cap/spend-cap rejections log NOTHING box-side today.
- **Box-side asks for the sibling lane** (contribute into their NOTES; do not
  edit their service): per-turn uuid, a domain field, log cap rejections
  (maps to `capped`), optionally hash the IP at source.
- **Ships with:** the 046 DDL as the next free migration (FK to `sites`,
  clients_db) in the same commit as the puller + a CHAT-001 register update.

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
| standing five + rulings recorded | — | DONE 08-10 (`409c40a8d`) |
| clients columns migration (2.1) | — | DONE 08-10, LIVE (`fe6b99d05`; gate submission owed) |
| admin Customers API + tab (2.2) | image rolls | BUILT 08-10 (`fe6b99d05`, `a84d544d1`) |
| transcripts ingestion (2.3) | build next | DESIGNED 08-10 (see §2.3) |
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
