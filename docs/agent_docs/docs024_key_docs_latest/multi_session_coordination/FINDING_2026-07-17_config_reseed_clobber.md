# FINDING — DB config re-seeds clobber concurrent config work (the config analog of WIP-bundling)

*Filed 2026-07-17 from the "diagnosis fixloop 3" thread. The coordination
workstream's shipped rules (git commit-per-task, build-from-committed-HEAD, the
090 coverage check) all address CODE and WORK-ITEM collisions. This is a THIRD
collision surface they don't yet cover: **live-mutable DB config**.*

## What happened
- ~17:55Z: this thread applied the v7 fix-proposer seed (added the code_lookup
  verify tier — new step, rewired run_checks→code_lookup, extended 3 reviewer
  prompts).
- ~18:33Z: another session applied a v8 fix-proposer seed (added 3 reviewer
  seats). Its full re-seed of `agent_definitions.default_config` **silently
  overwrote the v7 wiring** — the code_lookup step, the run_checks rewire, and
  the reviewer-prompt edits were all gone. The Go action stayed in the binary;
  only the DB wiring that used it was lost.
- Detected only because this thread re-verified the config after an unrelated
  image deploy. Nothing flagged the clobber.

## Why the shipped rules don't catch it
- git rules govern the working tree / commits. Agent definitions live in
  `agent_definitions.default_config` (jsonb, clients_db) and are edited by
  `psql` seed scripts, **live immediately, no image, no commit, no git trace**.
- A seed that does `UPDATE ... SET default_config = <whole object>` (full
  re-seed) is the config equivalent of `git add -A` — it takes the entire
  object from ONE author's view and writes it, discarding any concurrent
  field-level change another session made.

## The mitigation that worked here (candidate rule)
**Patch-style seeds, not full re-seeds.** The v7 seed was written as targeted
idempotent `jsonb_set`/guarded `UPDATE`s on specific paths, each a no-op if
already applied. Re-applying it ON TOP of v8 composed cleanly — because v8 kept
the same structure, every v7 anchor still resolved. A full-object re-seed cannot
compose; a patch seed usually can.

Proposed additions to the coordination practice (owner's call, like the git rules):
1. **Seeds edit the narrowest jsonb path that carries the change** — never
   `SET default_config = $whole$...$whole$` for a shared agent def that other
   threads also patch. Use `jsonb_set` on the specific step/field.
2. **Guard every seed UPDATE** (`WHERE ... NOT <already applied>`) so it is
   idempotent and re-runnable after someone else's concurrent patch.
3. **Back up before a full re-seed** (`bak_agentdef_<type>_<date>`) — already
   convention for risky edits; make it mandatory for whole-object writes.
4. Consider a lightweight "who last touched this agent def" surface (the digest
   already reads agent-config snapshots) so a clobber is visible next digest.

## Scope
`fix-proposer` is the hot spot (fixloop + concept-register + this thread all
edit it), but ANY shared agent def edited by more than one workstream has this
exposure. The diagnose-agent def was edited 5+ times across threads this week.
