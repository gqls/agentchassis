# HANDOFF — 2026-08-11 — continue here

**Lane:** `diagnosis_schema_visibility` — commission item 5, owner ruling 2026-08-10.
**Read first:** `PLAN_2026-08-10_…` (design + why), `NOTES_…` (the missteps),
`RUNBOOK_…` (every command, with its gotcha).

---

## 1. State in one paragraph

Item 5 is **DONE, APPROVED, and LIVE**. The diagnosis loop's evidence bundle was
showing the verdicter rows from six tables while describing the columns of one;
the Schema section is now self-describing and always carries the tables the
gather reads. Council **APPROVED** round 1 (`df9dae6c`), its one real objection
closed. Pod-verified on **`agent-chassis:v1.0.1284`**, both replicas, with a
negative control. The remaining work is **not item 5** — it is reading the
outcome of the `090` re-run that item 5 unblocked, and then choosing the next
commission item.

## 2. What shipped

| commit | what |
|---|---|
| `5f8a326fc` | the fix: always-list, filter notice, derivation test |
| `15ca136ab` | ratchet line (not a registrable mechanism) |
| `436736212` | standing five + commission corrections + WRONG_CALLS |
| `55fc8fc35` | closed the nested-steps caveat |
| `e2afedaaf` | council response: the count-degradation test, `Council-Reviewed:` |
| `a41dec8e5` | owner-log entry |
| `83a44dd6a` | pod verification + the index trap |
| `3ea02ffa7` | LANDMINE: `code_symbols` stores methods receiver-qualified |

Three behaviours, all in `diagnose_load_runtime_action.go`:

1. `schemaAlwaysTables` — the tables this action renders rows from are listed
   whatever `schema_include_patterns` says. Beats the exclude denylist too, and
   sorts first so `schema_table_cap` truncation cannot reach them.
2. `schemaFilterNotice` — the section states its coverage (`31 of 433`), that
   absence is not non-existence, and that an unlisted table is readable via the
   existing read-only `data_request` channel **without a human**.
3. `TestSchemaAlwaysTablesCoverTablesThisActionQueries` — re-derives the list
   from the action's own SQL and fails when a query names an uncovered table.

## 3. THE 090 CAME BACK — item 5 VERIFIED, and it reordered the commission

**Run `90f6f55f-c014-4537-880c-0f1ae2b82e0b`, 2026-08-11. Verdict UNVERIFIABLE,
`stopped: iteration-cap`.** Read the two halves apart — one is success.

**Item 5 PASSES the commission's bar.** Across all five bundle iterations: no
`42703` anywhere; `orchestration_states(…)` in the Schema section; the FILTERED
notice present; and a real `data_request` against `orchestration_states` that
**executed** with the correct columns and returned `(0 rows)` — an answer, not an
error. The bar was explicitly *not* "the table appears in the bundle".

**The stop reason is the signal.** `074beb8a` stopped at `scope-not-narrowing` —
it could not proceed. This one ran five full iterations and hit the cap: not
blind, just nothing to find.

**Because the evidence expired.** `hero_deployed` **0**, `logo_deployed` **0**,
§1's decisive row `3e46be5b-…` **gone**. Third state this bug has been seen in
(0 of 1,667 → 2 each → 0). ⚠ **Not** a 24h-TTL claim: 2,110 rows are retained,
oldest 2026-07-13, so those particular orchestrations were pruned by something
nobody has chased. `[UNVERIFIED]`

### What follows — this is the part that changes the plan

**Item 1(a) is no longer a search problem, and another `090` would burn credits
reproducing this result.** The commission assumed item 5 would unblock item 1;
item 5 unblocked the *tool*, and the tool proved the *data* is gone.

**Item 2 is now item 1's unblocker** — which the commission did not anticipate.
Making the three silent readers log the key they wanted *and the keys the map
actually held* writes this evidence into `agent_error_log` at the moment of
failure, a store nothing prunes on an orchestration's schedule. The alternative
is forcing a fresh site build that deploys a hero/logo and reading
`collected_data` before the prune — a race against the thing that just erased the
last one. Written into `bugs_open/236` §5b so the fixing thread sees it there.

## 4. Decisions waiting on the owner

**None of these are blocked on me — they are choices.** Full framing was given in
chat 2026-08-11; the short form:

1. **Which commission item next.** Commission order is 5 → 2 → 1 → 3. Item 2
   (make the three silent hero/logo readers log) is hours and independent. Item 1
   is large and mostly investigation, and its next move depends on the `090`
   above. Item 3 spans three layers and needs a routing call (below).
2. **Item 3's routing: council gate or architecture RFC?** It changes a client
   return signature and adds a field to a shared adapter response payload. The
   commission says *"plausibly architecture-scope… if in doubt write the RFC —
   the cost is one document"*.
3. **Item 3's modelling question:** `deploy_commit` is per-component, but a page
   is many components deployed across possibly several commits. The column's
   original author never answered whether the page level wants it too.
4. **Item 1's design decision** is explicitly reserved to the owner (Design 1 vs
   Design 2), and the commission warns the census's *"0 breaks"* premise for
   Design 2 is contradicted by production — **the baseline must be re-measured
   before either design is scored.**

## 5. Live facts, dated — re-measure, do not trust

| fact | value | when |
|---|---|---|
| chassis image carrying item 5 | `v1.0.1284` | 2026-08-11 |
| public tables | 433 | 2026-08-10 |
| Schema section before / after | 26 / 31 tables | 2026-08-10 |
| bundle size | ~80,000 chars; item 5 adds ~3,175 (+4%) | 2026-08-10 |
| code index | 6,170 symbols, all with bodies, live branch | 2026-08-11 09:49 |
| council verdict | APPROVED, 1 med + 5 low, 2 abstained | `df9dae6c` |

## 6. Traps this lane paid for (the full versions are in NOTES)

- **A failed watcher says nothing about the thing watched.** My first verdict
  poll exited 1 having "concluded" the run was gone; the cluster had lost DNS and
  every query returned empty. The arm handling "no data" must distinguish
  *queried and found nothing* from *could not query*.
- **`code_symbols` stores Go methods receiver-qualified.** A bare-name lookup
  returns 0 rows and no error. Now a LANDMINE (`3ea02ffa7`) because the loop's
  cite-or-abstain rule acts on that absence.
- **A substring from a heading runs to the END of the document**, not the end of
  the section — that nearly cancelled half this fix. Bound it and print
  `length()`.
- **A keyword hit may be a column name** (`relevance_score` matched "relevance").
- **A source-scanning test reads prose.** Stripping comments is not enough;
  ordinary string literals carry English too.
- **Pick a positive control you have confirmed is present.** Mine
  (`schema_table_cap`) was absent too, so it proved nothing.

## 7. If you change this code

The always-list is enforced by a test that scans **backtick literals only**. If
you add a query to `diagnose_load_runtime_action.go`, write it as a backtick
literal — a double-quoted `SELECT…FROM` fails the guard on purpose, because the
scan cannot read it and would otherwise pass while blind.
