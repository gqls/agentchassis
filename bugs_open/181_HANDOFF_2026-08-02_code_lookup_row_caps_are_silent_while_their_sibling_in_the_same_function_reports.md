# 181 — the code-lookup's three row caps are silent, while a sibling cap eight lines away reports itself

> **KEPT IN `bugs_open/` BY OWNER DIRECTION, 2026-08-06 — do NOT move this to
> `bugs_closed/` on the strength of the closure section at the foot.** The fix
> is shipped, live on `v1.0.1259`, pod-verified on both replicas and **induced in
> production both directions in one run** (see § CLOSED at the end, which records
> the evidence and is accurate as evidence). This thread had moved the file to
> `bugs_closed/` and the owner directed it back. So: the closure section documents
> the fix; **the file's LOCATION is the owner's call, not this file's own
> verdict.** Two follow-ups in that section are also genuinely still open and
> unstarted — re-examining the five capped `landmine-verifier` renders, and the
> codebase-wide silent-cap inventory the council's `bug_historian` seat asked for.

**Filed 2026-08-02 by the `bugfix_172_agent_state_cap` lane. OPEN, UNOWNED.**
**LATENT as far as this corpus can say — and the corpus CANNOT say, see § Measured.**

## Why this exists

Filed at the council gate's request. Reviewing `bugs_open/172`'s fix, the
`bug_historian` seat approved but objected (medium) on exactly one ground:

> "this is the THIRD pass over the same 'hard cap silently discards its tail'
> mechanism in this loop family … Nothing in this plan audits for or lints against
> other unaudited caps in the diagnose loop … **the next occurrence will again be
> independently rediscovered.**"

It asked for an inventory. This is the inventory's one true positive.

**The chain so far, and it is the point:** `bd003f67a` (2026-07-20) audited
`diagnose_load_runtime_action.go` for this shape and missed `164`; `164`'s own audit
missed `172`; `172` was filed for a **count-based** cap and did not see the
**shared-budget** cap three lines below it. Each pass narrowed to the shape it
happened to grep for. This is the fourth site, in the sibling file.

## The defect

`platform/orchestration/actions/diagnose_code_lookup_action.go`, `answerCodeCheck`
— three query arms, one per `code_check` kind, each capped by the same `row_cap`
(default **40**, `:373`):

| line | arm | cap | reports it? |
|---|---|---|---|
| `:548` | `symbol` lookup | `LIMIT $n` = `rowCap` | **no** |
| `:602` | `content` search | `LIMIT $3` = `rowCap` | **no** |
| `:648` | `ls` listing | `LIMIT $3` = `rowCap` | **no** |

Each arm's only terminal branch is `if n == 0 { b.WriteString(scope.emptyAnswer(...)) }`.
There is **no `n == rowCap` branch anywhere in the file** — so a search that returned
exactly 40 of 300 matches renders identically to one that found exactly 40 and
stopped because that was all there was.

**What makes it a defect rather than a knob:** the same function's *own* sibling cap
does report, eight lines above the first arm (`:402`):

```go
fmt.Fprintf(&b, "\n> %d further code_check(s) dropped (max_checks=%d) — coverage was capped, not complete.\n", dropped, maxChecks)
```

So the file already holds the convention, applies it to *how many checks are
answered*, and does not apply it to *how much each check returns*. A reviewer
reading a bundle sees the first cap declare itself and reasonably infers the absence
of such a line means nothing was capped.

`ls` is the worst arm: it is `ORDER BY path`, so the discarded set is an
**alphabetical tail** — the identical shape `bugs_closed/164` was filed and fixed
for, in the sibling file, by the lane that then filed `172`.

## Measured — and read this before quoting the file

**The retained bundle corpus cannot answer whether this has fired, and my first two
attempts to make it answer produced confident numbers from an empty population.**

The render writes a per-check delimiter (`[code_check %d] kind=… query=…`, `:394`),
so the natural instrument is to split retained bundles on it and count rows per
block. Result:

