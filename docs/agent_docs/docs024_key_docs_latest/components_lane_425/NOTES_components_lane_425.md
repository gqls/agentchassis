# NOTES — components lane (`bugs_open/425`)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

> **Created 2026-09-03 15:10Z, late.** This lane ran for two days on a single
> `HANDOFF_..._continue_here.md` doing the work of all five standing documents, which is why its
> corrections had to be written as struck-through banners inside the handoff rather than as dated
> entries here. **PLAN was never created and is not being back-filled** — a plan written after the
> decisions is a report, and the decisions with their reasons are already in the handoff's §0
> series where they were made. RUNBOOK, NOTES, README and SUMMARY now exist.

---

## 2026-09-03, afternoon session — the open defect was answered, and by someone else's bug

**Picked up:** the handoff's cold-start told me to read batches `…000691` then `…000690`, and named
the open defect as "the producer fix does not execute on *some* rerenders — 16 candidates
eliminated, model refuted".

**What I found before running anything.** Reading `rerender_page_sections_action.go` alongside the
handoff's §0h, the file's own comment on `c.plan = plan` names the defect outright: *"Compute it
here and drop it, and every re-render silently renders the page's OWN STORED DATA back at itself…
Pinned by `rerender_page_sections_resolved_data_test.go`."* That comment is dated 2026-09-03 and
cites `bugs_open/454`, filed by the `bugs_open/427` lane at 11:00Z from the far end of the same
mechanism. **The handoff mentioned `454` zero times** — the two lanes were working one cause from
opposite ends and the newer file had already named it.

**So the first useful act was not an experiment.** It was `grep -n 'c.plan' <file>` and reading the
one bug file that the code comment pointed at.

**What I then did, and why it was still worth doing.** `454` was proven on an `event-list`
component whose save was refused by the `450` tool-page guard, and on a designblog deck that
already carried the key. Neither is a clean positive for **this** lane's class. So I filed batch
`…000692` on garden-tools.uk `/care` — the handoff's own "tightest pair" partner, and the one page
in the pair whose stored deck was **OLD shape**. Complete 14:05:11Z: `excerpt` absent → present,
site-name suffix stripped, 4 excerpt elements with none empty, live at the served bytes by
14:05:05Z. First run, unambiguous.

**MISSTEP I inherited and then had to unpick — and it is the lane's most expensive claim.** The
handoff's §0i, headed *"THE PATH-SPLIT MODEL IS REFUTED — a rerender DID produce the new shape"*,
is **itself wrong**. Its two cited re-renders wrote rows that carry `excerpt`, but the state each
of them REPLACED already carried it too: a BUILD (`empty_section` / page-build-handler) put the key
there on 09-02 20:51:05 (designblog) and 23:02:20 (websitepromotion), and every later re-render
merged `stored ⊕ nil` and carried it forward. The disconfirming column was in the same
`page_component_history` row the original claim was built from. Consequence: a correct model (§2's
build-resolves/re-render-does-not) was retracted, and sixteen eliminations were then measured
across 17 instances hunting a per-instance difference that does not exist. Logged in
`WRONG_CALLS.md` and corrected in place in the handoff.

**MISSTEP of my own, caught at the query rather than after it.** My first read of the two queued
batches was `WHERE id IN ('00000000-...-688', ...)` and returned zero rows — those are **batch**
ids, not item ids. A zero-row result there reads exactly like "the items are gone", which is the
same false-absence shape this estate keeps paying for. Fixed by querying `batch_id`; recorded in
the RUNBOOK next to the command.

**Two dispatches were spent before I arrived, and the reason generalises.** `…000688` and
`…000691` both ran on baselines that **already carried** `excerpt`. On a path that persists
`stored ⊕ fresh`, a preserved value and a re-resolved one are byte-identical, so both hypotheses
predict "key present" and the dispatch cannot come out any other way. §0k had diagnosed exactly
this after `688` ran; `691` was then filed with the same flaw, justified because it ran on the page
that had reproduced the defect four times. **The page was never the axis.** Written up as a
LANDMINE, because it fires when you *choose a target* — before there is any symptom to warn you.

**Three confirmations I did not have to arrange.** Ordinary fleet traffic repaired old-shape decks
while I was writing: dartsonline `/guides-index` 13:55:46 and `/index` 13:56:32, both
`section_data_resolved`, both filed by `image-build-handler`, both old → NEW. Those are better
evidence than my own dispatch precisely because nobody chose them. The class went 5/17 new-shape
yesterday to **9/17 at 15:05Z**, and it is draining without a wave.

**A count that goes stale DOWNWARD.** Every staleness warning in this estate is about addition. This
lane's central census falls as the fleet re-renders, so a figure quoted from a document overstates
the remaining damage. Same rule, opposite sign: re-run it.

