# NOTES — actioning the 2026-08-05 code-review triage

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

The triage that precedes this is `HANDOFF_2026-08-05_code_review_triage.md` (commit
`2bd82d8d6`). It sorted 15 findings by owning lane and checked each premise. This file
records what happened when they were actioned.

---

## 1. The handoff's central premise had expired within the hour

The handoff's operating rule was "contribute into the owning lane rather than compete", and
it deliberately filed nothing and fixed nothing on that basis. By the time this session
picked it up, **two of the three lanes had closed**:

```
bugs_closed/194_...four_of_six_save_page_sections_callers...md   (3016d9fbb, 08-05 11:01)
bugs_closed/195_...a_workflow_rejected_by_validation...md        (0173f2398, 08-05 11:01)
bugs_open/156_...content_identical_duplicate_slots...md          still open (d196fd74e)
```

Both lanes' `HANDOFF_2026-08-05_continue_here.md` were grepped for
`code.?review|triage|F1|F7|F15` — **zero hits in either**. So the findings were not picked
up by the lanes before they closed, and 10 of 15 became unowned rather than someone else's
in-flight work. That is what licensed fixing them here.

**The transferable bit:** the handoff was written at 11:02 and its ownership table was
already stale at 11:03. Ownership on this tree is not a property you record, it is a
property you re-read.

## 2. Ownership by FILE mis-assigned two findings

The handoff documents its own method (§"How I triaged", step 1): `git log -1 --format=... --
<file>` per file. That is file-granularity, and it is wrong when two lanes touch one file:

```
$ git log -1 -- platform/orchestration/actions/save_page_sections_action.go
84b7d561c 08-04 21:26 fix(156): collapse byte-identical duplicate sections...

$ git blame -L 624,628 -- platform/orchestration/actions/save_page_sections_action.go
47ee3ebce4 (cqls 2026-08-04 624) ... all five lines
```

F11 and F12 were triaged to the **156** lane (open, owned — so "contribute, do not compete")
when line-level blame puts them in **194** (closed, unowned — so mine to fix). The two lanes
committed to this file 26 minutes apart. **Blame the LINES the finding cites, not the file.**

## 3. F5 — measured, and the measurement redirected it, then refuted it

Filed as "unbounded `agent_error_log` writes … no rate limit, no dedupe and no `site_id`".
The write-side facts are all true. The predicted harm is not.

Measured (live DB, 2026-08-05):

| question | answer |
|---|---|
| rows from the new writer (`context->>'classification' = 'transient'`) | 160 in its first 71 min (09:10:50Z → 10:22:06Z) ≈ 135/hr |
| the table's historical peak, all writers | ~332 rows/hr |
| this writer's share of today | 166 of 2,445 ≈ **7%** |
| when the 20× explosion started | **08-03**, two days BEFORE this code shipped |
| `site_id` on those rows | 0 of 160 |
| `domain` on those rows | `''` on all 160 — **not NULL, and not a domain** |

Two corrections to my own working notes, both caught before they were written down as
findings:

- **`count(domain) = 160` does not mean the domain is populated.** `count()` counts
  non-NULL, and `LogAgentError`'s INSERT writes `domain` as a bare `$2` with no `NULLIF`,
  unlike `site_id`'s `NULLIF($1,'')::uuid`. All 160 carry the empty string. Fleet-wide that
  asymmetry means 3,189 rows read `''` where they mean "unknown" and only 122 read NULL, so
  `WHERE domain IS NULL` under-reports "no domain" by 26×. Recorded here; NOT fixed, because
  changing a shared writer's stored shape is a seam change and not what F5 asked for.
- **The eviction scenario is narrower than filed.** `diagnose_load_runtime_action.go:267` is
  indeed `ORDER BY occurred_at DESC LIMIT $3`, but the same query filters
  `($1::uuid IS NULL OR site_id = $1::uuid) AND ($2::text IS NULL OR domain = $2::text)`.
  With `site_id` NULL and `domain` `''`, these rows match **neither** a site-scoped nor a
  domain-scoped diagnosis. Only a fully unscoped load sees them.

And then the finding's core premise failed outright — see §4.

## 4. MISSTEP — I read a working reaper's output as proof there was no reaper

Full account in `WRONG_CALLS.md` (2026-08-05, code-review triage). In brief: 7,449 rows
spanning exactly 30 days, plus a `--include=*.go` grep for delete/reap/prune that returned
nothing, and I wrote "**No reaper** — a full month is retained."

