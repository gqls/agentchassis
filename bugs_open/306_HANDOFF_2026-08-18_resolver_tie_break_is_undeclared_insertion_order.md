# 306 — the resolver's conflict winner is decided by undeclared insertion order, in a population where 9.4% of candidate sets contain a genuinely different page

**Filed 2026-08-18** by the `staged_component_build` lane, on the owner's direction
("take it out of an RFC and into a bugs_open so it is dealt with" — this was RFC_029 §10.12's
finding, and §10.12 now points here rather than restating it).

> ## Why this is filed on first-hand verification instead of a `090` run
> Per the owner ruling of 2026-07-31, a structural claim needs the diagnosis loop OR a stated
> substitution. Substitution stated: every claim below is a direct quote of code read this
> session (`unified_extractor.go`, exact lines cited) plus a 139-run-per-population measurement
> against `orchestration_states.collected_data` whose disconfirming results were available and
> in two cases **came back non-zero** (13/139 different pages; differing loop-iteration specs).
> There is no inferred mechanism anywhere in this file. The full measurement record with the
> queries is RFC_029 §10.12 and the lane RUNBOOK ("Are the recorded conflicts DIFFERENT VALUES
> or the same value in two shapes?", 2026-08-18).

## 1. The mechanism (all read at code, 2026-08-18, build v1.0.1309 / `f0117fb8b`)

`findFieldRecursive` (`platform/orchestration/datahelpers/unified_extractor.go:503`) is the
whole-tree-search arm of input resolution. Since RFC_029 Phase 1 it collects ALL candidates and
resolves conflicts to a stable winner. **The winner is chosen by:**

1. `sort.SliceStable(candidates, … depth < depth)` — **depth ONLY** (line ~612);
2. ties at equal depth keep **insertion order** into the slice;
3. insertion order is: the `~unwrap` hop (`tryUnwrapMapPatterns`) is appended at line ~606
   **before** the sorted-key recursion into sibling keys.

So at equal depth, whatever the unwrap hop found beats every sibling — and nothing declares
that, no comment states it, and no test pins it. The comment on the sort explains why it does
not sort by path string; it does not say what DOES break the tie.

## 2. Why the tie-break is currently load-bearing (measured)

Population: every `page-build-handler` orchestration of 24 h to 2026-08-18 ~14:00Z carrying a
`page_content.retry_payload` subtree — **n = 139**.

- **13/139 (9.4%) candidate sets contain a GENUINELY DIFFERENT page**, not a shape variant:
  `disclaimer` vs `contact-index`, `index` vs `fuel-pricing-framework`,
  `guide-mortgage-scorecard` vs `scorecard-simulator`, `case-study-data-pipeline-companies-house`
  vs `case-study-automated-intelligence-pipeline`. The loser is a stale page inside a
  `retry_payload` subtree left by an earlier step.
- Both candidates sit at **equal depth 1** (`~unwrap.current_page` vs the retry-payload path is
  deeper, but `input_data.current_page` is the depth-1 sibling), and the winner in all 139 was
  `~unwrap.current_page` — which resolves to `input_data.current_page`, the page the run was
  dispatched for. **The right page won every time. It wins by rule 2+3 above.**
- The resolved value **is read**: `page-build-handler` and `page-content-writer` consume
  `current_page` (unlike `build-dispatch-loop`, whose 63% of conflict rows land in a slot
  nothing reads — see RFC_029 §10.12's four-population table).

**So there is no live wrong-page defect today.** The defect is that correctness in a
9.4%-genuinely-ambiguous population rests on an implementation accident. Any of these innocent
edits flips it: reordering the two append sites in `collectFieldCandidates`; sorting candidates
by (depth, path); "simplifying" the stable sort to an unstable one. None would fail a test.

## 3. The second, adjacent instance: the unwrap hop itself still coin-flips

`tryUnwrapMapPatterns` (`unified_extractor.go:~676`) pattern 1 does
`for key, val := range m` over a Go map and returns the FIRST `*_result` key carrying a
`result` child — **unsorted map iteration**, i.e. the exact nondeterminism RFC_029's
collect-all rewrite removed, surviving one call INSIDE the function that was made
deterministic. **Measured 0/139 able to fire in the population above** (no root `*_result`
object there has a `result` child, so pattern 3 → `input_data` always wins), so this is NOT a
live defect and must not be reported as one. It is where the coin flip re-enters the moment any
workflow stores a `*_result` object with a `result` child at root — a shape other workflows do
produce (`{X}_result.result` is pattern 1 precisely because it was once common).

## 3a. STATUS 2026-08-19 — candidates 1+2 BUILT, council APPROVED (all-approve, corr `96ac93e6`), commit `846496906`; INERT UNTIL THE NEXT CHASSIS ROLL

Candidate 1 shipped as a `rank` field on `fieldCandidate` (direct ≺ `~unwrap` ≺ sibling, inherited
from the first hop off the root — which equals historical append order at every level because the
collector is depth-first and exhausts each root branch before the next); the sort reads it.
Candidate 2 shipped as a `sort.Strings` over pattern 1's `*_result` keys. Six tests in
`unified_extractor_tiebreak_test.go`, incl. the 13/139 production shape; mutation-proved both
ways (rank inverted → 2 fail while every previously pinned winner still passes; sort removed →
determinism test fails 30/30). **Still OPEN:** candidate 3 (shrink the ambiguous population by
skipping `retry_payload` subtrees) — needs its own blast-radius check and is NOT bundled; and
the bug stays open until the roll makes the declared tie-break LIVE. Close when: the roll is
verified by label+digest AND `TestTieBreakUnwrapHopBeatsSibling` is in the stamped revision.

## 4. Fix candidates, ordered by what closes the door

1. **Declare the tie-break and pin it** (small, no behaviour change): make the sort
   `(depth, source-class)` where source-class ranks `~unwrap`/`input_data`-rooted paths ahead of
   siblings — i.e. the CURRENT winner, now stated — and add a test with two equal-depth
   candidates that fails if the preference inverts. Closes the reorder door permanently.
2. **Sort pattern 1's keys** in `tryUnwrapMapPatterns` (one line + test): removes the last
   unsorted-map selection on the search path. Behaviour-identical today (0/139 fire), which is
   exactly why now is the cheap moment.
3. **Shrink the ambiguous population** so the tie-break stops mattering: the 13/139 exist
   because `retry_payload` subtrees retain stale pages. Adding `retry_payload` to
   `isInfrastructureKey`'s skip list would stop the search descending into retry echoes at all —
   **but this widens what the skip list means and needs its own blast-radius check** (who relies
   on recovering fields from a retry payload?). Do not bundle it with 1+2.
4. Long term this population shrinks further under RFC_029's `ensureCoreFields` gating decision
   (open with the owner) — not this bug's to decide.

## 5. How to verify

- Unit: two equal-depth candidates, different values, assert the declared winner; mutation-prove
  by inverting the preference (test must fail). Pattern-1 determinism: a map with two `*_result`
  keys, assert the same pick across runs (today it flips under `-count=100`).
- Live: the RUNBOOK's four-step method (resolve recorded `candidate_paths` against
  `collected_data`; the instrument stores paths, never values).

## 6. Relations

RFC_029 §10.12 (the measurement record) · §10.11 (why the field-list prune cannot reach this
population) · `bugs_open/085` (what an absent current_page does downstream) · the Phase 2
decision, which is gated behind this bug's population going to zero or being declared safe.