**A peer's sweep resolved a question I had answered wrongly.** `site_delivery_and_editor` measured
boxingonline `/articles/index.html` serving 14 empty `article-card__category` and 2 empty
`article-card__excerpt`, and asked whether it was a second component, an unguarded template, or a
data gap. It is `bugs_open/457`: six orphan `page_components` rows, each holding its own frozen
`rendered_html`, all assembled into one page. Summing per-row counts gives **14 and 2** — both axes
match their served figures, from two instruments sharing no code. **I had told them 425's class
"had not reached" that page; it never will, because the class was never on it.** And `454`'s fix
structurally cannot repair those rows: `component_id` is NULL, so `resolveComponent` misses and
each takes a carry branch. Corrected in `bugs_open/457` and in the handoff's new §0q.

**Could not verify one thing and am not going to pretend otherwise:** `boxingonline.com` has no A
record from this machine, so `probe-page-url.sh` fails its own sibling control and refuses to
answer. Every served reading for that domain in today's write-ups is the peer lane's, attributed.
My independent instrument on those pages is `page_components.rendered_html`, and where the two can
be compared they agree exactly.

---

## 2026-09-03 15:20Z — both canaries confirmed, and a commit trap I walked into

**Both `…000693` canaries came back CONFIRMED and did not disagree**, which is worth recording
because this lane's standing rule is that they do. `26be9662` robot-hands.com
`/learning-center-hub` (complete 15:14:51Z) and `75424e19` homegarden.uk `/comparisons-index`
(complete 15:16:01Z): `excerpt` absent → present, site-name suffix stripped, 8 and 3 elements
respectively with none empty, each attributed by `source_item_id` with the before-state projected
from the same history row. So the repair holds across three sites, three page types, item counts of
3/4/8, and a baseline nine days stale.

**One thing to carry:** robot-hands' `rendered_html` **shrank** 7,681 → 6,753 B on a successful
repair — the stripped suffixes and the collapsed empty elements remove more than the excerpt text
adds — and it did not trip the section shrink floor. A byte decrease on this repair is not a
regression signal, and anyone building an acceptance check on "html grew" would have filed it as one.

**Census, and it moves fast enough to matter:** 5/17 new-shape yesterday, 9/17 at 15:05Z, **11/17 at
15:18:56Z**. Thirteen minutes for two instances. Anyone quoting a figure from these documents
overstates the remaining damage.

### MISSTEP — my pathspec commit deleted another lane's bug file out from under itself

**What happened.** At ~16:09 local I appended a CONTRIB to `bugs_open/454_HANDOFF_…md`. At 16:13:36
the `bugs_open/427` lane ran `git mv` on that file to `bugs_closed/`, closing the bug. At 16:13 my
commit ran with the explicit pathspec CLAUDE.md mandates, naming the **old** path. A pathspec commit
reads the working tree at the paths it names and ignores the index — the file was not there any
more, so git recorded a **477-line deletion with no corresponding add**. `git ls-tree -r HEAD |
grep 454_HANDOFF` returned **zero rows**: the bug existed at neither path.

**Why this is worse than the documented version of the trap.** `LANDMINES` already carries this
collision twice, both times from the **mover's** side, where the file ends up in **both**
directories — which overstates the backlog and is caught by `check_bug_file_duplicated`. From the
**editor's** side it ends up in **neither**, so the estate's most-repeated instruction — grep
`bugs_open/` and `bugs_closed/` before you file — silently returns nothing, and a live bug reads as
never filed. Nothing checks for that.

**What I did, and the judgement I want on the record.** I did *not* commit the restoration myself,
even though it would have taken one command. The file had been written 60 seconds earlier, so the
moving session was plainly mid-close: committing it would have taken a same-file passenger of their
in-flight closure text *and* put their close decision under my commit message. I messaged them with
the exact trace instead. Their `git mv` add was already staged, so their next commit restored HEAD
by itself, within two minutes. They confirmed the trace from their end and folded both of my
corrections into the closed file as visible `CORRECTED` blocks rather than silent fixes.

**What caught it.** The commit's own output — `delete mode 100644 bugs_open/454_…` and `488
deletions(-)` against an expected ~490 insertions. I read the deletion count, did not recognise it,
and stopped. **The yellow commit-scope block did not and could not help**: it listed the file under
`bugs_open/` as part of my own area, which is exactly what it should say. The tell was the word
`delete`, in a line I could easily have scrolled past.

**Also resolved a numbering collision I caused:** we had each written a `## 15.` into that file
within minutes. Theirs is the closure note and is canonical; I renumbered **mine** to `§15a` with a
line saying so, and did not touch their text.

Written up as a LANDMINE with the two-command check (read your own `git show --stat` for
`delete mode`, then ask `git ls-tree -r HEAD` — exactly one row is correct, zero means this fired,
two means the documented sibling fired) and the prophylactic that actually shortens the window:
**commit an append to a bug file promptly and on its own, especially when that bug's fix has just
gone live and someone is verifying it.**

