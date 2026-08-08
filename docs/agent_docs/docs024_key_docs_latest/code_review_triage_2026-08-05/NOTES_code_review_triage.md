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
  > **CORRECTED 2026-08-06 11:2xZ — see §16.** Three things here are wrong or stale. The file
  > moved (`agenterrors/agenterrors.go:94`, RFC_012 option B, same day). The figures roughly
  > doubled (**9,949** `''` / **128** NULL, **79×**). And the framing is misleading: the
  > `NULLIF` on `site_id` is **compelled** by the `::uuid` cast (`''::uuid` errors), so this
  > INSERT's internal asymmetry evidences no intent at all. The real shape is **three
  > copy-paste clones breaking a 10-writer convention**, not one writer lacking one.
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

## 16. 2026-08-06 11:2xZ — the `domain` finding was real but MIS-FRAMED, and it is a clone family of three

Picked the lane up at 11:20Z. The 20:45Z check (§13/§2a) is ~9h off and `CONTENT_DATA_REGRESSION`
is **still 0 rows in all history** [MEASURED 11:2xZ] — nothing to do there but wait. F13 re-checked:
`git diff --numstat` still **25/4**, last commit still `bcb6afbe8`, so still another session's
uncommitted work and still not actionable.

What did move is the `domain` asymmetry the handoff parked in §3 as "real, fleet-wide, nobody's".
Three corrections to my own earlier account of it.

### 16a. CORRECTION — the file:line is dead, and the figures had roughly doubled

> **CORRECTED 2026-08-06 11:2xZ.** §5 above, `PLAN` §6 and the handoff §3 all locate this in
> `LogAgentError`'s INSERT in `platform/orchestration/agent_error_log.go`. **That INSERT is no
> longer there.** RFC_012 option B (owner ruling **2026-08-06**, i.e. the same day) moved the one
> INSERT into the leaf package: it is now
> `platform/orchestration/agenterrors/agenterrors.go:89–102`, and `LogAgentError` is a thin
> forwarder onto `agenterrors.Write`. **The asymmetry travelled with it unchanged** — a refactor
> that picked this exact query up and put it down elsewhere did not surface it, because moving a
> query preserves it verbatim. Worth stating as its own lesson: *a refactor is not a review of
> what it moves.*

Re-measured [MEASURED 08-06 11:2xZ, `agent_error_log`, ~~all history~~ **the ≤30-day retained
window — see §17c, this table is reaped and §4 already knew it**]:

```
domain IS NULL ....   128        (was 122)
domain = ''    .... 9,949        (was 3,189)
domain <> ''   .... 4,688        (was 4,155)
total          .. 14,765        (was ~7,466)
```

So `WHERE domain IS NULL` now sees **128 of 10,077** rows that have no domain — **1.3%**, an
under-report of **79×**, not the 26× recorded yesterday. The figure is not stable and should not
be quoted without its date; the ratio drifts because the writers that produce `''` are the
high-volume ones (16c).

### 16b. CORRECTION — `NULLIF` on the sibling columns is COMPELLED, so that INSERT shows no intent

My earlier framing was "`site_id` gets `NULLIF`, `domain` does not" — read as an inconsistency
someone overlooked. That reading is wrong about the cause. `site_id` and `work_item_id` are
**uuid**; `domain` is **text** (`\d agent_error_log`). And [MEASURED — this could have come out
otherwise, so it is evidence]:

```
SELECT ''::uuid;  -- ERROR:  invalid input syntax for type uuid: ""
SELECT ''::text;  -- legal, returns the empty string
```

`NULLIF($1,'')::uuid` is therefore **mandatory** for that INSERT to execute at all on an unset
site_id. It was never a decision about NULL semantics. **So within that single writer there is no
evidence of intent about `domain` either way** — and an argument from its internal asymmetry is an
argument from a cast requirement. The intent is only visible across the fleet, which is 16c.

**MISSTEP, recorded because I nearly wrote it down.** From the cast finding I inferred "there is no
convention for text columns, so bare `$2` is normal". **Refuted in one grep:** `NULLIF($n,'')` on
plain text columns appears **32 times across 24 files**. The convention exists; I had reasoned from
one site to a fleet-wide claim without counting. `WRONG_CALLS.md` gets this.

### 16c. What it actually is: 20 writers, one convention, and three clones that break it

`grep -rn "INSERT INTO agent_error_log" --include=*.go` finds **20 non-test writers** — worth
noting on its own, because `bugs_closed/034` closed partly on the claim of "**One**
`agent_error_log` writer". That consolidation held for the two sites 034 was about and has since
been overtaken by nineteen more. Read each one's `VALUES` clause (`RUNBOOK` R13):

| how `domain` is written | writers | stored when unset |
|---|---|---|
| `NULLIF($n, '')` | **10** | NULL ✅ |
| column omitted entirely | **7** | NULL ✅ (column default) |
| **bare `$2`** | **3** | **`''`** ❌ |

