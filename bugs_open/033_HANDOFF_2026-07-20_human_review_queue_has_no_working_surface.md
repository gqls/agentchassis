# 033 — the human-review queue has no working surface: 292 items, none ever actioned through it

**Filed:** 2026-07-20 by the reasoning-dataset thread.
**Severity:** latent, accumulating. Nothing errors. No site reports a failure.
Items route to `needs_human_review` correctly and then stop existing as far as
the platform is concerned.
**Status:** OPEN. Needs an **owner decision on intent** before any code — see
§"The question that has to be answered first".

> **READ §"GROUNDED 2026-07-20" AT THE FOOT OF THIS FILE FIRST.** Everything above
> was re-checked against the live system that day and three load-bearing claims
> failed, including the title. A full working review surface **does** exist; it is
> hardcoded to show the newest 50 items and therefore displays **0 of the 208**
> build-pipeline review items, reporting the queue as empty. Most of the fix needs
> no owner ruling. The original text is left intact below as filed.

> **UPDATE 2026-07-20 (bugfix thread) — the three display bugs are FIXED & LIVE;
> the queue is now visible. Bug stays OPEN for the drain.**
> - **Visibility (bugs 1–3): DONE & DEPLOYED, chassis + dashboard `v1.0.1141`.**
>   `HandleListWorkItems` is paged (limit/offset, default 200) and returns
>   server-side `status_counts`/`type_counts`/`total`/`truncated`; the dashboard
>   filters server-side, shows "showing N of M" + Load more, and gained a pipeline
>   selector (build/content/all). Verified: the handler's exact SQL returns 208
>   `needs_human_review` live (was 0 in the window); unit tests pass and the
>   count-scoping regression test fails on the reintroduced bug; the new symbols
>   are in the running pod. NOT exercised: the auth-wrapped HTTP call end-to-end
>   (no admin token to hand) — the owner sees it on first connect. Commit
>   `c11a804bd`.
> - **Access: DONE & LIVE.** Owner chose a VPN over a public Ingress (2026-07-20).
>   WireGuard deployed (`06d860c6b` + swept `deployment.yaml`); pod Running, `wg0`
>   up with 2 peers, and from inside the tunnel's network the dashboard resolves
>   via kube-dns and returns HTTP 200. External UDP handshake unverified (needs the
>   owner's client). Also fixed the two broken makefile targets (`7c969599c`).
> - **Owner decision taken:** the ~175 deliberate escalations → **split it**:
>   auto-drain what can be (wire `reconcile_section_data_action`, re-validate the
>   stale 121), queue the rest. **NOT yet built** — this is the remaining work and
>   why the bug is still OPEN. D2 (residue aging) and D3 (identity/auth) are still
>   open owner decisions.

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

---

# GROUNDED 2026-07-20 (bugfix thread) — the title of this file is wrong

Every figure below was re-checked against the live DB and the current tree on
2026-07-20, after the filing above. **Three of this file's load-bearing claims do
not survive.** The defect is real and worse than described, but it is not the one
in the title.

## > **CORRECTED: a full working surface EXISTS.** It is not absent. It is blind.

`frontends/admin-dashboard/src/App.tsx` contains a complete HITL review surface —
not a stub:

| affordance | file:line |
|---|---|
| work-item list, per-site and cross-site | `App.tsx:397`, `:2354`, `:2362` |
| `needs_human_review` filter + count + badge | `App.tsx:28`, `:958`, `:1011` |
| **Approve & Continue** → `POST /approve` | `App.tsx:1160`, `:694` |
| **Save & Rebuild** (writes spec, creates rebuild item, resolves) | `App.tsx:1169` |
| **Retry** → `POST /retry` | `App.tsx:1180`, `:651` |
| **Resolve / Reject / Skip** → `POST /resolve` | `App.tsx:1184-1189`, `:662` |
| editable review forms, incl. auto-built `needs_section_data` forms | `App.tsx:60`, `:492` |

So "no working surface" is false, and fix candidate 1 ("register
`HandleConfirmWorkItem` — one line") is not the fix. `HandleConfirmWorkItem` is
indeed unreachable (confirmed: grep returns only its definition), but it is a
*duplicate* of the already-registered, already-wired `HandleApproveWorkItem`.
Registering it would change nothing.

## Why the queue was never drained: anyone who looked saw an empty queue

`HandleListWorkItems` hardcodes `limit := 50` (`site_admin_handlers.go:483`) with
`ORDER BY wi.created_at DESC` (`:519`). The frontend requests **no** `status`
param and filters client-side — its own comment states the false belief:

```js
// App.tsx:440-441
// Load all non-complete items, filter client-side for accurate counts
let path = `/work-items?pipeline=build`;
```

It does not load all non-complete items. It loads the newest 50 of them. Measured
against live data:

```
non-complete build items                       : 687
the cross-site "All Items" view can ever show  :  50
of those 50, needs_human_review                :   0     ← the whole queue is invisible
needs_human_review in the build pipeline       : 208
```

**The cross-site view shows zero of 208.** The "Needs Review (N)" dropdown count
(`App.tsx:958`) is computed over the same 50-row window, so it reads **0** — the
queue does not merely look empty, it is *reported* as empty. Nobody ignored 292
items; the dashboard told them there were none.

The per-site view is partially better and badly uneven, because the 50 applies
per site:

| site | review items | visible in UI |
|---|---|---|
| leopardessconsulting.co.uk | 80 | 37 |
| robot-hands.com | 22 | **4** |
| gamesdesign.co.uk | 6 | **0** |
| relojistas.com | 17 | 17 |
| finetuning.uk / dartsonline.com | 12 / 12 | 12 / 12 |
| **build-pipeline total** | **208** | **128** |

And `pipeline=build` is hardcoded at the only fetch site (`App.tsx:441`) with no
UI control to change it, so the **94 `content`-pipeline review items are
unreachable through the dashboard by any route**.

Third blocker: the dashboard has **no Ingress anywhere in the repo** — `base/service.yaml:8`
is `type: ClusterIP` and `base/kustomization.yaml` lists only deployment+service.
Access is `kubectl port-forward` only. (`makefile:1161 port-forward-admin` is also
broken — it forwards to port 80, the Service listens on 8080; the working target
is `makefile:2070 dashboard-port-forward`.)

## > **CORRECTED: the 303 items are not the population the approve flow was built for**

The designed flow is `checkpoint_for_review` → item with `on_approve` in spec →
dashboard "Approve & Continue" → `HandleApproveWorkItem` applies the spec and
creates the follow-on item. It is coherent and fully implemented. **It has never
run, at any stage:**

```
agent_definitions (LIVE) referencing checkpoint_for_review : 0
site_work_items carrying spec->'on_approve'               : 0  (of 303)
site_work_items carrying spec->'checkpoint' , ever        : 0  (of 5,622)
```

`App.tsx:1160` renders **Approve & Continue** only when `spec.checkpoint === true`.
No item in the platform's history has ever had that key, so the button has never
rendered — which is the real reason `HandleApproveWorkItem` has never run.
(`checkpoint_for_review` was itself unregistered until 2026-07-17; `registry.go:619`
records that any workflow referencing it failed validation.)

So the 303 are a **second, different population**: escalations from discovery
checks. They split in two, and the split matters more than queue-vs-bin:

- **~175 deliberate, documented escalations** — `unresolved_cta` (66),
  `cta_names_unknown_destination` (47), `voice_tells` (25),
  `required_fields_missing` (20), `image_source_unsatisfiable` (17). Each carries
  a source comment explaining that a human must decide (e.g.
  `check_voice_tells.go:142` *"never an unreviewed auto-rewrite"*;
  `check_required_fields_missing.go:149` *"needs_human_review keeps it out of the
  dispatch loop while the dedup key holds it open"*). These were **filed
  correctly**.
- **~59 residue** — `content_rewrite` (24), `needs_page` (19),
  `needs_content_page` (16). **No producer emits these at this status.** They
  arrive via `page-build-handler` and `tool-improver`, whose LIVE config uses
  `fail_work_item` with `status_override: needs_human_review` on validation
  failure. Nobody chose human review for these; they are parked failures.

The residue carries a trap: `FailWorkItemAction`'s `status_override` branch
updates status **without incrementing `attempt_count`**
(`load_work_item_actions.go:897-905`). Live, they sit at `attempt_count` 0–2 of
`max_attempts` 3 — they *look* retryable and can never age out to `failed`, while
sitting in a status the dispatch loop structurally cannot see
(`claim_work_item_action.go:102` and `load_work_item_actions.go:559` both filter
`status IN ('triaged','approved')`).

Note `'approved'` there: it is a status **no Go code ever writes**.
`HandleApproveWorkItem` writes `status='complete'`. The gated-approval path
(`approval_mode != 'auto'` → wait → approve → `'approved'` → loader picks it up)
is vestigial — all 5,622 rows are `approval_mode='auto'`.

The one genuine automated consumer, `reconcile_section_data_action.go:114-116`
(re-opens `needs_section_data` when query-sourced data later resolves — 48 items
of the queue), is registered as an action but wired to **0 live agents**. Its own
header says *"Not wired here — pick the host agent and add a step."*

## > **CORRECTED: `resolution_path` is no longer unwritten**

Two rows now carry it, written by hand at 17:58 on 2026-07-20 by the robot-hands
thread. Still no Go code writes it; the column works, and threads use it as the
de-facto surface. `approved_by` remains 0 across all 5,622 rows.

## Staleness — an argument neither branch of the original question makes

121 of the 126 page-linked review items point at pages that have been **rebuilt
since the item was filed**. Nothing re-validates an item against the page it
describes, so an unknown but large share of the backlog is describing pages that
no longer exist in that form. Any "drain the queue" plan that does not re-validate
first will spend human attention on findings about superseded pages.

Sampled evidence that the queue also contains outright false positives — this
item is **correct as built**:

```
summary : CTA "Get in touch" on how-we-work (call-to-action):
          lands in an excluded area (contact/legal/about)
spec.href: /contact.html
spec.fix : "Product decision: build the promised page, or rewrite the copy/link."
```

`check_misdirected_cta.go:267` fires on the destination alone — `ctaAreaExcluded(a.Href)`
— with no label/destination agreement test on that branch. So "Get in touch",
"Ask a question", "Describe the job", "Walk us through your problem" → `/contact.html`
(all correct) are flagged identically to the genuinely broken "Start Ranking Free"
→ `/contact.html` that `bugs_open/023` is built on. **18 of the 47 are the
excluded-area class.** This belongs to the active `cta_link_integrity` workstream
(`bugs_open/023`), not here — flagged, not fixed, to avoid two threads on one check.

## What is a bug, and what is actually an owner decision

The original file says "nothing should be built until the queue-or-bin question is
decided". Most of the above is not waiting on that question:

**Bugs — no ruling needed:**
1. `limit := 50` + client-side filtering ⇒ cross-site view shows 0 of 208, and the
   count reads 0. Pass `status`/`item_type` through to the backend (it already
   supports both, `site_admin_handlers.go:497,513`) and paginate.
2. `pipeline=build` hardcoded ⇒ 94 content items unreachable.
3. No Ingress ⇒ port-forward only. Also `makefile:1161` targets the wrong port.

**Genuine owner decisions:**
- **D1** — the ~175 deliberate escalations: will a human ever work these? The
  checks were written on the assumption that someone would.
- **D2** — the ~59 residue: should a `status_override` failure age out to `failed`
  and re-enter retry, or stay parked? (Today it does neither.)
- **D3** — identity: there is no auth context in these handlers, so "record who
  decided" is blocked on an auth decision, not a one-liner. Unchanged from the
  original filing, and still correct.

**Status after grounding:** OPEN. Bugs 1–3 are the reason the queue was never
worked and can be fixed without the ruling. D1–D3 still need the owner.

## Related

- `bugs_open/023` / `cta_link_integrity` — owns the CTA check; independently found
  the same delivery gap ("filed somewhere nobody reads … a **delivery** gap, and
  it is a different fix", `NOTES_cta_link_integrity.md:196`). The 18 excluded-area
  false positives above are theirs.
- `bugs_open/035` — `updated_at` is not maintained; the `status_override` branch
  above is one of the paths that does not touch it.
- `bugs_open/017` / `work_item_completion_integrity` — the completion end of the
  same lifecycle.
- `bugs_open/032`, `bugs_open/021` — verification coverage; same family of
  "the mechanism exists but almost nothing uses it".
- `reasoning_dataset/PLAN_capture_gaps_and_volume.md` §Gap 2 — the fuller
  write-up, including the corrected figures.
