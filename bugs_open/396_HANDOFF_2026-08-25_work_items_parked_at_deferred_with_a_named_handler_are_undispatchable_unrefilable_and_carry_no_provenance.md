# 396 — 52 work items sit at `deferred` with a named `handler_agent`: undispatchable, un-promotable, un-re-filable, and carrying no record of who parked them

> **The title said 114 when filed. Corrected to 52 the same day — 62 of the 114 ARE stamped, in `result` rather than `spec`. See the correction block below.**

**Filed 2026-08-25** by the `deferred_work_item_park` lane, spun out of `bugs_open/328` at the
owner's direction after one of these rows blocked 328's own dispatch.
**Status: OPEN, UNOWNED. ROOT CAUSE ESTABLISHED 2026-08-25 (see the correction block) — deliberate hand-run per-site queue holds, because the platform has no park verb.**
**Severity: medium.** Nothing is breaking for a customer. The harm is that work the platform
decided to do disappears silently, **and the page it belonged to becomes unrequestable** — a fresh
dispatch for it fails `23505` on `idx_swi_dedup`, a failure that reads as *"already queued"* and
means *"queued and abandoned"*.
**Class:** silent state + missing provenance. ~~**Root cause NOT established**~~ — **ESTABLISHED 2026-08-25, one hour after filing, by an independent review: deliberate hand-run per-site queue holds. See the correction block.**

> **On the 090 loop: RUN, and it did NOT confirm.** Intake `4623672c-d942-4dfe-a7a4-41bdbf500c5c`,
> run `6061299a-cb6a-497f-b5eb-d31b3bb7771c`, 4 iterations, verdict **UNVERIFIABLE — "NOT CONFIRMED
> (stopped: iteration-cap) … Hand to a human with the full trail; do NOT auto-conclude."** It
> refuted its own candidate (`createDeferredItems` — it inserts `needs_section_data` at
> `needs_human_review` and does not write `handler_agent` at all) and ended with **zero remaining
> named candidates in the read code**. §5 records what it found that I had missed, and the one lead
> it left that does not exist.

> ## ⚠⚠ CORRECTED 2026-08-25, ~1 HOUR AFTER FILING — the headline was wrong about 62 of the 114, and the root cause IS established
>
> An independent review (Claude Fable 5, commissioned by the owner) refuted the central claim and
> **answered the open question**. Every point below was re-verified first-hand before this block was
> written.
>
> **1. "No provenance" is FALSE for 62 of the 114. The stamps are in `result`, and I only ever
> enumerated `spec`.** [MEASURED 2026-08-25] `result.deferred_by` = **62** rows:
> `loancalculator_rebuild_thread` (60, with `deferred_from_status='detected'` and a
> `deferred_reason` naming the owner-ordered rebuild AND its release condition) and
> `apis-uk-bees-lane` (2, with `deferred_at` and a full reason). **The genuinely untraceable
> population is 52, not 114.**
> ⚠ **The method error is the one this very file teaches in §5.** I wrote *"do not test for the
> stamps you know, enumerate the stamps that exist"* — and then ran that enumeration against
> **one column**. `result` is an established provenance channel here; migration `442` stamps
> `result.repair_284` for exactly this purpose.
>
> **2. ROOT CAUSE ESTABLISHED: deliberate per-site queue holds, run as hand psql by named sessions,
> because the platform has no park verb.** The hypothesis §4 listed as *"no evidence beyond the
> absence of alternatives"* had abundant evidence — **in the repo, in other lanes' documents**,
> which I never searched. I grepped the *code* for a writer on a tree where sessions run psql by
> hand and write down what they did.
>
> | rows | site | parked | the writer's own record |
> |---|---|---|---|
> | 60 | loancalculator.co.uk | 08-11, 08-12 | self-stamped `result.deferred_by='loancalculator_rebuild_thread'` |
> | 38 | mortgagecalculator.co.uk | 07-31 – 08-05 | `mortgagecalculator_couk_adoption`'s documented **15-second auto-defer backstop** — the verbatim SQL is in `HANDOFF_2026-08-03_continue_here.md:81-90`: `UPDATE site_work_items SET status='deferred', updated_at=NOW() WHERE … AND status IN ('triaged','approved')`. **It sets nothing else, which is exactly why these rows carry no stamp.** |
> | 14 | idea.uk | 08-04 22:03 | commit `90a4fb812`: *"Holds: 12× undeployed_asset + footer needs_rerender + deactivated_component → `deferred` (UPDATE 14)"* — matches the surviving rows type-for-type |
> | 2 | apis.uk | 08-24 | self-stamped `result.deferred_by='apis-uk-bees-lane'` |
>
> **3. §4's "best lead" (a discovery-run side effect) is REFUTED.** idea.uk's 14 were **created**
> 07-31 19:18 and **parked** 08-04 22:03 [MEASURED]. The discovery run I found completing at 22:03
> was co-occurring with the **park**, not the birth — and the park was a human's `UPDATE 14`.
>
> **4. The lane that made 38 of them had already forgotten it.**
> `mortgagecalculator_couk_adoption/NOTES…:2844` says **"[UNVERIFIED] what deferred them … a
> hand-park at adoption is the obvious guess and I did not establish it"** — written by the lane
> whose own HANDOFF, in the same directory, carries the recipe it ran. **I inherited that
> `[UNVERIFIED]` and re-derived the mystery from scratch.**
>
> §§1–5 below are left as filed, with the wrong claims struck where they appear. §6 (fix
> candidates) is **sharpened, not withdrawn** — see §6a.

