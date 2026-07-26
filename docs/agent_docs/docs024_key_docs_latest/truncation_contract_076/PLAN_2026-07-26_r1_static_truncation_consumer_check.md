# PLAN — R1: catch a truncation-tolerating step with no reader BEFORE it runs

**Opened 2026-07-26 (evening), continuing from
`HANDOFF_2026-07-26_continue_here.md` §R1.** This is the workstream doc set the
handoff said to start if anyone picked up R1 — it is a workstream that will
outlive one session, so it gets the standing five.

## The problem, in one paragraph

`bugs_closed/076` shipped a **runtime** guard: a step that sets
`tolerate_truncation: true` may only keep a cut LLM response if some other step
in its workflow reads the `__truncated` marker. That guard is a **floor that
fires late** — it only speaks when a response is actually cut, in production, and
it speaks by failing that run. The bad *config* can sit in the fleet for weeks
before anyone learns it is bad. R1 is the layer that reads the config itself and
says so before any run exists.

Raised by the council seat `guardian` (medium, corr `470678f4`); the submission
had never discussed that axis, which was a fair hit.

## The constraint that shapes everything

**There is no registration or deploy step to hook.** CLAUDE.md's own invariant is
that *DB config is live immediately*: a seed can add `tolerate_truncation: true`
to a live workflow with no build, no deploy, no restart. So "validate at workflow
registration time" — the literal shape the council asked for — **does not have a
place to attach**. Whatever we build is a check someone or something *runs*, not
a gate the config must pass through.

Two layers, chosen because they fail in different directions:

| layer | catches | misses | when it speaks |
|---|---|---|---|
| **L1** live-DB lint (`103_LINT_truncation_consumer.py`) | everything actually in the fleet, however it got there — seed, patch, hand-run `UPDATE` | anything not yet applied; only speaks when run | on demand, after any seeding |
| **L2** pre-commit check (`pattern-check.py`) | a **committed seed** that arms tolerance in a workflow it also defines | anything that reaches the DB without passing through a commit | at the moment of the edit, free, no cluster |

L1 is the authority (it reads the real workflow). L2 is the cheap one that fires
where the author can still fix it. Neither blocks — see the ADVISORY doctrine in
`pattern-check.py`'s header, and the handoff's ruling that a **blocking** gate on
this would make a bad seed take the fleet down.

> **CORRECTION to the brief, made before writing any code.** The handoff proposed
> L2 as *"a check inside `scripts/pattern-check.py` or the migration lint, so a
> **seed file** introducing the bad config is caught at commit time"*, and
> reasoned that this "catches the common path — most config arrives as a
> committed seed". **Measured, that is only half true, and the half it gets wrong
> would have produced a check that fires on correct work.** All three files in the
> repo that arm `tolerate_truncation` today —
> `sql_for_agents/177_council_tolerate_truncation.sql`,
> `PATCH_fix_proposer_021_…`, `PATCH_feature_designer_022_…` — are **`jsonb_set`
> patches**: they name the flag and the target steps, but the *workflow* they
> patch is in the database, so nothing in the file says whether a reader exists.
> A textual check over those files can only ever guess, and on all three the
> honest guess is wrong (all three targets ARE guarded, by
> `diagnose_council_decide`).
>
> So L2 is scoped to the case a file can actually answer: a SQL file that
> **embeds a full workflow** (`"steps": {…}`) and arms `tolerate_truncation`
> inside it. That shape is common — **62** SQL files in `docs/` embed a workflow
> containing `execute_llm_prompt` — and **zero** of them arm tolerance today, so
> the check fires **0 times on the entire corpus** while its true-positive case is
> exactly the bug's next instance. Patch-style files are deliberately skipped and
> the reason is stated in the code, because "why doesn't this fire on 177?" is the
> first question a reader will have.

## The landmine the handoff named, and how this answers it

The measurement query in the handoff hard-codes
`('diagnose_council_decide','verify_report_prose')` — **a hand copy of the Go
registry `truncationAwareActions`**. Two hand-maintained lists that must agree is
precisely the drift class this platform keeps paying for (the council gate's `099`
roster mirror exists for it; `102_LINT` exists because of another instance).

**So no checker in this work holds a copy.** `scripts/truncation_registry.py`
parses the registry out of `platform/orchestration/actions/truncation_guard.go`
at run time; L1 and L2 both import it. If the registry moves or is rewritten in a
form the parser cannot read, the parser **raises** — it never falls back to a
remembered list, because a stale list would make the lint report a clean fleet
that is not clean.

## Phasing

1. **P1 — parser.** `scripts/truncation_registry.py`, plus a cross-check that
   every parsed name is a real registered action in `registry.go` (a typo'd
   registry entry is itself a 076 re-opening).
2. **P2 — L1 lint.** `…/fixloop_eg_dartsonline/103_LINT_truncation_consumer.py`,
   modelled on its sibling `102_LINT_council_seat_parity.py`: read-only, no
   credits, advisory by default, `--strict` for a caller that wants an exit code.
3. **P3 — L2 check.** `check_truncation_without_reader` in `pattern-check.py`,
   measured against the corpus before it is wired in — the same bar every other
   check in that file was held to.
4. **P4 — validation by induction.** The fleet is clean (37/37 guarded), so a
   clean report proves nothing at all. Seed a deliberate offender **and** a
   guarded control, confirm the lint separates them, delete both.
5. **P5 — record.** Case file, handoff, `016b` pointer, commits by pathspec.

## Decisions, with reasons

- **L1 lives in `fixloop_eg_dartsonline/` as `103_…`, not in `scripts/`.** It is
  a live-DB lint in the 097–102 family and its nearest sibling (`102`) is there.
  `scripts/` holds the git-time checks. The parser is the exception: it is shared
  by both, so it goes in `scripts/` and L1 reaches it via `git rev-parse`.
- **No council submission.** The gate's scope is `platform/`, `internal/`,
  `pkg/` (owner ruling 2026-07-17, enforced client-side in `097` line 78). This
  work touches `scripts/` and `docs/` only, so `097` would refuse it, and
  `FORCE=1` past a scope refusal is not what that escape hatch is for.
  **[If any Go changes later, that flips.]**
- **L1 reports the `accepts_truncated` hatch users separately.** That is R2 in the
  handoff — a config flag that declares "my action handles a partial", trusted and
  verified by nothing. R1 was always the layer where R2 could be validated; making
  the hatch users *visible* in the same report is the cheapest form of that, and
  today the list is empty, so the report also proves the hatch is unused.
- **Not doing: the chassis startup scan** (the handoff's third shape). It needs a
  Go change, a council round and an image roll to add a layer that catches, at the
  next roll, a subset of what L1 catches on demand today. If L1 turns out to be
  run too rarely, that is the reason to revisit — not before.
- **Not doing: a blocking gate.** Explicitly ruled out in the handoff without an
  owner ruling, and `pattern-check.py`'s header records where a blocking check on
  a shared tree ends up (a fleet-wide outage, then permanently disabled).

## Done looks like

An offending seed is caught before it can be exercised — proven by inducing one,
not by a clean report — and the report shows zero offenders on today's fleet
(re-grounded 2026-07-26 21:5x: 37 tolerating steps, `council-gate` 16,
`fix-proposer` 16, `feature-designer` 5, **all guarded**).
