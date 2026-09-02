# 426 — the check for "applied by hand, never recorded" already exists, is free, and is mandated by CLAUDE.md. Seven live instances from one lane sat in it unnoticed for a month.

**Filed:** 2026-09-02 · **Branch:** `087_towards_multiple_domains` · **Status:** OPEN, evidence
first-hand, not fixed.
**Severity:** medium — nothing is broken *now*, but the state it fails to surface is the one that
blocked the migration runner for three days (migration 151), and the population is unmeasured
fleet-wide.
**Class:** undriven mechanism / attention gap. **NOT** a missing detector — see §2, which is the
whole point of the file.
**Filed by:** the `bugfix_314_council_scope` lane, at the request of and with evidence contributed
by the **`dispatch_throughput`** lane, which found the seven instances by auditing itself after an
unrelated heads-up. Credit for the finding is theirs.

---

## 1. The situation, in plain terms

A database migration on this estate can be applied in two ways. Normally the runner applies it and
writes a row into a ledger table (`schema_migrations`) saying so. Sometimes a migration is applied
**by hand** — deliberately, because the ordering matters — and then somebody has to tell the ledger
after the fact (`run-migrations.sh --record-only <file> --note '<what was verified>'`).

If that second step is missed, the file sits in the directory looking, to the runner, exactly like
a migration that has never been applied. The next person who runs `--apply` **replays it**. A
migration that is not written to be safely re-runnable then dies on a raw duplicate-key error
(23505) that reads like broken SQL and is not. Migration 151 did this and **blocked the runner for
3 days**.

## 2. The part that makes this worth filing: the check is already there

`scripts/migration/run-migrations.sh` **in its default, no-argument, dry-run mode** already
computes precisely this set. Read 2026-09-02:

- `:328-330` — builds `PENDING` as every appliable file with no ledger row. **That IS the census.**
- `:338-339` — prints it: `Pending (N):` then the filenames.
- `:341-356` — runs the idempotency lint over exactly that set (`bugs_open/007`).
- `:358-390` — probes each one in a doomed transaction to ask the file itself whether it has
  already been applied.

It costs nothing, it is read-only in the dry-run path, it needs no arguments, and **CLAUDE.md
already mandates it**: *"Migration runner practice — dry-run per SESSION and after every roll."*

So the honest statement of this bug is **not** "no check exists". It is:

> **A free, mandated, already-built check went unrun long enough for seven instances to accumulate
> in a single lane, and nothing anywhere noticed that it had not been run.**

This estate has a standing lesson for exactly this shape — *a silent mechanism is usually
UNDRIVEN, not missing* — and a fix that adds a second detector would be the wrong move twice over:
it would duplicate a working check, and it would leave the actual failure (nobody looks) untouched.

## 3. Evidence — the seven, contributed by the `dispatch_throughput` lane

[MEASURED 2026-09-02, by that lane, reported to this one] **Seven runner-appliable migrations
applied by hand with zero rows in a 481-file ledger**: `582`, `583`, `584`, `637` (August) and
`671`, `672`, `673` (Sunday). All seven have since been `--record-only`'d with dated verification
notes — the runner's own sanctioned remedy. **The repair is done; the reason they were invisible
for a month is not.**

**A second, sharper finding inside it — the safety net mis-reports rather than fails.** The probe
(§2, `:358-390`) degrades gracefully by matching a guard's `RAISE` message against `/already/i`: a
file that says "already applied" is understood to have been applied out of band. That contract is
on the **message vocabulary**, and nobody has ever written it down. Migration `672` has a drift arm
whose message reads *"pre_query md5 % is not 671's shipped text — drifted, investigate"* — no
"already" in it. Because `673` later edited the same task row, a probe of `672` lands on exactly
that arm and reports **"drifted, investigate"** when the true state is **"applied by hand, fine"**.

That is the dangerous half: a **truthful message that reads as danger when the state is safe**.
An operator who trusts the probe stops and investigates; one who has seen it cry wolf once starts
discounting it. Neither is the intended behaviour, and it is invisible from the runner's side
because the runner is faithfully reporting what the guard said.

## 4. Why a per-commit checker is the WRONG home (ruled out here so nobody re-proposes it)

The obvious suggestion — and it was made — is to add this to `scripts/pattern-check.py`, which
already runs on every commit in every session. It should not go there:

