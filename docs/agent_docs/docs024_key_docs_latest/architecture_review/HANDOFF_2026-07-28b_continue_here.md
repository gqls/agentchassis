# HANDOFF — architecture seat, continue here (2026-07-28, evening)

**COLD-START ENTRY POINT. Supersedes `HANDOFF_2026-07-28_continue_here.md`**, which
is still correct about layer 1 but predates the whole layer-1b council sequence.
Go back to it for the layer-1 verification recipe; go to
`HANDOFF_2026-07-27_continue_here.md` §5 and `…-27b` §5 for the Go-contract
landmines, which are unchanged and still the most expensive in this directory.

Prose: `SUMMARY_2026-07-28_the_seat_can_see.md` · `README_where_we_are.md`
(owner's log, append-only) · `NOTES_architecture_seat.md` (technical log + every
misstep) · `RUNBOOK_architecture_seat.md` · `DECISIONS_…_architecture_seat.md`.

---

## 1. State — everything below is verified, not remembered

| thing | state |
|---|---|
| D11 layer 2 (routing) | **LIVE** |
| **D11 layer 1 (symbol bodies)** | **LIVE & PROVEN** — 4,992 rows / **4,992 bodies**, `ref=086_experience_loop`, survived the roll to `v1.0.1194` |
| `bugs_open/108` | **CLOSED** — both defects fixed and live |
| Index freshness | keys on the indexed **commit**, not the row clock (108 defect A, live) |
| Indexer ref | **PINNED** by migration 252 to `086_experience_loop` |
| `review_architecture` | **still 0 reviews** — rate limit, not fault |
| **D11 layer 1b (markdown)** | **IN COUNCIL, round 7 in flight** — see §2 |
| **`bugs_open/135`** | **NEW, filed today, unowned** — see §3 |

**⚠ REVERSAL TRIGGER, do not lose:** migration 252 pins the indexer's ref to
`086_experience_loop`. **Change that literal to `'main'` AS PART OF the merge.**
A no-rows `pre_query` makes the scheduler SKIP the task — it does not fall back —
so the refresh would stop silently.

## 2. D11 layer 1b — markdown into the index. Round 7 in flight.

`SUBMISSION_CORR=7ba5b8c4-0e10-46db-9fc4-2bd0584e943a`, plan committed at
`SUBMISSION_2026-07-28_markdown_into_the_index.json`. Check it:

```sql
SELECT created_at, metadata->>'decision', metadata->>'unreadable'
FROM diagnosis_artifacts WHERE correlation_id='7ba5b8c4-0e10-46db-9fc4-2bd0584e943a'
  AND kind='council_report' ORDER BY created_at;
```

**If APPROVED**: build it, commit with `Council-Reviewed: 7ba5b8c4-…` on the
IMPLEMENTATION commit (never on docs), then image → roll → pod-grep → reindex →
run the VERIFY. **If REVISE**: read §4 before touching anything.

**What it does:** relax `code_symbols_kind_check` to admit `'doc'`; index one row
per markdown heading-section (`symbol`=heading, `body`=section text,
`content`=heading+path only — the body must stay out of `content_hash` or every
row re-embeds); from **four durable sources only**.

**The globs ARE the design** — widening them is the failure mode, not a tuning
knob. `docs024` holds 1,415 `.md` files that are superseded by design; a seat
citing a stale handoff as evidence is precisely what this workstream exists to
prevent.

```
bugs_open/*.md · bugs_closed/*.md · WRONG_CALLS.md ·
016b_debugging_guide_8_consolidated.md · docs026_concept_register/register/*.md
```

**Sizing: ~3,472 sections against 4,992 code rows (+70%).** NOT the ~1,749/+35%
quoted in rounds 1–4 — that predated the register-glob fix and counted none of the
110 register files.

## 3. `bugs_open/135` — the prune has no floor (filed today, UNOWNED)

`DELETE FROM code_symbols WHERE repo=$1 AND commit_sha IS DISTINCT FROM $2` runs
unconditionally at the end of every index. A partial fetch, moved directory or
permissions change yields a small-but-nonzero run that upserts what it saw and
**deletes everything else** — and reports `pruned: 4000` as a *success counter*.

