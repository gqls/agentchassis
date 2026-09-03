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
