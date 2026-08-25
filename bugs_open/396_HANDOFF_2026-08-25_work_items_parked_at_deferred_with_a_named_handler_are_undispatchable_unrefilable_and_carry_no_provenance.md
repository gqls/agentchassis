# 396 — 114 work items sit at `deferred` with a named `handler_agent`: undispatchable, un-promotable, un-re-filable, and carrying no record of who parked them or why

**Filed 2026-08-25** by the `deferred_work_item_park` lane, spun out of `bugs_open/328` at the
owner's direction after one of these rows blocked 328's own dispatch.
**Status: OPEN, UNOWNED.**
**Severity: medium.** Nothing is breaking for a customer. The harm is that work the platform
decided to do disappears silently, **and the page it belonged to becomes unrequestable** — a fresh
dispatch for it fails `23505` on `idx_swi_dedup`, a failure that reads as *"already queued"* and
means *"queued and abandoned"*.
**Class:** silent state + missing provenance. **Root cause NOT established — see §5.**

> **On the 090 loop: RUN, and it did NOT confirm.** Intake `4623672c-d942-4dfe-a7a4-41bdbf500c5c`,
> run `6061299a-cb6a-497f-b5eb-d31b3bb7771c`, 4 iterations, verdict **UNVERIFIABLE — "NOT CONFIRMED
> (stopped: iteration-cap) … Hand to a human with the full trail; do NOT auto-conclude."** It
> refuted its own candidate (`createDeferredItems` — it inserts `needs_section_data` at
> `needs_human_review` and does not write `handler_agent` at all) and ended with **zero remaining
> named candidates in the read code**. §5 records what it found that I had missed, and the one lead
> it left that does not exist.

## 1. The one-paragraph version

The estate has a deliberate, well-built park: `status='deferred'` **with an empty `handler_agent`**
means *"work we can see and cannot act on"* — a roadmap row, not a dispatch. It has six commented
writers, two consumers (`diagnose_triage_action.go:361`, `fixloop_digest_action.go:358`) and a
drain (`work_item_retraction.go:205`) that counts parked rows separately *precisely so the park
cannot empty unnoticed*. **That mechanism is correct and is not this bug.** What this bug is about
is 114 rows in a shape no Go writer produces — parked, but still naming a real handler — which
therefore look like ordinary queued work to every reader, are selected by nothing, and cannot be
re-filed by the detector that found them.

## 2. Evidence [MEASURED 2026-08-25 12:47Z — re-run before quoting, this grows]

Computed, not eyeballed (`GROUP BY` over a `CASE`, not a count read off a table):

| population | rows | types | verdict |
|---|---|---|---|
| traceable — `spec.parked_by = 'migration_389'` | **87** | 1 | correct, stamped, **and OWNED** (§7) |
| traceable — `spec.deferred_reason` (`canary_replan_407`, owner-sanctioned 08-12) | **4** | 2 | correct, stamped |
| correct convention — **empty** `handler_agent` (75 also declare `spec.not_dispatchable`) | **98** | 6 | **not a defect** |
| **UNTRACEABLE — named handler, no stamp of any kind** | **114** | **18** | **this bug** |
| **total `deferred`** | **303** | | |

The 114, measured together with a control that proves the instrument works:

| | |
|---|---|
| carry `handled_by` | **0** |
| carry `error` | **0** |
| `attempt_count > 0` | **0** |
| ever `triaged_at` | **1** |
| ever `claimed_at` | **1** |

⚠ **The control matters**: `handled_by` is populated on **7,114 of 7,329** `complete` rows (plus 156
`cancelled`, 131 `needs_human_review`, 76 `failed`), so the zero above is a real absence, not a dead
column.

**So 113 of the 114 never entered the dispatch queue at all.** They were never triaged, never
claimed, never attempted, and carry no error.

### The worked instance, and what it cost

`mortgagecalculator.co.uk` `/guides/mortgage-scorecard/index.html` held
`page_rerender_guide-mortgage-scorecard_62b5978e…_assemble` at `deferred` from **2026-08-03**,
`attempt_count=0`, `triaged_at` NULL, `handler_agent='page-rerender'`. It sat **22 days**. It
blocked `bugs_open/328`'s dispatch for that page with a `23505`. Re-armed to `triaged`, it
**completed in 2 minutes** and the page served correctly.

