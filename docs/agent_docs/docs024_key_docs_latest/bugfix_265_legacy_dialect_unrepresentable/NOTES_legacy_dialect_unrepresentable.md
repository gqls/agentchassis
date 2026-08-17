# NOTES — bug 265, legacy `input_schema` dialect made unrepresentable (append-only, newest at the bottom)

## 2026-08-16 — session 1

**Bug selection.** Session was named for 213 and 274; both closed at HEAD (`git ls-tree HEAD --
bugs_closed/` shows both; 213 by owner ruling 08-15, 274 live+proven). Census of `bugs_open/`:
last-commit date per file, edits by live transcripts (grep `"file_path":…bugs_open/NNN` in
`~/.claude/projects/…/*.jsonl` modified <36 h), header status keywords, `who-owns.py`. 265 was
the one open, unowned (no workstream dir; one live *Read* of the helper by the 260 lane, no
edits), platform-level candidate. 207/217 also surfaced as fixed-live-proven but still in
`bugs_open/` — closure housekeeping for their lane, not taken here.

**Re-verification** — see the bug file's 2026-08-16 section for the table. Two things moved:
the count is 3 not 4 (LMC converted loans-consolidation on 08-15), and the producer is hand
SQL, not the component-creator. The second is the decisive one for the design.

**Misstep avoided (recorded because it was close):** my first instinct was to write the
birth-path Go check as *the* fix — that is what the bug file's candidate 3 pointed at. One
`GROUP BY created_from, source_agent_type` showed the component-creator has never emitted the
dialect and every legacy row was `manual`. Had I not run it, the "fix" would have guarded a door
none of the four rows came through, and the next hand seed would have walked straight in. The
check that caught it: **read the provenance columns before choosing the seam.**

**Design decisions and reasons** — in the PLAN. Short form: CHECK constraint (the seam every
producer crosses) + behaviour-preserving conversion of the 3 rows in the same migration + Go
birth check for legibility + comment/tripwire rewritten to cite the constraint. Not: a CronJob
(constraint doesn't drift), deleting the reader's projection (fail-open again), rewriting
applied seeds.

**Probe run of 437** — COMMIT swapped for `SELECT …; ROLLBACK;` via sed into the scratchpad;
`psql -v ON_ERROR_STOP=1`. Output: guards passed, `UPDATE 3`, `NOTICE: 437: 3 legacy rows
converted…`, ALTER, COMMENT, then rollback. Live table read back: legacy 3, constraint absent,
backup absent. Converted rows captured → Go fixture. First attempt of the probe printed 11
lines (psql command tags + 3 rows); filtered `^{` for the fixture.

**Tests** — `TestLegacyDialectConversionMatchesProjection` (3 rows) and
`TestIsLegacyInputSchemaDialect` (6 shapes) green. Mutations: fixture minus report-dossier's
`required` → FAIL "projection mismatch"; predicate `return false` → FAIL on `legacy` and
`fields and properties`. Restored; `git diff` on the helper checked to be only the intended
hunks (the mutation restore was a first-occurrence replace and could have hit the wrong
function — it did not, verified by reading the func body).

**Same-file passenger.** `store_generated_component_action.go` was already ` M` at session
start: one hunk at ~:890 (`WHERE status = 'active'` → `PageWantedLivePredicateFor("")`), the
185 lane's, uncommitted. The helper exists at HEAD (`git grep … HEAD` hit
`datahelpers/links.go:361`), so it compiles. Named in the commit message per the pathspec
passenger practice rather than reverting another session's WIP in the tree.

**Council** — submitted `aba82416-de79-4452-8730-3e35ca0a15bb`, 6 edits, 11 grounded_in.
Verdict pending at time of writing; commit carries `Council-Submitted:`.

**`go vet ./platform/orchestration/actions/`** reports one pre-existing "unreachable code" in
`load_component_library_actions.go:207` — not mine, not touched.

**Two content questions surfaced for other lanes, not fixed:** report-dossier `body` marked
`source: llm` (its seed says never LLM-authored) — gripper-dossier lane's call; `site-header`
carries v2 field defs with no `fields` wrapper — a third shape the reader cannot see, harmless
today. Both in the bug file.

**10:22Z verdict:** APPROVED round 1, all reviewers. **10:24Z applied 437 by hand** (`psql -v
ON_ERROR_STOP=1 < file`): `UPDATE 3`, NOTICE, ALTER, COMMENT, COMMIT. Read back: census 0,
constraint present, three rows `fields`=t. Induced: legacy INSERT → `violates check constraint
"chk_input_schema_no_legacy_dialect"`, rolled back, scratch count 0. Recorded `--record-only`
with the note. Register CLC-015 + bug STATUS updated to "constraint LIVE, Go awaits roll".
Runner dry-run skipped (took >110 s this session — many pending files from other lanes).

## 2026-08-16 (afternoon) — CLOSED

