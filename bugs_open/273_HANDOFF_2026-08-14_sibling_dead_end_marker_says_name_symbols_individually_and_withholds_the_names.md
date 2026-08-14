# 273 — the sibling section's dead-end marker says "Name symbols individually" and withholds the names

**Filed 2026-08-14, `silent_hero_logo_readers` lane. This is `bugs_closed/261` §8 follow-up 2,
promoted to its own number the way follow-up 1 became `bugs_open/269` (now closed).**
Status: **FIX IN TREE with tests, NOT LIVE** — inert until a chassis image rolls, per the
fixed-AND-live bar.

## 0. Why no `090` run (OWNER RULING 2026-07-31 declaration)

This file asserts a root cause without a diagnosis-loop run, substituting equivalent first-hand
verification, declared here:

- the mechanism is a **cap read directly in the code** (`diagnose_assemble_bundle_action.go`,
  `siblingSignatures` — per-file share `capChars/len(scoped)`, floor 600), not an inference across
  tiers;
- the harm is **already witnessed on a live run and recorded** by the 261 lane: run
  `dbcc4259-ab84-494b-a48b-1df647209a40` (090 on `bugs_open/236`), 4 iterations, UNVERIFIABLE, its
  `needed_evidence` naming three functions that sat behind this marker (261 §1, §8.2);
- the fix is **self-evidencing at the unit level**: the mutation run below watches the new tests
  fail against the pre-fix marker and pass against the fix, in one session.

## 1. The defect

`siblingSignatures` lists same-file signatures so the model can name a sibling in `next_scope`.
Each file gets `siblingSigCap/len(scoped)` chars (floor 600) — roughly 5–10 signature lines — and
the elided remainder collapses to a `+N more` marker. Since `bugs_open/267`, a file whose whole
body exceeds `max_body_chars` gets the honest dead-end wording:

> `+N more in this file — not listed here, and the bare file path will NOT render: the whole file
> exceeds the 60000-char body budget. Name symbols individually.`