The three outliers are **one copy-paste family**, with a byte-identical 13-column block:

```
platform/orchestration/agenterrors/agenterrors.go:94
platform/orchestration/actions/store_generated_component_action.go:1358
platform/orchestration/actions/component_write_guard.go:320
    NULLIF($1, '')::uuid, $2, NULLIF($3, '')::uuid, $4,
```

**So this is clone-and-drift against an existing majority convention, not a missing convention.**
That matters for the fix: it is three characters at three sites converging onto what 17 of 20
writers already do — not a novel contract.

### 16d. The causal claim, confirmed by a partition that could have come out otherwise

Writer shape predicts stored shape, with **zero overlap**, across all 14,765 rows **then retained**
(§17c: a ≤30-day window, not all time) [MEASURED 08-06 11:2xZ]. Grouping by `error_code`:

- Every code carrying `domain=''` is a **classifier-generated** code — `LLM_API_ERROR` 3,481,
  `PROCESSING_FAILED` 3,399, `TIMEOUT` 1,499, `UNKNOWN` 1,458, `VALIDATION_ERROR_DROPPED` 89,
  `PARSE_ERROR` 21, `CHILD_ORCHESTRATION_FAILED` 13, `INCOMING_MESSAGE_REJECTED` 5,
  `UNROUTED_IMAGE_KIND` 1 — i.e. exactly the codes `agenterrors.ClassifyError` and the coordinator
  path emit, which is the bare-`$2` generic writer. **`is_null` is 0 for every one of them.**
- Every one of the 128 NULL rows carries a code belonging to a `NULLIF` writer or a
  column-omitting one: `TRUNCATION_DEGRADED_REVIEW` 41, `CONTENT_LINK_REPAIR_DETAIL` 30,
  `REVIEW_SUPERSEDED_BY_PASSING_SAVE` 25, `CONTENT_CLAIMS_FLOOR_DETAIL` 17,
  `CONTENT_DATA_LINK_AUDIT` 10, `FIX_PLAN_VALIDATION_REFUSED` 3,
  `CONTENT_DUPLICATE_SECTIONS_COLLAPSED` 1, `CONTENT_DATA_ENVELOPE` 1.

**The two sets do not intersect.** A per-code mix would have refuted the mapping; there is none.
This is why the damage is lopsided — the one generic writer used by the whole coordinator path is
in the broken family, so it produces the bulk of the table.

### 16e. Why it still bites, and what is NOT claimed

The consumer that loses most is the one §5 already identified:
`diagnose_load_runtime_action.go:267` filters
`($1::uuid IS NULL OR site_id = $1::uuid) AND ($2::text IS NULL OR domain = $2::text)`. A row with
`site_id` NULL and `domain` `''` matches **neither** a site-scoped nor a domain-scoped diagnosis —
only a fully unscoped load. That was 160 rows when §5 measured it; the `''` population is now
**9,949**.

**NOT claimed, and do not let this read as more than it is:** I have not audited every consumer, so
"~79× under-report" is a property of the STORED DATA, not a measured harm to a named report. I have
not run `090` on it. And the fix is still a stored-shape change on the fleet's highest-volume error
writer — cheap to type, and therefore exactly the kind of change whose blast radius wants counting
rather than assuming. **`WHERE domain IS NULL` is the wrong predicate for "no domain" on this table
today; use `COALESCE(domain,'') = ''` until the writers converge.**

## 17. 2026-08-07 01:1xZ — THE OWED 20:45Z READ, done 4.5h late. Verdict: STILL ZERO (branch 3)

Read at **2026-08-07 01:1xZ**, i.e. **4h31m after** the 2026-08-06 20:45Z due time. The lateness
costs one thing and only one: `orchestration_states` had reaped further, so the caller counts had
to be reassembled from the durable record rather than read off in one query (§17b). Everything
else was still measurable.

**Live at the read:** `v1.0.1261`, pods `agent-chassis-c9c6d45cf-7c8wq` / `-nmscp`, both started
**2026-08-06 19:54Z** — so the binary serving the due-time window is the one still running, and
this is a *fifth* pod generation (1254 → 1256 → 1257 → … → 1261). Pod-probed both replicas
[MEASURED 01:1xZ]: `incoming_sections_with_content_data` **1 / 1**, misspelled control
`require_save_without_sections_metadata` **0 / 0**. §12's gap is unchanged — presence, not absence
of the old code.

### 17a. The read itself

```
CONTENT_DATA_REGRESSION, grouped by agent_type ....... 0 rows
```

**This is branch 3 of the three decided in §13: "still zero".** Not branch 1 (no
`page-build-handler` rows) and — the one that mattered — **not branch 2: no `page-rerender` row,
so `PBP-031`'s stop condition has NOT fired and the per-caller opt-in is not blocked by this
read.**