There is a reaper. `scheduled_tasks.database-cleanup`, `enabled=t`, `interval_seconds=3600`,
`last_triggered_at = 2026-08-05 10:28:36Z` — minutes before I looked. Its `pre_query`:

```sql
DELETE FROM agent_error_log
WHERE (resolved = true  AND occurred_at < NOW() - INTERVAL '14 days')
   OR (resolved = false AND occurred_at < NOW() - INTERVAL '30 days')
```

```
 oldest_row                    | thirty_day_line               | consistent_with_a_working_reaper
 2026-07-06 10:51:54.004514+00 | 2026-07-06 10:30:51.605173+00 | t
```

21 minutes inside the line. The figure I read as "no retention" **was** the retention.

Two errors: the grep searched Go for a mechanism that lives in a SQL `pre_query` column; and
"oldest row is 30 days old" is produced identically by a working 30-day reaper and by no
reaper at all, so it could never have come out otherwise. Per the SessionStart landmine I did
at least read `enabled` + `last_triggered_at` rather than `last_completed_at`, which is
agent-written and would have proved nothing either way.

**Verdict on F5: the second FALSE POSITIVE of this review, after F4.** Both are absence
claims made without a lookup — the exact failure the handoff itself names for F4, reproduced
by me while checking it. No bug filed; the write-side facts are recorded here instead.

## 5. F7 — confirmed, and worse than the handoff had it

The handoff had the collision as "an operator copying the declaration between steps of one
agent". It is more concrete than that: **one live definition already holds both steps.**

```
 page-build-handler | save_sections    | save_page_sections    | (no require key)
 page-build-handler | validate_content | validate_page_content | require_sections_metadata=true
```

So `require_sections_metadata` means "warn me the stat audit could not run" on one step of
`page-build-handler` and would mean "refuse the save outright" on another. Live count of
`save_page_sections` steps carrying the key: **0** — RFC_010 shipped it seeded on nobody —
so the rename cost nothing and was taken.

**The collision had already misled a reader in our own tree.**
`save_sections_link_repair.go:13` stated the live page-build plan "sets both
`sections_metadata_field` and `require_sections_metadata: true`", as though one step's config
carried both. It does not — different keys, different steps, different actions, different
meanings. Corrected in place with a dated note rather than deleted, because the wrong comment
is evidence for the finding. Its conclusion survives: only the `sections_metadata_field` half
ever supported it.

## 6. F10 — the remedy is impossible, and the finding's premise is backwards

F10 asked for the local INSERT to be replaced by `orchestration.LogAgentError`, on the ground
that "the package's own precedent forbids the third copy".

```
platform/orchestration/coordinator.go:23:  ".../platform/orchestration/actions"
```

`orchestration` imports `actions`, so `actions` importing `orchestration` is a cycle. And
about **20 files** in `platform/orchestration/actions` each carry their own
`INSERT INTO agent_error_log`. The package's precedent *is* the copy, and it is structural.
Only the secondary half was actionable: `orchestration_id` was dropped, so a
`CONTENT_DATA_REGRESSION` row could not be joined to its run. Fixed.

## 7. F2 — the guard is now proven by the reviewer's own mutation

The two `agentbase` tests called only `messaging.*` functions and touched no `agentbase`
symbol, so neither could catch the drift its comment claimed to guard. Replaced with an AST
walk over `agent.go`. Mutation run, failing closed (abort unless the file is clean at HEAD,
abort unless the mutation is confirmed landed):

```
mutation landed: 1192: if match := messaging.MatchedValidationNeedle(err.Error()); match != "" {
--- FAIL: TestAgentbaseClassifiesThroughTheSharedPermanentClassifier
    agent.go no longer calls messaging.MatchedPermanentFailure anywhere
    agent.go classifies with the legacy messaging.MatchedValidationNeedle in [processMessage]
restored; git diff --numstat -> empty
```

**Why AST and not grep:** `server.go:113` names `MatchedPermanentFailure` in a *comment*. A
grep-based guard would be green on that comment alone while the real call site was reverted —
the "source-scanning test makes your COMMENTS load-bearing" landmine, avoided by parsing with
`parser.ParseFile(..., 0)`, which retains no comments.

## 8. Council — one verdict back, and its objection was right

