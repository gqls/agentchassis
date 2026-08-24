# NOTES — bugs_open/358: agent_error_log finding codes are write-only and expire unread

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

## 2026-08-22 — session start: ownership + re-validation

- `scripts/who-owns.py 358` → no owning workstream; the `bugfix_238_regeneration_key_loss` lane
  cites it as a cross-reference only, and their NOTES (2026-08-22 entry) say 358 was filed AT
  their request with this class deliberately handed to whoever picks it up. No open
  `site_work_items` row touches the class. This lane now owns it; this directory is the record.
- **Census re-run against the live DB (see RUNBOOK), two deltas since the bug file was written
  this morning — neither invalidates the class:**
  1. `resolved` is now **48**, not 0. All 48 set today 10:40 UTC by
     `resolved_by = 'content-loss-check:healed'` / `'content-loss-check:row_gone'` on
     `CONTENT_KEY_LOSS` (40) and `STRUCTURAL_KEY_CARRY_MISS` (8) — the 238 lane's consumer
     shipped (`cba51ad1d`) and ran. So "the resolved workflow has never been used once" is now
     stale BY ONE DAY, and the first user is exactly the reader-ships-with-writer pattern 358
     holds up as the positive example. The §8 trap stands: resolving halves remaining life to
     14 days; content-loss-check extracts (heals) before resolving.
  2. New code `CONTENT_KEY_LOSS` (72 rows, all 2026-08-22, agent_type `content-loss-check`) —
     written AND consumed by the same binary. Not a new member of the unread class.
- Commit `0ce242d9c` (today) added a THIRD recorder in the validation-gate family:
  `CONTENT_VALIDATION_WARNING_DETAIL`. 0 rows yet (inert until image roll). Reader status being
  verified — if unread, the class grew by one TODAY, while the bug file was being written,
  which is §3's self-sustaining mechanism observed live.
- Headline counts still hold: 45,507 total (was 45,426), `RESOLVER_CONFLICTING_CANDIDATES`
  9,617 (was 9,615) and still the loudest, oldest row still 2026-07-23 (retention live),
  `REVIEW_SUPERSEDED_BY_PASSING_SAVE` (25 rows, 07-23 only) days from deletion,
  `TRUNCATION_DEGRADED_REVIEW` dies ~08-25.

## 2026-08-22 — the census method itself is unreliable, which IS the finding

Trying to reproduce 358 §2.1's grep census produced junk on the first two attempts, and that is
worth recording rather than tidying away, because it is the argument for a registry.

- **`grep -ohE 'ErrorCode:\s*"..."'` plus a regex for `*Code = "..."` constants returned 47
  strings including `A`, `CODE`, `FIRST`, `SECOND`, `SOME_CODE`, `X_AUDIT` — test fixtures — while
  MISSING every one of `validationDetailErrorCode`, `linkRepairErrorCode`, `claimsFloorErrorCode`
  and nine more.** The recovery was to stop pattern-matching and resolve each identifier by name.
- **A fourth blindness, beyond the const one 358 §3.2 records:** several codes reach the table as
  a **positional argument** to `LogActionError(ctx, params, siteID, domain, action, code, …)` —
  `tool_birth_instance_scope_refused`, `RETRACTION_AUDIT`, `component_write_shared_blocked`,
  `PLAN_PAGE_SAME_NAME_IDENTITY_HELD`. An `ErrorCode:` grep cannot see any of them.
- **Conclusion that shaped the design:** if three careful attempts at the source census disagree,
  the source is the wrong authority. `SELECT DISTINCT error_code` cannot be got wrong this way.

## 2026-08-22 — CORRECTION 4 to the bug file: `BUILD_DISPATCH_STALLED`'s closed loop is not live

358 §2.2 lists it as one of three codes with an automated reader, evidence for the
"reader-with-writer from birth" law. Both halves are real IN THE FILE
(`214_build_dispatch_watchdog.sql:104` reads, `:120` writes) and **neither is live**:

```sql
SELECT filename FROM schema_migrations WHERE filename LIKE '214%';   -- 0 rows
SELECT name FROM scheduled_tasks WHERE name ILIKE '%watchdog%';       -- 0 rows
```

Migration 214 was never applied. Its zero row count reads as "quiet" and means "absent" — the
two are indistinguishable from the table's side, which is exactly why the check reports the
registered-minus-observed direction and never fails on it.

## 2026-08-22 — the `090` came back UNVERIFIABLE, and its REASON is evidence for the design

Run correlation `c965bfec-993a-4b2b-88ba-d44549c81df1`, filed on the five-writers claim.
Verdict: **UNVERIFIABLE (stopped: scope-not-narrowing)**. **This is not a refutation and must not
be reported as a confirmation either.** What it said it still needed:

- the `pgxpool`-vs-`*sql.DB` question could not be settled — *"the symbol search for pgxpool
  found 0 rows but the index holds no `type` kind declarations at all (kinds present: alias,
  const, func, interface, method, struct, var), so that 0 is unrepresentable-not-absent"*;
- `214_build_dispatch_watchdog.sql` was **absent from the bundle entirely** — *"SQL files are
  outside the .go-only indexed corpus and cannot be code_request'd"*;
- `cmd/content-loss-check/main.go` returned 0 rows, *"explicitly flagged as possible index
  staleness on unpushed work, not proof the binary is absent"*.

So the loop's static tier is `.go`-only, index-backed, and holds no `type` declarations — and the
claim under test is precisely about writers spread across Go, SQL and `cmd/`, two of which it
cannot see. **A claim can be outside what an instrument can grade without being outside what can
be verified**: all five files were read first-hand and their type declarations quoted. Recorded
here rather than in `WRONG_CALLS.md` because nothing was claimed wrongly — but the *shape* is
worth knowing before filing a 090 about anything that is not Go.

## 2026-08-22 — council round 1: REVISE, and the objection was right

`be1fd678-0836-4f32-90a6-8927b2463fee`, gated by `editquality` at HIGH:

> *"Both this edit and edit 5 call `findingCodeRegistryCodesExcept(t, ...)` but no edit in the
> plan defines this helper… Without it these test files will not compile — the repointing this
> edit exists to do is a no-op until the helper exists."*

Correct, and it is the second time this lane has nearly shipped an edit whose stated effect
depended on something not in the plan. Auditing the plan the same way found **a second omission
of exactly the same kind**: `cmd/config-key-audit/main.go`'s mode dispatch, without which
`--finding-codes` is unreachable — a mode that exists and cannot be called. Round 2 resubmitted
on the same correlation with both, and with sketches taken from the working code rather than
composed. Cost of the round: one resubmission. Cost of not having it: a commit whose two
headline edits were inert.

## 2026-08-22 — what is built, and the controls it passes

`--finding-codes` (mode 16 of `cmd/config-key-audit`), registry of 53 entries, 19 tests.
Against the live table: **43 normalised codes observed (44 raw — the two
`tool_crosslink_not_emitted:*` variants collapse), 0 findings, 32 unruled, 10 registered but
unobserved.** Controls, all run this session (RUNBOOK has the commands):

| control | result |
|---|---|
| real live list | exit 0 |
| + `TEST_UNREGISTERED_X` | exactly one `undeclared` finding, exit 1 |
| remove it again | exit 0 — so the check can come out clean, i.e. it discriminates |
| a raw `tool_crosslink_not_emitted:<new reason>` | exit 0, collapses to the family key |
| `consumed` reader repointed at a real file that does not name its code | `reader-does-not-name-code`, exit 1 |
| the true reader | exit 0 — so the rejection was about the reader, not about any change |
| empty input | refuses: empty stdout, exit 2 from the compiled binary |

⚠ **The registry mutation was done on a COPY** (`scratchpad/reg_wrong_reader.json` via
`--registry`), never on the shipped file — `WRONG_CALLS.md` 2026-08-22 records a session mutating
a shared file in place to prove a guard and another session committing it during the window.
That is also why the roster test's predicate was split out pure: it can be mutation-proved
against a fixture instead.

⚠ **`go run` collapses the child's exit status** (LANDMINES): the empty-input control read as
`exit=1` under `go run` and `exit=2` from the compiled binary. The signal to branch on is
**empty stdout**, exactly as `audit-optional-key-budget.sh` records. First reading of that control
was logged as a failure before the landmine was recalled — no claim was published, so it is here
rather than in `WRONG_CALLS.md`.

## 2026-08-22 — council round 2: APPROVED, with three advisory objections, all three checked

`be1fd678-0836-4f32-90a6-8927b2463fee`, round 2, **approved** (13 reviewers, 4 abstained, no
gating objection). All three advisories came from `guardian` and none was waved through:

1. **[medium] a new `_test.go` in `platform/orchestration/actions` — a package essentially every
   pipeline imports — so a compile failure there blocks `go test ./...` for all of them, not just
   the two tests it fixes.** True. Exercised in isolation before and after repointing:
   `go test ./platform/orchestration/actions/` green, and green again from `git archive HEAD`.
