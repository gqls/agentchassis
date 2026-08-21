# 351 — every calculator in the library is invisible to the selector, because the completeness guard tests for a WRAPPER TAG and calls a complete widget truncated

Filed 2026-08-21 by the `remortgagecalculator.uk` lane, out of the residual left when
`bugs_open/311` closed. **311 is CLOSED and correctly so** — its own fix (divert on collision) is
live and demand-proven, and the originating page now serves a real calculator. This file is the
half 311's closure did **not** fix, named in its close message as *"residuals stay open in their
own files"*.

**Status: OPEN, not started. Diagnosis complete and calibrated; no code written.**
Discussed with the `311 continued` lane the same evening — they verified the decisive arm in code
(see §4) and offered the work either way. `plan_sections_action.go` is territory that lane has
mapped but does not own.

## The bug in one paragraph

The component selector finds section components by `section_type`. **All 22 active calculators in
the library carry `section_type = NULL`, so none of them can ever be selected** — while every other
category is ~100% shelved. That alone would be a data problem. The reason it is a *code* problem is
what happens next: even when a calculator IS found by another path, `sectionTemplateValid` drops it,
because that predicate tests for the literal string `</section>` as a proxy for "not truncated" —
and a calculator is a `<div>`-wrapped self-contained widget that never contains one. **Measured: 0
of 22 pass the current guard; 22 of 22 are structurally complete.** So the platform cannot reuse a
calculator it already owns, and every site that wants one pays a full LLM generation instead — which
is what exposed `remortgagecalculator.uk` to `bugs_open/345`'s retry loop and burned 11 attempts
across five items on that one page.

## Evidence

### The shelving gap is categorical, not scattered

Live, `is_active AND component_level='section' AND forked_from IS NULL`:

| category | total | with `section_type` |
|---|---|---|
| custom | 25 | 25 |
| content | 22 | 22 |
| general | 22 | 21 |
| interactive-platform | 11 | 11 |
| **calculators** | **22** | **0** |

### The guard calls every one of them truncated, and every one is complete

All 22 templates run through the platform's own predicates, in-package:

```
read=22   sectionTemplateValid PASS = 0/22   toolTemplateValid PASS = 22/22
```

`sectionTemplateValid` (`plan_sections_action.go:1784`) is, verbatim, "empty → true; `len < 100` →
true; else `strings.Contains(html, "</section>")`". Its own comment says the only invalid case is
*"a long template with no closing `</section>` tag — the signature of a truncated LLM generation"*.
For this population that signature is wrong 22 times out of 22. `b89f91e1` ("Mortgages Repayment",
4,451 B) opens `<div class="calc-grid">` and closes with a terminated IIFE and `</script>`.

> **⚠ The first version of this measurement was VACUOUS and I nearly reported it.** Exporting via
> `COPY` with a tab separator, the tab was escaped, `SplitN` matched nothing, zero templates were
> read, and the harness dutifully printed `0/22` for BOTH predicates — a zero from a loop that never
> executed. Re-run per-file with an explicit `read` counter and a `t.Fatal` when it is zero. **Any
> re-run of this must keep the vacuity guard**; the shape of the wrong answer here is identical to
> the shape of the right one.

### The code predicted this recurrence

`componentTemplateValid`'s header records `bugs_open/024`: *"`loadSingleComponentSchema` was still
rejecting self-contained tool templates on the `'</section>'` marker and returning nil, silently …
the council's bug_historian seat predicted the second call site from this platform's documented
history of the same filter existing twice, and it was right."* That fix gave `component_level='tool'`
a structural predicate. **These 22 sit at `component_level='section'` and never got it.** Same
defect, same marker, one level over.

## Why the obvious fixes are wrong — this is the load-bearing section

**Do NOT backfill `section_type`.** `bugs_open/311` refuted it and was right; this file supplies the
reason it lacked. Backfilling makes all 22 selectable and then the guard drops all 22 — motion with
no effect, and 311's phrase "actively harmful for guard-dropped ones" is exactly the outcome.

**Do NOT widen the `041` backstop's gate alone** (`component_selector.go:346`, gated on
`norm := NormalizeComponentFunction(sectionType); norm != sectionType`, so it can never fire for an
already-kebab name like `mortgages-repayment`). **Verified in code by the `311 continued` lane:**
both loaders route through `componentTemplateValid` (`plan_sections_action.go:1794`), whose header
reads *"A component that is dropped here is invisible downstream — no error, no work item — which is
what made the original defect cost three fix cycles."* So a widened backstop would find the
incumbent by `function`, decline to raise `needs_new_component`, and the loader would then drop the
row at the same defective predicate — **no queued item and no rendered section. Strictly worse than
today.**

