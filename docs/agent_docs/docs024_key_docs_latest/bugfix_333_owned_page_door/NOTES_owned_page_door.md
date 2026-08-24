# NOTES — 333 owned-page door (append-only, newest at the bottom)

## 2026-08-24 — lane opened, bug re-validated, shape decided

**Ownership first.** `scripts/who-owns.py 333` says OWNED-or-recently-active, but reading what it points at:
the three lanes that touched the file are `bugfix_277_required_fields_repair` (CLOSED 08-22),
`bugfix_367_router_remit` (their change is migration 574, config only, and their own CONTRIB says it does
not touch `writeWorkItem`), and `bugfix_301_owned_guard_ordering` (CLOSED, this bug is its residual). No
`bugfix_333_*` directory existed. The bug file's own header says **"Status: OPEN, unowned."** So: resumed here.
⚠ `who-owns.py` reads COMMITS — it cannot see a session mid-fix. Checked the tree too: every file this lane
will touch was clean at HEAD `0bce1db39`.

**Bug still valid [MEASURED 2026-08-24 14:15Z, live table].** Not a stale filing:

| what | figure |
|---|---|
| `wont_fix` owned-page refusals at `page-build-handler` since 08-19 | **83** (5 today) |
| new owned-page filings at that handler since 08-19 | **78** — `tool-generator` 64 (last **11:48Z today**), `backfill-353` 8, `offer-analysis` 3, `internal-linker` 2, `generic` 1 |
| legacy open rows | 59 failed / 36 unresolved / 16 needs_human_review |
| owned pages fleet-wide | **176** on 13 sites, 96 named `tool-*` |
| of 88 refused rows since 08-19, carrying `page_id` | **83** (5 without — those are the `spec.page_name`-only producers) |

The `tool-generator` filings are almost all one site: webdesign.co.uk 63, loanandmortgagecalculator.co.uk 1.

**Which handlers actually refuse an owned page** [MEASURED 2026-08-24, live + archive, all history] — this is
the positive control that made the "read the declaration" design possible, because it could have come out
otherwise:

| handler | on owned pages |
|---|---|
| `page-build-handler` | 112 failed / 85 wont_fix / 74 complete (pre-guard) / 36 unresolved |
| `page-rerender` | **5,040 complete** |
| `section-editor` | 44 complete / 1 failed |
| `tool-generator` | 43 complete |

And exactly one live agent declares the opt-in: `SELECT type FROM agent_definitions WHERE
jsonb_path_exists(default_config,'$.workflow.steps.*.config.refuse_owned_page ? (@ == true)') …` → **one row,
`page-build-handler`**, with or without the `is_active`/snapshot filters.

### MISSTEP 1 — I wrote a query against a shape I had not checked

`jsonb_array_elements(default_config->'workflow'->'steps')` → `ERROR: cannot extract elements from an object`.
`steps` is an OBJECT keyed by step name, not an array. Cost: one round trip. **The check:** `SELECT
jsonb_typeof(default_config->'workflow'->'steps')` before writing any query over it — census run afterwards:
**194 live agents object, 5 absent, 0 arrays.** That census is now load-bearing (the door's jsonpath assumes
the object shape), so it is recorded here rather than left in scrollback.

### MISSTEP 2, and it is the one that matters — I nearly shipped a park that nothing could ever close

My first plan followed `bugs_closed/077`'s convention **literally**: re-type the demoted row to
`capability_gap`, key it `capability_gap:owned_page:<original key>`. It had precedent, a live consumer, and
two council seats on `bugs_open/342` endorsing exactly that shape a day earlier. It was still wrong.

A red-team pass over the plan found it: **a detector retracts its own finding by matching
`(item_type, item_key)`** — `resolveWorkItems`, `work_items_common.go:443-457`, and `deferred` is NOT in
`workItemClosedStatuses`, so a keep-type parked row IS retractable. A re-typed row matches no retraction at
all. Once the page were fixed by another route, the parked row would sit at `deferred` for ever holding its
dedup slot — the "nothing swept it, the items age for ever" hole those same seats caught on 342.

I had written *"`deferred` … retraction counts it"* in the plan as if it were a property of the status. It is
a property of the **type and key**, which I was about to change in the same breath.

