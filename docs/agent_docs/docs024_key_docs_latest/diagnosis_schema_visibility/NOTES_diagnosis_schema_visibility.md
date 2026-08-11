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

### Later, same session — closing the caveat I had flagged but not resolved

The RUNBOOK warns that the `jsonb_each` over `default_config #> '{workflow,steps}'`
walks **top-level steps only** and is blind to a step nested in a loop's
`sub_workflow` (the commission records that shape returning 3 of 6 elsewhere). I
had used exactly that shape to claim "no live config overrides the Go defaults"
— and then put that claim in the council submission. Flagging a caveat is not
closing it, so I re-ran it a second way that **can** see nested steps:

```sql
SELECT type,
       (default_config::text LIKE '%schema_include_patterns%'),
       (default_config::text LIKE '%schema_exclude_patterns%'),
       (default_config::text LIKE '%schema_full%'),
       (default_config::text LIKE '%diagnose_load_runtime%')
FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false
  AND default_config::text LIKE '%diagnose_load_runtime%';
-- 4 rows; f|f|f|t on every one
```

Same four agents, none of the three keys anywhere in the JSON at any depth. The
two methods are blind in **different** ways and agree, which is the version of
agreement that means something.

**A misstep inside the check itself.** My intended positive control was
`schema_table_cap` — "does this LIKE match a key I know exists?" It returned
**0**, because that key is not configured anywhere either. So the control proved
nothing about the LIKE; a zero from a broken query and a zero from an honest
absence look identical, which is the whole reason to run a control. The control
that actually works was already in the result: the `diagnose_load_runtime` column
returns `t` on all four rows, so the text scan demonstrably matches when there is
something to match. **Pick a control you have independently confirmed is present
— not one you assume is.**

### 2026-08-11 — the council verdict, READ (not just its decision field)

**APPROVED**, round 1, `df9dae6c`. 1 medium + 5 low advisory objections, none
high-severity, 2 seats abstained. An approval is not a reason to skip reading it
— three objections were worth acting on and are now closed.

> **A watcher trap fired first, and is worth recording.** My background poll for
> the verdict exited 1 and its loop had already broken out with
> `terminal-or-missing:` and an EMPTY status. That was **not** the council
> finishing: the cluster API had a transient DNS failure (`server misbehaving`),
> every `kubectl` in the loop returned empty, and my `case ""` arm read that
> connectivity failure as "the row is gone". **A failed watcher is not evidence
> about the thing being watched.** Re-queried directly once connectivity
> returned; the run had in fact completed at 22:39 the previous evening. Same
> family as the `||true` watcher entry in MEMORY — the arm that handles "no
> data" must distinguish "queried and found nothing" from "could not query".

**1. `editquality`, MEDIUM — "no sketch shows `schemaFilterNotice` being invoked
or its output concatenated into what `gatherSchema` returns."** Correct about the
*submission*, wrong about the *code*: the wiring is
`return schemaFilterNotice(tables, total, full) + sb.String()`, and
`TestGatherSchemaAlwaysListSurvivesTheFilters` already asserts `2 of 433` appears
in the returned string. The seat said as much — *"likely just sketch
incompleteness rather than a design flaw, hence object not veto"*. **The lesson
is about submissions, not code: a sketch that omits the call site invites exactly
this objection, and the reviewer cannot tell the two cases apart.** Show the
wiring line next time, even when it is one line.

**2. `diagnosis_guardian`, `missing` — "confirm the count query degrades to
silent notice-omission rather than failing `gatherSchema`; plan claims this but
no test enumerated for it."** **A fair hit: the claim was true and untested.**
Added `TestGatherSchemaSurvivesTheCountQueryFailing` — the count errors, the
gather still returns nil error, the listing still renders, and no notice appears
(neither `FILTERED` nor a ` of 0` denominator). Proven by mutation: changing the
degradation to `return "", err` fails it. This is the "observability never costs
a diagnosis" invariant, now pinned rather than asserted.

**3. `reuse_agent` + `editquality` — no evidence a `[]string -> []interface{}`
helper or an "always-include regardless of filter" convention was searched for
before adding one.** Also fair — I had grepped for `toIfaceSlice`/`toInterfaceSlice`/
`asIfaceSlice` but never wrote the result down, and never searched for the
always-include pattern at all. Both searched properly now and both came back
empty: no function anywhere in `platform/`, `internal/` or `pkg/` matches the
signature `func f([]string) []interface{}` except mine, no name collision, and no
`alwaysInclude`/`forceInclude`/`includeAlways` convention exists. **Recording the
search is part of the search** — an unrecorded grep is indistinguishable from no
grep, which is precisely what the seat objected to.

