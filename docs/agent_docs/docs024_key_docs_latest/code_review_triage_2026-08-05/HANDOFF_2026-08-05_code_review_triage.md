# HANDOFF — triage of the 2026-08-05 `/code-review` findings

A `/code-review` run over the working diff returned **15 findings**. I have **triaged, not
fixed**, them: ownership resolved, each claim checked against code or the live DB where that
was cheap, and two rated differently from the reviewer. **Nothing here is mine** — every
cluster belongs to a lane that committed it on the evening of 2026-08-04.

**Start here, then go to the cluster you own.** Do not work findings outside your cluster
without checking `scripts/who-owns.py` first — three lanes are involved and all three were
active last night.

---

## Ownership — three lanes, one fleet item

| cluster | files | owning commit | findings |
|---|---|---|---|
| **195 lane** — permanent/transient classifier | `platform/messaging/validation_drop.go`, `platform/agentbase/agent.go`, `platform/errors/errors.go`, `validation_drop_test.go`, `agent_test.go` | `28ef7a044` 08-04 20:55 | **F1, F2, F5, F8, F14** |
| **194 lane** — `save_page_sections` metadata source | `platform/orchestration/actions/save_sections_metadata_source.go` | `47ee3ebce` 08-04 21:00 | **F3, F4, F7, F9, F10** |
| **156 lane** — duplicate-section collapse | `platform/orchestration/actions/save_page_sections_action.go` | `84b7d561c` 08-04 21:26 | **F6, F11, F12** |
| **fleet / unowned** | `.gitignore`, `check_endpoint_health_action.go` | — | **F15, F13** |

---

## My triage verdicts, in the order I would act

### 1. F15 — 141 MB of untracked binaries in the repo root. **DO THIS FIRST.**

**Verified, and it is worse than "worth tidying".** `config-key-audit` (86M), `scheduler`
(28M), `git-adapter` (25M), `reasoningset` (3.1M), plus `live.html` and `clearideas.bash`:
all untracked, and `git check-ignore` matches **none** of them.

Why it is the top item despite being the least interesting: this tree is **forward-only**
(no resets, no amends) and CLAUDE.md records that sessions here still run `git add -A` — it
names the commit where that swept another thread's work. **One such commit puts 141 MB into
history permanently**, slowing every clone and CI checkout for ever, with no way back.

It costs one line each in `.gitignore` and is pure protection. I did not do it only because
`.gitignore` is shared and someone may be deliberately holding those paths — **confirm
nobody needs them tracked, then add them.**

### 2. F7 — `require_sections_metadata` collides with a LIVE key of different meaning. **194 lane.**

**Verified against the live DB.** The key already exists on **`content-reviewer`** and
**`page-build-handler`**, where `validate_page_content_stats.go` reads it to emit a
*warning-level* `stat_audit_unavailable` issue. The 194 change gives the same spelling, on a
`save_page_sections` step, the meaning **"refuse the save and return an error"**.

That is a genuine trap: an operator copying the declaration between steps of one agent, or
any `jsonb_set` sweep over `{workflow,steps,*,config,require_sections_metadata}` — the
natural way to roll a declaration out — arms a hard refusal on the fleet's highest-traffic
save path. **Rename one of them.** Cheap now, expensive after the first sweep.

### 3. F1 — the `de.Retryable` early return contradicts its own comment. **195 lane. Real, latent.**

**Code shape confirmed** (`validation_drop.go:112`): `if de.Retryable { return "" }` returns
before the substring fallback, while the comment six lines above states *"This change only
ever ADDS permanent classifications; it removes none."* For any `AsRetryable` DomainError
whose text contains a validation needle, it now **removes** one.

**Severity check I ran that the fix should not skip:** `grep -rn "AsRetryable(" --include=*.go`
finds **no live producer** outside comments. So it is latent — the builder invites the first
one, and the first one silently reinstates the infinite-retry loop the drop path exists to
prevent. **Fix the comment or the code so they agree**; the reviewer is right that they
currently do not.

### 4. F9 — the regression detector's "before" query is scoped narrower than the DELETE it predicts. **194 lane.**

