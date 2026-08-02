# 181 — the code-lookup's three row caps are silent, while a sibling cap eight lines away reports itself

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

## Related

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