The denominator, which is what makes the zero mean anything:

```
page_components since the roll .. 55 rows | 16 pages | 55 of 55 with content_data
                                  2026-08-05 20:52:27Z → 2026-08-06 20:15:36Z
of those, on v1.0.1261 only ..... 11 rows |  3 pages | 11 of 11 with content_data
                                  (pod start 2026-08-06 19:54:23Z)
```

**Callers, reassembled (§17b explains why it cannot be one query): 48 runs across THREE callers,
all COMPLETED.**

```
page-rerender ....... 44   = 29 (durable record, 20:54:05Z→08:48:25Z) + 15 (11:36:53Z→20:15:33Z)
page-build-handler ...  3   11:51:21Z → 20:39:42Z      <- NEW: a third caller, and see below
page-rebuild .........  1   08:32:35Z
```

**`page-build-handler` has now run, three times, and produced no row.** That is worth saying
plainly because it is the caller **F9's widening was about** — branch 1's subject. So the widened
predicate has now been exercised by its own target caller and did **not** over-fire. That is a
genuinely stronger result than yesterday's single-caller position, and it is the one thing this
read adds beyond "still zero".

**The honest sentence, per §13's instruction not to round up:** *48 runs across three callers and
55 saves over 16 pages produced no regression row, on a predicate proven able to fire at the unit
level (`TestShouldReportContentDataLoss` case 1) — and every one of those 55 saves carried
`content_data`, so the report stayed quiet for a demonstrable reason rather than for want of
traffic. It is NOT proven end to end: nothing here exercises predicate → INSERT →
`agent_error_log`.* Known limit, unchanged from §14b: only the last save per page survives in
`page_components` (55 rows, 16 pages), so intermediate saves left no trace and the detector is
their only witness.

### 17b. The ~24h reaper premise is CONFIRMED — and my first check of it was blind in the exact way `RUNBOOK` R6 warns about

§14c took the denominator early on the premise that `orchestration_states` reaps terminal rows at
~24h. **That premise is correct**, and is now characterised rather than believed. But the first
query I ran to check it said the opposite, and I nearly recorded that.

```
min(created_at) over ALL of orchestration_states  ->  2026-07-13, i.e. 24 DAYS
```

Which reads as "there is no 24h reaper" and would have retro-justified not taking the denominator
early. **It is the R6 trap verbatim** — a figure produced identically by a working reaper and by
none at all — because the reaper's status set does not cover every row. Splitting by status
discriminates it in one query [MEASURED 01:1xZ]:

```
COMPLETED   2056 rows | oldest 2026-08-06 00:46:30Z | 1d 00:33   <- reaped
FAILED      1678 rows | oldest 2026-08-06 00:37:00Z | 1d 00:42   <- reaped
CANCELLED     24 rows | oldest 2026-07-19           | 18d        <- NOT reaped
RUNNING       15 rows | oldest 2026-07-29           |  8d        <- NOT reaped
INITIALIZED    1 row  | oldest 2026-07-13           | 24d        <- the row that poisoned min()
```

Then read the reaper instead of inferring it — it is SQL in a `scheduled_tasks` column, invisible
to a Go grep (`RUNBOOK` R6 again). Task **`database-cleanup`**, enabled, hourly, last triggered
**2026-08-07 00:46:08Z**:

```sql
DELETE FROM orchestration_states
WHERE status IN ('COMPLETED', 'FAILED') AND updated_at < NOW() - INTERVAL '24 hours'
```

**The boundary matches the reaper's own last run minus 24h to within 30 seconds** (ran 00:46:08Z;
oldest surviving COMPLETED is 00:46:30Z the previous day). That is as clean a confirmation as this
kind of claim gets. Three refinements worth carrying:

1. It is **`updated_at`**, not `created_at` — a long-running orchestration survives longer than its
   creation time suggests.
2. **Only `COMPLETED` and `FAILED`.** `CANCELLED` is *not* in the set, which is why 24 cancelled
   rows go back to 07-19. Whether that is deliberate or an oversight is **[UNMEASURED]** — I have
   not found an argument either way and am not filing one. (Clause 4 separately reaps
   `EXECUTING_STEP`/`AWAITING_RESPONSES` stale >4h, which is why the 07-29 `RUNNING` rows also
   persist: `RUNNING` is in neither set.)
3. So **§14c's decision to take the denominator early was right**, and for a better reason than it
   gave: not "the table reaps at 24h" but "the table reaps *these two statuses* at 24h on
   `updated_at`, and every row this lane cares about is `COMPLETED`."

### 17c. CORRECTION to §16 — I wrote "all history" three times and this table does not have all history

