# NOTES — copy gate: `name` fields and the identity/display split

Append-only, newest at the bottom. The missteps are not an appendix; they are the point.

---

## 2026-09-04 — session `420 425`, picking the lane up

**Why this lane exists at all.** Asked to look at bugs 420 and 425, take whichever had no active
thread. Both 425 and the *other* 420 (billing email) had commits within the hour, so both were left
alone. This 420 had zero commits since it was filed on 2026-08-31; the owning lane
(`copy_quality_two_stage`) confirmed by message: *"Take it — I'm not on it and won't be."*

⚠ **Two unrelated bugs share the number 420.** Resolve by slug; `git log` the FILE PATH.

### What the bug file said, and why its fix was unsafe

Candidate 1 was "drop `name` from `nonProseFieldRe`" — a one-line change. Two things made it wrong,
and neither was in the file:

1. That regex was **the only thing** keeping identity fields away from the LLM repair path. The
   `copy_quality_two_stage` lane confirmed at the code that `AcceptNegationRewrite` had
   `invented_name` but **no `dropped_name`** — nothing stopped a rewrite *losing* an identifier.
   Figures are protected symmetrically; names were not.
2. The `components` lane (`bugs_open/425`) found that `name` carries **two opposite contracts by
   producer**. Listing items: `name` IS `pages.name` and the item's url is built from it. Directory
   items: `name` is the display value, prose by design.

**The general shape, which is the transferable part:** one field-name list was answering *"is this
worth scanning?"* and *"may I overwrite this?"* — questions whose false answers cost opposite
things. Filed as a LANDMINES entry; `sectionAssetKeyLike` is the unsplit fourth member and its
mutating consumer executes a `DELETE`.

### The census, and why the structural test won

Over all 1,729 `name`-bearing objects at any depth: 908 with a non-empty `url` sibling had **0**
prose-shaped names; 825 with no `url` key had 752; **0** with an empty/null url. Zero crossover,
reproduced by the components lane from a query sharing no code.

Rejected the **stem test** (`url` ends `/<name>.html`): only 188 of 908, because `guide-x` lives at
`/guides/x/index.html`. Rejected the **lexical test** (case / slug shape) even though it partitions
the same items *perfectly* — perfect agreement corroborates that the partition is real; it does not
license using a property of today's producers as a guarantee.

## MISSTEPS

### 1. I offered the owner a remedy without checking the remedy runs the repair

Put "Mechanism, then rerender the 23 pages" on a menu. **He chose it.** Only when planning the
rerender did I check: `page-rerender` has neither `rewrite_negations` nor `copy_gate_annotate`, and
`rerender_page_sections_action.go:3-9` re-renders stored `content_data` *"WITHOUT invoking the
content writer (no LLM)"*. The defect lives in `content_data.name`. All 23 rerenders would have
completed successfully and repaired nothing.

Caught by myself, one step too late — I had to go back and correct the menu I had given him.
The check is one query (RUNBOOK §5). **An option put to the owner is a claim and needs the same
evidence as any other claim**; offering feels like asking, so it escapes the discipline that
asserting attracts. Full entry in `WRONG_CALLS.md`.

### 2. I wrote a mutation-check claim without running the mutation

Listed in the new test header: *"stop threading `parent` through `walkContentSlice` →
TestWalkerReachesNestedItems fails"*. **False.** That test's fixture is maps-inside-lists all the
way down, every step handled by `walkContentMap`, which has the enclosing object already; `parent`
is only consulted for a string sitting *directly* in a list. Removing the threading leaves the test
passing.

Caught by actually running all twelve claims before committing. Eleven were real; this one was
reasoning about which code path *ought* to be involved. **I was one commit from reproducing, in the
same file, the exact defect I was there to fix** — the file's previous header claimed a mutation
check that the fixture (`"orchestrator"`, `"planner"` — single tokens the VALUE test excludes
anyway) made impossible. Fixed by giving `parent` a genuinely load-bearing test
(`TestIdentityAppliesToStringsInsideANamedList`). 14 mutations then run by hand, all 14 caught.

### 3. I nearly attributed another lane's red tests to my change

`go test ./platform/orchestration/actions/` failed on three tests after my edits. None were mine:
`llm_budget_ladder_test.go` is another session's **untracked** file, and two others fail at plain
HEAD. Established with a **no-overlay control** — `verify-head-builds.sh --test` on HEAD alone
produced an identical failure set. On a tree this many sessions share, a red local test is no more
attributable to you than a green one is a green HEAD.

### 4. My working-tree HEAD reading went "backwards" and I nearly filed it as a problem

Read HEAD as `848a98e16` mid-session, then `29d611750` later, which looked like a reset on a
forward-only tree. `git merge-base --is-ancestor 848a98e16 HEAD` says YES — it is an ancestor, so
HEAD moved forward normally and other sessions had simply committed in between. **Ask git about
git** rather than inferring from two readings taken at different times.

### 5. The council's round-1 REVISE found a real defect in my submission, and it was not cosmetic

Round 1 came back **REVISE**, gating objection from `editquality`:

> "The summary and rationale explicitly claim 'Add `name` to headlineFieldRe (a card name is a
> heading)' as part of closing the IsHeadlineField dual-purpose defect, **but no edit modifies
> headlineFieldRe/IsHeadlineField's regex** — only the ORDER of the identity check relative to the
> headline branch is changed. Without that regex change, `t.Headline` will not become true for
> `name` fields…"