- `diagnosis_artifacts` `kind='bundle'`: **276 rows**, of which **0** contain a
  rendered `[code_check …]` block.
- The 22 bundles matching `code_check` at all match it inside **quoted source code**
  — excerpts of `diagnose_load_runtime_action.go` pulled in by this lane's own
  diagnosis run today. Not renders. Not evidence.

So: **`[UNMEASURED]`, and not measurable from this corpus.** Either this action's
output does not reach the bundles that are retained, or it has not run in the
window. That question is the first job for whoever picks this up, and it is
interesting in its own right — a code-search facility with no rendered output in 276
retained bundles is either mis-plumbed or unused, and both are worth knowing.

**Do not repeat my mistake.** My first query filtered on the delimiter and returned
"0 blocks at the cap", which reads exactly like "the cap never fires". It was an
empty population. The guard is one extra column: **select the denominator in the
same query as the finding** — `count(*)` alongside `count(*) FILTER (WHERE …)` —
so a zero numerator can never be read without its zero denominator beside it.

## Fix candidates, ordered by what closes the door

1. **Report the cap per arm, reusing `:402`'s wording.** `n == rowCap` → append
   `> this listing is capped at %d row(s) — there may be more; narrow the query or
   raise row_cap.` Three call sites, one shared line. Cheapest, and it matches what
   the file already does one screen up.
2. **Make the render structurally unable to omit it** — have each arm return
   `(rows, capped)` and let one writer emit both the rows and the notice, so a
   future fourth arm cannot forget. Closer to unrepresentable; a slightly larger
   diff.
3. Raise `row_cap`. **Weakest** — it moves the cliff without making it visible, the
   knob `145` was refused for.

## How to verify a fix

- **Induce it.** Point a `symbol` check at a query with >40 matches (`code_symbols`
  has plenty) and assert the notice renders; assert a 39-row answer renders
  byte-identically to today.
- **Negative control:** an arm returning fewer than `row_cap` rows must not gain a
  line, or every existing bundle baseline moves.
- **First, answer the § Measured question** — if the action's output genuinely never
  reaches a bundle, the fix is still correct but the priority changes a great deal.

## Contribution from the 163 lane (2026-08-03) — your symbol-arm baseline has moved, and your fix candidate 2 got its single writer

Filed here per the `bugs_closed/164` ruling (an adjacent change in the function you have
claimed must be named, not silent). The `bugfix_163_symbol_lookup` lane (commit `c3b02f035`)
rewrote the **symbol arm's** query construction — path half now binds to the `path` column,
and a path-qualified miss re-runs name-only and reports both facts. Consequences for this bug:

- **Your verification recipe's byte-identical baseline no longer holds for the symbol arm.**
  "A 39-row answer renders byte-identically to today" is still true for `content` and `ls`,
  but a symbol answer may now carry a `note:` line (line-ref degrade), a `-- searched:` line
  (0-row predicate narration), or an `ELSEWHERE` block (path miss disclosure). Re-baseline
  from `c3b02f035`, not from 2026-08-02.
- **Helpful to your fix candidate 2:** the symbol arm's rows (primary AND fallback) now flow
  through one `renderSymbolRows(ctx, db, clause, args, rowCap, codeOut, docs)` — the "one
  writer" your candidate 2 asks for. A `n == rowCap` notice added THERE covers both symbol
  branches at one site; `content` and `ls` remain inline loops, untouched.
- **No cap behaviour was changed**: `rowCap` still flows to `LIMIT` unreported in all three
  arms. Your defect stands exactly as filed.

- **`bugs_open/172`** (this lane) — the two caps in the sibling file; the fix landed
  `3761a04ca` / `c8031e284`, council `d47b826e-6fc6-42ad-a2ef-62b1f1ba0b88` APPROVED.
- **`bugs_closed/164`**, **`bd003f67a`** — passes one and two over this family.
- **016b §9** *"A hard cap that silently discards its input's tail rewrites meaning"*
  and its new neighbour *"A budget SHARED across a named set is spent by its busiest
  member"* (added by this lane, 2026-08-02).