## 3. Why the shape is fatal — three predicates nobody reads together

| what | where | selects |
|---|---|---|
| dispatch | `claim_work_item_action.go:102` | `status IN ('triaged','approved')` |
| promotion | `triage_detect_items_action.go` | `status='detected'` |
| dedup slot **held** | `idx_swi_dedup` (read it from `pg_indexes`) | `status <> ALL (complete, verified, rejected, wont_fix, failed, unresolved, cancelled)` |

`deferred` is absent from the first two and absent from the third's *exclusion* list. Undispatchable,
un-promotable, **and still holding its `(site_id, item_key)` slot.**

## 4. What is EXCLUDED, with the evidence — so a future reader does not re-walk it

| candidate | verdict | on what |
|---|---|---|
| `FailWorkItemAction` + `status_override` | **OUT** | it stamps `handled_by = agentType`; **0 of 114** carry it. And all four live `status_override` values are `needs_human_review` (recursive walk over `agent_definitions`: `component-template-fixer>judged_refusal`, `>park_refused`, `page-build-handler>mark_needs_review`, `tool-improver>refuse_mangled_write`) |
| migration `217_..._handler_agent_not_null` | **OUT** | backfills `handler_agent = ''` **WHERE … IS NULL** — the opposite direction; it cannot put a name on anything |
| a later router that stamps `handler_agent` | **OUT** | no `UPDATE … SET handler_agent = <name>` exists anywhere in `platform/`, `internal/`, `pkg/`, `cmd/` |
| `refreshOpenWorkItem` | **OUT** | updates the DESCRIPTION only — status, priority and handler explicitly untouched — and only evidence/citation paths use `refreshOnConflict` |
| dispatched, then parked on failure | **OUT** | 113 of 114 never triaged/claimed/attempted, and none carries an `error` |
| any Go path at all writing this shape | **OUT** | no `UPDATE site_work_items … SET status='deferred'` exists; **all six** Go writers of the status pair it with `HandlerAgent: ""`, deliberately and with comments |
| a discovery-run side effect | **OPEN — best lead, timing only** | `agent_error_log` retains to 07-24 and covers every bulk-park minute. At **08-04 22:03–22:04**, the minute idea.uk's 14 rows were parked, its `completeness-`, `design-` and `quality-discovery-agent` all logged `complete` — and every parked row on that site carries `source='discovery'` |
| a hand-run `psql` UPDATE by an earlier session | **OPEN — untested** | no evidence beyond the absence of alternatives, **which is not evidence** |

⚠ **`plan_sections_action.go`'s four `deferred` hits are a DIFFERENT `deferred`** — a section-plan
status (`"ready" \| "deferred" \| "skipped"`, declared `:906`), not a work-item status. Counting them
inverts the conclusion, because it makes it look as though a Go path *does* produce this shape.

## 5. What the diagnosis loop added, and the one thing it got wrong

**It found a stamp I never looked for.** I had checked for `spec.parked_by` and
`spec.not_dispatchable` — the two conventions I knew — and the loop surfaced a **third**,
`spec.deferred_reason`, carrying a full owner-sanctioned explanation (`canary_replan_407`, corr
`b23b19c7`). It is only **4 rows**, so the headline barely moves (118 → 114), but the method lesson
is the point: **I checked for the stamps I knew about; I should have enumerated what stamps exist.**
`SELECT k, count(*) FROM site_work_items, LATERAL jsonb_object_keys(spec) k WHERE …` is one query
and it is now in the lane RUNBOOK.

⚠ **Its remaining lead is void, and I checked rather than inheriting it.** The verdict says *"still
needed: the full body of `HandleUpdateWorkItem` — the only call site the index shows that SETs
`handler_agent` on an UPDATE"*. **There is no such symbol**: `grep -rn "func Handle.*WorkItem"`
across the repo returns nothing, and every `handler_agent` reference under
`internal/core-manager/admin/` is an INSERT column list. The loop flagged its own caveat (only a
signature was indexed, not a body) and that caveat is the tell — **the code index is stale or the
symbol is gone.** Do not spend a round on it.

