# Persistence & data flow — storing idea.uk requests and results

You want requests and results in a database, eventually the main framework DB,
but not via a route that opens a security hole. This is the design. The short
version: **your instinct is correct — an internet-facing service should never
hold credentials to, or have a network path into, the core database.** Data
crosses that boundary one-way, through a drop the trusted side ingests — the
same presigned-B2 mechanism the Thunder adapter already uses. The framework DB
ends up with everything; the exposed box never touches it.

---

## 1. The security crux (why not connect directly)

idea.uk is the most exposed surface you have: public endpoints, untrusted input,
payment webhooks, sitting on a VM outside the cluster. The core Postgres holds
the whole platform. If idea.uk held core-DB credentials (or the cluster firewall
let the box dial Postgres), then a compromise of idea.uk — the likeliest thing
to be compromised — would hand the attacker a route into everything.

So the rule, standard tiered/DMZ thinking:

- **Exposed tier (idea.uk)** only ever *writes* to a scoped external drop.
- **Trusted tier (the chassis)** *reads* from the drop and ingests into Postgres.
- Initiative and DB credentials live entirely on the trusted side.

Even total compromise of idea.uk then yields no database access — just the
ability to write junk into one storage prefix, which the ingest step validates.

---

## 2. The design — three tiers, one-way flow

```
┌─────────────────────────────┐
│  idea.uk box (exposed)      │
│                             │
│  engine + service           │
│   ├── LOCAL STORE           │  operational state: orders, paid/unpaid,
│   │   (SQLite or JSON)      │  Stripe idempotency, report pending review.
│   │                         │  Low-latency, what a request reads/writes live.
│   └── on each terminal event│
│       writes an immutable   │
│       record ──────────────▶│  Backblaze B2  (the DEAD-DROP)
└─────────────────────────────┘   bucket/prefix: idea-events/
                                   scoped WRITE-ONLY key (or presigned PUT)
                                            │
                                   (no inbound path to the cluster — the box
                                    pushes out to B2; nothing reaches in)
                                            │
                                            ▼
┌──────────────────────────────────────────────────────────────┐
│  Framework (trusted, in-cluster)                               │
│                                                                │
│  scheduled_tasks row ──fires──▶ idea-ingest agent              │
│     every ~300s                   reads new B2 records,        │
│                                   INSERTs into Postgres        │
│                                   (idempotent on order id)     │
│                                            │                   │
│                                            ▼                   │
│                                   Postgres: idea_orders        │  ← system of
│                                   (business_intel or public)   │     record
└──────────────────────────────────────────────────────────────┘
```

Two stores, two jobs:
- **Local store** = operational. The service reads/writes it during a request's
  lifecycle. Also a durable buffer until records are shipped to B2.
- **Framework Postgres** = system of record. Queryable, integrated, where
  analytics and any cross-platform joins live. Fed one-way via B2.

---

## 3. The local store on the box — SQLite or stay JSON

Today it's a JSON file (`store.go`, `IDEA_DB_PATH`). For a single small instance
that's adequate, but you asked for a database. The honest trade-off:

| | JSON (today) | SQLite (recommended) | Local Postgres |
|---|---|---|---|
| Real ACID / concurrent-safe | mutex only | yes | yes |
| SQL queryable on the box | no | yes | yes |
| Server to run/secure on the exposed box | none | none (in-process) | yes — extra surface |
| New dependency | none | one (pure-Go driver) | one + a server |
| Backup | copy the file | copy the file | pg_dump |

**Recommendation: SQLite**, via the pure-Go driver `modernc.org/sqlite` (no cgo,
so the static binary still builds). It's a real database — ACID, SQL — with no
server to secure and no network surface, and the whole thing is one file you
copy to B2 nightly. `store.go` already has a small interface, so swapping the
JSON implementation for a SQLite one is contained.

**The one honest cost to flag:** this adds your *first* third-party dependency,
breaking the current "stdlib-only, `GOPROXY=off` just works" property. You'd
`go mod vendor` once with the network on, then offline builds use the vendored
copy. If preserving stdlib-purity matters more than local queryability, **stay on
JSON** — the box's local store doesn't have to be fancy, because the *database*
you'll actually query is the framework Postgres (§5). Either local choice works
with the rest of this design; the dead-drop and ingest are identical.

My lean: if you mainly want durability + the data in a real DB, **stay JSON on
the box and put the database in the framework** (§5) — it keeps the exposed tier
dead-simple (no driver, no DB to secure) and is the cleaner separation. Upgrade
the box to SQLite only if you specifically want to run SQL *on the box*.

---

## 4. The channel — why B2, and the alternatives