- **Filing method, declared:** this file asserts a **code-local** claim — three named
  sites lack a branch that a fourth site in the same function has. It was verified by
  reading all four sites and the file's every `rowCap` reference, not by the `090`
  loop. Per CLAUDE.md's 2026-07-31 ruling that substitution is stated, not silent:
  the claim is a direct reading of control flow with no cross-cutting inference in
  it. The part that IS uncertain — whether it fires — is marked `[UNMEASURED]` above
  rather than argued.

---

## SUPERSEDED 2026-08-06 — § Measured is answered: the cap has FIRED, five times, and the right corpus was `llm_call_log`

The original § Measured stands as written for what it checked: retained
`diagnosis_artifacts` bundles genuinely hold 0 rendered `[code_check]` blocks.
The mistake was the corpus, not the query. The action's output travels
`results_text` → the step result → `collected_data` → **embedded in the next
LLM prompt** — so the durable record is `llm_call_log.prompt_rendered`
(retention back to 2026-03), not the bundle table. Measured 2026-08-06, with
the denominator in the same query per this file's own warning:

```
llm_call_log: 49,723 calls total; 39 carry rendered [code_check ...] blocks
blocks parsed across the 39: 233
row-count distribution: mostly 0–26; one block at 39; FIVE at exactly 40 — the cap
all five capped blocks: agent_type = landmine-verifier (07-31 ×4, 08-03 ×1)
```

Re-running the five capped queries against the live index (denominator vs the
40 rendered):

| query | true matches | discarded |
|---|---|---|
| `content` "site_work_items" | 82 | 42 (51%) |
| `content` "page-build-handler" | 43 | 3 |
| `ls` "platform/orchestration/actions/" | **305** | **265 (87%)** — an alphabetical tail, `ORDER BY path` |
| `content` "success" | 279 | 239 (86%) |

*(Caveat: today's index vs 07-31/08-03 renders — the index has grown since,
but the magnitudes dwarf any growth.)* The one 39-row block (uncapped,
provably complete) renders in exactly the same style as the five capped ones —
the indistinguishability this file names, observed in the wild. And the agent
that read all five as complete is **landmine-verifier — the agent whose whole
job is deciding whether recorded landmines still apply to current code.**

**This also answers the "first job" question the § posed:** the output is not
mis-plumbed and not unused — it reaches prompts; the bundle corpus was the
wrong instrument. Also found while measuring: `answerCodeCheck` has a
**fourth consumer** this file did not name — the diagnose loop's own
`code_requests` lane (`diagnose_load_runtime_action.go:483`, `code_row_cap`
default 40) — which inherits any fix to the shared function unedited.

## FIX SHIPPED IN CODE 2026-08-06 — commit `df281f6ba`, council-submitted `f22f7ff1`

Design: candidate 2's shape, with **exact** cap detection. Two file-local
helpers — `probeLimit(rowCap)` binds `LIMIT rowCap+1` (the extra row is never
rendered; its arrival IS the fact "more matches exist"; `rowCap <= 0` passes
through unchanged to preserve the `LIMIT 0` → `emptyAnswer` edge) and
`rowCapNotice(capped, rowCap)` (single wording site; empty string when not
capped, which is what keeps every uncapped answer byte-identical). All three
arms bind the probe, break at the cap, and end by rendering the notice. The
symbol arm's notice lives inside `renderSymbolRows` — the single writer whose
own doc comment (163 lane) reserved the spot — so the ELSEWHERE fallback is
covered structurally. `answerCodeCheck`'s doc comment states the invariant a
fourth kind must follow. Probing not inference, deliberately: a false "more
matches exist" on a genuinely complete exactly-at-cap answer is this defect
inverted, and `formatRowsText` (`diagnose_load_runtime_action.go:647`) already
detects its cap by exactly this probe — the convention was in-family.

