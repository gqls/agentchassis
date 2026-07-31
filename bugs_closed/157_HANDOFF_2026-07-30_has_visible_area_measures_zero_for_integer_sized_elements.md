# 157 — `has_visible_area` measures 0 on any axis whose rendered size is a whole number, and reports it as "too small to see or click"

> **CLOSED 2026-07-31 18:08 UTC — FIXED AND LIVE, proven at the artefact and with
> a negative control.** Live in `browser-runner-adapter` **`v1.0.1216`**.
> Commits `71680ad513` (fix), `b90990bf4` (gofmt sweep), `f15e00a47` (round 2,
> answering the council). Council **APPROVED round 1** —
> `07639093-3d76-40f4-953b-c3708dac6a1a`, 12 seats, 5 abstained, 0 unreadable, not
> gated by truncation; its 3 advisory objections all closed in `f15e00a47`.
>
> **The three closing checks, in the order the guardian seat asked for them:**
>
> 1. **Pod-grep, long marker + positive controls in one exec** (`bugs_open/153`: a
>    roll is not evidence your fix shipped). On `browser-runner-adapter-fbb78fbbb-rpjl8`
>    (`v1.0.1216`) the new marker `non-numeric w/h in result` → **1**, with five
>    pre-existing controls → **1** each. The same marker was **0** on `v1.0.1215`
>    two hours earlier, so this is a measured transition, not a green reading.
> 2. **The acceptance re-run, in the cluster, against the deployed binary** —
>    correlation `bce1da22-6b47-4fef-bef7-7ef62b488ab4`: **21 passed / 0 failed / 1
>    skipped**, against a pre-fix baseline of **18 passed / 3 failed**. The three
>    failures became passes and nothing else moved. The single skip is `mobile-fit`
>    reporting `not run on profile desktop` — a legitimate **profile gate**, not a
>    `not implemented` skip, which matters because an all-skipped fence records a
>    PASS plus a 7-day cooldown (`LANDMINES.md` 2026-07-30). The two measurements
>    that were wrong now read: `#vtc-c1` **24x24** on desktop AND mobile (was
>    `0x0`), `#vtc-verdict` **386x47** desktop / **143x94** mobile — that mobile
>    figure being the discriminating one-integral-axis case that read `0x94`.
> 3. **A NEGATIVE CONTROL, so the fix cannot be confused with the check being
>    switched off.** Run offline through `try_fence.go`, which calls the fleet's own
>    evaluator (`RunChecksAction.Execute`) against the live URL, with two controls
>    added to the real fence: `has_visible_area` on `#vtc-c1` at an impossible
>    `5000x5000` floor, and one on a selector that does not exist. **Both still
>    FAILED on both profiles** while the four real checks passed — and the
>    impossible-threshold control failed while *printing the true measurement*
>    (`renders 24x24 … needs at least 5000x5000`), which proves in one line that the
>    decode is real AND the threshold comparison still fires. The absent-element
>    control returned `no element matches … after settle` rather than a measurement
>    error, confirming the `{found:false, w:0, h:0}` path the council flagged.
>
> **Follow-on done:** work item `975c3be4-a310-4d7c-aece-f837478d084d`
> (`improve_tool`, "Tier-4 acceptance failed for tool-ai-vendor-trust-checklist")
> was **cancelled as a false positive** — a ticket this bug manufactured against a
> tool that was never defective. That is the severity argument in the filing made
> concrete: this defect did not merely fail to catch things, it *created work*.

**Filed** 2026-07-30. Class: **false NEGATIVE in a live Tier-4 check type** (not a
missing capability, not a vacuous pass). Found while building
`tool-ai-vendor-trust-checklist` as the first end-to-end exercise of the S0–S7
staged build ladder (`features_open/027`, lane `staged_component_build`).

**Severity is about trust, not volume.** The check is three days old, is the one
the ladder lane calls its most valuable instrument, and it fails in the direction
that manufactures work: it accuses a correct element of being invisible, and its
message names a cause that is not there.

---

## Symptom

An acceptance run against a page whose controls are visibly present and clickable:

```
[FAIL] desktop  first-box-is-a-real-target
  #vtc-c1 renders 0x0 on desktop — present in the DOM but too small to see or
  click (needs at least 20x20). A collapsed flex/grid child is the usual cause:
  check that its parent establishes a height.

[FAIL] mobile   verdict-has-area
  #vtc-verdict renders 0x94 on mobile — present in the DOM but too small to see
  or click (needs at least 24x24).
```

In the same run, on the same page, **the interaction checks that CLICK `#vtc-c1`
all passed** — Playwright's `Click()` enforces actionability (visible, stable,
enabled), so the element it just clicked cannot have been 0x0.

