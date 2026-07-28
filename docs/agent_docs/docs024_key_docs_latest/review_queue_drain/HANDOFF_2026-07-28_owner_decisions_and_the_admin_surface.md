# HANDOFF 2026-07-28 — the owner has taken the 083 decisions; the admin surface is the way he makes them

**Written by:** the brochure_component_library thread (fundamentallyai.com), which
arrived at this workstream sideways and is handing it back. That thread stays on
fundamentallyai.com; **this work belongs here**, in `review_queue_drain/`, which already
owns `bugs_open/033` and the drain. **Do not open a new workstream** — I nearly did, and
checking first is why this file is in this directory.

**Read first:** `PLAN_2026-07-25_review_queue_drain.md` (the two prior owner rulings —
*"split it: auto-drain what can be, queue the rest"* and *"this queue is a queue, not a
bin"*), then `README_where_we_are.md`, then `NOTES_review_queue_drain.md` from the bottom.

---

## 1. What the owner decided, 2026-07-28

He was shown the classification below and said **"I will take all of them"** — i.e. he
takes every decision himself, working through them in the web admin dashboard. He asked
for the admin API to be checked first, which is done (§3).

The decisions he is taking, from `bugs_open/083`:

| # | decision | recommendation given |
|---|---|---|
| A | the **50** items that need a human ANSWER | promote — put them where he reads them |
| B | the **186** advisory / machine-fixable items | demote — a periodic report, not a queue |
| C | the **5** `capability_gap` rows | leave — they are deliberate roadmap rows |
| D | the structural rule | an item type must name who reads it; **"nobody" must be an answerable and refusable answer** |

**He has not yet executed these.** Nothing in this handoff should be actioned as though
he had. What he wants is the *surface* to work through them on.

> **CORRECTED 2026-07-28 (dashboard session): decision A has no execution step — the 50
> are already at `needs_human_review`, on the screen he reads. See the correction in §4.**

## 2. The numbers he is deciding against, measured 2026-07-27/28

```sql
SELECT status, count(*), count(DISTINCT item_type), min(created_at)::date
  FROM site_work_items WHERE status IN ('needs_human_review','detected') GROUP BY 1;
--  needs_human_review | 325 | 23 | 2026-03-15
--  detected           | 157 | 23 | 2026-07-14
```

> **The correction that reframed this bug, and it is the important line in this file.**
> `bugs_open/083` was being read as *unroutable findings need a route* — 298 non-terminal
> items across 13 types with no handler, plus 9 naming `human-review`, which is not a
> registered agent. I was one query from recommending we surface those 298 on the
> dashboard, having found its handler only accepts `status='needs_human_review'`
> (`internal/core-manager/admin/confirm_work_item_handler.go:79`) — *"one status value
> from a human's screen"*.
>
> **There are already 325 items at that status, oldest 2026-03-15.** The queue humans
> *can* see is fuller and older than the one they cannot. **Routing is not the
> bottleneck; readership is.** A router would move rows between three unread lists.
> (Note this is a *lower* number than the 370 the PLAN measured on 07-25 — the drain
> work has moved it, so re-measure rather than quoting either.)

The 298 unroutable, by what would actually resolve them:

| kind | items | types |
|---|---|---|
| **needs a human ANSWER** | **50** | `needs_section_data`, `owned_page_review`, `incomplete_page_group` |
| advisory / machine-fixable in principle | 186 | `unresolved_cta`, `cta_names_unknown_destination`, `required_fields_missing`, `voice_tells`, `dead_control`, `image_source_unsatisfiable`, `image_url_404`, `contact_form_undeliverable`, `truncated_component` |
| deliberate roadmap row | 5 | `capability_gap` |

**The 50 are blocked work, not backlog.** Seventeen are leopardessconsulting.co.uk
sections waiting since **2026-03-15** for information only the owner can supply —
*"Section 'pricing' on engagement-model needs: Pricing tier names…"*. The platform has
been asking a question for four and a half months and it was never put to anyone.

**Do not quote "298 broken things".** `capability_gap` is *supposed* to have no handler.
Which of the other twelve types are deliberate is undecided work.

## 3. The admin surface — checked 2026-07-28, and the one thing that is missing

**Working:**

```
core-manager /health                        -> 200
GET /api/v1/admin/work-items  (no token)    -> 401 {"error":"Authorization header required"}
admin-dashboard :8080                       -> 200
```

The 401 is the useful result: route registered, service alive, auth enforced. Both
deployments are `1/1`, two replicas each.

**The full HITL API already exists** (`internal/core-manager/api/server.go:210`):

```
POST   /api/v1/admin/work-items
GET    /api/v1/admin/work-items
GET    /api/v1/admin/work-items/:item_id
PATCH  /api/v1/admin/work-items/:item_id
POST   /api/v1/admin/work-items/:item_id/retry
POST   /api/v1/admin/work-items/:item_id/resolve
POST   /api/v1/admin/work-items/:item_id/approve
```

**The dashboard is same-origin**: its nginx proxies `/api/v1/` → core-manager and
`/api/v1/auth/` → auth-service. So **one port-forward gives a fully working dashboard,
login included** — no separate API tunnel:

```bash
kubectl port-forward -n ai-persona-system svc/admin-dashboard 8080:8080
# then http://localhost:8080
```

**What is missing: there are no ingresses at all, in any namespace.** `admin-dashboard`
is `ClusterIP`; the only NodePorts are `ingress-nginx` itself and `wireguard`
(51820/UDP, created 2026-07-20 — the VPN route the PLAN's visibility half mentions).
So the dashboard is **not reachable from a browser without either the VPN or a
port-forward**. That is the gap between "the admin panel works" (it does) and "the owner
can sit down and work through 50 decisions" (he cannot yet, conveniently).

**Kubeconfig token expires every 3 days** — a fleet-wide `Unauthorized` means expiry,
not breakage, and only the owner can refresh it. Budget for that before a long session.

## 4. What this workstream should do next, in order

1. **Confirm the owner's access route works end to end** — him, his browser, a real
   login, one item opened. Not a `curl` from inside the cluster; that is what I did and
   it proves the service, not the journey. Decide with him: port-forward, VPN, or an
   ingress worth creating.
2. **Make the 50 findable.** They are `status='detected'` with no handler, so the
   dashboard cannot see them at all (it filters on `needs_human_review`). Promoting them
   is decision A and is his to make — but the *filter* may be the honest fix rather than
   rewriting 50 rows' status, and that is a design question for this thread.

   > **CORRECTED 2026-07-28 (dashboard session) — the status claim above is wrong, and
   > it changes this step from "build something" to "already done".** Measured today:
   > all 50 sit at **`status='needs_human_review'`** (42 `needs_section_data` + 6
   > `owned_page_review` + 2 `incomplete_page_group`), all in the **build** pipeline —
   > the dashboard's default view, reachable via the existing Needs Review filter plus
   > the server-built item-type dropdown, with the full action set including the
   > auto-built `needs_section_data` form. Nothing needs promoting or filtering; there
   > is no design question left in this step. The `detected` population is a
   > *different* set (157 rows, 24 types, all but `image_url_404` carrying a live
   > `handler_agent`) — its defect is the dead promoter,
   > `bugs_open/083_HANDOFF_2026-07-26_detected_findings_never_reach_a_handler.md`,
   > not readership. What caught the error: one `GROUP BY status` over the three item
   > types. Full evidence in `NOTES_review_queue_drain.md` (2026-07-28). The 083 file
   > itself is accurate — the error was introduced here, in the paraphrase.
3. **Decision B** — the 186. A report needs a reader and a cadence; a queue that nobody
   drains is what this whole bug is about, so do not build a second one.
4. **Decision D is the one that closes the door.** An item type that names no reader
   should not be writable. That is a schema/write-path change, not a dashboard change,
   and it is the only candidate here that prevents the next instance.

## 5. Landmines

- **`complete` is not proof.** Verify against the row and the rendered surface, never the
  status. This queue is full of items whose status has never meant what it says.
- **A queue with no drain rate is a claim about intent the data contradicts.** The owner
  ruled on 2026-07-25 that *"the queue should not FILL"*; it is at 325 and the oldest row
  predates that ruling by four months.
- **`human-review` is not a registered agent.** 9 items across 5 types name it as their
  handler and are unclaimable — `bugs_closed/077`'s exact mechanism. Do not add a tenth.
- **Re-measure before quoting any number here.** The PLAN said 370 on 07-25, this file
  says 325 on 07-28. Both were true when taken.
- **This thread's own near-miss, recorded because it would have cost a day:** I found a
  neat mechanism ("one status value from the screen"), and it was wrong. The check that
  killed it was counting the queue it would have routed into. *Before proposing a route,
  measure the destination.*
