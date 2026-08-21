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