## 1. The one-paragraph version

The estate has a deliberate, well-built park: `status='deferred'` **with an empty `handler_agent`**
means *"work we can see and cannot act on"* — a roadmap row, not a dispatch. It has six commented
writers, two consumers (`diagnose_triage_action.go:361`, `fixloop_digest_action.go:358`) and a
drain (`work_item_retraction.go:205`) that counts parked rows separately *precisely so the park
cannot empty unnoticed*. **That mechanism is correct and is not this bug.** What this bug is about
is rows in a shape no Go writer produces — parked, but still naming a real handler — which
therefore look like ordinary queued work to every reader, are selected by nothing, and cannot be
re-filed by the detector that found them. **They are made by hand, by sessions holding a site's
queue while they rebuild it, because the platform offers no park verb that would stamp them.**

## 2. Evidence [MEASURED 2026-08-25 12:47Z — re-run before quoting, this grows]

Computed, not eyeballed (`GROUP BY` over a `CASE`, not a count read off a table):

| population | rows | types | verdict |
|---|---|---|---|
| traceable — `spec.parked_by = 'migration_389'` | **87** | 1 | correct, stamped, **and OWNED** (§7) |
| traceable — `spec.deferred_reason` (`canary_replan_407`, owner-sanctioned 08-12) | **4** | 2 | correct, stamped |
| correct convention — **empty** `handler_agent` (75 also declare `spec.not_dispatchable`) | **98** | 6 | **not a defect** |
| ~~UNTRACEABLE — named handler, no stamp~~ **CORRECTED: traceable in `result`, not `spec`** — `result.deferred_by` | **62** | 11 | `loancalculator_rebuild_thread` 60, `apis-uk-bees-lane` 2 |
| **UNTRACEABLE — no stamp in EITHER `spec` or `result`** | **52** | **10** | **this bug** — mortgagecalculator 38, idea.uk 14, both attributable to hand-parks via lane documents |
| **total `deferred`** | **303** | | |

The population as originally measured (**114**; the 62 stamped-in-`result` rows are a subset and were
not separated at the time), with a control that proves the instrument works:

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

~~**So 113 of the 114 never entered the dispatch queue at all.**~~ **CORRECTED — this over-reads
`triaged_at`.** The mortgagecalculator parks were made by a backstop whose own WHERE clause is
`AND status IN ('triaged','approved')`, so those rows **were** in the claimable queue when they
were swept; a row born `'triaged'` never gets a `triaged_at` stamp, because only the promoter
writes it. What survives, and is what the exclusion in §4 actually rests on: **never claimed,
never attempted, and carrying no `error`.**

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

## 6c. LIVE STATE 2026-08-26 — both fixes are IN THE BINARY and the config is applied