Screenshots of the same URL at 1366x1500 and 390x900 show the checkbox rendering
as a normal ~24px control and the verdict box at full column width, on both
profiles.

## Root cause `[VERIFIED — code path read in the exact module the repo builds]`

`internal/adapters/browserrunner/run_checks_action.go:703-721`:

```go
v, err := c.page.Evaluate(`(sel) => { ... return {found: true, w: r.width, h: r.height}; }`, selector)
...
w, _ := m["w"].(float64)   // <-- silently 0 when the value is not float64
h, _ := m["h"].(float64)
```

`github.com/mxschmitt/playwright-go@v0.6100.0/js_handle.go:109-114` (the version
in `go.mod:23`):

```go
if v, ok := vMap["n"]; ok {
    if math.Ceil(v.(float64))-v.(float64) == 0 {
        return int(v.(float64))     // integral numbers come back as int
    }
    return v.(float64)              // fractional numbers come back as float64
}
```

So the library returns **`int` for a whole number** and `float64` only for a
fractional one. The comma-ok assertion in `VisibleArea` therefore discards every
integral measurement and yields **0**, and the check compares 0 against its
threshold and fails.

**This is why the failures look arbitrary but are not.** Every observation fits:

| element | real size | decoded as | reported |
|---|---|---|---|
| `#vtc-c1` (CSS `width:24px; height:24px`) | 24 x 24 | `int`, `int` | **0x0** |
| `#vtc-verdict` desktop (text-driven box) | 386.xx x 47.xx | `float64`, `float64` | 386x47 (correct) |
| `#vtc-verdict` mobile | integral width x 94.xx | `int`, `float64` | **0**x94 |

The mobile row is the discriminating one: **one axis integral, one fractional,
and exactly the integral axis reads 0.** No theory about collapsed flex children
predicts that.

## Who is affected

- **Blast radius is exactly two lines.** `grep -n '\.(float64)'
  internal/adapters/browserrunner/*.go` returns only `run_checks_action.go:718`
  and `:719`. `HorizontalOverflow` and the other evaluators decode into typed
  structs and are unaffected. So this is `has_visible_area` alone.
- **The elements most likely to hit it are the ones the check exists to police.**
  Anything deliberately sized in whole pixels — icon buttons, checkboxes, avatars,
  spinners, `width:24px` touch targets — lands on an integer. Text-driven boxes
  usually land on a fraction and pass, which is why the check has looked fine.
- **`min_width`/`min_height` cannot be tuned around it.** The measured value is
  0, so every positive threshold fails.

## Fix candidates, ordered by what closes the door

1. **Coerce the number instead of asserting one concrete type** (~6 lines, in
   `VisibleArea`). Makes the bad state unrepresentable at the only place it can
   occur:
   ```go
   func evalNumber(v interface{}) float64 {
       switch n := v.(type) {
       case float64: return n
       case int:     return float64(n)
       case int64:   return float64(n)
       case int32:   return float64(n)
       case json.Number: f, _ := n.Float64(); return f
       }
       return 0
   }
   ```
   Then `w, h := evalNumber(m["w"]), evalNumber(m["h"])`.
   **Prefer this.** It is local, it needs no change to any fence, and it fixes
   every existing false failure without touching the check's semantics.

2. **Return the rect from JS as a string or a float-forced number**, e.g.
   `w: r.width + 0.0000001` or `String(r.width)`. Rejected: it hides the type
   problem in the payload rather than fixing the decode, and the next evaluator
   added to this file repeats the bug.

3. **Decode into a typed struct** the way the other evaluators already do. Sound,
   and more consistent with the file, but a larger edit than (1) for the same
   result.

Whichever is taken: **a `0` from a successful measurement must be distinguishable
from a `0` that means "not decoded".** The current code cannot tell them apart,
which is the reason the wrong answer is stated so confidently. Consider failing
loudly on an undecodable value rather than treating it as zero — a bookkeeping
failure should not present as a layout verdict.

## How to verify a fix

1. Re-run acceptance for `tool-ai-vendor-trust-checklist` on
   `leopardessconsulting.co.uk` — a page **known good by screenshot** whose
   checkbox is exactly `24px`:
   ```
   ./docs/leopardessconsulting/scripts/tool_acceptance_run.sh \
     4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk \
     tool-ai-vendor-trust-checklist
   ```
   Expect `first-box-is-a-real-target` to report `24x24` and pass, and
   `verdict-has-area` to pass on **mobile** as well as desktop. Baseline before
   the fix: 18 pass / 3 fail / 0 unexpected skips, correlation
   `dc952633-4bc9-4395-a12f-4e13118f6540`.
