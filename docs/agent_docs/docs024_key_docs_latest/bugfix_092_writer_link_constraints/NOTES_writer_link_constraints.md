# NOTES — bugs_open/092, the writer's link constraints

Append-only, newest at the bottom. Missteps are not an appendix — they are the point.

---

## 2026-07-31 — picking the bug, and the check that made the pick trustworthy

`who-owns.py 092` named `bugfix_079_phantom_link_gate` as the likely owner, but that lane
**closed 079 on 2026-07-29** (`c275e6959`), and its last touch on this file was a 07-28
correction. Ownership checks read COMMITS, so a session mid-fix is invisible to them — the
memory index records four collisions caused by exactly that. So I also grepped the live
session transcripts for this bug's **code symbols** rather than its number:

```bash
grep -c -E 'prepare_link_context|PrepareLinkContextAction|link_constraint_text|InjectLinkConstraints' *.jsonl
```

Nine transcripts hit. Every one turned out to be either the MEMORY.md index line quoting the
landmine, or a workflow listing that happens to name the step — none was a session working
it. The one that looked alarming (`8871a7d4`, actively editing and reading `bugs_open/071`,
which shares this class) was on `bugs_open/142`, the undeployed-asset detector.

**The number-grep alone would have been misleading in both directions**: it under-counts
(the sessions reading `071` never type "092") and over-counts (three of the nine hits are
the same memory line loaded into three different contexts).

## The bug was still live — and the second query is what decided the fix

Re-ran the filer's own measurement: **26 of 26** writer runs at `page_count 0`, latest
15:36 UTC today. But the query that actually determined the design was one the bug file did
not have, run over **all** `page-content-writer` orchestrations rather than only those that
recorded a link context:

```
writer_runs | has_input_site_id | has_site_record | has_toplevel_site_id | has_db_sync
         26 |                26 |               0 |                    0 |           0
```

Two things fell out of it at once: candidate 2 is dead (nothing to point `pages_field` at,
confirming the 07-27 triage from the other side), **and** the site id is at
`input_data.site_id`, which the package's shared `extractSiteID` does not look at. Without
that second column I would have wired the DB query to `extractSiteID`, it would have
returned `""` on every real run, and the fix would have failed **exactly the same silent way
as the bug** — an empty list and a warning nobody reads. That is the trap this bug is made
of, and I nearly reproduced it inside the fix.

## Design turn: which predicate, and why "agree with the gate" beat "be correct"

Three functions already load a site's pages, under **two** different predicates:

| function | predicate |
|---|---|
| `loadValidPagePaths` (the deploy gate) | `status NOT IN ('deleted','archived')` |
| `loadResolverPageSet` | `status NOT IN ('deleted','archived')` |
| `loadActivePagesForLinkContext` | `status = 'active'` |

My first instinct was to pick the "safest" one — the strictest, so the writer can never be
told about a page that will not be live. **That is the wrong criterion.** The property that
matters is not strictness, it is *agreement*: `validate_page_content` decides what is a
`phantom_link` using the gate's predicate, so if the writer's allow-list comes from a
different set, a writer that obeys its instructions can have its links flagged. I took the
gate's.

Measured before relying on it: `pages.status` fleet-wide is **only** `active` (449) and
`archived` (23), so the two predicates are the same set today. That measurement is what
turns "they agree" from an assumption into a fact, and it is also why unifying all three
call sites would have been a shared-mechanism change with no present benefit — so I did not.

## Misstep 1 — I wrote a truncation count that would have been wrong by orders of magnitude

The page-list cap is implemented as `LIMIT limit+1`, which detects overflow. I then wrote:

```go
omitted = len(pages) - limit   // always exactly 1
```

and put it in the action output as `pages_omitted_by_cap`. On a 4,000-page site that field
would have read **1**. A confident, precise-looking, wrong-by-three-orders-of-magnitude
number — worse than no number, because nobody re-checks a figure that looks measured.

Caught by re-reading my own diff before committing, not by a test. Fixed by asking for the
real total with a `count(*)` only on the rare over-cap path, and using `-1` as an explicit
"over the cap, exact figure unreadable" marker rather than inventing a plausible one.

**The general shape, which is the transferable part:** a probe designed to answer a
*boolean* ("is there more?") was reused as if it answered a *quantity*. `limit+1` can only
ever prove "at least one more". Logged in `WRONG_CALLS.md`.

## Misstep 2 — my `-1` marker was invisible to my own guard

Having introduced `-1`, the two call sites still read `if truncated > 0`. So the one case
that means "truncated, and I could not even tell you by how much" would have been reported
as **no truncation at all** — the silent cap, reintroduced by the fix for the silent cap.
Changed to `!= 0` with the reasoning written beside it.