Tests (`diagnose_code_lookup_rowcap_test.go`, 10 tests, sqlmock per the 172
precedent, cap lowered to 2): per-arm induced-at-cap (the `ExpectQuery` LIMIT
arg = 3 is the structural guard against reverting the probe bind); per-arm
exactly-at-cap negative controls in equality form (same rows rendered at
rowCap=2 and rowCap=99 must be EQUAL — immune to the post-163 baseline drift
this file warns about); the capped-ELSEWHERE branch; the notice-lands-before-
the-deferred-doc-block placement; pure tests for both helpers. **Mutation-
tested both ways**: reverting the probe bind fails the at-cap test
(`argument 2 expected [int64 - 3] does not match actual [int64 - 2]`);
switching to `n >= rowCap` inference fails all three negative controls
("a notice here is a false claim"). Verified from a clean `git archive HEAD`
overlay, not the shared dirty tree.

**Residual, known and one line wider, deliberately not fixed here:** a capped
ELSEWHERE fallback that matches only doc rows (`altCode == 0`) discards its
builder — now including its cap notice — while the doc rows still flush under
an "answered: 0 rows" answer. Pre-existing D12 × 163 interplay; belongs to
whichever thread takes that incoherence, not to this fix.

**OPEN until the roll** (fixed-AND-live bar): needs the next chassis image +
pod-verify (`grep -ac "this answer is CAPPED"` + a positive control, both
replicas — budget ~100s per grep, see the 08-05 LANDMINE on slow BusyBox grep
over these binaries), then an induced live check: an `ls` code_check on
`platform/orchestration/actions/` (305 paths this morning) must carry the
notice; any under-cap check must not.

**Follow-up owed, recorded here so it survives this session:** the five capped
landmine-verifier renders are identifiable in `llm_call_log` (timestamps
above). Per the fleet practice that *a pass from a blind check outlives the
blindness*, someone should re-examine whether any landmine-verifier verdict
rested on one of those capped listings read as complete — re-run those checks
post-roll or flag the affected verdicts.

### Council verdict, read in full 2026-08-06 — APPROVED, 12/12 voted, 0 unreadable

Correlation `f22f7ff1-7293-4d41-bcde-edc188fa6218`, ~3.5 minutes submit→verdict.
One advisory objection, `bug_historian`, and it is the same seat making the
same structural point that produced this very file: *"fixing the 4th instance
of a named recurring mechanism without a companion inventory/audit of the
mechanism itself (all silent row-caps in this codebase) all but guarantees a
5th. Router: proceed, but consider requiring the inventory as a condition or
immediate follow-up rather than deferring indefinitely."*

**SECOND FOLLOW-UP OWED, therefore — a codebase-wide silent-cap inventory.**
The 172 round's inventory covered the diagnose loop and found this file's
defect; the seat now asks for the whole codebase. Whoever takes it, the
method that worked here and the trap to avoid are both on record: audit
*every* discard site (SQL `LIMIT` binds, slice truncations `x = x[:cap]`,
loop `break`s on a budget) and for each ask "does the OUTPUT say so?" —
grepping only for the shape the last bug happened to have is precisely how
passes one to three each missed the next site (this file's own chain
paragraph). 016b §9's family entry now carries the probe-not-inference
convention to apply at any SQL-LIMIT site found wanting. Scope honestly:
sites whose output feeds an LLM prompt or a human-read artefact first; a
cap on an internal working set that nothing renders is a different, lesser
class. Not started this session — recorded so it cannot be silently dropped.

The code commit (`df281f6ba`) carries `Council-Submitted:` and is credited
automatically by the `098` report now the correlation is approved — no amend,
per the council runbook. **Never copy `Council-Reviewed:` onto it retroactively.**

---

## CLOSED 2026-08-06 — FIXED AND LIVE on v1.0.1259, and INDUCED IN PRODUCTION, both directions in one run

### 1. The binary carries it — pod-verified, both replicas, with controls