Roll to **v1.0.1304** at 10:41Z. `build provenance` line already out of `--tail` range on
agent-chassis (the standing landmine — it is a STARTUP line and the service is busy), so the
binary was probed instead, with controls: stamp `5de6cddbe` present on both replicas, new
literal present, **the pre-fix tripwire sentence ABSENT** (the negative control that makes it
evidence rather than a coincidental match), and `58b0111ac` confirmed an ancestor of the stamp.
Constraint re-verified live; refusal induced earlier; **demand control** — two real component
writes (`tool-ab-test-calculator` 12:17Z via tool-deployer, `tool-css-specificity-calculator`
12:19Z via tool-generator) passed the live constraint, so it is on the write path and does not
refuse legitimate writes. Tripwire firings since the roll: 0. Bug moved to `bugs_closed/`.

**MISSTEP, caught before committing (WRONG_CALLS row filed).** Repointing `bugs_open/265` →
`bugs_closed/265` in code comments is right for source files — it is the exact stale-pointer
trap this very change corrected for `026`. I applied the same sed to
`sql_for_agents/437_*.sql`, which was **already applied**, and `schema_migrations` stores an
**md5 checksum** of the file as applied. A comment-only edit would have drifted it and made a
correctly-applied migration read as tampered-with. Caught by asking what columns the ledger has
before trusting the edit; reverted, and `md5sum` now matches the recorded
`601f395002755ff05fa915f400dce8fe` exactly. The general form: **an applied migration is
history, and "fix the stale pointer" is a habit for source, not for history.**

**PASSENGERS in `2079aca42` (WRONG_CALLS.md), named so their authors can find them.** My commit
added 124 lines where my own row was ~30 — the contended-ledger tell. Two other sessions'
entries were sitting uncommitted in the file and went in under my message:

- *"a KNOWN-but-tool-level name vanishes while an unknown name routes to creation ⇒ there is a
  second branch"*
- *"I put a subagent's fleet counts in a commit message without grounding them, and they were
  10% stale"*

Nothing is lost and forward-only holds — this is a review/bisect problem, not a data one. What
caught it was reading the `--numstat` instead of trusting the pathspec: **"I followed the rule"
is not "I checked the result."** `WRONG_CALLS.md` and `LANDMINES.md` take appends from several
sessions a day, so the same-file case is the normal case there, not the exotic one.

## 2026-08-17 — the roll-blocked sweep, and 093 closed

**Session spanned midnight.** Everything dated 2026-08-16 above was correct at the time; it is
now 08-17. The "fresh build" is **v1.0.1305**, rolled 2026-08-16 22:07Z.

**265 regression check first:** its literal is present on both v1.0.1305 replicas and the
pre-fix tripwire sentence is still absent. The fix survived the roll — worth checking, because
one release tag can ship several source revisions (`bugs_open/249`).

**Build stamp: `6a782274b`** (chassis v1.0.1305, both replicas). The provenance log line had
scrolled again. Another lane had just published the same stamp for reasoning-agent; I verified
it on the chassis myself rather than inheriting it, because the stamp is per-SERVICE.

> **⚠ MY NEGATIVE CONTROL WAS WORTHLESS AND SAID SO — 40 zeros are PRESENT in any binary.**
> I probed `grep -a "0000…0" /proc/1/exe` as the must-be-absent control and it came back
> PRESENT. The standing landmine warns against a *discovery* grep for "some 40-hex string"
> (it matches Go's digit table); this is the adjacent trap: **a fabricated needle must be
> fabricated-but-PLAUSIBLE.** Zero-filled runs occur naturally in a binary's data. Replaced it
> with two real commit shas that are not the stamp — both absent, so the probe discriminates.
> Had I used only the zeros control I would have concluded the probe was unreliable and thrown
> away a sound measurement.

**The sweep.** Regenerated the roll-blocked candidate list and screened it by owning-lane
activity: of ~30 candidates, **exactly one** (`093`) had a quiet lane; everything else is
actively owned, so contribute-don't-compete applies and I left them alone.

**093 CLOSED** — and it was not roll-blocked at all. Its fix (`72effdbca`) shipped on
2026-07-26, the *same day* the file's last section declared the gap unchanged; the two crossed
and nobody re-read it for three weeks. Proof it runs, not merely ships: an `unregistered_stat`
finding with pattern `case-studies-grid.card1_stat_value` (component.field ⇒ stored
`content_data`), and `ItemType: "claims_unverified"` exists in exactly one place fleet-wide so
the build-time gate cannot have produced it. Full account in `bugs_closed/093`.

**Near-miss on the way:** I first grepped for callers of `scanStoredStatClaims` *within*
`discovery_checks/` while excluding the stats file, found only test callers, and was one step
from writing "a helper with no callers". Widening to the repo found two real call sites. The
narrow grep was my own construction, not a tool defect.

**NEGATIVE RESULT — I did not build the detector this suggested, on purpose.** Three bugs in
one day were already-fixed-but-open (145, 072, 093), so the obvious move was a check for
"bug file is stale". I measured the candidate signal first: *commits naming bug N since file N
was last touched*. It ranks `210` (32), `252` (23), `203` (20) at the top — and those are the
**most active lanes**, not the stalest files. The signal measures activity and would misfire
loudest on the work that most needs to be left alone; a guard that refuses good work gets
switched off, and then it protects nothing (`component_write_guard.go`'s own calibration note).
So: the duplicate detector I did ship is precise (a filename in two directories is
unambiguous); a "fix already shipped" detector is not cheaply precise, and 093 was found by
reading, not by a signal. Recorded so the next session does not rebuild it and reach the same
wall.