2. **[medium] the roster read couples two previously self-contained unit tests in a core package
   to a file under `docs/`; a stage that copies only source now fails them for an unrelated
   reason.** A real new coupling. **Measured:** nothing runs `go test` in a stripped context today
   — no Dockerfile in the tree runs it, no CI workflow directory exists, and `.dockerignore`
   strips `*.md`, not this `.json`. So it is real but unexercised. **Not removed**, because the
   only way to remove it is to compile a copy of the roster into the package — the third
   hand-maintained roster, i.e. the exact drift being retired. Instead the failure message now
   names the coupling, says the fix is to carry the file, and says explicitly not to work around
   it by hard-coding the list back. Failing rather than skipping is deliberate: a skip lets a
   collision through in the one environment where nobody is watching.
3. **[low] confirm no other script parses `os.Args[1]` positionally in a way the new branch could
   shadow.** Checked all seven callers: every one passes either **no args** (`audit-config-keys.sh`,
   the default mode) or an explicit `--<mode>` flag. `--finding-codes` is a distinct literal and
   shadows nothing.

The seat's own note is worth keeping: *"No architecture-change signal here… Every production-path
edit is additive… Blast radius is real but bounded to CI/test-time coupling in a shared package
plus a new docs->test-time file dependency."* That matches the plan's §8 scope argument, so the
council-gate-not-RFC call was right.

**Round cost:** one resubmission. **Round value:** the round-1 gating objection caught two edits
that would not have compiled, and round 2's advisories caught a coupling I had created without
naming. Both rounds found something real, which is now 4 of 6 for this lane's experience of the
gate.

---

## 2026-08-22, later — Task A: the expiring code extracted, and the silence explained

Owner ruled four things on the handoff's open questions (chat, 2026-08-22): extract the expiring
evidence AND establish whether it is fixed or blind; retention gets a longer clock for deliberate
findings and `resolved` stops shortening a row's life; dispositions for the 32 are PROPOSED by the
session and ratified by him in batches; backlog enforcement stays visible-only pending batch 1.

### A1 — extracted

All 41 `TRUNCATION_DEGRADED_REVIEW` rows to
`EVIDENCE_truncation_degraded_review_2026-08-22.json` (48,785 B, parses, `row_count` 41 and
`len(rows)` 41 agree). **24 distinct rounds, 12 seats.** Distribution: edit-quality 20, prior-art
5, guardian 3, architecture 3, then eight seats with 1–2. `agent_type` is **generic 37 /
council-gate 4** — i.e. this is overwhelmingly the DIAGNOSIS council, not the gate.

### A2 — fixed, not blind. And the code is misnamed for its own population

The handoff framed this as "either truncation stopped or the detector went blind, and from this
table they look identical". Both halves turned out to need correcting.

**First: what the 41 rows actually record.** Branch counts: `unsalvageable_invalid_json` 27,
`salvaged_from_invalid_json` 13, `producer_marker` **1**. So 40 of 41 come from the
`json.Valid(rb) == false` arm at `diagnose_council_decide_action.go:341`, whose log line calls it
"invalid JSON, likely truncated at max_tokens". **The token data refutes that for those 40:**

| check | result |
|---|---|
| review calls whose `output_tokens = max_tokens`, all retained history | **5**, ever |
| — of those, dated before this writer existed (2026-07-26) | 4 (all 2026-07-18) |
| — the fifth | 2026-08-02, `review_adoption_guardian`, `max_tokens=120` |
| reviewer replies ending in `}` | **11,514 of 11,519** |

Only the `producer_marker` row is a confirmed truncation. The other 40 record *"this seat's output
did not parse"*, cause unmeasured. The name asserts a mechanism the rows do not carry.

**Second: the silence since 2026-08-02 is a FIX.** The control is the sibling channel written by
the *same function*, before the damage record and on every round: `diagnosis_artifacts`
`kind='council_report'`, whose metadata carries `unreadable` and `gated_by_truncation`
unconditionally (`:468` — the unconditional emission is bugs_open/138's deliberate design, so its
absence means "old row", not "measured clean").

| week | council reports | carrying the key | unreadable seats |
|---|---|---|---|
| 2026-07-20 | 162 | 0 | **17** |
| 2026-07-27 | 202 | 104 | **24** |
| 2026-08-03 | 148 | 148 | **0** |
| 2026-08-10 | 198 | 198 | **0** |
| 2026-08-17 | 248 | 248 | **0** |

That is a **demand control**: the recording path is proven executing 248 times in the last week
and reporting clean, which a zero row count in `agent_error_log` could never establish on its own.
Had it been blind, `unreadable` would be non-zero with no matching finding row.

**[CORRELATED, NOT PROVEN CAUSAL] what changed:** reviewer ceilings moved 8000 → 16000 across the
same weeks — 5 calls at 16k in wk 07-20, 447, 636, 1,369, **1,756** by wk 08-17 — while failed
review calls fell 59/58 → 23/9/18. Plausible mechanism (more headroom ⇒ replies complete ⇒ JSON
parses) and a tight temporal match, but I did not find the commit that raised them and am not
claiming it.

**Not a live bug:** `review_adoption_guardian` at `max_tokens=120` on 2026-08-02 guaranteed
truncation (42-char reply). It runs at 8000 today with outputs ~300 tokens. One-off, corrected.

### Missteps in this session's own measurement

1. **A vacuous control, caught by looking at the denominator.** My first ceiling query compared
   `total_output_tokens >= wire_max_output_tokens` and returned a clean 0 for every week. Both
   columns are **100% NULL for review calls** (0 of 11,698 populated) — the comparison was against
   NULL and could not have come out any other way. The populated pair is `output_tokens` vs
   `max_tokens`. This is exactly the `WRONG_CALLS.md` shape: *a measurement that could not have
   disconfirmed*. Caught only because I ran a `count(col)` over the columns before trusting them.
2. **Reached for `severity` as the retention discriminator and had to abandon it.** Findings are
   written as error/warning/info and plumbing as error/fatal/warning, with three codes emitting
   mixed severities (`CONTENT_CLAIMS_FLOOR_DETAIL`, `CONTENT_DATA_ENVELOPE`,
   `FIX_PLAN_VALIDATION_REFUSED`). Nothing in the row except the code itself separates a
   deliberate finding from a timeout. Recorded because it is the obvious idea and the next person
   will have it too.
3. **`cd` persists between Bash calls in this harness.** A `cd` into the lane directory silently
   made the next two repo-root commands fail with "No such file or directory". Use absolute paths
   or re-`cd`.

### Measured, for Task C

**0 scheduled tasks select on `error_code` at all** (`SELECT name FROM scheduled_tasks WHERE
pre_query LIKE '%error_code%'` → 0 rows, 2026-08-22). Every reader that exists is in-process,
inside the binary that wrote the row. So "something automated will read this" is false by default
when judging the 32, not a matter of opinion.

---

## 2026-08-22, later still — Task B: the retention migration, and a verify block that passed a mutant

### The collision I did not expect, and the design it forced

`566_database_cleanup_reaps_every_terminal_status.sql` was already sitting **untracked** in
`sql_for_agents/`, written by the `bugs_open/354` lane, editing **arm 3** of the very
`pre_query` my change edits **arm 1** of. It pins the whole text by md5 before AND after. Two
migrations pinning one 90-line row means whoever lands second is refused and has to re-derive.

So 567 accepts **either** known text — `c26ccf49…` (pre-566) or `b4deb963…` (post-566) — and
refuses anything else. That keeps the property 566's header rightly argues for (the input is
known EXACTLY, so `replace()` is deterministic) while removing the ordering constraint. Its
negative controls assert arm 3 survives **in whichever form it is in**. Different arm, different
table, disjoint anchors.

I also messaged the other lane — and **messaged the wrong session**. `ListAgents` showed
`bugs_open/307` twice; a `SendMessage` to the bare name silently picks one, and it picked the
lane that does not own 566. That session replied, corrected me, and established what I could not:
**the 354 session has ENDED**, 566 is still unapplied (`schema_migrations` rows for `566%`: 0),
and the live md5 is still the pre-566 form. Re-verified both here rather than taking it on trust.
Address a duplicate-named session as `name [ref]`.

### The misstep that matters: my verify block passed a mutation test

I wrote the migration's verify block with what I thought was a behavioural control (3e): does any
known FINDING code match the short-retention list? Then I mutated a copy to put
`CONTENT_LINK_REPAIR_DETAIL` straight into the shipped list — the exact defect 3e exists to catch.

**It passed.** Twice over:

1. **3e compared against its own hard-coded `ARRAY[…]`**, not the list that actually goes into the
   `pre_query`. Editing the shipped list changed nothing 3e looked at. *A check on a copy of the
   thing is not a check on the thing.*
2. **Its population could not answer anyway.** It filtered on `CONTENT_LINK_REPAIR_DETAIL`,
   `TRUNCATION_DEGRADED_REVIEW` and `RETRACTION_AUDIT` — and NONE of them has a row older than 30
   days (oldest 2026-07-27, 07-26 and 08-05). The `EXISTS` was false whatever the list said.

