# HANDOFF — webdesign.uk build service, continue here (2026-07-31)

**COLD-START ENTRY POINT.** Written so a genuinely fresh Claude Code session —
zero context, new terminal — can pick this workstream up without re-reading the
whole PLAN first. Prose and detail live in the other four standing docs:
`PLAN_2026-07-28_webdesign_uk_build_service.md` (the design, all rulings),
`NOTES_webdesign_uk_build_service.md` (evidence + every misstep),
`RUNBOOK_webdesign_uk_build_service.md` (the commands), `README_where_we_are.md`
(the owner's plain-prose log, append-only — read the tail).

---

## 1. Read this before touching P4 — the sequencing problem

**You were pointed here to work on P4 ("automate the trigger"). P4 cannot be
meaningfully built yet, and this is not a formality — it is missing its subject.**

P4 is *"the pull action + a scheduled task"* that fetches **paid orders** from
the webdesign.uk box. There is no box. There is no order storage. There is no
Stripe integration, test or live. **P1, P2 and P3 — all of which come before P4
in this plan's own phasing (§9) — have not been started.** Nothing has been
built for this product yet except:

- the plan itself, fully ruled on (§11);
- `platform/fetchguard`, a genuinely reusable SSRF guard, shipped 2026-07-31
  and already wired into an unrelated live bug (`bugs_open/159`) it found along
  the way — real, but not part of *this* product's own build;
- one live Fable-5 probe (§7c) confirming org retention and a floor cost.

So: **the honest next task is P1** (the shopfront — minimal page, the LLM chat,
Stripe test mode, orders on the box; §9 of the PLAN, resized by the §5.1
ruling to include the chat's spend/abuse controls from day one, not as later
polish). If you were told specifically to start P4 ahead of P1 — for instance,
because the owner wants the intake API *contract* settled early, before the box
exists — say so explicitly in your own opening note and treat §3 below as a
**design-only** exercise: work out the shape, do not expect to submit it to the
council with real evidence behind it, because there will be nothing running on
the other end to point at. Submitting a seam with no real consumer to measure
against is exactly the kind of unmeasured claim CLAUDE.md's council-submission
rules ("measure the blast-radius claim yourself; do not ask the reviewer to")
are designed to catch — you would not be able to satisfy that requirement
honestly until P1 exists.

**If in doubt: build P1 first.** It is a working business with zero platform
code (P1–P3 all are, per §9), it is what proves anyone wants this at all before
more is spent on it, and it is what P4 needs to exist before it can be real.

## 2. State, as of 2026-07-31

| phase | state |
|---|---|
| P0 — decide and measure | **DONE**, bar one item that cannot close until later phases exist (below) |
| P1 — shopfront, fake door | **NOT STARTED** |
| P2 — free teaser | **NOT STARTED** |
| P3 — manual fulfilment, real money | **NOT STARTED** |
| **P4 — automate the trigger** | **NOT STARTED — this handoff's nominal subject, blocked on P1** |
| P5 — automate the seeding | NOT STARTED |
| P6 — isolation decision executed | NOT STARTED |
| P7 — ownership/handover | NOT STARTED |

**P0's one open item:** measuring one *real* site build end to end via
`llm_call_log`. Structurally the same problem as P4 — there is no build
pipeline wired to this product yet to measure. It closes naturally once P4/P5
exist, not before.

**Ownership check run before writing this handoff** (`git log`, `who-owns.py`,
`site_work_items`) — **all clean, nobody else is working this lane.** Every
commit on this directory is from one prior session. Re-run the same three
checks yourself before you start, per CLAUDE.md — this file goes stale within
hours on a shared tree.

## 3. What P4 actually is, when its time comes

> **⚠ SUPERSEDED IN PART, same day — read
> `PLAN_2026-07-31_p4_order_intake.md` before this section.** P4 was planned
> out on 2026-07-31 and turned out to be **far smaller than this section
> implies**: `build_queue`, `seed_build_queue` and a live
> `build-pipeline-trigger` (firing every 120s, right now) already provide the
> whole queue → site → work-item → build chain. P4 reduces to **one action +
> one scheduled task that put a paid order into `build_queue` as one row**.
> That plan also corrects two of this workstream's own stale premises and
> flags the one cheap test that must pass before any of it is trusted
> (`build_queue`'s last real use was 2026-03-22). The list below is kept
> because its *obligations* (§4) and *landmines* (§5) are unchanged.

From PLAN §4.1 (the trust-boundary ruling, already decided — do not re-litigate
it): the webdesign.uk box holds paid orders; **the cluster polls outbound to
collect them.** The box never dials in, never holds a cluster credential. This
is the one architectural property everything else in this workstream is built
to protect — see §4.1's full reasoning before touching it.

> **⚠ ADDED 2026-07-31, before any P4 planning got further than grounding —
> read this first, it changes the shape of item 1 below.** The obvious starting
> point is the **`http_request` action, already in the registry**, categorised
> `external`, described as *"Make an HTTP request to an external endpoint"*.
> **It is a STUB.** `HTTPRequestAction` (`generic_actions.go:130-155`) reads
> `url`/`method`, makes **no network call whatsoever**, and returns a hardcoded
> `{"status": 200, "body": {"success": true, "data": "mock response"}}` — its
> own comment says *"For now, return mock response"*. A P4 built on it would
> report healthy runs forever while collecting nothing. **0 live agent
> definitions reference it** (checked), so nothing is fooled today.
> **Consequence for P4: item 1 below is a genuinely new action, or the honest
> implementation of that stub** — and if you implement the stub, that is a
> shared-mechanism change touching every future caller, so it carries the full
> §4 obligations rather than being a quiet fill-in. Either way it must use
> `fetchguard.NewClient` (§5), not a bare `&http.Client{}`. Also in
> `LANDMINES.md`, footprint `http_request`.

So P4 is, concretely:

1. **An action** (new, in `platform/orchestration/actions/`) that calls the
   box's own API (built in P1/P3) over HTTPS, asks for orders in a `paid`
   state it has not yet collected, and marks them collected. **See the warning
   above — do not assume `http_request` does this.**
2. **A scheduled task** (`scheduled_tasks` row) that fires this action on some
   interval — start conservative (hourly?) rather than tight, since nothing
   about "next day or so" fulfilment needs low latency.
3. **What happens to a collected order** is P5's job (seed a `sites` row +
   `site_specs` aspects, oufe-style — see PLAN §7a for what "productised, not
   bespoke" implies about how much briefing detail this needs), so P4's own
   scope is genuinely just get-the-order-into-the-cluster, not build-the-site.

## 4. The obligations that make P4 different from an ordinary fix

**This is a shared mechanism, so it is architecture-scope even though it is
additive** — CLAUDE.md's "Platform seams and the ordering exemption" section,
owner rulings 2026-07-28 and 2026-07-29. Concretely, when you build it:

- **Register it in the concept register in the SAME commit that ships it** —
  not later. Name its landmine and its open review question explicitly (the
  proposal's own register entries, e.g. `DBI-025` added for `fetchguard` this
  week, are the house style to copy).
- **Measure the blast-radius claim yourself before submitting to council.**
  "No collision is possible" is a query, not an argument (`bugs_closed/124`'s
  own lesson, cited in CLAUDE.md). For P4 specifically: how many other things
  poll or write to `scheduled_tasks`/this action's target table, and could a
  second poller double-collect an order? Check it, cite the query, don't ask
  the reviewer to.
- **The owner's 2026-07-29 ruling retired the "ordering constraint" condition**
  — you do not need to hold this out of the fleet before review; HEAD is
  shared and any session's roll ships your commit regardless. What you do
  still owe: registration in the same commit, and submitting to the council
  before or alongside it (`097_TRIGGER_council_review_v1.sh`).
- **Tell other consumers, don't just measure them.** If this action or its
  scheduled task shares infrastructure with anything else already dispatching
  work (the existing `build-dispatch-loop`, `build-pipeline-trigger` — see
  `agent_definitions` for what already exists before inventing a new dispatch
  path), name them in the submission and say what changes about their
  guarantee, not just that you checked nothing breaks.

## 5. Landmines specific to this lane (fuller list: `LANDMINES.md`, footprint-searchable)

- **A "per-IP" limiter behind Cloudflare/Caddy is usually one global bucket** —
  `bugs_open/139`'s finding, directly relevant to P1's chat/teaser spend
  controls. Key on `CF-Connecting-IP`, verify with `count(DISTINCT ip) > 1`
  from two networks, not one test machine.
- **`platform/httpguard` is inbound-only; `platform/fetchguard` (new, shipped
  this week) is the outbound/SSRF one.** Any new fetch this product adds — the
  teaser reading a customer's domain, above all — uses `fetchguard.NewClient`,
  not a bare `&http.Client{}`.
- **A model swap is not config-only if the call layer passes rejected params.**
  Checked clean for Fable 5 as of 2026-07-31 (PLAN §7b) — re-check if the call
  layer (`platform/aiservice/anthropic.go`) has changed since.
- **`agent_definitions_backup` keeps the SOURCE row's `id` and `created_at`** —
  order any snapshot lookup by `snapshot_taken_at`, not `created_at`, or a
  config diff lies about what changed.

## 6. Owner rulings already made — do not re-ask these

Full detail in PLAN §11 and §7a/§7b. In brief: real LLM chat (not a stepped
form) for the intake box; briefing questionnaire stays **optional**, re-open
that decision only at P5; trust boundary is pull-only from day one, isolation
method decided at P3 not before; full sites at a high quality-based price;
guarantee + fee model both hinge on **acceptance** (before = refund only,
preview comes down; after = customer's changes paid, **our defects free**);
builds run on `claude-fable-5`; preview host is a different, shorter domain,
still to be supplied by the owner; the "about a thousand sites" figure in
webdesign.co.uk's copy is accepted as forward-looking, not to be re-litigated
here.

## 7. Immediate next step

Confirm with the owner whether this thread is building **P1** (recommended,
per §1 above) or genuinely wants P4's design work done ahead of it. Either way,
re-run the three ownership checks in §2 first — this file is a snapshot.
