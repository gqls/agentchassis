# 273 — the sibling section's dead-end marker says "Name symbols individually" and withholds the names

**Filed 2026-08-14, `silent_hero_logo_readers` lane. This is `bugs_closed/261` §8 follow-up 2,
promoted to its own number the way follow-up 1 became `bugs_open/269` (now closed).**
Status: ~~FIX IN TREE with tests, NOT LIVE~~ **CLOSED 2026-08-15 — FIXED AND LIVE on chassis
`v1.0.1300`** (§9: build point `a2a691213`, both replicas, controls run, both 273 commits
ancestors). **One caveat, stated where it can't be missed:** the live-behaviour witness is still
outstanding — **zero** diagnosis bundles of any kind have been assembled since the roll, so there
has been no organic demand to exercise the new tail. The defect's code no longer exists in the
running binary and the branch is mutation-proven in tests; the first live bundle that scopes an
over-budget file is the remaining witness. §5's recipe (with its demand control) stands for
whoever sees one.

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

## 4. The fix (committed `4f3f0be7d` + post-verdict follow-up; council APPROVED `ba3f6047`)

`platform/orchestration/actions/diagnose_assemble_bundle_action.go`:

1. The loop collects the canonical handles of skipped functions (declaration order).
2. In exactly the dead-end branch (`fitsBudget` known && !fits), `writeDeadEndTail` appends the
   elided handles compactly — no per-line path or signature — with the explicit instruction
   `Name them individually in next_scope as `<path>:<handle>``. Canonical, not bare: a bare
   method name resolves to the first same-named function and returns the wrong body (269).
3. The tail is bounded by `siblingDeadEndTailCap = 4000` per file. Past it, the marker counts the
   residual and names the one remedy that can enumerate the rest (`code_request` kind `symbol`,
   query = the path) — it never trails off silently, which would be this bug one layer further
   down. **ADDED 2026-08-14 post-verdict (§8d): `siblingDeadEndTailTotalCap = 12000` bounds the
   tails' AGGREGATE across the section**, implementing the council's recommendation; a dead-end
   file arriving after the aggregate is spent gets the overflow arm immediately.
4. **The tail is exempt from the section's global guard** (`capChars + capChars/4`). Counting it
   would evict the whole sibling section on the motivating case — one scoped over-budget file
   builds head (= full share) + tail > guard, and the model's only map of the file would be
   replaced by "further files omitted". Worst case is stated in the code:
   `capChars*5/4 + (dead-end files × 4000)` beside a 60,000-char body budget.
5. Every other branch byte-identical: could-fit and unknown-size markers unchanged, no-skip files
   unchanged, no change to any censused marker phrase ("did not fit" / "could not be read" /
   "read it whole" / "NO next_scope can render this path" — see the LANDMINES entry on counting
   wasted iterations by marker strings; the new tail avoids all of them, asserted by test).

Tests: `diagnose_assemble_deadend_tail_test.go`, five tests (the fifth, aggregate-cap, added
post-verdict per §8d). Mutation-proven 2026-08-14, twice: with the old marker restored behind
`if false`, `TestDeadEndTail_ElidedHandlesAreListedCanonically`, `_OverflowIsCountedAndGivenARemedy`
and `_DoesNotEvictItsOwnSection` all FAIL while the byte-identity pin passes both sides (it is the
negative control); and with the aggregate accounting disabled the same way, exactly
`_AggregateCapBoundsTheSumOfTails` FAILs. Full package green post-fix.

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

## 8. The council round — APPROVED first round, 3 advisory objections, all discharged here

Correlation `ba3f6047-a2e5-4ce6-ac0e-edf0bb88c4e3`, decided 2026-08-14 14:23Z:
`approved with 3 advisory objection(s) — none high-severity`; 5 abstained; guardian, reuse,
tooling-provenance, diagnosis-guardian, render-guardian, debug-historian all approve.

- **8a — editquality (medium): "the sketch doesn't show that `skippedHandles` reaches the dead-end
  branch with the right data."** It does, structurally: both branches are the SAME per-file loop
  iteration. `skippedHandles` is declared inside `for _, f := range scoped`, filled by the same
  `f.Functions` pass that increments `skipped`, and the dead-end branch runs immediately after in
  that iteration. For a `wholeFileOmitted` file, `inScope[path]` registers with an EMPTY named
  set, so every function flows through the same loop. The objection is an artifact of sketch
  elision — this lane's trap 4 again, though only medium this time. Asserted at function level by
  all five tests and at action level by the existing
  `TestOverCapAdvice_BareFilePathSuppressionIsWiredThroughTheRealAction`.