`countExistingRowsWithContentData` filters `build_status = 'deployed'`; the DELETE it exists
to predict carries **no** `build_status` predicate. So content_data destroyed on a
`needs_rebuild` or mid-build row is **never reported** — the exact condition the detector was
built to end. Same family as `bugs_closed/185`'s whole thesis (`build_status = 'deployed'` is
not "this page is live"), and worth handing to the 194 lane with that reference.

### 5. F5 — unbounded `agent_error_log` writes on every non-permanent failure. **195 lane.**

One INSERT per failed message per attempt, fleet-wide, with no rate limit, no dedupe and no
`site_id`. The same table is bounded elsewhere against exactly this
(`maxReportedConditionsPerResponse = 10`). The diagnosis loader reads it newest-first, so a
retry storm evicts informative rows from the window it feeds. Also synchronous with a 5s DB
timeout in front of `handleProcessingError`. **Judgement call for the lane** — it is
defensible if the volume is genuinely low, but the volume has not been measured and the
reviewer is right that nothing bounds it.

### 6. F2, F14 — tests that assert nothing, or assert the wrong thing. **195 lane.**

- **F2:** `TestAgentbaseUsesSharedPermanentClassifier` calls `messaging.MatchedPermanentFailure`
  directly and never exercises any `agentbase` code, so it cannot detect the drift its own
  comment claims to guard. The reviewer supplied the mutation: revert `agent.go:1192` and the
  suite still passes. **This is the "a guard proved on nothing" shape this estate keeps
  logging** — worth fixing precisely because it currently reads as coverage.
- **F14:** two packages hard-assert that `MatchedValidationNeedle` **misses** `WORKFLOW_INVALID`,
  pinning a known defect as required behaviour. Anyone later case-folding the comparison — a
  plausible one-line improvement — breaks two tests in packages they did not touch. Pin the
  miss on the classifier under test, not on the legacy helper.

### 7. F3 — the implicit metadata default can pre-empt an explicitly configured `html_field`. **194 lane. Needs the lane's judgement.**

The reviewer's mechanism is plausible and specific (`ExtractNestedField` auto-unwraps a
`.response` map at every path segment, so another agent's reply can resolve the default
path). I did **not** verify the reachability end-to-end — that needs the 194 lane's knowledge
of what actually lands in `collected_data` on that path. **Treat as unconfirmed-but-credible**
and check the reachability before either fixing or dismissing it.

### 8. F8, F10, F11, F12, F13 — real but minor. **Batch them.**

- **F8** (`errors.go:218`): `IsRetryable`/`GetRetryAfter` still use bare `err.(*DomainError)`
  in the very file that added `AsDomainError` to stop callers doing that. No live caller, so
  the first one inherits the bug.
- **F10**: a fourth verbatim copy of the `agent_error_log` INSERT instead of calling
  `LogAgentError`, dropping `orchestration_id` — so a `CONTENT_DATA_REGRESSION` row cannot be
  joined to the run that caused it. The package's own precedent forbids the third copy.
- **F11**: `sections_with_content_data` is counted *before* the insert loop, so it counts
  sections the save then discards (locked slots, unresolvable stubs).
- **F12**: an unconditional extra DB round-trip on the highest-traffic save path in two cases
  where the answer provably cannot change the outcome.
- **F13**: a comment's own insertion invalidated the `:216`/`:215` line anchors it cites.
  Trivial, but the reviewer notes two more instances in the same diff.

---

## The one finding I rate DIFFERENTLY from the reviewer

### F4 — **FALSE POSITIVE. Do not act on it.**

The reviewer asserts the `declared_absent` exemption is *"unreachable in production — no live
agent definition sets `expects_no_sections_metadata`"*, and builds a failure scenario in which
`tool-recreation-handler` emits `CONTENT_DATA_REGRESSION` noise on every run.

**Checked against the live DB: exactly one agent carries the key, and it is
`tool-recreation-handler`** — the very agent in the scenario. The exemption is reachable, it
is seeded on the right definition, and the predicted noise does not occur.