- that script has **no database connection** at commit time, so it could only guess at the ledger;
- it would become a **fifth hand-maintained copy** of the "which files are migrations" vocabulary,
  which is the defect `bugs_closed/314` exists to remove (four copies; see that file, and
  `scripts/council-scope.sh`'s header for why the copies drift);
- the failing moment is not commit time. These files were *committed* correctly. They were
  **applied** and not recorded.

## 5. Fix candidates, ordered by what closes the door

1. **(Preferred) Drive the existing dry run on a schedule and make its silence legible.** The
   estate already has a proven shape for this: a daily CronJob that runs a check and writes **one
   `doc_notes` row per run, including clean ones** — so that a MISSING row means *the job did not
   run*, which is distinguishable from *all clear*. Three live checks already use it
   (`optional-key-budget-check`, `single-owner-carriers-check`, `instance-token-adoption-check`),
   and the "quiet runs must still write a row" rule is RFC_022's lesson, learned the hard way.
   Contributed as a candidate, not a prescription, by the `dispatch_throughput` lane; **how it gets
   driven is this bug's actual question.** Makes the bad state visible without adding a detector.
2. **Contract the probe's message vocabulary** (fixes §3's second finding, and is independent of
   candidate 1). Either require every guard arm that can fire post-apply to carry `already`, or —
   better, because it does not rely on every future author reading a rule — stop keying graceful
   degrade on prose at all and give the probe a structured signal (a `SQLSTATE`, or a sentinel the
   guard raises). A contract enforced by message text is not a contract.