| Channel | Inbound path to cluster? | Latency | Fit |
|---|---|---|---|
| **B2 dead-drop** (recommended) | **none** | minutes (polling) | reuses Thunder's presigned-B2 pattern; most secure |
| Kafka topic | yes — broker reachable from the box (SASL/TLS) | seconds | append-only, topic-scoped; better than raw PG but still a path in |
| Narrow HTTPS ingest API | yes — one cluster endpoint exposed | seconds | one authenticated append-only endpoint; box pushes |
| Direct Postgres connection | yes — full DB reachable | real-time | **the hole. No.** |

**B2 wins** for you because it needs *no inbound path to the cluster at all* and
it's the pattern you already run (`prepare_artefact_url` in the Thunder adapter).
The box writes each record to `idea-events/` with a B2 application key **scoped
to write that one bucket/prefix** — or, tighter, the trusted side issues a
short-lived presigned PUT per write so the box holds no standing credential at
all (exactly `prepare_artefact_url`). Eventual (polling) consistency is fine for
orders/results; you're not joining on them in real time.

If you later want near-real-time (e.g. a live ops dashboard), Kafka is the
upgrade — but start with B2; it's the smallest secure thing.

---

## 5. The framework table (proposed — to reconcile with the live schema)

Your schema conventions (from the `\d` dump): `public` and a `business_intel`
schema owned by a separate restricted role `clients_user`; uuid PKs via
`gen_random_uuid()`; `jsonb` for flexible data; `timestamptz` defaulting `now()`;
snake_case; `status` text columns. `business_intel` already holds the
collected/analytics data (`businesses`, `business_prices`, `data_observations`,
`discovery_candidates`, …).

**Where it belongs:** idea.uk orders are product/analytics data, and
`business_intel` is the analytics tier owned by the restricted `clients_user`
role — which doubles as a *second* security boundary (the ingest can use that
limited role, not a superuser). So `business_intel.idea_orders` is the natural
home. If you'd rather treat it as operational, `public` works too. Your call;
it's a one-line schema-qualifier difference.

Proposed shape — **this is a starting point, NOT final DDL; per your rule we
verify against the live schema first** (exact role grants, whether to FK
`sites(id)` since idea.uk may be registered as a site, naming):

```sql
-- PROPOSED. Reconcile with live schema before creating.
CREATE TABLE business_intel.idea_orders (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_ref       text NOT NULL UNIQUE,        -- the box's order id (idempotency key for ingest)
    site_id         uuid,                          -- optional FK → public.sites(id) if idea.uk is a site
    domain          text NOT NULL,                 -- the customer's business/domain analysed
    audience        text,                          -- stated audience
    notes           text,                          -- optional "anything specific"
    status          text NOT NULL DEFAULT 'requested',  -- requested|confirmed|paid|delivered|refunded
    stripe_session  text,                          -- Stripe checkout/session ref (no card data)
    report_md       text,                          -- the delivered report (markdown)
    result_meta     jsonb DEFAULT '{}'::jsonb,     -- model/effort/cost/risk-summary, etc.
    requested_at    timestamptz NOT NULL,
    delivered_at    timestamptz,
    ingested_at     timestamptz NOT NULL DEFAULT now(),
    raw             jsonb DEFAULT '{}'::jsonb       -- the full record as received from B2 (audit)
);
```

`order_ref` is the idempotency key: the ingest does
`INSERT … ON CONFLICT (order_ref) DO UPDATE …` so re-reading a B2 record (or a
status progressing requested→paid→delivered) updates the same row instead of
duplicating. No card data is ever stored — only Stripe's opaque session/payment
references.

Optional second table, `business_intel.idea_taster_events` — lightweight rows
for the free audience-check (no payment), valuable as a demand/conversion signal
(how many tasters → how many £29 orders). Worth it if you want that analytics;
skip if it's noise.

---

## 6. The ingest job (reuses your existing machinery)

No new infrastructure — this is the `scheduled_tasks` + agent + action pattern
you already have:

1. A `scheduled_tasks` row (`interval_seconds` ~300) targeting a small
   `idea-ingest` agent.
2. The agent lists new objects under `idea-events/` in B2, reads each, and
   upserts into `business_intel.idea_orders` (idempotent on `order_ref`), then
   marks the B2 object processed (move to `idea-events/ingested/` or tag it).
3. Runs as the restricted `clients_user` role — it only needs INSERT/UPDATE on
   that one table.

The chassis *pulls*; the box never connects in. If ingest is down, records pile
up safely in B2 and get caught up on the next run (replayable).

---

## 7. Security properties (the recap)

- The exposed box holds **no core-DB credentials** and has **no network path**
  into the cluster. Worst case on compromise: write junk to one B2 prefix, which
  the ingest validates.
- The drop is **one-way**; the trusted side has the initiative.
- Ingest runs as a **restricted role** on a **single table** — least privilege
  even within the trusted tier.
- **Decoupled**: idea.uk's uptime and storage tech are independent of the
  cluster; the framework can rebuild its view by replaying B2.
- **No card data** anywhere in our store — only Stripe's opaque references.

---

## 8. Phased — you can do this in two steps

