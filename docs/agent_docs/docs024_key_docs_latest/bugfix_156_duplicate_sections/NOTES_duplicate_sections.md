# NOTES — bugfix 156, the prevention half

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — picking the bug, and how many were already taken

The fleet is busy: 35+ session transcripts touched in the last three hours. Choosing an
unowned bug was most of the opening cost, and `who-owns.py` alone was not enough, because it
reads **commits** and a session mid-fix is invisible to it.

What I actually did: greped every `.jsonl` transcript modified in the last 180 minutes for
`bugs_open/NNN` and counted. That is the check the memory line "EVERY ownership check is
LAGGING — grep live `.jsonl` transcripts" describes, and it changed the answer three times:

- **093** (stat audit, one guarded call site) looked ideal — unstarted, structural, council
  verdict already written. Rejected: its own post-roll triage says it is **blocked on
  `bugs_open/083`**, the check shipped and has never run because `improvement-sweep` is off,
  and 083 has an active thread (11 mentions in a live transcript). Fixing 093 today would
  produce another correct mechanism behind something that never runs.
- **071** (the gate detects every broken link then discards it) is the highest-value bug in
  the directory and `who-owns.py` returns **OWNED or recently active** on it, naming two live
  workstreams. Not mine to start a competing fix on.
- **156** — zero mentions across every live transcript, last commit 2026-07-31, and the file
  itself says of the remaining half: *"What remains, and is unowned."*

## The premise moved between filing and today, and in the direction that matters

The bug file's census: 17 duplicate `(page_id, slot_name)` groups, 6 of them content-identical
(vonc). Re-run today: **12 groups, 0 content-identical.** The vonc rows were fixed on 07-30.

So the honest framing of this lane changed before I wrote a line: **there is no damage to
repair.** The guard is judged entirely on whether it could ever destroy a live section, and
its value is preventing a recurrence of one measured incident. I would rather state that than
let a reader infer the guard is cleaning something up.

## The bug file's own fix candidate is wrong, and its own footnote says why

`156` candidate 1 proposes collapsing on `(slot_name, md5(content_data))`. Its own census
footnote records that finetuning.uk/our-position-on-ai has two rows whose `content_data` is
**NULL on both**. Under the proposed key those two are "identical" — so shipping the file's
literal recommendation would have deleted a live section on a shape the file itself flags,
one paragraph apart. The two halves were written for different purposes (the footnote warns a
future *measurer*, the candidate instructs a future *fixer*) and nothing joined them up.

Corrected rule, and the reason it cannot over-collapse: collapse only when **every column the
INSERT binds is equal, `position` excluded**. Then the collapsed row would have been
indistinguishable from its survivor in the database. `rendered_html` in the key is what fixes
the NULL-content case; `component_id` in the key is free narrowing.

## What the planning pass found that I had not

Ran the plan through `fable`. Three things came back that I had missed on my own read, and
one correction to my own brief:

1. **My brief said `save_page_sections_action.go` was dirty in the shared tree** (the 194
   lane's WIP) and I designed around a same-file collision. By the time the plan came back
   that session had committed and the file was clean. Stale-by-minutes, exactly as CLAUDE.md
   warns about session-start `git status`. The sibling-file shape is still right — it is the
   house pattern — but the collision argument I built it on had evaporated. Re-checked before
   editing rather than trusting either reading.
2. **The locked-slot path manufactures a duplicate of locked copy.** With a doubled list the
   first copy of a locked slot consumes the lock and is discarded (`lr.consumed = true`); the
   **second copy falls through and is INSERTed beside the locked row.** Nobody had recorded
   that; it is an independent argument for collapsing before the insert loop rather than
   relying on a post-hoc detector.
3. **A doubled list halves the content-regression guard and doubles the completeness floor's
   numerator.** Both guards are currently measuring a lie whenever this defect fires: a page
   truly cut to 13% of its text reads as 26% and passes the 25% floor. So the placement is
   not a style choice — it is what makes four other guards measure the truth.
4. **Plan-guard parity.** The repair half refuses to delete a repetition the effective plan
   specifies (council-forced, owner decision 07-31). A save-time collapse that ignores the
   plan would make the two halves disagree about the same question. Same helper, same per-slot
   accounting — but the **opposite failure direction**, because the repair's conservative
   move is "refuse to delete" and a collapse guard's is "refuse to collapse".

---

## 2026-08-04 (later) — implementation

Written as `save_sections_dedup.go` (pure seam + DB seam + record) with the call site at
`save_page_sections_action.go` immediately after the arrival diagnostic. Test file carries a
named mutation per test in its header, because a test that could not have come out otherwise
is not evidence.