`128d4fd1` (195 cluster) → **APPROVED**, 2 low-severity advisory objections. Both `editquality`
and `guardian` flagged the same thing: *"'Zero callers fleet-wide' is asserted from a
session-side grep … not a Go-tooling proof"*.

Fair, and answered with a check that could fail. Renaming both functions and rebuilding is a
compiler-verified zero-caller proof:

```
renamed symbols: 2   (IsRetryable -> IsRetryableZZPROOF, GetRetryAfter -> GetRetryAfterZZPROOF)
build-exit=0  vet-exit=0     # any caller anywhere would be an undefined symbol
restored; numstat empty
```

Scope of that proof, stated: it covers this module. An external importer of
`platform/errors` would not be seen — this is a single-module repo, so that is not a live
gap, but the claim is "nothing in this module calls them", not "nothing anywhere".

## 9. F13 — confirmed, and deliberately NOT fixed

The anchors are wrong exactly as filed: the comment cites `:215`/`:216`, but
`markTaskComplete` is at **:235** and the `config["task_name"]` read at **:237**. The
comment's own 21-line insertion shifted them — `216 + 21 = 237`.

Not fixed, because `check_endpoint_health_action.go` carries **25 insertions / 4 deletions of
another session's uncommitted work**, and the comment block holding those anchors is entirely
on the `+` side of their diff. Editing it would alter their in-flight work, and committing
that path would sweep it. Left for its author, with the corrected numbers above.

## 10. F6 — cannot be actioned; its content does not exist in the repo

F6 appears only in the handoff's ownership table (156 cluster,
`save_page_sections_action.go`). It is not described in any of the numbered verdict sections,
and the original `/code-review` output was never saved — the workstream directory contained
only the handoff. So there is no statement of what F6 claims. Recorded as an open gap rather
than guessed at.

**Transferable:** a triage that summarises N findings should either carry every finding's
text or say where the raw output lives. Ten of fifteen were fully recoverable from the
handoff's prose; F6 was not.

## 11. CORRECTION — my F3/F7 census used the query a LANDMINE says under-reports

> **CORRECTED 2026-08-05, same session.** §5 and the F3 work above both rest on a census of
> `save_page_sections` callers that I ran with a **top-level** walk:
> `LATERAL jsonb_each(ad.default_config #> '{workflow,steps}')`. It returns **3**. The true
> count is **6** — the step is nested in a loop `sub_workflow` for `pageflow-builder`,
> `page-rebuild` and `site-work-orchestrator`. `bugs_closed/194` says six in its own handoff,
> and `LANDMINES.md` carries the correct query under a footprint naming the exact config keys
> I was censusing.

Re-measured with the nested walk (`jsonb_path_query(default_config, '$.**.steps')`):

```
 page-build-handler      | save_sections | (no require key) | page_content.response.sections_metadata |      | validation_result.clean_html
 pageflow-builder        | save_sections | (no require key) | page_content.response.sections_metadata |      | assembled_page.html
 page-rebuild            | save_sections | (no require key) | page_content.response.sections_metadata |      | assembled_page.html
 page-rerender           | save_sections | (no require key) | rerender_sections.sections_metadata     |      |
 site-work-orchestrator  | save_sections | (no require key) | page_content.response.sections_metadata |      | assembled_page.html
 tool-recreation-handler | save_sections | (no require key) |                                        | true | validation_result.clean_html
```

**Both conclusions survive.** F7: `require_sections_metadata` is absent on all **six**, so the
rename really was free. F3: all **six** are explicit (five name `sections_metadata_field`, the
sixth declares `expects_no_sections_metadata`), so the implicit default is unreachable.

That is luck rather than method — had any of the three invisible callers carried the key,
"the rename is free" would have shipped as a false claim about live config. What is now
corrected in place: the code comment at the F3 warning site, this file, the handoff's
disposition table, `LANDMINES.md`, and three concept-register entries. The commit message of
`fa30062cc` and council submission `cb575682` both say "three" and **cannot** be corrected —
forward-only forbids an amend — so the correction lives here and in the code.

Full account in `WRONG_CALLS.md`. The instruction I failed to follow is already in memory:
grep `LANDMINES.md` for the symbol you are about to trust, BEFORE you trust it. The
SessionStart hook only matches landmines against files already dirty in the tree, and this one
guarded a file I had not yet touched when I ran the query.

## 12. LIVE on v1.0.1254 — pod-verified, both replicas, with a discriminating negative