Either fault alone makes it vacuous; I shipped both in one block. Rewritten to read `q` — the
shipped text — in both directions, and both mutants are now caught:

| mutant | before | after |
|---|---|---|
| a finding code added to the list | **PASSED** | `567: the short-retention list names CONTENT_LINK_REPAIR_DETAIL, which the registry classes as a FINDING` |
| a plumbing code removed from the list | not tested | `567: the short-retention list is MISSING DISPATCH_UNRESOLVABLE` |

**And my first attempt at the mutation test was itself vacuous** — the `sed` did not match, so I
ran the unmodified file and read its PASS as a result. That is the third vacuous measurement in
one session (after the NULL-column ceiling query). The fix is mechanical and I have adopted it:
**gate the mutation on the substitution having happened** —
`[ "$A" -eq $((B+1)) ] || { echo "MUTATION DID NOT TAKE — refusing"; exit 1; }` — before running
anything. A mutation test that silently tests the original is worse than no mutation test,
because it produces a PASS you then trust.

### What is in 567, and the one design decision worth arguing with

Default is **KEEP**: the list names what expires EARLY (30 days), everything else lives 365. A
code that is new, misspelled or forgotten is RETAINED, so drift can only over-retain, never delete
unread — the opposite failure direction from the rule it replaces. The two RFC_029 resolver codes
stay at 30 days deliberately: they are 10,103 of 45,553 rows, a 365-day clock would cost ~585,000
rows, and their design says frequency is the evidence, not history. That is **no change** for that
lane, but it is a decision taken about their data and it is flagged in the submission's risks.

`severity` was the obvious discriminator and was measured and rejected — findings are written as
error/warning/info, plumbing as error/fatal/warning, three codes emit both.

**Council:** `bae8d694-6095-4adb-b14d-346d31bfb73e`, submitted before applying. Admission
dry-run passed first (free), which is worth doing every time.

---

## 2026-08-22 ~18:50 UTC — the council round did NOT run, and the reason is fleet-wide

**`Council-Submitted: bae8d694-6095-4adb-b14d-346d31bfb73e` will never resolve to a verdict.**
Recorded here so no later reader takes the trailer for a review that happened. The run reached
`complete_invalid` / COMPLETED, which per the runbook means refused before any seat ran — *"absent
note + empty execution_path means nothing was reviewed, not nothing was wrong."*

The step error:

> `step review_editquality failed: … AI endpoint unavailable: provider=anthropic
> model=claude-sonnet-5 … status 400: "You have reached your specified API usage limits. You will
> regain access on 2026-09-01 at 00:00 UTC."`

**This is not this lane's problem and not the council's.** [MEASURED 2026-08-22 18:50 UTC]

| fact | value |
|---|---|
| last SUCCESSFUL llm call, fleet-wide | **2026-08-22 18:15:51 UTC** |
| successes since | **0** |
| failures in the 20 min after | 22 |
| agent types that have hit it | 15 — council-gate, diagnose-agent, page-content-writer, component-creator, webdesign-agent, site-review-agent, experience-planner, landmine-verifier, and 7 more |

**Why this episode is not the usual rate-limit blip, and the check that separates them.** There
have been five usage-limit episodes in the last fortnight (08-10: 7 refusals, 08-14: 28, 08-17: 4,
08-19: 5, 08-21: 3). **On every one, successes CONTINUED after the first refusal** — the fleet kept
working and the day ended normally. Today the first refusal is 18:15:36 and the last success is
18:15:51, fifteen seconds later, and then nothing at all. That is the discriminator: a rolling rate
limit interleaves, a spent cap stops dead. [INFERRED from that shape + the API's own message; the
date 2026-09-01 is the provider's statement, not something I measured.]

**Consequence for this lane, stated plainly:** migration 567 is **APPLIED AND LIVE, UNREVIEWED**.
Not because the gate was skipped — it was submitted first, admission-checked, and dispatched — but
because no seat could run. Its own evidence stands on its verify block (mutation-proven in both
directions) and the end-to-end rolled-back control. **Re-submit after 2026-09-01** and read the
verdict then; if it comes back REVISE, the code is already on the shared branch and live, which is
exactly the situation the `Council-Submitted:` trailer exists to keep honest.

**Consequence for everyone else:** no council gate, no 090 diagnosis loop, no content generation,
no checker agents until the cap clears. Anything queued will fail at its first LLM step. This is an
owner-level (billing/plan) matter — reported in chat 2026-08-22.

---

## 2026-08-22 — ORPHANED WORK IN `sql_for_agents/`, recorded so it cannot vanish quietly

Not mine, not adopted, and deliberately not committed by me. Recorded because uncommitted work on
this tree is not safe (CLAUDE.md: the next `git add -A` from any lane sweeps it into an unrelated
commit and nothing records what it was), and because it currently looks *in flight* when it is not.

| fact | state as of 2026-08-22 19:00 UTC |
|---|---|
| files | `566_database_cleanup_reaps_every_terminal_status.sql` + its `_ROLLBACK` |
| tracked? | **no** — untracked in `git status` |
| applied? | **no** — 0 rows in `schema_migrations` for `566%` |
| complete? | yes — full file, guard + edit + verify, ends `COMMIT;` |
| author | the `bugs_open/354` lane, session **ENDED** (confirmed by the `bugs_open/307 [abdc1e]` lane, which held the other half of that name) |
| what it fixes | arm 3 of `database-cleanup` names `'COMPLETED','FAILED'` literally while arm 4 skips `is_terminal` rows, so a terminal status named by neither is **never deleted**. `CANCELLED` is already in that position: 24 rows, oldest 34 days, against a 24h norm |
| its `before` md5 is now stale | it pins `c26ccf49…`; after 567 the live text is `7f4321d4…`. Its anchor `WHERE status IN ('COMPLETED', 'FAILED')` is untouched and still occurs exactly once — 567 asserts that as a negative control — so swapping its two md5 literals is the whole fix |

**Why I did not adopt it.** I have not read its arm-3 change closely enough to vouch for it, and
committing another lane's finished work under my name is precisely what the commit-per-task rule
exists to prevent. The `bugs_open/307 [abdc1e]` lane declined for the same reason and surfaced the
choice; we independently reached the same answer, which is some evidence it is the right one.
Surfaced to the owner instead — "finished work from an ended session, sitting untracked, fixing a
leak that is losing rows today" is his call.

> **CORRECTED 2026-08-22 ~19:05 UTC — the entry above is WRONG about the outage's likely duration,
> and the error is mine.** I wrote that this episode "is not the usual rate-limit blip" because
> "on every [prior] one, successes CONTINUED after the first refusal" while today's stopped dead.
> **That discriminator was computed on DAILY buckets, and daily buckets cannot see it.** "Last
> success later than first refusal" does not mean the fleet worked around the cap — it means it
> *recovered later the same day*. Hourly, the prior episodes look identical to today:
> 2026-08-10 had **0 successes in the 16:00 AND 17:00 hours** before 100 in the 18:00 hour;
> 2026-08-14 had **0 in the 16:00 hour** before 24 in the 17:00. Today's 18:00 hour (91 ok, 34
> capped) is the same *transition* hour those had. Nothing distinguishes this episode at all.
>
> `LANDMINES.md` ~2051–2059 already records three recurrences of this exact 400 — **each stating a
> reset weeks out, each cleared in 1–3 hours because the owner raised the cap** — and says
> explicitly that *"a stated three-week reset that in fact lasted two hours is exactly the shape
> that gets copied forward into other lanes' docs as a premise."* Which is what I did, here and in
> `README_where_we_are.md`. Caught by the `bugs_open/307 [abdc1e]` lane pointing at the entry;
> the hourly figures above are my own re-measurement, not their word.
>
> **I also did not grep LANDMINES for the symptom** — the SessionStart hook only matches entries
> footprinted on files already dirty, and this one is footprinted on a table and an error string.
> Logged in `WRONG_CALLS.md`.
>
> **What stands:** the outage is real and fleet-wide, the council round genuinely did not run, and
> 567 is genuinely live-unreviewed. **What does not:** any claim about how long it will last. The
> right escalation is *"the cap is hit, please raise it"* — precedent is hours, not weeks — and the
> only proof of a lift is a non-zero success count in the CURRENT hour, never the absence of
> failures.

---

## 2026-08-23 — the cap lifted overnight, `reader_sink` is live, and 567's owed review is submitted

**The cap lifted, and the landmine's precedent held exactly.** [MEASURED 2026-08-23 ~12:50 UTC]
`llm_call_log` by hour: 09:00 = 1 ok / 3 capped, 10:00 = 15 ok / 3 capped, **11:00 = 40 ok / 0
capped**, 12:00 = 22 ok / 0. So the outage ran ~18:15Z → ~10:00Z, **not** to 2026-09-01 as the 400
body stated. That is the fourth recurrence to clear early and the fourth time the stated reset was
the vendor's worst case rather than a forecast. **Verified on the SUCCESS side in the CURRENT hour**,
which is the only proof of a lift — the failures stop appearing either way.