Chassis **`v1.0.1341`**. Binary-probed on **both** replicas with a present-control
(`repairOutboundPageLinks`) and an absent-control (`zzzNotARealSymbol396zzz`) in the same run —
`honour_site_lock`, `lock_except_item_ids` and `WORK_ITEM_STATUS_OVERRIDE_REFUSED` all **PRESENT**.
⚠ The `build provenance` line had scrolled (pods 9 h old): empty there means *not in range*, not
*unstamped*.

| piece | state |
|---|---|
| `status_override` allow-list (council **APPROVED** `9c16eb83`) | **LIVE** |
| `sites.lock_except_item_ids` (migration `632`) | **APPLIED**, inert |
| `honour_site_lock` arm in `LoadWorkItemsAction` | **LIVE** |
| migration `633` (the held config half) | **APPLIED 2026-08-26**, hold condition met and proven |

`633` verified at the artefact with a query it does not contain: gate **names** the column ✓, gate
**KEEPS** `s.locked_at IS NULL` ✓, `load_items honour_site_lock=true` ✓, negative control
(`site-work-orchestrator` steps with the key) **0** ✓. Inert on apply — 0 sites carry an exception
list, so the new clause evaluates identically to the old for every row.

⚠ **`--record-only` REFUSES a `_HOLD` file** as an uppercase sidecar, so its ledger row was written
by hand (same workaround as `610_..._HOLD`). **Held migrations are otherwise invisible to
`schema_migrations` — "was the held half applied?" can only be answered from the live config.**

### ⛔ TWO THINGS ARE OWED AND NEITHER IS POSSIBLE TODAY

1. **The end-to-end exercise:** lock a site, except one item, confirm that item dispatches and its
   siblings do not.
2. **The council round.** `175df761` r2 ran and died `complete_invalid` — *"no reviewer produced a
   readable opinion (5 abstained, 12 unreadable) — a council with no opinions cannot decide."*
   **It was never judged.** Resubmit when the fleet is back; fix the duplicate file paths in
   `submission_396_site_lock_exception_r2.json` first (8 edits, two paths listed twice).

**Both are blocked by a fleet-wide outage, not by this lane:** Anthropic credits exhausted, last
success `2026-08-25 23:46:29Z`, **631 consecutive failures** since. The build queue is stalled at
**1,399 `triaged` / 0 `claimed`**, oldest triaged 08-18. Already diagnosed by other lanes and the
owner is already notified — `bugs_open/243`'s class. **Do not re-file it.**

## 6b. ⚠ SUPERSEDES §6a's candidate 1 — the fix is the EXISTING site lock, not a park verb (2026-08-25, after council `ed821065` REVISE)

**§6a said the primary fix was a park verb. That was wrong, and the council's `prior_art` seat
caught it as a HIGH.** `sites.locked_at` / `locked_by` already exist, are live on **3 of 51** sites
with a real reason in `locked_by`, gate `build-pipeline-trigger > find_dispatchable_site`, and have
admin lock/unlock endpoints. **The lock mutates NO work-item row** — so it strands nothing, holds no
`idx_swi_dedup` slot, and **the 22-day stall in §2 could not have happened under it.** Migration
`621`'s verb makes the wrong tool tidier.

**What was actually missing is narrower: the lock is all-or-nothing.** That is why the
`mortgagecalculator_couk_adoption` lane wrote the 15-second backstop that minted 38 of the 52 — its
own handoff calls the site lock *"(a)"* and item status *"(b) the finer control"*.

**SHIPPED INSTEAD** (council `175df761`, register **WII-036**):

| artefact | state |
|---|---|
| `sites.lock_except_item_ids uuid[]` — migration `632` | **APPLIED**, inert by construction |
| `siteLockExceptionSQL()` + `honour_site_lock` opt-in in `LoadWorkItemsAction` | **COMMITTED, INERT until the next chassis roll** |
| migration `633_..._HOLD.sql` — the config half | **HELD. Do not apply before the roll.** |

