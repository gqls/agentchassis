# PLAN — let the diagnosis loop see `orchestration_states` (commission item 5)

**Authority:** owner ruling 2026-08-10, item 5 of
`rfc012_await_findings/COMMISSION_2026-08-10_owner_rulings_five_pieces.md` —
*"yes, let diagnosis loop see the orchestration states file."*

**Why this item first:** the commission's own suggested order is 5 → 2 → 1 → 3,
because **item 5 unblocks item 1's research**. Item 1(a) — where `image_url` is
lost between the writer and the three hero/logo readers — is a question whose
evidence lives in `orchestration_states`, and two `090` runs on it returned
UNVERIFIABLE *for reasons unrelated to the hypothesis*.

---

## 1. The problem, as found

The commission stated the symptom and marked the cause `[UNVERIFIED]`:

> ⚠ **`[UNVERIFIED]` — I did not establish what populates `runtime.schema`.** I
> looked and did not find it in the `diagnose-agent` step configs by the obvious
> queries; that is a gap in my check, not evidence of absence. **Find the producer
> first** — likely a `load_schema_hint`-style step or a Go runtime-bundle builder
> — and change the table list there. Do not patch the renderer.

**The producer is `gatherSchema`** (`diagnose_load_runtime_action.go:763` as
found), returned under the `schema` key at the end of `DiagnoseLoadRuntimeAction`
and rendered by `diagnose_assemble_bundle_action.go:306`. The commission's guess
("a `load_schema_hint`-style step") was wrong; it is a Go helper in the gather
action itself. Not a config change, so it needs an image.

**The mechanism is a relevance include**, not an omission anyone made
deliberately about `orchestration_states`:

```go
"schema_include_patterns": []interface{}{"site%", "page%", "content%", "flow%"},
```

### The finding that resized the job

The commission framed this as one missing table. It is not. Measured against the
live DB, 2026-08-10 (query in the RUNBOOK):

| | |
|---|---|
| public tables | **433** |
| tables the include selects | **26** |
| tables this action **renders evidence rows from** | 6 |
| …of those, present in the Schema section | **1** (`site_work_items`) |

So `agent_error_log`, `orchestration_states`, `agent_definitions`,
`llm_call_log` and `code_symbols` were all absent. **The bundle showed the
verdicter rows from six tables while telling it the columns of one.**

### The second half, which the commission did not name

The section is headed `## Schema (live tables)` and simply stops. It never says
it is filtered. Verified on the most recent stored bundle: its 8,819-char Schema
section contains no occurrence of "filter", "not exhaustive" or "truncated" —
the only match for "relevance" is a **column name**, `relevance_score float8`.

So **a filtered-out table and a non-existent table render identically**, and the
verdict prompt's cite-or-abstain rule acts on absence. That is the same
empty-vs-absent trap the code tier already guards (`codeEvidenceLine`,
`bodyCoverageNote`) — the schema tier just never got the same treatment.

This is what actually stopped run `074beb8a`. It did not merely fail to find the
table; it concluded the table's identity was **unknowable without a human**:

> *"the `orchestration_states` table isn't in the bundle's Schema section, so its
> real primary-key/id column is unknown and must be confirmed by a human before
> requerying; guessing again would likely fail the same way."*

Adding the table without adding the notice would have fixed this one run and
left the next missing table to be discovered exactly the same expensive way.

---

## 2. Design, and why not the obvious version

The commission was explicit about the trap:

> a hard-coded table list is the same drift class this estate keeps filing … If
> the list is a literal, adding `orchestration_states` fixes today and leaves the
> next missing table to be discovered the same expensive way. **Prefer deriving
> it.**

### Decision 1 — derive the list from what the action reads

The always-list is **the tables this action itself renders rows from**. That is a
real derivation rather than a taste judgement: the evidence sections and the
schema section are driven off one list, so the bundle cannot show a row from a
table whose columns it hides.

It is enforced, not just documented. `TestSchemaAlwaysTablesCoverTablesThisActionQueries`
re-scans this action's own SQL and fails when a query names a table the list does
not carry. **That test is the anti-drift mechanism; the list is just its output.**

> **Why not derive it at runtime** (parse the SQL, or introspect what was
> queried)? Because the failure mode would move from "a test fails in CI" to "the
> bundle is quietly wrong again", which is the thing being fixed. A compile-time
> list with a test that re-derives it fails **loudly and early**.

### Decision 2 — the always-list beats BOTH filters, and sorts first