> **CORRECTED 2026-08-07 01:1xZ.** §16 and the `LANDMINES` entry both describe the `domain` census
> as `agent_error_log`, **all history**. It is not. The same `database-cleanup` task's *first*
> clause reaps this table:
>
> ```sql
> DELETE FROM agent_error_log
> WHERE (resolved = true  AND occurred_at < NOW() - INTERVAL '14 days')
>    OR (resolved = false AND occurred_at < NOW() - INTERVAL '30 days')
> ```
>
> So the 14,765-row population is a **≤30-day window** (≤14 days for anything resolved), not all
> history. **And the aggravating detail: §4 OF THIS FILE already had this.** Yesterday I
> discovered that same reaper the hard way — §4 is titled "MISSTEP — I read a working reaper's
> output as proof there was no reaper" and quotes the 14/30-day `DELETE` verbatim. Then today I
> wrote "all history" about the same table, three times, twelve sections further down. **The
> check is: grep your own NOTES for the table before characterising its population**
> (`grep -n "agent_error_log" NOTES_*.md` would have surfaced §4 instantly). A misstep recorded
> in the right file still has to be *retrieved* to be worth anything, and a long append-only log
> is exactly where a fact goes to be forgotten by its own author.
> **What survives the correction:** the 79× under-report and the exact writer↔shape
> partition are properties of *the population a reader can actually query*, which is the thing the
> operational advice is about — so `COALESCE(domain,'') = ''` stands unchanged. **What does not:**
> any claim about totals since the beginning, and the implication that the 26×→79× move was pure
> growth (some of it is the mix ageing out). The ratio is a moving figure over a rolling window and
> must never be quoted without its date.

**The same caveat lands on this section's own headline, so state it rather than inherit it.**
"`CONTENT_DATA_REGRESSION` = 0 rows in all history" is really "0 rows in the ≤30-day retained
window". For *this* read that is sound — the roll was 2026-08-05, well inside the window, so the
post-roll question is answered. But **§13's stronger claim that "the report has never fired in any
version" is not establishable from this table** beyond 30 days, and no version of it existed 30
days ago anyway. Net effect on the verdict: none. Net effect on how the verdict may be quoted:
say "no regression in the retained window since the roll", never "never fired".

## 18. 2026-08-07 01:3xZ — the `090` verdict: UNVERIFIABLE, and the verdict itself was nearly lost

Filed the `domain` divergence at 01:16Z (§16, intake `94144fbc`, run correlation
**`a7b1e113-8857-4161-ad2b-f3b7387e33e9`** — the run one is the artifact key). It completed at
01:30Z. Three separate things came out of it and only the first is about `domain`.

### 18a. Verdict: **UNVERIFIABLE** — and the mechanism was NOT refuted

Not CONFIRMED, not REFUTED. The loop's own words on the mechanism:

> "The mechanism (bare `$2` for domain) is real, but at a different symbol than named."

Its three blocking gaps, all traceable to **one** cause — **it cannot see code written after
2026-07-28**:

1. **"the code index has zero hits for `package agenterrors` and zero indexed paths under
   `platform/orchestration/agenterrors/`"**, so it read
   `platform/orchestration/agent_error_log.go:LogAgentError` and found it does **its own**
   bare-parameter INSERT with no forwarding — and correctly reported that as inconsistent with my
   claim that the INSERT lives in `agenterrors.Write` and `LogAgentError` merely forwards.
   **Both descriptions are accurate about different trees.** Mine is HEAD; the loop's is the
   pre-RFC_012 file. RFC_012 landed 2026-08-06; the index is frozen at `d98010e8b` (07-28).
2. Its symbol searches for the NULLIF-wrapping siblings (`save_sections_dedup.go`,
   `discovery_checks.go`, `content_data_envelope_guard.go`, `save_sections_claims_guard.go`)
   returned **zero** — same cause, these are post-07-28 files too.
3. It could not attribute the dominant `domain=''` cluster to a specific writer, because those
   bodies were not in scope.

**The loop was honest about its own blindness**, which is worth recording as the mechanism working:
one of its `code_requests` reads *"Re-check after the index catches up whether the `agenterrors`
package the hypothesis's location claim depends on exists at all, given the prior 0-row answer was
against a 9-day-stale index."* That is `bugs_open/108`'s fix behaving exactly as MEMORY says it
does — **it reports stale, it does not claim fresh.**

**So the honest reading: UNVERIFIABLE here means "the loop could not read the tree", not "the claim
is doubtful".** It independently re-derived the bare-`$2` mechanism from the code it *could* see,
and its own state query reproduced the shape. **What it could not do is check my location claim,
and that is the one part a stale index guarantees it cannot check.** Re-running is pointless until
the index moves — and the fix for that is the known LANDMINE (migration 252 pins the index ref to
`086_experience_loop`; it wants `'main'`). **Do NOT resubmit this to `090` before then**: it will
return UNVERIFIABLE again for the same reason and cost another run.

### 18b. THE VERDICT WAS COMPUTED AND THEN LOST — and every status said success

This is the more serious finding and it is not about `domain` at all.

