# NOTES — the work-item terminal-write contract (bugs_open/307)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-20 (session start) — research, and what it changed about the plan

Picked up `bugs_open/307` on the owner's instruction. `scripts/who-owns.py 307` says
OWNED/active (`staged_component_build`), but that lane's own §7 ends *"whoever builds it
should treat this section as the spec's skeleton"* and its handoff lists 307 as open work —
so this is contribution into a shared account, not a competing fix. Re-checked at start of
implementation: **0** council runs in the last 12h mention the seam.

**Bug re-validated before building** (it is two days old and this tree moves):
- Code unchanged: `FailWorkItemAction`'s three branches all still `WHERE id = $1`;
  `UpdateWorkItemStatusAction` still increments `attempt_count` on both arms with no ladder
  and no guard. `git log` on both files shows no fix landed.
- Still bleeding: `failed` rows at `attempt_count=1` with `handled_by` NULL continue daily
  (72 on 08-18, 2 on 08-19).

### The 090 diagnosis run — filed, and it FAILED, which is worth recording precisely

Fired `090` (intake corr `87f3b06f-91d6-4220-9533-796bce5cb196`, run corr
`103ad179-6674-4f0f-9d7d-79a5372cfdc9`). It assembled **four** evidence bundles
(63k/37k/92k/62k chars) whose contents agree with our reading, and then its **`verdict`
step died**:

```
step verdict failed: … response truncated: stop_reason=max_tokens
(output_tokens=32000 reached the configured cap, 0 chars recovered)
```

The intake item went terminally `failed` at `attempt_count=1` — correctly, because the
diagnose lane runs `max_attempts=1` by design. **So there is no gradable verdict.**
[MEASURED — the item row and the four `diagnosis_artifacts` rows]

Two things follow, and I am recording both rather than the convenient one:
1. **My symptom statement packed THREE mechanisms into one filing** (no-delay ladder,
   unladdered writer, LLM-only classifier) when the trigger's own guidance is *"one
   coherent bug per run"*. That is the most likely contributor to a 32k-token verdict.
   The next filing on this seam should carry one mechanism.
2. The bug was filed under the 2026-07-31 ruling's **named** escape hatch (first-hand
   verification, declared in the bug's own header). We proceed on that, plus this session's
   independent re-verification in code, in the live DB, and across three sweeps — not on a
   verdict we do not have. Owner asked; owner answered: proceed, record the failed run.

### What the research actually changed (i.e. what I would have got wrong)

- **I would have used `blocked` as the cooldown state.** The bug file suggests it. It is
  unusable: `feasibility-recheck` (enabled, every 600s) releases *every* `blocked` row whose
  handler exists — no timestamp condition — **and clears `error`**, destroying the reason.
  That is why the table holds **0** `blocked` rows and has held 0 for its whole history:
  not unused, continuously drained. [MEASURED]
- **I would have written backoff literals into Go.** `reaper_policies` already exists
  (migration 335, register SCH-024) and RFC_018 explicitly names `site_work_items` as the
  second consumer it is waiting for. Literals would have been the third hand-rolled copy of
  a mechanism the architecture seat already objected to once.
- **I would have keyed the burst detector on `site_id`.** In `agent_error_log`, `site_id` is
  **89.8% NULL** over 7 days (and 2025 of 2084 rows NULL *inside the outage window itself*),
  while `domain` is **0%** NULL. A "≥2 distinct sites" rule would have failed to fire during
  the exact outage it was written for. [MEASURED]
- **I would have swapped `isAIUnavailable` for the shared `RetryDisposition`.** Their needle
  lists disagree in both directions; swapping silently drops EOF/401/402/credit/api-key
  coverage. Layering keeps both.
- **I assumed the outage was the main damage.** It is not. The unladdered path costs
  **141 items in 14 days** (52% of all failures, 71% in the last 48h), 139 at
  `attempt_count=1` of 3 — in fair weather. The incident was the alarm, not the fire.

### One prediction I am writing down BEFORE it can be checked