Not just the include. If it only beat the include, adding `%log%` to the
denylist would silently drop `agent_error_log` and `llm_call_log` and
reintroduce this bug. It also sorts first, so `schema_table_cap` truncation
cannot reach it (31 of a 120 cap today, but the ordering is what makes that
safe rather than lucky).

**Open point for review, disclosed in the submission:** this means the
always-list overrides an operator's explicit denylist. Deliberate — these are
the tables the bundle shows rows from — but it is a judgement a reviewer may
disagree with.

### Decision 3 — say that the listing is filtered

`schemaFilterNotice` states the coverage (`31 of 433`), that absence is not
non-existence, and — the load-bearing sentence — that an unlisted table can be
read **through the read-only `data_request` channel that already exists**,
without a human. That is precisely the step `074beb8a` declined to take.

Suppressed when `schema_full` is set or nothing was withheld, so the notice can
never be a false claim about the listing. If the total-count query fails it
degrades to **silence**, never to a fabricated denominator.

### Decision 4 — scope routing: council gate, not architecture review

Checked against `PROCESS_architecture_review.md` as worded (owner ruling
2026-08-10 item 4, *"it **adds**, changes or removes an exported symbol other
packages depend on"*): every symbol here is **unexported and package-local**
(`gatherSchema`, `schemaAlwaysTables`, `schemaFilterNotice`, `stringsAsIface`);
the new config key is one action's step config, not a shared vocabulary; and
nothing changes what a shared mechanism **guarantees** (owner ruling 2026-07-29
§1). → council gate. Submitted `df9dae6c-b7ca-4605-8dd4-26462ce4b20b`.

---

## 3. What is NOT in scope

- **Not** a change to diagnosis logic, the verdict prompt, or the two-evidence-
  family guard. This is observability of the evidence corpus only.
- **Not** item 1. Establishing where `image_url` is lost is the *point* of
  unblocking the loop, but it is a separate piece of work and the commission
  wants a re-measured baseline first (its §1(b) warning that the census's
  "0 breaks" premise is contradicted by production still stands).
- **Not** the other diagnosis tiers. `code_symbols` is now described in the
  schema section, but whether the code tier has an analogous blind spot is
  unexamined. `[UNMEASURED]`

---

## 4. Verification — and why the obvious check is not enough

The commission set the bar and it is deliberately disconfirmable:

> **Verify — and make it disconfirmable.** Do not just check the table appears in
> the bundle text. **Re-run a `090` whose evidence lives in `orchestration_states`**
> and confirm its `data_request` now executes instead of erroring 42703.

| step | state |
|---|---|
| unit + behavioural tests green, guard proven by mutation | **DONE** |
| new SQL runs live, returns 31 tables incl. all six evidence tables | **DONE** |
| committed | **DONE** (`5f8a326fc`) |
| council verdict read and acted on | **DONE** — `df9dae6c` APPROVED; its one real gap (the count-degradation test) closed in `e2afedaaf` |
| chassis image built + rolled | **DONE** — `agent-chassis:v1.0.1284`, 2026-08-11 |
| pod-grep a literal this change ADDED, and a negative control | **DONE** — both replicas; the OLD concatenated query literal greps **0**, and an impossible-string sanity control also 0. Two mechanisms (`strings`, then `grep -a /proc/1/exe` after CLAUDE.md retired the first) |
| **re-run the `090`; its `data_request` executes, not 42703** | **DONE — PASSED.** Run `90f6f55f`: no `42703` in any of 5 iterations; the `data_request` against `orchestration_states` executed with the correct columns and returned `(0 rows)` |

**VERIFIED 2026-08-11. Item 5 is complete on the bar the commission set.**

> ⚠ The run's *verdict* is still `UNVERIFIABLE` — but on `iteration-cap`, not
> `scope-not-narrowing`, and for a reason that is not this change's: **the bug's
> evidence has since expired** (`hero_deployed` 0, `logo_deployed` 0, the decisive
> row gone). The loop searched properly and found nothing left to find. That is
> the harness working, not failing — and it reorders the commission, because
> item 1(a) now needs item 2's logging rather than another `090`. See the
> HANDOFF §3 and `bugs_open/236` §5b.

Two greppable literals this change adds, for the post-roll pod check:
`This listing is FILTERED, not the whole database` and
`you do not need a human to confirm it`. A negative control is available too —
the pre-change binary has neither, and `schema_always_tables` appears in no
earlier image.