```
site_work_items.status ......................... complete
orchestration_states (all three) ............... COMPLETED
diagnosis_artifacts for the correlation ........ 5 rows, EVERY ONE kind='bundle'
                                                 0 report, 0 verdict, 0 diagnosis
```

A reader doing the obvious thing — check the status, then read the report — finds unambiguous
success and **no report**, with nothing anywhere saying why. The verdict existed the whole time, in
`orchestration_states.collected_data->'verdict'` of the `diagnose-agent` run, and
**`collected_data` is on the 24-hour COMPLETED reaper (§17b), so it had ~24h to live.** I recovered
it by enumerating `jsonb_object_keys(collected_data)` and finding `verdict`, `diagnosis` and
`diagnosis_note` keys — not by any documented route.

What actually happened, from the pods the rows name themselves (`agent_error_log.pod_name` —
`agent-diagnose-agent-1db78640-zdgbs`, `agent-diagnose-orchestrator-19e0254b-cgm7t`; **these are
dedicated per-agent pods, NOT the chassis**):

```
workflow completed but its result could not be delivered to the parent
  (failed_transient): message validation failed          × 2, step_name='complete', 01:30:51/53Z
```

`coordinator.go:3754` is doing the **right** thing here — its own comment says so: *"A workflow
whose result never reached its parent did not succeed from the parent's point of view"*, so it
calls `notifyParentOfFailure`. The message body carries
`error.code = CHILD_ORCHESTRATION_FAILED`. The defect is downstream of that correct decision: the
**work item was still marked `complete`**, and no artifact recorded the loss.

**Two traps for anyone reading this table, both first-hand:**
- **`agent_error_log.error_code` for these rows is `UNKNOWN`** (the classifier's fallback). The
  informative code, `CHILD_ORCHESTRATION_FAILED`, exists only in the Kafka message body. So you
  cannot find this class by filtering `error_code`; filter `error_message LIKE '%could not be
  delivered to the parent%'`.
- **Do not grep the chassis for it.** I did, got a clean zero on both replicas, and the zero was
  the wrong-service trap: a positive control showed 219 log lines in the window and **0** mentions
  of `diagnose`. The rows' own `pod_name` column is the answer.

**Root cause NOT diagnosed and NOT asserted.** `reply_to_request_id` is empty (`""`) in the failure
message I read, and `DeliverReply` is passed `[]byte(replyToRequestID)` as the key — so "the
success reply carried an empty key and validation rejected it" is a **[HYPOTHESIS, UNMEASURED]**,
not a finding: the message I inspected is the *failure* notification, a different message from the
one that was rejected. Anyone taking this further should capture the rejected reply, not infer from
its replacement.

### 18c. The loop surfaced a fact I did not have, and it CORRECTS §16's volume claim

Its state citation disagreed with my §16 figures, so I re-measured rather than take it on trust
[MEASURED 01:3xZ]. It is corroborated — 13,783 `''` / 4,742 real / **128 NULL** / 18,653 total,
against the loop's 13,765 / 4,742 / 128 four minutes earlier. My own 08-06 11:2xZ figure was
**9,949** `''`. So the empty-string population grew by ~3,800 in 14 hours, and:

```
vet-practice-verifier | scrape_website | 6752   2026-08-04 20:02Z → 2026-08-07 01:32Z
vet-practice-verifier | (no step)      | 2726
vet-intel             | (no step)      | 2612
generic               | call_dispatch  |  439   (07-08 → 07-26, the old steady-state)
```

> **CORRECTED 2026-08-07 — §16 attributed the VOLUME to the wrong thing.** §16 says the damage is
> lopsided because "the one generic writer used by the whole coordinator path … produces the bulk
> of the table". The *mechanism* half stands — those rows do come through the bare-`$2` writer. The
> *volume* half is wrong: **12,090 of 13,783 (88%) are one live incident in the vet lane**, dated
> from 2026-08-04 and **still firing while I write this** (max 01:32:33Z). This is not steady-state
> generic traffic; it is a burst. Consequences: (a) the under-report ratio is now **~109×**
> (13,911 no-domain, 128 NULL) and moving fast — it was 26× on 08-05 and 79× on 08-06, so **the
> ratio is an incident metric, not a property of the defect**, and any figure quoted without its
> timestamp is worthless; (b) the `domain` defect's *severity* should be argued from the mechanism
> and the reader-side blindness, **never from the row count**, which will fall when the vet lane
> stops failing.

**Separately, and not mine:** that burst is `PROCESSING_FAILED` 5,225 / `LLM_API_ERROR` 4,684 /
`TIMEOUT` 2,060, and the loop's own runtime citation caught it mid-failure — *"scrape_website
(scrape_web) fatal: … An SSL/TLS certificate error occurred"*. **~12,000 error rows in 2.5 days
from one lane, still climbing, is worth an owner's attention on its own account.** Candidate
neighbours, asserted as candidates only: `bugs_open/205` (2026-08-06, "two records retry forever")
and `bugs_open/183`. I have not investigated and am filing nothing.

## 19. 2026-08-07 08:3xZ — the code index is CURRENT, and §18a's blocker is cleared

Migration **332** repointed `scheduled_tasks.code-index-refresh` from the dead
`086_experience_loop` to the live `087_towards_multiple_domains`; nudged the schedule; the run
completed. Full procedure and its six gotchas: `RUNBOOK` **R14**.

```
BEFORE  086_experience_loop          4,992 symbols  commit_time 2026-07-28 10:31:33Z
AFTER   087_towards_multiple_domains 5,754 symbols  commit_time 2026-08-07 01:53:30Z
        668 paths, ONE commit_sha, and 086 leftovers = 0 (clean cutover)
```

`commit_time` matches the remote pushed tip `2c3041f7d` exactly — the indexer fetches the pushed
tip, so my own later local commits are legitimately absent.

**Verified with a discriminating pair, not just a count.** The positive control is the very thing
the `090` run reported "zero hits" for, and it cannot exist in a corpus frozen at 07-28:

```
agenterrors ....................  5   <- POSITIVE: created 2026-08-06 (RFC_012)
save_sections_dedup ............  7   \
discovery_checks ............... 490   |  the four sibling writers whose symbol
content_data_envelope_guard ....  7   |  searches returned zero in §18a
save_sections_claims_guard .....  5   /
agenterrorz ....................  0   <- NEGATIVE: the LIKE discriminates
```

**So all three of §18a's blocking gaps are now resolvable, and the "DO NOT RESUBMIT" advice there
is spent** — it was conditional on the stale corpus and that condition is gone. A fresh `090` on
the `domain` divergence would now be able to check the location claim it could not check before.
That is a live option, not a done thing: it costs a run, and **the `090` delivery defect in §18b is
NOT fixed**, so anyone resubmitting must still recover the verdict from
`collected_data->'verdict'` within 24h rather than expect a report artifact.

**One thing this does NOT fix, and it is the reason R14 exists rather than a one-line note:** the
pin is a constant, so it goes stale again silently the day work moves to `088_*` — daily job still
green, `updated_at` fresh, `commit_time` ageing. 252 rejected the self-deriving alternative for
reasons that still hold. **Repointing it is part of cutting a branch**, now recorded in
`LANDMINES.md`, `RUNBOOK` R14 and the memory index.

## 20. 2026-08-08 15:1xZ — the resubmitted 090, a refactor that superseded §16, and the STOP CONDITION FIRED

Resubmitted the `domain` divergence on the now-current index (§19). Run correlation
**`61153a74-a285-4815-a533-bd89ed8b6d07`**. Three outcomes, and the third is the consequential one.

### 20a. Verdict UNVERIFIABLE again — but it earned its run by finding a refactor I had missed

`diagnosis_artifacts`: **4 rows, all `kind='bundle'`, zero reports** — §18b's delivery defect
reproduced exactly. This time the watcher captured `collected_data->'verdict'` to disk on terminal,
by design, so nothing was at risk. **And the previous run's rows were gone when I checked (0 rows
for `a7b1e113`), which confirms the 24h reaper claim was real** — recovering that verdict when I did
was the only chance anyone had to read it.

