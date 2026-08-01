# NOTES — bugs_open/162 (append-only, newest at the bottom)

## 2026-07-31 ~22:30 — picking it up

Chose 162 off `bugs_open/` after checking it was unowned three ways, because one check
is not enough here: `scripts/who-owns.py 162` (one commit, the filing, no owning
workstream), a grep of all 33 transcripts modified in the last 3 hours for
`bugs_open/162`, and a grep of the same for the CODE symbols
(`diagnose_persist_fix_plan`, `273_fix_proposer`, `plan_validation_refusal`). The symbol
grep is the one that matters — `who-owns.py` reads commits and is blind to a session
mid-fix. It found session `31f37a59` with 139 hits, which is the `099` lane; last
written 19:25, i.e. idle for 3 hours. The bug file itself says the quiet part out loud:
"**unowned by this lane on purpose** — `fix-proposer` belongs to the fix-loop lane."

## Validity check — still valid, and one thing the file gets slightly wrong

The census query (RUNBOOK §1) returned FOUR consumers of `diagnose_persist_fix_plan`,
not the three 162 names: `feature-designer`, `fix-proposer`, `council-gate`, and
`council-gate-036scratch`. Checked before treating it as a finding — the fourth is
`is_active=false` with **0 runs in 30 days**, so 162 counting three is not wrong in
substance, and I have not written it up as a defect. Recording it because the next
person to run that query will see four and wonder.

`fix-proposer.persist_plan` still had `next_step=select_panel`, no `repair_step`, no
router. **Still valid.** Go half live on both replicas (2 hits, positive control 11,
negative control 0 — controls in the SAME exec, per the fleet landmine that a roll is
not evidence).

## The check that would have caught a fleet-wide regression, and it was a DATA check

The new router is `conditional_branch` on `plan_persisted.plan_valid == true`. I read
`conditional_branch_action.go` and satisfied myself that a dotted path resolves from
collected_data root and that `compareValues(nil, "true")` is false, so an unresolvable
field routes to `repair_plan` rather than onward to a council. **That reasoning was
correct and it was not sufficient**, because it answers "does the parser handle this
syntax" and the real risk was the opposite one: if a step's result were *wrapped* the
way `execute_llm_prompt` leaves its object under `<output_field>.result`, the field
would never resolve, EVERY valid plan would route to repair, and that loop is **not
bounded** — the repair counter counts refusal artefacts, and a misrouted valid plan
writes none, so it spins to `fuel_budget`.

Measured instead of reasoned: live rows carry `plan_persisted` **unwrapped** at
collected_data root, keys `[edit_count, files, persisted, plan_json, plan_valid,
summary]`, `plan_valid=true`. Hazard excluded. The general lesson is the one already in
`MEMORY.md` in another form — a config key that resolves to nothing looks exactly like
a live one, and only the data can tell you which you have.

## MISSTEP — I nearly closed this on a measurement whose retention I had not checked

I ran a census of `fix-proposer` runs by `current_step`/`status` intending to state how
often the defect had actually bitten. `orchestration_states` holds roughly today only,
so the query returned a handful of rows and would have supported a confident,
meaningless number. Worse, the defect **destroys its own evidence**: a non-opted-in
refusal writes no artefact, no `agent_error_log` row, and leaves
`orchestration_states.error` NULL, so there is no population to count even inside
retention. The honest statement is `[UNMEASURED, AND UNMEASURABLE IN RETROSPECT]`, and
that is what went in the bug file — not a number.

## Decision — did NOT induce a refusal on fix-proposer, deliberately

The 099 lane's runbook gives the method (arm `persist_plan.config.max_edits` to 1) and
records its own misstep: it armed before checking the queue. I checked first, and found
**three other lanes' `fix-proposer` runs in flight**. The arm is a fleet-wide edit to a
shared live agent with a ~30-minute dispatch window, so every run in that window trips.
Declined. Recorded in the bug file as a **gap, not a pass** — the route
`persist_plan → check_plan_valid → repair_plan → persist_plan` is `[UNVERIFIED]` on
`fix-proposer` specifically. It is proven on `feature-designer`, which runs the same
shared Go code and the same-shaped graph (3 refusals on 2026-07-31: 2 routed to repair,
1 exhausted terminal).

## The residual, and taking it to the council rather than just fixing it

Verifying turned up something 162 does not mention: `planValidationRefusal` has FIVE
terminal exits and only ONE wrote the operator-facing `agent_error_log` row, while the
comment on `planRefusalErrorCode` told the reader that row's absence was dispositive.
Measured: 3 rows fleet-wide, all `feature-designer`, all from that one exit.

I did not just fix it, because fixing it means changing behaviour that **four tests
guard with the words "must not touch the DB"**, written days earlier by the lane that
owns the action. That is a scope judgement, not a cleanup, so it went to the council
with the question posed explicitly in the rationale
(`SUBMISSION_CORR 7b1eb170-50b9-4c2f-b5d5-25fd6cf88c2b`). **APPROVED**, 9 reviewers, 3
advisory objections, none high-severity — and the verdict arrived in about 4 minutes,
against the ~30 minutes CLAUDE.md budgets for.