The chassis rolled to `v1.0.1254` (pods `agent-chassis-d69d4467c-dvn8k` / `-fc8pq`, ~6 min old
when probed). Per `bugs_open/153` a roll is not evidence a fix shipped and the image carries no
provenance, so this was proven at the pod:

| literal | dvn8k | fc8pq | means |
|---|---|---|---|
| `refuse_save_without_sections_metadata` | 1 | 1 | F7's rename is in the binary |
| `incoming_sections_with_content_data` | 1 | 1 | `6e607da1e` — my LAST code commit — is in the binary |
| `require_save_without_sections_metadata` (wrong spelling) | 0 | 0 | the probe can return 0; it is not matching everything |

`incoming_sections_with_content_data` is the load-bearing one: it exists only after the final
code commit, so its presence dates the image past *all* of this lane's code.

**No removed-unique-literal negative was available, and that is worth stating rather than
faking.** The memory rule asks for a string the change REMOVED (expect 0) alongside one it
added. This change set cannot supply one: the edits were comments (absent from binaries), Go
identifier renames (`requireSectionsMetadataKey` → `refuseSaveWithoutMetadataKey` — identifiers
are not string literals), and one SQL fragment. The two candidates both fail:
`require_sections_metadata` survives in the binary from `validate_page_content_stats.go:235`,
and `build_status = 'deployed'` has **51** other occurrences in Go. So the negative here is a
wrong-spelling control, which proves the probe discriminates but NOT that the old code is gone.
The added-literal positives carry the actual claim.

Comment-only commit `1c6a3cab6` is unverifiable by this method by construction — comments are
not in the binary. Nothing to verify; noted so a reader does not go looking.

## 13. The owed check that F9 just changed the meaning of

Concept register `PBP-031` (`page-build-pipeline.md`) carries a verify-later: 24h post-roll,
`SELECT agent_type, count(*) FROM agent_error_log WHERE error_code='CONTENT_DATA_REGRESSION'
GROUP BY 1` — *"zero rows for page-build-handler and page-rerender is the pass; any
page-rerender row means the report's predicate is misconceived and the follow-up opt-in must
not proceed"*.

**F9 widened that predicate** (dropped `build_status = 'deployed'` from the count feeding it),
so the check now tests a different, wider condition than when it was written. Measured at the
roll:

```
CONTENT_DATA_REGRESSION, all history: 0 rows
save_page_sections rows, last 25 min: 0
any agent_error_log rows, last 25 min: 128 (newest 2026-08-05 20:52:32Z)  <- fleet IS live
```

**This is a PASS, and a weak one — say so.** Zero for both agent types is the stated pass
condition, but the code has been live for minutes and the report has **never fired in any
version**, so a zero here distinguishes nothing: it is what a correct widening and a broken one
both produce. The fleet is demonstrably running (128 rows), but no `save_page_sections`
traffic crossed the window, so even the absence is not yet informative about this path.

**Re-check due 2026-08-06 ~20:45Z** (24h from the roll). What each outcome means, decided now
rather than after seeing it: rows for `page-build-handler` only → the widening is catching the
non-deployed states F9 was about, which is the intended effect. **Any `page-rerender` row →
the register's own stop condition fires and the per-caller opt-in must not proceed** — and
because F9 widened the predicate, the first question is whether the row is a genuine
non-deployed loss or the widening over-firing. Still zero → the mechanism remains unexercised
and F9 is deployed-but-unproven, which is the state to record, not to round up to "verified".

## 14. 2026-08-06 morning — preparing the 20:45Z read, and a correction to how §13 measures it

Picked the lane up at 07:55Z, ~13h before the owed check. Four things changed or were wrong.

### 14a. The image rolled again — §12's proof was against pods that no longer exist

`v1.0.1254` → **`v1.0.1256`**, pods `agent-chassis-7d4d7b9669-2r8f2` / `-6f2ps`, restarted
**07:24:04Z / 07:24:28Z**. §12's evidence named `d69d4467c-dvn8k` / `-fc8pq`, which are gone.
A successor citing §12 would be citing a dead pod, so the probe was re-run [MEASURED 08-06 08:0xZ]:

| literal | 2r8f2 | 6f2ps | means |
|---|---|---|---|
| `incoming_sections_with_content_data` | 1 | 1 | the lane's LAST code commit is in the new image |
| `require_save_without_sections_metadata` (wrong spelling) | 0 | 0 | the probe still discriminates |