## The fix, calibrated fleet-wide in BOTH directions

**Make `sectionTemplateValid` structural** — `UnbalancedStructuralTags` + `endsCleanly`, i.e.
`toolTemplateValid`'s semantics — instead of testing for a wrapper tag. Leave both callers alone;
`componentTemplateValid` stays the single gate.

Calibrated over **all 148** active section-level components with a template:

| | count |
|---|---|
| pass → pass (unchanged) | 124 |
| fail → fail (unchanged) | 0 |
| **fail → PASS (rescued)** | **23** |
| **pass → FAIL (regressed)** | **1** |

The 23 rescued are the 22 calculators plus `f02d244b`. **The single regression is a FALSE one and
must be fixed as part of this change**, not accepted:

- `6c41404d` `about-commercial-block` (`content`, 3,699 B) ends `…</div>\n</section>{{end}}`.
- It is complete. The whole section is wrapped in a conditional, so it legitimately ends on a Go
  template action rather than on markup.
- `endsCleanly` (`component_write_guard.go:318`) is `strings.HasSuffix(strings.TrimSpace(s), ">")` —
  the last character here is `}`, so it returns false.

**So the change is two parts:** the structural predicate, **plus** tolerating a trailing template
action before the `endsCleanly` test (trim `{{…}}` from the tail, or accept `}}`).

**Blast radius of that second part, measured because `endsCleanly` is shared with tool validation:**

```
section  total=148  ends-on-template-action = 1   (the regression above)
tool     total=121  ends-on-template-action = 0
```

**Zero tools end that way**, so the tolerance cannot loosen tool truncation-detection on today's
fleet. With both parts: **23 rescued, 0 regressed.**

⚠ **Calibrate again before shipping, and in both directions** — the `311 continued` lane's warning,
from the 024/303 round: the `</section>` marker misclassifies BOTH ways, and *"4 of the 8
genuinely-cut tools CONTAINED `</section>` and passed the marker"*. The numbers above are today's
corpus (148 sections / 121 tools); a later corpus can differ, and every `pass → FAIL` flip needs a
hand check like the one above rather than a count.

## What this does NOT need, and why

**A writer hunt is probably a dead class.** The 22 are `created_at` 2026-08-13/14/15 with
`created_from='manual'` — one adoption batch (the LMC calculator work), not a generator. The current
store path cannot reproduce it: 311's diversion writes `section_type` on every new row (the row
created for this very site tonight, `5d3bc513`, carries `section_type='mortgages-repayment'`), and
the regeneration UPDATE self-heals a NULL via `COALESCE`. **Check whether the manual/adoption route
is scheduled to run again before hunting a writer** — if it is not, the fix above ends the class.

**If you later choose to backfill anyway, decide the ordering deliberately.** Seven incumbents now
also have a diverted twin whose `section_type` IS that vocabulary. After the fix, Path 1's
function-match surfaces the incumbent while the selector's `section_type` match surfaces the twin;
after a backfill, both match by `section_type`, and which wins depends on path order and the
selector's `is_active` ordering. Nothing is damaged today because `page_components` already binds
each page, but a replan walks into the resolver-conflict-window landmine. Either decide it in the
migration or state that incumbents stay Path-1-only — silence is the only wrong answer.

## How to verify a fix

1. Re-run the calibration above; assert **read=148** (never trust a zero without it), rescued ≥ 23,
   regressed = 0, and hand-check any new flip.
2. On a site with a planned calculator section and no existing binding, the selector must return the
   library incumbent and **no `needs_new_component` item is filed at all** — that is the whole point:
   a reused calculator costs no LLM generation and cannot enter `345`'s retry loop.
3. At the artefact: the page serves the calculator's inputs.

## Related

- `bugs_closed/311` — the CREATE half (fixed, live, demand-proven). This file is its REUSE residual;
  read its CONTRIB sections for the measurements that led here.
- `bugs_open/345` — the retry loop this defect feeds. Reuse would have avoided generation entirely.
- `bugs_open/024` / `303` — the same `</section>` marker defect at `component_level='tool'`, and the
  precedent for the remedy and for calibrating it.
- `bugs_open/309` — owns the field-source vocabulary guard that refused the generated template.

---

## REFINEMENT 2026-08-21 (later, from the `311 continued` lane) — the fix moves DOWN a level, and my proposed shape was wrong

Two corrections to the fix as written above, both from that lane, both verified here before recording.

### There is a FOURTH caller of `endsCleanly`, and it makes the special-case version harmful

My blast-radius measurement covered `toolTemplateValid` and missed a live enforcement site. Every
caller, enumerated (`grep -rn "endsCleanly(" platform/ internal/ pkg/ cmd/`, non-test):