**The cheap check that would have caught it:** before re-typing or re-keying a row, grep for who CLOSES rows
of that type — `resolveWorkItems`, `loadAuditRetractionCandidates`, the revalidator's `coveredItemTypes()` —
and ask whether the new identity still matches. One grep. **Transferable half: reusing a convention's SHAPE
is not reusing its LIFECYCLE.** 342's landmine says "reusing a type is not reusing its contract"; this is the
same trap from the other end — I was reusing the contract and breaking the identity the contract keys on.

Not promoted to `WRONG_CALLS.md`: the claim never left the plan file, and that file is this session's
scratch, not a shared doc. Recorded here because the *next* person to reach for 077's shape needs it.

### Other findings from the same red-team pass, all acted on

- **Existing seam tests would pass VACUOUSLY.** The door adds queries before the INSERT; sqlmock in ordered
  mode returns "was not expected", and the door's fail-open swallows exactly that error and inserts anyway.
  So `tool_content_item_test.go`, `work_item_created_honesty_test.go`, `nav_rebuild_request_test.go`,
  `tool_cross_link_items_test.go` and `write_render_audit_findings_test.go` would stay green while proving
  nothing. They are re-scripted, and the discriminating negatives in the new test file call
  `ExpectationsWereMet` so an unconsumed probe FAILS rather than passing quietly.
- **`recurrenceExpected` must be TRUE on the parked row.** Retraction sets `complete`, which the two-strike
  block counts as a strike; two retractions in 7 days would brand the third finding `unresolved` — 333's own
  loop under a new label.
- **Probe order is declaration-first.** The `agent_definitions` probe is indexed and cheap and gates the
  `pages` read, so a generic-page insert at a non-declaring handler pays one indexed lookup, not two.
- **`ErrNoRows` on the page means "not owned", not "unreadable"** — at the door a stale `spec.page_id` is
  ordinary. The 208/assemble guard keeps its stricter posture; only the door's reader is lenient.
- **Fail-open inside a transaction is partly fiction.** A server-side SQL error aborts the tx and the INSERT
  then fails anyway. The comment says so rather than claiming a safety the code does not have.

**Probe cost [MEASURED 2026-08-24, EXPLAIN ANALYZE on the live DB].** Declaration probe: index scan on
`idx_agent_definitions_type_version`, **0.278 ms** warm / 4.2 ms cold, 33 buffer hits (216 rows in the table).
Page policy read: `pages_pkey` index scan, **2.7 ms** execution.
⚠ That second measurement took **16 minutes of wall clock** to come back, and none of it was the query: a
`psql` session was 52 minutes into a base64 `COPY` over `page_components`, an `ALTER TABLE pages ADD COLUMN
noindex` was queued behind it, and every `pages` read in the fleet was waiting on that lock. **Execution time
is what the door pays; wall clock that day measured someone else's DDL.** Flagged to the owner in
`README_where_we_are.md`.

## 2026-08-24 — built, tested, committed, submitted

**Committed `6ab0b3434`** (code + tests + lane docs) and **`68734b771`** (register WII-028, two landmines,
016b §9, the bug file). Council round dispatched, corr `9813dec8-5ce1-48ab-bb77-e3f601f9f64c`; commit carries
`Council-Submitted:`, not `Council-Reviewed:` — **the verdict is not read yet and must not be claimed.**

### What is in the code

Door at the TOP of `writeWorkItem`, before the anti-churn brake, behind `DISABLE_OWNED_PAGE_DOOR_DEMOTION`
(ships ARMED). Declaration probe first (indexed, 0.278 ms), page policy read second — so the common case pays
one lookup and never touches `pages`. On a hit, `ownedPageParkedItem` (pure) rewrites the item; the error text
rides the existing conditional `extraCols/argN` slot, so 16-arg callers are byte-identical.

### Mutation proofs — both directions, run, not asserted

| mutation | result |
|---|---|
| `if false && …` on the door's guard | positives FAIL (row lands at requested status with the handler intact) |
| re-type + re-key the parked row | `TestOwnedPageParkedItem_KeepsIdentityTakesTheSignal` FAILS on both item_type and item_key |

Tree restored from a scratchpad copy after each; package green at the end of both.

### MISSTEP 3 — my parity test resolved migration 488 by NUMBER, and there are two 488s