2. **Keep a negative control in the same run**, or the fix cannot be told from
   the check being switched off: point one `has_visible_area` at an element that
   really is collapsed and confirm it still fails.
3. Grep the running pod for the fix with a LONG marker plus a pre-existing
   control in the same exec — a short marker returns 0 on a binary that supports
   it, because Go compiles short literals to immediate comparisons that never
   reach rodata.

## Note for anyone reading the vendor-trust tool's PLAN

Those three failures are **this bug, not a defect in that tool.** Its checkbox is
`24px` on purpose (WCAG 2.2 target size) and set in pixels rather than rem so a
site-level root font-size cannot shrink it. **Do not "fix" it by making the size
fractional** — that would silence the gate on the one page that currently
demonstrates the bug. Recorded in the tool's PLAN addendum
(`doc_plans`, `subject_type='tool'`, `subject_key='tool-ai-vendor-trust-checklist'`).

## Cross-links

- `bugs_open/084` — owned/ported pages get no *automatic* browser verification.
  Different class (coverage, not measurement), but the same tier. This run had to
  be fired by hand, which is 084's point.
- `LANDMINES.md` — entry added 2026-07-30, footprinted on
  `run_checks_action.go` `VisibleArea` and the `has_visible_area` check type.
- Concept register `TL-034` records `has_visible_area` as **built**; that status
  is right, and this bug is why "built" and "trustworthy" are not the same row.

---

## The fix as shipped (2026-07-31, `71680ad513`)

Fix candidate 1 was taken, and **broadened from two lines to the file's numeric
decode contract** for one reason found while implementing it:
`HorizontalOverflow`, 200 lines below `VisibleArea` in the same file, **already
hand-rolled the correct `int`/`float64` switch** —

```go
// JS numbers come back as float64/int; tolerate 2px of rounding.
switch n := m["over"].(type) {
case float64: ...
case int:     ...
```

So the file held the same fact twice, right in one place and wrong in the other.
A two-line patch to `VisibleArea` would have left that duplication standing and
the next evaluator added to this file free to repeat the bug. Both sites now go
through one decoder:

```go
func evalNumber(v interface{}) (float64, bool)   // float64 | int | int64 | int32 | json.Number
```

documented in its own comment as *the* way to read a number out of a
`page.Evaluate()` result here. Its accepted set is **exactly** `float64` and
`int` — what `parseValue` can emit — after the council's edit-quality seat
correctly called a wider "defensive" set dead code (see below). `HorizontalOverflow` is behaviour-preserving by
construction: same accepted types, same `> 2` comparison, same fall-back to
`info.Clipped` on an undecodable value.

**The `bool` is the load-bearing half, and it is the part the original bug report
asked for.** `VisibleArea` now returns an **error** on an undecodable value rather
than 0. The call site already reports `err` as `could not measure <sel>: …`,
distinct from the threshold message — so a bookkeeping failure can no longer
present as a layout verdict. That indistinguishability is precisely why the check
stated a cause ("a collapsed flex/grid child") it had never observed.

### Blast radius, re-measured at fix time rather than carried forward

`grep -rln playwright --include='*.go' .` (excluding tests) returns **exactly two
files**: `run_checks_action.go` and `render_audit_action.go`. Of the three
production `Evaluate` sites, `render_audit_action.go:282` decodes via a
`json.Marshal` → `json.Unmarshal` round-trip, which treats `int` and `float64`
identically and **is unaffected**. This independently confirms the original
report's two-line claim, by a different route (import census rather than a
`.(float64)` grep — a grep proves an absence only for the spelling it searches).

### Why the existing tests could never have caught it, and what was done instead

`browserPage.VisibleArea` is **already typed** `(float64, float64, bool, error)`.
The fault lives *below* that interface, in the decode of playwright's own result,
so `fakePage` cannot express it — **`TestHasVisibleArea` was green through the
entire life of the bug**, including its five table cases and the exact 1146x0
shape. This is the general lesson: *a test double that is typed cannot reproduce a
type-decode fault.*

So the decode is tested **directly**, at two levels:
`TestEvalNumberDecodesEveryShapePlaywrightReturns` over every shape playwright-go
can return plus the ones that must report `!ok`, and
`TestDecodeAreaHandlesEveryShapeTheEvalScriptReturns` over the whole result map
including the `{found:false, w:0, h:0}` not-present shape. Per the "a quiet test passes when the RULE is gone" rule, it was
**confirmed red before being accepted**: the `int` case was mutated to return 0
(reproducing the old comma-ok behaviour) and only that subtest failed —

