# PLAN — `bugs_open/106`: the concept register's coverage sensor had no cadence

**Opened** 2026-07-28. Ownership checked first: `scripts/who-owns.py 106` → no
owning workstream; last commit touching the file was a triage sweep on 07-27. The
concept-register directory has recent commits, but they are *other threads adding
entries*, not work on the detector. Taken up here.

## 1. What was already done, and what was left

106 was filed 2026-07-27 and **both its recommended candidates were implemented
the same day** by the concept-register side:

- **Candidate 2 — DONE.** `covers-through:` stamps on all 109 register files. The
  *misleading* half is closed: nobody can now mistake the register for complete.
- **Candidate 1 — BUILT, not wired.** `102_CHECK_register_coverage.py` +
  `102_coverage_ratchet.txt` exist, work, and have already caught real gaps.

The bug then stayed open for one specific reason, stated in its own post-roll
triage and worth quoting because it is the whole scope of this fix:

> *"This file's candidate 1 says the point is to make drift visible **on a cadence
> instead of by coincidence**. The cadence does not exist … It is not in
> `.githooks/`, not in `scripts/pattern-check.py`, not in `.github/`, and not a
> `scheduled_tasks` row. So the sensor runs when a human remembers to run it —
> which is the same 'detected by coincidence' mechanism this bug was filed about,
> moved one step earlier. **A fourth tool that must be invoked by coincidence does
> not retire it.**"*

## 2. The fix

`check_register_coverage` in `scripts/pattern-check.py` — advisory, on the commit
path, already run by `.githooks/pre-commit`.

**Trigger: a commit that CREATES a workstream directory the register has never
heard of.** Not a periodic sweep.

Why that trigger, and it is the design decision worth defending: **a cron reports
drift up to a week late, to nobody in particular.** This fires at the instant the
gap is created, in front of the one person who can close it in ten seconds. The
standing practice is to put the check where the error is *made* — the same reason
a warning went into `092_TRIGGER_experience_plan.sh` rather than only into a doc
this week.

Three properties that keep it from becoming wallpaper:

- **Only NEW directories fire.** 43 existing workstreams are uncovered and on the
  ratchet; that is accepted backlog, and flagging active work on it every commit
  is exactly how a check gets ignored.
- **It imports the sensor rather than reimplementing `is_covered()`.** Two
  hand-maintained copies of one matching rule is the drift class this platform
  keeps filing bugs about (`idx_swi_dedup` ↔ `workItemTerminalStatuses` is the
  standing example). There is one implementation; this calls it.
- **Two ways to go quiet, both correct.** Add a register entry (if the directory
  will hold a reusable mechanism) or add it to the ratchet (if it is site content
  or a one-off). The message says both.

## 3. Measured before inclusion, per `pattern-check.py`'s own bar

That file requires a new check to state its fire rate against the real history.

```
1,500 commits scanned   fires: 4   rate: 0.27%   false positives: 0
```

Quieter than every existing check (README 0.7%, SUMMARY 2.0%, twin ~2%) — and
that is the correct shape here, because the population is "commits that create a
brand-new workstream", which is genuinely rare. **Precision matters more than rate
for a check this quiet, so all four fires were inspected**: `memory_index`,
`bugs_sweep_2026_07`, `bugfix_066_spawn_image_tag`, `gemini_content_provider`.
All four are real, and the last two are *precisely the pair* 106's own triage
records the sensor finding on 07-27 — now caught at creation instead of days later
by a human who happened to run it.

## 4. Verification — the induced gap 106 demands

106 says: *"induce the gap, because a report that is green on a register somebody
has just hand-patched proves only that the patch happened."* Done, three arms:

| arm | setup | result |
|---|---|---|
| 1 | two new uncovered workstreams staged | **both fire** |
| 2 | A added to the ratchet | **only B fires** — ratchet suppression works |
| 3 | B given a register entry instead | **B goes quiet, A still fires** |

Plus the negative control: on a commit touching no workstream directory it prints
nothing and costs 40 ms.

## 5. Scope deliberately NOT taken

- **Content staleness.** The sensor asks only whether a subsystem is *represented*,
  never whether an entry is *accurate*. That is stage-2 verification's job and far
  more expensive; conflating them is how a coverage check becomes an audit nobody
  runs. Noted because this thread hit a live instance of the accuracy problem today
  — see NOTES, "the other half".
- **The other sensor inputs** candidate 1 imagined (agent types, action names,
  migration files). The implemented sensor covers workstream directories only. Not
  widened here: the wiring is the bug, and widening the sensor is a separate change
  with its own fire-rate measurement.
- **R2 / the mission-review promotion question** raised in 106's triage. It belongs
  to the concept-register workstream and is blocked on an empty denominator, which
  no amount of waiting fixes.