`TestOwnedPageRefusalPathMatchesMigration488` matched `488_` and took the first hit, which is
`488_meta_description_backfiller_agent.sql` — a completely unrelated migration. It failed with a message I had
written asserting that "migration 488 no longer writes the path", which was false about the migration it named.
Caught on the first run. **Had the directory sorted the other way it would have PASSED by luck.** Fixed to select
on the SLUG (`refuses_owned_pages`) and exclude `_ROLLBACK`/`_VERIFY`/`_SUPERSEDED`. The rule was already in
CLAUDE.md — *"a bare number is ambiguous … resolve by slug"* — and I had read it this session. Filed in
`WRONG_CALLS.md`; the cheap check is `ls docs/agent_docs/sql_for_agents/ | grep '^488'`, which prints four files.

### MISSTEP 4 — I inserted the door's test expectation after EVERY `ExpectBegin()`

Blanket insertion by script across 8 files broke 4 tests whose FIRST transaction is a page-conflict resolve that
never reaches the write seam (`apply_gap_plan`). The helper belongs after the `ExpectBegin()` of the transaction
that actually contains the `INSERT INTO site_work_items`. Reverted with `git checkout --` (the 8 files were
clean, verified first) and re-applied per test function, then per transaction. The placement rule is now in the
LANDMINES entry so the next person does not rediscover it.

### The vacuity sweep — the number, and how it was got

Instrumented the door's probe-error branch with a **print** (not a panic — a panic aborts the run at the first
offender and you fix one and believe you are done), ran `go test -v` over the whole package, and collected test
names with `awk '/^=== RUN/{n=$3} /VACUITYPROBE/{print n}'`. **21 test cases across 8 files** were reaching the
door with an unscripted probe and passing anyway. All re-scripted; two shared helpers left behind for the next
author. Instrumentation removed and the tree rebuilt before committing.

### Another lane's work in the tree, twice

1. `flag_page_image_rebuild_action.go` is dirty with `+4` lines adding `recurrenceExpected: true` (the comment
   cites `bugs_open/326`), which makes the two-strike COUNT unreachable while
   `TestFlagPageImageRebuild_PlanMemberWithoutSections_StillEmits` still expects it. **Not mine**, proved by
   re-running with my door disarmed via the kill switch — still fails. Not committed; told the 326 lane.
2. The `bugs_open/326` session messaged mid-work to sequence: their deferral change edits the same three regions
   of `writeWorkItem`. Agreed option 1 (I land first), told them the exact shape of my hunks, and messaged again
   at commit. Their `retry_after` and my park cannot interact, because a parked row sets `recurrenceExpected` and
   so skips the brake entirely.

### Verification against HEAD, not against the working tree

`scripts/verify-head-builds.sh --test --with <16 files> ./platform/orchestration/...` → **OK against HEAD
`f1f0adb8f`**, all `platform/orchestration/*` packages green. This is the check that separates my change from the
two other lanes' WIP sitting in the same tree. ⚠ Running it without a package argument fails on `test/website_builder`
(Kafka at localhost) and a pre-existing `test/unit/orchestration` build failure — neither is a statement about
the change under test, and reading it as one would have been a false red.

## 2026-08-24 — council round 1: REVISE, and it was worth the round

Verdict `9813dec8`: **REVISE**, 15 reviewers, 2 abstained, `decided_by: gating objection from editquality`.
Round 2 resubmitted on the same correlation (run `7efd5471`). **Two objections changed the code; five more were
checkable and were CHECKED rather than argued.** The gating one asked a question I had not asked myself.

### The gating HIGH — and it was RIGHT about the column and WRONG about the coverage

`editquality`: the door keys on `item.pageID != nil`, i.e. on the `site_work_items.page_id` COLUMN, and a
standing landmine says that column is NULL on most rows even when the page exists. If the ~26 producers mostly
identify their target by `affected_url`/`entity_id`/`spec` instead, *"the door silently skips the majority of the
population the diagnosis measured"*.

**[MEASURED 2026-08-24, live + archive, all history at `page-build-handler`]** — 3,819 of 5,259 carry `page_id`
(**72.6%**), so the landmine is right. But the split by producer is the answer:

| producer | rows | with page_id |
|---|---|---|
| design-audit | 1,508 | 97% |
| completeness-discovery-agent | 597 | 100% |
| content-gap-planner | 433 | 100% |
| generic (discovery) | 247 | 100% |
| tool-generator | 230 | 100% |
| tool-suggester / tool-deployer / backfill-353 / quality-discovery / offer-analysis | 126 / 77 / 74 / 57 / 41 | 100% |
| **page-rerender** | 399 | **0%** |
| **reconcile_site_plan** | 384 | **0%** |
| **image-build-handler** | 262 | **0%** |
| render_directory / render_news_section / plan_sections | 92 / 80 / 33 | **0%** |

