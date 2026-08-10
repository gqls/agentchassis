# NOTES — diagnosis bundle Schema section

Append-only, newest at the bottom. Missteps included on purpose — they are the
part the next thread cannot rederive.

---

## 2026-08-10 — session 1: found the producer, found it was bigger, shipped it inert

### Finding the producer (the commission's `[UNVERIFIED]` gap)

The commission guessed *"likely a `load_schema_hint`-style step or a Go
runtime-bundle builder"*. It is the latter: `gatherSchema` in
`diagnose_load_runtime_action.go`, returned under the `schema` key and rendered
by `diagnose_assemble_bundle_action.go:306`. One grep for `runtime.schema` found
the reader; the file listing (`ls | grep diagnos`) made `diagnose_load_runtime_action.go`
the obvious producer candidate and it was right first try. **~2 minutes** — worth
recording because the commission budgeted this as the unknown part of the job.

The mechanism was not an omission about `orchestration_states` specifically. It
is a **relevance include** — `site%|page%|content%|flow%` — that predates the
sections which read other tables. It was correct when written.

### The finding that resized the job

Checked whether the include covers the tables the gather **itself** reads:

```
agent_error_log|f   site_work_items|t   orchestration_states|f
agent_definitions|f llm_call_log|f      code_symbols|f
```

**Five of six.** 26 of 433 public tables in the section. So this was never "one
missing table" — the bundle was showing rows from six tables and describing one.
That reframed the fix from "add `orchestration_states` to a list" to "make the
list derive from what the action reads", which is what the commission had asked
for in the abstract (*"prefer deriving it"*) without knowing it was already
five-sixths wrong.

### MISSTEP 1 — a substring that measured the whole document

To test "does the section already say it is filtered?" I ran:

```sql
substring(collected_data->>'bundle' from position('## Schema' in ...))
```

with **no length argument**. That runs to the end of the ~80,000-char bundle, not
the end of the 8,819-char section. It returned `says_it_is_filtered = t`.

**Had I believed it, I would have dropped the notice half of the fix** — and the
notice is the half that generalises. The negatives in the same query were safe
(searching a superset and finding nothing still proves absence), which is *why*
the single positive was the dangerous cell. Caught it because a positive from a
query whose bounds I had not thought about is exactly the shape to distrust.

Bounded version, and the answer inverted: no match for "filter", "not
exhaustive" or "truncated" anywhere in the section.

### MISSTEP 2 — the keyword hit that was a column name

The bounded query still reported `says_relevance = t`. Printed the surrounding
130 chars: `relevance_score float8` — **a column in the listing**, not a notice.
A keyword search over a document that *contains a database schema* will match
schema words; every hit needs its context printed before it counts.

### MISSTEP 3 — my source-scanning test scored prose as code

`TestSchemaAlwaysTablesCoverTablesThisActionQueries` re-derives the table list
by scanning the action's SQL. First run failed with: a table called **`the`**.

Source: `"Code questions this diagnosis asked, answered from the code_symbols
index\n"` — an ordinary string literal. `\bFROM\s+(\w+)` matched `from the`.

I already knew comments were load-bearing here and had stripped them. **Stripping
comments was not enough: ordinary `"…"` string literals carry prose too.** The
tempting fix — add `"the"` to the not-a-table list — would have made the guard
weaker with every sentence anyone added to this file. Scanned **backtick literals
only** instead, which is where every SQL statement in the file lives.

### MISSTEP 4 — and then the blind-spot guard caught my own code

Scanning backticks only creates a new hole: SQL written in double quotes is
invisible, so the test would **pass while blind**. Added a guard that fails if
`SELECT…FROM` appears in a double-quoted string.

It immediately failed — on `gatherSchema`'s own query, which **I had just written
that way**. Real positive, my code, fixed by moving the query to a backtick
literal. (The first version of that guard matched bare `FROM` and tripped on the
same prose as misstep 3; requiring `SELECT` before `FROM` discriminates.)

Worth the three iterations: the guard is now the thing that stops the next
evidence section reintroducing this bug, and it has demonstrated it can fail.

### Proving the guard, rather than trusting a green run

Mutated the code — removed `orchestration_states` from the always-list — and both
tests failed with the intended message. Restored, re-ran, green. A guard that has
never been seen to fail is not evidence.

### Live checks before committing

- New SQL run verbatim against the live DB: **31 tables** (was 26), all six
  evidence tables present. The mock proves shape; only Postgres proves it parses.
  (`EXECUTE` cannot be wrapped in a subquery — had to inline the literals.)
- All four live agents with a `diagnose_load_runtime` step override **none** of
  the schema keys → the Go defaults are what production runs.
- One non-test caller of `gatherSchema`.
- Size: +3,175 chars on an ~80k bundle (+4%); cap is 120 tables, listing is 31.

### State at end of session

Committed `5f8a326fc` with `Council-Submitted: df9dae6c-b7ca-4605-8dd4-26462ce4b20b`.
Ratchet line committed `15ca136ab` (the pattern check's register-blind-spot flag;
nothing here is callable by another workstream, so it ratchets rather than
registers).

**INERT.** Go changes need an image, releases are whole-fleet and the owner runs
them. The commission's real verification — re-run a `090` whose evidence lives in
`orchestration_states` and confirm its `data_request` executes instead of 42703 —
is blocked on that roll. Nothing about this work should be described as working
in production yet.

> One flag on the commit is **not** from this work: pattern-check's
> `logged-model-output` at `diagnose_load_runtime_action.go:669`. That line is
> from commit `68933d0c9` and writes SQL result rows, not model output; my edit
> only shifted its line number under the checker.