The verdict says the code **confirms the structural claim**, then asks for runtime evidence. What
makes the run worth its cost is a citation I had never seen:

```
platform/orchestration/actions/log_action_error.go:LogActionEntry
  base := actionErrorEntry(params, entry.SiteID, entry.Domain)
  return agenterrors.Write(ctx, params.DB, logger, entry)
```

> **CORRECTION TO §16 — my writer census is SUPERSEDED, and by a commit dated the day AFTER it.**
> `f930de86b` (2026-08-07 02:24) — *"RFC_012 B: the copy class is retired — 18 hand-copied
> `agent_error_log` INSERTs become one writer, and the 19th arrived DURING the work"*. Re-measured
> [MEASURED 08-08 15:1xZ]: **the 20 INSERT sites are now 3** —
> `agenterrors/agenterrors.go:89` (the one writer), `store_generated_component_action.go:1353`,
> `internal/agents/contentcreator/claims_guard.go:184`. Several files §16 listed as **`NULLIF`
> writers** now build an `agenterrors.Entry` and call `LogActionEntry`, which **never touches
> `entry.Domain`** and forwards to the bare-`$2` INSERT.
>
> **So §16's shape is inverted, and both halves matter.** The bad news: this is no longer "3 clones
> breaking a 10-writer convention" — the convention was **consolidated onto the broken shape**, so
> the defect is now fleet-wide by construction. The good news, and it is bigger: **the fix collapses
> from three sites to one line.** `NULLIF($2,'')` at `agenterrors.go:94` fixes every consolidated
> caller at once; only `store_generated_component_action.go:1353` needs its own (and
> `contentcreator/claims_guard.go` omits the column, so it already writes NULL).
>
> **The lesson is not "I was wrong", it is that a structural census has a shelf life of about a
> day on this tree.** §16 was measured correctly and became false while sitting in a committed
> document, because a refactor landed under it. A census is evidence about a moment; date it, and
> re-run it before acting on it.

