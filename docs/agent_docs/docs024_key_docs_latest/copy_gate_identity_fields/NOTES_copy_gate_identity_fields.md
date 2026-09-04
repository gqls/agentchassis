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