The gap §12 states honestly — no removed-unique-literal negative is available for this change
set — is unchanged and still applies. This re-verifies presence, not absence of the old code.

**The general point, which is not in the handoff:** a roll you did not perform silently
invalidates your pod-verification. `bugs_open/153` says a roll is not evidence your fix shipped;
the converse also holds — **a later roll is not evidence it still ships, and it retires your
pod names.** Re-probe rather than cite.

### 14b. CORRECTION — §13's pairing check cannot do the job it is given

> **CORRECTED 2026-08-06.** §13 and the handoff both propose
> `... WHERE action='save_page_sections' AND occurred_at > <roll>` "to confirm the path ran at
> all". **It cannot.** `agent_error_log` records *errors*. A `save_page_sections` that runs
> perfectly writes no row, so zero there means "no errors", never "no traffic" — and reading it
> as traffic would have turned a healthy fleet into apparent silence. Measured at 08:00Z: still
> 0 rows for that action, while the table took **3,096** rows fleet-wide since the roll.

The denominator that does work is the save's own side effect. `save_page_sections` DELETEs then
INSERTs the agent-writable rows, so `page_components.created_at` is fresh on every save
[MEASURED 08-06 08:0xZ, window = roll → now]:

```
rows_inserted 35 | distinct_pages 10 | rows_with_content_data 35
window 2026-08-05 20:52:27Z → 2026-08-06 02:21:09Z
```

**All 35 of 35 carry content_data.** That is the discriminating part: the report fires only when
an incoming save carries *none*, so it stayed quiet for a demonstrable reason rather than for
want of traffic.

**Its limit, stated rather than left to be found:** only the LAST save per page survives — 23
runs (below) collapsed to 10 pages, so ~13 intermediate saves left no `page_components` trace.
For those the detector is the only witness, which is circular if the detector is broken. §14d
is what breaks the circularity.

### 14c. The only caller that ran is `page-rerender` — the very one the stop condition names

Callers are separable in `workflow_plan` by their distinctive `sections_metadata_field` (the
§11 census): only `page-rerender` names `rerender_sections.sections_metadata`. Since the roll:

```
rerender_sections.sections_metadata | 23 orchestrations | all COMPLETED
window 2026-08-05 20:54:05Z → 2026-08-06 02:21:07Z
```

That window ends 2s before the last `page_components` insert, which is what ties the two
measurements to the same events.

**Captured now, deliberately, because it will not survive to 20:45Z.** `orchestration_states`
reaps terminal rows at ~24h (MEMORY, bugfix 003), so the 20:54Z runs age out around 20:54Z
today — *while the owed check is being read*. The read-out would have found the error-log zero
and no longer been able to establish which callers produced it. **When a check is scheduled for
T+24h and its denominator lives in a 24h-retention table, the denominator must be taken early.**

**MISSTEP — my first version of this query could not have come out otherwise.** I tested step
completion with `execution_metadata->'completed_steps' @> to_jsonb(step_name)` and got 0 across
all 23, which I nearly wrote down as "the step never ran". `completed_steps` is a **number**,
not an array of names, so the containment test was type-mismatched and returned false for every
row regardless of truth. The tell was that 0 contradicted the 35 inserted rows in the same
window. The counter is also simply dead: **0 on all 23 runs that are `status='COMPLETED'`** — so
it is not a usable execution signal for anyone else either. `WRONG_CALLS.md` gets this one.

### 14d. There IS a positive control, and it passes at HEAD

The worry a zero cannot answer is "can this detector fire at all". It can, at the predicate
level: `TestShouldReportContentDataLoss` case 1 ("the 194 signature") asserts `want: true`.
Run against **committed HEAD `61df92ff0`** via `git archive` — not the working tree, which holds
two other sessions' WIP *in this very package* (trap #7) — all four cases and both
`countExistingRowsWithContentData` tests **PASS**.

Scope, honestly: this proves the *decision function* fires. It does not exercise
predicate → INSERT → `agent_error_log` end to end. So a 20:45Z zero means "no regression
occurred on 23 page-rerender runs", not "the whole path is proven".

### 14e. 2b (the F3 warning) — settled, and by config rather than logs