**And the runtime evidence the loop asked for now exists — but not where it looked.** Its
`data_request` queried three guard error codes; all their rows are **07-31 → 08-05, i.e.
PRE-consolidation** [MEASURED], so that query could not discriminate. The discriminating cut is
post-consolidation rows by domain shape:

```
rows since 2026-08-07 12:00Z, by error_code:  is_null = 0  ON EVERY ONE OF 8 CODES
  (UNKNOWN 206 ''/237 real · VALIDATION_ERROR_DROPPED 2 '' · CONTENT_LINK_REPAIR_DETAIL 2 real · …)
```

Before the consolidation those same guard codes produced **NULL** (§16d: `CONTENT_LINK_REPAIR_DETAIL`
30 NULL, `CONTENT_CLAIMS_FLOOR_DETAIL` 17 NULL, …). **The NULL-producing shape is gone from the
fleet** — a row now lands `''` or a real domain, never NULL. That is the mechanism confirmed at
runtime, and it is what the loop wanted; it simply asked the codes whose traffic predates the change.

### 20b. `CONTENT_DATA_REGRESSION` HAS FIRED, and it is a `page-rerender` row

Found incidentally while measuring the above. **This supersedes §17a's "still zero" verdict.**

```
agent_type   page-rerender          <- PBP-031's stop condition names exactly this
occurred_at  2026-08-08 15:14:35.326571Z
site         vetcomparison.uk   page  tool-cma-obligation-checker   build_status  deployed
```

`PBP-031` verbatim: *"any `page-rerender` row means the report's predicate is misconceived and the
follow-up opt-in must not proceed."* **So the stop condition is FIRED and the per-caller opt-in of
`refuse_save_without_sections_metadata` MUST NOT PROCEED.** §17a's verdict stood for ~38 hours.

§13 said settle one question first — genuine loss, or the F9 widening over-firing? **Neither
branch as written, and the real answer vindicates the register's wording.**

- **Not the widening.** The page is `build_status='deployed'`, so the pre-F9 predicate would have
  caught it too. [MEASURED]
- **The save proceeded and did null the column.** The new `page_components` row was created at
  `15:14:35.357283Z` — **31ms after the warning** — with `content_data` NULL and 13,967 chars of
  `rendered_html` intact. [MEASURED]
- **But there was no substantive content to lose.** The pre-save snapshot
  (`page_component_history`, `source='save_page_sections_overwrite'`, `15:14:35.336054Z`) holds
  `content_data` of type `object` and **2 characters** — i.e. `{}`. [MEASURED]
- **And the predicate's precondition is `content_data IS NOT NULL`**
  (`save_sections_metadata_source.go:226`, read at source), **which `{}` satisfies.** So the
  warning's "had 1 component(s) holding structured content_data" is true only in the IS-NOT-NULL
  sense.

**This is the same defect family as the `domain` finding, in the report that was tracking it: a
non-NULL-but-EMPTY value read as present.** `content_data IS NOT NULL` counts `{}` exactly as
`count(domain)` counts `''`. The register guessed "the predicate is misconceived" and was right,
for a reason nobody had identified.

**NOT claimed:** whether `{}` → NULL matters behaviourally to the rerender path is **[UNMEASURED]**
— PBP-031 says a NULL escalates the whole page to a full LLM rebuild, so the transition may have a
real cost even though no *content* was lost. That is the next question, not a settled one. **One
row is not a rate**, either: this is a single occurrence on one tool page, not a measured frequency.

## 21. 2026-08-08 16:0xZ — the `domain` fix is AT THE COUNCIL GATE

Submitted the one-line fix. **`SUBMISSION_CORR = 5d6501ba-db0e-4963-a1ca-e4736d41210e`**
(submission file kept in this lane: `SUBMISSION_2026-08-08_domain_nullif.json`).
Fresh chassis confirmed live first: **`v1.0.1266`**, pods `agent-chassis-856dff6b46-f86mr` /
`-smr28`, started 2026-08-08 16:00:40/16:01:01Z.

**The change:** `NULLIF($2,'')` on the `domain` parameter at two sites —
`agenterrors/agenterrors.go:94` (the one writer since RFC_012 B, so it covers every consolidated
caller) and `store_generated_component_action.go:1358` (the only unconsolidated straggler that
names the column). The third surviving INSERT,
`internal/agents/contentcreator/claims_guard.go:184`, omits `domain` and already stores NULL —
deliberately not edited. Plus a source-scan regression test, **submitted with its own limitation
stated**: `NULLIF` is server-side, so `sqlmock` still sees the bind arg as `''` and **a unit test
cannot prove this behaviour** — it guards the SQL text only, and the behavioural proof is the
post-deploy artefact check.