⚠⚠ **THE ORDERING IS NOT BUREAUCRACY.** `find_dispatchable_site` selects a **SITE**, not an item.
`LoadWorkItemsAction`, which runs next, has **never** checked `sites.locked_at` — the lock has
exactly ONE gate. Apply `633` before the binary and a locked site with an exception list dispatches
its **entire** queue, on precisely the sites somebody deliberately locked. Both halves, binary
first.

⚠ Two misreadings that cost this lane time and will cost the next reader more:
`load_work_item_actions.go:134` **looks** like a second gate and is not — it is inside
`WriteBuildItemsAction`, and **its log line misnames its own function** as
`"LoadWorkItemsAction: site is locked, skipping"`. And the selector's SQL lives under
`config.query`, **not** `config.pre_query`.

**Migration 621 (the park verb) is kept, demoted, and re-labelled** — see WII-034's amendment. It
is for parking specific items on an **unlocked** site, where the alternative is the raw `UPDATE`
that caused this bug. It is **not** the answer to "hold this site", and two objections stand
against parking as a mechanism at all: a parked row **still holds its dedup slot**, and
`work_item_retraction.go:205` can **close a deliberate park without any unpark route**.

**Candidate 2 (the `status_override` allow-list) shipped and was APPROVED** — council `9c16eb83`,
Go, inert until the roll.

## 6a. What the corrected root cause does to those candidates — it SHARPENS them

**The cause is not a bug in a writer. It is a MISSING VERB.** Four sessions each needed to hold a
site's queue while rebuilding it; the platform offers no way to do that, so each improvised the
same `UPDATE … SET status='deferred'` by hand, and only the ones who thought of it left a stamp.
That reframes the ranking:

1. **Candidate 1 is now the primary fix, and it is a FEATURE, not a guard.** Give the estate a park
   verb that records `parked_by` / `parked_reason` / `parked_from_status` and a release condition,
   in ONE place. An operator with a supported way to park stops inventing one. Migration `389` and
   `loancalculator_rebuild_thread` both show what a good stamp looks like — they simply had no
   shared implementation to call.
2. **The estate has grown at least SIX ad-hoc provenance conventions** for the same act:
   `spec.parked_by`, `spec.deferred_reason`, `spec.not_dispatchable`, `result.deferred_by` (+
   `deferred_reason` / `deferred_from_status` / `deferred_at`), `result.repair_284` (migration
   `442`), and a reason appended to `created_by`. **That divergence is the argument for
   standardising, and it is stronger evidence than the 52 unstamped rows are.**
3. **Candidate 2 (`status_override`) stands, and gains a sibling.** Two registered actions take a
   status from step config — `create_work_item_action.go:194-222` (`config["status"]` +
   `config["handler_agent"]`) and `UpdateWorkItemStatusAction` (`v3_site_actions.go:5953+`). A
   recursive walk over every `agent_definitions` row finds exactly **one** step fleet-wide
   configured `deferred` (`improvement-loop`/`create_work_item`, `capability_gap`, **empty**
   handler) and **zero** `status_override='deferred'` [MEASURED 2026-08-25]. **Cite that walk as the
   evidence, not the literal-grep absence** — a grep over source cannot clear a config-driven door.
4. ⚠ **Candidate 4 ("decide the 52", and the 62) MUST route through the owning lanes.** The 60
   loancalculator rows carry a live release condition — *"un-park after rebuild verify"* — and
   blanket re-arming would fire 60 rows another lane is deliberately holding. The 2 apis.uk rows
   state their own unblock condition. The mortgagecalculator 38 belong to an adoption lane with an
   active NOTES file. **Ask the holders; do not sweep.**

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

## CONTRIB 2026-08-25 (from the `bugfix_333_owned_page_door` lane) — your named-handler split is correct, and 11 of the 114 DO carry creation provenance

Prompted by the `bugfix_392_link_context_unread` lane, which worried your census might conflate the
owned-page door's parked rows (40 `content_rewrite`, `deferred`) with this bug's population. **It does
not — verified**: the door clears `handler_agent` and leads `error` with `OWNED_PAGE_GUARD`, so your
named-handler bucket excludes them by construction, and your "correct convention" row counts them.
For any future census: door rows = `handler_agent=''` AND `error LIKE 'OWNED_PAGE_GUARD%'`.