The handoff asks for a pod-log grep. Both pods return **0** — but they restarted at 07:24Z, so
that grep covers ~35 minutes and is nearly worthless as evidence. **The invariant is config, not
logs, so measure the config** — nothing can age out of it [MEASURED 08-06 08:1xZ, nested walk]:
all **six** callers still explicit (five name `sections_metadata_field`, `tool-recreation-handler`
declares `expects_no_sections_metadata=true`); zero would-warn rows. The warning remains
structurally unreachable until someone adds a seventh caller. `RUNBOOK` R11 carries the query,
which reports the warn condition as a column so it cannot be missed by eyeballing six rows.

## 15. 2026-08-06 10:00Z — third roll of the morning, and my own §14c query was wrong twice

A fresh chassis build landed while this lane was open: **`v1.0.1257`**, pods
`agent-chassis-5b9fd84984-hqc5d` / `-qvzkg`, restarted **09:52:08Z / 09:52:29Z**. That is the
**third pod generation in under three hours** (1254 → 1256 → 1257). Re-probed both replicas
[MEASURED 08-06 10:0xZ]: added literal **1 / 1**, misspelled control **0 / 0**. The lane's
source files are untouched at HEAD since `1c6a3cab6`, so this is a rebuild of the same code.

**The operational point, now demonstrated three times over: on this cluster a pod-verification
has a half-life measured in hours.** Do not cite pod names in a durable doc without a date and
an image tag beside them, and re-probe rather than quote. `RUNBOOK` R12.

### 15a. Re-measured, and the traffic has grown

```
CONTENT_DATA_REGRESSION, all history .... 0 rows          (unchanged)
page_components since roll ............. 47 rows / 13 pages / 47 with content_data
                                          20:52:27Z → 2026-08-06 08:47:13Z
```

Was 35 / 10 / 35 at 08:00Z. **Still 47 of 47 carrying `content_data`** — the denominator grew
and the property held.

### 15b. CORRECTION — the caller query in §14c and `RUNBOOK` R10 was wrong in two ways

> **CORRECTED 2026-08-06 10:0xZ.** §14c identified callers by fingerprinting
> `sections_metadata_field` out of `workflow_plan`. That was unnecessary and partly wrong.
> **`orchestration_states` has an `owner_agent_type` column** — my first `\d` was truncated at
> 22 lines and I never looked past it, then built a fingerprint scheme to recover exactly what
> that column states outright.

Two defects, both of which the fingerprint hid:

1. **The fingerprint is not unique.** `page_content.response.sections_metadata` is carried by
   **four** definitions (page-build-handler, pageflow-builder, page-rebuild,
   site-work-orchestrator). Only `page-rerender`'s is distinctive — so the scheme happened to
   answer the one question I was asking and would have been ambiguous for any other.
2. **`count(*)` over the walk counts STEP OCCURRENCES, not runs.** The fingerprint query
   reported 2 for the second caller; `owner_agent_type` reports **1 orchestration**. Both are
   right about different things.

The exact figures, via `owner_agent_type` [MEASURED 08-06 10:0xZ]:

```
page-rerender | COMPLETED | 29 | 2026-08-05 20:54:05Z → 2026-08-06 08:48:25Z
page-rebuild  | COMPLETED |  1 |            2026-08-06 08:32:35Z
```

**A second caller has now run** — `page-rebuild`, once, at 08:32Z. So tonight's read is no
longer single-caller: page-rerender 29 runs, page-rebuild 1.

### 15c. The runtime plan expands loops — `workflow_plan` is NOT the definition

Chasing the 2-vs-1 discrepancy turned up something worth its own note. `page-rebuild`'s
`workflow_plan` holds **two distinct step names** with `action=save_page_sections`:

```
save_sections
build_pages_loop_iter_0_save_sections     <- runtime clone, materialised per loop iteration
```

Identical config on both. So the second is the loop `sub_workflow` **expanded at dispatch**,
not a second configuration site — which is why the `agent_definitions` census in §14e correctly
shows page-rebuild once, and **the F3 conclusion is unaffected**.

**The distinction to keep, because §11's misstep and this one are the same mistake from
opposite ends:** `agent_definitions.default_config` is the **config** census (what is
declared); `orchestration_states.workflow_plan` is the **traffic** census (what ran, with loops
already unrolled into `*_loop_iter_N_*` instances). Counting config sites in the runtime plan
over-counts; counting traffic in the definitions is impossible. §11 used the wrong walk over
the right table; §15b used the right walk over the table that answers a different question.
