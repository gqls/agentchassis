# NOTES — the `deferred` work-item park

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-25 — lane opened out of `bugs_open/328`, and the first framing was already too big

Opened after 328's closure hit one of these rows. The owner asked for the general case.

### How it started, and the number I led with was three populations stacked

My first report to the owner said **"297 deferred rows, every one with `attempt_count = 0`, 205
naming a real handler"**, and framed all of it as *"jobs nothing will ever pick up"*.

That was too big, and I knew part of it within minutes of looking properly. The 297 is **three
unrelated populations**:

| population | rows | verdict |
|---|---|---|
| `deferred` + **empty** handler, most stamping `spec.not_dispatchable` | 75 (of the 216 unstamped) | **CORRECT BY DESIGN** — the estate's roadmap convention |
| `parked_by = migration_389`, all `contrast_failure` | 87 | **traceable AND owned** (`bugs_open/296`, lane `bugfix_131_contrast_ratio_check`, ACTIVE) |
| named handler, no `parked_by`, no `not_dispatchable` | **118** | the actual question |

**The first population is not damage and saying so would have started an argument with six
well-commented code sites.** `discovery_checks/remit.go` calls it a "double lock"; the convention
has a live consumer (`diagnose_triage_action.go:361`, `fixloop_digest_action.go:358`, both reading
`(item_type='capability_gap' OR status='deferred')`) and a live drain
(`work_item_retraction.go:205`, which counts `parked` separately *because* "the park draining
unnoticed is exactly what this counter exists to make visible"). That is a designed mechanism with
both ends wired, and my headline had it as a leak.

⚠ **The transferable half: a status is not a population.** `WHERE status='deferred'` looked like one
finding and was three. The discriminator that separates them — `spec ? 'parked_by'` and
`handler_agent <> ''` — costs one extra `GROUP BY` and I did not have it in the first query.

### MISSTEP 1 — I ran a test that structurally could not answer its question, and briefly believed it

To ask *"were these rows born `deferred`, or moved there later?"* I measured
`updated_at - created_at` and got **"0 of 205 born deferred"** — a clean, decisive-looking result
that I nearly wrote into a bug file.

**It means nothing.** `trg_site_work_items_updated_at` is BEFORE UPDATE FOR EACH ROW and bumps
`updated_at` on *every* write, so a row born `deferred` and later touched by anything is
**indistinguishable** from one created in another status and deferred later. `site_work_items`
keeps no status history, so the question has no answer from that column at all.

What caught it: recognising the shape from a landmine I had read this morning — *"a periodic write
to an open work item makes it UNREAPABLE for ever; `trg_site_work_items_updated_at` bumps
`updated_at` on EVERY write"* (`bugfix_213`). The same trigger, the same column, a different wrong
conclusion.

**This is the second time in one session I have taken an arithmetic or inferential shortcut on my
own output and had to retract it** (the first: the 328 closure's "12 dead anchors / 13 of 21
pages", both wrong, `WRONG_CALLS.md`). Both were caught, neither by re-reading — by a second
instrument disagreeing.

### The clustering is real, and it is still unexplained

Deferred rows with a named handler cluster into **ten one-minute events**, mostly **one site across
many item types**:

| event | site | rows | item types |
|---|---|---|---|
| 08-12 13:31 | loancalculator.co.uk | 43 | 8 |
| 08-11 12:31 | *fleet-wide, 14 sites* | 87 | 1 (`contrast_failure` — this is migration 389) |
| 08-11 18:20 | loancalculator.co.uk | 17 | 3 |
| 08-02 23:31 | mortgagecalculator.co.uk | 16 | 3 |
| 08-04 22:03 | idea.uk | 14 | 3 |
| 08-03 11:02 | mortgagecalculator.co.uk | 12 | 1 |

None of the non-389 rows carries `handled_by` or `error`. **[UNMEASURED]** whether these are bulk
*parks* or bulk *touches* — see misstep 1; the timestamps cannot tell them apart.

### What IS established first-hand, by reading every site rather than grepping

- **No Go path anywhere does `UPDATE … SET status='deferred'`.** The admin endpoints
  (`site_admin_handlers.go`) only set `complete` and `triaged`.
- **All six Go writers of work-item `deferred` pair it with `HandlerAgent: ""`**, deliberately and
  with comments: `remit.go:202`, `write_audit_findings_action.go:427` and `:584`,
  `load_work_item_actions.go:279`, `check_palette_contrast.go:138`,
  `check_content_duplication.go:251`.
- ⚠ **`plan_sections_action.go`'s four `deferred` hits are a DIFFERENT `deferred`** — a section-plan
  status (`"ready" | "deferred" | "skipped"`, `:906`), not a work-item status. Counting them would
  have made it look as though a Go path *does* produce the shape, inverting the conclusion. Caught
  by opening the file instead of trusting the grep line.
- `refreshOpenWorkItem` (`load_work_item_actions.go:~2116`) updates **description only** — status,
  priority and handler are explicitly untouched — and only evidence/citation paths use
  `refreshOnConflict`, none of the item types in question.

So the shape has **no producer in the codebase**, which is exactly the kind of claim that should not
be filed on my own reading.

### Filed the diagnosis loop rather than the bug

Owner ruling 2026-07-31: a `bugs_open/` file asserting a cross-cutting or structural root cause is
not filed until it has been through `090`, or the session states plainly why it substituted
equivalent first-hand verification. This is squarely that class — cross-cutting, and the cause is
by definition *not* where the symptom is, since the symptom is rows and the cause is whatever wrote
them.

- intake `4623672c-d942-4dfe-a7a4-41bdbf500c5c`
- run `6061299a-cb6a-497f-b5eb-d31b3bb7771c` ← the key artifacts are written under

Symptom authored to the house rules: states the MECHANISM, points at the tables and symbols, asserts
**no counts** (the loop fetches and cites them), no downstream-consequence clauses, one bug, and
explicitly excludes `parked_by=migration_389` as owned by another active lane.

### Migration 389 is the model, and it deserves saying out loud

`389_park_contrast_failures_and_reenable_improvement_sweep.sql` is the one bulk park in the estate
that can be audited after the fact: a precondition that `RAISE EXCEPTION`s if the premise is already
gone, `spec.parked_from_status` / `parked_reason` (naming the bug AND the restore condition) /
`parked_by`, a `GET DIAGNOSTICS` row-count assertion against the pre-count, and a negative control
proving nothing else moved. Every one of its 87 rows is traceable 14 days later. **Whatever made the
other 118 left nothing.**