- **Phase 1 (now, on the box):** keep the local store (JSON is fine; SQLite if
  you want local SQL), and add B2 record-writing on each terminal event. Data is
  now durable and *off the box*. Nothing in the cluster yet, no hole opened.
- **Phase 2 (when ready, in the framework):** create the table (reconciled with
  the live schema) and the `idea-ingest` scheduled task. Data becomes queryable
  in the framework DB. The box doesn't change again.

This matches what you said — it doesn't have to go directly into the framework
DB; it gets there safely, on your timetable, without the exposed service ever
reaching in.

---

## 9. Open decisions for you

1. **Local store:** stay JSON (simplest, recommended) or upgrade to SQLite
   (local SQL, one vendored dependency)?
2. **Schema home:** `business_intel` (analytics tier, restricted role —
   recommended) or `public` (operational)?
3. **Taster events:** store them too (conversion analytics) or orders only?
4. **Channel latency:** B2 polling (recommended) now; Kafka later only if you
   need near-real-time.

Tell me 1–3 and I'll (a) write the box-side B2 record-writing, and (b) draft the
finalised DDL + the `idea-ingest` agent/scheduled-task — after a proper look at
the live schema, per the usual rule.

---

## 10. DECISIONS LOCKED (2026-06-04) — supersedes the menu in §3–§5

1. **Local store: JSON** (no SQLite). The box stays stdlib-only;
   `GOPROXY=off` keeps working. The local store remains the operational buffer.
2. **Schema home: a new `ecommerce` schema** (not `public`, not
   `business_intel`) for the commerce records, and the **idea reports live in
   `clients_db`**.
3. **Store taster events** as a conversion signal.
4. **B2 polling now, Kafka later.**

### Refined table layout

Commerce/funnel data → new `ecommerce` schema:

- `ecommerce.orders` — the transaction record. `id uuid pk`, `order_ref text
  unique` (idempotency key), `status` (requested|confirmed|paid|delivered|
  refunded), `stripe_session` (opaque ref, **no card data**), `amount_gbp`,
  `domain`, `audience`, `notes`, `requested_at`, `paid_at`, `delivered_at`,
  `ingested_at`, `raw jsonb`. **No report content here** — commerce only.
- `ecommerce.taster_events` — the free audience-check, as a funnel signal.
  `id uuid pk`, `event_ref text unique`, `domain`, `audience`,
  `verdict_summary text` (short), `created_at`, `ingested_at`, `raw jsonb`.
  Lets you measure tasters → paid conversion.

The deliverable content → `clients_db`:

- `idea_reports` — `id uuid pk`, `order_ref text` (links to `ecommerce.orders`
  **by value**, see the fork below), `domain`, `report_md text`,
  `candidates jsonb`, `model_meta jsonb`, `created_at`, `ingested_at`.

### ⚠ One fork to confirm: is `clients_db` a database or a schema?

This materially changes the implementation:

- **If `clients_db` is a separate Postgres _database_** (the `_db` suffix and the
  existing `clients_user` role suggest a client-facing DB may already exist):
  Postgres **cannot** foreign-key or join across databases. So
  `ecommerce.orders` and `clients_db…idea_reports` are linked **by the shared
  `order_ref` value**, not by a DB-enforced FK. Any cross-set view needs an
  app-level lookup or `postgres_fdw`. This is a perfectly fine design — it cleanly
  separates *commerce* from *client deliverable* (different DBs, possibly
  different roles/backups) — it just means no referential integrity between the
  two, and the ingest agent writes to two connections.
- **If you meant `clients_db` as a _schema_** in the same database as
  `ecommerce`: then `ecommerce.orders` ↔ `clients_db.idea_reports` is a normal
  cross-schema FK and joins work as usual.

**My read:** you most likely mean a separate database (the naming + the role).
That's clean; I'll design the ingest to write the commerce record to
`ecommerce.*` and the report to `clients_db` keyed by the same `order_ref`, with
no cross-DB FK. **Confirm database-vs-schema and I'll finalise the DDL
accordingly.**

### Ingest with two destinations

The `idea-ingest` scheduled task reads each B2 record and writes:
- the order/taster row → `ecommerce.*` (upsert on `order_ref`/`event_ref`),
- the report (for paid, delivered orders) → `clients_db` (upsert on `order_ref`).

Both idempotent. If `clients_db` is a separate database the agent holds two
restricted connections (one per DB), each least-privilege on its own
schema/table. Still no path from the exposed box to either — the box only writes
to B2.

### Kafka later (the planned upgrade)

When near-real-time is wanted, the box publishes the same record shape to a
Kafka topic instead of (or alongside) B2, and a consumer does the same two
upserts. The record schema is identical, so it's a transport swap, not a
redesign — design the B2 record now as if it were a Kafka message (a
self-contained event with `order_ref`, `type`, `status`, payload) and the Kafka
move is cheap.