**`reader_sink` is live** (`4156ecca0`), owner-approved from the batch 1 proposal. The point is in
that commit and the registry `_doc`; what belongs here is how it was verified, because the field's
whole purpose is to not be satisfiable by typing:

- **All five existing `consumed` entries were audited against the query their reader actually
  issues, BEFORE the field was filled.** `reconcile_superseded_reviews_action.go:228` →
  `SELECT context FROM agent_error_log …`; `cmd/content-loss-check` `dispositionFamily` →
  `SELECT … FROM agent_error_log WHERE error_code IN ($1,$2,$3)` (three codes);
  `page_build_failure_guard.go:131` → `SELECT count(*) FROM agent_error_log WHERE error_code = $1`.
  Every one reads this table, so the field is honest and the check did not go red.
- **Three mutants on a registry COPY** (`--registry <copy>`, never the shipped file): field removed
  → `consumed-without-reader-sink`, exit 1; sink the reader never mentions →
  `reader-sink-not-in-reader`, exit 1; foreign sink the reader DOES mention → **reported, not
  failed**. So the shipped 0 findings / 0 foreign sinks is a measurement.

**⚠ THE CHECK IS WEAKER THAN IT LOOKS, and this is stated in the code, the `_doc` and here.** It
proves the reader file **mentions** the sink — not that it selects this code from it. A reader
naming several tables can satisfy it wrongly, and `cmd/content-loss-check/main.go` names several.
It is deliberately at the same strength as the existing `reader` check and uses no parsing, because
this mode's founding decision was that it parses nothing so no comment can become load-bearing. It
is still the check that catches the motivating case: 563's prompt template never mentions
`agent_error_log` at all.

**A tree that would not build, and a hook that got it right.** The working tree currently fails
`go test ./cmd/config-key-audit/` because another session's **uncommitted** `platform/livespec/`
rename left a **committed** test (`livedeclarations_test.go:151`) referencing
`livespec.DeferredDeclarations`, which no longer exists. Not mine, not touched. I tested against a
clean `git archive HEAD` with only my two files overlaid. Worth recording: the pre-commit hook
printed *"optional-key parity: NOT CHECKED (the tree does not build — not a parity claim)"* —
which is exactly the right behaviour and the same discipline this lane keeps arguing for. A check
that cannot run must say so rather than pass.

**567's owed review is submitted: `9dc2e6b4-a8fd-476c-8080-ae23567e25c5`.** A FRESH submission, not
a `RESUBMIT_CORR` — the first round (`bae8d694`) produced no verdict at all, so there is nothing to
revise against. Its rationale opens by stating the migration is **already applied and live**, so no
seat reviews it as a plan, and cites the owner ruling of 2026-07-29 that review here is after the
fact by design. If a seat objects the remedy is a follow-up migration; the rollback exists and is
lossy in one direction.

### 567 — APPROVED, round 1 (`9dc2e6b4-a8fd-476c-8080-ae23567e25c5`), and the one objection worth chasing

11 seats, 6 abstained, **0 unreadable**, one advisory objection set. No high severity, nothing
gating. Five objections in total, all `low` except one:

**[medium, guardian] "Plan never addresses whether `database-cleanup`'s pre_query is one of the
objects `platform/livespec` tracks. If it is, this ships with stale livespec silently."**
**CHECKED — it is not.** `livespec.go` declares exactly two `scheduled_task` objects:
`claimed-item-timeout` (:171) and `build-pipeline-trigger` (:184). `database-cleanup` is absent, so
there is nothing to go stale and the objection does not bite. **The seat was right to ask** — it is
the correct question and I had not asked it. *Second-order, and NOT mine:* `database-cleanup` is a
guarded live object edited by migrations 466, 566 and 567 and arguably belongs in that list; that
is the `363`/livespec lane's call, and that session is mid-rename in the package right now.

The four `low` ones, and what I did with each:

1. *"the post-566 md5 `b4deb963…` has no supporting evidence in `grounded_in` — only the pre-566
   hash is measured. If wrong the migration safely refuses rather than corrupts, but the value is
   asserted, not shown."* **Correct, and a fair hit.** I took that value from 566's own header
   rather than measuring it, and never said so. It is unfalsifiable now: 566 is unapplied and its
   author's session has ended, so the text that hash describes has never existed live. It cost
   nothing because the guard fails loud — but "asserted, not shown" is exactly right and it is the
   `[UNMEASURED]` marker rule I would apply to anyone else's number.
2. *"the arm-3 form check hardcodes two exact substrings from 566 … a third, unanticipated form of
   arm 3 would REFUSE the whole migration rather than skip that assertion."* True, disclosed, and
   the trade I chose deliberately: refusing loudly beats applying against a text I cannot recognise.
3. *"~15k/yr growth plus an unindexed hot predicate (`split_part(error_code…)`) on a shared table is
   a blast-radius concern for every consumer of `agent_error_log`, not just this lane."* This is my
   own risk 3 handed back, correctly, as **not committed to a follow-up**. Worth one: the sweep now
   filters on an unindexed expression hourly.
4. *"two migrations racing one row across sessions … worth a tracked follow-up to reconcile rather
   than perpetual either/or guards."* Agreed. The either/or guard is a bridge, not a home.

---

## 2026-08-23 — the ratchet, the index, and a number I stated two ways in one paragraph

**The backlog is capped, as a RATCHET** (`3ed3e4a8c`). Owner said "cap it"; a flat target would
have been red from day one against a backlog of 32, and this checker's own header says why that is
self-defeating. Above the cap is a finding, below it is a nudge naming the number to lower to,
absent is a finding. Live: *"32 unruled, exactly at the cap — the backlog cannot grow"*, and a copy
with the cap at 31 breaches, so the clean reading is a measurement.

**The index is live** (`570`, applied by hand, recorded). Measured at the artefact after applying,
not just in the trial: the strike ladder now plans `Index Scan using idx_error_log_code_time`,
**Buffers: shared read=3** against 8,018 before.

**The council's objection had a wrong premise and a right conclusion, and both halves were worth
measuring.** The seat worried about 567's *sweep*; the sweep never needed an index (already driven
by `idx_error_log_time`, `split_part` only a Filter, 7 buffers). The *readers* needed one badly.
Had I acted on the objection as stated I would have added an index for a query that did not want
one and never looked at the three that did.

> **CORRECTED 2026-08-23 — a small one, mine, and exactly the discipline this lane keeps preaching.**
> `570`'s submission says *"four indexes before this change"* and then **lists five**
> (`agent_error_log_pkey`, `idx_error_log_agent`, `idx_error_log_site`, `idx_error_log_time`,
> `idx_error_log_unresolved`). The table had **five**, or four besides the primary key. I inherited
> "four indexes, none on it" from the original handoff and repeated it without recounting, then
> pasted a list that contradicted it in the same sentence. Nothing turns on it — the negative
> control names all five explicitly and passed — but a count repeated from another doc without
> re-counting is the exact failure the owner's dated-census rule exists for, and it survived into a
> council submission.

**⚠ `/tmp` IS FULL — 16 GB tmpfs at 100%, measured 2026-08-23.** This breaks Go builds with
`link: mapping output file failed: no space left on device`, which reads like a toolchain fault and
is not one. Worse, **the standard HEAD-verification recipe in this lane's own handoff starts
`rm -rf /tmp/h && mkdir /tmp/h`** — so the check every session is told to run to prove HEAD compiles
is the check that now fails first. Workaround: build somewhere off `/tmp` **and** set
`TMPDIR=<that dir>`, because the Go linker's work directory follows `TMPDIR`, not the build path.
Fleet-wide, not this lane's to fix.

---

## 2026-08-23 — the orphaned 566 was ADOPTED and APPLIED; this entry is closed

Closing the loop on the entry above, so it stops reading as an open question. **Not this lane's
work and not a claim on this lane** — recorded here only because this is where the orphan was
written down, and a record that outlives its resolution is how the next reader wastes an hour.

The owner directed a session to pick it up. `566_database_cleanup_reaps_every_terminal_status.sql`
and its `_ROLLBACK` are now **tracked, applied and recorded**: commit `ccc851a42`, applied
2026-08-23 17:46Z, present in `schema_migrations`.

**The entry above was right on every checkable point**, which is worth saying because it was
written by a lane that deliberately did not read the SQL closely:

| the entry's claim | how it held up |
|---|---|
| before-md5 `c26ccf49` stale, live text `7f4321d4` | **correct**, unchanged a day later |
| the anchor is untouched and occurs exactly once | **correct** — asserted against the live text before applying |
| "swapping its two md5 literals is the whole fix" | **correct**, and it was the only edit the SQL needed |
| the leak is real and losing rows | **correct** — still 24 `CANCELLED` rows, oldest 2026-07-19, now 35 days |