The guard half is currently **unprovable from data**: both writers write `failed`, so no
query can separate "the handler said failed and it stood" from "the loop overwrote it".
9 `wont_fix` rows exist as of 2026-08-19. **Prediction: once the guard ships, the count of
`wont_fix`/`needs_human_review` rows that later flip to `failed` should be 0, and before it
ships that count is not measurable at all.** If someone later reports the guard "did
nothing", that is the expected reading of a prophylactic — check the skip log line, not the
row count.

## 2026-08-20 (implementation) — appended as it happens

### Build, and the four things that changed shape while building

**1. `build-dispatch-watchdog` does not exist.** The plan named three SQL read
sites, on the strength of `docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql`
and a research sweep that quoted its pre_query verbatim. It is not live:

```
$ git status --short docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql
?? docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql     -- UNTRACKED
$ git log --oneline -1 -- .../214_build_dispatch_watchdog.sql          -- (no commits)
SELECT name FROM scheduled_tasks WHERE name ILIKE '%watchdog%';        -- 0 rows
```
Some session wrote it, never committed it, never applied it. **A file in
`sql_for_agents/` is not a live task — the live inventory is the table**, and a
research sweep that reads the repo will keep reporting this one as real. So 506
patches TWO read sites, not three, and the "false BUILD_DISPATCH_STALLED" risk
the plan listed does not exist. Recorded in 506's own header too, where the next
reader of that file will be standing. [MEASURED 2026-08-20]

> **CROSS-LINK added 2026-08-22 — the same absent task, found independently from the other side.**
> The `bugs_open/358` lane reached `214` from the finding-code registry rather than from the
> dispatch reads, and their half is the more useful one: `BUILD_DISPATCH_STALLED` is registered
> as a code **with a closed automated loop**, and the loop is real *in the file* but inert,
> because both halves live inside the `pre_query` of a task that was never created. So the
> code's **zero row count reads as "quiet" and actually means "absent"** — a registered
> mechanism and a silent one are indistinguishable from that direction. Their entry:
> `docs/agent_docs/docs026_concept_register/register/debugging.md:613`.
> Re-verified here 2026-08-22: still untracked, **0** rows in `schema_migrations` for `214%`,
> **0** `scheduled_tasks` rows named `build-dispatch-watchdog`.
> Linked so the two findings stop being re-derived separately — I proved the task does not
> exist, they proved something *depends on it existing*.

**2. The live inventory of tasks that UPDATE `site_work_items.status` is five**,
and it matches what the research said: `claimed-item-timeout` (120s, and it runs
its OWN copy of the ladder), `detected-item-promoter` (900s), `feasibility-recheck`
(600s), `held-pair-canary-escalation` (86400s), `stale-work-item-reaper` (3600s).
None reads `retry_after`. The claimed-item-timeout duplicate ladder is left alone
in this change and named in the submission as the remaining divergence.

**3. `handled_by` nearly got broken by my own fix.** The helper takes an
`agentType` and writes it to `handled_by`; `update_work_item_status` has no agent
identity to pass, so it would have written `''`. That column being **NULL** is the
documented tell that separates the two writers — bug 307 §2.2 attributes its 12
attempt-1 deaths with it, and other lanes' censuses use it. Writing `''` would have
populated the column without identifying anyone and silently invalidated all of
them mid-stream. Caught before commit; the SQL now reads
`handled_by = COALESCE(NULLIF($3, ''), handled_by)`, which also keeps every bind
parameter referenced (the Go-side branch I tried first left `$3` unused and would
have failed at runtime with "bind message supplies 5 parameters").

**4. Two of my own test fixtures were wrong, and the code was right.** Both
failures were the burst probe not firing: the detector refuses to query on an
empty signature, because an empty one would collapse every blank-message failure
in the fleet into one group that matches itself. My fixtures assumed a probe on
cases with no readable `__step_error.message`. Fixed in the tests, and the reason
is now stated per case rather than inferred.

**Also:** a Go raw-string literal cannot contain a backtick, and I put
`` `handled_by IS NULL` `` inside a SQL comment inside one. It compiled as a
syntax error 20 lines later. And `nullableJSON` already existed in this package
returning `[]byte` (which lib/pq sends as bytea and will not cast to jsonb), so
the text form is a separate `jsonbTextOrNil` rather than a change to a helper
this work has no business touching.

