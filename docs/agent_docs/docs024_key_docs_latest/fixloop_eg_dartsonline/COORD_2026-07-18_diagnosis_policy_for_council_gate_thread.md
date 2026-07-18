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

---

## ADDENDUM 2026-07-18 — a blind spot in 099_SYNC_gate_roster.py (found doing the Sonnet 5 migration)

Moving the council to Sonnet 5 (owner decision D1: model → claude-sonnet-5,
reviewer max_tokens 3000 → 8000) exposed a real limitation in the mirror:

- **The sync detects roster STRUCTURE drift (added/removed seats, routing) but
  NOT config-VALUE drift inside an existing seat** (`config.ai_service.model`,
  `max_tokens`, and anything else `transform_step` copies verbatim). I migrated
  `fix-proposer`, ran `099 --apply`, and it reported *"Already in sync — nothing
  written"* because the seat NAMES and routing were unchanged — leaving the gate
  reviewers on `claude-sonnet-4-6 @ 3000` while fix-proposer was on
  `claude-sonnet-5 @ 8000`. Silent divergence, exactly the drift the mirror
  exists to prevent.
- I patched `council-gate` directly with the same idempotent patch-style loop
  (0NN_council_sonnet5_migration.sql, generalised to both types) to align them —
  this is NOT the structural hand-patch CLAUDE.md warns against (no seat/routing
  change), only a config-value the mirror provably can't propagate. Both councils
  are now `claude-sonnet-5`, reviewers `>= 8000`. Backups:
  `bak_agentdef_fixproposer_sonnet5_20260718`, `bak_agentdef_councilgate_sonnet5_20260718`.

**Suggested fix for the mirror (your tool, your call):** in the delta check,
compare each mirrored step's `config.ai_service` (and ideally a hash of the whole
transformed step) between fix-proposer and gate, not just the seat-name set —
and write when they differ even if the roster is structurally identical. Until
then, a model/max_tokens change to the council needs applying to BOTH types
directly (the migration seed does this), and CLAUDE.md's "run the mirror, don't
hand-patch" holds for SEAT changes but not for in-seat config changes.