One thing the entry could not have known, and the next person computing an md5 guard should:
**the after-md5 must be computed IN THE DATABASE**, with the same `replace()` expression the
migration runs. A locally-hashed copy of the column disagrees — `length()` counts CHARACTERS
while `md5()` hashes BYTES, and this row holds a multi-byte character, so the two sides differ
by 3 bytes. Had the after-md5 been derived locally, the migration's own byte-exact assertion
would have aborted the apply. (This is a known family — `LANDMINES.md` already carries it at
"`length()` on stored HTML is CHARACTERS", so no new entry was added for it. It is also *loud*
when it goes wrong, which is why it is not landmine-shaped.)

**On the judgement not to adopt it.** Two lanes independently declined and surfaced it to the
owner instead, and that was the right call rather than an over-cautious one: the adoption needed
a re-derived guard, a re-measured premise and a blast-radius check, none of which is "just commit
someone's finished file". What the two lanes did that mattered was **write it down** — the file
was untracked, and one `git add -A` from any lane would have swept it into an unrelated commit
with no record of what it was.

---

## 2026-08-23 — PHASE 2 BUILT: the check has a clock, and the blocker was real rather than theoretical

**The blocker, measured before it was designed around.** Phase 2 could not simply be packaged.
Two of `--finding-codes`' arms open the Go file a `consumed` entry names, and no check image in
this estate ships a source tree. Run the compiled binary against the live `DISTINCT error_code`
list with a source-less `--root`:

```
exit=1, 5 findings, e.g.
  [reader-unreadable] CONTENT_DATA_REGRESSION
      reader cmd/content-loss-check/main.go:325 cannot be read: no such file or directory
```

`[MEASURED 2026-08-23]`. The same run with `--no-source`: exit 0. So the naive deploy would have
produced a **red job every morning against a registry with nothing wrong with it** — and this
lane's whole subject is checks that cannot fail. One that always fails is the same defect with
the sign flipped, and a permanently red check is a disabled one.

**Why the arms moved instead of the source.** They grade the registry against Go source, and
**both halves change only by commit** — so a daily re-run in the cluster is arithmetically
incapable of coming out differently than it did at build time. Shipping ~17MB of source to
recompute a constant is not a check. The narrower form (COPY the three files named as readers
today) is worse: a fourth hand-maintained roster, silently wrong the first time a `consumed`
entry names a fourth file.

**The honesty condition, and it is the half that mattered.** Before this, those arms had **no
automatic runner at all** — and that was worth measuring rather than reading, because the entire
remedy rests on it. `[MEASURED 2026-08-23]` at committed HEAD:

```bash
go test ./cmd/config-key-audit/ -run 'BudgetCron' -v -count=1   # what the EXISTING hook runs
#   → exactly 4 tests, all TestBudgetCron*; grep -c ShippedRegistry → 0
go test ./cmd/config-key-audit/ -run 'TestShippedRegistryIsSelfConsistent' -v -count=1
#   → RUNS and PASSES                                        ← the control
```

The control is what makes it evidence: the test's absence is the **filter**, not a missing or
broken test. `scripts/check-finding-code-registry.sh` (new, wired in as hook 5) gives them a
runner, scoped to the registry, the mode's source, and — computed **from the registry at run
time**, never a hand-kept list — any file a live `consumed` entry names as its reader.

### Missteps and things that came out otherwise

- **The build refused, and that was the good direction.** `.dockerignore` excludes `docs/`
  wholesale with a `!` un-ignore line per acks file, and I had not added a third:
  `"/app/docs/.../finding_code_registry.json": not found`. The file's own comment predicts this
  exactly — *"A NEW ack-shipping check needs a line HERE as well as in its dockerfile — without
  one the COPY fails at build time, which is the loud direction and how this line got written;
  the quiet direction would be worse."* **The warning worked.** No new landmine: this one is
  already written, and its author was right that a `COPY` which refuses to build cannot ship an
  inert check. The quiet direction is `component-source-vocabulary-check` v1.0.1326 — deployed
  clean, reported success at every layer, exit 2 on every scheduled run.
- **A control that fired when it should have been silent, and the cause was my restore, not the
  check.** Testing the hook script in a scratch repo, case (C) — the *unmodified* registry —
  reported a failure. The check was right; `git checkout -- <file>` restores from the **index**,
  which still held the broken copy I had just staged. `git checkout HEAD -- <file>` fixed it.
  This is LANDMINES' `git checkout <file>` entry meeting the index, in a throwaway repo where it
  cost nothing — in the shared tree it costs another session's uncommitted work.
- **A scoping test that looked like a failure and was not.** Renaming `DEPLOY_STAMP_REFUSED_ON_SKIP`
  → `..._RENAMED` in its reader file produced no report. Not a scoping bug: the reader arm is a
  **substring** check, and the renamed constant still contains the original string. Removing the
  code entirely (`..._WITHHELD_ON_SKIP`) reported immediately, with no registry edit at all —
  which is the clause that catches a session deleting a query out from under a `consumed` entry.
  The substring limit is pre-existing and stated in the code; worth re-recording because it makes
  a whole class of rename invisible to this arm.
- **My own hook caught the shared tree on its first live run**, and said the right thing:
  *"finding-code registry: NOT CHECKED (the tree does not build — not a registry claim)"*. Another
  session's uncommitted `platform/livespec` rename had the package unbuildable. A check that
  cannot run must say so rather than pass — the same discipline this lane keeps arguing for,
  arriving unprompted.

### Council round 1: REVISE, and the gating objection was right in FORM

`be252395-9d51-4427-b2ae-5f581337b16d`. 9 reviewers, 8 abstained. Gated by **prior_art_librarian**,
HIGH: *"the plan's SourceArmsSkipped message points users to 'scripts/check-finding-code-registry.sh
(pre-commit)' as where the two arms now run, but that script is explicitly 'out of council scope' —
not created by this submission. If it does not already exist, this plan ships a report field that
names a runner that is not yet real, i.e. a forward reference presented as fact."*

**The fact was fine — the script shipped in the SAME commit as the flag (`6b2bfb800`) — and the
objection was still correct.** Reviewers see the plan, not the tree. I had followed the runbook's
"name the companion file in the edit's `rationale`, not its `file`" guidance for an out-of-scope
path, and the effect was a report string that, *from the plan alone*, pointed at nothing. That is
precisely this seat's remit. Round 2 lists the script as edit 6 with its real content, names the
landing commit, and answers the seat's second missing item with the measurement above rather than
another reading of the shell script. **A REVISE round is cheaper than the defect it finds**, and
this one found a real gap between what was true and what was showable.

### Council round 2: REVISE again — and the same defect as round 1, which is the actual finding

Two HIGH objections (`editquality`, `debug_historian`), saying one thing: *"no edit modifies that
CMD, its Dockerfile, or its CronJob manifest to actually pass `--no-source` … the plan builds the
capability to fix the bug but never applies it to the failing job."*

**True of the plan, false of the tree** — the CMD was already committed in `875a56bac`, and
`docker inspect --format '{{json .Config.Cmd}}'` returns it verbatim. **And that is exactly what
round 1 gated on.** After round 1 I fixed the *instance* (listed the script as an edit) and did not
generalise, so the identical shape walked into round 2. I had been leaving `build/` and
`deployments/` out of `edits` because they are out of council scope — but **scope governs what gets
REVIEWED, not what a reviewer is allowed to SEE.** Logged in `WRONG_CALLS.md`; the cheap check is
one question, asked of every load-bearing sentence before submitting: *can a reviewer confirm this
from the edits alone?*

**The reuse seat's medium objection was right and is FIXED rather than answered.**
`check-finding-code-registry.sh` re-implemented `check-optional-key-parity.sh`'s shape — staged
diff, relevance filter, `go test` on the same package, build-failure-vs-real-failure, always exit 0.
*"Two independently-built mechanisms solving overlapping problems in the same package, each
maintained separately, with no unification once both exist."* This lane has no standing to argue:
retiring pairs of hand-maintained things that must stay in sync is most of what `bugs_open/358` is.
`scripts/lib/precommit-gotest.sh` now carries the shared mechanics and both scripts source it
(`83b7e0994`). Not shared, deliberately: each caller's relevance predicate and its guidance text.

> **THE REFACTOR INTRODUCED A REGRESSION AND MY OWN CONTROL CAUGHT IT — worth more than the
> refactor.** Passing the failure headline into the could-not-tell line printed
> `optional-key budget parity: DRIFTED (RFC_022): NOT CHECKED (the tree does not build)` — one line
> asserting a finding and disclaiming one in the same breath. On a tree that often fails to compile
> because of another session's WIP, keeping an undecidable case **legible as undecidable** is the
> single behaviour that helper exists to hold, and merging those two strings quietly destroyed it.
> Caught only because I re-ran all three cases (healthy / real failure / unbuildable) for **both**
> callers — including the pre-existing RFC_022 guard I had not come to change and which every
> session's commits reach. Fixed with a separate neutral `subject`, and the reason is written at the
> parameter so the next author cannot re-merge them. New LANDMINES entry: that helper is now
> fleet-critical and looks like an ordinary library.