The 0% producers are a **different kind of item**: name-only ACTION REQUESTS filing `needs_page` for a page
identified by name. The content findings this bug is about carry the column ~100% of the time. On the measured
defect population the door sees **83 of 88 (94%)**, which matches the 08-19 figure exactly. Of the 1,440
`page_id`-NULL rows, **1,438 carry `spec.page_name`** — so the gap is bounded, quantified, and now a named
non-scope item instead of an unstated assumption.

The same objection's medium — *does the door's column match the join behind the [MEASURED] figures?* — is
answered by construction: the figures join `pages p ON p.id = w.page_id`, the column the door reads.

**What I take from this:** I had the 83-of-88 measurement in my own plan and did NOT put it in the submission.
The reviewer could not see it, asked exactly the right question, and the answer took one query I should have
included. **An objection about coverage is cheap to pre-empt and expensive to leave to a round.**

### The two that changed the code

- **`guardian` [medium] — the novel SQL was on the hot path.** The declaration probe (`jsonb_path_exists` over
  `default_config`) ran for every page-bearing dispatchable write through a seam ~26 producers share; a bad
  jsonpath on an older Postgres or a permissions change aborts the transaction for **all** of them.
  **Inverted: the boring statement now gates the novel one.** A `pages` PK read runs first; the jsonpath only for
  an owned page (176 estate-wide). Cost: 2.7 ms PK read on the common path instead of a 0.278 ms index scan.
  ⚠ I had chosen the original order on COST grounds and never weighed FAILURE MODE. Both probes are cheap; only
  one of them is novel, and novelty is what belongs off a shared hot path.
- **`bug_historian` [low] — fail-open was uncountable.** Both stand-down branches now log the stable literal
  `OWNED_PAGE_DOOR_PROBE_FAILED`. Deliberately not an `agent_error_log` row: that write rides the same
  transaction, so the one failure it must survive is the one that would break it.

### The five that were checkable — all checked, one refuted a claim of mine

- **`prior_art_librarian` [medium]: the jsonpath is blind to `sub_workflow`-nested config, so "exactly ONE live
  agent declares it" may be false.** REFUTED by the widest probe available: `$.**.refuse_owned_page ? (@ == true)`
  with no `is_active` and no snapshot filter returns the **same single agent**. The positive control survives a
  probe that could not have missed a nested declaration. (The seat was right that I had asserted it on the narrow
  path only — the check was one query.)
- **`guardian` [medium]: does anything switch on `raiseToolContentItem`'s new return string?** No. Both call
  sites assign it and pass it into a result map key `content_item` — verified by grepping every use of the
  variable, not just the call.
- **`reuse_agent` [medium]: is there an existing generic "handler declares boolean flag X" helper?** No —
  grepped `jsonb_path_exists` and `workflow,steps` across `platform/`, `internal/`, `cmd/`. First of its kind;
  the register now says a SECOND flag must parameterise it rather than copy it.
- **`constitution` [low]:** `handlerExpr` is a placeholder or column reference at every call site, never user
  data — same posture as the helper it mirrors.
- **`architecture` [medium]:** `deferred` is now overloaded a third way. Recorded in the register in the seat's
  own terms: **a fourth overload must introduce a real reason-code column, not another string convention.**

### The one I answered with a record rather than code, and said so

`bug_historian` [medium]: the 9 raw-INSERT bypassers keep the identical failure mode, and *"disclosure in the
risks section is good practice but is not a fix"*. Agreed — it is not. It is now a named non-scope item with its
measured size, and the register lists the bypassers BY PATH with the warning that the set grows unobserved. The
real remedy is a promoter-side backstop for the ones born `detected`, and taking it inside a REVISE on the seam
itself would widen a shared-seam change mid-review.

### A demand-control warning from a consumer, which I did not have

The `staged_component_build`/353 lane replied to my notification with a measurement that changes how I must read
my own verification: **cross-link emission has been at ZERO since 08-21** (13 tool births 08-22→08-24, 13 of 13
emitted nothing — 8 stopped before the emitter's Guard 2, 5 at it), because `add_tool` specs carry
`related_pages` only when `tool-suggester` wrote them. So **an empty parked bucket after the roll will not mean
the door is inert** — on the current producer mix nothing reaches that emitter's write at all. Their stated
discriminating setup: an `add_tool` item whose spec DOES carry `related_pages` naming an owned page. Folded into
the verification section.