**Correct, and I checked before accepting it.** The change *is* in the committed code
(`negation_content.go:253`) and mutation M5 caught its removal. What was missing was the **edit in
the plan**. I had eight edit slots, used them on the pieces that felt substantial, and let a
one-token regex change ride inside a neighbouring edit's prose.

**Why that is not a paperwork complaint.** Implemented from the plan exactly as submitted,
`t.Headline` never becomes true for a `name` field — so the heading floor is never selected and the
ordering fix guards a severity that never applies. **A fix that silently does two thirds of
nothing**, which is this bug's own failure shape one level along. The reviewer reached that
conclusion from the plan alone, without the tree.

Round 2 (same correlation, `RESUBMIT_CORR`) makes it edit 3, standalone, with its own mutation
check. To stay inside the 8-edit cap I first merged the two *test* edits — **refused server-side:
one edit = one file.** Merged the two same-file predicate edits instead, which also keeps the
headline edit prominent rather than folded away again.

**The lesson:** the plan is the artefact under review, not the tree. A change I could see in my
editor was invisible to the only reader who mattered, and "it is in the commit" would have been a
worse answer than fixing the plan.

### 6. I wrote a subagent's finding into three permanent records without reading the deciding lines, and it was backwards

Claimed, in `LANDMINES.md`, CQ-037 + its index row, and `bugs_open/420`, all within an hour:
*"`sectionAssetKeyLike` is the highest-consequence member of this class — widening it means a
deleted live section for the repair."*

**Backwards.** `remove_duplicate_page_sections_action.go:153` groups by
`SectionIdentityKey(s.Slot, s.Raw)` — the **RAW blob**. Normalised text appears once, at `:148`, as
an 80-char eligibility gate, so widening the shared list yields **FEWER** deletions and cannot make
two raw blobs collide. The consumer the list actually moves is the **read-only** detector
(`check_content_duplication.go:658`).

And the file is the estate's *worked mitigation* of the class I was writing up —
`section_text.go:105-124` states it in capitals: *"IDENTITY IS THE RAW BLOB, NOT THE NORMALISED
PROSE — AND THAT IS DELIBERATE."* **I cited that header as "the mitigation to copy" in the same
bullet where I called the file an open hole.** Both cannot be true and I did not notice.

**How.** The Explore subagent reported it accurately — its words were that the vonc.com deletion was
*"prevented by adding slot+component equality, not by tuning the list"*. I compressed "dangerous
shape, already fixed" into "dangerous", kept the alarming half, dropped the half that made it safe.

**What caught it:** going to verify before filing it as a bug, because filing would have asserted it
a fourth time. Two greps and one `sed` of the deciding function. Nothing external would have caught
it — all three records agreed with each other.

**The rule this earns, and it is not the one I already knew:** every durable figure in this session
was measured, dated and controlled, and the `[INFERRED]`/`[UNMEASURED]` marker discipline did
nothing here — because a sentence relayed from a subagent **arrives already phrased as a finding**.
So: *a claim sourced from a subagent gets the same evidence bar as one I would otherwise mark
`[INFERRED]`.* Corrected in four places plus a follow-up to the components lane, who I had told it
might be worth their picking up.

## DECISIONS AND THEIR REASONS

- **The guard went in the JUDGE, not the walker.** A filter at the enumeration point is bypassable
  by any future caller that enumerates fields itself; a rejection reason is not. The
  `copy_quality_two_stage` lane proposed this and it is stronger than what I had planned.
- **The identity flag rides on the yielded field rather than filtering the walk**, so
  `total = exempt + withinBudget + targets` still reconciles with the annotation's count. The
  walker's own header forbids the two consumers scanning different sets.
- **The identity exemption sits ABOVE the headline branch, and that ordering is the whole
  protection** — `name` is now headline-class, and a headline hit is never forgiven by the budget,
  so below it an identity field would be *forced* to the model. Pinned by a test whose stated
  mutation is the reorder.
- **Two exported predicates were planned; one exported const shipped instead.** Neither predicate
  has an external caller, and exporting a helper nothing calls is its own trap. The exported
  surface is `NegationTextField.Identity` and `IdentityNameWithURLSibling`.
- **One commit, not two.** The plan called for the heading floor as a separate commit for
  reviewability. Rejected on a shared tree: the intermediate state ships a fix that refuses two
  thirds of its own repairs, and any session's next roll would have shipped exactly that. The
  commit message names the floor's exact hunk instead.

## CORRECTION MADE TO ANOTHER LANE'S FILE

`repair_ordering_register_action.go` said the shared `gutted` floor is "40% of the original
length". It has not been a bare proportion since `7cc16a5d0` (2026-09-03) — it is
`wordCount(to) < 5 || len(to) < len(from)/4`. Corrected visibly in a comment block this change
already touched, and disclosed in the commit message. The argument there is unaffected; only the
number was stale.

## NOT DONE, DELIBERATELY

No `PLAN_*.md` here — the approved plan is at `/home/ant/.claude/plans/zany-swimming-ritchie.md`
and `bugs_open/420` §RESOLUTION is the authoritative design record; a third copy would drift.
No `SUMMARY_*.md` — rarity is part of that document's design, and one bug fix that has not yet
reached production is not a milestone.
