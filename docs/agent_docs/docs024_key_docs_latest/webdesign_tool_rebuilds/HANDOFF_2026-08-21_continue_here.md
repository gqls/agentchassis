# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-21 ~10:45Z. Supersedes `HANDOFF_2026-08-19_continue_here.md`.

**UPDATED 2026-08-21 15:05Z — PHASE A COMPLETE: 28 of 63 replaced (24 serve-confirmed; #25–#28
retired and graded, their serve-grades ride the queue — NO race exposure anywhere, retires all done).
FIRST TASK for a fresh session: batch serve-grades for regex-tester (`310accdf`), jwt-inspector
(`7c9deeee`), token-calculator, clip-path, and the oklch re-fix (`1fe89947`, control set agreed with
the 286/331 lane in NOTES 2026-08-21 14:15Z). Controls per tool pinned in NOTES. Then Phase C
(13 external-`<script src>` tools — read the superseded 08-19 handoff's Phase C section; brief from
BROWSER behaviour, retire the S3 asset with the slot). ⚠ ATTENDANCE RULE (rewritten twice, final
form): file, then HOLD THE TURN with a foreground poll loop (`for … do <check>; sleep 10; done`,
timeout ≤600s per call, chain calls) until the retire lands — background watchers deliver LATE
(measured 6h), and turn-end is not knowable in advance. The seed-496 incident (fixed by owner-lane
hotfix 532 within the hour) and TL-047's absence-arm proof are in NOTES 2026-08-21 14:15/14:30Z.**

Read: this file → `PLAN_2026-08-15_…` (design + owner rulings) → `RUNBOOK_…` (every command — note
the REWRITTEN "retire race" section) → `NOTES_…` (evidence, newest at bottom; the 2026-08-20
17:05Z entry is the day's hardest lesson) → `SUMMARY_2026-08-20_the_walls_came_down.md`.

## The recipe — unchanged six steps, ONE rule rewritten

The six steps stand as in the superseded handoff (read live script → brief from intent → file with
gates → grade the RUN → grade the COMPONENT by mechanism → retire IMMEDIATELY → serve-grade
cache-busted with old-only negatives). **The rewritten rule: "attend" means THE SESSION'S TURN STAYS
ALIVE, polling, from filing until the retire lands. A background watcher is NOT attendance** — its
firing is on time but its delivery waits for the next user interaction and is unbounded: measured SIX
HOURS on 2026-08-20, during which the oklch page publicly served BOTH tools (WRONG_CALLS entry +
RUNBOOK "retire race" section). If your turn must end, do not file.

## The chassis roll (v1.0.1321, pods 2026-08-20 19:51Z, build revision `0483e7f4e`)

- **TL-048 IS LIVE** (this lane's `bugs_open/339` seam fix, commit `e3dee9243`, council APPROVED
  round 1 corr `2ff9e215` at 17:18Z): tool pages' meta description is COMPOSED-ONLY — both creators
  pass the empty candidate. **Demand proof owed and scheduled: on #24's build, read
  `pages.meta_description` for the social-card page — it MUST be the composed sentence ("An
  interactive Social Card Studio, free to run in the browser…"), never the item's brief. Record the
  result in NOTES and TL-048's verify-later.** (The old ported page's meta was human-shaped and
  ≤320, so the DIFFERENCE to look for is: not the brief.)
- **TL-047's gate commit `c0d60a97d` is ALSO in 1321** — the `replace_existing` regenerate-in-place
  path now waits ONLY on seed `496_tool_generator_replace_existing_HOLD.sql` (the 331 lane's; close
  order in commit `138c8efaa`). Once applied, RE-FIXES need no deactivate/rename/retire race.
  **First queued re-fix: the oklch CSS fallback-order defect** (NOTES 17:05Z — the emitted pair is
  `oklch` then hex, so the hex wins everywhere and the oklch line is dead; correct order is hex
  first). Do NOT hand-edit the component.
- RFC_036 §9.3 (`e24bc9c0f`) and the 303 fix (`6d962bcf8`) remain ancestors — nothing regressed.

## What is LEFT (40 ported slots), by phase

- **Phase A remainder — 4 simple tools** after #24: `tool-regex-tester` (7,609 — the richest defect
  set: the regex runs against HTML-ESCAPED text, dead draft code ships with the author's inner
  monologue, hardcoded result hexes, no match count; analysis in NOTES-adjacent prep 2026-08-20),
  `tool-jwt-inspector` (7,645 — injects a fabricated `_human_exp` field INTO the decoded payload;
  stale output on cleared input; no expired-verdict; must state it decodes-not-verifies),
  `tool-token-calculator` (7,837 — 2024-era models/prices presented as current in 2026; estimate
  disclaimer buried in a comment), `tool-clip-path` (7,909 — mouse-only, dead on every phone;
  edge-clicks self-intersect the polygon).