**Recorded because an unchallenged false positive is expensive**: it names a specific agent
and a specific recurring symptom, so the next reader has every reason to believe it. One
query settles it:

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%expects_no_sections_metadata%';
```

**The general lesson, which is the reviewer's own doctrine turned on itself:** an
absence claim ("no live definition sets X") is exactly the shape this estate requires a
lookup for, and the review asserted it without one. Apply the same bar to a review's findings
as to a submission's claims.

---

## How I triaged, so you can extend it

Cheap and mechanical, in this order:

1. **Ownership first** — `git log -1 --format='%h %ad %s' -- <file>` per file. Three commits
   accounted for 13 of 15 findings, which is what makes the split by lane obvious.
2. **Is the code shape still there?** Read the cited lines. (All confirmed; nothing had moved.)
3. **Is it reachable?** This is what separates F1 (latent — no live `AsRetryable` producer)
   from F7 (live — the key is seeded on two agents today) and kills F4 outright.
4. **Do not re-derive the reviewer's reasoning** where it is sound and checkable — check the
   premise instead. Every verdict above rests on a grep or a query, not on re-reading prose.

## What I deliberately did not do

- **No fixes.** Every cluster belongs to an active lane; per CLAUDE.md, contribute into the
  owning lane rather than starting a competing fix. `scripts/who-owns.py` before routing.
- **No bug files filed.** F7, F9 and F5 are probably worth `bugs_open/` entries if their lanes
  will not take them promptly — but filing them *at* a lane that is mid-flight is the exact
  duplicate-work failure `who-owns.py` exists to prevent. Ask the lane first.
- **No `.gitignore` edit** — see F15; it is shared and needs one confirmation, not a
  unilateral change.

---

# ACTIONED — 2026-08-05, later the same day

Appended, not edited: the triage above stands as it was written. Three of its judgements were
overtaken by events or by measurement, and those are marked as corrections below rather than
changed in place.

**The operating premise expired within the minute.** "Every cluster belongs to an active lane"
was true at 11:02 and false at 11:03: `194` and `195` both closed to `bugs_closed/`, and
neither lane's `HANDOFF_2026-08-05_continue_here.md` mentions this review (grepped for
`code.?review|triage|F1|F7|F15` — zero hits in either). Ten of fifteen findings were therefore
unowned, and were fixed.

## Disposition of all fifteen

| # | verdict | what happened |
|---|---|---|
| F15 | confirmed | Fixed, `79c713bff`. The `.gitignore` already had a block for this exact accident (a 93MB binary committed 2026-07-20, `bca5d8255`) — four lines appended to it. |
| F1 | confirmed, latent | Fixed, `f887ed1ad`. Comment corrected; code was right. |
| F2 | confirmed | Fixed, `f887ed1ad`. **Both** agentbase tests had the flaw, not the one named. Replaced with an AST guard, proven by the reviewer's own mutation. |
| F8 | confirmed | Fixed, `f887ed1ad`. Zero callers re-proven by the compiler after a council objection to the grep. |
| F14 | confirmed | Fixed, `f887ed1ad`. |
| F7 | confirmed, worse than filed | Fixed, `fa30062cc`. See correction 2. |
| F9 | confirmed | Fixed, `fa30062cc`. |
| F3 | latent, unreachable | Made visible (warning), not fixed, `fa30062cc`. All three live callers are explicit. |
| F10 | **remedy refuted** | Secondary half fixed, `fa30062cc`. See correction 3. |
| F11 | confirmed | Fixed, `6e607da1e`. See correction 1 — not 156's. |
| F12 | confirmed | Fixed, `6e607da1e`. See correction 1 — not 156's. |
| F5 | **FALSE POSITIVE** | Not filed. See correction 4. |
| F4 | false positive | Unchanged — the triage above was right. |
| F13 | confirmed | **Deliberately not fixed.** See below. |
| F6 | unactionable | See below. |

## Corrections to the triage above

> **CORRECTION 1 — the ownership table is wrong for F11 and F12.** §"How I triaged" step 1
> uses `git log -1 -- <file>`, which is file-granularity. Two lanes touched
> `save_page_sections_action.go` 26 minutes apart on 08-04: the file's last commit is
> `84b7d561c` (**156**, open), but `git blame -L 624,628` puts both findings' lines in
> `47ee3ebce` (**194**, closed). They were unowned, not 156's. **Blame the lines a finding
> cites, not the file.**

> **CORRECTION 2 — F7 is more concrete than "an operator copying the declaration".** One live
> definition already holds both meanings: `page-build-handler` has a `save_sections` step
> (action `save_page_sections`) and a `validate_content` step (action `validate_page_content`,
> `require_sections_metadata=true`). And the collision had **already misled a reader in this
> repo** — `save_sections_link_repair.go:13` described the two keys as one step's config.
> Corrected in place, kept as evidence. Renamed the save-side key to
> `refuse_save_without_sections_metadata`; measured `0` live save steps carrying the old one.

> **CORRECTION 3 — F10's primary remedy is impossible, and its premise is backwards.**
> `platform/orchestration/coordinator.go:23` imports `actions`, so calling
> `orchestration.LogAgentError` from there is an import cycle; ~20 files in that package each
> carry a local INSERT for exactly that reason. "The package's own precedent forbids the third
> copy" — the precedent *is* the copy. Only the dropped `orchestration_id` was real; fixed.

> **CORRECTION 4 — F5 is the review's SECOND false positive, not a judgement call.** Rated
> above as "defensible if the volume is genuinely low, but the volume has not been measured".
> Measured: the write-side facts hold, but the predicted harm does not.
> `scheduled_tasks.database-cleanup` (enabled, hourly, `last_triggered_at` minutes old) deletes
> unresolved rows past 30 days — the oldest surviving row sat **21 minutes** inside that line.
> The rows also carry `site_id` NULL **and** `domain = ''`, so they match neither a
> site-scoped nor a domain-scoped diagnosis load; only a fully unscoped one sees them. The
> writer is ~7% of today's volume, and the 20× explosion began 08-03, two days before this code
> shipped. **No bug filed** — one asserting unbounded growth on a bounded table would be a
> durable false claim.
>
> Recorded in `WRONG_CALLS.md`: I reached "no reaper" first, from a `*.go` grep (the reaper is
> SQL in a `scheduled_tasks.pre_query` column) and from "oldest row is 30 days old" — a figure
> produced identically by a working 30-day reaper and by no reaper at all. **Both false
> positives in this review are absence claims made without a lookup, which is the lesson this
> handoff draws for F4 and which I then reproduced while checking it.**

## Two that were NOT actioned, and why

**F13 — confirmed, left for its owner.** The anchors are wrong exactly as filed:
`markTaskComplete` is at **:235** and the `config["task_name"]` read at **:237**, against the
comment's `:215`/`:216` — the comment's own 21-line insertion shifted them (`216 + 21 = 237`).
Not fixed because `check_endpoint_health_action.go` carries **25 insertions / 4 deletions of
another session's uncommitted work**, and the comment block holding those anchors is entirely
on the `+` side of their diff. Editing it would alter their in-flight work, and a pathspec
commit cannot exclude a same-file passenger. **Whoever owns that file: the corrected numbers
are above.**

**F6 — unactionable, and this is a gap in the triage.** F6 appears only in the ownership table
and is described in none of the numbered verdict sections. The original `/code-review` output
was never saved — this directory held only the handoff — so there is no record of what F6
claimed. Flagged rather than guessed at. *A triage summarising N findings should carry every
finding's text, or say where the raw output lives.*

## Paper trail

`PLAN_2026-08-05` · `NOTES_` · `RUNBOOK_` · `README_where_we_are` · `SUMMARY_2026-08-05`, all
in this directory. `WRONG_CALLS.md` (the reaper misstep). Commits `79c713bff` (F15),
`f887ed1ad` (195 cluster), `fa30062cc` (194 cluster), `6e607da1e` (F11/F12), `90678beba`
(wrong-calls). Council: `128d4fd1` **APPROVED** (2 low-severity advisories, both answered);
`cb575682` and `d0d2c97a` pending at the time of writing — **the code is already on the shared
branch, so a REVISE/REJECTED on either is still owed a response.**