`debug_historian`'s medium — *verify at the running pod, not the commit hash* — is accepted and
answered by not making the claim: the CronJob does not exist in the cluster yet
(`kubectl get cronjob finding-code-registry-check` → NotFound), because deploying is the owner's
whole-fleet release. One refinement on the lore it cites: the `strings | grep <symbol>` recipe was
**retired 2026-08-11** after three confidently wrong readings in a day. The binary states its own
provenance now, and this mode prints it (`built from: <sha>`), so "did my change ship?" is
`git merge-base --is-ancestor` against what the artefact says.

### Council round 3: REVISE — the reuse fix was real and my SKETCH was stale, so it read as dead code

Two HIGH (`editquality`, `reuse_agent`): *"edit 8's own rationale says 'BOTH scripts source it', but
edit 4's sketch still contains its own inline `go test`/OUT/RC/grep logic — it never calls
`precommit_run_gotest`. And `check-optional-key-parity.sh` … is not in the edit list at all. As
shown, the shared helper is dead code with no caller."* Plus a third: the kustomizations and the
makefile were again named in prose only.

**All correct.** I had built round 3 by appending edits to round 2's JSON and never regenerated the
sketch for the file I had since refactored. The runbook says exactly this — *"On a resubmit, update
the `sketch` fields — not just the `rationale`. Reviewers judge the sketch"* — and I did what it
says not to. **Third round, same defect, and that repetition is the actual finding**: rounds 1 and 2
I fixed the instance and did not generalise.

**The structural fix, in round 4:** every sketch is now GENERATED from the committed diff
(`git diff e7320fc23 HEAD -- <file>` / `git show HEAD:<file>`), so a sketch cannot disagree with the
tree; every touched file is present — eight as edits, and the four the cap excludes pasted verbatim
at the head of `grounded_in` rather than described. **Scope governs what gets REVIEWED, not what a
reviewer may SEE**, and an out-of-scope path listed as an edit costs nothing.

Worth stating plainly for the next author: the council found **no defect in the code** across three
rounds. Every gating objection was about the plan failing to show what the tree already contained.
That is still worth the rounds — an unshowable claim and a false one are indistinguishable from the
reviewer's chair, which is the same discipline this lane applies to `[MEASURED]` markers.

### Council round 4: APPROVED — 10 seats, 7 abstained, no gating objection

`be252395-9d51-4427-b2ae-5f581337b16d`, verdict 2026-08-23 18:55Z. Four rounds, and **the council
never found a defect in the code** — every gating objection across rounds 1–3 was the plan failing
to show what the tree already contained. That is still worth the rounds: from the reviewer's chair
an unshowable claim and a false one are indistinguishable, which is the same discipline this lane
applies to its own `[MEASURED]` markers.

Five advisory objections. **Checked, not waved past:**

1. **[medium, editquality] "no edit shows the CLI flag registration in `main.go` … leaving the
   causal chain from 'docker CMD passes `--no-source`' to 'auditFindingCodes receives nil'
   unshown."** The premise is conditional and resolves to **no change was needed**:
   `main.go:251-253` already dispatches `emitFindingCodes(os.Args[2:])`, and the flag is parsed
   inside that function (`findingcodes.go:634`). The 12 added lines `git diff` shows in `main.go`
   over the same range belong to **another lane** (`d0930af6f`, bugs 326), not to this change. So
   the chain is complete: CMD → `os.Args` → `emitFindingCodes` → nil `sourceReader` →
   `auditFindingCodes`. **The seat was right to ask** — I asserted "every touched file is present"
   and did not show that `main.go` was untouched, which is the same evidence gap in miniature.
2. **[medium, constitution] the `IMAGE_TAG` bump is fleet-shared plumbing and "the plan gives no
   evidence other targets were checked against this bump."** Correct that the evidence was
   missing, and measuring it found a real hazard worth writing down. `[MEASURED 2026-08-23]`
   **1** image exists at `v1.0.1331` (mine) against **32** overlays at `v1.0.1330` — and
   `deploy-agents` does not deploy the tag an overlay names, it **`sed`s every overlay to
   `$(IMAGE_TAG)`** and applies. So a bare `make deploy-agents` in this window would repoint 33
   services at a tag only one image exists at, and `ImagePullBackOff` reads as RUNNING. The full
   `make release` builds and pushes first, which is exactly why the procedure is what it is. **New
   LANDMINES entry**; the bump is not the mistake, deploying without building is.
3. **[medium, constitution] the shared pre-commit helper "should get architecture-level sign-off
   beyond this single council pass."** Recorded rather than actioned, with the reasoning: under the
   2026-07-29 ruling an RFC is owed when a change alters what a shared mechanism **guarantees**, and
   both guards' behaviour is unchanged — proved by re-running all three cases for both callers,
   including the pre-existing RFC_022 guard. What is real is the blast radius, and that is why it
   already has a LANDMINES entry naming it fleet-critical. Available to the architecture track on
   the owner's call; this lane is not the one to escalate its own change.
4. **[low, constitution] both guards now break together if the helper regresses**, where the risk
   was previously distributed. True, acknowledged, and the deliberate trade for removing the drift
   the reuse seat gated on. Also in the landmine.
5. **[low, tooling_provenance] no `doc_notes` entry records THIS decision for the subject.** Partly
   met: the landmine is synced into `doc_notes` by `landmines-sync.py`, so the fleet-critical-helper
   trap is readable by agents. The `--no-source` shape itself is recorded in the concept register
   (**DBG-075**), which is this estate's sanctioned home for "what exists and is callable" — and
   CLAUDE.md forbids hand-writing landmine rows into `doc_notes` directly.

---

## 2026-08-24 — closing out: 570 APPROVED, 566 landed, batch 1 ratified, and one defect this lane left behind

**570 APPROVED round 1** (`cf916059-39ff-4c20-8e66-7a8d326344e9`, 10 seats, 7 abstained, 0
unreadable). Two `low` objections, and **one of them is factually wrong in a way worth recording
because it is a submission-practice lesson, not a reviewer error:**

> *[debug_historian, low]* "this migration … ships no rollback script."

**It does** — `570_agent_error_log_indexes_its_own_code_ROLLBACK.sql`, written and committed in the
same commit as the forward file. The seat could not see it: `_ROLLBACK` files are **out of council
scope** (`scripts/council-scope.sh`), so I did not list one as an edit, and **an artefact excluded
from scope is invisible to the reviewer, who then reasonably reads its absence as a defect.**
⚠ **So: name the rollback in the RATIONALE even though it cannot be an edit.** Costs one sentence
and prevents a whole objection built on a true absence from the submission and a false absence from
the tree. The other objection (non-`CONCURRENTLY` `CREATE INDEX` taking a SHARE lock) was my own
risk 1 handed back, correctly, and was sub-second in practice.

**566 landed, and the order-independent design paid for itself.** Another session adopted the
orphan (`ccc851a42`, *"354 remainder: adopt and APPLY the orphaned 566"*) using the post-567 md5
`7f4321d4…` this lane handed over, applied it, and recorded it. Arm 3 now reads the vocabulary; my
arm 1 is untouched beside it. **Both migrations edit one 90-line row and neither had to re-derive
anything** — which was the entire point of accepting either known text instead of pinning one.

**Batch 1 is ratified** (by the thread the owner carried 358 to): `25 unruled, exactly at the cap`.
The ratchet behaved as designed — seven codes ruled, cap lowered in step, and the count cannot now
climb back without a finding.

### The defect I left behind, and fixed: `580`

567 replaced arm 1's rule and **left the comment above it**, so until today the live sweep read
`-- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)` — false in both
halves. Fixed by `580` (applied, recorded, `Council-Submitted: a105d04a`).

Worth stating why a *comment* earned a migration: `pre_query` is a **text column**, not source you
read beside the code that contradicts it. An operator inspecting `database-cleanup` sees the
comment and the SQL together with nothing else to check it against. A stale comment in a repo file
is corrected by the diff below it; **a stale comment in a live config row is the only description of
the row there is.** And it is precisely this lane's own subject — a record that misleads its reader.

**A fourth vacuous first-attempt, same shape as the other three.** My first mutant for 580's
behaviour guard added a second `SET` clause and failed on `ERROR: multiple assignments to same
column "pre_query"` — a **syntax** error, which proves nothing about the guard. The valid mutant
nests the rewrite inside the same call — `replace(replace(pre_query, <comment>, <new>), '365 days',
'30 days')` — which is syntactically fine and hides a real behaviour change inside a comment-only
migration. That one **is** caught. The standing check: *a mutant that fails must fail for the
reason under test — read the error, do not read the exit code.*

### And a false alarm of my own, worth the same treatment

My close-out check for "did 567 survive the chassis roll" was
`pre_query LIKE '%365 days%' AND pre_query NOT LIKE '%14 days%'` → it reported **GONE**. It had not
gone: `'14 days'` matched the **stale comment**, not a live rule. I tested a proxy (the string
anywhere in the row) instead of the thing (a live 14-day predicate). Ironically the false alarm is
what found `580`'s real defect — but it could as easily have sent someone re-applying a migration
that was already in place.

