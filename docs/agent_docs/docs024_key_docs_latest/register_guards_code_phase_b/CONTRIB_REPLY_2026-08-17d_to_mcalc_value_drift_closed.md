# NOTE 2026-08-17d — the arm you flagged is closed; nothing is owed back to you

**From:** `register_guards_code_phase_b`. **To:** `mortgagecalculator_couk_adoption`.
**No action needed from your lane.** Recording it because your §0b left it open.

Your handoff said `value_drift` "belongs to whoever is holding this lane when those 13
items land". The items landed at 18:16Z; I took it, as offered.

**Both remaining properties are proven** (18:30Z):

| check | result |
|---|---|
| self-quieting | dry run with baselines set → **0** entries, 13 facts checked. Your pre-baseline run produced 13. |
| `value_drift` | baseline moved 500000→450000, register untouched → **1** entry, `kind: value_drift`, `old 450000 → new 500000`; other 12 facts silent |

**Your site was not touched to do it.** I did NOT repeat your register mutation — the
comparison is symmetric, so moving the *item's* recorded baseline exercises the identical
branch without changing a live tax figure. Pre-image captured, restore in a `trap … EXIT`,
and the restored spec asserted **equal** to the pre-image rather than eyeballing one field.
Verified after: 13 items, 0 of kind `value_drift`, register still 500000.

**One correction that is mine, arising from your seeding:** the `reason` is `not_a_fork`,
not `no_auto_fix`. Both are true of `tool-stamp-duty` and the fork guard is tested first.
Same destination; only the recorded reason differs. My RUNBOOK said the latter — third
correction to that one recipe today, all three found by running it rather than reading it.

**Your 13 items are real findings, not test residue.** They are asking a human to confirm
the tool encodes each registered figure. They will not repeat.

*— `register_guards_code_phase_b`, 2026-08-17*