> **Timestamp trap, nearly a wrong call:** the `council_report` row read
> `2026-07-31 21:51+00` and I had submitted at 22:47 by the wall clock, so my first
> reading was "this verdict predates my submission, it is somebody else's". The DB is
> **UTC**; the shell is **BST (+01)**. 21:51+00 *is* 22:51 BST, four minutes after. I
> caught it before acting on it, but the shape — a timestamp that looks like it refutes
> your own causality — is worth having written down.

## MISSTEP, AND THE MOST USEFUL THING IN THIS FILE — the guard I added shipped INERT

The guardian seat's objection was that the containment claim ("only opted-in consumers
can reach the new write") was asserted from line numbers rather than tested, and asked
for a test pinning it. I wrote `TestPlanRefusal_OptedOutNeverRecords` the obvious way:
register no expectations, call the function, assert `mock.ExpectationsWereMet()`.

It passed. Then I mutated the code — moved `recordPlanRefusal` above the opt-in check,
which is precisely the edit that would silently change a non-opted-in consumer's
contract — and **it still passed**. So did all three pre-existing "must not touch the
DB" assertions. **Four guards, none of them discriminating, one of them written by me
about ten minutes earlier in direct response to a council objection.**

The cause: `mock.ExpectationsWereMet()` reports expectations that were **registered and
not consumed**. With none registered it is trivially satisfied. It never sees an
*unexpected* call — and `recordPlanRefusal` is best-effort by design, so it swallows the
driver error too, leaving nothing for the caller to observe.

Fixed with `dbTouchWatcher()`: an observed `zap` logger, keyed on the three messages
that every DB path here emits on failure (and with sqlmock holding no expectations,
every call fails). The same mutation now fails three tests. Retrofitted the three
pre-existing guards too — leaving assertions in place after proving them inert is worse
than never having written them, because the next reader counts them as coverage.

**What this cost, and what it nearly cost:** nothing, because the mutation check is
cheap and I ran it. It nearly cost the exact thing the guardian seat was worried about,
delivered with a test next to it certifying the opposite. `MEMORY.md` already carries
"a quiet-test passes when the RULE is gone — assert the mechanism FIRED"; this is that
lesson arriving through a different door, and the door is the assertion API reading like
English while meaning something narrower.

**Sizing, honestly:** 33 files in the tree call `ExpectationsWereMet`. None is *wholly*
inert (every one registers expectations somewhere), so the failure is per-TEST, not
per-file. I found and fixed 4, all in this one file. Whether the pattern exists in the
other 32 files is **[UNMEASURED]** — a per-test audit was out of scope for this bug and
is a decent candidate for a sweep.

## What shipped where

- **Config (LIVE immediately):** SQL 273 applied 2026-07-31 ~22:40. Verified at the
  live row, not at the migration's own verification block. Re-application now raises.
- **Go (committed, NOT live until the next chassis roll):** `417d6fd87`. So the
  refusal-visibility half of this work is *fixed and inert* until someone rolls — which
  is why 162 closes on the config half, whose defect it actually names, and the Go half
  is recorded as pending a roll rather than claimed as done.

## 2026-08-01, closing out — two more, and the first is the same bug I had just written up

**I printed "build: clean" from a shell line that could not report otherwise.** Final
verification pass, written as `go build ./... 2>&1 | head -5; echo "build: clean"`. The
`echo` is unconditional — it runs whether the build succeeded or not. `go build` had in
fact FAILED, and my own verification step announced the opposite. Caught only because the
test command on the next line failed loudly and I went looking for why.

This is **the same failure mode as the landmine I had appended twenty minutes earlier**:
an assertion that cannot fail, sitting where evidence is supposed to be. I wrote the
entry about `mock.ExpectationsWereMet()` and then made the shell-script version of it in
my own verification. Recorded because that is the point of WRONG_CALLS — knowing the
shape does not stop you producing it; only checking the exit code does.
**The check:** `; echo` is not a verifier. Gate it (`&& echo OK || echo FAILED`) or print
`${PIPESTATUS[0]}`. A "verified" line that a failure cannot suppress is decoration.

**What it revealed, which was NOT mine:** `go build ./...` fails in the WORKING TREE on
`apply_gap_plan_action.go:467` — `undefined: resolveNewPageConflict`, the 081 lane's
uncommitted WIP. **HEAD is clean.** Confirmed properly rather than assumed, by building a
`git archive HEAD` extraction in a scratch dir: `go build ./...` exit 0, the refusal suite
green, `./platform/...` green. This matters because `make build-<service>` archives from
committed HEAD, so an image built now carries my change and not their half-finished
symbol. A broken working tree is not a broken build here — but the only way to know which
you have is to build HEAD, not the tree.

**My LANDMINES.md and WRONG_CALLS.md appends were swept into another session's commit**
(`7f08dbbc5`, a provocation-pipeline commit) between my write and my own commit. Verified
present at HEAD, so nothing was lost and forward-only holds. Exactly the hazard CLAUDE.md
names — "committing per task stops you sweeping up others' WIP; it cannot stop a session
that still runs `git add -A` from sweeping up yours". Noted in the close commit so the
trail is followable, since a reader looking for those entries in my commit will not find
them there.