3. **Make `--record-only` harder to forget at the moment of the rename.** A `_HOLD` file *cannot*
   be recorded while it carries its suffix — `run-migrations.sh:245-250` refuses a sidecar by
   design — so the sequence is forced: apply by hand → rename → record. The window between rename
   and record is structural. A commit-time nudge on a `_HOLD`→plain rename ("this file is now
   appliable; is it recorded?") would fire at the one moment the author is definitely looking.
   Cheaper than 1, narrower, and it only covers the `_HOLD` path, not the plain-file path that
   produced six of the seven.
4. **Do nothing.** Named because it is the status quo and the runner does eventually catch these —
   noisily, at whoever runs `--apply` next, which is by construction not the person who caused it.

## 6. How to verify a fix

A positive control is essential: "the report is empty" is equally consistent with "nothing is
wrong" and "the job did not run" — which is this bug's entire subject.

- **Positive:** plant a known appliable-and-unrecorded file (or `--record-only` one *back out* in a
  scratch DB) and confirm the mechanism names it.
- **Negative:** with the estate clean, confirm the mechanism still emits its heartbeat — a
  `doc_notes` row, or whatever the chosen channel is — **on the clean result too**. If a clean run
  is silent, the fix has reproduced the bug.
- **For candidate 2:** induce `672`'s drift arm against an already-applied-then-edited row and
  confirm the probe reports it as applied-out-of-band rather than "drifted, investigate".

## 7. On process, stated rather than omitted

No `090` diagnosis run, deliberately, per the owner ruling of 2026-07-31's named escape hatch.
There is no structural claim here needing a loop: the mechanism is read at source
(`run-migrations.sh:328-390`, cited by line above), the seven instances are first-hand from the
lane that found and repaired them, and the "check exists but is undriven" conclusion follows from
reading the script rather than from inference. What this file asserts that is NOT first-hand is the
**fleet-wide population** — see §8.

## 8. Explicitly NOT measured

**`[UNMEASURED]` How many appliable-and-unrecorded migrations exist fleet-wide right now.** Seven
came from one lane that happened to look. The number across all lanes is unknown, and the first
action for whoever picks this up is one command:

```sh
./scripts/migration/run-migrations.sh          # default dry run; reads Pending (N)
```

Do that **before** designing anything — it sizes the bug, and per §6 it is also the positive
control that proves the mechanism works at all.

## 9. Related

- `bugs_open/007` — the idempotency class (Class C) this protects against; the runner's lint and
  `pattern-check.py`'s commit-time half both exist because of it.
- `bugs_closed/314` — the four copies of the migration vocabulary, and why a fifth must not be
  created here. Its live residual (the drifted copy in `pattern-check.py`) is being fixed by this
  file's filing lane, separately.
- `CLAUDE.md`, "Migration runner practice" — the mandate that already exists and was not followed.
- RFC_022 / `optional-key-budget-check` — the precedent for candidate 1, including why a clean run
  must still write a row.

## 10. UPDATE 2026-09-02 (later, same day) — candidate 1 FIXED and LIVE; candidate 2 is the open residual

**Status change: STILL OPEN, but only for candidate 2 now.** Candidate 1 (§5.1 —
drive the existing dry run on a schedule, make its silence legible) is built,
deployed, and proven against production. This is not "committed, inert until a
roll" — there is no image or roll involved; the mechanism is a CronJob, applied
directly, and already ran twice for real against the live cluster before this note
was written.

**Before building anything, re-ran the dry run myself to check the bug was still
real.** It was — worse, in fact: **34 live "applied by hand, unrecorded" instances
right now**, not the seven the bug documents (those seven, from `dispatch_throughput`'s
own repair, correctly did not reappear). Full measurement, precedent research, and
the build+deploy trail: `docs/agent_docs/docs024_key_docs_latest/
bugfix_426_migration_ledger_drift/` (NOTES + PLAN, the standing five).

**What shipped:** `deployments/kustomize/services/migration-ledger-drift-check/` —
a daily CronJob (`45 7 * * *` UTC), `postgres:16-alpine`, no custom image. It
fetches the real migrations directory via a sparse, shallow, partial `git clone`
(measured 1.8s for the 16MB directory) and runs `scripts/migration/run-migrations.sh`
**completely unmodified** (only `PSQL_CMD`/`MIGRATIONS_DIR` overridden) — zero
duplicate vocabulary, per this file's own §4. Writes ONE `doc_notes` row
(`subject_type='pipeline'`, `subject_key='migration-ledger-drift'`) every run,
including clean runs, per §5's RFC_022 precedent.

**Live-tested twice against production**, not just designed: first run exposed a
real defect (exiting non-zero on findings caused a full re-run + a permanently
"Failed" Job, because unlike `single-owner-carriers-check`'s rare findings, this
check's findings are the ordinary daily state — 72 of 158 pending on the very
first run); fixed and re-run cleanly. Second run: single pod, exit 0, real
`doc_notes` row, **158 pending / 81 clean / 72 needing review (34 ALREADY, 38
INCONCLUSIVE, 0 DUP)** — confirmed by reading the row directly from
`doc_notes`, not by trusting Job status. Full trail, including the two mid-flight
defects found and fixed before/after deploying: this workstream's PLAN §3/§3b.

**Candidate 2 (§3's second finding — the probe's `ALREADY` verdict only fires when
a guard's message matches `/already/i`) is DELIBERATELY NOT fixed by this change**,
and is the reason this bug stays open rather than closing. Mitigated at the report
layer instead (the new job's NEEDS-REVIEW bucket includes `INCONCLUSIVE`, not just
`ALREADY` — confirmed live: 38 of today's 72 flagged items are exactly this
class), which the `dispatch_throughput` lane flagged independently and unprompted
when told about the 34-vs-7 measurement (credited in this workstream's NOTES). But
the underlying contract in `run-migrations.sh`'s `probe_file()` is still prose-
matched, not structural, and fixing that properly means editing a script ~30
concurrent sessions rely on interactively — a separate, higher-stakes change,
deliberately left for whoever picks this up next. See the workstream PLAN §1 for
the full reasoning on why it wasn't bundled in.

**Verification owed to a future thread, not done here:** watch for the job's first
*scheduled* (not manually triggered) run at 07:45 UTC and confirm it behaves the
same as the manual triggers did.

---

## CONTRIB 2026-09-02 (editorial_design_uplift) — §8's `[UNMEASURED]` number is **34**, and it is not `Pending (N)`

§8 names the fleet-wide population as the first action for whoever picks this up, and gives
`run-migrations.sh` as the command. Ran it twice today, either side of the `v1.0.1355` roll:

| | pre-roll (~15:00 BST) | post-roll (~22:20 BST) |
|---|---|---|
| `Pending (N)` | **164** | **176** |

**But `Pending` is the wrong number for this bug, and quoting it would overstate the problem by 5×.**
It counts every appliable file with no ledger row — most of which are genuinely unapplied and
perfectly safe. The population §2 actually describes — *applied by hand, never recorded, therefore
replayed by the next `--apply`* — is the subset the runner's own probe flags:

> **34 files** reported `!! <file> — LIKELY ALREADY APPLIED; its own guard raised:`

`[MEASURED 2026-09-02, post-roll]`. First twelve: `234`, `235`, `265`, `272`, `273`, `303`, `343`,
`345`, `363`, `370`, `433`, `434`. So the seven the `dispatch_throughput` lane found by auditing
itself were roughly a fifth of it, and the rest have never been looked at.

**Two process findings, both bearing on §5's "nobody runs it" diagnosis:**

1. **The check takes OVER FIVE MINUTES.** My first attempt died on a `timeout 300`. Whatever drives
   it on a clock needs a generous limit, and a human running it interactively will be tempted to
   abandon it.
2. **Piped through `tail` it prints NOTHING until it finishes** — which is indistinguishable from a
   hang, and is what made me check whether the process was alive. Anyone who tried it once, piped,
   and gave up would reasonably conclude it was broken. **That is a candidate contributing cause for
   the attention gap this bug is about**, and it is cheap to fix in the runbook line rather than in
   the script: run it unpiped, in the background, and read the file.

**Caveat on the 34, stated because this bug is about false confidence in a census:** `LIKELY ALREADY
APPLIED` is the runner's probe *interpreting a guard's RAISE message*, and §3 already documents that
this contract is on message vocabulary — migration `672`'s drift arm says *"drifted, investigate"*
and would be read as a genuine problem rather than as applied-out-of-band. **So 34 is the count of
files whose guards said something the probe could match, not a proven count of applied-and-unrecorded
files.** It is a floor, and it is the best available number until §5 candidate 2 (a structured signal
instead of prose matching) exists.

Not dispatched, nothing recorded, no file `--record-only`'d — the 34 are other lanes' migrations and
each needs its own verification before anyone records it.
