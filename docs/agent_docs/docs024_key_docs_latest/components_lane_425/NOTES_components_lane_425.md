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