**The lead you may want** [MEASURED 2026-08-25 ~14:0xZ, live]: the 11 `content_rewrite` rows at
`page-build-handler` inside your 114 resolve by `created_by` to exactly two lanes —
`voiceh-rollout` (9 rows, 2026-08-08, all on `rebuild_policy='generic'` pages) and
`apis-uk-bees-lane` (2 rows, 2026-08-24, `page_id` resolving to no live page row). Your §"no stamp of
any kind" is too strong for these: `created_by` is populated, and since your own control shows 113 of
114 never entered dispatch, the CREATOR is the PARKER for them — the rows were born at `deferred` by
whatever those two lanes ran. The cheap next step for this slice: grep those lanes' dirs/scripts for a
direct `'deferred'` write (`grep -rn "deferred" docs/agent_docs/docs024_key_docs_latest/*voiceh* …`),
which should name the writer this file's Go-writer sweep correctly proves does not exist in
`platform/`. Also checked at your request: both `Status: "deferred"` arms in
`write_audit_findings_action.go` set `HandlerAgent: ""` — no Go writer contradiction; your central
claim stands.

---

> **NOTE from the noted.co.uk lane, 2026-08-25 late (drive-by, not taking the bug):**
> commit `2b46afbe6`'s `WORK_ITEM_STATUS_OVERRIDE_REFUSED` (work_items_common.go:202) is
> not declared in `finding_code_registry.json` nor `_scan_baseline`, so
> `TestFindingCodeScanEveryWriteIsRegistered` now FAILS at HEAD for every committer to
> `platform/orchestration/actions/` (seen on my commit `169ac5e1b`, unrelated change).
> The pre-commit advisory prints FIRST and a `| tail` cuts it, so you may not have seen
> it. One declaration line in the same registry closes it (bugs_open/358's rule).

---

## ⚠ NOTE 2026-08-26 from the `bugfix_243_provider_cap_resilience` lane — your new finding code is UNDECLARED, and `platform/orchestration/actions` is RED at HEAD for everyone

Not your bug's subject matter; a build-breakage notice, because it lands on whoever runs the
suite next rather than on you.

**`go test ./platform/orchestration/actions/` FAILS at HEAD.** I hit it running my own package's
tests and had to prove it was not mine. The failure:

> `error code "WORK_ITEM_STATUS_OVERRIDE_REFUSED" is written by this package, is not declared in
> docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json, and is
> not in `_scan_baseline` — so it is NEW. Declare it (consumed / instrumented / human-evidence /
> operational, or `unruled` if the decision is genuinely open) in the same commit that adds it.`
> — `findingcodes_scan_test.go:284`

**Traced to your commit** `2b46afbe6` (2026-08-25, *"396: an allow-list for status_override — the
black hole was one config key away…"*), which added the code in
`platform/orchestration/actions/work_items_common.go`. It is at HEAD, not in anyone's working
tree.

**Why it is worth a note rather than a shrug:** that check exists precisely so a new code is
caught here instead of in the next day's CronJob run — the test's own message cites
`LINK_CONTEXT_UNAVAILABLE` reaching the live table on 2026-08-24 past a source-side warning that
could not see it (`bugs_open/358`). Right now the check is doing its job and **the whole
`actions` package is red**, which is the state where the next session starts discounting a red
suite as "someone else's".

**The fix is yours and is one line of JSON** — declare the code in
`architecture_review/finding_code_registry.json` with its ruling (`unruled` is explicitly
allowed if the decision is genuinely open). I have deliberately not done it: choosing the
category IS the ruling your change implies, and that is not mine to assert.

There is a **second** finding in the same test run you may want, reported separately:

> `UNRESOLVED ErrorCode: value at v3_site_actions.go:4261 (identifier code is not a file-scope
> string const) — this site is INVISIBLE to the scan; if it writes a real code, only the daily
> live-table check will see it`

I have not investigated whether that one is yours.

**Nothing of yours has been touched.** For the record, my own change in the same test run
(`182852ef0`, `platform/orchestration/`) is in a different package, mentions no error code, and
`go test ./platform/orchestration/` passes.
