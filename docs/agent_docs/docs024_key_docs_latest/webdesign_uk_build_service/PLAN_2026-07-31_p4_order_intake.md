# PLAN — P4: getting a paid order into the build pipeline

**Written 2026-07-31.** Design only — nothing built. Grounds P4 (§9 of
`PLAN_2026-07-28_webdesign_uk_build_service.md`) against what the platform
actually has, rather than against what that plan assumed it would need.

**The headline: P4 is roughly a tenth of the size the original plan implied,
and P5 is perhaps half.** The queue → site → work-item → build chain already
exists and is running in production every two minutes. P4 is not "build a
dispatch pipeline"; it is **get one row into one existing table**.

---

## 1. What already exists — measured 2026-07-31, not assumed

| thing | state |
|---|---|
| `build_queue` table | **EXISTS** — `domain` (UNIQUE), `direction` (jsonb), `status`, `priority`, `batch_id` |
| `seed_build_queue` action | **EXISTS** — reads `status='queued'` rows, upserts the `sites` row, derives the first work item from `direction`, inserts it, marks the row `seeded` |
| `build-pipeline-trigger` scheduled task | **LIVE AND FIRING** — every 120s, `enabled=t`; last fired 2026-07-31 12:35:51Z, completed 12:35:51Z |
| downstream dispatch | `build-dispatch-loop` via `spawn_agent`/`call_agent`, already wired in that agent's workflow |

**So the pipeline is idle, not absent.** It polls every two minutes and finds
nothing, because nothing writes to `build_queue`: **2 rows in its entire
history, both `seeded`, the most recent 2026-03-22** — over four months ago.

`direction` is the contract that decides what happens to a queued domain
(`seed_build_queue_action.go:239-295`, `seedDetermineFirstItem`):

| `direction` | first work item | handler |
|---|---|---|
| `null` | `needs_domain_research` | `domain-research-classifier` |
| `{"objective": "..."}` | `needs_domain_research` (+ objective in spec) | `domain-research-classifier` |
| `{"brief_complete": true}` | `needs_site_plan` (skips research/briefing) | `site-planner` |
| `{"adopt_from": "..."}` | `needs_site_adoption` | `site-adoption-agent` |
| `{"fork_from": "..."}` | `needs_site_plan` | `site-planner` |

**That table is most of P5.** "Brief → the pipeline does the right thing" is
already expressed as a data contract; P5's remaining question is what to *put*
in `direction` for a paying customer, not how to make the pipeline obey it.

## 2. What P4 therefore is

**One action + one scheduled task, and nothing else.**

1. **A new action** — call it `collect_external_orders` — which:
   - GETs the webdesign.uk box's orders endpoint (built in P1/P3) over HTTPS,
     asking for orders in a `paid`, not-yet-collected state;
   - for each, `INSERT INTO build_queue (domain, direction, priority)`;
   - acknowledges collection back to the box so it stops offering them.
2. **A `scheduled_tasks` row** firing it. Interval: start at **900s (15 min)**,
   not the 120s the build trigger uses — fulfilment is "next day or so" (PLAN
   §7a), so latency is worth nothing here and a slower poll is a smaller
   blast radius while this is new.

Everything after the `INSERT` is existing, running code.

**It must NOT reach for the `http_request` action.** That is a registered stub
that returns a hardcoded `{"status":200,...,"data":"mock response"}` and makes
no network call — see `LANDMINES.md`, footprint `http_request`. The new action
does its own HTTP, and it must use `fetchguard.NewClient` (shipped 2026-07-31,
register `DBI-025`) rather than a bare `&http.Client{}` — the box is our own
host, but a compromised or misconfigured one redirecting inward is exactly the
case `fetchguard` exists for, and using it here costs nothing.

## 3. Two of this workstream's own premises were out of date — corrected

Both were carried into `PLAN_2026-07-28` §5.3a from the oufe seed's preamble.
Both were true when written and are not now. **Checked at source today**, and
the correction matters because §5.3a's fabrication argument rests on them.

**(a) `evidence_base` absence no longer disables the whole claims layer.**
`validate_page_content.go:302-338` now splits the two halves deliberately
(`bugs_open/104`):

> *"banned claims are FLEET-WIDE and scan with or without a site register, so a
> site nobody has armed — and every new site on its first build — is still
> protected against the universal shapes; the numeric scan stays strictly
> opt-in on the register's presence."*

So a `build_queue`-created site with no `site_specs` **does** get banned-claim
protection. What it does **not** get is `checkUnregisteredNumbers` — and that
is precisely the check that catches invented figures *about the customer's own
business* ("serving 500 clients since 2009"). **§5.3a's evidence_base control
is therefore still needed, for a narrower and more precise reason than the
plan states.**