⚠ **`spec.reason` is NOT park provenance.** It appears on 22 of the 114 and records why the item was
*detected* (`cta_links_stale`, `not_built`, `no_style_collection`, `post_reconcile_assembly`), not
why it was parked. Reading it as a trace would shrink the population on a false basis.

## 6. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make an untraceable park impossible.** A park writes `spec.parked_by` + `parked_reason` or it
   does not happen — enforced where the write happens, not by convention. **Migration `389` is
   already the model** and should be the template: a precondition that `RAISE EXCEPTION`s if the
   premise is gone, `parked_from_status`/`parked_reason`/`parked_by`, a `GET DIAGNOSTICS` row-count
   assertion, and a negative control proving nothing else moved. Its 87 rows are still fully
   auditable 14 days on; the other 114 left nothing.
2. **Close the `status_override` hole before it mints more.** It is a bare step-config string
   written straight into `status` with no allow-list, no enum and no CHECK — validated against
   none of the three predicates in §3. The four live values agree **by convention, not by
   constraint**, and `"deferred"` is the natural word for "park this". Already recorded in
   `LANDMINES.md` (2026-08-25).
3. **Give the park a drain, or give it a reader.** Today nothing moves a row out of `deferred`
   except `work_item_retraction.go`, and only for item types with a registered check. A periodic
   report of parked-with-handler rows would at least stop the population growing unseen —
   **detection is not the gap here, but visibility is.**
4. **Decide the 114.** Re-arm or close, per item type. ⚠ **Re-arming is not free**: each is a real
   dispatch onto a live customer site, and `bugs_open/328`'s experience is that a re-render carries
   every platform change since that page last rendered. 328 re-armed exactly one, deliberately.
5. *(Rejected as a first move)* changing `idx_swi_dedup`'s predicate so `deferred` releases the
   slot. It is in lockstep with `workItemTerminalStatuses`, and **migration 157 already broke that
   pair once fleet-wide with SQLSTATE 42P10.** Architecture-scope, not a patch.

## 7. Out of scope — do not touch

**The 87 `contrast_failure` rows parked by migration 389 are `bugs_open/296`**, and that bug is
**OWNED** by the `bugfix_131_contrast_ratio_check` lane (`scripts/who-owns.py 296` → ACTIVE, 17
commits/14d). They are traceable, deliberate, and being drained by a live retraction. Contribute
into `bugs_open/296` if anything here bears on them; do not start a competing fix.

**The 98 empty-handler rows are correct** and their mechanism is designed. Reporting them as damage
is the mistake this lane made in its first hour, and it would put you in an argument with six
well-commented code sites that are right.

## 8. How to verify a fix

- The split query in §2 — **computed**, not read off a table — shows population 4 at **0**, and
  populations 1–3 unchanged. Both halves: a fix that emptied the correct populations too would pass
  a bare "population 4 is zero" check.
- A **new** park, made after the fix, carries `parked_by` and `parked_reason`.
- **Induce it**: attempt an untraceable park and watch it be refused. A guard that has never been
  seen to fire is a guard you are assuming.
- `status_override` set to a value outside the dispatcher's vocabulary is refused at config-load or
  at write, with the refusal visible — not silently honoured.

## 9. Where the working record lives

`docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/` — PLAN, NOTES (append-only,
newest at the bottom; the cold-start read, and it carries the missteps), RUNBOOK (the split query,
the `jsonb_object_keys` enumeration, and the ⚠ that `updated_at - created_at` **cannot** tell
"born deferred" from "deferred later", because `trg_site_work_items_updated_at` bumps on every
write and the table keeps no status history), README_where_we_are.
`LANDMINES.md`: the `status_override` entry (2026-08-25), plus a correction the same day to the
older entry claiming `deferred` is only undispatchable when the handler is empty — it is
undispatchable either way, because `claim_work_item_action.go:102` filters on **status**.
