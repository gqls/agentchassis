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

### While the loop ran: a path that CAN write the shape, and does not

My "no Go path writes `deferred` with a named handler" rested on one grep pattern
(`SET status ... deferred`), which a **parameterised** update would walk straight past. Checked it
properly — every `UPDATE site_work_items … SET status` in the repo, with its value:

```bash
grep -rn -A4 "UPDATE site_work_items" --include=*.go platform/ internal/ pkg/ cmd/ \
  | grep -E "status *= *\\\$"
```

Three parameterised hits. Two are `v3_site_actions.go:6302/6312` — `UpdateWorkItemStatusAction`,
whose `$2` is the caller's `newStatus`. The third is the interesting one:

**`load_work_item_actions.go:1259` — `FailWorkItemAction` honours a step-config key
`status_override`**, and writes it straight into `status`:

```sql
UPDATE site_work_items SET error = $2, status = $3, handled_by = $4 WHERE id = $1
```

It touches `error`, `status` and `handled_by` — and **leaves `handler_agent` alone**. So a step
configured `status_override: "deferred"` produces *exactly* the shape under investigation: parked,
named handler, no `parked_by`.

And the comment immediately above it names the agents that use the key —
*"component-template-fixer ×2, page-build-handler, tool-improver"* — which are **precisely** the
handlers on the untraceable rows (`page_component_status_drift` → `component-template-fixer`,
`improve_tool` → `tool-improver`, `content_rewrite`/`needs_page`/`needs_content_page` →
`page-build-handler`). That is a very good-looking lead.

**And it is not the answer.** Two things refute it, and I checked both rather than stopping at the
resemblance:

1. **`FailWorkItemAction` stamps `handled_by = agentType`.** Every one of the bulk-parked rows has
   `handled_by` NULL/empty. A row written by this path would name its writer.
2. **Every live `status_override` in the fleet is `needs_human_review`, not `deferred`** — read
   with a recursive walk over `agent_definitions`, all four of them:
   `component-template-fixer>judged_refusal`, `component-template-fixer>park_refused`,
   `page-build-handler>mark_needs_review`, `tool-improver>refuse_mangled_write`.

> ⚠ **What this DOES establish, and it is worth a landmine on its own: the black hole is ONE CONFIG
> KEY away.** `status_override` is an ordinary step-config string with no allow-list — nothing
> validates it against the statuses the dispatcher, the promoter and `idx_swi_dedup` actually
> understand. A session setting `status_override: "deferred"` on any refusal step would silently
> mint undispatchable, un-promotable, un-re-filable rows at production rate, and every field on
> them would look healthy. The four live values are `needs_human_review` **by convention, not by
> constraint.**

So the shape still has no live producer, and the resemblance that looked decisive was a near-miss.
**[UNMEASURED]** what actually wrote the 118 — that is the loop's question, and I have deliberately
not guessed at it here. The comfortable answer (earlier sessions running `psql` by hand) remains
untested, and I have no evidence for it beyond the absence of alternatives, which is not evidence.

### `FailWorkItemAction` conclusively ruled OUT — with a control that proves the instrument works

The path stamps `handled_by = agentType` unconditionally. So "do the 118 carry `handled_by`?" is a
decisive test — **provided the column is actually written somewhere**, or a zero means nothing.
Both halves in one run:

| the 118 (deferred, named handler, no stamp) | |
|---|---|
| rows | **118** |
| with `handled_by` | **0** |
| with `error` | **0** |
| with `attempt_count > 0` | **0** |
| ever `triaged_at` | **1** |
| ever `claimed_at` | **1** |

**Positive control — is `handled_by` written at all?** Yes, heavily: **7,114 of 7,329** `complete`
rows carry it, plus 156 of 732 `cancelled`, 131 of 963 `needs_human_review`, 76 of 179 `failed`.
So the zero above is a real absence, not a dead column. (All **303** `deferred` rows carry none,
migration 389's 87 included — consistent, since a migration would not set it.)

**Two things this establishes:**

1. **`FailWorkItemAction` did not write these rows.** Not one carries its fingerprint.
2. **117 of the 118 never entered the dispatch queue at all** — never triaged, never claimed, never
   attempted. So the producer acted on rows in a *pre-dispatch* state, or created them parked. This
   rules out "dispatched, then parked after a failure", which was my second-favourite hypothesis
   and the one the `error` column would have evidenced.

⚠ It still does **not** distinguish born-deferred from moved-to-deferred — `triaged_at` is NULL in
both cases. That question may simply be unanswerable from this table, and the bug file must say so
rather than choosing the comfortable answer.

### Two more candidates raised and killed, and one real correlation left standing

**Candidate: migration `217_site_work_items_handler_agent_not_null.sql` backfilled handler names
onto correctly-parked rows.** This would have dissolved the whole finding — the 118 would be
honest roadmap rows whose empty handler was later filled in, and the "wrong shape" would be an
artefact of the migration rather than a defect. **REFUTED by reading it**: 217 backfills
`handler_agent = ''` **WHERE handler_agent IS NULL** — it collapses NULL onto empty, the opposite
direction, and then sets `DEFAULT ''` + `NOT NULL`. It cannot put a name on anything.

**Candidate: a later router stamps `handler_agent` onto rows born deferred-and-empty.** No such
writer exists — nothing in `platform/`, `internal/`, `pkg/` or `cmd/` does `UPDATE … SET
handler_agent = <a name>`. The only occurrences are `claim_work_item_action.go:173` (setting an
`error` when a handler is *missing*) and test fixtures.

**The correlation that IS real, and is the best lead left.** `agent_error_log` retains to
2026-07-24, so it covers every bulk-park minute. At **08-04 22:03–22:04**, the minute idea.uk's 14
rows were parked, idea.uk's `completeness-discovery-agent`, `design-discovery-agent` and
`quality-discovery-agent` all logged `complete` — **a full discovery run finishing on that exact
site at that exact minute**, and every one of those parked rows carries `source='discovery'`.

⚠ **[UNMEASURED] and I am deliberately not concluding from it.** A discovery run completing at the
same minute is consistent with the discovery write path parking them — and equally consistent with
discovery merely *touching* them (bumping `updated_at`) while they were already parked, which is
misstep 1's trap wearing a new hat. **The same ambiguity, one layer along: I still cannot tell a
park from a touch, and a co-occurring actor is not a writing actor.** Two of my four hypotheses
today have died on exactly this distinction; the third would too if I let it.

What is now excluded, so the loop's answer can be checked against it:

| candidate | verdict | on what evidence |
|---|---|---|
| `FailWorkItemAction` + `status_override` | **OUT** | stamps `handled_by`; 0 of 118 carry it; all 4 live values are `needs_human_review` |
| migration 217 backfill | **OUT** | backfills to `''`, not to a name |
| a later `handler_agent` router | **OUT** | no such writer exists in the repo |
| `refreshOpenWorkItem` | **OUT** | description only; and none of these item types uses `refreshOnConflict` |
| dispatched-then-parked-on-failure | **OUT** | 117 of 118 never triaged, claimed or attempted; 0 carry `error` |
| a discovery-run side effect | **OPEN — best lead** | timing correlation only |
| a hand-run `psql` UPDATE | **OPEN — untested** | no evidence beyond absence of alternatives, which is not evidence |