```
--- FAIL: TestEvalNumberDecodesEveryShapePlaywrightReturns/int_—_what_playwright_returns_for_a_whole_number
```

Two more tests pin the behaviour end-to-end: `TestIntegralSizesClearTheDefaultFloor`
(a 24x24 control against the 24x24 floor, and `#vtc-verdict`'s discriminating
one-integral-one-fractional mobile case) and
`TestUnmeasurableAreaReportsMeasurementFailureNotTooSmall`, which asserts the
detail says `could not measure` and does **not** say `too small to see or click`.

### What the council caught (round 1 APPROVED, 3 advisory items, all closed)

Worth reading even though the verdict was approve — two of the three were fair
hits, and one of them was a real hole.

1. **Three seats independently flagged the same gap** (edit-quality *medium*,
   guardian *low*, architecture *watch-item*): the submission's own risk note said
   the not-present shape `{found:false, w:0, h:0}` must keep decoding cleanly
   rather than trip the new error path — **and nothing tested it.** *"A
   self-identified risk with no test closing it is a fixable gap."* Correct, and
   the stake was real: had `found:false` errored, a selector matching nothing would
   have flipped from the deliberate `no element matches … after settle` FAIL to a
   `could not measure` bookkeeping failure — changing what the check reports for
   the commonest case. Closed by splitting the post-`Evaluate` half out as
   `decodeArea()`, testable without a browser, with a 7-case table over the shapes
   the eval script actually returns. **Mutation-proven:** making `decodeArea` error
   on `!found` fails exactly that subtest and no other.
2. **`evalNumber` was too WIDE, not too narrow.** It took `int64`/`int32`/
   `json.Number` "defensively"; the edit-quality seat called it dead code, since
   `parseValue` emits only `int` and `float64`. It was worse than dead — **a
   speculative arm makes the contract untestable and quietly widens what counts as
   a measurement.** Narrowed to exactly those two, with the three dropped types
   now asserted as NEGATIVE cases so the narrowness is checked rather than assumed.
   Anything else is a loud `could not measure`, which is the designed behaviour.
3. **`datahelpers.ToFloat64` already exists with the identical signature** —
   `(v interface{}) (float64, bool)` — and the reuse seat was right that it should
   have been searched for before a new helper was written. **The local one was kept
   deliberately**, and the reason is now at the function: `ToFloat64` also
   `ParseFloat`s a **string**, and rejected fix candidate 2 above was *"return the
   rect from JS as a string"*. A decoder that silently parses strings would make
   that payload change work by accident and hide the very class of type problem
   this bug is about. The *not looking* is logged in `WRONG_CALLS.md`; the
   divergence is a reasoned choice, not the miss.

Two further notes taken but not code changes. The **bug_historian** pointed out
that a grep for a helper *name* cannot prove no other site decodes inline; census
over both playwright-importing files now shows `.(float64)` appearing **zero**
times outside a comment, and every remaining assertion on an `Evaluate` result map
is `.(string)` or `.(bool)` — unambiguous, because only playwright's `"n"` branch
is value-dependent. It also noted the **sibling defect in the same file is still
open**: an unknown check type is SKIPPED and an all-skipped fence PASSES
(`LANDMINES.md` 2026-07-30) — not in scope here, but the subsystem still has one
instance of the broader silent-false-positive pattern. The **guardian** asked that
the live acceptance re-run be *a condition of closing this bug rather than a risk
footnote* — agreed, and that is what the STATUS banner at the top now says.

### What still has to happen — the verification is NOT complete

1. **A build and roll of `browser-runner-adapter`.** Not done by this session:
   another session had an in-flight release at `v1.0.1215` (uncommitted
   `IMAGE_TAG` and overlay bumps across every service), and building over it
   would have collided with their release.
2. **Pod-grep, not a tag and not `git log`** (`bugs_open/153`: a roll is not
   evidence your fix shipped). Long marker plus a positive control in the same
   exec — a short literal returns 0 on a binary that supports it:
   ```
   kubectl -n ai-persona-system exec <pod> -- sh -c \
     'for s in "non-numeric w/h in result" "too small to see or click" \
               "in the live DOM after settle"; do \
        printf "%s => " "$s"; grep -ac "$s" /app/browser-runner-adapter || true; done'
   ```
   Expect the first to become **1**. It was **0** with the other two at **1** on
   2026-07-31 15:26 UTC.
3. **The acceptance re-run in "How to verify a fix" above, WITH its negative
   control.** Point one `has_visible_area` at a genuinely collapsed element in the
   same run, or a green board cannot be told from the check having been switched
   off — which is the failure mode this very check type already has form for
   (the skipped-check PASS + 7-day cooldown, `LANDMINES.md` 2026-07-30).