Found by `bug_historian` reading the call site during review of the markdown plan.
**Pre-existing and independent of that plan.** The file carries the design six
council rounds produced, including the hard-won negative: **`site_work_items`
cannot be the durable surface** (`site_id`/`handler_agent`/`pipeline` are all
`NOT NULL`; this indexer is repo-wide) — `doc_notes` fits.

## 4. THE LESSON OF THIS SEQUENCE — read before resubmitting anything

Six rounds went **5/3 → 5/3 → 8/2 → 10/1 → 6/4 → 9/3**. It stopped converging at
round 4, and the diagnosis matters more than the plan:

> **ALL SIX of round 6's objections were about a helper I invented in round 6.
> None was about markdown indexing, which had drawn no objection since round 4.**

A prune-safety sub-feature **accreted in response to review**: objection → I added
a floor → the floor needed a durable surface → the surface needed a helper → the
helper drew six objections. **Each fix added mechanism, and each new mechanism
became fresh surface to object to.** Round 7 splits it out (§3) and resubmits the
core alone.

**The general rule, worth applying beyond this workstream: when a plan's
objections migrate off its subject and onto machinery added during review, the
plan has accreted a second change. Split it — do not iterate it.** Rounds spent on
the accretion are rounds not spent on the thing you came to do, and the accretion
is usually someone else's bug that review surfaced.

## 5. Landmines earned today

- **The council caught three things that would have SHIPPED BROKEN**: the register
  glob pointed at 28 superseded planning docs and missed all 110 register entries;
  a prune floor that guarded only the new rows; and a "durable surface" that was
  pseudocode for an **impossible row**. Take REVISE seriously — see §6.
- **`site_work_items` is site-scoped and cannot take a repo-wide row.**
  `site_id`/`handler_agent`/`pipeline` are `NOT NULL`. `doc_notes` has a nullable
  `site_id` and is append-only (no dedup contract).
- **A remembered figure is a stale figure.** `~1,749` was quoted in four
  consecutive rounds and was wrong by ~2×. It was caught by a *low*-severity
  objection about **methodology**, not about the number.
- **`^func …` greps cannot see methods with receivers.** A prior-art search
  anchored that way has a hole exactly where a stateful helper would live. The
  code index itself is the fix — it stores methods as `(*Type).Method`.
- **`editquality` went `unreadable` in rounds 2–4 and HUNG for 2h34m** in one
  attempt (`bugs_open/119`). When it finally parsed in round 5 it raised the
  first HIGH of the sequence. **A missing seat is invisible in the verdict —
  always read `unreadable`, never `abstained`.**
- **A wedged council run looks exactly like a queued one.** The exit condition is
  the peer check: if runs created *after* yours have COMPLETED, resubmit. That
  and `unreadable` are the two things to check before rewriting a plan.

## 6. How to read a REVISE, given the measurements

Grounded 2026-07-28 over 7 days: **per ROUND 51% approve; per SUBMISSION 76%**
(78 of 103 correlations eventually approved). CLAUDE.md's "~80%" is the
per-submission figure and is sound; read as per-round it badly misleads.
**A REVISE or two is the median path, not a signal the plan is bad.**

## 7. Still open, unchanged

1. **The seat still has 0 reviews.** `review_architecture` sits only on
   `feature-designer`, which refuses anything without an owner-approved
   `capability_gap` spec. Of 6 such items only one has both `owner_approval` and
   `code_pointers`, and it belongs to another thread. **Do NOT manufacture a
   review by firing at someone else's ticket.**
2. **`council-gate` still gets no code answers** — deliberate (it has no
   reproposer); the fix is surfacing code results into its verdict note.
3. **D11 layer 3** — a seat cannot look things up *while reasoning*, only next
   round. `[UNSCOPED]`, wants its own RFC.
4. **Two orphaned council orchestrations from ~16:55 today** never recovered
   across a roll — recorded `[UNDIAGNOSED]` in NOTES, not chased.