### Mutation proof — the five behaviours that must not silently rot

Each mutation applied to a clean copy at HEAD, tests run, code restored:

| mutation | caught by |
|---|---|
| drop `wont_fix` from the guard list | `GuardListIsNotTheCompletionPathList`, `TheGuardIsActuallyInTheStatement` |
| never build the guard clause | `TheGuardIsActuallyInTheStatement` |
| never stamp `retry_after` | `RetryAfterIsActuallyStamped` |
| drop the agent-type leg of the burst conjunction | `BurstThresholdsRequireTheConjunction/volume_alone,_one_agent_type` |
| let `max_attempts=1` items be released | `AOneShotLaneIsNeverReleased` |

The first two matter most: a mock decides how many rows come back, so a test that
only asserts "skipped" passes against a statement with no guard at all. The SQL
text has to be read.

### Migration validation

Both files, then both rollbacks, executed against the live schema inside one
transaction and rolled back: `ALTER TABLE / COMMENT / INSERT 0 1 / DO / UPDATE 1 /
UPDATE 1 / DO / … / ROLLBACK`. Each of 506's two UPDATEs matches exactly one row,
and both DO-blocks pass — including 506's assertion that the selector still
carries every pre-existing dispatchability clause, so a hand-typed near-copy
cannot silently drop the `depends_on` or claim guards.

## 2026-08-20 (evening) — council round 1: REVISE, and it found a real defect in my code

Verdict `revise`, decided by a gating objection from `editquality`. Two of the objections were
real defects, both mine, and both are in `WRONG_CALLS.md` as separate entries.

**The gating one, and it is the worse of the two.** I gave `update_work_item_status`'s **`complete`**
arm the failure-path guard list. That list deliberately omits `failed` and `unresolved`, because
moving a row through them is what a retry *is*. On the completion path that omission means a
`complete` write can silently overwrite a row that already failed or was given up — **the exact
defect class this change exists to close, reintroduced by the change, one arm over.** My rationale
called it *"the same guard its two siblings already have"*; it was false in precisely the two
entries that mattered. The seat compared my list against the sibling I had cited. I never did.

Fixed with a separate `workItemCompletionGuardStatuses`, placed *adjacent* to the failure list so
the difference is visible, plus two tests: the exact delta between the lists, and a source read of
the call site (the constants can be right while the interpolation is wrong — which is what
happened). Mutation-proved three ways.

**The measurement one.** `guardian` noticed that my "141 of 270 failed rows in 14 days" came from
`site_work_items` alone, which the archiver drains after ~7 days. Archive-inclusive: **401 of 558,
72%.** I had understated my own case threefold — and this lane's own RUNBOOK, written the same
session, says to `UNION ALL` the archive for exactly this. Writing a check down is not running it.

**Three seats independently refused to accept "named as a residual" as remediation** for
`claimed-item-timeout`'s fifth copy of the ladder. They were right that prose is not a ticket; it
is now `bugs_open/341`. The line worth keeping is `bug_historian`'s: *"the platform's own history
says the untreated sibling is where the next incident originates, not a footnote."*

**`bug_historian` also caught a false claim I had made twice** — "reapable at 48h". Clearing
`claimed_at` makes a row *eligible* for the stale reaper, but every write bumps `updated_at` and
the reaper keys on that, so each failure restarts the clock. Correct: *48h after the last write*.
Corrected in the register and the bug file.

**`debug_historian` (high) on migration 506's blind `jsonb_set`.** Right as discipline, and it did
not bite — verified mine was the only write to that row and no other lane's migration touched that
agent in the window. But the objection is about the *next* run, not this one: a ROLLBACK is what
somebody executes months later under pressure. The sidecar now gates on a pre-state md5 of both
read sites and aborts naming expected/found. Mutation-tested by corrupting the expected md5.

**`architecture` (medium, explicitly non-gating)** asked for a recorded entry — `RFC_043`. Its
second sentence is the one I had got wrong: *"the 08-18 owner ruling authorises the BEHAVIOUR; it
doesn't exempt the MECHANISM implementing it from architectural review — those are separate
questions."* I had treated the ruling as settling both.