**Not actioned, deliberately:** `guardian`'s "bundle-size growth has no enforced
ceiling" (true; `schema_table_cap` bounds table COUNT, not chars — but the
always-list is six entries and the measured cost is +4%, so a char cap would be
machinery for a bound nothing is near), and `tooling_provenance`'s doc_notes
hygiene point (the standing five exist in this lane; the travelling-docs
`doc_notes` channel is a different surface and adding a row for a point fix would
dilute it). Both recorded here rather than silently dropped.

**`architecture` asked for one thing to be written into the COMMISSION**, and it
has been: the notice is instructive prompt text, approved as a `point_fix`
because it documents an EXISTING read-only channel — but *"if this phrasing
pattern gets reused across other diagnosis actions it could accumulate into a de
facto shared vocabulary without ever passing through architecture review."* The
**second** action to do this needs an RFC; this one is the precedent that makes
it the second.

### 2026-08-11 — LIVE on v1.0.1284, pod-verified, and the code-tier blocker checked too

**The roll carries it.** `agent-chassis:v1.0.1284`, both replicas (pods 23m old at
check). Grepped with a positive AND a negative control, per the standing rule
that a roll is not evidence your fix shipped:

| grep | 6j5xn | rvrdg | means |
|---|---|---|---|
| `This listing is FILTERED, not the whole database` | 1 | 1 | POS — the notice is compiled in |
| `you do not need a human to confirm it` | 1 | 1 | POS — the self-serve sentence |
| `schema_always_tables` | 1 | 1 | POS — the config key |
| `data_type FROM information_schema.columns WHERE ` | **0** | **0** | **NEG — the OLD query literal is GONE** |
| `no orchestration rows for this correlation/site` | 1 | 1 | CTRL — an untouched sibling literal, proves the grep pipeline |

The negative control is the one that matters: the pre-change binary built that
query by `+`-concatenating constants, so `…columns WHERE ` was a real literal in
it. Zero on both pods means this is the new binary, not a lucky match.

### The OTHER blocker on the 236 run — checked before spending credits

Re-reading `074beb8a`, it named **two** harness failures, and this change fixes
only one. The other: it could not read the bodies of `storeActionResult`,
`processAwaitResponse` and `applyResponseToState` — *"failed to load from
coordinator.go in this checkout"*. Firing a re-run without checking that would
have risked a second UNVERIFIABLE for the unfixed reason and taught us nothing
about this fix.

Checked: the index is fresh (6,170 symbols, **all** with bodies, `ref` = the live
working branch, updated today 09:49). All three functions are present **with
bodies**, including the merge function at the heart of the bug —
`(*SagaCoordinator).applyResponseToState`, 4,746 chars, and its body really does
contain `existingData` and `response_received_at`, the merge logic §5 quotes. So
both blockers are clear and the re-run is worth its credits.

> **MISSTEP — I nearly filed "the merge function is missing from the index".**
> My first query was `WHERE symbol IN ('applyResponseToState', …)` and it
> returned **nothing**, which looked like a decisive second blocker. It is a
> **method**, and the index stores methods under their receiver-qualified name:
> `(*SagaCoordinator).applyResponseToState`. A bare-name exact match cannot find
> it. This is the SAME family as yesterday's four wrong calls — the query
> encoded an assumption (that symbols are stored bare) and its zero answered
> *that* question, not the one I asked. What caught it: checking the source
> before believing the absence, because a zero from an index is exactly the
> shape that should never be trusted on its own. Sanity check that would have
> caught it instantly: `count(*)` for the file was **92** against **91** source
> funcs, so nothing was missing at all.
>
> **This is a real trap for anyone writing a `code_request` too**: searching the
> index for a Go method by its bare name returns nothing, and the loop's own
> cite-or-abstain rule acts on absence.

**090 re-run fired:** `RUN_CORRELATION_ID=90f6f55f-c014-4537-880c-0f1ae2b82e0b`.
Symptom names the mechanism and points at the tables/symbols, asserts no counts,
and names §5's refutation as context so the loop is not asked to confirm a story
already refuted.