- **Phase C — 13 external-`<script src>` tools**: brief must come from browser behaviour; the S3
  asset must be retired with the slot (TL-032).
- **Phase B — the ≥8 KB tools (~22)** incl. the FIVE rich apps (mind-map, meme-generator, logic
  architect, micro-CMS, pasteboard): owner's standing instruction — LAST, one at a time, each seen
  at the served page by a person. `tool-meme-generator`'s library claim `6ae53f32…` is handled by
  the live §9.3 fork path.
- **Done means 63/63, each graded at the served bytes with a cache-buster and old-only negative
  controls.** Nothing is platform-blocked any more.

## Decisions that belong to the OWNER (none block Phase A/C)

1. **When to spend review attention on the five Phase B rich apps** — each needs a person at the
   served page by your own ruling; the lane can sequence everything else first (recommended).
2. **Whether/when to apply seed 496** (arms `replace_existing`; the 331 lane holds it deliberately —
   their close order says roll first, and the roll has now happened). It unlocks race-free re-fixes,
   starting with the oklch fallback order.
3. **The "tool asserts something untrue about itself" class (10+ sightings)** still has no
   mechanism: is it a Tier-2 acceptance criterion, a checker, or a standing human step in every
   brief (the current answer, which has caught all ten)? The lane recommends: leave it human until
   Phase A/C are done — the rebuild reads every script anyway — then decide with the full corpus.
4. **Track 2 (rules 16+17 sub-checks in `check_tool_health.go`)** — small, decided shape (superseded
   handoff §Track 2), deferred because the actions package carries other lanes' WIP and every rebuild
   already audits by construction. Build it when the grind stops, or assign it to a quiet lane.

## Open threads (not blockers)

- **Serve-grades**: all 23 confirmed; #24 onward each needs its own after its rerender.
- **`bugs_open/339` residue is NOT ours**: the 12-row live repair + the growing NON-tool writer
  sub-class belong to the `meta_description_never_backfilled` lane (split recorded in 339 §7b).
- **Council coverage**: commit `e3dee9243` carries `Council-Submitted: 2ff9e215…` — verdict is
  APPROVED, 098 credits it automatically; nothing owed.
- **Site-wide rerender sweeps run ~every 2h** (~119 rows, one per page, drain ~0.7/min alphabetical)
  — the margin gate in the RUNBOOK's filing section handles them; `ahead=0` is AMBIGUOUS (nothing
  queued vs next-in-line) — disambiguate by querying the page's open rerenders directly.
- The lane's briefs are 800–1,100 chars, so even pre-TL-048 they drew the composed fallback via the
  length signal; the 200–320 window was CLOSED by TL-048, not narrowed.

## Traps (additions since 2026-08-19; older ones in the superseded handoff still hold)

- **If this lane ever adopts a `?` optional-explicit config wire** (credited to the
  staged-component-build lane, 2026-08-21): (1) the wire and its entry in
  `architecture_review/optional_explicit_wire_acks.json` must travel in the SAME commit, or the
  `config-key-audit --optional-explicit-wires` gate goes red on someone else's clock; (2) the `?`
  marker only parses on chassis ≥ v1.0.1321 — on an earlier binary the migration applies cleanly and
  the field silently falls back to the whole-tree search (matters for ROLLBACKS too; LANDMINES has
  both version numbers). Also: an acknowledgement in that file is the WIRE AUTHOR'S claim — never
  confirm one for a wire you did not write (this lane declined exactly that request, 2026-08-21).
  ⚠ And the gate is NOT yet a CronJob (2026-08-21): `--optional-explicit-wires --report` runs by
  HAND only, so a missing ack will not be caught for you. One more source-of-truth rule from the
  same exchange: **a migration's own snapshot row (`agent_definitions_backup`, written in its
  transaction) is the load-bearing timestamp — `agent_definitions.updated_at` is only the LAST
  write and lies the moment anything else touches the row.**

- **A watcher's FIRING is not its DELIVERY** — see the rewritten recipe rule. Two same-day
  near-misses and one 6-hour public double-tool page are one mechanism.
- **`already_exists=true` on the LATEST orchestration row does not mean nothing was built** —
  a Kafka response failure (bugfix-040 class) makes the retry short-circuit against run 1's own
  component. Enumerate ALL runs in the item's window; and `error` text persists on rows that later
  complete (the 336 file's census trap, seen live on this lane's own #22).
- **Old ids shared by both versions are not negative controls** — `cssOutput`, `layersContainer`,
  `verdict`, `meshCanvas` have all now collided; ALWAYS grep the NEW template for each candidate.
- **The register index count and TL numbers move under you** — TL-044/447 and TL-046 were both
  renumbered mid-write by concurrent lanes; grep for your number before writing it.