**"Name symbols individually" is unsatisfiable for the elided tail.** The model can only name a
symbol it has been shown. The symbols behind the marker are, by construction, ones retrieval did
not surface (or they would be in scope) — so for them there is *no path at all*: the bare-path
re-read is refuted by the file's size, and the names were never rendered. It is the same defect
class 267 fixed one section up (an invitation the bundle's own arithmetic refutes), one layer down
(an instruction whose required input is withheld).

## 2. The recorded harm case

261 §8.2, verbatim: iteration 4 of the `236` run had a scope collapsed to five symbols (three of
them copies of a trivial `getMapKeys`); the per-file cap hid the functions the run needed behind
`_(+79 more in this file — put the bare file path in next_scope to see it whole)_` (pre-267
wording; the file was `platform/orchestration/coordinator.go`). The run returned UNVERIFIABLE with
the evidence in the index the whole time.

## 3. The arithmetic `[MEASURED 2026-08-14]`

- `coordinator.go`: **169,139 bytes** against the 60,000 default `max_body_chars` — can never
  render whole — and **91 functions**, of which a 600-char share shows ~5–10. Disconfirmable: a
  file under 60,000 bytes, or with few functions, would have refuted the dead-end premise.
- Complete compact tail of *canonical* handles for the worst real files (receiver-qualified
  spelling, after ~10 head lines): `coordinator.go` **2,715 chars**; `v3_site_actions.go` (85
  fns) **1,935**; `data_helpers.go` (71 fns) **1,231**. Measured by generating the handles from
  source, not estimated. So a 4,000-char per-file tail bound covers every file in the repo today.
- The alternative remedies all inherit their own caps: a `code_request` kind `symbol` with the
  path as query enumerates `code_symbols` rows for the file but is row-capped and that cap's
  silence is `bugs_open/181` (open); `ls` lists paths, not symbols.

## 4. The fix (in tree, commit pending council correlation)

`platform/orchestration/actions/diagnose_assemble_bundle_action.go`:

1. The loop collects the canonical handles of skipped functions (declaration order).
2. In exactly the dead-end branch (`fitsBudget` known && !fits), `writeDeadEndTail` appends the
   elided handles compactly — no per-line path or signature — with the explicit instruction
   `Name them individually in next_scope as `<path>:<handle>``. Canonical, not bare: a bare
   method name resolves to the first same-named function and returns the wrong body (269).
3. The tail is bounded by `siblingDeadEndTailCap = 4000` per file. Past it, the marker counts the
   residual and names the one remedy that can enumerate the rest (`code_request` kind `symbol`,
   query = the path) — it never trails off silently, which would be this bug one layer further
   down.
4. **The tail is exempt from the section's global guard** (`capChars + capChars/4`). Counting it
   would evict the whole sibling section on the motivating case — one scoped over-budget file
   builds head (= full share) + tail > guard, and the model's only map of the file would be
   replaced by "further files omitted". Worst case is stated in the code:
   `capChars*5/4 + (dead-end files × 4000)` beside a 60,000-char body budget.
5. Every other branch byte-identical: could-fit and unknown-size markers unchanged, no-skip files
   unchanged, no change to any censused marker phrase ("did not fit" / "could not be read" /
   "read it whole" / "NO next_scope can render this path" — see the LANDMINES entry on counting
   wasted iterations by marker strings; the new tail avoids all of them, asserted by test).

Tests: `diagnose_assemble_deadend_tail_test.go`, four tests. Mutation-proven 2026-08-14: with the
old marker restored behind `if false`, `TestDeadEndTail_ElidedHandlesAreListedCanonically`,
`_OverflowIsCountedAndGivenARemedy` and `_DoesNotEvictItsOwnSection` all FAIL; the byte-identity
pin passes both sides (it is the negative control). Full related set
(`DeadEndTail|OverCapAdvice|SiblingSignatures|SiblingSpelling|Assemble|Bundle`) green post-fix.

## 5. How to verify live, after a roll

1. Stamp + ancestry + control, per the recipe in
   `silent_hero_logo_readers/HANDOFF_2026-08-13_continue_here.md` §2 (precheck the startup line is
   in range first).
2. Behaviour, with a demand control: dispatch or await a diagnosis whose scope touches an
   over-budget file (`coordinator.go` is the natural one). In the bundle
   (`diagnosis_artifacts` kind='bundle'), the sibling section for that file must carry
   `The elided handles:` followed by canonical handles. **A zero is evidence only if a dead-end
   file was actually scoped** — check the section names the file and says
   "the whole file exceeds the" first.
3. The census phrases must not have moved: the 267 §4b `cap_only`/`resolver_only` counts, re-run
   across the roll, must not step at the roll boundary for bundles whose scope had no dead-end
   file.

## 6. What this deliberately does not fix

- **Types and package-level values are still not listed as siblings** — `siblingSignatures` is
  functions-only by design (`:715` comment). A dead-end file's types/consts remain enumerable only
  via `code_request`/`overCapAdvice`'s largest-that-fit list. Separate question; widening the
  section's kind coverage changes its size arithmetic and belongs to its own case if wanted.
- **`bugs_open/181`** (code-lookup row caps silent) — the overflow remedy inherits it; fixing 181
  is that file's work.
- **261 §8 follow-up 3** (`knownScopeIdentities` omits `values`, cosmetic) and **follow-up 4**
  (precedent check — since run, see 261 §9b) stay where they are.

## 7. Relations

`bugs_closed/261` (parent, §8.2 is the harm record) · `bugs_closed/267` (the satisfiability rule
this extends to the sibling section) · `bugs_closed/269` (why the handles are canonical) ·
`bugs_open/236` (the starved run) · `bugs_open/181` (the remedy's own cap) ·
`architecture_review/RFC_027` (the handle grammar's ownership — this adds no new producer;
handles come from `analysis.CanonicalSymbolName`, the grammar's one owner).