**What the seats could not check, and what I did about it.** `reaper_policies`, `scheduled_tasks`
and `schema_migrations` are outside the council's 11-table allowlist, so `prior_art_librarian` and
`guardian` both correctly flagged the `reaper_policies` existence claim as unverifiable-from-here.
Evidence went into a dated `doc_notes` row (the documented remedy) rather than widening the
allowlist to win my own round.

**Round 2 submitted** on the same trail correlation (`RESUBMIT_CORR=4cdec68b…`), sketches updated
rather than only the rationale, and the full round-1 evidence carried forward — each round is
judged standalone.

**The pattern in both of my defects.** The gating one was in the *by-the-way* edit — the one-line
guard I added while already in the file, which had no test among the fifteen. The measurement one
was a check I had written down that morning and did not run. Neither was a gap in knowledge. Both
were the parts I wasn't looking at because I was confident about the parts I was.

## 2026-08-21 — the roll happened, and natural traffic beat the canary to the transient arm

*(Fresh session continuing the lane post-approval. Everything below measured this morning.)*

**The roll.** v1.0.1320, both replicas started 2026-08-20 16:09Z, stamp `a255551e0` read from
`/proc/1/exe` on both pods with the positive control (9911 hits of `gqls/agentchassis`);
superseded by v1.0.1321 at 19:51Z (stamp `0483e7f4e`, control 9926). `git merge-base
--is-ancestor` true for all three commits (`069015add`, `5e1a0ac1e`, `29b32d0d8`) against BOTH
stamps — the round-1 completion-guard fix was aboard from the first live minute; no window ran
the wrong list.

**First natural proof — the transient arm end to end.** Two `page_rerender` items
(`5f52413b…`, `6de492b9…`) hit a real Kafka topic-creation failure ~18:34Z/~18:54Z on the 20th:
error prefixed `transient (ai_unavailable)`, `attempt_count=0` (release did NOT consume),
`retry_after` stamped (18:34:00Z / 18:54:33Z), re-claimed only AFTER the stamp
(18:34:25Z / 18:56:58Z), both then **complete**. The owner's ruling, observed un-induced.

**The demand-controlled zero.** 288 `agent_error_log` events since 16:09Z, and **zero** terminal
`failed` work-item writes (both sides of the 341 carve-out zero) against the pre-fix
archive-inclusive ~29/day with 72% at attempt 1. 18 hours of that is signal.

**Read carefully, not over-read:** one `needs_imagery` row completed at `attempt_count=1`, no
`retry_after` — that is the parked/complete-write increment residual (§8 of the bug file), not
ladder evidence. The 4 `wont_fix` rows touched since the roll are all NEW OWNED_PAGE_GUARD
refusals; no decision status observed overwritten. Current pods (19:51Z) log no
`work item failure ladder` lines yet — the two releases predate the restart, their lines died
with the old pods; log evidence for future events only.

