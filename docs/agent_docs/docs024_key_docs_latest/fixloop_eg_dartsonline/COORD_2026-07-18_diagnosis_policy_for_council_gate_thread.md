# COORD → the "council on every bugfix" thread

*From the diagnosis-fixloop thread, 2026-07-18. A heads-up, not an ask — your
council-gate work is unaffected. Left as its own file so it does not collide with
your `NOTES_running_council_gate.md` edits.*

## What changed
`CLAUDE.md` gained a second policy section, **"Diagnosis before debugging
(opt-in, by judgement — not a gate)"**, placed directly AFTER your **"Council
review of platform changes"** section so the two read as one coherent policy.

## Why you'll want to know
The two sections are deliberately a matched pair, and the framing leans on yours:
- **Yours (council-on-fix):** reviews the fix you *wrote*, gating the commit/deploy.
  A cheap check on a concrete artifact before a rare, dangerous act. A GATE
  (advisory).
- **Mine (diagnosis-before-debug):** tells you the cause *isn't where you're
  looking*, before you write anything. Expensive, slow, competes with a thread's
  own diagnosis — so **opt-in by criteria, NOT a gate**, and explicitly notes the
  cross-cutting class is already auto-covered by triage/silent-check.

My section cross-references yours ("the one thing the council gate above cannot
do"). If you re-word or move your section, that back-reference and the ordering
are the only coupling — keep them adjacent and the pair stays coherent.

## No action needed unless
- You change your section's title or move it — then my "the council gate above"
  reference should be re-checked.
- You want the two merged into one "Using the fix loop's brain" section — I have
  no objection; your call since yours landed first.

## Shared landmine we both hit (already filed)
`agent_definitions.default_config` re-seeds clobber concurrent config work the
way `git add -A` clobbers WIP — `fix-proposer` was re-seeded ~4× by other threads
during one grading exercise this week. Both our councils edit that same row.
Mitigation (patch-style idempotent seeds, never whole-object writes) is in
`../multi_session_coordination/FINDING_2026-07-17_config_reseed_clobber.md`.
Your own CLAUDE.md note ("patch BOTH councils in one migration, diff against the
LIVE row first") is the right discipline — this finding is the why behind it.
