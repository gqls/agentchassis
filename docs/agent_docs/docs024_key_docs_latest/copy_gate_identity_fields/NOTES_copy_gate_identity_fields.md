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

### 6a. …and the correction tripped the append-only ledger check, which was worth answering rather than waving through

`scripts/pattern-check.py` flagged **7 lines removed from `LANDMINES.md`**, a fleet-wide
append-only ledger where a removed line is most likely another session's entry.

Checked rather than assumed, and gated on the COUNT first because the recorded trap here is that
`git diff | grep '^-[^-]'` cannot see a deleted markdown bullet:

```bash
git diff --numstat <sha>^ <sha> -- …/LANDMINES.md          # 65 added, 7 deleted
git diff <sha>^ <sha> -- …/LANDMINES.md | grep '^-' | grep -v '^---'
```

All 7 are my own false bullet from an hour earlier. Nothing of another session's was touched, and
the original claim survives struck-through inside the replacement, so the record of what was
believed is intact — which is the property the ledger exists to protect. **The advisory was right
to fire**: nothing downstream can tell a deleted entry from one never written, and "it was mine"
is only knowable by running the diff.

### 7. I diagnosed a fleet-wide outage from a step name and a stored string, and escalated the wrong mechanism

Round 2 reached `complete_invalid`. I reported to the `dispatch_throughput` lane that the spend
governor was withholding every council submission and that its routing disagreed with its own
`admitted: true`. **They disarmed the council gate estate-wide on that report.**

**Wrong.** The council ran; its reviewers returned nothing readable. The cause was the **account** —
`"Your credit balance is too low to access the Anthropic API"` (400) on every LLM call fleet-wide
11:21:12Z–11:56:48Z, 91 of 107 reviewer calls. The gate admitted every run.

Three one-query checks would each have caught it and I ran none: `__step_error` (I saw `error:
NULL` and stopped — `bugs_open/099`'s landmine, *in my own memory index*, says exactly that the
detail is in `__step_error`); the step's own `description` (*"or a reviewer/decide step errored"*);
and the existence of `complete_withheld` as a **separate** terminal.

**Why it was persuasive:** the other lane's SQL composed `"WITHHELD at shed level 0 …"`
**unconditionally on every run**, so the blob held a decision (`admitted: true`) and a narrative
(`WITHHELD`) that contradicted each other — and I believed the narrative, then built a case
explaining the contradiction instead of asking which was load-bearing. Their own author was fooled
for ten minutes too.

**And my best evidence could not discriminate.** The clean onset — last approval 11:09:35Z, six
runs, none since — is equally predicted by a governor withhold and by an API outage. Specific,
dated, reproducible, and it tested nothing.

**Cost:** a live config change to a shared mechanism, on my say-so. The report was still right to
send — reviews were dead and nobody had been told. **Being right that something is broken does not
license confidence about why.** Full entry in `WRONG_CALLS.md`.

### 8. Round 2's gating objection was about a landmine I had written that morning

`prior_art_librarian`, HIGH: *"The LANDMINES list already carries entries footprinted on the exact
new symbol names this edit introduces — `identityContentField`, `isProseContentField`,
`neverProseFieldRe` … either these symbols already exist and this is an undisclosed rebuild, or the
landmine is a standing warning about exactly this design that the author has not engaged with."*

**Neither. The entry is mine, commit `f9b219a1f`, written hours earlier as this change's own
writeup.** Its footprints name the new symbols precisely because they are the symbols this change
introduces.

The seat did nothing wrong — it followed the house rule that a landmine touching your change must
be read before judging, found one, and could not tell it from prior art I had walked past. **A
landmine written before the review is, to the reviewer, indistinguishable from a warning the author
ignored.** That is a property of the estate's own process, not a misunderstanding, and the fix is
one line of disclosure in the submission: say the entry is yours and give its commit. Round 3 does.

It also caught my `sectionAssetKeyLike` claim from the *other* direction than I did — that my
"never filed anywhere" grep was worthless because the grep's own hit was that landmine. Right; and
by then I had refuted the underlying claim outright, so the paragraph is withdrawn from the
submission rather than repaired.

Two smaller round-2 objections, both accepted: the heading floor's provenance reads as me declaring
authority inside the submission that introduces it (restated in round 3 with how the owner actually
ruled, plus the alternative he declined); and my rationale mis-numbered the very edit it was proving
I had fixed — I renumbered when merging two same-file edits to fit the 8-edit cap and did not update
the prose.

**Reviewers re-ran the census and got 947 / 824 / 0 against my 908 / 825 / 0.** Both correct: the
population grows by addition, which is why it was dated. The cell the guard keys on — **zero
empty/null url siblings** — held in both runs.

### 9. My own edit was swept into another lane's commit, in the file this lane keeps missteps in

I edited `WRONG_CALLS.md`, and before I could commit it another session committed that file by
pathspec and took my change with it (`f2be9beda`, 13:19). Nothing was lost — the text is in HEAD —
but it is now attributed to a commit about a different lane's finding.

This is the documented same-file passenger trap, and worth one line here because of where it
landed: **the file the estate uses to record its own mistakes is high-traffic and shared**, so an
edit to it is more likely than most to be swept. Commit it in its own narrow commit the moment it
is written, not at the end of a batch.

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