**Still owed live (the canary's three arms, owner-authorised this morning):** attempt-ladder
scaling on a non-transient failure, honest terminal at `max_attempts`, and the guard skip.
Owner also ruled this morning: **close on fair-weather proof**, converting §5's outage arm to a
dated standing watch with a reopen trigger (the `bugs_closed/006` shape).

## 2026-08-21 — LIVE and proven; and I got the roll boundary wrong TWICE

**The authoritative account of the roll and the first natural proof is `bugs_open/307` §9**,
written by the session that picked the lane up post-approval. It is better evidence than what I
assembled independently, in three specific ways, and I am recording the comparison rather than a
second version of the same thing:

- it found the roll at **v1.0.1320, 16:09Z on 08-20** (stamp `a255551e0`), superseded that evening
  by v1.0.1321 (`0483e7f4e`) — so the contract has been live **continuously since 16:09Z**,
  including the round-1 completion-guard fix, i.e. **no live window ever ran the wrong guard list**;
- it verified by **binary probe on both replicas with a positive control**, where I used
  `service_binary_capabilities` + `merge-base` (both valid; theirs has no shelf life either and
  does not depend on the registry being written);
- it checked the thing I did not: **the re-claim timestamps**, proving the claim path *honoured*
  the stamp (18:34:25Z re-claim against an 18:34:00Z stamp) rather than merely that a stamp existed.
  Mine proved the stamp was written; theirs proved it was obeyed.

### My own misstep, twice over, on the same fact

I took the roll boundary from `kubectl get pods … startTime` = **19:51Z**, measured against it, and
got the self-contradicting result *"0 failures since roll"* alongside *"2 rows carry a retry_after
stamp"*. Both cannot be true. I then corrected to **18:30Z** by reading the stamped rows' own
timestamps — and that was **also wrong**, because those two rows are the earliest rows I happened to
find, not the earliest moment the code was live. The answer was 16:09Z and it came from the
artefact, not from arithmetic on symptoms.

**A pod's `startTime` is not the roll time.** It is when *that pod* last started; this fleet's
ephemeral per-job pods restart constantly, and a Deployment replica can restart hours after the
image changed. **The roll time is a property of the IMAGE and its stamp, not of any pod's clock** —
and my second guess repeated the first mistake's shape: I inferred a cause boundary from the
earliest effect I had found.

What saved it both times was a contradiction *inside my own output*. That is worth more than the
correction: a query whose two halves disagree is a gift, and the instinct to reconcile rather than
pick the convenient half is the whole of it.

### What I did measure that stands, with its demand control

The rate comparison, archive-inclusive and excluding `Claim timed out%` (the 341 writer's own path):
**18.8 early deaths per 16 h pre-roll** (14-day average) against **0 observed** post-roll. §9's
figure is the same finding at a different scale — ~29/day pre-fix against 0 terminal `failed`
writes in 18 h, with **288** `agent_error_log` events on the demand side. Either way the zero is
interpretable rather than merely quiet, which is the only reason to publish it.

### The `handled_by` reading that will confuse the next person

Both released rows show `handled_by = build-dispatch-loop`, and the release path sets it **NULL**.
That is the *completion* writing it afterwards: release (NULL) → re-claim → `complete_work_item`
(agentType). Not a contradiction, and not evidence the release did not happen — the evidence for
that is `attempt_count = 0` with `spec.transient_releases = 1`.

### What remains, and it is now specified rather than aspirational

Three arms nothing natural has demanded: the attempt-consuming ladder's **scaling** (30m → 60m),
the **honest terminal** `failed` at 3 of 3 (§5(c) live), and the **guard skip**. The owner has
authorised a synthetic canary for exactly these; the choreography, its race condition and its
teardown are in this lane's RUNBOOK ("The close canary"). Not yet run — verified 0 rows matching
`canary_307%` at 10:3xZ.

## 2026-08-21, midday — the canary's verdict: ladder right, and two defects around it. Close BLOCKED.

Full record in the bug file §9b; this is the working log of how it unfolded.

- Canary f4f15466 inserted 10:31:34Z (empty dispatchable queue → picked on the next 60s tick).
  Attempt 1: ladder perfect (`triaged`/1/+30m, claim cleared, error preserved) — **then
  `complete` two seconds later.** That is `bugs_open/344`: mark_complete runs on every
  returned saga, the complete_error child reports plain "complete", and `triaged` is in no
  completion guard. A NATURAL case (0c65f9fa, mortgagecalculator) did the same 42s earlier.
- Attempt 2: scaling proven (+60m), false-greened again.
- Attempt 3: **the terminal write itself errored — SQLSTATE 42P18.** backoff=0 on the terminal
  attempt collapses the retry_after fragment, $4 leaves the text but not the bind list. Both
  writers share it (a natural fail_work_item hit at 10:41:29Z). Terminal `failed` has been
  UNREACHABLE through the Go ladder since the roll; items cycle at max−1 via the sweep
  (95fe67da is doing it now). Fixed `bc80dde4a` (statements+args assembled together, class
  pinned by a placeholder-audit test, mutation-proven), Council-Submitted `df0748bf`, inert
  until the next roll. WRONG_CALLS row: none of the fifteen tests could 42P18, and none asked
  what the NEXT writer does to the row.
- Canary DELETEd, verified 0. The needs_diagnosis filed in the window is the natural render
  failure's, not ours. Guard-skip arm not demanded (overtaken by events) — owed to the
  post-roll re-run.
- Peer coordination: the original 307 session made contact mid-canary; division agreed (they
  take 341 candidate 2, we hold 307 + the 42P18 fix; 344 filed unowned). CONTRIB sent to the
  mcalc lane, whose "nothing flagged the item" second-defect account is contradicted by their
  own row's `attempt_count=1` + `retry_after` stamp — their 090 run 0b498cf8 will adjudicate.

**New close bar** (§9b): 42P18 fix live at the artefact on the next roll → 344 fixed or
owner-dispositioned → the same canary recipe re-run clean, all three arms plus the guard flip →
then close with §5's outage arm as the standing watch, per the owner's morning ruling.

## 2026-08-21, evening — CLOSED: one canary, every arm, on v1.0.1322

The 16:54Z roll carried BOTH repairs (my 42P18 fix and this lane's other session's 344
candidate 1 — `0f80f5ea1`, committed inside the build window; my earlier "not yet committed"
read was wrong, a head-truncated multi-path `git log` hid it and the canary's changed behaviour
exposed it — the cheap check was `git log -1 -- <file>`, single path).

Canary c192a2b2, full pass (bug file §9c has the table): attempt 1 `triaged`/1/+30m AND
`completed_at` NULL (the 344 fix working — yesterday's binary stamped it complete in 2s);
guard flip to `wont_fix` mid-claim → the failure write SKIPPED (row untouched, job pod logged
the skip line verbatim); attempt 2 `triaged`/2/+60m, survived completion; attempt 3 terminal
`failed` 3-of-3, `retry_after` NULL, completion refused. Torn down, 0 rows by key. Natural
census: 0 false greens since the roll; the §9b damage row was re-driven past the defect by its
own lane.

Council df0748bf: APPROVED round 1, 2 advisories, both answered with a change or a measurement
— the 42703-recovery test (`54032b2dd`) and the 13-steps/11-agents caller census (nested walk,
recorded in §9c). Coverage: the two pre-verdict commits carried Council-Submitted and resolve
automatically; the test commit carries Council-Reviewed on a read verdict.

**307 moved to bugs_closed** per the fixed-AND-live bar + the owner's morning ruling (outage
arm → standing watch with reopen trigger, written into the closed file). What this lane still
owes lives in 344 (sweep SQL half + close) and 341/524_HOLD — the other session's, by the
agreed division.

## 2026-08-21 (evening) — the roll landed, 307 closed, and my 344 round came back REVISE

**Fleet: one stamp.** `bac189921`, 59 pods, `service_binary_capabilities` — and briefly a MIXED
fleet (12 pods new / 275 old) while it rolled, which is worth remembering: a stamp query taken
mid-roll answers "what is running" with two rows, and picking either one is wrong. Both of my
fixes and the peer's 42P18 fix are ancestors of the current stamp; the negative control (later
HEAD) is not.

**307 is CLOSED** (`bugs_closed/`, by the peer session, on the owner's fair-weather ruling with
§5's outage arm as a standing watch). Its §9c holds the evidence table — cite that rather than
re-proving it.

**344's Go half is live and canary-proven**: attempts 1 and 2 re-triaged with cooldown stamps and
SURVIVED the loop's completion call (`completed_at` NULL where the previous binary stamped
`complete` in two seconds), the guard-skip arm fired with my log line captured verbatim from an
ephemeral per-job pod, and the natural false-green census reads **0** since the roll.

### The council came back REVISE on 344, and the gating objection was about a LIVE path

`guardian`, high: `ClaimWorkItemAction` is the fleet's central claim gate, and I had written that
the rendered SQL was *"byte-identical"* to the hand-typed clause it replaced **without ever
asserting it**. By the time the verdict arrived that code was already serving the whole fleet. It
was fine — measured 28 claims / 26 completions / 0 false greens since the roll — but "it turned out
fine" is not the same as "I checked", and the seat was objecting to the second thing.

Three other seats were right too:
- **The census I kept quoting had three different values.** My submission said "five Go sites", the
  test comment said "three consumers", and the truth is **four call sites across three files**.
  Two seats independently flagged the inconsistency, and both said the same thing: an unclosed
  census is where a missed duplicate hides. **This was my third undercount of the same population
  in two days.**
- **The skip had no durable record** — a log line only, and pod logs rotate in ~90s
  (`bugs_closed/034`'s shape). Now stamped on the row as `result.completion_skipped`.
- **The classification was a second, unsynchronised read.** Now a `CASE` inside the same `UPDATE`
  that records it, so the window is gone by construction rather than documented.

### The finding that matters more than any of them

Mutation-testing the durable-record fix returned **`[no tests to run]`**. The change I had just
made to answer a review objection **had no assertion behind it at all**. That is the 42P18 failure
from two days ago wearing different clothes: a fix believed because it was written, not because
anything would notice if it stopped working.

**The habit that falls out of it, and it is cheap: after writing code to answer a review, delete
that code and run the tests.** If nothing goes red, you have not answered the objection — you have
only described an answer. It has now caught two of my own changes in three days.

### 341/524 released

The HOLD's condition was "344 candidate 1 **LIVE**, not committed", and the distinction earned its
keep exactly as written: the fix was committed in the morning, the owner deferred the roll to the
evening, and releasing on the commit would have stamped cooldowns for a day while `mark_complete`
was still unguarded. Applied and recorded after the roll; the sweep now stamps `retry_after` from
`reaper_policies`.

**And the sweep needed no completion predicate after all** — §5b (which I recorded from the peer)
was wrong for the reason §2b already gave: all three arms select `status='claimed'`, which a
re-triaged row is not. Measured 0/0. **I wrote the correction and then failed to apply it to the
next claim I accepted**, which is the more transferable half of that mistake.

### 524's live status, stated precisely — and the misattribution I nearly published

First read after applying `524`: *"rows stamped by the sweep so far: 1"*. **Wrong, and wrong in the
flattering direction.** My query was `retry_after IS NOT NULL AND error LIKE 'Claim timed out%'` —
which asks "does this row have a stamp AND has the sweep ever touched it", not "did the sweep write
this stamp". The row it found carries `retry_after 13:51:37`, **five hours before `524` was
applied**: the stamp came from the **Go ladder** earlier in that item's life, and the sweep's later
terminal write simply left it alone, because pre-`524` the sweep did not touch the column at all.
(It is also terminal at 3/3, where `524` writes NULL — a second tell I had in front of me.)

**Correct attribution needs the write TIME, not the column's presence:** `error LIKE 'Claim timed
out%' AND updated_at > <when 524 applied>` → **0**.

> **CORRECTED minutes later — the boundary I used was in the FUTURE.** I wrote that query with a
> hard-coded `'2026-08-21 19:00:00+00'`, estimating when I had applied `524`. The real
> `schema_migrations.applied_at` is **18:44:22Z**, and the clock at the time of the query was
> **18:50Z** — so I was counting writes after a moment that had not yet happened, and the 0 was
> **guaranteed regardless of what the sweep had done.** Re-run against the real boundary: still
> **0**, with **0** claimed rows older than the 40-minute threshold and 3 sweep ticks elapsed. The
> conclusion is unchanged and the method was invalid, which are different things.
>
> This is the discipline rule's own case — *"a `[MEASURED]` figure is only evidence if the
> measurement could have come out otherwise"* — and it is the **fourth** measurement error of one
> family in this session: a filter that could not match its target, mutations that never applied,
> an attribution by column-presence rather than write-time, and now a window that had not opened.
> **Never hard-code a boundary you can read.** `applied_at` was one join away, and a boundary taken
> from the system cannot be in the future by accident. With the demand control beside it: **0 claimed
rows exist at all**, let alone any older than the sweep's 40-minute threshold, and the task last
fired at 18:46Z with nothing to do.

So: **`524` is applied and structurally verified** (the `pre_query` carries the stamp, reads
`reaper_policies`, both auto-complete arms intact, and the `EXPLAIN` gate parsed it) **but NOT yet
proven on live traffic** — because it has had no work, not because it failed. That zero is
"untested", and the demand control is what licenses saying so rather than guessing.

**Third attribution error of the same family in three days** (after the `agent_error_log` filter
that could not match its own target, and the mutations that never applied): each time I asked a
question whose answer I could not distinguish from the one I wanted. The habit that keeps catching
them is the same one — *before believing a count, name the row it would have to be about, and check
that row is the row you mean.*