**Blast radius MEASURED before submitting, not delegated to the reviewer** (the 2026-07-28 ruling):

- **Exactly one Go reader filters on `domain`** — `diagnose_load_runtime_action.go:269`,
  `AND ($2::text IS NULL OR domain = $2::text)`. **Its behaviour does not change**, and the reason
  is the *binding*, not the predicate: `nullText("")` returns nil (`:542`), so an unscoped
  diagnosis binds NULL and the clause short-circuits; a scoped one binds a real domain, which
  rows carrying `''` never matched and rows carrying NULL will not match either.
- **Zero** hits sweeping `*.sql`/`*.sh`/`*.py`/`*.ts`/`*.tsx` for `domain IS NULL`,
  `count(domain)` or `domain = ''` against this table.
- So **no consumer's result set changes.** The only change is what is stored going forward.

**Named, not smuggled:** no backfill — the 13,783 existing `''` rows keep their value, so the
table holds a **mixed shape indefinitely** and `COALESCE(domain,'') = ''` stays the safe predicate
for historical queries. And the case is argued **from the mechanism and the reader-side trap, never
from the row count**, because 88% of that count is one live vet-lane incident (§18c) and will fall
when it stops.

**Deliberately NOT done in this session:** the code edit is not written and nothing is committed.
The submission describes it; a successor writes it. If you commit before the verdict lands, use
`Council-Submitted: 5d6501ba-db0e-4963-a1ca-e4736d41210e` — **never `Council-Reviewed:` on a
verdict you have not read** (098 buckets that as MISMATCH).

Read the verdict **by this correlation**, never `ORDER BY created_at DESC LIMIT 1`:

```sql
SELECT created_at, metadata->>'decision', left(body,4000) FROM diagnosis_artifacts
WHERE correlation_id='5d6501ba-db0e-4963-a1ca-e4736d41210e' AND kind='council_report'
ORDER BY created_at DESC;
```

## 22. 2026-08-08 16:13Z — **APPROVED**, and both objections answered rather than filed away

`5d6501ba-db0e-4963-a1ca-e4736d41210e` → **`approved`**, *"approved with 1 advisory objection(s) —
none high-severity"*. Six seats reviewed (editquality, reuse_agent, tooling_provenance, guardian,
diagnosis_guardian, compliance), six abstained on relevance. `gated_by_truncation: false` — worth
checking, given the architecture seat's first reviews were 2 of 3 truncated. **~14 minutes end to
end, not the ~30 budgeted.**

The seats independently reached the framing this lane arrived at the hard way — reuse_agent: *"the
opposite of the founding-incident pattern … extending the one existing mechanism instead of adding
a parallel one"*; guardian: *"blast-radius work is unusually thorough already"*.

**Both criticisms were checkable, so I checked them instead of banking the approval.**

1. **guardian (low, advisory):** *"the one gap is trusting 'only two remaining hand-rolled sites'
   without an independent check."* **Fair — my census was scoped to `platform/ internal/`.**
   Re-run repo-wide, and again case/whitespace-insensitively
   (`grep -rniE "insert[[:space:]]+into[[:space:]]+agent_error_log"`): **3 sites, the same three**
   [MEASURED 08-08 16:1xZ]. The claim holds repo-wide. This is the "a grep proves absence only for
   the spelling *and the scope* it searches" trap, and the seat caught the scope half.
2. **reuse_agent (missing):** *"did not verify whether this exact fix has already been through
   council."* **Checked:** two prior `council_report`s mention both `agenterrors` and `NULLIF`, and
   both are correlation `5c2bc265` — **RFC_012 B's own consolidation**, REJECTED 08-07 08:39
   (for presenting an 8-file sample of 34 as representative) and APPROVED 08-08 15:25 on
   resubmission. **Not this fix.** No duplicate.

> **Context a successor should have, though it is not this lane's to relitigate: the consolidation
> my one-line fix sits on top of was itself rejected once, on blast-radius/sampling grounds.** It
> shipped anyway and was approved a day later — which is exactly CLAUDE.md's "review here is after
> the fact, by design". It also means `agenterrors.Write` became fleet-wide plumbing *between* a
> rejection and an approval, and the `''` shape rode along unremarked in both rounds.

**STATE: approved, and the CODE IS NOT WRITTEN.** The submission describes the edits; nothing is
implemented and nothing is committed against them. That is the next action — see the handoff.
The commit that ships it carries `Council-Reviewed: 5d6501ba-db0e-4963-a1ca-e4736d41210e`
(I have read this verdict, so the trailer is honest). **Docs-only commits carry no trailer** —
the gate refuses docs client-side and `098` joins on platform-code commits.