| site | use |
|---|---|
| `plan_sections_action.go:1651` | log field only |
| `plan_sections_action.go:1853` | `toolTemplateValid` — the one I measured |
| `create_tool_component_action.go:173` | log field only |
| **`component_write_guard.go:260`** | **live WRITE-time regression check** |

That fourth one is:

```go
if endsCleanly(currentHTML) && !endsCleanly(newHTML) {
    issues = append(issues, "replacement ends mid-token (%q) where the current template ends on a
    closed tag — the completion was cut mid-stream")
}
```

**So the same false positive already refuses legitimate work at BIRTH:** a generator rewriting a
section into a conditional wrapper produces `current` ending `>` and `new` ending `{{end}}`, and is
refused as "cut mid-stream". `6c41404d` presumably predates this guard or arrived through the manual
adoption route.

**Consequence for the fix: repair `endsCleanly` ITSELF, not `sectionTemplateValid`'s use of it.**
Special-casing only the section predicate would leave the write guard refusing exactly the shape the
loader newly accepts — a drift pair between two guards making the same judgement, which is the
class `componentTemplateValid` was created to end (its own header: two call sites making an
identical judgement, and the first fix patched only one). Stated cost of the shared fix: it slightly
loosens the write-time regression check for the tolerated shape. That is acceptable only because the
tolerance is narrow — see next.

### My proposed shape ("accept a trailing `}}`") was WRONG and would have passed real truncations

The correct rule is **strip trailing `{{end}}` repetitions (whitespace-tolerant), THEN require `>`**.
Never accept a bare `}}` suffix.

The discriminating case: a template genuinely cut immediately after any complete mid-template action
also ends `}}`. A bare-suffix rule passes that cut. Strip-then-check does not — after removing
complete trailing `{{end}}` tokens, the remainder of a mid-cut ends on prose or an open tag, not
`>`. And `{{end}}` is the ONLY action that legitimately terminates a template (a conditional or
range wrapper, possibly nested — hence strip repeatedly); a tail of `{{if …}}`, `{{range …}}` or a
bare placeholder is suspicious in every case.

### Measured over BOTH populations, which also covers the `:260` direction

Proposed rule implemented as `endsCleanlyV2` (regex-strip `\s*\{\{-?\s*end\s*-?\}\}\s*$` repeatedly,
then `HasSuffix(">")`) and run against every active section- and tool-level template:

```
rows read: section=148  tool=121  total=269
verdict flips (all directions):                                        1
rows a BARE '}}' rule would wrongly admit that strip-then-check refuses: 0
```

The single flip is the intended rescue — `6c41404d`, `false → true`, tail
`{{end}} {{end}} </div> </section>{{end}}`. Because `endsCleanly` is the shared function, this one
census covers `component_write_guard.go:260` as well as `toolTemplateValid`: exactly one row's
verdict changes anywhere, and at `:260` the effect is to stop one false "cut mid-stream" refusal
while leaving every true refusal intact.

**Two honesty notes on that table, because the numbers flatter the change:**

- The `0` in the second row means today's corpus contains **no genuine mid-action truncation**, so
  the bare-`}}` rule and the strip rule agree on everything we can currently see. Strip-then-check is
  therefore **not proven better on today's data — it is better on data we have not seen**, which is
  precisely the case a truncation guard exists for. Recorded as reasoning, not as a measurement.
- The repeated-strip loop is **defensive, not demand-proven**: `6c41404d` needs one strip (its other
  `{{end}}`s are mid-template). Nested trailing wrappers are plausible and cost nothing to handle,
  but no row in the corpus exercises the loop.

### Fix, restated

1. `endsCleanly` (`component_write_guard.go:318`) — strip trailing `{{end}}` repetitions, then
   require `>`. This is the whole change; all four call sites inherit it coherently.
2. `sectionTemplateValid` (`plan_sections_action.go:1784`) — replace the `</section>` substring test
   with the structural pair (`UnbalancedStructuralTags` + `endsCleanly`), i.e. `toolTemplateValid`'s
   semantics. Leave both callers alone; `componentTemplateValid` stays the single gate.

Combined, over the corpus above: **23 rescued, 0 regressed, 1 verdict flip fleet-wide.**

### Implementation notes owed to whoever takes it

- **Both files carry other sessions' uncommitted edits on some evenings.** Check the tree
  immediately before committing, pathspec `plan_sections_action.go` AND
  `component_write_guard.go` explicitly, and expect the same-file passenger — that lane's own 345 Go
  half rode into an unrelated session's commit this very evening.
- The calibration above (both directions, plus the `:260` note) is the council submission's review
  story and should land in it rather than be re-derived.
- Re-run the calibration against the live corpus before shipping; assert the read counts.
