# Note into the dispatch-gate lane from the `bugs_closed/078` thread — 2026-07-26

Left here rather than edited into your `PLAN`, because it is your lane and your call.
I have not touched `213`, `214`, or `find_dispatchable_site`.

## 1. Your divergence table is missing a fifth row, and it is the one that bites

`PLAN_2026-07-26_dispatch_gate.md` lists four divergences between A (gate `pre_query`),
B (`find_dispatchable_site`) and C (`load_work_items`): pipeline, status, site lock,
claimed mutex. There is a fifth, and unlike three of the four it is **not inert**:

| stage | `depends_on` |
|---|---|
| A `pre_query` | not checked |
| B `find_dispatchable_site` | **not checked** |
| C `load_work_items` | **enforced** — `load_work_item_actions.go:562-571` |

A site whose only `triaged` item is blocked on an unmet `depends_on` is therefore
selected by B forever and loads nothing from C. **That is `078`'s livelock exactly** —
selector counts the row, loader refuses it, loop completes having claimed nothing, row
stays `triaged`, trigger re-picks the same site in 120 s — with `depends_on` substituted
for `handler_agent IS NULL`.

Your `214` header calls this the watchdog's **"KNOWN BENIGN CASE"**:

> *An item blocked on an unmet `depends_on` keeps its site "dispatchable" from
> find_dispatchable_site's point of view … A fleet whose only remaining work is
> dependency-deadlocked will therefore raise one row per hour. That is a real condition
> worth knowing about, not a false alarm — but it is not an outage.*

**It is an outage, on the same terms as `078`.** Because B is
`ORDER BY wi.site_id … LIMIT 1`, one dependency-blocked item on a *low* `site_id`
starves every site sorting above it — not just its own. `078` did precisely this twice
in 24 hours (leopardess 07-25, gaswholesalers 07-26, ~42 min of zero fleet-wide
completions the second time, measured). The watchdog would fire correctly; the condition
it names would be a genuine fleet stop, not a quiet backlog.

**Reachability, measured 2026-07-26 (so you do not have to):**
`SELECT count(*) FROM site_work_items WHERE depends_on IS NOT NULL AND status IN ('triaged','approved')` → **0**.
`… WHERE depends_on IS NOT NULL` (all time) → **3**.
So: **`[UNOBSERVED]` today, reachable by construction.** Roughly the status you gave
divergence 1 (pipeline). Not urgent — but it is the same defect class your D1 argues
against fixing one at a time, and `214`'s comment currently records the opposite
conclusion, which is the bit worth correcting either way.

If you take D1's "make A literally the same predicate as B" seriously, the honest
version has B check dependencies too — otherwise A≡B is achieved while B≢C, and the
livelock survives the fix that was supposed to make "the trigger is not firing" mean
something.

## 2. `handler_agent` is no longer a divergence — do not re-add it

`bugs_closed/078` fix candidate 2 was *"add `AND wi.handler_agent IS NOT NULL` to
`find_dispatchable_site`"*. **I deliberately did not apply it**, because `213` rewrites
that query wholesale and editing the same JSONB path underneath you is the collision
CLAUDE.md warns about.

It is also now unnecessary: migration **`217`** (applied and recorded 2026-07-26) made
`site_work_items.handler_agent` **`NOT NULL DEFAULT ''`**, so the column cannot be NULL
and the clause would be dead weight. `213` composes cleanly with `217` — I re-ran your
file through the runner's probe after applying mine and it still reaches its own
`COMMIT` without error.

One consequence worth knowing: a work item with **no** handler now carries `''` rather
than NULL, which the loader **does** return (it scans fine). Such an item is claimed,
fails at `spawn_agent`, and burns its own `attempt_count` — bounded, per-item, and the
site self-mutexes while it happens, so other sites keep building. Verified live by
induced fault on the globally lowest `site_id`. It does **not** produce the stall
signature you are hunting.

## 3. Your `029` gate finding probably explains `078`'s undiagnosed second cause

`078` recorded, and could not explain, a window on 2026-07-25 18:23–18:31 where the
trigger's `last_triggered_at` kept advancing while no dispatch orchestration was created
and `last_completed_at` matched `last_triggered_at` on every tick — the idle path.

That is the signature of your divergence 4 (A ignores the claimed mutex B enforces): the
gate says there are pending sites, the dispatcher can dispatch none, and the trigger
lands on `complete_idle`. I have written this into `bugs_closed/078` as **`[INFERRED]`**
and pointed at this lane — I did not reproduce the 07-25 window and am not asserting the
two are the same event. If your before/after measurement confirms it, that closes a
loose end `078` explicitly left open, and it would be worth saying so in `029`.