### 2026-08-24 — CLOSED. Final state, verified at the artefact

| | |
|---|---|
| registry check | **0 findings**, exit 0 — 41 codes observed, all declared |
| retention parity | **0 disagreements** (registry ↔ live sweep) |
| unruled backlog | **25, exactly at the cap** — cannot grow without a finding |
| `567` retention rule | LIVE · `570` index LIVE · `580` honest comment LIVE |
| council | `9dc2e6b4` **approved** · `cf916059` **approved** · `a105d04a` **approved** |
| HEAD | builds, checked with this lane's own `verify-head-builds.sh` |

**The last thing this lane found was its own broken undo.** Both `567`'s and `580`'s rollbacks used
`regexp_replace(..., 'n')` over multi-line patterns and replaced **nothing** while reporting
`UPDATE 1`. Found only because a council objection claimed `580` had no rollback and I ran it to
prove otherwise. The objection was wrong about the file and **right about the capability**. Fixed,
both now dry-run clean, filed as a `LANDMINES.md` entry footprinted on `scheduled_tasks.pre_query`
and `regexp_replace`.

⚠ **That landmine entry is in HEAD but was carried by another session's commit** (`5f038dd6d`,
the loancalculator lane) — it sat in the shared working tree between my append and my commit and
rode along as a same-file passenger, which is the hazard `CLAUDE.md` documents. Nothing lost; the
remainder (`WRONG_CALLS.md`) went in `fcc19cdef`. Recorded so the trail is accurate rather than
tidy.

**Handed on, not left undone:** phase 2 (the daily CronJob) and the remaining 25 rulings belong to
the thread the owner moved 358 to. The ratchet is what makes leaving them safe — the count can go
down at that lane's pace and cannot climb back.

---

## 2026-08-24 — DEPLOYED, and it caught two on its first live day. One of them exposed a false claim of mine

**The CronJob is live.** The fleet release rolled it out; `kubectl get cronjob
finding-code-registry-check` shows it at `docker.io/aqls/finding-code-registry-check:v1.0.1334` —
the release's `sed` rewrote my overlay's `v1.0.1331` pin, exactly as the makefile does it. Verified
at the artefact, never at the make target.

**Never wait for the schedule.** A manual Job (`kubectl create job --from=cronjob/...`) is how this
was proven, and it is why the finding arrived today rather than tomorrow morning.

### It exited 1 on its first run, and the finding was real

```
[undeclared] LINK_CONTEXT_UNAVAILABLE
```

`[MEASURED 2026-08-24]` 2 rows, both 14:21Z — the CronJob caught it at 16:05Z, **about two hours
after the code's first row.** Not bookkeeping either: both rows are `page-content-writer` hitting a
**query timeout (SQLSTATE 08P01)** while loading linkable pages, `severity=error`, `degraded=true`,
outcome *"writer instructed to emit NO internal links"*. **Two pages were written with no internal
links because of a DB timeout, and nothing reads the row that says so.** That belongs to
`bugfix_092_writer_link_constraints` (quiet 14d), not to this lane, but it is the strongest argument
in the registry for giving a code a real reader.

A second arrived within the hour: `CONTENT_LINK_SUPPRESSED_UNSHIPPED`. Its own doc comment
introduces it as *"a FOURTH code beside CONTENT_LINK_REPAIR_DETAIL, CONTENT_LINK_REPAIR_SKIPPED and
CONTENT_DATA_LINK_AUDIT"* and reasons from their record shape — **which is §3.1 happening live**:
the record shape propagates, and the missing reader propagates with it. Three of those four are
`human-evidence` today.

Both declared `human-evidence` on a measurement (no reader anywhere: Go, SQL, python, shell, live
`agent_definitions` 0, `scheduled_tasks` 0), **by this lane rather than by their authors**, and the
entries say so. The owner may overrule; the measurement is not in doubt. Declaring them `unruled`
would have taken the backlog to 27 against a cap of 25 and left the job red on a different finding.

> **A design observation for the owner, not acted on.** The cap conflates two populations: the
> INHERITED backlog (should shrink — that is the ratchet's point) and NEWLY-BORN codes arriving from
> other lanes at an unpredictable rate (must be ruled promptly, not parked). With the cap sitting
> exactly at the count, any new code is an immediate breach, and the only compliant moves are "rule
> it now" or "raise the cap" — neither available to whoever happens to see the red job. It worked
> out here because the measurement was decisive. It will not always be.

### The finding inside the finding: a file I cited for two days did not exist

`findingcodes.go` has said since 2026-08-22 that a source scan *"is kept as an EARLY WARNING at
commit time (`findingcodes_scan_test.go` in the actions package)"*. **That file did not exist.** It
passed a council round, went into `DBG-075`, and was repeated in this lane's handoff. Nobody ran
`ls` on the path.

What stood in for it was `TestPackageErrorCodeConstantsAreRegistered`, walking a hand-written list
of **eleven** constants — catching only codes someone remembered to add, i.e. the one case that does
not need catching. A **third hand-maintained roster**, inside the change whose register entry boasts
of retiring two. `LINK_CONTEXT_UNAVAILABLE` walked straight past it.

**Now built, and it DISCOVERS rather than lists** (`findingcodes_scan_test.go`, submitted
`2e5f687d-5753-441b-91f3-406c84a98394`): go/ast over every non-test file in the package, collecting
each `ErrorCode:` value and **resolving constant identifiers** — §3.2's own trap ("grep the CONSTANT,
not just the literal") applied to the scanner instead of to a human. Both codes that caught the
package out are `ErrorCode: <const>`, so a literal-only scanner would have been silent on them while
looking like it worked; a test pins exactly that.

**It ratchets.** `[MEASURED]` 13 codes are written by the package and declared nowhere — **all 13
with zero rows in the window**, so the live-table authority is blind to them by construction and
they are precisely the population that blind spot leaves. Failing on them would be red-on-day-one,
which this mode's own design refuses. They sit in `_scan_baseline`; only a code in neither list
fails; the list may only shrink; a baseline entry that becomes declared is reported, never failed.

> **THE NEAR-MISS, and it went the right way.** My first instinct was to delete the eleven-name list
> outright as the redundant roster. Checking per entry first: **ten** are discoverable by the scan,
> and the eleventh — `deployStampRefusedErrorCode` — is **not**, because it is passed
> **positionally** at `page_build_failure_guard.go:111`, which no `ErrorCode:` scan can see.
> Deleting it would have dropped a code's coverage while looking like a tidy-up. It survives,
> narrowed **11 → 1**, and is now itself checked: `TestTheHandListHoldsOnlyWhatTheScanCannotSee`
> fails if an entry becomes discoverable. **The question that saved it was "would the replacement
> actually find THIS one?", per entry — not "is the replacement broadly better?".**

Also corrected in that header: *"that blind spot is harmless BY CONSTRUCTION"*. An unfired code
costs nothing **today**, a weaker claim. The population is thirteen, not zero. Both retractions are
in `WRONG_CALLS.md`.

### Housekeeping, from the CLAUDE.md rule that landed today

I had hand-rolled `git archive HEAD | tar` five times this session — **2.6 GB** of scratch trees,
the exact pattern the new "Checking that committed HEAD still builds" section describes. Reaped to
244 KB. Everything since uses `scripts/verify-head-builds.sh`, which takes repeated `--with <file>`
(needed here: the test and the registry had to be overlaid together, and with only the test the
baseline read as missing).

### The scan's council round: APPROVED — and all four objections say my sketch was truncated over the logic

`2e5f687d-5753-441b-91f3-406c84a98394`, 2026-08-24: **approved**, 10 seats, 7 abstained, no gating
objection. Three seats independently made the same point, and it is a fair hit:

> *[medium, editquality]* "The sketch … is heavily truncated **right at the point where it would show
> the constant-resolution pass, the baseline-comparison/ratchet logic, and the vacuity guard — the
> three mechanisms the rationale cites as the actual fix**. None of that logic is visible to verify
> against the diagnosis; only the header comment and a directory-walk stub are shown."

**Mechanically diagnosed, because it has a mechanical cause rather than a judgement one.** The file
opens with **42 lines of header comment**; my sketch generator truncated the diff at 95 lines in
FILE ORDER. So roughly half the budget went to prose, and what got cut was exactly the code under
review. **The fix is not "be more careful": strip comment lines before truncating, or emit the logic
first.** Recorded in the RUNBOOK as a rule, because this is the FOURTH round in two days lost to
"the plan did not show what the tree contains" — and the first three were omissions, while this one
is a truncation POLICY. Same family, new door.

**The second medium — `_scan_baseline` is load-bearing and not an edit** — is answered by a control
I had already run rather than by argument. `scanBaseline` does `t.Fatalf` when the key is absent
(`findingcodes_scan_test.go:237`), so the seat's "either fails flat on day one or silently no-ops"
cannot happen silently; and I hit that exact message earlier the same day, when I overlaid the test
without the registry. It is shown verbatim in `grounded_in` because the registry lives under `docs/`
and a comment-only sketch is refused client-side — but the seat is right that shown-in-evidence is
weaker than shown-as-edit, and that is the same lesson as above.

**The low one** — `TestTheHandListHoldsOnlyWhatTheScanCannotSee` is referenced by edit 2's comment
but lives in edit 1's truncated tail — is the same cause a third time. No code change owed on any of
the four; what was owed was a better sketch, and that is now a rule rather than an intention.

## 2026-08-24, evening — the fresh roll closes the day-one arc, and the owner's decisions restated

**v1.0.1335 verified at the artefact.** The CronJob's pinned image carries build stamp
`48f55f218…`; `git merge-base --is-ancestor` confirms both the declaration commit (`2f6a36124`)
and the scan/baseline commit (`fb1b8a9be`) are ancestors — so the deployed registry is the
55-code one, not yesterday's. A manual in-cluster Job: **exit 0, 0 findings, 43 observed / 55
declared, retention parity checked, clean row written.** The arc is closed: red on day one →
declared same day → next roll ships the declarations → green, with every hop verified at the
artefact rather than inferred from the tag.

**The owner's decisions in this lane, restated in one place** (they are scattered across four
docs and a proposal, and the chain of effects is worth seeing whole):