```
agent-chassis-5cf5db5bd8-54xsx   grep -ac "this answer is CAPPED"            -> 1   (the fix)
                                 grep -ac "coverage was capped, not complete" -> 6   (positive control, pre-existing)
                                 grep -ac "zzq_control_181_xyzzy"            -> 0   (negative control)
agent-chassis-5cf5db5bd8-ldx5z   identical: 1 / 6 / 0
```

**The provenance was checked, not assumed** (the fleet landmine: a roll is not
evidence your fix shipped — the image may predate your commit). Commit
`df281f6ba` at 09:43:01 **+0100** = 08:43 UTC; the pod's binary is dated
**10:39 UTC**, ~2h later. Note the trap in that comparison: `git log` prints BST
here and `kubectl` prints UTC, so a naive read makes a shipped fix look
un-shipped. `grep -ac`, not `strings | grep -c`, and budgeted ~280s per exec
per the 08-05 landmine on slow BusyBox grep over these ~95MB binaries.

### 2. It FIRES — induced live, and the negative control is in the SAME run

Orchestration `f370cfe4-8ea0-4e25-b924-36eb4e423c4f`, `COMPLETED @ complete`,
two checks in one payload so the positive and negative observations cannot be
two runs compared against each other:

| check | rows rendered | cap notice |
|---|---|---|
| `ls` `platform/orchestration/actions/` — 305 distinct paths vs cap 40 | **40** | **PRESENT** |
| `ls` `platform/orchestration/actions/diagnose_code_lookup` — under cap | 1 | **absent** |

The notice as it actually rendered:

```
> this answer is CAPPED (row_cap=40): the query matched more rows than are shown,
and rows are ordered by path — the missing matches sort after those shown. Treat
absence from this listing as UNKNOWN, not absent; narrow the query or raise row_cap.
```

That is this file's own § How-to-verify satisfied at the artefact: the over-cap
answer declares itself, the under-cap answer is byte-unchanged.

**How the induction was driven, because the obvious route cannot do it.** All
three live consumers take their `code_checks` from an LLM step
(`review_*.result.code_checks`, `derived.result.code_checks`), so driving any
of them costs credits AND lets the model choose the queries — it cannot aim a
specific over-cap check on purpose. A **temporary** agent
(`diagnosis-181-capcheck`) reading `code_check_fields:
["input_data.code_checks"]` made the induction exact, deterministic and
LLM-free, with `row_cap` deliberately left unset so the default 40 — the value
every real consumer runs with — was the value under test. **The probe row has
been DELETED and its absence proven** (`count(*) = 0` for that type in any
state, snapshots and soft-deletes included). Seed kept out of the repo on
purpose: it is a throwaway, not a mechanism, and registering it would be the
opposite of this bug's own lesson.

### 3. Noticed while inducing, NOT this bug's defect

The rendered answer opened with `!! CODE INDEX STALE: it describes commit
d98010e8 (ref 086_experience_loop), committed 2026-07-28 (9d ago)`. That is the
already-known pinned-ref landmine (migration 252 pins the index ref to a
feature branch), not a regression here — and the freshness banner doing its job
is the positive control for the render being read at all. It does mean the 305
figure is a count over a 9-day-stale index; irrelevant to what was under test
(the cap fired against whatever the index holds, which is the behaviour), but
worth stating rather than letting the number read as current HEAD.

### 4. Both follow-ups remain OPEN and are NOT closed by this

1. **Re-examine the five capped `landmine-verifier` renders** (timestamps in
   § SUPERSEDED) for any verdict that rested on a capped listing read as
   complete. Unstarted.
2. **The codebase-wide silent-cap inventory** the council's `bug_historian`
   seat asked for, so a fifth instance is found by audit rather than
   rediscovery. Unstarted; method and trap recorded above.

Neither blocks closure of *this* defect — the three named caps now report
themselves, live and induced — but both are real, named, and belong to whoever
picks them up.

No `WRONG_CALLS.md` entry for this bug: no claim made during the fix turned out
false. The one mismeasurement in the story — the bundle-corpus search — was the
filing session's, was correctly marked `[UNMEASURED]` rather than asserted, and
is superseded above with the method that worked.