**COORDINATION NOTE on the same incident, 15:25Z.** The `bugs_open/427` lane and I each wrote a
LANDMINES entry for that commit trap, independently, **two minutes apart** — theirs `6653293ee` at
16:19 local, mine `ec0f5b1e2` at 16:21. I did grep `pathspec` before writing and found the three
existing entries; theirs had not landed yet. So "grep before you file" worked exactly as intended
and still produced a duplicate, because the window was shorter than the write. The two entries are
complementary rather than redundant — theirs carries the recovery command
(`git show <deleting-commit>^:<old-path>`) and the mover-side norm, mine carries the
absent-from-both-directories consequence and the prophylactic — so I cross-referenced mine to
theirs naming what each half holds, rather than deleting or rewriting anyone's text. **The general
point: after a shared incident, expect the other party to be writing it up at the same moment, and
cross-reference rather than re-grep.** Same shape as the `437`/`425` mutual-invisibility note
already in the handoff.

---

## 2026-09-03 15:45Z — CLASS COMPLETE (17/17), and a rule of mine over-flagged by 6×

**Batch `…000694` drained in six minutes and the class is finished.** All six repaired, each with
`excerpt` absent → present confirmed from the archiving write, all with 0 empty elements:

| page | complete | items | html |
|---|---|---|---|
| idea.uk `/guides-index` | 15:23:02 | 7 → **9** | 6,987 B |
| homegarden `/this-month-index` | 15:23:50 | 3 | 3,062 B |
| homegarden `/home-maintenance-index` | 15:24:26 | 3 | 3,127 B |
| homegarden `/garden-index` | 15:25:14 | 3 | 3,205 B |
| homegarden `/january-index` | 15:26:00 | 3 | 3,060 B |
| homegarden `/shed-and-outbuildings-index` | 15:26:41 | 3 | 2,709 B |

`[MEASURED 2026-09-03 15:27Z]` **17 new / 0 old.** The deck class is closed: 5/12 yesterday, 9/8 at
15:05Z, 11/6 at 15:18Z, **17/0 at 15:27Z**. idea.uk picked up two extra items on the way (7 → 9) —
the resolver returned more eligible posts than the stale snapshot held, which is the mechanism
working rather than a discrepancy.

### MISSTEP — "an orphan row is unrepairable" over-flags by 6×, and I quoted a count I never took

I told **four** lanes that a `page_components` row with `component_id` NULL can never be repaired
by a re-render, because `resolveComponent` misses and the row takes a carry branch — and I attached
`[MEASURED 2026-09-03] 8 such rows on 3 pages`. The `bugs_open/384` lane challenged it. I checked
rather than adopted, and **they are right on both counts.**

`resolveComponent` (`rerender_page_sections_action.go:361-393`) does not give up on an empty
`componentID`; it falls through to `schemas[s.slotName]`. And `loadComponentSchemas`
(`plan_sections_action.go:1981-2002`) indexes **by both `Name` and `Function`** — its own comment
says so. So a NULL-id row resolves whenever its slot name matches either column.

`[MEASURED 2026-09-03 15:45Z]` **14** such rows on **7** pages: **12 resolve** (every one by
`function`), **2 stranded** — finetuning.uk `/blog` (`article-grid`) and gamesdesign.co.uk
`/game-jelly-invaders` (`section`). **Neither is on boxingonline**, so the six rows I built the
whole argument on do resolve: a re-render would refresh them and clear the empty elements, and what
it cannot fix is that six of them exist. Deletion remains the remedy — for the duplication, not for
the emptiness. That is a materially different instruction than the one I gave.

**Two errors, and the second is the worse one.** The rule was over-wide *and* the figure "8 such
rows on 3 pages" was **inherited from `bugs_open/457`'s own earlier census and repeated as though I
had measured it**, complete with a `[MEASURED]` marker and today's date. The estate's rule is that
a marker proves a measurement was *claimed*, not *complete*; this is the sharper form — a marker on
a number I never took, which reads as first-hand and is not. The plain predicate gives 14 on 7.

**The trap inside the correct predicate, which is the transferable part.** Slot names match
`function`, **not** `name`. **Zero** of the 14 rows match by `name`, so the obvious screen
`WHERE cc.name = pc.slot_name` returns **14 of 14 stranded** — clean, plausible, and exactly the
wrong answer I had asserted from reasoning. The 384 lane caught it only because a **known-good
control** (`content-listing`) also came back false, twenty minutes after they had watched that same
component repair. Nothing in the result would have shown it. Correct form:

```sql
pc.component_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM content_components cc
                   WHERE (cc.name = pc.slot_name OR cc.function = pc.slot_name) AND cc.is_active)
```

**Corrected in five places** — `bugs_open/457`, `bugs_open/425`'s §0q source, the lane handoff, my
CONTRIB in `bugs_closed/454`, and both prose docs — as visible dated blocks, not silent fixes. The
`427` lane folded my wrong version into `454`'s **closure note**, so that one is theirs to amend and
they have been told; a contributor's error that reaches a closure note is the contributor's to chase.

**What I did right, and it is the only reason this was cheap:** I verified the challenge at the code
and re-ran the census before changing anything, exactly as I had asked the `384` lane to do with my
correction to them an hour earlier. Adopting a peer's correction on assertion is the same failure as
asserting it yourself.
