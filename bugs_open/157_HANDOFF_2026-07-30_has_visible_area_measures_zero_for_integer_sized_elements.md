# 157 — `has_visible_area` measures 0 on any axis whose rendered size is a whole number, and reports it as "too small to see or click"

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
