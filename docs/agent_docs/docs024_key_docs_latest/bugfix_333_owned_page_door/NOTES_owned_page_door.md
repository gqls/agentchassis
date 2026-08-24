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
