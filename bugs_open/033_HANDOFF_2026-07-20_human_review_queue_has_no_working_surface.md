# 033 — the human-review queue has no working surface: 292 items, none ever actioned through it

**Filed:** 2026-07-20 by the reasoning-dataset thread.
**Severity:** latent, accumulating. Nothing errors. No site reports a failure.
Items route to `needs_human_review` correctly and then stop existing as far as
the platform is concerned.
**Status:** OPEN. Needs an **owner decision on intent** before any code — see
§"The question that has to be answered first".

---

## Observed

```
site_work_items WHERE status='needs_human_review' : 292
  oldest                                          : 2026-03-15  (4 months)
  newest                                          : 2026-07-20  (today)
  ever carrying approved_by                       : 0
  ever resolved via the admin API                 : 0
```

Arrival rate is **increasing**, not draining:

| month | items entering `needs_human_review` |
|---|---|
| 2026-03 | 4 |
| 2026-04 | 33 |
| 2026-05 | 31 |
| 2026-06 | 8 |
| **2026-07** | **216** (47 of them `cta_names_unknown_destination`) |

## The surface that exists, and has never been used

Three routes are registered for actioning a review item
(`internal/core-manager/api/server.go:210-219`):

| handler | file:line | writes |
|---|---|---|
| `HandleRetryWorkItem` | `site_admin_handlers.go:719` | `status='triaged'`, resets `attempt_count`. No identity. |
| `HandleResolveWorkItem` | `site_admin_handlers.go:774` | `status='complete'`, `result = jsonb_build_object('resolution',$2,'resolved_by','admin')` |
| `HandleApproveWorkItem` | `site_admin_handlers.go:817` | `status='complete'`, `result` with `'approved_by','admin'` |

**None has ever run.** No row anywhere carries `result->>'resolved_by' = 'admin'`
or a non-NULL `approved_by` column. A fourth handler, `HandleConfirmWorkItem`
(`internal/core-manager/admin/confirm_work_item_handler.go:42`), is **fully
implemented** — transactional, creates a follow-up item, marks the review item
complete — and is **never registered in `server.go`** (grep returns exactly one
hit: its own definition). It is unreachable code.

Also dead: the `approved_by` and `resolution_path` **columns**
(`sql_for_tables/018_site_work_items.sql:51,37`) are written by no Go code
anywhere in the repo. `resolution_path` appears in no `.go`, `.ts`, `.tsx` or
`.js` file at all.

## What actually resolves items — and why it matters

Items *do* get resolved. **Eight** carry a real `result->>'resolution'`, and they
are good:

> *"Cancelled 2026-07-14 before dispatch: the item bundled a FALSE boots failure
> (stale `.tool-container` anchor, fixed by PLAN supersede 148) with a GENUINE
> mobile-overflow failure that is NOT the tool's (vonc site footer,
> `div.footer-legal` — routed to component-template-fixer as a `responsive_fix`).
> Dispatching it would have sent tool-improver chasing a stale contract."*

Every one was written by a **working Claude thread via direct SQL**, not through
the API — all eight have an empty `resolved_by`, which the API would have stamped
`'admin'`. Their statuses are `cancelled` (7) and `complete` (1
`section_source_drift`); **none is a `needs_human_review` item**.

So the honest picture is not "nobody resolves anything". It is:

- the **intended** surface (admin API) is unused and partly unwired;
- the **de facto** surface (a thread writing SQL) works, produces genuinely good
  reasoning, records no identity, and is invoked ~8 times in four months;
- and the `needs_human_review` queue specifically has **never** been drained by
  either.

> **Correction to an earlier claim by this thread.** In
> `reasoning_dataset/PLAN_capture_gaps_and_volume.md` I wrote that the resolution
> JSONB was empty "0 of 4,570". That query was scoped to `status='complete'` and
> missed the seven `cancelled` rows. The reasons are being captured — rarely, ad
> hoc, and without identity. Corrected there too.

## The question that has to be answered first

**Is `needs_human_review` meant to be a queue, or a bin?**

- **If a queue** — it needs a surface someone actually opens, and the four
  handlers plus two dead columns are most of the work already done (one of them
  just needs registering). 292 items is a backlog, and the July spike says the
  producers are getting more productive while the consumer does not exist.
- **If a bin** — i.e. "park this, it is not worth a human's time" — then the
  status is misnamed, the items should expire, and the checks routing 216 items a
  month into it should be re-tuned instead. `cta_names_unknown_destination` alone
  is 47 in July; that looks like a check firing into a void.

**Nothing should be built until that is decided.** Wiring up the surface would be
wasted work if the answer is "bin", and re-tuning the producers would be wrong if
the answer is "queue".

## Fix candidates (after the decision, not before)

**If queue:**
1. Register `HandleConfirmWorkItem` in `server.go` — it is written and tested-shaped
   already. One line. (Check it still matches the current schema first; it has
   never run.)
2. Write the real identity, not the literal `'admin'`. Note there is **no auth
   context in these handlers** — no user id, claims or subject is available
   (`grep` for `userID|claims|c.Get(` in `site_admin_handlers.go` returns
   nothing), so "record who decided" is blocked on an auth decision and is not a
   one-liner. Say so rather than shipping another hardcoded `'admin'`.
3. Populate the `approved_by` and `resolution_path` columns, or drop them. Two
   dead columns that look authoritative are worse than none.

**If bin:** add an expiry/`wont_fix` sweep, and open a separate item against
whichever checks are the top producers.

## Why a dataset thread noticed

Human overrides are the highest-quality label obtainable — the only ones
expressing *preference* rather than mere success — and the eight that exist are
excellent. That is our interest and we declare it. But it is not the argument:
292 items, four months old, arriving faster than ever and read by nobody is a
platform problem whether or not anyone ever trains on it.

## Related

- `bugs_open/017` / `work_item_completion_integrity` — the completion end of the
  same lifecycle.
- `bugs_open/032`, `bugs_open/021` — verification coverage; same family of
  "the mechanism exists but almost nothing uses it".
- `reasoning_dataset/PLAN_capture_gaps_and_volume.md` §Gap 2 — the fuller
  write-up, including the corrected figures.
