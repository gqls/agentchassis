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