1. **Propose-and-ratify (2026-08-22):** the session proposes dispositions with evidence attached;
   he ratifies in batches. A disposition is a decision about what the estate *accepts losing* —
   sessions can measure, only the owner can accept. Effect: batch 1 was one word to ratify and one
   commit to apply.
2. **"Cap it" (the unruled backlog):** implemented as a RATCHET because a flat cap below the live
   backlog is red from day one, and his instruction was about stopping growth, not about
   punishing history. Effect: the backlog cannot grow, and every batch he ratifies lowers the cap
   in the same commit.
3. **Phase 2 first, then batch 1 (2026-08-23):** the clock before more rulings. Effect: the check
   went live a day sooner — and caught two new codes on its first day, which no amount of
   batch-ruling would have seen. The ordering refinement (batch 1 committed *before* the image
   build) is why the deployed image was current on arrival.
4. **Batch 1 ratified in full:** seven codes to human-evidence, recording today's truth rather
   than ambition. 32 → 25.
5. **`reader_sink` approved:** `consumed` must name the table its reader actually reads. Effect:
   `component_validation_rejected` cannot wear a green badge over an unread row.
6. **Standing rulings of his that shaped the build:** opt-in-default-OFF (2026-08-02) is why
   `--no-source` has the shape it has; review-after-the-fact (2026-07-29) is why we committed and
   reviewed without pretending we could hold; whole-fleet release is why the deploy waited for
   him — and is exactly the discipline the IMAGE_TAG landmine documents; the dated-census rule
   (2026-08-22) is why the hand list's staleness was legible as a class; and his council-scope
   widening (2026-08-23) is what made these commits reviewable at all — and finding 098 blind to
   the widened class fell straight out of it.

**A Fable review pass over the shipped code is running** at the owner's request; findings, if
any, land below.

### The Fable review pass DID NOT COMPLETE — and what I checked myself instead

The owner asked for a second pass by Fable. The agent hit the API session limit mid-review (resets
23:50 London) and returned **no findings** — its last message was *"now let me replicate the AST
scan mechanically"*. **Nothing below is Fable's verdict; it is owed, and the relaunch is in the
handoff.** I ran the three items on its checklist a reviewer would most plausibly catch:

1. **Could a real test failure be swallowed as "NOT CHECKED"?** The shared hook helper classified a
   build failure by a bag of words — `cannot find|undefined:|syntax error` — which a test's OWN
   failure message can contain. `[MEASURED]` no test in scope emits those tokens today, so no live
   exposure; **tightened anyway** to the toolchain's unambiguous `FAIL <pkg> [build failed]` /
   `[setup failed]` marker, with the reason at the line. Re-proved all three cases for both callers
   in a scratch tree from `verify-head-builds.sh --with` (`KEEP_TREE=1`, removed after).
2. **Are all 13 `_scan_baseline` codes genuine `agent_error_log` writes, or could the scan be
   collecting some other struct's `ErrorCode:` field?** Six sites showed no `agenterrors` marker
   within 8 lines; on reading, all six are `agenterrors.Finding{…}` literals batched through
   `LogActionFindings`. **All 13 are real.** (And that is a second struct type the scan correctly
   sees — `Finding` as well as `Entry`.)
3. **Prefix collisions between the baseline and the declared set.** `[MEASURED]` exactly one:
   **`UNKNOWN` is a prefix of `UNKNOWN_HANDLER_VERDICT`.** No live `LIKE 'UNKNOWN%'` exists, so it
   is not a query hazard today — but the checker's prefix rule is unconditional (pairwise over every
   declared code), so **the day someone declares `UNKNOWN_HANDLER_VERDICT` the check exits 1 on
   `prefix-collision`**, and they will not know why. Two remedies, neither mine to pick: rename the
   code (the writer is `complete_work_item_verification.go:394`), or scope the rule to family
   prefixes. Recorded in the handoff as known friction rather than fixed by relaxing a rule that
   exists because live `LIKE` families are real.

Not checked here and still owed to Fable: an independent read of the AST walk for edge cases I am
too close to see, and the `--no-source` guard's completeness.

### The Fable pass, second attempt: COMPLETED — five findings, all real, all fixed and proven

Relaunched with a budget-aware brief (read, do not rebuild; what was already checked; ~12 tool
calls). Its verdict, most severe first, and what each became:

1. **HIGH — the scan test is a hand-run tool wearing a commit-time label.** *"`.githooks/pre-commit`
   runs only `check-optional-key-parity.sh` … and `check-finding-code-registry.sh` (`go test
   ./cmd/config-key-audit/`). Neither runs `go test ./platform/orchestration/actions/`."* **Correct,
   and it is yesterday's wrong call one level up**: yesterday the scan test did not exist; today it
   existed and *nothing ran it*. The exact commit shape that produced `LINK_CONTEXT_UNAVAILABLE`
   would still have walked through the hook. **Fixed:** the hook keeps two relevance sets and runs
   two packages; a staged actions file whose STAGED content carries `ErrorCode:` triggers the
   actions package (whole package, no `-run`). Proven in a scratch tree: new code + registry
   untouched → reported by name; actions file without `ErrorCode:` → silent. → `WRONG_CALLS.md`.
2. **MEDIUM — `const b = a` at an `ErrorCode:` site was dropped silently.** Pass 1 recorded only
   string-literal consts; pass 2 had no `default:`. In the missing-a-new-code direction, and not in
   the stated-limits block. **Fixed:** alias chains resolve (bounded fixpoint); every unresolvable
   `ErrorCode:` value is REPORTED by file:line. Proven: an alias probe with a new code now fails
   where the old scanner passed it.
3. **LOW — the const map was keyed by bare identifier across every scope.** `ast.Inspect` walked
   function bodies, so a `var code = "X"` inside any function would have misattributed the three
   real `ErrorCode: code` sites. **Fixed:** `f.Decls`, file-scope `const` only. Proven: a local-var
   probe passes (reported unresolved) where the old scanner would have failed falsely.
4. **LOW — `TestTheHandListHoldsOnlyWhatTheScanCannotSee`'s "control" was a `t.Log`.** Under `-run`
   it passed vacuously against a broken scanner — the `-run`-as-roster shape my own scripts warn
   against. **Fixed:** `scanOrFatal`, one entry point for all three tests.
5. **LOW — the reader-check comment overstated itself** ("a constant declared to it in the same
   file"); the check is `strings.Contains` with no resolution. **Fixed:** the comment now says
   exactly what passes and why.

Plus a nit (pipe under `pipefail` in the fleet-critical helper → here-string), taken.

**What Fable verified HOLDS** (so nobody re-spends on it): vacuity thresholds vs the package's real
size (354 files, ~36 codes); ratchet state consistent; key normalisation; the hand list's one entry
genuinely positional; `--no-source` nil-`src` reachability; every script path returns 0; all five
`consumed` readers carry both the code and the sink; 9/9 `human-evidence` `why` fields name the
window.

**The unresolved report, on the real package** (`-v`): exactly the four sites Fable named —
`component_write_guard.go:501`, `log_action_error.go:252`, `v3_site_actions.go:4197`, `:4261` — all
local variables, i.e. the stated runtime blind spot, now *visible* as such rather than silent.

*(Housekeeping: the `WRONG_CALLS.md` addendum for the repeat wrong call was swept into the `364`
lane's commit `0fe414745` as a same-file passenger before my pathspec commit reached it — the
documented shared-tree behaviour. It is at HEAD, nothing is lost, forward-only holds. Code fixes
for Fable's findings: `bce49226a`, council `4d5c1523`.)*
