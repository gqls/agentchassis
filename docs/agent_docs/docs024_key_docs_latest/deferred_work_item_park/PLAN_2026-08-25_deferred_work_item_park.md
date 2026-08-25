# PLAN 2026-08-25 — work items parked at `deferred` with a named handler: undispatchable, untraceable, and blocking their own re-file

## Why this lane exists

Opened 2026-08-25 by the `bugs_open/328` lane, which hit one of these rows while closing 328 and
was asked by the owner to take the general case on.

The instance: `mortgagecalculator.co.uk /guides/mortgage-scorecard/index.html` held
`page_rerender_guide-mortgage-scorecard_…_assemble` at `status='deferred'` since **2026-08-03**,
`attempt_count=0`, `triaged_at` NULL, with `handler_agent='page-rerender'`. It had sat **22 days**.
Nothing was ever going to run it, and because `deferred` is not in `idx_swi_dedup`'s
terminal-status list it also held the `(site_id, item_key)` slot, so a fresh dispatch for that page
**failed 23505** — a failure that reads as *"already queued"* and means *"queued and abandoned"*.
Re-armed to `triaged`, it completed **2 minutes** later.

## What is actually true [MEASURED 2026-08-25, live DB + repo at HEAD]

**The platform HAS a deliberate, well-designed park**, and it is not this. The convention is
`status='deferred'` **+ empty `handler_agent`** = *"work we can SEE and cannot ACT on"* — a roadmap
row, not a dispatch. It has a live consumer (`diagnose_triage_action.go:361` and
`fixloop_digest_action.go:358`, both reading `(item_type='capability_gap' OR status='deferred')`)
and a live drain (`work_item_retraction.go:205`, which counts `parked` separately precisely because
*"the park draining unnoticed is exactly what this counter exists to make visible"*).

**Every Go writer of work-item `deferred` obeys it.** Verified by reading each one, not by grep
alone: `discovery_checks/remit.go:202`, `write_audit_findings_action.go:427` and `:584`,
`load_work_item_actions.go:279` (the `capability_gap` arm), `discovery_checks/check_palette_contrast.go:138`,
`discovery_checks/check_content_duplication.go:251` — **all six pair it with `HandlerAgent: ""`**,
deliberately, with comments saying why. Several also stamp `spec.not_dispatchable` explaining the
row to a human reader.

⚠ `plan_sections_action.go`'s four `deferred` sites are a **different `deferred`** — a section-plan
status (`"ready" | "deferred" | "skipped"`, `:906`), not a work-item status. Do not count them.

**And `UPDATE … SET status='deferred'` appears in NO Go path at all.** The admin endpoints
(`site_admin_handlers.go`) only set `complete` and `triaged`.

### So the shape in question has no producer in the codebase

| population | rows | trace |
|---|---|---|
| `parked_by = migration_389` (all `contrast_failure`) | **87** | full — `parked_by`, `parked_reason`, `parked_from_status` |
| no `parked_by`, **empty** handler | **75 of 216** | `spec.not_dispatchable` — the honest roadmap convention |
| **no `parked_by`, NAMED handler, no `not_dispatchable`** | **118** | **NONE** |

**118 rows, ~20 item types, ~8 producers, and no record of who parked them or why.**

⚠ **Migration `389` is the model and should be the template for any future park**: a precondition
check that aborts if the premise is gone, `spec.parked_from_status` / `parked_reason` /
`parked_by`, a row-count assertion against a pre-count, and a negative control proving nothing else
moved. The 87 rows it made are fully traceable *and* owned (`bugs_open/296`, lane
`bugfix_131_contrast_ratio_check`, **ACTIVE**) — **they are out of scope here and must not be
touched.**

### What I have NOT established, and must not assert

- ~~"Every one of these was created in another status and moved to `deferred` later"~~ — **the test
  I first ran cannot show that.** `trg_site_work_items_updated_at` bumps `updated_at` on *every*
  write, so a row born `deferred` and later touched is indistinguishable from one deferred later by
  `updated_at - created_at`. The clustering (single site, many item types, one minute) is real and
  unexplained, but it is equally consistent with a bulk *touch*.
- Whether the 118 are hand-run `psql` parks by earlier sessions, a Go path I have not found, or
  something else. **This is the open question**, and it is why the diagnosis loop was filed rather
  than a bug.

## Approach

1. **Diagnosis loop first** (owner ruling 2026-07-31 — a cross-cutting structural root cause is not
   filed until it has been through `090` or the session says why it substituted equivalent
   verification). Filed 2026-08-25: intake `4623672c-d942-4dfe-a7a4-41bdbf500c5c`, run
   `6061299a-cb6a-497f-b5eb-d31b3bb7771c`. Symptom deliberately asserts no counts and excludes
   `migration_389`'s rows.
2. **Then file the bug** in `/bugs_open/`, citing the verdict — CONFIRMED or REFUTED, both are
   results.
3. **Fix candidates, to be ordered by what makes the bad state unrepresentable** (not yet chosen):
   - a park that cannot be untraceable — require `spec.parked_by`/`parked_reason`, the 389 shape;
   - an un-park route — nothing today moves a row out of `deferred` except the retraction sweep,
     and that only covers types with a registered check;
   - a periodic report of parked-with-handler rows, so the population cannot grow silently again;
   - make the bad state unrepresentable at the index: `deferred` either releases the dedup slot or
     is refused with a handler set.
4. **Do not touch `contrast_failure`.** Contribute into `bugs_open/296` if anything here bears on
   it.

## What "done" looks like

A named cause for the 118, a decision on whether they should be re-armed or closed, and a control
that stops the population regrowing without a trace. **Re-arming is not free** — each one is a real
dispatch onto a live customer site, and the 328 lane's own experience is that a re-render carries
every platform change since that page last rendered.