**(b) `bugs_open/063` is closed — the email check is now fail-CLOSED.**
It is `bugs_closed/063`, and `validate_page_content.go:980` now emits
*"Email '%s' asserted but the site has no registered contact address — no
email may be published"*. A site with no registered email no longer passes a
fabricated one; it rejects **any** published email. That is materially
stronger than §5.3a assumed, and it means §5.3a's control (1) — "emit no
contact block unless supplied" — is **already enforced at the platform level**
rather than being something this product must build.

> **Method note.** Both corrections came from checking a cited claim at its
> source instead of re-quoting the doc that carried it. Two superseded
> premises in one afternoon's research is the rate this repo's own
> `prior-art-search-goes-stale` and `a-closed-bugs-scope-out-expires` entries
> would predict — and both had been repeated in this workstream's plan as
> settled fact.

## 4. The real gaps P4/P5 must close

With (a) and (b) corrected, the honest list is shorter and sharper:

1. **`seed_build_queue` creates no `site_specs` aspects at all** (grepped:
   no `site_specs`, `aspect` or `evidence_base` anywhere in that file), and
   **`upsertSite` inserts only `(domain, name, network_id, status)`** — no
   email, no company name. So a customer's site arrives with **no register**,
   hence no numeric-claim scan, and no contact details for the platform to
   verify a published address against.
2. **That is P5's actual job**, restated precisely: not "make the pipeline
   accept a brief" (it already does, via `direction`), but **write the
   customer's questionnaire answers into `sites` (email, phone, company_name)
   and into a seeded `evidence_base` aspect** before the first page is built,
   so the honesty guards this product's whole pitch depends on are armed.
3. **`build_queue.domain` is UNIQUE**, which cuts both ways:
   - **Good:** a natural idempotency key. A re-collected order (network retry,
     lost ack) can `ON CONFLICT (domain) DO NOTHING` and cannot double-build.
   - **Bad:** a second legitimate order for the same domain — a re-order after
     a refund, or a customer wanting a rebuild — is silently swallowed by the
     same clause. **Decide this deliberately** rather than discovering it: most
     likely the collector should reject-and-alert on a conflicting domain whose
     previous row is `seeded`, not silently drop it.

## 5. Verify before trusting any of §1

**`build_queue` last did real work on 2026-03-22.** "Exists and once ran" is
not "works today" — this codebase changes ~1,500 commits/week, and the two
historical rows predate essentially all of the current pipeline. Before P4 is
built on this assumption, **put one test row through it end to end**:

```sql
INSERT INTO build_queue (domain, direction, priority)
VALUES ('<a throwaway domain>', '{"objective":"..."}'::jsonb, 100);
```

…then watch `build-pipeline-trigger` (fires within 120s) and confirm a `sites`
row and a `site_work_items` row actually appear. If they do not, P4's premise
is wrong and this plan needs rewriting before any code is committed. **That
single test is the cheapest thing in this document and the most load-bearing.**

## 6. Obligations (unchanged from the handoff, restated so they travel)

A new action in the shared registry is an **architecture-scope shared
mechanism** (CLAUDE.md, owner rulings 2026-07-28 / 2026-07-29), so:

- **Register it in the concept register in the same commit that ships it**,
  with its landmine and open review question written down.
- **Measure the blast-radius claim yourself before submitting to council.**
  For this action specifically: does anything else write `build_queue`, and can
  two pollers double-collect? Both are queries — run them, cite them.
- **Name and tell the other consumers.** `build-pipeline-trigger` and
  `seed_build_queue` are existing machinery this feeds; say what changes about
  their guarantee (they go from never-firing-usefully to processing real,
  paid customer work).
- Submit to the council before or alongside the commit; the ordering-exemption
  condition was retired, so do not claim a constraint you do not have.

## 7. Open for the owner

1. **Poll interval** — 15 min proposed. Anything faster buys nothing given
   next-day fulfilment; anything slower delays the human's review window.
2. **Repeat domains** (§4.3) — reject-and-alert, or allow a rebuild path?
   This needs a product answer before the collector's conflict clause is
   written, because the two produce different customer-visible behaviour.
3. **Does P4 wait for P1 after all?** This plan can be *written* now (it is),
   but it cannot be *built* honestly until there is a box with an orders
   endpoint to poll — there is nothing to point a council submission's
   evidence at. Recommend building P1 first and keeping this as the settled
   contract it can be built against.

## Related

- `PLAN_2026-07-28_webdesign_uk_build_service.md` — §4.1 (pull-only trust
  boundary, already ruled), §7a (the offer), §9 (phasing), §5.3a (the
  fabrication controls this plan corrects two premises of).
- `HANDOFF_2026-07-31_continue_here.md` — cold-start entry point.
- `LANDMINES.md` — `http_request` (the stub), `platform/httpguard` (inbound
  only; `fetchguard` is the outbound one).
- `seed_build_queue_action.go`, `site_db_actions.go:1016` (`upsertSite`),
  `validate_page_content.go:302-338` and `:980`.