- **8b — editquality (medium): "`canon` might be a bare name, reintroducing 269."** It is
  `analysis.CanonicalSymbolName(fn)`, computed at the top of the same loop (269's fix), receiver-
  qualified for methods. `TestDeadEndTail_ElidedHandlesAreListedCanonically` requires the
  `(*Big).methodNumberNN` spelling in the rendered tail.
- **8c — editquality (low): "`tailLen` could be stale across files."** It is declared inside the
  per-file loop (`tailLen := 0` after the listed/skipped early-continue), fresh each iteration.
- **8d — bug_historian (medium): "per-file bound only; N dead-end files add N×4000 uncounted chars
  — the `bugs_closed/062` shape. Recommend a total cap per bundle."** **IMPLEMENTED as
  recommended, same day**: `siblingDeadEndTailTotalCap = 12000`, allowance =
  `min(siblingDeadEndTailCap, remaining)`, and at zero the overflow arm fires on the first handle
  — a count and the `code_request` remedy, never silence. Worst-case uncounted bytes are now
  `totalCap + one marker's prose per dead-end file`. The recommendation was right to refuse the
  "N is small" answer: this repo already holds **8 files over the 60,000 budget**, so N was only
  socially bounded. New test `TestDeadEndTail_AggregateCapBoundsTheSumOfTails`, mutation-proven
  (aggregate accounting disabled behind `if false` → exactly that test fails).
- **8e — bug_historian (low): "does a sibling elision marker with the same defect exist
  elsewhere?"** Surveyed: the body section's `overCapAdvice` lists the largest-fitting symbols
  and cross-references this section — which for dead-end files is now complete for functions; the
  workflow-step forwarding cap reports its excluded count (the eba040a9 audit's fix); the
  remaining capped listings live in `diagnose_code_lookup_action.go` and are `bugs_open/181`'s
  scope (named in §6). Types/values stay unenumerated by design (§6).
- **8f — reuse_agent asked for confirmation that `code_request` kind `symbol` exists and matches
  the call shape.** It does: `ValidCodeRequestKind` (`pkg/diagnose/loop.go:185-191`) accepts the
  closed set `symbol|content|ls`, and the symbol arm parses path tokens into `path ILIKE` clauses
  (`diagnose_code_lookup_action.go:1578-1590`), so `query = "<path>"` enumerates that file's
  indexed rows.

## 9. Live proof, 2026-08-15 — and the honest zero

Chassis `v1.0.1300`, pods `agent-chassis-6c68fcc549-8lb6d` / `-hptsr`, started ~2026-08-14 21:00Z.

- **Startup stamp out of log range on both replicas** (11h-old pods; first line in `--tail=100000`
  range was 06:47Z/07:36Z on 08-15) — as the §2-recipe landmine predicts, this means "not in
  range", NOT "unstamped". Fell back to the binary probe.
- **Binary probe, discovery done safely:** extracted every distinct 40-hex string from
  `/proc/1/exe` (78 of them) and intersected with real git objects — **exactly one is a commit**:
  `a2a691213dfbe11d38549f128870ef41cbf24a83` (2026-08-14 20:16Z). This sidesteps the
  digit-table trap because junk strings are not git objects. Same stamp PRESENT on replica 2;
  control (today's post-roll commit `f8cfa131b`) ABSENT on both replicas and correctly NOT an
  ancestor of the stamp.
- **Ancestry:** `4f3f0be7d` (fix) IN · `e57ecdf1c` (aggregate cap) IN. Both aboard.
- **Behaviour, with the demand control run first:** bundles assembled since the roll = **0**, so
  `fix_witness = 0` and `dead_end_demand = 0` are **unreadable zeros** — exactly the shape §5
  warns about. Nothing has asked the diagnosis loop anything since the roll. The closure rests on
  the binary proof plus the mutation-proven unit branch; the organic-demand witness transfers to
  §5, unexpired.

Query used (rerunnable):
```sql
SELECT count(*) AS bundles_since_roll,
       count(*) FILTER (WHERE body LIKE '%the whole file exceeds the%') AS dead_end_demand,
       count(*) FILTER (WHERE body LIKE '%The elided handles:%')        AS fix_witness,
       min(created_at), max(created_at)
FROM diagnosis_artifacts
WHERE kind='bundle' AND created_at >= '2026-08-14 20:16+00';
```