Both missteps are the same family: I added a sentinel and did not re-run the checks that
consume it. The memory index has a whole group for this ("your own action can silence your
own detector") and I still did it twice in one file.

## Mutation-checking the tests, because "9 tests pass" is not evidence

The failure mode of this action is *producing nothing and returning nil*, so a test that
asserts "no error" passes against the bug. I wrote nine tests claiming, in comments, that
each fails against the behaviour it guards — then actually ran six mutations rather than
leaving the claim as prose:

| mutation | caught by |
|---|---|
| restore `return ""` on an empty list | `EmptySiteInstructsNoInternalLinks`, `RecordsUnavailable…` |
| drop `input_data.site_id` from the resolver | 5 tests, incl. `ResolvesSiteIDFromInputData` |
| restore the URL synthesis | `NeverSynthesisesAURLFromAName` |
| delete the whole database branch (= the bug as filed) | 6 tests |
| delete the durable-record call | `RecordsUnavailable…`, `TreatsAQueryFailureAs…` |
| reverse the source precedence | `PrefersTheDatabaseOverCollectedData` |

All six caught, each by the test that claims to guard it. This cost about ten minutes and is
the difference between a test suite and a decoration.

## The build was already broken at HEAD, and it is not mine

`go build ./...` fails on `cmd/reasoningset/main.go:504` (three declared-and-not-used vars).
The file is **unmodified in my tree**, so it is broken *at committed HEAD* — someone's
in-flight commit. `./platform/...` builds and tests clean, which is what this change
touches. Verified my change against a clean `git archive HEAD` checkout rather than against
the working tree, because the tree carries several other sessions' edits.

## Trap paid for: `git archive HEAD` into the scratchpad filled a SHARED tmpfs

The recommended way to check "does my change compile against HEAD, not against the tree" is
to extract `git archive HEAD` into the scratchpad. Two such checkouts plus their build
caches came to **~440MB**, and `/tmp` is a **16G tmpfs shared by every concurrent session**
— around thirty of them, several holding 800MB–1.7GB each. It hit 100% while I was writing
docs, and the symptom is that Bash commands start returning
`the temp filesystem … is full … child process's stdout/stderr writes failed with ENOSPC`
— the command may still have *worked* (my `cat >>` append landed; only the output capture
was lost), which is a genuinely confusing failure. Filed in `LANDMINES.md`.

## What I did NOT do, and why

- **Did not widen `extractSiteID`.** It has five other callers, several treating `""` as
  "skip this work". Adding `input_data.site_id` there would silently change what five
  unrelated actions resolve — a shared-semantics change inside a bug patch, which is the
  shape the guardian seat vetoed `bugs_closed/124` for.
- **Did not touch the dead `link_constraints` config block** on `page-content-writer`.
  Deleting the Go function makes it *provably* unread; removing the config is a live change
  with its own risk profile and no benefit today.
- **Did not fix `render_context.available_pages`**, which is empty for the same reason.
  Nothing reads it; recorded in the bug file so it is not rediscovered as a new bug.
- **Did not repair a single deployed page.** This is prevention. The live 404s belong to
  `071` and `097`.

## Council round 1 — APPROVED, and the objection I most disagreed with was the one I was wrong about

`4b8c5e21-011b-40f0-819a-3dfa4b4c7b1d`, approved, 6 advisory objections, none high, 6 seats
abstained on relevance. Turnaround was **~8 minutes**, not the ~30 the runbook budgets.

**`reuse_agent` caught something I had already talked myself out of.** I *considered* calling
`loadValidPagePaths` directly, rejected it because it returns a `PageURLIndex` with no titles,
and wrote a comment asking the two queries to stay in step. The seat's phrasing is the part
worth keeping: *"they checked, named the existing mechanism, and then copied it instead of
calling it — the precise failure mode the founding incident describes, just one layer more
sophisticated (documented duplication vs. blind duplication)."*

That is right, and my reasoning contained a false dichotomy: the choice was never "call the
whole function or copy the whole query". Factoring the **predicate** alone costs one constant
and makes drift unrepresentable, while leaving each caller its own projection. I had framed
reuse as all-or-nothing and picked "nothing".

**Misstep 3 — I asserted what a mutation would do, and it did the opposite.** I wrote in the
new test's doc comment: *"MUTATION: inline the predicate back into either query and this
fails on that query's matcher."* Then I ran it. It **passed**. The matcher compares the
constant's *text*, and an inlined copy of the same text still matches.

On reflection the test is right and my comment was wrong: two copies that say the same thing
have not drifted. What the test actually pins is **divergence** — I verified that by adding
`'draft'` to the gate's list, which does fail it. The comment now states both, including
what it does not catch.

The general lesson, and it is the third time in one session: **a claim about what a check
would do is not the check.** I wrote a confident mutation prediction into a comment that
would have been read by every future maintainer as verified. The only reason it is not still
there is that I ran it — and I nearly did not, because it was "obvious".

**Two objections dissolved on measurement, which is the cheaper answer than an edit:**

- `editquality` gated on the deletion-safety claim being un-confirmable from SQL. It is
  confirmable: **no `agent_definitions` row references `InjectLinkConstraints` or
  `inject_link_constraints`**, and it is not a registered action name.
- `guidelines` raised DECLARED CONTRACTS. **0 of 185 active agents declare `input_contract`
  or `output_contract`** — the guideline is inert fleet-wide, exactly as that seat suspected.
  Worth reusing: this objection will recur on every submission until the guideline is fixed
  or retired, and one query settles it.

**`bug_historian`'s sibling audit was the most valuable objection**, because it found
something outside my change: of `extractSiteID`'s five callers, two do not fail loudly, and
one (`ExtractAndSyncLinksAction`) returns a *success-shaped* `{"links_extracted": N,
"persisted": false}` when the site id does not resolve. `link_registry` has **0 rows
all-history**. I stopped short of the obvious conclusion — the action runs on one agent with
**0 orchestrations in the retained window**, so "the exposure fires" and "the agent never
runs" are indistinguishable, and asserting either would reproduce this bug's own error.
Recorded as `[UNDETERMINED]`, contributed to `bugs_open/165` (which owns that table) rather
than filed as a competing bug.

## Why this bug is NOT being closed, against the standing instruction to close it

`CLAUDE.md`: *"The bar is **fixed AND live** — a fix committed but inert until the next roll
stays OPEN, because the defect is still reproducible until it ships."* Go changes are inert
until an image is built and rolled. This lane committed; it did not build or roll, because
the task was framed as "commit files for the next chassis build" and a fleet roll would ship
every other session's committed work too — not this lane's call to make unprompted.

So `092` stays in `bugs_open/` with a status banner naming exactly what closes it: a fresh
writer run recording `page_count > 0` and `source: database`, plus a pod-grep for
`LINK_CONTEXT_UNAVAILABLE` **with a positive control in the same exec**.

Incidental, and worth knowing before anyone builds: **`cmd/reasoningset` does not compile at
HEAD** (`main.go:504`, three declared-and-not-used variables), unmodified in this tree, so
it is someone's committed breakage. It does **not** block the chassis: each service's
dockerfile builds only its own `./cmd/<service>`, and `go build ./cmd/agent-chassis/`
succeeds. `go build ./...` does not, which will mislead the next session that runs it.

## 2026-07-31 19:09–19:16Z — LIVE on v1.0.1219, induced, CLOSED

**Pod-grep, both replicas, both directions.** New strings present (`LINK_CONTEXT_UNAVAILABLE`,
"There are NO pages available to link to"), **removed** strings absent (`## INTERNAL LINKS`
from the deleted duplicate, `## Internal Links` from the old heading), control
`PrepareLinkContextAction` = 8. The *removed* rows are the half a one-directional check
misses: a new string proves the binary contains it, but only the absent old strings prove it
is the NEW binary rather than one that merely happens to include the literal.

**A gap I have to state rather than paper over:** the two commits (`2e1bfb39e` fix,
`9a57d2395` review answer) are **indistinguishable by pod-grep**, because the second only
extracts a shared constant that Go constant-folds into byte-identical SQL. There is no marker
to look for and it would be dishonest to imply otherwise. Its guarantee is enforced by a test,
not by the binary.

**Induction, and why the target mattered more than the dispatch.** Writer runs are irregular
(26 today, in morning bursts, then roughly hourly), so waiting was not a plan. I picked
`loancalculator.co.uk/guide-can-i-overpay` specifically because it is `rebuild_policy='owned'`
on an `active` (not `deployed`) site: `save_page_sections` refuses `owned` pages, so **the run
could not write anything even if everything else went wrong**, while `prepare_link_context`
runs long before that refusal. Confirmed after: handler `complete_error`, writer `complete`,
and `pages.updated_at` for the target still 2026-07-30 22:10 — untouched.

That is the inverse of the trap the 079 lane hit: they induced successfully and got a **null
result** (`checked_links: 0` — the page had no anchors, so the repair had nothing to repair).
My verification sits *upstream* of content generation — `link_context` is recorded before a
single word is written — so the run's content outcome cannot make the evidence vacuous. Worth
knowing when choosing an induction target: **ask where in the pipeline your evidence is
recorded, and pick a target that guarantees you reach it, not one that guarantees success.**

The row:

```
 pages | source   | db_consulted | degraded | text_len | reason
    27 | database | true         | false    |     2739 | 27 linkable page(s) read from the pages table
```

against `0`/`null`/`0` on all 26 pre-fix runs, and **27 of 27** listed addresses match a
stored `pages.url` exactly — so the synthesis path is not merely unreached, the output is
provably the stored truth. `agent_error_log` has **0** `LINK_CONTEXT_UNAVAILABLE` rows, which
is the correct result and also confirms the loud arm is not firing spuriously.

Fixed AND live AND proven ⇒ moved to `bugs_closed/092`.

**Residual, unchanged by any of this:** it repairs no already-deployed page (that is `071`
and `097`), and the `extractSiteID` sibling exposure remains `[UNDETERMINED]` with the
measurement contributed to `165`.
